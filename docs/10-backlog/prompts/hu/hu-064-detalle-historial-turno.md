---
prompt_id: "PROMPT-HU-064-v1"
version: "1.3"
kind: "hu"
status: "executed"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-064"
related_hu:
  - "HU-060"
  - "HU-063"
  - "HU-065"
issue: 120
issue_url: "https://github.com/bcaceres19/barberia/issues/120"
suggested_issue_title: "feat(booking): implementar HU-064 detalle e historial de turno"
branch: "feat/120-hu064-detalle-historial-turno"
pr: 121
pr_url: "https://github.com/bcaceres19/barberia/pull/121"
depends_on:
  - "HU-063 integrada en main mediante PR #118 (cumplido)"
  - "Issue real propio con CA-064-01 a CA-064-08 (cumplido: #120)"
rules:
  - "RN-CIT-01"
  - "RN-HIS-01"
  - "RN-HIS-02"
  - "RN-RES-02"
  - "RN-RES-03"
  - "RN-DAT-01"
  - "RN-DAT-02"
  - "RN-DIS-07"
  - "RN-TEN-01"
decisions:
  - "DEC-004"
  - "DEC-007"
  - "DEC-014"
  - "DEC-016"
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
  - "DEC-041"
  - "DEC-045"
  - "DEC-046"
  - "DEC-074"
  - "DEC-075"
acceptance_criteria:
  - "CA-064-01"
  - "CA-064-02"
  - "CA-064-03"
  - "CA-064-04"
  - "CA-064-05"
  - "CA-064-06"
  - "CA-064-07"
  - "CA-064-08"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/00-control/glosario.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/estados-citas.md"
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
  - "docs/06-api/estandar-openapi.md"
  - "docs/10-backlog/plan-bloques.md"
  - "database/migrations/20260827110000_create_appointment_core.sql"
  - "api/openapi/paths/private-appointments.yaml"
  - "apps/api/internal/modules/booking"
  - "apps/web/src/modules/agenda"
created_at: "2026-08-31"
updated_at: "2026-09-01T04:50:00Z"
supersedes: null
superseded_by: null
---

# Implementar HU-064: detalle e historial de un turno

## Instrucción para el agente

Implementa únicamente la lectura privada del detalle y del historial de un turno. Esta HU no cambia ninguna cita: no añade edición, reprogramación, cancelación, cierre ni corrección de estado. La lista diaria debe seguir siendo mínima; contacto, nota e historial solo aparecen dentro del detalle autenticado.

El prompt está en `draft` porque `issue: pending` y porque depende de `HU-063` integrada. No modifiques el repositorio hasta que ambas dependencias consten como cumplidas, el issue real esté enlazado y exista una rama propia.

## Objetivo

Que el barbero abra un turno desde la agenda, comprenda su estado y datos vigentes, revise un historial completo e inmutable y vuelva a la misma fecha/barbero sin filtrar datos personales ni sustituir snapshots históricos por el catálogo actual.

## Preflight obligatorio

1. Verifica árbol limpio, `main` actualizada y `HU-063` integrada con sus CA/evidencias reales.
2. Localiza el issue propio. Mientras `issue: pending` o la dependencia siga pendiente, detente antes de cambiar código.
3. Con dependencias cumplidas, actualiza este prompt y crea `feat/<issue>-hu064-detalle-historial-turno`.
4. Ejecuta Graphify sobre `booking`, `appointment`, historial, `customer`, `staff`, rutas de agenda, cliente API y utilidades de tiempo.
5. Lee todos los `source_docs`; revisa especialmente privacidad, append-only, consultas con límite y plantilla P0 “Detalle de turno”.
6. Comprueba que el vocabulario de eventos coincide con `DEC-041` y la migración aplicada; no edites esa migración.

## Alcance incluido

- `GET` privado del detalle por `appointmentId` y lectura paginada/acotada del historial, contract-first.
- Respuesta de detalle con persona atendida, cliente que reservó, contacto opcional, nota de cliente, barbero, origen, intervalo, zona, estado y snapshots.
- Token opaco de versión/concurrencia en la representación para que una mutación posterior pueda usar una precondición sin exponer `updated_at` como regla de negocio.
- Historial ordenado con evento, actor seguro, instante, motivo permitido y cambios anterior/nuevo persistidos.
- Consultas tenant-aware con selección explícita, límite/cursor estable, RLS real y plan medido.
- Ruta Vue desde cada fila de agenda y regreso conservando fecha/barbero de `HU-063`.
- Carga, historial vacío, error/reintento, privacidad, pruebas y evidencia responsive/accesible.

## Fuera de alcance

- Cualquier `POST`, `PUT`, `PATCH` o `DELETE` sobre la cita o su historial.
- Acciones futuras deshabilitadas o promesas de notificación.
- Datos de contacto o nota en la lista diaria, URL, logs, errores, métricas o telemetría.
- Unir con el catálogo vigente para reemplazar snapshots.
- Exponer `customerId`, IDs de actores, `barbershopId`, SQL, RLS o timestamps internos de control.
- Crear una migración por intuición; un índice nuevo exige consulta y `EXPLAIN` que demuestren necesidad.

## Estado existente que debe conservarse

- `appointment`, `appointment_history` y `appointment_history_change` existen desde `HU-060`, con RLS forzada y sin privilegios `UPDATE`/`DELETE` sobre historial.
- El vocabulario de ocho eventos está cerrado en inglés `snake_case`; la interfaz traduce al español.
- `appointment` conserva snapshots de servicio y enlaza cliente/barbero de forma tenant-aware.
- El índice `idx_appointment_history_shop_appointment_occurred` ya cubre barbería, cita e instante; mídelo antes de agregar otro.
- `HU-062`/`HU-063` muestran una proyección mínima sin contacto y proporcionan el contexto de fecha/barbero.

## Trabajo requerido

### 1. Contrato OpenAPI

1. Diseña operaciones privadas coherentes para detalle y colección de historial, con `operationId` únicos, seguridad explícita, cursor/límite y errores RFC 9457.
2. Usa request/response direccionales cerrados; la respuesta de detalle incluye un token opaco de concurrencia estable para esa representación.
3. Documenta qué campos son snapshots y que el contacto no pertenece a la lista diaria.
4. Recurso mal formado, inexistente o ajeno usa el mismo `404`; no distingas causas.

### 2. Backend Go y PostgreSQL

1. Añade casos de uso de lectura al núcleo `booking`; dominio/aplicación no importan Chi ni pgx.
2. Consulta la cita por `(barbershop_id, appointment_id)` y el historial con cursor estable; selecciona columnas explícitas.
3. Resuelve nombres visibles de actor sin exponer correo, IDs internos ni generar N+1. Si falta un nombre autorizado, usa una etiqueta segura coherente con el tipo de actor.
4. Produce el token opaco de versión de manera determinista a partir del estado de concurrencia vigente; no conviertas su contenido en API pública.
5. Verifica `EXPLAIN (ANALYZE, BUFFERS)` con historial representativo. No cambies índices si el actual sirve.
6. Conserva las pruebas de privilegios append-only y agrega dos tenants, límites, cursor y cancelación de contexto.

### 3. Frontend Vue

1. Añade ruta lazy de detalle dentro de `agenda`, enlazada desde cada fila operable sin convertir toda la tarjeta en un control ambiguo.
2. Presenta hora/zona, persona, contacto, servicio snapshot, precio, estado, nota e historial según la plantilla P0.
3. Nunca guardes contacto en estado global, query string, local storage o mensajes de error.
4. Traduce eventos/estados con mapeos exhaustivos tipados; no muestres valores técnicos crudos.
5. Volver conserva fecha y barbero. No renderices botones de acciones no implementadas.
6. Implementa carga, vacío de historial, paginación/carga adicional, error y reintento con foco correcto.

## Pruebas y evidencia

- Dominio/modelo: snapshots, zona, estados, actores, token opaco y orden estable.
- HTTP/contrato: `200`, `400`, `401`, `404`, cursor inválido, límite, schemas exactos y ausencia de campos internos.
- PostgreSQL real: dos tenants, historia vacía/llena, empate de instante, páginas sin omisión/duplicado, plan e intentos `UPDATE`/`DELETE` rechazados.
- Privacidad: inspecciona respuesta diaria, errores y logs para comprobar ausencia de teléfono, correo y nota.
- Componente/router: abrir, volver, carga, vacío, historia paginada, error/reintento, teclado, foco y axe-core.
- E2E: desde una fecha no actual abre turno, verifica snapshots/historial, vuelve a la misma fecha/barbero.
- Evidencia: 320, 360, 768 y 1280 px, zoom 200 %, teclado, contraste y foco.

## Documentación y trazabilidad

- Actualiza OpenAPI/CHANGELOG, cliente generado, README API/web, matriz, historial, plan y estado real de este prompt.
- Documenta consulta, cota/cursor, `EXPLAIN` y decisión de reutilizar o crear índice.
- Mantén `HU-065` pendiente y no declares que existe una acción de reprogramación.

## Verificación final

```text
pnpm run openapi:check-config
pnpm run openapi:lint
pnpm run openapi:bundle

cd apps/api
gofmt -l .
go vet ./...
go test -race ./...
go build ./...

cd ../web
pnpm run format
pnpm run lint
pnpm run typecheck
pnpm run test:unit
pnpm run build
pnpm run test:e2e

cd ../..
git diff --check
graphify update .
```

Ejecuta integraciones contra PostgreSQL 14 real con dos tenants. Entrega `Criterio | Estado | Prueba o evidencia` para `CA-064-01`–`CA-064-08`, junto con el plan de consulta y evidencia de privacidad.

## Git y PR

- Rama sugerida: `feat/<issue>-hu064-detalle-historial-turno`.
- Commit/título: `feat(booking): implementa HU-064 detalle e historial de turno`.
- Usa `Closes #<issue>` solo con ocho criterios y evidencia completos; si no, `Refs #<issue>`.
- No hagas push directo, force push, merge de `main`, edites migraciones aplicadas ni mezcles `HU-065`.
