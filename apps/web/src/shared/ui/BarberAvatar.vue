<script setup lang="ts">
/**
 * BarberAvatar - Retrato cuadrado de un barbero (issue #189, atlas
 * panel-agenda-eventos evento 12): radio 2px, filete de latón, monograma
 * derivado de `fullName` en la serif del wordmark sobre tinta cuando no hay
 * fotografía. Siempre decorativo (aria-hidden): el nombre adyacente visible
 * es la identidad accesible, nunca falta junto a un retrato.
 *
 * Con `photoUrl` muestra la fotografía real del barbero (DEC-104, evento 12
 * del atlas); si la imagen no carga (borrada en otro dispositivo, red caída)
 * vuelve sola al monograma, nunca deja un icono de imagen rota.
 *
 * Algoritmo de iniciales idéntico a `StaffPage.vue` (issue #189 lo duplica
 * deliberadamente en vez de extraerlo: StaffPage queda fuera de alcance de
 * esta entrega y migra a este componente en un issue propio).
 */
import { computed, ref, watch } from 'vue'

interface Props {
  fullName: string
  /** closed: 28px (26px móvil) — selector cerrado. option: 34px (32px
   * móvil) — cada opción de la lista desplegada. row: 52px (44px móvil) —
   * fila de la lista de equipo. hero: 120px (104px móvil) — detalle y
   * edición del barbero. */
  size?: 'closed' | 'option' | 'row' | 'hero'
  photoUrl?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  size: 'closed',
  photoUrl: null,
})

function initials(fullName: string): string {
  const parts = fullName.trim().split(/\s+/).filter(Boolean)
  if (parts.length === 0) return ''
  if (parts.length === 1) return parts[0]!.charAt(0).toUpperCase()
  return (parts[0]!.charAt(0) + parts[parts.length - 1]!.charAt(0)).toUpperCase()
}

const monogram = computed(() => initials(props.fullName))

// Una fotografía que falla al cargar cae al monograma; una URL nueva (otra
// versión de la imagen) vuelve a intentarlo.
const photoFailed = ref(false)
watch(
  () => props.photoUrl,
  () => {
    photoFailed.value = false
  },
)
const showPhoto = computed(() => !!props.photoUrl && !photoFailed.value)

const classes = computed(() => ['barber-avatar', `barber-avatar--${props.size}`])
</script>

<template>
  <span :class="classes" aria-hidden="true">
    <img
      v-if="showPhoto"
      class="barber-avatar__photo"
      :src="photoUrl ?? undefined"
      alt=""
      decoding="async"
      @error="photoFailed = true"
    />
    <span v-else class="barber-avatar__monogram">{{ monogram }}</span>
  </span>
</template>

<style scoped>
.barber-avatar {
  --avatar-size: 28px;

  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: var(--avatar-size);
  height: var(--avatar-size);
  overflow: hidden;
  /* Tinta ligeramente más clara que el fondo de página (atlas: --ink-2,
     #16243a) para que el retrato se distinga incluso antes de leer el
     monograma o el filete. */
  background-color: #16243a;
  border: var(--border-width-normal) solid rgb(184 149 90 / 55%);
  border-radius: 2px;
}

.barber-avatar--option {
  --avatar-size: 34px;
}

.barber-avatar--row {
  --avatar-size: 52px;
}

.barber-avatar--hero {
  --avatar-size: 120px;
}

@media (max-width: 767px) {
  .barber-avatar--closed {
    --avatar-size: 26px;
  }

  .barber-avatar--option {
    --avatar-size: 32px;
  }

  .barber-avatar--row {
    --avatar-size: 44px;
  }

  .barber-avatar--hero {
    --avatar-size: 104px;
  }
}

.barber-avatar__photo {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.barber-avatar__monogram {
  font-family: var(--font-display);
  font-size: calc(var(--avatar-size) * 0.38);
  line-height: 1;
  letter-spacing: -0.04em;
  white-space: nowrap;
  color: var(--color-brand-accent-surface);
}
</style>
