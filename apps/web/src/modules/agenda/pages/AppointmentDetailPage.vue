<script setup lang="ts">
// Detalle e historial de un turno (HU-064): persona atendida, cliente que
// reservó con su contacto opcional y nota, servicio snapshot (DEC-004),
// precio, estado e historial inmutable paginado, siguiendo la plantilla P0
// "Detalle de turno" (estandar-diseno-visual.md §10). "Volver" conserva la
// fecha/barbero de origen (HU-063) mediante `route.query`, capturados al
// entrar. Fuera de alcance a propósito: cualquier edición, reprogramación,
// cancelación o cambio de estado (ninguna acción se renderiza).
import { computed, onMounted, ref } from 'vue'
import { useRoute, type LocationQueryRaw } from 'vue-router'
import { BaseAlert, BaseBadge, BaseButton } from '@/shared/ui'
import { formatInstantInTimezone } from '@/shared/time/formatInstant'
import {
  fetchAppointmentDetail,
  fetchAppointmentHistory,
  fetchBarbershopTimezone,
} from '../api/appointmentsApi'
import type { AppointmentDetail, HistoryEntry } from '../model/appointmentDetail'
import { HISTORY_EVENT_LABELS, historyFieldLabel } from '../model/appointmentDetail'
import { APPOINTMENT_STATUS_BADGE_CLASS, APPOINTMENT_STATUS_LABELS } from '../model/dailyAgenda'

type PageStatus = 'loading' | 'ready' | 'not-found' | 'error'
type HistoryStatus = 'loading' | 'ready' | 'error'

const route = useRoute()

const appointmentId = computed(() => String(route.params.appointmentId ?? ''))

// backQuery se captura UNA sola vez al entrar (no reactivo a cambios
// posteriores de route.query de esta misma pantalla, que no tiene
// ninguno): exactamente la fecha/barbero desde los que se abrió el turno
// (HU-063), para que "Volver" regrese al mismo lugar de la agenda.
const backQuery: LocationQueryRaw = {
  ...(typeof route.query.date === 'string' ? { date: route.query.date } : {}),
  ...(typeof route.query.barberId === 'string' ? { barberId: route.query.barberId } : {}),
}

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
const statusBadgeClass = computed(() =>
  detail.value ? APPOINTMENT_STATUS_BADGE_CLASS[detail.value.status] : '',
)

const timeRangeLabel = computed(() => {
  if (!detail.value || !barbershopTimezone.value) return ''
  const start = formatInstantInTimezone(detail.value.startsAt, barbershopTimezone.value)
  const end = new Intl.DateTimeFormat('es-CO', {
    timeZone: barbershopTimezone.value,
    timeStyle: 'short',
  }).format(new Date(detail.value.endsAt))
  return `${start} – ${end}`
})

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
      <header class="appointment-detail-page__header">
        <h1 id="appointment-detail-page-title" class="appointment-detail-page__title">
          {{ detail.attendeeName }}
        </h1>
        <BaseBadge :class="statusBadgeClass" size="sm" dot :label="statusLabel">
          {{ statusLabel }}
        </BaseBadge>
      </header>

      <dl class="appointment-detail-page__facts">
        <div class="appointment-detail-page__fact">
          <dt>Hora</dt>
          <dd>
            {{ timeRangeLabel }}
            <span v-if="barbershopTimezone">· Zona {{ barbershopTimezone }}</span>
          </dd>
        </div>
        <div class="appointment-detail-page__fact">
          <dt>Persona atendida</dt>
          <dd>{{ detail.attendeeName }}</dd>
        </div>
        <div class="appointment-detail-page__fact">
          <dt>Servicio</dt>
          <dd>
            {{ detail.serviceName }} · {{ detail.durationMinutes }} min · {{ detail.priceAmount }}
            {{ detail.currency }}
          </dd>
        </div>
        <div class="appointment-detail-page__fact">
          <dt>Barbero</dt>
          <dd>{{ detail.barberFullName }}</dd>
        </div>
        <div class="appointment-detail-page__fact">
          <dt>Cliente que reservó</dt>
          <dd>{{ detail.customerFullName }}</dd>
        </div>
        <div
          v-if="detail.customerPhone || detail.customerEmail"
          class="appointment-detail-page__fact"
        >
          <dt>Contacto</dt>
          <dd>
            <span v-if="detail.customerPhone">{{ detail.customerPhone }}</span>
            <span v-if="detail.customerPhone && detail.customerEmail"> · </span>
            <span v-if="detail.customerEmail">{{ detail.customerEmail }}</span>
          </dd>
        </div>
        <div v-if="detail.customerNote" class="appointment-detail-page__fact">
          <dt>Nota del cliente</dt>
          <dd>{{ detail.customerNote }}</dd>
        </div>
      </dl>

      <section class="appointment-detail-page__history" aria-labelledby="appointment-history-title">
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
              <div class="appointment-detail-page__history-main">
                <span class="appointment-detail-page__history-event">{{ eventLabel(entry) }}</span>
                <span class="appointment-detail-page__history-meta">
                  {{ occurredAtLabel(entry) }} · {{ entry.actorLabel }}
                </span>
              </div>
              <p v-if="entry.reason" class="appointment-detail-page__history-reason">
                {{ entry.reason }}
              </p>
              <ul v-if="entry.changes.length > 0" class="appointment-detail-page__history-changes">
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
    </template>
  </section>
</template>

<style scoped>
.appointment-detail-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  max-width: 720px;
  padding: var(--space-4);
  margin: 0 auto;
}

.appointment-detail-page__back {
  margin: 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
}

.appointment-detail-page__back a {
  color: var(--color-action-primary);
}

.appointment-detail-page__state {
  padding: var(--space-4);
  color: var(--color-text-secondary);
}

.appointment-detail-page__empty {
  padding: var(--space-2) 0;
  color: var(--color-text-secondary);
}

.appointment-detail-page__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.appointment-detail-page__title {
  margin: 0;
  font-size: var(--font-size-heading-lg);
  color: var(--color-text-primary);
}

.appointment-detail-page__facts {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  margin: 0;
}

.appointment-detail-page__fact {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.appointment-detail-page__fact dt {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  font-weight: 500;
  color: var(--color-text-secondary);
}

.appointment-detail-page__fact dd {
  margin: 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-text-primary);
}

.appointment-detail-page__history {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding-top: var(--space-4);
  border-top: var(--border-width-normal) solid var(--color-border-subtle);
}

.appointment-detail-page__history-title {
  margin: 0;
  font-size: var(--font-size-heading-md);
  color: var(--color-text-primary);
}

.appointment-detail-page__history-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding: 0;
  margin: 0;
  list-style: none;
}

.appointment-detail-page__history-item {
  padding: var(--space-3);
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: var(--radius-md);
}

.appointment-detail-page__history-main {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--space-2);
}

.appointment-detail-page__history-event {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  font-weight: 600;
  color: var(--color-text-primary);
}

.appointment-detail-page__history-meta {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
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
</style>
