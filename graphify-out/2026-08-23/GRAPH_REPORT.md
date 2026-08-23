# Graph Report - agent-hu011  (2026-08-23)

## Corpus Check
- 317 files · ~336,255 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2115 nodes · 4029 edges · 163 communities (132 shown, 31 thin omitted)
- Extraction: 83% EXTRACTED · 17% INFERRED · 0% AMBIGUOUS · INFERRED: 676 edges (avg confidence: 0.79)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `4af0b5a4`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Estados de las citas y máquina de transiciones
- Matriz de trazabilidad
- Respuestas a dudas pendientes del proyecto
- NewRouter
- devDependencies
- compilerOptions
- Plan de bloques de construcción del MVP
- dependencies
- compilerOptions
- Reglas de negocio
- package.json
- Translate
- RecoveryService
- graphify SKILL.md
- tsconfig.vitest.json
- NewSessionService
- .prettierrc.json
- router/index.ts
- tsconfig.json
- apps/web index.html
- apps/web/e2e README
- loginApi.test.ts
- LoginForm.vue
- openapi.d.ts
- BaseInput.vue
- system-barbershop
- 3. Bloque B0 · Cimientos, seguridad y primeras pantallas
- 3. Prompts por historia
- Alcance del MVP
- Estrategia de pruebas y controles de calidad
- BarbershopID
- PostgreSQL, multi-tenancy y consultas
- Load
- Inicio del prompt
- BaseDialog.vue
- httpserver/idempotency_test.go
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
- RequestID
- vitest-axe.d.ts
- BaseAlert.vue
- scripts
- eslint-plugin-vue
- SecurityHeaders
- acceso-evidencia-responsiva.spec.ts
- vitest-axe
- @playwright/test
- @types/node
- @vue/eslint-config-prettier
- @vue/test-utils
- jsdom
- NewPhoneChallengeService
- @vue/eslint-config-typescript
- assertProblem
- NewThrottleService
- setupTestDB
- httpapi/contract_test.go
- recovery_test.go
- Implementar HU-005: inicio de sesión seguro del barbero
- typescript
- Implementar HU-006: sesión persistente y cierre de sesión
- Implementar HU-010: pantalla de acceso
- vite
- buildRouter
- .Now
- contract_logout_test.go
- Resolve
- PhoneChallengeForm.vue
- WriteProblem
- whatsapp_meta.go
- NewPurgeService
- Implementar HU-008: recuperación de acceso con código
- NewDualChannelRecoverySender
- sessionStore.ts
- Implementar HU-007: defensa escalonada contra abuso en el acceso
- Implementar HU-012: cascarón del panel privado
- newResendSenderAgainstTestServer
- LoginPage.test.ts
- LoginPage.vue
- dto.go
- Internal
- As
- BaseBadge.vue
- echoBodyHandler
- fakeRepository
- resetForFreshLogin
- TestContract_TranslateAlwaysFillsRequiredFields
- BaseButton.vue
- requireSession.ts
- contract_session_test.go
- routes.test.ts
- reto-telefonico.spec.ts
- web/package.json
- challengeApi.test.ts
- PrivateShell.test.ts
- sessionContextApi.test.ts
- installSessionHandling.test.ts
- AppHeader.test.ts
- fakeHasher
- panel-evidencia-responsiva.spec.ts
- failingDummyHasher
- openapi-typescript
- RecoveryRequestStep.vue
- DB
- .Begin
- idempotency/idempotency_test.go
- newLoginHandlerWithThrottle
- setupTestDB
- RecoveryVerifyStep.vue
- RecoveryResetStep.vue
- RecoveryPage.vue
- Context
- Implementar HU-011: pantalla de recuperación de acceso
- ThrottleService
- .ValidateAndRenewSession
- LoginForm.test.ts
- RecoveryPage.test.ts
- recoveryApi.ts
- recoveryApi.test.ts
- RecoveryResetStep.test.ts
- RecoveryVerifyStep.test.ts
- recuperacion.spec.ts

## God Nodes (most connected - your core abstractions)
1. `buildRouter()` - 50 edges
2. `Reglas de negocio` - 49 edges
3. `As()` - 45 edges
4. `Respuestas a dudas pendientes del proyecto` - 43 edges
5. `setupTestDB()` - 42 edges
6. `Matriz de trazabilidad` - 42 edges
7. `Historial de cambios documentales` - 40 edges
8. `DB` - 33 edges
9. `assertProblem()` - 29 edges
10. `Load()` - 29 edges

## Surprising Connections (you probably didn't know these)
- `CLAUDE.md (root) — reglas graphify` --semantically_similar_to--> `.claude/CLAUDE.md — trigger de graphify`  [INFERRED] [semantically similar]
  CLAUDE.md → .claude/CLAUDE.md
- `README.md — Sistema de agenda para barberías` --semantically_similar_to--> `Mapa documental y fuentes de verdad`  [INFERRED] [semantically similar]
  README.md → docs/README.md
- `Reprogramación o cancelación rápida por emergencia` --references--> `Reglas de negocio`  [INFERRED]
  respuesta-manuales/respuesta-propuestas-oc.txt → docs/01-producto/reglas-negocio.md
- `Potestad del barbero para cancelar citas` --references--> `Reglas de negocio`  [INFERRED]
  respuesta-manuales/respuesta-propuestas-oc.txt → docs/01-producto/reglas-negocio.md
- `Optimización de horarios del barbero` --references--> `Reglas de negocio`  [INFERRED]
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

## Communities (163 total, 31 thin omitted)

### Community 0 - "Estados de las citas y máquina de transiciones"
Cohesion: 0.26
Nodes (12): RN-CIT-03: Solo transiciones de estado definidas, RN-CIT-04: Corrección auditada de un estado erróneo, RN-CIT-05: Cierre configurable de citas vencidas, Estado: cancelled_by_barber, Estado: cancelled_by_customer, Estado: completed, Estado: confirmed, DEC-017: no_show elevado a P0 (+4 more)

### Community 1 - "Matriz de trazabilidad"
Cohesion: 0.06
Nodes (72): AGENTS.md — Reglas de desarrollo del proyecto, OpenAPI contract changelog, components/headers README, components/parameters README, components/responses README, components/schemas README, components/security-schemes README, examples README (+64 more)

### Community 2 - "Respuestas a dudas pendientes del proyecto"
Cohesion: 0.05
Nodes (43): Respuestas a dudas pendientes del proyecto, Notificaciones y recordatorios creados en el momento, Elección de PostgreSQL, Multi-tenant compartido con buenas prácticas de consulta, Retención de datos de 24 meses configurable, Arquitectura backend económica en recursos, Bloqueos de horario manuales y recurrentes, Horario laboral configurable por el barbero (+35 more)

### Community 3 - "NewRouter"
Cohesion: 0.07
Nodes (36): Time, alwaysUnauthorized(), chiWalkFinds(), Handler, Mux, T, TestPrivateSubrouter_BypassingItSkipsTheMiddleware(), TestPrivateSubrouter_MiddlewareAppliesToRoutesRegisteredOnIt() (+28 more)

### Community 4 - "devDependencies"
Cohesion: 0.13
Nodes (15): devDependencies, axe-core, eslint, prettier, @vitejs/plugin-vue, vitest, vue-tsc, @vue/tsconfig (+7 more)

### Community 5 - "compilerOptions"
Cohesion: 0.09
Nodes (22): compilerOptions, allowImportingTsExtensions, erasableSyntaxOnly, lib, module, moduleDetection, noEmit, noFallthroughCasesInSwitch (+14 more)

### Community 6 - "Plan de bloques de construcción del MVP"
Cohesion: 0.12
Nodes (16): 1. Para qué sirve este documento, 2. Criterio de ordenamiento, 3. Mapa de bloques, 4. Detalle de cada bloque, 5. Reglas de secuencia, 6. Cobertura de las 45 funciones P0, 7. Qué queda fuera de este plan, 8. Estado y siguiente paso (+8 more)

### Community 7 - "dependencies"
Cohesion: 0.29
Nodes (7): dependencies, openapi-fetch, vue, vue-router, openapi-fetch, vue, vue-router

### Community 8 - "compilerOptions"
Cohesion: 0.10
Nodes (19): compilerOptions, allowArbitraryExtensions, erasableSyntaxOnly, noFallthroughCasesInSwitch, noUnusedLocals, noUnusedParameters, paths, tsBuildInfoFile (+11 more)

### Community 9 - "Reglas de negocio"
Cohesion: 0.10
Nodes (31): DEC-005: Anticipación mínima y ventana máxima elevadas a P0, DEC-018: Valores iniciales de configuración, DEC-020: Bloqueos recurrentes y calendario colombiano activable, DEC-027: Correo y WhatsApp oficial como canales del MVP, Reglas de negocio, RN-BLQ-03: Un bloqueo puede solaparse con citas existentes, RN-CAN-03: El barbero cancela en cualquier momento, RN-CAN-04: La cancelación libera el horario de inmediato (+23 more)

### Community 10 - "package.json"
Cohesion: 0.14
Nodes (13): description, devDependencies, @redocly/cli, name, private, scripts, openapi:bundle, openapi:check-config (+5 more)

### Community 11 - "Translate"
Cohesion: 0.27
Nodes (15): newProblem(), T, TestTranslate_IdempotencyConflict_MapsTo409(), TestTranslate_IdempotencyLocked_MapsTo409WithoutWaiting(), TestTranslate_Internal_NeverExposesCause(), TestTranslate_Invalid_ExposesSafeMessage(), TestTranslate_MaxBytesError_MapsToPayloadTooLarge(), TestTranslate_NotFound_ExposesSafeMessage() (+7 more)

### Community 12 - "RecoveryService"
Cohesion: 0.06
Nodes (53): Time, NormalizeEmail(), errInvalidCredentials(), Logger, NewRecoveryRequestHandler(), NewRecoveryResetPasswordHandler(), NewRecoveryVerifyHandler(), discardRecoveryLogger() (+45 more)

### Community 13 - "graphify SKILL.md"
Cohesion: 0.18
Nodes (11): CLAUDE.md (root) — reglas graphify, .claude/CLAUDE.md — trigger de graphify, graphify reference: add & watch, graphify reference: exports, graphify reference: extraction spec, graphify reference: GitHub clone & merge, graphify reference: hooks & CLAUDE.md integration, graphify reference: query/path/explain (+3 more)

### Community 14 - "tsconfig.vitest.json"
Cohesion: 0.20
Nodes (9): compilerOptions, tsBuildInfoFile, types, extends, include, node, src/**/*.spec.ts, jsdom (+1 more)

### Community 15 - "NewSessionService"
Cohesion: 0.07
Nodes (48): NewLogoutHandler(), T, newLogoutHandler(), TestLogoutHandler_MissingPrincipalInContext_ReturnsSafe500(), TestLogoutHandler_ServiceFailure_ReturnsSafe500(), TestLogoutHandler_Success_RevokesAndClearsCookie(), Context, Time (+40 more)

### Community 16 - ".prettierrc.json"
Cohesion: 0.40
Nodes (4): printWidth, semi, singleQuote, trailingComma

### Community 38 - "LoginForm.vue"
Cohesion: 0.18
Nodes (16): attemptedSubmit, emailError, emit, errorCount, fieldErrors, handleEmailInput(), handlePasswordInput(), onRetry() (+8 more)

### Community 45 - "openapi.d.ts"
Cohesion: 0.21
Nodes (10): components, $defs, operations, paths, RFC-9457, webhooks, httpClient, isProblem() (+2 more)

### Community 48 - "BaseInput.vue"
Cohesion: 0.14
Nodes (16): classes, describedBy, emit, errorId, handleBlur(), handleChange(), handleFocus(), handleInput() (+8 more)

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

### Community 55 - "BarbershopID"
Cohesion: 0.32
Nodes (15): Pool, T, setupTestDB(), TestContextCancellation(), TestHealthCheck(), TestInTenantTx_CA002_01(), TestInTenantTx_CA002_02(), TestInTenantTx_CA002_03() (+7 more)

### Community 56 - "PostgreSQL, multi-tenancy y consultas"
Cohesion: 0.19
Nodes (15): Atlas CLI, Flyway, golang-migrate, Goose, RN-CON-03: La base de datos debe impedir el cruce, RN-DIS-05: Intervalos semiabiertos [inicio, fin), RN-TEN-01: Ninguna barbería accede a datos de otra, Restricción de exclusión appointment_no_overlap (+7 more)

### Community 57 - "Load"
Cohesion: 0.17
Nodes (34): getEnv(), getEnvCSV(), getEnvDuration(), getEnvInt(), Duration, Load(), redactDSN(), requireTLS() (+26 more)

### Community 58 - "Inicio del prompt"
Cohesion: 0.07
Nodes (25): 1. Resultado ejecutivo, 2. Alcance y límites de la revisión, 3. Fortalezas que deben conservarse, 4.1 Bloqueantes de seguridad, concurrencia o privacidad, 4.2 Integridad y consistencia funcional, 4.3 Privilegios, rendimiento y pruebas, 4. Hallazgos priorizados, 5. Decisiones que deben registrarse antes de codificar (+17 more)

### Community 59 - "BaseDialog.vue"
Cohesion: 0.09
Nodes (21): classes, close(), closeButtonRef, dialogRef, emit, focusableElementsRef, handleBackdropClick(), handleCloseClick() (+13 more)

### Community 60 - "httpserver/idempotency_test.go"
Cohesion: 0.19
Nodes (23): Request, ResponseWriter, IdempotencyFingerprint(), IdempotencyKeyFromRequest(), doIdempotentRequest(), doIdempotentRequestObject(), Handler, HandlerFunc (+15 more)

### Community 61 - "Completar HU-009: sistema visual base"
Cohesion: 0.15
Nodes (13): Alcance incluido, Auditoría inicial requerida, Completar HU-009: sistema visual base, Documentación y trazabilidad, Estado existente que debes conservar, Fuera de alcance, Git y PR, Instrucción para Claude o Codex (+5 more)

### Community 62 - "Implementar HU-003: contrato HTTP base y errores uniformes"
Cohesion: 0.15
Nodes (12): Alcance incluido, Documentación y trazabilidad, Estado que debes conservar, Fuera de alcance, Git y PR, Implementar HU-003: contrato HTTP base y errores uniformes, Instrucción para Claude o Codex, Objetivo (+4 more)

### Community 63 - "Implementar HU-004: idempotencia reutilizable"
Cohesion: 0.15
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

### Community 73 - "RequestID"
Cohesion: 0.16
Nodes (15): T, TestRecover_DoesNotLogRawPanicValue(), TestRecover_RespondsWithProblemJSON(), TestRequestID_AcceptsValidClientID(), TestRequestID_RejectsInvalidClientID(), TestWriteProblem_SetsContentTypeAndStatus(), Handler, Logger (+7 more)

### Community 74 - "vitest-axe.d.ts"
Cohesion: 0.50
Nodes (3): Assertion, AsymmetricMatchersContaining, vitest

### Community 75 - "BaseAlert.vue"
Cohesion: 0.12
Nodes (18): alertRef, borderVar, classes, emit, focusableElementsRef, handleActionClick(), handleDismiss(), handleKeyDown() (+10 more)

### Community 76 - "scripts"
Cohesion: 0.18
Nodes (11): scripts, build, dev, format, format:write, generate:api, lint, preview (+3 more)

### Community 78 - "SecurityHeaders"
Cohesion: 0.33
Nodes (4): Handler, SecurityHeaders(), T, TestSecurityHeaders_SetsExpectedHeaders()

### Community 87 - "NewPhoneChallengeService"
Cohesion: 0.06
Nodes (56): NewChallengeHandler(), NewChallengeVerifyHandler(), Context, T, Time, newChallengeHandler(), newVerifyHandler(), testChallengeCfg() (+48 more)

### Community 91 - "assertProblem"
Cohesion: 0.11
Nodes (44): DefaultCookieConfig(), NewLoginHandler(), assertProblem(), doLogin(), Repository, Request, ResponseRecorder, T (+36 more)

### Community 92 - "NewThrottleService"
Cohesion: 0.20
Nodes (17): Context, Time, NewThrottleService(), Context, T, Time, testThrottleCfg(), TestThrottleService_BelowThreshold_ReturnsNil() (+9 more)

### Community 93 - "setupTestDB"
Cohesion: 0.06
Nodes (74): HMACHex(), Context, NewPhoneChallengeRepository(), T, testChallengeCfg(), testCodeHash(), TestPhoneChallengeRepository_RequestChallenge_NonEscalatedIP_NotAccepted(), TestPhoneChallengeRepository_RequestChallenge_ResendCooldown_RejectsImmediate() (+66 more)

### Community 94 - "httpapi/contract_test.go"
Cohesion: 0.10
Nodes (26): T, TestContract_ChallengeOperations_MethodPathSecurityAndResponses(), TestContract_ChallengeRequestSchema_MatchesDTOFields(), TestContract_ChallengeVerifyRequestSchema_MatchesDTOFields(), T, TestContract_RecoveryOperations_MethodPathSecurityAndResponses(), TestContract_RecoveryRequestSchema_MatchesDTOFields(), TestContract_RecoveryResetPasswordRequestSchema_MatchesDTOFields() (+18 more)

### Community 95 - "recovery_test.go"
Cohesion: 0.12
Nodes (27): NewArgon2Hasher(), parseEncodedHash(), T, TestArgon2Hasher_HashThenVerify_Succeeds(), TestArgon2Hasher_TwoUsersSamePassword_ProduceDistinctEncodedValues(), TestArgon2Hasher_Verify_MalformedEncodedHash_NeverPanicsOrErrors(), TestArgon2Hasher_Verify_WrongPasswordFails(), Context (+19 more)

### Community 96 - "Implementar HU-005: inicio de sesión seguro del barbero"
Cohesion: 0.15
Nodes (13): Alcance incluido, Documentación y trazabilidad, Estado existente que debes conservar, Fuera de alcance, Git y PR, Implementar HU-005: inicio de sesión seguro del barbero, Instrucción para Claude o Codex, Objetivo (+5 more)

### Community 98 - "Implementar HU-006: sesión persistente y cierre de sesión"
Cohesion: 0.17
Nodes (12): Alcance incluido, Documentación y trazabilidad, Estado existente que debes conservar, Fuera de alcance, Git y PR, Implementar HU-006: sesión persistente y cierre de sesión, Instrucción para Claude o Codex, Objetivo (+4 more)

### Community 99 - "Implementar HU-010: pantalla de acceso"
Cohesion: 0.15
Nodes (12): Alcance incluido, Documentación y trazabilidad, Estado existente que debes conservar, Fuera de alcance, Git y PR, Implementar HU-010: pantalla de acceso, Instrucción para Claude o Codex, Objetivo (+4 more)

### Community 101 - "buildRouter"
Cohesion: 0.07
Nodes (59): recoveryCapture, buildRouter(), Logger, Mux, main(), run(), buildRouterWithRecoveryCapture(), doRecoveryRequest() (+51 more)

### Community 102 - ".Now"
Cohesion: 0.34
Nodes (18): T, newTestService(), requireInternal(), requireUnauthorized(), TestLogin_CancelledContext_NeverCallsRepository(), TestLogin_CreateSessionError_ReturnsInternal(), TestLogin_InactiveUser_IsIndistinguishableFromUnknownEmail(), TestLogin_LookupCredentialError_ReturnsInternal() (+10 more)

### Community 103 - "contract_logout_test.go"
Cohesion: 0.43
Nodes (6): T, TestContract_LogoutOperation_MethodPathSecurityAndResponses(), TestContract_OpenAPIYAML_RegistersSessionCookieSecurityScheme(), TestContract_SessionCookieSecurityScheme_MatchesRuntimeConfig(), logoutOperation, logoutPathFile

### Community 105 - "Resolve"
Cohesion: 0.25
Nodes (21): Request, ParseTrustedProxies(), Resolve(), Request, T, mustTrusted(), newRequest(), TestParseTrustedProxies_InvalidCIDRErrors() (+13 more)

### Community 106 - "PhoneChallengeForm.vue"
Cohesion: 0.11
Nodes (18): requestChallenge(), verifyChallenge(), canRequest, canVerify, code, codeDigitsOnly, emit, onRequestCode() (+10 more)

### Community 107 - "WriteProblem"
Cohesion: 0.15
Nodes (17): Request, ResponseWriter, Request, ResponseWriter, Request, ResponseWriter, Handler, Request (+9 more)

### Community 108 - "whatsapp_meta.go"
Cohesion: 0.15
Nodes (18): Client, Context, NewMetaWhatsAppSender(), Request, Server, T, newMetaSenderAgainstTestServer(), TestMetaWhatsAppSender_Send_ErrorStatus_ReturnsErrorWithoutLeakingBody() (+10 more)

### Community 109 - "NewPurgeService"
Cohesion: 0.10
Nodes (22): main(), run(), Context, NewPurgeService(), Context, T, TestPurgeService_CancelledContext_NeverCallsRepository(), TestPurgeService_LoginThrottleError_SkipsRest() (+14 more)

### Community 110 - "Implementar HU-008: recuperación de acceso con código"
Cohesion: 0.11
Nodes (18): Alcance incluido, Bloqueos resueltos: `DP-SEG-11` → `DEC-063`, `DP-SEG-12` → `DEC-064`, `CT-006` → `DEC-065`, Criterios de aceptación y cierre, `CT-006` · No enumeración frente a destino enmascarado, Documentación y trazabilidad, `DP-NOT-05` · Proveedor oficial concreto, `DP-SEG-11` · Política de contraseña nueva, `DP-SEG-12` · Parámetros y autorización del flujo (+10 more)

### Community 111 - "NewDualChannelRecoverySender"
Cohesion: 0.19
Nodes (11): Context, NewDualChannelRecoverySender(), Context, T, TestDualChannelRecoverySender_BothChannelsFail_ReturnsCombinedError(), TestDualChannelRecoverySender_OneChannelFails_StillCallsTheOther(), TestDualChannelRecoverySender_SendsBothChannels(), DualChannelRecoverySender (+3 more)

### Community 112 - "sessionStore.ts"
Cohesion: 0.18
Nodes (13): fetchSessionContext(), SessionContext, SessionContextOutcome, refreshFromServer(), reportUnauthorized(), setAuthenticated(), setConnectionLost(), setUnauthenticated() (+5 more)

### Community 113 - "Implementar HU-007: defensa escalonada contra abuso en el acceso"
Cohesion: 0.12
Nodes (16): Alcance incluido, Bloqueos resueltos: `CT-005` → `DEC-061`, `DP-SEG-10` → `DEC-062`, Criterios de aceptación y cierre, `CT-005` · La sexta solicitud exige el reto, no la quinta, Documentación y trazabilidad, `DP-SEG-10` · Reto telefónico completo, Estado existente que debes conservar, Fuera de alcance (+8 more)

### Community 114 - "Implementar HU-012: cascarón del panel privado"
Cohesion: 0.12
Nodes (16): Alcance incluido, Bloqueo resuelto: `DP-SEG-09` → `DEC-060`, Criterios de aceptación y cierre, Defecto conocido, no introducido por esta HU: cookie `Secure` no persiste en WebKit sobre `http://localhost`, Documentación y trazabilidad, Estado existente que debes conservar, Evidencia real (ejecución del 2026-08-17), Fuera de alcance (+8 more)

### Community 115 - "newResendSenderAgainstTestServer"
Cohesion: 0.25
Nodes (11): Client, Context, NewResendEmailSender(), Server, T, newResendSenderAgainstTestServer(), TestResendEmailSender_Send_ErrorStatus_ReturnsError(), TestResendEmailSender_Send_PostsExactCodeToRecipient() (+3 more)

### Community 116 - "LoginPage.test.ts"
Cohesion: 0.29
Nodes (6): axeOptions, buildRouter(), loginMock, mountPage(), requestChallengeMock, verifyChallengeMock

### Community 117 - "LoginPage.vue"
Cohesion: 0.15
Nodes (16): login(), LoginCredentials, LoginOutcome, RFC-9457, LoginScreenState, attemptLogin(), email, isSubmitting (+8 more)

### Community 118 - "dto.go"
Cohesion: 0.15
Nodes (14): Time, validateLoginRequest(), BarbershopSummary, ChallengeAcceptedResponse, ChallengeRequest, ChallengeVerifyRequest, LoginRequest, LoginResponse (+6 more)

### Community 119 - "Internal"
Cohesion: 0.18
Nodes (11): Error, Kind, errInvalidChallenge(), Context, ErrInvalidSession(), ChallengeRequired(), IdempotencyConflict(), IdempotencyLocked() (+3 more)

### Community 120 - "As"
Cohesion: 0.41
Nodes (11): As(), T, TestAs_ReturnsFalseForForeignError(), TestAs_UnwrapsWrappedError(), TestIdempotencyConflict_CarriesSafeMessage(), TestIdempotencyLocked_CarriesSafeMessage(), TestInternal_WrapsCauseWithoutExposingIt(), TestInvalid_CarriesSafeMessage() (+3 more)

### Community 121 - "BaseBadge.vue"
Cohesion: 0.18
Nodes (10): borderVar, classes, dotColorVar, emit, handleDismiss(), Props, style, surfaceVar (+2 more)

### Community 122 - "echoBodyHandler"
Cohesion: 0.24
Nodes (9): BodyLimit(), Handler, echoBodyHandler(), Request, ResponseWriter, T, TestBodyLimit_AllowsBodyWithinLimit(), TestBodyLimit_ProducesRecognizableMaxBytesError() (+1 more)

### Community 123 - "fakeRepository"
Cohesion: 0.31
Nodes (6): Context, Time, createSessionCall, fakeRepository, lookupCall, resolveCall

### Community 124 - "resetForFreshLogin"
Cohesion: 0.18
Nodes (9): loggingOut, onLogout(), router, items, NavItem, onRetry(), resetForFreshLogin(), retryBootstrap() (+1 more)

### Community 125 - "TestContract_TranslateAlwaysFillsRequiredFields"
Cohesion: 0.54
Nodes (7): findRepoRoot(), T, isZeroValue(), loadProblemSchema(), TestContract_ProblemStructMatchesOpenAPISchema(), TestContract_TranslateAlwaysFillsRequiredFields(), problemSchema

### Community 126 - "BaseButton.vue"
Cohesion: 0.29
Nodes (5): classes, emit, handleClick(), Props, axeOptions

### Community 127 - "requireSession.ts"
Cohesion: 0.48
Nodes (5): requireSession(), CYCLE_PATHS, isSafeInternalRedirect(), ensureBootstrapped(), sessionState

### Community 128 - "contract_session_test.go"
Cohesion: 0.47
Nodes (5): T, TestContract_OpenAPIYAML_RegistersSessionContextPath(), TestContract_SessionContextOperation_MethodPathSecurityAndResponses(), sessionContextOperation, sessionContextPathFile

### Community 129 - "routes.test.ts"
Cohesion: 0.29
Nodes (5): installSessionHandling(), onUnauthorized(), authRoutes, authenticated, fetchSessionContextMock

### Community 131 - "web/package.json"
Cohesion: 0.40
Nodes (4): name, private, type, version

### Community 133 - "PrivateShell.test.ts"
Cohesion: 0.40
Nodes (3): mockState, retryBootstrapMock, BootstrapState

### Community 136 - "AppHeader.test.ts"
Cohesion: 0.67
Nodes (3): buildRouter(), mountHeader(), postMock

### Community 149 - "RecoveryRequestStep.vue"
Cohesion: 0.13
Nodes (16): attemptedSubmit, email, emit, fieldError, handleEmailInput(), isSubmitting, onSubmit(), status (+8 more)

### Community 150 - "DB"
Cohesion: 0.15
Nodes (10): Context, Repository, Time, New(), Context, Pool, NewDB(), ValidBarbershopID() (+2 more)

### Community 151 - ".Begin"
Cohesion: 0.17
Nodes (12): Context, Duration, parseOutcome(), Queries, Coordinator, Decision, Fingerprint, Key (+4 more)

### Community 152 - "idempotency/idempotency_test.go"
Cohesion: 0.23
Nodes (19): ComputeFingerprint(), ParseKey(), contains(), T, indexOf(), TestComputeFingerprint_DifferentContentProducesDifferentFingerprint(), TestComputeFingerprint_IsDeterministic(), TestComputeFingerprint_MatchesStoredFormat() (+11 more)

### Community 153 - "newLoginHandlerWithThrottle"
Cohesion: 0.26
Nodes (9): Context, Repository, T, Time, newLoginHandlerWithThrottle(), TestLoginHandler_Escalated_Returns429WithRetryAfterAndNeverTouchesRepository(), TestLoginHandler_NotEscalated_ProceedsNormally(), countingRepository (+1 more)

### Community 154 - "setupTestDB"
Cohesion: 0.51
Nodes (16): NewSQLCoordinator(), fingerprintOf(), T, setupTestDB(), TestAbort_ReleasesInProgressClaim_ButNeverCompleted(), TestBegin_AbandonedClaim_ReportsConflictInProgress(), TestBegin_CrossTenantIsolation(), TestBegin_DifferentFingerprint_ConflictsWithoutEffect() (+8 more)

### Community 155 - "RecoveryVerifyStep.vue"
Cohesion: 0.16
Nodes (16): requestRecovery(), verifyRecovery(), attemptedSubmit, canResend, code, emit, fieldError, handleCodeInput() (+8 more)

### Community 156 - "RecoveryResetStep.vue"
Cohesion: 0.18
Nodes (15): resetRecoveryPassword(), attemptedSubmit, confirmPassword, emit, fieldErrors, handleConfirmPasswordInput(), handleNewPasswordInput(), isSubmitting (+7 more)

### Community 159 - "RecoveryPage.vue"
Cohesion: 0.14
Nodes (9): email, headingRef, maskedEmail, maskedPhone, resetToken, router, Step, STEP_NUMBERS (+1 more)

### Community 160 - "Context"
Cohesion: 0.26
Nodes (6): Context, Time, Credential, failingRepository, stubClock, stubRepository

### Community 162 - "Implementar HU-011: pantalla de recuperación de acceso"
Cohesion: 0.15
Nodes (12): Alcance incluido, Documentación y trazabilidad, Estado existente que debe conservarse, Fuera de alcance, Git y PR, Implementar HU-011: pantalla de recuperación de acceso, Instrucción para Claude o Codex, Objetivo (+4 more)

### Community 166 - "ThrottleService"
Cohesion: 0.27
Nodes (9): Repository, NewLoginService(), Context, LoginService, PasswordHasher, Repository, ThrottleService, TokenGenerator (+1 more)

### Community 169 - ".ValidateAndRenewSession"
Cohesion: 0.48
Nodes (3): Context, Repository, Time

### Community 172 - "LoginForm.test.ts"
Cohesion: 0.33
Nodes (5): LoginServerErrorSummary, axeOptions, MountFormOptions, RouterLinkStub, RFC-9457

### Community 173 - "RecoveryPage.test.ts"
Cohesion: 0.33
Nodes (6): axeOptions, buildRouter(), mountPage(), requestRecoveryMock, resetRecoveryPasswordMock, verifyRecoveryMock

### Community 187 - "recoveryApi.ts"
Cohesion: 0.70
Nodes (3): RecoveryRequestOutcome, RecoveryResetPasswordOutcome, RecoveryVerifyOutcome

### Community 190 - "RecoveryVerifyStep.test.ts"
Cohesion: 0.40
Nodes (3): axeOptions, requestRecoveryMock, verifyRecoveryMock

## Knowledge Gaps
- **539 isolated node(s):** `system-barbershop`, `schemaDoc`, `ChallengeRequest`, `ChallengeAcceptedResponse`, `ChallengeVerifyRequest` (+534 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **31 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `buildRouter()` connect `buildRouter` to `NewRouter`, `ThrottleService`, `Resolve`, `RecoveryService`, `whatsapp_meta.go`, `NewSessionService`, `NewDualChannelRecoverySender`, `newResendSenderAgainstTestServer`, `DB`, `NewPhoneChallengeService`, `Load`, `assertProblem`, `NewThrottleService`, `setupTestDB`, `recovery_test.go`?**
  _High betweenness centrality (0.101) - this node is a cross-community bridge._
- **Why does `DB` connect `DB` to `buildRouter`, `BarbershopID`, `setupTestDB`, `httpserver/idempotency_test.go`, `setupTestDB`, `httpapi/contract_test.go`?**
  _High betweenness centrality (0.056) - this node is a cross-community bridge._
- **Why does `NewRouter()` connect `NewRouter` to `buildRouter`, `RequestID`, `WriteProblem`, `Translate`, `Internal`, `assertProblem`?**
  _High betweenness centrality (0.031) - this node is a cross-community bridge._
- **Are the 44 inferred relationships involving `buildRouter()` (e.g. with `NewChallengeHandler()` and `NewChallengeVerifyHandler()`) actually correct?**
  _`buildRouter()` has 44 INFERRED edges - model-reasoned connections that need verification._
- **Are the 15 inferred relationships involving `Reglas de negocio` (e.g. with `Reprogramación o cancelación rápida por emergencia` and `Persistencia de historial de bloqueos`) actually correct?**
  _`Reglas de negocio` has 15 INFERRED edges - model-reasoned connections that need verification._
- **Are the 43 inferred relationships involving `As()` (e.g. with `TestPhoneChallengeService_Request_CodeGeneratorFails_ReturnsInternalWithoutCallingRepository()` and `TestPhoneChallengeService_Request_SenderFails_ReturnsInternalButStillRegistered()`) actually correct?**
  _`As()` has 43 INFERRED edges - model-reasoned connections that need verification._
- **Are the 32 inferred relationships involving `setupTestDB()` (e.g. with `TestPhoneChallengeRepository_RequestChallenge_NonEscalatedIP_NotAccepted()` and `TestPhoneChallengeRepository_RequestChallenge_ResendCooldown_RejectsImmediate()`) actually correct?**
  _`setupTestDB()` has 32 INFERRED edges - model-reasoned connections that need verification._