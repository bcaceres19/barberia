---
titulo: "Supuestos temporales del proyecto"
version: "1.0"
estado: "Vigente"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-05"
documentos_relacionados:
  - "registro-decisiones.md"
  - "../01-producto/reglas-negocio.md"
  - "../01-producto/alcance-mvp.md"
---

# Supuestos temporales del proyecto

## 1. Regla de uso

Un supuesto permite avanzar mientras se valida una condición del mundo real; no reemplaza una decisión. Las respuestas de [respuesta-dudas-pendientes.txt](../../respuesta-manuales/respuesta-dudas-pendientes.txt) cerraron siete supuestos anteriores. Se conservan sus códigos en la sección 3 para mantener trazabilidad.

## 2. Supuestos activos

### SUP-002 · Las barberías del piloto usan `America/Bogota`

- **Enunciado:** el piloto ocurre en Colombia y la interfaz usa la zona de cada barbería, inicialmente `America/Bogota`.
- **Riesgo si es falso:** bajo para el modelo, alto para una interfaz mal configurada.
- **Mitigación:** todos los instantes conservan zona horaria (`RN-DIS-07`).
- **Validación:** confirmar zona al dar de alta cada barbería.

### SUP-005 · Un servicio ocupa un solo barbero

- **Enunciado:** ningún servicio requiere dos personas o recursos simultáneos.
- **Riesgo si es falso:** el cálculo de disponibilidad tendría que asignar varios recursos.
- **Validación:** revisar los catálogos de los participantes.

### SUP-007 · El cliente reserva desde un celular con navegador

- **Enunciado:** el enlace se abre principalmente desde WhatsApp, correo o redes en un teléfono.
- **Riesgo si es falso:** cambia la prioridad responsive, no el modelo.
- **Validación:** métrica agregada de tipo de dispositivo, sin identificar personas.

### SUP-008 · El barbero usa un celular de gama media

- **Enunciado:** opera de pie, con una mano y un teléfono Android de gama media.
- **Riesgo si es falso:** una interfaz pesada provoca abandono.
- **Validación:** probar el flujo completo en el dispositivo real de al menos un participante.

### SUP-009 · La conexión puede ser lenta e intermitente, pero existe

- **Enunciado:** no se requiere operación totalmente fuera de línea.
- **Riesgo si es falso:** la sincronización posterior puede crear turnos cruzados y cambia la arquitectura.
- **Validación:** comprobar conectividad en cada local antes del piloto.

### SUP-010 · El volumen del piloto es bajo

- **Enunciado:** menos de 1.000 turnos al mes entre los 2 o 3 participantes.
- **Riesgo si es falso:** bajo; PostgreSQL soporta varios órdenes más, pero deben revisarse los planes.
- **Validación:** medir turnos y consultas por día.

### SUP-011 · El propietario es el único desarrollador y operador

- **Enunciado:** una sola persona construye, despliega y presta soporte.
- **Riesgo si es falso:** ninguno; más capacidad permitiría revisar el alcance.
- **Consecuencia:** monolito modular y operación sencilla (`DEC-023`).

### SUP-014 · El horario base es semanal y cambia poco

- **Enunciado:** cada barbero tiene un horario semanal; recurrencias, fechas explícitas y excepciones cubren variaciones.
- **Riesgo si es falso:** la configuración se vuelve laboriosa.
- **Validación:** observar el número de excepciones creadas por barbero durante el piloto.

### SUP-015 · El precio del servicio es informativo

- **Enunciado:** el MVP no cobra anticipos ni integra una pasarela.
- **Riesgo si es falso:** pagos, reembolsos y obligaciones asociadas cambian el producto.
- **Validación:** confirmarlo por escrito con los participantes antes de iniciar.

## 3. Supuestos cerrados

| Código | Estado de cierre | Decisión |
| --- | --- | --- |
| `SUP-001` | Retirado: el piloto puede incluir independientes o barberías con varios barberos | `DEC-019`, `DEC-028` |
| `SUP-003` | Retirado: se permiten turnos que crucen medianoche si caben en el horario | `DEC-020` |
| `SUP-004` | Retirado: el MVP soporta uno o varios barberos | `DEC-019` |
| `SUP-006` | Confirmado como regla: el cliente no crea cuenta y usa enlace aleatorio | `DEC-022` |
| `SUP-012` | Confirmado: soporte < 1 hora la primera semana y mismo día después | `DEC-031` |
| `SUP-013` | Confirmado: respaldo anterior durante la primera semana | `DEC-028` |
| `SUP-016` | Sustituido por prerrequisito: política y aviso antes de tratar datos del piloto | `DEC-030` |

## 4. Prioridad de validación

1. `SUP-009`: conectividad real de cada local.
2. `SUP-008`: rendimiento en el teléfono real.
3. `SUP-015`: ausencia de pagos o anticipos en el MVP.
4. `SUP-014`: frecuencia real de excepciones al horario.
