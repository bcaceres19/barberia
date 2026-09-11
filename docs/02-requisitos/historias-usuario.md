---
titulo: "Historias de usuario y criterios de aceptación"
version: "1.38"
estado: "Propuesta"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-09-10"
documentos_relacionados:
  - "../01-producto/alcance-mvp.md"
  - "../01-producto/reglas-negocio.md"
  - "estados-citas.md"
  - "../10-backlog/plan-bloques.md"
  - "../10-backlog/prompts/README.md"
  - "../00-control/matriz-trazabilidad.md"
  - "../00-control/dudas-pendientes.md"
  - "../00-control/contradicciones.md"
---

# Historias de usuario y criterios de aceptación

> **Estado del contenido: Propuesta.** Estas historias **derivan** de funciones P0 y reglas confirmadas; no crean alcance por sí solas y requieren aprobación antes de implementarse. B0, `HU-020`–`HU-024` y `HU-040`–`HU-042` ya están integradas en `main`. Las dudas históricas de B0 quedaron resueltas por `DEC-050`–`DEC-066`; `DP-SER-01`–`DP-SER-03` del lote de B1 quedaron resueltas por `DEC-067`–`DEC-069`; `CT-008` del lote de B2 quedó resuelta por `DEC-070`. Los issues [#90](https://github.com/bcaceres19/barberia/issues/90), [#95](https://github.com/bcaceres19/barberia/issues/95), [#98](https://github.com/bcaceres19/barberia/issues/98) y [#100](https://github.com/bcaceres19/barberia/issues/100) conservan seguimientos visuales/E2E parciales de B2. `HU-060`–`HU-062` están integradas en `main` mediante [PR #105](https://github.com/bcaceres19/barberia/pull/105), [PR #109](https://github.com/bcaceres19/barberia/pull/109) y [PR #113](https://github.com/bcaceres19/barberia/pull/113); `DP-CIT-01`–`DP-CIT-05` quedaron resueltas por `DEC-071`–`DEC-075`. `HU-063` (navegación de la agenda por fecha) tiene issue real [#116](https://github.com/bcaceres19/barberia/issues/116) y está implementada en `feat/116-hu063-navegacion-agenda`, integrada en `main` mediante [PR #118](https://github.com/bcaceres19/barberia/pull/118). `HU-064` (detalle e historial) tiene issue real [#120](https://github.com/bcaceres19/barberia/issues/120) e integrada en `main` mediante [PR #121](https://github.com/bcaceres19/barberia/pull/121). `DP-CIT-06` quedó resuelta el 2026-09-01 como `DEC-076` (bloqueo duro frente a un intervalo bloqueado, mismo tratamiento que un cruce de citas); `HU-065` (reprogramación T2) está implementada contra el issue real [#123](https://github.com/bcaceres19/barberia/issues/123), integrada en `main` mediante [PR #125](https://github.com/bcaceres19/barberia/pull/125).

---

## 1. Cómo leer este documento

Cada historia lleva un código `HU-{nnn}` estable y sus criterios de aceptación llevan `CA-{HU}-{nn}`, según [glosario.md](../00-control/glosario.md), sección 8. **Los códigos no se renumeran ni se reutilizan.**

Una historia describe una capacidad verificable de principio a fin. No es una tarea técnica; una tarea que no se puede demostrar delante del propietario no es una historia.

Los criterios de aceptación son la **definición contractual de terminado**. Si un criterio no se puede comprobar con una prueba automatizada o con una evidencia concreta, está mal redactado.

La secuencia de bloques vive en [plan-bloques.md](../10-backlog/plan-bloques.md). El prompt operativo individual de cada historia vive en el [catálogo de prompts persistentes](../10-backlog/prompts/README.md); `prompts-implementacion.md` se conserva solo como antecedente histórico.

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
| B2 · Horario laboral y bloqueos | `HU-040` – `HU-042` | Integradas en `main` ([PR #93](https://github.com/bcaceres19/barberia/pull/93), [PR #96](https://github.com/bcaceres19/barberia/pull/96), [PR #99](https://github.com/bcaceres19/barberia/pull/99)); seguimientos parciales en `#90`, `#95`, `#98` y `#100` |
| B3 · Agenda, estados e integridad | `HU-060` – `HU-068` | `HU-060`–`HU-068` integradas en `main`, issues reales [#225](https://github.com/bcaceres19/barberia/issues/225)–[#227](https://github.com/bcaceres19/barberia/issues/227); tercer lote de B3 completo, T3/cierre automático pendientes |
| B4 · Reserva pública y disponibilidad | `HU-090` – `HU-099` | Redactadas como propuesta en issue [#240](https://github.com/bcaceres19/barberia/issues/240); implementación bloqueada por decisiones `DP-PUB-01`–`DP-PUB-06`/`CT-011` |
| B5 · Notificaciones y recordatorios | `HU-130` – | Pendientes |
| B6 · Operación, privacidad y piloto | `HU-150` – | Pendientes |

---

## 3. Bloque B0 · Cimientos, seguridad y primeras pantallas

Orden de construcción recomendado: `HU-001` → `HU-002` → `HU-003` → `HU-004` → `HU-009` → `HU-005` → `HU-006` → `HU-010` → `HU-012` → `HU-007` → `HU-008` → `HU-011`.

La base transversal de experiencia (`HU-009`) se adelanta a las pantallas para compartir comportamiento accesible y estados. Cada rediseño o pantalla nueva conserva los mockups y la firma cromática NAVA, sin impedir que la pantalla resuelva libremente su composición, escala y tecnología de estilos.

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
| Decisiones | `DEC-026`, `DEC-052`, `DEC-061`, `DEC-062`, `DEC-081` |
| Actor | Propietario (protege), barbero (afectado si se excede) |
| Depende de | `HU-003`, `HU-005` |
| Bloquea | — |
| Riesgo | Sin límite, un ataque automatizado prueba miles de contraseñas; con un límite mal calibrado, el barbero legítimo queda fuera en plena jornada. |

**Historia**

> Como propietario del sistema, necesito que el formulario de acceso limite los intentos por IP y exija una prueba adicional cuando se supera el umbral, para frenar el abuso sin castigar al barbero que se equivocó dos veces.

> **Bloqueo resuelto:** `DEC-026` fijaba el umbral inicial de **5 solicitudes por IP** sin la duración de la ventana ni la del escalamiento. `DP-SEG-06` quedó resuelta el 2026-08-11 como `DEC-052`: ventana de 15 minutos, escalamiento a verificación adicional de 24 horas. `CT-005` y `DP-SEG-10` quedaron resueltas el 2026-08-17 como `DEC-061`/`DEC-062`: las cinco primeras solicitudes se evalúan con normalidad y la sexta exige completar un código de 6 dígitos, 5 min y 5 intentos antes de evaluar la contraseña. `DEC-081` sustituyó el WhatsApp único: correo es el canal predeterminado y la configuración del evento puede usar correo, WhatsApp oficial o ambos, siempre sobre contactos verificados resueltos por el servidor.

**Alcance incluido**

- Conteo por IP con ventana de 15 minutos y umbral de 5 solicitudes (`DEC-052`), ambos configurables.
- Escalamiento: al superar el umbral (la sexta solicitud, `DEC-061`), esa solicitud y las siguientes exigen completar el reto OTP de `DEC-062`/`DEC-081` antes de evaluar la contraseña, durante 24 horas o hasta completarlo con éxito.
- Respuesta `429` con formato uniforme y con indicación de cuándo reintentar.
- Configuración expuesta como parámetros, no como números incrustados en el código.
- Los contadores no almacenan datos personales; la IP se guarda de forma acotada y con vencimiento.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-007-01` | Dadas hasta cinco solicitudes desde la misma IP dentro de la ventana, entonces todas se evalúan normalmente, incluida la contraseña (`DEC-061`). |
| `CA-007-02` | La sexta solicitud desde esa IP dentro de la ventana exige completar el reto OTP de `DEC-062`/`DEC-081` y no evalúa la contraseña hasta lograrlo; canal y destino se resuelven en servidor desde configuración y contactos verificados. |
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
| Decisiones | `DEC-026`, `DEC-051`, `DEC-063`, `DEC-064`, `DEC-065`, `DEC-081` |
| Actor | Barbero |
| Depende de | `HU-005`, `HU-007` |
| Bloquea | `HU-011` |
| Riesgo | Un barbero sin acceso en plena jornada pierde su agenda; un mecanismo de recuperación débil es la vía más común de secuestro de cuentas. |

**Historia**

> Como barbero que olvidó su contraseña, quiero recuperar el acceso con un código enviado por los canales configurados, para volver a mi agenda el mismo día sin depender de que alguien me responda.

> **Bloqueo resuelto:** `DEC-026` definía el mecanismo (código al teléfono verificado) sin fijar canal ni proveedor. `DP-SEG-05` quedó resuelta el 2026-08-11 como `DEC-051`: WhatsApp oficial y correo, reutilizando el proveedor ya habilitado por `DEC-027`. El 2026-08-17 se resolvieron las últimas cuatro dudas: `DP-SEG-11` (política de contraseña) como `DEC-063`, `DP-SEG-12` (formato/vigencia/intentos/token de reinicio del código) como `DEC-064`, `CT-006` (respuesta idéntica frente a destino enmascarado) como `DEC-065`, y `DP-NOT-05` (proveedor/adaptador real de WhatsApp y correo: Meta Cloud API + Resend) como `DEC-066`. `HU-008` ya no depende de ninguna decisión pendiente; solo espera que `HU-007` se integre en `main`.

**Alcance incluido**

- Migración con la tabla de códigos de recuperación: hash del código, vencimiento corto, intentos, marca de uso, `barbershop_id` y RLS.
- Solicitud de recuperación, verificación del código y establecimiento de contraseña nueva.
- Envío según la configuración del evento: correo por defecto, WhatsApp oficial o ambos (`DEC-081`), con los proveedores oficiales de `DEC-066`.
- Código de un solo uso, con vencimiento, con límite de intentos y con reenvío controlado.
- Invalidación de las sesiones activas al cambiar la contraseña.
- Destino mostrado enmascarado en la interfaz y en las respuestas.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-008-01` | Dado un correo registrado, cuando se solicita recuperación, entonces se envía un mismo código por el canal o canales configurados sobre contactos verificados —correo por defecto, WhatsApp oficial o ambos— y la respuesta de `POST /recovery/request` es idéntica a la de un correo no registrado, sin destino en el cuerpo (`DEC-065`, `DEC-081`). |
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

### HU-009 · Base de experiencia accesible y dirección visual NAVA

| Campo | Valor |
| --- | --- |
| Función | Soporte transversal de todas las pantallas P0 |
| Reglas | Criterios no funcionales de UX y accesibilidad |
| Decisiones | `DEC-033`, `DEC-035`, `DEC-077`, `DEC-078`, `DEC-079` |
| Actor | Barbero y cliente (indirectamente) |
| Depende de | — |
| Bloquea | `HU-010`, `HU-011`, `HU-012` y toda pantalla posterior |
| Riesgo | Sin una base transversal de comportamiento, cada pantalla puede resolver foco, estados y errores de forma incompatible. La solución visual, en cambio, puede ser propia de cada pantalla siempre que conserve NAVA / Tailored Grid. |

**Historia**

> Como barbero que trabaja de pie y con las manos ocupadas, quiero que todos los controles de la aplicación se vean y se comporten igual, con áreas táctiles cómodas y textos legibles, para no equivocarme entre cliente y cliente.

**Alcance incluido**

- Componentes de comportamiento en `apps/web/src/shared/ui/` que realmente usarán las pantallas de B0: botón, campo de texto, alerta, insignia y diálogo. **No se crean componentes sin uso real.**
- Variantes tipadas cuando simplifiquen comportamiento o accesibilidad; la composición visual puede definirse en el componente, módulo o pantalla según el problema que resuelva.
- Patrones de estado de pantalla: inicial, carga, actualización, vacío, error recuperable, error de campo, conflicto y éxito.
- Foco visible, orden de tabulación coherente, controles cómodos, contraste verificable y respeto de `prefers-reduced-motion`.
- Evidencia responsive en 320, 360, 768 y 1280 px.

**Alcance excluido**

- Una configuración funcional de temas o personalización por barbería: requiere una decisión de producto y no nace del mockup.
- Una dependencia visual nueva sin justificar licencia, accesibilidad, mantenimiento e impacto en el bundle.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-009-01` | Todo rediseño o pantalla nueva respeta los mockups, la firma cromática y el contraste editorial/funcional NAVA; composición, escala, nombres de tokens y tecnología permanecen libres. |
| `CA-009-02` | Cada componente interactivo tiene foco visible conforme al estándar y es operable solo con teclado. |
| `CA-009-03` | Los controles son cómodos de usar y conservan el objetivo de usabilidad de 44 × 44 px cuando la composición lo permite; cualquier excepción mantiene operabilidad, foco y legibilidad. |
| `CA-009-04` | Cada componente tiene prueba de componente que cubre sus estados: normal, deshabilitado, cargando y error cuando apliquen. |
| `CA-009-05` | El contraste real de texto y señales necesarias cumple WCAG 2.2 AA, verificado sobre las combinaciones elegidas por la pantalla. |
| `CA-009-06` | Existe evidencia visual en los cuatro anchos de referencia y ninguna pantalla produce desplazamiento horizontal. |
| `CA-009-07` | Las animaciones, si existen, apoyan la tarea y se reducen o eliminan con `prefers-reduced-motion`; no se exige una duración visual fija. |

**Pruebas obligatorias**

- Pruebas de componente con la herramienta definida en `estrategia-pruebas.md` sección 5.2.
- Verificación accesible automatizada y revisión manual con teclado.

**Terminado cuando** las pantallas de `HU-010` a `HU-012` pueden reutilizar el comportamiento accesible de la base y expresar NAVA / Tailored Grid sin que una regla visual rígida les impida resolver su propio contenido.

---

### HU-010 · Pantalla de acceso

| Campo | Valor |
| --- | --- |
| Función | `F-AUTH-01` |
| Reglas | Criterios no funcionales de UX; `RN-DAT-02` |
| Decisiones | `DEC-033`, `DEC-056`, `DEC-078` |
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
| Decisiones | `DEC-026`, `DEC-078` |
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
| Decisiones | `DEC-033`, `DEC-056`, `DEC-060`, `DEC-078` |
| Actor | Barbero |
| Depende de | `HU-006`, `HU-009`, `HU-010` |
| Bloquea | Todas las pantallas privadas de B1 en adelante |
| Riesgo | Si cada pantalla resuelve por su cuenta la sesión, la navegación y los errores, el mismo problema se implementa siete veces de siete maneras. |

**Historia**

> Como barbero, quiero que la aplicación recuerde mi sesión, me lleve a mi agenda al entrar, me muestre siempre en qué barbería estoy y me avise con claridad cuando se pierda la conexión, para no quedarme mirando una pantalla en blanco.

**Alcance incluido**

- Estructura de aplicación autenticada: contexto de la barbería, navegación principal y área de contenido, con una composición visual libre dentro de NAVA / Tailored Grid, construida sobre la ruta `/panel` y el guard mínimo que ya entrega `HU-010` (`DEC-056`) — sin crear un guard paralelo ni cambiar la ruta.
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
| Decisiones | `DEC-007`, `DEC-024`, `DEC-033`, `DEC-037`, `DEC-078` |
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
- Personalización funcional de colores, tipografías o tema por barbería: no forma parte de esta HU y requiere una decisión de producto; la composición interna de la pantalla sí es libre dentro de NAVA / Tailored Grid.

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
| Decisiones | `DEC-019`, `DEC-024`, `DEC-033`, `DEC-037`, `DEC-078` |
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

> **B2 integrado con seguimientos abiertos.** Las tres historias separan las dos funciones P0 del bloque: HU-040 configura la jornada semanal; HU-041 resuelve fechas especiales y festivos; HU-042 administra bloqueos puntuales y recurrentes. CT-008 quedó resuelta como DEC-070 (las FK del modelo físico de B2 usan ON DELETE RESTRICT, no CASCADE). HU-040 está integrada en main (PR #93, issue real #90 abierto por CA-040-08 parcial); HU-041 está integrada en main (PR #96, issue real #95 abierto por CA-041-08 parcial); HU-042 está integrada en main (PR #99, issue real #98 con seguimiento de UI/E2E en #100).

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
| Estado | Integrada en main (PR #93, squash-merge 2026-08-25, CI 4/4 en verde); prompt PROMPT-HU-040-v1 executed; issue real #90 abierto por CA-040-08 parcial |
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
| Estado | Integrada en `main`; prompt PROMPT-HU-041-v1.3 executed; issue real [#95](https://github.com/bcaceres19/barberia/issues/95), abierto por `CA-041-08` parcial; [PR #96](https://github.com/bcaceres19/barberia/pull/96) |
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
| Estado | Integrada en `main` mediante [PR #99](https://github.com/bcaceres19/barberia/pull/99); prompt `PROMPT-HU-042-v1` `executed`; issues [#98](https://github.com/bcaceres19/barberia/issues/98) y [#100](https://github.com/bcaceres19/barberia/issues/100) conservan seguimientos parciales |
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


## 6. Bloque B3 · Agenda, estados e integridad

> **B3 en ejecución.** Las primeras tres historias separaron la última defensa de datos (`HU-060`), la primera escritura vertical de negocio (`HU-061`) y la pantalla operativa mínima (`HU-062`); las tres están integradas. El segundo lote amplía esa base en cortes independientes: navegación por fecha (`HU-063`), detalle con historial (`HU-064`) y reprogramación auditada T2 (`HU-065`). `HU-063` tiene issue real ([#116](https://github.com/bcaceres19/barberia/issues/116)) y está implementada, integrada en `main` mediante [PR #118](https://github.com/bcaceres19/barberia/pull/118). `HU-064` tiene issue real ([#120](https://github.com/bcaceres19/barberia/issues/120)) e integrada en `main` mediante [PR #121](https://github.com/bcaceres19/barberia/pull/121). `HU-065` tiene issue real ([#123](https://github.com/bcaceres19/barberia/issues/123)) e integrada en `main` mediante [PR #125](https://github.com/bcaceres19/barberia/pull/125) (`DP-CIT-06` resuelta como `DEC-076`).

Orden recomendado: `HU-060` → `HU-061` → `HU-062` → `HU-063` → `HU-064` → `HU-065`. La edición de servicio/datos (T3), cancelación, completar, marcar inasistencia, corregir estados y cierre automático se redactan después de verificar este segundo lote.

---

### HU-060 · Núcleo persistente de citas sin cruces

| Campo | Valor |
| --- | --- |
| Función | `F-DISP-02`; base técnica de `F-CITA-*` y `F-EST-03` |
| Reglas | `RN-CON-01`, `RN-CON-03`, `RN-DIS-05`, `RN-DIS-07`, `RN-HIS-01`, `RN-HIS-02`, `RN-RES-02`, `RN-RES-03`, `RN-TEN-01` |
| Decisiones | `DEC-002`, `DEC-004`, `DEC-007`, `DEC-014`, `DEC-016`, `DEC-019`, `DEC-020`, `DEC-024`, `DEC-035`–`DEC-041`, `DEC-045`, `DEC-046`, `DEC-070` |
| Actor | Propietario del producto y capacidades futuras de citas |
| Depende de | `HU-040`–`HU-042` integradas; criterio de salida de B2 verificado; modelo de referencia revisado, no copiado ciegamente |
| Bloquea | `HU-061` y toda escritura, lectura o transición de citas de B3–B5 |
| Estado | Integrada en `main` (rama `feat/104-hu060-nucleo-citas`, issue real [#104](https://github.com/bcaceres19/barberia/issues/104), integrada mediante [PR #105](https://github.com/bcaceres19/barberia/pull/105)); prompt `PROMPT-HU-060-v1` `executed` |
| Riesgo | Sin una exclusión real y un historial append-only, dos procesos pueden guardar turnos cruzados o reescribir la evidencia operativa. |

**Historia**

> Como propietario del producto, quiero que clientes, citas y su historial inicial se persistan con aislamiento por barbería y exclusión temporal en PostgreSQL, para que ninguna interfaz o proceso pueda guardar turnos cruzados ni alterar el rastro original.

**Alcance incluido**

- Migración Atlas mínima para `customer`, `appointment`, `appointment_history` y `appointment_history_change`, revisada contra la sección D del modelo físico de referencia.
- Cinco estados técnicos cerrados; `confirmed`, `completed` y `no_show` ocupan agenda, mientras las dos cancelaciones la liberan.
- Intervalos `timestamptz` semiabiertos `[inicio, fin)`, duración planificada coherente con el snapshot y exclusión GiST por barbería y barbero solo para estados que ocupan agenda.
- Snapshots aprobados de nombre, duración, precio y moneda del servicio; un cambio posterior de catálogo no los sincroniza en silencio.
- FK tenant-aware, RLS forzada, índices de agenda/historial, privilegios mínimos y ausencia de borrado físico de citas e historial para `barberia_app`.
- Operación interna transaccional que pueda crear cliente, cita `confirmed` y evento `appointment_created` como una unidad, sin exponer todavía un endpoint.
- Fixtures y pruebas con PostgreSQL 14 real, dos tenants, inserción directa y dos conexiones concurrentes.

**Alcance excluido**

- Endpoint HTTP, formulario Vue, agenda visible o creación pública.
- Decidir la reconciliación de clientes manuales sin teléfono (`DP-CIT-01`). La base solo conserva formas válidas y permite ausencia significativa.
- Validar jornada efectiva, bloqueos, asignación servicio-barbero o permisos de una cita concreta; pertenecen al caso de uso de `HU-061`.
- Modificar, reprogramar, cancelar, completar, marcar inasistencia o corregir estados.
- Tokens públicos, recordatorios, notificaciones, cierre automático, anonimización o métricas del piloto.
- Copiar el DDL de referencia sin reevaluar bloqueo, datos existentes, RLS, índices y recuperación conforme a Atlas.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-060-01` | El esquema migrado representa `customer`, `appointment` y el historial normalizado con FK tenant-aware, RLS forzada y sin `jsonb`, cascadas ni datos duplicados fuera de los snapshots aprobados. |
| `CA-060-02` | Una inserción SQL directa de dos citas que ocupan agenda y se cruzan para el mismo barbero es rechazada por PostgreSQL; dos citas contiguas y dos citas simultáneas de barberos distintos son aceptadas. |
| `CA-060-03` | El criterio derivado de ocupación incluye exactamente `confirmed`, `completed` y `no_show`; una cita cancelada no participa en la exclusión, sin repetir una lista divergente en otra restricción. |
| `CA-060-04` | `ends_at - starts_at` coincide con `duration_minutes_snapshot`, incluso al cruzar medianoche o una transición DST, y los instantes nunca dependen de la zona de sesión. |
| `CA-060-05` | La creación interna persiste cliente, cita `confirmed` y `appointment_created` en una sola transacción; cualquier fallo revierte las tres escrituras. |
| `CA-060-06` | `appointment_history` y sus cambios aceptan solo inserción para el rol de aplicación; `UPDATE` y `DELETE` directos fallan y una corrección futura solo podrá añadir otra entrada. |
| `CA-060-07` | Con el rol real y dos tenants, leer o relacionar IDs ajenos devuelve cero filas o falla por integridad/RLS sin revelar existencia; teléfono y correo solo son únicos dentro de la barbería conforme a `DEC-045`/`DEC-046`. |
| `CA-060-08` | La migración valida y aplica desde vacío y desde la versión anterior con datos representativos, conserva `atlas.sum` y documenta bloqueo, avance correctivo y pruebas en README/matriz. |

**Pruebas obligatorias**

- SQL directo de estados, duración, snapshots, contigüidad, cruces total/parciales, multi-barbero y cancelación que libera la restricción.
- PostgreSQL real con dos tenants y `barberia_app`: RLS, FK compuestas, grants, ausencia de `DELETE`, historial append-only y cierre por falta de contexto.
- Dos conexiones reales que intentan crear el mismo cruce; exactamente una persiste, coordinadas con barreras observables y nunca con `sleep`.
- Repositorio/transacción Go para éxito y rollback de cliente+cita+historial, sin Chi ni detalles HTTP.
- Atlas sobre base vacía y actualización desde la migración anterior con datos, incluido segundo `apply` sin cambios.

**Terminado cuando** PostgreSQL hace imposible almacenar citas cruzadas que ocupan agenda y existe una primitiva transaccional auditada lista para que `HU-061` exponga la creación manual, sin adelantar UI ni transiciones.

---

### HU-061 · Creación manual de turnos

| Campo | Valor |
| --- | --- |
| Función | `F-CITA-03` |
| Reglas | `RN-CIT-02`, `RN-RES-01`–`RN-RES-03`, `RN-CON-01`, `RN-CON-03`, `RN-DIS-04`–`RN-DIS-07`, `RN-HIS-01`, `RN-TEN-01`, `RN-DAT-01`, `RN-DAT-02`, `RN-IDE-01` |
| Decisiones | `DEC-002`, `DEC-004`–`DEC-007`, `DEC-016`, `DEC-019`, `DEC-020`, `DEC-024`, `DEC-035`–`DEC-040`, `DEC-043`, `DEC-045`, `DEC-046`, `DEC-071`, `DEC-072`, `DEC-073` |
| Actor | Barbero autenticado |
| Depende de | `HU-060` integrada (PR #105); B2 integrado y verificado |
| Bloquea | `HU-062`, modificaciones, reprogramaciones, cancelación y recorridos posteriores de B3 |
| Estado | Implementada; issue real [#107](https://github.com/bcaceres19/barberia/issues/107); prompt `PROMPT-HU-061-v1` `executed` |
| Riesgo | Una cita manual que salte asignaciones, jornada, bloqueos o reconciliación de cliente puede reservar tiempo imposible, duplicar personas o contradecir la agenda pública futura. |

**Historia**

> Como barbero, quiero registrar desde el celular un turno recibido por WhatsApp, teléfono o en persona, para que ocupe la agenda con las mismas garantías de integridad que una reserva pública aunque no tenga todos los datos de contacto.

**Alcance incluido**

- Operación privada contract-first para crear una cita manual `confirmed`, protegida por sesión, tenant e `Idempotency-Key`.
- Selección de barbero, servicio (solo activos ya asignados al barbero elegido, `DEC-072`), persona atendida, fecha/hora y contacto opcional; reconciliación de `customer` por teléfono (`DEC-045`) o por correo cuando falta el teléfono (`DEC-071`), y rechazo si el intervalo coincide con un bloqueo vigente (`DEC-073`).
- Cualquier minuto válido, incluida una cita que ya empezó, sin aplicar anticipación mínima, ventana máxima ni rejilla pública.
- Fin planificado derivado de la duración vigente del servicio y snapshots inmutables de servicio/precio/moneda al crear.
- Transacción única de cliente, cita e historial usando el núcleo de `HU-060`; conflicto uniforme si PostgreSQL rechaza un cruce.
- Formulario “Nuevo turno” dentro del panel privado, con resumen, zona IANA de la barbería y conservación de datos no sensibles ante error.
- Punto de integración explícito con `schedule` y `catalog`, sin leer tablas ajenas desde el núcleo `booking`.

**Alcance excluido**

- Reinterpretar `DEC-071`–`DEC-073` durante la implementación; se aplican literalmente.
- Reserva pública, consulta de franjas alternativas, token de cliente o cancelación pública.
- Editar, reprogramar, cancelar, completar, marcar inasistencia o corregir una cita existente.
- Enviar confirmaciones o crear la maquinaria de recordatorios; B5 conserva esa preocupación. La ausencia de B5 se documenta, no se simula como envío exitoso.
- Crear servicios, barberos, horarios o bloqueos desde el formulario de cita.
- Permitir un `barbershopId`, estado, precio, moneda, duración snapshot o actor confiado desde el body.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-061-01` | Con sesión válida y datos permitidos por `DEC-071`–`DEC-073`, crear guarda exactamente una cita `confirmed` y un único `appointment_created`, y la devuelve con identificador opaco. |
| `CA-061-02` | La cita manual acepta cualquier minuto válido y queda exenta de anticipación mínima, ventana máxima y rejilla pública; un instante equivocado por zona o un intervalo fuera de la jornada permitida se rechaza sin escritura parcial. |
| `CA-061-03` | El servidor deriva servicio, duración, precio, moneda, tenant, actor y fin planificado desde fuentes autorizadas; el body no puede sobrescribir snapshots ni pertenencia. |
| `CA-061-04` | La identidad/reutilización de `customer` (`DEC-071`), el servicio permitido para el barbero (`DEC-072`) y el efecto de un bloqueo vigente (`DEC-073`) coinciden literalmente con esas decisiones, con pruebas de cada rama aprobada. |
| `CA-061-05` | Una cita que se cruza con otra que ocupa agenda responde `409` uniforme; contigüidad exacta es válida y la base de datos sigue siendo la última defensa bajo carrera. |
| `CA-061-06` | Repetir la misma clave e intención devuelve el resultado lógico original sin duplicar cliente, cita ni historial; una clave concurrente o reutilizada con otro contenido responde según `DEC-043`/`RN-IDE-01`. |
| `CA-061-07` | Recurso ajeno o inexistente responde `404` seguro; dos tenants reales no pueden leer ni escribir la cita, cliente, servicio o barbero del otro. |
| `CA-061-08` | El formulario usa “turno”, conserva campos no sensibles, evita doble toque, gestiona carga/error/conflicto/éxito, funciona con teclado y foco y reflowa a 320, 360, 768 y 1280 px con zoom 200 %. |

**Pruebas obligatorias**

- Dominio/aplicación para duración, zona, horario, datos opcionales y las tres ramas de `DEC-071`–`DEC-073`.
- HTTP/contrato para `201`, repetición, `400`, `401`, `404`, `409`, `422`, cuerpo excesivo, campos desconocidos y RFC 9457.
- PostgreSQL real con dos tenants, rollback atómico, contigüidad, cruce, cita ya iniciada y carrera de dos conexiones.
- Componente para validación, doble envío, foco, conflicto y conservación del formulario; `vitest-axe` sin violaciones conocidas.
- E2E real del barbero que crea un turno manual fuera de la rejilla pública, recarga y observa una sola cita, con evidencia responsive/accesible.

**Terminado cuando** el barbero puede crear un único turno manual auditable desde el panel, con decisiones de cliente/servicio/bloqueo aprobadas y sin adelantar edición, estados, reserva pública o notificaciones.

---

### HU-062 · Agenda diaria de hoy

| Campo | Valor |
| --- | --- |
| Función | `F-CITA-01`; prepara `F-CITA-02` sin implementar navegación por fecha |
| Reglas | `RN-CIT-01`, `RN-RES-02`, `RN-RES-03`, `RN-DIS-07`, `RN-TEN-01` y estados confirmados de `estados-citas.md` |
| Decisiones | `DEC-007`, `DEC-016`, `DEC-019`, `DEC-024`, `DEC-033`–`DEC-038`, `DEC-047`, `DEC-078` |
| Actor | Barbero autenticado |
| Depende de | `HU-061` integrada (cumplida); `DP-CIT-04` y `DP-CIT-05` resueltas como `DEC-074`–`DEC-075` (cumplida); issue real [#111](https://github.com/bcaceres19/barberia/issues/111) (cumplida) |
| Bloquea | Navegación por fecha, detalle, modificación, reprogramación, cancelación y transiciones visibles |
| Estado | Implementada contra el issue real [#111](https://github.com/bcaceres19/barberia/issues/111): backend y frontend con pruebas reales; `CA-062-08` (E2E/evidencia responsiva) pendiente de ejecución contra un stack real, mismo estado que `HU-061` |
| Riesgo | Una agenda que usa la zona del dispositivo, oculta estados terminales o elige un barbero implícito muestra una operación distinta de la realidad persistida. |

**Historia**

> Como barbero, quiero abrir el panel y ver cronológicamente los turnos de hoy con su hora, persona, servicio y estado, para operar la jornada desde el celular sin buscar en otras pantallas.

**Alcance incluido**

- Consulta privada de la agenda de una fecha civil en la zona IANA de la barbería, con un turno nocturno visible en cada agenda cuyo rango interseca su intervalo (`DEC-075`) y paginada o acotada conforme al volumen real del día.
- Apertura predeterminada en hoy con selector obligatorio de un barbero (`DEC-074`), sin vista consolidada ni vínculo inferido entre `staff_user` y `barber`.
- Lista cronológica accesible con hora, persona atendida, servicio snapshot, barbero cuando corresponda y etiqueta visible de cada estado.
- Citas activas y terminales del día; las terminales pueden atenuarse o separarse, pero no desaparecen ni pierden legibilidad.
- Acción “Nuevo turno” enlazada al formulario real de `HU-061`.
- Estados de carga inicial, actualización, vacío y error recuperable; fecha completa y zona visibles.
- Cliente TypeScript derivado de OpenAPI, módulo Vue `agenda` y pruebas de consulta/contrato/componente/E2E.

**Alcance excluido**

- Anterior, siguiente y selector de fecha (`F-CITA-02`); esta historia solo abre y recarga hoy.
- Detalle completo, teléfono, correo, nota, historial o acciones sobre una cita.
- Crear desde un hueco espacial, arrastrar, calendario semanal/mensual o tiempo real.
- Modificar, reprogramar, cancelar, completar, marcar `no_show` o corregir estados.
- Añadir un enlace automático `staff_user`→`barber` prohibido por `DEC-047`, o cualquier vista consolidada de varios barberos (`DEC-074` la excluye).
- Cálculo de disponibilidad pública, huecos ofrecibles, recordatorios o notificaciones.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-062-01` | Al entrar al panel, la agenda consulta y muestra la fecha civil completa de hoy en la zona de la barbería, aunque dispositivo y servidor usen otras zonas. |
| `CA-062-02` | La vista exige un `barberId` explícito (selector obligatorio, `DEC-074`); nunca presupone que el usuario autenticado sea una fila `barber` ni ofrece una vista consolidada de varios barberos. |
| `CA-062-03` | Los turnos se ordenan por `startsAt` y cada fila muestra hora, persona atendida, servicio snapshot, barbero cuando corresponda y una etiqueta de estado en español que no depende solo del color. |
| `CA-062-04` | Las citas activas y terminales que cruzan medianoche aparecen en cada agenda diaria cuyo rango civil interseca su intervalo (`DEC-075`), sin duplicación dentro de la misma agenda ni pérdida en los extremos `[inicio, fin)`. |
| `CA-062-05` | La respuesta y la interfaz no exponen teléfono, correo, nota ni campos internos innecesarios; un tenant no obtiene filas de otro y un alcance ajeno responde de forma segura. |
| `CA-062-06` | Carga, actualización, vacío y error recuperable conservan contexto y anuncian el estado; “Nuevo turno” navega al flujo real de `HU-061`. |
| `CA-062-07` | La consulta usa índice y límite verificables para el acceso por barbería, alcance de barbero y rango diario; `EXPLAIN (ANALYZE, BUFFERS)` evita un recorrido completo con volumen representativo. |
| `CA-062-08` | En 320 y 360 px existe una lista completa operable con una mano y objetivos de 44 px; 768/1280 px, teclado, foco, zoom 200 %, axe-core y contraste no pierden información ni acciones. |

**Pruebas obligatorias**

- Dominio/modelo de vista para día civil, zona, orden, medianoche y traducción de estados.
- HTTP/contrato para hoy con agenda vacía/llena, `401`, `404`, límites, campos exactos y dos tenants.
- PostgreSQL real para límites del día, citas nocturnas conforme a `DEC-075`, estados terminales, aislamiento y plan de consulta con volumen representativo.
- Componente con cero, una y varias citas, actualización/error, estados visibles y enlace a “Nuevo turno”.
- E2E en navegador real con dispositivo en otra zona, varios estados y el selector obligatorio de barbero; evidencia a 320, 360, 768 y 1280 px y accesibilidad.

**Terminado cuando** la pantalla principal abre en hoy y presenta una lista cronológica, tenant-aware y accesible que coincide con la base de datos, sin adelantar navegación, detalle o cambios sobre la cita.

---

### HU-063 · Navegación de la agenda por fecha

| Campo | Valor |
| --- | --- |
| Función | `F-CITA-02`; amplía la lectura diaria de `F-CITA-01` |
| Reglas | `RN-CIT-01`, `RN-DIS-05`, `RN-DIS-07`, `RN-TEN-01` y estados confirmados de `estados-citas.md` |
| Decisiones | `DEC-007`, `DEC-016`, `DEC-019`, `DEC-020`, `DEC-024`, `DEC-033`–`DEC-038`, `DEC-047`, `DEC-074`, `DEC-075`, `DEC-078` |
| Actor | Barbero autenticado |
| Depende de | `HU-062` integrada mediante PR #113; contrato diario con parámetro `date` vigente |
| Bloquea | `HU-064` y recorridos posteriores que deben regresar a la fecha operativa de origen |
| Estado | Integrada en `main` mediante [PR #118](https://github.com/bcaceres19/barberia/pull/118) contra el issue real [#116](https://github.com/bcaceres19/barberia/issues/116), que permanece abierto por E2E/evidencia responsiva pendientes; prompt `PROMPT-HU-063-v1` `executed` |
| Riesgo | Sumar 24 horas, usar la zona del dispositivo o aceptar respuestas fuera de orden puede mostrar otro día, perder turnos nocturnos o reemplazar una consulta reciente con datos obsoletos. |

**Historia**

> Como barbero, quiero moverme al día anterior, al siguiente o a una fecha elegida sin perder el barbero seleccionado, para consultar y preparar jornadas distintas de hoy desde la misma agenda.

**Alcance incluido**

- Controles estables de día anterior, selector de fecha y día siguiente sobre la pantalla real de `HU-062`; la primera entrada continúa abriendo hoy.
- Cálculo por fechas civiles consecutivas en la zona IANA de la barbería, nunca mediante una duración fija de 24 horas ni mediante la zona del dispositivo.
- Reutilización de `GET /private/barbers/{barberId}/appointments/daily-agenda?date=AAAA-MM-DD`; no se crea una segunda consulta ni otra semántica diaria.
- Conservación del `barberId` explícito de `DEC-074` al cambiar de fecha y de la fecha elegida al cambiar de barbero.
- Fecha completa, zona y contexto visible durante carga, actualización, vacío, error y reintento; una respuesta antigua no puede sobrescribir la selección más reciente.
- Turnos nocturnos visibles en cada día intersecado conforme a `DEC-075`, con límites semiabiertos idénticos a `HU-062`.
- Estado navegable coherente con recarga, historial del navegador y enlaces posteriores del módulo, sin convertirlo en una agenda semanal o mensual.

**Alcance excluido**

- Cambiar el selector obligatorio por una vista consolidada de varios barberos o inferir `staff_user`→`barber`.
- Calendario semanal/mensual, arrastrar, redimensionar, desplazamiento infinito, tiempo real o visualización de huecos.
- Detalle, contacto, historial o cualquier operación de escritura sobre un turno.
- Modificar el contrato diario, la consulta PostgreSQL o sus índices salvo que una prueba demuestre un defecto real; esta historia consume la capacidad ya integrada.
- Usar la fecha del servidor o del dispositivo como autoridad después de que la barbería ya proporcionó su zona.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-063-01` | La primera entrada abre la fecha civil completa de hoy en la zona de la barbería y conserva el selector obligatorio de un único barbero de `DEC-074`. |
| `CA-063-02` | Anterior y siguiente avanzan exactamente un día civil incluso en jornadas locales de 23 o 25 horas; nunca saltan ni repiten una fecha por DST. |
| `CA-063-03` | Elegir una fecha válida consulta ese mismo `AAAA-MM-DD`; cancelar o introducir una fecha inválida conserva la selección y los datos confirmados anteriores sin enviar una petición ambigua. |
| `CA-063-04` | Cambiar de fecha conserva el `barberId` vigente y cambiar de barbero conserva la fecha elegida; ninguna combinación produce una vista consolidada ni consulta un tenant distinto. |
| `CA-063-05` | Un turno que cruza medianoche aparece en cada fecha cuyo intervalo civil interseca, exactamente como en `CA-062-04`, sin duplicarse dentro del mismo día. |
| `CA-063-06` | Si dos consultas se resuelven fuera de orden, solo la correspondiente a la selección vigente actualiza la pantalla; un error conserva fecha/barbero, contenido confirmado y un reintento accesible. |
| `CA-063-07` | Recargar, usar atrás/adelante o volver desde una ruta hija restaura una combinación coherente de fecha y barbero sin depender de estado global nuevo. |
| `CA-063-08` | Los tres controles conservan posición, nombre accesible, foco visible y objetivo de 44 px; la agenda reflowa a 320, 360, 768 y 1280 px, funciona con teclado y zoom 200 %, y axe-core no reporta violaciones conocidas. |

**Pruebas obligatorias**

- Unitarias del modelo de navegación para fin/inicio de mes y año, día local de 23/25 horas y zona del dispositivo distinta.
- Cliente/contrato para parámetro `date`, codificación, cancelación de solicitud y descarte de respuestas obsoletas.
- Componente para anterior, selector, siguiente, cambio de barbero, carga/actualización/vacío/error/reintento e historial del navegador.
- Regresión de PostgreSQL/HTTP de `HU-062` para extremos `[inicio, fin)`, nocturnidad de `DEC-075` y dos tenants; no se duplica una suite si el backend no cambia.
- E2E real: iniciar en hoy, navegar a anterior/siguiente/fecha elegida, recargar, volver y observar la fecha/barbero correctos con dispositivo en otra zona; evidencia responsive/accesible.

**Terminado cuando** la agenda existente navega por cualquier fecha civil con los tres controles aprobados, conserva el barbero y el contexto de navegación y nunca altera la semántica diaria, la zona o el aislamiento ya verificados en `HU-062`.

---

### HU-064 · Detalle e historial de un turno

| Campo | Valor |
| --- | --- |
| Función | Profundiza `F-CITA-01` y hace visible `F-EST-03` |
| Reglas | `RN-CIT-01`, `RN-HIS-01`, `RN-HIS-02`, `RN-RES-02`, `RN-RES-03`, `RN-DAT-01`, `RN-DAT-02`, `RN-DIS-07`, `RN-TEN-01` |
| Decisiones | `DEC-004`, `DEC-007`, `DEC-014`, `DEC-016`, `DEC-019`, `DEC-024`, `DEC-033`–`DEC-041`, `DEC-045`, `DEC-046`, `DEC-074`, `DEC-075` |
| Actor | Barbero autenticado |
| Depende de | `HU-063` integrada mediante PR #118; núcleo e historial append-only de `HU-060` vigentes |
| Bloquea | `HU-065` y las acciones posteriores que necesitan contexto, estado e historial visibles |
| Estado | Implementada contra el issue real [#120](https://github.com/bcaceres19/barberia/issues/120), que permanece abierto por E2E/evidencia responsiva pendientes: backend y frontend con pruebas reales; `CA-064-08` (E2E/evidencia responsiva) pendiente de ejecución contra un stack real, mismo estado que `HU-061`/`HU-062`/`HU-063`. Integrada en `main` mediante [PR #121](https://github.com/bcaceres19/barberia/pull/121); prompt `PROMPT-HU-064-v1` `executed` |
| Riesgo | Cargar el nombre actual del catálogo, exponer contacto en la lista o presentar un historial incompleto puede reescribir el pasado, filtrar datos personales o impedir auditar una disputa. |

**Historia**

> Como barbero, quiero abrir un turno y ver su hora, persona, contacto, servicio, estado e historial, para entender el compromiso y sus cambios antes de decidir cualquier acción.

**Alcance incluido**

- Lectura privada tenant-aware del detalle de una cita por identificador opaco, separada de la proyección mínima de agenda.
- Persona atendida y cliente que reservó, contacto opcional, nota del cliente cuando exista, barbero, origen, intervalo, zona, estado y snapshots de servicio, duración, precio y moneda.
- Servicio mostrado exclusivamente desde los snapshots de la cita; el catálogo vigente no sustituye ni corrige el dato histórico.
- Token opaco de versión en la representación, destinado a precondiciones de mutaciones posteriores sin convertir `updated_at` en una regla de negocio pública.
- Historial cronológico append-only con tipo de evento traducido, actor seguro, instante y cambios anterior/nuevo necesarios; motivo solo cuando pertenece al evento y está autorizado.
- Lectura acotada o paginada por cursor del historial con orden estable, sin `SELECT *`, N+1 ni datos técnicos innecesarios.
- Ruta de detalle desde cada fila de la agenda y regreso que conserva fecha y barbero de `HU-063`.
- Estados de carga, vacío de historial, error y reintento; los datos de contacto solo aparecen dentro de la superficie privada de detalle.

**Alcance excluido**

- Modificar, reprogramar, cancelar, completar, marcar inasistencia o corregir una cita.
- Añadir botones futuros deshabilitados, confirmaciones falsas o una acción contextual que todavía no exista.
- Editar o borrar entradas de historial, reconstruir eventos ausentes o sintetizar cambios que no fueron persistidos.
- Mostrar teléfono, correo, nota o historial en la lista diaria, navegación global, URL, logs, métricas o trazas.
- Usar el nombre/precio/duración actuales del servicio en lugar de los snapshots, o exponer `customerId`, IDs internos de actores y detalles de RLS.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-064-01` | Una sesión válida abre el detalle de una cita propia; un ID inexistente, mal formado o de otra barbería responde `404` uniforme sin revelar existencia ni ejecutar una lectura fuera del tenant. |
| `CA-064-02` | El detalle muestra persona atendida, cliente/contacto disponibles, nota opcional, barbero, origen, intervalo en la zona de la barbería, estado, snapshots exactos y un token opaco de versión; un cambio posterior de catálogo no altera la representación. |
| `CA-064-03` | Teléfono, correo y nota solo viajan en la operación de detalle autenticada y no aparecen en la respuesta diaria, logs, URL, errores ni telemetría. |
| `CA-064-04` | El historial se ordena por `occurredAt` y un desempate estable; cada evento muestra actor comprensible, instante, tipo y pares anterior/nuevo persistidos, sin inventar valores ausentes. |
| `CA-064-05` | Las entradas y cambios siguen siendo append-only: la entrega no concede `UPDATE`/`DELETE`, no modifica una migración aplicada y conserva las pruebas de `CA-060-06`. |
| `CA-064-06` | La paginación o cota del historial no omite ni duplica eventos entre páginas y la consulta usa el índice tenant-aware existente o uno nuevo justificado mediante `EXPLAIN (ANALYZE, BUFFERS)`. |
| `CA-064-07` | Abrir y volver conserva la fecha y el barbero de la agenda; carga, historial vacío, error y reintento mantienen contexto, anuncian el estado y no exhiben acciones no implementadas. |
| `CA-064-08` | La pantalla sigue la plantilla P0 de detalle, reflowa a 320, 360, 768 y 1280 px, soporta teclado/zoom 200 %, mantiene objetivos de 44 px y axe-core no reporta violaciones conocidas. |

**Pruebas obligatorias**

- Dominio/modelo de vista para snapshots, estados, actores, cambios, zona y orden estable.
- HTTP/contrato para detalle e historial: `200`, `400`, `401`, `404`, cursor/límite, forma exacta y ausencia de datos internos.
- PostgreSQL real con dos tenants, varios eventos empatados por instante, paginación sin duplicados, plan de consulta e intentos directos de modificar/borrar historial rechazados.
- Prueba de privacidad que inspecciona lista diaria, errores y logs para confirmar que no contienen contacto ni nota.
- Componente y router para abrir/volver, carga, historial vacío/lleno, error/reintento, foco y `vitest-axe`.
- E2E real desde una fecha distinta de hoy: abrir un turno, comprobar snapshots e historial, volver a la misma agenda y aportar evidencia responsive/accesible.

**Terminado cuando** el barbero puede inspeccionar un turno y su rastro completo desde la agenda sin cambiarlo, sin sustituir snapshots por catálogo vigente y sin ampliar la exposición de datos personales.

---

### HU-065 · Reprogramación auditada de un turno

| Campo | Valor |
| --- | --- |
| Función | `F-CITA-05`; transición `T2` de `estados-citas.md` |
| Reglas | `RN-CIT-01`, `RN-CIT-03`, `RN-CON-01`, `RN-CON-03`, `RN-DIS-05`, `RN-DIS-07`, `RN-HIS-01`, `RN-HIS-02`, `RN-TEN-01`, `RN-IDE-01` |
| Decisiones | `DEC-002`, `DEC-004`, `DEC-007`, `DEC-014`, `DEC-016`, `DEC-020`, `DEC-024`, `DEC-035`–`DEC-038`, `DEC-043`, `DEC-074`, `DEC-075`, `DEC-076`, `DEC-078` |
| Actor | Barbero autenticado |
| Depende de | `HU-064` integrada en `main` mediante [PR #121](https://github.com/bcaceres19/barberia/pull/121) (cumplido); `DP-CIT-06` resuelta como `DEC-076` (cumplido) |
| Bloquea | Modificación T3, cancelación y demás transiciones visibles de B3 |
| Estado | Implementada contra el issue real [#123](https://github.com/bcaceres19/barberia/issues/123), que permanece abierto por E2E/evidencia responsiva pendientes, integrada en `main` mediante [PR #125](https://github.com/bcaceres19/barberia/pull/125); prompt `PROMPT-HU-065-v1` `executed` |
| Riesgo | Una reprogramación no atómica, sin precondición de versión o sin la exclusión PostgreSQL puede perder un cambio concurrente, dejar el historial incompleto o guardar dos turnos cruzados. |

**Historia**

> Como barbero, quiero mover un turno confirmado a otra fecha y hora disponibles, para atender un acuerdo con el cliente sin perder el intervalo anterior ni crear un cruce en la agenda.

**Alcance incluido**

- Operación privada contract-first e idempotente para `T2`: cambia únicamente `starts_at`/`ends_at` de una cita `confirmed`; conserva barbero, servicio, duración, precio, persona, cliente, origen y estado.
- Nuevo inicio estrictamente futuro, expresado como fecha/hora civil en la zona de la barbería; el fin se deriva de `duration_minutes_snapshot`, nunca del cliente.
- Validación de jornada efectiva mediante los puertos de `schedule`; un bloqueo vigente en el nuevo intervalo rechaza la reprogramación con el mismo conflicto uniforme que un cruce de citas (`DEC-076`). Un bloqueo ya solapado nunca mueve la cita por sí solo.
- Exclusión GiST como última defensa ante cruce con otra cita que ocupa agenda; contigüidad exacta sigue permitida.
- Actualización de intervalo y evento `appointment_rescheduled` con valores anterior/nuevo en una sola transacción tenant-aware.
- Precondición explícita sobre la representación leída: si otra operación cambió el turno antes de confirmar, responde conflicto y obliga a recargar en vez de aplicar un último escritor silencioso.
- Mismo intervalo como no-op exitoso, sin nueva entrada de historial ni efectos secundarios; repetición de la misma clave/intención devuelve el resultado lógico original.
- Flujo desde el detalle de `HU-064`, con resumen antes de confirmar y retorno al detalle/agenda en la fecha nueva.

**Alcance excluido**

- Cambiar barbero, servicio, duración, precio, persona atendida, cliente, contacto, nota o estado; `T3` y edición general tendrán historias propias.
- Reprogramar una cita terminal, reabrir una cancelada o crear una cita nueva como efecto oculto.
- Reprogramación masiva, arrastrar en calendario, sugerencias de franjas o cálculo de disponibilidad pública.
- Enviar avisos o regenerar recordatorios antes de B5. La operación deja el hecho de negocio e historial correctos; la ausencia temporal de B5 se documenta y no se simula como envío exitoso.
- Modificar la migración aplicada de `HU-060`, relajar la exclusión o resolver concurrencia solo con una consulta previa.
- Permitir con advertencia o exigir un paso previo de retirar/exceptuar el bloqueo dentro del mismo flujo: `DEC-076` fija el rechazo directo, mismo tratamiento que un cruce de citas.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-065-01` | Con cita propia `confirmed`, precondición vigente y nuevo inicio válido, la operación actualiza una vez el intervalo y agrega un único `appointment_rescheduled` con `starts_at`/`ends_at` anteriores y nuevos dentro de la misma transacción. |
| `CA-065-02` | El servidor conserva todos los campos fuera de T2, deriva el fin desde la duración snapshot y usa la zona de la barbería; el body no puede cambiar tenant, barbero, servicio, duración, precio, estado ni actor. |
| `CA-065-03` | Un inicio que no es futuro, una cita terminal o una transición distinta de `confirmed`→`confirmed` se rechaza con el código documentado y sin modificar cita ni historial. |
| `CA-065-04` | Un intervalo fuera de la jornada efectiva se rechaza sin escritura; ante un bloqueo vigente, contrato, dominio, interfaz y pruebas coinciden literalmente con el `DEC-*` que resuelva `DP-CIT-06`. Una cita ya afectada por un bloqueo permanece intacta hasta una acción explícita. |
| `CA-065-05` | Un cruce total, parcial o de un minuto responde `409`; la contigüidad exacta es válida y una carrera real de dos reprogramaciones incompatibles no deja más de una agenda válida en PostgreSQL. |
| `CA-065-06` | El mismo intervalo es un no-op sin historial; la misma `Idempotency-Key` e intención repite el resultado original, contenido distinto entra en conflicto y una precondición obsoleta responde `409` con instrucción segura de recargar. |
| `CA-065-07` | Un ID ajeno o inexistente responde `404`; con dos tenants reales no se puede leer, reprogramar ni relacionar recursos cruzados y ningún log contiene datos personales. |
| `CA-065-08` | El formulario conserva datos ante error, evita doble toque, muestra intervalo/zona/consecuencia, pide confirmación explícita y gestiona éxito/conflicto; reflowa a 320, 360, 768 y 1280 px con teclado, foco, zoom 200 % y axe-core limpio. |

**Pruebas obligatorias**

- Dominio/aplicación para fecha futura, zona/DST, fin derivado, mismo intervalo, estado terminal, jornada, cada rama aprobada al resolver `DP-CIT-06` y precondición obsoleta.
- HTTP/contrato para éxito/repetición, `400`, `401`, `404`, `409`, `422`, campos desconocidos, `Idempotency-Key`, precondición y RFC 9457.
- PostgreSQL real con dos tenants: atomicidad cita+historial+cambios, contigüidad, cruces, rollback y carrera coordinada de dos conexiones sin `sleep`.
- Componente para resumen, confirmación, doble envío, conservación de datos, conflicto de agenda/versión, foco y `vitest-axe`.
- E2E real: abrir una cita futura, reprogramarla a otro día, verificar el historial y ambas agendas; repetir la solicitud y comprobar un solo evento, con evidencia responsive/accesible.

**Terminado cuando** una cita `confirmed` puede moverse una sola vez y de forma auditable a otro intervalo válido, con conflicto seguro ante cruces o datos obsoletos y sin adelantar T3, cancelación, estados ni notificaciones.

---

### HU-066 · Cancelación de un turno por el barbero

| Campo | Valor |
| --- | --- |
| Función | `F-CITA-06`; transición `T6` de `estados-citas.md` |
| Reglas | `RN-CIT-01`, `RN-CIT-03`, `RN-CAN-03`, `RN-CAN-04`, `RN-HIS-01`, `RN-HIS-02`, `RN-TEN-01`, `RN-IDE-01` |
| Decisiones | `DEC-011`, `DEC-012`, `DEC-014`, `DEC-016`, `DEC-017`, `DEC-024`, `DEC-035`–`DEC-038`, `DEC-041`, `DEC-043`, `DEC-077`–`DEC-080` |
| Actor | Barbero autenticado |
| Depende de | `HU-064` y `HU-065` integradas en `main` |
| Bloquea | Cancelación pública de B4 y efectos de notificación de B5 |
| Estado | Implementada contra el issue real [#225](https://github.com/bcaceres19/barberia/issues/225); prompt `PROMPT-HU-066-v1` `executed`; [PR #235](https://github.com/bcaceres19/barberia/pull/235) abierto contra `main`, pendiente de CI/merge |
| Riesgo | Una cancelación no atómica puede liberar la agenda sin dejar rastro, atribuirla al actor equivocado o aplicar un cambio sobre una versión ya obsoleta. |

**Historia**

> Como barbero, quiero cancelar un turno confirmado incluso después de su hora de inicio, para resolver una contingencia operativa, liberar la franja cuando todavía sea futura y conservar quién tomó la decisión.

**Alcance incluido**

- Comando privado y explícito para `T6`, con sesión, `Idempotency-Key` y precondición basada en el token opaco de versión del detalle.
- Transición exclusiva `confirmed` → `cancelled_by_barber`, sin límite temporal; el servidor deriva tenant, actor y estado destino.
- Cambio de estado y evento `appointment_cancelled_by_barber` dentro de una sola transacción tenant-aware.
- La cita cancelada sale de la restricción de exclusión; una franja futura queda disponible inmediatamente salvo que exista otro bloqueo o cita vigente.
- Repetición exacta por el mismo actor como éxito sin evento adicional; una cita ya terminada de otra forma produce conflicto y obliga a recargar.
- Acción desde el detalle, visible solo para `confirmed`, con consecuencia explícita, confirmación, espera, conflicto, error recuperable y foco administrado.

**Alcance excluido**

- Cancelación iniciada por el cliente, política de plazo o motivo obligatorio de `RN-CAN-01`/`RN-CAN-02` (B4).
- Cancelación masiva, selección de citas afectadas por un servicio o bloqueo y reapertura de una cita cancelada.
- Enviar avisos, invalidar recordatorios o afirmar entrega por correo/WhatsApp antes de B5.
- Cambiar intervalo, servicio, duración, precio, persona, contacto, nota o barbero.
- Añadir un motivo obligatorio para T6: ninguna fuente vigente lo exige.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-066-01` | Con una cita propia `confirmed`, precondición vigente y clave nueva, la operación cambia una vez a `cancelled_by_barber` e inserta un único evento `appointment_cancelled_by_barber` con actor barbero dentro de la misma transacción. |
| `CA-066-02` | La cancelación se permite antes, durante o después del intervalo mientras la cita siga `confirmed`; ningún reloj del cliente impone un límite y el servidor no acepta un estado o actor enviado por el body. |
| `CA-066-03` | Tras cancelar una cita futura, su intervalo deja de participar en la exclusión en la misma confirmación; una cita nueva contigua o coincidente solo se rechaza si otra cita o bloqueo vigente todavía ocupa la franja. |
| `CA-066-04` | Una cita `completed`, `no_show` o cancelada por otro actor responde `409` sin cambiar estado ni historial; una cancelación repetida por el barbero devuelve éxito sin duplicar el evento. |
| `CA-066-05` | La misma `Idempotency-Key` e intención reproduce el resultado lógico, contenido distinto entra en conflicto y una precondición obsoleta responde `409` distinguible con instrucción segura de recargar. |
| `CA-066-06` | Un ID inexistente, mal formado o de otra barbería no permite inferir existencia ni mutar datos: responde según el contrato uniforme y dos tenants reales permanecen aislados. |
| `CA-066-07` | La operación no cambia intervalo, snapshots, cliente, persona, contacto, nota ni barbero; tampoco crea intentos de notificación o recordatorios ficticios y ningún log contiene datos personales. |
| `CA-066-08` | El detalle ofrece la acción solo para `confirmed`, comunica la consecuencia antes de confirmar, evita doble toque, conserva contexto ante error y representa el resultado terminal en 320, 360, 768 y 1280 px, teclado y zoom 200 %, con axe-core limpio. |

**Pruebas obligatorias**

- Dominio/aplicación para cita activa antes/durante/después del intervalo, estados terminales, repetición y precondición obsoleta.
- HTTP/contrato para `200`, `400`, `401`, `404`, `409`, campos desconocidos, `Idempotency-Key`, precondición y RFC 9457.
- PostgreSQL real con dos tenants: atomicidad estado+historial, rollback, liberación de la exclusión y carrera coordinada entre cancelar y otra mutación.
- Componente y E2E desde el detalle: confirmación, doble envío, conflicto, error, regreso a la agenda y nueva ocupación válida de la franja futura.
- Evidencia responsive/accesible; el estado terminal reutiliza el tratamiento de `detalle-turno-eventos/10-turno-cancelado.png`, adaptando la atribución sin inventar un mockup de diálogo.

**Terminado cuando** un barbero puede cancelar una cita activa una sola vez, sin límite temporal, con liberación e historial atómicos y sin adelantar cancelación pública ni B5.

---

### HU-067 · Cierre manual como atendido o no asistió

| Campo | Valor |
| --- | --- |
| Función | `F-EST-01`, `F-EST-02`, `F-EST-04`; transiciones manuales `T4` y `T7` |
| Reglas | `RN-CIT-01`, `RN-CIT-03`, `RN-CIT-05`, `RN-HIS-01`, `RN-HIS-02`, `RN-TEN-01`, `RN-IDE-01` |
| Decisiones | `DEC-002`, `DEC-014`, `DEC-016`, `DEC-017`, `DEC-018`, `DEC-024`, `DEC-035`–`DEC-038`, `DEC-041`, `DEC-043`, `DEC-077`–`DEC-080` |
| Actor | Barbero autenticado |
| Depende de | `HU-066` integrada en `main` sin romper el núcleo de `HU-060` |
| Bloquea | `HU-068` y cierre automático posterior |
| Estado | Implementada contra el issue real [#226](https://github.com/bcaceres19/barberia/issues/226); prompt `PROMPT-HU-067-v1` `executed`; [PR #237](https://github.com/bcaceres19/barberia/pull/237) integrado en `main` |
| Riesgo | Cerrar antes de tiempo o duplicar el evento falsea las métricas del piloto; confundir `completed` con `no_show` exige después una corrección auditable. |

**Historia**

> Como barbero, quiero cerrar manualmente un turno cuya hora de inicio ya pasó como atendido o no asistió, para que la agenda refleje el resultado real y las métricas del piloto no confundan ambos casos.

**Alcance incluido**

- Dos comandos privados explícitos para `T4` manual y `T7`; no existe un `PATCH status` genérico.
- Solo una cita propia `confirmed` cuyo `starts_at` ya pasó según el reloj del servidor puede cerrarse.
- Transición a `completed` con evento `appointment_completed` o a `no_show` con evento `appointment_no_show`, atómica y tenant-aware.
- `completed` y `no_show` siguen ocupando agenda; el intervalo y los snapshots no se reescriben aunque el servicio terminara antes o después.
- `Idempotency-Key`, token opaco de versión y regla de repetición: el mismo resultado no duplica historial; un resultado distinto exige T8 y responde conflicto.
- Acciones contextuales desde el detalle con lenguaje “Marcar como atendido” y “Marcar que no asistió”, confirmación y estados accesibles.

**Alcance excluido**

- Cierre automático, configuración de modo/demora X, job del worker o aviso de citas vencidas.
- Aviso de inasistencia, invalidación de recordatorios y cualquier efecto de B5.
- Corrección T8, cancelaciones, T3, medición de duración real o desplazamiento de citas siguientes.
- Marcar `completed` o `no_show` desde una cita terminal o antes de `starts_at`.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-067-01` | Con cita propia `confirmed`, `starts_at` pasado, versión vigente y clave nueva, completar cambia una vez a `completed` e inserta un único `appointment_completed` con actor barbero en la misma transacción. |
| `CA-067-02` | Bajo las mismas precondiciones, marcar inasistencia cambia una vez a `no_show` e inserta un único `appointment_no_show`; ambas representaciones quedan distinguibles por texto además del color. |
| `CA-067-03` | Exactamente en `starts_at` la acción ya es válida según el reloj del servidor; antes de ese instante responde `422` sin modificar cita, historial ni idempotencia de forma que impida reintentar. |
| `CA-067-04` | Ambos resultados conservan intervalo, snapshots y demás datos y continúan ocupando agenda; no se crea otra cita cruzada sobre su intervalo ni se recalcula `ends_at` por duración real. |
| `CA-067-05` | Repetir el mismo resultado devuelve éxito sin evento adicional; intentar el resultado contrario u operar sobre cualquier otro terminal responde `409` y orienta a la futura corrección T8. |
| `CA-067-06` | Idempotencia concurrente, contenido divergente y token obsoleto siguen el protocolo existente; una carrera entre completar, marcar inasistencia y cancelar deja exactamente un estado terminal y un evento aplicable. |
| `CA-067-07` | Un recurso ajeno o inexistente no se distingue, dos tenants reales permanecen aislados y errores/logs no exponen contacto, nota, nombres ni valores internos. |
| `CA-067-08` | Las acciones aparecen solo cuando el turno `confirmed` ya puede cerrarse, anuncian consecuencia y espera, conservan foco/contexto ante error y muestran los estados terminales con evidencia en 320, 360, 768 y 1280 px, teclado, zoom 200 % y axe-core limpio. |

**Pruebas obligatorias**

- Tabla completa de dominio para ambos comandos: antes/exacto/después de `starts_at`, cada estado terminal, repetición y reloj cancelable.
- HTTP/contrato cerrado para ambas rutas y todos los errores, sin aceptar `status`, actor ni timestamps arbitrarios.
- PostgreSQL real con dos tenants, atomicidad, historial exacto, exclusión conservada y carrera de tres operaciones mediante barreras, nunca `sleep`.
- Componentes y E2E para las dos confirmaciones, espera, éxito, conflicto de versión/estado y estado terminal visible en detalle y agenda.
- Evidencia responsive/accesible; `completed` usa el evento `09-turno-completado.png` del atlas de detalle y `no_show` su variante documentada.

**Terminado cuando** el barbero registra exactamente uno de los dos resultados reales de un turno iniciado, sin liberar su intervalo, duplicar historial ni simular automatización o avisos.

---

### HU-068 · Corrección auditada de un resultado terminal

| Campo | Valor |
| --- | --- |
| Función | `F-EST-02`, `F-EST-03`, `F-EST-04`; transición `T8` |
| Reglas | `RN-CIT-01`, `RN-CIT-03`, `RN-CIT-04`, `RN-HIS-01`, `RN-HIS-02`, `RN-CON-01`, `RN-CON-03`, `RN-TEN-01`, `RN-IDE-01` |
| Decisiones | `DEC-011`, `DEC-012`, `DEC-014`, `DEC-016`, `DEC-017`, `DEC-024`, `DEC-035`–`DEC-038`, `DEC-041`, `DEC-043`, `DEC-077`–`DEC-080` |
| Actor | Barbero autenticado |
| Depende de | `HU-066` y `HU-067` integradas |
| Bloquea | Verificación completa de T8 y del criterio de salida de B3 |
| Estado | Implementada contra el issue real [#227](https://github.com/bcaceres19/barberia/issues/227); prompt `PROMPT-HU-068-v1` `executed`, integrada en `main` mediante [PR #239](https://github.com/bcaceres19/barberia/pull/239) |
| Riesgo | Editar el evento original o reocupar una franja sin verificar la exclusión destruye la auditoría o permite dos turnos superpuestos. |

**Historia**

> Como barbero, quiero corregir un resultado terminal marcado por error indicando el motivo, para que el estado vigente sea correcto sin borrar la decisión anterior ni reabrir silenciosamente el turno.

**Alcance incluido**

- Comando privado y explícito para T8 con estado terminal destino, motivo obligatorio, `Idempotency-Key` y token opaco de versión.
- Corrección solo entre `completed`, `cancelled_by_customer`, `cancelled_by_barber` y `no_show`; nunca hacia `confirmed`.
- Cambio de estado más evento `appointment_status_corrected`, con estado anterior/nuevo y motivo, dentro de una sola transacción tenant-aware.
- Los eventos anteriores permanecen intactos; la corrección agrega evidencia y el detalle muestra la secuencia completa.
- Al pasar de cancelado a `completed`/`no_show`, se vuelven a aplicar la validación temporal del resultado y la exclusión PostgreSQL; un intervalo ya ocupado produce `409` sin escritura.
- Al pasar de `completed`/`no_show` a cancelado, la cita deja de ocupar agenda en la misma transacción.

**Alcance excluido**

- Corregir hacia `confirmed`, reabrir o recrear automáticamente una cita; se crea una nueva tras validar disponibilidad.
- Editar o borrar historial, reemplazar el evento equivocado o ocultar una corrección posterior.
- Cambiar intervalo, snapshots, persona, cliente, contacto, nota o barbero.
- Notificar el cambio, regenerar recordatorios o ejecutar efectos de B5.
- Configuración/cierre automático y edición T3.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-068-01` | Con cita terminal propia, destino terminal distinto, motivo válido, versión vigente y clave nueva, la operación cambia una vez el estado y agrega un único `appointment_status_corrected` con actor, estado anterior/nuevo y motivo en la misma transacción. |
| `CA-068-02` | `confirmed`, un valor desconocido o un motivo vacío/blanco se rechazan según el contrato; una solicitud nueva al mismo estado responde como éxito sin efecto. Ninguno de esos casos modifica cita, historial ni registros previos. |
| `CA-068-03` | Ninguna corrección ejecuta `UPDATE`/`DELETE` sobre historial; los eventos originales y sus cambios conservan orden y contenido, y el nuevo evento queda visible después de ellos. |
| `CA-068-04` | Corregir desde un estado cancelado hacia `completed` o `no_show` exige que `starts_at` ya haya pasado y que el intervalo pueda volver a ocupar agenda; un cruce responde `409` y la transacción completa hace rollback. |
| `CA-068-05` | Corregir desde `completed`/`no_show` hacia un estado cancelado libera la exclusión en la misma confirmación; no crea disponibilidad pasada ni ignora bloqueos/citas que afecten una futura reutilización. |
| `CA-068-06` | Repetición exacta no duplica el evento; contenido distinto con la misma clave, versión obsoleta o carrera con otra corrección deja un único estado vigente y un historial coherente. |
| `CA-068-07` | Un ID ajeno o inexistente responde sin revelar existencia; dos tenants reales permanecen aislados y motivo, contacto y nombres no aparecen en logs ni errores técnicos. |
| `CA-068-08` | El detalle muestra la acción solo en estados terminales, exige motivo y destino explícitos, resume la consecuencia sobre agenda, administra foco/errores y presenta el historial corregido en 320, 360, 768 y 1280 px, teclado, zoom 200 % y axe-core limpio. |

**Pruebas obligatorias**

- Matriz de dominio con los 12 cambios posibles entre cuatro terminales, destinos inválidos, motivo, frontera temporal y repetición.
- HTTP/contrato para éxito y errores; request cerrado sin `confirmed`, actor, tenant ni campos de la cita.
- PostgreSQL real con dos tenants: append-only, atomicidad, liberación/reocupación de exclusión, rollback por cruce y carrera coordinada.
- Componente y E2E: corregir `completed` → `no_show`, ver ambos eventos, probar conflicto de versión y demostrar que una cancelada no revive a `confirmed`.
- Evidencia responsive/accesible en los anchos normativos; sin mockup exacto asignado, la composición es libre dentro de NAVA / Tailored Grid.

**Terminado cuando** un error de clasificación terminal puede corregirse con motivo y rastro inmutable, sin reapertura, pérdida de auditoría, fuga tenant o cruce de agenda.

---

## 7. Bloque B4 · Reserva pública y disponibilidad

> **Lote documental, no autorización de implementación.** `HU-090`–`HU-099` derivan exclusivamente del alcance P0 vigente. B4 sigue bloqueado hasta que B3 cumpla su criterio de salida. `DP-PUB-01` quedó resuelta como `DEC-082`; las decisiones abiertas `DP-PUB-02`–`DP-PUB-06` y `CT-011` deben resolverse antes de pasar el prompt afectado a `ready`.

Orden recomendado: `HU-090` → `HU-091` → `HU-092` → `HU-093` → `HU-094` → `HU-095` → `HU-096` → `HU-097` → `HU-098` → `HU-099`.

---

### HU-090 · Entrada pública de reservas de una barbería

| Campo | Valor |
| --- | --- |
| Función | `F-PUB-01` |
| Reglas | `RN-TEN-01`, `RN-DAT-01`, `RN-DAT-02` |
| Decisiones | `DEC-016`, `DEC-019`, `DEC-022`, `DEC-024`, `DEC-033`–`DEC-039`, `DEC-077`–`DEC-079`, `DEC-082` |
| Actor | Visitante sin cuenta |
| Depende de | Criterio de salida de B3 |
| Bloquea | `HU-091`–`HU-099` |
| Estado | Propuesta; `DP-PUB-01` resuelta (`DEC-082`); prompt `PROMPT-HU-090-v1` en `draft`, `issue: pending` hasta abrir el issue real |
| Riesgo | Un identificador público enumerable o una resolución tenant incorrecta expone barberías ajenas o abre el flujo con contexto falso. |

**Historia**

> Como cliente, quiero abrir el enlace público de una barbería sin registrarme, para iniciar una reserva con el negocio correcto.

**Alcance incluido**

- Ruta pública que resuelve una barbería habilitada desde un identificador no confiado y muestra nombre, zona horaria y contacto público mínimo.
- Estado de carga, enlace inválido/no disponible y recuperación sin revelar IDs internos ni la existencia de otro tenant.
- Cascarón público responsive y accesible, separado del panel autenticado.
- `public_slug` generado automáticamente desde el nombre de la barbería (slugify + sufijo numérico ante colisión), único globalmente, editable desde la configuración existente de `HU-020` con ruptura sin redirección del enlace anterior (`DEC-082`).

**Alcance excluido**

- Listar servicios, barberos o franjas; formular datos; crear o cancelar citas.
- Cuentas de cliente, búsqueda por teléfono/correo, portal o bot.
- Redirección de un slug anterior, historial de slugs o un tercer estado distinguible al deshabilitar (`DEC-082` ya lo descarta).

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-090-01` | Un enlace válido abre sin sesión el contexto público exacto de una barbería habilitada y presenta la hora en su zona. |
| `CA-090-02` | Un identificador mal formado, desconocido o no publicable produce una respuesta pública uniforme sin IDs internos ni pistas de otro tenant. |
| `CA-090-03` | Ningún parámetro del cliente fija `barbershopId`; API, aplicación y PostgreSQL resuelven y aíslan el tenant de forma coherente. |
| `CA-090-04` | La pantalla solo muestra datos públicos aprobados y logs/errores no contienen contacto privado, tokens ni datos personales. |
| `CA-090-05` | Carga, error y reintento conservan contexto; 320, 360, 768 y 1280 px, teclado, foco, zoom 200 % y axe-core quedan verificados. |

**Pruebas obligatorias:** dominio/HTTP para resolución uniforme; PostgreSQL real con dos tenants; componente/router y E2E de enlace válido/inválido; evidencia responsive y accesible.

**Terminado cuando** el visitante entra sin cuenta al contexto público correcto, sin poder seleccionar ni inferir otro tenant.

---

### HU-091 · Catálogo público de servicios disponibles

| Campo | Valor |
| --- | --- |
| Función | `F-PUB-02` |
| Reglas | `RN-SER-01`, `RN-SER-02`, `RN-SER-03`, `RN-SER-04`, `RN-TEN-01`, `RN-DAT-02` |
| Decisiones | `DEC-002`–`DEC-004`, `DEC-016`, `DEC-019`, `DEC-024`, `DEC-067`–`DEC-069`, `DEC-077`–`DEC-079` |
| Actor | Visitante en una barbería pública |
| Depende de | `HU-090`; `HU-022`–`HU-024` integradas |
| Bloquea | `HU-092`, `HU-094`–`HU-097` |
| Estado | Propuesta; prompt `PROMPT-HU-091-v1` en `draft`, `issue: pending` |
| Riesgo | Ofrecer servicios inactivos, sin barbero o con datos no vigentes conduce a reservas imposibles o engañosas. |

**Historia**

> Como cliente, quiero ver los servicios activos, su duración y precio vigentes, para elegir qué deseo reservar.

**Alcance incluido**

- Lectura pública tenant-aware de servicios activos con nombre, descripción, duración, precio y moneda COP.
- Solo aparecen servicios con al menos una asignación vigente; un servicio inactivo desaparece sin afectar snapshots de citas existentes.
- Selección accesible de un servicio y estados carga, vacío, error y reintento.

**Alcance excluido**

- Administrar catálogo o asignaciones, calcular horas, crear citas o exponer IDs de tenant.
- Mostrar servicios inactivos, precios históricos o datos internos de auditoría.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-091-01` | La consulta pública devuelve únicamente servicios activos y asignados de la barbería resuelta, con duración positiva, precio vigente y COP. |
| `CA-091-02` | Desactivar un servicio lo retira de nuevas lecturas; reactivarlo con asignación válida lo recupera sin alterar citas existentes. |
| `CA-091-03` | Un servicio o asignación de otra barbería nunca aparece ni puede seleccionarse, verificado con dos tenants reales. |
| `CA-091-04` | Vacío, carga, error y reintento son distinguibles y la selección no depende solo de color. |
| `CA-091-05` | La lista y nombres largos reflowan en los cuatro anchos normativos, teclado/zoom 200 % y axe-core limpio. |

**Pruebas obligatorias:** consulta y filtros en PostgreSQL real; contrato público; componente de selección; E2E con servicio activo/inactivo y dos tenants.

**Terminado cuando** el cliente puede elegir exclusivamente un servicio realmente ofrecido por la barbería.

---

### HU-092 · Selección pública de barbero

| Campo | Valor |
| --- | --- |
| Función | `F-PUB-04` (selección) |
| Reglas | `RN-TEN-01`, `RN-CON-01`, `RN-DAT-02` |
| Decisiones | `DEC-019`, `DEC-024`, `DEC-047`, `DEC-068`, `DEC-077`–`DEC-079` |
| Actor | Cliente que ya eligió un servicio |
| Depende de | `HU-091`; `HU-021`/`HU-023` integradas |
| Bloquea | `HU-094`–`HU-097` |
| Estado | Propuesta; prompt `PROMPT-HU-092-v1` en `draft`, `issue: pending` |
| Riesgo | Permitir un barbero no asignado al servicio produce disponibilidad y reservas inválidas. |

**Historia**

> Como cliente, quiero elegir quién me atiende cuando hay varias opciones y omitir ese paso cuando solo hay una, para avanzar con la menor fricción posible.

**Alcance incluido**

- Lista pública de barberos pertenecientes a la barbería y asignados al servicio activo elegido.
- Preselección automática sin paso adicional cuando existe exactamente uno; elección explícita cuando existen varios.
- Revalidación de servicio y asignación en cada lectura posterior; cambios concurrentes no conservan una selección inválida.

**Alcance excluido**

- Preferencias, ranking, “cualquiera”, asignación automática, perfiles, fotos o baja de barberos.
- Consultar disponibilidad o crear la cita.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-092-01` | Con un único barbero asignado, queda seleccionado y el flujo no agrega un paso; con varios, exige una elección explícita. |
| `CA-092-02` | Solo se listan barberos de la barbería pública con asignación vigente al servicio activo elegido. |
| `CA-092-03` | Un ID ajeno, inexistente o ya no asignado se rechaza de forma uniforme antes de consultar horas o crear cita. |
| `CA-092-04` | Cambiar el servicio limpia una selección incompatible y conserva una compatible únicamente tras revalidarla. |
| `CA-092-05` | Estados y selector funcionan con teclado, foco visible, lector, zoom 200 % y los cuatro anchos normativos. |

**Pruebas obligatorias:** matriz 0/1/N barberos; PostgreSQL real con asignaciones cruzadas; contrato, componente y E2E de preselección/elección.

**Terminado cuando** el flujo conserva exactamente un barbero elegible y nunca inventa asignación automática.

---

### HU-093 · Configuración de reserva y cancelación pública

| Campo | Valor |
| --- | --- |
| Función | `F-DISP-04`, `F-DISP-05`, `F-CITA-07`; configuración de `RN-DIS-06` |
| Reglas | `RN-DIS-04`, `RN-DIS-06`, `RN-CAN-01`, `RN-CAN-02`, `RN-TEN-01` |
| Decisiones | `DEC-005`, `DEC-006`, `DEC-010`, `DEC-018`, `DEC-024`, `DEC-077`–`DEC-079` |
| Actor | Barbero autenticado |
| Depende de | B3; `HU-020`/`HU-012`; `DP-PUB-02` resuelta |
| Bloquea | `HU-094`, `HU-095`, `HU-099` |
| Estado | Propuesta; prompt `PROMPT-HU-093-v1` en `draft`, `issue: pending` |
| Riesgo | Rangos inventados o una política ambigua pueden ocultar horarios válidos o permitir cancelaciones contrarias a la decisión del negocio. |

**Historia**

> Como barbero, quiero configurar anticipación, ventana, rejilla y política de cancelación pública, para ajustar la reserva autónoma a mi operación.

**Alcance incluido**

- Configuración tenant-aware de anticipación mínima, ventana máxima, paso de rejilla, plazo de cancelación, actor permitido fuera de plazo y motivo obligatorio.
- Valores iniciales exactos de `DEC-018`; rangos y combinación predeterminada de política solo después de `DP-PUB-02`.
- Lectura y actualización contract-first con precondición de versión, validación cerrada y UI accesible.

**Alcance excluido**

- Cambiar citas existentes, aplicar límites públicos a citas manuales o configurar notificaciones.
- Inventar máximos, mínimos o defaults no confirmados.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-093-01` | Una barbería nueva recibe 60 minutos, 3 días, rejilla de 15 y plazo de cancelación de 20 minutos; los demás defaults coinciden con la decisión de `DP-PUB-02`. |
| `CA-093-02` | Solo valores dentro de rangos aprobados se guardan; campos desconocidos, combinaciones incoherentes y versión obsoleta se rechazan sin cambios parciales. |
| `CA-093-03` | La configuración de A nunca se lee ni modifica desde B y el body no acepta `barbershopId`. |
| `CA-093-04` | Cambiar la política afecta evaluaciones futuras, no reescribe citas, historial ni cancelaciones ya realizadas. |
| `CA-093-05` | El formulario explica unidades y consecuencias, conserva datos ante error y cumple responsive, teclado, zoom 200 % y axe-core. |

**Pruebas obligatorias:** dominio de rangos/fronteras; PostgreSQL real con dos tenants; contrato y conflicto de versión; componente y E2E de persistencia.

**Terminado cuando** la barbería controla todas las variables públicas confirmadas sin alterar la creación manual ni decisiones históricas.

---

### HU-094 · Motor de disponibilidad pública real

| Campo | Valor |
| --- | --- |
| Función | `F-DISP-01` |
| Reglas | `RN-DIS-01`–`RN-DIS-07`, `RN-BLQ-01`–`RN-BLQ-04`, `RN-CON-01`, `RN-CON-03`, `RN-CAN-04`, `RN-TEN-01` |
| Decisiones | `DEC-002`, `DEC-005`–`DEC-009`, `DEC-012`, `DEC-018`–`DEC-020`, `DEC-024`, `DEC-070`, `DEC-073`, `DEC-076` |
| Actor | Cliente con servicio y barbero válidos |
| Depende de | `HU-091`–`HU-093`; B2/B3; `DP-PUB-03` resuelta |
| Bloquea | `HU-095`, `HU-097`, `HU-099` |
| Estado | Propuesta; prompt `PROMPT-HU-094-v1` en `draft`, `issue: pending` |
| Riesgo | Restar mal un solo factor ofrece un turno imposible o esconde capacidad vendible. |

**Historia**

> Como cliente, quiero recibir únicamente horas donde el servicio completo cabe de verdad, para no confirmar un turno incompatible con la agenda.

**Alcance incluido**

- Proyección pura por fecha civil, zona de barbería, barbero y servicio vigentes.
- Intersección de jornada recurrente, excepciones/festivos, bloqueos vigentes y citas que ocupan agenda; unión de solapes antes de restar.
- Aplicación de duración snapshot candidata, `[inicio, fin)`, anticipación, ventana y rejilla aprobada.
- Resultado determinista, ordenado y sin reservar temporalmente ninguna franja.

**Alcance excluido**

- UI de calendario, creación de cita, hold de franja, caché distribuida o predicción de demanda.
- Elegir cómo reinicia la rejilla tras una interrupción antes de resolver `DP-PUB-03`.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-094-01` | Solo aparece un inicio si el intervalo completo cabe en jornada efectiva y no interseca cita ocupante ni bloqueo vigente. |
| `CA-094-02` | Hueco exacto y contigüidad son válidos; un minuto insuficiente, cierre excedido o duración no múltiplo de rejilla se resuelve conforme a reglas. |
| `CA-094-03` | Festivo, excepción, recurrencia, vacaciones, emergencia, cruce de medianoche y restricciones solapadas producen una resta única y correcta. |
| `CA-094-04` | Anticipación y ventana se evalúan con reloj del servidor; consultar no crea filas, bloqueos, sesiones ni efectos de negocio. |
| `CA-094-05` | Dos tenants y dos barberos simultáneos permanecen aislados; cancelación futura libera la franja si ninguna otra restricción la cubre. |
| `CA-094-06` | La consulta tiene límite de rango/costo aprobado, plan de consulta justificado y pruebas de rendimiento reproducibles sin datos reales. |

**Pruebas obligatorias:** tabla exhaustiva de dominio; PostgreSQL real con dos tenants y once factores; límites temporales/zona; contrato sin efectos y plan de consulta.

**Terminado cuando** el backend produce la misma disponibilidad que las reglas, incluso en fronteras y combinaciones adversas.

---

### HU-095 · Exploración pública de fechas y horarios

| Campo | Valor |
| --- | --- |
| Función | `F-PUB-03` |
| Reglas | `RN-DIS-01`–`RN-DIS-07`, `RN-CON-04`, `RN-DAT-02` |
| Decisiones | `DEC-005`–`DEC-007`, `DEC-018`–`DEC-020`, `DEC-077`–`DEC-079` |
| Actor | Cliente con servicio y barbero elegidos |
| Depende de | `HU-094` |
| Bloquea | `HU-096`, `HU-097` |
| Estado | Propuesta; prompt `PROMPT-HU-095-v1` en `draft`, `issue: pending` |
| Riesgo | Una pantalla que conserva horas obsoletas o usa la zona del dispositivo induce al cliente a confirmar otra hora. |

**Historia**

> Como cliente, quiero explorar fechas y horas disponibles en la zona de la barbería, para escoger un turno válido con claridad.

**Alcance incluido**

- Navegación limitada por anticipación/ventana y consulta bajo demanda del motor `HU-094`.
- Selección de una fecha y una franja con intervalo/duración visibles; zona de barbería explícita.
- Carga, día sin horas, error, reintento y refresco cuando cambian servicio o barbero.

**Alcance excluido**

- Reservar temporalmente, mostrar “otras personas mirando”, confirmar cita o sugerir alternativas tras conflicto.
- Calendario infinito, zona del dispositivo o horas pasadas.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-095-01` | Solo se pueden explorar fechas dentro de la ventana pública y solo se seleccionan horas devueltas por `HU-094`. |
| `CA-095-02` | La hora, fecha, duración y zona de barbería se anuncian juntas; cambiar zona del dispositivo no cambia el turno mostrado. |
| `CA-095-03` | Cambiar servicio o barbero invalida y vuelve a consultar la selección; una respuesta tardía no sobrescribe el contexto nuevo. |
| `CA-095-04` | Día vacío, carga, error y reintento conservan selecciones previas válidas y no afirman que una consulta reserve la hora. |
| `CA-095-05` | Teclado, foco, lector, objetivos de 44 px, zoom 200 %, axe-core y 320/360/768/1280 px quedan verificados. |

**Pruebas obligatorias:** modelo de estado y carreras de respuestas; componente; contrato; E2E desde otro huso horario y día sin disponibilidad.

**Terminado cuando** el cliente elige una franja real y comprende inequívocamente cuándo ocurrirá, sin adquirir un hold inexistente.

---

### HU-096 · Datos del cliente y persona atendida

| Campo | Valor |
| --- | --- |
| Función | `F-PUB-05`, `F-PUB-06` |
| Reglas | `RN-RES-01`–`RN-RES-03`, `RN-DAT-01`, `RN-DAT-02`, `RN-TEN-01` |
| Decisiones | `DEC-016`, `DEC-022`, `DEC-045`, `DEC-046`, `DEC-077`–`DEC-079` |
| Actor | Cliente con una franja elegida |
| Depende de | `HU-095`; `DP-PUB-04` resuelta |
| Bloquea | `HU-097` |
| Estado | Propuesta; prompt `PROMPT-HU-096-v1` en `draft`, `issue: pending` |
| Riesgo | Reconciliar contactos de forma ambigua puede unir personas distintas; pedir datos extra incumple minimización. |

**Historia**

> Como cliente, quiero indicar mis datos y si el turno es para otra persona, para que la barbería sepa a quién atender y a quién contactar.

**Alcance incluido**

- Nombre, teléfono y correo obligatorios; nota opcional; selector “para mí/otra persona” y segundo nombre solo cuando difiere.
- Normalización y validación server-side; `attendeeName` siempre resuelto y notificaciones futuras dirigidas al cliente.
- Política de reconciliación pública por teléfono/correo únicamente después de `DP-PUB-04`.

**Alcance excluido**

- Documento, dirección, fecha de nacimiento, datos sensibles o contacto de la persona atendida.
- Cuenta de cliente, libreta de contactos, consentimiento jurídico no aprobado o envío de mensajes.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-096-01` | Para sí mismo, el nombre atendido se deriva del cliente sin pedirlo dos veces; para otra persona, un nombre no vacío adicional es obligatorio. |
| `CA-096-02` | Nombre, teléfono y correo son obligatorios, nota es opcional y ningún otro dato personal aparece en UI o request. |
| `CA-096-03` | Validaciones y normalización coinciden en contrato/backend; errores por campo conservan todo dato no sensible escrito. |
| `CA-096-04` | Coincidencias y conflictos de teléfono/correo siguen exactamente `DP-PUB-04`, siempre dentro de la barbería. |
| `CA-096-05` | Datos personales no aparecen en URL, logs, telemetría ni errores; pruebas usan valores ficticios. |
| `CA-096-06` | Formulario condicional, mensajes y foco funcionan con teclado, autofill, zoom 200 %, axe-core y cuatro anchos normativos. |

**Pruebas obligatorias:** dominio de identidad; contrato cerrado; PostgreSQL real con coincidencias/conflictos y dos tenants; componente y E2E para sí/tercero.

**Terminado cuando** el sistema distingue cliente y persona atendida con los datos mínimos y sin fusionar identidades ambiguas.

---

### HU-097 · Confirmación pública concurrente e idempotente

| Campo | Valor |
| --- | --- |
| Función | `F-PUB-04` (creación), `F-DISP-03` |
| Reglas | `RN-RES-01`–`RN-RES-03`, `RN-DIS-03`–`RN-DIS-07`, `RN-CON-01`–`RN-CON-06`, `RN-CNF-01`, `RN-HIS-01`, `RN-IDE-01`, `RN-TEN-01` |
| Decisiones | `DEC-005`–`DEC-008`, `DEC-012`, `DEC-013`, `DEC-016`, `DEC-019`, `DEC-020`, `DEC-022`, `DEC-024`, `DEC-041`–`DEC-046`, `DEC-073` |
| Actor | Cliente que confirma el resumen público |
| Depende de | `HU-090`–`HU-096`; `DP-PUB-05` y `DP-PUB-06` resueltas; `CT-011` resuelta |
| Bloquea | `HU-098`, `HU-099`, B5 |
| Estado | Propuesta; prompt `PROMPT-HU-097-v1` en `draft`, `issue: pending` |
| Riesgo | Una carrera mal resuelta crea dos citas o pierde datos; una confirmación parcial deja cita sin historial o acceso del cliente. |

**Historia**

> Como cliente, quiero confirmar una sola vez el turno elegido y recibir alternativas si alguien se adelantó, para terminar sin duplicados ni perder mis datos.

**Alcance incluido**

- Comando público contract-first con `Idempotency-Key`; revalida barbería, servicio, asignación, jornada, bloqueo, límites y franja al confirmar.
- Cliente, cita `confirmed`, snapshots, `appointment_created` y credencial de acceso se persisten con atomicidad definida por `DP-PUB-05`/`CT-011`.
- Exclusión PostgreSQL como última defensa; exactamente un ganador bajo N solicitudes concurrentes.
- Perdedor recibe conflicto comprensible, conserva formulario y obtiene alternativas conforme a `DP-PUB-06`.
- Carrera reserva/bloqueo conserva la cita confirmada y registra/expone la afectación al barbero según `RN-CON-06`.

**Alcance excluido**

- Hold, pago, aprobación manual, lista de espera, CAPTCHA no decidido o notificaciones de B5.
- Confiar en disponibilidad consultada, timestamps, precio, duración, estado, actor o tenant enviados por el cliente.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-097-01` | Una intención válida crea exactamente una cita `confirmed` de origen `public`, con snapshots e historial inicial en la unidad atómica aprobada. |
| `CA-097-02` | El servidor revalida todo; recursos inactivos/ajenos, ventana vencida, bloqueo o jornada inválida no producen cita ni historial parcial. |
| `CA-097-03` | N confirmaciones simultáneas de una franja dejan una ganadora; las demás reciben `409` controlado y alternativas aprobadas sin perder formulario. |
| `CA-097-04` | Misma clave/contenido reproduce el resultado; misma clave/contenido distinto o clave en curso sigue `DEC-043`; doble toque no duplica efectos. |
| `CA-097-05` | Una carrera con bloqueo nunca borra una cita confirmada y deja al barbero la afectación exigida por `RN-CON-06`. |
| `CA-097-06` | Con dos tenants no hay relaciones cruzadas; logs, errores e idempotencia no conservan teléfono, correo, nota ni token en claro. |
| `CA-097-07` | El resumen exige confirmación explícita, bloquea doble toque y maneja éxito/conflicto/red con foco y datos conservados en los anchos normativos y axe-core limpio. |

**Pruebas obligatorias:** dominio/HTTP; PostgreSQL real con dos tenants, rollback y barreras de concurrencia sin `sleep`; idempotencia; componente; E2E ganador/perdedor y carrera con bloqueo.

**Terminado cuando** cada intención pública produce cero o una cita completa y el conflicto ofrece una recuperación segura.

---

### HU-098 · Confirmación y acceso privado del cliente a su turno

| Campo | Valor |
| --- | --- |
| Función | `F-PUB-07` |
| Reglas | `RN-CNF-01`, `RN-CNF-02`, `RN-DAT-01`–`RN-DAT-03`, `RN-TEN-01` |
| Decisiones | `DEC-016`, `DEC-022`, `DEC-024`, `DEC-025`, `DEC-049`, `DEC-077`–`DEC-079` |
| Actor | Cliente con enlace aleatorio del turno |
| Depende de | `HU-097`; `DP-PUB-05` y `CT-011` resueltas |
| Bloquea | `HU-099`, B5 |
| Estado | Propuesta; prompt `PROMPT-HU-098-v1` en `draft`, `issue: pending` |
| Riesgo | Filtrar o registrar el token permite consultar/cancelar una cita ajena; una expiración inventada puede dejar al cliente sin acceso. |

**Historia**

> Como cliente, quiero ver la confirmación y consultar mi turno mediante un enlace seguro, para conservar fecha, hora y política sin crear una cuenta.

**Alcance incluido**

- Pantalla posterior a la creación y ruta `/customer` autenticada solo por credencial aleatoria larga.
- Lectura mínima de la cita, barbería, persona atendida, servicio, barbero, intervalo/zona, estado y política vigente de cancelación.
- Token almacenado solo como hash, respuesta uniforme para inválido/expirado/revocado y redacción de URL en logs.
- Emisión, vigencia, rotación, revocación y entrega por correo según `DP-PUB-05`/`CT-011`; no se inventan.

**Alcance excluido**

- Portal, login, búsqueda por contacto, edición/reprogramación por cliente o historial interno.
- Mostrar IDs, notas internas, otros turnos o datos de otros clientes.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-098-01` | Tras confirmar, el cliente ve una representación coherente de su única cita y el acceso posterior requiere la credencial aprobada. |
| `CA-098-02` | Token inválido, vencido, revocado o de cita anonimizada produce respuesta uniforme; el valor en claro no se persiste ni registra. |
| `CA-098-03` | La lectura expone solo datos necesarios y nunca historial técnico, contacto completo innecesario, IDs internos o recursos relacionados. |
| `CA-098-04` | Hora y política se muestran en zona de barbería y el estado vigente nunca implica cancelación por falta de respuesta. |
| `CA-098-05` | Dos tenants y tokens distintos no se cruzan; una credencial no autoriza consultar otra cita. |
| `CA-098-06` | Carga, error/reintento, estado terminal y enlace inválido cumplen teclado, foco, zoom 200 %, axe-core y cuatro anchos. |

**Pruebas obligatorias:** token/hash y amenazas; HTTP customer; PostgreSQL real con dos tenants; redacción de logs; componente y E2E creación→consulta.

**Terminado cuando** el cliente recupera solo su turno mediante una credencial segura y revocable, sin cuenta ni exposición lateral.

---

### HU-099 · Cancelación pública conforme a la política

| Campo | Valor |
| --- | --- |
| Función | `F-PUB-08`, consumo de `F-CITA-07` |
| Reglas | `RN-CAN-01`–`RN-CAN-04`, `RN-CIT-03`, `RN-HIS-01`, `RN-HIS-02`, `RN-DIS-04`, `RN-IDE-01`, `RN-TEN-01` |
| Decisiones | `DEC-010`–`DEC-012`, `DEC-014`, `DEC-016`–`DEC-018`, `DEC-022`, `DEC-024`, `DEC-041`, `DEC-043` |
| Actor | Cliente con acceso válido a su turno |
| Depende de | `HU-093`, `HU-094`, `HU-098`; `DP-PUB-02`/`DP-PUB-05` resueltas |
| Bloquea | Criterio de salida de B4 y efectos de B5 |
| Estado | Propuesta; prompt `PROMPT-HU-099-v1` en `draft`, `issue: pending` |
| Riesgo | Evaluar el plazo en el navegador o aplicar mal la política cancela compromisos sin autorización o deja una franja bloqueada. |

**Historia**

> Como cliente, quiero cancelar mi turno desde su enlace cuando la política lo permita, para avisar a la barbería y liberar la hora.

**Alcance incluido**

- Comando customer explícito T5, con token válido, `Idempotency-Key`, reloj del servidor y precondición de versión.
- Dentro del plazo se permite; exactamente en el límite y fuera de plazo se aplica la política configurada, incluido motivo cuando corresponda.
- Estado `cancelled_by_customer`, motivo aprobado y evento del cliente en una transacción; libera la exclusión inmediatamente.
- Repetición exacta es no-op; otro terminal, versión obsoleta o intención divergente produce conflicto seguro.
- UI explica política, contacto con barbería cuando se rechaza y consecuencia antes de confirmar.

**Alcance excluido**

- Cancelación por barbero, masiva, reprogramación, reapertura, penalidades o pagos.
- Aviso al barbero/confirmación por canales e invalidación de recordatorios hasta B5; no se simulan envíos.

**Criterios de aceptación**

| Código | Criterio |
| --- | --- |
| `CA-099-01` | Una cita propia `confirmed` cancelable cambia una vez a `cancelled_by_customer` y agrega un único evento con actor cliente en la misma transacción. |
| `CA-099-02` | El reloj del servidor evalúa antes, exacto y después del límite; fuera de plazo se permite o rechaza exactamente según configuración. |
| `CA-099-03` | Cuando la política exige motivo, vacío/blanco se rechaza; cuando no lo exige no se inventa ni solicita como obligatorio. |
| `CA-099-04` | Al confirmar la cancelación, la franja futura vuelve a `HU-094` salvo otra cita/bloqueo; el pasado nunca se ofrece. |
| `CA-099-05` | Repetición exacta no duplica historial; otro terminal, versión obsoleta, clave divergente o token inválido no modifica nada ni revela existencia. |
| `CA-099-06` | UI y contrato explican rechazo y contacto seguro; confirmación, espera, error y estado cancelado cumplen responsive, teclado, foco, zoom y axe-core. |

**Pruebas obligatorias:** tabla de política y fronteras temporales; HTTP/idempotencia; PostgreSQL real con dos tenants, atomicidad/liberación y carrera; componente y E2E cancelar→reconsultar franja.

**Terminado cuando** el cliente cancela solo cuando está autorizado, deja rastro correcto y libera de inmediato una franja futura.

---

## 8. Historias pendientes de redacción

| Bloque | Rango reservado | Se redacta cuando |
| --- | --- | --- |
| B1 | `HU-025` – | `HU-020`–`HU-024` implementadas (`DEC-067`–`DEC-069` propagadas); redactar lo restante solo después de revisar el criterio de salida de B1 |
| B2 | `HU-040` – `HU-042` | Integradas en `main` (PR `#93`, `#96`, `#99`); seguimientos parciales en issues `#90`, `#95`, `#98` y `#100` |
| B3 | `HU-069` – | `HU-066`–`HU-068` ya integradas en `main`; continuar con T3 y cierre automático solo después de revisar este lote y resolver cualquier duda de semántica de snapshots/configuración |
| B4 | `HU-100` – | `HU-090`–`HU-099` ya redactadas; `DP-PUB-01` resuelta (`DEC-082`), `HU-090` en curso (issue [#243](https://github.com/bcaceres19/barberia/issues/243)); continuar con el resto solo después de resolver `DP-PUB-02`–`DP-PUB-06`/`CT-011` |
| B5 | `HU-130` – | B4 cumple su criterio de salida |
| B6 | `HU-150` – | B5 cumple su criterio de salida |

Redactar por anticipado las historias de un bloque lejano produce texto que hay que reescribir: lo que se aprende construyendo el bloque anterior cambia la historia siguiente.

---

## 9. Dudas que bloqueaban historias de B0

| Duda | Historia antes bloqueada | Qué faltaba decidir | Resuelta como |
| --- | --- | --- | --- |
| `DP-SEG-04` | `HU-005`, `HU-006` | Mecanismo concreto de la sesión larga y su duración exacta | `DEC-050` |
| `DP-SEG-05` | `HU-008`, `HU-011` | Canal y proveedor del código de recuperación | `DEC-051` |
| `DP-SEG-06` | `HU-007` | Duración de la ventana del límite por IP y del escalamiento | `DEC-052` |

Las tres se resolvieron el 2026-08-11 (ver [dudas-pendientes.md](../00-control/dudas-pendientes.md) y [registro-decisiones.md](../00-control/registro-decisiones.md)). `HU-005`–`HU-008` y `HU-011` ya no dependen de una duda abierta; siguen sujetas al criterio de salida de B0 antes de implementarse.
