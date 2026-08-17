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
| `migrations/atlas.sum` | Generado y validado con Atlas v1.3.0 |
| `testdata/dos_barberias.sql` | Escenario de HU-001 con dos barberías |
| `testdata/hu005_credenciales_sesiones.sql` | HU-005 · credenciales y una sesión vigente por barbería, sobre `dos_barberias.sql` |
| `testdata/hu007_reto_telefonico.sql` | HU-007 · teléfono verificado para dos usuarios (uno por barbería) sobre `dos_barberias.sql`/`hu005_credenciales_sesiones.sql`; deja otros dos sin verificar a propósito |
| `testdata/notification_lease_fixture.sql`, `testdata/customer_anonymization_fixture.sql`, `testdata/rls_suite_fixture.sql` | Fixtures de `modelo-fisico-referencia.sql` (issues #5, #6, #7) — no dependen de migraciones aplicadas más allá de las de arriba |
| `tests/hu001_aislamiento_rls.sql` | CA-001-01 a CA-001-06 con el rol real |
| `tests/hu005_aislamiento_credenciales_sesiones.sql` | HU-005 · `staff_credential`/`staff_session`/funciones `SECURITY DEFINER` con dos tenants y el rol real |
| `tests/hu007_defensa_abuso.sql` | HU-007 · `DDL-AUT-01` (sin DML directo de `barberia_app`, purga exclusiva de `barberia_worker`), `DEC-061` (sexta solicitud escala) y `DEC-062` (reto telefónico: tres condiciones, código atado a su IP, verificar limpia el escalamiento) con el rol real |
| `tests/rls_suite.sql` | Suite RLS completa de las 23 tablas restantes (issue #7, `DDL-RLS-01`) |
| `seeds/` | Vacío |
| `modelo-fisico-referencia.sql` | Diseño completo de B1–B6 y del prerrequisito de teléfono de A.3 (HU-008). **No es una migración** |

Están aplicadas las siete migraciones listadas arriba. El resto del
modelo —barberos, servicios, horario, bloqueos, citas, clientes,
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
