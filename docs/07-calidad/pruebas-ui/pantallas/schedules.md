# Horarios, excepciones y festivos

Scope **schedules**, ruta **/panel/horarios**. Cuenta: entrada scope=schedules de apps/web/.auth/nava-qa/current.json, leída privadamente; nunca credenciales del chat o suites ajenas.

Fuentes: HU-040,HU-041 en [historias](../../../02-requisitos/historias-usuario.md), DEC-109,DEC-115 en [decisiones](../../../00-control/registro-decisiones.md), RN en [reglas](../../../01-producto/reglas-negocio.md). Leer fragmentos de la capacidad; norma prevalece sobre código. Duda = registrar y BLOCKED.

Leer [protocolo](../02-protocolo-agentes.md), [transversal](../04-transversal.md) e [informe](../06-informe.md). Preparar sus estados mediante UI en su tenant; fechas relativas a servidor/America/Bogota. Rutas parametrizadas se obtienen navegando con su slug/IDs, no IDs ajenos.

| Caso | Acción | Resultado esperado |
| --- | --- | --- |
| UI-SCHEDULES-01 | Elegir profesional → editar semana → F5 | Día/tramos persistidos en tenant propio. |
| UI-SCHEDULES-02 | Tramos contiguos/cruzados/invertidos/medianoche | Semántica RN-DIS-05 y contrato, sin 500. |
| UI-SCHEDULES-03 | Excepción de fecha: tramos/cerrada → público | Prioridad de excepción sobre semana. |
| UI-SCHEDULES-04 | Dos excepciones misma fecha | Rechazo/actualización del contrato sin duplicado. |
| UI-SCHEDULES-05 | Festivos y excepción para trabajar | RN-BLQ-02 y prioridad de HU-041. |
| UI-SCHEDULES-06 | Cambiar profesional con consulta lenta/dos pestañas | Contexto correcto y concurrencia normativa. |
| UI-SCHEDULES-07 | Tablero móvil/teclado/zoom | DEC-109; no exigir atlas #196 como referencia exacta. |

Aplicar UI-T donde corresponda, 320/360/768/1280, teclado y zoom 200%. Todos los casos tienen estado/evidencia. Guardado tras F5; apoyo no disponible BLOCKED, no PASS por leer código. No modificar producto ni cuentas de otro scope.
