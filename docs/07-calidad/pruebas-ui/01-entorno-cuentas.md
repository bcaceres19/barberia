# Entorno, cuentas y acceso

## Preparación del coordinador

1. Releer AGENTS.md, DEC-050/057/106/116 y estrategia-pruebas.md. Aplicar local-app-startup: identificar procesos, conexiones y contenedor reales, Atlas y recorrido autenticado. No crear otra BD porque el sandbox no pueda verla.
2. Comprobar apps/api/.env.local sin mostrar sus valores. NAVA_QA_DATABASE_URL recibe DATABASE_URL del migrador; APP_DATABASE_URL conserva barberia_app. El aprovisionador exige loopback, base terminada en _test/_qa, local/test y ambos DSN apuntando al mismo destino.
3. Con PostgreSQL y esquema listos, desde raíz:

~~~sh
bash tools/qa/start-local.sh
~~~

En otra terminal:

~~~sh
cd apps/web
pnpm dev --host 127.0.0.1
~~~

El wrapper conserva .env.local. Quita del entorno de ese proceso credenciales Meta/Resend y selecciona OTP_PROVIDER=meta, marcador local. **Una captura sola no impide un envío real:** el remitente interno actúa primero. No arrancar worker/proveedores reales.

4. Verificar /health, /health/db, frontend y proxy. Abrir /acceso en Chromium, autenticar, ver Agenda y recibir 200 en Barberos y agenda diaria. Un puerto disponible no equivale a app funcional.
5. Ejecutar **node tools/qa/verify-ui-ready.mjs**. Este smoke acredita preparación, no la campaña completa.

## Cuentas y fixture

Manifiesto local: **apps/web/.auth/nava-qa/current.json** (0600, directorio 0700, ignorado por Git). Hay copia por campaña en ese directorio. Leer programáticamente solo la cuenta del scope asignado; nunca volcar JSON, contraseña, correo, teléfono, cookie o códigos en herramientas/chat/informe/stdout.

20 cuentas activas en 20 tenants: una por scope del inventario, journeys e isolation-a/isolation-b. Una cuenta inactive con contraseña válida y acceso deshabilitado, en access. Cada contraseña es aleatoria distinta, correo @example.test y teléfono sintético +999… exclusivo del marcador local.

Cada tenant tiene dos profesionales, tres servicios activos de 30/45/60 minutos y 20000/30000/10000 COP; los dos primeros asignados a ambos, tercero sin asignación. Jornada 08:00–20:00 todos los días America/Bogota, slug público y política DEFAULT 60 min/3 días/rejilla 15/cancelación 20/tardía permitida con motivo. Crear turnos mediante UI con fechas relativas; no fixtures fechadas que caducan.

Para aprovisionar sin iniciar API:

~~~sh
set -a
source apps/api/.env.local
set +a
export NAVA_QA_DATABASE_URL="$DATABASE_URL"
node tools/qa/provision-ui.mjs
~~~

Dos transacciones: tenants/usuarios con migrador; credenciales/datos con app y RLS. Si falla la segunda puede quedar la primera: reejecutar el mismo manifiesto completa sin borrar/restaurar. Tras éxito, un marcador privado <campaña>.provisioned impide volver a sembrar esa campaña al reiniciar API: no restaura asignaciones, horarios ni contraseñas modificadas. No borrar el marcador para refrescar pruebas; usar una campaña nueva. Si se reemplazó/vació la BD, el smoke lo detecta y se crea campaña nueva. Para campaña limpia usar --new-run, conservando las anteriores. No borrar/resetear tenants ajenos.

Después de recuperación, actualizar solo la contraseña de recovery en current.json y su copia de campaña, de forma privada; no cambiar cuentas de otras tareas.

## Navegador y códigos

Playwright está en apps/web. Instalar Chromium solo si falta; comprobar chromium.launch, no basta node_modules. Cada tarea controla su proceso/contexto propio. Contextos separan cookies, no datos de PostgreSQL: por eso se separan tenants.

La API de QA captura challenge.json y recovery.json en el directorio privado. Verificar destino y antigüedad antes de consumir. Reto y recuperación van en serie: cada archivo conserva el último envío.

Mi turno: capturar en memoria accessToken del POST REAL iniciado desde la UI de confirmación; navegar sin imprimirlo. Esta API no tiene variable de captura de confirmación: no inventarla ni esperar correo real.

No bajar umbrales de throttle. En QA local el coordinador puede asignar IP sintética distinta por tarea con route.continue/x-forwarded-for y APP_TRUSTED_PROXIES=127.0.0.1/32,::1/128. Mantener la IP estable dentro de cada caso de abuso. No sustituir rutas de negocio por mocks.
