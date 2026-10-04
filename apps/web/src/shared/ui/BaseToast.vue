<script setup lang="ts">
/**
 * BaseToast - Aviso emergente tipo acordeón (DEC-095, estandar-diseno-visual.md §6.8).
 * Superficie tinta con sello de estado: cerrado muestra la palabra de estado y
 * el título; abierto añade el detalle, la referencia y las acciones. Es
 * presentacional: el tiempo en pantalla, la cola y el acordeón entre avisos
 * los gobierna `shared/model/toastStore`, y `ToastRegion` los conecta.
 *
 * Accesibilidad: `role="alert"` solo para `danger`, `role="status"` para el
 * resto; nunca roba el foco. El detalle cerrado se anuncia mediante un texto
 * solo para lector de pantalla, porque el panel oculto no entra en el árbol
 * accesible. La palabra de estado es la señal además del color (WCAG 2.2 AA
 * 1.4.1) y la barra inferior solo es un apoyo visual del tiempo restante.
 */
import { computed, ref, useId } from 'vue'
import type { ToastActionIcon, ToastVariant } from '@/shared/model/toastStore'

interface Props {
  variant: ToastVariant
  title: string
  detail?: string
  /** Referencia segura para soporte (por ejemplo `request_id`). */
  reference?: string
  actionLabel?: string
  actionIcon?: ToastActionIcon
  /** Panel abierto (acordeón). */
  expanded: boolean
  /** Cuenta atrás detenida: abierto, con cursor encima o con foco. */
  held: boolean
  durationMs: number
}

const props = withDefaults(defineProps<Props>(), {
  detail: '',
  reference: '',
  actionLabel: '',
  actionIcon: 'arrow',
})

const emit = defineEmits<{
  toggle: []
  dismiss: []
  action: []
  hover: [value: boolean]
  focus: [value: boolean]
}>()

const rootRef = ref<HTMLElement | null>(null)
const panelId = `base-toast-panel-${useId()}`

// Palabra de estado: mismo vocabulario que BaseAlert (issue #212).
const STATUS_WORD: Record<ToastVariant, string> = {
  success: 'Confirmación',
  info: 'Nota',
  warning: 'Atención',
  danger: 'Error',
}

// Trazos del sello (viewBox 24): check, signo de información, exclamación y cruz.
const MARK_PATH: Record<ToastVariant, string> = {
  success: 'M5 12.5l4.5 4.5L19 7.5',
  info: 'M12 10.5v7M12 6.5v.01',
  warning: 'M12 5.5v8M12 18.5v.01',
  danger: 'M6.5 6.5l11 11M17.5 6.5l-11 11',
}

const ACTION_PATH: Record<ToastActionIcon, string> = {
  retry: 'M20 12a8 8 0 1 1-2.34-5.66M20 4v5h-5',
  arrow: 'M5 12h14M13 6l6 6-6 6',
}

const role = computed(() => (props.variant === 'danger' ? 'alert' : 'status'))

const fillStyle = computed(() => ({
  animationDuration: `${props.durationMs}ms`,
  animationPlayState: props.held ? 'paused' : 'running',
}))

// Solo el foco de teclado detiene la cuenta atrás: un clic de ratón deja el
// foco en el botón, y con eso el aviso no volvería a cerrarse solo. Si el
// navegador no distingue `:focus-visible`, se trata como foco de teclado.
function onFocusIn(event: FocusEvent) {
  let keyboard = true
  try {
    keyboard = (event.target as HTMLElement).matches(':focus-visible')
  } catch {
    keyboard = true
  }
  if (keyboard) emit('focus', true)
}

// El foco pasa entre los botones del propio aviso sin que cuente como salir.
function onFocusOut(event: FocusEvent) {
  const next = event.relatedTarget as Node | null
  if (next && rootRef.value?.contains(next)) return
  emit('focus', false)
}
</script>

<template>
  <div
    ref="rootRef"
    :class="['base-toast', `base-toast--${variant}`]"
    :role="role"
    @mouseenter="emit('hover', true)"
    @mouseleave="emit('hover', false)"
    @focusin="onFocusIn"
    @focusout="onFocusOut"
  >
    <button
      type="button"
      class="base-toast__head"
      :aria-expanded="expanded"
      :aria-controls="panelId"
      @click="emit('toggle')"
    >
      <span class="base-toast__mark" aria-hidden="true">
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2.2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path :d="MARK_PATH[variant]" />
        </svg>
      </span>
      <span class="base-toast__text">
        <span class="base-toast__status">{{ STATUS_WORD[variant] }}</span>
        <span class="base-toast__title">{{ title }}</span>
      </span>
      <span
        :class="['base-toast__chevron', { 'base-toast__chevron--open': expanded }]"
        aria-hidden="true"
      >
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <polyline points="6 9 12 15 18 9" />
        </svg>
      </span>
    </button>

    <span v-if="detail && !expanded" class="base-toast__sr">{{ detail }}</span>

    <div :id="panelId" :class="['base-toast__panel', { 'base-toast__panel--open': expanded }]">
      <div class="base-toast__panel-clip">
        <div class="base-toast__body">
          <p v-if="detail" class="base-toast__detail">{{ detail }}</p>
          <p v-if="reference" class="base-toast__reference">ID de solicitud · {{ reference }}</p>
          <div class="base-toast__actions">
            <button
              v-if="actionLabel"
              type="button"
              class="base-toast__action"
              @click="emit('action')"
            >
              <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2.2"
                stroke-linecap="round"
                stroke-linejoin="round"
                aria-hidden="true"
              >
                <path :d="ACTION_PATH[actionIcon]" />
              </svg>
              <span>{{ actionLabel }}</span>
            </button>
            <button type="button" class="base-toast__dismiss" @click="emit('dismiss')">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2.4"
                stroke-linecap="round"
                aria-hidden="true"
              >
                <path d="M6 6l12 12M18 6L6 18" />
              </svg>
              <span>Descartar</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="base-toast__bar" aria-hidden="true">
      <div class="base-toast__fill" :style="fillStyle" />
    </div>
  </div>
</template>

<style scoped>
/* Etiqueta de sastre sobre tinta (DEC-095). El filete perimetral es el latón
   de baja opacidad del header del shell: separa el aviso tanto de las
   pantallas claras como de las que ya están sobre tinta (agenda). */
.base-toast {
  --toast-accent: var(--color-success-on-strong);
  --toast-radius: 3px;
  --toast-hairline: color-mix(in srgb, var(--color-on-strong) 20%, transparent);

  position: relative;
  overflow: hidden;
  background-color: var(--color-surface-strong);
  color: var(--color-on-strong);
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent);
  border-radius: var(--toast-radius);
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--color-on-strong) 8%, transparent),
    0 2px 6px rgb(16 27 43 / 20%),
    0 22px 44px -14px rgb(16 27 43 / 58%);
  font-family: var(--font-sans);
  animation: base-toast-in 160ms ease-out;
}

.base-toast--info {
  --toast-accent: var(--color-info-on-strong);
}

.base-toast--warning {
  --toast-accent: var(--color-warning-on-strong);
}

.base-toast--danger {
  --toast-accent: var(--color-danger-on-strong);
}

@keyframes base-toast-in {
  from {
    opacity: 0;
    transform: translateY(-6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.base-toast__head {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  width: 100%;
  min-height: 64px;
  padding: var(--space-3) var(--space-1) var(--space-3) 14px;
  background: transparent;
  border: 0;
  color: inherit;
  text-align: left;
  cursor: pointer;
  transition: background-color var(--motion-duration-fast) var(--motion-easing-standard);
}

.base-toast__head:hover {
  background-color: color-mix(in srgb, var(--color-on-strong) 5%, transparent);
}

.base-toast__mark {
  display: flex;
  flex: none;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  color: var(--toast-accent);
  background-color: color-mix(in srgb, var(--color-on-strong) 5%, transparent);
  border: var(--border-width-normal) solid var(--toast-accent);
  border-radius: 2px;
}

.base-toast__text {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: var(--space-1);
  min-width: 0;
}

.base-toast__status {
  font-size: var(--font-size-caption);
  font-weight: 600;
  line-height: 12px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--color-brand-accent-surface);
}

.base-toast__title {
  font-size: 15px;
  font-weight: 600;
  line-height: 20px;
  color: var(--color-on-strong);
}

.base-toast__chevron {
  display: flex;
  flex: none;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: var(--control-height-icon);
  color: var(--color-on-strong-muted);
  transition:
    transform var(--motion-duration-base) var(--motion-easing-standard),
    color var(--motion-duration-fast) var(--motion-easing-standard);
}

.base-toast__head:hover .base-toast__chevron {
  color: var(--color-on-strong);
}

.base-toast__chevron--open {
  transform: rotate(180deg);
}

/* Solo lector de pantalla: anuncia el detalle mientras el panel está cerrado. */
.base-toast__sr {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}

/* Acordeón: la fila de la cuadrícula pasa de 0fr a 1fr. `visibility` quita el
   panel cerrado del orden de tabulación sin cortar la animación de cierre. */
.base-toast__panel {
  display: grid;
  grid-template-rows: 0fr;
  visibility: hidden;
  transition:
    grid-template-rows var(--motion-duration-base) var(--motion-easing-standard),
    visibility 0s linear var(--motion-duration-base);
}

.base-toast__panel--open {
  grid-template-rows: 1fr;
  visibility: visible;
  transition-delay: 0s;
}

.base-toast__panel-clip {
  min-height: 0;
  overflow: hidden;
}

/* Línea punteada tipo etiqueta; el texto se alinea bajo el título. */
.base-toast__body {
  margin: 0 14px;
  padding: var(--space-3) 0 var(--space-2) 44px;
  border-top: var(--border-width-normal) dashed var(--toast-hairline);
}

.base-toast__detail {
  margin: 0;
  font-size: var(--font-size-body-sm);
  line-height: 21px;
  color: var(--color-on-strong-soft);
}

.base-toast__reference {
  margin: 10px 0 0;
  font-size: var(--font-size-caption);
  line-height: 16px;
  letter-spacing: 0.04em;
  color: var(--color-on-strong-muted);
  font-variant-numeric: tabular-nums;
}

.base-toast__actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-top: 10px;
}

/* Los botones miden 40px y su zona activa se extiende a 44px (`::after`),
   el objetivo táctil de estandar-diseno-visual.md §8. */
.base-toast__action,
.base-toast__dismiss {
  position: relative;
  display: inline-flex;
  align-items: center;
  height: 40px;
  border-radius: 2px;
  cursor: pointer;
  transition:
    background-color var(--motion-duration-fast) var(--motion-easing-standard),
    border-color var(--motion-duration-fast) var(--motion-easing-standard),
    color var(--motion-duration-fast) var(--motion-easing-standard);
}

.base-toast__action::after,
.base-toast__dismiss::after {
  position: absolute;
  top: -2px;
  right: 0;
  bottom: -2px;
  left: 0;
  content: '';
}

.base-toast__action {
  gap: var(--space-2);
  padding: 0 14px 0 var(--space-3);
  background-color: color-mix(in srgb, var(--color-on-strong) 6%, transparent);
  border: var(--border-width-normal) solid var(--toast-accent);
  color: var(--color-on-strong);
  font-size: var(--font-size-body-sm);
  font-weight: 600;
}

.base-toast__action svg {
  flex: none;
  color: var(--toast-accent);
}

.base-toast__action:hover {
  background-color: color-mix(in srgb, var(--color-on-strong) 14%, transparent);
  border-color: var(--color-on-strong);
}

.base-toast__action:active {
  background-color: color-mix(in srgb, var(--color-on-strong) 20%, transparent);
}

.base-toast__dismiss {
  gap: 6px;
  margin-left: auto;
  padding: 0 10px;
  background: transparent;
  border: 0;
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.base-toast__dismiss:hover {
  background-color: color-mix(in srgb, var(--color-on-strong) 8%, transparent);
  color: var(--color-on-strong);
}

.base-toast button:focus-visible {
  outline: 2px solid var(--color-on-strong);
  outline-offset: 2px;
}

.base-toast__head:focus-visible {
  outline-offset: -4px;
}

/* Regla del tiempo restante: riel tenue con relleno del color del estado. */
.base-toast__bar {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: 2px;
  background-color: color-mix(in srgb, var(--color-on-strong) 10%, transparent);
}

.base-toast__fill {
  height: 100%;
  background-color: var(--toast-accent);
  transform-origin: left center;
  animation: base-toast-drain linear forwards;
}

@keyframes base-toast-drain {
  from {
    transform: scaleX(1);
  }
  to {
    transform: scaleX(0);
  }
}

/* Sin traslación ni transiciones; la regla sigue informando el tiempo. */
@media (prefers-reduced-motion: reduce) {
  .base-toast {
    animation: none;
  }

  .base-toast__head,
  .base-toast__chevron,
  .base-toast__panel,
  .base-toast__action,
  .base-toast__dismiss {
    transition: none;
  }
}
</style>
