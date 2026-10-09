<script setup lang="ts">
// Página de captura de datos del cliente y persona atendida (HU-096,
// CA-096-01 a CA-096-06) y confirmación pública concurrente (HU-097,
// CA-097-01 a CA-097-07). Sin mockup asignado: composición libre dentro de
// NAVA / Tailored Grid (DEC-078), alojada en el cascarón de la reserva
// pública (DEC-111): lo que se rellena vive en una hoja de papel marfil
// sobre el lienzo de tinta. Valida/normaliza en el cliente con las
// mismas reglas que el backend aplicará (identity.go), conserva los datos
// ante un error de validación o conflicto (CA-096-03, CA-097-03) y resuelve
// `attendeeName` sin pedirlo dos veces cuando el cliente reserva para sí
// mismo (CA-096-01, RN-RES-02).
import { computed, nextTick, ref } from 'vue'
import { BaseAlert, BaseButton, BaseInput } from '@/shared/ui'
import { confirmPublicAppointment } from '../api/confirmPublicAppointmentApi'
import { useBookingChrome } from '../model/bookingChrome'
import { newIdempotencyKey } from '../model/idempotencyKey'
import { usePublicVocabulary } from '../model/publicVocabulary'
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
const v = usePublicVocabulary()

// El cascarón dibuja el progreso: al confirmar, se completa.
const chrome = useBookingChrome()

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
      chrome.completed.value = true
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
  <main class="pb-page pb-page--center pb-page--fit">
    <p class="pb-eyebrow">
      {{
        confirmState === 'confirmed'
          ? 'Reserva completa'
          : reviewing
            ? 'Última revisión'
            : 'Casi listo'
      }}
    </p>
    <h1 class="pb-title">{{ confirmState === 'confirmed' ? 'Todo listo' : 'Tus datos' }}</h1>
    <div class="pb-rule" aria-hidden="true"></div>

    <section class="pb-sheet customer-details">
      <form
        v-if="!reviewing"
        ref="formRef"
        class="pb-sheet__form customer-details__rise"
        novalidate
        @submit.prevent="handleSubmit"
      >
        <fieldset class="customer-details__fieldset">
          <legend class="customer-details__legend">¿Para quién es el turno?</legend>
          <div
            class="customer-details__radio-group"
            role="radiogroup"
            aria-label="¿Para quién es el turno?"
          >
            <label class="customer-details__radio">
              <input v-model="attendeeChoice" type="radio" name="attendee-choice" value="self" />
              <span>Para mí</span>
            </label>
            <label class="customer-details__radio">
              <input v-model="attendeeChoice" type="radio" name="attendee-choice" value="other" />
              <span>Para otra persona</span>
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
          class="customer-details__rise"
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

        <BaseButton type="submit" variant="primary" size="lg" class="customer-details__submit"
          >Ver resumen</BaseButton
        >
      </form>

      <!-- Resumen y confirmación real (HU-096/HU-097, CA-097-07): exige
           confirmación explícita antes de reservar (RN-DIS-03), nunca
           insinúa que el turno ya quedó reservado hasta que el servidor lo
           confirme. -->
      <div
        v-else-if="confirmState !== 'confirmed'"
        ref="summaryRef"
        class="customer-details__summary customer-details__rise"
        role="status"
        tabindex="-1"
      >
        <h2 class="pb-sheet__heading">Revisa tus datos</h2>
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

        <!-- DEC-122: si el negocio publica su agenda en Google Calendar, Google invita al
             cliente al turno en el correo que dio. El aviso es condicional porque esta
             pantalla no sabe si el negocio lo usa; nunca promete una invitación. -->
        <p class="customer-details__invite-note">
          Si el negocio usa Google Calendar, recibirás una invitación a este turno en
          <strong>{{ email }}</strong
          >.
        </p>

        <BaseAlert v-if="confirmState === 'error'" variant="danger" role="alert">
          {{ errorMessage }}
        </BaseAlert>

        <BaseAlert v-if="confirmState === 'schedule-conflict'" variant="warning" role="alert">
          <p class="customer-details__conflict-line">Ese horario se acaba de ocupar.</p>
          <p v-if="alternatives.length > 0" class="customer-details__conflict-line">
            Estas horas siguen libres:
          </p>
          <ul v-if="alternatives.length > 0" class="customer-details__alternatives">
            <li v-for="alt in alternatives" :key="alt.startsAt">
              <BaseButton type="button" variant="soft" @click="chooseAlternative(alt.startsAt)">
                {{ formatLocalDateTime(alt.startsAt) }}
              </BaseButton>
            </li>
          </ul>
          <p v-else class="customer-details__conflict-line">
            No quedan horarios cercanos disponibles. Elige otro día.
          </p>
        </BaseAlert>

        <div class="customer-details__actions">
          <BaseButton
            type="button"
            variant="secondary"
            size="lg"
            :disabled="confirmState === 'submitting'"
            @click="editAgain"
          >
            Editar
          </BaseButton>
          <BaseButton
            type="button"
            variant="primary"
            size="lg"
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
        class="customer-details__confirmed customer-details__rise"
        role="status"
        tabindex="-1"
      >
        <span class="customer-details__seal" aria-hidden="true">
          <svg viewBox="0 0 96 96" focusable="false">
            <rect
              class="customer-details__seal-diamond"
              pathLength="1"
              x="22"
              y="22"
              width="52"
              height="52"
            />
            <path class="customer-details__seal-check" pathLength="1" d="M33 49l11 11 21-23" />
          </svg>
          <i
            v-for="spark in 8"
            :key="spark"
            class="customer-details__spark"
            :style="{
              '--spark-angle': `${(spark - 1) * 45}deg`,
              '--spark-delay': `${0.7 + (spark % 3) * 0.06}s`,
            }"
          ></i>
        </span>
        <h2 class="pb-sheet__heading">¡Tu turno quedó confirmado!</h2>
        <dl v-if="confirmedAppointment" class="customer-details__summary-list">
          <div class="customer-details__summary-row">
            <dt>{{ v.Business }}</dt>
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
        <p class="customer-details__note">
          Te enviamos un correo con el enlace para consultar tu turno cuando quieras.
        </p>
      </div>
    </section>
  </main>
</template>

<style scoped>
/* Todo lo que sigue vive sobre el panel de tinta levantada (`.pb-sheet`), el
   mismo lienzo de las demás pantallas de la reserva: los campos, el resumen
   y la confirmación se recomponen con los tokens de tinta y latón. `BaseInput`
   y `BaseButton` se pensaron para superficies claras, así que aquí se les
   ajustan sus variables y colores sin tocar el componente compartido. */
.customer-details__rise {
  animation: pb-rise 0.55s var(--pb-ease, ease) 0.1s backwards;
}

/* Los campos del formulario entran escalonados, uno tras otro. */
.pb-sheet__form > * {
  animation: pb-rise 0.6s var(--pb-ease, ease) calc(0.35s + var(--i, 0) * 0.07s) backwards;
}

.pb-sheet__form > :nth-child(2) {
  --i: 1;
}

.pb-sheet__form > :nth-child(3) {
  --i: 2;
}

.pb-sheet__form > :nth-child(4) {
  --i: 3;
}

.pb-sheet__form > :nth-child(5) {
  --i: 4;
}

.pb-sheet__form > :nth-child(6) {
  --i: 5;
}

.pb-sheet__form > :nth-child(7) {
  --i: 6;
}

.customer-details__fieldset {
  min-width: 0;
  padding: 0;
  margin: 0;
  border: none;
}

/* Rótulo reglado: versalitas de latón, el mismo de BaseInput. */
.customer-details__legend,
.customer-details__label {
  padding: 0;
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  line-height: 14px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--pb-brass);
}

.customer-details__legend {
  margin: 0 0 var(--space-2);
}

/* ── Campos de BaseInput sobre tinta ───────────────────────────────────── */

.customer-details :deep(.base-input) {
  --input-bg: color-mix(in srgb, #000 46%, var(--color-surface-strong));
  --input-border-color: var(--pb-line-strong);
  --input-border-base-color: color-mix(in srgb, var(--pb-brass) 75%, transparent);
  --input-focus-ring:
    0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--pb-brass),
    0 0 22px color-mix(in srgb, var(--pb-brass) 35%, transparent);

  color: var(--pb-text);
}

.customer-details :deep(.base-input::placeholder) {
  color: var(--pb-muted);
  opacity: 1;
}

.customer-details :deep(.base-input:hover:not(:disabled):not(.base-input--invalid)) {
  background-color: color-mix(in srgb, #000 38%, var(--color-surface-strong));
  border-color: color-mix(in srgb, var(--pb-brass) 60%, transparent);
  border-bottom-color: var(--pb-brass);
}

.customer-details :deep(.base-input:focus-visible) {
  background-color: color-mix(in srgb, #000 52%, var(--color-surface-strong));
  border-color: var(--pb-brass);
  border-bottom-color: var(--pb-brass);
}

.customer-details :deep(.base-input:-webkit-autofill),
.customer-details :deep(.base-input:-webkit-autofill:hover),
.customer-details :deep(.base-input:-webkit-autofill:focus) {
  -webkit-text-fill-color: var(--pb-text);
  caret-color: var(--pb-text);
}

.customer-details :deep(.base-input__label) {
  letter-spacing: 0.12em;
  color: var(--pb-brass);
}

.customer-details :deep(.base-input__required) {
  color: var(--color-danger-on-strong);
}

.customer-details :deep(.base-input--invalid),
.customer-details :deep(.base-input--invalid:hover) {
  border-color: var(--color-danger-on-strong);
  border-bottom-color: var(--color-danger-on-strong);
}

.customer-details :deep(.base-input--invalid:focus-visible) {
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-danger-on-strong);
}

.customer-details :deep(.base-input__hint) {
  color: var(--pb-muted);
}

.customer-details :deep(.base-input__error),
.customer-details__error {
  color: var(--color-danger-on-strong);
}

/* «Para mí / Para otra persona» como control segmentado. El `<input>` nativo
   sigue siendo el control (foco, flechas, formulario); solo se oculta su
   caja y la etiqueta hace de segmento. */
.customer-details__radio-group {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(128px, 1fr));
  gap: var(--space-2);
}

.customer-details__radio {
  position: relative;
  display: flex;
  min-height: var(--control-height);
  cursor: pointer;
}

.customer-details__radio input {
  position: absolute;
  inset: 0;
  z-index: 1;
  width: 100%;
  height: 100%;
  margin: 0;
  cursor: pointer;
  opacity: 0;
}

.customer-details__radio span {
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: center;
  padding: 0 var(--space-4);
  font-size: var(--font-size-body);
  font-weight: 500;
  color: var(--pb-text);
  text-align: center;
  background-color: transparent;
  border: var(--border-width-normal) solid var(--pb-line-strong);
  border-bottom: var(--border-width-emphasis) solid var(--pb-line-strong);
  border-radius: 2px;
  transition:
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard),
    color var(--motion-duration-base) var(--motion-easing-standard),
    transform var(--motion-duration-base) var(--pb-ease, ease);
}

.customer-details__radio:hover input:not(:checked) + span {
  background-color: var(--pb-wash);
  border-color: var(--pb-brass);
  transform: translateY(-1px);
}

.customer-details__radio input:checked + span {
  font-weight: 600;
  color: var(--pb-on-brass);
  background-color: var(--pb-brass);
  border-color: var(--pb-brass);
  border-bottom-color: color-mix(in srgb, var(--pb-brass) 70%, #000);
}

/* El segmento elegido también lleva un rombo: no depende solo del relleno. */
.customer-details__radio input:checked + span::before {
  content: '';
  width: 7px;
  height: 7px;
  margin-right: var(--space-2);
  background-color: var(--pb-on-brass);
  transform: rotate(45deg);
  animation: pb-pop 0.4s cubic-bezier(0.3, 1.5, 0.5, 1);
}

.customer-details__radio input:focus-visible + span {
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--pb-brass);
}

.customer-details__field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

/* Área de nota: el mismo campo reglado que BaseInput (recuesto oscuro, filete
   perimetral y línea base de latón de 2px). */
.customer-details__textarea {
  padding: var(--space-3) var(--space-4);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  line-height: var(--font-size-body-line);
  color: var(--pb-text);
  resize: vertical;
  background-color: color-mix(in srgb, #000 46%, var(--color-surface-strong));
  border: var(--border-width-normal) solid var(--pb-line-strong);
  border-bottom: var(--border-width-emphasis) solid
    color-mix(in srgb, var(--pb-brass) 75%, transparent);
  border-radius: 2px;
  outline: none;
  transition:
    border-color var(--motion-duration-fast) var(--motion-easing-standard),
    box-shadow var(--motion-duration-fast) var(--motion-easing-standard),
    background-color var(--motion-duration-fast) var(--motion-easing-standard);
}

.customer-details__textarea:hover {
  background-color: color-mix(in srgb, #000 38%, var(--color-surface-strong));
  border-color: color-mix(in srgb, var(--pb-brass) 60%, transparent);
  border-bottom-color: var(--pb-brass);
}

.customer-details__textarea:focus-visible {
  background-color: color-mix(in srgb, #000 52%, var(--color-surface-strong));
  border-color: var(--pb-brass);
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--pb-brass),
    0 0 22px color-mix(in srgb, var(--pb-brass) 35%, transparent);
}

.customer-details__textarea[aria-invalid='true'] {
  border-color: var(--color-danger-on-strong);
  border-bottom-color: var(--color-danger-on-strong);
}

.customer-details__field-foot {
  display: flex;
  gap: var(--space-2);
  justify-content: space-between;
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
}

.customer-details__counter {
  color: var(--pb-muted);
}

/* ── Botones de BaseButton en latón sobre tinta ────────────────────────── */

.customer-details :deep(.base-button) {
  font-weight: 600;
}

.customer-details :deep(.base-button--primary) {
  color: var(--pb-on-brass);
  background-color: var(--pb-brass);
  border-color: var(--pb-brass);
  border-bottom-width: var(--border-width-emphasis);
  border-bottom-color: color-mix(in srgb, var(--pb-brass) 70%, #000);
}

.customer-details :deep(.base-button--primary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--pb-brass) 86%, #fff);
  border-color: color-mix(in srgb, var(--pb-brass) 86%, #fff);
  border-bottom-color: color-mix(in srgb, var(--pb-brass) 70%, #000);
  box-shadow: 0 10px 24px -12px color-mix(in srgb, var(--pb-brass) 70%, transparent);
}

.customer-details :deep(.base-button--primary:active:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--pb-brass) 88%, #000);
  border-color: color-mix(in srgb, var(--pb-brass) 88%, #000);
}

.customer-details :deep(.base-button--secondary),
.customer-details :deep(.base-button--soft) {
  color: var(--pb-brass);
  background-color: transparent;
  border-color: var(--pb-brass);
  border-bottom-width: var(--border-width-emphasis);
}

.customer-details :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)),
.customer-details :deep(.base-button--soft:hover:not(:disabled):not(.base-button--loading)) {
  background-color: var(--pb-wash-strong);
  border-color: var(--pb-brass);
}

.customer-details :deep(.base-button--secondary:active:not(:disabled):not(.base-button--loading)),
.customer-details :deep(.base-button--soft:active:not(:disabled):not(.base-button--loading)) {
  background-color: var(--pb-wash-strong);
}

.customer-details :deep(.base-button:focus-visible) {
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-on-strong);
}

.customer-details__submit {
  width: 100%;
  margin-top: var(--space-2);
}

.customer-details__summary,
.customer-details__confirmed {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  outline: none;
}

.customer-details__confirmed .pb-sheet__heading {
  text-align: center;
}

.customer-details__summary-list {
  display: flex;
  flex-direction: column;
  margin: 0;
  border-top: 1px solid var(--pb-line-strong);
}

.customer-details__invite-note {
  margin: 0;
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
  color: var(--pb-muted);
  overflow-wrap: anywhere;
}

.customer-details__summary-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1) var(--space-6);
  justify-content: space-between;
  padding: var(--space-3) 0;
  border-bottom: 1px solid var(--pb-line);
  animation: pb-rise 0.5s var(--pb-ease, ease) backwards;
}

.customer-details__summary-row:nth-child(2) {
  animation-delay: 0.08s;
}

.customer-details__summary-row:nth-child(3) {
  animation-delay: 0.16s;
}

.customer-details__summary-row:nth-child(4) {
  animation-delay: 0.24s;
}

.customer-details__summary-row:nth-child(5) {
  animation-delay: 0.32s;
}

.customer-details__summary-row:nth-child(n + 6) {
  animation-delay: 0.4s;
}

.customer-details__summary-row dt {
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  line-height: 24px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--pb-brass);
}

.customer-details__summary-row dd {
  min-width: 0;
  margin: 0;
  overflow-wrap: anywhere;
  color: var(--pb-text);
  text-align: right;
}

/* Avisos (error y conflicto) sobre tinta: baño tenue del color de estado y
   filete lateral levantado al lienzo oscuro. */
.customer-details :deep(.base-alert) {
  color: var(--pb-text);
  background-color: color-mix(
    in srgb,
    var(--color-danger-on-strong) 12%,
    var(--color-surface-strong)
  );
  border-left-color: var(--color-danger-on-strong);
}

.customer-details :deep(.base-alert--warning) {
  background-color: color-mix(
    in srgb,
    var(--color-warning-on-strong) 12%,
    var(--color-surface-strong)
  );
  border-left-color: var(--color-warning-on-strong);
}

.customer-details :deep(.base-alert__title),
.customer-details :deep(.base-alert__text) {
  color: var(--pb-text);
}

.customer-details :deep(.base-alert__status) {
  color: var(--pb-brass);
}

.customer-details__conflict-line {
  margin: 0;
}

.customer-details__conflict-line + .customer-details__conflict-line,
.customer-details__alternatives {
  margin-top: var(--space-2);
}

.customer-details__alternatives {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  padding: 0;
  list-style: none;
}

.customer-details__actions {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--space-3);
}

.customer-details__actions :deep(.base-button) {
  width: 100%;
}

@media (min-width: 480px) {
  .customer-details__actions {
    grid-template-columns: 1fr 1.6fr;
  }
}

.customer-details__note {
  margin: 0;
  color: var(--pb-soft);
  text-align: center;
}

/* Sello de confirmación: el rombo se traza, el check se dibuja y ocho
   chispas de latón salen disparadas. Después queda quieto. */
.customer-details__seal {
  position: relative;
  display: block;
  align-self: center;
  width: 104px;
  height: 104px;
}

.customer-details__seal svg {
  width: 100%;
  height: 100%;
  overflow: visible;
  fill: none;
  stroke: var(--pb-brass);
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.customer-details__seal-diamond {
  transform: rotate(45deg);
  transform-origin: 48px 48px;
  stroke-dasharray: 1;
  animation: seal-draw 0.9s cubic-bezier(0.55, 0, 0.2, 1) 0.15s backwards;
}

.customer-details__seal-check {
  stroke: var(--color-success-on-strong);
  stroke-width: 3;
  stroke-dasharray: 1;
  animation: seal-draw 0.55s cubic-bezier(0.55, 0, 0.2, 1) 0.8s backwards;
}

.customer-details__spark {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 5px;
  height: 5px;
  margin: -2.5px 0 0 -2.5px;
  background-color: var(--pb-brass);
  opacity: 0;
  transform: rotate(var(--spark-angle)) translateY(0) rotate(45deg);
  animation: seal-spark 1s cubic-bezier(0.2, 0.7, 0.2, 1) var(--spark-delay, 0.8s) backwards;
}

@keyframes seal-draw {
  from {
    stroke-dashoffset: 1;
  }
  to {
    stroke-dashoffset: 0;
  }
}

@keyframes seal-spark {
  0% {
    opacity: 1;
    transform: rotate(var(--spark-angle)) translateY(0) rotate(45deg) scale(1);
  }
  100% {
    opacity: 0;
    transform: rotate(var(--spark-angle)) translateY(-74px) rotate(45deg) scale(0.4);
  }
}

@media (prefers-reduced-motion: reduce) {
  .customer-details__rise,
  .pb-sheet__form > *,
  .customer-details__summary-row,
  .customer-details__seal-diamond,
  .customer-details__seal-check,
  .customer-details__radio input:checked + span::before {
    animation: none;
  }

  .customer-details__spark {
    display: none;
  }

  .customer-details__radio span,
  .customer-details__textarea {
    transition: none;
  }

  .customer-details__radio:hover input:not(:checked) + span {
    transform: none;
  }
}
</style>
