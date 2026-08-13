# apps/api

Módulo Go único del backend: procesos `api` y `worker`. Ver
[`docs/04-arquitectura/backend-go.md`](../../docs/04-arquitectura/backend-go.md)
y [`docs/03-desarrollo/estandar-backend-go.md`](../../docs/03-desarrollo/estandar-backend-go.md)
antes de agregar código.

## Estado

- Configuración, logger, servidor HTTP con `/health`
- Paquetes de módulo vacíos (`internal/modules/*`) con su comentario de responsabilidad
- **Database**: pool pgx v5, `InTenantTx` para trabajo tenant-aware, health check de BD
- **HTTP (HU-003)**: router Chi v5 con las tres audiencias
  (`/api/v1/public`, `/api/v1/customer`, `/api/v1/private`, todavía sin
  operaciones de negocio), middleware base completo y errores uniformes
  RFC 9457. Ver la sección siguiente.

## Requisitos

- Go 1.25 o superior.
- PostgreSQL 14+ con extensiones `btree_gist`
- Variables de entorno (ver `internal/platform/config`)

## Comandos

```bash
go run ./cmd/api
go run ./cmd/worker
go build ./...
gofmt -l .
go vet ./...
go test ./...
go test -race ./internal/platform/database/...
```

## Patrón obligatorio: errores uniformes y router (HU-003)

`internal/platform/httpserver.NewRouter` monta las tres audiencias y aplica,
en este orden, el middleware que no depende de autenticación, tenant ni
idempotencia (`docs/04-arquitectura/backend-go.md` sección 6, pasos 1-6):
identificador de solicitud, recuperación de pánico, límite de tamaño del
cuerpo, timeout por contexto, registro estructurado y cabeceras de
seguridad. Los pasos 7-11 (rate limit, autenticación, tenant/RLS,
idempotencia) llegan con las historias que los necesitan.

### Cómo señalar un error desde dominio o servicios

El dominio y los servicios **nunca** importan `net/http` ni Chi. Devuelven
un `*apperr.Error`:

```go
import "system-barbershop/internal/platform/apperr"

func (s *MiServicio) Buscar(ctx context.Context, id string) (*Turno, error) {
    turno, err := s.repo.Buscar(ctx, id)
    if errors.Is(err, sql.ErrNoRows) {
        // Mismo Kind para "no existe" y "es de otra barbería": la capa
        // HTTP no tiene forma de distinguirlos (CA-003-03, RN-TEN-01).
        return nil, apperr.NotFound("no existe un turno con ese identificador")
    }
    if err != nil {
        return nil, apperr.Internal(err) // la causa nunca llega al cliente
    }
    return turno, nil
}
```

El handler HTTP es el único lugar que traduce ese error:

```go
turno, err := servicio.Buscar(r.Context(), id)
if err != nil {
    httpserver.WriteProblem(w, httpserver.Translate(err, httpserver.RequestIDFromContext(r.Context())))
    return
}
```

`Translate` es el único punto que decide `type`, `title`, `status`, `code`
y `detail`: nunca construyas un `httpserver.Problem` a mano fuera de ahí,
porque eso es lo que garantiza que `detail` no filtre SQL, rutas de
archivo, nombres de proveedor ni versiones (CA-003-02).

### Logger: lista permitida de campos

`httpserver.RequestLogger` solo registra `request_id`, `method`, `route`
(el patrón de ruta que resolvió Chi, nunca `r.URL.Path` crudo: una ruta de
`/api/v1/customer` lleva un token en la URL), `status` y `duration_ms`.
Cualquier log de negocio dentro de un handler o servicio sigue la misma
regla (RN-DAT-02): nunca nombre, teléfono, correo, contraseña, token,
cookie ni cabecera `Authorization`.

## Patrón obligatorio: operaciones tenant-aware (HU-002)

Toda operación de datos ocurre dentro de una transacción con `app.barbershop_id`
fijado con alcance local. Es **IMPOSIBLE** escribir una consulta de negocio sin
contexto: la API no lo permite.

### Cómo se abre una operación tenant-aware

```go
import (
    "context"
    "system-barbershop/internal/platform/database"
)

func (s *MiServicio) MiOperacion(ctx context.Context, shop database.BarbershopID) error {
    return s.db.InTenantTx(ctx, shop, func(ctx context.Context, q database.Queries) error {
        // Aquí SÍ puedes ejecutar consultas. El contexto de barbería YA está fijado.
        var count int
        err := q.QueryRow(ctx, "SELECT count(*) FROM mi_tabla").Scan(&count)
        return err
    })
}
```

### Por qué no existe una vía alternativa

- `InTenantTx` es la **única** función exportada que entrega un ejecutor de
  consultas (`Queries`). El pool NO expone `Query`, `QueryRow`, `Exec` ni `Begin`.
- `BarbershopID` es un tipo propio: no puedes pasar un `uuid.UUID` ni un `string`
  por error (evita confundir ids de usuario, cita, etc.).
- La secuencia interna es atómica: adquirir conexión → `BEGIN` → `set_config`
  local → ejecutar callback → `COMMIT` (o `ROLLBACK` ante error/pánico).
- Si `set_config` falla, se hace `ROLLBACK` y se devuelve error **SIN ejecutar
  el callback**. Jamás se continúa con contexto vacío (CA-002-05).
- El ejecutor (`Queries`) deja de ser válido al retornar el callback. Si
  alguien lo guarda en un struct, la transacción ya estará cerrada.

### La trampa de `SET LOCAL` y por qué NO se concatena

**ESTO NO FUNCIONA** (PostgreSQL no acepta parámetros en `SET LOCAL`):

```go
// INCORRECTO - no compila como consulta parametrizada
tx.Exec(ctx, "SET LOCAL app.barbershop_id = $1", shopID)
```

**LA FORMA CORRECTA** usa `set_config`, que SÍ es parametrizable:

```go
// CORRECTO - parametrizable, alcance local a la transacción
tx.Exec(ctx, "SELECT set_config('app.barbershop_id', $1::text, true)", shopID)
```

El tercer argumento `true` significa "local a la transacción": revierte solo al
terminar la transacción, sin necesidad de `RESET`. Ese comportamiento es
justamente lo que verifica CA-002-03.

**NUNCA** construyas esa sentencia por concatenación, ni siquiera "porque el
valor ya es un uuid validado". La regla es la forma de la llamada, no la
confianza en el valor. Concatenar el identificador en la cadena SQL introduce
inyección SQL en el punto exacto del que depende TODO el aislamiento entre
barberías.

### Excepción autorizada: comprobación de salud

La única lectura fuera del patrón `InTenantTx` es `DB.HealthCheck()` (y el
handler HTTP `DatabaseHealthHandler`). No tiene barbería: verifica
conectividad únicamente. Su respuesta indica `ok` o `unavailable` y **NADA
MÁS** (sin versión, host, usuario, error del driver).

## Pruebas de integración

Las pruebas en `internal/platform/database/*_test.go` requieren PostgreSQL real
con las migraciones aplicadas y `database/testdata/dos_barberias.sql` cargado.
Conéctate como `barberia_app` (rol SIN `BYPASSRLS`, SIN propiedad).

Ejecución:

```bash
export TEST_DATABASE_URL="postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"
go test -race ./internal/platform/database/...
```

Criterios verificados (numeración según docs/02-requisitos/historias-usuario.md,
corregida: antes CA-002-05/06 aparecían aquí intercambiados con
CA-002-04/05):
- CA-002-01: contexto local en cada operación
- CA-002-02: aislamiento concurrente A/B sobre mismo pool (-race)
- CA-002-03: sin residuo de contexto tras devolver conexión al pool
- CA-002-04: ninguna función exportada ejecuta consultas de negocio sin
  contexto (el pool no expone Query/QueryRow/Exec/Begin sueltos)
- CA-002-05: un fallo al fijar el contexto aborta la operación; el callback
  nunca se ejecuta con contexto vacío
- CA-002-06: el dominio y los servicios no importan database ni pgx
  (testdeps_test.go)

Además, sin ser una de las seis CA, TestContextCancellation verifica que
cancelar el context de la petición cancela la consulta en curso en
PostgreSQL (regla de diseño (f) de InTenantTx).

## Migraciones

Gobernadas por Atlas CLI v1.3.0 (fijado, no `latest`). Ver `database/README.md`.

```bash
atlas migrate hash --dir file://database/migrations
atlas migrate validate --env local
atlas migrate apply --env local
```

`atlas.sum` se versiona. NO se ejecutan al arrancar la aplicación (CA-001-07).