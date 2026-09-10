# Atlas visual de `/panel/bloqueos`

Paquete raster de diseño para #197 y `HU-042`. Representa bloqueos puntuales,
series semanales y retiro lógico según la interfaz vigente. No implementa la
ruta ni añade capacidades pendientes.

## Viewports

- `desktop/bloqueos/`: 15 estados a 1440 × 1024 px.
- `mobile/bloqueos/`: 15 estados a 840 × 1870 px, equivalentes al reflow móvil
  de 420 × 935 CSS px usado por el atlas consolidado.
- Principal 1280 × 1024: `desktop/bloqueos/04a-principal-1280.png`.
- Principal 320 × 800 CSS px: `mobile/bloqueos/04a-principal-320.png`
  (exportación 800 × 2000 px).
- Principal 768 × 1024 CSS px: `tablet/bloqueos/04-principal-768.png`
  (exportación 1080 × 1440 px).

## Matriz de eventos

Cada archivo numerado existe con el mismo nombre en escritorio y móvil.

| Archivo | Estado representado |
| --- | --- |
| `01-carga-inicial.png` | Carga inicial de contexto y bloqueos. |
| `02-sin-barberos.png` | Vacío sin personas configurables. |
| `03-error-carga.png` | Error inicial recuperable. |
| `04-principal.png` | Puntuales y series vigentes para un barbero. |
| `05-vacio.png` | Sin bloqueos puntuales ni series semanales. |
| `06-error-bloqueos.png` | Error recuperable del barbero seleccionado. |
| `07-dialogo-bloqueo-puntual.png` | Alta puntual con fecha, hora y tipo. |
| `08-bloqueo-validacion.png` | Fin anterior al inicio. |
| `09-bloqueo-guardando.png` | Bloqueo puntual en persistencia. |
| `10-dialogo-serie-semanal.png` | Alta de recurrencia semanal. |
| `11-serie-validacion.png` | Duración de serie fuera de rango. |
| `12-serie-guardando.png` | Serie en persistencia. |
| `13-bloqueo-conflicto-idempotencia.png` | Intento lógico recuperable y datos conservados. |
| `14-serie-sin-conexion.png` | Error de red conservando la serie. |
| `15-retirando-bloqueo.png` | Retiro lógico puntual en curso. |

## Contrato visual y funcional

- Mantiene literalmente la firma tinta/marfil/latón, tipografías, alertas,
  diálogos, filas editoriales y dock del resto del atlas; no usa sidebar.
- El selector conserva visible el barbero y la zona IANA. Los intervalos son
  `[inicio, fin)` y las cifras horarias mantienen lectura tabular.
- Puntual y serie semanal tienen estructura, copy y formularios distintos. Los
  siete tipos autorizados siguen siendo `Descanso`, `Almuerzo`, `No disponible`,
  `Día libre`, `Festivo`, `Vacaciones` y `Emergencia`.
- Validación, idempotencia y red conservan los datos. El retiro es lógico, no
  abre una confirmación inexistente y bloquea solo la fila afectada.

No se presenta edición, impacto/conteo de citas, reprogramación, cancelación,
calendario de disponibilidad, `date_list` ni excepciones de serie pendientes.

## Procedencia y revisión

Se cruzaron `BlocksPage.vue`, sus modelos/validaciones, `RN-BLQ-01`–`04`,
`DEC-008`, `DEC-009`, `DEC-020` y el estado documentado de `HU-042`. Las
referencias fueron `08-bloqueos-puntuales-y-series.png` y los atlas consolidados
de Servicios y Servicios por barbero. Las composiciones se generaron con
ImageGen; solo se normalizaron sus dimensiones al formato del atlas.

Se abrieron principal, carga, validación puntual, serie, error y reflow estrecho.
Los 33 PNG se decodificaron sin error. Esta revisión es visual del artefacto, no
QA funcional ni certificación WCAG.
