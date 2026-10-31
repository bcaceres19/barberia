# Campaña #301 · Informe consolidado de pruebas de UI (primera pasada y reencolados)

- Issue: #301 · Prompt de orquestación: PROMPT-ORCH-301-UI-LUNA-v1 · Campaña: `qa-ba991bb4eb`
- SHA ejecutado: `fd4fabe` (rama `test/301-pruebas-ui`) · Fecha: 2026-10-30
- Modelo real: **Claude Sonnet 5.5** (esfuerzo por defecto) como subagentes de Claude Code. El kit pide GPT-6 Luna / medium, que este runtime no ofrece; queda registrado como desviación.
- Entorno: BD PostgreSQL efímera `barberia_qa` con las 39 migraciones, API local y Vite, 20 cuentas y 20 tenants sintéticos aislados (uno por scope). Todos los proveedores externos **SIMULATED** (sin Google real, WhatsApp, correo ni Calendar): el mapa y la geocodificación se probaron simulados. La evidencia de `public-entry` que el agente rotuló como REAL se reclasificó a SIMULATED.
- Navegador: Chromium propio por agente (`playwright-core`), viewports 320/360/768/1280, teclado, zoom 200 % (emulado a 640 px × 2 cuando no hubo zoom real) y axe-core.
- Evidencia cruda (capturas, trazas, tokens) queda en `apps/web/test-results/ui-qa/qa-ba991bb4eb/` (ignorada, privada). Esta copia pública no contiene credenciales, cookies, tokens ni URL de Mi turno.

> Resultado: **no se encontró fallo en los casos ejecutados que no esté listado abajo**. No se afirma ausencia de bugs ni cobertura total.

## Conteos

| Scope          |   Casos |    PASS |   FAIL | BLOCKED | NOT_RUN |    N/A |
| -------------- | ------: | ------: | -----: | ------: | ------: | -----: |
| access         |      18 |      16 |      1 |       0 |       0 |      1 |
| recovery       |      26 |      18 |      0 |       0 |       1 |      7 |
| shop           |      30 |      17 |      0 |       4 |       6 |      3 |
| policy         |      25 |      17 |      0 |       1 |       4 |      3 |
| staff          |      26 |      14 |      2 |       0 |      10 |      0 |
| catalog        |      25 |      12 |      5 |       1 |       7 |      0 |
| assignments    |      23 |      19 |      1 |       2 |       0 |      1 |
| schedules      |      24 |      19 |      1 |       1 |       3 |      0 |
| agenda         |      24 |      17 |      2 |       1 |       4 |      0 |
| manual         |      25 |      22 |      0 |       1 |       1 |      1 |
| detail         |      27 |      20 |      1 |       3 |       3 |      0 |
| public-entry   |      18 |      12 |      0 |       2 |       0 |      4 |
| public-catalog |      22 |      13 |      1 |       0 |       0 |      8 |
| public-barber  |      24 |      15 |      0 |       0 |       3 |      6 |
| public-slots   |      25 |      11 |      2 |       1 |       1 |     10 |
| public-confirm |      26 |      13 |      3 |       0 |       5 |      5 |
| customer       |      28 |      15 |      1 |       2 |       5 |      5 |
| journeys       |      24 |      15 |      2 |       1 |       0 |      6 |
| **Total**      | **440** | **285** | **22** |  **20** |  **53** | **60** |

`executed` en el índice de prompts significa ejecución informada, no éxito de toda la app. Los BLOCKED/NOT_RUN no se convirtieron en PASS.

## Hallazgos por severidad

### P1 — flujo o guardado roto

| ID                    | Pantalla                       | Resumen                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | Norma                         |
| --------------------- | ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------- |
| F-DETAIL-01           | detalle de turno / alta manual | El panel acepta crear y reprogramar turnos fuera de la jornada (07:00, 21:00, 03:00); el servicio solo consulta bloqueos.                                                                                                                                                                                                                                                                                                                                                                                                            | CA-065-04, CA-061-02          |
| PCONF-F-01 (= J-F-01) | reserva pública                | Reintento tras respuesta perdida con la **misma** Idempotency-Key devuelve 409 `slot-conflict` en lugar de repetir el 201; la UI dice «Ese horario se acaba de ocupar» aunque la cita ya es del cliente. Reproducido 3 veces (UI y API). Causa verificada por lectura: `publicbooking/confirm.go` revalida disponibilidad antes de consultar el registro de idempotencia (que ocurre dentro de `CreatePublicAppointment`). Elegir una alternativa tras ese 409 crearía una segunda cita. El alta manual del panel sí es idempotente. | RN-IDE-01, CA-097-04, DEC-043 |

### P2 — fallo recuperable, responsive o accesible

| ID                    | Pantalla             | Resumen                                                                                                                                                                     |
| --------------------- | -------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ACCESS-F-01           | sesión               | La renovación deslizante actualiza BD pero no el `Expires` de la cookie (DEC-050).                                                                                          |
| RECOVERY-F-02         | recuperación         | Duda: reintento tras perder la respuesta del reset muestra «El enlace venció» aunque la contraseña ya cambió (RN-IDE-01 frente a DEC-064).                                  |
| STAFF-F1 / J-F-02     | bloqueos             | Crear un bloqueo sobre turnos confirmados no avisa de las citas afectadas ni marca la agenda (RN-BLQ-03). HU-042/HU-062 podrían diferirlo: brecha posiblemente documentada. |
| HALL-CAT-01 (= XS-F1) | catálogo             | El impacto de desactivar un servicio siempre dice 0 citas futuras (`apps/api/internal/modules/catalog/postgres/repository.go:594`).                                         |
| HALL-CAT-02           | catálogo             | Reactivar con nombre activo duplicado responde 500 (esperado 409 controlado).                                                                                               |
| HALL-CAT-03           | catálogo             | Total y paginador obsoletos tras crear.                                                                                                                                     |
| HALL-CAT-04           | catálogo             | Botón primario ilegible al hover (contraste 1.16:1).                                                                                                                        |
| ASSIGN-F-1            | asignaciones         | El foco se pierde tras alternar una casilla con Espacio (UI-T-14).                                                                                                          |
| F-SCH-01              | jornadas             | Editar una excepción sobre una fecha ocupada devuelve 404 en lugar de 409.                                                                                                  |
| AGENDA-H-02           | agenda               | Controles de 40 px frente a los 44 px exigidos (CA-062-08, CA-063-08).                                                                                                      |
| PENTRY-F-01           | entrada pública      | Sin aviso si se cae la red y luego se pulsa «Reservar un turno» (falla el import dinámico; solo F5 recupera).                                                               |
| PC-F1                 | catálogo público     | A zoom 200 % el pie se superpone a la lista y a la acción; probablemente del shell.                                                                                         |
| PSLOTS-F-01           | horarios públicos    | La tira de fechas se aplasta (24–35 px) en viewports bajos; la lista de horas queda en ~112 px (CA-095-05).                                                                 |
| PCONF-F-02            | confirmación pública | A 320×640 y 360×640 el resumen se solapa con el pie.                                                                                                                        |
| CUST-F2               | Mi turno             | La etiqueta «ERROR» del enlace inválido tiene contraste 2.04:1 (serio en axe); probable en otros usos de `PageState`.                                                       |

### P3 — detalle (resumen)

- **access:** F-02 reenviar no invalida el código anterior; F-03 copy de Retry-After (86399 s); F-04 un 401 por sesión vencida no conserva el destino.
- **recovery:** F-01 el texto sugiere envío a ambos canales.
- **policy:** campo vacío guarda 0; copy del aviso de conflicto; Enter no guarda; concatenación de dígitos (2.5→25).
- **shop:** DUDA-SHOP-01 una pestaña obsoleta revierte en silencio un campo no tocado; OBS-SHOP-01 la ficha UI-SHOP-05 está obsoleta por DEC-134.
- **staff:** F2 emoji partido en iniciales; F3 desborde a 320 px; O1 retirar bloqueo sin confirmación; O2 maxlength trunca en silencio.
- **catalog:** HALL-CAT-05 contraste «Inactivo» 3.49:1; HALL-CAT-06 campos numéricos filtran a dígitos.
- **schedules:** F-SCH-02 foco a BODY.
- **agenda:** H-01 preselecciona el primer profesional con dos (DEC-074); H-03 tras fallo de red se ven turnos del día anterior bajo la fecha nueva; H-04/H-05 cosméticos.
- **manual:** F01 resumen desborda; F02 mensaje de formato con hora vacía; F03 foco al documento tras enviar.
- **detail:** F-DETAIL-02 historial con UTC crudo y códigos internos; F-DETAIL-03 precio sin formato y «a. m..».
- **public-entry:** F-02 título estático; F-03 «Reintentar» en slug inexistente. **public-catalog:** PC-F2, PC-F3.
- **customer:** F1 copy «este enlace deja de funcionar» contradice DEC-136; F3 foco a BODY al «Mantener mi turno»; F4 `/mi-turno/` sin token queda en blanco.
- **journeys:** F-03 formato de precio inconsistente; F-04 fechas UTC y estados internos en el historial.

## Dudas para el propietario

Resueltas el 2026-10-31 en `DEC-140` (y `DEC-127` para la 3). Las que siguen abiertas necesitan una decisión del propietario.

| # | Duda | Estado |
| --- | --- | --- |
| 1 | Reproducir el 201 frente a revalidar (`RN-IDE-01`/`DEC-043` y `DEC-064`) | Resuelta: en la recuperación prevalece `DEC-064`; en la reserva pública se reproduce (#393). |
| 2 | `CA-092-03`: 200 vacío o 404 con un ID ajeno | Resuelta: se conserva y `CA-092-03` se reescribió. |
| 3 | `HU-092`/`DEC-104` excluyen fotos y perfiles | Resuelta: `DEC-127` registrada (retroactiva) y `HU-092` actualizada. |
| 4 | `CA-099-05`: 200 sin cambios si el cliente ya canceló | Abierta: es deliberado en el código (idempotente, sin duplicar historial), pero `DEC-136` habla de «rechazo de estado» ante otra clave. Hay que aclarar cuál manda. |
| 5 | HU dueña del aviso de citas afectadas al crear un bloqueo (`RN-BLQ-03`) | Abierta: asignarla a la HU de bloqueos de barbero con un criterio nuevo. |
| 6 | Agenda con 2+ profesionales: ¿se preselecciona el primero? | Abierta: `DEC-074` pide un «selector obligatorio» sin default implícito, pero `CA-063-01` solo exige que el selector exista. Decisión del propietario. |
| 7 | Teléfonos nacionales de 6 dígitos | Abierta: no hay mínimo de dígitos por país; endurecerlo exige una decisión. |
| 8 | Concurrencia en catálogo, equipo y marca (UI-T-08) | Resuelta: «última escritura gana»; UI-T-08 no aplica. |
| 9 | `DEC-004`/`RN-SER-04`: elegir a qué turnos se aplica el precio nuevo | Abierta: no es un defecto (los turnos guardan su precio); falta una HU en el backlog. |
| 10 | Panel `independiente` con más de un barbero | Resuelta: los selectores reaparecen (`DEC-115`, `DEC-134`). |
| 11 | `BaseAlert` en `<h4>` frente a axe `heading-order` | Resuelta: se mantiene (`HU-009`); excepción aceptada. |
| 12 | `PUT` de la política de reservas ante un reintento idéntico | Resuelta: se mantiene el 409; el aviso es neutral (#441). |

## Capacidades no verificadas

Mapa, geocodificación, OTP y correo reales (el OTP real por WhatsApp Meta se probó aparte en #86), Google Calendar, worker, recordatorios, retención, zoom real al 200 %, frontera exacta de anticipación/ventana y plazo de cancelación (requiere reloj controlado), RLS más allá de las peticiones cruzadas por HTTP (cubiertas también en #87).

## Segunda tanda: reencolados y correcciones (2026-10-30 y 31)

Con apoyo SQL acotado por tenant (cada valor tocado se restauró) se reencolaron los casos BLOCKED/NOT_RUN que dependían de él. El detalle por caso queda en el `reencolado.md` privado de cada scope; aquí solo el resultado.

**Reencolados sin fallo:**

- UI-T-10 (sesión vencida) en los 9 scopes que la tenían pendiente; public-entry 02b y 07b; public-slots 03b; recovery CA-008-10; detail 05/07 (turnos movidos al pasado).
- Repeticiones de public-confirm 05 y T-06 tras la corrección de PCONF-F-01: 1 cita, 1 token y la misma clave de idempotencia.
- **customer 03b:** un token revocado y uno vencido responden el mismo 404 uniforme que un token desconocido (API), con el mismo aviso en la UI a 320 y 1280 px y sin datos del turno.
- **public-confirm 06b:** con un bloqueo real creado desde el panel mientras el resumen estaba abierto, «Confirmar turno» da 409 `slot-conflict`, 0 citas y alternativas (RN-CON-06 se cumple). Hallazgo de copy corregido en #445.
- **public-barber EXTRA-CALIFICACION:** con reseñas reales insertadas, promedio 4,0 y «2 reseñas»; la reseña oculta no cuenta ni se ve; un comentario con `<b>` se muestra literal. La base exige el turno `completed` (trigger). Hallazgo de layout corregido en #447.
- **public-barber EXTRA-ULTIMA-ASIGNACION (DEC-114):** retirar las dos asignaciones de un servicio da 204 sin 409; un tercer DELETE da 404 uniforme; el servicio sale del catálogo público y de los servicios del barbero (cita manual, DEC-072), sigue en el catálogo privado y la URL directa dice «Este servicio no tiene barberos disponibles». Restaurado.
- **UI-AGENDA-07 y J-10 (perfil individual):** con un solo barbero y rol `independiente` no aparece selector de barbero en Agenda, Horarios, Servicios ni Nuevo turno, y «Barberos» pasa a «Mi perfil»; sin desborde y axe sin violaciones a 360 y 1280 px. No se creó una cita manual dentro de este recorrido.
- Los scopes policy, schedules, agenda, manual, staff, catalog, shop, public-barber y customer volvieron a correrse con resultado mayoritariamente PASS y los hallazgos de abajo.

**04c (límite exacto del plazo de cancelación):** no es verificable desde la UI sin controlar el reloj del servidor. Queda cubierto en la capa más baja: `booking/cancel_customer_test.go` fija con reloj inyectado que «exactamente en el límite cuenta como dentro de plazo» y que un instante después ya no.

**Siguen sin ejecutarse o bloqueados por norma:**

- public-slots 02b (frontera de anticipación y ventana: requiere reloj controlado; hay pruebas de dominio, no de UI).
- UI-T-15 por conflicto en shop (la pantalla no tiene 409) y UI-T-08 de staff, catalog y barberos (BLOCKED: no hay política de concurrencia, duda 8).
- Los NOT_RUN restantes de cada `report.md` privado.

### Hallazgos corregidos

Cada uno en su issue, rama y PR, con prueba en la capa más baja y verificación en Chromium.

| Hallazgo                                  | PR   |
| ----------------------------------------- | ---- |
| PCONF-F-01 (P1)                           | #393 |
| F-DETAIL-01 (P1)                          | #395 |
| HALL-CAT-01                               | #397 |
| HALL-CAT-02                               | #400 |
| F-SCH-01                                  | #402 |
| ACCESS-F-01                               | #404 |
| CUST-F2                                   | #406 |
| HALL-CAT-04                               | #408 |
| AGENDA-H-02                               | #410 |
| HALL-CAT-03                               | #412 |
| ASSIGN-F-1                                | #414 |
| PSLOTS-F-01                               | #416 |
| PCONF-F-02                                | #418 |
| PC-F1                                     | #420 |
| PENTRY-F-01                               | #422 |
| RECOVERY F-01 (copy, P3)                  | #424 |
| ruta desconocida en blanco (P3) ¹         | #426 |
| HAL-SHOP-01 (P2, nuevo)                   | #428 |
| F-RQ-01 (P2, nuevo)                       | #430 |
| F-RQ-PB-01 (P3, nuevo)                    | #432 |
| F-RQ-02: `startsAt` ilegible o pasado     | #437 |
| HALL-RQ-CAT-02: total falso en el pie     | #439 |
| F-POLICY-05 y OBS-SHOP-03: copy del 409   | #441 |
| F-RQ-CU-01: Mi turno al cambiar de token  | #443 |
| F-RQ-03: copy del conflicto por bloqueo   | #445 |
| F-RQ-PB-02: pie sobre «Continuar» (nuevo) | #447 |

¹ Cubre `OBS-SHOP-05` y `customer F4` (`/mi-turno/` sin token): una ruta final con «No encontramos esta página» sustituye la pantalla en blanco, sin revelar qué rutas privadas existen.

### Sin corregir por decisión o duda abierta

- **RECOVERY-F-02:** se mantiene el comportamiento, porque `DEC-064` prevalece sobre `RN-IDE-01` en el reintento del reset (duda 1).
- **STAFF-F1 / J-F-02:** sin dueño normativo claro para el aviso de citas afectadas al bloquear (duda 5).
- **F1 de customer:** «este enlace deja de funcionar» sigue en pantalla; el texto contradice `DEC-136` y espera confirmación.
- **RQ-F2 (P3):** el título de `BaseAlert` es un `<h4>` y axe marca `heading-order` tras el `h1`. Es una decisión de HU-009, usada en cinco pantallas y excluida de axe a propósito (`e2e/evidence/auth-fidelidad-desviaciones.md`); cambiarlo afectaría a toda la app (duda 11).
- **Si el `PUT` de la política debe tolerar un reintento idéntico** tras una respuesta perdida: sin norma; solo se corrigió el texto (duda 12).

Las dudas 8 a 12 están en la tabla de «Dudas para el propietario».

## Correcciones

La campaña no corrige producto en su propio PR: cada hallazgo confirmado se atiende en su issue, rama y PR (tabla arriba). Con esta tanda todos los casos tienen un estado con evidencia; los que siguen sin ejecutarse están listados arriba con su motivo (reloj, falta de norma). El criterio de #301 que exige subagentes GPT-6 Luna / medium no se cumplió (se usó Claude Sonnet 5.5, desviación registrada al inicio), así que el cierre de #301 queda a decisión del propietario.
