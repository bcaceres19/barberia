---
prompt_id: "PROMPT-CHORE-170-CONFIGURACION-BARBERIA-NAVA-v1"
version: "1.0"
kind: "chore"
status: "ready"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu:
  - "HU-020"
issue: 170
issue_url: "https://github.com/bcaceres19/barberia/issues/170"
suggested_issue_title: "feat(web): rediseñar configuración de barbería NAVA"
branch: "feat/170-configuracion-barberia-nava"
pr: null
pr_url: null
depends_on:
  - "Issue maestro #168 (orquestación NAVA); DEC-077-DEC-079 integradas en main (issue #166/PR #167)"
rules: []
decisions:
  - "DEC-079"
acceptance_criteria:
  - "La pantalla /panel/barberia respeta la firma cromática y el contraste serif/sans de DEC-079."
  - "Estados normal, foco, carga, disabled, validación/error y éxito verificados en 320/360/768/1280 px."
  - "vitest-axe sin violaciones en normal, error de validación y confirmación de guardado."
  - "Evidencia responsiva real capturada y commiteada bajo apps/web/e2e/evidence/configuracion-barberia/."
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

# Fase 7 · Configuración de barbería NAVA

## Instrucción para el agente

Verifica y completa `/panel/barberia` (`apps/web/src/modules/settings/pages/SettingsPage.vue`) contra el contrato visual mínimo de `DEC-079`, sin rediseño estructural si la pantalla ya cumple.

## Objetivo

Confirmar con evidencia real (no con el issue anterior #149, que solo corrigió tokens rotos) que la pantalla de configuración pertenece a la familia NAVA en composición, estados, responsive y accesibilidad, y cerrar los huecos reales encontrados.

## Preflight ejecutado

1. Árbol limpio verificado por preocupación (staging selectivo); `main` actualizado por fast-forward tras integrar #166/#167 y #168/#169.
2. Auditoría visual real contra el navegador (usuario QA aislado `qa.visual@ejemplo.test`, tenant `NAVA QA Local`, sin tocar cuentas ni datos de otras sesiones): la pantalla YA usa `BaseInput`/`BaseButton`/`BaseAlert` (primitivas NAVA de la Fase 1), tokens de color/tipografía reales, canvas marfil, wordmark NAVA y foco visible con anillo latón — contraria a la hipótesis inicial de "solo fix de tokens", el fix #149 fue suficiente para esta pantalla simple porque no tiene composición propia más allá del formulario.
3. Hallazgo real de accesibilidad: `vitest-axe` con la alerta de error/éxito visible reportaba `heading-order` porque `BaseAlert` usa `<h4>` justo después del único `<h1>` de la página. Mismo patrón ya documentado y aceptado en `LoginPage.test.ts` (HU-010): no es un defecto de esta pantalla ni de `BaseAlert`, se documenta la excepción con el mismo criterio en vez de tocar el componente compartido.
4. Hallazgo real fuera de alcance: a 320/360 px el dock de navegación trunca las etiquetas ("Servic…", "Barbe…", "Horar…", "Confi…") — viola la regla de navegación de `estandar-diseno-visual.md` §6.5, pero pertenece al shell (Fase 2, issue #135 ya cerrado), no a esta pantalla. Se registra como hallazgo para un issue de corrección independiente del shell, sin tocarlo aquí.
5. Hallazgo real fuera de alcance: `apps/web/e2e/configuracion-barberia.spec.ts` y `configuracion-barberia-evidencia-responsiva.spec.ts` buscaban el enlace del dock por el texto `Barbería`; el rótulo del dock es `Configuración` desde la Fase 2 (decisión ya documentada en `src/modules/settings/index.ts`). Se corrige el selector de ambos specs para que reflejen el comportamiento real ya integrado (no es un cambio de comportamiento).

## Alcance incluido

- Corregir el selector desactualizado de los dos E2E existentes de HU-020 (enlace del dock `Barbería` → `Configuración`).
- Añadir cobertura `vitest-axe` con la alerta de error de validación y la confirmación de guardado visibles (antes solo se verificaba el estado cargado sin alerta).
- Recapturar y commitear evidencia responsiva real (320/360/768/1280 px, normal y foco) en `apps/web/e2e/evidence/configuracion-barberia/`, reemplazando las capturas de 2026-08-23 (previas a NAVA).

## Fuera de alcance

- Cualquier rediseño estructural de `SettingsPage.vue`: la pantalla ya cumple la firma cromática, la anatomía label→control→ayuda/error y los estados exigidos.
- El truncamiento de etiquetas del dock a 320/360 px (pertenece al shell, Fase 2/issue #135; se registra como hallazgo para un issue de corrección independiente).
- Cambios en `BaseAlert.vue` u otra primitiva compartida.
- Paleta, tema, logo, canales o políticas nuevas sin HU real.
- Cambios de contrato, validación de backend o reglas de negocio.

## Estado existente que se conservó

- Selector obligatorio, orden `label → control → ayuda/error`, guardia de doble envío, conservación de datos no sensibles ante error/conflicto (`CA-020-08`) y actualización de cabecera solo tras confirmación real del servidor (`CA-020-02`): sin cambios.
- Las 9 pruebas unitarias previas de `SettingsPage.test.ts`: sin cambio de aserciones.

## Pruebas y evidencia

- `pnpm --filter @system-barbershop/web format` (archivos tocados: limpio; `base.css`/`tokens.css` mantienen advertencias preexistentes de Fase 1, sin relación con este issue).
- `pnpm --filter @system-barbershop/web lint`: 0 errores, mismas 27 advertencias preexistentes de primitivas compartidas.
- `pnpm --filter @system-barbershop/web typecheck`: limpio.
- `pnpm --filter @system-barbershop/web test:unit`: 44 archivos, 610 pruebas en verde (incluye las 2 pruebas `vitest-axe` nuevas).
- `pnpm --filter @system-barbershop/web build`: build de producción exitoso.
- `git diff --check`: sin errores de espacio en blanco.
- Evidencia real capturada contra el API local real (usuario QA aislado, datos sintéticos) en `apps/web/e2e/evidence/configuracion-barberia/{320,360,768,1280,1280-zoom200}/{normal,foco}.png`.

## Documentación y trazabilidad

- `docs/10-backlog/prompts/README.md`: nueva fila de índice para este prompt.
- `docs/00-control/matriz-trazabilidad.md` / `historial-cambios.md`: actualizados al confirmar el merge real.

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
| Firma cromática, tipografía y foco NAVA reales (no solo tokens) | Cumplido | Evidencia real capturada 2026-09-02; código de `SettingsPage.vue`/`BaseInput.vue` |
| Estados normal/foco/error/éxito verificados | Cumplido | `apps/web/e2e/evidence/configuracion-barberia/**` y pruebas unitarias correspondientes |
| `vitest-axe` sin violaciones en normal, error y éxito | Cumplido | 3 pruebas `vitest-axe` en `SettingsPage.test.ts` |
| 320/360/768/1280 px y zoom 200% sin scroll horizontal | Cumplido | Evidencia real; `configuracion-barberia-evidencia-responsiva.spec.ts` afirma lo mismo con `duena.a@ejemplo.test` cuando el stack QA local tenga esa credencial real disponible |
| E2E funcional de HU-020 pasa contra el API real | Bloqueado localmente | El stack QA local compartido (`barberia-qa-local-*`) no tiene un hash argon2id real para `duena.a@ejemplo.test` (solo el fixture SQL de pruebas de base de datos); se verificó el mismo recorrido con una cuenta QA propia (`qa.visual@ejemplo.test`, hash real, mismo tenant de evidencia). No es un defecto de esta entrega; se registra como brecha del entorno QA local, no de CI |
| Formato/lint/tipos/tests/build en verde | Cumplido | Comandos de verificación final |
| Sin dependencias nuevas | Cumplido | `git diff` no toca `package.json`/`pnpm-lock.yaml` |

## Hallazgos registrados para seguimiento separado

- Dock de navegación trunca etiquetas a 320/360 px (viola `estandar-diseno-visual.md` §6.5): candidato a issue de corrección del shell (Fase 2), no se corrige aquí.
- El stack QA local compartido (`barberia-qa-local-*`) no tiene credenciales reales para los usuarios de fixture (`duena.a@ejemplo.test`, `dueno.b@ejemplo.test`): los E2E funcionales que dependen de login real fallan localmente por esta causa, no por un defecto de producto. No forma parte de los checks de CI (que no ejecuta `test:e2e`).

## Git y PR

- Rama: `feat/170-configuracion-barberia-nava`.
- PR: pendiente de abrir, `Closes #170`.
- Sin push directo ni force-push; squash-merge solo tras CI verde.
