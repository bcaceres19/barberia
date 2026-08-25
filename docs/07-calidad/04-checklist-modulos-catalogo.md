---
titulo: "Checklist de configuración, barberos y catálogo de servicios"
version: "1.0"
estado: "Herramienta operativa, no normativa"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-25"
documentos_relacionados:
  - "README.md"
  - "02-checklist-transversal.md"
  - "../01-producto/reglas-negocio.md"
---

# Checklist de configuración, barberos y catálogo de servicios

Cubre `SettingsPage`, `StaffPage`, `CatalogPage` (incluido el ciclo de vida `HU-024`) y `BarberServicesPage`. Ejecuta primero [`02-checklist-transversal.md`](02-checklist-transversal.md) sobre cada pantalla.

## 1. Configuración de la barbería (`SettingsPage`)

Límites reales (`settingsValidation.ts`): nombre máx. 120, zona horaria máx. 64, correo de contacto máx. 254 (vacío = válido = "sin contacto"), teléfono en formato E.164 estricto `+[1-9][0-9]{7,14}` (vacío = válido).

- [ ] Nombre de 120 caracteres exactos (aceptado) vs. 121 (rechazado).
- [ ] Zona horaria vacía (rechazada), de 64 caracteres (aceptada), de 65 (rechazada en cliente), y una cadena con forma de zona pero **no reconocida por PostgreSQL** (por ejemplo `America/Nowhere`): el cliente la deja pasar (no valida contra IANA), así que el 422 real llega del servidor — confirmar que se muestra como **error general del formulario**, no como error del campo (así lo documenta el propio modelo), y que no queda una escritura parcial.
- [ ] Correo de contacto: vacío (válido, debe guardarse como `null`, nunca como `""`), formato inválido (`admin@`, `admin@dominio`, `@dominio.com`), 254 caracteres válidos vs. 255.
- [ ] Teléfono de contacto: vacío (válido → `null`), sin `+` inicial, con espacios o guiones (`+57 300 123 4567`, `+57-300-1234567`) — el patrón es estricto, verificar que se **rechaza** y no se guarda "silenciosamente limpiado".
- [ ] Guardar la sección de configuración deja el resto de secciones del panel (barberos, catálogo) sin tocar — no existe un único botón "Guardar" global (`estandar-diseno-visual.md` §10): confirmar que cada sección persiste de forma independiente.
- [ ] Dos pestañas con la sesión abierta, cada una edita un campo distinto de la misma sección y guarda casi al mismo tiempo: ¿el segundo guardado sobrescribe por completo el primero (incluidos los campos que la segunda pestaña no tocó, porque envía el formulario completo), o se pierde el cambio de la primera pestaña sin aviso?
- [ ] Cambiar el nombre de la barbería: la cabecera (`AppHeader`) se actualiza sin recargar la página completa.

## 2. Barberos (`StaffPage`)

Límite real (`staffValidation.ts`): nombre completo máx. 120 caracteres. **No exige unicidad** — dos barberos con el mismo nombre son válidos a propósito.

- [ ] Nombre vacío o solo espacios → rechazado. Nombre de 120 caracteres exacto → aceptado; 121 → rechazado.
- [ ] Crear **dos barberos con el nombre idéntico** exacto (permitido por diseño): verificar que la lista, el selector de `BarberServicesPage` y cualquier otro lugar que los muestre los distingue de forma inequívoca (por ejemplo por orden estable o algún dato adicional) — un usuario real no debe poder confundirlos ni asignarle un servicio al barbero equivocado por ambigüedad de la interfaz.
- [ ] Nombre con emoji/unicode (`José 🪒 Ñ`) y nombre con `<script>`: se guarda y se muestra igual (texto literal) en la lista de barberos y en el selector de `BarberServicesPage`.
- [ ] Crear varios barberos hasta cruzar el límite de una página de la lista paginada por cursor: navegar hacia adelante y hacia atrás no debe duplicar ni saltarse un barbero, incluso si se crea uno nuevo mientras se está paginando.
- [ ] Doble clic rápido en "Añadir barbero" con el mismo nombre completado: ¿crea uno o dos barberos? (staff no usa `Idempotency-Key` como sí lo usa `catalog` — confirmar cuál es el comportamiento real, es un caso de interés específico).

## 3. Catálogo de servicios (`CatalogPage`, incluye ciclo de vida `HU-024`)

Límites reales (`catalogValidation.ts`): nombre máx. 120, descripción máx. 500 (opcional), duración entera entre 1 y 1440 minutos, precio con patrón `^[0-9]{1,10}(\.[0-9]{1,2})?$` y mayor que cero estricto. Moneda siempre `COP`, no editable. Nombre único **solo entre servicios activos** de la misma barbería (`DEC-067`).

### Campos

- [ ] Nombre: vacío, 120 (aceptado), 121 (rechazado), duplicado exacto de un servicio **activo** existente (rechazado, `name-conflict`), duplicado exacto de un servicio **inactivo/desactivado** (¿se acepta, según la regla "único solo entre activos"? — verificar explícitamente, es el caso de interés más filoso de esta pantalla).
- [ ] Descripción: vacía (válida), 500 (aceptada), 501 (rechazada), con saltos de línea, con HTML.
- [ ] Duración: `0` y negativa (rechazadas), `1` y `1440` (aceptadas), `1441` (rechazada), no entera (`30.5`, rechazada por el patrón de solo dígitos), con ceros a la izquierda (`007`, ¿se interpreta como 7?), texto no numérico (`treinta`).
- [ ] Precio: vacío (rechazado), `0` (rechazado, debe ser mayor que cero), negativo (rechazado por el patrón, que no admite signo), `45000.50` (aceptado), `45000.999` (rechazado por el patrón, solo 2 decimales), `45,000` con coma de miles (rechazado por el patrón), 10 dígitos enteros exactos (aceptado, límite del patrón `{1,10}`) vs. 11 dígitos (rechazado), notación científica `4.5e4` (rechazada por el patrón).
- [ ] Moneda: confirmar que la interfaz nunca ofrece un control para cambiarla y que ninguna petición manual desde el propio navegador (editando el payload) logra crear un servicio con otra moneda sin que el servidor la reescriba a `COP` o la rechace.

### Ciclo de vida (`HU-024`, `RN-SER-03`)

- [ ] Desactivar un servicio activo: la vista previa de impacto (`affectedAppointments`) se muestra **antes** de confirmar. Como `appointment` todavía no existe en este bloque (`DEC-069`), el valor siempre debería ser `0` — confirmar que la pantalla lo maneja como un caso normal (recuento cero), no como un error o un estado vacío raro.
- [ ] Doble clic rápido en "Desactivar" sobre el mismo servicio: debe quedar desactivado una sola vez, sin dos entradas de historial ni un error confuso en el segundo clic (`idempotency-conflict` esperado y bien comunicado, no un 500).
- [ ] Desactivar un servicio ya desactivado (por ejemplo reabriendo una pestaña vieja con el botón todavía visible): `transition-conflict` esperado — verificar que el mensaje mostrado es comprensible, no un error técnico crudo.
- [ ] Reactivar un servicio ya activo (mismo caso, otra pestaña vieja): mismo criterio, `transition-conflict` con mensaje claro.
- [ ] Reactivar un servicio desactivado **cuyo nombre ahora choca** con un servicio activo nuevo creado mientras estaba desactivado (alguien reutilizó el nombre): ¿qué pasa? Es un caso límite no cubierto explícitamente por `DeactivateServiceOutcome`/`ReactivateServiceOutcome` en el modelo — documentar el comportamiento real observado, y si es un 500 o un estado inconsistente, es un hallazgo `Alto` o `Bloqueante`.
- [ ] Un servicio recién desactivado, ¿sigue apareciendo como asignable en `BarberServicesPage`? Ver también §4 de este archivo — es la intersección entre dos módulos y el tipo de bug que una prueba unitaria de un solo módulo no puede ver.
- [ ] Servicio desactivado: confirmar que **desaparece** del enlace público futuro (hoy no existe ese enlace todavía, pero si existe algún listado "activo" en el panel, confirmar que ya no aparece ahí) y que su historial (fecha de creación, etc.) sigue siendo consultable.

## 4. Servicios por barbero (`BarberServicesPage`, `HU-023`)

- [ ] Marcar la misma casilla (asignar) dos veces con clics muy rápidos: una sola fila de asignación creada, no una petición duplicada visible como error.
- [ ] Asignar un servicio a un barbero, y en otra pestaña **desactivar ese mismo servicio** desde `CatalogPage`; volver a la pestaña de asignación sin recargar: ¿la casilla sigue marcada como si nada, o refleja de algún modo que el servicio ya no está activo? Si no refleja nada, verificar al menos que la próxima carga real de la lista sí lo hace correctamente.
- [ ] Intentar desasignar la **última** asignación activa de un servicio activo (`DEC-068`: se rechaza retirar la última asignación de un servicio activo) — confirmar que el rechazo se comunica con un mensaje claro y no como un 500 ni un error genérico de red.
- [ ] Dos desasignaciones concurrentes sobre la misma última asignación (dos pestañas, clic casi simultáneo): exactamente una debe tener éxito o ambas deben fallar de forma controlada; nunca debe quedar el servicio activo sin ningún barbero asignado por una condición de carrera.
- [ ] Con muchos barberos y muchos servicios (cruzando el límite de una página en cualquiera de los dos selectores paginados por cursor): asignar mientras se pagina no debe duplicar ni perder una fila visible.
- [ ] Barbero o servicio con nombre idéntico a otro (ver §2 y §3): confirmar que el selector de esta pantalla permite distinguir cuál es cuál sin ambigüedad antes de asignar.
