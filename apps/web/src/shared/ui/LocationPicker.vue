<script setup lang="ts">
/**
 * LocationPicker - Marca una ubicación moviendo el mapa o con la posición del dispositivo
 * (DEC-138, HU-103). El pin va fijo al centro: la persona mueve el MAPA hasta que el pin
 * queda sobre el lugar, así que no necesita saber su dirección. Emite `update:modelValue`
 * con el punto y un texto: la dirección aproximada del punto (geocodificación inversa, mejor
 * esfuerzo) o «Ubicación marcada en el mapa» si no hay una. Lo que el padre ya sabe (una
 * sugerencia elegida) recentra el mapa sin emitir de vuelta.
 *
 * Accesibilidad: el botón «Usar mi ubicación» y el autocompletado son la alternativa a
 * arrastrar el mapa; el mapa admite teclado (flechas) y la ubicación elegida se anuncia. La
 * detección solo se pide al pulsar el botón. Si el mapa no carga, la detección sigue
 * funcionando. La posición es dato personal (RN-DAT-02): no se registra ni se guarda fuera
 * del formulario.
 */
import { computed, onBeforeUnmount, onMounted, ref, useId, watch } from 'vue'
import {
  createMapPickerEngine,
  DEFAULT_START,
  DEFAULT_START_ZOOM,
  FALLBACK_ADDRESS,
  FOCUS_ZOOM,
  roundPoint,
  samePoint,
  type LatLng,
  type MapPickerEngine,
  type MapPickerHandle,
} from '../model/mapPicker'
import { isGeolocationSupported, locateUser, type LocateResult } from '../model/geolocation'
import type { PlaceSelection } from '../model/placesAutocomplete'
import BaseButton from './BaseButton.vue'

const MAX_ADDRESS_LENGTH = 200

interface Props {
  modelValue: PlaceSelection | null
  /** Dónde abrir el mapa si aún no hay ubicación (p. ej. la de la barbería). */
  start?: LatLng | null
  label?: string
  error?: string
  disabled?: boolean
  /** Inyectables para pruebas; por defecto el motor del entorno y la geolocalización real. */
  engine?: MapPickerEngine | null
  locate?: () => Promise<LocateResult>
  canLocate?: boolean
}
const props = withDefaults(defineProps<Props>(), {
  start: null,
  label: 'Tu ubicación',
  error: '',
  disabled: false,
  engine: undefined,
  locate: () => locateUser(),
  canLocate: () => isGeolocationSupported(),
})
const emit = defineEmits<{ 'update:modelValue': [value: PlaceSelection] }>()

const uid = useId()
const labelId = `${uid}-label`
const hintId = `${uid}-hint`
const errorId = `${uid}-error`

const engine = computed(() => (props.engine === undefined ? createMapPickerEngine() : props.engine))

const mapEl = ref<HTMLElement | null>(null)
const mapState = ref<'loading' | 'ready' | 'failed'>('loading')
const locating = ref(false)
const notice = ref('')

let handle: MapPickerHandle | null = null
let unmounted = false
let lastPoint: LatLng | null = props.modelValue ? pointOf(props.modelValue) : null
let describeToken = 0
let describeAbort: AbortController | undefined

function pointOf(selection: PlaceSelection): LatLng {
  return { latitude: selection.latitude, longitude: selection.longitude }
}

function trimAddress(text: string): string {
  return [...text].slice(0, MAX_ADDRESS_LENGTH).join('')
}

/** Marca el punto: emite ya con el texto de respaldo y lo mejora si hay dirección aproximada. */
function commit(raw: LatLng) {
  const point = roundPoint(raw)
  lastPoint = point
  emit('update:modelValue', { address: FALLBACK_ADDRESS, ...point })

  describeAbort?.abort()
  describeAbort = new AbortController()
  const mine = ++describeToken
  const client = engine.value
  if (!client) return
  client
    .describe(point, describeAbort.signal)
    .then((text) => {
      // Una descripción tardía de un punto que la persona ya movió se descarta.
      if (unmounted || mine !== describeToken || !text) return
      emit('update:modelValue', { address: trimAddress(text), ...point })
    })
    .catch(() => {
      // Mejor esfuerzo: se queda el texto de respaldo.
    })
}

onMounted(async () => {
  const client = engine.value
  const el = mapEl.value
  if (!client || !el) {
    mapState.value = 'failed'
    return
  }
  const known = props.modelValue ? pointOf(props.modelValue) : null
  const startPoint = known ?? props.start ?? DEFAULT_START
  const zoom = known ? FOCUS_ZOOM : props.start ? 15 : DEFAULT_START_ZOOM
  try {
    const mounted = await client.mount(el, startPoint, zoom, commit)
    if (unmounted) {
      mounted.destroy()
      return
    }
    handle = mounted
    mapState.value = 'ready'
  } catch {
    if (!unmounted) mapState.value = 'failed'
  }
})

// Una selección que viene de fuera (sugerencia elegida) recentra el mapa; la que emitimos
// nosotros misma no, o el mapa saltaría bajo el dedo de la persona.
watch(
  () => props.modelValue,
  (value) => {
    if (!value) return
    const point = pointOf(value)
    if (lastPoint && samePoint(point, lastPoint)) return
    lastPoint = point
    handle?.setCenter(point, FOCUS_ZOOM)
  },
)

const NOTICES: Record<Exclude<LocateResult['kind'], 'success'>, string> = {
  denied:
    'No tenemos permiso para ver tu ubicación. Puedes permitirlo en tu navegador o marcarla en el mapa.',
  unavailable: 'No pudimos detectar tu ubicación. Márcala moviendo el mapa.',
  timeout:
    'Tardó demasiado en detectar tu ubicación. Inténtalo de nuevo o márcala moviendo el mapa.',
  unsupported: 'Este navegador no puede detectar tu ubicación. Márcala moviendo el mapa.',
}

async function onLocate() {
  if (locating.value || props.disabled) return
  locating.value = true
  notice.value = ''
  const result = await props.locate()
  if (unmounted) return
  locating.value = false
  if (result.kind !== 'success') {
    notice.value = NOTICES[result.kind]
    return
  }
  lastPoint = result.point
  handle?.setCenter(result.point, FOCUS_ZOOM)
  commit(result.point)
  notice.value =
    result.accuracyMeters !== null && result.accuracyMeters > 100
      ? `Ubicamos tu posición con una precisión de unos ${result.accuracyMeters} m. Ajusta el pin si hace falta.`
      : 'Ubicamos tu posición. Ajusta el pin si hace falta.'
}

onBeforeUnmount(() => {
  unmounted = true
  describeAbort?.abort()
  handle?.destroy()
  handle = null
})

const describedBy = computed(() => [hintId, props.error ? errorId : ''].filter(Boolean).join(' '))
</script>

<template>
  <div class="location-picker" :aria-busy="mapState === 'loading'">
    <p :id="labelId" class="location-picker__label">{{ label }}</p>

    <BaseButton
      v-if="canLocate"
      type="button"
      variant="secondary"
      :loading="locating"
      :disabled="disabled || locating"
      @click="onLocate"
    >
      {{ locating ? 'Buscando tu ubicación…' : 'Usar mi ubicación' }}
    </BaseButton>

    <p :id="hintId" class="location-picker__hint">
      Mueve el mapa hasta que el pin quede sobre el lugar donde te atendemos. No necesitas saber la
      dirección.
    </p>

    <div class="location-picker__stage">
      <div
        ref="mapEl"
        class="location-picker__map"
        role="group"
        :aria-labelledby="labelId"
        :aria-describedby="describedBy"
      />
      <span v-if="mapState === 'ready'" class="location-picker__pin" aria-hidden="true" />
      <p v-if="mapState === 'loading'" class="location-picker__overlay" role="status">
        Cargando el mapa…
      </p>
    </div>

    <p v-if="mapState === 'failed'" class="location-picker__warning" role="status">
      No pudimos cargar el mapa. Busca tu dirección o usa tu ubicación.
    </p>
    <p class="location-picker__notice" role="status">{{ notice }}</p>
    <p v-if="modelValue" class="location-picker__chosen" role="status">
      Ubicación marcada: <strong>{{ modelValue.address }}</strong>
    </p>
    <p v-if="error" :id="errorId" class="location-picker__error" role="alert">{{ error }}</p>
  </div>
</template>

<style scoped>
.location-picker {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
  color: inherit;
}

.location-picker__label {
  margin: 0;
  font-size: var(--font-size-body-sm, 14px);
  font-weight: 600;
}

.location-picker__hint,
.location-picker__notice,
.location-picker__chosen,
.location-picker__warning,
.location-picker__error {
  margin: 0;
  font-size: var(--font-size-caption, 13px);
  overflow-wrap: anywhere;
}

.location-picker__notice:empty {
  display: none;
}

.location-picker__error {
  color: var(--color-danger-text, currentColor);
}

.location-picker__stage {
  position: relative;
  align-self: stretch;
}

.location-picker__map {
  position: relative;
  width: 100%;
  height: clamp(200px, 34vh, 280px);
  overflow: hidden;
  border: 1px solid color-mix(in srgb, currentColor 24%, transparent);
  border-radius: 2px;
  background: color-mix(in srgb, currentColor 8%, transparent);
}

/* El pin queda fijo al centro: la punta (el vértice inferior) marca el punto exacto. */
.location-picker__pin {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 22px;
  height: 22px;
  margin: -34px 0 0 -11px;
  background: #d93025;
  border: 2px solid #fff;
  border-radius: 50% 50% 50% 0;
  box-shadow: 0 2px 6px rgb(0 0 0 / 40%);
  transform: rotate(-45deg);
  pointer-events: none;
}

.location-picker__pin::after {
  position: absolute;
  top: 5px;
  left: 5px;
  width: 8px;
  height: 8px;
  content: '';
  background: #fff;
  border-radius: 50%;
}

.location-picker__overlay {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  margin: 0;
  font-size: var(--font-size-caption, 13px);
  pointer-events: none;
}

/* Modo simulado (solo desarrollo): botones que mueven el punto. */
.location-picker__map :deep(.map-picker-sim__surface) {
  display: grid;
  height: 100%;
  place-items: center;
  font-size: var(--font-size-caption, 13px);
  opacity: 0.7;
}

.location-picker__map :deep(.map-picker-sim__button) {
  position: absolute;
  right: 8px;
  width: 44px;
  height: 44px;
  font: inherit;
  color: inherit;
  cursor: pointer;
  background: var(--color-surface-strong, #1b2230);
  border: 1px solid color-mix(in srgb, currentColor 40%, transparent);
  border-radius: 2px;
}

.location-picker__map :deep(.map-picker-sim__button:nth-of-type(1)) {
  top: 8px;
}
.location-picker__map :deep(.map-picker-sim__button:nth-of-type(2)) {
  top: 60px;
}
.location-picker__map :deep(.map-picker-sim__button:nth-of-type(3)) {
  top: 112px;
}
.location-picker__map :deep(.map-picker-sim__button:nth-of-type(4)) {
  top: 164px;
}
</style>
