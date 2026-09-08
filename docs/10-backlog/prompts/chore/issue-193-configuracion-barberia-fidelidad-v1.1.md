---
prompt_id: "PROMPT-CHORE-193-CONFIGURACION-BARBERIA-FIDELIDAD-v1.1"
version: "1.1"
kind: "chore"
status: "in_progress"
target_agents: ["codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-020"
related_hu: ["HU-009", "HU-012"]
issue: 193
issue_url: "https://github.com/bcaceres19/barberia/issues/193"
branch: "chore/193-configuracion-barberia-fidelidad"
pr: 231
pr_url: "https://github.com/bcaceres19/barberia/pull/231"
depends_on: ["Issue #192 integrado en main con CI verde", "Issue maestro #184"]
rules: ["RN-TEN-01", "RN-DAT-02"]
decisions: ["DEC-007", "DEC-077", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-020-01", "CA-020-02", "CA-020-03", "CA-020-04", "CA-020-05", "CA-020-06", "CA-020-07", "CA-020-08"]
created_at: "2026-09-08"
updated_at: "2026-09-08"
supersedes: "PROMPT-CHORE-193-CONFIGURACION-BARBERIA-FIDELIDAD-v1"
superseded_by: null
---

# Fidelidad de `/panel/barberia` al atlas NAVA

Esta versión conserva el prompt v1 situado en la rama documental
`docs/224-hu066-hu068`, sin reescribirlo. Ejecuta el issue
[#193](https://github.com/bcaceres19/barberia/issues/193) desde la rama
indicada, después del squash merge de #192 (`ce5c47e`).

## Alcance ejecutado

- Recompone la región de trabajo de `/panel/barberia` según el panel derecho y
  la vista móvil de `06-barberos-y-barberia.png`: encabezado editorial,
  formulario de 540 px, confirmación, etiquetas, campos y CTA de alta
  densidad.
- Conserva nombre, zona IANA, contactos, validación, guardado, prevención de
  doble envío y recuperación sin pérdida de valores de `HU-020`.
- Mantiene todos los nombres accesibles y los estados de carga, error de
  carga, validación, red, fallo inesperado y éxito. Los textos de ayuda se
  retiran solo visualmente porque no existen en el mockup; la semántica nativa
  de cada control y la validación permanecen intactas.
- No modifica API, backend, OpenAPI, cliente generado, migraciones, permisos
  ni rutas; tampoco incorpora logo, tema, canales, políticas u horarios.

## Contrato y evidencia visual

| Región | Referencia / tratamiento final |
| --- | --- |
| Panel exacto | Mitad Configuración de `docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/06-barberos-y-barberia.png` (lámina 1536×1024). |
| Escritorio | Columna clara centrada de 540 px, título editorial, confirmación verde y cuatro campos compactos con CTA de ancho total. |
| Móvil | Título corto visible `Barbería`, controles de 28 px y CTA de 30 px sin perder sus etiquetas, foco ni reflow. |
| Estados | Éxito, validación, carga, reintento y error de red conservan el contrato HU-020; los valores del formulario nunca aparecen en URL, logs o datos reales de evidencia. |
| Comparación | `apps/web/e2e/evidence/configuracion-barberia/fidelidad-193/comparacion/{desktop,mobile}/guardado-lado-a-lado.png` y `guardado-diff.png`. |
| Responsive | Capturas medidas mediante `window.innerWidth/innerHeight`: 320×720, 360×800, 768×1024, 1280×900 y 640×450 para zoom 200 %. |

La comparación deja una desviación heredada explícita: el `PrivateShell` de
#187 emplea cabecera y dock horizontales, mientras la lámina histórica dibuja
sidebar en escritorio y barra telefónica en móvil. No se altera ese shell
transversal en #193; la región de configuración clara, su jerarquía, controles
y estados sí se reproduce y la comparación no oculta dicha diferencia.

## Puerta final

Antes de integrar: auto-revisión, `git diff --check`, formato, lint, tipos,
componentes, E2E Chromium, build, comparación visual, evidencia responsive y
checks aplicables. El PR se integra por squash cuando esos controles estén en
verde.

## Pruebas ejecutadas

- `pnpm.cmd --dir apps/web exec vitest run src/modules/settings/pages/__tests__/SettingsPage.test.ts` — 11 pruebas pasan, incluidos los estados axe del formulario.
- `pnpm.cmd --dir apps/web typecheck`, `pnpm.cmd --dir apps/web lint` y `pnpm.cmd --dir apps/web build` — pasan; lint conserva únicamente warnings preexistentes fuera de este cambio.
- `pnpm.cmd --dir apps/web exec playwright test e2e/configuracion-barberia-fidelidad-mock.spec.ts --project=chromium-desktop` — 5 pruebas pasan con API simulada y datos ficticios.
- `node apps/web/e2e/configuracion-barberia-fidelidad-comparacion.mjs` — genera lado a lado y diff; sus métricas quedan en `.../comparacion/resumen.txt`.
