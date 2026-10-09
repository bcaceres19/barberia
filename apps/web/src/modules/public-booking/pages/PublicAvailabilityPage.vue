<script setup lang="ts">
// Página coordinadora de la exploración pública de fechas y horarios
// (HU-095, CA-095-01 a CA-095-05). Sin mockup asignado: composición libre
// dentro de NAVA / Tailored Grid (DEC-078), alojada en el cascarón de la
// reserva pública (DEC-111). Consume el motor de
// disponibilidad de HU-094 en una sola consulta por contexto (slug +
// serviceId + barberId): el contrato no admite una fecha como parámetro,
// devuelve todos los inicios válidos de la ventana pública vigente en una
// sola respuesta, así que "explorar fechas" es una agrupación en el
// cliente sobre datos ya recibidos (`groupSlotsByCivilDate`), nunca una
// consulta nueva por cada fecha visitada (RN-DIS-03: consultar no reserva).
//
// Revalidación (CA-095-03): un cambio de `serviceId`/`barberId` (mismo
// componente reutilizado por vue-router al cambiar los parámetros de ruta)
// dispara una nueva carga; una respuesta tardía de una carga ya
// reemplazada se descarta con un token monótono, mismo patrón que
// PublicBarberSelectionPage.vue.
import { computed, nextTick, ref, watch } from 'vue'
import { formatCivilDateFull } from '@/shared/time/civilDate'
import { formatTimeInTimezone } from '@/shared/time/formatInstant'
import BookingStateScreen from '../components/BookingStateScreen.vue'
import { listPublicAvailability } from '../api/listPublicAvailabilityApi'
import {
  civilDateChip,
  groupSlotsByCivilDate,
  type AvailabilityDay,
} from '../model/availabilityCalendar'
import type { PublicAvailabilitySlot } from '../model/publicAvailabilityOutcome'
import { usePublicVocabulary } from '../model/publicVocabulary'

interface Props {
  /** Identificador del enlace público, tal como llega del parámetro de
   * ruta `:slug`. No confiado: el servidor lo vuelve a resolver desde cero. */
  slug: string
  /** Identificador del servicio activo ya elegido, tal como llega del
   * parámetro de ruta `:serviceId`. No confiado: el servidor revalida
   * pertenencia, vigencia y asignación en cada lectura. */
  serviceId: string
  /** Identificador del barbero ya elegido, tal como llega del parámetro de
   * ruta `:barberId`. No confiado: el servidor revalida asignación vigente
   * al servicio activo en cada lectura. */
  barberId: string
}
const props = defineProps<Props>()
const v = usePublicVocabulary()

type ScreenState =
  | { status: 'loading' }
  | { status: 'success'; days: AvailabilityDay[]; durationMinutes: number; timezone: string }
  | { status: 'not-found' }
  | { status: 'network-error' }
  | { status: 'unexpected-error'; requestId?: string }

const screenState = ref<ScreenState>({ status: 'loading' })
const activeDayIndex = ref(0)
const activeSlotIndex = ref(0)
// Selección global (no por día): navegar entre días nunca descarta una
// franja ya elegida en otro día de la misma ventana (CA-095-04, "conservan
// selecciones previas válidas"). Solo un contexto nuevo (servicio/barbero
// distinto) o un reintento tras error la reinicia, porque en ese caso la
// ventana completa se vuelve a resolver desde cero.
const selectedSlot = ref<PublicAvailabilitySlot | null>(null)
const slotRefs = ref<HTMLElement[]>([])
const dayRefs = ref<HTMLElement[]>([])
const ticketRef = ref<HTMLElement | null>(null)

function setSlotRef(el: Element | { $el?: Element } | null, index: number) {
  if (el instanceof HTMLElement) slotRefs.value[index] = el
}

function setDayRef(el: Element | { $el?: Element } | null, index: number) {
  if (el instanceof HTMLElement) dayRefs.value[index] = el
}

let requestToken = 0

async function load() {
  const token = ++requestToken
  screenState.value = { status: 'loading' }
  activeDayIndex.value = 0
  activeSlotIndex.value = 0
  selectedSlot.value = null
  slotRefs.value = []
  dayRefs.value = []

  const outcome = await listPublicAvailability(props.slug, props.serviceId, props.barberId)
  // Una respuesta tardía de una carga ya reemplazada (contexto cambió de
  // nuevo mientras esta esperaba) se descarta: nunca sobrescribe el estado
  // de la carga vigente (CA-095-03).
  if (token !== requestToken) return

  switch (outcome.kind) {
    case 'success': {
      const { availability } = outcome
      const days = groupSlotsByCivilDate(availability.slots, availability.timezone)
      screenState.value = {
        status: 'success',
        days,
        durationMinutes: availability.durationMinutes,
        timezone: availability.timezone,
      }
      return
    }
    case 'not-found':
      screenState.value = { status: 'not-found' }
      return
    case 'network-error':
      screenState.value = { status: 'network-error' }
      return
    case 'unexpected-error':
      screenState.value = { status: 'unexpected-error', requestId: outcome.requestId }
  }
}

watch(() => [props.slug, props.serviceId, props.barberId], load, { immediate: true })

const retry = () => {
  void load()
}

const days = computed(() => (screenState.value.status === 'success' ? screenState.value.days : []))

const activeDay = computed<AvailabilityDay | null>(() => days.value[activeDayIndex.value] ?? null)

const timezone = computed(() =>
  screenState.value.status === 'success' ? screenState.value.timezone : '',
)

const durationMinutes = computed(() =>
  screenState.value.status === 'success' ? screenState.value.durationMinutes : 0,
)

// Al cambiar de día, el índice con roving tabindex apunta a la franja ya
// elegida si pertenece a este día; si no, vuelve al primer elemento. La
// selección en sí (`selectedSlot`) nunca se toca aquí.
watch(activeDayIndex, async () => {
  slotRefs.value = []
  const day = activeDay.value
  const preserved = day?.slots.findIndex((s) => s.startsAt === selectedSlot.value?.startsAt) ?? -1
  activeSlotIndex.value = preserved >= 0 ? preserved : 0
  // La tira de fechas sigue al día activo (también cuando se cambia con
  // Anterior/Siguiente); `scrollIntoView` solo desplaza la tira, no la página.
  await nextTick()
  dayRefs.value[activeDayIndex.value]?.scrollIntoView?.({
    inline: 'center',
    block: 'nearest',
    behavior: 'smooth',
  })
})

function chooseDay(index: number) {
  activeDayIndex.value = index
}

function goToPreviousDay() {
  if (activeDayIndex.value > 0) activeDayIndex.value -= 1
}

function goToNextDay() {
  if (activeDayIndex.value < days.value.length - 1) activeDayIndex.value += 1
}

function selectSlot(slot: PublicAvailabilitySlot) {
  selectedSlot.value = slot
  const index = activeDay.value?.slots.findIndex((s) => s.startsAt === slot.startsAt) ?? -1
  if (index >= 0) activeSlotIndex.value = index
  void revealTicket()
}

// La barra de acción fija al pie puede tapar la ficha recién aparecida en una
// ventana baja: se trae a la vista (el `scroll-margin-bottom` de la ficha
// reserva el alto de la barra) sin mover el foco.
async function revealTicket() {
  await nextTick()
  ticketRef.value?.scrollIntoView?.({ block: 'nearest', behavior: 'smooth' })
}

async function focusActiveSlot() {
  await nextTick()
  slotRefs.value[activeSlotIndex.value]?.focus()
}

function moveActiveSlot(delta: number) {
  const count = activeDay.value?.slots.length ?? 0
  if (count === 0) return
  activeSlotIndex.value = (activeSlotIndex.value + delta + count) % count
  void focusActiveSlot()
}

function onSlotsKeydown(event: KeyboardEvent) {
  switch (event.key) {
    case 'ArrowDown':
    case 'ArrowRight':
      event.preventDefault()
      moveActiveSlot(1)
      break
    case 'ArrowUp':
    case 'ArrowLeft':
      event.preventDefault()
      moveActiveSlot(-1)
      break
    case 'Home':
      event.preventDefault()
      activeSlotIndex.value = 0
      void focusActiveSlot()
      break
    case 'End':
      event.preventDefault()
      activeSlotIndex.value = (activeDay.value?.slots.length ?? 1) - 1
      void focusActiveSlot()
      break
    case ' ':
    case 'Enter': {
      event.preventDefault()
      const slot = activeDay.value?.slots[activeSlotIndex.value]
      if (slot) selectSlot(slot)
      break
    }
  }
}

function slotTimeLabel(slot: PublicAvailabilitySlot): string {
  return formatTimeInTimezone(slot.startsAt, timezone.value)
}

// CA-095-02: hora, fecha, duración y zona siempre anunciadas juntas; nunca
// una hora suelta sin su fecha/zona/duración. Cambiar la zona del
// dispositivo no cambia nada aquí, porque `timezone`/`formatTimeInTimezone`
// nunca leen la zona local (RN-DIS-07).
const selectionSummary = computed(() => {
  if (!selectedSlot.value) return ''
  const day = days.value.find((d) =>
    d.slots.some((s) => s.startsAt === selectedSlot.value!.startsAt),
  )
  if (!day) return ''
  const date = formatCivilDateFull(day.civilDate)
  const time = slotTimeLabel(selectedSlot.value)
  return `${date}, ${time} · zona horaria ${v.value.ofTheBusiness}: ${timezone.value} · dura ${durationMinutes.value} min`
})

// Piezas de la ficha visual de la franja elegida. La oración completa
// (`selectionSummary`) sigue siendo lo que anuncia la tecnología de apoyo;
// la ficha repite lo mismo por partes y se oculta para ella.
const selectedDateLabel = computed(() => {
  const slot = selectedSlot.value
  if (!slot) return ''
  const day = days.value.find((d) => d.slots.some((s) => s.startsAt === slot.startsAt))
  return day ? formatCivilDateFull(day.civilDate) : ''
})

const selectedTimeLabel = computed(() =>
  selectedSlot.value ? slotTimeLabel(selectedSlot.value) : '',
)

const selectedShortLabel = computed(() => {
  const slot = selectedSlot.value
  if (!slot) return ''
  const day = days.value.find((d) => d.slots.some((s) => s.startsAt === slot.startsAt))
  if (!day) return ''
  const chip = civilDateChip(day.civilDate)
  return `${selectedTimeLabel.value} · ${chip.weekday} ${chip.day} ${chip.month}`
})

const failedRequestId = computed(() =>
  screenState.value.status === 'unexpected-error' ? screenState.value.requestId : undefined,
)
</script>

<template>
  <BookingStateScreen
    v-if="screenState.status !== 'success'"
    :status="screenState.status"
    :request-id="failedRequestId"
    title="Elegir fecha y hora"
    loading-headline="Cargando disponibilidad…"
    @retry="retry"
  />

  <main v-else class="pb-page pb-page--bar pb-page--fit">
    <p class="pb-eyebrow">Fecha y hora</p>
    <h1 class="pb-title">Elige fecha y hora</h1>
    <div class="pb-rule" aria-hidden="true"></div>

    <!-- CA-095-04: ninguna franja en toda la ventana pública vigente
         (festivo total, servicio/barbero sin agenda compatible, etc.) es
         un vacío genuino, distinto de carga o error. -->
    <div v-if="days.length === 0" class="pb-empty">
      <span class="pb-empty__mark" aria-hidden="true"></span>
      <p class="pb-empty__message">
        No hay franjas disponibles en la ventana de reserva vigente. Vuelve a intentarlo más
        adelante.
      </p>
    </div>

    <template v-else>
      <p class="pb-zone">Horas en la zona horaria {{ v.ofTheBusiness }}: {{ timezone }}</p>

      <div class="pb-days" role="group" aria-label="Días con disponibilidad">
        <button
          v-for="(day, index) in days"
          :key="day.civilDate"
          :ref="(el) => setDayRef(el as Element | null, index)"
          type="button"
          class="pb-day"
          :style="{ '--pb-i': index }"
          :aria-current="index === activeDayIndex ? 'date' : undefined"
          @click="chooseDay(index)"
        >
          <span class="pb-day__dow">{{ civilDateChip(day.civilDate).weekday }}</span>
          <span class="pb-day__num">{{ civilDateChip(day.civilDate).day }}</span>
          <span class="pb-day__month">{{ civilDateChip(day.civilDate).month }}</span>
        </button>
      </div>

      <div class="pb-daybar">
        <button
          type="button"
          class="pb-step-btn"
          :disabled="activeDayIndex === 0"
          aria-label="Ver el día anterior con disponibilidad"
          @click="goToPreviousDay"
        >
          ‹ Anterior
        </button>
        <p class="pb-daybar__label" role="status" aria-live="polite">
          {{ activeDay ? formatCivilDateFull(activeDay.civilDate) : '' }}
        </p>
        <button
          type="button"
          class="pb-step-btn"
          :disabled="activeDayIndex >= days.length - 1"
          aria-label="Ver el día siguiente con disponibilidad"
          @click="goToNextDay"
        >
          Siguiente ›
        </button>
      </div>

      <ul
        class="pb-slots"
        role="radiogroup"
        aria-label="Horas disponibles"
        @keydown="onSlotsKeydown"
      >
        <li
          v-for="(slot, index) in activeDay?.slots ?? []"
          :key="slot.startsAt"
          :ref="(el) => setSlotRef(el as Element | null, index)"
          class="pb-slot"
          :class="{ 'pb-slot--selected': selectedSlot?.startsAt === slot.startsAt }"
          :style="{ '--pb-i': index }"
          role="radio"
          :aria-checked="selectedSlot?.startsAt === slot.startsAt"
          :tabindex="index === activeSlotIndex ? 0 : -1"
          @click="selectSlot(slot)"
        >
          {{ slotTimeLabel(slot) }}
        </li>
      </ul>

      <!-- Nunca insinúa que consultar o elegir una hora aquí reserva el
           turno (RN-DIS-03, fuera de alcance de HU-095): "elegida", no
           "reservada" ni "confirmada". La oración completa la anuncia la
           tecnología de apoyo; la ficha visual la repite por partes. -->
      <p v-if="selectionSummary" class="pb-sr-only" role="status">
        Franja elegida: {{ selectionSummary }}
      </p>
      <div
        v-if="selectionSummary"
        :key="selectedSlot?.startsAt"
        class="pb-ticket"
        aria-hidden="true"
      >
        <span class="pb-ticket__kicker">Franja elegida</span>
        <span class="pb-ticket__date">{{ selectedDateLabel }}</span>
        <span class="pb-ticket__time">{{ selectedTimeLabel }}</span>
        <span class="pb-ticket__foot">Dura {{ durationMinutes }} min · {{ timezone }}</span>
      </div>

      <Transition name="pb-bar">
        <div v-if="selectedSlot" class="pb-actionbar">
          <p class="pb-actionbar__summary">
            <span class="pb-actionbar__label">Tu horario</span>
            <span class="pb-actionbar__value">{{ selectedShortLabel }}</span>
          </p>
          <RouterLink
            :to="{
              name: 'reserva-publica-cliente',
              params: {
                slug: props.slug,
                serviceId: props.serviceId,
                barberId: props.barberId,
                startsAt: selectedSlot.startsAt,
              },
            }"
            class="pb-cta"
          >
            <span>Continuar</span>
            <span class="pb-cta__arrow" aria-hidden="true">→</span>
          </RouterLink>
        </div>
      </Transition>
    </template>
  </main>
</template>
