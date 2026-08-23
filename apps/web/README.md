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
la pantalla de acceso cuando el login responde 429, y la **configuración
básica de la barbería (HU-020)**: pantalla "Barbería" (`/panel/barberia`)
para consultar y actualizar nombre, zona horaria y contacto opcional, con
actualización inmediata de la cabecera tras guardar. Ver las secciones
siguientes. El resto de carpetas de `modules/` conserva su `index.ts` de
marcador de responsabilidad futura.

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
