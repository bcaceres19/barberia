---
prompt_id: "PROMPT-TEST-QA-B0-B1-CHROME-v1"
version: "1.0"
kind: "test"
status: "executed"
target_agents:
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
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
issue: null
issue_url: null
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on:
  - "HU-001 a HU-012 integradas en main (B0)"
  - "HU-020, HU-021, HU-022, HU-023 integradas en main (B1)"
  - "HU-024 integrada en main mediante PR #84 (squash-merge 2026-08-25T14:51:04Z, verificado con `gh pr view 84`); esto cierra B1 por completo, aunque docs/00-control/matriz-trazabilidad.md y docs/10-backlog/plan-bloques.md todavía dicen 'pendiente de revisión/CI/merge' — hallazgo documental fuera del alcance de este prompt"
rules: []
decisions:
  - "DEC-055"
  - "DEC-056"
  - "DEC-058"
  - "DEC-059"
  - "DEC-060"
  - "DEC-061"
  - "DEC-062"
  - "DEC-063"
  - "DEC-064"
  - "DEC-065"
  - "DEC-066"
  - "DEC-067"
  - "DEC-068"
  - "DEC-069"
  - "DEC-039"
acceptance_criteria:
  - "CA-007-*"
  - "CA-008-*"
  - "CA-010-*"
  - "CA-011-*"
  - "CA-012-*"
  - "CA-020-*"
  - "CA-021-*"
  - "CA-022-*"
  - "CA-023-*"
  - "CA-024-*"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/10-backlog/plan-bloques.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "apps/api/README.md"
  - "apps/web/README.md"
created_at: "2026-08-25"
updated_at: "2026-08-25"
execution_report: "test/2026-08-25-qa-manual-b0-b1-chrome-mcp-report.md"
supersedes: null
superseded_by: null
---

# QA manual en navegador real de B0 y B1 (HU-001–HU-012, HU-020–HU-024)

## Instrucción para el agente

Ejecuta, con un MCP de Chrome que controle un navegador real, los recorridos descritos abajo contra la aplicación corriendo en local. Este prompt **no autoriza cambios en el repositorio**: es verificación manual de comportamiento ya integrado en `main`, no implementación ni corrección. Si encuentras un defecto, repórtalo en la tabla final; no lo arregles ni abras rama/PR por tu cuenta.

## Objetivo

Producir un informe verificable (paso, resultado esperado, resultado observado, evidencia) de que los recorridos de aceptación de **B0** (`HU-001`–`HU-012`) y **B1** (`HU-020`–`HU-024`) funcionan en un navegador real, no solo en la suite Playwright. Los dos bloques ya están integrados en `main`:

- B0: `HU-001`–`HU-012`.
- B1: `HU-020`, `HU-021`, `HU-022`, `HU-023` y `HU-024` (PR #84, mergeado 2026-08-25). B2–B6 no existen todavía: no los busques ni reportes su ausencia como defecto.

## Preflight obligatorio

1. No se crea rama ni issue: esta ejecución no modifica el repositorio.
2. Lee completamente `AGENTS.md`, `CLAUDE.md`, `docs/02-requisitos/historias-usuario.md` (secciones `HU-001`–`HU-012` y `HU-020`–`HU-024`) y `docs/03-desarrollo/estandar-diseno-visual.md` (anchos de referencia y accesibilidad).
3. Ejecuta `graphify query "pantallas del panel privado B0 B1"` si necesitas ubicar componentes adicionales; no es obligatorio para correr los recorridos ya descritos aquí.
4. Levanta el entorno real siguiendo `apps/api/README.md` y `apps/web/README.md`:
   - PostgreSQL real con las migraciones aplicadas vía Atlas (no un contenedor vacío).
   - `go run ./cmd/api` corriendo contra esa base.
   - `npm run dev` en `apps/web` (por defecto `http://localhost:5173`).
   - Dos usuarios reales con hash argon2id real (no el fixture `database/testdata/hu005_credenciales_sesiones.sql`, que solo sirve para RLS): uno para la barbería "A" y otro para una segunda barbería "B", usados para las pruebas de aislamiento entre tenants.
5. Usa las mismas variables de entorno que la suite E2E real en vez de inventar credenciales: `E2E_EMAIL`/`E2E_PASSWORD` (tenant A) y `E2E_EMAIL_B`/`E2E_PASSWORD_B` (tenant B). Si no están definidas, pide al operador que las provea antes de continuar; no adivines contraseñas.
6. Para los pasos de HU-007/HU-008 que requieren leer un código de verificación real (WhatsApp/correo), coordina con quien levantó el entorno el mecanismo de captura (`APP_PHONE_CHALLENGE_CAPTURE_FILE` u otro definido en `apps/api/README.md`). Si no hay forma de leer el código real, ejecuta solo las ramas de error (código incorrecto/vencido, umbral, no enumeración) y marca la rama de éxito como "no verificable en este entorno" en vez de inventar un resultado.

## Alcance incluido

- Recorridos manuales en Chrome real de B0: acceso, sesión persistente, cierre de sesión, guard de rutas privadas, defensa contra abuso (HU-007), recuperación de acceso (HU-008/HU-011), aislamiento entre tenants, formato de error, accesibilidad/responsivo.
- Recorridos manuales en Chrome real de B1: configuración de la barbería (HU-020), barberos (HU-021), catálogo de servicios (HU-022), servicios por barbero (HU-023), ciclo de vida de servicios — desactivar/reactivar (HU-024).
- Captura de evidencia (capturas de pantalla por paso relevante, mensajes de error literales, estado de red cuando aplique).

## Fuera de alcance

- B2 a B6: no están implementados: no navegues rutas de horarios, agenda, reserva pública, notificaciones ni operación/backup, y no reportes su ausencia como defecto.
- Cualquier corrección de código, migración o documento. Los hallazgos se reportan, no se arreglan aquí.
- Actualizar `docs/00-control/matriz-trazabilidad.md` o `docs/10-backlog/plan-bloques.md` para reflejar el merge de PR #84 — es un hallazgo documental a reportar, no a corregir en este prompt.
- Crear datos de negocio permanentes: usa nombres con marca de tiempo (mismo patrón que `apps/web/e2e/*.spec.ts`, p. ej. `QA Manual Barbero ${Date.now()}`) para no chocar con datos reales ni dejar ambigüedad sobre qué fue creado por esta ejecución.

## Estado existente que debe conservarse

- Rutas ya definidas en `apps/web/src/app/router/index.ts`: `/acceso`, `/recuperar-acceso`, `/panel`, `/panel/barberia`, `/panel/barberos`, `/panel/servicios`, `/panel/servicios-por-barbero`.
- Selectores estables ya usados por la suite Playwright real (`apps/web/e2e/*.spec.ts`): estos prompts los citan porque un MCP de Chrome puede ubicarlos por rol/etiqueta accesible sin depender de clases CSS.

## Trabajo requerido

Ejecuta cada bloque en orden. Para cada paso numerado, registra: resultado esperado (ya escrito), resultado observado, y evidencia (captura de pantalla o texto literal visto).

### Bloque B0 — Cimientos, seguridad y primeras pantallas

**1. Acceso (HU-010) y guard de rutas privadas (HU-012)**
1. Sin sesión, navega a `/panel`. Debe redirigir a `/acceso?redirect=%2Fpanel` (o `/acceso?redirect=/panel`).
2. Verifica el encabezado "Inicia sesión" y los campos con etiqueta "Correo" y "Contraseña", más el botón "Iniciar sesión".
3. Envía credenciales inválidas. Debe aparecer una alerta que contenga "Revisa tu correo y contraseña".
4. Envía las credenciales reales del tenant A (`E2E_EMAIL`/`E2E_PASSWORD`). Debe terminar en `/panel` (el `redirect` guardado en el paso 1 debe respetarse) con el encabezado "Panel del barbero" visible.
5. Verifica que la cabecera muestra el nombre real de la barbería (no vacío).

**2. Sesión persistente (HU-006/HU-012)**
6. Con sesión iniciada, recarga `/panel` directamente (F5). Debe seguir en `/panel` con el mismo encabezado.
7. Cierra la pestaña/ventana y ábrela de nuevo apuntando a `/panel` (reutilizando el mismo perfil/contexto de navegador, cookie `HttpOnly` incluida). Debe entrar sin pedir credenciales de nuevo.
8. Haz clic en "Cerrar sesión". Debe volver a `/acceso`. Vuelve a navegar a `/panel`: debe redirigir de nuevo a `/acceso` (la sesión quedó revocada en el servidor, no solo limpiada en el cliente).

**3. Defensa escalonada contra abuso (HU-007)**
9. Desde `/acceso`, intenta iniciar sesión con contraseña incorrecta 5 veces seguidas para el mismo correo.
10. En el sexto intento (dentro de la ventana de 15 minutos) debe aparecer "Demasiados intentos" y "Verifica tu teléfono", en vez del mensaje genérico de credenciales inválidas.
11. Haz clic en "Enviar código por WhatsApp". Debe aparecer un mensaje que empiece con "Si tu cuenta existe" (mensaje deliberadamente genérico, no confirma si el número es real).
12. Ingresa un código de 6 dígitos incorrecto (p. ej. `000000`) en el campo "Código de 6 dígitos" y confirma con "Verificar código". Debe aparecer "El código no es válido o venció".
13. Si tienes forma de leer el código real emitido (ver preflight punto 6), ingrésalo y confirma: debe iniciar sesión y llegar a "Panel del barbero". Si no tienes acceso al código real, marca este paso como "no verificable en este entorno" en vez de simularlo.

**4. Recuperación de acceso (HU-008/HU-011)**
14. Navega a `/recuperar-acceso`. Ingresa el correo del tenant A y confirma con "Enviar código". Debe avanzar a "Paso 2 de 3".
15. Si tienes el código real (preflight punto 6): ingrésalo en "Código de 6 dígitos" y confirma con "Verificar código". Debe avanzar a "Paso 3 de 3" y mostrar el destino enmascarado (nunca el correo/teléfono completo en claro).
16. Define una contraseña nueva y guarda. Debe aparecer un mensaje que mencione que se cerraron todas las sesiones activas.
17. Haz clic en "Ir al acceso" e inicia sesión con la contraseña nueva. Debe entrar normalmente a `/panel`.
18. Rama de error: repite el flujo con un código vencido o inventado. Debe rechazar con un mensaje que contenga "El código no es válido", sin distinguir la causa exacta (no enumeración), y permanecer en "Paso 2 de 3".

**5. Aislamiento entre tenants (transversal a B0)**
19. En un contexto de navegador aparte (perfil distinto o ventana de incógnito), inicia sesión con las credenciales del tenant B (`E2E_EMAIL_B`/`E2E_PASSWORD_B`). Verifica que el nombre de barbería mostrado es distinto al del tenant A.
20. Con la sesión del tenant A activa, abre la consola de DevTools y ejecuta una petición `fetch` a un recurso propio del tenant B usando un id real (por ejemplo, un `barberId` de B obtenido en el paso 19 vía `fetch('/api/v1/private/barbers?limit=50', {credentials:'include'})`). Debe responder `404`, nunca `200` ni datos de B.

**6. Formato de error y accesibilidad/responsivo (transversal a B0)**
21. Provoca un error conocido (por ejemplo, credenciales inválidas) y revisa en DevTools → Network la respuesta: `Content-Type` debe ser `application/problem+json` con los campos del contrato RFC 9457 (`type`, `title`, `status`, `detail`).
22. Redimensiona el viewport a 320, 360, 768 y 1280 px para `/acceso`, `/recuperar-acceso` y `/panel`. En cada ancho: sin scroll horizontal, sin elementos superpuestos, todo control visible y usable.
23. Navega `/acceso` solo con teclado (Tab/Shift+Tab/Enter): el foco debe ser visible y debe poder completarse el login sin usar el mouse.

### Bloque B1 — Identidad de la barbería y catálogo

**7. Configuración de la barbería (HU-020)**
24. Inicia sesión como tenant A y entra a "Barbería" en la navegación (`/panel/barberia`). Verifica el encabezado "Barbería" y los campos "Nombre", "Zona horaria", "Correo de contacto", "Teléfono de contacto".
25. Cambia nombre, correo y teléfono de contacto; deja la zona horaria en un valor IANA válido (p. ej. `America/Bogota`) y guarda con "Guardar cambios". Recarga la página: los tres valores deben persistir.
26. Intenta guardar con una zona horaria inválida (p. ej. `COT`). Debe rechazarse (mensaje de error visible) y, tras recargar, el campo "Nombre" debe mostrar el valor anterior a este intento, no el que escribiste en el intento rechazado (sin escritura parcial).

**8. Barberos (HU-021)**
27. Entra a "Barberos" (`/panel/barberos`). Verifica el encabezado "Barberos".
28. Haz clic en "Agregar barbero", completa "Nombre" con un valor único con marca de tiempo y guarda. El nuevo barbero debe aparecer en la lista sin recargar.
29. Intenta guardar el diálogo "Agregar barbero" con el nombre vacío: debe mostrar "Escribe el nombre del barbero." y no cerrar el diálogo.
30. Edita el barbero recién creado con el botón "Editar {nombre}", cambia el nombre y guarda: la lista debe reflejar el nombre nuevo y ya no mostrar el anterior.
31. Aislamiento: crea un barbero en el tenant B; con la sesión del tenant A, confirma que ese nombre no aparece en la lista de A.

**9. Catálogo de servicios (HU-022)**
32. Entra a "Servicios" (`/panel/servicios`, enlace exacto "Servicios"). Verifica el encabezado "Servicios".
33. Agrega un servicio con nombre único con marca de tiempo, duración `30` y precio `20000.00`. Debe aparecer en la lista con "30 min" y el precio formateado.
34. Edita el mismo servicio: cambia duración a `45` y precio a `50000.00`; guarda y confirma que la fila refleja ambos cambios.
35. Intenta crear un servicio con nombre vacío: debe mostrar "Escribe el nombre del servicio." sin cerrar el diálogo.
36. Intenta crear un servicio con precio `0`: debe mostrar "El precio debe ser mayor que cero." y conservar el nombre ya escrito (no se pierde lo tecleado).
37. Intenta crear un segundo servicio activo con el mismo nombre que uno ya existente: debe mostrar "Ese nombre ya está en uso".
38. Aislamiento: confirma que un servicio creado en el tenant B no aparece en la lista del tenant A.

**10. Servicios por barbero (HU-023)**
39. Entra a "Servicios por barbero" (`/panel/servicios-por-barbero`). Verifica el encabezado "Servicios por barbero".
40. Selecciona un barbero en el selector "Barbero" y marca la casilla de un servicio para asignarlo. Recarga la página, vuelve a seleccionar el mismo barbero: la casilla debe seguir marcada (persistencia real).
41. Asigna ese mismo servicio a un segundo barbero: debe poder marcarse de forma independiente sin afectar la asignación del primero (vuelve a comprobar que el primero sigue marcado).
42. Con un servicio que tiene un único barbero asignado activo, intenta desmarcar esa única asignación: debe rechazarse, mostrar un mensaje que contenga "es el único barbero asignado a este servicio activo" y la casilla debe permanecer marcada.
43. Aislamiento: el selector "Barbero" de la barbería A nunca debe listar barberos de la barbería B; una consulta directa (`fetch` en consola) a `/api/v1/private/barbers/{idDeB}/services` desde la sesión de A debe responder `404`.

**11. Ciclo de vida de servicios: desactivar y reactivar (HU-024, cierre de B1)**
44. En "Servicios", localiza un servicio activo: debe mostrar la insignia "Activo" y un botón "Desactivar {nombre}".
45. Haz clic en "Desactivar {nombre}". Debe abrirse el diálogo "Desactivar servicio", mostrar brevemente "Consultando el impacto real…" y luego, como no existen citas todavía (B3 no implementado), "No hay citas futuras que se vean afectadas ahora mismo."
46. Confirma con el botón "Desactivar". El diálogo debe cerrarse y la fila del servicio debe mostrar ahora la insignia "Inactivo" y un botón "Reactivar {nombre}" en el lugar de "Desactivar {nombre}". El servicio **no debe desaparecer** de la lista (nunca se borra).
47. Recarga la página: el estado "Inactivo" debe persistir.
48. Haz clic en "Reactivar {nombre}". Debe abrirse el diálogo "Reactivar servicio" con el texto de que vuelve a ofrecerse sin cambiar duración, precio ni asignaciones. Confirma con "Reactivar": la insignia debe volver a "Activo" y el botón a "Desactivar {nombre}".
49. Verifica que la duración y el precio del servicio son los mismos de antes de desactivar/reactivar (no se alteraron).
50. Verifica en "Servicios por barbero" que las asignaciones de ese servicio a barberos no cambiaron por el ciclo de desactivar/reactivar.

**12. Responsivo y accesibilidad de B1 (transversal)**
51. Repite la verificación de anchos (320, 360, 768, 1280 px) y navegación solo por teclado para `/panel/barberia`, `/panel/barberos`, `/panel/servicios` y `/panel/servicios-por-barbero`.

## Pruebas y evidencia

- Una captura de pantalla por cada paso donde el resultado sea visualmente verificable (formularios, mensajes de error, insignias de estado, diálogos).
- El texto literal exacto de cada mensaje de error/éxito mostrado, para comparar contra el texto esperado citado arriba.
- Para los pasos de red (formato de error, aislamiento entre tenants vía `fetch`), el código de estado HTTP y el `Content-Type` observados.
- Para los pasos de responsivo, una captura por ancho evaluado.

## Documentación y trazabilidad

- Este prompt no actualiza documentación normativa. Si algún paso falla, no edites `historias-usuario.md`, `matriz-trazabilidad.md` ni ningún criterio de aceptación: repórtalo en la tabla final para que el propietario decida si es un defecto real o una duda a registrar en `docs/00-control/dudas-pendientes.md`.
- Reporta explícitamente, como hallazgo aparte (no como defecto de producto), que `docs/00-control/matriz-trazabilidad.md` y `docs/10-backlog/plan-bloques.md` describen `HU-024`/PR #84 como "pendiente de revisión/CI/merge" cuando en realidad el PR ya está fusionado en `main` (confirmado con `gh pr view 84`, `mergedAt: 2026-08-25T14:51:04Z`).

## Verificación final

No hay comandos de build/lint/test que correr: este prompt es exploración manual sobre la aplicación ya corriendo. Entrega una tabla:

`Paso | Bloque/HU | Resultado esperado | Resultado observado | Evidencia`

Y una segunda tabla solo con las filas cuyo resultado observado difiera del esperado, priorizadas por impacto.

No declares cumplido un paso que no ejecutaste realmente en el navegador; si un paso no fue verificable (por ejemplo, por falta de acceso a un código real de WhatsApp), dilo explícitamente en vez de omitirlo o darlo por bueno.

## Git y PR

No aplica: esta ejecución no crea rama, commit ni PR. Si durante la ejecución identificas que hace falta un cambio de código, documenta el hallazgo con el paso exacto que lo reproduce y detente ahí — la decisión de abrir issue y prompt de corrección es del propietario del proyecto.
