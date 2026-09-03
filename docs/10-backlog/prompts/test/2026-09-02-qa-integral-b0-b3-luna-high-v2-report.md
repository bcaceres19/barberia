# Informe de ejecución — QA integral B0–B3 y NAVA

- Prompt: `PROMPT-TEST-QA-PLATAFORMA-B0-B3-LUNA-v2`
- Fecha: 2026-09-02
- Ambiente: local dedicado, PostgreSQL real en Docker, API local, worker local y frontend Vite
- Resultado ejecutivo: **INCONCLUSA**
- Alcance de cambios: no se modificó producto ni se agregaron secretos al repositorio.

## Evidencia ejecutada

1. PostgreSQL dedicado levantado y migrado con Atlas: 16/16 migraciones aplicadas, sin pendientes.
2. API local saludable: `/health` respondió 200.
3. Worker local iniciado correctamente.
4. Flujo manual en navegador integrado:
   - solicitud de recuperación desde `/recuperar-acceso`;
   - recepción del correo real mediante Resend;
   - verificación del OTP proporcionado por el operador;
   - cambio de contraseña;
   - confirmación visual de revocación de sesiones;
   - nuevo inicio de sesión y llegada al panel.
5. Navegación manual autenticada: Agenda, Servicios, Barberos, Horarios, Configuración y Servicios por barbero cargaron datos del tenant QA sin errores visibles.
6. Validación manual de formularios obligatorios de servicio y nuevo turno.
7. Suite unitaria web: **44 archivos, 608 pruebas aprobadas**.
8. Suite Playwright configurada para Chromium escritorio/móvil, Firefox y WebKit: **388 pruebas iniciadas; 71 aprobadas, 193 fallidas y 124 no ejecutadas**.

## Hallazgos confirmados

### QA-01 — Configuración de remitente vacía en el contenedor

- Severidad: bloqueante para el envío real del correo.
- Evidencia: el primer intento registró `envío de código de recuperación simulado (sin credenciales de proveedor configuradas)`; la API key estaba presente, pero `APP_RESEND_FROM_ADDRESS` tenía longitud cero.
- Acción aplicada en ambiente local: se recreó únicamente el contenedor de API con el remitente verificado documentado. La segunda solicitud generó un correo real y el flujo completo pasó.
- Estado: **resuelto en el ambiente local actual**; debe quedar explícito en el procedimiento de arranque para no repetir la configuración incompleta.

### QA-02 — Expectativas E2E NAVA desactualizadas frente a la UI actual

- Severidad: alta para la confiabilidad de la campaña.
- Evidencia: varias pruebas esperan el encabezado `Inicia sesión`, mientras la pantalla actual muestra `Accede a NAVA`; las pruebas del panel esperan `Agenda de hoy`, mientras la UI observada muestra `Agenda`.
- Impacto: falsos negativos en los escenarios visuales, de navegación y autenticación.
- Estado: **abierto**; requiere reconciliar contratos de texto accesible y actualizar las aserciones con una nueva versión trazable del prompt/pruebas.

### QA-03 — Suite E2E no aislada del estado de autenticación y throttle local

- Severidad: alta para la campaña.
- Evidencia: los escenarios que dependen de login quedaron en `/acceso`; el escenario de credenciales inválidas recibió el reto telefónico en vez del mensaje esperado. La ejecución paralela comparte IP/estado de throttle y no parte de un fixture aislado por worker.
- Impacto: fallos en cascada que impiden evaluar Agenda, catálogo, horarios, citas y regresiones de tenant.
- Estado: **abierto**; requiere preparación reproducible de base de datos/estado de throttle y credenciales E2E explícitas para dos tenants.

### QA-04 — Recuperación E2E automatizada depende de captura local no habilitada

- Severidad: media para la campaña; no afecta el envío real validado manualmente.
- Evidencia: `recuperacion.spec.ts` espera `APP_RECOVERY_CAPTURE_FILE`, pero la corrida real usó Resend y no habilitó captura de códigos.
- Impacto: el recorrido automatizado termina esperando un archivo inexistente.
- Estado: **abierto**; separar el escenario automatizado con sender de captura del escenario manual real con Resend.

### QA-05 — Avisos de infraestructura de pruebas

- Severidad: baja/no concluyente.
- Evidencia: jsdom/axe emite `HTMLCanvasElement.prototype.getContext` no implementado; Firefox emite advertencias internas y tuvo algunos timeouts de navegación/cierre.
- Impacto: ruido y posibles flakies; no se observó un defecto funcional confirmado en el flujo manual.
- Estado: **pendiente de saneamiento del runner**.

## Resultado por puerta

- G0–G2: **observadas como operativas** en el ambiente local dedicado.
- G3–G5: **parcialmente observadas** mediante navegación manual y suite unitaria.
- G6–G8: **inconclusas** por fallos de preparación, aserciones desactualizadas y dependencia de captura/fixtures.
- Declaración final: **INCONCLUSA**. No se autoriza declarar la plataforma APROBADA con esta ejecución.

## Evidencia visual y técnica

- Captura del panel autenticado: `../../../../.codex/visualizations/2026/09/02/01a0634b-7a63-72d1-8891-8de628b84e7a/nava-acceso.png` (artefacto local del entorno).
- Evidencia Playwright generada: `../../../apps/web/evidence/`.
- Informe HTML de Playwright: `../../../apps/web/playwright-report/index.html`.

No se incluyen en este informe correos personales, contraseñas, OTP, API keys ni tokens de sesión.
