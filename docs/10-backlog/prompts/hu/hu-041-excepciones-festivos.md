---
prompt_id: "PROMPT-HU-041-v1"
version: "1.2"
kind: "hu"
status: "in_progress"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-041"
related_hu:
  - "HU-040"
  - "HU-042"
  - "HU-021"
issue: 95
issue_url: "https://github.com/bcaceres19/barberia/issues/95"
suggested_issue_title: "feat(schedule): implementar HU-041 excepciones de jornada y festivos"
branch: "feat/95-hu041-excepciones-festivos"
pr: null
pr_url: null
depends_on:
  - "HU-040 integrada en main (PR #93)"
  - "Criterio de salida de B1 cumplido; HU-020–HU-024 integradas en main"
  - "CT-008 resuelta como DEC-070 (FK de B2 en ON DELETE RESTRICT)"
  - "Issue real #95 creado y enlazado"
rules:
  - "RN-TEN-01"
  - "RN-BLQ-02"
  - "RN-DIS-05"
  - "RN-DIS-07"
decisions:
  - "DEC-007"
  - "DEC-019"
  - "DEC-020"
  - "DEC-024"
  - "DEC-033"
  - "DEC-034"
  - "DEC-035"
  - "DEC-036"
  - "DEC-037"
  - "DEC-038"
  - "DEC-039"
  - "DEC-040"
acceptance_criteria:
  - "CA-041-01"
  - "CA-041-02"
  - "CA-041-03"
  - "CA-041-04"
  - "CA-041-05"
  - "CA-041-06"
  - "CA-041-07"
  - "CA-041-08"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-de-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/00-control/glosario.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/prioridades.md"
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
  - "database/migrations"
  - "api/openapi/README.md"
  - "api/openapi/openapi.yaml"
  - "api/openapi/paths/schedules.yaml"
  - "apps/api/README.md"
  - "apps/api/internal/modules/schedule"
  - "apps/web/README.md"
  - "apps/web/src/modules/schedules"
created_at: "2026-08-25"
updated_at: "2026-08-25"
supersedes: null
superseded_by: null
---

# Implementar HU-041: excepciones de jornada y festivos

## Instrucción para el agente

Este prompt pasó a ready y ahora está in_progress: el issue real es #95 (https://github.com/bcaceres19/barberia/issues/95), CT-008 quedó resuelta como DEC-070 (docs/00-control/registro-decisiones.md) — las FK de working_hour_override/working_hour_override_segment hacia barber usan ON DELETE RESTRICT, ya aplicado en database/modelo-fisico-referencia.sql §C.3/C.3b — y la rama feat/95-hu041-excepciones-festivos implementa contrato, migración, backend Go, frontend Vue y pruebas reales descritos abajo. CA-041-08 (carga/vacío/error/éxito, teclado y responsive) queda Parcial: E2E escrito en apps/web/e2e/excepciones-festivos.spec.ts pero no ejecutado contra Chromium real, mismo estado documentado que HU-040. Actualiza pr/pr_url reales en este archivo y en el índice del catálogo antes de fusionar.

Implementa únicamente la configuración por fecha y el calendario festivo de esta HU. No adelantes bloqueos de descanso/emergencia, agenda, citas ni cálculo público de disponibilidad.

## Objetivo

Que cada barbero pueda activar o desactivar el calendario colombiano de festivos y configurar qué ocurre en una fecha concreta: día cerrado, horario especial o apertura manual de un festivo, sin alterar la configuración de otro barbero.

## Preflight obligatorio

1. Verifica árbol limpio respecto de cambios ajenos, main actualizada y HU-040 integrada.
2. Confirma issue real, rama feat/<issue>-hu041-excepciones-festivos y estado del prompt antes de ejecutar.
3. Lee completamente los source_docs y valida que no haya una duda o contradicción nueva.
4. Comprueba la resolución vigente de CT-008; detente si no existe un DEC-* compatible con AGENTS.md.
5. Ejecuta Graphify sobre schedule, working_hour, barber, sesión/tenant y el contrato OpenAPI.
6. Revisa que el calendario sea por barbero, no por barbería global, y que los ejemplos no contengan datos reales.

## Alcance incluido

- Configuración privada holiday_calendar_enabled independiente por barbero.
- Resolución de festivos colombianos para una fecha y zona de barbería, con bloqueo por defecto cuando la opción está activa.
- Excepción por fecha para cerrar completamente un día o abrirlo con uno o varios tramos especiales.
- Apertura manual de un festivo completo o con horario reducido; la excepción manual prevalece sobre el bloqueo automático.
- Consulta, creación, edición y retiro de excepciones con validación tenant-aware.
- Modelo normalizado para cabecera de excepción y tramos; sin jsonb ni listas codificadas en una columna.
- Contrato OpenAPI, caso de uso Go, persistencia Atlas, cliente tipado, pantalla Vue y pruebas reales.

## Fuera de alcance

- Bloqueos de break, lunch, unavailable, vacation o emergency: HU-042.
- Citas afectadas, reprogramación, cancelación y notificaciones.
- Cálculo completo de disponibilidad pública y agenda diaria.
- Cambio de zona horaria de la barbería o conversión masiva de instantes.
- Crear una tabla de festivos global editable por la barbería; el calendario colombiano es una capacidad normativa, no un catálogo libre.
- Borrado físico en cascada o cualquier comportamiento no resuelto por CT-008.

## Estado existente que debe conservarse

- working_hour y barber.holiday_calendar_enabled son insumos del modelo de referencia, no migraciones aplicadas.
- HU-020 es la autoridad de la zona IANA de la barbería y HU-021 la autoridad del recurso barber.
- Un barbero puede abrir un festivo con horario normal o reducido; esa decisión no se comparte con otros barberos.
- La disponibilidad futura debe poder consultar una resolución de jornada efectiva; no materialices datos sintéticos en tablas de citas inexistentes.
- Reutiliza el módulo schedule, el cliente generado y los componentes visuales existentes.

## Trabajo requerido

### 1. Semántica y contrato

1. Define contract-first las operaciones para leer/cambiar el toggle por barbero y gestionar excepciones por fecha.
2. Declara schemas cerrados, zona mostrada, errores RFC 9457, seguridad de sesión y trazabilidad RN-BLQ-02/DEC-020.
3. La respuesta de una excepción debe distinguir día cerrado de día abierto con tramos; no envíes un objeto ambiguo con campos incompatibles.
4. No aceptes barbershopId como autoridad. Un barbero ajeno o una excepción ajena deben responder con el mismo 404 seguro.
5. Expón un puerto interno de resolución de jornada efectiva para que disponibilidad lo consuma después; no inventes endpoints de reserva ni consultas de appointment.

### 2. Datos y backend Go

1. Crea la migración solo después de resolver CT-008; usa la forma normalizada de cabecera y segmentos, RLS forzada y grants mínimos.
2. Garantiza una sola excepción por barbero y fecha y que un día cerrado no tenga segmentos.
3. Valida segmentos no solapados dentro de la excepción, intervalos [inicio, fin), duración positiva y cruce de medianoche permitido según DEC-020.
4. Resuelve el calendario colombiano de forma determinista y aislada en el dominio; si una dependencia nueva fuera necesaria, justifícala antes de incorporarla.
5. La precedencia debe ser explícita: excepción manual abierta/cerrada > festivo automático > horario semanal.
6. No conviertas instantes ya almacenados ni uses la zona del navegador como fuente de verdad.
7. No uses ON DELETE CASCADE salvo decisión explícita de CT-008; conserva evidencias y RLS por tenant.
8. Los logs no incluyen nombre, teléfono, correo, razones libres ni tokens; usa IDs y tipos de operación.

### 3. Frontend Vue

1. Integra la configuración por barbero en una pantalla de horarios y excepciones, sin mezclarla con el editor de bloqueos de HU-042.
2. Permite elegir fecha, cerrado/abierto y tramos especiales; muestra la zona IANA de la barbería.
3. Comunica que activar el calendario bloquea festivos por defecto y muestra la acción para abrir uno.
4. Gestiona estados de carga, vacío, error recuperable, conflicto y éxito; conserva valores no sensibles ante fallos.
5. Usa texto/semántica además de color, foco contenido, teclado y objetivos táctiles de 44 px.
6. Prueba reflow sin pérdida a 320, 360, 768 y 1280 px y zoom 200 %.

## Pruebas y evidencia

- Dominio: fechas festivas colombianas, toggle independiente por barbero, precedencia, día cerrado, horario especial, medianoche, contiguidad y solape.
- PostgreSQL real: dos tenants, varios barberos, unicidad por fecha, segmento bajo cabecera cerrada, RLS, grants y decisión de FK de CT-008.
- HTTP/contrato: lectura/actualización del toggle, CRUD de excepción, 401, 404, 409, 422 y RFC 9457.
- Componente/E2E: activar, cerrar festivo, abrir festivo reducido, editar excepción, retirar excepción y comprobar aislamiento.
- Responsive/accesible: 320, 360, 768 y 1280 px, teclado, foco, zoom 200 %, axe-core y contraste.
- No declares implementado el cálculo completo de disponibilidad ni el flujo de citas afectadas.

## Documentación y trazabilidad

- Actualiza OpenAPI/CHANGELOG, README, migraciones, atlas.sum, pruebas SQL, matriz, historial, modelo de referencia y plan si cambian.
- Registra la fuente reproducible del calendario colombiano y cualquier dependencia aprobada; no ocultes una decisión de proveedor en código.
- Propaga cualquier resolución de CT-008 antes de actualizar este prompt a ready.
- Actualiza el catálogo con issue, rama, PR y estado reales.

## Verificación final

~~~text
pnpm run openapi:lint
pnpm run openapi:bundle
pnpm run db:validate
pnpm run db:test
cd apps/api && gofmt -w . && go vet ./... && go test -race ./... && go build ./...
cd apps/web && pnpm run lint && pnpm run typecheck && pnpm run test:coverage && pnpm run build && pnpm run test:e2e
git diff --check
graphify update .
~~~

Entrega Criterio | Estado | Prueba o evidencia para CA-041-01–CA-041-08, incluyendo la precedencia aplicada en cada caso.

## Git y PR

- Rama: feat/<issue>-hu041-excepciones-festivos.
- Commit/título: feat(schedule): implementa HU-041 excepciones de jornada y festivos.
- Usa Closes #<issue> solo con los ocho criterios verificados; en otro caso usa Refs #<issue>.
- No hagas push directo, force push, merge manual de main, DDL al arrancar ni cambios a migraciones aplicadas.

