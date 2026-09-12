---
titulo: "Reglas de negocio"
version: "1.5"
estado: "Vigente"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-09-11"
documentos_relacionados:
  - "../00-control/glosario.md"
  - "../00-control/supuestos.md"
  - "../00-control/dudas-pendientes.md"
  - "../00-control/registro-decisiones.md"
  - "../02-requisitos/estados-citas.md"
  - "alcance-mvp.md"
---

> **Estado del contenido:** Las reglas de este catálogo están confirmadas. “Turno” es la etiqueta mostrada al usuario; “cita” se conserva como término técnico.
>
---

## 1. Cómo leer este documento

Cada regla tiene un código estable (`RN-{ÁREA}-{nn}`) que se cita desde requisitos, historias de usuario, endpoints y casos de prueba. **Los códigos no se renumeran.**

Cuando una regla y un documento posterior se contradigan, **manda esta página**.

Los términos usados aquí están definidos en [../00-control/glosario.md](../00-control/glosario.md). En particular, la diferencia entre **cita activa** y **cita que ocupa agenda** es indispensable para leer el bloque `RN-CON`.

**Resumen:** 52 reglas confirmadas. Las dudas documentales del lote del 5 de agosto de 2026 están resueltas.

---

## 2. RES — Reservas y personas involucradas

### RN-RES-01 · Un cliente puede tener varias citas

**Estado: Decisión confirmada**

Una misma persona puede tener más de una cita activa, en fechas iguales o distintas, para sí misma o para terceros.

**Ejemplo:** Andrés reserva el martes a las 9:00 para él y el mismo martes a las 9:30 para su hijo.

**Excepción:** ninguna en el MVP. No se establece un tope de citas activas por cliente.

**Casos límite:**
- Dos citas bajo los mismos datos de contacto y a la misma hora con **barberos distintos**: permitido si corresponden a personas atendidas distintas.
- Dos citas consecutivas del mismo cliente (9:00 y 9:30): permitido y es un caso frecuente y deseable (padre e hijo).

**Módulos afectados:** reserva pública, agenda del barbero.

**Pruebas asociadas:** crear dos citas para el mismo cliente en horarios distintos; crear dos citas consecutivas.

---

### RN-RES-02 · Toda cita identifica a la persona atendida

**Estado: Decisión confirmada**

Cada cita registra el nombre de la persona que recibirá el servicio. En el caso habitual se reutiliza el nombre de quien reserva; el formulario solo solicita un segundo nombre cuando el cliente indica que reserva para otra persona.

**Ejemplo:** Juan reserva y en el formulario indica que el servicio es para "Mateo Ruiz". El barbero ve en su agenda "Mateo Ruiz — Corte clásico" y sabe a quién esperar.

**Excepción:** ninguna. Aunque la interfaz no duplique el campo, el dato resuelto debe quedar almacenado en la cita.

**Casos límite:**
- Nombre de la persona atendida vacío: rechazar.
- El cliente indica solo “mi hijo”: aceptado como texto no vacío, pero la interfaz muestra un ejemplo para orientar un nombre reconocible.

**Módulos afectados:** reserva pública, creación manual de citas, agenda, notificaciones.

**Pruebas asociadas:** reservar para sí mismo; reservar para un tercero; rechazo con persona atendida vacía.

---

### RN-RES-03 · Quien reserva y quien es atendido pueden ser distintos

**Estado: Decisión confirmada**

El cliente y la persona atendida pueden ser roles separados. Por defecto son la misma persona; si difieren, los datos de contacto y las notificaciones corresponden **siempre al cliente**, no a la persona atendida.

**Ejemplo:** Juan reserva para Mateo. Los recordatorios y avisos de cambio llegan a Juan.

**Excepción:** en la creación manual, el barbero puede registrar a la persona atendida sin datos de contacto propios.

**Casos límite:**
- Persona atendida menor de edad: no se recogen sus datos de contacto. Ver `RN-DAT-01`.
- Se solicita cancelar por parte de la persona atendida, que no tiene acceso: no contemplado en el MVP; debe pedirlo a quien reservó o al barbero.

**Módulos afectados:** modelo de datos, notificaciones, cancelación.

**Pruebas asociadas:** verificar que la notificación se dirige al cliente y no a la persona atendida.

---

## 3. SER — Servicios y duraciones

### RN-SER-01 · Todo servicio tiene una duración configurable

**Estado: Decisión confirmada**

La duración se expresa en minutos enteros y la define el barbero. Representa el tiempo planificado que ocupa la cita y permite calcular su hora final; el servicio real puede terminar antes o después sin desplazar automáticamente las citas siguientes (`DEC-002`).

**Ejemplo:** "Corte clásico" 30 min · "Corte + barba" 60 min · "Corte + barba + cejas" 90 min.

**Excepción:** ninguna.

**Casos límite:**
- Duración de 0 o negativa: rechazar.
- Duración mayor que la jornada laboral completa: el servicio nunca tendrá disponibilidad. Se propone permitirlo pero advertir en la interfaz.

**Módulos afectados:** servicios, disponibilidad, citas.

**Pruebas asociadas:** crear servicios de 30, 45 y 90 minutos; rechazo de duración inválida.

---

### RN-SER-02 · La duración no está limitada a 30, 60 o 90 minutos

**Estado: Decisión confirmada**

Esos son valores frecuentes, no un conjunto cerrado. Cualquier duración permitida por la configuración es válida.

**Ejemplo:** "Perfilado de barba" de 20 minutos; "Corte de niño" de 25 minutos.

**Excepción:** ninguna.

**Casos límite:** duraciones que no son múltiplo del paso de la rejilla de horarios. Ver `RN-DIS-06`.

**Módulos afectados:** servicios, disponibilidad.

**Pruebas asociadas:** servicio de 25 minutos con rejilla de 15; verificar franjas ofrecidas.

---

### RN-SER-03 · Un servicio desactivado deja de ofrecerse pero conserva su historial

**Estado: Decisión confirmada** (`DEC-003`)

Los servicios no se eliminan físicamente. Al desactivarlos dejan de aparecer en el enlace público, pero las citas pasadas y futuras que los referencian siguen siendo legibles. Antes de completar la desactivación, el sistema advierte cuántas citas futuras quedan afectadas y permite al barbero mantenerlas o cancelarlas.

**Motivo:** eliminar un servicio rompería el historial y las citas ya agendadas.

**Ejemplo:** el barbero deja de ofrecer "Mechas". Las citas de meses anteriores siguen mostrando "Mechas" en el historial.

**Excepción:** ninguna prevista. El sistema nunca decide por el barbero qué citas cancelar.

**Casos límite:**
- Desactivar un servicio con citas futuras ya agendadas: se conservan por defecto; el barbero puede seleccionar cuáles cancelar y cada cancelación debe notificar al cliente.
- Reactivar un servicio previamente desactivado: permitido.
- Alcance por bloque (`DEC-069`): `HU-024` (B1) construye la advertencia como un recuento real de impacto antes de confirmar, sin cancelación selectiva ni protección de concurrencia sobre el conteo (no aplicable todavía porque `appointment` no existe en B1). La cancelación selectiva con notificación y el bloqueo optimista del conteo se completan en B3 junto con `appointment`, sin cambiar la forma de esta regla.

**Módulos afectados:** servicios, reserva pública, agenda, notificaciones, historial.

**Pruebas asociadas:** desactivar servicio con varias citas futuras; verificar advertencia completa, conservación por defecto, cancelación selectiva con notificación y ausencia del servicio en el enlace público.

---

### RN-SER-04 · Los cambios de catálogo no alteran citas sin decisión explícita

**Estado: Decisión confirmada** (`DEC-004`)

Cada cita almacena los datos de servicio, duración y precio con los que fue creada. Modificar el catálogo no cambia silenciosamente las citas existentes. En un cambio de precio, el barbero elige de forma explícita si lo aplica solo a citas nuevas o también a las futuras ya agendadas.

**Motivo:** si cambiar un servicio de 30 a 45 minutos recalculara todas las citas existentes, podría generar cruces masivos en una agenda ya confirmada.

**Ejemplo:** el barbero sube "Corte clásico" de 30 a 40 minutos. Las 12 citas ya agendadas conservan 30 minutos; la número 13 se crea con 40.

**Excepción:** el barbero puede editar manualmente una cita concreta o aplicar un cambio de precio a citas futuras ya agendadas. Cada cita modificada conserva el valor anterior en el historial y notifica al cliente por el canal vigente.

**Casos límite:**
- El barbero intenta aplicar una nueva duración a varias citas: debe resolver antes cualquier cruce y confirmar el alcance; nunca se mutan en segundo plano.
- El precio cambia sin aplicarse a citas existentes: no se notifica a esos clientes porque su cita no cambió.
- El aviso usa correo, WhatsApp oficial o ambos según la configuración de `RN-REC-05`.

**Módulos afectados:** servicios, citas, disponibilidad, historial.

**Pruebas asociadas:** modificar duración y precio con citas futuras; verificar el alcance elegido, ausencia de cambios silenciosos, historial y aviso de toda cita realmente modificada.

---

## 4. DIS — Disponibilidad

### RN-DIS-01 · Solo se ofrece un horario si cabe el servicio completo

**Estado: Decisión confirmada**

Un horario de inicio se muestra únicamente si el intervalo `[inicio, inicio + duración)` está enteramente libre y dentro del horario laboral.

**Ejemplo:** la jornada termina a las 18:00 y el servicio dura 60 minutos. Las 17:30 **no** se ofrecen. Las 17:00 sí.

**Excepción:** ninguna. Esta regla resuelve uno de los problemas centrales del producto.

**Casos límite:**
- Hueco de 45 minutos entre dos citas y servicio de 60: no se ofrece ninguna franja en ese hueco.
- Hueco de exactamente 30 minutos y servicio de 30: **sí** se ofrece por el intervalo semiabierto confirmado en `RN-DIS-05`.
- El servicio termina exactamente al cierre de la jornada: se ofrece.

**Módulos afectados:** disponibilidad, reserva pública.

**Pruebas asociadas:** hueco insuficiente; hueco exacto; franja final de la jornada.

---

### RN-DIS-02 · Factores que la disponibilidad debe considerar

**Estado: Decisión confirmada**

El cálculo debe tener en cuenta, como mínimo: horario laboral, duración del servicio, citas que ocupan agenda, descansos, almuerzos, bloqueos, días libres, festivos, vacaciones, emergencias, configuraciones especiales del barbero y las reglas de anticipación aprobadas.

**Ejemplo:** martes con jornada 9:00–18:00, almuerzo 13:00–14:00, cita confirmada 10:00–11:00 y bloqueo de emergencia 16:00–18:00. Las franjas ofrecidas salen de restar todo eso.

**Excepción:** ninguna.

**Casos límite:** dos factores que se solapan entre sí (almuerzo dentro de un día libre) deben restar una sola vez, no dos.

**Módulos afectados:** disponibilidad.

**Pruebas asociadas:** día con los cinco tipos de restricción activos simultáneamente.

---

### RN-DIS-03 · Un horario consultado no queda reservado

**Estado: Decisión confirmada**

Consultar la disponibilidad no bloquea ni aparta el horario. Varios clientes pueden ver la misma franja libre al mismo tiempo.

**Ejemplo:** tres personas abren el enlace y las tres ven las 10:00 disponibles.

**Excepción:** ninguna en el MVP. No existe una reserva temporal de franja.

**Casos límite:** el cliente demora diez minutos en llenar el formulario y otro confirma antes. Se aplica `RN-CON-02` y `RN-CON-05`.

**Módulos afectados:** disponibilidad, reserva pública, concurrencia.

**Pruebas asociadas:** dos sesiones consultando el mismo horario.

---

### RN-DIS-04 · Anticipación mínima y ventana máxima

**Estado: Decisión confirmada** (`DEC-005`, `DEC-018`)

No se puede reservar con menos de X minutos de anticipación ni con más de Y días de anticipación. Ambos valores son configurables por barbería. Los valores iniciales son X = 60 minutos y Y = 3 días. Rango permitido: X entero en `[0, 1440]` minutos (0 desactiva el mínimo); Y entero en `[1, 90]` días (`DEC-083`).

**Motivo:** sin anticipación mínima, un cliente puede reservar a las 10:00 para las 10:05 mientras el barbero está atendiendo y no mira el teléfono. Sin ventana máxima, aparecen reservas a ocho meses que casi nunca se cumplen.

**Ejemplo:** con X = 60 minutos, a las 10:00 la primera franja ofrecida es la de las 11:00.

**Excepción:** el barbero, al crear una cita manual, **no** está sujeto ni a la anticipación mínima ni a la ventana máxima. Puede registrar una cita para dentro de cinco minutos, una que ya empezó o una posterior a la ventana pública, siempre que cumpla las reglas de integridad de agenda.

**Casos límite:**
- Reserva enviada justo en el límite: se evalúa contra el instante de confirmación en el servidor, no en el navegador.
- Ventana máxima que cae en un día no laborable: no requiere tratamiento especial.

**Módulos afectados:** disponibilidad, reserva pública, configuración.

**Pruebas asociadas:** reserva pública por debajo de la anticipación mínima; reserva pública más allá de la ventana; citas manuales equivalentes sin esos límites.

**Fuente de los valores:** `DEC-018`.

---

### RN-DIS-05 · Los intervalos son semiabiertos

**Estado: Decisión confirmada** (`DEC-020`)

Todo intervalo de cita o de bloqueo se interpreta como `[inicio, fin)`: incluye el instante de inicio y excluye el de fin.

**Motivo:** dos citas consecutivas —una que termina a las 10:00 y otra que empieza a las 10:00— **no** se cruzan. Con límites cerrados, toda agenda apretada daría conflicto falso.

**Ejemplo:** cita A de 9:30 a 10:00 y cita B de 10:00 a 10:30 coexisten sin conflicto.

**Excepción:** ninguna. La regla se aplica de forma uniforme a citas, bloqueos y horario laboral.

**Casos límite:**
- Cita de duración cero: prohibida, `ends_at` debe ser estrictamente mayor que `starts_at`.
- Cita que cruza la medianoche: permitida si el intervalo completo cabe en el horario laboral configurado del barbero.

**Módulos afectados:** modelo de datos, restricción de base de datos, disponibilidad, pruebas de concurrencia.

**Pruebas asociadas:** citas exactamente contiguas deben permitirse; solapamiento de un minuto debe rechazarse.

> `DEC-002` define la duración planificada; `DEC-020` confirma de forma independiente los extremos `[inicio, fin)`.

---

### RN-DIS-06 · Paso de la rejilla de horarios

**Estado: Decisión confirmada** (`DEC-006`, `DEC-018`, `DEC-084`)

Las franjas ofrecidas al cliente se generan cada N minutos desde el inicio de cada tramo laboral disponible. El valor inicial de N es 15 minutos.

**Motivo:** ofrecer cualquier minuto produce listas inmanejables en un celular; un paso demasiado grande desperdicia huecos.

**Ejemplo:** con N = 15 y jornada desde las 9:00, se ofrecen 9:00, 9:15, 9:30…

**Excepción:** el barbero, en la creación manual, puede fijar cualquier hora sin ceñirse a la rejilla.

**Casos límite:**
- Un servicio de 25 minutos con rejilla de 15 deja huecos residuales de 5 minutos. Es aceptable y previsible.
- Tras un bloqueo, cita u otra interrupción que termina a las 13:47 (fuera de la rejilla original del tramo), la rejilla **reinicia desde ese instante** (`DEC-084`, resuelve `DP-PUB-03`): la siguiente franja ofrecida es 13:47 y desde ahí se generan 14:02, 14:17…, en vez de esperar al siguiente múltiplo de la rejilla original del tramo.

**Módulos afectados:** disponibilidad, interfaz pública.

**Pruebas asociadas:** generación de franjas con rejilla de 15 y servicio de 25 minutos.

**Fuente del valor:** `DEC-018`.

---

### RN-DIS-07 · Toda hora se almacena con zona horaria

**Estado: Decisión confirmada** (`DEC-007`)

Los instantes se almacenan con zona horaria y se presentan siempre en la zona horaria de la barbería.

**Motivo:** un error de zona horaria desplaza la agenda completa una o más horas. Es uno de los riesgos técnicos principales del producto.

**Ejemplo:** una cita mostrada como "14 de agosto, 9:00" corresponde a un instante inequívoco en `America/Bogota`.

**Excepción:** los registros técnicos internos pueden expresarse en UTC.

**Casos límite:**
- Cliente que reserva desde otro país o con el reloj del teléfono mal configurado: la hora mostrada y almacenada es la de la barbería.
- El servidor está configurado en otra zona: no debe influir en el resultado.

**Módulos afectados:** todo el sistema.

**Pruebas asociadas:** crear cita con el dispositivo en otra zona horaria y verificar la hora almacenada y mostrada.

---

## 5. BLQ — Bloqueos y horario laboral

### RN-BLQ-01 · Tipos de bloqueo que el barbero puede configurar

**Estado: Decisión confirmada**

Descansos, almuerzos, horas no disponibles, días libres, festivos, vacaciones, bloqueos por emergencia y horarios especiales. Pueden definirse como intervalo puntual, recurrencia o lista explícita de fechas, con excepciones individuales.

**Ejemplo:** almuerzo diario de 13:00 a 14:00; vacaciones del 15 al 30 de diciembre; emergencia de 16:00 a 18:00 el día de hoy.

**Excepción:** ninguna.

**Casos límite:**
- Bloqueo de varios días: debe poder crearse en una sola operación.
- Al editar una recurrencia, la interfaz distingue esta instancia, esta y las siguientes o toda la serie.

**Módulos afectados:** bloqueos, disponibilidad, agenda.

**Pruebas asociadas:** crear cada tipo de bloqueo y verificar su efecto en la disponibilidad.

---

### RN-BLQ-02 · Los festivos pueden trabajarse o bloquearse

**Estado: Decisión confirmada**

Cada barbero puede activar el calendario colombiano de festivos. Al activarlo, esos días se bloquean por defecto; el barbero puede abrir un festivo completo o configurar un horario especial para trabajarlo. Con la opción desactivada, el sistema no crea bloqueos de festivo.

**Ejemplo:** el barbero decide trabajar el 20 de julio con horario reducido.

**Excepción:** la decisión de un barbero no altera el calendario de otro barbero de la misma barbería.

**Casos límite:** un festivo abierto manualmente prevalece sobre el bloqueo automático de ese barbero y fecha.

**Módulos afectados:** bloqueos, disponibilidad, configuración.

**Pruebas asociadas:** festivo trabajado con horario normal; festivo bloqueado.

---

### RN-BLQ-03 · Un bloqueo puede solaparse con citas ya existentes

**Estado: Decisión confirmada** (`DEC-008`)

Crear un bloqueo **nunca** falla por chocar con citas ya agendadas. El sistema debe advertir qué citas quedan afectadas y ofrecer un flujo de pocos pasos para reprogramarlas o cancelarlas, pero **no** debe hacerlo automáticamente. Cada acción realizada notifica al cliente correspondiente.

**Motivo:** la emergencia es real y el barbero necesita bloquear su agenda de inmediato. Impedírselo porque hay citas encima sería exactamente lo contrario de lo que necesita. Y cancelar sus citas automáticamente le quitaría el control que `RN-RET-01` le reserva.

**Ejemplo:** el barbero tiene una urgencia médica y bloquea de 15:00 a 19:00. Tiene tres citas dentro. El sistema crea el bloqueo, muestra las tres y le pregunta qué hacer con cada una.

**Excepción:** ninguna.

**Casos límite:**
- Un bloqueo sobre una cita ya `completed`: se permite, sin advertencia, porque nada se puede hacer al respecto.
- Bloqueo creado mientras un cliente está confirmando esa franja: ver `RN-CON-06`.
- Reprogramar (`T2`) una cita *hacia* un intervalo con un bloqueo vigente es el caso inverso y no queda cubierto por esta regla: se **rechaza** con el mismo tratamiento que un cruce de citas (`DEC-076`), nunca con la advertencia sin bloqueo que esta regla describe para el sentido contrario.

**Módulos afectados:** bloqueos, citas, notificaciones, disponibilidad.

**Pruebas asociadas:** crear bloqueo sobre tres citas confirmadas; verificar que el bloqueo se crea, que las citas siguen activas, que se listan al barbero y que solo las acciones elegidas generan cambios y notificaciones.

> **Consecuencia técnica importante:** los bloqueos **no** entran en la restricción de exclusión de la base de datos junto a las citas. Se documentará en `05-backend/concurrencia.md` *(pendiente de creación)*.

---

### RN-BLQ-04 · Los bloqueos se eliminan de forma lógica

**Estado: Decisión confirmada** (`DEC-009`)

Al eliminar un bloqueo se marca como eliminado y deja de restar disponibilidad, pero el registro se conserva.

**Motivo:** permite investigar por qué una franja apareció o desapareció, que es una de las causas de conflicto que el producto debe poder auditar.

**Ejemplo:** el barbero elimina por error su bloqueo de vacaciones; queda constancia de quién y cuándo.

**Excepción:** ninguna.

**Casos límite:** un bloqueo eliminado no debe reaparecer nunca en el cálculo de disponibilidad.

**Módulos afectados:** bloqueos, disponibilidad, auditoría.

**Pruebas asociadas:** eliminar bloqueo y verificar que la franja vuelve a ofrecerse y que el registro persiste.

---

## 6. CIT — Administración de citas

### RN-CIT-01 · Facultades del barbero sobre las citas

**Estado: Decisión confirmada**

El barbero puede crear, modificar, reprogramar y cancelar citas, cambiar su estado mediante transiciones permitidas y consultar el historial de modificaciones.

**Ejemplo:** el cliente llama para adelantar su cita; el barbero la reprograma desde la agenda.

**Excepción:** un estado terminal no participa en las transiciones ordinarias, pero el barbero puede ejecutar la corrección auditada de `RN-CIT-04`.

**Casos límite:** reprogramar a un horario ocupado debe fallar con el mismo conflicto que una reserva pública.

**Módulos afectados:** agenda, citas, historial.

**Pruebas asociadas:** cada operación sobre una cita confirmada; rechazo sobre una cita cancelada.

---

### RN-CIT-02 · El barbero puede crear citas manuales

**Estado: Decisión confirmada**

Las citas recibidas por WhatsApp, teléfono o de forma presencial se registran manualmente y son indistinguibles de las públicas en cuanto a su efecto sobre la agenda.

**Ejemplo:** entra un cliente al local y pide cita para el jueves; el barbero la crea desde su celular.

**Excepción:** la cita manual no está sujeta a la anticipación mínima (`RN-DIS-04`) y puede omitir el correo del cliente.

**Casos límite:**
- Cliente sin teléfono ni correo: permitido, pero entonces no recibirá recordatorios y la interfaz debe advertirlo.
- Cliente sin teléfono pero con correo: `customer` se busca/reutiliza por correo dentro de la barbería; sin teléfono ni correo, la creación siempre inserta una fila nueva, nunca reconcilia por nombre (`DEC-071`).
- Servicio elegido no asignado al barbero elegido: se rechaza; la cita manual solo admite servicios activos ya asignados a ese barbero (`DEC-072`).
- Intervalo elegido coincide con un bloqueo vigente: se rechaza con el mismo tratamiento que un cruce de citas; el barbero debe retirar o exceptuar el bloqueo antes de crear la cita (`DEC-073`).
- Cita manual que se cruza con una existente: se rechaza igual que una pública (`RN-CON-01`).

**Módulos afectados:** agenda, citas, notificaciones.

**Pruebas asociadas:** crear cita manual sin datos de contacto; crear cita manual solapada.

---

### RN-CIT-03 · Solo se permiten las transiciones de estado definidas

**Estado: Decisión confirmada**

El estado de una cita cambia únicamente mediante transiciones explícitamente permitidas, ejecutadas por el actor autorizado.

**Ejemplo:** una cita cancelada no puede volver a confirmarse; debe crearse una nueva.

**Excepción:** la corrección auditada de `RN-CIT-04` es una operación explícita distinta de editar o borrar el historial.

**Casos límite:** ver la tabla completa en [../02-requisitos/estados-citas.md](../02-requisitos/estados-citas.md).

**Módulos afectados:** citas, API, historial.

**Pruebas asociadas:** intentar cada transición prohibida y verificar el rechazo.

---

### RN-CIT-04 · Un estado marcado por error se corrige con rastro

**Estado: Decisión confirmada** (`DEC-017`)

El barbero puede corregir el resultado de una cita marcado por error. La operación conserva el estado anterior, el nuevo estado, el actor, el instante y el motivo; nunca edita ni elimina la entrada original.

**Ejemplo:** una cita marcada `completed` por error se corrige a `no_show` y ambas acciones quedan visibles.

**Excepción:** una cancelación no se reabre a `confirmed`. Si el turno sí debe ocurrir, se crea una nueva cita en una franja todavía disponible.

**Módulos afectados:** citas, estados, historial, métricas.

**Pruebas asociadas:** corregir `completed` a `no_show`; intentar reabrir una cancelada con la franja ocupada; verificar el historial completo.

---

### RN-CIT-05 · El cierre de citas vencidas es configurable

**Estado: Decisión confirmada** (`DEC-017`)

Cada barbería elige entre:

1. cierre manual, con aviso al finalizar la jornada o al abrir el día siguiente; o
2. cierre automático como `completed` después de X horas configurables.

El cierre automático registra al sistema como actor y usa `ends_at` como instante efectivo del resultado para las métricas, aunque el trabajo se ejecute después.

**Excepción:** una cita ya cerrada, cancelada o corregida no vuelve a procesarse.

**Módulos afectados:** estados, agenda, tareas programadas, métricas.

**Pruebas asociadas:** ambos modos, reintento del trabajo, cita ya cerrada y cambio de configuración.

---

## 7. CAN — Cancelaciones

### RN-CAN-01 · El cliente puede cancelar hasta cierto tiempo antes de la cita

**Estado: Decisión confirmada** (`DEC-010`, `DEC-018`)

El plazo inicial es de 20 minutos y es configurable por barbería. Rango permitido: entero en `[0, 10080]` minutos (0 desactiva la cancelación propia del cliente; tope de 7 días) (`DEC-083`).

**Ejemplo:** con el valor inicial, un cliente con cita a las 15:00 puede cancelar por su cuenta hasta las 14:40.

**Excepción:** el barbero puede cancelar en cualquier momento, sin plazo (`RN-CAN-03`). Después del plazo, el cliente solo puede cancelar si la política configurable de `RN-CAN-02` lo permite.

**Casos límite:**
- Solicitud enviada exactamente en el límite: se evalúa contra el reloj del servidor.
- Cliente que intenta cancelar pasado el plazo: se evalúa la política de la barbería; si no está autorizado, recibe un mensaje que le indica contactar al barbero.

**Módulos afectados:** cancelación pública, notificaciones, configuración.

**Pruebas asociadas:** cancelar dentro del plazo; cancelar fuera del plazo; cancelar en el límite exacto.

> El valor de 20 minutos fue confirmado en `DEC-018`.

---

### RN-CAN-02 · La política de cancelación fuera de plazo es configurable

**Estado: Decisión confirmada** (`DEC-010`)

Vencido el plazo de `RN-CAN-01`, la configuración de la barbería determina si la cita solo puede cancelarla el barbero o si también puede hacerlo el cliente (campo `permite_cliente`), y si esa cancelación exige motivo (campo `motivo_obligatorio`), ambos booleanos independientes. Default de una barbería nueva: `permite_cliente = true`, `motivo_obligatorio = true` (`DEC-083`).

**Motivo:** protege al barbero de las cancelaciones de último minuto sin dejar la agenda bloqueada por una cita que ya no ocurrirá.

**Ejemplo:** el cliente intenta cancelar a las 14:50 una cita de las 15:00. En una barbería la acción se rechaza y se indica contactar al barbero; en otra se permite si registra el motivo exigido.

**Excepción:** el barbero conserva siempre la facultad de cancelar una cita activa (`RN-CAN-03`).

**Casos límite:** la interfaz pública debe mostrar la política vigente, por qué el botón no está disponible cuando corresponda y qué hacer entonces. Un cambio de configuración no altera citas ya canceladas.

**Módulos afectados:** cancelación pública, agenda, configuración, historial.

**Pruebas asociadas:** verificar las combinaciones de actor autorizado y motivo obligatorio dentro y fuera del plazo.

---

### RN-CAN-03 · El barbero puede cancelar en cualquier momento

**Estado: Decisión confirmada** (`DEC-011`)

No hay límite temporal para la cancelación por parte del barbero, ni siquiera después de la hora de inicio.

**Motivo:** las emergencias del barbero no tienen horario.

**Ejemplo:** el barbero se enferma y cancela las citas del día.

**Excepción:** no puede cancelar una cita ya en estado terminal.

**Casos límite:** cancelación masiva de un día completo: deseable, pero se evalúa como función P1. Ver [prioridades.md](prioridades.md).

**Módulos afectados:** agenda, notificaciones.

**Pruebas asociadas:** cancelar una cita cuya hora ya pasó pero que sigue confirmada.

---

### RN-CAN-04 · Una cancelación libera el horario de inmediato

**Estado: Decisión confirmada** (`DEC-012`)

En cuanto una cita pasa a un estado cancelado, su intervalo deja de ocupar agenda y vuelve a ofrecerse públicamente.

**Motivo:** es la conducta esperada y recupera ingresos para el barbero.

**Ejemplo:** se cancela la cita de las 10:00; a los pocos segundos otro cliente puede reservar esa franja.

**Excepción:** si la franja está además cubierta por un bloqueo, sigue sin ofrecerse.

**Casos límite:** cancelar una cita cuya hora ya pasó no genera disponibilidad, porque el pasado no se ofrece.

**Módulos afectados:** disponibilidad, restricción de base de datos, recordatorios.

**Pruebas asociadas:** cancelar y volver a reservar la misma franja en la misma transacción de prueba.

---

## 8. CON — Concurrencia

> Este es el bloque de reglas más crítico del producto. Su incumplimiento produce el peor fallo posible: dos clientes en la puerta a la misma hora.

### RN-CON-01 · Dos citas que ocupan agenda no pueden cruzarse para el mismo barbero

**Estado: Decisión confirmada**

Para un mismo barbero, no pueden coexistir dos citas en estados que ocupan agenda cuyos intervalos se solapen, ni siquiera parcialmente.

**Ejemplo:** existe una cita de 10:00 a 11:00. No puede crearse otra de 10:30 a 11:30, ni de 9:30 a 10:30, ni de 10:15 a 10:45.

**Excepción:** las citas canceladas no cuentan (`RN-CAN-04`). Barberos distintos de la misma barbería pueden tener citas simultáneas.

**Casos límite:**
- Citas exactamente contiguas (10:00–11:00 y 11:00–12:00): están **permitidas** por `RN-DIS-05`.
- Solapamiento de un solo minuto: **prohibido**.
- Reprogramar una cita hacia un horario ocupado: se rechaza igual que una creación.

**Módulos afectados:** base de datos, creación de citas, reprogramación, creación manual.

**Pruebas asociadas:** cruce total, cruce parcial por delante, cruce parcial por detrás, contigüidad exacta.

---

### RN-CON-02 · Gana quien confirme válidamente primero

**Estado: Decisión confirmada**

Ante solicitudes simultáneas por la misma franja, obtiene la cita quien complete válidamente la confirmación en primer lugar.

**Ejemplo:** dos clientes envían el formulario con 200 milisegundos de diferencia. Uno obtiene la cita; el otro recibe un conflicto.

**Excepción:** ninguna. No hay prioridad por antigüedad del cliente ni por orden de apertura de la página.

**Casos límite:** empate real a nivel de milisegundos: lo resuelve la base de datos de forma determinista; el resultado es arbitrario pero siempre consistente.

**Módulos afectados:** creación de citas, base de datos.

**Pruebas asociadas:** N solicitudes simultáneas sobre la misma franja; exactamente una debe tener éxito.

---

### RN-CON-03 · La base de datos debe impedir técnicamente el cruce

**Estado: Decisión confirmada**

La prevención de cruces **no** puede descansar en el frontend, ni en una consulta previa a la inserción, ni en estado en memoria, ni en bloqueos internos del proceso de la aplicación. Debe existir una restricción en la base de datos que haga imposible almacenar dos citas cruzadas.

**Ejemplo:** aunque dos procesos del servidor comprueben simultáneamente que las 10:00 están libres y ambos intenten insertar, la base de datos acepta uno solo.

**Excepción:** ninguna. Esta es la última línea de defensa.

**Casos límite:**
- Migración de datos: si existieran cruces previos, la restricción no podría crearse. Debe verificarse antes.
- Reintentos automáticos: no deben eludir la restricción.

**Módulos afectados:** base de datos, migraciones, creación y reprogramación de citas.

**Pruebas asociadas:** intento de inserción directa en la base de datos, saltándose la aplicación, que debe fallar.

> Implementación técnica confirmada con rangos y restricción de exclusión en [../05-backend/base-datos.md](../05-backend/base-datos.md).

---

### RN-CON-04 · Consultar no expulsa a nadie

**Estado: Decisión confirmada**

Que otro usuario abra o consulte el mismo horario no debe interrumpir, expulsar ni invalidar la sesión de nadie.

**Ejemplo:** el cliente está llenando el formulario y otra persona consulta la misma franja. No ocurre nada; puede seguir.

**Excepción:** ninguna.

**Casos límite:** mostrar de forma proactiva "alguien más está viendo este horario" queda **fuera del MVP**: obliga a mantener estado en tiempo real y añade ansiedad sin resolver el problema.

**Módulos afectados:** reserva pública, disponibilidad.

**Pruebas asociadas:** dos sesiones concurrentes en el mismo formulario sin interferencia.

---

### RN-CON-05 · Quien pierde recibe un error controlado y alternativas

**Estado: Decisión confirmada**

Si la franja deja de estar disponible al confirmar, el cliente debe recibir un mensaje comprensible —no un error técnico— y una lista de horarios alternativos cercanos.

**Ejemplo:** *"Ese horario se acaba de ocupar. Estas horas siguen libres hoy: 10:30, 11:00, 14:00."*

**Excepción:** ninguna.

**Casos límite:**
- No queda ninguna alternativa ese día: ofrecer el siguiente día con disponibilidad.
- Los datos ya escritos en el formulario **no** deben perderse.

**Módulos afectados:** reserva pública, disponibilidad, manejo de errores.

**Pruebas asociadas:** provocar el conflicto y verificar el mensaje, las alternativas y la conservación del formulario.

---

### RN-CON-06 · Un bloqueo creado durante una reserva no invalida la cita ya confirmada

**Estado: Decisión confirmada** (`DEC-013`)

Si el barbero crea un bloqueo mientras un cliente confirma esa misma franja, y la cita alcanza a confirmarse, la cita **permanece activa**. El sistema advierte de inmediato al barbero, aplicando `RN-BLQ-03`, para que decida mantenerla, reprogramarla o cancelarla.

**Motivo:** es la consecuencia directa de que los bloqueos no compitan por la restricción de la base de datos. Es preferible una cita que el barbero ve y decide, a una cita destruida en silencio.

**Ejemplo:** a las 14:00:03 el barbero bloquea de 16:00 a 18:00; a las 14:00:04 se confirma una cita a las 16:30. La cita existe y el barbero la ve marcada como afectada.

**Excepción:** ninguna.

**Casos límite:** si la cita se confirma **antes** del bloqueo, el resultado es idéntico. El orden no cambia el desenlace.

**Módulos afectados:** bloqueos, citas, agenda, notificaciones.

**Pruebas asociadas:** ejecutar creación de bloqueo y reserva de forma simultánea en ambos órdenes; verificar que la cita persiste, queda marcada como afectada y produce una advertencia.

---

## 9. CNF — Confirmaciones

### RN-CNF-01 · Las reservas pueden aprobarse automáticamente

**Estado: Decisión confirmada**

Una reserva pública válida queda confirmada en el acto, sin requerir aprobación previa del barbero.

**Ejemplo:** el cliente completa el formulario y ve de inmediato "Tu cita quedó confirmada".

**Excepción:** la confirmación manual opcional se evalúa como función futura. Ver [prioridades.md](prioridades.md).

**Casos límite:** ninguno relevante en el MVP.

**Módulos afectados:** reserva pública, estados de citas.

**Pruebas asociadas:** verificar que una reserva pública nace en estado `confirmed`.

---

### RN-CNF-02 · La falta de respuesta no cancela la cita

**Estado: Decisión confirmada**

Durante el MVP, ninguna cita se cancela automáticamente porque el cliente no responda a una confirmación o a un recordatorio.

**Ejemplo:** el cliente no responde al recordatorio. La cita sigue confirmada y el barbero la espera.

**Excepción:** ninguna. La cancelación automática por falta de confirmación está **fuera del alcance** del MVP.

**Casos límite:** ninguno.

**Módulos afectados:** recordatorios, estados de citas.

**Pruebas asociadas:** verificar que ningún proceso programado cambia el estado de una cita por ausencia de respuesta.

---

## 10. RET — Retrasos

### RN-RET-01 · Un retraso no cancela la cita automáticamente

**Estado: Decisión confirmada**

Los retrasos deben poder comunicarse y administrarse, pero la decisión final sobre la cita es siempre del barbero. Puede atender tarde, reprogramar o cancelar y enviar un aviso a los clientes siguientes afectados.

**Ejemplo:** el cliente avisa que llega 15 minutos tarde. El barbero decide si lo espera, lo reprograma o lo cancela.

**Excepción:** ninguna.

**Casos límite:**
- Un retraso que empuja la cita sobre la siguiente: el sistema **no** reacomoda la agenda solo; el barbero selecciona qué turnos avisar o modificar.
- Cliente que no llega y no avisa: se resuelve con el estado `no_show`, no con un retraso.

**Módulos afectados:** citas, notificaciones, agenda.

**Pruebas asociadas:** registrar un retraso y verificar que el estado de la cita no cambia.

---

## 11. HIS — Historial y auditoría

### RN-HIS-01 · Toda cita tiene historial de cambios

**Estado: Decisión confirmada**

Debe registrarse quién realizó cada cambio, cuándo, qué cambió y, cuando sea razonable, los valores anterior y nuevo. La información debe bastar para auditoría, soporte y resolución de conflictos.

**Ejemplo:** *"14/08/2026 10:32 — Barbero Carlos reprogramó la cita: hora de 15:00 a 16:30."*

**Excepción:** no se registran consultas de lectura, solo cambios.

**Casos límite:**
- Cambio originado por el cliente desde el enlace público: el actor se registra como el cliente, identificado por la cita.
- Cambio originado por un proceso automático: el actor se registra como el sistema.

**Módulos afectados:** citas, historial, soporte.

**Pruebas asociadas:** verificar la entrada de historial para cada tipo de cambio.

---

### RN-HIS-02 · El historial no se modifica ni se elimina

**Estado: Decisión confirmada** (`DEC-014`)

Las entradas de historial son de solo escritura. No se editan ni se borran, ni siquiera cuando la cita se elimina lógicamente.

**Motivo:** un historial alterable no sirve para resolver un conflicto, que es precisamente su razón de existir.

**Ejemplo:** una cita cancelada conserva íntegro el rastro de sus cambios previos.

**Excepción:** una solicitud válida de eliminación anonimiza los datos personales conforme a `RN-DAT-03`, sin borrar los hechos operativos ni las entradas de historial.

**Casos límite:** una entrada creada por error queda registrada; se corrige con una entrada nueva, no borrando la anterior.

**Módulos afectados:** historial, protección de datos.

**Pruebas asociadas:** intentar modificar y eliminar una entrada de historial; ambas deben rechazarse.

---

## 12. REC — Recordatorios y notificaciones

### RN-REC-01 · Un cambio esencial invalida y regenera los recordatorios

**Estado: Decisión confirmada**

Cuando una cita cambia de fecha, hora, servicio o duración, el sistema debe: (1) invalidar los recordatorios anteriores, (2) generar los nuevos, (3) registrar el cambio, (4) notificar al cliente, (5) evitar mensajes duplicados y (6) garantizar que los reintentos no produzcan efectos repetidos.

**Ejemplo:** una cita se mueve de las 15:00 a las 16:30. Con la anticipación inicial de 30 minutos, el recordatorio de las 14:30 queda invalidado y se programa uno nuevo para las 16:00.

**Excepción:** un cambio que no afecte fecha, hora, servicio ni duración —por ejemplo, corregir la ortografía de un nombre— no regenera recordatorios.

**Casos límite:**
- El recordatorio antiguo ya estaba siendo enviado en el instante del cambio: debe existir una verificación final antes del envío. Ver `RN-REC-03`.
- Dos reprogramaciones seguidas en un minuto: debe quedar un solo recordatorio vigente, no tres.

**Módulos afectados:** recordatorios, notificaciones, tareas programadas, historial.

**Pruebas asociadas:** reprogramar dos veces seguidas y verificar que existe exactamente un recordatorio vigente.

---

### RN-REC-02 · No se envían recordatorios de citas canceladas

**Estado: Decisión confirmada**

**Ejemplo:** el cliente cancela a las 8:00 una cita de las 15:00. El recordatorio de las 14:00 no se envía.

**Excepción:** ninguna.

**Casos límite:** cancelación ocurrida en el mismo minuto del envío programado: la verificación final debe leer el estado vigente.

**Módulos afectados:** recordatorios, tareas programadas.

**Pruebas asociadas:** cancelar una cita con recordatorio inminente y verificar que no se envía.

---

### RN-REC-03 · Nunca se envía información de un horario anterior

**Estado: Decisión confirmada**

Todo mensaje debe reflejar el estado vigente de la cita en el momento del envío, no el que tenía cuando se programó.

**Ejemplo:** un recordatorio no puede decir "tu cita es a las 15:00" si ya fue movida a las 16:30.

**Excepción:** ninguna. Este es el fallo más dañino de los recordatorios: destruye la confianza del cliente y del barbero.

**Casos límite:** el mensaje debe construirse en el momento del envío, no almacenarse ya redactado al programarlo.

**Módulos afectados:** recordatorios, plantillas, notificaciones.

**Pruebas asociadas:** programar recordatorio, reprogramar la cita, ejecutar el envío y verificar el contenido.

---

### RN-REC-04 · Se conserva evidencia de cada intento de envío

**Estado: Decisión confirmada** (`DEC-015`)

Cada intento de envío deja constancia con su resultado, aunque falle. La evidencia vive en un registro técnico específico de envíos, notificaciones o eventos y **no** en el historial general de negocio de la cita.

**Motivo:** sin evidencia es imposible responder a "yo nunca recibí el aviso", que es exactamente el tipo de conflicto que el producto debe poder resolver.

**Ejemplo:** tres intentos registrados: dos fallidos por red y uno entregado.

**Excepción:** ninguna.

**Casos límite:** los registros no deben contener el número de teléfono ni el correo en claro. Ver `RN-DAT-02`.

**Módulos afectados:** notificaciones, observabilidad, soporte, almacenamiento técnico.

**Pruebas asociadas:** forzar fallos y un éxito de envío; verificar todos los intentos, su separación del historial general y la ausencia de datos personales en claro.

---

### RN-REC-05 · Canales y eventos son configurables

**Estado: Decisión confirmada** (`DEC-027`)

El MVP soporta correo electrónico y la interfaz oficial de WhatsApp. La barbería puede activar uno o ambos y elegirlos por evento: confirmación, recordatorio, reprogramación, cancelación, retraso e inasistencia.

Para los códigos OTP de recuperación y reto adicional de acceso, la configuración del evento admite correo, WhatsApp oficial o ambos; correo es el valor predeterminado (`DEC-081`). El servidor usa contactos verificados almacenados y la interfaz nunca acepta un destino arbitrario durante el reto. Una entrega por ambos canales comparte código y operación lógica, pero conserva evidencia independiente de cada intento conforme a `RN-REC-04`.

**Excepción:** el aviso de inasistencia está desactivado por defecto. No se usan integraciones de WhatsApp no oficiales.

**Módulos afectados:** configuración, clientes, notificaciones.

**Pruebas asociadas:** matriz de eventos y canales; barbería con un canal; barbería con ambos; inasistencia desactivada.

---

### RN-REC-06 · Los recordatorios se programan al cambiar la cita

**Estado: Decisión confirmada** (`DEC-018`, `DEC-032`)

Al crear o modificar una cita se crean o actualizan, dentro de la misma operación transaccional, sus recordatorios vigentes. La anticipación inicial es 30 minutos y la cantidad inicial es 1, configurable entre 0 y 3.

Un trabajador ejecuta programaciones persistidas; no descubre tardíamente qué citas debían tener recordatorio.

**Excepción:** una cita sin canal habilitado conserva la programación como no enviable y muestra una advertencia al barbero; no inventa un destinatario.

**Módulos afectados:** citas, recordatorios, tareas programadas.

**Pruebas asociadas:** creación, reprogramación, cantidad 0/1/3, cancelación y caída del trabajador.

---

## 13. TEN — Aislamiento entre barberías

### RN-TEN-01 · Ninguna barbería accede a datos de otra

**Estado: Decisión confirmada**

Toda consulta y toda modificación quedan restringidas a la barbería del usuario autenticado. Conocer o adivinar un identificador de otra barbería no debe dar acceso a nada.

PostgreSQL aplica RLS sobre `barbershop_id` como defensa obligatoria, además de los filtros de la aplicación (`DEC-024`).

**Ejemplo:** el barbero A cambia el identificador en una petición por el de una cita del barbero B. Debe recibir un rechazo, no los datos.

**Excepción:** ninguna.

**Casos límite:**
- El enlace público expone datos de una sola barbería, los estrictamente necesarios para reservar.
- Un usuario que en el futuro pertenezca a varias barberías debe operar en una a la vez.

**Módulos afectados:** todo el backend, base de datos, API.

**Pruebas asociadas:** para cada endpoint privado, intentar acceder a un recurso de otra barbería y verificar el rechazo.

---

## 14. DAT — Datos personales

### RN-DAT-01 · Minimización de datos

**Estado: Decisión confirmada**

Solo se solicitan los datos estrictamente necesarios para prestar el servicio. En la reserva pública son obligatorios nombre, teléfono y correo; la nota es opcional y el nombre de la persona atendida solo se pide aparte si difiere. No se piden documento de identidad, dirección, fecha de nacimiento ni datos sensibles.

**Ejemplo:** quien reserva para sí mismo completa nombre, teléfono y correo; no repite su nombre.

**Excepción:** ninguna en el MVP.

**Casos límite:** una cita manual puede omitir correo cuando el barbero acepta que no podrá enviar por ese canal.

**Módulos afectados:** reserva pública, modelo de datos, protección de datos.

**Pruebas asociadas:** verificar que el formulario público no solicita campos fuera de la lista aprobada.

---

### RN-DAT-02 · Los registros técnicos no contienen datos personales

**Estado: Decisión confirmada**

Nombres, teléfonos, correos y contenidos de mensajes no se escriben en los registros técnicos. Se usan identificadores o valores enmascarados.

**Ejemplo:** el registro dice `notificación 8f3a → cliente 41b2: entregada`, no `enviado a +57 300 123 4567`.

**Excepción:** ninguna.

**Casos límite:** los mensajes de error de terceros pueden incluir el destinatario y deben depurarse antes de registrarse.

**Módulos afectados:** observabilidad, notificaciones, manejo de errores.

**Pruebas asociadas:** revisar los registros generados por un flujo completo de reserva y notificación.

---

### RN-DAT-03 · Los datos personales vencidos se anonimizan

**Estado: Decisión confirmada, sujeta a revisión jurídica** (`DEC-025`)

Nombre, teléfono, correo y tokens públicos se conservan 24 meses por defecto, con plazo parametrizado. Al vencer o prosperar una solicitud válida, se anonimizan sin eliminar la cita, el intervalo, el servicio, el estado ni su historial operativo.

**Excepción:** ampliar el plazo exige una base legal documentada; no basta cambiar el parámetro.

**Módulos afectados:** clientes, citas, historial, privacidad, tareas programadas.

**Pruebas asociadas:** anonimización vencida, revocación de token, conservación de métricas y ausencia del dato anterior en logs.

---

## 15. IDE — Idempotencia

### RN-IDE-01 · Las operaciones críticas son seguras ante reintentos

**Estado: Decisión confirmada**

Repetir una operación crítica —por doble toque, pérdida de conexión o reintento automático— debe producir el mismo resultado que ejecutarla una sola vez, sin citas duplicadas, sin entradas de historial repetidas y sin notificaciones adicionales.

**Ejemplo:** el cliente toca "Confirmar", se le va la señal, vuelve a tocar. Queda una sola cita y recibe un solo mensaje.

**Excepción:** ninguna. Aplica a creación, reprogramación, cancelación, cambios de estado, generación de recordatorios y envío de notificaciones.

**Casos límite:**
- Misma clave de idempotencia con contenido distinto: debe rechazarse como conflicto, no ejecutarse.
- Reintento después de que la operación ya haya tenido éxito: debe devolver el resultado original.

**Módulos afectados:** API, citas, notificaciones, tareas programadas.

**Pruebas asociadas:** doble envío simultáneo; reintento con la misma clave; reintento con la misma clave y contenido distinto.

---

## 16. PRO — Propiedad del proyecto

### RN-PRO-01 · El propietario es dueño exclusivo de los activos

**Estado: Decisión confirmada**

El propietario desarrolla el software, cubre inicialmente los costos, presta inicialmente el soporte y es dueño exclusivo del código, la marca, la documentación, el diseño y los demás activos.

**Ejemplo:** ningún documento del proyecto describe copropiedad, sociedad ni cesión de propiedad intelectual.

**Excepción:** ninguna.

**Casos límite:** ninguno.

**Módulos afectados:** documentación, acuerdos comerciales.

---

### RN-PRO-02 · El colaborador no adquiere propiedad

**Estado: Decisión confirmada**

El barbero colaborador aporta contactos, validación y ventas. Su participación **no** genera propiedad, copropiedad ni derechos sobre el software o la marca.

**Ejemplo:** el colaborador consigue diez barberías. Sigue sin tener participación en el producto.

**Excepción:** ninguna.

**Casos límite:** ninguno.

**Módulos afectados:** documentación comercial.

---

### RN-PRO-03 · Referidos y comisiones fuera del MVP

**Estado: Decisión confirmada para el MVP: fuera de alcance** (`DEC-001`)

La versión actual no incluye gestión de referidos ni automatiza comisiones. El propietario y el colaborador decidirán por escrito entre porcentaje o monto fijo antes del primer pago o cliente referido remunerado.

**Impacto de reabrirlo sin definirlo:** riesgo de desacuerdo con el colaborador justo cuando el producto empiece a generar ingresos.

**Módulos afectados:** documentación comercial, referidos.

Ver `DEC-029`; la condición está definida, pero no existe todavía una cifra autorizada.

---

## 17. Resumen por estado

| Estado | Reglas |
| --- | --- |
| **Decisión confirmada** (52; `RN-PRO-03` fuera del MVP y `RN-DAT-03` sujeta a revisión jurídica) | RN-RES-01/02/03 · RN-SER-01/02/03/04 · RN-DIS-01/02/03/04/05/06/07 · RN-BLQ-01/02/03/04 · RN-CIT-01/02/03/04/05 · RN-CAN-01/02/03/04 · RN-CON-01/02/03/04/05/06 · RN-CNF-01/02 · RN-RET-01 · RN-HIS-01/02 · RN-REC-01/02/03/04/05/06 · RN-TEN-01 · RN-DAT-01/02/03 · RN-IDE-01 · RN-PRO-01/02/03 |
| **Reglas íntegramente en propuesta** (0) | — |
| **Duda pendiente** (0) | — |

## 18. Reglas con condición externa

No quedan reglas de negocio abiertas. Antes del piloto todavía deben cumplirse condiciones externas: adaptación de política de datos y acuerdo escrito (`DEC-030`), verificación de proveedores oficiales de notificación (`DEC-027`) y restauración probada (`DEC-031`).
