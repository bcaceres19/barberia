<script setup lang="ts">
// Barra de guardado flotante (DEC-110): aparece al pie de la columna cuando hay
// cambios de la barbería sin guardar, con un rombo que late, el resumen de qué
// cambió y las dos acciones. Mientras no hay nada pendiente no ocupa espacio ni
// entra en el orden de foco (`v-if`). El resultado se anuncia con
// `role="status"` y el rombo es decorativo.
import { BaseButton } from '@/shared/ui'

defineProps<{
  /** Resumen legible de lo pendiente ("Marca y vocabulario"). */
  summary: string
  saving: boolean
}>()
const emit = defineEmits<{ discard: []; save: [] }>()
</script>

<template>
  <div class="save-bar" role="status" aria-live="polite">
    <p class="save-bar__message">
      <span class="save-bar__diamond" aria-hidden="true" />
      <span class="save-bar__text">
        <strong>Cambios sin guardar</strong>
        <span class="save-bar__summary">{{ summary }}</span>
      </span>
    </p>
    <div class="save-bar__actions">
      <BaseButton type="button" variant="secondary" :disabled="saving" @click="emit('discard')">
        Descartar
      </BaseButton>
      <BaseButton
        type="button"
        variant="primary"
        :loading="saving"
        :disabled="saving"
        @click="emit('save')"
      >
        Guardar cambios
      </BaseButton>
    </div>
  </div>
</template>

<style scoped>
.save-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 16px 12px 20px;
  background: var(--color-field-strong-raised);
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
  border-left: 3px solid var(--color-brand-accent-surface);
  border-radius: 3px;
  box-shadow: 0 -10px 30px -12px rgb(16 27 43 / 55%);
  animation: save-bar-enter 380ms cubic-bezier(0.34, 1.56, 0.64, 1) both;
}

.save-bar__message {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0;
  min-width: 0;
}

.save-bar__diamond {
  flex: 0 0 auto;
  width: 8px;
  height: 8px;
  background: var(--color-brand-accent-surface);
  transform: rotate(45deg);
  animation: save-bar-pulse 2.4s ease-in-out infinite;
}

.save-bar__text {
  display: flex;
  flex-direction: column;
  min-width: 0;
  font-size: var(--font-size-caption);
  line-height: 17px;
  color: var(--color-on-strong-muted);
}

.save-bar__text strong {
  font-size: var(--font-size-body-sm);
  font-weight: 600;
  color: var(--color-on-strong);
}

.save-bar__actions {
  display: flex;
  flex: 0 0 auto;
  gap: 10px;
}

.save-bar :deep(.base-button) {
  --btn-focus-ring: 0 0 0 2px var(--color-field-strong-raised), 0 0 0 4px var(--color-focus);

  height: 36px;
  padding-inline: 16px;
  font-size: var(--font-size-caption);
}

.save-bar :deep(.base-button--primary) {
  background-color: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  border-color: var(--color-brand-accent-surface);
}

.save-bar :deep(.base-button--primary:hover:not(:disabled):not(.base-button--loading)) {
  filter: brightness(92%);
}

.save-bar :deep(.base-button--primary:active:not(:disabled):not(.base-button--loading)) {
  filter: brightness(84%);
}

.save-bar :deep(.base-button--secondary) {
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.save-bar :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
}

@keyframes save-bar-enter {
  from {
    opacity: 0;
    transform: translateY(24px);
  }

  to {
    opacity: 1;
    transform: none;
  }
}

@keyframes save-bar-pulse {
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

@media (max-width: 640px) {
  .save-bar {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
    padding: 12px;
  }

  .save-bar__actions :deep(.base-button) {
    flex: 1;
    height: 44px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .save-bar,
  .save-bar__diamond {
    animation: none;
  }
}
</style>
