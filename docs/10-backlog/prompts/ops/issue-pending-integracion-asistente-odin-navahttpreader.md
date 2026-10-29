---
prompt_id: "PROMPT-OPS-ODIN-NAVAHTTPREADER-v1"
version: "1.0"
kind: "ops"
status: "draft"
target_agents:
  - "any"
repository: "externo — repositorio propio del asistente de IA de Odin, fuera de bcaceres19/barberia"
base_branch: null
primary_hu: null
related_hu: []
issue: "pending"
issue_url: null
suggested_issue_title: "ops(odin): adaptador NavaHttpReader sobre la API pública de NAVA"
branch: null
pr: null
pr_url: null
depends_on: []
rules:
  - "RN-DIS-01"
  - "RN-DIS-02"
  - "RN-DIS-03"
  - "RN-DIS-04"
  - "RN-DIS-05"
  - "RN-DIS-06"
  - "RN-DIS-07"
  - "RN-TEN-01"
decisions: []
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - "api/openapi/openapi.yaml"
  - "api/openapi/paths/public-booking.yaml"
created_at: "2026-10-22"
updated_at: "2026-10-22"
supersedes: null
superseded_by: null
---

# Adaptador `NavaHttpReader` sobre la API pública de NAVA

## Instrucción para el agente

Este prompt no modifica el repositorio `bcaceres19/barberia`. Está dirigido a Odin (u otro agente) para que construya o complete, **en su propio repositorio**, el adaptador de solo lectura que consulta la API pública ya existente de NAVA. No implica crear issue, rama ni PR en `barberia`; por eso queda `status: draft` e `issue: pending` de forma permanente, salvo que el propietario decida versionar esta integración también dentro de `barberia`.

## Objetivo

Que el asistente de IA de Odin calcule disponibilidad de una barbería NAVA (demo: `barberia-demo`) leyendo exclusivamente la API pública HTTP de NAVA, sin copiar su lógica de disponibilidad, sin leer `apps/api/.env.local`, sin credenciales de base de datos y sin tocar su PostgreSQL.

## Bloqueo que resuelve este prompt

Odin reportó: necesitaba las credenciales del rol de aplicación de NAVA (`apps/api/.env.local`) para compilar y correr su propia copia del API y así calcular disponibilidad. Hizo bien en no rodear el control de permisos que le negó leer ese archivo. **No hacía falta**: NAVA ya expone esa disponibilidad por un endpoint público, de solo lectura, sin autenticación, documentado y estable. Odin no necesita compilar, ejecutar ni tener credenciales del API de NAVA en ningún momento; solo necesita una URL base HTTP de una instancia de NAVA ya corriendo.

## Alcance incluido

- Un cliente HTTP (`NavaHttpReader` o el nombre que use el proyecto de Odin) que consuma, en este orden, los cuatro endpoints públicos de `GET /public/barbershops/...` descritos abajo.
- Manejo de los contratos de error/vacío exactamente como los documenta el OpenAPI de NAVA (ver "Contrato exacto").
- Pruebas del adaptador en el repositorio de Odin contra una instancia de NAVA real (local o la que el propietario indique), no contra mocks que inventen la forma de la respuesta.

## Fuera de alcance

- Reimplementar o aproximar el cálculo de disponibilidad (jornada, excepciones, bloqueos, citas) dentro del proyecto de Odin: eso es exactamente lo que este adaptador evita duplicar.
- Cualquier escritura: el adaptador es de solo lectura. Reservar o confirmar un turno (`POST /public/.../appointments`) está fuera de este prompt.
- Cualquier acceso a PostgreSQL de NAVA, a `apps/api/.env.local` o a su proceso compilado.
- Cambios en el repositorio `barberia`.

## Contrato exacto (fuente: `api/openapi/paths/public-booking.yaml` en `barberia`, rama `main`)

Base de la API pública: `{BASE_URL}/api/v1/public` (sin autenticación, `security: []` en los cuatro).

1. `GET /barbershops/{slug}` — `operationId: resolvePublicBarbershop`. Resuelve la barbería por su slug público. `404` uniforme si el slug no existe, está mal formado o la barbería no es publicable (RN-TEN-01): esa respuesta nunca distingue la causa.
2. `GET /barbershops/{slug}/services` — `operationId: listPublicServices`. Catálogo de servicios activos.
3. `GET /barbershops/{slug}/services/{serviceId}/barbers` — `operationId: listPublicBarbers`. Barberos con asignación vigente a ese servicio.
4. `GET /barbershops/{slug}/services/{serviceId}/barbers/{barberId}/availability` — `operationId: listPublicAvailability`. Proyecta los inicios públicos válidos (jornada efectiva, excepciones, festivos, bloqueos vigentes y citas que ocupan agenda; RN-DIS-01 a RN-DIS-07). Responde `200` con `slots: []` — nunca un error distinto — cuando `serviceId`/`barberId` no tienen forma de UUID, son ajenos a la barbería, están inactivos o sin asignación vigente, o simplemente no hay franjas: la causa no se distingue (mismo criterio que el 404 del punto 1). Es de solo lectura: no reserva ni crea nada (RN-DIS-03).

Leer la descripción completa de cada operación en `api/openapi/paths/public-booking.yaml` antes de implementar el parseo; ahí están los esquemas de petición/respuesta exactos, no los adivines.

## Datos de prueba verificados en vivo (2026-10-22)

Instancia local efímera, solo para que Odin valide su adaptador mientras desarrolla. **No es una URL estable**: el propietario de `barberia` debe indicar la URL base definitiva cuando Odin integre esto contra un ambiente real.

- `BASE_URL` de prueba: `http://localhost:8082`
- `slug`: `barberia-demo` (barbería "Barbería Demo NAVA")
- `serviceId` ("Corte y barba", 60 min): `d0000021-0000-4000-8000-000000000002`
- `barberId` #1 ("Barbero Uno"): `d0000011-0000-4000-8000-000000000001`
- `barberId` #2 ("Barbero Dos"): `d0000011-0000-4000-8000-000000000002`

Se verificó en vivo que los cuatro endpoints responden `200` con datos reales, incluida disponibilidad con franjas (`slots`) no vacías para ambos barberos.

## Trabajo requerido

1. Implementar el cliente HTTP con los cuatro `GET` del contrato exacto, deserializando sus respuestas documentadas (no inferidas).
2. Implementar el manejo de los dos casos "vacío, no error": `404` uniforme en el paso 1, `200` con `slots: []` en el paso 4. El adaptador no debe traducir ninguno de los dos en una excepción genérica que oculte cuál fue.
3. Probar el adaptador contra la instancia indicada arriba (o la que el propietario de `barberia` levante en el momento de la prueba) para los dos `barberId` del tenant demo.
4. No commitear ni mencionar la URL local efímera como si fuera definitiva en el código de Odin; parametrizarla por configuración.

## Pruebas y evidencia

- Prueba de integración del adaptador contra una instancia NAVA real (no contra un mock de forma inventada), cubriendo: barbería existente, barbería inexistente (404 uniforme), servicio/barbero sin franjas (200 con `slots: []`), y al menos un barbero con franjas reales.

## Documentación y trazabilidad

- Si Odin decide fijar esta integración de forma permanente, debe documentar en su propio repositorio la URL base de producción/staging de NAVA que usará y quién la provee.
- Este archivo no se referencia desde la matriz de trazabilidad de `barberia` porque no cambia su código; se conserva aquí solo para que la conversación que lo originó no se pierda.

## Verificación final

```text
N/A en este repositorio: no hay comandos de barberia que verificar. La verificación vive en el repositorio de Odin.
```

## Git y PR

- No aplica en `barberia`: este prompt no abre rama ni PR en `bcaceres19/barberia`.
- El commit/PR del adaptador `NavaHttpReader` ocurre en el repositorio propio de Odin, con sus propias convenciones.
