<script setup lang="ts">
/**
 * BarberSelect - Selector de barbero con retrato (issue #189, atlas
 * panel-agenda-eventos evento 12): el `<select>` nativo no puede mostrar un
 * `BarberAvatar` por opción, así que esta pantalla necesita un listbox
 * propio. Local al módulo `agenda`: solo `BarberAvatar` se promueve a
 * shared/ui, este widget de selección le pertenece únicamente a `/panel`.
 *
 * No decide enrutamiento: emite `update:modelValue` con el `barberId`
 * elegido y quien lo monta (DailyAgendaPage.vue) sigue siendo la única
 * fuente de verdad de `route.query.barberId` (onBarberSelectChange).
 *
 * Patrón ARIA "Collapsible Listbox" (WAI-ARIA APG): botón disparador con
 * aria-haspopup="listbox"/aria-expanded, lista role="listbox" con opciones
 * role="option" y tabindex reglado a mano (roving tabindex) — el navegador
 * no ofrece un listbox nativo con contenido enriquecido por opción.
 */
import {
  computed,
  nextTick,
  onMounted,
  onUnmounted,
  ref,
  useId,
  type ComponentPublicInstance,
} from 'vue'
import BarberAvatar from '@/shared/ui/BarberAvatar.vue'
import type { BarberSummary } from '../model/appointment'

interface Props {
  modelValue: string | null
  barbers: BarberSummary[]
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), { disabled: false })

const emit = defineEmits<{
  'update:modelValue': [barberId: string]
}>()

const listboxId = useId()
const rootRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const optionRefs = ref<HTMLElement[]>([])
const open = ref(false)
const activeIndex = ref(0)

const selectedBarber = computed(() => props.barbers.find((b) => b.id === props.modelValue) ?? null)

function setOptionRef(el: Element | ComponentPublicInstance | null, index: number) {
  if (el instanceof HTMLElement) optionRefs.value[index] = el
}

async function openList() {
  if (props.disabled || props.barbers.length === 0) return
  const currentIndex = props.barbers.findIndex((b) => b.id === props.modelValue)
  activeIndex.value = currentIndex >= 0 ? currentIndex : 0
  open.value = true
  await nextTick()
  optionRefs.value[activeIndex.value]?.focus()
}

function closeList(returnFocus = true) {
  open.value = false
  if (returnFocus) triggerRef.value?.focus()
}

function toggle() {
  if (open.value) closeList()
  else void openList()
}

function moveActive(delta: number) {
  const count = props.barbers.length
  if (count === 0) return
  activeIndex.value = (activeIndex.value + delta + count) % count
  optionRefs.value[activeIndex.value]?.focus()
}

function selectActive() {
  const barber = props.barbers[activeIndex.value]
  if (!barber) return
  emit('update:modelValue', barber.id)
  closeList()
}

function selectBarber(barberId: string) {
  emit('update:modelValue', barberId)
  closeList()
}

function onTriggerKeydown(event: KeyboardEvent) {
  if (event.key === 'ArrowDown' || event.key === 'ArrowUp' || event.key === 'Enter') {
    event.preventDefault()
    void openList()
  }
}

function onListKeydown(event: KeyboardEvent) {
  switch (event.key) {
    case 'ArrowDown':
      event.preventDefault()
      moveActive(1)
      break
    case 'ArrowUp':
      event.preventDefault()
      moveActive(-1)
      break
    case 'Home':
      event.preventDefault()
      activeIndex.value = 0
      optionRefs.value[0]?.focus()
      break
    case 'End':
      event.preventDefault()
      activeIndex.value = props.barbers.length - 1
      optionRefs.value[activeIndex.value]?.focus()
      break
    case 'Enter':
    case ' ':
      event.preventDefault()
      selectActive()
      break
    case 'Escape':
      event.preventDefault()
      closeList()
      break
    case 'Tab':
      closeList(false)
      break
  }
}

// Cierre al hacer click fuera: el widget no tiene contrafondo modal (no es
// un diálogo, BaseDialog.vue), así que escucha directamente en document,
// mismo criterio que el keydown de document en BaseAlert.vue.
function onDocumentClick(event: MouseEvent) {
  if (!open.value) return
  if (rootRef.value && !rootRef.value.contains(event.target as Node)) closeList(false)
}

onMounted(() => document.addEventListener('mousedown', onDocumentClick))
onUnmounted(() => document.removeEventListener('mousedown', onDocumentClick))
</script>

<template>
  <div ref="rootRef" class="barber-select">
    <button
      id="daily-agenda-barber-select"
      ref="triggerRef"
      type="button"
      class="barber-select__trigger"
      :disabled="disabled || barbers.length === 0"
      aria-haspopup="listbox"
      :aria-expanded="open"
      :aria-controls="listboxId"
      @click="toggle"
      @keydown="onTriggerKeydown"
    >
      <BarberAvatar v-if="selectedBarber" :full-name="selectedBarber.fullName" size="closed" />
      <span v-else class="barber-select__trigger-icon" aria-hidden="true" />
      <span
        class="barber-select__trigger-label"
        :class="{ 'barber-select__trigger-label--placeholder': !selectedBarber }"
      >
        {{ selectedBarber?.fullName ?? 'Selecciona un barbero' }}
      </span>
      <span class="barber-select__chevron" aria-hidden="true" />
    </button>

    <ul
      v-show="open"
      :id="listboxId"
      class="barber-select__list"
      role="listbox"
      aria-label="Barbero"
      @keydown="onListKeydown"
    >
      <li
        v-for="(barber, index) in barbers"
        :key="barber.id"
        :ref="(el) => setOptionRef(el, index)"
        class="barber-select__option"
        :class="{ 'barber-select__option--selected': barber.id === modelValue }"
        role="option"
        :aria-selected="barber.id === modelValue"
        tabindex="-1"
        @click="selectBarber(barber.id)"
      >
        <BarberAvatar :full-name="barber.fullName" size="option" />
        <span class="barber-select__option-name">{{ barber.fullName }}</span>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.barber-select {
  position: relative;
}

/* Control reglado sobre tinta (atlas panel-agenda-eventos): caja tenue con
   línea base de latón de doble peso, el mismo lenguaje que BaseInput. */
.barber-select__trigger {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  min-height: 44px;
  padding: var(--space-2) var(--space-3);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-on-strong);
  text-align: left;
  background-color: rgb(244 240 231 / 5%);
  border: var(--border-width-normal) solid rgb(244 240 231 / 16%);
  border-bottom: var(--border-width-emphasis) solid var(--color-accent-brass);
  border-radius: 2px;
  cursor: pointer;
}

.barber-select__trigger[aria-expanded='true'] {
  border-bottom-color: var(--color-accent-brass);
}

.barber-select__trigger:disabled {
  cursor: not-allowed;
  opacity: 0.42;
  border-bottom-color: rgb(244 240 231 / 28%);
}

.barber-select__trigger:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus);
}

.barber-select__trigger-icon {
  width: 28px;
  height: 28px;
  flex-shrink: 0;
  background-color: #16243a;
  border: var(--border-width-normal) solid rgb(184 149 90 / 55%);
  border-radius: 2px;
}

.barber-select__trigger-label {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Sin barbero resuelto (evento 08 del atlas): el placeholder se distingue
   por color, no solo por su texto. */
.barber-select__trigger-label--placeholder {
  color: var(--color-on-strong-muted);
}

.barber-select__chevron {
  flex-shrink: 0;
  width: 8px;
  height: 8px;
  border-right: var(--border-width-normal) solid var(--color-accent-brass);
  border-bottom: var(--border-width-normal) solid var(--color-accent-brass);
  transform: rotate(45deg) translateY(-2px);
}

.barber-select__list {
  position: absolute;
  z-index: var(--layer-menu);
  top: calc(100% - 1px);
  left: 0;
  width: 100%;
  padding: 0;
  margin: 0;
  list-style: none;
  background-color: #16243a;
  border: var(--border-width-normal) solid rgb(244 240 231 / 16%);
  border-top: var(--border-width-emphasis) solid var(--color-accent-brass);
  border-radius: 2px;
  box-shadow: var(--shadow-dialog);
}

.barber-select__option {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  min-height: 44px;
  padding: var(--space-2) var(--space-3);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-on-strong);
  cursor: pointer;
}

/* Filete tenue entre opciones, no un borde perimetral por fila. */
.barber-select__option + .barber-select__option {
  border-top: var(--border-width-normal) solid rgb(244 240 231 / 8%);
}

.barber-select__option:hover,
.barber-select__option:focus-visible {
  outline: none;
  background-color: var(--color-overlay-hover);
}

.barber-select__option--selected {
  background-color: rgb(184 149 90 / 12%);
}

.barber-select__option--selected .barber-select__option-name {
  font-weight: 600;
}

.barber-select__option-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
