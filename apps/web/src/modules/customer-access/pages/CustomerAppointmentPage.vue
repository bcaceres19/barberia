<script setup lang="ts">
// Página coordinadora de la lectura del turno del cliente por token
// (HU-098, CA-098-01 a CA-098-06). Sin mockup asignado: composición libre
// dentro de NAVA / Tailored Grid (DEC-078). Posee estado y reintento; no
// conoce la forma RFC 9457 del contrato (eso queda dentro de
// `api/getCustomerAppointmentApi.ts`). Cancelar el turno es HU-099 (T5). Fuera de
// alcance a propósito: cuenta, portal, edición y reprogramación.
import { computed, nextTick, onMounted, ref } from 'vue'
import { BaseAlert, BaseBadge, BaseButton, NavaWordmark, PageState } from '@/shared/ui'
import { buildVocabularyFromTerms } from '@/shared/model'
import { formatInstantInTimezone, formatTimeInTimezone } from '@/shared/time/formatInstant'
import { getCustomerAppointment } from '../api/getCustomerAppointmentApi'
import { createReview } from '../api/createReviewApi'
import { cancelCustomerAppointment } from '../api/cancelCustomerAppointmentApi'
import AppointmentCancellation from '../components/AppointmentCancellation.vue'
import AppointmentReview from '../components/AppointmentReview.vue'
import { newIdempotencyKey } from '../model/idempotencyKey'
import {
  APPOINTMENT_STATUS_BADGE_VARIANT,
  appointmentStatusLabel,
  type CancelAppointmentInput,
  type CancellationRejection,
  type CustomerAppointment,
  type ReviewInput,
} from '../model/customerAppointmentOutcome'

interface Props {
  /** Credencial en claro del enlace, tal como llega del parámetro de ruta
   * `:token` (app/router, `props: true`). No confiada: el servidor la
   * hashea y la vuelve a resolver desde cero (CA-098-02); nunca se registra
   * ni se persiste en este componente más allá de reenviarla al API. */
  token: string
}
const props = defineProps<Props>()

type ScreenState =
  | { status: 'loading' }
  | { status: 'success'; appointment: CustomerAppointment }
  | { status: 'not-found' }
  | { status: 'network-error' }
  | { status: 'unexpected-error'; requestId?: string }

const screenState = ref<ScreenState>({ status: 'loading' })
// Reloj del dispositivo al cargar: solo alimenta la pista local de la política de
// cancelación; la decisión real es del servidor (HU-099).
const nowMs = ref(Date.now())

async function load() {
  screenState.value = { status: 'loading' }
  const outcome = await getCustomerAppointment(props.token)
  nowMs.value = Date.now()
  switch (outcome.kind) {
    case 'success':
      screenState.value = { status: 'success', appointment: outcome.appointment }
      return
    case 'not-found':
      screenState.value = { status: 'not-found' }
      return
    case 'network-error':
      screenState.value = { status: 'network-error' }
      return
    case 'unexpected-error':
      screenState.value = { status: 'unexpected-error', requestId: outcome.requestId }
  }
}

onMounted(load)

// Reseña del turno (HU-100, DEC-128): la página posee el envío; el componente solo
// presenta el formulario. Un `409` significa que el turno ya no admite reseña (la
// calificó otra pestaña, o cambió de estado): se recarga para mostrar lo vigente.
const reviewSubmitting = ref(false)
const reviewError = ref('')

async function submitReview(input: ReviewInput) {
  if (screenState.value.status !== 'success' || reviewSubmitting.value) return
  reviewSubmitting.value = true
  reviewError.value = ''
  const outcome = await createReview(props.token, input)
  reviewSubmitting.value = false

  switch (outcome.kind) {
    case 'success':
      screenState.value = {
        status: 'success',
        appointment: { ...screenState.value.appointment, canReview: false, review: outcome.review },
      }
      return
    case 'not-reviewable':
    case 'not-found':
      reviewError.value = 'Este turno ya no admite una reseña.'
      await load()
      return
    case 'validation-error':
      reviewError.value = 'Revisa la calificación y el comentario e inténtalo de nuevo.'
      return
    case 'network-error':
      reviewError.value = 'No pudimos conectar. Revisa tu conexión e inténtalo de nuevo.'
      return
    case 'unexpected-error':
      reviewError.value = outcome.requestId
        ? `No pudimos guardar tu reseña. Si continúa, comparte este código con soporte: ${outcome.requestId}.`
        : 'No pudimos guardar tu reseña. Inténtalo de nuevo en unos segundos.'
  }
}

// Cancelación del turno (HU-099, T5): la página posee el envío y la clave de
// idempotencia. Se reutiliza la clave solo al reintentar la MISMA intención (mismo
// motivo) tras un fallo de red o una respuesta inesperada; cualquier otra cosa es un
// intento nuevo con clave nueva (RN-IDE-01, DEC-043).
const cancelSubmitting = ref(false)
const cancelError = ref('')
const cancelRejection = ref<CancellationRejection>(null)
const cancelNotice = ref('')
const justCancelled = ref(false)
const cancelledNotice = ref<HTMLElement | null>(null)
let lastCancelAttempt: { reason: string | null; key: string } | null = null

function cancelAttemptKey(reason: string | null): string {
  if (lastCancelAttempt && lastCancelAttempt.reason === reason) return lastCancelAttempt.key
  lastCancelAttempt = { reason, key: newIdempotencyKey() }
  return lastCancelAttempt.key
}

async function submitCancel(input: CancelAppointmentInput) {
  if (screenState.value.status !== 'success' || cancelSubmitting.value) return
  const { appointment } = screenState.value
  cancelSubmitting.value = true
  cancelError.value = ''
  cancelRejection.value = null
  const outcome = await cancelCustomerAppointment(
    props.token,
    appointment.versionToken,
    cancelAttemptKey(input.reason),
    input,
  )
  cancelSubmitting.value = false

  switch (outcome.kind) {
    case 'success':
      lastCancelAttempt = null
      justCancelled.value = true
      screenState.value = {
        status: 'success',
        appointment: { ...appointment, status: 'cancelled_by_customer', canReview: false },
      }
      await nextTick()
      cancelledNotice.value?.focus()
      return
    case 'window-closed':
    case 'reason-required':
      lastCancelAttempt = null
      cancelRejection.value = outcome.kind
      return
    case 'validation-error':
      cancelError.value = 'Revisa el motivo e inténtalo de nuevo.'
      return
    case 'changed':
      lastCancelAttempt = null
      cancelNotice.value =
        'Tu turno cambió mientras tanto, así que no lo cancelamos. Esto es lo que figura ahora.'
      await load()
      return
    case 'idempotency-in-progress':
      cancelError.value =
        'Tu cancelación sigue en curso. Espera unos segundos e inténtalo de nuevo.'
      return
    case 'not-found':
      lastCancelAttempt = null
      screenState.value = { status: 'not-found' }
      return
    case 'network-error':
      cancelError.value = 'No pudimos conectar. Revisa tu conexión e inténtalo de nuevo.'
      return
    case 'unexpected-error':
      cancelError.value = outcome.requestId
        ? `No pudimos cancelar tu turno. Si continúa, comparte este código con soporte: ${outcome.requestId}.`
        : 'No pudimos cancelar tu turno. Inténtalo de nuevo en unos segundos.'
  }
}

const retry = () => {
  void load()
}

// Vocabulario de la barbería del turno (DEC-119): hasta que el servidor responde, o si
// no responde, valen los valores iniciales.
const v = computed(() =>
  buildVocabularyFromTerms(
    screenState.value.status === 'success' ? screenState.value.appointment.vocabulary : null,
  ),
)

// DEC-136: el enlace sigue vivo tras cancelar y muestra el turno cancelado, sin
// política de cancelación ni acciones.
const isCancelled = computed(
  () =>
    screenState.value.status === 'success' &&
    screenState.value.appointment.status.startsWith('cancelled'),
)

const statusLabel = computed(() =>
  screenState.value.status === 'success'
    ? appointmentStatusLabel(screenState.value.appointment.status, v.value)
    : '',
)
const statusBadgeVariant = computed(() =>
  screenState.value.status === 'success'
    ? APPOINTMENT_STATUS_BADGE_VARIANT[screenState.value.appointment.status]
    : 'neutral',
)

// CA-098-04: hora y política se muestran en la zona de la barbería, nunca
// en la del dispositivo (mismo criterio que formatInstantInTimezone en
// public-booking/pages/PublicBarbershopEntryPage.vue).
const timeRangeLabel = computed(() => {
  if (screenState.value.status !== 'success') return ''
  const { appointment } = screenState.value
  const start = formatInstantInTimezone(appointment.startsAt, appointment.timezone)
  const end = formatTimeInTimezone(appointment.endsAt, appointment.timezone)
  return `${start} – ${end}`
})

// cancellationPolicyLabel resume la política VIGENTE (HU-093) sin afirmar
// una acción que HU-098 no ofrece todavía (la cancelación es HU-099):
// nunca implica que la falta de respuesta cancele el turno (RN-CNF-02).
const cancellationPolicyLabel = computed(() => {
  if (screenState.value.status !== 'success') return ''
  const { appointment } = screenState.value
  if (!appointment.lateCancellationClientAllowed) {
    return `Puedes cancelar sin costo hasta ${appointment.cancellationDeadlineMinutes} minutos antes de tu turno. Pasado ese plazo, contacta directamente ${v.value.toTheBusiness}.`
  }
  const reasonNote = appointment.lateCancellationReasonRequired ? ' indicando el motivo' : ''
  return `Puedes cancelar hasta ${appointment.cancellationDeadlineMinutes} minutos antes sin costo, o después${reasonNote} conforme a la política ${v.value.ofTheBusiness}.`
})

const unexpectedErrorMessage = computed(() => {
  const state = screenState.value
  if (state.status !== 'unexpected-error') return ''
  return state.requestId
    ? `Inténtalo de nuevo. Si continúa, comparte este código con soporte: ${state.requestId}.`
    : 'Inténtalo de nuevo en unos segundos.'
})
</script>

<template>
  <main
    v-if="screenState.status !== 'success'"
    class="customer-appointment customer-appointment--state"
  >
    <!-- Encabezado accesible pero visualmente oculto: PageState solo aporta
         un <h2>; esta pantalla pública no tiene ningún otro <h1>, así que
         lo provee aquí (page-has-heading-one, axe-core). -->
    <h1 class="visually-hidden">Tu turno</h1>
    <PageState
      v-if="screenState.status === 'loading'"
      variant="loading"
      headline="Buscando tu turno…"
      role="status"
    />
    <PageState
      v-else-if="screenState.status === 'not-found'"
      variant="danger"
      status-label="Error"
      headline="No pudimos abrir ese enlace"
      role="alert"
    >
      <template #default
        >Puede haber vencido, haberse usado para cancelar, o estar escrito de forma incompleta.
        Revisa el correo de confirmación y vuelve a intentarlo.</template
      >
      <template #action>
        <BaseButton variant="secondary" @click="retry">Reintentar</BaseButton>
      </template>
    </PageState>
    <PageState
      v-else-if="screenState.status === 'network-error'"
      variant="warning"
      status-label="Atención"
      headline="No pudimos conectar"
      role="alert"
    >
      <template #default>Revisa tu conexión e inténtalo de nuevo.</template>
      <template #action>
        <BaseButton variant="primary" @click="retry">Reintentar</BaseButton>
      </template>
    </PageState>
    <PageState
      v-else
      variant="danger"
      status-label="Error"
      headline="Ocurrió un error inesperado"
      role="alert"
    >
      <template #default>{{ unexpectedErrorMessage }}</template>
      <template #action>
        <BaseButton variant="primary" @click="retry">Reintentar</BaseButton>
      </template>
    </PageState>
  </main>

  <main v-else class="customer-appointment customer-appointment--ready">
    <div class="customer-appointment__card">
      <NavaWordmark variant="ink" size="sm" />

      <header class="customer-appointment__header">
        <h1 class="customer-appointment__title">{{ screenState.appointment.barbershopName }}</h1>
        <p class="customer-appointment__status" role="status">
          <BaseBadge :variant="statusBadgeVariant" size="sm" :label="statusLabel">
            {{ statusLabel }}
          </BaseBadge>
        </p>
      </header>

      <dl class="customer-appointment__facts">
        <div class="customer-appointment__fact">
          <dt>Hora</dt>
          <dd>{{ timeRangeLabel }} · Zona {{ screenState.appointment.timezone }}</dd>
        </div>
        <div class="customer-appointment__fact">
          <dt>Persona atendida</dt>
          <dd>{{ screenState.appointment.attendeeName }}</dd>
        </div>
        <div class="customer-appointment__fact">
          <dt>Servicio</dt>
          <dd>
            {{ screenState.appointment.serviceName }} ·
            {{ screenState.appointment.durationMinutes }} min
          </dd>
        </div>
        <div class="customer-appointment__fact">
          <dt>{{ v.Professional }}</dt>
          <dd>{{ screenState.appointment.barberName }}</dd>
        </div>
      </dl>

      <BaseAlert v-if="cancelNotice" variant="warning" role="status">{{ cancelNotice }}</BaseAlert>

      <div
        v-if="justCancelled"
        ref="cancelledNotice"
        class="customer-appointment__cancelled"
        tabindex="-1"
      >
        <BaseAlert variant="success" role="status">
          Cancelaste tu turno. El horario quedó libre.
        </BaseAlert>
      </div>
      <BaseAlert v-else-if="isCancelled" variant="warning" role="status">
        Este turno está cancelado. El horario quedó libre.
      </BaseAlert>
      <BaseAlert v-else variant="plain" role="status">
        {{ cancellationPolicyLabel }}
      </BaseAlert>

      <AppointmentCancellation
        v-if="screenState.appointment.status === 'confirmed'"
        :appointment="screenState.appointment"
        :vocabulary="v"
        :now-ms="nowMs"
        :submitting="cancelSubmitting"
        :rejection="cancelRejection"
        :error-message="cancelError"
        @submit="submitCancel"
      />

      <AppointmentReview
        :can-review="screenState.appointment.canReview"
        :review="screenState.appointment.review"
        :professional-name="screenState.appointment.barberName"
        :submitting="reviewSubmitting"
        :error-message="reviewError"
        @submit="submitReview"
      />
    </div>
  </main>
</template>

<style scoped>
.customer-appointment {
  display: flex;
  min-height: 100dvh;
  align-items: center;
  justify-content: center;
  padding: var(--space-6) var(--space-4);
}

.customer-appointment--state {
  background-color: var(--color-surface-strong);
}

.customer-appointment--ready {
  background-color: var(--color-canvas);
}

.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  margin: -1px;
  padding: 0;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.customer-appointment__card {
  display: flex;
  width: 100%;
  max-width: 480px;
  flex-direction: column;
  align-items: center;
  gap: var(--space-5);
  text-align: center;
}

.customer-appointment__cancelled {
  width: 100%;
  text-align: left;
}

.customer-appointment__cancelled:focus-visible {
  outline: var(--focus-ring-width, 2px) solid var(--color-focus);
  outline-offset: 2px;
}

.customer-appointment__header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
}

.customer-appointment__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: 32px;
  line-height: 40px;
  letter-spacing: -0.015em;
  color: var(--color-text-primary);
}

.customer-appointment__status {
  display: inline-flex;
  margin: 0;
}

.customer-appointment__facts {
  display: flex;
  width: 100%;
  flex-direction: column;
  gap: 0;
  margin: 0;
  text-align: left;
  border-top: var(--border-width-normal) solid var(--color-border-subtle);
}

.customer-appointment__fact {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: var(--space-3) 0;
  border-bottom: var(--border-width-normal) solid var(--color-border-subtle);
}

.customer-appointment__fact dt {
  font-family: var(--font-family-base);
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.customer-appointment__fact dd {
  margin: 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-text-primary);
}

@media (min-width: 1024px) {
  .customer-appointment__title {
    font-size: 40px;
    line-height: 48px;
  }
}
</style>
