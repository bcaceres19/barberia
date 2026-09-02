---
prompt_id: "PROMPT-ORCH-NAVA-FRONTEND-v1"
version: "1.4"
kind: "orchestration"
status: "draft"
target_agents:
  - "claude"
  - "codex"
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
issue: "pending"
issue_url: null
suggested_issue_title: "chore(web): orquestar el rediseño incremental de todas las pantallas NAVA"
branch: null
pr: null
pr_url: null
depends_on:
  - "DEC-077, DEC-078 y DEC-079 registradas; estándar visual y mockups NAVA disponibles"
  - "Un issue real por cada fundación o pantalla que se vaya a migrar"
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
  - "CA-ORCH-NAVA-01: existe un inventario comprobado de todas las rutas y pantallas P0 implementadas, con su mockup, HU, contrato, pruebas y issue hijo."
  - "CA-ORCH-NAVA-02: cada familia visual se entrega en una rama y PR independientes, sin mezclar dominio, API o capacidades no autorizadas."
  - "CA-ORCH-NAVA-03: toda pantalla rediseñada respeta la firma cromática, el contraste serif/sans y la familia de los mockups de DEC-079."
  - "CA-ORCH-NAVA-04: cada entrega cubre interacción, estados, 320/360/768/1280 px, zoom 200 %, teclado, foco, contraste y movimiento reducido."
  - "CA-ORCH-NAVA-05: las pruebas de componente, vitest-axe, E2E afectadas, lint, tipos y build pasan o el bloqueo real queda documentado sin declarar cumplimiento."
  - "CA-ORCH-NAVA-06: al finalizar, todas las rutas P0 existentes del inventario están rediseñadas o tienen un bloqueo explícito y trazable."
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/README.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/estados-citas.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/04-arquitectura/frontend.md"
  - "docs/10-backlog/plan-bloques.md"
  - "docs/10-backlog/evidence/ui-redesign-nava-2026-09-02/README.md"
  - "docs/10-backlog/prompts/README.md"
  - "apps/web/src/app/router/index.ts"
  - "apps/web/src/modules/auth/routes.ts"
  - "apps/web/src/modules/agenda/routes.ts"
  - "apps/web/src/modules/catalog/routes.ts"
  - "apps/web/src/modules/staff/routes.ts"
  - "apps/web/src/modules/barberServices/routes.ts"
  - "apps/web/src/modules/schedules/routes.ts"
  - "apps/web/src/modules/settings/routes.ts"
created_at: "2026-09-01"
updated_at: "2026-09-02"
supersedes: null
superseded_by: null
---

# Orquestación de la adopción incremental del frontend NAVA

## Instrucción para Claude o el agente ejecutor

Usa este archivo como mapa de coordinación, no como autorización para rediseñar todo el frontend en una sola rama. La identidad NAVA, la firma cromática, la familia de los mockups aprobados y la frontera funcional son normativas; la composición, las medidas, los componentes y la tecnología visual son libres. Cada fundación o pantalla debe llegar mediante un issue real y, cuando el cambio sea no trivial, un prompt hijo autocontenido que atienda una sola preocupación primaria.

Este prompt permanece en `draft` mientras `issue: pending`. No cambies código desde este prompt ni lo marques `ready` hasta registrar un issue real de orquestación o planificación. La existencia de una pantalla en la especificación tampoco autoriza una capacidad P0 pendiente, P1, P2 o excluida.

Cuando el prompt tenga un issue real y pase a `ready`, ejecútalo como orquestación persistente: prepara y entrega cada familia mediante un prompt hijo, issue, rama y PR propios; espera que la dependencia anterior esté integrada antes de crear la siguiente rama desde `main` actualizado. No uses una rama de orquestación para acumular el rediseño. Continúa hasta cubrir todo el inventario P0 existente o hasta identificar un bloqueo normativo o externo real.

Si faltan números de issue para las entregas hijas, produce primero el backlog propuesto con criterios verificables y detente antes de modificar código. No inventes números. Solo crea issues o realiza acciones remotas si la persona que ejecuta este prompt lo autorizó expresamente.

## Objetivo

Coordinar una evolución gradual y verificable desde la interfaz actual hacia **NAVA / Tailored Grid**, de modo que cada rediseño o pantalla nueva respete los mockups y la firma cromática aprobados, pueda resolver libremente su composición y revise de forma completa estados, responsive, accesibilidad y pruebas, sin alterar contratos, reglas de negocio, seguridad, auditoría ni alcance.

El resultado de esta orquestación no es “todo NAVA en un PR”. Es un backlog trazable de entregas pequeñas, cada una con issue, dependencias, criterio observable, rama, pruebas y evidencia.

## Preflight obligatorio

1. Comprueba que el árbol esté limpio y no sobrescribas cambios ajenos.
2. Actualiza `main` mediante fast-forward antes de preparar cualquier entrega hija.
3. Ejecuta la consulta Graphify apropiada cuando exista `graphify-out/graph.json`.
4. Lee completamente `AGENTS.md`, `CLAUDE.md` y cada `source_docs` de este prompt.
5. Comprueba dudas y contradicciones. Si falta una decisión funcional, registra el bloqueo; no la completes desde el diseño.
6. Comprueba la clasificación de la pantalla en `especificacion-frontend-nava.md`: `P0 existente`, `P0 pendiente`, `P1`, `P2` o `Excluido`.
7. Para una pantalla existente, inventaría rutas, guardas, contratos, estados, permisos, copy y pruebas antes de editar.
8. Localiza o crea con autorización el issue real de la entrega hija. Actualiza su prompt antes de ejecutar.
9. Crea desde `main` actualizada la rama corta `<tipo>/<issue>-<descripcion>` indicada por el estándar Git.

## Alcance incluido

- Descomponer la adopción NAVA en issues y prompts hijos independientes.
- Mantener el orden de dependencias entre fundaciones, shell y pantallas.
- Hacer que las pantallas nuevas P0 aprobadas nazcan con NAVA desde su primer issue.
- Migrar pantallas existentes sin modificar su dominio o API, salvo que el mismo issue lo autorice y trace expresamente.
- Verificar la intención visual elegida, responsive, accesibilidad, estados asíncronos, rendimiento y pruebas de cada entrega.
- Actualizar trazabilidad documental y el catálogo de prompts a medida que cada entrega cambia de estado.

## Fuera de alcance

- Un rediseño masivo de `apps/web` en una única rama o PR.
- Implementar B3, B4, B5 o B6 solo porque su pantalla está especificada.
- Crear ingresos, reportes, inventario, caja, pagos, comisiones o perfiles de cliente.
- Crear el estado `in_progress`; “Ahora” es únicamente una señal temporal.
- Crear una vista consolidada multi-barbero mientras `DEC-074` siga gobernando P0.
- Añadir settings de tema, temas por tenant, app nativa o capacidades fuera del alcance P0.
- Incorporar una dependencia o asset sin licencia, revisión de accesibilidad y medición de rendimiento.
- Modificar OpenAPI, persistencia o reglas como efecto lateral de un cambio visual.

## Estado existente que debe conservarse

- `apps/web` ya tiene Vue 3, TypeScript estricto, Vue Router, cliente tipado y módulos por capacidad.
- Existen primitivas compartidas como botones, inputs, diálogos, alertas e insignias; deben auditarse y evolucionarse, no duplicarse por pantalla.
- B0, B1, B2 y `HU-060`–`HU-065` tienen entregas integradas con distintos seguimientos E2E/responsive pendientes registrados en la matriz.
- Los estilos actuales pueden convivir temporalmente con una nueva composición hasta que el issue de la pantalla defina y verifique su migración. La documentación de NAVA no significa que todo el CSS ya esté rediseñado.
- La agenda P0 exige un solo barbero seleccionado (`DEC-074`) y los turnos nocturnos aparecen en cada día intersectado (`DEC-075`).
- Los estados P0 son `confirmed`, `completed`, `cancelled_by_customer`, `cancelled_by_barber` y `no_show`.
- La interfaz usa “turno”; `appointment` permanece en API, datos y código (`DEC-016`).

## Inventario vinculante de pantallas actuales

“Todas las pantallas” significa todas las rutas P0 ya implementadas y las capas visibles que forman parte de ellas. Antes de crear issues, vuelve a comprobar el router: si apareció otra ruta P0 existente desde la versión de este prompt, agrégala al inventario con su HU y contrato; si es P0 pendiente, P1, P2 o excluida, no la implementes.

| Orden | Ruta o superficie | Implementación actual | Alcance visual que debe quedar completo | Referencia principal |
| --- | --- | --- | --- | --- |
| 0 | Fundaciones compartidas | `src/styles/{tokens.css,base.css}` y `src/shared/ui/{NavaWordmark,BaseButton,BaseInput,BaseAlert,BaseBadge,BaseDialog}.vue` | Firma cromática, tipografía, foco, controles, alertas, badges, capas y estados reutilizados realmente por las pantallas. | `03-componentes-formularios-alertas.png` |
| 1 | Shell privado | `src/modules/auth/layouts/PrivateShell.vue`, `AppHeader.vue`, `AppNav.vue` | Contexto NAVA/tenant, acción “Nuevo turno”, navegación sin truncamiento, safe areas, foco al navegar, escritorio y móvil. | `04-servicios-configuracion-navegacion.png` |
| 2 | `/acceso` | `src/modules/auth/pages/LoginPage.vue`, `LoginForm.vue`, `PhoneChallengeForm.vue` | Acceso inicial, validación, envío/carga, error, desafío por abuso y sesión vencida; sin navegación privada. | `01-autenticacion-movil.png` |
| 3 | `/recuperar-acceso` | `RecoveryPage.vue`, `RecoveryRequestStep.vue`, `RecoveryVerifyStep.vue`, `RecoveryResetStep.vue` | Solicitud, verificación, nueva contraseña, reenvío, errores uniformes y confirmación persistente en los tres pasos. | `01-autenticacion-movil.png` |
| 4 | `/panel` | `src/modules/agenda/pages/DailyAgendaPage.vue` | Selector obligatorio de barbero, fecha, carga/vacío/error, lista y representación temporal equivalente, estados canónicos y turnos nocturnos. | Derivar de `02` y `04`, más §7.2 de la especificación NAVA. |
| 5 | `/panel/turnos/nuevo` | `NewAppointmentPage.vue` | Formulario completo, agrupación semántica, resumen previo, validación, conflicto, doble envío y reflow móvil. | `02-nuevo-turno-responsive.png` |
| 6 | `/panel/turnos/:appointmentId` | `AppointmentDetailPage.vue` | Detalle, historial paginado, carga/vacío/error, estado terminal y diálogo o formulario de reprogramación sin perder contexto. | Derivar de `02` y `03`, más §§7.3–7.5 de la especificación. |
| 7 | `/panel/barberia` | `src/modules/settings/pages/SettingsPage.vue` | Formulario de barbería, ayudas, validación, guardado/carga/error y éxito persistente. | `04-servicios-configuracion-navegacion.png` |
| 8 | `/panel/barberos` | `src/modules/staff/pages/StaffPage.vue` | Lista, vacío, alta, edición/renombrado, errores y acciones permitidas; sin inventar foto o desactivación. | Derivar de `03` y `04`, más §7.6 de la especificación. |
| 9 | `/panel/servicios` | `src/modules/catalog/pages/CatalogPage.vue` | Catálogo, alta/edición, estados activo/inactivo, confirmación de desactivación, impacto real y reactivación. | `03-componentes-formularios-alertas.png` y `04-servicios-configuracion-navegacion.png` |
| 10 | `/panel/servicios-por-barbero` | `src/modules/barberServices/pages/BarberServicesPage.vue` | Selector de barbero, asignaciones, carga/vacío/error/conflicto y rechazo al retirar la última asignación activa. | Derivar de `03` y `04`, más §7.7 de la especificación. |
| 11 | `/panel/horarios` | `src/modules/schedules/pages/SchedulesPage.vue` | Jornada semanal, segmentos, días cerrados, excepciones/festivos y formularios soportados, con tablas transformadas en móvil. | Derivar de `02`–`04`, más §7.8 de la especificación. |
| 12 | `/panel/bloqueos` | `src/modules/schedules/pages/BlocksPage.vue` | Lista, alta, edición de ocurrencia/serie solo donde exista, recurrencia, estados y conflictos, sin simular turnos afectados pendientes. | Derivar de `02`–`04`, más §7.8 de la especificación. |

La carpeta canónica de mockups es `docs/10-backlog/evidence/ui-redesign-nava-2026-09-02/`. Las capturas anteriores de `docs/10-backlog/evidence/ui-visual-2026-09-02/` sirven como línea base del sistema actual, no como destino visual.

## Contrato visual que Claude debe aplicar

### Firma cromática

- Tinta NAVA `#101B2B`: wordmark, acciones primarias, shell y alto contraste.
- Marfil `#F4F0E7`: canvas y fondo dominante.
- Blanco `#FFFFFF`: campos, diálogos y superficies de lectura.
- Grafito `#2A2D32` y `#5E625F`: texto principal y secundario.
- Piedra `#E8E2D8` y `#C9C0B2`: superficies atenuadas, divisores y bordes.
- Salvia `#748477`: acento secundario y apoyo positivo.
- Latón `#B8955A` y `#765C2F`: selección, detalle editorial y foco visible.
- Hover/activo de tinta: `#18283D` y `#0A1420`.
- Estados: éxito `#EAF0EB/#325D43/#748477`; advertencia/conflicto `#F8F1DF/#775019/#9A6A24`; peligro `#F8EDEC/#8A2C2C/#A43A3A`; información `#E9EEF3/#23405B/#667D93`; inactivo `#EEECE8/#56514A/#C9C0B2`.

No es obligatorio usar todos los colores en cada pantalla ni conservar estos nombres como tokens. Sí es obligatorio que la paleta NAVA domine y que un color adicional tenga necesidad semántica o de legibilidad, contraste medido y justificación; no puede crear otra identidad de marca. El color nunca es la única señal.

### Tipografía y carácter

- Conserva contraste entre voz editorial serif para NAVA/display y sans-serif funcional para controles, datos, ayuda y estados.
- Prefiere `Instrument Serif`/`Instrument Sans` solo si los assets, licencia, self-hosting y rendimiento están resueltos; de lo contrario usa los fallbacks aprobados sin cargar fuentes remotas desde componentes.
- Usa jerarquía editorial, alineación cuidada, líneas finas, profundidad contenida y bajo ruido visual.
- Evita tijeras, navajas, bigotes y postes como decoración; no uses neón, estética gaming, vidrio ornamental, sombras profundas ni tarjetas grandes redondeadas para cada fragmento.
- Mantén el wordmark `NAVA`; no inventes un logotipo de la barbería ni reemplaces el nombre real del tenant.

### Componentes y estados

- Primaria: tinta NAVA, etiqueta verbo + resultado, carga explícita y prevención de doble envío.
- Secundaria: menor peso; destructiva: tratamiento de peligro y consecuencia nombrada.
- Campos: label persistente; placeholder no sustituye label; orden `label → control → ayuda/error`; requerido/opcional explícito; foco visible.
- Formularios largos: agrupación semántica por secciones o pasos; escritorio puede usar columnas; móvil refluye sin miniaturizar ni perder datos.
- Alertas: superficie, borde, título, mensaje accionable e icono cuando ayude; éxito importante y conflicto son persistentes, no solo toast.
- Diálogos: título, cierre accesible, foco contenido y restaurado, Escape cuando sea seguro; bottom sheet móvil solo para decisiones breves.
- Navegación: ninguna etiqueta truncada. La solución `Agenda / Servicios / Barberos / Más` con hoja para `Horarios`, `Configuración` y `Servicios por barbero` está aprobada como posibilidad, no como única composición.

### Libertad permitida

Puedes decidir por pantalla composición, grid, tamaños, espaciado, radios, sombras, movimiento, navegación, componentes locales o compartidos y tecnología de estilos. Puedes conservar CSS, adoptar CSS Modules, `<style scoped>`, Tailwind, utility-first o una biblioteca si el issue justifica licencia, mantenimiento, accesibilidad y bundle. No instales Tailwind ni otra dependencia automáticamente.

La libertad no permite cambiar contratos, copy normativo, obligatoriedad, permisos, estados de dominio, rutas o funciones para acomodar el diseño. Tampoco permite copiar como datos reales los ejemplos sintéticos de los mockups.

## Plan de entregas obligatorio

### Fase 0 · Inventario y backlog, sin cambios de código

1. Confirma el inventario anterior contra las rutas, páginas, componentes y pruebas actuales.
2. Para cada fila registra HU, CA, RN, DEC, requests/responses, estados, permisos, pruebas y deuda visual conocida.
3. Captura una línea base solo con datos sintéticos en 360 y 1280 px.
4. Propón los issues hijos indicados abajo con criterios verificables. Si no están autorizados o no tienen número real, detente y entrega el backlog; no abras ramas ni modifiques código.
5. Crea o actualiza un prompt hijo autocontenido por issue y registra ambos en el catálogo.

### Fase 1 · Fundaciones compartidas, solo si el inventario demuestra que hacen falta

**Issue sugerido:** `chore(web): alinear primitivas compartidas con los mockups NAVA`.

- Reutiliza la paleta ya existente cuando coincide con el estándar; no reescribas tokens por estética si no hace falta.
- Adapta únicamente primitivas usadas por entregas inmediatas: botón, campo, alerta, badge, diálogo y wordmark.
- Conserva props y contratos públicos o migra consumidores en el mismo issue cuando sea imprescindible.
- Añade pruebas de variantes, foco, carga, disabled, descripción accesible y diálogo.
- No rediseñes páginas de negocio dentro de esta rama.

### Fase 2 · Shell y navegación privada

**Issue sugerido:** `feat(web): rediseñar shell y navegación responsive NAVA`.

- Cubre `PrivateShell`, `AppHeader` y `AppNav` como una preocupación transversal.
- Preserva el guard, la rehidratación de sesión, rutas, tenant activo y cierre de sesión.
- Evita las seis etiquetas comprimidas observadas. Implementa una navegación que no trunque; `Más` es una posibilidad aprobada.
- Verifica safe areas, teclado, foco al cambiar ruta, zoom 200 % y contenido que no queda oculto por una barra fija.

### Fase 3 · Autenticación y recuperación

**Issue sugerido:** `feat(web): rediseñar acceso y recuperación NAVA`.

- Cubre `/acceso`, desafío por abuso y los tres pasos de `/recuperar-acceso` como una única familia de seguridad visual.
- Mantén mensajes neutrales, límites, destino enmascarado, reglas de contraseña, reenvío y respuestas actuales.
- No resuelvas dentro del rediseño `DP-NOT-06`, `DP-SEG-13`, `CT-009` o `CT-010`; si siguen abiertos, conserva el contrato vigente.
- Verifica error inline, resumen cuando corresponda, sesión vencida, carga, disabled, pegar código, mostrar/ocultar contraseña y regreso seguro.

### Fase 4 · Agenda diaria

**Issue sugerido:** `feat(web): rediseñar agenda diaria NAVA`.

- Cubre solo `/panel` y sus estados de lectura.
- Mantén selector obligatorio de un barbero, navegación de fecha y regla de intersección para turnos nocturnos.
- En escritorio puede existir una representación temporal; siempre conserva alternativa de lista accesible.
- No construyas vista consolidada multi-barbero ni acciones de ciclo de vida pendientes.

### Fase 5 · Nuevo turno

**Issue sugerido:** `feat(web): rediseñar formulario de nuevo turno NAVA`.

- Cubre `/panel/turnos/nuevo` contra el mockup 02.
- Mantén barbero antes de servicio, campos y opcionalidad reales, zona horaria, resumen previo, idempotencia y conflictos `409`.
- Conserva selecciones válidas después de error o conflicto; no inventes precio, duración o recordatorios si el contrato no los aporta.

### Fase 6 · Detalle, historial y reprogramación

**Issue sugerido:** `feat(web): rediseñar detalle e historial del turno NAVA`.

- Cubre `/panel/turnos/:appointmentId`, paginación de historial y la transición T2 ya implementada.
- Mantén el vocabulario “turno”, snapshots históricos, permisos, conflicto por cruce/bloqueo y versión vigente.
- No agregues edición T3, cancelación, completar, no asistencia ni corrección terminal hasta sus HU/issues.

### Fase 7 · Configuración de barbería

**Issue sugerido:** `feat(web): rediseñar configuración de barbería NAVA`.

- Cubre `/panel/barberia` con el mockup 04.
- Mantén nombre, zona horaria y contactos autorizados; no agregues paleta, tema, logo, canales o políticas sin HU.
- El éxito es persistente y el error conserva los datos no sensibles escritos.

### Fase 8 · Barberos

**Issue sugerido:** `feat(web): rediseñar gestión de barberos NAVA`.

- Cubre `/panel/barberos`: lista, vacío, alta y renombrado.
- No inventes foto, orden, desactivación, eliminación o vínculo automático con `staff_user`.
- Los formularios cortos pueden usar diálogo o bottom sheet si mantienen foco y legibilidad.

### Fase 9 · Servicios y ciclo de vida

**Issue sugerido:** `feat(web): rediseñar catálogo de servicios NAVA`.

- Cubre `/panel/servicios` con mockups 03 y 04.
- Mantén nombre único activo, duración planificada, precio COP mayor que cero, estado, impacto real, desactivación y reactivación.
- La acción destructiva muestra consecuencia y recuento reales; no calcula impacto en Vue ni promete cambiar turnos existentes.

### Fase 10 · Servicios por barbero

**Issue sugerido:** `feat(web): rediseñar asignación de servicios por barbero NAVA`.

- Cubre `/panel/servicios-por-barbero` sin cambiar su contrato.
- Mantén selector explícito, carga/vacío/error, asignación idempotente y rechazo de la última asignación activa.
- Si el producto decide integrarla visualmente dentro de Servicios, conserva la ruta compatible y trata la fusión como parte explícita del issue; no la elimines por el mockup.

### Fase 11 · Horarios y excepciones

**Issue sugerido:** `feat(web): rediseñar horarios y excepciones NAVA`.

- Cubre `/panel/horarios`: jornada semanal, segmentos, días cerrados y excepciones/festivos ya soportados.
- Mantén fechas civiles, zona horaria, intervalos semiabiertos y cruces de medianoche.
- Transforma tablas/grids para móvil; no uses scroll horizontal como único mecanismo.

### Fase 12 · Bloqueos

**Issue sugerido:** `feat(web): rediseñar bloqueos de agenda NAVA`.

- Cubre `/panel/bloqueos`: lista, creación, edición y alcance ocurrencia/serie solo donde exista comportamiento real.
- Mantén tipos, recurrencia, borrado lógico, idempotencia y bloqueo duro frente a nuevos turnos/reprogramación.
- No implementes todavía el flujo P0 pendiente de decisiones sobre turnos afectados.

### Fase 13 · Revisión integral

**Issue sugerido:** `test(web): validar rediseño integral NAVA en rutas P0 existentes`.

- Se ejecuta solo después de integrar las fases aplicables.
- Recorre todas las rutas del inventario, estados globales, navegación cruzada y retorno de foco.
- Compara visualmente con los mockups como familia, no píxel a píxel.
- Registra cualquier defecto nuevo en un issue propio; no mezcles correcciones no triviales en la rama de QA.

### Trabajo futuro no autorizado

Reserva pública, gestión pública del turno, rutas de sistema P0 pendientes y capacidades P1/P2 se rediseñan cuando tengan HU, contrato e issue reales. No prepares su código anticipadamente.

## Protocolo de ejecución por entrega

Claude Code debe repetir este ciclo para cada fase autorizada, sin acumular varias fases en una rama:

1. Parte de `main` actualizada por fast-forward y comprueba que el árbol de trabajo esté limpio. Si hay cambios ajenos, no los descartes, no los reformatees y no los incluyas.
2. Lee la página, sus componentes, validaciones, cliente tipado, pruebas y contrato antes de proponer cambios. Ejecuta la pantalla actual cuando el entorno lo permita y conserva una línea base con datos sintéticos.
3. Confirma el issue real y completa el prompt hijo con alcance, HU/CA/RN/DEC, estados, exclusiones, pruebas y evidencia. Solo entonces cambia ese prompt a `ready`.
4. Crea desde `main` la rama declarada. No uses la rama de documentación ni una rama anterior ya integrada.
5. Implementa la menor sección vertical que deje completa la pantalla o familia comprometida: apariencia, interacción, responsive, accesibilidad, estados y pruebas. No hagas una sustitución mecánica global de CSS.
6. Reutiliza una primitiva compartida cuando exista repetición real y estable; conserva el componente dentro del módulo cuando su semántica solo pertenezca a esa capacidad. No crees un sistema abstracto por anticipación.
7. Compara el resultado con el mockup correcto como lenguaje de familia —paleta, jerarquía, densidad, carácter editorial y estados—, no como calco píxel a píxel. Documenta cualquier desviación intencional y su razón funcional o accesible.
8. Ejecuta pruebas de componente, accesibilidad, E2E afectado, lint, tipos y build. Captura evidencia con datos sintéticos y revisa el diff en busca de PII, secretos, cambios funcionales accidentales y dependencias no autorizadas.
9. Actualiza prompt hijo, catálogo, matriz e historial con el estado real. Usa Conventional Commits y abre un PR enfocado con su issue enlazado.
10. Espera checks, revisión, correcciones y merge. Antes de la siguiente fase, vuelve a `main`, actualiza por fast-forward y repite el ciclo.

No alteres archivos generados del cliente OpenAPI, backend, OpenAPI, migraciones ni base de datos para resolver un problema visual. Si la pantalla no puede rediseñarse sin cambiar un contrato o una regla, registra el bloqueo y separa ese trabajo en otro issue.

## Plantilla de trabajo para cada entrega hija

Cada prompt hijo debe declarar:

1. una sola preocupación primaria y un issue real;
2. pantalla(s) y estados exactos incluidos;
3. HU, CA, RN y DEC aplicables;
4. rutas, contratos y pruebas existentes a preservar;
5. mockup de referencia, componentes NAVA nuevos o modificados y dueño de módulo;
6. anchos y estados que requieren evidencia;
7. exclusiones funcionales explícitas del mockup;
8. comandos de verificación;
9. rama y PR reales cuando existan.

No copies este prompt entero dentro de un prompt hijo. Enlázalo y transcribe solo el contexto necesario para que el hijo sea autocontenido.

## Matriz de regresión E2E por familia

La tabla indica el mínimo existente que debe inspeccionarse y ejecutarse cuando resulte afectado. No conviertas nombres de archivos en garantía de cobertura: si falta un estado exigido por el issue, agrega la prueba en la capa más baja adecuada. Si una ruta no tiene E2E dedicado, créalo en el issue de esa pantalla o deja un seguimiento explícito sin declarar cobertura.

| Familia | Specs E2E existentes que deben revisarse |
| --- | --- |
| Shell y navegación | `panel.spec.ts`, `panel-evidencia-responsiva.spec.ts` |
| Acceso y recuperación | `acceso.spec.ts`, `acceso-evidencia-responsiva.spec.ts`, `recuperacion.spec.ts`, `reto-telefonico.spec.ts` |
| Agenda diaria | `agenda-diaria.spec.ts`, `agenda-navegacion-fecha.spec.ts`, `panel-evidencia-responsiva.spec.ts` |
| Nuevo turno | `nuevo-turno.spec.ts` |
| Detalle e historial | `agenda-detalle-historial.spec.ts` |
| Reprogramación | `agenda-reprogramacion-turno.spec.ts` |
| Configuración de barbería | `configuracion-barberia.spec.ts`, `configuracion-barberia-evidencia-responsiva.spec.ts` |
| Barberos | `barberos.spec.ts`, `barberos-evidencia-responsiva.spec.ts` |
| Servicios | `servicios.spec.ts`, `servicios-ciclo-vida.spec.ts`, `servicios-evidencia-responsiva.spec.ts` |
| Servicios por barbero | `servicios-por-barbero.spec.ts`, `servicios-por-barbero-evidencia-responsiva.spec.ts` |
| Horarios y excepciones | `horarios.spec.ts`, `excepciones-festivos.spec.ts` |
| Bloqueos | No existe un spec dedicado en la línea base: comprobar de nuevo y crear cobertura dentro del issue si sigue ausente. |

## Pruebas y evidencia mínimas por entrega visible

- Prueba de componente de interacciones y estados afectados.
- `vitest-axe` con el estado normal y cada diálogo/error relevante abierto.
- E2E del recorrido P0 afectado o seguimiento explícito en el issue si el entorno real impide ejecutarlo.
- Evidencia 320, 360, 768 y 1280 px cuando cambia la composición, organizada bajo `apps/web/e2e/evidence/<familia>/` o la ubicación vigente definida por la estrategia de pruebas.
- Zoom 200 %, teclado completo, foco no oculto y reducción de movimiento.
- Nombres largos, lista vacía, una opción, muchas opciones y conflictos aplicables.
- Estados normal, hover/foco, carga, disabled, validación/error, vacío y éxito cuando existan en la pantalla; no fabriques estados que el contrato no soporta.
- Contraste verificado de las combinaciones realmente usadas.
- Comparación visual razonada contra el mockup asignado; no agregues snapshots de píxeles frágiles como única prueba.
- `git diff --check` y revisión de que no entraron funciones, datos o dependencias no autorizadas.

## Documentación y trazabilidad

- Actualiza `docs/00-control/matriz-trazabilidad.md` e `historial-cambios.md` cuando cambie cobertura o estado real.
- Actualiza `docs/10-backlog/prompts/README.md` al crear o cambiar un prompt hijo.
- Si una entrega altera el sistema global, actualiza `estandar-diseno-visual.md` y `especificacion-frontend-nava.md` en el mismo PR.
- Si una implementación descubre una contradicción funcional, registra `DP-*` o `CT-*` antes de codificar la respuesta.
- No marques este prompt o un hijo `executed` hasta que issue, pruebas, evidencia y PR reflejen el estado real.

## Verificación final aplicable al frontend

```text
pnpm --filter @system-barbershop/web format
pnpm --filter @system-barbershop/web lint
pnpm --filter @system-barbershop/web typecheck
pnpm --filter @system-barbershop/web test:unit
pnpm --filter @system-barbershop/web build
pnpm --filter @system-barbershop/web test:e2e
git diff --check
```

Si Playwright requiere un stack real no disponible, no declares el E2E cumplido: conserva el spec, registra el seguimiento en el issue y entrega la evidencia que sí exista.

Entrega una tabla:

`Criterio | Estado | Prueba o evidencia`

No declares cumplido aquello que no esté probado.

## Formato de respuesta de Claude Code

En su primera respuesta, Claude debe informar: rama actual, estado del árbol, issue/prompt habilitante encontrado, fase que puede ejecutar y cualquier bloqueo. Si este archivo continúa con `issue: pending`, debe entregar únicamente el backlog de issues hijos propuesto y detenerse antes de editar código.

Al cerrar cada entrega, debe resumir:

- issue, prompt hijo, rama, commit y PR;
- archivos y pantallas modificados;
- decisiones visuales adoptadas y libertades usadas;
- estados, anchos y recorridos comprobados;
- resultados exactos de pruebas y enlaces o rutas de evidencia;
- dependencias añadidas con justificación, o confirmación de que no añadió ninguna;
- pendientes o bloqueos reales, sin presentar como terminado lo que siga sin probar.

## Mensaje de continuación después de la Fase 0

El propietario puede pegar el siguiente mensaje en la misma sesión de Claude Code para autorizar la coordinación completa. “Hacer todo” significa continuar sin pedir confirmación rutinaria hasta completar el resultado, pero conserva issues, ramas y PR independientes; nunca significa acumular el rediseño en un único PR.

```text
Continúa con la orquestación completa del rediseño NAVA. Esta es mi autorización expresa para crear y actualizar los issues reales necesarios en GitHub, guardar los prompts hijos, crear ramas, realizar commits, hacer push, abrir PRs, vigilar sus checks y aplicar las correcciones de alcance que sean necesarias. Puedes hacer squash-merge únicamente cuando las protecciones, checks y aprobaciones exigidas por el repositorio estén satisfechas. No me pidas confirmación entre fases para acciones rutinarias ya incluidas en este alcance.

Primero resuelve los habilitantes de forma segura:

1. Conserva íntegros todos los cambios existentes del árbol. No descartes, sobrescribas, reformatees ni incluyas trabajo ajeno. Clasifica los archivos sucios por preocupación y usa staging selectivo o un worktree limpio cuando haga falta.
2. Termina la entrega documental del issue #166 con solo los archivos que realmente correspondan a su alcance; haz commit, push y abre o actualiza su PR. Espera checks y la aprobación requerida antes del merge. Si algún cambio del prompt maestro o del catálogo queda fuera de #166, crea un issue documental independiente y aísla esa entrega en vez de mezclarla.
3. Crea un issue maestro real para esta orquestación con el título y criterios sugeridos en el frontmatter. Actualiza `issue`, `issue_url`, `status` y el catálogo con el número real; no inventes identificadores. El prompt solo puede pasar a `ready` cuando #166 y DEC-077–DEC-079 estén integrados en `main`.
4. Cuando los habilitantes estén integrados, actualiza `main` por fast-forward y crea los issues hijos que sigan siendo necesarios. Como mínimo revisa las fases 6, 7, 9, 10 y 13 detectadas en la Fase 0. No des por conforme una pantalla únicamente porque tenga un issue anterior cerrado: audita todas las rutas contra los nuevos mockups y DEC-079; si una fase previa no cumple la familia visual, los estados o la evidencia exigida, crea un issue de corrección independiente. Si ya cumple y la evidencia es válida, no la reescribas.
5. Ejecuta cada issue hijo de principio a fin y en orden seguro: prompt hijo `ready`, rama nueva desde `main`, implementación, pruebas, evidencia, documentación, commit, push, PR, checks, revisión y merge permitido; después actualiza `main` antes de iniciar el siguiente. No combines varias capacidades en una rama o PR.
6. Continúa hasta que todas las rutas P0 existentes estén rediseñadas o verificadas contra los mockups, y finaliza con la Fase 13 de QA integral. Corrige dentro del mismo issue solo defectos pertenecientes a su alcance; abre seguimientos separados para los demás.

No instales Tailwind ni otra dependencia salvo que un issue demuestre y documente su necesidad. No cambies backend, OpenAPI, base de datos, reglas de negocio, permisos ni alcance para acomodar el diseño. Usa únicamente datos sintéticos en pruebas y capturas.

Detente antes del resultado final solo ante un bloqueo externo real que no puedas resolver de forma segura, como una aprobación humana obligatoria, una decisión funcional faltante o un secreto/servicio inaccesible. Si ocurre, deja el PR listo y reporta una sola vez el bloqueo exacto, la URL afectada y la acción mínima que necesito realizar. De lo contrario, sigue trabajando sin devolverme elecciones rutinarias.

Al finalizar entrega la tabla global `Ruta | Issue/PR | Estado | Pruebas | Evidencia`, los comandos ejecutados y cualquier seguimiento todavía abierto. No declares completado nada que no esté integrado o probado.
```

## Git y PR

- Cada entrega hija usa Conventional Commits y una rama `<tipo>/<issue>-<descripcion>`.
- Usa `Closes #<issue>` solo si el PR cubre el issue completo; de lo contrario, `Refs #<issue>`.
- No hagas push directo, force push ni merge de `main`.
- Un PR no debe mezclar fundaciones, shell, varias capacidades y cambios de dominio.
