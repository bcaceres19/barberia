# Recuperación

Scope **recovery**, ruta **/recuperar-acceso**. Cuenta: entrada scope=recovery de apps/web/.auth/nava-qa/current.json, leída privadamente; nunca credenciales del chat o suites ajenas.

Fuentes: HU-008,HU-011 en [historias](../../../02-requisitos/historias-usuario.md), DEC-064,DEC-065,DEC-092,DEC-094 en [decisiones](../../../00-control/registro-decisiones.md), RN en [reglas](../../../01-producto/reglas-negocio.md). Leer fragmentos de la capacidad; norma prevalece sobre código. Duda = registrar y BLOCKED.

Leer [protocolo](../02-protocolo-agentes.md), [transversal](../04-transversal.md) e [informe](../06-informe.md). Preparar sus estados mediante UI en su tenant; fechas relativas a servidor/America/Bogota. Rutas parametrizadas se obtienen navegando con su slug/IDs, no IDs ajenos.

| Caso | Acción | Resultado esperado |
| --- | --- | --- |
| UI-RECOVERY-01 | Solicitar correo propio y desconocido | Respuesta pública indistinguible; código solo para cuenta válida. |
| UI-RECOVERY-02 | Solicitar WhatsApp de su cuenta ficticia | Marcador local, destino proviene de cuenta autorizada. |
| UI-RECOVERY-03 | Código correcto, incorrecto, usado y vencido preparado | Solo válido autoriza siguiente paso; apoyo temporal ausente BLOCKED. |
| UI-RECOVERY-04 | Intentos máximos y reenvío dentro del cooldown | Límites de abuso y error uniforme, no 500. |
| UI-RECOVERY-05 | Contraseña nueva inválida, confirmación distinta y válida | Política de recuperación aplicada antes de cambiar. |
| UI-RECOVERY-06 | Dos sesiones abiertas → reset → intentar ambas | Ambas revocadas; login nuevo válido, anterior inválido CA-008-05. |
| UI-RECOVERY-07 | Atrás, F5, doble clic por paso | No reutiliza autorización consumida; actualizar contraseña privada en manifiesto. |

Aplicar UI-T donde corresponda, 320/360/768/1280, teclado y zoom 200%. Todos los casos tienen estado/evidencia. Guardado tras F5; apoyo no disponible BLOCKED, no PASS por leer código. No modificar producto ni cuentas de otro scope.
