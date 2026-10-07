# Barberos, fotos y bloqueos

Scope **staff**, ruta **/panel/barberos**. Cuenta: entrada scope=staff de apps/web/.auth/nava-qa/current.json, leída privadamente; nunca credenciales del chat o suites ajenas.

Fuentes: HU-021,HU-042 en [historias](../../../02-requisitos/historias-usuario.md), DEC-047,DEC-104,DEC-105,DEC-107 en [decisiones](../../../00-control/registro-decisiones.md), RN en [reglas](../../../01-producto/reglas-negocio.md). Leer fragmentos de la capacidad; norma prevalece sobre código. Duda = registrar y BLOCKED.

Leer [protocolo](../02-protocolo-agentes.md), [transversal](../04-transversal.md) e [informe](../06-informe.md). Preparar sus estados mediante UI en su tenant; fechas relativas a servidor/America/Bogota. Rutas parametrizadas se obtienen navegando con su slug/IDs, no IDs ajenos.

| Caso | Acción | Resultado esperado |
| --- | --- | --- |
| UI-STAFF-01 | Crear/renombrar → F5 → selector en Horarios | Nombre íntegro y tenant correcto. |
| UI-STAFF-02 | Vacío/espacios/120/121/duplicado/unicode | 120/duplicado válidos; vacío y 121 rechazados. |
| UI-STAFF-03 | Crear suficientes filas → paginar/buscar/resize | Paginador legible, sin omisiones/duplicados. |
| UI-STAFF-04 | Foto sintética válida → cambiar/quitar → F5 | Persistencia, sin fotos reales. |
| UI-STAFF-05 | Archivo inválido/excesivo/error de red | Límites del contrato; foto anterior intacta. |
| UI-STAFF-06 | Bloqueo rango/recurrente y excepción desde fila | Solo profesional propio, persistencia tras recarga. |
| UI-STAFF-07 | Retirar bloqueo → disponibilidad pública propia | Retirada lógica, hueco válido recuperado. |
| UI-STAFF-08 | Bloquear sobre turno confirmado | Afectación visible, turno no cancelado RN-BLQ-03. |
| UI-STAFF-09 | Abrir /panel/bloqueos y teclado en diálogos | Redirección a Barberos y foco correcto. |

Aplicar UI-T donde corresponda, 320/360/768/1280, teclado y zoom 200%. Todos los casos tienen estado/evidencia. Guardado tras F5; apoyo no disponible BLOCKED, no PASS por leer código. No modificar producto ni cuentas de otro scope.
