---
prompt_id: "PROMPT-HU-006-v1"
version: "1.2"
kind: "hu"
status: "in_progress"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-006"
related_hu:
  - "HU-005"
  - "HU-012"
issue: 45
issue_url: "https://github.com/bcaceres19/barberia/issues/45"
suggested_issue_title: "feat(auth): implementar HU-006 sesión persistente y cierre"
branch: "feat/45-hu006-sesion-persistente"
pr: null
pr_url: null
depends_on:
  - "HU-005 integrada"
rules:
  - "RN-TEN-01"
  - "RN-DAT-02"
decisions:
  - "DEC-024"
  - "DEC-026"
  - "DEC-034"
  - "DEC-035"
  - "DEC-037"
  - "DEC-038"
  - "DEC-050"
  - "DEC-055"
  - "DEC-058"
acceptance_criteria:
  - "CA-006-01"
  - "CA-006-02"
  - "CA-006-03"
  - "CA-006-04"
  - "CA-006-05"
  - "CA-006-06"
  - "CA-006-07"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/04-arquitectura/backend-go.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/06-api/estandar-openapi.md"
  - "database/modelo-fisico-referencia.sql"
  - "apps/api/README.md"
  - "api/openapi/openapi.yaml"
created_at: "2026-08-13"
updated_at: "2026-08-13"
supersedes: "docs/10-backlog/prompts-implementacion.md#prompt-de-hu-006--sesión-persistente-y-cierre-de-sesión"
superseded_by: null
---

# Implementar HU-006: sesión persistente y cierre de sesión

## Instrucción para Claude o Codex

Implementa únicamente `HU-006` sobre la autenticación integrada de `HU-005`. Protege el área privada, renueva sesiones activas y permite revocar la sesión actual de forma inmediata en el servidor. No agregues recuperación de acceso, rate limiting, panel Vue ni cierre masivo de dispositivos.

## Objetivo

Una cookie de sesión válida debe sobrevivir al cierre del navegador y autorizar solicitudes privadas durante una vigencia deslizante de 30 días. Una sesión vencida o revocada no ejecuta ninguna operación; cerrar sesión invalida solo el dispositivo actual; toda ruta bajo `/api/v1/private` pasa por un middleware inventariable y probado, sin excepción alguna (`CT-003` resuelta por `DEC-055`: el login ya no vive en ese subrouter).

## Preflight obligatorio

1. Comprueba árbol limpio, `main` actualizada por fast-forward y ausencia de cambios ajenos que puedas sobrescribir.
2. Consulta Graphify por `HU-006`, middleware privado, `staff_session`, renovación, revocación, routing y OpenAPI si existe el grafo.
3. Lee completamente cada `source_docs`. Confirma que `HU-005` está integrada, su migración aplicada y su tabla de criterios en verde.
4. `CT-003` está resuelta (`DEC-055`): el login vive en `/api/v1/public/auth/login`, fuera de `/api/v1/private`. La prueba estructural de `CA-006-04` exige el middleware en el 100 % del subrouter privado, sin ninguna lista de excepciones para el login ni para ninguna otra ruta.
5. `DP-SEG-08` está resuelta (`DEC-058`): `HU-005` (issue `#44`) probó el aislamiento de sesión entre barberías solo a nivel PostgreSQL/RLS porque no existía ningún endpoint privado real. Esta historia añade `CA-006-07`, que completa esa verificación end-to-end usando el logout —el primer endpoint privado real— con dos tenants reales.
6. Busca o crea, con autorización, un issue que cubra solo `HU-006`. Con `issue: pending` o dependencia incumplida (`HU-005` sin integrar), conserva `status: draft` y no cambies código.
7. Registra issue y URL, cambia a `ready`, crea `feat/<issue>-hu006-sesion-persistente` desde `main` y actualiza el prompt a `in_progress` con la rama real.

## Alcance incluido

- Middleware de sesión para las rutas privadas protegidas, en el orden aprobado respecto a seguridad, rate limit futuro, tenant/RLS e idempotencia.
- Lectura segura del token opaco desde cookie, hash antes de consultar y resolución del tenant sin aceptar `barbershop_id` del cliente.
- Validación de sesión no revocada/no vencida, pertenencia al usuario y barbería, y usuario activo.
- Actualización de `last_used_at` y `expires_at` para mantener 30 días desde la última actividad, sin revivir sesiones vencidas.
- Operación contract-first de cierre de sesión que fija `revoked_at` en el servidor y elimina la cookie con los mismos atributos con que fue creada.
- Respuestas uniformes `401` sin filtrar si el token era desconocido, vencido, revocado o de un usuario inactivo.
- Inventario automático de rutas y prueba estructural de cobertura de middleware sobre el 100 % de `/api/v1/private`, sin lista de excepciones (`DEC-055`).
- Verificación end-to-end de que la sesión de la barbería A nunca ejecuta el logout de la barbería B (`CA-006-07`, `DEC-058`), con dos tenants reales.

## Fuera de alcance

- Cambiar correo, contraseña, algoritmo de hash o emisión inicial de `HU-005` salvo una corrección indispensable y trazable.
- Revocar todos los dispositivos, listar sesiones o administrar dispositivos.
- Recuperación de acceso y revocación por cambio de contraseña de `HU-008`.
- Rate limiting, IP o verificación telefónica de `HU-007`.
- Pantalla de acceso, cascarón privado o manejo de sesión en Vue (`HU-010`/`HU-012`).
- JWT, refresh token, caché de revocación en memoria o almacenamiento del token opaco en claro.
- Borrar físicamente una sesión para representar logout; `DEC-050` exige revocación observable mediante `revoked_at`.

## Estado existente que debes conservar

- `HU-005` debe haber creado `staff_session`, la cookie, la resolución previa del tenant y el núcleo del módulo `auth`.
- El valor en claro solo vive en la cookie; PostgreSQL conserva SHA-256 hexadecimal y globalmente único.
- Toda operación de datos usa `InTenantTx`; no expongas el pool ni ejecutes SQL desde middleware o handlers fuera de un repositorio/adaptador con dueño claro.
- El logger HTTP de `HU-003` no registra ruta cruda, cookie, autorización ni tokens. Los problemas reutilizan su traductor central.
- Las migraciones aplicadas son inmutables. Si `HU-006` necesita un cambio de esquema no previsto, crea una migración roll-forward solo después de demostrar la necesidad.

## Trabajo requerido

1. Audita el flujo de sesión de `HU-005` y separa con claridad extracción/transporte HTTP, autenticación de aplicación y persistencia. El núcleo no importa Chi, `net/http` ni pgx.
2. Extrae la cookie sin aceptar el token en query string, fragmento, body, `localStorage` o encabezado alternativo no aprobado. Rechaza ausencia, forma inválida y tamaño excesivo antes de tocar PostgreSQL.
3. Calcula el hash y resuelve el tenant mediante la función estrecha aprobada. Abre la transacción con ese tenant y vuelve a comprobar sesión, usuario y barbería bajo RLS; la resolución previa no sustituye la autorización final.
4. Acepta solo sesiones con `revoked_at IS NULL`, `expires_at > now()` y usuario activo. Usa reloj inyectado. Una sesión vencida nunca se renueva ni se reabre.
5. En cada uso válido, actualiza `last_used_at` y extiende `expires_at` a 30 días desde la actividad conforme a `DEC-050`, dentro de una transacción corta. Evita una carrera en la que una renovación posterior resucite una sesión que acaba de revocarse.
6. Propaga un contexto tipado con identificadores opacos de usuario, sesión y barbería. El tenant nunca se toma de URL, body o cabecera controlada por el cliente.
7. Monta el middleware una sola vez en el subrouter protegido, cubriendo el 100 % de `/api/v1/private` sin allowlist ni excepción de arranque sin sesión (`DEC-055`: el login ya no está en este subrouter).
8. Declara cierre de sesión en OpenAPI antes del handler. La operación revoca solo la fila de la sesión actual, es idempotente en su efecto observable y limpia la cookie con Path/Domain/SameSite/Secure coherentes. No toca las otras sesiones del usuario.
9. Mantén los `401` indistinguibles para token desconocido, vencido, revocado, usuario inactivo o desajuste tenant/usuario. Registra solo `request_id` e identificadores opacos permitidos, nunca el token ni su cookie.
10. Crea un inventario estructural desde el router real. La prueba falla al registrar una nueva ruta protegida fuera del middleware; no hay excepciones aprobadas que mantener (`DEC-055`).

## Pruebas y evidencia

- Unitarias: token ausente/malformado, hash, sesión válida, vencida, revocada, usuario inactivo, desajuste usuario/tenant, reloj y fallos de repositorio.
- HTTP: reutilizar la cookie en un cliente/navegador nuevo dentro de vigencia funciona sin credenciales; ausencia, expiración y revocación devuelven el mismo `401`; el material nunca aparece en URL, body, problemas ni logs.
- Persistencia real: `last_used_at` y `expires_at` avanzan con actividad válida; no cambian para sesión vencida/revocada; RLS y dos tenants impiden cruces.
- Logout: fija `revoked_at`, limpia la cookie y el mismo material falla de inmediato; repetir el logout no ejecuta una operación privada ni produce un error revelador.
- Dos sesiones del mismo usuario: cerrar A invalida A y conserva B (`CA-006-06`).
- Dos tenants reales: la sesión de A invoca el logout con la sesión de B (o viceversa) y el resultado es rechazo uniforme sin afectar la sesión objetivo; completa `CA-005-05` de `HU-005` (`CA-006-07`, `DEC-058`).
- Carrera real renovación/logout coordinada sin `sleep`: el resultado final siempre queda revocado y ninguna actualización posterior elimina `revoked_at` o reabre la sesión.
- Prueba estructural: inventario de rutas bajo la topología aprobada, middleware presente en todas las protegidas y allowlist exacta para excepciones autorizadas.
- OpenAPI: esquema de cookie, seguridad por operación, `401`, logout, headers y ejemplos ficticios coinciden con handlers y cliente futuro.
- `go test -race` para middleware, renovación y carrera de revocación.

Entrega una tabla `Criterio | Estado | Prueba o evidencia` para `CA-006-01` a `CA-006-07`. No declares `CA-006-01` solo con `Max-Age`: demuestra reutilización después de recrear el cliente; no declares `CA-006-04` con una lista manual desconectada del router; no declares `CA-006-07` sin ejecutar el logout real con la sesión de un tenant distinto.

## Documentación y trazabilidad

- Actualiza `apps/api/README.md` con vigencia, renovación, revocación, topología de rutas y ejecución de pruebas, sin valores de sesión reales.
- Actualiza OpenAPI, matriz, historial y documentación de operación realmente afectadas.
- Si aparece una migración, documenta bloqueo, aplicación desde versión anterior y avance/recuperación; versiona `atlas.sum`.
- Actualiza este prompt y el índice con issue, rama, PR y estado reales. Un cambio material durante ejecución crea `v2` y enlaza `supersedes`.

## Verificación final

Ejecuta los comandos equivalentes del repositorio para:

```text
gofmt, go vet y go build
go test ./...
go test -race ./...
pruebas PostgreSQL con dos tenants y rol barberia_app
prueba concurrente de renovación frente a revocación
openapi:check-config
openapi:lint
openapi:bundle
pruebas HTTP, contrato e inventario de rutas
atlas migrate validate si existe cambio SQL
git diff --check
graphify update .
```

No simules la revocación solo borrando la cookie del cliente y no uses caché local como autoridad de sesión.

## Git y PR

- Commits y título: `feat(auth): implementa sesión persistente y cierre` o equivalente Conventional Commits.
- Abre un PR borrador contra `main`; usa `Closes #<issue>` solo con los seis criterios probados.
- Incluye evidencia de persistencia, expiración, revocación inmediata, aislamiento entre dispositivos/tenants, inventario de rutas y contrato.
- No hagas push directo, force push, merge manual ni reescribas `main`.
