<script setup lang="ts">
/**
 * BaseButton - Botón base del sistema visual.
 * Variantes: primary | secondary | soft | ghost | danger
 * (estandar-diseno-visual.md §8.1: primaria, secundaria, texto, peligro)
 * Estados: default | hover | active | focus-visible | disabled | loading
 * Tamaños: md (44px, estándar) | lg (48px, principal móvil, §6.2). No
 * existe una variante más pequeña: el estándar visual no define un tamaño
 * de botón por debajo del objetivo táctil de 44 × 44 px (CA-009-03), así
 * que este componente no inventa uno.
 * Accesibilidad: aria-pressed (toggle), aria-busy (carga), foco visible con
 * --color-focus, prefers-reduced-motion.
 */
import { computed } from 'vue'

interface Props {
  /** Variante visual del botón */
  variant?: 'primary' | 'secondary' | 'soft' | 'ghost' | 'danger'
  /** Tamaño del botón: md (44px) o lg (48px, acción principal móvil) */
  size?: 'md' | 'lg'
  /** Si el botón está deshabilitado */
  disabled?: boolean
  /** Si muestra estado de carga. Bloquea la interacción como disabled, sin
   * sustituir la idempotencia del servidor (estandar-frontend-vue.md §6). */
  loading?: boolean
  /** Si es un botón toggle (mantiene estado pressed) */
  pressed?: boolean
  /** Tipo nativo del botón */
  type?: 'button' | 'submit' | 'reset'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'md',
  disabled: false,
  loading: false,
  pressed: false,
  type: 'button',
})

const emit = defineEmits<{
  click: [event: MouseEvent]
}>()

const classes = computed(() => {
  const base = 'base-button'
  return [
    base,
    `${base}--${props.variant}`,
    `${base}--${props.size}`,
    props.disabled ? `${base}--disabled` : '',
    props.loading ? `${base}--loading` : '',
    props.pressed ? `${base}--pressed` : '',
  ]
    .filter(Boolean)
    .join(' ')
})

// El atributo nativo `disabled` ya impide el click y la activación por
// teclado del navegador mientras disabled o loading estén activos; este
// guardia es una defensa adicional explícita, no la única barrera real.
const handleClick = (event: MouseEvent) => {
  if (props.disabled || props.loading) {
    event.preventDefault()
    event.stopPropagation()
    return
  }
  emit('click', event)
}
</script>

<template>
  <button
    :class="classes"
    :type="type"
    :disabled="disabled || loading"
    :aria-disabled="disabled || loading"
    :aria-pressed="pressed"
    :aria-busy="loading"
    @click="handleClick"
  >
    <span class="base-button__content">
      <span v-if="loading" class="base-button__spinner" aria-hidden="true">
        <svg
          class="base-button__spinner-svg"
          viewBox="0 0 24 24"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <circle
            class="base-button__spinner-circle"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="3"
            stroke-linecap="round"
            stroke-dasharray="31.4 31.4"
          />
        </svg>
      </span>
      <slot />
    </span>
  </button>
</template>

<style scoped>
.base-button {
  --btn-height: var(--control-height);
  --btn-height-lg: var(--control-height-primary-mobile);
  --btn-padding-x: var(--space-4);
  --btn-padding-x-lg: var(--space-5);
  --btn-gap: var(--space-2);
  --btn-font-size: var(--font-size-body);
  --btn-font-weight: 500;
  --btn-radius: 2px;
  --btn-transition:
    background-color var(--motion-duration-fast) var(--motion-easing-standard),
    border-color var(--motion-duration-fast) var(--motion-easing-standard),
    color var(--motion-duration-fast) var(--motion-easing-standard),
    box-shadow var(--motion-duration-fast) var(--motion-easing-standard);
  --btn-focus-ring: 0 0 0 2px var(--color-surface), 0 0 0 4px var(--color-focus);

  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--btn-gap);
  height: var(--btn-height);
  padding: 0 var(--btn-padding-x);
  font-family: var(--font-family-base);
  font-size: var(--btn-font-size);
  font-weight: var(--btn-font-weight);
  line-height: 1;
  border: var(--border-width-normal) solid transparent;
  border-radius: var(--btn-radius);
  cursor: pointer;
  text-decoration: none;
  white-space: nowrap;
  transition: var(--btn-transition);
  outline: none;
}

@media (prefers-reduced-motion: reduce) {
  .base-button {
    transition: none;
  }
}

.base-button:focus-visible {
  box-shadow: var(--btn-focus-ring);
}

.base-button--lg {
  height: var(--btn-height-lg);
  padding: 0 var(--btn-padding-x-lg);
}

/* Variant: primary */
.base-button--primary {
  background-color: var(--color-action-primary);
  color: var(--color-on-strong);
  border-color: var(--color-action-primary);
}

.base-button--primary:hover:not(:disabled):not(.base-button--loading) {
  background-color: var(--color-action-primary-hover);
  border-color: var(--color-action-primary-hover);
}

.base-button--primary:active:not(:disabled):not(.base-button--loading) {
  background-color: var(--color-action-primary-active);
  border-color: var(--color-action-primary-active);
}

/* Variant: secondary — filete de latón con filete inferior acentuado
   (issue #212): la línea base recibe el mismo doble espesor que el campo
   reglado, no un borde uniforme de 1px. */
.base-button--secondary {
  background-color: var(--color-surface);
  color: var(--color-text-primary);
  border-color: var(--color-accent-brass);
  border-bottom-width: var(--border-width-emphasis);
}

.base-button--secondary:hover:not(:disabled):not(.base-button--loading) {
  background-color: var(--color-surface-muted);
  border-color: var(--color-accent-brass);
}

.base-button--secondary:active:not(:disabled):not(.base-button--loading) {
  background-color: var(--color-border-subtle);
}

/* Variant: soft */
.base-button--soft {
  background-color: var(--color-action-soft);
  color: var(--color-action-primary);
  border-color: transparent;
}

.base-button--soft:hover:not(:disabled):not(.base-button--loading) {
  background-color: var(--color-action-soft);
  border-color: var(--color-action-soft-border);
}

.base-button--soft:active:not(:disabled):not(.base-button--loading) {
  background-color: var(--color-action-soft-active);
}

/* Variant: ghost — misma familia reglada que secondary: filete de latón y
   filete inferior acentuado en vez de borde transparente. */
.base-button--ghost {
  background-color: transparent;
  color: var(--color-action-primary);
  border-color: var(--color-accent-brass);
  border-bottom-width: var(--border-width-emphasis);
}

.base-button--ghost:hover:not(:disabled):not(.base-button--loading) {
  background-color: var(--color-action-soft);
}

.base-button--ghost:active:not(:disabled):not(.base-button--loading) {
  background-color: var(--color-action-soft-active);
}

/* Variant: danger */
.base-button--danger {
  background-color: var(--color-danger-action);
  color: var(--color-on-strong);
  border-color: var(--color-danger-action);
}

/* El estándar visual no define un tono más oscuro de peligro para
   hover/active (solo el único "Peligro" de la tabla de contraste, §4.2):
   en vez de inventar un hexadecimal nuevo sin decisión del propietario,
   estos estados oscurecen el mismo token con un multiplicador numérico,
   igual que el resto del archivo ya usa opacity para disabled/secundario. */
.base-button--danger:hover:not(:disabled):not(.base-button--loading) {
  filter: brightness(90%);
}

.base-button--danger:active:not(:disabled):not(.base-button--loading) {
  filter: brightness(80%);
}

/* Disabled */
.base-button--disabled,
.base-button:disabled {
  opacity: 0.48;
  cursor: not-allowed;
  pointer-events: none;
}

/* Loading */
.base-button--loading {
  position: relative;
  color: transparent;
}

.base-button--loading .base-button__spinner {
  position: absolute;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* El giro del spinner es un indicador de progreso indeterminado, no una
   transición de estado: por eso no usa --motion-duration-fast/base (esas
   cubren transiciones de 120-200 ms que explican un cambio, CA-009-07).
   Sigue apagándose con prefers-reduced-motion más abajo. */
.base-button__spinner-svg {
  width: 1em;
  height: 1em;
  animation: base-button-spin 0.8s linear infinite;
}

@keyframes base-button-spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .base-button__spinner-svg {
    animation: none;
    opacity: 0.6;
  }
}

.base-button__spinner-circle {
  animation: base-button-spin-dash 1.4s ease-in-out infinite;
}

@keyframes base-button-spin-dash {
  0% {
    stroke-dasharray: 1, 200;
    stroke-dashoffset: 0;
  }
  50% {
    stroke-dasharray: 90, 200;
    stroke-dashoffset: -35;
  }
  100% {
    stroke-dasharray: 90, 200;
    stroke-dashoffset: -124;
  }
}

/* Pressed (para toggle) */
.base-button--pressed {
  box-shadow: inset 0 2px 4px var(--color-overlay-hover);
}

/* Icon-only (cuando solo hay icono) */
.base-button:has(> .base-button__content > :only-child:not(:has(+ *))) {
  padding-left: calc(var(--btn-padding-x) - 2px);
  padding-right: calc(var(--btn-padding-x) - 2px);
}

.base-button--lg:has(> .base-button__content > :only-child:not(:has(+ *))) {
  padding-left: calc(var(--btn-padding-x-lg) - 2px);
  padding-right: calc(var(--btn-padding-x-lg) - 2px);
}
</style>
