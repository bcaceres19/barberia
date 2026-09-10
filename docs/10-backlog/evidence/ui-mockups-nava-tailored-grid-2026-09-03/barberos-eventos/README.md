# Atlas visual de `/panel/barberos`

Referencia de alta fidelidad NAVA / Tailored Grid para la ruta `/panel/barberos`.
Cada evento tiene una lámina equivalente de escritorio y móvil. Las imágenes son
contrato visual del viewport representado; no crean funciones, datos ni reglas de
negocio nuevas.

## Viewports

- `desktop/barberos/`: 1440 × 1024 px.
- `mobile/barberos/`: 840 × 1870 px.
- Estado principal adicional a 1280 × 1024: `desktop/barberos/03a-principal-1280.png`.
- Estado principal adicional a 320 × 800 CSS px: `mobile/barberos/03a-principal-320.png`
  (exportación 800 × 2000 px).
- Estado principal adicional a 768 × 1024 CSS px: `tablet/barberos/03-principal-768.png`
  (exportación 1080 × 1440 px).

## Matriz de eventos

| Archivo | Estado representado |
| --- | --- |
| `01-carga-inicial.png` | Carga inicial con estructura de la lista en espera. |
| `02-lista-un-barbero.png` | Lista resuelta con una persona. |
| `03-lista-cuatro-barberos.png` | Estado principal con cuatro personas. |
| `04-lista-nombre-largo.png` | Reflow de un nombre extenso. |
| `05-lista-paginada.png` | Página con más resultados disponibles. |
| `06-cargando-mas.png` | Solicitud de la siguiente página en curso. |
| `07-vacio.png` | Equipo vacío con orientación para registrar el primero. |
| `08-error-carga.png` | Error recuperable de carga y acción `Reintentar`. |
| `09-dialogo-agregar.png` | Alta abierta en estado inicial. |
| `10-alta-validacion.png` | Alta con nombre requerido vacío. |
| `11-alta-guardando.png` | Alta en curso con controles bloqueados. |
| `12-alta-conflicto-idempotencia.png` | Conflicto del intento anterior conservando el formulario. |
| `13-alta-sin-conexion.png` | Error de conexión durante el alta. |
| `14-alta-error-inesperado.png` | Error inesperado durante el alta. |
| `15-alta-aplicada.png` | Lista resuelta después de agregar una persona. |
| `16-dialogo-renombrar.png` | Edición de nombre abierta. |
| `17-renombrar-validacion.png` | Edición con nombre requerido vacío. |
| `18-renombrar-guardando.png` | Cambio de nombre en curso con controles bloqueados. |
| `19-renombrar-no-disponible.png` | La persona ya no está disponible. |
| `20-renombrar-sin-conexion.png` | Error de conexión durante el cambio de nombre. |
| `21-renombrar-error-inesperado.png` | Error inesperado durante el cambio de nombre. |
| `22-renombrado-aplicado.png` | Lista resuelta con el nombre actualizado. |

## Contrato visual

- Firma cromática: tinta `#101B2B`, marfil `#F4F0E7`, latón `#B8955A` y
  grises funcionales de la familia NAVA.
- Voz editorial serif en títulos y nombres; controles y mensajes en sans-serif.
- Composición Tailored Grid idéntica a Agenda, Nuevo turno y Detalle: cabecera
  plana, contenido desde el mismo margen, filete editorial, bordes finos, radios
  contenidos y navegación inferior persistente.
- Carga, vacío y error de página usan el mismo `PageState`; los errores dentro
  de diálogos usan la misma nota al margen de `BaseAlert`, con superficie y
  filete semánticos, rótulo en versalitas y texto explícito.
- Los diálogos conservan contexto mediante velo oscuro; en móvil se presentan
  como paneles amplios y táctiles.
- Éxito se comunica mediante la lista ya actualizada, sin notificación inventada.

## Límites funcionales

Estas láminas solo representan capacidades existentes: listar, paginar, agregar y
renombrar barberos, además de sus estados de carga, vacío, validación y error. No
incluyen fotografía, cargo, correo, teléfono, eliminación, desactivación, filtros
ni otras acciones no autorizadas por la ruta actual.

## Procedencia

Rasterizado con Chromium y las mismas fuentes autoalojadas, tokens, shell,
`PageState`, alertas, campos, botones y navegación que los atlas de Agenda,
Nuevo turno y Detalle. No se usó interpretación generativa para color o
componentes. Los 47 PNG se decodificaron y verificaron en sus dimensiones finales.
