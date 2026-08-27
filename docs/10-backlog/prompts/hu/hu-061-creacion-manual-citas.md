---
prompt_id: "PROMPT-HU-061-v1"
version: "1.0"
kind: "hu"
status: "ready"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-061"
related_hu:
  - "HU-023"
  - "HU-040"
  - "HU-041"
  - "HU-042"
  - "HU-060"
  - "HU-062"
issue: "107"
issue_url: "https://github.com/bcaceres19/barberia/issues/107"
suggested_issue_title: "feat(booking): implementar HU-061 creación manual de turnos"
branch: null
pr: null
pr_url: null
depends_on:
  - "HU-060 integrada en main con CA-060-01 a CA-060-08 verificadas (PR #105)"
  - "DP-CIT-01 resuelta como DEC-071 y propagada"
  - "DP-CIT-02 resuelta como DEC-072 y propagada"
  - "DP-CIT-03 resuelta como DEC-073 y propagada"
rules:
  - "RN-CIT-02"
  - "RN-RES-01"
  - "RN-RES-02"
  - "RN-RES-03"
  - "RN-CON-01"
  - "RN-CON-03"
  - "RN-DIS-04"
  - "RN-DIS-05"
  - "RN-DIS-06"
  - "RN-DIS-07"
  - "RN-HIS-01"
  - "RN-TEN-01"
  - "RN-DAT-01"
  - "RN-DAT-02"
  - "RN-IDE-01"
decisions:
  - "DEC-002"
  - "DEC-004"
  - "DEC-005"
  - "DEC-006"
  - "DEC-007"
  - "DEC-016"
  - "DEC-019"
  - "DEC-020"
  - "DEC-024"
  - "DEC-035"
  - "DEC-036"
  - "DEC-037"
  - "DEC-038"
  - "DEC-039"
  - "DEC-040"
  - "DEC-043"
  - "DEC-045"
  - "DEC-046"
  - "DEC-071"
  - "DEC-072"
  - "DEC-073"
acceptance_criteria:
  - "CA-061-01"
  - "CA-061-02"
  - "CA-061-03"
  - "CA-061-04"
  - "CA-061-05"
  - "CA-061-06"
  - "CA-061-07"
  - "CA-061-08"
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
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/06-api/estandar-openapi.md"
  - "docs/10-backlog/plan-bloques.md"
  - "api/openapi/README.md"
  - "api/openapi/openapi.yaml"
  - "apps/api/README.md"
  - "apps/api/internal/modules/booking"
  - "apps/api/internal/modules/catalog"
  - "apps/api/internal/modules/schedule"
  - "apps/web/README.md"
  - "apps/web/src/modules/agenda"
  - "apps/web/src/modules/catalog"
  - "apps/web/src/modules/schedules"
  - "apps/web/src/shared/api"
created_at: "2026-08-26"
updated_at: "2026-08-27"
supersedes: null
superseded_by: null
---

# Implementar HU-061: creación manual de turnos

## Instrucción para el agente

`DP-CIT-01`–`DP-CIT-03` ya están resueltas y propagadas (`DEC-071`–`DEC-073`, 2026-08-27) y el issue real es [#107](https://github.com/bcaceres19/barberia/issues/107); este prompt está `ready`. Aplica esas tres decisiones literalmente, sin reinterpretarlas por facilidad técnica.

Entrega una sola capacidad vertical: el barbero crea un turno manual desde el panel y la operación termina en una única cita `confirmed` con historial, aislamiento e idempotencia. No implementes edición, reprogramación, cancelación, estados terminales, reserva pública ni notificaciones.

## Objetivo

Que un barbero autenticado registre un turno recibido por teléfono, WhatsApp o en persona desde su celular, en cualquier minuto válido y sin límites públicos de anticipación/ventana, con decisiones aprobadas sobre cliente, asignación de servicio y bloqueos vigentes.

## Preflight obligatorio

1. Verifica árbol limpio, `main` actualizada por fast-forward y `HU-060` integrada con sus ocho criterios probados (PR #105).
2. Lee `DEC-071`, `DEC-072` y `DEC-073` (resoluciones de `DP-CIT-01`–`DP-CIT-03`) en `docs/00-control/registro-decisiones.md` antes de codificar.
3. Actualiza `status` a `in_progress` y `branch` en los metadatos de este archivo y en el índice; crea `feat/107-hu061-creacion-manual-turnos` desde `main`.
4. Ejecuta Graphify sobre `booking`, `catalog`, `schedule`, `staff`, sesión/tenant, idempotencia, navegación privada y componentes base.
5. Lee completamente todos los `source_docs` y verifica que contrato, migraciones y cliente generado vigentes coincidan con la rama base.
6. Revisa los seguimientos de B2 (`#90`, `#95`, `#98`, `#100`) y conserva cualquier corrección integrada de horario/bloqueos/UI.

## Alcance incluido

- `POST /api/v1/private/appointments` o path equivalente aprobado contract-first, con `SessionCookie` e `Idempotency-Key` requeridos.
- Caso de uso `booking.CreateManualAppointment` independiente de Chi y PostgreSQL.
- Barbero, servicio, persona atendida, instante local interpretado en la zona de la barbería, contacto opcional y nota opcional con minimización.
- Fin planificado y snapshots derivados en servidor del servicio autorizado; estado y origen derivados, no aceptados del body.
- `DEC-071`–`DEC-073` (resoluciones de `DP-CIT-01`–`DP-CIT-03`) implementadas de forma literal y trazable.
- Exención manual de anticipación mínima, ventana máxima y rejilla pública; conserva jornada, integridad y cualquier otra restricción aprobada.
- Reutilización de idempotencia de `HU-004`/`DEC-043`, transacción de `HU-060` y traducción segura del conflicto de exclusión.
- Formulario Vue “Nuevo turno”, cliente tipado generado, validación de forma, resumen y estados visibles.

## Fuera de alcance

- Reinterpretar `DEC-071`–`DEC-073` durante la implementación; se aplican tal como quedaron redactadas.
- Consultar disponibilidad pública o proponer franjas alternativas; B4 es dueña.
- Crear/editar barberos, servicios, asignaciones, horarios, excepciones o bloqueos.
- Modificar, reprogramar, cancelar, completar, marcar `no_show` o corregir una cita.
- Token público, enlace del cliente, confirmación por proveedor, programación/envío de recordatorios o worker.
- Permitir duración, precio, moneda, snapshot, estado, origen, actor o `barbershopId` desde el request.
- Agregar dependencia visual/calendario o estado global sin la justificación exigida.

## Estado existente que debe conservarse

- `HU-060` debe proveer las tablas y la primitiva atómica cliente+cita+historial; no dupliques esa transacción en el handler.
- `catalog` posee servicios/asignaciones; `schedule` posee jornada efectiva y bloqueos. `booking` colabora mediante puertos pequeños, nunca SQL contra tablas de otro módulo.
- `HU-004` ya implementa idempotencia y `DEC-043` exige conflicto inmediato para la misma clave concurrente.
- `DEC-005`/`DEC-006` eximen solo anticipación, ventana y rejilla; no son una autorización genérica para saltar integridad.
- El frontend ya tiene shell privado, navegación, cliente OpenAPI, `BaseInput`, `BaseSelect`, botones, alertas y tokens aprobados; reutilízalos.
- La interfaz dice “turno”; contrato, código y datos dicen `appointment` (`DEC-016`).

## Trabajo requerido

### 1. Cerrar el contrato antes de codificar

1. `DEC-071`–`DEC-073` ya están propagadas a reglas, historia y matriz; verifica que siguen vigentes antes de diseñar el request.
2. Diseña un request cerrado con solo campos que el barbero aporta. Documenta unidad, zona, opcionalidad, longitudes y ejemplos ficticios.
3. Declara `201`, repetición lógica, `400`, `401`, `404`, `409`, `422` y `500` con RFC 9457; no uses `default`.
4. Incluye `x-business-rules` y `x-decisions` reales; no describas tablas o SQL en OpenAPI.
5. Lint y bundle antes de generar el cliente TypeScript. El archivo generado nunca se edita a mano.

### 2. Caso de uso y adaptadores Go

1. Deriva tenant/actor desde la sesión. Resuelve barbero, servicio, asignación y jornada/bloqueo mediante puertos según las decisiones aprobadas.
2. Interpreta la fecha/hora civil con la zona IANA de la barbería y deriva `ends_at` desde la duración del servicio; no uses la zona del servidor/dispositivo.
3. Aplica la política manual: sin anticipación mínima, ventana máxima ni rejilla, incluida una cita que ya empezó cuando las demás reglas la permitan.
4. Construye snapshots desde el catálogo real y llama una vez a la primitiva transaccional de `HU-060` dentro del protocolo de idempotencia.
5. Traduce cruce, recurso ajeno, transición no permitida, validación e idempotencia a errores estables sin datos personales.
6. No programes notificaciones ni recordatorios ficticios. Documenta el efecto pendiente de B5 sin afirmar que ocurrió.

### 3. Frontend Vue

1. Agrega “Nuevo turno” al shell/ruta privada mediante la API pública mínima del módulo `agenda`/`booking`, sin importar internals de `catalog` o `schedules`.
2. Ordena el formulario: servicio/barbero según decisión, persona, fecha/hora, contacto, nota y resumen.
3. Muestra la zona de la barbería junto a la hora. Conserva datos no sensibles ante `409`, `422` o fallo recuperable.
4. Evita doble toque y genera una clave de idempotencia por intención; no reutilices una clave tras cambiar el payload.
5. Gestiona carga de opciones, vacío, errores por campo, error global, conflicto y éxito persistente. El éxito navega o enlaza a la agenda real solo si `HU-062` ya existe; de lo contrario, muestra un resumen honesto.
6. Reutiliza tokens/componentes; verifica 44 px, teclado, foco, zoom 200 %, reducción de movimiento y 320/360/768/1280 px.

## Pruebas y evidencia

- Dominio/servicio: cualquier minuto, cita ya iniciada, zona distinta, duración/snapshots y todas las ramas de `DEC-071`–`DEC-073`.
- HTTP/contrato: request válido, desconocidos, exceso, auth, recurso ajeno, conflicto, validación, idempotencia igual/distinta/concurrente y error interno seguro.
- PostgreSQL real: dos tenants, contigüidad, cruces, rollback atómico y carrera de dos conexiones sin `sleep`.
- Componentes: carga de catálogos, datos opcionales, errores, resumen, doble toque, conservación y foco; axe-core.
- E2E: barbero crea turno manual fuera de la rejilla pública, recarga y observa una sola cita; repetir con dispositivo en otra zona y con conflicto real.
- Evidencia visual/accesible: 320, 360, 768, 1280 px, zoom 200 %, teclado, foco y contraste; no solo capturas.

## Documentación y trazabilidad

- Actualiza OpenAPI/CHANGELOG, cliente generado, README de API/web, matriz, historial, plan y este catálogo.
- Registra explícitamente qué efectos de T1 quedan para B5; no marques recordatorios/notificaciones como cumplidos.
- Si el contrato necesita una decisión no cubierta por `DP-CIT-01`–`DP-CIT-03`, registra otra duda antes de codificar.

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

Ejecuta las pruebas de migración/RLS de `HU-060` y las integraciones de `HU-061` contra PostgreSQL 14 real. Entrega `Criterio | Estado | Prueba o evidencia` para `CA-061-01`–`CA-061-08`; separa cualquier efecto pendiente de B5.

## Git y PR

- Rama sugerida después de crear issue: `feat/<issue>-hu061-creacion-manual-turnos`.
- Commit/título: `feat(booking): implementa HU-061 creación manual de turnos`.
- Usa `Closes #<issue>` solo con los ocho criterios y decisiones resueltas; si queda un efecto propio incompleto, usa `Refs #<issue>`.
- No hagas push directo, force push, merge de `main`, edición de migraciones aplicadas ni cambios ajenos de B2/B4/B5.
