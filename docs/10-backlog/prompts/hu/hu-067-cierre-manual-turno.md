---
prompt_id: "PROMPT-HU-067-v1"
version: "1.0"
kind: "hu"
status: "blocked"
target_agents: ["codex", "claude"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-067"
related_hu: ["HU-004", "HU-060", "HU-064", "HU-066", "HU-068"]
issue: 226
issue_url: "https://github.com/bcaceres19/barberia/issues/226"
suggested_issue_title: "feat(booking): implementar HU-067 cierre manual de turno"
branch: null
pr: null
pr_url: null
depends_on:
  - "HU-066 integrada en main con CI verde"
  - "Issue #226 con CA-067-01 a CA-067-08"
rules: ["RN-CIT-01", "RN-CIT-03", "RN-CIT-05", "RN-HIS-01", "RN-HIS-02", "RN-TEN-01", "RN-IDE-01"]
decisions: ["DEC-002", "DEC-014", "DEC-016", "DEC-017", "DEC-018", "DEC-024", "DEC-035", "DEC-036", "DEC-037", "DEC-038", "DEC-041", "DEC-043", "DEC-077", "DEC-078", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-067-01", "CA-067-02", "CA-067-03", "CA-067-04", "CA-067-05", "CA-067-06", "CA-067-07", "CA-067-08"]
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/matriz-trazabilidad.md"
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
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/06-api/estandar-openapi.md"
  - "docs/10-backlog/plan-bloques.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/detalle-turno-eventos/README.md"
  - "api/openapi/paths/private-appointments.yaml"
  - "apps/api/internal/modules/booking"
  - "apps/web/src/modules/agenda"
created_at: "2026-09-08"
updated_at: "2026-09-08"
supersedes: null
superseded_by: null
---

# Implementar HU-067: cierre manual como atendido o no asistió

## Instrucción para el agente

Implementa únicamente T4 manual y T7. Este prompt está `blocked` hasta que HU-066 esté integrada; cuando se cumpla, relee las fuentes, actualiza metadatos a `ready` y ejecuta el issue #226 sin mezclar T8 ni cierre automático.

## Objetivo

Que el barbero cierre una cita iniciada como `completed` o `no_show`, con un único resultado terminal y un evento auditable, sin liberar el intervalo ni falsear métricas.

## Preflight obligatorio

1. Detente si HU-066 no está integrada en `main` con CI verde.
2. Comprueba árbol limpio, actualiza `main`, lee todos los `source_docs` y revisa dudas/contradicciones.
3. Confirma issue #226, registra `ready`/`in_progress` y crea `feat/226-hu067-cierre-manual-turno`.
4. Ejecuta Graphify sobre `booking`, estados, historial, exclusión, idempotencia y detalle/agenda Vue.

## Alcance incluido

- Dos comandos contract-first: completar y marcar inasistencia.
- Solo `confirmed` con `starts_at <= reloj del servidor`.
- Estado + evento correspondiente dentro de una transacción tenant-aware.
- Ambos terminales siguen ocupando agenda y conservan intervalo/snapshots.
- Idempotencia, versión, conflictos, UI contextual, pruebas y evidencia.

## Fuera de alcance

- T8, cancelación, T3, cierre automático, configuración del modo o worker.
- Aviso de inasistencia, notificaciones o recordatorios de B5.
- Duración real o movimiento de citas siguientes.

## Estado existente que debe conservarse

- Los valores `completed`/`no_show`, eventos y predicado de ocupación existen desde HU-060.
- HU-064 expone detalle/historial/versión; HU-066 deja el patrón de comando terminal.
- El atlas de detalle representa `completed` y documenta la variante `no_show`.

## Trabajo requerido

1. Declara dos operaciones cerradas en OpenAPI; no aceptes `status`, actor ni tiempo efectivo desde el cliente.
2. Implementa servicios de dominio independientes de transporte/persistencia y un reloj inyectable.
3. Valida la frontera exacta de `starts_at` con reloj del servidor.
4. Actualiza estado e historial de forma atómica, sin tocar intervalo o snapshots.
5. Conserva exclusión; traduce repetición igual a no-op y destino distinto a conflicto.
6. Resuelve carreras completar/no-show/cancelar mediante la fila/versionado e idempotencia existentes.
7. Añade acciones claras al detalle y refresca detalle/historial/agenda tras éxito.

## Pruebas y evidencia

- Dominio/HTTP para ambos comandos, frontera temporal, terminales, versión e idempotencia.
- PostgreSQL real con dos tenants y carrera coordinada de tres resultados.
- Componentes y E2E para completar y no-show, incluido el historial.
- 320, 360, 768, 1280 px, teclado, foco, zoom 200 % y axe-core.

## Documentación y trazabilidad

- Actualiza contrato/CHANGELOG/cliente, README, matriz, historial, plan y este prompt.
- Si surge una duda material, regístrala y vuelve el prompt a `blocked` antes de codificar.

## Verificación final

```text
pnpm run openapi:check-config
pnpm run openapi:lint
pnpm run openapi:bundle
cd apps/api && gofmt -l . && go vet ./... && go test -race ./... && go build ./...
cd ../web && pnpm run format && pnpm run lint && pnpm run typecheck && pnpm run test:unit && pnpm run build && pnpm run test:e2e
cd ../.. && pnpm run db:atlas:validate && git diff --check && graphify update .
```

Entrega `Criterio | Estado | Prueba o evidencia` para `CA-067-01`–`CA-067-08`.

## Git y PR

- Rama: `feat/226-hu067-cierre-manual-turno`.
- Commit/PR: `feat(booking): implementa HU-067 cierre manual de turno`.
- `Closes #226` solo con todo el issue; de lo contrario `Refs #226`.
