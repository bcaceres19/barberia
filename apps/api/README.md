# apps/api

Módulo Go único del backend: procesos `api` y `worker`. Ver
[`docs/04-arquitectura/backend-go.md`](../../docs/04-arquitectura/backend-go.md)
y [`docs/03-desarrollo/estandar-backend-go.md`](../../docs/03-desarrollo/estandar-backend-go.md)
antes de agregar código.

## Estado

- Configuración, logger, servidor HTTP con `/health`
- Paquetes de módulo vacíos (`internal/modules/*`) con su comentario de responsabilidad
- **Database**: pool pgx v5, `InTenantTx` para trabajo tenant-aware, health check de BD

## Requisitos

- Go 1.23 o superior.
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

Criterios verificados:
- CA-002-01: contexto local en cada operación
- CA-002-02: aislamiento concurrente A/B sobre mismo pool (-race)
- CA-002-03: sin residuo de contexto tras devolver conexión al pool
- CA-002-04: callback nunca ejecutado si falla set_config
- CA-002-05: dominio no importa database ni pgx (testdeps_test.go)
- CA-002-06: cancelación de context propaga a PostgreSQL

## Migraciones

Gobernadas por Atlas CLI v1.3.0 (fijado, no `latest`). Ver `database/README.md`.

```bash
atlas migrate hash --dir file://database/migrations
atlas migrate validate --env local
atlas migrate apply --env local
```

`atlas.sum` se versiona. NO se ejecutan al arrancar la aplicación (CA-001-07).