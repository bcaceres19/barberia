---
titulo: "Catálogo de prompts persistentes"
version: "1.42"
estado: "Obligatorio para prompts reutilizables"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-09-01"
documentos_relacionados:
  - "../../../AGENTS.md"
  - "../../../CLAUDE.md"
  - "../../../CONTRIBUTING.md"
  - "../../00-control/registro-decisiones.md"
  - "../../02-requisitos/historias-usuario.md"
  - "../../03-desarrollo/flujo-git-github.md"
  - "../../03-desarrollo/especificacion-frontend-nava.md"
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
| [PROMPT-ORCH-NAVA-FRONTEND-v1](orchestration/adopcion-frontend-nava.md) | `orchestration` | `HU-005`–`HU-012`, `HU-020`–`HU-024`, `HU-040`–`HU-042`, `HU-060`–`HU-065`; B3–B4 futuras solo con HU real | `pending` | `draft` | `DEC-077`; exige un issue y prompt independiente por fundación o pantalla | Sin rama/PR; mapa para Claude/Codex, no autoriza una migración masiva |
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
| [PROMPT-HU-020-v1](hu/hu-020-configuracion-barberia.md) | `hu` | `HU-020` | [#67](https://github.com/bcaceres19/barberia/issues/67) | `executed` | Criterio de salida de B0 verificado 2026-08-23 (CI verde sobre `main` `a57354b`); integrada en `main` (squash-merge 2026-08-23) | `feat/67-hu020-configuracion-barberia` / [PR #70](https://github.com/bcaceres19/barberia/pull/70), integrada en `main` |
| [PROMPT-HU-021-v1](hu/hu-021-registro-listado-barberos.md) | `hu` | `HU-021` | [#68](https://github.com/bcaceres19/barberia/issues/68) | `executed` | `HU-020` (#67) integrada en `main` 2026-08-23 | Rama `feat/68-hu021-barberos`; [PR #72](https://github.com/bcaceres19/barberia/pull/72), integrada en `main` 2026-08-24 con CI verde |
| [PROMPT-HU-022-v1](hu/hu-022-catalogo-servicios.md) | `hu` | `HU-022` | [#75](https://github.com/bcaceres19/barberia/issues/75) | `executed` | `HU-020`/`HU-021` integradas; `DP-SER-01` resuelta (`DEC-067`) | Rama `feat/75-hu022-catalogo-servicios`; [PR #79](https://github.com/bcaceres19/barberia/pull/79), integrada en `main` con CI verde |
| [PROMPT-HU-023-v1](hu/hu-023-asignacion-servicios-barberos.md) | `hu` | `HU-023` | [#76](https://github.com/bcaceres19/barberia/issues/76) | `executed` | `HU-022` integrada ([PR #79](https://github.com/bcaceres19/barberia/pull/79)); `DP-SER-02` resuelta (`DEC-068`) | Rama `feat/76-hu023-servicios-barberos`; [PR #83](https://github.com/bcaceres19/barberia/pull/83), integrada en `main` |
| [PROMPT-HU-024-v1](hu/hu-024-ciclo-vida-servicios.md) | `hu` | `HU-024` | [#77](https://github.com/bcaceres19/barberia/issues/77) | `executed` | `HU-022`/`HU-023` integradas; `DP-SER-03` resuelta (`DEC-069`) | Rama `feat/77-hu024-ciclo-vida-servicios`; [PR #84](https://github.com/bcaceres19/barberia/pull/84), integrada en `main` |
| [PROMPT-HU-040-v1](hu/hu-040-horario-laboral.md) (v1.3) | `hu` | `HU-040` | [#90](https://github.com/bcaceres19/barberia/issues/90), abierto por `CA-040-08` parcial | `executed` | B1 cerrado; `CT-008` resuelta (`DEC-070`) | Rama `feat/90-hu040-horario-laboral`; [PR #93](https://github.com/bcaceres19/barberia/pull/93), integrada en `main` con CI verde |
| [PROMPT-HU-041-v1](hu/hu-041-excepciones-festivos.md) (v1.3) | `hu` | `HU-041` | [#95](https://github.com/bcaceres19/barberia/issues/95), abierto por `CA-041-08` parcial | `executed` | `HU-040` integrada ([PR #93](https://github.com/bcaceres19/barberia/pull/93)); `CT-008` resuelta (`DEC-070`) | Rama `feat/95-hu041-excepciones-festivos`; [PR #96](https://github.com/bcaceres19/barberia/pull/96), integrada en `main` con CI verde |
| [PROMPT-HU-042-v1](hu/hu-042-bloqueos-agenda.md) | `hu` | `HU-042` | [#98](https://github.com/bcaceres19/barberia/issues/98), abierto por seguimiento parcial (issue [#100](https://github.com/bcaceres19/barberia/issues/100)) | `executed` | `HU-040`/`HU-041` integradas; `CT-008` resuelta (`DEC-070`); backend completo y probado contra PostgreSQL real | Rama `feat/98-hu042-bloqueos-agenda`; [PR #99](https://github.com/bcaceres19/barberia/pull/99), integrada en `main` con CI verde |
| [PROMPT-HU-060-v1](hu/hu-060-nucleo-citas.md) | `hu` | `HU-060` | [#104](https://github.com/bcaceres19/barberia/issues/104) | `executed` | B2 integrado; seguimientos `#90`/`#95`/`#98`/`#100` verificados sin impacto en el contrato interno de B3 | Rama `feat/104-hu060-nucleo-citas`; integrada mediante [PR #105](https://github.com/bcaceres19/barberia/pull/105) |
| [PROMPT-HU-061-v1](hu/hu-061-creacion-manual-citas.md) | `hu` | `HU-061` | [#107](https://github.com/bcaceres19/barberia/issues/107) (abierto: E2E/evidencia responsiva pendientes) | `executed` | `HU-060` integrada (PR #105); `DP-CIT-01`–`DP-CIT-03` resueltas (`DEC-071`–`DEC-073`) | Integrada mediante [PR #109](https://github.com/bcaceres19/barberia/pull/109) |
| [PROMPT-HU-062-v1](hu/hu-062-agenda-diaria.md) | `hu` | `HU-062` | [#111](https://github.com/bcaceres19/barberia/issues/111) | `executed` | Todas cumplidas | `feat/111-hu062-agenda-diaria`, integrada mediante [PR #113](https://github.com/bcaceres19/barberia/pull/113); E2E/evidencia responsiva pendientes (issue #111 permanece abierto) |
| [PROMPT-HU-063-v1](hu/hu-063-navegacion-agenda.md) (v1.3) | `hu` | `HU-063` | [#116](https://github.com/bcaceres19/barberia/issues/116) | `executed` | Todas cumplidas | `feat/116-hu063-navegacion-agenda`, integrada mediante [PR #118](https://github.com/bcaceres19/barberia/pull/118); E2E/evidencia responsiva pendientes (issue #116 permanece abierto) |
| [PROMPT-HU-064-v1](hu/hu-064-detalle-historial-turno.md) (v1.3) | `hu` | `HU-064` | [#120](https://github.com/bcaceres19/barberia/issues/120) | `executed` | `CA-064-01`–`CA-064-07` cumplidas; `CA-064-08` (E2E/evidencia responsiva) pendiente | `feat/120-hu064-detalle-historial-turno`, integrada mediante [PR #121](https://github.com/bcaceres19/barberia/pull/121); E2E/evidencia responsiva pendientes (issue #120 permanece abierto) |
| [PROMPT-HU-065-v1](hu/hu-065-reprogramacion-turno.md) (v1.3) | `hu` | `HU-065` | [#123](https://github.com/bcaceres19/barberia/issues/123) | `executed` | `CA-065-01`–`CA-065-07` cumplidas; `CA-065-08` (E2E/evidencia responsiva) pendiente | `feat/123-hu065-reprogramacion-turno`, integrada mediante [PR #125](https://github.com/bcaceres19/barberia/pull/125); E2E/evidencia responsiva pendientes (issue #123 permanece abierto) |
| [PROMPT-DOCS-INFORME-ACADEMICO-v1](docs/pending-informe-academico-secciones-1-a-4-2.md) | `docs` | — | [#81](https://github.com/bcaceres19/barberia/issues/81) | `superseded` | Sustituido por v2 tras retroalimentación sobre lenguaje y audiencia | Rama `docs/81-informe-academico`; [PR #82](https://github.com/bcaceres19/barberia/pull/82) |
| [PROMPT-DOCS-INFORME-ACADEMICO-v2](docs/informe-academico-secciones-1-a-4-2-v2-gerencial.md) | `docs` | — | [#81](https://github.com/bcaceres19/barberia/issues/81) | `superseded` | Sustituido por v3 para incorporar identidad institucional, equipo confirmado y sistema visual sobrio | Rama `docs/81-informe-academico`; [PR #82](https://github.com/bcaceres19/barberia/pull/82); SHA-256 `6cad6bcc…78bbe0` |
| [PROMPT-DOCS-INFORME-ACADEMICO-v3](docs/informe-academico-secciones-1-a-4-2-v3-identidad-institucional.md) | `docs` | — | [#81](https://github.com/bcaceres19/barberia/issues/81) | `executed` | Identidad institucional, responsabilidades compartidas, primer participante, Arial y color solo para estados | 26 páginas y 45 RU; QA visual completo; SHA-256 `73554078…72859`; rama `docs/81-informe-academico`; [PR #82](https://github.com/bcaceres19/barberia/pull/82) |
| [PROMPT-TEST-QA-B0-B1-CHROME-v1](test/2026-08-25-qa-manual-b0-b1-chrome-mcp.md) | `test` | `HU-005`–`HU-012`, `HU-020`–`HU-024` | — (no modifica el repositorio) | `executed` | B0 y B1 recorridas; informe persistido con pasos, observaciones, evidencias y pendientes | Sin rama/PR: QA manual en navegador real vía MCP de Chrome; [informe](test/2026-08-25-qa-manual-b0-b1-chrome-mcp-report.md) |
| [PROMPT-TEST-OTP-EMAIL-RESEND-v1](test/issue-86-otp-correo-resend.md) | `test` | `HU-008`, `HU-011` | [#86](https://github.com/bcaceres19/barberia/issues/86) | `ready` | HU-008/HU-011 integradas; credenciales y buzón Resend se aportan de forma segura al ejecutar | Persistido en `docs/86-otp-correo-prompt` / [PR #129](https://github.com/bcaceres19/barberia/pull/129); ejecución futura en `test/86-otp-correo-resend` |
| [PROMPT-TEST-QA-B2-B3-CHROME-v1](test/2026-09-01-qa-manual-b2-b3-chrome-mcp.md) | `test` | `HU-040`–`HU-042`, `HU-060`–`HU-065` | — (no modifica el repositorio) | `ready` | Todas las nueve historias integradas en `main`; issues de seguimiento reales (`#90`,`#95`,`#98`,`#100`,`#107`,`#111`,`#116`,`#120`,`#123`) | Sin rama/PR: QA manual en navegador real vía MCP de Chrome, reproduce a mano los `.spec.ts` ya escritos de siete de las nueve historias; pendiente de ejecución |

## 10. Lista de control al guardar o entregar

- [ ] El archivo está basado en `TEMPLATE.md` y no depende del chat.
- [ ] Tipo, HU, issue, estado, dependencias y fuentes son reales.
- [ ] Si cambia el repositorio, hay issue antes de pasar a `ready`.
- [ ] Alcance excluido y criterio de terminado son explícitos.
- [ ] Pruebas y documentación afectada están enumeradas.
- [ ] El índice de esta página está actualizado.
- [ ] La respuesta al usuario enlaza el archivo persistido.
- [ ] No contiene secretos, tokens ni datos personales reales.
