---
titulo: "Registro de decisiones"
version: "1.49"
estado: "Vigente"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-10-09"
documentos_relacionados:
  - "contradicciones.md"
  - "matriz-trazabilidad.md"
  - "historial-cambios.md"
  - "dudas-pendientes.md"
  - "../01-producto/reglas-negocio.md"
  - "../01-producto/alcance-mvp.md"
  - "../03-desarrollo/flujo-git-github.md"
  - "../03-desarrollo/estandar-diseno-visual.md"
  - "../03-desarrollo/especificacion-frontend-nava.md"
  - "../../respuesta-manuales/respuesta-propuestas-oc.txt"
  - "../../respuesta-manuales/respuesta-dudas-pendientes.txt"
---

# Registro de decisiones

## 1. Uso y autoridad

Cada código `DEC-*` es estable y no se reutiliza. Este registro normaliza respuestas del propietario sin borrar su fuente. Si una respuesta no alcanza para decidir el asunto al que fue asociada, se registra únicamente lo que sí decidió y se conserva abierta la duda restante.

**Responsable de las decisiones de esta versión:** propietario del proyecto. La identidad nominal confirmada es NAVA, según `DEC-077`.

## 2. Resumen

| Código | Fecha | Decisión | Regla o duda principal | Estado |
| --- | --- | --- | --- | --- |
| `DEC-001` | 2026-08-05 | Referidos y comisiones quedan fuera de la versión actual | `RN-PRO-03`, `DP-COM-02` | Confirmada para el MVP |
| `DEC-002` | 2026-08-05 | La duración agenda un tiempo planificado, no garantiza el tiempo real | `RN-SER-01` | Confirmada; extremos resueltos después por `DEC-020` |
| `DEC-003` | 2026-08-05 | Desactivar un servicio exige advertencia y decisión sobre citas futuras | `RN-SER-03` | Confirmada |
| `DEC-004` | 2026-08-05 | Los cambios no se aplican silenciosamente a citas existentes | `RN-SER-04` | Confirmada |
| `DEC-005` | 2026-08-05 | Los límites de reserva son configurables y no restringen la cita manual | `RN-DIS-04` | Confirmada; valores fijados por `DEC-018` |
| `DEC-006` | 2026-08-05 | El paso de la rejilla es configurable por el barbero | `RN-DIS-06` | Confirmada; valor fijado por `DEC-018` |
| `DEC-007` | 2026-08-05 | Las horas conservan zona y se muestran en la zona de la barbería | `RN-DIS-07` | Confirmada |
| `DEC-008` | 2026-08-05 | Un bloqueo urgente puede afectar citas y abre un flujo asistido | `RN-BLQ-03` | Confirmada |
| `DEC-009` | 2026-08-05 | Los bloqueos se eliminan lógicamente | `RN-BLQ-04` | Confirmada |
| `DEC-010` | 2026-08-05 | La política de cancelación fuera de plazo es configurable | `RN-CAN-02` | Confirmada; plazo fijado por `DEC-018` |
| `DEC-011` | 2026-08-05 | El barbero puede cancelar una cita activa en cualquier momento | `RN-CAN-03` | Confirmada |
| `DEC-012` | 2026-08-05 | Cancelar libera inmediatamente la franja futura | `RN-CAN-04` | Confirmada |
| `DEC-013` | 2026-08-05 | Una carrera entre bloqueo y reserva no borra la cita confirmada | `RN-CON-06` | Confirmada |
| `DEC-014` | 2026-08-05 | El historial es inmutable y las correcciones son nuevas entradas | `RN-HIS-02` | Confirmada; anonimización resuelta por `DEC-025` |
| `DEC-015` | 2026-08-05 | Los intentos de envío viven en un registro técnico separado | `RN-REC-04` | Confirmada |
| `DEC-016` | 2026-08-05 | “Turno” es el término de usuario y “cita” el término técnico | `DP-PRD-01` | Confirmada |
| `DEC-017` | 2026-08-05 | `no_show` es P0; los estados admiten corrección auditada y cierre configurable | `DP-PRD-02/09/10` | Confirmada |
| `DEC-018` | 2026-08-05 | Se fijan los valores iniciales configurables de agenda y recordatorios | `DP-PRD-03/04/05`, `DP-UX-05`, `DP-NOT-02/03/04` | Confirmada |
| `DEC-019` | 2026-08-05 | El MVP admite uno o varios barberos y permite elegirlos | `DP-PRD-06`, `DP-UX-01` | Confirmada |
| `DEC-020` | 2026-08-05 | Calendario flexible, recurrencia, cruces de medianoche e intervalos semiabiertos | `DP-PRD-07`, `DP-BE-02/03/04` | Confirmada |
| `DEC-021` | 2026-08-05 | El barbero controla retrasos y avisos; no hay desplazamiento automático | `DP-PRD-08` | Confirmada |
| `DEC-022` | 2026-08-05 | Se fijan datos del formulario y acceso público por enlace aleatorio | `DP-UX-02/03/04` | Confirmada |
| `DEC-023` | 2026-08-05 | Backend en Go y frontend web en TypeScript | `DP-BE-01` | Confirmada; frameworks fijados por `DEC-033` y `DEC-034` |
| `DEC-024` | 2026-08-05 | PostgreSQL compartido con RLS por barbería | `DP-BD-01/02` | Confirmada |
| `DEC-025` | 2026-08-05 | Retención configurable de 24 meses y anonimización posterior | `DP-BD-03`, `DP-LEG-02` | Confirmada, sujeta a revisión jurídica |
| `DEC-026` | 2026-08-05 | Acceso con contraseña, defensa escalonada e incidente documentado | `DP-SEG-01/02/03` | Confirmada |
| `DEC-027` | 2026-08-05 | WhatsApp oficial y correo son canales configurables | `DP-NOT-01` | Confirmada |
| `DEC-028` | 2026-08-05 | Piloto gratuito de 4 semanas con contingencia y transición operativa | `DP-PIL-01/02/03`, `DP-COM-03` | Confirmada |
| `DEC-029` | 2026-08-05 | Precio y compensación se acuerdan después del piloto y antes de cobrar | `DP-COM-01/02` | Diferida con condición de reapertura |
| `DEC-030` | 2026-08-05 | Política de datos y acuerdo escrito son prerrequisitos del piloto | `DP-LEG-01/03` | Confirmada, sujeta a revisión jurídica |
| `DEC-031` | 2026-08-05 | Respaldos diarios, alojamiento económico y soporte comprometido | `DP-OPS-01/02/03` | Confirmada |
| `DEC-032` | 2026-08-05 | Los recordatorios automáticos son P0 | `CT-001` | Confirmada; contradicción resuelta |
| `DEC-033` | 2026-08-05 | El frontend usa Vue 3, TypeScript y Vite | Arquitectura frontend | Confirmada |
| `DEC-034` | 2026-08-05 | El backend usa Chi v5 sobre `net/http` | Arquitectura backend | Confirmada |
| `DEC-035` | 2026-08-05 | Se adoptan estándares obligatorios de código, pruebas y base de datos | Ingeniería y calidad | Confirmada |
| `DEC-036` | 2026-08-06 | Atlas administra las migraciones SQL versionadas de PostgreSQL | Evolución de base de datos | Confirmada |
| `DEC-037` | 2026-08-06 | OpenAPI 3.1.2 y Redocly gobiernan el contrato HTTP | Documentación y contrato API | Confirmada |
| `DEC-038` | 2026-08-06 | GitHub Flow, Conventional Commits y squash gobiernan la entrega de cambios | Ramas, commits y pull requests | Confirmada |
| `DEC-039` | 2026-08-06 | Un sistema visual ligero gobierna colores, medidas, componentes y pantallas | Interfaz completa | Confirmada |
| `DEC-040` | 2026-08-11 | Modelo de roles PostgreSQL: propietario `NOLOGIN`, `barberia_app` y `barberia_worker` separados, aprovisionados fuera de Atlas | `DDL-SEC-01`, `DDL-SEC-02`, `DDL-SEC-04` | Confirmada |
| `DEC-041` | 2026-08-11 | Vocabulario de `appointment_history.event_type` en inglés `snake_case` | `DDL-DOC` (contradicción interna de `estados-citas.md`) | Confirmada |
| `DEC-042` | 2026-08-11 | Fecha ancla de retención/anonimización: última actividad del cliente | `DEC-025`, `DDL-PRI-01` | Confirmada, sujeta a revisión jurídica |
| `DEC-043` | 2026-08-11 | Semántica concurrente de idempotencia: bloqueo consultivo transaccional y `409`/`425` inmediato | `DDL-IDEM-01` | Confirmada |
| `DEC-044` | 2026-08-11 | Versión mínima de PostgreSQL: se confirma 14, sin cambio sobre el código existente | `DDL` preflight | Confirmada |
| `DEC-045` | 2026-08-11 | `customer` se identifica y reutiliza por teléfono único por barbería (upsert tenant-aware) | `DDL-BIZ-03` | Confirmada |
| `DEC-046` | 2026-08-11 | La unicidad de correo de `customer` es por barbería, no global | `DDL-BIZ-03` | Confirmada |
| `DEC-047` | 2026-08-11 | `barber` se recorta al alcance de `HU-021`: alta, listado y renombrar; sin borrado, desactivación, orden ni vínculo automático con `staff_user` | `DDL-BIZ-02` | Confirmada |
| `DEC-048` | 2026-08-11 | Segundo y tercer recordatorio en 24 horas y 2 horas antes de la cita | `DEC-018`, `DDL-BIZ-01` | Confirmada |
| `DEC-049` | 2026-08-11 | La anonimización cubre todas las copias de datos personales, no solo `customer` | `DEC-025`, `DDL-PRI-01` | Confirmada, sujeta a revisión jurídica |
| `DEC-050` | 2026-08-11 | Sesión larga del barbero: cookie `HttpOnly`+`Secure`+`SameSite`, token opaco revocable, 30 días con renovación por uso | `DP-SEG-04`, `CA-006-02` | Confirmada |
| `DEC-051` | 2026-08-11 | Código de recuperación de acceso: WhatsApp oficial y correo, mismo proveedor de `DEC-027` | `DP-SEG-05` | Confirmada |
| `DEC-052` | 2026-08-11 | Límite de acceso: ventana de 15 minutos, escalamiento a verificación telefónica de 24 horas | `DP-SEG-06` | Confirmada |
| `DEC-053` | 2026-08-11 | Protocolo de lease de `notification_claim_due` (claim/CAS/recuperación); `retention_claim_due_customers` se mantiene sin lease | `DDL-CON-01`, `DDL-CON-02`, `DDL-OPS-01` | Confirmada |
| `DEC-054` | 2026-08-11 | Fórmula concreta de última actividad (DEC-042), marcador de anonimización, alcance de `appointment_history_change` y por qué `idempotency_record.response_body` no se toca | `DEC-042`, `DEC-049`, `DDL-PRI-01` | Confirmada, sujeta a revisión jurídica |
| `DEC-055` | 2026-08-13 | Resuelve `CT-003`: el inicio de sesión se mueve a `/api/v1/public/auth/login`; deja de estar bajo `/api/v1/private` | `CT-003`, `CA-006-04` | Confirmada |
| `DEC-056` | 2026-08-13 | Resuelve `CT-004`: `CA-010-01` se divide entre `HU-010` (navega a `/panel` protegido, mínimo) y `HU-012` (cascarón completo verificable) | `CT-004`, `CA-010-01`, `CA-012-01` | Confirmada |
| `DEC-057` | 2026-08-13 | Resuelve `DP-SEG-07`: cookie de sesión `barberia_session`, `Path=/api/v1`, `SameSite=Lax`, sin `Domain`, 30 días | `DEC-050`, `DP-SEG-07` | Confirmada |
| `DEC-058` | 2026-08-13 | Resuelve `DP-SEG-08`: `CA-005-05`/`CA-005-01` se dividen entre `HU-005` (aislamiento a nivel PostgreSQL/RLS) y `HU-006` (verificación end-to-end contra el logout real, nuevo `CA-006-07`) | `DP-SEG-08`, `CA-005-01`, `CA-005-05`, `CA-006-07` | Confirmada |
| `DEC-059` | 2026-08-13 | Resuelve `DP-UX-06`: el enlace de recuperación de `CA-010-08` navega a `/recuperar-acceso`, ruta real que declara que la recuperación aún no está disponible, hasta que `HU-011` la construya | `DP-UX-06`, `CA-010-08` | Confirmada |
| `DEC-067` | 2026-08-24 | Resuelve `DP-SER-01`: catálogo con moneda COP fija, sin servicios gratuitos (precio > 0) y nombres únicos entre servicios activos de la misma barbería | `DP-SER-01`, `HU-022` | Confirmada |
| `DEC-068` | 2026-08-24 | Resuelve `DP-SER-02`: un servicio activo debe conservar al menos un barbero asignado; se rechaza retirar la última asignación | `DP-SER-02`, `HU-023` | Confirmada |
| `DEC-069` | 2026-08-24 | Resuelve `DP-SER-03`: B1 construye un recuento simple de impacto al desactivar, sin protección de concurrencia; la advertencia completa con cancelación queda para B3 | `DP-SER-03`, `HU-024` | Confirmada |
| `DEC-070` | 2026-08-25 | Resuelve `CT-008`: siete FK de B2 pasan de `ON DELETE CASCADE` a `ON DELETE RESTRICT` | `CT-008` | Confirmada |
| `DEC-071` | 2026-08-27 | Resuelve `DP-CIT-01`: cliente manual sin teléfono se reconcilia por correo dentro de la barbería; sin teléfono ni correo, siempre fila nueva | `DP-CIT-01`, `HU-061` | Confirmada |
| `DEC-072` | 2026-08-27 | Resuelve `DP-CIT-02`: la cita manual solo puede usar servicios activos ya asignados al barbero elegido | `DP-CIT-02`, `HU-061` | Confirmada |
| `DEC-073` | 2026-08-27 | Resuelve `DP-CIT-03`: un bloqueo vigente rechaza la creación manual (bloqueo duro), igual que un cruce de citas | `DP-CIT-03`, `HU-061` | Confirmada |
| `DEC-074` | 2026-08-28 | Resuelve `DP-CIT-04`: la agenda diaria abre con selector obligatorio de un barbero, sin vista consolidada inicial | `DP-CIT-04`, `HU-062` | Confirmada |
| `DEC-075` | 2026-08-28 | Resuelve `DP-CIT-05`: un turno que cruza medianoche aparece en cada agenda diaria cuyo intervalo intersecta | `DP-CIT-05`, `HU-062` | Confirmada |
| `DEC-076` | 2026-09-01 | Resuelve `DP-CIT-06`: reprogramar hacia un bloqueo vigente se rechaza como conflicto | `DP-CIT-06`, `HU-065` | Confirmada |
| `DEC-077` | 2026-09-01 | La plataforma se llama NAVA y adopta la dirección visual Tailored Grid, sin ampliar el alcance funcional del MVP | Identidad y sistema visual frontend | Confirmada |
| `DEC-078` | 2026-09-02 | Se retiran las restricciones visuales prescriptivas y se libera la creación del diseño dentro de NAVA / Tailored Grid | Dirección visual, arquitectura frontend y UX | Confirmada |
| `DEC-079` | 2026-09-02 | Los mockups aprobados y la firma cromática NAVA gobiernan rediseños y pantallas nuevas, con libertad de composición y herramientas | Dirección visual de nuevas entregas frontend | Confirmada |
| `DEC-106` | 2026-10-02 | Skill compartido para arranque local con esquema y agenda comprobados | Operación local de agentes; issue #287 | Confirmada |
| `DEC-107` | 2026-10-02 | Barberos admite modo numerado para tabla adaptada al viewport y conserva cursor para selectores | HU-021; issue #288 | Confirmada |
| `DEC-108` | 2026-10-03 | El rechazo de credenciales en `/acceso` se avisa como aviso emergente, acotando DEC-095 | Acceso; `docs/03-desarrollo/estandar-diseno-visual.md` §6.8 | Confirmada |
| `DEC-109` | 2026-10-03 | Horarios adopta el tablero semanal sobre tinta del panel y deja de tener el atlas #196 como referencia exacta | HU-040, HU-041; `docs/03-desarrollo/estandar-diseno-visual.md` | Confirmada |
| `DEC-110` | 2026-10-03 | Marca, vocabulario y apariencia configurables de la barbería | `HU-025`; issue #292; amplía `DEC-039`, `DEC-077` | Confirmada |
| `DEC-111` | 2026-10-03 | La reserva pública adopta el lienzo de tinta del panel con un cascarón persistente | HU-090 a HU-097; `docs/03-desarrollo/estandar-diseno-visual.md` | Confirmada |
| `DEC-112` | 2026-10-03 | Escala tipográfica y movimiento compartidos del panel; Servicios por barbero sobre tinta | `docs/03-desarrollo/estandar-diseno-visual.md`; HU-023 | Confirmada |
| `DEC-113` | 2026-10-03 | Estados vacíos del panel como escena animada | `docs/03-desarrollo/estandar-diseno-visual.md` | Confirmada |
| `DEC-114` | 2026-10-04 | Se puede retirar a cualquier barbero de un servicio, también al último; sustituye a `DEC-068` | `DP-SER-02`, `HU-023`, `CA-023-05`, `CA-023-06` | Confirmada |
| `DEC-115` | 2026-10-04 | Perfil del panel por barbería: `shop` (con equipo) o `solo` (barbero individual), solo presentación | `HU-025`; issue #294; amplía `DEC-110` | Confirmada (elección del ajuste); composición del panel sujeta a revisión |
| `DEC-116` | 2026-10-08 | «Mi perfil» del barbero individual es una tarjeta propia, no una fila de tabla; amplía `DEC-115` | `HU-025`; issue #302 | Confirmada |
| `DEC-119` | 2026-10-08 | NAVA es multirrubro: el vocabulario del negocio llega a toda la interfaz (panel completo en #308; reserva pública, acceso del cliente y mensajes en #309); amplía `DEC-110` | `HU-025`; issues #308 y #309 | Confirmada |
| `DEC-120` | 2026-10-07 | Pruebas UI con Luna medium, prompts persistentes y cuentas sintéticas aisladas | Propietario; issue #300 | Confirmada |
| `DEC-121` | 2026-10-08 | El backend sube a Go 1.26.9 porque la línea 1.25 no tiene versión corregida de 9 vulnerabilidades de la librería estándar; amplía `DEC-023` | `DEC-035`; issue #317 | Confirmada |

## 3. Decisiones detalladas

### DEC-001 · Referidos y comisiones fuera de la versión actual

- **Fecha:** 2026-08-05.
- **Decisión:** la gestión de referidos y la definición o automatización de comisiones no forman parte del MVP ni de la versión actual. Solo se reabren mediante una decisión de alcance para una versión posterior y antes de activar cualquier pago por referidos.
- **Responsable:** propietario del proyecto.
- **Motivo:** es una capacidad de otra versión y no valida la operación básica de la agenda.
- **Alternativas descartadas:** definir ahora porcentaje y condiciones; automatizar referidos en el MVP; tratar el silencio como un acuerdo comercial.
- **Documentos afectados:** `RN-PRO-03`, `DP-COM-02`, `alcance-mvp.md`, `prioridades.md`, `glosario.md`.
- **Fuente:** `respuesta-propuestas-oc.txt`, línea 1.

### DEC-002 · Duración planificada frente a duración real

- **Fecha:** 2026-08-05.
- **Decisión:** la duración configurada calcula la hora final planificada y el espacio reservado en agenda. Un servicio real puede terminar antes o después; esa diferencia no reescribe automáticamente la cita.
- **Responsable:** propietario del proyecto.
- **Motivo:** el tiempo de un servicio es una estimación operativa necesaria para reservar, no una medición exacta de ejecución.
- **Alternativas descartadas:** asumir que el servicio siempre termina exactamente en `ends_at`; recalcular automáticamente las citas siguientes por la duración real; interpretar esta respuesta como aprobación de intervalos semiabiertos.
- **Documentos afectados:** `RN-SER-01`, `RN-DIS-05`, `DP-BE-04`, `glosario.md`.
- **Fuente:** `respuesta-propuestas-oc.txt`, líneas 2–4.
- **Límite histórico:** esta decisión no respondió los extremos; `DEC-020` los resolvió después como `[inicio, fin)`.

### DEC-003 · Desactivación de servicios con citas futuras

- **Fecha:** 2026-08-05.
- **Decisión:** al desactivar un servicio, el sistema conserva el historial, informa cuántas citas futuras quedan afectadas y permite al barbero mantenerlas o cancelarlas. No toma la decisión automáticamente. Toda cancelación resultante notifica a cada cliente afectado.
- **Responsable:** propietario del proyecto.
- **Motivo:** mantener la trazabilidad sin quitar al barbero el control de compromisos ya adquiridos.
- **Alternativas descartadas:** impedir la desactivación; borrar el servicio; cancelar todas las citas automáticamente; conservarlas sin advertencia.
- **Documentos afectados:** `RN-SER-03`, `F-SERV-01`, `F-CITA-06`, `F-NOT-02`.
- **Fuente:** `respuesta-propuestas-oc.txt`, líneas 5–9.

### DEC-004 · Aplicación y comunicación de cambios de servicio

- **Fecha:** 2026-08-05.
- **Decisión:** una cita conserva los datos con los que fue creada y ningún cambio de catálogo la modifica en silencio. Ante un cambio de precio, el barbero elige si se aplica solo a nuevas citas o también a las futuras ya agendadas. Todo cambio aplicado a una cita existente debe quedar en historial y notificarse al cliente por el canal vigente.
- **Responsable:** propietario del proyecto.
- **Motivo:** evitar cambios sorpresivos para el cliente y cruces provocados por mutaciones masivas de duración.
- **Alternativas descartadas:** aplicar siempre a todas las citas; aplicar siempre solo a las nuevas sin opción; modificar citas existentes sin historial ni aviso; fijar aquí WhatsApp o correo antes de resolver `DP-NOT-01`.
- **Documentos afectados:** `RN-SER-04`, `F-SERV-01`, `F-CITA-04`, `F-NOT-02`, `DP-NOT-01`.
- **Fuente:** `respuesta-propuestas-oc.txt`, líneas 10–13.

### DEC-005 · Límites configurables de agendamiento

- **Fecha:** 2026-08-05.
- **Decisión:** la anticipación mínima y la ventana máxima son configurables por barbería. Las citas creadas manualmente por el barbero no están sujetas a ninguno de esos dos límites, aunque sí a las reglas de integridad de agenda.
- **Responsable:** propietario del proyecto.
- **Motivo:** el barbero conoce su capacidad de reacción y necesita registrar compromisos recibidos por otros canales.
- **Alternativas descartadas:** valores globales inmodificables; ausencia de límites para clientes; aplicar los límites públicos a las citas manuales.
- **Documentos afectados:** `RN-DIS-04`, `DP-PRD-04`, `DP-PRD-05`, `F-DISP-04`, `F-DISP-05`, `estados-citas.md`.
- **Fuente:** `respuesta-propuestas-oc.txt`, líneas 14–16.

### DEC-006 · Rejilla configurable

- **Fecha:** 2026-08-05.
- **Decisión:** el barbero configura el paso en minutos de la rejilla de horarios ofrecidos al cliente. La creación manual puede usar cualquier minuto válido.
- **Responsable:** propietario del proyecto.
- **Motivo:** equilibrar cantidad de opciones y aprovechamiento de huecos según la operación de cada barbería.
- **Alternativas descartadas:** paso fijo para todas las barberías; ofrecer cualquier minuto al cliente; obligar la cita manual a seguir la rejilla.
- **Documentos afectados:** `RN-DIS-06`, `DP-UX-05`, `F-DISP-01`, `glosario.md`.
- **Fuente:** `respuesta-propuestas-oc.txt`, líneas 17–18.

### DEC-007 · Manejo de zonas horarias

- **Fecha:** 2026-08-05.
- **Decisión:** cada instante se almacena de forma inequívoca con zona horaria y la interfaz presenta la hora de la barbería, incluso si el cliente consulta desde otro país.
- **Responsable:** propietario del proyecto.
- **Motivo:** existen clientes o contactos en Estados Unidos y España; depender de la zona del dispositivo desplazaría las citas.
- **Alternativas descartadas:** usar la zona del cliente; usar la del servidor; guardar horas locales sin zona.
- **Documentos afectados:** `RN-DIS-07`, `F-CONF-01`, `F-DISP-01`, `SUP-002`, `glosario.md`.
- **Fuente:** `respuesta-propuestas-oc.txt`, líneas 19–21.

### DEC-008 · Bloqueos urgentes sobre citas existentes

- **Fecha:** 2026-08-05.
- **Decisión:** un bloqueo puede crearse aunque afecte citas. El sistema identifica las citas afectadas y ofrece un flujo de pocos pasos para que el barbero las reprograme o cancele; cada acción notifica a los clientes correspondientes. No hay reprogramación ni cancelación automática.
- **Responsable:** propietario del proyecto.
- **Motivo:** una emergencia debe poder bloquearse de inmediato sin perder control ni comunicación.
- **Alternativas descartadas:** rechazar el bloqueo; mover o cancelar citas automáticamente; crear el bloqueo sin mostrar afectados.
- **Documentos afectados:** `RN-BLQ-03`, `F-HOR-02`, `F-CITA-05`, `F-CITA-06`, `F-NOT-02`.
- **Fuente:** `respuesta-propuestas-oc.txt`, líneas 22–24.

### DEC-009 · Eliminación lógica de bloqueos

- **Fecha:** 2026-08-05.
- **Decisión:** eliminar un bloqueo solo lo retira del cálculo de disponibilidad; el registro y sus movimientos permanecen para auditoría.
- **Responsable:** propietario del proyecto.
- **Motivo:** no perder la explicación histórica de por qué una franja estuvo bloqueada.
- **Alternativas descartadas:** borrado físico; historial editable; mantener el bloqueo activo después de eliminarlo.
- **Documentos afectados:** `RN-BLQ-04`, `F-HOR-02`, `F-EST-03`, `glosario.md`.
- **Fuente:** `respuesta-propuestas-oc.txt`, línea 25.

### DEC-010 · Política configurable de cancelación fuera de plazo

- **Fecha:** 2026-08-05.
- **Decisión:** una vez vencido el plazo, la configuración de la barbería determina si puede cancelar solo el barbero o también el cliente y si se exige un motivo. La interfaz aplica y explica la política vigente.
- **Responsable:** propietario del proyecto.
- **Motivo:** cada barbero necesita ajustar el control de cancelaciones de último minuto a su operación.
- **Alternativas descartadas:** prohibición fija para todo cliente; permiso incondicional para todo cliente; política global no configurable; no conservar el motivo cuando sea obligatorio.
- **Documentos afectados:** `RN-CAN-01`, `RN-CAN-02`, `F-PUB-08`, `F-CITA-07`, `estados-citas.md`, `glosario.md`.
- **Fuente:** `respuesta-propuestas-oc.txt`, líneas 26–29.

### DEC-011 · Facultad de cancelación del barbero

- **Fecha:** 2026-08-05.
- **Decisión:** el barbero puede cancelar cualquier cita activa en cualquier momento si lo considera necesario. No puede cambiar una cita que ya esté en estado terminal.
- **Responsable:** propietario del proyecto.
- **Motivo:** el barbero conserva la decisión operativa final ante emergencias.
- **Alternativas descartadas:** imponerle el plazo del cliente; cancelación automática por el sistema; permitir alterar estados terminales.
- **Documentos afectados:** `RN-CAN-03`, `F-CITA-06`, `estados-citas.md`.
- **Fuente:** `respuesta-propuestas-oc.txt`, líneas 30–31.

### DEC-012 · Liberación inmediata tras cancelar

- **Fecha:** 2026-08-05.
- **Decisión:** una cita cancelada deja de ocupar una franja futura inmediatamente, salvo que otro bloqueo la cubra.
- **Responsable:** propietario del proyecto.
- **Motivo:** recuperar tiempo vendible y evitar pérdida de ingresos.
- **Alternativas descartadas:** retener la franja hasta la hora original; liberación manual; ignorar bloqueos concurrentes.
- **Documentos afectados:** `RN-CAN-04`, `F-DISP-01`, `F-PUB-08`, `F-CITA-06`, `estados-citas.md`.
- **Fuente:** `respuesta-propuestas-oc.txt`, líneas 32–34.

### DEC-013 · Carrera entre bloqueo y reserva

- **Fecha:** 2026-08-05.
- **Decisión:** si una cita alcanza a confirmarse en una franja que el barbero acaba de bloquear, la cita no se elimina ni cancela. El sistema advierte de inmediato al barbero, sin importar el orden de llegada, y él decide si la mantiene, reprograma o cancela.
- **Responsable:** propietario del proyecto.
- **Motivo:** una cita visible y resoluble es preferible a perderla silenciosamente.
- **Alternativas descartadas:** borrar la cita; cancelarla automáticamente; ocultar el conflicto; hacer depender el resultado del orden de milisegundos.
- **Documentos afectados:** `RN-CON-06`, `RN-BLQ-03`, `F-DISP-03`, `F-HOR-02`, `F-NOT-02`.
- **Fuente:** `respuesta-propuestas-oc.txt`, líneas 35–39.

### DEC-014 · Inmutabilidad del historial

- **Fecha:** 2026-08-05.
- **Decisión:** las entradas de historial no se modifican ni se eliminan. Un error se corrige con una entrada posterior que conserve el rastro.
- **Responsable:** propietario del proyecto.
- **Motivo:** la trazabilidad deja de ser confiable si se puede reescribir.
- **Alternativas descartadas:** edición; borrado físico; corrección que sustituya la entrada original.
- **Documentos afectados:** `RN-HIS-02`, `F-EST-03`, `estados-citas.md`, `DP-LEG-02`.
- **Fuente:** `respuesta-propuestas-oc.txt`, líneas 40–41.
- **Límite histórico:** el mecanismo legal no se decidió aquí; `DEC-025` confirmó después la anonimización con conservación del historial.

### DEC-015 · Registro técnico de intentos de envío

- **Fecha:** 2026-08-05.
- **Decisión:** cada intento de envío se conserva en un registro técnico específico de notificaciones o eventos, separado del historial general de negocio de la cita.
- **Responsable:** propietario del proyecto.
- **Motivo:** diagnosticar entregas y fallos sin contaminar el historial funcional ni exponer datos personales.
- **Alternativas descartadas:** mezclar intentos con el historial general; registrar solo el envío exitoso; no conservar evidencia de fallos.
- **Documentos afectados:** `RN-REC-04`, `F-OPS-04`, `F-NOT-02`, `RN-DAT-02`, `glosario.md`.
- **Fuente:** `respuesta-propuestas-oc.txt`, líneas 42–44.

### DEC-016 · Vocabulario de usuario y vocabulario técnico

- **Fecha:** 2026-08-05.
- **Decisión:** la interfaz, las notificaciones y la comunicación comercial usan **turno**, por ser la voz coloquial del mercado. La documentación técnica, API, datos y código conservan **cita** / `appointment` como nombre estable de la entidad.
- **Motivo:** adoptar la palabra pedida por el propietario sin introducir una migración puramente nominal ni mezclar nombres técnicos.
- **Límite de interpretación:** la respuesta pidió “usar turnos” por sonar coloquial; esa motivación se aplica a los puntos de contacto con el usuario, no a identificadores internos.
- **Fuente:** `respuesta-dudas-pendientes.txt`, línea 1.

### DEC-017 · Estados, correcciones y cierre de citas

- **Fecha:** 2026-08-05.
- **Decisión:** `no_show` forma parte del P0. El barbero puede corregir un estado registrado por error mediante una operación auditada que conserva el valor anterior. La barbería elige entre cierre manual asistido de citas vencidas o cierre automático después de X horas; el modo y X son configurables.
- **Regla de integridad:** la corrección cambia una clasificación terminal por otra. Una cita cancelada que deba volver a estar activa se reemplaza por una nueva cita tras validar disponibilidad. El cierre automático registra como actor al sistema y como hora efectiva `ends_at`; nunca reescribe el intervalo planificado.
- **Motivo:** obtener métricas fiables sin dejar errores de operación irreparables y respetar la preferencia de cada barbero.
- **Fuente:** `respuesta-dudas-pendientes.txt`, líneas 2 y 22–31.

### DEC-018 · Valores iniciales configurables

- **Fecha:** 2026-08-05.
- **Decisión:** cancelación pública: 20 minutos; anticipación mínima: 60 minutos; ventana máxima: 3 días; rejilla: 15 minutos; recordatorio: 30 minutos antes; cantidad: 1 por defecto, configurable entre 0 y 3; aviso de inasistencia: configurable y desactivado por defecto.
- **Motivo:** convertir las respuestas y rangos en una configuración inicial reproducible, sin quitar al barbero la facultad de ajustarla.
- **Límite de interpretación:** se eligen 3 días dentro del rango “1 a 3” porque permite el ejemplo explícito de reservar miércoles para viernes.
- **Fuente:** `respuesta-dudas-pendientes.txt`, líneas 3–10, 51 y 99–104.

### DEC-019 · Modelo multi-barbero y selección pública

- **Fecha:** 2026-08-05.
- **Decisión:** una barbería puede tener uno o varios barberos desde el MVP. La disponibilidad y la exclusión de cruces operan por barbero. El cliente elige barbero cuando hay varios; cuando hay uno, queda preseleccionado sin un paso adicional.
- **Motivo:** el piloto puede ocurrir con un independiente o con una barbería de cuatro barberos y el modelo no debe cambiar entre ambos casos.
- **Fuente:** `respuesta-dudas-pendientes.txt`, líneas 11–13 y 32.

### DEC-020 · Calendarios e intervalos flexibles

- **Fecha:** 2026-08-05.
- **Decisión:** cada barbero puede activar el calendario de festivos colombianos para bloquearlos por defecto y abrir manualmente los que trabaje. Los bloqueos admiten instancias puntuales, recurrencias y listas explícitas de fechas con excepciones. Una cita puede cruzar medianoche si cabe completa en el horario laboral configurado. Todo intervalo es semiabierto `[inicio, fin)`.
- **Motivo:** representar tanto horarios habituales como barberías que trabajan de noche, sin falsos cruces entre turnos consecutivos.
- **Fuente:** `respuesta-dudas-pendientes.txt`, líneas 14–17 y 60–73.

### DEC-021 · Gestión manual de retrasos

- **Fecha:** 2026-08-05.
- **Decisión:** ante un retraso, el barbero decide atender, reprogramar o cancelar y puede avisar a los siguientes clientes de la demora. El sistema no mueve automáticamente una cadena de turnos.
- **Motivo:** conservar el juicio operativo del barbero y evitar cambios masivos que el cliente no aceptó.
- **Fuente:** `respuesta-dudas-pendientes.txt`, líneas 18–21.

### DEC-022 · Datos de reserva y acceso del cliente

- **Fecha:** 2026-08-05.
- **Decisión:** el formulario público exige nombre, teléfono y correo; la nota es opcional. El nombre de la persona atendida se toma del reservante y solo se pide aparte al reservar para otra persona. Se acepta texto no vacío con un ejemplo claro. La confirmación envía por correo un enlace con token aleatorio largo para consultar o cancelar; un portal por teléfono/correo y un bot quedan fuera del MVP.
- **Motivo:** reunir los datos necesarios para correo, WhatsApp y calendario sin duplicar el nombre en el caso habitual.
- **Fuente:** `respuesta-dudas-pendientes.txt`, líneas 33–50.

### DEC-023 · Stack de implementación

- **Fecha:** 2026-08-05.
- **Decisión:** backend en Go como monolito modular; frontend web en TypeScript; PostgreSQL como persistencia. Los procesos programados usan el mismo código y una tabla transaccional de trabajos, sin microservicios ni infraestructura distribuida en el piloto.
- **Motivo:** bajo consumo de memoria, despliegue simple, mantenimiento por una persona y una interfaz moderna desacoplada del backend.
- **Alternativas descartadas:** microservicios; una pila pesada de aplicación empresarial; elegir un único lenguaje sacrificando el costo operativo pedido.
- **Fuente:** `respuesta-dudas-pendientes.txt`, líneas 52–59.

### DEC-024 · PostgreSQL y aislamiento por fila

- **Fecha:** 2026-08-05.
- **Decisión:** PostgreSQL con esquema compartido, `barbershop_id` obligatorio y seguridad a nivel de fila (RLS). La aplicación fija el contexto de barbería dentro de cada transacción; las políticas de la base de datos son la última defensa.
- **Motivo:** el propietario eligió la estrategia de aislamiento fuerte y pidió que escale a varias barberías.
- **Condición:** índices, consultas selectivas, paginación, planes revisados y pruebas de aislamiento son obligatorios; se detallan en `05-backend/base-datos.md`.
- **Fuente:** `respuesta-dudas-pendientes.txt`, líneas 74–81.

### DEC-025 · Retención y anonimización

- **Fecha:** 2026-08-05.
- **Decisión:** los datos personales se conservan 24 meses por defecto, con valor parametrizado en base de datos. Al vencer el plazo o prosperar una solicitud válida, se anonimizan nombre, teléfono y correo, conservando cita, tiempos, servicio, estado e historial operativo.
- **Motivo:** equilibrar constancia de uso y minimización de datos.
- **Límite:** requiere revisión jurídica colombiana; la parametrización no permite ampliar el plazo sin base legal documentada.
- **Fuente:** `respuesta-dudas-pendientes.txt`, líneas 82–85 y 127.

### DEC-026 · Autenticación, abuso e incidentes

- **Fecha:** 2026-08-05.
- **Decisión:** el barbero accede con correo y contraseña, mantiene una sesión de larga duración y recupera el acceso mediante código al teléfono verificado. El formulario aplica un límite inicial de 5 solicitudes por IP en la ventana configurada y, al superarlo, exige verificación telefónica. Antes del piloto existe un procedimiento escrito para contener, evaluar, avisar y registrar incidentes.
- **Motivo:** mantener baja fricción y elevar la prueba solo cuando exista señal de abuso.
- **Fuente:** `respuesta-dudas-pendientes.txt`, líneas 86–93.

### DEC-027 · Canales de notificación

- **Fecha:** 2026-08-05.
- **Decisión:** el MVP soporta correo electrónico y la interfaz oficial de WhatsApp. Cada barbería habilita uno o ambos y define el canal por tipo de evento. No se usan integraciones no oficiales.
- **Motivo:** ajustarse a la operación real del barbero sin depender de un único canal.
- **Consecuencia:** nombre, teléfono y correo son obligatorios en la reserva pública; las credenciales, plantillas y costos del proveedor se verifican antes del piloto.
- **Fuente:** `respuesta-dudas-pendientes.txt`, líneas 94–98.

### DEC-028 · Condiciones operativas y comerciales del piloto

- **Fecha:** 2026-08-05.
- **Decisión:** piloto gratuito de 4 semanas para 2 o 3 participantes; puede extenderse otras 4 semanas si la madurez lo exige. Los participantes reciben 3 meses de membresía gratuita después. El método anterior se mantiene solo la primera semana. Si el sistema se suspende, el propietario entrega manualmente las citas futuras y pendientes.
- **Motivo:** reducir el riesgo inicial sin mantener una doble agenda que impida medir adopción.
- **Fuente:** `respuesta-dudas-pendientes.txt`, líneas 105–117 y 121–125.

### DEC-029 · Precio y compensación diferidos

- **Fecha:** 2026-08-05.
- **Decisión:** no se fija precio antes del piloto; al terminar se calcula a partir de costos y utilidad. La compensación del colaborador se define por escrito como porcentaje o monto fijo antes del primer pago o cliente referido remunerado.
- **Estado:** diferida con condición de reapertura; no autoriza comisión, cobro ni función de referidos en el MVP.
- **Fuente:** `respuesta-dudas-pendientes.txt`, líneas 114–120.

### DEC-030 · Requisitos legales mínimos del piloto

- **Fecha:** 2026-08-05.
- **Decisión:** antes de publicar el flujo se redactan un aviso visible y una política de tratamiento adaptada al proyecto, y se firma un acuerdo breve con el colaborador. Antes de producción comercial, una persona con criterio jurídico colombiano revisa esos textos.
- **Motivo:** no recoger datos reales ni operar responsabilidades sobre la base de un texto genérico sin adaptar.
- **Advertencia:** esta decisión es un requisito de producto, no asesoría legal.
- **Fuente:** `respuesta-dudas-pendientes.txt`, líneas 126 y 128.

### DEC-031 · Respaldo, alojamiento y soporte

- **Fecha:** 2026-08-05.
- **Decisión:** copias diarias con 30 días de retención y restauración probada; servicios gestionados gratuitos para desarrollo y piloto; VPS dedicado económico al pasar a producción. El propietario responde en menos de una hora durante la primera semana y el mismo día después, y atiende cualquier alerta operativa.
- **Límite de interpretación:** “VPN dedicado” se normaliza a “VPS dedicado”, porque el contexto se refiere al servidor que aloja la aplicación.
- **Fuente:** `respuesta-dudas-pendientes.txt`, líneas 129–133.

### DEC-032 · Recordatorios automáticos en P0

- **Fecha:** 2026-08-05.
- **Decisión:** `F-NOT-03` se eleva a P0 y se construye junto con `F-NOT-01` y `F-NOT-02`. La programación se crea o actualiza en la misma operación transaccional que crea o modifica la cita; un trabajador solo ejecuta lo ya programado.
- **Motivo:** las notificaciones y los recordatorios deben existir desde el inicio y no puede haber maquinaria de regeneración sin recordatorios reales.
- **Fuente:** `respuesta-dudas-pendientes.txt`, líneas 134–136.
- **Contradicción resuelta:** `CT-001`.

### DEC-033 · Framework y herramientas del frontend

- **Fecha:** 2026-08-05.
- **Decisión:** el frontend usa Vue 3, TypeScript y Vite. Se adopta Composition API con `<script setup>`, Vue Router y carga diferida por ruta. El estado permanece local o en composables; Pinia solo se incorpora cuando exista estado compartido real.
- **Motivo:** priorizar velocidad de desarrollo y mantenimiento por una persona, con runtime contenido y sin la estructura inicial de Angular ni las decisiones adicionales que exige React como biblioteca.
- **Límites:** sin Nuxt, SSR, microfrontends, biblioteca visual pesada ni soporte offline completo en el MVP.
- **Documentos afectados:** `04-arquitectura/stack-despliegue-operacion.md`, `04-arquitectura/frontend.md`, `matriz-trazabilidad.md`.
- **Fuente:** instrucción directa del propietario del 2026-08-05.

### DEC-034 · Router HTTP del backend

- **Fecha:** 2026-08-05.
- **Decisión:** el backend usa Chi v5 como router y compositor de middleware sobre `net/http`. Los handlers siguen siendo `http.Handler` estándar y los módulos de dominio, servicios y repositorios no dependen de Chi.
- **Motivo:** `net/http` puro ya cubre routing básico, pero Chi reduce el código repetitivo de grupos y cadenas de middleware necesarios para rutas públicas, autenticación, tenant/RLS, rate limiting e idempotencia sin introducir un framework pesado.
- **Alternativas descartadas:** `net/http.ServeMux` sin router adicional, por mayor composición manual; Gin, Echo, Fiber o MVC completo, por aportar una superficie mayor que la necesaria para este monolito.
- **Límites:** Chi solo se usa en la capa HTTP; no se incorpora ORM, contenedor de inyección ni abstracción de contexto propia.
- **Documentos afectados:** `04-arquitectura/stack-despliegue-operacion.md`, `04-arquitectura/backend-go.md`, `matriz-trazabilidad.md`.
- **Fuente:** instrucción directa del propietario del 2026-08-05.

### DEC-035 · Estándares de ingeniería y calidad

- **Fecha:** 2026-08-05.
- **Decisión:** el desarrollo aplica código limpio adaptado al MVP, organización por capacidades, documentación de contratos e invariantes, pruebas según riesgo y diseño PostgreSQL en tercera forma normal por defecto. El repositorio se organiza en `apps/api`, `apps/web` y `database`; la integración de datos usa PostgreSQL real y los recorridos P0 conservan pruebas E2E.
- **Motivo:** agenda, concurrencia, historial y multi-tenancy exigen reglas verificables desde el primer incremento, pero el tamaño del proyecto no justifica capas, abstracciones o suites costosas sin riesgo real.
- **Alternativas descartadas:** estilo decidido por cada archivo; paquetes globales por capa técnica; pruebas solo E2E o solo unitarias; cobertura porcentual como única señal; esquema desnormalizado por conveniencia; migraciones mutables.
- **Límites:** no impone una cantidad fija de líneas ni una abstracción por archivo; permite excepciones justificadas sin relajar seguridad, integridad tenant-aware ni reglas P0. No crea todavía código, migraciones ni casos automatizados.
- **Documentos afectados:** `03-desarrollo/estandar-backend-go.md`, `03-desarrollo/estandar-frontend-vue.md`, `03-desarrollo/estrategia-pruebas.md`, `05-backend/estandar-base-datos.md`, `04-arquitectura/backend-go.md`, `04-arquitectura/frontend.md`, `04-arquitectura/stack-despliegue-operacion.md`, `AGENTS.md`, `matriz-trazabilidad.md`.
- **Fuente:** instrucción directa del propietario del 2026-08-05.

### DEC-036 · Gestión y trazabilidad de migraciones con Atlas

- **Fecha:** 2026-08-06.
- **Decisión:** Atlas CLI administra migraciones SQL versionadas. El directorio y `atlas.sum` se versionan en Git; Atlas registra versiones aplicadas en `atlas_schema_revisions`; CI valida y promueve exactamente el mismo artefacto entre ambientes. Las migraciones corren en una etapa separada con rol migrador, no al iniciar el API o worker.
- **Motivo:** el proyecto necesita SQL explícito para RLS, exclusiones, funciones y roles, junto con integridad del historial, `dry-run`, estado por ambiente y automatización reproducible sin introducir un servicio permanente.
- **Alternativas descartadas:** Goose y `golang-migrate`, por ofrecer un flujo más limitado de aplicación/versión; Flyway, por mayor peso operativo para el MVP; cambios manuales o migraciones ejecutadas por la aplicación, por debilitar privilegios y trazabilidad.
- **Límites:** el MVP no depende de Atlas Cloud, Atlas Pro, migraciones declarativas directas a producción ni `down` automático. Linting avanzado, drift administrado o aprobaciones de Atlas requieren evaluación posterior de costo y salida.
- **Documentos afectados:** `05-backend/migraciones-atlas.md`, `05-backend/estandar-base-datos.md`, `05-backend/base-datos.md`, `03-desarrollo/estrategia-pruebas.md`, `04-arquitectura/stack-despliegue-operacion.md`, `AGENTS.md`, `matriz-trazabilidad.md`.
- **Fuente:** instrucción directa del propietario del 2026-08-06 y evaluación de documentación oficial vigente de Atlas, Goose, `golang-migrate` y Flyway.

### DEC-037 · Estándar OpenAPI y gobierno del contrato HTTP

- **Fecha:** 2026-08-06.
- **Decisión:** el API usa OpenAPI 3.1.2 en YAML multiarchivo y enfoque contract-first. Redocly CLI fijado en el lockfile realiza lint, bundle y documentación. Las rutas usan `/api/v1`; los errores siguen RFC 9457; cada operación declara seguridad, ejemplos, idempotencia y trazabilidad `RN-*`/`DEC-*` cuando corresponda.
- **Motivo:** frontend y backend necesitan un contrato único, validable y generable antes de implementar 45 funciones P0. OpenAPI 3.1.2 conserva JSON Schema 2020-12 y soporte estable de la herramienta de documentación elegida.
- **Alternativas descartadas:** OpenAPI 3.2 inmediato, porque la generación HTML obligatoria todavía no declara soporte completo; OpenAPI 3.0, por su modelo de schemas anterior; generar la especificación desde comentarios Go o mantener documentación manual paralela, por riesgo de deriva.
- **Límites:** esta decisión crea el estándar, no el contrato concreto ni sus endpoints. El mecanismo exacto de autenticación y las herramientas de codegen/verificación de handlers se seleccionan al diseñar la primera entrega, siempre desde el bundle OpenAPI.
- **Documentos afectados:** `06-api/estandar-openapi.md`, `04-arquitectura/backend-go.md`, `04-arquitectura/frontend.md`, `04-arquitectura/stack-despliegue-operacion.md`, `03-desarrollo/estandar-backend-go.md`, `03-desarrollo/estandar-frontend-vue.md`, `03-desarrollo/estrategia-pruebas.md`, `AGENTS.md`, `matriz-trazabilidad.md`.
- **Fuente:** instrucción directa del propietario del 2026-08-06 y documentación oficial vigente de OpenAPI, Redocly y RFC 9457.

### DEC-038 · Flujo de ramas, commits y pull requests

- **Fecha:** 2026-08-06.
- **Decisión:** el proyecto usa GitHub Flow con `main` como única rama permanente, ramas cortas por issue, Conventional Commits y pull request obligatorio. El método de integración es squash; el título del PR forma el commit final. `main` debe bloquear force push y eliminación, exigir historial lineal, conversaciones resueltas y los checks disponibles. Mientras trabaje una sola persona se requieren cero aprobaciones, pero al incorporarse otro colaborador se exige al menos una aprobación ajena al último cambio.
- **Motivo:** el MVP necesita una historia simple, revisable y fácil de revertir. Ramas `develop`, por ambiente o por versión aumentarían la coordinación y el riesgo de divergencia sin existir versiones soportadas en paralelo.
- **Alternativas descartadas:** Git Flow con `develop` y `release/*`; commits directos a `main`; merge commits; ramas `staging` o `production`; exigir una aprobación ficticia durante el desarrollo individual.
- **Límites:** la decisión define el estándar y crea plantillas locales, pero no inicializa Git, crea el repositorio remoto ni configura reglas en GitHub. Ruleset o protección equivalente, checks, responsables reales de `CODEOWNERS` y firma obligatoria se activan cuando exista el repositorio y las identidades correspondientes.
- **Documentos afectados:** `03-desarrollo/flujo-git-github.md`, `CONTRIBUTING.md`, `.github/PULL_REQUEST_TEMPLATE.md`, `.github/ISSUE_TEMPLATE/`, `.gitattributes`, `.gitignore`, `04-arquitectura/stack-despliegue-operacion.md`, `AGENTS.md`, `matriz-trazabilidad.md`.
- **Fuente:** instrucción directa del propietario del 2026-08-06 y documentación oficial vigente de GitHub Flow, rulesets, squash merge, issues, pull requests y Conventional Commits 1.0.0.

### DEC-039 · Sistema visual y composición de pantallas

- **Fecha:** 2026-08-06.
- **Decisión:** el frontend adopta un sistema visual ligero y obligatorio, basado en tokens semánticos, con paleta azul marino, cobre y neutros, escala espacial de 4 px, tipografía del sistema y composición móvil primero. El flujo público y el panel comparten componentes, estados y patrones de pantalla; el objetivo mínimo de accesibilidad es WCAG 2.2 nivel AA y el objetivo táctil ordinario es 44 × 44 px.
- **Motivo:** el barbero opera desde un celular de gama media y el cliente debe reservar con rapidez; permitir estilos por módulo produciría diferencias de color, tamaño e interacción que aumentarían errores y mantenimiento.
- **Alternativas descartadas:** decidir estilos pantalla por pantalla; incorporar una biblioteca visual pesada antes de validar los flujos; permitir colores o CSS libres por barbería; construir modo oscuro y múltiples temas durante el MVP.
- **Límites:** el MVP usa un único tema claro. Cada barbería puede mostrar nombre y logotipo, pero no sustituir colores, tipografía, radios o estados semánticos. La decisión fija el estándar, no crea todavía componentes Vue, logotipo, maquetas finales ni una dependencia de iconos.
- **Documentos afectados:** `03-desarrollo/estandar-diseno-visual.md`, `03-desarrollo/estandar-frontend-vue.md`, `03-desarrollo/estrategia-pruebas.md`, `04-arquitectura/frontend.md`, `AGENTS.md`, `CONTRIBUTING.md`, `matriz-trazabilidad.md`.
- **Fuente:** instrucción directa del propietario del 2026-08-06 y WCAG 2.2 del W3C.

### DEC-040 · Modelo de roles PostgreSQL y aprovisionamiento

- **Fecha:** 2026-08-11.
- **Decisión:** un propietario de esquema/objetos `NOLOGIN` controla tablas, índices, secuencias y funciones. `barberia_migrator` es el único rol de login que ejecuta Atlas y no posee objetos por defecto. `barberia_app` (API) permanece tenant-scoped, sin superusuario, `CREATEDB`, `CREATEROLE` ni `BYPASSRLS`. Se crea `barberia_worker`, separado del API, sin acceso general de handlers y sin `BYPASSRLS`, exclusivo para reclamar/finalizar trabajos globales (notificaciones, retención). Los tres roles se aprovisionan con un administrador **antes** de que Atlas se conecte por primera vez; ninguna migración crea el rol con el que ya está conectada.
- **Responsable:** propietario del proyecto, a partir de la recomendación técnica de `docs/05-backend/revision-ddl-seguridad-2026-08-11.md`.
- **Motivo:** `DDL-SEC-01` mostró que la primera migración no controla su propio ownership de forma reproducible; `DDL-SEC-02` y `DDL-SEC-04` mostraron que API y worker comparten privilegios más amplios de lo necesario, ampliando el impacto de una inyección SQL o una credencial comprometida.
- **Alternativas descartadas:** dejar que el primer administrador que se conecte sea dueño accidental de los objetos; un único rol compartido entre API y worker; usar el rol migrador también como propietario de login.
- **Documentos afectados:** `05-backend/estandar-base-datos.md`, `05-backend/migraciones-atlas.md`, una migración correctiva roll-forward (no editar `20260807170000`), documentación de bootstrap de despliegue.
- **Fuente:** `docs/05-backend/revision-ddl-seguridad-2026-08-11.md`, hallazgos `DDL-SEC-01`, `DDL-SEC-02`, `DDL-SEC-04`; aprobación explícita del propietario el 2026-08-11.

### DEC-041 · Vocabulario de `appointment_history.event_type`

- **Fecha:** 2026-08-11.
- **Decisión:** `event_type` usa valores en inglés y `snake_case`, con el prefijo de la entidad: `appointment_created`, `appointment_rescheduled`, `appointment_service_changed`, `appointment_completed`, `appointment_cancelled_by_customer`, `appointment_cancelled_by_barber`, `appointment_no_show`, `appointment_status_corrected`.
- **Responsable:** propietario del proyecto.
- **Motivo:** `docs/02-requisitos/estados-citas.md` sección 9 documentaba valores en español con puntos (`cita.creada`, `cita.cancelada_por_cliente`...) mientras la sección 11 exige "valores en inglés y minúsculas", igual que la columna `status`. Se elige el vocabulario en inglés para no crear una excepción de idioma dentro del mismo esquema.
- **Alternativas descartadas:** conservar el vocabulario en español de la sección 9 y reescribir la regla general de la sección 11 para excluir `event_type`.
- **Documentos afectados:** `02-requisitos/estados-citas.md` (sección 9, corregir los ocho valores), `database/modelo-fisico-referencia.sql` (`appointment_history_event_type_ck`), `contradicciones.md` (`CT-002`).
- **Fuente:** hallazgo sin código propio en `docs/05-backend/revision-ddl-seguridad-2026-08-11.md`, sección 5, punto 2; aprobación explícita del propietario el 2026-08-11.

### DEC-042 · Fecha ancla de retención y anonimización

- **Fecha:** 2026-08-11.
- **Decisión:** los 24 meses de `DEC-025` se cuentan desde la última actividad del cliente: la más reciente entre la fecha de su última cita (creación o última transición registrada) y su último contacto directo (por ejemplo, una solicitud de cancelación o acceso por token). Un cliente sin actividad reciente en ninguna barbería vence y puede anonimizarse; un cliente activo nunca vence mientras siga interactuando.
- **Responsable:** propietario del proyecto, sujeta a revisión jurídica colombiana como toda `DEC-025`.
- **Motivo:** `DEC-025` fijó el plazo pero no el punto de partida; contarlo solo desde la creación penalizaría a clientes recurrentes antiguos y contarlo solo desde la última cita ignoraría otras interacciones legítimas (cancelaciones, consultas por token).
- **Alternativas descartadas:** contar únicamente desde la creación del registro de `customer`; contar únicamente desde la última cita, ignorando otro tipo de interacción.
- **Documentos afectados:** `01-producto/reglas-negocio.md` (regla de retención), `database/modelo-fisico-referencia.sql` (cálculo de vencimiento), matriz de anonimización derivada de `DEC-049`.
- **Fuente:** `docs/05-backend/revision-ddl-seguridad-2026-08-11.md`, sección 5, punto 3; aprobación explícita del propietario el 2026-08-11.

### DEC-043 · Semántica concurrente de idempotencia

- **Fecha:** 2026-08-11.
- **Decisión:** una segunda solicitud concurrente con la misma clave de idempotencia intenta un `pg_try_advisory_xact_lock` derivado de `(tenant, hash de la clave)`. Si no puede tomar el lock de inmediato, responde `409 Conflicto` (o `425 Too Early` si el contrato OpenAPI lo prefiere) sin esperar a que la primera transacción termine. No hay espera acotada ni reproducción del resultado de otra transacción.
- **Responsable:** propietario del proyecto.
- **Motivo:** `DDL-IDEM-01` mostró que la restricción única actual hace que la segunda inserción espere sin límite la resolución de la primera transacción, violando la regla de transacciones cortas y la semántica documentada de `RN-IDE-01`/`estados-citas.md` sección 10.
- **Alternativas descartadas:** espera acotada con `lock_timeout` y réplica del resultado confirmado, descartada por añadir latencia y complejidad de reintento en el cliente sin beneficio claro para el volumen esperado del piloto.
- **Documentos afectados:** `01-producto/reglas-negocio.md` (`RN-IDE-01`), `02-requisitos/estados-citas.md` (sección 10), migración correctiva de `20260807170100_create_idempotency_record.sql` (roll-forward), `06-api/estandar-openapi.md` si se adopta `425`.
- **Fuente:** `docs/05-backend/revision-ddl-seguridad-2026-08-11.md`, hallazgo `DDL-IDEM-01`; aprobación explícita del propietario el 2026-08-11.

### DEC-044 · Versión mínima de PostgreSQL

- **Fecha:** 2026-08-11.
- **Decisión:** se confirma formalmente PostgreSQL 14 como versión mayor mínima soportada, sin cambio sobre lo ya exigido por `20260807170000_create_tenant_foundation.sql` y `apps/api/README.md`.
- **Responsable:** propietario del proyecto.
- **Motivo:** el código ya impone ese piso; fijar 16 o superior obligaría a reescribir una migración inmutable sin beneficio demostrado y podría reducir opciones de hospedaje económico (`DEC-031`).
- **Alternativas descartadas:** elevar el piso a PostgreSQL 16.
- **Documentos afectados:** ninguno requiere cambio; queda como decisión formal de referencia para el preflight de `docs/10-backlog/prompt-endurecimiento-ddl.md`.
- **Fuente:** `docs/05-backend/revision-ddl-seguridad-2026-08-11.md`, sección 5, punto 1; aprobación explícita del propietario el 2026-08-11.

### DEC-045 · Identidad y reutilización de `customer` por teléfono

- **Fecha:** 2026-08-11.
- **Decisión:** el teléfono identifica de forma única a un cliente dentro de cada barbería. Una reserva pública hace upsert por `(barbershop_id, phone)`: si el teléfono ya existe en esa barbería, actualiza nombre y correo del cliente existente en lugar de crear una fila nueva.
- **Responsable:** propietario del proyecto.
- **Motivo:** evita duplicar clientes por cada reserva y refleja que, dentro de una misma barbería, el mismo número suele corresponder a la misma persona.
- **Alternativas descartadas:** crear un cliente nuevo en cada reserva sin unicidad de teléfono, que multiplicaría filas y dificultaría el historial de un cliente recurrente.
- **Documentos afectados:** `01-producto/reglas-negocio.md`, `database/modelo-fisico-referencia.sql` (restricción única `(barbershop_id, phone)` y lógica de upsert).
- **Fuente:** `docs/05-backend/revision-ddl-seguridad-2026-08-11.md`, hallazgo `DDL-BIZ-03`; aprobación explícita del propietario el 2026-08-11.

### DEC-046 · Unicidad de correo de `customer`

- **Fecha:** 2026-08-11.
- **Decisión:** la unicidad de correo de `customer` es por barbería, no global. La misma persona puede ser cliente de varias barberías con el mismo correo sin conflicto.
- **Responsable:** propietario del proyecto.
- **Motivo:** coherente con el aislamiento fuerte por RLS de `DEC-024`: cada barbería es un tenant independiente y no existe todavía un concepto de identidad de cliente compartida entre barberías.
- **Alternativas descartadas:** unicidad global de correo, descartada porque introduciría acoplamiento entre tenants que `DEC-024` no contempla y complicaría la anonimización por barbería.
- **Documentos afectados:** `database/modelo-fisico-referencia.sql` (restricción única `(barbershop_id, lower(email))`).
- **Fuente:** `docs/05-backend/revision-ddl-seguridad-2026-08-11.md`, hallazgo `DDL-BIZ-03`; aprobación explícita del propietario el 2026-08-11.

### DEC-047 · Alcance de `barber` recortado a `HU-021`

- **Fecha:** 2026-08-11.
- **Decisión:** el modelo físico de `barber` se recorta a lo que `HU-021` autoriza: identificador, tenant, nombre y timestamps, con alta, listado y renombrar. Se retiran del modelo de referencia el borrado, la desactivación, el orden manual y el vínculo automático con `staff_user` hasta que una historia futura los apruebe explícitamente.
- **Responsable:** propietario del proyecto.
- **Motivo:** `DDL-BIZ-02` encontró que el modelo de referencia incluye un ciclo de vida completo que `HU-021` no pidió ("Como barbero, quiero registrar, renombrar y consultar..."), y `HU-021` bloquea explícitamente `DELETE`.
- **Alternativas descartadas:** mantener el ciclo de vida completo y registrar una ampliación de alcance de `HU-021` ahora, descartada porque el propietario prefiere no comprometerse a ese alcance antes de necesitarlo.
- **Documentos afectados:** `database/modelo-fisico-referencia.sql` (tabla `barber`), `02-requisitos/historias-usuario.md` (sin cambio, ya refleja este alcance).
- **Fuente:** `docs/05-backend/revision-ddl-seguridad-2026-08-11.md`, hallazgo `DDL-BIZ-02`; aprobación explícita del propietario el 2026-08-11.
- **Actualización (2026-10-09):** `DEC-100` levantó parcialmente la restricción sobre el vínculo con `staff_user`: existe `barber.staff_user_id` opcional, que solo el propio usuario establece (migración `20261009120000_add_barber_staff_user_link.sql`, issue [#322](https://github.com/bcaceres19/barberia/issues/322)). Borrado, desactivación y orden manual siguen fuera.

### DEC-048 · Segundo y tercer recordatorio

- **Fecha:** 2026-08-11.
- **Decisión:** cuando la barbería configura más de un recordatorio, el segundo se envía 24 horas antes de la cita y el tercero 2 horas antes, además del primero a 30 minutos (`DEC-018`). El orden es 1 → 30 minutos, 2 → 24 horas, 3 → 2 horas, contado hacia atrás desde `starts_at`.
- **Responsable:** propietario del proyecto.
- **Motivo:** `DEC-018` fijó la cantidad (0 a 3, por defecto 1) y el primer recordatorio, pero no los valores del segundo y tercero; `DDL-BIZ-01` señaló que sin backfill ni valores por defecto, cero filas de recordatorio equivale a ninguno.
- **Alternativas descartadas:** ninguna alternativa concreta fue solicitada por el propietario; se adoptó la recomendación técnica directamente.
- **Documentos afectados:** `01-producto/reglas-negocio.md`, `database/modelo-fisico-referencia.sql` (backfill de la regla ordinal 1/30 y valores 2/1440 y 3/120 en minutos).
- **Fuente:** `docs/05-backend/revision-ddl-seguridad-2026-08-11.md`, hallazgo `DDL-BIZ-01`; aprobación explícita del propietario el 2026-08-11.

### DEC-049 · Matriz completa de anonimización

- **Fecha:** 2026-08-11.
- **Decisión:** la anonimización de `DEC-025` no se limita a `customer.full_name/phone/email`. Cubre también `appointment.attendee_name` y `customer_note`, `appointment_history.reason` y los valores anterior/nuevo de `appointment_history_change` que contengan datos personales, revoca y/o anonimiza `appointment_access_token`, y redacta o purga `idempotency_record.response_body` y cualquier identificador de proveedor que permita correlacionar con la persona. Se conservan cita, tiempos, servicio, estado e historial operativo, conforme al texto original de `DEC-025`.
- **Responsable:** propietario del proyecto, sujeta a revisión jurídica colombiana como toda `DEC-025`.
- **Motivo:** `DDL-PRI-01` mostró que una anonimización limitada a `customer` deja identificable a la persona a través de otras copias de su nombre y de texto libre asociado a la cita, incumpliendo el espíritu de minimización de `DEC-025` y `RN-DAT-03`.
- **Alternativas descartadas:** anonimizar solo `customer` conforme al texto literal de `DEC-025`, descartada por dejar la anonimización incompleta según el propio hallazgo de la revisión.
- **Documentos afectados:** `01-producto/reglas-negocio.md` (`RN-DAT-03`), `database/modelo-fisico-referencia.sql` (función de anonimización), diseño de worker de retención (Fase 5 de `docs/10-backlog/prompt-endurecimiento-ddl.md`).
- **Fuente:** `docs/05-backend/revision-ddl-seguridad-2026-08-11.md`, hallazgo `DDL-PRI-01`; aprobación explícita del propietario el 2026-08-11.

### DEC-050 · Mecanismo y duración de la sesión larga del barbero

- **Fecha:** 2026-08-11.
- **Decisión:** la sesión del barbero se sostiene con una cookie `HttpOnly`, `Secure` y `SameSite` que contiene un token opaco y aleatorio. El token se almacena con su hash en `staff_session` (`token_hash`, `revoked_at`, `expires_at`), nunca en claro. Dura 30 días desde la última actividad, con renovación deslizante en cada uso; revocar (cierre de sesión, cambio de contraseña, incidente) es un `UPDATE` que fija `revoked_at`, efectivo de inmediato.
- **Responsable:** propietario del proyecto.
- **Motivo:** `DEC-026` confirmó una "sesión de larga duración" sin fijar mecanismo ni plazo; `CA-006-02` exige que sea revocable, lo que descarta un token firmado sin registro server-side. Un token opaco revocable en base de datos es el único mecanismo de los evaluados que cumple la revocación inmediata.
- **Alternativas descartadas:** cookie con token opaco pero expiración fija de 90 días sin renovación (obliga a reautenticar cada 3 meses incluso con uso diario, sin beneficio de seguridad claro sobre la renovación deslizante); JWT firmado sin revocación inmediata en base de datos (incumple `CA-006-02`).
- **Documentos afectados:** `01-producto/reglas-negocio.md`, `database/modelo-fisico-referencia.sql` (sección `staff_session`, ya diseñada con los campos necesarios), `dudas-pendientes.md` (cierra `DP-SEG-04`).
- **Fuente:** `dudas-pendientes.md` §2 bis, `DP-SEG-04`; aprobación explícita del propietario el 2026-08-11.

### DEC-051 · Canal del código de recuperación de acceso

- **Fecha:** 2026-08-11.
- **Decisión:** el código de recuperación de acceso se envía por WhatsApp oficial y por correo, reutilizando el mismo proveedor y la misma integración ya habilitados por `DEC-027` para notificaciones. No se agrega un proveedor de SMS.
- **Responsable:** propietario del proyecto.
- **Motivo:** evitar sumar un proveedor, un costo y una verificación previa al piloto adicionales; el barbero ya cuenta con WhatsApp y correo como canales de contacto configurados desde `DEC-027`.
- **Límite de interpretación:** `DEC-026` fijaba el código "al teléfono verificado"; esta decisión lo extiende explícitamente a que el mismo código también llegue por correo, como canal adicional, no exclusivo del teléfono.
- **Alternativas descartadas:** SMS con un proveedor nuevo (agrega costo/proveedor sin necesidad, dado que WhatsApp oficial ya está disponible); WhatsApp único sin respaldo por correo (el propietario pidió explícitamente ambos canales).
- **Documentos afectados:** `01-producto/reglas-negocio.md`, `database/modelo-fisico-referencia.sql` (sección `staff_recovery_code`), `dudas-pendientes.md` (cierra `DP-SEG-05`).
- **Fuente:** `dudas-pendientes.md` §2 bis, `DP-SEG-05`; aprobación explícita del propietario el 2026-08-11.

### DEC-052 · Ventana y escalamiento del límite de acceso por IP

- **Fecha:** 2026-08-11.
- **Decisión:** el umbral de 5 solicitudes por IP (`DEC-026`) se mide en una ventana deslizante de 15 minutos. Al superarlo, la IP queda sujeta a verificación telefónica obligatoria durante 24 horas antes de volver al comportamiento normal.
- **Responsable:** propietario del proyecto.
- **Motivo:** 15 minutos es la ventana habitual contra fuerza bruta automatizada, corta para frenar intentos en ráfaga sin acumular falsos positivos de tráfico legítimo disperso; 24 horas de exigencia telefónica es suficientemente disuasivo para un atacante sin bloquear repetidamente al mismo barbero en la misma jornada tras resolver una verificación.
- **Alternativas descartadas:** ventana de 1 hora con escalamiento de 1 hora (ambos más indulgentes; más tiempo para que un atacante intente sin activar el control, y el escalamiento se apaga demasiado rápido para disuadir un segundo intento el mismo día).
- **Documentos afectados:** `01-producto/reglas-negocio.md`, `database/modelo-fisico-referencia.sql` (sección `login_throttle`), `dudas-pendientes.md` (cierra `DP-SEG-06`).
- **Fuente:** `dudas-pendientes.md` §2 bis, `DP-SEG-06`; aprobación explícita del propietario el 2026-08-11.

### DEC-053 · Protocolo de lease de `notification_claim_due` y alcance de `retention_claim_due_customers`

- **Fecha:** 2026-08-11.
- **Decisión:** `notification_claim_due` reclama con un `UPDATE ... SET status = 'processing', claim_token, claimed_at, lease_expires_at` sobre un `SELECT ... FOR UPDATE SKIP LOCKED` (lease configurable entre 30 y 3600 segundos, `p_limit` entre 1 y 200), en vez de solo bloquear la fila. La misma llamada recupera leases vencidos, sin una función aparte: su `WHERE` acepta tanto `pending` vencido como `processing` cuyo `lease_expires_at` ya pasó. `notification_finalize_claim` cierra por CAS de `claim_token` con tres desenlaces: `sent` (terminal), `retry` (suelta el lease y vuelve a `pending`, fallo temporal) y `permanent_failure` (pasa a `skipped`, reutilizando el estado que ya existía para RN-REC-06 en vez de crear un estado terminal nuevo; el motivo distingue un caso del otro en `notification_attempt`). Un CAS que no encuentra el token coincidente devuelve `false` sin lanzar excepción. `barberia_app` pierde el privilegio de columna sobre `claim_token`/`claimed_at`/`lease_expires_at`/`sent_at`: solo conserva `UPDATE` de `status`/`cancelled_at` para cancelar: el `CHECK` de forma ya impide cancelar una fila `processing`. `retention_claim_due_customers` NO adopta el mismo protocolo: su efecto (anonimizar) es una escritura SQL ejecutada en la MISMA transacción que la reclama, sin llamada de red de por medio, así que el bloqueo de fila del `SELECT ... FOR UPDATE SKIP LOCKED` basta; solo se le añade la misma validación de `p_limit` (1-200) y `p_now NOT NULL`. Si la anonimización completa (`DEC-049`) termina partiéndose en varias transacciones, esta parte de la decisión se revisa y adopta el mismo protocolo de lease.
- **Responsable:** propietario del proyecto.
- **Motivo:** `DDL-CON-01` mostró que la reclamación original no cambiaba estado ni creaba lease: al confirmar la transacción de reclamo, la fila volvía a estar disponible para otro trabajador, y mantener la transacción abierta durante el envío violaba la regla de transacciones cortas de `estandar-base-datos.md` §11 (nunca llamar correo o WhatsApp dentro de una transacción). `DDL-CON-02` exige poder probar estas garantías con PostgreSQL real y al menos dos conexiones. `DDL-OPS-01` exige que `p_limit` rechace `NULL`, cero, negativos y lotes excesivos en toda función global.
- **Alternativas descartadas:** un estado terminal `failed` nuevo para el fallo permanente, descartado porque el motivo (canal deshabilitado vs. reintentos agotados) ya se distingue en `notification_attempt` y no aporta valor de negocio distinguirlo también en el estado de `notification_schedule`, a costa de ampliar el vocabulario y las pruebas de forma; una función de recuperación de leases separada de `notification_claim_due`, descartada porque el mismo índice parcial y el mismo `WHERE` sirven ambos casos sin duplicar lógica ni añadir una segunda llamada por ciclo del worker; aplicar el mismo protocolo de lease a `retention_claim_due_customers` "por si acaso", descartado porque no hay E/S externa que proteger hoy y añadir columnas de lease sin uso real solo aumenta la superficie a probar.
- **Documentos afectados:** `docs/05-backend/estandar-base-datos.md` (§8, si se documenta el patrón de lease), `database/modelo-fisico-referencia.sql` (`notification_schedule`, `notification_claim_due`, `notification_finalize_claim`, `retention_claim_due_customers`), `database/tests/notification_lease_concurrency.sql`.
- **Fuente:** `docs/05-backend/revision-ddl-seguridad-2026-08-11.md`, hallazgos `DDL-CON-01`, `DDL-CON-02`, `DDL-OPS-01`; issue `#5`.

### DEC-054 · Fórmula de última actividad, marcador de anonimización y alcance de `appointment_history_change`

> `DEC-053` (arriba) se registró en paralelo por el PR del issue #5; ambas decisiones tocan `retention_claim_due_customers` sin contradecirse: `DEC-053` fija que la función no lleva lease, `DEC-054` fija la fórmula concreta de su fecha ancla. Se numeró `DEC-054` para no reutilizar un código ya tomado por el otro PR en curso (regla de este registro: ningún código se reasigna).

- **Fecha:** 2026-08-11.
- **Decisión:** cuatro precisiones necesarias para implementar `DEC-042`/`DEC-049` en SQL, sin las cuales el issue no se puede codificar sin inventar respuesta:
  1. **Fórmula de última actividad (`DEC-042`):** `GREATEST(customer.created_at, MAX(appointment.created_at), MAX(appointment_history.occurred_at) de sus citas, MAX(appointment_access_token.issued_at) de sus citas)`. La emisión del token (`issued_at`) es el mejor proxy disponible de "acceso por token": el esquema no registra el instante en que el cliente de verdad abre el enlace, solo cuándo el sistema lo emite. Añadir esa columna es una ampliación de esquema fuera del alcance de este issue.
  2. **Marcador de anonimización:** el texto literal `'Cliente anonimizado'` sustituye `customer.full_name` y `appointment.attendee_name`, y reemplaza (nunca borra) los valores no nulos de `appointment_history.reason` y de `appointment_history_change.previous_value`/`new_value` cuando el campo es personal. Es literal y único a propósito: `customer_anonymized_ck` lo exige tal cual, así que "¿está anonimizado?" es una comparación exacta, no una convención de texto libre.
  3. **Alcance de `appointment_history_change`:** solo se redactan las filas cuyo `field_name` sea `'attendee_name'` o `'customer_note'` — los únicos dos campos personales de `appointment`. Un cambio de `status`, `starts_at` o `service_id` no identifica a nadie y se conserva íntegro, porque es la evidencia operativa que `DEC-025`/`RN-DAT-03` piden conservar.
  4. **`idempotency_record.response_body` no se toca:** su TTL máximo es 86 400 s (`idempotency_begin`, `p_ttl_seconds` ≤ 1 día) y el mínimo de retención posible es 1 mes (`barbershop_retention_months_ck`). Para cualquier configuración válida, una respuesta idempotente ligada a la actividad de un cliente ya fue purgada por `idempotency_purge_expired` mucho antes de que ese cliente llegue a ser candidato de anonimización. La tabla tampoco tiene columna que la correlacione con `customer_id`; añadirla solo para cubrir un caso ya imposible sería alcance fuera de este issue.
- **Responsable:** propietario del proyecto, sujeta a revisión jurídica colombiana como toda `DEC-025`.
- **Motivo:** `DDL-PRI-01` exigía "aprobar primero una matriz de datos personales y una fecha ancla" antes de codificar; `DEC-042`/`DEC-049` fijaron el principio pero no la fórmula ni el marcador exactos, y AGENTS.md prohíbe inventar esa traducción sin registrarla.
- **Alternativas descartadas:** un estado/columna nuevo en `appointment_access_token` para registrar el acceso real del cliente (descartada por alcance: es una ampliación de esquema, no una corrección de la anonimización); redactar `appointment_history_change` por heurística de contenido en vez de por `field_name` (descartada por indeterminista y no verificable con un `CHECK`); añadir `customer_id` a `idempotency_record` para poder purgarla por cliente (descartada porque el TTL ya lo vuelve innecesario en todo escenario válido).
- **Documentos afectados:** `database/modelo-fisico-referencia.sql` (`customer_anonymized_ck`, `retention_claim_due_customers`, `customer_anonymize`).
- **Fuente:** `docs/05-backend/revision-ddl-seguridad-2026-08-11.md`, hallazgo `DDL-PRI-01`; `DEC-042`, `DEC-049`; issue `#6`.

### DEC-055 · Resolución de `CT-003`: audiencia del inicio de sesión

- **Fecha:** 2026-08-13.
- **Decisión:** el inicio de sesión de `HU-005` se mueve de `/api/v1/private/auth/login` a `/api/v1/public/auth/login`. Es la única operación del módulo `auth` en la audiencia pública; el resto (cierre de sesión de `HU-006`) permanece bajo `/api/v1/private`. `CA-006-04` no necesita ninguna lista de excepciones: como el login ya no vive bajo `/private`, la prueba estructural que inventaría rutas protegidas puede exigir el middleware de sesión en el 100 % de ese subrouter sin excepción alguna.
- **Responsable:** propietario del proyecto.
- **Motivo:** de las tres opciones registradas en `CT-003`, evita mantener y probar indefinidamente una lista cerrada de excepciones de arranque sin sesión dentro de `/private`, y sigue el patrón ya usado en el árbol de rutas (`backend-go.md` §7) de separar audiencias `public`/`private` en vez de introducir una tercera.
- **Alternativas descartadas:** login bajo `/private` con lista cerrada de excepciones (opción 1) — descartada por el costo permanente de mantener y auditar esa lista y por complicar la prueba estructural de `CA-006-04`; prefijo de autenticación separado (opción 3) — descartada por introducir una tercera audiencia a documentar sin necesidad frente a la distinción `public`/`private` ya existente.
- **Documentos afectados:** `docs/02-requisitos/historias-usuario.md` (`HU-005` alcance incluido), `docs/04-arquitectura/backend-go.md` §7 (árbol de rutas), `docs/10-backlog/prompts/hu/hu-005-inicio-sesion.md`, `docs/10-backlog/prompts/hu/hu-006-sesion-persistente.md`.
- **Fuente:** `docs/00-control/contradicciones.md`, `CT-003`; aprobación explícita del propietario el 2026-08-13.

### DEC-056 · Resolución de `CT-004`: destino de `CA-010-01` dividido entre `HU-010` y `HU-012`

- **Fecha:** 2026-08-13.
- **Decisión:** `CA-010-01` se divide entre `HU-010` y `HU-012` (opción 1 de `CT-004`). `HU-010` implementa y verifica que, con credenciales válidas, la aplicación navega a una única ruta privada real y protegida, `/panel`, con un guard mínimo propio de `HU-010` (sin sesión válida redirige al acceso) y contenido mínimo de marcador de posición autenticado — sin cabecera, sin navegación general. `HU-012` reutiliza y generaliza ese guard a todas las rutas privadas en vez de crear uno paralelo o cambiar la ruta, y construye ahí mismo la cabecera, navegación y demás estructura del cascarón descrita en su alcance. La verificación end-to-end de "llegar al panel" completo (cabecera, navegación) queda en `CA-012-01`/`CA-012-02` de `HU-012`, no en `HU-010`.
- **Responsable:** propietario del proyecto.
- **Motivo:** de las tres opciones registradas en `CT-004`, dividir el criterio es el cambio de menor alcance: no amplía `HU-010` más allá de "pantalla de acceso" con responsabilidades de layout/navegación que pertenecen a `HU-012`, y no reordena un backlog B0 ya planificado.
- **Alternativas descartadas:** mover un cascarón privado mínimo a `HU-010` (opción 2) — descartada por ampliar su alcance con responsabilidades de `HU-012`; reordenar y redefinir dependencias (opción 3) — descartada por su mayor impacto en la planificación ya fijada de B0.
- **Documentos afectados:** `docs/02-requisitos/historias-usuario.md` (`HU-010` `CA-010-01` y alcance, `HU-012` alcance), `docs/10-backlog/prompts/hu/hu-010-pantalla-acceso.md`, futuro prompt de `HU-012`.
- **Fuente:** `docs/00-control/contradicciones.md`, `CT-004`; aprobación explícita del propietario el 2026-08-13.

### DEC-057 · Resolución de `DP-SEG-07`: atributos de la cookie de sesión

- **Fecha:** 2026-08-13.
- **Decisión:** la cookie de sesión de `HU-005`/`HU-006` se llama `barberia_session`, con `Path=/api/v1`, `SameSite=Lax`, sin `Domain` explícito (host-only) y `Max-Age` de 30 días, además de `HttpOnly`+`Secure` ya fijados por `DEC-050`. Confirma el valor ya implementado en el PR de `HU-005` (issue `#44`).
- **Responsable:** propietario del proyecto.
- **Motivo:** `Lax` es el valor recomendado para una cookie de sesión de primer nivel que no necesita viajar en navegación cross-site; `Path=/api/v1` cubre tanto el login público como las rutas privadas futuras sin exponerla a otras rutas del mismo host; sin `Domain` explícito, el navegador la ata al host exacto, más restrictivo y sin riesgo de fuga a subdominios no revisados.
- **Alternativas descartadas:** `SameSite=Strict` — descartada por no aportar protección adicional relevante a este flujo y complicar la navegación desde enlaces externos (p. ej. WhatsApp/correo) sin beneficio de seguridad claro; `Domain` explícito — descartado por ampliar innecesariamente el alcance de la cookie a subdominios no auditados.
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-SEG-07`), PR de `HU-005` (issue `#44`).
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-SEG-07`; aprobación explícita del propietario el 2026-08-13.

### DEC-058 · Resolución de `DP-SEG-08`: evidencia de aislamiento dividida entre `HU-005` y `HU-006`

- **Fecha:** 2026-08-13.
- **Decisión:** `CA-005-05` y la parte de `CA-005-01` referida a "solicitudes privadas posteriores" se dan por cumplidos en `HU-005` con evidencia a nivel de PostgreSQL/RLS (aislamiento de `staff_session`/`staff_credential` con dos tenants) y de estructura HTTP, sin exigir un endpoint privado real todavía. La verificación end-to-end contra una operación privada real queda como un nuevo criterio `CA-006-07` de `HU-006`: la cookie de sesión de la barbería A nunca ejecuta el logout (u otra operación privada real) sobre datos de B, probado contra el primer endpoint privado real que `HU-006` construye.
- **Responsable:** propietario del proyecto.
- **Motivo:** mismo patrón que `DEC-056`/`CT-004` — dividir el criterio es el cambio de menor alcance: desbloquea el merge de `HU-005` ya, sin inventar un endpoint de demostración que no tiene aprobación de ninguna fuente, y sin crear una dependencia circular con `HU-006` (que ya depende de `HU-005` integrada).
- **Alternativas descartadas:** esperar a tener un endpoint real antes de mergear `HU-005` — descartada por ser circular (`HU-006` depende de `HU-005` integrada) y bloquear todo el bloque B0 sin una vía de salida; adelantar un endpoint mínimo de demostración dentro de `HU-005` — descartada por invadir el alcance de `HU-006` y arriesgar declarar cumplido un criterio con un endpoint inventado, prohibido explícitamente por el prompt de `HU-005`.
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-SEG-08`), `docs/02-requisitos/historias-usuario.md` (`HU-005` `CA-005-01`/`CA-005-05`, `HU-006` nuevo `CA-006-07`), `docs/10-backlog/prompts/hu/hu-006-sesion-persistente.md`.
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-SEG-08`; aprobación explícita del propietario el 2026-08-13.

### DEC-059 · Resolución de `DP-UX-06`: destino provisional del enlace de recuperación

- **Fecha:** 2026-08-13.
- **Decisión:** el enlace de recuperación de acceso de `CA-010-08` navega a `/recuperar-acceso`, una ruta real y operable, cargada de forma diferida dentro de `modules/auth`. La página declara explícitamente, con un aviso informativo, que la recuperación de acceso todavía no está disponible y que `HU-011` la construirá; no simula el flujo de `HU-011` ni presenta una experiencia de producto terminada. Confirma la interpretación ya implementada en el PR de `HU-010` (issue `#46`).
- **Responsable:** propietario del proyecto.
- **Motivo:** de las salidas posibles, es la que no engaña al barbero (a diferencia de un enlace roto o una simulación del flujo de `HU-011`) ni invade el alcance de `HU-011`, que construirá el flujo real de recuperación.
- **Alternativas descartadas:** un canal de contacto/soporte alterno — descartado por no estar aprobado en ningún `DEC-*` ni tener un canal de soporte definido para el MVP; dejar `CA-010-08` sin cumplir hasta que exista `HU-011` — descartado porque el criterio solo exige un enlace visible y operable, no el flujo completo, y bloquear el merge de `HU-010` por esto sería más costoso que la interpretación honesta ya aplicada.
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-UX-06`), PR de `HU-010` (issue `#46`).
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-UX-06`; aprobación explícita del propietario el 2026-08-13.

### DEC-060 · Resolución de `DP-SEG-09`: operación autoritativa de contexto de sesión para `HU-012`

- **Fecha:** 2026-08-17.
- **Decisión:** se añade una operación privada de solo lectura, `GET /api/v1/private/auth/session`, protegida por `SessionCookie` y montada sobre el mismo `SessionMiddleware` que ya usan las demás rutas de `/api/v1/private` (misma validación y renovación deslizante de `DEC-050`, sin duplicar lógica de sesión). Responde `200` con un payload mínimo — `{ barbershop: { id, name }, expiresAt }` — usando `barbershop.name`, columna ya existente en `barbershop` desde la migración fundacional (`20260807170000_create_tenant_foundation.sql`), aislada por tenant mediante el RLS ya vigente (`DEC-024`); y `401` con el mismo `UnauthorizedProblem` uniforme que `logout` ante cookie ausente, inválida, vencida o revocada. El payload no incluye `staffUserID`, correo ni nombre del barbero: no lo exige ningún criterio de `HU-012` y mantiene el patrón de `auth.Principal` (identificadores opacos, sin datos personales). Esta operación reemplaza el marcador provisional `sessionStorage` de `HU-010` (`DP-SEG-08`) como fuente de rehidratación de `HU-012`; el frontend la llama al abrir o recargar la aplicación, y la cookie `HttpOnly` sigue siendo la única prueba real de sesión en cada solicitud.
- **Responsable:** propietario del proyecto.
- **Motivo:** de las alternativas evaluadas, es la que resuelve `CA-012-01`/`CA-012-04` sin crear una segunda fuente de verdad ni ampliar el alcance de otra HU: reutiliza exactamente la validación de sesión que `HU-006` ya probó, y `barbershop.name` es un dato real y aislado por tenant, no un valor inventado ni un adelanto de la edición que construirá `HU-020`.
- **Alternativas descartadas:** ampliar `LoginResponse` con el contexto de barbería — descartada porque "iniciar sesión" y "¿sigo autenticado ahora?" son preguntas distintas y mezclarlas duplicaría la validación de sesión en dos formas; reutilizar `logout` como lectura — descartada porque `logout` es destructivo por diseño y no debe tener un efecto de solo consulta; mantener el marcador local de `sessionStorage` como fuente de bootstrap — descartada porque no sobrevive al cierre del navegador y no es una fuente autoritativa (`DP-SEG-09`).
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-SEG-09`), `docs/00-control/matriz-trazabilidad.md`, `docs/02-requisitos/historias-usuario.md` (`HU-012` `CA-012-01`/`CA-012-04`), `docs/10-backlog/prompts/hu/hu-012-cascaron-panel-privado.md`, `docs/10-backlog/prompts/README.md`; futura implementación de `HU-012` (issue `#56`) debe crear primero la operación en `api/openapi/paths/private-auth.yaml` contract-first.
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-SEG-09`; aprobación explícita del propietario el 2026-08-17.

### DEC-061 · Resolución de `CT-005`: la escalada ocurre en la sexta solicitud, no en la quinta

- **Fecha:** 2026-08-17.
- **Decisión:** con el valor inicial `p_threshold = 5` (`DEC-026`/`DEC-052`), las cinco primeras solicitudes dentro de la ventana se evalúan con normalidad, incluida la contraseña; la **sexta** solicitud es la que exige completar el reto telefónico (`DEC-062`) antes de evaluar la contraseña. Se corrige `login_throttle_register_attempt` en `database/modelo-fisico-referencia.sql` para escalar cuando `attempt_count + 1 > p_threshold` (antes `>= p_threshold`), y `p_threshold` se documenta como "cantidad de solicitudes permitidas sin reto".
- **Responsable:** propietario del proyecto.
- **Motivo:** `DP-SEG-02`, la duda original detrás de `DEC-026`, usa literalmente "al **superarlo**" (superar 5), no "al alcanzarlo"; es la fuente más cercana a la intención original, y de las tres opciones de `CT-005` es la que exige el menor cambio de redacción, porque `HU-007`/`CA-007-01`/`CA-007-02` ya usan la palabra "superar".
- **Alternativas descartadas:** la quinta ya exige el reto (opción 2) — descartada por contradecir "superar", ya usado en `DEC-026` y `HU-007`; redefinir `p_threshold` como "cantidad permitida" (opción 3) — es funcionalmente idéntica a la opción 1 una vez fijado el operador `>`, así que se adopta como aclaración de nomenclatura dentro de la opción 1, no como alternativa separada.
- **Documentos afectados:** `database/modelo-fisico-referencia.sql` (función y comentario de la sección A.4), `docs/00-control/contradicciones.md` (cierra `CT-005`), `docs/02-requisitos/historias-usuario.md` (`CA-007-01`/`CA-007-02`), `docs/10-backlog/prompts/hu/hu-007-defensa-abuso.md`.
- **Fuente:** `docs/00-control/contradicciones.md`, `CT-005`; aprobación explícita del propietario el 2026-08-17.

### DEC-062 · Resolución de `DP-SEG-10`: reto telefónico completo de `HU-007`

- **Fecha:** 2026-08-17.
- **Decisión:** al escalar (`DEC-061`), el login responde `429` uniforme sin evaluar la contraseña, indicando que se requiere verificación telefónica. El barbero completa el reto con dos operaciones públicas nuevas:
  - `POST /api/v1/public/auth/challenge { email }` → siempre `202` con mensaje genérico ("si la cuenta existe y la IP está en verificación, se envió un código por WhatsApp"), sin importar si el correo existe, si el teléfono está verificado o si la IP está realmente escalada; solo se envía un mensaje real por WhatsApp oficial cuando las tres condiciones se cumplen. Límite propio: 1 solicitud cada 60 segundos y máximo 3 solicitudes activas por IP dentro de la misma ventana de 15 minutos de `DEC-052`, para que el propio reto no se vuelva un vector de bombardeo del teléfono de un tercero.
  - `POST /api/v1/public/auth/challenge/verify { email, code }` → verifica un código numérico de 6 dígitos, vigente 5 minutos, máximo 5 intentos, un solo uso; responde exactamente igual (mismo status/schema) ante código incorrecto, vencido, agotado o cuenta inexistente. Éxito: limpia `escalated_until` y `attempt_count` de esa IP en `login_throttle`, dentro de la misma función `SECURITY DEFINER` que ya la escribe, y responde `204`. El barbero reintenta el login normalmente después; no hay token adicional que enlazar, porque la IP ya dejó de estar escalada.
  - Almacenamiento: tabla nueva `auth_phone_challenge` (`staff_user_id` FK, `ip_hash` igual al HMAC de `login_throttle` para atar el reto a la IP concreta que lo pidió, `code_hash = HMAC-SHA256(código, secreto de despliegue)` — mismo patrón que el HMAC de IP, nunca `SHA-256` simple de un espacio de 10⁶ —, `created_at`, `expires_at`, `attempts_count`, `consumed_at`, `invalidated_at`). RLS tenant-aware vía `staff_user_id → barbershop_id`. `login_throttle` en sí no gana columnas nuevas: sigue sin correo, usuario o barbería (`DEC-052`).
- **Responsable:** propietario del proyecto.
- **Motivo:** cumple `DEC-026`/`DEC-052` (verificación telefónica que desbloquea el acceso) sin crear un oráculo de cuentas (respuesta uniforme en ambas operaciones) ni un vector de abuso nuevo (límite de solicitud propio, código atado a la IP escalada específica). Reutiliza los mismos patrones criptográficos ya aprobados del propio `HU-007` (HMAC con secreto de despliegue) en vez de inventar uno nuevo.
- **Alternativas descartadas:** reutilizar el código de recuperación de `HU-008` — descartada explícitamente por el propio prompt de `HU-007` y porque `HU-008` depende de `HU-007` integrada, lo que crearía una dependencia circular; enviar el código sin límite de reenvío propio — descartado por abrir un vector de bombardeo del teléfono de un tercero solo con conocer su correo; exigir un token de un solo uso entre `verify` y el siguiente login — descartado por innecesario: limpiar `escalated_until` ya es suficiente y evita inventar un artefacto de autorización adicional sin una razón de seguridad clara.
- **Documentos afectados:** `database/modelo-fisico-referencia.sql` (nueva sección de diseño, tabla `auth_phone_challenge` — la implementación real crea la migración Atlas; esta sección es insumo, no migración aplicada), `docs/00-control/dudas-pendientes.md` (cierra `DP-SEG-10`), `docs/02-requisitos/historias-usuario.md` (`CA-007-02`/`CA-007-03`/`CA-007-04`), `docs/10-backlog/prompts/hu/hu-007-defensa-abuso.md`.
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-SEG-10`; aprobación explícita del propietario el 2026-08-17.

### DEC-063 · Resolución de `DP-SEG-11`: política mínima de la contraseña nueva

- **Fecha:** 2026-08-17.
- **Decisión:** la contraseña nueva de `CA-008-08` debe cumplir: longitud mínima 10 caracteres, longitud máxima 128 (límite defensivo de tamaño de entrada para Argon2id, no una regla de seguridad); sin exigencia de composición (no se fuerza mayúscula, número ni símbolo); se rechaza si es idéntica al correo de la cuenta o a la contraseña actual. No se valida contra una lista externa de contraseñas filtradas en este MVP — es una decisión de proveedor aparte, fuera de `DP-SEG-11`, y queda como mejora futura, no como bloqueo de `HU-008`. El mensaje de rechazo indica exactamente cuál regla incumple, nunca un mensaje genérico.
- **Responsable:** propietario del proyecto.
- **Motivo:** valores concretos y accionables que no dependen de un proveedor externo nuevo ni de heurísticas subjetivas; 10 caracteres es más estricto que el mínimo típico de 8 para compensar la ausencia de reglas de composición.
- **Alternativas descartadas:** exigir composición (mayúscula+número+símbolo) — descartada por evidencia documentada de que ese tipo de regla empeora la calidad real de las contraseñas elegidas; verificar contra una lista de contraseñas filtradas (estilo HaveIBeenPwned) — descartada por introducir una dependencia/proveedor externo nuevo sin evaluar costo ni disponibilidad, ampliando el alcance de `HU-008` igual que `DP-NOT-05`.
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-SEG-11`), `docs/02-requisitos/historias-usuario.md` (`CA-008-08`), `docs/10-backlog/prompts/hu/hu-008-recuperacion-acceso.md`.
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-SEG-11`; aprobación explícita del propietario el 2026-08-17.

### DEC-064 · Resolución de `DP-SEG-12`: parámetros y autorización del código de recuperación

- **Fecha:** 2026-08-17.
- **Decisión:**
  - Formato: 6 dígitos numéricos generados con el mismo generador criptográfico que ya usa `CryptoTokenGenerator` para tokens de sesión (`crypto/rand`, distribución uniforme `000000`–`999999`).
  - Representación almacenada: `HMAC-SHA256(código, secreto de despliegue)` — mismo patrón que el HMAC de IP de `HU-007` (`DEC-062`); nunca `SHA-256` simple del código, porque un espacio de 10⁶ se recupera trivialmente offline sin un secreto.
  - Vigencia: 15 minutos.
  - Intentos máximos: 5; al agotarse, el código se invalida por completo (`CA-008-03`) y exige solicitar uno nuevo.
  - Reenvío: cooldown de 60 segundos entre solicitudes y máximo 3 solicitudes activas por cuenta por hora; cada reenvío invalida atómicamente el código anterior en la misma escritura que crea el nuevo (`CA-008-07`), nunca deja dos códigos vigentes.
  - Identificador de la solicitud: ninguno nuevo — solicitar y verificar se hacen siempre contra `email`, igual que el login, para no crear un identificador correlacionable entre pasos.
  - Autorización entre verificar y cambiar contraseña: al verificar con éxito, el servidor emite un token opaco de reinicio (mismo generador y patrón hash que el token de sesión de `HU-005`/`HU-006`: `CryptoTokenGenerator` + `HashToken`), vigente 5 minutos, de un solo uso, devuelto una única vez en el cuerpo de la respuesta de verificación exitosa. Cambiar contraseña exige `email` + ese token; se consume atómicamente en la misma transacción que actualiza la credencial y revoca las sesiones (`CA-008-05`); reutilizarlo o presentarlo vencido responde el mismo error uniforme que un token inválido.
- **Responsable:** propietario del proyecto.
- **Motivo:** reutiliza los mismos patrones criptográficos y de token opaco que `HU-005`/`HU-006`/`HU-007` (`DEC-062`) ya implementaron y probaron, en vez de inventar un mecanismo nuevo; los valores de vigencia/intentos/reenvío son coherentes con los ya aprobados para el reto de `HU-007`, ajustados a un flujo voluntario (15 min, no una emergencia de sesión bloqueada).
- **Alternativas descartadas:** JWT firmado como token de reinicio — descartado por añadir una dependencia de firma/verificación nueva cuando el patrón opaco+hash ya existe, probado y auditado; código alfanumérico más largo — descartado por complicar copiarlo desde WhatsApp/correo sin ganancia de seguridad relevante una vez que el HMAC con secreto de despliegue ya impide fuerza bruta offline.
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-SEG-12`), `docs/02-requisitos/historias-usuario.md` (`CA-008-02`/`CA-008-03`/`CA-008-04`/`CA-008-07`), `docs/10-backlog/prompts/hu/hu-008-recuperacion-acceso.md`.
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-SEG-12`; aprobación explícita del propietario el 2026-08-17.

### DEC-065 · Resolución de `CT-006`: respuesta idéntica en la solicitud, destino enmascarado solo tras verificar

- **Fecha:** 2026-08-17.
- **Decisión:** opción 1 de `CT-006`. `POST /recovery/request` (paso 1) responde siempre igual — mismo status, schema y cuerpo genérico ("si la cuenta existe, se envió un código a los medios registrados") — sin importar si el correo existe o si el teléfono está verificado; nunca incluye el destino, ni siquiera enmascarado. El destino enmascarado (teléfono tipo `+57 *** *** 12` y/o correo tipo `b***@c***.com`) solo aparece en el cuerpo de la respuesta exitosa de `POST /recovery/verify` (paso 2, `DEC-064`): llegar a ese punto ya exige haber recibido y transcrito el código real, así que mostrar el destino ahí no abre un oráculo nuevo — quien no tiene el código nunca llega a verlo.
- **Responsable:** propietario del proyecto.
- **Motivo:** de las opciones registradas en `CT-006`, es la que preserva la no enumeración estricta de `CA-008-01` en el paso público sin prueba de posesión, y solo revela el destino después de una prueba de posesión real (el código correcto), que es exactamente lo que `CA-008-06` necesita para no mostrar el teléfono completo en ningún momento.
- **Alternativas descartadas:** marcador fijo no derivado del dato real (opción 2) — descartada por no darle al barbero información útil real y no resolver `CA-008-06`; cuerpo distinto aceptando el riesgo de enumeración (opción 3) — descartada por contradecir literalmente `CA-008-01`, que exige respuesta idéntica.
- **Documentos afectados:** `docs/00-control/contradicciones.md` (cierra `CT-006`), `docs/02-requisitos/historias-usuario.md` (`CA-008-01`/`CA-008-06`), `docs/10-backlog/prompts/hu/hu-008-recuperacion-acceso.md`.
- **Fuente:** `docs/00-control/contradicciones.md`, `CT-006`; aprobación explícita del propietario el 2026-08-17.

### DEC-066 · Resolución de `DP-NOT-05`: proveedor de entrega para `HU-008`

- **Fecha:** 2026-08-17.
- **Decisión:** WhatsApp: **Meta WhatsApp Cloud API en modo directo** (sin BSP intermediario), usando una plantilla de categoría "Authentication" pre-aprobada por Meta para el código de recuperación (es el único tipo de plantilla que Meta permite enviar fuera de la ventana de conversación de 24 horas sin recargo de "marketing", y su formato está pensado exactamente para OTP). Correo: proveedor transaccional **Resend** (API REST simple, dominio verificado, capa gratuita adecuada al volumen del piloto de 2–3 barberías); si el propietario ya opera con AWS, **SES** es alternativa aprobada equivalente sin requerir una nueva decisión. Ambos adaptadores viven detrás de un puerto único en el módulo `notification` (forma exacta a definir por la implementación siguiendo `estandar-backend-go.md`, p. ej. `SendRecoveryCode(ctx, phone, email, code string) error`), con timeout corto (5 s) por canal, sin reintento síncrono dentro de la solicitud HTTP (para no mantener abierta la transacción de PostgreSQL durante red, mismo criterio ya aplicado en `HU-007`), y tolerancia a fallo parcial: si un canal falla y el otro no, la respuesta al cliente sigue siendo el mismo mensaje genérico de `DEC-065`, porque el fallo del proveedor no debe filtrarse como señal de existencia de cuenta; el fallo se registra internamente sin destinatario ni contenido, identificable solo por `requestId`. Las credenciales (token de Meta, `phone_number_id`, API key de Resend) se leen de variables de entorno de despliegue, nunca se comitean, y la configuración se valida al arrancar.
- **Responsable:** propietario del proyecto (elección delegada explícitamente: "elige un default razonable para el piloto").
- **Motivo:** Meta Cloud API directo evita un BSP de pago adicional mientras el volumen es de piloto, y su plantilla de autenticación es el mecanismo oficial soportado por WhatsApp para OTP, sin depender de aprobación manual de contenido libre en cada envío. Resend tiene una API mínima y nivel gratuito suficiente para un piloto, sin atar el despliegue a AWS si el propietario no lo usa ya.
- **Alternativas descartadas:** BSP intermedio (Twilio/360dialog/Infobip) — descartado por el costo mensual y la capa de abstracción adicional innecesaria para el volumen del piloto; revisable si el producto escala. SendGrid/Postmark para correo — descartados solo por preferencia de simplicidad frente a Resend; son alternativas igualmente válidas si el propietario ya tiene cuenta con alguno.
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-NOT-05`), `docs/00-control/matriz-trazabilidad.md`, `docs/02-requisitos/historias-usuario.md` (`CA-008-01`), `docs/10-backlog/prompts/hu/hu-008-recuperacion-acceso.md`, `apps/api/README.md` (documentar variables de entorno esperadas, sin valores).
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-NOT-05`; aprobación explícita del propietario el 2026-08-17.

### DEC-067 · Resolución de `DP-SER-01`: moneda, gratuidad y nombres del catálogo

- **Fecha:** 2026-08-24.
- **Decisión:** el catálogo de `HU-022` usa moneda **COP fija** para todo el MVP (sin campo de moneda configurable por barbería ni por servicio); el precio debe ser estrictamente mayor que cero (`price > 0`), por lo que **no se permiten servicios gratuitos**; y dos servicios **activos** de la misma barbería no pueden compartir nombre exacto (unicidad parcial tenant-aware sobre `is_active = true`), mientras que un servicio inactivo puede conservar un nombre igual al de uno activo nuevo (no bloquea reactivación futura ni renombrar). Confirma la propuesta ya presente en `database/modelo-fisico-referencia.sql` §B.3 como decisión de producto, no solo como insumo técnico.
- **Responsable:** propietario del proyecto.
- **Motivo:** COP fija evita construir selección/almacenamiento de moneda y su validación ISO sin un caso de uso real todavía (el piloto opera en una sola región); exigir precio positivo evita que "gratis" se confunda con un campo vacío o un error de captura, y mantiene el precio como una señal comercial real; la unicidad por nombre entre servicios activos evita catálogos ambiguos en la pantalla y en el futuro enlace público, sin impedir que un nombre se reutilice tras desactivar el original.
- **Alternativas descartadas:** moneda configurable por barbería — descartada por ampliar el alcance de `HU-022` con un campo, validación y ejemplos que ningún criterio exige todavía; permitir precio 0 — descartada por no tener un caso de uso aprobado (cortesías/promociones quedan fuera del MVP) y por complicar innecesariamente la regla de validación; permitir nombres duplicados entre servicios activos — descartada por degradar la lista y la futura selección pública sin beneficio claro.
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-SER-01`), `docs/02-requisitos/historias-usuario.md` (`HU-022`, `CA-022-02`–`CA-022-04`), `docs/10-backlog/prompts/hu/hu-022-catalogo-servicios.md`; futura implementación (issue `#75`) debe reflejar `currency = 'COP'` fijo (no columna editable), `CHECK (price > 0)` y un índice único parcial `(barbershop_id, name) WHERE is_active`.
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-SER-01`; aprobación explícita del propietario el 2026-08-24.

### DEC-068 · Resolución de `DP-SER-02`: última asignación de un servicio activo

- **Fecha:** 2026-08-24.
- **Decisión:** un servicio activo debe conservar al menos un barbero asignado. La operación de desasignación que dejaría a un servicio activo sin ningún barbero se **rechaza** (bloqueo duro, `409`/`422` según el contrato), no se advierte ni se permite silenciosamente. Esta regla aplica únicamente mientras el servicio esté activo: desactivar el servicio (`HU-024`) sí puede dejarlo sin asignaciones, porque un servicio inactivo no se ofrece de todas formas.
- **Responsable:** propietario del proyecto.
- **Motivo:** de las tres opciones registradas en `DP-SER-02`, el bloqueo duro es la única que garantiza que "servicio activo" siga significando "servicio reservable" en todo momento, sin depender de que el barbero recuerde reasignarlo después ni de una advertencia que puede ignorarse; evita además tener que definir y probar un estado intermedio ("activo mostrable" vs. "activo sin oferta real") que ningún criterio de `HU-022`/`HU-023` exige.
- **Alternativas descartadas:** solo advertencia — descartada por dejar un estado inconsistente (servicio activo sin oferta real) como resultado normal del flujo, sin ganancia frente al bloqueo; sin restricción — descartada por la misma razón, agravada porque tampoco exige confirmación explícita.
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-SER-02`), `docs/02-requisitos/historias-usuario.md` (`HU-023`, `CA-023-05`, `CA-023-06`), `docs/10-backlog/prompts/hu/hu-023-asignacion-servicios-barberos.md`; futura implementación (issue `#76`) debe rechazar la desasignación cuando sea la última fila activa de `barber_service` para ese `service_id`, verificado dentro de la misma transacción (carrera de dos desasignaciones concurrentes de las dos últimas filas).
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-SER-02`; aprobación explícita del propietario el 2026-08-24.

### DEC-069 · Resolución de `DP-SER-03`: alcance B1/B3 de la advertencia de citas futuras

- **Fecha:** 2026-08-24.
- **Decisión:** `HU-024` (B1) construye un recuento simple de impacto antes de confirmar la desactivación: el sistema calcula y muestra cuántas citas futuras quedarían afectadas usando el estado real disponible en ese momento (que en B1 siempre será 0, porque `appointment` todavía no existe en la cadena migrada), sin protección contra cambios concurrentes entre previsualizar y confirmar (sin token de versión ni bloqueo optimista). El contrato, el dominio y la interfaz quedan preparados para que B3 rellene ese conteo con datos reales de `appointment` sin cambiar la forma de la operación. La cancelación selectiva de citas y cualquier protección de concurrencia sobre el conteo quedan explícitamente fuera de B1 y se construyen junto con `appointment` en B3.
- **Responsable:** propietario del proyecto.
- **Motivo:** de las opciones registradas en `DP-SER-03`, es la que no inventa un conteo cero como caso especial documentado en el código de negocio (sería una mentira de dominio: "cero" es simplemente el resultado real de una consulta sobre una tabla vacía, no un atajo), y no exige construir un mecanismo de concurrencia para proteger un dato que en B1 nunca cambia entre previsualizar y confirmar (no puede existir una carrera real sin `appointment`); tampoco difiere toda la funcionalidad a B3, lo que dejaría `HU-024` sin ninguna advertencia y rompería `RN-SER-03` antes de tiempo.
- **Alternativas descartadas:** recuento con bloqueo optimista en B1 — descartada por construir protección de concurrencia contra una carrera que no puede ocurrir todavía (no hay `appointment` que cambie el conteo entre las dos llamadas), costo sin beneficio real hasta B3; diferir toda advertencia a B3 — descartada por dejar `HU-024` sin cumplir la parte de `RN-SER-03` que sí es alcanzable ahora (mostrar impacto antes de confirmar), aunque ese impacto sea siempre cero en B1.
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-SER-03`), `docs/01-producto/reglas-negocio.md` (`RN-SER-03`, nota de alcance B1/B3), `docs/02-requisitos/historias-usuario.md` (`HU-024`, `CA-024-01`, `CA-024-04`), `docs/10-backlog/prompts/hu/hu-024-ciclo-vida-servicios.md`; futura implementación (issue `#77`) expone la previsualización como una operación real que consulta el estado vigente sin simular ni cachear un valor, y sin agregar un mecanismo de versión/lease que ningún dato concurrente todavía justifica.
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-SER-03`; aprobación explícita del propietario el 2026-08-24.

### DEC-070 · Resolución de `CT-008`: `RESTRICT` en vez de `ON DELETE CASCADE` en el modelo físico de B2

- **Fecha:** 2026-08-25.
- **Decisión:** opción 1 de `CT-008`. Las siete FK de `database/modelo-fisico-referencia.sql` §C que apuntaban a `barber` o a su propia cabecera con `ON DELETE CASCADE` (`working_hour`, `working_hour_override`, `working_hour_override_segment`, `time_block_series`, `time_block_series_date`, `time_block_series_exception`, `time_block`) pasan a `ON DELETE RESTRICT`. Ninguna migración de B2 crea borrado en cascada; si en el futuro se aprueba borrar o limpiar un `barber`, esa historia debe definir explícitamente qué pasa con su horario y sus bloqueos (operación propia, no un efecto colateral de la FK) antes de tocar estas restricciones.
- **Responsable:** propietario del proyecto.
- **Motivo:** `barber` no tiene borrado físico en su alcance vigente (`HU-021`, DDL-BIZ-02: "sin borrado, desactivación... una historia futura los aprueba explícitamente si llegan a hacer falta") y su propia FK hacia `barbershop` ya usa `RESTRICT`; usar `RESTRICT` en las FK de B2 es la opción consistente con ese precedente y con la prohibición general de borrado en cascada de `AGENTS.md`, sin inventar una excepción nueva que documentar y mantener. Además, varias de estas tablas (`time_block_series`, `time_block`) ya tienen su propia eliminación lógica (`deleted_at`), así que una cascada física sería además redundante con el mecanismo de retención que ya existe.
- **Alternativas descartadas:** excepción documentada de `AGENTS.md` para tablas de configuración (opción 2) — descartada por no tener ningún caso de uso que la necesite hoy (nada borra físicamente un `barber` todavía) y por ampliar una prohibición general con una excepción que solo serviría a un borrado que ni siquiera existe; otro mecanismo de retención a medida (opción 3) — descartada por ser innecesaria cuando `RESTRICT` ya basta mientras `barber` no tenga borrado físico.
- **Documentos afectados:** `docs/00-control/contradicciones.md` (cierra `CT-008`), `database/modelo-fisico-referencia.sql` §C (siete FK cambiadas de `CASCADE` a `RESTRICT`, con comentario en cada una), `docs/10-backlog/prompts/hu/hu-040-horario-laboral.md`, `docs/10-backlog/prompts/hu/hu-041-excepciones-festivos.md`, `docs/10-backlog/prompts/hu/hu-042-bloqueos-agenda.md`, `docs/10-backlog/plan-bloques.md`.
- **Fuente:** `docs/00-control/contradicciones.md`, `CT-008`; aprobación explícita del propietario el 2026-08-25.

### DEC-071 · Resolución de `DP-CIT-01`: identidad/reutilización de `customer` sin teléfono en una cita manual

- **Fecha:** 2026-08-27.
- **Decisión:** cuando la cita manual omite el teléfono (`RN-CIT-02` lo permite) pero el cliente sí da correo, `customer` se busca/reutiliza por correo dentro de la misma barbería, reutilizando la unicidad de correo por tenant que ya fija `DEC-046`. Cuando faltan ambos —teléfono y correo—, la creación siempre inserta una fila `customer` nueva; no se reconcilia por nombre ni por ningún otro dato, porque el nombre no es un identificador confiable (dos clientes distintos pueden compartir nombre) y forzar una coincidencia por texto arriesgaría fusionar personas distintas o dejar historial cruzado entre ellas.
- **Responsable:** propietario del proyecto.
- **Motivo:** mantiene un único criterio de identidad por dato disponible, en el mismo orden de confiabilidad que ya usa el sistema: teléfono (`DEC-045`) primero, correo (`DEC-046`) como respaldo cuando falta teléfono, y fila nueva solo cuando no hay ningún dato de contacto verificable con el que reconciliar; no introduce un mecanismo de fusión manual ni heurísticas de coincidencia de texto que ningún criterio exige todavía.
- **Alternativas descartadas:** siempre fila nueva sin teléfono, incluso con correo presente — descartada por desperdiciar el correo como identificador ya disponible y fragmentar el historial de un mismo cliente en varias filas sin necesidad; reconciliación por nombre — descartada por el riesgo real de fusionar personas distintas con nombres coincidentes o comunes.
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-CIT-01`), `docs/01-producto/reglas-negocio.md` (`RN-CIT-02`, caso límite nuevo), `docs/02-requisitos/historias-usuario.md` (`HU-061`, `CA-061-04`), `docs/10-backlog/prompts/hu/hu-061-creacion-manual-citas.md`; futura implementación debe aplicar upsert por correo dentro del tenant solo cuando falte el teléfono, coherente con la unicidad ya vigente de `DEC-046`.
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-CIT-01`; aprobación explícita del propietario el 2026-08-27.

### DEC-072 · Resolución de `DP-CIT-02`: alcance de la asignación servicio-barbero en una cita manual

- **Fecha:** 2026-08-27.
- **Decisión:** una cita manual solo puede crearse con un servicio que esté activo **y** asignado al barbero elegido (fila vigente en `barber_service` de `HU-023`); un servicio activo de la barbería pero no asignado a ese barbero se rechaza, aunque exista y esté activo.
- **Responsable:** propietario del proyecto.
- **Motivo:** mantiene coherencia con la única fuente de verdad que ya existe para "qué presta cada barbero" (`barber_service`, `HU-023`) y con `RN-DIS-02`, que la disponibilidad pública usará exactamente igual; permitir en la creación manual un servicio que el barbero no tiene asignado produciría una cita que la agenda y la futura reserva pública no podrían explicar de forma consistente, y obligaría a mantener dos reglas distintas de "servicio válido para un barbero" en el mismo sistema.
- **Alternativas descartadas:** cualquier servicio activo de la barbería sin exigir asignación — descartada por crear una segunda noción de "servicio válido" solo para la creación manual, divergente de `HU-023` y de la disponibilidad pública, sin ningún caso de uso documentado que la requiera.
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-CIT-02`), `docs/02-requisitos/historias-usuario.md` (`HU-061`, `CA-061-04`), `docs/10-backlog/prompts/hu/hu-061-creacion-manual-citas.md`; futura implementación valida la asignación vigente `barber_service` antes de aceptar la cita, con el mismo error `422`/`409` uniforme que otras validaciones de negocio de `HU-061`.
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-CIT-02`; aprobación explícita del propietario el 2026-08-27.

### DEC-073 · Resolución de `DP-CIT-03`: conducta de la creación manual frente a un bloqueo vigente

- **Fecha:** 2026-08-27.
- **Decisión:** si el intervalo elegido cae dentro de la jornada laboral pero coincide con un bloqueo vigente de cualquier tipo (`break`, `lunch`, `unavailable`, `day_off`, `holiday`, `vacation`, `emergency`), la creación manual se **rechaza** con el mismo tratamiento de conflicto que un cruce entre dos citas (bloqueo duro, sin advertencia que permita continuar). El barbero debe retirar o exceptuar el bloqueo desde las pantallas reales de `HU-041`/`HU-042` antes de poder crear la cita en ese intervalo.
- **Responsable:** propietario del proyecto.
- **Motivo:** de las tres opciones registradas en `DP-CIT-03`, el bloqueo duro es la única coherente con el precedente ya sentado por `DEC-068` (rechazar en vez de advertir cuando una operación dejaría el sistema en un estado que otra regla de negocio prohíbe) y con `RN-CIT-02` ("la cita manual... es indistinguible de las públicas en cuanto a su efecto sobre la agenda"): un bloqueo vigente significa que ese tiempo no está disponible para ningún origen de cita, manual o pública, y permitir una excepción silenciosa solo para el flujo manual reintroduciría la ambigüedad que la exclusión de `HU-060` existe para eliminar.
- **Alternativas descartadas:** permitir con advertencia — descartada porque delega en el barbero, en el momento de más prisa (creando una cita mientras atiende a alguien), una decisión que ya tiene un flujo propio y auditable en `HU-041`/`HU-042` (editar o exceptuar el bloqueo); exigir retirar/exceptuar el bloqueo como paso obligatorio dentro del mismo formulario de cita — descartada por mezclar dos preocupaciones distintas (gestión de bloqueos y creación de citas) en una sola pantalla y una sola transacción, cuando `HU-041`/`HU-042` ya resuelven la gestión de bloqueos de forma completa y auditada.
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-CIT-03`), `docs/01-producto/reglas-negocio.md` (`RN-CIT-02`, caso límite nuevo), `docs/02-requisitos/historias-usuario.md` (`HU-061`, `CA-061-04`), `docs/10-backlog/prompts/hu/hu-061-creacion-manual-citas.md`; futura implementación consulta la jornada efectiva/bloqueos vigentes de `schedule` antes de confirmar y responde el mismo `409` uniforme que usa para un cruce de citas.
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-CIT-03`; aprobación explícita del propietario el 2026-08-27.

### DEC-074 · Resolución de `DP-CIT-04`: vista inicial de la agenda diaria en una barbería con varios barberos

- **Fecha:** 2026-08-28.
- **Decisión:** la agenda diaria de `HU-062` abre siempre con un **selector obligatorio de un barbero**; no existe una vista consolidada de todos los barberos a la vez ni un default implícito que muestre citas sin que el staff elija explícitamente de quién. En una barbería con un solo barbero el selector puede preseleccionar la única opción existente sin exigir un paso adicional (mismo patrón que `DEC-019`/`DP-UX-01` ya aplica a la selección del cliente), pero la autoridad de la vista sigue siendo "un barbero a la vez", nunca una consolidación.
- **Responsable:** propietario del proyecto.
- **Motivo:** `DEC-047` prohíbe inventar un vínculo automático `staff_user`→`barber`; sin esa autoridad, una vista consolidada por defecto obligaría a decidir de forma no verificable cómo entrelazar turnos de varios barberos en una sola lista (orden, color, agrupación) sin ninguna fuente que lo exija todavía. El selector obligatorio no inventa ninguna asociación: exige que la persona que opera la pantalla declare explícitamente de qué barbero quiere ver la agenda, igual que ya hace el cliente en `DP-UX-01` al reservar.
- **Alternativas descartadas:** vista consolidada por defecto — descartada por requerir una regla de fusión/orden entre barberos que ninguna decisión previa fija, y por arriesgar que un barbero vea u opere sobre turnos de otro sin una decisión explícita de alcance; "otro criterio" (recordar el último barbero visto, orden alfabético) — descartado por añadir estado de preferencia por usuario que esta historia no necesita y que puede decidirse después sin costo, sin bloquear la entrega mínima.
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-CIT-04`), `docs/02-requisitos/historias-usuario.md` (`HU-062`, criterios de aceptación), `docs/10-backlog/prompts/hu/hu-062-agenda-diaria.md`; futura implementación exige un `barberId` explícito en la operación de agenda diaria y en la pantalla, sin vista "todos".
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-CIT-04`; aprobación explícita del propietario el 2026-08-28.

### DEC-075 · Resolución de `DP-CIT-05`: pertenencia diaria de un turno que cruza medianoche

- **Fecha:** 2026-08-28.
- **Decisión:** un turno cuyo intervalo `[starts_at, ends_at)` cruza la medianoche aparece **en cada agenda diaria cuyo rango civil interseca ese intervalo**: en la agenda del día en que empieza y también en la agenda del día siguiente, donde termina. No se elige un único "día dueño" del turno.
- **Responsable:** propietario del proyecto.
- **Motivo:** `DEC-020` ya permite que un tramo de horario/bloqueo cruce medianoche cuando cabe completo en la jornada configurada, y `HU-060` persiste el mismo tipo de intervalo semiabierto para `appointment`. Elegir "solo el día de inicio" ocultaría el turno de la agenda del día en que realmente se atiende al cliente durante la madrugada, que es precisamente cuando el barbero más necesita verlo listado; la intersección es la única regla que mantiene la agenda de cada día como un reflejo fiel de "qué pasa en mi negocio hoy", consistente con el objetivo de `HU-062`.
- **Alternativas descartadas:** solo el día civil de inicio — descartada porque un barbero que abre la agenda del día siguiente (donde transcurre la segunda mitad del turno) no vería el turno en curso, contradiciendo el objetivo de la historia de mostrar "los turnos de hoy".
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-CIT-05`), `docs/02-requisitos/historias-usuario.md` (`HU-062`, criterios de aceptación), `docs/10-backlog/prompts/hu/hu-062-agenda-diaria.md`; futura implementación filtra por intersección de rango (`starts_at < fin_civil AND ends_at > inicio_civil`), no por igualdad de fecha de `starts_at`, y un turno nocturno puede aparecer una vez en cada una de las dos agendas sin duplicarse dentro de la misma.
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-CIT-05`; aprobación explícita del propietario el 2026-08-28.

### DEC-076 · Resolución de `DP-CIT-06`: reprogramación (T2) hacia un intervalo con bloqueo vigente

- **Fecha:** 2026-09-01.
- **Decisión:** si el nuevo intervalo de una reprogramación (`T2`, `HU-065`) coincide con un bloqueo vigente del mismo barbero (cualquier tipo: `break`, `lunch`, `unavailable`, `day_off`, `holiday`, `vacation`, `emergency`), la reprogramación se **rechaza** con el mismo tratamiento de conflicto que un cruce entre dos citas: mismo `409` uniforme, sin advertencia que permita continuar. El barbero debe retirar o exceptuar el bloqueo desde las pantallas reales de `HU-041`/`HU-042` antes de poder reprogramar hacia ese intervalo.
- **Responsable:** propietario del proyecto.
- **Motivo:** mismo criterio que `DEC-073` (creación manual frente a un bloqueo vigente): un bloqueo vigente significa que ese tiempo no está disponible para ningún efecto sobre la agenda del barbero, sea una cita nueva o una reprogramada. `RN-BLQ-03` autoriza que un bloqueo se **cree** sobre una cita ya existente sin moverla (el barbero decide después, con aviso), pero eso no implica lo inverso: permitir que el barbero **elija** reprogramar voluntariamente hacia un bloqueo ya vigente reintroduciría, por el lado de T2, la misma ambigüedad que `DEC-073` cerró por el lado de la creación manual. Tratar ambos casos de forma distinta obligaría a la interfaz a explicar por qué "crear aquí" se rechaza pero "mover aquí" no, sin ninguna razón de negocio que lo justifique.
- **Alternativas descartadas:** permitir con advertencia — descartada por la misma razón que en `DEC-073`: delega en el barbero, en medio de un flujo de reprogramación, una decisión que ya tiene su propio flujo auditable en `HU-041`/`HU-042`; exigir retirar/exceptuar el bloqueo dentro del mismo formulario de reprogramación — descartada por mezclar dos preocupaciones distintas (gestión de bloqueos y reprogramación de una cita) en una sola pantalla y una sola transacción.
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-CIT-06`), `docs/01-producto/reglas-negocio.md` (`RN-BLQ-03`, caso límite nuevo), `docs/02-requisitos/historias-usuario.md` (`HU-065`, criterios de aceptación), `docs/10-backlog/prompts/hu/hu-065-reprogramacion-turno.md`; futura implementación de `T2` consulta la jornada efectiva/bloqueos vigentes de `schedule` (mismo puerto que `ManualBookingService`) antes de confirmar y responde el mismo `409` uniforme que usa para un cruce de citas, nunca una rama de "advertencia".
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-CIT-06`; aprobación explícita del propietario el 2026-09-01.

### DEC-077 · Identidad NAVA y dirección visual Tailored Grid

- **Fecha:** 2026-09-01.
- **Decisión:** la plataforma y el proyecto adoptan el nombre **NAVA**. El frontend adopta como dirección canónica **NAVA / Tailored Grid**: elegancia editorial sobria inspirada en sastrería contemporánea y hospitalidad boutique, con tinta azul marino, marfil cálido, grafito, latón discreto, salvia y piedra; wordmark `NAVA`; contraste entre `Instrument Serif` para marca/display e `Instrument Sans` para interfaz; líneas finas, radios pequeños, sombras mínimas y navegación inferior estable. La implementación conserva un tema claro único, tokens semánticos, diseño móvil primero, objetivo táctil de 44 × 44 px y WCAG 2.2 AA.
- **Alcance funcional:** la elección es visual y nominal; no convierte en requisitos las funciones dibujadas en un mockup conceptual. Ingresos, caja, reportes, inventario, cuentas de cliente, modo oscuro y personalización por tenant continúan fuera del MVP. “Ahora” es una señal temporal, no el estado futuro `in_progress`. La agenda P0 continúa mostrando un barbero seleccionado a la vez según `DEC-074`; la cuadrícula consolidada multi-barbero queda diferida hasta que una decisión e historia la autoricen. La interfaz conserva “turno” y el código/API conserva `appointment`, según `DEC-016`.
- **Convivencia con `DEC-039`:** esta decisión sustituye de `DEC-039` la paleta concreta azul/cobre/neutros fríos, la tipografía exclusivamente del sistema y el shell con navegación lateral. Conserva su gobierno por tokens, escala espacial de 4 px, componentes ligeros, tema claro único, accesibilidad, rendimiento y prohibición de estilos libres por módulo o barbería.
- **Tipografía:** los assets de Instrument se incorporan únicamente mediante un issue real que documente licencia, WOFF2 self-hosted, pesos, fallback, `font-display` y medición de rendimiento. Hasta entonces se usan los fallbacks definidos por el estándar; no se carga una fuente remota desde un componente.
- **Identidad del tenant:** NAVA identifica la plataforma. El nombre de la barbería sigue siendo dato y contexto principal del tenant; en el flujo público se muestra primero la barbería y como firma secundaria “Reservas con NAVA”. Un logotipo de barbería no se inventa ni se vuelve obligatorio sin una capacidad de carga y almacenamiento aprobada.
- **Implementación incremental:** la documentación fija el destino, pero no autoriza una migración masiva ni parcial del código. Cada fundación o pantalla se adopta con issue, rama, pruebas, evidencia responsive y preservación del contrato existente. La especificación distingue `P0 existente`, `P0 pendiente`, `P1`, `P2` y `Excluido` para impedir que un agente implemente por inferencia visual.
- **Alternativas descartadas:** mantener “Block Party” como identidad final — descartada por ser más estridente que la orientación elegante solicitada; copiar literalmente el mockup — descartada porque contradice alcance y decisiones vigentes; rediseñar cada pantalla sin sistema — descartado por producir divergencia; migrar todo el frontend en una rama — descartado por mezclar preocupaciones y aumentar el riesgo de regresión.
- **Documentos afectados:** `README.md`, `docs/03-desarrollo/estandar-diseno-visual.md`, `docs/03-desarrollo/especificacion-frontend-nava.md`, `docs/03-desarrollo/estandar-frontend-vue.md`, `docs/04-arquitectura/frontend.md`, `docs/07-calidad/02-checklist-transversal.md`, `docs/README.md`, `docs/00-control/matriz-trazabilidad.md`, `docs/10-backlog/prompts/README.md` y el prompt de orquestación NAVA. `apps/web` queda deliberadamente sin cambios hasta que exista un issue de implementación.
- **Fuente:** selección explícita del propietario de la propuesta NAVA / Tailored Grid y solicitud de orientar el diseño hacia una estética más seria y elegante, 2026-09-01.

**Actualización posterior:** `DEC-078` conserva la identidad y la intención visual de esta decisión, pero retira el carácter obligatorio de sus tokens, medidas, tema único, tecnología de estilos y prohibiciones estéticas o de estilos locales.

### DEC-078 · Libertad de creación visual dentro de NAVA / Tailored Grid

- **Fecha:** 2026-09-02.
- **Decisión:** se retiran las restricciones visuales prescriptivas que fijaban paleta, tokens, tipografías, escalas, radios, sombras, layouts, inventario de componentes, tema y tecnología de estilos. La creación del diseño de cada pantalla y flujo de la plataforma es libre, siempre que respete la dirección NAVA / Tailored Grid elegida el 2026-09-01. CSS, Tailwind CSS, CSS Modules, utility-first, una biblioteca visual u otra herramienta quedan permitidos según lo que mejor materialice el diseño y el resultado se justifique cuando agregue una dependencia.
- **Qué permanece obligatorio:** buenas prácticas de accesibilidad y UX, HTML semántico, teclado y foco visible, WCAG 2.2 AA, contraste real, señales que no dependan solo del color, controles cómodos, responsive sin pérdida de contenido, estados de carga/vacío/error/conflicto/éxito, movimiento reducido, rendimiento, pruebas, seguridad, privacidad, contratos, reglas de negocio y alcance P0/P1/P2.
- **Qué no autoriza:** no convierte el mockup en fuente de nuevas funciones, no cambia el vocabulario `turno`, no añade settings o personalización por tenant, no permite una apariencia genérica que contradiga NAVA y no modifica API, datos, permisos, disponibilidad, auditoría, idempotencia o aislamiento entre barberías.
- **Convivencia con `DEC-039` y `DEC-077`:** conserva de esas decisiones la identidad NAVA / Tailored Grid, la intención editorial y las garantías de calidad; sustituye su carácter obligatorio en materia de tokens, paleta, medidas, componentes, tema único, tecnología de estilos y prohibiciones estéticas o de estilos locales. Las referencias concretas del mockup quedan como orientación, no como contrato visual.
- **Herramientas:** no se adopta Tailwind ni otra dependencia automáticamente en este cambio documental. Cada issue de implementación puede elegir CSS u otra herramienta y debe documentar, cuando aplique, versión, licencia, alcance, accesibilidad, mantenimiento e impacto en el bundle.
- **Documentos afectados:** `AGENTS.md`, `docs/03-desarrollo/estandar-diseno-visual.md`, `docs/03-desarrollo/especificacion-frontend-nava.md`, `docs/03-desarrollo/estandar-frontend-vue.md`, `docs/03-desarrollo/estrategia-pruebas.md`, `docs/04-arquitectura/frontend.md`, `docs/02-requisitos/historias-usuario.md` (`HU-009`), `docs/07-calidad/02-checklist-transversal.md`, `docs/README.md`, `README.md`, `docs/00-control/matriz-trazabilidad.md`, `docs/00-control/historial-cambios.md` y el prompt de orquestación NAVA. No modifica el código de `apps/web`.
- **Fuente:** instrucción explícita del propietario del 2026-09-02 para eliminar los estándares rígidos de diseño, conservar el diseño elegido el día anterior y permitir la herramienta necesaria para implementarlo; issue documental `#166`.

**Actualización posterior:** `DEC-079` conserva la libertad de composición, medidas, componentes y tecnología, pero restablece para rediseños y pantallas nuevas el carácter obligatorio de la firma cromática NAVA y de la familia visual mostrada en los mockups aprobados.

### DEC-079 · Contrato visual mínimo para rediseños y pantallas nuevas

- **Fecha:** 2026-09-02.
- **Decisión:** toda pantalla, flujo o componente visible nuevo y todo rediseño completo debe pertenecer de forma reconocible a la familia de los mockups NAVA aprobados el 2026-09-02. Son obligatorios los anclajes cromáticos de tinta `#101B2B`, marfil `#F4F0E7`, blanco `#FFFFFF`, grafito `#2A2D32`/`#5E625F`, piedra `#E8E2D8`/`#C9C0B2`, salvia `#748477` y latón `#B8955A`/`#765C2F`, junto con la relación entre voz editorial serif y controles funcionales sans-serif. Las paletas semánticas de éxito, advertencia/conflicto, peligro, información e inactivo quedan fijadas en el estándar visual y el color nunca comunica solo.
- **Libertad que se conserva:** cada pantalla decide composición, grid, tamaños, espaciado, radios, sombras, movimiento, navegación, componentes y tecnología de estilos. Puede usar CSS, CSS Modules, `<style scoped>`, Tailwind, utility-first o una biblioteca visual; una dependencia nueva se justifica. Se permiten tintes y derivados accesibles de la paleta, y un color adicional solo cuando cubre una necesidad semántica o de legibilidad documentada sin crear otra identidad de marca.
- **Aplicación incremental:** la regla se aplica al crear o rediseñar. No obliga a migrar de inmediato pantallas existentes que el issue no toca y una corrección aislada no se convierte automáticamente en un rediseño. Cuando sí se rediseña una pantalla, se entrega completa en identidad, estados, responsive y accesibilidad, sin regiones principales a medio migrar.
- **Mockups aprobados:** `docs/10-backlog/evidence/ui-redesign-nava-2026-09-02/01-autenticacion-movil.png`, `02-nuevo-turno-responsive.png`, `03-componentes-formularios-alertas.png` y `04-servicios-configuracion-navegacion.png`. Son contrato de familia visual y patrones para las áreas representadas, no fuente de datos, copy, funciones ni coordenadas de píxel.
- **Qué no autoriza:** no crea pagos, caja, reportes, inventario, cuentas de cliente, vista multi-barbero, personalización por tenant ni ninguna capacidad fuera del alcance; no reemplaza reglas, HU, contratos, seguridad, privacidad, idempotencia o aislamiento entre barberías.
- **Convivencia con `DEC-077` y `DEC-078`:** concreta la identidad de `DEC-077` y matiza únicamente la libertad cromática y el carácter puramente orientativo de los mockups declarados por `DEC-078`. Permanecen vigentes la libertad compositiva y técnica de `DEC-078` y todas las garantías de calidad.
- **Documentos afectados:** `AGENTS.md`, `docs/03-desarrollo/estandar-diseno-visual.md`, `docs/03-desarrollo/especificacion-frontend-nava.md`, `docs/03-desarrollo/estandar-frontend-vue.md`, `docs/04-arquitectura/frontend.md`, `docs/02-requisitos/historias-usuario.md`, `docs/00-control/matriz-trazabilidad.md`, `docs/00-control/historial-cambios.md`, `docs/README.md`, el handoff de mockups y el prompt de orquestación NAVA. No modifica `apps/web`.
- **Fuente:** instrucción explícita del propietario del 2026-09-02: convertir los mockups generados en estándar para rediseños y pantallas nuevas, conservar un apartado libre y respetar los colores e identidad definidos; issue documental `#166`.

**Actualización posterior:** `DEC-080` conserva la libertad creativa cuando no existe una referencia exacta, pero convierte un mockup expresamente asignado en contrato visual medible para el viewport y estado que representa.

### DEC-080 · Fidelidad medible cuando se asigna un mockup exacto

- **Fecha:** 2026-09-03.
- **Decisión:** el trabajo visual tiene dos modos explícitos. En **identidad guiada**, sin una referencia exacta asignada, se conserva la libertad compositiva y técnica de `DEC-078` dentro de la identidad de `DEC-079`. En **fidelidad al mockup**, cuando un issue, prompt o instrucción asigna una imagen concreta a una pantalla/componente o pide igualarla, reproducirla, implementarla o corregirla contra ella, la imagen es contrato visual para el viewport y estado representados. La app real debe reproducir colores, proporciones, jerarquía tipográfica, escala de controles e iconos, centrado, alineaciones, densidad, bordes y estados; no basta con usar la paleta o contener elementos parecidos.
- **Verificación:** antes de editar se abre la referencia a resolución original, se mide su estructura y se captura una línea base de la app. Después se renderiza la app real en el mismo viewport, se comprueba el tamaño efectivo, y se revisan captura anterior/final, comparación lado a lado y overlay o diff. Métricas DOM, estilos computados, pruebas automatizadas y screenshots aislados son evidencia auxiliar, no sustituyen la comparación visual. Ninguna entrega se declara terminada con diferencias primarias sin clasificar.
- **Tolerancia y adaptación:** el estándar visual fija tolerancias de revisión para regiones, texto e iconos. Una diferencia visible importante se corrige aunque caiga dentro del umbral. Se permiten desviaciones únicamente por contenido o contrato real, regla superior, accesibilidad, privacidad, seguridad, responsive, rendimiento o rasterización del navegador; se documentan la causa y el tratamiento.
- **Libertad que se conserva:** CSS, CSS Modules, estilos scoped, utility-first, Tailwind o bibliotecas siguen permitidos con las justificaciones vigentes. En fidelidad, esa libertad corresponde a cómo producir el resultado y a decisiones no representadas, no a reinterpretar la geometría visible asignada. En viewports que el mockup no representa, el reflow sigue libre siempre que conserve la jerarquía y las garantías de calidad.
- **Frontera funcional:** un mockup nunca crea datos, copy normativo, funciones, endpoints, permisos, estados de dominio ni destinos. Si una representación visual contradice una autoridad superior, se conserva la autoridad y se registra la desviación.
- **Ayuda de ejecución:** `.claude/skills/nava-mockup-fidelity/SKILL.md` operacionaliza esta decisión para Claude Code y `.claude/skills/browser-viewport-verification/SKILL.md` demuestra el viewport efectivo. Son ayudas derivadas; no sustituyen este registro ni el estándar.
- **Convivencia con `DEC-077`–`DEC-079`:** mantiene la identidad, paleta y libertad técnica ya aprobadas. Matiza solo la libertad de composición y medidas cuando el propietario ha entregado una referencia exacta para ese resultado.
- **Documentos afectados:** `AGENTS.md`, `CLAUDE.md`, `.claude/{CLAUDE.md,settings.json,skills/{nava-mockup-fidelity,browser-viewport-verification}/}`, `docs/03-desarrollo/{estandar-diseno-visual.md,especificacion-frontend-nava.md,estrategia-pruebas.md}`, `docs/00-control/{matriz-trazabilidad.md,historial-cambios.md}` y el prompt de orquestación NAVA v3. La configuración del proyecto fija `claude-sonnet-5` con `effortLevel: high`; no modifica `apps/web`.
- **Fuente:** instrucción explícita del propietario tras comparar el login implementado con el mockup —centrado, proporciones, color, escala tipográfica e iconos no podían quedar a interpretación— y solicitud del 2026-09-03 de actualizar prompts o skills para evitar la repetición; issue documental `#208`.

**Actualización posterior:** `DEC-086` traslada la ayuda de ejecución de esta decisión al skill compartido `.agents/skills/visual-qa/` (Claude Code y Codex); `nava-mockup-fidelity` y `browser-viewport-verification` quedan como alias heredados. La decisión no cambia.

### DEC-081 · Canal configurable y destino verificado para códigos OTP de autenticación

- **Fecha:** 2026-09-03.
- **Decisión:** los códigos OTP de recuperación y del reto adicional de acceso se entregan conforme a la configuración del evento de la barbería: **correo electrónico por defecto**, **WhatsApp oficial** o **ambos**. Cuando se habilitan ambos, una emisión usa el mismo código y conserva una sola operación lógica, aunque produzca dos intentos de entrega trazables. La interfaz informa el canal o combinación configurada, pero no pide ni acepta un correo o teléfono de destino escrito por la persona durante el reto.
- **Destino y seguridad:** el servidor resuelve exclusivamente contactos verificados y almacenados para la cuenta dentro de la barbería. Antes de que el contrato vigente autorice revelar un destino, el texto permanece condicional y uniforme (`Si la cuenta existe...` / `Si la cuenta puede continuar...`); ningún estado distingue cuenta inexistente, código incorrecto, vencido o agotado. Si el canal configurado no tiene un contacto verificado utilizable, la operación responde de forma uniforme y registra internamente el fallo sin probar silenciosamente otro canal no configurado.
- **Configuración:** “cliente” en la instrucción del propietario se interpreta como la barbería/tenant que configura el sistema, no como una cuenta del consumidor final; el producto no crea cuentas de cliente. Correo es el valor inicial cuando todavía no existe una preferencia explícita. Cambiar el canal requiere la capacidad de configuración, permisos y contrato correspondientes.
- **Impacto sobre decisiones previas:** sustituye la obligación de WhatsApp único del reto en `DEC-062`, sustituye el envío siempre simultáneo de recuperación de `DEC-051` y descarta la solicitud intermedia de “exactamente un canal” registrada en `DP-NOT-06`. Conserva los proveedores oficiales de `DEC-066`, la no enumeración de `DEC-065`, la vigencia/intentos de `DEC-064` y la evidencia por intento de `RN-REC-04`.
- **Separación de alcance:** esta decisión resuelve `DP-NOT-06`, `DP-SEG-13`, `CT-009` y `CT-010`, pero no autoriza a implementar configuración, contrato, proveedores o persistencia dentro del issue visual `#188`. El atlas de autenticación representa las variantes para evitar ambigüedad; la adaptación funcional requiere issue y pruebas propios antes de modificar API o backend.
- **Documentos afectados:** `docs/00-control/{dudas-pendientes.md,contradicciones.md,matriz-trazabilidad.md}`, `docs/01-producto/reglas-negocio.md`, `docs/02-requisitos/historias-usuario.md`, `docs/03-desarrollo/{estandar-diseno-visual.md,especificacion-frontend-nava.md}` y `docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/auth-eventos/README.md`.
- **Fuente:** observación explícita del propietario del 2026-09-03 al revisar los mockups: correo por defecto, opción de WhatsApp o ambos según configuración del sistema, notificación unificada y un mockup separado por evento/canal.

**Actualización posterior:** `DEC-092` (2026-09-19) sustituye esta decisión únicamente para el paso 1 de la recuperación de acceso: la persona elige WhatsApp o correo y escribe su valor. El reto adicional de acceso conserva `DEC-081` sin cambios.

### DEC-082 · Resolución de `DP-PUB-01`: generación, mutabilidad y conducta al deshabilitar del slug público

- **Fecha:** 2026-09-10.
- **Decisión:** el identificador público del enlace de reservas de una barbería (`barbershop.public_slug`) se **genera automáticamente a partir del nombre** de la barbería (slugify: minúsculas, dígitos y guiones, formato `^[a-z0-9]([a-z0-9-]{1,38}[a-z0-9])$` ya propuesto en `database/modelo-fisico-referencia.sql`) la primera vez que el nombre se guarda; una colisión de unicidad global añade un sufijo numérico determinístico (`-2`, `-3`, ...) hasta resolverla. El barbero no lo escribe a mano en la creación.
- **Nota (2026-10-08, `DEC-117`):** el sufijo numérico determinístico de esta decisión se sustituye por un código aleatorio y la edición manual sigue sin implementarse; el resto —respuesta uniforme, unicidad global, resolución sin contexto de tenant— se mantiene.
- **Mutabilidad:** el slug es **editable** desde la configuración de la barbería (misma pantalla/capacidad que `HU-020`), sujeta a la misma validación de formato y unicidad global. Al cambiarlo, el slug anterior **deja de resolver**: cualquier visita posterior a ese enlace recibe la misma respuesta uniforme que un identificador inexistente (`CA-090-02`); no hay redirección ni periodo de gracia. No se conserva un historial de slugs previos como capacidad activa del MVP.
- **Conducta al deshabilitar:** cuando la barbería no está habilitada para reservas públicas (inactiva, o el slug no es publicable por cualquier motivo de negocio), el enlace responde con la **misma respuesta genérica y uniforme** que un identificador mal formado, desconocido o de otro tenant — sin distinguir "no existe" de "existe pero está apagado" y sin exponer IDs internos. Esto ya lo exige `CA-090-02` como parte del aislamiento de tenant; esta decisión confirma que la conducta de deshabilitar no introduce una tercera respuesta distinguible.
- **Unicidad y resolución:** el slug se compara sin distinguir mayúsculas y es único **globalmente** (no por tenant), porque se resuelve antes de que exista contexto de tenant alguno — coherente con el índice único parcial `idx_barbershop_public_slug` y la función `SECURITY DEFINER` ya bosquejados en `database/modelo-fisico-referencia.sql`.
- **Responsable:** propietario del proyecto.
- **Motivo:** generar el slug automáticamente evita que crear una barbería dependa de que el barbero acierte un formato técnico antes de operar; hacerlo editable (con enlace viejo roto, sin redirección) da margen para corregir un nombre comercial sin acumular la complejidad de mantener alias históricos, que ningún criterio de aceptación exige todavía. La respuesta uniforme al deshabilitar reutiliza exactamente la garantía de no fuga que `CA-090-02` y `RN-TEN-01` ya exigen para cualquier slug inválido, sin inventar un tercer estado público.
- **Alternativas descartadas:** slug elegido manualmente por el barbero — descartada porque introduce un paso de validación adicional en la creación sin que ninguna historia lo requiera; slug editable con redirección del enlace viejo — descartada por el costo de mantener alias históricos y resolución en cadena en la ruta pública, sin fundamento contractual (`CA-090-02` no lo exige); mensaje específico de "barbería en pausa" al deshabilitar — descartada por contradecir directamente la no-fuga exigida por `CA-090-02`/`RN-TEN-01` (revelaría que el negocio existió).
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-PUB-01`), `docs/01-producto/reglas-negocio.md` (`RN-TEN-01`, `RN-DAT-01`, `RN-DAT-02`, caso límite del slug), `docs/02-requisitos/historias-usuario.md` (`HU-090`, `CA-090-01`/`CA-090-02`), `docs/10-backlog/prompts/hu/hu-090-entrada-publica-reservas.md`, `docs/10-backlog/plan-bloques.md` (cierre de B3, desbloqueo de B4); futura implementación de `HU-090` aplica slugify + sufijo determinístico en la escritura del nombre de la barbería, valida formato/unicidad global en el mismo `UPDATE`, y resuelve el enlace público con la respuesta uniforme única para inválido/desconocido/deshabilitado.
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-PUB-01`; aprobación explícita del propietario el 2026-09-10 (generación automática, edición con ruptura de enlace viejo, respuesta uniforme al deshabilitar).

### DEC-083 · Resolución de `DP-PUB-02`: rangos configurables y política de cancelación tardía por defecto

- **Fecha:** 2026-09-11.
- **Decisión:** los valores iniciales de `RN-DIS-04` (anticipación mínima 60 min, ventana máxima 3 días), `RN-DIS-06` (rejilla 15 min) y `RN-CAN-01` (plazo de cancelación 20 min) quedan confirmados por `DEC-005`, `DEC-006`, `DEC-010` y `DEC-018`; esta decisión fija además el **rango permitido** que la configuración de cada barbería puede usar alrededor de esos valores:
  - Anticipación mínima: entero en minutos, rango `[0, 1440]` (0 desactiva el mínimo; tope de 24 horas).
  - Ventana máxima: entero en días, rango `[1, 90]`.
  - Rejilla de horarios: entero en minutos, restringido al conjunto `{5, 10, 15, 20, 30, 60}` (evita pasos arbitrarios que compliquen la proyección de franjas).
  - Plazo de cancelación del cliente: entero en minutos, rango `[0, 10080]` (0 desactiva la cancelación propia del cliente; tope de 7 días).
  - Política de cancelación fuera de plazo (`RN-CAN-02`): dos campos booleanos independientes, `permite_cliente` (si el cliente puede cancelar vencido el plazo) y `motivo_obligatorio` (si esa cancelación exige motivo). **Default de una barbería nueva:** `permite_cliente = true`, `motivo_obligatorio = true` (el cliente puede cancelar tarde, pero debe registrar el motivo). El barbero conserva siempre su facultad de `RN-CAN-03` sin plazo ni motivo.
- **Responsable:** propietario del proyecto.
- **Motivo:** sin un rango explícito, la validación de `HU-093` no tiene contra qué comparar la entrada del formulario y cualquier límite quedaría inventado por quien implemente. Los rangos elegidos son generosos alrededor de los valores iniciales ya confirmados (no los sustituyen) y evitan configuraciones absurdas (anticipación de una semana, rejilla de 7 minutos) sin impedir casos de negocio razonables. El default `permite_cliente = true` con motivo obligatorio prioriza la flexibilidad del cliente sobre el bloqueo total, dejando trazabilidad del motivo para disputas.
- **Alternativas descartadas:** dejar la rejilla como entero libre — descartada porque una rejilla no múltiplo de 5 complica la lectura humana de las franjas sin beneficio; default `permite_cliente = false` (solo barbero) — descartada por instrucción explícita del propietario, que prefiere no bloquear al cliente por defecto.
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-PUB-02`), `docs/01-producto/reglas-negocio.md` (`RN-DIS-04`, `RN-DIS-06`, `RN-CAN-01`, `RN-CAN-02`), `docs/10-backlog/prompts/hu/hu-093-configuracion-reserva-cancelacion.md`, `docs/10-backlog/prompts/hu/hu-094-motor-disponibilidad-publica.md`, `docs/10-backlog/plan-bloques.md`; futura implementación de `HU-093` valida estos rangos en el `CHECK`/dominio de aplicación y en el formulario, con los defaults aquí fijados para barberías sin configuración explícita.
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-PUB-02`; aprobación explícita del propietario el 2026-09-11 (propón rangos razonables; default de cancelación tardía: cliente también puede, motivo obligatorio).

### DEC-084 · Resolución de `DP-PUB-03`: reinicio de la rejilla tras una interrupción

- **Fecha:** 2026-09-11.
- **Decisión:** cuando un bloqueo, cita o cualquier otra restricción termina en un instante que no coincide con la rejilla original del tramo laboral, la generación de franjas **reinicia el conteo desde ese instante** (no conserva el anclaje original del tramo). Esto confirma como definitiva la propuesta que `RN-DIS-06` ya dejaba anotada en sus casos límite.
- **Ejemplo:** jornada 9:00–18:00 con rejilla de 15 min y un bloqueo que termina a las 13:47; la siguiente franja ofrecida es 13:47, y desde ahí se generan 14:02, 14:17… hasta el siguiente evento o el cierre del tramo, en vez de esperar hasta 14:00 (próximo múltiplo de la rejilla original del tramo).
- **Alcance:** aplica a cualquier interrupción dentro de un mismo tramo laboral (bloqueo, cita, excepción parcial); no reabre `RN-DIS-04` (anticipación/ventana) ni `RN-CON-01`/`RN-CON-03` (solapes), que se siguen aplicando después de generar las franjas.
- **Responsable:** propietario del proyecto.
- **Motivo:** reiniciar desde el instante real aprovecha huecos que de otro modo se perderían (p. ej. el hueco de 13 minutos entre 13:47 y 14:00 en el ejemplo), y es la única lectura de `RN-DIS-06` que ya tenía redacción previa como propuesta; conservar el anclaje original exigiría inventar una regla nueva sin base en el documento de reglas de negocio.
- **Alternativas descartadas:** conservar el anclaje del tramo original (franjas siempre en los mismos minutos fijos del día) — descartada por desperdiciar huecos reales y por no tener respaldo en `RN-DIS-06`, que solo registraba el reinicio como propuesta.
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-PUB-03`), `docs/01-producto/reglas-negocio.md` (`RN-DIS-06`, confirma el caso límite), `docs/02-requisitos/historias-usuario.md` (`HU-094`), `docs/10-backlog/prompts/hu/hu-094-motor-disponibilidad-publica.md`; futura implementación de `HU-094` reinicia el punto de generación de la rejilla en el instante exacto en que termina cada restricción dentro del tramo.
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-PUB-03`; aprobación explícita del propietario el 2026-09-11 (reinicia desde el instante).

### DEC-085 · Resolución de `DP-PUB-04`: reconciliación pública de `customer` por teléfono o correo

- **Fecha:** 2026-09-11.
- **Decisión:** en la reserva pública, el `customer` se reconcilia dentro de la barbería activa por **teléfono o correo, sin exigir que ambos coincidan**: una persona puede cambiar de correo conservando el teléfono, o de teléfono conservando el correo, y muy rara vez cambia ambos a la vez.
  - Si el teléfono **o** el correo dado coincide con una fila existente (y el otro campo no coincide con ninguna fila, o coincide con la misma fila), se **reutiliza esa fila** y se actualiza el dato que cambió (el campo no coincidente se sobrescribe con el valor nuevo dado).
  - Si teléfono y correo coinciden cada uno con una fila **distinta** (conflicto), **no se fusionan**: se **crea un `customer` nuevo** con los datos dados en vez de adivinar cuál de las dos identidades es la correcta.
  - Todo lo anterior ocurre siempre dentro del tenant (`RN-TEN-01`); nunca se reconcilia contra un `customer` de otra barbería.
- **Responsable:** propietario del proyecto.
- **Motivo:** exigir coincidencia de ambos campos perdería la reconciliación en el caso común (cambiar uno de los dos datos de contacto con el tiempo); fusionar automáticamente ante un conflicto de filas distintas arriesga mezclar identidades reales por una coincidencia parcial, algo que ninguna historia exige y que sería difícil de deshacer después.
- **Alternativas descartadas:** teléfono como clave única de reconciliación (ignora coincidencias de solo correo) — descartada porque el propietario indicó explícitamente que cualquiera de los dos puede ser el dato estable; correo como clave única — descartada por el mismo motivo; fusión automática eligiendo un campo como desempate en el caso conflictivo — descartada por el propietario a favor de no fusionar y crear un registro nuevo.
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-PUB-04`), `docs/02-requisitos/historias-usuario.md` (`HU-096`, `CA-096-04`), `docs/10-backlog/prompts/hu/hu-096-datos-cliente-persona-atendida.md`; futura implementación de `HU-096` aplica esta política exactamente en el paso de reconciliación de `customer` antes de crear la cita.
- **Fuente:** `docs/00-control/dudas-pendientes.md`, `DP-PUB-04`; aprobación explícita del propietario el 2026-09-11 (coincidencia por cualquiera de los dos campos; conflicto entre filas distintas crea un `customer` nuevo en vez de fusionar).

### DEC-086 · Infraestructura compartida de agentes: AGENTS.md como autoridad, skills canónicos y Graphify opcional

- **Fecha:** 2026-09-13.
- **Decisión:** Claude Code y Codex comparten una sola infraestructura de agentes:
  - `AGENTS.md` es la autoridad común. `CLAUDE.md` la importa (`@AGENTS.md`) y solo añade lo específico de Claude Code, incluida la autorización permanente de squash-merge del propietario. Se retira `.claude/CLAUDE.md`.
  - Los skills del proyecto viven en `.agents/skills/<skill>/SKILL.md`: `task-brief`, `ui-direction`, `visual-qa`, `change-review` y `generacion-mockups-nava`. `.claude/skills/<skill>/SKILL.md` es un adaptador con descripción idéntica que ordena leer el canónico. No se crea un skill nuevo sin un fallo repetido y demostrable.
  - `visual-qa` absorbe el contenido de `nava-mockup-fidelity` y `browser-viewport-verification`, que quedan como alias no invocables por el modelo para preservar las rutas citadas por prompts persistentes.
  - Graphify es opcional: sin hooks `PreToolUse` y con el skill invocable solo mediante `/graphify`.
  - `tools/ai/validate-agent-system.sh --strict` comprueba la coherencia y debe pasar tras cambiar skills o instrucciones de agentes.
- **Responsable:** propietario del proyecto.
- **Motivo:** el conocimiento de NAVA, las pruebas y la trazabilidad ya existían, pero se activaban de forma desigual entre agentes. `CLAUDE.md` duplicaba reglas y los hooks de Graphify añadían latencia, contexto y permisos en cada lectura o búsqueda, además de versionar una ruta absoluta de una sola máquina.
- **Alternativas descartadas:** instalar completos el plugin Product Design de OpenAI o el repositorio de skills de Anthropic (amplían permisos y contexto, y duplican NAVA/Playwright). Crear skills separados de prompt-architect, design-system, creative-director, design-critic, UX o revisores de arquitectura, frontend y código (añaden triggers redundantes). Adaptadores con `@import` (Claude Code no los expande dentro de `SKILL.md`) o con symlinks (frágiles en Windows). Borrar los skills visuales antiguos (rompería prompts persistentes).
- **Convivencia con `DEC-080`:** la decisión de fidelidad no cambia; solo cambia la ayuda derivada que la operacionaliza.
- **Documentos afectados:** `AGENTS.md`, `CLAUDE.md`, `.claude/{CLAUDE.md (retirado),settings.json,skills/**}`, `.agents/skills/**`, `tools/ai/validate-agent-system.sh`, `docs/03-desarrollo/auditoria-infraestructura-agentes-ia.md` y `docs/00-control/historial-cambios.md`. No modifica `apps/`, CI ni `graphify-out/`.
- **Fuente:** auditoría de infraestructura de agentes del 2026-09-13 y solicitud explícita del propietario de implementarla; issue [#259](https://github.com/bcaceres19/barberia/issues/259).

### DEC-087 · `graphify-out/` es local y regenerable; no se versiona

- **Fecha:** 2026-09-13.
- **Decisión:** el grafo de Graphify (`graphify-out/`) deja de versionarse. Se ignora completo en `.gitignore`, se retira del índice con `git rm -r --cached` sin reescribir la historia y cada persona lo regenera en local con `graphify update .` cuando quiera usarlo. Ningún prompt, estándar ni flujo puede exigir que exista en un clon.
- **Responsable:** propietario del proyecto.
- **Motivo:** son 107 MB y 110 archivos derivados del código que cambian con cada `graphify update`. Los hooks locales `post-commit`/`post-checkout` de Graphify los reconstruyen en cada commit o cambio de rama, lo que dejaba el árbol de trabajo sucio y producía diffs y conflictos de merge ajenos a cada entrega. Desde `DEC-086` Graphify es opcional, así que no hay motivo para compartir su salida.
- **Alternativas descartadas:** seguir versionándolo (ruido y conflictos permanentes); publicarlo fuera del repositorio (añade infraestructura sin un consumidor que la necesite); purgar sus blobs de la historia de Git (reescribe `main`, contrario a `DEC-038`).
- **Verificación:** `tools/ai/validate-agent-system.sh --strict` advierte si `graphify-out/` vuelve a estar versionado o deja de estar ignorado.
- **Documentos afectados:** `.gitignore`, `tools/ai/validate-agent-system.sh`, `docs/03-desarrollo/auditoria-infraestructura-agentes-ia.md` y `docs/00-control/historial-cambios.md`.
- **Fuente:** instrucción explícita del propietario del 2026-09-13 ("déjalo local, no lo subas al repo, es algo que se puede generar"); issue [#261](https://github.com/bcaceres19/barberia/issues/261).

### DEC-088 · Momentos de actualización del grafo local y skill `graphify-refresh`

- **Fecha:** 2026-09-13.
- **Decisión:** el grafo local de Graphify se actualiza solo cuando va a usarse, nunca por evento de git:
  - justo antes de consultarlo, o al empezar una tarea que cruza módulos, si se integró código desde su `built_at_commit`: `graphify update .` (local, sin IA);
  - tras un bloque documental grande desde la última reextracción semántica (por defecto 30 archivos): se **propone** `/graphify . --update` (incremental) y solo se ejecuta con confirmación del propietario, porque usa IA. La construcción completa `/graphify` queda para cuando no existe grafo;
  - nunca después de cada commit, cambio de rama o cambio solo documental por debajo del umbral.
- **Implementación:** `tools/ai/graphify-freshness.sh` (solo lectura) calcula la recomendación (`SKIP`, `UPDATE`, `SEMANTIC_UPDATE_SUGGESTED`, `BUILD_FIRST`) y avisa si los hooks git de Graphify reaparecen. El skill compartido `graphify-refresh` la aplica. Esta decisión amplía el catálogo de `DEC-086` por instrucción explícita del propietario.
- **Responsable:** propietario del proyecto.
- **Motivo:** tras retirar los hooks (`DEC-087`), el grafo dejaría de reflejar el código si nadie lo actualiza. Actualizarlo en cada evento reintroduce el gasto eliminado. Hacerlo al consultarlo mantiene respuestas correctas con coste mínimo.
- **Alternativas descartadas:** volver a los hooks `post-commit`/`post-checkout` o a `graphify watch` (coste continuo); tarea programada con cron (actualiza aunque nadie consulte); criterio manual sin script (no verificable).
- **Documentos afectados:** `AGENTS.md`, `CLAUDE.md`, `.agents/skills/graphify-refresh/`, `.claude/skills/graphify-refresh/`, `tools/ai/graphify-freshness.sh` y `docs/00-control/historial-cambios.md`.
- **Fuente:** instrucción explícita del propietario del 2026-09-13 de generar un skill que actualice Graphify en los momentos recomendados; issue [#264](https://github.com/bcaceres19/barberia/issues/264).

### DEC-089 · Resolución de `DP-PUB-05`: entropía, vigencia, rotación y revocación del token de acceso al turno

- **Fecha:** 2026-09-15.
- **Decisión:** el token de acceso público al turno (`F-PUB-07`, `appointment_access_token`) se genera como 32 bytes aleatorios criptográficos (256 bits, `crypto/rand`), viaja una sola vez en la URL del enlace enviado por correo y el servidor conserva únicamente `SHA-256(token)` en `token_hash` -mismo criterio que ya fija `database/modelo-fisico-referencia.sql` sección E.1-. Vigencia: 90 días desde `issued_at`, sin rotación en cada consulta (el mismo enlace sirve para toda su vigencia; leer o cancelar el turno no emite un token nuevo). Revocación: inmediata al anonimizar la cita (ya exigido por `DEC-049`) y al cancelarla (`HU-099`). Emisión: una sola vez, dentro de la misma transacción atómica que `HU-097` usa para confirmar la cita; el MVP no ofrece reemisión ni reenvío del enlace.
- **Responsable:** propietario del proyecto.
- **Motivo:** 256 bits de entropía hacen inviable la fuerza bruta sin necesitar HMAC (a diferencia del código numérico de recuperación de `DEC-064`, cuyo espacio de 10⁶ sí lo exige); 90 días cubre consultar la cita, su política de cancelación y el historial reciente sin dejar el enlace vigente indefinidamente; no rotar evita que el cliente pierda acceso si vuelve a abrir el mismo correo días después; emitir una sola vez en la transacción de confirmación evita una superficie de reemisión no aprobada por ninguna historia.
- **Alternativas descartadas:** vigencia de 30 días (más estricta, pero corta el acceso a una política de cancelación o un historial que el cliente aún podría necesitar consultar); token de un solo uso que rota en cada acceso (más restrictivo, pero rompe "volver a ver mi turno" desde el mismo enlace guardado); reenvío del enlace por contacto verificado (amplía el alcance de `HU-098` con una operación y una superficie de abuso nuevas, sin que ninguna historia lo pida).
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-PUB-05`), `docs/02-requisitos/historias-usuario.md` (`HU-097`, `HU-098`, `HU-099`), `docs/10-backlog/prompts/hu/{hu-097-confirmacion-publica-concurrente.md,hu-098-acceso-cliente-turno.md,hu-099-cancelacion-publica.md}`.
- **Fuente:** decisión explícita del propietario del 2026-09-15 (opción recomendada, `AskUserQuestion`), issue documental [#240](https://github.com/bcaceres19/barberia/issues/240).

### DEC-090 · Resolución de `DP-PUB-06`: orden y acotación de las franjas alternativas tras perder una carrera de confirmación

- **Fecha:** 2026-09-15.
- **Decisión:** al perder la carrera de confirmación de una franja (`HU-097`, `RN-CON-05`), el servidor calcula, dentro del mismo barbero y el mismo servicio, hasta 3 franjas válidas cronológicamente más cercanas a la elegida (antes o después, dentro de la ventana pública vigente). Si el día civil de la franja perdida ya no tiene ninguna franja libre restante, la búsqueda salta directo al primer inicio disponible del siguiente día con franjas, siempre con el mismo barbero -nunca ofrece otro barbero ni otro servicio.
- **Responsable:** propietario del proyecto.
- **Motivo:** mantiene el contexto que el cliente ya eligió (mismo barbero, mismo servicio) en vez de forzarlo a decidir de nuevo desde cero; acotar a 3 evita una lista larga que retrase la recuperación de un conflicto que ya de por sí es una interrupción; saltar al siguiente día con franjas evita una respuesta vacía cuando el día completo se agotó.
- **Alternativas descartadas:** incluir otros barberos elegibles del mismo servicio cuando el elegido no tiene nada cercano (más útil en el caso límite, pero introduce un cambio implícito de barbero que ninguna historia pide y complica la UI/las pruebas); no calcular alternativas y reenviar a la pantalla completa de disponibilidad de `HU-095` (más simple de construir, pero pierde el contexto de "casi lo lograste" que motiva `RN-CON-05`).
- **Documentos afectados:** `docs/00-control/dudas-pendientes.md` (cierra `DP-PUB-06`), `docs/01-producto/reglas-negocio.md` (`RN-CON-05`), `docs/02-requisitos/historias-usuario.md` (`HU-097`), `docs/10-backlog/prompts/hu/hu-097-confirmacion-publica-concurrente.md`.
- **Fuente:** decisión explícita del propietario del 2026-09-15 (opción recomendada, `AskUserQuestion`), issue documental [#240](https://github.com/bcaceres19/barberia/issues/240).

### DEC-091 · Resolución de `CT-011`: secuencia entre la confirmación por correo de B4 y la maquinaria de notificaciones de B5

- **Fecha:** 2026-09-15.
- **Decisión:** `HU-097` adelanta a B4 el envío mínimo del correo de confirmación con el enlace de acceso al turno (`F-PUB-07`), usando directamente el proveedor ya integrado y verificado en producción (Resend, `PROMPT-TEST-OTP-EMAIL-RESEND-v1`) sin cola, reintentos automáticos ni selección de proveedor -un envío síncrono simple dentro (o inmediatamente después de) la transacción de confirmación-. B5 construye después, sobre esa base ya funcionando, la maquinaria completa de notificaciones (transacción propia, reintentos, múltiples proveedores y recordatorios) sin que `HU-097` deba reescribirse.
- **Responsable:** propietario del proyecto.
- **Motivo:** de las cuatro opciones registradas en `contradicciones.md` (CT-011), esta es la de menor riesgo y menor alcance: reutiliza infraestructura de envío de correo ya probada con código real (OTP vía Resend) en vez de inventar una nueva, no exige adelantar todo el núcleo transaccional de B5 ni posponer el envío real a una fase futura sin fecha, y no simula un envío (el correo sale de verdad).
- **Alternativas descartadas:** dividir `F-PUB-07` para que B4 solo muestre el enlace en pantalla y B5 complete el correo real antes del piloto (el cliente puede perder el enlace si cierra la pestaña antes de que B5 exista); adelantar el núcleo transaccional completo de B5 para que B4 lo consuma desde el inicio (amplía el alcance y el tiempo antes de poder cerrar `HU-097` sin necesidad, dado que un envío síncrono simple ya es suficiente para el piloto).
- **Documentos afectados:** `docs/00-control/contradicciones.md` (cierra `CT-011`), `docs/00-control/dudas-pendientes.md`, `docs/01-producto/alcance-mvp.md` (`F-PUB-07`), `docs/10-backlog/plan-bloques.md` (B4/B5), `docs/02-requisitos/historias-usuario.md` (`HU-097`, `HU-098`), `docs/10-backlog/prompts/hu/{hu-097-confirmacion-publica-concurrente.md,hu-098-acceso-cliente-turno.md}`.
- **Fuente:** decisión explícita del propietario del 2026-09-15 (opción recomendada, `AskUserQuestion`), issue documental [#240](https://github.com/bcaceres19/barberia/issues/240).

### DEC-092 · La persona elige WhatsApp o correo y escribe su valor en el paso 1 de recuperación de acceso

- **Fecha:** 2026-09-19.
- **Decisión:** en el paso 1 de `HU-011` la pantalla ofrece dos botones, **WhatsApp** y **Correo**. Al elegir uno aparece un mensaje que menciona únicamente ese canal y pide su valor —el número de WhatsApp o el correo de la cuenta—, con el mismo texto condicional de hoy (“Si existe, te enviaremos un código de un solo uso por <canal>”). El código se entrega solo por el canal elegido, nunca por ambos. Mientras no se elija un canal la pantalla no muestra campo ni acción de envío.
- **Alcance de la sustitución:** solo para la **recuperación de acceso** (`HU-008`/`HU-011`, paso 1). Sustituye, para ese flujo, el envío por el canal o canales configurados por evento y la regla “la interfaz nunca acepta un destino escrito por la persona” de `DEC-081`, y el envío simultáneo por WhatsApp y correo que el código actual aplica. El reto adicional de acceso de `HU-007` conserva `DEC-081` sin cambios.
- **Lo que se conserva:** la respuesta idéntica exista o no la cuenta (`DEC-065`), la vigencia, los intentos, el reenvío y el token de reinicio de `DEC-064`, los proveedores oficiales de `DEC-066`, el envío únicamente a contactos verificados y el destino enmascarado solo tras verificar el código (`CA-008-06`). Elegir un canal no permite escribir un destino distinto del registrado: el valor escrito solo identifica la cuenta.
- **Responsable:** propietario del proyecto.
- **Motivo:** hoy la pantalla afirma que el código llega “por WhatsApp y correo”, y la persona no controla ni sabe por cuál canal le conviene recibirlo; elegirlo ella evita mensajes por un canal que no usa y hace explícito qué dato debe escribir.
- **Alternativas descartadas:** mantener la configuración por barbería de `DEC-081` sin elección de la persona (no atiende la instrucción); enviar siempre por ambos canales (comportamiento actual, menciona canales que la persona no eligió).
- **Dudas derivadas:** `DP-SEG-14` (identificar la cuenta por teléfono antes de resolver el tenant), `DP-SEG-15` (identificador de los pasos 2 y 3) y `DP-SEG-16` (canal elegido sin contacto verificado o no habilitado), resueltas por `DEC-093` el mismo 2026-09-19.
- **Documentos afectados:** `docs/00-control/{dudas-pendientes.md,matriz-trazabilidad.md,historial-cambios.md}`, `docs/01-producto/reglas-negocio.md`, `docs/02-requisitos/historias-usuario.md` (`HU-008`, `HU-011`).
- **Fuente:** instrucción explícita del propietario del 2026-09-19 al revisar la pantalla “Solicita tu código” en el navegador (dos botones, mensaje solo del canal elegido y solicitud de su valor), issue documental [#274](https://github.com/bcaceres19/barberia/issues/274).

### DEC-093 · Resolución de `DP-SEG-14`, `DP-SEG-15` y `DP-SEG-16`: identificar la cuenta por teléfono, identificador de los pasos 2 y 3 y canal sin contacto verificado

- **Fecha:** 2026-09-19.
- **Decisión:** (1) **Teléfono (`DP-SEG-14`).** El número escrito en el paso 1 se normaliza a E.164 (se descartan espacios, guiones, puntos y paréntesis) y identifica la cuenta solo si coincide con **exactamente una** cuenta activa con ese número verificado en todo el sistema (función `SECURITY DEFINER` `auth_recovery_resolve_phone`, migración `20260919200000`). Con ninguna coincidencia, o con más de una (mismo número en barberías distintas), la respuesta sigue siendo la genérica `202` y no se envía nada. Un valor que no es E.164 es un error de forma `422`, idéntico exista o no la cuenta. (2) **Pasos 2 y 3 (`DP-SEG-15`).** `verify` y `reset-password` reutilizan el mismo canal y valor del paso 1, con esquemas cerrados por canal (`email` o `phone`); el servidor traduce el teléfono al correo con la misma función y sigue usando sin cambios las funciones de recuperación por correo. Un teléfono sin coincidencia única falla con el mismo `401` uniforme que un código incorrecto o un token inválido. Ninguna respuesta devuelve el otro dato completo. (3) **Canal sin contacto verificado (`DP-SEG-16`).** Ambos botones están siempre visibles; si el canal elegido no tiene un contacto verificado utilizable o el proveedor de ese canal no está configurado, la respuesta es la genérica `202`, no se envía nada y nunca se prueba el otro canal (coherente con `DEC-081` y `DEC-065`). El paso 2 nombra solo el canal elegido y ofrece «Elegir otro canal», que vuelve al paso 1 sin revelar por qué no llegó el código.
- **Responsable:** propietario del proyecto.
- **Motivo:** el teléfono no es único entre barberías (`staff_user.phone` no tiene restricción de unicidad), así que elegir una cuenta entre varias dejaría recuperar la de otra barbería; exigir una coincidencia única evita ese secuestro sin añadir una restricción de datos. Reutilizar canal y valor en los tres pasos evita un identificador nuevo y mantiene las funciones de recuperación por correo y sus pruebas intactas. No sustituir un canal por otro conserva la promesa de `DEC-092`.
- **Riesgo residual aceptado:** la búsqueda por teléfono añade una consulta que un atacante podría medir; es el mismo tipo de diferencia de tiempo que ya existe entre una cuenta con y sin código enviado, y las defensas siguen siendo la respuesta idéntica de `DEC-065` y el cooldown y límite de reenvío de `DEC-064`. No se añade un retardo artificial.
- **Alternativas descartadas:** elegir una de varias cuentas con el mismo número (recupera la cuenta equivocada); exigir teléfono único por sistema (cambio de datos que ninguna historia pide y rompe barberías que comparten un número); un identificador nuevo para los pasos 2 y 3 (más superficie sin beneficio); usar el otro canal como respaldo (contradice `DEC-092`).
- **Documentos afectados:** `docs/00-control/{dudas-pendientes.md,matriz-trazabilidad.md,historial-cambios.md}`, `docs/01-producto/reglas-negocio.md`, `docs/02-requisitos/historias-usuario.md` (`HU-008`, `HU-011`), `api/openapi/{paths/public-auth.yaml,components/schemas/Recovery*Request.yaml,CHANGELOG.md}` (0.22.0), `database/{migrations/20260919200000_add_auth_recovery_resolve_phone.sql,tests/hu008_recuperacion_acceso.sql,testdata/hu008_recuperacion_canal.sql,README.md}`, `apps/api/README.md`.
- **Fuente:** aceptación explícita del propietario el 2026-09-19 de las tres respuestas recomendadas de `DP-SEG-14`–`DP-SEG-16` (`AskUserQuestion`), issue [#276](https://github.com/bcaceres19/barberia/issues/276).

### DEC-094 · La recuperación de acceso por correo no exige teléfono verificado

- **Fecha:** 2026-09-19.
- **Decisión:** en el paso 1 de la recuperación de acceso, el requisito de contacto verificado se evalúa **solo sobre el canal elegido**. **Correo:** basta una cuenta activa con ese correo, que ya es su identificador de acceso; el código se envía a ese correo aunque la cuenta no tenga teléfono o su teléfono no esté verificado. **WhatsApp:** se mantiene `DEC-093`: el número identifica la cuenta solo si coincide con exactamente una cuenta activa con ese número verificado. Nunca se exige tener ambos contactos y el otro canal no se usa como respaldo. Un teléfono sin verificar no se devuelve como destino de envío.
- **Alcance de la sustitución:** enmienda `DEC-092` y `DEC-093` únicamente en la frase «sobre contactos verificados» para el canal Correo, y reescribe `CA-008-01` de `HU-008` en esa parte. No cambia la respuesta idéntica de `DEC-065`, la vigencia, los intentos, el cooldown ni el límite de reenvío de `DEC-064`, ni los proveedores de `DEC-066`.
- **Responsable:** propietario del proyecto.
- **Motivo:** la implementación de `DEC-093` filtraba la cuenta por teléfono verificado también con el canal Correo, de modo que una cuenta activa sin teléfono nunca recibía el código y la respuesta 202 genérica lo ocultaba. El esquema no tiene marca de correo verificado y el correo es el identificador con el que la persona ya inicia sesión, así que exigir además un teléfono no añade una prueba de posesión del correo, solo impide usarlo.
- **Riesgo residual aceptado:** quien controle la bandeja del correo de la cuenta puede restablecer su contraseña sin poseer el teléfono. Es el modelo habitual de recuperación por correo; se mitigan con el código de un solo uso de 15 minutos, cinco intentos, cooldown y límite de reenvío de `DEC-064`, y con la revocación de sesiones al cambiar la contraseña.
- **Alternativas descartadas:** marca `email_verified_at` con un flujo de verificación de correo (historia nueva y deja sin recuperación por correo a las cuentas actuales); mantener la exigencia de teléfono verificado (contradice la instrucción del propietario); enviar por el otro canal cuando falta el elegido (contradice `DEC-092`).
- **Documentos afectados:** `docs/00-control/{matriz-trazabilidad.md,historial-cambios.md}`, `docs/01-producto/reglas-negocio.md`, `docs/02-requisitos/historias-usuario.md` (`HU-008`).
- **Fuente:** instrucción explícita del propietario del 2026-09-19 («si se manda el correo se mande al correo, si se manda el número se mande al número, no que espere que tenga los dos») y su confirmación de la opción «Correo basta» (`AskUserQuestion`), issue [#278](https://github.com/bcaceres19/barberia/issues/278).

### DEC-095 · Avisos emergentes tipo acordeón compartidos por todas las pantallas

- **Fecha:** 2026-09-19.
- **Decisión:** el resultado de una acción se notifica además con un aviso emergente (toast) de una cola única compartida (`shared/model/toastStore`, API `useToast`). El tiempo en pantalla depende de la variante: confirmación 5 s, información 6 s, advertencia 8 s y error 12 s. La cuenta atrás se detiene mientras el aviso está abierto, con el cursor encima o con foco de teclado. Se muestran como máximo 4 a la vez (al superarlo se descarta el más antiguo) y funcionan como acordeón: abrir uno cierra el que estuviera abierto. `role="alert"` solo para errores y `role="status"` para el resto; nunca roba el foco y Escape descarta el aviso con foco. Solo hay una región a la vez: el cascarón privado aloja la suya entre la cabecera y la navegación (`meta.toastHost: 'shell'`) y `App.vue` monta la de ventana en las pantallas sin cascarón (acceso, recuperación, reserva pública).
- **Alcance de la sustitución:** no sustituye ninguna decisión. Conserva `estandar-diseno-visual.md` §6.3: un éxito importante permanece dentro de la tarea y el aviso solo lo acompaña; los errores de formulario o diálogo siguen en línea. Un fallo de conexión o un error inesperado que no pertenece a ningún campo se avisa como toast con la acción de reintentar y una referencia segura (`request_id`), nunca un dato personal.
- **Responsable:** propietario del proyecto.
- **Motivo:** las alertas fijas dispersas ocupaban espacio, quedaban obsoletas y no distinguían un mensaje que necesita lectura (un error) de una confirmación breve. Una cola única evita que cada pantalla reimplemente tiempos, pausas y accesibilidad, y mantiene el resultado persistente como fuente de verdad.
- **Riesgo residual aceptado:** un aviso efímero puede pasar inadvertido; por eso nunca es el único registro de un éxito ni de un error de formulario, y la pausa por foco, cursor y apertura cubre WCAG 2.2.1.
- **Alternativas descartadas:** una dependencia de toasts de terceros (una cola de este tamaño no la justifica); mostrar todos los resultados solo en línea (no cubre fallos de conexión sin campo asociado); tiempo único para todas las variantes (un error no se alcanza a leer en 5 s).
- **Documentos afectados:** `docs/03-desarrollo/estandar-diseno-visual.md` (§6.8), `docs/00-control/historial-cambios.md`.
- **Fuente:** diseño e implementación de avisos emergentes del propietario, entregados en el issue [#280](https://github.com/bcaceres19/barberia/issues/280); el propietario los confirma al aprobar su PR. El número `DEC-094` que llevaba el borrador quedó ocupado por #279 y se renumeró a `DEC-095`.

### DEC-096 · Proveedor OTP WhatsApp reversible durante la indisponibilidad de Meta

- **Fecha:** 2026-09-19.
- **Decisión:** Meta WhatsApp Cloud API se conserva como proveedor `meta`. Mientras su verificación esté indisponible, el despliegue puede seleccionar `OTP_PROVIDER=twilio` para usar Twilio Verify mediante un canal configurado; volver a `OTP_PROVIDER=meta` solo requiere reiniciar con esa configuración. Los controladores, contratos HTTP y casos de uso no conocen el proveedor.
- **Persistencia y seguridad:** con Meta la aplicación mantiene el HMAC del código generado localmente. Con Twilio, Verify es la autoridad del OTP: la aplicación persiste solo una marca opaca para conservar cooldown, límite, expiración y el paso posterior del dominio, nunca el código ni su hash. El estado `approved` es el único que autoriza consumir el reto local. La vigencia local no puede superar los 600 s de Verify cuando se selecciona Twilio. Los errores del proveedor se traducen detrás del puerto OTP y no exponen payloads, secretos ni destinos completos.
- **Alcance:** aplica al reto de acceso de `HU-007` y al canal WhatsApp de recuperación de `HU-008`; correo, rutas existentes, OpenAPI y Meta no se sustituyen. La configuración Twilio solo es obligatoria cuando se selecciona ese proveedor.
- **Infraestructura separada:** este cambio de código no crea recursos. Twilio Verify WhatsApp requiere un Verify Service y un sender propio asociado a WABA/Messaging Service; las políticas y verificación de Meta siguen siendo requisitos operativos para producción.
- **Documentos afectados:** `apps/api`, `database/migrations`, `database/modelo-fisico-referencia.sql`, `apps/api/README.md`; issue [#282](https://github.com/bcaceres19/barberia/issues/282).
- **Fuente:** instrucción explícita del propietario el 2026-09-19 para habilitar Twilio temporalmente, de forma reversible, sin acoplar autenticación a un proveedor.

### DEC-097 · Sandbox de Twilio aislado para desarrollo de OTP WhatsApp

- **Fecha:** 2026-09-19.
- **Decisión:** se habilita `OTP_PROVIDER=twilio_sandbox` exclusivamente en `local` y `test`, como adaptación de Twilio Programmable Messaging para probar entrega WhatsApp mientras Meta/WABA no está disponible. No sustituye `OTP_PROVIDER=twilio`, que conserva Twilio Verify para un sender productivo, ni altera `meta`.
- **Seguridad y límites:** Sandbox vuelve al modelo local de generación y HMAC del OTP; Twilio solo transmite el mensaje y nunca es autoridad de verificación. Requiere `TWILIO_ACCOUNT_SID`, credenciales Twilio y el sender E.164 `TWILIO_WHATSAPP_SANDBOX_FROM`; la configuración falla fuera de local/test. Solo puede enviar texto libre durante la ventana de servicio de 24 horas que inicia `join`; fuera de ella el Sandbox exige una plantilla preaprobada y no admite plantillas propias. La activación, unión de destinatarios y renovación de sesión son manuales en la consola Twilio; no se automatizan ni se usan para producción.
- **Alcance:** conserva los mismos handlers, contratos, límites y puertos de HU-007/HU-008. No crea migraciones, recursos productivos, endpoints ni dependencias.
- **Fuente:** petición explícita del propietario tras confirmar que Meta Business permanece rechazado; issue [#282](https://github.com/bcaceres19/barberia/issues/282).

### DEC-098 · SMS de Twilio Verify como contingencia temporal de OTP

- **Fecha:** 2026-09-19.
- **Decisión:** `OTP_PROVIDER=twilio` usa `TWILIO_VERIFY_CHANNEL=sms` por defecto mientras Meta/WABA no está disponible. El mismo Verify Service crea y valida los códigos; no se genera ni persiste un OTP de aplicación. `TWILIO_VERIFY_CHANNEL=whatsapp` conserva el retorno reversible cuando exista el sender propio requerido por Meta.
- **Compatibilidad:** handlers, rutas, cuerpos JSON y el literal `channel: "whatsapp"` del contrato de recuperación no cambian. La interfaz presenta el destino como teléfono, sin exponer el proveedor o canal de infraestructura, para no prometer WhatsApp cuando se configura SMS.
- **Seguridad y alcance:** aplica únicamente a HU-007 y HU-008 dentro de DEC-096. Conserva vigencia máxima de 600 s, límites existentes y la validación explícita de `approved`; no modifica notificaciones de citas ni crea recursos externos.
- **Fuente:** instrucción explícita del propietario el 2026-09-19 para usar SMS temporalmente mientras evalúa WhatsApp; issue [#282](https://github.com/bcaceres19/barberia/issues/282).

### DEC-099 · Publicación unidireccional NAVA → Google Calendar por barbero

- **Fecha:** 2026-09-26.
- **Decisión:** se admite una integración con Google Calendar cuyo dueño es un barbero, **solo en la dirección NAVA → Google**: cada barbero conecta su propia cuenta y su propio calendario, y la conexión pertenece a `(barbería, barbero)`. Lo que el barbero o cualquier persona cambie, mueva o elimine en Google Calendar **no** afecta a NAVA: no reprograma ni cancela citas, no crea bloqueos y no altera la disponibilidad. Las citas solo se cancelan y reprograman desde la app. NAVA comprueba de forma periódica únicamente que los eventos que publicó sigan existiendo, para restaurarlos si se borraron por error (`DEC-101`); ese chequeo nunca modifica el dominio. Quedan fuera calendarios de clientes, de administradores, un calendario global, de recursos o sucursales, y cualquier otro proveedor.
- **Principio de autoridad:** NAVA es la única autoridad sobre las citas y sus reglas (`RN-CON-03`, `DEC-073`, `DEC-076`, máquina de estados). Google Calendar es una vista de solo lectura funcional de la agenda del barbero.
- **Responsable:** propietario del proyecto.
- **Motivo:** el barbero quiere ver sus citas en el calendario que ya usa y recibir su aviso de proximidad, sin que un cambio hecho en otra herramienta rompa las reglas de la agenda ni exija resolver conflictos entre dos fuentes.
- **Alternativas descartadas:** sincronización bidireccional con webhooks, `syncToken` y eventos externos como bloqueos (propuesta inicial del 2026-09-26, descartada por el propietario el mismo día por complejidad y riesgo); un calendario único para toda la barbería, por mezclar agendas personales y violar `DEC-019`.
- **Efecto sobre el alcance:** `alcance-mvp.md` sigue excluyendo la lectura desde calendarios externos y la integración bidireccional; esta decisión solo admite publicar.
- **Documentos afectados:** `01-producto/alcance-mvp.md`, `historial-cambios.md`, `dudas-pendientes.md`, catálogo de prompts. Cada capacidad se entrega en su propio issue y PR según [PROMPT-ORCH-GCAL-BARBERO-v1](../10-backlog/prompts/orchestration/google-calendar-barbero.md); las `HU-*` se redactan en la primera entrega de cada capacidad, sin reservar números aquí.
- **Fuente:** instrucción del propietario el 2026-09-26 («app → google calendar y no a viceversa»); issue documental [#284](https://github.com/bcaceres19/barberia/issues/284).

### DEC-100 · Vínculo explícito y opcional entre `barber` y `staff_user`

- **Fecha:** 2026-09-26.
- **Decisión:** se aprueba la excepción que `DEC-047` dejaba a «una historia futura»: `barber.staff_user_id` nullable, clave foránea compuesta `(barbershop_id, staff_user_id)` hacia `staff_user`, y unicidad parcial por `(barbershop_id, staff_user_id)` cuando no es nulo, de modo que cada usuario es como máximo un barbero y cada barbero tiene como máximo un usuario.
- **Quién lo asigna (resuelve `DP-INT-01`):** el propio usuario autenticado selecciona cuál barbero es él. La operación actúa siempre sobre el principal autenticado y nunca acepta un `staff_user_id` en la solicitud, así que nadie asigna el vínculo de otra persona. Un barbero ya vinculado a otro usuario no puede tomarse (conflicto). El usuario puede cambiar o quitar su propio vínculo. Nunca se infiere por nombre, correo ni teléfono, y `staff_user.id` nunca se asume igual a `barber.id`. El dueño que no atiende clientes puede no tener barbero vinculado.
- **Efecto sobre Google Calendar:** un usuario sin barbero vinculado no puede conectar Google Calendar. Si el usuario cambia o quita su vínculo, se desconecta la conexión de Google del barbero anterior (revocando y borrando sus credenciales), para que quien tome ese barbero después no herede una cuenta ajena. Esta consecuencia es una propuesta mía, sujeta a la revisión del propietario.
- **Responsable:** propietario del proyecto.
- **Motivo:** sin este vínculo el sistema no puede saber qué barbero corresponde a la persona autenticada, y la conexión OAuth debe pertenecer al barbero, no a un usuario arbitrario.
- **Alternativas descartadas:** tabla de vínculo aparte, porque hoy nada justifica varias cuentas por barbero; que cualquier usuario asigne el vínculo de otro, o que solo lo haga un propietario, porque el propietario dijo que cada usuario selecciona a su barbero.
- **Documentos afectados:** `database/modelo-fisico-referencia.sql`, diccionario y diagrama de datos, OpenAPI de barberos. Se entrega como prerrequisito propio ([PROMPT-FEAT-GCAL-01-v1](../10-backlog/prompts/hu/gcal-01-vinculo-barbero-usuario.md)).
- **Fuente:** aprobación explícita del propietario el 2026-09-26 («el usuario selecciona al barbero»); issue [#284](https://github.com/bcaceres19/barberia/issues/284).

### DEC-101 · Semántica de la publicación de citas y bloqueos en Google Calendar

- **Fecha:** 2026-09-26.
- **Decisión:**
  1. **Vínculo.** Cada cita o bloqueo publicado se asocia con su evento mediante `google_calendar_event_link` (`resource_type` `appointment` o `time_block`, id del evento y `etag`). Nunca se identifica por título ni por hora. El evento lleva propiedades extendidas privadas con identificadores opacos (`navaResourceType`, `navaResourceId`, `navaConnectionId`).
  2. **Citas.** Al confirmarse una cita se crea el evento; reprogramar actualiza el mismo evento; `cancelled_by_customer` y `cancelled_by_barber` **eliminan** el evento; `completed` y `no_show` lo conservan. Los estados de NAVA no cambian por ningún dato de Google.
  3. **Bloqueos.** Se publican los bloqueos puntuales de origen `manual` de tipo `break`, `lunch`, `unavailable`, `day_off`, `vacation` y `emergency`; `holiday` (calendario automático) no. Las series de bloqueo de NAVA quedan fuera de la primera fase (`DP-INT-02`).
  4. **Cambios hechos en Google.** NAVA no los lee para cambiar su dominio ni reacciona a ellos. Si el evento fue modificado en Google, la siguiente actualización originada en NAVA lo restablece al estado canónico. Si fue **eliminado** en Google, un trabajo periódico del worker lo detecta y lo vuelve a crear en pocos minutos, porque un borrado accidental no debe dejar la cita sin publicar; solo se restauran las citas confirmadas futuras y los bloqueos vigentes futuros. La única forma de dejar de publicar una cita es cancelarla desde la app, que elimina el evento y cierra su vínculo, y una ausencia (`404`/`410`) durante esa cancelación se trata como éxito. La cadencia del chequeo se fija y documenta en la entrega; el chequeo consulta solo los eventos publicados por NAVA (por sus propiedades extendidas privadas) y nunca usa `syncToken` ni webhook.
  5. **Recordatorio.** La anticipación del aviso la define el barbero en la sección de configuración de NAVA, como un número entero de minutos (por ejemplo 30) entre 0 y 40320 (límite de Google), guardado en su conexión. Con valor, cada evento se publica con un único recordatorio emergente (`reminders.useDefault = false` y ese `override`); sin valor, usa los recordatorios predeterminados del calendario del barbero. Cambiarlo actualiza los eventos futuros ya publicados. NAVA no envía un aviso propio.
  6. **Privacidad.** El título del evento de una cita es `Nombre del cliente — Servicio`; la descripción contiene solo el servicio y el estado. Nunca incluye teléfono, correo, notas privadas, tokens ni identificadores internos fuera de las propiedades extendidas privadas.
  7. **Zona horaria.** Todo se convierte con la zona de `barbershop` (`shops.Timezone`), nunca una zona fija; el evento representa el mismo instante.
  8. **Ventana inicial.** Al conectar, se publican las citas y bloqueos futuros desde el instante actual hasta 6 meses hacia adelante; lo pasado no se publica. Después, todo cambio se publica según ocurre.
  9. **Idempotencia.** Reintentar un trabajo o repetir un cambio no crea un segundo evento: el vínculo persistido decide entre crear y actualizar.
  10. **Estados de la conexión.** `connected`, `reauth_required`, `error` y `disconnected`. Un error permanente (token revocado, permiso denegado, calendario eliminado) cambia el estado y nunca detiene ni deshace las citas de NAVA.
  11. **Desconexión.** Revoca el token, elimina las credenciales almacenadas y detiene los trabajos pendientes de esa conexión. No borra citas ni elimina los eventos ya creados en Google (`DP-INT-03`, resuelta el 2026-09-26).
- **Responsable:** propietario del proyecto.
- **Motivo:** conserva la autoridad de `booking` y `schedule`, evita un segundo origen de verdad y da una regla única por estado.
- **Alternativas descartadas:** marcar el evento como cancelado en Google en vez de eliminarlo, por dejar ruido en el calendario; leer cambios de Google (`DEC-099`).
- **Documentos afectados:** `database/modelo-fisico-referencia.sql`, contrato OpenAPI de la conexión, diccionario de datos.
- **Fuente:** instrucciones del propietario del 2026-09-26 (semántica funcional, respuestas sobre título, cancelación, ventana y desconexión, y limitación a NAVA → Google); issue [#284](https://github.com/bcaceres19/barberia/issues/284).

### DEC-102 · Cola propia y protección de tokens de Google Calendar

- **Fecha:** 2026-09-26.
- **Decisión:**
  1. **Outbox propio.** Como la maquinaria de notificaciones de B5 no existe, el módulo de integración tendrá su tabla `google_calendar_sync_job`, escrita en la misma transacción que el cambio de negocio mediante un puerto que definen `booking` y `schedule` (que no importan Google). La reclamación usa `UPDATE … SET status='processing', claim_token, lease_expires_at … RETURNING`, envío fuera de transacción y finalización por CAS sobre `claim_token`, según `DDL-CON-01`. El worker (`barberia_worker`, sin contexto de tenant, solo funciones de claim/finalización) la consume. Los errores temporales (`429`, `5xx`, tiempo de espera) usan backoff exponencial con tope y número máximo de intentos. Ninguna llamada a Google ocurre dentro de una transacción de PostgreSQL de negocio; una caída de Google nunca impide crear una cita.
  2. **Tokens.** El refresh token se persiste cifrado con AES-256-GCM en la aplicación, con clave de 32 bytes desde el secreto `GOOGLE_CALENDAR_TOKEN_ENCRYPTION_KEY` e identificador de clave en la fila para rotación. El access token no se persiste. Ningún token, secreto ni código OAuth aparece en logs, respuestas HTTP, frontend ni Git.
  3. **OAuth.** Flujo de código de autorización con PKCE, `access_type=offline` y `prompt=consent`. El `state` es de un solo uso, de vigencia corta y queda ligado a barbería, barbero y sesión. Alcances mínimos: `https://www.googleapis.com/auth/calendar.events` más `openid email` para mostrar la cuenta conectada. `calendar.events` es un alcance sensible: la verificación de Google para producción se documenta en la entrega.
  4. **Sin webhook ni `watch`.** Al no importar cambios de Google (`DEC-099`), no hay endpoint público de notificaciones, canales, `syncToken` ni renovación de canales. El worker suma un trabajo periódico de restauración de eventos borrados (`DEC-101`), que solo lista los eventos publicados por NAVA.
  5. **Dependencias y configuración.** Se admite una dependencia oficial de Google (`golang.org/x/oauth2` y `google.golang.org/api/calendar/v3`), aislada en el módulo, con justificación y `govulncheck` en la entrega. Las variables son `GOOGLE_CALENDAR_CLIENT_ID`, `GOOGLE_CALENDAR_CLIENT_SECRET`, `GOOGLE_CALENDAR_REDIRECT_URI` y `GOOGLE_CALENDAR_TOKEN_ENCRYPTION_KEY`; sin ellas la integración queda desactivada y el resto del producto funciona.
  6. **Ambientes.** La URI de redirección de OAuth exige HTTPS fuera de local; `http://localhost` se admite solo en local. La entrega documenta ambas.
- **Responsable:** propietario del proyecto.
- **Motivo:** el estándar no fija cifrado de credenciales de terceros ni cola de trabajos, y la integración los necesita antes de escribir código.
- **Alternativas descartadas:** esperar a B5 y compartir su cola, por bloquear la funcionalidad; llamar a Google dentro de la transacción de negocio, por acoplar la reserva al servicio externo.
- **Documentos afectados:** `05-backend/estandar-base-datos.md` (referencia), `04-arquitectura/stack-despliegue-operacion.md`, README de `apps/api`, migraciones y contrato en cada entrega.
- **Fuente:** instrucciones del propietario del 2026-09-26 (outbox, tokens y worker); aprobación explícita del cifrado AES-256-GCM con clave por variable de entorno el 2026-09-26; issue [#284](https://github.com/bcaceres19/barberia/issues/284).

### DEC-103 · Excepción puntual a la paginación por cursor: `GET /private/services` pasa a paginación por número de página, con búsqueda por nombre

- **Fecha:** 2026-09-28.
- **Contexto:** el panel "Servicios" mostraba el catálogo con una lista de scroll infinito ("Cargar más" sobre paginación por cursor, CA-022-01). El propietario pidió una tabla con paginador y buscador; un paginador numerado con total exacto (`1, 2, 3…`, "34 servicios") exige conocer el total de filas y saltar a una posición arbitraria, algo que la paginación por cursor documentada en `docs/06-api/estandar-openapi.md` §6.10 ("listas potencialmente crecientes usan paginación por cursor") no ofrece sin una consulta `COUNT` aparte de todos modos.
- **Decisión:** `GET /private/services` (`operationId: listServices`, HU-022) reemplaza `cursor`/`limit` por `page`/`pageSize` (mismos valores 1/20/50 de mínimo/por defecto/máximo, solo reinterpretados como número de página en vez de cursor) y agrega `search` (filtro opcional, coincidencia parcial insensible a mayúsculas, solo sobre `name`). La respuesta reemplaza `nextCursor` por `page`, `pageSize`, `total` y `totalPages` (con piso de 1). Esta es una excepción PUNTUAL y documentada a la regla general de §6.10, acotada a este único endpoint: el resto de listados del contrato (`BarberListResponse`, `AssignmentListResponse` de HU-023, `ScheduleListResponse`, etc.) conserva paginación por cursor sin cambios, porque ninguno de ellos tiene hoy un requisito de paginador numerado que lo justifique.
- **Alcance excluido:** no se crea un índice de búsqueda de texto (`pg_trgm`/GIN) para acelerar `ILIKE`: el catálogo de servicios de una barbería es una colección acotada por el uso real (decenas, no miles, de filas), y añadir ese índice sin evidencia de un problema de rendimiento real sería una abstracción sin necesidad demostrada (`AGENTS.md`, "Calidad"). Se revisará si el volumen real de alguna barbería lo justifica.
- **Compatibilidad:** cambio incompatible del contrato (`api/openapi/CHANGELOG.md` 0.24.0): ambos lados (backend y `apps/web`) se migran en el mismo cambio; consumidor único, sin período de transición necesario.
- **Fuente:** instrucción explícita del propietario el 2026-09-28 ("colócalos en una tabla con paginador y con su buscador... ajusta lo que debas ajustar en el back y front"), con confirmación explícita entre paginador Anterior/Siguiente (cumple §6.10 sin excepción) y paginador numerado con total (requiere esta excepción) — el propietario eligió el paginador numerado.

### DEC-104 · Fotografía opcional del barbero: almacenamiento en PostgreSQL y primer contenido binario del contrato

- **Fecha:** 2026-09-30.
- **Contexto:** al rediseñar la pantalla «Barberos» según el lenguaje ya aprobado en Agenda y Servicios, el propietario pidió «un espacio para la imagen del barbero» y ajustar back y front si hacía falta. `docs/03-desarrollo/especificacion-frontend-nava.md` §14 dejaba la «carga de fotos de barberos» como pendiente deliberado (sin fuente ni política), `docs/06-api/estandar-openapi.md` §10 exige un caso aprobado antes de añadir contenido binario y `BarberAvatar` ya tenía el prop `photoUrl` sin consumidor.
- **Decisión:** cada barbero puede tener **una** fotografía opcional. Se guarda en la tabla nueva `barber_photo` (`bytea`, una fila por barbero con clave `(barbershop_id, barber_id)`, FK compuesta `ON DELETE RESTRICT`, RLS con las mismas políticas por tenant que `barber` más `DELETE`), no en `barber`: el listado del equipo solo lee `updated_at` (LEFT JOIN) y nunca los bytes. El contrato añade `PUT`/`GET`/`DELETE /private/barbers/{barberId}/photo` (cuerpo binario `image/jpeg` o `image/png`, sin base64 ni JSON) y `photoUpdatedAt` (nullable, obligatorio) a `BarberResponse`; `photoUpdatedAt` es la versión de la imagen (URL versionada y `ETag`, `304` a `If-None-Match`, `Cache-Control: private`). El `PUT` reemplaza y es idempotente por naturaleza, así que no lleva `Idempotency-Key`; el `DELETE` es idempotente y **borra** la fila (retirar la imagen de una persona no es ocultarla).
- **Política de la imagen:** el navegador recorta al centro en cuadrado, reduce a 512 px y re-codifica a JPEG (sin EXIF). El servidor valida por los **bytes** (`image.DecodeConfig`, no por `Content-Type` ni extensión) formato JPEG/PNG coincidente con la cabecera, máximo 512 KiB y entre 64 y 1024 px por lado; el esquema repite tipo y tamaño en `CHECK`. No se conserva nombre de archivo ni metadatos. Es un dato personal (imagen de una persona) con retención mientras el barbero la conserve.
- **Alcance excluido:** sin almacenamiento de objetos ni CDN (no existe hoy en la infraestructura y sería una dependencia sin necesidad demostrada; se revisará si el volumen real lo justifica); sin fotografía en las pantallas públicas de reserva (`HU-092` no cambia: exigiría un endpoint público y una decisión sobre exposición de imágenes de personas); sin recorte manual ni edición avanzada; sin foto de clientes ni de la barbería.
- **Compatibilidad:** cambio compatible del contrato (`api/openapi/CHANGELOG.md` 0.25.0): campo nuevo en respuestas y operaciones nuevas. El alta idempotente ya persistida antes de este cambio (vigencia 24 h) se repite sin `photoUpdatedAt`; el cliente lo trata como «sin fotografía».
- **Alternativas descartadas:** columna `bytea` en `barber` (arrastra los bytes en cada lectura del equipo); almacenamiento de objetos externo (dependencia e infraestructura nuevas); base64 dentro de JSON (prohibido por §10 y 33 % más pesado); confiar en el `Content-Type` (permitiría almacenar y servir cualquier archivo con apariencia de imagen); `multipart/form-data` (más superficie que un cuerpo binario único sin campos adicionales).
- **Fuente:** instrucción explícita del propietario el 2026-09-30 («un espacio para la imagen del barbero, ajusta ya sea back y front si es necesario para este tema»).

### DEC-105 · Bloqueos integrados en Barberos

- **Fecha:** 2026-09-30.
- **Decisión:** consultar, crear y retirar los bloqueos puntuales y las series semanales existentes desde una ventana modal amplia, abierta con «Bloquear» en cada barbero de «Barberos», sin expandir sus filas. La ventana presenta las listas en dos columnas en escritorio y apiladas en móvil, con scroll interno, cierre por Escape y restauración de foco. Los formularios conservan su diálogo propio sobre la consulta. Esta composición responde al ajuste posterior solicitado por el propietario el mismo día. «Bloqueos» deja de ser un destino independiente; su URL antigua redirige a Barberos. El rediseño es de identidad guiada NAVA / Tailored Grid, sin mockup exacto asignado.
- **Invariantes:** se mantienen HU-042, RN-BLQ-01/03/04, RN-TEN-01, RN-IDE-01 y la zona horaria de la barbería. No cambia la disponibilidad, no se cancelan ni reprograman turnos automáticamente y los seguimientos de #100 continúan separados.
- **Fuente:** instrucción explícita del propietario el 2026-09-30; issue [#286](https://github.com/bcaceres19/barberia/issues/286).
- **Nota de conciliación:** al actualizar main, las decisiones locales de paginación de Servicios y fotografía ocupaban DEC-099/100 ya publicados para Google Calendar en #285. Se preserva el contenido de ambas, reasignando las locales a DEC-103/104 y actualizando sus referencias; no se cambia su comportamiento. La migración aplicada de fotografía conserva su referencia histórica a DEC-100 por inmutabilidad; hoy corresponde a DEC-104.

### DEC-106 · Arranque local con esquema y recorrido comprobados

- **Fecha:** 2026-10-02.
- **Decisión:** el propietario autoriza crear `local-app-startup` en `.agents/skills/`, con adaptador de Claude y descubrimiento en AGENTS.md, para peticiones de encender, reiniciar o comprobar la app local. El skill reutiliza Docker/configuración/procesos, separa roles API y migrador, verifica Atlas y comprueba el recorrido autenticado de agenda antes de declarar la app lista. No agrega DDL al inicio del API ni reemplaza estándares.
- **Responsable:** propietario del proyecto.
- **Motivo:** PostgreSQL y `/health/db` respondían, pero la agenda fallaba con 500 en barberos por un esquema desfasado. La mera conectividad no es un criterio suficiente de arranque.
- **Límites:** encender la app no autoriza rotaciones de secretos, reparaciones masivas de permisos, `migrate set`, cambios de funcionalidades ni despliegues. Si falta sesión, se declara pendiente la verificación funcional. Migraciones y reconciliaciones conservan DEC-036/040 y la autorización aplicable.
- **Fuente:** petición explícita del propietario del 2026-10-02 («Genera una skill, prompt o algo, para que al momento de prender la app no vuelva a pasar este tipo de temas»); issue [#287](https://github.com/bcaceres19/barberia/issues/287).

### DEC-107 · Paginador de Barberos adaptado al espacio disponible

- **Fecha:** 2026-10-02.
- **Decisión:** la tabla de Barberos sigue el paginador numerado de Servicios y solicita solo las filas que caben en el viewport; recalcula al redimensionar y muestra la última página tras un alta, respetando el orden por fecha de alta. `GET /private/barbers` añade `page`/`pageSize` como modo optativo, con total exacto, página efectiva y totalPages (piso 1), preservando `cursor`/`limit` y su respuesta para selectores existentes. No se mezclan modos. PageSize admite 1–50, por defecto 20; una página fuera de rango se ajusta a la última. Ambos modos mantienen orden `(created_at,id)` y RN-TEN-01. Total y filas se leen en una misma sentencia PostgreSQL para compartir snapshot.
- **Alcance:** excepción puntual a §6.10 OpenAPI para la tabla de Barberos; complementa DEC-103 sin sustituir el cursor de otros listados. Sin búsqueda nueva, migraciones, límite de equipo ni funciones nuevas.
- **Fuente:** petición del propietario del 2026-10-02 de paginador como Servicios y cantidad de filas que soporte la pantalla sin scroll; issue [#288](https://github.com/bcaceres19/barberia/issues/288).

### DEC-108 · El rechazo de credenciales en `/acceso` se avisa como aviso emergente

- **Fecha:** 2026-10-03.
- **Decisión:** cuando el servidor rechaza el inicio de sesión (credenciales inválidas o formato rechazado), `/acceso` lo comunica con un aviso emergente de error de la cola compartida (`DEC-095`), arriba a la derecha en escritorio y a todo el ancho superior en móvil, que se retira solo a los 12 s, en vez de una alerta fija roja bajo el formulario. El texto no cambia y sigue sin distinguir correo inexistente de contraseña incorrecta (`CA-005-02`, `CA-010-02`); el correo se conserva y la contraseña se limpia como hasta ahora. Los errores locales de campo (`Escribe tu contraseña`) y el resumen de campos siguen en línea.
- **Alcance de la sustitución:** acota `DEC-095` («los errores de formulario o diálogo siguen en línea») solo para el rechazo del servidor en el inicio de sesión. Recuperación de acceso no cambia: sus errores de código o política siguen en línea.
- **Responsable:** propietario del proyecto.
- **Motivo:** una alerta roja fija permanecía tras el error y competía con el formulario; el propietario ya había indicado que las notificaciones van en la esquina superior derecha y se retiran con el tiempo.
- **Riesgo residual aceptado:** un aviso efímero puede pasar inadvertido; el campo de contraseña vacío tras el rechazo y el `role="alert"` del aviso de error lo compensan, y la pausa por foco, cursor y apertura cubre WCAG 2.2.1.
- **Alternativas descartadas:** mantener la alerta en línea (contradice la instrucción); mostrar ambas (duplica el mensaje).
- **Documentos afectados:** `docs/03-desarrollo/estandar-diseno-visual.md` (§6.8), `apps/web/e2e/evidence/auth-fidelidad-desviaciones.md`.
- **Fuente:** instrucción explícita del propietario del 2026-10-03 («esa notificación roja… las notificaciones deben aparecer en la esquina superior derecha y que se vayan eliminando temporalmente»).

### DEC-109 · Horarios adopta el tablero semanal sobre tinta del panel

- **Fecha:** 2026-10-03.
- **Decisión:** `/panel/horarios` abandona la composición clara del atlas de #196 y adopta el lenguaje que ya comparten Agenda, Servicios y Barberos: superficie tinta, paneles hundidos, latón de marca, versalitas espaciadas, rombos de estado, campos reglados con línea base de latón y movimiento con `prefers-reduced-motion`. La semana se dibuja como un tablero: cada día lleva una pista horaria con una barra de latón por tramo sobre una regla común (07–21 h, que crece solo si un tramo se sale de ella), el marcador del día y la hora vigentes en la zona de la barbería, cifras de la semana derivadas de los tramos ya cargados y la lista real de tramos, con **Editar** y **Retirar**, bajo cada pista. Los controles nativos se sustituyen por los compartidos del panel: selector de barbero con retrato (`BaseSelect`), hora (`BaseTimePicker`), fecha (`BaseDatePicker`) y un selector de día de la semana propio; el diálogo muestra una vista previa del tramo junto a los ya existentes de ese día.
- **Composición sin desplazamiento (ajuste del mismo día):** en escritorio (≥ 1100 px) la pantalla cabe en la altura del panel: la cabecera reúne título, selector de barbero, cifras y «Agregar tramo»; el tablero ocupa la columna izquierda y a su derecha un acordeón con tres apartados —calendario de festivos, excepciones de jornada y próximos festivos— del que se abre uno solo a la vez (excepciones primero; la insignia de cada encabezado resume su estado o conteo). La lista del apartado abierto se desplaza dentro de su cuerpo si no cabe, nunca la página. En pantallas angostas el acordeón pasa bajo el tablero y los apartados crecen con su contenido. Cada encabezado es un botón con `aria-expanded` dentro de un `h2`, y un apartado cerrado no recibe foco ni se lee con lector de pantalla.
- **Alcance:** solo presentación e interacción. No cambia ningún contrato OpenAPI, regla de negocio, validación ni texto de error; el servidor sigue decidiendo solapes (`CA-040-04`) y fechas duplicadas (`CA-041-05`). Las cifras, «Hoy», «Con excepción» y la vista previa son derivados de datos ya cargados, no funciones nuevas ni almacenamiento nuevo.
- **Responsable:** propietario del proyecto.
- **Motivo:** la pantalla era la única del panel que seguía clara y descuadrada (tarjetas inferiores desalineadas, controles nativos blancos sobre un panel de tinta).
- **Riesgo residual aceptado:** la pista horaria es decorativa (`aria-hidden`); la información completa vive en la lista de tramos, que sigue siendo la fuente accesible. Un tramo que cruza medianoche se recorta al fin de la regla y se rotula `(+1)` en su hora de fin.
- **Alternativas descartadas:** conservar el atlas #196 (contradice la instrucción); mover tramos arrastrando sobre la pista (la guía prohíbe el arrastre para mover segmentos).
- **Documentos afectados:** `docs/03-desarrollo/estandar-diseno-visual.md` (referencias por familia), `docs/03-desarrollo/especificacion-frontend-nava.md` (§7.8), `apps/web/e2e/evidence/horarios/rediseno/`.
- **Fuente:** instrucción explícita del propietario del 2026-10-03 de ajustar Horarios al diseño global del panel, con animaciones, campos y botones refinados y colores como Agenda, Servicios y Barberos.

### DEC-110 · Marca, vocabulario y apariencia configurables de la barbería

- **Fecha:** 2026-10-03.
- **Decisión:** la pantalla de Configuración (`/panel/barberia`) pasa de un formulario de cuatro campos a un centro de ajustes sobre tinta (mismo lenguaje de Agenda, Servicios y Barberos) con dos clases de ajuste, y cada sección dice cuál es. **Este dispositivo** (modo, tamaño de texto, animaciones): se guardan en el navegador (`localStorage`), se aplican al instante y no pasan por el servidor. **Toda la barbería** (nombre, color de acento, vocabulario, zona horaria, contacto): se editan en un borrador y solo un guardado confirmado por el servidor los aplica.
  1. **Modo.** Tinta (el panel oscuro de siempre, valor inicial), Marfil (claro, con los anclajes marfil/tinta/latón oscuro de `estandar-diseno-visual.md` §4.1) y Automático (sigue `prefers-color-scheme` en vivo). Solo vale dentro del cascarón privado: el acceso, la recuperación y la reserva pública no cambian.
  2. **Tamaño de texto.** Pequeño 90 %, Normal 100 %, Grande 112,5 % y Muy grande 125 %, escalando texto y geometría con `zoom` sobre `<html>` (el CSS del producto declara píxeles; `rem` no alcanzaría). El factor se limita para que el ancho efectivo nunca baje de 320 px y `--viewport-height` compensa `100dvh`.
  3. **Animaciones.** «Reducir animaciones» se suma a `prefers-reduced-motion`: una persona puede pedir menos movimiento aunque su sistema no lo pida, nunca más del que su sistema permite.
  4. **Color de acento.** Lista cerrada de seis claves (`brass`, `emerald`, `sapphire`, `ruby`, `amethyst`, `copper`), cada una con un valor para Tinta y otro para Marfil cuyo contraste con la superficie del modo cumple WCAG 2.2 AA como texto y como relleno de botón con texto de la superficie. Una prueba (`brandPalette.test.ts`) mide cada par; añadir un acento exige migración, contrato y esa medición. El latón sobre Tinta es el valor de `tokens.css`: una barbería que no cambia nada ve exactamente lo de siempre. Los colores semánticos de estado, el wordmark `NAVA` y la tipografía no cambian.
  5. **Vocabulario.** La barbería elige la palabra con la que llama a su negocio (por defecto «barbería») y a quien atiende los turnos (por defecto «barbero»/«barberos»), cada una con su género gramatical para concordar artículos y participios («esta estilista», «estilistas registradas»). Es presentación pura: `turno`, los estados autorizados, la tabla `barber` y los contratos de la API no se renombran. Cubre navegación, Barberos, Agenda, Nuevo turno, Detalle del turno, Servicios, Servicios por barbero y Configuración; Horarios y las pantallas públicas conservan el vocabulario inicial hasta su propio cambio. `DEC-119` lo extiende a Horarios y al resto del panel (#308) y, después, a las pantallas públicas (#309).
  6. **Zona horaria.** Selector con buscador sobre el catálogo IANA del navegador, reloj en vivo de la zona elegida y atajo a la zona del dispositivo que nunca se aplica solo; el servidor sigue confirmando contra `pg_timezone_names` (`CA-020-03`, `RN-DIS-07`).
- **Contrato y datos:** recurso nuevo `GET`/`PATCH /private/settings/brand` (seis campos siempre presentes; cuerpo cerrado; `422` sin escritura ante un acento fuera de lista o un término inválido) y seis columnas con `CHECK` en `barbershop` (`brand_accent`, `business_term`, `business_term_gender`, `professional_term`, `professional_term_plural`, `professional_term_gender`; migración `20261003120000_add_barbershop_brand.sql`, defaults idénticos a la interfaz anterior). El contrato de `HU-020` conserva exactamente sus cuatro campos (`CA-020-07`). Sin versión de concurrencia: son preferencias de presentación y la última escritura gana. Escribir toca `updated_at` y, como con `HU-020`, invalida el `If-Match` de una política de reserva leída antes (`HU-093`).
- **Guardado:** acota `especificacion-frontend-nava.md` §7.6/§7.9 («una sección guarda por separado»): datos básicos y marca son dos recursos distintos pero comparten una barra de guardado que envía solo los que cambiaron y confirma cada uno por separado (si uno falla, el otro queda guardado y lo escrito se conserva). Reglas de reserva pública conserva su pantalla y su guardado propios (`HU-093`).
- **Alcance de la sustitución:** amplía `DEC-039` (tema claro único, «construir modo oscuro y múltiples temas durante el MVP» descartado, sin sustitución de colores por barbería) y `estandar-diseno-visual.md` §3.2 («no personalización por tenant») y §4.3 («no se ofrece personalización de color por barbería»), únicamente en lo descrito aquí. Sigue prohibido un color libre, un logotipo propio, personalizar la reserva pública o los correos, y alterar los colores semánticos. Sustituye la fidelidad de `configuracion-barberia` a su mockup (`#170`, `#193`): la pantalla adopta el lenguaje del panel y su evidencia pasa a `apps/web/e2e/evidence/configuracion/rediseno/`.
- **Responsable:** propietario del proyecto.
- **Motivo:** el propietario pidió alinear Configuración con Agenda, Servicios y Barberos y ampliarla con las opciones de apariencia y marca que la app puede tener (modo claro/oscuro, aumento de letra, color del latón, cambiar «barbería» por la palabra que se quiera, zona horaria).
- **Riesgo residual aceptado:** (a) `zoom` no es estándar en todos los navegadores antiguos (Firefox lo soporta desde la 126); donde no exista, la preferencia no tiene efecto y la interfaz queda a 100 %. (b) Con «Muy grande» el dock de siete destinos recorta etiquetas largas con puntos suspensivos. (c) Un acento personalizado puede parecerse a un estado semántico (el rubí al peligro): el estado nunca depende solo del color. (d) La marca viaja con una solicitud adicional al autenticarse y una copia local evita el parpadeo; si el servidor no responde, el panel conserva la copia o los valores iniciales.
- **Alternativas descartadas:** color libre por barbería (rompe el contraste comprobado y la identidad); guardar las preferencias de pantalla en el servidor por usuario (no hay HU de preferencias por usuario y el tamaño de letra es del aparato); ampliar `GET`/`PATCH /private/settings/barbershop` con los campos nuevos (viola `CA-020-07`); renombrar `barber` o los contratos para reflejar el vocabulario (innecesario y rompería clientes); derivar el plural y el género de la palabra (el español no es regular).
- **Documentos afectados:** `docs/03-desarrollo/estandar-diseno-visual.md` (§3.2, §4.3, §6.9), `docs/03-desarrollo/especificacion-frontend-nava.md` (§4.1, §7.6, §9.4, §10.1), `docs/02-requisitos/historias-usuario.md` (`HU-020`, `HU-025`), `api/openapi/` (0.27.0), `database/README.md`, `database/tests/brand_marca_vocabulario.sql`, `docs/00-control/{matriz-trazabilidad,historial-cambios}.md`.
- **Fuente:** instrucción explícita del propietario del 2026-10-03 y issue [#292](https://github.com/bcaceres19/barberia/issues/292).

### DEC-111 · La reserva pública adopta el lienzo de tinta del panel con un cascarón persistente

- **Fecha:** 2026-10-03.
- **Decisión:** las cinco pantallas de `/reservar/:slug…` (entrada, servicio, barbero, horario y datos) abandonan el lienzo marfil con tarjetas blancas y adoptan el lenguaje que ya comparten Agenda, Servicios, Barberos y el acceso: tinta fija, latón de marca, serif editorial en títulos y nombres, versalitas espaciadas, filas de divisor fino, campos reglados y movimiento que respeta `prefers-reduced-motion`. Lo que la persona rellena (datos, resumen y confirmación) vive en un panel de tinta levantada con filete de latón (no en papel claro, que deslumbra sobre el lienzo oscuro); `BaseInput`, `BaseButton` y `BaseAlert` se reutilizan ajustando sus variables y colores dentro de la pantalla, sin una variante oscura nueva en el componente compartido.
- **Cascarón persistente:** las cinco pantallas pasan a ser hijas de una ruta `PublicBookingLayout`, con las mismas URL, nombres y parámetros. El cascarón conserva entre pasos el fondo animado (la mesa de patronaje del acceso recompuesta para una columna de lectura; un foco de latón sigue al cursor solo con puntero fino), la cabecera, el enlace de retorno al paso anterior, el progreso («Paso N de 4», rombos que se rellenan y una regla de latón que se dibuja) y la transición de paso (adelante entra desde la derecha, atrás desde la izquierda). Firma al pie «Reservas con NAVA» (`especificacion-frontend-nava.md` §2.1). La página de datos avisa al cascarón, mediante un `provide`/`inject` acotado a este árbol y no estado global, cuando el servidor confirma el turno: el progreso se completa y el retorno deja de ofrecerse.
- **Interacción nueva, sin función nueva:** selección con rombo que se rellena y dibuja su check, filete lateral y un destello; barra de acción fija al pie con la selección y «Continuar»; tira de días con disponibilidad sobre «Anterior/Siguiente» (los mismos datos ya recibidos, sin consulta nueva); ficha de papel de la franja elegida; sello animado de confirmación; reloj de la barbería que avanza cada 15 s; control segmentado para «¿Para quién es el turno?» sobre los mismos `<input type="radio">`.
- **Modo de conformidad:** identidad guiada (sin mockup exacto asignado, `DEC-078`/`DEC-080`).
- **Alcance:** solo presentación e interacción. No cambia ningún contrato OpenAPI, regla de negocio (`RN-DIS-03`: elegir una franja no reserva; `RN-DIS-07`: horas siempre en la zona de la barbería), validación, texto de error ni el orden del flujo. El cascarón no hereda `data-app-theme` ni el acento del panel (`DEC-110`): la reserva pública sigue sin personalización por barbería.
- **Accesibilidad:** la evidencia de la reserva dejó de desactivar `color-contrast` en axe-core, porque la pantalla ya no usa los pares claros que el estándar daba por medidos; se mide en cada pantalla y estado con movimiento reducido. Los estados de página sobre tinta usan los tintes levantados `--color-danger-on-strong` y `--color-warning-on-strong` en vez del rojo/ámbar de las superficies claras. El progreso expone `aria-current="step"` y anuncia los pasos completados; la oración completa de la franja elegida sigue anunciándose en un `role="status"` y la ficha visual se oculta a la tecnología de apoyo para no repetirla.
- **Responsable:** propietario del proyecto.
- **Motivo:** la reserva pública era la única superficie de cara a la persona usuaria que seguía clara, plana y sin movimiento, mientras el acceso y el panel ya compartían identidad de tinta y latón; el propietario pidió alinearla, animarla, refinar campos y botones y distribuir los objetos con criterio de experiencia.
- **Riesgo residual aceptado:** (a) el precio se sigue mostrando como lo entrega el contrato (`45000.00 COP`), sin formato local, para no alterar el texto verificado por las pruebas; (b) los pasos 1 a 4 no muestran el nombre de la barbería porque el contrato de servicios, barberos y disponibilidad no lo devuelve; solo la entrada lo muestra y el resto conserva la firma «Reservas con NAVA», brecha anterior a esta decisión; (c) el fondo animado y el foco de cursor consumen composición de GPU en equipos modestos y se apagan con `prefers-reduced-motion`; (d) `Mi turno` (`/mi-turno/:token`) conserva su aspecto anterior hasta que un issue propio lo autorice.
- **Alternativas descartadas:** una variante oscura de `BaseInput` (duplica el campo reglado y su contraste); conservar el marfil y solo añadir animaciones (deja la identidad dividida); consultar el perfil de la barbería en cada paso para mostrar su nombre (llamada extra o estado compartido sin HU); animar con una biblioteca (el movimiento es CSS sobre `transform`/`opacity`/`stroke-dashoffset`, sin dependencia nueva).
- **Documentos afectados:** `docs/03-desarrollo/estandar-diseno-visual.md` (§2 referencias, §6.10, §12), `docs/03-desarrollo/especificacion-frontend-nava.md` (§7.10), `docs/00-control/historial-cambios.md`, `apps/web/e2e/evidence/reserva-publica/rediseno/`. Issue pendiente de abrir antes de integrar.
- **Fuente:** instrucción explícita del propietario del 2026-10-03 de ajustar la pantalla de reservas públicas al diseño del resto del producto, con animaciones de eventos, campos, botones y colores como Agenda, Servicios y Barberos.

### DEC-112 · Escala tipográfica y movimiento compartidos del panel; Servicios por barbero sobre tinta

- **Fecha:** 2026-10-04.
- **Decisión:** Agenda, Servicios, Barberos, Servicios por barbero, Horarios y Configuración dejan de escribir tamaños de letra sueltos y consumen una sola escala por tokens. Piso legible: `--font-size-caption` (12 px), reservado a versalitas espaciadas y datos auxiliares; ningún texto del panel baja de ahí (antes había 8, 9, 10 y 11 px). Texto corrido: `--font-size-body-sm` (14), `--font-size-body` (16) y `--font-size-body-lg` (18). Títulos en serif: `--font-size-title-item` (21), `--font-size-title-section` (28) y `--font-size-title-page` (36, tope; antes 40). El vocabulario de movimiento (`styles/motion.css`: `nv-rise`, `nv-fade`, `nv-slide`, `nv-pop`, `nv-wipe`, `nv-lift` y los keyframes `nava-*`) se comparte entre pantallas con un índice `--i` para escalonar; todo es `transform`/`opacity`/`clip-path` y se apaga con `prefers-reduced-motion` y `data-motion="reduced"`.
- **Cambios de pantalla:** acciones de fila de 36 px con texto de 14 px (antes 24–34 px con 11–12 px); encabezados de tabla en versalitas de latón; anchura de lectura común de 920 px en las listas; botón de acción principal de 40 px; dock con filete de latón que se dibuja y transición de salida entre pantallas; Agenda con carril que se dibuja, fichas que crecen desde su hora, rombo con latido en «Ahora» y lista que se reescalona al cambiar de persona o de fecha; rótulo «Fecha» alineado a la izquierda en móvil. «Servicios por barbero» abandona la superficie clara y los controles nativos: ficha de la persona con medidor y regla de latón, filas de interruptor con casilla propia y estado como rombo con rótulo, esqueleto con brillo.
- **Modo de conformidad:** identidad guiada. Agenda conserva su estructura de fidelidad al mockup `panel-agenda-eventos` (#189): solo cambian tamaños de texto, ritmo y movimiento; las proporciones de las regiones principales no se tocan.
- **Alcance:** solo presentación e interacción. Sin cambio de contrato OpenAPI, regla de negocio, texto de error, selector accesible ni orden de flujo.
- **Responsable:** propietario del proyecto.
- **Motivo:** el propietario pidió refinar el diseño de estas pantallas, animar eventos, afinar campos, botones y colores como en las ya trabajadas, ordenar los objetos y corregir textos demasiado pequeños o demasiado grandes.
- **Riesgo residual aceptado:** (a) el precio sigue mostrándose como lo entrega el contrato (`45000.00 COP`), sin formato local, para no alterar el texto verificado por las pruebas; (b) con «Muy grande» el texto mínimo de 12 px escala hasta ~15 px y las filas de altura fija de Servicios y Barberos dependen del recorte con puntos suspensivos; (c) el monograma del retrato pequeño (`BarberAvatar`, 26–34 px) conserva un texto decorativo menor de 12 px porque es `aria-hidden` y el nombre visible lo acompaña.
- **Alternativas descartadas:** subir solo los tamaños sin tokens (volvería a divergir); una biblioteca de animación (el movimiento es CSS puro); rediseñar la estructura de Agenda (contradice la fidelidad ya medida en #189).
- **Documentos afectados:** `docs/03-desarrollo/estandar-diseno-visual.md` (§5.3, §6.11), `docs/00-control/historial-cambios.md`, `apps/web/src/styles/{tokens,motion}.css`. Issue pendiente de abrir antes de integrar.
- **Fuente:** instrucción explícita del propietario del 2026-10-04.

### DEC-113 · Estados vacíos del panel como escena animada

- **Fecha:** 2026-10-04.
- **Decisión:** los estados «sin turnos», «sin servicios», «sin barberos» y «sin resultado de búsqueda» dejan de ser un párrafo suelto o un divisor solitario y pasan a `EmptyScene` (`shared/ui`): ilustración de línea de latón que se dibuja sola, cuadrícula de puntos con brillo, rombos que flotan, filete superior que se traza, titular serif, ayuda y la acción real con un destello periódico. Cinco escenas: huecos de agenda sobre una regla de horas con el marcador «ahora», pila de fichas de catálogo, tres marcos de retrato con el central por llenar, dos marcos unidos por un vínculo y una búsqueda. Se aplica en Agenda (sin turnos y sin barberos), Servicios (vacío y búsqueda sin resultado), Barberos, Horarios y Servicios por barbero (sin barberos y sin servicios).
- **Fondo vivo:** el escenario del cascarón añade, detrás de todas las pantallas del panel, dos luces tenues que derivan y una cuadrícula de patronaje que avanza (`estandar-diseno-visual.md` §6.13); las pantallas pasan a raíz transparente.
- **Alcance:** solo presentación. Los titulares conservan su texto exacto; se añaden ayudas breves y, en Servicios, la acción «Agregar servicio» dentro del vacío. Ninguna ilustración usa tijeras, navajas ni postes (`estandar-diseno-visual.md` §3.1). La ilustración es `aria-hidden`; `prefers-reduced-motion` y `data-motion="reduced"` la dejan dibujada y quieta.
- **Responsable:** propietario del proyecto.
- **Motivo:** el propietario consideró los estados vacíos «muy simples» frente al resto de las pantallas.
- **Riesgo residual aceptado:** cuatro animaciones infinitas suaves (brillo, rombos, latido del «+», destello del botón) consumen composición mínima y se apagan con movimiento reducido.
- **Documentos afectados:** `docs/03-desarrollo/estandar-diseno-visual.md` (§6.12), `docs/00-control/historial-cambios.md`, `apps/web/src/shared/ui/EmptyScene.vue`.
- **Fuente:** instrucción explícita del propietario del 2026-10-04.

### DEC-114 · Retirar a cualquier barbero de un servicio, también al último (sustituye a `DEC-068`)

- **Fecha:** 2026-10-04.
- **Decisión:** la desasignación de un servicio a un barbero (`DELETE /private/barbers/{barberId}/services/{serviceId}`) se acepta siempre que la asignación exista, aunque sea la última de un servicio activo. Un servicio activo puede quedar sin barberos; mientras no tenga ninguno no se ofrece en la reserva pública (el catálogo público ya exige al menos una asignación, `HU-091`, `CA-091-01`) y vuelve a ofrecerse al asignar otro barbero. Desaparece el `409` de `DEC-068` y con él la verificación transaccional con bloqueo de fila. Retirar una asignación inexistente sigue respondiendo `404` uniforme, de modo que la operación es segura ante repetición. `DEC-068` queda sustituida en lo que exigía conservar al menos un barbero; el resto de `HU-023` (asignar, aislamiento por tenant, asociación sin datos propios) no cambia.
- **Responsable:** propietario del proyecto.
- **Motivo:** el propietario consideró que el mensaje «no puedes retirar la última asignación activa» no tenía sentido para quien administra el equipo y pidió poder retirar un servicio de cualquier peluquero. La razón de `DEC-068` (que «servicio activo» siga significando «servicio reservable») no se pierde: la reserva pública ya filtra por asignación, así que un servicio sin barberos simplemente no se ofrece.
- **Alternativas descartadas:** mantener el bloqueo duro de `DEC-068` — descartada por decisión del propietario; advertir y pedir confirmación antes de dejar el servicio sin barberos — no solicitada, queda disponible como mejora futura.
- **Riesgo residual aceptado:** un servicio activo sin barberos sigue apareciendo en el catálogo privado y en Servicios por barbero, pero no en la reserva pública ni en la cita manual (que exige la asignación, `DEC-072`). Si el servicio era el único que ofrecía la barbería, la reserva pública queda sin servicios hasta asignar otro barbero.
- **Compatibilidad:** `api/openapi/` pasa a 0.28.0: `DELETE` elimina la response `409` (`LastActiveAssignmentConflictProblem`). Única consumidora: la pantalla «Servicios por barbero». Las migraciones aplicadas no cambian: el `COMMENT` de `barber_service` y el índice `idx_barber_service_service` conservan su texto histórico.
- **Documentos afectados:** `docs/02-requisitos/historias-usuario.md` (`HU-023`, `CA-023-05`, `CA-023-06`), `docs/00-control/{dudas-pendientes,matriz-trazabilidad,historial-cambios}.md`, `docs/03-desarrollo/especificacion-frontend-nava.md`, `docs/07-calidad/{04-checklist-modulos-catalogo,05-matriz-combinaciones}.md`, `api/openapi/` (`paths/barber-services.yaml`, `CHANGELOG.md`), `apps/api/README.md`, `apps/web/README.md`.
- **Fuente:** instrucción explícita del propietario del 2026-10-04.
### DEC-115 · Perfil del panel: barbero individual

- **Fecha:** 2026-10-04.
- **Decisión:** cada barbería elige uno de dos perfiles de panel, guardado en `barbershop.panel_profile` y expuesto como `panelProfile` en `GET`/`PATCH /private/settings/brand` (el recurso de `DEC-110`): `shop` (valor inicial, el panel completo de siempre) o `solo` (barbero individual). Es **presentación pura**: no limita cuántos barberos existen, no cambia reglas de agenda ni autorización, y un barbero independiente sigue siendo una barbería con un solo barbero (`DEC-019`, glosario): no se crea un caso especial en el modelo. En el `PATCH` el campo es optativo; si falta, el perfil guardado no cambia, de modo que un cliente que no lo conoce no lo reinicia.
- **Qué cambia en `solo`:**
  1. **Navegación.** El dock queda en Agenda, Servicios, Mi perfil, Horarios, Configuración y Reserva pública. «Mi perfil» es la pantalla de Barberos vista como ficha del único barbero (nombre, foto, bloqueos), sin «Agregar», encabezados de tabla ni paginador. «Servicios por barbero» deja de ser un destino y su ruta redirige a Servicios, también cuando el perfil llega después de entrar. En el dock móvil suben Servicios, Mi perfil y Horarios; Configuración y Reserva pública quedan en «Más».
  2. **Sin selector de barbero** en Agenda, Nuevo turno y Horarios cuando hay exactamente un barbero (`DEC-074` sigue vigente: la selección vive en `barberId`; solo desaparece el control). La Agenda suma una línea de «mi día» derivada de los turnos ya cargados (cuántos quedan, el que está en curso o el siguiente); no consulta nada nuevo.
  3. **Servicios.** Cada servicio activo gana el interruptor «Lo ofrezco», que es la asignación de `HU-023` al único barbero vista desde el servicio; solo cambia cuando el servidor confirma, y un servicio nuevo queda ofrecido al crearlo (si esa asignación falla, el servicio existe y se avisa).
  4. **Configuración.** Una sección «Perfil del panel» cambia el perfil en ambos sentidos y se guarda con el resto de la marca.
  Con más de un barbero (o ninguno) las pantallas recuperan sus controles de elección y de alta, aunque el perfil sea `solo`.
- **Quién lo decidió:** el propietario, el 2026-10-04, al elegir entre tres formas de definir el perfil (ajuste de la barbería, derivado de los datos, rol por usuario).
- **Propuestas de implementación sujetas a su revisión:** la composición exacta del dock, el rótulo «Mi perfil», la línea de «mi día» y que el panel individual siga mostrando la elección donde haya más de un barbero.
- **Dependencia:** apagar «Lo ofrezco» en el único barbero retira la última asignación de un servicio activo. `DEC-068` lo rechazaba y `DEC-114` lo permite; mientras `DEC-114` no esté integrada en `main`, el API responde `409`, la pantalla conserva el estado y lo avisa, y el interruptor solo puede encenderse.
- **Contrato y datos:** columna `panel_profile text NOT NULL DEFAULT 'shop'` con `CHECK (panel_profile IN ('shop', 'solo'))` (migración `20261004120000_add_barbershop_panel_profile.sql`, sin tabla, política RLS ni `GRANT` nuevos); `panelProfile` siempre presente en la respuesta y optativo en la solicitud (OpenAPI 0.28.0, cambio compatible). Sin versión de concurrencia: es una preferencia y la última escritura gana.
- **Responsable:** propietario del proyecto.
- **Motivo:** un independiente carga hoy con gestión de equipo, selector de barbero y una matriz «Servicios por barbero» que no necesita; el propietario pidió una zona para barberos individuales con solo lo necesario.
- **Alternativas descartadas:** derivar el perfil de cuántos barberos hay (implícito, y oculta la gestión de equipo justo cuando hace falta añadir un segundo barbero); un rol por usuario dentro de una barbería con varios barberos (exige implementar `DEC-100`, autorización y RLS: un cambio crítico mucho mayor); un caso especial en el modelo de datos (contradice `DEC-019`); un destino aparte para barberos individuales en lugar de un ajuste.
- **Límites:** no concede ni retira permisos, no cambia la reserva pública ni agrega funciones de negocio (reportes, ingresos, clientes). El enlace público de reserva no se muestra en el panel porque ningún endpoint privado lo expone; hacerlo exige su propia decisión (resuelto por `DEC-117`, que lo muestra en Configuración).
- **Documentos afectados:** `docs/02-requisitos/historias-usuario.md` (`HU-025`, `CA-025-10`–`CA-025-14`), `docs/03-desarrollo/especificacion-frontend-nava.md` (§5.1), `docs/00-control/{matriz-trazabilidad,historial-cambios}.md`, `api/openapi/` (0.28.0), `database/{README.md,tests/panel_perfil_barbero_individual.sql,testdata/ui_barbero_individual_294.sql}`, `.github/workflows/ci.yml`.
- **Fuente:** elección explícita del propietario del 2026-10-04 y issue [#294](https://github.com/bcaceres19/barberia/issues/294).

### DEC-116 · «Mi perfil» del barbero individual como tarjeta propia (amplía `DEC-115`)

- **Fecha:** 2026-10-08.
- **Decisión:** en el perfil de panel `solo`, «Mi perfil» deja de reutilizar la fila de la tabla de equipo (`DEC-115`, punto 1 de «Qué cambia en `solo`») y pasa a ser su propia tarjeta: retrato grande con marco de esquinas, rol, nombre, los mismos datos (en NAVA desde, foto) y las mismas acciones (editar, bloquear), sin encabezados de tabla ni fila sin cabecera. Sigue siendo `StaffPage` quien carga los datos y abre el diálogo de edición (`MyProfileCard.vue` es solo de presentación, recibe los datos por props y emite `edit`); la tarjeta reemplaza la tabla únicamente cuando hay exactamente un barbero bajo perfil `solo`. Con más de un barbero (o ninguno) la pantalla conserva el comportamiento ya fijado por `DEC-115`: tabla de equipo completa, o el estado vacío de «Crear mi perfil».
- **Quién lo decidió:** el propietario, el 2026-10-08, al ver la pantalla real y señalar que una fila de tabla sin cabecera seguía leyéndose como una tabla, no como la ficha de una sola persona.
- **Alcance:** presentación pura, igual que `DEC-115`. No cambia datos, autorización, reglas de agenda ni el modelo (`barbershop.panel_profile` sin cambios); no toca la dependencia de `DEC-114` ya registrada en `DEC-115`.
- **Responsable:** propietario del proyecto.
- **Documentos afectados:** `docs/00-control/historial-cambios.md`.
- **Fuente:** instrucción explícita del propietario del 2026-10-08, issue [#302](https://github.com/bcaceres19/barberia/issues/302).

### DEC-117 · Enlace público único e inadivinable, visible y copiable desde la app (sustituye en parte a `DEC-082`)

- **Fecha:** 2026-10-08.
- **Decisión:**
  1. **Formato.** El `public_slug` de una barbería pasa a ser `<nombre>-<código>`: la base derivada del nombre (`shops.SlugBase`, sin cambios respecto de `DEC-082`) más un código aleatorio de 6 símbolos de `crypto/rand` sobre el alfabeto `a–z`, `2–9` sin los confundibles `i`, `l`, `o`, `0`, `1` (31⁶ ≈ 887 millones de combinaciones por nombre). Ejemplo: `corte-fino-k7x2m9`. Sigue cumpliendo `barbershop_public_slug_ck` (3 a 40 caracteres; la base se recorta para dejar sitio al código). **Sustituye el sufijo numérico determinístico `-2`, `-3`** de `DEC-082`: dos barberías con el mismo nombre ya no se distinguen por un número predecible, sino por un código que no se puede deducir. La unicidad global sigue garantizada por `idx_barbershop_public_slug`; una colisión real del código reintenta con otro dentro de la misma transacción (hasta 10 intentos; agotarlos es un error, no un bucle).
  2. **Cuándo se genera.** Al guardar el nombre por primera vez (`HU-020`, como en `DEC-082`) **o** en la primera lectura del enlace desde Configuración, lo que ocurra antes. Una barbería sin slug —que nunca guardó Configuración— deja de quedarse sin enlace público. Se genera una sola vez: renombrar la barbería no cambia el slug ni rompe el enlace compartido.
  3. **Visible y recuperable.** `GET /private/settings/public-link` (autenticado, sin parámetros) devuelve `{ slug }` de la barbería de la sesión y, si no existe, lo genera en esa misma transacción (con `FOR UPDATE`, de modo que dos primeras lecturas simultáneas convergen en un único slug). Configuración gana la sección «Enlace público» (07; «Reservas» pasa a 08) con la URL completa —el origen sale del navegador, el API solo conoce el slug—, «Copiar enlace» (con respaldo: si el portapapeles falla, el texto queda seleccionado y se explica) y «Abrir». Cubre «por si se le borra o se les pierde»: el enlace siempre está en la app. La respuesta no se almacena en caché.
  4. **Aislamiento.** El enlace de una barbería solo abre el flujo público de esa barbería (`RN-TEN-01`, `CA-090-02`, `CA-090-03`): el slug se resuelve a un único `barbershop_id` con `public_resolve_barbershop_by_slug`, el cliente nunca fija el tenant, y el endpoint privado deriva la barbería exclusivamente de la sesión (no hay forma de pedir el enlace de otra). El código aleatorio añade que el enlace de otra barbería no se pueda adivinar ni enumerar a partir de su nombre.
- **No cambia:** la resolución pública y su respuesta uniforme ante un slug inválido, desconocido o deshabilitado (`DEC-082`); el flujo de reserva; el esquema (no hay migración: la columna, el `CHECK` y el índice únicos ya existen).
- **Slugs existentes:** se conservan tal cual. Reescribirlos rompería enlaces que las barberías ya compartieron y `DEC-082` descarta la redirección; las barberías con un slug anterior (sin código) siguen resolviendo. Solo los slugs que se generen desde ahora llevan código.
- **Fuera de alcance (decisión del propietario, 2026-10-08):** regenerar el enlace y editarlo a mano. `DEC-082` preveía la edición («editable desde la configuración»); sigue sin implementarse y no se ofrece. Si el enlace se filtrara no hay forma de invalidarlo desde la app: queda como mejora futura con su propia decisión (invalidar el anterior, sin redirección, como en `DEC-082`).
- **Responsable:** propietario del proyecto.
- **Motivo:** el propietario pidió que cada empresa o barbero independiente tenga una URL pública y única con su nombre, que pueda tomarla desde la app si la pierde, y que con ella no se pueda entrar a las plataformas de agendamiento de otros clientes.
- **Alternativas descartadas:** solo el nombre con sufijo numérico (`DEC-082`: legible pero adivinable y enumerable por nombre); un código opaco sin nombre (máxima privacidad, pero la URL pierde el nombre de la empresa que se pidió); un UUID en la URL (largo, ilegible y poco compartible); reescribir los slugs existentes (rompe enlaces ya compartidos).
- **Riesgo residual aceptado:** (a) las barberías con slug anterior a esta decisión conservan un enlace sin código, adivinable por su nombre; no hay forma de regenerarlo hasta que exista esa mejora; (b) el enlace es compartible por diseño y no es un secreto de autenticación: quien lo tiene puede iniciar una reserva en esa barbería, nunca ver datos privados (`CA-090-04`).
- **Documentos afectados:** `docs/02-requisitos/historias-usuario.md` (`HU-090`, `CA-090-06`), `docs/00-control/{matriz-trazabilidad,historial-cambios,glosario}.md`, `docs/03-desarrollo/especificacion-frontend-nava.md` (Configuración), `api/openapi/` (0.30.0, `GET /private/settings/public-link`), `apps/api/internal/modules/shops`, `apps/web/src/modules/settings`.
- **Fuente:** instrucción explícita del propietario del 2026-10-08 y issue [#304](https://github.com/bcaceres19/barberia/issues/304).

### DEC-118 · Slug público sin guiones y con código de 8 caracteres (amplía `DEC-117`)

- **Fecha:** 2026-10-08.
- **Decisión:** el `public_slug` que se genera pasa de `<nombre>-<código de 6>` (p. ej. `mateo-barbero-53dx3x`) a `<nombre><código de 8>` **todo junto, sin guiones** (p. ej. `mateobarberok7x2m9q4`). La base del nombre elimina cualquier carácter fuera de `a–z`, `0–9` tras plegar acentos y pasar a minúsculas (`Mateo · Barbero` → `mateobarbero`); el código sube a 8 símbolos del mismo alfabeto sin confundibles (31⁸ ≈ 852 mil millones de combinaciones por nombre) y se pega a la base sin separador. La base se recorta a 32 caracteres para que el total no pase de 40 y el código nunca se corte. Sigue cumpliendo `barbershop_public_slug_ck` (el guion es permitido por el patrón, no obligatorio); no hay migración.
- **Se mantiene de `DEC-117`:** generación en el primer guardado o en la primera lectura del enlace, único global con reintento ante colisión, solo ver y copiar, aislamiento por tenant, slugs existentes conservados.
- **Slugs existentes:** los generados antes (con guiones y 6 símbolos) siguen resolviendo y no se reescriben.
- **Responsable:** propietario del proyecto.
- **Motivo:** al ver la pantalla real, el propietario señaló que con guiones el enlace «queda todo raro» y pidió un código algo más largo («no tanto») y todo junto.
- **Alternativas descartadas:** conservar el guion solo ante el código (`mateobarbero-k7x2m9q4`: es lo que el propietario descartó); subir a 10 o más símbolos (el propietario pidió «no tanto»; 8 ya hace inviable enumerar).
- **Riesgo residual aceptado:** sin separador, el límite entre nombre y código no es visible; es solo estético, la unicidad la garantiza el índice, no el formato.
- **Documentos afectados:** `docs/00-control/{historial-cambios,glosario}.md`, `docs/02-requisitos/historias-usuario.md` (`CA-090-06`), `api/openapi/` (0.30.1), `apps/api/internal/modules/shops`.
- **Fuente:** instrucción explícita del propietario del 2026-10-08 e issue [#306](https://github.com/bcaceres19/barberia/issues/306).

### DEC-119 · NAVA multirrubro: el vocabulario del negocio llega a toda la interfaz (amplía `DEC-110`)

- **Fecha:** 2026-10-08.
- **Decisión:** NAVA deja de ser un producto solo para barberías: lo usan también manicuristas, zonas de belleza, spas, tatuadoras y otros negocios que comparten el mismo modelo (un horario por profesional, un flujo de personas y turnos). La palabra con la que cada negocio se nombra a sí mismo y a su profesional (`DEC-110`, con plural y género gramatical) deja de aplicarse solo a una parte del panel y pasa a regir **toda superficie de texto** de la aplicación. Se entrega en dos issues independientes:
  1. **Issue [#308](https://github.com/bcaceres19/barberia/issues/308), el panel completo:** Horarios (título, selector, estados vacío, de carga y de error, bloqueos, excepciones, festivos y avisos), Nuevo turno, Detalle del turno, Política de reserva, validaciones de nombre del negocio y del profesional, y las etiquetas de estado y de historial «Cancelado por el barbero» (que pasan a decir «por la manicurista» o la palabra configurada). Las sugerencias del selector de vocabulario se amplían a otros rubros («estudio de uñas», «centro de estética», «estudio de tatuajes»; «manicurista», «esteticista», «masajista», «tatuador», «tatuadora») sin quitar el texto libre. El acceso y la recuperación, que no conocen aún a ninguna barbería, dejan de decir «barbería» y usan una palabra neutra («Gestión precisa para tu negocio.»). Una prueba estática (`vocabularyCoverage.test.ts`) impide reintroducir «barbero/barbería» como texto fijo en esas pantallas.
  2. **Issue [#309](https://github.com/bcaceres19/barberia/issues/309), lo que ve el cliente:** reserva pública, acceso al turno por enlace (`HU-098`) y mensajes salientes. `GET /public/barbershops/{slug}` y `GET /customer/appointments/{token}` devuelven un objeto `vocabulary` (`businessTerm`, `businessTermGender`, `professionalTerm`, `professionalTermPlural`, `professionalTermGender`; OpenAPI 0.31.0), leído con la misma transacción tenant-aware que ya lee el perfil, sin migración, sin función `SECURITY DEFINER` nueva y sin exponer el acento ni el perfil del panel. El cascarón de la reserva pública lo lee una vez por enlace y lo ofrece a todas las pantallas (cada una puede abrirse directamente); no monta las pantallas del recorrido hasta tenerlo, salvo la entrada, que lee el mismo perfil y usa un texto de carga neutro. El texto de «enlace no encontrado» deja de nombrar al profesional porque, sin enlace válido, no hay vocabulario que consultar. El único mensaje saliente que nombraba a la barbería (el código de verificación del sandbox de WhatsApp, que no tiene tenant) dice «NAVA»; el correo de confirmación ya usa el nombre de la barbería.
- **Lo que sustituye de decisiones anteriores:** el punto 5 de `DEC-110` («Horarios y las pantallas públicas conservan el vocabulario inicial hasta su propio cambio») queda resuelto para Horarios y el resto del panel, y para las pantallas públicas, el acceso del cliente y los mensajes queda resuelto por #309. La frase de `DEC-110` «sigue prohibido … personalizar la reserva pública o los correos» se levanta **únicamente para el vocabulario**: el color de acento, el logotipo y el modo siguen sin personalizarse en la reserva pública (`DEC-111`).
- **Lo que no cambia:** los identificadores técnicos (tabla `barber`, rutas como `/panel/barberos` y `/reservar/…/barbero`, estados `cancelled_by_barber`, eventos de `DEC-041`, claves del contrato), el término de dominio «turno» (`DEC-016`) y los estados autorizados. Es presentación pura: sin migración, sin migración en #309 ni cambio de contrato en #308 (en #309 solo se agrega la propiedad `vocabulary` a dos respuestas, compatible) y sin nuevas reglas de negocio.
- **Responsable:** propietario del proyecto.
- **Motivo:** el propietario indicó que NAVA ya no será solo de barberías y que quien la use debe poder escribir su propia palabra («barbero, barbería, manicurista, tatuadora…») en vez de ver siempre «barbería» o «barbero».
- **Alternativas descartadas:** renombrar la tabla `barber` y los contratos a `professional` (migración inmutable y contrato roto, ya descartado por `DEC-110`); un único «rubro» con plantillas cerradas de palabras (menos flexible que la palabra libre con sugerencias); derivar el género de la palabra (el español no es regular).
- **Riesgo residual aceptado:** (a) la concordancia solo conoce masculino y femenino; una palabra de género ambiguo («masajista») usa el género que el negocio elija. (b) El «Calendario de festivos colombianos» de Horarios sigue siendo específico de Colombia; ampliarlo a otros países queda fuera de esta decisión. (c) La URL pública y las rutas conservan palabras técnicas («barbero», «barberos»), que el usuario puede ver en la barra de direcciones.
- **Documentos afectados:** `docs/00-control/{registro-decisiones,historial-cambios,matriz-trazabilidad}.md`, `docs/02-requisitos/historias-usuario.md` (`HU-025`, `CA-025-15`), `docs/03-desarrollo/especificacion-frontend-nava.md` (§ vocabulario), `docs/03-desarrollo/estandar-diseno-visual.md`, `api/openapi/` (0.31.0, `PublicVocabulary`, `PublicBarbershopProfile`, `CustomerAppointmentResponse`), `apps/api/internal/modules/{publicbooking,customeraccess,notification}`, `apps/web/src/modules/{schedules,agenda,settings,staff,auth,public-booking,customer-access}`, `apps/web/src/shared/model/vocabulary.ts`.
- **Fuente:** instrucción explícita del propietario del 2026-10-08 e issues [#308](https://github.com/bcaceres19/barberia/issues/308) y [#309](https://github.com/bcaceres19/barberia/issues/309).

### DEC-120 · Campañas de interfaz con agentes Luna y fixtures privados

- **Nota de numeración (2026-10-08):** esta decisión se redactó en la rama `test/300-exploracion-ui-luna` como `DEC-116`; al integrarla, `main` ya usaba ese código para «Mi perfil» (issue #302), y como los códigos no se reutilizan se renumeró a `DEC-120`.
- **Fecha:** 2026-10-07.
- **Decisión:** por solicitud explícita del propietario se crea ui-app-testing canónico y adaptador Claude para «prueba toda la app». Orquesta GPT-6 Luna / medium por pantalla, máximo dos simultáneos, contexto mínimo, prompts persistentes y cuentas/tenants/navegadores independientes; access/recovery seriales y cierre cruzado.
- **Fuente y responsable:** propietario, instrucción de crear documentos, usuarios y skill; [#300](https://github.com/bcaceres19/barberia/issues/300).
- **Límites:** UI real sobre API/PostgreSQL locales, sin fixes ni terceros reales. Credenciales/códigos/enlaces/trazas crudas privados e ignorados. Sesión opaca DEC-050. El skill no cambia modelo del hilo actual ni promete delegación en otro runtime.
- **Criterios:** inventario, esperados trazables, estado por caso y evidencia responsive/accesible/persistencia. Preparación no equivale a campaña completa.
- **Artefactos:** AGENTS.md, .agents/skills/ui-app-testing, adaptador .claude, docs/07-calidad/pruebas-ui, prompts y tools/qa. Validación estricta del sistema obligatoria.

### DEC-121 · El backend sube a Go 1.26.9 (amplía `DEC-023`)

- **Fecha:** 2026-10-08.
- **Decisión:** `apps/api/go.mod` pasa de `go 1.25.0` / `toolchain go1.25.13` a `go 1.26.0` / `toolchain go1.26.9`, y CI instala Go 1.26. `DEC-023` fija Go como lenguaje del backend sin atar una versión menor; esta decisión fija la mínima soportada por seguridad.
- **Motivo:** `govulncheck` (control obligatorio de CI) detectó nueve vulnerabilidades de la librería estándar alcanzables por el código (`net/http`, `net/textproto`, `crypto/tls`: GO-2026-6603, 6605, 6607, 6608, 6610, 6611, 6612, 6613 y 6617). Todas se corrigen en Go 1.26.9; la línea 1.25 no publica versión corregida. Sin la subida, el job de Go falla en todos los PR y en `main`.
- **Alcance:** solo versión del lenguaje y del toolchain; ningún cambio de código de aplicación, contrato ni dependencias de terceros. Verificado con `go vet`, `go build`, `go test -race ./...` contra PostgreSQL real y `govulncheck ./...` (0 vulnerabilidades alcanzables).
- **Responsable:** propietario del proyecto (elección explícita del 2026-10-08, ante tres opciones: subir a 1.26, esperar un parche de 1.25 o integrar con el control en rojo).
- **Alternativas descartadas:** esperar un parche de 1.25 (puede no existir y deja la integración bloqueada); integrar con el check de seguridad en rojo (contradice la regla de integrar solo con CI en verde).
- **Fuente:** instrucción explícita del propietario del 2026-10-08 e issue [#317](https://github.com/bcaceres19/barberia/issues/317).

### DEC-122 · Invitación al cliente como asistente del evento de Google Calendar (amplía `DEC-099` y `DEC-101`)

- **Fecha:** 2026-10-09.
- **Decisión:** cuando la cita tiene correo del cliente y el barbero tiene Google Calendar conectado, el evento que NAVA publica en el calendario del barbero incluye a ese correo como **asistente**. Google envía la invitación y, en cuentas Google, el evento aparece en el calendario del cliente. No hay cuentas, OAuth ni tokens de clientes: el evento sigue siendo del barbero y la única conexión es la de `DEC-099`.
  1. **Cuándo se invita.** Solo con correo presente en la cita y conexión del barbero en estado `connected`. Sin correo no se invita; la cita y el evento del barbero no cambian.
  2. **Ciclo de vida.** Crear el evento envía la invitación (`sendUpdates=all`); reprogramar actualiza el mismo evento y Google avisa al cliente; cancelar desde la app elimina el evento y Google envía la cancelación. `completed` y `no_show` conservan el evento sin enviar avisos.
  3. **Sin retorno.** Aceptar, rechazar o borrar la invitación en el calendario del cliente no modifica citas, bloqueos ni disponibilidad de NAVA (`DEC-099`). NAVA no lee respuestas de asistentes.
  4. **Privacidad.** El correo del cliente solo se envía como campo de asistente; no aparece en título, descripción ni propiedades extendidas. El evento fija `guestsCanModify=false`, `guestsCanInviteOthers=false` y `guestsCanSeeOtherGuests=false`. El título sigue siendo `Nombre del cliente — Servicio` (`DEC-101`). La confirmación pública avisa al cliente, solo cuando dio correo, de que recibirá una invitación de Google Calendar.
  5. **Conexión del barbero ausente o caída.** Sin conexión `connected` no se invita ni se encola; una caída de Google nunca pierde la cita (`DEC-102`). Un cliente sin cuenta Google recibe el correo de invitación estándar con el archivo `.ics`.
- **Responsable:** propietario del proyecto.
- **Motivo:** el propietario pidió que el cliente también reciba la cita en su Google Calendar usando el correo que entrega al reservar. Google no permite escribir en el calendario de una persona sin su consentimiento OAuth; invitar como asistente es la vía soportada que no exige cuentas ni credenciales de clientes.
- **Alternativas descartadas:** OAuth de cada cliente para escribir en su calendario (contradice `DEC-099`, obliga a guardar tokens de clientes y amplía la verificación de Google); botón «Agregar a mi calendario» con `.ics` (válido, pero no automático).
- **Documentos afectados:** `alcance-mvp.md`, prompts `gcal-03` y `gcal-04`, y, al implementarse, contrato OpenAPI, diccionario de datos y política de privacidad.
- **Fuente:** instrucción explícita del propietario del 2026-10-09 («el sistema, por medio del correo que ofrece el cliente, agende la cita en su Google Calendar igualmente al barbero»); issue [#321](https://github.com/bcaceres19/barberia/issues/321).
