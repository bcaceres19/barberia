# Disponibilidad pública

Scope **public-slots**, ruta **/reservar/:slug/servicios/:serviceId/barbero/:barberId/horario**. Cuenta: entrada scope=public-slots de apps/web/.auth/nava-qa/current.json, leída privadamente; nunca credenciales del chat o suites ajenas.

Fuentes: HU-094,HU-095 en [historias](../../../02-requisitos/historias-usuario.md), DEC-083,DEC-111 en [decisiones](../../../00-control/registro-decisiones.md), RN en [reglas](../../../01-producto/reglas-negocio.md). Leer fragmentos de la capacidad; norma prevalece sobre código. Duda = registrar y BLOCKED.

Leer [protocolo](../02-protocolo-agentes.md), [transversal](../04-transversal.md) e [informe](../06-informe.md). Preparar sus estados mediante UI en su tenant; fechas relativas a servidor/America/Bogota. Rutas parametrizadas se obtienen navegando con su slug/IDs, no IDs ajenos.

| Caso | Acción | Resultado esperado |
| --- | --- | --- |
| UI-PUBLIC-SLOTS-01 | Fecha válida/profesional/servicio | Solo huecos donde cabe servicio completo. |
| UI-PUBLIC-SLOTS-02 | Antes/en/después de anticipo/ventana con reloj real | Fronteras servidor RN-DIS-04; apoyo ausente BLOCKED. |
| UI-PUBLIC-SLOTS-03 | Cambiar duración/jornada/excepción/bloqueo propios | Recalcula RN-DIS-02. |
| UI-PUBLIC-SLOTS-04 | Crear contiguo/cruzado en panel propio | Sin cruces; contiguos válidos si cabe. |
| UI-PUBLIC-SLOTS-05 | Cambio de fecha/profesional con consulta lenta | Respuesta vieja no pisa selección CA-095-03. |
| UI-PUBLIC-SLOTS-06 | Vacío/offline/reintento | No afirma reservar al consultar, conserva selección válida. |
| UI-PUBLIC-SLOTS-07 | Timezone del dispositivo diferente | Mismo turno en zona barbería CA-095-02. |

Aplicar UI-T donde corresponda, 320/360/768/1280, teclado y zoom 200%. Todos los casos tienen estado/evidencia. Guardado tras F5; apoyo no disponible BLOCKED, no PASS por leer código. No modificar producto ni cuentas de otro scope.
