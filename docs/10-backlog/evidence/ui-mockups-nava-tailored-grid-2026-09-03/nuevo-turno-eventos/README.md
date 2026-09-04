# Mockups de `/panel/turnos/nuevo` (Nuevo turno) por viewport y evento

## Propósito

Este directorio reemplaza como referencia de implementación la lámina compuesta `03-nuevo-turno.png` de `ui-mockups-nava-tailored-grid-2026-09-02`. La lámina original se conserva como historial visual, pero no debe entregarse a un agente de implementación: cada archivo de este directorio representa una sola pantalla y un solo evento, igual que ocurrió en `/acceso` con `02-acceso-recuperacion.png` y en `/panel` con `01-agenda-diaria-responsive.png`.

La lámina compuesta mostraba, en su tercer panel, botones de cuenta decorativos y navegación duplicada en el encabezado (`Nuevo turno` y `Cerrar sesión` como botones de la barra superior, repetidos luego en el dock). Esa es la misma inconsistencia que ya se corrigió al construir [`auth-eventos`](../auth-eventos/README.md) y [`panel-agenda-eventos`](../panel-agenda-eventos/README.md): el encabezado y el dock de este directorio son el mismo cascarón ya validado en esos dos atlas, sin controles de cuenta decorativos.

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
| 04 | `04-formulario-vacio.png` | Formulario recién abierto: nada elegido, sin resumen. |
| 05 | `05-cargando-servicios.png` | Barbero elegido, servicios cargando: el campo de servicio queda deshabilitado. |
| 06 | `06-sin-servicios-asignados.png` | El barbero elegido no tiene servicios activos asignados. |
| 07 | `07-error-servicios.png` | Falla la carga de servicios de ese barbero. |
| 08 | `08-formulario-completo.png` | Formulario completo, resumen visible antes del envío. |
| 09 | `09-error-validacion.png` | Envío con campos inválidos: error por campo más alerta global de validación. |
| 10 | `10-guardando.png` | Envío en curso: botón deshabilitado con «Guardando…», datos conservados. |
| 11 | `11-conflicto-horario.png` | Conflicto `409`: la hora recién se ocupó; el dato no se pierde y el error vive también junto al campo de hora. |
| 12 | `12-turno-registrado.png` | Éxito persistente: resumen honesto porque `HU-062` (agenda diaria como destino de navegación) no aplica a este enlace. |

## Contrato visual común

- Paleta aprobada (`apps/web/src/styles/tokens.css`): tinta `#101B2B`, marfil `#F4F0E7`, pergamino `#E8E2D8` (`--color-surface-muted`), blanco `#FFFFFF`, grafito `#2A2D32`, piedra `#C9C0B2` y latón `#B8955A`/`#765C2F`.
- **Las tarjetas numeradas del formulario usan pergamino, no blanco puro**, igual que las fichas de turno de `panel-agenda-eventos`: el mismo canvas de tinta se lee mejor con superficies claras atenuadas que con blanco puro.
- **Cada sección del formulario lleva un número en un círculo con filete de latón**, tomado literalmente de la lámina compuesta original; el título de la sección va en serif y el subtítulo en grafito secundario.
- **Los campos de texto son superficies blancas dentro de la tarjeta de pergamino**, con borde piedra y radio pequeño; un campo obligatorio agrega un asterisco en latón oscuro, nunca en rojo (el rojo se reserva para error).
- **El resumen del turno es una superficie de tinta** (`--ink-2`), no una tarjeta clara: aparece solo cuando hay contenido que resumir (`hasSummaryContent` del código real) y antes del CTA, igual que en la lámina original.
- **El botón primario (`Registrar turno`) es sólido en latón** sobre el canvas de tinta; en `saving` se atenúa por opacidad y su rótulo cambia a «Guardando…», sin duplicar un segundo indicador de carga.
- **Los errores conviven en dos niveles cuando aplica**: un mensaje junto al campo afectado (borde y texto en rojo) y una alerta global arriba del CTA. El evento `11` muestra ambos a la vez porque así ocurre en el código real (`saveErrorDetail` global + el mismo dato conservado en el campo de hora).
- **Los datos no se pierden en ningún estado de error.** Los eventos `09`, `10` y `11` parten del formulario ya lleno (evento `08`), nunca de un formulario vacío, porque el código real conserva `attendeeName`, `customerFullName`, etc. ante `409`/`422`/fallo recuperable.
- Los rótulos de sección, campo, placeholder y alerta son los del módulo real `apps/web/src/modules/agenda/pages/NewAppointmentPage.vue`. Cuando un mockup necesita un texto que el código todavía no tiene, procede de una decisión `DEC-*` registrada.
- Los datos mostrados (barbero, cliente, teléfono, correo, precio) son ficticios.
- Ninguna línea de texto se sale de su contenedor en ningún viewport: el generador falla si detecta desborde horizontal.

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
