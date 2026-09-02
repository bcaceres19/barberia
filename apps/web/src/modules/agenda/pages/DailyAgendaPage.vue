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
  minutesIntoCivilDate,
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

// Línea temporal horizontal de escritorio (estandar-diseno-visual.md §11.2,
// especificacion-frontend-nava.md §7.2): "no es fuente de disponibilidad",
// solo posiciona visualmente las fichas ya devueltas por el servidor. La
// lista cronológica de arriba sigue siendo la fuente accesible equivalente;
// esta franja se marca aria-hidden y sus enlaces quedan fuera del tabulado
// (mismo turno, doble forma de abrir el detalle sería redundante para
// teclado/lector).
const MIN_TIMELINE_SPAN_MINUTES = 4 * 60

const timelineBounds = computed(() => {
  if (!barbershopTimezone.value || !selectedDate.value || entries.value.length === 0) return null
  const tz = barbershopTimezone.value
  const date = selectedDate.value
  const starts = entries.value.map((e) => minutesIntoCivilDate(e.startsAt, date, tz))
  const ends = entries.value.map((e) => minutesIntoCivilDate(e.endsAt, date, tz))
  const startHour = Math.floor(Math.min(...starts) / 60) * 60
  const endHour = Math.ceil(Math.max(...ends) / 60) * 60
  const span = Math.max(endHour - startHour, MIN_TIMELINE_SPAN_MINUTES)
  return { start: startHour, end: startHour + span }
})

const timelineTicks = computed(() => {
  const bounds = timelineBounds.value
  if (!bounds) return []
  const ticks: { minute: number; label: string }[] = []
  for (let minute = bounds.start; minute <= bounds.end; minute += 60) {
    const hour = Math.floor(minute / 60) % 24
    ticks.push({ minute, label: `${String(hour).padStart(2, '0')}:00` })
  }
  return ticks
})

function timelinePercent(minute: number): string {
  const bounds = timelineBounds.value
  if (!bounds) return '0%'
  return `${((minute - bounds.start) / (bounds.end - bounds.start)) * 100}%`
}

function timelineSlipStyle(entry: DailyAgendaEntry): { left: string; width: string } {
  const bounds = timelineBounds.value
  if (!bounds || !selectedDate.value || !barbershopTimezone.value)
    return { left: '0%', width: '0%' }
  const start = minutesIntoCivilDate(entry.startsAt, selectedDate.value, barbershopTimezone.value)
  const end = minutesIntoCivilDate(entry.endsAt, selectedDate.value, barbershopTimezone.value)
  const span = bounds.end - bounds.start
  // Ancho mínimo visual del 4%: una ficha muy corta sigue siendo legible en
  // la línea de tiempo sin que eso cambie su duración real.
  const width = Math.max(((end - start) / span) * 100, 4)
  return { left: timelinePercent(start), width: `${width}%` }
}

// "Ahora" (estandar-diseno-visual.md §7.2, §9.3): decorativo respecto al
// estado del turno, nunca cambia `confirmed`. Solo se calcula al montar/
// recargar la agenda (sin reloj en vivo): un dato de referencia visual, no
// una fuente de disponibilidad que necesite exactitud al segundo.
const nowMarkerPercent = computed(() => {
  const bounds = timelineBounds.value
  if (!bounds || !selectedDate.value || !barbershopTimezone.value || !isViewingToday.value) {
    return null
  }
  const nowMinute = minutesIntoCivilDate(
    new Date().toISOString(),
    selectedDate.value,
    barbershopTimezone.value,
  )
  if (nowMinute < bounds.start || nowMinute > bounds.end) return null
  return timelinePercent(nowMinute)
})
</script>

<template>
  <section class="daily-agenda-page" aria-labelledby="daily-agenda-page-title">
    <header class="daily-agenda-page__header">
      <div>
        <h1 id="daily-agenda-page-title" class="daily-agenda-page__title">Agenda</h1>
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

            <template v-else>
              <!-- Línea temporal horizontal de escritorio: presentación
                   visual adicional del mismo turno de la lista de abajo, que
                   sigue siendo la fuente accesible equivalente (§7.2, §8.1).
                   aria-hidden + tabindex="-1" evitan una segunda forma
                   redundante de llegar al mismo detalle para teclado/lector. -->
              <div v-if="timelineBounds" class="daily-agenda-page__timeline" aria-hidden="true">
                <div class="daily-agenda-page__timeline-track">
                  <div
                    v-for="tick in timelineTicks"
                    :key="tick.minute"
                    class="daily-agenda-page__timeline-tick"
                    :style="{ left: timelinePercent(tick.minute) }"
                  >
                    <span class="daily-agenda-page__timeline-tick-label">{{ tick.label }}</span>
                  </div>

                  <div
                    v-if="nowMarkerPercent"
                    class="daily-agenda-page__timeline-now"
                    :style="{ left: nowMarkerPercent }"
                  >
                    <span class="daily-agenda-page__timeline-now-label">Ahora</span>
                  </div>

                  <RouterLink
                    v-for="entry in entries"
                    :key="`timeline-${entry.id}`"
                    tabindex="-1"
                    class="daily-agenda-page__timeline-slip"
                    :class="{
                      'daily-agenda-page__timeline-slip--terminal': entry.status !== 'confirmed',
                    }"
                    :style="timelineSlipStyle(entry)"
                    :to="{
                      name: 'agenda-detalle-turno',
                      params: { appointmentId: entry.id },
                      query: withQuery({}),
                    }"
                  >
                    <span class="daily-agenda-page__timeline-slip-time">{{
                      entryTime(entry)
                    }}</span>
                    <span class="daily-agenda-page__timeline-slip-name">{{
                      entry.attendeeName
                    }}</span>
                  </RouterLink>
                </div>
              </div>

              <ul
                class="daily-agenda-page__list"
                :aria-label="`Turnos de ${selectedBarber?.fullName}`"
              >
                <li
                  v-for="entry in entries"
                  :key="entry.id"
                  class="daily-agenda-page__item"
                  :class="{ 'daily-agenda-page__item--terminal': entry.status !== 'confirmed' }"
                >
                  <RouterLink
                    class="daily-agenda-page__item-main"
                    :to="{
                      name: 'agenda-detalle-turno',
                      params: { appointmentId: entry.id },
                      query: withQuery({}),
                    }"
                  >
                    <span class="daily-agenda-page__item-time">{{ entryTime(entry) }}</span>
                    <span class="daily-agenda-page__item-name">{{ entry.attendeeName }}</span>
                    <span class="daily-agenda-page__item-service">{{ entry.serviceName }}</span>
                  </RouterLink>
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
  font-size: var(--font-size-h1);
  line-height: var(--font-size-h1-line);
  font-weight: var(--font-weight-h1);
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
  /* Ficha (§6.1, §7.2): al menos 64px en móvil, no el objetivo táctil
     mínimo de 44px. */
  min-height: 64px;
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

/* La fila abre el detalle del turno (HU-064) por su bloque principal
   (hora/persona/servicio), no por toda la tarjeta: el badge de estado queda
   fuera del enlace, sin volverse un control ambiguo. */
.daily-agenda-page__item-main {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: var(--space-2);
  min-width: 0;
  min-height: 44px;
  color: inherit;
  text-decoration: none;
  border-radius: var(--radius-sm);
  outline: none;
}

.daily-agenda-page__item-main:hover {
  text-decoration: underline;
}

.daily-agenda-page__item-main:focus-visible {
  box-shadow:
    0 0 0 2px var(--color-surface),
    0 0 0 4px var(--color-focus);
}

.daily-agenda-page__item-time {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: var(--color-text-primary);
}

/* Persona atendida como texto principal de la ficha (§6.1). */
.daily-agenda-page__item-name {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  font-weight: 600;
  color: var(--color-text-primary);
}

.daily-agenda-page__item-service {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  color: var(--color-text-secondary);
}

/* Línea temporal horizontal de escritorio (§7.2, §11.2): oculta por defecto,
   visible solo desde 1024px, donde hay espacio real para un eje legible. */
.daily-agenda-page__timeline {
  display: none;
}

@media (min-width: 1024px) {
  .daily-agenda-page__timeline {
    display: block;
    overflow-x: auto;
  }

  .daily-agenda-page__timeline-track {
    position: relative;
    height: 104px;
    margin-top: var(--space-6);
    padding: 0 var(--space-2);
    background-color: var(--color-surface);
    border: var(--border-width-normal) solid var(--color-border-subtle);
    border-radius: var(--radius-md);
  }

  .daily-agenda-page__timeline-tick {
    position: absolute;
    top: 0;
    bottom: 0;
    border-left: var(--border-width-normal) solid var(--color-border-subtle);
  }

  .daily-agenda-page__timeline-tick-label {
    position: absolute;
    top: calc(-1 * var(--space-6));
    left: var(--space-1);
    white-space: nowrap;
    font-size: var(--font-size-caption);
    font-variant-numeric: tabular-nums;
    color: var(--color-text-secondary);
  }

  /* Latón oscuro (foco/énfasis secundario, §4.1): el marcador "Ahora" no es
     una acción primaria y no reutiliza el color de acción. */
  .daily-agenda-page__timeline-now {
    position: absolute;
    top: 0;
    bottom: 0;
    z-index: 1;
    border-left: var(--border-width-emphasis) solid var(--color-focus);
  }

  .daily-agenda-page__timeline-now-label {
    position: absolute;
    top: calc(-1 * var(--space-6));
    left: var(--space-1);
    white-space: nowrap;
    font-size: var(--font-size-caption);
    font-weight: 600;
    color: var(--color-focus);
  }

  .daily-agenda-page__timeline-slip {
    position: absolute;
    top: var(--space-3);
    bottom: var(--space-3);
    display: flex;
    min-width: 64px;
    flex-direction: column;
    justify-content: center;
    gap: 2px;
    overflow: hidden;
    padding: var(--space-1) var(--space-2);
    background-color: var(--color-action-soft);
    border: var(--border-width-normal) solid var(--color-action-soft-border);
    border-radius: var(--radius-sm);
    color: var(--color-action-primary);
    text-decoration: none;
  }

  .daily-agenda-page__timeline-slip--terminal {
    background-color: var(--color-inactive-surface);
    border-color: var(--color-inactive-border);
    color: var(--color-inactive-text);
  }

  .daily-agenda-page__timeline-slip-time {
    font-size: var(--font-size-caption);
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }

  .daily-agenda-page__timeline-slip-name {
    overflow: hidden;
    font-size: var(--font-size-caption);
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>
