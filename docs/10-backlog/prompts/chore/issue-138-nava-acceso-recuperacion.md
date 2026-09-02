---
prompt_id: "PROMPT-CHORE-138-NAVA-ACCESO-RECUPERACION-v1"
version: "1.0"
kind: "chore"
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
  - "HU-010"
  - "HU-011"
issue: "138"
issue_url: "https://github.com/bcaceres19/barberia/issues/138"
suggested_issue_title: "feat(web): NAVA en acceso, recuperación y sesión vencida"
branch: "feat/138-nava-acceso-recuperacion"
pr: 139
pr_url: "https://github.com/bcaceres19/barberia/pull/139"
depends_on:
  - "Fase 2 (issue #135, PR #136) integrada en main"
rules: []
decisions:
  - "DEC-055"
  - "DEC-056"
  - "DEC-058"
  - "DEC-077"
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/10-backlog/prompts/orchestration/adopcion-frontend-nava.md"
created_at: "2026-09-01"
updated_at: "2026-09-02"
supersedes: null
superseded_by: null
---

# Fase 3 · Acceso, recuperación y sesión vencida

## Instrucción para el agente

Implementa únicamente la Fase 3 (Seguridad y acceso) del prompt de orquestación `PROMPT-ORCH-NAVA-FRONTEND-v1`: recomponer la identidad visual de `LoginPage.vue`/`RecoveryPage.vue` y cerrar el hueco de "Sesión vencida", sin tocar ningún control de seguridad ni regla de negocio.

## Objetivo

`LoginPage.vue` y `RecoveryPage.vue` muestran `NavaWordmark` en vez de un texto "Barbería" fijo; el título de acceso es el literal "Accede a NAVA" del estándar; la redirección por sesión vencida (`installSessionHandling.ts`) ya tenía mecanismo, ahora tiene mensaje visible.

## Preflight ejecutado

1. Árbol limpio verificado; `main` actualizado por fast-forward tras integrar Fase 2 (PR #136).
2. Auditoría de `LoginPage.vue`, `RecoveryPage.vue`, `PhoneChallengeForm.vue`, los tres pasos de recuperación e `installSessionHandling.ts`: se confirmó que el copy de botones/pasos ya coincidía con los textos canónicos del estándar, y que `?motivo=sesion-expirada` ya se empujaba sin que ninguna pantalla lo leyera.
3. Issue real creado (`#138`) antes de escribir código; rama `feat/138-nava-acceso-recuperacion` desde `main` actualizada.

## Alcance incluido

- `NavaWordmark` en `LoginPage.vue` y `RecoveryPage.vue` (reemplaza `<p>Barbería</p>`).
- Título de acceso → "Accede a NAVA" (texto literal de `especificacion-frontend-nava.md` §7.1, no interpretación propia).
- Mensaje contextual "Tu sesión venció" en `LoginPage.vue` cuando `route.query.motivo === 'sesion-expirada'` y no hay un envío en curso/completado; desaparece en cuanto se intenta un envío.

## Fuera de alcance

- Cualquier control de seguridad, mensaje neutral anti-enumeración, límite de intentos o regla de negocio (`DEC-055/056/058/061-066`).
- `PhoneChallengeForm.vue` y los tres pasos de recuperación: su copy ya era conforme, sin cambios.
- Otras pantallas de negocio.

## Estado existente que se conservó

- Lógica completa de `attemptLogin` (CA-010-01 a CA-010-04, CA-012-02): doble envío bloqueado, credenciales inválidas limpian solo la contraseña, error de red conserva ambos campos, redirect seguro tras éxito. Sin tocar el `<script>` salvo el nuevo `computed` de sesión vencida.
- Los tres pasos de `RecoveryPage.vue` (solicitud, verificación, reset) y su máquina de estados (`Step`), sin cambios.

## Pruebas y evidencia

- `pnpm --filter @system-barbershop/web format`/`lint` (0 errores, mismos 27 warnings preexistentes)/`typecheck`: limpios.
- `pnpm --filter @system-barbershop/web test:unit`: 44 archivos, 594 pruebas en verde (591 previas + 3 nuevas: muestra el mensaje con `?motivo=sesion-expirada`, no lo muestra en una visita normal, lo oculta tras un envío).
- `pnpm --filter @system-barbershop/web build`: build de producción exitoso.
- `git diff --check`: sin errores de espacio en blanco.
- Verificación visual en navegador real: no realizada esta sesión (mismo motivo que la Fase 2, sin backend/PostgreSQL local); CI levanta el stack completo.

## Documentación y trazabilidad

- `docs/10-backlog/prompts/README.md`: nueva fila de índice.
- `docs/00-control/historial-cambios.md`: entrada de esta entrega con confirmación real de CI/merge.

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
| Wordmark NAVA en acceso y recuperación | Cumplido | `LoginPage.vue`, `RecoveryPage.vue` |
| Título "Accede a NAVA" | Cumplido | `LoginPage.vue` |
| Mensaje "Sesión vencida" visible y contextual | Cumplido | `LoginPage.test.ts` (3 pruebas nuevas) |
| Sin regresión en flujo de acceso (CA-010-*, CA-012-02) | Cumplido | Pruebas preexistentes de `LoginPage.test.ts` sin cambio de aserciones, todas en verde |
| Ningún control de seguridad modificado | Cumplido | Diff de PR #139 limitado a copy/composición visual |
| Formato/lint/tipos/tests/build en verde | Cumplido | Comandos de verificación final |
| CI real (GitHub Actions) en verde | Cumplido | `gh pr checks 139` (4/4 verde); `gh pr view 139 --json state,mergedAt,mergeCommit`: `MERGED`, `2026-09-02T03:48:14Z`, commit `e43ae8c` |

## Git y PR

- Commit: `feat(web): NAVA en acceso, recuperación y sesión vencida (Fase 3)`.
- PR: [#139](https://github.com/bcaceres19/barberia/pull/139), `Closes #138`.
- Sin push directo ni force-push; sin merge de `main` fuera del squash-merge autorizado tras CI verde.
