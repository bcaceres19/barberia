---
prompt_id: "PROMPT-HU-009-v1"
version: "1.0"
kind: "hu"
status: "executed"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-009"
related_hu: []
issue: "42"
issue_url: "https://github.com/bcaceres19/barberia/issues/42"
suggested_issue_title: "feat(web): completar HU-009 sistema visual base"
branch: "feat/42-hu009-sistema-visual"
pr: 43
pr_url: "https://github.com/bcaceres19/barberia/pull/43"
depends_on:
  - "Secuencia de ejecución posterior a HU-004"
rules:
  - "Criterios no funcionales de UX y accesibilidad"
decisions:
  - "DEC-033"
  - "DEC-035"
  - "DEC-038"
  - "DEC-039"
acceptance_criteria:
  - "CA-009-01"
  - "CA-009-02"
  - "CA-009-03"
  - "CA-009-04"
  - "CA-009-05"
  - "CA-009-06"
  - "CA-009-07"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/prioridades.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/04-arquitectura/frontend.md"
  - "apps/web/package.json"
  - "apps/web/src/shared/ui"
created_at: "2026-08-12"
updated_at: "2026-08-13"
supersedes: "docs/10-backlog/prompts-implementacion.md#prompt-de-hu-009--sistema-visual-base-en-componentes"
superseded_by: null
---

# Completar HU-009: sistema visual base

## Instrucción para Claude o Codex

Implementa únicamente `HU-009` mediante una auditoría y refactor de los componentes existentes. No dupliques componentes ni construyas pantallas de `HU-010`, `HU-011` o `HU-012`. Produce un pull request borrador con pruebas, evidencia responsive y accesible.

## Objetivo

Dejar en `apps/web/src/shared/ui/` un sistema base coherente y reutilizable para botón, campo de texto, alerta, insignia y diálogo. Todas las decisiones visuales deben provenir de tokens semánticos; los controles serán operables con teclado, táctiles, responsive y compatibles con movimiento reducido.

## Preflight obligatorio

1. Verifica árbol limpio, `main` actualizada por fast-forward y ausencia de cambios ajenos que puedan sobrescribirse.
2. Consulta Graphify por `HU-009`, `shared/ui`, tokens, pruebas de componentes y dependencias si el grafo está disponible.
3. Lee completamente cada `source_docs` y revisa el código actual antes de proponer componentes nuevos.
4. Confirma la secuencia solicitada: este prompt se ejecuta después de integrar `HU-004`, aunque la HU normativa no tenga dependencia funcional propia.
5. Busca un issue que cubra exactamente `HU-009`. Si no existe, créalo con el título sugerido, alcance, criterios, evidencia y exclusiones.
6. Registra issue/URL en este archivo y en el índice, cambia a `ready`, crea `feat/<issue>-hu009-sistema-visual` y cambia a `in_progress` con la rama real.

## Alcance incluido

- Auditar y completar los componentes existentes `BaseButton`, `BaseInput`, `BaseAlert`, `BaseBadge` y `BaseDialog`, o sus equivalentes reales.
- Variantes semánticas por intención mediante props tipadas; nunca colores, tamaños o `z-index` libres.
- Patrones reutilizables para estado inicial, carga, actualización, vacío, error recuperable, error de campo, conflicto y éxito, sin inventar componentes que todavía no tengan uso real.
- Foco visible, navegación coherente, áreas táctiles mínimas, contraste, nombres accesibles y movimiento reducido.
- Pruebas de componente, verificación accesible automatizada, revisión manual por teclado y evidencia a 320, 360, 768 y 1280 px.
- Checks de frontend en CI si aún no existen.

## Fuera de alcance

- Pantallas de acceso, recuperación o cascarón privado.
- Modo oscuro, selector de tema, personalización por barbería o biblioteca visual externa.
- Iconografía o dependencias nuevas sin necesidad demostrada y justificación normativa.
- Galería, playground o ruta de producción creada solo para pruebas.
- CSS local específico de una pantalla futura.

## Estado existente que debes conservar

- Ya existen `tokens.css` y cinco componentes base; parte de ellos, identifica brechas y refactoriza en lugar de crear duplicados.
- La dependencia permitida es `app → modules → shared`; `shared` no importa módulos ni aplicación.
- TypeScript se mantiene estricto y no se usa `any`, estado global ni una API visual de valores libres.
- Los nombres de props y eventos existentes se conservan cuando sean correctos; cualquier ruptura debe estar justificada, buscada en todos los consumidores y cubierta por pruebas.
- El criterio formal “Terminado cuando” de la HU se revalidará al construir `HU-010` a `HU-012`; esta entrega demuestra que la base está lista, no inventa evidencia futura.

## Auditoría inicial requerida

Antes de editar, busca y registra brechas de:

- colores hexadecimales, `rgb`, tipografías, radios, espacios, sombras, duraciones o tamaños literales fuera de los tokens permitidos;
- props como `color`, `class`, `size` o `zIndex` que permitan eludir el sistema semántico;
- controles menores de 44 × 44 px;
- estilos de foco ausentes, ocultos o dependientes solo de color;
- elementos con `aria-hidden` que aún reciben foco;
- IDs aleatorios o relaciones label/description/error inestables;
- diálogos sin gestión de foco, cierre por teclado, retorno de foco o bloqueo de fondo;
- pruebas que solo comprueban clases internas en lugar del comportamiento accesible.

Incluye el resultado de esta auditoría en el PR y enlaza cada brecha corregida con una prueba.

## Trabajo requerido

1. Revisa `tokens.css` contra el estándar. Agrega únicamente tokens semánticos necesarios para los cinco componentes y los patrones aprobados; no agregues una segunda paleta.
2. `BaseButton`: variantes por intención, estados normal/deshabilitado/cargando, nombre accesible estable, foco visible y objetivo táctil mínimo. El estado de carga no debe provocar doble envío.
3. `BaseInput`: label real, ayuda y error asociados con IDs deterministas, atributos de validez, foco visible y propagación tipada del valor. No uses placeholder como etiqueta.
4. `BaseAlert`: tono semántico, icono decorativo cuando aplique y semántica apropiada (`status` o `alert`) sin anunciar contenido dos veces.
5. `BaseBadge`: texto comprensible además del color y contraste suficiente; no la conviertas en control si solo comunica estado.
6. `BaseDialog`: nombre/descripción accesibles, foco inicial controlado, trampa de foco, cierre con Escape cuando esté permitido, retorno de foco y fondo no interactivo mientras esté abierto. Evita colisiones con diálogos anidados o múltiples instancias.
7. Representa los patrones de pantalla mediante composiciones/tokens o componentes solo si hay uso real inmediato dentro del alcance; no construyas abstracciones especulativas.
8. Respeta reflow, zoom al 200 %, teclado, lector de pantalla y `prefers-reduced-motion`. Si hay animaciones, usa tokens entre 120 y 200 ms y desactívalas cuando se solicita movimiento reducido.
9. Si la CI actual no valida frontend, agrega un job mínimo reproducible para instalación bloqueada, lint/tipos, pruebas, build y controles accesibles aplicables.

## Pruebas y evidencia

- Pruebas de componente con Vitest y Vue Test Utils, o las herramientas fijadas por el repositorio, para estados normal, deshabilitado, cargando y error cuando apliquen.
- Interacciones por teclado y foco observable para todos los componentes interactivos.
- Nombre, rol, descripción, error y estado consultados por semántica accesible, no por clases privadas.
- Diálogo: apertura, recorrido de tabulación, Escape, foco de retorno y fondo no interactivo.
- Verificación automatizada WCAG 2.2 AA con la herramienta aprobada; registra herramienta y resultado.
- Evidencia visual a 320, 360, 768 y 1280 px, sin desplazamiento horizontal; añade comprobación con zoom 200 %.
- Evidencia de `prefers-reduced-motion` y duraciones permitidas.
- Un harness de prueba o desarrollo puede usarse para capturas, pero no debe quedar accesible como ruta de producción.

Entrega una tabla `Criterio | Estado | Prueba o evidencia` para `CA-009-01` a `CA-009-07`. Marca como verificación posterior, sin falsos positivos, la condición que depende de `HU-010` a `HU-012`.

## Documentación y trazabilidad

- Documenta la API pública de los componentes, variantes permitidas y ejemplos mínimos en el lugar definido por el frontend.
- Actualiza matriz, historial, estándar visual o estrategia de pruebas solo cuando el cambio los afecte realmente.
- Registra y justifica toda dependencia nueva; evita agregarla si el navegador o las herramientas actuales resuelven el requisito con seguridad.
- Actualiza metadatos e índice del catálogo con issue, rama, PR y estado reales.

## Verificación final

Ejecuta los scripts equivalentes definidos en `apps/web/package.json` para:

```text
instalación con lockfile inmutable
lint
comprobación de tipos
pruebas de componentes
verificación accesible
build de producción
E2E o harness de evidencia responsive aplicable
git diff --check
graphify update .
```

No actualices snapshots sin revisar el comportamiento y no ocultes pruebas inestables.

## Git y PR

- Commits y título: `feat(web): completa sistema visual base` o equivalente Conventional Commits.
- Abre PR borrador contra `main` y usa `Closes #<issue>` solo si cubre el issue completo.
- Incluye matriz de criterios, auditoría antes/después, resultados de CI, evidencia responsive/accesible, riesgos y recuperación.
- No hagas merge, push directo, force push ni reescribas `main`.
