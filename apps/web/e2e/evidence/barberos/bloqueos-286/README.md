# Bloqueos dentro de Barberos · #286

Modo de conformidad: **identidad guiada** NAVA / Tailored Grid, según la instrucción del propietario del 2026-09-30 y DEC-105. No se asignó un mockup exacto.

La fila de cada profesional conserva detalle, edición y fotografía, y añade «Bloquear». La acción abre una ventana modal amplia, sin expandir la fila. El panel contextual de la ventana muestra sus bloqueos puntuales y series, fechas, horas, motivo y vigencia; permite crear y retirar mediante los contratos existentes de HU-042. La URL antigua redirige a Barberos. Formularios con identidad y resumen previo; pie fijo, errores en línea, idempotencia conservada y cierre bloqueado al guardar. Los registros entran/salen con transición; la ventana respeta movimiento reducido. En móvil las listas se apilan y se recorren con scroll interno, conservando la cabecera de cierre. Escape cierra únicamente el formulario superior y devuelve el foco a su acción; al cerrar la consulta vuelve a «Bloquear».

## Verificación

- `pnpm typecheck`, `pnpm build`: correctos; componentes expuestos de forma diferida por sus índices públicos y compuestos en `app`.
- `pnpm lint`: sin errores; 45 advertencias previas en otros componentes compartidos y de agenda/autenticación.
- Vitest: 93 pruebas correctas (14 de BarberBlocksPanel, 32 de StaffPage, 13 de AppNav y 34 de BaseDialog). Incluye formularios, errores/reintentos, retiro, paginación y descarte de una respuesta atrasada tras retirar un bloqueo; la regresión de Escape consumido dentro de diálogos superpuestos tiene prueba de componente.
- `pnpm exec playwright test e2e/barberos-bloqueos.spec.ts --project=chromium-desktop --project=chromium-mobile --project=firefox --workers=1 --reporter=line`.
- Los tres proyectos E2E pasaron; no se registraron errores de ejecución (`pageerror`).
- E2E contra API y PostgreSQL locales reales: bloqueo nocturno, serie semanal, consulta tras cambiar de profesional, retiro, formulario conservado ante error, redirección y aislamiento de dos tenants (404 al consultar el barbero ajeno).
- Viewports medidos con `innerWidth`/`innerHeight`: 320×740, 360×800, 768×1024, 1280×900 y 640×450. Este último verifica reflow equivalente al espacio disponible a 200%; no sustituye una medición de zoom físico del navegador.
- Axe WCAG A/AA, contraste incluido: pantalla o diálogo activo, una vez terminadas las transiciones finitas (sin desactivar las animaciones). Escape, restauración de foco y `prefers-reduced-motion` comprobados. Sin desbordamiento horizontal en los anchos medidos.
- `pnpm openapi:lint`: válido. `atlas migrate validate --dir file://database/migrations`: válido; ninguna migración ni checksum cambiado por esta integración.

## Capturas

Cada motor guarda `*-{320,360,768,1280,640}-equipo.png`, `*-panel.png`, `*-listas-fin.png` y `*-formulario.png`, más los formularios puntuales y semanales completos. Todos los datos son sintéticos. Las capturas muestran la ventana de consulta y el formulario sobre ella; las de listas-fin prueban el contenido inferior de la zona de scroll; una sola captura de viewport no representa toda la lista móvil.

Referencias de escritorio: [equipo](chromium-desktop-1280-equipo.png), [panel](chromium-desktop-1280-panel.png), [formulario](chromium-desktop-1280-formulario.png).

Referencias móviles: [equipo 320](chromium-desktop-320-equipo.png), [formulario 320](chromium-desktop-320-formulario.png), [equipo Firefox 320](firefox-320-equipo.png).

## Reproducción local

Aplicar `database/testdata/ui_bloqueos_barberos_286.sql` con rol migrador únicamente en la BD local de pruebas y arrancar API/Vite. La suite usa esas dos cuentas ficticias y conserva sesiones solo en `apps/web/.auth/bloqueos-286-{1,2}.json`, ignorado por Git. Así, ejecutar varios motores no consume innecesariamente el umbral de HU-007. Si faltan, la suite inicia sesión con las credenciales ficticias del fixture; si el entorno ya alcanzó el umbral, preparar sesiones efímeras de esos usuarios de prueba o esperar la ventana, sin desactivar los controles de acceso. En esta verificación se aprovisionaron sesiones efímeras de los dos usuarios sintéticos para continuar después de alcanzar el límite local.

Las pruebas retiran lógicamente los registros que crean incluso al fallar; no borran registros ni modifican cuentas del propietario.

## Alcance y reversión

Seguimientos `date_list`, edición/excepciones de series y turnos afectados permanecen en #100 y las HU dueñas. Revertir la composición y la navegación restaura el acceso separado sin migrar ni perder registros.

Los cambios locales previos de fotografías, servicios y agenda se conservaron. La actualización de main requirió conciliar DEC-099/100 locales (Servicios/fotografía) con las ya publicadas para Google Calendar: ahora son DEC-103/104. La migración aplicada conserva su referencia histórica original y su checksum.

Resultado de visual-qa: **passed**, tras revisar capturas reales de equipo, panel y formularios en escritorio y móvil. Auto-revisión `change-review`: sin hallazgos accionables P0–P2 dentro del alcance; los seguimientos de #100 continúan excluidos.
