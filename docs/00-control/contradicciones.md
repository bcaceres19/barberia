---
titulo: "Registro de contradicciones"
version: "1.1"
estado: "Vigente"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-11"
documentos_relacionados:
  - "registro-decisiones.md"
  - "dudas-pendientes.md"
  - "../01-producto/alcance-mvp.md"
  - "../01-producto/prioridades.md"
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

## 4. Historial de estado

| Fecha | Código | Cambio | Evidencia |
| --- | --- | --- | --- |
| 2026-07-30 | `CT-001` | Detectada en alcance y prioridades | Versiones 0.1 de ambos documentos |
| 2026-08-05 | `CT-001` | Trasladada al registro canónico; permanece abierta | Revisión del archivo de respuestas manuales |
| 2026-08-05 | `CT-001` | Resuelta: recordatorios automáticos elevados a P0 | `DEC-027`, `DEC-032` |
| 2026-08-11 | `CT-002` | Detectada y resuelta en la misma sesión: `event_type` adopta inglés `snake_case` | `DEC-041` |
