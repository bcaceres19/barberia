# Atlas visual de `/panel/servicios-por-barbero`

Paquete raster de diseño para #195 y `HU-023`. Representa la elección de una
persona y las asignaciones reales del catálogo, incluida la protección de la
última asignación activa. No implementa la ruta ni modifica sus reglas.

## Viewports

- `desktop/servicios-por-barbero/`: 17 estados a 1440 × 1024 px.
- `mobile/servicios-por-barbero/`: 17 estados a 840 × 1870 px, equivalentes a
  420 × 935 CSS px.
- Principal 1280 × 1024: `desktop/servicios-por-barbero/03a-principal-1280.png`.
- Principal 320 × 800 CSS px: `mobile/servicios-por-barbero/03a-principal-320.png`
  (exportación 800 × 2000 px).
- Principal 768 × 1024 CSS px:
  `tablet/servicios-por-barbero/03-principal-768.png` (1080 × 1440 px).

## Matriz de eventos

Cada archivo existe en las carpetas equivalentes de escritorio y móvil.

| Archivo | Estado representado |
| --- | --- |
| `01-carga-inicial.png` | Carga paralela de barberos y servicios. |
| `02-sin-barberos.png` | Vacío sin barberos; orienta hacia la sección correcta. |
| `03-sin-servicios.png` | Vacío sin catálogo; orienta hacia Servicios. |
| `04-error-carga.png` | Error inicial recuperable. |
| `05-principal-un-barbero.png` | Selector con una persona y asignaciones mixtas. |
| `06-principal-equipo-completo.png` | Principal con contexto de equipo y servicios compartibles. |
| `07-cambio-barbero-cargando.png` | Nueva selección con asignaciones en carga. |
| `08-error-carga-asignaciones.png` | Fallo recuperable del barbero seleccionado. |
| `09-asignando-servicio.png` | Alta de una asociación en curso. |
| `10-asignacion-aplicada.png` | Asociación confirmada por el servidor. |
| `11-retirando-servicio.png` | Retiro de una asociación en curso. |
| `12-retiro-aplicado.png` | Retiro confirmado. |
| `13-ultima-asignacion-rechazada.png` | Rechazo de la última asignación activa; conserva la casilla. |
| `14-barbero-o-servicio-no-disponible.png` | Recurso inexistente o ajeno. |
| `15-cambio-sin-conexion.png` | Error de red conservando el estado anterior. |
| `16-cambio-error-inesperado.png` | Error inesperado conservando el estado anterior. |
| `17-nombres-largos-reflow.png` | Persona y servicio extensos sin clipping. |

## Contrato visual y funcional

- Reutiliza sin reinterpretación el cascarón, paleta, tipografías, dock,
  `PageState`, `BaseAlert`, monogramas y filas del resto del atlas.
- El selector mantiene visible la persona editada; cada fila expresa
  `Asignado` o `Sin asignar` además de la casilla. `Guardando…` bloquea solo la
  asociación afectada y evita aparentar un guardado optimista.
- En móvil el destino se agrupa bajo `Más`, igual que Configuración; en
  escritorio `Servicios por barbero` queda activo directamente.
- El rechazo de la última asignación usa la misma superficie/error del atlas,
  conserva la selección y explica la única solución autorizada: asignar otra
  persona antes de retirar la actual.

No incluye precio/duración por persona, comisión, orden manual, horarios,
disponibilidad, citas, desactivación de servicios, roles ni borrados.

## Procedencia y revisión

Se cruzaron `BarberServicesPage.vue`, sus modelos/pruebas, `DEC-019` y
`DEC-068`. Las referencias visuales fueron los atlas de Barberos y Servicios,
además del cascarón compartido de Agenda y Barbería. Los PNG se rasterizaron con
Chromium usando fuentes autoalojadas y los mismos tokens/componentes visuales;
no se dejó un generador como parte de la entrega.

Se abrieron muestras de principal, carga contextual, error, última asignación y
nombres largos. Los 37 PNG se decodificaron sin error y coinciden con las
dimensiones declaradas. Esta es revisión visual, no QA funcional ni
certificación WCAG.
