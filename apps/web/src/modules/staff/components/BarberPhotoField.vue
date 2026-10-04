<script setup lang="ts">
/**
 * BarberPhotoField - Espacio para la fotografía del barbero dentro de los
 * diálogos de alta y edición (DEC-104). Muestra el retrato (o su monograma),
 * deja elegir un archivo o soltarlo sobre el marco, lo recorta al centro en un
 * cuadrado y lo reduce en el navegador (photo/preparePhoto.ts), y publica el
 * cambio pendiente como `PhotoDraft`. No habla con el servidor: quien lo monta
 * decide cuándo enviarlo (al pulsar «Guardar»).
 *
 * Accesibilidad: el marco es decorativo (el nombre del barbero ya está en el
 * formulario); todo se opera con botones reales. El estado ("nueva foto lista",
 * "se quitará al guardar") y los errores se anuncian en una región `aria-live`.
 * `prefers-reduced-motion` apaga el destello, el pulso y la escala del marco.
 */
import { computed, onBeforeUnmount, ref, useId, watch } from 'vue'
import { BarberAvatar, BaseButton } from '@/shared/ui'
import { NO_PHOTO_CHANGE, type PhotoDraft } from '../model/photoDraft'
import {
  PHOTO_ACCEPT_ATTRIBUTE,
  preparePhoto,
  type PreparePhotoFailure,
} from '../photo/preparePhoto'

interface Props {
  fullName: string
  /** Fotografía que el barbero ya tiene guardada (null si no tiene). */
  currentUrl: string | null
  draft: PhotoDraft
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), { disabled: false })

const emit = defineEmits<{ 'update:draft': [draft: PhotoDraft] }>()

const fileInputRef = ref<HTMLInputElement | null>(null)
const hintId = `${useId()}-hint`
const preparing = ref(false)
const dragging = ref(false)
const failure = ref<PreparePhotoFailure | null>(null)

const failureMessages: Record<PreparePhotoFailure, string> = {
  'unsupported-type': 'Ese archivo no es una imagen JPG, PNG o WebP.',
  'too-large': 'La imagen es demasiado pesada. Elige una de menos de 12 MB.',
  'too-small': 'La imagen es muy pequeña. Elige una de al menos 64 × 64 píxeles.',
  unreadable: 'No pudimos leer esa imagen. Prueba con otra.',
}

// Foto que se ve en el marco: la nueva elegida, ninguna si se pidió quitarla, o
// la guardada.
const shownUrl = computed<string | null>(() => {
  if (props.draft.kind === 'new') return props.draft.previewUrl
  if (props.draft.kind === 'remove') return null
  return props.currentUrl
})

const hasChange = computed(() => props.draft.kind !== 'none')
const canRemove = computed(() => shownUrl.value !== null)
const canUndo = computed(() => hasChange.value && props.currentUrl !== null)

const chooseLabel = computed(() => (shownUrl.value ? 'Cambiar foto' : 'Elegir foto'))

const statusMessage = computed(() => {
  if (failure.value) return failureMessages[failure.value]
  if (preparing.value) return 'Preparando la foto…'
  if (props.draft.kind === 'new') return 'Foto lista. Se guardará al pulsar «Guardar».'
  if (props.draft.kind === 'remove') return 'La foto se quitará al pulsar «Guardar».'
  return ''
})

function openPicker() {
  if (props.disabled || preparing.value) return
  fileInputRef.value?.click()
}

async function useFile(file: File) {
  if (props.disabled) return
  failure.value = null
  preparing.value = true
  const result = await preparePhoto(file)
  preparing.value = false

  if (result.kind === 'error') {
    failure.value = result.reason
    return
  }
  emit('update:draft', { kind: 'new', blob: result.blob, previewUrl: result.previewUrl })
}

async function onFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  // Se vacía para que elegir el mismo archivo otra vez vuelva a disparar change.
  input.value = ''
  if (file) await useFile(file)
}

function onDragOver(event: DragEvent) {
  if (props.disabled) return
  event.preventDefault()
  dragging.value = true
}

function onDragLeave() {
  dragging.value = false
}

async function onDrop(event: DragEvent) {
  event.preventDefault()
  dragging.value = false
  const file = event.dataTransfer?.files?.[0]
  if (file) await useFile(file)
}

function onRemove() {
  failure.value = null
  emit('update:draft', props.currentUrl ? { kind: 'remove' } : NO_PHOTO_CHANGE)
}

function onUndo() {
  failure.value = null
  emit('update:draft', NO_PHOTO_CHANGE)
}

// Las URL de vista previa son objetos del navegador: se liberan al reemplazarlas
// o descartarlas, y al desmontar el campo.
watch(
  () => props.draft,
  (_next, previous) => {
    if (previous?.kind === 'new') URL.revokeObjectURL(previous.previewUrl)
  },
)

onBeforeUnmount(() => {
  if (props.draft.kind === 'new') URL.revokeObjectURL(props.draft.previewUrl)
})
</script>

<template>
  <div class="barber-photo-field">
    <div
      class="barber-photo-field__frame"
      :class="{
        'barber-photo-field__frame--dragging': dragging,
        'barber-photo-field__frame--empty': !shownUrl,
        'barber-photo-field__frame--busy': preparing,
      }"
      aria-hidden="true"
      @dragover="onDragOver"
      @dragleave="onDragLeave"
      @drop="onDrop"
    >
      <!-- La clave reinicia la entrada animada cada vez que cambia la imagen
           mostrada (elegir, quitar, deshacer). -->
      <BarberAvatar
        :key="shownUrl ?? 'monogram'"
        class="barber-photo-field__portrait"
        size="hero"
        :full-name="fullName"
        :photo-url="shownUrl"
      />
      <span v-if="preparing" class="barber-photo-field__busy">
        <span class="barber-photo-field__busy-diamond" />
      </span>
      <span v-else-if="dragging" class="barber-photo-field__drop">Suelta la foto</span>
    </div>

    <div class="barber-photo-field__side">
      <p class="barber-photo-field__label">Foto</p>
      <p :id="hintId" class="barber-photo-field__hint">
        JPG, PNG o WebP. Se recorta en cuadrado desde el centro.
      </p>

      <div class="barber-photo-field__actions">
        <BaseButton
          type="button"
          variant="secondary"
          :disabled="disabled || preparing"
          :aria-describedby="hintId"
          @click="openPicker"
        >
          {{ chooseLabel }}
        </BaseButton>
        <BaseButton
          v-if="canRemove"
          type="button"
          variant="danger"
          :disabled="disabled || preparing"
          @click="onRemove"
        >
          Quitar foto
        </BaseButton>
        <BaseButton
          v-if="canUndo"
          type="button"
          variant="ghost"
          :disabled="disabled || preparing"
          @click="onUndo"
        >
          Deshacer
        </BaseButton>
      </div>

      <p
        class="barber-photo-field__status"
        :class="{ 'barber-photo-field__status--error': failure }"
        :role="failure ? 'alert' : 'status'"
        aria-live="polite"
      >
        {{ statusMessage }}
      </p>
    </div>

    <!-- Oculto (display: none): el botón visible lo abre. Fuera del orden de
         tabulación y del árbol de accesibilidad para no duplicar el control. -->
    <input
      ref="fileInputRef"
      class="barber-photo-field__input"
      type="file"
      :accept="PHOTO_ACCEPT_ATTRIBUTE"
      tabindex="-1"
      aria-hidden="true"
      data-testid="barber-photo-input"
      @change="onFileChange"
    />
  </div>
</template>

<style scoped>
.barber-photo-field {
  display: flex;
  align-items: center;
  gap: var(--space-5);
  padding: var(--space-4);
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: var(--radius-sm);
}

.barber-photo-field__input {
  display: none;
}

/* Marco: cuadrado hundido con esquinas de latón (los cuatro "cortes" del
   Tailored Grid). Vacío, lleva el borde punteado de una zona para soltar. */
.barber-photo-field__frame {
  position: relative;
  display: grid;
  flex: 0 0 auto;
  place-items: center;
  width: 136px;
  height: 136px;
  background-color: var(--color-surface-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: 2px;
  transition:
    border-color var(--motion-duration-base) var(--motion-easing-standard),
    background-color var(--motion-duration-base) var(--motion-easing-standard);
}

.barber-photo-field__frame--empty {
  border-style: dashed;
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent);
}

.barber-photo-field__frame--dragging {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
  border-style: solid;
  border-color: var(--color-brand-accent-surface);
}

.barber-photo-field__frame::before,
.barber-photo-field__frame::after {
  content: '';
  position: absolute;
  width: 14px;
  height: 14px;
  pointer-events: none;
  border: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
}

.barber-photo-field__frame::before {
  top: -4px;
  left: -4px;
  border-right: 0;
  border-bottom: 0;
}

.barber-photo-field__frame::after {
  right: -4px;
  bottom: -4px;
  border-top: 0;
  border-left: 0;
}

.barber-photo-field__portrait {
  animation: barber-photo-enter 320ms cubic-bezier(0.34, 1.56, 0.64, 1) both;
}

.barber-photo-field__busy,
.barber-photo-field__drop {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  background-color: rgb(16 27 43 / 72%);
  color: var(--color-on-strong);
  font-size: var(--font-size-caption);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.barber-photo-field__busy-diamond {
  width: 22px;
  height: 22px;
  background-color: var(--color-brand-accent-surface);
  transform: rotate(45deg);
  animation: barber-photo-spin 1s cubic-bezier(0.55, 0, 0.3, 1) infinite;
}

.barber-photo-field__side {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: var(--space-2);
}

.barber-photo-field__label {
  margin: 0;
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.barber-photo-field__hint {
  margin: 0;
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
}

.barber-photo-field__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: var(--space-1);
}

.barber-photo-field__status {
  min-height: var(--font-size-caption-line);
  margin: 0;
  color: var(--color-success-on-strong);
  font-size: var(--font-size-caption);
  line-height: var(--font-size-caption-line);
}

.barber-photo-field__status--error {
  color: var(--color-danger-on-strong);
}

/* Botones sobre tinta: mismo par latón-fantasma / rojo-fantasma que los
   diálogos de Servicios (BaseButton está calibrado para superficie clara). */
.barber-photo-field__actions :deep(.base-button) {
  height: 36px;
  padding-inline: 14px;
  font-size: var(--font-size-caption);
  --btn-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);
}

.barber-photo-field__actions :deep(.base-button--secondary) {
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.barber-photo-field__actions
  :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
}

.barber-photo-field__actions :deep(.base-button--danger) {
  background-color: transparent;
  color: var(--color-danger-on-strong);
  border-color: color-mix(in srgb, var(--color-danger-on-strong) 50%, transparent);
  border-bottom-color: var(--color-danger-on-strong);
}

.barber-photo-field__actions
  :deep(.base-button--danger:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-danger-on-strong) 12%, transparent);
  filter: none;
}

.barber-photo-field__actions :deep(.base-button--ghost) {
  background-color: transparent;
  color: var(--color-on-strong-muted);
  border-color: transparent;
}

.barber-photo-field__actions
  :deep(.base-button--ghost:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-on-strong) 8%, transparent);
  color: var(--color-on-strong);
}

@keyframes barber-photo-enter {
  from {
    opacity: 0;
    transform: scale(0.86);
  }

  to {
    opacity: 1;
    transform: scale(1);
  }
}

@keyframes barber-photo-spin {
  0% {
    transform: rotate(45deg) scale(1);
  }

  50% {
    transform: rotate(225deg) scale(0.7);
  }

  100% {
    transform: rotate(405deg) scale(1);
  }
}

@media (max-width: 480px) {
  .barber-photo-field {
    flex-direction: column;
    align-items: stretch;
    gap: var(--space-4);
  }

  .barber-photo-field__frame {
    align-self: center;
  }
}

@media (prefers-reduced-motion: reduce) {
  .barber-photo-field__portrait,
  .barber-photo-field__busy-diamond {
    animation: none;
  }

  .barber-photo-field__frame {
    transition: none;
  }
}
</style>
