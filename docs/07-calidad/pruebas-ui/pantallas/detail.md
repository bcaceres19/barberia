# Detalle e historial

Scope **detail**, ruta **/panel/turnos/:appointmentId**. Cuenta: entrada scope=detail de apps/web/.auth/nava-qa/current.json, leída privadamente; nunca credenciales del chat o suites ajenas.

Fuentes: HU-064,HU-065,HU-066,HU-067,HU-068 en [historias](../../../02-requisitos/historias-usuario.md), DEC-076 en [decisiones](../../../00-control/registro-decisiones.md), RN en [reglas](../../../01-producto/reglas-negocio.md). Leer fragmentos de la capacidad; norma prevalece sobre código. Duda = registrar y BLOCKED.

Leer [protocolo](../02-protocolo-agentes.md), [transversal](../04-transversal.md) e [informe](../06-informe.md). Preparar sus estados mediante UI en su tenant; fechas relativas a servidor/America/Bogota. Rutas parametrizadas se obtienen navegando con su slug/IDs, no IDs ajenos.

| Caso | Acción | Resultado esperado |
| --- | --- | --- |
| UI-DETAIL-01 | Crear propio → detalle → F5 | Datos/snapshots/cronología reales. |
| UI-DETAIL-02 | Reprogramar válido → Mi turno | Hora/versión nueva y evento anterior/nuevo. |
| UI-DETAIL-03 | Reprogramar a cruce/bloqueo/versión antigua | Conflicto atómico sin historial parcial. |
| UI-DETAIL-04 | Cancelar por barbero según RN-CAN-03 | Estado/motivo/actor correctos. |
| UI-DETAIL-05 | Cerrar attended/no_show según reloj/transición | Solo transiciones HU-067, sin cierre arbitrario. |
| UI-DETAIL-06 | Corregir resultado con motivo | Evento nuevo y anterior intacto RN-CIT-04/RN-HIS-02. |
| UI-DETAIL-07 | Corregir a estado que ocupa franja ya ocupada | Revalida, conflicto atómico HU-068. |
| UI-DETAIL-08 | Doble clic/reintento por acción y dos pestañas | Un efecto/evento y concurrencia controlada. |
| UI-DETAIL-09 | ID inexistente/de isolation-b | No contenido ajeno, 404 donde corresponda. |

Aplicar UI-T donde corresponda, 320/360/768/1280, teclado y zoom 200%. Todos los casos tienen estado/evidencia. Guardado tras F5; apoyo no disponible BLOCKED, no PASS por leer código. No modificar producto ni cuentas de otro scope.
