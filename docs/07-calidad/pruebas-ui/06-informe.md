# Informe y evidencia

Salida exclusiva: apps/web/test-results/ui-qa/<campaña>/<scope>/report.md y results.json. Coordinador consolida en esa campaña; copia pública sanitizada puede ir en docs/10-backlog/evidence/ui-qa-<issue>/ desde rama de campaña.

## Campos

Campaña, issue, prompt_id/versión, scope, SHA ejecutado, modelo/esfuerzo reales, runtime, fecha, navegador, viewports efectivos y perfil. Cuenta identificada por scope; jamás credencial.

Tabla **caso | PASS/FAIL/BLOCKED/NOT_RUN/N/A | observado | evidencia relativa | REAL/SIMULATED**. results.json conserva lista cases con id/status/observed/evidence/mode y los mismos metadatos, sin cuerpos completos.

Hallazgo: severidad, ruta redaccionada, precondiciones, pasos, esperado con RN/DEC/HU, observado, request_id/status, reproducción e impacto. P0 corrupción/fuga/pérdida; P1 flujo o guardado roto; P2 fallo recuperable/responsive/accesible; P3 detalle. Infraestructura no lista = BLOCKED, no bug del producto.

Registrar conteos, capacidades no verificadas y siguientes casos exactos.

## Privacidad

Trazas/HAR, perfiles, login/OTP y enlaces Mi turno pueden incluir secretos incluso de prueba. Guardar crudos en salida ignorada y privada (directorio 0700, archivos sensibles 0600), nunca adjuntar/versionar sin revisar. Ocultar inputs de contraseña/OTP y segmentos con token en screenshots. No imprimir cookies, cuerpos de autenticación ni URL de Mi turno.

Evitar aserciones que impriman token/contraseña en su error; registrar solo resultado booleano redaccionado. request_id y status son evidencia utilizable.

PASS exige aserción/evidencia; cargar pantalla no acredita guardado, RLS ni expiración. executed en índice significa ejecución informada, no éxito de toda la app. No convertir pendientes/bloqueos en PASS.
