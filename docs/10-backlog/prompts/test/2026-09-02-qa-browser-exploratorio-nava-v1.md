---
prompt_id: "PROMPT-TEST-QA-NAVA-BROWSER-EXPLORATORY-v1"
version: "1.0"
kind: "test"
status: "ready"
target_agents:
  - "codex"
target_model: "gpt-5.6-luna"
reasoning_effort: "high"
repository: "bcaceres19/barberia"
base_branch: "main"
source_revision_floor: "HEAD"
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
issue: null
issue_url: null
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on:
  - "PROMPT-TEST-QA-PLATAFORMA-B0-B3-LUNA-v2"
  - "API Go local, PostgreSQL real desechable y migraciones Atlas aplicadas"
  - "Credenciales ficticias E2E preparadas de forma segura para la barbería A"
rules:
  - "RN-RES-01"
  - "RN-RES-02"
  - "RN-SER-03"
  - "RN-SER-04"
  - "RN-DIS-04"
  - "RN-DIS-05"
  - "RN-DIS-07"
  - "RN-BLQ-03"
  - "RN-BLQ-04"
  - "RN-CIT-01"
  - "RN-CIT-02"
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
  - "DEC-007"
  - "DEC-016"
  - "DEC-019"
  - "DEC-020"
  - "DEC-024"
  - "DEC-039"
  - "DEC-043"
  - "DEC-057"
  - "DEC-058"
  - "DEC-062"
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
  - "CA-009-*"
  - "CA-010-*"
  - "CA-011-*"
  - "CA-012-*"
  - "CA-020-*"
  - "CA-021-*"
  - "CA-022-*"
  - "CA-023-*"
  - "CA-024-*"
  - "CA-040-*"
  - "CA-041-*"
  - "CA-042-*"
  - "CA-060-*"
  - "CA-061-*"
  - "CA-062-*"
  - "CA-063-*"
  - "CA-064-*"
  - "CA-065-*"
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
  - "docs/10-backlog/prompts/test/2026-09-02-qa-integral-b0-b3-luna-high-v2.md"
  - "apps/api/README.md"
  - "apps/web/README.md"
  - "apps/web/e2e"
  - ".github/workflows/ci.yml"
created_at: "2026-09-02"
updated_at: "2026-09-02"
execution_report: "Reconocimiento parcial ejecutado en Codex In-app Browser; backend/API bloqueados en el entorno observado. Ver sección Baseline observado."
reusable: true
supersedes: null
superseded_by: null
---

# QA exploratorio browser-first de NAVA B0-B3

## Instrucción para el agente

Ejecuta una campaña de pruebas visual, funcional y de accesibilidad sobre la plataforma realmente implementada en `main`, usando primero el navegador integrado de Codex y el stack local real. Usa el prompt integral `PROMPT-TEST-QA-PLATAFORMA-B0-B3-LUNA-v2` como matriz normativa y este prompt como su continuación enfocada en observar la UI, recorrerla y encontrar defectos reproducibles.

Este prompt es de solo lectura respecto del repositorio. No autoriza cambios en código, pruebas, snapshots, migraciones, contrato, documentación normativa, issues, ramas, commits ni pull requests. Sí autoriza levantar y detener procesos locales creados por la campaña, usar una base de pruebas desechable creada por la campaña, sembrar datos ficticios y guardar evidencia temporal fuera del repositorio.

Si encuentras un bug:

1. captura la evidencia antes de continuar;
2. repítelo en una sesión limpia o en el mismo estado controlado;
3. clasifícalo como `PRODUCT-FAIL`, `TEST-GAP`, `ENVIRONMENT-BLOCKED`, `KNOWN-GAP`, `DOC-CONTRADICTION` o `INTERMITTENT`;
4. no lo corrijas, no relajes una aserción y no abras un issue;
5. no presentes una limitación del navegador integrado como defecto del producto sin confirmarla con el runner E2E o una interacción manual equivalente.

## Baseline observado antes de la campaña completa

Este baseline fue observado el 2026-09-02 en el árbol de trabajo disponible, sin alterar código:

| Observación | Resultado | Clasificación inicial |
| --- | --- | --- |
| Frontend Vite en `http://127.0.0.1:5173` | Carga la ruta `/acceso` y redirige desde `/` | `PASS` de arranque visual |
| Identidad de acceso | Wordmark `NAVA` y H1 `Accede a NAVA` | `PASS` visual NAVA |
| Recuperación pública | `/recuperar-acceso` muestra `Solicita tu código`, paso 1 de 3 | `PASS` de navegación pública |
| Acceso a 320 px | Sin scroll horizontal; inputs de 44 px y CTA de 48 px | `PASS` observado |
| Reflow equivalente a zoom 200 % | En viewport 640 px no aparece scroll horizontal ni recorte | `PASS` observado; confirmar con zoom real |
| Foco y validación vacía | Orden correo → contraseña → CTA → recuperación; resumen de errores recibe foco | `PASS` observado |
| 429 simulado | Caso E2E visual de bloqueo 320 px pasó con respuesta `application/problem+json` simulada | `PASS` de estado simulado, no del API real |
| API local en `127.0.0.1:8080` | No respondió; no hubo backend disponible | `ENVIRONMENT-BLOCKED` |
| PostgreSQL/Docker/Go | No hubo conexión utilizable a Docker; `go` y `psql` no estuvieron disponibles en PATH | `ENVIRONMENT-BLOCKED` |
| E2E visual normal 320 px | Falló antes de capturar porque espera H1 `Inicia sesión`, pero la UI actual muestra `Accede a NAVA` | `TEST-GAP`/`DOC-CONTRADICTION` por reconciliar; no es aún bug de producto |

La discrepancia `Inicia sesión`/`Accede a NAVA` debe verificarse contra `DEC-077`, `docs/03-desarrollo/especificacion-frontend-nava.md` y la pantalla actual antes de cambiar cualquier prueba. No la conviertas automáticamente en una corrección de producto. La dificultad de escribir mediante una API de automatización concreta tampoco es evidencia suficiente de que un campo Vue no acepte teclado; confírmala con el runner E2E y el control manual equivalente.

## Objetivo observable

Entregar un informe reproducible que permita saber:

1. cómo se ve y se comporta cada pantalla implementada de B0 a B3 en el navegador;
2. qué recorridos pasan y cuáles fallan con el stack real;
3. qué fallos son de producto y cuáles son del entorno, de datos o de expectativas desactualizadas;
4. si los cambios NAVA conservan jerarquía, responsive, foco, objetivos táctiles y navegación;
5. qué bugs nuevos o regresiones deben convertirse después en prompts de fix separados.

No uses “se ve bien”, “todo bien” ni el estado verde de CI sin registrar observación, selector, viewport y evidencia.

## Preflight y seguridad

1. Registra fecha/hora UTC, rama, SHA, `git status --short`, SO/arquitectura, Node, pnpm, Go, PostgreSQL, Atlas y navegadores.
2. Conserva todos los cambios ajenos del árbol. No hagas `pull`, `reset`, `checkout`, `stash`, limpieza automática ni regeneración de artefactos versionados.
3. Confirma que las URLs sean locales y que la base sea desechable antes de sembrar.
4. Usa dos tenants ficticios y dos contextos de navegador aislados; redacta IDs, cookies, tokens, claves de idempotencia, `request_id`, correos y teléfonos.
5. Guarda screenshots, trazas y logs en `%TEMP%/nava-qa-browser-<timestamp>` o equivalente. Las specs históricas que escriben en `apps/web/e2e/evidence/` no deben sobrescribir capturas versionadas: si se ejecutan, conserva la captura en una ubicación temporal y restaura solamente cualquier archivo que la campaña haya modificado accidentalmente.
6. No uses proveedores reales, correo real, WhatsApp real, credenciales reales ni secretos en navegador, terminal, reporte o archivos versionados.
7. Si el API, PostgreSQL o el navegador requerido faltan, marca cada caso dependiente como `ENVIRONMENT-BLOCKED` y continúa solo con los casos independientes. No inyectes cookies ni fabriques un login exitoso para declarar que el panel real funciona.

## Preparación del stack

Cuando el entorno lo permita:

- levanta PostgreSQL 14+ dedicado y desechable;
- aplica migraciones con Atlas v1.3.0 desde vacío;
- verifica segundo `apply` sin pendientes y registra `status`;
- carga fixtures ficticios equivalentes a `database/testdata/` para dos tenants;
- levanta `go run ./cmd/api` en `127.0.0.1:8080` y Vite en `127.0.0.1:5173`;
- comprueba `/health` sin exponer versión, DSN, usuario ni stack trace;
- prepara credenciales E2E ficticias solo por variables seguras del proceso.

Si solo está disponible Vite, ejecuta la línea pública y visual: `/`, `/acceso`, `/recuperar-acceso` y los estados de validación. Reporta el panel y las mutaciones como bloqueados por la dependencia ausente, nunca como aprobados.

## Recorrido browser-first

### B0 — acceso, recuperación y shell

En el navegador integrado, y luego con el runner E2E si existe:

- `/` redirige a `/acceso`;
- acceso vacío muestra errores asociados y mueve el foco al resumen cuando hay más de un error;
- correo inválido y contraseña vacía no generan petición de negocio;
- credencial inválida no enumera cuentas y conserva únicamente lo permitido por la regla;
- fallo de red conserva los datos, muestra `Reintentar` y permite un reintento sin recargar;
- durante el envío el botón se deshabilita y dos toques producen una sola solicitud;
- 429 muestra bloqueo y tiempo seguro sin revelar detalles;
- recuperación navega por sus tres pasos y conserva el correo entre pasos, sin enviar OTP real;
- `/panel` sin sesión conserva el destino y vuelve al acceso;
- con sesión real, F5, cierre/reapertura, logout y sesión expirada tienen el resultado esperado.

### B1 — configuración, barberos, servicios y asignaciones

Con sesión real del tenant A:

- inspecciona `/panel/barberia`, `/panel/barberos`, `/panel/servicios` y `/panel/servicios-por-barbero`;
- crea, edita, pagina y recarga datos ficticios;
- verifica límites, unicode, HTML literal, precio COP exacto, moneda sin `$`, filas mínimas y estados de carga/error;
- confirma que desactivar/reactivar y asignar/desasignar mantienen las reglas de última asignación;
- verifica que diálogos atrapan foco, Escape lo devuelve al disparador y el texto accesible no duplica asteriscos;
- repite una lectura de un ID de A desde B y confirma `404` sin dato transitorio.

### B2 — horarios, excepciones y bloqueos

Con datos ficticios:

- inspecciona `/panel/horarios` y `/panel/bloqueos` en estado vacío y con datos;
- crea jornada simple, partida, contigua y nocturna;
- intenta solapes y rangos inválidos;
- verifica festivo automático, excepción cerrada/especial, retiro y bloqueo puntual/serie;
- conserva como `KNOWN-GAP` lo que el alcance excluye explícitamente por el seguimiento #100;
- comprueba que el retiro lógico no pierde la evidencia esperada y que las acciones no mutan el tenant B.

### B3 — turnos y agenda

Con PostgreSQL real y dos barberos ficticios:

- crea turno manual dentro de la próxima hora y fuera de la rejilla pública;
- confirma que solo aparecen servicios activos asignados al barbero elegido;
- prueba contigüidad exacta, cruce de un minuto, bloqueo vigente y turno que cruza medianoche;
- revisa `/panel`, selector obligatorio de barbero, fecha/zona de barbería y agenda vacía/llena;
- navega anterior/siguiente/fecha, F5, atrás/adelante y cambio rápido de barbero/fecha;
- abre detalle, snapshots, contacto opcional, historial paginado y vuelta conservando query;
- reprograma un turno confirmado, verifica agenda anterior/nueva, un único evento, no-op, bloqueo, cruce, versión obsoleta, estado inválido y reintento idempotente.

## Matriz visual y accesible mínima

Para `/acceso`, `/recuperar-acceso`, `/panel`, `/panel/turnos/nuevo`, `/panel/turnos/{id}`, `/panel/barberia`, `/panel/barberos`, `/panel/servicios`, `/panel/servicios-por-barbero`, `/panel/horarios` y `/panel/bloqueos`, registra al menos:

| Vista | Anchos | Estados | Entrada | Verificación |
| --- | --- | --- | --- | --- |
| Público y privado | 320, 360, 768, 1024, 1280 | vacío, datos, carga, error, éxito | teclado y táctil | `scrollWidth <= clientWidth`, contenido no tapado |
| Diálogo | 320, 1280 | abierto, error, envío | Tab, Shift+Tab, Escape | foco atrapado, retorno de foco, sin recorte |
| Zoom | 1280 al 200 % real o equivalente documentado | datos y error | teclado | reflow, no pérdida de controles |
| Preferencia | anchos representativos | carga y transición | `prefers-reduced-motion` | sin dependencia de animación para comprender el estado |

En cada control mide el rectángulo real; botones, campos, enlaces y áreas de checkbox deben cumplir 44×44 px cuando el estándar lo exige. Comprueba H1/H2, nombre/rol/estado, `aria-current`, `aria-invalid`, `aria-describedby`, foco visible, contraste según la política del estándar y ausencia de scroll horizontal.

## Red, errores y concurrencia

- Usa red lenta, offline, timeout y reintento sobre datos ficticios.
- Para estados puramente visuales puede interceptarse una respuesta, pero el informe debe decir `SIMULADO`; no cuenta como verificación de persistencia.
- Verifica `400/401/404/409/422/429` y un `500` por el mecanismo de prueba existente, sin corromper datos.
- Comprueba `application/problem+json`, código de problema, `request_id` abreviado, mensaje seguro y ausencia de PII/secretos.
- Ejecuta doble envío, dos pestañas y carreras de crear/reprogramar/desasignar contra PostgreSQL real.
- Tras cada fallo de negocio cuenta filas, historial y efectos finales; cero éxito falso tras timeout/offline.

## Triage de hallazgos

Usa la plantilla íntegra de `docs/07-calidad/06-plantilla-hallazgo.md` y añade:

```text
ID | severidad | clasificación | HU/RN/DEC/CA/fix | SHA
URL | navegador | viewport | zona | red | precondición
pasos numerados | esperado | observado | frecuencia
evidencia temporal | request_id/status/code redactados | impacto
```

Severidad:

- Bloqueante: fuga de tenant, cruce persistido, corrupción, secreto/PII expuesto o 500 sin recuperación en P0.
- Alto: dato persistido incorrecto, historial/idempotencia rota o flujo P0 sin salida.
- Medio: error de interacción, teclado, responsive o mensaje que dificulta completar.
- Bajo: detalle visual/copy sin efecto funcional ni accesible.

No registres como bug B4–B6, T3, cancelación/completar/no-show, vista consolidada multi-barbero, modo oscuro, sexto ítem temporal permitido o seguimientos explícitamente fuera de alcance.

## Criterios de salida

El resultado es `APROBADA` solo si el stack real está disponible, los recorridos y combinaciones aplicables tienen evidencia, los recursos cruzados devuelven `404`, no hay cruces/duplicados/fugas/500 reproducibles y no quedan hallazgos Bloqueantes o Altos. Si falla una expectativa por título o prueba desactualizada, clasifícala como `TEST-GAP`/`DOC-CONTRADICTION` hasta reconciliar la fuente normativa. Si falta API, PostgreSQL, navegador o credenciales ficticias seguras, el resultado es `INCONCLUSA`, nunca “aprobada con supuestos”.

## Informe final obligatorio

Entrega:

1. Resultado ejecutivo: `APROBADA`, `NO APROBADA` o `INCONCLUSA`, con máximo tres razones.
2. Manifiesto: SHA, árbol inicial/final, stack, DB/Atlas, navegador, zonas y límites.
3. Tabla de rutas/pantallas × viewport × estado × evidencia.
4. Puertas G0-G8 y recorridos B0-B3 con `PASS/FAIL/BLOCKED/N-A`.
5. Matrices PA/PB/PC y regresiones R01-R13 del prompt integral.
6. Hallazgos completos, separados de bloqueos de entorno, gaps de pruebas y capacidades fuera de alcance.
7. Próximo paso recomendado: crear un prompt de fix o pedir una decisión; no ejecutar ese paso dentro de esta campaña.

No modifiques este prompt con resultados posteriores: si cambia materialmente el alcance, crea una nueva versión que enlace `supersedes` y actualiza el índice.
