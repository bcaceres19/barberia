<script setup lang="ts">
/**
 * BaseBadge - Etiqueta/etiqueta de estado del sistema visual.
 * Variantes: neutral | primary | success | warning | danger | info
 * Tamaños: sm | md (default)
 * Con punto de estado opcional (dot), dismissible
 * Accesibilidad: aria-label si solo icono, role="status" para estados
 */
import { computed } from 'vue'

interface Props {
  /** Variante semántica */
  variant?: 'neutral' | 'primary' | 'success' | 'warning' | 'danger' | 'info'
  /** Tamaño */
  size?: 'sm' | 'md'
  /** Si muestra punto de estado */
  dot?: boolean
  /** Si se puede cerrar */
  dismissible?: boolean
  /** Label para accesibilidad (requerido si solo icono/punto) */
  label?: string
  /** Contorno sobre superficie clara/oscura en vez de relleno: el sistema de
   * rellenos claros desaparece sobre pergamino o tinta (issue #189, atlas
   * panel-agenda-eventos). El borde y el texto siguen los del variant/status;
   * solo cambia el fondo. */
  outline?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'neutral',
  size: 'md',
  dot: false,
  dismissible: false,
  outline: false,
})

const emit = defineEmits<{
  dismiss: [event: MouseEvent]
}>()

const classes = computed(() => {
  const base = 'base-badge'
  return [
    base,
    `${base}--${props.variant}`,
    `${base}--${props.size}`,
    props.dot ? `${base}--dot` : '',
    props.dismissible ? `${base}--dismissible` : '',
    props.outline ? `${base}--outline` : '',
  ]
    .filter(Boolean)
    .join(' ')
})

const surfaceVar = computed(() => {
  const map: Record<string, string> = {
    neutral: 'var(--color-inactive-surface)',
    primary: 'var(--color-action-soft)',
    success: 'var(--color-success-surface)',
    warning: 'var(--color-warning-surface)',
    danger: 'var(--color-danger-surface)',
    info: 'var(--color-info-surface)',
  }
  return map[props.variant] || map.neutral
})

const textVar = computed(() => {
  const map: Record<string, string> = {
    neutral: 'var(--color-inactive-text)',
    primary: 'var(--color-action-primary)',
    success: 'var(--color-success-text)',
    warning: 'var(--color-warning-text)',
    danger: 'var(--color-danger-text)',
    info: 'var(--color-info-text)',
  }
  return map[props.variant] || map.neutral
})

const borderVar = computed(() => {
  const map: Record<string, string> = {
    neutral: 'var(--color-inactive-border)',
    primary: 'var(--color-action-soft-border)',
    success: 'var(--color-success-border)',
    warning: 'var(--color-warning-border)',
    danger: 'var(--color-danger-border)',
    info: 'var(--color-info-border)',
  }
  return map[props.variant] || map.neutral
})

// El color del punto SIEMPRE deriva de variant: no existe una vía para que
// un consumidor pase un color libre (auditoría HU-009 — estandar-diseno-visual.md
// §15.4: "Props describen intención, no valores libres").
const dotColorVar = computed(() => {
  const map: Record<string, string> = {
    neutral: 'var(--color-inactive-text)',
    primary: 'var(--color-action-primary)',
    success: 'var(--color-success-text)',
    warning: 'var(--color-warning-text)',
    danger: 'var(--color-danger-text)',
    info: 'var(--color-info-text)',
  }
  return map[props.variant] || map.neutral
})

const dismissSvg = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>`

const handleDismiss = (event: MouseEvent) => {
  event.stopPropagation()
  emit('dismiss', event)
}

const style = computed(() => ({
  '--badge-surface': surfaceVar.value,
  '--badge-text': textVar.value,
  '--badge-border': borderVar.value,
  '--badge-dot-color': dotColorVar.value,
}))
</script>

<template>
  <span
    :class="classes"
    :style="style"
    :role="dot ? 'status' : undefined"
    :aria-label="label"
    :aria-live="dot ? 'polite' : undefined"
  >
    <span v-if="dot" class="base-badge__dot" aria-hidden="true" />
    <slot />
    <button
      v-if="dismissible"
      type="button"
      class="base-badge__dismiss"
      @click="handleDismiss"
      :aria-label="label ? `Cerrar ${label}` : 'Cerrar etiqueta'"
    >
      <span class="base-badge__dismiss-icon" aria-hidden="true" v-html="dismissSvg" />
    </button>
  </span>
</template>

<style scoped>
.base-badge {
  --badge-height-sm: 24px;
  --badge-height-md: 28px;
  --badge-padding-x-sm: var(--space-2);
  --badge-padding-x-md: var(--space-3);
  --badge-font-size-sm: var(--font-size-caption);
  --badge-font-size-md: var(--font-size-body-sm);
  --badge-font-weight: 500;
  --badge-line-height-sm: var(--font-size-caption-line);
  --badge-line-height-md: var(--font-size-body-sm-line);
  --badge-radius: var(--radius-pill);
  --badge-gap: var(--space-1);
  --badge-dot-size-sm: 6px;
  --badge-dot-size-md: 8px;
  /* 24×24: suelo absoluto de WCAG 2.2 (SC 2.5.8), no los 44×44 del botón
   * de icono independiente. Una insignia es una etiqueta compacta en línea
   * con el texto que la rodea (excepción "Inline" de la misma regla);
   * forzar 44×44 aquí rompería el propósito del componente. */
  --badge-dismiss-size: 24px;
  --badge-border-width: var(--border-width-normal);
  --badge-transition:
    background-color var(--motion-duration-fast) var(--motion-easing-standard),
    border-color var(--motion-duration-fast) var(--motion-easing-standard),
    color var(--motion-duration-fast) var(--motion-easing-standard);

  display: inline-flex;
  align-items: center;
  gap: var(--badge-gap);
  height: var(--badge-height-md);
  padding: 0 var(--badge-padding-x-md);
  font-family: var(--font-family-base);
  font-size: var(--badge-font-size-md);
  font-weight: var(--badge-font-weight);
  line-height: var(--badge-line-height-md);
  color: var(--badge-text);
  background-color: var(--badge-surface);
  border: var(--badge-border-width) solid var(--badge-border);
  border-radius: var(--badge-radius);
  white-space: nowrap;
  transition: var(--badge-transition);
}

@media (prefers-reduced-motion: reduce) {
  .base-badge {
    transition: none;
  }
}

.base-badge--sm {
  height: var(--badge-height-sm);
  padding: 0 var(--badge-padding-x-sm);
  font-size: var(--badge-font-size-sm);
  line-height: var(--badge-line-height-sm);
}

.base-badge--dot {
  padding-left: calc(var(--badge-padding-x-md) + var(--badge-dot-size-md) + var(--badge-gap));
}

.base-badge--dot.base-badge--sm {
  padding-left: calc(var(--badge-padding-x-sm) + var(--badge-dot-size-sm) + var(--badge-gap));
}

.base-badge__dot {
  position: relative;
  left: calc(-1 * (var(--badge-padding-x-md) + var(--badge-dot-size-md) + var(--badge-gap)));
  width: var(--badge-dot-size-md);
  height: var(--badge-dot-size-md);
  border-radius: 50%;
  flex-shrink: 0;
  background-color: var(--badge-dot-color);
}

.base-badge--sm .base-badge__dot {
  left: calc(-1 * (var(--badge-padding-x-sm) + var(--badge-dot-size-sm) + var(--badge-gap)));
  width: var(--badge-dot-size-sm);
  height: var(--badge-dot-size-sm);
}

.base-badge--dismissible {
  padding-right: calc(var(--badge-padding-x-md) + var(--badge-dismiss-size) + var(--space-1));
}

.base-badge--sm.base-badge--dismissible {
  padding-right: calc(var(--badge-padding-x-sm) + var(--badge-dismiss-size) + var(--space-1));
}

.base-badge__dismiss {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: var(--badge-dismiss-size);
  height: var(--badge-dismiss-size);
  margin-left: var(--space-1);
  padding: 0;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  color: var(--badge-text);
  opacity: 0.64;
  cursor: pointer;
  transition:
    opacity var(--motion-duration-fast) var(--motion-easing-standard),
    background-color var(--motion-duration-fast) var(--motion-easing-standard);
}

@media (prefers-reduced-motion: reduce) {
  .base-badge__dismiss {
    transition: none;
  }
}

.base-badge__dismiss:hover {
  opacity: 1;
  background-color: var(--color-overlay-hover);
}

.base-badge__dismiss:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-surface),
    0 0 0 4px var(--color-focus);
}

.base-badge__dismiss-icon {
  width: 12px;
  height: 12px;
}

/* Contorno (issue #189): el borde y el texto ya declarados por variant/status
 * bastan; solo se anula el relleno para que la insignia se apoye en contorno
 * sobre pergamino o tinta en vez de competir con esas superficies. */
.base-badge--outline {
  background-color: transparent;
}
</style>
