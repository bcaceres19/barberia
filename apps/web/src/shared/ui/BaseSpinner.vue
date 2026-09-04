<script setup lang="ts">
/**
 * BaseSpinner - Indicador de progreso indeterminado del sistema visual.
 * Anillo de latón interrumpido (no un círculo completo) con el rombo del
 * divisor NAVA fijo al centro (mismo motivo que `.auth-split__rule::after`
 * en AuthSplitLayout.vue, aquí componentizado por primera vez: issue #189).
 * Tamaños: sm | md | lg (default). Tono: brass (default) | success | warning
 * | danger | info.
 * Accesibilidad: decorativo (aria-hidden) cuando un rótulo visible lo
 * acompaña (caso normal, dentro de PageState); role="img" + aria-label
 * cuando se usa solo, sin texto adyacente.
 * `prefers-reduced-motion: reduce` detiene la rotación por completo: el
 * anillo queda estático, el rombo central nunca giró.
 */
import { computed } from 'vue'

interface Props {
  /** Tamaño del anillo */
  size?: 'sm' | 'md' | 'lg'
  /** Tinte semántico; latón por defecto (espera neutra) */
  tone?: 'brass' | 'success' | 'warning' | 'danger' | 'info'
  /** Texto accesible cuando no hay un rótulo visible adyacente */
  label?: string
}

const props = withDefaults(defineProps<Props>(), {
  size: 'lg',
  tone: 'brass',
})

const toneVar = computed(() => {
  const map: Record<string, string> = {
    brass: 'var(--color-accent-brass)',
    success: 'var(--color-success-text)',
    warning: 'var(--color-warning-text)',
    danger: 'var(--color-danger-text)',
    info: 'var(--color-info-text)',
  }
  return map[props.tone] || map.brass
})

const classes = computed(() => ['base-spinner', `base-spinner--${props.size}`])

const style = computed(() => ({ '--spinner-tone': toneVar.value }))
</script>

<template>
  <span
    :class="classes"
    :style="style"
    :role="label ? 'img' : undefined"
    :aria-label="label"
    :aria-hidden="label ? undefined : 'true'"
  >
    <svg class="base-spinner__ring" viewBox="0 0 40 40" fill="none">
      <circle
        class="base-spinner__track"
        cx="20"
        cy="20"
        r="17"
        stroke-width="2"
        stroke-linecap="round"
      />
    </svg>
    <span class="base-spinner__mark" />
  </span>
</template>

<style scoped>
.base-spinner {
  --spinner-size-sm: 20px;
  --spinner-size-md: 40px;
  --spinner-size-lg: 64px;
  --spinner-mark-size: 8px;

  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: var(--spinner-size-lg);
  height: var(--spinner-size-lg);
}

.base-spinner--sm {
  width: var(--spinner-size-sm);
  height: var(--spinner-size-sm);
  --spinner-mark-size: 4px;
}

.base-spinner--md {
  width: var(--spinner-size-md);
  height: var(--spinner-size-md);
  --spinner-mark-size: 6px;
}

.base-spinner__ring {
  width: 100%;
  height: 100%;
  animation: base-spinner-rotate 900ms linear infinite;
}

@media (prefers-reduced-motion: reduce) {
  .base-spinner__ring {
    animation: none;
  }
}

@keyframes base-spinner-rotate {
  to {
    transform: rotate(360deg);
  }
}

/* Anillo interrumpido (~75% de la circunferencia, 2π·17 ≈ 107): un círculo
 * completo se leería como un aro decorativo, no como progreso en curso. */
.base-spinner__track {
  stroke: var(--spinner-tone);
  stroke-dasharray: 80 107;
  opacity: 0.9;
}

/* Rombo central fijo: no hereda la rotación del anillo (mismo motivo que
 * AuthSplitLayout.vue: el divisor NAVA no gira, solo el progreso lo rodea). */
.base-spinner__mark {
  position: absolute;
  width: var(--spinner-mark-size);
  height: var(--spinner-mark-size);
  background-color: var(--spinner-tone);
  transform: rotate(45deg);
}
</style>
