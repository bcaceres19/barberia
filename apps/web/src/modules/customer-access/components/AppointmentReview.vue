<script setup lang="ts">
// Sección «Tu opinión» del acceso al turno (HU-100, DEC-128). Presentacional: la
// página posee el estado y llama al API. Tres casos excluyentes: el turno admite
// reseña (formulario), el cliente ya la dejó (se muestra de solo lectura) o ninguno
// (no renderiza nada: un turno no completado nunca ofrece «Calificar», CA-100-01).
import { computed, ref } from 'vue'
import { BaseAlert, BaseButton, StarRating, StarRatingInput } from '@/shared/ui'
import { commentLength, REVIEW_COMMENT_MAX_LENGTH } from '@/shared/model/rating'
import type { CustomerReview, ReviewInput } from '../model/customerAppointmentOutcome'

interface Props {
  canReview: boolean
  review: CustomerReview | null
  /** Nombre de quien atendió, para el título «Califica a …». */
  professionalName: string
  submitting?: boolean
  /** Mensaje de error del envío ya redactado por la página. */
  errorMessage?: string
}
const props = withDefaults(defineProps<Props>(), { submitting: false, errorMessage: '' })
const emit = defineEmits<{ submit: [input: ReviewInput] }>()

const rating = ref<number | null>(null)
const comment = ref('')
const ratingError = ref('')

const length = computed(() => commentLength(comment.value))
const overLimit = computed(() => length.value > REVIEW_COMMENT_MAX_LENGTH)
const commentError = computed(() =>
  overLimit.value ? `El comentario no puede superar ${REVIEW_COMMENT_MAX_LENGTH} caracteres.` : '',
)

function onSubmit() {
  ratingError.value = rating.value === null ? 'Elige de 1 a 5 estrellas.' : ''
  if (rating.value === null || overLimit.value) return
  const trimmed = comment.value.trim()
  emit('submit', { rating: rating.value, comment: trimmed === '' ? null : trimmed })
}
</script>

<template>
  <section v-if="canReview" class="appointment-review" aria-labelledby="appointment-review-title">
    <h2 id="appointment-review-title" class="appointment-review__title">
      Califica a {{ professionalName }}
    </h2>
    <p class="appointment-review__lead">Tu opinión ayuda a otras personas a elegir. Es pública.</p>

    <form class="appointment-review__form" novalidate @submit.prevent="onSubmit">
      <StarRatingInput
        v-model="rating"
        legend="Tu calificación"
        :error="ratingError"
        :disabled="submitting"
      />

      <div class="appointment-review__field">
        <label for="appointment-review-comment" class="appointment-review__label">
          Comentario <span class="appointment-review__optional">(opcional)</span>
        </label>
        <textarea
          id="appointment-review-comment"
          v-model="comment"
          class="appointment-review__textarea"
          rows="4"
          name="comment"
          :disabled="submitting"
          :aria-invalid="!!commentError"
          aria-describedby="appointment-review-hint"
        />
        <div class="appointment-review__foot">
          <p id="appointment-review-hint" class="appointment-review__hint">
            No escribas teléfonos, direcciones ni nombres de otras personas.
          </p>
          <span
            class="appointment-review__counter"
            :class="{ 'appointment-review__counter--over': overLimit }"
            aria-hidden="true"
            >{{ length }} / {{ REVIEW_COMMENT_MAX_LENGTH }}</span
          >
        </div>
        <p v-if="commentError" class="appointment-review__error" role="alert">
          {{ commentError }}
        </p>
      </div>

      <BaseAlert v-if="errorMessage" variant="danger" role="alert">{{ errorMessage }}</BaseAlert>

      <BaseButton type="submit" variant="primary" :loading="submitting">Enviar reseña</BaseButton>
    </form>
  </section>

  <section
    v-else-if="props.review"
    class="appointment-review"
    aria-labelledby="appointment-review-title"
  >
    <h2 id="appointment-review-title" class="appointment-review__title">Gracias por tu reseña</h2>
    <p class="appointment-review__lead">Así la ven quienes eligen a {{ professionalName }}.</p>
    <StarRating :rating="props.review.rating" :count="1" :show-value="false" />
    <p v-if="props.review.comment" class="appointment-review__comment">
      {{ props.review.comment }}
    </p>
  </section>
</template>

<style scoped>
.appointment-review {
  display: flex;
  width: 100%;
  flex-direction: column;
  align-items: stretch;
  gap: var(--space-3);
  padding-top: var(--space-5);
  text-align: left;
  border-top: var(--border-width-normal) solid var(--color-border-subtle);
}

.appointment-review__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: 24px;
  line-height: 32px;
  color: var(--color-text-primary);
}

.appointment-review__lead,
.appointment-review__comment {
  margin: 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-text-secondary);
}

.appointment-review__comment {
  color: var(--color-text-primary);
  overflow-wrap: anywhere;
}

.appointment-review__form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  color: var(--color-text-primary);
}

.appointment-review__field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.appointment-review__label {
  font-family: var(--font-family-base);
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.appointment-review__optional {
  font-weight: 400;
  color: var(--color-text-secondary);
}

.appointment-review__textarea {
  width: 100%;
  min-height: 96px;
  padding: 12px var(--space-4);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  line-height: var(--font-size-body-line);
  color: var(--color-text-primary);
  resize: vertical;
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-control);
  border-radius: 2px;
}

.appointment-review__textarea:focus-visible {
  outline: var(--focus-ring-width, 2px) solid var(--color-focus);
  outline-offset: 2px;
}

.appointment-review__foot {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}

.appointment-review__hint {
  margin: 0;
  font-family: var(--font-family-base);
  font-size: 13px;
  color: var(--color-text-secondary);
}

.appointment-review__counter {
  flex: none;
  font-family: var(--font-family-base);
  font-size: 13px;
  color: var(--color-text-secondary);
}

.appointment-review__counter--over,
.appointment-review__error {
  margin: 0;
  font-family: var(--font-family-base);
  font-size: 13px;
  color: var(--color-danger-text);
}
</style>
