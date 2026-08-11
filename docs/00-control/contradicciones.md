---
titulo: "Registro de contradicciones"
version: "1.0"
estado: "Vigente"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-05"
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

## 4. Historial de estado

| Fecha | Código | Cambio | Evidencia |
| --- | --- | --- | --- |
| 2026-07-30 | `CT-001` | Detectada en alcance y prioridades | Versiones 0.1 de ambos documentos |
| 2026-08-05 | `CT-001` | Trasladada al registro canónico; permanece abierta | Revisión del archivo de respuestas manuales |
| 2026-08-05 | `CT-001` | Resuelta: recordatorios automáticos elevados a P0 | `DEC-027`, `DEC-032` |
