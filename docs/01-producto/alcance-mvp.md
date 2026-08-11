---
titulo: "Alcance del MVP"
version: "1.0"
estado: "Borrador"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-05"
documentos_relacionados:
  - "reglas-negocio.md"
  - "prioridades.md"
  - "../02-requisitos/estados-citas.md"
  - "../00-control/dudas-pendientes.md"
  - "../00-control/supuestos.md"
  - "../00-control/registro-decisiones.md"
---

> **Estado del contenido:** Hay 45 funciones P0 confirmadas. No queda ninguna elevación pendiente del lote de decisiones del 5 de agosto de 2026.
>
> Este documento no convierte automáticamente las propuestas en decisiones aprobadas.

---

## 1. Objetivo del MVP

**Construir la versión mínima que permita a un barbero real operar toda su agenda durante su jornada, sin volver a WhatsApp para agendar, y sin que ocurra un solo cruce de horarios.**

El MVP no busca ser un producto completo. Busca responder una sola pregunta:

> ¿Un barbero independiente que hoy agenda por WhatsApp cambia su forma de trabajar y sostiene el cambio durante varias semanas?

Todo lo que no ayude a responder esa pregunta está fuera.

### Lo que el MVP debe demostrar

| Debe demostrar | No pretende demostrar todavía |
| --- | --- |
| Que un barbero puede operar su día completo en la plataforma | Que el producto es rentable |
| Que sus clientes reservan solos, sin ayuda | Que funciona para barberías de cinco sillas |
| Que no se producen cruces de horarios | Que se puede vender a escala |
| Que el barbero recibe menos mensajes preguntando disponibilidad | Que el barbero pagará |

> **Advertencia sobre la última fila:** el piloto puede producir una conversación sobre precio y una intención declarada de pagar. Eso **no** es pago efectivo y ningún documento debe presentarlo como tal.

---

## 2. Usuarios iniciales

### Usuario principal — El barbero

Barbero independiente o dueño de una barbería pequeña en Armenia, Quindío. Trabaja de pie, con las manos ocupadas, y responde su teléfono entre cliente y cliente. Su herramienta es un celular de gama media, con conexión que no siempre es buena.

**Lo que necesita:** ver su día de un vistazo, agendar en menos de treinta segundos, bloquear tiempo cuando surge un imprevisto y dejar de contestar "¿tienes espacio hoy?".

**Lo que no tolera:** una aplicación que exija sentarse a configurar, que sea lenta, o que le haga perder una cita.

### Usuario secundario — El cliente

Persona que quiere cortarse el pelo. Abre un enlace que le llegó por WhatsApp o por Instagram. **No va a crear una cuenta, ni descargar nada, ni recordar una contraseña.**

**Lo que necesita:** ver qué horas hay libres y reservar en menos de un minuto.

### Usuario operativo — El propietario

Desarrolla, opera y soporta el producto. Necesita poder diagnosticar un problema, restaurar datos y responder al barbero el mismo día.

Detalle completo en `usuarios-iniciales.md` *(pendiente de creación)*.

---

## 3. Problemas que el MVP resuelve

Ordenados por el daño que causan hoy.

| # | Problema | Cómo lo ataca el MVP | Regla asociada |
| --- | --- | --- | --- |
| 1 | Cruces de horarios y doble reserva | Restricción en la base de datos que hace imposible almacenar dos citas cruzadas | `RN-CON-01`, `RN-CON-03` |
| 2 | Olvido de citas | Recordatorios automáticos P0 | `RN-REC-01`, `DEC-032` |
| 3 | Tiempo perdido informando horarios disponibles | Enlace público que muestra la disponibilidad real y actualizada | `RN-DIS-01` |
| 4 | Interrupciones mientras atiende a un cliente | El cliente reserva solo, sin intervención del barbero | `RN-CNF-01` |
| 5 | Horarios que parecen libres pero no caben | Solo se ofrece una franja si cabe el servicio completo | `RN-DIS-01` |
| 6 | Dificultad para bloquear descansos, días libres o emergencias | Bloqueos de todo tipo, incluso sobre citas ya agendadas | `RN-BLQ-01`, `RN-BLQ-03` |
| 7 | No saber a quién se atiende cuando alguien reserva por otro | Persona atendida como dato obligatorio y distinto del cliente | `RN-RES-02`, `RN-RES-03` |
| 8 | Falta de aviso ante cambios o cancelaciones | Notificación de cambios esenciales | `RN-REC-01` |
| 9 | Falta de historial para resolver conflictos | Historial de cambios inalterable en cada cita | `RN-HIS-01`, `RN-HIS-02` |
| 10 | Plataformas caras y sobredimensionadas | Alcance deliberadamente reducido y operación de bajo costo | — |

### Problemas reconocidos que el MVP **no** resuelve

Se dejan explícitos para que nadie asuma lo contrario:

- **Dependencia de WhatsApp para conversar.** El barbero seguirá usando WhatsApp para hablar con sus clientes. El MVP le quita la tarea de *agendar*, no la de conversar.
- **Retrasos que desordenan el día.** El sistema nunca reacomoda la agenda automáticamente. La gestión y el aviso asistido están clasificados como `F-CITA-10` P1; mientras no existan, el barbero usa contacto directo (`RN-RET-01`).
- **Clientes que no se presentan.** Se pueden registrar y contar, pero el MVP no incorpora ningún mecanismo para evitarlos.

---

## 4. Funciones P0

**Estado: Función del MVP.** Estas 45 funciones constituyen el alcance confirmado. El detalle por función se desarrollará en `10-backlog/mvp-p0.md` *(pendiente de creación)*.

### Acceso — `AUTH`

| Código | Función | Nota |
| --- | --- | --- |
| `F-AUTH-01` | Autenticación del barbero | Correo, contraseña y sesión larga. `DEC-026` |
| `F-AUTH-02` | Recuperación de acceso | Código al teléfono verificado. `DEC-026` |

### Configuración — `CONF`, `SERV`, `HOR`

| Código | Función | Nota |
| --- | --- | --- |
| `F-CONF-01` | Configuración básica de la barbería | Nombre, zona horaria, datos de contacto |
| `F-CONF-02` | Gestión de barberos | Uno o varios barberos por barbería; servicios y horario por barbero. `DEC-019` |
| `F-SERV-01` | Gestión de servicios | Crear, editar y desactivar con control sobre citas futuras. `RN-SER-03`, `RN-SER-04` |
| `F-SERV-02` | Duraciones configurables | Minutos enteros y tiempo planificado, sin lista cerrada. `RN-SER-01`, `RN-SER-02` |
| `F-HOR-01` | Configuración del horario laboral | Por día de la semana |
| `F-HOR-02` | Descansos, bloqueos y festivos | Puntuales, recurrentes, listas de fechas, excepciones y calendario colombiano activable. `DEC-020` |

### Reserva pública — `PUB`

| Código | Función | Nota |
| --- | --- | --- |
| `F-PUB-01` | Enlace público de reservas | Sin registro ni contraseña |
| `F-PUB-02` | Consulta pública de servicios | Solo servicios activos |
| `F-PUB-03` | Consulta pública de fechas y horarios | Disponibilidad real |
| `F-PUB-04` | Selección de barbero y creación pública | Elige cuando hay varios; nace en `confirmed`. `DEC-019` |
| `F-PUB-05` | Identificación de quien reserva | Nombre, teléfono y correo obligatorios. `DEC-022` |
| `F-PUB-06` | Identificación de la persona atendida | Se reutiliza el reservante; campo adicional solo si difiere. `RN-RES-02` |
| `F-PUB-07` | Confirmación y acceso al turno | Pantalla, correo y enlace aleatorio largo para consultar o cancelar. `DEC-022` |
| `F-PUB-08` | Cancelación por parte del cliente | Sujeta al plazo y a la política configurable posterior. `RN-CAN-01`, `RN-CAN-02` |

### Disponibilidad e integridad — `DISP`

| Código | Función | Nota |
| --- | --- | --- |
| `F-DISP-01` | Cálculo real de disponibilidad | Con los once factores de `RN-DIS-02` |
| `F-DISP-02` | Prevención de citas cruzadas | Garantizada por la base de datos. `RN-CON-03` |
| `F-DISP-03` | Manejo de concurrencia real | Error controlado y alternativas. `RN-CON-05` |
| `F-DISP-04` | Anticipación mínima configurable | Valor inicial: 60 minutos. `DEC-018` |
| `F-DISP-05` | Ventana máxima configurable | Valor inicial: 3 días. `DEC-018` |

### Agenda del barbero — `CITA`

| Código | Función | Nota |
| --- | --- | --- |
| `F-CITA-01` | Agenda diaria del barbero | Pantalla principal del producto |
| `F-CITA-02` | Navegación por fecha | Día anterior, siguiente, selector |
| `F-CITA-03` | Creación manual de citas | `RN-CIT-02` |
| `F-CITA-04` | Modificación de citas | Servicio, duración, datos |
| `F-CITA-05` | Reprogramación de citas | Con las mismas garantías de concurrencia |
| `F-CITA-06` | Cancelación por parte del barbero | Sin límite temporal. `RN-CAN-03` |
| `F-CITA-07` | Política y motivo de cancelación fuera de plazo | Configurable por barbería. `RN-CAN-02`, `DEC-010` |

### Estados e historial — `EST`

| Código | Función | Nota |
| --- | --- | --- |
| `F-EST-01` | Estados explícitos de citas | Ver [estados-citas.md](../02-requisitos/estados-citas.md) |
| `F-EST-02` | Transiciones y correcciones controladas | `RN-CIT-03`, `RN-CIT-04`, `RN-CIT-05` |
| `F-EST-03` | Historial de cambios inmutable | `RN-HIS-01`, `RN-HIS-02` |
| `F-EST-04` | Estado “no asistió” | P0 para preservar las métricas del piloto. `DEC-017` |

### Notificaciones — `NOT`

| Código | Función | Nota |
| --- | --- | --- |
| `F-NOT-01` | Invalidación y regeneración de recordatorios | `RN-REC-01`, `DEC-032` |
| `F-NOT-02` | Notificación de cambios esenciales | Reprogramación, cancelación y cambios aplicados a citas existentes |
| `F-NOT-03` | Recordatorios automáticos | 30 minutos y uno por defecto; configurable hasta tres. Correo y WhatsApp oficial. `DEC-018`, `DEC-027`, `DEC-032` |

### Seguridad — `SEG`

| Código | Función | Nota |
| --- | --- | --- |
| `F-SEG-01` | Separación de datos entre barberías | `RN-TEN-01` |
| `F-SEG-02` | Protección de datos personales | `RN-DAT-01`, `RN-DAT-02` |
| `F-SEG-03` | Protección escalonada contra abuso | Límite por IP y código telefónico al superar el umbral. `DEC-026` |

### Operación — `OPS`

| Código | Función | Nota |
| --- | --- | --- |
| `F-OPS-01` | Manejo de errores y conexiones inestables | Requisito de la realidad del usuario |
| `F-OPS-02` | Prevención de solicitudes duplicadas | Doble toque |
| `F-OPS-03` | Comportamiento idempotente en operaciones críticas | `RN-IDE-01` |
| `F-OPS-04` | Registro técnico mínimo para soporte | Incluye intentos de envío separados del historial general y sin datos personales. `RN-REC-04`, `RN-DAT-02` |
| `F-OPS-05` | Copias de seguridad | Diarias, retención de 30 días. `DEC-031` |
| `F-OPS-06` | Procedimiento básico de recuperación | Debe estar **probado**, no solo escrito |
| `F-OPS-09` | Procedimiento de incidentes | Contener, evaluar, avisar y registrar. `DEC-026` |

**Total del alcance P0: 45 funciones confirmadas.**

---

## 5. Funciones P1

**Estado: Mixto.** `F-CITA-10` está confirmada como P1; las demás clasificaciones conservan su estado de propuesta. Ver [prioridades.md](prioridades.md), sección 4.

### Confirmadas para una entrega posterior al núcleo P0

Costo muy bajo, valor concentrado en la calidad de lo que el piloto enseñará.

| Código | Función | Qué se pierde sin ella |
| --- | --- | --- |
| `F-CITA-08` | Notas internas del barbero | El barbero seguirá usando papel |

### Resto de P1

| Código | Función | Alternativa mientras no exista |
| --- | --- | --- |
| `F-CITA-09` | Búsqueda de clientes | Desplazarse por la agenda |
| `F-CITA-10` | Gestión y comunicación de retrasos | P1 confirmada: el barbero decide y avisa manualmente; no hay desplazamiento automático. `DEC-021` |
| `F-OPS-07` | Métricas básicas *(sin interfaz)* | Consultas directas a la base de datos durante el piloto |

### P2

| Código | Función | Alternativa |
| --- | --- | --- |
| `F-CITA-11` | Agenda semanal | Agenda diaria con navegación |
| `F-CITA-12` | Repetición rápida de citas | Crearla manualmente |
| `F-OPS-08` | Exportación básica de agenda | Captura de pantalla |

---

## 6. Funciones excluidas

**Estado: Fuera de alcance.** No deben trabajarse sin una nueva decisión de producto. La justificación completa y la condición de reconsideración de cada una se desarrollarán en `funciones-excluidas.md` *(pendiente de creación)*.

### De producto

Pagos en línea · facturación electrónica · inventario · nómina · contabilidad completa · comisiones internas avanzadas · programa de fidelización · aplicación móvil nativa · marketplace público de barberías · CRM completo · mercadeo automatizado.

### De inteligencia artificial y automatización

Inteligencia artificial para gestionar citas · chatbot avanzado · cancelación automática por falta de confirmación *(prohibida expresamente por `RN-CNF-02`)*.

### De alcance organizativo

Múltiples sedes con administración compleja · roles administrativos altamente configurables.

### De integraciones

Integraciones contables · integraciones avanzadas con redes sociales · integración bidireccional completa con calendarios externos.

### De arquitectura

Microservicios · event sourcing completo · infraestructura distribuida compleja · motor de reglas genérico · sistema de colas avanzado sin una necesidad validada.

> **Sobre las exclusiones de arquitectura:** no se posponen por falta de tiempo, sino porque **contradicen** el principio de que el producto debe poder ser operado por una sola persona. Reconsiderarlas exige un cambio en ese principio, no solo más presupuesto.

### Desplazadas a Futura

Confirmación manual opcional · gestión básica de referidos comerciales.

---

## 7. Decisiones de alcance cerradas

No quedan dudas de este lote que cambien qué se construye. Las incorporaciones principales son:

| Decisión | Efecto en alcance |
| --- | --- |
| `DEC-017` | `F-EST-04` pasa a P0 y se incorpora el cierre configurable |
| `DEC-019` | El MVP administra uno o varios barberos |
| `DEC-020` | Bloqueos recurrentes, festivos activables y turnos nocturnos entran en horario/disponibilidad |
| `DEC-026` | Defensa escalonada e incidente escrito entran en P0 |
| `DEC-027` | Correo y WhatsApp oficial son canales del MVP |
| `DEC-032` | `F-NOT-03` pasa a P0; `CT-001` queda resuelta |

---

## 8. Criterios para iniciar el piloto

**Estado: Decisión confirmada.** El desarrollo verificable, con responsable por punto, irá en `08-piloto/criterios-inicio.md` *(pendiente de creación)*.

Ningún punto es negociable el día antes de empezar. **Si uno falla, el piloto no arranca.**

### Funcional
- [ ] Flujo público de reserva completo, de principio a fin
- [ ] Agenda diaria operativa con navegación por fecha
- [ ] Servicios y horario laboral configurables
- [ ] Los siete tipos de bloqueo funcionando
- [ ] Recurrencias, excepciones y calendario colombiano de festivos funcionando
- [ ] Una barbería de prueba con más de un barbero puede configurarse y reservarse
- [ ] Reprogramación funcionando
- [ ] Cancelación por barbero y por cliente funcionando
- [ ] Historial visible y completo
- [ ] Notificaciones esenciales operativas
- [ ] Recordatorios automáticos operativos por correo y WhatsApp oficial según configuración
- [ ] `no_show`, corrección auditada y cierre de citas vencidas operativos

### Integridad
- [ ] Restricción de base de datos que impide cruces, **verificada intentando insertar directamente en la base de datos**
- [ ] Pruebas de concurrencia superadas: N solicitudes simultáneas, exactamente una cita creada
- [ ] Pruebas de idempotencia superadas: doble toque, reintento con la misma clave, reintento con contenido distinto
- [ ] Recordatorios antiguos correctamente invalidados tras una reprogramación

### Seguridad
- [ ] Aislamiento entre barberías probado endpoint por endpoint con una segunda barbería de prueba
- [ ] Políticas RLS activas y rol de aplicación sin `BYPASSRLS`
- [ ] Límite por IP y escalamiento a código telefónico verificados
- [ ] Cero vulnerabilidades críticas conocidas
- [ ] Registros técnicos revisados: sin datos personales

### Experiencia
- [ ] Flujo completo verificado en un celular de gama media real, no solo en el navegador de escritorio
- [ ] Flujo utilizable con una sola mano
- [ ] Comportamiento verificado con conexión lenta e inestable
- [ ] Doble toque en "Confirmar" no crea dos citas

### Operación
- [ ] Copias de seguridad automáticas funcionando
- [ ] **Restauración probada al menos una vez de forma completa**
- [ ] Canal de soporte acordado con el barbero y tiempo de respuesta comprometido
- [ ] Procedimiento de contingencia escrito: qué hace el barbero si la plataforma falla en plena jornada
- [ ] Procedimiento de incidentes escrito y ensayado
- [ ] Política de datos visible y acuerdo breve con cada participante
- [ ] Compromiso de soporte: menos de una hora la primera semana y el mismo día después

> El punto de restauración probada es el que más se suele omitir y el más caro de descubrir tarde. Una copia de seguridad que nunca se restauró no es una copia de seguridad.

---

## 9. Criterios de éxito del piloto

**Estado: Propuesta.** Todos los valores numéricos son propuestas, no compromisos. El desarrollo completo irá en `08-piloto/metricas-piloto.md` *(pendiente de creación)*.

El marco operativo sí está confirmado: 4 semanas, 2 o 3 participantes, método anterior solo durante la primera semana y posible extensión de otras 4 semanas (`DEC-028`).

### Umbrales absolutos — cualquier incumplimiento invalida el piloto

| Criterio | Umbral |
| --- | --- |
| Citas cruzadas activas | **Cero** |
| Citas perdidas o corrompidas | **Cero** |
| Accesos entre barberías | **Cero** |
| Datos personales expuestos | **Cero** |

### Uso real

| Criterio | Umbral propuesto | Por qué importa |
| --- | --- | --- |
| Días trabajados con la agenda abierta y actualizada | ≥ 80 % | Uso real, no interés |
| Citas reales registradas en la plataforma | ≥ 80 % del total | Si el barbero mantiene una agenda paralela, no adoptó la herramienta |
| Reservas completadas por el cliente sin ayuda | ≥ 70 % de las públicas | Si el barbero tiene que explicar cómo reservar, el flujo falló |
| Clientes reales que usaron el enlace | ≥ 10 personas distintas | Volumen mínimo para que el dato signifique algo |

### Señales cualitativas

| Señal | Cómo se verifica |
| --- | --- |
| Menos mensajes preguntando disponibilidad | Comparación declarada por el barbero al inicio y al final |
| Menos interrupciones mientras atiende | Entrevista final |
| Facilidad de uso | El barbero opera sin consultar al propietario en la última semana |
| Intención concreta de continuar | Lo pide sin que se le ofrezca |
| Conversación sobre precio | El barbero la inicia |
| Referidos | Menciona a otro barbero por su nombre |

### Clasificación del resultado

| Resultado | Condición |
| --- | --- |
| **Exitoso** | Se cumplen los umbrales absolutos, los cuatro de uso real y hay intención concreta de continuar |
| **Parcialmente exitoso** | Umbrales absolutos cumplidos, uso real por debajo del umbral pero con una causa identificada y corregible |
| **No concluyente** | Volumen insuficiente, interrupción del piloto o problema técnico que impidió el uso normal |
| **Fallido** | Se incumple un umbral absoluto, o el barbero abandona el uso antes del final |

> **Advertencia obligatoria:** ninguno de estos criterios demuestra que el barbero pagará. Interés, intención de usar, uso real, intención de pagar y pago efectivo son cinco cosas distintas, y el piloto solo puede medir las tres primeras.

---

## 10. Lo que hace que este alcance sea creíble

| Riesgo de alcance | Contención |
| --- | --- |
| El conjunto P0 crece durante el desarrollo | Toda incorporación exige quitar otra cosa ([prioridades.md](prioridades.md), sección 5) |
| Aparecen funciones "pequeñas" no documentadas | Si no tiene código `F-`, no se construye |
| Una función excluida vuelve por la puerta de atrás | La lista de exclusiones es explícita y verificable |
| Se confunde propuesta con decisión | Etiquetas de estado obligatorias en toda afirmación |
