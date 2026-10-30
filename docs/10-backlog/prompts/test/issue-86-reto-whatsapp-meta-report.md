---
prompt_id: "PROMPT-FIX-86-OTP-WHATSAPP-META-v2"
report_for: "docs/10-backlog/prompts/fix/issue-86-integrar-otp-whatsapp-meta-v2.md"
issue: 86
issue_url: "https://github.com/bcaceres19/barberia/issues/86"
executed_at: "2026-10-30"
executed_by: "agente Claude (sesión interactiva), operador bcaceres19"
environment: "local: API real (APP_ENVIRONMENT=local) sobre PostgreSQL 14 con migraciones al día, Meta WhatsApp Cloud API en modo `text` (DEC-124), un único destinatario de prueba controlado por el operador"
---

# Evidencia: reto OTP de acceso (HU-007) con WhatsApp real por Meta (issue #86, paso 13)

## Resumen

Un código real llegó por Meta WhatsApp Cloud API al contacto verificado de una cuenta sintética de
desarrollo y permitió completar el login tras el reto. Las ramas de código reutilizado, inválido y
cuenta inexistente responden idénticas. La recuperación por correo (pasos 15–17) ya quedó
verificada con Resend en [issue-86-otp-correo-resend-report.md](issue-86-otp-correo-resend-report.md).

## Resultado

| Criterio                                                                                            | Estado | Evidencia segura                                                                                  |
| --------------------------------------------------------------------------------------------------- | ------ | ------------------------------------------------------------------------------------------------- |
| Cinco intentos fallidos permitidos; la sexta solicitud exige el reto antes de evaluar la contraseña | PASS   | Cinco `401` y a la sexta `429` con el reto visible en la UI                                       |
| Se solicita el código (`challenge`)                                                                 | PASS   | `202`; el API registró `whatsapp aceptado por Meta` (solo `wamid` y modo, sin teléfono ni código) |
| Un OTP real llega por WhatsApp al contacto verificado que resuelve el servidor                      | PASS   | El operador recibió el código de 6 dígitos en su teléfono                                         |
| Código válido (`challenge/verify`)                                                                  | PASS   | `204`; el escalamiento de la IP se limpia                                                         |
| Login tras el reto con la contraseña correcta y sesión privada                                      | PASS   | `login` `200` y `GET /private/auth/session` `200`                                                 |
| Código reutilizado                                                                                  | PASS   | `401` uniforme; la fila del reto queda consumida (1 de 1)                                         |
| Código inválido                                                                                     | PASS   | `401` uniforme, igual que el reutilizado                                                          |
| Cuenta inexistente con un código                                                                    | PASS   | `401` uniforme, no revela si la cuenta existe                                                     |
| Código vencido y límite de intentos                                                                 | N/A    | No se repitieron con proveedor real; los cubren las pruebas Go/SQL automatizadas existentes       |

## Hallazgo operativo (no es un defecto de producto)

La primera solicitud devolvió `202` y Meta la aceptó, pero **no se entregó**: en modo `text` Meta solo
entrega si el destinatario escribió al número emisor en las últimas 24 horas (DEC-124). Con esa
ventana abierta la entrega funcionó. La API de Meta acepta la petición de forma síncrona y falla
después, así que el registro `whatsapp aceptado por Meta` no prueba la entrega: para saberlo hay
que mirar el webhook de estados (DEC-126) o al destinatario. Para no depender de la ventana hace
falta una plantilla `AUTHENTICATION` aprobada en el WABA (hoy solo existen `hello_world` y una de
prueba de categoría `UTILITY`).

## Datos sensibles revisados

| Dato sensible                 | Comprobación                                                      | Resultado    |
| ----------------------------- | ----------------------------------------------------------------- | ------------ |
| Credenciales de Meta          | Se leyeron del entorno local sin imprimirlas                      | No expuestas |
| Teléfono completo             | Ausente de este informe, de los logs del API y de la evidencia    | No expuesto  |
| Código OTP                    | No se persiste en archivos ni se publica; un solo uso             | No publicado |
| Estado de la cuenta de prueba | Teléfono y verificación revertidos a su estado original al cerrar | Restaurado   |

## Conclusión

Paso 13 de #86 verificado de extremo a extremo con un proveedor real. Con el informe de Resend
(pasos 15–17), la próxima acción de #86 queda cubierta.
