---
prompt_id: "PROMPT-ORCH-NAVA-FRONTEND-v1"
version: "1.0"
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
suggested_issue_title: "docs(web): planificar adopción incremental del diseño NAVA"
branch: null
pr: null
pr_url: null
depends_on:
  - "DEC-077 registrada y especificación NAVA integrada en main"
  - "Un issue real por cada fundación o pantalla que se vaya a migrar"
rules:
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
  - "DEC-039"
  - "DEC-074"
  - "DEC-075"
  - "DEC-077"
acceptance_criteria: []
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
  - "docs/10-backlog/prompts/README.md"
created_at: "2026-09-01"
updated_at: "2026-09-01"
supersedes: null
superseded_by: null
---

# Orquestación de la adopción incremental del frontend NAVA

## Instrucción para Claude o el agente ejecutor

Usa este archivo como mapa de coordinación, no como autorización para rediseñar todo el frontend en una sola rama. La identidad NAVA y su especificación son normativas, pero cada fundación o pantalla debe llegar mediante un issue real y, cuando el cambio sea no trivial, un prompt hijo autocontenido que atienda una sola preocupación primaria.

Este prompt permanece en `draft` mientras `issue: pending`. No cambies código desde este prompt ni lo marques `ready` hasta registrar un issue real de orquestación o planificación. La existencia de una pantalla en la especificación tampoco autoriza una capacidad P0 pendiente, P1, P2 o excluida.

## Objetivo

Coordinar una migración gradual y verificable desde la interfaz actual al sistema **NAVA / Tailored Grid**, de modo que cada nueva pantalla nazca en el estándar NAVA y cada pantalla existente se migre completa —estados, responsive, accesibilidad y pruebas incluidos— sin alterar contratos, reglas de negocio, seguridad, auditoría ni alcance.

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
- Verificar tokens, tipografía, responsive, accesibilidad, estados asíncronos y pruebas de cada entrega.
- Actualizar trazabilidad documental y el catálogo de prompts a medida que cada entrega cambia de estado.

## Fuera de alcance

- Un rediseño masivo de `apps/web` en una única rama o PR.
- Implementar B3, B4, B5 o B6 solo porque su pantalla está especificada.
- Crear ingresos, reportes, inventario, caja, pagos, comisiones o perfiles de cliente.
- Crear el estado `in_progress`; “Ahora” es únicamente una señal temporal.
- Crear una vista consolidada multi-barbero mientras `DEC-074` siga gobernando P0.
- Añadir modo oscuro, temas por tenant, app nativa, animaciones decorativas o biblioteca visual pesada.
- Cargar fuentes desde CDN o incorporar assets sin licencia y medición.
- Modificar OpenAPI, persistencia o reglas como efecto lateral de un cambio visual.

## Estado existente que debe conservarse

- `apps/web` ya tiene Vue 3, TypeScript estricto, Vue Router, cliente tipado y módulos por capacidad.
- Existen primitivas compartidas como botones, inputs, diálogos, alertas e insignias; deben auditarse y evolucionarse, no duplicarse por pantalla.
- B0, B1, B2 y `HU-060`–`HU-065` tienen entregas integradas con distintos seguimientos E2E/responsive pendientes registrados en la matriz.
- Los tokens actuales son legado hasta que un issue de fundaciones los migre de forma completa. La documentación de NAVA no significa que el CSS ya esté migrado.
- La agenda P0 exige un solo barbero seleccionado (`DEC-074`) y los turnos nocturnos aparecen en cada día intersectado (`DEC-075`).
- Los estados P0 son `confirmed`, `completed`, `cancelled_by_customer`, `cancelled_by_barber` y `no_show`.
- La interfaz usa “turno”; `appointment` permanece en API, datos y código (`DEC-016`).

## Plan de entregas recomendado

### Fase 1 · Fundaciones NAVA

Crear un issue propio para:

- migrar `tokens.css` y `base.css` a los valores de `estandar-diseno-visual.md` 2.0;
- incorporar `Instrument Serif` e `Instrument Sans` como WOFF2 self-hosted, con licencia, fallbacks y medición;
- adaptar solo las primitivas realmente utilizadas (`BaseButton`, `BaseInput`, `BaseDialog`, `BaseAlert`, `BaseBadge`, etc.);
- documentar transición y evitar que una pantalla quede mitad legado/mitad NAVA.

No modificar pantallas de negocio salvo una ruta de laboratorio o la mínima evidencia necesaria para validar las primitivas.

### Fase 2 · Shell privado

Issue propio para `PrivateAppShell`, wordmark, header, dock de escritorio/móvil, safe areas, foco tras navegación y estados de sesión. Conservar rutas, guardas y permisos. No agregar destinos excluidos. Destinos P0: Agenda, Servicios, Barberos, Horarios y Configuración; “Nuevo turno” es acción, no módulo.

### Fase 3 · Seguridad y acceso

Separar por preocupación compatible con los issues existentes:

- acceso y desafío por abuso;
- solicitud/verificación de recuperación;
- nueva contraseña y resultado;
- sesión vencida.

No cambiar respuestas neutrales ni controles de seguridad al rediseñar.

### Fase 4 · Operación de agenda

Entregas independientes y ordenadas:

1. agenda diaria + selector obligatorio de barbero + fecha;
2. nuevo turno manual;
3. detalle e historial;
4. reprogramación;
5. futuras acciones de ciclo de vida solo cuando tengan HU/issue.

La representación temporal de escritorio siempre tiene alternativa de lista. No implementar el mosaico multi-barbero del concepto.

### Fase 5 · Configuración del negocio

Separar al menos:

- barbería y barberos;
- catálogo de servicios;
- asignación de servicios por barbero;
- horario semanal;
- excepciones/festivos;
- bloqueos y sus alcances realmente soportados.

No combinar todas las configuraciones en un formulario o PR.

### Fase 6 · Reserva pública

Solo cuando B4 tenga historias, contrato e issues reales. Separar el shell público y el recorrido si los riesgos lo justifican, manteniendo una experiencia coherente de principio a fin. Cubrir servicio, barbero condicional, fecha/hora, datos, revisión, confirmación, conflicto de slot y gestión por token sin cuenta.

### Fase 7 · P1/P2

No preparar código anticipado. Cuando una capacidad sea priorizada, crear su decisión/HU/issue y volver a leer la especificación para reutilizar los patrones reservados.

## Plantilla de trabajo para cada entrega hija

Cada prompt hijo debe declarar:

1. una sola preocupación primaria y un issue real;
2. pantalla(s) y estados exactos incluidos;
3. HU, CA, RN y DEC aplicables;
4. rutas, contratos y pruebas existentes a preservar;
5. componentes NAVA nuevos o modificados, con dueño de módulo;
6. anchos y estados que requieren evidencia;
7. exclusiones explícitas del mockup;
8. comandos de verificación;
9. rama y PR reales cuando existan.

No copies este prompt entero dentro de un prompt hijo. Enlázalo y transcribe solo el contexto necesario para que el hijo sea autocontenido.

## Pruebas y evidencia mínimas por entrega visible

- Prueba de componente de interacciones y estados afectados.
- `vitest-axe` con el estado normal y cada diálogo/error relevante abierto.
- E2E del recorrido P0 afectado o seguimiento explícito en el issue si el entorno real impide ejecutarlo.
- Evidencia 320, 360, 768 y 1280 px cuando cambia la composición.
- Zoom 200 %, teclado completo, foco no oculto y reducción de movimiento.
- Nombres largos, lista vacía, una opción, muchas opciones y conflictos aplicables.
- Contraste verificado de combinaciones nuevas.
- `git diff --check` y revisión de que no entraron hexadecimales o tamaños arbitrarios en módulos.

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

## Git y PR

- Cada entrega hija usa Conventional Commits y una rama `<tipo>/<issue>-<descripcion>`.
- Usa `Closes #<issue>` solo si el PR cubre el issue completo; de lo contrario, `Refs #<issue>`.
- No hagas push directo, force push ni merge de `main`.
- Un PR no debe mezclar fundaciones, shell, varias capacidades y cambios de dominio.
