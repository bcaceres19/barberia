<script setup lang="ts">
// app compone las dos capacidades por sus índices públicos; ningún módulo
// conoce los componentes privados de su hermano (DEC-105, issue #286).
import { ref } from 'vue'
import { StaffPage } from '@/modules/staff'
import { BarberBlocksPanel } from '@/modules/schedules'
import { barberPhotoUrl } from '@/shared/api'
import { useVocabulary } from '@/shared/composables'
import { BaseButton, BaseDialog } from '@/shared/ui'

// Palabras de la barbería (DEC-110).
const v = useVocabulary()
const isBlocksOpen = ref(false)
const selectedBarber = ref<{ id: string; name: string; photoUrl: string | null }>()

function openBlocks(id: string, name: string, photoUrl: string | null) {
  selectedBarber.value = { id, name, photoUrl }
  isBlocksOpen.value = true
}
</script>

<template>
  <StaffPage>
    <template #barber-actions="{ barber }">
      <BaseButton
        variant="secondary"
        class="block-action"
        :aria-label="`Bloquear a ${barber.fullName}`"
        aria-haspopup="dialog"
        @click="openBlocks(barber.id, barber.fullName, barberPhotoUrl(barber))"
      >
        <svg
          class="block-action-icon"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
          aria-hidden="true"
        >
          <rect x="5" y="10" width="14" height="11" rx="1" />
          <path d="M8 10V6a4 4 0 0 1 8 0v4" />
        </svg>
        Bloquear
      </BaseButton>
    </template>
  </StaffPage>
  <BaseDialog
    v-model="isBlocksOpen"
    :title="`Bloqueos de ${selectedBarber?.name ?? v.professional}`"
    size="xl"
    content-class="blocks-workspace-dialog"
    :close-on-backdrop="false"
  >
    <BarberBlocksPanel
      v-if="isBlocksOpen && selectedBarber"
      :key="selectedBarber.id"
      :barber-id="selectedBarber.id"
      :barber-name="selectedBarber.name"
      :photo-url="selectedBarber.photoUrl"
    />
  </BaseDialog>
</template>

<style scoped>
.block-action.base-button.base-button--secondary {
  background-color: transparent;
  color: var(--color-danger-on-strong);
  border-color: color-mix(in srgb, var(--color-danger-on-strong) 50%, transparent);
  border-bottom-color: var(--color-danger-on-strong);
}
.block-action.base-button.base-button--secondary:hover:not(:disabled):not(.base-button--loading) {
  background-color: color-mix(in srgb, var(--color-danger-on-strong) 12%, transparent);
  border-color: color-mix(in srgb, var(--color-danger-on-strong) 50%, transparent);
  border-bottom-color: var(--color-danger-on-strong);
}
.block-action.base-button.base-button--secondary:active:not(:disabled):not(.base-button--loading) {
  background-color: color-mix(in srgb, var(--color-danger-on-strong) 20%, transparent);
}

.block-action :deep(.base-button__content) {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.block-action-icon {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
}
</style>

<style>
/* Superficie propia del diálogo teletransportado; conserva foco, pila y
   movimiento reducido del componente compartido (DEC-105, #286). */
.blocks-workspace-dialog {
  --color-text-primary: var(--color-on-strong);
  --color-text-secondary: var(--color-on-strong-muted);
  --color-surface: var(--color-surface-strong);
  --color-border-subtle: var(--color-field-strong-border);
  width: min(100%, 1092px);
  border: 1px solid var(--color-field-strong-border);
  border-top: 2px solid var(--color-brand-accent-surface);
  border-radius: 2px;
}
.blocks-workspace-dialog .base-dialog__header {
  padding: 16px 28px;
}
.blocks-workspace-dialog .base-dialog__title {
  font-size: 14px;
  font-weight: 500;
  line-height: 1.5;
  overflow-wrap: anywhere;
}
.blocks-workspace-dialog .base-dialog__content {
  padding: 0;
  overscroll-behavior: contain;
}
.blocks-workspace-dialog .blocks-panel {
  border: 0;
}
@media (max-width: 767px) {
  .blocks-workspace-dialog .base-dialog__header {
    padding: 12px 16px;
  }
}
</style>
