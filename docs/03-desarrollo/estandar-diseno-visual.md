---
titulo: "Estándar visual NAVA para rediseños y pantallas nuevas"
version: "5.2"
estado: "Normativo para rediseños y pantallas nuevas"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-09-03"
documentos_relacionados:
  - "../00-control/registro-decisiones.md"
  - "../01-producto/alcance-mvp.md"
  - "../01-producto/reglas-negocio.md"
  - "../02-requisitos/estados-citas.md"
  - "../04-arquitectura/frontend.md"
  - "../10-backlog/evidence/ui-redesign-nava-2026-09-02/README.md"
  - "../10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/README.md"
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

La identidad NAVA, su firma cromática, el contraste editorial/funcional y las cualidades visuales de este documento son obligatorios dentro del alcance anterior. La composición concreta, el grid, los tamaños, el espaciado, los radios, las sombras, la navegación, el inventario de componentes y la tecnología de estilos permanecen libres únicamente cuando no existe una referencia exacta asignada o para decisiones que esa referencia no representa. Si un issue, prompt o instrucción asigna un mockup concreto o pide igualarlo, reproducirlo o corregir la interfaz contra él, se activa el modo de fidelidad de la sección 2.1.

Fuentes normativas: `DEC-077`, `DEC-078`, `DEC-079` y `DEC-080`. `DEC-079` matiza a `DEC-078` en la identidad obligatoria; `DEC-080` distingue libertad creativa sin referencia exacta de fidelidad medible cuando sí existe un mockup asignado.

Si un mockup contradice una regla de negocio, el alcance, una HU, un contrato, la seguridad o la accesibilidad, prevalece la fuente normativa correspondiente. El mockup nunca crea funciones.

## 2. Referencias visuales aprobadas

Las siguientes láminas forman la referencia canónica para el lenguaje visual de las familias que representan:

| Referencia | Familia cubierta |
| --- | --- |
| [Referencia maestra Tailored Grid](../10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/00-referencia-tailored-grid.png) | Composición editorial, carriles temporales, densidad ordenada y relación entre escritorio operativo y flujo móvil. Sus módulos fuera de alcance no se copian. |
| [Autenticación móvil](../10-backlog/evidence/ui-redesign-nava-2026-09-02/01-autenticacion-movil.png) | Acceso, validación y recuperación. |
| [Acceso y recuperación por viewport y evento](../10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/auth-eventos/README.md) | Referencia exacta de `/acceso` y `/recuperar-acceso`: un PNG por pantalla o evento, con pares equivalentes en escritorio y móvil. Sustituye la lámina compuesta como objetivo de implementación. |
| [Nuevo turno responsive](../10-backlog/evidence/ui-redesign-nava-2026-09-02/02-nuevo-turno-responsive.png) | Formularios extensos, secciones, reflow y conflictos. |
| [Componentes, formularios y alertas](../10-backlog/evidence/ui-redesign-nava-2026-09-02/03-componentes-formularios-alertas.png) | Botones, campos, foco, estados, alertas y diálogo. |
| [Servicios, configuración y navegación](../10-backlog/evidence/ui-redesign-nava-2026-09-02/04-servicios-configuracion-navegacion.png) | Shell, navegación responsive, listas operativas, modal y éxito persistente. |

El [atlas integral de mockups](../10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/README.md) amplía estas referencias a las once rutas implementadas, sus estados transversales y conceptos P0 pendientes. Las láminas rotuladas como no implementadas reservan intención visual, pero no crean alcance, contrato ni autorización de desarrollo.

Para acceso y recuperación se asigna siempre el archivo individual del viewport y evento que se implementa; no se usa `02-acceso-recuperacion.png` como objetivo de fidelidad. Los códigos de seis dígitos adoptan seis casillas visuales. Ese patrón solo es conforme si sigue representando un único valor lógico, admite pegar el código completo, distribuye sus dígitos, soporta avance, retroceso y teclado numérico, conserva foco visible y expone un nombre e instrucciones accesibles coherentes.

En autenticación, el error local se muestra debajo de su campo y el resumen global después de las acciones/enlaces del formulario, sin repetir literalmente los errores locales. `Input`, `PasswordInput`, `OtpInput` e `InlineAlert` mantienen una sola anatomía visual entre acceso y recuperación. Durante un envío, campos y acción principal quedan bloqueados en todos los viewports. Las alertas entran en `160 ms` mediante opacidad y una traslación vertical máxima de `6 px`, sin mover el foco; `prefers-reduced-motion` elimina la traslación y cualquier retraso no esencial.

### 2.1 Modos de conformidad

Antes de diseñar o editar, el issue y el PR declaran uno de estos modos:

| Modo | Cuándo aplica | Contrato visual |
| --- | --- | --- |
| **Identidad guiada** | No existe un mockup exacto asignado a la pantalla/estado, o la instrucción pide explorar o inspirarse sin reproducir una lámina. | Son obligatorias la identidad, firma cromática y cualidades NAVA. Composición, escala, grid, espaciado, componentes y navegación permanecen libres. |
| **Fidelidad al mockup** | Un issue, prompt o instrucción asigna una imagen concreta a la pantalla/componente o pide igualar, reproducir, implementar o corregir contra esa referencia. | La imagen es contrato visual para el viewport y estado representados. Deben reproducirse sus colores, proporciones, jerarquía tipográfica, escalas de controles e iconos, centrado, alineaciones, densidad, bordes y tratamiento de estados. |

En modo de fidelidad no basta con que la pantalla “se sienta NAVA”, use los mismos tokens o contenga los mismos elementos. El resultado renderizado debe compararse con la referencia. La libertad permanece en la técnica de implementación, el reflow de anchos no representados y las decisiones ausentes de la imagen.

En ambos modos se conservan el copy, los datos y las funciones reales. No se copian textos sintéticos, personas, cifras, módulos o acciones que solo aparezcan como decoración del mockup. La implementación puede desviarse por una regla o contrato superior, contenido real, accesibilidad, privacidad, seguridad, responsive o rendimiento; cada desviación se registra con su causa y tratamiento.

### 2.2 Medición y comparación en modo de fidelidad

Antes de editar se abre la imagen a resolución original y se identifica el panel, viewport y estado aplicables. Se registra como mínimo:

- tamaño del panel de referencia y relación de aspecto;
- límites y proporción de marca, título, contenido/formulario y acción principal;
- alineaciones, centrado, distribución de espacio y densidad;
- tamaño y line-height de texto cuando se conozcan, o jerarquía y wrapping observables cuando deban inferirse;
- altura de controles, caja visual de iconos, bordes y espacios principales;
- valores cromáticos normativos y estados representados.

Después de implementar se captura la app real en el mismo viewport y estado. La revisión incluye comparación lado a lado y overlay o diff de imagen; estilos computados y métricas DOM sirven de apoyo, pero nunca sustituyen mirar la comparación.

Como tolerancia de revisión, los límites de regiones principales y alineaciones clave no se apartan más de 4 CSS px o 2 % de la dimensión relevante —lo que sea mayor—; texto y line-height conocidos no se apartan más de 1 CSS px, e iconos equivalentes no más de 2 CSS px. Rasterización de fuentes, antialiasing y longitud del copy real se revisan por su efecto, no se tratan por sí solos como defectos. Una diferencia visible importante no queda aprobada solo por caer dentro de un umbral numérico.

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
| Latón oscuro | `#765C2F` | Foco visible, acción suave y, desde el issue #212, las versalitas espaciadas y filetes de los controles reglados (`--color-accent-brass` en `styles/tokens.css`: rótulos de campo, ranuras de código, palabra de estado de alerta, filete de botón secundario/fantasma). |

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
- `BaseButton` (issue #212, `shared/ui`): radio `2px` en todas las variantes. La primaria usa tinta llena; secundaria y fantasma llevan filete de latón oscuro (`--color-accent-brass`) con el borde inferior acentuado a `2px`, la misma línea base doble-espesor que `BaseInput`.

### 6.2 Campos y formularios

- Todo campo conserva label visible. Placeholder y ayuda complementan, no sustituyen el label.
- La anatomía `label → control → ayuda/error` se mantiene estable y el error aparece junto al campo.
- Requerido y opcional se comunican por texto o marca entendible, no solo por color.
- El foco de teclado es visible y no queda oculto por barras, diálogos o teclado virtual.
- Un formulario extenso puede agruparse por pasos, secciones o columnas. En móvil se transforma a una columna o a pasos completos sin perder datos ni acciones.
- Un error de envío muestra resumen persistente cuando existen varios campos afectados; el foco se mueve de forma útil sin encerrar a la persona.
- **Campo reglado** (issue #212, `BaseInput`/`OtpInput` en `shared/ui`): superficie de papel blanco apoyada en una línea base de tinta de `2px` (`--color-action-primary`), con filete perimetral de `1px` a baja opacidad (`--color-border-subtle`) y radio `2px`. El rótulo va sobre el campo en versalitas espaciadas de latón oscuro (11px, `letter-spacing: 0.08em`, mayúsculas) — el mismo tratamiento que "PASO 1 DE 3" y "ACCESO SEGURO" del mockup de referencia. No lleva iconos decorativos de sobre ni de candado: el rótulo ya nombra el campo; el slot `leading`/`trailing` de `BaseInput` sigue existiendo en la API para quien lo necesite, pero `auth` dejó de usarlo. El conmutador de contraseña es la palabra `Mostrar`/`Ocultar` en versalitas subrayadas, no un glifo de ojo. En estado de error, el filete completo (perímetro y línea base) pasa a color de peligro. `OtpInput` (nuevo, `shared/ui`) repite exactamente esta construcción en seis ranuras estrechas con el dígito compuesto en `--font-display`: un solo valor lógico de seis dígitos hacia afuera, con pegado y distribución automática, avance, retroceso, flechas y `autocomplete="one-time-code"`. Ya está disponible en `shared/ui`, pero ninguna pantalla lo consume todavía — la adopción en `/acceso` y `/recuperar-acceso` es una fase posterior.

### 6.3 Alertas y estados

- Error, conflicto, éxito, información y advertencia usan superficie, borde, icono o forma, título y explicación accionable.
- Un éxito importante permanece dentro de la tarea; no depende solo de un toast efímero.
- Un conflicto conserva las selecciones válidas y explica qué puede hacer la persona.
- Un error inesperado no expone stack, proveedor, token ni datos personales; puede mostrar un `request_id` seguro cuando el contrato lo proporcione.
- **Alerta como nota al margen** (issue #212, `BaseAlert` en `shared/ui`): filete lateral de `4px` del color de estado sobre fondo apenas teñido, radio `2px` — ya no un recuadro con borde perimetral uniforme e icono circular genérico. Sobre el título aparece la palabra de estado en versalitas de latón oscuro: `Error` (`danger`), `Atención` (`warning`), `Nota` (`info`), `Confirmación` (`success`); esa palabra es la que cumple "icono, texto y estructura además del color" (WCAG 2.2 AA 1.4.1). El control de descarte se rotula `Descartar` en vez de un icono con `aria-label` aparte. La variante `plain` (nueva) es la excepción sin relleno para contenido informativo que no comunica un estado del sistema — filete lateral de latón, sin fondo teñido, sin palabra de estado; el título se compone como versalita única en vez de repetirse también como titular en negrita. Es la que usa la caja "Requisitos de la contraseña" de recuperación de acceso.

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

### 6.6 Cabecera de página y carril de registro

- `PageHeader` (`shared/ui`) es la cabecera repetible de toda pantalla privada: título editorial (`h1`, único por vista), contexto opcional bajo el título, enlace de retorno opcional y una región de acciones que se apila en móvil. Sustituye el encabezado ad-hoc que cada pantalla redefinía por separado.
- `RecordRow` (`shared/ui`) es el renglón de carril para listas de registros (barberos, servicios, tramos, bloqueos): divisor fino entre filas en vez de tarjeta blanca con borde, con regiones `leading`/principal/`trailing`. No es interactivo por sí mismo; cuando la fila completa navega, el consumidor coloca el enlace o botón dentro de una de sus regiones.
- Ambos son opcionales: una pantalla que no encaje en el patrón puede seguir componiendo su propio encabezado o lista, siempre dentro de la firma cromática y los patrones de esta sección.

## 7. Zona de libertad creativa

En modo de identidad guiada, cada pantalla o flujo puede decidir libremente:

- composición, columnas, orden visual y uso del espacio;
- tamaños tipográficos, pesos y escala;
- espaciado, grid, radios, bordes, sombras y movimiento;
- header, rail, sidebar, dock, tabs, menú o navegación contextual;
- componentes locales o compartidos;
- CSS estándar, `<style scoped>`, CSS Modules, Tailwind, utility-first, biblioteca visual o una solución híbrida;
- fotografía, ilustración, textura o ausencia de imagen, cuando aporten a la tarea y no contradigan la identidad.

En modo de fidelidad, la misma libertad aplica a la tecnología y a todo lo que el mockup no representa; no autoriza cambiar la geometría, escala o jerarquía visible del panel asignado por preferencia del implementador.

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
7. modo de conformidad (`identidad guiada` o `fidelidad al mockup`).
8. en modo de fidelidad, mediciones objetivo, captura anterior, captura final, comparación lado a lado y overlay/diff en el viewport efectivo comprobado.

La pantalla está terminada cuando:

- pertenece visualmente a la misma familia que los mockups aprobados;
- usa la firma cromática NAVA o derivados permitidos con contraste comprobado;
- conserva la libertad de composición sin romper jerarquía ni operabilidad;
- cubre estados, responsive, teclado, foco y movimiento reducido;
- no amplía el alcance ni contradice contrato o reglas;
- aporta pruebas y evidencia proporcionales.

En modo de fidelidad, además:

- las regiones principales cumplen las tolerancias de la sección 2.2 o documentan una desviación autorizada;
- se verificó el viewport efectivo con `window.innerWidth`/`window.innerHeight` o mecanismo equivalente;
- se revisaron la app real y su comparación visual, no solo el código, los tests, un snapshot aislado o métricas DOM;
- no queda una diferencia visible sin clasificar en color, proporción, tipografía, controles, iconos, centrado, alineación, densidad o estado;
- el resultado se marca explícitamente `PASS` o `FAIL`; `PASS` no es válido mientras exista una diferencia primaria sin explicar.

No se rechaza una solución por usar CSS local, Tailwind u otra herramienta. En identidad guiada tampoco se rechaza por otra composición, tamaño, radio o variante propia. En fidelidad sí se rechaza una desviación visual no justificada respecto del mockup asignado. En ambos modos se rechaza si sustituye la identidad cromática, pierde el lenguaje NAVA, copia funciones no autorizadas o incumple las garantías anteriores.

## 12. Referencias

- `DEC-077`: identidad NAVA y dirección Tailored Grid.
- `DEC-078`: libertad de composición, componentes y herramientas.
- `DEC-079`: mockups y firma cromática obligatorios para rediseños y pantallas nuevas.
- `DEC-080`: dos modos de conformidad y fidelidad medible cuando existe un mockup exacto asignado.
- [Handoff de mockups NAVA](../10-backlog/evidence/ui-redesign-nava-2026-09-02/README.md).
- [Atlas integral NAVA / Tailored Grid](../10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/README.md).
- [Especificación de frontend NAVA](especificacion-frontend-nava.md).
- [Estándar de frontend Vue](estandar-frontend-vue.md).
- [Estrategia de pruebas](estrategia-pruebas.md).
- [WCAG 2.2](https://www.w3.org/TR/WCAG22/).
