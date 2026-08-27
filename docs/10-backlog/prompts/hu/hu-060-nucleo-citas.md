---
prompt_id: "PROMPT-HU-060-v1"
version: "1.0"
kind: "hu"
status: "executed"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-060"
related_hu:
  - "HU-040"
  - "HU-041"
  - "HU-042"
  - "HU-061"
issue: "104"
issue_url: "https://github.com/bcaceres19/barberia/issues/104"
suggested_issue_title: "feat(booking): implementar HU-060 núcleo de citas sin cruces"
branch: "feat/104-hu060-nucleo-citas"
pr: null
pr_url: null
depends_on:
  - "HU-040, HU-041 y HU-042 integradas en main"
  - "Criterio de salida de B2 verificado; reconciliar seguimientos abiertos #90, #95, #98 y #100 que afecten B3"
  - "Issue real propio creado y enlazado antes de pasar a ready o ejecutar"
rules:
  - "RN-CON-01"
  - "RN-CON-03"
  - "RN-DIS-05"
  - "RN-DIS-07"
  - "RN-HIS-01"
  - "RN-HIS-02"
  - "RN-RES-02"
  - "RN-RES-03"
  - "RN-TEN-01"
decisions:
  - "DEC-002"
  - "DEC-004"
  - "DEC-007"
  - "DEC-014"
  - "DEC-016"
  - "DEC-019"
  - "DEC-020"
  - "DEC-024"
  - "DEC-035"
  - "DEC-036"
  - "DEC-037"
  - "DEC-038"
  - "DEC-039"
  - "DEC-040"
  - "DEC-041"
  - "DEC-045"
  - "DEC-046"
  - "DEC-070"
acceptance_criteria:
  - "CA-060-01"
  - "CA-060-02"
  - "CA-060-03"
  - "CA-060-04"
  - "CA-060-05"
  - "CA-060-06"
  - "CA-060-07"
  - "CA-060-08"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/00-control/glosario.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/estados-citas.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/04-arquitectura/backend-go.md"
  - "docs/05-backend/base-datos.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/06-api/estandar-openapi.md"
  - "docs/10-backlog/plan-bloques.md"
  - "database/README.md"
  - "database/modelo-fisico-referencia.sql"
  - "database/migrations"
  - "database/tests"
  - "apps/api/README.md"
  - "apps/api/internal/modules/booking"
  - "apps/api/internal/platform/database"
created_at: "2026-08-26"
updated_at: "2026-08-27"
supersedes: null
superseded_by: null
---

# Implementar HU-060: núcleo persistente de citas sin cruces

## Instrucción para el agente

Implementa únicamente la base persistente y la primitiva transaccional interna de `HU-060`. Este prompt permanece en `draft` mientras `issue: pending`; no cambies código, contrato ni migraciones hasta crear o recibir un issue real, actualizar los metadatos y comprobar el criterio de salida de B2.

`database/modelo-fisico-referencia.sql` sección D es un insumo revisable, no una migración autorizada. No lo copies sin reconciliarlo con las quince migraciones aplicadas, `DEC-040`, `DEC-045`/`DEC-046`, la prohibición de cascadas y los estándares vigentes.

## Objetivo

Que PostgreSQL 14 haga imposible persistir citas cruzadas que ocupan agenda para el mismo barbero y que el módulo `booking` disponga de una operación interna capaz de guardar cliente, cita `confirmed` e historial inicial en una sola transacción, sin endpoint ni pantalla todavía.

## Preflight obligatorio

1. Comprueba árbol limpio y no sobrescribas cambios ajenos; actualiza `main` solo mediante fast-forward.
2. Confirma que `HU-040`–`HU-042` están integradas y revisa si los issues `#90`, `#95`, `#98` o `#100` contienen un cambio que afecte el contrato interno de B3.
3. Crea o localiza el issue real de `HU-060`; actualiza `issue`, `issue_url`, `status` y, al iniciar, `branch` en este archivo y en el índice.
4. Crea `feat/<issue>-hu060-nucleo-citas` desde `main` actualizada. No reutilices una rama de B2.
5. Ejecuta Graphify sobre `booking`, `schedule`, `catalog`, `staff`, transacciones tenant-aware, idempotencia y migraciones cuando exista `graphify-out/graph.json`.
6. Lee completamente todos los `source_docs`. Comprueba dudas y contradicciones; no uses `DP-CIT-01` para decidir lógica de reconciliación manual en esta HU.
7. Inspecciona las migraciones ya aplicadas y diseña una única migración ascendente nueva; ninguna migración anterior se edita.

## Alcance incluido

- Tablas normalizadas `customer`, `appointment`, `appointment_history` y `appointment_history_change` con comentarios de dueño, retención y clasificación.
- Estados cerrados `confirmed`, `completed`, `cancelled_by_customer`, `cancelled_by_barber`, `no_show` y criterio derivado único `occupies_schedule`.
- Snapshots de servicio autorizados por `DEC-004`, sin sincronización automática desde `service`.
- Intervalos `timestamptz` `[inicio, fin)` y restricción `EXCLUDE USING gist` por `barbershop_id`, `barber_id` y rango solo cuando `occupies_schedule`.
- FK compuestas tenant-aware en `ON DELETE RESTRICT`, RLS habilitada/forzada, políticas explícitas, grants mínimos e índices respaldados por consultas.
- Historial y cambios append-only para `barberia_app`; citas sin `DELETE` físico.
- Núcleo `booking` independiente de Chi/PostgreSQL y adaptador `booking/postgres` con una primitiva transaccional interna para cliente+cita+evento inicial.
- Migración Atlas, `atlas.sum`, fixtures y pruebas PostgreSQL reales con dos tenants y dos conexiones.

## Fuera de alcance

- Cualquier path OpenAPI, handler HTTP, formulario Vue o pantalla de agenda.
- Resolver cómo se reutiliza un cliente manual sin teléfono (`DP-CIT-01`). La persistencia puede aceptar ausencia significativa, pero el caso de uso se define en `HU-061` tras la decisión.
- Consultar jornada efectiva, bloqueos o asignaciones servicio-barbero para autorizar una cita.
- Modificar, reprogramar, cancelar, completar, marcar inasistencia o corregir estados.
- Crear `appointment_access_token`, programación de recordatorios, intentos de notificación, worker, anonimización o métricas.
- Cambiar servicios, horarios, bloqueos o idempotencia existentes.

## Estado existente que debe conservarse

- `barbershop`, `staff_user`, `barber`, `service`, `barber_service`, horario, excepciones y bloqueos ya tienen migraciones inmutables.
- `database/modelo-fisico-referencia.sql` propone la sección D y ya fue auditado, pero no refleja automáticamente cada diferencia de las migraciones reales de B1/B2.
- `DEC-045` identifica/reutiliza `customer` por teléfono dentro del tenant; `DEC-046` solo fija unicidad de correo por tenant.
- `states-citas.md` define exactamente los cinco estados, el criterio de ocupación y los ocho `event_type` confirmados por `DEC-041`.
- `idempotency_record` y su protocolo concurrente de `DEC-043` ya existen; esta HU no expone una operación idempotente HTTP.
- El módulo Go dueño es `apps/api/internal/modules/booking`; `schedule` y `catalog` no deben importar sus tablas.

## Trabajo requerido

### 1. Diseño y migración

1. Compara la sección D del modelo de referencia con el esquema migrado real y documenta toda diferencia necesaria antes de generar SQL.
2. Mantén `customer` y los snapshots en 3FN con las únicas desnormalizaciones ya aprobadas; no agregues `jsonb`, arrays relacionales ni datos de contacto en historial por comodidad.
3. Protege estados, origen, duración, intervalos, nombres, moneda, motivos y marcas temporales con constraints declarativas.
4. Implementa la exclusión con `btree_gist` ya disponible; contigüidad debe ser válida y el predicado debe depender del criterio derivado único.
5. Aplica RLS/`FORCE ROW LEVEL SECURITY`, políticas tenant/admin y grants exactos. `appointment`, `appointment_history` y `appointment_history_change` no admiten `DELETE` del API; el historial tampoco `UPDATE`.
6. Añade índices para agenda diaria, historial por cita y consultas de integridad que tengan un consumidor demostrado. Verifícalos con `EXPLAIN` y volumen representativo.
7. Describe bloqueo esperado, actualización desde datos existentes y avance correctivo; no dependas de down migrations destructivas.

### 2. Dominio y persistencia Go

1. Crea tipos de dominio cerrados para estado, origen, intervalo y snapshots; el núcleo no importa Chi, pgx ni DTO HTTP.
2. Define interfaces junto al consumidor y una operación interna que reciba datos ya autorizados; `HU-061` añadirá la política de creación manual.
3. Dentro de una única `InTenantTx`, crea o vincula la fila `customer` que el llamador ya decidió, inserta `appointment` y `appointment_created`; un fallo en cualquier paso revierte todo.
4. Traduce `exclusion_violation`, FK/unique/check y RLS a errores internos tipados, sin filtrar SQL ni IDs ajenos.
5. No abras una transacción por tabla, no llames red y no escribas logs con nombre, teléfono, correo, nota o motivo.

### 3. Evidencia y documentación

1. Añade fixture propio de dos barberías separado de suites existentes y datos exclusivamente ficticios.
2. Actualiza `database/README.md`, matriz, historial y modelo de referencia solo donde el esquema real lo exija.
3. Mantén `HU-061` bloqueada: no marques resuelta ninguna `DP-CIT-*` ni declares que existe creación manual visible.

## Pruebas y evidencia

- PostgreSQL: constraints, cinco estados, criterio de ocupación, snapshots, duración, RLS, FK tenant-aware y grants con el rol real.
- Exclusión directa: cruce total, parciales por ambos extremos, un minuto, contigüidad, barberos distintos, tenants distintos y filas canceladas.
- Tiempo: medianoche y DST con zona de barbería distinta de la sesión PostgreSQL.
- Historial: `INSERT` permitido; `UPDATE`/`DELETE` rechazados en cabecera y cambios.
- Transacción Go: éxito, error tras cliente, error tras cita y cancelación de contexto; ninguna escritura parcial.
- Carrera de dos conexiones sin `sleep`: exactamente una cita cruzada persiste.
- Atlas: hash, validate, apply desde vacío, apply desde versión anterior con datos, segundo apply sin cambios y `status` limpio.

## Documentación y trazabilidad

- Actualiza matriz, historial, `database/README.md`, diagrama/diccionario cuando exista y el estado real de este prompt.
- No agregues una operación OpenAPI ficticia. Si aparece la necesidad, pertenece a `HU-061`.
- Registra cualquier duda o contradicción nueva antes de cambiar el comportamiento o el DDL.

## Verificación final

```text
cd database
atlas migrate hash
atlas migrate validate --env local
atlas migrate apply --env local
atlas migrate status --env local

cd ../apps/api
gofmt -l .
go vet ./...
go test -race ./...
go build ./...

cd ../..
git diff --check
graphify update .
```

Ejecuta además las suites SQL de `HU-060` con `psql -v ON_ERROR_STOP=1` contra PostgreSQL 14 real y el rol `barberia_app`; no declares RLS, exclusión ni concurrencia cumplidas con dobles.

Entrega una tabla `Criterio | Estado | Prueba o evidencia` para `CA-060-01`–`CA-060-08`, más los resultados de base vacía, actualización con datos y carrera de dos conexiones.

## Git y PR

- Rama sugerida después de crear issue: `feat/<issue>-hu060-nucleo-citas`.
- Commit/título: `feat(booking): implementa HU-060 núcleo de citas sin cruces`.
- Usa `Closes #<issue>` solo si los ocho criterios y todas las pruebas obligatorias están completos; de lo contrario, `Refs #<issue>` y enumera lo pendiente.
- No hagas push directo, force push, merge manual de `main`, DDL al arrancar ni cambios a migraciones aplicadas.
