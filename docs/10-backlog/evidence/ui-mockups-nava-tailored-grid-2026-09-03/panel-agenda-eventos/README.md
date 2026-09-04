# Mockups de `/panel` (agenda diaria) por viewport y evento

## Propósito

Este directorio reemplaza como referencia de implementación la lámina compuesta `01-agenda-diaria-responsive.png` de `ui-mockups-nava-tailored-grid-2026-09-02`. La lámina original se conserva como historial visual, pero no debe entregarse a un agente de implementación: cada archivo de este directorio representa una sola pantalla y un solo evento para reducir ambigüedad, igual que ocurrió en `/acceso` con `02-acceso-recuperacion.png`.

La estructura es deliberadamente simétrica:

- `desktop/panel/` y `mobile/panel/` contienen los mismos doce eventos de `/panel`;
- un mismo número y nombre describen el mismo estado funcional en ambos viewports; termina el par antes de dar el evento por cerrado;
- encabezado, filtros, navegación por fecha, lista y dock son los mismos componentes en escritorio y móvil — móvil no introduce encabezado propio ni una segunda forma de navegar, y agrupa los destinos secundarios en «Más» sin sacarlos del árbol.

## Panel (agenda diaria)

| N.º | Archivo en cada viewport | Evento representado |
| --- | --- | --- |
| 01 | `01-agenda-lista.png` | Agenda de hoy cargada: línea temporal decorativa y lista accesible equivalente, ambas dibujadas. |
| 02 | `02-carga-contexto.png` | Carga inicial: barbería, barberos y zona horaria aún sin resolver, sin filtros, fecha ni acción. |
| 03 | `03-error-contexto.png` | Fallo al cargar el contexto inicial: sin contexto resuelto tampoco hay CTA. |
| 04 | `04-sin-barberos.png` | Sin barberos activos: no se inventa selección ni se ofrece un turno que no podría asignarse. |
| 05 | `05-carga-agenda.png` | Primera carga de la agenda con barbero y fecha ya resueltos: esqueleto que conserva la geometría del contenido por llegar. |
| 06 | `06-actualizando-fecha.png` | Cambio de fecha: agenda anterior atenuada, actualización anunciada, marcador «Ahora» ausente porque la fecha mostrada ya no es hoy. |
| 07 | `07-error-agenda.png` | Error recuperable de la agenda: barbero y fecha se conservan para reintentar sin volver a elegirlos. |
| 08 | `08-barbero-no-disponible.png` | El barbero solicitado ya no está disponible: la selección queda vacía porque conservarla contradiría el mensaje. |
| 09 | `09-dia-sin-turnos.png` | Día válido sin turnos: se conserva el eje del día vacío con el marcador «Ahora», y «Nuevo turno» vive una sola vez, dentro del estado vacío. |
| 10 | `10-zona-horaria-no-disponible.png` | Zona horaria no confirmada: la agenda sigue visible y solo se bloquea la navegación por fecha. |
| 11 | `11-turno-nocturno.png` | Turno que cruza medianoche: el eje llega hasta las 02:00 y la marca «Cambio de día» cae exactamente sobre las 00:00. |
| 12 | `12-seleccion-barbero.png` | Selección de barbero: la lista desplegada muestra retrato con foto y con monograma conviviendo. |

## Contrato visual común

- Paleta aprobada (`apps/web/src/styles/tokens.css`): tinta `#101B2B`, marfil `#F4F0E7`, pergamino `#E8E2D8` (`--color-surface-muted`), blanco `#FFFFFF`, grafito `#2A2D32`, piedra `#C9C0B2` y latón `#B8955A`/`#765C2F`.
- **Las fichas y filas usan pergamino, no blanco puro.** Sobre tinta el blanco daba ~17:1 y deslumbra en una pantalla de uso continuo.
- **Los turnos terminales (`completed`, `cancelled_by_customer`, `cancelled_by_barber`, `no_show`) cambian de material, no de peso.** Un turno vigente (`confirmed`) es superficie de papel; uno cerrado es un registro con contorno sobre la tinta. Un relleno gris los dejaba pesando igual o más que un turno vigente.
- Las insignias de estado sobre pergamino se apoyan en contorno y texto del color de estado: `confirmed` usa el par info, `completed` el par éxito, `cancelled_by_customer` el par inactivo, `cancelled_by_barber` el par peligro y `no_show` el par advertencia.
- **Calendario.** Cada ficha se posiciona y dimensiona desde su propia hora contra el eje: el borde izquierdo y el ancho salen de `startsAt`/`endsAt`, nunca de un reparto arbitrario. Las etiquetas del eje y las marcas (`Ahora`, `Cambio de día`) viven en bandas separadas y nunca se superponen; cada marca traza su línea vertical por el carril, y la de medianoche cae exactamente sobre `00:00`. El carril es una región definida con guía por hora y media hora tenue entre cada par. La ficha muestra rango horario, persona y servicio cuando la duración le da ancho; con poco ancho conserva solo la hora de inicio.
- **Estados de página.** Cuando la carga, el error o el vacío ocupan toda el área de contenido, se componen centrados con `PageState`: indicador o divisor NAVA, rótulo de estado en versalitas, titular en serif, cuerpo y acción real. La alerta como nota al margen (`BaseAlert`) se reserva para cuando acompaña contenido que sigue visible (evento `10`). El texto de espera y el de vacío se componen como titular en serif; `Reintentar` es un botón real, no una etiqueta contorneada. Con barbero y fecha ya resueltos, la espera usa esqueleto con la geometría del contenido por llegar (evento `05`); sin contexto resuelto se usa el estado centrado (evento `02`). Si el mensaje nombra otra sección del producto, ese destino se resalta en latón (evento `04`, «Barberos»).
- **La línea temporal es decorativa** (`aria-hidden`, `tabindex="-1"`) y la lista es la representación accesible equivalente. Las dos se conservan en los dos viewports.
- **El marcador «Ahora» es contextual y decorativo:** no cambia `confirmed` ni crea disponibilidad, y solo existe cuando la fecha mostrada es hoy.
- **Retrato del barbero.** Dondequiera que un barbero se ve o se elige, su nombre va con un retrato cuadrado de radio `2 px` y filete de latón: `28 px` en el selector cerrado y `34 px` en cada opción de la lista (`26`/`32 px` en móvil). Nunca aparece un retrato sin nombre, y es decorativo para lectores de pantalla porque el nombre ya da la identidad. Esta entrega implementa solo el monograma derivado de `fullName`; la variante con fotografía (evento `12`, mezclada deliberadamente con monograma en el mismo plantel) queda representada en el mockup pero no se implementa — `Barber` no declara un campo de foto.
- Los rótulos de acción, los cinco rótulos de estado (`Confirmado`, `Completado`, `Cancelado por el cliente`, `Cancelado por el barbero`, `No se presentó`) y la línea `{fecha} · Zona {timezone}` son los del módulo real `apps/web/src/modules/agenda`. Cuando un mockup necesita un texto que el código todavía no tiene, procede de una decisión `DEC-*` registrada.
- Los datos mostrados (barberos, clientes, servicios, horarios) son ficticios.
- Ninguna línea de texto se sale de su contenedor en ningún viewport: el generador falla si detecta desborde horizontal o si un evento de escritorio no cabe en `1024 px` de alto.

## Uso por agentes de implementación

1. Seleccionar un solo evento y viewport.
2. Abrir únicamente el PNG correspondiente y este README; no usar la lámina compuesta como objetivo.
3. Comparar la app real y el PNG en el mismo viewport efectivo.
4. Corregir por ciclos cortos y guardar captura, comparación lado a lado y overlay o diff.
5. Repetir el mismo número en el viewport hermano antes de declarar el evento terminado.

El mockup gobierna composición, color, proporción, jerarquía, escala de controles e iconos, centrado, alineación, densidad y estado visual. Las HU, reglas, decisiones, contrato API y código real gobiernan el comportamiento y el texto cuando exista una discrepancia funcional.

## Procedencia

Los 24 archivos se generan con [`tools/mockups/panel-agenda-eventos`](../../../../../tools/mockups/panel-agenda-eventos), con el mismo criterio que `tools/mockups/auth-eventos`: `content.mjs` declara cada evento una sola vez y `render.mjs` lo compone en HTML con los tokens de `apps/web/src/styles/tokens.css` y las fuentes auto-hosteadas del proyecto, y lo rasteriza con el Chromium de Playwright que ya usa la suite e2e.

```bash
node tools/mockups/panel-agenda-eventos/render.mjs
```

Un evento se declara una vez y se dibuja en los dos viewports, así que escritorio y móvil no pueden divergir en copy, orden ni estado. La posición y el ancho de cada ficha en la línea temporal se calculan desde su propia hora (`start`/`end` en minutos desde medianoche), de modo que una ficha no puede quedar desalineada de su horario ni una marca puede taparle la etiqueta a una hora del eje. El generador aborta con código distinto de cero si algún evento desborda.
