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
  APPOINTMENT_STATUS_BADGE_VARIANT,
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

function statusBadgeVariant(entry: DailyAgendaEntry) {
  return APPOINTMENT_STATUS_BADGE_VARIANT[entry.status]
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

// Zoom del carril (pedido explícito del propietario, 2026-09-04: "habilita
// que se pueda hacer zoom al calendario para que se parta en 15 en 15...
// y se puedan ver bien las cards"). No representado en el atlas panel-
// agenda-eventos (una imagen estática sin controles): es una función
// nueva, fuera del modo "fidelidad medible" de DEC-080 para el resto de
// esta pantalla, autorizada aparte. Cada nivel reduce el paso del eje
// (60min → 30min → 15min) y ensancha el carril en la misma proporción;
// timelineBounds/timelinePercent no cambian, solo el ancho en px del
// contenedor que las interpreta.
const TIMELINE_ZOOM_LEVELS = [1, 2, 4] as const
const timelineZoomIndex = ref(0)
const timelineZoom = computed(() => TIMELINE_ZOOM_LEVELS[timelineZoomIndex.value]!)
const canZoomInTimeline = computed(() => timelineZoomIndex.value < TIMELINE_ZOOM_LEVELS.length - 1)
const canZoomOutTimeline = computed(() => timelineZoomIndex.value > 0)

function zoomTimelineIn() {
  if (canZoomInTimeline.value) timelineZoomIndex.value += 1
}

function zoomTimelineOut() {
  if (canZoomOutTimeline.value) timelineZoomIndex.value -= 1
}

const timelineTickStepMinutes = computed(() => {
  const zoom = timelineZoom.value
  if (zoom >= 4) return 15
  if (zoom >= 2) return 30
  return 60
})

const timelineZoomLabel = computed(() => {
  const step = timelineTickStepMinutes.value
  return step === 60 ? '1 h' : `${step} min`
})

const timelineBounds = computed(() => {
  if (!barbershopTimezone.value || !selectedDate.value) return null
  const tz = barbershopTimezone.value
  const date = selectedDate.value

  // Pedido explícito del propietario (2026-09-04): el carril siempre
  // cubre el día completo (00:00–24:00), no una ventana recortada
  // alrededor de los turnos reales — el zoom (arriba) ya resuelve ver el
  // detalle de una franja concreta. Un día vacío que no es "hoy" conserva
  // el comportamiento anterior (sin línea temporal): sin turnos ni "Ahora"
  // que centrar no hay nada real que el eje esté mostrando.
  if (entries.value.length === 0 && !isViewingToday.value) return null
  if (entries.value.length === 0) return { start: 0, end: 24 * 60 }

  // El fin usa la versión SIN recortar (issue #189): un turno que termina el
  // día siguiente (evento 11 del atlas) extiende el eje hasta esa hora real
  // en vez de cortarlo en medianoche, para que la ficha se vea completa y la
  // marca "Cambio de día" tenga carril donde dibujarse. En cualquier otro
  // caso el fin es medianoche (24:00), nunca antes.
  const ends = entries.value.map((e) => minutesSinceCivilMidnight(e.endsAt, date, tz))
  const lastEnd = Math.max(...ends)
  return { start: 0, end: Math.max(24 * 60, lastEnd) }
})

// `align` evita que la etiqueta centrada de la primera o la última hora se
// salga del carril (issue #189): centrar (translateX(-50%)) es correcto
// para una marca interior, pero en los dos extremos la mitad de la
// etiqueta cae fuera del carril — ahí se ancla hacia adentro en vez de
// centrarse, sin recortar el texto ni disparar scroll horizontal.
const timelineTicks = computed(() => {
  const bounds = timelineBounds.value
  if (!bounds) return []
  const step = timelineTickStepMinutes.value
  const ticks: { minute: number; label: string; align: 'start' | 'center' | 'end' }[] = []
  for (let minute = bounds.start; minute <= bounds.end; minute += step) {
    const hour = Math.floor(minute / 60) % 24
    const mins = minute % 60
    const align = minute === bounds.start ? 'start' : minute === bounds.end ? 'end' : 'center'
    ticks.push({
      minute,
      label: `${String(hour).padStart(2, '0')}:${String(mins).padStart(2, '0')}`,
      align,
    })
  }
  return ticks
})

// Guías de media hora (issue #189): tenues, sin etiqueta — solo marcan el
// carril entre cada par de horas para que una ficha corta se pueda leer
// contra el eje sin contar píxeles. Con zoom >1 el eje ya etiqueta cada
// 30/15min (arriba), así que esta guía intermedia deja de aportar nada
// nuevo y se apaga.
const timelineHalfHourGuides = computed(() => {
  const bounds = timelineBounds.value
  if (!bounds || timelineZoom.value > 1) return []
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
  // El umbral se mide en ancho real (%×zoom), no en el % crudo respecto al
  // carril: con el carril ensanchado por zoom, una ficha que antes no
  // cabía con las tres líneas ahora sí tiene espacio real, aunque su %
  // respecto a timelineBounds no haya cambiado.
  return timelineSlipWidthPercent(entry) * timelineZoom.value >= 8.5 ? 'full' : 'compact'
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
              variant="ghost"
              :disabled="!canNavigateDates"
              aria-label="Día anterior"
              @click="goToPreviousDay"
            >
              Anterior
            </BaseButton>
            <div class="daily-agenda-page__date-input-wrap">
              <BaseInput
                type="date"
                label="Fecha"
                class="daily-agenda-page__date-input"
                :model-value="selectedDate ?? ''"
                :disabled="!canNavigateDates"
                @change="onDateInputChange"
              />
              <!-- El valor de un input[type=date] siempre es ISO (YYYY-MM-DD)
                   aunque el navegador lo pinte con el formato del locale del
                   SO (issue #189: el atlas muestra "2026-09-03" literal). Esta
                   capa superpuesta —no interactiva— reemplaza visualmente ese
                   render nativo por el mismo valor ya en ISO, sin tocar la
                   interacción real (el campo de abajo sigue enfocable,
                   editable por teclado y con el picker nativo). -->
              <span
                v-if="selectedDate"
                class="daily-agenda-page__date-display"
                aria-hidden="true"
                >{{ selectedDate }}</span
              >
            </div>
            <BaseButton
              type="button"
              variant="ghost"
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
            <div v-if="timelineBounds" class="daily-agenda-page__timeline-wrapper">
              <!-- Control de zoom (pedido explícito del propietario,
                   2026-09-04): no representado en el atlas — es una imagen
                   estática sin controles — y por eso vive fuera del bloque
                   aria-hidden de abajo, como un control real más de la
                   pantalla. Cada nivel reduce el paso del eje (1h → 30min →
                   15min) y ensancha el carril en la misma proporción, para
                   que una ficha corta deje de ir en modo compacto. -->
              <div class="daily-agenda-page__timeline-zoom">
                <button
                  type="button"
                  class="daily-agenda-page__timeline-zoom-btn"
                  :disabled="!canZoomOutTimeline"
                  aria-label="Alejar el calendario"
                  @click="zoomTimelineOut"
                >
                  −
                </button>
                <span class="daily-agenda-page__timeline-zoom-label">{{ timelineZoomLabel }}</span>
                <button
                  type="button"
                  class="daily-agenda-page__timeline-zoom-btn"
                  :disabled="!canZoomInTimeline"
                  aria-label="Acercar el calendario"
                  @click="zoomTimelineIn"
                >
                  +
                </button>
              </div>

              <div class="daily-agenda-page__timeline" aria-hidden="true">
                <div
                  class="daily-agenda-page__timeline-scroll"
                  :style="{ width: `${timelineZoom * 100}%` }"
                >
                  <!-- Eje (horas) y marcas (Ahora/Cambio de día) del atlas son
                       dos filas propias, apiladas ANTES del carril (.axis →
                       .marks → .track en
                       tools/mockups/panel-agenda-eventos/render.mjs) — nunca
                       texto flotando dentro del carril con un top negativo.
                       Con eso una marca nunca puede quedar encima de una hora
                       en punto (issue #189, reporte en vivo: "está por
                       encima del tiempo"): ocupan bandas verticales
                       disjuntas, no compiten por el mismo espacio aunque
                       coincidan en x. -->
                  <div class="daily-agenda-page__timeline-axis">
                    <span
                      v-for="tick in timelineTicks"
                      :key="tick.minute"
                      class="daily-agenda-page__timeline-tick-wrap"
                      :style="{ left: timelinePercent(tick.minute) }"
                    >
                      <span
                        class="daily-agenda-page__timeline-tick-label"
                        :class="`daily-agenda-page__timeline-tick-label--${tick.align}`"
                        >{{ tick.label }}</span
                      >
                    </span>
                  </div>

                  <div class="daily-agenda-page__timeline-marks">
                    <span
                      v-if="nowMarkerPercent"
                      class="daily-agenda-page__timeline-mark-label daily-agenda-page__timeline-mark-label--now"
                      :style="{ left: nowMarkerPercent }"
                      >Ahora</span
                    >
                    <span
                      v-if="dayChangeMarkerPercent"
                      class="daily-agenda-page__timeline-mark-label daily-agenda-page__timeline-mark-label--day-change"
                      :style="{ left: dayChangeMarkerPercent }"
                      >Cambio de día</span
                    >
                  </div>

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
                    />

                    <div
                      v-if="nowMarkerPercent"
                      class="daily-agenda-page__timeline-mark daily-agenda-page__timeline-mark--now"
                      :style="{ left: nowMarkerPercent }"
                    />

                    <div
                      v-if="dayChangeMarkerPercent"
                      class="daily-agenda-page__timeline-mark daily-agenda-page__timeline-mark--day-change"
                      :style="{ left: dayChangeMarkerPercent }"
                    />

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
                        timelineSlipDetail(entry) === 'full'
                          ? entryTimeRange(entry)
                          : entryTime(entry)
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
                  :variant="statusBadgeVariant(entry)"
                  size="sm"
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
  color: var(--color-on-strong-muted);
}

.daily-agenda-page__cta {
  flex-shrink: 0;
}

.daily-agenda-page__updating {
  margin: 0;
  padding: var(--space-2) 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  color: var(--color-on-strong-muted);
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

.daily-agenda-page__date-input-wrap {
  position: relative;
  flex: 1 1 180px;
  min-width: 160px;
}

/* Anterior/Siguiente/Fecha se apoyan directamente sobre
   --color-surface-strong (tinta): BaseButton--ghost y BaseInput están
   calibrados para superficie clara (blanco/latón oscuro), pero el atlas
   panel-agenda-eventos usa latón claro sobre un campo casi transparente
   para este control (.btn--ghost/.control en tools/mockups/
   panel-agenda-eventos/render.mjs). Se sobrescribe solo aquí, no en los
   componentes compartidos, porque el resto de sus usos (auth-eventos)
   sí está sobre superficie clara y necesita el latón oscuro. */
.daily-agenda-page__date-nav :deep(.base-button--ghost) {
  color: var(--color-brand-accent-surface);
  border-color: rgb(184 149 90 / 50%);
  border-bottom-color: var(--color-brand-accent-surface);
}

.daily-agenda-page__date-nav
  :deep(.base-button--ghost:hover:not(:disabled):not(.base-button--loading)) {
  background-color: rgb(184 149 90 / 12%);
}

.daily-agenda-page__date-input :deep(.base-input__label) {
  color: var(--color-brand-accent-surface);
}

.daily-agenda-page__date-input :deep(.base-input) {
  --input-bg: rgb(244 240 231 / 5%);
  --input-border-color: rgb(244 240 231 / 16%);
  --input-border-base-color: var(--color-brand-accent-surface);
  color: var(--color-on-strong);
}

/* El control del atlas (.control en tools/mockups/panel-agenda-eventos/
   render.mjs) es solo texto, sin icono: el calendario propio del navegador
   se oculta. El campo entero — no solo el icono — sigue abriendo el
   selector nativo al hacer clic en Chromium, así que la interacción no se
   pierde. El texto propio del navegador (formato del locale del SO, no el
   ISO literal del atlas) también se vuelve transparente: la capa
   `.daily-agenda-page__date-display` de abajo lo reemplaza visualmente sin
   tocar el valor real ni la edición por teclado. */
.daily-agenda-page__date-input :deep(.base-input::-webkit-calendar-picker-indicator) {
  display: none;
}

.daily-agenda-page__date-input :deep(.base-input::-webkit-datetime-edit) {
  color: transparent;
}

.daily-agenda-page__date-input :deep(.base-input:disabled),
.daily-agenda-page__date-input :deep(.base-input--disabled) {
  background-color: var(--input-bg);
  border-color: var(--input-border-color);
  border-bottom-color: rgb(244 240 231 / 28%);
  color: var(--color-on-strong-muted);
  opacity: 0.42;
}

.daily-agenda-page__date-display {
  position: absolute;
  right: var(--space-4);
  bottom: 0;
  left: var(--space-4);
  display: flex;
  align-items: center;
  height: var(--control-height);
  overflow: hidden;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-on-strong);
  text-overflow: ellipsis;
  white-space: nowrap;
  pointer-events: none;
}

.daily-agenda-page__date-input:has(:disabled) ~ .daily-agenda-page__date-display {
  color: var(--color-on-strong-muted);
  opacity: 0.42;
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

  .daily-agenda-page__date-input-wrap {
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
  /* 3px literal, no --border-width-emphasis (2px): igual que
     .daily-agenda-page__timeline-slip, calcado del filete de .row en el
     atlas (border-left:3px solid var(--brass-deep)). */
  border-left: 3px solid var(--color-accent-brass);
  border-radius: 2px;
}

/* Turno terminal (issue #189): cambia de MATERIAL, no de peso — un relleno
   gris u opacidad reducida lo dejaba pesando igual o más que un turno
   vigente. Pasa a ser un registro con contorno sobre la tinta: sigue siendo
   un hecho del día (RN-CIT-04), solo dejó de ser el foco de atención. */
.daily-agenda-page__item--terminal {
  background-color: transparent;
  /* Más visible que el 24% original del atlas (issue #189, reporte en
     vivo: "las tarjetas transparentes... ni se notan"). */
  border-color: rgb(244 240 231 / 45%);
  border-left-color: rgb(244 240 231 / 55%);
}

/* El nombre conserva color pleno incluso en un turno terminal — sigue
   siendo el dato principal de la fila (atlas: .row--terminal solo atenúa
   hora y servicio, nunca la persona). */
.daily-agenda-page__item--terminal .daily-agenda-page__item-name {
  color: var(--color-on-strong);
}

.daily-agenda-page__item--terminal .daily-agenda-page__item-time,
.daily-agenda-page__item--terminal .daily-agenda-page__item-service {
  color: var(--color-on-strong-muted);
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
  letter-spacing: 0.02em;
  color: var(--color-text-secondary);
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

  .daily-agenda-page__date-input-wrap {
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
.daily-agenda-page__timeline-wrapper {
  display: none;
}

.daily-agenda-page__timeline {
  display: none;
}

@media (min-width: 1024px) {
  .daily-agenda-page__timeline-wrapper {
    display: block;
  }

  .daily-agenda-page__timeline-zoom {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }

  .daily-agenda-page__timeline-zoom-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    padding: 0;
    background-color: transparent;
    color: var(--color-brand-accent-surface);
    border: 1px solid rgb(184 149 90 / 50%);
    border-radius: 2px;
    font-size: 14px;
    line-height: 1;
    cursor: pointer;
  }

  .daily-agenda-page__timeline-zoom-btn:hover:not(:disabled) {
    background-color: rgb(184 149 90 / 12%);
  }

  .daily-agenda-page__timeline-zoom-btn:disabled {
    opacity: 0.42;
    cursor: not-allowed;
  }

  .daily-agenda-page__timeline-zoom-label {
    min-width: 40px;
    font-size: var(--font-size-caption);
    font-variant-numeric: tabular-nums;
    text-align: center;
    color: var(--color-on-strong-muted);
  }

  .daily-agenda-page__timeline {
    display: block;
    /* El carril es 100% ancho relativo (posiciones en %, nunca px fijos) al
       nivel de zoom por defecto: no necesita scroll propio. `auto`, no
       `hidden` (issue #189): con zoom > 1 el carril interno
       (.timeline-scroll) se ensancha más allá de este contenedor a
       propósito y necesita poder desplazarse; al zoom por defecto nunca
       sobra ancho real (la etiqueta de la última hora ya se ancla hacia
       adentro, ver --end más abajo), así que `auto` no dibuja una barra
       de scroll que no haga falta. */
    overflow-x: auto;
  }

  .daily-agenda-page__timeline-scroll {
    min-width: 100%;
  }

  /* Eje, marcas y carril: tres filas propias apiladas, calcadas de
     .axis/.marks/.track en tools/mockups/panel-agenda-eventos/render.mjs
     (issue #189). Nunca texto posicionado con un top negativo dentro del
     carril — así una marca ("Ahora"/"Cambio de día") nunca puede quedar
     encima de una hora en punto: viven en bandas verticales disjuntas, no
     compiten por el mismo espacio aunque coincidan en x. */
  .daily-agenda-page__timeline-axis {
    position: relative;
    height: 18px;
  }

  .daily-agenda-page__timeline-tick-wrap {
    position: absolute;
    top: 0;
  }

  .daily-agenda-page__timeline-tick-label {
    display: block;
    white-space: nowrap;
    font-size: var(--font-size-caption);
    /* Mismo motivo que .timeline-mark-label: sin esto hereda
       line-height:24px de body y desborda la fila de 18px del eje. */
    line-height: 1;
    font-variant-numeric: tabular-nums;
    color: var(--color-on-strong-muted);
  }

  /* Una marca interior se centra sobre su línea; en los dos extremos del
     eje centrar saca la mitad de la etiqueta fuera del carril, así que ahí
     se ancla hacia adentro en su lugar (issue #189: esto era la causa real
     del scroll horizontal — no un carril genuinamente más ancho que la
     pantalla, solo una etiqueta de borde sangrando unos px). translateX en
     vez de left/right: el wrap ya lleva su posición por `:style`, y un
     estilo en línea siempre gana sobre cualquier `left`/`right` de clase. */
  .daily-agenda-page__timeline-tick-label--center {
    transform: translateX(-50%);
  }

  .daily-agenda-page__timeline-tick-label--end {
    transform: translateX(-100%);
  }

  .daily-agenda-page__timeline-marks {
    position: relative;
    height: 22px;
  }

  /* "Ahora" (latón, foco/énfasis secundario, §4.1): etiqueta como insignia
     rellena — es la marca de mayor prioridad del eje. */
  .daily-agenda-page__timeline-mark-label {
    position: absolute;
    top: 0;
    transform: translateX(-50%);
    padding: 4px 9px;
    white-space: nowrap;
    /* Literal 10px + .12em, no --font-size-caption (12px) ni el
       letter-spacing de otras insignias (issue #189, reporte en vivo:
       "está muy grande") — .mark em en el atlas usa exactamente estos dos
       valores en escritorio, más angostos que el resto del sistema de
       insignias porque esta es la única marca que flota sola sobre el eje,
       no dentro de una fila con más contexto alrededor. */
    font-size: 10px;
    /* Sin esto hereda line-height:24px de body (src/styles/base.css) — el
       atlas nunca fija un line-height para .mark em porque su body no
       redefine el valor por defecto del navegador (issue #189, reporte en
       vivo: "está muy grande" / "muy pegado al calendario"). Con la
       herencia de 24px el badge medía 32px de alto real pese a su
       font-size de 10px, desbordaba la fila de 22px y terminaba
       superponiendo el borde del carril en vez de dejar aire antes de él. */
    line-height: 1;
    font-weight: 600;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    border-radius: 2px;
  }

  .daily-agenda-page__timeline-mark-label--now {
    color: var(--color-surface-strong);
    background-color: var(--color-brand-accent-surface);
  }

  /* "Cambio de día": misma familia que "Ahora" pero de menor prioridad —
     contorno en vez de relleno. Cuando un turno nocturno que empieza "hoy"
     hace coincidir "Ahora" y "Cambio de día" en el mismo x (evento 11), la
     segunda insignia baja una fila dentro de esta misma banda en vez de
     superponerse. */
  .daily-agenda-page__timeline-mark-label--day-change {
    top: 22px;
    color: var(--color-brand-accent-surface);
    background-color: transparent;
    border: var(--border-width-normal) solid var(--color-brand-accent-surface);
  }

  .daily-agenda-page__timeline-marks:has(.daily-agenda-page__timeline-mark-label--day-change) {
    height: 44px;
  }

  /* El carril es una región definida: un borde propio, no solo marcas de
     hora flotando sobre el fondo. */
  .daily-agenda-page__timeline-track {
    position: relative;
    height: 96px;
    /* Sin margen: en el atlas .axis/.marks/.track se apilan sin espacio
       (issue #189, reporte en vivo: "está muy separado ahora"). */
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

  .daily-agenda-page__timeline-mark {
    position: absolute;
    top: 0;
    bottom: 0;
    z-index: 1;
    border-left: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
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
    /* La base (.daily-agenda-page__timeline-slip) fija `border: 0` en los
       cuatro lados y solo repone el izquierdo — un turno vigente no
       necesita perímetro, apoya en el relleno de pergamino. Terminal SÍ
       necesita un perímetro real (issue #189, reporte en vivo: "no está
       la caja literal"): sin `border-width`/`border-style` propios aquí,
       `border-color` no tenía nada que colorear y la ficha quedaba sin
       contorno visible, solo con el filete izquierdo. Más visible que el
       24% original del atlas ("las tarjetas transparentes... ni se
       notan"). */
    border: var(--border-width-normal) solid rgb(244 240 231 / 45%);
    border-left: 3px solid rgb(244 240 231 / 55%);
    color: var(--color-on-strong);
  }

  /* Igual que la fila de lista: la hora se atenúa, la persona conserva
     color pleno pero baja de peso (atlas: .slip--terminal redefine
     .slip__time a --on-ink-2 y .slip__name a --on-ink/500, nunca 600). */
  .daily-agenda-page__timeline-slip--terminal .daily-agenda-page__timeline-slip-time {
    color: var(--color-on-strong-muted);
  }

  .daily-agenda-page__timeline-slip--terminal .daily-agenda-page__timeline-slip-name {
    color: var(--color-on-strong);
    font-weight: 500;
  }

  .daily-agenda-page__timeline-slip-time {
    font-size: 11px;
    font-weight: 400;
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.03em;
    color: var(--color-text-secondary);
  }

  .daily-agenda-page__timeline-slip-name {
    overflow: hidden;
    font-size: 13px;
    font-weight: 600;
    color: var(--color-text-primary);
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

/* Rediseño local del estado de turno (issue #189, pedido explícito del
   propietario 2026-09-04: "parecen botones" — el badge con caja del atlas
   es fiel al mockup, pero al lado de Anterior/Siguiente/Nuevo turno (todas
   cajas con borde) se lee como un control más, no como un dato de la
   fila. Se abandona la caja por un rótulo — punto de color + versalitas
   espaciadas, el mismo vocabulario que "BARBERO"/"FECHA"/"AHORA" en esta
   misma pantalla — nunca inventando un color nuevo: sigue siendo el color
   de la variante (`--badge-text`), solo cambia cómo se presenta. */
:deep(.base-badge) {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: auto;
  padding: 0;
  background-color: transparent;
  border: none;
  border-radius: 0;
  font-size: 11px;
  font-weight: 600;
  line-height: 14px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

:deep(.base-badge)::before {
  content: '';
  flex-shrink: 0;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  /* currentColor, no --badge-dot-color: el punto sigue el mismo color de
     texto que ya resuelve la variante (y su atenuado terminal de abajo),
     sin depender de la mecánica de padding negativo del prop `dot`
     original de BaseBadge (pensada para su badge con caja). */
  background-color: currentColor;
}

/* Terminal (atlas .row--terminal .badge: pierde su color de estado
   individual): mismo criterio que el resto del material terminal de la
   fila — el texto de la insignia ("Completado", "Cancelado...") sigue
   distinguiendo el estado, el color ya no compite con un turno vigente. */
.daily-agenda-page__item--terminal :deep(.base-badge) {
  color: var(--color-on-strong-muted);
}

@media (max-width: 1023px) {
  :deep(.base-badge) {
    margin-top: 4px;
    font-size: 10px;
  }
}
</style>
