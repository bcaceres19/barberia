---
prompt_id: "PROMPT-TEST-174-REVISION-INTEGRAL-NAVA-v1"
version: "1.1"
kind: "test"
status: "executed"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu:
  - "HU-005"
  - "HU-006"
  - "HU-007"
  - "HU-008"
  - "HU-009"
  - "HU-010"
  - "HU-011"
  - "HU-012"
  - "HU-020"
  - "HU-021"
  - "HU-022"
  - "HU-023"
  - "HU-024"
  - "HU-040"
  - "HU-041"
  - "HU-042"
  - "HU-060"
  - "HU-061"
  - "HU-062"
  - "HU-063"
  - "HU-064"
  - "HU-065"
issue: 174
issue_url: "https://github.com/bcaceres19/barberia/issues/174"
suggested_issue_title: "test(web): validar rediseño integral NAVA en rutas P0 existentes"
branch: "test/174-revision-integral-nava"
pr: 181
pr_url: "https://github.com/bcaceres19/barberia/pull/181"
depends_on:
  - "Fases 6, 7, 9, 10 (issues #170, #171, #172, #173) integradas en main"
rules: []
decisions:
  - "DEC-079"
acceptance_criteria:
  - "Se recorrieron las 11 rutas P0 del inventario con datos sintéticos reales."
  - "0 violaciones axe-core reales, 0 errores de consola y sin scroll horizontal a 320 px en ninguna ruta."
  - "Cualquier defecto nuevo queda registrado en un issue propio, no corregido en esta rama."
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/10-backlog/prompts/orchestration/adopcion-frontend-nava.md"
created_at: "2026-09-02"
updated_at: "2026-09-02"
supersedes: null
superseded_by: null
---

# Fase 13 · Revisión integral NAVA

## Instrucción para el agente

Recorre las rutas P0 del inventario NAVA con datos sintéticos reales tras integrar las fases 6, 7, 9 y 10, y registra cualquier defecto nuevo en un issue propio sin corregirlo en esta rama.

## Objetivo

Cerrar el plan de orquestación NAVA (issue maestro #168) con una revisión integral real, no una declaración de cumplimiento sin evidencia.

## Preflight ejecutado

1. Árbol limpio verificado; `main` actualizado por fast-forward tras integrar #166–#173 (las cuatro fases pendientes detectadas en la Fase 0).
2. Se confirmó que otra sesión (Codex, `PROMPT-TEST-QA-PLATAFORMA-B0-B3-LUNA-v2`, archivos sin commitear en este árbol) ejecutó de forma independiente una QA integral más amplia (B0–B3) el mismo día y encontró, sin coordinación previa, el mismo tipo de brecha que esta Fase 13 documenta: specs E2E con aserciones de texto/URL desactualizadas frente a la UI NAVA real, y un umbral de intentos de HU-007 compartido por todo el entorno local. No se tocó ni se incorporó ese informe: es de otra sesión.

## Alcance incluido

Recorrido real (usuario QA aislado `qa.visual@ejemplo.test`, datos sintéticos, sin tocar cuentas ni datos de otras sesiones) de las 11 rutas P0 con superficie propia del inventario:

`/acceso`, `/recuperar-acceso`, `/panel` (agenda), `/panel/turnos/nuevo`, `/panel/turnos/:appointmentId` (detalle + diálogo de reprogramación), `/panel/barberia`, `/panel/barberos`, `/panel/servicios`, `/panel/servicios-por-barbero`, `/panel/horarios`, `/panel/bloqueos`.

Por cada ruta: captura a 1280 px y 320 px, verificación de scroll horizontal a 320 px, `axe-core` ejecutado en vivo sobre el DOM real (`color-contrast` y `heading-order` desactivados con el mismo criterio ya documentado en `LoginPage.test.ts`/`SettingsPage.test.ts`) y captura de errores de consola/página. Resultado consolidado en `apps/web/e2e/evidence/revision-integral-nava/report.json`.

Un hallazgo puntual (dock de navegación aparentemente superpuesto a mitad de la pantalla "Horarios" en la captura `fullPage` de 1280 px) se verificó por separado con una captura de viewport tras un scroll real: el dock permanece correctamente fijo al pie sin superponerse a contenido. Es un artefacto conocido de las capturas `fullPage` de Playwright con elementos `position: fixed` (se congelan en su posición de viewport al componer la imagen completa), no un defecto real — mismo criterio que ya se aceptó para las capturas `1280-zoom200` de fases anteriores.

## Fuera de alcance

- Cualquier corrección no trivial: el único hallazgo real conocido (truncamiento de etiquetas del dock a 320/360 px) ya está registrado por separado en el issue [#176](https://github.com/bcaceres19/barberia/issues/176) (shell, Fase 2) y no se corrige aquí.
- Rediseño de cualquier pantalla: ninguna requirió cambios de código en esta revisión.
- Ejecución completa de la suite Playwright existente contra el stack QA local compartido: bloqueada por dos brechas de entorno ya documentadas (ver "Pruebas y evidencia").

## Resultado por ruta

| Ruta | Scroll horizontal 320 px | Violaciones axe reales | Errores de consola |
| --- | --- | --- | --- |
| `/acceso` | No | 0 | 0 |
| `/recuperar-acceso` | No | 0 | 0 |
| `/panel` (agenda) | No | 0 | 0 |
| `/panel/turnos/nuevo` | No | 0 | 0 |
| `/panel/turnos/:appointmentId` | No | 0 | 0 |
| `/panel/barberia` | No | 0 | 0 |
| `/panel/barberos` | No | 0 | 0 |
| `/panel/servicios` | No | 0 | 0 |
| `/panel/servicios-por-barbero` | No | 0 | 0 |
| `/panel/horarios` | No | 0 | 0 |
| `/panel/bloqueos` | No | 0 | 0 |

Ninguna ruta produjo un defecto nuevo que exigiera un issue de corrección. El único hallazgo abierto conocido (truncamiento de etiquetas del dock) ya tenía su propio issue (#176) desde la Fase 7.

## Pruebas y evidencia

- `pnpm --filter @system-barbershop/web format`: archivos tocados por esta fase (ninguno de código) limpios; `base.css`/`tokens.css` mantienen las advertencias preexistentes de la Fase 1, sin relación.
- `pnpm --filter @system-barbershop/web lint`: 0 errores, mismas 27 advertencias preexistentes.
- `pnpm --filter @system-barbershop/web typecheck`: limpio.
- `pnpm --filter @system-barbershop/web test:unit`: 44 archivos, 614 pruebas en verde.
- `pnpm --filter @system-barbershop/web build`: build de producción exitoso.
- `git diff --check`: sin errores de espacio en blanco.
- Recorrido real de las 11 rutas con `axe-core` en vivo y captura de consola: `apps/web/e2e/evidence/revision-integral-nava/` (22 capturas + `report.json`).
- **`test:e2e` completo: no se declara cumplido.** Dos brechas de entorno, ya confirmadas de forma independiente por la QA integral B0–B3 de otra sesión el mismo día, bloquean la ejecución local contra el stack QA compartido (`barberia-qa-local-*`):
  1. Las cuentas de fixture (`duena.a@ejemplo.test`, `dueno.b@ejemplo.test`) no tienen un hash argon2id real en ese stack, solo el fixture SQL de pruebas de base de datos (ya documentado en las Fases 6/7/9/10).
  2. Varias aserciones (`toHaveURL(/\/panel$/)`) asumen que el login redirige siempre a `/panel` sin parámetros; con un tenant de un solo barbero, la agenda añade `?barberId=&date=` a la URL tras el login. Esto es un comportamiento correcto de la agenda para tenants de un barbero, no un defecto: la aserción es específica de la forma del tenant de cada spec (duena.a tiene dos barberos) y no se generaliza aquí para no alterar specs fuera del alcance de esta fase.
  Ninguna de las dos bloquea CI (no ejecuta `test:e2e`, solo `test:unit`).

## Documentación y trazabilidad

- `docs/10-backlog/prompts/README.md`: nueva fila de índice; fila de la Fase 9 (#171) confirmada como `executed`.
- `docs/00-control/matriz-trazabilidad.md` / `historial-cambios.md`: actualizados para cerrar la aplicación del contrato visual de `DEC-079` en el inventario P0 vigente.
- El issue maestro [#168](https://github.com/bcaceres19/barberia/issues/168) se cierra al fusionar esta entrega: todas sus fases hijas (6, 7, 9, 10, 13) quedan integradas.

## Verificación final

```text
pnpm --filter @system-barbershop/web format
pnpm --filter @system-barbershop/web lint
pnpm --filter @system-barbershop/web typecheck
pnpm --filter @system-barbershop/web test:unit
pnpm --filter @system-barbershop/web build
git diff --check
```

| Criterio | Estado | Prueba o evidencia |
| --- | --- | --- |
| 11 rutas P0 recorridas con datos sintéticos reales | Cumplido | `apps/web/e2e/evidence/revision-integral-nava/` |
| 0 violaciones axe reales, 0 errores de consola, sin scroll horizontal a 320 px | Cumplido | `report.json`; tabla de resultado por ruta |
| Ningún defecto nuevo sin registrar | Cumplido | Único hallazgo conocido ya registrado en #176 |
| Formato/lint/tipos/tests/build en verde | Cumplido | Comandos de verificación final (614/614 pruebas) |
| `test:e2e` completo contra el stack QA local | Bloqueado, no declarado cumplido | Dos brechas de entorno documentadas arriba, confirmadas por una QA independiente el mismo día; fuera de los checks de CI |

## Git y PR

- Rama: `test/174-revision-integral-nava`.
- PR: [#181](https://github.com/bcaceres19/barberia/pull/181), `Closes #174`, integrada en `main` con CI 4/4 verde (commit `548fa0f`, tras corregir el formato Prettier de `report.json`).
- Sin push directo ni force-push; squash-merge solo tras CI verde.
- Issue maestro [#168](https://github.com/bcaceres19/barberia/issues/168) cerrado manualmente tras confirmar que sus cinco fases hijas (6, 7, 9, 10, 13) están integradas.
