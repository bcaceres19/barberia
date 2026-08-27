---
prompt_id: "PROMPT-HU-062-v1"
version: "1.0"
kind: "hu"
status: "draft"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-062"
related_hu:
  - "HU-021"
  - "HU-060"
  - "HU-061"
issue: "pending"
issue_url: null
suggested_issue_title: "feat(agenda): implementar HU-062 agenda diaria de hoy"
branch: null
pr: null
pr_url: null
depends_on:
  - "HU-060 y HU-061 integradas en main"
  - "DP-CIT-04 resuelta como DEC-* y propagada"
  - "DP-CIT-05 resuelta como DEC-* y propagada"
  - "Issue real propio creado y enlazado antes de pasar a ready o ejecutar"
rules:
  - "RN-CIT-01"
  - "RN-RES-02"
  - "RN-RES-03"
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
acceptance_criteria:
  - "CA-062-01"
  - "CA-062-02"
  - "CA-062-03"
  - "CA-062-04"
  - "CA-062-05"
  - "CA-062-06"
  - "CA-062-07"
  - "CA-062-08"
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
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/04-arquitectura/backend-go.md"
  - "docs/04-arquitectura/frontend.md"
  - "docs/05-backend/base-datos.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/06-api/estandar-openapi.md"
  - "docs/10-backlog/plan-bloques.md"
  - "api/openapi/README.md"
  - "api/openapi/openapi.yaml"
  - "apps/api/README.md"
  - "apps/api/internal/modules/booking"
  - "apps/api/internal/modules/staff"
  - "apps/web/README.md"
  - "apps/web/src/app/router"
  - "apps/web/src/modules/agenda"
  - "apps/web/src/modules/auth"
  - "apps/web/src/shared/time"
  - "apps/web/src/shared/ui"
created_at: "2026-08-26"
updated_at: "2026-08-26"
supersedes: null
superseded_by: null
---

# Implementar HU-062: agenda diaria de hoy

## Instrucción para el agente

Implementa únicamente la lectura y pantalla de la agenda de hoy. No ejecutes este prompt mientras `issue: pending`, `HU-061` no esté integrada o `DP-CIT-04`/`DP-CIT-05` no tengan decisiones confirmadas y propagadas. `DEC-047` prohíbe inventar un vínculo automático entre `staff_user` y `barber`; la selección o consolidación inicial y la pertenencia diaria de turnos nocturnos deben seguir literalmente las resoluciones del propietario.

No adelantes anterior/siguiente/selector de fecha, detalle completo ni acciones sobre citas. Esta historia prepara el contrato diario para que `F-CITA-02` lo reutilice después, pero la interfaz solo abre y recarga hoy.

## Objetivo

Que al abrir el panel privado el barbero vea una lista cronológica, accesible y tenant-aware de los turnos de hoy, calculados en la zona IANA de la barbería y con el alcance multi-barbero aprobado.

## Preflight obligatorio

1. Verifica árbol limpio, `main` actualizada y `HU-060`/`HU-061` integradas con pruebas y contrato vigentes.
2. Lee las decisiones que cierren `DP-CIT-04` y `DP-CIT-05`. Si falta alguna o no está propagada a historia/matriz/prompt, detente sin tocar código.
3. Crea o localiza el issue real, actualiza metadatos/índice y crea `feat/<issue>-hu062-agenda-diaria` desde `main`.
4. Ejecuta Graphify sobre `booking`, consulta de `appointment`, `staff`, shell privado, router, zona horaria, cliente API y componentes base.
5. Lee completamente los `source_docs`; revisa especialmente `estandar-diseno-visual.md` §§7, 10 y 11.2 y la semántica de estados/citas nocturnas.
6. Verifica el estado real de evidencia responsive de B2 para no introducir ni copiar un patrón ya marcado como parcial.

## Alcance incluido

- Operación privada de lectura de agenda diaria con fecha civil explícita o default interno de hoy, seguridad de sesión y alcance multi-barbero según `DP-CIT-04`.
- Rango civil calculado en la zona IANA de la barbería y pertenencia de citas nocturnas exactamente como resuelva `DP-CIT-05`.
- Respuesta mínima con ID, `startsAt`, `endsAt`, persona atendida, snapshots de servicio, estado y barbero cuando corresponda.
- Consulta tenant-aware, orden estable y límite/paginación proporcional al volumen; índice/plan verificado.
- Página principal de agenda que abre en hoy, muestra fecha completa, zona, lista cronológica y acción “Nuevo turno”.
- Estados carga inicial, actualización, vacío y error recuperable; etiquetas españolas de los cinco estados usando el patrón visual aprobado.
- Cliente tipado generado, pruebas Go/HTTP/PostgreSQL/Vue/E2E y evidencia responsive/accesible.

## Fuera de alcance

- Controles de día anterior, siguiente o selector de fecha; se entregan en la historia dueña de `F-CITA-02`.
- Vista semanal/mensual, arrastrar, redimensionar, WebSocket o actualización en tiempo real.
- Detalle de contacto, nota, historial o acciones de editar/reprogramar/cancelar/cerrar/corregir.
- Cálculo y visualización de huecos disponibles; B4 es dueña de disponibilidad pública.
- Notificaciones, recordatorios, métricas o cierre automático.
- Resolver `DP-CIT-04`/`DP-CIT-05` mediante defaults de código o crear FK/campo `staff_user`→`barber`.
- Exponer teléfonos, correos, notas, motivos o IDs internos innecesarios en la lista.

## Estado existente que debe conservarse

- `HU-060` define ocupación e intervalos; `HU-061` crea citas manuales reales. Esta HU solo lee.
- `barbershop.timezone` ya existe y `shared/time` tiene utilidades que reciben zona explícita; no uses `Date`/locale del dispositivo como autoridad.
- El shell privado, rutas lazy, navegación y componentes base ya existen. El módulo `agenda` tiene un stub que debe reemplazarse sin importar internos de otros módulos.
- El estándar visual exige lista equivalente en móvil, objetivos de 44 px, estado con texto+señal y “Nuevo turno” como acción principal.
- Los snapshots de `appointment` son la fuente visible de servicio; no hagas join para mostrar el nombre actual del catálogo y reescribir historia.
- Las citas terminales siguen siendo hechos del día y no se borran; solo pueden separarse o atenuarse sin perder legibilidad.

## Trabajo requerido

### 1. Contrato y consulta

1. Propaga las decisiones de `DP-CIT-04` y `DP-CIT-05`. Si su aplicación requiere un cambio material de criterios, crea una versión nueva del prompt antes de ejecutar.
2. Diseña la operación bajo `Appointments`, con seguridad explícita, fecha `format: date`, límites, orden y errores RFC 9457. No documentes SQL.
3. Define la consulta de citas nocturnas de acuerdo con `DP-CIT-05`; no asumas por conveniencia si se filtra por inicio o por intersección.
4. Mantén la respuesta cerrada y mínima; usa nombres técnicos en inglés y descripciones en español.
5. Implementa repositorio con filtro explícito por tenant, alcance de barbero aprobado y rango; ordena por `starts_at`, luego ID estable.
6. Verifica `EXPLAIN (ANALYZE, BUFFERS)` con volumen representativo. Ajusta índice solo si una consulta demostrada lo necesita y mediante migración nueva.

### 2. Backend Go

1. Añade `ListDailyAgenda` al núcleo `booking` con reloj y zona explícitos; no conozca Chi ni SQL.
2. Deriva tenant de sesión y valida el alcance multi-barbero según la decisión. Recurso ajeno responde como no encontrado.
3. Convierte fecha civil a dos instantes inequívocos en la zona de barbería, incluidos cambios DST; no uses 24 horas fijas si el día local dura 23/25 horas.
4. Mapea filas a un DTO de respuesta sin contacto/nota/historial. Traduce errores sin filtrar datos.
5. Prueba cancelación de contexto, límites, agenda vacía y múltiples estados.

### 3. Frontend Vue

1. Sustituye el stub de agenda por una ruta lazy y página dueña del módulo, sin mover reglas al shell.
2. Al entrar, consulta hoy según la zona devuelta por el servidor o el contrato aprobado; muestra fecha completa y zona.
3. Aplica el selector/vista consolidada de `DP-CIT-04` con estados carga/vacío/error y actualización local.
4. Renderiza lista cronológica: hora, persona, servicio snapshot, barbero cuando corresponda y `AppointmentStatusBadge` o equivalente basado en tokens.
5. La acción primaria “Nuevo turno” navega al flujo real de `HU-061`. No agregues acciones futuras deshabilitadas como promesa falsa.
6. En móvil no uses una cuadrícula espacial obligatoria. En escritorio puede haber composición ampliada solo si conserva la lista accesible equivalente.
7. Verifica teclado, foco, lector semántico, 44 px, reducción de movimiento, zoom 200 % y 320/360/768/1280 px.

## Pruebas y evidencia

- Dominio: hoy en zona explícita, servidor/dispositivo en otra zona, día de 23/25 h, medianoche conforme a `DP-CIT-05`, orden y estados.
- PostgreSQL real: dos tenants, alcance multi-barbero aprobado, citas que empiezan/terminan en los límites, nocturnas, terminales e índice con volumen.
- HTTP/contrato: agenda vacía/llena, fecha inválida, auth, recurso ajeno, límites y forma exacta sin datos personales innecesarios.
- Componentes: carga, actualización, vacío, error/reintento, uno/varios turnos, cinco estados y “Nuevo turno”; axe-core.
- E2E: login, apertura en hoy, dispositivo en otra zona, agenda cronológica con estados y navegación a crear turno.
- Responsive/accesible: 320, 360, 768, 1280 px, zoom 200 %, teclado, foco, contraste y teléfono real antes de versión.

## Documentación y trazabilidad

- Actualiza OpenAPI/CHANGELOG, cliente generado, README API/web, matriz, historial, plan y estado real del prompt.
- Documenta el plan de consulta y cualquier índice nuevo. No agregues migración si el índice de `HU-060` ya sirve.
- Mantén `F-CITA-02` pendiente: esta entrega abre hoy, pero no permite cambiar de fecha.
- Registra dudas nuevas antes de inventar defaults, filtros o campos.

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

Ejecuta las integraciones contra PostgreSQL 14 real con dos tenants. Entrega `Criterio | Estado | Prueba o evidencia` para `CA-062-01`–`CA-062-08`, junto con el `EXPLAIN`, evidencias responsive y aclaración expresa de que `F-CITA-02` sigue fuera de alcance.

## Git y PR

- Rama sugerida después de crear issue: `feat/<issue>-hu062-agenda-diaria`.
- Commit/título: `feat(agenda): implementa HU-062 agenda diaria de hoy`.
- Usa `Closes #<issue>` solo con los ocho criterios y la evidencia visual/E2E completa; si queda evidencia propia parcial, usa `Refs #<issue>` y mantén el issue abierto.
- No hagas push directo, force push, merge de `main`, edición de migraciones aplicadas ni implementación oportunista de navegación/acciones futuras.
