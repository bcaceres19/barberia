---
prompt_id: "PROMPT-CHORE-CONTROLES-REGLADOS-SHARED-UI-v1"
version: "1.0"
kind: "chore"
status: "executed"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-009"
related_hu:
  - "HU-010"
  - "HU-011"
  - "HU-012"
issue: "212"
issue_url: "https://github.com/bcaceres19/barberia/issues/212"
suggested_issue_title: "chore(web): controles reglados NAVA en shared/ui"
branch: "chore/212-controles-reglados-shared-ui"
pr: null
pr_url: null
depends_on:
  - "Issue maestro #211"
  - "PR #210 (issue #188) integrado o cerrado"
rules: []
decisions:
  - "DEC-039"
  - "DEC-077"
  - "DEC-078"
  - "DEC-079"
  - "DEC-080"
acceptance_criteria:
  - "CA-009-07"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/auth-eventos/README.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
created_at: "2026-09-03"
updated_at: "2026-09-03"
supersedes: "PROMPT-CHORE-LENGUAJE-REGLADO-NAVA-AUTH-v1"
superseded_by: null
---

# Controles reglados NAVA en `shared/ui`

## Instrucción para el agente

Esta entrega es **exclusivamente visual**. No cambia comportamiento, reglas de negocio, textos de producto, contratos ni máquinas de estado. Implementa únicamente la preocupación descrita aquí, de forma autónoma, sin ampliar el alcance ni inventar decisiones.

Modelo y esfuerzo requeridos: **Claude Sonnet 5 con `effortLevel: high`**. Carga y sigue las skills `nava-mockup-fidelity` y `browser-viewport-verification` antes de editar y antes de declarar terminado cualquier componente.

Esta es la **fase 1 de dos**. La fase 2 (composición de `/acceso` y `/recuperar-acceso`) vive en [PROMPT-CHORE-FIDELIDAD-ACCESO-RECUPERACION-v1](issue-pending-fidelidad-acceso-recuperacion.md) y **no** forma parte de este issue. No adelantes su trabajo.

## Issue

[#212 · chore(web): controles reglados NAVA en shared/ui](https://github.com/bcaceres19/barberia/issues/212), fase 1 del issue maestro [#211](https://github.com/bcaceres19/barberia/issues/211).

## Objetivo

Los controles compartidos de `apps/web/src/shared/ui` hablan el lenguaje reglado editorial de NAVA, y las once rutas P0 que los consumen siguen funcionando, siendo legibles y pasando sus pruebas.

## Referencia visual asignada

El contrato normativo es la sección **“Contrato visual común”** de [`ui-mockups-nava-tailored-grid-2026-09-03/auth-eventos/README.md`](../../evidence/ui-mockups-nava-tailored-grid-2026-09-03/auth-eventos/README.md), en sus entradas *Campo reglado*, *Ranuras de código* y *Alerta como nota al margen*.

Como apoyo visual del aspecto final de cada control, usa estos PNG del mismo directorio:

| Control | Estados | Archivos |
| --- | --- | --- |
| Campo: reposo y relleno | reposo, con valor | `desktop/acceso/01-inicial.png`, `desktop/acceso/05-sin-conexion.png` |
| Campo: error | error con mensaje bajo el campo | `desktop/acceso/03-validacion.png`, `desktop/recuperacion/08-contrasena-validacion.png` |
| Campo: deshabilitado | envío en curso | `desktop/acceso/02-enviando.png` |
| Conmutador de contraseña | `Mostrar` en versalitas | `desktop/recuperacion/07-contrasena-inicial.png` |
| Ranuras de código | vacías y llenas en error | `desktop/acceso/07-reto-envio-correo.png`, `desktop/acceso/10-reto-codigo-invalido-correo.png` |
| Alerta `danger` | con título y cuerpo | `desktop/acceso/04-credenciales-invalidas.png` |
| Alerta `warning` con acción | botón `Reintentar` dentro | `desktop/acceso/05-sin-conexion.png` |
| Alerta `info` descartable | con `Descartar` | `desktop/acceso/06-sesion-vencida.png` |
| Alerta `success` | confirmación | `desktop/recuperacion/11-completado.png` |
| Alerta sin relleno | caja de requisitos | `desktop/recuperacion/07-contrasena-inicial.png` |
| Botones | primario, deshabilitado, cargando, fantasma | `desktop/acceso/01-inicial.png`, `desktop/acceso/02-enviando.png`, `desktop/acceso/07-reto-envio-correo.png` |

Los PNG son referencia de lectura. No los edites ni ejecutes su generador (`tools/mockups/auth-eventos/render.mjs`).

## Alcance incluido

1. `apps/web/src/shared/ui/BaseInput.vue`
2. `apps/web/src/shared/ui/BaseAlert.vue`
3. `apps/web/src/shared/ui/BaseButton.vue`
4. `apps/web/src/shared/ui/OtpInput.vue` (nuevo) y su export en `index.ts`
5. `apps/web/src/styles/tokens.css` solo si el lenguaje reglado exige un token nuevo o el ajuste de uno existente
6. Corrección de la regresión que este cambio introduzca en cualquier pantalla consumidora
7. Pruebas de componente y evidencia visual de los cuatro controles

## Fuera de alcance

- **Cualquier cambio de composición en pantallas.** No muevas bloques, no cambies el orden de alertas, no reubiques botones ni enlaces. Eso es la fase 2.
- **Cualquier cambio funcional.** No toques validaciones, máquinas de estado, endpoints, `apps/api`, contrato OpenAPI ni base de datos.
- **Cualquier cambio de copy de producto.** El único texto nuevo autorizado es el que pertenece al propio control: el rótulo `Mostrar`/`Ocultar` del conmutador de contraseña, la palabra de estado de la alerta (`Error`, `Atención`, `Nota`, `Confirmación`) y el rótulo `Descartar`. No reescribas mensajes de validación, títulos ni cuerpos de alerta: los recibe el componente por prop o slot.
- Sustituir `OtpInput` en las pantallas que hoy usan un `BaseInput` de código. El componente se crea y se prueba en esta fase; **se adopta en la fase 2**.
- Rediseñar `BaseBadge`, `BaseDialog`, `EmptyState`, `PageHeader`, `RecordRow` o `NavaWordmark`, salvo la corrección mínima si heredan un estilo roto por este cambio.

## Estado existente que debe conservarse

- **API pública de cada componente.** Props, eventos, slots y nombres no cambian, salvo adiciones estrictamente necesarias y retrocompatibles. Las 17 pantallas consumidoras no deben requerir edición para seguir compilando.
- **`:show-required-marker="false"`** y el resto del contrato de `BaseInput` que el issue #188 dejó estable.
- **Accesibilidad.** Nombre accesible del conmutador de contraseña y su estado presionado; `role` de las alertas gobernado por quien las usa; foco visible en todos los controles; objetivo táctil mínimo de 44 × 44 px.
- **Estado `loading`** de `BaseButton` con indicador de progreso, y su apagado bajo `prefers-reduced-motion`.
- **Gobierno por tokens.** Los componentes consumen variables semánticas, nunca un hexadecimal ni un número suelto. Si necesitas un valor que no existe, añade el token y actualiza `docs/03-desarrollo/estandar-diseno-visual.md`.

## Trabajo requerido

### 1. Baseline antes de editar

Captura los cuatro controles en su estado actual, en las pantallas donde ya viven, a 1440 y 390 px. Completa la tabla de contrato de `nava-mockup-fidelity` con geometría, color, tipografía y espaciado medidos sobre los PNG de referencia.

### 2. `BaseInput` — campo reglado

Superficie de papel blanco apoyada en una línea base de tinta de `2 px`, con filete perimetral de `1 px` a baja opacidad y radio `2 px`. El rótulo va sobre el campo en versalitas espaciadas de latón, con el mismo tratamiento tipográfico de `PASO 1 DE 3` y `ACCESO SEGURO`. El valor y el conmutador se apoyan en la línea base, no centrados en la caja.

Retira los iconos decorativos de sobre y candado: el rótulo ya nombra el campo. El slot `leading` se conserva en la API pero deja de usarse en `auth`; si otra pantalla lo usa hoy, decide con evidencia si mantenerlo o retirarlo en su propia fase, y déjalo documentado.

El conmutador de contraseña pasa de glifo de ojo a la palabra `Mostrar`/`Ocultar` en versalitas de latón, con filete inferior. Conserva su nombre accesible y su estado presionado.

Estados obligatorios, todos verificados: reposo, foco visible, relleno, error, deshabilitado.

### 3. `OtpInput` — ranuras de código (componente nuevo)

Seis ranuras que repiten la construcción del campo reglado en formato estrecho, con el dígito compuesto en la serif del wordmark apoyado en su línea base. No es un decorado: es el control real, y debe cumplir el contrato del README de referencia.

- Mantiene **un solo valor lógico de seis dígitos** hacia afuera (`modelValue` string), no seis campos independientes.
- Soporta pegar el código completo con distribución automática entre ranuras.
- Soporta avance al escribir, retroceso al borrar, y flechas izquierda/derecha.
- `inputmode="numeric"`, `autocomplete="one-time-code"`, solo dígitos.
- Foco visible por ranura y un nombre accesible coherente para el grupo, con el rótulo asociado.
- Estados: reposo, foco, relleno, error, deshabilitado.

Estas interacciones no son “funcionalidad de producto”: son lo que hace que el control exista. Sin ellas la ranura es una imagen.

### 4. `BaseAlert` — nota al margen

Filete lateral de `4 px` del color de estado sobre fondo apenas teñido, radio `2 px`. Sobre el título aparece la palabra de estado en versalitas: `Error` (`danger`), `Atención` (`warning`), `Nota` (`info`), `Confirmación` (`success`).

Esa palabra es la que cumple “icono, texto y estructura además del color” (WCAG 2.2 AA, 1.4.1). Retira el glifo circular anterior. Conserva el slot de acción —con el filete inferior de los demás controles— y el control de descarte, ahora rotulado `Descartar`.

Añade una variante **sin relleno** para contenido informativo que no comunica un estado del sistema, con filete lateral de latón: es la que usa la caja de requisitos de contraseña. En esa variante, el título se compone como versalita y no se repite como titular.

### 5. `BaseButton`

Radio `2 px`. Primario en tinta; secundario y fantasma con filete de latón y filete inferior acentuado. Conserva alturas, objetivo táctil, estado `loading` y estado deshabilitado.

### 6. Barrido de regresión

`BaseInput` y `BaseAlert` los consumen 17 archivos en 8 módulos: `agenda`, `auth`, `barberServices`, `catalog`, `schedules`, `settings`, `staff` y el shell privado.

Recorre en navegador real las once rutas P0 y confirma en cada una que ningún formulario, alerta o botón quedó roto, ilegible, sin contraste o desalineado. Corrige la regresión que introduzca tu cambio. **No rediseñes esas pantallas**: si una diferencia estaba ahí antes de tu cambio, se registra y se deja para su propia fase.

## Pruebas y evidencia

- **Componente.** Amplía `apps/web/src/shared/ui/__tests__` con: los cinco estados de `BaseInput`; el conmutador `Mostrar`/`Ocultar` con su nombre accesible y estado presionado; las cinco variantes de `BaseAlert` incluida la palabra de estado, el slot de acción y el descarte; los estados de `BaseButton`; y para `OtpInput` el pegado completo, el avance, el retroceso, las flechas, el filtrado a dígitos, el valor lógico único y el nombre accesible.
- **Actualiza** las pruebas existentes que asuman la iconografía anterior. Si una aserción deja de aplicar, cámbiala y explica por qué en el PR. No relajes ni elimines una prueba para que pase.
- **Accesibilidad.** `vitest-axe` sobre cada componente en cada estado, con 0 violaciones. Verifica contraste de texto y de los filetes que portan significado, foco visible, orden de tabulación y 200 % de zoom.
- **E2E.** Ejecuta la suite completa. Las suites de las once rutas deben seguir en verde; corrige los selectores que dependan de la iconografía retirada, conservando `getByLabel(..., { exact: true })`.
- **Evidencia visual.** Por control y estado: captura actual, comparación lado a lado con el PNG de apoyo y overlay o diff. Guarda bajo `apps/web/e2e/evidence/` siguiendo la convención de las suites de evidencia responsiva.
- **Anchos obligatorios.** 320, 360, 768 y 1280 px. Verifica `window.innerWidth`/`innerHeight` reales antes de capturar.

## Documentación y trazabilidad

- Actualiza `docs/03-desarrollo/estandar-diseno-visual.md` con el lenguaje reglado de controles y todo token nuevo.
- Actualiza `docs/00-control/matriz-trazabilidad.md` e `historial-cambios.md` en lo que realmente resulte afectado.
- Si concluyes que este lenguaje debe regir todo el producto y no solo estos controles, **regístralo como duda**; no lo resuelvas por tu cuenta ni lo extiendas a otras pantallas.
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

Una fila por control y estado, y una por cada una de las once rutas del barrido de regresión. No declares cumplido aquello que no esté probado.

## Git y PR

- Rama `chore/212-controles-reglados-shared-ui` desde `main` actualizada.
- Commit y PR: `chore(web): controles reglados NAVA en shared/ui`.
- `Closes #212` solo si se cubre el issue completo; en caso contrario `Refs #212`.
- No hagas push directo, force push ni merge de `main`.
