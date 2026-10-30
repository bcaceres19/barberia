# Campaña #301 · Informe consolidado de pruebas de UI (primera pasada)

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

1. ¿El contrato exige reproducir el 201 cuando la franja ya la ocupa esa misma solicitud (RN-IDE-01 / DEC-043 frente a la revalidación previa)? Aplica también a RECOVERY-F-02 (DEC-064).
2. CA-092-03 pide «rechazar de forma uniforme» un ID ajeno: el API devuelve 200 con lista vacía (público) y 404 en el POST. ¿Es lo normativo? (public-barber DUDA-01, journeys D-03, public-slots).
3. HU-092/DEC-104 excluyen fotos y perfiles en la selección pública, pero la pantalla los muestra por decisiones posteriores: ¿se actualiza HU-092?
4. CA-099-05 pide conflicto ante versión obsoleta; el servicio responde 200 sin cambios si la cita ya está cancelada por el cliente (customer D1).
5. ¿El aviso de citas afectadas al crear un bloqueo tiene HU dueña (journeys D-02)?
6. Agenda: ¿preselección del primer profesional con dos o más (DEC-074 / CA-063-01)?
7. Teléfonos nacionales de 6 dígitos (+57 300555) se aceptan porque la regla es solo E.164.

## Capacidades no verificadas

Mapa, geocodificación, OTP y correo reales (el OTP real por WhatsApp Meta se probó aparte en #86), Google Calendar, worker, recordatorios, retención, zoom real al 200 %, frontera exacta de anticipación/ventana y plazo de cancelación (requiere reloj controlado), RLS más allá de las peticiones cruzadas por HTTP (cubiertas también en #87).

## Pendiente para reencolar (siguiente tanda, con apoyo SQL acotado del coordinador)

- UI-T-10 (expirar solo la sesión QA con `expires_at = issued_at + interval '1 microsecond'`): shop, policy, staff, catalog, assignments, schedules, agenda, manual, detail.
- detail: UI-DETAIL-05/07 (mover citas propias al pasado). customer: 03b (revocar y vencer un token propio). public-entry: 02b y 07b (restaurando el valor). public-barber: reseñas/estrellas (INSERT sintético en `barber_review`) y retirar la última asignación (DEC-114). journeys: UI-J-10 (rol `independiente` en un tenant de un solo profesional) y elegir una alternativa tras el conflicto de UI-J-03. assignments-06.
- public-slots: 02b (reloj) y 03b; public-confirm: 06b (franja bloqueada), UI-T-04/05/07/15 y repetir 05/T-06 cuando se corrija PCONF-F-01; customer: 04c y NOT_RUN.
- Los NOT_RUN por scope listados en cada `report.md` privado (shop X-03, T-02/04/09/15/16; staff 10; catalog 7; etc.).
- Repetir UI-J-09 público tras la corrección de PCONF-F-01.

## Correcciones

Esta campaña no corrige producto. Cada hallazgo confirmado se atiende en su propio issue, rama y PR.
