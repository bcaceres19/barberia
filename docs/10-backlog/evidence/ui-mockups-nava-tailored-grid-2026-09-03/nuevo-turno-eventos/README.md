# Mockups de `/panel/turnos/nuevo` (Nuevo turno) por viewport y evento

## Propósito

Este directorio reemplaza como referencia de implementación la lámina compuesta `03-nuevo-turno.png` de `ui-mockups-nava-tailored-grid-2026-09-02`. La lámina original se conserva como historial visual, pero no debe entregarse a un agente de implementación: cada archivo de este directorio representa una sola pantalla y un solo evento, igual que ocurrió en `/acceso` con `02-acceso-recuperacion.png` y en `/panel` con `01-agenda-diaria-responsive.png`.

La lámina compuesta mostraba, en su tercer panel, botones de cuenta decorativos y navegación duplicada en el encabezado (`Nuevo turno` y `Cerrar sesión` como botones de la barra superior, repetidos luego en el dock). Esa es la misma inconsistencia que ya se corrigió al construir [`auth-eventos`](../auth-eventos/README.md) y [`panel-agenda-eventos`](../panel-agenda-eventos/README.md): el encabezado y el dock de este directorio son el mismo cascarón ya validado en esos dos atlas, sin controles de cuenta decorativos.

**Revisión del 2026-09-04 (segundo pase sobre el formulario).** Las cuatro tarjetas con borde propio pasaron a ser bandas de una sola hoja, la columna de rótulos creció hasta que los subtítulos dejaran de partirse, el CTA se alineó con la columna de campos y el filete de latón quedó reservado para los campos ya resueltos. Los veinticuatro PNG se volvieron a generar para alinear la pantalla con los dos atlas hermanos antes de implementarla. La primera versión heredaba de la lámina compuesta un formulario de tarjetas de pergamino con campos blancos, una espera que era un indicador suelto en una pantalla vacía, un éxito reducido a una alerta pequeña arriba a la izquierda y un ancho de `900 px` centrado que dejaba media pantalla libre y desalineaba el encabezado del de `/panel`. Nada de eso existía en `auth-eventos` ni en `panel-agenda-eventos`, así que el atlas quedaba fuera del sistema justo en la pantalla donde más se escribe. Los doce eventos, sus nombres y su copy no cambian.

La estructura es deliberadamente simétrica:

- `desktop/nuevo-turno/` y `mobile/nuevo-turno/` contienen los mismos doce eventos;
- un mismo número y nombre describen el mismo estado funcional en ambos viewports;
- encabezado y dock son los mismos componentes en escritorio y móvil, ya presentes en `panel-agenda-eventos`.

## Nuevo turno (`/panel/turnos/nuevo`)

| N.º | Archivo en cada viewport | Evento representado |
| --- | --- | --- |
| 01 | `01-carga-contexto.png` | Carga inicial: barberos y zona horaria aún sin resolver. |
| 02 | `02-error-contexto.png` | Fallo al cargar barberos/zona horaria: sin contexto no hay formulario. |
| 03 | `03-sin-barberos.png` | Sin barberos activos: no hay formulario que ofrecer. |
| 04 | `04-formulario-vacio.png` | Formulario recién abierto: nada elegido. Único evento sin resumen, porque `hasSummaryContent` todavía es falso. |
| 05 | `05-cargando-servicios.png` | Barbero elegido, servicios cargando: el campo de servicio queda deshabilitado y el resumen ya muestra el barbero. |
| 06 | `06-sin-servicios-asignados.png` | El barbero elegido no tiene servicios activos asignados. |
| 07 | `07-error-servicios.png` | Falla la carga de servicios de ese barbero. |
| 08 | `08-formulario-completo.png` | Formulario completo, resumen cerrado antes del envío. |
| 09 | `09-error-validacion.png` | Envío con campos inválidos: error por campo más alerta global de validación. |
| 10 | `10-guardando.png` | Envío en curso: botón deshabilitado con «Guardando…», datos conservados. |
| 11 | `11-conflicto-horario.png` | Conflicto `409`: la hora recién se ocupó; el dato no se pierde y el error vive también junto al campo de hora. |
| 12 | `12-turno-registrado.png` | Éxito persistente: resumen honesto porque `HU-062` (agenda diaria como destino de navegación) no aplica a este enlace. |

## Contrato visual común

- Paleta aprobada (`apps/web/src/styles/tokens.css`): tinta `#101B2B`, marfil `#F4F0E7`, pergamino `#E8E2D8` (`--color-surface-muted`), blanco `#FFFFFF`, grafito `#2A2D32`, piedra `#C9C0B2` y latón `#B8955A`/`#765C2F`.
- **El formulario vive sobre el canvas de tinta, no sobre tarjetas de pergamino.** El pergamino está reservado para REGISTROS —las fichas y filas de turno de `panel-agenda-eventos`, y la ficha del turno recién creado del evento `12`—, no para el cromo de un formulario: campos blancos sobre pergamino sobre tinta apilaban tres materiales para una sola tarea.
- **Las cuatro secciones son bandas de UNA hoja, separadas por filete**, no cuatro tarjetas apiladas. La hoja es una superficie translúcida (`rgba(244,240,231,.035)`) con contorno tenue y filete de latón de `3 px` a la izquierda —la misma familia de material que el carril y el esqueleto de la agenda—, y cada banda se separa de la siguiente con una línea de `1 px`. Con borde propio por sección, el formulario se leía como una pila rayada de objetos sueltos en vez de un documento.
- **Los campos son el control ya resuelto en `panel-agenda-eventos`**: superficie translúcida, contorno tenue, filete inferior de `2 px`, radio `2 px` y rótulo en versalitas de latón. Un campo obligatorio agrega un asterisco en latón, nunca en rojo (el rojo se reserva para error). El estado deshabilitado (servicio sin barbero elegido o cargando) se atenúa y pierde el filete.
- **El latón del filete inferior marca el campo ya resuelto; uno vacío lleva filete neutro.** Ocho subrayados dorados a la vez convertían el formulario en un muestrario de oro y no distinguían lo hecho de lo pendiente: compara el evento `04` (todo neutro) con el `08` (todo en latón) y el `05` (solo el barbero). Es un estado real, derivable del valor del campo.
- **Cada banda se compone en dos columnas en escritorio** (rótulo numerado a la izquierda en `288 px`, campos a la derecha) y en una sola en móvil; el CTA se sangra hasta la columna de campos, para que la acción caiga bajo los controles y no bajo los títulos de sección. Con esa densidad el formulario completo entra en un viewport de `1024 px` de alto —el generador falla si deja de entrar— en vez de exigir scroll con media pantalla vacía al costado.
- **El número de sección es un cuadrado de `2 px` de radio con filete de latón y cifra en serif**, no un círculo: el círculo era el único elemento redondo de todo el sistema.
- **El resumen es el `<dl>` real del código**: título «Resumen» y solo cuatro entradas —Barbero, Servicio, Persona atendida, Fecha y hora—, cada una visible únicamente cuando su dato existe. Aparece en cuanto hay UNA selección hecha (`hasSummaryContent`), no solo con el formulario completo, así que también acompaña a los eventos `05`–`07` y `09`. En escritorio vive en una columna lateral de `340 px` junto al CTA; en móvil la retícula colapsa y queda entre el último campo y el CTA, que es donde `especificacion-frontend-nava.md` §7.3 lo pide. El nombre del barbero va con su monograma, igual que en la agenda.
- **El botón primario (`Registrar turno`) es sólido en latón** sobre el canvas de tinta; en `saving` se atenúa por opacidad y su rótulo cambia a «Guardando…», sin duplicar un segundo indicador de carga.
- **Los errores conviven en dos niveles cuando aplica**: un mensaje junto al campo afectado (filete inferior y texto en rojo) y la alerta global, que en escritorio encabeza la columna lateral y en móvil cae justo encima del CTA. El evento `11` muestra ambos a la vez porque así ocurre en el código real (`saveErrorDetail` global + el mismo dato conservado en el campo de hora).
- **Estados de página.** La espera inicial usa esqueleto con la geometría del formulario por llegar y un rótulo de estado en versalitas de latón, igual que el evento `05` de `panel-agenda-eventos`; el error de contexto y el vacío se componen centrados con divisor NAVA, rótulo de estado, titular en serif, cuerpo y acción real, y el destino nombrado en el mensaje (`"Barberos"`) se resalta en latón. Ninguno de los tres es ya un renglón suelto en una pantalla vacía.
- **El éxito (evento `12`) es la pantalla completa**, no una alerta pequeña flotando arriba a la izquierda: divisor, rótulo «Confirmación», titular en serif, ficha de pergamino con el turno creado y la acción para registrar otro.
- **Tres tintes de estado levantados al canvas de tinta** (`--danger-ink`, `--warning-ink`, `--success-ink`): los pares del sistema están calculados para superficies claras y `#A43A3A` sobre `#101B2B` no alcanza AA. Son el mismo matiz aclarado, y solo se usan para texto y filetes sobre tinta; las alertas y las fichas siguen usando los pares claros del sistema.
- **Los datos no se pierden en ningún estado de error.** Los eventos `09`, `10` y `11` parten del formulario ya lleno (evento `08`), nunca de un formulario vacío, porque el código real conserva `attendeeName`, `customerFullName`, etc. ante `409`/`422`/fallo recuperable.
- Los rótulos de sección, campo, placeholder y alerta son los del módulo real `apps/web/src/modules/agenda/pages/NewAppointmentPage.vue`. Cuando un mockup necesita un texto que el código todavía no tiene, procede de una decisión `DEC-*` registrada.
- **Los valores de fecha y hora son los que produce el código, no una tipografía libre**: el campo de fecha muestra lo que dibuja el control nativo `type="date"` en es-CO (`04/09/2026`), el de hora el valor de `type="time"` en 24 h (`11:30`) —que es el que el resumen concatena— y el resumen escribe el día completo con el mismo `Intl.DateTimeFormat` de `formatCivilDateFull`. El turno de ejemplo cae un día después del que muestra `panel-agenda-eventos`: registrar un turno en el pasado contradiría la pantalla.
- **El contador de la nota refleja el límite real** de `validateCustomerNote` (500 caracteres), no un tope inventado.
- Los datos mostrados (barbero, cliente, teléfono, correo, precio) son ficticios.
- Ninguna línea de texto se sale de su contenedor en ningún viewport y ningún evento de escritorio excede `1024 px` de alto: el generador falla con código distinto de cero en ambos casos.

## Uso por agentes de implementación

1. Seleccionar un solo evento y viewport.
2. Abrir únicamente el PNG correspondiente y este README; no usar la lámina compuesta como objetivo.
3. Comparar la app real y el PNG en el mismo viewport efectivo.
4. Corregir por ciclos cortos y guardar captura, comparación lado a lado y overlay o diff.
5. Repetir el mismo número en el viewport hermano antes de declarar el evento terminado.

El mockup gobierna composición, color, proporción, jerarquía, escala de controles e iconos, centrado, alineación, densidad y estado visual. Las HU, reglas, decisiones, contrato API y código real gobiernan el comportamiento y el texto cuando exista una discrepancia funcional. `idempotency-conflict`, `not-found`, `network-error` y `unexpected-error` (`SaveStatus` real) comparten el mismo tratamiento visual de alerta global sin resaltado de campo que se ve en el evento `09`; no se declaran como eventos aparte porque no aportan una composición distinta.

## Procedencia

Los 24 archivos se generan con [`tools/mockups/nuevo-turno-eventos`](../../../../../tools/mockups/nuevo-turno-eventos), con el mismo criterio que `tools/mockups/panel-agenda-eventos` y `tools/mockups/auth-eventos`: `content.mjs` declara cada evento una sola vez y `render.mjs` lo compone en HTML con los tokens de `apps/web/src/styles/tokens.css` y las fuentes auto-hosteadas del proyecto, y lo rasteriza con el Chromium de Playwright que ya usa la suite e2e.

```bash
node tools/mockups/nuevo-turno-eventos/render.mjs
```

Un evento se declara una vez y se dibuja en los dos viewports, así que escritorio y móvil no pueden divergir en copy, orden ni estado. El generador aborta con código distinto de cero si algún evento desborda el ancho de su viewport.

Este directorio es evidencia de diseño para el issue [#190](https://github.com/bcaceres19/barberia/issues/190) (`chore/190-nuevo-turno-fidelidad`, prompt en [`docs/10-backlog/prompts/chore/issue-190-nuevo-turno-fidelidad.md`](../../../prompts/chore/issue-190-nuevo-turno-fidelidad.md)). No modifica `apps/web` ni autoriza implementación automática de conceptos no representados aquí.
