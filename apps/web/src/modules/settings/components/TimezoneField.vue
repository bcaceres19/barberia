<script setup lang="ts">
// Selector de zona horaria con buscador (DEC-110, RN-DIS-07): un combobox
// accesible (patrón ARIA "list autocomplete") sobre el catálogo IANA del
// navegador, más un reloj en vivo de la zona elegida y la sugerencia de la zona
// de este dispositivo. La zona del dispositivo NUNCA se aplica sola: solo se
// ofrece como atajo, porque toda hora de la app sigue la zona de la barbería.
// El servidor confirma la zona contra su catálogo al guardar (CA-020-03).
import { computed, onBeforeUnmount, onMounted, ref, useId, watch } from 'vue'
import {
  cityOf,
  clockParts,
  detectDeviceTimezone,
  filterTimezones,
  listTimezones,
  offsetLabel,
} from '../model/timezones'

const props = defineProps<{
  modelValue: string
  label: string
  error?: string
  disabled?: boolean
}>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const uid = useId()
const inputId = `tz-input-${uid}`
const listId = `tz-list-${uid}`
const errorId = `tz-error-${uid}`
const optionId = (index: number) => `tz-option-${uid}-${index}`

const MAX_VISIBLE = 60

const zones = computed(() => listTimezones([props.modelValue]))
const deviceZone = detectDeviceTimezone()

const open = ref(false)
const typing = ref(false)
const query = ref('')
const active = ref(0)
const inputRef = ref<HTMLInputElement | null>(null)
const listRef = ref<HTMLUListElement | null>(null)

// Mientras la persona no teclea, el campo muestra la zona vigente.
const inputText = computed(() => (typing.value ? query.value : props.modelValue))

// Sin teclear, la zona vigente va primera (con su marca) y el resto sigue en
// orden alfabético: la lista se abre sobre lo que ya está elegido, no sobre
// "Abidjan". Al teclear, el orden es el del filtro.
const matches = computed(() => {
  if (typing.value) return filterTimezones(zones.value, query.value)
  return [props.modelValue, ...zones.value.filter((zone) => zone !== props.modelValue)]
})
const visible = computed(() => matches.value.slice(0, MAX_VISIBLE))

const now = ref(new Date())
let timer: ReturnType<typeof setInterval> | undefined
onMounted(() => {
  timer = setInterval(() => {
    now.value = new Date()
  }, 1000)
})
onBeforeUnmount(() => clearInterval(timer))

const clock = computed(() => clockParts(props.modelValue, now.value))
const offset = computed(() => offsetLabel(props.modelValue, now.value))

function openList() {
  if (props.disabled) return
  open.value = true
  const current = visible.value.indexOf(props.modelValue)
  active.value = current >= 0 ? current : 0
  scrollActiveIntoView()
}

function closeList() {
  open.value = false
  typing.value = false
  query.value = ''
}

function choose(zone: string) {
  emit('update:modelValue', zone)
  closeList()
  inputRef.value?.focus()
}

function onInput(event: Event) {
  typing.value = true
  query.value = (event.target as HTMLInputElement).value
  open.value = true
  active.value = 0
}

function scrollActiveIntoView() {
  void Promise.resolve().then(() => {
    document.getElementById(optionId(active.value))?.scrollIntoView?.({ block: 'nearest' })
  })
}

function move(delta: number) {
  if (!open.value) {
    openList()
    return
  }
  const count = visible.value.length
  if (!count) return
  active.value = (active.value + delta + count) % count
  scrollActiveIntoView()
}

function onKeydown(event: KeyboardEvent) {
  switch (event.key) {
    case 'ArrowDown':
      event.preventDefault()
      move(1)
      break
    case 'ArrowUp':
      event.preventDefault()
      move(-1)
      break
    case 'Home':
      if (open.value) {
        event.preventDefault()
        active.value = 0
        scrollActiveIntoView()
      }
      break
    case 'End':
      if (open.value) {
        event.preventDefault()
        active.value = Math.max(0, visible.value.length - 1)
        scrollActiveIntoView()
      }
      break
    case 'Enter':
      if (open.value && visible.value[active.value]) {
        event.preventDefault()
        choose(visible.value[active.value]!)
      }
      break
    case 'Escape':
      if (open.value) {
        // El primer Escape cierra la lista; no se propaga a un diálogo.
        event.preventDefault()
        event.stopPropagation()
        closeList()
      }
      break
    case 'Tab':
      closeList()
      break
  }
}

function onBlur(event: FocusEvent) {
  // Pasar el foco a una opción (clic en la lista) no debe cerrar antes de elegir.
  if (event.relatedTarget instanceof Node && listRef.value?.contains(event.relatedTarget)) return
  closeList()
}

watch(
  () => props.disabled,
  (disabled) => {
    if (disabled) closeList()
  },
)

const showDeviceHint = computed(
  () => !!deviceZone && deviceZone !== props.modelValue && !props.disabled,
)

defineExpose({ focus: () => inputRef.value?.focus() })
</script>

<template>
  <div class="tz-field" :class="{ 'tz-field--invalid': !!error, 'tz-field--open': open }">
    <label :for="inputId" class="tz-field__label">{{ label }}</label>

    <div class="tz-field__combo">
      <input
        :id="inputId"
        ref="inputRef"
        class="tz-field__input"
        type="text"
        role="combobox"
        autocomplete="off"
        autocapitalize="off"
        spellcheck="false"
        aria-autocomplete="list"
        aria-haspopup="listbox"
        :aria-expanded="open"
        :aria-controls="listId"
        :aria-activedescendant="open && visible.length ? optionId(active) : undefined"
        :aria-invalid="!!error"
        :aria-describedby="error ? errorId : undefined"
        :value="inputText"
        :disabled="disabled"
        placeholder="Busca por ciudad o país"
        @input="onInput"
        @focus="($event.target as HTMLInputElement).select()"
        @click="openList"
        @keydown="onKeydown"
        @blur="onBlur"
      />
      <span class="tz-field__chevron" aria-hidden="true" />

      <Transition name="tz-pop">
        <ul
          v-if="open"
          :id="listId"
          ref="listRef"
          class="tz-field__list"
          role="listbox"
          :aria-label="label"
        >
          <li
            v-for="(zone, index) in visible"
            :id="optionId(index)"
            :key="zone"
            class="tz-field__option"
            :class="{
              'tz-field__option--active': index === active,
              'tz-field__option--selected': zone === modelValue,
            }"
            role="option"
            :aria-selected="zone === modelValue"
            @mousedown.prevent="choose(zone)"
            @mousemove="active = index"
          >
            <span class="tz-field__city">{{ cityOf(zone) }}</span>
            <span class="tz-field__zone">{{ zone }}</span>
            <span class="tz-field__offset">{{ offsetLabel(zone, now) }}</span>
          </li>
          <li v-if="visible.length === 0" class="tz-field__empty" role="presentation">
            No encontramos esa zona. Prueba con el nombre de una ciudad cercana.
          </li>
          <li
            v-else-if="matches.length > visible.length"
            class="tz-field__more"
            role="presentation"
          >
            Mostrando {{ visible.length }} de {{ matches.length }}: sigue escribiendo para afinar.
          </li>
        </ul>
      </Transition>
    </div>

    <p v-if="error" :id="errorId" class="tz-field__error" role="alert">{{ error }}</p>

    <div v-if="clock" class="tz-field__clock" aria-live="off">
      <span class="tz-field__time" aria-hidden="true">
        <span class="tz-field__digits">{{ clock.hours }}</span
        ><span class="tz-field__colon">:</span
        ><span class="tz-field__digits">{{ clock.minutes }}</span>
      </span>
      <span class="tz-field__clock-copy">
        <span class="tz-field__clock-title"
          >Hora en {{ cityOf(modelValue) }}
          <span class="tz-field__clock-offset">· {{ offset }}</span></span
        >
        <span class="tz-field__clock-day">{{ clock.day }}</span>
      </span>
    </div>

    <button
      v-if="showDeviceHint"
      type="button"
      class="tz-field__device"
      @click="emit('update:modelValue', deviceZone!)"
    >
      Usar la zona de este dispositivo
      <span class="tz-field__device-zone">{{ deviceZone }}</span>
    </button>
  </div>
</template>

<style scoped>
.tz-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.tz-field__label {
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.08em;
  line-height: 14px;
  text-transform: uppercase;
}

.tz-field__combo {
  position: relative;
}

.tz-field__input {
  width: 100%;
  height: 44px;
  padding: 0 40px 0 14px;
  color: var(--color-on-strong);
  background: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-on-strong) 12%, transparent);
  border-bottom: var(--border-width-normal) solid var(--color-brand-accent-surface);
  border-radius: 3px;
  font-family: var(--font-sans);
  font-size: var(--font-size-body-sm);
  transition:
    border-color var(--motion-duration-base) var(--motion-easing-standard),
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    box-shadow var(--motion-duration-base) var(--motion-easing-standard);
}

.tz-field__input::placeholder {
  color: var(--color-on-strong-muted);
}

.tz-field__input:hover:not(:disabled) {
  border-color: color-mix(in srgb, var(--color-on-strong) 26%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.tz-field__input:focus-visible {
  outline: none;
  background: color-mix(in srgb, var(--color-on-strong) 6%, transparent);
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus);
}

.tz-field--invalid .tz-field__input {
  background: color-mix(in srgb, var(--color-danger-on-strong) 7%, transparent);
  border-bottom-color: var(--color-danger-on-strong);
}

.tz-field__input:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.tz-field__chevron {
  position: absolute;
  top: 50%;
  right: 16px;
  width: 7px;
  height: 7px;
  border-right: 1.5px solid var(--color-on-strong-muted);
  border-bottom: 1.5px solid var(--color-on-strong-muted);
  pointer-events: none;
  transform: translateY(-70%) rotate(45deg);
  transition: transform var(--motion-duration-base) var(--motion-easing-standard);
}

.tz-field--open .tz-field__chevron {
  transform: translateY(-30%) rotate(225deg);
}

.tz-field__list {
  position: absolute;
  z-index: var(--layer-menu);
  top: calc(100% + 6px);
  right: 0;
  left: 0;
  max-height: 280px;
  margin: 0;
  padding: 4px 0;
  overflow-y: auto;
  list-style: none;
  background: var(--color-field-strong-raised);
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent);
  border-radius: 3px;
  box-shadow: 0 18px 36px -12px rgb(16 27 43 / 55%);
  scrollbar-width: thin;
  scrollbar-color: color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent)
    transparent;
}

.tz-field__option {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  grid-template-areas:
    'city offset'
    'zone offset';
  column-gap: 12px;
  align-items: center;
  min-height: 44px;
  padding: 6px 14px;
  border-left: 2px solid transparent;
  color: var(--color-on-strong);
  cursor: pointer;
}

.tz-field__option--active {
  background: color-mix(in srgb, var(--color-brand-accent-surface) 14%, transparent);
  border-left-color: var(--color-brand-accent-surface);
}

/* Sobre el tinte de la opción activa el acento no alcanza AA como texto. */
.tz-field__option--active .tz-field__offset {
  color: var(--color-on-strong);
}

.tz-field__option--selected .tz-field__city::after {
  content: ' ✓';
  color: var(--color-brand-accent-surface);
}

.tz-field__city {
  grid-area: city;
  overflow: hidden;
  font-size: var(--font-size-body-sm);
  line-height: 18px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tz-field__zone {
  grid-area: zone;
  overflow: hidden;
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  line-height: 15px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tz-field__offset {
  grid-area: offset;
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.06em;
  white-space: nowrap;
}

.tz-field__empty,
.tz-field__more {
  padding: 10px 14px;
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  line-height: 18px;
}

.tz-field__error {
  margin: 0;
  color: var(--color-danger-on-strong);
  font-size: var(--font-size-caption);
  line-height: 16px;
}

/* Reloj de la zona elegida: dígitos serif en latón, dos puntos que laten. */
.tz-field__clock {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 14px;
  background: color-mix(in srgb, var(--color-brand-accent-surface) 8%, transparent);
  border-left: 3px solid var(--color-brand-accent-surface);
  border-radius: 2px;
}

.tz-field__time {
  display: inline-flex;
  align-items: baseline;
  font-family: var(--font-display);
  font-size: var(--font-size-title-page);
  line-height: 1;
  color: var(--color-on-strong);
  font-variant-numeric: tabular-nums;
}

.tz-field__colon {
  margin-inline: 2px;
  color: var(--color-brand-accent-surface);
  animation: tz-colon 2s steps(1, end) infinite;
}

.tz-field__clock-copy {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.tz-field__clock-title {
  font-size: var(--font-size-body-sm);
  font-weight: 600;
  line-height: 18px;
  color: var(--color-on-strong);
}

.tz-field__clock-offset {
  font-weight: 400;
  color: var(--color-on-strong-muted);
}

.tz-field__clock-day {
  font-size: var(--font-size-caption);
  line-height: 17px;
  color: var(--color-on-strong-muted);
}

.tz-field__device {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px 8px;
  align-self: flex-start;
  min-height: 44px;
  padding: 6px 2px;
  color: var(--color-brand-accent-surface);
  background: transparent;
  border: 0;
  border-radius: 3px;
  font-family: var(--font-sans);
  font-size: var(--font-size-caption);
  font-weight: 600;
  text-align: left;
  text-decoration: underline;
  text-underline-offset: 3px;
  cursor: pointer;
}

.tz-field__device:hover {
  text-decoration-thickness: 2px;
}

.tz-field__device:focus-visible {
  outline: var(--border-width-emphasis) solid var(--color-focus);
  outline-offset: 2px;
}

.tz-field__device-zone {
  color: var(--color-on-strong-muted);
  font-weight: 400;
  text-decoration: none;
}

.tz-pop-enter-active,
.tz-pop-leave-active {
  transition:
    opacity var(--motion-duration-fast) var(--motion-easing-standard),
    transform var(--motion-duration-base) cubic-bezier(0.34, 1.56, 0.64, 1);
  transform-origin: top center;
}

.tz-pop-enter-from,
.tz-pop-leave-to {
  opacity: 0;
  transform: translateY(-6px) scale(0.98);
}

@keyframes tz-colon {
  0%,
  100% {
    opacity: 1;
  }

  50% {
    opacity: 0.2;
  }
}

@media (prefers-reduced-motion: reduce) {
  .tz-field__colon {
    animation: none;
  }
}
</style>
