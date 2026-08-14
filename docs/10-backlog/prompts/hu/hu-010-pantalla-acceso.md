---
prompt_id: "PROMPT-HU-010-v1"
version: "1.1"
kind: "hu"
status: "executed"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-010"
related_hu:
  - "HU-005"
  - "HU-007"
  - "HU-009"
  - "HU-011"
  - "HU-012"
issue: 46
issue_url: "https://github.com/bcaceres19/barberia/issues/46"
suggested_issue_title: "feat(web): implementar HU-010 pantalla de acceso"
branch: "feat/46-hu010-pantalla-acceso"
pr: 52
pr_url: "https://github.com/bcaceres19/barberia/pull/52"
depends_on:
  - "HU-005 integrada"
  - "HU-009 integrada"
rules:
  - "RN-DAT-02"
  - "Criterios no funcionales de UX y accesibilidad"
decisions:
  - "DEC-026"
  - "DEC-033"
  - "DEC-035"
  - "DEC-037"
  - "DEC-038"
  - "DEC-039"
  - "DEC-050"
  - "DEC-056"
acceptance_criteria:
  - "CA-010-01"
  - "CA-010-02"
  - "CA-010-03"
  - "CA-010-04"
  - "CA-010-05"
  - "CA-010-06"
  - "CA-010-07"
  - "CA-010-08"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/04-arquitectura/frontend.md"
  - "docs/06-api/estandar-openapi.md"
  - "apps/web/package.json"
  - "apps/web/src/app/router/index.ts"
  - "apps/web/src/modules/auth"
  - "apps/web/src/shared/api"
  - "apps/web/src/shared/ui"
  - "api/openapi/openapi.yaml"
created_at: "2026-08-13"
updated_at: "2026-08-14"
supersedes: "docs/10-backlog/prompts-implementacion.md#prompt-de-hu-010--pantalla-de-acceso"
superseded_by: null
---

# Implementar HU-010: pantalla de acceso

## Instrucción para Claude o Codex

Implementa únicamente `HU-010` en Vue contra el contrato real de `HU-005`. Compón la pantalla con el sistema visual de `HU-009`, cliente tipado derivado del bundle OpenAPI y pruebas de componente/E2E. No construyas el panel de `HU-012`, la recuperación de `HU-011` ni el backend de rate limiting de `HU-007`.

## Objetivo

Entregar una pantalla de acceso móvil primero, clara y accesible: correo y contraseña, envío único, mensajes no enumerables, recuperación visible y manejo seguro de red lenta. Un acceso correcto navega a `/panel`, ruta privada real y protegida por un guard mínimo propio de esta historia (`DEC-056`); no construyas cabecera, navegación ni el cascarón completo, que pertenecen a `HU-012`.

## Preflight obligatorio

1. Comprueba árbol limpio, `main` actualizada por fast-forward y ausencia de cambios ajenos que puedas sobrescribir.
2. Consulta el Graphify de raíz y el de `apps/web` por `HU-010`, router, módulo `auth`, `shared/api`, componentes base y pruebas si están disponibles.
3. Lee completamente cada `source_docs`. Confirma que `HU-005` y `HU-009` están integradas y que sus contratos/pruebas están en verde.
4. `CT-004` está resuelta (`DEC-056`): `CA-010-01` se divide entre esta historia (navega a `/panel` con guard mínimo propio) y `HU-012` (cascarón completo, verificación end-to-end). No adelantes cabecera, navegación ni ningún contenido del cascarón de `HU-012`: el guard de `/panel` es deliberadamente mínimo para que `HU-012` lo reutilice sin duplicarlo.
5. Confirma el destino navegable del enlace de recuperación. `CA-010-08` exige un enlace visible, pero el flujo completo pertenece a `HU-011`; no publiques una ruta rota ni implementes la recuperación dentro de esta HU. Si las fuentes no definen una transición incremental válida, registra la duda antes de codificar.
6. Localiza un issue que cubra exactamente `HU-010`; si no existe, créalo o solicita autorización con el título sugerido, ocho criterios, evidencia responsive y exclusiones.
7. Registra issue/URL, cambia a `ready` solo con las dependencias cumplidas (`CT-004` ya está resuelta por `DEC-056`), crea `feat/<issue>-hu010-pantalla-acceso` desde `main` y actualiza el estado a `in_progress`.

## Alcance incluido

- Ruta de acceso cargada de forma diferida dentro de `apps/web/src/modules/auth` y registrada desde `app/router`.
- Pantalla conforme a la plantilla visual P0: marca discreta, título, formulario estrecho, una acción principal y recuperación visible.
- Campos de correo y contraseña con etiquetas, ayuda/errores asociados, atributos de autocompletado apropiados y validación de forma sin revelar existencia de cuentas.
- Cliente HTTP tipado generado o derivado del bundle OpenAPI de `HU-005`; no DTO manual paralelo ni `fetch` desde componentes.
- Estados inicial, enviando, credenciales inválidas, error de red recuperable y respuesta `429` documentada. La integración E2E del escalamiento se completa en `HU-007`; esta HU solo puede probar el estado de presentación contra el contrato existente.
- Prevención de doble envío, conservación segura de datos durante errores recuperables y navegación posterior a `/panel` (`DEC-056`).
- Ruta privada `/panel` y su guard mínimo: sin sesión válida redirige al acceso; con sesión válida muestra un marcador de posición autenticado, sin cabecera ni navegación general — ese trabajo pertenece a `HU-012`, que reutiliza este guard sin crear uno paralelo.
- Pruebas de componente, axe-core, E2E del acceso real y evidencia a 320, 360, 768 y 1280 px.

## Fuera de alcance

- Implementar o modificar credenciales, cookies, sesiones o handlers de `HU-005`/`HU-006`, salvo corregir una incompatibilidad de contrato mediante su propio issue.
- Lógica de rate limit o verificación telefónica de `HU-007`.
- Flujo de solicitud, código, reenvío y cambio de contraseña de `HU-011`.
- Cascarón, navegación privada, restauración de sesión y logout visual de `HU-012`.
- Crear un destino privado provisional, una ruta de recuperación rota o una pantalla “próximamente” para aparentar cumplimiento.
- Nuevos colores, medidas, tipografías, variantes visuales o dependencia de componentes. Si falta una primitiva compartida material, registra un fix separado de `HU-009`.
- Pinia, estado global, almacenamiento de contraseña/token o una biblioteca de formularios sin necesidad demostrada.

## Estado existente que debes conservar

- `HU-009` ya entrega `BaseButton`, `BaseInput`, `BaseAlert`, `BaseBadge`, `BaseDialog`, tokens y pruebas. Reutiliza su API pública; no copies estilos ni componentes dentro de `auth`.
- La dependencia es `app → modules → shared`. `auth` puede importar `shared`, pero `shared` no importa `auth` y un módulo no consume internos de otro.
- TypeScript permanece estricto, sin `any`; estados asíncronos se representan con una unión discriminada.
- Todo acceso HTTP pasa por `shared/api`; los tipos nacen del bundle y el código generado no se edita.
- La cookie es `HttpOnly`: JavaScript no necesita ni puede leer el token. La UI reacciona al resultado HTTP, no administra material de sesión.
- `HU-007` ocurre después en el orden B0. Una prueba de componente puede representar un `429` ya documentado, pero no afirma que el escalamiento backend exista.

## Trabajo requerido

1. Audita router, `modules/auth`, `shared/api` y componentes reales. Elimina o adapta solo placeholders dentro de esta HU; no reorganices módulos no relacionados.
2. Establece una ruta diferida y una página coordinadora pequeña. La página posee estado/navegación; el formulario emite una intención tipada y no conoce detalles RFC 9457 ni llama `fetch`.
3. Si aún no existe generación/derivación del cliente OpenAPI, elige la herramienta mínima conforme a `DEC-037`: documenta necesidad, alternativas, licencia, mantenimiento, superficie transitiva y comando reproducible. Genera desde el bundle, fija versiones/lockfile y nunca edites el resultado.
4. Implementa correo y contraseña con `BaseInput`; usa `type="email"`, `autocomplete="username"` y `autocomplete="current-password"` cuando corresponda. No pongas credenciales en URL, query, historial, console, analytics, mensajes de error ni almacenamiento persistente.
5. Valida forma y campos obligatorios en cliente para ayudar, sin sustituir al backend. Credencial incorrecta y cuenta inexistente muestran el mismo mensaje accionable y conservan el correo sin afirmar si existe.
6. Durante envío, conserva el ancho/jerarquía del botón, cambia el texto a una acción explícita, deshabilita nuevos envíos y garantiza una sola llamada ante doble click/toque/Enter. No agregues `Idempotency-Key` a login salvo que el contrato lo exija.
7. Ante error de red, conserva correo y el estado estrictamente necesario en memoria, muestra `Reintentar` sin recargar y evita persistir la contraseña. Define de forma explícita si la contraseña se conserva solo en memoria conforme a las fuentes; registra una duda si los criterios y la política de seguridad no permiten una interpretación única.
8. Mapea problemas por `status`, `type` o `code`, nunca por `detail`. Un error inesperado muestra texto seguro y `requestId` cuando exista. Un `429` documentado explica el bloqueo en lenguaje simple y respeta `Retry-After` si está disponible, sin fingir que `HU-007` ya está implementada.
9. Tras éxito, navega exactamente a `/panel` (`DEC-056`). Implementa solo el guard mínimo de esa ruta única (sin sesión válida redirige al acceso); no construyas componentes del panel, cabecera, navegación ni un guard global — `HU-012` reutiliza este guard, no crea uno paralelo.
10. Muestra un enlace semántico y operable a recuperación usando la transición incremental aprobada. No uses un botón disfrazado de enlace, `href="#"` ni un destino inexistente.
11. Compón solo con tokens y componentes aprobados. Verifica 44 × 44 px, foco visible, orden de tabulación, zoom 200 %, una mano a 320/360 px, movimiento reducido y ausencia de desplazamiento horizontal.

## Pruebas y evidencia

- Componente: render inicial, etiquetas/ayuda, validación, envío válido, credenciales inválidas, error de red/reintento, `429`, error inesperado, botón cargando y doble toque/Enter con una sola solicitud.
- Accesibilidad: consultas por rol/etiqueta, errores asociados, foco llevado al resumen cuando haya varios, orden de teclado, enlace de recuperación y `vitest-axe` sin violaciones aplicables.
- Seguridad: la contraseña no aparece en URL, historial, storage, console, errores serializados ni capturas; el cliente usa cookies sin leer el token.
- Cliente tipado: request/response/problem derivados del bundle y prueba de mapeo por código estable; búsqueda que demuestre ausencia de DTO duplicado o `fetch` directo en el módulo.
- E2E contra API real local: acceso correcto, credenciales inválidas indistinguibles y fallo recuperable. El destino de éxito es la navegación real a `/panel` y su guard mínimo (`DEC-056`); la verificación del cascarón completo queda para `HU-012`.
- El estado `429` tiene prueba de componente contra el contrato; el E2E de umbral/escalamiento queda explícitamente asociado a `HU-007`, sin declararlo ejecutado ahora.
- Evidencia visual de normal, foco, carga, error y estado bloqueado a 320, 360, 768 y 1280 px; zoom 200 %, teclado y sin scroll horizontal.
- Build confirma carga diferida y que el flujo de acceso no descarga módulos privados futuros.

Entrega una tabla `Criterio | Estado | Prueba o evidencia` para `CA-010-01` a `CA-010-08`. `CA-010-01` solo puede quedar cumplido con la navegación real a `/panel` y su guard mínimo (`DEC-056`); no lo declares cumplido con la evidencia end-to-end completa del panel, que corresponde a `CA-012-01`/`CA-012-02` de `HU-012`. `CA-010-08` exige un enlace navegable, no texto decorativo.

## Documentación y trazabilidad

- Documenta la API pública del módulo `auth`, comando de generación del cliente y estados de error solo en el lugar que corresponda; no dupliques OpenAPI.
- Actualiza matriz, historial, router/README y evidencia visual realmente afectada.
- Si aparece una brecha material en un componente de `HU-009`, crea un issue/prompt separado en vez de ampliar silenciosamente esta HU.
- Actualiza este prompt y el índice con issue, rama, PR y estado reales. Un cambio material después de iniciar crea una nueva versión.

## Verificación final

Ejecuta los scripts reales de `apps/web/package.json` y del contrato:

```text
pnpm install --frozen-lockfile (en el workspace correspondiente)
pnpm run openapi:check-config
pnpm run openapi:lint
pnpm run openapi:bundle
generación/verificación reproducible del cliente tipado
pnpm run format
pnpm run lint
pnpm run typecheck
pnpm run test:unit
pnpm run build
pnpm run test:e2e para el recorrido afectado
git diff --check
graphify update .
```

No actualices snapshots sin revisar el comportamiento, no ocultes pruebas inestables y no declares evidencia de `HU-007`, `HU-011` o `HU-012` que todavía no exista.

## Git y PR

- Commits y título: `feat(web): implementa pantalla de acceso` o equivalente Conventional Commits.
- Abre un PR borrador contra `main`; usa `Closes #<issue>` solo si los ocho criterios están probados conforme a las resoluciones vigentes.
- Incluye contrato/cliente, pruebas, evidencia responsive y accesible, navegación de éxito, seguridad de credenciales, bundle y tabla de criterios.
- No hagas push directo, force push, merge manual ni reescribas `main`.
