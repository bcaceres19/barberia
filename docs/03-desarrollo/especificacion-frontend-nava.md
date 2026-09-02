---
titulo: "Especificación integral de experiencia y pantallas NAVA"
version: "1.0"
estado: "Obligatorio para desarrollo incremental"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-09-01"
documentos_relacionados:
  - "../00-control/registro-decisiones.md"
  - "../01-producto/alcance-mvp.md"
  - "../01-producto/reglas-negocio.md"
  - "../02-requisitos/estados-citas.md"
  - "../02-requisitos/historias-usuario.md"
  - "../04-arquitectura/frontend.md"
  - "../10-backlog/plan-bloques.md"
  - "estandar-diseno-visual.md"
  - "estandar-frontend-vue.md"
  - "estrategia-pruebas.md"
---

# Especificación integral de experiencia y pantallas NAVA

## 1. Propósito y forma de uso

Este documento convierte la dirección visual **NAVA / Tailored Grid** aprobada por el propietario en un contrato de experiencia para `apps/web`. Debe permitir que Claude, Codex o una persona implemente cada pantalla en entregas futuras sin reinterpretar el mockup, inventar funciones ni crear estilos locales.

Esta especificación define:

- la identidad y el lenguaje visual de NAVA;
- la arquitectura de información del panel privado y de la reserva pública;
- la composición, jerarquía, interacción y responsive de cada familia de pantallas prevista;
- los componentes compartidos y sus responsabilidades;
- los estados de carga, vacío, error, conflicto y éxito;
- la frontera entre lo que pertenece a P0, lo diferido y lo excluido.

No autoriza por sí sola cambios de código. Cada adopción debe partir de una historia o issue real, una rama corta y pruebas proporcionales al comportamiento afectado. El código actual puede conservar temporalmente el sistema anterior hasta que el issue correspondiente migre una base o una pantalla completa; no se permiten migraciones visuales parciales dentro de una misma pantalla.

Fuentes normativas: `DEC-016`, `DEC-039`, `DEC-074`, `DEC-075` y `DEC-077`. Si un mockup, este documento y una regla de negocio difieren, prevalecen el registro de decisiones, el alcance, las reglas, las historias y los contratos vigentes, en ese orden documental.

## 2. Dirección de producto y marca

### 2.1 Nombre

- El producto y la aplicación se llaman **NAVA**.
- El wordmark visible se escribe `NAVA`, en mayúsculas, sin traducirlo ni agregarle “App”, “Studio” o “Barber”.
- El nombre de la barbería es dato del tenant y nunca se sustituye por NAVA.
- En el panel privado, NAVA identifica la plataforma y el nombre de la barbería identifica el contexto operativo.
- En la reserva pública, la barbería tiene la jerarquía principal y aparece la firma secundaria “Reservas con NAVA”. El cliente no debe creer que reserva en una barbería llamada NAVA.
- La decisión de producto no certifica disponibilidad de marca, dominio o registro legal; esa validación requiere una tarea separada antes de una salida comercial que la exija.

### 2.2 Personalidad

NAVA debe sentirse como sastrería contemporánea y hospitalidad boutique: precisa, serena, cálida y profesional. La interfaz toma del concepto “Tailored Grid” las líneas editoriales, la estructura temporal, la tipografía con contraste y el uso medido de marfil, tinta y latón.

Debe evitar:

- estética de fiesta, neón, graffiti, gaming o alto contraste fluorescente;
- clichés de barbería como bigotes, postes, navajas o tijeras repetidas como decoración;
- tarjetas grandes y redondeadas para cada fragmento de contenido;
- sombras profundas, degradados, vidrio, brillos o texturas falsas;
- fotografías de stock, avatares ficticios o cifras de demostración en producción;
- texto corporativo grandilocuente. NAVA habla de forma breve, humana y operativa.

### 2.3 Promesa de experiencia

1. La siguiente tarea se reconoce en menos de tres segundos.
2. Tiempo, persona, servicio y estado forman siempre la jerarquía principal de un turno.
3. El panel favorece densidad ordenada; la reserva pública favorece calma y una decisión por paso.
4. La elegancia proviene de proporción, tipografía, alineación y espacio, no de decoración.
5. Ninguna decisión visual debilita accesibilidad, rendimiento o exactitud del dominio.

## 3. Frontera de alcance

### 3.1 Leyenda

| Marca | Significado para implementación |
| --- | --- |
| `P0 existente` | Pantalla o capacidad ya construida; puede migrarse a NAVA con issue real sin cambiar su contrato. |
| `P0 pendiente` | Forma parte del MVP, pero se implementa solo cuando su HU, reglas, contrato e issue estén listos. |
| `P1` | Mejora posterior al MVP; esta especificación reserva el patrón, no autoriza construirlo. |
| `P2` | Expansión posterior; requiere priorización y decisiones adicionales. |
| `Excluido` | No debe aparecer en navegación, métricas, placeholders ni mocks de producción. |

### 3.2 Qué se conserva del mockup elegido

- wordmark NAVA y contraste entre serif editorial y sans funcional;
- lienzo marfil, superficies contenidas, tinta azul marino y latón discreto;
- agenda estructurada por tiempo, marcador de hora actual y foco en el próximo turno;
- barra de acciones contextual y navegación inferior estable;
- reserva móvil por servicio, barbero, fecha y hora;
- líneas finas, radios pequeños y composición con sensación de ficha editorial.

### 3.3 Qué no se infiere del mockup

| Elemento conceptual | Tratamiento normativo |
| --- | --- |
| `$742K`, ventas o ingresos | Excluido del MVP. No hay pagos ni métrica financiera autorizada. |
| Ocupación o rendimiento | Solo se muestra cuando una HU y un contrato definan fórmula, ventana, fuente y casos límite. Nunca se calcula de forma ad hoc en Vue. |
| Reportes | No existe como destino P0. Una métrica operativa futura no crea automáticamente un módulo de reportes. |
| Inventario | Excluido del MVP. No aparece en el dock. |
| Agenda con todas las filas de barberos | No es la vista P0. `DEC-074` exige seleccionar un barbero; la vista consolidada queda diferida. |
| Estado “En curso” | No existe en la máquina P0. El marcador “Ahora” expresa tiempo, no cambia el estado `confirmed`. |
| Perfil o cuenta de cliente | Excluido del flujo público P0. El cliente gestiona un turno mediante enlace de acceso, sin cuenta. |
| Modo oscuro o color por barbería | Excluido. El MVP tiene un tema claro NAVA. |
| Aplicación nativa | Excluida. La experiencia es web responsive. |

## 4. Fundamentos visuales obligatorios

Los valores exactos viven en [estandar-diseno-visual.md](estandar-diseno-visual.md). Esta sección define cómo se perciben y se aplican.

### 4.1 Color

- **Tinta NAVA** domina wordmark, texto, navegación y acción principal.
- **Marfil** es el lienzo cálido; no se sustituye por blanco azulado.
- **Blanco** se reserva para controles o superficies que deben despegarse del lienzo.
- **Latón oscuro** sirve para foco, detalles interactivos secundarios y énfasis sobrio.
- **Latón decorativo** puede rellenar un distintivo o una superficie editorial con texto tinta; no se usa para texto pequeño sobre marfil.
- **Salvia** y **piedra** suavizan contenido secundario. Los estados funcionales usan sus tokens semánticos, no el color de marca por aproximación.

### 4.2 Tipografía

- `Instrument Serif` se reserva al wordmark y títulos editoriales de alta jerarquía. No se usa en labels, tablas, inputs ni párrafos operativos.
- `Instrument Sans` gobierna toda la interfaz funcional.
- Horas, precios y métricas usan cifras tabulares.
- Mayúsculas completas solo se permiten en el wordmark y microetiquetas de hasta cuatro palabras; se añade espaciado de letras y nunca se usan en párrafos.
- Si las fuentes todavía no están incorporadas mediante un issue que documente licencia, archivos WOFF2 y rendimiento, se usan los fallbacks canónicos; no se enlazan CDN o Google Fonts desde un componente.

### 4.3 Forma, profundidad y ritmo

- Controles: radio 4 px. Superficies: 6 px. Diálogos: 8 px. Las insignias de estado pueden ser pill.
- Bordes de 1 px estructuran. Bordes de 2 px se reservan para foco, selección o error.
- Las tarjetas no flotan por defecto. La separación normal es fondo + línea + espacio.
- La sombra elevada solo aparece en menú, barra sticky, panel inferior o diálogo.
- La escala espacial de 4 px es obligatoria. Las columnas, reglas y baseline deben alinear contenidos relacionados.

### 4.4 Iconografía e imagen

- SVG de línea coherente, caja 20 o 24 px, trazo visual de 1.75–2 px.
- Todo icono funcional vive dentro de un control nombrado; color solo no comunica estado.
- Avatares reales solo si el producto incorpora una fuente y política para fotos. Hasta entonces se usa monograma accesible o ninguna imagen.
- No se generan retratos ficticios para representar barberos o clientes reales.

## 5. Arquitectura de información

### 5.1 Panel privado

La navegación principal P0 se agrupa así:

1. **Agenda**: agenda diaria, cambio de fecha, detalle, nuevo turno y acciones de ciclo de vida.
2. **Servicios**: catálogo y asignación de servicios por barbero.
3. **Barberos**: listado y configuración del equipo.
4. **Horarios**: jornada base, excepciones, festivos y bloqueos.
5. **Configuración**: barbería, reglas de reserva, cancelación, recordatorios y canales conforme existan sus HU.

“Nuevo turno” es una acción primaria global, no un módulo de navegación. “Bloqueos” puede aparecer como subvista visible de Horarios. No se crean destinos P0 para Clientes, Caja, Reportes, Inventario, Finanzas ni Ajustes genéricos sin una capacidad aprobada.

### 5.2 Navegación por ancho

| Ancho | Patrón |
| --- | --- |
| 320–767 px | Header compacto; contenido a una columna; dock inferior `Agenda · Nuevo · Horarios · Más`; acciones de formulario sticky cuando ayuden. |
| 768–1023 px | Header completo; dock inferior o rail compacto según espacio real; contenido en una o dos columnas. No se cambia el orden semántico. |
| 1024 px o más | Header editorial con NAVA, contexto y acción principal; contenido en grid; dock inferior de escritorio. No hay sidebar lateral en la dirección NAVA. |

El dock inferior de escritorio es parte de la identidad seleccionada. Permanece visible sin cubrir contenido, tiene altura objetivo de 64 px y usa icono más texto. En móvil suma `env(safe-area-inset-bottom)`.

### 5.3 Shell privado

```text
┌──────────────────────────────────────────────────────────────┐
│ NAVA · nombre de barbería        contexto       Nuevo turno │
├──────────────────────────────────────────────────────────────┤
│ título + descripción breve                    acción local   │
│ filtros / resumen autorizado                                │
│                                                              │
│ contenido principal NAVA                                    │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│ Agenda · Servicios · Barberos · Horarios · Configuración    │
└──────────────────────────────────────────────────────────────┘
```

- El header no contiene métricas sin contrato.
- La sesión y salida se ubican en el menú de contexto del usuario, no como ítem dominante del dock.
- El destino activo se identifica con superficie, texto y `aria-current`; nunca solo con color.
- El contenido reserva padding inferior suficiente para el dock.

### 5.4 Shell público

```text
┌──────────────────────────────┐
│ nombre de barbería           │
│ Reservas con NAVA            │
├──────────────────────────────┤
│ Paso n de m                  │
│ Título de la decisión        │
│ ayuda breve                  │
│                              │
│ opciones / formulario        │
│                              │
│ resumen compacto persistente │
├──────────────────────────────┤
│ Atrás             Continuar  │
└──────────────────────────────┘
```

- No tiene navegación del panel, publicidad ni solicitud de registro.
- El cuerpo no supera 720 px; el formulario de lectura larga no supera 640 px.
- El CTA principal permanece alcanzable, sin ocultar errores ni el último control.
- El progreso cuenta solo pasos reales. Si hay un único barbero, se omite esa decisión y se recalcula el total.

## 6. Componentes del sistema NAVA

Los nombres son responsabilidades de diseño, no obligación de crear un archivo por adelantado. Se construyen solo cuando un issue los usa realmente.

| Componente conceptual | Responsabilidad | Reglas clave |
| --- | --- | --- |
| `NavaWordmark` | Identidad de plataforma | Texto/SVG accesible; variantes tinta e invertida; sin imagen raster. |
| `PrivateAppShell` | Header, contenido y dock | Maneja safe areas, foco al navegar y contexto de sesión. |
| `PublicBookingShell` | Marca de barbería, progreso y acciones | Sin dependencias del bundle privado. |
| `PageHeader` | Título, ayuda y acción de pantalla | Un solo `h1`; la acción baja a ancho completo en móvil si hace falta. |
| `DesktopDock` / `MobileDock` | Navegación primaria | `aria-current`, targets de 44 px, sin overflow horizontal oculto. |
| `MetricRibbon` | Resumen operativo aprobado | Solo datos con contrato; label, valor, contexto y estado de carga. |
| `BarberSelector` | Selección obligatoria de barbero | Select/listbox accesible; no opción consolidada en P0. |
| `DateNavigator` | Anterior, fecha y siguiente | Tres controles estables; fecha completa anunciable; zona de barbería. |
| `AgendaTimeline` | Representación temporal escritorio | No es fuente de disponibilidad; tiene equivalente en lista. |
| `AppointmentSlip` | Turno compacto | Hora, persona, servicio, estado; foco y enlace único al detalle. |
| `NowMarker` | Hora actual | Etiqueta “Ahora”; decorativo respecto al estado del turno. |
| `NextTurnStrip` | Próximo turno accionable | Solo si hay dato; no inventa estado “en curso”. |
| `ContextPanel` | Resumen o acciones de la selección | No duplica toda la pantalla ni oculta acción crítica en móvil. |
| `AppointmentStatusBadge` | Estado P0 | Usa etiquetas de `estados-citas.md`, texto + icono/forma. |
| `ServiceOptionRow` | Selección de servicio | Nombre, duración, precio vigente y estado seleccionado. |
| `BarberOption` | Preferencia de barbero | Nombre real; monograma opcional; “Sin preferencia” solo si lo permite el flujo. |
| `DateStrip` | Selección de día público | Fecha legible; días no disponibles realmente disabled. |
| `TimeSlotGrid` | Franjas disponibles | Botones de mínimo 44 px; selección única; zona visible. |
| `BookingStepProgress` | Progreso público | Texto “Paso n de m” y nombre del paso; no depende de una barra visual. |
| `BookingSummary` | Selección persistente | Servicio, barbero, fecha, hora, duración y precio; permite editar el paso. |
| `BottomActionBar` | Acciones sticky | No tapa foco, errores o contenido; respeta safe area. |
| `BaseDialog` / `BaseBottomSheet` | Confirmación breve | Misma semántica; foco contenido y retorno al disparador. No depende de drag. |
| `AsyncState` | Carga, vacío, error o conflicto | Mensaje específico, acción posible y `request_id` solo en error inesperado. |

### 6.1 Turno como “ficha”

El `AppointmentSlip` usa el lenguaje de una ficha impresa:

1. hora y duración en cifras tabulares;
2. nombre de la persona atendida como texto principal;
3. servicio como texto secundario;
4. estado con badge completo;
5. barbero solo cuando aporta contexto; en la agenda P0 ya está definido por el selector;
6. indicador de conflicto o acción pendiente mediante texto e icono.

No muestra teléfono o correo en la lista. No contiene varios botones pequeños: la ficha abre el detalle; como máximo admite una acción rápida frecuente con nombre accesible y target completo.

## 7. Inventario completo de pantallas

### 7.1 Acceso y seguridad

| Pantalla / estado | Prioridad | Composición NAVA | Acción principal |
| --- | --- | --- | --- |
| Acceso de personal | `P0 existente` | Wordmark, título “Accede a NAVA”, teléfono/correo según contrato, contraseña y recuperación visible. Sin navegación privada. | Iniciar sesión |
| Desafío por abuso | `P0 existente` | Explica el paso sin revelar reglas de seguridad; control de verificación y alternativa accesible. | Verificar |
| Solicitud de recuperación | `P0 existente` | Un dato de contacto, mensaje neutro que no enumera cuentas. | Enviar código |
| Verificación de código | `P0 existente` | Destino enmascarado, código segmentado solo si pega/teclado funcionan, contador y reenvío. | Verificar código |
| Nueva contraseña | `P0 existente` | Reglas visibles antes de error, confirmación, opción mostrar/ocultar. | Guardar contraseña |
| Recuperación completada | `P0 existente` | Confirmación persistente y enlace de regreso. | Volver a acceder |
| Sesión vencida | `P0 existente` | Mensaje en contexto, preserva solo retorno seguro y nunca datos sensibles escritos. | Volver a iniciar sesión |

Reglas particulares:

- En 1024 px o más, el formulario ocupa una columna de 400–480 px y puede compartir el lienzo con una composición editorial vacía, nunca con funciones simuladas.
- En móvil, NAVA queda arriba, el formulario comienza sin hero que empuje la tarea fuera de pantalla.
- Los mensajes de seguridad son neutrales y no confirman si una cuenta existe.

### 7.2 Agenda diaria

**Prioridad:** `P0 existente`, sujeta a `DEC-074` y `DEC-075`.

Orden obligatorio:

1. `PageHeader`: “Agenda” + fecha legible + “Nuevo turno”.
2. `BarberSelector`: selección explícita; si solo existe uno, se muestra como valor fijo o preseleccionado sin opción “Todos”.
3. `DateNavigator`: anterior, selector de fecha, siguiente y retorno a “Hoy” cuando corresponda.
4. Resumen permitido: cantidad de turnos del día y próximo turno; solo si vienen del contrato o se derivan sin ambigüedad del resultado completo recibido.
5. Contenido cronológico.
6. Próximo turno o acciones contextuales cuando exista.

Escritorio:

- puede mostrar una línea temporal horizontal para **el barbero seleccionado**;
- el eje de tiempo tiene intervalos consistentes y se alinea con las fichas;
- los huecos son espacio, no tarjetas vacías repetidas;
- el marcador “Ahora” cruza el eje y lleva texto; no cambia `confirmed`;
- una lista cronológica accesible conserva la información equivalente y puede ser la misma estructura semántica con otra presentación CSS;
- el mosaico simultáneo de varios barberos queda reservado para `P2` o una decisión posterior.

Móvil:

- lista vertical por hora, sin exigir scroll horizontal;
- selector y fecha pueden ser sticky si no cubren el foco;
- cada ficha tiene al menos 64 px y ordena hora → persona → servicio → estado;
- “Nuevo turno” puede recibir fecha y hora prellenadas al iniciarse desde un hueco, pero el backend revalida todo.

Estados:

- sin barberos activos: explicación y enlace a Barberos si la persona tiene permiso;
- día sin turnos: “No hay turnos para esta fecha” + “Crear turno”;
- carga de otra fecha: conserva la fecha elegida, evita parpadeo completo y anuncia actualización;
- turno nocturno: aparece en cada día que intersecta y muestra fecha/hora de inicio/fin sin ambigüedad;
- error: conserva barbero y fecha, permite reintentar;
- estado terminal: sigue legible, con menor énfasis que `confirmed`, nunca oculto por defecto.

### 7.3 Nuevo turno manual

**Prioridad:** `P0 existente`.

Secciones y orden:

1. barbero;
2. servicio válido para ese barbero;
3. fecha, hora inicial y duración planificada;
4. cliente/persona atendida y contacto según `RN-CIT-02`;
5. resumen;
6. confirmación.

En escritorio se usa formulario de máximo 760 px con resumen lateral sticky solo si cabe. En móvil, secciones apiladas y resumen antes del CTA. Cambiar barbero invalida de forma explícita un servicio incompatible; no se borra silenciosamente. La hora manual no se fuerza a la rejilla pública. Los conflictos de jornada, bloqueo o cruce conservan datos y llevan el foco a fecha/hora.

Textos canónicos: “Nuevo turno”, “Persona atendida”, “Hora de inicio”, “Duración prevista”, “Crear turno”, “Creando…”. Nunca “Nueva cita”.

### 7.4 Detalle e historial del turno

**Prioridad:** `P0 existente`.

- Encabezado: fecha/hora, estado y acción contextual permitida.
- Resumen principal: persona atendida, servicio, barbero, duración y datos de contacto autorizados.
- Historial: línea temporal inmutable con evento, actor, fecha/hora y cambios comprensibles; paginación/carga adicional según contrato.
- Acciones: reprogramar, editar lo permitido, cancelar, completar, marcar no asistencia o corregir estado solo si la máquina y la HU lo autorizan.
- Acciones destructivas o terminales se confirman en diálogo/bottom sheet con consecuencia, motivo cuando aplique e idempotencia del servidor.
- La acción primaria cambia por estado; no se renderizan botones disabled para transiciones imposibles solo para “llenar” la pantalla.

### 7.5 Reprogramar y editar turno

| Pantalla | Prioridad | Regla de composición |
| --- | --- | --- |
| Reprogramar fecha/hora | `P0 existente` | Diálogo o página corta con intervalo vigente, nueva fecha/hora, fin calculado como vista previa, versión vigente y resumen. |
| Editar servicio/duración `T3` | `P0 pendiente` | Flujo dedicado cuando exista HU; compara “Antes / Después” y no reutiliza una simple edición de texto. |
| Cancelar por barbería | `P0 pendiente` | Confirmación con turno, consecuencia, motivo y notificación resultante. |
| Completar turno | `P0 pendiente` | Confirmación breve desde detalle/próximo turno; muestra resultado persistente. |
| Marcar no asistencia | `P0 pendiente` | Acción diferenciada de cancelar; no comparte color/etiqueta de peligro sin texto. |
| Corregir estado terminal | `P0 pendiente` | Vista de alto impacto con estado actual, destino permitido, motivo y advertencia de auditoría. |

Los estados pendientes no se implementan hasta tener historia, contrato e issue. Su presencia en esta tabla solo garantiza coherencia visual futura.

### 7.6 Barbería y equipo

| Pantalla | Prioridad | Composición | Acción principal |
| --- | --- | --- | --- |
| Configuración de barbería | `P0 existente` | Nombre, zona horaria y campos autorizados agrupados; una sección guarda por separado. | Guardar cambios |
| Lista de barberos | `P0 existente` | Filas sobrias con nombre, estado y servicios/horario como enlaces contextuales. | Añadir barbero |
| Alta de barbero | `P0 existente` | Formulario corto; identidad operativa, sin foto obligatoria. | Guardar barbero |
| Renombrar/editar barbero | `P0 existente` | Mismo patrón del alta; impacto visible si el contrato lo exige. | Guardar cambios |
| Desactivar barbero | `P0 pendiente` | No se diseña como delete; requiere HU sobre impacto y retención. | Según decisión futura |

La lista no muestra tarjetas con retratos ficticios. En escritorio puede usar filas de 64–72 px; en móvil, filas etiquetadas con menú de acciones. Servicios y horario abren su módulo correspondiente con barbero preseleccionado, sin duplicar formularios.

### 7.7 Servicios y asignaciones

| Pantalla / estado | Prioridad | Reglas NAVA |
| --- | --- | --- |
| Catálogo de servicios | `P0 existente` | Lista o tabla responsiva con nombre, duración, precio y estado; filtro activo/inactivo; acción “Nuevo servicio”. |
| Crear servicio | `P0 existente` | Formulario corto; nombre, duración planificada y precio según contrato. |
| Editar servicio | `P0 existente` | Muestra impacto de duración/precio según reglas; no promete cambiar turnos existentes. |
| Desactivar servicio | `P0 existente` | Previsualiza cantidad real de turnos afectados; consecuencia y confirmación explícitas. |
| Reactivar servicio | `P0 existente` | Confirmación ligera; resultado visible en la fila. |
| Servicios por barbero | `P0 existente` | Selector de barbero + lista de servicios con asignación; rechazo claro al intentar retirar la última asignación activa. |

No se usa un interruptor silencioso para desactivar. El precio conserva formato local, pero el almacenamiento y cálculo dependen del contrato. Una fila inactiva sigue legible y lleva la etiqueta “Inactivo”.

### 7.8 Horarios, excepciones y bloqueos

| Pantalla | Prioridad | Composición |
| --- | --- | --- |
| Horario semanal | `P0 existente` | Selector de barbero, siete días en orden local, segmentos por día, cerrado explícito y resumen de zona horaria. |
| Excepciones y festivos | `P0 existente` | Próximas excepciones por fecha, tipo y segmentos; crear/editar en formulario dedicado o panel. |
| Lista de bloqueos | `P0 existente` | Próximos y pasados diferenciados; fecha, intervalo, tipo, recurrencia y estado. |
| Crear bloqueo | `P0 existente` | Barbero, tipo, intervalo, recurrencia autorizada y resumen previo. |
| Editar serie/ocurrencia | `P0 existente` | Selector explícito de alcance solo donde el backend/UI lo soporte; la UI actual puede tener seguimiento parcial y no se simulan controles pendientes. |
| Turnos afectados por bloqueo | `P0 pendiente` | Primero confirma el bloqueo; luego lista cada turno y permite decisiones individuales. |

En móvil, la semana no se comprime en siete columnas ilegibles: cada día es una sección. En escritorio puede usar grid, manteniendo controles y labels reales. No se usa drag para mover segmentos. Crear un bloqueo no cancela ni reprograma turnos existentes en silencio.

### 7.9 Configuración operativa futura del MVP

Estas pantallas aparecen únicamente con sus HU y contratos:

- reglas de anticipación y ventana de reserva;
- intervalos/rejilla pública;
- política y límite de cancelación;
- recordatorios;
- canales de WhatsApp/correo y estado de configuración;
- cierre operativo y reglas relacionadas con estados, si están aprobadas.

Cada capacidad es una sección o subruta con guardado propio, descripción del efecto y valores de frontera. “Configuración” no es un formulario único que guarde todo. Los secretos y credenciales no se muestran ni se almacenan como campos ordinarios del frontend.

Pantallas complementarias:

- **Textos legales y privacidad (`P0 pendiente`)**: páginas públicas de lectura larga, con nombre de barbería/plataforma, fecha de vigencia, navegación de regreso y ancho de unas 72 caracteres. No se generan textos jurídicos desde el diseño ni se publican borradores sin revisión.
- **Estado de canales (`P0 pendiente`)**: puede mostrar canal configurado/no configurado y última validación cuando exista contrato. No revela tokens, credenciales ni respuesta cruda del proveedor.
- **Operación de notificaciones y retención (`sin pantalla P0`)**: workers, colas, leases y anonimización son capacidades backend/operativas. No se crea un dashboard administrativo salvo que una HU futura defina actor, permiso, datos y acciones.

### 7.10 Reserva pública

**Prioridad:** `P0 pendiente` (B4). No se implementa solo por existir este diseño.

Secuencia adaptable:

1. **Servicio** — “Elige un servicio”. Opción con nombre, duración y precio; solo servicios activos y reservables.
2. **Barbero** — “Elige un barbero”. Se omite si solo hay uno válido; “Sin preferencia” solo si el dominio lo permite.
3. **Fecha y hora** — “Elige fecha y hora”. Fechas disponibles primero, luego franjas; zona horaria de la barbería visible.
4. **Tus datos** — nombre, teléfono, correo y nota conforme al contrato; “Reservo para otra persona” revela solo el dato necesario.
5. **Revisión** — “Revisa tu turno”. Resumen completo, política relevante y enlaces “Editar” a cada paso.
6. **Resultado** — “Tu turno está confirmado”. Fecha, hora, servicio, barbero, instrucciones del enlace de gestión y siguiente acción.

Reglas:

- sin cuenta, contraseña, descarga de aplicación ni upsell;
- el paso conserva datos no sensibles al volver;
- un slot ocupado al confirmar produce un estado de conflicto dentro del paso de fecha/hora, conserva los datos y ofrece alternativas reales del backend;
- doble toque deja el CTA en carga y se apoya en idempotencia del servidor;
- una fecha sin disponibilidad no se presenta como seleccionable;
- precio y duración son datos, no controles decorativos;
- no se promete envío por un canal que el backend no haya confirmado.

Estados dedicados:

| Estado | Mensaje/acción esperada |
| --- | --- |
| Barbería o enlace no disponible | Explicación neutra; no filtra estado interno; permite volver al canal de origen. |
| Sin servicios reservables | “No hay servicios disponibles para reservar en línea”. |
| Sin fechas en ventana | “No encontramos horarios en estas fechas” + cambiar servicio/barbero cuando aplique. |
| Franja ocupada al confirmar | “Ese horario acaba de ocuparse” + alternativas cercanas reales. |
| Error recuperable | Conserva selecciones y ofrece “Reintentar”. |
| Confirmación | Resumen persistente y acceso de gestión; no depende solo de un toast. |

### 7.11 Gestión pública de un turno

**Prioridad:** `P0 pendiente`.

- La URL contiene un token opaco; nunca se expone en texto, logs o analytics.
- La pantalla representa un solo turno y no crea una cuenta de cliente.
- Muestra barbería, fecha/hora/zona, servicio, barbero, estado y política aplicable.
- La cancelación pública solo aparece dentro del límite vigente; exige confirmación y muestra el resultado persistente.
- Un enlace inválido, vencido, revocado o ya usado recibe un mensaje seguro sin distinguir detalles que faciliten enumeración.
- Un turno ya cancelado o terminal muestra su estado; no ofrece transiciones imposibles.

### 7.12 Pantallas diferidas y excluidas

| Capacidad | Estado | Reserva de diseño |
| --- | --- | --- |
| Notas internas del turno | `P1` | Sección privada del detalle, claramente separada de notas visibles al cliente. |
| Búsqueda/lista/detalle de clientes | `P1` | Patrón de lista y detalle; no se añade al dock P0. |
| Gestión de retraso | `P1` | Señal contextual en agenda y notificación; requiere reglas antes de UI. |
| Métricas operativas | `P1` | `MetricRibbon` solo con definiciones y contrato; sin ingresos. |
| Agenda semanal | `P2` | Vista temporal adicional; conserva lista accesible. |
| Repetir turno | `P2` | Acción desde un turno existente con revisión explícita. |
| Exportar agenda | `P2` | Acción secundaria; no módulo permanente. |
| Vista consolidada multi-barbero | `P2` | Puede recuperar el mosaico del concepto solo tras resolver interacción y alcance. |
| Clientes con cuenta / “Mis turnos” | `Excluido` | No se maqueta como si existiera. |
| Pagos, caja, ingresos, comisiones | `Excluido` | Sin navegación, métricas ni placeholders. |
| Inventario, nómina, contabilidad | `Excluido` | Sin navegación, métricas ni placeholders. |
| Marketplace, marketing, CRM, fidelización | `Excluido` | Sin navegación ni banners. |
| Multi-sede, modo oscuro, temas por tenant | `Excluido` | No se anticipan variantes en componentes P0. |

### 7.13 Estados globales y rutas de sistema

| Pantalla / región | Prioridad | Tratamiento NAVA |
| --- | --- | --- |
| Arranque/rehidratación de sesión | `P0 existente` | Shell estable o vista mínima con wordmark y progreso nombrado; no muestra un panel privado hasta validar sesión y tenant. |
| Ruta no encontrada | `P0 pendiente` | “No encontramos esta página”, explicación breve y acción al destino seguro; no inventa navegación pública/privada. |
| Sin permiso | `P0 pendiente` | “No tienes permiso para ver esta sección”; conserva privacidad y ofrece volver. No se presenta como 404 si el contrato exige otra conducta. |
| Sin conexión | `P0 pendiente` | Aviso persistente, contenido anterior solo si es seguro y acciones que requieren servidor bloqueadas con motivo; reintento al recuperar red. |
| Error inesperado | `P0 pendiente` | Mensaje seguro, `request_id` copiable y “Reintentar”/volver; nunca stack, proveedor o dato sensible. |
| Mantenimiento | `Excluido` | Solo se incorporará si backend/operación define una señal real; no se usa como error genérico conveniente. |

Estas vistas usan el mismo shell cuando la sesión/contexto ya es confiable. En el flujo público nunca revelan existencia de tenant, turno o cuenta más allá de lo autorizado. La aplicación no finge funcionamiento offline completo: solo preserva lectura o datos locales cuando sea seguro y avisa qué acción no se envió.

## 8. Comportamiento responsive por patrón

### 8.1 Reflow, no miniaturización

- 320 px debe conservar todo contenido y acción; no se reduce tipografía para hacer caber una tabla.
- Una tabla se transforma en filas etiquetadas. Un grid semanal se transforma en secciones por día.
- La agenda temporal tiene lista equivalente; el scroll horizontal nunca es el único camino para revisar turnos.
- Acciones secundarias pasan a menú o bloque inferior, manteniendo nombre accesible y orden.
- Un diálogo de decisión breve puede transformarse en bottom sheet; un flujo de varios campos se transforma en página.

### 8.2 Densidad

| Contexto | Densidad objetivo |
| --- | --- |
| Panel móvil | Filas de 56–72 px; texto 14–16 px; un dato principal y metadatos esenciales. |
| Panel escritorio | Grid de 12 columnas; separación de 24–32 px entre regiones; filas de 52–64 px. |
| Reserva pública móvil | Controles de 48 px; texto base 16 px; opciones de 64 px o más. |
| Formularios escritorio | 560–760 px; máximo dos campos por fila si siguen un orden lógico. |

### 8.3 Barras fijas

- El contenido agrega padding equivalente a la barra más safe area.
- Al enfocar un campo, `scroll-margin` evita que el teclado o la barra lo tape.
- La barra no oculta el resumen de errores.
- En zoom 200 %, deja de ser fija si el espacio restante resulta insuficiente.

## 9. Estados, interacción y accesibilidad

### 9.1 Estados asíncronos mínimos

Cada página y región que consulta datos define:

1. carga inicial con geometría estable;
2. actualización local conservando el contenido anterior cuando sea seguro;
3. vacío con causa y siguiente acción real;
4. error recuperable con reintento;
5. error no recuperable con salida segura;
6. conflicto de negocio con datos conservados;
7. éxito persistente, no solo toast;
8. permisos/sesión vencida sin filtrar información.

### 9.2 Foco y teclado

- `:focus-visible` usa el token de foco NAVA de 2 px con separación de 2 px.
- Navegar de página mueve foco al `h1` o al contenedor principal anunciado; no lo deja en un dock desmontado.
- Diálogos contienen foco, cierran con Escape cuando es seguro y lo devuelven al disparador.
- Fecha, hora, servicio y barbero se pueden operar con teclado sin simular controles incompletos.
- Ninguna acción necesaria aparece solo en hover.

### 9.3 Semántica

- Un `h1` por página y jerarquía sin saltos usados solo por tamaño.
- Listas, tablas y grids conservan roles nativos preferentemente; ARIA complementa, no reemplaza HTML.
- Estados y resultados se anuncian con regiones adecuadas sin repetir todo el contenido.
- El marcador “Ahora” no interrumpe lectura ni se anuncia como un cambio de estado del turno.
- Iconos decorativos quedan fuera del árbol accesible.

### 9.4 Movimiento

- Transiciones de 120 o 200 ms solo para continuidad, selección o entrada de una capa.
- `prefers-reduced-motion` elimina desplazamientos y animación continua.
- No hay parallax, rebotes, conteos animados ni introducciones que retrasen la tarea.

## 10. Contenido y vocabulario

### 10.1 Término canónico

La interfaz siempre usa **turno**. API, TypeScript generado y backend usan `appointment`. No se introducen “cita”, “booking” o “reserva” para nombrar el objeto ya creado. “Reserva” puede describir el proceso público: “Reserva tu turno”.

### 10.2 Voz

- Títulos: tarea concreta — “Elige un servicio”, “Revisa tu turno”.
- Botones: verbo + resultado — “Crear turno”, “Guardar horario”, “Cancelar turno”.
- Ayuda: una frase; explica efecto o formato, no repite el label.
- Error: qué ocurrió + qué puede hacer la persona.
- Éxito: qué cambió + siguiente paso.
- Confirmaciones: nombran el objeto; nunca “Aceptar” como acción destructiva.

### 10.3 Fechas, horas y dinero

- Se muestran en español y en la zona horaria de la barbería.
- Una experiencia usa una sola convención horaria coherente; las horas siempre incluyen minutos.
- Fechas importantes son legibles (“sábado, 24 de mayo”), con formato corto solo como apoyo.
- Dinero se formatea según la moneda definida por el contrato; no se concatena `$` ni se asumen pesos desde el componente.
- Las cifras tabulares no eliminan su label ni contexto.

## 11. Implementación incremental para agentes

### 11.1 Regla de entrega

Cada issue migra una preocupación completa y verificable. No se debe:

- cambiar solo colores de una pantalla sin cubrir estados, responsive y accesibilidad;
- migrar todo el frontend en una única rama;
- mezclar rediseño con cambio de contrato o regla de negocio sin que el issue lo autorice;
- crear todos los componentes conceptuales de §6 por adelantado;
- modificar snapshots para aceptar diferencias sin revisión visual;
- agregar datos, botones o rutas porque aparecen en el mockup.

### 11.2 Orden recomendado

1. Fundaciones: tokens, tipografía self-hosted, base CSS y primitivas afectadas.
2. Shell privado: wordmark, header, navegación y safe areas.
3. Acceso/recuperación y estados de sesión.
4. Agenda diaria, navegación de fecha, nuevo turno y detalle.
5. Configuración, barberos, servicios y asignaciones.
6. Horarios, excepciones y bloqueos.
7. Acciones restantes de B3 cuando existan sus HU.
8. Reserva y gestión pública cuando B4 tenga contrato e issue.
9. Capacidades P1/P2 solo tras priorización.

La secuencia puede cambiar por el backlog, pero una pantalla nunca depende de un componente NAVA inexistente sin incluirlo en su mismo issue o en una dependencia ya integrada.

### 11.3 Conservación de comportamiento

Antes de cambiar una pantalla existente, el agente debe inventariar:

- rutas y guardas;
- requests/responses tipados;
- estados de dominio y permisos;
- pruebas de componente y E2E;
- copy normativo;
- estados responsive y accesibles existentes.

El rediseño preserva esos contratos. Una diferencia funcional detectada se registra como duda/defecto y no se “arregla” dentro del cambio visual sin trazabilidad.

## 12. Pruebas y evidencia por pantalla

Toda pantalla NAVA modificada debe aportar:

- pruebas de componente para interacción, estado, foco y nombre accesible;
- `vitest-axe` sobre el estado normal y sobre diálogos/errores relevantes;
- E2E del recorrido P0 afectado cuando el comportamiento visible cambia;
- evidencia en 320, 360, 768 y 1280 px cuando cambia composición;
- zoom 200 %, teclado completo y `prefers-reduced-motion`;
- contraste de cada combinación nueva;
- textos largos: barbería, servicio y persona;
- datos límite reales del contrato: cero elementos, una opción, muchas opciones, turnos nocturnos y conflictos aplicables;
- prueba en teléfono real antes del piloto.

La evidencia debe incluir al menos carga, vacío, error, conflicto, éxito y disabled cuando esos estados existan. Una captura estática no sustituye la prueba de interacción.

## 13. Criterio de terminado de una pantalla NAVA

- [ ] La pantalla corresponde a una HU/issue real y no amplía alcance.
- [ ] Usa NAVA como plataforma y conserva el nombre de la barbería como contexto.
- [ ] Consume tokens y primitivas; no contiene hexadecimales, radios o sombras locales arbitrarios.
- [ ] Usa “turno” y copy orientado a tarea.
- [ ] Implementa todos los estados asíncronos aplicables.
- [ ] Mantiene una acción principal por región.
- [ ] Funciona en 320/360/768/1280 px y zoom 200 %.
- [ ] Controles táctiles alcanzan 44 × 44 px y las barras no ocultan foco o contenido.
- [ ] Teclado, foco, lector, reducción de movimiento y contraste cumplen WCAG 2.2 AA.
- [ ] No introduce reportes, inventario, pagos, métricas o estados no autorizados.
- [ ] Pruebas, evidencia y documentación se actualizaron.
- [ ] La implementación conserva API, reglas, auditoría, idempotencia y aislamiento existentes.

## 14. Pendientes deliberados

Esta especificación no decide:

- el registro jurídico o dominio de la marca NAVA;
- un logotipo gráfico adicional al wordmark tipográfico;
- carga de fotos de barberos;
- vista consolidada de varios barberos;
- fórmula o catálogo de métricas;
- personalización visual por tenant;
- comportamiento de capacidades sin HU o regla cerrada.

Esos asuntos requieren decisión o issue propio. Hasta entonces, el agente usa el patrón más simple permitido y no completa el vacío con supuestos.
