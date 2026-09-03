---
prompt_id: "PROMPT-TEST-QA-PLATAFORMA-B0-B3-LUNA-v2"
version: "2.0"
kind: "test"
status: "executed"
executed_at: "2026-09-02"
result: "INCONCLUSA"
result_file: "2026-09-02-qa-integral-b0-b3-luna-high-v2-report.md"
target_agents:
  - "codex"
target_model: "gpt-5.6-luna"
reasoning_effort: "high"
repository: "bcaceres19/barberia"
base_branch: "main"
source_revision_floor: "495edd0"
primary_hu: null
related_hu:
  - "HU-001"
  - "HU-002"
  - "HU-003"
  - "HU-004"
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
issue: null
issue_url: null
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on:
  - "B0-B3 implementados e integrados en main hasta HU-065"
  - "Fundaciones y pantallas NAVA integradas mediante PR #133, #136, #139, #142, #145, #154, #161 y #163"
  - "Regresiones cerradas mediante PR #127, #147, #150, #156 y #159"
  - "Adaptador y prueba manual OTP por correo integrados mediante PR #130 y documentación hasta #165"
  - "Seguimientos abiertos que no deben confundirse con defectos nuevos: #86, #98, #100, #107, #111, #116, #120 y #123"
rules:
  - "RN-RES-01"
  - "RN-RES-02"
  - "RN-RES-03"
  - "RN-SER-01"
  - "RN-SER-02"
  - "RN-SER-03"
  - "RN-SER-04"
  - "RN-DIS-04"
  - "RN-DIS-05"
  - "RN-DIS-06"
  - "RN-DIS-07"
  - "RN-BLQ-01"
  - "RN-BLQ-02"
  - "RN-BLQ-03"
  - "RN-BLQ-04"
  - "RN-CIT-01"
  - "RN-CIT-02"
  - "RN-CIT-03"
  - "RN-CON-01"
  - "RN-CON-03"
  - "RN-CON-05"
  - "RN-HIS-01"
  - "RN-HIS-02"
  - "RN-TEN-01"
  - "RN-DAT-01"
  - "RN-DAT-02"
  - "RN-IDE-01"
decisions:
  - "DEC-002"
  - "DEC-004"
  - "DEC-007"
  - "DEC-014"
  - "DEC-016"
  - "DEC-019"
  - "DEC-020"
  - "DEC-024"
  - "DEC-026"
  - "DEC-033"
  - "DEC-034"
  - "DEC-035"
  - "DEC-036"
  - "DEC-037"
  - "DEC-039"
  - "DEC-043"
  - "DEC-050"
  - "DEC-051"
  - "DEC-055"
  - "DEC-056"
  - "DEC-057"
  - "DEC-058"
  - "DEC-060"
  - "DEC-061"
  - "DEC-062"
  - "DEC-063"
  - "DEC-064"
  - "DEC-065"
  - "DEC-066"
  - "DEC-067"
  - "DEC-068"
  - "DEC-069"
  - "DEC-070"
  - "DEC-071"
  - "DEC-072"
  - "DEC-073"
  - "DEC-074"
  - "DEC-075"
  - "DEC-076"
  - "DEC-077"
acceptance_criteria:
  - "CA-001-* a CA-012-*"
  - "CA-020-* a CA-024-*"
  - "CA-040-* a CA-042-*"
  - "CA-060-* a CA-065-*"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/02-requisitos/estados-citas.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/06-api/estandar-openapi.md"
  - "docs/07-calidad/01-metodologia-y-uso.md"
  - "docs/07-calidad/02-checklist-transversal.md"
  - "docs/07-calidad/03-checklist-autenticacion-sesion.md"
  - "docs/07-calidad/04-checklist-modulos-catalogo.md"
  - "docs/07-calidad/05-matriz-combinaciones.md"
  - "docs/07-calidad/06-plantilla-hallazgo.md"
  - "docs/10-backlog/prompts/test/2026-08-25-qa-manual-b0-b1-chrome-mcp-report.md"
  - "docs/10-backlog/prompts/test/issue-86-otp-correo-resend-report.md"
  - "apps/api/README.md"
  - "apps/web/README.md"
  - "apps/web/e2e"
  - ".github/workflows/ci.yml"
created_at: "2026-09-02"
updated_at: "2026-09-02"
execution_report: null
reusable: true
supersedes: "PROMPT-TEST-QA-B2-B3-CHROME-v1"
superseded_by: null
---

# QA integral y combinatoria de NAVA B0-B3 con Codex Luna high

## Configuración del agente

- Modelo exacto: `gpt-5.6-luna`.
- Esfuerzo de razonamiento: `high`.
- Usa un único agente coordinador. La campaña comparte una base y datos ordenados; no paralelices recorridos que muten el mismo tenant ni uses subagentes sobre el mismo estado.
- Usa el navegador integrado de Codex con sesión real cuando esté disponible. Playwright en modo normal o headed es la alternativa para la suite automatizada y la evidencia que el navegador integrado no pueda capturar.
- El prompt ya expresa objetivo, límites, evidencia, criterios de éxito y formato. No reemplaces estos criterios por una instrucción vaga como “piensa más” o “prueba todo”.

## Instrucción para el agente

Ejecuta una campaña de pruebas completa sobre la plataforma **realmente implementada** en `main`: B0, B1, B2 y B3 hasta `HU-065`. Combina una línea base automatizada, los recorridos Playwright existentes, exploración manual por pares, pruebas entre dos tenants, condiciones de red/sesión y regresiones dirigidas a los defectos ya corregidos.

Este prompt es reutilizable y **no autoriza cambios en archivos versionados, producto, contrato, migraciones, documentación, issues o pull requests**. Sí autoriza acciones locales reversibles necesarias para probar: leer archivos y logs, levantar servicios locales, crear una base de prueba desechable, aplicar las migraciones en esa base, sembrar datos ficticios, ejecutar pruebas, controlar el navegador y guardar evidencia temporal fuera del repositorio.

Si encuentras un defecto:

1. conserva la evidencia antes de seguir;
2. confirma si es reproducible sin alterar el producto;
3. clasifícalo y repórtalo;
4. no lo corrijas, no cambies una expectativa para hacerlo pasar y no abras un issue sin autorización expresa.

## Objetivo observable

Entregar un informe que permita responder, con evidencia y sin inferencias:

1. qué SHA, entorno, motores y combinaciones se probaron;
2. si las suites automatizadas y los recorridos visibles B0-B3 pasan sobre PostgreSQL real;
3. si las fronteras críticas —tenant, zona horaria, intervalos, concurrencia, idempotencia, sesión y datos personales— se conservan;
4. si cada defecto cerrado listado en la matriz de regresión permanece corregido;
5. qué escenarios fallaron, quedaron bloqueados o no aplican, con causa diferenciada;
6. qué capacidades están fuera de alcance y por tanto no deben reportarse como bugs.

No uses “todo bien”, “parece correcto” ni el estado verde de CI como sustituto de una ejecución real.

## Límites de autonomía y seguridad

### Puedes hacer sin pedir confirmación

- inspeccionar Git, código, contrato, pruebas, logs y documentación;
- ejecutar comandos locales no destructivos de formato, lint, tipos, build y pruebas;
- iniciar y detener procesos locales creados por esta campaña;
- usar una base PostgreSQL **dedicada y desechable** para QA;
- crear datos ficticios reconocibles por el prefijo `QA-LUNA-<fecha-hora-UTC>`;
- usar dos contextos de navegador aislados para los tenants A y B;
- capturar screenshots, trazas y respuestas redactadas en un directorio temporal fuera del repositorio.

### Requiere detenerse o pedir autorización

- probar contra piloto o producción;
- usar, copiar, borrar o modificar datos reales;
- pegar secretos, OTP, cookies o credenciales en el chat, reporte, consola visible o archivo versionado;
- enviar correo o WhatsApp real si el operador no preparó explícitamente un buzón/canal de prueba y sus variables seguras;
- editar código, snapshots, migraciones, documentación o umbrales;
- abrir/comentar/cerrar issues, crear ramas, commits o PR;
- destruir una base, contenedor o directorio que la campaña no haya creado y validado como propio.

Ante una posible fuga entre tenants, corrupción, 500 reproducible o pérdida de datos, detén las mutaciones del flujo afectado, guarda evidencia redactada y reporta el hallazgo inmediatamente. Puedes continuar con módulos independientes solo si el entorno sigue siendo seguro.

## Preflight obligatorio

### 1. Identidad exacta de la ejecución

Registra antes de probar:

```text
fecha/hora UTC
git branch --show-current
git rev-parse HEAD
git status --short
SO y arquitectura
versiones de Go, Node, pnpm, PostgreSQL, Atlas y navegadores
```

- Debes estar en `main` limpia o en una copia de solo lectura del SHA que el operador indicó.
- El SHA debe contener al menos `495edd0`; si se prueba una revisión posterior, registra el SHA real y revisa sus cambios desde ese piso antes de reutilizar esta matriz.
- Si el árbol tiene cambios ajenos, no los sobrescribas, guardes en stash ni reviertas. Informa el bloqueo o trabaja en otra copia autorizada.
- No hagas `pull`, `reset`, `checkout`, rebase ni limpieza automática.

### 2. Lectura y reconciliación de fuentes

Lee completamente cada `source_doc`. Luego:

1. inventaría las rutas reales de `apps/web/src/app/router/index.ts` y los módulos exportados;
2. inventaría los `test(...)` actuales de `apps/web/e2e/*.spec.ts`;
3. compara los límites de cada formulario con sus validadores actuales;
4. confirma en `git log` que los fixes/PR de la sección “Regresiones obligatorias” están en el SHA probado;
5. registra contradicciones o documentación desactualizada como hallazgo documental, sin reinterpretar el comportamiento.

Si una fuente normativa contradice este prompt, manda la fuente y el escenario queda `BLOQUEADO-POR-CONTRADICCIÓN` hasta documentar la duda. No inventes el resultado esperado.

### 3. Entorno aislado

Levanta el stack mediante las recetas vigentes de `apps/api/README.md`, `apps/web/README.md` y `.github/workflows/ci.yml`:

- PostgreSQL real, no SQLite ni mocks;
- migraciones Atlas aplicadas desde cero en una base dedicada;
- rol de aplicación real sin `BYPASSRLS` y rol worker cuando la suite lo requiera;
- API Go local;
- frontend Vue local;
- dos tenants A/B y dos sesiones de navegador realmente separadas.

No hardcodees un número histórico de migraciones. Registra lo que informe `atlas migrate status` para el SHA probado.

### 4. Datos mínimos controlados

Crea en la base desechable, por interfaces públicas siempre que sea posible:

| Entidad | Tenant A | Tenant B |
| --- | --- | --- |
| Usuario autenticable | 1 activo y, si la receta lo soporta, 1 inactivo | 1 activo |
| Barberos | 2 con nombres distintos; 2 con nombre idéntico para la prueba de ambigüedad | 1 |
| Servicios | 1 activo asignado a ambos barberos, 1 activo asignado a uno, 1 inactivo | 1 activo |
| Horario | jornada partida y un tramo nocturno controlado | jornada simple |
| Excepciones/bloqueos | excepción cerrada, especial, bloqueo puntual y serie | 1 bloqueo |
| Turnos | contiguos, uno futuro confirmado, uno que cruza medianoche | 1 futuro confirmado |

Usa únicamente nombres, correos y teléfonos ficticios. En capturas y tablas identifica los tenants como A/B y redacta IDs, correos, teléfonos, cookies, tokens de versión, claves de idempotencia y `request_id` salvo los últimos caracteres necesarios para correlación.

### 5. Condiciones previas y salida segura

- Confirma espacio suficiente para screenshots/trazas temporales.
- Confirma que ninguna URL apunta a producción.
- Confirma que la base es desechable antes de sembrar.
- Guarda la evidencia en `%TEMP%/nava-qa-<timestamp>` o equivalente, nunca dentro del repositorio.
- Al terminar, detén solo los procesos creados por la campaña. No elimines recursos si no puedes demostrar que son los exactos creados por ella.

## Alcance incluido

### Automatizado

- OpenAPI, Go, PostgreSQL/RLS/concurrencia, frontend, componentes y build.
- Todos los specs Playwright existentes en `apps/web/e2e`.
- Chromium escritorio/móvil como puerta mínima; Firefox y WebKit como campaña previa a versión.
- Repetición diagnóstica acotada de cualquier fallo para distinguir fallo estable de intermitente, sin usar retries para declararlo verde.

### Manual y exploratorio

- acceso, reto, recuperación, sesión, shell y navegación privada;
- configuración, barberos, servicios, ciclo de vida y asignaciones;
- horarios, festivos, excepciones y bloqueos implementados en UI;
- creación manual, agenda diaria, navegación por fecha, detalle/historial y reprogramación;
- responsive, teclado, foco, zoom 200 %, reduced motion, estados asíncronos y objetivos táctiles;
- red normal/lenta/offline, doble acción, atrás/adelante, F5, dos pestañas y sesión vencida;
- aislamiento A/B en listas, selectores, URLs y endpoints;
- verificación dirigida de fixes y fases NAVA integradas.

## Fuera de alcance: no son defectos de esta campaña

- B4 reserva pública, B5 notificaciones/recordatorios y B6 operación/piloto aún no implementados.
- T3 edición general, cancelación, completar, `no_show`, corrección de estado y cierre automático posteriores a `HU-065`.
- Vista consolidada multi-barbero: `DEC-074` exige un barbero a la vez.
- Estado `in_progress`: “Ahora” es una señal visual, no un estado de turno.
- Ingresos, caja, reportes, inventario, cuenta de cliente, modo oscuro y personalización por tenant.
- Fuentes Instrument self-hosted mientras no exista la entrega específica aprobada; los fallbacks vigentes son válidos.
- Sexto ítem temporal “Servicios por barbero” en el dock hasta la Fase 5.
- En `HU-042`, controles de `date_list`, excepciones de serie y edición por `scope`; están trazados en el issue #100 y no están conectados a la UI.
- Envío real por WhatsApp. El reto de `HU-007` se prueba con el mecanismo local documentado salvo que exista una autorización separada para proveedor real.
- Repetir el correo OTP real con Resend sin buzón y variables seguras del operador. La prueba existente está documentada; la falta de credenciales no es un bug.
- Corrección de cualquier hallazgo o actualización de documentación normativa.

## Estrategia de ejecución por puertas

Cada puerta termina en `PASS`, `FAIL`, `BLOCKED` o `NOT-APPLICABLE`. Un `BLOCKED` nunca cuenta como `PASS`.

### G0 — Integridad del entorno y contrato de la campaña

Debe pasar antes de mutar datos:

- revisión/identidad de Git registrada;
- base y URLs confirmadas como locales y desechables;
- dos tenants disponibles;
- fuentes y límites reconciliados;
- secretos ausentes del repositorio y de la salida capturable.

Si G0 falla, entrega diagnóstico y no inventes resultados del resto.

### G1 — Línea base automática

Ejecuta desde el directorio correspondiente y conserva comando, código de salida, duración y resumen. Adapta la sintaxis al SO, no el significado:

```text
# raíz
pnpm run openapi:check-config
pnpm run openapi:lint
pnpm run openapi:bundle

# apps/api
gofmt -l .
go vet ./...
go build ./...
go test -race -count=1 ./...

# apps/web
pnpm run format
pnpm run lint
pnpm run typecheck
pnpm run test:unit
pnpm run build
```

Para Atlas, reproduce los pasos no destructivos del job de CI contra la base desechable: `validate`, `apply` desde vacío, segundo `apply` sin cambios y `status` sin pendientes. No edites migraciones ni `atlas.sum`.

Después de comandos que generen artefactos, ejecuta `git status --short` y `git diff --check`. Si aparece una diferencia versionada, no la reviertas ni la aceptes: repórtala como divergencia reproducible del comando.

Ante un fallo automatizado:

1. guarda el primer log;
2. repite solo la prueba exacta hasta 3 ejecuciones adicionales con `-count=1` o equivalente;
3. clasifica `estable N/N`, `intermitente N/4` o `no reproducible 1/4`;
4. no uses el rerun verde para borrar el fallo inicial ni cambies la prueba.

### G2 — E2E Chromium obligatorio

Con el stack real y datos aislados:

```text
pnpm exec playwright test --project=chromium-desktop --project=chromium-mobile
```

- Cero `.only`, `.skip` nuevo, snapshot actualizado o aserción relajada.
- Cada fallo conserva screenshot, trace y mensaje.
- Ejecuta en modo headed el caso fallido si hace falta entender el comportamiento visible; el headed no reemplaza el resultado de la suite.

### G3 — Compatibilidad antes de versión

```text
pnpm exec playwright test --project=firefox --project=webkit
```

Un motor ausente o no instalable queda `BLOCKED-ENVIRONMENT`, no `PASS`. Un fallo solo de motor sigue siendo hallazgo de producto hasta demostrar una incompatibilidad externa documentada.

### G4 — Recorrido manual nominal B0-B3

Reproduce los specs existentes por sus títulos y aserciones actuales, no desde memoria. Como mínimo:

| ID | Módulo/HU | Recorrido nominal que debe completarse |
| --- | --- | --- |
| N01 | B0 auth | login válido, no enumeración, cinco intentos normales y sexto con reto, código incorrecto y rama local correcta si hay captura segura |
| N02 | B0 recovery | solicitar, verificar, cambiar contraseña, revocar sesiones y entrar con la nueva; reuso/vencimiento rechazados |
| N03 | B0 session/shell | ruta privada conserva destino, F5/cierre-reapertura conserva sesión, logout revoca, sesión expirada muestra “Tu sesión venció” |
| N04 | B1 settings | editar nombre/contacto/zona, cabecera inmediata, F5 persistente, zona inválida sin escritura parcial |
| N05 | B1 staff | crear, renombrar, paginar, nombres duplicados/unicode/HTML literal, aislamiento B |
| N06 | B1 catalog | crear/editar/listar, límites, duplicado activo, precio exacto COP, desactivar/reactivar y conservar asignaciones/snapshots |
| N07 | B1 assignment | asignar a dos barberos, persistir, rechazar última asignación activa y aislar B |
| N08 | B2 hours | tramo, jornada partida, contigüidad, solape rechazado, editar/retirar, tramo nocturno |
| N09 | B2 exceptions | festivos on/off, cerrada, especial, fecha duplicada, editar/retirar, precedencia por barbero |
| N10 | B2 blocks | siete tipos disponibles, puntual, serie, rango inválido, persistencia y retiro lógico; reconocer exclusiones #100 |
| N11 | B3 create | crear dentro de la próxima hora y fuera de rejilla; servicio no asignado ausente; cruce y bloqueo rechazados |
| N12 | B3 agenda | barbero obligatorio, hoy/zona, agenda vacía/llena, cambio de barbero, turno nocturno en ambos días |
| N13 | B3 navigation | anterior/siguiente/fecha, F5, atrás/adelante, zona de barbería, respuesta vieja descartada |
| N14 | B3 detail | snapshots/contacto privado/historial, catálogo modificado no reescribe pasado, volver conserva fecha/barbero, ID ajeno/inexistente 404 |
| N15 | B3 reschedule | éxito a otro día, agenda anterior/nueva, un evento; no-op; cruce/bloqueo/versión obsoleta/estado inválido; reintento idempotente |

### G5 — Exploración combinatoria por pares

Ejecuta primero todas las filas de `docs/07-calidad/05-matriz-combinaciones.md`. Luego completa las tres matrices de esta versión. Si al reconciliar el código aparece un valor nuevo, agrégalo al **informe temporal de la ejecución**, no a la documentación versionada, y asegura que cada par relevante aparece al menos una vez.

#### Matriz A — formularios, red y acción

| ID | Flujo | Clase de datos | Viewport | Red | Acción/temporalidad | Resultado mínimo |
| --- | --- | --- | --- | --- | --- | --- |
| PA01 | Login | válido típico | 320 | normal | doble toque | una sola navegación/sesión; foco visible |
| PA02 | Login | correo inexistente | 768 | lenta | envío único | mensaje indistinguible; carga explícita |
| PA03 | Recovery paso 2 | código de 5/7 dígitos o letras | 1280 | normal | F5 y reintento | cliente rechaza; flujo recuperable o reinicio claro |
| PA04 | Settings | email/teléfono vacíos, zona válida | 360 | offline al enviar | volver atrás | formulario conservado; sin escritura parcial |
| PA05 | Settings | límites 120/254/64 exactos | 1280 | lenta | dos pestañas | comportamiento observado sin 500; F5 confirma qué persistió |
| PA06 | Staff | 120 chars unicode/emoji | 320 | normal | doble clic | layout íntegro; resultado sin ambigüedad ni bloqueo |
| PA07 | Service | 120/500, duración 1, precio decimal válido | 768 | normal | crear y F5 | round-trip exacto; sin `$`; moneda real |
| PA08 | Service | 121/501, duración 1441, precio 3 decimales | 1280 | lenta | envío | errores juntos en cliente; no petición de negocio |
| PA09 | Working hour | contiguo al tramo existente | 360 | normal | guardar y editar | contigüidad válida; persistencia exacta |
| PA10 | Exception/block | motivo 200 y rango válido | 768 | offline y reintento | doble toque | una sola entidad; texto conservado |
| PA11 | New appointment | sin teléfono/correo, hora fuera de rejilla | 320 | lenta | salir y volver | creación válida si integridad cumple; estado explícito |
| PA12 | Reschedule | fecha/hora válida | 1280 | offline tras enviar | mismo reintento | un intervalo y un evento de historial |

#### Matriz B — tiempo, integridad y concurrencia

| ID | Operación | Intervalo/estado | Concurrencia | Tenant | Resultado mínimo |
| --- | --- | --- | --- | --- | --- |
| PB01 | Crear turno | contiguo exacto `[fin A = inicio B]` | una pestaña | A | permitido |
| PB02 | Crear turno | cruce de 1 minuto | dos envíos coordinados | A | exactamente uno válido; conflicto controlado |
| PB03 | Crear turno | cruza medianoche y cabe en jornada | una | A | persiste y aparece una vez en cada día intersecado |
| PB04 | Crear turno | sobre bloqueo vigente | una | A | `409`, cero cita/historial nuevo |
| PB05 | Crear bloqueo | sobre cita confirmada | una | A | bloqueo se crea; cita no se mueve/cancela automáticamente |
| PB06 | Desasignar | dos últimas asignaciones | dos pestañas | A | nunca queda servicio activo sin asignación |
| PB07 | Reprogramar | dos destinos incompatibles | dos solicitudes coordinadas | A | una gana; precondición/conflicto seguro para la otra |
| PB08 | Reprogramar | mismo intervalo | reintento misma clave | A | no-op sin historial duplicado |
| PB09 | Leer/mutar recurso | ID válido de A desde sesión B | una | B→A | `404`, sin señal transitoria en UI |
| PB10 | Logout/mutación | sesión revocada en pestaña 1 | pestaña 2 intenta guardar | A | `401` y redirección clara; no mutación |

#### Matriz C — presentación y acceso

Aplica a `/acceso`, `/recuperar-acceso`, `/panel`, `/panel/turnos/nuevo`, `/panel/turnos/{id}`, `/panel/barberia`, `/panel/barberos`, `/panel/servicios`, `/panel/servicios-por-barbero`, `/panel/horarios` y `/panel/bloqueos`.

| ID | Viewport | Estado visible | Entrada | Preferencia | Verificación |
| --- | --- | --- | --- | --- | --- |
| PC01 | 320 | normal con texto límite | táctil | normal | sin scroll horizontal; acción al alcance; dock no tapa contenido |
| PC02 | 360 | diálogo abierto | teclado | reduced motion | foco atrapado/visible; Escape devuelve foco; transición atenuada |
| PC03 | 768 | error recuperable | teclado | normal | mensaje y reintento accesibles; datos no sensibles conservados |
| PC04 | 1024 | carga/actualización | mouse | normal | cambio al patrón escritorio sin contenido duplicado |
| PC05 | 1279 | lista larga | táctil | normal | dock inferior vigente; paginación/CTA alcanzables |
| PC06 | 1280 | éxito con varias filas | teclado | normal | contenido centrado; jerarquía NAVA; acciones en orden lógico |
| PC07 | 1280 + zoom 200 % | diálogo + error | teclado | reduced motion | reflow, no recorte ni control inalcanzable |
| PC08 | 320 | unicode/HTML literal | táctil y teclado | normal | texto literal, sin ejecución, mojibake ni corte multibyte |

En cada pantalla mide objetivos táctiles con el rectángulo real del elemento interactivo o de su `label`, no por inspección visual aproximada. El mínimo es 44×44 px.

### G6 — Regresiones obligatorias de bugs y cambios NAVA

No declares “fix verificado” por encontrar el commit o una prueba unitaria. Ejecuta el caso visible y conserva evidencia.

| ID | Issue / PR | Regresión que debe demostrarse |
| --- | --- | --- |
| R01 | issue #51 / PR #127 | Cada `BaseInput required` muestra **exactamente un** asterisco visual; el nombre accesible no duplica “*”. Muestrea acceso, recuperación, configuración y un diálogo de negocio en 320/1280 px. |
| R02 | #146 / #147 | “Detalle de turno” usa jerarquía H1 real y “Historial” H2, no tamaño de párrafo heredado; mantiene reflow/zoom 200 %. |
| R03 | #149 / #150 | “Barberos”, “Servicios”, “Horarios”, “Servicios por barbero”, “Barbería” y “Horarios y bloqueos” usan la jerarquía/tokens vigentes. En Bloqueos, bordes/radios no caen en `#ccc` ni en una variable inexistente. |
| R04 | #155 / #156 | Servicios muestra `precio + moneda` sin `$` literal, conserva el decimal exacto del contrato, cifras tabulares y fila mínima de 64 px; prueba precio largo en 320 px. |
| R05 | #157 / #159 | En Servicios por barbero, el área real del checkbox+label mide al menos 44×44; clic en el texto alterna, teclado funciona y el rechazo de la última asignación conserva la casilla. |
| R06 | #138 / #139 | Acceso/recuperación muestran wordmark NAVA; título “Accede a NAVA”; una redirección por sesión expirada muestra “Tu sesión venció” y el aviso se limpia al reintentar. |
| R07 | #135 / #136 | Shell privado conserva wordmark+barbería, CTA “Nuevo turno”, dock fijo y `aria-current`; el dock no tapa contenido en 320/360/1024/1279/1280 ni con safe-area simulada. |
| R08 | #141 / #142 | Agenda conserva lista cronológica accesible; la línea temporal aparece solo desde 1024, queda fuera del tab order, posiciona el turno nocturno sin fecha errónea y “Ahora” solo aparece en hoy. |
| R09 | #144 / #145 | Nuevo turno actualiza el resumen previo al CTA para barbero, servicio, persona y fecha/hora; no inventa precio/duración y no usa colores fallback rotos. |
| R10 | #153 / #154 | Filas de Barberos muestran monograma y jerarquía sin perder el nombre completo, acciones, 64 px mínimos ni accesibilidad con nombres duplicados/unicode. |
| R11 | #160 / #161 | Horarios usa jerarquía de fila y cifras tabulares sin alterar creación, edición, retiro, contigüidad ni solapes. |
| R12 | #162 / #163 | Bloqueos muestra ficha y jerarquía de sección sin perder los siete tipos, motivo, rango, retiro ni responsive. |
| R13 | #86 / #130 | Siempre ejecuta las pruebas automatizadas de selección de remitente: local/test permite correo único con Resend completo; piloto/producción sigue exigiendo canal dual. Repite el correo real solo si el operador preparó buzón y secretos de forma segura; de lo contrario marca `NOT-RERUN-AUTHORIZATION` y enlaza el informe existente, sin declarar fallo. |

### G7 — Aislamiento tenant endpoint por endpoint

Para cada familia implementada toma un ID real de A y úsalo desde B. Debe responder `404`, nunca `200`, `403` ni datos parciales:

- sesión/contexto privado;
- configuración de barbería;
- barberos;
- servicios;
- asignaciones servicio-barbero;
- horarios recurrentes;
- calendario/excepciones;
- bloqueos puntuales/series;
- agenda diaria;
- detalle/historial de turno;
- reprogramación.

Además verifica que ningún nombre de A aparece, ni brevemente, en listas o selectores de B durante carga, error, retry, atrás/F5 o respuesta fuera de orden.

### G8 — Privacidad, errores y observabilidad

Provoca de forma controlada `400/401/404/409/422/429`, error de red y un `500` simulado por el mecanismo de prueba existente. Verifica:

- `application/problem+json` y forma RFC 9457 donde corresponda;
- mensaje seguro y accionable;
- `request_id` correlacionable sin datos personales;
- ausencia de contraseña, OTP, cookie, token, email/teléfono completos, nota privada y stack trace;
- reintento seguro y conservación de datos no sensibles;
- ningún éxito falso tras timeout/offline.

No provoques un 500 mediante corrupción manual de datos ni cambios de código.

## Criterios de severidad y triage

Usa la clasificación de `docs/07-calidad/01-metodologia-y-uso.md`:

- **Bloqueante:** fuga tenant, cruce persistido, corrupción/pérdida irreversible, secreto o dato personal expuesto, 500 sin recuperación en un P0.
- **Alto:** dato persistido incorrecto, historial/idempotencia rota, flujo P0 sin salida.
- **Medio:** error/mensaje/teclado/responsive que dificulta pero permite completar.
- **Bajo:** detalle visual/copy sin efecto funcional ni accesible.

Antes de atribuir un fallo al producto, separa:

| Clasificación | Criterio |
| --- | --- |
| `PRODUCT-FAIL` | Resultado reproducible contradice RN/DEC/CA/estándar sobre entorno válido |
| `ENVIRONMENT-BLOCKED` | Falta herramienta, navegador, servicio o variable; no se observó el producto |
| `TEST-DATA-BLOCKED` | Precondición ficticia no pudo crearse sin incumplir límites |
| `KNOWN-GAP` | Capacidad explícitamente fuera de alcance o seguimiento ya trazado |
| `DOC-CONTRADICTION` | Fuentes no permiten definir un esperado único |
| `INTERMITTENT` | El mismo caso alterna éxito/fallo bajo iguales condiciones; conserva frecuencia |

Cada hallazgo usa íntegra la plantilla de `docs/07-calidad/06-plantilla-hallazgo.md` y añade SHA, navegador, viewport, zona, red y frecuencia.

## Evidencia mínima

### Por escenario

Registra:

```text
ID | HU/RN/DEC/CA/fix | precondición | combinación exacta | esperado |
observado | PASS/FAIL/BLOCKED/N-A | evidencia | duración | intento n/N
```

- Screenshot solo cuando demuestra estado visual; no reemplaza aserciones.
- En red: método/path redactado, status, media type, code de problema y `request_id` abreviado.
- En datos: consulta/aserción y conteo, sin exponer DSN ni PII.
- En responsive: captura por pantalla/estado/ancho relevante y medición de scroll (`scrollWidth <= clientWidth`).
- En accesibilidad: recorrido de teclado, foco, nombre/rol, objetivo táctil y resultado axe cuando aplique.
- En concurrencia/idempotencia: número de solicitudes, éxitos, conflictos y filas/eventos finales.

No captures DevTools con cookies, headers `Authorization`, OTP, DSN o variables de entorno visibles.

## Orden recomendado de sesiones

La campaña es larga; ejecútala en sesiones con checkpoint de resultados, no con saltos arbitrarios:

1. S0 — G0 y manifiesto.
2. S1 — G1 automatizado.
3. S2 — G2 Chromium.
4. S3 — B0 auth/sesión/shell + R01/R06/R07/R13.
5. S4 — B1 settings/staff/catalog/asignaciones + matrices A/C + R03/R04/R05/R10.
6. S5 — B2 horarios/excepciones/bloqueos + matrices B/C + R03/R11/R12.
7. S6 — B3 creación/agenda/navegación/detalle/reprogramación + matrices B/C + R02/R08/R09.
8. S7 — G7 tenant, G8 errores/privacidad y dos pestañas.
9. S8 — G3 Firefox/WebKit y cierre.

Dentro de cada sesión, usa los IDs en orden. Si algo queda bloqueado, registra el bloqueo y continúa con los casos independientes; nunca lo omitas de la tabla.

## Criterios de salida

La campaña solo puede concluir `APROBADA` si:

1. G0-G2 pasan; G3 pasa para una candidata a versión o queda explícitamente fuera del objetivo acordado de la ejecución.
2. Todos los recorridos N01-N15 y filas PA/PB/PC aplicables tienen resultado real.
3. R01-R12 pasan. R13 automático pasa; la repetición de correo real puede quedar `NOT-RERUN-AUTHORIZATION` sin reprobar la campaña.
4. Aislamiento A/B produce cero lecturas/mutaciones cruzadas y todos los recursos ajenos se ocultan con `404`.
5. Cero cruces persistidos, duplicados críticos, pérdida de historial, fuga de PII o 500 reproducible.
6. Cero hallazgos Bloqueantes o Altos abiertos.
7. Los hallazgos Medios/Bajos incluyen reproducción y evidencia; no se esconden para aprobar.
8. Ninguna prueba fue ignorada, reescrita, reintentada indefinidamente ni marcada verde sin ejecución.
9. `git status --short` final no muestra modificaciones versionadas creadas por la campaña.

Si una condición falla, el resultado es `NO APROBADA`; si no pudo observarse por una dependencia externa, es `INCONCLUSA`, nunca `APROBADA CON SUPUESTOS`.

## Formato obligatorio del informe final

Entrega en la respuesta final, sin modificar el repositorio:

1. **Resultado ejecutivo:** `APROBADA`, `NO APROBADA` o `INCONCLUSA`, con tres razones máximas.
2. **Manifiesto:** SHA, fecha, stack, DB/Atlas, navegadores, zonas y límites de evidencia.
3. **Puertas:** tabla `G0-G8 | estado | comando/recorrido | evidencia | duración`.
4. **Cobertura funcional:** tabla por bloque/HU y CA cubiertos.
5. **Cobertura combinatoria:** filas PA/PB/PC ejecutadas, pares pendientes y justificación.
6. **Regresiones:** tabla `R01-R13 | issue/PR | resultado | evidencia`.
7. **Aislamiento e integridad:** endpoint/familia, status A→B/B→A, conteo final y evidencia.
8. **Responsive/accesibilidad:** pantalla × ancho × estado × teclado/zoom/axe/objetivo táctil.
9. **Hallazgos:** uno por plantilla completa, ordenados por severidad y reproducibilidad.
10. **Bloqueos y no ejecutados:** motivo exacto; no mezclar con fallos.
11. **Capacidades fuera de alcance reconocidas:** confirma que no se reportaron como bugs.
12. **Próximo paso recomendado:** reportar, pedir decisión o preparar un prompt de fix separado; no ejecutar ese paso.

Termina con una tabla breve:

`Criterio de salida | Estado | Evidencia`

No declares cumplido aquello que no observaste. Si el operador pide persistir el informe o abrir issues, eso requiere autorización separada, issue real y flujo Git conforme a `AGENTS.md`.
