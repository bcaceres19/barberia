---
prompt_id: "PROMPT-CHORE-195-GENERACION-MOCKUP-SERVICIOS-BARBERO-PENPOT-v1"
version: "1.0"
kind: "chore"
status: "blocked"
target_agents: ["codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-023"
related_hu: ["HU-009", "HU-021", "HU-022"]
issue: 195
issue_url: "https://github.com/bcaceres19/barberia/issues/195"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on: ["Issue #194 integrado en main con CI verde", "Issue maestro #184", "Penpot MCP conectado"]
rules: ["RN-SER-03", "RN-SER-04", "RN-TEN-01", "RN-DAT-02", "RN-IDE-01"]
decisions: ["DEC-019", "DEC-068", "DEC-077", "DEC-078", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-023-01", "CA-023-02", "CA-023-03", "CA-023-04", "CA-023-05", "CA-023-06", "CA-023-07", "CA-023-08"]
source_docs:
  - "AGENTS.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "apps/web/src/modules/barberServices"
  - "apps/web/src/modules/staff"
  - "apps/web/src/modules/catalog"
  - "apps/web/src/styles/tokens.css"
created_at: "2026-09-06"
updated_at: "2026-09-07"
artifacts:
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/servicios-por-barbero-eventos/README.md"
supersedes: null
superseded_by: null
---

# Genera el mockup editable de Servicios por barbero con Penpot

## Avance raster del 2026-09-07

Por instrucción directa del propietario se generó el atlas raster suplementario
`docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/servicios-por-barbero-eventos/`.
Incluye 17 estados pareados en escritorio/móvil y principales adicionales a
320, 768 y 1280 px, reutilizando literalmente el lenguaje de Barberos y
Servicios. Este avance no sustituye la fuente editable Penpot, la rama o el PR
exigidos por este prompt y no cambia su estado: #194 continúa abierto.

## Instrucción para Codex

Genera en Penpot el paquete de diseño de `/panel/servicios-por-barbero` para #195. El resultado es diseño previo editable y exportado; no implementes ni cambies reglas.

## Perfil y preflight

```text
Modelo: gpt-5.6-sol
Esfuerzo: medium
MCP: Penpot MCP local
```

Espera #194 integrado. Lee `AGENTS.md` y todos los `source_docs`, actualiza `main` y crea `chore/195-mockup-servicios-barbero`. Inspecciona la interfaz, cliente y pruebas reales. Comienza Penpot con lectura y trabaja en `Rutas pendientes`, sección `195/Servicios-por-barbero`, reutilizando componentes de #192 y #194.

## Estados y contenido autorizado

- principal con selector de barbero y lista real de servicios asignados/no asignados;
- carga de contexto, vacío sin barberos, vacío sin servicios y error recuperable;
- selección de barbero y carga de asignaciones;
- asignando, desasignando, éxito y error;
- rechazo visible al intentar retirar la última asignación de un servicio activo.

Un mismo servicio puede pertenecer a varios barberos. La asociación no contiene precio o duración propios ni orden manual.

## Trabajo requerido

1. Crea o reutiliza componentes Penpot para selector de barbero, monograma, fila de servicio, control de asignación, alerta y estado. Nombra `195/<componente>/<variante>`.
2. Diseña el principal en 320×800, 360×800, 768×1024 y 1280×1024; estados secundarios materiales en 360×800 y 1280×1024.
3. Haz comprensible qué barbero se edita, qué servicios cambiarán y cuándo una acción está guardando. No uses color como única señal ni controles ambiguos sin nombre accesible.
4. El rechazo de la última asignación debe conservar el estado anterior y explicar la solución autorizada; no ofrezcas desactivar servicio dentro de esta ruta si el código no lo hace.
5. Revisa listas largas, nombres extensos, foco/teclado, targets táctiles, contraste, clipping y overflow. Abre cada exportación.
6. Exporta y compacta en `docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/servicios-por-barbero-eventos/{desktop,mobile}/servicios-por-barbero/` y crea README con matriz, decisiones, exclusiones y procedencia Penpot.

## Fuera de alcance y entrega

No cambies app, API, backend, migraciones o reglas. No agregues precio/duración por profesional, comisión, orden, roles, borrado de barbero/servicio ni horarios.

Ejecuta `git diff --check`, revisa que solo haya diseño/documentación y actualiza prompt e índice. Entrega `Estado | Viewport | Frame | PNG | Revisión`. Commit/PR: `chore(design): genera mockups de servicios por barbero`; usa `Refs #195`.
