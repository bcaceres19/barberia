---
prompt_id: "PROMPT-CHORE-FIDELIDAD-ACCESO-RECUPERACION-v1"
version: "1.0"
kind: "chore"
status: "ready"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-010"
related_hu:
  - "HU-005"
  - "HU-007"
  - "HU-008"
  - "HU-011"
issue: "213"
issue_url: "https://github.com/bcaceres19/barberia/issues/213"
suggested_issue_title: "chore(web): fidelidad visual de acceso y recuperación con los mockups por evento"
branch: "chore/213-fidelidad-acceso-recuperacion"
pr: null
pr_url: null
depends_on:
  - "Issue #212 (PROMPT-CHORE-CONTROLES-REGLADOS-SHARED-UI-v1) integrado en main mediante PR #214"
  - "Issue maestro #211"
  - "PR #210 (issue #188) integrado o cerrado"
rules: []
decisions:
  - "DEC-062"
  - "DEC-063"
  - "DEC-064"
  - "DEC-065"
  - "DEC-077"
  - "DEC-079"
  - "DEC-080"
  - "DEC-081"
acceptance_criteria:
  - "CA-005-02"
  - "CA-010-02"
  - "CA-010-03"
  - "CA-010-04"
  - "CA-010-05"
  - "CA-010-07"
  - "CA-011-06"
  - "CA-011-07"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/auth-eventos/README.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/00-control/registro-decisiones.md"
created_at: "2026-09-03"
updated_at: "2026-09-04"
supersedes: "PROMPT-CHORE-LENGUAJE-REGLADO-NAVA-AUTH-v1"
superseded_by: null
---

# Fidelidad visual de `/acceso` y `/recuperar-acceso`

## Instrucción para el agente

Trabaja en modo **fidelidad de mockup**. Se te asignan 46 imágenes de referencia exactas y la app real debe reproducirlas.

Esta entrega es **exclusivamente visual**: composición, geometría, jerarquía, color, densidad y estado visual. **No cambia comportamiento, validaciones, endpoints, máquinas de estado ni textos de producto.** Cuando un mockup y el código difieran en *texto* o en *comportamiento*, manda el código: registras la diferencia en la tabla de desviaciones y la deja para el ajuste funcional posterior. Cuando difieran en *composición*, manda el mockup.

Modelo y esfuerzo requeridos: **Claude Sonnet 5 con `effortLevel: high`**. Carga y sigue las skills `nava-mockup-fidelity` y `browser-viewport-verification` antes de editar y antes de declarar terminada cualquier pantalla.

Esta es la **fase 2 de dos**. La fase 1 entregó los controles reglados en `shared/ui` ([PROMPT-CHORE-CONTROLES-REGLADOS-SHARED-UI-v1](issue-pending-controles-reglados-shared-ui.md)) y debe estar integrada en `main` antes de empezar. Aquí los consumes; no los rediseñas.

## Issue

[#213 · chore(web): fidelidad visual de acceso y recuperación con los mockups por evento](https://github.com/bcaceres19/barberia/issues/213), fase 2 del issue maestro [#211](https://github.com/bcaceres19/barberia/issues/211).

La fase 1 quedó integrada en `main` mediante PR [#214](https://github.com/bcaceres19/barberia/pull/214) (commit `bd738ef`), así que esta fase está desbloqueada.

## Estado heredado de la fase 1

El PR [#214](https://github.com/bcaceres19/barberia/pull/214) entregó los controles reglados, pero cerró con verificaciones
declaradas **no realizadas**. Léelo antes de empezar y trátalo así:

**Llega sin verificación visual y cae dentro de tus propios eventos, así que lo cierras tú:**

- la variante `plain` de `BaseAlert`, que es la caja «Requisitos de la contraseña» de los
  eventos `07`, `08` y `09` de recuperación;
- el paso 3 del flujo de recuperación completo, eventos `07` a `11`.

No los des por buenos porque la fase 1 los declaró implementados: son parte de tu evidencia
por evento y se comparan contra su PNG como cualquier otro.

**Llega sin verificación y queda FUERA de tu alcance —no lo arregles, ni lo uses como
excusa para no avanzar—:**

- `/panel/turnos/:id` (detalle de turno) y el diálogo «Agregar servicio» de
  `/panel/servicios`, no recorridos en el barrido de regresión de la fase 1;
- la suite E2E completa de Playwright de las once rutas P0, que la fase 1 **no ejecutó**.

Ese último punto te afecta de forma directa: cuando corras `pnpm test:e2e` es posible que
falle algo que la fase 1 introdujo y nadie detectó. **Antes de tocar nada, ejecuta la suite
E2E sobre `main` sin tus cambios y guarda ese resultado como línea base.** Si un fallo ya
existía, no es tuyo: repórtalo en [#212](https://github.com/bcaceres19/barberia/issues/212), que sigue abierto, y no lo
arrastres a tu PR. Si aparece después de tu cambio, sí es tuyo y lo corriges.

**Token nuevo disponible:** `--color-accent-brass` en `apps/web/src/styles/tokens.css`,
que gobierna las versalitas y filetes de latón del lenguaje reglado. Úsalo; no escribas el
hexadecimal ni reutilices `--color-focus`, que coincide en valor pero no en concepto.

## Objetivo

`/acceso` y `/recuperar-acceso` reproducen los 23 eventos representados, en escritorio (1440 × 1024) y móvil (420 CSS px), con el comportamiento actual intacto.

## Referencia asignada

Directorio `docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/auth-eventos/`:

- `desktop/acceso/01..12` y `mobile/acceso/01..12`;
- `desktop/recuperacion/01..11` y `mobile/recuperacion/01..11`;
- su `README.md` es el contrato visual normativo y trae la tabla que nombra cada evento.

Un mismo número describe el mismo estado funcional en ambos viewports; termina el par antes de dar el evento por cerrado. **No** uses `ui-mockups-nava-tailored-grid-2026-09-02/02-acceso-recuperacion.png`: es lámina histórica, explícitamente descartada como objetivo. No edites los PNG ni ejecutes su generador.

## Alcance incluido

`apps/web/src/modules/auth`: `AuthSplitLayout.vue`, `LoginForm.vue`, `LoginPage.vue`, `PhoneChallengeForm.vue`, `RecoveryPage.vue`, `RecoveryRequestStep.vue`, `RecoveryVerifyStep.vue`, `RecoveryResetStep.vue`, más la adopción de `OtpInput` donde hoy hay un `BaseInput` de código, y las pruebas y evidencia de lo modificado.

## Fuera de alcance

- **Todo lo funcional.** No toques `validation/`, `api/`, `model/`, `guards/`, `bootstrap/`, `apps/api`, el contrato OpenAPI ni la base de datos. Si una diferencia visual parece exigir un cambio funcional, **no lo hagas**: anótalo en la tabla de desviaciones.
- **Configuración de canal de `DEC-081`.** Esa decisión declara de forma expresa que no autoriza implementar configuración, contrato, proveedores ni persistencia de canal dentro del alcance visual. Requiere issue funcional propio.
- Rediseñar `shared/ui`. Si un control necesita un ajuste, corrígelo en la fase 1 o abre seguimiento; no lo parchees localmente dentro de `auth`.
- Rediseñar cualquier otra ruta.

## Estado existente que debe conservarse

- **Seguridad y no enumeración.** `CA-005-02`/`CA-010-02`: el rechazo de credenciales nunca distingue correo inexistente de contraseña incorrecta. `DEC-062`/`DEC-064`/`DEC-065`: código incorrecto, vencido, agotado o de cuenta inexistente comparten respuesta uniforme. `CA-010-07`: las credenciales no se serializan a almacenamiento ni a la URL.
- **Máquinas de estado sin cambios de forma:** `LoginScreenState`, `LoginOutcome`, `ChallengeRequestOutcome`, `ChallengeVerifyOutcome`, `RecoveryRequestOutcome`, `RecoveryVerifyOutcome`, `RecoveryResetPasswordOutcome`.
- **Textos actuales** de `loginValidation.ts`, `recoveryValidation.ts`, los resúmenes de error de `LoginPage.vue` y los títulos y cuerpos de alerta de los tres pasos de recuperación. Los mockups ya usan esas cadenas: si encuentras una diferencia, gana el código.
- **Accesibilidad ya lograda:** `role="alert"`/`role="status"`, foco al encabezado en cada paso (`CA-011-07`), foco al resumen con más de un error (`CA-010-05`), `:show-required-marker="false"` en `/acceso`, y `getByLabel(..., { exact: true })` en la suite E2E.
- **Guardia de doble envío** de `attemptLogin` y el cooldown de reenvío de 60 s.

## Trabajo requerido

### 1. Contrato medido antes de editar

Abre cada PNG a resolución original —no una miniatura— y completa la tabla de contrato de `nava-mockup-fidelity` por evento. Captura baseline de la app real en los dos viewports antes de tocar código.

### 2. Cascarón y encabezado

- Escritorio: panel de tinta a la izquierda con wordmark, divisor de latón con rombo, lema y pie en versalitas; columna de contenido centrada a la derecha, **centrada verticalmente**, no anclada arriba.
- Móvil: una sola columna. Nunca se miniaturiza una vista dentro de la otra.
- **Ambas rutas** muestran el wordmark con la regla de latón a su ancho completo. Hoy `/recuperar-acceso` no lo tiene.
- Recuperación conserva `Paso N de 3` en versalitas de latón sobre el título del paso.

### 3. Ancla de alertas

- En `/acceso`, la alerta global va **después** del botón y de `¿Olvidaste tu contraseña?`. Esto implica mover el `BaseAlert` de `serverError`, que hoy se renderiza al inicio del formulario en `LoginForm.vue`. Es un cambio de orden en la plantilla, no de lógica.
- En recuperación se conserva el mismo patrón, después del grupo de acciones y de `Volver al acceso` cuando ese enlace exista.
- **Excepción:** cuando la alerta explica por qué la pantalla está en ese estado y la acción es su remedio —`10-contrasena-enlace-vencido` y `11-completado`— la alerta precede al botón. El motivo nunca va después del remedio.

### 4. Sección del reto en `/acceso`

- El reto deja de ser una tarjeta sobrepuesta: abre una sección de la misma columna con la regla de latón y rombo, seguida de la línea de canal en versalitas y su nota.
- Con el reto activo, credenciales y `Iniciar sesión` se muestran bloqueados y la acción primaria pasa al reto. **Nunca dos botones primarios en la misma vista.**
- `Volver al acceso` **no** aparece dentro de `/acceso`.
- Adopta `OtpInput` en lugar del `BaseInput` de código, conservando el mismo valor de seis dígitos que el componente ya entrega.
- Los mockups `07`–`12` representan tres canales. El backend vigente resuelve uno solo. Renderiza la línea de canal desde un valor derivado del servidor con un único valor posible hoy; **no** inventes configuración, endpoint, campo de contrato ni preferencia persistida, y **no** permitas escribir o elegir un destino. Deja las variantes restantes anotadas como pendientes del issue funcional de `DEC-081`.
- Los estados del reto que los mockups **no** representan —por ejemplo la fase previa a solicitar el código— conservan su composición actual dentro de NAVA. No inventes su diseño.

### 5. Adopción de `OtpInput` en recuperación

`RecoveryVerifyStep.vue` sustituye su `BaseInput` de código por `OtpInput`, con el mismo `modelValue`, el mismo error y el mismo rótulo. Eventos `05` y `06`.

### 6. Coherencia entre estado y mensaje

Los mockups fijan qué muestra cada evento; todo esto ya es el comportamiento actual, así que es cuestión de reflejarlo, no de cambiarlo:

- código rechazado ⇒ seis ranuras **llenas** y en error;
- credenciales rechazadas ⇒ contraseña limpia, correo conservado (`CA-010-02`);
- fallo de red ⇒ ambos datos conservados y botón `Reintentar` **visible** dentro de la alerta (`actionLabel` ya existe en el resumen; el mockup exige que se vea);
- sesión vencida ⇒ campos vacíos, alerta `info` descartable;
- envío en curso ⇒ todos los campos de la operación y la acción principal bloqueados, verbo en gerundio con indicador de progreso y geometría conservada.

### 7. Desviación conocida a registrar, no a implementar

Los eventos `03-validacion` y `08-contrasena-validacion` muestran un resumen global que **explica** qué falta, mientras que `LoginForm.vue` hoy **enumera** literalmente cada error de campo. Es una diferencia de texto: **conserva el comportamiento actual** y regístrala en la tabla de desviaciones para el ajuste funcional posterior. Sí adopta la posición, el color y la forma que fija el mockup.

## Pruebas y evidencia

- **Componente.** Actualiza `auth/components/__tests__` y `auth/pages/__tests__` a la nueva composición: orden de la alerta global, sección del reto, adopción de `OtpInput`, bloqueo durante el reto. Si una aserción deja de aplicar, cámbiala y explica por qué en el PR; no la relajes ni la elimines para que pase.
- **E2E.** `acceso.spec.ts`, `recuperacion.spec.ts`, `reto-telefonico.spec.ts`, `acceso-evidencia-responsiva.spec.ts` y `recuperacion-evidencia-responsiva.spec.ts`.
- **Accesibilidad.** `axe-core` en vivo sobre ambas rutas en cada estado con contenido nuevo, con 0 violaciones. Verifica foco visible, orden de tabulación, 200 % de zoom y `prefers-reduced-motion`.
- **Evidencia visual, obligatoria por evento.** Los 23 eventos × 2 viewports: captura de la app real, comparación lado a lado con el PNG y overlay o diff. Guarda bajo `apps/web/e2e/evidence/` con la convención de las suites de evidencia responsiva. Verifica `window.innerWidth`/`innerHeight` reales antes de capturar; no confíes en un redimensionado no comprobado.
- **Anchos obligatorios adicionales:** 320, 360, 768 y 1280 px.

## Documentación y trazabilidad

- Actualiza `docs/00-control/matriz-trazabilidad.md` e `historial-cambios.md` en lo que realmente resulte afectado.
- Entrega una **tabla de desviaciones** con toda diferencia entre mockup y app que no se cerró, su motivo y a qué issue futuro pertenece. Las diferencias de texto y comportamiento van aquí, no al código.
- Actualiza los metadatos de este archivo (`status`, `issue`, `branch`, `pr`) y su fila en el índice [`docs/10-backlog/prompts/README.md`](../README.md).

## Verificación final

```text
cd apps/web
pnpm format
pnpm lint
pnpm typecheck
pnpm test:unit
pnpm test:e2e
pnpm build
```

Entrega una tabla:

`Criterio | Estado | Prueba o evidencia`

Una fila por cada uno de los 23 eventos en cada viewport (46 en total) y una por cada criterio `CA-*` de los metadatos. No declares cumplido aquello que no esté probado, y no cierres el evento en un viewport sin haber cerrado su hermano.

## Git y PR

- Rama `chore/213-fidelidad-acceso-recuperacion` desde `main` actualizada.
- Commit y PR: `chore(web): fidelidad visual de acceso y recuperación con los mockups por evento`.
- `Closes #213` solo si se cubre el issue completo; en caso contrario `Refs #213`.
- No hagas push directo, force push ni merge de `main`.
