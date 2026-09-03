---
prompt_id: "PROMPT-ORCH-NAVA-ATLAS-v3"
version: "3.0"
kind: "orchestration"
status: "ready"
target_agents:
  - "claude"
  - "codex"
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
issue: 184
issue_url: "https://github.com/bcaceres19/barberia/issues/184"
suggested_issue_title: "chore(web): ejecutar el rediseño integral según el atlas NAVA Tailored Grid"
branch: null
pr: null
pr_url: null
depends_on:
  - "El atlas del issue #183 está integrado en `main`"
  - "El protocolo de fidelidad visual del issue #208 está integrado en `main`"
  - "Cada familia de pantalla se ejecuta mediante su propio issue, rama y PR"
rules:
  - "RN-TEN-01"
  - "RN-DAT-01"
  - "RN-DAT-02"
  - "RN-IDE-01"
  - "RN-CIT-02"
  - "RN-CIT-04"
  - "RN-CIT-05"
  - "RN-DIS-02"
  - "RN-DIS-04"
  - "RN-DIS-05"
  - "RN-DIS-06"
  - "RN-BLQ-01"
  - "RN-BLQ-02"
  - "RN-BLQ-03"
  - "RN-BLQ-04"
decisions:
  - "DEC-016"
  - "DEC-019"
  - "DEC-020"
  - "DEC-074"
  - "DEC-075"
  - "DEC-077"
  - "DEC-078"
  - "DEC-079"
  - "DEC-080"
acceptance_criteria:
  - "CA-ORCH-ATLAS3-01: las once rutas P0 implementadas quedan asignadas a un panel exacto del atlas y a una entrega visual trazable."
  - "CA-ORCH-ATLAS3-02: cada pantalla aplica modo de fidelidad al mockup; colores, proporciones, jerarquía, escalas, centrado, alineaciones, densidad y estados coinciden en el viewport representado dentro de las tolerancias normativas."
  - "CA-ORCH-ATLAS3-03: ninguna fase se cierra con auditoría, cambio de tokens, recaptura o métricas DOM sin un rediseño observable y comparación visual."
  - "CA-ORCH-ATLAS3-04: cada ruta conserva comportamiento, copy normativo, API, seguridad, privacidad, permisos y reglas existentes; un mockup no crea funciones."
  - "CA-ORCH-ATLAS3-05: cada familia aporta baseline, captura final, lado a lado y overlay/diff vistos por el implementador, con viewport efectivo comprobado."
  - "CA-ORCH-ATLAS3-06: cada pantalla cubre 320, 360, 768 y 1280 px, zoom 200 %, teclado, foco, movimiento reducido y estados aplicables."
  - "CA-ORCH-ATLAS3-07: cada preocupación usa issue, prompt hijo cuando corresponda, rama y PR propios; no se hace una migración masiva en una sola rama."
  - "CA-ORCH-ATLAS3-08: formato, lint, tipos, pruebas de componente, axe, E2E afectadas y build pasan antes de integrar."
  - "CA-ORCH-ATLAS3-09: Chrome DevTools se usa sobre la app real con la cuenta QA indicada por el propietario; consola y red quedan sin regresiones del rediseño."
  - "CA-ORCH-ATLAS3-10: el informe final marca cada ruta PASS o FAIL y no declara rediseño completo si existe una diferencia primaria no explicada."
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - ".claude/skills/nava-mockup-fidelity/SKILL.md"
  - ".claude/skills/nava-mockup-fidelity/references/acceptance-gates.md"
  - ".claude/skills/browser-viewport-verification/SKILL.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/estados-citas.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/04-arquitectura/frontend.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/README.md"
  - "docs/10-backlog/evidence/ui-redesign-nava-2026-09-02/README.md"
  - "docs/10-backlog/prompts/README.md"
  - "apps/web/package.json"
  - "apps/web/src/app/router/index.ts"
created_at: "2026-09-03"
updated_at: "2026-09-03"
supersedes: "PROMPT-ORCH-NAVA-ATLAS-v2"
superseded_by: null
---

# Rediseño integral NAVA con fidelidad verificable al atlas

## Instrucción para Claude Code

Ejecuta este archivo completo de principio a fin como orquestador, Senior Product Designer, UI/UX Designer e implementador frontend. El objetivo no es revisar si la app “parece NAVA”: debes **rediseñar materialmente todas las pantallas P0 implementadas y hacer que cada una coincida con su panel asignado del atlas** en color, proporción, jerarquía tipográfica, escala de controles e iconos, centrado, alineación, densidad y estados.

Este trabajo es netamente visual. Conserva funciones, copy normativo, rutas, contratos, permisos, seguridad, privacidad y reglas de negocio. Una imagen nunca autoriza datos, botones, módulos, estados o endpoints inexistentes.

La orden “haz todo” significa completar todas las familias de forma persistente y secuencial, no mezclar toda la app en una sola rama. Crea o usa un issue hijo, prompt persistente, rama y PR pequeño por preocupación; espera checks, integra según las reglas del repositorio, actualiza `main` y continúa con la siguiente familia.

## Modelo y sesión

Para una sesión nueva de Claude Code usa:

```text
claude --model claude-sonnet-5 --effort high
```

Comprueba con `claude --version` que Claude Code sea 2.1.197 o posterior y con `/status` que el modelo efectivo sea `claude-sonnet-5` y el esfuerzo sea `high`. No escribas `hight`. No cambies silenciosamente a otro modelo o esfuerzo; si el proveedor o la cuenta no ofrecen esa combinación, informa el bloqueo exacto antes de modificar código.

La elección de modelo no sustituye ninguna puerta de aceptación de este prompt.

## Reglas de persistencia

Trabaja con autonomía hasta completar el resultado autorizado. Puedes gestionar issues hijos, prompts, ramas, commits, pushes, PRs y checks, pero debes respetar `AGENTS.md`, `CONTRIBUTING.md` y las protecciones de `main`. No uses force-push, no reescribas `main`, no borres cambios ajenos y no combines preocupaciones.

Solo detente por:

1. contradicción normativa que cambie producto, contrato o seguridad;
2. credencial, autorización o aprobación humana realmente obligatoria;
3. infraestructura externa fallando de forma reproducible;
4. cambios ajenos que no puedan aislarse con seguridad;
5. modelo solicitado no disponible sin posibilidad autorizada de sustitución.

Una elección rutinaria de CSS, layout o componente no es un bloqueo.

## Cuenta QA y privacidad

- Usa exclusivamente la cuenta QA que el propietario indique en la conversación de ejecución.
- Si falta URL o cuenta, pide una vez los datos mínimos al llegar a la inspección autenticada; avanza mientras tanto en trabajo estático que no dependa de ellos.
- Introduce credenciales solo en la interfaz del navegador. No las guardes en archivos, comandos, variables versionadas, logs, prompts, screenshots ni reportes.
- No uses datos de producción ni incluyas información personal real en evidencias.
- No evadas login, OTP, rate limits, permisos o sesión.

## Autoridad

Aplica esta precedencia:

1. `AGENTS.md`, seguridad, privacidad y Git;
2. decisiones, reglas, HU y criterios de aceptación;
3. OpenAPI, backend real y tipos generados;
4. estándar visual y especificación frontend NAVA;
5. panel exacto del atlas asignado a la ruta;
6. este prompt y sus prompts hijos.

Si la imagen contradice una autoridad superior, conserva la autoridad y documenta la desviación visual. No uses esa excepción para cambiar una preferencia visual que sí puede reproducirse.

## Skill obligatoria

Para cada familia, antes de editar y otra vez antes de declarar terminada:

1. carga y sigue `.claude/skills/nava-mockup-fidelity/SKILL.md`;
2. completa `.claude/skills/nava-mockup-fidelity/references/acceptance-gates.md`;
3. usa `.claude/skills/browser-viewport-verification/SKILL.md` para demostrar el viewport efectivo.

Todas las rutas de este prompt usan **modo de fidelidad al mockup**, no “identidad guiada”. La libertad creativa solo cubre técnica de implementación, reflow de anchos no representados y aspectos ausentes de la imagen.

## Preflight obligatorio

1. Lee completamente todos los `source_docs`; no te limites a este resumen.
2. Verifica rama, remoto, HEAD y árbol de trabajo. Aísla cambios ajenos sin descartarlos.
3. Confirma que #183 y #208 estén integrados en `main` y que #184 siga abierto.
4. Actualiza `main` por fast-forward y crea cada rama desde ese punto.
5. Ejecuta Graphify si `graphify-out/graph.json` existe.
6. Inspecciona router, componentes, estilos, pruebas y E2E reales; no confíes en un inventario viejo.
7. Ejecuta baseline de formato, lint, tipos, unitarias y build; separa fallos previos de regresiones.
8. Arranca el stack con los scripts oficiales y abre la app real con Chrome DevTools.
9. Inicia sesión por la UI con la cuenta QA indicada.
10. Captura antes de editar cada ruta en 320, 360, 768 y 1280 px. Verifica el viewport real; si el resize falla, usa el harness Playwright descrito por la skill.
11. Crea la matriz `ruta → panel exacto → estado → cambio obligatorio → evidencia antes/final/comparación`.

## Contrato cromático

Usa los valores declarados por `DEC-079` y el estándar:

- tinta `#101B2B`;
- marfil `#F4F0E7`;
- blanco `#FFFFFF`;
- grafito `#2A2D32` y `#5E625F`;
- piedra `#E8E2D8` y borde `#C9C0B2`;
- salvia `#748477`;
- latón `#B8955A` y latón oscuro `#765C2F`;
- paleta semántica del estándar para éxito, advertencia/conflicto, error, información e inactivo.

Los colores declarados deben coincidir en CSS computado. No “aproximes” el azul, marfil o latón. El color no comunica por sí solo.

## Atlas y mapa vinculante

Abre a resolución original el [README del atlas](../../evidence/ui-mockups-nava-tailored-grid-2026-09-02/README.md) y cada PNG aplicable. Identifica el panel exacto, no toda la lámina de forma ambigua.

| Orden | Referencia | Rutas o superficie |
| --- | --- | --- |
| 0 | `00-referencia-tailored-grid.png` | Lenguaje maestro, carriles, contraste y proporciones generales; no crea módulos. |
| 1 | `14-componentes-formularios-alertas.png` | Fundaciones, inputs, botones, foco, alertas, modal y estados compartidos. |
| 2 | `01-agenda-diaria-responsive.png` | `/panel`. |
| 3 | `02-acceso-recuperacion.png` | `/acceso`, `/recuperar-acceso` y estados de ambos flujos. |
| 4 | `03-nuevo-turno.png` | `/panel/turnos/nuevo`. |
| 5 | `04-detalle-historial-reprogramacion.png` | `/panel/turnos/:appointmentId`. |
| 6 | `05-servicios-y-asignacion.png` | `/panel/servicios`, `/panel/servicios-por-barbero`. |
| 7 | `06-barberos-y-barberia.png` | `/panel/barberos`, `/panel/barberia`. |
| 8 | `07-horarios-y-excepciones.png` | `/panel/horarios`. |
| 9 | `08-bloqueos-puntuales-y-series.png` | `/panel/bloqueos`. |

`09-reserva-publica-pasos-1-a-3.png` a `13-configuracion-operativa-comunicaciones.png` son reserva conceptual P0. No implementes esas funciones salvo que una HU, contrato, ruta e issue posteriores las autoricen expresamente.

## Puerta visual por pantalla

Antes de tocar código, completa la tabla de contrato de la skill con mediciones de la imagen original y de la app actual. Como mínimo mide:

- panel y canvas;
- bloque de marca;
- heading y copy;
- ancho, alto y posición del contenido/formulario;
- controles e iconos;
- acción primaria;
- alertas, diálogos y estados aplicables.

Implementa la región completa. No cierres una fase por:

- cambiar solo tokens o fondos;
- agrandar únicamente el wordmark;
- corregir un único centrado;
- actualizar screenshots o snapshots;
- demostrar que las clases o variables existen;
- afirmar que la pantalla ya pertenece a NAVA.

Después, renderiza la app en el viewport exacto del panel y compara:

1. referencia original;
2. app anterior;
3. app final;
4. lado a lado;
5. overlay o diff.

Aplica las tolerancias de `estandar-diseno-visual.md` y de la skill. Corrige cualquier diferencia primaria inexplicada aunque una cifra aislada parezca aceptable. La comparación debe ser vista por el implementador; generar el archivo sin abrirlo no satisface la puerta.

## Verificación con Chrome DevTools

Para cada ruta:

1. recorre el flujo real con clic y teclado;
2. comprueba 320, 360, 768 y 1280 px, además de zoom 200 %;
3. verifica `window.innerWidth` y `window.innerHeight` después de cada resize;
4. inspecciona nombres accesibles, headings, labels, orden y restauración de foco;
5. revisa estilos computados de colores, fuentes, tamaños, overflow y objetivos táctiles;
6. revisa consola y Network; no dejes errores nuevos ni solicitudes inesperadas;
7. recorre carga, vacío, error, conflicto, éxito y disabled cuando existan;
8. guarda evidencia determinista sin credenciales ni datos personales.

Playwright y Vitest complementan la revisión. No sustituyen la inspección manual de la app real. Si Chrome no puede producir un viewport exacto, usa Playwright para esas capturas y conserva Chrome para interacción, accesibilidad, consola, red y estilos.

## Frontera funcional estricta

Puedes cambiar composición, HTML semántico, CSS, componentes visuales, iconos, responsive, estados visibles ya existentes, pruebas y evidencia. No cambies backend, OpenAPI, migraciones, payloads, permisos, reglas, persistencia o capacidades.

No agregues ingresos, caja, reportes, inventario, CRM, pagos, marketplace, cuenta del cliente, multi-barbero P0, estado `in_progress`, notificaciones ficticias, avatar Admin ni fotografía obligatoria. Si encuentras un bug funcional independiente, regístralo en otro issue y continúa donde no bloquee.

CSS, scoped CSS, CSS Modules, utility-first, Tailwind o bibliotecas están permitidos. No instales Tailwind automáticamente. Toda dependencia nueva documenta licencia, mantenimiento, accesibilidad, integración, impacto de bundle y razón por la cual lo existente no basta.

## Secuencia de entregas

Ejecuta, como mínimo, estas preocupaciones por separado:

1. fundaciones compartidas de controles, formularios, alertas y diálogos;
2. shell y navegación;
3. acceso y recuperación;
4. agenda diaria;
5. nuevo turno;
6. detalle, historial y reprogramación;
7. barberos;
8. configuración de barbería;
9. servicios y ciclo de vida;
10. servicios por barbero;
11. horarios y excepciones;
12. bloqueos;
13. revisión integral.

Puedes unir dos rutas solo si comparten la misma preocupación, componentes y pruebas y el PR sigue siendo pequeño. Documenta la razón. Después de cada merge actualiza `main` antes de crear la siguiente rama.

## Pruebas por PR

Ejecuta los comandos reales disponibles en el repositorio, incluyendo sus equivalentes para:

```text
pnpm --filter @system-barbershop/web format
pnpm --filter @system-barbershop/web lint
pnpm --filter @system-barbershop/web typecheck
pnpm --filter @system-barbershop/web test:unit
pnpm --filter @system-barbershop/web build
pnpm --filter @system-barbershop/web test:e2e
git diff --check
```

Incluye prueba de componente para semántica, interacción, foco y estados; `vitest-axe` en estados relevantes; E2E del recorrido afectado; evidencia de los cuatro anchos; y revisión real de la comparación. No reduzcas umbrales, ignores pruebas inestables ni actualices snapshots sin mirar el cambio.

## Formato de progreso

Después de cada familia informa:

| Campo | Contenido obligatorio |
| --- | --- |
| Familia | rutas y estado implementado |
| Issue / rama / PR | valores y enlaces reales |
| Mockup | archivo, panel, viewport y estado exactos |
| Cambios | geometría, tipografía, colores, controles, iconos y estados |
| Evidencia | baseline, final, lado a lado, overlay/diff y viewports medidos |
| Diferencias | desviaciones restantes, autoridad y tratamiento |
| Pruebas | comandos y resultados reales |
| Fidelidad | `PASS` o `FAIL` |
| Siguiente | siguiente familia desbloqueada |

No pidas confirmación entre fases para decisiones rutinarias ya cubiertas. No uses `PASS` mientras exista una diferencia primaria sin explicar.

## Entrega final

Entrega una matriz de las once rutas con:

1. panel exacto y modo de fidelidad;
2. issue, rama y PR;
3. evidencia anterior, final, lado a lado y overlay/diff;
4. resultado en 320, 360, 768, 1280 y zoom 200 %;
5. teclado, foco, accesibilidad, consola, red y estados recorridos;
6. pruebas y build;
7. diferencias justificadas;
8. dependencias visuales añadidas o confirmación de ninguna;
9. bugs funcionales registrados aparte;
10. `PASS` o `FAIL` por ruta y del conjunto.

No declares “rediseño completo” si falta una ruta, un estado aplicable, evidencia comparable o una diferencia visible primaria sigue sin explicación.

## Texto corto para iniciar

```text
Ejecuta de principio a fin `docs/10-backlog/prompts/orchestration/redisenio-integral-nava-tailored-grid-v3.md` con Claude Sonnet 5 y esfuerzo high. Carga obligatoriamente la skill `nava-mockup-fidelity` en cada familia. Esto no es una inspiración ni una auditoría: rediseña las once rutas P0 en modo de fidelidad al panel exacto del atlas, conserva toda la funcionalidad y verifica la app real con Chrome DevTools, viewports efectivos y comparación antes/final/lado a lado/overlay. No declares una pantalla terminada con diferencias primarias sin explicar; gestiona issues, prompts hijos, ramas, PRs y checks hasta completar todas las fases.
```
