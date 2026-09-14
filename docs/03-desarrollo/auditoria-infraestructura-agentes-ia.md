---
titulo: "Auditoría y arquitectura de infraestructura de agentes (Claude Code y Codex)"
version: "1.1"
estado: "Vigente; aplicada por el issue #259"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-09-13"
documentos_relacionados:
  - "../../AGENTS.md"
  - "../../CLAUDE.md"
  - "../00-control/registro-decisiones.md"
  - "estandar-diseno-visual.md"
  - "estrategia-pruebas.md"
---

# Auditoría y arquitectura propuesta para Claude Code y Codex

**Estado:** aplicado mediante el issue [#259](https://github.com/bcaceres19/barberia/issues/259) y `DEC-086`. Las desviaciones respecto del borrador original se registran en la sección 12.  
**Fecha:** 2026-09-13.  
**Alcance:** instrucciones, skills, prompts, herramientas de diseño, revisión, pruebas visuales, hooks, MCP y coste de contexto.  
**No incluye:** cambios funcionales de NAVA, dependencias de producto, configuración de CI ni eliminación de artefactos existentes.

## Resumen ejecutivo

El repositorio ya tiene mejores fundamentos de producto, diseño y pruebas que la mayoría de paquetes externos evaluados. El problema principal no es falta de conocimiento: es que ese conocimiento se activa de forma desigual entre Claude y Codex, mientras Graphify añade un coste automático desproporcionado.

La arquitectura recomendada conserva `AGENTS.md` como autoridad compartida, usa cuatro skills canónicos bajo `.agents/skills/`, expone adaptadores mínimos a Claude en `.claude/skills/`, y mantiene NAVA en sus documentos normativos en vez de copiarlo a otro “design-system skill”. El skill existente `generacion-mockups-nava` completa un catálogo total de cinco capacidades de proyecto.

No se recomienda instalar el plugin completo Product Design de OpenAI ni el repositorio de skills de Anthropic. Se adaptan solamente sus patrones fuertes: brief mínimo, direcciones visuales diferenciadas, auditoría basada en capturas y QA esperado-versus-real. Figma se conserva como integración bajo demanda, no como contexto global obligatorio.

## 1. Diagnóstico del sistema actual

### Inventario relevante

| Elemento | Estado observado | Diagnóstico |
|---|---:|---|
| `AGENTS.md` | 62 líneas / 6.1 KB | Tamaño sano y autoridad correcta; no necesita una reducción agresiva. |
| `CLAUDE.md` | 28 líneas / 2.9 KB | Corto, pero duplica reglas de `AGENTS.md` y fuerza Graphify en trabajo rutinario. |
| `.claude/CLAUDE.md` | 7 líneas / 620 B | Segunda capa redundante que vuelve a activar Graphify y fidelidad visual. |
| `.claude/settings.json` | 25 líneas | Hooks `PreToolUse` disparan Graphify en lectura, búsqueda y shell; el archivo además contiene cambios locales del usuario. |
| skill `graphify` | 737 líneas / 41.8 KB | Demasiado grande, amplio y operativo para ser contexto rutinario. Incluye red, APIs externas y comandos de mantenimiento; debe ser explícito. |
| skills visuales Claude | 2 skills / 137 líneas | Buenos y concretos, pero exclusivos de Claude y parcialmente solapados. |
| skill de mockups Codex | 54 líneas | Bien delimitado y útil; genera artefactos visuales, no código. |
| prompts persistentes | 112 archivos / 1.4 MB | Trazabilidad valiosa; deben descubrirse por issue/HU, no cargarse en bloque. |
| `graphify-out/` | 107 MB | El grafo consultado devolvió una respuesta amplia y truncada; no justifica uso obligatorio en cada tarea. |
| evidencia de backlog | 148 MB | Correcta como evidencia bajo demanda, inadecuada como contexto inicial. |

### Fortalezas existentes

- `docs/03-desarrollo/estandar-diseno-visual.md` ya define identidad, color, tipografía, composición, accesibilidad y dos modos de conformidad.
- `docs/03-desarrollo/especificacion-frontend-nava.md` ya contiene inventario de pantallas, componentes, estados, voz y comportamiento responsive.
- `docs/03-desarrollo/estrategia-pruebas.md` exige componentes, E2E, evidencia visual, 320/360/768/1280 px, zoom 200 %, teclado y contraste.
- Vue, Vitest, Playwright y axe ya están instalados; no hace falta otra dependencia para el ciclo visual.
- La disciplina de decisiones, reglas, historias, issues y prompts persistentes evita que un agente convierta un mockup en alcance funcional.

### Duplicaciones y contradicciones

1. `CLAUDE.md` repite prompts persistentes y fidelidad ya gobernados por `AGENTS.md` y los documentos normativos.
2. Claude tiene dos skills visuales que Codex no descubre; Codex tiene el skill de mockups que Claude no descubre.
3. Graphify está descrito como skill especializado, pero las instrucciones y hooks lo convierten en requisito de casi toda operación.
4. La consulta obligatoria del grafo contradice el objetivo de progressive disclosure: expande contexto antes de demostrar que el grafo es necesario.
5. Las decisiones visuales mencionan nombres de skills exclusivos de Claude; la capacidad debería ser compartida y el documento normativo no debería depender del proveedor.

## 2. Comparación: Anthropic, OpenAI, comunidad y soluciones propias

| Fuente | Aporte fuerte | Limitación para NAVA | Decisión |
|---|---|---|---|
| Anthropic `frontend-design` | Parte de producto, audiencia y dirección; combate estética genérica; incluye segunda pasada crítica. | Es demostrativo y genérico; no conoce NAVA, P0, fidelity mode ni pruebas del repo. | Adaptar principios, no instalar el paquete. |
| Anthropic `canvas-design` | Composición gráfica y salida PNG/PDF. | Está orientado a artefactos estáticos, no a UI Vue operativa. | Descartar para producto; usar image generation solo cuando se pidan mockups. |
| Anthropic `web-artifacts-builder` | Montaje rápido de artefactos web. | Inicializa React/Tailwind/shadcn e instala dependencias; choca con Vue y la arquitectura existente. | No instalar. |
| Anthropic skill/plugin development | Buen estándar abierto y empaquetado. | Duplicaría herramientas de creación disponibles en cada agente. | Usar como referencia de compatibilidad, no copiar al repo. |
| OpenAI Product Design | Brief mínimo, tres opciones visuales, audit y `design-qa` con comparación real. | Plugin grande, orientado a prototipos y Work Mode; impone archivos/flujo propios y duplica NAVA/Playwright. | Adaptar los gates útiles en dos skills pequeños. |
| OpenAI Figma | Design-to-code, creación de vistas, tokens/componentes y capturas verificadas. | Requiere cuenta, permisos y un archivo objetivo; escribir en Figma es estado externo. | Útil bajo demanda; no añadir MCP repo-global. |
| OpenAI `skill-creator` | Buen diseño de triggers, progressive disclosure y validación. | Es herramienta del agente, no conocimiento del producto. | Usarlo para mantener los skills, no versionarlo otra vez. |
| Graphify comunitario/local | Puede ayudar en cambios realmente transversales y code intelligence. | 737 líneas, hooks frecuentes, salida grande, red/API opcional y señal ruidosa en esta auditoría. | Mantener opcional; retirar activación automática. |
| Skills propios | Integran autoridad, NAVA, issue/HU/RN/DEC y pruebas reales del repositorio. | Requieren mantenimiento disciplinado. | Mejor ajuste si se limita el catálogo a cinco. |

Fuentes primarias: [Anthropic Skills](https://code.claude.com/docs/en/skills), [Anthropic memory/CLAUDE.md](https://code.claude.com/docs/en/memory), [Anthropic hooks](https://code.claude.com/docs/en/hooks), [repositorio de ejemplos de Anthropic](https://github.com/anthropics/skills), [Codex Skills](https://learn.chatgpt.com/docs/build-skills), [Codex AGENTS.md](https://learn.chatgpt.com/docs/agent-configuration/agents-md), [plugins oficiales de OpenAI](https://github.com/openai/plugins) y [Product Design `design-qa`](https://github.com/openai/plugins/blob/1dc195897af4161d039b80d8471ec0a10c9bbc89/plugins/product-design/skills/design-qa/SKILL.md).

## 3. Tabla de skills: ESSENTIAL / USEFUL / OPTIONAL / REDUNDANT / REMOVE

| Capacidad o skill | Clase | Motivo |
|---|---|---|
| `task-brief` propuesto | **ESSENTIAL** | Fusiona prompt architect, discovery y planificación adaptativa sin crear documentos grandes. |
| `ui-direction` propuesto | **ESSENTIAL** | Fusiona creative director y routing del design system; obliga a direcciones compositivas, no simples recolores. |
| `visual-qa` propuesto | **ESSENTIAL** | Une design critic, UX review, viewport verification y comparación esperada-real con evidencia. |
| `change-review` propuesto | **ESSENTIAL** | Une architecture, frontend y code review con lentes condicionales y comportamiento read-only. |
| `generacion-mockups-nava` existente | **USEFUL** | Produce atlas raster cuando la exploración realmente necesita imágenes; está bien acotado. |
| `nava-mockup-fidelity` existente | **USEFUL → REDUNDANT** | Absorbido por `visual-qa` (gates A–D en `references/fidelity-gates.md`). Queda como alias no invocable por el modelo, porque más de 50 prompts persistentes citan su ruta. |
| `browser-viewport-verification` existente | **USEFUL → REDUNDANT** | Absorbido por `visual-qa` (`references/viewport-verification.md`). Queda como alias no invocable por el modelo por el mismo motivo. |
| `graphify` existente | **OPTIONAL** | Útil solo para preguntas transversales; no debe ser precondición ni hook global. |
| `prompt-architect` separado | **REDUNDANT** | Queda dentro de `task-brief`; separarlo añade otro trigger y handoff. |
| `project-context` separado | **REDUNDANT** | Un índice documental y el discovery de `task-brief` son suficientes. |
| `design-system` separado | **REDUNDANT** | NAVA ya está definido de forma normativa; duplicarlo generaría drift. |
| `creative-director` separado | **REDUNDANT** | Se integra en `ui-direction`. |
| `design-critic` separado | **REDUNDANT** | Se integra en `visual-qa`. |
| `ux-reviewer` separado | **REDUNDANT** | Es un lente de `visual-qa`, no otro paquete. |
| `architecture-reviewer` separado | **REDUNDANT** | Es un lente de `change-review`. |
| `frontend-reviewer` separado | **REDUNDANT** | Es un lente de `change-review`. |
| `code-reviewer` separado | **REDUNDANT** | Es la responsabilidad base de `change-review`. |
| Anthropic `frontend-design` íntegro | **OPTIONAL / no instalar** | Buena inspiración; un skill NAVA pequeño aporta mejor señal y menos contexto. |
| Anthropic `canvas-design` | **REMOVE / no instalar** | Arte estático, no UI de producto. |
| Anthropic `web-artifacts-builder` | **REMOVE / no instalar** | Stack y bootstrap incompatibles; dependencia y red innecesarias. |
| OpenAI Product Design completo | **OPTIONAL / no instalar** | Potente para prototipos generales; demasiado amplio y duplicado para el repo. |
| Figma oficial | **USEFUL** | Excelente cuando hay Figma explícito; acceso bajo demanda y con confirmación de destino. |

## 4. Arquitectura propuesta de agentes

```text
Solicitud
  └─ AGENTS.md (autoridad siempre activa)
      ├─ task-brief (solo si hay ambigüedad/riesgo)
      │   └─ issue + DEC/RN/HU + código mínimo afectado
      ├─ ui-direction (solo UI nueva/rediseño sin target exacto)
      │   └─ NAVA + ruta/estado + 2–3 direcciones
      ├─ implementación normal del agente
      ├─ visual-qa (solo cambios visibles renderizados)
      │   └─ Playwright/Figma/screenshots bajo demanda
      └─ change-review (revisión independiente/pre-PR)
```

Los skills son capacidades, no una tubería obligatoria. Una corrección pequeña puede pasar directamente de instrucciones a implementación y test; una migración crítica usa `task-brief` y `change-review`; una pantalla nueva usa el ciclo completo.

### Skill, subagente, comando, hook, MCP o script

| Necesidad | Mecanismo | Razón |
|---|---|---|
| Interpretar requisitos y elegir contexto | Skill | Requiere juicio contextual. |
| Dirección creativa | Skill; subagente opcional si se pide revisión independiente | Requiere razonamiento, no ejecución determinística. |
| Review independiente | Skill o subagente explícito | Conviene aislamiento, pero no siempre justifica otro agente. |
| Validar estructura/frontmatter/adaptadores | Script | Resultado determinístico, rápido y barato. |
| Lint/typecheck/tests | Scripts npm/Go existentes y CI | Ya existen; no encapsularlos en prosa. |
| Visual QA | Skill que orquesta Playwright y capturas | Mezcla juicio visual con acciones repetibles. |
| Figma | Plugin/MCP oficial bajo demanda | Es una integración externa con estado y permisos. |
| Reglas normativas | `AGENTS.md` y documentación | Deben ser autoridad estable, no workflow opcional. |
| Graphify | Comando/skill explícito | La señal no justifica hooks sobre cada operación. |

## 5. Arquitectura de contexto y progressive disclosure

### Siempre activo

- `AGENTS.md` solamente como autoridad común.
- `CLAUDE.md` importa `@AGENTS.md` y conserva solo lo específico de Claude: adaptadores, Graphify opcional y la autorización permanente de merge del propietario.
- Los nombres y descripciones de skills; nunca sus cuerpos completos hasta activación.

### Bajo demanda

- Issue y prompt exactos.
- Secciones relevantes de decisiones, alcance, reglas, estándares y contrato.
- Skill especializado que coincide con el trabajo.
- Código vecino y pruebas de la capacidad afectada.
- Capturas, Figma, atlas y evidencias solo en tareas visuales.
- Graphify solo cuando una búsqueda textual y el ownership local no resuelven un cambio transversal.

Claude recomienda mantener `CLAUDE.md` por debajo de 200 líneas y permite importar `@AGENTS.md`; los imports se cargan, por lo que importar documentos grandes no ahorra tokens. Codex concatena instrucciones desde raíz hasta el directorio actual con un presupuesto predeterminado de 32 KiB. Ambos usan descriptions de skills para decidir activación y cargan el cuerpo después, por lo cual los triggers deben ser discriminantes. Véanse [Anthropic memory](https://code.claude.com/docs/en/memory), [Anthropic Skills](https://code.claude.com/docs/en/skills), [Codex AGENTS.md](https://learn.chatgpt.com/docs/agent-configuration/agents-md) y [Codex Skills](https://learn.chatgpt.com/docs/build-skills).

## 6. Arquitectura compartida Claude / Codex

```text
.agents/skills/                 # contenido canónico compatible con Agent Skills
  task-brief/SKILL.md
  ui-direction/SKILL.md
  visual-qa/SKILL.md
  change-review/SKILL.md
  generacion-mockups-nava/...

.claude/skills/                 # adaptadores mínimos; no segunda implementación
  <skill>/SKILL.md              # frontmatter idéntico + instrucción de leer el canónico
  nava-mockup-fidelity/         # alias heredado (disable-model-invocation)
  browser-viewport-verification/# alias heredado (disable-model-invocation)
  graphify/                     # herramienta opcional, solo /graphify

AGENTS.md                       # reglas compartidas
CLAUDE.md                       # @AGENTS.md + excepciones Claude
tools/ai/validate-agent-system.sh
```

### Compartido

- Objetivos, triggers, workflows, rúbricas y referencias normativas.
- Autoridad de `AGENTS.md`.
- Scripts determinísticos y evidencia producida por tests.

### Específico de Claude

- Adaptadores de descubrimiento en `.claude/skills`.
- Hooks y permisos de `.claude/settings.json`.
- Importación `@AGENTS.md` y cualquier instrucción de UI propia del cliente.

### Específico de Codex

- Descubrimiento canónico en `.agents/skills`.
- Plugins del entorno, incluida Figma, sin configuración duplicada en el repo.
- Skills del sistema como `skill-creator` y `openai-docs`, que no deben copiarse al proyecto.

No se recomienda `.ai/` en la primera iteración: añadiría otra raíz y otro adaptador sin contenido compartido adicional. Si luego se publica un paquete reutilizable entre proyectos, su fuente puede vivir en un repositorio dedicado y generar/verificar ambos adaptadores; NAVA seguirá dentro de este repositorio.

## 7. Arquitectura de prompting

`task-brief` adapta su profundidad:

- **Pequeña:** objetivo + verificación, sin plantilla visible.
- **Estándar:** objetivo, usuario/contexto, alcance, requisitos, restricciones, aceptación, verificación y unknowns.
- **Crítica:** añade riesgos, rollback, observabilidad y decisiones pendientes.

Reglas esenciales:

1. Resolver dudas con fuentes normativas antes de preguntar.
2. Hacer una suposición solo si es reversible y no cambia alcance, seguridad, contrato o semántica de datos.
3. Hacer una sola pregunta enfocada si la respuesta cambia materialmente la solución.
4. No inventar números de issue, `RN-*`, `DEC-*`, `HU-*` ni comportamientos API.
5. Persistir únicamente prompts destinados a ejecución futura, usando el catálogo actual.

Esto evita tanto “programar antes de pensar” como transformar cada petición pequeña en un PRD de varias páginas.

## 8. Workflow visual propuesto

### Sin referencia exacta: identidad guiada

```text
brief de producto → 2 direcciones (3 solo si el impacto lo amerita)
→ recomendación → contrato visual compacto → implementación
→ render 320/360/768/1280 → estados/a11y → crítica → corrección
```

Las direcciones deben cambiar composición o interacción: por ejemplo, ritmo editorial temporal frente a flujo guiado de decisión. Un cambio de paleta no constituye una dirección.

### Con referencia exacta: fidelidad

```text
abrir original → medir viewport/estado/densidad → baseline real
→ implementar → captura real → side-by-side + overlay/diff
→ corregir P0–P2 → recapturar el mismo estado → pass/block
```

Dos iteraciones son el valor predeterminado; una tercera se justifica si un P1/P2 converge. No existe bucle ilimitado. La rúbrica combina jerarquía, tipografía, spacing, color, assets, contenido, estados, responsive, teclado, foco, contraste, motion y consola.

Figma entra solamente cuando el usuario proporciona/solicita un archivo, frame, biblioteca o entrega Figma. El plugin actual tiene permiso “Allow low-risk actions” heredado del valor global; para trabajos reales debe confirmarse el archivo/nodo destino y evitar escrituras amplias no solicitadas.

## 9. Seguridad de skills y herramientas externas

### Criterio de admisión

Antes de integrar: revisar frontmatter, instrucciones completas, referencias cargadas, scripts, manifest, red, credenciales, escrituras, borrados, instalación de dependencias, contenido externo no confiable y mantenimiento.

### Hallazgos

- **Graphify:** el skill incluye operaciones de red, posible Gemini mediante variable de entorno, clonación y comandos de limpieza. Los hooks de Claude ejecutan comandos con permisos del usuario antes de herramientas frecuentes. Debe quedar explícito y sus scripts no deben ejecutarse por defecto.
- **Anthropic web-artifacts-builder:** instala/usa un stack alternativo; se rechaza por superficie de red, dependencias y conflicto arquitectónico.
- **OpenAI Product Design:** oficial y mantenido, pero con capacidades de lectura/escritura, navegador, generación, publicación y convenciones propias. Instalarlo completo expande permisos y contexto sin necesidad.
- **Figma:** integración oficial útil, pero toda escritura modifica estado externo. Usar lectura primero, confirmar archivo/nodo y aplicar permisos de “ask before writes” si se desea un límite más estricto.
- **Skills propuestos:** solo instrucciones; no red, paquetes ni comandos destructivos. El único script propuesto es read-only y valida archivos locales.

Anthropic advierte que los hooks de comandos se ejecutan con permisos completos del usuario, por lo que deben tratarse como código de producción: [seguridad de hooks](https://code.claude.com/docs/en/hooks).

## 10. Plan de implementación por fases

| Fase | Contenido | Estado |
|---|---|---|
| 1 — issue y rama | Issue #259, rama `chore/259-infraestructura-agentes` en un worktree separado, sin mezclar la HU en curso. | Hecha |
| 2 — núcleo mínimo | Cuatro skills canónicos, adaptadores Claude (incluido `generacion-mockups-nava`), `CLAUDE.md` reducido, catálogo en `AGENTS.md` y validador. | Hecha |
| 3 — migración visual | `visual-qa` incorpora íntegros el flujo de fidelidad, las condiciones de fallo, las desviaciones permitidas, los gates A–D y la verificación de viewport. Los skills antiguos pasan a alias. `DEC-086` actualiza la ayuda de ejecución de `DEC-080`. | Hecha |
| 4 — Graphify | Hooks `PreToolUse` retirados en un commit separado; skill `graphify` solo invocable con `/graphify`. `graphify-out/` deja de versionarse y se regenera en local (`DEC-087`, #261). | Hecha |
| 5 — pruebas y refinamiento | Validador determinístico con casos negativos, descubrimiento real en Claude Code y Codex. Medir activaciones erróneas en 5–10 tareas reales y ajustar descripciones, sin añadir skills. | Validación inicial hecha; la medición continua queda pendiente |

## 11. Cambios implementados, descartados y pendientes

### Implementados

- `.agents/skills/{task-brief,ui-direction,visual-qa,change-review}/SKILL.md` y `visual-qa/references/{fidelity-gates.md,viewport-verification.md}`.
- `.claude/skills/<skill>/SKILL.md` como adaptadores con descripción idéntica para los cinco skills canónicos.
- Alias `nava-mockup-fidelity` (incluida la ruta `references/acceptance-gates.md`) y `browser-viewport-verification`.
- `CLAUDE.md` con `@AGENTS.md`; retiro de `.claude/CLAUDE.md`; sección "Skills de agentes" en `AGENTS.md`.
- `.claude/settings.json` sin hooks; `graphify` con `disable-model-invocation: true`.
- `tools/ai/validate-agent-system.sh [--strict]`.

### Deliberadamente no instalados

- Plugin Product Design completo de OpenAI.
- Repositorio/marketplace completo de skills de Anthropic.
- `web-artifacts-builder`, `canvas-design` y otro skill de design system.
- Nuevas dependencias, MCP repo-global, hooks caros, browser adicional o framework visual.

### Pendientes

- ~~Decidir el versionado de `graphify-out/`~~ Resuelto por `DEC-087` (issue [#261](https://github.com/bcaceres19/barberia/issues/261)): la carpeta se ignora y se regenera solo en local.
- Medir durante 5–10 tareas reales las activaciones erróneas y ajustar solo las descripciones.
- Una reinstalación o actualización de Graphify (`graphify install`) puede restaurar sus hooks, su bloque en `CLAUDE.md` o la invocación automática del skill. Tras actualizarlo, ejecutar el validador con `--strict`, que lo detecta.

### Criterios de validación final

- Codex descubre los skills canónicos por descripción y Claude descubre los adaptadores.
- Los cuerpos no se cargan hasta activación.
- Un prompt de backend no activa skills visuales.
- "Hazlo más moderno" activa brief y dirección visual, pero una corrección CSS exacta no abre una exploración completa.
- Una referencia exacta salta direcciones alternativas y activa `visual-qa` en modo fidelidad.
- La revisión permanece read-only si no se pide corregir.
- `tools/ai/validate-agent-system.sh --strict` pasa sin fallos ni advertencias.
- No se duplican tokens, tipografía ni componentes NAVA dentro de los skills.

## 12. Desviaciones respecto del borrador original

| Borrador | Aplicado | Motivo |
|---|---|---|
| Adaptadores Claude con `@../../../.agents/skills/<skill>/SKILL.md` | Adaptador con la instrucción explícita de leer el canónico y resolver `references/` desde allí | La documentación de Claude Code solo expande `@imports` en `CLAUDE.md`, no en el cuerpo de un `SKILL.md`: el adaptador habría mostrado una línea literal. Los symlinks se descartan porque el repositorio también se usa desde Windows. |
| `CLAUDE.md` con "no mergear sin petición explícita" | Se conserva literal la autorización permanente de squash-merge con CI en verde | La regla del borrador contradecía una instrucción vigente del propietario (2026-08-13). |
| Retirar `nava-mockup-fidelity` y `browser-viewport-verification` | Alias con `disable-model-invocation: true` que redirigen a `visual-qa` | Más de 50 prompts persistentes citan esas rutas, y `AGENTS.md` prohíbe reescribir en silencio prompts usados por otro agente. Como alias no compiten por activación. |
| `visual-qa` resumido | Incorpora íntegros el flujo de 10 pasos, qué debe coincidir, las desviaciones permitidas, las condiciones de fallo, los gates A–D con tolerancias y la verificación de viewport con el arnés Playwright | Criterio de equivalencia de la fase 3: el borrador omitía tolerancias, gates y el fallo conocido de `resize_window`. |
| Validador con `rg` y comprobación de `@import` | `grep`/`awk` portables; descripciones idénticas entre canónico y adaptador, enlaces relativos existentes, skills solo de Claude no invocables, rutas absolutas en settings y modo `--strict` | `rg` no está garantizado en todas las máquinas ni en CI, y la igualdad de descripciones evita desvíos silenciosos. |
| Graphify: solo retirar hooks | También `disable-model-invocation` en el skill `graphify` | Su descripción pedía usar el grafo "para cualquier pregunta sobre el código", lo que mantenía la activación automática sin hooks. |
| `generacion-mockups-nava` sin cambios | Frontmatter entre comillas y una línea para agentes sin generación raster | Claude Code no tiene ImageGen: debe declararlo y detenerse en vez de sustituir las imágenes por HTML o SVG. |
