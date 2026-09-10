---
prompt_id: "PROMPT-CHORE-196-GENERACION-MOCKUP-HORARIOS-PENPOT-v1"
version: "1.0"
kind: "chore"
status: "blocked"
target_agents: ["codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-040"
related_hu: ["HU-009", "HU-041"]
issue: 196
issue_url: "https://github.com/bcaceres19/barberia/issues/196"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on: ["Issue #195 integrado en main con CI verde", "Issue maestro #184", "Penpot MCP conectado"]
rules: ["RN-DIS-01", "RN-DIS-02", "RN-DIS-03", "RN-DIS-05", "RN-DIS-07", "RN-TEN-01"]
decisions: ["DEC-007", "DEC-019", "DEC-020", "DEC-070", "DEC-077", "DEC-078", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-040-01", "CA-040-02", "CA-040-03", "CA-040-04", "CA-040-05", "CA-040-08", "CA-041-01", "CA-041-08"]
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
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/horarios-eventos/README.md"
supersedes: null
superseded_by: null
---

# Genera el mockup editable de Horarios y excepciones con Penpot

## Avance raster del 2026-09-08

Por instrucción directa del propietario se generó el atlas raster suplementario
`docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/horarios-eventos/`.
Incluye 16 estados pareados en escritorio/móvil y principales adicionales a
320, 768 y 1280 px, alineados con el lenguaje visual consolidado. Este avance no
sustituye la fuente editable Penpot, la rama, el PR ni las dependencias exigidas
por este prompt; por eso conserva el estado `blocked`.

## Instrucción para Codex

Genera en Penpot el paquete de diseño de `/panel/horarios` para #196. Representa jornada semanal, segmentos, días cerrados, festivos y excepciones reales. No implementes cambios en la app.

## Perfil y preflight

```text
Modelo: gpt-5.6-sol
Esfuerzo: medium
MCP: Penpot MCP local
```

Espera #195 integrado. Lee por completo `AGENTS.md` y `source_docs`, actualiza `main` y crea `chore/196-mockup-horarios`. Inspecciona vistas, formularios, tipos y pruebas vigentes. En Penpot comienza con lectura y usa `Rutas pendientes`, sección `196/Horarios`, reutilizando selector y patrones previos.

## Estados y contenido autorizado

- principal con barbero seleccionado, semana, días abiertos/cerrados y uno o varios segmentos;
- carga, sin barberos, horario vacío y error recuperable;
- alta/edición/retiro de segmento, validación y solape;
- excepción de fecha: abierta, cerrada y con segmentos;
- festivo aplicable y excepción explícita;
- guardando, éxito y error que conserva la edición.

Las horas se interpretan en la zona IANA de la barbería; intervalos son `[inicio, fin)`. Un tramo puede cruzar medianoche solo cuando las reglas vigentes lo permiten.

## Trabajo requerido

1. Crea componentes Penpot para selector, fila de día, segmento, editor, excepción, badge de festivo, alerta y estado. Nombra `196/<componente>/<variante>`.
2. Diseña el principal en 320×800, 360×800, 768×1024 y 1280×1024; estados secundarios materiales en 360×800 y 1280×1024.
3. Evita una tabla ilegible en móvil: aplica reflow preservando día, estado, segmentos y acciones. No miniaturices controles.
4. Distingue horario semanal, excepción de fecha y festivo mediante jerarquía y copy, no solo color. Representa solapes y errores junto al control afectado.
5. Revisa múltiples segmentos, cruce de medianoche, nombres largos, foco, teclado, targets táctiles, contraste, clipping y overflow. Abre cada exportación.
6. Exporta y compacta en `docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/horarios-eventos/{desktop,mobile}/horarios/` y crea README con matriz, reglas visualizadas, decisiones, exclusiones y procedencia Penpot.

## Fuera de alcance y entrega

No cambies app, backend, OpenAPI, migraciones, recurrencia, zona o reglas. No agregues plantillas, calendario mensual, disponibilidad calculada, citas o automatismos inexistentes.

Ejecuta `git diff --check`, verifica diff de solo diseño/documentación y actualiza prompt e índice. Entrega `Estado | Viewport | Frame | PNG | Revisión`. Commit/PR: `chore(design): genera mockups de horarios`; usa `Refs #196`.
