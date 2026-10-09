# Selección pública de profesional

Scope **public-barber**, ruta **/reservar/:slug/servicios/:serviceId/barbero**. Cuenta: entrada scope=public-barber de apps/web/.auth/nava-qa/current.json, leída privadamente; nunca credenciales del chat o suites ajenas.

Fuentes: HU-092 en [historias](../../../02-requisitos/historias-usuario.md), DEC-104,DEC-111 en [decisiones](../../../00-control/registro-decisiones.md), RN en [reglas](../../../01-producto/reglas-negocio.md). Leer fragmentos de la capacidad; norma prevalece sobre código. Duda = registrar y BLOCKED.

Leer [protocolo](../02-protocolo-agentes.md), [transversal](../04-transversal.md) e [informe](../06-informe.md). Preparar sus estados mediante UI en su tenant; fechas relativas a servidor/America/Bogota. Rutas parametrizadas se obtienen navegando con su slug/IDs, no IDs ajenos.

| Caso | Acción | Resultado esperado |
| --- | --- | --- |
| UI-PUBLIC-BARBER-01 | Elegir servicio | Solo profesionales asignados propios; dos iniciales. |
| UI-PUBLIC-BARBER-02 | Asignar/desasignar en su panel → F5 | Lista actualizada. |
| UI-PUBLIC-BARBER-03 | Foto/monograma/nombre largo sintéticos | Accesible y sin deformación. |
| UI-PUBLIC-BARBER-04 | Servicio ajeno/inactivo/sin profesionales | Sin fuga, estado/error del contrato. |
| UI-PUBLIC-BARBER-05 | Seleccionar → avanzar → atrás/cambiar servicio | Invalida selecciones incompatibles. |

Aplicar UI-T donde corresponda, 320/360/768/1280, teclado y zoom 200%. Todos los casos tienen estado/evidencia. Guardado tras F5; apoyo no disponible BLOCKED, no PASS por leer código. No modificar producto ni cuentas de otro scope.
