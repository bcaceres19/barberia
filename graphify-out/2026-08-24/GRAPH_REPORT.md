# Graph Report - agent-a6a8806684c76a6e0  (2026-08-23)

## Corpus Check
- 376 files · ~487,895 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2625 nodes · 5384 edges · 206 communities (171 shown, 35 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 987 edges (avg confidence: 0.79)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `65df8fb2`
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
- recovery_handler_test.go
- graphify SKILL.md
- tsconfig.vitest.json
- .Now
- .prettierrc.json
- router/index.ts
- tsconfig.json
- apps/web index.html
- apps/web/e2e README
- loginApi.test.ts
- LoginForm.vue
- settings/index.ts
- staff/index.ts
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
- Key
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
- Principal
- NewThrottleService
- setupTestDB
- staff/httpapi/contract_test.go
- recovery_test.go
- Implementar HU-005: inicio de sesión seguro del barbero
- typescript
- Implementar HU-006: sesión persistente y cierre de sesión
- Implementar HU-010: pantalla de acceso
- vite
- buildRouter
- NewService
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
- As
- StaffPage.vue
- BaseBadge.vue
- echoBodyHandler
- SettingsPage.vue
- AppNav.vue
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
- PhoneChallengeService
- panel-evidencia-responsiva.spec.ts
- challenge_handler_test.go
- openapi-typescript
- Implementar HU-020: configuración básica de la barbería
- Implementar HU-021: registro y listado de barberos
- .ServeHTTP
- Validation
- Context
- setupWorkerTestDB
- RecoveryRequestStep.vue
- DB
- run
- setupTestDB
- NewCapturingPhoneCodeSender
- contract_recovery_test.go
- RecoveryVerifyStep.vue
- RecoveryResetStep.vue
- sessionContextOutcome.ts
- Repository
- RecoveryPage.vue
- contract_challenge_test.go
- settingsApi.test.ts
- Implementar HU-011: pantalla de recuperación de acceso
- SettingsPage.test.ts
- BaseDialog.test.ts
- configuracion-barberia-evidencia-responsiva.spec.ts
- formatInstant.ts
- navItem.ts
- .ValidateAndRenewSession
- staff/httpapi/handler_test.go
- NewService
- LoginForm.test.ts
- RecoveryPage.test.ts
- Fingerprint
- httpserver/idempotency_test.go
- assertProblem
- newSessionMiddleware
- RecoveryService
- newLoginHandlerWithThrottle
- DecodeCursor
- StaffPage.test.ts
- ThrottleService
- Service
- Context
- NewArgon2Hasher
- newLogoutHandler
- recoveryApi.ts
- recoveryApi.test.ts
- RecoveryResetStep.test.ts
- RecoveryVerifyStep.test.ts
- newBarberResponse
- Context
- MaskEmail
- staffApi.test.ts
- .Login
- recuperacion.spec.ts
- spySessionRepo
- PhoneChallengeConfig
- Repository
- NewCapturingRecoveryCodeSender
- resetForFreshLogin
- URLParam
- barberos-evidencia-responsiva.spec.ts
- stubHasher

## God Nodes (most connected - your core abstractions)
1. `buildRouter()` - 71 edges
2. `As()` - 56 edges
3. `Reglas de negocio` - 49 edges
4. `NewService()` - 45 edges
5. `Respuestas a dudas pendientes del proyecto` - 43 edges
6. `setupTestDB()` - 42 edges
7. `DB` - 42 edges
8. `Matriz de trazabilidad` - 42 edges
9. `Historial de cambios documentales` - 40 edges
10. `Translate()` - 36 edges

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

## Communities (206 total, 35 thin omitted)

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

### Community 12 - "recovery_handler_test.go"
Cohesion: 0.18
Nodes (26): Logger, NewRecoveryRequestHandler(), NewRecoveryResetPasswordHandler(), NewRecoveryVerifyHandler(), discardRecoveryLogger(), Logger, RecoveryRepository, T (+18 more)

### Community 13 - "graphify SKILL.md"
Cohesion: 0.18
Nodes (11): CLAUDE.md (root) — reglas graphify, .claude/CLAUDE.md — trigger de graphify, graphify reference: add & watch, graphify reference: exports, graphify reference: extraction spec, graphify reference: GitHub clone & merge, graphify reference: hooks & CLAUDE.md integration, graphify reference: query/path/explain (+3 more)

### Community 14 - "tsconfig.vitest.json"
Cohesion: 0.20
Nodes (9): compilerOptions, tsBuildInfoFile, types, extends, include, node, src/**/*.spec.ts, jsdom (+1 more)

### Community 15 - ".Now"
Cohesion: 0.08
Nodes (49): Context, T, Time, newTestService(), requireInternal(), requireUnauthorized(), TestLogin_CancelledContext_NeverCallsRepository(), TestLogin_CreateSessionError_ReturnsInternal() (+41 more)

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
Cohesion: 0.11
Nodes (18): classes, close(), closeButtonRef, dialogRef, emit, focusableElementsRef, handleBackdropClick(), handleCloseClick() (+10 more)

### Community 60 - "Key"
Cohesion: 0.09
Nodes (67): createBarber(), Barber, Repository, T, newRepository(), setupTestDB(), TestCreate_DifferentKeysSameName_CreatesTwoDistinctBarbers(), TestCreate_FirstExecution_PersistsAndReturnsProceed() (+59 more)

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
Cohesion: 0.15
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
Cohesion: 0.26
Nodes (17): NewPhoneChallengeService(), T, testChallengeCfg(), TestCryptoPhoneCodeGenerator_ProducesSixDigits(), TestPhoneChallengeService_Request_Accepted_SendsCode(), TestPhoneChallengeService_Request_AcceptedButEmptyPhone_NeverSends(), TestPhoneChallengeService_Request_CodeGeneratorFails_ReturnsInternalWithoutCallingRepository(), TestPhoneChallengeService_Request_NotAccepted_NeverSends() (+9 more)

### Community 91 - "Principal"
Cohesion: 0.16
Nodes (18): Handler, validSessionCookieShape(), NewSessionContextHandler(), T, newSessionContextHandler(), TestSessionContextHandler_MissingPrincipalInContext_ReturnsSafe500(), TestSessionContextHandler_RepositoryFailure_ReturnsSafe500(), TestSessionContextHandler_Success_ReturnsMinimalPayload() (+10 more)

### Community 92 - "NewThrottleService"
Cohesion: 0.20
Nodes (17): Context, Time, NewThrottleService(), Context, T, Time, testThrottleCfg(), TestThrottleService_BelowThreshold_ReturnsNil() (+9 more)

### Community 93 - "setupTestDB"
Cohesion: 0.07
Nodes (73): HMACHex(), Context, NewPhoneChallengeRepository(), T, testChallengeCfg(), testCodeHash(), TestPhoneChallengeRepository_RequestChallenge_NonEscalatedIP_NotAccepted(), TestPhoneChallengeRepository_RequestChallenge_ResendCooldown_RejectsImmediate() (+65 more)

### Community 94 - "staff/httpapi/contract_test.go"
Cohesion: 0.08
Nodes (43): findRepoRoot(), operation, schemaDoc, T, loadYAML(), TestContract_CookieNameMatchesSecuritySchemeReference(), TestContract_LoginOperation_MethodPathSecurityAndResponses(), TestContract_LoginRequestSchema_MatchesDTOFields() (+35 more)

### Community 95 - "recovery_test.go"
Cohesion: 0.20
Nodes (18): Context, T, mustHash(), newTestRecoveryService(), TestRecoveryService_ChangePassword_InvalidToken_NeverHashesOrWrites(), TestRecoveryService_ChangePassword_RaceLosesToken_ReturnsUniformError(), TestRecoveryService_ChangePassword_SameAsCurrent_Rejected(), TestRecoveryService_ChangePassword_Success() (+10 more)

### Community 96 - "Implementar HU-005: inicio de sesión seguro del barbero"
Cohesion: 0.15
Nodes (13): Alcance incluido, Documentación y trazabilidad, Estado existente que debes conservar, Fuera de alcance, Git y PR, Implementar HU-005: inicio de sesión seguro del barbero, Instrucción para Claude o Codex, Objetivo (+5 more)

### Community 98 - "Implementar HU-006: sesión persistente y cierre de sesión"
Cohesion: 0.15
Nodes (12): Alcance incluido, Documentación y trazabilidad, Estado existente que debes conservar, Fuera de alcance, Git y PR, Implementar HU-006: sesión persistente y cierre de sesión, Instrucción para Claude o Codex, Objetivo (+4 more)

### Community 99 - "Implementar HU-010: pantalla de acceso"
Cohesion: 0.15
Nodes (12): Alcance incluido, Documentación y trazabilidad, Estado existente que debes conservar, Fuera de alcance, Git y PR, Implementar HU-010: pantalla de acceso, Instrucción para Claude o Codex, Objetivo (+4 more)

### Community 101 - "buildRouter"
Cohesion: 0.07
Nodes (82): barberBody, barberListBody, barbershopSettingsBody, recoveryCapture, buildRouter(), Logger, Mux, main() (+74 more)

### Community 102 - "NewService"
Cohesion: 0.11
Nodes (46): Service, NewGetBarbershopSettingsHandler(), NewUpdateBarbershopSettingsHandler(), Context, fakeRepository, Request, T, requestWithPrincipal() (+38 more)

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
Cohesion: 0.20
Nodes (15): Request, ResponseWriter, Request, ResponseWriter, Request, ResponseWriter, Request, ResponseWriter (+7 more)

### Community 108 - "whatsapp_meta.go"
Cohesion: 0.15
Nodes (18): Client, Context, NewMetaWhatsAppSender(), Request, Server, T, newMetaSenderAgainstTestServer(), TestMetaWhatsAppSender_Send_ErrorStatus_ReturnsErrorWithoutLeakingBody() (+10 more)

### Community 109 - "NewPurgeService"
Cohesion: 0.19
Nodes (13): Context, NewPurgeService(), Context, T, TestPurgeService_CancelledContext_NeverCallsRepository(), TestPurgeService_LoginThrottleError_SkipsRest(), TestPurgeService_PhoneChallengeError_SkipsRecoveryButKeepsLoginThrottleCount(), TestPurgeService_PurgeOnce_ReturnsAllThreeCounts() (+5 more)

### Community 110 - "Implementar HU-008: recuperación de acceso con código"
Cohesion: 0.11
Nodes (18): Alcance incluido, Bloqueos resueltos: `DP-SEG-11` → `DEC-063`, `DP-SEG-12` → `DEC-064`, `CT-006` → `DEC-065`, Criterios de aceptación y cierre, `CT-006` · No enumeración frente a destino enmascarado, Documentación y trazabilidad, `DP-NOT-05` · Proveedor oficial concreto, `DP-SEG-11` · Política de contraseña nueva, `DP-SEG-12` · Parámetros y autorización del flujo (+10 more)

### Community 111 - "NewDualChannelRecoverySender"
Cohesion: 0.19
Nodes (11): Context, NewDualChannelRecoverySender(), Context, T, TestDualChannelRecoverySender_BothChannelsFail_ReturnsCombinedError(), TestDualChannelRecoverySender_OneChannelFails_StillCallsTheOther(), TestDualChannelRecoverySender_SendsBothChannels(), DualChannelRecoverySender (+3 more)

### Community 112 - "sessionStore.ts"
Cohesion: 0.24
Nodes (11): fetchSessionContext(), installSessionHandling(), onUnauthorized(), refreshFromServer(), reportUnauthorized(), setAuthenticated(), setConnectionLost(), setUnauthenticated() (+3 more)

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
Cohesion: 0.13
Nodes (18): login(), LoginCredentials, LoginOutcome, RFC-9457, LoginScreenState, CYCLE_PATHS, isSafeInternalRedirect(), attemptLogin() (+10 more)

### Community 118 - "dto.go"
Cohesion: 0.15
Nodes (14): Time, validateLoginRequest(), BarbershopSummary, ChallengeAcceptedResponse, ChallengeRequest, ChallengeVerifyRequest, LoginRequest, LoginResponse (+6 more)

### Community 119 - "As"
Cohesion: 0.19
Nodes (19): Error, Kind, As(), ChallengeRequired(), IdempotencyConflict(), IdempotencyLocked(), Internal(), NotFound() (+11 more)

### Community 120 - "StaffPage.vue"
Cohesion: 0.07
Nodes (38): createBarber(), fetchBarbers(), renameBarber(), toBarber(), toPage(), Barber, BarberPage, newIdempotencyKey() (+30 more)

### Community 121 - "BaseBadge.vue"
Cohesion: 0.18
Nodes (10): borderVar, classes, dotColorVar, emit, handleDismiss(), Props, style, surfaceVar (+2 more)

### Community 122 - "echoBodyHandler"
Cohesion: 0.24
Nodes (9): BodyLimit(), Handler, echoBodyHandler(), Request, ResponseWriter, T, TestBodyLimit_AllowsBodyWithinLimit(), TestBodyLimit_ProducesRecognizableMaxBytesError() (+1 more)

### Community 123 - "SettingsPage.vue"
Cohesion: 0.13
Nodes (26): fetchBarbershopSettings(), saveBarbershopSettings(), toSettings(), BarbershopSettings, BarbershopSettingsFormValues, toFormValues(), FetchBarbershopSettingsOutcome, SaveBarbershopSettingsOutcome (+18 more)

### Community 124 - "AppNav.vue"
Cohesion: 0.33
Nodes (3): baseItems, items, props

### Community 125 - "TestContract_TranslateAlwaysFillsRequiredFields"
Cohesion: 0.54
Nodes (7): findRepoRoot(), T, isZeroValue(), loadProblemSchema(), TestContract_ProblemStructMatchesOpenAPISchema(), TestContract_TranslateAlwaysFillsRequiredFields(), problemSchema

### Community 126 - "BaseButton.vue"
Cohesion: 0.29
Nodes (5): classes, emit, handleClick(), Props, axeOptions

### Community 127 - "requireSession.ts"
Cohesion: 0.36
Nodes (6): requireSession(), onRetry(), ensureBootstrapped(), retryBootstrap(), sessionState, setChecking()

### Community 128 - "contract_session_test.go"
Cohesion: 0.47
Nodes (5): T, TestContract_OpenAPIYAML_RegistersSessionContextPath(), TestContract_SessionContextOperation_MethodPathSecurityAndResponses(), sessionContextOperation, sessionContextPathFile

### Community 129 - "routes.test.ts"
Cohesion: 0.36
Nodes (7): updateBarbershopName(), authRoutes, privateShellChildRoutes, privateShellRoute(), authenticated, buildRouter(), fetchSessionContextMock

### Community 131 - "web/package.json"
Cohesion: 0.40
Nodes (4): name, private, type, version

### Community 133 - "PrivateShell.test.ts"
Cohesion: 0.40
Nodes (3): mockState, retryBootstrapMock, BootstrapState

### Community 136 - "AppHeader.test.ts"
Cohesion: 0.67
Nodes (3): buildRouter(), mountHeader(), postMock

### Community 137 - "PhoneChallengeService"
Cohesion: 0.29
Nodes (7): errInvalidChallenge(), Context, NewCryptoPhoneCodeGenerator(), CryptoPhoneCodeGenerator, PhoneChallengeRepository, PhoneChallengeService, PhoneCodeGenerator

### Community 140 - "challenge_handler_test.go"
Cohesion: 0.27
Nodes (16): T, newChallengeHandler(), newVerifyHandler(), testChallengeCfg(), TestChallengeHandler_Accepted_Returns202WithGenericMessage(), TestChallengeHandler_AcceptedVsNotAccepted_IdenticalResponse(), TestChallengeHandler_MissingEmail_Returns422(), TestChallengeHandler_NotAccepted_StillReturns202() (+8 more)

### Community 143 - "Implementar HU-020: configuración básica de la barbería"
Cohesion: 0.11
Nodes (17): 1. Contrato HTTP antes del código, 2. Migración y persistencia, 3. Backend Go, 4. Frontend Vue, Alcance incluido, Documentación y trazabilidad, Ejecución real (2026-08-23), Estado existente que debes conservar (+9 more)

### Community 144 - "Implementar HU-021: registro y listado de barberos"
Cohesion: 0.12
Nodes (16): 1. Contrato HTTP antes del código, 2. Migración y seguridad de datos, 3. Backend Go, 4. Frontend Vue, Alcance incluido, Documentación y trazabilidad, Estado existente que debes conservar, Fuera de alcance (+8 more)

### Community 145 - ".ServeHTTP"
Cohesion: 0.16
Nodes (10): Request, ResponseWriter, Request, ResponseWriter, PrincipalFromContext(), Request, ResponseWriter, newBarbershopSettingsResponse() (+2 more)

### Community 146 - "Validation"
Cohesion: 0.28
Nodes (12): NormalizeContact(), NormalizeContactEmail(), NormalizeName(), errContactEmailInvalid(), errContactPhoneInvalid(), errNameRequired(), errNameTooLong(), errTimezoneInvalid() (+4 more)

### Community 147 - "Context"
Cohesion: 0.29
Nodes (4): Context, Time, stubPhoneChallengeRepository, stubThrottleRepository

### Community 148 - "setupWorkerTestDB"
Cohesion: 0.29
Nodes (8): Context, NewPurgeRepository(), T, setupWorkerTestDB(), TestPurgeRepository_PurgeLoginThrottle_DeletesOnlyExpired(), TestPurgeRepository_PurgePhoneChallenges_RunsWithoutError(), TestPurgeRepository_PurgeRecoveryCodes_RunsWithoutError(), PurgeRepository

### Community 149 - "RecoveryRequestStep.vue"
Cohesion: 0.13
Nodes (16): attemptedSubmit, email, emit, fieldError, handleEmailInput(), isSubmitting, onSubmit(), status (+8 more)

### Community 150 - "DB"
Cohesion: 0.24
Nodes (6): Context, Pool, NewDB(), ValidBarbershopID(), DB, Row

### Community 151 - "run"
Cohesion: 0.20
Nodes (9): main(), run(), Handler, Server, New(), Logger, NewLogger(), parseLevel() (+1 more)

### Community 152 - "setupTestDB"
Cohesion: 0.51
Nodes (10): T, restoreShopA(), sameBarbershop(), samePointerValue(), setupTestDB(), TestGet_ReturnsTheFourAuthorizedFields(), TestGetAndUpdate_TenantAIsolatedFromTenantB(), TestUpdate_InvalidTimezone_WritesNothing() (+2 more)

### Community 153 - "NewCapturingPhoneCodeSender"
Cohesion: 0.29
Nodes (7): Context, NewCapturingPhoneCodeSender(), T, TestCapturingPhoneCodeSender_InnerFailure_NeverWritesCapture(), TestCapturingPhoneCodeSender_WritesPhoneAndCodeToFile(), CapturingPhoneCodeSender, PhoneCodeSender

### Community 154 - "contract_recovery_test.go"
Cohesion: 0.33
Nodes (8): operation, T, TestContract_RecoveryOperations_MethodPathSecurityAndResponses(), TestContract_RecoveryRequestSchema_MatchesDTOFields(), TestContract_RecoveryResetPasswordRequestSchema_MatchesDTOFields(), TestContract_RecoveryVerifyRequestSchema_MatchesDTOFields(), TestContract_RecoveryVerifyResponseSchema_MatchesDTOFields(), recoveryPathFile

### Community 155 - "RecoveryVerifyStep.vue"
Cohesion: 0.16
Nodes (16): requestRecovery(), verifyRecovery(), attemptedSubmit, canResend, code, emit, fieldError, handleCodeInput() (+8 more)

### Community 156 - "RecoveryResetStep.vue"
Cohesion: 0.18
Nodes (15): resetRecoveryPassword(), attemptedSubmit, confirmPassword, emit, fieldErrors, handleConfirmPasswordInput(), handleNewPasswordInput(), isSubmitting (+7 more)

### Community 157 - "sessionContextOutcome.ts"
Cohesion: 0.38
Nodes (4): SessionContext, SessionContextOutcome, authenticated, fetchSessionContextMock

### Community 158 - "Repository"
Cohesion: 0.36
Nodes (4): Context, Repository, Time, New()

### Community 159 - "RecoveryPage.vue"
Cohesion: 0.14
Nodes (9): email, headingRef, maskedEmail, maskedPhone, resetToken, router, Step, STEP_NUMBERS (+1 more)

### Community 160 - "contract_challenge_test.go"
Cohesion: 0.38
Nodes (6): operation, T, TestContract_ChallengeOperations_MethodPathSecurityAndResponses(), TestContract_ChallengeRequestSchema_MatchesDTOFields(), TestContract_ChallengeVerifyRequestSchema_MatchesDTOFields(), challengePathFile

### Community 161 - "settingsApi.test.ts"
Cohesion: 0.33
Nodes (3): getMock, patchMock, settingsBody

### Community 162 - "Implementar HU-011: pantalla de recuperación de acceso"
Cohesion: 0.17
Nodes (12): Alcance incluido, Documentación y trazabilidad, Estado existente que debe conservarse, Fuera de alcance, Git y PR, Implementar HU-011: pantalla de recuperación de acceso, Instrucción para Claude o Codex, Objetivo (+4 more)

### Community 163 - "SettingsPage.test.ts"
Cohesion: 0.33
Nodes (4): fetchMock, loadedSettings, saveMock, updateBarbershopNameMock

### Community 164 - "BaseDialog.test.ts"
Cohesion: 0.50
Nodes (3): openDialogStack, axeOptions, TwoDialogsHost

### Community 169 - ".ValidateAndRenewSession"
Cohesion: 0.48
Nodes (3): Context, Repository, Time

### Community 170 - "staff/httpapi/handler_test.go"
Cohesion: 0.16
Nodes (37): Service, NewCreateBarberHandler(), NewGetBarberHandler(), NewListBarbersHandler(), NewRenameBarberHandler(), decodeProblem(), Request, ResponseRecorder (+29 more)

### Community 171 - "NewService"
Cohesion: 0.21
Nodes (27): NewService(), Barber, Context, T, mustBeNotFound(), mustBeValidation(), TestCreate_EmptyName_RejectedWithoutTouchingRepository(), TestCreate_NameExactly120Characters_Accepted() (+19 more)

### Community 172 - "LoginForm.test.ts"
Cohesion: 0.33
Nodes (5): LoginServerErrorSummary, axeOptions, MountFormOptions, RouterLinkStub, RFC-9457

### Community 173 - "RecoveryPage.test.ts"
Cohesion: 0.33
Nodes (6): axeOptions, buildRouter(), mountPage(), requestRecoveryMock, resetRecoveryPasswordMock, verifyRecoveryMock

### Community 174 - "Fingerprint"
Cohesion: 0.13
Nodes (17): Barber, Context, fakeRepository, Barber, Barber, Context, Repository, Time (+9 more)

### Community 175 - "httpserver/idempotency_test.go"
Cohesion: 0.19
Nodes (23): Request, ResponseWriter, IdempotencyFingerprint(), IdempotencyKeyFromRequest(), doIdempotentRequest(), doIdempotentRequestObject(), Handler, HandlerFunc (+15 more)

### Community 176 - "assertProblem"
Cohesion: 0.29
Nodes (20): assertProblem(), doLogin(), Repository, Request, ResponseRecorder, T, newHandler(), readAndRestore() (+12 more)

### Community 177 - "newSessionMiddleware"
Cohesion: 0.28
Nodes (17): NewSessionMiddleware(), doPrivateRequest(), Handler, HandlerFunc, ResponseRecorder, T, newSessionMiddleware(), protectedHandler() (+9 more)

### Community 178 - "RecoveryService"
Cohesion: 0.24
Nodes (13): NormalizeEmail(), errInvalidRecoveryCode(), errInvalidResetToken(), Context, NewRecoveryService(), ValidateNewPassword(), PasswordHasher, RecoveryCodeSender (+5 more)

### Community 179 - "newLoginHandlerWithThrottle"
Cohesion: 0.19
Nodes (14): DefaultCookieConfig(), NewLoginHandler(), Context, Repository, T, Time, newLoginHandlerWithThrottle(), TestLoginHandler_Escalated_Returns429WithRetryAfterAndNeverTouchesRepository() (+6 more)

### Community 180 - "DecodeCursor"
Cohesion: 0.22
Nodes (15): DecodeCursor(), EncodeCursor(), Time, NormalizeFullName(), T, TestDecodeCursor_DoesNotAcceptAnArbitraryForgedCursor(), TestDecodeCursor_RejectsGarbageAsClientError(), TestEncodeDecodeCursor_RoundTrips() (+7 more)

### Community 181 - "StaffPage.test.ts"
Cohesion: 0.16
Nodes (12): axeOptions, createMock, fetchMock, fourBarbers, mountPage(), mountReady(), oneBarber, openDialogElement() (+4 more)

### Community 182 - "ThrottleService"
Cohesion: 0.26
Nodes (10): NewChallengeHandler(), NewChallengeVerifyHandler(), Repository, NewLoginService(), Context, LoginService, ThrottleService, Clock (+2 more)

### Community 183 - "Service"
Cohesion: 0.23
Nodes (9): LooksLikeBarberID(), errBarberNotFound(), errFullNameRequired(), errFullNameTooLong(), Barber, Context, Repository, validateFullName() (+1 more)

### Community 184 - "Context"
Cohesion: 0.26
Nodes (6): Context, Time, Credential, failingRepository, stubClock, stubRepository

### Community 185 - "NewArgon2Hasher"
Cohesion: 0.27
Nodes (9): NewArgon2Hasher(), parseEncodedHash(), T, TestArgon2Hasher_HashThenVerify_Succeeds(), TestArgon2Hasher_TwoUsersSamePassword_ProduceDistinctEncodedValues(), TestArgon2Hasher_Verify_MalformedEncodedHash_NeverPanicsOrErrors(), TestArgon2Hasher_Verify_WrongPasswordFails(), Argon2Hasher (+1 more)

### Community 186 - "newLogoutHandler"
Cohesion: 0.44
Nodes (7): NewLogoutHandler(), T, newLogoutHandler(), TestLogoutHandler_MissingPrincipalInContext_ReturnsSafe500(), TestLogoutHandler_ServiceFailure_ReturnsSafe500(), TestLogoutHandler_Success_RevokesAndClearsCookie(), LogoutHandler

### Community 187 - "recoveryApi.ts"
Cohesion: 0.70
Nodes (3): RecoveryRequestOutcome, RecoveryResetPasswordOutcome, RecoveryVerifyOutcome

### Community 190 - "RecoveryVerifyStep.test.ts"
Cohesion: 0.40
Nodes (3): axeOptions, requestRecoveryMock, verifyRecoveryMock

### Community 191 - "newBarberResponse"
Cohesion: 0.28
Nodes (8): Time, Barber, newBarberListResponse(), newBarberResponse(), BarberListResponse, BarberResponse, CreateBarberRequest, UpdateBarberRequest

### Community 192 - "Context"
Cohesion: 0.36
Nodes (3): Context, stubRecoveryRepository, stubRecoverySender

### Community 193 - "MaskEmail"
Cohesion: 0.39
Nodes (6): MaskEmail(), MaskPhone(), T, TestMaskEmail(), TestMaskEmail_NeverReturnsFullValue(), TestMaskPhone()

### Community 194 - "staffApi.test.ts"
Cohesion: 0.25
Nodes (4): barberBody, getMock, patchMock, postMock

### Community 195 - ".Login"
Cohesion: 0.29
Nodes (4): Time, errInvalidCredentials(), Context, Session

### Community 197 - "spySessionRepo"
Cohesion: 0.48
Nodes (3): Context, Time, spySessionRepo

### Community 198 - "PhoneChallengeConfig"
Cohesion: 0.38
Nodes (4): Context, fakePhoneChallengeRepository, PhoneChallengeConfig, requestChallengeCall

### Community 199 - "Repository"
Cohesion: 0.47
Nodes (3): Context, Repository, New()

### Community 200 - "NewCapturingRecoveryCodeSender"
Cohesion: 0.50
Nodes (3): Context, NewCapturingRecoveryCodeSender(), CapturingRecoveryCodeSender

### Community 201 - "resetForFreshLogin"
Cohesion: 0.50
Nodes (4): loggingOut, onLogout(), router, resetForFreshLogin()

### Community 202 - "URLParam"
Cohesion: 0.67
Nodes (3): Request, RequestWithURLParam(), URLParam()

### Community 203 - "barberos-evidencia-responsiva.spec.ts"
Cohesion: 0.50
Nodes (3): axeScriptPath, evidenceDir, viewports

## Knowledge Gaps
- **624 isolated node(s):** `system-barbershop`, `schemaDoc`, `ChallengeRequest`, `ChallengeAcceptedResponse`, `ChallengeVerifyRequest` (+619 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **35 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `buildRouter()` connect `buildRouter` to `NewRouter`, `PhoneChallengeService`, `recovery_handler_test.go`, `.Now`, `DB`, `NewCapturingPhoneCodeSender`, `staff/httpapi/handler_test.go`, `newSessionMiddleware`, `RecoveryService`, `newLoginHandlerWithThrottle`, `ThrottleService`, `NewArgon2Hasher`, `newLogoutHandler`, `Load`, `Key`, `NewCapturingRecoveryCodeSender`, `NewPhoneChallengeService`, `Principal`, `NewThrottleService`, `setupTestDB`, `NewService`, `Resolve`, `whatsapp_meta.go`, `NewDualChannelRecoverySender`, `newResendSenderAgainstTestServer`?**
  _High betweenness centrality (0.120) - this node is a cross-community bridge._
- **Why does `DB` connect `DB` to `buildRouter`, `Repository`, `Fingerprint`, `httpserver/idempotency_test.go`, `setupWorkerTestDB`, `BarbershopID`, `setupTestDB`, `staff/httpapi/contract_test.go`, `Key`, `setupTestDB`, `Repository`?**
  _High betweenness centrality (0.050) - this node is a cross-community bridge._
- **Why does `As()` connect `As` to `NewService`, `NewService`, `Translate`, `WriteProblem`, `Key`, `.Now`, `httpserver/idempotency_test.go`, `.ServeHTTP`, `DecodeCursor`, `NewPhoneChallengeService`, `echoBodyHandler`, `NewThrottleService`?**
  _High betweenness centrality (0.049) - this node is a cross-community bridge._
- **Are the 65 inferred relationships involving `buildRouter()` (e.g. with `NewChallengeHandler()` and `NewChallengeVerifyHandler()`) actually correct?**
  _`buildRouter()` has 65 INFERRED edges - model-reasoned connections that need verification._
- **Are the 54 inferred relationships involving `As()` (e.g. with `TestPhoneChallengeService_Request_CodeGeneratorFails_ReturnsInternalWithoutCallingRepository()` and `TestPhoneChallengeService_Request_SenderFails_ReturnsInternalButStillRegistered()`) actually correct?**
  _`As()` has 54 INFERRED edges - model-reasoned connections that need verification._
- **Are the 15 inferred relationships involving `Reglas de negocio` (e.g. with `Reprogramación o cancelación rápida por emergencia` and `Persistencia de historial de bloqueos`) actually correct?**
  _`Reglas de negocio` has 15 INFERRED edges - model-reasoned connections that need verification._
- **Are the 42 inferred relationships involving `NewService()` (e.g. with `TestCreateBarberHandler_ConflictFingerprint_Returns409()` and `TestCreateBarberHandler_Locked_Returns409()`) actually correct?**
  _`NewService()` has 42 INFERRED edges - model-reasoned connections that need verification._