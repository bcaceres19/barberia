---
prompt_id: "PROMPT-OPS-GCAL-05-PUESTA-EN-MARCHA-v1"
version: "1.0"
kind: "ops"
status: "draft"
target_agents:
  - "any"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu: []
issue: "pending"
issue_url: null
suggested_issue_title: "docs(integraciones): guía de puesta en marcha y revisión integral de Google Calendar"
branch: null
pr: null
pr_url: null
depends_on:
  - "PROMPT-FEAT-GCAL-03-v1 integrado"
  - "PROMPT-FEAT-GCAL-04-v1 integrado"
rules:
  - "RN-TEN-01"
decisions:
  - "DEC-099"
  - "DEC-101"
  - "DEC-102"
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
  - "docs/04-arquitectura/stack-despliegue-operacion.md"
  - ".agents/skills/change-review/SKILL.md"
created_at: "2026-09-26"
updated_at: "2026-09-26"
supersedes: null
superseded_by: null
---

# Puesta en marcha y verificación integral de Google Calendar

## Instrucción para el agente

Implementa únicamente la preocupación descrita en este archivo. Trabaja de forma autónoma hasta producir la entrega definida, sin ampliar el alcance ni inventar decisiones. Es una entrega de la orquestación [PROMPT-ORCH-GCAL-BARBERO-v1](../orchestration/google-calendar-barbero.md); léela para conocer el orden y el principio de autoridad (NAVA manda, Google es una vista).

## Objetivo

La integración queda revisada de punta a punta, operable en local, desarrollo y producción, y con las verificaciones de seguridad y tenancy documentadas.

## Preflight obligatorio

1. Comprueba que el árbol esté limpio y no sobrescribas cambios ajenos; trabaja en un worktree si hay cambios sin relación.
2. Actualiza `main` mediante fast-forward y verifica que cada dependencia de los metadatos esté integrada.
3. Lee completamente `AGENTS.md`, `CLAUDE.md` y cada archivo de `source_docs`, y las `DEC-*` enlazadas.
4. Comprueba `dudas-pendientes.md` y `contradicciones.md`. Si falta una decisión, registra el bloqueo antes de cambiar código.
5. Este prompt nace con `issue: pending` y en `draft`. Antes de ejecutarlo crea el issue real con criterios verificables (título sugerido en los metadatos), actualiza este archivo y el índice, y solo entonces pásalo a `ready`.
6. Si esta entrega introduce o toca una historia, redáctala en `docs/02-requisitos/historias-usuario.md` con su número real y criterios `CA-*` antes de codificar; no inventes números.
7. Crea la rama `<tipo>/<issue>-<descripcion>` desde `main` actualizada.

## Alcance incluido

- Revisión independiente con `change-review` de las entregas 01 a 04 (arquitectura, tenancy, tokens, idempotencia, aislamiento de NAVA frente a Google).
- Guía operativa: Google Cloud, pantalla de consentimiento y verificación para producción, dominio, redirect URI por ambiente, rotación de la clave de tokens y respuesta ante `reauth_required`.
- Comprobación de los casos aplicables de la orquestación contra las pruebas existentes; los que falten se registran como issues, no se ocultan.

## Fuera de alcance

- Nuevas funciones o cambios de contrato.

## Estado existente que debe conservarse

- Todo el código integrado en las entregas anteriores.

## Trabajo requerido

1. Reúne la matriz caso → prueba con enlaces reales.
2. Ejecuta la revisión y registra hallazgos con severidad.
3. Redacta la guía sin credenciales reales ni datos personales.

## Pruebas y evidencia

- Cada caso aplicable enlaza una prueba ejecutada o un issue abierto.
- Comprobación manual guiada contra una cuenta de prueba de Google, si el propietario la provee; de lo contrario se declara no ejecutada.

## Documentación y trazabilidad

- Actualiza matriz de trazabilidad, historial, contrato OpenAPI, diccionario y diagrama de datos que realmente resulten afectados.
- Actualiza el README de `apps/api` y el documento de despliegue.
- Actualiza los metadatos y el índice de este catálogo con issue, rama, PR y estado reales.

## Verificación final

```text
cd apps/api && go vet ./... && go test ./... && go test -race ./... && govulncheck ./...
cd apps/web && pnpm typecheck && pnpm lint && pnpm test:unit && pnpm build
pnpm run openapi:lint (y el bundle/cliente tipado definidos por el repositorio)
tools/ai/validate-agent-system.sh --strict
```

Entrega una tabla `Criterio | Estado | Prueba o evidencia`. No declares cumplido aquello que no esté probado.

## Git y PR

- Commit/PR con Conventional Commits, sin secretos ni datos personales.
- `Closes #<issue>` solo si se cubre el issue completo; de lo contrario `Refs #<issue>`.
- No hagas push directo, force push ni merge de `main`; integra por PR y squash tras CI verde.
