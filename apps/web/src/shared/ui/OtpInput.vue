<script setup lang="ts">
/**
 * OtpInput - Ranuras de código reglado (issue #212, auth-eventos/README.md
 * "Contrato visual común"). Seis posiciones que repiten la construcción del
 * campo reglado de BaseInput en formato estrecho: no es un componente de
 * kit de formularios genérico con caja redondeada, pertenece a la misma
 * familia visual (superficie de papel, línea base de tinta, radio 2px,
 * rótulo en versalitas de latón).
 *
 * Expone un único valor lógico de seis dígitos hacia afuera (`modelValue`
 * string), nunca seis campos independientes: quien consume este componente
 * no debe ensamblar el código a partir de refs individuales.
 */
import { computed, nextTick, ref, useId } from 'vue'

const LENGTH = 6

interface Props {
  /** Valor completo del código (hasta 6 dígitos), v-model */
  modelValue?: string
  /** Rótulo del grupo, asociado por aria-labelledby */
  label?: string
  /** Mensaje de error (activa el estado invalid) */
  error?: string
  /** Si está deshabilitado */
  disabled?: boolean
  /** Nombre del campo lógico (no se refleja en el DOM de cada ranura) */
  name?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  disabled: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  blur: [event: FocusEvent]
  focus: [event: FocusEvent]
  complete: [value: string]
}>()

const groupId = `otp-input-${useId()}`
const labelId = computed(() => (props.label ? `${groupId}-label` : undefined))
const errorId = computed(() => (props.error ? `${groupId}-error` : undefined))
const hasError = computed(() => !!props.error)

const digits = computed(() => {
  const value = props.modelValue.replace(/[^0-9]/g, '').slice(0, LENGTH)
  return Array.from({ length: LENGTH }, (_, index) => value[index] ?? '')
})

const slotRefs = ref<(HTMLInputElement | null)[]>([])
const setSlotRef = (el: unknown, index: number) => {
  slotRefs.value[index] = el as HTMLInputElement | null
}

const focusSlot = async (index: number) => {
  const clamped = Math.max(0, Math.min(LENGTH - 1, index))
  await nextTick()
  slotRefs.value[clamped]?.focus()
  slotRefs.value[clamped]?.select()
}

const emitValue = (nextDigits: string[]) => {
  const value = nextDigits.join('')
  emit('update:modelValue', value)
  if (value.length === LENGTH) emit('complete', value)
}

const onInput = (index: number, event: Event) => {
  const target = event.target as HTMLInputElement
  const cleaned = target.value.replace(/[^0-9]/g, '')

  if (!cleaned) {
    const next = [...digits.value]
    next[index] = ''
    emitValue(next)
    target.value = ''
    return
  }

  // Un input puede llegar con más de un carácter (autocompletado del
  // teclado, entrega por voz): se distribuye desde esta posición en vez
  // de descartar el resto, igual que un pegado.
  const next = [...digits.value]
  let cursor = index
  for (const char of cleaned) {
    if (cursor >= LENGTH) break
    next[cursor] = char
    cursor += 1
  }
  emitValue(next)
  target.value = next[index] ?? ''
  focusSlot(Math.min(cursor, LENGTH - 1))
}

const onKeydown = (index: number, event: KeyboardEvent) => {
  if (event.key === 'Backspace') {
    const current = digits.value[index]
    if (!current && index > 0) {
      event.preventDefault()
      const next = [...digits.value]
      next[index - 1] = ''
      emitValue(next)
      focusSlot(index - 1)
    }
    return
  }
  if (event.key === 'ArrowLeft') {
    event.preventDefault()
    focusSlot(index - 1)
    return
  }
  if (event.key === 'ArrowRight') {
    event.preventDefault()
    focusSlot(index + 1)
  }
}

const onPaste = (index: number, event: ClipboardEvent) => {
  const pasted = event.clipboardData?.getData('text') ?? ''
  const cleaned = pasted.replace(/[^0-9]/g, '')
  if (!cleaned) return
  event.preventDefault()

  const next = [...digits.value]
  let cursor = index
  for (const char of cleaned) {
    if (cursor >= LENGTH) break
    next[cursor] = char
    cursor += 1
  }
  emitValue(next)
  focusSlot(Math.min(cursor, LENGTH - 1))
}

const onFocus = (event: FocusEvent) => {
  ;(event.target as HTMLInputElement).select()
  emit('focus', event)
}

const onBlur = (event: FocusEvent) => {
  emit('blur', event)
}

const slotsClasses = (index: number) => {
  const base = 'otp-input__slot'
  return [
    base,
    hasError.value ? `${base}--invalid` : '',
    digits.value[index] ? `${base}--filled` : '',
  ]
    .filter(Boolean)
    .join(' ')
}
</script>

<template>
  <div class="otp-input" :class="{ 'otp-input--disabled': disabled }">
    <span v-if="label" :id="labelId" class="otp-input__label">{{ label }}</span>

    <div
      class="otp-input__slots"
      role="group"
      :aria-labelledby="labelId"
      :aria-describedby="errorId"
    >
      <input
        v-for="(digit, index) in digits"
        :key="index"
        :ref="(el) => setSlotRef(el, index)"
        type="text"
        inputmode="numeric"
        autocomplete="one-time-code"
        pattern="[0-9]*"
        maxlength="1"
        :class="slotsClasses(index)"
        :value="digit"
        :disabled="disabled"
        :aria-invalid="hasError"
        :aria-label="`Dígito ${index + 1} de ${LENGTH}`"
        @input="onInput(index, $event)"
        @keydown="onKeydown(index, $event)"
        @paste="onPaste(index, $event)"
        @focus="onFocus"
        @blur="onBlur"
      />
    </div>

    <div v-if="error" :id="errorId" class="otp-input__error" aria-live="assertive" role="alert">
      {{ error }}
    </div>
  </div>
</template>

<style scoped>
.otp-input {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  width: 100%;
}

.otp-input__label {
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  line-height: 14px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-accent-brass);
}

.otp-input__slots {
  display: flex;
  gap: var(--space-2);
}

/* Misma construcción del campo reglado (BaseInput) en formato estrecho:
   superficie blanca, filete perimetral de baja opacidad, línea base de
   tinta de 2px, radio 2px. El dígito se compone en la serif del wordmark,
   no en la sans funcional del resto del control. */
.otp-input__slot {
  flex: 1 1 0;
  min-width: 0;
  height: var(--control-height);
  padding: 0;
  font-family: var(--font-display);
  font-size: 22px;
  line-height: 1;
  text-align: center;
  color: var(--color-text-primary);
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-bottom: var(--border-width-emphasis) solid var(--color-action-primary);
  border-radius: 2px;
  outline: none;
  transition:
    border-color var(--motion-duration-fast) var(--motion-easing-standard),
    box-shadow var(--motion-duration-fast) var(--motion-easing-standard);
}

@media (prefers-reduced-motion: reduce) {
  .otp-input__slot {
    transition: none;
  }
}

.otp-input__slot:hover:not(:disabled):not(.otp-input__slot--invalid) {
  border-color: var(--color-text-secondary);
  border-bottom-color: var(--color-action-primary);
}

.otp-input__slot:focus-visible {
  border-color: var(--color-focus);
  box-shadow:
    0 0 0 2px var(--color-surface),
    0 0 0 4px var(--color-focus);
}

.otp-input__slot--filled {
  border-bottom-color: var(--color-action-primary);
}

.otp-input__slot--invalid {
  border-color: var(--color-danger-border);
  border-bottom-color: var(--color-danger-border);
}

.otp-input__slot--invalid:focus-visible {
  box-shadow:
    0 0 0 2px var(--color-surface),
    0 0 0 4px var(--color-danger-border);
}

.otp-input__slot:disabled {
  background-color: var(--color-surface-muted);
  border-color: var(--color-border-subtle);
  border-bottom-color: var(--color-border-subtle);
  color: var(--color-text-secondary);
  cursor: not-allowed;
  opacity: 0.64;
}

.otp-input__error {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
  color: var(--color-danger-text);
}
</style>
