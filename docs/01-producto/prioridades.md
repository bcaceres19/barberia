---
titulo: "Sistema de prioridades y clasificación de funciones"
version: "1.0"
estado: "Borrador"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-05"
documentos_relacionados:
  - "alcance-mvp.md"
  - "reglas-negocio.md"
  - "../00-control/glosario.md"
  - "../00-control/dudas-pendientes.md"
  - "../02-requisitos/estados-citas.md"
  - "../00-control/registro-decisiones.md"
  - "../00-control/contradicciones.md"
---

> **Estado del contenido:** Las clasificaciones afectadas por las respuestas del propietario están confirmadas. Los elementos no respondidos expresamente conservan su estado previo.
>
> Este documento no convierte automáticamente las propuestas en decisiones aprobadas.

---

## 1. Para qué sirve este documento

[alcance-mvp.md](alcance-mvp.md) dice **qué** entra en el MVP. Este documento dice **por qué**, y define el método para resolver las discusiones futuras sobre alcance sin volver a empezar la conversación cada vez.

Toda función que alguien proponga debe pasar por el filtro de la sección 3 antes de entrar al backlog.

---

## 2. Niveles de prioridad

| Nivel | Definición | Prueba de fuego |
| --- | --- | --- |
| **P0** | Indispensable para operar el producto o iniciar el piloto. | Si falta, **el piloto no puede empezar** o el producto produce un error grave de agenda. |
| **P1** | Importante, pero el piloto puede comenzar sin ella si existe una alternativa razonable. | Si falta, el barbero se las arregla con WhatsApp, papel o una llamada. |
| **P2** | Conveniente para mejorar la experiencia o la operación. | Si falta, nadie se queja de forma espontánea. |
| **Futura** | Debe posponerse hasta después de validar el MVP. | Construirla ahora consumiría tiempo que hace falta para validar la hipótesis principal. |
| **Fuera de alcance** | No debe incorporarse sin una nueva decisión de producto. | Cambia lo que el producto **es**, no solo lo que hace. |

### Reglas de aplicación

1. **P0 no es "lo que más me gusta".** Es lo que hace imposible arrancar si falta.
2. Una función no puede estar simultáneamente en dos niveles.
3. Una función clasificada como *Fuera de alcance* no puede aparecer como P0 en ningún otro documento. Si esto ocurre, es una contradicción y debe registrarse.
4. Subir una función de P1 a P0 exige justificar **qué se cae** si no está.
5. **El número de funciones P0 solo puede reducirse, salvo decisión explícita del propietario.** `DEC-005` y `DEC-010` son excepciones registradas que ampliaron la línea base; un crecimiento sin `DEC-*` es una ampliación no autorizada.

---

## 3. Filtro obligatorio para toda función propuesta

Antes de asignar prioridad, toda función debe responder estas diez preguntas.

| # | Pregunta | Para qué sirve |
| --- | --- | --- |
| 1 | ¿Qué problema resuelve? | Descarta funciones que solo suenan bien |
| 2 | ¿A qué usuario beneficia: barbero, cliente o propietario? | El propietario tiende a construir para sí mismo |
| 3 | ¿Con qué frecuencia ocurre el problema? | Un problema mensual no justifica una función |
| 4 | ¿Qué impacto tiene en el negocio si no se resuelve? | Distingue lo incómodo de lo costoso |
| 5 | ¿Qué complejidad técnica tiene? | Alta complejidad exige un beneficio proporcional |
| 6 | ¿Qué riesgos introduce? | Algunas funciones dañan la integridad de la agenda |
| 7 | ¿Qué costo operativo genera al mes? | Mensajería, almacenamiento, soporte |
| 8 | ¿Existe una alternativa manual aceptable? | Si la hay, casi nunca es P0 |
| 9 | ¿Puede posponerse sin bloquear el piloto? | Definición práctica de P0 |
| 10 | ¿Cómo se sabrá si valió la pena? | Sin métrica, no hay aprendizaje |

**Si la respuesta a la 8 es "sí, y no es tan grave", la función no es P0.**

---

## 4. Clasificación de las funciones P1 en evaluación

Cada función señalada para evaluación fue analizada con el filtro anterior y actualizada con `DEC-016`–`DEC-032`.

### 4.1 Resumen

| Función | Clasificación actual | Motivo en una línea |
| --- | --- | --- |
| Recordatorios automáticos | **P0 confirmada** | `DEC-032` resuelve `CT-001` |
| Anticipación mínima | **P0 confirmada** | 60 minutos iniciales. `DEC-018` |
| Ventana máxima de reservas | **P0 confirmada** | 3 días iniciales. `DEC-018` |
| Política y motivo de cancelación fuera de plazo | **P0 confirmada** | `DEC-010`; configurable por barbería |
| Estado "no asistió" | **P0 confirmada** | `DEC-017`; indispensable para las métricas |
| Notas internas del barbero | **P1, construir antes del piloto** | Un campo de texto; alto valor percibido |
| Búsqueda de clientes | **P1** | Necesaria para crear citas manuales rápido |
| Gestión y comunicación de retrasos | **P1 confirmada** | El barbero decide y avisa; no se desplaza la agenda automáticamente. `DEC-021` |
| Métricas básicas | **P1 sin interfaz** | Se resuelve con consultas directas durante el piloto |
| Agenda semanal | **P2** | La agenda diaria con navegación cubre el MVP |
| Repetición rápida de citas | **P2** | Muy útil, claramente posponible |
| Exportación básica de agenda | **P2** | Alternativa: captura de pantalla |
| Festivos y recurrencias | **P0 dentro de `F-HOR-02`** | Calendario activable, recurrencias y excepciones. `DEC-020` |
| Confirmación manual opcional | **Futura** | Contradice la aprobación automática confirmada |
| Gestión básica de referidos | **Futura** | Una hoja de cálculo basta durante el piloto |

### 4.2 Decisiones y elevaciones a P0

La anticipación mínima, la ventana máxima, la política de cancelación, `no_show` y los recordatorios automáticos quedaron confirmados para P0.

---

#### Recordatorios automáticos → **P0**

**Estado: Decisión confirmada** (`DEC-032`; `CT-001` resuelta)

| Pregunta | Respuesta |
| --- | --- |
| Problema | Olvido de citas, el primero de la lista de problemas del producto |
| Usuario | Barbero (pierde el ingreso) y cliente (pierde el turno) |
| Frecuencia | Varias veces por semana |
| Impacto | Cada olvido es una franja perdida que ya no se recupera |
| Complejidad | Media: tareas programadas, invalidación y regeneración, reintentos |
| Riesgo | Recordatorios duplicados o con horario viejo; costo por mensaje |
| Costo operativo | Depende del canal. Es el motivo real por el que se dudó de su prioridad |
| Alternativa manual | El barbero escribiendo por WhatsApp uno por uno, que es exactamente lo que el producto promete eliminar |
| ¿Posponible? | No sin renunciar a medir el problema principal |

**Motivo de la elevación:**

1. **El alcance del MVP ya lo presupone.** El listado de funciones P0 incluye "invalidación y regeneración de recordatorios". No se puede invalidar lo que no existe. Esta tensión se registra como contradicción (ver sección 6).
2. **El criterio de éxito del piloto incluye reducir las citas olvidadas.** Sin recordatorios, esa medición es imposible y el piloto queda parcialmente ciego.
3. **Es la razón por la que un barbero adopta la herramienta.** Una agenda sin recordatorios es un calendario; el barbero ya tiene uno.

El costo se controla con configuración por barbería y proveedores oficiales de correo y WhatsApp (`DEC-027`).

**Alcance confirmado:** uno por defecto, configurable hasta tres, 30 minutos antes por defecto y canales de correo y WhatsApp oficial configurables, con invalidación y regeneración conforme a `RN-REC-01`.

---

#### Anticipación mínima y ventana máxima → **P0 confirmadas**

**Estado: Decisión confirmada** (`DEC-005`)

| Pregunta | Respuesta |
| --- | --- |
| Problema | Un cliente reserva a las 10:00 para las 10:05, con el barbero atendiendo y sin mirar el teléfono |
| Usuario | Barbero |
| Frecuencia | Baja en volumen, alta en daño cuando ocurre |
| Impacto | El cliente llega a una cita que el barbero no sabe que existe. Es el peor fallo de confianza posible en el primer piloto |
| Complejidad | **Muy baja**: un valor de configuración y un filtro en el cálculo de disponibilidad |
| Riesgo | Ninguno relevante |
| Costo operativo | Ninguno |
| Alternativa manual | Ninguna. El barbero no puede impedirlo de otro modo |
| ¿Posponible? | Técnicamente sí; en la práctica, no conviene |

**Motivo de la elevación:** las reglas de negocio confirmadas ya incluyen "las reglas de anticipación aprobadas" entre los factores obligatorios del cálculo de disponibilidad (`RN-DIS-02`). Es decir, la regla confirmada ya asume que existen. Además, es de las funciones más baratas del catálogo: dejarla fuera no ahorra tiempo apreciable y sí expone el piloto a su fallo más embarazoso.

La misma decisión confirmó la ventana máxima configurable: el barbero puede limitar cuánto hacia el futuro reservan los clientes y las citas manuales quedan exentas de ambos límites.

Los valores iniciales son 60 minutos y 3 días (`DEC-018`).

---

#### Política y motivo de cancelación fuera de plazo → **P0 confirmada**

**Estado: Decisión confirmada** (`DEC-010`)

La política configurable determina si, vencido el plazo, cancela solo el barbero o también el cliente y si se exige motivo. No se conserva como una mejora analítica P1 porque ahora forma parte de la conducta operativa aprobada de la cancelación pública.

---

### 4.3 Funciones P1 recomendadas para construir antes del piloto

`no_show` dejó esta categoría y pasó a P0. La nota interna permanece como mejora posterior al núcleo.

| Función | Costo estimado | Qué se pierde sin ella |
| --- | --- | --- |
| Notas internas del barbero | Muy bajo | El barbero seguirá anotando en papel lo que el sistema no guarda, y esa es una señal directa de que la herramienta no reemplazó su método anterior |

---

### 4.4 Funciones desplazadas a Futura, con su justificación

Los festivos ya no están en esta sección: `DEC-020` los integra en `F-HOR-02` como calendario colombiano activable, con excepciones por barbero.

#### Confirmación manual opcional → **Futura**

Contradice directamente `RN-CNF-01` (aprobación automática) como conducta por defecto, y obliga a introducir el estado `pending` con todas sus consecuencias sobre la restricción de la base de datos.

**Condición para reconsiderarla:** que aparezcan reservas falsas o abusivas en el piloto y la aprobación automática deje de ser viable.

#### Gestión básica de referidos → **Futura**

Con un colaborador y menos de diez barberías, una hoja de cálculo es superior a cualquier función que se construya. `DEC-001` confirmó que la gestión de referidos y las comisiones quedan fuera de la versión actual; sus condiciones deben definirse antes de reactivar esa capacidad en una versión posterior.

**Condición para reconsiderarla:** más de un colaborador activo, o más de veinte barberías referidas.

---

### 4.5 Métricas básicas: una decisión deliberada

**Estado: Propuesta**

Se propone **no construir pantalla de métricas** para el piloto, y en su lugar:

1. Asegurar que el modelo de datos permita responder las preguntas del piloto mediante consultas directas a la base de datos.
2. Que el propietario ejecute esas consultas semanalmente durante el piloto.

**Motivo:** las métricas del piloto las necesita el **propietario**, no el barbero. Construir una pantalla para un solo usuario que además tiene acceso directo a la base de datos es trabajo desperdiciado. En cambio, **verificar que los datos necesarios se estén registrando sí es P0**, porque un dato no capturado no se recupera después.

**Consecuencia:** el diseño del modelo de datos debe validarse explícitamente contra la lista de métricas del piloto antes de dar por cerrada la fase de diseño.

---

## 5. Cómo se decide una discusión de alcance

Cuando aparezca la tentación de agregar algo al MVP:

1. Aplicar el filtro de la sección 3.
2. Si es P0, identificar **qué función P0 existente se puede quitar o simplificar** a cambio. El conjunto P0 tiene un tamaño finito.
3. Si no se puede quitar nada, probablemente no era P0.
4. Registrar la decisión con su motivo, para no repetir la discusión en un mes.

> Señal de alarma: si el conjunto P0 crece en dos ocasiones seguidas sin que nada salga, el alcance se está desbordando. Es uno de los riesgos principales del proyecto.

---

## 6. Contradicción resuelta

> La fuente canónica es [../00-control/contradicciones.md](../00-control/contradicciones.md), código `CT-001`, resuelta por `DEC-032`.

**Documentos en conflicto:** listado de funciones P0 frente a listado de funciones P1 en evaluación.

**Contradicción:** el conjunto P0 incluye *"invalidación y regeneración de recordatorios"* y *"notificación de cambios esenciales"*, mientras que *"recordatorios automáticos"* aparece como candidato a P1. **No es posible invalidar y regenerar recordatorios que no existen.**

**Impacto:** si se construye solo la maquinaria de invalidación sin los recordatorios, se habrá invertido en infraestructura sin ninguna función visible, y el piloto no podrá medir la reducción de olvidos, que es uno de sus criterios de éxito.

**Soluciones posibles:**
1. Elevar los recordatorios automáticos a P0 y mantener la maquinaria de invalidación en P0. *(Recomendada.)*
2. Bajar la maquinaria de invalidación a P1, junto con los recordatorios, y arrancar el piloto sin ningún tipo de recordatorio.
3. Construir solo notificaciones de cambio (avisos inmediatos al reprogramar o cancelar), sin recordatorios programados. Es más barato, pero no ataca el olvido, que es el problema principal.

**Resolución:** opción 1. `F-NOT-03` es P0, junto con `F-NOT-01` y `F-NOT-02`. Los canales son correo y WhatsApp oficial configurables (`DEC-027`).

**Estado: Decisión confirmada.**

---

## 7. Estado de la clasificación

| Categoría | Cantidad | Estado |
| --- | --- | --- |
| Funciones P0 del listado inicial | 37 | Confirmadas como P0 |
| Incorporaciones P0 por decisión formal | 3 | Confirmadas por `DEC-005` y `DEC-010` |
| Elevaciones adicionales a P0 | 2 | `F-NOT-03` y `F-EST-04` confirmadas |
| Nuevas funciones P0 operativas | 3 | `F-CONF-02`, `F-SEG-03` y `F-OPS-09` |
| P1 recomendadas después del núcleo | 1 | Notas internas |
| P1 restantes | 3 | Pendientes de aprobación |
| P2 | 3 | Pendientes de aprobación |
| Desplazadas a Futura | 2 | Confirmación manual y referidos |
| Fuera de alcance | 24 | Confirmadas como excluidas |

El desarrollo del backlog con todos los atributos por función se hará en `10-backlog/backlog-maestro.md` *(pendiente de creación)*, tomando esta clasificación como entrada.
