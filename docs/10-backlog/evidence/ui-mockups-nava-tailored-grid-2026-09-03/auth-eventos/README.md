# Mockups de acceso y recuperación por viewport y evento

## Propósito

Este directorio reemplaza como referencia de implementación la lámina compuesta `02-acceso-recuperacion.png`. La lámina original se conserva como historial visual, pero no debe entregarse a un agente de implementación: cada archivo de este directorio representa una sola pantalla y un solo evento para reducir ambigüedad.

La estructura es deliberadamente simétrica:

- `desktop/acceso/` y `mobile/acceso/` contienen los mismos doce eventos de `/acceso`;
- `desktop/recuperacion/` y `mobile/recuperacion/` contienen los mismos once eventos de `/recuperar-acceso`;
- un mismo número y nombre describen el mismo estado funcional en ambos viewports;
- escritorio hace reflow al cascarón dividido NAVA y móvil usa una sola columna; nunca se miniaturiza una vista dentro de la otra.

## Acceso

| N.º | Archivo en cada viewport | Evento representado |
| --- | --- | --- |
| 01 | `01-inicial.png` | Formulario de acceso en reposo. |
| 02 | `02-enviando.png` | Envío en curso, campos y acción bloqueados. |
| 03 | `03-validacion.png` | Validación local de correo y contraseña vacíos. |
| 04 | `04-credenciales-invalidas.png` | Rechazo uniforme de credenciales, sin enumerar cuentas. |
| 05 | `05-sin-conexion.png` | Error recuperable de red con reintento y datos preservados. |
| 06 | `06-sesion-vencida.png` | Retorno al acceso por sesión vencida. |
| 07 | `07-reto-envio-correo.png` | Reto adicional con entrega por correo configurado, canal predeterminado. |
| 08 | `08-reto-envio-whatsapp.png` | Reto adicional con entrega por WhatsApp oficial configurado. |
| 09 | `09-reto-envio-ambos.png` | Reto adicional con entrega del mismo código por correo y WhatsApp. |
| 10 | `10-reto-codigo-invalido-correo.png` | Código uniforme no válido en la variante correo. |
| 11 | `11-reto-codigo-invalido-whatsapp.png` | Código uniforme no válido en la variante WhatsApp. |
| 12 | `12-reto-codigo-invalido-ambos.png` | Código uniforme no válido en la variante de ambos canales. |

## Recuperación de acceso y cambio de contraseña

| N.º | Archivo en cada viewport | Evento representado |
| --- | --- | --- |
| 01 | `01-solicitud-inicial.png` | Paso 1: solicitud del código en reposo. |
| 02 | `02-solicitud-validacion.png` | Paso 1: correo vacío o inválido. |
| 03 | `03-solicitud-enviando.png` | Paso 1: solicitud en curso. |
| 04 | `04-solicitud-sin-conexion.png` | Paso 1: error recuperable de red. |
| 05 | `05-verificacion-inicial.png` | Paso 2: ingreso del código y contador de reenvío. |
| 06 | `06-verificacion-codigo-invalido.png` | Paso 2: código incorrecto, vencido o agotado. |
| 07 | `07-contrasena-inicial.png` | Paso 3: contraseña nueva con política visible. |
| 08 | `08-contrasena-validacion.png` | Paso 3: longitud insuficiente y confirmación distinta. |
| 09 | `09-contrasena-enviando.png` | Paso 3: actualización en curso. |
| 10 | `10-contrasena-enlace-vencido.png` | Paso 3: token de reinicio inválido o vencido. |
| 11 | `11-completado.png` | Recuperación completada y sesiones revocadas. |

## Contrato visual común

- Paleta aprobada: tinta `#101B2B`, marfil `#F4F0E7`, blanco `#FFFFFF`, grafito `#2A2D32`, piedra `#C9C0B2`, latón `#B8955A` y salvia para éxito.
- El wordmark NAVA, el título y el formulario permanecen centrados y a escala legible. La regla de latón abarca el ancho completo del wordmark.
- Los labels no muestran asteriscos. La obligatoriedad se comunica semánticamente y mediante validación accesible.
- `Input`, `PasswordInput`, `OtpInput` y `InlineAlert` son contratos visuales compartidos: acceso y recuperación reutilizan los mismos colores, borde, altura, tipografía, foco, error y estado deshabilitado. No se crean versiones locales con otro tono o escala.
- **Campo reglado.** El control es una superficie de papel blanco apoyada en una línea base de tinta de `2 px`, con filete perimetral de `1 px` a baja opacidad y radio `2 px`. El rótulo va sobre el campo en versalitas espaciadas de latón, el mismo tratamiento de `PASO 1 DE 3` y `ACCESO SEGURO`. No lleva iconos decorativos de sobre ni de candado: el rótulo ya nombra el campo. El conmutador de contraseña es la palabra `Mostrar` en versalitas, no un glifo de ojo; así el control tiene nombre accesible propio sin depender de una forma.
- **Ranuras de código.** Las seis posiciones repiten exactamente la construcción del campo reglado en formato estrecho, y el dígito se compone en la serif del wordmark apoyado en la línea base. No son un componente aparte con caja redondeada: pertenecen a la misma familia que los demás controles. La implementación debe conservar un valor lógico de seis dígitos y soportar pegado completo, distribución automática, avance, retroceso, teclado numérico, foco visible y un nombre accesible coherente.
- **Alerta como nota al margen.** Filete lateral de `4 px` del color de estado sobre fondo apenas teñido, radio `2 px`, y una palabra de estado en versalitas —`Error`, `Atención`, `Nota`, `Confirmación`— sobre el título. Esa palabra es la que cumple “icono, texto y estructura además del color”: no se usa un glifo circular genérico. La acción de recuperación dentro de la alerta hereda el mismo filete inferior de los demás controles.
- **Sección del reto.** El reto adicional no es una tarjeta sobrepuesta: abre una sección de la misma columna con la regla de latón y rombo que ya divide el panel oscuro, seguida del canal configurado en versalitas. Se lee como continuación del formulario, no como un módulo pegado encima.
- La contraseña nueva admite entre 10 y 128 caracteres y debe diferir del correo y de la contraseña actual. No se inventan requisitos de mayúsculas, números o símbolos.
- El error específico vive inmediatamente debajo de su campo, nunca dentro del control. El resumen global explica qué puede hacer la persona y no repite literalmente cada error local.
- En `/acceso`, cualquier alerta global aparece después del botón y de `¿Olvidaste tu contraseña?`, exactamente en el ancla usada por `06-sesion-vencida.png`; no aparece entre el título y los campos. En recuperación se conserva el mismo patrón después del grupo de acciones y de `Volver al acceso` cuando ese enlace exista.
- Las alertas se expresan con icono, texto y estructura además del color. Entran en `160 ms` con opacidad `0 → 1` y traslación vertical `-6 px → 0`, curva `ease-out`, sin desplazar el foco ni reservar menos espacio del necesario. Con `prefers-reduced-motion: reduce` se elimina la traslación y la aparición es inmediata o una transición de opacidad no esencial de hasta `80 ms`.
- Durante cualquier envío, escritorio y móvil bloquean **todos** los campos de la operación y la acción principal, cambian el verbo a gerundio con indicador de progreso, impiden doble envío y conservan la geometría.
- El canal del OTP no lo escribe ni lo elige mediante un destino arbitrario la persona en esta pantalla. El servidor resuelve contactos verificados según la configuración del evento: correo por defecto, WhatsApp oficial o ambos. La respuesta y el error permanecen uniformes para no enumerar cuentas.
- Los datos escritos son ficticios y los destinos aparecen enmascarados.
- Cada pantalla muestra una sola acción primaria. Cuando el reto adicional está activo, las credenciales y `Iniciar sesión` quedan bloqueadas y la acción primaria pasa al reto; no compiten dos botones oscuros en la misma vista.
- El estado de los campos no contradice al mensaje: un código rechazado se muestra con las seis casillas llenas y en error; un rechazo de credenciales limpia la contraseña y conserva el correo (`CA-010-02`); un fallo de red conserva ambos datos y ofrece `Reintentar`; una sesión vencida llega con los campos vacíos.
- Los rótulos de acción y los mensajes de validación son los del módulo real `apps/web/src/modules/auth`. Cuando un mockup necesita un texto que el código todavía no tiene, procede de una decisión `DEC-*` registrada.
- `Volver al acceso` solo aparece en `/recuperar-acceso`. El reto de `/acceso` no lo incluye porque la persona ya está en esa ruta.
- Ninguna línea de texto se sale de su contenedor en ningún viewport: el generador falla si detecta desborde horizontal o si un evento de escritorio no cabe en `1024 px` de alto.

## Uso por agentes de implementación

1. Seleccionar una sola ruta, viewport y evento.
2. Abrir únicamente el PNG correspondiente y este README; no usar la lámina compuesta como objetivo.
3. Comparar la app real y el PNG en el mismo viewport efectivo.
4. Corregir por ciclos cortos y guardar captura, comparación lado a lado y overlay o diff.
5. Repetir el mismo número en el viewport hermano antes de declarar el evento terminado.

El mockup gobierna composición, color, proporción, jerarquía, centrado, alineación, densidad y estado visual. Las HU, reglas, decisiones, contrato API y código real gobiernan el comportamiento y el texto seguro cuando exista una discrepancia funcional.

## Procedencia

La primera separación se generó con ImageGen en modo de edición/composición precisa usando como fuentes `00-referencia-tailored-grid.png`, `02-acceso-recuperacion.png`, `14-componentes-formularios-alertas.png` y los estados reales del módulo `apps/web/src/modules/auth`. Tras agotarse temporalmente esa herramienta, la revisión del 2026-09-03 se recompuso de forma determinista. La estructura separada, las seis casillas OTP, el ancla de alertas, el bloqueo durante envío y las tres variantes de canal fueron confirmados por el propietario el 2026-09-03.

Los 46 archivos se regeneran con [`tools/mockups/auth-eventos`](../../../../../tools/mockups/auth-eventos): `content.mjs` declara cada evento una sola vez y `render.mjs` lo compone en HTML con los tokens de `apps/web/src/styles/tokens.css` y las fuentes auto-hosteadas del proyecto, y lo rasteriza con el Chromium de Playwright que ya usa la suite e2e.

```bash
node tools/mockups/auth-eventos/render.mjs
```

Un evento se declara una vez y se dibuja en los dos viewports, así que escritorio y móvil no pueden divergir en copy, orden ni estado. Se usa HTML en lugar de SVG posicionado a mano porque el motor de maquetación resuelve el ajuste de línea real: el desborde de texto fuera de la alerta en las variantes móviles del reto venía de líneas colocadas manualmente. El generador aborta con código distinto de cero si algún evento desborda.

### Correcciones de la revisión del 2026-09-03

Un segundo pase reemplazó el lenguaje de los controles. Campos, casillas de código y alertas venían de un vocabulario genérico de kit de formularios —cajas blancas redondeadas con iconos de sobre, candado y ojo; seis recuadros de OTP dentro de una tarjeta con etiqueta en píldora oscura; alertas pastel con glifo circular— que no pertenecía a la composición editorial de NAVA y hacía que el reto pareciera un añadido de otro diseño. Los tres se reconstruyeron sobre reglas, versalitas y la serif del wordmark, según describe el contrato visual de arriba.

Sobre la primera composición separada se corrigieron: rótulos y mensajes que no coincidían con `apps/web/src/modules/auth` (`Ingresar` → `Iniciar sesión`, `Escribe un correo válido.` → `Escribe tu correo.`, entre otros); alertas colocadas después de la acción que explicaban en `10-contrasena-enlace-vencido` y `11-completado`; estados de campo que contradecían su mensaje (código rechazado con casillas vacías, credenciales rechazadas con la contraseña aún escrita, sesión vencida con datos precargados); el desborde de texto de las alertas móviles del reto; el doble botón primario y el `Volver al acceso` fuera de contexto en el reto de `/acceso`; la falta del botón `Reintentar` ante un fallo de red; el wordmark ausente en recuperación; y el ritmo vertical, que pegaba cada error en línea al label siguiente y las casillas OTP a su botón.
