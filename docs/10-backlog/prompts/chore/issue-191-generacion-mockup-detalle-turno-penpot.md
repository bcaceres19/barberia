---
prompt_id: "PROMPT-CHORE-191-GENERACION-MOCKUP-DETALLE-TURNO-PENPOT-v1"
version: "1.0"
kind: "chore"
status: "blocked"
target_agents: ["codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-064"
related_hu: ["HU-009", "HU-012", "HU-063", "HU-065"]
issue: 191
issue_url: "https://github.com/bcaceres19/barberia/issues/191"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on: ["Issue #190 integrado en main con CI verde", "Issue maestro #184", "Penpot MCP conectado"]
rules: ["RN-CIT-01", "RN-CIT-03", "RN-HIS-01", "RN-HIS-02", "RN-RES-02", "RN-RES-03", "RN-DAT-01", "RN-DAT-02", "RN-DIS-07", "RN-TEN-01"]
decisions: ["DEC-004", "DEC-007", "DEC-014", "DEC-016", "DEC-041", "DEC-074", "DEC-075", "DEC-076", "DEC-077", "DEC-078", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-064-01", "CA-064-02", "CA-064-03", "CA-064-04", "CA-064-07", "CA-064-08", "CA-065-03", "CA-065-04", "CA-065-06", "CA-065-08"]
source_docs:
  - "AGENTS.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/estados-citas.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/detalle-turno-eventos/README.md"
  - "apps/web/src/modules/agenda"
  - "apps/web/src/styles/tokens.css"
created_at: "2026-09-06"
updated_at: "2026-09-06"
supersedes: null
superseded_by: null
---

# Genera el mockup editable de Detalle de turno con Penpot

## Instrucción para Codex

Genera o consolida en Penpot el paquete de diseño de `/panel/turnos/:appointmentId`: ficha, historial paginado y diálogo de reprogramación. No implementes el diseño en Vue. El resultado debe ser una fuente editable y exportaciones que luego puedan asignarse como contrato visual al issue #191.

## Perfil de ejecución

```text
Modelo: gpt-5.6-sol
Esfuerzo: medium
MCP: Penpot MCP local
```

No uses Astra ni subas el esfuerzo en el primer pase. Verifica que Penpot exponga inspección, escritura y exportación. Empieza con `high_level_overview`; no muestres claves MCP.

## Preflight obligatorio

1. Espera a que #190 esté integrado con CI verde. Lee por completo `AGENTS.md` y los `source_docs`.
2. Comprueba árbol limpio, actualiza `main` por fast-forward y crea `chore/191-mockup-detalle-turno` sin descartar cambios ajenos.
3. Inspecciona ruta, página Vue, componentes, pruebas y tipos para inventariar únicamente estados y textos reales.
4. Abre visualmente el atlas `detalle-turno-eventos`. Reutilízalo como insumo válido si está completo; no lo dupliques a ciegas. El objetivo es llevar la fuente a Penpot y resolver inconsistencias documentadas, no rediseñar funcionalidad.
5. En el archivo de Penpot conectado, crea o reutiliza `Rutas pendientes` y una sección `191/Detalle-turno`.

## Estados que debe representar

- principal `confirmed` con ficha, contacto privado autorizado, snapshots e historial completo;
- carga de detalle, `404` uniforme y error recuperable;
- historial cargando, vacío, error y paginación con “Cargar más”;
- terminales `completed`, `cancelled_by_customer`, `cancelled_by_barber` y `no_show`, sin acciones futuras;
- diálogo inicial, guardando, conflicto de agenda, versión obsoleta, error de red y éxito aplicado.

Los estados que solo cambien copy pueden documentarse como variantes de un mismo componente; no crees láminas redundantes.

## Trabajo requerido

1. Crea componentes editables para shell, ficha, hechos, badge, evento de historial, alerta, campo y diálogo. Usa layout flexible, nombres `191/<componente>/<variante>` y los tokens NAVA aprobados.
2. Resuelve primero el estado principal en 320×800, 360×800, 768×1024 y 1280×1024. Deriva cada estado secundario material en 360×800 y 1280×1024.
3. Mantén el término visible “turno”; nunca dibujes `appointment`, token de versión, IDs, datos de otro tenant o valores técnicos. Los cambios de horario se presentan en la zona de la barbería.
4. Conserva fecha y barbero al volver a la agenda. Reprogramar solo aparece donde la capacidad vigente lo permite; no agregues completar, cancelar, cambiar servicio, duración, persona o notificaciones.
5. Revisa visualmente cada frame exportable, contraste AA, jerarquía, clipping, textos largos, teclado/foco representado y objetivos táctiles.
6. Exporta PNG al atlas vigente en `docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/detalle-turno-eventos/{desktop,mobile}/detalle-turno/` y actualiza allí el README con matriz de estados, decisiones, exclusiones, página/sección/frame de Penpot y procedencia.

## Fuera de alcance

- Cambiar `apps/web`, API, backend, migraciones, paginación, concurrencia o reglas T2.
- Inventar T3, completar, cancelar, `no_show` como acción, notificaciones o botones decorativos.
- Usar datos personales reales o declarar cumplidos criterios funcionales solo porque existen mockups.

## Verificación y entrega

Abre cada PNG exportado y compáralo con su frame. Ejecuta `git diff --check`, revisa que solo existan prompts/evidencia/documentación y entrega `Estado | Viewport | Frame Penpot | PNG | Revisión`. Actualiza este prompt y el índice con estado, rama y PR reales. Commit/PR: `chore(design): genera mockups de detalle de turno`; usa `Refs #191`, no `Closes #191`.
