# Atlas integral de mockups NAVA / Tailored Grid

## Propósito

Este atlas reúne la referencia visual elegida por el propietario y los mockups necesarios para orientar rediseños y pantallas nuevas de NAVA. La dirección final conserva los colores de la familia ya aprobada y toma de Tailored Grid su precisión editorial, sus líneas finas, su organización por carriles y su contraste entre superficies operativas oscuras y superficies de lectura claras.

No propone que toda la aplicación sea oscura ni que toda la aplicación sea clara. La síntesis esperada es:

- tinta NAVA para marca, navegación, tiempo y acciones operativas dominantes;
- marfil y blanco para formularios, configuración y lectura prolongada;
- grafito y piedra para texto, divisores y estados neutros;
- salvia para éxito o disponibilidad;
- latón para selección, foco editorial y acciones destacadas;
- rojo únicamente para error o acción destructiva.

Los valores normativos están en [`estandar-diseno-visual.md`](../../../03-desarrollo/estandar-diseno-visual.md). Los colores impresos dentro de una imagen son aproximaciones visuales y no sustituyen ese documento.

## Referencia maestra

| Archivo | Función |
| --- | --- |
| [`00-referencia-tailored-grid.png`](00-referencia-tailored-grid.png) | Imagen elegida por el propietario. Define el carácter editorial, la densidad ordenada, el uso del tiempo como estructura y la convivencia de escritorio operativo con flujo móvil. Sus módulos de clientes, reportes, inventario, métricas y vista multi-barbero no forman parte del alcance vigente y no deben copiarse. |

## Cobertura implementada

Estas láminas representan rutas o superficies existentes. La implementación debe validar textos, datos, reglas, estados y acciones contra el código, las HU y el contrato real; el mockup no crea comportamiento.

| Archivo | Rutas o patrones cubiertos |
| --- | --- |
| [`01-agenda-diaria-responsive.png`](01-agenda-diaria-responsive.png) | `/panel`: agenda de un barbero seleccionado, fecha, carril temporal, lista accesible, carga, vacío y error. |
| [`02-acceso-recuperacion.png`](02-acceso-recuperacion.png) | Lámina histórica de `/acceso` y `/recuperar-acceso`. No se usa como objetivo de implementación porque reúne varias pantallas. La referencia ejecutable separada por viewport y evento está en [`auth-eventos`](../ui-mockups-nava-tailored-grid-2026-09-03/auth-eventos/README.md). |
| [`03-nuevo-turno.png`](03-nuevo-turno.png) | `/panel/turnos/nuevo`: selección, fecha/hora, datos de contacto, resumen, validación y conflicto recuperable. |
| [`04-detalle-historial-reprogramacion.png`](04-detalle-historial-reprogramacion.png) | `/panel/turnos/:appointmentId`: hechos, estado, historial paginado, reprogramación y conflicto. |
| [`05-servicios-y-asignacion.png`](05-servicios-y-asignacion.png) | `/panel/servicios` y `/panel/servicios-por-barbero`: catálogo, alta/edición, estado, asignación, guardado y rechazo de última asignación. |
| [`06-barberos-y-barberia.png`](06-barberos-y-barberia.png) | `/panel/barberos` y `/panel/barberia`: lista, alta/edición, vacío, error, foco y configuración básica guardada. |
| [`07-horarios-y-excepciones.png`](07-horarios-y-excepciones.png) | `/panel/horarios`: tramos semanales, calendario de festivos, excepciones, solapamiento y reflow móvil. |
| [`08-bloqueos-puntuales-y-series.png`](08-bloqueos-puntuales-y-series.png) | `/panel/bloqueos`: bloqueos puntuales, series semanales, creación, retiro, error y móvil. |
| [`14-componentes-formularios-alertas.png`](14-componentes-formularios-alertas.png) | Anatomía de botones, campos, foco, disabled, carga, alertas, diálogo y jerarquía de acciones. |

## Conceptos P0 pendientes

Estas láminas están rotuladas como `Concepto P0 pendiente · no implementado`. Sirven para reservar dirección visual, no para afirmar que la función existe ni para autorizar su desarrollo sin HU, issue, contrato y reglas vigentes.

| Archivo | Concepto cubierto |
| --- | --- |
| [`09-reserva-publica-pasos-1-a-3.png`](09-reserva-publica-pasos-1-a-3.png) | Servicio, preferencia de barbero, fecha y hora en la reserva pública sin cuenta. |
| [`10-reserva-publica-pasos-4-a-6.png`](10-reserva-publica-pasos-4-a-6.png) | Datos mínimos, revisión y confirmación de la reserva pública. |
| [`11-gestion-publica-y-estados.png`](11-gestion-publica-y-estados.png) | Consulta mediante enlace seguro, cancelación, éxito, enlace inválido, permiso, offline, error, carga y reintento. |
| [`12-ciclo-operativo-turno.png`](12-ciclo-operativo-turno.png) | Edición de servicio/duración, cancelación por barbería, completado, no asistencia, corrección terminal e historial auditable. |
| [`13-configuracion-operativa-comunicaciones.png`](13-configuracion-operativa-comunicaciones.png) | Anticipación, ventana, rejilla, cancelación, recordatorios, canales, privacidad y estados de conexión. |

## Reglas de lectura para diseño e implementación

1. Conservar la paleta NAVA; no reemplazarla por una identidad distinta.
2. Usar Tailored Grid como principio de composición, no como plantilla rígida: cada pantalla puede elegir su grid, densidad, navegación, componentes y tecnología.
3. Reservar las superficies oscuras para orientación, contexto operativo, tiempo, navegación o una acción dominante. Mantener claros los formularios y textos largos.
4. En escritorio, favorecer tablas, carriles y paneles alineados. En móvil, hacer reflow a una columna; nunca miniaturizar el escritorio.
5. Mantener una acción primaria clara por región. Las acciones destructivas usan confirmación y lenguaje explícito.
6. Mostrar estados con texto, icono y estructura además del color. Los errores deben preservar lo escrito y ofrecer recuperación cuando sea posible.
7. Mantener foco visible, labels persistentes, controles táctiles cómodos, contraste WCAG 2.2 AA y orden de teclado coherente.
8. Usar fotografías solo si el producto dispone de ellas y existe fallback accesible. Los avatares con iniciales son suficientes; la imagen no vuelve obligatoria una foto.
9. La navegación móvil de referencia muestra `Agenda`, `Servicios`, `Barberos` y `Más`; los destinos restantes viven en `Más` para evitar truncamiento.
10. La tecnología visual sigue libre. CSS, CSS Modules, utility-first, Tailwind o una biblioteca pueden usarse si el issue justifica dependencia, licencia, bundle, rendimiento y accesibilidad.

## Límites de alcance

No copiar de la imagen maestra ni introducir desde estos mockups:

- panel multi-barbero en P0;
- clientes o CRM;
- ventas, caja, ingresos o métricas comerciales;
- reportes o inventario;
- marketplace, pagos o upselling;
- cuenta o contraseña para el cliente de la reserva pública;
- estado `in_progress`;
- secretos, tokens, datos personales reales o destinos de notificación inseguros.

## Verificación mínima al implementar

- 320, 360, 768 y 1280 px, además del teléfono del piloto;
- zoom de texto al 200 %, teclado, foco y restauración de foco;
- `prefers-reduced-motion` y ausencia de movimiento imprescindible;
- carga, vacío, error, conflicto, éxito, deshabilitado y offline cuando apliquen;
- pruebas de componente para controles interactivos y E2E para recorridos P0 afectados;
- contraste de las combinaciones realmente implementadas, no inferido de la imagen raster.

## Generación y revisión

Las láminas se produjeron con el modo integrado de ImageGen, usando `00-referencia-tailored-grid.png`, la familia NAVA previamente aprobada y la inspección de las 11 rutas reales como fuentes de composición y contenido. Se descartaron las variantes preliminares que aclaraban demasiado la identidad o mostraban navegación y módulos no válidos. Las versiones finales corrigieron la navegación móvil, la selección activa, los controles de cuenta decorativos y la iconografía cliché.

Este cambio corresponde al issue [#183](https://github.com/bcaceres19/barberia/issues/183). No modifica `apps/web` ni autoriza implementación automática de los conceptos pendientes.

El handoff ejecutable para Claude Code está en [`redisenio-integral-nava-tailored-grid-v2.md`](../../prompts/orchestration/redisenio-integral-nava-tailored-grid-v2.md). Su issue maestro es [#184](https://github.com/bcaceres19/barberia/issues/184) y exige entregas independientes por familia.
