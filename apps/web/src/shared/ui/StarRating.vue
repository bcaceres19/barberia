<script setup lang="ts">
/**
 * StarRating - Promedio de estrellas de solo lectura (DEC-128, HU-100).
 * Cinco estrellas con relleno proporcional (4,5 pinta cuatro y media) y, opcional,
 * el valor y el total al lado. El conjunto es una sola imagen para el lector de
 * pantalla con el texto completo («4,5 de 5 · 12 reseñas»); los glifos y el texto
 * visible son decorativos para no leer dos veces. Sin reseñas visibles muestra
 * estrellas vacías y «Sin reseñas todavía», nunca «0».
 * Hereda el color del contexto: `--star-filled` y `--star-empty` lo ajustan en
 * superficies claras u oscuras.
 */
import { computed, useId } from 'vue'
import { MAX_RATING, ratingSummaryLabel, formatRating, reviewCountLabel } from '../model/rating'

interface Props {
  /** Promedio de 1 decimal, o `null` sin reseñas visibles. */
  rating: number | null
  /** Total de reseñas visibles. */
  count?: number
  /** Tamaño de las estrellas. */
  size?: 'sm' | 'md'
  /** Muestra el valor y el total junto a las estrellas. */
  showValue?: boolean
  /** Texto accesible propio; por defecto el resumen «4,5 de 5 · 12 reseñas». Una sola
   * reseña lo necesita («4 de 5 estrellas»): «1 reseña» la confundiría con un promedio. */
  label?: string
}

const props = withDefaults(defineProps<Props>(), { count: 0, size: 'md', showValue: true })

const uid = useId()
const hasRating = computed(() => props.rating !== null && props.count > 0)
const label = computed(() => props.label ?? ratingSummaryLabel(props.rating, props.count))

// Fracción de cada estrella pintada: la i-ésima (0..4) se llena con lo que falte
// del promedio, entre 0 y 1.
const fills = computed(() =>
  Array.from({ length: MAX_RATING }, (_, i) =>
    hasRating.value ? Math.min(1, Math.max(0, (props.rating ?? 0) - i)) : 0,
  ),
)

const STAR_PATH =
  'M12 2.5l2.94 5.96 6.58.96-4.76 4.64 1.12 6.55L12 17.52l-5.88 3.09 1.12-6.55L2.48 9.42l6.58-.96L12 2.5z'
</script>

<template>
  <span class="star-rating" :class="`star-rating--${size}`" role="img" :aria-label="label">
    <span class="star-rating__stars" aria-hidden="true">
      <svg v-for="(fill, i) in fills" :key="i" class="star-rating__star" viewBox="0 0 24 24">
        <defs>
          <linearGradient :id="`${uid}-${i}`" x1="0" x2="1" y1="0" y2="0">
            <stop :offset="`${fill * 100}%`" stop-color="var(--star-filled)" />
            <stop :offset="`${fill * 100}%`" stop-color="var(--star-empty)" />
          </linearGradient>
        </defs>
        <path
          :d="STAR_PATH"
          :fill="`url(#${uid}-${i})`"
          stroke="var(--star-filled)"
          stroke-width="1"
          stroke-linejoin="round"
        />
      </svg>
    </span>
    <span v-if="showValue" class="star-rating__value" aria-hidden="true">
      <template v-if="hasRating">
        <strong>{{ formatRating(rating) }}</strong>
        <span class="star-rating__count">({{ reviewCountLabel(count) }})</span>
      </template>
      <template v-else>Sin reseñas todavía</template>
    </span>
  </span>
</template>

<style scoped>
.star-rating {
  --star-filled: var(--star-filled-color, var(--color-accent-brass));
  --star-empty: var(--star-empty-color, transparent);

  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-family: var(--font-family-base);
  line-height: 1;
}

.star-rating__stars {
  display: inline-flex;
  gap: 2px;
}

.star-rating--md .star-rating__star {
  width: 18px;
  height: 18px;
}

.star-rating--sm .star-rating__star {
  width: 14px;
  height: 14px;
}

.star-rating__value {
  font-size: var(--font-size-caption, 13px);
  color: inherit;
}

.star-rating__value strong {
  font-weight: 700;
}

.star-rating__count {
  margin-left: 4px;
  opacity: 0.82;
}
</style>
