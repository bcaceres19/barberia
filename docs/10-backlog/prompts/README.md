---
titulo: "Catálogo de prompts persistentes"
version: "1.11"
estado: "Obligatorio para prompts reutilizables"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-23"
documentos_relacionados:
  - "../../../AGENTS.md"
  - "../../../CLAUDE.md"
  - "../../../CONTRIBUTING.md"
  - "../../00-control/registro-decisiones.md"
  - "../../02-requisitos/historias-usuario.md"
  - "../../03-desarrollo/flujo-git-github.md"
  - "../prompts-implementacion.md"
---

# Catálogo de prompts persistentes

## 1. Propósito

Este directorio es la ubicación canónica de todo prompt que deba sobrevivir a una conversación y ser ejecutado, revisado o integrado por otra persona, Codex, Claude u otro agente. Un prompt guardado conserva contexto, alcance y trazabilidad; un prompt que vive solo en el chat no forma parte del repositorio.

Los prompts son artefactos derivados. Nunca sustituyen las decisiones, reglas, historias, criterios ni estándares que enlazan. Ante una diferencia, manda la fuente normativa y el prompt se corrige antes de ejecutarse.

## 2. Qué debe guardarse

Se guarda cualquier prompt preparado para una ejecución posterior, entre otros:

- implementación o finalización de una `HU-*`;
- corrección de un defecto (`fix` o `hotfix`);
- auditoría de seguridad, datos, arquitectura u operación;
- revisión de una HU, issue, PR o migración;
- creación o reparación de pruebas;
- cambios de documentación, CI, dependencias u operación;
- orquestación de varias entregas independientes.

No se guardan respuestas explicativas ordinarias, comandos desechables ni preguntas de aclaración que no vayan a reutilizarse.

## 3. Organización y nombres

Los subdirectorios se crean al aparecer el primer artefacto de su tipo:

| Tipo | Ruta y nombre | Ejemplo |
| --- | --- | --- |
| Historia | `hu/hu-<numero>-<slug>.md` | `hu/hu-003-contrato-http.md` |
| Corrección | `fix/issue-<numero>-<slug>.md` | `fix/issue-148-evitar-cita-duplicada.md` |
| Auditoría | `audit/<fecha>-<slug>.md` o `audit/issue-<numero>-<slug>.md` | `audit/2026-08-12-seguridad-db.md` |
| Revisión de PR | `review/pr-<numero>-<slug>.md` | `review/pr-42-rls-catalogo.md` |
| Pruebas | `test/issue-<numero>-<slug>.md` | `test/issue-171-concurrencia.md` |
| Documentación | `docs/issue-<numero>-<slug>.md` | `docs/issue-35-persistencia-prompts.md` |
| CI/operación | `ci/issue-<numero>-<slug>.md` u `ops/issue-<numero>-<slug>.md` | `ci/issue-182-validar-openapi.md` |
| Orquestación | `orchestration/<slug>.md` | `orchestration/cierre-b0.md` |

El nombre no se cambia solo porque después se asigne issue o PR: esos enlaces viven en los metadatos y en este índice. Un cambio material del cuerpo después de iniciar la ejecución crea una nueva versión con sufijo `-v2`, `-v3`, etc., y enlaza `supersedes`.

## 4. Metadatos obligatorios

Cada archivo parte de [TEMPLATE.md](TEMPLATE.md) y conserva estos campos:

| Campo | Regla |
| --- | --- |
| `prompt_id` | Identificador estable, por ejemplo `PROMPT-HU-003-v1` |
| `version` | Versión del contenido del prompt |
| `kind` | `hu`, `fix`, `audit`, `review`, `test`, `docs`, `ci`, `ops`, `chore`, `spike` u `orchestration` |
| `status` | Uno de los estados de la sección 5 |
| `target_agents` | Agentes capaces de ejecutarlo: `claude`, `codex`, `human` o `any` |
| `repository` / `base_branch` | Destino explícito de la ejecución |
| `primary_hu` | `HU-*` primaria o `null` |
| `related_hu` | Otras HU relacionadas, sin mezclar su implementación |
| `issue` / `issue_url` | Número y enlace real, o `pending`/`null` mientras sea borrador |
| `suggested_issue_title` | Obligatorio cuando `issue: pending` |
| `branch` / `pr` / `pr_url` | Estado real de entrega; nunca valores inventados |
| `depends_on` | HU, issues, PR o decisiones que deben estar terminados |
| `rules`, `decisions`, `acceptance_criteria` | Identificadores normativos aplicables |
| `source_docs` | Archivos que el ejecutor debe releer |
| `created_at` / `updated_at` | Fecha ISO del artefacto |
| `supersedes` / `superseded_by` | Relación de versión cuando aplique |

No se escriben secretos, credenciales, tokens, datos personales reales, DSN ni material de sesión en metadatos, cuerpo, ejemplos o evidencia.

## 5. Estados y transición

| Estado | Significado |
| --- | --- |
| `draft` | Falta issue, decisión, dependencia o revisión; no se ejecuta |
| `ready` | Autocontenido, fuentes vigentes, dependencias satisfechas e issue real enlazado |
| `in_progress` | Existe rama o ejecución activa; se registran rama y PR cuando aparezcan |
| `executed` | Produjo el PR, informe o resultado solicitado; no implica que el PR esté integrado |
| `blocked` | Una dependencia o decisión impide continuar y está identificada |
| `superseded` | Otra versión lo reemplaza; se conserva por trazabilidad y se enlaza la sucesora |

Reglas de transición:

1. `issue: pending` implica siempre `status: draft` para cualquier prompt que pueda modificar el repositorio.
2. Antes de `ready`, se vuelve a leer cada fuente y se verifica que no haya duda o contradicción abierta.
3. Al iniciar, se actualizan `status`, `issue` y `branch` en la misma rama de trabajo.
4. Al abrir PR se registra su número/enlace y se usa `Closes #n` solo si cubre todo el issue; en caso contrario, `Refs #n`.
5. Al producir el resultado se usa `executed`; el índice distingue si el PR sigue abierto o ya fue integrado.
6. El cuerpo usado por una ejecución `in_progress` no se reescribe de forma material. Una corrección crea una versión nueva y marca la anterior `superseded`.

## 6. Contenido autocontenido

Otro agente debe poder abrir un único archivo y comenzar sin leer el chat que lo originó. Por ello todo prompt incluye:

1. objetivo y resultado observable;
2. preflight de Git, issue, rama y dependencias;
3. fuentes normativas que debe leer completamente;
4. alcance incluido y excluido;
5. estado existente que debe conservar o reutilizar;
6. instrucciones de implementación o revisión;
7. criterios de aceptación y pruebas;
8. documentación y trazabilidad a actualizar;
9. comandos de verificación;
10. forma de entrega, commit, PR y evidencia.

No se usan frases como “lo anterior”, “según lo que hablamos”, “estos tres cambios” o referencias a contenido visible únicamente en otro chat.

## 7. Una preocupación por prompt

Un prompt mutable atiende una preocupación primaria y un issue. Si un trabajo necesita varias HU o fixes:

1. se crea un archivo por HU o issue;
2. cada uno conserva rama, pruebas y PR propios;
3. opcionalmente se crea un prompt de `orchestration/` que solo define orden, dependencias y criterio de avance;
4. nunca se usa la orquestación para mezclar commits o saltarse las dependencias.

## 8. Relación con documentos anteriores

[`../prompts-implementacion.md`](../prompts-implementacion.md), [`../prompts-detallados-b0.md`](../prompts-detallados-b0.md) y los prompts históricos del directorio padre se conservan como antecedentes. Desde esta versión, todo prompt nuevo o revisado vive aquí de manera individual. Cuando exista un archivo individual para la misma preocupación, su versión y estado en este catálogo identifican el prompt operativo; las fuentes normativas siguen teniendo precedencia sobre ambos.

## 9. Índice actual

| Prompt | Tipo | HU | Issue | Estado | Dependencias | Rama / PR |
| --- | --- | --- | --- | --- | --- | --- |
| [PROMPT-HU-003-v1](hu/hu-003-contrato-http.md) | `hu` | `HU-003` | [#37](https://github.com/bcaceres19/barberia/issues/37) | `executed` | `HU-002` terminada | `feat/37-hu003-contrato-http` / [#38](https://github.com/bcaceres19/barberia/pull/38) |
| [PROMPT-HU-004-v1](hu/hu-004-idempotencia.md) | `hu` | `HU-004` | [#40](https://github.com/bcaceres19/barberia/issues/40) | `executed` | `HU-003` integrada | `feat/40-hu004-idempotencia` / [#41](https://github.com/bcaceres19/barberia/pull/41) |
| [PROMPT-HU-009-v1](hu/hu-009-sistema-visual.md) | `hu` | `HU-009` | [#42](https://github.com/bcaceres19/barberia/issues/42) | `executed` | Secuencia tras `HU-004` | `feat/42-hu009-sistema-visual` / [#43](https://github.com/bcaceres19/barberia/pull/43) |
| [PROMPT-HU-005-v1](hu/hu-005-inicio-sesion.md) (v1.2) | `hu` | `HU-005` | [#44](https://github.com/bcaceres19/barberia/issues/44) | `executed` | `HU-002`/`HU-003` integradas; `CT-003` resuelta (`DEC-055`) | [PR #49](https://github.com/bcaceres19/barberia/pull/49), integrada en `main`; `DP-SEG-08` cerrada por `DEC-058`/`HU-006` |
| [PROMPT-HU-006-v1](hu/hu-006-sesion-persistente.md) (v1.2) | `hu` | `HU-006` | [#45](https://github.com/bcaceres19/barberia/issues/45) | `executed` | `HU-005` integrada; `CT-003` resuelta (`DEC-055`); `DP-SEG-08` resuelta (`DEC-058`) | [PR #50](https://github.com/bcaceres19/barberia/pull/50), integrada en `main` |
| [PROMPT-HU-010-v1](hu/hu-010-pantalla-acceso.md) (v1.1) | `hu` | `HU-010` | [#46](https://github.com/bcaceres19/barberia/issues/46) | `executed` | `HU-005`/`HU-009` integradas; `CT-004` resuelta (`DEC-056`) | [PR #52](https://github.com/bcaceres19/barberia/pull/52), integrada en `main`; `DP-UX-06` cerrada por `DEC-059` |
| [PROMPT-HU-012-v1](hu/hu-012-cascaron-panel-privado.md) (v1.2) | `hu` | `HU-012` | [#56](https://github.com/bcaceres19/barberia/issues/56) | `executed` | `HU-006`/`HU-009`/`HU-010` integradas; `DP-SEG-09` resuelta (`DEC-060`) | [PR #59](https://github.com/bcaceres19/barberia/pull/59), integrada en `main` |
| [PROMPT-HU-007-v1](hu/hu-007-defensa-abuso.md) (v1.3) | `hu` | `HU-007` | [#57](https://github.com/bcaceres19/barberia/issues/57) | `executed` | `DEC-061`/`DEC-062` resuelven `CT-005`/`DP-SEG-10`; `HU-012` integrada en `main` | [PR #61](https://github.com/bcaceres19/barberia/pull/61), integrada en `main` |
| [PROMPT-HU-008-v1](hu/hu-008-recuperacion-acceso.md) (v1.4) | `hu` | `HU-008` | [#58](https://github.com/bcaceres19/barberia/issues/58) | `executed` | `DEC-063`–`DEC-066` resuelven `DP-SEG-11`/`DP-SEG-12`/`CT-006`/`DP-NOT-05`; `HU-007` integrada en `main` | [PR #63](https://github.com/bcaceres19/barberia/pull/63), integrada en `main` |
| [PROMPT-HU-011-v1](hu/hu-011-pantalla-recuperacion.md) | `hu` | `HU-011` | [#64](https://github.com/bcaceres19/barberia/issues/64) | `executed` | `HU-008`/`HU-009`/`HU-010` integradas en `main` | `feat/64-hu011-pantalla-recuperacion` / [#66](https://github.com/bcaceres19/barberia/pull/66) |
| [PROMPT-HU-020-v1](hu/hu-020-configuracion-barberia.md) | `hu` | `HU-020` | [#67](https://github.com/bcaceres19/barberia/issues/67) | `executed` | Criterio de salida de B0 verificado 2026-08-23 (CI verde sobre `main` `a57354b`); PR abierto, pendiente de CI/merge | `feat/67-hu020-configuracion-barberia` / PR pendiente de abrir |
| [PROMPT-HU-021-v1](hu/hu-021-registro-listado-barberos.md) | `hu` | `HU-021` | [#68](https://github.com/bcaceres19/barberia/issues/68) | `draft` | `HU-020` (#67) debe integrarse en `main` | Rama / PR pendientes |

## 10. Lista de control al guardar o entregar

- [ ] El archivo está basado en `TEMPLATE.md` y no depende del chat.
- [ ] Tipo, HU, issue, estado, dependencias y fuentes son reales.
- [ ] Si cambia el repositorio, hay issue antes de pasar a `ready`.
- [ ] Alcance excluido y criterio de terminado son explícitos.
- [ ] Pruebas y documentación afectada están enumeradas.
- [ ] El índice de esta página está actualizado.
- [ ] La respuesta al usuario enlaza el archivo persistido.
- [ ] No contiene secretos, tokens ni datos personales reales.
