# Series por fechas, excepciones y edición por alcance · #100

Modo de conformidad: **identidad guiada** NAVA / Tailored Grid dentro de la ventana de bloqueos de Barberos (DEC-105). No se asignó un mockup exacto; los diálogos nuevos reutilizan la gramática de los existentes (tinta, filete de latón, pie fijo, errores en línea).

## Qué cubre

- **Agregar bloqueo por fechas** (`date_list`): tipo, hora, duración, vigencia (con o sin fin), motivo y lista de fechas elegidas; la serie se crea con todas sus fechas en una sola operación idempotente. Una fecha repetida o fuera de la vigencia se explica antes de enviar.
- **Fechas** (solo series por fechas) y **Excepciones** (cualquier serie): agregar y retirar fechas explícitas; suprimir una instancia con motivo opcional («esta instancia no») y restaurarla.
- **Editar** una serie: «Toda la serie» (`whole`) o «Esta fecha y las siguientes» (`this_and_following`, solo `weekly`): la serie original queda recortada el día anterior al corte y se crea una nueva. Una serie por fechas solo admite `whole` (el contrato lo rechaza para `date_list`).

Sin cambios de contrato, migración ni regla de negocio: se usan las operaciones de HU-042 ya integradas.

## Verificación

- Vitest: cliente tipado (`timeBlocksApi.test.ts`), diálogos (`SeriesDialogs.test.ts`) e integración en el panel (`BarberBlocksPanel.test.ts`).
- E2E contra API y PostgreSQL reales: `pnpm exec playwright test e2e/barberos-bloqueos-series.spec.ts --project=chromium-desktop --project=chromium-mobile --project=firefox` (los tres pasaron). Recorre alta por fechas, fechas y excepciones (204/409 reales), edición `whole` y `this_and_following` (comprueba en el API que la original termina el día anterior), persistencia tras recargar, foco de retorno, Escape y errores recuperables.
- Axe WCAG A/AA (incluido contraste) sobre cada diálogo y el panel, terminadas las transiciones; sin desbordamiento horizontal a 320×740, 360×800, 768×1024 y 1280×900; botones de registro de al menos 44 px de alto y dentro del viewport.
- Hallazgo corregido durante la verificación: el hover del botón secundario sobre la tinta del diálogo dejaba texto claro sobre fondo claro (contraste 1,13); ahora `.block-dialog` fija el fondo oscuro del hover.

## Capturas

Solo se versiona `chromium-desktop-*`: `alta-fechas`, `fechas-excepciones`, `editar-serie`, `editar-desde` y, por ancho (320/360/768/1280), `panel`, `registro`, `fechas` y `alta`. Datos sintéticos.

## Reproducción local

Aplicar `database/testdata/ui_bloqueos_barberos_286.sql` con rol migrador únicamente en la BD local de pruebas y arrancar API/Vite; la suite reutiliza esas cuentas ficticias y su sesión en `apps/web/.auth/` (ignorado por Git). Cada ejecución crea series con un rótulo único y las retira al terminar.
