<script setup lang="ts">
// Vista previa de la política de reserva: traduce lo que se está editando a lo
// que verá la persona que reserva. Es solo lectura y deriva de los valores del
// formulario (no consulta ni crea nada). La regla dibuja la ventana de reserva
// desde «ahora»: el tramo rayado es la anticipación mínima, que todavía no se
// puede reservar. Debajo, las franjas de una mañana de ejemplo con la rejilla
// elegida. El gráfico es decorativo (`aria-hidden`); la lista de datos lleva
// el mismo contenido en texto para lectores de pantalla.
import { computed } from 'vue'
import {
  blockedShareOfWindow,
  formatDuration,
  sampleSlotLabels,
  type BookingPolicyFormValues,
} from '../model/bookingPolicy'

const props = defineProps<{ policy: BookingPolicyFormValues }>()

const blockedPercent = computed(
  () => blockedShareOfWindow(props.policy.minAdvanceMinutes, props.policy.maxAdvanceDays) * 100,
)

const leadText = computed(() =>
  props.policy.minAdvanceMinutes > 0
    ? `desde ${formatDuration(props.policy.minAdvanceMinutes)} de antelación`
    : 'sin antelación mínima',
)

const windowText = computed(() => {
  const days = props.policy.maxAdvanceDays
  return `hasta ${days} ${days === 1 ? 'día' : 'días'} por delante`
})

const cancelText = computed(() =>
  props.policy.cancellationDeadlineMinutes > 0
    ? `hasta ${formatDuration(props.policy.cancellationDeadlineMinutes)} antes del turno`
    : 'no puede cancelar por su cuenta',
)

const lateText = computed(() => {
  if (props.policy.cancellationDeadlineMinutes <= 0) return 'no aplica'
  if (!props.policy.lateCancellationClientAllowed) return 'no puede cancelar'
  return props.policy.lateCancellationReasonRequired
    ? 'puede cancelar, indicando un motivo'
    : 'puede cancelar sin motivo'
})

const slots = computed(() => sampleSlotLabels(props.policy.slotGridMinutes))
</script>

<template>
  <aside class="policy-preview" aria-labelledby="policy-preview-title">
    <p class="policy-preview__kicker">Vista previa</p>
    <h2 id="policy-preview-title" class="policy-preview__title">Así lo ve tu cliente</h2>

    <div class="policy-preview__window" aria-hidden="true">
      <div class="policy-preview__track">
        <span class="policy-preview__blocked" :style="{ width: `${blockedPercent}%` }" />
        <span class="policy-preview__open" :style="{ left: `${blockedPercent}%` }" />
        <span class="policy-preview__now" />
      </div>
      <div class="policy-preview__legend">
        <span>Ahora</span>
        <span>+{{ policy.maxAdvanceDays }} d</span>
      </div>
    </div>

    <dl class="policy-preview__facts">
      <div class="policy-preview__fact">
        <dt>Reserva</dt>
        <dd>{{ leadText }}, {{ windowText }}</dd>
      </div>
      <div class="policy-preview__fact">
        <dt>Franjas</dt>
        <dd>una cada {{ policy.slotGridMinutes }} min</dd>
      </div>
      <div class="policy-preview__fact">
        <dt>Cancelación</dt>
        <dd>{{ cancelText }}</dd>
      </div>
      <div class="policy-preview__fact">
        <dt>Fuera de plazo</dt>
        <dd>{{ lateText }}</dd>
      </div>
    </dl>

    <div v-if="slots.length > 0" class="policy-preview__sample" aria-hidden="true">
      <span class="policy-preview__sample-label">Ejemplo de franjas</span>
      <ul class="policy-preview__slots">
        <li
          v-for="(slot, index) in slots"
          :key="`${policy.slotGridMinutes}-${slot}`"
          class="policy-preview__slot"
          :style="{ '--slot-index': index }"
        >
          {{ slot }}
        </li>
      </ul>
    </div>
  </aside>
</template>

<style scoped>
.policy-preview {
  position: sticky;
  top: 24px;
  align-self: start;
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 20px 20px 22px;
  background: color-mix(in srgb, var(--color-on-strong) 3%, transparent);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-left: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  border-radius: 3px;
  animation: policy-preview-enter 520ms var(--motion-easing-standard) 220ms both;
}

.policy-preview__kicker {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.08em;
  line-height: 14px;
  text-transform: uppercase;
}

.policy-preview__kicker::before {
  content: '';
  width: 5px;
  height: 5px;
  background: currentColor;
  transform: rotate(45deg);
  animation: policy-preview-pulse 2.4s ease-in-out infinite;
}

.policy-preview__title {
  margin: -10px 0 0;
  font-family: var(--font-display);
  font-size: var(--font-size-title-item);
  font-weight: 400;
  line-height: 27px;
  color: var(--color-on-strong);
}

/* --- Ventana de reserva ---------------------------------------------------- */

.policy-preview__window {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.policy-preview__track {
  position: relative;
  height: 14px;
  overflow: hidden;
  background: color-mix(in srgb, var(--color-on-strong) 5%, transparent);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: 2px;
}

.policy-preview__blocked,
.policy-preview__open {
  position: absolute;
  top: 0;
  bottom: 0;
  transition:
    width 520ms var(--motion-ease-spring, cubic-bezier(0.34, 1.56, 0.64, 1)),
    left 520ms var(--motion-ease-spring, cubic-bezier(0.34, 1.56, 0.64, 1));
}

/* Tramo cerrado: rayado fino, el «aún no» de la anticipación mínima. */
.policy-preview__blocked {
  left: 0;
  background: repeating-linear-gradient(
    135deg,
    color-mix(in srgb, var(--color-on-strong) 28%, transparent) 0 2px,
    transparent 2px 6px
  );
}

/* Tramo abierto: latón lleno que se dibuja al entrar. */
.policy-preview__open {
  right: 0;
  background: color-mix(in srgb, var(--color-brand-accent-surface) 70%, transparent);
  transform-origin: left center;
  animation: policy-preview-draw 900ms var(--motion-easing-standard) 420ms both;
}

.policy-preview__now {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: 2px;
  background: var(--color-on-strong);
}

.policy-preview__legend {
  display: flex;
  justify-content: space-between;
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  letter-spacing: 0.08em;
  line-height: 14px;
  text-transform: uppercase;
}

/* --- Datos en texto --------------------------------------------------------- */

.policy-preview__facts {
  display: flex;
  flex-direction: column;
  margin: 0;
}

.policy-preview__fact {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 10px 0;
  border-top: var(--border-width-normal) solid var(--color-field-strong-border);
}

.policy-preview__fact dt {
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.08em;
  line-height: 14px;
  text-transform: uppercase;
}

.policy-preview__fact dd {
  margin: 0;
  font-size: var(--font-size-body-sm);
  line-height: 20px;
  color: var(--color-on-strong);
}

/* --- Franjas de ejemplo ------------------------------------------------------ */

.policy-preview__sample {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.policy-preview__sample-label {
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  letter-spacing: 0.08em;
  line-height: 14px;
  text-transform: uppercase;
}

.policy-preview__slots {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.policy-preview__slot {
  padding: 5px 10px;
  font-family: var(--font-display);
  font-size: var(--font-size-body);
  line-height: 20px;
  color: var(--color-on-strong);
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent);
  border-radius: 2px;
  animation: policy-preview-slot 360ms var(--motion-easing-standard) both;
  animation-delay: calc(var(--slot-index, 0) * 45ms);
}

@keyframes policy-preview-enter {
  from {
    opacity: 0;
    transform: translateX(16px);
  }

  to {
    opacity: 1;
    transform: none;
  }
}

@keyframes policy-preview-draw {
  from {
    clip-path: inset(0 100% 0 0);
  }

  to {
    clip-path: inset(0);
  }
}

@keyframes policy-preview-slot {
  from {
    opacity: 0;
    transform: translateY(6px) scale(0.94);
  }

  to {
    opacity: 1;
    transform: none;
  }
}

@keyframes policy-preview-pulse {
  0%,
  100% {
    opacity: 1;
    transform: rotate(45deg) scale(1);
  }

  50% {
    opacity: 0.45;
    transform: rotate(45deg) scale(0.7);
  }
}

@media (max-width: 1099px) {
  .policy-preview {
    position: static;
    animation-name: policy-preview-rise;
  }
}

@keyframes policy-preview-rise {
  from {
    opacity: 0;
    transform: translateY(14px);
  }

  to {
    opacity: 1;
    transform: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .policy-preview,
  .policy-preview__kicker::before,
  .policy-preview__open,
  .policy-preview__slot {
    animation: none;
  }

  .policy-preview__blocked,
  .policy-preview__open {
    transition: none;
  }
}
</style>
