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
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter, type LocationQueryRaw } from 'vue-router'
import { useMinHoldLoading, useVocabulary } from '@/shared/composables'
import { capitalize, isSoloProfile } from '@/shared/model'
import { BaseAlert, BaseBadge, BaseButton, DiamondLoader, EmptyScene, PageState } from '@/shared/ui'
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
import { summarizeDay } from '../model/daySummary'
import {
  APPOINTMENT_STATUS_BADGE_VARIANT,
  APPOINTMENT_STATUS_LABELS,
  type DailyAgendaEntry,
} from '../model/dailyAgenda'
import AgendaSkeleton from '../components/AgendaSkeleton.vue'
import BaseDatePicker from '@/shared/ui/BaseDatePicker.vue'
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

// agendaDate es la fecha a la que pertenecen `entries`: se fija en el mismo
// instante que ellas, al llegar la respuesta. Todo lo que se deriva de los
// turnos (eje, fichas, "Ahora", "En proceso") usa esta fecha y no
// selectedDate: si usara la seleccionada, al cambiar de día la agenda anterior
// se recalcularía contra la fecha nueva y saltaría antes de que llegue la
// siguiente.
const agendaDate = ref<string | null>(null)
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

const agendaDateLabel = computed(() =>
  agendaDate.value ? formatCivilDateFull(agendaDate.value) : null,
)

const canNavigateDates = computed(() => barbershopTimezone.value !== null)

// Día civil vigente en la barbería: el calendario lo marca y ofrece «Hoy».
const todayCivilDate = computed(() =>
  barbershopTimezone.value ? getCivilDateInTimezone(barbershopTimezone.value) : null,
)

// isViewingToday decide el texto del estado vacío ("hoy" vs. una fecha
// explícita): sin zona conocida se asume "hoy" (mismo criterio degradado
// que HU-062, que nunca navegaba y siempre mostraba "hoy").
const isViewingToday = computed(() => {
  if (!agendaDate.value || !barbershopTimezone.value) return true
  return agendaDate.value === getCivilDateInTimezone(barbershopTimezone.value)
})

const emptyStateDateText = computed(() =>
  isViewingToday.value ? 'hoy' : `el ${agendaDateLabel.value}`,
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

// useMinHoldLoading (issue reportado 2026-09-28: "al entrar a Agenda o
// Servicios se ve como se genera el objeto... quiero el tema de la carga
// para tapar ese evento"): el rombo del evento 02 se muestra de inmediato
// (sigue siendo `pageStatus.value = 'loading'` lo primero que pasa), pero
// no cede el paso a "ready"/"load-error" hasta que pase un mínimo, para que
// una respuesta rápida no lo retire a medio parpadeo dejando ver la
// construcción cruda de los controles/el eje por debajo.
// Palabras de la barbería (DEC-110): con los valores iniciales, el texto de siempre.
const v = useVocabulary()

const { start: startPageHold, hold: holdPageReveal } = useMinHoldLoading()

async function loadPage() {
  pageStatus.value = 'loading'
  startPageHold()
  const [barbersOutcome, timezoneOutcome] = await Promise.all([
    fetchBarberSummaries(),
    fetchBarbershopTimezone(),
  ])

  barbershopTimezone.value = timezoneOutcome.kind === 'success' ? timezoneOutcome.timezone : null

  if (barbersOutcome.kind !== 'success') {
    holdPageReveal(() => {
      pageStatus.value = 'load-error'
    })
    return
  }

  barbers.value = barbersOutcome.items
  holdPageReveal(() => {
    pageStatus.value = 'ready'
    // El esqueleto empieza a verse justo ahora: su retención mínima cuenta
    // desde este instante, no desde que se pidió la agenda (ver loadAgenda).
    if (agendaStatus.value === 'loading') startAgendaHold()
  })

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

// Mismo criterio que startPageHold/holdPageReveal (issue 2026-09-28), pero
// solo para la PRIMERA carga ('loading', AgendaSkeleton): 'updating' ya
// tiene su propio transition-delay (.daily-agenda-page--updating, más
// abajo) para el caso "cambio de fecha/barbero con contenido previo", que
// no es el problema reportado y no se toca.
const { start: startAgendaHold, hold: holdAgendaReveal } = useMinHoldLoading()

async function loadAgenda(barberId: string, date: string | null) {
  const requestId = ++agendaRequestSeq
  const isFirstLoad = !hasLoadedEntriesOnce.value
  agendaStatus.value = isFirstLoad ? 'loading' : 'updating'
  // Mientras el spinner de página siga visible (pageStatus 'loading') el
  // esqueleto aún no se ve: su retención arranca al revelarse la página. Sin
  // esto, ambos temporizadores vencían con ~10 ms de diferencia y el
  // esqueleto parpadeaba un solo fotograma entre el spinner y la agenda
  // (issue reportado 2026-09-30: "al acceder a la pantalla de agenda, hay
  // un espasmo").
  if (isFirstLoad && pageStatus.value === 'ready') startAgendaHold()

  const outcome = await fetchDailyAgenda(barberId, date ?? undefined)

  // Una selección posterior (fecha, barbero o ambas) ya reemplazó el
  // destino vigente: esta respuesta llegó fuera de orden y se descarta sin
  // tocar el contenido ya mostrado (nunca sustituye la selección vigente
  // con una respuesta obsoleta).
  if (requestId !== agendaRequestSeq) return

  function apply() {
    switch (outcome.kind) {
      case 'success':
        entries.value = [...outcome.items].sort((a, b) => a.startsAt.localeCompare(b.startsAt))
        agendaDate.value = date
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

  if (isFirstLoad && pageStatus.value === 'ready') holdAgendaReveal(apply)
  else apply()
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

function goToNewAppointment() {
  void router.push({ name: 'agenda-nuevo-turno' })
}

// Ritmo propio del rombo de cada turno: duración y desfase derivados del id
// (hash estable), así los rombos no laten a la vez y el orden no cambia al
// volver a pintar la lista. El turno en proceso late algo más rápido.
function diamondStyle(entry: DailyAgendaEntry): Record<string, string> {
  let hash = 0
  for (const char of entry.id) hash = (hash * 31 + char.charCodeAt(0)) >>> 0
  const spread = (hash % 1000) / 1000
  const phase = ((hash >>> 10) % 1000) / 1000
  const [min, range] = isEntryInProgress(entry) ? [4, 2] : [6, 5]
  const duration = min + spread * range
  return {
    '--diamond-duration': `${duration.toFixed(1)}s`,
    '--diamond-delay': `-${(phase * duration).toFixed(1)}s`,
  }
}

// El turno en curso se muestra como "En proceso" aunque su estado siga siendo
// `confirmed`: es solo presentación de la agenda, no cambia el dato.
function statusLabel(entry: DailyAgendaEntry): string {
  if (isEntryInProgress(entry)) return 'En proceso'
  return APPOINTMENT_STATUS_LABELS[entry.status]
}

function statusBadgeVariant(entry: DailyAgendaEntry) {
  return APPOINTMENT_STATUS_BADGE_VARIANT[entry.status]
}

// Perfil de barbero individual (DEC-115): con un solo barbero no hay nada que
// elegir, así que el selector se oculta. La selección sigue viviendo en
// `route.query.barberId` (DEC-074): solo desaparece el control.
const hideBarberPicker = computed(() => isSoloProfile.value && barbers.value.length === 1)

// «Mi día» (DEC-115): una frase con lo que queda de hoy, derivada de los turnos ya
// cargados. Solo para el día en curso: en otra fecha «siguiente» no significaría nada.
// Un día sin turnos ya tiene su propio estado vacío, así que no repite el mensaje.
const soloDaySummary = computed<string | null>(() => {
  if (!isSoloProfile.value || !isViewingToday.value) return null
  if (agendaStatus.value !== 'ready' || !barbershopTimezone.value) return null
  const summary = summarizeDay(entries.value, Date.now())
  if (summary.total === 0) return null
  if (summary.remaining === 0) return 'No te quedan turnos por atender hoy.'
  const count = `${summary.remaining} ${summary.remaining === 1 ? 'turno por atender' : 'turnos por atender'}`
  if (summary.inProgress) {
    return `${count} · En curso hasta las ${formatAgendaTime(summary.inProgress.endsAt)}`
  }
  if (summary.next) {
    return `${count} · Siguiente a las ${formatAgendaTime(summary.next.startsAt)} · ${summary.next.serviceName}`
  }
  return count
})

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
  if (!barbershopTimezone.value || !agendaDate.value) return null
  const tz = barbershopTimezone.value
  const date = agendaDate.value

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
  if (!bounds || !agendaDate.value || !barbershopTimezone.value) return 0
  const start = minutesIntoCivilDate(entry.startsAt, agendaDate.value, barbershopTimezone.value)
  // Fin sin recortar, igual que en timelineBounds: la ficha ocupa su
  // duración real aunque cruce medianoche.
  const end = minutesSinceCivilMidnight(entry.endsAt, agendaDate.value, barbershopTimezone.value)
  const span = bounds.end - bounds.start
  // Ancho mínimo visual del 4%: una ficha muy corta sigue siendo legible en
  // la línea de tiempo sin que eso cambie su duración real.
  return Math.max(((end - start) / span) * 100, 4)
}

// Ficha ensanchada al pasar el cursor o tocarla: crece hacia la derecha, salvo
// las que arrancan cerca del final del carril, que se anclan por su borde
// derecho y crecen hacia la izquierda para no salirse del contenedor. El
// umbral (~18 % del carril a zoom 1) equivale a unos 200 px, el ancho al que
// se ensancha una ficha (--expandido en CSS); con zoom el carril es más ancho,
// así que el mismo ancho en px es un porcentaje menor.
function timelineSlipStyle(entry: DailyAgendaEntry): {
  left?: string
  right?: string
  width: string
} {
  const bounds = timelineBounds.value
  if (!bounds || !agendaDate.value || !barbershopTimezone.value) return { left: '0%', width: '0%' }
  const start = minutesIntoCivilDate(entry.startsAt, agendaDate.value, barbershopTimezone.value)
  const width = timelineSlipWidthPercent(entry)
  const leftPercent = ((start - bounds.start) / (bounds.end - bounds.start)) * 100
  if (leftPercent > 100 - 18 / timelineZoom.value) {
    return { right: `${Math.max(0, 100 - leftPercent - width)}%`, width: `${width}%` }
  }
  return { left: timelinePercent(start), width: `${width}%` }
}

// Ficha «seleccionada» con el dedo: en táctil no hay cursor, así que el primer
// toque la ensancha (nombre completo, horario y servicio) y el segundo abre el
// detalle. Con ratón el clic navega directo; el ensanche lo da :hover.
const expandedSlipId = ref<string | null>(null)
let lastPointerType = 'mouse'

function onSlipPointerDown(event: PointerEvent) {
  lastPointerType = event.pointerType
}

function onSlipClick(event: MouseEvent, entryId: string) {
  if (lastPointerType === 'mouse' || expandedSlipId.value === entryId) return
  event.preventDefault()
  expandedSlipId.value = entryId
}

function onDocumentPointerDown(event: PointerEvent) {
  if (!expandedSlipId.value) return
  const target = event.target
  if (!(target instanceof Element) || !target.closest('.daily-agenda-page__timeline-slip')) {
    expandedSlipId.value = null
  }
}

onMounted(() => document.addEventListener('pointerdown', onDocumentPointerDown))
onUnmounted(() => document.removeEventListener('pointerdown', onDocumentPointerDown))

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

function isEntryCancelled(entry: (typeof entries.value)[number]): boolean {
  return entry.status === 'cancelled_by_customer' || entry.status === 'cancelled_by_barber'
}

// Pendiente: confirmado que aún no está en curso (por atender, o ya pasado de
// hora y sin resolver). Es el caso "por hacer" de la agenda.
function isEntryPending(entry: (typeof entries.value)[number]): boolean {
  return entry.status === 'confirmed' && !isEntryInProgress(entry)
}

// Turno en curso: confirmado y con `ahora` dentro de [inicio, fin). Es solo
// presentación (no cambia `confirmed`) y, como el marcador "Ahora", se evalúa
// al pintar la agenda, sin reloj en vivo. Solo aplica al día en curso.
function isEntryInProgress(entry: (typeof entries.value)[number]): boolean {
  if (entry.status !== 'confirmed' || !isViewingToday.value) return false
  const nowMs = Date.now()
  return nowMs >= new Date(entry.startsAt).getTime() && nowMs < new Date(entry.endsAt).getTime()
}

// "Ahora" (estandar-diseno-visual.md §7.2, §9.3): decorativo respecto al
// estado del turno, nunca cambia `confirmed`. Solo se calcula al montar/
// recargar la agenda (sin reloj en vivo): un dato de referencia visual, no
// una fuente de disponibilidad que necesite exactitud al segundo.
const nowMarkerPercent = computed(() => {
  const bounds = timelineBounds.value
  if (!bounds || !agendaDate.value || !barbershopTimezone.value || !isViewingToday.value) {
    return null
  }
  const nowMinute = minutesIntoCivilDate(
    new Date().toISOString(),
    agendaDate.value,
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
  <section
    class="daily-agenda-page"
    :class="{ 'daily-agenda-page--updating': agendaStatus === 'updating' }"
    :aria-busy="agendaStatus === 'updating'"
    aria-labelledby="daily-agenda-page-title"
  >
    <header class="daily-agenda-page__header nv-rise">
      <div>
        <h1 id="daily-agenda-page-title" class="daily-agenda-page__title">Agenda</h1>
        <p v-if="selectedDateLabel" class="daily-agenda-page__date">
          {{ selectedDateLabel
          }}<template v-if="barbershopTimezone"> · Zona {{ barbershopTimezone }}</template>
        </p>
        <p v-if="soloDaySummary" class="daily-agenda-page__summary">{{ soloDaySummary }}</p>
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
    <div v-if="pageStatus === 'loading'" class="daily-agenda-page__loading" role="status">
      <DiamondLoader :label="`Cargando ${v.professionals}…`" />
    </div>

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
      <EmptyScene v-if="barbers.length === 0" scene="team" role="status">
        <template #title>Aún no tienes {{ v.professionalsRegistered }}.</template>
        <template #hint>
          <template v-if="isSoloProfile">
            Crea tu perfil en
            <RouterLink :to="{ name: 'staff-barberos' }">«Mi perfil»</RouterLink>
            para ver tu agenda.
          </template>
          <template v-else>
            Agrega {{ v.oneProfessional }} en la sección
            <RouterLink :to="{ name: 'staff-barberos' }">«{{ v.Professionals }}»</RouterLink>
            para ver su agenda.
          </template>
        </template>
      </EmptyScene>

      <template v-else>
        <div class="daily-agenda-page__controls nv-rise" style="--i: 1">
          <div v-if="!hideBarberPicker" class="daily-agenda-page__picker">
            <label for="daily-agenda-barber-select" class="daily-agenda-page__label">{{
              v.Professional
            }}</label>
            <BarberSelect
              compact
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
              <label
                id="daily-agenda-date-label"
                for="daily-agenda-date-picker"
                class="daily-agenda-page__label"
                >Fecha</label
              >
              <BaseDatePicker
                trigger-id="daily-agenda-date-picker"
                :model-value="selectedDate"
                :today="todayCivilDate"
                :disabled="!canNavigateDates"
                label-id="daily-agenda-date-label"
                @update:model-value="goToDate"
              />
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
            :headline="`${capitalize(v.thisProfessional)} ya no está disponible`"
            role="alert"
          >
            Elige {{ v.anotherProfessional }} {{ v.professional }} en la lista.
          </PageState>

          <!-- Evento 07: error recuperable de la agenda, barbero/fecha
               conservados para reintentar. -->
          <PageState
            v-else-if="agendaStatus === 'error'"
            variant="warning"
            status-label="Atención"
            :headline="`No pudimos cargar la agenda ${v.ofThisProfessional}`"
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
            <div
              v-if="timelineBounds"
              class="daily-agenda-page__timeline-wrapper nv-rise"
              style="--i: 2"
            >
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
                      :class="`daily-agenda-page__timeline-tick-wrap--${tick.align}`"
                      :style="{ left: timelinePercent(tick.minute) }"
                    >
                      <span class="daily-agenda-page__timeline-tick-label">{{ tick.label }}</span>
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

                  <div class="daily-agenda-page__timeline-track nv-wipe">
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
                      v-for="(entry, index) in entries"
                      :key="`timeline-${entry.id}`"
                      tabindex="-1"
                      class="daily-agenda-page__timeline-slip"
                      :class="{
                        'daily-agenda-page__timeline-slip--terminal': entry.status !== 'confirmed',
                        'daily-agenda-page__timeline-slip--current': isEntryInProgress(entry),
                        'daily-agenda-page__timeline-slip--pending': isEntryPending(entry),
                        'daily-agenda-page__timeline-slip--completed': entry.status === 'completed',
                        'daily-agenda-page__timeline-slip--no-show': entry.status === 'no_show',
                        'daily-agenda-page__timeline-slip--cancelled': isEntryCancelled(entry),
                        'daily-agenda-page__timeline-slip--expanded': expandedSlipId === entry.id,
                        'daily-agenda-page__timeline-slip--end':
                          timelineSlipStyle(entry).right !== undefined,
                      }"
                      :style="{ ...timelineSlipStyle(entry), '--i': index + 3 }"
                      :to="{
                        name: 'agenda-detalle-turno',
                        params: { appointmentId: entry.id },
                        query: withQuery({}),
                      }"
                      @pointerdown="onSlipPointerDown"
                      @click.capture="onSlipClick($event, entry.id)"
                    >
                      <span class="daily-agenda-page__timeline-slip-time">{{
                        timelineSlipDetail(entry) === 'full' || expandedSlipId === entry.id
                          ? entryTimeRange(entry)
                          : entryTime(entry)
                      }}</span>
                      <span class="daily-agenda-page__timeline-slip-name">{{
                        entry.attendeeName
                      }}</span>
                      <span
                        v-if="timelineSlipDetail(entry) === 'full' || expandedSlipId === entry.id"
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
            <EmptyScene
              v-if="entries.length === 0"
              scene="agenda"
              class="daily-agenda-page__empty-state"
            >
              <template #title>
                No hay turnos para {{ selectedBarber?.fullName }} {{ emptyStateDateText }}.
              </template>
              <template #hint>
                Cuando alguien reserve, o registres un turno a mano, aparecerá aquí en su hora.
              </template>
              <template #action>
                <BaseButton type="button" variant="primary" @click="goToNewAppointment">
                  Nuevo turno
                </BaseButton>
              </template>
            </EmptyScene>

            <ul
              v-else
              :key="`${selectedBarberId}-${selectedDate}`"
              class="daily-agenda-page__list"
              :aria-label="`Turnos de ${selectedBarber?.fullName}`"
            >
              <li
                v-for="(entry, index) in entries"
                :key="entry.id"
                class="daily-agenda-page__item nv-rise nv-lift"
                :style="{ '--i': index + 3 }"
                :class="{
                  'daily-agenda-page__item--terminal': entry.status !== 'confirmed',
                  'daily-agenda-page__item--current': isEntryInProgress(entry),
                  'daily-agenda-page__item--pending': isEntryPending(entry),
                  'daily-agenda-page__item--completed': entry.status === 'completed',
                  'daily-agenda-page__item--no-show': entry.status === 'no_show',
                  'daily-agenda-page__item--cancelled': isEntryCancelled(entry),
                }"
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
                  :style="diamondStyle(entry)"
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
  /* Transparente: la tinta y el fondo animado los pone el cascarón. */
  background: transparent;
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
    font-size: var(--font-size-title-page);
    line-height: 46px;
  }
}

.daily-agenda-page__date {
  margin: var(--space-1) 0 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  color: var(--color-on-strong-muted);
}

/* «Mi día» del barbero individual: la única línea de latón del encabezado, para
   que lo que sigue en la jornada se lea antes que la fecha. */
.daily-agenda-page__summary {
  margin: var(--space-2) 0 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  font-weight: 500;
  color: var(--color-brand-accent-surface);
}

.daily-agenda-page__cta {
  flex-shrink: 0;
  /* Sobre tinta, el blanco puro de la variante secundaria deslumbra: el CTA
     usa el mismo latón sobre fondo transparente que Anterior/Siguiente (ver
     `.daily-agenda-page__date-nav`), para que todos los botones de la
     pantalla sean de una sola familia. */
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.daily-agenda-page__cta:hover:not(:disabled) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

/* Aviso solo para lector de pantalla: fuera del flujo para que empezar una
   carga no empuje la agenda hacia abajo y la devuelva al llegar la
   respuesta (el "espasmo" al cambiar de fecha o barbero). El aviso visual es
   el atenuado de la agenda anterior, de abajo. */
.daily-agenda-page__updating {
  position: absolute;
  width: 1px;
  height: 1px;
  margin: -1px;
  padding: 0;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}

/* La agenda anterior se atenúa solo si la respuesta tarda: con la transición
   demorada, una carga rápida (decenas de ms) no parpadea. Al volver a
   'ready' la clase sale y el regreso a opacidad plena es inmediato. */
.daily-agenda-page__timeline-wrapper,
.daily-agenda-page__list,
.daily-agenda-page__empty-state {
  transition: opacity 0.15s ease;
}

.daily-agenda-page--updating .daily-agenda-page__timeline-wrapper,
.daily-agenda-page--updating .daily-agenda-page__list,
.daily-agenda-page--updating .daily-agenda-page__empty-state {
  opacity: 0.55;
  transition-delay: 0.25s;
}

@media (prefers-reduced-motion: reduce) {
  .daily-agenda-page__timeline-wrapper,
  .daily-agenda-page__list,
  .daily-agenda-page__empty-state {
    transition: none;
  }
}

.daily-agenda-page__loading {
  display: flex;
  flex: 1;
  flex-direction: column;
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
  font-size: var(--font-size-caption);
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
  display: flex;
  flex: 1 1 180px;
  flex-direction: column;
  gap: var(--space-2);
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
  --btn-height: 40px;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.daily-agenda-page__date-nav
  :deep(.base-button--ghost:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
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
  background-color: transparent;
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
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
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent);
  border-left-color: var(--color-brand-accent-surface);
}

/* Estados de la ficha, todos con la misma receta: velo tenue + contorno y
   filete del color del estado, sobre el azul general. Van después de
   --terminal para ganarle el contorno.
   - Pendiente (confirmado por atender): amarillo, el "por hacer".
   - Completado: verde apagado de la familia de éxito, el "ya hecho".
   - Cancelado: rosa apagado de la familia de peligro.
   - No se presentó: lila apagado, distinto de los otros tres y del latón. */
.daily-agenda-page__item--pending {
  background-color: color-mix(in srgb, var(--color-pending-on-strong) 8%, transparent);
  border-color: color-mix(in srgb, var(--color-pending-on-strong) 50%, transparent);
  border-left-color: var(--color-pending-on-strong);
}

.daily-agenda-page__item--pending :deep(.base-badge) {
  color: var(--color-pending-on-strong);
}

.daily-agenda-page__item--completed {
  background-color: color-mix(in srgb, var(--color-success-on-strong) 7%, transparent);
  border-color: color-mix(in srgb, var(--color-success-on-strong) 40%, transparent);
  border-left-color: var(--color-success-on-strong);
}

.daily-agenda-page__item.daily-agenda-page__item--completed :deep(.base-badge) {
  color: var(--color-success-on-strong);
}

.daily-agenda-page__item--no-show {
  background-color: color-mix(in srgb, var(--color-no-show-on-strong) 8%, transparent);
  border-color: color-mix(in srgb, var(--color-no-show-on-strong) 45%, transparent);
  border-left-color: var(--color-no-show-on-strong);
}

.daily-agenda-page__item.daily-agenda-page__item--no-show :deep(.base-badge) {
  color: var(--color-no-show-on-strong);
}

/* Turno cancelado (por el cliente o por el barbero): rosa apagado de la
   familia de peligro, con un velo tenue, para separarlo a simple vista de los
   completados sin gritar como una alerta. Va después de --terminal para
   ganarle el color de contorno. */
.daily-agenda-page__item--cancelled {
  background-color: color-mix(in srgb, var(--color-danger-on-strong) 7%, transparent);
  border-color: color-mix(in srgb, var(--color-danger-on-strong) 45%, transparent);
  border-left-color: var(--color-danger-on-strong);
}

.daily-agenda-page__item.daily-agenda-page__item--cancelled :deep(.base-badge) {
  color: var(--color-danger-on-strong);
}

/* Turno en curso ("En proceso"): el azul general de la agenda (tinta), sin
   velo, con contorno y filete en azul claro — el azul es su color de estado. */
.daily-agenda-page__item--current {
  background-color: var(--color-surface-strong);
  border-color: color-mix(in srgb, var(--color-info-on-strong) 55%, transparent);
  border-left-color: var(--color-info-on-strong);
}

.daily-agenda-page__item.daily-agenda-page__item--current :deep(.base-badge) {
  color: var(--color-info-on-strong);
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
  font-size: var(--font-size-body);
  font-weight: 400;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.02em;
  color: var(--color-on-strong-soft);
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
  font-size: var(--font-size-body-lg);
  font-weight: 600;
  color: var(--color-on-strong);
}

.daily-agenda-page__item-service {
  font-family: var(--font-family-base);
  margin-top: 2px;
  font-size: var(--font-size-body-sm);
  line-height: 16px;
  color: var(--color-on-strong-soft);
}

@media (prefers-reduced-motion: reduce) {
  .daily-agenda-page__timeline-slip,
  .daily-agenda-page__timeline-mark--now::after {
    animation: none;
  }
}

/* Reflejo sutil al pasar sobre un turno de la lista: sube el velo y el
   nombre toma el latón; el desplazamiento de 3px viene de `.nv-lift`. */
.daily-agenda-page__item {
  position: relative;
  transition:
    transform var(--motion-duration-base) var(--motion-ease-out),
    box-shadow var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard),
    background-color var(--motion-duration-base) var(--motion-easing-standard);
}

@media (hover: hover) {
  .daily-agenda-page__item:hover {
    background-color: color-mix(in srgb, var(--color-on-strong) 5%, transparent);
  }
}

.daily-agenda-page__item-main:hover {
  text-decoration: none;
}

.daily-agenda-page__item:hover .daily-agenda-page__item-name {
  color: var(--color-brand-accent-surface);
}

@media (max-width: 1023px) {
  .daily-agenda-page__cta {
    width: 100%;
  }

  /* El rótulo «Fecha» encabeza la fila completa (Anterior · fecha ·
     Siguiente) alineado a la izquierda, igual que «Barbero» sobre su
     selector: el contenedor del selector se disuelve y cada pieza ocupa su
     área de la cuadrícula. */
  .daily-agenda-page__date-nav {
    display: grid;
    align-items: end;
    width: 100%;
    column-gap: var(--space-2);
    row-gap: var(--space-2);
    grid-template-columns: minmax(0, 1fr) minmax(0, 1.65fr) minmax(0, 1.05fr);
    grid-template-areas:
      'label label label'
      'prev date next';
  }

  .daily-agenda-page__date-input-wrap {
    display: contents;
  }

  .daily-agenda-page__date-input-wrap > .daily-agenda-page__label {
    grid-area: label;
  }

  .daily-agenda-page__date-input-wrap > :not(.daily-agenda-page__label) {
    grid-area: date;
    min-width: 0;
  }

  .daily-agenda-page__date-nav > :first-child {
    grid-area: prev;
  }

  .daily-agenda-page__date-nav > :last-child {
    grid-area: next;
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
    font-size: var(--font-size-body-sm);
    color: var(--color-on-strong-soft);
  }

  .daily-agenda-page__item-name {
    font-size: var(--font-size-body);
    line-height: 22px;
  }

  .daily-agenda-page__item-service {
    font-size: var(--font-size-body-sm);
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
    width: 32px;
    height: 32px;
    padding: 0;
    background-color: transparent;
    color: var(--color-brand-accent-surface);
    border: 1px solid color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
    border-radius: 2px;
    font-size: var(--font-size-body);
    line-height: 1;
    cursor: pointer;
    transition:
      background-color var(--motion-duration-fast) var(--motion-easing-standard),
      transform var(--motion-duration-fast) var(--motion-ease-out);
  }

  .daily-agenda-page__timeline-zoom-btn:active:not(:disabled) {
    transform: scale(0.9);
  }

  .daily-agenda-page__timeline-zoom-btn:hover:not(:disabled) {
    background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
  }

  .daily-agenda-page__timeline-zoom-btn:disabled {
    opacity: 0.42;
    cursor: not-allowed;
  }

  .daily-agenda-page__timeline-zoom-label {
    min-width: 48px;
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
     del scroll horizontal). translateX en vez de left/right: el wrap ya
     lleva su posición por `:style`, y un estilo en línea siempre gana sobre
     cualquier `left`/`right` de clase. */
  .daily-agenda-page__timeline-tick-wrap--center {
    transform: translateX(-50%);
  }

  /* El desplazamiento va en el wrap, no en la etiqueta: el wrap es el que
     define el área desplazable del carril, y con `left: 100%` su caja (aún
     sin desplazar) sobresalía ~31 px del borde derecho y daba un scroll
     horizontal de más incluso a zoom 1. */
  .daily-agenda-page__timeline-tick-wrap--end {
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
    font-size: var(--font-size-caption);
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
    border: var(--border-width-normal) solid
      color-mix(in srgb, var(--color-on-strong) 16%, transparent);
    border-radius: 2px;
  }

  .daily-agenda-page__timeline-guide {
    position: absolute;
    top: 0;
    bottom: 0;
    border-left: var(--border-width-normal) dashed
      color-mix(in srgb, var(--color-on-strong) 10%, transparent);
  }

  .daily-agenda-page__timeline-tick {
    position: absolute;
    top: 0;
    bottom: 0;
    border-left: var(--border-width-normal) dashed
      color-mix(in srgb, var(--color-on-strong) 24%, transparent);
  }

  .daily-agenda-page__timeline-mark {
    position: absolute;
    top: 0;
    bottom: 0;
    z-index: 1;
    border-left: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  }

  /* Rombo con latido sobre la línea de "Ahora": el único elemento de la
     pantalla que se mueve sin que nadie lo toque, y solo porque el tiempo
     real avanza. */
  .daily-agenda-page__timeline-mark--now::before {
    content: '';
    position: absolute;
    top: -5px;
    left: -6px;
    width: 10px;
    height: 10px;
    background-color: var(--color-brand-accent-surface);
    transform: rotate(45deg);
  }

  /* Onda que sale del rombo: borde que escala y se apaga. Solo transform y
     opacity, por pasos; animar `box-shadow` forzaba un repintado por cuadro. */
  .daily-agenda-page__timeline-mark--now::after {
    content: '';
    position: absolute;
    top: -5px;
    left: -6px;
    width: 10px;
    height: 10px;
    border: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
    opacity: 0;
    transform: rotate(45deg);
    animation: agenda-now-ring 2.4s steps(24, end) infinite;
  }

  @keyframes agenda-now-ring {
    from {
      opacity: 0.9;
      transform: rotate(45deg) scale(1);
    }

    to {
      opacity: 0;
      transform: rotate(45deg) scale(2.8);
    }
  }

  /* Cada ficha crece desde su hora de inicio (borde izquierdo) con el
     escalonado de la lista: el día se "dibuja" en el carril en vez de
     aparecer de golpe. Solo escala/opacidad, nunca el ancho real. */
  .daily-agenda-page__timeline-slip {
    transform-origin: left center;
    animation: nava-grow-x 520ms var(--motion-ease-out) backwards;
    animation-delay: calc(min(var(--i, 0), 14) * var(--motion-stagger));
    transition:
      box-shadow var(--motion-duration-base) var(--motion-easing-standard),
      background-color var(--motion-duration-base) var(--motion-easing-standard);
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
    /* Tinte del estado; la ficha ensanchada lo apoya sobre tinta opaca. */
    --slip-tint: transparent;
    background-color: var(--slip-tint);
    border: 0;
    border-left: 3px solid var(--color-accent-brass);
    border-radius: 2px;
    color: var(--color-on-strong);
    text-decoration: none;
  }

  .daily-agenda-page__timeline-slip--current {
    background-color: var(--color-surface-strong);
    border: var(--border-width-normal) solid
      color-mix(in srgb, var(--color-info-on-strong) 55%, transparent);
    border-left: 3px solid var(--color-info-on-strong);
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
    border: var(--border-width-normal) solid
      color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent);
    border-left: 3px solid var(--color-brand-accent-surface);
    color: var(--color-on-strong);
  }

  /* Pendiente (amarillo) y completado (verde): misma receta que la ficha de
     lista; el carril fija `border: 0` en la base, por eso llevan su contorno
     completo. Después de --terminal para ganarle el color. */
  .daily-agenda-page__timeline-slip--pending {
    --slip-tint: color-mix(in srgb, var(--color-pending-on-strong) 8%, transparent);
    background-color: var(--slip-tint);
    border: var(--border-width-normal) solid
      color-mix(in srgb, var(--color-pending-on-strong) 50%, transparent);
    border-left: 3px solid var(--color-pending-on-strong);
  }

  .daily-agenda-page__timeline-slip--completed {
    --slip-tint: color-mix(in srgb, var(--color-success-on-strong) 7%, transparent);
    background-color: var(--slip-tint);
    border-color: color-mix(in srgb, var(--color-success-on-strong) 40%, transparent);
    border-left-color: var(--color-success-on-strong);
  }

  .daily-agenda-page__timeline-slip--no-show {
    --slip-tint: color-mix(in srgb, var(--color-no-show-on-strong) 8%, transparent);
    background-color: var(--slip-tint);
    border-color: color-mix(in srgb, var(--color-no-show-on-strong) 45%, transparent);
    border-left-color: var(--color-no-show-on-strong);
  }

  /* Cancelado: rosa apagado; después de --terminal para ganarle el contorno. */
  .daily-agenda-page__timeline-slip--cancelled {
    --slip-tint: color-mix(in srgb, var(--color-danger-on-strong) 7%, transparent);
    background-color: var(--slip-tint);
    border-color: color-mix(in srgb, var(--color-danger-on-strong) 45%, transparent);
    border-left-color: var(--color-danger-on-strong);
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

  /* Ficha ensanchada (cursor encima o tocada): nunca menos ancha que su
     contenido, así el nombre se lee entero; sobre tinta opaca y por encima de
     las vecinas para que no se transparente el texto de otra ficha. El hover
     solo cuenta con puntero fino: en táctil queda «pegado» tras el toque y la
     selección la da --expanded. */
  .daily-agenda-page__timeline-slip--expanded {
    z-index: 3;
    min-width: max-content;
    overflow: visible;
    background: linear-gradient(var(--slip-tint), var(--slip-tint)), var(--color-surface-strong);
    box-shadow: var(--shadow-dialog);
  }

  @media (hover: hover) {
    .daily-agenda-page__timeline-slip:hover {
      z-index: 3;
      min-width: max-content;
      overflow: visible;
      background: linear-gradient(var(--slip-tint), var(--slip-tint)), var(--color-surface-strong);
      box-shadow: var(--shadow-dialog);
    }

    .daily-agenda-page__timeline-slip:hover .daily-agenda-page__timeline-slip-name {
      overflow: visible;
      text-overflow: clip;
    }
  }

  .daily-agenda-page__timeline-slip--expanded .daily-agenda-page__timeline-slip-name,
  .daily-agenda-page__timeline-slip--expanded .daily-agenda-page__timeline-slip-service {
    overflow: visible;
    text-overflow: clip;
  }

  .daily-agenda-page__timeline-slip-time {
    font-size: var(--font-size-caption);
    font-weight: 400;
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.03em;
    color: var(--color-on-strong-soft);
  }

  .daily-agenda-page__timeline-slip-name {
    overflow: hidden;
    font-size: var(--font-size-body-sm);
    font-weight: 600;
    color: var(--color-on-strong);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .daily-agenda-page__timeline-slip-service {
    overflow: hidden;
    font-size: var(--font-size-caption);
    color: var(--color-on-strong-soft);
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
  font-size: var(--font-size-caption);
  font-weight: 600;
  line-height: 14px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

:deep(.base-badge)::before {
  content: '';
  flex-shrink: 0;
  width: 5px;
  height: 5px;
  /* Rombo (cuadrado girado 45°) en vez de círculo: el marcador de estado de
     esta pantalla, igual que el rombo de la marca. */
  transform: rotate(45deg);
  /* currentColor, no --badge-dot-color: el punto sigue el mismo color de
     texto que ya resuelve la variante (y su atenuado terminal de abajo),
     sin depender de la mecánica de padding negativo del prop `dot`
     original de BaseBadge (pensada para su badge con caja). */
  background-color: currentColor;
}

/* Parpadeo suave y esporádico del rombo: casi todo el ciclo queda quieto y
   solo un tramo corto baja la opacidad y la escala, sin movimiento lateral.
   Duración y desfase salen de `--diamond-duration`/`--diamond-delay` (uno
   propio por turno, ver diamondStyle) para que no se sincronicen. */
@keyframes agenda-diamond-blink {
  0%,
  62%,
  100% {
    opacity: 1;
    transform: rotate(45deg) scale(1);
  }

  80% {
    opacity: 0.55;
    transform: rotate(45deg) scale(0.85);
  }
}

:deep(.base-badge)::before {
  animation: agenda-diamond-blink var(--diamond-duration, 7s) ease-in-out var(--diamond-delay, 0s)
    infinite;
}

@media (prefers-reduced-motion: reduce) {
  :deep(.base-badge)::before {
    animation: none;
  }
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
    font-size: var(--font-size-caption);
  }
}
</style>
