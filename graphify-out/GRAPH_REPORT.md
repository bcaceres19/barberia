# Graph Report - barberia  (2026-08-12)

## Corpus Check
- 152 files · ~167,253 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 758 nodes · 1045 edges · 73 communities (65 shown, 8 thin omitted)
- Extraction: 90% EXTRACTED · 10% INFERRED · 0% AMBIGUOUS · INFERRED: 106 edges (avg confidence: 0.74)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `1f121ede`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Estados de las citas y máquina de transiciones
- Matriz de trazabilidad
- Respuestas a dudas pendientes del proyecto
- run
- devDependencies
- compilerOptions
- Plan de bloques de construcción del MVP
- scripts
- compilerOptions
- Reglas de negocio
- package.json
- OpenAPI entry document
- api/openapi README
- graphify SKILL.md
- tsconfig.vitest.json
- clock.go
- .prettierrc.json
- router/index.ts
- tsconfig.json
- apps/web index.html
- apps/web/e2e README
- BaseAlert.vue
- system-barbershop
- 3. Bloque B0 · Cimientos, seguridad y primeras pantallas
- 3. Prompts por historia
- Alcance del MVP
- Estrategia de pruebas y controles de calidad
- database_test.go
- PostgreSQL, multi-tenancy y consultas
- Load
- Inicio del prompt
- BaseDialog.vue
- BaseInput.vue
- Completar HU-009: sistema visual base
- Implementar HU-003: contrato HTTP base y errores uniformes
- Implementar HU-004: idempotencia reutilizable
- <Título del prompt>
- Catálogo de prompts persistentes
- Prompt de implementación · Llaves asimétricas para la seguridad del API
- Prompts detallados de implementación · B0 en curso
- Sistema de agenda para barberías: reserva pública en línea con garantía de integridad de horarios
- TestNoDatabaseImportsInDomain
- idempotency_concurrency_two_connections.sh
- notification_lease_concurrency_two_connections.sh

## God Nodes (most connected - your core abstractions)
1. `Reglas de negocio` - 49 edges
2. `Respuestas a dudas pendientes del proyecto` - 43 edges
3. `Matriz de trazabilidad` - 42 edges
4. `Historial de cambios documentales` - 40 edges
5. `Dudas pendientes y resoluciones` - 23 edges
6. `Estados de las citas y máquina de transiciones` - 20 edges
7. `Load()` - 18 edges
8. `Estrategia de pruebas y controles de calidad` - 18 edges
9. `Respuestas a propuestas de reglas de negocio` - 16 edges
10. `compilerOptions` - 15 edges

## Surprising Connections (you probably didn't know these)
- `CLAUDE.md (root) — reglas graphify` --semantically_similar_to--> `.claude/CLAUDE.md — trigger de graphify`  [INFERRED] [semantically similar]
  CLAUDE.md → .claude/CLAUDE.md
- `README.md — Sistema de agenda para barberías` --semantically_similar_to--> `Mapa documental y fuentes de verdad`  [INFERRED] [semantically similar]
  README.md → docs/README.md
- `Pull request template` --conceptually_related_to--> `OpenAPI entry document`  [INFERRED]
  .github/PULL_REQUEST_TEMPLATE.md → api/openapi/openapi.yaml
- `Reprogramación o cancelación rápida por emergencia` --references--> `Reglas de negocio`  [INFERRED]
  respuesta-manuales/respuesta-propuestas-oc.txt → docs/01-producto/reglas-negocio.md
- `Potestad del barbero para cancelar citas` --references--> `Reglas de negocio`  [INFERRED]
  respuesta-manuales/respuesta-propuestas-oc.txt → docs/01-producto/reglas-negocio.md

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Governance Decision Chain** — docs_00_control_registro_decisiones, docs_00_control_dudas_pendientes, docs_00_control_contradicciones, docs_00_control_historial_cambios, docs_00_control_matriz_trazabilidad [INFERRED 0.80]
- **Graphify Reference Docs** — claude_skills_graphify_skill, claude_skills_graphify_references_query, claude_skills_graphify_references_update, claude_skills_graphify_references_exports, claude_skills_graphify_references_extraction_spec, claude_skills_graphify_references_add_watch, claude_skills_graphify_references_github_and_merge, claude_skills_graphify_references_hooks, claude_skills_graphify_references_transcribe [EXTRACTED 0.90]
- **Empty Scaffold READMEs** — apps_api_readme, apps_web_readme, database_readme, api_openapi_readme [INFERRED 0.75]
- **Red de decisiones que confirman el alcance P0 del MVP** — docs_01_producto_alcance_mvp_doc, docs_01_producto_prioridades_doc, docs_01_producto_reglas_negocio_doc, docs_02_requisitos_estados_citas_doc, docs_01_producto_prioridades_dec_032, docs_02_requisitos_estados_citas_dec_017, docs_01_producto_alcance_mvp_f_not_03, docs_01_producto_alcance_mvp_f_est_04 [INFERRED 0.80]
- **Pila y estándares del backend Go/Chi/PostgreSQL** — docs_03_desarrollo_estandar_backend_go_doc, docs_04_arquitectura_backend_go_doc, docs_04_arquitectura_stack_despliegue_operacion_doc, docs_05_backend_base_datos_doc, docs_05_backend_estandar_base_datos_doc, docs_04_arquitectura_backend_go_dec_034, docs_03_desarrollo_estandar_backend_go_dec_035 [EXTRACTED 0.85]
- **Garantía de integridad de agenda sin cruces de citas** — docs_01_producto_reglas_negocio_rn_con_01, docs_01_producto_reglas_negocio_rn_con_03, docs_01_producto_reglas_negocio_rn_dis_05, docs_05_backend_base_datos_appointment_no_overlap, docs_02_requisitos_estados_citas_confirmed, docs_02_requisitos_estados_citas_doc [EXTRACTED 0.90]
- **Comparación de herramientas de migración de base de datos** — concept_atlas_cli, concept_goose, concept_golang_migrate, concept_flyway [INFERRED 0.70]
- **Decisiones de seguridad para el piloto** — respuesta_manuales_respuesta_dudas_pendientes_dp_seg_01, respuesta_manuales_respuesta_dudas_pendientes_dp_seg_02, respuesta_manuales_respuesta_dudas_pendientes_dp_seg_03 [EXTRACTED 0.75]
- **Reglas de disponibilidad y agendamiento** — respuesta_manuales_respuesta_propuestas_oc_rn_dis_04, respuesta_manuales_respuesta_propuestas_oc_rn_dis_05, respuesta_manuales_respuesta_propuestas_oc_rn_dis_06, respuesta_manuales_respuesta_propuestas_oc_rn_dis_07 [EXTRACTED 0.75]

## Communities (73 total, 8 thin omitted)

### Community 0 - "Estados de las citas y máquina de transiciones"
Cohesion: 0.26
Nodes (12): RN-CIT-03: Solo transiciones de estado definidas, RN-CIT-04: Corrección auditada de un estado erróneo, RN-CIT-05: Cierre configurable de citas vencidas, Estado: cancelled_by_barber, Estado: cancelled_by_customer, Estado: completed, Estado: confirmed, DEC-017: no_show elevado a P0 (+4 more)

### Community 1 - "Matriz de trazabilidad"
Cohesion: 0.08
Nodes (62): AGENTS.md — Reglas de desarrollo del proyecto, apps/api README, apps/web README, CONTRIBUTING.md — Cómo contribuir, database/migrations README, database README, database/seeds README, database/testdata README (+54 more)

### Community 2 - "Respuestas a dudas pendientes del proyecto"
Cohesion: 0.05
Nodes (43): Respuestas a dudas pendientes del proyecto, Notificaciones y recordatorios creados en el momento, Elección de PostgreSQL, Multi-tenant compartido con buenas prácticas de consulta, Retención de datos de 24 meses configurable, Arquitectura backend económica en recursos, Bloqueos de horario manuales y recurrentes, Horario laboral configurable por el barbero (+35 more)

### Community 3 - "run"
Cohesion: 0.08
Nodes (34): main(), run(), main(), run(), DatabaseHealthHandler(), HealthHandler(), T, TestRecover_DoesNotLogRawPanicValue() (+26 more)

### Community 4 - "devDependencies"
Cohesion: 0.06
Nodes (31): devDependencies, eslint, eslint-plugin-vue, jsdom, @playwright/test, prettier, @types/node, typescript (+23 more)

### Community 5 - "compilerOptions"
Cohesion: 0.09
Nodes (22): compilerOptions, allowImportingTsExtensions, erasableSyntaxOnly, lib, module, moduleDetection, noEmit, noFallthroughCasesInSwitch (+14 more)

### Community 6 - "Plan de bloques de construcción del MVP"
Cohesion: 0.12
Nodes (16): 1. Para qué sirve este documento, 2. Criterio de ordenamiento, 3. Mapa de bloques, 4. Detalle de cada bloque, 5. Reglas de secuencia, 6. Cobertura de las 45 funciones P0, 7. Qué queda fuera de este plan, 8. Estado y siguiente paso (+8 more)

### Community 7 - "scripts"
Cohesion: 0.10
Nodes (19): dependencies, vue, vue-router, name, private, scripts, build, dev (+11 more)

### Community 8 - "compilerOptions"
Cohesion: 0.10
Nodes (19): compilerOptions, allowArbitraryExtensions, erasableSyntaxOnly, noFallthroughCasesInSwitch, noUnusedLocals, noUnusedParameters, paths, tsBuildInfoFile (+11 more)

### Community 9 - "Reglas de negocio"
Cohesion: 0.10
Nodes (31): DEC-005: Anticipación mínima y ventana máxima elevadas a P0, DEC-018: Valores iniciales de configuración, DEC-020: Bloqueos recurrentes y calendario colombiano activable, DEC-027: Correo y WhatsApp oficial como canales del MVP, Reglas de negocio, RN-BLQ-03: Un bloqueo puede solaparse con citas existentes, RN-CAN-03: El barbero cancela en cualquier momento, RN-CAN-04: La cancelación libera el horario de inmediato (+23 more)

### Community 10 - "package.json"
Cohesion: 0.14
Nodes (13): description, devDependencies, @redocly/cli, name, private, scripts, openapi:bundle, openapi:check-config (+5 more)

### Community 11 - "OpenAPI entry document"
Cohesion: 0.67
Nodes (3): OpenAPI contract changelog, OpenAPI entry document, redocly.yaml — configuración Redocly

### Community 12 - "api/openapi README"
Cohesion: 0.15
Nodes (7): components/headers README, components/parameters README, components/responses README, components/schemas README, components/security-schemes README, examples README, api/openapi README

### Community 13 - "graphify SKILL.md"
Cohesion: 0.18
Nodes (11): CLAUDE.md (root) — reglas graphify, .claude/CLAUDE.md — trigger de graphify, graphify reference: add & watch, graphify reference: exports, graphify reference: extraction spec, graphify reference: GitHub clone & merge, graphify reference: hooks & CLAUDE.md integration, graphify reference: query/path/explain (+3 more)

### Community 14 - "tsconfig.vitest.json"
Cohesion: 0.20
Nodes (9): compilerOptions, tsBuildInfoFile, types, extends, include, node, src/**/*.spec.ts, jsdom (+1 more)

### Community 15 - "clock.go"
Cohesion: 0.40
Nodes (3): Clock, System, Time

### Community 16 - ".prettierrc.json"
Cohesion: 0.40
Nodes (4): printWidth, semi, singleQuote, trailingComma

### Community 48 - "BaseAlert.vue"
Cohesion: 0.06
Nodes (31): alertRef, borderVar, classes, emit, focusableElementsRef, handleActionClick(), handleDismiss(), handleKeyDown() (+23 more)

### Community 51 - "3. Bloque B0 · Cimientos, seguridad y primeras pantallas"
Cohesion: 0.09
Nodes (22): 1. Cómo leer este documento, 2. Estado del catálogo, 3. Bloque B0 · Cimientos, seguridad y primeras pantallas, 4. Bloque B1 · Identidad de la barbería y catálogo, 5. Historias pendientes de redacción, 6. Dudas que bloqueaban historias de B0, Convención de las tablas de cabecera, Historias de usuario y criterios de aceptación (+14 more)

### Community 52 - "3. Prompts por historia"
Cohesion: 0.09
Nodes (22): 1. Cómo usar estos prompts, 2. Preámbulo obligatorio, 3. Prompts por historia, 4. Prompts de B1 preparados, todavía bloqueados por B0, 5. Prompt de revisión, 6. Mantenimiento de este documento, Prompt de `HU-001` · Esquema inicial con aislamiento por barbería, Prompt de `HU-002` · Contexto de barbería en cada solicitud (+14 more)

### Community 53 - "Alcance del MVP"
Cohesion: 0.20
Nodes (12): DEC-019: El MVP administra uno o varios barberos, Alcance del MVP, F-CITA-10: Gestión y comunicación de retrasos (P1), F-EST-04: Estado no_show (P0), F-NOT-03: Recordatorios automáticos (P0), DEC-010: Política de cancelación fuera de plazo elevada a P0, DEC-032: Recordatorios automáticos elevados a P0 (resuelve CT-001), Sistema de prioridades y clasificación de funciones (+4 more)

### Community 54 - "Estrategia de pruebas y controles de calidad"
Cohesion: 0.13
Nodes (24): DEC-016, JSON Schema Draft 2020-12, OpenAPI Specification 3.1.2, Redocly CLI, RFC 9457 Problem Details, RN-IDE-01: Operaciones críticas seguras ante reintentos, DEC-035: Estándar de código backend/base de datos obligatorio, Estándar de código del backend en Go (+16 more)

### Community 55 - "database_test.go"
Cohesion: 0.15
Nodes (22): Context, Pool, NewDB(), Pool, T, setupTestDB(), TestContextCancellation(), TestHealthCheck() (+14 more)

### Community 56 - "PostgreSQL, multi-tenancy y consultas"
Cohesion: 0.19
Nodes (15): Atlas CLI, Flyway, golang-migrate, Goose, RN-CON-03: La base de datos debe impedir el cruce, RN-DIS-05: Intervalos semiabiertos [inicio, fin), RN-TEN-01: Ninguna barbería accede a datos de otra, Restricción de exclusión appointment_no_overlap (+7 more)

### Community 57 - "Load"
Cohesion: 0.20
Nodes (23): getEnv(), getEnvDuration(), getEnvInt(), Load(), redactDSN(), requireTLS(), baseLocalEnv(), T (+15 more)

### Community 58 - "Inicio del prompt"
Cohesion: 0.07
Nodes (25): 1. Resultado ejecutivo, 2. Alcance y límites de la revisión, 3. Fortalezas que deben conservarse, 4.1 Bloqueantes de seguridad, concurrencia o privacidad, 4.2 Integridad y consistencia funcional, 4.3 Privilegios, rendimiento y pruebas, 4. Hallazgos priorizados, 5. Decisiones que deben registrarse antes de codificar (+17 more)

### Community 59 - "BaseDialog.vue"
Cohesion: 0.12
Nodes (18): classes, close(), closeButtonRef, dialogRef, emit, focusableElementsRef, handleBackdropClick(), handleCloseClick() (+10 more)

### Community 60 - "BaseInput.vue"
Cohesion: 0.15
Nodes (15): classes, describedBy, emit, errorId, handleBlur(), handleChange(), handleFocus(), handleInput() (+7 more)

### Community 61 - "Completar HU-009: sistema visual base"
Cohesion: 0.15
Nodes (13): Alcance incluido, Auditoría inicial requerida, Completar HU-009: sistema visual base, Documentación y trazabilidad, Estado existente que debes conservar, Fuera de alcance, Git y PR, Instrucción para Claude o Codex (+5 more)

### Community 62 - "Implementar HU-003: contrato HTTP base y errores uniformes"
Cohesion: 0.17
Nodes (12): Alcance incluido, Documentación y trazabilidad, Estado que debes conservar, Fuera de alcance, Git y PR, Implementar HU-003: contrato HTTP base y errores uniformes, Instrucción para Claude o Codex, Objetivo (+4 more)

### Community 63 - "Implementar HU-004: idempotencia reutilizable"
Cohesion: 0.17
Nodes (12): Alcance incluido, Documentación y trazabilidad, Estado existente que debes conservar, Fuera de alcance, Git y PR, Implementar HU-004: idempotencia reutilizable, Instrucción para Claude o Codex, Objetivo (+4 more)

### Community 64 - "<Título del prompt>"
Cohesion: 0.17
Nodes (12): Alcance incluido, Documentación y trazabilidad, Estado existente que debe conservarse, Fuera de alcance, Git y PR, Instrucción para el agente, Objetivo, Preflight obligatorio (+4 more)

### Community 65 - "Catálogo de prompts persistentes"
Cohesion: 0.18
Nodes (11): 10. Lista de control al guardar o entregar, 1. Propósito, 2. Qué debe guardarse, 3. Organización y nombres, 4. Metadatos obligatorios, 5. Estados y transición, 6. Contenido autocontenido, 7. Una preocupación por prompt (+3 more)

### Community 66 - "Prompt de implementación · Llaves asimétricas para la seguridad del API"
Cohesion: 0.22
Nodes (9): 1.1 Esto no es una historia nueva: resuelve una duda existente, 1.2 Qué gana realmente este proyecto con llaves asimétricas, 1.3 El diseño que sí satisface la documentación, 1. Léase esto antes que el prompt, 2. Decisión que debe existir antes de codificar, 3. Prompt, 4. Efecto de esta decisión en el modelo de datos, 5. Riesgo que este documento deja consignado (+1 more)

### Community 68 - "Prompts detallados de implementación · B0 en curso"
Cohesion: 0.29
Nodes (7): 1. Relación con `prompts-implementacion.md`, 2. Estado real del repositorio al escribir estos prompts, 3. Prompt detallado · `HU-002` — Contexto de barbería en cada solicitud, 4. Prompt detallado · `HU-009` — Sistema visual base en componentes, 5. Qué hacer al cerrar estas dos historias, Por qué estas dos y en paralelo, Prompts detallados de implementación · B0 en curso

### Community 69 - "Sistema de agenda para barberías: reserva pública en línea con garantía de integridad de horarios"
Cohesion: 0.40
Nodes (4): 1. Introducción, Cómo se corresponde con el ejemplo de referencia, Nota sobre las referencias, Sistema de agenda para barberías: reserva pública en línea con garantía de integridad de horarios

## Knowledge Gaps
- **321 isolated node(s):** `system-barbershop`, `Clock`, `httprobProblem`, `requestIDKey`, `semi` (+316 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **8 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Estándar de documentación OpenAPI` connect `Estrategia de pruebas y controles de calidad` to `Estados de las citas y máquina de transiciones`, `Matriz de trazabilidad`, `Reglas de negocio`?**
  _High betweenness centrality (0.093) - this node is a cross-community bridge._
- **Why does `Registro de decisiones` connect `Matriz de trazabilidad` to `PostgreSQL, multi-tenancy y consultas`, `Estrategia de pruebas y controles de calidad`?**
  _High betweenness centrality (0.093) - this node is a cross-community bridge._
- **Why does `Reglas de negocio` connect `Reglas de negocio` to `Estados de las citas y máquina de transiciones`, `PostgreSQL, multi-tenancy y consultas`, `Alcance del MVP`, `Estrategia de pruebas y controles de calidad`?**
  _High betweenness centrality (0.085) - this node is a cross-community bridge._
- **Are the 15 inferred relationships involving `Reglas de negocio` (e.g. with `Reprogramación o cancelación rápida por emergencia` and `Persistencia de historial de bloqueos`) actually correct?**
  _`Reglas de negocio` has 15 INFERRED edges - model-reasoned connections that need verification._
- **What connects `system-barbershop`, `Clock`, `httprobProblem` to the rest of the system?**
  _321 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Matriz de trazabilidad` be split into smaller, more focused modules?**
  _Cohesion score 0.08085317460317461 - nodes in this community are weakly interconnected._
- **Should `Respuestas a dudas pendientes del proyecto` be split into smaller, more focused modules?**
  _Cohesion score 0.048726467331118496 - nodes in this community are weakly interconnected._