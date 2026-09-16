<script setup lang="ts">
// Página coordinadora de la lectura del turno del cliente por token
// (HU-098, CA-098-01 a CA-098-06). Sin mockup asignado: composición libre
// dentro de NAVA / Tailored Grid (DEC-078). Posee estado y reintento; no
// conoce la forma RFC 9457 del contrato (eso queda dentro de
// `api/getCustomerAppointmentApi.ts`). Fuera de alcance a propósito:
// cuenta, portal, edición, reprogramación o cancelación (HU-099).
import { computed, onMounted, ref } from 'vue'
import { BaseAlert, BaseBadge, BaseButton, NavaWordmark, PageState } from '@/shared/ui'
import { formatInstantInTimezone, formatTimeInTimezone } from '@/shared/time/formatInstant'
import { getCustomerAppointment } from '../api/getCustomerAppointmentApi'
import {
  APPOINTMENT_STATUS_BADGE_VARIANT,
  APPOINTMENT_STATUS_LABELS,
  type CustomerAppointment,
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

async function load() {
  screenState.value = { status: 'loading' }
  const outcome = await getCustomerAppointment(props.token)
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

const retry = () => {
  void load()
}

const statusLabel = computed(() =>
  screenState.value.status === 'success'
    ? APPOINTMENT_STATUS_LABELS[screenState.value.appointment.status]
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
    return `Puedes cancelar sin costo hasta ${appointment.cancellationDeadlineMinutes} minutos antes de tu turno. Pasado ese plazo, contacta directamente a la barbería.`
  }
  const reasonNote = appointment.lateCancellationReasonRequired ? ' indicando el motivo' : ''
  return `Puedes cancelar hasta ${appointment.cancellationDeadlineMinutes} minutos antes sin costo, o después${reasonNote} conforme a la política de la barbería.`
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
          <dt>Barbero</dt>
          <dd>{{ screenState.appointment.barberName }}</dd>
        </div>
      </dl>

      <BaseAlert variant="plain" role="status">
        {{ cancellationPolicyLabel }}
      </BaseAlert>
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
