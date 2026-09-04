---
prompt_id: "PROMPT-CHORE-189-AGENDA-DIARIA-FIDELIDAD-v2"
version: "2.0"
kind: "chore"
status: "superseded"
target_agents:
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-062"
related_hu:
  - "HU-009"
  - "HU-012"
  - "HU-063"
issue: "189"
issue_url: "https://github.com/bcaceres19/barberia/issues/189"
suggested_issue_title: null
branch: "chore/189-agenda-diaria-fidelidad"
pr: null
pr_url: null
depends_on:
  - "Atlas panel-agenda-eventos integrado mediante PR #216"
  - "Controles reglados shared/ui integrados mediante PR #214"
  - "Implementación inicial de fidelidad en commit 9cd9e55"
rules: []
decisions:
  - "DEC-074"
  - "DEC-075"
  - "DEC-077"
  - "DEC-079"
  - "DEC-080"
acceptance_criteria:
  - "CA-062-01"
  - "CA-062-02"
  - "CA-062-03"
  - "CA-063-01"
  - "CA-063-02"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/panel-agenda-eventos/README.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/00-control/registro-decisiones.md"
created_at: "2026-09-03"
updated_at: "2026-09-03"
supersedes: "PROMPT-CHORE-189-AGENDA-DIARIA-FIDELIDAD-v1"
superseded_by: "PROMPT-CHORE-189-AGENDA-DIARIA-FIDELIDAD-v3"
---

# Fidelidad visual de `/panel` con datos mock de diseño

## Instrucción para el agente

Continúa el issue #189 en modo de fidelidad al atlas `panel-agenda-eventos`.
La persona propietaria autorizó el 2026-09-03 usar datos simulados para
montar, capturar y comparar el diseño. Esta versión sustituye la exigencia de
depender del backend/QA local de v1 para la evidencia visual.

## Objetivo

Montar determinísticamente los doce estados del atlas de `/panel` con mocks
de red exclusivos de Playwright, en escritorio y móvil, para poder iterar la
composición y producir evidencia visual sin depender de login, PostgreSQL,
credenciales, OTP o rate limit locales.

## Alcance incluido

- Un harness de Playwright local a `apps/web/e2e` que intercepte sesión,
  barberos, zona horaria y agenda diaria con respuestas sintéticas tipadas.
- Evidencia de los 12 eventos del atlas en sus dos viewports, usando solo
  identidades ficticias y fechas fijas.
- Actualización de los E2E de agenda que dependían del encabezado y del
  selector nativo reemplazado por `BarberSelect`.
- Correcciones visuales que se descubran al comparar la app real renderizada
  con los PNG asignados.

## Fuera de alcance

- No introducir mocks en `src/`, en el cliente HTTP, rutas de producción,
  modelos de dominio, API, backend, OpenAPI o base de datos.
- No convertir los fixtures de diseño en disponibilidad, autorización,
  comportamiento de agenda o datos persistidos.
- No usar credenciales, datos personales, tokens ni una cuenta real para la
  evidencia de esta versión.
- La fotografía del barbero sigue fuera de alcance: el mock usa monogramas.

## Estado existente que debe conservarse

- `DailyAgendaPage`, `BarberSelect`, `AgendaSkeleton`, `BaseSpinner`,
  `PageState` y `BarberAvatar` del commit `9cd9e55` son el punto de partida.
- `DEC-074`: una única selección de barbero; no se crea una cuadrícula
  multi-barbero.
- `DEC-075`: el cambio de día es decorativo y los turnos nocturnos conservan
  su intervalo real.
- La línea temporal sigue `aria-hidden` y la lista es su equivalente
  accesible.
- En texto o comportamiento manda el código vigente; toda diferencia visual
  restante se registra en la tabla de desviaciones existente.

## Trabajo requerido

1. Instala interceptores Playwright antes de navegar a `/panel`: sesión
   autenticada sintética, contexto de barberos, zona y agenda. No hagas login
   ni llames al backend real.
2. Modela de forma explícita los eventos 01–12 del README del atlas: cargando
   contexto, error de contexto, sin barberos, carga/actualización/error de
   agenda, no encontrado, vacío, zona no disponible, nocturno y selector
   abierto, además de la agenda cargada.
3. Captura la app en el viewport efectivo correspondiente a cada PNG y
   verifica `innerWidth`/`innerHeight` desde la prueba. Mantén también los
   anchos 320, 360, 768 y 1280 y zoom aproximado al 200 % como evidencia
   responsive.
4. Compara cada captura con el PNG de referencia; corrige primero estructura,
   escala, jerarquía, superficie, alineación y estado. Conserva la tabla de
   desviaciones para diferencias autorizadas.
5. Mantén los E2E funcionales contra backend como pruebas separadas: este
   harness visual mockeado no pretende sustituirlos.

## Pruebas y evidencia

- Vitest de `DailyAgendaPage` y los tres componentes compartidos, incluida
  navegación deshabilitada sin zona, `not-found`, horario nocturno y avatar.
- Playwright mockeado de los 12 eventos en escritorio y móvil, sin red real,
  con `axe-core` donde la página tenga contenido nuevo y con evidencia bajo
  `apps/web/e2e/evidence/panel/fidelidad-189/`.
- Formato, lint, typecheck y build. Si los E2E de backend continúan fallando,
  se registran como infraestructura ajena y no se ocultan.

## Documentación y trazabilidad

- Actualiza `panel-fidelidad-189-desviaciones.md` con el cambio de estrategia
  y las diferencias no cerradas.
- Actualiza los metadatos de este prompt, el prompt v1 y el índice del
  catálogo al cambiar estado, abrir PR o terminar.

## Entrega

Entrega `Criterio | Estado | Prueba o evidencia`, una fila para cada evento y
viewport y una para cada `CA-*`. Usa `Refs #189` mientras no se hayan cerrado
todas las diferencias y la evidencia completa. No hagas push, merge ni crees
un PR sin autorización explícita.
