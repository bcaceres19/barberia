<script setup lang="ts" generic="T extends string">
// Grupo de opciones excluyentes accesible (DEC-110): un `radiogroup` real con
// inputs nativos ocultos dentro de cada etiqueta, así el navegador resuelve las
// flechas, el foco y el anuncio de "n de m" sin código propio. El aspecto de
// cada opción lo pone el consumidor por el slot; aquí solo viven la semántica,
// el foco visible y la entrada escalonada.
import { useId } from 'vue'

const props = defineProps<{
  modelValue: T
  options: readonly { value: T; label: string; hint?: string }[]
  /** Nombre accesible del grupo. */
  label: string
  disabled?: boolean
}>()
const emit = defineEmits<{ 'update:modelValue': [value: T] }>()
const name = `option-group-${useId()}`
</script>

<template>
  <div class="option-group" role="radiogroup" :aria-label="label">
    <label
      v-for="(option, index) in options"
      :key="option.value"
      class="option-group__option"
      :class="{ 'option-group__option--selected': option.value === props.modelValue }"
      :style="{ '--option-index': index }"
    >
      <input
        class="option-group__input"
        type="radio"
        :name="name"
        :value="option.value"
        :checked="option.value === props.modelValue"
        :disabled="disabled"
        @change="emit('update:modelValue', option.value)"
      />
      <slot :option="option" :selected="option.value === props.modelValue" />
    </label>
  </div>
</template>

<style scoped>
.option-group {
  display: flex;
  flex-wrap: wrap;
  gap: var(--option-group-gap, 10px);
}

.option-group__option {
  position: relative;
  display: block;
  cursor: pointer;
  animation: option-enter 320ms var(--motion-easing-standard) both;
  animation-delay: calc(var(--option-index, 0) * 45ms);
}

/* Oculto a la vista pero operable: sigue siendo el control real. */
.option-group__input {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  margin: 0;
  opacity: 0;
  cursor: pointer;
  /* Por encima del contenido de la opción: un clic sobre un elemento con
     transform (el rombo del acento) llega igualmente al radio real. */
  z-index: 1;
}

.option-group__input:disabled {
  cursor: not-allowed;
}

.option-group__option:has(.option-group__input:disabled) {
  opacity: 0.5;
  cursor: not-allowed;
}

/* El foco visible lo dibuja el contenedor de la opción, no el input invisible. */
.option-group__option:has(.option-group__input:focus-visible) {
  outline: var(--border-width-emphasis) solid var(--color-focus);
  outline-offset: 3px;
  border-radius: 3px;
}

@keyframes option-enter {
  from {
    opacity: 0;
    transform: translateY(6px);
  }

  to {
    opacity: 1;
    transform: none;
  }
}
</style>
