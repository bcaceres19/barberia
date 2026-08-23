---
prompt_id: "PROMPT-HU-020-v1"
version: "1.0"
kind: "hu"
status: "in_progress"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-020"
related_hu:
  - "HU-003"
  - "HU-006"
  - "HU-009"
  - "HU-012"
  - "HU-021"
issue: 67
issue_url: "https://github.com/bcaceres19/barberia/issues/67"
suggested_issue_title: "feat(settings): implementar HU-020 configuración básica de la barbería"
branch: "feat/67-hu020-configuracion-barberia"
pr: null
pr_url: null
depends_on:
  - "Criterio de salida de B0 verificado contra docs/10-backlog/plan-bloques.md"
  - "HU-003, HU-006, HU-009 y HU-012 integradas en main"
  - "Issue real de HU-020 creado con CA-020-01 a CA-020-08"
rules:
  - "RN-DIS-07"
  - "RN-TEN-01"
  - "RN-DAT-02"
  - "Criterios no funcionales de UX y accesibilidad"
decisions:
  - "DEC-007"
  - "DEC-024"
  - "DEC-033"
  - "DEC-034"
  - "DEC-035"
  - "DEC-036"
  - "DEC-037"
  - "DEC-038"
  - "DEC-039"
  - "DEC-040"
  - "DEC-060"
acceptance_criteria:
  - "CA-020-01"
  - "CA-020-02"
  - "CA-020-03"
  - "CA-020-04"
  - "CA-020-05"
  - "CA-020-06"
  - "CA-020-07"
  - "CA-020-08"
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
  - "docs/06-api/estandar-openapi.md"
  - "docs/10-backlog/plan-bloques.md"
  - "database/README.md"
  - "database/modelo-fisico-referencia.sql"
  - "database/migrations/20260807170000_create_tenant_foundation.sql"
  - "api/openapi/README.md"
  - "api/openapi/openapi.yaml"
  - "api/openapi/paths/settings.yaml"
  - "api/openapi/paths/private-auth.yaml"
  - "apps/api/README.md"
  - "apps/api/internal/modules/shops/doc.go"
  - "apps/api/internal/platform/database/database.go"
  - "apps/api/internal/modules/auth/session_service.go"
  - "apps/api/cmd/api/main.go"
  - "apps/web/README.md"
  - "apps/web/src/app/router/index.ts"
  - "apps/web/src/modules/auth/index.ts"
  - "apps/web/src/modules/auth/routes.ts"
  - "apps/web/src/modules/auth/model/sessionStore.ts"
  - "apps/web/src/modules/auth/layouts/PrivateShell.vue"
  - "apps/web/src/modules/settings/index.ts"
  - "apps/web/src/shared/api"
  - "apps/web/src/shared/ui"
created_at: "2026-08-23"
updated_at: "2026-08-23"
issue_created_evidence: "gh issue create bcaceres19/barberia#67; B0 verificado: origin/main a57354b, CI verde (job Go: hu001/hu005/hu007/hu008 suites + go test -race; job database: atlas validate/apply + suites RLS/idempotencia; job web: lint/typecheck/test:unit/build), ver https://github.com/bcaceres19/barberia/actions"
supersedes: "docs/10-backlog/prompts-implementacion.md#prompt-de-hu-020--configuración-básica-de-la-barbería"
superseded_by: null
---

# Implementar HU-020: configuración básica de la barbería

## Instrucción para Claude o Codex

Implementa únicamente `HU-020` como una entrega vertical contract-first: lectura y actualización autenticadas del nombre, zona horaria IANA, correo y teléfono de contacto de la barbería activa; persistencia PostgreSQL; módulo Go `shops`; módulo Vue `settings`; y actualización inmediata del nombre que muestra el cascarón privado de `HU-012`.

Issue real creado: [#67](https://github.com/bcaceres19/barberia/issues/67), con `CA-020-01` a `CA-020-08` enlazados a `docs/02-requisitos/historias-usuario.md`. Criterio de salida de B0 verificado el 2026-08-23 contra `docs/10-backlog/plan-bloques.md`: `origin/main` en `a57354b` con CI en verde (suites `hu001`/`hu005`/`hu007`/`hu008`, `go test -race`, `atlas migrate validate/apply`, idempotencia con dos conexiones reales, `lint`/`typecheck`/`test:unit`/`build` del frontend). Estado pasa a `ready` y luego `in_progress` en rama `feat/67-hu020-configuracion-barberia`.

## Objetivo

Que un barbero autenticado abra la sección “Barbería”, consulte y actualice únicamente la configuración básica de su propia barbería y vea el nombre nuevo en la cabecera sin recargar la aplicación. La zona configurada, no la del navegador ni la del servidor, debe gobernar toda presentación de horas. Un error de validación o de red no produce escritura parcial, no pierde datos no sensibles del formulario y nunca expone información de otro tenant.

## Preflight obligatorio

1. Comprueba árbol limpio y no sobrescribas cambios ajenos. Actualiza `main` mediante fast-forward y verifica que las entregas de B0 siguen integradas y en verde.
2. Demuestra el criterio de salida de B0 punto por punto contra `docs/10-backlog/plan-bloques.md`; que todas las HU estén mergeadas es evidencia necesaria, pero no sustituye esa verificación. Si un punto no puede demostrarse, conserva este prompt en `draft` o pásalo a `blocked` con la causa real.
3. Ejecuta consultas Graphify sobre `HU-020`, `shops`, `settings`, `InTenantTx`, `SessionContext`, `PrivateShell`, el router privado y las políticas de `barbershop`. Usa `graphify path` o `graphify explain` cuando la relación no sea evidente.
4. Lee completamente cada `source_docs`. Los directorios listados significan revisar su API pública, archivos relevantes y pruebas; no editar código generado ni cargar cachés o evidencia binaria sin necesidad.
5. Confirma el estado real: la migración fundacional ya contiene `barbershop.name`, `barbershop.timezone`, `updated_at` y RLS; todavía no contiene `contact_email` ni `contact_phone`; `modelo-fisico-referencia.sql` sí los diseña, pero no es una migración; `shops`, `settings` y `api/openapi/paths/settings.yaml` siguen como marcadores vacíos.
6. Revisa `dudas-pendientes.md` y `contradicciones.md`. Si las fuentes vigentes discrepan sobre campos, límites, semántica del `PATCH`, ruta o actualización de la cabecera, registra la duda o contradicción antes de codificar; no uses el modelo de referencia ni este prompt para reemplazar una decisión normativa.
7. Crea o localiza el issue con el título sugerido y criterios verificables. Sustituye `pending`/`null` por el número y URL reales y actualiza el índice. Solo entonces cambia a `ready`; al empezar crea `feat/<issue>-hu020-configuracion-barberia` desde `main` actualizada y cambia a `in_progress`.

## Alcance incluido

- Operaciones privadas para consultar y actualizar la configuración básica de la barbería derivada de `auth.Principal`; ningún identificador de tenant llega como autoridad desde URL, body, query o cabecera.
- Nombre obligatorio, recortado y con el límite ya aplicado por `barbershop_name_ck`; zona IANA obligatoria; correo de contacto opcional normalizado a minúsculas; teléfono opcional normalizado a E.164; valores opcionales vacíos persistidos como `NULL`.
- Validación de la zona contra `pg_timezone_names` dentro del flujo de persistencia. No se aceptan `COT`, `UTC-5`, offsets sueltos ni la zona detectada del dispositivo como fuente de verdad.
- Migración Atlas roll-forward que agrega solo `contact_email` y `contact_phone` a la tabla aplicada `barbershop`, con restricciones, comentarios y privilegios consistentes con las migraciones vigentes.
- Contrato OpenAPI 3.1.2, handler, caso de uso, repositorio PostgreSQL, cliente TypeScript generado y pantalla Vue responsive/accesible.
- Ruta diferida de “Barbería” dentro del único cascarón privado de `/panel`, y entrada de navegación compuesta mediante APIs públicas de módulos.
- Actualización inmediata y controlada del nombre de la barbería en el estado de sesión/cabecera después de que el servidor confirme el guardado.
- Pruebas unitarias, HTTP/contrato, integración con PostgreSQL real y dos tenants, componente, E2E y evidencia responsive/accesible.

## Fuera de alcance

- Aprovisionar, crear, seleccionar, cambiar o eliminar barberías.
- `public_slug`, logotipo, enlace público, reserva pública o cualquier parte de `F-PUB-01`.
- Barberos, servicios, horarios, bloqueos, agenda, citas, disponibilidad, políticas de reserva/cancelación, recordatorios, canales o proveedores.
- Personalización de paleta, tipografía, radios, tamaños o tema por barbería; `DEC-039` permite un solo tema claro y el nombre/logotipo no sustituyen los tokens.
- Reescribir instantes existentes al cambiar la zona. Un instante persistido sigue siendo el mismo; solo cambia su presentación futura.
- Catálogo nuevo de zonas servido por un endpoint, selector basado en la zona del navegador o dependencia de fechas/zonas sin justificación.
- Estado global paralelo, Pinia por conveniencia, ruta privada paralela, otro guard de sesión o llamadas `fetch` desde componentes.

## Estado existente que debes conservar

- `HU-002` hace obligatoria `database.DB.InTenantTx`; el ejecutor de consultas no sale de la transacción y el contexto de tenant se fija con `set_config(..., true)`.
- `HU-003` entrega router Chi, errores RFC 9457, `request_id`, límite de body y registro técnico sin datos personales. No construyas problemas HTTP ni logs paralelos.
- `HU-006` monta `SessionMiddleware` una sola vez sobre el subrouter privado y propaga `auth.Principal`. Toda ruta nueva se registra sobre ese subrouter después del middleware.
- `HU-012` entrega `GET /private/auth/session`, `sessionStore`, el único guard privado, `PrivateShell`, `AppHeader` y `AppNav`. La cabecera hoy obtiene el nombre real de `barbershop.name`; evoluciona su API pública mínima para aceptar el nombre confirmado por HU-020, sin duplicar el estado ni importar internos entre módulos.
- `HU-009` entrega tokens y `BaseButton`, `BaseInput`, `BaseAlert`, `BaseBadge` y `BaseDialog`; compón con ellos y no inventes otra primitiva visual sin justificarla.
- `20260807170000_create_tenant_foundation.sql` ya está aplicada y es inmutable. Conserva su `barbershop_name_ck`, `barbershop_timezone_ck`, trigger de `updated_at`, políticas RLS y grants; cualquier corrección se hace con una migración nueva.
- `database/modelo-fisico-referencia.sql` sección B.1 incluye `contact_email`, `contact_phone` y también `public_slug`. Solo los dos primeros pertenecen a esta HU: no copies la sección completa.
- El cliente web usa `openapi-fetch` y tipos generados desde `api/openapi/dist/openapi.yaml`; el archivo generado nunca se edita a mano.

## Trabajo requerido

### 1. Contrato HTTP antes del código

1. Define en `api/openapi/paths/settings.yaml` una operación privada `GET /private/settings/barbershop` y una `PATCH /private/settings/barbershop`, o registra antes de implementar por qué otra forma encaja mejor con el estándar. Registra los `$ref` en `openapi.yaml` y declara el tag estable correspondiente.
2. Usa `SessionCookie`, JSON `camelCase`, nombres técnicos en inglés, `operationId` estables, ejemplos completamente ficticios, `x-business-rules` con `RN-DIS-07`, `RN-TEN-01`, `RN-DAT-02` y `x-decisions` pertinentes.
3. El request de `PATCH` contiene solo `name`, `timezone`, `contactEmail` y `contactPhone`; rechaza campos desconocidos y nunca acepta `barbershopId`. Documenta de forma idéntica en schema y backend cómo se expresan los contactos vacíos y cómo se normalizan a ausencia.
4. Declara longitudes y formatos que el backend aplica realmente. Conserva los límites existentes de nombre (120) y zona (64); usa formato `email` y E.164 para contactos sin crear reglas más restrictivas que las fuentes. Si hace falta fijar un límite no aprobado, regístralo antes de inventarlo.
5. Documenta `200`, `400` para JSON/forma inválida, `401`, `404` cuando la fila no sea visible en el tenant, `422` para validación procesable y `500`, todos con los componentes RFC 9457 vigentes. La respuesta exitosa devuelve la representación canónica guardada.
6. Ejecuta lint y bundle antes de implementar; regenera el cliente TypeScript únicamente desde el bundle validado.

### 2. Migración y persistencia

1. Crea una migración nueva con Atlas. No edites ninguna de las ocho migraciones aplicadas ni regeneres el hash para ocultar una modificación histórica.
2. Agrega exclusivamente `barbershop.contact_email` y `barbershop.contact_phone`, ambas anulables, con restricciones coherentes con la normalización aprobada, comentarios de clasificación/retención y sin `public_slug`.
3. Evalúa bloqueo y actualización con datos existentes: las columnas nuevas deben admitir la actualización desde la versión actual sin inventar valores de contacto. Valida base vacía, actualización desde la última versión con las dos barberías de prueba y `atlas.sum`.
4. No intentes validar IANA mediante un `CHECK` sobre `pg_timezone_names`: no es una expresión inmutable. La aplicación comprueba la zona contra ese catálogo dentro de la operación; la base conserva el límite estructural vigente.
5. Conserva RLS `ENABLE`/`FORCE`, políticas y privilegios actuales de `barbershop`. Prueba con `barberia_app` real que A lee/actualiza A, no ve ni modifica B, y que ausencia de contexto falla cerrado. Usa `barberia_migrator`/`barberia_owner` según el flujo vigente de `DEC-040`; nunca superusuario para declarar la prueba de aislamiento cumplida.
6. Añade testdata/prueba SQL específica de HU-020 sin datos personales reales y actualiza los README de datos que enumeran migraciones, fixtures y pruebas.

### 3. Backend Go

1. Implementa en `internal/modules/shops` dominio, errores, puertos, servicio y adaptadores `postgres`/`httpapi`. Chi, `net/http`, JSON, pgx y `database.DB` no entran al dominio ni al servicio.
2. El handler obtiene `auth.Principal` del contexto y pasa únicamente sus identificadores opacos. Si falta el principal, es un defecto interno de wiring; nunca toma el tenant de datos controlados por el cliente.
3. Normaliza nombre y contactos antes de persistir: recorta nombre; convierte correo a minúsculas; convierte contactos vacíos a `NULL`; valida correo y teléfono sin registrar sus valores. El nombre vacío o inválido y los contactos mal formados producen errores de campo `422`.
4. Dentro de una sola `InTenantTx`, comprueba la zona con una consulta parametrizada exacta a `pg_timezone_names` y ejecuta el `UPDATE` filtrando también por `barbershop_id`. Zona inválida implica cero escritura: prueba que nombre y contactos anteriores permanecen intactos.
5. La lectura y actualización seleccionan solo las cuatro propiedades autorizadas. No devuelvas ni registres credenciales, usuario de staff, detalles SQL o datos de contacto en logs; conserva `request_id` en errores inesperados.
6. Registra ambos handlers sobre el subrouter privado real, después de `SessionMiddleware`; amplía la prueba estructural de inventario para demostrar que no existe una ruta de settings fuera del middleware.
7. Escribe pruebas unitarias del servicio y validaciones, pruebas HTTP contra el bundle y pruebas PostgreSQL de dos tenants. No uses mocks de base para afirmar RLS ni zona IANA.

### 4. Frontend Vue

1. Convierte `modules/settings` en un módulo real con API por intención, modelos de vista, validación, formulario, página y pruebas. Los componentes usan el cliente tipado; no duplican DTO ni inspeccionan `Problem.detail` para decidir comportamiento.
2. Expón desde `settings/index.ts` la ruta diferida de la sección. Compónla como hija del único `/panel` mediante la API pública del router/cascarón existente; no hagas que `auth` importe internos de `settings` ni crees otro guard.
3. Agrega “Barbería” a la navegación mediante una superficie pública pequeña. Mantén la dirección `app → modules → shared` y usa `index.ts` u orquestación en `app` cuando dos módulos deban coordinarse.
4. Modela estados discriminados: carga inicial, listo/sin cambios, guardando, error por campo, error recuperable y éxito. Evita doble envío, conserva nombre/zona/contactos ante fallo recuperable y cancela o ignora respuestas obsoletas si el usuario repite la carga.
5. Después de un `PATCH 200`, actualiza el nombre del contexto de sesión mediante una función estrecha exportada por la API pública de `auth` o coordinación equivalente en `app`. Solo el valor confirmado por el servidor cambia la cabecera; un fallo no aplica optimismo ni obliga a recargar. La siguiente rehidratación sigue usando `GET /private/auth/session` como autoridad.
6. No uses la zona del navegador para reemplazar la guardada. Si creas una función compartida de presentación para probar `CA-020-04`, debe recibir siempre la zona de la barbería de forma explícita y usar APIs de plataforma; no agregues una biblioteca de fechas por esta HU.
7. Etiquetas, ayuda, errores asociados, resumen/foco, anuncio de guardado, teclado, objetivo táctil y reflow siguen `estandar-diseno-visual.md`; todo color, tamaño, radio, espacio y movimiento viene de tokens existentes.

## Pruebas y evidencia

- Dominio/servicio: nombre vacío/espacios/límite, correo y teléfono ausentes/válidos/inválidos, normalización, zona IANA válida/inválida y actualización atómica sin escritura parcial.
- PostgreSQL real: lectura/actualización A, aislamiento B, intento con identificador ajeno en cada entrada manipulable, ausencia de contexto, rol sin `BYPASSRLS`, constraints, migración vacía y upgrade con datos.
- HTTP/contrato: `GET`/`PATCH` exitosos, JSON o campo desconocido, `401`, `404`, `422` por cada campo y `500` sin detalles internos; handler probado contra el bundle real.
- Privacidad: inspecciona logs de un flujo exitoso y fallido y demuestra que no contienen nombre, correo ni teléfono de contacto, cookie, token o cuerpo crudo.
- Componente: carga, valores iniciales, edición, contactos vacíos, error de campo, error recuperable con datos conservados, doble envío bloqueado, éxito y cabecera actualizada solo tras confirmación.
- `CA-020-04`: usa un instante conocido y dos contextos de navegador con zonas de dispositivo distintas; ambos deben producir la misma hora al formatear con la zona configurada de la barbería. Incluye un control negativo que demuestre que omitir `timeZone` sí habría dependido del dispositivo.
- E2E real: iniciar sesión → abrir “Barbería” → editar → guardar → comprobar cabecera → recargar → comprobar persistencia. Ejecuta contra API y PostgreSQL reales, sin interceptar la operación que se declara probada.
- Evidencia visual a 320, 360, 768 y 1280 px, zoom 200 %, teclado completo, foco visible/no oculto y `vitest-axe`/axe-core sin violaciones aplicables. No actualices capturas sin inspeccionarlas.

Entrega una tabla `Criterio | Estado | Prueba o evidencia` para `CA-020-01` a `CA-020-08`. No declares cumplido un criterio con una prueba que no ejerza su capa real.

## Documentación y trazabilidad

- Actualiza `api/openapi/CHANGELOG.md`, README de API/web/base de datos, lista de migraciones/pruebas, matriz de trazabilidad e historial documental donde realmente resulte afectado.
- Actualiza `database/modelo-fisico-referencia.sql` solo si la migración aprobada descubre una diferencia real; no lo uses para simular que la migración existe.
- Actualiza este prompt y `docs/10-backlog/prompts/README.md` con issue, rama, PR, estado y evidencia reales en cada transición.
- No declares B1 abierto o HU-021 desbloqueada hasta que HU-020 esté integrada y sus dependencias reales estén satisfechas.

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
# aplicar en base vacía y en actualización desde la versión anterior
# ejecutar database/tests/hu020_configuracion_barberia.sql con barberia_app real

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

Documenta la preparación real de PostgreSQL y las variables de prueba sin copiar DSN, contraseñas ni secretos. `gofmt -l .` debe quedar vacío. Una suite omitida, inestable o sustituida por mocks se reporta como no verificada.

## Git y PR

- Rama esperada después de asignar issue: `feat/<issue>-hu020-configuracion-barberia`.
- Commits y título del PR: `feat(settings): implementa HU-020 configuración básica de la barbería` o equivalente Conventional Commits coherente con todo el corte vertical.
- Usa `Closes #<issue>` solo si los ocho criterios y la documentación están cubiertos; de lo contrario usa `Refs #<issue>` y conserva el estado real.
- Abre PR contra `main`, auto-revisa el diff completo, adjunta tabla de criterios, evidencia responsive/accesible, riesgo de migración y recuperación roll-forward.
- No hagas push directo, force push, merge manual, edición de migraciones aplicadas ni reescritura de `main`.
