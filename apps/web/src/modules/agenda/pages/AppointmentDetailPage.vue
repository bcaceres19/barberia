<script setup lang="ts">
// Detalle e historial de un turno (HU-064) más reprogramación T2 (HU-065):
// persona atendida, cliente que reservó con su contacto opcional y nota,
// servicio snapshot (DEC-004), precio, estado e historial inmutable
// paginado, siguiendo la plantilla P0 "Detalle de turno"
// (estandar-diseno-visual.md §10). "Volver" conserva la fecha/barbero de
// origen (HU-063) mediante `route.query`; tras reprogramar con éxito, la
// fecha se actualiza a la del nuevo inicio para no dejar la fecha vieja
// presentada como vigente. Fuera de alcance a propósito: T3, cancelación o
// cualquier otro cambio de estado (ninguna otra acción se renderiza).
import { computed, onMounted, ref } from 'vue'
import { useRoute, type LocationQueryRaw } from 'vue-router'
import { BaseAlert, BaseButton, BaseDialog, BaseInput } from '@/shared/ui'
import { formatInstantInTimezone } from '@/shared/time/formatInstant'
import { getCivilDateInTimezone } from '@/shared/time/civilDate'
import {
  fetchAppointmentDetail,
  fetchAppointmentHistory,
  fetchBarbershopTimezone,
  rescheduleAppointment,
} from '../api/appointmentsApi'
import { newIdempotencyKey } from '../model/idempotencyKey'
import type { AppointmentDetail, HistoryEntry } from '../model/appointmentDetail'
import { HISTORY_EVENT_LABELS, historyFieldLabel } from '../model/appointmentDetail'
import { APPOINTMENT_STATUS_LABELS } from '../model/dailyAgenda'

type PageStatus = 'loading' | 'ready' | 'not-found' | 'error'
type HistoryStatus = 'loading' | 'ready' | 'error'
type RescheduleStatus =
  | 'idle'
  | 'saving'
  | 'version-conflict'
  | 'invalid-state'
  | 'conflict'
  | 'idempotency-conflict'
  | 'validation-error'
  | 'network-error'
  | 'unexpected-error'

const route = useRoute()

const appointmentId = computed(() => String(route.params.appointmentId ?? ''))

// backDate/backBarberId inician con la fecha/barbero de origen (HU-063) y
// SOLO backDate cambia después, tras reprogramar con éxito (§3.5 del
// trabajo requerido): "Volver" siempre regresa a donde el turno vigente
// realmente está, nunca a una fecha que dejó de mostrarlo.
const backDate = ref<string | null>(typeof route.query.date === 'string' ? route.query.date : null)
const backBarberId = ref<string | null>(
  typeof route.query.barberId === 'string' ? route.query.barberId : null,
)
const backQuery = computed<LocationQueryRaw>(() => ({
  ...(backDate.value ? { date: backDate.value } : {}),
  ...(backBarberId.value ? { barberId: backBarberId.value } : {}),
}))

const pageStatus = ref<PageStatus>('loading')
const detail = ref<AppointmentDetail | null>(null)
const barbershopTimezone = ref<string | null>(null)

const historyStatus = ref<HistoryStatus>('loading')
const historyItems = ref<HistoryEntry[]>([])
const historyNextCursor = ref<string | null>(null)
const historyLoadingMore = ref(false)

const statusLabel = computed(() =>
  detail.value ? APPOINTMENT_STATUS_LABELS[detail.value.status] : '',
)
const attendeeInitials = computed(() => {
  if (!detail.value) return ''
  return detail.value.attendeeName
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part.slice(0, 1).toUpperCase())
    .join('')
})

// El atlas usa una fotografía editorial, pero ese activo no pertenece al
// contrato de la cita. El monograma conserva la jerarquía sin inventar una
// imagen ni exponer un dato adicional.
const FACT_ICONS: Record<string, string> = {
  time: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9"></circle><path d="M12 7v5l3 2"></path></svg>`,
  attendee: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="7" r="3.5"></circle><path d="M4.5 21c.6-4.2 3.5-6.5 7.5-6.5s6.9 2.3 7.5 6.5"></path></svg>`,
  service: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M6 3c3 3 4 6 4 9s-1 6-4 9"></path><path d="M18 3c-3 3-4 6-4 9s1 6 4 9"></path><path d="M8.5 8.5 15.5 15.5"></path><path d="M15.5 8.5 8.5 15.5"></path></svg>`,
  barber: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="7" r="3.5"></circle><path d="M4.5 21c.6-4.2 3.5-6.5 7.5-6.5s6.9 2.3 7.5 6.5"></path><path d="M3.5 12.5h3M17.5 12.5h3"></path></svg>`,
  contact: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="5" width="18" height="14" rx="2"></rect><path d="m4 7 8 6 8-6"></path></svg>`,
  note: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M5 4h14v12H9l-4 4V4Z"></path></svg>`,
  calendar: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="5" width="16" height="15" rx="2"></rect><path d="M8 3v4M16 3v4M4 10h16"></path></svg>`,
}

function factIcon(name: keyof typeof FACT_ICONS): string {
  return FACT_ICONS[name]
}

function historyIcon(entry: HistoryEntry): string {
  return entry.eventType === 'appointment_rescheduled'
    ? `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><path d="M4 17.5V20h2.5L18 8.5 15.5 6 4 17.5Z"></path><path d="m14.5 7 2.5 2.5"></path></svg>`
    : FACT_ICONS.calendar
}

const timeRangeLabel = computed(() => {
  if (!detail.value || !barbershopTimezone.value) return ''
  const start = formatInstantInTimezone(detail.value.startsAt, barbershopTimezone.value)
  const end = new Intl.DateTimeFormat('es-CO', {
    timeZone: barbershopTimezone.value,
    timeStyle: 'short',
  }).format(new Date(detail.value.endsAt))
  return `${start} – ${end}`
})

// Reprogramar turno (HU-065, T2): solo visible sobre un turno confirmado
// (RN-* del prompt — T3/cancelación no existen todavía, así que ningún
// otro estado ofrece una acción propia).
const canReschedule = computed(() => detail.value?.status === 'confirmed')

const isRescheduleOpen = ref(false)
const rescheduleDate = ref('')
const rescheduleTime = ref('')
const rescheduleStatus = ref<RescheduleStatus>('idle')
const rescheduleErrorDetail = ref<string | null>(null)
// Clave mutable de módulo (no `ref`), mismo criterio que
// NewAppointmentPage.vue: se reutiliza tal cual entre reintentos del MISMO
// intento (un doble toque reproduce la misma petición en vez de duplicarla)
// y solo se renueva al abrir el diálogo de nuevo, que es aquí el límite
// natural de "intento lógico" (el diálogo se cierra y reabre, a diferencia
// de un formulario de página completa).
let rescheduleIdempotencyKey = newIdempotencyKey()

function civilTimeInTimezone(iso: string, timezone: string): string {
  return new Intl.DateTimeFormat('en-GB', {
    timeZone: timezone,
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(iso))
}

// addMinutesToTimeString: aritmética pura de minutos sobre "HH:MM", sin
// Date/Intl/zona horaria. Es una vista previa barata para el barbero
// mientras escribe; el cómputo autoritativo del nuevo fin ocurre en el
// servidor contra duration_minutes_snapshot en la zona real de la
// barbería, así que esta función nunca necesita ser exacta ante DST.
function addMinutesToTimeString(time: string, minutes: number): string {
  const [hours, mins] = time.split(':').map(Number)
  if (hours === undefined || mins === undefined || Number.isNaN(hours) || Number.isNaN(mins)) {
    return ''
  }
  const totalMinutesInDay = 24 * 60
  const total =
    (((hours * 60 + mins + minutes) % totalMinutesInDay) + totalMinutesInDay) % totalMinutesInDay
  const endHours = Math.floor(total / 60)
  const endMins = total % 60
  return `${String(endHours).padStart(2, '0')}:${String(endMins).padStart(2, '0')}`
}

const rescheduleNewEndLabel = computed(() => {
  if (!detail.value || !rescheduleTime.value) return ''
  return addMinutesToTimeString(rescheduleTime.value, detail.value.durationMinutes)
})

function onOpenReschedule() {
  if (!detail.value || !barbershopTimezone.value) return
  rescheduleDate.value = getCivilDateInTimezone(
    barbershopTimezone.value,
    new Date(detail.value.startsAt),
  )
  rescheduleTime.value = civilTimeInTimezone(detail.value.startsAt, barbershopTimezone.value)
  rescheduleStatus.value = 'idle'
  rescheduleErrorDetail.value = null
  rescheduleIdempotencyKey = newIdempotencyKey()
  isRescheduleOpen.value = true
}

function onRescheduleDialogClosed() {
  isRescheduleOpen.value = false
}

async function onSubmitReschedule() {
  if (!detail.value || rescheduleStatus.value === 'saving') return
  if (!rescheduleDate.value || !rescheduleTime.value) return

  rescheduleStatus.value = 'saving'
  rescheduleErrorDetail.value = null

  const startsAt = `${rescheduleDate.value}T${rescheduleTime.value}:00`
  const outcome = await rescheduleAppointment(
    appointmentId.value,
    { startsAt },
    detail.value.versionToken,
    rescheduleIdempotencyKey,
  )

  switch (outcome.kind) {
    case 'success': {
      isRescheduleOpen.value = false
      if (barbershopTimezone.value) {
        backDate.value = getCivilDateInTimezone(
          barbershopTimezone.value,
          new Date(outcome.startsAt),
        )
      }
      await loadPage()
      return
    }
    // 'not-found' aquí solo puede significar que el turno desapareció
    // mientras el diálogo estaba abierto (eliminado o de otra barbería):
    // la única salida coherente es la misma que un conflicto de versión,
    // recargar el detalle real.
    case 'not-found':
      rescheduleStatus.value = 'version-conflict'
      return
    case 'version-conflict':
      rescheduleStatus.value = 'version-conflict'
      return
    case 'invalid-state':
      rescheduleStatus.value = 'invalid-state'
      return
    case 'conflict':
      rescheduleErrorDetail.value = outcome.detail
      rescheduleStatus.value = 'conflict'
      return
    case 'validation-error':
      rescheduleErrorDetail.value = outcome.detail
      rescheduleStatus.value = 'validation-error'
      return
    case 'idempotency-conflict':
      rescheduleStatus.value = 'idempotency-conflict'
      return
    case 'network-error':
      rescheduleStatus.value = 'network-error'
      return
    case 'unexpected-error':
      rescheduleStatus.value = 'unexpected-error'
  }
}

// onReloadAfterConflict recarga el detalle real (versionToken/estado
// vigentes) tras un conflicto de versión o de estado: la intención del
// barbero (fecha/hora elegidas) no puede reenviarse contra datos obsoletos,
// así que cerrar y recargar es la única salida segura.
function onReloadAfterConflict() {
  isRescheduleOpen.value = false
  void loadPage()
}

async function loadPage() {
  pageStatus.value = 'loading'
  const [detailOutcome, timezoneOutcome] = await Promise.all([
    fetchAppointmentDetail(appointmentId.value),
    fetchBarbershopTimezone(),
  ])
  barbershopTimezone.value = timezoneOutcome.kind === 'success' ? timezoneOutcome.timezone : null

  switch (detailOutcome.kind) {
    case 'success':
      detail.value = detailOutcome.detail
      pageStatus.value = 'ready'
      void loadHistory()
      return
    case 'not-found':
      pageStatus.value = 'not-found'
      return
    default:
      pageStatus.value = 'error'
  }
}

onMounted(loadPage)

function onRetryLoad() {
  void loadPage()
}

async function loadHistory(cursor?: string) {
  if (cursor) {
    historyLoadingMore.value = true
  } else {
    historyStatus.value = 'loading'
    historyItems.value = []
  }

  const outcome = await fetchAppointmentHistory(appointmentId.value, cursor)
  switch (outcome.kind) {
    case 'success':
      historyItems.value = cursor ? [...historyItems.value, ...outcome.items] : outcome.items
      historyNextCursor.value = outcome.nextCursor
      historyStatus.value = 'ready'
      break
    default:
      historyStatus.value = 'error'
  }
  historyLoadingMore.value = false
}

function onRetryHistory() {
  void loadHistory()
}

function onLoadMoreHistory() {
  if (historyNextCursor.value) void loadHistory(historyNextCursor.value)
}

function eventLabel(entry: HistoryEntry): string {
  return HISTORY_EVENT_LABELS[entry.eventType] ?? entry.eventType
}

function occurredAtLabel(entry: HistoryEntry): string {
  if (!barbershopTimezone.value) return ''
  return formatInstantInTimezone(entry.occurredAt, barbershopTimezone.value)
}
</script>

<template>
  <section class="appointment-detail-page" aria-labelledby="appointment-detail-page-title">
    <p class="appointment-detail-page__back">
      <RouterLink :to="{ name: 'panel', query: backQuery }">← Volver a la agenda</RouterLink>
    </p>

    <div
      v-if="pageStatus === 'loading'"
      class="appointment-detail-page__state"
      role="status"
      aria-live="polite"
    >
      <p>Cargando el detalle del turno…</p>
    </div>

    <BaseAlert
      v-else-if="pageStatus === 'not-found'"
      variant="warning"
      title="Este turno ya no está disponible"
      role="alert"
    >
      Puede haberse eliminado o pertenecer a otra barbería.
    </BaseAlert>

    <BaseAlert
      v-else-if="pageStatus === 'error'"
      variant="warning"
      title="No pudimos cargar este turno"
      role="alert"
    >
      Revisa tu conexión e inténtalo de nuevo.
      <template #action>
        <BaseButton type="button" variant="secondary" @click="onRetryLoad">Reintentar</BaseButton>
      </template>
    </BaseAlert>

    <template v-else-if="detail">
      <header class="appointment-detail-page__hero">
        <div class="appointment-detail-page__avatar" aria-hidden="true">{{ attendeeInitials }}</div>
        <div class="appointment-detail-page__identity">
          <h1 id="appointment-detail-page-title" class="appointment-detail-page__title">
            {{ detail.attendeeName }}
          </h1>
          <p class="appointment-detail-page__status" role="status">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1.8"
              aria-hidden="true"
            >
              <circle cx="12" cy="12" r="8"></circle>
              <path d="m8.5 12 2.2 2.2 4.8-5" stroke-linecap="round" stroke-linejoin="round"></path>
            </svg>
            {{ statusLabel }}
          </p>
        </div>
        <BaseButton
          v-if="canReschedule"
          type="button"
          variant="secondary"
          class="appointment-detail-page__reschedule"
          @click="onOpenReschedule"
        >
          <span
            class="appointment-detail-page__button-icon"
            aria-hidden="true"
            v-html="factIcon('calendar')"
          />
          Reprogramar turno
        </BaseButton>
      </header>

      <div class="appointment-detail-page__body">
        <dl class="appointment-detail-page__facts">
          <div class="appointment-detail-page__fact">
            <span
              class="appointment-detail-page__fact-icon"
              aria-hidden="true"
              v-html="factIcon('time')"
            />
            <dt>Hora</dt>
            <dd>
              {{ timeRangeLabel }}
              <span v-if="barbershopTimezone">· Zona {{ barbershopTimezone }}</span>
            </dd>
          </div>
          <div class="appointment-detail-page__fact">
            <span
              class="appointment-detail-page__fact-icon"
              aria-hidden="true"
              v-html="factIcon('attendee')"
            />
            <dt>Persona atendida</dt>
            <dd>{{ detail.attendeeName }}</dd>
          </div>
          <div class="appointment-detail-page__fact">
            <span
              class="appointment-detail-page__fact-icon"
              aria-hidden="true"
              v-html="factIcon('service')"
            />
            <dt>Servicio</dt>
            <dd>
              {{ detail.serviceName }} · {{ detail.durationMinutes }} min · {{ detail.priceAmount }}
              {{ detail.currency }}
            </dd>
          </div>
          <div class="appointment-detail-page__fact">
            <span
              class="appointment-detail-page__fact-icon"
              aria-hidden="true"
              v-html="factIcon('barber')"
            />
            <dt>Barbero</dt>
            <dd>{{ detail.barberFullName }}</dd>
          </div>
          <div class="appointment-detail-page__fact">
            <span
              class="appointment-detail-page__fact-icon"
              aria-hidden="true"
              v-html="factIcon('attendee')"
            />
            <dt>Cliente que reservó</dt>
            <dd>{{ detail.customerFullName }}</dd>
          </div>
          <div
            v-if="detail.customerPhone || detail.customerEmail"
            class="appointment-detail-page__fact"
          >
            <span
              class="appointment-detail-page__fact-icon"
              aria-hidden="true"
              v-html="factIcon('contact')"
            />
            <dt>Contacto</dt>
            <dd>
              <span v-if="detail.customerPhone">{{ detail.customerPhone }}</span>
              <span v-if="detail.customerPhone && detail.customerEmail"> · </span>
              <span v-if="detail.customerEmail">{{ detail.customerEmail }}</span>
            </dd>
          </div>
          <div v-if="detail.customerNote" class="appointment-detail-page__fact">
            <span
              class="appointment-detail-page__fact-icon"
              aria-hidden="true"
              v-html="factIcon('note')"
            />
            <dt>Nota del cliente</dt>
            <dd>{{ detail.customerNote }}</dd>
          </div>
        </dl>

        <section
          class="appointment-detail-page__history"
          aria-labelledby="appointment-history-title"
        >
          <h2 id="appointment-history-title" class="appointment-detail-page__history-title">
            Historial
          </h2>

          <div
            v-if="historyStatus === 'loading'"
            class="appointment-detail-page__state"
            role="status"
            aria-live="polite"
          >
            <p>Cargando historial…</p>
          </div>

          <BaseAlert
            v-else-if="historyStatus === 'error'"
            variant="warning"
            title="No pudimos cargar el historial"
            role="alert"
          >
            Revisa tu conexión e inténtalo de nuevo.
            <template #action>
              <BaseButton type="button" variant="secondary" @click="onRetryHistory">
                Reintentar
              </BaseButton>
            </template>
          </BaseAlert>

          <template v-else>
            <p v-if="historyItems.length === 0" class="appointment-detail-page__empty">
              Sin eventos registrados todavía.
            </p>

            <ol v-else class="appointment-detail-page__history-list" aria-label="Eventos del turno">
              <li
                v-for="entry in historyItems"
                :key="entry.id"
                class="appointment-detail-page__history-item"
              >
                <span
                  class="appointment-detail-page__history-icon"
                  aria-hidden="true"
                  v-html="historyIcon(entry)"
                />
                <div class="appointment-detail-page__history-main">
                  <span class="appointment-detail-page__history-event">{{
                    eventLabel(entry)
                  }}</span>
                  <span class="appointment-detail-page__history-meta">
                    {{ occurredAtLabel(entry) }} · {{ entry.actorLabel }}
                  </span>
                </div>
                <p v-if="entry.reason" class="appointment-detail-page__history-reason">
                  {{ entry.reason }}
                </p>
                <ul
                  v-if="entry.changes.length > 0"
                  class="appointment-detail-page__history-changes"
                >
                  <li v-for="change in entry.changes" :key="change.fieldName">
                    {{ historyFieldLabel(change.fieldName) }}: {{ change.previousValue ?? '—' }} →
                    {{ change.newValue ?? '—' }}
                  </li>
                </ul>
              </li>
            </ol>

            <BaseButton
              v-if="historyNextCursor"
              type="button"
              variant="secondary"
              :loading="historyLoadingMore"
              @click="onLoadMoreHistory"
            >
              Cargar más
            </BaseButton>
          </template>
        </section>
      </div>

      <BaseDialog
        v-model="isRescheduleOpen"
        title="Reprogramar turno"
        size="xl"
        placement="bottom"
        content-class="appointment-detail-page__reschedule-dialog"
        @close="onRescheduleDialogClosed"
      >
        <form
          class="appointment-detail-page__dialog-form"
          novalidate
          @submit.prevent="onSubmitReschedule"
        >
          <BaseAlert
            v-if="rescheduleStatus === 'version-conflict'"
            variant="warning"
            title="Este turno cambió mientras lo editabas"
            role="alert"
          >
            <template #action>
              <BaseButton type="button" variant="secondary" @click="onReloadAfterConflict">
                Recargar
              </BaseButton>
            </template>
          </BaseAlert>
          <BaseAlert
            v-if="rescheduleStatus === 'invalid-state'"
            variant="warning"
            title="Este turno ya no se puede reprogramar"
            role="alert"
          >
            Su estado cambió mientras lo editabas.
            <template #action>
              <BaseButton type="button" variant="secondary" @click="onReloadAfterConflict">
                Recargar
              </BaseButton>
            </template>
          </BaseAlert>
          <BaseAlert
            v-if="rescheduleStatus === 'conflict'"
            variant="danger"
            title="No pudimos reprogramar el turno"
            role="alert"
          >
            {{ rescheduleErrorDetail }}
          </BaseAlert>
          <BaseAlert
            v-if="rescheduleStatus === 'validation-error'"
            variant="danger"
            title="Revisa la nueva fecha y hora"
            role="alert"
          >
            {{ rescheduleErrorDetail }}
          </BaseAlert>
          <BaseAlert
            v-if="rescheduleStatus === 'idempotency-conflict'"
            variant="danger"
            title="No pudimos completar el intento anterior"
            role="alert"
          >
            Inténtalo de nuevo.
          </BaseAlert>
          <BaseAlert
            v-if="rescheduleStatus === 'network-error'"
            variant="warning"
            title="No pudimos conectar"
            role="alert"
          >
            Revisa tu conexión e inténtalo de nuevo. No perdiste lo que elegiste.
          </BaseAlert>
          <BaseAlert
            v-if="rescheduleStatus === 'unexpected-error'"
            variant="danger"
            title="Ocurrió un error inesperado"
            role="alert"
          >
            Inténtalo de nuevo en unos segundos. No perdiste lo que elegiste.
          </BaseAlert>

          <div class="appointment-detail-page__dialog-grid">
            <p class="appointment-detail-page__dialog-current">
              <strong>Horario actual</strong>
              {{ timeRangeLabel }}
              <span v-if="barbershopTimezone">Zona {{ barbershopTimezone }}</span>
            </p>
            <BaseInput
              v-model="rescheduleDate"
              type="date"
              name="rescheduleDate"
              label="Nueva fecha"
              required
              :disabled="rescheduleStatus === 'saving'"
            />
            <BaseInput
              v-model="rescheduleTime"
              type="time"
              name="rescheduleTime"
              label="Nueva hora"
              required
              :disabled="rescheduleStatus === 'saving'"
            />
            <p v-if="rescheduleNewEndLabel" class="appointment-detail-page__dialog-preview">
              <strong>Vista previa</strong>
              {{ rescheduleTime }} – {{ rescheduleNewEndLabel }}
              <span v-if="barbershopTimezone">Zona {{ barbershopTimezone }}</span>
            </p>
          </div>

          <div class="appointment-detail-page__dialog-actions">
            <BaseButton type="button" variant="secondary" @click="isRescheduleOpen = false">
              Cancelar
            </BaseButton>
            <BaseButton
              type="submit"
              variant="primary"
              :loading="rescheduleStatus === 'saving'"
              :disabled="rescheduleStatus === 'saving'"
            >
              Confirmar
            </BaseButton>
          </div>
        </form>
      </BaseDialog>
    </template>
  </section>
</template>

<style scoped>
.appointment-detail-page {
  display: flex;
  position: relative;
  flex-direction: column;
  gap: var(--space-6);
  min-height: 100%;
  padding: 24px 40px 16px;
  background: var(--color-canvas);
}

.appointment-detail-page__body {
  display: grid;
  grid-template-columns: minmax(0, 1.78fr) minmax(260px, 0.82fr);
  gap: 48px;
  align-items: start;
}

.appointment-detail-page__back {
  margin: 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
}

.appointment-detail-page__back a {
  color: var(--color-action-primary);
  text-decoration: none;
}

.appointment-detail-page__state {
  padding: var(--space-6);
  color: var(--color-text-secondary);
}

.appointment-detail-page__empty {
  padding: var(--space-2) 0;
  color: var(--color-text-secondary);
}

.appointment-detail-page__hero {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  min-height: 74px;
}

.appointment-detail-page__avatar {
  display: grid;
  width: 72px;
  height: 72px;
  flex: 0 0 72px;
  place-items: center;
  overflow: hidden;
  background: var(--color-surface-strong);
  border: 2px solid var(--color-border-subtle);
  border-radius: 50%;
  color: var(--color-on-strong);
  font-family: var(--font-display);
  font-size: 24px;
}

.appointment-detail-page__identity {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: var(--space-1);
}

.appointment-detail-page__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h1);
  line-height: var(--font-size-h1-line);
  font-weight: var(--font-weight-h1);
  color: var(--color-text-primary);
}

.appointment-detail-page__status {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  margin: 0;
  color: var(--color-success-text);
  font-size: var(--font-size-body);
  line-height: var(--font-size-body-line);
}

.appointment-detail-page__status svg {
  width: 20px;
  height: 20px;
}

.appointment-detail-page__reschedule {
  flex: 0 0 auto;
  min-width: 194px;
}

.appointment-detail-page__button-icon {
  display: inline-flex;
  width: 20px;
  height: 20px;
  flex: 0 0 20px;
}

.appointment-detail-page__button-icon :deep(svg) {
  width: 100%;
  height: 100%;
}

.appointment-detail-page__facts {
  display: flex;
  flex-direction: column;
  gap: 0;
  margin: 0;
  padding-top: 1px;
  border-top: var(--border-width-normal) solid var(--color-border-subtle);
}

.appointment-detail-page__fact {
  display: grid;
  grid-template-columns: 40px minmax(132px, 0.62fr) minmax(0, 1.35fr);
  align-items: center;
  min-height: 56px;
  gap: var(--space-2);
  padding: 9px 10px;
  border-bottom: var(--border-width-normal) solid var(--color-border-subtle);
}

.appointment-detail-page__fact-icon {
  width: 24px;
  height: 24px;
  color: var(--color-surface-strong);
}

.appointment-detail-page__fact-icon :deep(svg) {
  width: 100%;
  height: 100%;
}

.appointment-detail-page__fact dt {
  font-family: var(--font-family-base);
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.appointment-detail-page__fact dd {
  margin: 0;
  font-family: var(--font-family-base);
  font-size: 14px;
  line-height: 20px;
  color: var(--color-text-primary);
}

.appointment-detail-page__history {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  min-width: 0;
  padding-left: 24px;
  border-left: var(--border-width-normal) solid var(--color-border-subtle);
}

.appointment-detail-page__history-title {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h2);
  line-height: var(--font-size-h2-line);
  font-weight: var(--font-weight-h2);
  color: var(--color-text-primary);
}

.appointment-detail-page__history-list {
  display: flex;
  flex-direction: column;
  gap: 0;
  padding: 0;
  margin: 0;
  list-style: none;
}

/* Línea de tiempo vertical (atlas: puntos y conexión entre eventos), sin
   cambiar el orden ni el contenido de cada evento. */
.appointment-detail-page__history-item {
  position: relative;
  min-height: 74px;
  padding: 0 0 var(--space-5) 48px;
  border-left: var(--border-width-normal) solid var(--color-border-subtle);
}

.appointment-detail-page__history-item:last-child {
  padding-bottom: 0;
  border-left-color: transparent;
}

.appointment-detail-page__history-icon {
  position: absolute;
  top: 0;
  left: -18px;
  display: grid;
  width: 36px;
  height: 36px;
  place-items: center;
  background-color: var(--color-brand-accent-surface);
  color: var(--color-surface-strong);
  border-radius: 999px;
}

.appointment-detail-page__history-icon :deep(svg) {
  width: 19px;
  height: 19px;
}

.appointment-detail-page__history-item:first-child .appointment-detail-page__history-icon {
  background-color: var(--color-surface-strong);
  color: var(--color-on-strong);
}

.appointment-detail-page__history-main {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.appointment-detail-page__history-event {
  font-family: var(--font-family-base);
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.appointment-detail-page__history-meta {
  font-family: var(--font-family-base);
  font-size: 13px;
  color: var(--color-text-secondary);
}

.appointment-detail-page__history-reason {
  margin: var(--space-1) 0 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  color: var(--color-text-secondary);
}

.appointment-detail-page__history-changes {
  margin: var(--space-1) 0 0;
  padding-left: var(--space-4);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  color: var(--color-text-secondary);
}

.appointment-detail-page__dialog-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.appointment-detail-page__dialog-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-4);
  align-items: end;
}

.appointment-detail-page__dialog-current,
.appointment-detail-page__dialog-preview {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin: 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
  color: var(--color-text-primary);
}

.appointment-detail-page__dialog-current strong,
.appointment-detail-page__dialog-preview strong {
  color: var(--color-text-primary);
  font-size: var(--font-size-caption);
}

.appointment-detail-page__dialog-current span,
.appointment-detail-page__dialog-preview span {
  color: var(--color-text-secondary);
}

.appointment-detail-page__dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  flex-wrap: wrap;
}

:global(.appointment-detail-page__reschedule-dialog .base-dialog__header) {
  padding: 14px 18px;
}

:global(.appointment-detail-page__reschedule-dialog .base-dialog__title) {
  font-family: var(--font-display);
  font-size: 20px;
  font-weight: 400;
}

:global(.appointment-detail-page__reschedule-dialog .base-dialog__content) {
  padding: 12px 18px 16px;
}

@media (max-width: 767px) {
  .appointment-detail-page {
    gap: var(--space-4);
    padding: 12px 12px 0;
  }

  .appointment-detail-page__back {
    display: none;
  }

  .appointment-detail-page__hero {
    gap: var(--space-3);
    min-height: 102px;
  }

  .appointment-detail-page__avatar {
    width: 58px;
    height: 58px;
    flex-basis: 58px;
    font-size: 19px;
  }

  .appointment-detail-page__title {
    font-size: 25px;
    line-height: 30px;
  }

  .appointment-detail-page__status {
    font-size: 14px;
  }

  .appointment-detail-page__reschedule {
    position: absolute;
    top: 123px;
    right: 12px;
    left: 12px;
    width: calc(100% - 24px);
  }

  .appointment-detail-page__body {
    grid-template-columns: 1fr;
    gap: var(--space-4);
    padding-top: 44px;
  }

  .appointment-detail-page__fact {
    grid-template-columns: 26px minmax(100px, 0.86fr) minmax(0, 1.25fr);
    min-height: 44px;
    gap: 4px;
    padding: 7px 0;
  }

  .appointment-detail-page__fact-icon {
    width: 20px;
    height: 20px;
  }

  .appointment-detail-page__fact dt,
  .appointment-detail-page__fact dd {
    font-size: 11px;
    line-height: 15px;
  }

  .appointment-detail-page__history {
    gap: var(--space-3);
    padding: 2px 0 0;
    border-top: var(--border-width-normal) solid var(--color-border-subtle);
    border-left: none;
  }

  .appointment-detail-page__history-title {
    font-size: 20px;
    line-height: 25px;
  }

  .appointment-detail-page__history-item {
    min-height: 64px;
    padding-left: 42px;
  }

  .appointment-detail-page__history-icon {
    left: -15px;
    width: 30px;
    height: 30px;
  }

  .appointment-detail-page__history-event {
    font-size: 12px;
  }

  .appointment-detail-page__history-meta,
  .appointment-detail-page__history-reason,
  .appointment-detail-page__history-changes {
    font-size: 11px;
    line-height: 15px;
  }

  .appointment-detail-page__dialog-grid {
    grid-template-columns: 1fr;
    gap: var(--space-3);
  }

  .appointment-detail-page__dialog-actions {
    display: grid;
    grid-template-columns: 1fr 1fr;
  }

  :global(.appointment-detail-page__reschedule-dialog .base-dialog__container) {
    max-height: calc(100vh - 20px);
  }
}

@media (min-width: 768px) and (max-width: 1023px) {
  .appointment-detail-page {
    padding: 24px;
  }

  .appointment-detail-page__body {
    grid-template-columns: minmax(0, 1.3fr) minmax(240px, 0.7fr);
    gap: 28px;
  }

  .appointment-detail-page__dialog-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
