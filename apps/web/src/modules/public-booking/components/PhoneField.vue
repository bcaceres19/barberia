<script setup lang="ts">
// Teléfono con el prefijo en su propia casilla. El selector nativo conserva
// teclado, lectores de pantalla y la lista del sistema; encima se dibuja la
// cara del selector (bandera + prefijo) y el `<select>` queda transparente.
// Pegar un número con prefijo (`+57 300…`) elige el país y deja solo el
// número nacional en la casilla.
import { computed, useId } from 'vue'
import {
  PHONE_COUNTRIES,
  flagEmoji,
  nationalDigits,
  phoneCountry,
  splitInternationalPhone,
} from '../model/phoneCountries'

interface Props {
  /** ISO del país elegido (`v-model:country`). */
  country: string
  /** Número nacional, sin prefijo (`v-model:national`). */
  national: string
  label?: string
  error?: string
  required?: boolean
}
const props = withDefaults(defineProps<Props>(), {
  label: 'Teléfono',
  error: undefined,
  required: false,
})

const emit = defineEmits<{
  'update:country': [iso: string]
  'update:national': [value: string]
}>()

const uid = useId()
const inputId = `phone-field-${uid}`
const errorId = `${inputId}-error`

const selected = computed(() => phoneCountry(props.country))

function onSelect(event: Event) {
  emit('update:country', (event.target as HTMLSelectElement).value)
}

function onInput(event: Event) {
  const target = event.target as HTMLInputElement
  const pasted = splitInternationalPhone(target.value)
  if (pasted) {
    emit('update:country', pasted.country)
    target.value = pasted.national
    emit('update:national', pasted.national)
    return
  }
  // Solo dígitos: separadores copiados de una agenda no deben quedar a la vista.
  const digits = nationalDigits(target.value)
  if (digits !== target.value) target.value = digits
  emit('update:national', digits)
}
</script>

<template>
  <div class="phone-field">
    <label :for="inputId" class="phone-field__label">
      {{ label }}
      <span v-if="required" class="phone-field__required" aria-hidden="true">*</span>
    </label>

    <div class="phone-field__row" :class="{ 'phone-field__row--invalid': !!error }">
      <div class="phone-field__prefix">
        <span class="phone-field__face" aria-hidden="true">
          <span class="phone-field__flag">{{ flagEmoji(selected.iso) }}</span>
          <span class="phone-field__dial">{{ selected.dial }}</span>
          <span class="phone-field__chevron"></span>
        </span>
        <select
          class="phone-field__select"
          aria-label="Prefijo del país"
          autocomplete="tel-country-code"
          :value="country"
          @change="onSelect"
        >
          <option v-for="item in PHONE_COUNTRIES" :key="item.iso" :value="item.iso">
            {{ flagEmoji(item.iso) }} {{ item.name }} ({{ item.dial }})
          </option>
        </select>
      </div>

      <input
        :id="inputId"
        class="phone-field__input"
        type="tel"
        inputmode="tel"
        autocomplete="tel-national"
        placeholder="300 123 4567"
        :value="national"
        :required="required"
        :aria-required="required"
        :aria-invalid="!!error"
        :aria-describedby="error ? errorId : undefined"
        @input="onInput"
      />
    </div>

    <div v-if="error" :id="errorId" class="phone-field__error" role="alert">{{ error }}</div>
  </div>
</template>

<style scoped>
/* Mismo campo reglado que `BaseInput` sobre la hoja de tinta: recuesto oscuro,
   filete perimetral y línea base de latón. Aquí el filete envuelve las dos
   casillas para que se lean como un solo teléfono, con un tabique entre ellas. */
.phone-field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  min-width: 0;
}

.phone-field__label {
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  line-height: 14px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--pb-brass);
}

.phone-field__required {
  margin-left: 2px;
  color: var(--color-danger-on-strong);
}

.phone-field__row {
  display: flex;
  height: var(--control-height);
  background-color: color-mix(in srgb, #000 46%, var(--color-surface-strong));
  border: var(--border-width-normal) solid var(--pb-line-strong);
  border-bottom: var(--border-width-emphasis) solid
    color-mix(in srgb, var(--pb-brass) 75%, transparent);
  border-radius: 2px;
  transition:
    border-color var(--motion-duration-fast) var(--motion-easing-standard),
    box-shadow var(--motion-duration-fast) var(--motion-easing-standard),
    background-color var(--motion-duration-fast) var(--motion-easing-standard);
}

.phone-field__row:hover {
  background-color: color-mix(in srgb, #000 38%, var(--color-surface-strong));
  border-color: color-mix(in srgb, var(--pb-brass) 60%, transparent);
  border-bottom-color: var(--pb-brass);
}

.phone-field__row:focus-within {
  background-color: color-mix(in srgb, #000 52%, var(--color-surface-strong));
  border-color: var(--pb-brass);
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--pb-brass),
    0 0 22px color-mix(in srgb, var(--pb-brass) 35%, transparent);
}

.phone-field__row--invalid,
.phone-field__row--invalid:hover {
  border-color: var(--color-danger-on-strong);
  border-bottom-color: var(--color-danger-on-strong);
}

.phone-field__row--invalid:focus-within {
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-danger-on-strong);
}

.phone-field__prefix {
  position: relative;
  display: flex;
  flex: none;
  align-items: center;
  border-right: 1px solid var(--pb-line-strong);
}

.phone-field__face {
  display: flex;
  gap: var(--space-2);
  align-items: center;
  padding: 0 var(--space-3);
  font-size: var(--font-size-body);
  color: var(--pb-text);
  pointer-events: none;
}

.phone-field__flag {
  font-size: 20px;
  line-height: 1;
}

.phone-field__dial {
  font-variant-numeric: tabular-nums;
  font-weight: 500;
}

.phone-field__chevron {
  width: 6px;
  height: 6px;
  margin: -3px 0 0 2px;
  border-right: 1.5px solid var(--pb-brass);
  border-bottom: 1.5px solid var(--pb-brass);
  transform: rotate(45deg);
}

/* El selector real cubre la cara: recibe el toque y el foco, pero no se ve. */
.phone-field__select {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  margin: 0;
  color: var(--pb-text);
  cursor: pointer;
  background: transparent;
  border: 0;
  opacity: 0;
}

.phone-field__select option {
  color: #000;
}

.phone-field__input {
  flex: 1;
  min-width: 0;
  padding: 0 var(--space-4);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--pb-text);
  background: transparent;
  border: 0;
  outline: none;
}

.phone-field__input::placeholder {
  color: var(--pb-muted);
  opacity: 1;
}

.phone-field__error {
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
  color: var(--color-danger-on-strong);
}

/* El foco del selector (opacity 0) se dibuja en la cara. */
.phone-field__prefix:has(.phone-field__select:focus-visible) .phone-field__face {
  outline: 2px solid var(--pb-brass);
  outline-offset: -4px;
}

@media (prefers-reduced-motion: reduce) {
  .phone-field__row {
    transition: none;
  }
}
</style>
