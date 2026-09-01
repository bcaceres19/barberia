# apps/web

Aplicación Vue 3 + TypeScript + Vite del flujo público y del panel del
barbero. Ver
[`docs/04-arquitectura/frontend.md`](../../docs/04-arquitectura/frontend.md)
y [`docs/03-desarrollo/estandar-frontend-vue.md`](../../docs/03-desarrollo/estandar-frontend-vue.md)
antes de agregar código.

## Estado

Arranque de la aplicación, router, el **sistema visual base (HU-009)**
(tokens y cinco componentes en `src/shared/ui/`), la **pantalla de acceso
(HU-010)**, el **cascarón del panel privado (HU-012)** (guard generalizado
con sesión real, cabecera con barbería activa y coordinación única de 401),
la **defensa escalonada contra abuso (HU-007)**: reto telefónico inline en
la pantalla de acceso cuando el login responde 429, la **configuración
básica de la barbería (HU-020)**: pantalla "Barbería" (`/panel/barberia`)
para consultar y actualizar nombre, zona horaria y contacto opcional, con
actualización inmediata de la cabecera tras guardar, el **registro y
listado de barberos (HU-021)**: pantalla "Barberos" (`/panel/barberos`)
con lista paginada, alta con idempotencia real y edición del nombre; una
barbería con una persona y una con varias usan el mismo componente y el
mismo estado de datos, y el **catálogo básico de servicios (HU-022)**:
pantalla "Servicios" (`/panel/servicios`) con lista paginada, alta y
edición de nombre/descripción/duración/precio, precio en COP como string
decimal exacto y conflicto de nombre entre servicios activos (`DEC-067`)
distinguido del conflicto de idempotencia, y la **asignación de servicios a
barberos (HU-023)**: pantalla "Servicios por barbero"
(`/panel/servicios-por-barbero`) con un selector de barbero y una casilla
por servicio del catálogo; el mismo componente funciona con un barbero que
presta todo el catálogo y con un equipo de especialidades distintas. Cada
casilla se deshabilita mientras su propia solicitud está en curso (evita
doble envío) y solo cambia de estado tras la respuesta real del servidor;
retirar la última asignación activa de un servicio activo se rechaza
(`DEC-068`) y la casilla vuelve a marcarse, y el **ciclo de vida de
servicios (HU-024)**: en la misma pantalla "Servicios", desactivar/reactivar
con un `BaseDialog` que consulta el impacto real de citas futuras antes de
confirmar (siempre 0 en B1, `DEC-069`) y recarga el estado real si el
servidor responde un conflicto de transición, en vez de asumir éxito
optimista, y el **horario laboral recurrente (HU-040)**: pantalla
"Horarios" (`/panel/horarios`) con un selector de barbero y los siete días
ISO siempre visibles, agrupando los tramos existentes de cada uno; agregar,
editar y retirar un tramo con validación de campo en cliente (día/hora/
duración) y traducción del conflicto de solape real del servidor
(`code: "conflict"`) a un mensaje recuperable sin cerrar el diálogo. La
pantalla indica explícitamente la zona IANA de la barbería junto a las
horas (`CA-040-06`), nunca la del dispositivo, y las **excepciones de
jornada y festivos (HU-041)**: en la misma pantalla "Horarios", un
interruptor de calendario de festivos colombianos por barbero, un CRUD de
excepciones de jornada por fecha (día cerrado o abierto con tramos
especiales, con la misma traducción de conflicto de fecha duplicada
\-`CA-041-05`- a un mensaje recuperable) y una referencia de "Próximos
festivos colombianos" con un atajo para prellenar la fecha del diálogo de
alta. Ver las secciones siguientes. El resto de carpetas de `modules/`
conserva su `index.ts` de marcador de responsabilidad futura.

## Acceso del barbero (HU-010)

Fuente normativa: `docs/02-requisitos/historias-usuario.md` (HU-010) y
`docs/00-control/registro-decisiones.md` (`DEC-056`). Esta sección resume
la API pública y las decisiones de implementación; no las sustituye.

### Rutas

`src/modules/auth/routes.ts` expone `authRoutes`, importado una sola vez
desde `src/app/router/index.ts`:

- `/acceso` (`name: "acceso"`) — pantalla de acceso, diferida.
- `/panel` — cascarón privado (`src/modules/auth/layouts/PrivateShell.vue`),
  protegido por un único guard generalizado
  (`beforeEnter: requireSession`, `src/modules/auth/guards/requireSession.ts`,
  HU-012). Desde HU-020, `auth.privateShellRoute(children, extraNavItems)`
  es una fábrica: `src/app/router/index.ts` la invoca una sola vez,
  combinando las hijas y las entradas de navegación de `auth`
  (`privateShellChildRoutes`, hoy solo `""` → `name: "panel"`) con las de
  cualquier otro módulo con pantalla privada (`settings.
settingsPrivateShellChildRoutes`/`settingsNavItems`, hoy `barberia` →
  `name: "configuracion-barberia"`). Ningún módulo importa a otro para
  lograrlo (app → modules → shared, ver
  `src/shared/navigation/navItem.ts`); toda ruta privada futura se agrega
  igual, como hija de esta misma ruta padre, sin repetir el guard ni crear
  un segundo cascarón.
- `/recuperar-acceso` (`name: "recuperar-acceso"`) — destino real (no
  roto) del enlace de recuperación de `CA-010-08` mientras `HU-011` no
  existe. Ver `DP-UX-06` en `docs/00-control/dudas-pendientes.md`: ninguna
  fuente aprobada define todavía esta transición; la interpretación
  aplicada (declarada, no oculta) es un aviso explícito de "todavía no
  disponible", sin simular el flujo real de `HU-011`.

### Cascarón del panel privado (HU-012)

Fuente normativa: `docs/02-requisitos/historias-usuario.md` (HU-012) y
`docs/00-control/registro-decisiones.md` (`DEC-060`). Reemplaza el guard
mínimo y el marcador local `sessionStorage` de HU-010 (`DP-SEG-08`,
`sessionMarker.ts`, eliminado): ahora existe una lectura privada real,
`GET /private/auth/session`, y el guard la consulta antes de decidir.

- **`src/modules/auth/model/sessionStore.ts`**: único estado compartido
  entre rutas privadas de la aplicación (singleton reactivo, sin Pinia —
  `docs/03-desarrollo/estandar-frontend-vue.md` §113 solo la exige con
  estado compartido real; una sola pieza de estado no justifica la
  dependencia todavía, `DEC-035`). Expone `sessionState` (solo lectura),
  `ensureBootstrapped`/`retryBootstrap` (consulta el servidor, comparte una
  única solicitud en curso entre llamadas concurrentes) y la coordinación
  única de 401 (`reportUnauthorized`/`onUnauthorized`, `CA-012-03`):
  idempotente, así que varias respuestas 401 casi simultáneas producen una
  sola limpieza/notificación.
- **`src/modules/auth/guards/requireSession.ts`**: generalizado a toda
  ruta privada. Sin sesión real, redirige a `/acceso` conservando el
  destino pretendido como `?redirect=` (solo si es una ruta interna
  segura, nunca externa ni un ciclo hacia acceso/logout,
  `src/modules/auth/model/redirectTarget.ts`); `LoginPage.vue` lo recupera
  tras un acceso exitoso (`CA-012-02`).
- **`src/modules/auth/bootstrap/installSessionHandling.ts`**: único punto
  que conecta el middleware de respuesta del cliente HTTP
  (`openapi-fetch`, `client.use()`) con la coordinación de 401 del store, y
  esa coordinación con el router real. Se instala una sola vez desde
  `app/bootstrap/createApp.ts` — el único lugar con la instancia real del
  router — nunca desde dentro de `modules/auth` (`app → modules → shared`).
  Excluye deliberadamente `GET /private/auth/session`: el guard ya
  interpreta el 401 de ese endpoint por su cuenta, evitando una redirección
  compitiendo con la que el guard está resolviendo.
- **`src/modules/auth/layouts/PrivateShell.vue`**: cabecera
  (`AppHeader.vue`, barbería activa siempre visible, `CA-012-04`, y cierre
  de sesión, `CA-012-07`) + navegación (`AppNav.vue`, hoy solo "Panel": no
  inventa destinos de B1 en adelante que todavía no existen) + contenido,
  solo en estado `authenticated`; `checking`/`connection-lost` (con
  "Reintentar", `CA-012-05`) se renderizan en su lugar, nunca una pantalla
  en blanco.

### Cliente HTTP tipado

Todo acceso HTTP pasa por `src/shared/api/httpClient.ts`
(`openapi-fetch` + tipos generados). Comando reproducible para regenerar
los tipos tras un cambio de contrato:

```bash
# Desde la raíz del repositorio: relinta y regenera el bundle real.
pnpm run openapi:bundle
# Desde apps/web: genera los tipos TypeScript desde ese bundle.
pnpm run generate:api
```

`src/shared/api/generated/openapi.d.ts` es generado; nunca se edita a
mano (está excluido de Prettier y documentado como tal en `.prettierignore`).

### Estados de error del módulo `auth`

`src/modules/auth/api/loginApi.ts` traduce la respuesta real de
`POST /public/auth/login` a `LoginOutcome` (unión discriminada), mapeando
por `status` (nunca por `detail`): `success`, `invalid-credentials` (401),
`validation-error` (400/422), `rate-limited` (429, `DEC-061`/`DEC-062`
desde HU-007: sexta solicitud dentro de la ventana, lee `Retry-After`) y
`unexpected-error` (cualquier otro estado, con `requestId` cuando el
`Problem` lo trae). `LoginPage.vue` muestra `PhoneChallengeForm.vue`
(`src/modules/auth/api/challengeApi.ts`) cuando el estado es
`rate-limited`: pide el código de 6 dígitos y, al verificarlo, reintenta el
login automáticamente con las credenciales ya escritas. El mensaje de
"código enviado" es deliberadamente genérico (no enumeración, `DEC-062`):
nunca afirma que un WhatsApp real llegó.

### Pruebas

Componente (`src/modules/auth/components/__tests__`,
`src/modules/auth/layouts/__tests__`, `src/modules/auth/pages/__tests__`),
store/coordinador (`src/modules/auth/model/__tests__/sessionStore.test.ts`,
`src/modules/auth/bootstrap/__tests__/installSessionHandling.test.ts`),
cliente tipado (`src/modules/auth/api/__tests__`), enrutamiento/guard
(`src/modules/auth/__tests__/routes.test.ts`), accesibilidad (`vitest-axe`
integrado en las suites anteriores) y E2E contra el API real en local:
`e2e/acceso.spec.ts`/`e2e/acceso-evidencia-responsiva.spec.ts` (HU-010,
actualizados donde HU-012 cambió el comportamiento observable — el guard
ahora conserva `?redirect=`), `e2e/panel.spec.ts`/
`e2e/panel-evidencia-responsiva.spec.ts` (HU-012, capturas en
`e2e/evidence/panel/`) y `e2e/reto-telefonico.spec.ts` (HU-007: umbral,
reto telefónico completo con el código real capturado vía
`APP_PHONE_CHALLENGE_CAPTURE_FILE`, código incorrecto no enumerable —
ver `apps/api/README.md` sección "Defensa escalonada contra abuso
(HU-007)" para el candado de umbral al correr toda la suite E2E junta).
El recorrido E2E requiere `apps/api` corriendo
contra PostgreSQL real con las migraciones aplicadas y un usuario con un
hash argon2id real (no el hash ficticio de
`database/testdata/hu005_credenciales_sesiones.sql`, que solo sirve para
probar forma/RLS); variables `E2E_EMAIL`/`E2E_PASSWORD` sobrescriben las
credenciales de prueba por defecto.

## Configuración básica de la barbería (HU-020)

Fuente normativa: `docs/02-requisitos/historias-usuario.md` (HU-020) y
`docs/00-control/registro-decisiones.md` (`DEC-019`, `DEC-024`). Esta
sección resume la API pública y las decisiones de implementación; no las
sustituye.

### Ruta y navegación

`src/modules/settings/routes.ts` expone `settingsPrivateShellChildRoutes`
(`barberia` → `name: "configuracion-barberia"`, `src/modules/settings/
pages/SettingsPage.vue`, diferida) y `src/modules/settings/index.ts`
expone además `settingsNavItems` (`{ to: { name:
"configuracion-barberia" }, label: "Barbería" }`). `src/app/router/index.ts`
combina ambas con las de `auth` al construir el único cascarón privado (ver
"Cascarón del panel privado (HU-012)" más arriba); `settings` no importa
ningún interno de `auth` ni monta su propio guard.

### Formulario y estados

`SettingsPage.vue` carga `GET /private/settings/barbershop` al montar,
edita los cuatro campos autorizados (`name`, `timezone`, `contactEmail`,
`contactPhone`) y guarda con `PATCH` del mismo recurso
(`src/modules/settings/api/settingsApi.ts`, mismo patrón de outcome
discriminado por `status` que `loginApi.ts`). Estados explícitos: carga,
error de carga recuperable con "Reintentar", listo, guardando, error de
campo (validación de cliente en `src/modules/settings/validation/
settingsValidation.ts`, forma únicamente — nunca valida el catálogo IANA,
eso lo confirma el servidor), error de validación del servidor (incluida
una zona horaria no reconocida, `CA-020-03`), error de red y éxito. Un
error recuperable NUNCA borra lo que el barbero ya escribió (`CA-020-08`);
solo la respuesta `200` confirmada por el servidor reemplaza los valores
del formulario.

### Actualización inmediata de la cabecera

Tras un guardado exitoso, `SettingsPage.vue` llama a la función pública
estrecha `updateBarbershopName` (exportada por `auth`, HU-020) con el
nombre YA confirmado por el servidor: es el único punto del código que
actualiza la cabecera antes de la próxima rehidratación completa
(`GET /private/auth/session` sigue siendo la autoridad en cada recarga).
Ningún camino de este módulo aplica optimismo: un fallo de guardado nunca
invoca esta función.

### Presentación de instantes independiente del dispositivo (CA-020-04)

`src/shared/time/formatInstant.ts` (`formatInstantInTimezone`) es la única
pieza que HU-020 aporta para esta garantía: recibe la zona SIEMPRE
explícita (nunca la del dispositivo) y usa `Intl.DateTimeFormat` (API de
plataforma, sin librería de fechas nueva). Sin consumidor real todavía
(esta pantalla no muestra ningún instante formateado); queda lista para que
`B2`/`B3` la reutilicen en vez de reinventar el mismo cálculo.

### Pruebas

Componente (`src/modules/settings/pages/__tests__/SettingsPage.test.ts`),
cliente tipado (`src/modules/settings/api/__tests__/settingsApi.test.ts`),
validación (`src/modules/settings/validation/__tests__/
settingsValidation.test.ts`), presentación de instantes
(`src/shared/time/__tests__/formatInstant.test.ts`), accesibilidad
(`vitest-axe` integrado en `SettingsPage.test.ts`) y E2E contra el API real
en local: `e2e/configuracion-barberia.spec.ts` (edición/guardado con
persistencia real tras recargar, zona no reconocida sin pérdida de datos)
y `e2e/configuracion-barberia-evidencia-responsiva.spec.ts` (capturas en
`e2e/evidence/configuracion-barberia/`, un solo inicio de sesión
reutilizado entre los cinco breakpoints para no competir con el umbral de
`HU-007` al correrse junto al resto de la suite E2E).

## Registro y listado de barberos (HU-021)

Fuente normativa: `docs/02-requisitos/historias-usuario.md` (HU-021) y
`docs/00-control/registro-decisiones.md` (`DEC-019`, `DEC-024`, `DEC-043`,
`DEC-047`). Esta sección resume la API pública y las decisiones de
implementación; no las sustituye.

### Ruta y navegación

`src/modules/staff/routes.ts` expone `staffPrivateShellChildRoutes`
(`barberos` → `name: "staff-barberos"`, `src/modules/staff/pages/
StaffPage.vue`, diferida) y `src/modules/staff/index.ts` expone además
`staffNavItems` (`{ to: { name: "staff-barberos" }, label: "Barberos" }`).
`src/app/router/index.ts` combina ambas con las de `auth`/`settings` al
construir el único cascarón privado (ver "Cascarón del panel privado
(HU-012)"); `staff` no importa ningún interno de `auth` ni monta su propio
guard.

### Un mismo componente para barbería unipersonal y de equipo (CA-021-01/02)

`StaffPage.vue` renderiza `barbers.value` (un arreglo plano de `Barber`)
sin ninguna rama que distinga "un barbero" de "varios": una lista con 1
elemento y una con 4 pasan por el mismo `<li v-for>`. El alta hace
`unshift` del recurso confirmado por el servidor (nunca antes de esa
confirmación: sin actualización optimista); el renombrado reemplaza por
`id` en su misma posición, sin duplicar ni reordenar de forma inestable.

### Paginación por cursor (`src/modules/staff/api/staffApi.ts`)

`fetchBarbers(cursor?)` traduce `GET /private/barbers` (`cursor`/`limit`
como parámetros de consulta tipados por el cliente generado) a
`FetchBarbersOutcome`. `StaffPage.vue` carga la primera página al montar y
expone "Cargar más" cuando `nextCursor` no es `null`; las páginas
siguientes se concatenan por `id` (defensivo contra un doble clic muy
rápido), nunca se reemplaza la lista completa.

### Alta con idempotencia real (RN-IDE-01, DEC-043)

`src/modules/staff/model/idempotencyKey.ts` (`newIdempotencyKey`,
`crypto.randomUUID()`) genera la clave de UN intento lógico al abrir el
diálogo de alta; el mismo intento (por ejemplo, un reintento tras un error
de red con el diálogo todavía abierto) reutiliza la MISMA clave, y solo un
envío exitoso o cerrar y reabrir el diálogo la renueva. `createBarber`
(`staffApi.ts`) envía esa clave como cabecera `Idempotency-Key`
(`params.header`, no `headers`: forma exigida por `openapi-fetch` para
parámetros tipados del contrato). Un conflicto de idempotencia (`409`) se
expone como `idempotency-conflict`, distinto de `validation-error`: el
formulario no cambió, es un reintento el que necesita una clave nueva.

### Formulario y estados

`StaffPage.vue` usa `BaseDialog` (HU-009) para alta y edición, con
`BaseInput`/`BaseButton`/`BaseAlert` y validación de cliente
(`src/modules/staff/validation/staffValidation.ts`, solo forma: vacío,
solo espacios o más de 120 caracteres; nunca exige dos palabras ni
unicidad). Un error recuperable nunca borra lo escrito; solo la respuesta
confirmada por el servidor cierra el diálogo. La pantalla nunca muestra
placeholders de servicio, horario, disponibilidad, estado activo o usuario
vinculado (`DEC-047`): administra únicamente nombres.

### Hallazgo real de accesibilidad corregido en `BaseDialog` (HU-009)

HU-021 es la primera pantalla que usa `BaseDialog` en un flujo real del
panel. El E2E de evidencia (`e2e/barberos-evidencia-responsiva.spec.ts`)
contra Chromium real (no `jsdom`) expuso dos problemas que las pruebas de
componente existentes no cubrían, corregidos en `src/shared/ui/
BaseDialog.vue` en el mismo cambio:

- **El foco nunca entraba al diálogo al abrirlo.** `trapFocus()` llamaba
  `updateFocusableElements()` de forma síncrona dentro de
  `watch(isOpen, ...)` (flush `pre`, ANTES de que Vue aplicara el `v-show`
  que quita `display: none`); con el diálogo todavía oculto,
  `el.offsetParent !== null` era falso para todo candidato y el bloque
  completo -incluido el enfoque inicial- nunca se ejecutaba. `jsdom` no
  detectaba el defecto (su cálculo de `offsetParent`/reflow es menos
  estricto que un navegador real). Corregido envolviendo TODO el cuerpo de
  `trapFocus` en `nextTick`, no solo la llamada final a `.focus()`.
- **`landmark-unique` (axe-core) real con `AppHeader` (HU-012) en la misma
  página.** El `<header>` de `BaseDialog` sigue resolviendo como landmark
  `banner` incluso dentro de `[role="dialog"]` (la lista de excepciones de
  la spec HTML no incluye ese rol), así que con `AppHeader` ya presente en
  el cascarón privado, dos landmarks `banner` sin nombre único violaban la
  regla. Corregido cambiando ese contenedor de `<header>` a `<div>` (el
  diálogo ya se identifica por su propio `role="dialog"` +
  `aria-labelledby`; el contenedor interno no necesita ser landmark).

Ambos se verificaron con la suite de `BaseDialog.test.ts` (33 pruebas,
sigue en verde) y con el E2E real de HU-021 en Chromium de escritorio y
móvil.

### Pruebas

Componente (`src/modules/staff/pages/__tests__/StaffPage.test.ts`,
incluida la mutación in-place que `push`/asignación por referencia exige
clonar en los fixtures de prueba), cliente tipado (`src/modules/staff/
api/__tests__/staffApi.test.ts`), validación (`src/modules/staff/
validation/__tests__/staffValidation.test.ts`), accesibilidad
(`vitest-axe` integrado en `StaffPage.test.ts`, con `stubs.teleport` para
que las pruebas de `BaseDialog` queden dentro del árbol del wrapper) y E2E
contra el API real en local: `e2e/barberos.spec.ts` (alta de 1 y luego 3
más por el mismo camino de UI -CA-021-01/02-, renombrado con persistencia
real, nombre vacío rechazado, identificador real de otra barbería con
`404` idéntico y ausente del listado propio -CA-021-05-, verificado con
`fetch` autenticado por cookie dentro de la página) y
`e2e/barberos-evidencia-responsiva.spec.ts` (capturas en
`e2e/evidence/barberos/`, axe-core inyectado desde
`node_modules/axe-core/axe.min.js` en cada uno de los cinco breakpoints y
en el diálogo de alta, foco atrapado dentro del diálogo verificado con
`document.activeElement`).

Preparación de PostgreSQL para el E2E real (sin exponer DSN ni secretos):
un contenedor Postgres 14 efímero con las diez migraciones aplicadas vía
Atlas, `testdata/dos_barberias.sql` + `testdata/hu005_credenciales_sesiones.sql`

- `testdata/hu021_barberos.sql` cargados, y un hash argon2id REAL calculado
  en el momento con el mismo `Argon2Hasher` de producción (nunca embebido en
  el repositorio) sustituyendo el hash ficticio de
  `hu005_credenciales_sesiones.sql` para las dos cuentas usadas en el
  recorrido, exactamente el procedimiento que `e2e/configuracion-barberia.spec.ts`
  ya documenta como necesario.

## Catálogo básico de servicios (HU-022)

Fuente normativa: `docs/02-requisitos/historias-usuario.md` (HU-022) y
`docs/00-control/registro-decisiones.md` (`DEC-024`, `DEC-043`, `DEC-067`).
Esta sección resume la API pública y las decisiones de implementación; no
las sustituye.

### Ruta y navegación

`src/modules/catalog/routes.ts` expone `catalogPrivateShellChildRoutes`
(`servicios` → `name: "catalog-servicios"`, `src/modules/catalog/pages/
CatalogPage.vue`, diferida) y `src/modules/catalog/index.ts` expone
`catalogNavItems` (`{ to: { name: "catalog-servicios" }, label: "Servicios" }`).
`src/app/router/index.ts` combina ambas con las de `auth`/`settings`/`staff`
al construir el único cascarón privado; `catalog` no importa ningún interno
de otro módulo ni monta su propio guard.

### Precio como string decimal, nunca `number` (DEC-067)

`model/service.ts` declara `price: string` (nunca `number`): el contrato
expone un decimal exacto con hasta dos cifras
(`api/openapi/components/schemas/ServiceResponse.yaml`) y el formulario
(`validation/catalogValidation.ts`) valida esa misma forma con una
expresión regular antes de enviarla, sin convertir a punto flotante en
ningún punto del cliente. `currency` siempre llega `"COP"`: informativo,
nunca un campo editable (ningún formulario de esta pantalla lo declara).

### Conflicto de nombre distinto de conflicto de idempotencia (DEC-067)

`api/catalogApi.ts` distingue el `409` de nombre duplicado
(`code: "conflict"`) del `409` de idempotencia
(`code: "idempotency-conflict"`/`"idempotency-locked"`) leyendo el campo
`code` que `openapi-fetch` ya decodificó en `error`, nunca comparando
`detail`. `CatalogPage.vue` expone ambos como mensajes recuperables
distintos (`name-conflict` invita a elegir otro nombre; `idempotency-conflict`
invita a reintentar) sin perder lo que el barbero ya escribió.

### Edición parcial de cuatro campos en un solo formulario

A diferencia de `StaffPage.vue` (un único campo editable), el diálogo de
edición de `CatalogPage.vue` muestra los cuatro campos de catálogo
prellenados y los reenvía todos en cada guardado (el backend sí admite un
subconjunto parcial, pero esta pantalla no necesita esa granularidad:
edita el servicio completo de una vez). Una cadena vacía en "Descripción"
borra la descripción existente (equivalente a "sin descripción"), mismo
criterio que el backend.

### Hallazgo real de responsive corregido en `AppNav` (HU-012, compartido)

El E2E de evidencia (`e2e/servicios-evidencia-responsiva.spec.ts`) contra
Chromium real expuso que la cuarta entrada de navegación ("Servicios") ya
no cabía en una sola fila a 320/360 px sin desbordar el documento
(`document.documentElement.scrollWidth > clientWidth`): `AppNav.vue`
(`.app-nav__list`) no tenía manejo alguno para más de tres enlaces en el
ancho más angosto. Se corrigió entonces con `overflow-x: auto` sobre la
propia lista de enlaces (patrón de barra de pestañas desplazable dentro
de sí misma), que sí dejaba pasar la aserción del E2E porque esta solo
mide el desbordamiento de `document.documentElement`, no el de un
contenedor interno.

El QA manual en navegador real de `HU-020`–`HU-024`
(`docs/10-backlog/prompts/test/2026-08-25-qa-manual-b0-b1-chrome-mcp-report.md`)
encontró que ese patrón sí era un defecto real y visible a 320/360 px: el
menú queda con scroll horizontal interno, lo que incumple el criterio
literal de `estandar-diseno-visual.md` §16/§17 ("funciona desde 320 px sin
pérdida") y el ítem A de `docs/07-calidad/02-checklist-transversal.md`
("reflow sin scroll horizontal ni pérdida de contenido"). Se reemplazó por
`flex-wrap: wrap` sobre `.app-nav__list`: a 320/360 px los cinco enlaces
(`Panel`/`Barbería`/`Barberos`/`Servicios`/`Servicios por barbero`) se
reparten en dos líneas sin scroll de ningún tipo, aceptando el costo de un
alto de cabecera variable según cuántos módulos aporte cada barbería (antes
descartado por ese motivo) a cambio de no violar el criterio de "sin
pérdida" con un scroll no descubrible. Reverificado sin regresión con
`panel-evidencia-responsiva.spec.ts` en 320/360/768/1280 px (sin scroll
horizontal del documento, foco visible en teclado).

### Pruebas

Componente (`src/modules/catalog/pages/__tests__/CatalogPage.test.ts`:
carga, alta, edición, conflicto de nombre, conflicto de idempotencia,
doble envío bloqueado, ausencia de placeholders de asignación/activación/
citas, axe-core), cliente tipado
(`src/modules/catalog/api/__tests__/catalogApi.test.ts`: mapeo por
`status`/`code`, nunca `detail`), validación
(`src/modules/catalog/validation/__tests__/catalogValidation.test.ts`) y
E2E contra el API real en local: `e2e/servicios.spec.ts` (alta → listar →
editar → recargar con persistencia real, precio cero y nombre vacío
rechazados sin persistir, duraciones 25/30/45/90 sin catálogo cerrado,
nombre duplicado con `409` sin segunda fila -`DEC-067`-, identificador
real de otra barbería con `404` idéntico y ausente del listado propio) y
`e2e/servicios-evidencia-responsiva.spec.ts` (capturas en
`e2e/evidence/servicios/`, sin scroll horizontal, foco visible y sin
violaciones axe-core en los cinco breakpoints y en el diálogo de alta).

Preparación de PostgreSQL para el E2E real: mismo procedimiento que
"Registro y listado de barberos (HU-021)" arriba (contenedor Postgres 14
efímero, migraciones vía Atlas, `testdata/dos_barberias.sql` +
`testdata/hu005_credenciales_sesiones.sql` + `testdata/hu021_barberos.sql`

- `testdata/hu022_catalogo.sql`, hash argon2id real calculado en el
  momento, nunca embebido en el repositorio).

## Asignación de servicios a barberos (HU-023)

Fuente normativa: `docs/02-requisitos/historias-usuario.md` (HU-023) y
`docs/00-control/registro-decisiones.md` (`DEC-024`, `DEC-068`). Esta
sección resume la API pública y las decisiones de implementación; no las
sustituye.

### Módulo nuevo, sin importar internos de `staff` ni de `catalog`

`src/modules/barberServices/` es un módulo propio (no una extensión de
`staff` ni de `catalog`): la pantalla necesita datos de ambos (barberos y
servicios) pero ningún módulo del frontend expone su cliente API interno a
otro (`staffApi.ts`/`catalogApi.ts` son privados de sus módulos,
docs/03-desarrollo/estandar-frontend-vue.md §3). `api/barberServicesApi.ts`
llama DIRECTAMENTE al cliente HTTP compartido (`@/shared/api/httpClient`)
para `GET /private/barbers`/`GET /private/services` (solo id + nombre
visible, `BarberSummary`/`ServiceSummary`) igual que cualquier otro módulo
llama a ese mismo cliente: es la misma pequeña duplicación de forma que el
backend ya acepta entre `catalog.Cursor`/`staff.Cursor`, preferible a un
acoplamiento cruzado entre módulos. `routes.ts`/`index.ts` siguen el mismo
patrón de ruta hija diferida + `NavItem` que `staff`/`catalog`;
`src/app/router/index.ts` combina las cuatro.

### Casilla controlada: el estado solo cambia tras la respuesta del servidor

`BarberServicesPage.vue` no usa `v-model` sobre cada casilla: usa
`:checked="isAssigned(service.id)"` + `@change`, y en el handler
`onToggleService` guarda una referencia directa al `<input>` que disparó el
evento. Si el servidor rechaza el cambio (`404`, `409` de DEC-068, error de
red), el código fija `checkbox.checked = isAssigned(service.id)`
DIRECTAMENTE sobre ese elemento del DOM, sin depender de que Vue vuelva a
sincronizar la propiedad `checked` por sí solo: cuando el valor reactivo
subyacente no cambió de contenido (el servicio seguía asignado antes y
sigue asignado después de un rechazo), Vue no vuelve a tocar el DOM porque,
desde su óptica, nada cambió -aunque el navegador ya haya alternado la
casilla visualmente al hacer clic-. Cada casilla se deshabilita mientras su
propia solicitud está en curso (`pendingServiceIds`), lo que además evita
estructuralmente el doble envío (un `<input disabled>` no dispara `change`).

### Última asignación activa: mensaje recuperable, sin perder el resto de casillas (DEC-068)

Cuando `unassignService` responde `last-active-conflict`, la pantalla
muestra una alerta explicando que ese barbero es el único asignado al
servicio y revierte solo ESA casilla; las demás conservan su estado
(asignado/pendiente) sin verse afectadas. La selección de barbero tampoco
se pierde: el error es local a la casilla, no a toda la pantalla.

### Pruebas

Componente (`src/modules/barberServices/pages/__tests__/
BarberServicesPage.test.ts`: un barbero y cuatro, servicio compartido,
vacío sin barberos/sin servicios, error recuperable, cambio de barbero,
asignar/desasignar, último activo rechazado -DEC-068- con reversión de la
casilla, error de red, doble envío bloqueado, ausencia de horario/
disponibilidad/citas/precio por barbero, axe-core), cliente tipado
(`src/modules/barberServices/api/__tests__/barberServicesApi.test.ts`:
mapeo por `status`/`code`, nunca `detail`) y E2E contra el API real en
local: `e2e/servicios-por-barbero.spec.ts` (asignar con persistencia real,
un servicio compartido por varios barberos como recursos independientes,
última asignación activa rechazada con la casilla marcada de nuevo,
identificador real de otra barbería con `404`) y
`e2e/servicios-por-barbero-evidencia-responsiva.spec.ts` (capturas en
`e2e/evidence/servicios-por-barbero/`, sin scroll horizontal, foco visible
por teclado -incluida la propia casilla- y sin violaciones axe-core en los
cinco breakpoints).

Preparación de PostgreSQL para el E2E real: mismo procedimiento que
"Catálogo básico de servicios (HU-022)" arriba, con
`testdata/hu023_asignaciones.sql` además de las testdata previas.

## Ciclo de vida de servicios (HU-024)

Fuente normativa: `docs/02-requisitos/historias-usuario.md` (HU-024) y
`docs/00-control/registro-decisiones.md` (`DEC-069`). Esta sección resume la
API pública y las decisiones de implementación; no las sustituye.

### Misma pantalla, sin mezclar edición de catálogo

`CatalogPage.vue` (HU-022) gana dos acciones por servicio -"Desactivar" si
`isActive`, "Reactivar" si no- y un tercer `BaseDialog`, sin tocar los
diálogos de alta/edición existentes: el ciclo de vida nunca comparte
formulario con nombre/descripción/duración/precio (RN-SER-04). Cada
servicio muestra un `BaseBadge` "Activo"/"Inactivo" con texto explícito
(nunca solo color, CA-024-08).

### El diálogo consulta el impacto real antes de confirmar (CA-024-01)

Al abrir el diálogo de desactivación, `openDeactivateDialog` llama
`previewDeactivation(serviceId)` de inmediato: el mensaje de advertencia
("N cita(s) futura(s) se verían afectadas" o "no hay citas futuras") viene
SIEMPRE de esa respuesta real, nunca de un valor por defecto en el cliente.
El botón de confirmar queda deshabilitado mientras la previsualización está
en curso. `deactivateService`/`reactivateService` (POST) llevan su propia
`Idempotency-Key` por intento lógico -mismo criterio que `createService`
(HU-022): se genera al abrir el diálogo y se reutiliza en un reintento por
error de red, nunca en un intento lógico distinto.

### Conflicto de transición: recarga el estado real, nunca éxito optimista

Cuando el servidor responde `409` con `code: conflict` (CA-024-06: el
servicio ya estaba en el estado destino, alguien más lo cambió primero), la
página NO asume que su propio intento tuvo efecto: `reloadAfterConflict`
vuelve a pedir `fetchServices()` y reemplaza la fila con el estado real del
servidor, mostrando una alerta que explica qué pasó. Un `409` de
idempotencia (`idempotency-conflict`/`idempotency-locked`, RN-IDE-01) y un
`404` (servicio ya no existe) tienen su propio mensaje recuperable, sin
tocar la lista.

### Pruebas

Componente (`src/modules/catalog/pages/__tests__/CatalogPage.test.ts`:
badges Activo/Inactivo, previsualización real antes de confirmar, clave de
idempotencia fresca por intento, cancelar no muta nada, conflicto de
transición recarga el estado real, reactivar sin campos de formulario,
axe-core con el diálogo de desactivación abierto), cliente tipado
(`src/modules/catalog/api/__tests__/catalogApi.test.ts`) y E2E contra el
API real en local: `e2e/servicios-ciclo-vida.spec.ts` (previsualizar,
cancelar sin mutar, desactivar con persistencia real tras recargar,
reactivar sin alterar duración, identificador real de otra barbería con
`404` en las tres operaciones), escrito siguiendo el mismo patrón que
`servicios.spec.ts`/`servicios-por-barbero.spec.ts`. Pendiente de
ejecución contra Chromium real y de la evidencia responsiva en
`e2e/evidence/servicios/` en los cinco breakpoints (no forma parte de los
checks de CI, que solo corren `test:unit`).

Preparación de PostgreSQL para el E2E real: mismo procedimiento que
"Catálogo básico de servicios (HU-022)" arriba (sin testdata adicional:
`service.is_active`/`deactivated_at` ya existen desde esa migración).

## Horario laboral recurrente (HU-040)

Fuente normativa: `docs/02-requisitos/historias-usuario.md` (HU-040) y
`docs/00-control/registro-decisiones.md` (`DEC-020`, `DEC-043`, `DEC-070`).
Esta sección resume la API pública y las decisiones de implementación; no
las sustituye.

### Módulo nuevo, mismo patrón de selector que HU-023

`src/modules/schedules/` reutiliza el patrón de selector de barbero de
`barberServices` (`api/schedulesApi.ts` llama DIRECTAMENTE al cliente HTTP
compartido para `GET /private/barbers`, nunca a `staffApi.ts`), pero
agrupa el recurso por día ISO en vez de por casilla: `groupedByWeekday`
siempre construye las siete entradas de `ISO_WEEKDAYS`, con o sin tramos
(`CA-040-01`), así que un barbero recién creado sin horario y uno con
jornada partida en varios días muestran exactamente la misma estructura.

### Alta con `Idempotency-Key`, edición y retiro sin ella

`createWorkingHour` sigue el mismo patrón de intento lógico que
`createBarber`/`createService`: la clave se genera al abrir el diálogo
"Agregar tramo" y se reutiliza en un reintento por error de red, nunca en
un intento lógico distinto. `updateWorkingHour` (reemplaza el intervalo
completo) y `deleteWorkingHour` no llevan `Idempotency-Key` (PATCH/DELETE
sobre un recurso ya identificado por su propio id son naturalmente
repetibles, mismo criterio que `renameBarber`); retirar un tramo ya
retirado responde `404`, que la pantalla trata igual que un éxito (ya no
está, sin mostrar un error nuevo que exija reintento).

### El solape del servidor es la única fuente de verdad

`scheduleValidation.ts` valida forma (día 1-7, hora `HH:MM`, duración
`[1, 1440]`) pero NUNCA solape: esa comprobación depende del estado ya
persistido de los demás tramos, que solo el servidor conoce con certeza
(mismo criterio que `DEC-067`/`DEC-068` en otros módulos). Un `409` con
`code: "conflict"` se traduce a un mensaje recuperable que no cierra el
diálogo ("Este tramo se solapa con otro existente"), distinto del `409` de
idempotencia (`RN-IDE-01`).

### La zona de la barbería se indica explícitamente (CA-040-06)

`fetchBarbershopTimezone` llama a
`GET /private/settings/barbershop` (HU-020) directamente -mismo criterio
de pequeña duplicación entre módulos que `BarberSummary`/`ServiceSummary`
en `barberServices`- y la pantalla muestra "Horas en la zona horaria de la
barbería: {zona}" junto al título. `startsTime` viaja siempre como hora
civil `HH:MM` sin ninguna conversión de zona en ningún punto del cliente;
un fallo al obtener la zona no bloquea la pantalla (la hora ya es correcta
de todas formas), solo omite el indicador.

### Pruebas

Componente (`src/modules/schedules/pages/__tests__/SchedulesPage.test.ts`:
un tramo y jornada partida en varios días con el mismo componente, los
siete días siempre visibles, vacío sin barberos, error recuperable, cambio
de barbero, alta/edición/retiro -éxito, solape, validación de cliente,
error de red-, retiro sin doble envío, indicador de zona horaria,
axe-core), cliente tipado
(`src/modules/schedules/api/__tests__/schedulesApi.test.ts`: mapeo por
`status`/`code`, nunca `detail`) y validación
(`src/modules/schedules/validation/__tests__/scheduleValidation.test.ts`).
E2E escrito siguiendo el mismo patrón que
`servicios-por-barbero.spec.ts`/`servicios-ciclo-vida.spec.ts`:
`e2e/horarios.spec.ts` (alta con persistencia real tras recargar, jornada
partida, solape rechazado sin cerrar el diálogo, editar y retirar con
persistencia real, identificador real de otra barbería con `404`).
Pendiente de ejecución contra Chromium real y de la evidencia responsiva en
`e2e/evidence/horarios/` en los cinco breakpoints (no forma parte de los
checks de CI, que solo corren `test:unit`).

Preparación de PostgreSQL para el E2E real: mismo procedimiento que
"Catálogo básico de servicios (HU-022)" arriba, con
`testdata/hu040_horario.sql` además de las testdata previas.

## Excepciones de jornada y festivos (HU-041)

Fuente normativa: `docs/02-requisitos/historias-usuario.md` (HU-041) y
`docs/00-control/registro-decisiones.md` (`DEC-020`, `DEC-043`,
`DEC-070`). Esta sección resume la API pública y las decisiones de
implementación; no las sustituye.

### Mismo módulo y misma pantalla que HU-040, tres secciones nuevas por barbero

`src/modules/schedules/` gana `api/scheduleExceptionsApi.ts`,
`model/scheduleException.ts`/`exceptionOutcome.ts` y
`validation/exceptionValidation.ts`, todos siguiendo exactamente el
mismo patrón que sus equivalentes de HU-040. `SchedulesPage.vue` no gana
una pantalla nueva: al elegir un barbero, `selectBarber` dispara en
paralelo la carga de su horario semanal (HU-040), su interruptor de
calendario de festivos y sus excepciones de jornada. Tres secciones
nuevas aparecen debajo de los siete días: el interruptor de calendario de
festivos, la lista de excepciones (con alta/edición/retiro) y, si hay
datos, la referencia de "Próximos festivos colombianos" con un botón
"Registrar excepción" que prellena la fecha en el diálogo de alta.

### Forma abierta/cerrada, tramos dinámicos

`ValidateExceptionShape` en `exceptionValidation.ts` refleja la regla del
backend: cerrada nunca admite tramos (`CA-041-04`), abierta exige al
menos uno. Los diálogos de alta/edición modelan `isClosed` como un
`radiogroup` de dos opciones en vez de una casilla booleana sola, para
que la etiqueta describa el resultado ("Cerrado"/"Abierto con tramos
especiales") en lugar de un verbo ambiguo; al pasar a "Abierto" se
agrega automáticamente un primer tramo vacío. Los tramos son una lista
dinámica (`Agregar tramo`/`Quitar`, mismo par `startsTime`/
`durationMinutes` que HU-040), no un único campo fijo: una jornada de
festivo trabajado con un solo tramo reducido y un día especial con
varios tramos usan el mismo formulario.

### El solape y la fecha duplicada del servidor son la única fuente de verdad

Igual que HU-040, `exceptionValidation.ts` valida forma en cliente pero
nunca fecha duplicada (`CA-041-05`) ni solape contra el estado ya
persistido: un `409` con `code: "conflict"` se traduce a
`date-conflict` (mensaje recuperable, "Ya existe una excepción para esa
fecha" / "Ya existe otra excepción para esa fecha"), distinto del `409`
de idempotencia (`RN-IDE-01`) que sigue el mismo criterio que
`createWorkingHour`.

### El calendario de festivos es un interruptor por barbero, no una acción con `Idempotency-Key`

`updateHolidayCalendar` es un `PATCH` naturalmente repetible (mismo
criterio que `renameBarber`/`updateWorkingHour`): no lleva clave de
idempotencia. La casilla revierte a su valor real si la petición falla,
sin optimismo: solo una respuesta `success` cambia el estado visible.

### Festivos colombianos: un dato de referencia que no bloquea la pantalla

`fetchColombianHolidays` es la única llamada del módulo con un outcome de
un solo desenlace de error (`unavailable`, sin distinguir causas): es un
atajo de conveniencia para prellenar una fecha, nunca una condición
bloqueante. La pantalla filtra a fechas futuras (`>= hoy`) y omite la
sección por completo si la consulta falla o no hay festivos próximos.

### Pruebas

Componente (`src/modules/schedules/pages/__tests__/SchedulesPage.test.ts`,
sección "SchedulesPage · HU-041": interruptor de festivos reflejando su
estado real y su edición, mensaje recuperable en un fallo de red, listado
de excepciones, alta de una excepción cerrada, fecha duplicada sin cerrar
el diálogo, validación de cliente sin llamar al API, retiro, festivos
colombianos próximos con prellenado del diálogo de alta), cliente tipado
(`src/modules/schedules/api/__tests__/scheduleExceptionsApi.test.ts`:
mapeo por `status`/`code`, nunca `detail`) y validación
(`src/modules/schedules/validation/__tests__/exceptionValidation.test.ts`).
E2E escrito siguiendo el mismo patrón que `horarios.spec.ts`:
`e2e/excepciones-festivos.spec.ts` (activar el calendario de festivos con
persistencia real, alta cerrada y abierta con persistencia real, fecha
duplicada rechazada sin cerrar el diálogo, editar y retirar con
persistencia real, identificador real de otra barbería con `404` en el
calendario de festivos). Pendiente de ejecución contra Chromium real y de
la evidencia responsiva en `e2e/evidence/excepciones-festivos/` en los
cinco breakpoints (no forma parte de los checks de CI, que solo corren
`test:unit`), mismo estado que `horarios.spec.ts` (HU-040).

Preparación de PostgreSQL para el E2E real: mismo procedimiento que
"Horario laboral recurrente (HU-040)" arriba, con
`testdata/hu041_excepciones.sql` además de las testdata previas.

## Nuevo turno (HU-061)

Primera pantalla real del módulo `agenda` (`src/modules/agenda`):
`NewAppointmentPage.vue` (`/panel/turnos/nuevo`), el barbero autenticado
registra un turno manual recibido por teléfono, WhatsApp o en persona.
Elegir un barbero carga solo los servicios activos ya asignados a ese
barbero (`GET .../services` de HU-023 cruzado contra el catálogo activo de
HU-022, `DEC-072`): el formulario nunca ofrece un servicio que el
servidor rechazaría. `startsAt` se envía como fecha/hora civil
"AAAA-MM-DDTHH:MM:SS" sin desplazamiento de zona (los `<input
type="date">`/`type="time">` nativos ya la producen en el reloj del
dispositivo, que es el que el barbero ve): el servidor la interpreta
contra la zona IANA de la barbería (RN-DIS-07), sin ninguna conversión de
zona en el cliente. El éxito muestra un resumen honesto (servicio,
duración, precio) sin navegar automáticamente a la agenda: el barbero
vuelve a "Panel" (HU-062) por su cuenta.

### Pruebas

Componente
(`src/modules/agenda/pages/__tests__/NewAppointmentPage.test.ts`): carga
(vacío, error recuperable), selección de barbero → carga de servicios
asignados (`DEC-072`), validación de forma bloqueando el envío,
conflicto de agenda/bloqueo (`DEC-073`) mostrando el detalle del
servidor, conflicto de idempotencia, error de red conservando los datos
ya escritos, doble envío bloqueado y `vitest-axe` sin violaciones.
E2E escrito siguiendo el mismo patrón que
`e2e/servicios-por-barbero.spec.ts`: `e2e/nuevo-turno.spec.ts` (alta
manual dentro de la próxima hora, fuera de la anticipación mínima
pública; cruce con una cita existente del mismo barbero rechazado;
servicio no asignado ausente del selector). Pendiente de ejecución
contra Chromium real y de evidencia responsiva en los cuatro
breakpoints (no forma parte de los checks de CI, que solo corren
`test:unit`), mismo estado que `e2e/horarios.spec.ts` (HU-040) y
`e2e/excepciones-festivos.spec.ts` (HU-041).

## Agenda diaria de hoy (HU-062)

Reemplaza el marcador de posición de HU-012 en la raíz del panel privado:
`DailyAgendaPage.vue` (`/panel`, `name: 'panel'`) ahora vive en el módulo
`agenda` (`src/modules/agenda/routes.ts`), no en `auth`
(`auth/pages/PanelPage.vue` retirado, `auth/routes.ts` ya no registra
ninguna hija propia — `privateShellChildRoutes` queda `[]`). Selector
obligatorio de un barbero (`DEC-074`): la pantalla nunca ofrece una vista
consolidada de varios barberos ni infiere uno del `staff_user` autenticado.
Al elegir un barbero, `fetchDailyAgenda` pide
`GET /private/barbers/{barberId}/appointments/daily-agenda` con `date`
(desde HU-063, casi siempre un día civil explícito en vez de dejar que el
servidor decida "hoy"; ver la sección "Navegación de agenda por fecha
(HU-063)" más abajo). El encabezado muestra la fecha civil completa y la
zona con `formatCivilDateFull`/`formatTimeInTimezone`
(`src/shared/time/civilDate.ts`/`formatInstant.ts`, mismo criterio de zona
explícita obligatoria que `formatInstantInTimezone`, HU-020). Cada fila usa
`BaseBadge` con las clases de estado ya definidas en el sistema visual
(`base-badge--status-*`, sin usar hasta esta historia) y la etiqueta en
español de `APPOINTMENT_STATUS_LABELS`
(`src/modules/agenda/model/dailyAgenda.ts`): nunca solo color. Una cita
terminal se atenúa (`opacity`) pero nunca desaparece de la lista
(`RN-CIT-04`). "Nuevo turno" navega al formulario real de HU-061
(`router.push`, sin `RouterLink` porque `BaseButton` no admite un
destino). Cambiar de barbero descarta una respuesta de agenda que llegue
tarde para una selección ya reemplazada (mismo patrón que
`SchedulesPage.vue`/`selectBarber`, HU-040).

Fuera de alcance a propósito: detalle de cita y cualquier acción sobre una
cita existente. Anterior/siguiente/selector de fecha (`F-CITA-02`) llegó
con HU-063 (siguiente sección).

### Pruebas

Componente (`src/modules/agenda/pages/__tests__/DailyAgendaPage.test.ts`):
carga (vacío, error recuperable), selector obligatorio de barbero
(`DEC-074`), agenda vacía con mensaje explícito, aislamiento entre
selecciones rápidas de barbero (descarta una respuesta obsoleta), error
recuperable con reintento, barbero ya no disponible (404), navegación a
"Nuevo turno" y `vitest-axe` sin violaciones. `src/shared/time/formatInstant.ts`
gana `formatTimeInTimezone`/`formatFullDateInTimezone`, probadas en
`src/shared/time/__tests__/formatInstant.test.ts` con el mismo control
negativo de zona explícita que `formatInstantInTimezone` (HU-020).
`src/modules/auth/__tests__/routes.test.ts` se actualizó: la prueba del
guard del cascarón ya no depende de qué módulo sirve `/panel`, usa una
hija mínima propia en vez de `auth/pages/PanelPage.vue` (retirado).

E2E: `e2e/agenda-diaria.spec.ts` (apertura mostrando fecha completa y
zona con selector obligatorio, turno recién creado visible con estado
Confirmado, cambio de barbero muestra la agenda correcta, navegación a
"Nuevo turno", y un dispositivo en otra zona horaria — `timezoneId:
'America/Los_Angeles'` — mostrando la zona de la barbería, nunca la del
dispositivo). `e2e/panel.spec.ts`/`e2e/panel-evidencia-responsiva.spec.ts`
(HU-012) se actualizaron: el encabezado esperado pasó de "Panel del
barbero" a "Agenda de hoy". Pendiente de ejecución contra Chromium real
y de evidencia responsiva en los cuatro breakpoints (no forma parte de
los checks de CI, que solo corren `test:unit`), mismo estado que
`e2e/nuevo-turno.spec.ts` (HU-061).

## Navegación de agenda por fecha (HU-063)

Tres controles estables en `DailyAgendaPage.vue` (día anterior, selector de
fecha `BaseInput type="date"`, día siguiente; `estandar-diseno-visual.md`
§11.2): la fuente de verdad es `route.query` (`date` en `AAAA-MM-DD`,
`barberId`), nunca un store global. `src/shared/time/civilDate.ts` aporta
la aritmética de calendario que lo hace posible, siempre sin depender de la
zona del dispositivo (`RN-DIS-07`):

- `getCivilDateInTimezone(timezone, at?)`: día civil vigente AHORA en la
  zona explícita recibida (`Intl.DateTimeFormat('en-CA', …)`), correcto en
  un día local de 23/25 horas por DST.
- `shiftCivilDate(civilDate, deltaDays)`: aritmética pura en `Date.UTC`
  sobre el triplete año/mes/día, nunca suma 24h de reloj; cruza mes/año y
  es indiferente a DST porque la zona de la barbería no participa.
- `formatCivilDateFull(civilDate)`: formatea una fecha civil YA resuelta
  ("jueves, 31 de agosto de 2026") anclando a mediodía UTC y formateando en
  `timeZone: 'UTC'`, para que ninguna zona (ni siquiera un desfase extremo
  como UTC+14) pueda desplazar el día mostrado.

`syncFromRoute()` (`DailyAgendaPage.vue`) es el único punto que traduce
`route.query` a una carga de agenda: se invoca una vez al abrir la pantalla
(tras resolver barberos/zona) y en cada cambio posterior de `route.query`
vía un `watch`, cubriendo de forma uniforme recarga, atrás/adelante del
navegador y los `router.push` propios de esta pantalla (cambiar de fecha o
de barbero). Si la URL no trae una selección válida (`barberId`
inexistente, `date` ausente o mal formada), la normaliza con un único
`router.replace` — nunca crea una entrada de historial por el simple hecho
de abrir `/panel` — que vuelve a disparar el mismo `watch` ya con la URL
normalizada; nunca hay una segunda petición duplicada. Un contador
`agendaRequestSeq` descarta cualquier respuesta que llegue fuera de orden
(cambio rápido de fecha o de barbero), igual que `selectBarber` ya hacía en
HU-062. Sin zona de la barbería conocida (`fetchBarbershopTimezone`
`unavailable`), la navegación por fecha se deshabilita en vez de adivinar
con la zona del dispositivo; la agenda sigue siendo legible con el
comportamiento degradado de HU-062 (`date` ausente, "hoy" según el
servidor).

`agendaStatus` distingue `loading` (primera carga, sin nada que conservar)
de `updating` (ya hubo una agenda confirmada — para otra fecha o barbero —
que permanece visible con un indicador local "Actualizando…" mientras
llega la siguiente, `estandar-diseno-visual.md` §9 "Actualización"). Un
error tras una carga previa exitosa conserva esa agenda visible junto a la
alerta de reintento, en vez de reemplazar toda la pantalla.

Fuera de alcance a propósito: detalle, historial y cualquier acción sobre
una cita (ver "Detalle e historial de un turno (HU-064)" más abajo, y
HU-065 para reprogramación); vista semanal/mensual; vista consolidada de
varios barberos.

### Pruebas

`src/shared/time/__tests__/civilDate.test.ts`: fin de mes/año, año
bisiesto, aritmética estable alrededor de una transición DST y lectura del
día civil en dos zonas con desfase suficiente para caer en días distintos.

Componente (`src/modules/agenda/pages/__tests__/DailyAgendaPage.test.ts`,
bloque "navegación por fecha (HU-063)"): anterior/siguiente/selector de
fecha, recarga con `date`/`barberId` ya en la URL, atrás del navegador,
descarte de una respuesta fuera de orden tras navegar dos veces seguidas,
indicador "Actualizando…" con la agenda anterior visible, y navegación
deshabilitada sin zona de la barbería conocida.

E2E: `e2e/agenda-navegacion-fecha.spec.ts` (hoy → anterior → siguiente →
fecha elegida → recarga → atrás/adelante, con fecha y barbero correctos en
cada paso; un dispositivo en otra zona horaria calcula igual el día civil
de la zona de la barbería). Pendiente de ejecución contra Chromium real y
de evidencia responsiva en los cuatro breakpoints (no forma parte de los
checks de CI, que solo corren `test:unit`), mismo estado que
`e2e/agenda-diaria.spec.ts` (HU-062).

## Detalle e historial de un turno (HU-064)

`AppointmentDetailPage.vue` (ruta lazy `turnos/:appointmentId`, name
`agenda-detalle-turno`, hija de `agenda`, registrada después de
`turnos/nuevo`) abre desde cada fila de `DailyAgendaPage.vue`: el bloque
`daily-agenda-page__item-main` (hora/persona/servicio) es un `RouterLink`
hacia el detalle con `query: withQuery({})` (reenvía `date`/`barberId`
vigentes sin tocarlos); el badge de estado queda fuera del enlace a
propósito, para no convertir toda la fila en un control ambiguo.

La pantalla sigue la plantilla P0 "Detalle de turno"
(`estandar-diseno-visual.md` §10): hora con zona, persona atendida,
servicio (snapshot, nunca el catálogo vigente), barbero, cliente que
reservó con su contacto opcional y nota (solo visibles aquí, nunca en la
agenda diaria), estado e historial. Ninguna acción se renderiza (edición,
reprogramación, cancelación): `HU-064` es de solo lectura. "Volver a la
agenda" (`RouterLink`, mismo patrón que `RecoveryPage.vue`) usa
`backQuery`, capturado UNA sola vez al montar desde `route.query.date`/
`route.query.barberId`, para regresar exactamente a la fecha/barbero de
origen (`HU-063`) sin depender del historial del navegador.

`pageStatus` (`loading`/`ready`/`not-found`/`error`) gobierna el detalle;
`historyStatus` (`loading`/`ready`/`error`) gobierna el historial por
separado, con su propio `historyLoadingMore` para el botón "Cargar más" —
un error de historial nunca oculta un detalle ya cargado, ni viceversa.
`loadHistory(cursor?)` reemplaza `historyItems` en la primera página y
concatena en las siguientes, nunca al revés. `actorLabel` de cada entrada
llega ya resuelto y seguro desde el servidor (nunca correo ni un
identificador interno, `RN-HIS-01`); `HISTORY_EVENT_LABELS`/
`historyFieldLabel` (`model/appointmentDetail.ts`) traducen el vocabulario
técnico cerrado (ocho eventos, `DEC-041`) y los campos de
`appointment_history_change` al español, con una etiqueta de reserva para
cualquier campo no mapeado explícitamente.

### Pruebas

Componente (`src/modules/agenda/pages/__tests__/AppointmentDetailPage.test.ts`):
detalle listo con todos los campos (y sin sección "Contacto" cuando no hay
teléfono ni correo), no encontrado, error con reintento, historial listo
tras el detalle, motivo y cambios anterior/nuevo de una entrada, "Cargar
más" con dos páginas que conviven (nunca se reemplaza la primera), error de
historial con reintento, enlace "Volver" con `href` verificado
literalmente (fecha/barbero preservados), y `vitest-axe` sin violaciones.

E2E: `e2e/agenda-detalle-historial.spec.ts` (desde una fecha no actual,
abre el turno, verifica snapshots/contacto/nota/historial, vuelve a la
misma fecha/barbero; un `appointmentId` inexistente muestra el estado "no
disponible" explícito). Pendiente de ejecución contra Chromium real y de
evidencia responsiva en los cuatro breakpoints (no forma parte de los
checks de CI, que solo corren `test:unit`), mismo estado que
`e2e/agenda-navegacion-fecha.spec.ts` (HU-063).

## Sistema visual base (HU-009)

Fuente normativa completa:
[`docs/03-desarrollo/estandar-diseno-visual.md`](../../docs/03-desarrollo/estandar-diseno-visual.md)
(`DEC-039`). Esta sección resume la API pública de los cinco componentes
que ya existen; no la sustituye.

### Regla que gobierna a los cinco

Toda decisión visual sale de un token semántico de `tokens.css`
(`--color-*`, `--space-*`, `--radius-*`, `--motion-*`…), nunca de un
hexadecimal, píxel o milisegundo suelto dentro del componente. Ninguno
declara `class`, `color` ni `zIndex` como prop: `class` se hereda por el
fallthrough automático de Vue hacia el elemento raíz cuando el consumidor
lo necesita para layout; la apariencia semántica solo se elige con `variant`
(o `tone`, según el componente).

### BaseButton

```vue
<BaseButton variant="primary" size="lg" :loading="isSubmitting" @click="onSubmit">
  Confirmar turno
</BaseButton>
```

- `variant`: `primary` | `secondary` | `soft` | `ghost` | `danger`.
- `size`: `md` (44px, por defecto) | `lg` (48px, acción principal móvil).
  No existe `sm`: el estándar visual no define un botón por debajo del
  objetivo táctil de 44 px.
- `disabled`, `loading` (bloquea la interacción y evita doble envío sin
  sustituir la idempotencia del servidor), `pressed` (toggle), `type`.
- Evento `click`. El spinner de `loading` es un indicador de progreso
  indeterminado: se apaga con `prefers-reduced-motion`, pero no cuenta como
  la transición de 120–200 ms de `CA-009-07` (esa regla es para cambios de
  estado discretos, no para "trabajo en curso" de duración desconocida).

### BaseInput

```vue
<BaseInput v-model="email" type="email" label="Correo" :error="errors.email" required />
```

- `modelValue`, `type`, `label`, `hint`, `error` (sustituye a `hint` y
  activa `aria-invalid`), `placeholder` (nunca reemplaza a `label`),
  `disabled`, `readonly`, `required`, `autocomplete`, `name`, `id`
  (se genera uno estable con `useId()` si se omite), y los atributos de
  validación nativa (`pattern`, `minlength`, `maxlength`, `min`, `max`,
  `step`).
- Eventos: `update:modelValue`, `input`, `change`, `blur`, `focus`.
- Slots `leading`/`trailing`: solo decorativos (el wrapper lleva
  `aria-hidden`); un control interactivo ahí quedaría oculto para
  tecnología de asistencia pese a seguir siendo alcanzable con Tab.

### BaseAlert

```vue
<BaseAlert variant="danger" title="No se pudo confirmar el turno" dismissible @dismiss="onDismiss">
  Intenta de nuevo en unos segundos.
</BaseAlert>
```

- `variant`: `success` | `warning` | `danger` | `info` | `neutral`.
- `title`, `dismissible`, `role` (`alert` por defecto = `aria-live="assertive"`;
  `status` = `polite`).
- Slots `default` (cuerpo) y `action` (una acción de recuperación).
- Eventos `dismiss`, `action`.

### BaseBadge

```vue
<BaseBadge variant="info">Confirmada</BaseBadge>
<BaseBadge variant="neutral" dismissible label="Filtro barbero" @dismiss="onRemove">Juan</BaseBadge>
```

- `variant`: `neutral` | `primary` | `success` | `warning` | `danger` | `info`.
- `size`: `sm` | `md`. `dot` (punto de estado, color siempre derivado de
  `variant` — no existe un prop de color libre). `dismissible`, `label`
  (obligatorio en la práctica cuando el contenido no es texto legible por
  sí solo).
- El botón de cierre mide 24×24 px (el suelo de WCAG 2.2 SC 2.5.8), no
  44×44: una insignia es una etiqueta compacta en línea con el texto que
  la rodea, la excepción "Inline" de esa misma regla.

### BaseDialog

```vue
<BaseDialog v-model="open" title="Cancelar turno" description="Esta acción no se puede deshacer.">
  <p>¿Seguro que quieres cancelar?</p>
  <template #footer>
    <BaseButton variant="secondary" @click="open = false">Volver</BaseButton>
    <BaseButton variant="danger" @click="confirmCancel">Cancelar turno</BaseButton>
  </template>
</BaseDialog>
```

- `modelValue` (v-model), `title`, `description`, `size`
  (`sm` | `md` | `lg` | `full`), `closeOnBackdrop`, `closeOnEscape`,
  `showClose`.
- Slots `header` (reemplaza título + botón de cerrar), `default`, `footer`.
- Eventos `update:modelValue`, `close`, `open`. Métodos expuestos `open()`/`close()`.
- Gestiona foco inicial, trampa de foco, retorno de foco al disparador y
  bloqueo de scroll del body — todo contado por una pila compartida entre
  instancias (`openDialogStack`, exportado desde el propio `.vue`), así que
  dos diálogos abiertos a la vez no compiten por la misma tecla Escape ni
  desbloquean el scroll el uno al otro. No admite un prop `zIndex`: todos
  comparten la capa semántica `--layer-dialog`.

## Requisitos

- Node.js 20 o superior.
- pnpm.

## Comandos

```bash
pnpm install
pnpm dev
pnpm build
pnpm typecheck
pnpm lint
pnpm format
pnpm test:unit
pnpm test:e2e
pnpm generate:api   # regenera shared/api/generated/openapi.d.ts desde el bundle
```

`pnpm dev` usa el proxy de `vite.config.ts` (`/api` → `http://localhost:8080`
por defecto, configurable con `API_PROXY_TARGET`) para que `shared/api` use
rutas relativas de mismo origen, igual que en despliegue real. Pinia se
agrega solo si aparece estado compartido real entre rutas, según `DEC-033`.
