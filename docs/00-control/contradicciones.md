---
titulo: "Registro de contradicciones"
version: "1.8"
estado: "Vigente"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-25"
documentos_relacionados:
  - "../../AGENTS.md"
  - "registro-decisiones.md"
  - "dudas-pendientes.md"
  - "../01-producto/alcance-mvp.md"
  - "../01-producto/prioridades.md"
  - "../02-requisitos/historias-usuario.md"
  - "../04-arquitectura/backend-go.md"
---

# Registro de contradicciones

## 1. Regla de uso

Una contradicción existe cuando dos afirmaciones vigentes no pueden cumplirse al mismo tiempo. No se resuelve ocultando una de ellas: se elige una alternativa mediante una decisión `DEC-*`, se actualizan todos los documentos afectados y se conserva aquí el rastro de la resolución.

Estados permitidos: `Abierta`, `En análisis`, `Resuelta` y `Descartada por falso conflicto`.

## 2. Resumen

| Código | Contradicción | Estado | Dependencia | Momento límite |
| --- | --- | --- | --- | --- |
| `CT-001` | Recordatorios automáticos P0 frente a maquinaria P0 que depende de ellos | **Resuelta** | `DEC-027`, `DEC-032` | Cerrada el 2026-08-05 |
| `CT-002` | Vocabulario de `appointment_history.event_type`: español con puntos frente a regla general en inglés | **Resuelta** | `DEC-041` | Cerrada el 2026-08-11 |
| `CT-003` | El inicio de sesión está bajo `/api/v1/private`, pero toda ruta de ese prefijo exige una sesión previa | **Resuelta** | `DEC-055` | Cerrada el 2026-08-13 |
| `CT-004` | `HU-010` debe llegar al panel privado que solo construye `HU-012`, pero `HU-012` depende de `HU-010` | **Resuelta** | `DEC-056` | Cerrada el 2026-08-13 |
| `CT-005` | Umbral de `HU-007`: “superar 5 solicitudes” frente a escalar cuando el conteo alcanza 5 en el SQL de referencia | **Resuelta** | `DEC-061` | Cerrada el 2026-08-17 |
| `CT-006` | Recuperación no enumerable con respuesta idéntica frente a mostrar el destino real enmascarado | **Resuelta** | `DEC-065` | Cerrada el 2026-08-17 |
| `CT-008` | Modelo físico de B2 propone `ON DELETE CASCADE` frente a la prohibición de borrado en cascada de `AGENTS.md` | **Resuelta** | `DEC-070` | Cerrada el 2026-08-25 |

## 3. Contradicciones detalladas

### CT-001 · Recordatorios automáticos: ¿P0 o P1?

- **Detectada:** 2026-07-30.
- **Registrada formalmente:** 2026-08-05.
- **Estado:** **Resuelta**.
- **Responsable de resolver:** propietario del proyecto.
- **Documentos originalmente en conflicto:** las versiones anteriores de [alcance-mvp.md](../01-producto/alcance-mvp.md) incluían `F-NOT-01` en P0 y dejaban `F-NOT-03` como elevación propuesta; [prioridades.md](../01-producto/prioridades.md) conservaba la misma tensión.
- **Contradicción:** `F-NOT-01` exige invalidar y regenerar recordatorios, pero esa maquinaria no tiene objeto si `F-NOT-03`, que crea los recordatorios automáticos, no forma parte del mismo alcance.
- **Impacto:** se podría construir infraestructura sin función visible y el piloto no podría atribuir una reducción de olvidos a recordatorios que no existen.
- **Opciones:** (1) elevar `F-NOT-03` a P0 y mantener `F-NOT-01` en P0; (2) bajar ambos a P1 y arrancar sin recordatorios; (3) conservar solo notificaciones inmediatas de cambio, sin recordatorios programados.
- **Resolución:** opción 1. `F-NOT-03` se eleva a P0 y se construye con `F-NOT-01` y `F-NOT-02` (`DEC-032`).
- **Canales que permiten el cierre:** correo y WhatsApp oficial configurables por barbería y evento (`DEC-027`).
- **Evidencia:** `respuesta-manuales/respuesta-dudas-pendientes.txt`, líneas 94–104 y 134–136.

### CT-002 · Vocabulario de `appointment_history.event_type`

- **Detectada:** 2026-08-11, durante `docs/05-backend/revision-ddl-seguridad-2026-08-11.md`.
- **Registrada formalmente:** 2026-08-11.
- **Estado:** **Resuelta**.
- **Responsable de resolver:** propietario del proyecto.
- **Documentos originalmente en conflicto:** [estados-citas.md](../02-requisitos/estados-citas.md), sección 9 (valores en español con puntos: `cita.creada`, `cita.reprogramada`, `cita.servicio_modificado`, `cita.completada`, `cita.cancelada_por_cliente`, `cita.cancelada_por_barbero`, `cita.no_asistio`, `cita.estado_corregido`) frente a la sección 11 del mismo documento (regla general: "los valores se escriben en inglés y en minúsculas").
- **Contradicción:** un mismo documento fija una regla general de idioma para valores almacenados y luego enumera valores concretos de `event_type` que la incumplen.
- **Impacto:** una migración copiada literalmente de la sección 9 quedaría en español, inconsistente con `status` y con el resto del esquema en inglés.
- **Opciones:** (1) adoptar inglés `snake_case` para `event_type` y corregir la sección 9; (2) conservar el español de la sección 9 y excluir `event_type` de la regla general de la sección 11.
- **Resolución:** opción 1 (`DEC-041`). Valores: `appointment_created`, `appointment_rescheduled`, `appointment_service_changed`, `appointment_completed`, `appointment_cancelled_by_customer`, `appointment_cancelled_by_barber`, `appointment_no_show`, `appointment_status_corrected`.
- **Evidencia:** `docs/05-backend/revision-ddl-seguridad-2026-08-11.md`, sección 5, punto 2.

### CT-003 · Inicio de sesión dentro de un prefijo que exige sesión

- **Detectada y registrada:** 2026-08-13, al preparar el prompt individual de `HU-005` y `HU-006`.
- **Estado:** **Resuelta**.
- **Responsable de resolver:** propietario del proyecto.
- **Documentos en conflicto:** [historias-usuario.md](../02-requisitos/historias-usuario.md), alcance incluido de `HU-005` (inicio de sesión bajo `/api/v1/private`) y `CA-006-04` de `HU-006` (toda ruta bajo `/api/v1/private` exige sesión válida); [backend-go.md](../04-arquitectura/backend-go.md), secciones 6 y 7, ubica `/api/v1/private/auth/...` después del middleware de autenticación.
- **Contradicción:** el barbero todavía no posee una sesión cuando llama al inicio de sesión. Exigir autenticación a todas las rutas del prefijo vuelve circular esa operación; dejarla sin middleware incumple la formulación literal de `CA-006-04` y su prueba estructural.
- **Impacto:** el contrato OpenAPI, el árbol Chi y la prueba que inventaría todas las rutas privadas no pueden coincidir sin una excepción o un cambio de audiencia explícitos.
- **Opciones:** (1) conservar el login bajo `/api/v1/private/auth/...`, definir una lista mínima y cerrada de operaciones de arranque sin sesión y ajustar `CA-006-04`; (2) mover el login a una audiencia no autenticada, por ejemplo `/api/v1/public/auth/...`, y actualizar HU, arquitectura y contrato; (3) crear un prefijo de autenticación separado y documentar su seguridad.
- **Resolución:** opción 2. El login se mueve a `/api/v1/public/auth/login` (`DEC-055`); `CA-006-04` no requiere ninguna excepción porque el login ya no vive bajo `/private`.
- **Evidencia:** `HU-005` (alcance incluido), `CA-006-04`, `backend-go.md` §§6–7, `DEC-055`.

### CT-004 · Destino posterior al acceso antes de que exista el panel

- **Detectada y registrada:** 2026-08-13, al preparar el prompt individual de `HU-010`.
- **Estado:** **Resuelta**.
- **Responsable de resolver:** propietario del proyecto.
- **Documentos en conflicto:** [historias-usuario.md](../02-requisitos/historias-usuario.md), `CA-010-01` exige que un acceso correcto lleve al panel privado; la misma historia declara que `HU-010` bloquea `HU-012`, mientras `HU-012` es la historia que construye el cascarón del panel y depende de `HU-010`.
- **Contradicción:** `HU-010` no puede demostrar su destino de éxito contra un panel real sin implementar parte de `HU-012`, pero `HU-012` no puede empezar hasta que `HU-010` esté terminada.
- **Impacto:** el implementador tendría que inventar una pantalla o ruta provisional, ampliar el alcance de `HU-010` o declarar cumplido `CA-010-01` sin el resultado observable exigido.
- **Opciones:** (1) dividir `CA-010-01`: `HU-010` autentica y conserva un destino aprobado, y `HU-012` completa/verifica la llegada al panel; (2) mover un cascarón privado mínimo y expresamente delimitado a `HU-010`, reduciendo después el alcance de `HU-012`; (3) reordenar y redefinir dependencias para construir primero un destino privado real sin mezclar ambas HU.
- **Resolución:** opción 1 (`DEC-056`). `HU-010` navega a `/panel`, ruta privada real y protegida con un guard mínimo propio; `HU-012` reutiliza y generaliza ese guard y construye ahí la cabecera, navegación y demás estructura del cascarón. La llegada al panel completo se verifica en `CA-012-01`/`CA-012-02` de `HU-012`, no en `HU-010`.
- **Evidencia:** `HU-010` (`Depende de`, `Bloquea`, `CA-010-01`) y `HU-012` (`Depende de`, alcance incluido), `DEC-056`.

### CT-005 · Límite de cinco solicitudes: alcanzar frente a superar el umbral

- **Detectada y registrada:** 2026-08-14, al preparar el prompt individual de `HU-007` (issue `#57`).
- **Estado:** **Abierta**.
- **Responsable de resolver:** propietario del proyecto.
- **Documentos en conflicto:** `DEC-026` y `HU-007` describen un umbral inicial de cinco solicitudes y escalamiento “al superarlo”; `CA-007-02` dice “superado el umbral”; `database/modelo-fisico-referencia.sql`, función `login_throttle_register_attempt`, fija `escalated_until` cuando el conteo nuevo cumple `>= p_threshold` y comenta que escala al “alcanzar” el umbral.
- **Contradicción:** con `p_threshold = 5`, el SQL de referencia escala en la quinta solicitud, mientras la lectura literal de “superar cinco” permite evaluar cinco y escala desde la sexta. La diferencia cambia qué solicitud evalúa la contraseña y las pruebas de frontera.
- **Impacto:** implementar sin resolución haría que OpenAPI, servicio, SQL, frontend y pruebas pudieran aprobar semánticas incompatibles; un cambio posterior afectaría seguridad y podría bloquear antes de lo aprobado a un barbero legítimo.
- **Opciones:** (1) cinco solicitudes se evalúan normalmente y la sexta exige el reto; (2) la quinta ya exige el reto; (3) redefinir el parámetro como “cantidad de solicitudes normales permitidas” y alinear todos los textos/SQL con esa semántica explícita.
- **Resolución:** opción 1. Las cinco primeras solicitudes se evalúan con normalidad; la sexta exige el reto telefónico (`DEC-061`, coordinada con `DEC-062`/`DP-SEG-10`). Se corrige `login_throttle_register_attempt` en `database/modelo-fisico-referencia.sql` (`>= p_threshold` → `> p_threshold`).
- **Evidencia:** `DEC-026`; `HU-007`, `CA-007-01/02`; `database/modelo-fisico-referencia.sql` sección A.4; `DEC-061`.

### CT-006 · Respuesta idéntica frente a destino real enmascarado

- **Detectada y registrada:** 2026-08-14, al preparar el prompt individual de `HU-008` (issue `#58`).
- **Estado:** **Abierta**.
- **Responsable de resolver:** propietario del proyecto.
- **Documentos en conflicto:** `CA-008-01` exige que solicitar recuperación para un correo registrado y uno inexistente produzca respuestas idénticas; el alcance de `HU-008` y `CA-008-06` exigen mostrar el teléfono enmascarado en respuestas; `HU-011`/`CA-011-02` exigen mostrar el destino enmascarado durante el flujo.
- **Contradicción:** para un correo inexistente no existe teléfono/correo real que enmascarar. Devolver el destino real enmascarado solo para cuentas existentes cambia el cuerpo y enumera la cuenta; inventar un destino aparente o retrasar su exposición cambia el comportamiento que las fuentes no decidieron.
- **Impacto:** el contrato de solicitud, el flujo de tres pasos y las pruebas de no enumeración no pueden fijar un schema/ejemplo único sin elegir cuándo y cómo se muestra el destino.
- **Opciones:** (1) la solicitud siempre devuelve un mensaje genérico sin destino y el destino se muestra solo después de una prueba que ya demuestre posesión; (2) siempre devuelve un marcador fijo no derivado del dato real; (3) permite un cuerpo distinto y redefine explícitamente qué significa “idéntica”, aceptando y mitigando el riesgo de enumeración; (4) otra alternativa aprobada que conserve ambos objetivos.
- **Resolución:** opción 1 (`DEC-065`). El paso de solicitud siempre responde idéntico y sin destino; el destino enmascarado solo aparece en la respuesta exitosa de verificación, que ya exige haber recibido y transcrito el código real.
- **Evidencia:** `HU-008` alcance/`CA-008-01`/`CA-008-06`; `HU-011` alcance/`CA-011-02`; `DEC-065`.

### CT-007 · Mensajes distintos por causa frente al error uniforme de verificación

- **Detectada y registrada:** 2026-08-23, al implementar el prompt individual de `HU-011` (issue `#64`).
- **Estado:** **Resuelta** (misma entrada; ver resolución).
- **Responsable de resolver:** propietario del proyecto (decisión ya tomada en `DEC-064`/`DEC-065`, aplicada aquí por primera vez a `HU-011`).
- **Documentos en conflicto:** `CA-011-03` exige que un código incorrecto, vencido o agotado produzca "mensajes distintos y accionables"; `DEC-064`/`DEC-065`, ya implementadas en `HU-008` (`apps/api/internal/modules/auth/recovery.go`, función `Verify`: "cualquier fallo —código incorrecto, vencido, agotado o cuenta inexistente— devuelve exactamente el mismo error uniforme"), fijan que `POST /recovery/verify` responde el mismo `401` sin distinguir el motivo, por diseño de no enumeración.
- **Contradicción:** el backend integrado no expone ningún campo (`status`, `code` ni cabecera) que distinga incorrecto de vencido de agotado; cualquier mensaje distinto en el cliente tendría que inventar esa distinción sin que el servidor la respalde, o adivinarla, lo que el catálogo de prompts prohíbe explícitamente ("no inventes... códigos de error... que el contrato no exponga").
- **Impacto:** implementar `CA-011-03` de forma literal exigiría o bien romper la no enumeración (que el servidor sí distinga y por tanto revele más de lo que `DEC-064` decidió), o bien fabricar una distinción ficticia en el cliente sin base real. Ninguna de las dos es aceptable.
- **Opciones:** (1) un único mensaje uniforme ("El código no es correcto o ya venció") para las tres causas, consistente con el error real del servidor; (2) pedir al backend un `code` más granular, reabriendo `DEC-064`/`DEC-065` y el riesgo de enumeración que ya resolvieron; (3) inventar en el cliente una distinción sin respaldo del servidor.
- **Resolución:** opción 1, aplicando el mismo criterio que `DEC-064`/`DEC-065` ya fijaron para este mismo endpoint, y el mismo patrón que `HU-007` ya usa en producción para su reto telefónico análogo (`PhoneChallengeForm.vue`: "El código no es válido o venció", un único mensaje para incorrecto/vencido/agotado). `HU-011` implementa el paso de verificación con este único mensaje; el reenvío con cuenta regresiva (`CA-011-04`) sigue siendo un estado de cliente aparte, no derivado de este error.
- **Evidencia:** `HU-011` alcance/`CA-011-03`; `apps/api/internal/modules/auth/recovery.go` (`Verify`); `apps/api/README.md` sección "No enumeración (`DEC-065`) y destino enmascarado"; `apps/web/src/modules/auth/components/PhoneChallengeForm.vue` (precedente de `HU-007`); `apps/web/src/modules/auth/components/RecoveryVerifyStep.vue`.

### CT-008 · Borrado en cascada del modelo de B2 frente a la regla del repositorio

- **Detectada y registrada:** 2026-08-25, al preparar las historias y prompts de B2.
- **Estado:** **Resuelta**.
- **Responsable de resolver:** propietario del proyecto.
- **Documentos en conflicto:** `AGENTS.md` prohíbe el borrado en cascada; `database/modelo-fisico-referencia.sql`, sección C, proponía `ON DELETE CASCADE` en las FK de `working_hour`, `working_hour_override`, `working_hour_override_segment`, `time_block_series`, `time_block_series_date`, `time_block_series_exception` y `time_block` hacia `barber` o sus entidades padre.
- **Contradicción:** no se puede copiar literalmente el modelo de referencia y cumplir simultáneamente la regla de no borrado en cascada. La decisión afecta el comportamiento ante baja o limpieza administrativa de un barbero y las migraciones tenant-aware.
- **Impacto:** una migración ejecutada sin resolverla podría eliminar horarios, excepciones o bloqueos sin una operación explícita y perder evidencia operativa; cambiar las FK después de aplicarlas exige roll-forward.
- **Opciones:** (1) sustituir las cascadas por `RESTRICT`/desactivación y exigir limpieza o retención explícita; (2) aprobar una excepción limitada y documentada para tablas de configuración, actualizando `AGENTS.md`; (3) otra alternativa que conserve la trazabilidad y no borre datos de negocio implícitamente.
- **Resolución:** opción 1. Las siete FK pasan a `ON DELETE RESTRICT`; `barber` no tiene borrado físico en su alcance vigente (`HU-021`), así que `RESTRICT` no bloquea ninguna operación existente y deja explícita cualquier limpieza futura (`DEC-070`).
- **Evidencia:** `AGENTS.md`, sección Calidad; `database/modelo-fisico-referencia.sql`, sección C; `docs/05-backend/estandar-base-datos.md`, secciones 6, 9 y 10; `docs/00-control/registro-decisiones.md`, `DEC-070`.

---
## 4. Historial de estado

| Fecha | Código | Cambio | Evidencia |
| --- | --- | --- | --- |
| 2026-07-30 | `CT-001` | Detectada en alcance y prioridades | Versiones 0.1 de ambos documentos |
| 2026-08-05 | `CT-001` | Trasladada al registro canónico; permanece abierta | Revisión del archivo de respuestas manuales |
| 2026-08-05 | `CT-001` | Resuelta: recordatorios automáticos elevados a P0 | `DEC-027`, `DEC-032` |
| 2026-08-11 | `CT-002` | Detectada y resuelta en la misma sesión: `event_type` adopta inglés `snake_case` | `DEC-041` |
| 2026-08-13 | `CT-003` | Detectada y registrada como abierta: ubicación del login frente a protección total de `/api/v1/private` | `HU-005`, `CA-006-04`, `backend-go.md` §§6–7 |
| 2026-08-13 | `CT-004` | Detectada y registrada como abierta: destino al panel de `HU-010` frente a dependencia de `HU-012` | `CA-010-01`, dependencias de `HU-010`/`HU-012` |
| 2026-08-13 | `CT-003` | Resuelta: login se mueve a `/api/v1/public/auth/login` | `DEC-055` |
| 2026-08-13 | `CT-004` | Resuelta: `CA-010-01` dividido entre `HU-010` (destino `/panel` mínimo) y `HU-012` (cascarón completo) | `DEC-056` |
| 2026-08-14 | `CT-005` | Detectada y registrada como abierta: el SQL escala al alcanzar 5, mientras decisión/HU dicen superar 5 | `HU-007`, issue `#57`, `DP-SEG-10` |
| 2026-08-14 | `CT-006` | Detectada y registrada como abierta: respuesta de recuperación idéntica frente a destino real enmascarado | `HU-008`, `HU-011`, issue `#58` |
| 2026-08-17 | `CT-005` | Resuelta: la sexta solicitud exige el reto, no la quinta | `DEC-061` |
| 2026-08-17 | `CT-006` | Resuelta: destino enmascarado solo tras verificar el código | `DEC-065` |
| 2026-08-23 | `CT-007` | Detectada y resuelta en la misma entrada: mensaje uniforme para código incorrecto/vencido/agotado, no distinto | `HU-011`, issue `#64`, `DEC-064`, `DEC-065` |
| 2026-08-25 | `CT-008` | Detectada y registrada como abierta: las FK de B2 proponen `ON DELETE CASCADE`, contrario a la prohibición de borrado en cascada | `AGENTS.md`, `database/modelo-fisico-referencia.sql`, preparación de `HU-040`–`HU-042` |
| 2026-08-25 | `CT-008` | Resuelta: las siete FK pasan a `ON DELETE RESTRICT` | `DEC-070` |
