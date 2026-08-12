---
prompt_id: "PROMPT-HU-003-v1"
version: "1.0"
kind: "hu"
status: "draft"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-003"
related_hu: []
issue: "pending"
issue_url: null
suggested_issue_title: "feat(api): implementar HU-003 contrato HTTP base y errores uniformes"
branch: null
pr: null
pr_url: null
depends_on:
  - "HU-001 integrada"
  - "HU-002 integrada"
rules:
  - "RN-DAT-02"
  - "RN-TEN-01"
decisions:
  - "DEC-034"
  - "DEC-035"
  - "DEC-037"
  - "DEC-038"
acceptance_criteria:
  - "CA-003-01"
  - "CA-003-02"
  - "CA-003-03"
  - "CA-003-04"
  - "CA-003-05"
  - "CA-003-06"
  - "CA-003-07"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/prioridades.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/04-arquitectura/backend-go.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/06-api/estandar-openapi.md"
  - "api/openapi/openapi.yaml"
created_at: "2026-08-12"
updated_at: "2026-08-12"
supersedes: "docs/10-backlog/prompts-implementacion.md#prompt-de-hu-003--contrato-http-base-y-registro-sin-datos-personales"
superseded_by: null
---

# Implementar HU-003: contrato HTTP base y errores uniformes

## Instrucción para Claude o Codex

Implementa únicamente `HU-003` de extremo a extremo. Trabaja hasta dejar un pull request borrador verificable, sin implementar autenticación, idempotencia ni endpoints de negocio futuros. Las fuentes normativas prevalecen sobre este prompt; si detectas una contradicción, regístrala antes de codificar.

## Objetivo

Entregar la base HTTP del API en Chi v5 con las audiencias `/api/v1/public`, `/api/v1/customer` y `/api/v1/private`, errores RFC 9457 uniformes, correlación de solicitudes, recuperación segura de pánicos y registros estructurados que no contengan datos personales ni secretos. OpenAPI, handlers, pruebas y documentación deben coincidir.

## Preflight obligatorio

1. Comprueba que no sobrescribirás cambios ajenos y que `main` está actualizada por fast-forward.
2. Si existe `graphify-out/graph.json`, consulta el contexto de `HU-003`, contrato HTTP, middleware y logging con Graphify antes de navegar archivos manualmente.
3. Lee completamente cada archivo de `source_docs` y confirma que `HU-001` y `HU-002` están integradas y sus pruebas pasan.
4. Busca un issue abierto que cubra exactamente `HU-003`. Si no existe, crea uno con el título sugerido, criterios, pruebas y exclusiones de esta historia.
5. Actualiza en este archivo y en el índice del catálogo el número y URL reales del issue; cambia el estado a `ready`. No ejecutes este prompt mientras siga `issue: pending`.
6. Crea desde `main` la rama `feat/<issue>-hu003-contrato-http` y cambia el estado a `in_progress` con la rama real.

## Alcance incluido

- Router Chi v5 y los tres grupos de audiencia, aunque todavía no contengan operaciones de negocio.
- Orden de middleware definido en la arquitectura, incluida correlación y recuperación de pánico.
- Representación `application/problem+json` conforme a RFC 9457 con `type`, `title`, `status`, `detail`, `instance` e identificador de correlación.
- Cabecera de correlación estable en entrada, respuesta y cada registro relacionado.
- Logger estructurado con nivel configurable y referencias opacas.
- Traducción central de errores, incluido `404` para recursos de otro tenant.
- Componentes, cabeceras, respuestas y ejemplos de error en OpenAPI 3.1.2.
- Pruebas de unidad, HTTP y contrato exigidas por la HU.

## Fuera de alcance

- Autenticación, sesiones, recuperación de acceso o rate limiting.
- Idempotencia de `HU-004`.
- Migraciones o cambios de esquema no indispensables para esta HU.
- Endpoints ficticios o de negocio para demostrar la infraestructura.
- Datos personales reales en pruebas, fixtures, mensajes o evidencia.

## Estado que debes conservar

- El dominio y los servicios Go permanecen independientes de Chi, `net/http` y PostgreSQL.
- El acceso tenant-aware de `HU-002` sigue siendo el único camino permitido para operaciones de datos.
- Las migraciones aplicadas son inmutables.
- OpenAPI es contract-first: cambia primero el contrato y después adapta implementación y pruebas.

## Trabajo requerido

1. Audita el módulo Go existente y reutiliza paquetes y convenciones presentes; no crees carpetas globales `controllers`, `models`, `services`, `utils` o `helpers`.
2. Define tipos de error de aplicación fuera de la capa HTTP y una traducción explícita a problemas HTTP. El dominio no debe importar Chi.
3. Monta el router y middleware en el orden normativo. Acepta una cabecera de correlación válida o genera un valor opaco; nunca reflejes contenido arbitrario sin validación.
4. Centraliza la escritura de problemas: tipo de contenido correcto, estado coherente, instancia segura y detalle accionable que no revele SQL, rutas internas, nombres de proveedores, excepciones ni versiones.
5. Recupera pánicos en el límite HTTP, devuelve el problema `500`, registra la correlación sin datos sensibles y permite que el proceso continúe.
6. Diseña el logger con una lista permitida de campos. Prohíbe nombre, teléfono, correo, contraseña, tokens, cookies, cabeceras de autorización y contenido libre de mensajes.
7. Para recursos ajenos, traduce la ausencia por aislamiento a `404`, sin distinguir entre inexistente y perteneciente a otro tenant.
8. Actualiza OpenAPI con esquemas reutilizables, cabecera de correlación y ejemplos para cada respuesta de error declarada. No mantengas DTO manual paralelo si el proyecto deriva tipos del bundle.
9. Incorpora en CI las validaciones `openapi:lint` y `openapi:bundle` si aún no existen.

## Pruebas y evidencia

- `httptest` para rutas base, orden observable de middleware, correlación recibida/generada, formato y `Content-Type` de problemas.
- Pánico controlado que produzca `500`, correlación y registro seguro sin derribar el servidor de prueba.
- Error interno que pruebe que no se filtran consultas, rutas, proveedores, versiones ni nombres de tipos técnicos.
- Recurso de otro tenant que responda exactamente igual que uno inexistente: `404` y cuerpo no revelador.
- Captura del logger recorrida por patrones de correo, teléfono, token, contraseña, cookie y autorización; la prueba debe fallar ante un campo prohibido.
- Prueba estructural que impida importaciones de Chi desde dominio o servicios.
- Prueba de contrato contra el bundle OpenAPI y ejemplos válidos para todas las respuestas de error.

Entrega al final una tabla `Criterio | Estado | Prueba o evidencia` para `CA-003-01` a `CA-003-07`. No declares cumplido un criterio sin prueba.

## Documentación y trazabilidad

- Actualiza únicamente la documentación realmente afectada: README del API, contrato OpenAPI, matriz de trazabilidad e historial.
- Registra cualquier nueva decisión antes de usarla como supuesto.
- Actualiza este frontmatter y el índice del catálogo con issue, rama, PR y estado reales.
- Si el contenido del prompt necesita un cambio material una vez iniciada la ejecución, crea una versión nueva y enlaza `supersedes`; no reescribas silenciosamente esta versión.

## Verificación final

Ejecuta los comandos equivalentes definidos por el repositorio para:

```text
formato y análisis estático de Go
pruebas unitarias e HTTP de apps/api
pruebas de integración aplicables con PostgreSQL real y dos tenants
openapi:lint
openapi:bundle
validación del contrato contra los handlers
git diff --check
graphify update .
```

Conserva la salida o resumen preciso como evidencia del PR. No reduzcas umbrales ni omitas pruebas fallidas.

## Git y PR

- Usa commits y título `feat(api): implementa contrato HTTP base` o una variante Conventional Commits equivalente.
- Publica la rama y abre un PR borrador contra `main` con `Closes #<issue>` únicamente si cubre todo el issue.
- El PR debe explicar alcance, exclusiones, contrato, pruebas, riesgos, recuperación y evidencia de ausencia de datos sensibles.
- No hagas merge, push directo, force push ni reescribas `main`.
