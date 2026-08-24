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
distinguido del conflicto de idempotencia. Ver las secciones siguientes.
El resto de carpetas de `modules/` conserva su `index.ts` de marcador de
responsabilidad futura.

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
ancho más angosto. Corregido con `overflow-x: auto` sobre la propia lista
de enlaces (patrón de barra de pestañas desplazable, en vez de envolver a
varias líneas, que habría cambiado el ritmo vertical del cascarón).
Verificado sin regresión contra las evidencias responsivas ya integradas
de `HU-011`/`HU-012`/`HU-020`/`HU-021`.

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
