---
prompt_id: "PROMPT-HU-023-v1"
version: "1.1"
kind: "hu"
status: "draft"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-023"
related_hu:
  - "HU-021"
  - "HU-022"
  - "HU-024"
issue: "76"
issue_url: "https://github.com/bcaceres19/barberia/issues/76"
suggested_issue_title: "feat(catalog): implementar HU-023 asignación de servicios a barberos"
branch: "feat/76-hu023-servicios-barberos"
pr: null
pr_url: null
depends_on:
  - "HU-021 integrada en main (cumplido)"
  - "HU-022 integrada en main (pendiente)"
  - "DP-SER-02 resuelta como DEC-068 (cumplido)"
rules:
  - "RN-SER-03"
  - "RN-SER-04"
  - "RN-TEN-01"
  - "RN-DAT-02"
decisions:
  - "DEC-004"
  - "DEC-019"
  - "DEC-024"
  - "DEC-033"
  - "DEC-034"
  - "DEC-035"
  - "DEC-036"
  - "DEC-037"
  - "DEC-038"
  - "DEC-039"
  - "DEC-040"
  - "DEC-068"
acceptance_criteria:
  - "CA-023-01"
  - "CA-023-02"
  - "CA-023-03"
  - "CA-023-04"
  - "CA-023-05"
  - "CA-023-06"
  - "CA-023-07"
  - "CA-023-08"
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
  - "docs/10-backlog/prompts/hu/hu-021-registro-listado-barberos.md"
  - "docs/10-backlog/prompts/hu/hu-022-catalogo-servicios.md"
  - "database/README.md"
  - "database/modelo-fisico-referencia.sql"
  - "database/migrations"
  - "api/openapi/openapi.yaml"
  - "api/openapi/paths/staff.yaml"
  - "api/openapi/paths/catalog.yaml"
  - "apps/api/internal/modules/staff"
  - "apps/api/internal/modules/catalog"
  - "apps/web/src/modules/staff"
  - "apps/web/src/modules/catalog"
created_at: "2026-08-24"
updated_at: "2026-08-24"
supersedes: null
superseded_by: null
---

# Implementar HU-023: asignación de servicios a barberos

## Instrucción para Claude o Codex

`DP-SER-02` está resuelta como `DEC-068` y el issue real [#76](https://github.com/bcaceres19/barberia/issues/76) existe con `CA-023-01`–`CA-023-08`. Este prompt permanece en `draft` únicamente porque `HU-022` todavía no está integrada en `main`. No implementes hasta que lo esté; entonces pasa este prompt a `in_progress` e implementa únicamente la relación entre barberos y servicios existentes. No cambies el catálogo, el ciclo de vida, horarios, disponibilidad ni citas.

## Objetivo

Que un barbero autenticado configure qué servicios presta cada integrante de su barbería usando una asociación tenant-aware sin atributos duplicados. Un servicio activo debe conservar al menos un barbero asignado; retirar la última asignación activa se rechaza, conforme a `DEC-068`.

## Preflight obligatorio

1. Comprueba árbol limpio, actualiza `main` por fast-forward y verifica que HU-021/HU-022 estén integradas con CI verde.
2. Ejecuta Graphify sobre `staff`, `catalog`, `barber`, `service`, composición privada e `InTenantTx`; identifica el dueño de cada operación y evita ciclos.
3. Lee completamente cada `source_docs`, incluidas las APIs públicas y pruebas reales de HU-021/HU-022.
4. Confirma en `docs/00-control/registro-decisiones.md` que `DEC-068` está propagada. Si aparece una contradicción, registra bloqueo y detente.
5. Crea `feat/76-hu023-servicios-barberos` desde `main` actualizada y cambia este prompt a `in_progress` al iniciar.

## Alcance incluido

- Tabla de asociación pura `barber_service` con claves compuestas tenant-aware y sin datos del barbero o servicio.
- Consulta de servicios asignados a un barbero, asignación repetible sin duplicados y desasignación según `DP-SER-02`.
- Contrato OpenAPI privado y casos de uso en el módulo dueño, con colaboración explícita entre `catalog` y `staff` mediante APIs/puertos pequeños.
- Interfaz privada para gestionar asignaciones sin importar archivos internos de otros módulos.
- Migración Atlas mínima, RLS, grants `SELECT`/`INSERT`/`DELETE` solo sobre la asociación cuando la decisión lo permita.
- Pruebas con un y cuatro barberos, servicios compartidos y dos tenants reales.

## Fuera de alcance

- Crear, editar, desactivar o reactivar `service`; renombrar o borrar `barber`.
- Precio o duración por barbero, orden, capacidad, comisión o credenciales.
- Horarios, bloqueos, disponibilidad, selección pública, citas, snapshots, historial o notificaciones.
- Mutaciones directas de tablas ajenas desde el núcleo de otro módulo o una página que concentre internals de `staff` y `catalog`.

## Estado existente que debe conservarse

- HU-021 es dueña de `barber`; HU-022 será dueña de `service`. La nueva relación no mueve esas responsabilidades.
- El modelo de referencia B.4 propone una tabla de asociación en 2FN con FK compuestas. Revísalo contra la decisión final; no lo copies como migración aplicada.
- `auth.Principal` y `InTenantTx` son autoridades de tenant. Los identificadores de request nunca autorizan `barbershop_id`.
- El frontend compone rutas/nav desde APIs públicas mínimas. Si un módulo necesita una capacidad de otro, expón una intención estable o coordina en `app`; no importes páginas, clientes o modelos internos.
- No existe `appointment` en la cadena migrada actual. No simules citas ni uses su ausencia para decidir `DP-SER-02`.

## Trabajo requerido

### 1. Contrato HTTP

1. Diseña contract-first operaciones privadas para listar, asignar y desasignar. Prefiere una relación identificable (`barberId` + `serviceId`) y semántica HTTP naturalmente repetible; documenta cualquier conflicto real.
2. Declara `SessionCookie`, schemas cerrados, UUID, ejemplos ficticios, `x-business-rules` y `x-decisions`.
3. Recurso ajeno e inexistente deben ser indistinguibles. No expongas si el barbero o servicio existe en otro tenant.
4. El contrato no incluye nombre/duración/precio duplicados en la escritura ni campos futuros de horario/disponibilidad.
5. Regenera el cliente desde el bundle solo después de lint y pruebas de contrato verdes.

### 2. Migración y seguridad

1. Crea con Atlas únicamente `barber_service` y sus índices/políticas inmediatos.
2. Usa PK/FK compuestas por `barbershop_id`; impide asociar extremos de tenants distintos aun con SQL directo.
3. Habilita/fuerza RLS; política administrativa a `barberia_owner`; permisos mínimos al rol real de aplicación.
4. No uses `ON DELETE CASCADE` como autorización para borrar `barber` o `service`: esos verbos siguen prohibidos por sus tablas dueñas. Justifica la acción referencial de la asociación.
5. Prueba base vacía y upgrade desde HU-022 con datos, ausencia de contexto, A/B, permisos y `atlas.sum`.

### 3. Backend Go

1. Define la operación en el módulo que posee la intención de catálogo. La existencia/pertenencia de un barbero se consulta mediante un puerto explícito o una operación pública estable de `staff`, no mediante acoplamiento del núcleo a su repositorio.
2. Implementa asignación atómica y repetible, desasignación conforme a `DP-SER-02` y lectura ordenada sin duplicados.
3. Cada consulta filtra tenant además de RLS; traduce ajeno/inexistente al mismo `NotFound`.
4. Prueba carreras de dos asignaciones iguales, última desasignación y recursos cruzados con PostgreSQL real.
5. No registres nombres, cuerpos, cookies ni listas completas; usa IDs opacos y `request_id`.

### 4. Frontend Vue

1. Integra la gestión en una ruta/pantalla coherente con las APIs públicas de `staff` y `catalog`; conserva carga diferida y un solo cascarón.
2. Muestra servicios disponibles y asignados con controles accesibles; el mismo componente funciona con uno y cuatro barberos.
3. Aplica exactamente el estado permitido por `DP-SER-02`, incluidos vacío, última asignación, advertencia o bloqueo si fueron aprobados.
4. Evita doble envío, conserva selección ante error recuperable y reemplaza estado solo después de respuesta del servidor.
5. No muestres horarios, disponibilidad, citas, precio por barbero ni controles de ciclo de vida.

## Pruebas y evidencia

- Dominio/servicio: asignación nueva/repetida, servicio compartido, desasignación repetida, última asignación y errores de puertos.
- PostgreSQL: PK/FK tenant-aware, RLS, grants, A/B, sin contexto, carrera y ausencia de atributos duplicados.
- HTTP/contrato: lista, asignar, desasignar, `401`, `404` uniforme, `409`/`422` solo si la decisión los exige y campos desconocidos.
- Arquitectura: pruebas que impidan imports internos cruzados y acceso del núcleo a Chi/pgx/tablas de otro dueño.
- Componentes/E2E: un barbero, cuatro, servicio compartido, retiro conforme a la decisión, recarga y aislamiento.
- Responsive/accesible: 320, 360, 768, 1280 px, zoom 200 %, teclado, foco, objetivos táctiles y axe-core.

## Documentación y trazabilidad

- Actualiza OpenAPI/CHANGELOG, README, inventarios de migraciones/pruebas, matriz e historial según lo implementado.
- Actualiza el modelo físico si la decisión aprobada difiere de B.4, explicando integridad y normalización.
- Actualiza prompt/catálogo con issue, rama, PR y estado reales; no declares B1 cerrado ni `HU-024` implementada.

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

Entrega `Criterio | Estado | Prueba o evidencia` para `CA-023-01`–`CA-023-08` y una lista de exclusiones inspeccionadas en OpenAPI, SQL, Go y Vue.

## Git y PR

- Rama: `feat/<issue>-hu023-servicios-barberos`.
- Commits/título: `feat(catalog): implementa HU-023 asignación de servicios a barberos`.
- `Closes #<issue>` solo si todos los criterios están verificados; en otro caso `Refs #<issue>`.
- No hagas push directo, force push, merge manual, DDL al arrancar ni cambios a migraciones aplicadas.
