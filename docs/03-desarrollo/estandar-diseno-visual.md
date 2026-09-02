---
titulo: "Estándar visual NAVA para rediseños y pantallas nuevas"
version: "4.0"
estado: "Normativo para rediseños y pantallas nuevas"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-09-02"
documentos_relacionados:
  - "../00-control/registro-decisiones.md"
  - "../01-producto/alcance-mvp.md"
  - "../01-producto/reglas-negocio.md"
  - "../02-requisitos/estados-citas.md"
  - "../04-arquitectura/frontend.md"
  - "../10-backlog/evidence/ui-redesign-nava-2026-09-02/README.md"
  - "especificacion-frontend-nava.md"
  - "estandar-frontend-vue.md"
  - "estrategia-pruebas.md"
---

# Estándar visual NAVA para rediseños y pantallas nuevas

## 1. Propósito, alcance y autoridad

Este documento define el contrato visual mínimo de NAVA. Es obligatorio cuando un issue:

- crea una pantalla, flujo o componente visible nuevo;
- rediseña de forma completa una pantalla o región existente;
- extiende un patrón cubierto por los mockups de referencia;
- introduce una nueva base de estilos o una dependencia visual.

Una corrección funcional o accesible aislada no obliga a rediseñar la pantalla completa. Las pantallas existentes pueden conservar temporalmente su aspecto hasta que un issue autorice su rediseño; dentro de una pantalla rediseñada no se dejan regiones principales a medio migrar.

La identidad NAVA, su firma cromática, el contraste editorial/funcional y las cualidades visuales de este documento son obligatorios dentro del alcance anterior. Continúan libres la composición concreta, el grid, los tamaños, el espaciado, los radios, las sombras, la navegación, el inventario de componentes y la tecnología de estilos, siempre que el resultado conserve esa identidad y cumpla las garantías de calidad.

Fuentes normativas: `DEC-077`, `DEC-078` y `DEC-079`. `DEC-079` matiza a `DEC-078`: mantiene la libertad de creación y herramientas, pero vuelve obligatorios para rediseños y pantallas nuevas los anclajes cromáticos y el lenguaje visual reconocido en los mockups aprobados.

Si un mockup contradice una regla de negocio, el alcance, una HU, un contrato, la seguridad o la accesibilidad, prevalece la fuente normativa correspondiente. El mockup nunca crea funciones.

## 2. Referencias visuales aprobadas

Las siguientes láminas forman la referencia canónica para el lenguaje visual de las familias que representan:

| Referencia | Familia cubierta |
| --- | --- |
| [Autenticación móvil](../10-backlog/evidence/ui-redesign-nava-2026-09-02/01-autenticacion-movil.png) | Acceso, validación y recuperación. |
| [Nuevo turno responsive](../10-backlog/evidence/ui-redesign-nava-2026-09-02/02-nuevo-turno-responsive.png) | Formularios extensos, secciones, reflow y conflictos. |
| [Componentes, formularios y alertas](../10-backlog/evidence/ui-redesign-nava-2026-09-02/03-componentes-formularios-alertas.png) | Botones, campos, foco, estados, alertas y diálogo. |
| [Servicios, configuración y navegación](../10-backlog/evidence/ui-redesign-nava-2026-09-02/04-servicios-configuracion-navegacion.png) | Shell, navegación responsive, listas operativas, modal y éxito persistente. |

Respetar los mockups significa conservar su familia visual, color, jerarquía, densidad deliberada, claridad de estados y relación entre voz editorial y controles funcionales. No significa copiar coordenadas, textos sintéticos, datos, tamaños, radios o una composición exacta. La implementación puede apartarse de una distribución concreta cuando el contenido real, el ancho, la accesibilidad o el rendimiento lo exijan.

## 3. Identidad obligatoria

### 3.1 Personalidad

NAVA se reconoce como sastrería contemporánea y hospitalidad boutique: precisa, seria, elegante, cálida, serena y profesional. La interfaz debe mostrar:

- jerarquía editorial y alineación cuidada;
- superficies calmadas, líneas finas y profundidad contenida;
- una acción principal fácil de encontrar;
- lectura inmediata del tiempo, la persona, el servicio y el estado del turno;
- densidad ordenada en el panel privado y mayor calma en flujos públicos o de autenticación;
- copy breve, humano y operativo.

Se evitan como lenguaje dominante los clichés de barbería —tijeras, navajas, bigotes o postes usados como decoración—, el neón, la estética gaming, el vidrio ornamental, las sombras profundas, las tarjetas grandes y redondeadas para cada fragmento y la apariencia de dashboard genérico. Un recurso de ese tipo solo puede aparecer si resuelve una necesidad concreta, no desplaza la identidad y se justifica en la revisión visual.

### 3.2 Marca y contexto

- El wordmark se escribe `NAVA`, en mayúsculas, sin agregar “App”, “Studio” o “Barber”.
- En el panel privado, NAVA identifica la plataforma y el nombre de la barbería identifica el tenant activo.
- En la reserva pública, el nombre de la barbería conserva la jerarquía principal y NAVA aparece como firma secundaria.
- No se inventan logotipos de barberías ni personalización por tenant.

## 4. Firma cromática obligatoria

### 4.1 Colores base

Los siguientes colores son los anclajes de identidad. No es obligatorio usar todos en cada pantalla, pero la pantalla no puede sustituirlos por una paleta de marca diferente.

| Rol | Valor de referencia | Uso esperado |
| --- | --- | --- |
| Tinta NAVA | `#101B2B` | Wordmark, acciones primarias, shell y superficies de alto contraste. |
| Marfil cálido | `#F4F0E7` | Canvas y fondos dominantes. |
| Blanco | `#FFFFFF` | Campos, diálogos y superficies elevadas o de lectura. |
| Grafito principal | `#2A2D32` | Texto principal sobre superficies claras. |
| Grafito secundario | `#5E625F` | Ayuda, metadatos y texto secundario. |
| Piedra | `#E8E2D8` | Superficie atenuada, agrupación y separación suave. |
| Borde piedra | `#C9C0B2` | Divisores y bordes no interactivos. |
| Salvia | `#748477` | Acento secundario, bordes de control y apoyo de estados positivos. |
| Latón | `#B8955A` | Acento de marca medido, selección y detalles editoriales. |
| Latón oscuro | `#765C2F` | Foco visible y acción suave cuando mantenga contraste. |

Para interacción sobre tinta se admiten `#18283D` en hover y `#0A1420` en activo. Los tonos derivados mediante transparencia, mezcla o aclarado son válidos cuando mantienen la familia cromática y el contraste requerido.

### 4.2 Colores semánticos

| Estado | Superficie | Texto | Borde |
| --- | --- | --- | --- |
| Éxito | `#EAF0EB` | `#325D43` | `#748477` |
| Advertencia o conflicto | `#F8F1DF` | `#775019` | `#9A6A24` |
| Error o peligro | `#F8EDEC` | `#8A2C2C` | `#A43A3A` |
| Información | `#E9EEF3` | `#23405B` | `#667D93` |
| Inactivo | `#EEECE8` | `#56514A` | `#C9C0B2` |

El color nunca es la única señal: un estado incluye texto y, cuando ayude a reconocerlo, icono, forma o borde. La acción destructiva usa rojo únicamente cuando la consecuencia es destructiva; no se usa como acento decorativo.

### 4.3 Libertad dentro de la paleta

- No son obligatorios los nombres de variables ni la existencia de un archivo de tokens concreto.
- Una pantalla puede crear tintes, transparencias o combinaciones propias a partir de la paleta.
- Un color adicional exige una función semántica o de legibilidad que la paleta no cubra, contraste comprobado y justificación en el issue o PR. No puede convertirse en una segunda identidad de marca.
- Los estados de un turno pueden derivar de la paleta semántica, pero siempre conservan su texto canónico y una señal adicional al color.
- No se ofrece personalización de color por barbería salvo decisión de producto posterior.

## 5. Tipografía, forma y profundidad

### 5.1 Contraste tipográfico

El contraste entre una voz editorial serif y una voz funcional sans-serif es parte de la firma NAVA:

- el wordmark y los títulos editoriales pueden usar `Instrument Serif` o un fallback serif compatible;
- controles, labels, ayudas, mensajes, tablas y datos operativos usan `Instrument Sans` o un fallback sans-serif legible;
- la familia exacta, los pesos y la escala no son obligatorios si la entrega conserva ese contraste, tiene licencia válida, carga eficiente y legibilidad suficiente;
- una fuente nueva se incorpora de forma self-hosted o mediante el mecanismo aprobado por arquitectura; no se enlaza desde un componente sin revisar licencia, privacidad y rendimiento.

### 5.2 Forma y ritmo

Los mockups favorecen bordes finos, radios pequeños, sombras mínimas, agrupación por líneas y un ritmo editorial claro. Esta intención es obligatoria; los valores concretos son libres.

- La profundidad distingue capas y decisiones, no decora cada bloque.
- Las tarjetas no sustituyen automáticamente la jerarquía de página.
- El espacio puede ser compacto u holgado según la tarea, pero mantiene alineación y relaciones visibles entre label, control, ayuda y error.
- Iconos funcionales viven dentro de controles con nombre accesible. Los iconos decorativos no reemplazan texto ni crean clichés de marca.

## 6. Patrones cubiertos por los mockups

### 6.1 Botones y acciones

- La acción primaria usa tinta NAVA y una etiqueta que nombra el resultado.
- La acción secundaria mantiene menor peso visual; una acción suave o enlace no compite con la primaria.
- La acción destructiva usa el tratamiento de peligro y describe el efecto, por ejemplo “Desactivar servicio”.
- Carga, deshabilitado, foco, hover y activo son distinguibles. Deshabilitar durante una mutación no sustituye el texto de progreso.
- Una región no presenta varias acciones con el mismo peso sin una razón explícita.

### 6.2 Campos y formularios

- Todo campo conserva label visible. Placeholder y ayuda complementan, no sustituyen el label.
- La anatomía `label → control → ayuda/error` se mantiene estable y el error aparece junto al campo.
- Requerido y opcional se comunican por texto o marca entendible, no solo por color.
- El foco de teclado es visible y no queda oculto por barras, diálogos o teclado virtual.
- Un formulario extenso puede agruparse por pasos, secciones o columnas. En móvil se transforma a una columna o a pasos completos sin perder datos ni acciones.
- Un error de envío muestra resumen persistente cuando existen varios campos afectados; el foco se mueve de forma útil sin encerrar a la persona.

### 6.3 Alertas y estados

- Error, conflicto, éxito, información y advertencia usan superficie, borde, icono o forma, título y explicación accionable.
- Un éxito importante permanece dentro de la tarea; no depende solo de un toast efímero.
- Un conflicto conserva las selecciones válidas y explica qué puede hacer la persona.
- Un error inesperado no expone stack, proveedor, token ni datos personales; puede mostrar un `request_id` seguro cuando el contrato lo proporcione.

### 6.4 Diálogos y capas

- El diálogo tiene título, propósito, cierre accesible y acciones ordenadas por consecuencia.
- El foco entra, queda contenido mientras corresponde y vuelve al disparador al cerrar.
- En móvil, una decisión breve puede usar diálogo o bottom sheet; un flujo largo se convierte preferentemente en página o paso completo.
- El overlay no oculta el contexto hasta volverlo incomprensible ni permite interactuar accidentalmente con el fondo.

### 6.5 Navegación

- Ninguna etiqueta necesaria se trunca hasta perder significado.
- Escritorio puede usar header, rail, sidebar o una combinación; móvil puede usar dock, menú o `Más`.
- La lámina de servicios aprueba como patrón posible cuatro destinos visibles y `Más` para los restantes. No es una obligación si otra solución conserva claridad, orden, área segura y acceso por teclado.
- La navegación no inventa módulos ni convierte acciones globales en destinos sin respaldo de producto.

## 7. Zona de libertad creativa

Cada pantalla o flujo puede decidir libremente:

- composición, columnas, orden visual y uso del espacio;
- tamaños tipográficos, pesos y escala;
- espaciado, grid, radios, bordes, sombras y movimiento;
- header, rail, sidebar, dock, tabs, menú o navegación contextual;
- componentes locales o compartidos;
- CSS estándar, `<style scoped>`, CSS Modules, Tailwind, utility-first, biblioteca visual o una solución híbrida;
- fotografía, ilustración, textura o ausencia de imagen, cuando aporten a la tarea y no contradigan la identidad.

La libertad no permite sustituir la firma cromática, perder el contraste serif/sans característico, producir una apariencia genérica ajena a NAVA, copiar datos sintéticos del mockup como si fueran contrato, omitir estados o debilitar accesibilidad, responsive, rendimiento, seguridad o pruebas.

Los componentes compartidos se extraen cuando existe repetición real de comportamiento, semántica o accesibilidad. No se crea un inventario global por adelantado ni se obliga a que todas las pantallas tengan la misma composición.

## 8. Accesibilidad, responsive y estados obligatorios

1. **Semántica:** HTML semántico, jerarquía de encabezados, labels reales y ARIA solo cuando el HTML nativo no alcance.
2. **Teclado y foco:** toda acción necesaria funciona con teclado; el foco es visible, no queda cubierto y se restaura al cerrar una capa.
3. **Contraste:** las combinaciones implementadas cumplen WCAG 2.2 AA. Los valores de la paleta no eximen de medir la combinación real.
4. **Interacción táctil:** objetivo de usabilidad de 44 × 44 px cuando la composición lo permita; cualquier excepción conserva operabilidad y legibilidad.
5. **Responsive:** no se oculta información ni acción necesaria. Se revisan 320, 360, 768 y 1280 px, zoom 200 % y el teléfono del piloto.
6. **Estados:** carga, actualización, vacío, error recuperable, conflicto, éxito, sesión vencida, inactivo y deshabilitado se representan cuando aplican.
7. **Movimiento:** apoya la comprensión, no retrasa la tarea y respeta `prefers-reduced-motion`.
8. **Contenido límite:** nombres largos, unicode, cero y muchas opciones, conexión lenta y reintentos forman parte de la revisión.
9. **Privacidad:** ninguna captura, fixture, mensaje o log contiene datos personales reales, secretos o tokens.

## 9. Herramientas de implementación

No se impone una tecnología de estilos. Tailwind está permitido, pero no se instala automáticamente. Cualquier dependencia visual nueva documenta en el issue o PR:

- versión y licencia;
- alcance de uso;
- mantenimiento y compatibilidad;
- impacto en bundle, carga e interacción;
- compilación o purga de estilos;
- accesibilidad de componentes y estados;
- estrategia para conservar la firma cromática NAVA.

La elección técnica de una pantalla no se convierte por sí sola en obligación para todo el producto.

## 10. Frontera funcional

El diseño no modifica reglas, permisos, disponibilidad, auditoría, idempotencia, aislamiento entre barberías ni contratos. Continúan vigentes:

- el término visible `turno` y los estados autorizados;
- el selector obligatorio de un barbero para la agenda P0;
- la pertenencia diaria de turnos nocturnos definida por `DEC-075`;
- la exclusión de pagos, caja, ingresos, inventario, CRM, marketplace, cuenta de cliente y demás capacidades fuera del MVP;
- la autoridad del backend para validar conflictos y operaciones críticas.

## 11. Entrega y criterio de terminado

Cada rediseño o pantalla nueva parte de un issue real y documenta:

1. HU, reglas, decisiones y contrato afectados;
2. mockup o familia visual tomada como referencia;
3. cómo conserva la paleta y la identidad NAVA;
4. estados y anchos que se verificarán;
5. dependencia visual nueva, si existe, con su justificación;
6. pruebas de componente, accesibilidad, E2E y rendimiento aplicables.

La pantalla está terminada cuando:

- pertenece visualmente a la misma familia que los mockups aprobados;
- usa la firma cromática NAVA o derivados permitidos con contraste comprobado;
- conserva la libertad de composición sin romper jerarquía ni operabilidad;
- cubre estados, responsive, teclado, foco y movimiento reducido;
- no amplía el alcance ni contradice contrato o reglas;
- aporta pruebas y evidencia proporcionales.

No se rechaza una solución por usar CSS local, Tailwind, otra composición, otro tamaño, otro radio o una variante propia. Sí se rechaza si sustituye la identidad cromática, pierde el lenguaje NAVA, copia funciones no autorizadas del mockup o incumple las garantías anteriores.

## 12. Referencias

- `DEC-077`: identidad NAVA y dirección Tailored Grid.
- `DEC-078`: libertad de composición, componentes y herramientas.
- `DEC-079`: mockups y firma cromática obligatorios para rediseños y pantallas nuevas.
- [Handoff de mockups NAVA](../10-backlog/evidence/ui-redesign-nava-2026-09-02/README.md).
- [Especificación de frontend NAVA](especificacion-frontend-nava.md).
- [Estándar de frontend Vue](estandar-frontend-vue.md).
- [Estrategia de pruebas](estrategia-pruebas.md).
- [WCAG 2.2](https://www.w3.org/TR/WCAG22/).
