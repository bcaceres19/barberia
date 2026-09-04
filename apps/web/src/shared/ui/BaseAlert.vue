<script setup lang="ts">
/**
 * BaseAlert - Alerta/feedback del sistema visual, "nota al margen"
 * (issue #212, auth-eventos/README.md "Contrato visual común").
 * Variantes: success | warning | danger | info | neutral | plain
 * `plain` es la variante sin relleno para contenido informativo que no
 * comunica un estado del sistema (por ejemplo, requisitos de contraseña):
 * filete lateral de latón, sin fondo teñido, sin palabra de estado.
 * Con acción opcional (botón/link), dismissible.
 * Accesibilidad: role="alert" (assertive) o role="status" (polite),
 * aria-live, focus-visible en acción, focus trap si dismissible. La
 * palabra de estado en versalitas (Error/Atención/Nota/Confirmación) es
 * la que cumple "icono, texto y estructura además del color" (WCAG 2.2 AA
 * 1.4.1): ya no hay un glifo circular genérico.
 */
import { computed, ref, onMounted, onUnmounted, useSlots } from 'vue'

interface Props {
  /** Variante semántica de la alerta */
  variant?: 'success' | 'warning' | 'danger' | 'info' | 'neutral' | 'plain'
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
    plain: 'transparent',
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
    plain: 'var(--color-text-primary)',
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
    plain: 'var(--color-accent-brass)',
  }
  return map[props.variant] || map.info
})

// Palabra de estado en versalitas sobre el título (contrato visual del
// issue #212): sustituye el glifo circular anterior como el elemento que
// distingue el estado además del color (WCAG 2.2 AA 1.4.1). `neutral` y
// `plain` no representan un estado del sistema, así que no llevan palabra.
const isPlain = computed(() => props.variant === 'plain')
const statusWord = computed(() => {
  const map: Partial<Record<string, string>> = {
    danger: 'Error',
    warning: 'Atención',
    info: 'Nota',
    success: 'Confirmación',
  }
  return map[props.variant]
})

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

// Permite a un consumidor (por ejemplo, un resumen de validación con más
// de un error, CA-010-05) mover el foco al propio recuadro de la alerta
// sin duplicar su `role`/estructura en un wrapper aparte.
defineExpose({
  focus: () => alertRef.value?.focus(),
})

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
    <div class="base-alert__content">
      <div v-if="statusWord || title || slots.default" class="base-alert__text">
        <p v-if="statusWord" class="base-alert__status">{{ statusWord }}</p>
        <h4 v-if="title && !isPlain" class="base-alert__title">{{ title }}</h4>
        <p v-else-if="title && isPlain" class="base-alert__title base-alert__title--plain">
          {{ title }}
        </p>
        <slot />
      </div>

      <div v-if="hasActionSlot" class="base-alert__actions">
        <slot name="action" @click="handleActionClick" />
      </div>
    </div>

    <button v-if="dismissible" type="button" class="base-alert__dismiss" @click="handleDismiss">
      Descartar
    </button>
  </div>
</template>

<style scoped>
/* Nota al margen (issue #212): filete lateral de 4px del color de estado
   sobre fondo apenas teñido, radio 2px — ya no un recuadro con borde
   perimetral uniforme e icono circular. */
.base-alert {
  --alert-padding: var(--space-4);
  --alert-gap: var(--space-3);
  --alert-radius: 2px;
  --alert-border-width: 4px;
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
  border-left: var(--alert-border-width) solid var(--alert-border);
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

/* Foco del propio recuadro (consumidores que reciben `tabindex="-1"` para
   moverle el foco, p. ej. un resumen de validación con más de un error). */
.base-alert:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-surface),
    0 0 0 4px var(--color-focus);
}

.base-alert[v-show='false'] {
  opacity: 0;
  transform: translateY(-8px);
  pointer-events: none;
}

.base-alert__content {
  flex: 1;
  min-width: 0;
}

.base-alert__text {
  margin-bottom: var(--space-2);
}

/* Palabra de estado (Error/Atención/Nota/Confirmación): versalitas de
   latón sobre el título, misma familia tipográfica que el rótulo de
   BaseInput y las ranuras de OtpInput. */
.base-alert__status {
  margin: 0 0 var(--space-1) 0;
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-accent-brass);
}

.base-alert__title {
  margin: 0 0 var(--space-1) 0;
  font-family: var(--font-family-base);
  font-size: var(--alert-title-size);
  font-weight: 600;
  line-height: var(--font-size-body-line);
  color: var(--alert-text);
}

/* Variante sin relleno (caja de requisitos de contraseña): el título es
   la única línea de encabezado, compuesto como versalita de latón en vez
   de repetirse además como titular en negrita. */
.base-alert__title--plain {
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-accent-brass);
}

.base-alert__text:only-child .base-alert__title,
.base-alert__text:only-child .base-alert__title--plain {
  margin-bottom: 0;
}

.base-alert__actions {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

/* Control de descarte rotulado "Descartar" (issue #212), no un botón de
   icono anónimo: el texto visible es su propio nombre accesible (WCAG
   2.5.3 Label in Name), así que ya no lleva aria-label aparte. La altura
   mínima conserva el objetivo táctil de 44px (CA-009-03) aunque el
   contenido visible sea más compacto que el glifo anterior. */
.base-alert__dismiss {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: var(--control-height-icon);
  padding: 0 var(--space-2);
  margin: calc(-1 * var(--space-2)) calc(-1 * var(--space-2)) calc(-1 * var(--space-2)) 0;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition:
    color var(--motion-duration-fast) var(--motion-easing-standard),
    background-color var(--motion-duration-fast) var(--motion-easing-standard);
}

@media (prefers-reduced-motion: reduce) {
  .base-alert__dismiss {
    transition: none;
  }
}

.base-alert__dismiss:hover {
  color: var(--alert-text);
  background-color: var(--color-overlay-hover);
}

.base-alert__dismiss:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-surface),
    0 0 0 4px var(--color-focus);
}

/* Neutral usa texto secundario */
.base-alert--neutral {
  --alert-text: var(--color-text-secondary);
}
</style>
