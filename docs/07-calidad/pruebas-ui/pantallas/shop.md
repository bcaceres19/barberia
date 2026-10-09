# Barbería, marca y perfil

Scope **shop**, ruta **/panel/barberia**. Cuenta: entrada scope=shop de apps/web/.auth/nava-qa/current.json, leída privadamente; nunca credenciales del chat o suites ajenas.

Fuentes: HU-020,HU-025,HU-012 en [historias](../../../02-requisitos/historias-usuario.md), DEC-110,DEC-115 en [decisiones](../../../00-control/registro-decisiones.md), RN en [reglas](../../../01-producto/reglas-negocio.md). Leer fragmentos de la capacidad; norma prevalece sobre código. Duda = registrar y BLOCKED.

Leer [protocolo](../02-protocolo-agentes.md), [transversal](../04-transversal.md) e [informe](../06-informe.md). Preparar sus estados mediante UI en su tenant; fechas relativas a servidor/America/Bogota. Rutas parametrizadas se obtienen navegando con su slug/IDs, no IDs ajenos.

| Caso | Acción | Resultado esperado |
| --- | --- | --- |
| UI-SHOP-01 | Nombre/contactos válidos → guardar → F5 | Valor persistido y coherente con cabecera. |
| UI-SHOP-02 | Contactos vacíos/espacios/inválidos/límites | Normalización/null según contrato; no escritura parcial. |
| UI-SHOP-03 | Zona IANA válida e inválida | Validación del servidor, sin fechas desplazadas silenciosamente. |
| UI-SHOP-04 | Acento/vocabulario/apariencia → navegar/F5 | Marca persistida por barbería; apariencia local según DEC-110. |
| UI-SHOP-05 | Perfil equipo → individual → equipo | Profesional real DEC-115, sin duplicar ni inventar. |
| UI-SHOP-06 | Dos pestañas guardan cambios incompatibles | Marca/perfil siguen última escritura DEC-110/115; no exigir 409 sin versión contractual. Contrastar identidad/contactos con HU-020. |
| UI-SHOP-07 | Offline/doble clic/reintento | Un cambio real y datos no sensibles conservados. |

Aplicar UI-T donde corresponda, 320/360/768/1280, teclado y zoom 200%. Todos los casos tienen estado/evidencia. Guardado tras F5; apoyo no disponible BLOCKED, no PASS por leer código. No modificar producto ni cuentas de otro scope.
