---
prompt_id: "PROMPT-CHORE-187-SHELL-NAVEGACION-TAILORED-GRID-v1"
version: "1.0"
kind: "chore"
status: "executed"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-012"
related_hu: []
issue: "187"
issue_url: "https://github.com/bcaceres19/barberia/issues/187"
suggested_issue_title: "chore(web): shell y navegacion NAVA Tailored Grid"
branch: "chore/187-shell-navegacion-tailored-grid"
pr: 202
pr_url: "https://github.com/bcaceres19/barberia/pull/202"
depends_on:
  - "Fase 1 (issue #186/PR #199) integrada en main"
rules:
  - "RN-DIS-02"
  - "RN-DIS-04"
  - "RN-DIS-05"
decisions:
  - "DEC-077"
  - "DEC-078"
  - "DEC-079"
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/00-referencia-tailored-grid.png"
  - "docs/10-backlog/prompts/orchestration/redisenio-integral-nava-tailored-grid-v2.md"
  - ".claude/skills/browser-viewport-verification/SKILL.md"
created_at: "2026-09-02"
updated_at: "2026-09-02"
supersedes: null
superseded_by: null
---

# Fase 2 · Shell y navegación Tailored Grid

## Instrucción para el agente

Ejecuta únicamente la Fase 2 (Shell y navegación) de `PROMPT-ORCH-NAVA-ATLAS-v2`: traducir el contraste oscuro/claro del atlas al shell real y corregir la navegación móvil, sin añadir destinos fuera del router.

## Objetivo

El shell privado (`AppHeader`/`AppNav`) heredaba los tokens NAVA desde la Fase 1 anterior (#132/#199) pero no usaba la superficie tinta para orientación, y el dock listaba los seis destinos privados (Agenda, Servicios, Barberos, Horarios, Configuración, Servicios por barbero) en línea plana sin agrupar. `Bloqueos` (`/panel/bloqueos`) existía en el router desde HU-042 pero nunca tuvo entrada de navegación.

## Preflight ejecutado

1. Fase 1 (#186/PR #199) integrada en `main`.
2. Línea base capturada con Chrome DevTools real (cuenta QA `bacg20044@gmail.com`) en `/panel`: confirmó header y dock en superficie blanca plana, sin jerarquía tinta/marfil.
3. Auditoría de `apps/web/src/modules/*/index.ts`: confirmó que `schedulesNavItems` solo exportaba `Horarios`, sin entrada para `schedules-bloqueos` pese a que la ruta ya estaba registrada.
4. Se detectó que `mcp__claude-in-chrome__resize_window` no cambia el viewport real en este entorno (verificado con `window.innerWidth` en dos pestañas distintas); se creó la skill `browser-viewport-verification` y se usó el harness Playwright del repo (viewport real vía CDP) para la evidencia responsiva de esta fase.

## Alcance incluido

- `AppHeader`: superficie tinta (`--color-surface-strong`), `NavaWordmark` invertido, CTA "Nuevo turno" con el acento latón (`--color-brand-accent-surface`/`--color-brand-accent-text`, tokens ya existentes y sin consumidor hasta ahora). "Cerrar sesión" se conserva (`variant="secondary"`, ya legible sobre tinta).
- `AppNav`: `NavItem` gana un flag opcional `primary`; solo Agenda/Servicios/Barberos (marcados `primary: true`) quedan siempre visibles en el dock. El resto (Horarios, Bloqueos, Configuración, Servicios por barbero) vive en un menú "Más" (`BaseDialog`, reutiliza foco atrapado/Escape/restauración), junto con "Cerrar sesión".
- `schedulesNavItems` gana la entrada `Bloqueos` que faltaba.
- `useLogout` (nuevo composable en `auth/model`): extrae la lógica de cierre de sesión de `AppHeader` para que el menú "Más" la reutilice sin duplicar la guardia de doble envío ni la revocación de sesión.
- `RecordRow` (Fase 1) se reutiliza como fila del menú "Más": primer consumidor real de la primitiva.
- `panel-evidencia-responsiva.spec.ts` reescrito con un solo login reutilizado por prueba (mismo criterio que `barberos-evidencia-responsiva.spec.ts`): la versión anterior iniciaba sesión 20 veces en segundos (una por combinación prueba×ancho) y quedaba bloqueada por la propia defensa contra abuso de HU-007 al ejecutarse localmente. Se añadió cobertura del dock primario y del menú "Más" (incluida la llegada real a `/panel/bloqueos`).

## Fuera de alcance

- Composición de las 11 pantallas de negocio: Fases 3-12.
- Sidebar de escritorio: el shell conserva el dock fijo al pie en todos los anchos (comentario existente de `AppNav.vue`, decisión previa de `especificacion-frontend-nava.md`); no se introdujo un rail nuevo.
- Corrección de los otros specs E2E con aserciones desactualizadas (`acceso.spec.ts`, `panel.spec.ts`, `reto-telefonico.spec.ts` esperan encabezados que ya no existen, p. ej. "Agenda de hoy"/"Panel del barbero"): defecto preexistente no causado por este rediseño, no corregido aquí.

## Pruebas y evidencia

- `pnpm --filter @system-barbershop/web format`/`lint` (0 errores)/`typecheck`/`test:unit` (633 pruebas, incluye 11 de `AppNav.test.ts` reescritas para el menú "Más")/`build`: todos en verde.
- `git diff --check`: sin errores de espacio en blanco.
- Chrome DevTools real (cuenta QA, stack `barberia-qa-local`): header con `background-color: rgb(16, 27, 43)` y CTA con `rgb(184, 149, 90)`/`rgb(16, 27, 43)` confirmados por estilo computado; "Más" abre con foco en el botón de cerrar, Escape cierra y restaura el foco al disparador; navegación a `/panel/bloqueos` confirmada de punta a punta; consola sin errores.
- Playwright (`panel-evidencia-responsiva.spec.ts`, `chromium-desktop` y `chromium-mobile`, local contra `barberia-qa-local`): 2/2 pruebas en verde, evidencia real en 320/360/768/1280/1280-zoom200 (`normal.png`, `foco.png`, `mas-abierto.png`).

## Documentación y trazabilidad

- `docs/10-backlog/prompts/README.md`: nueva fila de índice para este prompt.
- `docs/00-control/historial-cambios.md`: entrada de esta entrega con el PR real.

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
| Dock sin truncamiento, con "Bloqueos" alcanzable | Cumplido | Evidencia Playwright 320/360/768/1280 + recorrido real en Chrome DevTools |
| Header en superficie tinta con foco destacado latón | Cumplido | Estilos computados verificados en Chrome DevTools real |
| Foco/Escape/restauración del menú "Más" | Cumplido | Recorrido real + `AppNav.test.ts` |
| Formato/lint/tipos/tests/build en verde | Cumplido | Comandos de verificación final |
| CI real (GitHub Actions) en verde | Cumplido | `gh pr checks 202` (4/4 verde); `gh pr view 202`: `MERGED`, `2026-09-03T02:59:00Z`, commit `9195b49` |

## Git y PR

- Commit: `chore(web): shell y navegacion Tailored Grid (#187)`.
- PR: [#202](https://github.com/bcaceres19/barberia/pull/202), `Refs #184`, `Closes #187`.
- Sin push directo ni force-push; squash-merge tras CI verde, rama eliminada.
