<script setup lang="ts">
// Calificación y reseñas de Google del negocio en la entrada pública (HU-101,
// DEC-129). Se piden al abrir la página; sin ficha, sin credencial o ante cualquier
// fallo la sección no existe y nunca muestra un error (CA-101-01, CA-101-03). Cada
// reseña lleva la atribución a su autor y a Google, y el enlace a la ficha real
// (CA-101-02). Las reseñas se despliegan bajo demanda para no alargar la entrada.
import { onMounted, ref } from 'vue'
import { StarRating } from '@/shared/ui'
import { formatRating } from '@/shared/model/rating'
import { getPublicGoogleReviews } from '../api/getPublicGoogleReviewsApi'
import type { PublicGoogleReviews } from '../model/publicGoogleReviewsOutcome'

interface Props {
  slug: string
}
const props = defineProps<Props>()

const reviews = ref<PublicGoogleReviews | null>(null)
const open = ref(false)

onMounted(async () => {
  const outcome = await getPublicGoogleReviews(props.slug)
  if (outcome.kind === 'success') reviews.value = outcome.reviews
})

const countFormat = new Intl.NumberFormat('es-CO')
</script>

<template>
  <section v-if="reviews" class="pb-google" aria-labelledby="pb-google-title">
    <h2 id="pb-google-title" class="pb-google__title">Opiniones en Google</h2>
    <div class="pb-google__summary">
      <StarRating
        :rating="reviews.rating"
        :count="reviews.ratingCount"
        size="md"
        :show-value="false"
        :label="`${formatRating(reviews.rating)} de 5 estrellas en Google`"
      />
      <p class="pb-google__score">
        <strong>{{ formatRating(reviews.rating) }}</strong>
        <span> · {{ countFormat.format(reviews.ratingCount) }} calificaciones en Google</span>
      </p>
    </div>

    <template v-if="reviews.reviews.length > 0">
      <button
        type="button"
        class="pb-google__toggle"
        :aria-expanded="open"
        aria-controls="pb-google-list"
        @click="open = !open"
      >
        {{ open ? 'Ocultar reseñas' : `Ver reseñas (${reviews.reviews.length})` }}
      </button>
      <ul v-if="open" id="pb-google-list" class="pb-google__list">
        <li v-for="(review, index) in reviews.reviews" :key="index" class="pb-google__item">
          <div class="pb-google__head">
            <StarRating
              :rating="review.rating"
              :count="1"
              size="sm"
              :show-value="false"
              :label="`${review.rating} de 5 estrellas`"
            />
            <span class="pb-google__author">
              <a
                v-if="review.authorUrl"
                :href="review.authorUrl"
                target="_blank"
                rel="noopener noreferrer"
                >{{ review.authorName }}</a
              >
              <template v-else>{{ review.authorName }}</template>
            </span>
            <span v-if="review.relativeTime" class="pb-google__when">{{
              review.relativeTime
            }}</span>
          </div>
          <p v-if="review.text" class="pb-google__text">{{ review.text }}</p>
        </li>
      </ul>
    </template>

    <p class="pb-google__attribution">
      Reseñas de Google
      <template v-if="reviews.mapsUrl">
        ·
        <a :href="reviews.mapsUrl" target="_blank" rel="noopener noreferrer"
          >Ver la ficha en Google Maps (nueva pestaña)</a
        >
      </template>
    </p>
  </section>
</template>
