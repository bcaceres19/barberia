---
prompt_id: "PROMPT-TEST-QA-B2-B3-CHROME-v1"
version: "1.0"
kind: "test"
status: "superseded"
target_agents:
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu:
  - "HU-040"
  - "HU-041"
  - "HU-042"
  - "HU-060"
  - "HU-061"
  - "HU-062"
  - "HU-063"
  - "HU-064"
  - "HU-065"
issue: null
issue_url: null
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on:
  - "HU-040 integrada en main mediante PR #93 (cumplido)"
  - "HU-041 integrada en main mediante PR #96 (cumplido)"
  - "HU-042 integrada en main mediante PR #99 (cumplido, UI parcial: ver Fuera de alcance)"
  - "HU-060 integrada en main mediante PR #105 (cumplido, sin pantalla propia)"
  - "HU-061 integrada en main mediante PR #109 (cumplido)"
  - "HU-062 integrada en main mediante PR #113 (cumplido)"
  - "HU-063 integrada en main mediante PR #118 (cumplido)"
  - "HU-064 integrada en main mediante PR #121 (cumplido)"
  - "HU-065 integrada en main mediante PR #125 (cumplido)"
rules:
  - "RN-TEN-01"
  - "RN-DIS-04"
  - "RN-DIS-05"
  - "RN-DIS-07"
  - "RN-BLQ-01"
  - "RN-BLQ-02"
  - "RN-BLQ-03"
  - "RN-BLQ-04"
  - "RN-CON-01"
  - "RN-CON-03"
  - "RN-CIT-01"
  - "RN-CIT-02"
  - "RN-CIT-03"
  - "RN-RES-01"
  - "RN-RES-02"
  - "RN-RES-03"
  - "RN-HIS-01"
  - "RN-HIS-02"
  - "RN-DAT-01"
  - "RN-DAT-02"
  - "RN-IDE-01"
decisions:
  - "DEC-002"
  - "DEC-004"
  - "DEC-007"
  - "DEC-014"
  - "DEC-016"
  - "DEC-019"
  - "DEC-020"
  - "DEC-024"
  - "DEC-033"
  - "DEC-034"
  - "DEC-035"
  - "DEC-036"
  - "DEC-037"
  - "DEC-038"
  - "DEC-039"
  - "DEC-040"
  - "DEC-041"
  - "DEC-043"
  - "DEC-045"
  - "DEC-046"
  - "DEC-070"
  - "DEC-071"
  - "DEC-072"
  - "DEC-073"
  - "DEC-074"
  - "DEC-075"
  - "DEC-076"
acceptance_criteria:
  - "CA-040-01 a CA-040-08"
  - "CA-041-01 a CA-041-08"
  - "CA-042-01 a CA-042-08 (parcial: ver Fuera de alcance)"
  - "CA-061-01 a CA-061-08"
  - "CA-062-01 a CA-062-08"
  - "CA-063-01 a CA-063-08"
  - "CA-064-01 a CA-064-08"
  - "CA-065-01 a CA-065-08"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "apps/api/README.md"
  - "apps/web/README.md"
  - "apps/web/e2e/horarios.spec.ts"
  - "apps/web/e2e/excepciones-festivos.spec.ts"
  - "apps/web/e2e/nuevo-turno.spec.ts"
  - "apps/web/e2e/agenda-diaria.spec.ts"
  - "apps/web/e2e/agenda-navegacion-fecha.spec.ts"
  - "apps/web/e2e/agenda-detalle-historial.spec.ts"
  - "apps/web/e2e/agenda-reprogramacion-turno.spec.ts"
  - "apps/web/src/modules/schedules/pages/SchedulesPage.vue"
  - "apps/web/src/modules/schedules/pages/BlocksPage.vue"
created_at: "2026-09-01"
updated_at: "2026-09-02"
execution_report: null
supersedes: null
superseded_by: "PROMPT-TEST-QA-PLATAFORMA-B0-B3-LUNA-v2"
---

# QA manual en navegador real de B2 y B3 (HU-040–HU-042, HU-060–HU-065)

## Instrucción para el agente

Ejecuta, con un MCP de Chrome que controle un navegador real (o Playwright en modo headed si no hay MCP de Chrome disponible en tu entorno), los recorridos descritos abajo contra la aplicación corriendo en local. Este prompt **no autoriza cambios en el repositorio**: es verificación manual de comportamiento ya integrado en `main`, no implementación ni corrección. Si encuentras un defecto, repórtalo en la tabla final con su evidencia; no lo arregles ni abras rama/PR por tu cuenta — esa decisión es del propietario del proyecto.

Todas las nueve historias listadas en `related_hu` ya están integradas en `main` (ver tabla de dependencias arriba, cada una con su PR real). Lo que falta y este prompt produce es exactamente lo que sus issues de seguimiento (`#90`, `#95`, `#98`, `#100`, `#107`, `#111`, `#116`, `#120`, `#123`) describen como pendiente: recorrido E2E contra un navegador real (no solo Vitest/jsdom) y evidencia visual responsiva/accesible en los cuatro anchos del estándar. La suite Playwright de este repositorio (`apps/web/e2e/*.spec.ts`) ya existe para siete de las nueve historias y **es el guión autoritativo**: contiene las acciones exactas y las aserciones literales ya validadas contra los mismos textos que usan las pruebas de componente. Tu trabajo es reproducir esas mismas acciones a mano en un navegador real, confirmar cada aserción y capturar evidencia — no inventar un recorrido paralelo.

## Objetivo

Producir un informe verificable (paso, resultado esperado, resultado observado, evidencia) que confirme, contra un navegador real:

1. Los siete recorridos ya escritos en `apps/web/e2e/*.spec.ts` para `HU-040`, `HU-041`, `HU-061`–`HU-065` se comportan igual que describen sus specs.
2. El recorrido de `HU-042` (bloqueos puntuales y series semanales — la única parte de esa historia con pantalla completa; ver Fuera de alcance) funciona según los pasos manuales de este mismo prompt (no existe spec Playwright para esta historia todavía).
3. Aislamiento entre tenants en los endpoints/pantallas nuevos de B2 y B3.
4. Evidencia responsiva (320, 360, 768, 1280 px) y de accesibilidad (teclado, foco, zoom 200%) de cada pantalla nueva de B2/B3: `/panel/horarios`, `/panel/bloqueos`, `/panel/turnos/nuevo`, `/panel` (agenda diaria, ya cubierta parcialmente en la ejecución previa de B0/B1 pero nunca con contenido de citas real), `/panel/turnos/{id}`.

## Preflight obligatorio

1. No se crea rama ni issue: esta ejecución no modifica el repositorio.
2. Lee completamente `AGENTS.md`, `CLAUDE.md`, las secciones `HU-040`–`HU-042` y `HU-060`–`HU-065` de `docs/02-requisitos/historias-usuario.md`, y `docs/03-desarrollo/estandar-diseno-visual.md` (anchos de referencia y accesibilidad).
3. Lee completamente cada archivo `apps/web/e2e/*.spec.ts` listado en `source_docs`: son el guión exacto que reproducirás a mano. Presta atención particular a los textos literales esperados (encabezados, mensajes de error, etiquetas) — son las mismas cadenas que debes ver en el navegador real.
4. Ejecuta `graphify query "pantallas de horarios, bloqueos y agenda del panel privado"` si necesitas ubicar componentes adicionales; no es obligatorio para correr los recorridos ya descritos aquí.
5. Levanta el entorno real siguiendo `apps/api/README.md` y `apps/web/README.md`:
   - PostgreSQL real con las diecisiete migraciones aplicadas vía Atlas (no un contenedor vacío; sigue exactamente la receta del job `go` de `.github/workflows/ci.yml` si necesitas construirlo desde cero).
   - `go run ./cmd/api` corriendo contra esa base.
   - `pnpm run dev` en `apps/web` (por defecto `http://localhost:5173`).
   - Dos usuarios reales con hash argon2id real, uno por barbería (tenant A y tenant B), para las pruebas de aislamiento.
6. Usa las mismas variables de entorno que la suite E2E real en vez de inventar credenciales: `E2E_EMAIL`/`E2E_PASSWORD` (tenant A) y `E2E_EMAIL_B`/`E2E_PASSWORD_B` (tenant B). Si no están definidas, pide al operador que las provea antes de continuar; no adivines contraseñas.
7. Cada recorrido de B3 (`HU-061`–`HU-065`) necesita un barbero, un servicio y la asignación servicio→barbero ya existentes antes de poder crear un turno; cada recorrido de B2 (`HU-040`–`HU-042`) necesita al menos un barbero existente. Créalos tú mismo dentro de la sesión con nombres marcados con timestamp (ver Alcance excluido) siguiendo el mismo patrón que los `beforeEach`/helpers de los propios `.spec.ts` (`addBarber`, `addService`, `assignServiceToBarber`) — no reutilices datos de una ejecución anterior de este prompt ni asumas que ya existen.

## Alcance incluido

- **HU-040** (`/panel/horarios`, sección de tramos semanales): recorridos de `apps/web/e2e/horarios.spec.ts` — agregar tramo y verificar persistencia bajo su día, jornada partida con dos tramos no solapados, tramo solapado rechazado sin cerrar el diálogo, editar/retirar un tramo, aislamiento entre tenants (`404` real al consultar el horario de otra barbería).
- **HU-041** (`/panel/horarios`, secciones "Calendario de festivos colombianos" y "Excepciones de jornada"): recorridos de `apps/web/e2e/excepciones-festivos.spec.ts` — activar/desactivar el calendario de festivos con persistencia, excepción cerrada, excepción abierta con tramo especial, fecha duplicada rechazada, editar/retirar una excepción, aislamiento entre tenants.
- **HU-042** (`/panel/bloqueos`, alcance ya implementado en pantalla): agregar bloqueo puntual, agregar serie semanal, retirar cada uno, persistencia tras recargar, los siete tipos de bloqueo del selector, aislamiento entre tenants — ver pasos manuales en la sección propia más abajo.
- **HU-061** (`/panel/turnos/nuevo`): recorridos de `apps/web/e2e/nuevo-turno.spec.ts` — turno manual dentro de la próxima hora fuera de la rejilla pública, turno que se cruza con uno existente rechazado, un servicio no asignado al barbero no aparece en el selector.
- **HU-062** (`/panel`, agenda de hoy): recorridos de `apps/web/e2e/agenda-diaria.spec.ts` — fecha completa y zona visibles con selector de barbero obligatorio, un turno recién creado aparece con estado "Confirmado", cambiar de barbero muestra la agenda correcta, "Nuevo turno" navega al formulario real, la agenda usa la zona de la barbería.
- **HU-063** (`/panel`, navegación por fecha): recorrido de `apps/web/e2e/agenda-navegacion-fecha.spec.ts` — hoy → anterior → siguiente → fecha elegida → recarga → atrás/adelante con fecha y barbero siempre correctos; anterior/siguiente calculan el día civil de la zona de la barbería, nunca la del dispositivo.
- **HU-064** (`/panel/turnos/{id}`): recorrido de `apps/web/e2e/agenda-detalle-historial.spec.ts` — abrir un turno desde una fecha no actual, verificar snapshots de servicio/contacto/nota y el historial, volver a la misma fecha/barbero; un `appointmentId` inexistente muestra el estado "no disponible" explícito.
- **HU-065** (`/panel/turnos/{id}`, acción "Reprogramar turno"): recorridos de `apps/web/e2e/agenda-reprogramacion-turno.spec.ts` — reprogramar a otro día con la agenda anterior vacía y la nueva mostrando el turno, un único evento "Turno reprogramado" en el historial; reintentar el mismo envío tras una caída de red simulada no duplica el evento.
- Evidencia responsiva (320, 360, 768, 1280 px, sin scroll horizontal ni solapamientos) y de accesibilidad (navegación solo por teclado, foco visible, zoom 200%) de cada pantalla nueva: `/panel/horarios`, `/panel/bloqueos`, `/panel/turnos/nuevo`, `/panel/turnos/{id}`, y `/panel` con contenido de citas real (no vacío, a diferencia de la ejecución previa de B0/B1).
- Aislamiento entre tenants vía `fetch` en consola DevTools para los endpoints nuevos: horario, excepciones, bloqueos, agenda diaria, detalle/historial de turno, reprogramación.

## Fuera de alcance

- **HU-042, controles todavía no conectados a la pantalla** (issue real [#100](https://github.com/bcaceres19/barberia/issues/100)): fechas explícitas de una serie `date_list` (agregar/retirar), excepciones de una serie ("esta instancia no"), y edición de serie por alcance (`scope=whole`/`scope=this_and_following`). El backend ya los soporta, pero `BlocksPage.vue` no tiene controles para ellos todavía — **no los busques en la UI y no reportes su ausencia como defecto**: es un hallazgo ya conocido y trazado en el issue #100.
- `T3` (edición general de un turno), cancelación, completar, marcar `no_show`, corrección de estado y cierre automático: no existen todavía (B3 sigue en construcción más allá de `HU-065`). No navegues buscando esas acciones ni reportes su ausencia.
- B4 (reserva pública), B5 (notificaciones) y B6 (operación/piloto): no están implementados. No navegues rutas públicas de reserva ni reportes ausencia de notificaciones — B5 se documenta explícitamente como pendiente en cada HU.
- Cualquier corrección de código, migración o documento. Los hallazgos se reportan, no se arreglan aquí.
- Crear datos de negocio permanentes: usa nombres con marca de tiempo (mismo patrón que `apps/web/e2e/*.spec.ts`, p. ej. `QA B2B3 Barbero ${Date.now()}`) para no chocar con datos reales ni dejar ambigüedad sobre qué fue creado por esta ejecución.
- Repetir los recorridos de B0/B1 ya cubiertos por `docs/10-backlog/prompts/test/2026-08-25-qa-manual-b0-b1-chrome-mcp.md` — ese informe ya existe; no lo dupliques.

## Estado existente que debe conservarse

- Rutas ya definidas en `apps/web/src/app/router/index.ts`: `/panel/horarios` (`SchedulesPage.vue`, HU-040 y HU-041 en una sola pantalla), `/panel/bloqueos` (`BlocksPage.vue`, HU-042), `/panel/turnos/nuevo` (`NewAppointmentPage.vue`, HU-061), `/panel` (`DailyAgendaPage.vue`, HU-062/HU-063), `/panel/turnos/{appointmentId}` (`AppointmentDetailPage.vue`, HU-064/HU-065).
- Selectores estables ya usados por la suite Playwright real: estos prompts los citan porque un MCP de Chrome puede ubicarlos por rol/etiqueta accesible sin depender de clases CSS. Usa exactamente los mismos: `getByRole('heading', { name: '<texto exacto>' })`, `getByLabel('<etiqueta exacta>')`, `getByRole('button', { name: '<texto exacto>' })`.
- El vocabulario cerrado de siete tipos de bloqueo (`BLOCK_TYPES`, `apps/web/src/modules/schedules/model/timeBlock.ts`): Descanso, Almuerzo, No disponible, Día libre, Festivo, Vacaciones, Emergencia — el selector de tipo siempre ofrece exactamente estas siete opciones, nunca texto libre.

## Trabajo requerido

Ejecuta cada bloque en orden. Para cada paso, registra: resultado esperado (el que describe el `.spec.ts` citado, o el paso manual de este documento), resultado observado, y evidencia (captura de pantalla o texto literal visto).

### Bloque 1 — HU-040 y HU-041 (`/panel/horarios`)

1. Inicia sesión como tenant A. Entra a "Horarios" (`/panel/horarios`). Verifica el encabezado "Horarios".
2. Ejecuta, en orden, cada escenario de `apps/web/e2e/horarios.spec.ts` (los cinco `test(...)` del archivo): agregar tramo, jornada partida, tramo solapado rechazado, editar/retirar, aislamiento entre tenants (crea o reutiliza un barbero real del tenant B para el `404`).
3. En la misma pantalla, ejecuta cada escenario de `apps/web/e2e/excepciones-festivos.spec.ts` (los seis `test(...)`): activar calendario de festivos, excepción cerrada, excepción abierta con tramo especial, fecha duplicada rechazada, editar/retirar excepción, aislamiento entre tenants.
4. Verifica que la sección "Próximos festivos colombianos" solo aparece cuando `colombianHolidays.length > 0` (si la fecha actual del entorno no tiene festivos próximos en la ventana que calcula el backend, esta sección puede estar legítimamente ausente — no lo reportes como defecto sin antes confirmar en el backend si hay festivos colombianos configurados en el rango esperado).

### Bloque 2 — HU-042 (`/panel/bloqueos`)

No existe spec Playwright para esta historia: sigue estos pasos manuales.

5. Entra a "Horarios y bloqueos" (`/panel/bloqueos`, enlace en la navegación privada — confirma que existe una entrada de navegación hacia esta ruta; si no la hay, repórtalo como hallazgo). Verifica el encabezado "Horarios y bloqueos".
6. Selecciona un barbero en el selector "Barbero". Si no tiene bloqueos, debe verse "Este barbero no tiene bloqueos puntuales vigentes." y "Este barbero no tiene series de bloqueo configuradas.".
7. Haz clic en "Agregar bloqueo". Debe abrirse el diálogo "Agregar bloqueo puntual" con los campos "Tipo de bloqueo" (selector con las siete opciones exactas listadas arriba), "Fecha de inicio"/"Hora de inicio", "Fecha de fin"/"Hora de fin", y "Motivo (opcional)". Completa un bloqueo de tipo "Descanso" en el futuro cercano y guarda con "Guardar bloqueo". Debe aparecer en "Próximos bloqueos puntuales" con el tipo, el rango horario y el botón "Retirar".
8. Recarga la página, vuelve a seleccionar el mismo barbero: el bloqueo debe seguir ahí (persistencia real).
9. Haz clic en "Retirar" sobre ese bloqueo. Debe desaparecer de la lista sin recargar.
10. Haz clic en "Agregar serie semanal". Debe abrirse el diálogo "Agregar serie semanal" con "Tipo de bloqueo", "Día de la semana", "Hora de inicio", "Duración (minutos)", "Vigente desde" y "Motivo (opcional)". Completa una serie de tipo "Almuerzo" y guarda con "Guardar serie". Debe aparecer en "Series semanales" con el día, la hora y la duración.
11. Recarga la página: la serie debe seguir ahí. Haz clic en "Retirar" sobre la serie: debe desaparecer sin recargar.
12. Intenta crear un bloqueo puntual con un rango de fechas/horas inválido (por ejemplo, fin antes que inicio) y confirma que se muestra "Revisa los datos del bloqueo." sin cerrar el diálogo.
13. Aislamiento: con la sesión del tenant A, ejecuta en consola DevTools un `fetch` a los bloqueos de un `barberId` real del tenant B (por ejemplo, `fetch('/api/v1/private/barbers/{idDeB}/time-blocks?limit=50', {credentials:'include'})`). Debe responder `404`.
14. Confirma explícitamente, y regístralo así en el informe (no como defecto, como estado ya conocido): los controles de `date_list`, excepciones de serie y edición por `scope` no existen en esta pantalla (ver Fuera de alcance, issue #100).

### Bloque 3 — HU-061, HU-062 y HU-063 (`/panel/turnos/nuevo`, `/panel`)

15. Crea un barbero y un servicio nuevos (nombres con timestamp) y asígnalo al barbero, igual que hacen los helpers `addBarber`/`addService`/`assignServiceToBarber` de los `.spec.ts` citados.
16. Ejecuta cada escenario de `apps/web/e2e/nuevo-turno.spec.ts` (los tres `test(...)`): turno manual dentro de la próxima hora, turno cruzado rechazado, servicio no asignado ausente del selector.
17. Ejecuta cada escenario de `apps/web/e2e/agenda-diaria.spec.ts` (los cinco `test(...)`): fecha/zona con selector obligatorio, turno recién creado visible con estado "Confirmado", cambiar de barbero muestra la agenda correcta, "Nuevo turno" navega al formulario, zona de la barbería (no del dispositivo).
18. Ejecuta el recorrido completo de `apps/web/e2e/agenda-navegacion-fecha.spec.ts` (los dos `test(...)`): hoy → anterior → siguiente → fecha elegida → recarga → atrás/adelante, y el cálculo de zona de la barbería para anterior/siguiente.

### Bloque 4 — HU-064 y HU-065 (`/panel/turnos/{id}`)

19. Ejecuta el recorrido completo de `apps/web/e2e/agenda-detalle-historial.spec.ts` (los dos `test(...)`): detalle abierto desde una fecha no actual con snapshots/historial, y un `appointmentId` inexistente.
20. Ejecuta los dos escenarios de `apps/web/e2e/agenda-reprogramacion-turno.spec.ts`: reprogramar a otro día con agenda anterior/nueva e historial correctos, y el reintento tras caída de red que no duplica el evento.
21. Adicional (no cubierto explícitamente por el spec): con un turno `confirmed`, abre el diálogo "Reprogramar turno" y cancela sin enviar (botón "Cancelar"). El diálogo debe cerrarse sin ningún cambio ni petición de red visible en la pestaña Network.

### Bloque 5 — Responsivo y accesibilidad de B2/B3 (transversal)

22. Para `/panel/horarios`, `/panel/bloqueos`, `/panel/turnos/nuevo`, `/panel` (con al menos un turno real creado en los bloques anteriores, para que la fila con badge de estado sea visible) y `/panel/turnos/{id}` (abriendo el turno creado en el bloque 3 o 4): redimensiona el viewport a 320, 360, 768 y 1280 px. En cada ancho: sin scroll horizontal, sin elementos superpuestos, todo control visible y usable, incluidos los diálogos abiertos ("Agregar tramo", "Agregar excepción", "Agregar bloqueo puntual", "Agregar serie semanal", "Reprogramar turno").
23. Navega cada una de esas cinco pantallas solo con teclado (Tab/Shift+Tab/Enter/Espacio): el foco debe ser visible en todo momento y cada flujo (crear un tramo, crear un bloqueo, reprogramar un turno) debe poder completarse sin usar el mouse.
24. Verifica zoom del navegador al 200% en `/panel/horarios` y `/panel/turnos/{id}`: sin contenido cortado ni controles inalcanzables.

## Pruebas y evidencia

- Una captura de pantalla por cada paso donde el resultado sea visualmente verificable (formularios, mensajes de error, listas, diálogos).
- El texto literal exacto de cada mensaje de error/éxito mostrado, para comparar contra el texto citado en el `.spec.ts` correspondiente o en este documento.
- Para los pasos de aislamiento entre tenants, el código de estado HTTP observado.
- Para los pasos de responsivo, una captura por ancho evaluado y por pantalla.
- Para los pasos de accesibilidad, indica explícitamente si el foco fue visible en cada control y si el flujo se completó sin mouse.

## Documentación y trazabilidad

- Este prompt no actualiza documentación normativa. Si algún paso falla, no edites `historias-usuario.md`, `matriz-trazabilidad.md` ni ningún criterio de aceptación: repórtalo en la tabla final para que el propietario decida si es un defecto real o una duda a registrar en `docs/00-control/dudas-pendientes.md`.
- Guarda tu informe final como `docs/10-backlog/prompts/test/2026-09-01-qa-manual-b2-b3-chrome-mcp-report.md` (mismo patrón que el informe de B0/B1) y actualiza el metadato `execution_report` de este mismo archivo con esa ruta, y `status` a `executed`, una vez termines.

## Verificación final

No hay comandos de build/lint/test que correr: este prompt es exploración manual sobre la aplicación ya corriendo. Entrega una tabla:

`Paso | Bloque/HU | Resultado esperado | Resultado observado | Evidencia`

Y una segunda tabla solo con las filas cuyo resultado observado difiera del esperado, priorizadas por impacto (P0 rompe un criterio de aceptación o el aislamiento entre tenants; P1 incumple un criterio literal como responsive/accesibilidad; P2 es un hallazgo documental sin impacto en el comportamiento).

No declares cumplido un paso que no ejecutaste realmente en el navegador; si un paso no fue verificable, dilo explícitamente en vez de omitirlo o darlo por bueno.

## Git y PR

No aplica: esta ejecución no crea rama, commit ni PR. Si durante la ejecución identificas que hace falta un cambio de código, documenta el hallazgo con el paso exacto que lo reproduce y detente ahí — la decisión de abrir issue y prompt de corrección es del propietario del proyecto.
