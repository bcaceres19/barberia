---
titulo: "Prompt de implementación · Llaves asimétricas para la seguridad del API"
version: "1.0"
estado: "Propuesta · requiere decisión previa del propietario"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-07"
documentos_relacionados:
  - "../00-control/dudas-pendientes.md"
  - "../00-control/registro-decisiones.md"
  - "../02-requisitos/historias-usuario.md"
  - "../04-arquitectura/backend-go.md"
  - "../06-api/estandar-openapi.md"
  - "prompts-implementacion.md"
---

# Prompt de implementación · Llaves asimétricas para la seguridad del API

> **Estado: Propuesta.** Un prompt no es fuente normativa. Este además **no se puede
> ejecutar todavía**: describe una capacidad que resuelve `DP-SEG-04`, una duda abierta.
> Su primer paso es escribir la decisión, no código.

---

## 1. Léase esto antes que el prompt

### 1.1 Esto no es una historia nueva: resuelve una duda existente

`DP-SEG-04` pregunta *"¿qué mecanismo sostiene la sesión larga del barbero y cuánto dura
exactamente?"* y registra dos opciones:

> Cookie `HttpOnly` + `Secure` + `SameSite` con token opaco revocable en base de datos,
> frente a un token firmado sin revocación inmediata. **La revocabilidad es requisito de
> `CA-006-02`.**

Adoptar llaves asimétricas es elegir la segunda familia. La documentación ya advierte que,
por sí sola, **incumple `CA-006-02`**:

> `CA-006-02` · Dado un cierre de sesión, cuando se reutiliza el material de sesión anterior,
> entonces la respuesta es de no autorizado y ninguna operación se ejecuta.

Un token firmado es válido hasta que vence, aunque el barbero haya cerrado sesión. Si el
servidor no consulta nada, no puede saber que fue revocado. Esto no es un detalle de
implementación: es un criterio de aceptación de una historia P0.

### 1.2 Qué gana realmente este proyecto con llaves asimétricas

Conviene ser exacto, porque el argumento habitual —"evita ir a la base de datos en cada
petición"— **no aplica aquí**. `HU-002` establece que toda operación de datos ocurre dentro
de una transacción con `app.barbershop_id` fijado: el viaje a PostgreSQL ya existe.
Verificar la sesión añade una búsqueda por índice a una transacción que de todos modos se
abre.

Lo que sí se gana, y es real:

| Beneficio | Por qué importa aquí |
| --- | --- |
| Contención ante compromiso | Quien obtenga la llave de verificación **no puede emitir** tokens. Con HMAC, verificar y firmar usan el mismo secreto. |
| Separación de responsabilidades | Solo el módulo `auth` posee la llave privada. El resto del sistema verifica. |
| Verificadores futuros | Un segundo servicio, un worker o una app móvil pueden validar sin recibir capacidad de emisión. |
| Rotación auditable | Con `kid` y un conjunto de llaves, rotar no obliga a cerrar todas las sesiones a la vez. |
| Webhooks de proveedores (B5) | `DEC-027` traerá WhatsApp y correo oficiales, que firman sus webhooks asimétricamente. El proyecto necesitará esta capacidad de todos modos. |

Lo que **no** se gana: no elimina la consulta de revocación, no simplifica el despliegue y
no hace el sistema más rápido de forma medible en este tamaño.

### 1.3 El diseño que sí satisface la documentación

Firma asimétrica **más** sesión revocable. No son alternativas: son capas distintas.

```text
Cookie HttpOnly ─── token de acceso firmado con Ed25519, vida corta (10 min)
                    claims: iss, aud, sub, bsid, sid, iat, exp, jti, kid
                              │
                              ├─ 1. verificar firma con la llave pública  → autenticidad
                              ├─ 2. verificar exp / iss / aud             → vigencia
                              └─ 3. leer staff_session por sid            → REVOCACIÓN (CA-006-02)
                                    revoked_at IS NULL AND expires_at > now()

Cookie HttpOnly ─── token de refresco OPACO, aleatorio, vida larga
                    solo su hash SHA-256 vive en staff_session.token_hash
                    revocar = escribir revoked_at. Efecto inmediato.
```

El paso 3 es el que cumple `CA-006-02`. Si el propietario decide **omitirlo** para tener
sesiones verdaderamente sin estado, entonces `CA-006-02` debe reescribirse y `HU-006` cambia
de alcance. Esa es una decisión de producto, no una elección del implementador.

---

## 2. Decisión que debe existir antes de codificar

El prompt de la sección 3 **empieza escribiendo `DEC-040`**. Debe cerrar, como mínimo:

| Punto | Por qué no se puede dejar abierto |
| --- | --- |
| Algoritmo y formato del token | Determina la dependencia y la superficie de ataque |
| Vida del token de acceso | Es la ventana en la que un token robado sigue sirviendo |
| Vida del token de refresco | Es la "sesión larga" que `DEC-026` menciona sin cuantificar |
| ¿Se consulta la sesión en cada petición? | Es literalmente `CA-006-02`: sí o no, con consecuencias |
| Dónde vive la llave privada | Sin esto no hay despliegue seguro posible |
| Período de rotación y de solapamiento | Sin esto, rotar significa expulsar a todos los barberos |
| Rotación del token de refresco | Determina si un refresco robado es detectable |

Valores **propuestos** para que el propietario los confirme o corrija (no son decisión):

- algoritmo **Ed25519 (EdDSA)**;
- acceso **10 minutos**; refresco **30 días** con renovación deslizante;
- **sí** se consulta `staff_session` en cada petición privada;
- llave privada en variable de entorno o gestor de secretos, **nunca** en el repositorio;
- rotación cada **90 días**, con la llave anterior aceptada para verificación **24 horas**;
- refresco **rotatorio**: cada uso emite uno nuevo e invalida el anterior.

---

## 3. Prompt

> Precede este prompt con el **preámbulo obligatorio** de
> [prompts-implementacion.md](prompts-implementacion.md) sección 2.

```text
CONTEXTO: implementas el mecanismo de sesión del área privada usando criptografía
asimétrica. Esto resuelve DP-SEG-04 y habilita HU-005 y HU-006.

PASO 0 — OBLIGATORIO, ANTES DE CUALQUIER CÓDIGO.
Verifica si existe en docs/00-control/registro-decisiones.md una decisión que resuelva
DP-SEG-04. Si NO existe:
  a) NO escribas código de sesión.
  b) Redacta la propuesta de DEC-040 con el formato exacto de las decisiones existentes:
     fecha, decisión, responsable, motivo, alternativas descartadas, documentos afectados
     y fuente. Debe cerrar los siete puntos de prompt-llaves-asimetricas.md sección 2.
  c) Declara de forma explícita si la decisión conserva o modifica CA-006-02. Si el
     propietario quiere sesiones sin consulta a base de datos, CA-006-02 y HU-006 deben
     reescribirse y eso se dice en la propuesta, no se resuelve en el código.
  d) Actualiza la fila de DP-SEG-04 en docs/00-control/dudas-pendientes.md.
  e) DETENTE y reporta que la decisión requiere aprobación del propietario.
No elijas tú el algoritmo, la vigencia ni la política de rotación: esos son exactamente
los puntos que la duda deja abiertos.

A partir de aquí, todo asume que DEC-040 existe y aprueba el diseño híbrido
(firma asimétrica + sesión revocable). Si aprobó otro diseño, sigue ese.

────────────────────────────────────────────────────────────────────────
1. DEPENDENCIA CRIPTOGRÁFICA
────────────────────────────────────────────────────────────────────────
La generación y verificación Ed25519 usan crypto/ed25519 de la biblioteca estándar. No
agregues una dependencia para eso.

Para el formato del token, evalúa y justifica por escrito, conforme a
estandar-backend-go.md §5.21:
  - PASETO v4.public: el formato fija el algoritmo, por lo que la confusión de algoritmos
    y "alg: none" no existen como clase de defecto;
  - JWT con una biblioteca vetada: más común, pero exige fijar el algoritmo esperado en el
    verificador y NUNCA leerlo del encabezado del token;
  - formato propio mínimo: NO. Escribir un formato de token a mano es exactamente el tipo
    de criptografía artesanal que produce vulnerabilidades silenciosas.
Escribe la justificación en el ADR o en el comentario del paquete, con licencia, tamaño y
superficie transitiva.

Si eliges JWT: el verificador declara el algoritmo esperado y rechaza cualquier token cuyo
encabezado no coincida. Un verificador que confía en el "alg" del token es un defecto
crítico, no una preferencia de estilo.

────────────────────────────────────────────────────────────────────────
2. PAQUETE DE LLAVES
────────────────────────────────────────────────────────────────────────
Crea apps/api/internal/platform/keyring (nombre que describa una responsabilidad; no lo
metas en "utils" ni en "crypto" genérico).

Debe ofrecer:
  - carga del conjunto de llaves desde configuración, mediante el paquete config, que es
    el único autorizado a leer el entorno;
  - una llave ACTIVA para firmar, identificada por un kid estable;
  - cero o más llaves ACEPTADAS solo para verificar, durante el solapamiento de rotación;
  - Sign(ctx, claims) y Verify(ctx, token) con errores tipados y distinguibles:
    firma inválida, token vencido, kid desconocido, emisor o audiencia incorrectos.

Reglas innegociables del material de llaves:
  - La llave privada NUNCA se escribe en el repositorio, en un log, en un mensaje de error,
    en una prueba ni en un fixture. Añade el patrón correspondiente a .gitignore.
  - El tipo que envuelve la llave privada implementa String() y MarshalJSON() devolviendo
    un valor redactado, para que un log accidental no la exponga. Esto es una defensa real:
    la mayoría de las fugas de secretos son un %v distraído.
  - La llave privada no se expone por ningún endpoint, ni siquiera de diagnóstico.
  - La llave pública sí puede publicarse. Si expones un conjunto de llaves, es de solo
    lectura, cacheable y NO revela cuál es la activa para firmar más allá del kid.
  - Arrancar sin llave activa configurada es un error fatal de arranque, no una advertencia
    con generación automática. Una llave generada al vuelo invalidaría todas las sesiones
    en cada reinicio y en cada réplica.

Genera las llaves con un comando explícito bajo apps/api/cmd/ o un script documentado, que
imprima la llave pública y escriba la privada donde el operador indique. Documenta el
procedimiento en apps/api/README.md.

────────────────────────────────────────────────────────────────────────
3. CLAIMS DEL TOKEN DE ACCESO
────────────────────────────────────────────────────────────────────────
Incluye exactamente: iss, aud, sub (staff_user_id), bsid (barbershop_id), sid (id de
staff_session), iat, exp, jti y el kid en el encabezado o su equivalente en el formato.

PROHIBIDO por RN-DAT-02: correo, nombre, teléfono, rol legible, o cualquier dato personal.
Un token viaja por logs de proxy, historiales y capturas. Solo identificadores opacos.

Valida SIEMPRE iss y aud. Un token emitido para otra audiencia no se acepta "porque la
firma es válida": la firma solo prueba quién lo emitió, no para qué.

Tolerancia de reloj: como máximo 60 segundos, configurable, aplicada solo a iat/nbf.
No amplíes la tolerancia para "arreglar" un servidor con la hora mal puesta.

────────────────────────────────────────────────────────────────────────
4. VERIFICACIÓN EN CADA SOLICITUD PRIVADA
────────────────────────────────────────────────────────────────────────
El middleware de /api/v1/private ejecuta, en este orden y sin saltarse ninguno:
  1. verificar la firma con la llave del kid indicado; kid desconocido -> no autorizado;
  2. verificar exp, iss y aud;
  3. abrir la transacción con el contexto de HU-002 usando bsid;
  4. leer staff_session por sid y exigir revoked_at IS NULL y expires_at > now();
  5. exigir que la sesión pertenezca al mismo staff_user_id y barbershop_id del token.
     Si no coinciden, no autorizado Y registro de seguridad: significa que un token válido
     apunta a una sesión que no le corresponde;
  6. exigir que el usuario siga activo (CA-005-07);
  7. actualizar last_used_at.

El paso 4 es el que cumple CA-006-02. No lo conviertas en caché en memoria sin invalidación
entre réplicas: con dos procesos, un cierre de sesión no llegaría al otro.

El identificador de barbería del contexto sale del token verificado y de la sesión, NUNCA
del cuerpo, de la ruta ni de un encabezado enviado por el cliente (backend-go.md §7,
RN-TEN-01). Si aparece un barbershop_id en una petición privada, ignóralo.

────────────────────────────────────────────────────────────────────────
5. TRANSPORTE
────────────────────────────────────────────────────────────────────────
Cookies, no encabezado Authorization, porque el consumidor es la SPA del mismo origen y una
cookie HttpOnly no es legible por JavaScript inyectado:
  - HttpOnly, Secure, SameSite=Strict, Path acotado, sin Domain amplio;
  - la cookie de refresco solo se envía a la ruta de refresco, no a todo el API;
  - ningún token viaja en la URL, en un fragmento ni en almacenamiento persistente
    (CA-006-05);
  - protección CSRF explícita, porque SameSite no es suficiente por sí solo en todos los
    navegadores y flujos. Documenta el mecanismo elegido y pruébalo.

────────────────────────────────────────────────────────────────────────
6. REFRESCO Y REVOCACIÓN
────────────────────────────────────────────────────────────────────────
  - El token de refresco es aleatorio y opaco, mínimo 256 bits de crypto/rand. NUNCA
    math/rand. Solo se guarda su hash SHA-256 en staff_session.token_hash.
  - Refresco rotatorio: cada uso emite uno nuevo e invalida el anterior en la misma
    transacción.
  - Detección de reutilización: si llega un refresco ya usado, revoca TODA la cadena de
    sesiones de ese usuario y registra el evento. Un refresco reutilizado significa robo o
    condición de carrera; ante la duda se cierra sesión, no se continúa.
  - Cerrar sesión escribe revoked_at. Es inmediato y verificable desde el servidor.
  - Cambiar la contraseña revoca TODAS las sesiones del usuario (CA-008-05).
  - Cerrar sesión en un dispositivo no toca los demás, salvo petición explícita
    (CA-006-06).

────────────────────────────────────────────────────────────────────────
7. ROTACIÓN DE LLAVES
────────────────────────────────────────────────────────────────────────
Procedimiento, documentado y ensayado al menos una vez antes del piloto:
  1. generar la llave nueva y publicarla como ACEPTADA para verificación;
  2. desplegar; todas las réplicas ya verifican con ambas;
  3. promover la nueva a ACTIVA para firmar;
  4. mantener la anterior como aceptada al menos la vida del token de acceso;
  5. retirarla.
Saltarse el paso 2 invalida los tokens en vuelo. Un despliegue que expulse a un barbero en
plena jornada es un fallo operativo, no un efecto secundario aceptable.

Una llave comprometida se retira de inmediato y se revocan todas las sesiones. Ese camino
se documenta en el procedimiento de incidentes de DEC-026 (F-OPS-09).

────────────────────────────────────────────────────────────────────────
8. CONTRATO Y OBSERVABILIDAD
────────────────────────────────────────────────────────────────────────
  - Declara primero en api/openapi/ las operaciones de sesión, refresco y cierre, con sus
    esquemas, errores RFC 9457 y ejemplos. Después implementa los handlers.
  - Los ejemplos del contrato NO contienen un token con apariencia real.
  - Registra: emisión, refresco, revocación, firma inválida, kid desconocido y token
    vencido. Con identificadores opacos y request_id. Nunca el token, ni siquiera truncado:
    un prefijo de token sigue siendo material de sesión.
  - Un pico de firmas inválidas es una señal de seguridad; debe poder consultarse.

────────────────────────────────────────────────────────────────────────
9. PRUEBAS OBLIGATORIAS
────────────────────────────────────────────────────────────────────────
Unitarias del keyring:
  - firma y verificación en el camino feliz;
  - token con un carácter alterado -> rechazado;
  - token firmado con OTRA llave -> rechazado;
  - kid desconocido -> rechazado;
  - vencido -> rechazado; dentro de la tolerancia de reloj -> aceptado;
  - iss o aud incorrectos -> rechazados;
  - si usas JWT: token con "alg" manipulado, incluido "none" y un intento de confusión
    a HMAC usando la llave pública como secreto -> rechazados. Esta prueba es obligatoria
    y no se omite "porque la biblioteca ya lo maneja";
  - String() y MarshalJSON() del tipo de llave privada NO revelan material.

Integración con PostgreSQL real:
  - token válido de una sesión revocada -> no autorizado (CA-006-02);
  - token válido cuya sesión pertenece a otro usuario -> no autorizado;
  - token de la barbería A no alcanza ningún recurso de B, verificado contra un endpoint
    privado real (CA-005-05);
  - usuario desactivado -> sin acceso y sin sesiones vigentes;
  - cambio de contraseña -> todas las sesiones caen;
  - refresco reutilizado -> toda la cadena revocada;
  - rotación: token firmado con la llave anterior sigue verificándose durante el
    solapamiento y deja de hacerlo después.

Prueba estructural: enumera las rutas privadas registradas y falla si alguna no pasa por el
middleware de sesión (CA-006-04).

Ejecuta go test -race para el refresco concurrente.

────────────────────────────────────────────────────────────────────────
10. LO QUE NO DEBES HACER
────────────────────────────────────────────────────────────────────────
  - No generes llaves automáticamente al arrancar.
  - No guardes la llave privada en el repositorio, en la imagen del contenedor ni en una
    variable de entorno registrada en logs de despliegue.
  - No uses la misma llave para sesiones, enlaces públicos de turno y webhooks: son tres
    propósitos y tres pares de llaves.
  - No firmes los tokens de acceso al turno de DEC-022 con este mecanismo sin una decisión
    aparte: RN-DAT-03 exige revocarlos al anonimizar, y un token firmado no se revoca solo.
  - No introduzcas un caché de revocación en memoria mientras existan varias réplicas.
  - No amplíes la vida del token de acceso para reducir refrescos: es la ventana de
    exposición.

Al terminar, enumera CA-005-01 a CA-005-07 y CA-006-01 a CA-006-06 y di con qué prueba
concreta se verifica cada uno. Declara explícitamente si CA-006-02 se cumple de forma
inmediata o con retraso, y cuánto.
```

---

## 4. Efecto de esta decisión en el modelo de datos

`database/modelo-fisico-referencia.sql` sección A.2 ya diseñó `staff_session` con
`token_hash`, `expires_at`, `revoked_at` y `last_used_at`. Ese diseño **sirve tal cual**
para el token de refresco de este esquema. Si `DEC-040` adopta refresco rotatorio con
detección de reutilización, la tabla necesita además:

```sql
-- Cadena de refrescos, para poder revocar toda la familia ante una reutilización.
ALTER TABLE staff_session
  ADD COLUMN parent_session_id uuid,
  ADD COLUMN chain_id          uuid NOT NULL DEFAULT gen_random_uuid(),
  ADD CONSTRAINT staff_session_barbershop_id_parent_id_fk
    FOREIGN KEY (barbershop_id, parent_session_id)
    REFERENCES staff_session (barbershop_id, id) ON DELETE SET NULL;

CREATE INDEX idx_staff_session_chain ON staff_session (barbershop_id, chain_id)
  WHERE revoked_at IS NULL;
```

Ese `ALTER` se agrega a la migración de `HU-005`, no como migración separada: `HU-005`
todavía no existe como migración aplicada, así que no hay inmutabilidad que respetar.

---

## 5. Riesgo que este documento deja consignado

Si el propietario aprueba llaves asimétricas **sin** la consulta de sesión del paso 4, el
resultado es una regresión de seguridad frente al diseño documentado: un barbero que cierra
sesión en un dispositivo perdido seguirá teniendo acceso durante toda la vida del token.
Con 10 minutos es un riesgo acotado y discutible; con vigencias largas —que es lo que
`DEC-026` pide con "sesión larga"— sería inaceptable.

La combinación peligrosa es exactamente **token firmado de vida larga sin verificación de
revocación**. El prompt está escrito para que esa combinación no pueda aparecer por
descuido.
