<script setup lang="ts">
/**
 * WeekdayPicker - Selector del día de la semana como siete fichas (L M X J V
 * S D). Patrón ARIA "radiogroup": un solo tope de tabulación (la ficha
 * elegida) y flechas para moverse; Inicio/Fin saltan a lunes/domingo. El
 * nombre accesible de cada ficha es el día completo, no la inicial.
 */
import { ref, useId } from 'vue'
import { ISO_WEEKDAYS } from '../model/workingHour'

interface Props {
  modelValue: number
  label: string
  disabled?: boolean
  error?: string
}

const props = withDefaults(defineProps<Props>(), { disabled: false, error: undefined })
const emit = defineEmits<{ 'update:modelValue': [value: number] }>()

const uid = useId()
const labelId = `weekday-label-${uid}`
const errorId = `weekday-error-${uid}`
const buttons = ref<HTMLButtonElement[]>([])

// X para miércoles: M ya es martes y la inicial repetida confundiría.
const SHORT: Record<number, string> = { 1: 'L', 2: 'M', 3: 'X', 4: 'J', 5: 'V', 6: 'S', 7: 'D' }

function choose(value: number) {
  if (props.disabled) return
  emit('update:modelValue', value)
}

function onKeydown(event: KeyboardEvent, index: number) {
  const last = ISO_WEEKDAYS.length - 1
  let next = index
  if (event.key === 'ArrowRight' || event.key === 'ArrowDown') next = index === last ? 0 : index + 1
  else if (event.key === 'ArrowLeft' || event.key === 'ArrowUp')
    next = index === 0 ? last : index - 1
  else if (event.key === 'Home') next = 0
  else if (event.key === 'End') next = last
  else return
  event.preventDefault()
  choose(ISO_WEEKDAYS[next]!.value)
  buttons.value[next]?.focus()
}
</script>

<template>
  <div class="weekday-picker">
    <span :id="labelId" class="weekday-picker__label">{{ label }}</span>
    <div
      class="weekday-picker__group"
      role="radiogroup"
      :aria-labelledby="labelId"
      :aria-describedby="error ? errorId : undefined"
    >
      <button
        v-for="(day, index) in ISO_WEEKDAYS"
        :key="day.value"
        :ref="(el) => (buttons[index] = el as HTMLButtonElement)"
        type="button"
        role="radio"
        class="weekday-picker__chip"
        :class="{ 'weekday-picker__chip--on': day.value === modelValue }"
        :aria-checked="day.value === modelValue"
        :aria-label="day.label"
        :title="day.label"
        :tabindex="day.value === modelValue ? 0 : -1"
        :disabled="disabled"
        @click="choose(day.value)"
        @keydown="onKeydown($event, index)"
      >
        {{ SHORT[day.value] }}
      </button>
    </div>
    <p v-if="error" :id="errorId" class="weekday-picker__error" role="alert">{{ error }}</p>
  </div>
</template>

<style scoped>
.weekday-picker {
  display: grid;
  gap: 8px;
  min-width: 0;
}

.weekday-picker__label {
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  font-weight: 600;
  line-height: 14px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.weekday-picker__group {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: 6px;
}

.weekday-picker__chip {
  position: relative;
  min-height: 44px;
  padding: 0;
  overflow: hidden;
  color: var(--color-on-strong);
  font-family: var(--font-display);
  font-size: var(--font-size-title-item);
  line-height: 1;
  background-color: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-on-strong) 14%, transparent);
  border-bottom: var(--border-width-emphasis) solid
    color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent);
  border-radius: 2px;
  cursor: pointer;
  transition:
    background-color var(--motion-duration-fast) var(--motion-easing-standard),
    color var(--motion-duration-fast) var(--motion-easing-standard),
    border-color var(--motion-duration-fast) var(--motion-easing-standard),
    transform var(--motion-duration-base) cubic-bezier(0.34, 1.56, 0.64, 1);
}

@media (hover: hover) {
  .weekday-picker__chip:hover:not(:disabled):not(.weekday-picker__chip--on) {
    background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
    border-color: color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
    transform: translateY(-2px);
  }
}

.weekday-picker__chip--on {
  color: var(--color-brand-accent-text);
  background-color: var(--color-brand-accent-surface);
  border-color: var(--color-brand-accent-surface);
  transform: translateY(-2px);
  animation: weekday-chip-pop var(--motion-duration-base) cubic-bezier(0.34, 1.56, 0.64, 1) both;
}

.weekday-picker__chip:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus);
}

.weekday-picker__chip:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.weekday-picker__error {
  margin: 0;
  color: var(--color-danger-on-strong);
  font-size: var(--font-size-body-sm);
}

@keyframes weekday-chip-pop {
  from {
    transform: translateY(0) scale(0.9);
  }

  to {
    transform: translateY(-2px) scale(1);
  }
}

@media (prefers-reduced-motion: reduce) {
  .weekday-picker__chip,
  .weekday-picker__chip--on {
    animation: none;
    transition: none;
  }
}
</style>
