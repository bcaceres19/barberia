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
- **Inicio de sesión (HU-005)**: `internal/modules/auth` implementa
  `POST /api/v1/public/auth/login` (`DEC-055`, sin middleware de
  autenticación). Ver la sección "Inicio de sesión (HU-005)" más abajo.
- **Sesión persistente y cierre de sesión (HU-006)**: middleware de sesión
  (paso 8 del orden de middleware) montado sobre el subrouter real de
  `/api/v1/private` que `httpserver.NewRouter` devuelve, renovación
  deslizante de 30 días y `POST /api/v1/private/auth/logout`, el primer
  endpoint privado real del sistema. Ver la sección "Sesión persistente y
  cierre de sesión (HU-006)" más abajo.
- **Contexto de sesión (HU-012)**: `GET /api/v1/private/auth/session`
  (`DEC-060`), segunda operación privada real, de solo lectura: rehidrata la
  cookie `HttpOnly` y devuelve la barbería activa. Ver la sección "Contexto
  de sesión (HU-012)" más abajo.
- **Defensa escalonada contra abuso (HU-007)**: conteo por IP de
  `POST /api/v1/public/auth/login` (`DEC-061`: la sexta solicitud dentro de
  la ventana exige completar el reto telefónico, no la quinta) y
  `POST /api/v1/public/auth/challenge`/`.../verify` (`DEC-062`). Purga
  periódica en `cmd/worker`. Ver la sección "Defensa escalonada contra abuso
  (HU-007)" más abajo.
- **Recuperación de acceso (HU-008)**: `POST /api/v1/public/auth/recovery/
  request`/`.../verify`/`.../reset-password` (`DEC-063`–`DEC-066`): código
  de un solo uso enviado por WhatsApp (Meta Cloud API) y correo (Resend),
  sin enumeración, con token de reinicio opaco y revocación total de
  sesiones al cambiar la contraseña. Purga periódica en `cmd/worker`. Ver la
  sección "Recuperación de acceso (HU-008)" más abajo.
- **Configuración básica de la barbería (HU-020)**: `internal/modules/shops`
  implementa `GET`/`PATCH /api/v1/private/settings/barbershop`: nombre, zona
  IANA (confirmada contra `pg_timezone_names` dentro de la misma transacción
  del `UPDATE`) y contacto opcional. Ver la sección "Configuración básica de
  la barbería (HU-020)" más abajo.
- **Registro y listado de barberos (HU-021)**: `internal/modules/staff`
  implementa `GET`/`POST /api/v1/private/barbers` y
  `GET`/`PATCH /api/v1/private/barbers/{barberId}`: lista paginada por
  cursor, alta protegida con `Idempotency-Key` (RN-IDE-01, `DEC-043`),
  lectura individual y renombrado. `404` idéntico para un barbero
  inexistente o de otra barbería (`CA-021-05`); sin borrado, desactivación,
  orden manual ni vínculo con `staff_user` (`DEC-047`). Ver la sección
  "Registro y listado de barberos (HU-021)" más abajo.
- **Catálogo básico de servicios (HU-022)**: `internal/modules/catalog`
  implementa `GET`/`POST /api/v1/private/services` y
  `GET`/`PATCH /api/v1/private/services/{serviceId}`: lista paginada por
  cursor, alta protegida con `Idempotency-Key` (RN-IDE-01, `DEC-043`),
  lectura individual y edición parcial de nombre/descripción/duración/
  precio. Precio en centavos enteros (`ParsePriceCOP`/`FormatPriceCOP`,
  nunca coma flotante); moneda COP fija, sin campo editable (`DEC-067`).
  `409` con `code: "conflict"` para un nombre ya usado por otro servicio
  activo de la misma barbería, distinto del `409` de idempotencia; `404`
  idéntico para un servicio inexistente o de otra barbería (`CA-022-06`);
  sin borrado, ciclo de activación ni vínculo con barberos (fuera de
  alcance de HU-022). Ver la sección "Catálogo básico de servicios
  (HU-022)" más abajo.
- **Asignación de servicios a barberos (HU-023)**: `internal/modules/
  catalog` (`AssignmentService`, dueño de la operación) implementa
  `GET /api/v1/private/barbers/{barberId}/services`,
  `PUT`/`DELETE .../{barberId}/services/{serviceId}`: lista paginada por
  cursor, asignación con semántica HTTP naturalmente repetible (sin
  `Idempotency-Key`: repetir el mismo `PUT` responde `200` en vez de `201`
  con el mismo `createdAt`, sin crear una segunda fila) y desasignación.
  Retirar la última asignación activa de un servicio activo responde `409`
  (`DEC-068`), verificado dentro de la misma transacción que bloquea la
  fila de `service` (`SELECT ... FOR UPDATE`) para resistir la carrera de
  dos desasignaciones concurrentes de las dos últimas filas de un mismo
  servicio. `catalog` colabora con `staff` SOLO a través de
  `catalog.BarberPort`, un puerto pequeño que `staff.BarberLookup`
  satisface de forma estructural: ningún paquete de un módulo importa al
  otro (verificado por `internal/platform/archtest/
  module_boundary_test.go`); `cmd/api` es la única raíz de composición que
  conecta ambos. `404` idéntico para un barbero o servicio inexistente o de
  otra barbería (`CA-023-04`); la asociación (`barber_service`) nunca
  duplica nombre, duración, precio ni estado (`CA-023-07`). Ver la sección
  "Asignación de servicios a barberos (HU-023)" más abajo.
- **Ciclo de vida de servicios (HU-024)**: `internal/modules/catalog`
  implementa `GET .../{serviceId}/deactivation-impact`,
  `POST .../deactivate`, `POST .../reactivate`, protegidas por el mismo
  protocolo de idempotencia que HU-004/HU-022 (`Idempotency-Key`,
  RN-IDE-01). El impacto de citas futuras es literal `0` en B1
  (`catalog.currentDeactivationImpact`, `DEC-069`): `appointment` no existe
  todavía, así que no se simula ni se construye un puerto que finja
  consultarlo. La transición bloquea la fila (`SELECT ... FOR UPDATE`)
  dentro de la misma transacción del `UPDATE`, mismo patrón que HU-023
  (`DEC-068`): una clave de idempotencia nueva sobre una transición que ya
  no aplica responde `409` sin romper la reclamación. Sin migración nueva:
  `service.is_active`/`deactivated_at` ya existían desde HU-022. Ver la
  sección "Ciclo de vida de servicios (HU-024)" más abajo.
- **Horario laboral recurrente (HU-040)**: `internal/modules/schedule`
  implementa `GET`/`POST /api/v1/private/barbers/{barberId}/working-hours` y
  `GET`/`PATCH`/`DELETE .../working-hours/{workingHourId}`: lista paginada
  por cursor ordenada por día ISO y hora de inicio, alta protegida con
  `Idempotency-Key` (RN-IDE-01, `DEC-043`), lectura individual, edición
  (reemplaza el intervalo completo) y retiro físico. Un tramo cuyo
  `startsTime` + `durationMinutes` cruza medianoche es válido (`DEC-020`);
  el solape con otro tramo del mismo barbero y día responde `409`
  (`code: "conflict"`), verificado dentro de la transacción que bloquea la
  fila de `barber` (`SELECT ... FOR UPDATE`) para resistir la carrera de
  dos altas concurrentes que parten de cero tramos existentes. `schedule`
  colabora con `staff` SOLO a través de `schedule.BarberPort`, mismo
  patrón que `catalog.BarberPort` frente a HU-023. `CT-008` se resolvió
  como `DEC-070`: la FK de `working_hour` hacia `barber` usa
  `ON DELETE RESTRICT`, no `CASCADE`. Ver la sección "Horario laboral
  recurrente (HU-040)" más abajo.
- **Excepciones de jornada y festivos (HU-041)**: extiende
  `internal/modules/schedule` con el interruptor de calendario colombiano
  de festivos por barbero (`GET`/`PATCH .../holiday-calendar`), el CRUD de
  excepciones de jornada por fecha (`GET`/`POST .../schedule-exceptions`,
  `GET`/`PATCH`/`DELETE .../schedule-exceptions/{exceptionId}`, alta
  protegida con `Idempotency-Key`) y el dato de referencia
  `GET /api/v1/private/schedule/colombian-holidays` (calculado, algoritmo
  de Pascua Meeus/Jones/Butcher + Ley 51 de 1983 "Ley Emiliani", sin tabla
  ni proveedor externo). A diferencia de HU-040, el solape de tramos SÍ
  vive en una restricción `EXCLUDE` de PostgreSQL
  (`working_hour_override_segment_no_overlap_excl`): la fecha es fija por
  cabecera, no una envolvente semanal recurrente, así que no hay
  fragilidad de tramo nocturno que evitar. `CT-008`/`DEC-070` (`ON DELETE
  RESTRICT`) aplica igual a las dos tablas nuevas. `schedule.
  ResolveEffectiveDay` es un puerto interno (nunca expuesto por HTTP) que
  aplica la precedencia de `CA-041-07`: excepción manual (abierta o
  cerrada) > festivo automático (solo si el barbero activó su calendario)
  > horario semanal de HU-040. Ver la sección "Excepciones de jornada y
  festivos (HU-041)" más abajo.
- **Núcleo persistente de citas (HU-060)**: nuevo módulo
  `internal/modules/booking` entrega únicamente la base de datos y la
  primitiva transaccional interna `BookingService.CreateInternal`: crea o
  vincula el cliente que el llamador ya decidió, inserta la cita
  `confirmed` y el evento `appointment_created`, todo dentro de UNA sola
  `InTenantTx`. Sin endpoint HTTP, sin operación OpenAPI y sin pantalla:
  `HU-061` expondrá la creación manual sobre este núcleo. La restricción
  `EXCLUDE USING gist (barbershop_id, barber_id, tstzrange(starts_at,
  ends_at, '[)')) WHERE (occupies_schedule)` es la última defensa contra
  cruces del mismo barbero (RN-CON-01/RN-CON-03); probada con inserción SQL
  directa (`database/tests/hu060_citas.sql`) y con una carrera real de dos
  conexiones en Go
  (`TestCreateInternal_ConcurrentOverlap_ExactlyOneSucceeds`,
  `internal/modules/booking/postgres`, mismo patrón que la carrera de
  HU-023). Bajo concurrencia real, PostgreSQL puede resolver la disputa del
  índice GiST como `deadlock_detected` (40P01) en vez de
  `exclusion_violation` (23P01) limpio para la transacción perdedora: el
  repositorio traduce ambos códigos al mismo conflicto tipado
  (`bookingpostgres.isDeadlockDetected`, hallazgo real verificado contra
  PostgreSQL 14, no una hipótesis). `appointment_history`/
  `appointment_history_change` son append-only para `barberia_app` (sin
  `UPDATE` ni `DELETE`, RN-HIS-02). `booking` no importa `schedule` ni
  `catalog`, y viceversa. Ver la sección "Núcleo persistente de citas
  (HU-060)" más abajo.

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
go test -race ./internal/modules/auth/...
go test -race ./internal/modules/notification/...
go test -race ./internal/modules/shops/...
go test -race ./internal/modules/staff/...
go test -race ./internal/modules/catalog/...
go test -race ./internal/modules/schedule/...
go test -race ./internal/modules/booking/...
go test -race ./cmd/api/...
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

## Inicio de sesión (HU-005)

`POST /api/v1/public/auth/login` (`DEC-055`, sin middleware de
autenticación: sería circular, `CT-003`) verifica correo y contraseña de un
barbero activo y emite una cookie de sesión opaca de 30 días (`DEC-050`).
`internal/modules/auth` sigue la estructura estándar de un módulo: el
núcleo (`domain.go`, `ports.go`, `service.go`, `password.go`, `token.go`) no
importa Chi, `net/http`, pgx ni `internal/platform/database` (CA-002-06);
`postgres/repository.go` traduce el puerto `auth.Repository` a
`database.DB`; `httpapi/` decodifica, valida la forma y traduce el
resultado a HTTP.

### Contraseñas: argon2id, parámetros documentados en código

`internal/modules/auth/password.go` (`Argon2Hasher`) deriva la contraseña
con `golang.org/x/crypto/argon2` en modo `argon2id`, memoria 19 MiB, 2
iteraciones, 1 hilo (recomendación mínima de OWASP Password Storage Cheat
Sheet), sal de 16 bytes generada con `crypto/rand` en cada `Hash`. El valor
codificado (formato PHC, `$argon2id$v=19$m=...,t=...,p=...$sal$hash`)
incluye algoritmo, versión, parámetros, sal y hash en un único texto:
`staff_credential` no tiene columna de sal separada (CA-005-03). Nunca se
registra, nunca se expone en una respuesta.

### No enumeración: correo inexistente = contraseña incorrecta = usuario inactivo (CA-005-02, CA-005-07)

`authn_resolve_login_tenant` (SQL) ya unifica "correo inexistente" y
"usuario inactivo" en el mismo resultado (`NULL`). `LoginService.Login`
(`service.go`) SIEMPRE ejecuta exactamente una llamada a
`PasswordHasher.Verify`, con un hash del mismo costo criptográfico —el real
si la credencial existe, o un hash señuelo precalculado una sola vez al
construir el servicio si no— y `Repository.LookupCredential`
(`postgres/repository.go`) ejecuta la MISMA forma de consultas SQL en
ambos casos (tenant real o tenant señuelo `00000000-...-000000000000`, con
un `staff_user_id` señuelo cuando no hay usuario que buscar). El resultado
observable es idéntico: mismo `401`, mismo `type`/`code`/`detail`, mismos
headers. La evidencia de esta simetría es **estructural** (mismo número de
llamadas a cada dependencia, mismo costo de hash), no una medición de
nanosegundos: ver `TestLogin_StructuralNonEnumeration_SameShapeForUnknownEmailAndWrongPassword`
en `internal/modules/auth/service_test.go`.

### Cookie de sesión: DP-SEG-07 (elección provisional)

`DEC-050` fija `HttpOnly`+`Secure`+`SameSite`, pero no el valor exacto de
`SameSite`, `Path`, `Domain` ni el nombre de la cookie; tampoco lo fija el
issue `#44`. `docs/00-control/dudas-pendientes.md` registra este vacío como
`DP-SEG-07` y documenta la elección aplicada mientras se confirma:
`httpapi.DefaultCookieConfig()` usa nombre `barberia_session`,
`Path=/api/v1`, `SameSite=Lax`, `Secure=true`, sin `Domain` explícito
(host-only), 30 días de vigencia. **Estos valores son provisionales**; si
el propietario confirma otros, se actualizan aquí, en
`api/openapi/components/security-schemes/SessionCookie.yaml` y en el
código en el mismo cambio.

### Bloqueo conocido: CA-005-05 y la parte de "solicitudes privadas posteriores" de CA-005-01

`CA-005-05` exige verificar el aislamiento de sesión "contra un endpoint
privado real". Ninguna operación privada existe todavía en el código
(`/api/v1/private` está montado sin operaciones desde HU-003) ni está
aprobada por ninguna fuente para HU-005: la primera operación privada real
(`POST /api/v1/private/auth/logout`) pertenece a HU-006, fuera del alcance
de este cambio. Siguiendo la instrucción explícita del prompt de HU-005
("si ninguna fuente define esa operación, registra la duda y detente; no
publiques una ruta de prueba"), esta parte queda **registrada como
`DP-SEG-08`** en `docs/00-control/dudas-pendientes.md`, sin un endpoint de
demostración inventado. Como evidencia parcial,
`internal/modules/auth/postgres/repository_test.go` y
`database/tests/hu005_aislamiento_credenciales_sesiones.sql` demuestran el
aislamiento de `staff_session` por tenant contra PostgreSQL real (RLS,
`SELECT`/`INSERT` cruzados rechazados), pero no hay todavía una prueba HTTP
de extremo a extremo con un endpoint privado real.

### Tabla de criterios de aceptación

| Criterio | Estado | Prueba o evidencia |
| --- | --- | --- |
| `CA-005-01` | Cumplido | Emisión de sesión asociada a usuario y barbería (`TestLogin_Success_CreatesSessionWithHashedToken`, `TestLoginHandler_Success_SetsCookieWithoutTokenInBody`, `TestCreateSession_PersistsRetrievableSessionScopedToTenant`). "Solicitudes privadas posteriores operan con ese contexto": completado por `HU-006` (`TestSession_HTTP_ReusedCookieOnFreshRequest_StaysAuthenticatedWithoutCredentials`, `DEC-058`). |
| `CA-005-02` | Cumplido | `TestLogin_StructuralNonEnumeration_SameShapeForUnknownEmailAndWrongPassword`, `TestLoginHandler_UnknownEmail_ReturnsIdenticalProblemToWrongPassword`, `TestLookupCredential_UnresolvedTenant_UsesDecoyWithoutError`, sección "CA-005-01/02/07" de `hu005_aislamiento_credenciales_sesiones.sql`. |
| `CA-005-03` | Cumplido | `password.go` (argon2id, parámetros OWASP, sal `crypto/rand` por llamada), `TestArgon2Hasher_TwoUsersSamePassword_ProduceDistinctEncodedValues`, `staff_credential_password_hash_ck` + sección "Formato" de la suite SQL. |
| `CA-005-04` | Cumplido | `TestLoginHandler_ThroughFullRouter_NeverLogsSensitiveValues`, `TestLoginHandler_Success_SetsCookieWithoutTokenInBody`, `TestLoginHandler_InternalRepositoryError_ReturnsSafe500`. |
| `CA-005-05` | Cumplido | Aislamiento de `staff_session` por tenant demostrado contra PostgreSQL real (`TestLookupCredential_CrossTenant_NeverLeaksAcrossShops`, `TestCreateSession_PersistsRetrievableSessionScopedToTenant`, sección RLS de la suite SQL). La verificación end-to-end "contra un endpoint privado real" la completa `HU-006` (`CA-006-07`, `TestLogout_HTTP_SessionOfShopA_NeverExecutesShopBsLogout`, `DEC-058`). |
| `CA-005-06` | Cumplido | `api/openapi/paths/public-auth.yaml` escrito antes del handler; `openapi:lint`/`openapi:bundle` en verde; `contract_test.go` (4 pruebas) compara los DTO Go y la operación real contra el YAML fuente. |
| `CA-005-07` | Cumplido | "No puede iniciar sesión": `TestResolveLoginTenant_InactiveAndUnknown_BothReturnNotFound`, `TestLogin_InactiveUser_IsIndistinguishableFromUnknownEmail`, `TestLoginHandler_InactiveUser_SameProblemAsUnknownEmail`, sección CA-005-07 de la suite SQL. "No conserva sesiones vigentes" tras una desactivación posterior a la emisión: cerrado por `HU-006` (`TestValidateAndRenewSession_InactiveUser_NeverRenews`), que reconfirma `is_active` en cada uso de la sesión. |

## Sesión persistente y cierre de sesión (HU-006)

`internal/modules/auth` extiende el módulo de `HU-005` con la validación de
una sesión ya emitida, su renovación deslizante y su cierre. El núcleo
(`session_service.go`) sigue sin importar Chi, `net/http`, pgx ni
`internal/platform/database` (CA-002-06); `postgres/session_repository.go`
implementa el puerto `auth.SessionRepository` contra `database.DB`;
`httpapi/middleware.go` y `httpapi/logout_handler.go` son la capa HTTP. No se
agregó ninguna migración: `staff_session` y la función
`authn_resolve_session_tenant` ya existían desde
`20260813120000_create_auth_credentials_and_sessions.sql` (`HU-005`), sin
consumidor hasta ahora.

### Middleware de sesión: dónde vive y cómo se monta

`httpserver.NewRouter` ahora devuelve dos valores: el `*chi.Mux` completo y
el subrouter real montado en `/api/v1/private`. `cmd/api.buildRouter` monta
`SessionMiddleware.RequireSession` con `private.Use(...)` **antes** de
registrar ninguna ruta sobre ese subrouter (chi exige ese orden) y toda ruta
privada futura se registra sobre él, nunca con el patrón completo sobre el
`*chi.Mux` (eso saltaría el middleware: ver el control negativo
`TestPrivateSubrouter_BypassingItSkipsTheMiddleware` en
`internal/platform/httpserver/private_router_wiring_test.go`). Es el paso 8
del orden de middleware (`docs/04-arquitectura/backend-go.md` sección 6);
`DEC-055` ya dejó el login fuera de este subrouter, así que no existe
ninguna excepción de arranque sin sesión que mantener (`CT-003` resuelta).

### Validación con renovación deslizante, en una única sentencia atómica

`SessionRepository.ValidateAndRenewSession` reconfirma, dentro de la
transacción tenant-aware (la resolución previa con
`authn_resolve_session_tenant` NUNCA sustituye esta autorización final):
`revoked_at IS NULL`, `expires_at > now` y `staff_user.is_active`, y si la
sesión sigue vigente extiende `last_used_at`/`expires_at` a 30 días desde el
uso (`DEC-050`) en la MISMA sentencia `UPDATE ... FROM ... RETURNING`. No
hace falta un `SELECT ... FOR UPDATE` aparte: el `UPDATE` ya toma el bloqueo
de fila necesario, así que una renovación y una revocación concurrentes
sobre la misma sesión quedan serializadas por PostgreSQL sin ninguna
coordinación adicional en Go, y el resultado final siempre queda revocado
(ver `TestValidateAndRenewSession_ConcurrentWithRevoke_FinalStateAlwaysRevoked`).

### Cierre de sesión: solo la sesión actual, efecto idempotente

`POST /api/v1/private/auth/logout` fija `revoked_at` únicamente en la fila
de la sesión que el middleware ya validó para esa solicitud
(`auth.Principal.SessionID`, nunca un identificador que llegue del cliente:
el contrato no acepta ninguno) y limpia la cookie con los mismos
`Path`/`SameSite`/`Secure`/`HttpOnly` con que `HU-005` la emitió. Repetir la
llamada con el mismo material ya revocado nunca vuelve a ejecutar la
operación: el middleware la detiene con el mismo `401` uniforme antes de
llegar al handler (`CA-006-02`).

### `CA-006-07`: aislamiento entre barberías contra el logout real

`DEC-058` dividió la verificación de `CA-005-05`/`CA-005-01` de `HU-005`
entre PostgreSQL/RLS (esa historia) y un endpoint privado real (esta). Como
el contrato de logout no acepta ningún identificador de sesión objetivo -el
tenant y la sesión se derivan exclusivamente de la cookie ya autenticada-,
no existe ningún canal por el que la sesión de una barbería pudiera afectar
la de otra; `TestLogout_HTTP_SessionOfShopA_NeverExecutesShopBsLogout`
(`apps/api/cmd/api/session_integration_test.go`) lo demuestra contra el
router de producción con dos tenants reales, y
`TestRevokeSession_WrongTenant_NeverRevokesAnotherShopsSession`
(`internal/modules/auth/postgres/session_repository_test.go`) añade la
defensa en profundidad a nivel de repositorio (RLS + filtro explícito de
`barbershop_id`).

### Tabla de criterios de aceptación

| Criterio | Estado | Prueba o evidencia |
| --- | --- | --- |
| `CA-006-01` | Cumplido | `TestSession_HTTP_ReusedCookieOnFreshRequest_StaysAuthenticatedWithoutCredentials`: una cookie persistida reutilizada en una solicitud completamente nueva sigue autenticada dentro de la vigencia, sin reenviar credenciales. |
| `CA-006-02` | Cumplido | `TestLogout_HTTP_ReusedCookieAfterLogout_Returns401AndNeverRunsTwice`, `TestValidateAndRenewSession_RevokedSession_NeverRenewsOrReopens`, `TestRevokeSession_RevokesOwnSession_AndIsIdempotent`. |
| `CA-006-03` | Cumplido | `TestValidateAndRenewSession_ExpiredSession_NeverRenewsOrReopens` (la fila permanece sin cambios; una sesión vencida exige un nuevo inicio de sesión). |
| `CA-006-04` | Cumplido | `TestPrivateRouteInventory_AllRegisteredRoutesRequireSession` camina el router REAL de producción con `chi.Walk` (sin lista manual) y confirma `401` uniforme para toda ruta bajo `/api/v1/private`; `TestPrivateSubrouter_BypassingItSkipsTheMiddleware` demuestra que la técnica detecta una ruta que escapara del middleware. |
| `CA-006-05` | Cumplido | `TestSessionMiddleware_MalformedCookie_Returns401WithoutTouchingRepository`, `TestSessionMiddleware_OversizedCookie_Returns401WithoutTouchingRepository` (forma inválida rechazada antes de tocar PostgreSQL), `TestSessionMiddleware_ThroughFullRouter_NeverLogsSessionMaterial`, `TestPrivateRoute_ThroughFullRouter_NeverLogsSessionMaterial` (ni el token ni su hash aparecen en logs ni en la URL: la cookie es el único transporte). |
| `CA-006-06` | Cumplido | `TestLogout_HTTP_ClosingOneDeviceDoesNotAffectAnother`: dos sesiones del mismo usuario, cerrar una conserva la otra. |
| `CA-006-07` | Cumplido | `TestLogout_HTTP_SessionOfShopA_NeverExecutesShopBsLogout` con dos tenants reales contra el logout real; `TestRevokeSession_WrongTenant_NeverRevokesAnotherShopsSession` como defensa en profundidad a nivel de repositorio. Completa `CA-005-05`/`CA-005-01` de `HU-005` (`DEC-058`). |

## Contexto de sesión (HU-012)

`GET /api/v1/private/auth/session` (`DEC-060`) resuelve `DP-SEG-09`: hasta
esta historia no existía ninguna lectura privada no destructiva para que el
frontend rehidratara la cookie `HttpOnly` al abrir o recargar la
aplicación. Se registra sobre el mismo subrouter `/api/v1/private`, después
del mismo `SessionMiddleware` que ya protege `logout` — sin validación de
sesión duplicada ni un segundo camino de autorización.

### Diferencia con logout: lectura, no mutación

A diferencia de `POST /private/auth/logout`, este endpoint nunca escribe
`Set-Cookie` ni cambia estado: repetirlo no tiene efecto adicional.
`SessionService.Context` solo compone el payload mínimo a partir del
`auth.Principal` ya validado y renovado por el middleware para esa misma
solicitud (`ExpiresAt` es el mismo valor que la renovación deslizante ya
fijó, sin una segunda consulta) y una lectura nueva y estrecha,
`SessionRepository.BarbershopName`, que lee la columna real `barbershop.name`
(existente desde `20260807170000_create_tenant_foundation.sql`) dentro de
la misma transacción tenant-aware, aislada por
`barbershop_select_tenant_policy` (RLS, `DEC-024`).

### Payload mínimo (RN-DAT-02)

`{ barbershop: { id, name }, expiresAt }`. Nunca incluye `staffUserID`,
correo ni nombre del barbero: ningún criterio de `HU-012` lo exige, y
ampliarlo rompería el patrón de `auth.Principal` (identificadores opacos,
sin datos personales) que el resto del módulo ya sigue.

### Tabla de criterios de aceptación (evidencia de backend)

| Criterio | Estado | Prueba o evidencia |
| --- | --- | --- |
| `CA-012-01` (parcial, backend) | Cumplido | La restauración real contra la cookie se demuestra en `TestSessionContext_HTTP_ValidSession_ReturnsRealBarbershopName`; la parte de frontend (bootstrap tras cerrar/reabrir el navegador) se documenta en `apps/web/README.md`. |
| `CA-012-04` (parcial, backend) | Cumplido | `TestSessionContext_HTTP_ValidSession_ReturnsRealBarbershopName` y `TestSessionContext_HTTP_TwoTenants_NeverCrossesBarbershopNames` prueban que el nombre viene de la columna real y nunca se cruza entre tenants. |

### Pruebas

- Unitarias: `TestSessionService_Context_*` (`internal/modules/auth/session_service_test.go`).
- HTTP con doble de repositorio: `TestSessionContextHandler_*` (`internal/modules/auth/httpapi/session_context_handler_test.go`).
- Contrato: `TestContract_SessionContextOperation_*`, `TestContract_OpenAPIYAML_RegistersSessionContextPath` (`internal/modules/auth/httpapi/contract_session_test.go`).
- PostgreSQL real: `TestBarbershopName_*` (`internal/modules/auth/postgres/session_repository_test.go`).
- Router de producción con dos tenants reales: `TestSessionContext_HTTP_*` (`cmd/api/session_context_integration_test.go`), incluida ausencia de material de sesión en logs (`CA-006-05`) y no exposición del nombre tras revocar la sesión.

## Defensa escalonada contra abuso (HU-007)

Protege `POST /api/v1/public/auth/login` de intentos automatizados sin
enumerar cuentas ni bloquear prematuramente a un barbero legítimo
(`DEC-026`, `DEC-052`, `DEC-061`, `DEC-062`).

### Umbral: la sexta solicitud, no la quinta (`DEC-061`, resolvió `CT-005`)

`login_throttle_register_attempt` cuenta por IP (HMAC-SHA256 con
`APP_AUTH_HMAC_SECRET`, nunca en claro; única tabla del sistema sin
`barbershop_id` ni RLS por diseño — asociarla a una barbería permitiría
diluir el límite repartiendo intentos entre barberías). Con el umbral
inicial de 5, las cinco primeras solicitudes dentro de la ventana se
evalúan con normalidad (contraseña incluida); la **sexta** exige el reto
telefónico antes de evaluar la contraseña (`LoginService.Login` nunca
resuelve tenant ni llama `PasswordHasher.Verify` en esa rama). Responde
`429` (`code: challenge-required`) con cabecera `Retry-After`.

### Reto telefónico (`DEC-062`)

- `POST /api/v1/public/auth/challenge { email }` → siempre `202` con el
  mismo mensaje genérico, exista o no la cuenta, esté o no el teléfono
  verificado y esté o no la IP realmente escalada (no enumeración). Límite
  propio: 1 solicitud/60 s y máximo 3 activas por IP en 15 min
  (`auth_phone_challenge_request`, misma función que resuelve las tres
  condiciones y devuelve el teléfono solo en el camino aceptado).
- `POST /api/v1/public/auth/challenge/verify { email, code }` → código de
  6 dígitos (HMAC-SHA256 con el mismo secreto, nunca `SHA-256` simple:
  10⁶ combinaciones son triviales de recuperar offline sin un secreto),
  vigente 5 min, máximo 5 intentos, atado a la IP concreta que lo pidió.
  Éxito: `204`, limpia `escalated_until`/`attempt_count` de esa IP en la
  misma transacción (`auth_phone_challenge_verify`); el barbero reintenta
  el login normalmente, sin token adicional.
- El envío real de WhatsApp es un marcador de posición
  (`auth.LoggingPhoneCodeSender`, solo registra que "habría" enviado, sin
  `phone` ni `code`): el proveedor real es `DEC-066`, decisión de HU-008.

### Variables de entorno nuevas

| Variable | Por defecto | Uso |
| --- | --- | --- |
| `APP_AUTH_HMAC_SECRET` | (obligatoria, ≥32 caracteres) | Firma el HMAC de IP y código. Nunca en el repositorio. |
| `APP_TRUSTED_PROXIES` | vacío (sin proxies confiables) | CIDR separados por coma; sin esto, `X-Forwarded-For` se ignora siempre y se usa `RemoteAddr`. |
| `APP_LOGIN_THROTTLE_WINDOW_SECONDS` | `900` | Ventana deslizante del contador. |
| `APP_LOGIN_THROTTLE_ESCALATION_SECONDS` | `86400` | Duración del escalamiento una vez activado. |
| `APP_LOGIN_THROTTLE_THRESHOLD` | `5` | Solicitudes permitidas sin reto; la siguiente lo exige. |
| `APP_LOGIN_THROTTLE_RETENTION_SECONDS` | `172800` | Debe ser ≥ `..._ESCALATION_SECONDS`. |
| `APP_LOGIN_THROTTLE_PURGE_LIMIT` | `500` | Lote de purga del worker (1-1000). |
| `APP_PHONE_CHALLENGE_EXPIRES_SECONDS` | `300` | Vigencia del código. |
| `APP_PHONE_CHALLENGE_MAX_ATTEMPTS` | `5` | Intentos antes de invalidar. |
| `APP_PHONE_CHALLENGE_RATE_WINDOW_SECONDS` | `900` | Ventana del límite de solicitudes del reto. |
| `APP_PHONE_CHALLENGE_RATE_MAX_ACTIVE` | `3` | Máximo de solicitudes del reto por IP en esa ventana. |
| `APP_PHONE_CHALLENGE_RESEND_COOLDOWN_SECONDS` | `60` | Mínimo entre reenvíos. |
| `APP_PHONE_CHALLENGE_PURGE_LIMIT` | `500` | Lote de purga del worker (1-1000). |

### Prueba local completa: capturar el código sin un WhatsApp real

`APP_PHONE_CHALLENGE_CAPTURE_FILE=<ruta>` hace que
`auth.CapturingPhoneCodeSender` escriba `{"phone":"...","code":"..."}` en
esa ruta además de "enviar". **Restringido a `APP_ENVIRONMENT=local`/`test`
por `cmd/api/main.go`: arrancar con esta variable fuera de esos ambientes
falla explícitamente.** Es lo que usa `apps/web/e2e/reto-telefonico.spec.ts`
para completar el recorrido de verificación real sin un proveedor real
("terceros se interceptan").

⚠️ **Al correr la suite E2E completa en local**: `acceso.spec.ts`,
`panel.spec.ts` y `panel-evidencia-responsiva.spec.ts` no aíslan su IP y
comparten el contador real con cualquier otra prueba que use el mismo
`apps/api` local. Con `fullyParallel: true` y varios navegadores/proyectos,
el volumen combinado de intentos de esas suites puede superar el umbral
por defecto (5) y bloquearlas con un 429 falso, con un escalamiento de 24
horas real. Para correr esas suites junto con HU-007, sube el umbral
temporalmente en ESE proceso `apps/api` (p. ej.
`APP_LOGIN_THROTTLE_THRESHOLD=500`); `reto-telefonico.spec.ts` en cambio
necesita el valor real (por defecto, 5) para ejercitar el escalamiento de
verdad, así que corre contra una instancia separada de `apps/api` con la
configuración por defecto. `apps/web/e2e/reto-telefonico.spec.ts` aísla
cada una de sus propias pruebas con una IP sintética vía
`X-Forwarded-For` (requiere `APP_TRUSTED_PROXIES` incluyendo el peer real
que ve el proceso `api`, p. ej. `127.0.0.1/32` en un run nativo).

### Tabla de criterios de aceptación

| Criterio | Estado | Prueba o evidencia |
| --- | --- | --- |
| `CA-007-01` | Cumplido | `TestThrottleRepository_SixthAttempt_Escalates`, `TestLoginHandler_NotEscalated_ProceedsNormally`, `hu007_defensa_abuso.sql`, E2E "las primeras cinco...". |
| `CA-007-02` | Cumplido | `TestLoginHandler_Escalated_Returns429WithRetryAfterAndNeverTouchesRepository` (repositorio de login nunca se toca), `PhoneChallengeRepository_VerifyChallenge_CorrectCode_SucceedsAndClearsThrottle`, E2E "completar el reto... y llega a /panel". |
| `CA-007-03` | Cumplido | `TestChallengeHandler_AcceptedVsNotAccepted_IdenticalResponse`, `TestPhoneChallengeService_Verify_UnknownAccountAndWrongCode_SameError`, E2E "un código incorrecto...". |
| `CA-007-04` | Cumplido | Semántica de ventana/escalamiento en `login_throttle_register_attempt`, comentada y probada en `hu007_defensa_abuso.sql`. |
| `CA-007-05` | Cumplido | Todos los valores vienen de `config.Config`/variables de entorno; ver tabla arriba. Ningún valor incrustado en código. |
| `CA-007-06` | Cumplido | `TestPurgeRepository_PurgeLoginThrottle_DeletesOnlyExpired`, `TestPurgeRepository_PurgePhoneChallenges_RunsWithoutError`, purga real vía `barberia_worker` (sin GRANT a `barberia_app`, `hu007_defensa_abuso.sql`). |
| `CA-007-07` | Cumplido | `internal/platform/clientip` (paquete completo de pruebas: peer directo, proxy confiable, multi-salto, peer no confiable con cabecera falsificada, IPv4/IPv6, entradas malformadas). |

### Pruebas

- Unitarias: `internal/modules/auth/throttle_test.go`, `phone_challenge_test.go`, `purge_test.go`, `phone_sender_capture_test.go`.
- Resolvedor de IP: `internal/platform/clientip/resolver_test.go`.
- HTTP con dobles: `internal/modules/auth/httpapi/challenge_handler_test.go`, `login_throttle_test.go`, `contract_challenge_test.go`.
- PostgreSQL real, incluida concurrencia (45 llamadas simultáneas sin incrementos perdidos): `internal/modules/auth/postgres/throttle_repository_test.go`, `phone_challenge_repository_test.go`, `purge_repository_test.go` (este último requiere `TEST_WORKER_DATABASE_URL` conectado como `barberia_worker`).
- SQL directo con el rol real: `database/tests/hu007_defensa_abuso.sql`.
- E2E contra API/PostgreSQL/navegador reales: `apps/web/e2e/reto-telefonico.spec.ts` (ver advertencia de umbral arriba).

## Recuperación de acceso (HU-008)

`POST /api/v1/public/auth/recovery/request`, `.../verify` y
`.../reset-password` (`DEC-063`–`DEC-066`) permiten a un barbero recuperar
acceso sin intervención del propietario: solicitar un código de un solo
uso, verificarlo y establecer una contraseña nueva. `internal/modules/auth`
extiende el módulo con `recovery.go` (núcleo: `RecoveryService`,
`ValidateNewPassword`), `postgres/recovery_repository.go` (puerto contra las
cuatro funciones `SECURITY DEFINER` de
`20260817190000_create_staff_recovery_code.sql`) y `httpapi/recovery_handler.go`
(las tres operaciones HTTP). El nuevo módulo `internal/modules/notification`
aporta los adaptadores reales de entrega
(`notification/whatsapp_meta.go`, `notification/email_resend.go`,
`notification/recovery_sender.go`) detrás de un puerto pequeño
(`auth.RecoveryCodeSender`) que `auth` consume sin conocer Meta ni Resend.

### Mismo patrón que HU-007: resolver la cuenta antes de tenant

Igual que `login_throttle`/`auth_phone_challenge`, `staff_recovery_code` no
tiene ningún `GRANT` directo para `barberia_app` (`DDL-AUT-01`): las tres
operaciones públicas corren ANTES de resolver `app.barbershop_id`
(resuelven la cuenta por correo dentro de la función), así que usan
`database.DB.CallSecurityDefinerRow`, no `InTenantTx`. Solo
`auth_recovery_purge_expired` está concedida a `barberia_worker`
(`cmd/worker`, purga periódica).

### Código, token de reinicio y contraseña: parámetros de `DEC-064`/`DEC-063`

Código de 6 dígitos (`crypto/rand`, mismo generador que el reto telefónico
de HU-007), almacenado como `HMAC-SHA256(código, APP_AUTH_HMAC_SECRET)`
—nunca `SHA-256` simple—, vigente 15 min por defecto, máximo 5 intentos,
reenvío con cooldown de 60 s y máximo 3/hora que invalida atómicamente el
código anterior (`idx_staff_recovery_code_active_per_user`, único código
vigente por usuario). Verificar con éxito emite un token de reinicio opaco
de un solo uso (mismo patrón `CryptoTokenGenerator`/`HashToken` que la
sesión de HU-005/006), vigente 5 min por defecto, que `reset-password`
consume exactamente una vez incluso bajo dos solicitudes concurrentes
(`FOR UPDATE` en `auth_recovery_change_password`). La contraseña nueva debe
tener 10–128 caracteres, distinta del correo de la cuenta y de la
contraseña actual (`ValidateNewPassword`, mensaje específico por regla
incumplida, `CA-008-08`); reutiliza `Argon2Hasher` sin cambiar sus
parámetros.

### No enumeración (`DEC-065`) y destino enmascarado

`POST .../request` responde SIEMPRE `202` con el mismo mensaje genérico,
exista o no la cuenta, esté o no el teléfono verificado y sin importar el
resultado interno de `RecoveryService.Request` —ese resultado se descarta
para la respuesta y solo se registra sin destinatario ni código
(`RN-DAT-02`)—. El teléfono/correo enmascarados (`mask.go`) solo aparecen en
la respuesta `200` de `.../verify`: llegar ahí ya exige haber recibido y
transcrito el código real, así que no abre un oráculo nuevo. `.../verify` y
`.../reset-password` devuelven exactamente el mismo error uniforme para
código incorrecto, vencido, agotado, cuenta inexistente o token de reinicio
desconocido/de otra cuenta/consumido/vencido.

### Cambio de contraseña + revocación total de sesiones (`CA-008-05`)

`auth_recovery_change_password` actualiza `staff_credential` y revoca TODAS
las sesiones activas del usuario en la misma operación de PostgreSQL: no
hay ventana en la que la credencial ya cambió pero una sesión antigua siga
viva, ni viceversa.

### Proveedores reales: Meta WhatsApp Cloud API + Resend (`DEC-066`)

`DualChannelRecoverySender` intenta SIEMPRE los dos canales, sin importar si
uno falla (tolerancia a fallo parcial, sin cambiar la respuesta genérica de
`DEC-065`); cada adaptador aplica un timeout de 5 s sin reintento síncrono
dentro de la solicitud HTTP. Sin las credenciales de despliegue configuradas
(típicamente local/test), `cmd/api.buildRouter` usa
`auth.LoggingRecoveryCodeSender` (marcador de posición que solo registra que
"habría" enviado, sin teléfono/correo/código) en su lugar — mismo patrón que
el reto telefónico de HU-007.

| Variable | Por defecto | Uso |
| --- | --- | --- |
| `APP_RECOVERY_CODE_EXPIRES_SECONDS` | `900` | Vigencia del código. |
| `APP_RECOVERY_CODE_MAX_ATTEMPTS` | `5` | Intentos antes de invalidar. |
| `APP_RECOVERY_RESEND_COOLDOWN_SECONDS` | `60` | Mínimo entre reenvíos. |
| `APP_RECOVERY_RESEND_WINDOW_SECONDS` | `3600` | Ventana del límite de reenvío. |
| `APP_RECOVERY_RESEND_MAX_PER_WINDOW` | `3` | Máximo de códigos por cuenta en esa ventana. |
| `APP_RECOVERY_RESET_TOKEN_EXPIRES_SECONDS` | `300` | Vigencia del token de reinicio. |
| `APP_RECOVERY_CODE_PURGE_LIMIT` | `500` | Lote de purga del worker (1–1000). |
| `APP_META_WHATSAPP_API_VERSION` | `v21.0` | Versión de Meta Graph API. |
| `APP_META_WHATSAPP_PHONE_NUMBER_ID` | (vacía) | Secreto: número emisor en Meta. |
| `APP_META_WHATSAPP_ACCESS_TOKEN` | (vacía) | Secreto: autenticación contra Meta Graph API. |
| `APP_META_WHATSAPP_TEMPLATE_NAME` | (vacía) | Plantilla "Authentication" pre-aprobada. |
| `APP_META_WHATSAPP_LANGUAGE_CODE` | `es` | Idioma de esa plantilla. |
| `APP_RESEND_API_KEY` | (vacía) | Secreto: autenticación contra Resend. |
| `APP_RESEND_FROM_ADDRESS` | (vacía) | Remitente verificado del correo. |
| `APP_RESEND_SUBJECT` | `Código de recuperación de acceso` | Asunto fijo del correo. |

### Prueba local completa: capturar el código sin un proveedor real

`APP_RECOVERY_CAPTURE_FILE=<ruta>` hace que
`auth.CapturingRecoveryCodeSender` escriba `{"phone":"...","email":"...","code":"..."}`
en esa ruta además de "enviar". Restringido a `APP_ENVIRONMENT=local`/`test`
por `cmd/api/main.go`, mismo doble candado que
`APP_PHONE_CHALLENGE_CAPTURE_FILE` de HU-007. Es lo que usa
`cmd/api/recovery_integration_test.go` para completar el recorrido de
recuperación real sin un proveedor real ("terceros se interceptan"); el E2E
visual de tres pasos pertenece a `HU-011`.

### Contención entre paquetes de prueba: solo 2 cuentas con teléfono verificado

`testdata/hu007_reto_telefonico.sql` deja exactamente dos cuentas con
teléfono verificado en todo el fixture (`duena.a`, `dueno.b`), y tanto
`internal/modules/auth/postgres/recovery_repository_test.go` como
`cmd/api/recovery_integration_test.go` necesitan un código de recuperación
real sobre esas mismas cuentas. El cooldown de reenvío de `DEC-064` usa
reloj de pared real (no un reloj inyectable, a diferencia de las pruebas
unitarias), y `go test ./...` ejecuta paquetes distintos en paralelo contra
el MISMO PostgreSQL: sin cuidado adicional, dos paquetes podrían competir
por el cooldown de la misma cuenta. Ambos archivos reintentan (con un plazo
acotado, nunca indefinido) la primera solicitud "real" de cada prueba hasta
que se acepta, en vez de asumir éxito a la primera; ver el comentario junto
a `requestRecoveryEventually`/`doRecoveryRequestEventually` en cada archivo.
Esto no sustituye ninguna prueba determinista del cooldown en sí (esas ya
existen con datos sintéticos que no compiten por la cuenta compartida).

### Tabla de criterios de aceptación

| Criterio | Estado | Prueba o evidencia |
| --- | --- | --- |
| `CA-008-01` | Cumplido | No enumeración de `.../request` (`TestRecoveryRepository_RequestRecovery_UnverifiedPhone_NotAccepted`, `_UnknownAccount_NotAccepted`, `hu008_recuperacion_acceso.sql`, `TestRecovery_System_NonenumerationThenWrongCodeRejected`); entrega real a los dos adaptadores aprobados (`whatsapp_meta_test.go`, `email_resend_test.go` contra un `httptest.Server` que afirma el cuerpo/cabeceras de la solicitud real, `recovery_sender_test.go` para el fallo parcial tolerado). |
| `CA-008-02` | Cumplido | Intentos/agotamiento con reloj y generador deterministas ficticios (`recovery_test.go`), concurrencia real de PostgreSQL sin `sleep` (`hu008_recuperacion_acceso.sql`, `TestRecoveryRepository_ChangePassword_ConcurrentSameToken_OnlyOneWinner` para la parte de un solo consumo). |
| `CA-008-03` | Cumplido | Vencimiento de código/token con reloj inyectado (`recovery_test.go`), vencido/usado/inválido nunca revive (`recovery_repository_test.go`, `hu008_recuperacion_acceso.sql`). |
| `CA-008-04` | Cumplido | `staff_recovery_code_code_hash_ck`/`reset_token_hash_ck` exigen 64 hex (HMAC-SHA256/SHA-256, nunca el valor en claro ni una representación trivial de invertir); auditoría de logs/fixtures sin código/token/contraseña/teléfono/correo completo (`TestRecoveryRequestHandler_ThroughFullRouter_NeverLogsSensitiveValues` y equivalentes en `recovery_handler_test.go`). |
| `CA-008-05` | Cumplido | `TestRecovery_System_FullJourney_RequestVerifyResetLogsInWithNewPassword` (contraseña anterior deja de autenticar, nueva sí, contra el login real), `hu008_recuperacion_acceso.sql` (revocación total de sesiones en la misma operación), `recovery_repository_test.go` (dos consumos concurrentes del mismo token, un solo ganador). |
| `CA-008-06` | Cumplido | Destino enmascarado solo en la respuesta exitosa de `.../verify` (`mask_test.go`, `TestRecovery_System_FullJourney...` verifica que `maskedPhone` nunca sea igual al valor completo). |
| `CA-008-07` | Cumplido | Límite propio de reenvío y un solo código vigente por usuario (`hu008_recuperacion_acceso.sql`, subtest "immediate resend rejected by cooldown" de `recovery_repository_test.go`, índice único parcial `idx_staff_recovery_code_active_per_user`). |
| `CA-008-08` | Cumplido | Política de `DEC-063` con mensaje específico por regla incumplida (`ValidateNewPassword`, `recovery_test.go`, `recovery_handler_test.go`). |

### Pruebas

- Unitarias con reloj/generador determinista ficticio: `internal/modules/auth/recovery_test.go`.
- HTTP con dobles: `internal/modules/auth/httpapi/recovery_handler_test.go`, `contract_recovery_test.go`.
- Adaptadores de notificación (dobles de `httptest.Server`, nunca red real): `internal/modules/notification/whatsapp_meta_test.go`, `email_resend_test.go`, `recovery_sender_test.go`.
- PostgreSQL real, incluida concurrencia: `internal/modules/auth/postgres/recovery_repository_test.go`.
- SQL directo con el rol real: `database/tests/hu008_recuperacion_acceso.sql`.
- E2E backend/sistema contra el router y PostgreSQL reales, proveedor interceptado: `cmd/api/recovery_integration_test.go`.

## Registro y listado de barberos (HU-021)

`internal/modules/staff` implementa las cuatro operaciones privadas de
`api/openapi/paths/staff.yaml`: `GET`/`POST /api/v1/private/barbers` y
`GET`/`PATCH /api/v1/private/barbers/{barberId}`. Núcleo (`domain.go`,
`errors.go`, `ports.go`, `service.go`) sin Chi, `net/http`, pgx ni
`internal/platform/database` (CA-002-06); `postgres/repository.go` traduce
el puerto `staff.Repository` a `database.DB`; `httpapi/` decodifica, valida
la forma y traduce el resultado a HTTP. Migración
`database/migrations/20260823130000_create_barber.sql` (tabla `barber`
mínima, DEC-047: sin `active`, `deleted_at`, `sort_order` ni vínculo con
`staff_user`).

### Un mismo camino para barbería unipersonal y de equipo (CA-021-01/02)

`barber` es una tabla ordinaria, sin ninguna bandera ni columna que
distinga "el único barbero" de "uno de varios": una barbería con una
persona tiene exactamente una fila `barber`, y agregar tres más produce
cuatro filas distintas con el mismo `INSERT`, el mismo handler y el mismo
componente Vue. `TestStaff_HTTP_OneThenFourBarbers_SamePathBothCases`
(`cmd/api/staff_integration_test.go`) lo demuestra contra el router real.

### Paginación por cursor (CA-021-02)

La colección no tiene un máximo de negocio aprobado (`docs/06-api/
estandar-openapi.md` §6.10), así que `GET /private/barbers` pagina por
cursor opaco (`staff.EncodeCursor`/`DecodeCursor`, base64 de
`{createdAt, id}`) en vez de offset, con orden estable `(created_at, id)`
respaldado por `idx_barber_shop_created_id`. `limit` se clampa en el
servicio a `[1, 50]`, por defecto 20 (`staff.DefaultListLimit`); un cursor
manipulado o de otra forma se rechaza como `400` sin tocar PostgreSQL
(nunca revela ni modifica un recurso ajeno: el cursor es solo una posición
de recorrido, la fila real sigue protegida por RLS).

### Alta idempotente: Begin, INSERT y Complete en una sola transacción (RN-IDE-01, DEC-043)

`POST /private/barbers` sigue al pie de la letra el patrón documentado en
"Patrón obligatorio: idempotencia reutilizable" de arriba, pero a
diferencia del ejemplo de esa sección (que solo demuestra el mecanismo), la
orquestación completa vive en `staff/postgres.Repository.Create`: dentro de
la MISMA `InTenantTx`, llama `Coordinator.Begin`, ejecuta el `INSERT
... RETURNING` solo si el resultado es `OutcomeProceed`, construye el
cuerpo JSON de la respuesta (mismo formato exacto que
`httpapi.BarberResponse`, ver la advertencia de mantenimiento manual en
ambos archivos) y llama `Coordinator.Complete` con ese mismo cuerpo. El
handler HTTP nunca decide el cuerpo por su cuenta: reproduce
`result.Response` byte a byte tanto en la primera ejecución como en una
repetición exacta (`CA-004-01`), reconstruyendo `Location` a partir del
`id` que ese mismo cuerpo ya contiene (funciona igual para `OutcomeProceed`
y `OutcomeReplay`, sin un campo aparte). La clave de idempotencia NO impone
unicidad de nombre: dos claves distintas pueden crear dos barberos con el
mismo `fullName` (`TestCreate_DifferentKeysSameName_CreatesTwoDistinctBarbers`).

### `404` idéntico para inexistente y de otra barbería (CA-021-05, RN-TEN-01)

`Get` y `Rename` filtran explícitamente por `barbershop_id` además de RLS
y colapsan "no existe" y "es de otra barbería" en el mismo
`apperr.NotFound`, igual que el resto del backend. Un identificador que ni
siquiera tiene forma de UUID se rechaza con el mismo `404` ANTES de tocar
PostgreSQL (`staff.LooksLikeBarberID`), para no arriesgar un error de tipo
(500) por una comparación `text` contra una columna `uuid`.
`TestStaff_HTTP_TwoTenants_CrossAccessAlwaysReturns404WithoutLeaking`
(`cmd/api/staff_integration_test.go`) compara campo a campo el `Problem` de
un identificador inexistente contra el de un barbero real de otra
barbería (solo `instance`/`requestId`, de correlación por solicitud,
pueden diferir).

### Nombre: recorte, largo y Unicode (CA-021-03)

`staff.NormalizeFullName` recorta espacios; el servicio rechaza vacío,
solo espacios o más de 120 caracteres (`staff.FullNameMaxLength`, igual
que `barber_full_name_ck`) usando `utf8.RuneCountInString` (cuenta
caracteres, no bytes: un nombre con tildes o `ñ` no se rechaza de forma
prematura). Sin unicidad: dos barberos de la misma barbería pueden
compartir nombre (`TestRename_DuplicateNameAcrossBarbers_Allowed`).

### Tabla de criterios de aceptación

| Criterio | Estado | Prueba o evidencia |
| --- | --- | --- |
| `CA-021-01` | Cumplido | `TestList_FourBarbersInATeamShop_ReturnedAsFourDistinctResources`, `TestStaff_HTTP_OneThenFourBarbers_SamePathBothCases`, `database/tests/hu021_barberos.sql` ("CA-021-01/02"). |
| `CA-021-02` | Cumplido | Los mismos anteriores; alta protegida con idempotencia: `TestCreate_FirstExecution_PersistsAndReturnsProceed`, `TestStaff_HTTP_Idempotency_ReplayAndConflict`. |
| `CA-021-03` | Cumplido | `TestCreate_EmptyName_RejectedWithoutTouchingRepository`, `TestCreate_WhitespaceOnlyName_Rejected`, `TestCreate_NameOver120Characters_Rejected`, `TestCreate_NameExactly120Characters_Accepted`, `TestCreate_UnicodeName_CountsRunesNotBytes`, `barber_full_name_ck` en `database/tests/hu021_barberos.sql`. |
| `CA-021-04` | Cumplido | `TestRename_OwnTenant_UpdatesByIDWithoutDuplicating`, `TestStaff_HTTP_CreateGetListRenameReload_FullJourney`. |
| `CA-021-05` | Cumplido | `TestGet_CrossTenant_NeverLeaksAnotherShopsBarber`, `TestRename_CrossTenant_NeverRenamesAnotherShopsBarber`, `TestStaff_HTTP_TwoTenants_CrossAccessAlwaysReturns404WithoutLeaking`, sección CA-021-05 de `hu021_barberos.sql`. |
| `CA-021-06` | Cumplido | Sección CA-021-06 de `database/tests/hu021_barberos.sql` (RLS forzada, `WITH CHECK` rechaza `barbershop_id` ajeno, rol real sin `BYPASSRLS`). |
| `CA-021-07` | Cumplido | Sin política/GRANT `DELETE` (`database/tests/hu021_barberos.sql`), contrato sin campos fuera de alcance (`TestContract_UpdateBarberRequestSchema_MatchesDTOFields`), `TestStaff_HTTP_UnknownField_Returns400`. |
| `CA-021-08` | Parcial (backend no aplica; ver `apps/web/README.md`) | Evidencia de frontend/accesibilidad documentada en `apps/web/README.md` y `apps/web/e2e/`. |

### Pruebas

- Unitarias: `internal/modules/staff/domain_test.go`, `service_test.go` (doble en memoria de `staff.Repository`).
- HTTP con dobles: `internal/modules/staff/httpapi/handler_test.go`.
- Contrato: `internal/modules/staff/httpapi/contract_test.go`.
- PostgreSQL real, incluida concurrencia con `-race` (dos conexiones reales, misma clave): `internal/modules/staff/postgres/repository_test.go`.
- SQL directo con el rol real: `database/tests/hu021_barberos.sql`.
- Router de producción con dos tenants reales: `cmd/api/staff_integration_test.go`.

## Catálogo básico de servicios (HU-022)

`internal/modules/catalog` implementa las cuatro operaciones privadas de
`api/openapi/paths/catalog.yaml`: `GET`/`POST /api/v1/private/services` y
`GET`/`PATCH /api/v1/private/services/{serviceId}`. Núcleo (`domain.go`,
`errors.go`, `ports.go`, `service.go`) sin Chi, `net/http`, pgx ni
`internal/platform/database` (CA-002-06); `postgres/repository.go` traduce
el puerto `catalog.Repository` a `database.DB`; `httpapi/` decodifica,
valida la forma y traduce el resultado a HTTP. Migración
`database/migrations/20260824140000_create_service.sql` (tabla `service`
mínima: nombre, descripción opcional, duración en minutos, precio en COP,
`is_active` presente pero sin ningún endpoint que lo cambie -prepara
HU-024 sin exponer su transición-).

El caso de uso se llama `catalog.CatalogService`, no solo `Service` (a
diferencia de `staff.Service`/`shops.Service`): el propio recurso de este
módulo ya se llama `Service`, y un paquete Go no puede declarar dos tipos
con el mismo nombre.

### Dinero exacto: centavos enteros, nunca coma flotante (DEC-067)

`catalog.ParsePriceCOP`/`FormatPriceCOP` convierten entre el string
decimal del contrato (`"45000.00"`) y un `int64` de centavos usando
únicamente aritmética de enteros (división/módulo, sin `strconv.ParseFloat`
en ningún punto). `catalog/postgres` traduce esos centavos a/desde
`pgtype.Numeric` (`numericFromCents`/`centsFromNumeric`) también con
aritmética de enteros (`big.Int` escalado por la potencia de diez que
corresponda a `Exp`), para que el valor que PostgreSQL almacena en
`numeric(12,2)` haga un viaje de ida y vuelta exacto sin redondeo
acumulado. Un precio cero o negativo se rechaza antes de tocar el
repositorio (`errPriceMustBePositive`); un precio con más de dos cifras
decimales o formato inválido, también.

### Nombre único entre servicios activos: conflicto, no validación (DEC-067)

`idx_service_active_name` (índice único parcial `(barbershop_id, name)
WHERE is_active`) es la única fuente de verdad de esta regla: el núcleo no
la duplica con una consulta previa (evitaría una carrera entre dos altas
simultáneas con el mismo nombre). `catalog/postgres.Repository.Create`/
`.Update` detectan el `unique_violation` (`pgconn.PgError`, código `23505`,
`ConstraintName == "idx_service_active_name"`) y lo traducen a
`NameTaken`/`CreateResult.NameTaken`/`UpdateResult.NameTaken`, nunca a un
error genérico; `CatalogService` los traduce a `apperr.Conflict`
(`KindConflict`, nuevo en `internal/platform/apperr`), que
`httpserver.Translate` mapea a `409` con `code: "conflict"` -distinto de
`idempotency-conflict`/`idempotency-locked`, que no dependen de ninguna
cabecera `Idempotency-Key`-. En `Create`, el conflicto ocurre DENTRO de la
misma `InTenantTx` que la reclamación de idempotencia: el `ROLLBACK`
resultante deshace también esa reclamación, dejando la clave libre para un
reintento legítimo con un nombre distinto (mismo criterio de `CA-004-06`).

### `404` idéntico para inexistente y de otra barbería (CA-022-06, RN-TEN-01)

Mismo patrón que `staff`: `Get` y `Update` filtran explícitamente por
`barbershop_id` además de RLS y colapsan "no existe" y "es de otra
barbería" en el mismo `apperr.NotFound`. Un identificador que ni siquiera
tiene forma de UUID se rechaza con el mismo `404` antes de tocar
PostgreSQL (`catalog.LooksLikeServiceID`).

### Edición parcial con `description` que se puede "borrar" (CA-022-04/05)

`catalog.UpdateFields.Description` es un `OptionalDescription{Set, Value}`,
no un simple puntero: `Set=false` significa "el cliente no envió este
campo, no tocar"; `Set=true, Value=nil` significa "borrar la descripción
existente". La capa HTTP deriva `Set` de la sola presencia del puntero
`*string` tras decodificar el JSON (una clave ausente dejó el puntero
`nil`; una clave presente, incluso con cadena vacía, lo dejó no-nulo). El
repositorio traduce esto a SQL con `CASE WHEN $4::boolean THEN $5::text
ELSE description END`, nunca con `COALESCE` a secas (que no podría
expresar "borrar"). Un `PATCH` sin ningún campo de catálogo presente se
rechaza como `apperr.Validation` antes de tocar el repositorio
(`errUpdateEmptyBody`); ninguna operación acepta `isActive`, asignaciones a
barberos ni alcance de propagación hacia una cita (RN-SER-04): el contrato
ni siquiera declara esos campos.

### Tabla de criterios de aceptación

| Criterio | Estado | Prueba o evidencia |
| --- | --- | --- |
| `CA-022-01` | Cumplido | `TestList_FourServicesInACatalog_ReturnedAsFourDistinctResources`, `TestCatalog_HTTP_CreateGetListUpdateReload_FullJourney`, `database/tests/hu022_catalogo.sql` ("CA-022-01"). |
| `CA-022-02` | Cumplido | Alta protegida con idempotencia: `TestCreate_FirstExecution_PersistsAndReturnsProceed`, `TestCreate_SameKeyAndContent_ReplaysWithoutCreatingASecondRow`, `TestCatalog_HTTP_Idempotency_ReplayAndConflict`. |
| `CA-022-03` | Cumplido | `TestCreate_Durations25_30_45_90_AllAccepted`, `TestCreate_DurationZeroNegativeOrAbove1440_Rejected`, `TestCreateServiceHandler_FractionalDuration_Returns400`, sección de `hu022_catalogo.sql`. |
| `CA-022-04` | Cumplido | `TestCreateServiceHandler_UnknownField_Returns400`, `TestCreate_PriceZeroOrNegative_Rejected`, `TestCreate_DuplicateActiveName_ReturnsNameTakenWithoutPersisting`, `TestCatalog_HTTP_DuplicateActiveName_Returns409`. |
| `CA-022-05` | Cumplido | Contrato sin campo de alcance sobre citas (`TestContract_UpdateServiceRequestSchema_MatchesDTOFields` rechaza `barberIds`/`appointments`); `TestCatalog_HTTP_CreateGetListUpdateReload_FullJourney` confirma que un cambio de precio/duración no toca `name`. |
| `CA-022-06` | Cumplido | `TestGet_CrossTenant_NeverLeaksAnotherShopsService`, `TestUpdate_CrossTenant_NeverEditsAnotherShopsService`, `TestCatalog_HTTP_TwoTenants_CrossAccessAlwaysReturns404WithoutLeaking`, sección CA-022-06 de `hu022_catalogo.sql`. |
| `CA-022-07` | Cumplido | `openapi:lint`/`openapi:bundle` en verde; `contract_test.go` (9 pruebas) compara los DTO Go y las cuatro operaciones reales contra el YAML fuente. |
| `CA-022-08` | Parcial (backend no aplica; ver `apps/web/README.md`) | Evidencia de frontend/accesibilidad documentada en `apps/web/README.md` y `apps/web/e2e/`. |

### Pruebas

- Unitarias: `internal/modules/catalog/domain_test.go`, `service_test.go` (doble en memoria de `catalog.Repository`).
- HTTP con dobles: `internal/modules/catalog/httpapi/handler_test.go`.
- Contrato: `internal/modules/catalog/httpapi/contract_test.go`.
- PostgreSQL real, incluida concurrencia con `-race` y precisión monetaria exacta: `internal/modules/catalog/postgres/repository_test.go`.
- SQL directo con el rol real: `database/tests/hu022_catalogo.sql`.
- Router de producción con dos tenants reales: `cmd/api/catalog_integration_test.go`.

## Asignación de servicios a barberos (HU-023)

`internal/modules/catalog` agrega las tres operaciones privadas de
`api/openapi/paths/barber-services.yaml`:
`GET /api/v1/private/barbers/{barberId}/services`,
`PUT`/`DELETE .../{barberId}/services/{serviceId}`. Núcleo nuevo
(`assignment.go`, `assignment_errors.go`, `assignment_ports.go`,
`assignment_service.go`) en el mismo paquete `catalog` (dueño de la
intención "qué servicios se prestan"), sin Chi, `net/http`, pgx ni
`internal/platform/database` (CA-002-06); `postgres/
assignment_repository.go` traduce `catalog.AssignmentRepository` a
`database.DB`; `httpapi/assignment_{dto,handler}.go` decodifica y traduce a
HTTP. Migración `database/migrations/20260824150000_create_barber_service.sql`
(tabla de asociación PURA `barber_service`: solo `barbershop_id`,
`barber_id`, `service_id`, `created_at`, sin duplicar nombre/duración/
precio/estado, CA-023-07).

### Colaboración entre módulos SOLO por puerto explícito (trabajo requerido §3.1)

`catalog.AssignmentService` necesita confirmar que un `barberId` pertenece
a la barbería vigente antes de asignarle un servicio, pero **el paquete
`catalog` nunca importa `staff`**, ni siquiera en sus propias pruebas
(`internal/platform/archtest/module_boundary_test.go` lo verifica
recorriendo los imports de ambos árboles de paquetes). En su lugar,
`catalog` declara su propio puerto pequeño:

```go
// catalog/assignment_ports.go
type BarberPort interface {
    Exists(ctx context.Context, barbershopID, barberID string) (bool, error)
}
```

`staff` expone `staff.BarberLookup` (`staff/lookup.go`), un adaptador de una
sola operación sobre `*staff.Service` que satisface esa interfaz de forma
puramente estructural, sin que `staff` importe `catalog` tampoco.
`cmd/api.buildRouter` (la única raíz de composición que conoce ambos
módulos) conecta ambos lados:

```go
assignmentService := catalog.NewAssignmentService(
    catalogpostgres.NewAssignmentRepository(db),
    staff.NewBarberLookup(staffService),
)
```

La existencia de `serviceId` la verifica `catalog` directamente (es su
propia tabla `service`, mismo módulo bajo prueba, no una dependencia
cruzada).

### Última asignación activa: rechazo con bloqueo de fila (DEC-068, CA-023-05/06)

`AssignmentRepository.Unassign` ejecuta, dentro de UNA sola `InTenantTx`:

1. `SELECT is_active FROM service WHERE id = $1 AND barbershop_id = $2
   FOR UPDATE`: bloquea la fila de `service` hasta el fin de la
   transacción. Dos desasignaciones concurrentes del MISMO `service_id` se
   serializan aquí, sin importar qué `barber_id` retire cada una.
2. Confirma que la asociación exista (si no, `UnassignOutcomeNotFound`).
3. Cuenta cuántas filas activas quedan para ese `service_id` (incluida la
   que se retiraría). Si el servicio está activo y esa cuenta es `<= 1`,
   sería la última: `UnassignOutcomeLastActiveConflict`, sin borrar nada.
4. En cualquier otro caso, `DELETE` y `UnassignOutcomeDeleted`.

El lock del paso 1 es lo que hace la carrera segura: una segunda
transacción que intente retirar la penúltima fila del mismo servicio espera
ahí; al reanudarse, ve la cuenta YA actualizada y rechaza correctamente si
le toca ser la última.
`TestUnassign_ConcurrentRaceOnLastTwoAssignments_ExactlyOneSucceeds`
(`internal/modules/catalog/postgres/assignment_repository_test.go`, con
`-race`, dos conexiones reales) demuestra esto contra PostgreSQL real.

### Semántica HTTP naturalmente repetible, sin `Idempotency-Key`

Asignar y desasignar son `PUT`/`DELETE` sobre un recurso identificado por
sus propios dos identificadores (`barberId`+`serviceId`): repetir la misma
solicitud produce el mismo estado final sin ambigüedad de contenido, así
que ninguna de las dos operaciones usa el protocolo de idempotencia de
RN-IDE-01/`DEC-043` (reservado a un `POST` que no sería idempotente por sí
mismo). Repetir `PUT` responde `200` (no `201`) con el `createdAt`
ORIGINAL, sin insertar una segunda fila (`INSERT ... ON CONFLICT DO
NOTHING`); repetir `DELETE` sobre una asociación ya retirada responde `404`.

### Tabla de criterios de aceptación

| Criterio | Estado | Prueba o evidencia |
| --- | --- | --- |
| `CA-023-01` | Cumplido | `TestAssign_New_CreatesRowAndIsVisibleInList`, `TestBarberServices_HTTP_AssignRepeatListUnassign_FullJourney`, E2E `servicios-por-barbero.spec.ts`. |
| `CA-023-02` | Cumplido | `TestAssign_Repeated_DoesNotCreateASecondRow`, `TestAssignmentService_Assign_Repeated_ReturnsAlreadyExistsWithoutError`, HTTP journey (200 con `createdAt` original en la repetición). |
| `CA-023-03` | Cumplido | `TestAssign_SameServiceToMultipleBarbers_EachIsAnIndependentResource`, E2E ("un mismo servicio se asigna a varios barberos"). |
| `CA-023-04` | Cumplido | `TestAssign_ServiceFromAnotherTenant_ServiceNotFound`, `TestRawSQL_CompositeFK_RejectsCrossTenantAssociation` (FK real, `23503`), `TestBarberServices_HTTP_TwoTenants_CrossAccessReturns404WithoutLeaking`, `hu023_asignaciones.sql` ("CA-023-04"). |
| `CA-023-05` | Cumplido | `TestUnassign_LastActiveAssignment_RejectedWithoutDeleting` (incluye reintento seguro), `TestBarberServices_HTTP_LastActiveAssignment_Returns409`, E2E DEC-068. |
| `CA-023-06` | Cumplido | `TestUnassign_ConcurrentRaceOnLastTwoAssignments_ExactlyOneSucceeds` (`-race`, dos conexiones reales), `hu023_asignaciones.sql` ("CA-023-06"). |
| `CA-023-07` | Cumplido | Esquema exacto verificado en `hu023_asignaciones.sql`; `TestBarberServices_HTTP_ResponseNeverIncludesNameDurationPrice`. |
| `CA-023-08` | Parcial (backend no aplica; ver `apps/web/README.md`) | Evidencia de frontend/accesibilidad documentada en `apps/web/README.md` y `apps/web/e2e/`. |

### Pruebas

- Unitarias: `internal/modules/catalog/assignment_service_test.go` (doble en memoria de `catalog.AssignmentRepository`/`BarberPort`); `internal/modules/staff/lookup_test.go`.
- PostgreSQL real, incluida la carrera de DEC-068 con `-race`: `internal/modules/catalog/postgres/assignment_repository_test.go`.
- SQL directo con el rol real: `database/tests/hu023_asignaciones.sql`.
- Router de producción con dos tenants reales: `cmd/api/barber_services_integration_test.go`.
- Arquitectura (sin imports cruzados `catalog`↔`staff`): `internal/platform/archtest/module_boundary_test.go`.

## Ciclo de vida de servicios (HU-024)

`internal/modules/catalog` agrega las tres operaciones privadas de
`api/openapi/paths/service-lifecycle.yaml`:
`GET /api/v1/private/services/{serviceId}/deactivation-impact`,
`POST .../deactivate`, `POST .../reactivate`. Mismo núcleo `catalog`, mismo
paquete que HU-022/HU-023 (`CatalogService.PreviewDeactivation`/
`Deactivate`/`Reactivate` en `service.go`); sin migración nueva:
`service.is_active`/`deactivated_at` y `service_deactivated_at_ck` ya
existían desde `20260824140000_create_service.sql` (HU-022 los preparó sin
exponer ningún endpoint que los cambiara).

### El impacto de citas futuras es literal, no un puerto simulado (DEC-069)

`CA-024-01`/`CA-024-04` exigen que la advertencia muestre el impacto REAL de
desactivar un servicio, consultado en el momento, nunca un valor por
defecto. `appointment` no existe todavía en la cadena migrada (B3), así que
"el impacto real" en B1 es, literalmente, cero: ninguna cita puede existir
para ningún servicio de ninguna barbería. `catalog.currentDeactivationImpact()`
(`domain.go`) documenta esta verdad explícitamente y es el ÚNICO lugar que
decide `AffectedAppointments`; no es un `Port`/adaptador que finja consultar
una tabla que no existe (`trabajo requerido §2`, "no inventes... un
adaptador que siempre devuelva cero... como dato falso"). Cuando B3 cree
`appointment`, esta única función se reemplaza por una consulta real, sin
tocar el resto de la historia.

### Transición atómica con bloqueo de fila, sin optimista (DEC-069)

`Repository.Deactivate`/`Reactivate` ejecutan, dentro de UNA sola
`InTenantTx` que además coordina la idempotencia de HU-004 (`Idempotency-Key`,
RN-IDE-01):

1. `Begin` reclama la clave de idempotencia (`OutcomeProceed` continúa;
   `OutcomeReplay` reproduce la respuesta guardada byte a byte; cualquier
   otro desenlace se traduce a `409` sin tocar la fila).
2. `lockServiceForTransition` bloquea la fila (`SELECT is_active FROM
   service WHERE id = $1 AND barbershop_id = $2 FOR UPDATE`): dos
   confirmaciones concurrentes del MISMO servicio se serializan aquí,
   idéntico patrón que `AssignmentRepository.Unassign` (HU-023, DEC-068).
   `found=false` (fila inexistente o de otra barbería) hace `ROLLBACK`
   entero, dejando la clave de idempotencia libre para un reintento
   legítimo (`CA-024-07`).
3. Si el estado ya coincide con el destino (desactivar un servicio ya
   inactivo, o viceversa), `InvalidTransition=true` y `ROLLBACK`: la
   transacción entera se revierte, la clave queda libre, y la capa HTTP
   traduce esto a `409` (`code: conflict`, CA-024-06) — nunca un éxito
   silencioso sobre una transición que ya no aplica.
4. En cualquier otro caso, el `UPDATE` condicionado y `Complete` cierran la
   reclamación de idempotencia con la respuesta final.

`DEC-069` es explícita en que este bloqueo NO es "protección de concurrencia
sobre el conteo de citas" (imposible de construir honestamente sin
`appointment`): es la misma guarda de estado que cualquier transición
atómica necesita sobre su propia fila, independiente de B1/B3.
`TestDeactivate_TwoRealConcurrentConnections_ExactlyOneSucceeds`
(`internal/modules/catalog/postgres/lifecycle_repository_test.go`, `-race`,
dos conexiones reales) demuestra que, ante dos confirmaciones simultáneas
con claves de idempotencia DISTINTAS sobre el mismo servicio, exactamente
una transiciona y la otra descubre `InvalidTransition` tras esperar el
lock — nunca un error, nunca una segunda escritura.

### `ServiceResponse` gana `isActive`/`deactivatedAt`, de solo lectura

El contrato de HU-022 declaraba explícitamente que `ServiceResponse` nunca
exponía estos dos campos ("el ciclo de activación pertenece a HU-024"). Con
HU-024 implementada, `isActive`/`deactivatedAt` pasan a ser parte
permanente de la representación canónica (`GET`/lista/alta/edición los
devuelven también), pero siguen siendo de solo lectura: `CreateServiceRequest`/
`UpdateServiceRequest` (HU-022) NUNCA los aceptan como entrada — solo
cambian mediante `deactivate`/`reactivate`.
`postgres.serviceResponseWire`/`deactivationResponseWire` declaran la MISMA
forma exacta que `httpapi.ServiceResponse`/`ServiceDeactivationResponse`
(mismo criterio de mantenimiento manual que HU-022 ya documentaba, ver
`TestCreate_StoredResponseBody_MatchesHTTPAPIWireShape`).

### Tabla de criterios de aceptación

| Criterio | Estado | Prueba o evidencia |
| --- | --- | --- |
| `CA-024-01` | Cumplido | `TestPreviewDeactivation_Found_ReturnsZeroAffectedAppointments`, `TestGetServiceDeactivationImpactHandler_Found_Returns200WithZero`, `TestServiceLifecycle_HTTP_PreviewDeactivateReactivate_FullJourney`. |
| `CA-024-02` | Cumplido | `TestDeactivate_ActiveService_SetsInactiveWithTimestamp`, `hu024_ciclo_vida.sql` ("CA-024-02"), journey HTTP completo. |
| `CA-024-03` | Cumplido | Ninguna operación de `appointment`/notificación existe en el código; alcance verificado por ausencia, no por prueba positiva (B3 todavía no existe). |
| `CA-024-04` | Cumplido | El impacto se recalcula dentro de `Deactivate` (nunca recibido del cliente); `TestDeactivateServiceHandler_Proceed_Returns200WithStoredBody` confirma `affectedAppointments` en la respuesta de confirmación. |
| `CA-024-05` | Cumplido | `TestReactivate_InactiveService_SetsActiveClearsTimestamp`, `hu024_ciclo_vida.sql` ("CA-024-05" ×2), journey HTTP. |
| `CA-024-06` | Cumplido | `TestDeactivate_Repeated_SameKey_ReturnsSameResponseWithoutSecondEffect` (replay), `TestDeactivate_AlreadyInactive_NewKey_ReturnsInvalidTransitionWithoutChangingRow` (clave nueva rechazada). |
| `CA-024-07` | Cumplido | `TestDeactivate_CrossTenant_NeverLeaksAnotherShopsService`, `TestDeactivate_UnknownID_ReturnsNotFound`, `TestServiceLifecycle_HTTP_TwoTenants_CrossAccessReturns404WithoutLeaking`, `hu024_ciclo_vida.sql` ("CA-024-07"). |
| `CA-024-08` | Parcial (backend no aplica; ver `apps/web/README.md`) | Evidencia de frontend/accesibilidad documentada en `apps/web/README.md`. |

### Pruebas

- Dominio: `internal/modules/catalog/service_test.go` (doble en memoria de `catalog.Repository`; 12 pruebas nuevas de `PreviewDeactivation`/`Deactivate`/`Reactivate`).
- PostgreSQL real, incluida la carrera de dos conexiones con `-race`: `internal/modules/catalog/postgres/lifecycle_repository_test.go` (9 pruebas).
- HTTP con dobles: `internal/modules/catalog/httpapi/handler_test.go` (15 pruebas nuevas) y `contract_test.go` (contrato contra el YAML fuente).
- SQL directo con el rol real: `database/tests/hu024_ciclo_vida.sql`.
- Router de producción con dos tenants reales: `cmd/api/service_lifecycle_integration_test.go` (4 pruebas de recorrido completo).

## Horario laboral recurrente (HU-040)

`internal/modules/schedule` (nuevo módulo) implementa las cinco operaciones
privadas de `api/openapi/paths/schedules.yaml` sobre `working_hour`
(`20260825160000_create_working_hour.sql`): listar (paginada por cursor,
orden `iso_weekday, starts_time, id`), crear (protegida con
`Idempotency-Key`, RN-IDE-01), consultar, editar (reemplaza el intervalo
completo) y retirar (físico; `working_hour` no tiene eliminación lógica).

### El solape vive en Go, no en un `EXCLUDE` de PostgreSQL

Un tramo nocturno (DEC-020) hace que una restricción de exclusión con
envolvente semanal sea frágil (mismo razonamiento documentado en el
comentario de la migración y en `estandar-base-datos.md §7`). En su lugar,
`schedule.IntervalsOverlap` compara dos intervalos `[inicio, fin)` en
minutos desde medianoche (semiabiertos: tramos contiguos NO se solapan,
CA-040-03), y `postgres.overlapsExisting` la invoca dentro de la misma
transacción que el `INSERT`/`UPDATE`, sobre los tramos existentes del mismo
`(barbershopID, barberID, isoWeekday)` ya bloqueados con
`SELECT ... FOR UPDATE`.

### Bloqueo de la fila de `barber`, no solo de `working_hour` (carrera de "cero tramos")

Si dos altas concurrentes para el MISMO barbero y día parten de un
conjunto de tramos existentes VACÍO, un `SELECT ... FOR UPDATE` sobre
`working_hour` no bloquea nada (no hay fila que bloquear): ambas
transacciones podrían insertar tramos que se solapan entre sí sin que
ninguna vea a la otra (phantom read clásico a `READ COMMITTED`).
`lockBarberForScheduleWrite` bloquea en cambio la fila de `barber` antes de
verificar solape, serializando todas las escrituras de horario de ese
barbero (mismo criterio que DEC-068 bloqueando la fila de `service`).
`TestCreate_ConcurrentOverlappingCreates_ExactlyOneSucceeds`
(`internal/modules/schedule/postgres/repository_test.go`, `-race`, dos
goroutines reales sobre el mismo pool) demuestra que exactamente una de dos
altas concurrentes que se solapan tiene éxito; la otra recibe
`Conflict=true` sin dejar dos tramos cruzados.

### `CT-008`/`DEC-070`: `ON DELETE RESTRICT`, no `CASCADE`

El modelo de referencia proponía `ON DELETE CASCADE` de `working_hour`
hacia `barber`; `AGENTS.md` prohíbe el borrado en cascada. Se resolvió como
`DEC-070`: la FK usa `ON DELETE RESTRICT` (`barber` no tiene borrado físico
en su alcance vigente, HU-021/DDL-BIZ-02), verificado contra PostgreSQL
real en `tests/hu040_horario.sql`, no solo documentado.

### `BarberPort`: mismo patrón de colaboración entre módulos que HU-023

`schedule.Service` verifica que `barberId` exista en la barbería activa
mediante `schedule.BarberPort`, satisfecho por `staff.NewBarberLookup`
(mismo puerto pequeño y explícito que `catalog.BarberPort`): ni `schedule`
importa `staff`, ni `staff` importa `schedule`; `cmd/api` conecta ambos.

### Tabla de criterios de aceptación

| Criterio | Estado | Prueba o evidencia |
| --- | --- | --- |
| `CA-040-01` | Cumplido | `TestList_OrderedByWeekdayThenStartsTime`, `TestList_FirstPage_NeverReturnsMoreThanLimit`, `TestWorkingHours_HTTP_CreateListGetUpdateDelete_FullJourney`. |
| `CA-040-02` | Cumplido | `TestCreate_SplitShift_TwoNonOverlappingSegmentsSameDay`, `hu040_horario.sql` ("CA-040-02/03"). |
| `CA-040-03` | Cumplido | `TestIntervalsOverlap_*` (dominio), `TestCreate_NightShift_CrossesMidnightWithoutError`, `TestCreate_ContiguousSegments_NoConflict`. |
| `CA-040-04` | Cumplido | `TestCreate_ExactSameStart_ReturnsConflict`, `TestCreate_PartialOverlap_ReturnsConflict`, `TestUpdate_OverlapsAnotherSegment_ReturnsConflictWithoutChangingAnything`, `TestWorkingHours_HTTP_OverlappingCreate_Returns409`, `TestWorkingHours_HTTP_InvalidField_Returns422`, `hu040_horario.sql` (CHECK de día/duración). |
| `CA-040-05` | Cumplido | `TestGet_CrossBarber_...`, `TestGet_CrossTenant_...`, `TestUpdate_CrossTenant_...`, `TestDelete_CrossTenant_...`, `TestDelete_ThenGet_NotFound`, `TestWorkingHours_HTTP_CrossTenantBarber_Returns404`, `hu040_horario.sql` ("CA-040-05"). |
| `CA-040-06` | Cumplido | `startsTime` viaja como hora civil `HH:MM` sin conversión alguna (`to_char(starts_time, 'HH24:MI')`); ver `apps/web/README.md` para la indicación explícita de zona en pantalla. |
| `CA-040-07`/`08` | Parcial (backend no aplica) | Evidencia de frontend/accesibilidad documentada en `apps/web/README.md`. |

### Pruebas

- Dominio: `internal/modules/schedule/domain_test.go`, `service_test.go` (doble en memoria de `schedule.Repository`/`BarberPort`).
- PostgreSQL real, incluida la carrera de dos altas concurrentes con `-race`: `internal/modules/schedule/postgres/repository_test.go`.
- HTTP con dobles y contrato: `internal/modules/schedule/httpapi/contract_test.go` (contrato contra el YAML fuente; sin `handler_test.go` propio: las rutas con dos parámetros de ruta se cubren mediante el router real, mismo criterio que `catalog/httpapi/assignment_handler.go`).
- SQL directo con el rol real: `database/tests/hu040_horario.sql` (incluida la verificación de `DEC-070` contra PostgreSQL real).
- Router de producción con dos tenants reales: `cmd/api/schedule_integration_test.go`.

## Excepciones de jornada y festivos (HU-041)

Extiende `internal/modules/schedule` (mismo módulo de HU-040, mismo
`schedule.Service`/`schedule.Repository`) sobre dos tablas nuevas de
`20260826090000_create_working_hour_override.sql`: `working_hour_override`
(cabecera: fecha efectiva, `is_closed`, `reason` opcional) y
`working_hour_override_segment` (tramos de un día abierto), más
`barber.holiday_calendar_enabled`.

### El solape SÍ vive en un `EXCLUDE` de PostgreSQL, a diferencia de HU-040

`working_hour_override_segment_no_overlap_excl` (`EXCLUDE USING gist`)
rechaza dos tramos que se solapan dentro de la MISMA excepción. La
justificación de HU-040 para mover el solape a Go (una restricción de
exclusión sobre una envolvente semanal es frágil frente a un tramo
nocturno) no aplica aquí: cada excepción está anclada a una fecha civil
fija de la cabecera, no a un día de la semana que se repite, así que no
hay envolvente que fragilizar. `UNIQUE (barbershop_id, barber_id,
effective_date)` hace atómica la detección de fecha duplicada
(`CA-041-05`) por el mismo motivo: sin recurrencia, la base es un
backstop seguro sin necesitar el truco de bloquear la fila de `barber`
que sí hizo falta en HU-040.

### `UpdateException`: ROLLBACK completo si el conflicto aparece a mitad de transacción

`postgres.Repository.UpdateException` hace `UPDATE` de la cabecera,
`DELETE` de los tramos previos e `INSERT` de los nuevos dentro de LA MISMA
transacción. Si el `INSERT` de los tramos nuevos choca contra el
`EXCLUDE` (o el `UPDATE` de la cabecera choca contra el `UNIQUE` de
fecha) DESPUÉS de que la cabecera o los tramos previos ya se modificaron,
el sentinela `errExceptionConflictInternal` fuerza el `ROLLBACK` de la
transacción COMPLETA (mismo patrón que `errOverlapConflictInternal` de
HU-040): la excepción nunca queda a medio reemplazar (cabecera nueva,
tramos vacíos).

### `ResolveEffectiveDay`: puerto interno de precedencia (`CA-041-07`), no expuesto por HTTP

Para una fecha civil concreta, en este orden: (1) una excepción manual de
esa fecha, abierta o cerrada, prevalece siempre; (2) en su ausencia, un
festivo colombiano (`ColombianHolidaysForYear`) bloquea el día SOLO si el
barbero activó `holiday_calendar_enabled`; (3) en cualquier otro caso, el
horario semanal de `working_hour` (HU-040) para el día ISO
correspondiente. Es un puerto interno para que la disponibilidad (B4) lo
consuma después; no crea citas ni expone disponibilidad pública por sí
mismo.

### Festivos colombianos: calculado, no una tabla ni un proveedor externo

`schedule.ColombianHolidaysForYear` implementa la Ley 51 de 1983 ("Ley
Emiliani"): dieciocho festivos por año — seis fijos que nunca se mueven,
dos ligados a Pascua que tampoco se mueven (Jueves y Viernes Santo), y
diez (siete fijos y tres ligados a Pascua) que se trasladan al lunes
siguiente cuando no caen ya en lunes. La fecha de Pascua usa el algoritmo
gregoriano anónimo (Meeus/Jones/Butcher), válido para cualquier año del
calendario gregoriano. `GET /api/v1/private/schedule/colombian-holidays`
expone este cálculo (parámetro `year` obligatorio) como dato de
referencia, igual para toda barbería y todo barbero: útil para que la
pantalla ofrezca "abrir este festivo" sin que el cliente reimplemente el
algoritmo.

### Tabla de criterios de aceptación

| Criterio | Estado | Prueba o evidencia |
| --- | --- | --- |
| `CA-041-01`/`02` | Cumplido | `TestGetHolidayCalendar_*`, `TestSetHolidayCalendar_*`, `TestHolidayCalendarEnabled_*`, `TestSetHolidayCalendarEnabled_*`, `TestHolidayCalendar_HTTP_GetThenUpdate_PersistsToggle`. |
| `CA-041-03` | Cumplido | `TestResolveEffectiveDay_HolidayAuto_WhenEnabledAndNoException`, `TestResolveEffectiveDay_HolidayIgnored_WhenCalendarDisabled`. |
| `CA-041-04` | Cumplido | `TestValidateExceptionShape_*` (dominio), `TestCreateException_OverlappingSegments_ReturnsConflictAndNothingPersists`, `TestScheduleExceptions_HTTP_InvalidShape_Returns422`, `hu041_excepciones.sql`. |
| `CA-041-05` | Cumplido | `TestCreateException_DuplicateDate_ReturnsConflict`, `TestUpdateException_ConflictWithAnotherDate_LeavesOriginalUntouched`, `TestScheduleExceptions_HTTP_DuplicateDate_Returns409`, `hu041_excepciones.sql`. |
| `CA-041-06` | Cumplido | `TestGetException_CrossBarber_*`, `TestGetException_CrossTenant_*`, `TestUpdateException_CrossTenant_*`, `TestDeleteException_CrossTenant_*`, `TestDeleteException_ThenGet_NotFound`, `TestScheduleExceptions_HTTP_CrossTenantBarber_Returns404`, `hu041_excepciones.sql` ("CA-041-06"). |
| `CA-041-07` | Cumplido | `TestResolveEffectiveDay_ManualClosed_PrevailsOverEverything`, `TestResolveEffectiveDay_ManualOpen_PrevailsOverHolidayAndWeekly`, `TestResolveEffectiveDay_Weekly_WhenNoExceptionAndNotHoliday`. |
| `CA-041-08` | Parcial (backend no aplica) | Evidencia de frontend/accesibilidad documentada en `apps/web/README.md`. |

### Pruebas

- Dominio: `internal/modules/schedule/colombian_holidays_test.go`, `exception_domain_test.go`, `exception_service_test.go` (doble en memoria de `schedule.Repository`/`BarberPort`).
- PostgreSQL real: `internal/modules/schedule/postgres/exception_repository_test.go`.
- HTTP con contrato: `internal/modules/schedule/httpapi/contract_test.go` (extendido con las ocho operaciones nuevas).
- SQL directo con el rol real: `database/tests/hu041_excepciones.sql` (incluida la verificación de `DEC-070` contra PostgreSQL real).
- Router de producción con dos tenants reales: `cmd/api/schedule_exceptions_integration_test.go`.

## Núcleo persistente de citas (HU-060)

Nuevo módulo `internal/modules/booking`, dueño de las tablas nuevas de
`20260827110000_create_appointment_core.sql`: `customer`, `appointment`,
`appointment_history`, `appointment_history_change`. HU-060 entrega
únicamente la base persistente y una primitiva transaccional interna;
ningún endpoint HTTP, operación OpenAPI ni pantalla existe todavía. `HU-061`
construirá la política de creación manual (incluida la reconciliación de
`DP-CIT-01`) sobre `booking.BookingService.CreateInternal`.

### `CreateInternal`: cliente + cita + historial, una sola transacción

`booking.BookingService.CreateInternal` valida `CreateInternalInput` (forma
cerrada de estado/origen/intervalo/snapshot/actor, sin tocar la base) y
delega en `booking.Repository.CreateInternal`
(`internal/modules/booking/postgres`), que dentro de UNA sola `InTenantTx`:
crea o vincula el cliente que `CustomerInput` ya decidió (`ExistingID`
resuelto por `SELECT ... WHERE barbershop_id = $1 AND id = $2`, tenant-aware;
`New` inserta una fila), inserta la cita `confirmed` y el evento
`appointment_created`. Un fallo en cualquier paso revierte los tres, incluido
un cliente nuevo que nunca queda persistido si la cita choca con la
exclusión (`TestCreateInternal_ScheduleConflict_RollsBackNewCustomer`).

### La restricción de exclusión es la última defensa, no el dominio Go

`appointment_barber_interval_excl` (`EXCLUDE USING gist (barbershop_id,
barber_id, tstzrange(starts_at, ends_at, '[)')) WHERE (occupies_schedule)`)
es quien de verdad impide dos citas cruzadas del mismo barbero, sea cual sea
lo que valide antes el dominio Go. `occupies_schedule` es una columna
generada y almacenada (`status IN ('confirmed', 'completed', 'no_show')`,
estados-citas.md §4/§11): el mismo criterio alimenta la exclusión y, en B4,
el cálculo de disponibilidad, sin poder divergir.

### Hallazgo real de concurrencia: `deadlock_detected` además de `exclusion_violation`

Verificado contra PostgreSQL 14 real con
`TestCreateInternal_ConcurrentOverlap_ExactlyOneSucceeds` (dos goroutines,
cada una con su propia conexión del pool, ejecutando `-race`): cuando dos
`INSERT` que se cruzan llegan de verdad al mismo tiempo, PostgreSQL puede
resolver la disputa del índice GiST como `deadlock_detected` (SQLSTATE
`40P01`) en vez de un `exclusion_violation` (`23P01`) limpio para la
transacción perdedora — la comprobación especulativa del índice puede crear
un ciclo de espera entre los dos `INSERT` simultáneos. `insertAppointment`
traduce ambos códigos al mismo `apperr.Conflict`
(`bookingpostgres.isDeadlockDetected`): esta es la única fuente posible de
un interbloqueo dentro de esa función, así que la traducción es segura.

### Snapshots de servicio: centavos en Go, `numeric(12,2)` en PostgreSQL

`ServiceSnapshot.PriceAmountCents` es un entero exacto, nunca coma flotante
(mismo criterio que `catalog.Service.PriceCents`); `formatPriceAmount`/
`parsePriceAmount` convierten hacia y desde el `numeric(12,2)` real de
`price_amount_snapshot`. `booking` no importa `catalog`: el snapshot llega
ya resuelto en el `CreateInternalInput` del llamador (HU-061 en adelante),
sin sincronización automática con el catálogo (`DEC-004`).

### Historial append-only

`appointment_history`/`appointment_history_change` no tienen `GRANT UPDATE`
ni `DELETE` para `barberia_app`, ni política RLS que los permita
(`database/tests/hu060_citas.sql`): una corrección futura (T8,
`appointment_status_corrected`) agregará otra entrada, nunca editará la
existente (RN-HIS-02, `DEC-014`).

### Tabla de criterios de aceptación

| Criterio | Estado | Prueba o evidencia |
| --- | --- | --- |
| `CA-060-01` | Cumplido | `database/tests/hu060_citas.sql` ("esquema OK", "CHECK OK" ×2), `20260827110000_create_appointment_core.sql`. |
| `CA-060-02` | Cumplido | `hu060_citas.sql` ("EXCLUDE OK"), `TestCreateInternal_ScheduleConflict_RollsBackNewCustomer`, `TestCreateInternal_ContiguousInterval_Succeeds`, `TestCreateInternal_DifferentBarber_SameWindow_Succeeds`. |
| `CA-060-03` | Cumplido | `hu060_citas.sql` ("occupies_schedule OK"), `TestStatus_OccupiesSchedule`, `TestCreateInternal_NewCustomer_PersistsAppointmentAndHistory`. |
| `CA-060-04` | Cumplido | `hu060_citas.sql` ("DEC-007 OK", medianoche y cambio de horario de verano con la sesión en otra zona). |
| `CA-060-05` | Cumplido | `TestCreateInternal_NewCustomer_PersistsAppointmentAndHistory`, `TestCreateInternal_ScheduleConflict_RollsBackNewCustomer`, `TestCreateInternal_ContextCancelled_NoPartialWrite`. |
| `CA-060-06` | Cumplido | `hu060_citas.sql` ("CHECK/append-only OK": `UPDATE`/`DELETE` rechazados con `insufficient_privilege` en ambas tablas). |
| `CA-060-07` | Cumplido | `hu060_citas.sql` ("RLS OK", "grants OK", "RN-TEN-01 OK"), `TestCreateInternal_ExistingCustomerFromOtherTenant_ReturnsNotFound`. |
| `CA-060-08` | Cumplido | `atlas migrate hash/validate/apply` desde vacío y desde la versión anterior con datos representativos (14 barberías/9 barberos preexistentes, verificados intactos tras aplicar), segundo `apply` sin cambios, `atlas migrate status` limpio (16/16, "Already at latest version"). |

### Pruebas

- Dominio: `internal/modules/booking/domain_test.go` (validación de forma pura, sin base de datos).
- PostgreSQL real: `internal/modules/booking/postgres/repository_test.go`, incluida la carrera real de dos conexiones (`TestCreateInternal_ConcurrentOverlap_ExactlyOneSucceeds`, `-race`).
- SQL directo con el rol real: `database/tests/hu060_citas.sql` (esquema, `CHECK`, exclusión, `occupies_schedule`, medianoche/DST, historial append-only, RLS, grants, aislamiento de tenant, `ON DELETE RESTRICT`).
- Sin HTTP ni router de producción: HU-060 no expone ningún endpoint.

## Pruebas de integración

Las pruebas en `internal/platform/database/*_test.go` requieren PostgreSQL real
con las migraciones aplicadas y `database/testdata/dos_barberias.sql` cargado.
Conéctate como `barberia_app` (rol SIN `BYPASSRLS`, SIN propiedad).

Ejecución:

```bash
export TEST_DATABASE_URL="postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"
go test -race ./internal/platform/database/...
```

Desde HU-007, `internal/modules/auth/postgres/purge_repository_test.go`
requiere ADEMÁS `TEST_WORKER_DATABASE_URL` conectado como `barberia_worker`
(las funciones de purga no están concedidas a `barberia_app`, `DDL-AUT-01`):

```bash
export TEST_WORKER_DATABASE_URL="postgres://barberia_worker@localhost:5432/barberia_test?sslmode=disable"
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

`internal/modules/auth/postgres/repository_test.go` requiere, además de
`database/testdata/dos_barberias.sql`,
`database/testdata/hu005_credenciales_sesiones.sql` cargado (credenciales y
una sesión vigente por barbería). Cubre resolución de tenant (activo,
inactivo, inexistente), búsqueda de credencial (tenant real y señuelo,
cruce de tenant rechazado por RLS) y persistencia de sesión aislada por
barbería.

```bash
psql "$TEST_DATABASE_URL" -v ON_ERROR_STOP=1 -f database/testdata/hu005_credenciales_sesiones.sql
go test -race ./internal/modules/auth/...
```

`internal/modules/auth/postgres/session_repository_test.go` (HU-006) usa el
mismo `barberia_app` y la misma testdata; crea sus propias sesiones con
tokens únicos por prueba (no depende de las sesiones fijas del fixture) y
cubre renovación deslizante, vencimiento, revocación, usuario inactivo,
tenant equivocado y la carrera real renovación/revocación bajo `-race`.

`cmd/api/session_integration_test.go` (HU-006) construye el router REAL de
producción (`buildRouter`, el mismo que usa `run()`) contra PostgreSQL real:
la prueba estructural de `CA-006-04` (inventario de rutas privadas vía
`chi.Walk`), persistencia de sesión (`CA-006-01`), reutilización tras logout
(`CA-006-02`), cierre en un dispositivo sin afectar otro (`CA-006-06`) y
aislamiento entre barberías contra el logout real (`CA-006-07`).

```bash
go test -race ./internal/modules/auth/postgres/...
go test -race ./cmd/api/...
```

## Migraciones

Gobernadas por Atlas CLI v1.3.0 (fijado, no `latest`). Ver `database/README.md`.

```bash
atlas migrate hash --dir file://database/migrations
atlas migrate validate --env local
atlas migrate apply --env local
```

`atlas.sum` se versiona. NO se ejecutan al arrancar la aplicación (CA-001-07).