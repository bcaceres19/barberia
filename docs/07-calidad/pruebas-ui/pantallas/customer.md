# Mi turno y cancelación pública

Scope **customer**, ruta **/mi-turno/:token**. Cuenta: entrada scope=customer de apps/web/.auth/nava-qa/current.json, leída privadamente; nunca credenciales del chat o suites ajenas.

Fuentes: HU-098,HU-099 en [historias](../../../02-requisitos/historias-usuario.md), DEC-022,DEC-083,DEC-091 en [decisiones](../../../00-control/registro-decisiones.md), RN en [reglas](../../../01-producto/reglas-negocio.md). Leer fragmentos de la capacidad; norma prevalece sobre código. Duda = registrar y BLOCKED.

Leer [protocolo](../02-protocolo-agentes.md), [transversal](../04-transversal.md) e [informe](../06-informe.md). Preparar sus estados mediante UI en su tenant; fechas relativas a servidor/America/Bogota. Rutas parametrizadas se obtienen navegando con su slug/IDs, no IDs ajenos.

| Caso | Acción | Resultado esperado |
| --- | --- | --- |
| UI-CUSTOMER-01 | Reserva real → accessToken → abrir Mi turno | Proyección mínima de una cita, sin staff. |
| UI-CUSTOMER-02 | Reprogramación/estado panel propio → F5 | Fecha/estado/política actuales; no historial técnico. |
| UI-CUSTOMER-03 | Token inválido/revocado/vencido preparado | Uniforme, sin imprimir token; apoyo ausente BLOCKED. |
| UI-CUSTOMER-04 | Cancelar antes/en/después del plazo y políticas propias | Reloj servidor y decisión normativa. |
| UI-CUSTOMER-05 | Motivo vacío/válido cuando requerido | Vacío no cambia, válido cancela una vez. |
| UI-CUSTOMER-06 | Cancelar → F5 → disponibilidad | Terminal e historial únicos, hueco futuro recuperado. |
| UI-CUSTOMER-07 | Dos clics/dos pestañas cancelando | Un efecto; terminal/versión obsoleta conforme contrato. |
| UI-CUSTOMER-08 | Credencial de otra cita/contextos distintos | Token solo autoriza su cita, sin enumeración/fuga. |

Aplicar UI-T donde corresponda, 320/360/768/1280, teclado y zoom 200%. Todos los casos tienen estado/evidencia. Guardado tras F5; apoyo no disponible BLOCKED, no PASS por leer código. No modificar producto ni cuentas de otro scope.
