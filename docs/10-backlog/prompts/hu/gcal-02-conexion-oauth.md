---
prompt_id: "PROMPT-FEAT-GCAL-02-CONEXION-OAUTH-v1"
version: "1.0"
kind: "hu"
status: "draft"
target_agents:
  - "any"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu: []
issue: "pending"
issue_url: null
suggested_issue_title: "feat(integraciones): conectar y desconectar Google Calendar por barbero"
branch: null
pr: null
pr_url: null
depends_on:
  - "PROMPT-FEAT-GCAL-01-v1 integrado (vínculo barbero–usuario)"
  - "PR del issue #284 integrado (DEC-102)"
rules:
  - "RN-TEN-01"
decisions:
  - "DEC-099"
  - "DEC-100"
  - "DEC-101"
  - "DEC-102"
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/06-api/estandar-openapi.md"
  - "docs/04-arquitectura/backend-go.md"
  - "docs/04-arquitectura/stack-despliegue-operacion.md"
  - "apps/api/internal/platform/config"
created_at: "2026-09-26"
updated_at: "2026-09-26"
supersedes: null
superseded_by: null
---

# Conexión OAuth de Google Calendar por barbero

## Instrucción para el agente

Implementa únicamente la preocupación descrita en este archivo. Trabaja de forma autónoma hasta producir la entrega definida, sin ampliar el alcance ni inventar decisiones. Es una entrega de la orquestación [PROMPT-ORCH-GCAL-BARBERO-v1](../orchestration/google-calendar-barbero.md); léela para conocer el orden y el principio de autoridad (NAVA manda, Google es una vista).

## Objetivo

Un barbero con usuario vinculado conecta su propia cuenta de Google, consulta el estado y la desconecta; el refresh token queda cifrado y nunca sale del backend.

## Preflight obligatorio

1. Comprueba que el árbol esté limpio y no sobrescribas cambios ajenos; trabaja en un worktree si hay cambios sin relación.
2. Actualiza `main` mediante fast-forward y verifica que cada dependencia de los metadatos esté integrada.
3. Lee completamente `AGENTS.md`, `CLAUDE.md` y cada archivo de `source_docs`, y las `DEC-*` enlazadas.
4. Comprueba `dudas-pendientes.md` y `contradicciones.md`. Si falta una decisión, registra el bloqueo antes de cambiar código.
5. Este prompt nace con `issue: pending` y en `draft`. Antes de ejecutarlo crea el issue real con criterios verificables (título sugerido en los metadatos), actualiza este archivo y el índice, y solo entonces pásalo a `ready`.
6. Si esta entrega introduce o toca una historia, redáctala en `docs/02-requisitos/historias-usuario.md` con su número real y criterios `CA-*` antes de codificar; no inventes números.
7. Crea la rama `<tipo>/<issue>-<descripcion>` desde `main` actualizada.

## Alcance incluido

- OpenAPI primero: estado, inicio de conexión (URL de autorización), callback y desconexión, con errores del estándar.
- Migración Atlas y RLS de `google_calendar_connection` (y del estado OAuth de un solo uso), en 3FN, con estados `connected`, `reauth_required`, `error`, `disconnected` y una sola conexión activa por barbero.
- Módulo dueño `apps/api/internal/modules/googlecalendar` con puertos: cliente Google (adaptador aislado), almacén de credenciales, cifrado AES-256-GCM con `key_id`.
- PKCE, `state` de un solo uso ligado a barbería, barbero y sesión; refresh del access token bajo demanda; revocación y borrado de credenciales al desconectar.
- Configuración `GOOGLE_CALENDAR_CLIENT_ID`, `GOOGLE_CALENDAR_CLIENT_SECRET`, `GOOGLE_CALENDAR_REDIRECT_URI` y `GOOGLE_CALENDAR_TOKEN_ENCRYPTION_KEY`; integración desactivada cuando falta alguna, sin fallar el arranque del API.
- Guía paso a paso de Google Cloud (proyecto, API, pantalla de consentimiento, cliente OAuth, redirect URI por ambiente, alcances mínimos, usuarios de prueba, verificación para producción).

## Fuera de alcance

- Crear, actualizar o eliminar eventos, trabajos y worker (`PROMPT-FEAT-GCAL-03-v1`).
- Leer cambios de Google, webhook, `watch` o `syncToken` (descartados por `DEC-099`).
- UI (`PROMPT-FEAT-GCAL-04-v1`) y cuentas de clientes o administradores.

## Estado existente que debe conservarse

- Los servicios de dominio siguen independientes de Chi y pgx; el SDK de Google vive solo en el adaptador.
- Roles `barberia_app` y `barberia_worker` separados (`DEC-040`); ningún DDL al arrancar.

## Trabajo requerido

1. Justifica la dependencia oficial de Google en el PR y ejecuta `govulncheck`.
2. Actualiza OpenAPI, ejecuta lint y genera el cliente.
3. Crea la migración y las pruebas de RLS con dos tenants.
4. Implementa el módulo, los handlers y la configuración.
5. Verifica que ninguna respuesta, log ni error contiene tokens, `client_secret`, `state` ni el código OAuth.

## Pruebas y evidencia

- Cifrado: ida y vuelta, clave incorrecta, rotación por `key_id`; nada en claro en la base de datos.
- OAuth con un servidor Google falso: éxito, `state` reutilizado o expirado, usuario sin barbero vinculado, permisos denegados, token revocado (`reauth_required`).
- Aislamiento: el barbero de la barbería A no lee ni desconecta la conexión de B.
- Refresh automático de un access token expirado.

## Documentación y trazabilidad

- Actualiza matriz de trazabilidad, historial, contrato OpenAPI, diccionario y diagrama de datos que realmente resulten afectados.
- Documenta las variables de entorno en `apps/api/README.md` y en el documento de despliegue; guía de Google Cloud en `docs/`.
- Actualiza los metadatos y el índice de este catálogo con issue, rama, PR y estado reales.

## Verificación final

```text
cd apps/api && go vet ./... && go test ./... && go test -race ./... && govulncheck ./...
pnpm run openapi:lint (y el bundle/cliente tipado definidos por el repositorio)
atlas migrate validate y pruebas contra PostgreSQL real con al menos dos tenants
tools/ai/validate-agent-system.sh --strict
```

Entrega una tabla `Criterio | Estado | Prueba o evidencia`. No declares cumplido aquello que no esté probado.

## Git y PR

- Commit/PR con Conventional Commits, sin secretos ni datos personales.
- `Closes #<issue>` solo si se cubre el issue completo; de lo contrario `Refs #<issue>`.
- No hagas push directo, force push ni merge de `main`; integra por PR y squash tras CI verde.
