# Mockups de `/panel/turnos/:appointmentId` (detalle, historial y reprogramación) por viewport y evento

## Propósito

Este directorio reemplaza como referencia de implementación la lámina compuesta
[`04-detalle-historial-reprogramacion.png`](../../ui-mockups-nava-tailored-grid-2026-09-02/04-detalle-historial-reprogramacion.png)
de `ui-mockups-nava-tailored-grid-2026-09-02`. La lámina se conserva como
historial visual, pero no debe entregarse a un agente de implementación: cada
archivo de este directorio representa una sola pantalla y un solo evento, igual
que ya ocurrió en `/acceso` con [`auth-eventos`](../auth-eventos/README.md), en
`/panel` con [`panel-agenda-eventos`](../panel-agenda-eventos/README.md) y en
`/panel/turnos/nuevo` con [`nuevo-turno-eventos`](../nuevo-turno-eventos/README.md).

La estructura es deliberadamente simétrica:

- `desktop/detalle-turno/` y `mobile/detalle-turno/` contienen los mismos
  dieciséis eventos;
- un mismo número y nombre describen el mismo estado funcional en ambos
  viewports;
- encabezado y dock son los mismos componentes ya validados en los tres atlas
  hermanos.

El turno representado es el mismo Mateo Rojas de las 09:00 que la agenda de
`panel-agenda-eventos` muestra en la fila de Julián Rodríguez: quien llega aquí
lo hace tocando esa fila, y «Volver a la agenda» regresa a esa fecha y a ese
barbero (HU-063).

## Detalle del turno (`/panel/turnos/:appointmentId`)

| N.º | Archivo en cada viewport | Evento representado |
| --- | --- | --- |
| 01 | `01-carga-detalle.png` | `pageStatus === 'loading'`: esqueleto con la geometría de la ficha y del historial por llegar. |
| 02 | `02-turno-no-disponible.png` | `not-found`: el turno se eliminó o es de otra barbería (RN-TEN-01, un solo mensaje sin distinguir la causa). |
| 03 | `03-error-detalle.png` | `error` recuperable al cargar el detalle, con «Reintentar». |
| 04 | `04-detalle-confirmado.png` | Referencia de la pantalla: turno `confirmed` con su historial completo. |
| 05 | `05-historial-cargando.png` | Ficha resuelta y `historyStatus === 'loading'`: son dos peticiones distintas y la ficha no espera al historial. |
| 06 | `06-historial-vacio.png` | Historial vacío: «Sin eventos registrados todavía.» |
| 07 | `07-error-historial.png` | Falla solo el historial: la ficha sigue completa y el reintento vive en su columna. |
| 08 | `08-historial-paginado.png` | Primera página del historial con cursor pendiente: el eje continúa hasta «Cargar más». |
| 09 | `09-turno-completado.png` | Terminal `completed`: sin acción propia, ficha en contorno. |
| 10 | `10-turno-cancelado.png` | Terminal `cancelled_by_customer`, con el `reason` del evento. |
| 11 | `11-dialogo-reprogramar.png` | Diálogo recién abierto: fecha y hora precargadas con el horario vigente. |
| 12 | `12-dialogo-guardando.png` | `saving`: campos deshabilitados y «Confirmar» bloqueado, sin perder lo elegido. |
| 13 | `13-dialogo-conflicto-agenda.png` | `conflict` (409): el `detail` real del backend dentro de la alerta. |
| 14 | `14-dialogo-version-obsoleta.png` | `version-conflict`: la única salida es «Recargar», así que la alerta lleva su acción. |
| 15 | `15-dialogo-sin-conexion.png` | `network-error`: lo elegido se conserva, como dice el propio mensaje. |
| 16 | `16-reprogramacion-aplicada.png` | Éxito: diálogo cerrado, hora vigente actualizada y un evento más en el historial. |

Cuatro estados reales no tienen archivo propio porque no aportan una
composición distinta, solo su copy:

- `invalid-state` («Este turno ya no se puede reprogramar» / «Su estado cambió
  mientras lo editabas.») usa exactamente el tratamiento del evento `14`,
  alerta de atención con acción «Recargar»;
- `validation-error`, `idempotency-conflict` y `unexpected-error` usan el del
  evento `13`, alerta de error sin acción, con sus títulos y cuerpos reales.

Lo mismo vale para los otros dos estados terminales: `cancelled_by_barber` y
`no_show` se componen como los eventos `09`/`10` y solo cambian la variante de
la insignia (`danger` y `warning`).

## Contrato visual común

- Paleta aprobada (`apps/web/src/styles/tokens.css`): tinta `#101B2B`, marfil
  `#F4F0E7`, pergamino `#E8E2D8` (`--color-surface-muted`), grafito `#2A2D32`,
  piedra `#C9C0B2` y latón `#B8955A`/`#765C2F`.
- **La ficha del turno es PERGAMINO; el historial se queda sobre tinta.** El
  pergamino es el material de los REGISTROS —las fichas y filas de turno de
  `panel-agenda-eventos`—, así que quien toca una fila de papel en `/panel`
  abre aquí esa misma fila en grande. Con las dos mitades sobre el mismo
  material, el registro y el libro de eventos habrían competido; con esta
  separación no hace falta ningún rótulo que explique cuál es cuál.
- **Un turno TERMINAL pierde el papel y pasa a contorno sobre tinta**, con la
  misma regla que su fila en la agenda: un turno vigente es una superficie
  presente, uno cerrado es un registro. Eso explica visualmente, sin texto, por
  qué el encabezado se queda sin la acción de reprogramar (eventos `09` y
  `10`); `canReschedule` solo es cierto con `confirmed`.
- **La ficha se compone como ficha, no como tabla.** Son los mismos siete
  rótulos del `<dl>` y en el mismo orden del template, pero con jerarquía:
  «Hora» encabeza en serif a tamaño de titular —es el primer `<dt>` del código
  y el dato por el que se abre esta pantalla—, los cinco hechos siguientes caen
  en dos columnas siguiendo el orden de lectura, y la nota del cliente cierra en
  serif, porque es una cita textual de una persona y no un campo más de la
  retícula. Siete renglones del mismo peso no tenían foco y dejaban media hoja
  de pergamino vacía a la derecha de una columna fija de rótulos.
- **El rótulo de cada hecho va en versalitas de latón oscuro** (`#765C2F` sobre
  pergamino, 4,85:1, AA para texto normal) y el valor en grafito. En móvil la
  retícula de dos columnas colapsa a una: en `420 px` no queda medida para el
  nombre del servicio ni para el contacto.
- **La zona horaria baja a su propia línea** bajo el horario. El template la
  anexa dentro del mismo `dd` como «· Zona X»; al quedar en línea aparte, el
  punto medio que la separaba sobra. Es el único cambio de puntuación del atlas.
- **No hay filete entre la ficha y el historial.** Papel y tinta ya son dos
  materiales distintos y separan las columnas por sí solos; una línea vertical
  entre ellas solo subrayaba el espacio libre bajo la ficha. Lo que ancla la
  columna del historial es el filete de latón bajo su título, el mismo remate
  editorial del wordmark y del divisor.
- **La insignia de estado acompaña al nombre en su misma línea de base**, como
  en el `<header>` real (título, insignia y acción en una fila); colgada debajo
  del titular quedaba como una etiqueta suelta. En móvil baja a su propia línea.
- **El eje del historial usa el rombo del divisor NAVA, no un punto redondo.**
  El código actual dibuja `border-radius: 999px`, y el círculo era el único
  elemento redondo de todo el sistema. El rombo SÓLIDO marca el evento más
  reciente ya cargado y solo cuando no queda cursor pendiente: en el evento
  `08` todos los rombos van huecos y el eje continúa, con un rombo punteado,
  hasta «Cargar más», porque el registro sigue.
- **Los eventos van del más antiguo al más reciente** (`ORDER BY occurred_at, id`
  en la API) y los campos modificados de cada evento en orden alfabético de
  `field_name` (`ORDER BY history_id, field_name`). Por eso «Hora de fin»
  aparece ANTES que «Hora de inicio»: es el orden real, no un error del atlas.
- **El nombre del evento y su sello comparten renglón**, como el
  `__history-main` real (`justify-content: space-between`): apilarlos alargaba
  cada evento sin aportar jerarquía. En móvil el sello baja bajo el nombre.
- **Cada campo modificado se compone con el rótulo en versalitas y los dos
  valores debajo**, separados por la flecha en latón. El texto conserva lo que
  dibuja el código —`historyFieldLabel(fieldName)`, valor anterior, `→`, valor
  nuevo—; solo cambia la disposición, para que una línea larga no se parta en
  medio de una fecha. **El rótulo del campo va en gris, no en latón**: con
  cuatro rótulos dorados por evento el historial se volvía un muestrario de oro
  y el latón dejaba de señalar nada. Aquí solo la flecha —el hecho del
  cambio— es de latón.
- **Sobre tinta la insignia de estado usa el par CLARO SÓLIDO del sistema**
  (`--*-surface` + `--*-text` + `--*-border`), no el tinte al 8 % que la agenda
  usa sobre pergamino: `#23405B` sobre `#101B2B` no alcanza AA. El punto de
  `BaseBadge` se dibuja como rombo, por la misma razón que el eje.
- **El diálogo de reprogramación es un formulario, así que vive sobre tinta**
  (`--ink-2` con filete superior de latón), nunca sobre pergamino, y sus dos
  campos son el control ya resuelto en los atlas hermanos: superficie
  translúcida, filete inferior de latón cuando el campo tiene valor, rótulo en
  versalitas y asterisco de obligatorio en latón, nunca en rojo. En `saving`
  los campos se atenúan y pierden el filete.
- **El horario vigente se enmarca dentro del diálogo** con filete de latón a la
  izquierda: es el contexto contra el que se decide la nueva hora, no un
  renglón gris más entre la alerta y los campos.
- **Dentro del diálogo la acción secundaria («Cancelar») es de contorno, no de
  marfil sólido.** Sobre `--ink-2`, un botón de canvas junto al latón sólido
  pesaba igual que la acción primaria; el contorno es el mismo `btn--ghost` que
  la agenda usa en su navegación por fecha.
- **Los estados de página completos** (`02` y `03`) se componen centrados con el
  divisor NAVA, rótulo de estado en versalitas, titular en serif y cuerpo, igual
  que en los atlas hermanos. El 404 no ofrece acción porque el código tampoco la
  ofrece: la salida es el enlace «Volver a la agenda» del encabezado.
- **El éxito no es una alerta.** Al confirmar, el código cierra el diálogo y
  recarga el detalle; no emite ningún mensaje de confirmación. El evento `16`
  representa exactamente eso —hora vigente nueva y un evento más en el
  historial— y no inventa una alerta que el código no tiene.
- **Los datos no se pierden en ningún estado de error del diálogo.** Los eventos
  `12` a `15` parten de la hora ya elegida (`11:30`), nunca de un diálogo
  reiniciado, porque el código conserva `rescheduleDate`/`rescheduleTime`.
- Todos los datos mostrados (persona, cliente, barbero, teléfono, correo,
  precio) son ficticios. El contacto aparece únicamente en el hecho «Contacto»,
  que es la superficie autorizada; `actorLabel` siempre es un nombre resuelto y
  seguro (RN-HIS-01), nunca un correo ni un identificador, y el `versionToken`
  no se dibuja en ninguna parte.
- Ninguna línea de texto se sale de su contenedor en ningún viewport y ningún
  evento de escritorio excede `1024 px` de alto: el generador falla con código
  distinto de cero en los tres casos.

## Diferencias con la lámina compuesta

La lámina de 2026-09-02 se compuso sobre canvas claro, con un selector de
cuenta decorativo junto al wordmark, «Nuevo turno» y «Cerrar sesión» como
botones de la barra superior repetidos luego en el dock, retratos fotográficos
de personas inexistentes, un icono por cada hecho de la ficha y los siete
hechos como una tabla de renglones iguales. Ninguna de esas piezas sobrevive
aquí, por las mismas razones ya registradas al construir `auth-eventos` y
`panel-agenda-eventos`: el cascarón es el ya validado, sin controles de cuenta
ni navegación duplicada, y la iconografía se limita al dock, donde el icono sí
distingue destinos. Los rótulos de la ficha ya nombran su contenido; siete
iconos nuevos habrían creado un vocabulario que el sistema no tiene, y sin
jerarquía la tabla dejaba el horario —el dato que se viene a buscar— pesando lo
mismo que el resto.

## Lo que hay que resolver antes de implementar

**Los valores de un cambio de horario se muestran formateados en la zona de la
barbería, no en crudo.** Hoy `appointment_history_change.previous_value` /
`new_value` guardan instantes RFC 3339 en UTC (`insertAppointmentRescheduledHistory`,
`apps/api/internal/modules/booking/postgres/repository.go`) y
`AppointmentDetailPage.vue` los imprime tal cual, así que el barbero ve
`2026-09-04T13:00:00Z`. El atlas los dibuja como
`4/09/2026, 8:00 a. m.`, con el mismo `formatInstantInTimezone` que ya usan el
rango de horas y la fecha de cada evento.

Es un cambio de presentación solo en el frontend —no toca backend, OpenAPI,
migraciones ni el contrato— y lo exigen tanto RN-DIS-07/CA-020-04 (todo instante
se presenta en la zona configurada de la barbería, nunca en la del dispositivo
ni en UTC) como el propio trabajo requerido del issue [#191](https://github.com/bcaceres19/barberia/issues/191),
que prohíbe mostrar valores técnicos. Aun así **cambia texto visible**, así que
debe confirmarse con el propietario antes de implementarlo; si se rechaza, el
tratamiento alternativo es dibujar el valor crudo con la misma composición de
dos líneas. Los campos que no son instantes (`Estado`, `Precio`, `Duración`)
siguen mostrándose tal como llegan.

Dos detalles más que el atlas representa fielmente y conviene no “corregir” al
implementar:

- el cuerpo de la alerta del evento `13` es el `detail` literal del backend
  («el barbero ya tiene una cita en ese intervalo»), en minúscula inicial;
- el evento `11` repite el horario vigente en «Horario actual» y en «Nuevo
  horario aproximado», porque el diálogo se abre precargado con ese horario.

## Uso por agentes de implementación

1. Seleccionar un solo evento y viewport.
2. Abrir únicamente el PNG correspondiente y este README; no usar la lámina
   compuesta como objetivo.
3. Comparar la app real y el PNG en el mismo viewport efectivo.
4. Corregir por ciclos cortos y guardar captura, comparación lado a lado y
   overlay o diff.
5. Repetir el mismo número en el viewport hermano antes de declarar el evento
   terminado.

El mockup gobierna composición, color, proporción, jerarquía, escala de
controles e iconos, centrado, alineación, densidad y estado visual. Las HU,
reglas, decisiones, contrato API y código real gobiernan el comportamiento y el
texto cuando exista una discrepancia funcional.

## Procedencia

Los 32 archivos se generan con
[`tools/mockups/detalle-turno-eventos`](../../../../../tools/mockups/detalle-turno-eventos),
con el mismo criterio que los tres atlas hermanos: `content.mjs` declara cada
evento una sola vez y `render.mjs` lo compone en HTML con los tokens de
`apps/web/src/styles/tokens.css` y las fuentes auto-hosteadas del proyecto, y lo
rasteriza con el Chromium de Playwright que ya usa la suite e2e.

```bash
node tools/mockups/detalle-turno-eventos/render.mjs
```

Un evento se declara una vez y se dibuja en los dos viewports, así que
escritorio y móvil no pueden divergir en copy, orden ni estado. Además de los
desbordes, el generador verifica que las fechas y horas calculadas en Node con
`Intl.DateTimeFormat` coincidan con las que produce el Chromium que rasteriza:
una diferencia de ICU dejaría el atlas prometiendo un formato que el navegador
no escribe.

Este directorio es evidencia de diseño para el issue
[#191](https://github.com/bcaceres19/barberia/issues/191), cuyo prompt vive en
[`docs/10-backlog/prompts/chore/issue-191-detalle-historial-reprogramacion-fidelidad.md`](../../../prompts/chore/issue-191-detalle-historial-reprogramacion-fidelidad.md)
y todavía apunta a la lámina compuesta: al ejecutarlo debe apuntar a este atlas.
No modifica `apps/web` ni autoriza implementación automática de conceptos no
representados aquí; en particular, los eventos de historial `appointment_completed`
y `appointment_cancelled_by_customer` que aparecen en los eventos `09` y `10`
pertenecen al vocabulario cerrado de DEC-041 y se dibujan porque el detalle debe
saber representarlos, no porque este atlas autorice construir completar o
cancelar un turno, que siguen fuera de alcance.
