---
prompt_id: "PROMPT-FEAT-GCAL-07-FRONTEND-CONEXION-v1"
version: "1.0"
kind: "hu"
status: "draft"
target_agents:
  - "any"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu: []
issue: "pending"
issue_url: null
suggested_issue_title: "feat(web): conectar, sincronizar y desconectar Google Calendar desde el área del barbero"
branch: null
pr: null
pr_url: null
depends_on:
  - "PROMPT-FEAT-GCAL-04-v1 integrado"
  - "PROMPT-FEAT-GCAL-06-v1 integrado"
rules: []
decisions:
  - "DEC-099"
  - "DEC-101"
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/06-api/estandar-openapi.md"
  - "apps/web/src/modules"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - ".agents/skills/visual-qa/SKILL.md"
created_at: "2026-09-26"
updated_at: "2026-09-26"
supersedes: null
superseded_by: null
---

# Interfaz de Google Calendar para el barbero

## Instrucción para el agente

Implementa únicamente la preocupación descrita en este archivo. Trabaja de forma autónoma hasta producir la entrega definida, sin ampliar el alcance ni inventar decisiones. Es una entrega de la orquestación [PROMPT-ORCH-GCAL-BARBERO-v1](../orchestration/google-calendar-barbero.md); léela para conocer el orden y el principio de autoridad (NAVA manda, Google es una vista).

## Objetivo

El barbero autenticado puede conectar, ver el estado, sincronizar ahora y desconectar Google Calendar con la identidad NAVA / Tailored Grid, y la agenda muestra los bloqueos y el origen en el historial.

## Preflight obligatorio

1. Comprueba que el árbol esté limpio y no sobrescribas cambios ajenos; trabaja en un worktree si hay cambios sin relación.
2. Actualiza `main` mediante fast-forward y verifica que cada dependencia de los metadatos esté integrada.
3. Lee completamente `AGENTS.md`, `CLAUDE.md` y cada archivo de `source_docs`, y las `DEC-*` enlazadas.
4. Comprueba `dudas-pendientes.md` y `contradicciones.md`. Si falta una decisión, registra el bloqueo antes de cambiar código.
5. Este prompt nace con `issue: pending` y en `draft`. Antes de ejecutarlo crea el issue real con criterios verificables (título sugerido en los metadatos), actualiza este archivo y el índice, y solo entonces pásalo a `ready`.
6. Si esta entrega introduce o toca una historia, redáctala en `docs/02-requisitos/historias-usuario.md` con su número real y criterios `CA-*` antes de codificar; no inventes números.
7. Crea la rama `<tipo>/<issue>-<descripcion>` desde `main` actualizada.

## Alcance incluido

- Módulo o sección del área del barbero con los estados No conectado, Conectando, Conectado, Sincronizando, Requiere reconexión y Error de sincronización, y las acciones Conectar, Sincronizar ahora y Desconectar (acción de doble clic segura).
- Cliente tipado del contrato; el componente no conoce la forma interna del API ni maneja tokens.
- Refresco de la agenda tras sincronizar; el bloqueo de Google se ve como un bloque de agenda sin etiquetas técnicas.
- Etiqueta «Origen: sincronización Google Calendar» en el historial de una cita cambiada por la integración.

## Fuera de alcance

- Cualquier cambio de backend o contrato.
- Calendarios de clientes o administradores.

## Estado existente que debe conservarse

- Sistema visual NAVA / Tailored Grid; sin mockup asignado, la creación visual es libre dentro de esa identidad (modo guiado por identidad de `visual-qa`).
- Avisos emergentes tipo acordeón (`DEC-095`) para éxito y error.

## Trabajo requerido

1. Ejecuta `visual-qa` antes de editar y antes de declarar terminado.
2. Implementa los estados con datos simulados del cliente tipado y accesibilidad de teclado y lector de pantalla.
3. Añade pruebas de componente y el recorrido E2E con el API simulado.
4. Captura evidencia responsive.

## Pruebas y evidencia

- Componente: cada estado, acciones deshabilitadas mientras sincroniza, error recuperable.
- E2E: conectar (redirección simulada), sincronizar, desconectar y refresco de agenda.
- Evidencia y verificación accesible a 320, 360, 768 y 1280 px.

## Documentación y trazabilidad

- Actualiza matriz de trazabilidad, historial, contrato OpenAPI, diccionario y diagrama de datos que realmente resulten afectados.
- Registra la evidencia visual en `docs/10-backlog/evidence/` según la convención vigente.
- Actualiza los metadatos y el índice de este catálogo con issue, rama, PR y estado reales.

## Verificación final

```text
cd apps/web && pnpm typecheck && pnpm lint && pnpm test:unit && pnpm build
Playwright E2E aplicable
tools/ai/validate-agent-system.sh --strict
```

Entrega una tabla `Criterio | Estado | Prueba o evidencia`. No declares cumplido aquello que no esté probado.

## Git y PR

- Commit/PR con Conventional Commits, sin secretos ni datos personales.
- `Closes #<issue>` solo si se cubre el issue completo; de lo contrario `Refs #<issue>`.
- No hagas push directo, force push ni merge de `main`; integra por PR y squash tras CI verde.
