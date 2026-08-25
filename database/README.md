# database

Modelo PostgreSQL, migraciones, seeds y pruebas de datos del MVP. Ver
[`docs/05-backend/estandar-base-datos.md`](../docs/05-backend/estandar-base-datos.md)
y [`docs/05-backend/migraciones-atlas.md`](../docs/05-backend/migraciones-atlas.md)
antes de agregar una migración.

## Estado

| Artefacto | Estado |
| --- | --- |
| `atlas.hcl` | Cuatro ambientes: `local`, `test`, `pilot`, `production` |
| `migrations/20260807170000_create_tenant_foundation.sql` | HU-001 · extensiones, roles, `barbershop`, `staff_user`, RLS |
| `migrations/20260807170100_create_idempotency_record.sql` | HU-004 · `idempotency_record` |
| `migrations/20260811145252_harden_roles_and_definer_functions.sql` | Correctiva · `barberia_owner`/`barberia_worker`, ownership, `DEC-040` |
| `migrations/20260811154100_harden_idempotency_concurrency.sql` | Correctiva · protocolo de idempotencia concurrente (`DEC-043`) |
| `migrations/20260811220000_harden_idempotency_fingerprint_format.sql` | Correctiva · `request_fingerprint` hexadecimal en minúsculas (`DDL-VAL-01`) |
| `migrations/20260813120000_create_auth_credentials_and_sessions.sql` | HU-005 · `staff_credential`, `staff_session`, `authn_resolve_login_tenant`, `auth_get_credential`, `authn_resolve_session_tenant` (secciones A.0–A.2 del modelo de referencia) |
| `migrations/20260817180000_create_login_throttle_and_phone_challenge.sql` | HU-007 · `staff_user.phone`/`phone_verified_at` (adelantado de A.3), `login_throttle`, `auth_phone_challenge` y sus funciones `SECURITY DEFINER` (secciones A.4–A.5 del modelo de referencia) |
| `migrations/20260817190000_create_staff_recovery_code.sql` | HU-008 · `staff_recovery_code` y sus cuatro funciones `SECURITY DEFINER` (`auth_recovery_request`/`verify`/`current_credential`/`change_password`/`purge_expired`), sección A.3 adaptada del modelo de referencia |
| `migrations/20260823120000_add_barbershop_contact_info.sql` | HU-020 · `barbershop.contact_email`/`contact_phone` opcionales, mismos patrones de forma que `staff_user.email`/`.phone`; sin `GRANT` nuevo (privilegio de tabla ya cubre las columnas) |
| `migrations/20260823130000_create_barber.sql` | HU-021 · tabla `barber` mínima (id, `barbershop_id`, `full_name`, timestamps), copiada de la sección B.2 de `modelo-fisico-referencia.sql` (`DEC-047`): RLS forzada, política administrativa a `barberia_owner`, políticas `SELECT`/`INSERT`/`UPDATE` para `barberia_app`, sin política ni `GRANT DELETE` |
| `migrations/20260824140000_create_service.sql` | HU-022 · tabla `service` mínima (id, `barbershop_id`, `name`, `description`, `duration_minutes`, `price_amount` `numeric(12,2)`, `price_currency` fija en `COP`, `is_active`/`deactivated_at` con `service_deactivated_at_ck`; HU-024 expone su transición vía `POST .../deactivate`/`.../reactivate`, sin migración nueva: la columna ya existía), basada en la sección B.3 de `modelo-fisico-referencia.sql` con tres diferencias exigidas por `DEC-067` (precio `> 0`, moneda fija a `COP`, índice único sin `lower(name)`): RLS forzada, índice único parcial `idx_service_active_name` `(barbershop_id, name) WHERE is_active`, sin política ni `GRANT DELETE` |
| `migrations/20260824150000_create_barber_service.sql` | HU-023 · tabla de asociación PURA `barber_service` (`barbershop_id`, `barber_id`, `service_id`, `created_at`; sin nombre/duración/precio/estado), basada en la sección B.4 de `modelo-fisico-referencia.sql` con dos diferencias: FK compuestas con `ON DELETE RESTRICT` (no `CASCADE`: `barber`/`service` no exponen `DELETE` al rol de aplicación, ese `CASCADE` nunca podría dispararse) e índices propios para la consulta por barbero (`CA-023-01`) y el conteo de `DEC-068`; RLS forzada, política administrativa a `barberia_owner`, políticas `SELECT`/`INSERT`/`DELETE` para `barberia_app` (sin `UPDATE`: nada editable) |
| `migrations/atlas.sum` | Generado y validado con Atlas v1.3.0 |
| `testdata/dos_barberias.sql` | Escenario de HU-001 con dos barberías |
| `testdata/hu005_credenciales_sesiones.sql` | HU-005 · credenciales y una sesión vigente por barbería, sobre `dos_barberias.sql` |
| `testdata/hu007_reto_telefonico.sql` | HU-007/HU-008 · teléfono verificado para dos usuarios (uno por barbería) sobre `dos_barberias.sql`/`hu005_credenciales_sesiones.sql`; deja otros dos sin verificar a propósito |
| `testdata/hu021_barberos.sql` | HU-021 · dos barberías DEDICADAS (`33333333.../44444444...`) a las pruebas de `staff/postgres` y `cmd/api`, separadas de `dos_barberias.sql` para que un conteo exacto de barberos en una página no dependa de otras suites |
| `testdata/hu022_catalogo.sql` | HU-022 · dos barberías DEDICADAS (`55555555.../66666666...`) a las pruebas de `catalog/postgres` y `cmd/api`, separadas de `dos_barberias.sql`/`hu021_barberos.sql` por el mismo motivo |
| `testdata/hu023_asignaciones.sql` | HU-023 · dos barberías DEDICADAS (`77777777.../88888888...`) a las pruebas de `catalog/postgres.AssignmentRepository` y `cmd/api`, separadas de `dos_barberias.sql`/`hu021_barberos.sql`/`hu022_catalogo.sql` por el mismo motivo |
| `testdata/notification_lease_fixture.sql`, `testdata/customer_anonymization_fixture.sql`, `testdata/rls_suite_fixture.sql` | Fixtures de `modelo-fisico-referencia.sql` (issues #5, #6, #7) — no dependen de migraciones aplicadas más allá de las de arriba |
| `tests/hu001_aislamiento_rls.sql` | CA-001-01 a CA-001-06 con el rol real |
| `tests/hu005_aislamiento_credenciales_sesiones.sql` | HU-005 · `staff_credential`/`staff_session`/funciones `SECURITY DEFINER` con dos tenants y el rol real |
| `tests/hu007_defensa_abuso.sql` | HU-007 · `DDL-AUT-01` (sin DML directo de `barberia_app`, purga exclusiva de `barberia_worker`), `DEC-061` (sexta solicitud escala) y `DEC-062` (reto telefónico: tres condiciones, código atado a su IP, verificar limpia el escalamiento) con el rol real |
| `tests/hu008_recuperacion_acceso.sql` | HU-008 · `DDL-AUT-01` (sin DML directo de `barberia_app` sobre `staff_recovery_code`, purga exclusiva de `barberia_worker`), no enumeración de `auth_recovery_request`, cooldown/límite de reenvío y unicidad del código activo (`CA-008-07`), intentos/agotamiento de `auth_recovery_verify` (`CA-008-02`/`03`) y verificación + cambio de contraseña con revocación total de sesiones en la misma operación (`CA-008-05`), todo envuelto en `BEGIN...ROLLBACK` para no dejar estado persistido |
| `tests/hu020_configuracion_barberia.sql` | HU-020 · columnas anulables, contacto válido/vacío (`CA-020-06`), `CHECK` de correo/teléfono, aislamiento de tenant sobre las columnas nuevas (`CA-020-05`) y por qué la validación IANA (`CA-020-03`) no vive en un `CHECK`, todo con el rol real |
| `tests/hu021_barberos.sql` | HU-021 · esquema mínimo sin ciclo de vida (`DEC-047`), RLS forzada/sin `BYPASSRLS`/sin `GRANT` ni política `DELETE` (`CA-021-06`/`07`), alta de 1 y de 4 barberos en la misma tabla (`CA-021-01`/`02`), aislamiento de tenant en lectura/escritura cruzada con id real (`CA-021-05`), `barber_full_name_ck` (`CA-021-03`) y el trigger `updated_at`, todo con el rol real |
| `tests/hu022_catalogo.sql` | HU-022 · esquema con precisión monetaria `numeric(12,2)` (nunca `real`/`double precision`), RLS forzada/sin `BYPASSRLS`/sin `GRANT` ni política `DELETE` (RN-SER-03), varios servicios distintos en el mismo catálogo (`CA-022-01`), aislamiento de tenant en lectura/escritura cruzada con id real (`CA-022-06`), nombre único entre servicios activos y su reutilización tras desactivar (`DEC-067`), `CHECK` de duración/precio/moneda (`CA-022-03`/`04`) y el trigger `updated_at`, todo con el rol real |
| `tests/hu023_asignaciones.sql` | HU-023 · esquema exacto de `barber_service` sin atributos duplicados (`CA-023-07`), PK/FK compuestas, RLS forzada/sin `BYPASSRLS`, grants exactos `SELECT`/`INSERT`/`DELETE` sin `UPDATE`, FK compuesta rechaza asociar un barbero y un servicio de barberías distintas aun con el rol real (`CA-023-04`), aislamiento de tenant en lectura (`CA-023-06`), asignación repetida sin duplicar y un servicio compartido por varios barberos (`CA-023-02`/`03`), todo con el rol real |
| `tests/hu024_ciclo_vida.sql` | HU-024 · `service_deactivated_at_ck` rechaza `is_active=false`/`deactivated_at NULL` y `is_active=true`/`deactivated_at` fijado en ambas direcciones, `barberia_app` transiciona activo↔inactivo sobre la MISMA fila sin duplicar (`CA-024-02`/`05`), aislamiento de tenant sobre la transición con id real (`CA-024-07`), y desactivar libera el nombre para un servicio activo nuevo ejercido con la transición real de `barberia_app` (`DEC-067`); RLS forzada, ausencia de `GRANT`/política `DELETE` y el esquema mínimo de `service` ya están cubiertos por `tests/hu022_catalogo.sql`, no se repiten aquí; la orquestación completa (idempotencia, transición inválida, la carrera real de dos conexiones) vive en `internal/modules/catalog/postgres`, todo con el rol real |
| `tests/rls_suite.sql` | Suite RLS completa de las 23 tablas restantes (issue #7, `DDL-RLS-01`) |
| `seeds/` | Vacío |
| `modelo-fisico-referencia.sql` | Diseño completo de B1–B6. **No es una migración** |

Están aplicadas las doce migraciones listadas arriba. El resto del
modelo —servicios, horario, bloqueos, citas, clientes,
notificaciones— vive en
[`modelo-fisico-referencia.sql`](modelo-fisico-referencia.sql) y se
convierte en migración cuando se abre la historia que lo necesita. Escribirlo
todo hoy como migraciones lo volvería inmutable antes de tener una sola línea
de negocio funcionando (`DEC-036`).

## Roles

| Rol | Uso | Privilegios |
| --- | --- | --- |
| `barberia_migrator` | Propietario de los objetos; ejecuta Atlas y carga `testdata/` | DDL; política administrativa sobre las tablas con RLS forzada |
| `barberia_app` | Único rol de los procesos `api` y `worker` | DML dentro de su barbería; **sin** `BYPASSRLS`, sin propiedad, sin DDL |

Las migraciones crean ambos roles **sin contraseña**. Cada ambiente la fija
fuera del repositorio:

```bash
psql "$DATABASE_ADMIN_URL" -c "ALTER ROLE barberia_app PASSWORD '<secreto>'"
```

Toda transacción de la aplicación debe fijar el contexto antes de tocar datos:

```sql
BEGIN;
SET LOCAL app.barbershop_id = '<uuid de la barbería>';
-- ... trabajo ...
COMMIT;
```

Sin ese ajuste, cualquier consulta a una tabla protegida **falla**; no
devuelve el conjunto completo (`CA-001-03`).

## Versión de PostgreSQL

**El número soportado sigue pendiente de una decisión `DEC-*` explícita.**

Lo que sí está fijado por el diseño: la primera migración verifica en tiempo
de ejecución `server_version_num >= 140000` y aborta por debajo. Es el mínimo
técnico que exige `FORCE ROW LEVEL SECURITY`, `gen_random_uuid()` en el
núcleo y `btree_gist` sobre `uuid` para la restricción de exclusión de
`docs/05-backend/base-datos.md` sección 3. Esa comprobación es un piso, no
la decisión: no inventar una respuesta a una duda sin registrarla
(`AGENTS.md`).

## Herramienta

[Atlas CLI](https://atlasgo.io/) con flujo de migraciones SQL versionadas
(`DEC-036`). No se usa el estado declarativo de Atlas contra piloto o
producción.

**Versión fijada: `v1.3.0`.** `migraciones-atlas.md` §3.3 prohíbe usar
`latest` automáticamente. El canal `-latest` del sitio de descargas entrega
compilaciones *canary* (por ejemplo `v1.3.1-...-canary`), que no son
candidatas a un flujo reproducible. CI debe instalar exactamente esta versión
y actualizarla mediante revisión, no de forma automática.

Instalación en Windows:

```powershell
$dir = "$env:LOCALAPPDATA\Atlas\bin"
New-Item -ItemType Directory -Force $dir | Out-Null
Invoke-WebRequest -Uri "https://release.ariga.io/atlas/atlas-windows-amd64-v1.3.0.exe" `
  -OutFile "$dir\atlas.exe" -UseBasicParsing
# Agregar $dir al PATH de usuario
```

En Linux y macOS se usa el binario equivalente
`atlas-<os>-<arch>-v1.3.0`. No se usa el instalador `atlasgo.sh` sin fijar
versión.

## Comandos

Desde `database/`, con Atlas CLI instalado y `DATABASE_URL` (u otra variable
según el ambiente) exportada:

```bash
atlas migrate new <descripcion>
atlas migrate hash
atlas migrate validate --env local
atlas migrate status --env local
atlas migrate apply --env local --dry-run
atlas migrate apply --env local
```

### Primera aplicación

`atlas.sum` todavía no existe: los dos archivos de `migrations/` se escribieron
a mano y hay que sellarlos antes de aplicarlos.

```bash
cd database
atlas migrate hash                       # genera atlas.sum sobre los dos archivos
atlas migrate validate --env local
atlas migrate apply    --env local --dry-run
atlas migrate apply    --env local

psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f tests/hu001_aislamiento_rls.sql
```

`atlas.sum` se versiona en Git junto con las migraciones y nunca se regenera
para "arreglar" un conflicto: una migración ya compartida no se edita, se
corrige con otra (`DEC-036`).

### Agregar la siguiente migración

1. Abrir la historia y comprobar que no depende de una duda sin `DEC-*`.
2. `atlas migrate new <descripcion>`.
3. Copiar la sección correspondiente de `modelo-fisico-referencia.sql` y
   revisarla contra el estándar y las reglas que cita.
4. `atlas migrate hash` y `atlas migrate validate --env local`.
5. Escribir la prueba en `tests/` con dos barberías y el rol real.
6. Aplicar en vacío y sobre la versión anterior con datos.

## Recuperación

**Pendiente.** El procedimiento de copia diaria, retención de 30 días y
restauración probada se documenta aquí cuando exista la primera migración y
el mecanismo de respaldo elegido, según
`docs/04-arquitectura/stack-despliegue-operacion.md` sección 5. Antes del
piloto, este documento debe registrar fecha, duración, resultado y
responsable de al menos una restauración completa.
