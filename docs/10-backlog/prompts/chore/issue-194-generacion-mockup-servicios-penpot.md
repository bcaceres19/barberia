---
prompt_id: "PROMPT-CHORE-194-GENERACION-MOCKUP-SERVICIOS-PENPOT-v1"
version: "1.0"
kind: "chore"
status: "blocked"
target_agents: ["codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-022"
related_hu: ["HU-009", "HU-024"]
issue: 194
issue_url: "https://github.com/bcaceres19/barberia/issues/194"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on: ["Issue #193 integrado en main con CI verde", "Issue maestro #184", "Penpot MCP conectado"]
rules: ["RN-SER-01", "RN-SER-02", "RN-SER-03", "RN-SER-04", "RN-TEN-01", "RN-DAT-02"]
decisions: ["DEC-003", "DEC-004", "DEC-067", "DEC-069", "DEC-077", "DEC-078", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-022-01", "CA-022-02", "CA-022-03", "CA-022-04", "CA-022-05", "CA-022-08", "CA-024-01", "CA-024-04", "CA-024-08"]
source_docs:
  - "AGENTS.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "apps/web/src/modules/catalog"
  - "apps/web/src/styles/tokens.css"
created_at: "2026-09-06"
updated_at: "2026-09-07"
artifacts:
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/servicios-eventos/README.md"
supersedes: null
superseded_by: null
---

# Genera el mockup editable de Servicios y ciclo de vida con Penpot

## Avance raster del 2026-09-07

Por instrucción directa del propietario se generó el atlas raster suplementario
`docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/servicios-eventos/`.
Incluye 38 estados pareados en escritorio/móvil y principales adicionales a
320, 768 y 1280 px, alineados con Agenda, Barberos y Barbería. Este avance no
sustituye la fuente editable Penpot, la rama o el PR exigidos por este prompt y
no cambia su estado: #193 continúa abierto.

## Instrucción para Codex

Genera en Penpot el paquete de diseño de `/panel/servicios` para #194. Debe cubrir catálogo, alta/edición y ciclo de vida existentes sin tocar la aplicación.

## Perfil y preflight

```text
Modelo: gpt-5.6-sol
Esfuerzo: medium
MCP: Penpot MCP local
```

Espera #193 integrado. Lee completamente `AGENTS.md` y los `source_docs`, actualiza `main` y crea `chore/194-mockup-servicios`. Inspecciona código, formularios, tipos y pruebas reales. En Penpot inicia con inspección de solo lectura y usa `Rutas pendientes`, sección `194/Servicios`.

## Estados y contenido autorizado

- principal con catálogo paginado, servicios activos/inactivos, nombre, duración y precio COP;
- carga, vacío y error recuperable;
- alta y edición: inicial, validación, guardando, conflicto y error;
- previsualización de impacto, confirmación de desactivación, desactivando y resultado;
- reactivación, incluidos conflicto de nombre y resultado.

El precio es COP fijo y mayor que cero; duración admite minutos enteros positivos sin presets cerrados. Los cambios del catálogo no reescriben citas históricas.

## Trabajo requerido

1. Crea componentes Penpot reutilizables para fila/registro, estado, badge activo, cifras tabulares, formulario, confirmación e impacto. Nombra `194/<componente>/<variante>`.
2. Diseña el principal en 320×800, 360×800, 768×1024 y 1280×1024; estados secundarios materiales en 360×800 y 1280×1024.
3. Prioriza nombre y legibilidad; no uses iconografía cliché como sustituto del nombre. Distingue crear, editar, desactivar y reactivar con copy y estructura, no solo color.
4. Representa errores de precio, duración, nombre duplicado, doble envío e impacto futuro sin prometer cancelación selectiva inexistente.
5. Verifica contraste, moneda y alineación numérica, nombres/descripciones largos, foco, objetivos táctiles, clipping y overflow. Abre cada PNG exportado.
6. Exporta y compacta en `docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/servicios-eventos/{desktop,mobile}/servicios/` y crea README con matriz estado/viewport/frame/PNG, decisiones, exclusiones y procedencia Penpot.

## Fuera de alcance y entrega

No cambies app, backend, OpenAPI, cliente, migraciones, reglas de impacto o asignaciones. No agregues moneda editable, servicios gratuitos, descuentos, impuestos, paquetes, inventario, cobros o propagación a citas.

Ejecuta `git diff --check`, confirma que el diff solo contiene diseño/documentación y actualiza prompt e índice. Entrega `Estado | Viewport | Frame | PNG | Revisión`. Commit/PR: `chore(design): genera mockups de servicios`; usa `Refs #194`.
