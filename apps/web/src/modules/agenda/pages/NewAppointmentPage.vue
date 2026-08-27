<script setup lang="ts">
// Pantalla "Nuevo turno" (HU-061): el barbero autenticado registra un
// turno manual recibido por teléfono, WhatsApp o en persona. No consulta
// disponibilidad pública ni propone franjas (B4 es dueña); no edita,
// reprograma, cancela ni cambia estado (fuera de alcance de HU-061). El
// éxito muestra un resumen honesto: HU-062 (agenda diaria) todavía no
// existe, así que esta pantalla no enlaza a ninguna vista de agenda real.
import { computed, ref, watch } from 'vue'
import { BaseAlert, BaseButton, BaseInput } from '@/shared/ui'
import {
  createManualAppointment,
  fetchAssignedServices,
  fetchBarberSummaries,
  fetchBarbershopTimezone,
} from '../api/appointmentsApi'
import { newIdempotencyKey } from '../model/idempotencyKey'
import type { BarberSummary, ServiceSummary } from '../model/appointment'
import type { CreatedManualAppointment } from '../model/appointmentOutcome'
import {
  buildStartsAt,
  validateAttendeeName,
  validateCustomerEmail,
  validateCustomerFullName,
  validateCustomerNote,
  validateCustomerPhone,
  validateStartsAt,
} from '../validation/appointmentValidation'

type PageStatus = 'loading' | 'ready' | 'load-error'
type ServicesStatus = 'idle' | 'loading' | 'ready' | 'error'
type SaveStatus =
  | 'idle'
  | 'saving'
  | 'validation-error'
  | 'conflict'
  | 'idempotency-conflict'
  | 'not-found'
  | 'network-error'
  | 'unexpected-error'

const pageStatus = ref<PageStatus>('loading')
const barbers = ref<BarberSummary[]>([])
const barbershopTimezone = ref<string | null>(null)

const selectedBarberId = ref('')
const servicesStatus = ref<ServicesStatus>('idle')
const services = ref<ServiceSummary[]>([])
const selectedServiceId = ref('')

const attendeeName = ref('')
const customerFullName = ref('')
const customerPhone = ref('')
const customerEmail = ref('')
const customerNote = ref('')
const startsAtDate = ref('')
const startsAtTime = ref('')

const fieldErrors = ref<Record<string, string | undefined>>({})
const attempted = ref(false)
const saveStatus = ref<SaveStatus>('idle')
const saveErrorDetail = ref<string | undefined>(undefined)
const created = ref<CreatedManualAppointment | null>(null)
let idempotencyKey = newIdempotencyKey()

async function loadPage() {
  pageStatus.value = 'loading'
  const [barbersOutcome, timezoneOutcome] = await Promise.all([
    fetchBarberSummaries(),
    fetchBarbershopTimezone(),
  ])
  barbershopTimezone.value = timezoneOutcome.kind === 'success' ? timezoneOutcome.timezone : null
  if (barbersOutcome.kind !== 'success') {
    pageStatus.value = 'load-error'
    return
  }
  barbers.value = barbersOutcome.items
  pageStatus.value = 'ready'
}

void loadPage()

watch(selectedBarberId, async (barberId) => {
  selectedServiceId.value = ''
  services.value = []
  if (!barberId) {
    servicesStatus.value = 'idle'
    return
  }
  servicesStatus.value = 'loading'
  const outcome = await fetchAssignedServices(barberId)
  if (selectedBarberId.value !== barberId) return
  if (outcome.kind === 'success') {
    services.value = outcome.items
    servicesStatus.value = 'ready'
  } else {
    servicesStatus.value = 'error'
  }
})

const startsAt = computed(() => buildStartsAt(startsAtDate.value, startsAtTime.value))

function validateAll(): boolean {
  const errors: Record<string, string | undefined> = {
    barberId: selectedBarberId.value ? undefined : 'Elige un barbero.',
    serviceId: selectedServiceId.value ? undefined : 'Elige un servicio.',
    attendeeName: validateAttendeeName(attendeeName.value),
    customerFullName: validateCustomerFullName(customerFullName.value),
    customerPhone: validateCustomerPhone(customerPhone.value),
    customerEmail: validateCustomerEmail(customerEmail.value),
    customerNote: validateCustomerNote(customerNote.value),
    startsAt: validateStartsAt(startsAt.value),
  }
  fieldErrors.value = errors
  return Object.values(errors).every((e) => e === undefined)
}

async function onSubmit() {
  if (saveStatus.value === 'saving') return
  attempted.value = true
  saveErrorDetail.value = undefined
  if (!validateAll()) {
    saveStatus.value = 'validation-error'
    return
  }

  saveStatus.value = 'saving'
  const outcome = await createManualAppointment(
    {
      barberId: selectedBarberId.value,
      serviceId: selectedServiceId.value,
      attendeeName: attendeeName.value.trim(),
      customerFullName: customerFullName.value.trim(),
      customerPhone: customerPhone.value.trim() === '' ? null : customerPhone.value.trim(),
      customerEmail: customerEmail.value.trim() === '' ? null : customerEmail.value.trim(),
      customerNote: customerNote.value.trim() === '' ? null : customerNote.value.trim(),
      startsAt: startsAt.value,
    },
    idempotencyKey,
  )

  switch (outcome.kind) {
    case 'success':
      created.value = outcome.appointment
      saveStatus.value = 'idle'
      return
    case 'conflict':
      saveErrorDetail.value = outcome.detail
      saveStatus.value = 'conflict'
      return
    case 'validation-error':
      saveErrorDetail.value = outcome.detail
      saveStatus.value = 'validation-error'
      return
    case 'idempotency-conflict':
      saveStatus.value = 'idempotency-conflict'
      return
    case 'not-found':
      saveStatus.value = 'not-found'
      return
    case 'network-error':
      saveStatus.value = 'network-error'
      return
    case 'unexpected-error':
      saveStatus.value = 'unexpected-error'
  }
}

function onStartNewAppointment() {
  created.value = null
  attendeeName.value = ''
  customerFullName.value = ''
  customerPhone.value = ''
  customerEmail.value = ''
  customerNote.value = ''
  startsAtDate.value = ''
  startsAtTime.value = ''
  fieldErrors.value = {}
  attempted.value = false
  saveStatus.value = 'idle'
  saveErrorDetail.value = undefined
  idempotencyKey = newIdempotencyKey()
}

function onRetryLoad() {
  void loadPage()
}
</script>

<template>
  <section class="new-appointment-page" aria-labelledby="new-appointment-title">
    <header class="new-appointment-page__header">
      <h1 id="new-appointment-title" class="new-appointment-page__title">Nuevo turno</h1>
    </header>

    <p v-if="barbershopTimezone" class="new-appointment-page__timezone">
      Horas en la zona horaria de la barbería: {{ barbershopTimezone }}
    </p>

    <div
      v-if="pageStatus === 'loading'"
      class="new-appointment-page__state"
      role="status"
      aria-live="polite"
    >
      <p>Cargando barberos…</p>
    </div>

    <BaseAlert
      v-else-if="pageStatus === 'load-error'"
      variant="warning"
      title="No pudimos cargar esta sección"
      role="alert"
    >
      Revisa tu conexión e inténtalo de nuevo.
      <template #action>
        <BaseButton type="button" variant="secondary" @click="onRetryLoad">Reintentar</BaseButton>
      </template>
    </BaseAlert>

    <template v-else>
      <p v-if="barbers.length === 0" class="new-appointment-page__empty">
        Aún no tienes barberos registrados. Agrega uno en la sección "Barberos" antes de registrar
        turnos.
      </p>

      <div
        v-else-if="created"
        class="new-appointment-page__summary"
        role="status"
        aria-live="polite"
      >
        <BaseAlert variant="success" title="Turno registrado">
          {{ created.attendeeName }} · {{ created.serviceName }} ({{ created.durationMinutes }} min)
          · {{ created.priceAmount }} {{ created.currency }}
        </BaseAlert>
        <p class="new-appointment-page__summary-note">
          El turno quedó confirmado en la agenda del barbero. Las notificaciones al cliente todavía
          no están disponibles (B5 las agrega más adelante).
        </p>
        <BaseButton type="button" variant="primary" @click="onStartNewAppointment">
          Registrar otro turno
        </BaseButton>
      </div>

      <form v-else class="new-appointment-page__form" novalidate @submit.prevent="onSubmit">
        <div class="new-appointment-page__field">
          <label for="new-appointment-barber" class="new-appointment-page__label">Barbero</label>
          <select
            id="new-appointment-barber"
            v-model="selectedBarberId"
            class="new-appointment-page__select"
          >
            <option value="" disabled>Elige un barbero</option>
            <option v-for="b in barbers" :key="b.id" :value="b.id">{{ b.fullName }}</option>
          </select>
          <p
            v-if="attempted && fieldErrors.barberId"
            class="new-appointment-page__error"
            role="alert"
          >
            {{ fieldErrors.barberId }}
          </p>
        </div>

        <div class="new-appointment-page__field">
          <label for="new-appointment-service" class="new-appointment-page__label">Servicio</label>
          <select
            id="new-appointment-service"
            v-model="selectedServiceId"
            class="new-appointment-page__select"
            :disabled="!selectedBarberId || servicesStatus === 'loading'"
          >
            <option value="" disabled>
              {{ selectedBarberId ? 'Elige un servicio' : 'Elige primero un barbero' }}
            </option>
            <option v-for="s in services" :key="s.id" :value="s.id">{{ s.name }}</option>
          </select>
          <p v-if="servicesStatus === 'loading'" role="status" aria-live="polite">
            Cargando servicios…
          </p>
          <p v-else-if="servicesStatus === 'ready' && services.length === 0">
            Este barbero no tiene servicios activos asignados.
          </p>
          <p v-else-if="servicesStatus === 'error'" role="alert">
            No pudimos cargar los servicios de este barbero.
          </p>
          <p
            v-if="attempted && fieldErrors.serviceId"
            class="new-appointment-page__error"
            role="alert"
          >
            {{ fieldErrors.serviceId }}
          </p>
        </div>

        <BaseInput
          v-model="attendeeName"
          type="text"
          label="Persona atendida"
          required
          :error="attempted ? fieldErrors.attendeeName : undefined"
        />
        <BaseInput
          v-model="customerFullName"
          type="text"
          label="Nombre del cliente"
          required
          :error="attempted ? fieldErrors.customerFullName : undefined"
        />
        <div class="new-appointment-page__form-row">
          <BaseInput
            v-model="customerPhone"
            type="tel"
            label="Teléfono (opcional)"
            placeholder="+573001234567"
            :error="attempted ? fieldErrors.customerPhone : undefined"
          />
          <BaseInput
            v-model="customerEmail"
            type="email"
            label="Correo (opcional)"
            :error="attempted ? fieldErrors.customerEmail : undefined"
          />
        </div>
        <p v-if="!customerPhone && !customerEmail" class="new-appointment-page__hint">
          Sin teléfono ni correo, el cliente no recibirá recordatorios.
        </p>

        <div class="new-appointment-page__form-row">
          <BaseInput v-model="startsAtDate" type="date" label="Fecha del turno" required />
          <BaseInput v-model="startsAtTime" type="time" label="Hora del turno" required />
        </div>
        <p
          v-if="attempted && fieldErrors.startsAt"
          class="new-appointment-page__error"
          role="alert"
        >
          {{ fieldErrors.startsAt }}
        </p>

        <BaseInput
          v-model="customerNote"
          type="text"
          label="Nota (opcional)"
          :error="attempted ? fieldErrors.customerNote : undefined"
        />

        <BaseAlert v-if="saveStatus === 'conflict'" variant="danger" role="alert">
          {{ saveErrorDetail }}
        </BaseAlert>
        <BaseAlert v-else-if="saveStatus === 'validation-error'" variant="danger" role="alert">
          {{ saveErrorDetail ?? 'Revisa los datos del turno.' }}
        </BaseAlert>
        <BaseAlert v-else-if="saveStatus === 'idempotency-conflict'" variant="danger" role="alert">
          Este intento ya estaba en curso o cambió mientras se procesaba. Recarga la página e
          inténtalo de nuevo.
        </BaseAlert>
        <BaseAlert v-else-if="saveStatus === 'not-found'" variant="danger" role="alert">
          El barbero o el servicio elegidos ya no están disponibles.
        </BaseAlert>
        <BaseAlert
          v-else-if="saveStatus === 'network-error' || saveStatus === 'unexpected-error'"
          variant="danger"
          role="alert"
        >
          No pudimos registrar el turno. Tus datos se conservaron; inténtalo de nuevo.
        </BaseAlert>

        <BaseButton type="submit" variant="primary" :disabled="saveStatus === 'saving'">
          {{ saveStatus === 'saving' ? 'Guardando…' : 'Registrar turno' }}
        </BaseButton>
      </form>
    </template>
  </section>
</template>

<style scoped>
.new-appointment-page__form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-width: 32rem;
}

.new-appointment-page__field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.new-appointment-page__label {
  font-weight: 600;
}

.new-appointment-page__select {
  min-height: 44px;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--color-border, #ccc);
  border-radius: 0.375rem;
}

.new-appointment-page__form-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.new-appointment-page__form-row > * {
  flex: 1 1 12rem;
}

.new-appointment-page__error {
  color: var(--color-danger, #b00020);
  margin: 0;
}

.new-appointment-page__hint {
  color: var(--color-text-secondary, #666);
  margin: 0;
}

.new-appointment-page__summary {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  max-width: 32rem;
}
</style>
