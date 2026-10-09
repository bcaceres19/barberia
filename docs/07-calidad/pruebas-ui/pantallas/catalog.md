# Servicios y ciclo de vida

Scope **catalog**, ruta **/panel/servicios**. Cuenta: entrada scope=catalog de apps/web/.auth/nava-qa/current.json, leída privadamente; nunca credenciales del chat o suites ajenas.

Fuentes: HU-022,HU-024 en [historias](../../../02-requisitos/historias-usuario.md), DEC-067,DEC-103 en [decisiones](../../../00-control/registro-decisiones.md), RN en [reglas](../../../01-producto/reglas-negocio.md). Leer fragmentos de la capacidad; norma prevalece sobre código. Duda = registrar y BLOCKED.

Leer [protocolo](../02-protocolo-agentes.md), [transversal](../04-transversal.md) e [informe](../06-informe.md). Preparar sus estados mediante UI en su tenant; fechas relativas a servidor/America/Bogota. Rutas parametrizadas se obtienen navegando con su slug/IDs, no IDs ajenos.

| Caso | Acción | Resultado esperado |
| --- | --- | --- |
| UI-CATALOG-01 | Servicio 25 min y precio decimal → F5 | Guardado exacto, sin lista cerrada de duraciones. |
| UI-CATALOG-02 | Nombre vacío/120/121 y descripción 500/501 | Límites correctos; modal/lista legibles. |
| UI-CATALOG-03 | Duración 0/1/1440/1441/30.5, precio 0/positivo/3 decimales/coma | 1–1440 enteros, COP >0 y dos decimales. |
| UI-CATALOG-04 | Nombre activo duplicado y mismo nombre en tenant ajeno | Conflicto solo tenant propio. |
| UI-CATALOG-05 | Crear turno → cambiar precio/nombre/duración | Snapshot del turno previo intacto RN-SER-04. |
| UI-CATALOG-06 | Desactivar con asignaciones/turno futuro → F5 | Advertencia/acción según HU-024; fuera de oferta, historial conservado. |
| UI-CATALOG-07 | Reactivar con nombre activo ya ocupado | Conflicto controlado de unicidad. |
| UI-CATALOG-08 | Buscar y paginar más de una página | Total/filtro coherentes tras mutación. |

Aplicar UI-T donde corresponda, 320/360/768/1280, teclado y zoom 200%. Todos los casos tienen estado/evidencia. Guardado tras F5; apoyo no disponible BLOCKED, no PASS por leer código. No modificar producto ni cuentas de otro scope.
