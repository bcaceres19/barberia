<script setup lang="ts">
/**
 * BaseDialog - Diálogo modal del sistema visual.
 * Con backdrop, focus trap, ESC para cerrar, click fuera para cerrar (opcional)
 * Tamaños: sm | md | lg | full (mobile)
 * Accesibilidad: role="dialog", aria-modal="true", aria-labelledby,
 * focus trap, focus restoration, focus-visible en elementos interactivos
 */
import { computed, ref, onMounted, onUnmounted, watch, nextTick } from 'vue'

interface Props {
  /** Si el diálogo está abierto (v-model) */
  modelValue?: boolean
  /** Título del diálogo (para aria-labelledby) */
  title?: string
  /** Descripción adicional (para aria-describedby) */
  description?: string
  /** Tamaño del diálogo */
  size?: 'sm' | 'md' | 'lg' | 'full'
  /** Si se cierra al hacer click en el backdrop */
  closeOnBackdrop?: boolean
  /** Si se cierra con ESC */
  closeOnEscape?: boolean
  /** Si mostrar botón de cerrar en header */
  showClose?: boolean
  /** Clases CSS adicionales */
  class?: string
  /** Z-index personalizado */
  zIndex?: number
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  size: 'md',
  closeOnBackdrop: true,
  closeOnEscape: true,
  showClose: true,
  class: '',
  zIndex: 40,
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  close: []
  open: []
}>()

const dialogRef = ref<HTMLDivElement | null>(null)
const overlayRef = ref<HTMLDivElement | null>(null)
const closeButtonRef = ref<HTMLButtonElement | null>(null)
const previouslyFocusedElement = ref<HTMLElement | null>(null)
const focusableElementsRef = ref<HTMLElement[]>([])

const isOpen = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})

const classes = computed(() => {
  const base = 'base-dialog'
  return [
    base,
    `${base}--${props.size}`,
    isOpen.value ? `${base}--open` : '',
    props.class,
  ]
    .filter(Boolean)
    .join(' ')
})

const overlayClasses = computed(() => {
  const base = 'base-dialog__overlay'
  return [
    base,
    isOpen.value ? `${base}--open` : '',
  ]
    .filter(Boolean)
    .join(' ')
})

const dialogId = `base-dialog-${Math.random().toString(36).slice(2, 9)}`
const titleId = `${dialogId}-title`
const descriptionId = `${dialogId}-description`

const style = computed(() => ({
  '--dialog-z-index': props.zIndex,
}))

const updateFocusableElements = () => {
  if (!dialogRef.value) return
  const elements = dialogRef.value.querySelectorAll<HTMLElement>(
    'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"]), [contenteditable="true"]'
  )
  focusableElementsRef.value = Array.from(elements).filter(
    el => !el.hasAttribute('disabled') && !el.hidden && el.offsetParent !== null
  )
}

const handleKeyDown = (event: KeyboardEvent) => {
  if (!isOpen.value) return

  if (event.key === 'Escape' && props.closeOnEscape) {
    event.preventDefault()
    close()
    return
  }

  if (event.key === 'Tab' && focusableElementsRef.value.length > 0) {
    const first = focusableElementsRef.value[0]
    const last = focusableElementsRef.value[focusableElementsRef.value.length - 1]

    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault()
      last.focus()
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault()
      first.focus()
    }
  }
}

const handleBackdropClick = (event: MouseEvent) => {
  if (props.closeOnBackdrop && event.target === overlayRef.value) {
    close()
  }
}

const handleCloseClick = () => {
  close()
}

const close = () => {
  isOpen.value = false
  emit('close')
}

/** Abre el diálogo (exponido para uso imperativo) */
const open = () => {
  isOpen.value = true
  emit('open')
}

// Exponer métodos para uso imperativo y tests
defineExpose({
  close,
  open,
})

const trapFocus = () => {
  updateFocusableElements()
  if (focusableElementsRef.value.length > 0) {
    // Focus en el primer elemento enfocable, o en el botón de cerrar si existe
    const firstFocusable = closeButtonRef.value || focusableElementsRef.value[0]
    nextTick(() => {
      firstFocusable?.focus()
    })
  }
}

const restoreFocus = () => {
  if (previouslyFocusedElement.value) {
    (previouslyFocusedElement.value as HTMLElement).focus()
    previouslyFocusedElement.value = null
  }
}

watch(isOpen, (newValue) => {
  if (newValue) {
    // Guardar elemento enfocado anteriormente
    previouslyFocusedElement.value = document.activeElement as HTMLElement
    // Prevenir scroll en body
    document.body.style.overflow = 'hidden'
    // Focus trap
    trapFocus()
  } else {
    // Restaurar scroll
    document.body.style.overflow = ''
    // Restaurar focus
    restoreFocus()
  }
})

onMounted(() => {
  document.addEventListener('keydown', handleKeyDown)
  if (isOpen.value) {
    trapFocus()
    document.body.style.overflow = 'hidden'
  }
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeyDown)
  document.body.style.overflow = ''
})
</script>

<template>
  <Teleport to="body">
    <div
      v-show="isOpen"
      ref="overlayRef"
      :class="overlayClasses"
      :style="style"
      @click="handleBackdropClick"
      @keydown="handleKeyDown"
      aria-hidden="true"
    >
      <div
        v-show="isOpen"
        ref="dialogRef"
        :class="classes"
        :id="dialogId"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="title ? titleId : undefined"
        :aria-describedby="description ? descriptionId : undefined"
      >
        <div class="base-dialog__container">
          <header v-if="title || showClose || $slots.header" class="base-dialog__header">
            <slot name="header">
              <h2 v-if="title" :id="titleId" class="base-dialog__title">{{ title }}</h2>
              <button
                v-if="showClose"
                ref="closeButtonRef"
                type="button"
                class="base-dialog__close"
                @click="handleCloseClick"
                aria-label="Cerrar diálogo"
              >
                <svg
                  class="base-dialog__close-icon"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  aria-hidden="true"
                >
                  <line x1="18" y1="6" x2="6" y2="18"></line>
                  <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
              </button>
            </slot>
          </header>

          <div v-if="description" :id="descriptionId" class="base-dialog__description">
            {{ description }}
          </div>

          <div class="base-dialog__content">
            <slot />
          </div>

          <footer v-if="$slots.footer" class="base-dialog__footer">
            <slot name="footer" />
          </footer>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.base-dialog__overlay {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-4);
  background-color: rgb(15 23 42 / 48%);
  backdrop-filter: blur(2px);
  z-index: var(--dialog-z-index, var(--layer-dialog));
  opacity: 0;
  visibility: hidden;
  transition: opacity 0.2s ease, visibility 0.2s ease;
}

@media (prefers-reduced-motion: reduce) {
  .base-dialog__overlay {
    transition: none;
  }
}

.base-dialog__overlay--open {
  opacity: 1;
  visibility: visible;
}

.base-dialog {
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - var(--space-4) * 2);
  background-color: var(--color-surface);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-dialog);
  z-index: var(--dialog-z-index, var(--layer-dialog));
  transform: scale(0.95) translateY(8px);
  opacity: 0;
  transition: transform 0.2s ease, opacity 0.2s ease;
}

@media (prefers-reduced-motion: reduce) {
  .base-dialog {
    transition: none;
  }
}

.base-dialog--open {
  transform: scale(1) translateY(0);
  opacity: 1;
}

/* Tamaños */
.base-dialog--sm .base-dialog__container {
  width: 100%;
  max-width: 360px;
}

.base-dialog--md .base-dialog__container {
  width: 100%;
  max-width: 520px;
}

.base-dialog--lg .base-dialog__container {
  width: 100%;
  max-width: 720px;
}

.base-dialog--full .base-dialog__container {
  width: 100%;
  max-width: 100%;
  max-height: 100vh;
  border-radius: 0;
  height: 100vh;
}

@media (min-width: 768px) {
  .base-dialog--full .base-dialog__container {
    max-width: 720px;
    max-height: calc(100vh - var(--space-4) * 2);
    border-radius: var(--radius-lg);
    height: auto;
  }
}

.base-dialog__container {
  display: flex;
  flex-direction: column;
  width: 100%;
  max-height: inherit;
  overflow: hidden;
}

.base-dialog__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-4) var(--space-5);
  border-bottom: var(--border-width-normal) solid var(--color-border-subtle);
  flex-shrink: 0;
}

.base-dialog__title {
  margin: 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-h3);
  font-weight: var(--font-weight-h3);
  line-height: var(--font-size-h3-line);
  color: var(--color-text-primary);
}

.base-dialog__close {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  padding: 0;
  background: transparent;
  border: none;
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: background-color 0.12s ease, color 0.12s ease;
}

@media (prefers-reduced-motion: reduce) {
  .base-dialog__close {
    transition: none;
  }
}

.base-dialog__close:hover {
  background-color: var(--color-surface-muted);
  color: var(--color-text-primary);
}

.base-dialog__close:focus-visible {
  outline: none;
  box-shadow: 0 0 0 2px var(--color-surface), 0 0 0 4px var(--color-focus);
}

.base-dialog__close-icon {
  width: 20px;
  height: 20px;
}

.base-dialog__description {
  padding: var(--space-2) var(--space-5) 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
  color: var(--color-text-secondary);
  flex-shrink: 0;
}

.base-dialog__content {
  flex: 1;
  overflow: auto;
  padding: var(--space-5);
  min-height: 0;
}

.base-dialog__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-3);
  padding: var(--space-4) var(--space-5);
  border-top: var(--border-width-normal) solid var(--color-border-subtle);
  flex-shrink: 0;
  flex-wrap: wrap;
}

/* Scrollbar styling */
.base-dialog__content::-webkit-scrollbar {
  width: 8px;
}

.base-dialog__content::-webkit-scrollbar-track {
  background: transparent;
}

.base-dialog__content::-webkit-scrollbar-thumb {
  background-color: var(--color-border-subtle);
  border-radius: 4px;
}

.base-dialog__content::-webkit-scrollbar-thumb:hover {
  background-color: var(--color-border-control);
}
</style>