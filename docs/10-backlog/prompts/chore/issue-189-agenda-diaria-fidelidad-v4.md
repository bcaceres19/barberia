---
prompt_id: "PROMPT-CHORE-189-AGENDA-DIARIA-FIDELIDAD-v4"
version: "4.0"
kind: "chore"
status: "executed"
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
  - "PROMPT-CHORE-189-AGENDA-DIARIA-FIDELIDAD-v3"
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
  - "docs/10-backlog/prompts/chore/issue-189-agenda-diaria-fidelidad-v3.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/panel-agenda-eventos/README.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "apps/web/e2e/evidence/panel-fidelidad-189-desviaciones.md"
created_at: "2026-09-03"
updated_at: "2026-09-03"
supersedes: "PROMPT-CHORE-189-AGENDA-DIARIA-FIDELIDAD-v3"
superseded_by: null
---

# Fidelidad integral del atlas de `/panel`

## Autorización y objetivo

La persona propietaria autorizó el 2026-09-03 extender la fidelidad de #189
al shell visible en los PNG del atlas. La app debe reproducir el diseño
propuesto para `/panel`: header, navegación, controles, timeline, estados,
selector, tarjetas y composición móvil/escritorio, con datos mock solamente.

## Alcance incluido

- Los componentes visibles del shell privado cuando se visita `/panel`, aun
  si son compartidos: `AppHeader`, `AppNav` y `PrivateShell`.
- `DailyAgendaPage`, `AgendaSkeleton`, `BarberSelect`, los badges de estado y
  el harness mock de `apps/web/e2e`.
- Comparación iterativa de los 24 PNG con capturas equivalentes y registro de
  toda excepción funcional no evitable.

## Límites

- No eliminar navegación, logout ni accesibilidad existentes solo porque un
  PNG estático no muestre la acción; la función prevalece y se compone de la
  forma menos intrusiva posible.
- No tocar API, OpenAPI, backend, base de datos, autenticación real ni
  introducir mocks en `src/`.
- No declarar fidelidad píxel a píxel mientras exista una desviación.

## Trabajo obligatorio

1. Medir y corregir primero el evento 01 en los dos viewports: espaciado,
   escala, filas, insignias, eje, selector y dock.
2. Repetir la comparación en carga (05), selector abierto (12), vacíos y
   errores antes de regenerar los 24 casos.
3. Usar los datos y reloj sintéticos del atlas; capturar móvil a `420×935 @2x`.
4. Conservar pruebas de componente y agregar/ajustar solo aserciones que
   describan la nueva geometría, nunca para ocultar una diferencia.
5. Ejecutar E2E mock, pruebas de componente, lint, typecheck y build.

## Entrega

No hagas push, merge ni PR sin autorización. Mantén `Refs #189` y actualiza
este prompt, el catálogo y las desviaciones con el resultado real.

## Resultado (2026-09-03)

Ejecutado. Se corrigieron, sobre el trabajo ya avanzado con Codex: "Cerrar
sesión" quedó invisible salvo con teclado (`clip: rect(0 0 0 0)` sin
alternativa visible al mouse) — restaurado como texto discreto siempre
visible, cumpliendo el límite explícito de este prompt ("no eliminar...
logout... la función prevalece"); el detalle de cada ficha de la línea
temporal (rango/nombre/servicio vs. solo hora+nombre) no distinguía ancho
real, mostrando las tres líneas siempre y desbordando en fichas angostas —
ahora depende del ancho real de la ficha, igual que el atlas; las insignias
"Ahora" y "Cambio de día" podían superponerse y volverse ilegibles cuando
coinciden en x (turno nocturno que empieza "hoy") — ahora viven en bandas
verticales separadas. Formato de hora corregido a 24h (ya lo traía el avance
de Codex). Verificado: `pnpm format`/`lint`/`typecheck`/`test:unit`
(708/709, 1 fallo confirmado no relacionado y no reproducible en aislado)
y `build` en verde; los 24 pares evento×viewport se regeneraron con
`panel-fidelidad-mock.spec.ts` y se revisaron visualmente contra el atlas.
Detalle completo en `apps/web/e2e/evidence/panel-fidelidad-189-desviaciones.md`.
No se hizo push ni se abrió PR — pendiente de autorización explícita antes
de ese paso.
