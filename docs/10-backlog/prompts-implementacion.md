---
titulo: "Prompts de implementación de B0 y primeras historias de B1"
version: "1.2"
estado: "Propuesta"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-12"
documentos_relacionados:
  - "../02-requisitos/historias-usuario.md"
  - "plan-bloques.md"
  - "../../AGENTS.md"
  - "../../CONTRIBUTING.md"
  - "prompts/README.md"
---

# Prompts de implementación de B0 y primeras historias de B1

> **Estado del contenido: Propuesta.** Un prompt es una herramienta de trabajo, **no** una fuente normativa. Si un prompt contradice `AGENTS.md`, una regla `RN-*`, una decisión `DEC-*` o un estándar de `docs/03-desarrollo/`, manda el documento normativo y el prompt se corrige.

> **Catálogo vigente:** este documento conserva el inventario histórico y prompts aún no migrados. Todo prompt nuevo o revisado se guarda individualmente en el [catálogo de prompts persistentes](prompts/README.md). Si existe allí una versión individual de la misma preocupación, el catálogo identifica cuál es la operativa y su estado real.

---

## 1. Cómo usar estos prompts

1. Verifique que la historia no esté bloqueada por una duda abierta o por el criterio de salida del bloque anterior ([historias-usuario.md](../02-requisitos/historias-usuario.md), sección 6, y [plan-bloques.md](plan-bloques.md)). Si lo está, **no empiece**.
2. Cree el issue con los criterios `CA-*` de la historia y la rama corta correspondiente.
3. Pegue el **preámbulo obligatorio** seguido del prompt de la historia.
4. Revise el resultado contra los criterios de aceptación, no contra la impresión general.
5. Actualice la matriz de trazabilidad y el historial de cambios en el mismo pull request.

Un prompt no sustituye la revisión. El agente que implementa no es quien decide si la historia está terminada.

### Prompts expandidos

Los prompts de este documento son el **índice** de las doce historias de B0 y de las dos
primeras historias propuestas de B1. Cuando una
historia entra en construcción, su versión detallada —decisiones técnicas concretas,
trampas conocidas y forma exacta de verificar cada criterio— vive en
[prompts-detallados-b0.md](prompts-detallados-b0.md). Si ambos difieren, manda el detallado
y este se corrige.

| Documento | Contenido |
| --- | --- |
| [prompts-detallados-b0.md](prompts-detallados-b0.md) | `HU-002` (backend) y `HU-009` (frontend), en construcción |
| [prompt-llaves-asimetricas.md](prompt-llaves-asimetricas.md) | Sesión con criptografía asimétrica. **Requiere `DEC-*` que resuelva `DP-SEG-04` antes de codificar** |

---

## 2. Preámbulo obligatorio

Va **antes** de cualquier prompt de historia, sin modificaciones.

```text
Trabajas en el repositorio del sistema de agenda para barberías.

Antes de escribir una sola línea, lee y respeta, en este orden de autoridad:
1. AGENTS.md
2. docs/00-control/registro-decisiones.md (las decisiones que cite la historia)
3. docs/01-producto/reglas-negocio.md (las reglas que cite la historia)
4. docs/02-requisitos/historias-usuario.md (la historia y sus criterios CA-*)
5. El estándar del área que toques:
   - Go: docs/03-desarrollo/estandar-backend-go.md y docs/04-arquitectura/backend-go.md
   - Vue: docs/03-desarrollo/estandar-frontend-vue.md y docs/04-arquitectura/frontend.md
   - Visual: docs/03-desarrollo/estandar-diseno-visual.md
   - PostgreSQL: docs/05-backend/estandar-base-datos.md, base-datos.md y migraciones-atlas.md
   - OpenAPI: docs/06-api/estandar-openapi.md
   - Pruebas: docs/03-desarrollo/estrategia-pruebas.md

Reglas innegociables:
- No inventes una respuesta a una duda o a una contradicción. Si la historia necesita
  una decisión que no existe, detente y regístrala en docs/00-control/dudas-pendientes.md.
- Contract-first: si la historia toca HTTP, actualiza api/openapi/ ANTES del handler.
- Toda migración se crea y valida con Atlas, es inmutable y no se ejecuta al arrancar
  la aplicación. Versiona atlas.sum.
- Toda tabla de negocio lleva barbershop_id y RLS. El rol de aplicación no tiene BYPASSRLS.
- Ningún log, prueba, fixture o comentario contiene datos personales, secretos ni tokens.
- Ningún componente Vue define colores, tamaños, radios o tipografías fuera de los tokens.
- No agregues dependencias nuevas sin justificarlas contra el estándar correspondiente.
- No uses `any`, ORM, borrado en cascada, jsonb ni estado global.
- Los comentarios explican decisiones e invariantes; no repiten el código. Todo TODO
  lleva referencia rastreable.

Entrega:
- Rama corta desde main: <tipo>/<issue>-<descripcion>. Conventional Commits.
- Un cambio coherente: código, contrato, migraciones, pruebas y documentación juntos.
- Actualiza docs/00-control/matriz-trazabilidad.md y docs/00-control/historial-cambios.md.
- Reporta con honestidad qué quedó fuera y por qué. No declares terminado lo que no probaste.

Al terminar, enumera cada criterio CA-* de la historia y di con qué prueba concreta
o evidencia se verifica. Un criterio sin verificación es un criterio incumplido.
```

---

## 3. Prompts por historia

### Prompt de `HU-001` · Esquema inicial con aislamiento por barbería

```text
Implementa HU-001 (docs/02-requisitos/historias-usuario.md).

Objetivo: la base de datos nace aislada por barbería y ninguna aplicación puede saltarse ese aislamiento.

Haz exactamente esto:
1. Migración inicial con Atlas en database/migrations/: extensiones necesarias, tabla
   barbershop y tabla de usuarios del área privada. Claves primarias, created_at y
   updated_at con zona horaria, restricciones de integridad explícitas. Tercera forma
   normal; ninguna desnormalización sin motivo escrito.
2. Roles: rol migrador propietario de los objetos y rol de aplicación SIN BYPASSRLS y sin
   propiedad. Documenta cómo se crean en database/README.md.
3. RLS: ENABLE + FORCE en toda tabla con datos de una barbería, con política USING y
   WITH CHECK contra current_setting('app.barbershop_id').
4. database/testdata/: dos barberías con usuarios propios, sin datos personales reales.
5. database/tests/: pruebas SQL que demuestran CA-001-01 a CA-001-04.
6. Verifica CA-001-05 aplicando la migración sobre una base vacía y validando atlas.sum.

No crees ninguna tabla de negocio (servicios, horarios, bloqueos, citas, clientes).
No crees tablas de sesión, recuperación ni idempotencia: llegan con HU-005, HU-008 y HU-004.

Criterios a satisfacer: CA-001-01 a CA-001-07.
```

---

### Prompt de `HU-002` · Contexto de barbería en cada solicitud

```text
Implementa HU-002 (docs/02-requisitos/historias-usuario.md). Depende de HU-001 terminada.

Objetivo: toda operación de datos ocurre dentro de una transacción con app.barbershop_id
fijado con alcance local, y es imposible escribir una consulta de negocio sin contexto.

Haz exactamente esto:
1. En apps/api/internal/platform/database: pool de conexiones configurable, apertura de
   transacción y fijación del contexto con alcance local a la transacción.
2. Expón UN solo punto autorizado para ejecutar trabajo tenant-aware. La firma debe hacer
   imposible omitir el identificador de barbería: si se puede olvidar, el diseño es incorrecto.
3. Un fallo al fijar el contexto aborta la operación. Nunca se continúa con contexto vacío.
4. Comprobación de salud que verifique conectividad sin exponer credenciales ni versiones.
5. Pruebas de integración con PostgreSQL real: dos barberías concurrentes sobre el mismo pool
   (CA-002-02) y verificación de que el contexto no sobrevive a la devolución al pool (CA-002-03).
6. Prueba que falle si el dominio importa el paquete de base de datos o tipos del driver.

El identificador de barbería llega como parámetro; la autenticación es HU-005.
Documenta el patrón obligatorio en apps/api/README.md.

Criterios a satisfacer: CA-002-01 a CA-002-06.
```

---

### Prompt de `HU-003` · Contrato HTTP base y registro sin datos personales

```text
Implementa HU-003 (docs/02-requisitos/historias-usuario.md). Depende de HU-002 terminada.

Objetivo: todos los errores del API se comportan igual, cada solicitud es rastreable y
ningún registro técnico contiene datos personales.

Haz exactamente esto:
1. Incorpora Chi v5 en cmd/api y monta los tres grupos de audiencia: /api/v1/public,
   /api/v1/customer y /api/v1/private. Respeta el orden de middleware de
   docs/04-arquitectura/backend-go.md sección 6.
2. Respuesta de error uniforme RFC 9457 (application/problem+json) con type, title, status,
   detail, instance e identificador de correlación. Un solo helper; ningún handler escribe
   errores a mano.
3. Identificador de solicitud propagado a la respuesta y a cada línea de registro.
4. Logger estructurado con nivel configurable. Prohibido registrar nombre, teléfono, correo,
   contraseña, token o contenido de mensajes. Usa referencias opacas.
5. Recuperación de pánico: 500 con formato uniforme, registro con correlación y proceso vivo.
6. Recurso de otra barbería responde 404, nunca 403, y el cuerpo no revela su existencia.
7. api/openapi/: componentes de error, cabeceras y ejemplos. `pnpm run openapi:lint` y el
   bundle deben pasar.
8. Prueba que inspeccione la salida del logger buscando patrones de correo, teléfono y token
   y falle si encuentra alguno (CA-003-04).

Los módulos de dominio no importan Chi. Si aparece un import de Chi fuera de la capa HTTP,
el diseño está mal.

Criterios a satisfacer: CA-003-01 a CA-003-07.
```

---

### Prompt de `HU-004` · Idempotencia reutilizable

```text
Implementa HU-004 (docs/02-requisitos/historias-usuario.md). Depende de HU-003 terminada.

Objetivo: existe un mecanismo único de idempotencia listo ANTES de la primera operación
crítica, conforme a RN-IDE-01.

Haz exactamente esto:
1. Migración Atlas con la tabla idempotency_record: barbershop_id, clave, huella del
   contenido, resultado almacenado, estado y vencimiento. RLS activa como el resto.
2. Middleware o servicio reutilizable que lea la cabecera Idempotency-Key y aplique:
   - misma clave + mismo contenido -> devuelve el resultado original, sin repetir efectos;
   - misma clave + contenido distinto -> 409, sin ejecutar nada;
   - dos solicitudes simultáneas con la misma clave -> el efecto ocurre exactamente una vez.
3. Registro vencido: la operación se evalúa de nuevo. Operación fallida: no deja una clave
   que impida reintentar legítimamente.
4. Declara la cabecera y sus respuestas en api/openapi/.
5. Pruébalo con un handler de prueba dentro del paquete de pruebas. NO publiques ningún
   endpoint ficticio en el contrato ni en el router de producción.
6. Pruebas de integración con PostgreSQL: repetición exacta, repetición divergente,
   concurrencia real ejecutada con -race, vencimiento y dos barberías (CA-004-04).

Documenta en apps/api/README.md que toda escritura crítica futura debe usar este mecanismo.

Criterios a satisfacer: CA-004-01 a CA-004-06.
```

---

### Prompt de `HU-005` · Inicio de sesión del barbero

```text
ANTES DE EMPEZAR: verifica que exista un DEC-* que resuelva DP-SEG-04 (mecanismo y duración
de la sesión larga). Si no existe, DETENTE, no elijas un mecanismo por tu cuenta y reporta
que la historia está bloqueada.

Implementa HU-005 (docs/02-requisitos/historias-usuario.md). Depende de HU-002 y HU-003.

Objetivo: el barbero entra con correo y contraseña sin que el sistema revele qué correos
existen y sin que ningún secreto quede legible en base de datos o en registros.

Haz exactamente esto:
1. Migración Atlas: tabla de credenciales y tabla de sesiones, ambas con barbershop_id y RLS.
2. Contraseña almacenada con función de derivación resistente a fuerza bruta, con sal única
   por usuario y parámetros documentados en el código como invariante. Nunca cifrado
   reversible ni hash simple.
3. Declara primero en api/openapi/ la operación de inicio de sesión bajo /api/v1/private,
   con su esquema, sus errores y sus ejemplos. Después implementa el handler.
4. Correo inexistente y contraseña incorrecta producen respuestas indistinguibles en cuerpo,
   código y tiempo perceptible (CA-005-02). Cuida el tiempo, no solo el mensaje.
5. La sesión emitida queda asociada al usuario y a su barbería; las solicitudes privadas
   posteriores operan con ese contexto usando el mecanismo de HU-002.
6. Usuario desactivado o eliminado no inicia sesión ni conserva sesiones vigentes.
7. Pruebas: unitarias del servicio, integración HTTP (éxito, credenciales inválidas, correo
   inexistente, aislamiento entre barberías) y verificación de que dos usuarios con la misma
   contraseña producen hashes distintos.

Nada de contraseñas, correos ni material de sesión en logs, errores o respuestas.

Criterios a satisfacer: CA-005-01 a CA-005-07.
```

---

### Prompt de `HU-006` · Sesión persistente y cierre de sesión

```text
ANTES DE EMPEZAR: requiere el mismo DEC-* de DP-SEG-04 que HU-005.

Implementa HU-006 (docs/02-requisitos/historias-usuario.md). Depende de HU-005 terminada.

Objetivo: la sesión sobrevive al cierre del navegador, se puede revocar de verdad y ninguna
ruta privada queda sin protección.

Haz exactamente esto:
1. Vigencia larga conforme al DEC-* vigente, con renovación controlada. Una sesión vencida
   NO se renueva sola.
2. Cierre de sesión que invalida la sesión en el servidor, no solo en el cliente.
   Reutilizar el material anterior debe responder no autorizado (CA-006-02).
3. Middleware de sesión que proteja todo /api/v1/private, con respuesta uniforme cuando
   la sesión falta o venció.
4. Registra el instante de último uso de la sesión para poder auditar accesos.
   Ese registro no incluye datos personales.
5. Escribe una prueba estructural que enumere las rutas privadas registradas y falle si
   alguna no pasa por el middleware de sesión (CA-006-04). Esta prueba es obligatoria:
   protege contra la ruta que alguien agregará dentro de seis meses sin recordarlo.
6. El material de sesión no viaja en la URL, no se registra y no es adivinable.

Criterios a satisfacer: CA-006-01 a CA-006-06.
```

---

### Prompt de `HU-007` · Defensa escalonada contra abuso

```text
ANTES DE EMPEZAR: verifica que exista un DEC-* que resuelva DP-SEG-06 (duración de la ventana
del límite por IP y del escalamiento). Si no existe, DETENTE y repórtalo. El umbral inicial
de 5 solicitudes por IP sí está confirmado por DEC-026; la ventana no.

Implementa HU-007 (docs/02-requisitos/historias-usuario.md). Depende de HU-003 y HU-005.

Haz exactamente esto:
1. Conteo por IP con umbral y ventana configurables por parámetro. Valor inicial del umbral: 5.
   Nada de números incrustados en el código.
2. Superado el umbral, la solicitud exige verificación telefónica ANTES de evaluar la
   contraseña, y responde 429 con el formato uniforme de HU-003 indicando cuándo reintentar.
3. La respuesta al superar el umbral no revela si el correo existe.
4. Transcurrida la ventana sin intentos, el conteo se reinicia solo.
5. Los datos de conteo vencen solos y no contienen correo, nombre ni teléfono.
6. Documenta y prueba cómo se obtiene la IP en el despliegue previsto. Una cabecera que el
   cliente pueda falsificar no puede ser la única fuente (CA-007-07).
7. Pruebas de integración: recorrido del umbral, escalamiento, expiración de la ventana,
   dos IP independientes y evasión por cabecera manipulada.

Criterios a satisfacer: CA-007-01 a CA-007-07.
```

---

### Prompt de `HU-008` · Recuperación de acceso con código

```text
ANTES DE EMPEZAR: verifica que exista un DEC-* que resuelva DP-SEG-05 (canal y proveedor del
código de recuperación). DEC-027 habilita correo y WhatsApp oficial para notificaciones de
turnos, NO para códigos de seguridad: no lo uses como sustituto. Si no existe el DEC-*,
DETENTE y repórtalo.

Implementa HU-008 (docs/02-requisitos/historias-usuario.md). Depende de HU-005 y HU-007.

Haz exactamente esto:
1. Migración Atlas: tabla de códigos de recuperación con hash del código, vencimiento corto,
   contador de intentos, marca de uso, barbershop_id y RLS.
2. Tres operaciones declaradas primero en api/openapi/: solicitar recuperación, verificar
   código y establecer contraseña nueva.
3. El código es de un solo uso, vence en un plazo corto configurable y se invalida por
   completo al agotar los intentos.
4. El código se guarda como hash. El valor enviado no queda en base de datos ni en logs.
5. Solicitar recuperación con un correo registrado y con uno no registrado produce respuestas
   idénticas (CA-008-01).
6. Al establecer la contraseña nueva, invalida TODAS las sesiones activas del usuario.
7. El teléfono se muestra enmascarado en respuestas y registros; nunca completo.
8. El reenvío tiene límite propio y no reinicia el vencimiento del código anterior sin
   invalidarlo.
9. La política mínima de contraseña se valida en el servidor y el error explica qué falta.
10. Adaptador de envío detrás de una interfaz; en pruebas se usa un doble. El proveedor real
    nunca se invoca desde una prueba.

Criterios a satisfacer: CA-008-01 a CA-008-08.
```

---

### Prompt de `HU-009` · Sistema visual base en componentes

```text
Implementa HU-009 (docs/02-requisitos/historias-usuario.md).

Objetivo: existen los componentes base que las pantallas de B0 van a usar, y ninguna pantalla
necesitará inventar un estilo propio.

Haz exactamente esto:
1. En apps/web/src/shared/ui/ crea SOLO los componentes que HU-010, HU-011 y HU-012 usarán:
   botón, campo de texto, alerta, insignia y diálogo. No crees componentes sin uso real.
2. Todo color, medida, radio, sombra y tipografía sale de apps/web/src/styles/tokens.css.
   Cero valores literales. Las props expresan intención (tone="danger"), nunca valores
   (color="#b91c1c").
3. Implementa los patrones de estado de docs/03-desarrollo/estandar-diseno-visual.md
   sección 9: inicial, carga, actualización, vacío, error recuperable, error de campo,
   conflicto y éxito.
4. Foco visible según el estándar, orden de tabulación coherente, área táctil mínima de
   44x44 px y respeto de prefers-reduced-motion.
5. Prueba de componente por cada componente, cubriendo sus estados reales.
6. Verificación de contraste WCAG 2.2 AA y evidencia visual en 320, 360, 768 y 1280 px.

Prohibido por DEC-039: modo oscuro, selector de tema, colores por barbería, CSS libre y
biblioteca visual externa. No agregues una dependencia de iconos sin justificarla.

Criterios a satisfacer: CA-009-01 a CA-009-07.
```

---

### Prompt de `HU-010` · Pantalla de acceso

```text
Implementa HU-010 (docs/02-requisitos/historias-usuario.md). Depende de HU-005 y HU-009.

Objetivo: la primera pantalla del producto, usable con una mano en un celular de gama media
con conexión mala.

Haz exactamente esto:
1. Módulo apps/web/src/modules/auth/, ruta cargada de forma diferida.
2. Composición según docs/03-desarrollo/estandar-diseno-visual.md sección 10: marca discreta,
   título, formulario estrecho y recuperación visible.
3. Cliente tipado generado desde el bundle OpenAPI. Prohibido escribir DTO manuales paralelos.
4. Estados: inicial, enviando, credenciales inválidas, error de red con datos conservados y
   bloqueo por umbral de HU-007 explicado en lenguaje simple.
5. El botón se deshabilita mientras la solicitud está en curso; el doble toque no produce dos
   solicitudes (CA-010-04).
6. Etiquetas asociadas a sus controles, error junto al campo y resumen accesible si hay varios.
7. La contraseña nunca va en la URL, ni se registra, ni queda en el historial de navegación.
8. Pruebas: componente del formulario, E2E de acceso correcto e incorrecto, evidencia
   responsive en los cuatro anchos y verificación con teclado.

Usa los componentes de HU-009. Si necesitas un estilo que no existe, el problema está en
HU-009, no en esta pantalla: arréglalo allá.

Criterios a satisfacer: CA-010-01 a CA-010-08.
```

---

### Prompt de `HU-011` · Pantalla de recuperación de acceso

```text
Implementa HU-011 (docs/02-requisitos/historias-usuario.md). Depende de HU-008, HU-009 y HU-010.

Objetivo: que un barbero recupere su acceso sin llamar a nadie.

Haz exactamente esto:
1. Tres pasos en el módulo auth: solicitar, verificar código, establecer contraseña nueva.
   El indicador muestra siempre paso actual y total.
2. Composición según docs/03-desarrollo/estandar-diseno-visual.md sección 10: paso actual,
   destino enmascarado, campo del código y reenvío con estado.
3. Mensajes distintos y accionables para código incorrecto, vencido y agotado. Sin lenguaje
   técnico y sin nombres de proveedor.
4. El reenvío muestra cuánto falta para poder pedirlo otra vez y lo impide antes de tiempo.
5. La política de contraseña se explica ANTES de escribir, no solo al fallar.
6. Al terminar: confirmación del cambio, aviso de que las sesiones se cerraron y salida al
   acceso.
7. Accesibilidad: el foco se mueve al encabezado de cada paso y el cambio se anuncia a un
   lector de pantalla.
8. Pruebas: componente por paso, E2E con código válido y con código vencido, evidencia
   responsive en 320 y 360 px.

Criterios a satisfacer: CA-011-01 a CA-011-08.
```

---

### Prompt de `HU-012` · Cascarón del panel privado

```text
Implementa HU-012 (docs/02-requisitos/historias-usuario.md). Depende de HU-006, HU-009 y HU-010.

Objetivo: la estructura sobre la que se montarán todas las pantallas privadas de B1 en adelante.

Haz exactamente esto:
1. Estructura autenticada según las estructuras base del estándar visual: cabecera con el
   nombre de la barbería activa, navegación principal y área de contenido.
2. Guarda de ruta: sin sesión válida redirige al acceso conservando el destino pretendido y
   volviendo a él después de entrar (CA-012-02).
3. Manejo central de respuestas no autorizadas: limpia la sesión local UNA sola vez e informa
   el motivo. Un bucle de redirección es un defecto, no un detalle.
4. Patrones globales: carga con esqueleto de geometría aproximada, error recuperable con
   "Reintentar" y aviso de conexión perdida sin perder el estado de la pantalla.
5. Cerrar sesión desde la cabecera invalida la sesión en el servidor y vuelve al acceso.
6. Cada ruta se carga de forma diferida; la navegación no recarga la aplicación completa.
7. Pantalla de inicio PROVISIONAL que muestre el estado de la sesión. No construyas la agenda:
   es B3 y depende de datos que todavía no existen.
8. Pruebas: componente de la guarda y del manejo de no autorizado, E2E de acceso, navegación,
   recarga con sesión persistente y cierre de sesión. Evidencia responsive.

El objetivo real: que agregar una pantalla nueva solo requiera declarar su ruta y su
contenido, sin repetir lógica de sesión ni de errores.

Criterios a satisfacer: CA-012-01 a CA-012-08.
```

---

## 4. Prompts de B1 preparados, todavía bloqueados por B0

> Preparar un prompt no abre el bloque. Antes de ejecutar cualquiera de los siguientes, B0 debe cumplir su criterio de salida y la historia debe estar aprobada por el propietario.

### Prompt de `HU-020` · Configuración básica de la barbería

```text
ANTES DE EMPEZAR: verifica y documenta que B0 cumple su criterio de salida. Si HU-003,
HU-006, HU-009 o HU-012 no están terminadas, o alguna duda abierta impide cerrar B0,
detente. Este prompt no autoriza a saltarse esa dependencia.

Implementa HU-020 (docs/02-requisitos/historias-usuario.md).

Objetivo: un barbero autenticado consulta y actualiza nombre, zona IANA, correo y teléfono
de contacto de SU barbería, y el cascarón privado refleja el nombre guardado sin recargar.

Haz exactamente esto:
1. Contract-first en api/openapi/: define GET y PATCH privados para la sección de barbería
   dentro de settings. Usa nombres técnicos en inglés, JSON camelCase, seguridad explícita,
   x-business-rules [RN-DIS-07, RN-TEN-01], x-decisions [DEC-007, DEC-024], ejemplos y
   respuestas 200, 401, 404, 422 y 500 que correspondan. El request NO acepta barbershopId.
2. Revisa database/modelo-fisico-referencia.sql sección B.1 como insumo, no como migración.
   Crea con Atlas una migración nueva que agregue únicamente contact_email y contact_phone
   a barbershop. No agregues public_slug: pertenece a F-PUB-01. No edites la migración de
   HU-001. Actualiza y valida atlas.sum.
3. En apps/api/internal/modules/shops separa dominio, puertos, caso de uso y adaptadores.
   El caso de uso obtiene el tenant de la identidad autorizada, valida nombre, correo,
   teléfono E.164 y zona contra pg_timezone_names, y ejecuta la escritura mediante el
   contexto transaccional de HU-002. Chi y PostgreSQL no entran al dominio.
4. PATCH modifica solo los campos permitidos. Un valor de zona como COT, UTC-5 o uno
   inexistente produce 422 sin escritura parcial. Cambiar la zona no reescribe ningún
   instante almacenado. Correo y teléfono vacíos se normalizan a ausencia de valor.
5. En apps/web/src/modules/settings crea la sección “Barbería” dentro del panel de HU-012.
   Usa el cliente derivado del bundle OpenAPI y los componentes de HU-009. No llames fetch
   desde el componente ni dupliques DTO. Tras éxito, actualiza el nombre visible del
   cascarón por la API pública del módulo, sin estado global nuevo ni recarga completa.
6. Implementa estados inicial, carga, error de campo, error recuperable y éxito. Conserva
   nombre, zona y contactos ante un fallo recuperable. Etiquetas, foco, anuncio de estado,
   teclado y objetivos táctiles cumplen el estándar visual.
7. Pruebas backend: unitarias de validación y zona; HTTP validado contra el bundle; integración
   PostgreSQL con dos barberías para lectura, actualización, ausencia de contexto y ataque
   con identificador ajeno. Inspecciona logs para demostrar que no contienen contacto.
8. Pruebas frontend: componente de todos los estados y E2E abrir -> editar -> guardar ->
   comprobar cabecera. Adjunta evidencia a 320, 360, 768 y 1280 px, zoom 200 % y teclado.
9. Actualiza la matriz de trazabilidad, el historial documental y los README afectados.

Fuera de alcance: aprovisionar/eliminar barberías, public_slug, reserva, cancelación,
recordatorios, proveedores, personalización visual y reescritura de instantes.

Criterios a satisfacer: CA-020-01 a CA-020-08. Al entregar, mapea cada criterio a una
prueba o evidencia concreta.
```

---

### Prompt de `HU-021` · Registro y listado de barberos

```text
ANTES DE EMPEZAR: verifica que B0 cumple su criterio de salida y que HU-020 está terminada
y aprobada. Si no, detente. No uses este prompt para adelantar B1 sobre una base incompleta.

Implementa HU-021 (docs/02-requisitos/historias-usuario.md).

Objetivo: representar con la misma entidad, API y pantalla a una barbería de una persona
y a otra de cuatro, con aislamiento por tenant y sin adelantar servicios, horarios o agenda.

Haz exactamente esto:
1. Contract-first en api/openapi/: define colección privada de barbers con GET y POST, y
   recurso privado con PATCH limitado al nombre. Declara seguridad, ejemplos, paginación
   solo si el estándar la exige para el límite elegido, x-business-rules [RN-TEN-01],
   x-decisions [DEC-019, DEC-024] y respuestas 200/201, 401, 404, 422 y 500. Nunca aceptes
   barbershopId en el body.
2. Revisa database/modelo-fisico-referencia.sql sección B.2 como diseño de referencia.
   Crea con Atlas una migración nueva para barber con id, barbershop_id, full_name y marcas
   de tiempo con zona, PK, UNIQUE tenant-aware cuando sea destino de FK, FK a barbershop
   con ON DELETE RESTRICT, trigger updated_at, RLS ENABLE + FORCE, políticas y privilegios.
   Implementa solo datos necesarios para HU-021: no agregues comportamiento de borrado,
   desactivación, orden manual ni vínculo automático con staff_user. Actualiza atlas.sum.
3. En apps/api/internal/modules/staff implementa dominio, puertos, casos de uso de listar,
   crear y renombrar, adaptador PostgreSQL y adaptador HTTP. El nombre se recorta, no puede
   quedar vacío y admite como máximo 120 caracteres. No impongas unicidad de nombres: esa
   regla no existe. El tenant proviene de la sesión y toda consulta filtra además por él.
4. Un identificador de otra barbería responde 404 y nunca revela existencia. La barbería
   unipersonal tiene una fila barber normal; no derives el profesional desde barbershop ni
   desde staff_user y no agregues una rama “single barber”.
5. En apps/web/src/modules/staff crea la pantalla “Barberos”: lista, vacío, error recuperable,
   alta y edición del nombre. Usa ruta diferida, cliente generado, componentes de HU-009 y
   el cascarón de HU-012. No muestres todavía servicios ni horarios ficticios.
6. Evita doble envío desde la interfaz. No incorpores DELETE, activate/deactivate, invitación,
   credenciales, permisos, barber_service, working_hours, disponibilidad ni agenda.
7. Pruebas backend: unitarias del nombre y la actualización; HTTP contra bundle; integración
   PostgreSQL real con dos tenants, escritura cruzada, 404 ajeno y escenarios de una y cuatro
   filas con exactamente el mismo camino de código.
8. Pruebas frontend: componente de lista, vacío, alta, edición, 422 y error recuperable; E2E
   añadir -> listar -> renombrar. Adjunta evidencia a 320, 360, 768 y 1280 px, zoom 200 %,
   teclado y foco del formulario o diálogo.
9. Actualiza la matriz de trazabilidad, el historial documental y los README afectados.

Criterios a satisfacer: CA-021-01 a CA-021-08. Al entregar, mapea cada criterio a una
prueba o evidencia concreta y reporta explícitamente el alcance excluido.
```

---

## 5. Prompt de revisión

Para revisar el trabajo de una historia antes de integrarlo. No lo ejecute el mismo agente que implementó.

```text
Revisa la implementación de HU-0XX contra docs/02-requisitos/historias-usuario.md.

Para cada criterio CA-0XX-nn responde: cumplido, incumplido o no verificable, y con qué
prueba o evidencia concreta. No aceptes "parece correcto".

Revisa además:
1. ¿Alguna tabla nueva quedó sin barbershop_id o sin RLS?
2. ¿Algún handler escribe errores fuera del helper uniforme?
3. ¿Algún log, prueba o fixture contiene datos personales, secretos o tokens?
4. ¿El contrato OpenAPI y los handlers dicen lo mismo?
5. ¿Algún componente Vue define color, tamaño, radio o tipografía fuera de los tokens?
6. ¿Hay una migración editada después de aplicada?
7. ¿El dominio importa Chi, el driver de PostgreSQL o tipos de transporte?
8. ¿Se agregó una dependencia sin justificación?
9. ¿La matriz de trazabilidad y el historial de cambios se actualizaron?
10. ¿Se declaró terminado algo que no se probó?

Reporta los hallazgos ordenados por gravedad. Si un criterio no se puede verificar, dilo
en vez de suponer.
```

---

## 6. Mantenimiento de este documento

- Un prompt se corrige cuando la historia cambia, cuando un estándar cambia o cuando el resultado real demostró que el prompt permitía una interpretación equivocada.
- Los prompts restantes de B1 a B6 se redactan al abrir cada bloque. `HU-020` y `HU-021` son una preparación anticipada solicitada por el propietario y conservan una guarda que impide ejecutarlos antes del cierre de B0.
- Este documento no es fuente de verdad de ninguna regla. Si alguien lo cita para justificar una conducta del producto, la cita es inválida.
