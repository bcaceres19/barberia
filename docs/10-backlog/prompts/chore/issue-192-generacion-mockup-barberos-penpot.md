---
prompt_id: "PROMPT-CHORE-192-GENERACION-MOCKUP-BARBEROS-PENPOT-v1"
version: "1.0"
kind: "chore"
status: "blocked"
target_agents: ["codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-021"
related_hu: ["HU-009", "HU-012"]
issue: 192
issue_url: "https://github.com/bcaceres19/barberia/issues/192"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on: ["Issue #191 integrado en main con CI verde", "Issue maestro #184", "Penpot MCP conectado"]
rules: ["RN-TEN-01", "RN-DAT-02"]
decisions: ["DEC-019", "DEC-047", "DEC-077", "DEC-078", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-021-01", "CA-021-02", "CA-021-03", "CA-021-04", "CA-021-05", "CA-021-07", "CA-021-08"]
source_docs:
  - "AGENTS.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/panel-agenda-eventos/README.md"
  - "apps/web/src/modules/staff"
  - "apps/web/src/styles/tokens.css"
created_at: "2026-09-06"
updated_at: "2026-09-06"
supersedes: null
superseded_by: null
---

# Genera el mockup editable de Barberos con Penpot

## Instrucción para Codex

Genera en Penpot el paquete de diseño de `/panel/barberos` para #192. Produce únicamente diseño editable, PNG y documentación; no cambies Vue ni comportamiento.

## Perfil y preflight

```text
Modelo: gpt-5.6-sol
Esfuerzo: medium
MCP: Penpot MCP local
```

Espera la integración de #191. Lee por completo `AGENTS.md` y los `source_docs`, actualiza `main` y crea `chore/192-mockup-barberos`. Verifica el modelo efectivo. Con Penpot conectado, comienza con inspección de solo lectura y trabaja en `Rutas pendientes`, sección `192/Barberos`. Si algo ya existe, reutiliza componentes y no sobrescribas frames ajenos.

## Estados y contenido autorizado

- estado principal con cuatro barberos y las acciones reales de añadir y renombrar;
- carga, vacío y error recuperable;
- formulario o diálogo de alta: inicial, validación, guardando, éxito y error;
- edición/renombrado: inicial, validación, guardando y error.

Usa solo nombre visible. Se permite monograma derivado del nombre como recurso decorativo; fotografía, rol, correo, teléfono, credenciales, orden manual, desactivación y eliminación están prohibidos por alcance.

## Trabajo requerido

1. Inspecciona componentes, textos, estados y pruebas reales de `staff`; no conviertas propuestas del diseño en funciones.
2. Crea componentes Penpot reutilizables para encabezado, fila de barbero, monograma, estado de página, campo y diálogo. Nombra `192/<componente>/<variante>`.
3. Diseña el principal en 320×800, 360×800, 768×1024 y 1280×1024; estados secundarios materiales en 360×800 y 1280×1024.
4. Expresa NAVA/Tailored Grid y la paleta obligatoria, pero trabaja con libertad compositiva. Haz clara la diferencia entre ver equipo, añadir y renombrar sin apoyarte solo en color.
5. Verifica contraste, lectura con nombres largos, foco visible, orden de lectura, objetivos táctiles y ausencia de overflow. Abre cada exportación.
6. Exporta y compacta en `docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/barberos-eventos/{desktop,mobile}/barberos/` y actualiza el README con matriz estado/viewport/frame/PNG, decisiones, exclusiones y procedencia Penpot.

## Fuera de alcance y entrega

No cambies app, backend, OpenAPI, cliente, datos o permisos. No inventes fotos, roles, cuentas, servicios, horarios ni eliminación. Los mockups no prueban aislamiento ni persistencia.

Ejecuta `git diff --check`, revisa que el diff solo contenga diseño/documentación y actualiza prompt e índice con datos reales. Entrega `Estado | Viewport | Frame | PNG | Revisión`. Commit/PR: `chore(design): genera mockups de barberos`; usa `Refs #192`.
