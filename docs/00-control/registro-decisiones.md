---
titulo: "Registro de decisiones"
version: "1.20"
estado: "Vigente"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-28"
documentos_relacionados:
  - "contradicciones.md"
  - "matriz-trazabilidad.md"
  - "historial-cambios.md"
  - "dudas-pendientes.md"
  - "../01-producto/reglas-negocio.md"
  - "../01-producto/alcance-mvp.md"
  - "../03-desarrollo/flujo-git-github.md"
  - "../03-desarrollo/estandar-diseno-visual.md"
  - "../../respuesta-manuales/respuesta-propuestas-oc.txt"
  - "../../respuesta-manuales/respuesta-dudas-pendientes.txt"
---

# Registro de decisiones

## 1. Uso y autoridad

Cada código `DEC-*` es estable y no se reutiliza. Este registro normaliza respuestas del propietario sin borrar su fuente. Si una respuesta no alcanza para decidir el asunto al que fue asociada, se registra únicamente lo que sí decidió y se conserva abierta la duda restante.

**Responsable de las decisiones de esta versión:** propietario del proyecto. La identidad nominal está pendiente de documentar.

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
