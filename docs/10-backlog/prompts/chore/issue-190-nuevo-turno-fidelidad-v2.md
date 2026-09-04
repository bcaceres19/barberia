---
prompt_id: "PROMPT-CHORE-190-NUEVO-TURNO-FIDELIDAD-v2"
version: "2.0"
kind: "chore"
status: "ready"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-061"
related_hu:
  - "HU-009"
  - "HU-012"
issue: 190
issue_url: "https://github.com/bcaceres19/barberia/issues/190"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on:
  - "Revisión 2026-09-04 del atlas nuevo-turno-eventos integrada en main (segundo pase del formulario)"
  - "Issue #189 integrado en main con CI verde"
  - "Fundaciones y shell de #186/#187 integrados"
  - "Issue maestro #184 (PROMPT-ORCH-NAVA-ATLAS-v3)"
rules:
  - "RN-CIT-01"
  - "RN-DAT-01"
  - "RN-IDE-01"
  - "RN-TEN-01"
decisions:
  - "DEC-071"
  - "DEC-072"
  - "DEC-073"
  - "DEC-077"
  - "DEC-079"
  - "DEC-080"
acceptance_criteria:
  - "CA-061-01"
  - "CA-061-02"
  - "CA-061-03"
  - "CA-061-04"
  - "CA-061-05"
  - "CA-061-06"
  - "CA-061-07"
  - "CA-061-08"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/nuevo-turno-eventos/README.md"
  - "docs/10-backlog/prompts/chore/issue-189-agenda-diaria-fidelidad-v4.md"
  - "docs/10-backlog/prompts/orchestration/redisenio-integral-nava-tailored-grid-v3.md"
  - "docs/10-backlog/prompts/hu/hu-061-creacion-manual-citas.md"
  - "apps/web/src/modules/agenda"
  - "apps/web/e2e/panel-fidelidad-mock.spec.ts"
created_at: "2026-09-04"
updated_at: "2026-09-04"
supersedes: "PROMPT-CHORE-190-NUEVO-TURNO-FIDELIDAD-v1"
superseded_by: null
---

# Fidelidad de `/panel/turnos/nuevo` al atlas NAVA (segundo pase)

## Instrucción para el agente

Implementa únicamente la fidelidad visual de `/panel/turnos/nuevo` contra el
atlas por evento [`nuevo-turno-eventos`](../../evidence/ui-mockups-nava-tailored-grid-2026-09-03/nuevo-turno-eventos/README.md),
evento por evento y viewport por viewport. Trabaja de forma autónoma hasta
producir la entrega definida, sin ampliar el alcance ni inventar decisiones.

Ejecuta con Claude Sonnet 5 en esfuerzo `high` y carga antes de editar las
skills `.claude/skills/nava-mockup-fidelity/SKILL.md` y
`.claude/skills/browser-viewport-verification/SKILL.md`. Este prompt asigna
una referencia exacta: aplica el modo de fidelidad de `AGENTS.md` §Calidad,
no la libertad compositiva.

## Por qué existe la v2

La v1 apuntaba a la primera versión del atlas, que heredaba de la lámina
compuesta un formulario de tarjetas de pergamino con campos blancos, una
espera reducida a un indicador suelto en pantalla vacía, un éxito como
alerta pequeña arriba a la izquierda y una columna de `900 px` centrada,
desalineada del encabezado de `/panel`. El atlas se regeneró el 2026-09-04
con un segundo pase sobre el formulario. **El objetivo de implementación son
los PNG vigentes en `main`, no los de la v1.** Si el árbol de trabajo local
todavía tiene esa revisión sin integrar, intégrala o descártala antes de
empezar: no la mezcles con la implementación.

## Objetivo

`/panel/turnos/nuevo` reproduce, en escritorio `1440×1024` y móvil
`420×935 @2x`, los doce eventos del atlas —composición, color, proporción,
jerarquía tipográfica, escala de controles e iconos, alineación, densidad y
estados—, conservando intacto el comportamiento de `HU-061`.

## Preflight obligatorio

1. Comprueba que el árbol esté limpio y no sobrescribas cambios ajenos.
2. Actualiza `main` por fast-forward y verifica que el atlas vigente incluye
   la revisión del 2026-09-04 (hoja continua, filete de latón solo en campos
   resueltos, CTA sangrado a la columna de campos). Si no está, detente.
3. Ejecuta `graphify query` sobre la zona antes de leer código en crudo.
4. Lee completos `AGENTS.md`, `CLAUDE.md` y cada `source_docs`.
5. Registra cualquier contradicción entre atlas, código y normativa antes de
   escribir código. No la resuelvas inventando.
6. Crea la rama `chore/190-nuevo-turno-fidelidad` desde `main` actualizada.
   Si ya existe local con trabajo previo, resuélvelo primero por separado.

## Alcance incluido

- `apps/web/src/modules/agenda/pages/NewAppointmentPage.vue` y los
  componentes que esa ruta necesite (por ejemplo un esqueleto propio,
  hermano de `AgendaSkeleton`).
- Reutilización de `shared/ui` ya existente: `PageHeader`, `PageState`,
  `BaseInput`, `BaseAlert`, `BaseButton`, `BaseSpinner`, `BarberAvatar`.
- Harness mock de evidencia en `apps/web/e2e` y su evidencia asociada.

## Fuera de alcance

- `apps/api`, OpenAPI, cliente generado, migraciones, campos, obligatoriedad
  o reglas de reconciliación.
- Disponibilidad pública, recordatorios, pagos, cambio de barbero posterior o
  nuevas acciones de cita.
- El issue #191 (detalle/historial/reprogramación) y cualquier otra pantalla.
- Mocks dentro de `src/`: los datos sintéticos viven solo en `apps/web/e2e`.

## Estado existente que debe conservarse

- Contrato y comportamiento de `HU-061`: campos, obligatoriedad, validación
  por campo, dependencia servicio→barbero, `Idempotency-Key`
  (`RN-IDE-01`), y conservación de datos ante `409`/`422`/fallo de red.
- `hasSummaryContent`: el resumen aparece con UNA selección hecha y cada
  entrada solo si su dato existe. No lo conviertas en un bloque fijo.
- Semántica accesible: `label` real por control, `role="alert"` en errores,
  `aria-live` en estados de carga y éxito, foco visible y orden de tabulación.
- Navegación, «Cerrar sesión» y accesibilidad del shell: no se eliminan
  porque un PNG estático no los muestre.
- `apps/web/src/modules/agenda/pages/__tests__/NewAppointmentPage.test.ts` y
  `apps/web/e2e/nuevo-turno.spec.ts` siguen pasando; ajusta solo aserciones
  que describan la nueva geometría, nunca para ocultar una diferencia.

## Contrato visual que hay que reproducir

El README del atlas manda; esto es el resumen operativo de lo no negociable:

1. **Hoja continua, no tarjetas.** Las cuatro secciones son bandas de una
   sola superficie translúcida con filete de latón de `3 px` a la izquierda y
   separadas entre sí por línea de `1 px`. Nada de pergamino en el cromo del
   formulario: el pergamino es material de registro (ficha del evento `12`).
2. **Bandas a dos columnas en escritorio** (rótulo numerado `288 px` +
   campos) y a una en móvil. El número de sección es un cuadrado de radio
   `2 px` con filete de latón y cifra en serif.
3. **Campo reglado sobre tinta**: superficie translúcida, contorno tenue,
   filete inferior de `2 px`, radio `2 px`, rótulo en versalitas de latón y
   asterisco de obligatorio en latón. `BaseInput` ya expone `--input-bg`,
   `--input-border-color` y `--input-border-base-color`: la variante sobre
   tinta se resuelve redefiniendo esos tokens en el ámbito de la página. Si
   propones tocar `shared/ui`, mide antes la regresión en las demás pantallas
   y justifícalo.
4. **El latón del filete marca el campo resuelto**; vacío o deshabilitado
   lleva filete neutro. Compara `04` (todo neutro), `05` (solo barbero) y
   `08` (todo en latón).
5. **CTA sangrado hasta la columna de campos**, misma medida que gobierna la
   columna de rótulos; en móvil ocupa el ancho completo.
6. **Resumen**: título «Resumen» y solo Barbero, Servicio, Persona atendida,
   Fecha y hora. En escritorio, columna lateral de `340 px` encabezada por la
   alerta global cuando exista; en móvil, entre el último campo y el CTA
   (§7.3). El barbero va con su monograma (`BarberAvatar`).
7. **Estados de página**: la espera inicial es esqueleto con la geometría del
   formulario más rótulo en versalitas de latón; error de contexto y vacío se
   componen centrados con divisor, rótulo, titular en serif, cuerpo y acción
   real, resaltando en latón el destino nombrado (`"Barberos"`).
8. **Éxito a pantalla completa** con ficha de pergamino, no alerta pequeña.
9. **Densidad**: en escritorio los doce eventos entran en `1024 px` de alto.
   Si tu implementación exige scroll, la densidad todavía no coincide.
10. **Copy y datos**: los textos son los del módulo real; el contador de la
    nota refleja el límite real de `validateCustomerNote` (500) y la fecha y
    hora del resumen usan `formatCivilDateFull` más el valor de `type="time"`.

## Trabajo requerido

1. Captura baseline de la app real en los doce estados y los dos viewports
   antes de tocar nada.
2. Corrige primero el evento `04` (formulario vacío) en ambos viewports:
   hoja, bandas, rejilla, campo reglado, rótulos y CTA. Es el evento que fija
   la geometría de los demás.
3. Sigue con `08` (completo, con resumen), `01` (esqueleto), `09` y `11`
   (errores en dos niveles) y `12` (éxito). Cierra con `02`, `03`, `05`,
   `06`, `07` y `10`.
4. Crea `apps/web/e2e/nuevo-turno-fidelidad-mock.spec.ts` con el mismo
   criterio que `panel-fidelidad-mock.spec.ts`: datos y reloj sintéticos solo
   dentro de Playwright, un escenario por evento con el MISMO nombre y número
   del atlas, capturando a `evidence/nuevo-turno/fidelidad-190/mock/` en
   `1440×1024` y `420×935 @2x`.
5. Compara cada par contra su PNG (lado a lado más overlay o diff) y corrige
   por ciclos cortos. No declares un evento cerrado sin su viewport hermano.
6. Registra toda desviación no evitable en
   `apps/web/e2e/evidence/nuevo-turno-fidelidad-190-desviaciones.md` con
   `Diferencia | Autoridad | Tratamiento`.

## Pruebas y evidencia

- Pruebas de componente de `NewAppointmentPage`: doble envío bloqueado,
  errores por campo, alerta global, conservación de datos ante `409` y
  aparición condicional del resumen.
- `nuevo-turno.spec.ts` y la nueva spec mock en verde.
- Evidencia responsive en `320/360/768/1280`, zoom `200 %`, recorrido de
  teclado con foco visible y `prefers-reduced-motion`.
- `vitest-axe` sobre los estados con alerta visible y sobre el éxito.

## Documentación y trazabilidad

- Actualiza este prompt y el índice de `docs/10-backlog/prompts/README.md`
  con rama, PR y estado reales.
- Si el atlas y el código real se contradicen en texto o comportamiento,
  manda el código y queda registrado en el archivo de desviaciones; si se
  contradicen en composición, manda el atlas.

## Verificación final

```text
pnpm --dir apps/web format
pnpm --dir apps/web lint
pnpm --dir apps/web typecheck
pnpm --dir apps/web test:unit
pnpm --dir apps/web test:e2e -- nuevo-turno.spec.ts nuevo-turno-fidelidad-mock.spec.ts
pnpm --dir apps/web build
```

Entrega dos tablas: `Criterio | Estado | Evidencia` (una fila por evento y
viewport) y `Diferencia | Autoridad | Tratamiento`. No declares cumplido
aquello que no esté probado ni comparado visualmente.

## Git y PR

- Commit/PR: `chore(web): reproduce Nuevo turno según el atlas NAVA`.
- `Closes #190` solo con toda la evidencia; en otro caso `Refs #190`.
- No hagas push directo, force push ni merge de `main`. No mezcles #191.
