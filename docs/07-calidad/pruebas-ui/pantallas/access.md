# Acceso y sesión

Scope **access**, ruta **/acceso**. Cuenta: entrada scope=access de apps/web/.auth/nava-qa/current.json, leída privadamente; nunca credenciales del chat o suites ajenas.

Fuentes: HU-005,HU-006,HU-007,HU-010,HU-012 en [historias](../../../02-requisitos/historias-usuario.md), DEC-050,DEC-057,DEC-061,DEC-108 en [decisiones](../../../00-control/registro-decisiones.md), RN en [reglas](../../../01-producto/reglas-negocio.md). Leer fragmentos de la capacidad; norma prevalece sobre código. Duda = registrar y BLOCKED.

Leer [protocolo](../02-protocolo-agentes.md), [transversal](../04-transversal.md) e [informe](../06-informe.md). Preparar sus estados mediante UI en su tenant; fechas relativas a servidor/America/Bogota. Rutas parametrizadas se obtienen navegando con su slug/IDs, no IDs ajenos.

| Caso | Acción | Resultado esperado |
| --- | --- | --- |
| UI-ACCESS-01 | Login válido → panel → F5 | Sesión y barbería correctas, agenda visible. |
| UI-ACCESS-02 | Correo desconocido, contraseña incorrecta e inactive, contextos distintos | Rechazo no enumerable; aviso emergente, correo conservado, contraseña limpia. |
| UI-ACCESS-03 | Vacío, correo inválido, 254/255 y contraseña 256/257 | Límites del formulario sin envío inválido ni 500. |
| UI-ACCESS-04 | Doble clic con respuesta real lenta | Una intención/sesión, botón bloquea segundo envío. |
| UI-ACCESS-05 | Offline al iniciar → reintentar | Datos conservados según CA-010-03; segundo intento real. |
| UI-ACCESS-06 | Umbral de intentos y siguiente solicitud, misma IP | Escalamiento exacto DEC-061; no confundir defensa con bug. |
| UI-ACCESS-07 | Reto correcto/incorrecto/usado/reenvío | Código del marcador local, límites y consumo único. |
| UI-ACCESS-08 | UI-S-01 a UI-S-06 de recorridos | Hash, cookie, renovación, expiración y revocación reales o apoyo BLOCKED. |

Aplicar UI-T donde corresponda, 320/360/768/1280, teclado y zoom 200%. Todos los casos tienen estado/evidencia. Guardado tras F5; apoyo no disponible BLOCKED, no PASS por leer código. No modificar producto ni cuentas de otro scope.
