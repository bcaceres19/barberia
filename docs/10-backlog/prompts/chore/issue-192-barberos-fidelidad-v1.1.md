---
prompt_id: "PROMPT-CHORE-192-BARBEROS-FIDELIDAD-v1.1"
version: "1.1"
kind: "chore"
status: "in_progress"
target_agents: ["codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-021"
related_hu: ["HU-009", "HU-012"]
issue: 192
issue_url: "https://github.com/bcaceres19/barberia/issues/192"
branch: "chore/192-barberos-fidelidad"
pr: 230
pr_url: "https://github.com/bcaceres19/barberia/pull/230"
depends_on: ["Issue #191 integrado en main con CI verde", "Issue maestro #184"]
rules: ["RN-TEN-01", "RN-DAT-02"]
decisions: ["DEC-047", "DEC-077", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-021-01", "CA-021-02", "CA-021-03", "CA-021-04", "CA-021-05", "CA-021-06", "CA-021-07", "CA-021-08"]
created_at: "2026-09-08"
updated_at: "2026-09-08"
supersedes: "PROMPT-CHORE-192-BARBEROS-FIDELIDAD-v1"
superseded_by: null
---

# Fidelidad de `/panel/barberos` al atlas NAVA

Esta versión preserva el prompt v1 ejecutado desde la rama documental
`docs/224-hu066-hu068`, que no estaba integrado en `main`; no reescribe su
cuerpo. Ejecuta el issue [#192](https://github.com/bcaceres19/barberia/issues/192)
en la rama indicada, después del squash merge de #191 (`d2a5687`).

## Alcance ejecutado

- Recompone la región de trabajo de `/panel/barberos` según el panel izquierdo
  y los estados móviles de `06-barberos-y-barberia.png`: título editorial,
  acción principal, tabla-registro, monogramas, acción de edición, lista móvil,
  estado vacío y diálogos de alta/renombrado.
- Conserva exactamente la lectura, alta, renombrado, validación, idempotencia,
  carga, paginación y errores existentes de `HU-021`.
- No incorpora foto, rol, orden, vínculo automático, desactivación ni borrado
  (`DEC-047`); los monogramas decorativos siguen siendo el fallback autorizado.
- No modifica backend, OpenAPI, cliente tipado, migraciones, permisos ni rutas.

## Contrato y evidencia visual

| Región | Referencia / tratamiento final |
| --- | --- |
| Panel exacto | Mitad Barberos de `docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/06-barberos-y-barberia.png` (lámina 1536×1024). |
| Lista escritorio | Tabla clara de 520 px, cuatro filas de 92 px, divisor fino, monograma de 48 px y acción editorial. |
| Lista móvil | Filas de 64 px, monograma de 36 px, chevrón y CTA circular de 44 px. |
| Diálogos/estados | Alta, error recuperable que conserva valor, vacío y foco del campo; se conservan nombres, teclado y focus trap de `BaseDialog`. |
| Comparación | `apps/web/e2e/evidence/barberos/fidelidad-192/comparacion/{desktop,mobile}/lista-lado-a-lado.png` y `lista-diff.png`, vistas por el implementador. |
| Responsive | Capturas medidas mediante `window.innerWidth/innerHeight`: 320×720, 360×800, 768×1024, 1280×900 y 640×450 para zoom 200 %. |

La comparación deja una desviación explícita: el `PrivateShell` integrado por
#187 usa cabecera y dock horizontales, mientras la lámina histórica dibuja una
sidebar en escritorio y una barra telefónica en móvil. No se altera ese shell
transversal en #192 porque es propiedad de #187/`HU-012`; la región Barberos
clara, su jerarquía, filas, acciones y estados sí se reprodujeron y la
comparación no oculta la diferencia heredada.

## Pruebas ejecutadas

- `pnpm.cmd --dir apps/web exec vitest run src/modules/staff/pages/__tests__/StaffPage.test.ts` — 18 pruebas pasan.
- `pnpm.cmd --dir apps/web typecheck` y `pnpm.cmd --dir apps/web lint` — pasan; warnings preexistentes fuera de StaffPage.
- `pnpm.cmd --dir apps/web exec playwright test e2e/barberos-fidelidad-mock.spec.ts --project=chromium-desktop` — 6 pruebas pasan, con API simulada sin datos reales.
- `pnpm.cmd --dir apps/web build` — pasa.
- `node apps/web/e2e/barberos-fidelidad-comparacion.mjs` — genera comparación y diff; el resumen queda junto a la evidencia.

## Puerta final

**PASS para la región asignada de Barberos**, con la desviación de shell
documentada arriba y atribuida al componente propietario ya integrado. Antes de
integrar: auto-revisión, `git diff --check`, checks aplicables y PR squash que
cierra #192.
