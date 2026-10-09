# Catálogo público

Scope **public-catalog**, ruta **/reservar/:slug/servicios**. Cuenta: entrada scope=public-catalog de apps/web/.auth/nava-qa/current.json, leída privadamente; nunca credenciales del chat o suites ajenas.

Fuentes: HU-091 en [historias](../../../02-requisitos/historias-usuario.md), DEC-067,DEC-111,DEC-114 en [decisiones](../../../00-control/registro-decisiones.md), RN en [reglas](../../../01-producto/reglas-negocio.md). Leer fragmentos de la capacidad; norma prevalece sobre código. Duda = registrar y BLOCKED.

Leer [protocolo](../02-protocolo-agentes.md), [transversal](../04-transversal.md) e [informe](../06-informe.md). Preparar sus estados mediante UI en su tenant; fechas relativas a servidor/America/Bogota. Rutas parametrizadas se obtienen navegando con su slug/IDs, no IDs ajenos.

| Caso | Acción | Resultado esperado |
| --- | --- | --- |
| UI-PUBLIC-CATALOG-01 | Abrir catálogo propio | Solo activos asignados; tercer servicio fixture ausente. |
| UI-PUBLIC-CATALOG-02 | Comparar con panel propio | Nombre/precio/COP/duración coherentes. |
| UI-PUBLIC-CATALOG-03 | Desactivar/retirar última asignación → F5 | Desaparece sin borrar historial. |
| UI-PUBLIC-CATALOG-04 | Elegir servicio → atrás/F5 | Selección coherente, no reserva implícita. |
| UI-PUBLIC-CATALOG-05 | Preparar vacío por UI y error de red | Estados distintos, reintento. |

Aplicar UI-T donde corresponda, 320/360/768/1280, teclado y zoom 200%. Todos los casos tienen estado/evidencia. Guardado tras F5; apoyo no disponible BLOCKED, no PASS por leer código. No modificar producto ni cuentas de otro scope.
