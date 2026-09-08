---
prompt_id: "PROMPT-HU-066-v1"
version: "1.0"
kind: "hu"
status: "ready"
target_agents:
  - "codex"
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-066"
related_hu: ["HU-004", "HU-060", "HU-064", "HU-065"]
issue: 225
issue_url: "https://github.com/bcaceres19/barberia/issues/225"
suggested_issue_title: "feat(booking): implementar HU-066 cancelación de turno por barbero"
branch: null
pr: null
pr_url: null
depends_on:
  - "HU-064 y HU-065 integradas en main"
  - "Issue #225 con CA-066-01 a CA-066-08"
rules: ["RN-CIT-01", "RN-CIT-03", "RN-CAN-03", "RN-CAN-04", "RN-HIS-01", "RN-HIS-02", "RN-TEN-01", "RN-IDE-01"]
decisions: ["DEC-011", "DEC-012", "DEC-014", "DEC-016", "DEC-017", "DEC-024", "DEC-035", "DEC-036", "DEC-037", "DEC-038", "DEC-041", "DEC-043", "DEC-077", "DEC-078", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-066-01", "CA-066-02", "CA-066-03", "CA-066-04", "CA-066-05", "CA-066-06", "CA-066-07", "CA-066-08"]
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
  - "docs/05-backend/base-datos.md"
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

# Implementar HU-066: cancelación de un turno por el barbero

## Instrucción para el agente

Implementa únicamente T6: `confirmed` → `cancelled_by_barber`. La operación debe ser privada, tenant-aware, idempotente, condicionada a la versión leída y atómica con su historial. No implementes cancelación pública ni efectos de B5.

## Objetivo

Que el barbero cancele una cita activa en cualquier momento, libere su intervalo cuando corresponda y conserve una atribución auditable sin aplicar cambios sobre una versión obsoleta.

## Preflight obligatorio

1. Comprueba árbol limpio, `main` actualizada por fast-forward e issue #225 abierto.
2. Lee completamente `AGENTS.md`, `CLAUDE.md`, `CONTRIBUTING.md` y cada `source_docs`.
3. Ejecuta Graphify sobre `booking`, historial, idempotencia, router privado y detalle Vue.
4. Confirma que `HU-064`/`HU-065` y el token opaco de versión están en `main`.
5. Revisa dudas y contradicciones; registra cualquier semántica ausente antes de codificar.
6. Actualiza este prompt a `in_progress`, registra la rama y crea `feat/225-hu066-cancelacion-barbero` desde `main`.

## Alcance incluido

- Comando OpenAPI privado y cerrado con `Idempotency-Key` y precondición de versión.
- T6 sin límite temporal, con tenant/actor/destino derivados en servidor.
- Estado y `appointment_cancelled_by_barber` en una transacción.
- Liberación inmediata de la exclusión para citas futuras canceladas.
- Repetición coherente, conflictos distinguibles y UI desde detalle.
- Pruebas en la capa más baja, PostgreSQL real, componente, E2E y evidencia responsive.

## Fuera de alcance

- Cancelación pública, plazo, motivo configurable o cancelación masiva.
- T3, T4, T7, T8, edición de datos o reapertura.
- Avisos, recordatorios, outbox o llamadas de red.
- Modificar una migración aplicada o relajar la exclusión.

## Estado existente que debe conservarse

- `appointment` ya restringe estados y la exclusión ignora los cancelados.
- `appointment_history` admite `appointment_cancelled_by_barber` y es append-only.
- `HU-004`/`DEC-043` gobiernan idempotencia concurrente.
- El detalle expone un `versionToken`; T2 demuestra precondición y traducción de conflictos.
- La agenda y el detalle usan “turno”; API, Go y datos conservan `appointment`.

## Trabajo requerido

1. Diseña el contrato antes del handler; no uses un `PATCH status` genérico.
2. Implementa caso de uso independiente de Chi/pgx con reloj solo para pruebas de antes/durante/después.
3. En una transacción tenant-aware, bloquea/condiciona la cita, verifica `confirmed` y versión, actualiza estado e inserta historial.
4. Trata repetición del mismo resultado como éxito sin nuevo evento y otro terminal como `409`.
5. Traduce recurso ajeno como `404` y conserva el protocolo de idempotencia en carreras.
6. Añade al detalle una confirmación accesible; ante conflicto conserva contexto y ofrece recargar.
7. Reutiliza del atlas solo el tratamiento terminal equivalente; el diálogo no tiene mockup exacto y se diseña dentro de NAVA.

## Pruebas y evidencia

- Dominio/HTTP: `200`, `400`, `401`, `404`, `409`, repetición, versión e idempotencia.
- PostgreSQL real con dos tenants: atomicidad, rollback, liberación de exclusión y carrera coordinada.
- Componente con confirmación, espera, error/conflicto, foco y `vitest-axe`.
- E2E: cancelar una futura, verificar historial/agenda y ocupar legítimamente la franja liberada.
- Capturas 320, 360, 768 y 1280 px; teclado y zoom 200 %.

## Documentación y trazabilidad

- Actualiza OpenAPI/CHANGELOG, cliente generado, README afectados, matriz, historial, plan y este prompt.
- Registra pruebas y cualquier migración Atlas nueva; no edites migraciones aplicadas.

## Verificación final

```text
pnpm run openapi:check-config
pnpm run openapi:lint
pnpm run openapi:bundle
cd apps/api && gofmt -l . && go vet ./... && go test -race ./... && go build ./...
cd ../web && pnpm run format && pnpm run lint && pnpm run typecheck && pnpm run test:unit && pnpm run build && pnpm run test:e2e
cd ../.. && pnpm run db:atlas:validate && git diff --check && graphify update .
```

Entrega `Criterio | Estado | Prueba o evidencia` para `CA-066-01`–`CA-066-08`. No declares cumplido lo no probado.

## Git y PR

- Rama: `feat/225-hu066-cancelacion-barbero`.
- Commit/PR: `feat(booking): implementa HU-066 cancelación por barbero`.
- Usa `Closes #225` solo si cubre todo el issue; si no, `Refs #225`.
- No hagas push directo, force push ni merge de `main`.
