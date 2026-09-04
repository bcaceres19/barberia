---
prompt_id: "PROMPT-CHORE-189-AGENDA-DIARIA-FIDELIDAD-v3"
version: "3.0"
kind: "chore"
status: "superseded"
target_agents:
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-062"
related_hu:
  - "HU-009"
  - "HU-012"
  - "HU-063"
issue: "189"
issue_url: "https://github.com/bcaceres19/barberia/issues/189"
branch: "chore/189-agenda-diaria-fidelidad"
pr: null
pr_url: null
depends_on:
  - "PROMPT-CHORE-189-AGENDA-DIARIA-FIDELIDAD-v2"
  - "Atlas panel-agenda-eventos integrado mediante PR #216"
rules: []
decisions:
  - "DEC-074"
  - "DEC-075"
  - "DEC-077"
  - "DEC-079"
  - "DEC-080"
acceptance_criteria:
  - "CA-062-01"
  - "CA-062-02"
  - "CA-062-03"
  - "CA-063-01"
  - "CA-063-02"
source_docs:
  - "AGENTS.md"
  - "docs/10-backlog/prompts/chore/issue-189-agenda-diaria-fidelidad-v2.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/panel-agenda-eventos/README.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "apps/web/e2e/evidence/panel-fidelidad-189-desviaciones.md"
created_at: "2026-09-03"
updated_at: "2026-09-03"
supersedes: "PROMPT-CHORE-189-AGENDA-DIARIA-FIDELIDAD-v2"
superseded_by: "PROMPT-CHORE-189-AGENDA-DIARIA-FIDELIDAD-v4"
---

# Fidelidad visual de `/panel` con mocks verificables y demo local

## Instrucción para el agente

Continúa #189 usando solo datos ficticios. La persona propietaria autorizó
además una vista local en vivo para revisar el diseño: un servidor efímero de
mock puede atender exclusivamente las cuatro lecturas que `/panel` consume,
mediante el proxy local de Vite. Esta ampliación sustituye la exclusividad de
interceptores Playwright de v2; no autoriza mocks en el producto.

## Alcance incluido

- `apps/web/e2e/panel-fidelidad-mock.spec.ts`: doce estados del atlas en
  escritorio `1440×1024 @1x` y móvil `420×935 @2x`, reloj fijo y capturas.
- `apps/web/e2e/mock-panel-api.mjs`: servidor local, temporal y aislado para
  la demo de `/panel`; solo sesión sintética, barberos, zona horaria y agenda.
- Ajustes de composición de `DailyAgendaPage` que pertenezcan al atlas:
  controles, eje, filas y estados, sin cambiar sus reglas de negocio.
- Registro explícito de diferencias en
  `apps/web/e2e/evidence/panel-fidelidad-189-desviaciones.md`.

## Fuera de alcance

- Cualquier mock en `src/`, cliente HTTP, API, OpenAPI, backend, PostgreSQL,
  autenticación real, fixtures persistentes o datos personales.
- Cambiar el shell privado (`AppHeader`, `AppNav`, `PrivateShell`), que afecta
  rutas ajenas y tiene su propia trazabilidad de #187.
- Declarar equivalencia píxel a píxel si queda una diferencia registrada.

## Verificación requerida

1. Usa el mismo conjunto de nombres, servicios, horarios y estados ficticios
   que el evento 01 del atlas para la agenda base.
2. Congela el reloj de las capturas para que «Ahora» sea reproducible.
3. Comprueba los 24 casos; guarda la evidencia bajo
   `apps/web/e2e/evidence/panel/fidelidad-189/mock/`.
4. Ejecuta las pruebas de componente pertinentes, lint, typecheck y build.
5. Conserva los E2E funcionales contra backend como una preocupación separada.

## Entrega

No hagas push, merge ni abras PR sin autorización. Usa `Refs #189`. Entrega
el resultado real de cada verificación y enlaza esta versión del prompt.
