---
prompt_id: "PROMPT-CHORE-LENGUAJE-REGLADO-NAVA-AUTH-v1"
version: "1.0"
kind: "chore"
status: "superseded"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-010"
related_hu:
  - "HU-005"
  - "HU-007"
  - "HU-008"
  - "HU-009"
  - "HU-011"
  - "HU-012"
issue: "pending"
issue_url: null
suggested_issue_title: "chore(web): lenguaje reglado NAVA en controles compartidos y fidelidad de acceso/recuperación"
branch: null
pr: null
pr_url: null
depends_on:
  - "PR #210 (issue #188) integrado o cerrado: toca los mismos archivos de `apps/web/src/modules/auth`"
  - "DEC-081 registrada en `docs/00-control/registro-decisiones.md`"
rules:
  - "RN-SEG-01"
  - "RN-REC-04"
decisions:
  - "DEC-039"
  - "DEC-062"
  - "DEC-063"
  - "DEC-064"
  - "DEC-065"
  - "DEC-077"
  - "DEC-078"
  - "DEC-079"
  - "DEC-080"
  - "DEC-081"
acceptance_criteria:
  - "CA-005-02"
  - "CA-007-02"
  - "CA-008-05"
  - "CA-008-06"
  - "CA-010-01"
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
  - "docs/02-requisitos/historias-usuario.md"
created_at: "2026-09-03"
updated_at: "2026-09-03"
supersedes: null
superseded_by:
  - "PROMPT-CHORE-CONTROLES-REGLADOS-SHARED-UI-v1"
  - "PROMPT-CHORE-FIDELIDAD-ACCESO-RECUPERACION-v1"
  - "PROMPT-ORCH-LENGUAJE-REGLADO-NAVA-v1"
---

# Lenguaje reglado NAVA en controles compartidos y fidelidad de acceso/recuperación

> **Sustituido el 2026-09-03.** El propietario pidió separar la entrega en dos preocupaciones
> independientes y acotar ambas a lo estrictamente visual. Este archivo se conserva por
> trazabilidad; no se ejecuta. Usa
> [PROMPT-CHORE-CONTROLES-REGLADOS-SHARED-UI-v1](issue-pending-controles-reglados-shared-ui.md),
> [PROMPT-CHORE-FIDELIDAD-ACCESO-RECUPERACION-v1](issue-pending-fidelidad-acceso-recuperacion.md)
> y su orquestación
> [PROMPT-ORCH-LENGUAJE-REGLADO-NAVA-v1](../orchestration/lenguaje-reglado-nava.md).

## Instrucción para el agente

Trabaja en modo **fidelidad de mockup**. Se te asignan 46 imágenes de referencia exactas. Implementa únicamente la preocupación descrita aquí, de forma autónoma, hasta que la app real reproduzca cada evento en su viewport. No amplíes el alcance ni inventes decisiones.

Modelo y esfuerzo requeridos: **Claude Sonnet 5 con `effortLevel: high`**. Carga y sigue las skills `nava-mockup-fidelity` y `browser-viewport-verification` antes de editar y antes de declarar terminada cualquier pantalla.

## Objetivo

`/acceso` y `/recuperar-acceso` reproducen, en escritorio (1440×1024) y móvil (420 CSS px), los 23 eventos representados en [`ui-mockups-nava-tailored-grid-2026-09-03/auth-eventos`](../../evidence/ui-mockups-nava-tailored-grid-2026-09-03/auth-eventos/README.md), y los controles compartidos de `apps/web/src/shared/ui` adoptan el lenguaje reglado que esos mockups fijan, sin romper las nueve rutas P0 restantes que los consumen.

## Referencia asignada

Directorio: `docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/auth-eventos/`.

- `desktop/acceso/01..12` y `mobile/acceso/01..12`;
- `desktop/recuperacion/01..11` y `mobile/recuperacion/01..11`;
- el `README.md` del directorio es el contrato visual normativo y la tabla de eventos.

Un mismo número describe el mismo estado funcional en ambos viewports. **No** uses `ui-mockups-nava-tailored-grid-2026-09-02/02-acceso-recuperacion.png`: es lámina histórica y está explícitamente descartada como objetivo.

Los PNG se regeneran de forma determinista con `node tools/mockups/auth-eventos/render.mjs`. Léelos como referencia; no los edites ni los regeneres como parte de esta entrega.

## Alcance incluido

1. `apps/web/src/shared/ui/BaseInput.vue`, `BaseAlert.vue` y `BaseButton.vue`: adoptar el lenguaje reglado descrito en el contrato visual.
2. `apps/web/src/shared/ui/OtpInput.vue`: componente nuevo de seis ranuras con un único valor lógico de seis dígitos.
3. `apps/web/src/modules/auth`: `LoginForm.vue`, `PhoneChallengeForm.vue`, `RecoveryRequestStep.vue`, `RecoveryVerifyStep.vue`, `RecoveryResetStep.vue`, `LoginPage.vue`, `RecoveryPage.vue`, `AuthSplitLayout.vue`.
4. Barrido de regresión visual y de pruebas en las rutas que consumen los controles modificados: `/panel`, `/panel/turnos/nuevo`, `/panel/turnos/:id`, `/panel/servicios`, `/panel/servicios-por-barbero`, `/panel/barberos`, `/panel/horarios`, `/panel/bloqueos`, `/panel/configuracion`.
5. Pruebas de componente, E2E y evidencia responsiva de lo modificado.

## Fuera de alcance

- **Configuración de canal de `DEC-081`.** Esa decisión resuelve `DP-NOT-06`, `DP-SEG-13`, `CT-009` y `CT-010`, pero declara de forma expresa que **no autoriza** implementar configuración, contrato, proveedores ni persistencia de canal dentro del alcance visual. No toques `apps/api`, el contrato OpenAPI, la base de datos ni los adaptadores de notificación.
- Rediseñar la composición de las nueve rutas P0 no representadas en estos mockups. Ahí solo corriges la regresión que introduzca tu propio cambio de `shared/ui`.
- Cambiar reglas de negocio, criterios de aceptación, endpoints o el modelo de sesión.
- Regenerar o modificar los PNG de referencia y el generador de `tools/mockups/`.

## Estado existente que debe conservarse

- **Comportamiento y seguridad.** `CA-005-02`/`CA-010-02`: un rechazo de credenciales nunca distingue correo inexistente de contraseña incorrecta. `DEC-062`/`DEC-064`/`DEC-065`: código incorrecto, vencido, agotado o de cuenta inexistente comparten respuesta uniforme. `CA-010-07`: las credenciales no se serializan a almacenamiento ni a la URL.
- **Máquinas de estado.** `LoginScreenState`, `LoginOutcome`, `ChallengeRequestOutcome`, `ChallengeVerifyOutcome`, `RecoveryRequestOutcome`, `RecoveryVerifyOutcome` y `RecoveryResetPasswordOutcome` no cambian de forma. Esta entrega es visual: no reescribas la coordinación de estado salvo lo que exija el punto 4 de *Trabajo requerido*.
- **Accesibilidad ya lograda.** `role="alert"`/`role="status"` existentes, foco al encabezado en cada paso (`CA-011-07`), foco al resumen con más de un error (`CA-010-05`), `:show-required-marker="false"` en `/acceso` (issue #188 retiró el asterisco visual), y `getByLabel(..., { exact: true })` en la suite E2E.
- **Guardia de doble envío** de `attemptLogin` y el cooldown de reenvío de 60 s (`RESEND_COOLDOWN_SECONDS`).
- **Validaciones y copys** de `validation/loginValidation.ts` y `validation/recoveryValidation.ts`: los mockups ya usan exactamente esas cadenas. Si un mockup y el código discrepan en texto, **manda el código** y se registra la diferencia; si discrepan en composición, manda el mockup.

## Trabajo requerido

### 1. Contrato medido antes de editar

Abre cada PNG a resolución original y completa la tabla de contrato de `nava-mockup-fidelity`. Captura baseline de la app real en los dos viewports antes de tocar código. Un intercambio de tokens o un refresco de capturas no es un rediseño.

### 2. Controles reglados en `shared/ui`

Implementa el lenguaje que fija el contrato visual del README de referencia:

- **`BaseInput`.** Superficie de papel blanco sobre línea base de tinta de `2 px`, filete perimetral de `1 px` a baja opacidad y radio `2 px`. Rótulo sobre el campo en versalitas espaciadas de latón, con el mismo tratamiento tipográfico de `PASO 1 DE 3` y `ACCESO SEGURO`. Sin iconos decorativos de sobre ni de candado. El conmutador de contraseña es la palabra `Mostrar`/`Ocultar` en versalitas, no un glifo de ojo; conserva `aria-pressed` o el nombre accesible equivalente que ya exponga el componente. Estados obligatorios: reposo, foco visible, error, deshabilitado.
- **`BaseAlert`.** Filete lateral de `4 px` del color de estado, fondo apenas teñido, radio `2 px`, y una palabra de estado en versalitas sobre el título: `Error`, `Atención`, `Nota`, `Confirmación`. Esa palabra es la que cumple “icono, texto y estructura además del color”; retira el glifo circular. Conserva el slot de acción y el de descarte, con el filete inferior de los demás controles.
- **`BaseButton`.** Radio `2 px`. Primario en tinta; secundario y fantasma con filete de latón y filete inferior acentuado. Conserva el estado `loading` con verbo en gerundio e indicador de progreso.
- **`OtpInput` (nuevo).** Seis ranuras que repiten la construcción del campo reglado en formato estrecho, con el dígito compuesto en la serif del wordmark apoyado en su línea base. Debe mantener **un solo valor lógico de seis dígitos** y soportar: pegado del código completo con distribución automática, avance y retroceso entre ranuras, teclado numérico (`inputmode="numeric"`, `autocomplete="one-time-code"`), foco visible y un nombre accesible coherente. No es un componente decorativo: es el control real.

Estos cuatro son contratos compartidos. No crees variantes locales con otro tono o escala dentro de `auth`.

### 3. Composición de `/acceso` y `/recuperar-acceso`

- Reproduce el cascarón: escritorio con panel de tinta y columna centrada, móvil en una sola columna. Bloque centrado verticalmente, no anclado arriba.
- Ambas rutas muestran el wordmark con la regla de latón a su ancho completo.
- **Ancla de alertas.** En `/acceso`, la alerta global aparece **después** del botón y de `¿Olvidaste tu contraseña?`. Esto implica mover el `BaseAlert` de `serverError`, que hoy se renderiza al inicio del formulario en `LoginForm.vue`.
- **Excepción de orden.** Cuando la alerta explica por qué la pantalla está en ese estado y la acción es su remedio —`10-contrasena-enlace-vencido` y `11-completado`— la alerta precede al botón. El motivo nunca va después del remedio.
- **Resumen global.** Explica qué puede hacer la persona y **no repite literalmente** cada error local. Hoy `LoginForm.vue` enumera los errores de campo tal cual; sustitúyelo por el resumen explicativo que muestran `03-validacion` y `08-contrasena-validacion`, conservando `role="alert"`, `tabindex="-1"` y el foco al resumen cuando hay más de un error.
- **Sección del reto.** El reto adicional deja de ser una tarjeta sobrepuesta: abre una sección de la misma columna con la regla de latón y rombo, seguida del canal configurado en versalitas. Durante el reto, credenciales y `Iniciar sesión` quedan bloqueadas y la acción primaria pasa al reto; nunca dos botones primarios en la misma vista. `Volver al acceso` **no** aparece dentro de `/acceso`.
- **Estado coherente con el mensaje.** Código rechazado ⇒ seis ranuras llenas y en error. Credenciales rechazadas ⇒ contraseña limpia y correo conservado (`CA-010-02`, ya implementado). Fallo de red ⇒ ambos datos conservados y botón `Reintentar` dentro de la alerta (`actionLabel` ya existe; el mockup exige que se vea). Sesión vencida ⇒ campos vacíos.

### 4. Canal del código sin implementar configuración

El backend vigente resuelve el reto por un solo canal. Los mockups representan las tres variantes de `DEC-081` (`correo` predeterminado, `WhatsApp oficial`, `ambos`).

- Modela el canal como **una prop o valor derivado del servidor**, con un único valor posible hoy, y renderiza la línea de canal a partir de él.
- **No** inventes configuración, endpoint, campo de contrato ni preferencia persistida, y **no** permitas que la persona escriba o elija un destino.
- El texto permanece condicional y uniforme (`Si la cuenta puede continuar…`), sin enumerar cuentas.
- Deja registrado en el PR que las variantes `08`/`09`/`11`/`12` quedan representadas pero no configurables hasta que exista el issue funcional de `DEC-081`.

### 5. Barrido de regresión

`BaseInput` y `BaseAlert` los consumen 17 archivos en 8 módulos (`agenda`, `auth`, `barberServices`, `catalog`, `schedules`, `settings`, `staff`). Recorre en navegador real las nueve rutas P0 no representadas, confirma que ningún formulario o alerta quedó roto, ilegible o sin contraste, y corrige la regresión que introduzca tu cambio. No rediseñes esas pantallas.

## Pruebas y evidencia

- **Componente.** `OtpInput`: pegado del código completo, avance, retroceso, teclado numérico, foco visible, nombre accesible y valor lógico único. `BaseInput`: conmutador `Mostrar`/`Ocultar` y los cuatro estados. `BaseAlert`: las cuatro variantes, la palabra de estado, el slot de acción y el de descarte.
- **Actualiza** las pruebas existentes de `shared/ui/__tests__`, `auth/components/__tests__` y `auth/pages/__tests__` que asuman la iconografía o el orden anteriores. No relajes una aserción para que pase: si el comportamiento cambió, cambia la aserción y explica por qué en el PR.
- **E2E.** `acceso.spec.ts`, `recuperacion.spec.ts`, `reto-telefonico.spec.ts` y las tres suites de evidencia responsiva de acceso, recuperación y panel.
- **Accesibilidad.** `axe-core` en vivo sobre `/acceso` y `/recuperar-acceso` en cada estado con contenido nuevo, con 0 violaciones. Verifica foco visible, orden de tabulación, 200 % de zoom y `prefers-reduced-motion`.
- **Evidencia visual.** Por cada evento: captura de la app real, comparación lado a lado con el PNG y overlay o diff. Guarda bajo `apps/web/e2e/evidence/` siguiendo la convención ya usada por las suites de evidencia responsiva. Verifica `window.innerWidth`/`innerHeight` reales antes de capturar; no confíes en un redimensionado no comprobado.
- **Anchos obligatorios.** 320, 360, 768 y 1280 px además de los dos viewports de referencia.

## Documentación y trazabilidad

- Actualiza `docs/00-control/matriz-trazabilidad.md` e `historial-cambios.md` en lo que realmente resulte afectado.
- Si el lenguaje reglado modifica el sistema visual general, actualiza `docs/03-desarrollo/estandar-diseno-visual.md` y `especificacion-frontend-nava.md`; si crees que hace falta una decisión nueva, **regístrala como duda antes de codificar**, no la resuelvas por tu cuenta.
- Actualiza los metadatos de este archivo (`status`, `issue`, `branch`, `pr`) y la fila correspondiente del índice [`docs/10-backlog/prompts/README.md`](../README.md).

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

Una fila por evento de los 23, más una por cada criterio `CA-*` de los metadatos y una por cada ruta del barrido de regresión. No declares cumplido aquello que no esté probado, y documenta toda desviación restante con su justificación.

## Git y PR

- Rama `chore/<issue>-lenguaje-reglado-nava-auth` desde `main` actualizada.
- Commit y PR: `chore(web): lenguaje reglado NAVA en controles compartidos y fidelidad de acceso/recuperación`.
- `Closes #<issue>` solo si se cubre el issue completo; en caso contrario `Refs #<issue>`.
- No hagas push directo, force push ni merge de `main`.
