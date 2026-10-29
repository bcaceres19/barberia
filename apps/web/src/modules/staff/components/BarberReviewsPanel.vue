<script setup lang="ts">
// Reseñas de un barbero para quien administra (HU-100, DEC-128, CA-100-05). Muestra
// el promedio y, bajo demanda, TODAS las reseñas —las ocultas incluidas, con la fecha
// en que se ocultaron— y permite ocultar o volver a mostrar cada una. Ocultar no
// borra: es reversible y solo saca la reseña del promedio y de la lista pública.
// Autocontenido (pide sus datos por `barberId`) para usarlo igual en el detalle de
// un barbero y en la ficha del perfil individual.
import { computed, ref, watch } from 'vue'
import { BaseButton, StarRating } from '@/shared/ui'
import { listBarberReviews, setReviewHidden } from '../api/reviewsApi'
import type { BarberReview, BarberReviewList } from '../model/review'

interface Props {
  barberId: string
  /** Promedio y total que ya trae la lista de barberos, para pintar sin pedir nada. */
  rating: number | null
  ratingCount: number
}
const props = defineProps<Props>()
// El promedio cambia al ocultar o mostrar: se avisa para que quien lo muestra se actualice.
const emit = defineEmits<{ changed: [summary: { rating: number | null; ratingCount: number }] }>()

type LoadState =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; reviews: BarberReviewList }
  | { status: 'error' }

const open = ref(false)
const state = ref<LoadState>({ status: 'idle' })
const busyId = ref<string | null>(null)
const actionError = ref('')
let requestToken = 0

const shown = computed(() =>
  state.value.status === 'success'
    ? { rating: state.value.reviews.rating, count: state.value.reviews.ratingCount }
    : { rating: props.rating, count: props.ratingCount },
)

async function load() {
  const token = ++requestToken
  state.value = { status: 'loading' }
  actionError.value = ''
  const outcome = await listBarberReviews(props.barberId)
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

async function toggleHidden(review: BarberReview) {
  if (busyId.value) return
  busyId.value = review.id
  actionError.value = ''
  const outcome = await setReviewHidden(props.barberId, review.id, review.hiddenAt === null)
  busyId.value = null
  if (outcome.kind !== 'success') {
    actionError.value = 'No pudimos actualizar la reseña. Inténtalo de nuevo.'
    return
  }
  // El promedio lo calcula el servidor: se vuelve a leer la lista completa.
  await load()
  if (state.value.status === 'success') {
    emit('changed', {
      rating: state.value.reviews.rating,
      ratingCount: state.value.reviews.ratingCount,
    })
  }
}

watch(
  () => props.barberId,
  () => {
    requestToken++
    open.value = false
    state.value = { status: 'idle' }
    actionError.value = ''
  },
)

const dateFormat = new Intl.DateTimeFormat('es-CO', { dateStyle: 'medium' })
function formatDate(iso: string): string {
  return dateFormat.format(new Date(iso))
}
</script>

<template>
  <section class="barber-reviews" aria-label="Reseñas">
    <div class="barber-reviews__summary">
      <StarRating :rating="shown.rating" :count="shown.count" />
      <BaseButton
        type="button"
        variant="secondary"
        :aria-expanded="open"
        :aria-controls="`barber-reviews-${barberId}`"
        @click="toggle"
      >
        {{ open ? 'Ocultar reseñas' : 'Ver reseñas' }}
      </BaseButton>
    </div>

    <div v-if="open" :id="`barber-reviews-${barberId}`" class="barber-reviews__panel">
      <p v-if="state.status === 'loading'" class="barber-reviews__note" role="status">
        Cargando reseñas…
      </p>
      <p v-else-if="state.status === 'error'" class="barber-reviews__note" role="alert">
        No pudimos cargar las reseñas.
        <BaseButton type="button" variant="secondary" @click="load">Reintentar</BaseButton>
      </p>
      <template v-else-if="state.status === 'success'">
        <p v-if="actionError" class="barber-reviews__error" role="alert">{{ actionError }}</p>
        <p v-if="state.reviews.items.length === 0" class="barber-reviews__note">
          Todavía no hay reseñas.
        </p>
        <ul v-else class="barber-reviews__list">
          <li
            v-for="review in state.reviews.items"
            :key="review.id"
            class="barber-reviews__item"
            :class="{ 'barber-reviews__item--hidden': review.hiddenAt !== null }"
          >
            <div class="barber-reviews__head">
              <StarRating
                :rating="review.rating"
                :count="1"
                size="sm"
                :show-value="false"
                :label="`${review.rating} de 5 estrellas`"
              />
              <time :datetime="review.createdAt" class="barber-reviews__date">{{
                formatDate(review.createdAt)
              }}</time>
              <span v-if="review.hiddenAt" class="barber-reviews__badge">
                Oculta desde el {{ formatDate(review.hiddenAt) }}
              </span>
            </div>
            <p v-if="review.comment" class="barber-reviews__comment">{{ review.comment }}</p>
            <BaseButton
              type="button"
              variant="secondary"
              :disabled="busyId !== null"
              :aria-label="`${review.hiddenAt ? 'Mostrar' : 'Ocultar'} la reseña de ${review.rating} estrellas del ${formatDate(review.createdAt)}`"
              @click="toggleHidden(review)"
            >
              {{ review.hiddenAt ? 'Mostrar' : 'Ocultar' }}
            </BaseButton>
          </li>
        </ul>
      </template>
    </div>
  </section>
</template>

<style scoped>
.barber-reviews {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  margin-top: var(--space-4);
}

.barber-reviews__summary {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  align-items: center;
  justify-content: space-between;
}

.barber-reviews__panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.barber-reviews__note,
.barber-reviews__error {
  margin: 0;
  font-size: var(--font-size-body-sm);
}

.barber-reviews__error {
  color: var(--color-danger-text, currentColor);
}

.barber-reviews__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding: 0;
  margin: 0;
  list-style: none;
}

.barber-reviews__item {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  align-items: flex-start;
  padding-bottom: var(--space-3);
  border-bottom: 1px solid currentColor;
  border-bottom-color: color-mix(in srgb, currentColor 18%, transparent);
}

.barber-reviews__item--hidden .barber-reviews__comment {
  opacity: 0.7;
}

.barber-reviews__head {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1) var(--space-3);
  align-items: center;
}

.barber-reviews__date {
  font-size: var(--font-size-caption, 13px);
}

.barber-reviews__badge {
  padding: 2px var(--space-2);
  font-size: var(--font-size-caption, 13px);
  font-weight: 600;
  border: 1px solid currentColor;
  border-radius: 999px;
}

.barber-reviews__comment {
  margin: 0;
  overflow-wrap: anywhere;
}
</style>
