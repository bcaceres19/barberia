---
prompt_id: "PROMPT-HU-021-v1"
version: "1.0"
kind: "hu"
status: "draft"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-021"
related_hu:
  - "HU-020"
  - "HU-002"
  - "HU-003"
  - "HU-004"
  - "HU-006"
  - "HU-009"
  - "HU-012"
issue: 68
issue_url: "https://github.com/bcaceres19/barberia/issues/68"
suggested_issue_title: "feat(staff): implementar HU-021 registro y listado de barberos"
branch: null
pr: null
pr_url: null
depends_on:
  - "Criterio de salida de B0 verificado contra docs/10-backlog/plan-bloques.md"
  - "HU-020 integrada en main con sus ocho criterios verificados"
  - "Issue real de HU-021 creado con CA-021-01 a CA-021-08"
rules:
  - "RN-TEN-01"
  - "RN-DAT-02"
  - "RN-IDE-01"
  - "Criterios no funcionales de UX y accesibilidad"
decisions:
  - "DEC-019"
  - "DEC-024"
  - "DEC-033"
  - "DEC-034"
  - "DEC-035"
  - "DEC-036"
  - "DEC-037"
  - "DEC-038"
  - "DEC-039"
  - "DEC-040"
  - "DEC-043"
  - "DEC-047"
acceptance_criteria:
  - "CA-021-01"
  - "CA-021-02"
  - "CA-021-03"
  - "CA-021-04"
  - "CA-021-05"
  - "CA-021-06"
  - "CA-021-07"
  - "CA-021-08"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/04-arquitectura/backend-go.md"
  - "docs/04-arquitectura/frontend.md"
  - "docs/05-backend/base-datos.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/05-backend/revision-ddl-seguridad-2026-08-11.md"
  - "docs/06-api/estandar-openapi.md"
  - "docs/10-backlog/plan-bloques.md"
  - "docs/10-backlog/prompts/hu/hu-020-configuracion-barberia.md"
  - "database/README.md"
  - "database/modelo-fisico-referencia.sql"
  - "database/migrations"
  - "database/tests/rls_suite.sql"
  - "api/openapi/README.md"
  - "api/openapi/openapi.yaml"
  - "api/openapi/paths/settings.yaml"
  - "apps/api/README.md"
  - "apps/api/internal/modules/staff/doc.go"
  - "apps/api/internal/platform/database/database.go"
  - "apps/api/internal/platform/idempotency"
  - "apps/api/internal/modules/auth/session_service.go"
  - "apps/api/cmd/api/main.go"
  - "apps/web/README.md"
  - "apps/web/src/app/router/index.ts"
  - "apps/web/src/modules/auth"
  - "apps/web/src/modules/settings"
  - "apps/web/src/modules/staff/index.ts"
  - "apps/web/src/shared/api"
  - "apps/web/src/shared/ui"
created_at: "2026-08-23"
updated_at: "2026-08-23"
supersedes: "docs/10-backlog/prompts-implementacion.md#prompt-de-hu-021--registro-y-listado-de-barberos"
superseded_by: null
---

# Implementar HU-021: registro y listado de barberos

## Instrucción para Claude o Codex

Implementa únicamente `HU-021` como una entrega vertical contract-first para consultar, listar, registrar y renombrar barberos de la barbería activa. Usa exactamente el mismo modelo, API y pantalla para una barbería con una persona y para otra con cuatro. No agregues borrado, activación/desactivación, orden manual, credenciales, relación con `staff_user`, servicios, horarios, disponibilidad ni agenda.

Issue real creado: [#68](https://github.com/bcaceres19/barberia/issues/68), con `CA-021-01` a `CA-021-08` enlazados a `docs/02-requisitos/historias-usuario.md`. Este prompt conserva `status: draft` porque `HU-020` (issue [#67](https://github.com/bcaceres19/barberia/issues/67)) todavía no está integrada en `main`: HU-021 no puede pasar a `ready` hasta que ese PR se mergee y el patrón de composición de rutas/navegación que deja HU-020 exista realmente.

## Objetivo

Que un barbero autenticado abra “Barberos”, vea únicamente el equipo de su barbería, agregue una persona y cambie su nombre. La barbería unipersonal conserva una fila `barber` ordinaria; añadir tres personas produce cuatro recursos distintos sin cambiar schema, endpoint, componentes ni camino de código. Conocer un UUID de otro tenant nunca revela ni modifica ese recurso.

## Preflight obligatorio

1. Comprueba árbol limpio, actualiza `main` por fast-forward y no sobrescribas cambios ajenos.
2. Verifica el criterio de salida de B0 y que el PR de HU-020 está integrado, sus checks siguen verdes y su contrato/migración/cliente/pantalla reales coinciden con la documentación. No avances usando únicamente el prompt de HU-020 como prueba de implementación.
3. Ejecuta consultas Graphify sobre `HU-021`, `staff`, `barber`, `InTenantTx`, `idempotency.Coordinator`, rutas privadas, composición de rutas de HU-020 y RLS. Usa `graphify path`/`explain` para los límites entre `auth`, `staff`, `app` y `shared`.
4. Lee completamente cada `source_docs`. En rutas de directorio revisa API pública, archivos relevantes y pruebas; no edites código generado, cachés ni migraciones aplicadas.
5. Confirma el estado real: la cadena de migraciones aplicada aún no contiene `barber`; la sección B.2 de `modelo-fisico-referencia.sql` y partes de `rls_suite.sql` son diseño/pruebas de referencia, no evidencia de una tabla migrada; los módulos Go/Vue `staff` siguen vacíos salvo que HU-020 los haya cambiado justificadamente.
6. Revisa dudas y contradicciones vigentes. `DEC-047` es obligatoria: retira de esta HU cualquier borrado, activación/desactivación, orden o vínculo automático con usuarios aunque el modelo de referencia histórico o código futuro lo sugiera.
7. Crea o localiza el issue con `CA-021-01` a `CA-021-08`. Actualiza número/URL y el índice; cambia a `ready` solo con dependencias satisfechas. Al iniciar crea `feat/<issue>-hu021-barberos` desde `main` actualizada y cambia a `in_progress`.

## Alcance incluido

- Entidad `barber` separada de `barbershop` y `staff_user`, con `id`, `barbershop_id`, `full_name`, `created_at` y `updated_at`.
- Contrato privado para lista paginada, lectura individual, alta y actualización parcial limitada al nombre. La lectura individual permite demostrar de forma directa el `404` de `CA-021-05`.
- Migración Atlas con FK `ON DELETE RESTRICT`, unicidad tenant-aware necesaria para relaciones futuras, trigger de `updated_at`, RLS forzada, políticas y grants mínimos, sin permiso ni política `DELETE`.
- Módulo Go `staff` con dominio, puertos, servicios, repositorio PostgreSQL y handlers HTTP independientes de Chi/pgx en su núcleo.
- Pantalla “Barberos” dentro del cascarón privado: lista, paginación cuando aplique, vacío, alta y edición del nombre, con estados y errores accesibles.
- Idempotencia de la creación mediante el mecanismo reutilizable de HU-004: repetir el mismo `POST` con la misma clave y contenido devuelve el mismo recurso sin crear una segunda fila; otro nombre con la misma clave produce conflicto. Esto no impone unicidad de nombres.
- Pruebas con PostgreSQL real y dos tenants, escenarios idénticos de una y cuatro personas, contrato/HTTP, componentes, E2E y evidencia responsive/accesible.

## Fuera de alcance

- Eliminar, archivar, desactivar, reactivar u ordenar barberos; no agregues columnas, endpoints, controles ni grants para esos comportamientos.
- Crear cuentas, credenciales o invitaciones; asignar roles/permisos; asociar automática o manualmente `staff_user` con `barber`.
- Correo, teléfono, fotografía, biografía, comisión u otros datos personales/profesionales del barbero.
- Servicios, `barber_service`, precios, duración, horario laboral, festivos, bloqueos, disponibilidad, selección pública o agenda.
- Inferir el barbero desde la sesión o desde `barbershop`; crear una rama `singleBarber`; usar el nombre de barbería o de usuario como sustituto de `barber.full_name`.
- Unicidad de `full_name`. Dos personas pueden compartir nombre; una restricción única o deduplicación por texto inventaría una regla.
- Nueva biblioteca de tablas/formularios, Pinia, estado global paralelo o estilos fuera de tokens/componentes aprobados.

## Estado existente que debes conservar

- `auth.Principal` es la única fuente del `barbershop_id` autenticado y el subrouter privado ya exige sesión. Ningún request de esta HU acepta tenant como autoridad.
- `database.DB.InTenantTx` es la única vía de consultas tenant-aware; cada consulta filtra explícitamente por `barbershop_id` además de RLS.
- `internal/platform/idempotency` implementa `RN-IDE-01`/`DEC-043`. `Begin`, creación y `Complete` deben compartir una sola `InTenantTx`; no repartas el advisory lock y el efecto entre transacciones.
- HU-020 debe haber dejado un patrón real para componer rutas/navegación privadas y coordinar módulos por sus `index.ts`; reutilízalo sin crear un segundo cascarón ni hacer imports de archivos internos de otro módulo.
- HU-009 aporta componentes y tokens; HU-012 aporta guard, cabecera, navegación y coordinación única de `401`.
- `modelo-fisico-referencia.sql` B.2 ya refleja `DEC-047`: úsalo como insumo revisable para la forma mínima de `barber`, nunca como migración aplicada ni como autorización para copiar secciones posteriores.
- `database/tests/rls_suite.sql` se construyó sobre el modelo físico completo. Conserva su cobertura útil, pero crea pruebas ejecutables de HU-021 sobre la cadena real de migraciones; no cites la suite de referencia como prueba de la nueva migración si no se ejecutó contra ella.
- El cliente TypeScript deriva del bundle OpenAPI y no se edita manualmente.

## Trabajo requerido

### 1. Contrato HTTP antes del código

1. Define una colección privada `GET /private/barbers` y `POST /private/barbers`, y un recurso `GET /private/barbers/{barberId}` y `PATCH /private/barbers/{barberId}`. Si la estructura de archivos vigente tras HU-020 exige otro archivo OpenAPI, conserva la separación por capacidad sin mezclar `staff` con servicios/agenda.
2. Declara tag `Staff` o el nombre estable adoptado en el contrato, `SessionCookie`, JSON `camelCase`, `operationId` por intención, ejemplos ficticios, `x-business-rules` (`RN-TEN-01`, `RN-DAT-02`, y `RN-IDE-01` en el alta) y `x-decisions` (`DEC-019`, `DEC-024`, `DEC-047`, más las transversales pertinentes).
3. `POST` recibe únicamente `fullName`, exige `Idempotency-Key`, responde `201` con `Location` y la representación creada; documenta replay/conflicto/operación en curso según el protocolo real de HU-004. La clave no convierte el nombre en único: dos claves distintas pueden crear dos barberos con el mismo nombre.
4. `PATCH` permite únicamente `fullName`, rechaza objetos vacíos y campos desconocidos y devuelve `200` con la representación actualizada. No expongas `active`, `deletedAt`, `sortOrder`, `staffUserId`, servicios u horarios en ningún schema.
5. La lista usa paginación por cursor porque la colección no tiene un máximo de negocio aprobado. Define `limit` con límites técnicos justificados, cursor opaco y orden estable (por ejemplo `created_at, id`); no uses offset ni inventes un máximo de barberos por barbería. El frontend debe poder recorrer páginas sin perder filas.
6. Documenta `200`/`201`, `400`, `401`, `404`, `409` donde aplica idempotencia, `422` y `500` con componentes RFC 9457; `404` es idéntico para inexistente y otro tenant. Regenera el cliente solo tras lint/bundle verdes.

### 2. Migración y seguridad de datos

1. Crea con Atlas una migración nueva que contenga únicamente la tabla mínima `barber` de HU-021 y sus objetos inmediatos. No copies `service`, `barber_service`, horarios ni otras secciones del modelo físico.
2. Implementa UUID PK, `barbershop_id NOT NULL`, FK a `barbershop(id) ON DELETE RESTRICT`, `UNIQUE (barbershop_id, id)` para FKs tenant-aware futuras, `full_name NOT NULL`, timestamps con zona, límite de 120 y `btrim(full_name) <> ''`, trigger `set_updated_at` e índices mínimos guiados por consultas reales.
3. No agregues unicidad por nombre. El servicio almacena el nombre recortado; el `CHECK` es defensa estructural, no sustituto de normalización.
4. Habilita y fuerza RLS. Crea política administrativa para `barberia_owner` y políticas separadas `SELECT`, `INSERT`, `UPDATE` para `barberia_app`, todas contra `current_setting('app.barbershop_id')`; concede solo esos tres verbos. No crees política ni `GRANT DELETE`.
5. Valida estructura y comportamiento con `barberia_app` real: sin contexto falla cerrado; A no ve/escribe/renombra B; `WITH CHECK` rechaza cambiar el tenant; `DELETE` se deniega; nombre inválido se rechaza; el rol no tiene `BYPASSRLS`.
6. Prueba aplicación desde cero y upgrade desde la versión con HU-020 y datos existentes. Evalúa bloqueo/recuperación, valida `atlas.sum` y documenta roll-forward; nunca edites una migración aplicada.
7. Crea fixture y prueba SQL propios de HU-021 con dos barberías y nombres sintéticos. Actualiza el inventario de migraciones, testdata y pruebas.

### 3. Backend Go

1. Implementa `internal/modules/staff` con entidad/valor de nombre, errores, puertos y servicios de `ListBarbers`, `GetBarber`, `CreateBarber` y `RenameBarber`; adaptadores en subpaquetes `postgres` y `httpapi`.
2. Recorta `fullName`, rechaza vacío/solo espacios/más de 120 caracteres y preserva caracteres Unicode válidos. No exige dos palabras ni unicidad.
3. Los handlers extraen `auth.Principal`; los servicios reciben identificadores opacos/tipados y nunca Chi, `net/http`, pgx, DTO o `database.DB`. El repositorio ejecuta todo dentro de `InTenantTx` y cada `SELECT`/`UPDATE` filtra además por `barbershop_id`.
4. Para `CreateBarber`, valida `Idempotency-Key`/huella en HTTP y ejecuta `Coordinator.Begin`, `INSERT` y `Complete` dentro de la misma transacción. Prueba replay byte a byte, clave reutilizada con otro contenido, ejecución concurrente y misma clave literal en dos tenants.
5. `GetBarber` y `RenameBarber` convierten tanto “no existe” como “pertenece a otro tenant” en el mismo `apperr.NotFound`; ningún error o tiempo observable revela la fila ajena. El `UPDATE ... RETURNING` renombra una sola fila y nunca crea/duplica.
6. La lista aplica cursor y orden estable dentro del tenant. Prueba páginas contiguas sin duplicados/omisiones, incluidos nombres iguales y timestamps cercanos.
7. Registra rutas únicamente sobre el subrouter privado, amplía el inventario estructural y usa `httpserver.Translate`/`WriteProblem`. No registres nombres, cuerpos, correo, teléfono, cookie, cursor crudo sensible ni datos personales.

### 4. Frontend Vue

1. Convierte `modules/staff` en un módulo real con API pública mínima, ruta diferida, funciones del cliente por intención, modelos discriminados, validación, página y componentes/pruebas.
2. Compón `/panel/barberos` como hija del cascarón existente y agrega “Barberos” a la navegación usando el patrón público entregado por HU-020. No importes archivos internos de `auth`/`settings`, no repitas guard y no descargues el módulo en el flujo público.
3. Implementa lista con claves por `barber.id`, carga inicial, estado vacío, error recuperable y paginación accesible si existe siguiente cursor. Una y cuatro filas usan el mismo componente y el mismo estado de datos.
4. Implementa alta y edición del nombre con formulario o diálogo accesible. Deshabilita mientras envía, evita doble toque y conserva el texto ante `422` o error recuperable. El alta genera/reutiliza una clave de idempotencia por intento lógico según el cliente de HU-004; un reintento de red del mismo intento no crea otra fila.
5. Tras alta exitosa, incorpora una sola vez el recurso confirmado en la lista; tras renombrar, reemplaza por `id` sin duplicar ni reordenar de forma inestable. No hagas actualización optimista que finja éxito antes del servidor.
6. No muestres placeholders de servicio, horario, disponibilidad, estado activo, usuario vinculado o agenda. La pantalla administra únicamente nombres.
7. Usa componentes/tokens aprobados, etiquetas y errores asociados, resumen/foco cuando aplique, anuncios de estado, teclado completo y objetivos táctiles ordinarios de 44 px.

## Pruebas y evidencia

- Dominio/servicio: nombre vacío, espacios, 120/121 caracteres, Unicode, recorte, nombres duplicados permitidos, renombrado por `id` sin duplicación.
- PostgreSQL real: esquema mínimo, FK/unique tenant-aware, RLS forzada, permisos exactos, `DELETE` denegado, A/B aislados, ausencia de contexto, inserción y cambio cruzados rechazados, rol real sin `BYPASSRLS`.
- Idempotencia: mismo `POST`/clave/contenido crea una fila y reproduce respuesta; misma clave/otro contenido da conflicto; dos conexiones concurrentes producen un efecto; misma clave literal en A y B no interfiere.
- HTTP/contrato: lista/paginación, lectura, alta `201 + Location`, renombrado, forma inválida, `401`, `404` indistinguible para ajeno/inexistente, `409`, `422`, `500`, campos desconocidos y ausencia total de operaciones fuera de alcance.
- Privacidad: logs de alta/lista/renombrado exitosos y fallidos no contienen `fullName`, body, cookie ni datos personales; usa identificadores opacos y `request_id`.
- Componente: carga, vacío, una fila, cuatro filas, página siguiente, alta, edición, `422`, error recuperable, doble envío, foco y ausencia de campos/acciones excluidos.
- E2E real: iniciar sesión → abrir “Barberos” → añadir → listar → renombrar → recargar; repite con una barbería inicialmente unipersonal y otra que termina con cuatro, por el mismo camino de UI/API. Incluye intento de consultar/editar un UUID de otro tenant y verifica `404` sin fuga.
- Evidencia en 320, 360, 768 y 1280 px, zoom 200 %, teclado, foco visible/no oculto, objetivos táctiles y axe-core; inspecciona capturas antes de conservarlas.

Entrega una tabla `Criterio | Estado | Prueba o evidencia` para `CA-021-01` a `CA-021-08` y una lista explícita de exclusiones inspeccionadas en contrato, migración, grants, Go y Vue.

## Documentación y trazabilidad

- Actualiza OpenAPI/CHANGELOG, README de API/web/base de datos, inventarios de migraciones/testdata/pruebas, matriz e historial según los artefactos reales.
- Si el DDL final difiere justificadamente de B.2, actualiza `modelo-fisico-referencia.sql` y explica la diferencia; no propagues columnas fuera de alcance.
- Actualiza este prompt y el catálogo con issue, rama, PR, estado y evidencia reales al iniciar, abrir PR, ejecutar, bloquear o sustituir.
- No declares completado B1: HU-021 solo cubre la parte de barberos de `F-CONF-02`; aún faltan las historias de servicios que el plan del bloque exige redactar.

## Verificación final

```text
# raíz
pnpm run openapi:check-config
pnpm run openapi:lint
pnpm run openapi:bundle

# database (Atlas v1.3.0 y PostgreSQL real)
atlas migrate hash --dir file://database/migrations
atlas migrate validate --env local
atlas migrate status --env local
atlas migrate apply --env local --dry-run
# aplicar en base vacía y actualizar desde HU-020
# ejecutar database/tests/hu021_barberos.sql con barberia_app real

# apps/api
gofmt -l .
go vet ./...
go test -race ./...
go build ./...

# apps/web
pnpm run generate:api
pnpm run format
pnpm run lint
pnpm run typecheck
pnpm run test:unit
pnpm run build
pnpm run test:e2e

# raíz
git diff --check
graphify update .
```

Documenta cómo se preparó PostgreSQL sin exponer DSN, claves o secretos. No reduzcas umbrales, no ocultes pruebas inestables, no sustituyas concurrencia/RLS por mocks y no uses las pruebas del modelo físico como evidencia si no corrieron sobre la cadena migrada.

## Git y PR

- Rama esperada después de asignar issue: `feat/<issue>-hu021-barberos`.
- Commits y título del PR: `feat(staff): implementa HU-021 registro y listado de barberos` o equivalente Conventional Commits para el corte vertical completo.
- Usa `Closes #<issue>` solo si `CA-021-01`–`CA-021-08` y exclusiones están verificadas; de lo contrario usa `Refs #<issue>` y conserva el estado real.
- Abre PR contra `main` con tabla de criterios, evidencia responsive/accesible, prueba de dos tenants, idempotencia, riesgo de migración y recuperación roll-forward.
- No hagas push directo, force push, merge manual, edición de migraciones aplicadas ni reescritura de `main`.
