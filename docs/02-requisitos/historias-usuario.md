---
titulo: "Historias de usuario y criterios de aceptación"
version: "1.1"
estado: "Propuesta"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-10"
documentos_relacionados:
  - "../01-producto/alcance-mvp.md"
  - "../01-producto/reglas-negocio.md"
  - "estados-citas.md"
  - "../10-backlog/plan-bloques.md"
  - "../10-backlog/prompts-implementacion.md"
  - "../00-control/matriz-trazabilidad.md"
  - "../00-control/dudas-pendientes.md"
---

# Historias de usuario y criterios de aceptación

> **Estado del contenido: Propuesta.** Estas historias **derivan** de funciones P0 y reglas ya confirmadas; no crean, amplían ni reinterpretan alcance. Requieren aprobación del propietario antes de implementarse. Tres historias de B0 dependen además de dudas abiertas (`DP-SEG-04`, `DP-SEG-05`, `DP-SEG-06`) y no deben codificarse mientras no exista la decisión correspondiente. `HU-020` y `HU-021` se prepararon por solicitud del propietario, pero su implementación continúa bloqueada hasta que B0 cumpla su criterio de salida.

---

## 1. Cómo leer este documento

Cada historia lleva un código `HU-{nnn}` estable y sus criterios de aceptación llevan `CA-{HU}-{nn}`, según [glosario.md](../00-control/glosario.md), sección 8. **Los códigos no se renumeran ni se reutilizan.**

Una historia describe una capacidad verificable de principio a fin. No es una tarea técnica; una tarea que no se puede demostrar delante del propietario no es una historia.

Los criterios de aceptación son la **definición contractual de terminado**. Si un criterio no se puede comprobar con una prueba automatizada o con una evidencia concreta, está mal redactado.

La secuencia de bloques vive en [plan-bloques.md](../10-backlog/plan-bloques.md). El prompt de implementación de cada historia vive en [prompts-implementacion.md](../10-backlog/prompts-implementacion.md).

### Convención de las tablas de cabecera

| Campo | Significado |
| --- | --- |
| Función | Código `F-*` del alcance que esta historia hace realidad |
| Reglas | Reglas `RN-*` que la historia debe cumplir y que sus pruebas verifican |
| Decisiones | Decisiones `DEC-*` que la gobiernan |
| Depende de | Historias que deben estar terminadas antes |
| Bloquea | Historias que no pueden empezar sin esta |
| Riesgo | Qué se rompe si la historia queda mal hecha |

---

## 2. Estado del catálogo

| Bloque | Historias | Estado |
| --- | --- | --- |
| B0 · Cimientos, seguridad y primeras pantallas | `HU-001` – `HU-012` | Redactadas en este documento |
| B1 · Identidad de la barbería y catálogo | `HU-020` – `HU-021` | Redacción parcial propuesta; implementación bloqueada por B0 |
| B2 · Horario laboral y bloqueos | `HU-040` – | Pendientes |
| B3 · Agenda, estados e integridad | `HU-060` – | Pendientes |
| B4 · Reserva pública y disponibilidad | `HU-090` – | Pendientes |
| B5 · Notificaciones y recordatorios | `HU-130` – | Pendientes |
| B6 · Operación, privacidad y piloto | `HU-150` – | Pendientes |

---

## 3. Bloque B0 · Cimientos, seguridad y primeras pantallas

Orden de construcción recomendado: `HU-001` → `HU-002` → `HU-003` → `HU-004` → `HU-009` → `HU-005` → `HU-006` → `HU-010` → `HU-012` → `HU-007` → `HU-008` → `HU-011`.

El sistema visual (`HU-009`) se adelanta a las pantallas porque construir la pantalla de acceso sin componentes base produce estilos locales que después hay que desmontar.

---

### HU-001 · Esquema inicial con aislamiento por barbería

| Campo | Valor |
| --- | --- |
| Función | `F-SEG-01` |
| Reglas | `RN-TEN-01`, `RN-DIS-07` |
| Decisiones | `DEC-024`, `DEC-035`, `DEC-036` |
| Actor | Propietario (operación) |
| Depende de | — |
| Bloquea | `HU-002` y todas las posteriores |
| Riesgo | Un aislamiento débil expone datos de una barbería a otra. Es uno de los cuatro umbrales absolutos del piloto y no admite corrección posterior barata. |

**Historia**

> Como propietario del sistema, necesito que la base de datos nazca con `barbershop_id` obligatorio, RLS activa y un rol de aplicación sin privilegios de omisión, para que ninguna barbería pueda leer ni modificar datos de otra aunque la aplicación tenga un defecto.

**Alcance incluido**

- Migración inicial administrada con Atlas: extensiones necesarias, tabla `barbershop` y tabla de usuarios del área privada, ambas con clave primaria, `created_at`/`updated_at` con zona horaria y las restricciones de integridad correspondientes.
- Roles de base de datos separados: rol migrador (propietario de objetos) y rol de aplicación **sin** `BYPASSRLS` y sin propiedad sobre las tablas.
- `ENABLE ROW LEVEL SECURITY` y `FORCE ROW LEVEL SECURITY` en toda tabla con datos de una barbería, con política de lectura y escritura contra `current_setting('app.barbershop_id')`.
- Datos de prueba en `database/testdata/` con **dos** barberías y usuarios propios de cada una.
- Pruebas SQL en `database/tests/` que demuestran el aislamiento.

**Alcance excluido**

- Cualquier tabla de negocio: servicios, horarios, bloqueos, citas, clientes. Pertenecen a B1–B3.
- Sesiones, códigos de recuperación e idempotencia: cada una llega con su propia historia y su propia migración.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-001-01` | Dado el rol de aplicación y el contexto de la barbería A fijado en la transacción, cuando se consulta la tabla de usuarios, entonces solo se devuelven filas de A y ninguna de B. |
| `CA-001-02` | Dado el rol de aplicación con contexto de A, cuando se intenta insertar o actualizar una fila con `barbershop_id` de B, entonces la operación es rechazada por la política `WITH CHECK`. |
| `CA-001-03` | Dada una transacción sin `app.barbershop_id` fijado, cuando se consulta cualquier tabla protegida, entonces no se devuelve ninguna fila y la operación falla de forma explícita, nunca devolviendo el conjunto completo. |
| `CA-001-04` | Dado el rol de aplicación, cuando se consultan sus atributos en el catálogo de PostgreSQL, entonces no posee `BYPASSRLS` ni es propietario de las tablas. |
| `CA-001-05` | Dada una base de datos vacía, cuando se aplica la migración con Atlas, entonces el resultado es idéntico al de aplicarla dos veces sobre la misma base y `atlas.sum` queda versionado y válido. |
| `CA-001-06` | Toda columna de instante almacena zona horaria; ninguna columna temporal usa un tipo sin zona. |
| `CA-001-07` | Ninguna migración se ejecuta al arrancar los procesos `api` o `worker`. |

**Pruebas obligatorias**

- Integración con PostgreSQL real y dos barberías, conforme a `estrategia-pruebas.md` sección 4.4 y 6.
- Prueba negativa de escritura cruzada y prueba de ausencia de contexto.
- Validación Atlas: `migrate lint`, `migrate validate` y aplicación sobre base vacía.

**Terminado cuando** las pruebas anteriores pasan en CI, la matriz de trazabilidad registra `HU-001` en la fila de `DEC-024` y `database/README.md` describe cómo levantar la base local.

---

### HU-002 · Contexto de barbería en cada solicitud autenticada

| Campo | Valor |
| --- | --- |
| Función | `F-SEG-01` |
| Reglas | `RN-TEN-01` |
| Decisiones | `DEC-024`, `DEC-034` |
| Actor | Propietario (operación) |
| Depende de | `HU-001` |
| Bloquea | `HU-003`, `HU-005` |
| Riesgo | Una conexión reutilizada con el contexto de otra barbería produce exactamente la fuga que `HU-001` intentaba impedir. |

**Historia**

> Como propietario del sistema, necesito que toda operación de datos ocurra dentro de una transacción con el identificador de barbería fijado con alcance local, para que las políticas de la base de datos tengan siempre un contexto correcto y nunca uno heredado.

**Alcance incluido**

- Paquete de acceso a datos en `apps/api/internal/platform/database`: pool, apertura de transacción y fijación de `app.barbershop_id` con alcance de transacción.
- Un único punto autorizado para ejecutar trabajo tenant-aware; ninguna utilidad permite omitir el contexto.
- Comprobación de salud que verifica conectividad sin exponer credenciales ni versiones internas.
- Prueba con dos barberías ejecutadas de forma concurrente sobre el mismo pool.

**Alcance excluido**

- Autenticación y sesión (`HU-005`): en esta historia el identificador de barbería llega desde el llamador, y las pruebas lo inyectan directamente.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-002-01` | Dada una operación de datos, cuando se ejecuta, entonces siempre ocurre dentro de una transacción con `app.barbershop_id` fijado con alcance local a esa transacción. |
| `CA-002-02` | Dadas dos operaciones concurrentes de barberías distintas sobre el mismo pool, cuando ambas terminan, entonces ninguna observó filas de la otra. |
| `CA-002-03` | Al terminar la transacción, el valor de contexto no persiste en la conexión devuelta al pool, verificado leyendo el ajuste en una operación posterior. |
| `CA-002-04` | No existe ninguna función exportada que ejecute consultas de negocio sin recibir el contexto de barbería. |
| `CA-002-05` | Un fallo al fijar el contexto aborta la operación; nunca continúa con contexto vacío. |
| `CA-002-06` | El dominio y los servicios no importan el paquete de base de datos ni tipos del controlador de PostgreSQL. |

**Pruebas obligatorias**

- Unitarias de dirección de dependencias (el dominio no compila contra el paquete de datos).
- Integración concurrente con dos barberías y verificación de residuo de contexto.

**Terminado cuando** existe una prueba que falla si alguien agrega una consulta sin contexto, y `apps/api/README.md` documenta el patrón obligatorio.

---

### HU-003 · Contrato HTTP base con errores uniformes y registro sin datos personales

| Campo | Valor |
| --- | --- |
| Función | `F-OPS-01`, `F-OPS-04` |
| Reglas | `RN-DAT-02`, `RN-TEN-01` |
| Decisiones | `DEC-034`, `DEC-037` |
| Actor | Barbero (recibe mensajes claros), propietario (diagnostica) |
| Depende de | `HU-002` |
| Bloquea | `HU-004`, `HU-005`, `HU-007`, `HU-008` |
| Riesgo | Si cada endpoint inventa su formato de error, el frontend acumula casos especiales y el soporte pierde trazabilidad. Si un log lleva un teléfono, se incumple un umbral absoluto del piloto. |

**Historia**

> Como barbero, quiero que cuando algo falle reciba una explicación clara y consistente en vez de un error técnico, y como propietario necesito poder rastrear esa falla en los registros sin que allí aparezca ni un solo dato personal.

**Alcance incluido**

- Incorporación de Chi v5 y montaje de los tres grupos de audiencia: `/api/v1/public`, `/api/v1/customer`, `/api/v1/private`.
- Orden de middleware según `04-arquitectura/backend-go.md` sección 6.
- Respuesta de error uniforme RFC 9457 (`application/problem+json`) con `type`, `title`, `status`, `detail`, `instance` e identificador de correlación.
- Identificador de solicitud propagado a la respuesta y a cada línea de registro.
- Registro estructurado con nivel configurable, sin nombre, teléfono, correo, contraseña, token ni contenido de mensajes; los identificadores se registran como referencias opacas.
- Respuesta `404` —no `403`— ante un recurso de otra barbería.
- Los componentes de error, cabeceras y ejemplos correspondientes en `api/openapi/`, con `openapi:lint` y `openapi:bundle` en verde.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-003-01` | Toda respuesta de error del API usa `application/problem+json` con los campos obligatorios y un identificador de correlación presente también en la cabecera de respuesta. |
| `CA-003-02` | Ningún mensaje de error expuesto contiene nombres de excepciones, rutas de archivo, consultas SQL, nombres de proveedores ni versiones de dependencias. |
| `CA-003-03` | Dado un recurso existente de otra barbería, cuando se solicita con una sesión válida de la barbería propia, entonces la respuesta es `404` y el cuerpo no revela que el recurso existe. |
| `CA-003-04` | Ejecutado un recorrido completo de solicitudes válidas y fallidas, la revisión de los registros producidos no encuentra ningún dato personal ni secreto. |
| `CA-003-05` | Un pánico en un handler produce `500` con el formato uniforme, deja registro con el identificador de correlación y no derriba el proceso. |
| `CA-003-06` | El documento OpenAPI pasa lint y bundle, y cada respuesta de error declarada tiene ejemplo. |
| `CA-003-07` | Los módulos de dominio no importan Chi. |

**Pruebas obligatorias**

- Pruebas HTTP con `httptest` para middleware, formato de error, propagación del identificador y recuperación de pánico.
- Prueba de contrato: los handlers responden conforme al bundle OpenAPI.
- Prueba que inspecciona la salida del logger buscando patrones de correo, teléfono y token.

**Terminado cuando** el contrato y los handlers no se contradicen y existe una prueba que falla si un log emite un campo prohibido.

---

### HU-004 · Idempotencia reutilizable para operaciones críticas

| Campo | Valor |
| --- | --- |
| Función | `F-OPS-02`, `F-OPS-03` |
| Reglas | `RN-IDE-01` |
| Decisiones | `DEC-035`, `DEC-037` |
| Actor | Cliente y barbero (indirectamente) |
| Depende de | `HU-003` |
| Bloquea | `HU-060` y siguientes (creación de citas) |
| Riesgo | Sin este mecanismo listo antes de la primera escritura crítica, el doble toque produce dos turnos y el piloto pierde credibilidad en su primer día. |

**Historia**

> Como propietario del sistema, necesito un mecanismo único de idempotencia disponible antes de que exista la primera operación crítica, para que repetir una solicitud por doble toque, mala señal o reintento automático nunca produzca un segundo efecto.

**Alcance incluido**

- Migración con la tabla de claves de idempotencia (`idempotency_record`), con `barbershop_id`, clave, huella del contenido, resultado almacenado, estado y vencimiento; RLS activa igual que el resto.
- Middleware o servicio reutilizable que interpreta la cabecera `Idempotency-Key`.
- Semántica: misma clave y mismo contenido devuelve el resultado original sin repetir efectos; misma clave con contenido distinto responde `409`; solicitudes concurrentes con la misma clave no ejecutan el efecto dos veces.
- Declaración de la cabecera y sus respuestas en `api/openapi/`.
- El mecanismo se prueba con un handler de prueba dentro del paquete de pruebas; **no se publica ningún endpoint ficticio en el contrato**.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-004-01` | Dada una operación ejecutada con éxito con la clave K, cuando se repite con la misma clave y el mismo contenido, entonces se devuelve el resultado original sin ejecutar el efecto de nuevo. |
| `CA-004-02` | Dada la clave K ya usada, cuando llega otra solicitud con la misma clave y contenido distinto, entonces la respuesta es `409` y no se ejecuta ningún efecto. |
| `CA-004-03` | Dadas dos solicitudes simultáneas con la misma clave, cuando ambas terminan, entonces el efecto ocurrió exactamente una vez y ambas respuestas son coherentes entre sí. |
| `CA-004-04` | Los registros de idempotencia respetan el aislamiento por barbería: una clave de A no interfiere con la misma clave literal en B. |
| `CA-004-05` | Un registro vencido no revive una respuesta antigua; la operación se evalúa de nuevo. |
| `CA-004-06` | Una operación fallida no deja una clave que impida reintentar legítimamente. |

**Pruebas obligatorias**

- Integración con PostgreSQL: repetición exacta, repetición divergente, concurrencia real con `-race`, vencimiento y dos barberías.

**Terminado cuando** el mecanismo está documentado en `apps/api/README.md` como obligatorio para toda operación de escritura crítica futura.

---

### HU-005 · Inicio de sesión del barbero con correo y contraseña

| Campo | Valor |
| --- | --- |
| Función | `F-AUTH-01` |
| Reglas | `RN-TEN-01`, `RN-DAT-02` |
| Decisiones | `DEC-026` |
| Actor | Barbero |
| Depende de | `HU-002`, `HU-003`; **duda abierta `DP-SEG-04`** |
| Bloquea | `HU-006`, `HU-007`, `HU-010`, `HU-012` |
| Riesgo | Es la puerta de todo el área privada: un defecto aquí compromete todos los datos del sistema. |

**Historia**

> Como barbero, quiero entrar a mi agenda con mi correo y mi contraseña, para empezar a trabajar sin recordar códigos ni instalar nada.

> **Bloqueo declarado:** `DEC-026` confirma "correo, contraseña y sesión larga", pero no fija el mecanismo de sesión ni su duración exacta. Ese punto está registrado como `DP-SEG-04` y debe resolverse con un `DEC-*` **antes** de escribir código. La historia se implementa contra la decisión que resuelva la duda; el resto de sus criterios no depende de ella.

**Alcance incluido**

- Migración con la tabla de credenciales y la de sesiones, ambas con `barbershop_id` y RLS.
- Almacenamiento de contraseña con función de derivación resistente a fuerza bruta y parámetros documentados; nunca cifrado reversible ni hash simple.
- Operación de inicio de sesión en `/api/v1/private` declarada primero en OpenAPI.
- Respuesta uniforme ante credenciales inválidas y ante correo inexistente: **el mismo mensaje y el mismo tiempo de respuesta**, para no revelar qué correos existen.
- Registro del intento sin exponer el correo en claro en los logs.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-005-01` | Dadas credenciales válidas, cuando el barbero inicia sesión, entonces obtiene una sesión asociada a su barbería y a su usuario, y las solicitudes privadas posteriores operan con el contexto de esa barbería. |
| `CA-005-02` | Dado un correo inexistente y dado un correo existente con contraseña incorrecta, entonces ambas respuestas son indistinguibles en cuerpo, código y tiempo perceptible. |
| `CA-005-03` | La contraseña se almacena mediante derivación de clave con sal única por usuario; la base de datos no contiene ninguna contraseña legible ni reversible. |
| `CA-005-04` | Ni la contraseña, ni el correo, ni el material de sesión aparecen en registros, mensajes de error o respuestas. |
| `CA-005-05` | Una sesión emitida para la barbería A no habilita ninguna operación sobre datos de B, verificado contra un endpoint privado real. |
| `CA-005-06` | La operación está declarada en OpenAPI con su esquema, sus errores y sus ejemplos antes de existir el handler. |
| `CA-005-07` | Un usuario desactivado o eliminado no puede iniciar sesión ni conservar sesiones vigentes. |

**Pruebas obligatorias**

- Unitarias del servicio de autenticación (verificación, normalización del correo, usuario inactivo).
- Integración HTTP: éxito, credenciales inválidas, correo inexistente, aislamiento entre barberías.
- Prueba de que el hash almacenado cambia con la misma contraseña en dos usuarios distintos.

**Terminado cuando** existe el `DEC-*` que resuelve `DP-SEG-04`, el contrato está publicado y todas las pruebas anteriores pasan.

---

### HU-006 · Sesión persistente, cierre de sesión y protección del área privada

| Campo | Valor |
| --- | --- |
| Función | `F-AUTH-01` |
| Reglas | `RN-TEN-01` |
| Decisiones | `DEC-026` |
| Actor | Barbero |
| Depende de | `HU-005` |
| Bloquea | `HU-012` |
| Riesgo | Una sesión que no se puede revocar convierte cualquier dispositivo perdido en un acceso permanente a los datos de los clientes. |

**Historia**

> Como barbero, quiero seguir dentro de la aplicación al día siguiente sin volver a escribir mi contraseña, y quiero poder cerrar sesión y que ese cierre sea inmediato y real.

**Alcance incluido**

- Vigencia larga de la sesión conforme al `DEC-*` que resuelva `DP-SEG-04`, con renovación controlada.
- Cierre de sesión que invalida la sesión en el servidor, no solo en el navegador.
- Middleware de sesión que protege todo `/api/v1/private` y responde de forma uniforme cuando falta o vence.
- Registro del instante de último uso para poder auditar accesos.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-006-01` | Dada una sesión iniciada, cuando el navegador se cierra y se vuelve a abrir dentro de la vigencia, entonces el barbero sigue autenticado sin volver a escribir la contraseña. |
| `CA-006-02` | Dado un cierre de sesión, cuando se reutiliza el material de sesión anterior, entonces la respuesta es de no autorizado y ninguna operación se ejecuta. |
| `CA-006-03` | Una sesión vencida no se renueva sola; exige un nuevo inicio de sesión. |
| `CA-006-04` | Toda ruta bajo `/api/v1/private` exige sesión válida; una ruta nueva sin protección explícita hace fallar una prueba. |
| `CA-006-05` | El material de sesión no es adivinable, no se registra y no viaja en la URL. |
| `CA-006-06` | Cerrar sesión en un dispositivo no invalida las sesiones de otros dispositivos, salvo que el barbero lo solicite explícitamente si esa opción se implementa. |

**Pruebas obligatorias**

- Integración HTTP: persistencia, revocación, vencimiento, ruta privada sin sesión.
- Prueba estructural que enumera las rutas privadas registradas y verifica que todas pasan por el middleware de sesión.

**Terminado cuando** el cierre de sesión es verificable desde el servidor y la prueba estructural protege contra rutas privadas olvidadas.

---

### HU-007 · Defensa escalonada contra abuso en el acceso

| Campo | Valor |
| --- | --- |
| Función | `F-SEG-03` |
| Reglas | `RN-DAT-02` |
| Decisiones | `DEC-026` |
| Actor | Propietario (protege), barbero (afectado si se excede) |
| Depende de | `HU-003`, `HU-005`; **duda abierta `DP-SEG-06`** |
| Bloquea | — |
| Riesgo | Sin límite, un ataque automatizado prueba miles de contraseñas; con un límite mal calibrado, el barbero legítimo queda fuera en plena jornada. |

**Historia**

> Como propietario del sistema, necesito que el formulario de acceso limite los intentos por IP y exija una prueba adicional cuando se supera el umbral, para frenar el abuso sin castigar al barbero que se equivocó dos veces.

> **Bloqueo declarado:** `DEC-026` fija el umbral inicial de **5 solicitudes por IP**, pero no la duración de la ventana ni la del escalamiento. Registrado como `DP-SEG-06`.

**Alcance incluido**

- Conteo por IP con ventana configurable y umbral configurable, con valor inicial 5.
- Escalamiento: superado el umbral, la solicitud exige verificación telefónica antes de evaluar la contraseña.
- Respuesta `429` con formato uniforme y con indicación de cuándo reintentar.
- Configuración expuesta como parámetros, no como números incrustados en el código.
- Los contadores no almacenan datos personales; la IP se guarda de forma acotada y con vencimiento.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-007-01` | Dadas N solicitudes desde la misma IP por debajo del umbral, entonces todas se evalúan normalmente. |
| `CA-007-02` | Superado el umbral dentro de la ventana, la siguiente solicitud desde esa IP exige verificación telefónica y no evalúa la contraseña. |
| `CA-007-03` | La respuesta al superar el umbral usa el formato uniforme, indica cuándo reintentar y no revela si el correo existe. |
| `CA-007-04` | Transcurrida la ventana sin nuevos intentos, el conteo se reinicia y el acceso normal se restablece. |
| `CA-007-05` | El umbral y la ventana se cambian por configuración, sin recompilar ni editar código. |
| `CA-007-06` | Los datos de conteo vencen solos y no contienen correo, nombre ni teléfono. |
| `CA-007-07` | El límite no depende de una cabecera que el cliente pueda falsificar libremente; la obtención de la IP está documentada y probada según el despliegue previsto. |

**Pruebas obligatorias**

- Integración: recorrido del umbral, escalamiento, expiración de la ventana, dos IP distintas independientes.
- Prueba de que una cabecera manipulada no evade el límite en la configuración de despliegue documentada.

**Terminado cuando** existe el `DEC-*` que resuelve `DP-SEG-06` y el escalamiento queda demostrado de extremo a extremo.

---

### HU-008 · Recuperación de acceso con código al teléfono verificado

| Campo | Valor |
| --- | --- |
| Función | `F-AUTH-02` |
| Reglas | `RN-DAT-01`, `RN-DAT-02` |
| Decisiones | `DEC-026` |
| Actor | Barbero |
| Depende de | `HU-005`, `HU-007`; **duda abierta `DP-SEG-05`** |
| Bloquea | `HU-011` |
| Riesgo | Un barbero sin acceso en plena jornada pierde su agenda; un mecanismo de recuperación débil es la vía más común de secuestro de cuentas. |

**Historia**

> Como barbero que olvidó su contraseña, quiero recuperar el acceso con un código enviado a mi teléfono verificado, para volver a mi agenda el mismo día sin depender de que alguien me responda.

> **Bloqueo declarado:** `DEC-026` define el mecanismo (código al teléfono verificado) pero no el canal ni el proveedor. `DEC-027` habilita correo y WhatsApp oficial **para notificaciones de turnos**, no para códigos de seguridad. Registrado como `DP-SEG-05`.

**Alcance incluido**

- Migración con la tabla de códigos de recuperación: hash del código, vencimiento corto, intentos, marca de uso, `barbershop_id` y RLS.
- Solicitud de recuperación, verificación del código y establecimiento de contraseña nueva.
- Código de un solo uso, con vencimiento, con límite de intentos y con reenvío controlado.
- Invalidación de las sesiones activas al cambiar la contraseña.
- Destino mostrado enmascarado en la interfaz y en las respuestas.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-008-01` | Dado un correo registrado, cuando se solicita recuperación, entonces se envía un código al teléfono verificado y la respuesta es idéntica a la de un correo no registrado. |
| `CA-008-02` | El código vence en un plazo corto configurable, se acepta una sola vez y queda inválido tras usarse. |
| `CA-008-03` | Superado el número de intentos fallidos, el código se invalida por completo y debe solicitarse uno nuevo. |
| `CA-008-04` | El código se almacena como hash; la base de datos no contiene el valor enviado. |
| `CA-008-05` | Al establecer la contraseña nueva, todas las sesiones activas del usuario quedan invalidadas. |
| `CA-008-06` | El teléfono se muestra enmascarado y nunca aparece completo en respuestas ni registros. |
| `CA-008-07` | El reenvío tiene límite propio y no reinicia el vencimiento del código anterior sin invalidarlo. |
| `CA-008-08` | La contraseña nueva se rechaza si no cumple la política mínima documentada, con un mensaje que explica qué falta. |

**Pruebas obligatorias**

- Integración: recorrido completo, código vencido, código ya usado, exceso de intentos, correo inexistente, invalidación de sesiones.
- Prueba del adaptador de envío con doble de prueba; el proveedor real no se invoca en pruebas.

**Terminado cuando** existe el `DEC-*` que resuelve `DP-SEG-05` y el recorrido completo funciona con el adaptador seleccionado.

---

### HU-009 · Sistema visual base implementado en componentes

| Campo | Valor |
| --- | --- |
| Función | Soporte transversal de todas las pantallas P0 |
| Reglas | Criterios no funcionales de UX y accesibilidad |
| Decisiones | `DEC-039`, `DEC-033`, `DEC-035` |
| Actor | Barbero y cliente (indirectamente) |
| Depende de | — |
| Bloquea | `HU-010`, `HU-011`, `HU-012` y toda pantalla posterior |
| Riesgo | Sin componentes base, cada pantalla inventa su propio botón; el resultado es una interfaz incoherente y un rediseño caro a mitad del proyecto. |

**Historia**

> Como barbero que trabaja de pie y con las manos ocupadas, quiero que todos los controles de la aplicación se vean y se comporten igual, con áreas táctiles cómodas y textos legibles, para no equivocarme entre cliente y cliente.

**Alcance incluido**

- Componentes base en `apps/web/src/shared/ui/` que realmente usarán las pantallas de B0: botón, campo de texto, alerta, insignia y diálogo. **No se crean componentes sin uso real.**
- Variantes por intención (`tone`), nunca por valor libre; todo color, medida, radio y tipografía proviene de `tokens.css`.
- Patrones de estado de pantalla: inicial, carga, actualización, vacío, error recuperable, error de campo, conflicto y éxito.
- Foco visible, orden de tabulación coherente, área táctil de 44 × 44 px y respeto de `prefers-reduced-motion`.
- Evidencia responsive en 320, 360, 768 y 1280 px.

**Alcance excluido**

- Modo oscuro, selector de tema, colores por barbería, CSS libre y biblioteca visual externa: prohibidos por `DEC-039`.
- Iconografía como dependencia nueva sin justificación.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-009-01` | Ningún componente contiene un color hexadecimal, un tamaño en píxeles arbitrario ni una fuente literal; todos consumen tokens semánticos. |
| `CA-009-02` | Cada componente interactivo tiene foco visible conforme al estándar y es operable solo con teclado. |
| `CA-009-03` | Los controles táctiles ordinarios miden al menos 44 × 44 px. |
| `CA-009-04` | Cada componente tiene prueba de componente que cubre sus estados: normal, deshabilitado, cargando y error cuando apliquen. |
| `CA-009-05` | El contraste de texto y de bordes cumple WCAG 2.2 AA, verificado con la herramienta definida en la estrategia de pruebas. |
| `CA-009-06` | Existe evidencia visual en los cuatro anchos de referencia y ninguna pantalla produce desplazamiento horizontal. |
| `CA-009-07` | Las animaciones, si existen, duran entre 120 y 200 ms y desaparecen con `prefers-reduced-motion`. |

**Pruebas obligatorias**

- Pruebas de componente con la herramienta definida en `estrategia-pruebas.md` sección 5.2.
- Verificación accesible automatizada y revisión manual con teclado.

**Terminado cuando** las pantallas de `HU-010` a `HU-012` se construyen sin agregar un solo estilo local nuevo.

---

### HU-010 · Pantalla de acceso

| Campo | Valor |
| --- | --- |
| Función | `F-AUTH-01` |
| Reglas | Criterios no funcionales de UX; `RN-DAT-02` |
| Decisiones | `DEC-039`, `DEC-033` |
| Actor | Barbero |
| Depende de | `HU-005`, `HU-009` |
| Bloquea | `HU-012` |
| Riesgo | Es la primera pantalla que el barbero ve del producto. Un formulario confuso o lento en un celular de gama media es una mala primera impresión difícil de revertir. |

**Historia**

> Como barbero, quiero una pantalla de acceso simple, con mi correo y mi contraseña visibles en un solo lugar y con la recuperación a la vista, para entrar rápido incluso con una conexión mala.

**Alcance incluido**

- Composición según `estandar-diseno-visual.md` sección 10: marca discreta, título, formulario estrecho y recuperación visible.
- Cliente tipado generado desde el bundle OpenAPI; sin DTO manuales paralelos.
- Estados: inicial, enviando, error de credenciales, error de red con datos conservados, y bloqueo por umbral de `HU-007` con explicación.
- Etiquetas asociadas, mensajes de error junto al campo y resumen accesible cuando haya varios.
- Ruta cargada de forma diferida.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-010-01` | Con credenciales válidas, el barbero entra y llega al panel privado. |
| `CA-010-02` | Con credenciales inválidas, el mensaje explica qué hacer, conserva el correo escrito y no indica si el correo existe. |
| `CA-010-03` | Ante un fallo de red, la pantalla conserva lo escrito y ofrece reintentar sin recargar. |
| `CA-010-04` | El botón de envío se deshabilita mientras la solicitud está en curso, y un doble toque no produce dos solicitudes. |
| `CA-010-05` | La pantalla es operable solo con teclado, con orden de tabulación coherente y foco visible. |
| `CA-010-06` | La pantalla es utilizable con una sola mano en 320 y 360 px, sin desplazamiento horizontal. |
| `CA-010-07` | La contraseña nunca se registra, ni se envía en la URL, ni queda en el historial de navegación. |
| `CA-010-08` | Existe un enlace visible a la recuperación de acceso. |

**Pruebas obligatorias**

- Prueba de componente del formulario (validación, estados, accesibilidad).
- E2E del recorrido de acceso, incluido el caso de credenciales inválidas.
- Evidencia responsive en los cuatro anchos.

**Terminado cuando** el recorrido completo de acceso funciona contra el API real en local y las evidencias están adjuntas al pull request.

---

### HU-011 · Pantalla de recuperación de acceso

| Campo | Valor |
| --- | --- |
| Función | `F-AUTH-02` |
| Reglas | Criterios no funcionales de UX; `RN-DAT-01`, `RN-DAT-02` |
| Decisiones | `DEC-026`, `DEC-039` |
| Actor | Barbero |
| Depende de | `HU-008`, `HU-009`, `HU-010` |
| Bloquea | — |
| Riesgo | Un flujo de recuperación confuso termina en una llamada al propietario; es exactamente el soporte que el MVP quiere evitar. |

**Historia**

> Como barbero que perdió su contraseña, quiero un flujo de pocos pasos que me diga en qué paso estoy, a dónde llegó el código y qué hacer si no llega, para recuperar mi acceso sin llamar a nadie.

**Alcance incluido**

- Composición según `estandar-diseno-visual.md` sección 10: paso actual, destino enmascarado, campo del código y reenvío con estado.
- Tres pasos: solicitar, verificar el código, establecer la contraseña nueva.
- Estados de error específicos: código incorrecto, código vencido, demasiados intentos y espera de reenvío con cuenta regresiva.
- Confirmación explícita de que las sesiones activas se cerraron.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-011-01` | El flujo muestra siempre el paso actual y el total de pasos. |
| `CA-011-02` | El destino del código se muestra enmascarado y nunca completo. |
| `CA-011-03` | Un código incorrecto, vencido o agotado produce mensajes distintos y accionables, sin lenguaje técnico. |
| `CA-011-04` | El reenvío indica cuánto falta para poder volver a solicitarlo y no permite pedirlo antes. |
| `CA-011-05` | Al terminar, la pantalla confirma el cambio, informa que las sesiones se cerraron y lleva al acceso. |
| `CA-011-06` | La política de contraseña se explica **antes** de que el barbero escriba, no solo al fallar. |
| `CA-011-07` | Recorrido operable con teclado, con foco movido al encabezado de cada paso y anunciado por lector de pantalla. |
| `CA-011-08` | Utilizable en 320 y 360 px sin desplazamiento horizontal. |

**Pruebas obligatorias**

- Pruebas de componente por paso.
- E2E del recorrido completo con código válido y con código vencido.
- Evidencia responsive y verificación accesible.

**Terminado cuando** un barbero puede recuperar el acceso de principio a fin sin intervención del propietario.

---

### HU-012 · Cascarón del panel privado

| Campo | Valor |
| --- | --- |
| Función | `F-OPS-01` (comportamiento ante fallos y conexión inestable) |
| Reglas | `RN-TEN-01`, criterios no funcionales de UX |
| Decisiones | `DEC-033`, `DEC-039` |
| Actor | Barbero |
| Depende de | `HU-006`, `HU-009`, `HU-010` |
| Bloquea | Todas las pantallas privadas de B1 en adelante |
| Riesgo | Si cada pantalla resuelve por su cuenta la sesión, la navegación y los errores, el mismo problema se implementa siete veces de siete maneras. |

**Historia**

> Como barbero, quiero que la aplicación recuerde mi sesión, me lleve a mi agenda al entrar, me muestre siempre en qué barbería estoy y me avise con claridad cuando se pierda la conexión, para no quedarme mirando una pantalla en blanco.

**Alcance incluido**

- Estructura de aplicación autenticada: cabecera con el nombre de la barbería, navegación principal y área de contenido, según las estructuras base del estándar visual.
- Guarda de ruta: sin sesión válida se redirige al acceso, conservando el destino pretendido.
- Manejo central de respuestas no autorizadas: la sesión se limpia y se informa el motivo, sin bucles de redirección.
- Patrones globales de estado: carga con esqueleto, error recuperable con "Reintentar" y aviso de conexión perdida.
- Una pantalla de inicio provisional que muestra el estado de la sesión; **no** es la agenda, que llega en B3.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-012-01` | Con sesión válida, al abrir la aplicación se entra directamente al panel sin pedir credenciales. |
| `CA-012-02` | Sin sesión, cualquier ruta privada redirige al acceso y, tras entrar, lleva al destino que se pretendía abrir. |
| `CA-012-03` | Ante una respuesta de no autorizado, la sesión local se limpia una sola vez y no se produce un bucle de redirección. |
| `CA-012-04` | La cabecera muestra siempre la barbería activa; el barbero nunca puede dudar de en qué contexto está operando. |
| `CA-012-05` | Una pérdida de conexión muestra un aviso comprensible con acción de reintento, sin perder el estado de la pantalla. |
| `CA-012-06` | Cada ruta se carga de forma diferida y la navegación no recarga la aplicación completa. |
| `CA-012-07` | Cerrar sesión desde la cabecera invalida la sesión en el servidor y devuelve al acceso. |
| `CA-012-08` | La estructura cumple los cuatro anchos de referencia y la navegación es operable con teclado. |

**Pruebas obligatorias**

- Pruebas de componente de la guarda de ruta y del manejo de no autorizado.
- E2E: acceso, navegación, recarga con sesión persistente y cierre de sesión.
- Evidencia responsive.

**Terminado cuando** una pantalla nueva se puede agregar declarando solo su ruta y su contenido, sin repetir lógica de sesión ni de errores.

---

## 4. Bloque B1 · Identidad de la barbería y catálogo

> **Apertura anticipada de documentación.** Estas dos historias se redactan antes del cierre de B0 por solicitud expresa del propietario. Esto no cambia la secuencia de construcción: ningún prompt de B1 se ejecuta mientras B0 no cumpla su criterio de salida.

Orden de construcción recomendado para esta parte del bloque: `HU-020` → `HU-021`. Las historias de servicios que completarán B1 se redactan al abrir formalmente el bloque.

---

### HU-020 · Configuración básica de la barbería

| Campo | Valor |
| --- | --- |
| Función | `F-CONF-01` |
| Reglas | `RN-DIS-07`, `RN-TEN-01` |
| Decisiones | `DEC-007`, `DEC-024`, `DEC-033`, `DEC-037`, `DEC-039` |
| Actor | Barbero autenticado |
| Depende de | Criterio de salida de B0; en particular `HU-003`, `HU-006`, `HU-009` y `HU-012` |
| Bloquea | `HU-021` y las pantallas de B1 que necesitan identificar la barbería activa |
| Riesgo | Una zona inválida o tomada del dispositivo desplaza todas las horas; una actualización sin aislamiento puede modificar la identidad o el contacto de otra barbería. |

**Historia**

> Como barbero, quiero consultar y actualizar el nombre, la zona horaria y los datos de contacto de mi barbería desde una sección propia, para que el sistema identifique correctamente el negocio y muestre todas las horas en su zona real.

**Alcance incluido**

- Lectura y actualización autenticadas de la fila `barbershop` correspondiente a la sesión; el cliente nunca envía un `barbershopId` confiable.
- Nombre obligatorio, zona horaria IANA obligatoria, correo de contacto opcional y teléfono de contacto opcional. El correo se normaliza en minúsculas y el teléfono, cuando exista, usa E.164.
- Validación de la zona contra el catálogo IANA disponible en PostgreSQL; no se acepta una abreviatura como `COT`, un desplazamiento como `UTC-5` ni la zona del navegador como fuente de verdad.
- Sección “Barbería” dentro del módulo de configuración, con guardado independiente de las secciones futuras de reserva, cancelación, recordatorios y canales.
- Contrato OpenAPI privado, caso de uso Go, persistencia PostgreSQL, cliente TypeScript derivado del bundle y pantalla Vue responsive.
- Actualización inmediata del nombre de la barbería en el cascarón privado después de guardar con éxito, sin recargar la aplicación completa.

**Alcance excluido**

- Alta o eliminación de barberías; el aprovisionamiento sigue siendo administrativo.
- Enlace público o `public_slug`, que pertenece a `F-PUB-01` en B4.
- Políticas de cancelación, límites de reserva, recordatorios, canales y proveedores.
- Conversión o reescritura masiva de instantes existentes al cambiar la zona: un instante almacenado nunca se modifica por una preferencia de presentación.
- Personalización de colores, tipografías o tema por barbería, prohibida por `DEC-039`.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-020-01` | Dada una sesión válida, cuando se abre la sección “Barbería”, entonces se muestran el nombre, la zona IANA y los datos de contacto de la barbería activa, y ningún dato de otra barbería. |
| `CA-020-02` | Dado un nombre no vacío y una zona IANA válida, cuando se guarda la sección, entonces la actualización persiste y el cascarón privado refleja el nuevo nombre sin recargar la aplicación. |
| `CA-020-03` | Dada una abreviatura, un desplazamiento UTC o un identificador inexistente, cuando se intenta guardar como zona, entonces la operación responde `422` con un error de campo comprensible y no modifica ningún dato. |
| `CA-020-04` | Dado un instante conocido y dos dispositivos en zonas distintas, cuando ambos consultan la barbería, entonces la hora resultante se calcula con la zona configurada de la barbería y coincide en ambos dispositivos. |
| `CA-020-05` | Dado un identificador de barbería ajeno incluido en URL, body o parámetro manipulado, cuando se intenta leer o actualizar la configuración, entonces se ignora como fuente de tenant o se responde `404`, y la barbería ajena permanece sin cambios. |
| `CA-020-06` | Cuando correo o teléfono están vacíos se almacenan como ausencia de valor; cuando están presentes se normalizan y validan, y nunca aparecen en registros técnicos. |
| `CA-020-07` | El contrato OpenAPI, el handler, el cliente tipado y el formulario coinciden en campos obligatorios, opcionales, límites y respuestas de error. |
| `CA-020-08` | La sección conserva los datos no sensibles ante un error recuperable, anuncia carga, error y éxito de forma accesible, funciona con teclado y hace reflow sin pérdida a 320, 360, 768 y 1280 px. |

**Pruebas obligatorias**

- Unitarias del caso de uso y de la validación de zona IANA, incluido un instante evaluado con zonas de dispositivo distintas.
- HTTP y contrato para lectura, actualización, `422`, no autenticado y recurso de otro tenant.
- Integración con PostgreSQL real y dos barberías, cubriendo RLS en lectura y actualización.
- Prueba de componente del formulario en carga, error de campo, error recuperable y éxito.
- E2E del recorrido privado abrir → editar → guardar → comprobar cabecera, más evidencia responsive y accesible.

**Terminado cuando** un barbero actualiza la identidad básica de su barbería de extremo a extremo, todas las pruebas pasan y ningún dato de contacto llega a los registros técnicos.

---

### HU-021 · Registro y listado de barberos

| Campo | Valor |
| --- | --- |
| Función | `F-CONF-02` (registro y listado; servicios y horario se completan en historias posteriores) |
| Reglas | `RN-TEN-01` |
| Decisiones | `DEC-019`, `DEC-024`, `DEC-033`, `DEC-037`, `DEC-039` |
| Actor | Barbero autenticado |
| Depende de | `HU-020` y criterio de salida de B0 |
| Bloquea | Historias de servicios por barbero, B2 y la selección pública de barbero en B4 |
| Riesgo | Modelar al barbero como un caso especial de la barbería unipersonal obliga a cambiar datos, API y agenda al incorporar el segundo barbero; omitir el tenant permite relaciones cruzadas. |

**Historia**

> Como barbero, quiero registrar, renombrar y consultar las personas que prestan servicios en mi barbería, para que el mismo sistema represente sin casos especiales tanto a un independiente como a un equipo de varios barberos.

**Alcance incluido**

- Entidad `barber` distinta de `barbershop` y de la cuenta `staff_user`, con identificador propio, `barbershop_id`, nombre visible y marcas de tiempo con zona.
- Lista privada de barberos de la barbería activa, alta de un barbero y edición de su nombre.
- La barbería con una sola persona conserva una fila `barber`; no se infiere el barbero desde `barbershop` ni desde la sesión.
- Pantalla “Barberos” en el módulo `staff`, con lista, estado vacío, formulario accesible para añadir y edición del nombre.
- Contrato OpenAPI privado, módulo Go `staff`, migración Atlas, RLS, cliente tipado y componentes Vue.

**Alcance excluido**

- Eliminar, desactivar o reactivar barberos: el efecto sobre citas futuras no está definido en el alcance confirmado y no se inventa en esta historia.
- Crear credenciales, invitar usuarios, asignar roles o enlazar automáticamente `staff_user` con `barber`.
- Asignar servicios, configurar horarios o calcular disponibilidad; llegan en las historias posteriores de B1 y B2.
- Selección pública de barbero y agenda por barbero, que pertenecen a B4 y B3 respectivamente.
- Datos personales adicionales al nombre necesario para mostrar al profesional.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-021-01` | Dada una barbería unipersonal, cuando se consulta su equipo, entonces existe una fila `barber` y el sistema no usa una forma de datos ni una rama de código especial para “barbero único”. |
| `CA-021-02` | Dada una barbería con un barbero, cuando se registran tres más, entonces la lista devuelve cuatro recursos distintos pertenecientes a la misma barbería. |
| `CA-021-03` | Dado un nombre vacío, compuesto solo por espacios o mayor de 120 caracteres, cuando se intenta crear o editar un barbero, entonces se responde `422` con error de campo y no se persiste el cambio. |
| `CA-021-04` | Dado un barbero existente de la barbería activa, cuando se modifica su nombre con un valor válido, entonces la lista muestra el nuevo nombre sin duplicar el recurso. |
| `CA-021-05` | Dado el identificador de un barbero de otra barbería, cuando se intenta consultarlo o editarlo, entonces se responde `404`; cuando se lista el equipo, nunca aparece en el resultado. |
| `CA-021-06` | La base de datos rechaza insertar o relacionar un `barber` con un `barbershop_id` distinto del contexto de la transacción, incluso mediante acceso directo con el rol de aplicación. |
| `CA-021-07` | El API no expone operaciones de borrado, desactivación, credenciales, servicios ni horarios para esta historia; el contrato y el cliente generado contienen únicamente lectura, alta y edición del nombre. |
| `CA-021-08` | La pantalla presenta estados de carga, vacío, error recuperable y éxito; añadir o editar es operable con teclado, mantiene objetivos táctiles de 44 px y funciona sin pérdida a 320, 360, 768 y 1280 px. |

**Pruebas obligatorias**

- Unitarias del dominio y del caso de uso para nombre válido/inválido y actualización sin duplicación.
- HTTP y contrato para lista, alta, edición, `422`, no autenticado y `404` de otro tenant.
- Integración con PostgreSQL real y dos barberías, incluidas RLS, escritura cruzada y la representación idéntica de una y cuatro personas.
- Pruebas de componente para lista, vacío, alta, edición y error recuperable.
- E2E privado del recorrido añadir → listar → renombrar, más evidencia responsive y accesible.

**Terminado cuando** una barbería con una o cuatro personas usa exactamente el mismo modelo, API y pantalla, sin acceso cruzado y sin adelantar servicios, horarios, cuentas o agenda.

---

## 5. Historias pendientes de redacción

| Bloque | Rango reservado | Se redacta cuando |
| --- | --- | --- |
| B1 | `HU-022` – | B0 cumple su criterio de salida; `HU-020` y `HU-021` ya están preparadas como propuesta |
| B2 | `HU-040` – | B1 cumple su criterio de salida |
| B3 | `HU-060` – | B2 cumple su criterio de salida |
| B4 | `HU-090` – | B3 cumple su criterio de salida |
| B5 | `HU-130` – | B4 cumple su criterio de salida |
| B6 | `HU-150` – | B5 cumple su criterio de salida |

Redactar por anticipado las historias de un bloque lejano produce texto que hay que reescribir: lo que se aprende construyendo el bloque anterior cambia la historia siguiente.

---

## 6. Dudas que bloquean historias de B0

| Duda | Historia bloqueada | Qué falta decidir |
| --- | --- | --- |
| `DP-SEG-04` | `HU-005`, `HU-006` | Mecanismo concreto de la sesión larga y su duración exacta |
| `DP-SEG-05` | `HU-008`, `HU-011` | Canal y proveedor del código de recuperación |
| `DP-SEG-06` | `HU-007` | Duración de la ventana del límite por IP y del escalamiento |

Ninguna de las tres se resuelve escribiendo código. Se registran en [dudas-pendientes.md](../00-control/dudas-pendientes.md) y esperan un `DEC-*`, conforme a `AGENTS.md`.
