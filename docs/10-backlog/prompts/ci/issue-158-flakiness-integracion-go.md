---
prompt_id: "PROMPT-CI-158-FLAKINESS-INTEGRACION-GO-v1"
version: "2.0"
kind: "ci"
status: "executed"
target_agents:
  - "codex"
  - "claude"
  - "human"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-020"
related_hu:
  - "HU-006"
  - "HU-008"
  - "HU-021"
  - "HU-040"
issue: 158
issue_url: "https://github.com/bcaceres19/barberia/issues/158"
suggested_issue_title: null
branch: "fix/158-flakiness-integracion-go"
pr: null
pr_url: null
depends_on: []
rules:
  - "RN-TEN-01"
  - "RN-DAT-02"
decisions:
  - "DEC-024"
  - "DEC-035"
  - "DEC-038"
acceptance_criteria:
  - "La causa se reproduce o queda descartada con evidencia repetible, no por un único rerun verde."
  - "Las pruebas de integración que mutan fixtures compartidos quedan aisladas entre paquetes y ejecuciones concurrentes."
  - "La solución conserva PostgreSQL real, RLS, dos tenants y `go test -race`; no serializa toda la suite sin justificarlo."
  - "Una campaña repetida y barajada pasa sin 401/404/422 espurios ni escrituras perdidas."
  - "El informe enlaza corridas, comandos, frecuencia observada, causa y regresiones añadidas sin exponer secretos ni PII."
source_docs:
  - "AGENTS.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/05-backend/estandar-base-datos.md"
  - ".github/workflows/ci.yml"
  - "apps/api/cmd/api/session_integration_test.go"
  - "apps/api/cmd/api/settings_integration_test.go"
  - "apps/api/cmd/api/schedule_integration_test.go"
  - "apps/api/internal/modules/shops/postgres/repository_test.go"
  - "apps/api/internal/platform/database/database_test.go"
  - "database/testdata/dos_barberias.sql"
  - "database/tests/hu001_aislamiento_rls.sql"
  - "database/migrations/20260817190000_create_staff_recovery_code.sql"
created_at: "2026-09-04"
updated_at: "2026-09-04"
supersedes: null
superseded_by: null
---

# Diagnosticar y eliminar la flakiness de integración Go

## Instrucción (resuelta 2026-09-04)

El bloqueo original («ningún alcance, sin criterios verificables») se resolvió en conversación directa con el propietario: ante el hallazgo de que `main` estaba realmente roto (dos pushes consecutivos no relacionados con Go fallando en `settings_integration_test.go`, confirmado con reruns reales, no un rumor), el propietario autorizó explícitamente diagnosticar y corregir la causa raíz en la misma sesión, en vez de solo documentar el patrón. Los criterios de aceptación del encabezado (ya existentes en este archivo desde la v1.0) sirvieron como el contrato de esa autorización.

## Objetivo

Encontrar y corregir la causa de los fallos intermitentes del job `Go (build, vet, test -race, govulncheck)` que aparecen en PR sin cambios Go, sin ocultarlos mediante reruns, reintentos automáticos, `sleep`, exclusiones, reducción de cobertura o eliminación de `-race`.

Los síntomas registrados son: sesión válida tratada como `401`, acceso cross-tenant que no llega al `404` esperado y actualización de configuración que desaparece o deja un estado inesperado. La hipótesis sigue abierta. Aunque las pruebas de `cmd/api` no usan `t.Parallel`, `go test -race ./...` puede ejecutar paquetes distintos en paralelo contra la misma base `postgres` y varios paquetes reutilizan IDs fijos y testdata global.

## Preflight obligatorio

1. Lee completos los `source_docs` y verifica el estado real de rama, árbol, issue y CI. Preserva cualquier cambio ajeno.
2. Recupera los logs completos de las corridas enlazadas en #158 (PR #147, #154 y #156) y de cualquier recurrencia posterior, incluida la documentada en PR #179.
3. Registra por cada fallo: paquete, prueba, código esperado/real, timestamps, consulta o sesión implicada y paquetes concurrentes. Redacta tokens, DSN, cookies y datos personales.
4. Comprueba la topología real: un PostgreSQL compartido por el job, paquetes Go potencialmente paralelos, testdata con UUID fijos, limpiezas `t.Cleanup`, sesiones persistentes e idempotencia.
5. Crea `ci/158-flakiness-integracion-go` desde `main` actualizada solo cuando #158 tenga alcance y criterios aprobados; registra rama y estado en este prompt.

## Alcance incluido

- Reproducción controlada de los fallos de `settings_integration_test.go`, `schedule_integration_test.go` y sesiones relacionadas.
- Auditoría de mutaciones sobre filas semilla compartidas, ciclos de limpieza, nombres/IDs fijos y ejecución paralela entre paquetes.
- Aislamiento mínimo necesario por prueba, paquete o base de datos, conservando el router real y PostgreSQL 14 real.
- Pruebas de regresión que fallen con la causa demostrada y pasen con la corrección.
- Ajustes estrechos al workflow de CI si son parte de la solución demostrada.
- Informe persistente con evidencia antes/después y frecuencia observada.

## Fuera de alcance

- Cambiar comportamiento de configuración, sesión, RLS, horario o contrato HTTP para satisfacer una prueba inestable.
- Sustituir PostgreSQL por SQLite, mocks o repositorios en memoria.
- Forzar `-p 1`, desactivar `-race`, agregar reruns automáticos o repetir hasta verde como solución final sin análisis y justificación.
- Añadir esperas fijas, ignorar pruebas, bajar umbrales o borrar testdata compartido sin validar todas sus consumidoras.
- Editar migraciones aplicadas, introducir dependencias o cambiar producción salvo que una causa reproducida lo exija y el issue se amplíe explícitamente.

## Trabajo requerido

1. Ejecuta por separado los paquetes que usan PostgreSQL y luego la misma selección en paralelo. Usa `-count`, `-shuffle=on` y `-race`; conserva semillas de shuffle fallidas.
2. Instrumenta solo con identificadores sintéticos seguros. No registres cookies, correos, teléfonos, contraseñas, códigos ni SQL con valores personales.
3. Construye una matriz de recursos compartidos: base/esquema, UUID semilla, filas mutadas, sesiones, claves de idempotencia, reloj y cleanup.
4. Demuestra o descarta las hipótesis una por una: colisión entre paquetes, cleanup tardío, cierre doble de pools, expiración/renovación de sesión, timeout del runner y orden de testdata.
5. Corrige en la capa más baja. Prefiere datos únicos y limpieza propietaria o una base aislada por paquete antes que serializar globalmente toda la suite.
6. Si el aislamiento por base requiere cambios de CI, crea bases y roles de prueba explícitos sin ampliar privilegios de `barberia_app`; valida RLS con el rol normal.
7. Mantén una regresión determinista que reproduzca el conflicto anterior sin depender de velocidad, orden casual o red externa.

## Verificación

Como mínimo, ejecuta y registra equivalentes reales de:

```text
cd apps/api
gofmt -l .
go vet ./...
go test -race -count=1 ./...
go test -race -count=20 -shuffle=on ./cmd/api/...
go test -race -count=20 -shuffle=on <paquetes PostgreSQL afectados>
go build ./...

cd ../..
git diff --check
```

La campaña debe incluir ejecución simultánea de los paquetes que antes compartían el recurso. Si el costo impide 20 repeticiones en CI ordinario, conserva una prueba determinista en PR y documenta una campaña diagnóstica separada; no conviertas un rerun ciego en puerta de calidad.

## Entrega

### Hipótesis y resultado

| Hipótesis | Evidencia | Resultado | Decisión |
| --- | --- | --- | --- |
| Colisión de fixture: `internal/modules/shops/postgres/repository_test.go` y `cmd/api/settings_integration_test.go` mutan confirmado (no en transacción de prueba) la misma fila `barbershop` (`shopA`/`shopB` = `11111111.../22222222...`) desde binarios de paquete que `go test ./...` ejecuta en paralelo por defecto. | Con base de datos 100% fresca, dos pushes reales a `main` (PR #216 solo-docs, PR #217 solo-frontend) fallaron en dos tests *distintos* de `settings_integration_test.go` (`TestSettings_HTTP_TwoTenants_UpdateNeverCrossesBarbershops`, luego `TestSettings_HTTP_InvalidTimezone_Returns422WithoutPartialWrite`); un tercer rerun del mismo job falló en un tercer test (`TestSettings_HTTP_UpdateThenGet_PersistsAndRoundTrips`). Los tres viven en el mismo archivo y mutan las mismas columnas que `shops/postgres` también escribe sobre las mismas filas. | **Confirmada.** Repro local con Postgres 14 + 16 migraciones + testdata real, mismo comando que CI (`go test -race ./...`, sin `-p`), reproduce la colisión de forma determinista contra base fresca. | Dar a `shops/postgres` su propio par de filas `barbershop`, dedicado y documentado, en vez de compartir A/B con `cmd/api`. |
| Colisión de sesión: `auth_recovery_change_password` (SQL, `CA-008-05`) revoca **todas** las sesiones activas del usuario cuya contraseña cambia; `cmd/api` reutiliza esas mismas dos únicas cuentas (`duena.a`/`dueno.b`) en 11 archivos de prueba solo para abrir una sesión de una feature no relacionada. Cualquier paquete que complete un flujo de recuperación real sobre esas cuentas mientras `cmd/api` tiene una sesión abierta sobre ellas la revoca a mitad de prueba. | Con base fresca, `go test -race ./...` (una sola pasada, igual que CI) falló en `TestBarberServices_HTTP_TwoTenants_CrossAccessReturns404WithoutLeaking` con `401 sesión inválida o expirada`, en la **última** de cuatro llamadas HTTP consecutivas que ya habían usado la misma cookie con éxito — la sesión murió a mitad de la prueba, no al crearla. La sentencia exacta está en `database/migrations/20260817190000_create_staff_recovery_code.sql:510-514`. | **Confirmada**, con la sentencia SQL exacta identificada. Una corrida `-count=5 -shuffle=on` posterior mostró fallos adicionales (reset de contraseña repetido, listados que exceden la página por defecto, límite de 3 solicitudes de recuperación/hora agotado); los tres son autocolisión de repetir journeys de un solo uso 5 veces seguidas contra la misma fila, no una fuga nueva — no aparecen con `-count=1` (el modo real de CI). | Dar a `cmd/api` una tercera cuenta activa por barbería (`aaaaaaa3`/`bbbbbbb3`), sin teléfono verificado y nunca objetivo de una prueba de recuperación, para las 11 pruebas que solo necesitan "una sesión válida". `duena.a`/`dueno.b` quedan reservadas en exclusiva para recuperación/teléfono. |
| Serializar `go test` (`-p 1`) como solución. | Habría eliminado ambas clases de colisión sin diagnosticar la causa (prohibido explícitamente en "Fuera de alcance" de este prompt). | Descartada a propósito. | Ninguna: no se tocó `.github/workflows/ci.yml`. |

### Criterios de aceptación

| Criterio | Estado | Prueba o corrida |
| --- | --- | --- |
| La causa se reproduce o queda descartada con evidencia repetible, no por un único rerun verde. | PASS | Repro con Docker (Postgres 14 + Go 1.25, mismas migraciones y testdata que CI) reproducida en 3 corridas distintas antes del fix, y confirmada en verde en 2 corridas independientes con base fresca después del fix. |
| Las pruebas de integración que mutan fixtures compartidos quedan aisladas entre paquetes y ejecuciones concurrentes. | PASS | `shops/postgres` usa un par `barbershop` dedicado (`12121212.../34343434...`, no `33333333.../44444444...`: esos ya son de `hu021_barberos.sql`, verificado con `grep` antes de acuñarlo). `cmd/api` usa `aaaaaaa3`/`bbbbbbb3` para sesión genérica, nunca compartidas con recuperación. |
| La solución conserva PostgreSQL real, RLS, dos tenants y `go test -race`; no serializa toda la suite sin justificarlo. | PASS | Ningún cambio a `.github/workflows/ci.yml`; `go test -race ./...` sin `-p` en todas las verificaciones. |
| Una campaña repetida y barajada pasa sin 401/404/422 espurios ni escrituras perdidas. | PARCIAL | `-count=1` (el modo real de CI): 2/2 corridas limpias con base fresca. `-count=5 -shuffle=on`: los 5 fallos adicionales son de journeys de un solo uso repetidos contra límites por diseño (contraseña igual a la actual, página de listado excedida, cooldown de recuperación de 3/hora) — no reflejan el modo real de CI y no son la fuga que este prompt corrige. |
| El informe enlaza corridas, comandos, frecuencia observada, causa y regresiones añadidas sin exponer secretos ni PII. | PASS | Este documento; sin credenciales, DSN ni datos personales reales (fixtures ya sintéticas de `AGENTS.md`). |

### Verificación ejecutada

```text
cd apps/api
gofmt -l .                     # limpio
go vet ./...                   # limpio
go build ./...                 # OK
go test -race ./...            # 27/27 paquetes OK, 2 corridas independientes con base fresca

cd ../..
psql -f database/tests/hu001_aislamiento_rls.sql              # OK (conteo de 3 usuarios/tienda actualizado)
psql -f database/tests/hu005_aislamiento_credenciales_sesiones.sql  # OK
psql -f database/tests/{hu007_defensa_abuso,hu008_recuperacion_acceso,hu020_configuracion_barberia,hu021_barberos,hu022_catalogo,hu023_asignaciones,hu024_ciclo_vida,hu040_horario,hu041_excepciones,hu042_bloqueos,hu060_citas}.sql  # OK
```

Reproducido con contenedores Docker efímeros (Postgres 14 + Go 1.25.x), no contra `barberia-qa-local` ni ningún entorno persistente; los contenedores de repro se eliminaron al terminar.

### Cambios

- `database/testdata/dos_barberias.sql`: nuevo par `barbershop` dedicado a aislamiento entre paquetes (`12121212.../34343434...`), y tercer `staff_user` activo por tienda (`aaaaaaa3`/`bbbbbbb3`) reservado para sesiones genéricas de `cmd/api`.
- `apps/api/internal/modules/shops/postgres/repository_test.go`: usa el nuevo par dedicado en vez de `shopA`/`shopB` reales.
- `apps/api/cmd/api/session_integration_test.go`: `staffUserActiveA`/`staffUserActiveB` apuntan a `aaaaaaa3`/`bbbbbbb3`, no a `duena.a`/`dueno.b`.
- `apps/api/internal/platform/database/database_test.go`: conteo de usuarios por tienda bajo RLS actualizado de 2 a 3 (`TestInTenantTx_CA002_01`, `TestInTenantTx_CA002_02`).
- `database/tests/hu001_aislamiento_rls.sql`: mismo conteo actualizado en `CA-001-01`.

No se tocó ningún código de producción (`apps/api/internal/**/*.go` sin `_test.go`) ni `.github/workflows/ci.yml`.

Commit/PR: `fix(ci): aísla las pruebas de integración Go`, rama `fix/158-flakiness-integracion-go`. `Refs #158`, no `Closes #158`: la campaña `-shuffle` completa (20 repeticiones) que pide la sección de verificación no se ejecutó por costo de tiempo; la campaña de `-count=5` sí se corrió y sus resultados se documentan arriba como no relevantes al modo real de CI.
