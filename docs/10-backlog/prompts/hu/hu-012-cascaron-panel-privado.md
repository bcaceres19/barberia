---
prompt_id: "PROMPT-HU-012-v1"
version: "1.2"
kind: "hu"
status: "executed"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-012"
related_hu:
  - "HU-006"
  - "HU-009"
  - "HU-010"
  - "HU-020"
issue: 56
issue_url: "https://github.com/bcaceres19/barberia/issues/56"
suggested_issue_title: "feat(web): implementar HU-012 cascarón del panel privado"
branch: "feat/56-hu012-cascaron-panel"
pr: 59
pr_url: "https://github.com/bcaceres19/barberia/pull/59"
depends_on:
  - "HU-006 integrada"
  - "HU-009 integrada"
  - "HU-010 integrada"
  - "DP-SEG-09 resuelta (DEC-060) y propagada a las fuentes"
rules:
  - "RN-TEN-01"
  - "RN-DAT-02"
  - "Criterios no funcionales de UX y accesibilidad"
decisions:
  - "DEC-024"
  - "DEC-026"
  - "DEC-033"
  - "DEC-035"
  - "DEC-037"
  - "DEC-038"
  - "DEC-039"
  - "DEC-050"
  - "DEC-056"
  - "DEC-057"
  - "DEC-060"
acceptance_criteria:
  - "CA-012-01"
  - "CA-012-02"
  - "CA-012-03"
  - "CA-012-04"
  - "CA-012-05"
  - "CA-012-06"
  - "CA-012-07"
  - "CA-012-08"
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
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/04-arquitectura/frontend.md"
  - "docs/04-arquitectura/backend-go.md"
  - "docs/06-api/estandar-openapi.md"
  - "apps/web/README.md"
  - "apps/web/src/app/router/index.ts"
  - "apps/web/src/modules/auth"
  - "apps/web/src/shared/api"
  - "apps/web/src/shared/ui"
  - "apps/api/README.md"
  - "apps/api/internal/modules/auth"
  - "api/openapi/openapi.yaml"
  - "api/openapi/paths/public-auth.yaml"
  - "api/openapi/paths/private-auth.yaml"
created_at: "2026-08-14"
updated_at: "2026-08-17"
supersedes: "docs/10-backlog/prompts-implementacion.md#prompt-de-hu-012--cascarón-del-panel-privado"
superseded_by: null
---

# Implementar HU-012: cascarón del panel privado

## Instrucción para Claude o Codex

Implementa únicamente `HU-012` como estructura reutilizable del área privada. Reutiliza y generaliza la ruta `/panel` y el guard mínimo entregados por `HU-010`; integra la sesión real de `HU-006`, el sistema visual de `HU-009` y el cliente derivado de OpenAPI. No construyas la agenda ni capacidades de B1.

`DP-SEG-09` quedó resuelta el 2026-08-17 como `DEC-060` (ver sección siguiente): este prompt ya no está bloqueado y su `status` es `ready`. Antes de codificar, relee `DEC-060` completa en `registro-decisiones.md` — fija la forma exacta de la operación autorizada; no la reinterpretes ni la amplíes.

## Objetivo

Permitir que un barbero con una cookie válida abra o recargue la aplicación y entre directamente al panel, vea siempre el nombre de su barbería activa, navegue entre rutas privadas sin recargar, recupere el destino pretendido después de autenticarse y cierre la sesión en el servidor desde la cabecera. Una expiración, revocación o pérdida de conexión debe producir un estado único, comprensible y recuperable, nunca una pantalla en blanco ni un bucle de redirección.

## Bloqueo resuelto: `DP-SEG-09` → `DEC-060`

Hasta el 2026-08-17 no existía una lectura privada no destructiva para rehidratar la cookie `HttpOnly` y obtener el contexto de barbería:

- `POST /api/v1/public/auth/login` devuelve solo `expiresAt`; el token vive únicamente en la cookie.
- `POST /api/v1/private/auth/logout` valida y destruye la sesión actual; no sirve como consulta.
- el guard de `HU-010` usa `expiresAt` en `sessionStorage`, marcador no autoritativo que no sobrevive al cierre de la pestaña o navegador;
- ningún contrato vigente entrega al frontend el nombre de la barbería activa antes de `HU-020`.

`DEC-060` fija la resolución exacta que debes implementar, sin reinterpretarla ni ampliarla:

- Nueva operación `GET /api/v1/private/auth/session`, seguridad `SessionCookie`, montada sobre el `SessionMiddleware` existente (misma validación y renovación deslizante que ya usa `logout`; no se crea una segunda ruta de validación de sesión).
- `200`: payload mínimo `{ barbershop: { id, name }, expiresAt }`. `barbershop.name` sale de la columna real `barbershop.name` (ya existe desde `20260807170000_create_tenant_foundation.sql`), aislada por tenant vía el RLS ya vigente (`DEC-024`). No agregues `staffUserID`, correo ni nombre del barbero al payload: no lo exige ningún `CA-012-*` y rompería el patrón de `auth.Principal` (identificadores opacos, sin datos personales).
- `401`: mismo `UnauthorizedProblem` uniforme que `logout` ante cookie ausente, inválida, vencida o revocada — ninguna rama de este endpoint puede distinguir esos casos en la respuesta.
- Este endpoint reemplaza `sessionMarker.ts`/`hasRememberedSession` (HU-010, `DP-SEG-08`) como fuente de bootstrap; el guard generalizado de esta historia lo llama al abrir/recargar la aplicación en vez de leer `sessionStorage`.

Relee `DEC-060` completa en `docs/00-control/registro-decisiones.md` antes de codificar; si detectas una ambigüedad que esa decisión no cubre, detente y regístrala como una duda nueva en vez de resolverla por tu cuenta.

## Preflight obligatorio

1. Comprueba árbol limpio, `main` actualizada por fast-forward y ausencia de cambios ajenos que puedas sobrescribir.
2. Consulta Graphify de raíz y `apps/web/graphify-out/graph.json` por `HU-012`, `requireSession`, `sessionMarker`, `PanelPlaceholderPage`, router, cliente HTTP, middleware de sesión, logout y respuestas `401`.
3. Lee completamente cada `source_docs`; los directorios listados significan revisar su API pública, archivos relevantes y pruebas, no cargar archivos generados o cachés sin necesidad.
4. Confirma en Git que `HU-006`, `HU-009` y `HU-010` están integradas y que sus suites pasan.
5. Verifica que `DP-SEG-09` esté resuelta y propagada. Si sigue abierta o la decisión no cubre tanto restauración de sesión como barbería activa, detente y deja `status: blocked`.
6. Confirma que el issue `#56` conserva exactamente `HU-012`. Actualiza este archivo a `ready` solo cuando todas las dependencias estén satisfechas.
7. Crea `feat/56-hu012-cascaron-panel` desde `main` actualizada y cambia este prompt a `in_progress`; no reutilices la rama documental del prompt.

## Alcance incluido

- Cascarón autenticado sobre `/panel`: cabecera con barbería activa, navegación principal y área de contenido.
- Generalización del guard de `HU-010` para todas las rutas privadas, con restauración de sesión autoritativa y destino pretendido seguro.
- Manejo central y único de `401`: limpiar estado local no sensible una sola vez, informar el motivo y redirigir sin bucles.
- Cierre de sesión desde la cabecera contra `POST /api/v1/private/auth/logout` y limpieza local posterior al resultado seguro.
- Estados globales aplicables: bootstrap/carga, conexión perdida, error recuperable con “Reintentar” y contenido previo conservado.
- Pantalla inicial provisional que muestra el estado de sesión y contexto; no representa una agenda inexistente.
- Rutas diferidas, navegación SPA, componentes y evidencia responsive/accesible en los cuatro anchos.
- Cambio HTTP mínimo que autorice `DP-SEG-09`, si esa es la resolución, actualizado contract-first junto con handler, cliente y pruebas.

## Fuera de alcance

- Agenda diaria, citas, servicios, horarios, bloqueos, configuración editable o cualquier entidad de B1–B3.
- Crear otro guard, ruta de panel o almacén de sesión paralelo a lo entregado por `HU-010`.
- Guardar token, cookie, contraseña o respuesta sensible en `localStorage`, `sessionStorage`, Pinia, logs o consola.
- Tratar un marcador del navegador como prueba de autorización; el backend conserva la autoridad en cada solicitud.
- Inventar el nombre de la barbería, persistirlo indefinidamente como autoridad o adelantar la edición de `HU-020`.
- Introducir Pinia, una biblioteca visual, iconos o dependencia nueva sin necesidad y justificación conforme a los estándares.
- Modo offline, sincronización en segundo plano, agenda simulada, múltiples barberías simultáneas o selector de tenant.
- Estilos visuales, colores, tamaños, tipografías o variantes fuera de tokens y componentes aprobados.

## Estado existente que debes conservar

- `HU-006` entrega cookie `barberia_session`, middleware privado, renovación deslizante y logout real con aislamiento de dos tenants.
- `HU-010` entrega `/acceso`, `/panel`, `requireSession`, `sessionMarker` y `PanelPlaceholderPage`. El marcador local está documentado como provisional; evolúcionalo, no dupliques su responsabilidad.
- `HU-009` entrega `BaseButton`, `BaseInput`, `BaseAlert`, `BaseBadge`, `BaseDialog`, tokens y pruebas accesibles. Reutiliza su API pública.
- `shared/api/httpClient.ts` y los tipos de `shared/api/generated` derivan del bundle OpenAPI; los componentes no llaman `fetch`.
- Toda ruta de `/api/v1/private` ya pasa por `SessionMiddleware`; conserva la prueba estructural de `CA-006-04`.
- `DEC-056` fija `/panel` y exige reutilizar/generalizar el guard existente, sin ruta o guard paralelo.

## Trabajo requerido después de desbloquear

1. Implementa exactamente la fuente autoritativa aprobada por `DP-SEG-09`. Si añade o cambia HTTP, actualiza primero OpenAPI 3.1.2 con seguridad `SessionCookie`, respuestas RFC 9457, ejemplos ficticios, `RN-*`/`DEC-*`; luego handler/servicio/repositorio, pruebas y cliente generado en la misma entrega.
2. Modela el bootstrap de la aplicación como estados discriminados y explícitos: comprobando, autenticado con contexto mínimo, no autenticado y fallo recuperable. Ningún estado local concede permisos por sí mismo.
3. Generaliza `requireSession` o su API pública en un solo camino para todas las rutas privadas. Conserva el destino pretendido como ruta interna validada; rechaza destinos externos o ciclos hacia acceso/logout.
4. Compón un layout privado desde `app` usando APIs públicas de módulos. Mantén `app → modules → shared`; no importes internos de otro módulo ni conviertas `shared` en dueño del dominio de autenticación.
5. Implementa una coordinación única para respuestas `401`. Múltiples solicitudes concurrentes no deben limpiar/redirigir repetidas veces ni producir navegación cíclica. Los errores inesperados conservan `requestId` cuando exista y nunca muestran detalles internos.
6. Integra el logout tipado. Deshabilita el control durante la solicitud, evita doble envío, y limpia el estado local incluso cuando el servidor confirme que la sesión ya no es válida, sin fingir éxito ante fallos de red ambiguos.
7. Construye cabecera, navegación y pantalla provisional con los tokens/patrones existentes. El nombre de la barbería viene de la fuente aprobada; carga, conexión perdida y reintento conservan el contenido no sensible.
8. Mantén todas las rutas diferidas y confirma que el flujo público no descarga el árbol privado antes de necesitarlo.
9. Elimina el comportamiento provisional que quede sustituido, sin borrar cobertura de `HU-010`; actualiza sus pruebas para demostrar la transición deliberada.

## Pruebas y evidencia

- Unitarias/componente: estados de bootstrap, sesión válida/ausente/vencida, destino pretendido válido e inválido, una sola limpieza ante varios `401`, fallo de red y reintento sin perder estado.
- Componente del cascarón: barbería activa siempre visible, navegación semántica, logout deshabilitado durante envío, foco, teclado, aviso de conexión y ausencia de violaciones automatizadas con axe-core.
- HTTP/Go si `DP-SEG-09` aprueba una operación: forma exacta, cookie ausente/inválida/válida, sesión revocada/vencida, dos tenants, no filtración de token y contrato contra el bundle.
- E2E contra API y PostgreSQL reales: acceso → `/panel` completo; cerrar y reabrir navegador con sesión vigente; recarga directa; ruta privada sin sesión → acceso → destino original; `401` sin bucle; conexión interrumpida con reintento; logout invalida servidor y vuelve a acceso.
- Evidencia en 320, 360, 768 y 1280 px, zoom 200 %, teclado completo, foco no oculto y navegación móvil/escritorio según el estándar.
- Repite los E2E afectados de `HU-010`; no actualices snapshots ni evidencia sin inspeccionar el resultado.

## Criterios de aceptación y cierre

Entrega una tabla `Criterio | Estado | Prueba o evidencia` para `CA-012-01` a `CA-012-08`.

- `CA-012-01` exige restauración real desde la cookie después de cerrar/reabrir, no solo continuidad en la misma pestaña.
- `CA-012-02` exige cualquier ruta privada y retorno al destino pretendido, no una excepción codificada solo para `/panel`.
- `CA-012-03` requiere una prueba de concurrencia o repetición que demuestre una única limpieza/redirección.
- `CA-012-04` no se cumple con texto fijo, fixture productivo ni dato local autoritativo.
- `CA-012-07` debe demostrar revocación en servidor reutilizando el material anterior.
- `CA-012-08` incluye comportamiento, reflow, teclado y evidencia; una captura aislada no basta.

### Evidencia real (ejecución del 2026-08-17)

| Criterio | Estado | Prueba o evidencia |
| --- | --- | --- |
| `CA-012-01` | Cumplido | `TestSessionContext_HTTP_ValidSession_ReturnsRealBarbershopName` (Go); `una sesión real sobrevive a cerrar y reabrir el navegador` y `una recarga directa de /panel conserva la sesión` (`e2e/panel.spec.ts`), contra cookie real, sin marcador local. |
| `CA-012-02` | Cumplido | `requireSession.ts` genérico sobre la ruta padre `/panel`; `routes.test.ts` (`preserves the intended destination...`); `e2e/panel.spec.ts` (`sin sesión, cualquier ruta privada redirige...`) y `LoginPage.test.ts` (`returns to the intended destination...`). |
| `CA-012-03` | Cumplido | `sessionStore.test.ts` (`reportUnauthorized notifies listeners exactly once for several near-simultaneous 401s`), `installSessionHandling.test.ts`. |
| `CA-012-04` | Cumplido | `AppHeader.test.ts`, `TestSessionContext_HTTP_ValidSession_ReturnsRealBarbershopName`/`TestSessionContext_HTTP_TwoTenants_NeverCrossesBarbershopNames` (Go, dos tenants reales), `e2e/panel.spec.ts` (`la cabecera muestra siempre la barbería activa real`). |
| `CA-012-05` | Cumplido | `PrivateShell.test.ts`, `e2e/panel.spec.ts` (`una pérdida de conexión ofrece Reintentar...`). |
| `CA-012-06` | Cumplido | Chunks separados por ruta en `pnpm run build`; `e2e/panel.spec.ts` (`la navegación entre acceso y panel no recarga la aplicación completa`). |
| `CA-012-07` | Cumplido | `AppHeader.test.ts` (logout real); `TestLogout_HTTP_*` (HU-006, reutilizado); `e2e/panel.spec.ts` (`cerrar sesión invalida el servidor...`, reinyecta la cookie capturada antes del logout en un contexto nuevo y confirma que el servidor la rechaza). |
| `CA-012-08` | Cumplido en Chromium/Firefox; **conocido, sin cubrir en WebKit** | `e2e/panel-evidencia-responsiva.spec.ts` en 320/360/768/1280 px + zoom 200 %; capturas en `apps/web/e2e/evidence/panel/`. Ver "Defecto conocido" abajo. |

Verificación final ejecutada completa: `openapi:check-config`/`lint`/`bundle` en verde; `gofmt -l .` vacío; `go vet ./...` y `go build ./...` en verde; `go test -race ./...` en verde (PostgreSQL 14 real, Docker, migraciones aplicadas); `generate:api`/`format`/`lint`/`typecheck`/`test:unit` (213 pruebas) en verde; `build` en verde; `test:e2e` (`playwright test`, las 4 plataformas): **164/180 en verde** — `chromium-desktop` 45/45, `chromium-mobile` 45/45, `firefox` 45/45, `webkit` 29/45.

### Defecto conocido, no introducido por esta HU: cookie `Secure` no persiste en WebKit sobre `http://localhost`

Los 16 fallos de `webkit` comparten una sola causa raíz: `GET /private/auth/session` responde `401` inmediatamente después de un `POST /public/auth/login` exitoso, porque WebKit no almacena una cookie `Secure=true` (`DEC-050`, no negociable) emitida sobre `http://localhost` sin TLS — a diferencia de Chromium/Firefox, que sí conceden esa excepción a `localhost` en desarrollo. `HU-010` nunca lo detectó porque su E2E no hacía ninguna llamada privada real tras el login (dependía del marcador `sessionStorage` que HU-012 reemplaza); `HU-012` es la primera en ejercitar `SessionCookie` de punta a punta en cada navegador, y por eso lo expone. No se corrige aquí: bajar `Secure` violaría `DEC-050`. Es una limitación exclusiva de pruebas locales por HTTP sin TLS; en despliegue real (HTTPS) no aplica. `playwright.config.ts` ya declara que WebKit/Firefox "se suman antes de una versión" (no bloquean cada PR; solo Chromium desktop/móvil sí). Registrado aquí para que el propietario decida si amerita una duda `DP-*` formal o un ajuste de entorno de pruebas (p. ej. HTTPS local) antes de exigir a WebKit en CI.

## Documentación y trazabilidad

- Actualiza OpenAPI/CHANGELOG y cliente generado solo si la resolución toca HTTP.
- Actualiza `apps/web/README.md` con bootstrap, rutas privadas, manejo `401` y comandos; `apps/api/README.md` si aparece una operación de sesión/contexto.
- Actualiza matriz e historial con pruebas reales y estado de `HU-012`; no declares cerrado B0.
- Actualiza este prompt y el catálogo con `status`, rama, PR y evidencia reales. Si aparece otro vacío normativo, regístralo antes de codificar.

## Verificación final

```text
# raíz
pnpm run openapi:check-config
pnpm run openapi:lint
pnpm run openapi:bundle

# apps/api, cuando exista cambio backend/HTTP
gofmt -l .
go vet ./...
go test -race ./...

# apps/web
pnpm run generate:api
pnpm run format
pnpm run lint
pnpm run typecheck
pnpm run test:unit
pnpm run build
pnpm run test:e2e

# raíz
git diff --check
```

Ejecuta cada grupo desde el directorio indicado. `gofmt -l .` debe producir salida vacía. Las pruebas que requieren API/PostgreSQL usan la preparación real documentada; no se sustituyen por mocks para declarar el E2E cumplido.

## Git y PR

- Rama: `feat/56-hu012-cascaron-panel`.
- Commit/PR: `feat(web): implementa HU-012 cascarón del panel privado`.
- Usa `Closes #56` solo cuando `DP-SEG-09` esté resuelta, `CA-012-01`–`CA-012-08` tengan evidencia y el issue esté cubierto completo; de lo contrario, `Refs #56` y conserva el prompt bloqueado o en progreso.
- Abre PR borrador, auto-revisa el diff y no hagas push directo, force push ni merge de `main`.

