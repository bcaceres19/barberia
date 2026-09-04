---
prompt_id: "PROMPT-CHORE-189-AGENDA-DIARIA-FIDELIDAD-v1"
version: "1.0"
kind: "chore"
status: "superseded"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-062"
related_hu:
  - "HU-009"
  - "HU-012"
  - "HU-063"
issue: "189"
issue_url: "https://github.com/bcaceres19/barberia/issues/189"
suggested_issue_title: null
branch: "chore/189-agenda-diaria-fidelidad"
pr: null
pr_url: null
depends_on:
  - "Atlas `panel-agenda-eventos` (24 PNG + README) y `tools/mockups/panel-agenda-eventos` integrados en main"
  - "Issue #212 (controles reglados en shared/ui) integrado en main mediante PR #214"
  - "Issue maestro #184"
rules: []
decisions:
  - "DEC-074"
  - "DEC-075"
  - "DEC-077"
  - "DEC-079"
  - "DEC-080"
acceptance_criteria:
  - "CA-062-01"
  - "CA-062-02"
  - "CA-062-03"
  - "CA-063-01"
  - "CA-063-02"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/panel-agenda-eventos/README.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/00-control/registro-decisiones.md"
created_at: "2026-09-03"
updated_at: "2026-09-03"
supersedes: null
superseded_by: "PROMPT-CHORE-189-AGENDA-DIARIA-FIDELIDAD-v2"
---

# Fidelidad visual de `/panel` (agenda diaria)

## Nota de persistencia (2026-09-03)

Este archivo se redactó en una sesión previa como parte del trabajo en curso de `chore/213-fidelidad-acceso-recuperacion`, pero nunca se guardó en un commit — existía solo en el árbol de trabajo de esa rama, sin relación con su propósito (fidelidad de `/acceso`/`/recuperar-acceso`). Por eso no aparece en el historial de `main` antes de esta entrega. Se persiste aquí, en la rama que realmente lo ejecuta, con el cuerpo original íntegro (instrucción, alcance, decisiones citadas) y el estado actualizado a `executed`, tal como exige `AGENTS.md`: "no se considera entregado si existe únicamente en un chat".

## Instrucción para el agente

Trabaja en modo **fidelidad de mockup**. Se te asignan 24 imágenes de referencia exactas y la app real debe reproducirlas.

Esta entrega es **exclusivamente visual**: composición, geometría, jerarquía, color, densidad y estado visual. **No cambia comportamiento, validaciones, endpoints, máquinas de estado ni textos de producto.** Cuando un mockup y el código difieran en _texto_ o en _comportamiento_, manda el código: registras la diferencia en la tabla de desviaciones. Cuando difieran en _composición_, manda el mockup.

Modelo y esfuerzo requeridos: **Claude Sonnet 5 con `effortLevel: high`**. Carga y sigue las skills `nava-mockup-fidelity` y `browser-viewport-verification` antes de editar y antes de declarar terminada cualquier pantalla.

## Issue

[#189 · chore(web): rediseno agenda diaria NAVA Tailored Grid](https://github.com/bcaceres19/barberia/issues/189), fase 4 del issue maestro [#184](https://github.com/bcaceres19/barberia/issues/184).

**Estado `blocked` original:** no comenzaba hasta que el atlas `panel-agenda-eventos` estuviera integrado en `main`. Se generó y se integró mediante PR [#216](https://github.com/bcaceres19/barberia/pull/216) el 2026-09-03, lo que desbloqueó esta ejecución.

## Referencia asignada

Directorio `docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/panel-agenda-eventos/`:

- `desktop/panel/01..12` y `mobile/panel/01..12`;
- su `README.md` es el contrato visual normativo y trae la tabla que nombra cada evento.

Un mismo número describe el mismo estado funcional en ambos viewports; termina el par antes de dar el evento por cerrado.

**No uses** `ui-mockups-nava-tailored-grid-2026-09-02/01-agenda-diaria-responsive.png`, que es lo que el cuerpo del issue #189 cita: es la lámina compuesta original y queda como historial visual. El atlas por evento la sustituye como objetivo de implementación, igual que ocurrió en `/acceso` con `02-acceso-recuperacion.png`.

No edites los PNG ni ejecutes su generador (`tools/mockups/panel-agenda-eventos/render.mjs`).

## Alcance incluido

1. `apps/web/src/modules/agenda/pages/DailyAgendaPage.vue` y los componentes que se extraigan de ella.
2. Componentes compartidos nuevos en `apps/web/src/shared/ui`, creados aquí y **adoptados solo por `/panel`** en esta entrega:
   - `BaseSpinner.vue` — anillo de latón interrumpido con el rombo del divisor NAVA al centro.
   - `PageState.vue` — estado de página centrado (marca o indicador, divisor, rótulo de estado, titular en serif, cuerpo y acción).
   - `BarberAvatar.vue` — retrato del barbero con monograma derivado de `fullName`.
3. Esqueleto de la agenda, local al módulo `agenda` porque su geometría es la de esta pantalla.
4. Pruebas de componente, E2E y evidencia responsiva de lo modificado.

Los tres componentes de `shared/ui` los van a consumir después otras rutas con sus propios issues. Créalos con API estable y pruebas, pero **no los adoptes fuera de `/panel`**: eso mezclaría preocupaciones, igual que `OtpInput` se creó en #212 y se adoptó en #213.

## Fuera de alcance

- **La fotografía del barbero.** El atlas la representa por instrucción del propietario, pero `Barber` es exactamente `id`, `fullName`, `createdAt` y `updatedAt`, y `apps/web/src/modules/staff/model/barber.ts` documenta de forma explícita que no se inventa un campo que el contrato no declare. Implementa **solo el monograma**, que se deriva de `fullName`. No añadas campo, carga, almacenamiento ni recorte de imagen: eso necesita `DEC-*` e issue propios.
- **Todo lo funcional.** No toques `api/`, `model/`, `guards/`, `apps/api`, el contrato OpenAPI ni la base de datos. Si una diferencia visual parece exigir un cambio funcional, **no lo hagas**: anótalo en la tabla de desviaciones.
- Rediseñar cualquier otra ruta, aunque herede los componentes nuevos.
- Agenda multi-barbero y estado `in_progress`, ya excluidos por el issue.

## Estado existente que debe conservarse

- **`DEC-074`:** la agenda es de un solo barbero seleccionado. La línea temporal no se convierte en cuadrícula multi-barbero.
- **`DEC-075` y navegación por fecha (`HU-063`):** `canNavigateDates` gobierna `Anterior`, `Fecha` y `Siguiente`. Cuando la zona horaria no se pudo confirmar, esos tres controles quedan deshabilitados **y la agenda sigue visible**.
- **La línea temporal es decorativa** (`aria-hidden`, `tabindex="-1"`) y la lista es la representación accesible equivalente. Las dos se conservan en los dos viewports; el mockup dibuja ambas justamente para que no se borre ninguna.
- **El marcador `Ahora` es contextual y decorativo:** no cambia `confirmed` ni crea disponibilidad, y solo existe cuando la fecha mostrada es hoy.
- **Textos actuales** de `DailyAgendaPage.vue` y `model/dailyAgenda.ts`, incluidos los cinco rótulos de estado (`Confirmado`, `Completado`, `Cancelado por el cliente`, `Cancelado por el barbero`, `No se presentó`) y la línea `{fecha} · Zona {timezone}`.
- **Accesibilidad ya lograda:** `role="status"`/`aria-live="polite"` en carga y actualización, `role="alert"` en los errores sin robar foco, y objetivo táctil mínimo de 44 px en el reintento.

## Trabajo requerido

### 1. Contrato medido antes de editar

Abre cada PNG a resolución original y completa la tabla de contrato de `nava-mockup-fidelity` por evento. Captura baseline de la app real en los dos viewports antes de tocar código.

### 2. Calendario

- Cada ficha se posiciona y dimensiona **desde su propia hora** contra el eje: el borde izquierdo y el ancho salen de `startsAt`/`endsAt`, nunca de un reparto arbitrario. Hoy la geometría no corresponde al horario.
- Las etiquetas del eje y las marcas (`Ahora`, `Cambio de día`) viven en **bandas separadas** y nunca se superponen. Cada marca traza además su línea vertical por el carril; la de medianoche cae exactamente sobre `00:00`.
- El carril es una región definida con guía por hora y media hora tenue entre cada par.
- La ficha muestra rango horario, persona y servicio cuando la duración le da ancho; con poco ancho conserva solo la hora de inicio.

### 3. Superficie y jerarquía

- Las fichas y filas usan **pergamino** (`--color-surface-muted`, `#E8E2D8`), no blanco puro: sobre tinta el blanco daba ~17:1 y deslumbra en una pantalla de uso continuo.
- **Los turnos terminales cambian de material, no de peso:** un turno vigente es superficie de papel; uno cerrado es un registro con contorno sobre la tinta. Un relleno gris los dejaba pesando igual o más que un turno vigente.
- Las insignias sobre pergamino se apoyan en contorno y texto del color de estado.

### 4. Estados de página

- Cuando la carga, el error o el vacío ocupan toda el área de contenido, se componen **centrados** con `PageState`: indicador o divisor NAVA, rótulo de estado en versalitas, titular en serif, cuerpo y acción real. Un renglón suelto arriba a la izquierda deja la pantalla pareciendo rota.
- La alerta como nota al margen (`BaseAlert` de #212) se reserva para cuando acompaña contenido que sigue visible: el evento `10`.
- El texto de espera y el de vacío se componen como **titular en serif**; el `Reintentar` es un botón real, no una etiqueta contorneada.
- Cuando barbero y fecha ya están resueltos, la espera usa **esqueleto** con la geometría del contenido por llegar (evento `05`). Sin contexto resuelto se usa el estado centrado (evento `02`).
- Si el mensaje nombra otra sección del producto, ese destino se resalta en latón (evento `04`, «Barberos»).

### 5. Coherencia entre estado y mensaje

Todo esto ya es el comportamiento actual: es cuestión de reflejarlo, no de cambiarlo.

- Barbero no disponible ⇒ la selección queda **vacía**, porque conservar al barbero inexistente contradice el mensaje que pide elegir otro.
- Zona horaria no confirmada ⇒ la agenda se dibuja y solo se bloquea la navegación por fecha.
- Sin contexto resuelto ⇒ no hay filtros, fecha ni `Nuevo turno`.
- Día válido sin turnos ⇒ se conserva el **eje del día vacío** con el marcador `Ahora`, y `Nuevo turno` vive dentro del estado vacío, sin repetirse en el encabezado.

### 6. Retrato del barbero

Dondequiera que un barbero se ve o se elige, su nombre va con un retrato cuadrado de radio `2 px` y filete de latón: `28 px` en el selector cerrado y `34 px` en cada opción de la lista (`26`/`32 px` en móvil). Nunca aparece un retrato sin nombre, y es decorativo para lectores de pantalla porque el nombre ya da la identidad.

Implementa **el monograma**: iniciales derivadas de `fullName` en la serif del wordmark sobre tinta. La variante con fotografía del evento `12` queda representada pero **no se implementa** (ver _Fuera de alcance_); deja el componente preparado para recibirla después sin rediseñarlo.

### 7. Paridad entre viewports

Encabezado, filtros, navegación por fecha, lista y dock son los mismos componentes en escritorio y móvil. Móvil no introduce encabezado propio ni una segunda forma de navegar; el dock agrupa los destinos secundarios en «Más» sin sacarlos del árbol.

## Pruebas y evidencia

- **Componente.** `BaseSpinner`, `PageState` y `BarberAvatar` con sus variantes y estados, incluido el monograma de un nombre de una sola palabra y de uno con tres. En `DailyAgendaPage`: posición y ancho de ficha derivados de la hora, ausencia del marcador `Ahora` fuera de hoy, selección vacía en `not-found`, agenda visible con navegación bloqueada sin zona horaria.
- **Actualiza** las pruebas existentes que asuman la composición anterior. Si una aserción deja de aplicar, cámbiala y explica por qué en el PR; no la relajes para que pase.
- **E2E.** `agenda-diaria.spec.ts`, `agenda-navegacion-fecha.spec.ts` y `panel-evidencia-responsiva.spec.ts`. Ejecuta la suite completa sobre `main` **antes** de tocar código y guarda esa línea base: si algo ya venía roto, repórtalo en su issue y no lo arrastres a tu PR.
- **Accesibilidad.** `axe-core` en vivo sobre `/panel` en cada estado con contenido nuevo, con 0 violaciones. Verifica foco visible, orden de tabulación, 200 % de zoom y `prefers-reduced-motion` (el indicador de progreso debe detenerse).
- **Evidencia visual, obligatoria por evento.** Los 12 eventos × 2 viewports: captura de la app real, comparación lado a lado con el PNG y overlay o diff. Guarda bajo `apps/web/e2e/evidence/` con la convención de las suites de evidencia responsiva. Verifica `window.innerWidth`/`innerHeight` reales antes de capturar.
- **Anchos obligatorios adicionales:** 320, 360, 768 y 1280 px.

## Documentación y trazabilidad

- Actualiza `docs/03-desarrollo/estandar-diseno-visual.md` con los tres componentes compartidos nuevos y todo token que añadas.
- Actualiza `docs/00-control/matriz-trazabilidad.md` e `historial-cambios.md` en lo que realmente resulte afectado.
- Entrega una **tabla de desviaciones** con toda diferencia entre mockup y app que no se cerró, su motivo y a qué issue futuro pertenece. La fotografía del barbero va ahí.
- Actualiza los metadatos de este archivo (`status`, `branch`, `pr`) y su fila en el índice [`docs/10-backlog/prompts/README.md`](../README.md).

## Verificación final

```text
cd apps/web
pnpm format
pnpm lint
pnpm typecheck
pnpm test:unit
pnpm test:e2e
pnpm build
```

Entrega una tabla:

`Criterio | Estado | Prueba o evidencia`

Una fila por cada uno de los 12 eventos en cada viewport (24 en total) y una por cada criterio `CA-*` de los metadatos. No declares cumplido aquello que no esté probado, y no cierres un evento en un viewport sin haber cerrado su hermano.

## Git y PR

- Rama `chore/189-agenda-diaria-fidelidad` desde `main` actualizada.
- Commit y PR: `chore(web): fidelidad visual de la agenda diaria con los mockups por evento`.
- `Closes #189` solo si se cubre el issue completo; en caso contrario `Refs #189`.
- No hagas push directo, force push ni merge de `main`.

## Resultado de la ejecución (2026-09-03)

Implementado con `Refs #189`, no `Closes #189`: el gate C completo de evidencia (24 pares evento×viewport con baseline/final/lado-a-lado/overlay) no se completó en esta sesión — ver `apps/web/e2e/evidence/panel-fidelidad-189-desviaciones.md` para el detalle de qué se verificó y qué queda pendiente. Comportamiento, `DEC-074`/`DEC-075`, textos y contrato de datos se conservaron sin cambios; la geometría de ficha desde `startsAt`/`endsAt` resultó ya correcta (no se reescribió). `e2e/agenda-diaria.spec.ts`/`e2e/agenda-navegacion-fecha.spec.ts` nunca habían pasado contra Chromium real — brecha preexistente reportada, no arrastrada a este PR.
