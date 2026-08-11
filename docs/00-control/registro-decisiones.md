---
titulo: "Registro de decisiones"
version: "1.8"
estado: "Vigente"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-11"
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
