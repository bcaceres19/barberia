---
prompt_id: "PROMPT-ORCH-NAVA-ATLAS-v2"
version: "2.1"
kind: "orchestration"
status: "superseded"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu:
  - "HU-005"
  - "HU-006"
  - "HU-007"
  - "HU-008"
  - "HU-009"
  - "HU-010"
  - "HU-011"
  - "HU-012"
  - "HU-020"
  - "HU-021"
  - "HU-022"
  - "HU-023"
  - "HU-024"
  - "HU-040"
  - "HU-041"
  - "HU-042"
  - "HU-060"
  - "HU-061"
  - "HU-062"
  - "HU-063"
  - "HU-064"
  - "HU-065"
issue: 184
issue_url: "https://github.com/bcaceres19/barberia/issues/184"
suggested_issue_title: "chore(web): ejecutar el rediseño integral según el atlas NAVA Tailored Grid"
branch: null
pr: null
pr_url: null
depends_on:
  - "El atlas del issue #183 debe quedar integrado en `main` antes del primer cambio de producto"
  - "La orquestación anterior #168 está ejecutada y cerrada; este prompt exige una nueva migración visual material basada en el atlas ampliado"
rules:
  - "RN-TEN-01"
  - "RN-DAT-01"
  - "RN-DAT-02"
  - "RN-IDE-01"
  - "RN-CIT-02"
  - "RN-CIT-04"
  - "RN-CIT-05"
  - "RN-DIS-02"
  - "RN-DIS-04"
  - "RN-DIS-05"
  - "RN-DIS-06"
  - "RN-BLQ-01"
  - "RN-BLQ-02"
  - "RN-BLQ-03"
  - "RN-BLQ-04"
decisions:
  - "DEC-016"
  - "DEC-019"
  - "DEC-020"
  - "DEC-074"
  - "DEC-075"
  - "DEC-077"
  - "DEC-078"
  - "DEC-079"
acceptance_criteria:
  - "CA-ORCH-ATLAS-01: el router real queda inventariado y todas las rutas P0 implementadas tienen una entrega visual asignada."
  - "CA-ORCH-ATLAS-02: cada pantalla recibe un rediseño material guiado por su lámina, no una mera auditoría o sustitución mínima de tokens."
  - "CA-ORCH-ATLAS-03: el resultado conserva la firma cromática NAVA y traduce Tailored Grid a jerarquía, carriles, líneas, densidad y contraste operativo/lectura."
  - "CA-ORCH-ATLAS-04: no se introduce ninguna función tomada de un mockup cuando no exista en HU, contrato y código."
  - "CA-ORCH-ATLAS-05: cada preocupación se entrega con issue, prompt hijo cuando corresponda, rama y PR independientes."
  - "CA-ORCH-ATLAS-06: cada pantalla cubre carga, vacío, error, éxito, conflicto, disabled y offline cuando sean aplicables al flujo real."
  - "CA-ORCH-ATLAS-07: todas las entregas visibles se comprueban en 320, 360, 768 y 1280 px, zoom 200 %, teclado, foco y movimiento reducido."
  - "CA-ORCH-ATLAS-08: formato, lint, tipos, pruebas de componente, accesibilidad, E2E afectadas y build pasan antes de integrar."
  - "CA-ORCH-ATLAS-09: ninguna entrega modifica backend, OpenAPI, base de datos o reglas salvo un issue independiente autorizado."
  - "CA-ORCH-ATLAS-10: una revisión integral final compara la app real con las catorce láminas, registra diferencias justificadas y no deja pantallas a medio migrar."
  - "CA-ORCH-ATLAS-11: cada ruta se inspecciona en ejecución con Chrome DevTools usando exclusivamente la cuenta QA indicada por el propietario, con evidencia antes/después."
  - "CA-ORCH-ATLAS-12: el trabajo permanece netamente visual; cualquier defecto funcional ajeno se registra aparte y no amplía los PR de rediseño."
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/README.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/estados-citas.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/04-arquitectura/frontend.md"
  - "docs/10-backlog/plan-bloques.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/README.md"
  - "docs/10-backlog/evidence/ui-redesign-nava-2026-09-02/README.md"
  - "docs/10-backlog/prompts/README.md"
  - "apps/web/package.json"
  - "apps/web/src/app/router/index.ts"
  - "apps/web/src/modules/auth/routes.ts"
  - "apps/web/src/modules/agenda/routes.ts"
  - "apps/web/src/modules/catalog/routes.ts"
  - "apps/web/src/modules/staff/routes.ts"
  - "apps/web/src/modules/barberServices/routes.ts"
  - "apps/web/src/modules/schedules/routes.ts"
  - "apps/web/src/modules/settings/routes.ts"
created_at: "2026-09-02"
updated_at: "2026-09-03"
supersedes: "PROMPT-ORCH-NAVA-FRONTEND-v1.5"
superseded_by: "PROMPT-ORCH-NAVA-ATLAS-v3"
---

# Rediseño integral de NAVA guiado por el atlas Tailored Grid

## Instrucción principal para Claude Code

Ejecuta este prompt completo como orquestador, Senior Product Designer, UI/UX Designer e implementador frontend. No entregues únicamente un análisis, una propuesta, un inventario o un mensaje diciendo que la app ya cumple. El propietario solicita un **rediseño visual material de todas las pantallas P0 actualmente implementadas**, guiado por el atlas integral NAVA / Tailored Grid.

La ejecución anterior de NAVA y sus PR cerrados son únicamente la línea base técnica. **No constituyen evidencia de que este rediseño nuevo ya esté hecho.** Está prohibido cerrar una fase con cambios mínimos de tokens, una recaptura de screenshots o una conclusión de “ya cumple” si la interfaz continúa viéndose sustancialmente igual. Cada ruta debe recibir una mejora visible de composición, jerarquía, densidad, navegación y adaptación responsive, sin cambiar su comportamiento de negocio.

Trabaja con autonomía y persistencia hasta completar todas las fases autorizadas. Puedes crear los issues hijos, prompts persistentes, ramas, commits, pushes y pull requests necesarios; seguir sus checks; corregir fallos introducidos por la entrega; y continuar con la siguiente fase cuando la anterior esté integrada. Respeta las protecciones y revisiones obligatorias del repositorio. No hagas `force push`, no reescribas `main` y no mezcles varias preocupaciones en una rama o PR.

Solo detente por uno de estos motivos:

1. una contradicción normativa que cambie funcionalidad o seguridad;
2. falta de autorización o credenciales para una operación remota necesaria;
3. una aprobación humana obligatoria que no puedas obtener;
4. un fallo de infraestructura externo reproducido y documentado;
5. riesgo de sobrescribir cambios ajenos que no puedas aislar de forma segura.

Una elección rutinaria de CSS, composición, spacing, grid, breakpoint o componente no es un bloqueo: resuélvela aplicando el estándar y documenta la decisión en el issue correspondiente.

## Cuenta y entorno de prueba

- Usa **exclusivamente la cuenta QA que el propietario indique en la conversación de ejecución**. No elijas otra cuenta, no crees una cuenta alternativa y no uses datos de producción para sustituirla.
- Si la URL o la cuenta todavía no fueron proporcionadas, pide una sola vez los datos mínimos al iniciar la etapa de navegador. Puedes avanzar en inspección estática mientras llegan, pero no declares QA visual completado.
- Introduce las credenciales únicamente mediante la interfaz del navegador. No las escribas en archivos, prompts persistentes, comandos, variables versionadas, fixtures, logs, comentarios, screenshots ni reportes.
- No muestres la contraseña en evidencias. Evita capturar o publicar información personal; usa únicamente datos sintéticos dentro de la cuenta QA.
- Conserva la sesión autorizada mientras sea válida y no intentes saltarte autenticación, OTP, limitadores o permisos.

## Resultado obligatorio

Al finalizar deben existir:

- todas las rutas P0 implementadas rediseñadas como una familia coherente NAVA / Tailored Grid;
- componentes compartidos suficientes para evitar once soluciones aisladas, sin crear un megacomponente;
- navegación usable sin etiquetas truncadas ni destinos inaccesibles;
- formularios, tablas, carriles, diálogos, alertas y estados adaptados a escritorio, tablet y móvil;
- una evidencia visual comparable con cada mockup aplicable;
- capturas antes/después de cada ruta obtenidas desde la app real mediante Chrome DevTools;
- pruebas de interacción, accesibilidad y E2E actualizadas;
- un informe final que distinga fidelidad lograda, diferencias justificadas y conceptos no implementados;
- issues, prompts hijos, ramas, commits y PRs trazables.

El rediseño debe ser perceptible. No cierres una fase alegando que “ya usa los colores” si su composición, jerarquía, densidad, estados o navegación todavía no reflejan la lámina correspondiente.

## Preflight obligatorio

1. Lee completamente `AGENTS.md`, `CLAUDE.md`, `CONTRIBUTING.md` y todos los `source_docs`.
2. Comprueba rama, HEAD, remoto, árbol de trabajo y cambios ajenos. No uses `reset --hard`, `checkout --`, `clean` ni stashes destructivos.
3. La sesión que produjo este prompt puede estar en `docs/183-atlas-mockups-nava` con cambios documentales sin commit. Aísla y finaliza primero el issue #183 sin incluir cambios ajenos. No empieces código de producto hasta que el atlas y este prompt estén disponibles en `main` mediante PR.
4. Actualiza `main` mediante fast-forward después de integrar #183.
5. Verifica que el issue maestro #184 existe y está abierto.
6. Inspecciona el router real; no confíes solo en el inventario de este archivo.
7. Ejecuta Graphify cuando `graphify-out/graph.json` esté disponible y úsalo para localizar dependencias compartidas.
8. Revisa contradicciones y dudas pendientes. Una duda visual rutinaria se resuelve; una duda que cambia dominio, contrato, seguridad o alcance se registra y bloquea únicamente la fase afectada.
9. Ejecuta una línea base antes de editar: formato, lint, tipos, unitarias y build del frontend. Distingue fallos previos de regresiones propias.
10. Arranca el stack mediante los scripts oficiales del repositorio y abre la app real con el MCP de Chrome DevTools.
11. Inicia sesión por la interfaz con la cuenta QA indicada por el propietario; no inyectes una sesión ni evadas el flujo.
12. Captura con Chrome DevTools una línea base visual de las once rutas en 320, 360, 768 y 1280 px antes de editar. Si una ruta necesita datos, créalos mediante flujos autorizados con valores sintéticos.
13. Guarda un inventario `ruta → estado real → mockup → cambio visual obligatorio → evidencia antes/después` y úsalo como puerta de cada fase.

## Autoridad y precedencia

Aplica este orden:

1. `AGENTS.md`, seguridad, privacidad y reglas de Git;
2. reglas `RN-*`, decisiones `DEC-*`, HU y criterios `CA-*`;
3. OpenAPI, comportamiento real del backend y tipos generados;
4. estándar visual NAVA y especificación frontend;
5. atlas integral y mockups anteriores;
6. este prompt y sus prompts hijos.

Si una imagen contradice una fuente superior, conserva la fuente superior y registra la diferencia. Un mockup define dirección visual; nunca crea una API, una regla, un dato o una acción.

## Dirección visual vinculante

La solución final sintetiza todo lo aprobado:

- **Tinta NAVA `#101B2B`** para identidad, navegación, contexto operativo, tiempo y acciones primarias.
- **Marfil `#F4F0E7` y blanco `#FFFFFF`** para canvas, formularios, configuración y lectura prolongada.
- **Grafito `#2A2D32` / `#5E625F`** para texto principal y secundario.
- **Piedra `#E8E2D8` / `#C9C0B2`** para divisores, bordes y fondos atenuados.
- **Salvia `#748477`** para estados positivos y apoyo secundario.
- **Latón `#B8955A` / `#765C2F`** para selección, detalle editorial y foco destacado.
- Rojo únicamente para errores y acciones destructivas.

No conviertas toda la app en una superficie oscura ni toda la app en una superficie blanca. Usa la tinta para orientación y operación; usa marfil/blanco para contenido y formularios. Mantén voz serif editorial en marca/títulos y sans funcional en controles/datos.

Tailored Grid debe sentirse en:

- alineaciones precisas y divisores finos;
- organización por carriles, tiempo, secciones o columnas cuando el contenido lo justifique;
- densidad alta pero ordenada en escritorio;
- una acción principal inequívoca;
- reflow real a una columna en móvil;
- prioridad visual para el siguiente trabajo operativo;
- profundidad contenida, radios discretos y ausencia de tarjetas innecesarias.

Evita clichés decorativos de barbería, gradientes ornamentales, glassmorphism, neón, sombras profundas, pills para todo y la estética de dashboard genérico de IA.

## Libertad de implementación

La composición, el grid, las medidas, el spacing, los radios, las sombras, la navegación concreta y el inventario de componentes permanecen libres dentro del contrato anterior.

Puedes conservar CSS/CSS Modules existentes o proponer utility-first, Tailwind u otra biblioteca. **No instales Tailwind automáticamente.** Una dependencia visual nueva solo puede entrar en el issue de fundaciones si documentas:

- por qué mejora este proyecto concreto;
- licencia y mantenimiento;
- impacto medido en bundle y carga;
- integración con Vue y estilos actuales;
- accesibilidad y posibilidad de retirada;
- por qué CSS existente no es suficiente.

Si no existe esa justificación, continúa con las herramientas presentes.

## Atlas vinculante y mapa de rutas

Lee primero [`README del atlas`](../../evidence/ui-mockups-nava-tailored-grid-2026-09-02/README.md). Abre cada PNG a resolución completa antes de implementar su familia.

| Orden | Archivo | Aplicación autorizada |
| --- | --- | --- |
| 0 | `00-referencia-tailored-grid.png` | Carácter, composición, carriles y contraste escritorio/móvil. No copiar módulos fuera de alcance. |
| 1 | `14-componentes-formularios-alertas.png` | Fundaciones, controles, foco, alertas, diálogo y estados. |
| 2 | `01-agenda-diaria-responsive.png` | `/panel`. |
| 3 | `02-acceso-recuperacion.png` | `/acceso` y `/recuperar-acceso`. |
| 4 | `03-nuevo-turno.png` | `/panel/turnos/nuevo`. |
| 5 | `04-detalle-historial-reprogramacion.png` | `/panel/turnos/:appointmentId`. |
| 6 | `05-servicios-y-asignacion.png` | `/panel/servicios` y `/panel/servicios-por-barbero`. |
| 7 | `06-barberos-y-barberia.png` | `/panel/barberos` y `/panel/barberia`. |
| 8 | `07-horarios-y-excepciones.png` | `/panel/horarios`. |
| 9 | `08-bloqueos-puntuales-y-series.png` | `/panel/bloqueos`. |

Las siguientes imágenes son **conceptos P0 pendientes** y no autorizan código salvo que el preflight demuestre que la HU, el contrato, la base de datos y la ruta ya fueron aprobados e implementados después de este prompt:

- `09-reserva-publica-pasos-1-a-3.png`;
- `10-reserva-publica-pasos-4-a-6.png`;
- `11-gestion-publica-y-estados.png`;
- `12-ciclo-operativo-turno.png`;
- `13-configuracion-operativa-comunicaciones.png`.

Registra esas cinco familias como reserva visual, no como trabajo silencioso dentro de los PR del rediseño existente.

## Inventario funcional que debe conservarse

Antes de editar, reconfirma estos comportamientos contra código y HU:

### Shell privado

- Un único shell privado protegido por sesión.
- Wordmark `NAVA` y nombre real del tenant.
- Acción global “Nuevo turno”.
- En móvil, cuatro destinos legibles: `Agenda`, `Servicios`, `Barberos`, `Más`; `Más` debe hacer accesibles `Horarios`, `Bloqueos`, `Configuración`, `Servicios por barbero` y cerrar sesión.
- No muestres notificaciones, avatar “AD”, rol “Admin” ni métricas si el producto no los implementa.

### Acceso y recuperación

- Conserva exactamente pasos, validación, protección contra abuso, reenvío y mensajes uniformes del flujo real.
- No expongas si una cuenta existe ni destinos no autorizados.
- No inventes redes sociales, registro público ni login de clientes.

### Agenda diaria

- Un solo barbero seleccionado en P0 (`DEC-074`).
- Navegación por fecha y zona horaria real.
- Carril temporal y lista accesible equivalentes.
- Estados canónicos y turnos nocturnos según `DEC-075`.
- “Ahora” es un indicador temporal, nunca `in_progress`.

### Nuevo turno

- Campos y obligatoriedad derivados del flujo manual real.
- Servicio, barbero, persona atendida, fecha/hora y resumen previo.
- Conservación de datos ante error o conflicto.
- Prevención de doble envío y respuesta autoritativa del backend.

### Detalle, historial y reprogramación

- Hechos del turno, estado, historial paginado y diálogo de reprogramación existentes.
- No añadas completar, no-show, cancelar, editar servicio/duración o corregir estado terminal hasta que esas capacidades estén realmente implementadas.

### Barberos y barbería

- Barberos: lista, alta y edición/renombrado soportados.
- No inventes fotografía obligatoria, roles, desactivación o eliminación.
- Barbería: nombre, zona horaria y contactos según contrato real.

### Servicios y servicios por barbero

- Catálogo, precio COP, duración, alta/edición y ciclo de vida real.
- Asignaciones mediante selección de barbero y controles existentes.
- Rechazo visible al retirar la última asignación activa.
- No uses iconografía cliché como sustituto del nombre del servicio.

### Horarios, excepciones y bloqueos

- Jornada semanal, segmentos, días cerrados, festivos y excepciones reales.
- Bloqueos puntuales y series semanales según la interfaz existente.
- No dibujes edición de bloqueos ni impacto en turnos si esa acción no existe en el código/contrato.

## Estados transversales obligatorios

Cada fase debe inventariar y representar solo los estados aplicables:

- inicial;
- carga o skeleton sin layout shift severo;
- vacío con siguiente acción útil;
- error recuperable;
- error de validación junto al campo y resumen cuando convenga;
- conflicto `409` sin perder datos;
- guardado/enviado;
- deshabilitado;
- foco visible;
- operación pendiente que impide doble envío;
- offline cuando el flujo realmente lo detecte o el estándar lo requiera;
- contenido largo, datos ausentes y listas extensas.

No dependas únicamente de toasts ni del color. Alertas importantes deben persistir junto a la tarea.

## Frontera netamente visual

Este trabajo modifica exclusivamente la capa de presentación del frontend y sus verificaciones:

- composición, layout y responsive;
- CSS, estilos compartidos y primitivas visuales;
- semántica y accesibilidad del markup cuando no cambie la regla funcional;
- navegación visual y exposición de rutas ya existentes;
- estados visibles que el flujo ya produce;
- pruebas de componente, selectores E2E y evidencia afectada por el nuevo diseño;
- documentación estrictamente necesaria para trazabilidad visual.

No aproveches el rediseño para cambiar validaciones de negocio, payloads, endpoints, permisos, persistencia, concurrencia, textos normativos, ciclo de vida o capacidades. Si Chrome DevTools descubre un bug funcional no causado por el rediseño, crea un issue separado con evidencia y continúa con las pantallas que no dependan de él; no lo soluciones dentro del PR visual.

## Verificación obligatoria con Chrome DevTools

Playwright, Vitest y screenshots automatizados complementan la revisión, pero **no sustituyen la inspección manual de la app real con Chrome DevTools**. En cada fase y en la revisión final:

1. Abre la ruta real en Chrome DevTools con la cuenta QA indicada.
2. Recorre el flujo con clic, teclado y navegación real; no te limites a cargar la URL.
3. Usa Device Mode para 320, 360, 768 y 1280 px y captura el viewport completo relevante.
4. Comprueba a 200 % de zoom que el contenido refluya y siga operable.
5. Inspecciona el DOM, accesible name, jerarquía de headings, labels, focus order y restauración de foco en diálogos.
6. Verifica en estilos computados los colores, fuentes, contraste, overflow y dimensiones táctiles realmente aplicados.
7. Revisa consola: cero errores nuevos, warnings relevantes explicados y ninguna exposición de datos.
8. Revisa Network: ninguna solicitud inesperada, ningún 4xx/5xx provocado por el rediseño y estados de carga/error visibles cuando correspondan.
9. Fuerza o prepara de forma segura estados de carga, vacío, error, conflicto y éxito disponibles en el flujo real; no los simules con HTML editado como evidencia final.
10. Compara lado a lado el resultado con el PNG asignado y captura diferencias justificadas.
11. Guarda evidencia por ruta y ancho con nombres deterministas, sin credenciales ni datos personales.

Una fase no puede marcarse terminada si solo se revisó código, Storybook, jsdom, Playwright headless o una imagen estática.

## Plan de ejecución obligatorio

### Fase 0 · Integrar el atlas y crear el backlog

1. Finaliza el issue #183 en su rama sin mezclar los cambios ajenos preexistentes.
2. Incluye el atlas, estándar 4.1, historial, este prompt y su índice en el PR documental correspondiente.
3. Espera checks y deja #183 integrado según el flujo del repositorio.
4. Actualiza `main` por fast-forward.
5. Audita router, componentes, estilos, pruebas y evidencia actual.
6. Crea bajo #184 los issues hijos necesarios con criterios observables y enlaces al mockup exacto.
7. Crea un prompt hijo autocontenido para cada issue no trivial y actualiza el catálogo.

El backlog mínimo separa estas preocupaciones:

1. fundaciones visuales compartidas;
2. shell y navegación;
3. acceso y recuperación;
4. agenda diaria;
5. nuevo turno;
6. detalle, historial y reprogramación;
7. barberos;
8. configuración de barbería;
9. servicios y ciclo de vida;
10. servicios por barbero;
11. horarios y excepciones;
12. bloqueos;
13. revisión integral.

Puedes fusionar únicamente dos rutas en un issue cuando compartan la misma preocupación, componentes y pruebas, y el PR siga siendo pequeño y revisable. Documenta la razón. No uses una sola rama para todo.

### Fase 1 · Fundaciones

- Audita `src/styles` y `src/shared/ui` antes de añadir primitivas.
- Establece variables/estilos para la firma cromática, tipografía, bordes, capas, foco y motion.
- Evoluciona botones, inputs, selects, alertas, badges, diálogos y loading reutilizados.
- Mantén componentes independientes del contrato interno del API.
- Añade páginas de demostración solo si son de desarrollo y no crean ruta productiva.

### Fase 2 · Shell y navegación

- Traduce el contraste oscuro/claro del atlas al shell real.
- Corrige navegación móvil sin truncamiento y con safe areas.
- Mantén activo, hover, focus y current route distinguibles.
- Asegura foco al navegar, skip link y encabezados correctos.
- No añadas destinos fuera del router.

### Fases 3–12 · Familias de pantalla

Para cada familia:

1. abre el mockup correspondiente y captura la pantalla real actual;
2. enumera similitudes, diferencias y restricciones funcionales;
3. define en el issue qué cambios harán el rediseño material;
4. implementa la pantalla completa, incluidos diálogos y estados;
5. actualiza pruebas de componente sin debilitar aserciones;
6. actualiza E2E para comportamiento y evidencia visual;
7. verifica responsive y accesibilidad antes del commit;
8. recorre la pantalla terminada en Chrome DevTools usando la cuenta QA indicada;
9. abre el PR con capturas reales antes/después y explicación de diferencias justificadas;
10. espera checks, corrige regresiones propias e integra según las reglas;
11. actualiza `main` antes de iniciar la siguiente rama.

### Fase 13 · Revisión integral

- Recorre las once rutas reales con datos sintéticos y al menos dos tenants cuando el entorno de prueba lo requiera.
- Usa la cuenta QA indicada por el propietario para todo el recorrido visible en Chrome DevTools.
- Compara cada pantalla con su mockup y el estándar, no solo con snapshots.
- Comprueba consistencia de shell, tipografía, controles, alertas, diálogos y navegación.
- Verifica que no haya scroll horizontal accidental, solapamientos, truncamiento, pérdida de foco ni contenido inaccesible.
- Ejecuta `axe-core` en vivo y pruebas automatizadas aplicables.
- Mide build/bundle y documenta cualquier dependencia visual agregada.
- Crea issues separados para defectos fuera del alcance; corrige dentro de esta fase únicamente regresiones del rediseño.
- Cierra #184 solo cuando todas las entregas hijas estén integradas o exista un bloqueo externo explícito aceptado por el propietario.

## Responsive y accesibilidad

Verifica como mínimo:

| Ancho/condición | Evidencia requerida |
| --- | --- |
| 320 px | Sin truncamiento, scroll horizontal accidental ni acciones inaccesibles. |
| 360 px | Flujo táctil completo y navegación legible. |
| 768 px | Transición tablet coherente, sin zonas vacías o densidad rota. |
| 1280 px | Jerarquía Tailored Grid y uso eficiente del espacio. |
| Zoom 200 % | Reflow, lectura y operación completas. |
| Teclado | Orden lógico, foco visible, Escape y restauración de foco. |
| `prefers-reduced-motion` | Ninguna información depende del movimiento. |

Objetivo de controles táctiles: 44 × 44 px cuando la composición lo permita. Cumple WCAG 2.2 AA en las combinaciones implementadas. No confíes en los colores aproximados impresos dentro del PNG: mide el CSS real.

## Pruebas por entrega

Cada PR debe incluir la capa más baja adecuada:

- pruebas unitarias de funciones visuales o de estado cuando exista lógica;
- pruebas de componente para interacción, semántica, foco y estados;
- `vitest-axe` con el estado relevante visible;
- E2E del recorrido P0 afectado;
- evidencia en los cuatro anchos;
- recorrido manual obligatorio con Chrome DevTools y la cuenta QA indicada;
- revisión de consola y red durante el recorrido real;
- build y presupuesto razonable de bundle.

No actualices snapshots a ciegas, no ignores pruebas inestables y no reduzcas umbrales.

## Verificación técnica mínima

Desde la raíz, usa los comandos equivalentes disponibles en el repositorio:

```text
pnpm --filter @system-barbershop/web format
pnpm --filter @system-barbershop/web lint
pnpm --filter @system-barbershop/web typecheck
pnpm --filter @system-barbershop/web test:unit
pnpm --filter @system-barbershop/web build
pnpm --filter @system-barbershop/web test:e2e
git diff --check
```

Si el workspace usa otra forma de filtro o scripts raíz, inspecciónalos y usa el comando real; no declares una prueba ejecutada si el comando no corresponde.

## Restricciones estrictas

- No modificar backend, OpenAPI, migraciones ni base de datos como efecto lateral visual.
- No incluir arreglos funcionales no visuales aunque se descubran durante la inspección; registrarlos aparte.
- No usar `any`, estado global o una dependencia nueva sin justificación.
- No introducir datos personales reales, secretos ni tokens en código, fixtures, screenshots o logs.
- No crear clientes/CRM, ventas, caja, métricas, reportes, inventario, pagos, marketplace o cuenta del cliente.
- No crear agenda multi-barbero en P0.
- No crear estado `in_progress`.
- No hacer fotos obligatorias de barberos.
- No copiar controles “Admin”, notificaciones o módulos decorativos de la imagen maestra.
- No modificar un mockup para esconder una discrepancia de implementación; documenta la diferencia.
- No conservar dos sistemas visuales completos en la misma pantalla al cerrar su fase.

## Git, issues y PR

- Un issue y una rama corta `<tipo>/<issue>-<descripcion>` por preocupación.
- Conventional Commits y título de PR.
- Cada PR enlaza #184 y cierra únicamente su issue hijo.
- No mezclar archivos ajenos ni cambios de otras sesiones.
- No push directo ni force push a `main`.
- Integración por squash después de checks y revisiones obligatorias.
- Elimina la rama después del merge cuando sea seguro.
- Actualiza prompt hijo, catálogo, matriz e historial con estados reales.

## Formato de progreso

Después de cada fase informa de forma compacta:

| Campo | Contenido |
| --- | --- |
| Fase | Nombre y familia visual |
| Issue / rama / PR | Enlaces y nombres reales |
| Mockup | Archivo usado |
| Cambios | Resultado visual observable |
| Diferencias | Desviaciones justificadas por reglas, contenido o accesibilidad |
| Pruebas | Comandos y resultado real |
| Evidencia | Chrome DevTools antes/después, 320/360/768/1280, zoom 200 %, consola, red y accesibilidad |
| Estado | Integrada, en revisión o bloqueada |
| Siguiente | Próxima fase desbloqueada |

No pidas confirmación entre fases cuando la siguiente acción ya está cubierta por este prompt y las reglas del repositorio.

## Entrega final

Entrega:

1. tabla de las once rutas y su estado final;
2. issues, ramas y PRs creados;
3. capturas antes/después por familia;
4. resultado de formato, lint, tipos, unitarias, E2E, build y `git diff --check`;
5. resumen de accesibilidad y responsive;
6. matriz de inspección Chrome DevTools por ruta, indicando cuenta QA utilizada de forma enmascarada, consola, red, teclado y estados recorridos;
7. dependencias añadidas y justificación, o confirmación de que no se añadió ninguna;
8. diferencias justificadas respecto de los mockups;
9. conceptos P0 pendientes que permanecieron sin implementar;
10. bugs funcionales descubiertos y registrados aparte, sin mezclarlos con el rediseño;
11. riesgos o bloqueos externos restantes;
12. confirmación explícita de que no se modificaron contratos, datos o reglas de negocio.

No declares “rediseño completo” mientras falte una ruta, un estado crítico, evidencia responsive o una fase pendiente sin bloqueo aceptado.

## Texto corto para iniciar la ejecución

```text
Lee completamente y ejecuta de principio a fin `docs/10-backlog/prompts/orchestration/redisenio-integral-nava-tailored-grid-v2.md`. Esta ejecución es netamente de rediseño visual: no te limites a auditar, recapturar pantallas, cambiar tokens o decir que ya cumple. Rediseña materialmente las once rutas P0 según el atlas, usa exclusivamente la cuenta QA que te indicaré en esta conversación y verifica cada pantalla real con Chrome DevTools en todos los anchos exigidos. Crea y gestiona los issues, prompts hijos, ramas, commits y PRs necesarios; continúa sin pedirme confirmaciones rutinarias y detente únicamente ante un bloqueo real definido por el propio prompt.
```
