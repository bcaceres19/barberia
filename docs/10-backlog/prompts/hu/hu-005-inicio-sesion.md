---
prompt_id: "PROMPT-HU-005-v1"
version: "1.1"
kind: "hu"
status: "ready"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-005"
related_hu:
  - "HU-006"
  - "HU-007"
  - "HU-010"
  - "HU-012"
issue: 44
issue_url: "https://github.com/bcaceres19/barberia/issues/44"
suggested_issue_title: "feat(auth): implementar HU-005 inicio de sesión seguro"
branch: null
pr: null
pr_url: null
depends_on:
  - "HU-002 integrada"
  - "HU-003 integrada"
rules:
  - "RN-TEN-01"
  - "RN-DAT-02"
decisions:
  - "DEC-024"
  - "DEC-026"
  - "DEC-034"
  - "DEC-035"
  - "DEC-036"
  - "DEC-037"
  - "DEC-038"
  - "DEC-040"
  - "DEC-050"
  - "DEC-055"
acceptance_criteria:
  - "CA-005-01"
  - "CA-005-02"
  - "CA-005-03"
  - "CA-005-04"
  - "CA-005-05"
  - "CA-005-06"
  - "CA-005-07"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/prioridades.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/04-arquitectura/backend-go.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/06-api/estandar-openapi.md"
  - "database/modelo-fisico-referencia.sql"
  - "database/README.md"
  - "apps/api/README.md"
  - "api/openapi/openapi.yaml"
created_at: "2026-08-13"
updated_at: "2026-08-13"
supersedes: "docs/10-backlog/prompts-implementacion.md#prompt-de-hu-005--inicio-de-sesión-del-barbero"
superseded_by: null
---

# Implementar HU-005: inicio de sesión seguro del barbero

## Instrucción para Claude o Codex

Implementa únicamente `HU-005` de extremo a extremo: contrato, migración, núcleo de autenticación, persistencia, HTTP y pruebas. Emite una sesión opaca revocable conforme a `DEC-050`, sin adelantar cierre de sesión, renovación deslizante completa, rate limiting, recuperación de acceso ni la pantalla Vue. Las fuentes normativas prevalecen sobre este prompt.

## Objetivo

Permitir que un barbero activo inicie sesión con correo y contraseña, reciba una cookie segura asociada a su usuario y barbería y pueda demostrar el aislamiento de una solicitud privada real. Un correo inexistente y una contraseña incorrecta deben ser indistinguibles; ninguna contraseña, dirección de correo o material de sesión puede aparecer en respuestas, registros o almacenamiento legible.

## Preflight obligatorio

1. Comprueba árbol limpio, `main` actualizada por fast-forward y ausencia de cambios ajenos que puedas sobrescribir.
2. Si existe `graphify-out/graph.json`, consulta `HU-005`, autenticación, sesiones, tenant/RLS, OpenAPI y migraciones antes de navegar archivos manualmente.
3. Lee completamente cada archivo de `source_docs`. Confirma en Git que `HU-002` y `HU-003` están integradas y que sus pruebas pasan.
4. `CT-003` está resuelta (`DEC-055`, 2026-08-13): el login vive en `/api/v1/public/auth/login`, sin middleware de autenticación. `HU-005`, `CA-006-04`, `backend-go.md` §7 y este prompt ya reflejan la resolución; no reabras la decisión ni inventes una audiencia distinta.
5. Confirma que no exista otra duda o contradicción abierta que afecte cookie, ruta, operación privada usada por `CA-005-05` o parámetros de seguridad. Registra cualquier vacío antes de implementar; no uses como decisión la propuesta histórica de llaves asimétricas de `docs/10-backlog/prompt-llaves-asimetricas.md`, descartada por `DEC-050`.
6. Localiza un issue abierto que cubra exactamente esta HU. Si no existe, créalo o solicita autorización para crearlo con el título sugerido, los siete criterios, pruebas y exclusiones de este archivo.
7. Sustituye `issue: pending` por el issue real, actualiza este archivo y el índice del catálogo y cambia a `ready` solo cuando el issue exista (`CT-003` ya está resuelta por `DEC-055`).
8. Crea `feat/<issue>-hu005-inicio-sesion` desde `main` actualizada y registra rama y estado `in_progress` antes de modificar código.

## Alcance incluido

- Operación de inicio de sesión contract-first en `/api/v1/public/auth/login`, sin middleware de autenticación (`DEC-055`).
- Migración Atlas de la parte de autenticación estrictamente necesaria: resolución previa del tenant, `staff_credential`, `staff_session`, RLS, privilegios, funciones estrechas e índices de las secciones A.0–A.2 del modelo físico de referencia, revisadas contra el estado aplicado real.
- Normalización del correo y verificación de contraseña con una función de derivación resistente a fuerza bruta, sal única y parámetros versionados/documentados.
- Token de sesión opaco generado con `crypto/rand`, persistido solo como hash y enviado una sola vez mediante cookie con los atributos exactos aprobados a partir de `DEC-050`.
- Caso de uso, puertos, repositorio PostgreSQL y adaptador HTTP dentro del módulo `auth`, conservando dominio y aplicación independientes de Chi y PostgreSQL.
- Validación mínima de sesión y operación privada real necesarias para probar `CA-005-01` y `CA-005-05`, únicamente si el issue define su forma exacta. No inventes un endpoint de demostración.
- OpenAPI 3.1.2, problemas RFC 9457, esquema de seguridad por cookie, ejemplos ficticios y pruebas de contrato.

## Fuera de alcance

- Cierre de sesión, cobertura estructural de todas las rutas privadas y renovación deslizante completa de `HU-006`.
- Límite por IP o verificación telefónica de `HU-007`.
- Recuperación, códigos y teléfono verificado de `HU-008`.
- Pantallas Vue de `HU-010` a `HU-012`.
- Tokens firmados, JWT, PASETO, llaves asimétricas o refresh tokens: `DEC-050` eligió un único token opaco revocable.
- Agregar las secciones A.3 o A.4 del modelo de referencia, editar migraciones aplicadas o copiar el modelo completo como una sola migración.
- Crear usuarios reales, un flujo de alta o un endpoint privado ficticio solo para satisfacer una prueba.

## Estado existente que debes conservar

- El router, los errores uniformes, la correlación y el logging seguro de `HU-003` ya existen y se reutilizan.
- `database.DB.InTenantTx` de `HU-002` es el único camino de consultas tenant-aware; el pool no se expone a los módulos.
- Solo cinco migraciones están aplicadas. `database/modelo-fisico-referencia.sql` es diseño revisable, no una migración; copia únicamente A.0–A.2 y vuelve a auditarlo.
- `barberia_owner`, `barberia_migrator`, `barberia_app` y `barberia_worker` siguen `DEC-040`; una función `SECURITY DEFINER` usa `search_path = ''`, nombres calificados, `REVOKE ... FROM PUBLIC` y un `GRANT` mínimo.
- `staff_credential` no concede `SELECT` amplio al rol del API; la lectura ocurre mediante una función acotada a un usuario y al tenant vigente.
- Las migraciones aplicadas son inmutables y `atlas.sum` debe acompañar cualquier archivo nuevo.

## Trabajo requerido

1. Audita la migración vigente de `staff_user`, los roles y las secciones A.0–A.2. Genera una migración con Atlas y adapta el SQL al ownership y privilegios reales; no pegues el modelo sin revisar bloqueos, objetos existentes y recuperación.
2. Implementa la resolución previa del tenant sin exponer si el correo existe. Después abre la transacción tenant-aware y obtiene una sola credencial mediante la función estrecha; no agregues una consulta global de credenciales.
3. Define un puerto de hash/verificación de contraseña en el núcleo de `auth`. Usa el algoritmo ya permitido por el diseño de referencia o registra y justifica cualquier cambio, incluidos parámetros, licencia y dependencia. El valor codificado debe incluir sal y parámetros; nunca cifrado reversible ni hash rápido simple.
4. Para correo inexistente, ejecuta una verificación contra un hash ficticio válido con los mismos parámetros. Devuelve exactamente el mismo status, problem type, code, detail y headers que para contraseña incorrecta; no uses `sleep` para maquillar diferencias. Documenta una estrategia de prueba estable para el tiempo perceptible en vez de una aserción frágil de nanosegundos.
5. Genera el token opaco con entropía criptográfica suficiente y codificación apta para cookie. Calcula su hash antes de persistirlo; el token en claro no entra al logger, a errores, fixtures, base de datos ni respuesta JSON.
6. Persiste la sesión asociada a `staff_user_id` y `barbershop_id`, con emisión y vencimiento de 30 días según `DEC-050`. Fija la cookie `HttpOnly`, `Secure`, `SameSite` y los demás atributos exactos que la resolución normativa o el issue hayan aprobado; si alguno sigue sin definir, registra el bloqueo antes de elegirlo.
7. Impide el acceso de usuarios inactivos o eliminados y asegura que una sesión suya no habilite solicitudes privadas. Una desactivación realizada directamente en el escenario de integración debe producir rechazo sin revelar el motivo.
8. Declara primero OpenAPI: request cerrado, contraseña `writeOnly`, respuesta sin secretos, `Set-Cookie` descrita sin valor funcional, seguridad explícita, problemas `400`/`401`/`422`/`429` solo cuando correspondan al comportamiento real, `500`, `X-Request-Id`, `x-business-rules` y `x-decisions`. Después implementa handler y contrato.
9. Usa una operación privada real aprobada para demostrar que la sesión de A fija el contexto de A y no alcanza B. Si ninguna fuente define esa operación, registra la duda y detente; no publiques una ruta de prueba.
10. Actualiza composición en `cmd/api` sin introducir estado global, lectura de entorno fuera de `config`, SQL en handlers ni importaciones de Chi/PostgreSQL desde el núcleo.

## Pruebas y evidencia

- Unitarias del servicio: normalización de correo, contraseña correcta/incorrecta, correo inexistente con hash ficticio, usuario inactivo, error de persistencia, cancelación de contexto y generador criptográfico fallido.
- Dos usuarios con la misma contraseña producen valores codificados distintos y ambos verifican correctamente.
- HTTP con `httptest`: éxito, JSON inválido, campo desconocido, faltante, credenciales incorrectas, correo inexistente, usuario inactivo, error interno seguro y atributos de `Set-Cookie` sin token en el body.
- La respuesta de correo inexistente y contraseña incorrecta es idéntica en cuerpo, código y headers funcionales; la evidencia de tiempo muestra el mismo trabajo criptográfico y una distribución sin diferencia perceptible, sin depender de `sleep`.
- PostgreSQL real 14 o la versión fijada, con dos tenants y `barberia_app`: migración desde vacío y desde versión anterior con datos; RLS/privilegios; relación cruzada rechazada; `SELECT` amplio de credenciales denegado; funciones `SECURITY DEFINER` acotadas.
- La cookie de A autentica la operación privada aprobada dentro de A y nunca permite leer/modificar B; token desconocido, alterado o perteneciente a usuario inactivo se rechaza de forma uniforme.
- Captura completa de logs del recorrido inspeccionada por correo, contraseña, cookie, token y hash: ocurrencias permitidas, cero.
- Contrato: lint, bundle, ejemplos válidos y handler verificado contra método, path, seguridad, status, headers, media types y schemas.

Entrega una tabla `Criterio | Estado | Prueba o evidencia` para `CA-005-01` a `CA-005-07`. No declares cumplido `CA-005-05` con un handler solo de pruebas ni `CA-005-02` solo porque el texto coincide.

## Documentación y trazabilidad

- Actualiza `apps/api/README.md` con el flujo de autenticación, configuración no secreta y forma de ejecutar pruebas; nunca documentes valores reales.
- Actualiza OpenAPI, `CHANGELOG`, matriz, historial, diagrama/diccionario y documentación de base de datos solo cuando el cambio los afecte realmente.
- Registra algoritmo/parámetros, cookie y operación privada en la fuente aprobada correspondiente si las fuentes actuales no los fijan.
- Actualiza metadatos e índice del catálogo con issue, rama, PR y estado reales. Un cambio material después de iniciar crea una nueva versión del prompt.

## Verificación final

Ejecuta los comandos equivalentes definidos por el repositorio para:

```text
atlas migrate hash
atlas migrate validate --env local
atlas migrate apply --env local --dry-run
aplicación en base vacía y actualización desde la versión anterior con datos
pruebas SQL/RLS con dos tenants y el rol real
gofmt, go vet, go build y go test del módulo
go test -race ./...
openapi:check-config
openapi:lint
openapi:bundle
pruebas HTTP y de contrato
git diff --check
graphify update .
```

No reduzcas umbrales, no ignores pruebas inestables y no uses SQLite o mocks para afirmar RLS, privilegios o aislamiento.

## Git y PR

- Commits y título: `feat(auth): implementa inicio de sesión seguro` o equivalente Conventional Commits.
- Abre un PR borrador contra `main`; usa `Closes #<issue>` solo cuando los siete criterios estén probados.
- Incluye contrato, migración/recuperación, parámetros de hash, seguridad de cookie, evidencia de no enumeración, aislamiento, logs y tabla de criterios.
- No hagas push directo, force push, merge manual ni reescribas `main`.
