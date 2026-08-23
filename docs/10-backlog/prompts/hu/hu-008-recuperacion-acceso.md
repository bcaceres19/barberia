---
prompt_id: "PROMPT-HU-008-v1"
version: "1.4"
kind: "hu"
status: "executed"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-008"
related_hu:
  - "HU-005"
  - "HU-006"
  - "HU-007"
  - "HU-011"
issue: 58
issue_url: "https://github.com/bcaceres19/barberia/issues/58"
suggested_issue_title: "feat(auth): implementar HU-008 recuperación de acceso"
branch: "feat/58-hu008-recuperacion-acceso"
pr: 63
pr_url: "https://github.com/bcaceres19/barberia/pull/63"
depends_on:
  - "HU-005 integrada"
  - "HU-006 integrada"
  - "HU-007 integrada"
rules:
  - "RN-TEN-01"
  - "RN-DAT-01"
  - "RN-DAT-02"
  - "RN-REC-04"
  - "RN-REC-05"
decisions:
  - "DEC-024"
  - "DEC-026"
  - "DEC-027"
  - "DEC-034"
  - "DEC-035"
  - "DEC-036"
  - "DEC-037"
  - "DEC-038"
  - "DEC-040"
  - "DEC-050"
  - "DEC-051"
  - "DEC-057"
  - "DEC-063"
  - "DEC-064"
  - "DEC-065"
  - "DEC-066"
acceptance_criteria:
  - "CA-008-01"
  - "CA-008-02"
  - "CA-008-03"
  - "CA-008-04"
  - "CA-008-05"
  - "CA-008-06"
  - "CA-008-07"
  - "CA-008-08"
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
  - "docs/04-arquitectura/stack-despliegue-operacion.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/05-backend/revision-ddl-seguridad-2026-08-11.md"
  - "docs/06-api/estandar-openapi.md"
  - "database/modelo-fisico-referencia.sql"
  - "database/README.md"
  - "apps/api/README.md"
  - "apps/api/internal/modules/auth"
  - "apps/api/internal/modules/notification"
  - "apps/api/cmd/api/main.go"
  - "apps/api/cmd/worker/main.go"
  - "api/openapi/openapi.yaml"
  - "api/openapi/paths/public-auth.yaml"
created_at: "2026-08-14"
updated_at: "2026-08-23"
supersedes: "docs/10-backlog/prompts-implementacion.md#prompt-de-hu-008--recuperación-de-acceso-con-código"
superseded_by: null
---

# Implementar HU-008: recuperación de acceso con código

## Instrucción para Claude o Codex

Implementa únicamente `HU-008` de extremo a extremo en backend: solicitar recuperación, verificar un código de un solo uso y establecer una contraseña nueva, con envío por WhatsApp oficial y correo conforme a `DEC-051`, no enumeración, invalidación de sesiones y pruebas reales. No construyas la pantalla Vue de `HU-011` ni conviertas el código de recuperación en el reto de `HU-007`.

`DP-SEG-11`, `DP-SEG-12`, `CT-006` y `DP-NOT-05` quedaron resueltas el 2026-08-17 como `DEC-063`–`DEC-066` (ver sección siguiente). `HU-007` se integró en `main` el 2026-08-17 ([PR #61](https://github.com/bcaceres19/barberia/pull/61)), así que este prompt ya está **`ready`**: no queda ninguna decisión ni dependencia pendiente. Confirma en preflight que tu copia local de `main` incluye ese commit antes de crear la rama.

## Objetivo

Permitir que un barbero recupere acceso sin intervención del propietario: una solicitud pública no revela si el correo existe; una cuenta válida recibe el mismo código por los canales oficiales aprobados; el código se guarda solo en una representación resistente a recuperación, vence, limita intentos y reenvíos y se consume una vez; una verificación exitosa autoriza exactamente un cambio de contraseña; el cambio actualiza la credencial e invalida todas las sesiones activas en una misma frontera consistente.

## Bloqueos resueltos: `DP-SEG-11` → `DEC-063`, `DP-SEG-12` → `DEC-064`, `CT-006` → `DEC-065`

### `DP-SEG-11` · Política de contraseña nueva

`DEC-063`: longitud 10–128 caracteres, sin exigencia de composición, rechazo si es igual al correo de la cuenta o a la contraseña actual; sin verificación contra lista externa de contraseñas filtradas en este MVP (queda como mejora futura). El mensaje de rechazo indica exactamente qué regla incumple.

### `DP-SEG-12` · Parámetros y autorización del flujo

`DEC-064`: código de 6 dígitos numéricos (`crypto/rand`); almacenado como `HMAC-SHA256(código, secreto de despliegue)`, nunca `SHA-256` simple; vigencia 15 min; máximo 5 intentos; reenvío con cooldown de 60 s y máximo 3/hora, invalidando atómicamente el código anterior. Solicitar y verificar se hacen siempre contra `email` (sin identificador opaco nuevo). Verificar con éxito emite un token opaco de reinicio de un solo uso, vigente 5 min (mismo patrón `CryptoTokenGenerator`/`HashToken` que el token de sesión de `HU-005`/`HU-006`), que el paso de cambiar contraseña debe presentar junto con `email`.

### `CT-006` · No enumeración frente a destino enmascarado

`DEC-065` (opción 1): `POST /recovery/request` siempre responde idéntico, sin destino, exista o no la cuenta. El destino enmascarado solo aparece en la respuesta exitosa de `POST /recovery/verify` — llegar ahí ya exige haber recibido y transcrito el código real, así que no abre un oráculo nuevo.

### `DP-NOT-05` · Proveedor oficial concreto

`DEC-066`: WhatsApp por **Meta WhatsApp Cloud API directo** (plantilla "Authentication" pre-aprobada, sin BSP intermediario); correo por **Resend** (o **SES** si el despliegue ya usa AWS). Puerto único en el módulo `notification`, timeout de 5 s por canal, sin reintento síncrono dentro de la solicitud HTTP, tolerante a fallo parcial de un canal sin cambiar la respuesta genérica de `DEC-065`. Credenciales por variable de entorno de despliegue, nunca en el repo; configuración validada al arrancar. Un puerto con doble de prueba (fake) no satisface por sí solo "funciona con el adaptador seleccionado": implementa también el adaptador real contra Meta/Resend.

Relee `DEC-063`–`DEC-066` completas en `docs/00-control/registro-decisiones.md` antes de codificar; si encuentras una ambigüedad que no cubran, detente y regístrala como duda nueva en vez de resolverla por inferencia.

## Preflight obligatorio

1. Comprueba árbol limpio, `main` actualizada por fast-forward y ausencia de cambios ajenos.
2. Consulta Graphify por `HU-008`, `staff_recovery_code`, `staff_credential`, `staff_session`, `PasswordHasher`, sesión masiva, `notification`, OpenAPI y worker.
3. Lee completamente cada `source_docs`; revisa la sección A.3 del modelo de referencia y los privilegios sensibles de `DEC-040`/`DDL-AUT-01`.
4. Confirma que `HU-005`, `HU-006` y `HU-007` están **integradas en `main`** (PR mergeado) y sus pruebas pasan. `HU-007` no se considera satisfecha por un estado `429` simulado en frontend, sino por el reto real de `DEC-062` funcionando end-to-end.
5. `DP-SEG-11`, `DP-SEG-12`, `CT-006` y `DP-NOT-05` ya están resueltas (`DEC-063`–`DEC-066`); relee las cuatro completas antes de codificar.
6. Confirma que el issue `#58` cubre solo `HU-008`; cambia a `ready` únicamente con todas las dependencias satisfechas.
7. Crea `feat/58-hu008-recuperacion-acceso` desde `main` actualizada y cambia el prompt a `in_progress`.

## Alcance incluido

- Migración Atlas roll-forward de `staff_recovery_code` con FK tenant-aware, RLS forzada, índices mínimos, restricciones temporales y privilegios revisados.
- Operaciones públicas contract-first para solicitar, verificar y establecer contraseña nueva, con schemas cerrados, no enumeración y errores RFC 9457.
- Generación criptográfica del código/identificadores según la decisión; persistencia solo de la representación aprobada, nunca el valor enviado.
- Código único vigente por usuario, vencimiento, máximo de intentos, consumo/invalidez terminal y reenvío que invalida el anterior de forma atómica.
- Entrega por WhatsApp oficial y correo mediante puerto/adaptador seleccionado; pruebas usan dobles y nunca red real.
- Teléfono/correo enmascarados únicamente en el momento y forma resueltos por `CT-006`.
- Política de contraseña aprobada, hash Argon2id reutilizado y cambio atómico con revocación de todas las sesiones del usuario.
- Integración con defensa de abuso de `HU-007` donde la decisión lo exija, sin compartir semánticas de reto por inferencia.
- Pruebas unitarias, HTTP/contrato, PostgreSQL real con dos tenants, concurrencia, proveedor y revisión de logs.

## Fuera de alcance

- Pantallas, indicador de pasos, cuenta regresiva o navegación Vue de `HU-011`.
- Reutilizar el código como verificación telefónica de `HU-007` sin una decisión explícita.
- Enviar mensajes reales desde pruebas, incluir credenciales del proveedor o agregar secretos al repositorio.
- Guardar código, contraseña, teléfono, correo o contenido de mensaje en logs, problemas HTTP, fixtures o evidencia.
- Revelar si una cuenta existe mediante body, status, headers funcionales, destino enmascarado, tiempos evitables o errores del proveedor.
- Permitir más de un código vigente, reiniciar silenciosamente el vencimiento anterior o aceptar el mismo código dos veces.
- Cambiar la contraseña y revocar sesiones en transacciones separadas que puedan dejar un estado parcial.
- Editar migraciones aplicadas, implementar la maquinaria general de notificaciones B5 o incorporar un proveedor distinto al aprobado.

## Estado existente que debes conservar

- `HU-005` entrega `staff_credential`, Argon2id, no enumeración del login, resolución previa del tenant y emisión de sesión opaca.
- `HU-006` entrega validación/renovación/revocación de sesión y logout; extiende su repositorio con una operación masiva explícita, no con SQL desde el handler.
- `HU-007` debe aportar la defensa pública integrada antes de esta historia. No dupliques resolvedor de IP, contador o configuración.
- `database/modelo-fisico-referencia.sql` A.3 diseña `staff_recovery_code`, índice activo único, RLS y campos temporales. Es referencia: audítala contra las decisiones nuevas y crea una migración, no la declares aplicada.
- `staff_user` contiene el teléfono/verificación que gobierna `DEC-026`; confirma el esquema migrado real y no confíes solo en el modelo de referencia.
- El módulo `notification` todavía no contiene un proveedor. Respeta los límites de módulos: `auth` consume un puerto de entrega pequeño; el adaptador no filtra tipos internos del proveedor al núcleo.
- `PasswordHasher.Hash` ya produce PHC Argon2id con sal aleatoria. Reutilízalo después de validar la política aprobada; no cambies parámetros de login sin issue propio.

## Trabajo requerido después de desbloquear

1. Diseña contract-first las tres operaciones y los artefactos aprobados. Define status, schemas, seguridad pública, límites de cuerpo, ejemplos ficticios, `Problem.code`, no enumeración, idempotencia/reintento cuando aplique y trazabilidad `RN-*`/`DEC-*`. Alinea `CT-006` antes de escribir handlers.
2. Crea la migración Atlas de `staff_recovery_code` basada en A.3 y las decisiones resueltas. Revisa ownership, RLS, FK compuestas, índices, `CHECK`, unicidad activa y grants; usa funciones estrechas `SECURITY DEFINER` solo si reducen realmente privilegios y cumplen `search_path=''`.
3. Modela el núcleo con reloj, generador de código/identificadores, hasher/verificador, repositorio y puerto de entrega. Servicios no importan `net/http`, Chi, pgx ni SDK del proveedor.
4. Solicitud: normaliza entrada, aplica defensa aprobada, mantiene respuesta no enumerable, invalida/reemplaza de forma atómica el código previo y persiste antes de cualquier E/S. Sigue la decisión de proveedor para entrega dual y fallos parciales; nunca mantengas transacción abierta durante red.
5. Usa una representación almacenada resistente a recuperación offline para el código corto exactamente como apruebe `DP-SEG-12`; un SHA-256 simple de un espacio pequeño no se vuelve seguro solo por llamarse “hash”.
6. Verificación: decide todo en una escritura/función atómica con reloj controlado. Un intento fallido incrementa sin carrera; al agotar, invalida; vencido/usado/invalido no revive; éxito consume o emite exactamente el artefacto aprobado para el tercer paso.
7. Cambio de contraseña: valida el artefacto de autorización y la política `DP-SEG-11`, deriva el nuevo hash y actualiza credencial más revocación de todas las sesiones en una misma transacción tenant-aware. El artefacto se consume una vez incluso bajo dos solicitudes concurrentes.
8. Traduce errores a respuestas seguras. Los mensajes internos/proveedor se depuran antes de log; `requestId` permite soporte sin destinatario ni código.
9. Implementa el adaptador seleccionado con timeout/cancelación/idempotencia y configuración validada. Las pruebas del contrato del proveedor se separan y no contienen credenciales.
10. Regenera cliente tipado para `HU-011`, pero no construyas sus páginas. La API pública del módulo debe permitir a la historia siguiente consumir el contrato sin importar internos de backend.

## Pruebas y evidencia

- Unitarias con reloj/generador determinista ficticio: solicitud, vencimiento, un solo uso, intentos 0/límite, reenvío antes/después de espera, política de contraseña, cancelación de contexto y errores de dependencias.
- No enumeración: correo existente/inexistente con mismo status/schema/headers resueltos; control de destino enmascarado según `CT-006`; ningún efecto observable no autorizado en el contrato.
- PostgreSQL 14 real y rol `barberia_app`: migración vacía/actualización con datos, RLS con dos tenants, FK/checks, índice activo único, código vencido/usado, intentos atómicos, reenvío concurrente y dos consumos simultáneos con un solo ganador.
- Integración de credencial/sesiones: contraseña anterior deja de autenticar, nueva autentica, todas las sesiones anteriores responden `401`, otra barbería no se modifica.
- HTTP/contrato: JSON inválido/campo desconocido/cuerpo excesivo, solicitudes válidas, errores seguros, código incorrecto/vencido/agotado/usado, autorización de reset inválida/reutilizada y política incumplida.
- Adaptador: doble de prueba confirma ambos canales y contenido mínimo; timeout, fallo de uno/ambos y reintento según decisión. Ninguna prueba llama al proveedor real.
- Auditoría automática/manual de logs, fixtures, SQL y respuestas: cero código, contraseña, token, teléfono/correo completo o body de proveedor.
- E2E backend/sistema del recorrido completo con proveedor interceptado. El E2E visual de tres pasos pertenece a `HU-011`.

## Criterios de aceptación y cierre

Entrega `Criterio | Estado | Prueba o evidencia` para `CA-008-01` a `CA-008-08`.

- `CA-008-01` exige envío real a ambos adaptadores aprobados para cuenta válida y contrato no enumerable para inexistente; un fake que solo devuelve éxito sin afirmar llamadas no basta.
- `CA-008-02`/`03` se prueban con reloj y concurrencia, no con espera real ni `sleep`.
- `CA-008-04` incluye demostrar que un dump de la fila no contiene el valor enviado ni una representación trivial de invertir.
- `CA-008-05` requiere todas las sesiones y cambio de credencial en frontera consistente, probado con más de una sesión.
- `CA-008-06` se cierra solo después de resolver `CT-006`; nunca expongas el dato completo.
- `CA-008-07` requiere límite propio de reenvío y ausencia de dos códigos activos.
- `CA-008-08` usa exclusivamente la política aprobada por `DP-SEG-11`, documentada antes de implementar.

## Documentación y trazabilidad

- Actualiza OpenAPI/CHANGELOG y cliente generado; `apps/api/README.md`, `database/README.md`, diccionario/diagrama si existen, matriz e historial.
- Versiona migración y `atlas.sum`; documenta bloqueo, datos existentes, roll-forward y configuración del proveedor sin secretos.
- Registra evidencia de dependencia/proveedor y la forma de operar fallos sin exponer destinatarios.
- Actualiza este prompt/catálogo con estado, rama y PR reales. `HU-011` no pasa a ejecutable solo porque el cliente se regenere.

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

# apps/web: solo contrato generado/consumidores afectados, sin HU-011 visual
pnpm run generate:api
pnpm run format
pnpm run lint
pnpm run typecheck
pnpm run test:unit
pnpm run build

# raíz
git diff --check
```

Ejecuta cada bloque desde su directorio. Prueba la migración desde vacío y desde la versión anterior con datos y roles reales. `gofmt -l .` debe quedar sin salida. Añade el E2E backend/sistema aprobado al comando/suite real del repositorio; no inventes un comando que no exista.

## Git y PR

- Rama: `feat/58-hu008-recuperacion-acceso`.
- Commit/PR: `feat(auth): implementa HU-008 recuperación de acceso`.
- Usa `Closes #58` solo con `HU-007` integrada, todos los bloqueos resueltos y `CA-008-01`–`CA-008-08` demostradas; en otro caso usa `Refs #58` y conserva el PR borrador.
- No mezcles `HU-011`, no hagas push directo/force push/merge de `main` y no omitas checks o revisión de secretos/datos personales.

