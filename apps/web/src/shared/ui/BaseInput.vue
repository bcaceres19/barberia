<script setup lang="ts">
/**
 * BaseInput - Campo de entrada base del sistema visual.
 * Tipos: text | email | password | number | tel | url | search
 * Estados: default | hover | focus | invalid | disabled | readonly
 * Con label, hint, error message, leading/trailing icons
 * Accesibilidad: aria-describedby para hint/error, aria-invalid, aria-readonly,
 * focus-visible ring, autocompletado nativo
 */
import { computed, ref, useId } from 'vue'

interface Props {
  /** Valor del input (v-model) */
  modelValue?: string | number
  /** Tipo de input HTML */
  type?: 'text' | 'email' | 'password' | 'number' | 'tel' | 'url' | 'search' | 'time'
  /** Label asociado */
  label?: string
  /** Texto de ayuda (hint) */
  hint?: string
  /** Mensaje de error (muestra estado invalid) */
  error?: string
  /** Placeholder */
  placeholder?: string
  /** Si está deshabilitado */
  disabled?: boolean
  /** Si es solo lectura */
  readonly?: boolean
  /** Si es requerido */
  required?: boolean
  /** Autocomplete */
  autocomplete?: string
  /** Nombre del campo */
  name?: string
  /** ID del input (se genera si no se proporciona) */
  id?: string
  /** Icono leading (slot name="leading") — decorativo únicamente; nunca un
   * control interactivo, porque el wrapper lleva aria-hidden. */
  /** Icono trailing (slot name="trailing") — misma restricción que leading. */
  /** Patrón de validación */
  pattern?: string
  /** Longitud mínima */
  minlength?: number
  /** Longitud máxima */
  maxlength?: number
  /** Valor mínimo (type=number) */
  min?: number | string
  /** Valor máximo (type=number) */
  max?: number | string
  /** Paso (type=number) */
  step?: number | string
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  disabled: false,
  readonly: false,
  required: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
  blur: [event: FocusEvent]
  focus: [event: FocusEvent]
  input: [event: Event]
  change: [event: Event]
}>()

const inputRef = ref<HTMLInputElement | null>(null)
// vue-tsc no asocia el ref="inputRef" del template con esta variable a
// efectos de noUnusedLocals: sin esta línea, "declared but never read".
void inputRef.value
// useId() (Vue 3.5+) genera un identificador estable por instancia, a
// diferencia de Math.random(): determinista dentro del ciclo de vida del
// componente y sin riesgo de colisión entre dos instancias montadas a la
// vez (docs/03-desarrollo/estandar-diseno-visual.md, auditoría HU-009:
// "IDs aleatorios o relaciones label/description/error inestables").
const generatedId = `base-input-${useId()}`
const inputId = computed(() => props.id || generatedId)
const hintId = computed(() => (props.hint ? `${inputId.value}-hint` : undefined))
const errorId = computed(() => (props.error ? `${inputId.value}-error` : undefined))
const describedBy = computed(() => {
  const ids = []
  if (hintId.value) ids.push(hintId.value)
  if (errorId.value) ids.push(errorId.value)
  return ids.length ? ids.join(' ') : undefined
})
const hasError = computed(() => !!props.error)

const classes = computed(() => {
  const base = 'base-input'
  return [
    base,
    hasError.value ? `${base}--invalid` : '',
    props.disabled ? `${base}--disabled` : '',
    props.readonly ? `${base}--readonly` : '',
  ]
    .filter(Boolean)
    .join(' ')
})

const wrapperClasses = computed(() => {
  const base = 'base-input__wrapper'
  return [
    base,
    props.disabled ? `${base}--disabled` : '',
    props.readonly ? `${base}--readonly` : '',
  ]
    .filter(Boolean)
    .join(' ')
})

const labelClasses = computed(() => {
  const base = 'base-input__label'
  return [base, props.required ? `${base}--required` : ''].filter(Boolean).join(' ')
})

const handleInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', target.value)
  emit('input', event)
}

const handleChange = (event: Event) => {
  emit('change', event)
}

const handleBlur = (event: FocusEvent) => {
  emit('blur', event)
}

const handleFocus = (event: FocusEvent) => {
  emit('focus', event)
}
</script>

<template>
  <div :class="wrapperClasses">
    <label v-if="label" :for="inputId" :class="labelClasses">
      {{ label }}
      <span v-if="required" class="base-input__required" aria-hidden="true">*</span>
    </label>

    <div class="base-input__input-wrapper">
      <span
        v-if="$slots.leading"
        class="base-input__icon base-input__icon--leading"
        aria-hidden="true"
      >
        <slot name="leading" />
      </span>

      <input
        ref="inputRef"
        :id="inputId"
        :type="type"
        :class="classes"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :readonly="readonly"
        :required="required"
        :autocomplete="autocomplete"
        :name="name"
        :aria-describedby="describedBy"
        :aria-invalid="hasError"
        :aria-readonly="readonly"
        :aria-disabled="disabled ? 'true' : undefined"
        :aria-required="required"
        :pattern="pattern"
        :minlength="minlength"
        :maxlength="maxlength"
        :min="min"
        :max="max"
        :step="step"
        @input="handleInput"
        @change="handleChange"
        @blur="handleBlur"
        @focus="handleFocus"
      />

      <span
        v-if="$slots.trailing"
        class="base-input__icon base-input__icon--trailing"
        aria-hidden="true"
      >
        <slot name="trailing" />
      </span>
    </div>

    <div v-if="hint && !error" :id="hintId" class="base-input__hint" aria-live="polite">
      {{ hint }}
    </div>

    <div v-if="error" :id="errorId" class="base-input__error" aria-live="assertive" role="alert">
      {{ error }}
    </div>
  </div>
</template>

<style scoped>
.base-input__wrapper {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  width: 100%;
}

.base-input__label {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  font-weight: 500;
  line-height: var(--font-size-body-sm-line);
  color: var(--color-text-primary);
}

.base-input__label--required::after {
  content: ' *';
  color: var(--color-danger-action);
  margin-left: 2px;
}

.base-input__input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.base-input {
  --input-height: var(--control-height);
  --input-padding-x: var(--space-4);
  --input-font-size: var(--font-size-body);
  --input-line-height: var(--font-size-body-line);
  --input-bg: var(--color-surface);
  --input-border: var(--border-width-normal) solid var(--color-border-control);
  --input-border-hover: var(--border-width-normal) solid var(--color-border-control);
  --input-border-focus: var(--border-width-emphasis) solid var(--color-focus);
  --input-border-invalid: var(--border-width-emphasis) solid var(--color-danger-border);
  --input-radius: var(--radius-md);
  --input-transition:
    border-color var(--motion-duration-fast) var(--motion-easing-standard),
    box-shadow var(--motion-duration-fast) var(--motion-easing-standard),
    background-color var(--motion-duration-fast) var(--motion-easing-standard);
  --input-focus-ring: 0 0 0 2px var(--color-surface), 0 0 0 4px var(--color-focus);

  width: 100%;
  height: var(--input-height);
  padding: 0 var(--input-padding-x);
  font-family: var(--font-family-base);
  font-size: var(--input-font-size);
  line-height: var(--input-line-height);
  color: var(--color-text-primary);
  background-color: var(--input-bg);
  border: var(--input-border);
  border-radius: var(--input-radius);
  outline: none;
  transition: var(--input-transition);
}

@media (prefers-reduced-motion: reduce) {
  .base-input {
    transition: none;
  }
}

.base-input::placeholder {
  color: var(--color-text-secondary);
  opacity: 0.64;
}

.base-input:hover:not(:disabled):not(.base-input--readonly):not(.base-input--invalid) {
  border-color: var(--color-text-secondary);
}

.base-input:focus-visible {
  border: var(--input-border-focus);
  box-shadow: var(--input-focus-ring);
}

.base-input:disabled,
.base-input--disabled {
  background-color: var(--color-surface-muted);
  border-color: var(--color-border-subtle);
  color: var(--color-text-secondary);
  cursor: not-allowed;
  opacity: 0.64;
}

.base-input:read-only,
.base-input--readonly {
  background-color: var(--color-surface-muted);
  border-color: var(--color-border-subtle);
  cursor: default;
}

.base-input--invalid {
  border: var(--input-border-invalid);
}

.base-input--invalid:focus-visible {
  box-shadow:
    0 0 0 2px var(--color-surface),
    0 0 0 4px var(--color-danger-border);
}

/* Iconos */
.base-input__icon {
  position: absolute;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
  pointer-events: none;
  flex-shrink: 0;
}

.base-input__icon--leading {
  left: var(--space-3);
}

.base-input__icon--trailing {
  right: var(--space-3);
}

.base-input:has(.base-input__icon--leading) {
  padding-left: calc(var(--input-padding-x) + 20px);
}

.base-input:has(.base-input__icon--trailing) {
  padding-right: calc(var(--input-padding-x) + 20px);
}

/* Hint y Error */
.base-input__hint,
.base-input__error {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
}

.base-input__hint {
  color: var(--color-text-secondary);
}

.base-input__error {
  color: var(--color-danger-text);
}
</style>
