<script setup lang="ts">
// Pantalla "Nuevo turno" (HU-061): el barbero autenticado registra un
// turno manual recibido por teléfono, WhatsApp o en persona. No consulta
// disponibilidad pública ni propone franjas (B4 es dueña); no edita,
// reprograma, cancela ni cambia estado (fuera de alcance de HU-061). El
// éxito muestra un resumen honesto: HU-062 (agenda diaria) todavía no
// existe, así que esta pantalla no enlaza a ninguna vista de agenda real.
import { computed, ref, watch } from 'vue'
import { BaseAlert, BaseButton, BaseInput } from '@/shared/ui'
import { formatCivilDateFull, isCivilDateString } from '@/shared/time/civilDate'
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

// Resumen antes del CTA (estandar-diseno-visual.md §10, "Crear/editar
// turno": secciones cortas en orden, resumen; especificacion-frontend-nava.md
// §7.3: "resumen antes del CTA" en móvil). Solo refleja selecciones ya
// hechas, sin inventar datos que el contrato de servicios no expone
// todavía (ServiceSummary hoy solo trae id/name, sin duración ni precio).
const selectedBarberName = computed(
  () => barbers.value.find((b) => b.id === selectedBarberId.value)?.fullName ?? '',
)
const selectedServiceName = computed(
  () => services.value.find((s) => s.id === selectedServiceId.value)?.name ?? '',
)
const summaryDateLabel = computed(() =>
  isCivilDateString(startsAtDate.value) ? formatCivilDateFull(startsAtDate.value) : '',
)
const hasSummaryContent = computed(
  () =>
    !!(
      selectedBarberName.value ||
      selectedServiceName.value ||
      attendeeName.value.trim() ||
      summaryDateLabel.value ||
      startsAtTime.value
    ),
)

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
        <div class="new-appointment-page__section">
          <div class="new-appointment-page__section-heading">
            <span class="new-appointment-page__section-number" aria-hidden="true">1</span>
            <div>
              <h2 class="new-appointment-page__section-title">Selecciona</h2>
              <p class="new-appointment-page__section-hint">Elige al barbero y el servicio.</p>
            </div>
          </div>

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
            <label for="new-appointment-service" class="new-appointment-page__label"
              >Servicio</label
            >
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
        </div>

        <div class="new-appointment-page__section">
          <div class="new-appointment-page__section-heading">
            <span class="new-appointment-page__section-number" aria-hidden="true">2</span>
            <div>
              <h2 class="new-appointment-page__section-title">Persona atendida</h2>
              <p class="new-appointment-page__section-hint">Indica quién recibirá el servicio.</p>
            </div>
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
        </div>

        <div class="new-appointment-page__section">
          <div class="new-appointment-page__section-heading">
            <span class="new-appointment-page__section-number" aria-hidden="true">3</span>
            <div>
              <h2 class="new-appointment-page__section-title">Fecha y hora</h2>
              <p class="new-appointment-page__section-hint">Define cuándo será el turno.</p>
            </div>
          </div>

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
        </div>

        <div class="new-appointment-page__section">
          <div class="new-appointment-page__section-heading">
            <span class="new-appointment-page__section-number" aria-hidden="true">4</span>
            <div>
              <h2 class="new-appointment-page__section-title">Nota (opcional)</h2>
              <p class="new-appointment-page__section-hint">
                Agrega información adicional si es necesario.
              </p>
            </div>
          </div>

          <BaseInput
            v-model="customerNote"
            type="text"
            label="Nota (opcional)"
            :error="attempted ? fieldErrors.customerNote : undefined"
          />
        </div>

        <div
          v-if="hasSummaryContent"
          class="new-appointment-page__resumen"
          aria-labelledby="new-appointment-resumen-title"
        >
          <h2 id="new-appointment-resumen-title" class="new-appointment-page__resumen-title">
            Resumen
          </h2>
          <dl class="new-appointment-page__resumen-list">
            <div v-if="selectedBarberName" class="new-appointment-page__resumen-item">
              <dt>Barbero</dt>
              <dd>{{ selectedBarberName }}</dd>
            </div>
            <div v-if="selectedServiceName" class="new-appointment-page__resumen-item">
              <dt>Servicio</dt>
              <dd>{{ selectedServiceName }}</dd>
            </div>
            <div v-if="attendeeName.trim()" class="new-appointment-page__resumen-item">
              <dt>Persona atendida</dt>
              <dd>{{ attendeeName }}</dd>
            </div>
            <div v-if="summaryDateLabel || startsAtTime" class="new-appointment-page__resumen-item">
              <dt>Fecha y hora</dt>
              <dd>
                <span v-if="summaryDateLabel">{{ summaryDateLabel }}</span>
                <span v-if="summaryDateLabel && startsAtTime"> · </span>
                <span v-if="startsAtTime">{{ startsAtTime }}</span>
              </dd>
            </div>
          </dl>
        </div>

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

        <BaseButton
          type="submit"
          variant="primary"
          size="lg"
          class="new-appointment-page__submit"
          :disabled="saveStatus === 'saving'"
        >
          {{ saveStatus === 'saving' ? 'Guardando…' : 'Registrar turno' }}
        </BaseButton>
      </form>
    </template>
  </section>
</template>

<style scoped>
.new-appointment-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  max-width: 720px;
  padding: var(--space-4);
  margin: 0 auto;
}

.new-appointment-page__header {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
}

.new-appointment-page__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h1);
  line-height: var(--font-size-h1-line);
  font-weight: var(--font-weight-h1);
  color: var(--color-text-primary);
}

.new-appointment-page__timezone {
  margin: 0;
  font-size: var(--font-size-body-sm);
  color: var(--color-text-secondary);
}

.new-appointment-page__state {
  padding: var(--space-4);
  color: var(--color-text-secondary);
}

.new-appointment-page__empty {
  padding: var(--space-4);
  color: var(--color-text-secondary);
}

.new-appointment-page__form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  /* Formulario legible (estandar-diseno-visual.md §7.1): máximo 640px. */
  max-width: 640px;
}

.new-appointment-page__section {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding-bottom: var(--space-5);
  border-bottom: var(--border-width-normal) solid var(--color-border-subtle);
}

.new-appointment-page__section:last-of-type {
  padding-bottom: 0;
  border-bottom: none;
}

.new-appointment-page__section-heading {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
}

.new-appointment-page__section-number {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: var(--border-width-normal) solid var(--color-brand-accent-surface);
  border-radius: 999px;
  color: var(--color-brand-accent-text);
  font-family: var(--font-display);
  font-size: var(--font-size-body);
}

.new-appointment-page__section-title {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h3);
  line-height: var(--font-size-h3-line);
  color: var(--color-text-primary);
}

.new-appointment-page__section-hint {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-body-sm);
  color: var(--color-text-secondary);
}

.new-appointment-page__field {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.new-appointment-page__label {
  font-weight: 600;
  color: var(--color-text-primary);
}

.new-appointment-page__select {
  min-height: var(--control-height);
  padding: var(--space-2) var(--space-3);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-text-primary);
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-control);
  border-radius: var(--radius-sm);
}

.new-appointment-page__form-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.new-appointment-page__form-row > * {
  flex: 1 1 12rem;
}

.new-appointment-page__error {
  margin: 0;
  color: var(--color-danger-text);
}

.new-appointment-page__hint {
  margin: 0;
  color: var(--color-text-secondary);
}

.new-appointment-page__summary {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  max-width: 640px;
}

/* Resumen antes del CTA (§7.3): superficie contenida, línea y espacio, sin
   sombra ni flotación (estandar-diseno-visual.md §6.3). */
.new-appointment-page__resumen {
  padding: var(--space-4);
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: var(--radius-md);
}

.new-appointment-page__resumen-title {
  margin: 0 0 var(--space-3);
  font-size: var(--font-size-body-sm);
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.new-appointment-page__resumen-list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2) var(--space-5);
  margin: 0;
}

.new-appointment-page__resumen-item {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.new-appointment-page__resumen-item dt {
  color: var(--color-text-secondary);
}

.new-appointment-page__resumen-item dt::after {
  content: ':';
}

.new-appointment-page__resumen-item dd {
  margin: 0;
  font-weight: 600;
  color: var(--color-text-primary);
}

.new-appointment-page__submit {
  width: 100%;
}
</style>
