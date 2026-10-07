# Casos transversales

IDs UI son operativos, no CA normativos. N/A exige una razón verificable.

| ID | Acción desde UI | Resultado esperado |
| --- | --- | --- |
| UI-T-01 | Guardar válido → F5 → abrir edición | Valores normalizados recuperados, no solo aviso |
| UI-T-02 | Comparar lista con selector de otra pantalla | Dato coherente, RN-SER-04 |
| UI-T-03 | Doble clic durante envío | Una intención/efecto, RN-IDE-01 |
| UI-T-04 | Atrás/adelante y URL privada directa | Datos coherentes y guardia de CA-006-04 |
| UI-T-05 | Offline antes de guardar → reconectar/reintentar | Datos no sensibles conservados y un efecto |
| UI-T-06 | Escritura real completa, perder respuesta → reintentar | Idempotencia, sin duplicados |
| UI-T-07 | Consulta lenta real, cambiar selección | Respuesta antigua no pisa contexto nuevo |
| UI-T-08 | Dos pestañas, versiones incompatibles | Conflicto/política normativa; no inventar last-write-wins |
| UI-T-09 | Logout en pestaña A, guardar en B | Rechazo/reauth sin éxito falso, CA-006-02 |
| UI-T-10 | Expirar en servidor sesión QA mientras formulario abierto | 401 y no escritura; apoyo del coordinador |
| UI-T-11 | Vacío/espacios/mínimo/máximo/máximo+1 | Validación por campo, sin 500 ni escritura parcial |
| UI-T-12 | Unicode/emoji/HTML literal en campo permitido | Íntegro tras F5 y HTML no ejecutado |
| UI-T-13 | 320/360/768/1280, resize con diálogo | Controles alcanzables, sin desbordes; medir innerWidth |
| UI-T-14 | Tab/Shift+Tab/Enter/Escape y zoom 200% | Foco, etiquetas y diálogo accesibles |
| UI-T-15 | Carga/vacío/error/conflicto/éxito/deshabilitado | Estados distinguibles y reintento cuando aplica |
| UI-T-16 | ID ajeno desde URL/selector, sesión propia | Sin fuga, 404 según CA-003-03/RN-TEN-01 |
| UI-T-17 | Consola y red durante recorrido | Sin excepciones ni 5xx inesperados |

UI-T-10: el coordinador usa rol autorizado, sin grants ni cambios globales de reloj. Limitar UPDATE a la sesión concreta (id o hash mantenido en memoria) y staff_user_id del scope, con revoked_at IS NULL. Fijar expires_at=issued_at+intervalo de 1 microsegundo, verificando que quede antes de la hora actual; el CHECK exige expires_at > issued_at, por eso no usar igualdad ni un valor anterior a issued_at. No tocar otras sesiones ni filas ajenas. Falta permiso = BLOCKED; simular 401 no acredita expiración real.

route.abort/retardo/proxy inyectan errores; route.fulfill solo acredita manejo visual SIMULATED. Nunca acredita persistencia, seguridad o negocio. axe-core en browser complementa teclado/zoom/contraste real: no declarar accesibilidad completa desde screenshots o jsdom.
