---
titulo: "Checklist transversal (toda pantalla y flujo)"
version: "1.0"
estado: "Herramienta operativa, no normativa"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-25"
documentos_relacionados:
  - "README.md"
  - "01-metodologia-y-uso.md"
  - "../03-desarrollo/estandar-diseno-visual.md"
  - "../01-producto/reglas-negocio.md"
  - "../03-desarrollo/estrategia-pruebas.md"
---

# Checklist transversal

Aplica a **cada** pantalla privada y cada formulario, además de los checks específicos de [`03`](03-checklist-autenticacion-sesion.md) y [`04`](04-checklist-modulos-catalogo.md). Repite este bloque completo por pantalla antes de darla por explorada.

## A. Responsive y visual (`estandar-diseno-visual.md` §7, §16)

- [ ] **320 px**: reflow sin scroll horizontal ni pérdida de contenido o de una acción necesaria.
- [ ] **360 px**: mismo criterio que 320 px; es el ancho de referencia de los E2E `*-evidencia-responsiva.spec.ts` ya existentes — compara contra esa evidencia si algo se ve distinto.
- [ ] **768 px**: la composición puede cambiar (una→dos columnas); verificar que no queda contenido huérfano ni duplicado a mitad de transición.
- [ ] **1024–1279 px**: panel con navegación lateral; el contenido no se estira sin límite.
- [ ] **1280 px+**: contenido centrado, máximo 1440 px en privado — verificar que no queda una franja vacía absurda ni el formulario estirado a todo el ancho.
- [ ] **Zoom de texto al 200 %**: ningún control ni texto queda cortado o inaccesible.
- [ ] **Nombre/texto en el límite máximo del campo** (ver longitudes exactas en [`05-matriz-combinaciones.md`](05-matriz-combinaciones.md)) mostrado en cada lugar donde se repite: lista, tarjeta, selector, insignia, cabecera — ¿trunca con elipsis legible, desborda la tarjeta, o rompe el layout de la fila?
- [ ] **Objetivo táctil ≥ 44×44 px** en cada botón/control, incluida la fila completa si es interactiva.
- [ ] **Foco visible** (anillo `#2563EB`) al navegar solo con teclado, incluido dentro de un diálogo.
- [ ] **`prefers-reduced-motion`** activado: transiciones y el giro de carga se atenúan o desaparecen, no se rompe la interfaz.
- [ ] **Color nunca es la única señal**: todo estado, error o selección tiene también texto o ícono.
- [ ] Los cinco estados de la pantalla existen y son distinguibles entre sí, no un mismo spinner genérico para todo: **inicial, carga, vacío, error recuperable, éxito**.

## B. Errores HTTP y manejo de fallas

- [ ] **Backend caído / 500**: mensaje seguro sin nombre de excepción ni stack, `request_id` visible y copiable, opción de reintentar.
- [ ] **Sesión vencida a mitad de un formulario** (401 en el envío): los datos no sensibles ya escritos no se pierden silenciosamente; redirección clara a iniciar sesión.
- [ ] **Red cortada durante el envío** (throttle a offline en las herramientas de desarrollo) y luego reintento: no debe quedar un recurso duplicado ni dos en curso.
- [ ] **Red lenta** (throttle a "Slow 3G"): el botón muestra un estado de carga explícito ("Guardando…", no solo deshabilitado en silencio) y bloquea reenvíos mientras está en curso.
- [ ] **Body/campo con una cadena extremadamente larga** (10 000+ caracteres pegados en un campo de texto libre): no debe colgar el navegador ni provocar un 500 sin mensaje.

## C. Idempotencia y doble acción (`RN-IDE-01`)

- [ ] **Doble clic rápido** sobre cada botón de acción primaria (crear, guardar, desactivar/reactivar, asignar): debe producir un único efecto, nunca dos recursos ni dos entradas de historial.
- [ ] Enviar un formulario, y **antes de que responda el servidor**, navegar a otra pantalla y volver: ¿la operación se completó, se duplicó o quedó en un limbo visible como pendiente para siempre?
- [ ] Doble envío con **la misma clave de idempotencia y contenido distinto** (solo aplica a `catalog`, que sí usa `Idempotency-Key`, ver `apps/web/src/modules/catalog/model/idempotencyKey.ts`): debe rechazarse como conflicto, no ejecutarse con el contenido nuevo.

## D. Aislamiento entre barberías (`RN-TEN-01`)

- [ ] Con dos sesiones abiertas en navegadores/perfiles distintos (barbería A y barbería B), copiar el identificador de un recurso de A (servicio, barbero, asignación) y sustituirlo en una petición o URL de B: debe responder **404**, nunca 403 (P10 de `estados-citas.md` — un 403 revelaría que el recurso existe).
- [ ] Un dato de A (nombre de servicio, nombre de barbero) nunca debe aparecer en ningún selector, lista o dropdown de B, ni siquiera brevemente durante una carga.

## E. Flujos no esperados

- [ ] **Botón "Atrás" del navegador** después de una mutación (crear/editar/desactivar): la lista mostrada, ¿queda desactualizada o se refresca correctamente al volver a esa vista?
- [ ] **Recargar la página (F5) a mitad de un formulario**: los datos no guardados se pierden, como es esperable, pero la pantalla no debe quedar en un estado roto (carga infinita, formulario vacío sin poder reintentar).
- [ ] **Navegación directa por URL** a una ruta privada sin sesión activa: redirección a login conservando el destino pretendido.
- [ ] **Dos pestañas con la misma sesión**: hacer logout en una, luego intentar una acción en la otra — debe fallar de forma clara (401 → redirección), no ejecutar la acción con una sesión ya inválida.
- [ ] **Redimensionar la ventana con un diálogo abierto**: el diálogo no debe quedar cortado, desplazado fuera de la vista o con el foco perdido.
- [ ] **Cerrar un diálogo con Escape** y con clic fuera: el foco vuelve al elemento que lo abrió.

## F. Almacenamiento e integridad de datos

- [ ] Crear/editar un recurso y **recargar con F5** (no solo confiar en el estado en memoria de Vue): el valor mostrado coincide exactamente con lo enviado, carácter por carácter.
- [ ] El mismo dato mostrado en **dos lugares distintos** (lista y selector de otro módulo, por ejemplo un barbero en `staff` y en `barberServices`) coincide siempre — sin una copia desactualizada.
- [ ] Un valor con **espacios al inicio/final únicamente** (`"   "`) se trata como vacío por cada validador (todos usan `trim()` en este proyecto) — verificar que ningún campo lo acepte como "con contenido".
- [ ] Un valor con **HTML/script** (`<script>alert(1)</script>`, `<img src=x onerror=alert(1)>`) guardado y mostrado después: debe verse como texto literal, nunca ejecutarse ni romper el layout de la fila que lo muestra.
- [ ] Un valor con **unicode/emoji** (`Bárbería 🪒 Ñoño`) se guarda y se muestra igual en cada lugar donde aparece, sin mojibake ni recorte a mitad de un carácter multibyte.
