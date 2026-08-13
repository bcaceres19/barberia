# apps/web

Aplicación Vue 3 + TypeScript + Vite del flujo público y del panel del
barbero. Ver
[`docs/04-arquitectura/frontend.md`](../../docs/04-arquitectura/frontend.md)
y [`docs/03-desarrollo/estandar-frontend-vue.md`](../../docs/03-desarrollo/estandar-frontend-vue.md)
antes de agregar código.

## Estado

Arranque de la aplicación, router con una pantalla provisional, y el
**sistema visual base (HU-009)**: tokens (`src/styles/tokens.css`,
`src/styles/base.css`) y cinco componentes en `src/shared/ui/` — `BaseButton`,
`BaseInput`, `BaseAlert`, `BaseBadge`, `BaseDialog`. Ver la sección
siguiente. Las carpetas de `modules/` conservan su `index.ts` de marcador de
responsabilidad futura; no hay componentes de negocio, cliente API ni
pantallas reales todavía.

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
```

`shared/api/generated` se agrega cuando exista un bundle OpenAPI que
generar; Pinia se agrega solo si aparece estado compartido real entre rutas,
según `DEC-033`.
