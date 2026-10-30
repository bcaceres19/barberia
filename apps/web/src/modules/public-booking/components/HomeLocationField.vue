<script setup lang="ts">
/**
 * HomeLocationField - Dónde atiende el barbero en un turno a domicilio (DEC-138, HU-103).
 * Tres formas equivalentes de indicarlo, para quien no se sabe su dirección: moverle el
 * mapa hasta que el pin quede sobre su puerta, el botón «Usar mi ubicación» o buscar una
 * dirección y elegir una sugerencia. Basta cualquiera; el valor es siempre un punto con su
 * texto (`PlaceSelection`), que es lo que el servidor guarda tal cual (`appointment_home_service_ck`).
 */
import { AddressAutocomplete, LocationPicker } from '@/shared/ui'
import type { LatLng } from '@/shared/model/mapPicker'
import type { PlaceSelection } from '@/shared/model/placesAutocomplete'

defineProps<{
  modelValue: PlaceSelection | null
  /** Dónde abrir el mapa si aún no hay ubicación (la de la barbería, si la tiene). */
  start?: LatLng | null
  error?: string
}>()
const emit = defineEmits<{ 'update:modelValue': [value: PlaceSelection | null] }>()
</script>

<template>
  <div class="home-location">
    <AddressAutocomplete
      :model-value="modelValue"
      label="Busca tu dirección (opcional)"
      hint="Si no te sabes la dirección, no hace falta: marca tu ubicación en el mapa."
      :show-map="false"
      :clear-on-input="false"
      @update:model-value="emit('update:modelValue', $event)"
    />
    <LocationPicker
      :model-value="modelValue"
      :start="start"
      label="Tu ubicación"
      :error="error"
      @update:model-value="emit('update:modelValue', $event)"
    />
  </div>
</template>

<style scoped>
.home-location {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
</style>
