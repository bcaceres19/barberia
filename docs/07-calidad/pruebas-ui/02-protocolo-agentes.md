# Protocolo y consumo

El coordinador persiste el issue y cada encargo antes de delegar. Subagente: **GPT-6 Luna / medium, fork_turns=none**, máximo dos activos. Recibe solo prompt guardado, scope, campaña, SHA, URL local, ruta del manifiesto privado y salida exclusiva; nunca credenciales ni historial completo.

El coordinador usa Luna si el cliente lo permite: un skill no cambia el modelo del hilo actual. Otro runtime registra sus limitaciones; no inventar un identificador ni afirmar que se usó Luna. No sustituir por Astra/Sol automáticamente. Registrar el modelo/esfuerzo desde los parámetros documentados de la delegación; la identidad genérica del mensaje de sistema no acredita una variante distinta.

## Por tarea

1. Leer protocolo, transversal, informe y SOLO su ficha. Consultar los fragmentos HU/RN/DEC que resuelvan el esperado; no cargar todo el repositorio ni Graphify.
2. Validar cuenta/scope, lanzar navegador/contexto propio y probar API/datos reales. Inspeccionar nombres accesibles actuales: iconos y asteriscos pueden formar parte del nombre; preferir roles y regex acotadas cuando corresponda, sin confundir un selector incorrecto con fallo de producto.
3. Ejecutar casos de ficha y transversales aplicables: estados, 320/360/768/1280, teclado y zoom 200%. Cada caso recibe PASS/FAIL/BLOCKED/NOT_RUN/N/A con razón y evidencia.
4. Crear/cambiar solo datos del tenant propio, prefijo QA. Fechas relativas a reloj servidor y America/Bogota.
5. Ante fallo conservar evidencia inmediatamente, reproducir una vez en contexto limpio si es seguro. No corregir producto, migraciones, contratos, snapshots ni pruebas.
6. Escribir resultados continuamente en salida exclusiva; cerrar solo su navegador. No detener servidores compartidos.

Tanda orientativa: hasta 12 casos o 20 minutos. Los restantes quedan NOT_RUN y se reencolan, nunca se omiten como si pasaran. Máximo dos intentos ante bloqueo de infraestructura; informar al coordinador. No volcar DOM/HAR/JSON ni repetir capturas enteras al modelo. Captura por viewport y por fallo relevante. Resumen <=400 palabras con enlace a salida detallada.

## Orden y estados

Smoke; access y recovery en serie, sin otros agentes autenticándose durante abuso. Después cola por pantallas con tenants propios, máximo dos concurrentes. Al final journeys y aislamiento en serie. Marca, perfil, políticas, bloqueos y contraseñas solo cambian su tenant/cuenta.

Solo el coordinador actualiza prompts e índice: ready → in_progress → executed al ejecutar la ficha y documentar resultados, o blocked por precondición. executed no significa todos PASS. Cambio material de cuerpo ejecutado genera nueva versión y supersedes. Un fix o prueba nueva permanente requiere issue/rama/PR independiente.

Los subagentes escriben solo apps/web/test-results/ui-qa/<campaña>/<scope>/. No modifican repositorio de producto ni estados compartidos del catálogo. Los casos sin resultado normativo se registran como duda y BLOCKED.

## Campañas posteriores y relevo

El coordinador consulta el issue de la campaña con GitHub (herramienta autenticada o gh issue view). Si está cerrado, crea un issue test con los mismos criterios de cobertura adaptados a las rutas actuales y enlaza el anterior; persiste nuevos prompts con ese issue, identificadores nuevos, dependencias y supersedes. Actualiza índice antes de delegar. La petición de probar autoriza esa preparación; no inventar números ni ejecutar prompts asociados al issue cerrado.

Si GitHub no responde, diagnosticar conectividad/permisos hasta dos intentos sin volcar credenciales. Continuar inventario y preflight local, pero registrar BLOCKED para el inicio de campaña sin issue verificable. Una interrupción de infraestructura no se clasifica como fallo de producto.

Cada tarea mantiene checkpoint.json privado en su salida: campaign, scope, SHA, prompt_id/version, modelo/esfuerzo documentado por el orquestador, casos terminados, caso en curso, identificadores de datos QA creados, último paso confirmado y pendientes. Guardar tras cada mutación comprobada. No incluir contraseñas, cookies, códigos ni token/URL de Mi turno.

Tras interrupción del runtime, leer checkpoint e informe, confirmar mismo SHA/campaña/tenant y autenticar de nuevo si hace falta. Antes de repetir una escritura, recuperar por UI el estado del dato registrado; si el resultado anterior es incierto, investigarlo sin repetir una creación a ciegas. Caso parcialmente ejecutado queda NOT_RUN o BLOCKED hasta aportar evidencia. Si cambió SHA o se modificó el encargo, crear una nueva versión/relevo persistente; no reescribir el cuerpo usado.
