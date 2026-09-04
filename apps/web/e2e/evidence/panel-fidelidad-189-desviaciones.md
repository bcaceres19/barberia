# Desviaciones visuales — issue #189

## Ejecución mock v2 — 2026-09-03

La evidencia de esta iteración usa únicamente respuestas sintéticas de
Playwright. El reloj está fijado a `2026-09-03T16:15:00Z`, que sitúa el
marcador «Ahora» a las 11:15 en `America/Bogota`. No se inició sesión ni se
consultó PostgreSQL, API, OTP ni datos reales.

Se capturaron los doce eventos del atlas en los viewports que declara su
generador: escritorio `1440×1024 @1x` y móvil `420×935 @2x`. Las capturas
están en `mock/desktop/` y `mock/mobile/`.

| Área | Estado | Motivo y seguimiento |
| --- | --- | --- |
| Agenda: controles, línea temporal, filas y estados | Ajustada | El mock usa los mismos barbero, horarios, nombres y estados ficticios del evento 01. El selector tiene ancho fijo en escritorio; fecha conserva sus tres controles en móvil; las filas presentan hora, persona, servicio y estado en la jerarquía del atlas. |
| Carril y marcador «Ahora» | Ajustada | El rango toma las horas reales de las citas y una hora de contexto por cada extremo, como el atlas. El reloj de la evidencia se fija para no depender del tiempo de ejecución. |
| Fotografía en selector | Diferencia aceptada | El atlas combina fotografía y monograma; el contrato `Barber` no declara imagen. Se muestra monograma, según el README del atlas y el alcance de #189. |
| Formato visible de `input[type=date]` | Dependiente del navegador | El PNG del atlas muestra `2026-09-03`; Chromium local presenta `03/09/2026`. El valor y el control nativo son el mismo (`2026-09-03`); no se reemplaza por una simulación visual que altere su interacción o accesibilidad. |
| AppHeader y dock | Fuera de alcance de #189 | El atlas fue generado con un encabezado mínimo y dock en flujo. La aplicación integra `AppHeader` y `AppNav` fijos del issue #187, con CTA y navegación agrupada en «Más». Cambiarlo afectaría todas las rutas privadas y contradice la decisión documentada en esos componentes; requiere una entrega propia de shell. |

No se declara equivalencia píxel a píxel mientras permanezcan las dos últimas
diferencias de shell y control nativo. La evidencia sí permite iterar y validar
de forma determinista la composición que pertenece a `DailyAgendaPage`.
