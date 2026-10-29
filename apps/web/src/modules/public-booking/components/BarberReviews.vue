<script setup lang="ts">
// Reseñas visibles del barbero elegido (HU-100, DEC-128, CA-100-04). Se piden
// bajo demanda al abrir el desplegable: la selección de barbero ya trae el
// promedio, y la lista solo interesa a quien quiere leerla. Cambiar de barbero
// descarta la lista y la respuesta tardía de una carga reemplazada.
import { ref, watch } from 'vue'
import { StarRating } from '@/shared/ui'
import { listPublicBarberReviews } from '../api/listPublicBarberReviewsApi'
import type { PublicBarberReviews } from '../model/publicBarberReviewsOutcome'

interface Props {
  slug: string
  barberId: string
  barberName: string
  ratingCount: number
}
const props = defineProps<Props>()

type LoadState =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; reviews: PublicBarberReviews }
  | { status: 'error' }

const open = ref(false)
const state = ref<LoadState>({ status: 'idle' })
let requestToken = 0

async function load() {
  const token = ++requestToken
  state.value = { status: 'loading' }
  const outcome = await listPublicBarberReviews(props.slug, props.barberId)
  if (token !== requestToken) return
  state.value =
    outcome.kind === 'success'
      ? { status: 'success', reviews: outcome.reviews }
      : { status: 'error' }
}

function toggle() {
  open.value = !open.value
  if (open.value && state.value.status !== 'success') void load()
}

watch(
  () => props.barberId,
  () => {
    requestToken++
    open.value = false
    state.value = { status: 'idle' }
  },
)

const dateFormat = new Intl.DateTimeFormat('es-CO', { dateStyle: 'medium' })
function formatDate(iso: string): string {
  return dateFormat.format(new Date(iso))
}
</script>

<template>
  <section class="pb-reviews" :aria-label="`Reseñas de ${barberName}`">
    <button
      type="button"
      class="pb-reviews__toggle"
      :aria-expanded="open"
      :aria-controls="`pb-reviews-${barberId}`"
      @click="toggle"
    >
      {{ open ? 'Ocultar reseñas' : `Ver reseñas (${ratingCount})` }}
    </button>
    <div v-if="open" :id="`pb-reviews-${barberId}`" class="pb-reviews__panel">
      <p v-if="state.status === 'loading'" class="pb-reviews__note" role="status">
        Cargando reseñas…
      </p>
      <p v-else-if="state.status === 'error'" class="pb-reviews__note" role="alert">
        No pudimos cargar las reseñas.
        <button type="button" class="pb-reviews__retry" @click="load">Reintentar</button>
      </p>
      <template v-else-if="state.status === 'success'">
        <p v-if="state.reviews.items.length === 0" class="pb-reviews__note">
          Todavía no hay reseñas.
        </p>
        <ul v-else class="pb-reviews__list">
          <li v-for="(review, index) in state.reviews.items" :key="index" class="pb-reviews__item">
            <div class="pb-reviews__head">
              <StarRating
                :rating="review.rating"
                :count="1"
                size="sm"
                :show-value="false"
                :label="`${review.rating} de 5 estrellas`"
              />
              <time class="pb-reviews__date" :datetime="review.createdAt">{{
                formatDate(review.createdAt)
              }}</time>
            </div>
            <p v-if="review.comment" class="pb-reviews__comment">{{ review.comment }}</p>
          </li>
        </ul>
      </template>
    </div>
  </section>
</template>
