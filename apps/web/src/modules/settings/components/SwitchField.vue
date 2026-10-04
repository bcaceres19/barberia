<script setup lang="ts">
// Interruptor accesible (`role="switch"`) con el lenguaje reglado de NAVA: una
// pista fina con un rombo que se desliza, no una píldora genérica. El estado
// se anuncia con `aria-checked` y también con texto ("Activado"/"Desactivado"),
// nunca solo con el color.
import { useId } from 'vue'

defineProps<{
  modelValue: boolean
  label: string
  hint?: string
  disabled?: boolean
}>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()
const id = useId()
</script>

<template>
  <div class="switch-field" :class="{ 'switch-field--on': modelValue }">
    <div class="switch-field__copy">
      <label :id="`${id}-label`" :for="`${id}-control`" class="switch-field__label">{{
        label
      }}</label>
      <p v-if="hint" :id="`${id}-hint`" class="switch-field__hint">{{ hint }}</p>
    </div>
    <button
      :id="`${id}-control`"
      type="button"
      role="switch"
      class="switch-field__control"
      :aria-checked="modelValue"
      :aria-labelledby="`${id}-label`"
      :aria-describedby="hint ? `${id}-hint` : undefined"
      :disabled="disabled"
      @click="emit('update:modelValue', !modelValue)"
    >
      <span class="switch-field__state" aria-hidden="true">{{
        modelValue ? 'Activado' : 'Desactivado'
      }}</span>
      <span class="switch-field__track" aria-hidden="true">
        <span class="switch-field__thumb" />
      </span>
    </button>
  </div>
</template>

<style scoped>
.switch-field {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.switch-field__copy {
  min-width: 0;
}

.switch-field__label {
  display: block;
  font-size: 14px;
  font-weight: 600;
  line-height: 20px;
  color: var(--color-on-strong);
}

.switch-field__hint {
  margin: 2px 0 0;
  font-size: 12px;
  line-height: 18px;
  color: var(--color-on-strong-muted);
}

.switch-field__control {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 10px;
  min-height: 44px;
  padding: 0 2px;
  background: transparent;
  border: 0;
  border-radius: 3px;
  color: var(--color-on-strong-muted);
  font-family: var(--font-sans);
  cursor: pointer;
}

.switch-field__control:focus-visible {
  outline: var(--border-width-emphasis) solid var(--color-focus);
  outline-offset: 2px;
}

.switch-field__state {
  min-width: 76px;
  text-align: right;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  transition: color var(--motion-duration-base) var(--motion-easing-standard);
}

.switch-field--on .switch-field__state {
  color: var(--color-brand-accent-surface);
}

.switch-field__track {
  position: relative;
  display: block;
  width: 46px;
  height: 24px;
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-on-strong) 32%, transparent);
  border-radius: 3px;
  background: color-mix(in srgb, var(--color-on-strong) 5%, transparent);
  transition:
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard);
}

.switch-field--on .switch-field__track {
  background: color-mix(in srgb, var(--color-brand-accent-surface) 22%, transparent);
  border-color: var(--color-brand-accent-surface);
}

/* Un rombo, el motivo de la casa, recorre la pista con un leve rebote. */
.switch-field__thumb {
  position: absolute;
  top: 50%;
  left: 5px;
  width: 12px;
  height: 12px;
  background: var(--color-on-strong-muted);
  transform: translateY(-50%) rotate(45deg);
  transition:
    left 280ms cubic-bezier(0.34, 1.56, 0.64, 1),
    background-color var(--motion-duration-base) var(--motion-easing-standard);
}

.switch-field--on .switch-field__thumb {
  left: 27px;
  background: var(--color-brand-accent-surface);
}

.switch-field__control:hover:not(:disabled) .switch-field__track {
  border-color: var(--color-brand-accent-surface);
}

.switch-field__control:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}
</style>
