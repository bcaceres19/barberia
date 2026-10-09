# Servicios por barbero

Scope **assignments**, ruta **/panel/servicios-por-barbero**. Cuenta: entrada scope=assignments de apps/web/.auth/nava-qa/current.json, leída privadamente; nunca credenciales del chat o suites ajenas.

Fuentes: HU-023 en [historias](../../../02-requisitos/historias-usuario.md), DEC-114,DEC-115 en [decisiones](../../../00-control/registro-decisiones.md), RN en [reglas](../../../01-producto/reglas-negocio.md). Leer fragmentos de la capacidad; norma prevalece sobre código. Duda = registrar y BLOCKED.

Leer [protocolo](../02-protocolo-agentes.md), [transversal](../04-transversal.md) e [informe](../06-informe.md). Preparar sus estados mediante UI en su tenant; fechas relativas a servidor/America/Bogota. Rutas parametrizadas se obtienen navegando con su slug/IDs, no IDs ajenos.

| Caso | Acción | Resultado esperado |
| --- | --- | --- |
| UI-ASSIGNMENTS-01 | Elegir barbero → asignar → F5 | Asignación y oferta pública coherentes. |
| UI-ASSIGNMENTS-02 | Retirar asignación no última y última | Ambas permitidas DEC-114; sin última no oferta pública. |
| UI-ASSIGNMENTS-03 | Cambiar barbero durante consulta lenta | Respuesta anterior no pisa nueva selección. |
| UI-ASSIGNMENTS-04 | Servicio inactivo/sin asignación | Sin oferta inválida ni obligación inventada. |
| UI-ASSIGNMENTS-05 | Doble toque y dos pestañas misma asignación | Efecto seguro/errores del contrato, sin 500. |
| UI-ASSIGNMENTS-06 | Cambiar a perfil individual propio → Lo ofrezco | Profesional real, sin equipo ficticio. |

Aplicar UI-T donde corresponda, 320/360/768/1280, teclado y zoom 200%. Todos los casos tienen estado/evidencia. Guardado tras F5; apoyo no disponible BLOCKED, no PASS por leer código. No modificar producto ni cuentas de otro scope.
