---
prompt_id: "PROMPT-TEST-198-REVISION-INTEGRAL-ATLAS-NAVA-v1"
version: "1.0"
kind: "test"
status: "blocked"
target_agents: ["claude", "codex", "human"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu: ["HU-010", "HU-011", "HU-012", "HU-020", "HU-021", "HU-022", "HU-023", "HU-024", "HU-040", "HU-041", "HU-042", "HU-061", "HU-062", "HU-063", "HU-064", "HU-065"]
issue: 198
issue_url: "https://github.com/bcaceres19/barberia/issues/198"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on:
  - "Issues #188 a #197 integrados en main con CI verde"
  - "Issues #212/#213 integrados o reconciliados sin regresiones"
  - "Issue maestro #184"
rules: ["RN-TEN-01", "RN-DAT-02", "RN-IDE-01"]
decisions: ["DEC-077", "DEC-078", "DEC-079", "DEC-080", "DEC-081"]
acceptance_criteria:
  - "Las once rutas P0 implementadas se recorren en sus estados aplicables."
  - "Cada ruta tiene comparación final contra su panel exacto y evidencia responsive/accesible."
  - "No hay regresiones funcionales, errores nuevos de consola/red, overflow ni pérdida de foco."
  - "Toda diferencia queda cerrada o enlazada a un issue propio; #184 no se declara completo con FAIL abiertos."
source_docs:
  - "AGENTS.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/README.md"
  - "docs/10-backlog/prompts/orchestration/redisenio-integral-nava-tailored-grid-v3.md"
  - "apps/web/src/app/router/index.ts"
  - "apps/web/e2e"
created_at: "2026-09-04"
updated_at: "2026-09-04"
supersedes: null
superseded_by: null
---

# Revisión integral final del atlas NAVA

Ejecuta el issue [#198](https://github.com/bcaceres19/barberia/issues/198) únicamente cuando todas las fases estén integradas. Esta entrega valida el conjunto; no reabre diseño ni mezcla correcciones sustanciales. Un defecto encontrado recibe issue y fix propios salvo una corrección documental/evidencia trivial dentro del alcance aprobado.

## Trabajo requerido

1. Lee todos los `source_docs`, confirma dependencias y crea `test/198-revision-integral-atlas-nava` desde `main` actualizada si se persistirá informe/evidencia.
2. Inventa cero rutas: audita el router real y construye matriz de las once rutas P0, panel exacto, estados, HU y evidencia esperada.
3. Recorre app real con datos sintéticos: acceso, recuperación, agenda, nuevo turno, detalle/reprogramación, servicios, asignaciones, barberos, barbería, horarios y bloqueos.
4. Para cada ruta revisa referencia/final/lado a lado/overlay o diff; 320/360/768/1280; zoom 200 %; teclado, foco, headings, axe, contraste, movimiento reducido y scroll accidental.
5. Revisa consola y Network, expiración/sesión, carga, vacío, error, conflicto, éxito y disabled aplicables. No persistas credenciales ni PII.
6. Ejecuta formato, lint, tipos, unitarias/componentes, suite E2E completa y build. Separa fallos preexistentes de regresiones y no uses reruns para ocultar #158.

## Entrega

Persiste un informe con `Ruta | Panel | Estados | Responsive | Accesibilidad | Consola/red | Pruebas | Fidelidad PASS/FAIL` y una tabla de desviaciones con issue real. Actualiza #184/#198 y el catálogo. Commit/PR: `test(web): valida el rediseño integral NAVA`. Usa `Closes #198` y cierra #184 solo si ninguna ruta queda en FAIL ni falta evidencia.
