<script setup lang="ts">
// Pantalla "Agenda de hoy" (HU-062): abre el panel privado mostrando
// cronológicamente los turnos de hoy de un único barbero, calculados en la
// zona IANA de la barbería. Sustituye el marcador de posición de HU-012
// (`auth/pages/PanelPage.vue`, retirado). Selector obligatorio de un
// barbero (DEC-074): nunca una vista consolidada de varios a la vez, ni un
// barbero implícito. Fuera de alcance a propósito: anterior/siguiente/
// selector de fecha (F-CITA-02), detalle de cita y cualquier acción sobre
// una cita existente (editar/cancelar/reprogramar/completar).
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { BaseAlert, BaseBadge, BaseButton } from '@/shared/ui'
import { formatFullDateInTimezone, formatTimeInTimezone } from '@/shared/time/formatInstant'
import {
  fetchBarberSummaries,
  fetchBarbershopTimezone,
  fetchDailyAgenda,
} from '../api/appointmentsApi'
import type { BarberSummary } from '../model/appointment'
import {
  APPOINTMENT_STATUS_BADGE_CLASS,
  APPOINTMENT_STATUS_LABELS,
  type DailyAgendaEntry,
} from '../model/dailyAgenda'

type PageStatus = 'loading' | 'ready' | 'load-error'
type AgendaStatus = 'idle' | 'loading' | 'ready' | 'error' | 'not-found'

const router = useRouter()

const pageStatus = ref<PageStatus>('loading')
const barbers = ref<BarberSummary[]>([])
const selectedBarberId = ref<string | null>(null)
// CA-062-01: zona IANA de la barbería, nunca la del dispositivo. null solo
// mientras carga o si la consulta falla; el encabezado se oculta hasta
// tenerla, en vez de mostrar una zona adivinada.
const barbershopTimezone = ref<string | null>(null)

const agendaStatus = ref<AgendaStatus>('idle')
const entries = ref<DailyAgendaEntry[]>([])

const selectedBarber = computed(
  () => barbers.value.find((b) => b.id === selectedBarberId.value) ?? null,
)

const todayLabel = computed(() =>
  barbershopTimezone.value ? formatFullDateInTimezone(barbershopTimezone.value) : null,
)

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

  if (barbers.value.length > 0) {
    await selectBarber(barbers.value[0]!.id)
  }
}

onMounted(loadPage)

function onRetryLoad() {
  void loadPage()
}

async function selectBarber(barberId: string) {
  selectedBarberId.value = barberId
  agendaStatus.value = 'loading'

  const outcome = await fetchDailyAgenda(barberId)
  // El barbero seleccionado pudo cambiar mientras la solicitud estaba en
  // vuelo (cambio rápido en el selector): descarta una respuesta obsoleta.
  if (selectedBarberId.value !== barberId) return

  switch (outcome.kind) {
    case 'success':
      entries.value = [...outcome.items].sort((a, b) => a.startsAt.localeCompare(b.startsAt))
      agendaStatus.value = 'ready'
      return
    case 'not-found':
      agendaStatus.value = 'not-found'
      return
    default:
      agendaStatus.value = 'error'
  }
}

function onBarberSelectChange(event: Event) {
  const barberId = (event.target as HTMLSelectElement).value
  void selectBarber(barberId)
}

function onRetryAgenda() {
  if (selectedBarberId.value) void selectBarber(selectedBarberId.value)
}

function goToNewAppointment() {
  void router.push({ name: 'agenda-nuevo-turno' })
}

function statusLabel(entry: DailyAgendaEntry): string {
  return APPOINTMENT_STATUS_LABELS[entry.status]
}

function statusBadgeClass(entry: DailyAgendaEntry): string {
  return APPOINTMENT_STATUS_BADGE_CLASS[entry.status]
}

function entryTime(entry: DailyAgendaEntry): string {
  if (!barbershopTimezone.value) return ''
  return formatTimeInTimezone(entry.startsAt, barbershopTimezone.value)
}
</script>

<template>
  <section class="daily-agenda-page" aria-labelledby="daily-agenda-page-title">
    <header class="daily-agenda-page__header">
      <div>
        <h1 id="daily-agenda-page-title" class="daily-agenda-page__title">Agenda de hoy</h1>
        <p v-if="todayLabel" class="daily-agenda-page__date">
          {{ todayLabel }} · Zona {{ barbershopTimezone }}
        </p>
      </div>
      <BaseButton
        v-if="pageStatus === 'ready' && barbers.length > 0"
        type="button"
        variant="primary"
        @click="goToNewAppointment"
      >
        Nuevo turno
      </BaseButton>
    </header>

    <div
      v-if="pageStatus === 'loading'"
      class="daily-agenda-page__state"
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
      <p v-if="barbers.length === 0" class="daily-agenda-page__empty">
        Aún no tienes barberos registrados. Agrega uno en la sección "Barberos" para ver su agenda.
      </p>

      <template v-else>
        <div class="daily-agenda-page__picker">
          <label for="daily-agenda-barber-select" class="daily-agenda-page__label">Barbero</label>
          <select
            id="daily-agenda-barber-select"
            class="daily-agenda-page__select"
            :value="selectedBarberId ?? ''"
            @change="onBarberSelectChange"
          >
            <option v-for="barber in barbers" :key="barber.id" :value="barber.id">
              {{ barber.fullName }}
            </option>
          </select>
        </div>

        <div
          v-if="agendaStatus === 'loading'"
          class="daily-agenda-page__state"
          role="status"
          aria-live="polite"
        >
          <p>Cargando la agenda de {{ selectedBarber?.fullName }}…</p>
        </div>

        <BaseAlert
          v-else-if="agendaStatus === 'not-found'"
          variant="warning"
          title="Este barbero ya no está disponible"
          role="alert"
        >
          Elige otro barbero en la lista.
        </BaseAlert>

        <BaseAlert
          v-else-if="agendaStatus === 'error'"
          variant="warning"
          title="No pudimos cargar la agenda de este barbero"
          role="alert"
        >
          Revisa tu conexión e inténtalo de nuevo.
          <template #action>
            <BaseButton type="button" variant="secondary" @click="onRetryAgenda">
              Reintentar
            </BaseButton>
          </template>
        </BaseAlert>

        <template v-else-if="agendaStatus === 'ready'">
          <p v-if="entries.length === 0" class="daily-agenda-page__empty">
            No hay turnos para {{ selectedBarber?.fullName }} hoy.
          </p>

          <ul
            v-else
            class="daily-agenda-page__list"
            :aria-label="`Turnos de hoy de ${selectedBarber?.fullName}`"
          >
            <li
              v-for="entry in entries"
              :key="entry.id"
              class="daily-agenda-page__item"
              :class="{ 'daily-agenda-page__item--terminal': entry.status !== 'confirmed' }"
            >
              <div class="daily-agenda-page__item-main">
                <span class="daily-agenda-page__item-time">{{ entryTime(entry) }}</span>
                <span class="daily-agenda-page__item-name">{{ entry.attendeeName }}</span>
                <span class="daily-agenda-page__item-service">{{ entry.serviceName }}</span>
              </div>
              <BaseBadge :class="statusBadgeClass(entry)" size="sm" dot :label="statusLabel(entry)">
                {{ statusLabel(entry) }}
              </BaseBadge>
            </li>
          </ul>
        </template>
      </template>
    </template>
  </section>
</template>

<style scoped>
.daily-agenda-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  max-width: 720px;
  padding: var(--space-4);
  margin: 0 auto;
}

.daily-agenda-page__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.daily-agenda-page__title {
  margin: 0;
  font-size: var(--font-size-heading-lg);
  color: var(--color-text-primary);
}

.daily-agenda-page__date {
  margin: var(--space-1) 0 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  color: var(--color-text-secondary);
}

.daily-agenda-page__state {
  padding: var(--space-4);
  color: var(--color-text-secondary);
}

.daily-agenda-page__empty {
  padding: var(--space-4);
  color: var(--color-text-secondary);
}

.daily-agenda-page__picker {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.daily-agenda-page__label {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  font-weight: 500;
  color: var(--color-text-primary);
}

.daily-agenda-page__select {
  min-height: 44px;
  padding: var(--space-2) var(--space-3);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-text-primary);
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: var(--radius-md);
}

.daily-agenda-page__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding: 0;
  margin: 0;
  list-style: none;
}

.daily-agenda-page__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  min-height: 44px;
  padding: var(--space-4);
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  flex-wrap: wrap;
}

/* Atenuada, nunca oculta ni borrada: una cita terminal sigue siendo un
   hecho del día (RN-CIT-04). */
.daily-agenda-page__item--terminal {
  opacity: 0.72;
}

.daily-agenda-page__item-main {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: var(--space-2);
  min-width: 0;
}

.daily-agenda-page__item-time {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  font-weight: 600;
  color: var(--color-text-primary);
}

.daily-agenda-page__item-name {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-text-primary);
}

.daily-agenda-page__item-service {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  color: var(--color-text-secondary);
}
</style>
