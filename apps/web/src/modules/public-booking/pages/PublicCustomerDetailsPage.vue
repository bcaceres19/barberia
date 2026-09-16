<script setup lang="ts">
// Página de captura de datos del cliente y persona atendida (HU-096,
// CA-096-01 a CA-096-06) y confirmación pública concurrente (HU-097,
// CA-097-01 a CA-097-07). Sin mockup asignado: composición libre dentro de
// NAVA / Tailored Grid (DEC-078). Valida/normaliza en el cliente con las
// mismas reglas que el backend aplicará (identity.go), conserva los datos
// ante un error de validación o conflicto (CA-096-03, CA-097-03) y resuelve
// `attendeeName` sin pedirlo dos veces cuando el cliente reserva para sí
// mismo (CA-096-01, RN-RES-02).
import { computed, nextTick, ref } from 'vue'
import { BaseButton, BaseInput } from '@/shared/ui'
import { confirmPublicAppointment } from '../api/confirmPublicAppointmentApi'
import { newIdempotencyKey } from '../model/idempotencyKey'
import type {
  ConfirmedPublicAppointment,
  PublicAppointmentAlternative,
} from '../model/confirmPublicAppointmentOutcome'
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
   * `:slug`. */
  slug: string
  /** Servicio ya elegido (HU-091). */
  serviceId: string
  /** Barbero ya elegido (HU-092). */
  barberId: string
  /** Franja ya elegida (HU-095), instante ISO absoluto. El servidor la
   * revalida por completo al confirmar (CA-097-02); esta página solo la
   * conserva y la envía tal cual. */
  startsAt: string
}
const props = defineProps<Props>()

type AttendeeChoice = 'self' | 'other'

const attendeeChoice = ref<AttendeeChoice>('self')
const fullName = ref('')
const phone = ref('')
const email = ref('')
const note = ref('')
const attendeeName = ref('')

const attempted = ref(false)
const reviewing = ref(false)

// selectedStartsAt permite reintentar con una alternativa (RN-CON-05,
// DEC-090) sin recargar la página ni perder el resto del formulario.
const selectedStartsAt = ref(props.startsAt)

type ConfirmState = 'idle' | 'submitting' | 'confirmed' | 'schedule-conflict' | 'error'
const confirmState = ref<ConfirmState>('idle')
const confirmedAppointment = ref<ConfirmedPublicAppointment | null>(null)
const alternatives = ref<PublicAppointmentAlternative[]>([])
const errorMessage = ref('')
// idempotencyKey se conserva mientras el reintento es la MISMA solicitud
// (red/error inesperado, RN-IDE-01): un contenido distinto (editar, elegir
// otra alternativa) siempre pide una clave nueva.
const idempotencyKey = ref<string | null>(null)

const formRef = ref<HTMLElement | null>(null)
const summaryRef = ref<HTMLElement | null>(null)
const confirmedRef = ref<HTMLElement | null>(null)

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
  confirmState.value = 'idle'
  alternatives.value = []
  errorMessage.value = ''
  idempotencyKey.value = null
  await nextTick()
  formRef.value?.querySelector<HTMLElement>('input, textarea')?.focus()
}

function formatLocalDateTime(iso: string): string {
  try {
    return new Date(iso).toLocaleString('es-CO', {
      dateStyle: 'full',
      timeStyle: 'short',
    })
  } catch {
    return iso
  }
}

// CA-097-07: bloquea doble toque (confirmState guarda 'submitting' hasta
// que el intento termina) y conserva los datos del formulario en
// cualquier desenlace (nunca se limpian fuera de 'confirmed').
async function handleConfirm() {
  if (confirmState.value === 'submitting') return

  if (!idempotencyKey.value) {
    idempotencyKey.value = newIdempotencyKey()
  }

  confirmState.value = 'submitting'
  errorMessage.value = ''

  const outcome = await confirmPublicAppointment(
    props.slug,
    props.serviceId,
    props.barberId,
    {
      startsAt: selectedStartsAt.value,
      fullName: fullName.value.trim(),
      phone: phone.value.trim(),
      email: email.value.trim(),
      note: note.value.trim() || null,
      forSomeoneElse: attendeeChoice.value === 'other',
      attendeeName: attendeeChoice.value === 'other' ? attendeeName.value.trim() : null,
    },
    idempotencyKey.value,
  )

  switch (outcome.kind) {
    case 'success':
      confirmedAppointment.value = outcome.appointment
      confirmState.value = 'confirmed'
      idempotencyKey.value = null
      await nextTick()
      confirmedRef.value?.focus()
      return
    case 'schedule-conflict':
      alternatives.value = outcome.alternatives
      confirmState.value = 'schedule-conflict'
      idempotencyKey.value = null
      return
    case 'not-found':
      errorMessage.value =
        'Este enlace de reserva ya no está disponible. Vuelve a intentarlo desde el inicio.'
      confirmState.value = 'error'
      idempotencyKey.value = null
      return
    case 'validation-error':
      errorMessage.value = outcome.detail
      confirmState.value = 'error'
      idempotencyKey.value = null
      return
    case 'idempotency-conflict':
      errorMessage.value = 'Hubo un problema al procesar tu confirmación. Intenta de nuevo.'
      confirmState.value = 'error'
      idempotencyKey.value = null
      return
    case 'network-error':
      errorMessage.value =
        'No pudimos conectar con el servidor. Verifica tu conexión e intenta de nuevo.'
      confirmState.value = 'error'
      return
    case 'unexpected-error':
      errorMessage.value = 'Ocurrió un error inesperado. Intenta de nuevo.'
      confirmState.value = 'error'
      return
  }
}

function chooseAlternative(startsAt: string) {
  selectedStartsAt.value = startsAt
  alternatives.value = []
  confirmState.value = 'idle'
  idempotencyKey.value = null
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

      <!-- Resumen y confirmación real (HU-096/HU-097, CA-097-07): exige
           confirmación explícita antes de reservar (RN-DIS-03), nunca
           insinúa que el turno ya quedó reservado hasta que el servidor lo
           confirme. -->
      <div
        v-else-if="confirmState !== 'confirmed'"
        ref="summaryRef"
        class="customer-details__summary"
        role="status"
        tabindex="-1"
      >
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
          <div class="customer-details__summary-row">
            <dt>Fecha y hora</dt>
            <dd>{{ formatLocalDateTime(selectedStartsAt) }}</dd>
          </div>
        </dl>

        <p
          v-if="confirmState === 'error'"
          class="customer-details__error customer-details__banner"
          role="alert"
        >
          {{ errorMessage }}
        </p>

        <div
          v-if="confirmState === 'schedule-conflict'"
          class="customer-details__conflict"
          role="alert"
        >
          <p>Ese horario se acaba de ocupar.</p>
          <p v-if="alternatives.length > 0">Estas horas siguen libres:</p>
          <ul v-if="alternatives.length > 0" class="customer-details__alternatives">
            <li v-for="alt in alternatives" :key="alt.startsAt">
              <BaseButton type="button" variant="soft" @click="chooseAlternative(alt.startsAt)">
                {{ formatLocalDateTime(alt.startsAt) }}
              </BaseButton>
            </li>
          </ul>
          <p v-else>No quedan horarios cercanos disponibles. Elige otro día.</p>
        </div>

        <div class="customer-details__actions">
          <BaseButton
            type="button"
            variant="secondary"
            :disabled="confirmState === 'submitting'"
            @click="editAgain"
          >
            Editar
          </BaseButton>
          <BaseButton
            type="button"
            variant="primary"
            :loading="confirmState === 'submitting'"
            @click="handleConfirm"
          >
            Confirmar turno
          </BaseButton>
        </div>
      </div>

      <!-- Confirmación real del servidor (CA-097-01, CA-097-07): el correo
           con el enlace de acceso ya se envió; el token en claro solo
           existe en esta respuesta (DEC-089), nunca se vuelve a mostrar. -->
      <div
        v-else
        ref="confirmedRef"
        class="customer-details__confirmed"
        role="status"
        tabindex="-1"
      >
        <h2 class="customer-details__summary-title">¡Tu turno quedó confirmado!</h2>
        <dl v-if="confirmedAppointment" class="customer-details__summary-list">
          <div class="customer-details__summary-row">
            <dt>Barbería</dt>
            <dd>{{ confirmedAppointment.barbershopName }}</dd>
          </div>
          <div class="customer-details__summary-row">
            <dt>Servicio</dt>
            <dd>{{ confirmedAppointment.serviceName }}</dd>
          </div>
          <div class="customer-details__summary-row">
            <dt>Atiende a</dt>
            <dd>{{ confirmedAppointment.attendeeName }}</dd>
          </div>
          <div class="customer-details__summary-row">
            <dt>Fecha y hora</dt>
            <dd>{{ formatLocalDateTime(confirmedAppointment.startsAt) }}</dd>
          </div>
        </dl>
        <p>Te enviamos un correo con el enlace para consultar tu turno cuando quieras.</p>
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

.customer-details__banner {
  margin: 0;
  padding: var(--space-3);
  background-color: var(--color-surface-muted);
  border-radius: 4px;
}

.customer-details__conflict {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  margin: 0;
  padding: var(--space-3);
  background-color: var(--color-surface-muted);
  border-radius: 4px;
}

.customer-details__alternatives {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin: 0;
  padding: 0;
  list-style: none;
}

.customer-details__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.customer-details__confirmed {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding: var(--space-4);
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: 4px;
}

@media (min-width: 1024px) {
  .customer-details__title {
    font-size: 40px;
    line-height: 48px;
  }
}
</style>
