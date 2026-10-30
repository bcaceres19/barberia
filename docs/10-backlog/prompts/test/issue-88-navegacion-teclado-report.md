---
report_for: "docs/10-backlog/prompts/test/2026-08-25-qa-manual-b0-b1-chrome-mcp-report.md"
issue: 88
issue_url: "https://github.com/bcaceres19/barberia/issues/88"
executed_at: "2026-10-30"
executed_by: "agente Claude (sesión interactiva), operador bcaceres19"
environment: "efímero: postgres:14 en :55432 con fixtures de CI, API :8081, Vite :5174; Chromium de Playwright a 1280×800, teclas entregadas al documento con `page.keyboard`"
---

# Evidencia: navegación por teclado en B0 y B1 (issue #88)

## Resumen

El límite de #88 era la herramienta (el navegador integrado no entregaba `Tab` ni `Enter`). Con
Chromium de Playwright las teclas sí llegan al documento y el recorrido completo funciona sin
ratón. No se encontró ningún defecto de producto.

## B0 · acceso (`/acceso`)

| Paso                                   | Resultado                                                                                   |
| -------------------------------------- | ------------------------------------------------------------------------------------------- |
| `Tab` desde el inicio de la página     | Correo → Contraseña → «Mostrar contraseña» → «Iniciar sesión» → «¿Olvidaste tu contraseña?» |
| Tras el último elemento                | El foco sale del documento (barra del navegador) y vuelve al inicio con el siguiente `Tab`  |
| `Shift+Tab` x3 desde el último         | Orden inverso exacto: enlace → «Iniciar sesión» → «Mostrar contraseña»                      |
| Escribir correo, `Tab`, escribir clave | El foco pasa a Contraseña; los valores se escriben sin tocar el ratón                       |
| `Enter` con el foco en Contraseña      | Envía el formulario y navega a `/panel`                                                     |
| Indicador de foco                      | Visible (contorno o sombra) en todos los controles recorridos                               |

## B1 · panel privado (`/panel`)

| Paso                      | Resultado                                                                                                                                     |
| ------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| `Tab` x14                 | Cerrar sesión → marca → Agenda → Servicios → Barberos → Horarios → Servicios por barbero → Reserva pública → Configuración, y vuelta al ciclo |
| `Shift+Tab` x3            | Recorre en orden inverso sin saltos ni trampas de foco                                                                                        |
| `Enter` sobre «Servicios» | Navega de `/panel` a `/panel/servicios`                                                                                                       |
| Indicador de foco         | Visible en todos los enlaces y botones                                                                                                        |

## Alcance y límites

- El reto del código de 6 dígitos con teclado (autoavance, retroceso, pegado) está cubierto por
  las pruebas de componente de `OtpInput`/`PhoneChallengeForm`; en esta sesión el acceso no pidió
  reto en la ejecución del recorrido por teclado, así que no se re-verificó a mano.
- Se probó a 1280 px; los anchos 320/360/768 los cubren las pruebas E2E de evidencia responsiva.

## Conclusión

Sin hallazgos. Cubre la próxima acción de #88 (Tab, Shift+Tab y Enter en B0/B1).
