<script setup lang="ts">
/**
 * BaseAlert - Alerta/feedback del sistema visual.
 * Variantes: success | warning | danger | info | neutral
 * Con acción opcional (botón/link), icono, dismissible
 * Accesibilidad: role="alert" (assertive) o role="status" (polite),
 * aria-live, focus-visible en acción, focus trap si dismissible
 */
import { computed, ref, onMounted, onUnmounted, useSlots } from 'vue'

interface Props {
  /** Variante semántica de la alerta */
  variant?: 'success' | 'warning' | 'danger' | 'info' | 'neutral'
  /** Título de la alerta */
  title?: string
  /** Si se puede cerrar */
  dismissible?: boolean
  /** Role ARIA (alert = assertive, status = polite) */
  role?: 'alert' | 'status'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'info',
  dismissible: false,
  role: 'alert',
})

const emit = defineEmits<{
  dismiss: []
  action: [event: MouseEvent]
}>()

const isVisible = ref(true)
const alertRef = ref<HTMLDivElement | null>(null)
const focusableElementsRef = ref<HTMLElement[]>([])

const classes = computed(() => {
  const base = 'base-alert'
  return [base, `${base}--${props.variant}`, props.dismissible ? `${base}--dismissible` : '']
    .filter(Boolean)
    .join(' ')
})

const surfaceVar = computed(() => {
  const map: Record<string, string> = {
    success: 'var(--color-success-surface)',
    warning: 'var(--color-warning-surface)',
    danger: 'var(--color-danger-surface)',
    info: 'var(--color-info-surface)',
    neutral: 'var(--color-inactive-surface)',
  }
  return map[props.variant] || map.info
})

const textVar = computed(() => {
  const map: Record<string, string> = {
    success: 'var(--color-success-text)',
    warning: 'var(--color-warning-text)',
    danger: 'var(--color-danger-text)',
    info: 'var(--color-info-text)',
    neutral: 'var(--color-inactive-text)',
  }
  return map[props.variant] || map.info
})

const borderVar = computed(() => {
  const map: Record<string, string> = {
    success: 'var(--color-success-border)',
    warning: 'var(--color-warning-border)',
    danger: 'var(--color-danger-border)',
    info: 'var(--color-info-border)',
    neutral: 'var(--color-inactive-border)',
  }
  return map[props.variant] || map.info
})

const iconSvg = computed(() => {
  const icons: Record<string, string> = {
    success: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>`,
    warning: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>`,
    danger: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="15" y1="9" x2="9" y2="15"></line><line x1="9" y1="9" x2="15" y2="15"></line></svg>`,
    info: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>`,
    neutral: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>`,
  }
  return icons[props.variant] || icons.info
})

const dismissSvg = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>`

const handleDismiss = () => {
  isVisible.value = false
  emit('dismiss')
}

const handleActionClick = (event: MouseEvent) => {
  emit('action', event)
}

const handleKeyDown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && props.dismissible) {
    handleDismiss()
    return
  }
  // Tab trap para alert dismissible. Se recalculan los elementos
  // enfocables en cada Tab, no solo al montar: un slot de acción puede
  // cambiar su contenido después del montaje inicial.
  if (event.key === 'Tab' && props.dismissible) {
    updateFocusableElements()
    if (!focusableElementsRef.value.length) return
    const first = focusableElementsRef.value[0]
    const last = focusableElementsRef.value[focusableElementsRef.value.length - 1]
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault()
      last.focus()
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault()
      first.focus()
    }
  }
}

onMounted(() => {
  if (props.dismissible && alertRef.value) {
    updateFocusableElements()
    // Focus en el botón de cerrar si no hay otra acción
    if (!hasActionSlot.value) {
      const dismissBtn = alertRef.value.querySelector('.base-alert__dismiss') as HTMLElement
      dismissBtn?.focus()
    }
    document.addEventListener('keydown', handleKeyDown)
  }
})

onUnmounted(() => {
  if (props.dismissible) {
    document.removeEventListener('keydown', handleKeyDown)
  }
})

const updateFocusableElements = () => {
  if (!alertRef.value) return
  const elements = alertRef.value.querySelectorAll<HTMLElement>(
    'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])',
  )
  focusableElementsRef.value = Array.from(elements).filter((el) => !el.hasAttribute('disabled'))
}

const slots = useSlots()
const hasActionSlot = computed(() => !!slots.action)

const style = computed(() => ({
  '--alert-surface': surfaceVar.value,
  '--alert-text': textVar.value,
  '--alert-border': borderVar.value,
}))
</script>

<template>
  <div
    v-show="isVisible"
    ref="alertRef"
    :class="classes"
    :style="style"
    :role="role"
    :aria-live="role === 'alert' ? 'assertive' : 'polite'"
    :aria-atomic="true"
    @keydown="handleKeyDown"
  >
    <div class="base-alert__icon" aria-hidden="true" v-html="iconSvg" />

    <div class="base-alert__content">
      <div v-if="title || slots.default" class="base-alert__text">
        <h4 v-if="title" class="base-alert__title">{{ title }}</h4>
        <slot />
      </div>

      <div v-if="hasActionSlot" class="base-alert__actions">
        <slot name="action" @click="handleActionClick" />
      </div>
    </div>

    <button
      v-if="dismissible"
      type="button"
      class="base-alert__dismiss"
      @click="handleDismiss"
      aria-label="Cerrar alerta"
    >
      <span class="base-alert__dismiss-icon" aria-hidden="true" v-html="dismissSvg" />
    </button>
  </div>
</template>

<style scoped>
.base-alert {
  --alert-padding: var(--space-4);
  --alert-gap: var(--space-3);
  --alert-radius: var(--radius-md);
  --alert-border-width: var(--border-width-normal);
  --alert-icon-size: 20px;
  --alert-font-size: var(--font-size-body);
  --alert-title-size: var(--font-size-body);
  --alert-line-height: var(--font-size-body-line);
  --alert-transition:
    opacity var(--motion-duration-base) var(--motion-easing-standard),
    transform var(--motion-duration-base) var(--motion-easing-standard);

  display: flex;
  align-items: flex-start;
  gap: var(--alert-gap);
  padding: var(--alert-padding);
  background-color: var(--alert-surface);
  border: var(--alert-border-width) solid var(--alert-border);
  border-radius: var(--alert-radius);
  color: var(--alert-text);
  font-family: var(--font-family-base);
  font-size: var(--alert-font-size);
  line-height: var(--alert-line-height);
  transition: var(--alert-transition);
  animation: base-alert-slide-in var(--motion-duration-base) ease-out;
}

@media (prefers-reduced-motion: reduce) {
  .base-alert {
    transition: none;
    animation: none;
  }
}

@keyframes base-alert-slide-in {
  from {
    opacity: 0;
    transform: translateY(-8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.base-alert[v-show='false'] {
  opacity: 0;
  transform: translateY(-8px);
  pointer-events: none;
}

.base-alert__icon {
  flex-shrink: 0;
  width: var(--alert-icon-size);
  height: var(--alert-icon-size);
  margin-top: 2px;
  color: var(--alert-text);
}

.base-alert__content {
  flex: 1;
  min-width: 0;
}

.base-alert__text {
  margin-bottom: var(--space-2);
}

.base-alert__title {
  margin: 0 0 var(--space-1) 0;
  font-family: var(--font-family-base);
  font-size: var(--alert-title-size);
  font-weight: 600;
  line-height: var(--font-size-body-line);
  color: var(--alert-text);
}

.base-alert__text:only-child .base-alert__title {
  margin-bottom: 0;
}

.base-alert__actions {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.base-alert__dismiss {
  /* Botón de icono independiente (estandar-diseno-visual.md §6.2): 44×44,
   * no el 28×28 anterior. CA-009-03. */
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: var(--control-height-icon);
  height: var(--control-height-icon);
  padding: 0;
  margin: calc(-1 * var(--space-3)) calc(-1 * var(--space-3)) calc(-1 * var(--space-3))
    var(--space-1);
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  color: var(--alert-text);
  opacity: 0.64;
  cursor: pointer;
  transition:
    opacity var(--motion-duration-fast) var(--motion-easing-standard),
    background-color var(--motion-duration-fast) var(--motion-easing-standard);
}

@media (prefers-reduced-motion: reduce) {
  .base-alert__dismiss {
    transition: none;
  }
}

.base-alert__dismiss:hover {
  opacity: 1;
  background-color: var(--color-overlay-hover);
}

.base-alert__dismiss:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-surface),
    0 0 0 4px var(--color-focus);
}

.base-alert__dismiss-icon {
  width: 16px;
  height: 16px;
}

/* Variantes dismissible: más padding a la derecha para el botón. El
   objetivo táctil real es --control-height-icon (44px); el margen negativo
   de .base-alert__dismiss recorta cuánto de eso sobresale visualmente. */
.base-alert--dismissible {
  padding-right: calc(var(--alert-padding) + var(--control-height-icon) - var(--space-3));
}

/* Neutral usa texto secundario */
.base-alert--neutral {
  --alert-text: var(--color-text-secondary);
}
</style>
