---
prompt_id: "PROMPT-ORCH-LENGUAJE-REGLADO-NAVA-v1"
version: "1.0"
kind: "orchestration"
status: "ready"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu:
  - "HU-005"
  - "HU-007"
  - "HU-008"
  - "HU-009"
  - "HU-010"
  - "HU-011"
  - "HU-012"
issue: "211"
issue_url: "https://github.com/bcaceres19/barberia/issues/211"
suggested_issue_title: "chore(web): adopción del lenguaje reglado NAVA en autenticación (issue maestro)"
branch: null
pr: null
pr_url: null
depends_on:
  - "PR #210 (issue #188) integrado o cerrado"
rules: []
decisions:
  - "DEC-077"
  - "DEC-079"
  - "DEC-080"
  - "DEC-081"
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/auth-eventos/README.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
created_at: "2026-09-03"
updated_at: "2026-09-03"
supersedes: "PROMPT-CHORE-LENGUAJE-REGLADO-NAVA-AUTH-v1"
superseded_by: null
---

# Adopción del lenguaje reglado NAVA en autenticación

## Issue maestro

[#211 · chore(web): adopción del lenguaje reglado NAVA en autenticación](https://github.com/bcaceres19/barberia/issues/211).

## Propósito

Este prompt **no modifica el repositorio**. Define orden, dependencias y criterio de avance entre dos entregas independientes, cada una con su issue, su rama, sus pruebas y su PR. No se usa para mezclar commits ni para saltarse una dependencia.

El destino de las dos fases son los 46 mockups por evento de [`ui-mockups-nava-tailored-grid-2026-09-03/auth-eventos`](../../evidence/ui-mockups-nava-tailored-grid-2026-09-03/auth-eventos/README.md), cuyo `README.md` es el contrato visual normativo.

Ambas fases son **exclusivamente visuales**. El ajuste funcional que estos mockups anticipan —configuración de canal de `DEC-081`, resumen de errores explicativo— se trata después, en sus propios issues, y está fuera del alcance de las dos.

## Fases

| Orden | Prompt | Preocupación | Bloquea a |
| --- | --- | --- | --- |
| 1 | [PROMPT-CHORE-CONTROLES-REGLADOS-SHARED-UI-v1](../chore/issue-pending-controles-reglados-shared-ui.md) · [#212](https://github.com/bcaceres19/barberia/issues/212) | Campo reglado, ranuras de código, alerta como nota al margen y botones en `apps/web/src/shared/ui`, con barrido de regresión en las once rutas P0 | Fase 2 |
| 2 | [PROMPT-CHORE-FIDELIDAD-ACCESO-RECUPERACION-v1](../chore/issue-pending-fidelidad-acceso-recuperacion.md) · [#213](https://github.com/bcaceres19/barberia/issues/213) | Composición de `/acceso` y `/recuperar-acceso` a fidelidad con los 23 eventos en los dos viewports | — |

## Por qué en este orden

La fase 2 consume los controles que crea la fase 1. Invertir el orden obliga a construir la composición sobre controles que van a cambiar debajo, y a rehacer la evidencia visual de los 46 eventos.

La fase 1 se aísla además porque su radio de impacto es mucho mayor que el de la fase 2: `BaseInput` y `BaseAlert` los consumen 17 archivos en 8 módulos, así que su riesgo real es la regresión en pantallas que nadie pidió rediseñar. Separarla permite verificar esa regresión sin mezclarla con el trabajo de fidelidad.

## Criterio de avance

La fase 2 **no comienza** hasta que la fase 1 esté integrada en `main` con:

1. CI en verde;
2. las once rutas P0 recorridas en navegador real sin regresión abierta;
3. `OtpInput` creado, probado y exportado desde `shared/ui`, aunque todavía no lo use ninguna pantalla;
4. su tabla `Criterio | Estado | Prueba o evidencia` entregada y completa.

Si la fase 1 termina con una desviación abierta que afecte a un control usado por autenticación, la fase 2 queda `blocked` hasta resolverla.

## Dependencia externa común

PR [#210](https://github.com/bcaceres19/barberia/pull/210) (issue [#188](https://github.com/bcaceres19/barberia/issues/188)) está abierto sobre `fix/188-acceso-recuperacion-fidelidad` y toca los mismos archivos de `apps/web/src/modules/auth`. Ninguna de las dos fases arranca antes de que ese PR esté integrado o cerrado.

## Trabajo posterior que estas fases no cubren

Se registra aquí solo para que no se cuele en el alcance de ninguna de las dos:

- **Configuración de canal del OTP (`DEC-081`).** Correo predeterminado, WhatsApp oficial o ambos, resueltos en servidor desde contactos verificados. Necesita issue funcional propio con contrato, proveedores, persistencia y pruebas de no enumeración. Los mockups `08`, `09`, `11` y `12` quedan representados pero no configurables hasta entonces.
- **Resumen de errores explicativo.** El contrato visual pide que el resumen global explique qué puede hacer la persona en vez de repetir literalmente cada error local; `LoginForm.vue` hoy los enumera. Es un cambio de texto y de comportamiento del componente, no de composición.

## Entrega

Este archivo no produce rama ni PR. Al cerrar cada fase se actualizan los metadatos del prompt hijo correspondiente y el índice [`docs/10-backlog/prompts/README.md`](../README.md).
