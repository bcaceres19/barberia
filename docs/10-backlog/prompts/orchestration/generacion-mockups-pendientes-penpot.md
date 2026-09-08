---
prompt_id: "PROMPT-ORCH-MOCKUPS-PENDIENTES-PENPOT-v2"
version: "2.0"
kind: "orchestration"
status: "blocked"
target_agents:
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu: ["HU-020", "HU-021", "HU-022", "HU-023", "HU-024", "HU-040", "HU-041", "HU-042", "HU-064", "HU-065"]
issue: 184
issue_url: "https://github.com/bcaceres19/barberia/issues/184"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on:
  - "Issue #190 integrado en main con CI verde"
  - "Penpot MCP conectado al archivo de diseño correcto"
  - "Cada paquete se integra antes de iniciar el siguiente"
rules: ["RN-TEN-01", "RN-DAT-01", "RN-DAT-02", "RN-IDE-01", "RN-SER-01", "RN-SER-02", "RN-SER-03", "RN-SER-04", "RN-DIS-01", "RN-DIS-02", "RN-DIS-03", "RN-DIS-05", "RN-DIS-07", "RN-BLQ-01", "RN-BLQ-02", "RN-BLQ-03", "RN-BLQ-04", "RN-CIT-03", "RN-HIS-01", "RN-HIS-02"]
decisions: ["DEC-016", "DEC-019", "DEC-020", "DEC-047", "DEC-067", "DEC-068", "DEC-069", "DEC-070", "DEC-076", "DEC-077", "DEC-078", "DEC-079", "DEC-080"]
acceptance_criteria:
  - "CA-ORCH-MP-01: cada ruta obtiene un paquete editable en Penpot y exportaciones rastreables sin modificar la aplicación."
  - "CA-ORCH-MP-02: la secuencia #191 a #197 respeta los bloqueos declarados y cada entrega usa issue, rama y PR propios."
  - "CA-ORCH-MP-03: cada mockup representa solo funciones y estados demostrables en HU, reglas, contrato o código vigente."
  - "CA-ORCH-MP-04: el estado principal existe en 320, 360, 768 y 1280 px; los estados secundarios materiales existen en 360 y 1280 px."
  - "CA-ORCH-MP-05: cada paquete incluye README, inventario de estados, decisiones visuales, exclusiones y exportaciones revisadas visualmente."
source_docs:
  - "AGENTS.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/10-backlog/prompts/README.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/panel-agenda-eventos/README.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/nuevo-turno-eventos/README.md"
created_at: "2026-09-06"
updated_at: "2026-09-07"
supersedes: "prompt de orquestación v1 retirado por estructura separada"
superseded_by: null
---

# Generación secuencial de mockups pendientes con Penpot MCP

## Instrucción para Codex

Genera en Penpot los mockups faltantes de las rutas #191–#197. Esta ejecución produce diseño editable, exportaciones y documentación de diseño; **no modifica `apps/web`, backend, OpenAPI, migraciones, pruebas de producto ni comportamiento**. Los prompts `*-fidelidad.md` existentes pertenecen a la implementación posterior y no son sustituidos por esta serie.

## Perfil de ejecución obligatorio

```text
Modelo: gpt-5.6-sol
Esfuerzo de razonamiento: medium
MCP: Penpot MCP, preferentemente local
```

Usa `gpt-5.6-sol/medium` como autor principal: tiene visión, herramientas y MCP, y ofrece suficiente criterio creativo sin el consumo de una sesión Astra de alta intensidad. No escales silenciosamente de modelo ni de esfuerzo. Si esa combinación no está disponible, informa el bloqueo antes de escribir en Penpot.

El modo local de Penpot MCP es el recomendado porque permite exportar al repositorio. Deben estar disponibles, como mínimo, las capacidades equivalentes a `high_level_overview`, `execute_code`, `penpot_api_info` y `export_shape`. Nunca copies al chat, archivos, logs o capturas una clave MCP. La única preparación manual admisible es abrir el archivo correcto de Penpot y conectar su plugin; desde ahí Codex realiza el trabajo.

## Preflight

1. Lee por completo `AGENTS.md` y todos los `source_docs`; revisa dudas y contradicciones aplicables.
2. Comprueba `main`, remoto, árbol de trabajo y estado real de la dependencia. No descartes ni sobrescribas cambios ajenos.
3. No inicies #191 hasta que #190 esté integrado en `main` con CI verde. Después, cada issue espera la integración del anterior.
4. Verifica el modelo efectivo y el esfuerzo de la tarea.
5. Con el archivo correcto abierto y el plugin conectado, ejecuta primero una inspección de solo lectura de páginas, componentes, tokens y página enfocada.
6. Crea o reutiliza una página dedicada llamada `Rutas pendientes`. Si la API no puede cambiar/crear página, pide únicamente que el usuario la cree o enfoque y continúa al recibirla.
7. Conserva páginas y frames ajenos. Todas las escrituras viven en secciones con prefijo `<issue>/`.

## Sistema común y economía de ejecución

- Reutiliza componentes, estilos y tokens Penpot; no dibujes cada frame desde cero ni dupliques componentes desconectados.
- Toma como familia visual los atlas aprobados de Agenda, Nuevo turno y Detalle, junto con `DEC-077`–`DEC-080` y los tokens reales. Trabaja en **identidad guiada**: estas rutas aún no tienen un mockup exacto nuevo que copiar.
- Primero resuelve el estado principal como componente. Luego deriva estados y viewports por variantes, conservando nombres, orden y contenido.
- Crea cuatro frames del estado principal: 320×800, 360×800, 768×1024 y 1280×1024. Crea solo 360×800 y 1280×1024 para cada estado secundario que cambie materialmente la composición.
- Usa contenido ficticio coherente entre rutas. No uses datos personales reales, identificadores internos, tokens, credenciales ni capturas de producción.
- Antes de exportar, abre y revisa visualmente cada frame. Corrige clipping, overflow, contraste, jerarquía, foco visible representado, objetivos táctiles y divergencias entre variantes.

## Secuencia vinculante

| Orden | Issue | Ruta | Prompt hijo |
| --- | --- | --- | --- |
| 1 | #191 | `/panel/turnos/:appointmentId` | [`issue-191-generacion-mockup-detalle-turno-penpot.md`](../chore/issue-191-generacion-mockup-detalle-turno-penpot.md) |
| 2 | #192 | `/panel/barberos` | [`issue-192-generacion-mockup-barberos-penpot.md`](../chore/issue-192-generacion-mockup-barberos-penpot.md) |
| 3 | #193 | `/panel/barberia` | [`issue-193-generacion-mockup-barberia-penpot.md`](../chore/issue-193-generacion-mockup-barberia-penpot.md) |
| 4 | #194 | `/panel/servicios` | [`issue-194-generacion-mockup-servicios-penpot.md`](../chore/issue-194-generacion-mockup-servicios-penpot.md) |
| 5 | #195 | `/panel/servicios-por-barbero` | [`issue-195-generacion-mockup-servicios-barbero-penpot.md`](../chore/issue-195-generacion-mockup-servicios-barbero-penpot.md) |
| 6 | #196 | `/panel/horarios` | [`issue-196-generacion-mockup-horarios-penpot.md`](../chore/issue-196-generacion-mockup-horarios-penpot.md) |
| 7 | #197 | `/panel/bloqueos` | [`issue-197-generacion-mockup-bloqueos-penpot.md`](../chore/issue-197-generacion-mockup-bloqueos-penpot.md) |

No generes todos los paquetes en una rama ni adelantes un prompt bloqueado. Al terminar cada hijo, integra su PR conforme a las reglas del repositorio, actualiza `main` y recién entonces pasa al siguiente.

## Forma de entrega por paquete

Cada prompt hijo debe dejar:

1. sección y frames editables en `Rutas pendientes`;
2. exportaciones PNG compactadas en el atlas vigente: `docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/<ruta>-eventos/{desktop,mobile}/<ruta>/`;
3. un `README.md` con matriz `evento → viewport → frame → exportación`, decisiones visuales, fuentes, desviaciones y exclusiones;
4. enlace o identificadores no secretos del archivo, página y sección de Penpot;
5. metadatos e índice de prompts actualizados con estado, rama y PR reales.

Los PNG son artefactos de diseño, no prueba de que la app ya cumple. Usa `Refs #<issue>`: el paquete no cierra por sí solo el issue de rediseño.

## Entrega final

Entrega una tabla `Issue | Ruta | Sección Penpot | Estados | Viewports | Evidencia | PR | Resultado`. No declares completa la serie si falta una exportación, un README o una dependencia no se integró.

## Texto corto para iniciar

```text
Ejecuta `docs/10-backlog/prompts/orchestration/generacion-mockups-pendientes-penpot.md` con gpt-5.6-sol y esfuerzo medium. Usa Penpot MCP local, comienza con inspección de solo lectura y genera únicamente los paquetes de mockups de #191 a #197, uno por rama y PR, respetando sus bloqueos. No cambies la aplicación ni cierres los issues: entrega Penpot editable, PNG revisados y README trazable por ruta.
```
