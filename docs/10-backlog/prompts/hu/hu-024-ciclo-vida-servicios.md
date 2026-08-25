---
prompt_id: "PROMPT-HU-024-v1"
version: "1.2"
kind: "hu"
status: "executed"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-024"
related_hu:
  - "HU-022"
  - "HU-023"
issue: "77"
issue_url: "https://github.com/bcaceres19/barberia/issues/77"
suggested_issue_title: "feat(catalog): implementar HU-024 desactivación y reactivación de servicios"
branch: "feat/77-hu024-ciclo-vida-servicios"
pr: 84
pr_url: "https://github.com/bcaceres19/barberia/pull/84"
depends_on:
  - "HU-022 y HU-023 integradas en main (cumplido)"
  - "DP-SER-03 resuelta como DEC-069 (cumplido)"
rules:
  - "RN-SER-03"
  - "RN-SER-04"
  - "RN-TEN-01"
  - "RN-IDE-01"
  - "RN-DAT-02"
decisions:
  - "DEC-003"
  - "DEC-004"
  - "DEC-024"
  - "DEC-033"
  - "DEC-034"
  - "DEC-035"
  - "DEC-036"
  - "DEC-037"
  - "DEC-038"
  - "DEC-039"
  - "DEC-040"
  - "DEC-043"
  - "DEC-069"
acceptance_criteria:
  - "CA-024-01"
  - "CA-024-02"
  - "CA-024-03"
  - "CA-024-04"
  - "CA-024-05"
  - "CA-024-06"
  - "CA-024-07"
  - "CA-024-08"
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
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/04-arquitectura/backend-go.md"
  - "docs/04-arquitectura/frontend.md"
  - "docs/05-backend/base-datos.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/06-api/estandar-openapi.md"
  - "docs/10-backlog/plan-bloques.md"
  - "docs/10-backlog/prompts/hu/hu-022-catalogo-servicios.md"
  - "docs/10-backlog/prompts/hu/hu-023-asignacion-servicios-barberos.md"
  - "database/README.md"
  - "database/modelo-fisico-referencia.sql"
  - "database/migrations"
  - "api/openapi/openapi.yaml"
  - "api/openapi/paths/catalog.yaml"
  - "apps/api/internal/modules/catalog"
  - "apps/web/src/modules/catalog"
created_at: "2026-08-24"
updated_at: "2026-08-25"
supersedes: null
superseded_by: null
---

# Implementar HU-024: desactivación y reactivación de servicios

## Instrucción para Claude o Codex

`DP-SER-03` está resuelta como `DEC-069` y el issue real [#77](https://github.com/bcaceres19/barberia/issues/77) existe con `CA-024-01`–`CA-024-08`. `HU-022` y `HU-023` ya están integradas en `main`: este prompt pasó a `executed` y produjo la implementación del ciclo de vida de `service` sin borrado físico, contra la rama `feat/77-hu024-ciclo-vida-servicios` y el PR #84 integrado. La cadena migrada todavía no contiene `appointment`; el impacto de citas futuras (`CA-024-01`/`CA-024-04`) es literal y honestamente `0` en B1 (`catalog.currentDeactivationImpact`, `DEC-069`), sin bloqueo optimista, sin tabla `appointment` parcial ni adaptador que simule una consulta contra algo que no existe todavía.

## Objetivo

Que el barbero desactive o reactive un servicio con una confirmación honesta y segura, conservando el recurso y sin alterar citas automáticamente. La división B1/B3 y la ausencia de protección de concurrencia sobre el conteo en B1 siguen literalmente `DEC-069`.

## Preflight obligatorio

1. Comprueba árbol limpio y `main` actualizada por fast-forward; verifica HU-022/HU-023 integradas y CI verde.
2. Ejecuta Graphify sobre `catalog`, `service`, idempotencia, rutas privadas y cualquier puerto aprobado para impacto de citas. No accedas directamente a tablas de una capacidad futura.
3. Lee completamente cada `source_docs` y confirma el estado real de migraciones, contrato y módulos.
4. Confirma en `docs/00-control/registro-decisiones.md` que `DEC-069` está propagada. Si aparece una contradicción, registra bloqueo y detente.
5. Crea `feat/77-hu024-ciclo-vida-servicios` desde `main` actualizada y cambia este prompt a `in_progress` antes de implementar.

## Alcance incluido

- Transiciones explícitas activo → inactivo e inactivo → activo sobre el mismo `service`.
- Marca temporal e invariantes de forma coherentes; nunca borrado físico ni recreación.
- Operación real de impacto/previsualización y confirmación solo en la forma aprobada por `DP-SER-03`.
- Desactivación/reactivación idempotentes, tenant-aware y protegidas frente a repetición/concurrencia.
- Interfaz accesible con advertencia, confirmación, conflicto, éxito y error recuperable.
- Pruebas PostgreSQL real, HTTP/contrato, dominio, componentes y E2E sin simular capacidades no aprobadas.

## Fuera de alcance

- Cancelar, reprogramar, editar o notificar citas; selección de citas afectadas si la decisión la asigna a B3.
- `DELETE` físico, purga, cambio de identificador o pérdida de historial.
- Cambiar nombre, descripción, duración, precio, moneda o asignaciones en la misma operación.
- Reserva pública, disponibilidad y ocultamiento en superficies que todavía no existen; sus historias futuras consumen el estado real.
- Crear una tabla `appointment` parcial, un adaptador que siempre devuelva cero o datos falsos de prueba en producción.

## Estado existente que debe conservarse

- HU-022 crea `service`; HU-023 crea `barber_service`. Esta historia evoluciona únicamente el ciclo de vida del servicio.
- `RN-SER-03` y `DEC-003` prohíben borrar y decidir automáticamente sobre citas. `RN-SER-04`/`DEC-004` prohíben propagar cambios de catálogo en silencio.
- El modelo de referencia propone `is_active` + `deactivated_at` y ausencia de `DELETE`; es insumo, no respuesta a `DP-SER-03`.
- `auth.Principal`, `InTenantTx` e idempotencia de HU-004 siguen siendo las fronteras obligatorias.
- El cliente OpenAPI es generado; el frontend usa el módulo `catalog` y el cascarón privado existentes.

## Trabajo requerido

### 1. Contrato y semántica

1. Define contract-first la consulta de impacto y los comandos de desactivar/reactivar según la decisión aprobada. No uses un `PATCH isActive` genérico que omita la confirmación exigida.
2. Declara `SessionCookie`, schemas cerrados, idempotencia, `x-business-rules`, `x-decisions` y RFC 9457.
3. Modela explícitamente impacto vigente/obsoleto, reintento y conflictos de estado si `DP-SER-03` los exige.
4. Recurso ajeno e inexistente producen el mismo `404`. Repetir la misma intención reproduce el mismo resultado; otra intención con la misma clave da conflicto.
5. No documentes cancelación de citas ni un conteo que la implementación real no pueda calcular.

### 2. Datos y concurrencia

1. Si HU-022 ya creó columnas de ciclo de vida, no dupliques ni reescribas su migración; añade solo el roll-forward necesario si la decisión cambia la forma.
2. Conserva el `CHECK` que relaciona estado y marca temporal, RLS forzada y ausencia de `DELETE` para el rol de aplicación.
3. Implementa cambio de estado y registro idempotente en una sola transacción corta.
4. Aplica el mecanismo aprobado contra TOCTOU entre impacto y confirmación; prueba una carrera real con dos conexiones, sin `sleep` como sincronización.
5. Valida upgrade con servicios activos/inactivos existentes, bloqueo, roll-forward, `atlas.sum` y PostgreSQL 14.

### 3. Backend Go

1. Añade comandos/valores al núcleo `catalog`; ninguna regla vive solo en el handler o SQL.
2. El puerto de impacto pertenece a la capacidad dueña de citas cuando exista. `catalog` consume un contrato pequeño; no consulta tablas ajenas desde el núcleo.
3. Traduce transición inválida, impacto obsoleto, idempotencia concurrente y `NotFound` a errores estables del contrato.
4. No mantengas una transacción abierta durante llamadas de red ni ejecutes efectos de citas/notificación.
5. Logs solo con IDs opacos, transición y `request_id`; nunca nombre del servicio, cuerpo, cookies o datos de clientes.

### 4. Frontend Vue

1. Añade acciones de desactivar/reactivar dentro de “Servicios” sin mezclar edición de catálogo.
2. La advertencia muestra únicamente impacto respaldado por el servidor y exige confirmación accesible; cancelar el diálogo no muta nada.
3. Gestiona estado obsoleto/conflicto con recarga de impacto, no con éxito optimista.
4. Deshabilita doble toque y conserva una clave por intento lógico; actualiza el recurso solo tras confirmación del servidor.
5. Estado inactivo se comunica con texto/semántica además de color; foco, teclado y lector de pantalla deben entender el flujo.

## Pruebas y evidencia

- Dominio: transiciones, repetición, clave conflictiva, impacto obsoleto y errores de puerto.
- PostgreSQL: invariantes temporales, RLS, grants sin `DELETE`, A/B, misma fila preservada y carrera real.
- HTTP/contrato: impacto, desactivar, reactivar, `401`, `404`, `409`, `422`, `500` y schemas cerrados.
- Privacidad: ninguna evidencia contiene nombres de clientes, citas, teléfonos, correos, tokens o cuerpos.
- Componentes/E2E: abrir advertencia, cancelar, confirmar, conflicto/reintento, recargar, reactivar y aislamiento.
- Responsive/accesible: 320, 360, 768, 1280 px, zoom 200 %, teclado, foco, objetivos táctiles, axe-core y contraste.

## Documentación y trazabilidad

- Actualiza contrato/CHANGELOG, README, migraciones/pruebas, matriz, historial y plan de bloques con la división real B1/B3.
- Actualiza el modelo de referencia si la resolución cambia columnas, índices o flujo; no conviertas una propuesta en decisión silenciosa.
- Actualiza prompt/catálogo con issue, rama, PR y estado. Solo declara B1 cerrado si todos sus criterios de salida están verificados.

## Verificación final

```text
pnpm run openapi:lint
pnpm run openapi:bundle
pnpm run db:validate
pnpm run db:test
cd apps/api && gofmt -w . && go vet ./... && go test -race ./... && go build ./...
cd apps/web && pnpm run lint && pnpm run typecheck && pnpm run test:coverage && pnpm run build && pnpm run test:e2e
git diff --check
graphify update .
```

Entrega `Criterio | Estado | Prueba o evidencia` para `CA-024-01`–`CA-024-08`, la decisión aplicada a cada punto de `DP-SER-03` y una lista de exclusiones inspeccionadas.

## Git y PR

- Rama: `feat/<issue>-hu024-ciclo-vida-servicios`.
- Commits/título: `feat(catalog): implementa HU-024 desactivación y reactivación de servicios`.
- `Closes #<issue>` solo si los ocho criterios y la división B1/B3 están probados; de lo contrario `Refs #<issue>`.
- No hagas push directo, force push, merge manual, DDL al arrancar ni cambios a migraciones aplicadas.
