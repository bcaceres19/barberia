# Sistema de agenda para barberías: reserva pública en línea con garantía de integridad de horarios

**Capítulo: Formulación de Proyectos de Software — Proyecto de Aula**

---

## 1. Introducción

Las micro y pequeñas empresas de servicios personales constituyen una porción significativa del tejido productivo latinoamericano y, al mismo tiempo, una de las menos alcanzadas por la transformación digital (CEPAL, 2020). En Colombia, los microestablecimientos dedicados a servicios concentran una fracción mayoritaria de las unidades económicas registradas, pero su adopción de herramientas informáticas se limita, en la mayoría de los casos, al uso de mensajería instantánea y redes sociales como reemplazo de sistemas de información propios (DANE, 2021). Este patrón no obedece a desconocimiento tecnológico, sino a una brecha de ajuste: las plataformas disponibles suelen estar dimensionadas para organizaciones con personal administrativo, capacidad de configuración y presupuesto recurrente, condiciones que un negocio de una sola persona no reúne. El resultado es que procesos críticos del negocio —entre ellos el agendamiento de citas— permanecen soportados por canales conversacionales que no fueron diseñados para ese propósito.

La barbería es un caso representativo de esta situación. Se trata de un negocio cuya unidad de producción es la cita: un intervalo de tiempo, asignado a un profesional específico, que solo puede venderse una vez. A diferencia de un producto de inventario, el tiempo no vendido no se recupera, y el tiempo vendido dos veces no se puede entregar. Toda la operación descansa, entonces, sobre la correcta administración de un recurso rígido y no acumulable. Sin embargo, el barbero independiente administra ese recurso mientras tiene las manos ocupadas: atiende a un cliente y, simultáneamente, responde por WhatsApp a quien pregunta por disponibilidad, anota una cita en un cuaderno o intenta recordar un compromiso que aceptó verbalmente. En este contexto se producen fallas que son estructurales y no accidentales: reservas cruzadas por doble asignación de la misma franja, citas olvidadas por ausencia de un recordatorio sistemático, interrupciones continuas durante la atención, y tiempo laboral consumido en informar horarios que el propio cliente podría consultar.

El problema de fondo es un problema de integridad de datos bajo concurrencia, que la ingeniería de software reconoce desde hace décadas como una de las fuentes más frecuentes de defectos difíciles de reproducir (Bernstein & Newcomer, 2009). Cuando dos clientes consultan la misma franja horaria y la solicitan con pocos segundos de diferencia, cualquier verificación realizada en la interfaz de usuario, o incluso una consulta previa a la inserción en el servidor, resulta insuficiente: entre la verificación y la escritura existe una ventana en la que el estado del sistema puede cambiar. Un diseño que confíe la prevención del cruce a la capa de presentación o a la lógica de aplicación producirá, con probabilidad no despreciable, el peor fallo posible del dominio: dos personas esperando en la puerta a la misma hora. Por esa razón, la formulación de este proyecto establece como restricción arquitectónica no negociable que la imposibilidad del cruce debe estar garantizada por la propia base de datos, mediante restricciones de exclusión sobre rangos temporales, y verificarse intentando insertar datos inconsistentes directamente en el motor, saltándose la aplicación.

A esta exigencia se suman otras que provienen del carácter multiempresa de la solución y del tratamiento de datos personales. Un sistema que aloje simultáneamente a varias barberías debe impedir que los datos de una sean accesibles desde otra, incluso ante el conocimiento o la adivinación de identificadores internos; el proyecto adopta para ello una defensa en profundidad que combina el filtrado en la aplicación con políticas de seguridad a nivel de fila (RLS) en PostgreSQL. De igual forma, los nombres, teléfonos y correos de los clientes están sujetos al régimen colombiano de protección de datos personales (Congreso de la República de Colombia, Ley 1581 de 2012), lo que impone minimización en la recolección, ausencia de datos personales en los registros técnicos y anonimización de la información vencida sin destruir el historial operativo que permite auditar la agenda. Estos requisitos no son accesorios: condicionan el modelo de datos y la observabilidad desde la formulación misma del proyecto, y no pueden incorporarse después sin rehacer decisiones estructurales.

Existen en el mercado plataformas comerciales de agendamiento para el sector de belleza y barbería que resuelven parcialmente estas necesidades. Su limitación, desde la perspectiva del usuario objetivo de este trabajo, es de escala y de costo: exigen procesos de configuración prolongados, incorporan módulos de facturación, inventario, mercadeo y fidelización que el barbero independiente no utiliza, y operan bajo suscripciones mensuales difíciles de sostener para un negocio unipersonal. Adicionalmente, muchas de ellas requieren que el cliente final cree una cuenta o instale una aplicación, una fricción que en la práctica devuelve la conversación a WhatsApp. La oportunidad que este proyecto identifica no consiste, por lo tanto, en construir una plataforma más completa, sino deliberadamente más pequeña: un alcance reducido a lo que un barbero real necesita para operar su jornada, con reserva pública sin registro para el cliente y con una operación de bajo costo sostenible por una sola persona.

En consecuencia, este proyecto formula el análisis, diseño, implementación y validación de un sistema de agenda para barberías, orientado a que un barbero independiente opere la totalidad de su jornada sin recurrir a WhatsApp para agendar y sin que se produzca un solo cruce de horarios. El alcance del producto mínimo viable quedó delimitado en 45 funciones de prioridad P0 y 52 reglas de negocio codificadas y trazables, que cubren la configuración de servicios y horarios, siete tipos de bloqueo con recurrencias y festivos, el cálculo real de disponibilidad, la reserva pública mediante enlace, la agenda diaria del barbero, la máquina de estados de las citas con historial inmutable, los recordatorios automáticos por correo y por la interfaz oficial de WhatsApp, y los procedimientos operativos de respaldo, recuperación e incidentes. La solución se construye sobre tecnologías de código abierto —backend en Go con Chi, frontend en Vue 3 con TypeScript y Vite, persistencia en PostgreSQL con migraciones versionadas mediante Atlas y contrato HTTP gobernado por OpenAPI—, decisión que reduce el costo de operación y mantiene la trazabilidad técnica del producto.

La validación del proyecto no se plantea en un entorno simulado, sino mediante un piloto en condiciones reales con dos o tres barberos de Armenia, Quindío, durante cuatro semanas, con criterios de aceptación definidos previamente y umbrales absolutos que, de incumplirse, invalidan el ejercicio: cero citas cruzadas activas, cero citas perdidas o corrompidas, cero accesos entre barberías y cero datos personales expuestos. Se hace explícito, además, un límite metodológico del piloto: puede medir uso real, facilidad de uso e intención declarada de continuar, pero no demuestra por sí mismo disposición efectiva de pago, y ningún resultado será presentado como si lo hiciera. Con este enfoque, el trabajo aporta evidencia sobre una pregunta de ingeniería concreta y verificable —si es posible garantizar integridad de agenda bajo concurrencia real con una arquitectura simple y de bajo costo— y ofrece, como subproducto, un cuerpo documental de decisiones, reglas y trazabilidad que puede ser reutilizado en proyectos de dominios equivalentes basados en la asignación de recursos temporales no acumulables.

---

## Nota sobre las referencias

Las citas incorporadas en el texto siguen el estilo APA usado en el informe de referencia. **Antes de entregar, el grupo debe verificar cada fuente y reemplazar las que no haya consultado directamente**, ya que no deben citarse trabajos no leídos. Las referencias empleadas arriba son:

| Cita en el texto | Fuente a verificar / completar |
| --- | --- |
| (CEPAL, 2020) | Informe de la Comisión Económica para América Latina y el Caribe sobre digitalización de mipymes. Verificar año, título exacto y URL. |
| (DANE, 2021) | Encuesta de Micronegocios (EMICRON) o Encuesta TIC del DANE. Verificar la publicación exacta y la cifra citada, o ajustar la afirmación a lo que la fuente realmente sostenga. |
| (Bernstein & Newcomer, 2009) | Bernstein, P. A., & Newcomer, E. (2009). *Principles of Transaction Processing* (2.ª ed.). Morgan Kaufmann. |
| (Ley 1581 de 2012) | Congreso de la República de Colombia. (2012). *Ley 1581 de 2012, por la cual se dictan disposiciones generales para la protección de datos personales.* Diario Oficial. |

**Fuentes internas del proyecto que respaldan los párrafos 3 a 7** (no requieren cita externa, pero conviene referenciarlas en el capítulo):

- Alcance, problemas atacados, funciones P0 y criterios del piloto: `docs/01-producto/alcance-mvp.md`
- Reglas de concurrencia, disponibilidad, historial y datos personales: `docs/01-producto/reglas-negocio.md` (`RN-CON-01` a `RN-CON-06`, `RN-DIS-01`, `RN-HIS-02`, `RN-DAT-01` a `RN-DAT-03`, `RN-TEN-01`)
- Decisiones de arquitectura y stack: `docs/00-control/registro-decisiones.md` (`DEC-023`, `DEC-024`, `DEC-036`, `DEC-037`)

---

## Cómo se corresponde con el ejemplo de referencia

| Párrafo | Función retórica | Equivalente en el informe Salgado |
| --- | --- | --- |
| 1 | Contexto amplio del fenómeno | Cambio climático global y GEI |
| 2 | Acotamiento al dominio específico y su vulnerabilidad | Ecosistemas de montaña, región andina y polinización |
| 3 | Núcleo del problema técnico | Falta de información sobre efectos en polinizadores |
| 4 | Restricciones adicionales del dominio | Bebederos artificiales como objeto manipulable |
| 5 | Estado del arte y vacío que se ocupa | Avances tecnológicos y monitoreo de fauna |
| 6 | Qué hace este trabajo y con qué tecnologías | Uso de sistemas *open source* para el monitoreo |
| 7 | Cómo se valida y qué aporta | Volumen de datos obtenido y utilidad futura |
