---
prompt_id: "PROMPT-HU-063-v1"
version: "1.2"
kind: "hu"
status: "in_progress"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-063"
related_hu:
  - "HU-062"
  - "HU-064"
issue: 116
issue_url: "https://github.com/bcaceres19/barberia/issues/116"
suggested_issue_title: "feat(agenda): implementar HU-063 navegación por fecha"
branch: "feat/116-hu063-navegacion-agenda"
pr: 118
pr_url: "https://github.com/bcaceres19/barberia/pull/118"
depends_on:
  - "HU-062 integrada en main mediante PR #113 (cumplido)"
  - "Issue real propio creado y enlazado (cumplido: #116)"
rules:
  - "RN-CIT-01"
  - "RN-DIS-05"
  - "RN-DIS-07"
  - "RN-TEN-01"
decisions:
  - "DEC-007"
  - "DEC-016"
  - "DEC-019"
  - "DEC-020"
  - "DEC-024"
  - "DEC-033"
  - "DEC-034"
  - "DEC-035"
  - "DEC-036"
  - "DEC-037"
  - "DEC-038"
  - "DEC-039"
  - "DEC-047"
  - "DEC-074"
  - "DEC-075"
acceptance_criteria:
  - "CA-063-01"
  - "CA-063-02"
  - "CA-063-03"
  - "CA-063-04"
  - "CA-063-05"
  - "CA-063-06"
  - "CA-063-07"
  - "CA-063-08"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/00-control/glosario.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/estados-citas.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/06-api/estandar-openapi.md"
  - "docs/10-backlog/plan-bloques.md"
  - "api/openapi/paths/private-appointments.yaml"
  - "apps/api/internal/modules/booking/agenda.go"
  - "apps/api/internal/modules/booking/postgres/repository.go"
  - "apps/web/src/modules/agenda"
  - "apps/web/src/shared/time"
created_at: "2026-08-31"
updated_at: "2026-08-31"
supersedes: null
superseded_by: null
---

# Implementar HU-063: navegación de la agenda por fecha

## Instrucción para el agente

Implementa únicamente `HU-063`: anterior, selector de fecha y siguiente sobre la agenda diaria real de `HU-062`. Conserva literalmente `DEC-074` (un `barberId` explícito, sin vista consolidada) y `DEC-075` (un turno nocturno aparece en cada día civil cuyo rango interseca).

El issue real [#116](https://github.com/bcaceres19/barberia/issues/116) ya está enlazado y este prompt pasó a `ready`. No adelantes detalle, historial ni acciones sobre citas de `HU-064`/`HU-065`.

## Objetivo

Que el barbero navegue a cualquier fecha civil desde la agenda, conserve el barbero seleccionado y pueda recargar, volver o avanzar sin que la zona del dispositivo, una respuesta obsoleta o un cambio de ruta muestre otro día.

## Preflight obligatorio

1. Comprueba árbol limpio y preserva cambios ajenos; actualiza `main` mediante fast-forward.
2. Verifica que `HU-062` y PR #113 estén integradas y revisa sus pendientes E2E/responsive para no declararlos resueltos por esta HU.
3. Issue real ya enlazado ([#116](https://github.com/bcaceres19/barberia/issues/116)); crea la rama `feat/116-hu063-navegacion-agenda` desde `main` actualizada.
4. Ejecuta Graphify sobre `agenda`, router, cliente generado, utilidades de tiempo y operación `listDailyAgenda`.
5. Lee completamente los `source_docs`, en especial diseño visual §§9, 10, 11.2 y 12.
6. Confirma que no exista una duda o contradicción abierta sobre fecha, barbero, medianoche o URL. Si aparece una, regístrala antes de codificar.

## Alcance incluido

- Tres controles estables: día anterior, selector de fecha y día siguiente; la entrada inicial sigue siendo hoy.
- Fecha civil calculada en la zona IANA de la barbería, incluso en días locales de 23/25 horas.
- Reutilización de `listDailyAgenda` y su parámetro `date`; ningún endpoint alternativo.
- Conservación recíproca de fecha y `barberId` al cambiar cualquiera de los dos.
- Estado navegable restaurable tras recarga, atrás/adelante y retorno desde rutas hijas, sin agregar estado global.
- Cancelación o descarte de peticiones obsoletas para impedir que una respuesta tardía reemplace la selección vigente.
- Carga, actualización, vacío, error y reintento con fecha, zona y barbero visibles.
- Pruebas de componente/E2E y evidencia responsive/accesible.

## Fuera de alcance

- Detalle, contacto, historial, edición, reprogramación, cancelación o estado de una cita.
- Vista semanal/mensual, desplazamiento infinito, arrastrar, huecos, disponibilidad pública o tiempo real.
- Vista consolidada o vínculo automático `staff_user`→`barber`.
- Cambiar semántica, datos o índices del backend ya verificado salvo defecto reproducible.
- Crear dependencias, tokens o componentes visuales fuera del sistema aprobado.

## Estado existente que debe conservarse

- `GET /private/barbers/{barberId}/appointments/daily-agenda` ya acepta `date` opcional `AAAA-MM-DD` y calcula el rango en la zona de la barbería.
- `AgendaService` usa reloj inyectado y días civiles, no sumas de 24 horas.
- La consulta PostgreSQL filtra por intersección y ya está probada con `DEC-075` y dos tenants.
- `DailyAgendaPage.vue` ya controla selección obligatoria de barbero, carga, vacío, error y acción “Nuevo turno”.
- La lista no contiene contacto ni nota y los estados terminales permanecen visibles.

## Trabajo requerido

### 1. Modelo de navegación

1. Define un único estado fuente para `date` en formato civil estricto; no guardes un instante UTC como fecha seleccionada.
2. Deriva anterior/siguiente con aritmética de calendario en la zona de la barbería; prueba cambios de mes, año y DST.
3. Representa fecha y barbero en estado de navegación restaurable. Conserva compatibilidad con la entrada `/panel` de `HU-062`.
4. Distingue selección solicitada de datos confirmados para conservar contenido durante actualización y error.

### 2. Cliente y concurrencia de lectura

1. Reutiliza el cliente tipado generado para enviar `date`; no dupliques DTO ni construyas parámetros libres.
2. Cancela la petición anterior o compara una identidad de solicitud antes de publicar el resultado.
3. Un aborto por navegación no se muestra como error; un fallo real conserva selección y ofrece reintento.
4. No cambies OpenAPI ni Go si el contrato actual satisface la historia. Si detectas un defecto, demuéstralo primero y mantén el cambio mínimo con sus regresiones.

### 3. Interfaz Vue

1. Añade anterior, selector y siguiente con posición estable conforme al estándar visual.
2. Mantén fecha completa y zona visibles; usa “turno” en la interfaz y nombres técnicos en código.
3. Conserva `barberId` al cambiar fecha y la fecha al cambiar barbero.
4. Sincroniza recarga y navegación del navegador sin ciclo de watchers, petición duplicada ni estado global.
5. Asegura objetivos de 44 px, teclado, foco, zoom 200 %, reducción de movimiento y reflow en los anchos aprobados.

## Pruebas y evidencia

- Modelo: hoy, anterior/siguiente, fin de mes/año, día de 23/25 horas y dispositivo en otra zona.
- Cliente: fecha exacta, `AbortSignal` o mecanismo equivalente, error real y respuestas fuera de orden.
- Componente: los tres controles, cambio de barbero, recarga/atrás/adelante, carga/actualización/vacío/error/reintento y axe-core.
- Regresión backend: contrato y pruebas de `HU-062` para límites, nocturnidad y dos tenants si ningún código backend cambia.
- E2E real: hoy → anterior → siguiente → fecha elegida → recarga → atrás/adelante, siempre con fecha/barbero correctos.
- Evidencia: 320, 360, 768 y 1280 px, zoom 200 %, teclado, foco y contraste.

## Documentación y trazabilidad

- Actualiza estado de `HU-063`, matriz, historial, plan, README web y este prompt con issue/rama/PR reales.
- Cambia OpenAPI/CHANGELOG/README API solo si el comportamiento HTTP realmente cambia.
- Conserva `HU-064`/`HU-065` como preocupaciones separadas y no marques evidencia pendiente de `HU-062` como completada sin ejecutarla.

## Verificación final

```text
pnpm run openapi:check-config
pnpm run openapi:lint
pnpm run openapi:bundle

cd apps/api
gofmt -l .
go vet ./...
go test -race ./...
go build ./...

cd ../web
pnpm run format
pnpm run lint
pnpm run typecheck
pnpm run test:unit
pnpm run build
pnpm run test:e2e

cd ../..
git diff --check
graphify update .
```

Entrega `Criterio | Estado | Prueba o evidencia` para `CA-063-01`–`CA-063-08`. No declares E2E o evidencia visual cumplidos si solo existe el spec.

## Git y PR

- Rama sugerida tras crear el issue: `feat/<issue>-hu063-navegacion-agenda`.
- Commit/título: `feat(agenda): implementa HU-063 navegación por fecha`.
- Usa `Closes #<issue>` solo si los ocho criterios y evidencia están completos; de lo contrario `Refs #<issue>` y deja seguimiento explícito.
- No hagas push directo, force push, merge de `main` ni mezcles `HU-064`/`HU-065`.
