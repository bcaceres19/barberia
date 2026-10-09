# Cliente y confirmación pública

Scope **public-confirm**, ruta **/reservar/:slug/servicios/:serviceId/barbero/:barberId/horario/:startsAt/cliente**. Cuenta: entrada scope=public-confirm de apps/web/.auth/nava-qa/current.json, leída privadamente; nunca credenciales del chat o suites ajenas.

Fuentes: HU-096,HU-097 en [historias](../../../02-requisitos/historias-usuario.md), DEC-045,DEC-046,DEC-091,DEC-111 en [decisiones](../../../00-control/registro-decisiones.md), RN en [reglas](../../../01-producto/reglas-negocio.md). Leer fragmentos de la capacidad; norma prevalece sobre código. Duda = registrar y BLOCKED.

Leer [protocolo](../02-protocolo-agentes.md), [transversal](../04-transversal.md) e [informe](../06-informe.md). Preparar sus estados mediante UI en su tenant; fechas relativas a servidor/America/Bogota. Rutas parametrizadas se obtienen navegando con su slug/IDs, no IDs ajenos.

| Caso | Acción | Resultado esperado |
| --- | --- | --- |
| UI-PUBLIC-CONFIRM-01 | Para sí, datos ficticios válidos → resumen → confirmar | Una public confirmed y nombre atendido derivado. |
| UI-PUBLIC-CONFIRM-02 | Para otra persona, nombre atendido vacío/válido | Obligatorio solo para otro RN-RES-03. |
| UI-PUBLIC-CONFIRM-03 | Nombre120/121, contacto vacío/inválido, nota500/501 | Validación y datos conservados, no PII en URL. |
| UI-PUBLIC-CONFIRM-04 | Volver del resumen y editar | Confirmación usa dato final visible. |
| UI-PUBLIC-CONFIRM-05 | Doble toque/respuesta perdida tras escribir | Una cita, accessToken solo memoria privada. |
| UI-PUBLIC-CONFIRM-06 | Franja ocupada/bloqueada tras consulta → confirmar | 409 y alternativas conservando formulario. |
| UI-PUBLIC-CONFIRM-07 | Confirmar → agenda/detalle propios → F5 | Persistencia y snapshots reales. |
| UI-PUBLIC-CONFIRM-08 | Token emitido → Mi turno sin sesión staff | Solo su cita; no acredita correo externo. |

Aplicar UI-T donde corresponda, 320/360/768/1280, teclado y zoom 200%. Todos los casos tienen estado/evidencia. Guardado tras F5; apoyo no disponible BLOCKED, no PASS por leer código. No modificar producto ni cuentas de otro scope.
