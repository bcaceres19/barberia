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
- **Idempotencia (HU-004)**: `internal/platform/idempotency` coordina con
  las funciones PostgreSQL ya aprobadas (`DEC-043`) para que una escritura
  crítica se ejecute como máximo una vez por clave. Reutilizable, sin
  endpoint propio todavía. Ver la sección "Patrón obligatorio: idempotencia
  reutilizable" más abajo.

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
go test -race ./internal/platform/idempotency/...
go test -race ./internal/platform/httpserver/...
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

## Patrón obligatorio: idempotencia reutilizable (HU-004)

RN-IDE-01 exige que **toda escritura crítica futura** (crear una cita,
reservar un turno, cualquier operación que un reintento de red podría
duplicar) use este mecanismo. `internal/platform/idempotency` no importa
`net/http` ni Chi: es un puerto de aplicación (`Coordinator`) con un
adaptador PostgreSQL concreto (`SQLCoordinator`) que llama exactamente a
`idempotency_begin`, `idempotency_complete` e `idempotency_abort`
(`database/migrations/20260811154100_harden_idempotency_concurrency.sql`,
`DEC-043`), sin duplicar en Go ninguna decisión que esas funciones ya
garantizan de forma atómica.

HU-004 entrega el mecanismo, **no** una operación de negocio: no hay
endpoint de idempotencia en el router ni en el contrato OpenAPI. Cada HU
que agregue una escritura crítica real es quien monta este patrón sobre su
propio handler.

### Regla que no se puede romper: una sola transacción

`Coordinator.Begin`, la ejecución del efecto y `Complete` o `Abort` deben
ocurrir dentro de la **misma** `database.DB.InTenantTx`. El motivo no es
estilo: el bloqueo que toma `Begin` (`pg_try_advisory_xact_lock`) tiene
alcance transaccional y se libera solo al terminar esa transacción (COMMIT
o ROLLBACK). Repartir `Begin` y el efecto en transacciones separadas anula
la protección de concurrencia que `DEC-043` exige.

### Cómo se usa

```go
import (
    "system-barbershop/internal/platform/database"
    "system-barbershop/internal/platform/httpserver"
    "system-barbershop/internal/platform/idempotency"
)

func (s *MiServicio) handler(w http.ResponseWriter, r *http.Request) {
    requestID := httpserver.RequestIDFromContext(r.Context())

    // 1. Cabecera: obligatoria, validada antes de tocar PostgreSQL.
    key, err := httpserver.IdempotencyKeyFromRequest(r)
    if err != nil {
        httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
        return
    }

    // 2. Huella canónica del contenido EXACTO recibido (nunca una
    //    re-serialización de un JSON ya decodificado).
    body, _ := io.ReadAll(r.Body)
    fingerprint := httpserver.IdempotencyFingerprint(r, body)

    var decision idempotency.Decision
    var respuesta idempotency.StoredResponse

    err = s.db.InTenantTx(r.Context(), shop, func(ctx context.Context, q database.Queries) error {
        d, err := s.coord.Begin(ctx, q, shop, key, "create_appointment", fingerprint, 24*time.Hour)
        if err != nil {
            return err
        }
        decision = d
        if d.Outcome != idempotency.OutcomeProceed {
            return nil // no ejecutar el efecto; el handler decide fuera de la tx
        }

        // 3. El efecto real ocurre AQUÍ, con q, dentro de esta misma tx.
        respuesta = ejecutarElEfecto(ctx, q)

        // 4. Confirmar solo tras el éxito del efecto.
        _, err = s.coord.Complete(ctx, q, shop, key, respuesta)
        return err
    })
    if err != nil {
        // El error del efecto (o su pánico, o la cancelación del contexto)
        // hace ROLLBACK de TODA la transacción, incluido el INSERT que
        // Begin hizo para reclamar la clave: un reintento legítimo
        // encuentra la clave libre de nuevo, sin llamar Abort a mano.
        httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
        return
    }

    switch decision.Outcome {
    case idempotency.OutcomeProceed:
        httpserver.WriteStoredResponse(w, respuesta) // primera ejecución
    case idempotency.OutcomeReplay:
        httpserver.WriteStoredResponse(w, decision.Response) // repetición exacta (CA-004-01)
    default:
        // conflicto de contenido/operación (RN-IDE-01) o ejecución en
        // curso (DEC-043): Translate ya sabe convertirlo en 409.
        httpserver.WriteProblem(w, httpserver.Translate(decision.AsError(), requestID))
    }
}
```

### Por qué no hace falta un `recover()` propio

Un pánico dentro del efecto se propaga a través de `Begin`/efecto/`Complete`
hasta el `defer` de `InTenantTx`, que sigue ejecutándose durante el
desenrollado del pánico y hace `ROLLBACK` igual que ante un error normal
(deshaciendo también el `INSERT` de `Begin`). Este paquete no necesita —ni
agrega— su propia recuperación de pánico: sería una garantía duplicada
sobre otra que `InTenantTx` ya ofrece.

### Cuándo sí llamar `Abort` explícitamente

Casi nunca hace falta: el `ROLLBACK` automático ante cualquier error ya deja
la clave libre para un reintento legítimo (`CA-004-06`). `Abort` existe
para el caso distinto en que un módulo necesita **conservar** el resto de
esa transacción (por ejemplo, dejar evidencia de auditoría del fallo) y
solo limpiar la reclamación de idempotencia antes de hacer `COMMIT`.
Decidir si ese caso aplica es responsabilidad del módulo que llama, no de
`internal/platform/idempotency`: este paquete no inventa esa política.

### Reglas que ya cumple el adaptador (verificadas contra PostgreSQL real)

- `CA-004-01`: misma clave y mismo contenido reproducen la respuesta
  original byte a byte, sin repetir el efecto.
- `CA-004-02`: misma clave con contenido distinto responde
  `409 idempotency-conflict`, sin ejecutar nada.
- `CA-004-03`: dos conexiones reales y simultáneas con la misma clave — el
  efecto ocurre exactamente una vez; la perdedora recibe
  `409 idempotency-locked` de inmediato, sin esperar (`DEC-043`, sin
  `lock_timeout` ni espera acotada).
- `CA-004-04`: la misma clave literal en dos barberías no interfiere (RLS
  tenant-aware, igual que el resto del esquema).
- `CA-004-05`: una clave vencida se trata como si no existiera; no revive
  una respuesta antigua.
- `CA-004-06`: un efecto fallido, con pánico o con contexto cancelado nunca
  deja bloqueado un reintento legítimo.
- Una fila `completed` nunca se aborta (cierra `DDL-IDEM-01`).

Ver `internal/platform/idempotency/postgres_test.go` y
`internal/platform/httpserver/idempotency_test.go` para la evidencia
completa, incluida la concurrencia real con `-race`.

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

`internal/platform/idempotency/postgres_test.go` y
`internal/platform/httpserver/idempotency_test.go` requieren el mismo
PostgreSQL real (misma `TEST_DATABASE_URL`, mismo `barberia_app`, mismas
migraciones y testdata). Cubren `CA-004-01` a `CA-004-06`, `RN-IDE-01`
(clave reutilizada con otra operación), el estado `conflict_in_progress`, y
`CA-004-03` con dos conexiones reales y simultáneas del pool bajo `-race`.

```bash
go test -race ./internal/platform/idempotency/...
go test -race ./internal/platform/httpserver/...
```

## Migraciones

Gobernadas por Atlas CLI v1.3.0 (fijado, no `latest`). Ver `database/README.md`.

```bash
atlas migrate hash --dir file://database/migrations
atlas migrate validate --env local
atlas migrate apply --env local
```

`atlas.sum` se versiona. NO se ejecutan al arrancar la aplicación (CA-001-07).