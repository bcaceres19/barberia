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
//
// Fidelidad visual con el atlas panel-agenda-eventos (issue #189): esta
// revisión es exclusivamente de composición/color/jerarquía/estado visual,
// nunca de comportamiento. Los comentarios que citan un evento del atlas
// (docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/
// panel-agenda-eventos/README.md) documentan a qué panel responde cada
// bloque.
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter, type LocationQueryRaw } from 'vue-router'
import { BaseAlert, BaseBadge, BaseButton, BaseInput, PageState } from '@/shared/ui'
import {
  formatCivilDateFull,
  getCivilDateInTimezone,
  isCivilDateString,
  minutesIntoCivilDate,
  minutesSinceCivilMidnight,
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
import AgendaSkeleton from '../components/AgendaSkeleton.vue'
import BarberSelect from '../components/BarberSelect.vue'

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

function onBarberSelect(barberId: string) {
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

function formatAgendaTime(instant: string): string {
  if (!barbershopTimezone.value) return ''
  return new Intl.DateTimeFormat('es-CO', {
    timeZone: barbershopTimezone.value,
    hour: '2-digit',
    minute: '2-digit',
    hourCycle: 'h23',
  }).format(new Date(instant))
}

function entryTimeRange(entry: DailyAgendaEntry): string {
  return `${formatAgendaTime(entry.startsAt)}–${formatAgendaTime(entry.endsAt)}`
}

function entryTime(entry: DailyAgendaEntry): string {
  return formatAgendaTime(entry.startsAt)
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
  if (!barbershopTimezone.value || !selectedDate.value) return null
  const tz = barbershopTimezone.value
  const date = selectedDate.value

  // Día vacío (issue #189, evento 09 del atlas): se conserva el eje con el
  // marcador "Ahora" en vez de ocultar la línea temporal. Sin turnos que
  // fijen un rango, y sin un concepto de horario comercial en este
  // contrato, el eje se centra en la hora actual — el único dato real
  // disponible — en vez de inventar un rango de negocio no verificable. Un
  // día vacío que no es "hoy" no tiene "Ahora" que centrar, así que
  // conserva el comportamiento anterior (sin línea temporal).
  if (entries.value.length === 0) {
    if (!isViewingToday.value) return null
    const nowMinute = minutesIntoCivilDate(new Date().toISOString(), date, tz)
    const start = Math.max(0, Math.floor(nowMinute / 60) - 2) * 60
    const end = Math.min(24, Math.ceil(nowMinute / 60) + 2) * 60
    return { start, end: Math.max(end, start + MIN_TIMELINE_SPAN_MINUTES) }
  }

  const starts = entries.value.map((e) => minutesIntoCivilDate(e.startsAt, date, tz))
  // El fin usa la versión SIN recortar (issue #189): un turno que termina el
  // día siguiente (evento 11 del atlas) extiende el eje hasta esa hora real
  // en vez de cortarlo en medianoche, para que la ficha se vea completa y la
  // marca "Cambio de día" tenga carril donde dibujarse.
  const ends = entries.value.map((e) => minutesSinceCivilMidnight(e.endsAt, date, tz))
  // El atlas conserva una hora de contexto antes y después del primer y el
  // último turno. El intervalo sigue naciendo de sus horas reales, pero evita
  // que una ficha extrema quede pegada al borde del carril.
  const startHour = Math.max(0, Math.floor(Math.min(...starts) / 60) * 60 - 60)
  // En jornada diurna el atlas deja dos horas completas tras el último
  // turno. El caso nocturno conserva una sola para que 22:00–02:00 sitúe
  // «Cambio de día» exactamente a mitad del carril, sin recortar el turno.
  const lastEnd = Math.max(...ends)
  const endHour = Math.ceil(lastEnd / 60) * 60 + (lastEnd > 24 * 60 ? 60 : 120)
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

// Guías de media hora (issue #189): tenues, sin etiqueta — solo marcan el
// carril entre cada par de horas para que una ficha corta se pueda leer
// contra el eje sin contar píxeles.
const timelineHalfHourGuides = computed(() => {
  const bounds = timelineBounds.value
  if (!bounds) return []
  const guides: number[] = []
  for (let minute = bounds.start + 30; minute < bounds.end; minute += 60) {
    guides.push(minute)
  }
  return guides
})

function timelinePercent(minute: number): string {
  const bounds = timelineBounds.value
  if (!bounds) return '0%'
  return `${((minute - bounds.start) / (bounds.end - bounds.start)) * 100}%`
}

function timelineSlipWidthPercent(entry: DailyAgendaEntry): number {
  const bounds = timelineBounds.value
  if (!bounds || !selectedDate.value || !barbershopTimezone.value) return 0
  const start = minutesIntoCivilDate(entry.startsAt, selectedDate.value, barbershopTimezone.value)
  // Fin sin recortar, igual que en timelineBounds: la ficha ocupa su
  // duración real aunque cruce medianoche.
  const end = minutesSinceCivilMidnight(entry.endsAt, selectedDate.value, barbershopTimezone.value)
  const span = bounds.end - bounds.start
  // Ancho mínimo visual del 4%: una ficha muy corta sigue siendo legible en
  // la línea de tiempo sin que eso cambie su duración real.
  return Math.max(((end - start) / span) * 100, 4)
}

function timelineSlipStyle(entry: DailyAgendaEntry): { left: string; width: string } {
  const bounds = timelineBounds.value
  if (!bounds || !selectedDate.value || !barbershopTimezone.value)
    return { left: '0%', width: '0%' }
  const start = minutesIntoCivilDate(entry.startsAt, selectedDate.value, barbershopTimezone.value)
  return { left: timelinePercent(start), width: `${timelineSlipWidthPercent(entry)}%` }
}

// La ficha muestra rango horario, persona y servicio cuando la duración le
// da ancho; con poco ancho conserva solo la hora de inicio y la persona,
// sin el rango completo ni el servicio (issue #189, atlas evento 01: solo
// la ficha más ancha del ejemplo — Samuel Díaz, 60min — lleva las tres
// líneas; el resto muestra hora de inicio + nombre). El umbral es el mismo
// porcentaje que ya gobierna el ancho real de la ficha, no un breakpoint de
// viewport aparte.
function timelineSlipDetail(entry: DailyAgendaEntry): 'full' | 'compact' {
  return timelineSlipWidthPercent(entry) >= 8.5 ? 'full' : 'compact'
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

// "Cambio de día" (issue #189, evento 11 del atlas): decorativa igual que
// "Ahora" — marca dónde el carril cruza medianoche cuando un turno de este
// día se extiende hasta el siguiente. Nunca decide a qué día pertenece una
// cita (eso ya lo resolvió el servidor, DEC-075): es solo geometría.
const dayChangeMarkerPercent = computed(() => {
  const bounds = timelineBounds.value
  if (!bounds || bounds.end <= 24 * 60) return null
  return timelinePercent(24 * 60)
})
</script>

<template>
  <section class="daily-agenda-page" aria-labelledby="daily-agenda-page-title">
    <header class="daily-agenda-page__header">
      <div>
        <h1 id="daily-agenda-page-title" class="daily-agenda-page__title">Agenda</h1>
        <p v-if="selectedDateLabel" class="daily-agenda-page__date">
          {{ selectedDateLabel
          }}<template v-if="barbershopTimezone"> · Zona {{ barbershopTimezone }}</template>
        </p>
      </div>
      <BaseButton
        v-if="pageStatus === 'ready' && barbers.length > 0"
        type="button"
        variant="secondary"
        class="daily-agenda-page__cta"
        @click="goToNewAppointment"
      >
        Nuevo turno
      </BaseButton>
    </header>

    <!-- Evento 02 del atlas: sin barbero/zona resueltos, nada que anticipar
         todavía — estado de página centrado con spinner, sin divisor. -->
    <PageState
      v-if="pageStatus === 'loading'"
      variant="loading"
      headline="Cargando barberos…"
      role="status"
    />

    <!-- Evento 03: fallo al cargar el contexto inicial. -->
    <PageState
      v-else-if="pageStatus === 'load-error'"
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

    <template v-else>
      <!-- Evento 04: sin barberos activos, sin selección inventada ni CTA. -->
      <PageState
        v-if="barbers.length === 0"
        variant="info"
        headline="Aún no tienes barberos registrados."
        role="status"
      >
        Agrega uno en la sección
        <RouterLink class="page-state__link" :to="{ name: 'staff-barberos' }"
          >«Barberos»</RouterLink
        >
        para ver su agenda.
      </PageState>

      <template v-else>
        <div class="daily-agenda-page__controls">
          <div class="daily-agenda-page__picker">
            <label for="daily-agenda-barber-select" class="daily-agenda-page__label">Barbero</label>
            <BarberSelect
              :model-value="selectedBarberId"
              :barbers="barbers"
              @update:model-value="onBarberSelect"
            />
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

        <!-- Evento 10: la zona horaria no se pudo confirmar. La agenda sigue
             visible (§ trabajo requerido 5): esta es una nota al margen que
             acompaña contenido que sigue visible, no un PageState de página
             completa. -->
        <BaseAlert
          v-if="pageStatus === 'ready' && !barbershopTimezone"
          variant="warning"
          title="No pudimos confirmar la zona horaria"
          role="alert"
        >
          La agenda sigue visible, pero la navegación por fecha queda bloqueada hasta recuperar ese
          dato.
          <template #action>
            <BaseButton type="button" variant="secondary" @click="onRetryLoad">
              Reintentar
            </BaseButton>
          </template>
        </BaseAlert>

        <!-- Evento 05: barbero y fecha ya resueltos, la espera conserva la
             geometría del contenido por llegar (esqueleto, no spinner). -->
        <AgendaSkeleton
          v-if="agendaStatus === 'loading'"
          :label="`Cargando la agenda de ${selectedBarber?.fullName}…`"
        />

        <template v-else>
          <!-- Evento 06: cambio de fecha, agenda anterior conservada. -->
          <p
            v-if="agendaStatus === 'updating'"
            class="daily-agenda-page__updating"
            role="status"
            aria-live="polite"
          >
            Actualizando…
          </p>

          <!-- Evento 08: el barbero solicitado ya no está disponible. -->
          <PageState
            v-if="agendaStatus === 'not-found'"
            variant="warning"
            status-label="Atención"
            headline="Este barbero ya no está disponible"
            role="alert"
          >
            Elige otro barbero en la lista.
          </PageState>

          <!-- Evento 07: error recuperable de la agenda, barbero/fecha
               conservados para reintentar. -->
          <PageState
            v-else-if="agendaStatus === 'error'"
            variant="warning"
            status-label="Atención"
            headline="No pudimos cargar la agenda de este barbero"
            role="alert"
          >
            Revisa tu conexión e inténtalo de nuevo.
            <template #action>
              <BaseButton type="button" variant="secondary" @click="onRetryAgenda">
                Reintentar
              </BaseButton>
            </template>
          </PageState>

          <template
            v-if="
              agendaStatus === 'ready' || (hasLoadedEntriesOnce && agendaStatus !== 'not-found')
            "
          >
            <!-- Línea temporal horizontal de escritorio: presentación
                 visual adicional del mismo turno de la lista de abajo, que
                 sigue siendo la fuente accesible equivalente (§7.2, §8.1).
                 aria-hidden + tabindex="-1" evitan una segunda forma
                 redundante de llegar al mismo detalle para teclado/lector.
                 Se dibuja también con la lista vacía (evento 09): el eje del
                 día vacío se conserva con el marcador "Ahora". -->
            <div v-if="timelineBounds" class="daily-agenda-page__timeline" aria-hidden="true">
              <div class="daily-agenda-page__timeline-track">
                <div
                  v-for="minute in timelineHalfHourGuides"
                  :key="`half-${minute}`"
                  class="daily-agenda-page__timeline-guide"
                  :style="{ left: timelinePercent(minute) }"
                />

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
                  class="daily-agenda-page__timeline-mark daily-agenda-page__timeline-mark--now"
                  :style="{ left: nowMarkerPercent }"
                >
                  <span class="daily-agenda-page__timeline-mark-label">Ahora</span>
                </div>

                <div
                  v-if="dayChangeMarkerPercent"
                  class="daily-agenda-page__timeline-mark daily-agenda-page__timeline-mark--day-change"
                  :style="{ left: dayChangeMarkerPercent }"
                >
                  <span class="daily-agenda-page__timeline-mark-label">Cambio de día</span>
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
                    timelineSlipDetail(entry) === 'full' ? entryTimeRange(entry) : entryTime(entry)
                  }}</span>
                  <span class="daily-agenda-page__timeline-slip-name">{{
                    entry.attendeeName
                  }}</span>
                  <span
                    v-if="timelineSlipDetail(entry) === 'full'"
                    class="daily-agenda-page__timeline-slip-service"
                    >{{ entry.serviceName }}</span
                  >
                </RouterLink>
              </div>
            </div>

            <!-- Evento 09: día válido sin turnos. El eje vacío se conserva
                 (arriba); esta composición local vive dentro del contenido,
                 no reemplaza la pantalla completa como PageState — por eso
                 no usa ese componente. "Nuevo turno" vive aquí una sola vez,
                 no se repite en el encabezado. -->
            <div v-if="entries.length === 0" class="daily-agenda-page__empty-state">
              <span class="daily-agenda-page__empty-divider" aria-hidden="true" />
              <p class="daily-agenda-page__empty-headline">
                No hay turnos para {{ selectedBarber?.fullName }} {{ emptyStateDateText }}.
              </p>
              <BaseButton type="button" variant="primary" @click="goToNewAppointment">
                Nuevo turno
              </BaseButton>
            </div>

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
                <RouterLink
                  class="daily-agenda-page__item-main"
                  :to="{
                    name: 'agenda-detalle-turno',
                    params: { appointmentId: entry.id },
                    query: withQuery({}),
                  }"
                >
                  <span class="daily-agenda-page__item-time">{{ entryTimeRange(entry) }}</span>
                  <span class="daily-agenda-page__item-details">
                    <span class="daily-agenda-page__item-name">{{ entry.attendeeName }}</span>
                    <span class="daily-agenda-page__item-service">{{ entry.serviceName }}</span>
                  </span>
                </RouterLink>
                <BaseBadge
                  :class="statusBadgeClass(entry)"
                  size="sm"
                  outline
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
/* Superficie tinta de punta a punta (estandar-diseno-visual.md §3): la
   agenda es tiempo y orientación operativa, no un formulario. Las fichas de
   turno y los estados siguen en superficie clara para conservar el
   contraste de lectura sobre el fondo oscuro (§6.6, mismo criterio que el
   atlas: "convivencia de escritorio operativo con flujo móvil"). */
.daily-agenda-page {
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: 100%;
  padding: 20px;
  background-color: var(--color-surface-strong);
}

@media (min-width: 1024px) {
  .daily-agenda-page {
    padding: 26px 40px 30px;
  }
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
  font-family: var(--font-display);
  font-size: var(--font-size-h1);
  line-height: var(--font-size-h1-line);
  font-weight: var(--font-weight-h1);
  color: var(--color-on-strong);
}

@media (min-width: 1024px) {
  .daily-agenda-page__title {
    font-size: 40px;
    line-height: 46px;
  }
}

.daily-agenda-page__date {
  margin: var(--space-1) 0 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  color: var(--color-on-strong);
  opacity: 0.64;
}

.daily-agenda-page__cta {
  flex-shrink: 0;
}

.daily-agenda-page__updating {
  margin: 0;
  padding: var(--space-2) 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  color: var(--color-on-strong);
  opacity: 0.8;
}

.daily-agenda-page__controls {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

@media (min-width: 640px) {
  .daily-agenda-page__controls {
    flex-direction: row;
    flex-wrap: wrap;
    align-items: flex-end;
  }

  .daily-agenda-page__picker {
    flex: 1 1 240px;
  }
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
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--color-brand-accent-surface);
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

@media (min-width: 1024px) {
  .daily-agenda-page__controls {
    flex-wrap: nowrap;
  }

  .daily-agenda-page__picker {
    flex: 0 0 280px;
  }

  .daily-agenda-page__date-nav {
    flex-wrap: nowrap;
  }

  .daily-agenda-page__date-input {
    flex: 0 0 150px;
    min-width: 150px;
  }
}

/* Evento 09 del atlas: divisor-titular-acción locales, mismo lenguaje
   visual que PageState pero sin ocupar toda el área de contenido — el eje
   del día vacío sigue visible arriba. */
.daily-agenda-page__empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-8) var(--space-4);
  text-align: center;
}

.daily-agenda-page__empty-divider {
  position: relative;
  width: 220px;
  height: 2px;
  background-color: var(--color-accent-brass);
}

.daily-agenda-page__empty-divider::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 9px;
  height: 9px;
  transform: translate(-50%, -50%) rotate(45deg);
  background-color: var(--color-accent-brass);
  box-shadow: 0 0 0 8px var(--color-surface-strong);
}

.daily-agenda-page__empty-headline {
  margin: 0;
  max-width: 32ch;
  font-family: var(--font-display);
  font-size: var(--font-size-h2);
  line-height: var(--font-size-h2-line);
  font-weight: var(--font-weight-h2);
  color: var(--color-on-strong);
}

.daily-agenda-page__list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 0;
  margin: 0;
  list-style: none;
}

/* Ficha (§6.1, §7.2): pergamino, no blanco puro — sobre tinta el blanco
   deslumbra en una pantalla de uso continuo (issue #189). El filete
   izquierdo de latón es la misma regla que BaseAlert usa para su nota al
   margen, aquí en el color neutro de "turno vigente". */
.daily-agenda-page__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  /* Ficha (§6.1, §7.2): al menos 64px en móvil, no el objetivo táctil
     mínimo de 44px. */
  min-height: 64px;
  min-height: 64px;
  padding: 13px 18px;
  background-color: var(--color-surface-muted);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-left: var(--border-width-emphasis) solid var(--color-accent-brass);
  border-radius: 2px;
}

/* Turno terminal (issue #189): cambia de MATERIAL, no de peso — un relleno
   gris u opacidad reducida lo dejaba pesando igual o más que un turno
   vigente. Pasa a ser un registro con contorno sobre la tinta: sigue siendo
   un hecho del día (RN-CIT-04), solo dejó de ser el foco de atención. */
.daily-agenda-page__item--terminal {
  background-color: transparent;
  border-color: rgb(244 240 231 / 24%);
  border-left-color: rgb(244 240 231 / 24%);
}

.daily-agenda-page__item--terminal .daily-agenda-page__item-time,
.daily-agenda-page__item--terminal .daily-agenda-page__item-name {
  color: var(--color-on-strong);
}

.daily-agenda-page__item--terminal .daily-agenda-page__item-service {
  color: var(--color-on-strong);
  opacity: 0.64;
}

/* La fila abre el detalle del turno (HU-064) por su bloque principal
   (hora/persona/servicio), no por toda la tarjeta: el badge de estado queda
   fuera del enlace, sin volverse un control ambiguo. */
.daily-agenda-page__item-main {
  display: flex;
  flex: 1;
  min-width: 0;
  align-items: center;
  gap: 20px;
  min-height: 0;
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
  flex: 0 0 118px;
  font-family: var(--font-family-base);
  font-size: 15px;
  font-weight: 400;
  font-variant-numeric: tabular-nums;
  color: var(--color-text-primary);
}

.daily-agenda-page__item-details {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0;
  line-height: 20px;
}

/* Persona atendida como texto principal de la ficha (§6.1). */
.daily-agenda-page__item-name {
  font-family: var(--font-family-base);
  font-size: 17px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.daily-agenda-page__item-service {
  font-family: var(--font-family-base);
  margin-top: 2px;
  font-size: 14px;
  line-height: 16px;
  color: var(--color-text-secondary);
}

@media (max-width: 1023px) {
  .daily-agenda-page__cta {
    width: 100%;
  }

  .daily-agenda-page__date-nav {
    display: grid;
    width: 100%;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1.65fr) minmax(0, 1.05fr);
  }

  .daily-agenda-page__date-input {
    min-width: 0;
  }

  .daily-agenda-page__item {
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
    padding: 13px 15px;
  }

  .daily-agenda-page__item-main {
    flex-direction: column;
    align-items: flex-start;
    min-height: 0;
    gap: 0;
  }

  .daily-agenda-page__item-time {
    flex-basis: auto;
    font-size: 13px;
    color: var(--color-text-secondary);
  }

  .daily-agenda-page__item-name {
    font-size: 16px;
    line-height: 22px;
  }

  .daily-agenda-page__item-service {
    font-size: 13px;
    line-height: 18px;
  }
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

  /* El carril es una región definida (issue #189, atlas panel-agenda-eventos):
     un borde propio, no solo marcas de hora flotando sobre el fondo. */
  .daily-agenda-page__timeline-track {
    position: relative;
    height: 96px;
    /* Banda superior propia para las marcas (Ahora/Cambio de día, ~32px) y
       otra más baja para las etiquetas de hora (~16px): issue #189 corrige
       que antes compartían la misma fila y podían superponerse cuando una
       marca caía cerca de una hora en punto. Espacio extra arriba: cuando
       "Ahora" y "Cambio de día" coinciden en x (turno nocturno que empieza
       hoy), la segunda insignia sube a una tercera banda propia. */
    margin-top: 62px;
    padding: 0;
    border: var(--border-width-normal) solid rgb(244 240 231 / 16%);
    border-radius: 2px;
  }

  .daily-agenda-page__timeline-guide {
    position: absolute;
    top: 0;
    bottom: 0;
    border-left: var(--border-width-normal) dashed rgb(244 240 231 / 10%);
  }

  .daily-agenda-page__timeline-tick {
    position: absolute;
    top: 0;
    bottom: 0;
    border-left: var(--border-width-normal) dashed rgb(244 240 231 / 24%);
  }

  .daily-agenda-page__timeline-tick-label {
    position: absolute;
    top: -20px;
    left: 0;
    white-space: nowrap;
    font-size: var(--font-size-caption);
    font-variant-numeric: tabular-nums;
    color: var(--color-on-strong);
    opacity: 0.64;
  }

  .daily-agenda-page__timeline-mark {
    position: absolute;
    top: 0;
    bottom: 0;
    z-index: 1;
    border-left: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  }

  /* "Ahora" (latón, foco/énfasis secundario, §4.1): etiqueta como insignia
     rellena — es la marca de mayor prioridad del eje. */
  .daily-agenda-page__timeline-mark-label {
    position: absolute;
    top: -28px;
    left: 50%;
    transform: translateX(-50%);
    padding: 4px 9px;
    white-space: nowrap;
    font-size: var(--font-size-caption);
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    border-radius: 2px;
  }

  .daily-agenda-page__timeline-mark--now .daily-agenda-page__timeline-mark-label {
    color: var(--color-surface-strong);
    background-color: var(--color-brand-accent-surface);
  }

  /* "Cambio de día": misma familia que "Ahora" pero de menor prioridad —
     contorno en vez de relleno. Una banda propia más arriba (issue #189):
     un turno nocturno que empieza "hoy" puede coincidir en x con "Ahora"
     (evento 11 con reloj cercano a medianoche), y las dos insignias
     necesitan no superponerse aunque caigan en el mismo punto del carril. */
  .daily-agenda-page__timeline-mark--day-change .daily-agenda-page__timeline-mark-label {
    top: -56px;
    color: var(--color-brand-accent-surface);
    background-color: transparent;
    border: var(--border-width-normal) solid var(--color-brand-accent-surface);
  }

  .daily-agenda-page__timeline-slip {
    position: absolute;
    top: 10px;
    bottom: 10px;
    display: flex;
    min-width: 0;
    flex-direction: column;
    justify-content: center;
    gap: 2px;
    overflow: hidden;
    padding: 0 12px;
    background-color: var(--color-surface-muted);
    border: 0;
    border-left: 3px solid var(--color-accent-brass);
    border-radius: 2px;
    color: var(--color-action-primary);
    text-decoration: none;
  }

  /* Mismo criterio de material que la ficha de lista: contorno sobre tinta,
     no un relleno gris que pese igual o más que un turno vigente. */
  .daily-agenda-page__timeline-slip--terminal {
    background-color: transparent;
    border-color: rgb(244 240 231 / 24%);
    border-left-color: rgb(244 240 231 / 24%);
    color: var(--color-on-strong);
  }

  .daily-agenda-page__timeline-slip-time {
    font-size: 11px;
    font-weight: 400;
    font-variant-numeric: tabular-nums;
  }

  .daily-agenda-page__timeline-slip-name {
    overflow: hidden;
    font-size: 13px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .daily-agenda-page__timeline-slip-service {
    overflow: hidden;
    font-size: 12px;
    color: var(--color-text-secondary);
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

:deep(.base-badge) {
  height: auto;
  padding: 5px 12px;
  border-radius: 2px;
  font-size: 12px;
  font-weight: 600;
  line-height: 14px;
}

@media (max-width: 1023px) {
  :deep(.base-badge) {
    margin-top: 4px;
    padding: 4px 9px;
    font-size: 11px;
  }
}
</style>
