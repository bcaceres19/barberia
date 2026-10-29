<script setup lang="ts">
/**
 * AddressAutocomplete - Dirección con autocompletado de Google Places y mapa de
 * confirmación (DEC-131, HU-103). Un combobox accesible (patrón ARIA 1.2 «list
 * autocomplete»): el texto escrito NO es una dirección válida; solo una sugerencia
 * elegida, resuelta a coordenadas reales, emite `update:modelValue` con la selección.
 * Escribir de nuevo la descarta (`null`), así el formulario nunca envía un texto libre.
 *
 * Teclado: ↓/↑ recorren las sugerencias, Enter elige la resaltada, Esc cierra la lista.
 * Una respuesta tardía de una consulta ya reemplazada se descarta. La dirección es dato
 * personal (RN-DAT-02): no se registra ni se guarda fuera del formulario.
 */
import { computed, onBeforeUnmount, ref, useId, watch } from 'vue'
import {
  createPlacesClient,
  MIN_QUERY_LENGTH,
  type PlaceSelection,
  type PlaceSuggestion,
  type PlacesClient,
} from '../model/placesAutocomplete'
import LocationMap from './LocationMap.vue'

interface Props {
  modelValue: PlaceSelection | null
  label: string
  hint?: string
  error?: string
  disabled?: boolean
  /** Inyectable para pruebas; por defecto el cliente real o simulado del entorno. */
  client?: PlacesClient | null
  /** Espera tras la última tecla antes de consultar. */
  debounceMs?: number
}
const props = withDefaults(defineProps<Props>(), {
  hint: '',
  error: '',
  disabled: false,
  client: undefined,
  debounceMs: 250,
})
const emit = defineEmits<{ 'update:modelValue': [value: PlaceSelection | null] }>()

const uid = useId()
const inputId = `${uid}-input`
const listId = `${uid}-list`
const statusId = `${uid}-status`
const hintId = `${uid}-hint`
const errorId = `${uid}-error`

const places = computed(() => (props.client === undefined ? createPlacesClient() : props.client))

const text = ref(props.modelValue?.address ?? '')
const suggestions = ref<PlaceSuggestion[]>([])
const open = ref(false)
const activeIndex = ref(-1)
const loading = ref(false)
const failed = ref(false)
const resolving = ref(false)

let timer: ReturnType<typeof setTimeout> | undefined
let token = 0
let controller: AbortController | undefined

const optionId = (i: number) => `${uid}-opt-${i}`
const activeDescendant = computed(() =>
  open.value && activeIndex.value >= 0 ? optionId(activeIndex.value) : undefined,
)
const statusText = computed(() => {
  if (resolving.value) return 'Ubicando la dirección…'
  if (loading.value) return 'Buscando direcciones…'
  if (failed.value) return 'No pudimos buscar direcciones. Inténtalo de nuevo.'
  if (open.value && suggestions.value.length === 0) return 'No encontramos esa dirección.'
  if (open.value) return `${suggestions.value.length} sugerencias disponibles.`
  return ''
})
const describedBy = computed(
  () =>
    [props.hint ? hintId : '', props.error ? errorId : '', statusId].filter(Boolean).join(' ') ||
    undefined,
)

function cancelPending() {
  token++
  clearTimeout(timer)
  controller?.abort()
  controller = undefined
}

function onInput(event: Event) {
  text.value = (event.target as HTMLInputElement).value
  // Editar el texto invalida la selección anterior: un texto libre no es una dirección.
  if (props.modelValue) emit('update:modelValue', null)
  failed.value = false
  activeIndex.value = -1
  cancelPending()
  if ([...text.value.trim()].length < MIN_QUERY_LENGTH) {
    suggestions.value = []
    open.value = false
    loading.value = false
    return
  }
  loading.value = true
  open.value = true
  const mine = token
  timer = setTimeout(() => void search(mine), props.debounceMs)
}

async function search(mine: number) {
  const client = places.value
  if (!client) return
  controller = new AbortController()
  try {
    const found = await client.suggest(text.value, controller.signal)
    if (mine !== token) return
    suggestions.value = found
    failed.value = false
  } catch {
    if (mine !== token) return
    suggestions.value = []
    failed.value = true
  } finally {
    if (mine === token) loading.value = false
  }
}

async function choose(suggestion: PlaceSuggestion) {
  const client = places.value
  if (!client) return
  cancelPending()
  open.value = false
  suggestions.value = []
  resolving.value = true
  failed.value = false
  const mine = token
  try {
    const selection = await client.resolve(suggestion.placeId)
    if (mine !== token) return
    if (selection) {
      text.value = selection.address
      emit('update:modelValue', selection)
    } else {
      failed.value = true
    }
  } catch {
    if (mine === token) failed.value = true
  } finally {
    if (mine === token) resolving.value = false
  }
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
    if (suggestions.value.length === 0) return
    event.preventDefault()
    open.value = true
    const last = suggestions.value.length - 1
    activeIndex.value =
      event.key === 'ArrowDown'
        ? activeIndex.value >= last
          ? 0
          : activeIndex.value + 1
        : activeIndex.value <= 0
          ? last
          : activeIndex.value - 1
  } else if (event.key === 'Enter') {
    const picked = suggestions.value[activeIndex.value]
    if (open.value && picked) {
      event.preventDefault()
      void choose(picked)
    }
  } else if (event.key === 'Escape' && open.value) {
    event.preventDefault()
    event.stopPropagation()
    open.value = false
    activeIndex.value = -1
  }
}

watch(
  () => props.modelValue,
  (value) => {
    if (value === null) return
    text.value = value.address
  },
)
onBeforeUnmount(cancelPending)
</script>

<template>
  <div class="address-autocomplete">
    <label :for="inputId" class="address-autocomplete__label">{{ label }}</label>
    <input
      :id="inputId"
      class="address-autocomplete__input"
      type="text"
      role="combobox"
      autocomplete="off"
      aria-autocomplete="list"
      aria-haspopup="listbox"
      :aria-expanded="open && suggestions.length > 0"
      :aria-controls="listId"
      :aria-activedescendant="activeDescendant"
      :aria-describedby="describedBy"
      :aria-invalid="!!error"
      :disabled="disabled"
      :value="text"
      @input="onInput"
      @keydown="onKeydown"
    />
    <ul
      v-show="open && suggestions.length > 0"
      :id="listId"
      class="address-autocomplete__list"
      role="listbox"
      :aria-label="label"
    >
      <li
        v-for="(suggestion, i) in suggestions"
        :id="optionId(i)"
        :key="suggestion.placeId"
        class="address-autocomplete__option"
        :class="{ 'address-autocomplete__option--active': i === activeIndex }"
        role="option"
        :aria-selected="i === activeIndex"
        @mousedown.prevent
        @click="choose(suggestion)"
      >
        {{ suggestion.text }}
      </li>
    </ul>
    <p v-if="hint" :id="hintId" class="address-autocomplete__hint">{{ hint }}</p>
    <p v-if="error" :id="errorId" class="address-autocomplete__error" role="alert">{{ error }}</p>
    <p :id="statusId" class="address-autocomplete__status" role="status">{{ statusText }}</p>
    <LocationMap
      v-if="modelValue"
      class="address-autocomplete__map"
      :latitude="modelValue.latitude"
      :longitude="modelValue.longitude"
      :address="modelValue.address"
    />
  </div>
</template>

<style scoped>
.address-autocomplete {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 6px;
  color: inherit;
}

.address-autocomplete__label {
  font-size: var(--font-size-body-sm, 14px);
  font-weight: 600;
}

.address-autocomplete__input {
  box-sizing: border-box;
  width: 100%;
  min-height: 44px;
  padding: 8px 12px;
  font: inherit;
  color: inherit;
  background: transparent;
  border: 1px solid color-mix(in srgb, currentColor 40%, transparent);
  border-radius: 2px;
}

.address-autocomplete__input:focus-visible {
  outline: 2px solid currentColor;
  outline-offset: 2px;
}

.address-autocomplete__input[aria-invalid='true'] {
  border-color: var(--color-danger-text, currentColor);
}

.address-autocomplete__list {
  z-index: 5;
  padding: 4px 0;
  margin: 0;
  list-style: none;
  background: var(--color-surface, #fff);
  color: var(--color-text, #111);
  border: 1px solid color-mix(in srgb, currentColor 30%, transparent);
  border-radius: 2px;
  box-shadow: 0 6px 18px rgb(0 0 0 / 18%);
}

.address-autocomplete__option {
  padding: 10px 12px;
  cursor: pointer;
  overflow-wrap: anywhere;
}

.address-autocomplete__option--active,
.address-autocomplete__option:hover {
  background: color-mix(in srgb, currentColor 10%, transparent);
}

.address-autocomplete__hint,
.address-autocomplete__error,
.address-autocomplete__status {
  margin: 0;
  font-size: var(--font-size-caption, 13px);
}

.address-autocomplete__error {
  color: var(--color-danger-text, currentColor);
}

.address-autocomplete__status:empty {
  display: none;
}

.address-autocomplete__map {
  margin-top: 6px;
}
</style>
