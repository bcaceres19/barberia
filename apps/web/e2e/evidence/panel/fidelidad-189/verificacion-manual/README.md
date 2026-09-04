# Verificación manual — issue #189

Capturas de la app real (`barberia-qa-local`, no mocks) tomadas con el harness Playwright del repo durante la implementación de la fidelidad visual de `/panel`. No sustituyen el gate C completo del atlas (24 eventos × 2 viewports con baseline/final/lado-a-lado/overlay) — ver `apps/web/e2e/evidence/panel-fidelidad-189-desviaciones.md`.

- `01-agenda-lista-1280.png` — evento 01, escritorio: tres turnos con los tres tratamientos de material (confirmado en pergamino, completado y cancelado en contorno sobre tinta), insignias en contorno, eje con carril delimitado y marca "Ahora".
- `01-agenda-lista-375.png` — mismo estado en 375px móvil: mismos componentes, sin línea temporal (oculta bajo 1024px por diseño).
- `09-dia-sin-turnos-1280.png` — evento 09: eje vacío conservado con "Ahora", composición local de divisor/titular/acción sin usar `PageState` de página completa.
- `12-seleccion-barbero-1280.png` — evento 12: listbox abierto con avatar en el disparador y en cada opción (monograma y variante de una sola inicial conviviendo).
- `12-seleccion-barbero-monograma-detalle.png` — recorte de la lista abierta que confirma el monograma de dos letras ("E1") para un nombre de varias palabras, legible tras el ajuste de `letter-spacing`/tamaño.
