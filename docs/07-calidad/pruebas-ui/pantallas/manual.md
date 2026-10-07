# Nuevo turno manual

Scope **manual**, ruta **/panel/turnos/nuevo**. Cuenta: entrada scope=manual de apps/web/.auth/nava-qa/current.json, leída privadamente; nunca credenciales del chat o suites ajenas.

Fuentes: HU-061 en [historias](../../../02-requisitos/historias-usuario.md), DEC-071,DEC-072,DEC-073 en [decisiones](../../../00-control/registro-decisiones.md), RN en [reglas](../../../01-producto/reglas-negocio.md). Leer fragmentos de la capacidad; norma prevalece sobre código. Duda = registrar y BLOCKED.

Leer [protocolo](../02-protocolo-agentes.md), [transversal](../04-transversal.md) e [informe](../06-informe.md). Preparar sus estados mediante UI en su tenant; fechas relativas a servidor/America/Bogota. Rutas parametrizadas se obtienen navegando con su slug/IDs, no IDs ajenos.

| Caso | Acción | Resultado esperado |
| --- | --- | --- |
| UI-MANUAL-01 | Profesional+servicio asignado+persona+fecha válida → agenda | Una cita manual, snapshots correctos. |
| UI-MANUAL-02 | Fecha próxima fuera de anticipación/rejilla pública | Permitida si cumple regla manual RN-CIT-02. |
| UI-MANUAL-03 | Servicio inactivo/no asignado | No elegible, servidor revalida. |
| UI-MANUAL-04 | Persona vacía/120/121, contactos opcionales, nota 500/501 | Límites manuales; no heredar obligatoriedad pública. |
| UI-MANUAL-05 | Franja cruzada y contigua | Cruce rechazado, contiguo válido RN-CON-01. |
| UI-MANUAL-06 | Franja con bloqueo vigente | Conflicto duro controlado sin escritura. |
| UI-MANUAL-07 | Doble envío/respuesta perdida/reintento | Una cita, campos no sensibles conservados. |
| UI-MANUAL-08 | Correo repetido sin teléfono y sin ambos contactos | Reconciliación DEC-071; sin ambos persona nueva. |

Aplicar UI-T donde corresponda, 320/360/768/1280, teclado y zoom 200%. Todos los casos tienen estado/evidencia. Guardado tras F5; apoyo no disponible BLOCKED, no PASS por leer código. No modificar producto ni cuentas de otro scope.
