<script setup lang="ts">
// Mapa de Google con marcador y la dirección en texto (DEC-130, HU-102). Lo usan
// la reserva pública y la vista previa de Configuración, así que vive en
// `shared/ui` y no depende de la paleta de ninguno: hereda el color del texto.
//
// El `iframe` lleva `title` (lector de pantalla) y no atrapa el teclado más allá
// de sus propios controles; para quien no pueda usar el mapa, la dirección va en
// texto y hay un enlace que lo abre en Google Maps. Sin forma de dibujar el mapa
// en este entorno (producción sin clave de navegador) no renderiza nada.
import { computed } from 'vue'
import { mapEmbedUrl, mapsLinkUrl } from '../model/mapEmbed'

const props = defineProps<{
  latitude: number
  longitude: number
  address: string
}>()

const embedUrl = computed(() => mapEmbedUrl(props.latitude, props.longitude))
const linkUrl = computed(() => mapsLinkUrl(props.latitude, props.longitude))
</script>

<template>
  <figure v-if="embedUrl" class="location-map">
    <iframe
      class="location-map__frame"
      :src="embedUrl"
      :title="`Mapa de ${address}`"
      loading="lazy"
      referrerpolicy="no-referrer-when-downgrade"
      allowfullscreen
    />
    <figcaption class="location-map__caption">
      <span class="location-map__address">{{ address }}</span>
      <a v-if="linkUrl" class="location-map__link" :href="linkUrl" target="_blank" rel="noopener">
        Abrir en Google Maps<span class="location-map__sr"> (se abre en una pestaña nueva)</span>
      </a>
    </figcaption>
  </figure>
</template>

<style scoped>
.location-map {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin: 0;
  color: inherit;
}

.location-map__frame {
  display: block;
  width: 100%;
  height: clamp(180px, 34vh, 260px);
  border: 1px solid color-mix(in srgb, currentColor 24%, transparent);
  border-radius: 2px;
  background: color-mix(in srgb, currentColor 8%, transparent);
}

.location-map__caption {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 16px;
  align-items: baseline;
  justify-content: space-between;
  font-size: 14px;
  line-height: 20px;
}

.location-map__address {
  min-width: 0;
  overflow-wrap: anywhere;
}

.location-map__link {
  color: inherit;
  text-decoration: underline;
  text-underline-offset: 3px;
  white-space: nowrap;
}

.location-map__link:focus-visible {
  outline: 2px solid currentColor;
  outline-offset: 3px;
}

.location-map__sr {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}
</style>
