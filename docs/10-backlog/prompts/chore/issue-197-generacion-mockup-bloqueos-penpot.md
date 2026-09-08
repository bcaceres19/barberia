---
prompt_id: "PROMPT-CHORE-197-GENERACION-MOCKUP-BLOQUEOS-PENPOT-v1"
version: "1.0"
kind: "chore"
status: "blocked"
target_agents: ["codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-042"
related_hu: ["HU-009", "HU-040", "HU-041"]
issue: 197
issue_url: "https://github.com/bcaceres19/barberia/issues/197"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on: ["Issue #196 integrado en main con CI verde", "Issue maestro #184", "Penpot MCP conectado"]
rules: ["RN-BLQ-01", "RN-BLQ-02", "RN-BLQ-03", "RN-BLQ-04", "RN-DIS-07", "RN-TEN-01"]
decisions: ["DEC-008", "DEC-009", "DEC-020", "DEC-070", "DEC-073", "DEC-076", "DEC-077", "DEC-078", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-042-01", "CA-042-02", "CA-042-03", "CA-042-04", "CA-042-05", "CA-042-06", "CA-042-07", "CA-042-08"]
source_docs:
  - "AGENTS.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "apps/web/src/modules/schedules"
  - "apps/web/src/styles/tokens.css"
created_at: "2026-09-06"
updated_at: "2026-09-08"
artifacts:
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/bloqueos-eventos/README.md"
supersedes: null
superseded_by: null
---

# Genera el mockup editable de Bloqueos con Penpot

## Avance raster del 2026-09-08

Por instrucción directa del propietario se generó el atlas raster suplementario
`docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/bloqueos-eventos/`.
Incluye 15 estados pareados en escritorio/móvil y principales adicionales a
320, 768 y 1280 px, sin presentar capacidades pendientes de #100. Este avance no
sustituye la fuente editable Penpot, la rama, el PR ni las dependencias exigidas
por este prompt; por eso conserva el estado `blocked`.

## Instrucción para Codex

Genera en Penpot el paquete de diseño de `/panel/bloqueos` para #197. Cubre bloqueos puntuales y series semanales que existan en la interfaz y contrato vigentes. No implementes ni amplíes capacidades.

## Perfil y preflight

```text
Modelo: gpt-5.6-sol
Esfuerzo: medium
MCP: Penpot MCP local
```

Espera #196 integrado. Lee `AGENTS.md` y todos los `source_docs`, verifica el estado real de #98/#100, actualiza `main` y crea `chore/197-mockup-bloqueos`. Inspecciona código, contrato visible y pruebas para separar implementado de pendiente. En Penpot empieza con lectura y usa `Rutas pendientes`, sección `197/Bloqueos`.

## Estados y contenido autorizado

- principal con barbero seleccionado y bloqueos vigentes;
- carga, sin barberos, vacío y error recuperable;
- creación puntual con los siete tipos autorizados;
- serie semanal y, solo si ya existen, lista explícita, excepción y retiro lógico;
- validación, solape/conflicto, guardando, éxito y error que conserva datos;
- retirada/confirmación solo en la forma realmente implementada.

Un bloqueo puede coexistir con una cita ya creada sin moverla automáticamente. No prometas conteo de citas afectadas si no existe en la interfaz/contrato.

## Trabajo requerido

1. Crea componentes Penpot para selector, tipo, intervalo, recurrencia, fila/ficha, estado, alerta y confirmación. Nombra `197/<componente>/<variante>`.
2. Diseña el principal en 320×800, 360×800, 768×1024 y 1280×1024; estados secundarios materiales en 360×800 y 1280×1024.
3. Diferencia claramente puntual, recurrente, excepción y retirado mediante estructura, etiquetas y copy. Los siete tipos no dependen únicamente del color.
4. Representa zona horaria, intervalos, fecha/serie y acciones vigentes sin crear edición si no existe. Conserva datos ante errores y doble envío.
5. Revisa listas largas, recurrencia, cruce de medianoche, foco, teclado, targets táctiles, contraste, clipping y overflow. Abre cada exportación.
6. Exporta y compacta en `docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/bloqueos-eventos/{desktop,mobile}/bloqueos/` y crea README con matriz, capacidades presentes/ausentes, decisiones, exclusiones y procedencia Penpot.

## Fuera de alcance y entrega

No cambies app, backend, OpenAPI, migraciones, reglas o impacto sobre citas. No agregues edición, calendario, disponibilidad, cancelación/reprogramación de turnos ni acciones que #100 siga dejando pendientes.

Ejecuta `git diff --check`, confirma que el diff solo contiene diseño/documentación y actualiza prompt e índice. Entrega `Estado | Viewport | Frame | PNG | Revisión`. Commit/PR: `chore(design): genera mockups de bloqueos`; usa `Refs #197`.
