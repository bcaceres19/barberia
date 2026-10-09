# Política pública

Scope **policy**, ruta **/panel/reserva-publica**. Cuenta: entrada scope=policy de apps/web/.auth/nava-qa/current.json, leída privadamente; nunca credenciales del chat o suites ajenas.

Fuentes: HU-093 en [historias](../../../02-requisitos/historias-usuario.md), DEC-083 en [decisiones](../../../00-control/registro-decisiones.md), RN en [reglas](../../../01-producto/reglas-negocio.md). Leer fragmentos de la capacidad; norma prevalece sobre código. Duda = registrar y BLOCKED.

Leer [protocolo](../02-protocolo-agentes.md), [transversal](../04-transversal.md) e [informe](../06-informe.md). Preparar sus estados mediante UI en su tenant; fechas relativas a servidor/America/Bogota. Rutas parametrizadas se obtienen navegando con su slug/IDs, no IDs ajenos.

| Caso | Acción | Resultado esperado |
| --- | --- | --- |
| UI-POLICY-01 | Leer iniciales → editar válido → F5 | Defaults 60/3/15/20 y tardía+motivo, antes de edición. |
| UI-POLICY-02 | Anticipación 0/1440, ventana 1/90 y valores fuera | Enteros/rangos; anticipo menor a ventana total. |
| UI-POLICY-03 | Cada rejilla permitida y una inválida | 5/10/15/20/30/60 persistidas; otra rechazada. |
| UI-POLICY-04 | Cancelación 0/10080 y fuera | Validación, sin escritura parcial. |
| UI-POLICY-05 | Permitir/rechazar tardía y motivo requerido | Coherencia de flags según contrato y UI. |
| UI-POLICY-06 | Dos pestañas actualizan versión antigua | Conflicto optimista controlado. |
| UI-POLICY-07 | Cambiar política → disponibilidad/Mi turno propios | Aplican nuevos valores reales y reloj servidor. |

Aplicar UI-T donde corresponda, 320/360/768/1280, teclado y zoom 200%. Todos los casos tienen estado/evidencia. Guardado tras F5; apoyo no disponible BLOCKED, no PASS por leer código. No modificar producto ni cuentas de otro scope.
