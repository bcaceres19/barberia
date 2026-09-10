<script setup lang="ts">
// Pantalla "Nuevo turno" (HU-061): el barbero autenticado registra un
// turno manual recibido por teléfono, WhatsApp o en persona. No consulta
// disponibilidad pública ni propone franjas (B4 es dueña); no edita,
// reprograma, cancela ni cambia estado (fuera de alcance de HU-061). El
// éxito muestra un resumen honesto: HU-062 (agenda diaria) no es destino de
// este enlace, así que esta pantalla no enlaza a ninguna vista de agenda.
//
// Fidelidad visual con el atlas nuevo-turno-eventos (issue #190, segundo
// pase del 2026-09-04): esta revisión es exclusivamente de composición,
// color, jerarquía y estado visual, nunca de comportamiento. Los comentarios
// que citan un evento del atlas (docs/10-backlog/evidence/
// ui-mockups-nava-tailored-grid-2026-09-03/nuevo-turno-eventos/README.md)
// documentan a qué panel responde cada bloque.
import { computed, nextTick, ref, watch } from 'vue'
import { BaseAlert, BaseButton, BaseInput, BarberAvatar, PageState } from '@/shared/ui'
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
import BarberSelect from '../components/BarberSelect.vue'
import NewAppointmentSkeleton from '../components/NewAppointmentSkeleton.vue'
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

// Tope real de `validateCustomerNote`: el contador de la nota refleja ese
// límite, no uno inventado (atlas, "Contrato visual común").
const NOTE_MAX_LENGTH = 500

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
const noteRef = ref<HTMLTextAreaElement | null>(null)

// La nota crece con su contenido en vez de dejar que la persona la
// arrastre a mano (`resize: none` en el estilo): el alto sigue al
// `scrollHeight` real del contenido en cada cambio, incluido el reinicio a
// '' de `onStartNewAppointment`.
function autoGrowNote() {
  const el = noteRef.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = `${el.scrollHeight}px`
}

watch(customerNote, () => {
  void nextTick(autoGrowNote)
})

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

// El campo de servicio depende del barbero (DEC-072): sin barbero elegido,
// mientras llegan sus servicios, o sin servicios activos que ofrecer, queda
// deshabilitado y pierde el filete de latón (atlas, eventos 04 y 05).
const isServiceDisabled = computed(
  () =>
    !selectedBarberId.value ||
    servicesStatus.value === 'loading' ||
    (servicesStatus.value !== 'idle' && services.value.length === 0),
)

// Rótulo de la opción vacía del `<select>`: sin esto, un barbero sin
// servicios activos (o cuya carga falló) mostraba "Elige un servicio" como
// única opción, prometiendo una lista que no existe.
const serviceSelectPlaceholder = computed(() => {
  if (!selectedBarberId.value) return 'Elige primero un barbero'
  if (servicesStatus.value === 'loading') return 'Cargando servicios…'
  if (servicesStatus.value === 'error') return 'No pudimos cargar los servicios'
  if (services.value.length === 0) return 'Sin servicios activos asignados'
  return 'Elige un servicio'
})

// Alerta global. En escritorio encabeza la columna lateral y en móvil cae
// justo encima del CTA: es el mismo nodo, la retícula lo recoloca.
const hasGlobalAlert = computed(() => saveStatus.value !== 'idle' && saveStatus.value !== 'saving')
// La columna lateral existe cuando lleva algo: alerta global, resumen o
// ambos. El evento 04 (formulario recién abierto) es el único sin ella.
const hasRail = computed(() => hasGlobalAlert.value || hasSummaryContent.value)

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
      // El conflicto es SIEMPRE sobre el instante elegido: además de la
      // alerta global, el mismo detalle del servidor marca el campo de hora
      // (atlas evento 11, dos niveles de error a la vez). No se inventa
      // texto: es el `detail` que devolvió el API.
      fieldErrors.value = { ...fieldErrors.value, startsAt: outcome.detail }
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

function onBarberSelect(barberId: string) {
  selectedBarberId.value = barberId
}
</script>

<template>
  <section class="new-appointment-page" aria-labelledby="new-appointment-title">
    <header class="new-appointment-page__header">
      <h1 id="new-appointment-title" class="new-appointment-page__title">Nuevo turno</h1>
      <p v-if="barbershopTimezone" class="new-appointment-page__timezone">
        Horas en la zona horaria de la barbería: {{ barbershopTimezone }}
      </p>
    </header>

    <!-- Evento 01: sin barberos ni zona resueltos, la espera conserva la
         geometría del formulario por llegar (esqueleto, no un indicador
         suelto sobre una pantalla vacía). -->
    <NewAppointmentSkeleton v-if="pageStatus === 'loading'" label="Cargando barberos…" />

    <!-- Evento 02: fallo al cargar el contexto inicial. -->
    <div
      v-else-if="pageStatus === 'load-error'"
      class="new-appointment-page__state new-appointment-page__state--warning"
    >
      <PageState
        variant="warning"
        status-label="Atención"
        headline="No pudimos cargar esta sección"
        role="alert"
      >
        Revisa tu conexión e inténtalo de nuevo.
        <template #action>
          <BaseButton type="button" variant="secondary" @click="onRetryLoad">Reintentar</BaseButton>
        </template>
      </PageState>
    </div>

    <template v-else>
      <!-- Evento 03: sin barberos activos no hay formulario que ofrecer. El
           destino nombrado en el mensaje se resalta en latón, sin convertirse
           en un enlace que el copy real no promete. -->
      <div v-if="barbers.length === 0" class="new-appointment-page__state">
        <PageState variant="info" headline="Aún no tienes barberos registrados." role="status">
          Agrega uno en la sección
          <span class="new-appointment-page__dest">"Barberos"</span> antes de registrar turnos.
        </PageState>
      </div>

      <!-- Evento 12: el éxito es la pantalla completa —divisor, rótulo,
           titular en serif y ficha de pergamino con el turno creado—, no una
           alerta pequeña flotando arriba a la izquierda. -->
      <div
        v-else-if="created"
        class="new-appointment-page__state new-appointment-page__state--success"
      >
        <PageState
          variant="success"
          status-label="Confirmación"
          headline="Turno registrado"
          role="status"
        >
          <span class="new-appointment-page__ticket">
            {{ created.attendeeName }} · {{ created.serviceName }} ({{
              created.durationMinutes
            }}
            min) · {{ created.priceAmount }} {{ created.currency }}
          </span>
          <span class="new-appointment-page__success-note">
            El turno quedó confirmado en la agenda del barbero. Las notificaciones al cliente
            todavía no están disponibles (B5 las agrega más adelante).
          </span>
          <template #action>
            <BaseButton
              type="button"
              variant="primary"
              class="new-appointment-page__submit"
              @click="onStartNewAppointment"
            >
              Registrar otro turno
            </BaseButton>
          </template>
        </PageState>
      </div>

      <form
        v-else
        class="new-appointment-page__layout"
        :class="{ 'new-appointment-page__layout--solo': !hasRail }"
        novalidate
        @submit.prevent="onSubmit"
      >
        <!-- Hoja continua: las cuatro secciones son bandas de UNA superficie
             translúcida con filete de latón a la izquierda, separadas entre sí
             por una línea de 1px. Nada de tarjetas de pergamino: el pergamino
             es material de REGISTRO (la ficha del evento 12), no cromo de
             formulario. -->
        <div class="new-appointment-page__form">
          <div class="new-appointment-page__sheet">
            <section class="new-appointment-page__section">
              <div class="new-appointment-page__section-head">
                <span class="new-appointment-page__section-number" aria-hidden="true">1</span>
                <div class="new-appointment-page__section-stack">
                  <h2 class="new-appointment-page__section-title">Selecciona</h2>
                  <p class="new-appointment-page__section-hint">Elige al barbero y el servicio.</p>
                </div>
              </div>

              <div class="new-appointment-page__section-body">
                <div class="new-appointment-page__row">
                  <div
                    class="new-appointment-page__field"
                    :class="{
                      'new-appointment-page__field--filled': !!selectedBarberId,
                      'new-appointment-page__field--error': attempted && !!fieldErrors.barberId,
                    }"
                  >
                    <label for="new-appointment-barber" class="new-appointment-page__label">
                      Barbero
                      <span class="new-appointment-page__required" aria-hidden="true">*</span>
                    </label>
                    <BarberSelect
                      :model-value="selectedBarberId || null"
                      :barbers="barbers"
                      trigger-id="new-appointment-barber"
                      placeholder="Elige un barbero"
                      @update:model-value="onBarberSelect"
                    />
                    <p
                      v-if="attempted && fieldErrors.barberId"
                      class="new-appointment-page__error"
                      role="alert"
                    >
                      {{ fieldErrors.barberId }}
                    </p>
                  </div>

                  <div
                    class="new-appointment-page__field"
                    :class="{
                      'new-appointment-page__field--filled': !!selectedServiceId,
                      'new-appointment-page__field--off': isServiceDisabled,
                      'new-appointment-page__field--error': attempted && !!fieldErrors.serviceId,
                    }"
                  >
                    <label for="new-appointment-service" class="new-appointment-page__label">
                      Servicio
                      <span class="new-appointment-page__required" aria-hidden="true">*</span>
                    </label>
                    <div class="new-appointment-page__select-wrap">
                      <select
                        id="new-appointment-service"
                        v-model="selectedServiceId"
                        class="new-appointment-page__select"
                        :disabled="isServiceDisabled"
                      >
                        <option value="" disabled>
                          {{ serviceSelectPlaceholder }}
                        </option>
                        <option v-for="s in services" :key="s.id" :value="s.id">
                          {{ s.name }}
                        </option>
                      </select>
                      <span class="new-appointment-page__caret" aria-hidden="true" />
                    </div>
                    <p
                      v-if="servicesStatus === 'loading'"
                      class="new-appointment-page__note new-appointment-page__note--loading"
                      role="status"
                      aria-live="polite"
                    >
                      <span class="new-appointment-page__note-spinner" aria-hidden="true" />
                      <span>Cargando servicios…</span>
                    </p>
                    <p
                      v-else-if="servicesStatus === 'ready' && services.length === 0"
                      class="new-appointment-page__note new-appointment-page__note--warning"
                    >
                      Este barbero no tiene servicios activos asignados.
                    </p>
                    <p
                      v-else-if="servicesStatus === 'error'"
                      class="new-appointment-page__note new-appointment-page__note--danger"
                      role="alert"
                    >
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
              </div>
            </section>

            <section class="new-appointment-page__section">
              <div class="new-appointment-page__section-head">
                <span class="new-appointment-page__section-number" aria-hidden="true">2</span>
                <div class="new-appointment-page__section-stack">
                  <h2 class="new-appointment-page__section-title">Persona atendida</h2>
                  <p class="new-appointment-page__section-hint">
                    Indica quién recibirá el servicio.
                  </p>
                </div>
              </div>

              <div class="new-appointment-page__section-body">
                <div class="new-appointment-page__row">
                  <BaseInput
                    v-model="attendeeName"
                    type="text"
                    label="Persona atendida"
                    placeholder="Persona atendida"
                    required
                    :class="{ 'new-appointment-page__input--filled': !!attendeeName }"
                    :error="attempted ? fieldErrors.attendeeName : undefined"
                  />
                  <BaseInput
                    v-model="customerFullName"
                    type="text"
                    label="Nombre del cliente"
                    placeholder="Nombre del cliente"
                    required
                    :class="{ 'new-appointment-page__input--filled': !!customerFullName }"
                    :error="attempted ? fieldErrors.customerFullName : undefined"
                  />
                </div>
                <div class="new-appointment-page__row">
                  <BaseInput
                    v-model="customerPhone"
                    type="tel"
                    label="Teléfono (opcional)"
                    placeholder="+573001234567"
                    :class="{ 'new-appointment-page__input--filled': !!customerPhone }"
                    :error="attempted ? fieldErrors.customerPhone : undefined"
                  />
                  <BaseInput
                    v-model="customerEmail"
                    type="email"
                    label="Correo (opcional)"
                    placeholder="Correo (opcional)"
                    :class="{ 'new-appointment-page__input--filled': !!customerEmail }"
                    :error="attempted ? fieldErrors.customerEmail : undefined"
                  />
                </div>
                <p v-if="!customerPhone && !customerEmail" class="new-appointment-page__hint">
                  Sin teléfono ni correo, el cliente no recibirá recordatorios.
                </p>
              </div>
            </section>

            <section class="new-appointment-page__section">
              <div class="new-appointment-page__section-head">
                <span class="new-appointment-page__section-number" aria-hidden="true">3</span>
                <div class="new-appointment-page__section-stack">
                  <h2 class="new-appointment-page__section-title">Fecha y hora</h2>
                  <p class="new-appointment-page__section-hint">Define cuándo será el turno.</p>
                </div>
              </div>

              <div class="new-appointment-page__section-body">
                <div class="new-appointment-page__row">
                  <BaseInput
                    v-model="startsAtDate"
                    type="date"
                    label="Fecha del turno"
                    required
                    :class="{ 'new-appointment-page__input--filled': !!startsAtDate }"
                  />
                  <BaseInput
                    v-model="startsAtTime"
                    type="time"
                    label="Hora del turno"
                    required
                    :class="{ 'new-appointment-page__input--filled': !!startsAtTime }"
                    :error="attempted ? fieldErrors.startsAt : undefined"
                  />
                </div>
              </div>
            </section>

            <section class="new-appointment-page__section">
              <div class="new-appointment-page__section-head">
                <span class="new-appointment-page__section-number" aria-hidden="true">4</span>
                <div class="new-appointment-page__section-stack">
                  <h2 class="new-appointment-page__section-title">Nota (opcional)</h2>
                  <p class="new-appointment-page__section-hint">
                    Agrega información adicional si es necesario.
                  </p>
                </div>
              </div>

              <div class="new-appointment-page__section-body">
                <div
                  class="new-appointment-page__field"
                  :class="{
                    'new-appointment-page__field--filled': !!customerNote,
                    'new-appointment-page__field--error': attempted && !!fieldErrors.customerNote,
                  }"
                >
                  <label for="new-appointment-note" class="new-appointment-page__label">
                    Nota (opcional)
                  </label>
                  <!-- Caja multilínea con contador (atlas evento 08): el tope
                       del contador es el de `validateCustomerNote`, y no se
                       fija `maxlength` para que un exceso siga produciendo el
                       mensaje de validación real en vez de truncarse en
                       silencio. -->
                  <textarea
                    id="new-appointment-note"
                    ref="noteRef"
                    v-model="customerNote"
                    class="new-appointment-page__textarea"
                    rows="2"
                    :aria-invalid="attempted && !!fieldErrors.customerNote"
                    :aria-describedby="
                      attempted && fieldErrors.customerNote
                        ? 'new-appointment-note-error'
                        : 'new-appointment-note-counter'
                    "
                  />
                  <div class="new-appointment-page__field-foot">
                    <span
                      v-if="attempted && fieldErrors.customerNote"
                      id="new-appointment-note-error"
                      class="new-appointment-page__error"
                      role="alert"
                    >
                      {{ fieldErrors.customerNote }}
                    </span>
                    <span v-else />
                    <span id="new-appointment-note-counter" class="new-appointment-page__counter">
                      {{ customerNote.length }}/{{ NOTE_MAX_LENGTH }}
                    </span>
                  </div>
                </div>
              </div>
            </section>
          </div>
        </div>

        <!-- Columna lateral: alerta global y resumen, en ese orden. En
             escritorio queda junto al CTA sin empujar el formulario; en móvil
             la retícula colapsa y cae entre el último campo y el CTA, que es
             donde especificacion-frontend-nava.md §7.3 lo pide. -->
        <aside v-if="hasRail" class="new-appointment-page__rail">
          <!-- El titular de la alerta va en el cuerpo, no en el prop `title`
               de BaseAlert: ese prop compone un <h4> y esta pantalla ya tiene
               un <h2> por banda, así que un titular de alerta rompería el
               orden de encabezados (axe `heading-order`). La composición del
               atlas —rótulo ERROR, línea destacada y cuerpo— se conserva
               entera. -->
          <BaseAlert v-if="saveStatus === 'conflict'" variant="danger" role="alert">
            <span class="new-appointment-page__alert-title">{{ saveErrorDetail }}</span>
          </BaseAlert>
          <BaseAlert v-else-if="saveStatus === 'validation-error'" variant="danger" role="alert">
            <span class="new-appointment-page__alert-title">
              {{ saveErrorDetail ?? 'Revisa los datos del turno.' }}
            </span>
          </BaseAlert>
          <BaseAlert
            v-else-if="saveStatus === 'idempotency-conflict'"
            variant="danger"
            role="alert"
          >
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
                <dt class="new-appointment-page__resumen-label">Barbero</dt>
                <dd class="new-appointment-page__resumen-value">
                  <BarberAvatar :full-name="selectedBarberName" size="closed" />
                  <span>{{ selectedBarberName }}</span>
                </dd>
              </div>
              <div v-if="selectedServiceName" class="new-appointment-page__resumen-item">
                <dt class="new-appointment-page__resumen-label">Servicio</dt>
                <dd class="new-appointment-page__resumen-value">
                  <span>{{ selectedServiceName }}</span>
                </dd>
              </div>
              <div v-if="attendeeName.trim()" class="new-appointment-page__resumen-item">
                <dt class="new-appointment-page__resumen-label">Persona atendida</dt>
                <dd class="new-appointment-page__resumen-value">
                  <span>{{ attendeeName }}</span>
                </dd>
              </div>
              <div
                v-if="summaryDateLabel || startsAtTime"
                class="new-appointment-page__resumen-item"
              >
                <dt class="new-appointment-page__resumen-label">Fecha y hora</dt>
                <dd class="new-appointment-page__resumen-value">
                  <span>
                    <span v-if="summaryDateLabel">{{ summaryDateLabel }}</span>
                    <span v-if="summaryDateLabel && startsAtTime"> · </span>
                    <span v-if="startsAtTime">{{ startsAtTime }}</span>
                  </span>
                </dd>
              </div>
            </dl>
          </div>
        </aside>

        <!-- CTA sangrado hasta la columna de campos: la misma medida que
             gobierna la columna de rótulos, para que la acción caiga bajo los
             controles y no bajo los títulos de sección. En móvil ocupa el
             ancho completo. -->
        <div class="new-appointment-page__actions">
          <BaseButton
            type="submit"
            variant="primary"
            class="new-appointment-page__submit"
            :disabled="saveStatus === 'saving'"
          >
            {{ saveStatus === 'saving' ? 'Guardando…' : 'Registrar turno' }}
          </BaseButton>
        </div>
      </form>
    </template>
  </section>
</template>

<style scoped>
/* Superficie de tinta de punta a punta, igual que /panel: el formulario vive
   SOBRE el canvas, no sobre tarjetas de pergamino (atlas nuevo-turno-eventos,
   "Contrato visual común"). Móvil primero (viewport 420 del atlas); la escala
   de escritorio (1440) entra en @media (min-width: 1024px). */
.new-appointment-page {
  display: flex;
  flex-direction: column;
  gap: 18px;
  min-height: 100%;
  padding: 20px;
  background-color: var(--color-surface-strong);
}

.new-appointment-page__header {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.new-appointment-page__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: 32px;
  line-height: 1.05;
  font-weight: var(--font-weight-h1);
  color: var(--color-on-strong);
}

.new-appointment-page__timezone {
  margin: 0;
  font-family: var(--font-family-base);
  font-size: 14px;
  line-height: 20px;
  color: var(--color-on-strong-muted);
}

/* ---------------------------------------------------------------- retícula */

.new-appointment-page__layout {
  display: grid;
  align-items: start;
  grid-template-columns: minmax(0, 1fr);
  grid-template-areas: 'form' 'rail' 'acts';
  row-gap: 16px;
}

.new-appointment-page__layout--solo {
  grid-template-areas: 'form' 'acts';
}

.new-appointment-page__form {
  grid-area: form;
  display: flex;
  flex-direction: column;
}

.new-appointment-page__rail {
  grid-area: rail;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.new-appointment-page__actions {
  grid-area: acts;
  display: flex;
  flex-direction: column;
}

/* ------------------------------------------------------ hoja del formulario */

/* Las cuatro secciones son bandas de UNA hoja separadas por filete, no cuatro
   tarjetas apiladas: con borde propio cada una, el formulario se leía como una
   pila rayada de objetos sueltos en vez de un documento. */
.new-appointment-page__sheet {
  background-color: rgb(244 240 231 / 3.5%);
  border: var(--border-width-normal) solid rgb(244 240 231 / 10%);
  border-left: 3px solid rgb(184 149 90 / 55%);
  border-radius: 2px;
}

.new-appointment-page__section {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 18px 16px;
}

.new-appointment-page__section + .new-appointment-page__section {
  border-top: var(--border-width-normal) solid rgb(244 240 231 / 10%);
}

.new-appointment-page__section-head {
  display: flex;
  align-items: flex-start;
  gap: 11px;
}

/* Cuadrado de radio 2px con filete de latón y cifra en serif: el círculo era
   el único elemento redondo de todo el sistema. */
.new-appointment-page__section-number {
  display: flex;
  flex: none;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border: var(--border-width-normal) solid rgb(184 149 90 / 60%);
  border-radius: 2px;
  font-family: var(--font-display);
  font-size: 14px;
  line-height: 1;
  color: var(--color-brand-accent-surface);
}

.new-appointment-page__section-stack {
  display: flex;
  min-width: 0;
  flex-direction: column;
}

.new-appointment-page__section-title {
  margin: 0;
  font-family: var(--font-display);
  font-size: 19px;
  line-height: 1.15;
  font-weight: 400;
  color: var(--color-on-strong);
}

.new-appointment-page__section-hint {
  margin: 2px 0 0;
  font-family: var(--font-family-base);
  font-size: 13px;
  line-height: 18px;
  color: var(--color-on-strong-muted);
  text-wrap: balance;
}

.new-appointment-page__section-body {
  display: flex;
  flex-direction: column;
  gap: 13px;
}

.new-appointment-page__row {
  display: flex;
  flex-direction: column;
  gap: 13px;
}

.new-appointment-page__hint {
  margin: 0;
  font-family: var(--font-family-base);
  font-size: 12px;
  line-height: 17px;
  color: var(--color-on-strong-muted);
}

/* ------------------------------------------------- campos reglados sobre tinta */

.new-appointment-page__field {
  display: flex;
  flex: 1;
  min-width: 0;
  flex-direction: column;
  gap: 6px;
}

.new-appointment-page__label {
  font-family: var(--font-sans);
  font-size: 10px;
  font-weight: 600;
  line-height: 14px;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--color-brand-accent-surface);
}

/* El campo reglado sobre tinta se resuelve redefiniendo los tokens que
   BaseInput ya expone (--input-bg / --input-border-color /
   --input-border-base-color), en el ámbito de esta página: shared/ui sigue
   calibrado para superficie clara, que es donde vive en el resto de la app
   (auth-eventos, diálogos de configuración). */
.new-appointment-page :deep(.base-input) {
  --input-height: 46px;
  --input-padding-x: 13px;
  --input-font-size: 15px;
  --input-line-height: 1.3;
  --input-bg: rgb(244 240 231 / 4%);
  --input-border-color: rgb(244 240 231 / 12%);
  --input-border-base-color: rgb(244 240 231 / 30%);
  --input-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);

  color: var(--color-on-strong);
}

.new-appointment-page :deep(.base-input__wrapper) {
  gap: 6px;
}

.new-appointment-page :deep(.base-input__label) {
  font-size: 10px;
  line-height: 14px;
  letter-spacing: 0.14em;
}

/* Asterisco de obligatorio en latón, nunca en rojo: el rojo se reserva para
   error (atlas, "Contrato visual común"). El margen negativo absorbe el
   espacio en blanco que la plantilla de BaseInput deja entre el rótulo y el
   asterisco: en versalitas espaciadas ese espacio se lee como un asterisco
   suelto, y el atlas lo pega al rótulo ("PERSONA ATENDIDA*"). */
.new-appointment-page :deep(.base-input__required),
.new-appointment-page__required {
  margin-left: -5px;
  color: var(--color-brand-accent-surface);
}

.new-appointment-page :deep(.base-input::placeholder) {
  color: var(--color-on-strong-muted);
  opacity: 1;
}

.new-appointment-page :deep(.base-input:hover:not(:disabled):not(.base-input--invalid)) {
  border-color: rgb(244 240 231 / 26%);
}

/* El latón del filete inferior marca el campo YA RESUELTO; uno vacío lleva
   filete neutro. Ocho subrayados dorados a la vez convertían el formulario en
   un muestrario de oro y ya no distinguían lo hecho de lo pendiente (compara
   los eventos 04, 05 y 08 del atlas). */
.new-appointment-page :deep(.new-appointment-page__input--filled .base-input) {
  background-color: rgb(244 240 231 / 6%);
  border-bottom-color: var(--color-brand-accent-surface);
}

.new-appointment-page :deep(.base-input:disabled),
.new-appointment-page :deep(.base-input--disabled) {
  background-color: var(--input-bg);
  border-color: var(--input-border-color);
  border-bottom-color: rgb(244 240 231 / 20%);
  color: var(--color-on-strong-muted);
  opacity: 0.45;
}

.new-appointment-page :deep(.base-input--invalid) {
  background-color: rgb(227 146 141 / 7%);
  border-color: var(--input-border-color);
  border-bottom-color: var(--color-danger-on-strong);
}

.new-appointment-page :deep(.base-input__error) {
  font-size: 12px;
  line-height: 16px;
  font-weight: 600;
  color: var(--color-danger-on-strong);
}

/* El control nativo de fecha/hora pinta su propio icono de calendario/reloj;
   el atlas compone estos dos campos como texto reglado, igual que /panel. El
   campo entero sigue abriendo el selector nativo al hacer clic. */
.new-appointment-page :deep(.base-input::-webkit-calendar-picker-indicator) {
  display: none;
}

/* Selector de servicio: `<select>` nativo (accesibilidad y comportamiento
   intactos) con la misma piel del campo reglado y el caret de latón del
   atlas. */
.new-appointment-page__select-wrap {
  position: relative;
  display: flex;
}

.new-appointment-page__select {
  width: 100%;
  height: 46px;
  padding: 0 34px 0 13px;
  font-family: var(--font-family-base);
  font-size: 15px;
  line-height: 1.3;
  color: var(--color-on-strong);
  background-color: rgb(244 240 231 / 4%);
  border: var(--border-width-normal) solid rgb(244 240 231 / 12%);
  border-bottom: var(--border-width-emphasis) solid rgb(244 240 231 / 30%);
  border-radius: 2px;
  outline: none;
  appearance: none;
}

/* La lista desplegada la pinta el sistema operativo sobre superficie clara:
   sin esto, texto marfil sobre blanco. */
.new-appointment-page__select option {
  color: var(--color-text-primary);
  background-color: var(--color-surface);
}

.new-appointment-page__select:focus-visible {
  border-color: var(--color-focus);
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus);
}

.new-appointment-page__select:disabled {
  cursor: not-allowed;
}

.new-appointment-page__caret {
  position: absolute;
  top: 50%;
  right: 14px;
  width: 7px;
  height: 7px;
  border-right: var(--border-width-normal) solid var(--color-brand-accent-surface);
  border-bottom: var(--border-width-normal) solid var(--color-brand-accent-surface);
  transform: translateY(-70%) rotate(45deg);
  pointer-events: none;
}

.new-appointment-page__field--filled .new-appointment-page__select {
  background-color: rgb(244 240 231 / 6%);
  border-bottom-color: var(--color-brand-accent-surface);
}

/* Deshabilitado (servicio sin barbero elegido o cargando): se atenúa y pierde
   el latón, nunca lo conserva (atlas eventos 04 y 05). */
.new-appointment-page__field--off .new-appointment-page__select,
.new-appointment-page__field--off .new-appointment-page__caret {
  opacity: 0.45;
}

.new-appointment-page__field--off .new-appointment-page__select {
  border-bottom-color: rgb(244 240 231 / 20%);
}

.new-appointment-page__field--error .new-appointment-page__select,
.new-appointment-page__field--error .new-appointment-page__textarea {
  background-color: rgb(227 146 141 / 7%);
  border-bottom-color: var(--color-danger-on-strong);
}

/* Selector de barbero (BarberSelect, con monograma): misma altura y piel que
   el resto de los campos de esta hoja. Sobre /panel el control vive suelto y
   siempre lleva su línea de latón; aquí obedece la misma regla de "resuelto"
   que los demás. */
.new-appointment-page__field :deep(.barber-select__trigger) {
  min-height: 46px;
  padding: 0 13px;
  gap: 10px;
  font-size: 15px;
  background-color: rgb(244 240 231 / 4%);
  border-color: rgb(244 240 231 / 12%);
  border-bottom-color: rgb(244 240 231 / 30%);
}

.new-appointment-page__field--filled :deep(.barber-select__trigger) {
  background-color: rgb(244 240 231 / 6%);
  border-bottom-color: var(--color-brand-accent-surface);
}

.new-appointment-page__field--error :deep(.barber-select__trigger) {
  background-color: rgb(227 146 141 / 7%);
  border-bottom-color: var(--color-danger-on-strong);
}

/* Sin barbero elegido el disparador dibuja un cuadrado vacío del tamaño del
   monograma; el atlas deja el campo con solo su placeholder. */
.new-appointment-page__field :deep(.barber-select__trigger-icon) {
  display: none;
}

.new-appointment-page__field :deep(.barber-select__trigger-label--placeholder) {
  color: var(--color-on-strong-muted);
}

.new-appointment-page__textarea {
  width: 100%;
  min-height: 70px;
  padding: 12px 13px;
  font-family: var(--font-family-base);
  font-size: 15px;
  line-height: 1.3;
  color: var(--color-on-strong);
  background-color: rgb(244 240 231 / 4%);
  border: var(--border-width-normal) solid rgb(244 240 231 / 12%);
  border-bottom: var(--border-width-emphasis) solid rgb(244 240 231 / 30%);
  border-radius: 2px;
  outline: none;
  overflow-y: hidden;
  resize: none;
}

.new-appointment-page__textarea:focus-visible {
  border-color: var(--color-focus);
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus);
}

.new-appointment-page__field--filled .new-appointment-page__textarea {
  background-color: rgb(244 240 231 / 6%);
  border-bottom-color: var(--color-brand-accent-surface);
}

.new-appointment-page__field-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.new-appointment-page__counter {
  font-family: var(--font-family-base);
  font-size: 11px;
  color: var(--color-on-strong-muted);
  font-variant-numeric: tabular-nums;
}

.new-appointment-page__error {
  margin: 0;
  font-family: var(--font-family-base);
  font-size: 12px;
  line-height: 16px;
  font-weight: 600;
  color: var(--color-danger-on-strong);
}

.new-appointment-page__note {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  font-family: var(--font-family-base);
  font-size: 12px;
  line-height: 16px;
  color: var(--color-on-strong-muted);
}

.new-appointment-page__note--loading {
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--color-brand-accent-surface);
}

.new-appointment-page__note--warning {
  color: var(--color-warning-on-strong);
}

.new-appointment-page__note--danger {
  color: var(--color-danger-on-strong);
}

.new-appointment-page__note-spinner {
  width: 13px;
  height: 13px;
  flex-shrink: 0;
  border: 1.5px solid rgb(184 149 90 / 22%);
  border-top-color: var(--color-brand-accent-surface);
  border-right-color: var(--color-brand-accent-surface);
  border-radius: 50%;
  transform: rotate(-38deg);
}

/* -------------------------------------------------------------- resumen */

.new-appointment-page__resumen {
  padding: 16px;
  background-color: #16243a;
  border: var(--border-width-normal) solid rgb(244 240 231 / 12%);
  border-top: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  border-radius: 2px;
}

.new-appointment-page__resumen-title {
  margin: 0;
  font-family: var(--font-sans);
  font-size: 10px;
  font-weight: 600;
  line-height: 14px;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--color-brand-accent-surface);
}

.new-appointment-page__resumen-list {
  display: flex;
  flex-direction: column;
  margin: 12px 0 0;
}

.new-appointment-page__resumen-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
  padding: 11px 0;
}

.new-appointment-page__resumen-item:first-child {
  padding-top: 0;
}

.new-appointment-page__resumen-item:last-child {
  padding-bottom: 0;
}

.new-appointment-page__resumen-item + .new-appointment-page__resumen-item {
  border-top: var(--border-width-normal) solid rgb(244 240 231 / 10%);
}

.new-appointment-page__resumen-label {
  font-family: var(--font-sans);
  font-size: 10px;
  font-weight: 600;
  line-height: 14px;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--color-on-strong-muted);
}

.new-appointment-page__resumen-value {
  display: flex;
  align-items: center;
  gap: 9px;
  margin: 0;
  font-family: var(--font-family-base);
  font-size: 15px;
  line-height: 21px;
  font-weight: 600;
  color: var(--color-on-strong);
}

/* ------------------------------------------------------- alerta y acciones */

/* Botón primario sólido en latón sobre el canvas de tinta: BaseButton
   --primary está calibrado en tinta para superficie clara, que es lo correcto
   en el resto de la app. El prefijo `.new-appointment-page` sube la
   especificidad por encima de `.base-button--primary:hover:not(...)` para que
   el orden de inserción de los estilos del componente no decida el resultado.
   En `saving` se atenúa por opacidad y su rótulo cambia a «Guardando…», sin
   duplicar un segundo indicador de carga. */
.new-appointment-page .new-appointment-page__submit {
  width: 100%;
  height: 48px;
  padding: 0 22px;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.02em;
  background-color: var(--color-brand-accent-surface);
  border-color: var(--color-brand-accent-surface);
  color: var(--color-surface-strong);
}

.new-appointment-page .new-appointment-page__submit:hover:not(:disabled) {
  background-color: var(--color-brand-accent-surface);
  border-color: var(--color-brand-accent-surface);
  filter: brightness(94%);
}

.new-appointment-page .new-appointment-page__submit:focus-visible {
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus);
}

.new-appointment-page .new-appointment-page__submit:disabled {
  opacity: 0.55;
}

/* La acción del éxito es el mismo botón sólido en latón, pero centrado bajo
   la ficha y sin la medida completa del CTA del formulario (atlas evento
   12). */
.new-appointment-page .new-appointment-page__state .new-appointment-page__submit {
  width: auto;
  align-self: center;
}

/* Nota al margen del sistema con perímetro completo, como en el atlas: sobre
   tinta un filete lateral suelto dejaba la alerta sin contorno contra el
   canvas. */
.new-appointment-page :deep(.base-alert) {
  --alert-padding: 13px 15px;
  --alert-title-size: 16px;
  --alert-font-size: 14px;
  --alert-line-height: 20px;

  border: var(--border-width-normal) solid var(--alert-border);
  border-left-width: 4px;
}

.new-appointment-page :deep(.base-alert__status) {
  font-size: 10px;
  line-height: 14px;
  letter-spacing: 0.15em;
  color: currentcolor;
  opacity: 0.85;
}

.new-appointment-page__alert-title {
  display: block;
  font-size: 16px;
  line-height: 23px;
  font-weight: 600;
  color: var(--color-surface-strong);
}

/* ------------------------------------------------- estados de página completos */

/* El destino nombrado dentro del mensaje se resalta en latón, igual que
   «Barberos» en el evento 04 del atlas de /panel. */
.new-appointment-page__dest {
  font-weight: 600;
  color: var(--color-brand-accent-surface);
}

/* PageState fija su color de marca por estilo en línea, calibrado para los
   pares claros del sistema: sobre tinta, #9a6a24 (atención) y #325d43
   (confirmación) no alcanzan AA. Aquí se levantan a los tintes sobre tinta
   del atlas; el `!important` es la única forma de ganarle a un estilo en
   línea, y queda contenido en esta pantalla para no alterar /panel, que ya
   se validó con su propio atlas. */
.new-appointment-page__state {
  display: flex;
  flex: 1;
  flex-direction: column;
}

.new-appointment-page__state :deep(.page-state) {
  max-width: 330px;
  margin-inline: auto;
  padding: 34px 0;
  gap: 12px;
}

.new-appointment-page__state :deep(.page-state__divider) {
  width: 166px;
}

.new-appointment-page__state--warning :deep(.page-state) {
  --page-state-mark: var(--color-warning-on-strong) !important;
}

.new-appointment-page__state--success :deep(.page-state) {
  --page-state-mark: var(--color-success-on-strong) !important;
}

.new-appointment-page__ticket,
.new-appointment-page__success-note {
  display: block;
}

/* Ficha del turno recién creado: pergamino, el mismo material con el que la
   agenda dibuja un turno vigente. */
.new-appointment-page__ticket {
  padding: 13px 15px;
  margin-bottom: 12px;
  font-size: 15px;
  line-height: 22px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  text-align: left;
  color: var(--color-text-primary);
  background-color: var(--color-surface-muted);
  border-left: 3px solid var(--color-accent-brass);
  border-radius: 2px;
}

/* ------------------------------------------------------------- escritorio */

@media (min-width: 1024px) {
  .new-appointment-page {
    gap: 18px;
    padding: 22px 40px 24px;
  }

  .new-appointment-page__title {
    font-size: 38px;
  }

  .new-appointment-page__timezone {
    font-size: 15px;
    line-height: 22px;
  }

  /* La misma medida gobierna la columna de rótulos de cada banda y la
     sangría del CTA, para que el botón caiga exactamente bajo los campos y
     no bajo el título de la sección. */
  .new-appointment-page__layout {
    --label-col: 288px;
    --band-gap: 28px;
    --band-pad: 26px;

    grid-template-columns: minmax(0, 1fr) 340px;
    grid-template-areas: 'form rail' 'acts rail';
    column-gap: 32px;
    row-gap: 0;
  }

  .new-appointment-page__layout--solo {
    grid-template-columns: minmax(0, 980px);
    grid-template-areas: 'form' 'acts';
  }

  .new-appointment-page__actions {
    margin-top: 16px;
    padding-left: calc(var(--band-pad) + var(--label-col) + var(--band-gap) + 3px);
  }

  .new-appointment-page__section {
    display: grid;
    grid-template-columns: var(--label-col) minmax(0, 1fr);
    gap: var(--band-gap);
    padding: 20px var(--band-pad);
  }

  .new-appointment-page__section-head {
    gap: 12px;
  }

  .new-appointment-page__section-number {
    width: 27px;
    height: 27px;
    font-size: 15px;
  }

  .new-appointment-page__section-title {
    font-size: 21px;
  }

  .new-appointment-page__section-hint {
    margin-top: 3px;
  }

  .new-appointment-page__section-body {
    gap: 14px;
  }

  .new-appointment-page__row {
    flex-direction: row;
    gap: 20px;
  }

  .new-appointment-page__hint {
    font-size: 13px;
    line-height: 18px;
  }

  .new-appointment-page__field {
    gap: 7px;
  }

  .new-appointment-page__label,
  .new-appointment-page :deep(.base-input__label) {
    font-size: 11px;
    line-height: 15px;
  }

  .new-appointment-page :deep(.base-input__wrapper) {
    gap: 7px;
  }

  .new-appointment-page :deep(.base-input) {
    --input-padding-x: 14px;
  }

  .new-appointment-page__select {
    padding: 0 34px 0 14px;
  }

  .new-appointment-page__textarea {
    min-height: 74px;
    padding: 12px 14px;
  }

  .new-appointment-page__counter {
    font-size: 11px;
  }

  .new-appointment-page__resumen {
    padding: 18px 20px;
  }

  .new-appointment-page__resumen-title {
    font-size: 11px;
    line-height: 15px;
  }

  .new-appointment-page__resumen-list {
    margin-top: 14px;
  }

  .new-appointment-page__resumen-item {
    padding: 12px 0;
  }

  .new-appointment-page__resumen-value {
    gap: 10px;
    font-size: 16px;
    line-height: 22px;
  }

  .new-appointment-page .new-appointment-page__submit {
    width: auto;
    min-width: 240px;
    height: 50px;
    padding: 0 30px;
    font-size: 16px;
    align-self: flex-start;
  }

  .new-appointment-page .new-appointment-page__state .new-appointment-page__submit {
    min-width: 0;
  }

  .new-appointment-page :deep(.base-alert) {
    --alert-padding: 15px 18px;
    --alert-title-size: 17px;
    --alert-font-size: 15px;
    --alert-line-height: 22px;
  }

  .new-appointment-page :deep(.base-alert__status) {
    font-size: 11px;
    line-height: 15px;
  }

  .new-appointment-page__alert-title {
    font-size: 17px;
    line-height: 24px;
  }

  .new-appointment-page__state :deep(.page-state) {
    max-width: 520px;
    padding: 40px 0;
    gap: 14px;
  }

  .new-appointment-page__state :deep(.page-state__divider) {
    width: 220px;
  }

  .new-appointment-page__ticket {
    padding: 14px 18px;
    font-size: 16px;
    line-height: 23px;
  }
}
</style>
