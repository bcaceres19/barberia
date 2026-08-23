---
prompt_id: "PROMPT-HU-011-v1"
version: "1.0"
kind: "hu"
status: "executed"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-011"
related_hu:
  - "HU-008"
  - "HU-009"
  - "HU-010"
  - "HU-012"
issue: 64
issue_url: "https://github.com/bcaceres19/barberia/issues/64"
suggested_issue_title: "feat(web): implementar HU-011 pantalla de recuperación de acceso"
branch: "feat/64-hu011-pantalla-recuperacion"
pr: 66
pr_url: "https://github.com/bcaceres19/barberia/pull/66"
depends_on:
  - "HU-008 integrada"
  - "HU-009 integrada"
  - "HU-010 integrada"
rules:
  - "RN-DAT-01"
  - "RN-DAT-02"
  - "Criterios no funcionales de UX y accesibilidad"
decisions:
  - "DEC-026"
  - "DEC-039"
  - "DEC-063"
  - "DEC-064"
  - "DEC-065"
  - "DEC-066"
acceptance_criteria:
  - "CA-011-01"
  - "CA-011-02"
  - "CA-011-03"
  - "CA-011-04"
  - "CA-011-05"
  - "CA-011-06"
  - "CA-011-07"
  - "CA-011-08"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/dudas-pendientes.md"
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
  - "apps/api/README.md"
  - "apps/web/package.json"
  - "apps/web/src/app/router/index.ts"
  - "apps/web/src/modules/auth"
  - "apps/web/src/shared/api"
  - "apps/web/src/shared/ui"
  - "api/openapi/openapi.yaml"
  - "api/openapi/paths/public-auth.yaml"
created_at: "2026-08-23"
updated_at: "2026-08-23"
supersedes: "docs/10-backlog/prompts-implementacion.md#prompt-de-hu-011--pantalla-de-recuperacion-de-acceso"
superseded_by: null
---

# Implementar HU-011: pantalla de recuperación de acceso

## Instrucción para Claude o Codex

Implementa únicamente `HU-011` en Vue contra el contrato real y ya integrado de `HU-008` (`POST /api/v1/public/auth/recovery/request`, `.../verify` y `.../reset-password`, ver `apps/api/README.md` y `api/openapi/paths/public-auth.yaml`). Compón la pantalla con el sistema visual de `HU-009` y el cliente tipado derivado del bundle OpenAPI que ya usa `HU-010`. No modifiques el backend de `HU-008`, la pantalla de acceso de `HU-010` ni el cascarón de `HU-012`.

## Objetivo

Que un barbero que perdió su contraseña complete un flujo de tres pasos (solicitar, verificar código, establecer contraseña nueva), siempre sabiendo en qué paso está, a dónde llegó el código y qué hacer si no llega, sin llamar al propietario.

## Preflight obligatorio

1. Comprueba árbol limpio, `main` actualizada por fast-forward y ausencia de cambios ajenos que puedas sobrescribir.
2. Consulta Graphify (`graphify query`, `graphify explain`) por `HU-011`, el módulo `auth`, el enlace de recuperación que ya publica `HU-010` y el contrato de `HU-008`.
3. Lee completamente cada `source_docs`. Confirma en `apps/api/README.md` los contratos exactos de `recovery/request`, `recovery/verify` y `recovery/reset-password` (payloads, respuestas, códigos de error, RFC 9457, no enumeración) que entregó `HU-008` (PR #63, integrado en `main`).
4. Confirma que `HU-009` y `HU-010` siguen integradas y en verde; localiza en `HU-010` el enlace de recuperación existente (`CA-010-08`) y su destino declarado — esta historia implementa esa ruta, no crea una paralela.
5. El issue real es [#64](https://github.com/bcaceres19/barberia/issues/64). Si al ejecutar ya no está vigente o su alcance cambió, detente y regístralo antes de codificar.
6. Cambia `status` a `in_progress`, crea `feat/64-hu011-pantalla-recuperacion` desde `main` y registra la rama en este archivo y en el índice.

## Alcance incluido

- Ruta de recuperación cargada de forma diferida dentro de `apps/web/src/modules/auth`, enlazada desde el destino que ya publica `HU-010`.
- Tres pasos con indicador de paso actual y total: solicitar (correo o teléfono según lo que exponga el contrato de `HU-008`), verificar código, establecer contraseña nueva.
- Composición según `estandar-diseno-visual.md` sección 10: paso actual, destino enmascarado (usa el enmascarado que ya calcula el backend, `apps/api/internal/modules/auth/mask.go`; no reimplementes el enmascarado en el cliente), campo de código y reenvío con estado.
- Mensajes distintos y accionables para código incorrecto, código vencido, demasiados intentos y espera de reenvío con cuenta regresiva basada en el `Retry-After`/campo temporal real que devuelve el contrato.
- Política de contraseña explicada antes de escribir (no solo al fallar), acorde a la política real que valida `HU-008` en el backend.
- Confirmación final explícita de cambio de contraseña y de que las sesiones activas se cerraron, con salida al acceso (`HU-010`).
- Pruebas de componente por paso, E2E del recorrido completo (código válido y código vencido), axe-core y evidencia responsive en 320 y 360 px.

## Fuera de alcance

- Cualquier cambio al backend, migraciones, contrato OpenAPI o pruebas SQL de `HU-008`. Si el contrato integrado no permite un criterio de esta HU, registra un defecto separado en vez de tocar `HU-008` aquí.
- Cabecera, navegación o cualquier contenido del cascarón de `HU-012`.
- Cambios a la pantalla de acceso de `HU-010` más allá de confirmar que su enlace apunta a la ruta real que esta historia entrega.
- Nuevos tokens, colores, tipografías o componentes visuales fuera de los que ya expone `HU-009`; si falta una primitiva, regístralo como fix separado.
- Pinia, estado global persistente o almacenamiento de la contraseña/código fuera de memoria del formulario.

## Estado existente que debe conservarse

- `HU-008` ya entrega el backend completo (handlers, repositorio, envío WhatsApp/email, purga, enmascarado, no enumeración) — reutilízalo tal cual, sin duplicar lógica de validación de política de contraseña ni de enmascarado en el frontend.
- `HU-009` entrega `BaseButton`, `BaseInput`, `BaseAlert`, `BaseBadge`, `BaseDialog`, tokens y pruebas; reutiliza su API pública.
- `HU-010` entrega el cliente HTTP tipado derivado del bundle OpenAPI y el patrón de página coordinadora + formulario tipado sin `fetch` directo; sigue el mismo patrón para los tres pasos de esta historia.
- La dependencia de módulos permanece `app → modules → shared`; `auth` no expone internos a otros módulos.

## Trabajo requerido

1. Audita el enlace de recuperación de `HU-010` y el router; define la ruta diferida de los tres pasos dentro de `modules/auth`.
2. Modela el estado del flujo como una máquina de tres pasos explícita (unión discriminada), con navegación entre pasos y sin retroceso que pierda el progreso ya confirmado por el servidor.
3. Paso 1 (solicitar): formulario mínimo según lo que exija el contrato de `HU-008`; respuesta siempre no enumerable, sin revelar si la cuenta existe.
4. Paso 2 (verificar código): campo de código, mensajes distintos para incorrecto/vencido/agotado (mapeados por `status`/`type`/`code` del contrato, nunca por `detail`), reenvío con cuenta regresiva real y bloqueo del reenvío anticipado.
5. Paso 3 (establecer contraseña): explica la política antes de escribir; valida en cliente para ayudar, sin sustituir la validación del backend.
6. Al confirmar el cambio: pantalla de éxito que declare explícitamente el cierre de sesiones activas (según lo que documente `HU-008`) y enlace de salida al acceso de `HU-010`.
7. Accesibilidad: foco movido al encabezado de cada paso al transicionar, anunciado a lector de pantalla; operable con teclado; sin desplazamiento horizontal en 320/360 px.
8. No inventes campos, códigos de error o tiempos que el contrato de `HU-008` no exponga; si falta algo necesario para un criterio, regístralo como duda antes de simular el dato.

## Pruebas y evidencia

- Componente por paso: render inicial, validación, envío, estados de error específicos (incorrecto, vencido, agotado, espera de reenvío), doble envío bloqueado.
- Accesibilidad: `vitest-axe` sin violaciones aplicables, foco por paso, orden de teclado.
- E2E contra API real local: recorrido completo con código válido y con código vencido.
- Evidencia visual responsive a 320 y 360 px como mínimo (agrega 768/1280 si el resto de la suite ya lo hace).
- Seguridad: el código y la contraseña nueva no aparecen en URL, historial, `console`, analytics ni almacenamiento persistente.

Entrega una tabla `Criterio | Estado | Prueba o evidencia` para `CA-011-01` a `CA-011-08`.

## Documentación y trazabilidad

- Actualiza matriz de trazabilidad, historial de cambios y el README de `apps/web` si documenta módulos/rutas.
- Actualiza este prompt y `docs/10-backlog/prompts/README.md` con rama, PR y estado reales al iniciar y al terminar.
- Si `HU-009` (formal, pendiente de `HU-011`) puede cerrarse tras esta entrega, señálalo en la documentación de control sin declarar tú mismo su cierre si depende de otra decisión pendiente.

## Verificación final

```text
pnpm install --frozen-lockfile (en apps/web)
pnpm run openapi:check-config
pnpm run openapi:lint
pnpm run openapi:bundle
pnpm run format
pnpm run lint
pnpm run typecheck
pnpm run test:unit
pnpm run build
pnpm run test:e2e para el recorrido de recuperación
git diff --check
graphify update .
```

No actualices snapshots sin revisar el comportamiento, no ocultes pruebas inestables y no declares evidencia de `HU-012` o de HU posteriores que todavía no exista.

## Git y PR

- Commits y título: `feat(web): implementa pantalla de recuperación de acceso` o equivalente Conventional Commits.
- Abre un PR contra `main` con `Closes #64` solo si los ocho criterios están probados.
- Incluye pruebas, evidencia responsive y accesible, tabla de criterios y confirmación de que el backend de `HU-008` no se modificó.
- No hagas push directo, force push, merge manual ni reescribas `main`.
