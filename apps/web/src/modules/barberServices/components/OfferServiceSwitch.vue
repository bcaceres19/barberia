<script setup lang="ts">
// Interruptor «Lo ofrezco» de un servicio para el barbero individual (DEC-115).
// Es el mismo lenguaje de interruptor de Configuración (pista fina y rombo que
// se desliza), en tamaño de fila. El estado se anuncia con `aria-checked` y con
// texto, nunca solo con el color. Solo cambia cuando el servidor confirma
// (`BarberOffering.setOffered`), así que mientras viaja la petición queda
// ocupado y no admite un segundo clic.
import { useToast } from '@/shared/composables'
import type { BarberOffering } from '../model/barberOffering'

const props = defineProps<{
  offering: BarberOffering
  serviceId: string
  serviceName: string
  /** Solo la pista, sin texto: para una columna con su propio encabezado. */
  compact?: boolean
}>()

const toast = useToast()

async function onToggle() {
  const next = !props.offering.isOffered(props.serviceId)
  const outcome = await props.offering.setOffered(props.serviceId, next)
  switch (outcome) {
    case 'changed':
      toast.success(next ? 'Ahora lo ofreces' : 'Ya no lo ofreces', {
        detail: next
          ? `«${props.serviceName}» aparece en tu reserva pública.`
          : `«${props.serviceName}» deja de aparecer en tu reserva pública.`,
      })
      return
    case 'network-error':
      toast.error('No pudimos guardar el cambio', {
        detail: 'Revisa tu conexión e inténtalo de nuevo.',
      })
      return
    case 'error':
      toast.error('No pudimos guardar el cambio', {
        detail: `«${props.serviceName}» sigue como estaba.`,
      })
      return
    case 'unchanged':
      return
  }
}
</script>

<template>
  <button
    v-if="offering.status.value === 'ready'"
    type="button"
    role="switch"
    class="offer-switch"
    :class="{ 'offer-switch--on': offering.isOffered(serviceId) }"
    :aria-checked="offering.isOffered(serviceId)"
    :aria-label="`Lo ofrezco: ${serviceName}`"
    :aria-busy="offering.isBusy(serviceId)"
    :disabled="offering.isBusy(serviceId)"
    @click="onToggle"
  >
    <span v-if="!compact" class="offer-switch__text" aria-hidden="true">{{
      offering.isOffered(serviceId) ? 'Lo ofrezco' : 'No lo ofrezco'
    }}</span>
    <span class="offer-switch__track" aria-hidden="true">
      <span class="offer-switch__thumb" />
    </span>
  </button>
</template>

<style scoped>
.offer-switch {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-height: 44px;
  padding: 0 2px;
  background: transparent;
  border: 0;
  border-radius: 3px;
  color: var(--color-on-strong-muted);
  font-family: var(--font-sans);
  cursor: pointer;
}

.offer-switch:focus-visible {
  outline: var(--border-width-emphasis) solid var(--color-focus);
  outline-offset: 2px;
}

.offer-switch__text {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  white-space: nowrap;
  transition: color var(--motion-duration-base) var(--motion-easing-standard);
}

.offer-switch--on .offer-switch__text {
  color: var(--color-brand-accent-surface);
}

.offer-switch__track {
  position: relative;
  display: block;
  flex: 0 0 auto;
  width: 46px;
  height: 24px;
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-on-strong) 32%, transparent);
  border-radius: 3px;
  background: color-mix(in srgb, var(--color-on-strong) 5%, transparent);
  transition:
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard);
}

.offer-switch--on .offer-switch__track {
  background: color-mix(in srgb, var(--color-brand-accent-surface) 22%, transparent);
  border-color: var(--color-brand-accent-surface);
}

/* Un rombo, el motivo de la casa, recorre la pista con un leve rebote. */
.offer-switch__thumb {
  position: absolute;
  top: 50%;
  left: 5px;
  width: 12px;
  height: 12px;
  background: var(--color-on-strong-muted);
  transform: translateY(-50%) rotate(45deg);
  transition:
    left 280ms cubic-bezier(0.34, 1.56, 0.64, 1),
    background-color var(--motion-duration-base) var(--motion-easing-standard);
}

.offer-switch--on .offer-switch__thumb {
  left: 27px;
  background: var(--color-brand-accent-surface);
}

.offer-switch:hover:not(:disabled) .offer-switch__track {
  border-color: var(--color-brand-accent-surface);
}

.offer-switch:disabled {
  cursor: progress;
  opacity: 0.6;
}

@media (prefers-reduced-motion: reduce) {
  .offer-switch__thumb {
    transition: none;
  }
}
</style>
