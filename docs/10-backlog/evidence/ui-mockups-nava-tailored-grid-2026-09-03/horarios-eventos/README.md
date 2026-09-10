# Atlas visual de `/panel/horarios`

Paquete raster de diseño para #196, `HU-040` y `HU-041`. Representa la jornada
semanal, festivos colombianos y excepciones por fecha que existen en la ruta. No
implementa cambios ni amplía reglas del producto.

## Viewports

- `desktop/horarios/`: 16 estados a 1440 × 1024 px.
- `mobile/horarios/`: 16 estados a 840 × 1870 px, equivalentes al reflow móvil
  de 420 × 935 CSS px usado por el atlas consolidado.
- Principal 1280 × 1024: `desktop/horarios/04a-principal-1280.png`.
- Principal 320 × 800 CSS px: `mobile/horarios/04a-principal-320.png`
  (exportación 800 × 2000 px).
- Principal 768 × 1024 CSS px: `tablet/horarios/04-principal-768.png`
  (exportación 1080 × 1440 px).

## Matriz de eventos

Cada archivo numerado existe con el mismo nombre en escritorio y móvil.

| Archivo | Estado representado |
| --- | --- |
| `01-carga-inicial.png` | Carga paralela de barberos y configuración. |
| `02-sin-barberos.png` | Vacío sin personas configurables. |
| `03-error-carga.png` | Error inicial recuperable. |
| `04-principal.png` | Jornada semanal completa, festivos y excepciones. |
| `05-horario-sin-tramos.png` | Siete días visibles sin tramos. |
| `06-error-horario.png` | Error de jornada del barbero seleccionado. |
| `07-dialogo-agregar-tramo.png` | Alta de tramo con zona e intervalo explícitos. |
| `08-tramo-validacion.png` | Hora requerida y duración fuera de rango. |
| `09-tramo-solape.png` | Solape recuperable conservando los datos. |
| `10-tramo-guardando.png` | Alta en curso y doble envío bloqueado. |
| `11-dialogo-excepcion-cerrada.png` | Excepción de fecha como día cerrado. |
| `12-dialogo-excepcion-segmentos.png` | Excepción abierta con múltiples segmentos. |
| `13-excepcion-fecha-duplicada.png` | Conflicto por excepción existente. |
| `14-excepcion-guardando.png` | Persistencia de excepción en curso. |
| `15-calendario-festivos-error.png` | Fallo recuperable al cambiar el calendario. |
| `16-retirando-excepcion.png` | Retiro de una excepción en curso. |

## Contrato visual y funcional

- Reutiliza el lienzo tinta `#101B2B`, marfil `#F4F0E7`, latón `#B8955A`,
  tipografía editorial, filetes, diálogos, alertas y dock del atlas vigente.
- Los siete días ISO permanecen visibles; un día sin jornada comunica `Sin
  tramos`. En móvil las filas hacen reflow y no se miniaturiza una tabla.
- La zona `America/Bogota` y la semántica `[inicio, fin)` se muestran junto al
  contexto editado. Se incluyen jornada partida y múltiples segmentos.
- La excepción por fecha reemplaza visualmente la jornada recurrente y el
  festivo aplicable; `Cerrado` y `Abierto con segmentos` se distinguen con copy,
  estructura y selección, no solo color.
- Los errores permanecen junto al control o sección afectada y conservan la
  edición. Guardados y retiros bloquean únicamente la operación en curso.

Quedan fuera plantillas, calendario mensual, disponibilidad calculada, citas,
automatismos, cambio de zona y recurrencias nuevas.

## Procedencia y revisión

Se cruzaron `SchedulesPage.vue`, modelos, validaciones y pruebas de `schedules`,
`RN-DIS-01`–`07`, `DEC-007`, `DEC-020`, `HU-040` y `HU-041`. Las referencias
visuales fueron `07-horarios-y-excepciones.png` y los atlas consolidados de
Servicios y Servicios por barbero. Las composiciones se generaron con ImageGen;
solo se normalizaron sus dimensiones al formato del atlas.

Se abrieron principal, carga, validación, excepción segmentada, error de
festivos y reflow estrecho. Los 35 PNG se decodificaron sin error. Esta revisión
es visual del artefacto, no QA funcional ni certificación WCAG.
