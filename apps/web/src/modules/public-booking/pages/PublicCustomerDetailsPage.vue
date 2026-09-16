<script setup lang="ts">
// Página de captura de datos del cliente y persona atendida (HU-096,
// CA-096-01 a CA-096-06). Sin mockup asignado: composición libre dentro de
// NAVA / Tailored Grid (DEC-078). No llama a ningún endpoint: HU-097
// (bloqueada hoy por DP-PUB-05/DP-PUB-06/CT-011, docs/00-control/
// dudas-pendientes.md) es quien revalida y persiste. Esta pantalla solo
// valida/normaliza en el cliente con las mismas reglas que el backend
// aplicará (identity.go), conserva los datos ante un error de validación
// (CA-096-03) y resuelve `attendeeName` sin pedirlo dos veces cuando el
// cliente reserva para sí mismo (CA-096-01, RN-RES-02).
import { computed, nextTick, ref } from 'vue'
import { BaseButton, BaseInput } from '@/shared/ui'
import {
  ATTENDEE_NAME_MAX_LENGTH,
  CUSTOMER_NOTE_MAX_LENGTH,
  validateAttendeeName,
  validateCustomerEmail,
  validateCustomerFullName,
  validateCustomerNote,
  validateCustomerPhone,
} from '../validation/customerIdentityValidation'

interface Props {
  /** Identificador del enlace público, tal como llega del parámetro de ruta
   * `:slug`. No se usa en esta página: no hay ninguna llamada al servidor. */
  slug: string
  /** Servicio ya elegido (HU-091). Igual que slug, viaja intacto para que
   * HU-097 reciba el contexto completo sin pedirlo de nuevo. */
  serviceId: string
  /** Barbero ya elegido (HU-092). */
  barberId: string
  /** Franja ya elegida (HU-095), instante ISO absoluto. Esta página no la
   * interpreta ni la revalida (HU-097 lo hará): solo la conserva. */
  startsAt: string
}
// Los cuatro campos viajan intactos en la URL (props: true) para que HU-097
// reciba el contexto completo sin pedirlo de nuevo; esta página no los lee
// -no hay ninguna llamada al servidor todavía- así que defineProps no se
// asigna a una variable (documenta el contrato de la ruta sin una lectura
// sin uso).
defineProps<Props>()

type AttendeeChoice = 'self' | 'other'

const attendeeChoice = ref<AttendeeChoice>('self')
const fullName = ref('')
const phone = ref('')
const email = ref('')
const note = ref('')
const attendeeName = ref('')

const attempted = ref(false)
const reviewing = ref(false)

const formRef = ref<HTMLElement | null>(null)
const summaryRef = ref<HTMLElement | null>(null)

const fieldErrors = computed(() => ({
  fullName: validateCustomerFullName(fullName.value),
  phone: validateCustomerPhone(phone.value),
  email: validateCustomerEmail(email.value),
  note: validateCustomerNote(note.value),
  attendeeName:
    attendeeChoice.value === 'other' ? validateAttendeeName(attendeeName.value) : undefined,
}))

const hasErrors = computed(() => Object.values(fieldErrors.value).some((message) => !!message))

// CA-096-01: para sí mismo, el nombre atendido se deriva del cliente sin
// pedirlo dos veces.
const resolvedAttendeeName = computed(() =>
  attendeeChoice.value === 'self' ? fullName.value.trim() : attendeeName.value.trim(),
)

async function handleSubmit() {
  attempted.value = true
  if (hasErrors.value) {
    await nextTick()
    const firstErrorField = document.querySelector<HTMLElement>('[aria-invalid="true"]')
    firstErrorField?.focus()
    return
  }
  reviewing.value = true
  await nextTick()
  summaryRef.value?.focus()
}

async function editAgain() {
  reviewing.value = false
  await nextTick()
  formRef.value?.querySelector<HTMLElement>('input, textarea')?.focus()
}
</script>

<template>
  <main class="customer-details">
    <div class="customer-details__container">
      <h1 class="customer-details__title">Tus datos</h1>

      <form v-if="!reviewing" ref="formRef" novalidate @submit.prevent="handleSubmit">
        <fieldset class="customer-details__fieldset">
          <legend class="customer-details__legend">¿Para quién es el turno?</legend>
          <div
            class="customer-details__radio-group"
            role="radiogroup"
            aria-label="¿Para quién es el turno?"
          >
            <label class="customer-details__radio">
              <input v-model="attendeeChoice" type="radio" name="attendee-choice" value="self" />
              Para mí
            </label>
            <label class="customer-details__radio">
              <input v-model="attendeeChoice" type="radio" name="attendee-choice" value="other" />
              Para otra persona
            </label>
          </div>
        </fieldset>

        <BaseInput
          v-model="fullName"
          type="text"
          label="Tu nombre"
          autocomplete="name"
          required
          :error="attempted ? fieldErrors.fullName : undefined"
        />

        <BaseInput
          v-if="attendeeChoice === 'other'"
          v-model="attendeeName"
          type="text"
          label="Nombre de la persona atendida"
          autocomplete="off"
          required
          :maxlength="ATTENDEE_NAME_MAX_LENGTH"
          :error="attempted ? fieldErrors.attendeeName : undefined"
        />

        <BaseInput
          v-model="phone"
          type="tel"
          label="Teléfono"
          placeholder="+573001234567"
          autocomplete="tel"
          required
          :error="attempted ? fieldErrors.phone : undefined"
        />

        <BaseInput
          v-model="email"
          type="email"
          label="Correo"
          autocomplete="email"
          required
          :error="attempted ? fieldErrors.email : undefined"
        />

        <div class="customer-details__field">
          <label for="customer-details-note" class="customer-details__label">Nota (opcional)</label>
          <textarea
            id="customer-details-note"
            v-model="note"
            class="customer-details__textarea"
            rows="2"
            :aria-invalid="attempted && !!fieldErrors.note"
            :aria-describedby="
              attempted && fieldErrors.note
                ? 'customer-details-note-error'
                : 'customer-details-note-counter'
            "
          />
          <div class="customer-details__field-foot">
            <span
              v-if="attempted && fieldErrors.note"
              id="customer-details-note-error"
              class="customer-details__error"
              role="alert"
            >
              {{ fieldErrors.note }}
            </span>
            <span v-else />
            <span id="customer-details-note-counter" class="customer-details__counter">
              {{ note.length }}/{{ CUSTOMER_NOTE_MAX_LENGTH }}
            </span>
          </div>
        </div>

        <BaseButton type="submit" variant="primary">Ver resumen</BaseButton>
      </form>

      <!-- Resumen local, sin llamar al servidor (HU-097 no existe todavía):
           permite revisar antes de que la confirmación real esté
           disponible, sin insinuar que el turno ya quedó reservado
           (RN-DIS-03). -->
      <div v-else ref="summaryRef" class="customer-details__summary" role="status" tabindex="-1">
        <h2 class="customer-details__summary-title">Revisa tus datos</h2>
        <dl class="customer-details__summary-list">
          <div class="customer-details__summary-row">
            <dt>Cliente</dt>
            <dd>{{ fullName }}</dd>
          </div>
          <div class="customer-details__summary-row">
            <dt>Atiende a</dt>
            <dd>{{ resolvedAttendeeName }}</dd>
          </div>
          <div class="customer-details__summary-row">
            <dt>Teléfono</dt>
            <dd>{{ phone }}</dd>
          </div>
          <div class="customer-details__summary-row">
            <dt>Correo</dt>
            <dd>{{ email }}</dd>
          </div>
          <div v-if="note.trim()" class="customer-details__summary-row">
            <dt>Nota</dt>
            <dd>{{ note }}</dd>
          </div>
        </dl>
        <BaseButton type="button" variant="secondary" @click="editAgain">Editar</BaseButton>
      </div>
    </div>
  </main>
</template>

<style scoped>
.customer-details {
  display: flex;
  min-height: 100dvh;
  justify-content: center;
  padding: var(--space-6) var(--space-4);
  background-color: var(--color-canvas);
}

.customer-details__container {
  display: flex;
  width: 100%;
  max-width: 640px;
  flex-direction: column;
  gap: var(--space-5);
}

.customer-details__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: 32px;
  line-height: 40px;
  letter-spacing: -0.015em;
  color: var(--color-text-primary);
}

.customer-details__container form,
.customer-details__summary {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.customer-details__fieldset {
  padding: 0;
  margin: 0;
  border: none;
}

.customer-details__legend {
  padding: 0;
  margin: 0 0 var(--space-2);
  font-weight: 600;
  color: var(--color-text-primary);
}

.customer-details__radio-group {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-4);
}

.customer-details__radio {
  display: flex;
  min-height: 44px;
  align-items: center;
  gap: var(--space-2);
  color: var(--color-text-primary);
}

.customer-details__field {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.customer-details__label {
  font-weight: 500;
  color: var(--color-text-primary);
}

.customer-details__textarea {
  padding: var(--space-3);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-text-primary);
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: 4px;
  resize: vertical;
}

.customer-details__field-foot {
  display: flex;
  justify-content: space-between;
  gap: var(--space-2);
}

.customer-details__error {
  color: var(--color-danger);
}

.customer-details__counter {
  color: var(--color-text-secondary);
}

.customer-details__summary {
  padding: var(--space-4);
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: 4px;
}

.customer-details__summary-title {
  margin: 0;
  font-size: 20px;
  color: var(--color-text-primary);
}

.customer-details__summary-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  margin: 0;
}

.customer-details__summary-row {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: var(--space-2);
  border-bottom: var(--border-width-normal) solid var(--color-border-subtle);
  padding-bottom: var(--space-2);
}

.customer-details__summary-row dt {
  font-weight: 600;
  color: var(--color-text-secondary);
}

.customer-details__summary-row dd {
  margin: 0;
  color: var(--color-text-primary);
  text-align: right;
}

@media (min-width: 1024px) {
  .customer-details__title {
    font-size: 40px;
    line-height: 48px;
  }
}
</style>
