---
titulo: "Registro de contradicciones"
version: "1.3"
estado: "Vigente"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-13"
documentos_relacionados:
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
