---
titulo: "Historias de usuario y criterios de aceptación"
version: "1.16"
estado: "Propuesta"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-25"
documentos_relacionados:
  - "../01-producto/alcance-mvp.md"
  - "../01-producto/reglas-negocio.md"
  - "estados-citas.md"
  - "../10-backlog/plan-bloques.md"
  - "../10-backlog/prompts-implementacion.md"
  - "../00-control/matriz-trazabilidad.md"
  - "../00-control/dudas-pendientes.md"
  - "../00-control/contradicciones.md"
---

# Historias de usuario y criterios de aceptación

> **Estado del contenido: Propuesta.** Estas historias **derivan** de funciones P0 y reglas confirmadas; no crean alcance por sí solas y requieren aprobación antes de implementarse. B0 y `HU-020`–`HU-024` ya están integradas en `main`. Las dudas históricas de B0 quedaron resueltas por `DEC-050`–`DEC-066`; `DP-SER-01`–`DP-SER-03` del lote de B1 quedaron resueltas por `DEC-067`–`DEC-069`; `CT-008` del lote de B2 quedó resuelta por `DEC-070`. `HU-040` tiene issue real [#90](https://github.com/bcaceres19/barberia/issues/90) y su prompt está `ready`; `HU-041`–`HU-042` siguen `draft` a la espera de que `HU-040` se integre.

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
| B1 · Identidad de la barbería y catálogo | `HU-020` – `HU-024` | Integradas en `main` ([PR #79](https://github.com/bcaceres19/barberia/pull/79), [PR #83](https://github.com/bcaceres19/barberia/pull/83), [PR #84](https://github.com/bcaceres19/barberia/pull/84)) |
| B2 · Horario laboral y bloqueos | `HU-040` – `HU-042` | `CT-008` resuelta (`DEC-070`); `HU-040` en ejecución (issue real [#90](https://github.com/bcaceres19/barberia/issues/90)); `HU-041`–`HU-042` siguen `draft` |
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
| Decisiones | `DEC-026`, `DEC-050`, `DEC-055`, `DEC-057`, `DEC-058` |
| Actor | Barbero |
| Depende de | `HU-002`, `HU-003` |
| Bloquea | `HU-006`, `HU-007`, `HU-010`, `HU-012` |
| Riesgo | Es la puerta de todo el área privada: un defecto aquí compromete todos los datos del sistema. |

**Historia**

> Como barbero, quiero entrar a mi agenda con mi correo y mi contraseña, para empezar a trabajar sin recordar códigos ni instalar nada.

> **Bloqueo resuelto:** `DEC-026` confirmaba "correo, contraseña y sesión larga" sin fijar el mecanismo de sesión ni su duración exacta. `DP-SEG-04` quedó resuelta el 2026-08-11 como `DEC-050`: cookie `HttpOnly`+`Secure`+`SameSite` con token opaco revocable en `staff_session`, 30 días con renovación por uso.

> **Bloqueo resuelto:** `CT-003` señalaba que declarar el login bajo `/api/v1/private` lo volvía circular, porque ese prefijo exige sesión previa. Quedó resuelta el 2026-08-13 como `DEC-055`: el login se mueve a `/api/v1/public/auth/login`, la única operación pública del módulo `auth`.

> **Bloqueo resuelto:** al implementar esta historia (issue `#44`) aparecieron dos vacíos nuevos. `DP-SEG-07` (atributos exactos de la cookie) quedó resuelta como `DEC-057`: `barberia_session`, `Path=/api/v1`, `SameSite=Lax`, sin `Domain`, 30 días. `DP-SEG-08` (`CA-005-05` exige un endpoint privado real que todavía no existe) quedó resuelta como `DEC-058`: el criterio se divide entre esta historia (aislamiento a nivel PostgreSQL/RLS) y `HU-006` (verificación end-to-end, nuevo `CA-006-07`).

**Alcance incluido**

- Migración con la tabla de credenciales y la de sesiones, ambas con `barbershop_id` y RLS.
- Almacenamiento de contraseña con función de derivación resistente a fuerza bruta y parámetros documentados; nunca cifrado reversible ni hash simple.
- Operación de inicio de sesión en `/api/v1/public/auth/login` declarada primero en OpenAPI (`DEC-055`).
- Respuesta uniforme ante credenciales inválidas y ante correo inexistente: **el mismo mensaje y el mismo tiempo de respuesta**, para no revelar qué correos existen.
- Registro del intento sin exponer el correo en claro en los logs.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-005-01` | Dadas credenciales válidas, cuando el barbero inicia sesión, entonces obtiene una sesión asociada a su barbería y a su usuario; el contexto de esa barbería queda probado a nivel de PostgreSQL/RLS en esta historia, y la verificación end-to-end contra solicitudes privadas reales se confirma en `CA-006-07` de `HU-006` (`DEC-058`). |
| `CA-005-02` | Dado un correo inexistente y dado un correo existente con contraseña incorrecta, entonces ambas respuestas son indistinguibles en cuerpo, código y tiempo perceptible. |
| `CA-005-03` | La contraseña se almacena mediante derivación de clave con sal única por usuario; la base de datos no contiene ninguna contraseña legible ni reversible. |
| `CA-005-04` | Ni la contraseña, ni el correo, ni el material de sesión aparecen en registros, mensajes de error o respuestas. |
| `CA-005-05` | Una sesión emitida para la barbería A no habilita ninguna operación sobre datos de B; en esta historia se prueba con aislamiento real de `staff_session`/`staff_credential` a nivel PostgreSQL/RLS con dos tenants, ya que `HU-005` no depende de ningún endpoint privado real. La verificación end-to-end contra una operación privada real (logout) se confirma en `CA-006-07` de `HU-006` (`DEC-058`). |
| `CA-005-06` | La operación está declarada en OpenAPI con su esquema, sus errores y sus ejemplos antes de existir el handler. |
| `CA-005-07` | Un usuario desactivado o eliminado no puede iniciar sesión ni conservar sesiones vigentes. |

**Pruebas obligatorias**

- Unitarias del servicio de autenticación (verificación, normalización del correo, usuario inactivo).
- Integración HTTP: éxito, credenciales inválidas, correo inexistente, aislamiento entre barberías.
- Prueba de que el hash almacenado cambia con la misma contraseña en dos usuarios distintos.

**Terminado cuando** el contrato está publicado y todas las pruebas anteriores pasan.

---

### HU-006 · Sesión persistente, cierre de sesión y protección del área privada

| Campo | Valor |
| --- | --- |
| Función | `F-AUTH-01` |
| Reglas | `RN-TEN-01` |
| Decisiones | `DEC-026`, `DEC-050`, `DEC-055`, `DEC-058` |
| Actor | Barbero |
| Depende de | `HU-005` |
| Bloquea | `HU-012` |
| Riesgo | Una sesión que no se puede revocar convierte cualquier dispositivo perdido en un acceso permanente a los datos de los clientes. |

**Historia**

> Como barbero, quiero seguir dentro de la aplicación al día siguiente sin volver a escribir mi contraseña, y quiero poder cerrar sesión y que ese cierre sea inmediato y real.

> **Bloqueo resuelto:** `CT-003` dejaba en duda si `CA-006-04` exigía una excepción para el login. Quedó resuelta el 2026-08-13 como `DEC-055`: el login vive en `/api/v1/public/auth/login`, fuera de `/api/v1/private`, así que `CA-006-04` no necesita ninguna excepción.

> **Criterio nuevo:** `DP-SEG-08`, detectada al implementar `HU-005` (issue `#44`), quedó resuelta como `DEC-058`: `CA-005-05` se prueba en `HU-005` solo a nivel PostgreSQL/RLS porque todavía no existe ningún endpoint privado real; esta historia añade `CA-006-07` para completar esa verificación end-to-end contra el logout, el primer endpoint privado real.

**Alcance incluido**

- Vigencia de 30 días con renovación por uso (`DEC-050`), token opaco revocable almacenado en `staff_session`.
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
| `CA-006-07` | Una sesión emitida para la barbería A nunca ejecuta el logout (ni ninguna otra operación privada real) sobre datos de la barbería B, verificado end-to-end contra el endpoint real (`DEC-058`, completa `CA-005-05` de `HU-005`). |

**Pruebas obligatorias**

- Integración HTTP: persistencia, revocación, vencimiento, ruta privada sin sesión.
- Prueba estructural que enumera las rutas privadas registradas y verifica que todas pasan por el middleware de sesión.
- Aislamiento end-to-end (`CA-006-07`): sesión de A contra el logout de B, con dos tenants reales.

**Terminado cuando** el cierre de sesión es verificable desde el servidor y la prueba estructural protege contra rutas privadas olvidadas.

---

### HU-007 · Defensa escalonada contra abuso en el acceso

| Campo | Valor |
| --- | --- |
| Función | `F-SEG-03` |
| Reglas | `RN-DAT-02` |
| Decisiones | `DEC-026`, `DEC-052`, `DEC-061`, `DEC-062` |
| Actor | Propietario (protege), barbero (afectado si se excede) |
| Depende de | `HU-003`, `HU-005` |
| Bloquea | — |
| Riesgo | Sin límite, un ataque automatizado prueba miles de contraseñas; con un límite mal calibrado, el barbero legítimo queda fuera en plena jornada. |

**Historia**

> Como propietario del sistema, necesito que el formulario de acceso limite los intentos por IP y exija una prueba adicional cuando se supera el umbral, para frenar el abuso sin castigar al barbero que se equivocó dos veces.

> **Bloqueo resuelto:** `DEC-026` fijaba el umbral inicial de **5 solicitudes por IP** sin la duración de la ventana ni la del escalamiento. `DP-SEG-06` quedó resuelta el 2026-08-11 como `DEC-052`: ventana de 15 minutos, escalamiento a verificación telefónica de 24 horas. `CT-005` (¿la quinta o la sexta solicitud exige el reto?) y `DP-SEG-10` (reto telefónico completo) quedaron resueltas el 2026-08-17 como `DEC-061` y `DEC-062`: las cinco primeras solicitudes se evalúan con normalidad, la sexta exige completar el reto en `POST /api/v1/public/auth/challenge` y `.../challenge/verify` (código de 6 dígitos por WhatsApp oficial, 5 min, 5 intentos) antes de evaluar la contraseña.

**Alcance incluido**

- Conteo por IP con ventana de 15 minutos y umbral de 5 solicitudes (`DEC-052`), ambos configurables.
- Escalamiento: al superar el umbral (la sexta solicitud, `DEC-061`), esa solicitud y las siguientes exigen completar el reto telefónico de `DEC-062` antes de evaluar la contraseña, durante 24 horas o hasta completarlo con éxito.
- Respuesta `429` con formato uniforme y con indicación de cuándo reintentar.
- Configuración expuesta como parámetros, no como números incrustados en el código.
- Los contadores no almacenan datos personales; la IP se guarda de forma acotada y con vencimiento.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-007-01` | Dadas hasta cinco solicitudes desde la misma IP dentro de la ventana, entonces todas se evalúan normalmente, incluida la contraseña (`DEC-061`). |
| `CA-007-02` | La sexta solicitud desde esa IP dentro de la ventana exige completar el reto telefónico de `DEC-062` y no evalúa la contraseña hasta lograrlo. |
| `CA-007-03` | La respuesta al superar el umbral usa el formato uniforme, indica cuándo reintentar y no revela si el correo existe. |
| `CA-007-04` | Transcurrida la ventana sin nuevos intentos, el conteo se reinicia; el reto telefónico sigue vigente si `escalated_until` no venció o no se completó con éxito (`DEC-062`). |
| `CA-007-05` | El umbral y la ventana se cambian por configuración, sin recompilar ni editar código. |
| `CA-007-06` | Los datos de conteo vencen solos y no contienen correo, nombre ni teléfono. |
| `CA-007-07` | El límite no depende de una cabecera que el cliente pueda falsificar libremente; la obtención de la IP está documentada y probada según el despliegue previsto. |

**Pruebas obligatorias**

- Integración: recorrido del umbral, escalamiento, expiración de la ventana, dos IP distintas independientes.
- Prueba de que una cabecera manipulada no evade el límite en la configuración de despliegue documentada.

**Terminado cuando** el escalamiento queda demostrado de extremo a extremo.

---

### HU-008 · Recuperación de acceso con código al teléfono verificado

| Campo | Valor |
| --- | --- |
| Función | `F-AUTH-02` |
| Reglas | `RN-DAT-01`, `RN-DAT-02` |
| Decisiones | `DEC-026`, `DEC-051`, `DEC-063`, `DEC-064`, `DEC-065` |
| Actor | Barbero |
| Depende de | `HU-005`, `HU-007` |
| Bloquea | `HU-011` |
| Riesgo | Un barbero sin acceso en plena jornada pierde su agenda; un mecanismo de recuperación débil es la vía más común de secuestro de cuentas. |

**Historia**

> Como barbero que olvidó su contraseña, quiero recuperar el acceso con un código enviado por WhatsApp y correo, para volver a mi agenda el mismo día sin depender de que alguien me responda.

> **Bloqueo resuelto:** `DEC-026` definía el mecanismo (código al teléfono verificado) sin fijar canal ni proveedor. `DP-SEG-05` quedó resuelta el 2026-08-11 como `DEC-051`: WhatsApp oficial y correo, reutilizando el proveedor ya habilitado por `DEC-027`. El 2026-08-17 se resolvieron las últimas cuatro dudas: `DP-SEG-11` (política de contraseña) como `DEC-063`, `DP-SEG-12` (formato/vigencia/intentos/token de reinicio del código) como `DEC-064`, `CT-006` (respuesta idéntica frente a destino enmascarado) como `DEC-065`, y `DP-NOT-05` (proveedor/adaptador real de WhatsApp y correo: Meta Cloud API + Resend) como `DEC-066`. `HU-008` ya no depende de ninguna decisión pendiente; solo espera que `HU-007` se integre en `main`.

**Alcance incluido**

- Migración con la tabla de códigos de recuperación: hash del código, vencimiento corto, intentos, marca de uso, `barbershop_id` y RLS.
- Solicitud de recuperación, verificación del código y establecimiento de contraseña nueva.
- Envío por WhatsApp oficial y por correo (`DEC-051`), mismo proveedor de `DEC-027`.
- Código de un solo uso, con vencimiento, con límite de intentos y con reenvío controlado.
- Invalidación de las sesiones activas al cambiar la contraseña.
- Destino mostrado enmascarado en la interfaz y en las respuestas.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-008-01` | Dado un correo registrado, cuando se solicita recuperación, entonces se envía un código al teléfono/correo verificados y la respuesta de `POST /recovery/request` es idéntica a la de un correo no registrado, sin destino en el cuerpo (`DEC-065`). |
| `CA-008-02` | El código vence en 15 minutos (`DEC-064`), se acepta una sola vez y queda inválido tras usarse. |
| `CA-008-03` | Superados 5 intentos fallidos (`DEC-064`), el código se invalida por completo y debe solicitarse uno nuevo. |
| `CA-008-04` | El código se almacena como `HMAC-SHA256` con secreto de despliegue (`DEC-064`); la base de datos no contiene el valor enviado ni una representación recuperable sin ese secreto. |
| `CA-008-05` | Al establecer la contraseña nueva usando el token de reinicio emitido en la verificación (`DEC-064`), todas las sesiones activas del usuario quedan invalidadas en la misma transacción. |
| `CA-008-06` | El destino (teléfono/correo) se muestra enmascarado únicamente en la respuesta exitosa de `POST /recovery/verify`, nunca en la solicitud (`DEC-065`), y nunca aparece completo en respuestas ni registros. |
| `CA-008-07` | El reenvío tiene cooldown de 60 s y máximo 3 por hora (`DEC-064`); cada reenvío invalida atómicamente el código anterior, nunca deja dos vigentes. |
| `CA-008-08` | La contraseña nueva se rechaza si no cumple la política de `DEC-063` (10-128 caracteres, no igual al correo ni a la contraseña actual), con un mensaje que explica qué falta. |

**Pruebas obligatorias**

- Integración: recorrido completo, código vencido, código ya usado, exceso de intentos, correo inexistente, invalidación de sesiones.
- Prueba del adaptador de envío con doble de prueba; el proveedor real no se invoca en pruebas.

**Terminado cuando** el recorrido completo funciona con el adaptador seleccionado.

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
| Decisiones | `DEC-039`, `DEC-033`, `DEC-056` |
| Actor | Barbero |
| Depende de | `HU-005`, `HU-009` |
| Bloquea | `HU-012` |
| Riesgo | Es la primera pantalla que el barbero ve del producto. Un formulario confuso o lento en un celular de gama media es una mala primera impresión difícil de revertir. |

**Historia**

> Como barbero, quiero una pantalla de acceso simple, con mi correo y mi contraseña visibles en un solo lugar y con la recuperación a la vista, para entrar rápido incluso con una conexión mala.

> **Bloqueo resuelto:** `CT-004` señalaba que `HU-010` no podía demostrar su destino de éxito contra el panel privado sin implementar parte de `HU-012`, que a su vez depende de `HU-010`. Quedó resuelta el 2026-08-13 como `DEC-056`: `CA-010-01` se divide — `HU-010` navega a `/panel`, ruta privada real y protegida con un guard mínimo propio, sin cabecera ni navegación; `HU-012` reutiliza ese guard y construye ahí el cascarón completo.

**Alcance incluido**

- Composición según `estandar-diseno-visual.md` sección 10: marca discreta, título, formulario estrecho y recuperación visible.
- Cliente tipado generado desde el bundle OpenAPI; sin DTO manuales paralelos.
- Estados: inicial, enviando, error de credenciales, error de red con datos conservados, y bloqueo por umbral de `HU-007` con explicación.
- Etiquetas asociadas, mensajes de error junto al campo y resumen accesible cuando haya varios.
- Ruta cargada de forma diferida.
- Ruta privada `/panel` y su guard mínimo (`DEC-056`): sin sesión válida redirige al acceso; con sesión válida muestra un marcador de posición autenticado, sin cabecera ni navegación general — esa estructura la construye `HU-012` reutilizando este guard, no uno paralelo.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-010-01` | Con credenciales válidas, el barbero entra y la aplicación navega a `/panel`, ruta privada real y protegida por el guard de esta historia; la verificación de la llegada al panel completo (cabecera, navegación) se confirma en `CA-012-01`/`CA-012-02` de `HU-012` (`DEC-056`). |
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
| Decisiones | `DEC-033`, `DEC-039`, `DEC-056`, `DEC-060` |
| Actor | Barbero |
| Depende de | `HU-006`, `HU-009`, `HU-010` |
| Bloquea | Todas las pantallas privadas de B1 en adelante |
| Riesgo | Si cada pantalla resuelve por su cuenta la sesión, la navegación y los errores, el mismo problema se implementa siete veces de siete maneras. |

**Historia**

> Como barbero, quiero que la aplicación recuerde mi sesión, me lleve a mi agenda al entrar, me muestre siempre en qué barbería estoy y me avise con claridad cuando se pierda la conexión, para no quedarme mirando una pantalla en blanco.

**Alcance incluido**

- Estructura de aplicación autenticada: cabecera con el nombre de la barbería, navegación principal y área de contenido, según las estructuras base del estándar visual, construida sobre la ruta `/panel` y el guard mínimo que ya entrega `HU-010` (`DEC-056`) — sin crear un guard paralelo ni cambiar la ruta.
- Guarda de ruta: sin sesión válida se redirige al acceso, conservando el destino pretendido; generaliza el guard de `/panel` de `HU-010` a todas las rutas privadas.
- Manejo central de respuestas no autorizadas: la sesión se limpia y se informa el motivo, sin bucles de redirección.
- Patrones globales de estado: carga con esqueleto, error recuperable con "Reintentar" y aviso de conexión perdida.
- Una pantalla de inicio provisional que muestra el estado de la sesión; **no** es la agenda, que llega en B3.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-012-01` | Con sesión válida, al abrir la aplicación se entra directamente al panel sin pedir credenciales; la restauración se verifica contra `GET /api/v1/private/auth/session` (`DEC-060`), nunca contra un marcador local. |
| `CA-012-02` | Sin sesión, cualquier ruta privada redirige al acceso y, tras entrar, lleva al destino que se pretendía abrir. |
| `CA-012-03` | Ante una respuesta de no autorizado, la sesión local se limpia una sola vez y no se produce un bucle de redirección. |
| `CA-012-04` | La cabecera muestra siempre la barbería activa, con el nombre devuelto por `GET /api/v1/private/auth/session` (`DEC-060`); el barbero nunca puede dudar de en qué contexto está operando. |
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

> **B1 integrado.** `HU-020`–`HU-024` están integradas en `main`; `DP-SER-01`–`DP-SER-03` quedaron resueltas por `DEC-067`–`DEC-069` y el cierre de B1 está respaldado por [PR #84](https://github.com/bcaceres19/barberia/pull/84).

Orden de construcción recomendado para esta parte del bloque: `HU-020` → `HU-021` → `HU-022` → `HU-023` → `HU-024`.

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

### HU-022 · Catálogo básico de servicios

| Campo | Valor |
| --- | --- |
| Función | `F-SERV-01`, `F-SERV-02` (alta, consulta y edición del catálogo; ciclo de activación en `HU-024`) |
| Reglas | `RN-SER-01`, `RN-SER-02`, `RN-SER-04`, `RN-TEN-01`, `RN-DAT-02` |
| Decisiones | `DEC-002`, `DEC-004`, `DEC-024`, `DEC-033`–`DEC-040`, `DEC-043`, `DEC-067` |
| Actor | Barbero autenticado |
| Depende de | `HU-020`, `HU-021` |
| Bloquea | `HU-023`, `HU-024`, disponibilidad y reserva pública |
| Estado | Integrada en `main` ([PR #79](https://github.com/bcaceres19/barberia/pull/79), squash-merge, CI 4/4 verde) |
| Riesgo | Tratar precio como punto flotante, aceptar precio 0 o nombres duplicados entre servicios activos, o propagar cambios a citas, rompería `DEC-067`, integridad e historia. |

**Historia**

> Como barbero, quiero crear, consultar y editar los servicios de mi barbería con su duración planificada y precio informativo, para mantener un catálogo reutilizable por todos los barberos y por la reserva futura.

**Alcance incluido**

- Recurso `service` perteneciente a la barbería, separado de `barber` y de cualquier cita, con nombre, descripción opcional, duración en minutos enteros, importe en COP, estado activo y marcas de tiempo.
- Lista paginada, lectura individual, alta idempotente y edición de los campos de catálogo autorizados.
- Duraciones positivas configurables, incluidos valores como 25 o 45 minutos; la interfaz no impone una lista cerrada de presets.
- Pantalla privada “Servicios” en el módulo `catalog`, compuesta dentro del cascarón existente sin duplicar guard, navegación ni cliente HTTP.
- Contrato OpenAPI privado, módulo Go `catalog`, migración Atlas mínima, RLS, cliente TypeScript generado y pruebas de dos tenants.
- Moneda COP fija (sin campo editable), precio estrictamente mayor que cero y nombre único entre servicios activos de la misma barbería, conforme a `DEC-067`.

**Alcance excluido**

- Asignar servicios a barberos (`HU-023`).
- Desactivar, reactivar o eliminar servicios (`HU-024`); no se expone `DELETE` físico.
- Horarios, disponibilidad, enlace público, citas, snapshots, cambios masivos, historial y notificaciones.
- Precios por barbero, descuentos, impuestos, promociones, paquetes, inventario, comisiones o cobro en línea.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-022-01` | Dada una sesión válida, cuando se abre “Servicios”, entonces se listan únicamente los servicios de la barbería activa con paginación estable y estados de carga, vacío y error recuperable. |
| `CA-022-02` | Dados nombre y duración válidos y un precio en COP estrictamente mayor que cero, cuando se crea un servicio con una clave idempotente, entonces se persiste una sola fila activa y un reintento exacto reproduce el mismo recurso. |
| `CA-022-03` | Dadas duraciones enteras positivas como 25, 30, 45 y 90 minutos, cuando se crean servicios, entonces todas son aceptadas sin redondeo ni catálogo cerrado; cero, negativas, fraccionarias o fuera del límite técnico documentado se rechazan. |
| `CA-022-04` | Dados campos vacíos, excesivos, desconocidos, un precio menor o igual a cero, o un nombre igual al de otro servicio activo de la misma barbería (`DEC-067`), cuando se crea o edita, entonces se responde `400`, `409` o `422` según el contrato y no se persiste un estado parcial. |
| `CA-022-05` | Cuando se cambia nombre, descripción, duración o precio del catálogo, entonces solo cambia `service`; ninguna operación de esta HU acepta alcance sobre citas ni simula propagación a recursos futuros (`RN-SER-04`). |
| `CA-022-06` | Dado un identificador de servicio de otra barbería, cuando se consulta o edita, entonces se responde el mismo `404` que para uno inexistente y RLS impide toda lectura o escritura cruzada. |
| `CA-022-07` | OpenAPI, handlers, cliente generado, dominio, migración y formulario coinciden en campos, límites, precisión monetaria, errores y trazabilidad `RN-*`/`DEC-*`. |
| `CA-022-08` | La pantalla permite crear y editar con teclado, evita doble envío, conserva datos ante error recuperable y funciona sin pérdida a 320, 360, 768 y 1280 px y zoom 200 %. |

**Pruebas obligatorias**

- Dominio y servicio para duración, nombre, descripción y valor monetario; casos definidos por `DEC-067` (precio > 0, nombre único entre activos).
- HTTP/contrato para lista, lectura, alta idempotente, edición, validación, conflicto, `401`, `404` tenant-aware y `500`.
- PostgreSQL 14 real con dos tenants, precisión `numeric`, restricciones, RLS forzada, grants mínimos y ausencia de `DELETE`.
- Componentes y E2E privado crear → listar → editar → recargar, con evidencia responsive y accesible.

**Terminado cuando** el catálogo básico funciona de extremo a extremo sin asignaciones, ciclo de vida ni citas, y `DEC-067` está incorporada en todas las capas.

---

### HU-023 · Asignación de servicios a barberos

| Campo | Valor |
| --- | --- |
| Función | `F-CONF-02` (servicios por barbero) |
| Reglas | `RN-SER-03`, `RN-SER-04`, `RN-TEN-01`, `RN-DAT-02` |
| Decisiones | `DEC-004`, `DEC-019`, `DEC-024`, `DEC-033`–`DEC-040`, `DEC-068` |
| Actor | Barbero autenticado |
| Depende de | `HU-021`, `HU-022` integradas en `main` (cumplido) |
| Bloquea | Horarios por barbero, disponibilidad y selección pública de barbero |
| Estado | Implementada; issue real [#76](https://github.com/bcaceres19/barberia/issues/76); [PR #83](https://github.com/bcaceres19/barberia/pull/83) abierto contra `main`, pendiente de revisión/CI/merge |
| Riesgo | Una relación sin tenant compuesto puede asociar recursos de barberías distintas; permitir retirar la última asignación de un servicio activo contradiría `DEC-068` y dejaría un servicio activo imposible de reservar. |

**Historia**

> Como barbero, quiero indicar qué servicios presta cada persona de mi equipo, para que la misma configuración represente correctamente una barbería independiente y una con varios especialistas.

**Alcance incluido**

- Asociación tenant-aware `barber_service` entre filas reales de `barber` y `service`, sin copiar nombre, duración ni precio.
- Consulta de asignaciones por barbero y operaciones explícitas para asignar y desasignar; desasignar la última fila activa de un servicio activo se rechaza (`DEC-068`).
- La misma ruta y pantalla funcionan para un barbero con todos los servicios, varios con servicios compartidos y un equipo con especialidades distintas.
- Interfaz dentro de la experiencia privada existente, usando APIs públicas mínimas de `staff` y `catalog` o composición en `app`; ningún módulo importa archivos internos de otro.
- Migración Atlas mínima, RLS, FK compuestas, contrato privado, módulo Go y pruebas con dos tenants.

**Alcance excluido**

- Crear, editar, desactivar o reactivar servicios; renombrar barberos.
- Horarios, disponibilidad, citas, selección pública, precios por barbero o duración distinta por profesional.
- Credenciales, roles, comisiones, orden manual, borrado de barberos o servicios.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-023-01` | Dado un barbero y servicios de su barbería, cuando se consultan sus asignaciones, entonces se devuelve el conjunto real sin duplicados y sin datos de otro tenant. |
| `CA-023-02` | Cuando se asigna un servicio permitido a un barbero, entonces la relación aparece al recargar y repetir exactamente la operación no crea una segunda fila. |
| `CA-023-03` | Un mismo servicio puede asignarse a varios barberos de la misma barbería y cada asociación sigue siendo un recurso independiente. |
| `CA-023-04` | Intentar asociar un barbero o servicio de otra barbería responde `404` o conflicto uniforme y la FK tenant-aware/RLS rechaza la relación incluso con el rol real de aplicación. |
| `CA-023-05` | Intentar retirar la última asignación activa de un servicio activo se rechaza con `409`/`422` (`DEC-068`); la operación es segura ante repetición y nunca borra el barbero, el servicio ni una cita. |
| `CA-023-06` | Un servicio activo nunca queda con cero asignaciones como resultado de esta historia; base de datos, API e interfaz coinciden en rechazar ese estado (`DEC-068`). |
| `CA-023-07` | La asociación no contiene nombre, duración, precio, estado ni columnas de otro dueño; el contrato no expone precios o duraciones por barbero. |
| `CA-023-08` | La interfaz anuncia guardado/error, funciona con teclado, evita doble envío y conserva una composición usable en 320, 360, 768 y 1280 px. |

**Pruebas obligatorias**

- Dominio/servicio para asignación repetida, desasignación y rechazo de la última asignación activa (`DEC-068`).
- PostgreSQL real con dos tenants, FK compuestas, RLS, permisos exactos y carreras de asignación repetida.
- HTTP/contrato para consulta, alta/baja de asociación, `401`, `404`, conflicto y campos desconocidos.
- Componente y E2E para un barbero, cuatro barberos, servicio compartido y aislamiento cruzado.

**Terminado cuando** cada barbero conserva un conjunto tenant-aware de servicios sin duplicar atributos ni adelantar disponibilidad, horarios o citas.

---

### HU-024 · Desactivación y reactivación de servicios

| Campo | Valor |
| --- | --- |
| Función | `F-SERV-01` (ciclo de vida sin borrado físico) |
| Reglas | `RN-SER-03`, `RN-SER-04`, `RN-TEN-01`, `RN-IDE-01` |
| Decisiones | `DEC-003`, `DEC-004`, `DEC-024`, `DEC-033`–`DEC-040`, `DEC-043`, `DEC-069` |
| Actor | Barbero autenticado |
| Depende de | `HU-022`, `HU-023` integradas en `main` |
| Bloquea | Cierre de B1, consulta pública de servicios y flujos futuros con citas |
| Estado | Integrada en `main` contra el issue real [#77](https://github.com/bcaceres19/barberia/issues/77); [PR #84](https://github.com/bcaceres19/barberia/pull/84) |
| Riesgo | Desactivar sin una advertencia real omite compromisos futuros; simular citas antes de B3 crea contrato ficticio; construir protección de concurrencia sobre un conteo que en B1 nunca cambia (`DEC-069`) sería alcance inventado. |

**Historia**

> Como barbero, quiero desactivar o reactivar un servicio sin borrarlo y con una advertencia honesta sobre su impacto, para dejar de ofrecerlo sin perder historial ni modificar citas automáticamente.

**Alcance incluido**

- Estado activo/inactivo coherente y marca temporal de desactivación, sin borrado físico.
- Previsualización real del impacto (recuento de citas futuras afectadas, siempre 0 en B1 porque `appointment` no existe todavía) y confirmación, sin protección de concurrencia sobre el conteo (`DEC-069`); la interfaz nunca inventa un conteo distinto del real ni ofrece cancelar citas inexistentes.
- Desactivación idempotente que conserva por defecto toda cita existente y nunca propaga cambios de catálogo en silencio.
- Reactivación explícita y auditable a nivel técnico, sin recrear el servicio ni sus asignaciones.
- Estados visuales de advertencia, confirmación, conflicto, éxito y error recuperable en la pantalla de servicios.

**Alcance excluido**

- `DELETE` físico, purga o reutilización del identificador.
- Cancelar, reprogramar, notificar o modificar citas; esas operaciones pertenecen a B3/B5 y solo se enlazan cuando existan.
- Reserva pública y disponibilidad; las historias futuras deben respetar `is_active` sin ampliar esta entrega.
- Cambiar nombre, duración o precio durante el mismo comando de ciclo de vida.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-024-01` | Antes de desactivar, el sistema presenta el impacto real disponible mediante una consulta real (siempre 0 en B1, conforme a `DEC-069`); nunca lo muestra como un valor por defecto sin consultar ni simula datos. |
| `CA-024-02` | Cuando se confirma la desactivación conforme al contrato vigente, entonces el servicio queda inactivo con su identificador e historial conservados y no se borra ninguna asociación o cita por cascada. |
| `CA-024-03` | Las citas existentes se mantienen por defecto y ninguna se cancela, reprograma o modifica sin una operación futura explícita de la capacidad dueña (`RN-SER-03`, `RN-SER-04`). |
| `CA-024-04` | La confirmación de desactivación vuelve a consultar el impacto real en la misma operación (sin depender de un valor cacheado de la previsualización); no requiere bloqueo optimista porque en B1 el conteo no puede cambiar entre ambas llamadas (`DEC-069`). |
| `CA-024-05` | Cuando se reactiva un servicio, vuelve a estado activo sin crear otra fila ni alterar duración, precio o asignaciones. |
| `CA-024-06` | Repetir desactivación o reactivación con la misma intención produce un solo efecto; una clave reutilizada con otra intención se rechaza conforme a `RN-IDE-01`. |
| `CA-024-07` | Un servicio ajeno o inexistente produce el mismo `404`; el rol de aplicación no dispone de `DELETE` físico y RLS impide cambios cruzados. |
| `CA-024-08` | La advertencia y confirmación son operables por teclado, gestionan foco, no dependen solo del color y funcionan en 320, 360, 768 y 1280 px y zoom 200 %. |

**Pruebas obligatorias**

- Dominio/servicio para transiciones, repetición, conflicto y decisión de impacto.
- PostgreSQL real con dos tenants: invariantes `is_active`/marca temporal, ausencia de `DELETE` y RLS.
- HTTP/contrato para impacto, desactivar, reactivar, `401`, `404`, `409`, `422` y error interno.
- Componente/E2E para advertencia, confirmación, cancelación del diálogo, reintento y reactivación, sin simular citas fuera del contrato aprobado.

**Terminado cuando** el servicio cambia de estado sin borrado ni efectos automáticos y la división B1/B3 de `DEC-069` está probada, no asumida.

---

## 5. Bloque B2 · Horario laboral y bloqueos

> **B2 en ejecución.** Las tres historias separan las dos funciones P0 del bloque: HU-040 configura la jornada semanal; HU-041 resuelve fechas especiales y festivos; HU-042 administra bloqueos puntuales y recurrentes. CT-008 quedó resuelta como DEC-070 (las FK del modelo físico de B2 usan ON DELETE RESTRICT, no CASCADE). HU-040 tiene issue real #90 y su prompt está ready; HU-041/HU-042 permanecen draft hasta que HU-040 se integre en main.

Orden recomendado: HU-040 → HU-041 → HU-042. La disponibilidad, las citas afectadas y la agenda consumen estas capacidades en B3/B4; no se implementan aquí.

---

### HU-040 · Horario laboral recurrente

| Campo | Valor |
| --- | --- |
| Función | F-HOR-01 |
| Reglas | RN-TEN-01, RN-DIS-05, RN-DIS-07 |
| Decisiones | DEC-007, DEC-019, DEC-020, DEC-024, DEC-033–DEC-040 |
| Actor | Barbero autenticado |
| Depende de | Criterio de salida de B1; HU-020–HU-024 integradas en main; CT-008 resuelta (DEC-070) |
| Bloquea | HU-041, HU-042 y el cálculo posterior de disponibilidad |
| Estado | Implementación completa en rama feat/90-hu040-horario-laboral; prompt PROMPT-HU-040-v1 in_progress; issue real #90; pendiente de PR/CI |
| Riesgo | Un tramo ambiguo, solapado o interpretado en la zona del dispositivo desplaza toda la disponibilidad futura. |

**Historia**

> Como barbero, quiero configurar los tramos de trabajo de cada barbero por día de la semana, para que el sistema sepa cuándo ofrecer turnos y pueda representar jornadas partidas o nocturnas.

**Alcance incluido**

- Consulta, alta, edición y retiro de tramos recurrentes por barbero y barbería activa.
- Días ISO del 1 al 7, varios tramos por día y duración en minutos enteros.
- Horas civiles interpretadas con la zona IANA de la barbería; no se usa la zona del navegador.
- Intervalos semiabiertos [inicio, fin), con tramos contiguos permitidos y solapes del mismo barbero/día rechazados.
- Tramos que cruzan medianoche, según DEC-020, sin convertir la configuración en instantes dependientes del servidor.
- Contrato OpenAPI, módulo schedule, persistencia tenant-aware, pantalla privada y cliente tipado.

**Alcance excluido**

- Fechas especiales, festivos, horarios de excepción y calendario colombiano: HU-041.
- Descansos, almuerzos, vacaciones, emergencias, recurrencias y listas explícitas: HU-042.
- Citas, agenda, disponibilidad pública, reservas y notificaciones.
- Credenciales, roles, eliminación de barberos y personalización visual por barbería.
- Cualquier FK o cascada del modelo físico distinta de ON DELETE RESTRICT (DEC-070).

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| CA-040-01 | Dada una barbería con uno o varios barberos, cuando se consulta el horario, entonces aparecen los siete días y solo los tramos de la barbería activa, sin filas de otro tenant. |
| CA-040-02 | Cuando se crean dos tramos no solapados el mismo día, entonces ambos persisten; una jornada partida no requiere una rama especial para el segundo tramo. |
| CA-040-03 | Una duración positiva representa el fin del tramo de forma determinista; los intervalos contiguos no chocan y un tramo nocturno puede cruzar medianoche conforme a DEC-020. |
| CA-040-04 | Un día fuera de 1–7, duración cero/negativa/excesiva, hora inválida o solape del mismo barbero responde con error de validación y no deja escritura parcial. |
| CA-040-05 | Un barbero puede editar o retirar un tramo propio; intentar leerlo o modificarlo desde otra barbería responde 404 y la base de datos rechaza la operación cruzada con el rol de aplicación. |
| CA-040-06 | La misma configuración se muestra con la zona de la barbería aunque el dispositivo use otra zona horaria; los instantes de una capacidad futura no se reescriben. |
| CA-040-07 | La pantalla contempla carga, vacío, error recuperable y éxito; conserva los datos no sensibles ante error, evita doble envío y es operable por teclado. |
| CA-040-08 | El contrato, handler, caso de uso, repositorio, cliente generado y pantalla coinciden en campos, límites y errores, con verificación responsive a 320, 360, 768 y 1280 px. |

**Pruebas obligatorias**

- Unitarias de dominio y aplicación para días, duración, tramos partidos, contiguidad, solape y medianoche.
- HTTP y contrato para consulta, alta, edición, retiro, no autenticado, recurso ajeno, validación y error uniforme.
- PostgreSQL real con dos barberías, RLS, FK en ON DELETE RESTRICT (DEC-070) y grants, lectura/escritura cruzada y ausencia de escritura parcial.
- Componente y E2E del selector de barbero, alta, edición, conflicto, retiro, recarga y aislamiento.
- Evidencia responsive y accesible en 320, 360, 768 y 1280 px, teclado, foco, zoom 200 %, axe-core y contraste.

**Terminado cuando** un barbero puede mantener la semana laboral real de cualquier profesional de su barbería, incluidos tramos partidos y nocturnos, sin cruces de tenant y sin adelantar excepciones, bloqueos o citas.

---

### HU-041 · Excepciones de jornada y festivos

| Campo | Valor |
| --- | --- |
| Función | F-HOR-02, parte de fechas especiales y festivos |
| Reglas | RN-TEN-01, RN-BLQ-02, RN-DIS-05, RN-DIS-07 |
| Decisiones | DEC-007, DEC-019, DEC-020, DEC-024, DEC-033–DEC-040 |
| Actor | Barbero autenticado |
| Depende de | HU-040 integrada en main y B1 cerrado |
| Bloquea | HU-042 y la resolución posterior de jornada efectiva |
| Estado | Propuesta; prompt PROMPT-HU-041-v1 en draft; issue pending |
| Riesgo | Compartir el calendario entre barberos o dejar que un festivo automático venza una apertura manual produce disponibilidad falsa. |

**Historia**

> Como barbero, quiero activar el calendario colombiano de festivos y definir excepciones para fechas concretas, para cerrar un día o trabajar con un horario especial sin alterar a los demás barberos.

**Alcance incluido**

- Toggle holiday_calendar_enabled independiente por barbero.
- Resolución determinista del calendario colombiano para la fecha consultada y la zona de la barbería.
- Excepción por fecha para cerrar completamente el día o abrirlo con uno o varios tramos propios.
- Apertura manual de un festivo completo o con horario reducido; la excepción manual prevalece sobre el bloqueo automático.
- Consulta, alta, edición y retiro de excepciones, con una sola cabecera por barbero y fecha.
- Puerto interno de resolución de jornada efectiva para consumos futuros, sin crear citas ni disponibilidad pública.
- Modelo normalizado de cabecera y tramos, contrato OpenAPI, módulo schedule, pantalla y cliente generado.

**Alcance excluido**

- Bloqueos de descanso, almuerzo, no disponible, vacaciones, días libres y emergencia: HU-042.
- Citas afectadas, reprogramación, cancelación, notificaciones y agenda.
- Cambio de zona de la barbería o conversión de instantes existentes.
- Catálogo editable de festivos, calendarios externos y reglas comerciales.
- Cualquier FK o cascada del modelo físico distinta de ON DELETE RESTRICT (DEC-070).

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| CA-041-01 | Dado un barbero con el calendario desactivado, entonces los festivos no agregan bloqueos automáticos; la decisión no cambia el resultado de otro barbero. |
| CA-041-02 | Dado un barbero con el calendario activado, entonces un festivo queda cerrado por defecto en la resolución de jornada efectiva. |
| CA-041-03 | Dado un festivo abierto manualmente, entonces prevalece la excepción y el barbero puede trabajar la jornada completa o uno o varios tramos reducidos. |
| CA-041-04 | Una excepción cerrada no admite tramos; una excepción abierta exige tramos válidos, sin solapes, con intervalos semiabiertos y medianoche permitida por DEC-020. |
| CA-041-05 | Existe como máximo una excepción por barbero y fecha; crear, editar o retirar una excepción no modifica el horario semanal ni la excepción de otro barbero. |
| CA-041-06 | Una fecha, barbero o excepción de otra barbería no se expone y una escritura directa con el rol de aplicación queda bloqueada por tenant/RLS. |
| CA-041-07 | La resolución de jornada aplica la precedencia excepción manual abierta/cerrada > festivo automático > horario semanal y usa siempre la zona de la barbería. |
| CA-041-08 | La pantalla y el contrato comunican carga, vacío, error, éxito y el efecto del toggle; son operables por teclado y reflowan sin pérdida a 320, 360, 768 y 1280 px. |

**Pruebas obligatorias**

- Unitarias de calendario colombiano, toggle independiente, precedencia, fecha cerrada, horario especial, medianoche, contiguidad y solape.
- HTTP y contrato para toggle y CRUD de excepciones, 401, 404, 409, 422 y RFC 9457.
- PostgreSQL real con dos tenants y varios barberos, unicidad por fecha, cabecera cerrada sin tramos, RLS, FK en ON DELETE RESTRICT (DEC-070) y grants.
- Componente y E2E para activar, cerrar festivo, abrir festivo reducido, editar, retirar y comprobar aislamiento.
- Evidencia responsive y accesible en 320, 360, 768 y 1280 px, teclado, foco, zoom 200 %, axe-core y contraste.

**Terminado cuando** cada barbero puede representar su decisión independiente sobre festivos y fechas especiales, la resolución es reproducible y ninguna pantalla promete todavía citas o disponibilidad.

---

### HU-042 · Bloqueos de agenda

| Campo | Valor |
| --- | --- |
| Función | F-HOR-02, bloqueos puntuales, recurrentes y listas explícitas |
| Reglas | RN-TEN-01, RN-BLQ-01, RN-BLQ-03, RN-BLQ-04, RN-DIS-05, RN-DIS-07, RN-IDE-01 |
| Decisiones | DEC-007, DEC-008, DEC-009, DEC-013, DEC-019, DEC-020, DEC-024, DEC-033–DEC-040, DEC-043 |
| Actor | Barbero autenticado |
| Depende de | HU-040 y HU-041 integradas en main |
| Bloquea | Proyección completa de disponibilidad y la integración B3 de citas afectadas |
| Estado | Propuesta; prompt PROMPT-HU-042-v1 en draft; issue pending |
| Riesgo | Rechazar una emergencia por una cita existente, borrar el registro o duplicar una recurrencia destruye la confianza operativa del barbero. |

**Historia**

> Como barbero, quiero bloquear descansos, almuerzos, días libres, vacaciones, emergencias y otros intervalos, de forma puntual o recurrente, para que el sistema deje de ofrecer esos espacios y conserve la explicación de cada cambio.

**Alcance incluido**

- Siete tipos: break, lunch, unavailable, day_off, holiday, vacation y emergency.
- Bloqueo puntual y definiciones recurrentes semanales o por lista explícita de fechas.
- Rango de vigencia, excepciones individuales de una serie y edición con alcance explícito: esta instancia, esta y las siguientes o toda la serie.
- Intervalos semiabiertos, duración positiva, cruce de medianoche y unión de bloqueos solapados sin descontar dos veces.
- Eliminación lógica con actor e instante, conservación del registro y exclusión inmediata de la proyección efectiva.
- Creación de un bloqueo no rechazada por una posible cita; la consulta y resolución de citas afectadas queda en B3/B5.
- Contrato OpenAPI, módulo schedule, persistencia Atlas, RLS, pantalla, cliente tipado y pruebas reales.

**Alcance excluido**

- Crear, modificar, reprogramar, cancelar o notificar citas.
- Lista de citas afectadas, decisiones individuales y avisos de RN-BLQ-03: B3/B5.
- Cálculo de reserva pública, agenda y disponibilidad final.
- Borrado físico, purga, calendarios externos y funciones P1/P2.
- Cambiar horario semanal o excepciones de jornada: HU-040/HU-041.
- Crear una tabla appointment parcial o un adaptador que devuelva conteos inventados.
- Cualquier FK o cascada del modelo físico distinta de ON DELETE RESTRICT (DEC-070).

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| CA-042-01 | Cada uno de los siete tipos puede crearse y consultarse como bloqueo puntual o definición de serie válida, sin valores fuera del vocabulario aprobado. |
| CA-042-02 | Una serie semanal exige día ISO y una lista explícita conserva sus fechas en filas normalizadas; una excepción elimina solo la instancia elegida. |
| CA-042-03 | La edición ofrece esta instancia, esta y las siguientes o toda la serie; la opción elegida no pierde ni duplica instancias futuras. |
| CA-042-04 | Los intervalos validan duración, rango, contiguidad, medianoche y semántica [inicio, fin); los solapamientos de bloqueos se proyectan como una sola unión temporal. |
| CA-042-05 | Retirar un bloqueo o serie marca su eliminación, registra actor e instante, lo excluye de la proyección y conserva el registro; no existe DELETE físico para el rol de aplicación. |
| CA-042-06 | Crear un bloqueo no falla por la posible existencia de citas y no cancela, reprograma ni notifica ninguna cita; el punto de integración deja RN-BLQ-03 trazado a B3/B5. |
| CA-042-07 | La lectura y escritura están aisladas por barbería, los reintentos críticos no duplican efectos y una carrera real no deja dos retiros o dos series para una misma intención. |
| CA-042-08 | La pantalla diferencia punto/serie/lista, comunica estado retirado, conserva datos ante error y funciona con teclado, foco, zoom 200 % y anchos 320, 360, 768 y 1280 px. |

**Pruebas obligatorias**

- Unitarias de los siete tipos, punto, semanal, lista, excepciones, edición por alcance, solape, medianoche, unión, retiro lógico e idempotencia.
- HTTP y contrato para punto, serie, fechas, excepciones, retiro, 401, 404, 409, 422 y errores RFC 9457.
- PostgreSQL real con dos tenants, checks, RLS, fecha hija inválida, serie cerrada/no vigente, grants sin DELETE, actor de retiro y FK en ON DELETE RESTRICT (DEC-070).
- Concurrencia con dos conexiones reales y barreras observables, nunca sleep.
- Componente y E2E con los siete tipos, recurrencia, fecha explícita, excepción, retiro y aislamiento.
- Evidencia responsive y accesible en 320, 360, 768 y 1280 px, teclado, foco, zoom 200 %, axe-core y contraste.
- Evidencia separada de que esta HU no consulta ni crea appointment y deja la lista de afectados para B3/B5.

**Terminado cuando** el barbero puede bloquear tiempo de forma puntual, recurrente y auditable sin que el sistema borre registros ni resuelva citas por su cuenta, y la parte pendiente de RN-BLQ-03 queda enlazada a las historias dueñas.

---


## 6. Historias pendientes de redacción

| Bloque | Rango reservado | Se redacta cuando |
| --- | --- | --- |
| B1 | `HU-025` – | `HU-020`–`HU-024` implementadas (`DEC-067`–`DEC-069` propagadas); redactar lo restante solo después de revisar el criterio de salida de B1 |
| B2 | `HU-040` – `HU-042` | `CT-008` resuelta (`DEC-070`); `HU-040` en ejecución (issue real `#90`); `HU-041`–`HU-042` requieren issue real propio antes de ejecutarse |
| B3 | `HU-060` – | B2 cumple su criterio de salida |
| B4 | `HU-090` – | B3 cumple su criterio de salida |
| B5 | `HU-130` – | B4 cumple su criterio de salida |
| B6 | `HU-150` – | B5 cumple su criterio de salida |

Redactar por anticipado las historias de un bloque lejano produce texto que hay que reescribir: lo que se aprende construyendo el bloque anterior cambia la historia siguiente.

---

## 7. Dudas que bloqueaban historias de B0

| Duda | Historia antes bloqueada | Qué faltaba decidir | Resuelta como |
| --- | --- | --- | --- |
| `DP-SEG-04` | `HU-005`, `HU-006` | Mecanismo concreto de la sesión larga y su duración exacta | `DEC-050` |
| `DP-SEG-05` | `HU-008`, `HU-011` | Canal y proveedor del código de recuperación | `DEC-051` |
| `DP-SEG-06` | `HU-007` | Duración de la ventana del límite por IP y del escalamiento | `DEC-052` |

Las tres se resolvieron el 2026-08-11 (ver [dudas-pendientes.md](../00-control/dudas-pendientes.md) y [registro-decisiones.md](../00-control/registro-decisiones.md)). `HU-005`–`HU-008` y `HU-011` ya no dependen de una duda abierta; siguen sujetas al criterio de salida de B0 antes de implementarse.
