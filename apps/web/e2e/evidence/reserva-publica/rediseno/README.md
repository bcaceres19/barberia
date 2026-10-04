# Reserva pública · rediseño sobre tinta (`DEC-111`)

Modo de conformidad: **identidad guiada** (sin mockup exacto asignado, `DEC-078`/`DEC-080`). No hay comparación de fidelidad contra una lámina; la evidencia es el recorrido real renderizado, con datos de prueba interceptados (`page.route`, ningún dato real).

Generada por `apps/web/e2e/reserva-publica-rediseno-evidencia.spec.ts` (Chromium de escritorio) en 320, 360, 768, 1280, 1440 y 1280 con zoom de texto 200 % (aproximado a 640 px).

| Captura                                                       | Estado                                                                                                      |
| ------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| `01-entrada`                                                  | Portada: nombre de la barbería, hora local, contacto y «Reservar un turno».                                 |
| `02-servicios`, `03-servicio-elegido`                         | Lista de servicios; elegido, con la barra de acción fija (capturada con la ventana al final del contenido). |
| `04-barberos`, `05-barbero-elegido`                           | Lista de barberos; elegido.                                                                                 |
| `06-horario`, `07-horario-elegido`                            | Tira de días, franjas; franja elegida con su ficha y la barra de acción.                                    |
| `08-datos`, `09-datos-errores`, `10-resumen`, `11-confirmado` | Panel de tinta: formulario, errores que conservan lo escrito, resumen y confirmación.                       |
| `12-carga`, `13-vacio`, `14-no-encontrado`                    | Estados de página sobre tinta.                                                                              |
| `15-foco-teclado`                                             | Foco visible y selección con flechas (solo 1280).                                                           |

## Verificación

- Sin desplazamiento horizontal en ningún ancho ni paso.
- axe-core **con `color-contrast` activo**, con movimiento reducido, en cada pantalla y estado (recorrido completo, conflicto de horario con alternativas, barbero único, no encontrado, sin conexión y error inesperado).
- Con `prefers-reduced-motion` no queda ninguna animación corriendo en el contenido.
- Teclado: flechas, espacio y foco visible; volver atrás retrocede el progreso.
- Firefox y Chromium móvil (Pixel 7): recorrido a 360 px y la verificación accesible, sin fallos. WebKit no está instalado en el entorno local y queda para CI.
