---
prompt_id: "PROMPT-HU-022-v1"
version: "1.1"
kind: "hu"
status: "in_progress"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-022"
related_hu:
  - "HU-020"
  - "HU-021"
  - "HU-023"
  - "HU-024"
issue: "75"
issue_url: "https://github.com/bcaceres19/barberia/issues/75"
suggested_issue_title: "feat(catalog): implementar HU-022 catálogo básico de servicios"
branch: "feat/75-hu022-catalogo-servicios"
pr: null
pr_url: null
depends_on:
  - "HU-020 y HU-021 integradas en main (cumplido)"
  - "DP-SER-01 resuelta como DEC-067 (cumplido)"
rules:
  - "RN-SER-01"
  - "RN-SER-02"
  - "RN-SER-04"
  - "RN-TEN-01"
  - "RN-DAT-02"
  - "RN-IDE-01"
decisions:
  - "DEC-002"
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
  - "DEC-043"
acceptance_criteria:
  - "CA-022-01"
  - "CA-022-02"
  - "CA-022-03"
  - "CA-022-04"
  - "CA-022-05"
  - "CA-022-06"
  - "CA-022-07"
  - "CA-022-08"
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
  - "database/README.md"
  - "database/modelo-fisico-referencia.sql"
  - "database/migrations"
  - "api/openapi/openapi.yaml"
  - "api/openapi/paths/catalog.yaml"
  - "apps/api/README.md"
  - "apps/api/internal/modules/catalog/doc.go"
  - "apps/web/README.md"
  - "apps/web/src/modules/catalog/index.ts"
  - "apps/web/src/modules/staff"
created_at: "2026-08-24"
updated_at: "2026-08-24"
supersedes: null
superseded_by: null
---

# Implementar HU-022: catálogo básico de servicios

## Instrucción para Claude o Codex

`DP-SER-01` está resuelta como `DEC-067` y el issue real [#75](https://github.com/bcaceres19/barberia/issues/75) existe con `CA-022-01`–`CA-022-08`. Ambas guardas quedan satisfechas: implementa `HU-022` como corte vertical contract-first para listar, consultar, crear y editar el catálogo básico de servicios. No adelantes asignaciones a barberos, ciclo de activación, disponibilidad, reserva ni citas.

## Objetivo

Que un barbero autenticado gestione nombre, descripción, duración planificada y precio informativo de los servicios de su barbería desde una pantalla real. Moneda COP fija, sin servicios gratuitos (`price > 0`) y nombre único entre servicios activos de la misma barbería, exactamente conforme a `DEC-067`. Un cambio de catálogo nunca modifica ni aparenta modificar citas existentes.

## Preflight obligatorio

1. Comprueba árbol limpio, actualiza `main` por fast-forward y no sobrescribas cambios ajenos.
2. Verifica en GitHub que `HU-020` y `HU-021` están integradas y que sus checks continúan verdes.
3. Ejecuta Graphify sobre `catalog`, `service`, `InTenantTx`, idempotencia, composición de rutas privadas y cliente OpenAPI; confirma los límites con `staff`, `settings`, `app` y `shared`.
4. Lee completamente cada `source_docs`. En rutas de directorio revisa archivos públicos, pruebas y migraciones aplicadas relevantes.
5. Confirma en `docs/00-control/registro-decisiones.md` que `DEC-067` está propagada a reglas, historia, criterios y modelo de referencia. Si aparece una contradicción, registra el bloqueo y detente.
6. Crea `feat/75-hu022-catalogo-servicios` desde `main` actualizada.

## Alcance incluido

- Recurso tenant-aware `service` con los campos aprobados por HU-022 y `DP-SER-01`.
- Lista por cursor, lectura individual, alta idempotente y edición parcial cerrada a campos de catálogo.
- Contrato OpenAPI privado bajo `/api/v1/private`, módulo Go `catalog`, repositorio PostgreSQL y pantalla Vue “Servicios”.
- Migración Atlas que crea únicamente `service` y sus objetos inmediatos; RLS forzada, índices guiados por consultas y grants mínimos sin `DELETE`.
- Duraciones enteras configurables sin lista cerrada; dinero como `numeric` y moneda ISO según el estándar y la decisión de producto.
- Pruebas de dominio, contrato, PostgreSQL real con dos tenants, componentes, E2E, responsive y accesibilidad.

## Fuera de alcance

- `barber_service`, asignar o retirar servicios de barberos (`HU-023`).
- Desactivar, reactivar o borrar servicios (`HU-024`).
- Citas, snapshots, actualización masiva, disponibilidad, horarios, enlace público, historial o notificaciones.
- Precios por profesional, impuestos, promociones, pagos, inventario, paquetes o comisiones.
- Pinia, biblioteca de tablas/formularios, dependencia nueva o estilos fuera del sistema visual sin justificación independiente.

## Estado existente que debe conservarse

- `internal/modules/catalog/doc.go`, `modules/catalog/index.ts` y `api/openapi/paths/catalog.yaml` son stubs intencionales; conviértelos en módulos reales sin crear otra capacidad paralela.
- `auth.Principal` y `database.DB.InTenantTx` siguen siendo las únicas autoridades de tenant. Cada consulta filtra `barbershop_id` además de RLS.
- La creación reutiliza el coordinador idempotente de HU-004 dentro de la misma transacción que `INSERT` y `Complete`.
- HU-020/HU-021 ya entregaron composición de rutas/navegación privadas, cliente tipado, patrones de formulario/lista y pruebas con dos tenants. Reutiliza sus APIs públicas, no sus archivos internos.
- La sección B.3 de `modelo-fisico-referencia.sql` es insumo revisable, no migración aplicada ni decisión para los puntos abiertos de `DP-SER-01`.
- El cliente TypeScript deriva del bundle OpenAPI y nunca se edita manualmente.

## Trabajo requerido

### 1. Contrato HTTP primero

1. Define colección privada `GET/POST /private/services` y recurso `GET/PATCH /private/services/{serviceId}` con tag estable `Catalog`.
2. Declara `SessionCookie`, JSON `camelCase`, schemas cerrados, ejemplos ficticios, `x-business-rules` y `x-decisions`.
3. `POST` exige `Idempotency-Key`, responde `201 + Location` y documenta replay, conflicto de huella y operación concurrente conforme al protocolo real.
4. `PATCH` rechaza cuerpo vacío y campos desconocidos; no acepta `isActive`, asignaciones, citas ni alcance de propagación.
5. La lista usa cursor opaco y orden estable. No inventes un máximo de servicios por barbería.
6. Documenta errores RFC 9457 reales (`400`, `401`, `404`, `409`, `422`, `500`) y regenera el cliente solo tras lint y bundle verdes.

### 2. Migración y datos

1. Genera con Atlas una migración nueva e inmutable para `service`, sin copiar `barber_service` ni tablas posteriores.
2. Implementa UUID, FK `barbershop`, unique tenant-aware, timestamps, checks y precisión monetaria coherentes con HU-022, el estándar y la resolución de `DP-SER-01`.
3. No agregues `DELETE`; el estado inicial activo prepara HU-024 sin exponer su transición.
4. Habilita/fuerza RLS y concede solo `SELECT`, `INSERT`, `UPDATE` a `barberia_app`; política administrativa a `barberia_owner`.
5. Valida base vacía y upgrade desde HU-021 con datos, `atlas.sum`, bloqueo, roll-forward y PostgreSQL 14 real.
6. Añade testdata sintética y pruebas SQL propias con dos barberías; no cites el modelo físico como ejecución.

### 3. Backend Go

1. Implementa el núcleo `catalog` con dominio, errores, puertos y servicios independientes de Chi, `net/http`, pgx y `database.DB`.
2. Valida/normaliza cada campo una sola vez en la autoridad adecuada; usa Unicode por caracteres y dinero decimal, nunca `float`.
3. Implementa repositorio `postgres` dentro de `InTenantTx`; lista por cursor y `404` idéntico para ajeno/inexistente.
4. Ejecuta alta idempotente completa en una transacción y prueba replay byte a byte, conflicto, concurrencia y misma clave literal en dos tenants.
5. Registra rutas solo en el subrouter privado y traduce errores con las utilidades existentes. Ningún log contiene nombre/descripcion del servicio ni body.

### 4. Frontend Vue

1. Convierte `modules/catalog` en módulo real con API pública mínima, ruta diferida, navegación, cliente por intención, modelos discriminados, validación, página y pruebas.
2. Compón `/panel/servicios` dentro del cascarón existente; no repitas guard ni importes internos de `auth`, `settings` o `staff`.
3. Implementa lista/paginación, vacío, carga y error recuperable; alta y edición accesibles con datos conservados ante fallo.
4. Genera una clave idempotente por intento lógico y reutilízala en reintentos del mismo alta; bloquear el botón no sustituye este protocolo.
5. Usa tokens/componentes aprobados; no muestres asignaciones, activación, disponibilidad ni citas.

## Pruebas y evidencia

- `CA-022-01`–`CA-022-08` con una prueba identificable por criterio.
- Dominio: límites, Unicode, duración 25/30/45/90, dinero y todos los casos resueltos por `DP-SER-01`.
- PostgreSQL: esquema mínimo, precisión, constraints, RLS, grants, A/B, sin contexto, upgrade y ausencia de `DELETE`.
- HTTP/contrato: forma exacta, campos desconocidos, paginación, idempotencia, `401`, `404`, `409`, `422`, `500`.
- Privacidad: logs de éxito/fallo sin nombre, descripción, cuerpo, cookie ni datos de contacto.
- Componente/E2E: crear → listar → editar → recargar, una y varias filas, red fallida, doble toque y aislamiento.
- Evidencia inspeccionada en 320, 360, 768 y 1280 px, zoom 200 %, teclado, foco y axe-core.

## Documentación y trazabilidad

- Actualiza OpenAPI/CHANGELOG, README de API/web/base de datos, inventarios de migraciones/pruebas, matriz e historial según artefactos reales.
- Si el DDL aprobado difiere justificadamente de B.3, actualiza el modelo físico de referencia y documenta la razón.
- Actualiza este prompt y el catálogo al iniciar, abrir PR, ejecutar, bloquear o sustituir; no declares B1 terminado.

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

Entrega `Criterio | Estado | Prueba o evidencia` para los ocho criterios y una lista de exclusiones inspeccionadas en OpenAPI, SQL, Go y Vue. No declares cumplido aquello que no esté probado.

## Git y PR

- Rama esperada: `feat/<issue>-hu022-catalogo-servicios`.
- Commits y título: `feat(catalog): implementa HU-022 catálogo básico de servicios` o equivalente Conventional Commits.
- Usa `Closes #<issue>` solo si los ocho criterios y exclusiones están verificados; de lo contrario `Refs #<issue>`.
- No hagas push directo, force push, merge manual, DDL al arrancar ni cambios a migraciones aplicadas.
