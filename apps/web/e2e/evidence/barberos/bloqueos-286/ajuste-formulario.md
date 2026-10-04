# Ajuste del formulario de bloqueo puntual · #286

Verificación del 2026-10-02 en modo de identidad guiada NAVA (`DEC-105`).

- Cabecera compacta: se elimina la ficha repetida; el formulario conserva el profesional en su nombre accesible y la ventana contextual.
- Campos sobre tinta, bordes discretos y línea base de latón; corrección del selector CSS que antes no alcanzaba al input real.
- Selectores puntual y semanal con lista propia, selección marcada, flechas, Home/End, Enter, Tab, Escape y cierre exterior. El primer Escape solo cierra la lista.
- Calendario personalizado existente promovido a `shared/ui/BaseDatePicker`, compartido con la agenda; fechas civiles y «Hoy» siempre en la zona de la barbería.
- Selector horario compacto en una sola vista: dos columnas para hora y minutos, escritura numérica directa y ajuste con flechas del teclado, sin listas de números. Hora 00–23, minutos exactos 00–59, avance de foco al completar dos dígitos de hora y borrador que solo se confirma con «Aplicar hora».
- Proporciones del selector horario ajustadas al campo: ancho del trigger entre 192 y 240 px y panel de 204 px de alto (antes 280 × 398 px), cifras de 24 px y edición/acciones de 44 px. La prueba responsive comprueba el ancho relativo al campo y una altura inferior a 220 px.
- Paneles flotantes fuera del scroll del modal, dentro del viewport, con foco contenido, cierre exterior y Escape que conserva el formulario.
- Capturas `*-formulario.png`, `*-selector.png`, `*-calendario.png`, `*-hora.png` y `*-minutos.png` en 320, 360, 768 y 1280 px y viewport 640 × 450 equivalente a zoom 200 %. Dimensiones efectivas verificadas y sin desborde horizontal.
- Inspección visual de escritorio, móvil y zoom; axe sin infracciones en formulario, lista, calendario y selector horario compacto, foco y movimiento reducido comprobados.
- 22 pruebas de componente afectadas aprobadas en el último ajuste (hora y bloqueos); las pruebas de calendario y agenda se verificaron al promover el calendario compartido. Cuatro E2E reales aprobados en Chromium escritorio y móvil: creación puntual y semanal, retiro, error con datos conservados, paginación e aislamiento con dos tenants; selección civil y recarga de la agenda comprobadas con el calendario compartido. El selector horario conserva pruebas de escritura parcial, flechas, cancelación, medianoche, validación y minutos exactos; la edición y el ajuste por teclado de minutos se comprueban en los cinco viewports.
- Formato, lint sin advertencias, tipos y build aprobados.

El entorno local necesitó los fixtures sintéticos del issue y sesiones de prueba en `.auth` (ignorado); no se alteraron límites de login ni cuentas reales. No se modificó el contrato HTTP ni la base de datos de producto.
