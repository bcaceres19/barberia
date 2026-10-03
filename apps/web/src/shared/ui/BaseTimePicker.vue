<script setup lang="ts">
import { computed, nextTick, ref, useId, watch } from 'vue'
import { usePickerPopover } from '@/shared/composables/usePickerPopover'

const props = withDefaults(
  defineProps<{
    modelValue: string
    label: string
    id?: string
    disabled?: boolean
    required?: boolean
  }>(),
  { id: undefined, disabled: false, required: false },
)
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()
const uid = useId()
const triggerId = computed(() => props.id ?? `time-picker-${uid}`)
const labelId = `time-label-${uid}`
const dialogId = `time-dialog-${uid}`
const titleId = `time-title-${uid}`
const hintId = `time-hint-${uid}`
const root = ref<HTMLElement | null>(null)
const trigger = ref<HTMLButtonElement | null>(null)
const panel = ref<HTMLElement | null>(null)
const open = ref(false)
type Part = 'hour' | 'minute'
const hourText = ref('09')
const minuteText = ref('00')
const pad = (value: number) => String(value).padStart(2, '0')
const validPart = (text: string, max: number) => /^\d{1,2}$/.test(text) && Number(text) <= max
const validHour = computed(() => validPart(hourText.value, 23))
const validMinute = computed(() => validPart(minuteText.value, 59))
const validTime = computed(() => validHour.value && validMinute.value)
const hour = computed(() => Number(hourText.value))
const minute = computed(() => Number(minuteText.value))
const labelledBy = computed(() => `${labelId} ${triggerId.value}`)
const { style } = usePickerPopover(open, root, trigger, panel, close)

function choose(part: Part, value: number) {
  if (part === 'hour') hourText.value = pad(value)
  else minuteText.value = pad(value)
}
function edit(part: Part, event: Event) {
  const text = (event.target as HTMLInputElement).value
  if (part === 'hour') hourText.value = text
  else minuteText.value = text
  if (validPart(text, part === 'hour' ? 23 : 59)) {
    if (part === 'hour' && text.length === 2) {
      const input = panel.value?.querySelector<HTMLInputElement>(`#minute-input-${uid}`)
      input?.focus({ preventScroll: true })
      input?.select()
    }
  }
}
function partKeydown(part: Part, event: KeyboardEvent) {
  const max = part === 'hour' ? 23 : 59
  const text = part === 'hour' ? hourText.value : minuteText.value
  let value = validPart(text, max) ? Number(text) : 0
  if (event.key === 'ArrowDown') value -= 1
  else if (event.key === 'ArrowUp') value += 1
  else if (event.key === 'PageDown') value -= 5
  else if (event.key === 'PageUp') value += 5
  else if (event.key === 'Home') value = 0
  else if (event.key === 'End') value = max
  else if (event.key === 'Enter') {
    event.preventDefault()
    apply()
    return
  } else return
  event.preventDefault()
  choose(part, (value + max + 1) % (max + 1))
}
async function show() {
  if (props.disabled) return
  const current = /^(?:[01]\d|2[0-3]):[0-5]\d$/.test(props.modelValue) ? props.modelValue : '09:00'
  hourText.value = current.slice(0, 2)
  minuteText.value = current.slice(3)
  open.value = true
  await nextTick()
  const input = panel.value?.querySelector<HTMLInputElement>('input')
  input?.focus({ preventScroll: true })
  input?.select()
}
function close(restoreFocus = true) {
  open.value = false
  if (restoreFocus) trigger.value?.focus()
}
function apply() {
  if (!validTime.value || props.disabled) return
  emit('update:modelValue', `${pad(hour.value)}:${pad(minute.value)}`)
  close()
}
function dialogKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    event.preventDefault()
    event.stopPropagation()
    close()
  }
  if (event.key !== 'Tab' || !panel.value) return
  event.stopPropagation()
  const focusables = [
    ...panel.value.querySelectorAll<HTMLElement>(
      'button:not(:disabled):not([tabindex="-1"]), input',
    ),
  ]
  const first = focusables[0]
  const last = focusables.at(-1)
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault()
    last?.focus()
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault()
    first?.focus()
  }
}
watch(
  () => props.disabled,
  (disabled) => {
    if (disabled) close(false)
  },
)
</script>

<template>
  <div ref="root" class="base-time-picker" :aria-owns="open ? dialogId : undefined">
    <label :id="labelId" :for="triggerId" class="base-time-picker__label">
      {{ label }}<span v-if="required" class="base-time-picker__sr-only"> (obligatoria)</span
      ><span v-if="required" aria-hidden="true"> *</span>
    </label>
    <button
      :id="triggerId"
      ref="trigger"
      type="button"
      class="base-time-picker__trigger"
      :disabled="disabled"
      :aria-labelledby="labelledBy"
      aria-haspopup="dialog"
      :aria-expanded="open"
      :aria-controls="dialogId"
      @click="open ? close() : show()"
      @keydown.down.prevent="show"
      @keydown.up.prevent="show"
    >
      <span :class="{ 'base-time-picker__placeholder': !modelValue }">{{
        modelValue || 'Elegir hora'
      }}</span>
      <svg
        viewBox="0 0 16 16"
        width="16"
        height="16"
        fill="none"
        stroke="currentColor"
        stroke-width="1.25"
        aria-hidden="true"
      >
        <circle cx="8" cy="8" r="6" />
        <path d="M8 4v4l3 2" />
      </svg>
    </button>
    <Teleport to="body">
      <Transition name="time-picker-pop">
        <div
          v-if="open"
          :id="dialogId"
          ref="panel"
          role="dialog"
          :aria-labelledby="titleId"
          :aria-describedby="hintId"
          class="base-time-picker__dialog"
          :style="style"
          @keydown="dialogKeydown"
        >
          <header class="base-time-picker__header">
            <div>
              <h3 :id="titleId">{{ label }}</h3>
              <p :id="hintId">Formato de 24 horas</p>
            </div>
            <button
              type="button"
              class="base-time-picker__close"
              :aria-label="`Cerrar selector de ${label.toLowerCase()}`"
              @click="close()"
            >
              ×
            </button>
          </header>
          <div class="base-time-picker__columns">
            <div class="base-time-picker__column">
              <label :id="`hour-label-${uid}`" :for="`hour-input-${uid}`">Hora</label>
              <input
                :id="`hour-input-${uid}`"
                :value="hourText"
                type="text"
                inputmode="numeric"
                maxlength="2"
                autocomplete="off"
                role="spinbutton"
                aria-valuemin="0"
                aria-valuemax="23"
                :aria-valuenow="validHour ? hour : undefined"
                :aria-invalid="!validHour"
                :aria-describedby="!validTime ? `time-error-${uid}` : undefined"
                @input="edit('hour', $event)"
                @keydown="partKeydown('hour', $event)"
                @focus="($event.target as HTMLInputElement).select()"
              />
            </div>
            <span class="base-time-picker__colon" aria-hidden="true">:</span>
            <div class="base-time-picker__column">
              <label :id="`minute-label-${uid}`" :for="`minute-input-${uid}`">Minutos</label>
              <input
                :id="`minute-input-${uid}`"
                :value="minuteText"
                type="text"
                inputmode="numeric"
                maxlength="2"
                autocomplete="off"
                role="spinbutton"
                aria-valuemin="0"
                aria-valuemax="59"
                :aria-valuenow="validMinute ? minute : undefined"
                :aria-invalid="!validMinute"
                :aria-describedby="!validTime ? `time-error-${uid}` : undefined"
                @input="edit('minute', $event)"
                @keydown="partKeydown('minute', $event)"
                @focus="($event.target as HTMLInputElement).select()"
              />
            </div>
          </div>
          <p
            v-if="!validTime"
            :id="`time-error-${uid}`"
            class="base-time-picker__error"
            role="alert"
          >
            Hora de 00 a 23 y minutos de 00 a 59.
          </p>
          <footer class="base-time-picker__footer">
            <button type="button" @click="close()">Cancelar</button>
            <button
              type="button"
              class="base-time-picker__apply"
              :disabled="!validTime"
              @click="apply"
            >
              Aplicar hora
            </button>
          </footer>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.base-time-picker__sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}
.base-time-picker {
  min-width: 0;
  display: grid;
  gap: 8px;
}
.base-time-picker__label {
  color: var(--color-on-strong-muted);
  font-size: 11px;
  font-weight: 600;
  line-height: 14px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
.base-time-picker__label span {
  color: var(--color-danger-on-strong);
}
.base-time-picker__trigger {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  width: 100%;
  min-height: 44px;
  padding: 0 14px;
  font: inherit;
  color: var(--color-on-strong);
  background: var(--color-field-strong);
  border: 1px solid var(--color-field-strong-border);
  border-bottom: 2px solid var(--color-brand-accent-surface);
  border-radius: 2px;
  cursor: pointer;
}
.base-time-picker__trigger svg {
  color: var(--color-brand-accent-surface);
  flex-shrink: 0;
}
.base-time-picker__placeholder {
  color: var(--color-on-strong-muted);
}
.base-time-picker__trigger:hover:not(:disabled),
.base-time-picker__trigger[aria-expanded='true'] {
  border-color: var(--color-brand-accent-surface);
}
.base-time-picker__trigger:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.base-time-picker__dialog {
  position: fixed;
  z-index: var(--layer-dialog);
  box-sizing: border-box;
  width: min(clamp(192px, var(--picker-trigger-width, 212px), 240px), calc(100vw - 16px));
  max-height: calc(100dvh - 16px);
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 10px;
  color: var(--color-on-strong);
  background: var(--color-field-strong);
  border: 1px solid var(--color-field-strong-border);
  border-top: 2px solid var(--color-brand-accent-surface);
  border-radius: 2px;
  box-shadow: var(--shadow-dialog);
}
.base-time-picker__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}
.base-time-picker__header h3 {
  margin: 0;
  font: 600 12px/1.4 var(--font-family-base);
}
.base-time-picker__header p {
  margin: 2px 0 0;
  color: var(--color-on-strong-muted);
  font-size: 12px;
}
.base-time-picker__dialog button {
  min-height: 44px;
  font-family: var(--font-family-base);
  font-size: 14px;
  color: var(--color-on-strong);
  background: transparent;
  border: 1px solid transparent;
  border-radius: 2px;
  cursor: pointer;
}
.base-time-picker__dialog button:hover {
  background: var(--color-field-strong-raised);
  border-color: var(--color-brand-accent-surface);
}
.base-time-picker__close {
  min-width: 44px;
  font-size: 24px !important;
}
.base-time-picker__columns {
  display: grid;
  grid-template-columns: 1fr 12px 1fr;
  gap: 4px;
  margin-top: 4px;
}
.base-time-picker__column {
  display: grid;
  gap: 4px;
  min-width: 0;
}
.base-time-picker__column label {
  text-align: center;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-on-strong-muted);
}
.base-time-picker__column input {
  box-sizing: border-box;
  width: 100%;
  height: 44px;
  padding: 0;
  text-align: center;
  font: 500 24px/1 var(--font-family-base);
  font-variant-numeric: tabular-nums;
  color: var(--color-on-strong);
  background: var(--color-field-strong-raised);
  border: 1px solid var(--color-field-strong-border);
  border-bottom: 2px solid var(--color-brand-accent-surface);
  border-radius: 2px;
}
.base-time-picker__colon {
  padding-top: 28px;
  color: var(--color-brand-accent-surface);
  font: 500 20px/1 var(--font-family-base);
  text-align: center;
}
.base-time-picker__column input::selection {
  background: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
}
.base-time-picker__error {
  margin: 12px 0 0;
  font-size: 12px;
  color: var(--color-danger-on-strong);
}
.base-time-picker__footer {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding-top: 8px;
  margin-top: 8px;
  border-top: 1px solid var(--color-field-strong-border);
}
.base-time-picker__footer button {
  padding: 0 8px;
  font-size: 12px;
}
.base-time-picker__footer .base-time-picker__apply {
  background: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  border-color: var(--color-brand-accent-surface);
}
.base-time-picker__footer button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.base-time-picker__trigger:focus-visible,
.base-time-picker__dialog button:focus-visible,
.base-time-picker__column input:focus-visible {
  outline: 2px solid var(--color-brand-accent-surface);
  outline-offset: 2px;
}
.time-picker-pop-enter-active,
.time-picker-pop-leave-active {
  transition: opacity 120ms ease;
}
.time-picker-pop-enter-from,
.time-picker-pop-leave-to {
  opacity: 0;
}
@media (prefers-reduced-motion: reduce) {
  .time-picker-pop-enter-active,
  .time-picker-pop-leave-active {
    transition: none;
  }
}
</style>
