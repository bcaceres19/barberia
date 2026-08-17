---
prompt_id: "PROMPT-HU-007-v1"
version: "1.1"
kind: "hu"
status: "blocked"
status_reason: "Solo pendiente de que HU-012 (feat/56-hu012-cascaron-panel) se integre en main, por orden de construcción de B0; ya no hay ninguna decisión normativa abierta."
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-007"
related_hu:
  - "HU-003"
  - "HU-005"
  - "HU-008"
  - "HU-010"
  - "HU-012"
issue: 57
issue_url: "https://github.com/bcaceres19/barberia/issues/57"
suggested_issue_title: "feat(auth): implementar HU-007 defensa escalonada contra abuso"
branch: null
pr: null
pr_url: null
depends_on:
  - "HU-003 integrada"
  - "HU-005 integrada"
  - "HU-012 integrada según el orden de construcción B0"
rules:
  - "RN-DAT-02"
decisions:
  - "DEC-026"
  - "DEC-034"
  - "DEC-035"
  - "DEC-036"
  - "DEC-037"
  - "DEC-038"
  - "DEC-040"
  - "DEC-052"
  - "DEC-061"
  - "DEC-062"
acceptance_criteria:
  - "CA-007-01"
  - "CA-007-02"
  - "CA-007-03"
  - "CA-007-04"
  - "CA-007-05"
  - "CA-007-06"
  - "CA-007-07"
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
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/04-arquitectura/backend-go.md"
  - "docs/04-arquitectura/frontend.md"
  - "docs/04-arquitectura/stack-despliegue-operacion.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/05-backend/revision-ddl-seguridad-2026-08-11.md"
  - "docs/06-api/estandar-openapi.md"
  - "database/modelo-fisico-referencia.sql"
  - "database/README.md"
  - "apps/api/README.md"
  - "apps/api/internal/platform/config"
  - "apps/api/internal/platform/httpserver"
  - "apps/api/internal/modules/auth"
  - "apps/api/cmd/api/main.go"
  - "apps/api/cmd/worker/main.go"
  - "apps/web/src/modules/auth"
  - "api/openapi/openapi.yaml"
  - "api/openapi/paths/public-auth.yaml"
created_at: "2026-08-14"
updated_at: "2026-08-17"
supersedes: "docs/10-backlog/prompts-implementacion.md#prompt-de-hu-007--defensa-escalonada-contra-abuso"
superseded_by: null
---

# Implementar HU-007: defensa escalonada contra abuso en el acceso

## Instrucción para Claude o Codex

Implementa únicamente `HU-007` alrededor del inicio de sesión real de `HU-005`: conteo seguro por IP, ventana y umbral configurables, escalamiento telefónico antes de evaluar la contraseña, respuesta uniforme, expiración y pruebas contra el despliegue previsto. No conviertas esta entrega en recuperación de contraseña ni reutilices sin autorización el código de `HU-008`.

`DP-SEG-10` y `CT-005` quedaron resueltas el 2026-08-17 como `DEC-062` y `DEC-061` (ver sección siguiente). Este prompt sigue `blocked`, pero únicamente porque el orden de construcción de B0 exige que `HU-012` esté **integrada en `main`** antes de empezar `HU-007` — no queda ninguna decisión normativa pendiente. Verifica en preflight que ese PR ya se mergeó antes de abrir rama.

## Objetivo

Frenar intentos automatizados sin enumerar cuentas ni bloquear prematuramente al barbero legítimo: las solicitudes desde una IP se cuentan de forma concurrente y no reconstruible, con valores iniciales de 5 solicitudes, ventana de 15 minutos y escalamiento de 24 horas conforme a `DEC-026`/`DEC-052`. Al cruzar la frontera aprobada, la contraseña no se evalúa hasta completar el reto telefónico aprobado; la respuesta y la experiencia indican de forma segura qué hacer y cuándo vuelve el acceso normal.

## Bloqueos resueltos: `CT-005` → `DEC-061`, `DP-SEG-10` → `DEC-062`

### `CT-005` · La sexta solicitud exige el reto, no la quinta

`DEC-061` fija que, con `p_threshold = 5`, las cinco primeras solicitudes dentro de la ventana se evalúan con normalidad (incluida la contraseña); la **sexta** es la que exige el reto. `login_throttle_register_attempt` en `database/modelo-fisico-referencia.sql` ya quedó corregida (`> p_threshold`, no `>= p_threshold`) — audítala contra la decisión al crear la migración real, no la copies sin verificar.

### `DP-SEG-10` · Reto telefónico completo

`DEC-062` fija el reto de extremo a extremo, sin margen de interpretación:

- `POST /api/v1/public/auth/challenge { email }` → siempre `202` con mensaje genérico, sin importar si el correo existe, si el teléfono está verificado o si la IP está realmente escalada; solo se envía un mensaje real por WhatsApp oficial cuando las tres condiciones se cumplen. Límite propio: 1 solicitud cada 60 s, máximo 3 activas por IP por ventana de 15 min.
- `POST /api/v1/public/auth/challenge/verify { email, code }` → código numérico de 6 dígitos, vigente 5 min, máximo 5 intentos, un solo uso; respuesta idéntica ante código incorrecto/vencido/agotado/cuenta inexistente. Éxito: limpia `escalated_until`/`attempt_count` de esa IP en `login_throttle` (misma función `SECURITY DEFINER`) y responde `204`; el barbero reintenta el login normalmente, sin token adicional que enlazar.
- Tabla nueva `auth_phone_challenge` (`staff_user_id` FK, `ip_hash` = mismo HMAC de `login_throttle`, `code_hash = HMAC-SHA256(código, secreto de despliegue)`, `created_at`, `expires_at`, `attempts_count`, `consumed_at`, `invalidated_at`), RLS tenant-aware vía `staff_user_id → barbershop_id`. `login_throttle` no gana columnas nuevas.

Relee `DEC-061`/`DEC-062` completas en `docs/00-control/registro-decisiones.md` antes de codificar; si encuentras una ambigüedad que esas decisiones no cubran, detente y regístrala como duda nueva en vez de resolverla por inferencia.

## Preflight obligatorio

1. Comprueba árbol limpio, `main` actualizada por fast-forward y ausencia de cambios ajenos.
2. Consulta Graphify por `HU-007`, `LoginService`, `LoginHandler`, configuración, orden de middleware, `login_throttle`, proxy/IP, `429`, `Retry-After`, worker y flujo de acceso Vue.
3. Lee completamente cada `source_docs`; revisa en particular la sección A.4 del modelo de referencia y el hallazgo `DDL-AUT-01`.
4. Confirma que `HU-003`, `HU-005` y, por orden B0, `HU-012` están **integradas en `main`** (PR mergeado, no solo abierto). `HU-003` aporta RFC 9457/logging; `HU-005` es el login que debe protegerse; `HU-012` es la única dependencia que puede seguir pendiente cuando leas esto — si aún no se mergeó, detente y conserva `status: blocked`.
5. `DP-SEG-10` y `CT-005` ya están resueltas (`DEC-062`, `DEC-061`); relee ambas completas antes de codificar.
6. Confirma el alcance exacto del issue `#57`; cambia este archivo a `ready` solo con `HU-012` ya integrada.
7. Crea `feat/57-hu007-defensa-abuso` desde `main` actualizada y cambia a `in_progress` antes de implementar.

## Alcance incluido

- Configuración validada y observable para umbral, ventana, escalamiento, retención/purga y política de proxies confiables; valores iniciales según decisiones vigentes.
- Derivación de IP desde `RemoteAddr` o encabezados emitidos solo por proxies expresamente confiables, con documentación del despliegue y defensa contra spoofing.
- HMAC-SHA256 de la IP con secreto de despliegue; nunca IP en claro ni hash simple de un espacio de baja entropía.
- Migración Atlas roll-forward de `login_throttle`, índice, funciones estrechas y privilegios mínimos; el modelo físico de referencia es insumo, no migración aplicada.
- Conteo atómico/concurrente antes de evaluar la contraseña y reto telefónico completo exactamente como quede aprobado.
- Respuesta `429` RFC 9457, `Retry-After` o información equivalente aprobada, no enumeración y traducción tipada en el cliente web.
- Purga acotada desde el worker con rol `barberia_worker`; API sin DML directo sobre la tabla.
- UI mínima necesaria para completar el reto aprobado y demostrar `CA-007-02` end-to-end, reutilizando el módulo/pantalla de acceso y el sistema visual existentes.
- Pruebas unitarias, HTTP, PostgreSQL real, concurrencia, configuración, proxy y E2E del umbral/reto.

## Fuera de alcance

- Recuperar o cambiar contraseña (`HU-008`/`HU-011`).
- Contar por correo, usuario o barbería, o guardar correo/teléfono/nombre junto al contador.
- Confiar ciegamente en `X-Forwarded-For`, `Forwarded` o cualquier cabecera enviada por el cliente.
- Guardar IP en claro, cifrada de forma reversible, con hash simple o en logs/fixtures/evidencia.
- Implementar CAPTCHA, bloqueo permanente, listas negras, reputación externa o servicio distribuido de rate limit.
- Evaluar la contraseña antes del reto una vez que el escalamiento esté activo.
- Editar una migración aplicada, copiar toda la sección A del modelo de referencia o conceder DML directo a `barberia_app`.
- Elegir el proveedor/canal o reutilizar el código de recuperación sin decisión expresa.

## Estado existente que debes conservar

- `HU-003` centraliza `Problem`, `requestId`, límites, logging seguro y traducción de errores. Reutilízalo; no escribas JSON de error ad hoc.
- `HU-005` implementa `POST /api/v1/public/auth/login`, trabajo uniforme para correo inexistente/contraseña incorrecta y emisión de sesión. El rate limit se integra sin romper la no enumeración ni duplicar el login.
- `HU-010` ya mapea `429` por `status` y muestra un estado explicativo, pero no implementa un reto. Evoluciona ese estado conforme a la decisión, sin DTO manual ni `fetch` en componentes.
- `database/modelo-fisico-referencia.sql` A.4 propone tabla global sin `barbershop_id`, HMAC de IP, UPSERT atómico, función de purga y separación `barberia_app`/`barberia_worker`. Audita y traslada solo lo aprobado a una migración nueva.
- La prueba de revisión DDL ya demostró que una versión `SELECT ... FOR UPDATE` perdía incrementos cuando la fila no existía. Conserva UPSERT atómico y vuelve a probarlo contra la migración real.
- El orden de middleware de `backend-go.md` ubica limitación pública antes de autenticación privada; el login sigue fuera de `/private` por `DEC-055`.

## Trabajo requerido después de desbloquear

1. Actualiza primero OpenAPI con la semántica aprobada: `429`, headers, `Problem.code`, respuestas no enumerables y las operaciones/schemas del reto si existen. Declara `security`, ejemplos ficticios, `x-business-rules` y `x-decisions`; lint/bundle antes del cliente.
2. Añade configuración tipada desde entorno para valores y confianza de proxies. Rechaza secretos ausentes/débiles, rangos inválidos, redes mal formadas y combinaciones de retención incompatibles; la configuración nunca se registra completa.
3. Implementa un resolvedor de IP pequeño y probado. Por defecto usa el par remoto; solo acepta la cadena reenviada cuando el peer inmediato pertenece a la lista de proxies confiables y aplica la semántica documentada del despliegue. Normaliza IPv4/IPv6 sin confundir puertos.
4. Deriva un identificador HMAC estable con secreto de despliegue. No expongas el HMAC al frontend ni lo uses como identificador público; documenta el efecto de rotar la clave según la decisión operativa aprobada.
5. Crea una migración Atlas nueva para la parte A.4 aprobada. Conserva ownership/RLS excepcional documentada, `search_path=''`, nombres calificados, `REVOKE ... FROM PUBLIC`, grants mínimos y validación de parámetros. No edites migraciones previas.
6. Encapsula contador y reto mediante puertos del módulo `auth` o plataforma con dueño claro. El núcleo no importa Chi/pgx; el handler no ejecuta SQL. Registra el intento y decide el escalamiento antes de `PasswordHasher.Verify` conforme a la frontera resuelta por `CT-005`.
7. Asegura que cuenta existente/inexistente, contraseña correcta/incorrecta y reto requerido no creen un oráculo adicional. La rama escalada no resuelve tenant ni evalúa contraseña antes de la prueba adicional si así lo fija la decisión.
8. Implementa el reto telefónico y su experiencia web exactamente como apruebe `DP-SEG-10`, con cliente generado, estados accesibles, doble envío controlado y datos no sensibles conservados.
9. Integra purga en lotes pequeños en `cmd/worker` reutilizando la configuración aprobada; no mantengas una transacción abierta durante E/S externa.
10. Actualiza el modelo de referencia solo para mantenerlo consistente con la migración/decisión; una diferencia material se registra antes de codificar.

## Pruebas y evidencia

- Unitarias con reloj inyectado: debajo de umbral, frontera exacta resuelta, escalamiento activo, ventana vencida, escalamiento aún vigente, dos IP y configuración inválida.
- Resolvedor de IP: conexión directa, proxy confiable, múltiples saltos según política aprobada, peer no confiable con cabecera falsa, IPv4, IPv6 y entradas malformadas.
- HMAC: estabilidad con misma clave/IP, separación por IP/clave y ausencia de IP en resultados/logs; nunca uses direcciones reales.
- PostgreSQL 14 real: migración en vacío y actualización con datos; grants/ownership; parámetros; purga; 40 o más llamadas concurrentes coordinadas sin `sleep` y sin incrementos perdidos; dos IP independientes.
- HTTP: solicitudes válidas debajo de umbral, cruce de frontera, `429` uniforme, header de reintento, correo existente/inexistente indistinguible, contraseña no evaluada durante reto y cabecera manipulada incapaz de evadir.
- Frontend/componente: estado/reto aprobado, foco, teclado, espera/reintento, doble toque y mapeo por `status`/`code`, nunca por `detail`.
- E2E con API/PostgreSQL reales: recorrido del umbral y reto telefónico completo, expiración de ventana y dos IP. Terceros se interceptan; no envíes mensajes reales.
- Auditoría de logs/fixtures/capturas: cero IP, correo, teléfono, contraseña, token o código en claro.

## Criterios de aceptación y cierre

Entrega `Criterio | Estado | Prueba o evidencia` para `CA-007-01` a `CA-007-07`.

- No declares `CA-007-02` con un simple `429`: exige completar el reto aprobado sin evaluar antes la contraseña.
- `CA-007-03` incluye forma uniforme, no enumeración e indicación verificable de reintento.
- `CA-007-04` distingue ventana del contador y escalamiento de 24 horas según la decisión; no acortes uno para hacer pasar la prueba.
- `CA-007-05` requiere cambiar parámetros por configuración de despliegue y reiniciar, sin recompilar ni editar código.
- `CA-007-06` exige purga real y revisión de datos/logs, no solo una columna `expires_at`.
- `CA-007-07` se prueba contra la topología documentada, incluyendo control negativo con cabecera falsificada.

## Documentación y trazabilidad

- Actualiza contrato/CHANGELOG, cliente generado, `apps/api/README.md`, `apps/web/README.md`, `database/README.md`, matriz e historial según el cambio real.
- Documenta topología de proxy prevista, fuente de IP y configuración/secreto sin incluir valores sensibles.
- Versiona migración y `atlas.sum`; registra bloqueo, compatibilidad y roll-forward.
- Actualiza este prompt/catálogo con estado, rama y PR reales. `DP-SEG-10`/`CT-005` ya están resueltas (`DEC-062`/`DEC-061`); no marques `ready` mientras `HU-012` no esté integrada en `main`.

## Verificación final

```text
# raíz
pnpm run openapi:check-config
pnpm run openapi:lint
pnpm run openapi:bundle

# database, con PostgreSQL real y roles documentados
atlas migrate hash --dir file://database/migrations
atlas migrate validate --env local
atlas migrate apply --env local --dry-run
atlas migrate apply --env local
atlas migrate status --env local

# apps/api
gofmt -l .
go vet ./...
go test -race ./...

# apps/web
pnpm run generate:api
pnpm run format
pnpm run lint
pnpm run typecheck
pnpm run test:unit
pnpm run build
pnpm run test:e2e

# raíz
git diff --check
```

Ejecuta los comandos desde el directorio al que corresponde cada bloque. Las migraciones se prueban además sobre la versión anterior con datos y con los roles reales; una aplicación vacía no basta. `gofmt -l .` debe quedar sin salida.

## Git y PR

- Rama: `feat/57-hu007-defensa-abuso`.
- Commit/PR: `feat(auth): implementa HU-007 defensa escalonada contra abuso`.
- Usa `Closes #57` solo con `DP-SEG-10`/`CT-005` resueltas y `CA-007-01`–`CA-007-07` probadas completas; de lo contrario usa `Refs #57` y no marques el PR listo.
- No mezcles `HU-008`, no hagas push directo/force push/merge de `main` y no omitas conversaciones ni checks.

