<script setup lang="ts" generic="T extends string | number">
import { computed, nextTick, onMounted, onUnmounted, ref, useId, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: T
    options: readonly { value: T; label: string }[]
    label: string
    id?: string
    disabled?: boolean
  }>(),
  { disabled: false, id: undefined },
)
const emit = defineEmits<{ 'update:modelValue': [value: T] }>()
const generatedId = useId()
const triggerId = computed(() => props.id ?? `select-${generatedId}`)
const listId = `select-options-${generatedId}`
const labelId = `select-label-${generatedId}`
const root = ref<HTMLElement | null>(null)
const trigger = ref<HTMLButtonElement | null>(null)
const list = ref<HTMLUListElement | null>(null)
const open = ref(false)
const active = ref(0)
const selected = computed(() => props.options.find((option) => option.value === props.modelValue))

function focusOption(index: number) {
  active.value = index
  const option = list.value?.children[index]
  if (option instanceof HTMLElement) {
    option.focus()
    option.scrollIntoView?.({ block: 'nearest' })
  }
}
async function show() {
  if (props.disabled || !props.options.length) return
  open.value = true
  await nextTick()
  focusOption(
    Math.max(
      0,
      props.options.findIndex((option) => option.value === props.modelValue),
    ),
  )
}
function close(restoreFocus = true) {
  open.value = false
  if (restoreFocus) trigger.value?.focus()
}
function choose(index: number) {
  const option = props.options[index]
  if (!option || props.disabled) return
  emit('update:modelValue', option.value)
  close()
}
function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Tab') {
    close(false)
    return
  }
  if (event.key === 'Escape') {
    // El primer Escape cierra la lista; el siguiente puede cerrar su diálogo.
    event.preventDefault()
    event.stopPropagation()
    close()
    return
  }
  const count = props.options.length
  if (!count) return
  if (['ArrowDown', 'ArrowUp', 'Home', 'End', 'Enter', ' '].includes(event.key)) {
    event.preventDefault()
    if (event.key === 'Enter' || event.key === ' ') choose(active.value)
    else if (event.key === 'Home') focusOption(0)
    else if (event.key === 'End') focusOption(count - 1)
    else focusOption((active.value + (event.key === 'ArrowDown' ? 1 : -1) + count) % count)
  }
}
function onOutside(event: MouseEvent) {
  if (open.value && event.target instanceof Node && !root.value?.contains(event.target))
    close(false)
}
function onFocusout(event: FocusEvent) {
  if (event.relatedTarget instanceof Node && !root.value?.contains(event.relatedTarget))
    close(false)
}
watch(
  () => props.disabled,
  (disabled) => {
    if (disabled) close(false)
  },
)
onMounted(() => document.addEventListener('mousedown', onOutside))
onUnmounted(() => document.removeEventListener('mousedown', onOutside))
</script>

<template>
  <div ref="root" class="base-select" @focusout="onFocusout">
    <label :id="labelId" :for="triggerId" class="base-select__label">{{ label }}</label>
    <button
      :id="triggerId"
      ref="trigger"
      type="button"
      class="base-select__trigger"
      :disabled="disabled || !options.length"
      aria-haspopup="listbox"
      :aria-expanded="open"
      :aria-controls="listId"
      @click="open ? close() : show()"
      @keydown.down.prevent="show"
      @keydown.up.prevent="show"
    >
      <span>{{ selected?.label }}</span>
      <span class="base-select__chevron" aria-hidden="true" />
    </button>
    <Transition name="select-pop">
      <ul
        v-if="open"
        :id="listId"
        ref="list"
        class="base-select__list"
        role="listbox"
        :aria-labelledby="labelId"
        @keydown="onKeydown"
      >
        <li
          v-for="(option, index) in options"
          :key="option.value"
          role="option"
          class="base-select__option"
          :aria-selected="option.value === modelValue"
          :tabindex="index === active ? 0 : -1"
          @click="choose(index)"
        >
          <span>{{ option.label }}</span
          ><span class="base-select__mark" aria-hidden="true" />
        </li>
      </ul>
    </Transition>
  </div>
</template>

<style scoped>
.base-select {
  position: relative;
  min-width: 0;
  display: grid;
  gap: 8px;
}
.base-select__label {
  font-size: 11px;
  font-weight: 600;
  line-height: 14px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-accent-brass);
}
.base-select__trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
  min-height: 44px;
  padding: 10px 14px;
  font: inherit;
  text-align: left;
  color: var(--color-text-primary);
  background: var(--color-surface);
  border: 1px solid var(--color-border-control);
  border-bottom: 2px solid var(--color-accent-brass);
  border-radius: 2px;
  cursor: pointer;
}
.base-select__trigger:hover:not(:disabled),
.base-select__trigger[aria-expanded='true'] {
  border-color: var(--color-accent-brass);
}
.base-select__trigger:focus-visible {
  outline: 2px solid var(--color-focus);
  outline-offset: 3px;
}
.base-select__trigger:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.base-select__chevron {
  width: 7px;
  height: 7px;
  border-right: 1px solid var(--color-accent-brass);
  border-bottom: 1px solid var(--color-accent-brass);
  transform: rotate(45deg);
  transition: transform 160ms ease;
}
.base-select__trigger[aria-expanded='true'] .base-select__chevron {
  transform: rotate(225deg);
}
.base-select__list {
  position: absolute;
  top: calc(100% - 1px);
  left: 0;
  z-index: var(--layer-menu);
  width: 100%;
  max-height: 264px;
  overflow-y: auto;
  overscroll-behavior: contain;
  margin: 0;
  padding: 4px 0;
  list-style: none;
  background: var(--color-surface);
  border: 1px solid var(--color-border-control);
  border-top: 2px solid var(--color-accent-brass);
  border-radius: 2px;
  box-shadow: var(--shadow-dialog);
}
.base-select__option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  min-height: 44px;
  padding: 10px 14px;
  border-left: 3px solid transparent;
  color: var(--color-text-primary);
  font-size: 14px;
  cursor: pointer;
}
.base-select__option + .base-select__option {
  border-top: 1px solid var(--color-border-control);
}
.base-select__option:hover,
.base-select__option:focus-visible {
  outline: none;
  background: var(--color-field-strong-raised);
  color: var(--color-on-strong);
  border-left-color: var(--color-accent-brass);
}
.base-select__option[aria-selected='true'] {
  font-weight: 600;
  border-left-color: var(--color-accent-brass);
}
.base-select__mark {
  width: 7px;
  height: 7px;
  flex-shrink: 0;
  background: var(--color-accent-brass);
  transform: rotate(45deg);
  visibility: hidden;
}
.base-select__option[aria-selected='true'] .base-select__mark {
  visibility: visible;
}
.select-pop-enter-active,
.select-pop-leave-active {
  transition:
    opacity 160ms ease,
    transform 160ms ease;
}
.select-pop-enter-from,
.select-pop-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
@media (prefers-reduced-motion: reduce) {
  .select-pop-enter-active,
  .select-pop-leave-active,
  .base-select__chevron {
    transition: none;
  }
}
</style>
