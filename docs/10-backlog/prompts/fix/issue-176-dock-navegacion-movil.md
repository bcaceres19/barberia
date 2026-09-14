---
prompt_id: "PROMPT-FIX-176-DOCK-NAVEGACION-MOVIL-v1"
version: "1.0"
kind: "fix"
status: "executed"
target_agents:
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-012"
related_hu:
  - "HU-009"
issue: "176"
issue_url: "https://github.com/bcaceres19/barberia/issues/176"
suggested_issue_title: "fix(web): el dock de navegacion trunca etiquetas a 320/360 px"
branch: "fix/176-dock-navegacion-movil"
pr: "258"
pr_url: "https://github.com/bcaceres19/barberia/pull/258"
depends_on:
  - "HU-012 integrada en main mediante PR #59"
  - "PR #202 (issue #187) integrada en main; posible correccion previa que debe verificarse"
rules:
  - "RN-TEN-01"
decisions:
  - "DEC-033"
  - "DEC-056"
  - "DEC-060"
  - "DEC-077"
  - "DEC-078"
  - "DEC-079"
acceptance_criteria:
  - "CA-012-08"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/10-backlog/prompts/chore/issue-187-shell-navegacion-tailored-grid.md"
created_at: "2026-09-13"
updated_at: "2026-09-13"
supersedes: null
superseded_by: null
---

# Verificar y cerrar el truncamiento del dock móvil de HU-012

## Perfil de ejecución requerido

- Modelo: `gpt-5.6-terra`.
- Intensidad de razonamiento: `high`.
- Ejecuta una sola preocupación: el bug del issue `#176`.
- Trabaja de forma autónoma hasta demostrar una de las dos salidas permitidas: **ya corregido** o **corregido mediante un PR nuevo**. No declares éxito por inspección estática solamente.

## Instrucción para el agente

Resuelve exclusivamente el issue [#176](https://github.com/bcaceres19/barberia/issues/176), vinculado a `HU-012`: en 320 px y 360 px ninguna etiqueta necesaria de la navegación privada puede truncarse hasta perder significado. Primero verifica si la solución ya integrada por el PR [#202](https://github.com/bcaceres19/barberia/pull/202) corrigió el defecto. Si el bug ya no se reproduce en `main`, no generes un cambio de código redundante: documenta evidencia actual en el issue y ciérralo como resuelto por ese PR. Si todavía se reproduce, conserva una prueba de regresión que falle, aplica la corrección mínima y entrega un PR propio.

Este es un prompt de corrección funcional de una HU, no un prompt de QA ni de ampliación visual. Las pruebas y la evidencia indicadas son obligatorias para demostrar el arreglo, pero no constituyen una entrega independiente.

## Objetivo

Dejar el issue `#176` correctamente resuelto y cerrado con evidencia verificable en la versión actual de `main`: el dock móvil de `HU-012` mantiene etiquetas comprensibles, todos los destinos reales siguen siendo alcanzables, no introduce desplazamiento horizontal y conserva navegación por teclado, foco y área segura en 320 px y 360 px.

## Contexto confirmado al redactar este prompt

- El issue `#176` permanece abierto y no tiene comentarios.
- El PR `#202`, integrado en `main` con squash `9195b49`, cerró el issue `#187` y cambió el dock móvil de seis destinos planos a `Agenda`, `Servicios`, `Barberos` y `Más`.
- `AppNav.vue` separa actualmente destinos `primary` y secundarios; los secundarios se muestran en un `BaseDialog` accesible junto con `Cerrar sesión`.
- `AppNav.test.ts` ya cubre la separación primary/secondary, apertura y cierre del diálogo, restauración de foco, logout y `vitest-axe`.
- `apps/web/e2e/panel-evidencia-responsiva.spec.ts` ya recorre 320/360/768/1280 px y zoom 200 %, comprueba ausencia de scroll horizontal y navega mediante `Más`.
- Existe evidencia histórica posterior al arreglo en `apps/web/e2e/evidence/panel/{320,360}/`; no la presentes como prueba actual sin volver a ejecutar el recorrido sobre el SHA revisado.
- No hay una duda `DP-*` ni una contradicción `CT-*` abierta que bloquee esta corrección.

## Preflight obligatorio

1. Ejecuta `git status --short --branch`. No sobrescribas ni incluyas cambios ajenos. Si el árbol está sucio, usa otro worktree o detente antes de mezclar archivos.
2. Verifica en GitHub que `#176` sigue abierto y relee completamente su cuerpo y comentarios. Comprueba también el estado y contenido de `#187` y del PR `#202`.
3. Actualiza `main` solo mediante fast-forward y registra el SHA exacto evaluado.
4. Ejecuta `graphify query "HU-012 AppNav dock navegación móvil truncamiento issue 176 PR 202"` cuando exista `graphify-out/graph.json`.
5. Lee completamente `AGENTS.md`, `CLAUDE.md`, `CONTRIBUTING.md` y todos los `source_docs` de los metadatos. Relee en particular `HU-012`, `CA-012-08` y `estandar-diseno-visual.md` §6.5.
6. Inspecciona `AppNav.vue`, `PrivateShell.vue`, los `NavItem` aportados por los módulos, `AppNav.test.ts` y `panel-evidencia-responsiva.spec.ts`. Confirma el número y los nombres reales de los destinos; no los deduzcas del issue antiguo.
7. Comprueba dudas y contradicciones vigentes. Si aparece una decisión nueva que cambie la navegación, registra el bloqueo antes de editar.
8. No crees rama todavía. Primero ejecuta la reproducción actual descrita abajo.

## Reproducción y decisión de salida

Ejecuta la aplicación real con datos sintéticos y una sesión válida. En Chromium, fija y verifica el viewport efectivo con `window.innerWidth` en 320 px y 360 px. En cada ancho:

1. Abre una ruta privada que muestre el shell completo.
2. Inspecciona cada etiqueta visible del dock y confirma que ninguna depende de elipsis para distinguirse.
3. Abre `Más` y confirma que todos los destinos secundarios reales son legibles y alcanzables.
4. Recorre el dock y el diálogo solo con teclado; confirma foco visible, `Escape`, restauración del foco y `aria-current` en el destino activo.
5. Confirma que `document.documentElement.scrollWidth <= document.documentElement.clientWidth` y que el safe area inferior no oculta controles.
6. Ejecuta `axe-core` en vivo sobre el dock cerrado y el diálogo abierto; exige cero violaciones atribuibles al cambio.
7. Guarda capturas actuales de ambos anchos y registra el SHA, comando, navegador y resultado.

Después de reproducir, toma exactamente una ruta:

- **Ruta A — ya corregido:** si todos los puntos pasan en `main`, no edites producto, pruebas ni documentación normativa. Añade al issue `#176` un comentario con SHA, relación explícita con PR `#202`, comandos, resultados y enlaces a evidencia versionada ya existente o nueva si debe persistirse. Cierra el issue como resuelto por el cambio integrado. No abras un PR vacío.
- **Ruta B — todavía falla:** si un punto falla, conserva una prueba de regresión que reproduzca el defecto, crea `fix/176-dock-navegacion-movil` desde `main` actualizada y aplica la corrección mínima. Solo entonces actualiza este prompt a `in_progress` con la rama real; no reescribas materialmente su cuerpo.

## Alcance incluido

- Diagnóstico actual del dock y el menú `Más` de `HU-012` en 320 px y 360 px.
- Corrección mínima en `apps/web/src/modules/auth/components/AppNav.vue`, `PrivateShell.vue` o el modelo `NavItem` únicamente si la reproducción demuestra que sigue siendo necesaria.
- Conservación de una prueba de regresión en la capa más baja que reproduzca fielmente el truncamiento o la pérdida de acceso.
- Evidencia en navegador real del estado final, incluida accesibilidad y reflow.
- Actualización estrictamente necesaria de trazabilidad y del catálogo si existe un nuevo cambio de repositorio.
- Comentario y cierre de `#176` cuando el resultado esté probado.

## Fuera de alcance

- Rediseñar pantallas de negocio, la cabecera, el shell completo o el lenguaje visual NAVA.
- Crear, eliminar, renombrar o reordenar destinos sin una fuente normativa vigente.
- Cambiar rutas, guards, rehidratación de sesión, tenant activo, contratos HTTP, backend o base de datos.
- Alterar el comportamiento del logout o esconderlo.
- Convertir el trabajo en una campaña general de QA o arreglar otros hallazgos.
- Tocar el issue de flakiness `#158`, los issues de pruebas `#86`–`#88` o cualquier issue `feat`/`chore` abierto.
- Agregar dependencias o usar snapshots como única evidencia.

## Invariantes que debes conservar

- Todas las rutas privadas reales continúan alcanzables sin escribir la URL manualmente.
- El destino activo conserva nombre accesible y `aria-current="page"`.
- El menú `Más` mantiene nombre accesible, foco atrapado, cierre con `Escape` y restauración al disparador.
- `Cerrar sesión` sigue accesible y revoca la sesión mediante el flujo existente.
- La navegación no recarga la aplicación completa y las rutas conservan carga diferida.
- No se introduce scroll horizontal ni se oculta contenido tras el dock o el safe area.
- No se usa color como única señal ni se debilitan objetivos táctiles, foco o contraste.

## Trabajo requerido si el defecto sigue presente

1. Añade primero una regresión automatizada que falle por el comportamiento observado. No escribas una aserción sobre una clase CSS si puedes observar nombre, visibilidad, geometría o acceso como usuario.
2. Corrige la causa más pequeña. Reutiliza la separación `primary`/secundaria y `BaseDialog` si siguen siendo apropiados; no construyas otro sistema de navegación paralelo.
3. Verifica todos los `extraItems` aportados por módulos para que el arreglo no dependa solo de un fixture reducido.
4. Repite la reproducción en 320 px y 360 px y ejecuta también 768 px, 1280 px y zoom de texto 200 % para detectar regresiones.
5. Actualiza `docs/00-control/matriz-trazabilidad.md` y `docs/00-control/historial-cambios.md` solo si hubo un cambio material. Actualiza los metadatos y el índice de este catálogo con rama, PR y estado reales.
6. Revisa el diff completo y separa cualquier hallazgo fuera de alcance en otro issue; no lo incorpores al PR `#176`.

## Pruebas y evidencia

- Prueba de componente de `AppNav` con el conjunto realista de destinos móviles y estados del menú.
- Regresión E2E de `HU-012` en 320 px y 360 px que compruebe etiquetas completas/comprensibles, ausencia de overflow y acceso a todos los destinos.
- Teclado: `Tab`, `Shift+Tab`, `Enter`/`Space`, `Escape`, foco visible y restauración de foco.
- `axe-core` en navegador real con el dock cerrado y `Más` abierto.
- Evidencia visual actual en 320/360 px y control de regresión en 768/1280 px más zoom 200 %.
- No actualices capturas ni snapshots sin inspeccionar y explicar la diferencia.

## Verificación final

Ejecuta desde la raíz, ajustando únicamente el comando E2E al entorno documentado del repositorio:

```text
pnpm --filter @system-barbershop/web format
pnpm --filter @system-barbershop/web lint
pnpm --filter @system-barbershop/web typecheck
pnpm --filter @system-barbershop/web test:unit -- AppNav.test.ts
pnpm --filter @system-barbershop/web test:unit
pnpm --filter @system-barbershop/web build
pnpm --filter @system-barbershop/web test:e2e -- panel-evidencia-responsiva.spec.ts
git diff --check
```

No ocultes una limitación de entorno. Si el E2E real no puede ejecutarse, el issue no se cierra como verificado y la entrega explica exactamente el bloqueo.

Entrega esta tabla en el comentario del issue y, si existe, en el PR:

| Criterio | Estado | Prueba o evidencia |
| --- | --- | --- |
| `CA-012-08`: navegación operable y responsive | `PASS`/`FAIL` | Comando, viewport y evidencia |
| Etiquetas comprensibles a 320/360 px | `PASS`/`FAIL` | Captura y comprobación observable |
| Todos los destinos reales alcanzables | `PASS`/`FAIL` | Recorrido y aserción E2E |
| Sin overflow horizontal ni solape por safe area | `PASS`/`FAIL` | Medición en navegador real |
| Teclado, foco y `aria-current` | `PASS`/`FAIL` | Recorrido y prueba |
| `axe-core` en vivo | `PASS`/`FAIL` | Resultado por estado |
| Formato, lint, tipos, unitarias y build | `PASS`/`FAIL` | Comandos y resultados |

## Git, issue y PR

- Si aplica la Ruta A, no crees rama, commit ni PR. Comenta y cierra `#176` con evidencia de que el PR `#202` ya resolvió el defecto.
- Si aplica la Ruta B, usa la rama `fix/176-dock-navegacion-movil`.
- Commit y título del PR: `fix(web): evita truncar el dock movil`.
- Usa `Closes #176` solo si el PR cubre el issue completo; de lo contrario, `Refs #176` y deja explícito qué falta.
- No hagas push directo, force push ni merge de `main`. Integra solo por squash cuando todos los checks aplicables estén en verde y se cumplan las reglas de revisión.

## Registro de ejecución

- 2026-09-13: preflight completado sobre `main` en `378bd52d6cf834befc293f2a236809541220b4e0`; `#176` seguía abierto sin comentarios y el PR `#202` estaba integrado como `9195b494e27820e0158887ff18ba27bc7b080321`.
- La comprobación estática y automatizada disponible pasó: `AppNav.test.ts` (12 pruebas), la suite unitaria completa (829 pruebas), lint (0 errores), tipos, build y formato.
- Ejecución completada el 2026-09-13: Chromium administrado por Playwright verificó el dock contra API y PostgreSQL temporales con datos sintéticos. La regresión E2E, las capturas actuales y la evidencia live de axe-core se integraron mediante el squash del PR [#258](https://github.com/bcaceres19/barberia/pull/258), SHA `101f78446c2a59db721a3947aac98760d006f33f`; GitHub cerró el issue `#176` automáticamente.
