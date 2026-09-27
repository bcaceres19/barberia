---
prompt_id: "PROMPT-FEAT-GCAL-01-VINCULO-BARBERO-USUARIO-v1"
version: "1.0"
kind: "hu"
status: "draft"
target_agents:
  - "any"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu: []
issue: "pending"
issue_url: null
suggested_issue_title: "feat(barberos): vincular un barbero con su usuario del área privada"
branch: null
pr: null
pr_url: null
depends_on:
  - "PR del issue #284 integrado (DEC-099 a DEC-102)"
  - "DP-INT-01 resuelta (quién asigna el vínculo)"
rules:
  - "RN-TEN-01"
decisions:
  - "DEC-019"
  - "DEC-047"
  - "DEC-099"
  - "DEC-100"
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/06-api/estandar-openapi.md"
  - "docs/02-requisitos/historias-usuario.md (HU-021)"
  - "apps/api/internal/modules/staff"
  - "apps/api/internal/modules/schedule (barber)"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - ".agents/skills/visual-qa/SKILL.md"
created_at: "2026-09-26"
updated_at: "2026-09-26"
supersedes: null
superseded_by: null
---

# Vínculo explícito entre barbero y usuario autenticado

## Instrucción para el agente

Implementa únicamente la preocupación descrita en este archivo. Trabaja de forma autónoma hasta producir la entrega definida, sin ampliar el alcance ni inventar decisiones. Es una entrega de la orquestación [PROMPT-ORCH-GCAL-BARBERO-v1](../orchestration/google-calendar-barbero.md); léela para conocer el orden y el principio de autoridad (NAVA manda, Google es una vista).

## Objetivo

Existe una asignación explícita, opcional y única de `barber.staff_user_id` que permite resolver qué barbero es el usuario autenticado, con aislamiento por barbería.

## Preflight obligatorio

1. Comprueba que el árbol esté limpio y no sobrescribas cambios ajenos; trabaja en un worktree si hay cambios sin relación.
2. Actualiza `main` mediante fast-forward y verifica que cada dependencia de los metadatos esté integrada.
3. Lee completamente `AGENTS.md`, `CLAUDE.md` y cada archivo de `source_docs`, y las `DEC-*` enlazadas.
4. Comprueba `dudas-pendientes.md` y `contradicciones.md`. Si falta una decisión, registra el bloqueo antes de cambiar código.
5. Este prompt nace con `issue: pending` y en `draft`. Antes de ejecutarlo crea el issue real con criterios verificables (título sugerido en los metadatos), actualiza este archivo y el índice, y solo entonces pásalo a `ready`.
6. Si esta entrega introduce o toca una historia, redáctala en `docs/02-requisitos/historias-usuario.md` con su número real y criterios `CA-*` antes de codificar; no inventes números.
7. Crea la rama `<tipo>/<issue>-<descripcion>` desde `main` actualizada.

## Alcance incluido

- Migración Atlas: `barber.staff_user_id` nullable, FK compuesta `(barbershop_id, staff_user_id)` a `staff_user`, unicidad parcial cuando no es nulo.
- Caso de uso y repositorio para asignar/desasignar el vínculo y para resolver «barbero del usuario autenticado» mediante un puerto consumible por otros módulos.
- Contrato OpenAPI (contract-first), cliente tipado y UI mínima de asignación en la pantalla de barberos.

## Fuera de alcance

- Cualquier uso de Google Calendar.
- Inferir el vínculo por nombre, correo o teléfono; borrado o desactivación de barberos (siguen fuera por `DEC-047`).
- Roles o permisos nuevos más allá de lo que resuelva `DP-INT-01`.

## Estado existente que debe conservarse

- `barber` y `HU-021` (alta, listado, renombrar) sin cambios de comportamiento; sin `DELETE` sobre `barber`.
- Migraciones aplicadas inmutables: solo una migración nueva y `atlas.sum` actualizado.
- RLS y `FORCE ROW LEVEL SECURITY` de `barber` tal como están.

## Trabajo requerido

1. Redacta la historia con su número real y los `CA-*`; si `DP-INT-01` sigue abierta, detente y regístralo.
2. Actualiza OpenAPI primero; genera lint y cliente.
3. Crea la migración con Atlas, con `COMMENT` de la columna y privilegios coherentes con el estándar.
4. Implementa dominio, servicio, repositorio y handler sin dependencias cruzadas entre módulos (puerto en el consumidor).
5. Implementa la UI mínima con estados de carga, error y vacío.

## Pruebas y evidencia

- Constraint: un mismo `staff_user` no puede vincularse a dos barberos; un `staff_user` de otra barbería no puede vincularse (FK compuesta), con dos tenants en PostgreSQL real.
- Servicio: asignar, reasignar, desasignar; usuario sin barbero devuelve el error de dominio esperado.
- Componente Vue interactivo y evidencia responsive/accesible a los anchos de la guía visual.

## Documentación y trazabilidad

- Actualiza matriz de trazabilidad, historial, contrato OpenAPI, diccionario y diagrama de datos que realmente resulten afectados.
- Actualiza `database/modelo-fisico-referencia.sql`, diccionario y diagrama de datos, y anota en `DEC-047` que `DEC-100` levantó parcialmente su restricción.
- Actualiza los metadatos y el índice de este catálogo con issue, rama, PR y estado reales.

## Verificación final

```text
cd apps/api && go vet ./... && go test ./... && go test -race ./... && govulncheck ./...
cd apps/web && pnpm typecheck && pnpm lint && pnpm test:unit && pnpm build
pnpm run openapi:lint (y el bundle/cliente tipado definidos por el repositorio)
atlas migrate validate y pruebas contra PostgreSQL real con al menos dos tenants
tools/ai/validate-agent-system.sh --strict
```

Entrega una tabla `Criterio | Estado | Prueba o evidencia`. No declares cumplido aquello que no esté probado.

## Git y PR

- Commit/PR con Conventional Commits, sin secretos ni datos personales.
- `Closes #<issue>` solo si se cubre el issue completo; de lo contrario `Refs #<issue>`.
- No hagas push directo, force push ni merge de `main`; integra por PR y squash tras CI verde.
