# Atlas visual de `/panel/servicios`

Paquete raster de diseño para #194, `HU-022` y `HU-024`. Representa el catálogo
real, alta, edición y ciclo de vida de servicios. No implementa la ruta ni añade
funciones, reglas o datos al producto.

## Viewports

- `desktop/servicios/`: 38 estados a 1440 × 1024 px.
- `mobile/servicios/`: 38 estados a 840 × 1870 px, equivalentes a 420 × 935 CSS px.
- Principal 1280 × 1024: `desktop/servicios/03a-principal-1280.png`.
- Principal 320 × 800 CSS px: `mobile/servicios/03a-principal-320.png`
  (exportación 800 × 2000 px). Usa dos registros —uno activo y uno inactivo—
  para comprobar reflow sin superponer el dock.
- Principal 768 × 1024 CSS px: `tablet/servicios/03-principal-768.png`
  (exportación 1080 × 1440 px).

## Matriz de eventos

Cada archivo de la tabla existe en `desktop/servicios/` y `mobile/servicios/`.

| Archivo | Estado representado |
| --- | --- |
| `01-carga-inicial.png` | Carga inicial del catálogo. |
| `02-lista-un-servicio.png` | Catálogo resuelto con un servicio. |
| `03-catalogo-activos-inactivos.png` | Principal con servicios activos e inactivos. |
| `04-nombre-descripcion-largos.png` | Nombre y descripción extensos con reflow. |
| `05-lista-paginada.png` | Página con más resultados disponibles. |
| `06-cargando-mas.png` | Carga de la siguiente página. |
| `07-vacio.png` | Catálogo vacío con acción para registrar el primero. |
| `08-error-carga.png` | Error recuperable y `Reintentar`. |
| `09-dialogo-agregar.png` | Alta inicial. |
| `10-alta-validacion.png` | Nombre requerido, duración inválida y precio cero. |
| `11-alta-guardando.png` | Alta en curso con controles bloqueados. |
| `12-alta-nombre-duplicado.png` | Conflicto con otro servicio activo. |
| `13-alta-conflicto-idempotencia.png` | Intento lógico en conflicto. |
| `14-alta-sin-conexion.png` | Error de red conservando datos. |
| `15-alta-error-inesperado.png` | Error inesperado conservando datos. |
| `16-alta-aplicada.png` | Servicio confirmado incorporado al catálogo. |
| `17-dialogo-editar.png` | Edición precargada. |
| `18-edicion-validacion.png` | Validación de nombre, duración y precio. |
| `19-edicion-guardando.png` | Edición en curso. |
| `20-edicion-nombre-duplicado.png` | Conflicto de nombre durante edición. |
| `21-edicion-no-disponible.png` | Servicio inexistente o ajeno. |
| `22-edicion-sin-conexion.png` | Error de red durante edición. |
| `23-edicion-error-inesperado.png` | Error inesperado durante edición. |
| `24-edicion-aplicada.png` | Lista con datos confirmados actualizados. |
| `25-desactivar-impacto-cargando.png` | Consulta del impacto real en curso. |
| `26-desactivar-sin-citas-futuras.png` | Confirmación sin citas futuras afectadas. |
| `27-desactivar-con-impacto.png` | Confirmación con recuento de citas futuras. |
| `28-desactivando.png` | Desactivación en curso. |
| `29-desactivar-conflicto-estado.png` | El estado cambió concurrentemente. |
| `30-desactivar-conflicto-idempotencia.png` | Conflicto del intento de transición. |
| `31-desactivar-sin-conexion.png` | Error de red al desactivar. |
| `32-desactivar-error-inesperado.png` | Error inesperado al desactivar. |
| `33-desactivado-aplicado.png` | Resultado confirmado como inactivo. |
| `34-dialogo-reactivar.png` | Confirmación de reactivación. |
| `35-reactivando.png` | Reactivación en curso. |
| `36-reactivar-conflicto-estado.png` | Estado ya reactivado concurrentemente. |
| `37-reactivar-sin-conexion.png` | Error de red al reactivar. |
| `38-reactivado-aplicado.png` | Resultado confirmado como activo. |

## Contrato visual y funcional

- Conserva literalmente la firma del atlas existente: tinta `#101B2B`, marfil
  `#F4F0E7`, latón `#B8955A`, Instrument Serif/Instrument Sans, cabecera plana,
  márgenes, dock, filete editorial, radios de 2 px y profundidad mínima.
- Las filas derivan de `RecordRow`; las cifras usan alineación tabular y los
  badges comunican `Activo`/`Inactivo` con texto, borde y material.
- `PageState`, campos, botones, velo y notas `BaseAlert` usan los mismos colores
  semánticos que Agenda, Barberos y Barbería. Peligro solo aparece en la acción
  destructiva y sus errores.
- COP es fijo y se muestra sin inventar símbolo `$`. Duración y precio son
  valores propios del servicio, nunca del barbero.
- Desactivar conserva servicio e historial, muestra impacto respaldado y no
  promete cancelación automática. Reactivar no crea otro registro.

Quedan fuera asignaciones, disponibilidad, horarios, paquetes, descuentos,
impuestos, cobros, inventario, servicios gratuitos y propagación silenciosa a
citas existentes.

## Procedencia y revisión

Se cruzaron `CatalogPage.vue`, sus modelos/pruebas, `RN-SER-01`–`04`,
`DEC-003`, `DEC-004`, `DEC-067` y `DEC-069`. Las referencias visuales fueron
`panel-agenda-eventos`, `barberos-eventos` y `barberia-eventos` del mismo atlas.
Los PNG se rasterizaron con Chromium usando las fuentes autoalojadas y los mismos
tokens/componentes visuales; no se dejó un generador como parte de la entrega.

Se abrieron muestras de principal, validación, impacto, conflicto y reflow. Los
79 PNG se decodificaron sin error y coinciden con las dimensiones declaradas.
Esta es revisión visual del artefacto, no QA funcional ni certificación WCAG.
