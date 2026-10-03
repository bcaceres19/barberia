---
name: local-app-startup
description: "Start or diagnose the local NAVA app: Docker PostgreSQL, Atlas migrations, Go API, Vue frontend and Playwright. Use when asked to turn on, restart or verify the local app. Check schema and authenticated agenda readiness before declaring success. Exclude production deployments and feature implementation."
---

# Arranque local de NAVA

El resultado es una app local utilizable, con esquema compatible y un recorrido real comprobado. PostgreSQL accesible o `/health` en verde no demuestran que la agenda funcione.

Fuente: solicitud del propietario del 2026-10-02, DEC-106, issue [#287](https://github.com/bcaceres19/barberia/issues/287). Sigue [migraciones Atlas](../../../docs/05-backend/migraciones-atlas.md), [estándar de BD](../../../docs/05-backend/estandar-base-datos.md) y [estrategia de pruebas](../../../docs/03-desarrollo/estrategia-pruebas.md).

## 1. Identificar el entorno y reutilizarlo

- Lee AGENTS.md, la configuración local y los comandos reales de `apps/api` y `apps/web`. No deduzcas el estado actual desde un README que diga que todavía no hay BD: contrasta código, migraciones y procesos.
- Identifica dónde están Docker, Go, pnpm, Atlas y el navegador. Si estás dentro de Flatpak y las herramientas viven en el anfitrión, usa `flatpak-spawn --host` para operar allí; su `/tmp` puede ser distinto. Respeta los controles de permisos de la sesión.
- Comprueba contenedores, puertos y procesos existentes antes de arrancar nada. Reutiliza la instancia correcta; inicia solo lo detenido. No crees otra BD ni otro servidor porque el sandbox no vea el anfitrión.
- Confirma el contenedor y la base de destino con consultas sin datos personales. `barberia-postgres`, `barberia_test`, 5432/8080/5173 son pistas del entorno observado, no valores que deban imponerse a otra configuración.
- No detengas procesos ajenos ni borres volúmenes, tablas o datos. No instales dependencias si las herramientas existentes bastan.

## 2. Cargar configuración sin exponer secretos

- Go no carga `.env.local` automáticamente. Carga la configuración existente en el entorno del proceso antes de ejecutar `go run ./cmd/api` desde `apps/api`.
- Verifica presencia y validez con `internal/platform/config`: al menos `APP_DATABASE_URL`, `APP_AUTH_HMAC_SECRET` y ambiente local. No imprimas valores ni el entorno completo; conserva el proveedor OTP y otros ajustes existentes.
- API: `barberia_app`. Atlas: `barberia_migrator`, miembro de `barberia_owner`, según DEC-040. Worker, cuando se solicite o el recorrido lo requiera: `barberia_worker` con `APP_WORKER_DATABASE_URL` propio. No uses el rol migrador ni `postgres` en la API.
- `APP_DATABASE_URL` es la conexión de la API; `DATABASE_URL` es la conexión de Atlas para `--env local`. No intercambies ambas solo por compartir host/base.
- Si falta una credencial, busca únicamente su configuración destinada al proyecto. No pruebes contraseñas adivinadas ni reutilices `POSTGRES_PASSWORD` para otro rol. Prender la app no autoriza rotar claves. Una generación/rotación requiere autorización específica y un destino local ignorado por Git, con permisos restringidos.
- Comprueba autenticación TCP con el rol app y la BD efectiva antes de arrancar. Una consulta por socket como `postgres` o `pg_isready` solo prueba disponibilidad del servidor.

## 3. Sincronizar esquema antes de declarar listo

Ejecuta desde `database/`, con la conexión del migrador en el entorno:

```sh
atlas migrate validate --env local
atlas migrate status --env local
atlas migrate apply --env local --dry-run
```

- Comprueba versión aplicada, siguiente versión, checksums y pendientes contra el directorio del checkout que ejecutará Go. No cambies una migración aplicada ni regeneres `atlas.sum` para ocultar discrepancias.
- Para una petición de arranque local, aplica las migraciones pendientes existentes y revisadas cuando el destino sea inequívocamente local, estén dentro del alcance autorizado y no haya operaciones destructivas ni dudas de compatibilidad. Guarda un respaldo privado antes de actualizar una base con datos. No expongas el dump ni lo versiones.
- Ejecuta `atlas migrate apply --env local` como paso separado; luego exige `atlas migrate status --env local` con cero pendientes. Nunca añadas DDL al arranque de Go.
- Ante precondiciones fallidas, permisos, checksum o historial desfasado, detén la aplicación de migraciones y diagnostica. Un objeto existente puede venir de SQL manual aunque Atlas no lo registre; un objeto invisible en `information_schema` puede existir sin permisos para el migrador.
- No arregles permisos mediante transferencias masivas, superusuario en la API, desactivar RLS o grants globales. Delimita objetos y privilegios conforme a la migración y registra la reparación; si necesita autorización adicional, solicítala sobre el cambio concreto.
- `atlas migrate set` no es un atajo de arranque. Solo se usa con autorización específica de reconciliación, respaldo y evidencia de equivalencia completa (definición, índices, restricciones, RLS, grants, comentarios y datos cuando aplique) para cada versión que marcaría. Comprueba que no salta versiones intermedias. Si no puedes demostrarlo, no lo ejecutes ni recrees/borras objetos para sortear la precondición.
- Una reparación que cambia SQL del repositorio requiere su issue/rama y las pruebas normativas; no inventes tablas para que desaparezca un 500.

## 4. Arrancar y comprobar el recorrido

- Inicia la API y el frontend solo si no están ya activos, con procesos persistentes y logs locales sin secretos. Vite debe apuntar a esta API mediante su proxy real. Inicia Playwright cuando se solicite o se necesite para comprobar la app.
- Espera disponibilidad con un plazo acotado. Comprueba HTTP 200 y cuerpo esperado en `/health` y `/health/db`; comprueba también que el frontend sirve y que su proxy llega al API. Revisa los logs si un proceso sale.
- Con una sesión local ya autorizada, abre `/panel` en Playwright. Si falta sesión, usa únicamente credenciales ficticias configuradas para este fin; no inventes un usuario ni leas datos personales para iniciar sesión. No dispares OTP real como parte del smoke. Si no puedes autenticarte, informa que la comprobación funcional quedó pendiente.
- Exige 200 en `GET /api/v1/private/barbers`, configuración de barbería y `GET /api/v1/private/barbers/{barberId}/appointments/daily-agenda` cuando exista un barbero. Verifica que se ve Agenda y no su estado de error. Un 401 no cuenta como prueba funcional exitosa; distingue falta de sesión de un 500.
- Con cero barberos, comprueba el estado vacío válido y registra que no se probó agenda por barbero. Comprueba los demás módulos solicitados, sin escribir datos de negocio como parte de un smoke de lectura.
- Si aplicaste migraciones, ejecuta las pruebas existentes apropiadas sobre PostgreSQL real y dos tenants; prefiere una base efímera o fixtures dedicados y declara sus efectos. Para cambios visibles de código aplica visual-qa; encender servicios no constituye un rediseño ni requiere crear capturas nuevas por defecto.

## 5. Cierre verificable

Informa URLs, qué reutilizaste y qué arrancaste, estado de Atlas, salud API/BD y resultado funcional autenticado. Distingue «servicios activos» de «app comprobada». Si falta configuración, hay pendientes o falla el recorrido, reporta el componente exacto y la evidencia; no declares todo listo.

Caso de regresión: `/health/db` respondió OK mientras `/barbers` devolvía 500 por falta de `barber_photo`. El arranque solo quedó comprobado después de sincronizar Atlas y obtener 200 en barberos y agenda. Esa secuencia debe quedar cubierta cada vez que se use este skill.
