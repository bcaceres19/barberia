---
prompt_id: "PROMPT-HU-040-v1"
version: "1.1"
kind: "hu"
status: "ready"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-040"
related_hu:
  - "HU-021"
  - "HU-041"
  - "HU-042"
issue: 90
issue_url: "https://github.com/bcaceres19/barberia/issues/90"
suggested_issue_title: "feat(schedule): implementar HU-040 horario laboral recurrente"
branch: null
pr: null
pr_url: null
depends_on:
  - "Criterio de salida de B1 cumplido; HU-020–HU-024 integradas en main (PR #84)"
  - "CT-008 resuelta como DEC-070 (FK de B2 en ON DELETE RESTRICT, no CASCADE)"
  - "Issue real #90 creado y enlazado"
rules:
  - "RN-TEN-01"
  - "RN-DIS-05"
  - "RN-DIS-07"
  - "RN-IDE-01"
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
  - "DEC-043"
acceptance_criteria:
  - "CA-040-01"
  - "CA-040-02"
  - "CA-040-03"
  - "CA-040-04"
  - "CA-040-05"
  - "CA-040-06"
  - "CA-040-07"
  - "CA-040-08"
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

# Implementar HU-040: horario laboral recurrente

## Instrucción para el agente

Este archivo es un prompt persistente de implementación. Sus dos guardas de bloqueo ya se resolvieron: el issue real es #90 (https://github.com/bcaceres19/barberia/issues/90) y CT-008 quedó resuelta como DEC-070 (docs/00-control/registro-decisiones.md) — las FK de working_hour, working_hour_override(_segment), time_block_series(_date/_exception) y time_block hacia barber usan ON DELETE RESTRICT, no CASCADE; database/modelo-fisico-referencia.sql §C ya se actualizó en consecuencia. El estado pasa a ready. Actualiza rama y PR reales en este archivo y en el índice del catálogo en la misma rama del issue.

Implementa únicamente la preocupación de esta HU. No adelantes excepciones por fecha, festivos, bloqueos, citas, disponibilidad pública ni agenda.

## Objetivo

Que un barbero autenticado pueda consultar y mantener los tramos recurrentes de trabajo de cada barbero de su barbería, con varios tramos por día, horarios nocturnos y validación de intervalos, para que las capacidades futuras calculen disponibilidad sobre una jornada real.

## Preflight obligatorio

1. Verifica árbol limpio respecto de cambios ajenos, main actualizada por fast-forward y el estado real de B1.
2. El issue real es #90; crea la rama feat/90-hu040-horario-laboral.
3. Lee por completo todos los source_docs, incluyendo el modelo físico ya actualizado por DEC-070 (FK en RESTRICT) únicamente como insumo de diseño.
4. CT-008 ya tiene un DEC-* vigente (DEC-070: RESTRICT) que resuelve las FK de B2; no reabras la decisión.
5. Ejecuta Graphify sobre schedule, barber, sesión/tenant, rutas privadas y el cliente OpenAPI; confirma que el módulo schedule es el dueño de la capacidad.
6. No sobrescribas cambios locales de otras personas ni edites migraciones aplicadas.

## Alcance incluido

- Recurso privado tenant-aware para consultar, crear, editar y retirar tramos de working_hour de un barbero de la barbería activa.
- Días ISO 1–7, varios tramos por día, duración entera positiva y representación de tiempo civil local.
- Intervalos semiabiertos [inicio, fin); un tramo puede cruzar medianoche conforme a DEC-020.
- Rechazo de tramos que se solapan para el mismo barbero y día de semana, sin escritura parcial.
- Contrato OpenAPI contract-first, caso de uso Go en schedule, persistencia PostgreSQL con RLS y cliente TypeScript generado.
- Pantalla privada de horarios que reutilice el cascarón, componentes y patrones visuales existentes.
- Pruebas unitarias, HTTP/contrato, PostgreSQL real con dos tenants, componente, E2E, responsive y accesibilidad.

## Fuera de alcance

- Excepciones por fecha, calendario colombiano, horarios especiales y festivos: HU-041.
- Descansos, vacaciones, emergencias, recurrencias de bloqueos y listas explícitas: HU-042.
- Cálculo de disponibilidad, citas, reserva pública, agenda y notificaciones.
- Alta/eliminación de barberías, credenciales, roles o eliminación física de barberos.
- Personalización de paleta, tipografía o tema por barbería.
- Copiar literalmente database/modelo-fisico-referencia.sql si contradice AGENTS.md o el DEC-* que resuelva CT-008.

## Estado existente que debe conservarse

- HU-020 aporta la zona horaria de barbershop; HU-021 aporta el recurso barber para uno o varios profesionales.
- apps/api/internal/modules/schedule/doc.go, apps/web/src/modules/schedules/index.ts y api/openapi/paths/schedules.yaml son stubs intencionales; extiéndelos sin crear un módulo paralelo.
- El tenant se deriva de la sesión y del contexto transaccional; el cliente nunca envía un barbershopId confiable.
- El modelo de referencia propone working_hour como tiempo civil local y una duración; es revisable y no sustituye la migración Atlas.
- Ninguna regla de horarios debe vivir solo en Chi, Vue o SQL; el dominio/aplicación conserva la validación.

## Trabajo requerido

### 1. Contrato HTTP

1. Define las operaciones privadas, sus paths, seguridad, schemas cerrados, ejemplos ficticios, errores RFC 9457, x-business-rules y x-decisions.
2. No aceptes tenant, zona horaria ni un barbero de otro tenant como autoridad del request.
3. Declara con claridad los límites de día ISO, hora, duración y conflicto de solape.
4. El contrato, el handler y el cliente generado deben coincidir; no edites manualmente el archivo generado.

### 2. Datos y backend Go

1. Diseña la migración con Atlas después de resolver CT-008; aplica tercera forma normal, FK tenant-aware, RLS forzada, grants mínimos e índices justificados.
2. Rechaza duration_minutes <= 0, valores mayores al límite aprobado, día fuera de 1–7 y starts_time inválido.
3. Valida el solape dentro del caso de uso con semántica [inicio, fin); una jornada partida es válida y dos tramos contiguos no se consideran solapados.
4. Trata el cruce de medianoche de forma determinista y prueba una jornada nocturna antes de integrarla con disponibilidad.
5. Usa consultas parametrizadas dentro de InTenantTx; no dependas solo de RLS y no introduzcas jsonb, ORM o una carpeta global.
6. Usa ON DELETE RESTRICT en la FK de working_hour hacia barber, conforme a DEC-070; retirar un tramo es una operación física propia (no hay regla de retención para working_hour), y no depende de borrar un barber, que no tiene borrado físico en su alcance vigente.
7. Los logs usan IDs opacos y request_id; nunca nombres, teléfonos, correos, cookies o tokens.

### 3. Frontend Vue

1. Presenta selector de barbero, siete días y tramos ordenados; permite añadir, editar y retirar un tramo sin duplicar el cliente HTTP.
2. Muestra la zona de la barbería junto a la hora y conserva datos no sensibles ante error recuperable.
3. Comunica solapes y validaciones junto al campo o fila correspondiente; no usa color como único indicador.
4. Evita doble envío, anuncia carga/error/éxito y respeta foco, teclado, objetivos táctiles y reducción de movimiento.
5. Verifica reflow sin pérdida funcional a 320, 360, 768 y 1280 px.

## Pruebas y evidencia

- Dominio: días, duración, varios tramos, contiguidad, solape, medianoche y zona.
- PostgreSQL real: base vacía y upgrade, dos tenants, RLS, FK/grants decididos por CT-008, lectura/escritura cruzada y ausencia de escritura parcial.
- HTTP/contrato: lista, alta, edición, retiro, 401, 404, 409, 422 y error uniforme.
- Componente/E2E: selector de barbero, vacío, alta, edición, conflicto, retiro, reintento y aislamiento.
- Responsive/accesible: 320, 360, 768 y 1280 px, teclado, foco, zoom 200 %, axe-core y contraste.
- No declares probado el cálculo de disponibilidad: pertenece a historias posteriores.

## Documentación y trazabilidad

- Actualiza contrato OpenAPI/CHANGELOG, README del API y frontend, migración, atlas.sum, pruebas SQL, modelo físico si cambia, matriz, historial y plan de bloques.
- Actualiza este prompt y el catálogo con issue, rama, PR y estado reales.
- Si la resolución de CT-008 cambia una fuente normativa, propaga el DEC-* antes de marcar la HU lista.
- Mantén explícito que la HU no implementa disponibilidad ni citas.

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

Entrega una tabla Criterio | Estado | Prueba o evidencia para CA-040-01–CA-040-08. No declares cumplido un criterio que dependa de una decisión o capacidad no implementada.

## Git y PR

- Rama: feat/<issue>-hu040-horario-laboral.
- Commit/título: feat(schedule): implementa HU-040 horario laboral recurrente.
- Usa Closes #<issue> solo si cubres los ocho criterios; en otro caso usa Refs #<issue>.
- No hagas push directo, force push, merge manual de main, DDL al arrancar ni cambios a migraciones aplicadas.

