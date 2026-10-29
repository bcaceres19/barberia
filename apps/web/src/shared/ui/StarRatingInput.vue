<script setup lang="ts">
/**
 * StarRatingInput - Calificación de 1 a 5 estrellas (DEC-128, HU-100).
 * Un grupo de radios nativos (teclado y lector sin código extra: flechas mueven la
 * selección, Tab entra y sale del grupo) con un `<fieldset>` y su `<legend>`. Cada
 * estrella tiene el nombre accesible «N de 5 estrellas»; el foco visible lo pinta
 * la estrella correspondiente. Controlado con `v-model` (número o `null`).
 */
import { computed, useId } from 'vue'
import { MAX_RATING } from '../model/rating'

interface Props {
  modelValue: number | null
  /** Texto de la leyenda del grupo. */
  legend: string
  disabled?: boolean
  /** Mensaje de error ya redactado; marca el grupo como inválido. */
  error?: string
}

const props = withDefaults(defineProps<Props>(), { disabled: false, error: undefined })
const emit = defineEmits<{ 'update:modelValue': [value: number] }>()

const uid = useId()
const name = `rating-${uid}`
const errorId = `${name}-error`
const stars = Array.from({ length: MAX_RATING }, (_, i) => i + 1)
const describedBy = computed(() => (props.error ? errorId : undefined))

const STAR_PATH =
  'M12 2.5l2.94 5.96 6.58.96-4.76 4.64 1.12 6.55L12 17.52l-5.88 3.09 1.12-6.55L2.48 9.42l6.58-.96L12 2.5z'
</script>

<template>
  <fieldset
    class="star-input"
    :disabled="disabled"
    :aria-invalid="!!error"
    :aria-describedby="describedBy"
  >
    <legend class="star-input__legend">{{ legend }}</legend>
    <div class="star-input__stars">
      <label v-for="n in stars" :key="n" class="star-input__option">
        <input
          class="star-input__radio"
          type="radio"
          :name="name"
          :value="n"
          :checked="modelValue === n"
          :aria-label="`${n} de ${MAX_RATING} estrellas`"
          @change="emit('update:modelValue', n)"
        />
        <svg
          class="star-input__star"
          :class="{ 'star-input__star--on': modelValue !== null && n <= modelValue }"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <path :d="STAR_PATH" stroke-width="1.2" stroke-linejoin="round" />
        </svg>
      </label>
      <span class="star-input__value" aria-hidden="true">
        {{ modelValue === null ? 'Toca una estrella' : `${modelValue} de ${MAX_RATING}` }}
      </span>
    </div>
    <p v-if="error" :id="errorId" class="star-input__error" role="alert">{{ error }}</p>
  </fieldset>
</template>

<style scoped>
.star-input {
  min-width: 0;
  margin: 0;
  padding: 0;
  border: 0;
}

.star-input__legend {
  padding: 0;
  margin-bottom: var(--space-2, 8px);
  font-family: var(--font-family-base);
  font-size: var(--font-size-caption, 13px);
  font-weight: 600;
  color: inherit;
}

.star-input__stars {
  display: flex;
  align-items: center;
  gap: 4px;
}

.star-input__option {
  position: relative;
  display: inline-flex;
  padding: 4px;
  cursor: pointer;
  border-radius: 2px;
}

.star-input__radio {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  margin: 0;
  cursor: pointer;
  opacity: 0;
}

.star-input__option:has(.star-input__radio:focus-visible) {
  outline: var(--focus-ring-width, 2px) solid var(--color-focus);
  outline-offset: 2px;
}

.star-input__star {
  width: 36px;
  height: 36px;
  fill: transparent;
  stroke: var(--star-filled-color, var(--color-accent-brass));
  transition: fill 120ms ease;
}

.star-input__star--on {
  fill: var(--star-filled-color, var(--color-accent-brass));
}

.star-input__value {
  margin-left: var(--space-2, 8px);
  font-family: var(--font-family-base);
  font-size: var(--font-size-caption, 13px);
}

.star-input__error {
  margin: var(--space-2, 8px) 0 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-caption, 13px);
  color: var(--color-danger-text);
}

@media (prefers-reduced-motion: reduce) {
  .star-input__star {
    transition: none;
  }
}
</style>
