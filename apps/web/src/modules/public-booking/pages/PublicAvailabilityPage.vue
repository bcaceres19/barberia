<script setup lang="ts">
// Página coordinadora de la exploración pública de fechas y horarios
// (HU-095, CA-095-01 a CA-095-05). Sin mockup asignado: composición libre
// dentro de NAVA / Tailored Grid (DEC-078). Consume el motor de
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
import { BaseButton, EmptyState, PageState } from '@/shared/ui'
import { formatCivilDateFull } from '@/shared/time/civilDate'
import { formatTimeInTimezone } from '@/shared/time/formatInstant'
import { listPublicAvailability } from '../api/listPublicAvailabilityApi'
import { groupSlotsByCivilDate, type AvailabilityDay } from '../model/availabilityCalendar'
import type { PublicAvailabilitySlot } from '../model/publicAvailabilityOutcome'

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

function setSlotRef(el: Element | { $el?: Element } | null, index: number) {
  if (el instanceof HTMLElement) slotRefs.value[index] = el
}

let requestToken = 0

async function load() {
  const token = ++requestToken
  screenState.value = { status: 'loading' }
  activeDayIndex.value = 0
  activeSlotIndex.value = 0
  selectedSlot.value = null
  slotRefs.value = []

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
watch(activeDayIndex, () => {
  slotRefs.value = []
  const day = activeDay.value
  const preserved = day?.slots.findIndex((s) => s.startsAt === selectedSlot.value?.startsAt) ?? -1
  activeSlotIndex.value = preserved >= 0 ? preserved : 0
})

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
  return `${date}, ${time} · zona horaria de la barbería: ${timezone.value} · dura ${durationMinutes.value} min`
})

const unexpectedErrorMessage = computed(() => {
  const state = screenState.value
  if (state.status !== 'unexpected-error') return ''
  return state.requestId
    ? `Inténtalo de nuevo. Si continúa, comparte este código con soporte: ${state.requestId}.`
    : 'Inténtalo de nuevo en unos segundos.'
})
</script>

<template>
  <main v-if="screenState.status !== 'success'" class="availability availability--state">
    <h1 class="visually-hidden">Elegir fecha y hora</h1>
    <PageState
      v-if="screenState.status === 'loading'"
      variant="loading"
      headline="Cargando disponibilidad…"
      role="status"
    />
    <PageState
      v-else-if="screenState.status === 'not-found'"
      variant="danger"
      status-label="Error"
      headline="No encontramos ese enlace"
      role="alert"
    >
      <template #default
        >Revisa que copiaste la dirección completa, o pídele al barbero que te la vuelva a
        compartir.</template
      >
      <template #action>
        <BaseButton variant="secondary" @click="retry">Reintentar</BaseButton>
      </template>
    </PageState>
    <PageState
      v-else-if="screenState.status === 'network-error'"
      variant="warning"
      status-label="Atención"
      headline="No pudimos conectar"
      role="alert"
    >
      <template #default>Revisa tu conexión e inténtalo de nuevo.</template>
      <template #action>
        <BaseButton variant="primary" @click="retry">Reintentar</BaseButton>
      </template>
    </PageState>
    <PageState
      v-else
      variant="danger"
      status-label="Error"
      headline="Ocurrió un error inesperado"
      role="alert"
    >
      <template #default>{{ unexpectedErrorMessage }}</template>
      <template #action>
        <BaseButton variant="primary" @click="retry">Reintentar</BaseButton>
      </template>
    </PageState>
  </main>

  <main v-else class="availability availability--list">
    <div class="availability__container">
      <h1 class="availability__title">Elige fecha y hora</h1>

      <!-- CA-095-04: ninguna franja en toda la ventana pública vigente
           (festivo total, servicio/barbero sin agenda compatible, etc.) es
           un vacío genuino, distinto de carga o error. -->
      <EmptyState
        v-if="days.length === 0"
        message="No hay franjas disponibles en la ventana de reserva vigente. Vuelve a intentarlo más adelante."
      />

      <template v-else>
        <p class="availability__timezone">
          Horas en la zona horaria de la barbería: {{ timezone }}
        </p>

        <div class="availability__day-nav">
          <BaseButton
            type="button"
            variant="secondary"
            :disabled="activeDayIndex === 0"
            aria-label="Ver el día anterior con disponibilidad"
            @click="goToPreviousDay"
          >
            ‹ Anterior
          </BaseButton>
          <p class="availability__day-label" role="status" aria-live="polite">
            {{ activeDay ? formatCivilDateFull(activeDay.civilDate) : '' }}
          </p>
          <BaseButton
            type="button"
            variant="secondary"
            :disabled="activeDayIndex >= days.length - 1"
            aria-label="Ver el día siguiente con disponibilidad"
            @click="goToNextDay"
          >
            Siguiente ›
          </BaseButton>
        </div>

        <ul
          class="availability__slots"
          role="radiogroup"
          aria-label="Horas disponibles"
          @keydown="onSlotsKeydown"
        >
          <li
            v-for="(slot, index) in activeDay?.slots ?? []"
            :key="slot.startsAt"
            :ref="(el) => setSlotRef(el as Element | null, index)"
            class="availability__slot"
            :class="{ 'availability__slot--selected': selectedSlot?.startsAt === slot.startsAt }"
            role="radio"
            :aria-checked="selectedSlot?.startsAt === slot.startsAt"
            :tabindex="index === activeSlotIndex ? 0 : -1"
            @click="selectSlot(slot)"
          >
            <span class="availability__slot-check" aria-hidden="true"></span>
            <span class="availability__slot-time">{{ slotTimeLabel(slot) }}</span>
          </li>
        </ul>

        <!-- Nunca insinúa que consultar o elegir una hora aquí reserva el
             turno (RN-DIS-03, fuera de alcance de HU-095): "elegida", no
             "reservada" ni "confirmada". -->
        <p v-if="selectionSummary" class="availability__summary" role="status">
          Franja elegida: {{ selectionSummary }}
        </p>
      </template>
    </div>
  </main>
</template>

<style scoped>
.availability {
  display: flex;
  min-height: 100dvh;
  justify-content: center;
  padding: var(--space-6) var(--space-4);
}

.availability--state {
  align-items: center;
  background-color: var(--color-surface-strong);
}

.availability--list {
  background-color: var(--color-canvas);
}

.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  margin: -1px;
  padding: 0;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.availability__container {
  display: flex;
  width: 100%;
  max-width: 640px;
  flex-direction: column;
  gap: var(--space-5);
}

.availability__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: 32px;
  line-height: 40px;
  letter-spacing: -0.015em;
  color: var(--color-text-primary);
}

.availability__timezone {
  margin: 0;
  color: var(--color-text-secondary);
}

.availability__day-nav {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: var(--space-2) var(--space-3);
}

/* `order: -1` + `flex-basis: 100%` deja la fecha siempre en su propia línea
   encima de los botones: una fecha larga ("miércoles, 16 de septiembre de
   2026") nunca compite por ancho con "‹ Anterior"/"Siguiente ›" en 320 px,
   sin scroll horizontal (evidencia responsiva). */
.availability__day-label {
  order: -1;
  flex-basis: 100%;
  margin: 0;
  text-align: center;
  font-weight: 600;
  color: var(--color-text-primary);
}

.availability__slots {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  padding: 0;
  margin: 0;
  list-style: none;
}

.availability__slot {
  display: flex;
  min-width: 44px;
  min-height: 44px;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-4);
  cursor: pointer;
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: 4px;
}

.availability__slot:hover {
  border-color: var(--color-accent-brass);
}

.availability__slot:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-canvas),
    0 0 0 4px var(--color-focus);
}

/* La selección se marca con el ícono de check Y el borde de énfasis, nunca
   solo con un cambio de color (CA-095-05, mismo criterio que
   PublicBarberSelectionPage.vue/CA-092-05). */
.availability__slot--selected {
  border-color: var(--color-accent-brass);
  border-width: var(--border-width-emphasis);
}

.availability__slot-check {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: 50%;
}

.availability__slot--selected .availability__slot-check {
  background-color: var(--color-accent-brass);
  border-color: var(--color-accent-brass);
}

.availability__slot-time {
  font-weight: 600;
  color: var(--color-text-primary);
}

.availability__summary {
  margin: 0;
  padding: var(--space-4);
  color: var(--color-text-primary);
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: 4px;
}

@media (min-width: 1024px) {
  .availability__title {
    font-size: 40px;
    line-height: 48px;
  }
}
</style>
