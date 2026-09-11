<script lang="ts">
// Pila de diálogos abiertos, en un <script> de módulo (no de setup): se
// evalúa una sola vez y por eso el array se comparte entre TODAS las
// instancias de BaseDialog montadas. Escape y el atrapado de foco de un
// diálogo particular solo actúan cuando su id es el último de la pila —el
// visualmente más alto—, así dos diálogos abiertos a la vez (uno anidado
// desde el otro) no compiten por la misma tecla (auditoría HU-009:
// "evita colisiones con diálogos anidados o múltiples instancias").
export const openDialogStack: string[] = []
</script>

<script setup lang="ts">
/**
 * BaseDialog - Diálogo modal del sistema visual.
 * Con backdrop, focus trap, ESC para cerrar, click fuera para cerrar (opcional)
 * Tamaños: sm | md | lg | full (mobile)
 * Accesibilidad: role="dialog", aria-modal="true", aria-labelledby,
 * focus trap, focus restoration, focus-visible en elementos interactivos
 */
import { computed, ref, useId, onMounted, onUnmounted, watch, nextTick } from 'vue'

interface Props {
  /** Si el diálogo está abierto (v-model) */
  modelValue?: boolean
  /** Título del diálogo (para aria-labelledby) */
  title?: string
  /** Descripción adicional (para aria-describedby) */
  description?: string
  /** Tamaño del diálogo */
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full'
  /** Ubicación visual dentro del viewport. El foco y el backdrop no cambian. */
  placement?: 'center' | 'bottom'
  /** Clase opcional para una superficie de diálogo con contrato visual propio. */
  contentClass?: string
  /** Si se cierra al hacer click en el backdrop */
  closeOnBackdrop?: boolean
  /** Si se cierra con ESC */
  closeOnEscape?: boolean
  /** Si mostrar botón de cerrar en header */
  showClose?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  size: 'md',
  placement: 'center',
  contentClass: undefined,
  closeOnBackdrop: true,
  closeOnEscape: true,
  showClose: true,
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
  return [base, `${base}--${props.size}`, props.contentClass, isOpen.value ? `${base}--open` : '']
    .filter(Boolean)
    .join(' ')
})

const overlayClasses = computed(() => {
  const base = 'base-dialog__overlay'
  return [base, `${base}--${props.placement}`, isOpen.value ? `${base}--open` : '']
    .filter(Boolean)
    .join(' ')
})

// useId() (Vue 3.5+) genera un identificador estable y único por instancia,
// a diferencia de Math.random(): sin riesgo de colisión entre dos diálogos
// montados a la vez (auditoría HU-009).
const dialogId = `base-dialog-${useId()}`
const titleId = `${dialogId}-title`
const descriptionId = `${dialogId}-description`

// Solo el diálogo visualmente más alto (el último de openDialogStack)
// reacciona a Escape y atrapa Tab. Ningún componente inventa un z-index
// propio (estandar-diseno-visual.md §6.3): todos comparten --layer-dialog.
const isTopmost = () => openDialogStack[openDialogStack.length - 1] === dialogId

const updateFocusableElements = () => {
  if (!dialogRef.value) return
  const elements = dialogRef.value.querySelectorAll<HTMLElement>(
    'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"]), [contenteditable="true"]',
  )
  focusableElementsRef.value = Array.from(elements).filter(
    (el) => !el.hasAttribute('disabled') && !el.hidden && el.offsetParent !== null,
  )
}

const handleKeyDown = (event: KeyboardEvent) => {
  if (!isOpen.value || !isTopmost()) return

  if (event.key === 'Escape' && props.closeOnEscape) {
    event.preventDefault()
    close()
    return
  }

  if (event.key === 'Tab') {
    // Recalculado en cada Tab (no solo al abrir, HU-068): un campo
    // deshabilitado al momento de abrir (p. ej. un botón "Confirmar"
    // condicionado a que el formulario esté completo) puede habilitarse
    // mientras el diálogo sigue abierto; una lista cacheada del primer
    // cálculo seguiría excluyéndolo y el atrapamiento de foco envolvería
    // antes de llegar a él.
    updateFocusableElements()
    if (focusableElementsRef.value.length === 0) return
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
  // Todo el cuerpo, incluida updateFocusableElements(), debe esperar a
  // nextTick: watch(isOpen, ...) corre en flush "pre" (ANTES de que Vue
  // aplique el v-show que quita display:none del diálogo), así que una
  // consulta síncrona aquí encuentra 0 elementos con offsetParent !== null
  // (todos siguen ocultos) y el bloque completo -incluido el enfoque
  // inicial- nunca se ejecuta. Reproducido contra Chromium real (Playwright,
  // HU-021: el foco nunca entraba al diálogo pese a que las pruebas con
  // jsdom, menos estrictas con offsetParent/reflow, no lo detectaban).
  nextTick(() => {
    updateFocusableElements()
    if (focusableElementsRef.value.length > 0) {
      // Focus en el primer elemento enfocable, o en el botón de cerrar si existe
      const firstFocusable = closeButtonRef.value || focusableElementsRef.value[0]
      firstFocusable?.focus()
    }
  })
}

const restoreFocus = () => {
  if (previouslyFocusedElement.value) {
    ;(previouslyFocusedElement.value as HTMLElement).focus()
    previouslyFocusedElement.value = null
  }
}

// El bloqueo de scroll del body se cuenta por diálogos abiertos (largo de
// la pila), no por este único diálogo: si un segundo diálogo se cierra
// mientras el primero sigue abierto, el body debe seguir bloqueado.
const lockBodyScroll = () => {
  document.body.style.overflow = 'hidden'
}
const unlockBodyScrollIfNoneOpen = () => {
  if (openDialogStack.length === 0) {
    document.body.style.overflow = ''
  }
}

const pushToDialogStack = () => {
  if (!openDialogStack.includes(dialogId)) {
    openDialogStack.push(dialogId)
  }
}
const removeFromDialogStack = () => {
  const index = openDialogStack.indexOf(dialogId)
  if (index !== -1) {
    openDialogStack.splice(index, 1)
  }
}

watch(isOpen, (newValue) => {
  if (newValue) {
    // Guardar elemento enfocado anteriormente
    previouslyFocusedElement.value = document.activeElement as HTMLElement
    pushToDialogStack()
    lockBodyScroll()
    // Focus trap
    trapFocus()
  } else {
    removeFromDialogStack()
    unlockBodyScrollIfNoneOpen()
    // Restaurar focus
    restoreFocus()
  }
})

onMounted(() => {
  document.addEventListener('keydown', handleKeyDown)
  if (isOpen.value) {
    pushToDialogStack()
    trapFocus()
    lockBodyScroll()
  }
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeyDown)
  if (isOpen.value) {
    removeFromDialogStack()
    unlockBodyScrollIfNoneOpen()
  }
})
</script>

<template>
  <Teleport to="body">
    <div
      v-show="isOpen"
      ref="overlayRef"
      :class="overlayClasses"
      @click="handleBackdropClick"
      @keydown="handleKeyDown"
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
          <!-- div, no <header>: dentro de [role="dialog"] un <header> sigue
               resolviendo como landmark "banner" (el hueco de la lista de
               excepciones de la spec HTML no incluye role="dialog"), y con
               AppHeader ya presente en la página produce dos landmarks
               "banner" sin nombre único (axe-core landmark-unique). El
               diálogo ya se identifica por su propio role="dialog" +
               aria-labelledby; este contenedor no necesita ser landmark. -->
          <div v-if="title || showClose || $slots.header" class="base-dialog__header">
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
          </div>

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
  /* Sin backdrop-filter: blur() — el estándar visual prohíbe desenfoques
   * decorativos (estandar-diseno-visual.md §6.3). El fondo no interactivo
   * es solo un scrim translúcido. */
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-4);
  background-color: var(--color-overlay-scrim);
  z-index: var(--layer-dialog);
  opacity: 0;
  visibility: hidden;
  transition:
    opacity var(--motion-duration-base) var(--motion-easing-standard),
    visibility var(--motion-duration-base) var(--motion-easing-standard);
}

.base-dialog__overlay--bottom {
  align-items: flex-end;
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
  z-index: var(--layer-dialog);
  transform: scale(0.95) translateY(8px);
  opacity: 0;
  transition:
    transform var(--motion-duration-base) var(--motion-easing-standard),
    opacity var(--motion-duration-base) var(--motion-easing-standard);
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

.base-dialog--xl .base-dialog__container {
  width: min(100%, 1092px);
}

.base-dialog--xl {
  width: min(100%, 1092px);
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
  /* Botón de icono independiente (estandar-diseno-visual.md §6.2): 44×44,
   * no los 36×36 anteriores. CA-009-03. */
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: var(--control-height-icon);
  height: var(--control-height-icon);
  padding: 0;
  margin: calc(-1 * var(--space-2));
  background: transparent;
  border: none;
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition:
    background-color var(--motion-duration-fast) var(--motion-easing-standard),
    color var(--motion-duration-fast) var(--motion-easing-standard);
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
  box-shadow:
    0 0 0 2px var(--color-surface),
    0 0 0 4px var(--color-focus);
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
