---
prompt_id: "PROMPT-HU-095-v1"
version: "1.0"
kind: "hu"
status: "in_progress"
target_agents: ["codex", "claude"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-095"
related_hu: ["HU-094", "HU-096"]
issue: 254
issue_url: "https://github.com/bcaceres19/barberia/issues/254"
suggested_issue_title: "feat(public-booking): implementar HU-095 fechas y horarios"
branch: "feat/254-hu095-fechas-horarios"
pr: null
pr_url: null
depends_on: ["HU-094 integrada (satisfecho, PR #253)", "Issue real con CA-095-01 a CA-095-05 (satisfecho, issue #254)"]
rules: ["RN-DIS-01", "RN-DIS-02", "RN-DIS-03", "RN-DIS-04", "RN-DIS-05", "RN-DIS-06", "RN-DIS-07", "RN-CON-04", "RN-DAT-02"]
decisions: ["DEC-005", "DEC-006", "DEC-007", "DEC-018", "DEC-019", "DEC-020", "DEC-077", "DEC-078", "DEC-079"]
acceptance_criteria: ["CA-095-01", "CA-095-02", "CA-095-03", "CA-095-04", "CA-095-05"]
source_docs: ["AGENTS.md", "CLAUDE.md", "CONTRIBUTING.md", "docs/00-control/registro-decisiones.md", "docs/01-producto/alcance-mvp.md", "docs/01-producto/reglas-negocio.md", "docs/02-requisitos/historias-usuario.md", "docs/03-desarrollo/estandar-frontend-vue.md", "docs/03-desarrollo/estandar-diseno-visual.md", "docs/03-desarrollo/estrategia-pruebas.md", "docs/06-api/estandar-openapi.md", "apps/web/src/modules"]
created_at: "2026-09-10"
updated_at: "2026-09-11"
supersedes: null
superseded_by: null
---

# Implementar HU-095: exploración pública de fechas y horarios

## Instrucción para el agente

Implementa solo la selección visual de fecha/franja consumiendo HU-094; no reserves ni confirmes.

## Objetivo

Que el cliente elija una franja real en la zona de la barbería sin interpretar la consulta como un hold.

## Preflight obligatorio

1. Confirma HU-094 integrada, issue real, rama y fuentes vigentes.
2. Ejecuta Graphify sobre flujo público, cliente generado y utilidades de fecha civil.
3. Determina si existe mockup exacto asignado; si existe, activa fidelidad y comparación visual.

## Alcance incluido

- Navegación dentro de ventana, carga bajo demanda y selección accesible.
- Hora/fecha/duración/zona explícitas y estados día vacío/error/reintento.
- Cancelación o descarte de respuestas tardías al cambiar contexto.

## Fuera de alcance

- Hold, tiempo real, confirmación, datos personales o alternativas tras conflicto.

## Estado existente que debe conservarse

- Cliente OpenAPI tipado, `shared/time` y arquitectura `app → modules → shared`.

## Trabajo requerido

1. Modela estado local explícito sin `any` ni store global.
2. Integra el cliente generado y evita carreras entre consultas.
3. Construye la UI NAVA con anuncios, foco y objetivos táctiles.

## Pruebas y evidencia

- Unitarias del modelo/carreras; componente; E2E con zona del dispositivo distinta y día vacío.
- Capturas 320/360/768/1280, teclado, zoom 200 % y axe-core.

## Documentación y trazabilidad

Actualiza README frontend, matriz, HU, evidencia y prompt/índice.

## Verificación final

```text
cd apps/web && pnpm run format && pnpm run lint && pnpm run typecheck && pnpm run test:unit && pnpm run build && pnpm run test:e2e
cd ../.. && git diff --check && graphify update .
```

Entrega tabla para `CA-095-01`–`CA-095-05`.

## Git y PR

- Commit/PR: `feat(public-booking): implementa HU-095 fechas y horarios`.
