<script setup lang="ts">
// Pantalla "Agenda de hoy" (HU-062) + navegación por fecha (HU-063): abre
// el panel privado mostrando cronológicamente los turnos de un día civil
// de un único barbero, calculado en la zona IANA de la barbería. Anterior/
// selector de fecha/siguiente (F-CITA-02) conservan posición estable
// (estandar-diseno-visual.md §11.2) y viven en `route.query` (`date`,
// `barberId`): sin estado global, recargable y con atrás/adelante reales.
// Selector obligatorio de un barbero (DEC-074): nunca una vista
// consolidada. Fuera de alcance a propósito: detalle de cita y cualquier
// acción sobre una cita existente (editar/cancelar/reprogramar/completar).
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter, type LocationQueryRaw } from 'vue-router'
import { BaseAlert, BaseBadge, BaseButton, BaseInput } from '@/shared/ui'
import { formatTimeInTimezone } from '@/shared/time/formatInstant'
import {
  formatCivilDateFull,
  getCivilDateInTimezone,
  isCivilDateString,
  shiftCivilDate,
} from '@/shared/time/civilDate'
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
// 'updating' (CA-063): ya hubo una agenda confirmada antes (para otra
// fecha/barbero) y esta pantalla la conserva visible mientras llega la
// siguiente; 'loading' es solo la primera carga, sin nada que conservar.
type AgendaStatus = 'idle' | 'loading' | 'updating' | 'ready' | 'error' | 'not-found'

const route = useRoute()
const router = useRouter()

const pageStatus = ref<PageStatus>('loading')
const barbers = ref<BarberSummary[]>([])
// CA-063: zona IANA de la barbería, nunca la del dispositivo. null solo
// mientras carga o si la consulta falla; sin ella la pantalla no puede
// resolver "hoy" ni ofrecer navegación por fecha (mismo criterio degradado
// que HU-062: la agenda del día del servidor sigue siendo legible, solo se
// oculta la fecha/zona en pantalla y los controles de navegación).
const barbershopTimezone = ref<string | null>(null)

// selectedBarberId/selectedDate reflejan la última selección ya sincronizada
// con route.query (fuente de verdad): se fijan justo antes de disparar la
// carga, nunca antes, para que la plantilla no muestre una selección que la
// URL todavía no confirma.
const selectedBarberId = ref<string | null>(null)
const selectedDate = ref<string | null>(null)

const agendaStatus = ref<AgendaStatus>('idle')
const entries = ref<DailyAgendaEntry[]>([])
const hasLoadedEntriesOnce = ref(false)

let agendaRequestSeq = 0

const selectedBarber = computed(
  () => barbers.value.find((b) => b.id === selectedBarberId.value) ?? null,
)

const selectedDateLabel = computed(() =>
  selectedDate.value ? formatCivilDateFull(selectedDate.value) : null,
)

const canNavigateDates = computed(() => barbershopTimezone.value !== null)

// isViewingToday decide el texto del estado vacío ("hoy" vs. una fecha
// explícita): sin zona conocida se asume "hoy" (mismo criterio degradado
// que HU-062, que nunca navegaba y siempre mostraba "hoy").
const isViewingToday = computed(() => {
  if (!selectedDate.value || !barbershopTimezone.value) return true
  return selectedDate.value === getCivilDateInTimezone(barbershopTimezone.value)
})

const emptyStateDateText = computed(() =>
  isViewingToday.value ? 'hoy' : `el ${selectedDateLabel.value}`,
)

function requestedBarberIdFromRoute(): string | null {
  const value = route.query.barberId
  return typeof value === 'string' && value.length > 0 ? value : null
}

function requestedDateFromRoute(): string | null {
  const value = route.query.date
  return typeof value === 'string' && isCivilDateString(value) ? value : null
}

// withQuery combina route.query con overrides; una clave con valor
// undefined/vacío se elimina en vez de quedar como "date=undefined"
// literal en la URL.
function withQuery(overrides: Record<string, string | undefined>): LocationQueryRaw {
  const merged: Record<string, string> = {}
  for (const [key, value] of Object.entries({ ...route.query, ...overrides })) {
    if (typeof value === 'string' && value.length > 0) merged[key] = value
  }
  return merged
}

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

  if (barbers.value.length > 0) await syncFromRoute()
}

onMounted(loadPage)

function onRetryLoad() {
  void loadPage()
}

// syncFromRoute (CA-063) es el único punto que traduce route.query a una
// carga de agenda: se llama una vez al abrir la pantalla (tras resolver
// barberos/zona) y en cada cambio posterior de `route.query` (recarga,
// atrás/adelante del navegador, o un router.push propio de esta pantalla).
// Si la URL no trae una selección válida la normaliza con un único
// router.replace (nunca crea una entrada de historial por abrir /panel);
// ese replace vuelve a disparar este mismo watcher, ya con la URL
// normalizada, así que nunca hay una segunda petición duplicada.
async function syncFromRoute() {
  // Guarda de ruta: esta pantalla solo posee `route.query` mientras es la
  // ruta activa ('panel'). Navegar a otra hija del panel (p. ej. "Nuevo
  // turno") no la desmonta en las pruebas de componente (que la montan
  // fuera de un <router-view>), así que sin esta guarda el watcher de
  // route.query seguiría reaccionando a una ruta que ya no es esta
  // pantalla.
  if (route.name !== 'panel' || barbers.value.length === 0) return

  const requestedBarberId = requestedBarberIdFromRoute()
  const resolvedBarberId =
    requestedBarberId && barbers.value.some((b) => b.id === requestedBarberId)
      ? requestedBarberId
      : barbers.value[0]!.id

  const resolvedDate = barbershopTimezone.value
    ? (requestedDateFromRoute() ?? getCivilDateInTimezone(barbershopTimezone.value))
    : null

  // La comparación usa el valor CRUDO de route.query.date (no el ya
  // validado por requestedDateFromRoute): una fecha inválida en la URL
  // ("2026-13-40", "hoy") nunca es igual a `resolvedDate` y por eso también
  // se normaliza, no solo una fecha ausente.
  const rawDateInUrl = typeof route.query.date === 'string' ? route.query.date : null
  const needsNormalization = requestedBarberId !== resolvedBarberId || rawDateInUrl !== resolvedDate

  if (needsNormalization) {
    await router.replace({
      query: withQuery({ barberId: resolvedBarberId, date: resolvedDate ?? undefined }),
    })
    return
  }

  selectedBarberId.value = resolvedBarberId
  selectedDate.value = resolvedDate
  await loadAgenda(resolvedBarberId, resolvedDate)
}

watch(
  () => [route.query.date, route.query.barberId],
  () => void syncFromRoute(),
)

async function loadAgenda(barberId: string, date: string | null) {
  const requestId = ++agendaRequestSeq
  agendaStatus.value = hasLoadedEntriesOnce.value ? 'updating' : 'loading'

  const outcome = await fetchDailyAgenda(barberId, date ?? undefined)

  // Una selección posterior (fecha, barbero o ambas) ya reemplazó el
  // destino vigente: esta respuesta llegó fuera de orden y se descarta sin
  // tocar el contenido ya mostrado (nunca sustituye la selección vigente
  // con una respuesta obsoleta).
  if (requestId !== agendaRequestSeq) return

  switch (outcome.kind) {
    case 'success':
      entries.value = [...outcome.items].sort((a, b) => a.startsAt.localeCompare(b.startsAt))
      agendaStatus.value = 'ready'
      hasLoadedEntriesOnce.value = true
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
  void router.push({ query: withQuery({ barberId }) })
}

function onRetryAgenda() {
  if (selectedBarberId.value) void loadAgenda(selectedBarberId.value, selectedDate.value)
}

function goToDate(newDate: string) {
  void router.push({ query: withQuery({ date: newDate }) })
}

function goToPreviousDay() {
  if (selectedDate.value) goToDate(shiftCivilDate(selectedDate.value, -1))
}

function goToNextDay() {
  if (selectedDate.value) goToDate(shiftCivilDate(selectedDate.value, 1))
}

function onDateInputChange(event: Event) {
  const value = (event.target as HTMLInputElement).value
  if (isCivilDateString(value)) goToDate(value)
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
        <p v-if="selectedDateLabel" class="daily-agenda-page__date">
          {{ selectedDateLabel }} · Zona {{ barbershopTimezone }}
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
        <div class="daily-agenda-page__controls">
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

          <div class="daily-agenda-page__date-nav">
            <BaseButton
              type="button"
              variant="secondary"
              :disabled="!canNavigateDates"
              aria-label="Día anterior"
              @click="goToPreviousDay"
            >
              Anterior
            </BaseButton>
            <BaseInput
              type="date"
              label="Fecha"
              class="daily-agenda-page__date-input"
              :model-value="selectedDate ?? ''"
              :disabled="!canNavigateDates"
              @change="onDateInputChange"
            />
            <BaseButton
              type="button"
              variant="secondary"
              :disabled="!canNavigateDates"
              aria-label="Día siguiente"
              @click="goToNextDay"
            >
              Siguiente
            </BaseButton>
          </div>
        </div>

        <div
          v-if="agendaStatus === 'loading'"
          class="daily-agenda-page__state"
          role="status"
          aria-live="polite"
        >
          <p>Cargando la agenda de {{ selectedBarber?.fullName }}…</p>
        </div>

        <template v-else>
          <p
            v-if="agendaStatus === 'updating'"
            class="daily-agenda-page__updating"
            role="status"
            aria-live="polite"
          >
            Actualizando…
          </p>

          <BaseAlert
            v-if="agendaStatus === 'not-found'"
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

          <template
            v-if="
              agendaStatus === 'ready' || (hasLoadedEntriesOnce && agendaStatus !== 'not-found')
            "
          >
            <p v-if="entries.length === 0" class="daily-agenda-page__empty">
              No hay turnos para {{ selectedBarber?.fullName }} {{ emptyStateDateText }}.
            </p>

            <ul
              v-else
              class="daily-agenda-page__list"
              :aria-label="`Turnos de ${selectedBarber?.fullName}`"
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
                <BaseBadge
                  :class="statusBadgeClass(entry)"
                  size="sm"
                  dot
                  :label="statusLabel(entry)"
                >
                  {{ statusLabel(entry) }}
                </BaseBadge>
              </li>
            </ul>
          </template>
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

.daily-agenda-page__updating {
  margin: 0;
  padding: var(--space-2) 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  color: var(--color-text-secondary);
}

.daily-agenda-page__empty {
  padding: var(--space-4);
  color: var(--color-text-secondary);
}

.daily-agenda-page__controls {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
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

/* HU-063: anterior/fecha/siguiente conservan posiciones estables
   (estandar-diseno-visual.md §11.2), sin reflow al cambiar de estado. */
.daily-agenda-page__date-nav {
  display: flex;
  align-items: flex-end;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.daily-agenda-page__date-input {
  flex: 1 1 180px;
  min-width: 160px;
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
