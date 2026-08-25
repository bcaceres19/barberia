---
prompt_id: "PROMPT-HU-042-v1"
version: "1.0"
kind: "hu"
status: "draft"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-042"
related_hu:
  - "HU-040"
  - "HU-041"
  - "HU-060"
issue: "pending"
issue_url: null
suggested_issue_title: "feat(schedule): implementar HU-042 bloqueos de agenda"
branch: null
pr: null
pr_url: null
depends_on:
  - "HU-040 y HU-041 integradas en main"
  - "Criterio de salida de B1 cumplido; HU-020–HU-024 integradas en main"
  - "CT-008 resuelta mediante un DEC-* antes de crear la migración"
  - "Issue real de HU-042 creado y enlazado antes de pasar a ready o ejecutar"
rules:
  - "RN-TEN-01"
  - "RN-BLQ-01"
  - "RN-BLQ-03"
  - "RN-BLQ-04"
  - "RN-DIS-05"
  - "RN-DIS-07"
  - "RN-IDE-01"
decisions:
  - "DEC-007"
  - "DEC-008"
  - "DEC-009"
  - "DEC-013"
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
  - "CA-042-01"
  - "CA-042-02"
  - "CA-042-03"
  - "CA-042-04"
  - "CA-042-05"
  - "CA-042-06"
  - "CA-042-07"
  - "CA-042-08"
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

# Implementar HU-042: bloqueos de agenda

## Instrucción para el agente

Este prompt permanece en draft: no existe issue real y CT-008 sigue abierta. No lo ejecutes ni inventes una resolución para las FK del modelo de referencia. La parte de impacto sobre citas se completa con la capacidad dueña de B3; esta HU no crea una tabla appointment parcial ni simula citas.

## Objetivo

Que el barbero pueda crear y mantener bloqueos puntuales y definiciones recurrentes que resten tiempo de trabajo, con siete tipos, fechas explícitas, excepciones individuales y eliminación lógica auditable.

## Preflight obligatorio

1. Verifica árbol limpio respecto de cambios ajenos, main actualizada y la integración de HU-040/HU-041.
2. Confirma issue real, rama feat/<issue>-hu042-bloqueos-agenda y metadatos actualizados antes de ejecutar.
3. Lee todos los source_docs y revisa la resolución de CT-008.
4. Si la contradicción sigue abierta, detente antes de crear contrato, migración o código.
5. Ejecuta Graphify sobre schedule, working_hour, excepciones, sesión/tenant y los stubs de contrato/frontend.
6. No hagas que el módulo de horarios importe tablas de citas; la integración con citas será un puerto o colaboración explícita en B3.

## Alcance incluido

- Bloqueo puntual con los siete tipos: break, lunch, unavailable, day_off, holiday, vacation, emergency.
- Definición recurrente semanal y lista explícita de fechas, con rango de vigencia y excepciones individuales.
- Intervalos semiabiertos, duración positiva, cruce de medianoche y validaciones de forma.
- Consulta de bloqueos vigentes e históricos; eliminación lógica que deja el registro y el actor, pero lo excluye de la proyección efectiva.
- Semántica segura ante reintentos de operaciones críticas y doble envío.
- Contrato OpenAPI, dominio/aplicación Go, migración Atlas, RLS, cliente tipado, pantalla Vue y pruebas reales.
- Punto de integración interno para que B3 pueda identificar citas afectadas sin que esta HU consulte appointment.

## Fuera de alcance

- Crear, modificar, reprogramar, cancelar o notificar citas.
- Resolver en esta HU la lista de citas afectadas de RN-BLQ-03; B3/B5 implementan la consulta, las decisiones del barbero y las notificaciones.
- Cálculo de disponibilidad pública y reserva.
- Borrado físico de bloqueos o series, purga, cascadas no aprobadas por CT-008.
- Calendario colombiano y excepciones de jornada: HU-041.
- Cambiar horarios recurrentes base: HU-040.
- Importar calendarios externos, pagos, inventario o funciones P1/P2.

## Estado existente que debe conservarse

- working_hour y las excepciones de HU-041 son las fuentes de jornada; los bloqueos se restan después como unión de intervalos.
- Los solapamientos entre bloqueos no deben descontarse dos veces de la proyección efectiva, aunque puedan conservarse como definiciones independientes.
- RN-BLQ-03 exige que un bloqueo urgente no sea rechazado por coincidir con citas; como appointment pertenece a B3, esta HU solo conserva el contrato de creación sin consultar esa tabla.
- RN-BLQ-04 exige eliminación lógica; no concedas DELETE al rol de aplicación sobre time_block ni time_block_series.
- El modelo físico de referencia propone time_block, time_block_series, time_block_series_date y time_block_series_exception; es insumo revisable y queda subordinado a AGENTS.md y CT-008.

## Trabajo requerido

### 1. Contrato HTTP

1. Define operaciones privadas separadas para bloqueo puntual, series, fechas explícitas, excepciones y retiro lógico; evita un endpoint genérico con payload ambiguo.
2. Declara los siete valores permitidos, weekly/lista explícita, rango efectivo, zona de la barbería, schemas cerrados y errores RFC 9457.
3. Explica en x-business-rules que crear el bloqueo no cancela ni reprograma citas y que la lista de citas afectadas llegará en B3.
4. No aceptes tenant del body; recursos ajenos o inexistentes responden 404 seguro.
5. Define una operación interna de proyección/consulta para disponibilidad futura sin materializar instancias de una serie en time_block.

### 2. Datos y backend Go

1. Resuelve CT-008 antes de la migración; conserva tercera forma normal, RLS forzada, grants mínimos, índices guiados por consultas y atlas.sum.
2. Usa checks para el vocabulario cerrado, intervalos, duración, rango efectivo, motivo acotado y marcas de eliminación coherentes.
3. Una serie semanal requiere día ISO; una lista explícita requiere filas de fecha; una excepción identifica la instancia excluida. Valida que una fecha explícita pertenezca a la serie.
4. Para editar una recurrencia implementa explícitamente “esta instancia”, “esta y las siguientes” o “toda la serie”; si requiere dividir una serie, documenta la operación y pruébala sin perder instancias.
5. La eliminación lógica fija actor/instante y deja de afectar la proyección; repetir la operación no crea un efecto adicional.
6. Usa transacciones cortas, parámetros y contexto de barbería. No llames red dentro de la transacción.
7. Modela la integración de citas como puerto pequeño y futuro, no como consulta directa desde schedule; no devuelvas conteos inventados.
8. Los logs solo incluyen IDs opacos, tipo, rango técnico y request_id; nunca razones libres con datos personales.

### 3. Frontend Vue

1. Construye la pantalla “Horarios y bloqueos” dentro del panel privado, con selector de barbero, próximos bloqueos y acceso a crear bloqueo.
2. Diferencia bloqueo puntual, semanal y lista de fechas; permite añadir excepciones individuales y retirar de forma lógica.
3. Muestra zona de la barbería, fechas y horas en formato local; explica que retirar no borra el registro.
4. No presentes botones de cancelar/reprogramar citas hasta que B3/B5 expongan esas operaciones reales.
5. Gestiona carga, vacío, error recuperable, conflicto, éxito y estado retirado; conserva datos escritos.
6. Verifica teclado, foco, 44 px, axe-core, zoom 200 %, reducción de movimiento y 320/360/768/1280 px.

## Pruebas y evidencia

- Dominio: los siete tipos, punto, semanal, lista, excepciones, edición por alcance, solape, medianoche, eliminación lógica e idempotencia.
- PostgreSQL real: dos tenants, RLS, checks, fecha hija inválida, serie cerrada/no vigente, grants sin DELETE, actor de retiro y decisión de FK de CT-008.
- HTTP/contrato: CRUD de punto/serie, fechas/excepciones, retiro, 401, 404, 409, 422, schemas cerrados y mensajes RFC 9457.
- Concurrencia: dos conexiones reales para crear/reintentar/retirar; barreras observables, nunca sleep.
- Componente/E2E: siete tipos, recurrente, fecha explícita, excepción, retiro y aislamiento entre tenants.
- Responsive/accesible: 320, 360, 768 y 1280 px, teclado, foco, zoom 200 %, axe-core y contraste.
- Prueba de integración futura documentada: la creación no consulta ni crea appointment; el impacto sobre citas queda trazado a B3/B5.

## Documentación y trazabilidad

- Actualiza OpenAPI/CHANGELOG, README, migraciones, atlas.sum, pruebas SQL, matriz, historial, plan y modelo de referencia si corresponde.
- Mantén enlazado el criterio de RN-BLQ-03 que queda pendiente de B3/B5; no lo marques completamente cubierto por una operación que no puede listar citas.
- Si CT-008 cambia por decisión, propaga el DEC-* a HUs, prompt, matriz y modelo antes de ejecutar.
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

Entrega Criterio | Estado | Prueba o evidencia para CA-042-01–CA-042-08, más una fila separada que indique qué parte de RN-BLQ-03 queda explícitamente para B3/B5.

## Git y PR

- Rama: feat/<issue>-hu042-bloqueos-agenda.
- Commit/título: feat(schedule): implementa HU-042 bloqueos de agenda.
- Usa Closes #<issue> solo cuando los ocho criterios propios estén verificados; usa Refs #<issue> si la integración B3/B5 permanece pendiente.
- No hagas push directo, force push, merge manual de main, DDL al arrancar ni cambios a migraciones aplicadas.

