---
prompt_id: "PROMPT-TEST-OTP-EMAIL-RESEND-v1"
version: "1.0"
kind: "test"
status: "executed"
target_agents:
  - "codex"
  - "claude"
  - "human"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-008"
related_hu:
  - "HU-011"
issue: 86
issue_url: "https://github.com/bcaceres19/barberia/issues/86"
suggested_issue_title: "test(qa): verificar flujo OTP real (reto HU-007 y recuperación HU-008/HU-011) con código real"
branch: "docs/86-otp-correo-prompt"
pr: 129
pr_url: "https://github.com/bcaceres19/barberia/pull/129"
depends_on:
  - "HU-008 integrada en main mediante PR #63"
  - "HU-011 integrada en main mediante PR #66"
  - "Issue #86 abierto"
rules:
  - "RN-DAT-01"
  - "RN-DAT-02"
  - "RN-REC-04"
decisions:
  - "DEC-026"
  - "DEC-051"
  - "DEC-063"
  - "DEC-064"
  - "DEC-065"
  - "DEC-066"
acceptance_criteria:
  - "CA-008-01"
  - "CA-008-02"
  - "CA-008-05"
  - "CA-008-06"
  - "CA-008-07"
  - "CA-008-08"
  - "CA-011-01"
  - "CA-011-02"
  - "CA-011-04"
  - "CA-011-05"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/00-control/contradicciones.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "apps/api/README.md"
  - "apps/web/README.md"
  - "apps/api/internal/modules/auth/recovery.go"
  - "apps/api/internal/modules/notification"
  - "apps/api/cmd/api/main.go"
  - "apps/web/src/modules/auth"
created_at: "2026-09-01"
updated_at: "2026-09-02"
report: "test/issue-86-otp-correo-resend-report.md"
execution_branch: "test/86-otp-correo-resend"
supersedes: null
superseded_by: null
---

# Habilitar y probar el OTP de recuperación únicamente por correo con Resend

## Instrucción para el agente

Trabaja únicamente sobre la parte de recuperación de acceso de `HU-008`/`HU-011` que el issue `#86` dejó sin verificar: entregar un OTP real por correo mediante Resend y completar manualmente el recorrido de recuperación. No implementes WhatsApp, no modifiques el reto telefónico de `HU-007` y no cambies la regla de producción fijada por `DEC-051`/`DEC-066`.

La excepción de correo único debe existir exclusivamente en `APP_ENVIRONMENT=local` o `test`. En `pilot` y `production`, Meta WhatsApp Cloud API y Resend continúan siendo obligatorios y el código debe seguir intentando ambos canales.

## Objetivo

Dejar una forma segura, explícita y probada de ejecutar en local el recorrido:

`solicitar recuperación → recibir OTP en un correo real → verificar OTP → cambiar contraseña → comprobar revocación de sesiones → iniciar sesión con la contraseña nueva`.

El resultado debe permitir que el operador configure solamente Resend durante esta prueba local, sin SIM, número virtual ni credenciales de Meta, sin almacenar secretos en el repositorio y sin relajar la configuración endurecida de piloto/producción.

## Preflight obligatorio

1. Comprueba que el árbol esté limpio. Actualiza `main` mediante fast-forward y confirma que los PR `#63` y `#66` están integrados.
2. Verifica que el issue real sea `#86`. Este trabajo cubre solo los pasos de recuperación `HU-008/HU-011`; no cubre el reto por WhatsApp de `HU-007`, por lo que el PR debe usar `Refs #86`, nunca `Closes #86`.
3. Si existe `graphify-out/graph.json`, consulta `HU-008`, `RecoveryCodeSender`, `DualChannelRecoverySender`, `ResendEmailSender`, `buildRouter` y `APP_RESEND_*` antes de editar.
4. Lee completamente `AGENTS.md`, `CLAUDE.md`, `CONTRIBUTING.md` y todos los `source_docs`. Relee en particular `DEC-051`, `DEC-064`, `DEC-065`, `DEC-066`, `HU-008` y `HU-011`.
5. Confirma el comportamiento actual:
   - `apps/api/cmd/api/main.go` solo construye el adaptador real cuando Meta y Resend están completos;
   - con únicamente `APP_RESEND_API_KEY`/`APP_RESEND_FROM_ADDRESS`, local cae hoy en `LoggingRecoveryCodeSender` y no envía correo;
   - `config.Load` exige ambos proveedores fuera de local/test;
   - `auth_recovery_request` requiere una cuenta activa con teléfono verificado, condición que este trabajo no debe eliminar.
6. Crea `test/86-otp-correo-resend` desde `main` actualizada y cambia este prompt a `in_progress`, registrando la rama real.
7. Antes de la prueba manual, el operador debe disponer de una cuenta Resend, un remitente/dominio verificado, una API key cargada de forma segura en el entorno y acceso al buzón destinatario. Nunca solicites que la API key se pegue en un chat, issue, commit, captura o log.

## Alcance incluido

- Permitir que el API seleccione un remitente de recuperación **solo por correo** cuando:
  - `APP_ENVIRONMENT` sea `local` o `test`;
  - `APP_RESEND_API_KEY` y `APP_RESEND_FROM_ADDRESS` estén completos;
  - las credenciales `APP_META_WHATSAPP_*` no estén completas.
- Conservar el remitente dual existente cuando Meta y Resend estén completos.
- Conservar `LoggingRecoveryCodeSender` cuando local/test no tenga Resend completo.
- Conservar el fallo de arranque existente de `config.Load` en `pilot`/`production` si falta cualquiera de los dos proveedores.
- Pruebas unitarias de selección y delegación del canal, sin llamadas de red reales.
- Documentación operativa mínima para ejecutar la prueba sin persistir secretos.
- Una prueba manual real del recorrido de recuperación con un correo controlado por el operador y evidencia redactada.

## Fuera de alcance

- Implementar, contratar o configurar Meta WhatsApp Cloud API, Twilio, Telnyx, números virtuales o SIM.
- Cambiar `DEC-051`, `DEC-066`, la historia `HU-008` o el requisito dual de piloto/producción.
- Permitir correo único en `pilot` o `production`, incluso mediante una variable ambigua o un valor por defecto.
- Eliminar la precondición de teléfono verificado de `auth_recovery_request` o crear una migración para alterar esa regla.
- Modificar el reto de abuso `HU-007`, que sigue dependiendo de verificación telefónica/WhatsApp y permanece pendiente dentro del issue `#86`.
- Cambiar OpenAPI, esquemas HTTP, base de datos, expiración, HMAC, límites, reenvío, política de contraseña o mensajes de no enumeración.
- Enviar correos reales desde pruebas automatizadas.
- Guardar o mostrar API keys, OTP, contraseña, token de reinicio, correo completo o teléfono completo en logs, fixtures, capturas, comentarios o reportes versionados.

## Estado existente que debe conservarse

- `auth.RecoveryService` ya genera el código criptográfico, persiste únicamente su HMAC y llama al puerto `RecoveryCodeSender` después de que PostgreSQL acepta la solicitud.
- `notification.ResendEmailSender` ya implementa la llamada REST real a Resend con timeout de 5 segundos y descarta el cuerpo de error para evitar fugas.
- `notification.DualChannelRecoverySender` ya intenta WhatsApp y correo aun cuando uno falle.
- `auth.LoggingRecoveryCodeSender` y `APP_RECOVERY_CAPTURE_FILE` ya permiten pruebas sin proveedor, pero no demuestran entrega real a un buzón.
- `config.Load` ya protege ambientes endurecidos y sus pruebas demuestran que producción rechaza Meta o Resend incompletos.
- La pantalla `/recuperar-acceso` y las tres operaciones HTTP de `HU-008/HU-011` ya están integradas; no reconstruyas el flujo.

## Trabajo requerido

1. Introduce una composición estrecha de recuperación por correo que satisfaga `auth.RecoveryCodeSender`, ignore únicamente el teléfono en la frontera de entrega y delegue `email`/`code` al `EmailSender`. Mantén el módulo `auth` independiente de Resend.
2. Extrae o ajusta la selección de remitente en `apps/api/cmd/api/main.go` para distinguir de forma explícita:
   - Meta completa + Resend completo → `DualChannelRecoverySender`;
   - solo Resend completo y ambiente local/test → remitente por correo;
   - Resend incompleto en local/test → `LoggingRecoveryCodeSender`.
3. No confíes únicamente en que `config.Load` normalmente rechaza una configuración: la función de selección también debe impedir que una construcción directa de `config.Config` active correo único en `pilot`/`production`.
4. Añade pruebas en la capa más baja adecuada para demostrar:
   - el remitente por correo hace exactamente una llamada con destinatario y código correctos;
   - propaga un error seguro del canal;
   - local/test con solo Resend elige correo;
   - ambos proveedores eligen envío dual;
   - local/test sin Resend usa el marcador seguro;
   - pilot/production nunca selecciona correo único;
   - las pruebas existentes que exigen ambos proveedores en producción siguen pasando.
5. Documenta en `apps/api/README.md` la matriz anterior y el procedimiento operativo. Los ejemplos deben usar marcadores, nunca valores reales.
6. Configura el entorno local fuera del repositorio:
   - `APP_ENVIRONMENT=local`;
   - `APP_RESEND_API_KEY` mediante el mecanismo seguro del operador;
   - `APP_RESEND_FROM_ADDRESS` con remitente verificado;
   - sin `APP_META_WHATSAPP_*` para esta prueba.
7. Prepara una cuenta local sintética activa con teléfono marcado como verificado y un correo real controlado por el operador. Hazlo de forma reversible y no comitees el correo ni lo incluyas en evidencia.
8. Ejecuta manualmente `/recuperar-acceso`:
   - solicita el código y confirma respuesta genérica/no enumerable;
   - comprueba la llegada del correo sin capturar el OTP ni el destinatario completo;
   - introduce el OTP en la interfaz y verifica que avanza al paso 3;
   - cambia la contraseña;
   - demuestra que una sesión anterior queda revocada;
   - inicia sesión con la contraseña nueva;
   - confirma que el código no puede reutilizarse.
9. Restaura o elimina solamente los datos sintéticos creados por la prueba mediante el procedimiento seguro del entorno. No borres datos reales ni alteres migraciones aplicadas.

## Pruebas y evidencia

- Unitarias de `notification`: delegación exacta y fallo de Resend mediante servidor HTTP falso; ninguna red externa.
- Unitarias de `cmd/api`: matriz de selección por ambiente y presencia de credenciales ficticias.
- Regresión de `config.Load`: producción continúa rechazando la ausencia de Meta o Resend.
- Suite Go completa con carrera cuando el entorno lo permita.
- Prueba manual en navegador real del recorrido de `HU-008/HU-011` con un correo recibido realmente.
- Evidencia permitida: timestamp, estado entregado, paso alcanzado, códigos HTTP, texto no sensible y capturas con correo/teléfono/OTP/token completamente redactados.
- Evidencia prohibida: API key, cabecera `Authorization`, OTP, contraseña, token de reinicio, destinatario completo o cuerpo bruto de Resend.

Si no existen credenciales Resend o acceso a un buzón real, completa solo las pruebas automatizadas, marca la prueba manual como **bloqueada por prerrequisito operativo** y no declares resuelto el issue `#86`.

## Documentación y trazabilidad

- Actualiza `apps/api/README.md` únicamente con el modo de prueba local y las garantías de producción.
- No cambies documentación normativa, OpenAPI, matriz de datos ni migraciones porque el contrato y la regla de producción no cambian.
- Actualiza los metadatos de este prompt y el índice del catálogo con estado, rama y PR reales.
- Al finalizar, deja un comentario en el issue `#86` diferenciando:
  - recuperación por correo `HU-008/HU-011`: verificada o bloqueada, con evidencia segura;
  - reto telefónico `HU-007`: aún pendiente y fuera de este PR.

## Verificación final

```text
# apps/api
gofmt -l .
go vet ./...
go test -race ./...

# raíz
git diff --check
git status --short
```

Ejecuta los comandos desde el directorio indicado. `gofmt -l .` debe quedar sin salida. No uses la recepción real del correo como sustituto de las pruebas automatizadas ni al revés.

Entrega dos tablas:

`Criterio | Estado | Prueba o evidencia`

`Dato sensible revisado | Resultado | Evidencia segura`

No declares cumplido aquello que no hayas ejecutado realmente.

## Git y PR

- Rama de ejecución: `test/86-otp-correo-resend`.
- Commit/PR: `test(auth): habilita prueba OTP por correo con Resend`.
- Usa `Refs #86`: el reto WhatsApp de `HU-007` permanece fuera de alcance, por lo que este PR no puede cerrar todo el issue.
- No hagas push directo, force push ni merge de `main`.
