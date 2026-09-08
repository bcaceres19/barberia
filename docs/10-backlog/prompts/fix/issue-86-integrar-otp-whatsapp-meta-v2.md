---
prompt_id: "PROMPT-FIX-86-OTP-WHATSAPP-META-v2"
version: "2.0"
kind: "fix"
status: "executed"
target_agents:
  - "codex"
target_model: "gpt-5.6-terra"
reasoning_effort: "high"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-007"
related_hu:
  - "HU-005"
  - "HU-008"
  - "HU-011"
issue: 86
issue_url: "https://github.com/bcaceres19/barberia/issues/86"
suggested_issue_title: null
branch: "fix/86-reto-otp-whatsapp-meta"
pr: 223
pr_url: "https://github.com/bcaceres19/barberia/pull/223"
depends_on:
  - "HU-007 integrada mediante PR #61"
  - "HU-008 integrada mediante PR #63"
  - "Issue #86 abierto"
  - "Cuenta Meta con WhatsApp Cloud API y recursos de prueba configurados fuera del repositorio"
  - "Destinatario de prueba controlado por el propietario"
  - "Plantilla Authentication con botón Copy Code aprobada o verificable durante el preflight"
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
  - "CA-007-01"
  - "CA-007-02"
  - "CA-007-03"
  - "CA-007-04"
  - "CA-007-05"
  - "CA-007-06"
  - "CA-007-07"
  - "El reto HU-007 usa MetaWhatsAppSender cuando la configuración Meta está completa y conserva el marcador seguro solo en local/test sin Meta."
  - "Un OTP generado por NAVA llega al destinatario de prueba y se verifica una sola vez sin exponer secretos ni PII."
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/04-arquitectura/backend-go.md"
  - "docs/04-arquitectura/stack-despliegue-operacion.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/06-api/estandar-openapi.md"
  - "docs/10-backlog/prompts/hu/hu-007-defensa-abuso.md"
  - "docs/10-backlog/prompts/test/issue-86-otp-correo-resend.md"
  - "docs/10-backlog/prompts/test/issue-86-otp-correo-resend-report.md"
  - "docs/10-backlog/prompts/fix/issue-pending-otp-canales-configurados-v2.md"
  - "apps/api/README.md"
  - "apps/api/internal/platform/config/config.go"
  - "apps/api/internal/modules/auth/phone_challenge.go"
  - "apps/api/internal/modules/auth/phone_sender.go"
  - "apps/api/internal/modules/auth/postgres/phone_challenge_repository.go"
  - "apps/api/internal/modules/auth/httpapi/challenge_handler.go"
  - "apps/api/internal/modules/notification/ports.go"
  - "apps/api/internal/modules/notification/whatsapp_meta.go"
  - "apps/api/internal/modules/notification/whatsapp_meta_test.go"
  - "apps/api/cmd/api/main.go"
  - "api/openapi/paths/public-auth.yaml"
  - "database/migrations/20260817180000_create_login_throttle_and_phone_challenge.sql"
  - "apps/web/src/modules/auth/components/PhoneChallengeForm.vue"
  - "apps/web/e2e/reto-telefonico.spec.ts"
created_at: "2026-09-07"
updated_at: "2026-09-07"
supersedes: "PROMPT-TEST-OTP-WHATSAPP-META-v1"
superseded_by: null
---

# Integrar y verificar el OTP real de HU-007 con Meta WhatsApp Cloud API

## Perfil de ejecución

Ejecuta este prompt con **GPT-5.6 Terra** y razonamiento **high**. Trabaja con autonomía hasta completar código, pruebas y documentación. La única pausa permitida es una acción humana inevitable de Meta o del teléfono receptor: aprobación de plantilla, renovación del token temporal o introducción local del OTP recibido. Nunca solicites que el token, el OTP, el teléfono completo, una contraseña, una cookie o un DSN se peguen en el chat.

## Objetivo

Cerrar la brecha real del issue #86 para `HU-007`: hoy NAVA genera, firma, persiste y verifica correctamente el reto OTP, y ya existe `notification.MetaWhatsAppSender`, pero `apps/api/cmd/api/main.go` continúa inyectando `auth.LoggingPhoneCodeSender` en `PhoneChallengeService`. Implementa la composición mínima que activa Meta WhatsApp Cloud API cuando su configuración está completa, conserva un fallback seguro solo para desarrollo sin credenciales y demuestra el recorrido real:

`sexta solicitud de login → challenge-required → solicitar reto → NAVA genera OTP → Meta lo entrega al contacto verificado → verificar OTP → desbloquear IP → impedir reutilización`.

Este trabajo prueba WhatsApp como canal real de desarrollo. No sustituye ni implementa la configuración completa por barbería/evento de `DEC-081`; esa capacidad sigue en `PROMPT-FIX-OTP-CANALES-CONFIGURADOS-v2` y necesita issue propio. No presentes este PR como cumplimiento total de `DEC-081`.

## Hechos del estado actual que debes confirmar

1. `auth.PhoneChallengeService` ya usa `crypto/rand`, genera seis dígitos, calcula `HMAC-SHA256` con `APP_AUTH_HMAC_SECRET` y entrega el código en claro solo al puerto `PhoneCodeSender` después de que PostgreSQL acepte la solicitud.
2. `auth_phone_challenge_request` resuelve en servidor el teléfono de `staff_user`; el request público recibe `email`, nunca un teléfono de entrega.
3. La migración vigente ya implementa un código por IP/cuenta, expiración de 300 segundos, máximo de cinco intentos, consumo único, cooldown de 60 segundos, máximo de tres retos activos por IP en 15 minutos, RLS y funciones `SECURITY DEFINER` estrechas.
4. `notification.MetaWhatsAppSender` ya envía una plantilla Authentication a `POST https://graph.facebook.com/{version}/{phone-number-id}/messages`, elimina el `+` del E.164, incluye el código en `body` y en el botón `sub_type=otp`, aplica timeout de cinco segundos y descarta cuerpos de error que puedan contener PII.
5. Ese adaptador ya participa en recuperación mediante `DualChannelRecoverySender`, pero no satisface directamente `auth.PhoneCodeSender`, cuyo método es `SendCode(ctx, phone, code)`.
6. `cmd/api/main.go` todavía construye siempre `auth.NewLoggingPhoneCodeSender(logger)` para `HU-007`; por ello llenar `apps/api/.env` no activa hoy un envío real.
7. Go no carga `.env` automáticamente y no se añadirá `godotenv` ni otra dependencia solo para hacerlo.
8. `apps/api/.env` está ignorado por Git y contiene configuración local del operador. No abras, imprimas, captures, serialices ni incluyas su contenido en resultados. Solo carga sus pares clave/valor en el proceso de forma silenciosa.

Si cualquiera de estos hechos ya cambió en `main`, adapta la solución al estado integrado y documenta la diferencia; no dupliques una composición ya existente.

## Preflight obligatorio

1. Comprueba `git status -sb`. El árbol compartido puede contener cambios ajenos: no los modifiques, no los incluyas en commits y no uses `reset`, `checkout --`, `clean` ni `stash` sobre trabajo del usuario.
2. Consulta el issue #86 y confirma que sigue abierto. Confirma también que PR #61 (`HU-007`) y PR #63 (`HU-008`) están en `main`.
3. Lee completamente todos los `source_docs` y localiza `DEC-061`, `DEC-062`, `DEC-066` y `DEC-081` en su fuente normativa.
4. Confirma con `git check-ignore -v apps/api/.env` que el archivo local no puede versionarse. No ejecutes `Get-Content` de ese archivo como comando visible ni muestres valores.
5. Carga el archivo en la sesión PowerShell sin producir salida:

   ```powershell
   Get-Content .env | Where-Object { $_ -match '^[^#][^=]*=' } | ForEach-Object {
     $name, $value = $_ -split '=', 2
     Set-Item -Path "Env:$name" -Value $value
   }
   ```

   Ejecútalo desde `apps/api`. Después valida únicamente presencia, nunca contenido, de `APP_META_WHATSAPP_API_VERSION`, `APP_META_WHATSAPP_PHONE_NUMBER_ID`, `APP_META_WHATSAPP_ACCESS_TOKEN`, `APP_META_WHATSAPP_TEMPLATE_NAME` y `APP_META_WHATSAPP_LANGUAGE_CODE`.
6. No agregues `META_WHATSAPP_WABA_ID` ni `META_WHATSAPP_TEST_RECIPIENT` a `config.Config`: son auxiliares de administración/Postman y el runtime de NAVA no debe decidir el destino desde ellas.
7. Confirma fuera de logs que la plantilla configurada es categoría `AUTHENTICATION`, está `APPROVED`, tiene botón `COPY_CODE` y que el idioma coincide exactamente. Si Meta exige renovar el token temporal o aprobar la plantilla, marca solo la prueba manual como bloqueada hasta que el operador lo resuelva; continúa con implementación y pruebas automatizadas.
8. Como el árbol principal está sucio, trabaja en un worktree aislado creado desde `origin/main` y usa la rama `fix/86-reto-otp-whatsapp-meta`. No copies el `.env` al worktree ni lo comitees; para la prueba real carga las variables desde el archivo local ignorado por una ruta conocida sin imprimirlas.
9. Cambia este prompt y el catálogo a `in_progress`, registrando la rama real, dentro de la rama de ejecución.

## Diseño requerido

### 1. Adaptación del puerto

Reutiliza `notification.MetaWhatsAppSender`; no crees un segundo cliente Meta. Añade la adaptación mínima y explícita entre `notification.WhatsAppSender.Send` y `auth.PhoneCodeSender.SendCode`.

Solución preferida: un tipo pequeño dentro de `internal/modules/notification`, por ejemplo `PhoneChallengeSender`, que reciba un `WhatsAppSender` y exponga:

```go
SendCode(ctx context.Context, phone, code string) error
```

El tipo no necesita importar `auth`: Go permite que satisfaga estructuralmente el puerto. No acoples `auth` a Meta ni muevas reglas OTP al módulo `notification`.

Agregar `SendCode` directamente a `MetaWhatsAppSender` es aceptable solo si las pruebas y la documentación muestran que no confunde el puerto genérico `WhatsAppSender` con el caso de uso. No dupliques el armado del payload.

### 2. Selección en la raíz de composición

Extrae o implementa en `apps/api/cmd/api/main.go` una selección probada, análoga a `selectRecoverySender`, con esta matriz:

| Meta completa | Ambiente | Remitente HU-007 |
| --- | --- | --- |
| sí | `local`/`test` | adaptador real sobre `MetaWhatsAppSender` |
| sí | `pilot`/`production` | adaptador real sobre `MetaWhatsAppSender` |
| no | `local`/`test` | `LoggingPhoneCodeSender` |
| no | `pilot`/`production` | configuración inválida; `config.Load` debe impedir el arranque |

“Meta completa” significa phone number ID, access token y template name no vacíos. API version y language code conservan sus defaults/configuración validada.

Una configuración Meta parcial debe fallar al arrancar incluso en local/test: no debe caer silenciosamente al remitente simulado y hacer creer al operador que probó Meta. Ausencia total en local/test sí conserva el marcador actual y las E2E sin terceros.

No exijas Resend para el reto `HU-007`. La selección dual de recuperación conserva su conducta vigente y no se reescribe en este PR.

### 3. Seguridad y errores

- Nunca registres configuración completa, token, teléfono, OTP, payload bruto, `Authorization` ni cuerpo de error de Meta.
- Conserva la respuesta pública `202` no enumerable de `/auth/challenge`, incluso si Meta responde error o timeout.
- Conserva el `401` uniforme de verificación para código incorrecto, vencido, agotado, reutilizado o cuenta inexistente.
- No añadas reintento síncrono, cola, broker, Redis, circuit breaker o webhook en este PR.
- No cambies vigencia, límites, HMAC, tabla, funciones SQL, endpoint, DTO ni copy solo para facilitar la prueba.
- No aceptes `phone` ni destinatario en ningún request.
- No introduzcas una variable para el teléfono receptor: el destino sale exclusivamente de `staff_user.phone` verificado.

## Archivos previstos

Modifica únicamente lo necesario, previsiblemente:

- `apps/api/cmd/api/main.go`: selección e inyección del remitente real.
- `apps/api/cmd/api/phone_sender_selection_test.go` (nuevo): matriz de selección.
- `apps/api/internal/modules/notification/phone_challenge_sender.go` (nuevo, si eliges adaptador): puente pequeño entre puertos.
- `apps/api/internal/modules/notification/phone_challenge_sender_test.go` (nuevo): delegación exacta y propagación de error seguro.
- `apps/api/internal/platform/config/config.go`: rechazo de configuración Meta parcial, si no existe todavía.
- `apps/api/internal/platform/config/config_test.go`: matriz ausente/completa/parcial y ambientes.
- `apps/api/README.md`: activación local real, carga silenciosa del `.env`, matriz de selección y procedimiento de prueba sin valores.
- Este prompt y `docs/10-backlog/prompts/README.md`: estado, rama, PR y resultado reales.

No se espera modificar:

- migraciones o `atlas.sum`;
- OpenAPI o cliente TypeScript;
- frontend Vue;
- `PhoneChallengeService` o su repositorio PostgreSQL;
- `whatsapp_meta.go`, salvo que una prueba demuestre un defecto real del payload frente a la plantilla aprobada.

Si necesitas tocar uno de esos elementos, explica primero por qué el contrato o comportamiento actual no basta y mantén el cambio dentro del issue #86. Una discrepancia material con `DEC-081` no se resuelve aquí: se deriva al prompt funcional pendiente.

## Pruebas automatizadas obligatorias

1. Adaptador:
   - delega exactamente teléfono y código una vez;
   - propaga error del canal sin añadir teléfono/código al mensaje;
   - satisface `auth.PhoneCodeSender` mediante una aserción de compilación en la capa apropiada.
2. Selección:
   - Meta completa selecciona remitente real en local y producción;
   - Meta totalmente ausente selecciona marcador solo en local/test;
   - configuración parcial falla en `config.Load` y nunca selecciona silenciosamente marcador;
   - el remitente de recuperación no cambia.
3. Cliente Meta existente:
   - conserva endpoint `{version}/{phone-number-id}/messages`;
   - `to` sin `+`;
   - plantilla e idioma exactos;
   - código en body y botón OTP;
   - timeout/error no filtra cuerpo del proveedor.
4. Auth:
   - solicitud no aceptada no llama al remitente;
   - solicitud aceptada llama una vez;
   - error del remitente no cambia el `202` público;
   - verificación conserva consumo único, expiración y máximo de intentos.
5. Regresión:
   - suites completas de `auth`, `notification`, `config` y `cmd/api`;
   - E2E existente con captura local y sin Meta sigue funcionando.

No hagas llamadas reales a Meta desde pruebas automatizadas.

## Prueba manual real

La prueba real se ejecuta después de que las automatizadas pasen.

1. Carga el `.env` local silenciosamente. Verifica solo que las claves requeridas estén presentes.
2. Arranca PostgreSQL 14+ con las migraciones vigentes y una cuenta sintética activa con credencial válida.
3. El operador debe asociar localmente a esa cuenta su número receptor controlado en formato E.164 y fijar `phone_verified_at` en una segunda sentencia, porque el trigger anula la verificación cuando cambia `phone`. No edites ni comitees `database/testdata/hu007_reto_telefonico.sql` con un número real.
4. Nunca muestres el número completo. Evidencia permitida: máscara, últimos dos dígitos, timestamp y estado.
5. Levanta `go run ./cmd/api` con Meta activa. Confirma que no aparece el mensaje de remitente simulado.
6. Desde la misma IP, realiza cinco solicitudes de login que todavía evalúan normalmente y una sexta que responde `429 challenge-required`; demuestra con doble/instrumentación automatizada que la contraseña no se evaluó en la sexta.
7. Ejecuta `POST /api/v1/public/auth/challenge` con el email de la cuenta. Debe responder `202` uniforme.
8. Confirma físicamente que el teléfono recibió un mensaje desde el número de prueba de Meta, con plantilla Authentication y botón Copy Code. No captures ni transcribas el OTP en archivos, logs, issue, PR o chat.
9. Introduce el OTP directamente en Postman o en la UI local; no lo envíes al agente por chat. `POST /api/v1/public/auth/challenge/verify` debe responder `204`.
10. Repite el mismo OTP y confirma el `401` uniforme. Prueba además un código ficticio distinto; no publiques el real.
11. Reintenta login con la credencial válida y confirma que la IP dejó de estar escalada.
12. Revisa logs y evidencia para demostrar ausencia de token, OTP, teléfono completo, email completo, contraseña, cookie, DSN y cuerpo bruto de Meta.
13. Revoca o deja vencer el token temporal cuando termine la sesión. Restaura únicamente los datos sintéticos o locales modificados para la prueba.

Si el operador todavía no tiene una plantilla `APPROVED`, el token venció o Meta rechaza el destinatario de prueba, registra el código de error sanitizado y marca la prueba manual como bloqueada. No falsees entrega con `APP_PHONE_CHALLENGE_CAPTURE_FILE` y no declares que Meta funcionó basándote en un mock.

## Errores que debes diagnosticar sin filtrar datos

- `401`/OAuth: token temporal vencido o permiso ausente.
- `400`: phone number ID, nombre/idioma de plantilla o componentes incompatibles.
- plantilla `PENDING`, `REJECTED`, `PAUSED` o `DISABLED`.
- destinatario no agregado/verificado en el entorno de prueba.
- Graph API version no soportada.
- Meta acepta con `wamid` pero no entrega: deja el seguimiento por webhook fuera de este PR y registra solo que falta evidencia de entrega.
- respuesta `202` sin llamada a Meta: solicitud no aceptada por las tres condiciones de HU-007, cooldown/límite, o selección equivocada de remitente.

Nunca incluyas el cuerpo bruto de error si puede contener teléfono o destinatario.

## Verificación final

```text
# apps/api
gofmt -l .
go vet ./...
go test -race ./internal/platform/config/...
go test -race ./internal/modules/auth/...
go test -race ./internal/modules/notification/...
go test -race ./cmd/api/...
go test -race ./...
go build ./...

# raíz
pnpm run openapi:check-config
pnpm run openapi:lint
pnpm run openapi:bundle
git diff --check
git status --short
```

`gofmt -l .` debe quedar sin salida. Si no cambió OpenAPI, los comandos confirman que el contrato vigente sigue válido; no regeneres ni comitees artefactos sin cambios. Ejecuta las pruebas PostgreSQL que requieran `TEST_DATABASE_URL` solo contra la base de prueba y roles documentados.

## Evidencia y entrega

Entrega dos tablas:

`Criterio | Estado | Prueba o evidencia segura`

`Dato sensible | Comprobación | Resultado`

Incluye como mínimo:

- selección real vs. marcador;
- sexta solicitud y `challenge-required`;
- Meta aceptó el mensaje;
- recepción física confirmada sin mostrar OTP/teléfono;
- verificación `204`;
- reutilización rechazada;
- login posterior desbloqueado;
- pruebas automatizadas ejecutadas;
- limitación expresa: configuración completa por barbería/evento de `DEC-081` sigue fuera de alcance.

Actualiza el prompt a `executed` solo si produjo el resultado/PR correspondiente. Si la implementación termina pero la entrega real queda bloqueada por Meta, deja claramente separados “código y pruebas automatizadas completos” de “prueba operativa bloqueada”.

## Git y PR

- Rama: `fix/86-reto-otp-whatsapp-meta`.
- Commit/PR: `fix(auth): conecta reto OTP con Meta WhatsApp`.
- Incluye solo rutas de esta preocupación; el `.env` nunca aparece en `git status`, staged diff ni PR.
- Usa `Refs #86` mientras falte cualquier comprobación manual del issue, incluida la evidencia de entrega real o los pendientes que el informe de correo dejó parciales.
- Usa `Closes #86` únicamente si todo el cuerpo vigente del issue queda demostrado, no solo el envío de WhatsApp.
- No hagas push directo, force push ni merge de `main`; no resuelvas conversaciones ni omitas checks por conveniencia.
