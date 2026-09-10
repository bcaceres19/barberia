---
prompt_id: "PROMPT-TEST-OTP-WHATSAPP-META-v1"
version: "1.0"
kind: "test"
status: "superseded"
target_agents:
  - "codex"
  - "claude"
  - "human"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-007"
related_hu:
  - "HU-005"
issue: 86
issue_url: "https://github.com/bcaceres19/barberia/issues/86"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on:
  - "Cuenta Meta Business habilitada para WhatsApp Cloud API"
  - "Número emisor y destinatario de prueba controlados por el operador"
  - "Plantilla OTP aprobada y credenciales inyectadas fuera del repositorio"
rules:
  - "RN-DAT-02"
  - "RN-REC-04"
  - "RN-REC-05"
decisions:
  - "DEC-026"
  - "DEC-052"
  - "DEC-061"
  - "DEC-062"
  - "DEC-066"
  - "DEC-081"
acceptance_criteria:
  - "La sexta solicitud escalada exige el reto y no evalúa la contraseña antes de verificarlo."
  - "Un OTP real llega mediante Meta WhatsApp Cloud API al contacto verificado resuelto por el servidor."
  - "Código válido, inválido, vencido, reutilizado y límite de intentos conservan la respuesta segura prevista."
  - "No se persisten ni publican credenciales, OTP, teléfonos completos, cookies o contraseñas."
  - "La evidencia distingue la verificación operativa de la cobertura automatizada ya existente."
source_docs:
  - "AGENTS.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/10-backlog/prompts/hu/hu-007-defensa-abuso.md"
  - "docs/10-backlog/prompts/test/issue-86-otp-correo-resend.md"
  - "docs/10-backlog/prompts/test/issue-86-otp-correo-resend-report.md"
  - "apps/api/README.md"
  - "apps/api/internal/modules/auth/phone_challenge.go"
  - "apps/api/internal/modules/notification/whatsapp_meta.go"
  - "apps/web/e2e/reto-telefonico.spec.ts"
created_at: "2026-09-04"
updated_at: "2026-09-07"
supersedes: null
superseded_by: "PROMPT-FIX-86-OTP-WHATSAPP-META-v2"
---

# Verificar el reto OTP de acceso con Meta WhatsApp Cloud API

> **No ejecutar.** Este guion asumía incorrectamente que el adaptador de
> Meta ya estaba conectado al reto de `HU-007`. Fue sustituido por
> `PROMPT-FIX-86-OTP-WHATSAPP-META-v2`, que primero corrige la composición
> real y después ejecuta la validación operativa.

## Instrucción

Ejecuta únicamente la parte aún abierta del issue [#86](https://github.com/bcaceres19/barberia/issues/86): la verificación manual real del reto adicional de `HU-007` por WhatsApp. El código y sus pruebas automatizadas ya existen; no los reimplementes ni repitas la entrega por correo de `HU-008/HU-011`.

Este prompt permanece `blocked` hasta que el operador confirme los tres prerrequisitos externos del encabezado. Nunca pidas que una credencial se pegue en el chat, issue, prompt, comando visible, captura o archivo versionado.

## Objetivo

Demostrar en un entorno local o de prueba controlado que una IP escalada recibe un OTP real en el contacto verificado de la cuenta, lo completa y solo entonces puede evaluar sus credenciales, conservando límites, no enumeración, aislamiento y privacidad.

## Preflight

1. Lee todos los `source_docs`, confirma que #86 sigue abierto y que la prioridad operativa permite retomar Meta.
2. Verifica que `MetaWhatsAppSender`, `PhoneChallengeService`, repositorio PostgreSQL, OpenAPI y E2E existentes corresponden al código integrado en `main`.
3. Comprueba la configuración fuera del repositorio: versión de Graph API, phone number ID, token, plantilla y código de idioma. No muestres sus valores.
4. Usa una cuenta sintética activa con teléfono verificado y un número receptor controlado por el operador; no uses datos del piloto o producción.
5. Crea una rama `test/86-reto-whatsapp-meta` solo si se va a persistir un informe o una regresión necesaria. Si la ejecución no cambia archivos, entrega evidencia segura en el issue sin inventar PR.

## Alcance incluido

- Provocar el umbral de `HU-007` de forma controlada y confirmar el `challenge-required`.
- Solicitar el reto, observar una entrega real por Meta, verificar el código y completar el login.
- Verificar código inválido, reenvío/cooldown, consumo único y ausencia de evaluación prematura de la contraseña.
- Revisar respuestas HTTP, consola, red y logs sanitizados.
- Persistir un informe que no contenga teléfono completo, OTP, token, cookie, contraseña ni cuerpo sensible del proveedor.

## Fuera de alcance

- Cambiar la política de canales de `DEC-081`; eso pertenece a `PROMPT-FIX-OTP-CANALES-CONFIGURADOS-v2` y requiere issue propio.
- Implementar otro proveedor, SMS, número virtual, fallback no configurado o un destino introducido por el cliente.
- Cambiar umbrales, vigencia, intentos, HMAC, sesión, contraseña, OpenAPI o migraciones para facilitar la prueba.
- Repetir la validación real de Resend ya documentada.

## Ejecución y evidencia

1. Arranca el stack oficial contra PostgreSQL 14 real con las migraciones vigentes y testdata sintético.
2. Inyecta secretos por el mecanismo seguro del operador; confirma que `.env`, history, salida del proceso y capturas no los exponen.
3. Genera cinco solicitudes fallidas permitidas y confirma que la sexta exige reto antes de evaluar contraseña.
4. Solicita el código y comprueba una entrega real en WhatsApp sin copiar el OTP a archivos ni salidas de herramienta.
5. Verifica el código desde la UI real, inicia sesión y confirma la sesión privada.
6. Ejecuta ramas inválida, reutilizada, vencida y límite de intentos con cuentas/datos sintéticos separados cuando sea necesario.
7. Revisa que cuenta inexistente, inactiva o sin contacto utilizable no revele su condición ni un destino completo.
8. Destruye solo los recursos efímeros creados para esta prueba y revoca cualquier token temporal según el procedimiento del operador.

Ejecuta además las pruebas automatizadas afectadas y `git diff --check` si existe un diff. Entrega:

`Criterio | Estado | Evidencia segura`

`Dato sensible | Comprobación | Resultado`

Usa `Refs #86` hasta completar todo lo que el issue mantiene abierto. No declares la entrega real solo a partir de un mock o servidor falso.
