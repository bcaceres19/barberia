---
prompt_id: "PROMPT-CHORE-193-GENERACION-MOCKUP-BARBERIA-PENPOT-v1"
version: "1.0"
kind: "chore"
status: "blocked"
target_agents: ["codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-020"
related_hu: ["HU-009", "HU-012"]
issue: 193
issue_url: "https://github.com/bcaceres19/barberia/issues/193"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on: ["Issue #192 integrado en main con CI verde", "Issue maestro #184", "Penpot MCP conectado"]
rules: ["RN-DIS-07", "RN-TEN-01", "RN-DAT-02"]
decisions: ["DEC-007", "DEC-024", "DEC-033", "DEC-037", "DEC-077", "DEC-078", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-020-01", "CA-020-02", "CA-020-03", "CA-020-04", "CA-020-06", "CA-020-07", "CA-020-08"]
source_docs:
  - "AGENTS.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "apps/web/src/modules/settings"
  - "apps/web/src/styles/tokens.css"
created_at: "2026-09-06"
updated_at: "2026-09-07"
artifacts:
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/barberia-eventos/README.md"
supersedes: null
superseded_by: null
---

# Genera el mockup editable de Configuración de barbería con Penpot

## Avance raster del 2026-09-07

Por instrucción directa del propietario se generó un atlas raster suplementario
alineado mediante rasterización determinista en
`docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/barberia-eventos/`.
Incluye doce estados pareados en escritorio/móvil y los cuatro viewports del
estado principal. Este avance no sustituye la fuente editable Penpot, la rama o
el PR exigidos por este prompt y no cambia su estado: #192 continúa abierto.

## Instrucción para Codex

Genera en Penpot el paquete de diseño de `/panel/barberia` para #193, limitado a la configuración básica real. Entrega fuente editable, exportaciones y README; no implementes la pantalla.

## Perfil y preflight

```text
Modelo: gpt-5.6-sol
Esfuerzo: medium
MCP: Penpot MCP local
```

Espera #192 integrado. Lee `AGENTS.md` y todos los `source_docs`, actualiza `main` y crea `chore/193-mockup-barberia`. Inspecciona la página, formulario, tipos y pruebas vigentes. En Penpot empieza con `high_level_overview` y trabaja en `Rutas pendientes`, sección `193/Barberia`, reutilizando el sistema común.

## Estados y contenido autorizado

- principal con nombre, zona horaria IANA, correo opcional y teléfono opcional;
- carga inicial y error recuperable;
- formulario sin cambios y con cambios pendientes;
- validación por nombre, zona, correo y teléfono;
- guardando, guardado exitoso y error que conserva datos.

La zona se presenta como IANA y la hora se explica desde la barbería, no desde el dispositivo. Los contactos se usan solo dentro del formulario y todos los ejemplos son ficticios.

## Trabajo requerido

1. Crea componentes Penpot para agrupación editorial, campo, selector de zona, alerta/estado y acción de guardado, con nombres `193/<componente>/<variante>`.
2. Diseña el principal en 320×800, 360×800, 768×1024 y 1280×1024; estados secundarios materiales en 360×800 y 1280×1024.
3. Mantén jerarquía y guardado independiente. Representa validación, disabled y foco con texto/forma además del color.
4. No incluyas logo de tenant, tema, colores configurables, políticas, horarios, recordatorios, proveedores o canales OTP/notificación.
5. Revisa contraste AA, labels persistentes, teclado/foco, objetivos táctiles, privacidad visual, textos largos, clipping y overflow. Abre cada exportación.
6. Exporta y compacta en `docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/barberia-eventos/{desktop,mobile}/barberia/` y documenta en README estados, viewports, frames, PNG, decisiones, exclusiones y procedencia Penpot.

## Fuera de alcance y entrega

No cambies app, API, backend, migraciones, cliente tipado ni datos. El mockup no prueba persistencia, normalización o aislamiento.

Ejecuta `git diff --check`, confirma que el diff solo contiene diseño/documentación y actualiza prompt e índice. Entrega `Estado | Viewport | Frame | PNG | Revisión`. Commit/PR: `chore(design): genera mockups de configuración de barbería`; usa `Refs #193`.
