<script setup lang="ts">
/**
 * EmptyState - Estado "sin datos todavía" del sistema visual NAVA / Tailored
 * Grid (estandar-diseno-visual.md §6.6, atlas: círculo decorativo + mensaje
 * + siguiente acción útil). Solo para una colección genuinamente vacía (por
 * ejemplo, "aún no tienes barberos"), no para "sin resultados para este
 * filtro" (esos siguen siendo texto simple junto al control que filtra).
 */
withDefaults(
  defineProps<{
    /** Mensaje principal. */
    message: string
  }>(),
  {},
)
</script>

<template>
  <div class="empty-state">
    <span class="empty-state__icon" aria-hidden="true">
      <slot name="icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <circle cx="12" cy="8" r="4" />
          <path d="M4 21c0-4 3.6-7 8-7s8 3 8 7" />
        </svg>
      </slot>
    </span>
    <p class="empty-state__message">{{ message }}</p>
    <div v-if="$slots.action" class="empty-state__action">
      <slot name="action" />
    </div>
  </div>
</template>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-8) var(--space-4);
  text-align: center;
}

.empty-state__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background-color: var(--color-surface-muted);
  color: var(--color-text-secondary);
}

.empty-state__icon svg {
  width: 32px;
  height: 32px;
}

.empty-state__message {
  margin: 0;
  max-width: 40ch;
  color: var(--color-text-secondary);
  font-size: var(--font-size-body);
}
</style>
