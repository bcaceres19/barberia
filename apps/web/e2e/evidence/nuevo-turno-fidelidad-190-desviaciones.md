# Desviaciones de `/panel/turnos/nuevo` frente al atlas `nuevo-turno-eventos`

Issue [#190](https://github.com/bcaceres19/barberia/issues/190) · rama
`chore/190-nuevo-turno-fidelidad` · atlas de referencia:
[`nuevo-turno-eventos`](../../../../docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/nuevo-turno-eventos/README.md)
(revisión del 2026-09-04, segundo pase del formulario).

Criterio del atlas: **el mockup gobierna composición, color, proporción,
jerarquía, escala, centrado, alineación, densidad y estado visual; las HU,
reglas, decisiones, contrato API y código real gobiernan el comportamiento y el
texto.** Lo que sigue es todo lo que no se pudo igualar, con su autoridad.

## Evidencia que sustenta esta tabla

- Capturas de la app por evento y viewport:
  `e2e/evidence/nuevo-turno/fidelidad-190/mock/{desktop,mobile}/`
  (`nuevo-turno-fidelidad-mock.spec.ts`).
- Comparación contra cada lámina:
  `e2e/evidence/nuevo-turno/fidelidad-190/comparacion/`, con el diff absoluto
  por evento y `resumen.txt` (porcentaje de píxeles distintos y diferencia de
  alto). Las láminas lado a lado no se versionan por peso: se recomponen con
  `node e2e/nuevo-turno-fidelidad-comparacion.mjs`, que las genera junto al
  diff a partir del atlas y de las capturas de `mock/`.
- Evidencia responsive, teclado y movimiento reducido:
  `e2e/evidence/nuevo-turno/fidelidad-190/responsiva/`
  (`nuevo-turno-evidencia-responsiva.spec.ts`).

## Criterio | Estado | Evidencia

Una fila por evento y viewport. «Píxeles distintos» es el porcentaje que
reporta `resumen.txt` sobre la región común, con umbral de 16/255 para no
contar el antialias del texto; en escritorio incluye el desplazamiento
uniforme de ~13 px que introduce el encabezado real (ver la tabla siguiente),
y en móvil se añade la diferencia de alto total entre lámina y app.

| Evento                     | Viewport           | Estado                           | Evidencia                                                       |
| -------------------------- | ------------------ | -------------------------------- | --------------------------------------------------------------- |
| 01 carga de contexto       | desktop 1440×1024  | Conforme                         | 5,7 % · `comparacion/desktop/01-carga-contexto-diff.png`        |
| 01 carga de contexto       | mobile 420×935 @2x | Conforme                         | 12,0 % · alto 2590 vs 2558 (+32 px)                             |
| 02 error de contexto       | desktop            | Conforme                         | 2,5 %                                                           |
| 02 error de contexto       | mobile             | Conforme                         | 6,6 % · mismo alto                                              |
| 03 sin barberos            | desktop            | Conforme                         | 1,8 %                                                           |
| 03 sin barberos            | mobile             | Conforme                         | 4,8 % · mismo alto                                              |
| 04 formulario vacío        | desktop            | Conforme                         | 6,5 % · fija la geometría de los demás                          |
| 04 formulario vacío        | mobile             | Conforme                         | 13,3 % · alto 2876 vs 2824 (+52 px)                             |
| 05 cargando servicios      | desktop            | Conforme                         | 6,9 % · servicio atenuado y sin latón                           |
| 05 cargando servicios      | mobile             | Conforme                         | 12,9 % · alto 3164 vs 3112 (+52 px)                             |
| 06 sin servicios asignados | desktop            | Conforme                         | 7,1 %                                                           |
| 06 sin servicios asignados | mobile             | Conforme                         | 13,1 % · alto 3164 vs 3112 (+52 px)                             |
| 07 error de servicios      | desktop            | Conforme                         | 7,1 %                                                           |
| 07 error de servicios      | mobile             | Conforme                         | 13,1 % · alto 3164 vs 3112 (+52 px)                             |
| 08 formulario completo     | desktop            | Conforme                         | 7,6 % · resumen lateral de 340 px, ocho filetes en latón        |
| 08 formulario completo     | mobile             | Conforme                         | 12,9 % · alto 3438 vs 3386 (+52 px); resumen entre campos y CTA |
| 09 error de validación     | desktop            | Conforme con desviación de texto | 8,7 % · alerta global sin segunda línea (ver tabla siguiente)   |
| 09 error de validación     | mobile             | Conforme con desviación de texto | 13,6 % · alto 3506 vs 3486 (+20 px)                             |
| 10 guardando               | desktop            | Conforme                         | 7,6 % · CTA atenuado con «Guardando…», sin segundo indicador    |
| 10 guardando               | mobile             | Conforme                         | 12,9 % · alto 3438 vs 3386 (+52 px)                             |
| 11 conflicto de horario    | desktop            | Conforme con desviación de texto | 8,8 % · alerta global + error junto al campo de hora            |
| 11 conflicto de horario    | mobile             | Conforme con desviación de texto | 13,6 % · alto 3664 vs 3644 (+20 px)                             |
| 12 turno registrado        | desktop            | Conforme con desviación de dato  | 4,3 % · importe crudo del API (ver tabla siguiente)             |
| 12 turno registrado        | mobile             | Conforme con desviación de dato  | 13,3 % · mismo alto                                             |

Guardas automáticas que respaldan la densidad, verificadas por
`nuevo-turno-fidelidad-mock.spec.ts` en cada uno de los 24 casos: el documento
no desborda a lo ancho, y en escritorio los doce eventos entran en `1024 px`
de alto sin exigir scroll dentro del cascarón. Es la misma condición con la
que aborta el generador del atlas.

## Diferencia | Autoridad | Tratamiento

| Diferencia                                                                                                                                                                                              | Autoridad                                                                                                                              | Tratamiento                                                                                                                                                         |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| El encabezado real lleva el botón «Cerrar sesión» a la derecha; el atlas dibuja solo marca y nombre de barbería.                                                                                        | Código (`AppHeader`, HU-012; el prompt de #190 lo protege de forma explícita: «no se eliminan porque un PNG estático no los muestre»). | Se conserva. Es cascarón compartido, fuera del alcance de esta pantalla.                                                                                            |
| Todo el contenido queda ~13 px más abajo que en la lámina.                                                                                                                                              | Código (el `AppHeader` real mide más alto que la barra del generador).                                                                 | Se conserva. El diff de escritorio muestra que las columnas, anchos y límites de banda coinciden en horizontal; el desplazamiento es uniforme y viene del cascarón. |
| El control nativo `type="time"` dibuja reloj de 12 h («11:30 a. m.»); el atlas lo dibuja en 24 h («11:30»).                                                                                             | Navegador (Chromium resuelve el formato por el locale del usuario, y `es-CO` es de 12 h).                                              | Se conserva el control nativo. El resumen sí escribe el valor de 24 h que el atlas muestra, porque ahí el texto lo compone el código.                               |
| El marcador de posición de `type="date"` aparece como «dd/mm/yyyy»; el atlas escribe «dd/mm/aaaa».                                                                                                      | Navegador (mismo motivo).                                                                                                              | Se conserva.                                                                                                                                                        |
| La ficha del éxito escribe «45000.00 COP»; el atlas escribe «$45.000 COP».                                                                                                                              | Código (`created.priceAmount` es el importe crudo del API; HU-061 no formatea moneda y no existe decisión que lo autorice).            | Se conserva el valor real. Formatear moneda es trabajo de producto, no de fidelidad visual.                                                                         |
| La alerta global de conflicto muestra el `detail` del servidor («el barbero ya tiene una cita en ese intervalo»); el atlas escribe «Ese horario acaba de ocuparse» + «Elige otra hora para continuar.». | Código (DEC-073: el detalle del `409` es del servidor).                                                                                | Se conserva el texto real. La composición —rótulo ERROR, línea destacada y cuerpo— sí se reproduce.                                                                 |
| La alerta de validación no lleva la segunda línea «Corrige los campos marcados para continuar.».                                                                                                        | Código (`saveStatus === 'validation-error'` solo produce una cadena; ese cuerpo no existe en el módulo ni en una `DEC-*`).             | Se conserva una sola línea.                                                                                                                                         |
| El atlas escribe el teléfono con espacios («+57 300 123 4567»); la evidencia usa «+573001234567».                                                                                                       | Código (`validateCustomerPhone` exige `^\+[1-9][0-9]{7,14}$`: con espacios el envío del evento 11 nunca habría llegado al `409`).      | La evidencia usa la forma que el propio placeholder del campo declara.                                                                                              |
| El evento 12 conserva la línea «Horas en la zona horaria de la barbería: America/Bogota»; el atlas no la dibuja.                                                                                        | Código (`barbershopTimezone` sigue resuelto después de crear el turno).                                                                | Se conserva: es información cierta y el atlas no la contradice, solo no la representó.                                                                              |
| El caret de los dos selectores es la punta de flecha ya usada por `BarberSelect` en `/panel`; el atlas dibuja un triángulo «▾».                                                                         | Sistema (issue #189 ya resolvió esa forma para el mismo control sobre tinta).                                                          | Se conserva la forma del sistema, idéntica en los dos campos de la banda 1.                                                                                         |
| El divisor de los estados de página es la regla de 2 px con rombo de `PageState`; el atlas compone dos filetes de 1 px con hueco alrededor del rombo.                                                   | Sistema (`PageState`, construcción fijada en el issue #189 y compartida con `/panel` y `/acceso`).                                     | Se conserva la construcción del sistema; solo se iguala el ancho (220 px escritorio, 166 px móvil).                                                                 |
| El titular de la alerta global no es un `<h4>` como en el resto de la app.                                                                                                                              | Accesibilidad (esta pantalla ya tiene un `<h2>` por banda: un `<h4>` de alerta rompe `heading-order` en axe).                          | Se compone como línea destacada dentro del cuerpo de `BaseAlert`, con el mismo peso y tamaño del atlas.                                                             |

## Decisiones que sí cambian código, y por qué

| Cambio                                                                                                                    | Motivo                                                                                                                                                                                                                                                                                                                    |
| ------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| El campo «Barbero» pasa de `<select>` nativo a `BarberSelect` (listbox con monograma).                                    | El atlas dibuja el retrato del barbero DENTRO del control cerrado (eventos 05–11) y una `<option>` nativa no puede llevarlo. `BarberSelect` ya existía en el módulo desde #189; solo se le agregaron `triggerId` y `placeholder` opcionales, con los valores actuales por defecto, para no tocar `/panel`.                |
| El campo «Nota» pasa de `BaseInput` de una línea a `<textarea>` local con contador.                                       | El atlas lo dibuja como caja multilínea con contador `0/500`. No se toca `shared/ui`: `BaseInput` no tiene variante multilínea y añadírsela habría afectado a todas las pantallas. No se fija `maxlength`, para que un exceso siga produciendo el mensaje real de `validateCustomerNote` en vez de truncarse en silencio. |
| Un `409` marca además el campo «Hora del turno» con el `detail` del servidor.                                             | El evento 11 del atlas muestra los dos niveles de error a la vez. No se inventa texto: es el mismo `detail` que ya alimenta la alerta global.                                                                                                                                                                             |
| Tres tokens nuevos en `tokens.css`: `--color-danger-on-strong`, `--color-warning-on-strong`, `--color-success-on-strong`. | El atlas los declara en su contrato visual («tres tintes de estado levantados al canvas de tinta»): `#a43a3a` sobre `#101b2b` no alcanza AA. Son aditivos —ningún estilo previo los referencia—, así que no pueden regresionar otra pantalla.                                                                             |
| `--page-state-mark` se sobrescribe con `!important` solo dentro de esta pantalla.                                         | `PageState` fija ese color por estilo en línea. Se levanta a los tintes sobre tinta sin tocar el componente, para no alterar `/panel`, que ya se validó contra su propio atlas.                                                                                                                                           |
