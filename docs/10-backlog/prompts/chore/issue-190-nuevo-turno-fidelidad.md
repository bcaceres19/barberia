---
prompt_id: "PROMPT-CHORE-190-NUEVO-TURNO-FIDELIDAD-v1"
version: "1.0"
kind: "chore"
status: "ready"
target_agents: ["claude", "codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-061"
related_hu: ["HU-009", "HU-012"]
issue: 190
issue_url: "https://github.com/bcaceres19/barberia/issues/190"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on:
  - "Issue #189 integrado en main con CI verde"
  - "Fundaciones y shell de #186/#187 integrados"
  - "Issue maestro #184"
rules: ["RN-CIT-01", "RN-DAT-01", "RN-IDE-01", "RN-TEN-01"]
decisions: ["DEC-071", "DEC-072", "DEC-073", "DEC-077", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-061-01", "CA-061-02", "CA-061-03", "CA-061-04", "CA-061-05", "CA-061-06", "CA-061-07", "CA-061-08"]
source_docs:
  - "AGENTS.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/README.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/nuevo-turno-eventos/README.md"
  - "docs/10-backlog/prompts/orchestration/redisenio-integral-nava-tailored-grid-v3.md"
  - "docs/10-backlog/prompts/hu/hu-061-creacion-manual-citas.md"
  - "apps/web/src/modules/agenda"
created_at: "2026-09-04"
updated_at: "2026-09-04"
supersedes: null
superseded_by: null
---

# Fidelidad de Nuevo turno al atlas NAVA

## Instrucción

Rediseña únicamente `/panel/turnos/nuevo` para reproducir, evento por evento, el atlas ejecutable de [`nuevo-turno-eventos`](../../evidence/ui-mockups-nava-tailored-grid-2026-09-03/nuevo-turno-eventos/README.md). La lámina compuesta `03-nuevo-turno.png` queda como historial visual, no como objetivo de implementación: reunía varias pantallas y arrastraba controles de cuenta decorativos en el encabezado ya corregidos en el atlas por evento. Es una entrega visual del issue [#190](https://github.com/bcaceres19/barberia/issues/190): conserva exactamente contrato, campos, obligatoriedad, validación, selección servicio-barbero, idempotencia y comportamiento de conflicto de `HU-061`.

`#189` y las fundaciones de `#186`/`#187` ya están integradas en `main` (2026-09-04): este prompt está `ready`. Crea `chore/190-nuevo-turno-fidelidad` desde `main` actualizada y registra la rama.

## Trabajo requerido

1. Lee completos los `source_docs`, inspecciona la ruta, componentes y pruebas reales, y preserva cambios ajenos.
2. Abre cada PNG de `nuevo-turno-eventos` a resolución original y mide canvas, tarjetas numeradas, encabezado, grupos de campos, resumen, CTA, errores y densidad para el evento correspondiente.
3. Captura baseline en la app real y cubre vacío, relleno, validación, envío, éxito y conflicto `409` sin pérdida de datos.
4. Reproduce composición, jerarquía, color, tipografía, espaciado, alineación, iconos y estados. Reutiliza `shared/ui`; no agregues controles sin comportamiento existente.
5. Entrega referencia/baseline/final/lado a lado/overlay o diff en el viewport del mockup y evidencia 320/360/768/1280, zoom 200 %, teclado, foco y movimiento reducido.

## Fuera de alcance

- Cambiar `apps/api`, OpenAPI, cliente generado, migraciones, campos, obligatoriedad o reglas de reconciliación.
- Añadir disponibilidad pública, recordatorios, pagos, cambio de barbero posterior o nuevas acciones de cita.
- Declarar fidelidad por tokens o métricas DOM sin inspeccionar la comparación visual.

## Pruebas y entrega

Conserva pruebas de componente para doble envío, errores y datos; ejecuta formato, lint, tipos, unitarias, E2E afectada y build. Revisa axe, consola y Network en la app real. Entrega `Criterio | Estado | Evidencia` y `Diferencia | Autoridad | Tratamiento`. Commit/PR: `chore(web): reproduce Nuevo turno según el atlas NAVA`. Usa `Closes #190` solo con toda la evidencia; no mezcles #191.
