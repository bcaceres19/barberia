# Plantilla de correos transaccionales · evidencia

Issue [#465](https://github.com/bcaceres19/barberia/issues/465), `DEC-143`.

Capturas de las vistas previas que genera `EMAIL_PREVIEW_DIR=<carpeta> go test ./internal/modules/notification -run WritePreviews`, con datos ficticios (cliente «Ana», «Barbería Centro»). Renderizadas con Chromium (Playwright) en contextos propios.

| Archivo | Qué muestra |
| --- | --- |
| `confirmado-320-light.png`, `confirmado-360-light.png`, `confirmado-768-light.png`, `confirmado-1280-light.png` | Confirmación de turno en los cuatro anchos de la guía visual, sin desbordamiento horizontal |
| `confirmado-360-dark.png`, `cancelado-360-dark.png` | Modo oscuro (`prefers-color-scheme: dark`) |
| `cancelado-360-light.png`, `reprogramado-360-light.png` | Avisos de `HU-130` |
| `codigo-360-light.png` | Correo de código (recuperación y acceso) |

Accesibilidad: axe-core 0 violaciones en las cuatro plantillas, en claro y oscuro, a 360 px. Contraste AA de cada par de color comprobado por `TestEmailTemplate_ColorPairsMeetAA`.
