# Agenda diaria

Scope **agenda**, ruta **/panel**. Cuenta: entrada scope=agenda de apps/web/.auth/nava-qa/current.json, leída privadamente; nunca credenciales del chat o suites ajenas.

Fuentes: HU-062,HU-063 en [historias](../../../02-requisitos/historias-usuario.md), DEC-074,DEC-075,DEC-115 en [decisiones](../../../00-control/registro-decisiones.md), RN en [reglas](../../../01-producto/reglas-negocio.md). Leer fragmentos de la capacidad; norma prevalece sobre código. Duda = registrar y BLOCKED.

Leer [protocolo](../02-protocolo-agentes.md), [transversal](../04-transversal.md) e [informe](../06-informe.md). Preparar sus estados mediante UI en su tenant; fechas relativas a servidor/America/Bogota. Rutas parametrizadas se obtienen navegando con su slug/IDs, no IDs ajenos.

| Caso | Acción | Resultado esperado |
| --- | --- | --- |
| UI-AGENDA-01 | Abrir panel en equipo → seleccionar profesional | Selector obligatorio, no agenda consolidada inventada. |
| UI-AGENDA-02 | Día vacío/día con turnos propios | Estados distintos, orden y horas de barbería. |
| UI-AGENDA-03 | Anterior/siguiente/calendario/hoy → F5 | Fecha según HU-063, sin datos de otro día. |
| UI-AGENDA-04 | Cambiar profesional/fecha con respuesta lenta | Respuesta tardía no pisa selección. |
| UI-AGENDA-05 | Turno que cruza medianoche creado por UI | Visible en cada día que intersecta DEC-075. |
| UI-AGENDA-06 | Abrir detalle → volver | Contexto/dato coherentes. |
| UI-AGENDA-07 | Perfil individual y móvil con texto 200%/scroll | Profesional real, dock no tapa controles. |

Aplicar UI-T donde corresponda, 320/360/768/1280, teclado y zoom 200%. Todos los casos tienen estado/evidencia. Guardado tras F5; apoyo no disponible BLOCKED, no PASS por leer código. No modificar producto ni cuentas de otro scope.
