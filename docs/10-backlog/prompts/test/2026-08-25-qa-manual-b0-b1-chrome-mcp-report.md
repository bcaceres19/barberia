---
report_id: "REPORT-TEST-QA-B0-B1-CHROME-2026-08-25"
version: "1.0"
kind: "test-report"
status: "executed"
prompt_id: "PROMPT-TEST-QA-B0-B1-CHROME-v1"
prompt_version: "1.0"
target_agent: "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
related_hu:
  - "HU-005"
  - "HU-006"
  - "HU-007"
  - "HU-008"
  - "HU-010"
  - "HU-011"
  - "HU-012"
  - "HU-020"
  - "HU-021"
  - "HU-022"
  - "HU-023"
  - "HU-024"
execution_date: "2026-08-25"
environment:
  browser: "Codex in-app browser"
  frontend: "http://localhost:5173"
  api: "http://localhost:8080"
  database: "PostgreSQL local barberia_test, migraciones Atlas aplicadas"
  source_revision: "main con PR #84 fusionado"
evidence_limits:
  - "Las capturas se emitieron durante la sesión de navegador; los resultados textuales y observables quedan consolidados en este informe."
  - "El navegador integrado no expone una consola DevTools/Network utilizable para los fetch cross-tenant."
  - "No se inventaron códigos OTP: se ejecutaron únicamente las ramas de error cuando no hubo código real."
---

# Informe de QA manual en navegador real: B0 y B1

## Resultado ejecutivo

Se recorrieron los 51 pasos del prompt asociado. Los recorridos funcionales de B1 (HU-020 a HU-024) se comportaron correctamente en la interfaz, incluida la persistencia, el aislamiento visible entre tenants y el ciclo de desactivar/reactivar servicios.

Quedaron como no verificables o con divergencia:

- No hubo OTP real: el usuario A no tiene teléfono verificado, no había desafíos activos en la base y no existían archivos de captura.
- No se pudo ejecutar el fetch cross-tenant desde DevTools con el navegador integrado.
- La API de control de teclado no consiguió trasladar el foco con Tab desde body ni enviar Enter al formulario.
- A 320 y 360 px, la navegación privada mantiene un scroll horizontal interno. El documento no desborda, pero el criterio literal del prompt exige no tener scroll horizontal.
- La documentación de control todavía describe PR #84 como pendiente, aunque ya está fusionado; se conserva como hallazgo documental y no se corrigió en esta ejecución.

## Estado por paso

| Paso | Bloque/HU | Resultado esperado | Resultado observado | Evidencia |
| --- | --- | --- | --- | --- |
| 1 | B0 / HU-010, HU-012 | /panel redirige a acceso con redirect | Redirigió a /acceso?redirect=/panel | URL observada en navegador |
| 2 | B0 / HU-010 | Encabezado, correo, contraseña y botón visibles | Se observaron Inicia sesión, Correo, Contraseña e Iniciar sesión | Snapshot DOM |
| 3 | B0 / HU-010 | Error contiene revisión de correo y contraseña | Alertas No pudimos iniciar tu sesión y Revisa tu correo y contraseña e inténtalo de nuevo. | Texto literal de alerta |
| 4 | B0 / HU-010 | Login A termina en /panel | Terminó en /panel con Panel del barbero | URL y snapshot DOM |
| 5 | B0 / HU-010 | Cabecera muestra barbería no vacía | Mostró la barbería A; durante HU-020 se cambió a un nombre sintético de QA y persistió | Banner y consulta de persistencia |
| 6 | B0 / HU-006, HU-012 | Recarga conserva sesión | Recarga conservó /panel y la sesión | URL y snapshot DOM |
| 7 | B0 / HU-006, HU-012 | Reapertura del mismo contexto conserva sesión | Nueva pestaña del mismo contexto entró a /panel sin pedir credenciales | URL y banner |
| 8 | B0 / HU-006, HU-012 | Logout y guard revocan acceso | Cerrar sesión llevó a /acceso; volver a /panel redirigió a acceso | URL observada |
| 9 | B0 / HU-007 | Cinco intentos incorrectos mantienen error genérico | Los cinco mostraron el error genérico esperado | Texto literal repetido |
| 10 | B0 / HU-007 | Sexto intento muestra bloqueo y reto telefónico | Mostró Demasiados intentos y Verifica tu teléfono; también espera aproximada de 1440 minutos | Alertas literales |
| 11 | B0 / HU-007 | Envío muestra mensaje no enumerativo | Mostró Si tu cuenta existe y tu teléfono está verificado, recibirás un código por WhatsApp. | Status literal |
| 12 | B0 / HU-007 | Código incorrecto muestra rechazo uniforme | Mostró El código no es válido o venció. Inténtalo de nuevo. | Alert literal |
| 13 | B0 / HU-007 | Código real inicia sesión | No verificable: no existía código real disponible | BD: teléfono no verificado y 0 desafíos activos; archivo de captura ausente |
| 14 | B0 / HU-008, HU-011 | Solicitud avanza a Paso 2 de 3 | Avanzó a Paso 2 de 3 y mostró mensaje no enumerativo | Snapshot y status |
| 15 | B0 / HU-008, HU-011 | Código real avanza a Paso 3 de 3 | No verificable: no existía código real de recuperación | BD: 0 códigos pendientes; archivo de captura ausente |
| 16 | B0 / HU-008, HU-011 | Cambio de contraseña revoca sesiones | No ejecutado por falta del código real y por requerir handoff para el envío final de cambio | Bloqueo documentado |
| 17 | B0 / HU-008, HU-011 | Login con nueva contraseña llega a /panel | No verificable porque el paso 16 no podía completarse | Dependencia del paso 16 |
| 18 | B0 / HU-008, HU-011 | Código inválido mantiene Paso 2 y no enumera causa | Permaneció en Paso 2; alertó El código no es válido y texto El código no es correcto o ya venció. Puedes reenviarlo o revisar lo que escribiste. | Alert y snapshot |
| 19 | B0 / tenants | Contexto aparte muestra barbería B distinta | Se inició sesión secuencialmente con B y mostró Barbería de prueba B; el IAB comparte cookies entre pestañas, por lo que no se pudo probar un perfil aislado real | Banner B y limitación del IAB |
| 20 | B0 / tenants | Fetch de recurso B desde A responde 404 | No verificable desde DevTools del navegador integrado | Sin consola Network disponible |
| 21 | B0 / RFC 9457 | Network devuelve problem+json con type, title, status y detail | Verificado suplementariamente contra API: HTTP 401, Content-Type application/problem+json, type /api/v1/problems/unauthorized, title No autorizado, status 401, detail correo o contraseña incorrectos | Respuesta HTTP directa |
| 22 | B0 / responsive | 320, 360, 768 y 1280 sin scroll horizontal ni solapamientos | acceso y recuperación no desbordaron. En panel, a 320 y 360 la lista app-nav__list mide 568 px y usa scroll horizontal interno; a 768 y 1280 no se observó overflow | Evaluación DOM y capturas a 320 |
| 23 | B0 / accesibilidad | Tab, Shift+Tab y Enter permiten login sin mouse, con foco visible | No verificable literalmente con el controlador IAB: al campo enfocado se le vio foco, pero Tab no trasladó el foco desde body y Enter no envió el formulario | Snapshot/captura de foco y limitación del controlador |
| 24 | B1 / HU-020 | Barbería y cuatro campos visibles | Se observaron encabezado Barbería, Nombre, Zona horaria, Correo de contacto y Teléfono de contacto | Snapshot DOM y captura |
| 25 | B1 / HU-020 | Nombre, contacto y zona válida persisten | Cambios sintéticos de QA persistieron tras recarga y en la BD local | Captura y consulta de barbershop |
| 26 | B1 / HU-020 | Zona inválida rechaza sin escritura parcial | COT mostró No pudimos guardar los cambios y rechazó con mensaje de zona IANA; tras recarga conservó el nombre anterior | Alert y recarga |
| 27 | B1 / HU-021 | Encabezado Barberos visible | Se observó Barberos | Snapshot DOM |
| 28 | B1 / HU-021 | Alta única aparece sin recarga | El barbero sintético apareció inmediatamente | Lista observada |
| 29 | B1 / HU-021 | Nombre vacío muestra Escribe el nombre del barbero. y conserva diálogo | Mensaje exacto mostrado y diálogo permaneció abierto | Alert y diálogo |
| 30 | B1 / HU-021 | Renombrado elimina nombre anterior de la lista | La lista mostró el nombre editado y dejó de mostrar el anterior | Lista observada |
| 31 | B1 / HU-021 | Barbero B no aparece en A | El barbero sintético de B no apareció en la lista de A | Listas A y B comparadas |
| 32 | B1 / HU-022 | Encabezado Servicios visible | Se observó Servicios | Snapshot DOM |
| 33 | B1 / HU-022 | Alta con 30 min y 20000.00 aparece formateada | Apareció Activo y 30 min · $ 20000.00 COP | Fila de servicio |
| 34 | B1 / HU-022 | Edición refleja 45 min y 50000.00 | La fila reflejó 45 min · $ 50000.00 COP | Fila observada |
| 35 | B1 / HU-022 | Nombre vacío muestra mensaje y conserva diálogo | Mostró Escribe el nombre del servicio. y no cerró el diálogo | Alert y diálogo |
| 36 | B1 / HU-022 | Precio cero rechaza y conserva nombre | Mostró El precio debe ser mayor que cero. y conservó el nombre | Alert y formulario |
| 37 | B1 / HU-022 | Nombre activo duplicado muestra conflicto | Mostró Ese nombre ya está en uso y Ya existe un servicio activo con ese nombre. Usa otro nombre. | Alert y diálogo |
| 38 | B1 / HU-022 | Servicio B no aparece en A | El servicio sintético de B no apareció en Servicios de A | Listas A y B comparadas |
| 39 | B1 / HU-023 | Encabezado Servicios por barbero visible | Se observó Servicios por barbero | Snapshot DOM |
| 40 | B1 / HU-023 | Asignación persiste tras recarga | La casilla siguió marcada al recargar y seleccionar el mismo barbero | Casilla y recarga |
| 41 | B1 / HU-023 | Segundo barbero se asigna independientemente | Se asignó al segundo y el primero conservó su asignación | Selector y casillas |
| 42 | B1 / HU-023 | No se permite retirar última asignación activa | Alertó No puedes retirar ...: es el único barbero asignado a este servicio activo. Asigna otro barbero antes de retirar este.; la casilla permaneció marcada | Alert literal y casilla |
| 43 | B1 / HU-023 | Selector A excluye B y fetch cruzado devuelve 404 | Selector A solo mostró barberos A; fetch 404 no verificable desde IAB | Selector observado; limitación DevTools |
| 44 | B1 / HU-024 | Servicio activo muestra insignia y botón Desactivar | Se observó Activo y Desactivar | Fila observada |
| 45 | B1 / HU-024 | Diálogo consulta impacto y muestra 0 citas futuras | Se observó el diálogo y No hay citas futuras que se vean afectadas ahora mismo.; el estado de carga fue demasiado rápido para capturarlo | Diálogo y texto literal |
| 46 | B1 / HU-024 | Confirmación deja fila Inactivo sin borrar | La fila permaneció, cambió a Inactivo y mostró Reactivar | Fila observada |
| 47 | B1 / HU-024 | Inactivo persiste tras recarga | El estado Inactivo persistió | Recarga y fila |
| 48 | B1 / HU-024 | Reactivación conserva registro y vuelve a Activo | El diálogo informó que no cambia duración, precio ni asignaciones; luego volvió a Activo | Diálogo y fila |
| 49 | B1 / HU-024 | Duración y precio no cambian | Se conservaron 45 min y 50000.00 COP | Fila antes/después |
| 50 | B1 / HU-024 | Asignaciones no cambian tras ciclo | Las asignaciones de ambos barberos se conservaron | Servicios por barbero |
| 51 | B1 / responsive y accesibilidad | Anchos y teclado correctos en pantallas B1 | A 768/1280 no hubo overflow de página. A 320/360 persistió el scroll horizontal interno del menú; teclado quedó no verificable por la limitación del controlador | Evaluación DOM, capturas y limitación IAB |

## Divergencias y pendientes

| Prioridad | Paso(s) | Hallazgo | Impacto | Siguiente evidencia necesaria |
| --- | --- | --- | --- | --- |
| P1 | 22, 51 | El menú privado tiene scroll horizontal interno a 320/360 px: app-nav__list 568 px frente a viewport de 320/360 px | Incumple el criterio literal de responsive; la navegación sigue siendo desplazable | Decidir si el patrón de navegación desplazable es aceptable o corregirlo y repetir responsive |
| P1 | 13, 15-17 | No hay código OTP real para cerrar los flujos de reto y recuperación | Éxito del canal y revocación de sesiones no quedan probados manualmente | Configurar teléfono/captura o proveedor de mensajería de pruebas y repetir |
| P1 | 20, 43 | No se pudo ejecutar el fetch cross-tenant desde consola DevTools | El aislamiento UI sí se observó, pero falta la comprobación directa solicitada | Repetir en Chrome con consola y Network disponibles |
| P2 | 23, 51 | El controlador IAB no permitió una navegación fiable con Tab/Enter | La accesibilidad por teclado queda sin verificación manual completa | Repetir con Chrome/Playwright o una sesión donde las teclas se entreguen al documento |
| P2 | Documentación | matriz-trazabilidad.md y plan-bloques.md aún dicen que PR #84 está pendiente | Trazabilidad desactualizada, sin impacto en el comportamiento observado | Actualización documental separada, fuera del alcance original del prompt |

## Datos sintéticos creados

La ejecución dejó datos de QA con marca temporal en la base local: cambio de configuración de la barbería A, dos barberos A, un barbero B, servicios A y B y asignaciones necesarias para probar HU-023/HU-024. No se eliminaron porque el prompt no autorizaba limpieza destructiva.

## Integridad de la ejecución

- No se modificó código de aplicación.
- No se modificaron historias de usuario, matriz de trazabilidad ni plan de bloques.
- No se incluyeron credenciales, contraseñas, tokens ni códigos OTP.
- La consola del navegador no mostró errores ni advertencias al cierre.
