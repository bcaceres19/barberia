---
titulo: "Glosario y convenciones de nombres"
version: "1.1"
estado: "Vigente"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-07"
documentos_relacionados:
  - "../01-producto/reglas-negocio.md"
  - "../02-requisitos/estados-citas.md"
  - "../02-requisitos/historias-usuario.md"
  - "../01-producto/prioridades.md"
  - "supuestos.md"
  - "dudas-pendientes.md"
  - "registro-decisiones.md"
  - "contradicciones.md"
---

> **Estado del contenido:** Vocabulario confirmado por `DEC-016` y decisiones relacionadas.

---

## 1. Para qué sirve este documento

Este glosario es la **fuente de verdad de los nombres**. Cuando dos documentos usan palabras distintas para la misma cosa, la documentación empieza a contradecirse sin que nadie lo note.

Regla de uso: **si un término aparece aquí, ningún otro documento debe usar un sinónimo para referirse a lo mismo.**

Si necesitas un término nuevo, agrégalo aquí primero.

> **Nota de construcción:** la documentación se está creando por fases. Los enlaces a documentos marcados como *(pendiente de creación)* todavía no resuelven.

---

## 2. Decisión de vocabulario más importante

**Estado: Decisión confirmada** (`DEC-016`)

El producto habla de **turnos** con barberos y clientes; la entidad técnica se llama **cita**.

| Palabra | Uso permitido |
| --- | --- |
| **Cita** | Nombre de la entidad en documentación técnica, base de datos, API y código. |
| **Turno** | Nombre mostrado en interfaz, notificaciones, ayuda y comunicación comercial. |
| **Reserva** | Permitido **solo** como nombre de la *acción* de crear una cita ("reservar", "flujo de reserva", "enlace de reservas"). Nunca como nombre de la entidad. |
| **Agendamiento**, **booking**, **appointment** (en prosa) | No usar. |

**Motivo:** el producto se vende como "turnos" porque así habla el barbero, pero mezclar tres nombres para la misma entidad dentro de 90 documentos garantiza contradicciones. Se separa el lenguaje comercial del lenguaje del sistema.

---

## 3. Términos de negocio

### Barbería
Negocio que presta servicios de barbería y que constituye la **unidad de aislamiento de datos** del sistema. Todo dato operativo pertenece a exactamente una barbería.

Entidad técnica: `barbershop`.

Un barbero independiente sin local también se representa como una barbería con un solo barbero. No se crea un caso especial.

### Barbero
Persona que presta el servicio y que ocupa la agenda. Es el recurso cuya disponibilidad se calcula.

Entidad técnica: `barber`.

El MVP contempla uno o varios barberos por barbería. Disponibilidad y cruces se calculan por barbero (`DEC-019`).

### Usuario
Persona con credenciales de acceso al área privada. En el MVP, todo usuario es un barbero o el dueño de la barbería.

Entidad técnica: `user`.

**El cliente no es un usuario.** El cliente no tiene contraseña ni inicia sesión en el MVP.

### Cliente
Persona que **realiza** la reserva. Es quien deja sus datos de contacto y quien recibe las notificaciones.

Entidad técnica: `customer`.

### Persona atendida
Persona que **recibe** el servicio. Por defecto es el cliente; solo se captura aparte cuando un cliente reserva para un familiar o amigo.

Entidad técnica: `attendee`.

**Estado: Decisión confirmada.** Toda cita identifica expresamente a la persona atendida (regla `RN-RES-02`).

> Distinción crítica: si Juan reserva para su hijo Mateo, el **cliente** es Juan (recibe la confirmación y el recordatorio) y la **persona atendida** es Mateo (es a quien el barbero espera en la silla). Confundir ambos es una de las causas de error que el producto busca eliminar.

### Servicio
Trabajo que presta el barbero, con nombre, duración y precio. La duración del servicio determina cuánto espacio ocupa una cita en la agenda.

Entidad técnica: `service`.

### Cita
Compromiso entre una barbería y un cliente para atender a una persona atendida, con un barbero, un servicio y un intervalo de tiempo definido.

Entidad técnica: `appointment`.

### Bloqueo
Intervalo en el que el barbero **no** atiende, por cualquier motivo.

Entidad técnica: `time_block`.

**Estado: Decisión confirmada.** Se unifican en una sola entidad todos los casos que la operación distingue verbalmente:

| Caso del negocio | Tipo propuesto (`kind`) |
| --- | --- |
| Descanso | `break` |
| Almuerzo | `lunch` |
| Horas no disponibles | `unavailable` |
| Día libre | `day_off` |
| Festivo bloqueado | `holiday` |
| Vacaciones | `vacation` |
| Emergencia | `emergency` |

**Motivo:** los siete casos tienen el mismo comportamiento (quitan disponibilidad en un intervalo). Crear siete entidades multiplica el código sin agregar capacidad. El `kind` conserva la intención para la interfaz y para las métricas.

### Horario laboral
Definición recurrente de los días y las franjas en que la barbería atiende normalmente. Es la base sobre la que se restan bloqueos y citas.

Entidad técnica: `working_hour` (DDL-NAM-01: nombre en singular). La excepción por fecha vive en
`working_hour_override` y `working_hour_override_segment`.

### Enlace público de reservas
Dirección web que el barbero comparte con sus clientes y que da acceso al flujo de reserva sin necesidad de iniciar sesión.

---

## 4. Términos de disponibilidad y tiempo

### Disponibilidad
Conjunto de horarios de inicio en los que **cabe completo** un servicio determinado, para un barbero determinado, en una fecha determinada.

**Estado: Decisión confirmada.** Un horario que no tiene espacio suficiente para completar el servicio **no** se muestra como disponible (regla `RN-DIS-01`).

### Franja (horario ofrecido)
Cada opción de hora de inicio que se le presenta al cliente. Una franja se ofrece solo si el intervalo `[inicio, inicio + duración del servicio)` está libre.

### Intervalo de una cita
Par de instantes `starts_at` y `ends_at`.

**Estado: Decisión confirmada.** El intervalo es **semiabierto**: incluye el inicio y excluye el fin, `[starts_at, ends_at)` (`DEC-020`).

**Motivo:** permite que una cita que termina a las 10:00 y otra que empieza a las 10:00 **no** se consideren cruzadas. Con límites cerrados en ambos extremos, todas las citas consecutivas darían falso conflicto.

### Duración planificada
Cantidad de minutos que una cita reserva en la agenda y con la que se calcula `ends_at`. El servicio real puede terminar antes o después sin reescribir automáticamente las citas siguientes.

**Estado: Decisión confirmada.** Ver `DEC-002` y `RN-SER-01`.

### Anticipación mínima
Tiempo mínimo que debe transcurrir entre el momento de la reserva y el inicio de la cita.

**Estado: Decisión confirmada.** Existe, es configurable, no limita la creación manual y su valor inicial es 60 minutos (`DEC-018`).

### Ventana máxima de reservas
Cuánto hacia el futuro puede reservar un cliente.

**Estado: Decisión confirmada.** Existe, es configurable, no limita la creación manual y su valor inicial es 3 días (`DEC-018`).

### Plazo de cancelación
Tiempo mínimo antes del inicio de la cita hasta el cual el cliente puede cancelar por su cuenta.

**Estado: Decisión confirmada.** Es configurable y su valor inicial es 20 minutos (`DEC-018`, `RN-CAN-01`).

Después del plazo, la configuración define si cancela solo el barbero o también el cliente y si el motivo es obligatorio (`DEC-010`, `RN-CAN-02`).

### Zona horaria de la barbería
Zona horaria en la que el barbero y sus clientes piensan las horas.

**Estado: Decisión confirmada** que cada barbería tiene una zona, que los instantes se almacenan inequívocamente y que la interfaz muestra la zona de la barbería (`DEC-007`, `RN-DIS-07`).

**Estado: Suposición temporal** que la única zona del piloto sea `America/Bogota` (UTC−05:00, sin horario de verano). Ver [supuestos.md](supuestos.md), `SUP-002`.

---

## 5. Términos de estado de las citas

Estos tres conceptos se confunden con facilidad. **No son sinónimos.**

### Cita activa
Cita en un estado **no terminal** que sigue vigente y aparece en la agenda futura del barbero.

En el MVP: únicamente el estado `confirmed`.

> Es el sentido en el que debe leerse la regla `RN-CON-01` ("dos citas activas no pueden cruzarse").

### Cita que ocupa agenda
Cita cuyo intervalo **impide** que se cree otra cita solapada para el mismo barbero. Es el conjunto exacto que entra en la restricción de la base de datos.

**Estado: Decisión confirmada.** En el MVP: `confirmed`, `completed` y `no_show`.

**Motivo:** una cita ya atendida sigue ocupando su lugar en el pasado; permitir crear otra encima corrompería el historial. Las canceladas, en cambio, liberan el espacio.

### Estado terminal
Estado del que la cita no sale mediante una transición ordinaria. Una clasificación equivocada puede corregirse con la operación auditada `RN-CIT-04`; no se edita el historial.

Ver el detalle completo en [../02-requisitos/estados-citas.md](../02-requisitos/estados-citas.md).

---

## 6. Términos técnicos

### Concurrencia
Situación en la que dos o más solicitudes intentan ocupar el mismo espacio de agenda al mismo tiempo.

### Condición de carrera
Defecto en el que el resultado depende del orden exacto en que se ejecutan operaciones simultáneas. En este producto, la condición de carrera principal es: dos clientes consultan el mismo horario libre y ambos confirman.

### Idempotencia
Propiedad por la cual repetir la misma solicitud produce el mismo resultado que ejecutarla una sola vez, **sin efectos secundarios adicionales**.

> Ejemplo concreto: el cliente toca "Confirmar" dos veces por nerviosismo o mala señal. Debe quedar **una** cita, **un** registro de historial y **una** notificación.

### Clave de idempotencia
Identificador que el cliente envía junto a una operación crítica para que el servidor pueda reconocer un reintento de la misma operación.

Entidad técnica: `idempotency_record`.

### Eliminación lógica
Marcar un registro como eliminado sin borrarlo físicamente, conservando su historial.

Campo propuesto: `deleted_at`.

**Estado: Decisión confirmada** para los bloqueos (`DEC-009`, `RN-BLQ-04`).

### Aislamiento entre barberías (multi-tenancy)
Garantía de que ninguna barbería puede leer ni modificar datos de otra.

**Estado: Decisión confirmada.** PostgreSQL aplica RLS por `barbershop_id` en un esquema compartido (`DEC-024`). Ver [../05-backend/base-datos.md](../05-backend/base-datos.md).

### Evento de dominio
Hecho ocurrido en el sistema que puede desencadenar consecuencias. Ejemplo: `appointment_rescheduled` (DEC-041).

### Notificación
Mensaje dirigido a una persona como consecuencia de un evento. Ejemplo: aviso de que su cita cambió de hora.

### Recordatorio
Notificación **programada para el futuro** cuyo propósito es evitar el olvido de una cita.

Valor inicial: uno, 30 minutos antes; configurable entre cero y tres (`DEC-018`).

### Intento de notificación
Cada ejecución individual de envío de una notificación, con su resultado. Una notificación puede tener varios intentos.

**Estado: Decisión confirmada.** Se conserva en un registro técnico de envíos o eventos separado del historial general de la cita y sin datos personales en claro (`DEC-015`, `RN-REC-04`, `RN-DAT-02`).

> Estos cuatro conceptos se separan deliberadamente. Mezclarlos es la causa habitual de recordatorios duplicados y de mensajes enviados con un horario que ya cambió.

---

## 7. Términos comerciales

### Propietario
Persona dueña exclusiva del código, la marca, la documentación, el diseño y los demás activos del proyecto. Desarrolla el software, cubre los costos iniciales y presta el soporte inicial.

**Estado: Decisión confirmada.**

### Colaborador
Barbero que aporta contactos, validación y ventas. **No adquiere propiedad alguna sobre el software ni sobre los activos del proyecto.** Puede recibir una comisión por clientes referidos, en condiciones todavía por documentar y aprobar.

**Estado: Decisión confirmada** en cuanto a la ausencia de propiedad.
**Estado: Decisión confirmada para el MVP:** referidos y comisiones quedan fuera. Antes del primer pago, el propietario y el colaborador acuerdan por escrito porcentaje o monto fijo (`DEC-001`, `DEC-029`).

### Piloto
Primera prueba operativa del producto durante 4 semanas con 2 o 3 participantes y clientes reales (`DEC-028`).

### Uso real
Que el barbero opere su agenda en la plataforma durante su jornada. **No** es lo mismo que interés, ni que intención de usar, ni que intención de pagar. Esta distinción es obligatoria en todos los documentos del piloto y comerciales.

---

## 8. Convenciones de códigos

Todo elemento identificable lleva un código estable. **Los códigos no se reutilizan ni se renumeran**: si un elemento se descarta, su código queda retirado.

| Prefijo | Significa | Ejemplo | Documento donde vive |
| --- | --- | --- | --- |
| `RN-{ÁREA}-{nn}` | Regla de negocio | `RN-CON-01` | [reglas-negocio.md](../01-producto/reglas-negocio.md) |
| `F-{ÉPICA}-{nn}` | Función / capacidad | `F-DISP-01` | [alcance-mvp.md](../01-producto/alcance-mvp.md) |
| `DP-{ÁREA}-{nn}` | Duda pendiente | `DP-NOT-01` | [dudas-pendientes.md](dudas-pendientes.md) |
| `SUP-{nnn}` | Supuesto | `SUP-002` | [supuestos.md](supuestos.md) |
| `DEC-{nnn}` | Decisión registrada | `DEC-001` | [registro-decisiones.md](registro-decisiones.md) |
| `CT-{nnn}` | Contradicción detectada | `CT-001` | [contradicciones.md](contradicciones.md) |
| `HU-{nnn}` | Historia de usuario | `HU-014` | [historias-usuario.md](../02-requisitos/historias-usuario.md) |
| `CA-{HU}-{nn}` | Criterio de aceptación | `CA-014-03` | [historias-usuario.md](../02-requisitos/historias-usuario.md), junto a su historia |
| `TC-{nnn}` | Caso de prueba | `TC-042` | `07-calidad/` *(pendiente)* |
| `AM-{nnn}` | Amenaza de seguridad | `AM-007` | `modelo-amenazas.md` *(pendiente)* |
| `MP-{nnn}` | Métrica del piloto | `MP-005` | `metricas-piloto.md` *(pendiente)* |

### Áreas para reglas de negocio (`RN`)

| Área | Cubre |
| --- | --- |
| `RES` | Reservas y personas involucradas |
| `SER` | Servicios y duraciones |
| `DIS` | Disponibilidad |
| `BLQ` | Bloqueos y horarios |
| `CIT` | Administración de citas |
| `CAN` | Cancelaciones |
| `CON` | Concurrencia |
| `CNF` | Confirmaciones |
| `RET` | Retrasos |
| `HIS` | Historial y auditoría |
| `REC` | Recordatorios y notificaciones |
| `TEN` | Aislamiento entre barberías |
| `DAT` | Datos personales |
| `IDE` | Idempotencia |
| `PRO` | Propiedad del proyecto |

### Épicas para funciones (`F`)

| Épica | Cubre |
| --- | --- |
| `AUTH` | Acceso y recuperación de cuenta |
| `CONF` | Configuración de la barbería |
| `SERV` | Servicios |
| `HOR` | Horario laboral y bloqueos |
| `PUB` | Flujo público de reserva |
| `DISP` | Cálculo de disponibilidad |
| `CITA` | Gestión de citas del barbero |
| `EST` | Estados y transiciones |
| `NOT` | Notificaciones y recordatorios |
| `SEG` | Seguridad y aislamiento |
| `OPS` | Operación, respaldos y soporte |

### Áreas para dudas pendientes (`DP`)

`PRD` producto · `UX` experiencia de usuario · `BE` backend · `BD` base de datos · `SEG` seguridad · `NOT` notificaciones · `PIL` piloto · `COM` comercial · `LEG` legal · `OPS` operación.

---

## 9. Etiquetas de estado del contenido

Toda afirmación relevante debe poder clasificarse en una de estas siete etiquetas.

| Etiqueta | Significa | Se puede construir sobre ella |
| --- | --- | --- |
| `Estado: Decisión confirmada` | Regla aprobada. Fuente de verdad. | Sí |
| `Estado: Propuesta` | Recomendación que requiere aprobación del propietario. | Sí, marcando la dependencia |
| `Estado: Suposición temporal` | Condición asumida para poder avanzar. Puede caer. | Sí, con riesgo documentado |
| `Estado: Duda pendiente` | Decisión que corresponde al propietario. | No sin registrar el impacto |
| `Estado: Función del MVP` | Capacidad necesaria para el piloto. | Sí |
| `Estado: Función futura` | Capacidad no necesaria para validar el MVP. | No en esta fase |
| `Estado: Fuera de alcance` | No debe trabajarse sin una nueva decisión de producto. | No |

### Reglas de uso

1. Una propuesta **nunca** se redacta como si estuviera aprobada.
2. Una suposición técnica **nunca** se convierte en regla de negocio.
3. Si un documento necesita una decisión que no existe, se registra la duda y se continúa con una suposición marcada.
4. Ninguna función puede estar simultáneamente en `Función del MVP` y en `Fuera de alcance`.

---

## 10. Convenciones de redacción

| Regla | Ejemplo correcto | Ejemplo incorrecto |
| --- | --- | --- |
| Nombres técnicos en inglés y `snake_case` | `starts_at`, `time_block` | `fechaInicio`, `FechaInicio` |
| Prosa en español | "la cita se reprograma" | "la cita se reschedulea" |
| Instantes con zona horaria | "2026-08-14 09:00 (America/Bogota)" | "2026-08-14 09:00" |
| Duraciones en minutos enteros | "45 minutos" | "0,75 horas" |
| Nunca afirmar validación inexistente | "hipótesis por validar en el piloto" | "los barberos prefieren…" |
| Nunca afirmar implementación | "se propone" / "deberá" | "el sistema hace" |

---

## 11. Decisión de vocabulario cerrada

`DEC-016` resolvió `DP-PRD-01`: “turno” es la palabra de usuario y “cita” el nombre técnico estable. No quedan dudas de nomenclatura abiertas.
