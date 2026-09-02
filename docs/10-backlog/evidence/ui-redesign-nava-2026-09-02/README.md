# Mockups de rediseño NAVA

## Propósito

Estas láminas traducen la dirección NAVA / Tailored Grid a referencias visuales para autenticación, formularios, estados, alertas, modales y navegación responsive. Se generaron a partir de la auditoría visual del 2 de septiembre de 2026 y quedaron aprobadas como familia visual para rediseños y pantallas nuevas mediante `DEC-079`.

Son referencia obligatoria de identidad, firma cromática, contraste tipográfico y tratamiento de patrones para las familias que representan. No son una especificación de píxeles, tamaños, grid, radios, nombres de tokens, biblioteca o tecnología de estilos. La implementación debe conservar contratos, reglas de negocio, vocabulario, alcance y pruebas vigentes.

## Artefactos

| Lámina | Alcance | Decisiones que ilustra |
| --- | --- | --- |
| [`01-autenticacion-movil.png`](01-autenticacion-movil.png) | Acceso inicial, validación y recuperación | Una tarea por vista, jerarquía editorial, labels persistentes, errores con icono/texto/borde y CTA inequívoco. |
| [`02-nuevo-turno-responsive.png`](02-nuevo-turno-responsive.png) | Formulario de nuevo turno en escritorio, móvil y conflicto | Agrupación semántica por secciones, reflow a una columna, resumen de errores persistente y conflicto recuperable sin perder contexto. |
| [`03-componentes-formularios-alertas.png`](03-componentes-formularios-alertas.png) | Botones, campos, estados, alertas y modal | Anatomía coherente, foco visible, estados distinguibles sin depender solo del color, jerarquía de acciones y confirmación destructiva explícita. |
| [`04-servicios-configuracion-navegacion.png`](04-servicios-configuracion-navegacion.png) | Servicios, alta, configuración y shell responsive | Densidad operativa en escritorio, formulario móvil, éxito persistente y navegación móvil sin etiquetas truncadas mediante `Más`. |

## Lectura para implementación

- Conservar la sensación de sastrería contemporánea y hospitalidad boutique: precisión, calma, calidez, jerarquía editorial y bajo ruido visual.
- Usar las composiciones como punto de partida. Una pantalla puede cambiar escala, espaciado, radios, profundidad o distribución si mejora el contenido; debe conservar la firma cromática y la relación serif/sans definidas en el estándar visual.
- Mantener una acción principal clara por región. Las acciones destructivas deben nombrar su resultado y pedir confirmación cuando la regla lo exija.
- Los campos deben conservar label, ayuda, error y estado de foco sin saltos ambiguos. El foco de teclado debe ser claramente visible.
- Los errores, conflictos y éxitos importantes permanecen junto a la tarea; no dependen solo de un toast ni del color.
- En móvil se hace reflow, no miniaturización. Ninguna etiqueta, acción o contenido necesario debe quedar truncado u oculto.
- La navegación de la cuarta lámina propone cuatro destinos visibles y `Más` para los restantes. Es una solución visual posible a la truncación observada, no un nuevo módulo ni una regla obligatoria.
- Los ejemplos de contenido son sintéticos. Antes de implementar, validar copy, formatos, opciones, obligatoriedad y estados contra la HU, el contrato y las reglas vigentes.
- Estos mockups no autorizan pagos, caja, ingresos, reportes, inventario, cuentas de cliente, vista multi-barbero ni personalización por tenant.
- No implican adoptar Tailwind ni otra dependencia. La herramienta de estilos se decide y justifica en el issue de implementación correspondiente.
- La fuente normativa de colores, libertades y excepciones es [`docs/03-desarrollo/estandar-diseno-visual.md`](../../../03-desarrollo/estandar-diseno-visual.md), no los valores aproximados que puedan extraerse de los píxeles de una imagen.

## Verificación esperada al implementar

- Evidencia en 320, 360, 768 y 1280 px, además del teléfono del piloto.
- Zoom de texto al 200 %, teclado completo, foco/restauración de foco y `prefers-reduced-motion`.
- Contraste WCAG 2.2 AA de las combinaciones realmente implementadas.
- Estados de carga, vacío, error, conflicto, éxito y deshabilitado cuando apliquen.
- Pruebas de componente para interacción y semántica; E2E para el recorrido P0 afectado.
- Controles táctiles cómodos, con objetivo de 44 × 44 px cuando la composición lo permita.

## Prompt set ejecutado

Se usó el modo integrado de ImageGen con la taxonomía `ui-mockup` y las cinco capturas principales de la auditoría como referencias de estructura y contenido. El bloque compartido pidió interfaz de producto implementable, español, identidad NAVA / Tailored Grid, superficies marfil, tinta azul marino, grafito, latón y salvia como base cromática, jerarquía editorial, controles accesibles, foco visible, estados redundantes y exclusión de funciones fuera del MVP.

Las cuatro solicitudes específicas fueron:

1. Tres vistas móviles de autenticación: acceso inicial, validación y recuperación.
2. Nuevo turno en escritorio y móvil, con resumen de errores y conflicto de horario.
3. Lámina de referencia para botones, campos, estados, alertas y modal destructivo.
4. Servicios, alta de servicio, configuración guardada y navegación responsive sin truncamiento.

La primera lámina recibió una edición puntual para retirar iconografía de tijeras y conservar el texto vigente de recuperación por WhatsApp y correo. No se utilizó la vía CLI ni se modificó código de `apps/web`.
