<script setup lang="ts">
/**
 * BaseButton - Botón base del sistema visual.
 * Implementa variantes: primary | secondary | soft | ghost | danger
 * Estados: default | hover | active | focus-visible | disabled | loading
 * Tamaños: sm (40×40 icon-only) | md (44×44) | lg (48×48 mobile-primary)
 * Accesibilidad: aria-pressed (toggle), aria-disabled, role="button" nativo,
 * focus-visible ring 2px offset 2px con --color-focus, prefers-reduced-motion
 */
import { computed, ref } from 'vue'

interface Props {
  /** Variante visual del botón */
  variant?: 'primary' | 'secondary' | 'soft' | 'ghost' | 'danger'
  /** Tamaño del botón */
  size?: 'sm' | 'md' | 'lg'
  /** Si el botón está deshabilitado */
  disabled?: boolean
  /** Si muestra estado de carga */
  loading?: boolean
  /** Si es un botón toggle (mantiene estado pressed) */
  pressed?: boolean
  /** Tipo nativo del botón */
  type?: 'button' | 'submit' | 'reset'
  /** Clases CSS adicionales */
  class?: string
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'md',
  disabled: false,
  loading: false,
  pressed: false,
  type: 'button',
  class: '',
})

const emit = defineEmits<{
  click: [event: MouseEvent]
}>()

const buttonRef = ref<HTMLButtonElement | null>(null)
void buttonRef // template ref, no runtime usage needed

const classes = computed(() => {
  const base = 'base-button'
  return [
    base,
    `${base}--${props.variant}`,
    `${base}--${props.size}`,
    props.disabled ? `${base}--disabled` : '',
    props.loading ? `${base}--loading` : '',
    props.pressed ? `${base}--pressed` : '',
    props.class,
  ]
    .filter(Boolean)
    .join(' ')
})

const handleClick = (event: MouseEvent) => {
  if (props.disabled || props.loading) {
    event.preventDefault()
    event.stopPropagation()
    return
  }
  emit('click', event)
}

const handleKeyDown = (event: KeyboardEvent) => {
  if (props.disabled || props.loading) {
    event.preventDefault()
    return
  }
  // Enter o Space activan el botón (comportamiento nativo)
  if (event.key === 'Enter' || event.key === ' ') {
    // El navegador maneja el click nativo para button type="button"
  }
}
</script>

<template>
  <button
    ref="buttonRef"
    :class="classes"
    :type="type"
    :disabled="disabled || loading"
    :aria-disabled="disabled || loading"
    :aria-pressed="pressed"
    :aria-busy="loading"
    @click="handleClick"
    @keydown="handleKeyDown"
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
  --btn-height-sm: 40px;
  --btn-height-lg: var(--control-height-primary-mobile);
  --btn-padding-x: var(--space-4);
  --btn-padding-x-sm: var(--space-3);
  --btn-padding-x-lg: var(--space-5);
  --btn-gap: var(--space-2);
  --btn-font-size: var(--font-size-body);
  --btn-font-weight: 500;
  --btn-radius: var(--radius-md);
  --btn-transition: background-color 0.12s ease, border-color 0.12s ease, color 0.12s ease, box-shadow 0.12s ease;
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

.base-button--sm {
  height: var(--btn-height-sm);
  padding: 0 var(--btn-padding-x-sm);
  min-width: var(--btn-height-sm);
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

/* Variant: secondary */
.base-button--secondary {
  background-color: var(--color-surface);
  color: var(--color-text-primary);
  border-color: var(--color-border-control);
}

.base-button--secondary:hover:not(:disabled):not(.base-button--loading) {
  background-color: var(--color-surface-muted);
  border-color: var(--color-border-control);
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
  background-color: #dbeafe;
}

/* Variant: ghost */
.base-button--ghost {
  background-color: transparent;
  color: var(--color-action-primary);
  border-color: transparent;
}

.base-button--ghost:hover:not(:disabled):not(.base-button--loading) {
  background-color: var(--color-action-soft);
}

.base-button--ghost:active:not(:disabled):not(.base-button--loading) {
  background-color: #dbeafe;
}

/* Variant: danger */
.base-button--danger {
  background-color: var(--color-danger-action);
  color: var(--color-on-strong);
  border-color: var(--color-danger-action);
}

.base-button--danger:hover:not(:disabled):not(.base-button--loading) {
  background-color: #991b1b;
  border-color: #991b1b;
}

.base-button--danger:active:not(:disabled):not(.base-button--loading) {
  background-color: #7f1d1d;
  border-color: #7f1d1d;
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

.base-button__spinner-svg {
  width: 1em;
  height: 1em;
  animation: base-button-spin 0.8s linear infinite;
}

@keyframes base-button-spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
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
  box-shadow: inset 0 2px 4px rgb(15 23 42 / 10%);
}

/* Icon-only (cuando solo hay icono) */
.base-button:has(> .base-button__content > :only-child:not(:has(+ *))) {
  padding-left: calc(var(--btn-padding-x) - 2px);
  padding-right: calc(var(--btn-padding-x) - 2px);
}

.base-button--sm:has(> .base-button__content > :only-child:not(:has(+ *))) {
  padding-left: calc(var(--btn-padding-x-sm) - 2px);
  padding-right: calc(var(--btn-padding-x-sm) - 2px);
}

.base-button--lg:has(> .base-button__content > :only-child:not(:has(+ *))) {
  padding-left: calc(var(--btn-padding-x-lg) - 2px);
  padding-right: calc(var(--btn-padding-x-lg) - 2px);
}
</style>