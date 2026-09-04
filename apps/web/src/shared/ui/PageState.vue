<script setup lang="ts">
/**
 * PageState - Estado de página centrado (issue #189, atlas
 * panel-agenda-eventos, eventos 02/03/04/07/08): para cuando la carga, el
 * error o el vacío ocupan TODA el área de contenido. Cuando el estado
 * acompaña contenido que sigue visible (p. ej. el evento 10, zona horaria no
 * confirmada), se usa `BaseAlert` como nota al margen en su lugar — este
 * componente no reemplaza esa variante.
 *
 * variant="loading": BaseSpinner + titular, sin divisor ni rótulo de estado
 * (evento 02). Los demás variants componen el divisor NAVA (regla-rombo-regla,
 * misma construcción que AuthSplitLayout.vue) con un rótulo de estado en
 * versalitas opcional (omitido para `info`, presente para `warning`/`danger`
 * con la palabra "Atención"/"Error", igual criterio que BaseAlert), titular
 * serif y cuerpo. La acción real (BaseButton) es responsabilidad del
 * consumidor vía el slot `action`.
 */
import { computed } from 'vue'
import BaseSpinner from './BaseSpinner.vue'

interface Props {
  variant?: 'loading' | 'info' | 'success' | 'warning' | 'danger'
  /** Rótulo de estado en versalitas, p. ej. "Atención"/"Error". Se omite si
   * no se pasa: no todo estado necesita nombrarse (evento 04, informativo). */
  statusLabel?: string
  /** Titular serif — el texto exacto del código, nunca reescrito aquí. */
  headline: string
  /** loading/info → status (polite); warning/danger → alert (assertive). */
  role?: 'status' | 'alert'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'info',
})

const isLoading = computed(() => props.variant === 'loading')

const resolvedRole = computed(() => {
  if (props.role) return props.role
  return props.variant === 'warning' || props.variant === 'danger' ? 'alert' : 'status'
})

const markVar = computed(() => {
  const map: Record<string, string> = {
    info: 'var(--color-accent-brass)',
    success: 'var(--color-success-text)',
    warning: 'var(--color-warning-border)',
    danger: 'var(--color-danger-action)',
  }
  return map[props.variant] || map.info
})

const style = computed(() => ({ '--page-state-mark': markVar.value }))
</script>

<template>
  <div
    class="page-state"
    :style="style"
    :role="resolvedRole"
    :aria-live="resolvedRole === 'alert' ? 'assertive' : 'polite'"
    aria-atomic="true"
  >
    <BaseSpinner v-if="isLoading" size="lg" tone="brass" aria-hidden="true" />
    <span v-else class="page-state__divider" aria-hidden="true" />

    <p v-if="statusLabel" class="page-state__status">{{ statusLabel }}</p>
    <h2 class="page-state__headline">{{ headline }}</h2>
    <p v-if="$slots.default" class="page-state__body"><slot /></p>
    <div v-if="$slots.action" class="page-state__action">
      <slot name="action" />
    </div>
  </div>
</template>

<style scoped>
.page-state {
  display: flex;
  flex: 1;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  padding: var(--space-8) var(--space-4);
  text-align: center;
}

/* Divisor NAVA regla-rombo-regla (misma construcción que
 * AuthSplitLayout.vue .auth-split__rule): línea de 2px con un rombo fijo al
 * centro, "recortado" con box-shadow del color de la superficie donde vive
 * PageState (tinta de punta a punta en /panel). */
.page-state__divider {
  position: relative;
  width: 220px;
  height: 2px;
  margin-bottom: var(--space-1);
  background-color: var(--page-state-mark);
}

.page-state__divider::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 9px;
  height: 9px;
  transform: translate(-50%, -50%) rotate(45deg);
  background-color: var(--page-state-mark);
  box-shadow: 0 0 0 8px var(--color-surface-strong);
}

.page-state__status {
  margin: 0;
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--page-state-mark);
}

.page-state__headline {
  margin: 0;
  max-width: 32ch;
  font-family: var(--font-display);
  font-size: var(--font-size-h2);
  line-height: var(--font-size-h2-line);
  font-weight: var(--font-weight-h2);
  color: var(--color-on-strong);
}

.page-state__body {
  margin: 0;
  max-width: 42ch;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  line-height: var(--font-size-body-line);
  color: var(--color-on-strong-muted);
}

/* Destino de otra sección del producto (evento 04, "Barberos"): el
 * consumidor marca el enlace con esta clase para heredar el latón — PageState
 * no expone un prop de color libre (props describen intención, no valores
 * libres, mismo criterio que BaseBadge). */
.page-state__body :deep(.page-state__link) {
  color: var(--color-accent-brass);
  font-weight: 600;
  text-decoration: none;
}

.page-state__body :deep(.page-state__link:hover) {
  text-decoration: underline;
}

.page-state__action {
  margin-top: var(--space-2);
}
</style>
