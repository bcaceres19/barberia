---
prompt_id: "PROMPT-TEST-OTP-EMAIL-RESEND-v1"
report_for: "docs/10-backlog/prompts/test/issue-86-otp-correo-resend.md"
issue: 86
issue_url: "https://github.com/bcaceres19/barberia/issues/86"
executed_at: "2026-09-02"
executed_by: "agente Claude (sesión interactiva), operador bcaceres19"
environment: "efímero: contenedor postgres:14 + contenedor golang:1.25-bookworm (docker run ad-hoc), destruidos al finalizar"
---

# Evidencia: prueba manual real de OTP por correo (Resend, issue #86)

## Resumen

Con el dominio `nava.wtf` ya verificado en Resend y `APP_RESEND_API_KEY` provista
por el operador, se completó el recorrido real de `HU-008`/`HU-011`
(`solicitar recuperación → recibir OTP en un correo real → verificar OTP →
cambiar contraseña → iniciar sesión con la contraseña nueva → confirmar que
el código no puede reutilizarse`) contra un entorno local efímero, con el
remitente único por correo (`EmailOnlyRecoverySender`) descrito en
`apps/api/README.md`.

La cuenta sintética usada fue `duena.a` (fixture `database/testdata/
hu007_reto_telefonico.sql`, teléfono ya verificado), con su correo
reapuntado —solo dentro de la base de datos efímera, nunca comiteado— a un
correo real controlado por el operador para poder recibir el envío de
Resend.

## Entorno usado (efímero, destruido al finalizar)

- Red Docker dedicada + contenedor `postgres:14` con las 16 migraciones de
  `database/migrations/` aplicadas en el orden documentado y los fixtures
  `dos_barberias.sql`/`hu005_credenciales_sesiones.sql`/
  `hu007_reto_telefonico.sql` cargados (mismo procedimiento que
  `.github/workflows/ci.yml`).
- Contenedor `golang:1.25-bookworm` ejecutando `go run ./cmd/api` con
  `APP_ENVIRONMENT=local`, `APP_RESEND_FROM_ADDRESS=no-responder@nava.wtf` y
  `APP_RESEND_API_KEY` inyectada por `docker run --env-file`, generado por el
  propio operador en su terminal — el agente nunca leyó el valor de la key.
- `apps/web` servido con `pnpm run dev`, proxy de Vite apuntando al
  contenedor de la API.
- Los tres componentes (Postgres, API, base de datos) se eliminaron
  (`docker rm -f`, `docker network rm`) al terminar; no queda estado
  persistente de esta prueba.

## Recorrido ejecutado

1. Pantalla real `/recuperar-acceso`, paso 1: se solicitó el código con el
   correo real de la cuenta sintética. Respuesta genérica confirmada.
2. El correo llegó a la bandeja real en menos de un minuto, remitente
   `no-responder@nava.wtf`, asunto "Código de recuperación de acceso". La
   duración registrada en el log del servidor para esa solicitud
   (`483 ms`) es consistente con una llamada REST real a Resend, frente a
   `2 ms` de una solicitud de control resuelta por el marcador de posición
   (`LoggingRecoveryCodeSender`, sin credenciales completas).
3. Paso 2: el código de 6 dígitos recibido se introdujo en la pantalla real
   y avanzó al paso 3.
4. Paso 3: la pantalla mostró el teléfono/correo enmascarados
   (`+57 *** *** 01`, `b***@g***.com`), confirmando `CA-008-06`.
5. Se estableció una contraseña nueva sintética (10–128 caracteres, distinta
   del correo y de la contraseña anterior) y la API respondió `204`. La UI
   mostró "Contraseña actualizada".
6. Se confirmó contra la API real (no solo la UI):
   - Reutilizar el mismo código ya consumido → `401` (rechazado).
   - Iniciar sesión con la contraseña **anterior** (la del fixture) →
     `401` (rechazado).
   - Iniciar sesión con la contraseña **nueva** → `200` (aceptado).

## Limitación honesta: revocación de sesión no re-demostrada en esta sesión

El paso "demuestra que una sesión anterior queda revocada" del prompt
**no** se completó de forma manual en esta ejecución: se creó una sesión
previa al cambio de contraseña (login exitoso, `200` contra
`GET /api/v1/private/auth/session`) y se intentó un segundo ciclo completo
de recuperación para verificar que esa sesión quedara revocada tras el
segundo cambio, pero el segundo `reset-password` por API falló dos veces por
errores de forma del propio agente (cuerpo sin `email`, luego un token mal
extraído de un archivo temporal) — nunca por un fallo del sistema bajo
prueba. Por disciplina de evidencia, **no se declara este criterio
verificado manualmente hoy**: `CA-008-05` sigue cubierto únicamente por la
suite automatizada existente y documentada en `apps/api/README.md`
(`TestRecovery_System_FullJourney_RequestVerifyResetLogsInWithNewPassword`,
`hu008_recuperacion_acceso.sql`), no por esta prueba manual.

## Tabla de criterios

| Criterio | Estado | Prueba o evidencia |
| --- | --- | --- |
| Entrega real por Resend (issue #86) | Cumplido | Correo real recibido en la bandeja del operador; `duration_ms` del log del servidor (483 ms) consistente con la llamada REST real, contra 2 ms del marcador de posición |
| No enumeración (`CA-008-01`) | Cumplido | Respuesta genérica `202` en la UI real y por API con correo inexistente |
| Destino enmascarado (`CA-008-06`) | Cumplido | Captura de la pantalla del paso 3: `+57 *** *** 01`, `b***@g***.com` |
| Código de un solo uso | Cumplido | Reintentar el mismo código tras el cambio de contraseña → `401` |
| Cambio de contraseña efectivo | Cumplido | Login con contraseña anterior → `401`; login con contraseña nueva → `200` |
| Revocación de sesión (`CA-008-05`) | **No verificado manualmente hoy** | Cubierto solo por la suite automatizada existente (ver limitación arriba); el segundo ciclo manual no llegó a completarse por un error del propio agente, no del sistema |

## Datos sensibles: revisión

| Dato sensible | Resultado | Evidencia segura |
| --- | --- | --- |
| `APP_RESEND_API_KEY` | Nunca vista por el agente | Inyectada al contenedor vía `docker run --env-file <ruta local, .gitignore>`, escrita por el operador en su propia terminal |
| OTP (dos códigos, uno por ciclo) | Nunca escrito en archivos, commits ni este reporte | Leído solo desde la vista de Gmail dentro del navegador automatizado y tecleado directamente en la UI/curl; ninguna herramienta de escritura a disco lo persistió |
| Contraseña nueva sintética | Nunca mostrada en texto plano en este reporte | Usada solo en memoria durante la sesión; la base de datos que la contenía (hasheada) fue destruida al finalizar |
| Correo real del operador | Usado solo dentro del entorno efímero | Nunca comiteado; la fila de `staff_user` que lo contenía vivía en un contenedor Postgres destruido al finalizar |
| Base de datos completa | Sin persistencia | `docker rm -f`/`docker network rm` al finalizar; sin volumen nombrado |

## Verificación final

No se ejecutó `go test -race ./...` en esta sesión (fuera de alcance: esta
prueba es la validación manual con correo real, no una re-ejecución de la
suite automatizada ya verificada en PR #130/CI). `gofmt`/`go vet`/`go build`
tampoco se repitieron: no se modificó código Go en esta sesión, solo se
ejecutó el binario ya existente en `main`.
