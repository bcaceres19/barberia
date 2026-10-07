# Entrada pública

Scope **public-entry**, ruta **/reservar/:slug**. Cuenta: entrada scope=public-entry de apps/web/.auth/nava-qa/current.json, leída privadamente; nunca credenciales del chat o suites ajenas.

Fuentes: HU-090 en [historias](../../../02-requisitos/historias-usuario.md), DEC-110,DEC-111 en [decisiones](../../../00-control/registro-decisiones.md), RN en [reglas](../../../01-producto/reglas-negocio.md). Leer fragmentos de la capacidad; norma prevalece sobre código. Duda = registrar y BLOCKED.

Leer [protocolo](../02-protocolo-agentes.md), [transversal](../04-transversal.md) e [informe](../06-informe.md). Preparar sus estados mediante UI en su tenant; fechas relativas a servidor/America/Bogota. Rutas parametrizadas se obtienen navegando con su slug/IDs, no IDs ajenos.

| Caso | Acción | Resultado esperado |
| --- | --- | --- |
| UI-PUBLIC-ENTRY-01 | Slug propio sin login | Datos públicos, marca y CTA reales. |
| UI-PUBLIC-ENTRY-02 | Slug inexistente y no publicable preparado | Respuesta uniforme sin existencia revelada. |
| UI-PUBLIC-ENTRY-03 | F5/móvil/entrar catálogo | Sin guardia staff, cascarón persistente. |
| UI-PUBLIC-ENTRY-04 | Error de red y reintento | Mensaje seguro y recuperación. |
| UI-PUBLIC-ENTRY-05 | Cambiar marca/vocabulario propios → F5 | Valores reales, sin dato privado. |

Aplicar UI-T donde corresponda, 320/360/768/1280, teclado y zoom 200%. Todos los casos tienen estado/evidencia. Guardado tras F5; apoyo no disponible BLOCKED, no PASS por leer código. No modificar producto ni cuentas de otro scope.
