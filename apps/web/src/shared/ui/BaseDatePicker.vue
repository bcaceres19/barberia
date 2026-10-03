<script setup lang="ts">
/**
 * BaseDatePicker - Calendario civil compartido por agenda y formularios.
 *
 * Sustituye al `<input type="date">`: el calendario nativo lo pinta cada
 * navegador (blanco y con su formato de locale, incluso sobre la tinta de la
 * agenda) y en Firefox su texto se encimaba con la capa ISO que lo cubría.
 * Aquí el campo es un disparador de texto ISO (el mismo valor que viaja en la
 * URL y al API) y el calendario se dibuja con los tokens de NAVA.
 *
 * Solo trabaja con fechas civiles "AAAA-MM-DD" (RN-DIS-07): nunca crea un
 * `Date` local ni lee la zona del dispositivo; qué día es "hoy" lo decide
 * quien lo monta, con la zona de la barbería.
 *
 * Patrón ARIA "Date Picker Dialog" (WAI-ARIA APG): disparador con
 * aria-haspopup="dialog", diálogo no modal con foco atrapado y una cuadrícula
 * role="grid" con roving tabindex (flechas ±1/±7 días, Inicio/Fin de semana,
 * RePág/AvPág ±mes, Mayús+RePág/AvPág ±año, Esc cierra).
 *
 * Zoom propio: la rueda sobre el calendario (o el pellizco en trackpad) y el
 * pellizco de dos dedos en pantalla táctil lo agrandan o reducen; `+`, `-` y
 * `0` hacen lo mismo con teclado. El factor solo escala tamaños del calendario
 * (`--dp-zoom`), nunca la página.
 *
 * No decide enrutamiento: emite `update:modelValue` y DailyAgendaPage sigue
 * siendo la única fuente de verdad de `route.query.date`.
 */
import { computed, nextTick, ref, useId, watch } from 'vue'
import { usePickerPopover } from '@/shared/composables/usePickerPopover'
import { formatCivilDateFull, isCivilDateString, shiftCivilDate } from '@/shared/time/civilDate'

interface Props {
  modelValue: string | null
  /** Día civil vigente en la barbería, para marcarlo y ofrecer «Hoy». Sin él
   * (zona desconocida) no se marca ningún día ni se muestra el atajo. */
  today?: string | null
  disabled?: boolean
  /** id del botón disparador (lo nombra el `<label for>` de la pantalla). */
  triggerId?: string
  placeholder?: string
  required?: boolean
  /** id del rótulo visible: el nombre accesible del disparador es
   * «<rótulo> <fecha>» para que un lector de pantalla oiga también el valor. */
  labelId?: string
}

const props = withDefaults(defineProps<Props>(), {
  today: null,
  disabled: false,
  triggerId: undefined,
  placeholder: 'Elegir fecha',
  required: false,
  labelId: undefined,
})

const emit = defineEmits<{
  'update:modelValue': [date: string]
}>()

const WEEKDAYS = [
  { abbr: 'lunes', short: 'L' },
  { abbr: 'martes', short: 'M' },
  { abbr: 'miércoles', short: 'X' },
  { abbr: 'jueves', short: 'J' },
  { abbr: 'viernes', short: 'V' },
  { abbr: 'sábado', short: 'S' },
  { abbr: 'domingo', short: 'D' },
]
const ZOOM_MIN = 0.85
const ZOOM_MAX = 1.6
const ZOOM_KEY_STEP = 1.1
// Rueda de ratón: ~14 % por muesca (deltaY 100). Pellizco de trackpad
// (rueda con ctrlKey): deltas de 1-10, de ahí un coeficiente mayor.
const WHEEL_ZOOM_RATE = 0.0015
const PINCH_WHEEL_ZOOM_RATE = 0.01
const WHEEL_LINE_HEIGHT = 33

const dialogId = useId()
const titleId = useId()
const rootRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const dialogRef = ref<HTMLElement | null>(null)
const open = ref(false)
// Día con el foco de teclado; el mes visible se deriva de él, así una sola
// fuente de verdad mueve cuadrícula, título y roving tabindex.
const focusedDate = ref('')
const generatedTriggerId = useId()
const effectiveTriggerId = computed(() => props.triggerId ?? `date-picker-${generatedTriggerId}`)
const zoom = ref(1)

function pad2(value: number): string {
  return String(value).padStart(2, '0')
}

function parts(date: string): [number, number, number] {
  const [year, month, day] = date.split('-').map(Number)
  return [year!, month!, day!]
}

function addMonths(date: string, delta: number): string {
  const [year, month, day] = parts(date)
  const target = new Date(Date.UTC(year, month - 1 + delta, 1))
  const lastDay = new Date(
    Date.UTC(target.getUTCFullYear(), target.getUTCMonth() + 1, 0),
  ).getUTCDate()
  return `${target.getUTCFullYear()}-${pad2(target.getUTCMonth() + 1)}-${pad2(Math.min(day, lastDay))}`
}

const monthTitle = computed(() => {
  if (!focusedDate.value) return ''
  const [year, month] = parts(focusedDate.value)
  const text = new Intl.DateTimeFormat('es-CO', {
    timeZone: 'UTC',
    month: 'long',
    year: 'numeric',
  }).format(new Date(Date.UTC(year, month - 1, 1, 12)))
  return text.charAt(0).toUpperCase() + text.slice(1)
})

interface DayCell {
  date: string
  day: number
  inMonth: boolean
  isToday: boolean
  isSelected: boolean
}

// Siempre seis semanas: la altura del calendario no salta al cambiar de mes.
const weeks = computed<DayCell[][]>(() => {
  if (!focusedDate.value) return []
  const [year, month] = parts(focusedDate.value)
  const first = `${year}-${pad2(month)}-01`
  const leading = (new Date(Date.UTC(year, month - 1, 1)).getUTCDay() + 6) % 7
  const start = shiftCivilDate(first, -leading)
  return Array.from({ length: 6 }, (_, week) =>
    Array.from({ length: 7 }, (_, weekday) => {
      const date = shiftCivilDate(start, week * 7 + weekday)
      const [cellYear, cellMonth, day] = parts(date)
      return {
        date,
        day,
        inMonth: cellYear === year && cellMonth === month,
        isToday: date === props.today,
        isSelected: date === props.modelValue,
      }
    }),
  )
})

function focusDay() {
  void nextTick(() => {
    const day = dialogRef.value?.querySelector<HTMLElement>(`[data-date="${focusedDate.value}"]`)
    day?.focus({ preventScroll: true })
    day?.scrollIntoView?.({ block: 'nearest' })
  })
}

const { style: popoverStyle, reposition } = usePickerPopover(
  open,
  rootRef,
  triggerRef,
  dialogRef,
  close,
)

function setZoom(next: number) {
  const clamped = Math.min(ZOOM_MAX, Math.max(ZOOM_MIN, next))
  const rounded = Math.round(clamped * 1000) / 1000
  if (rounded === zoom.value) return
  zoom.value = rounded
  void nextTick(reposition)
}

// La rueda sobre el calendario lo acerca (hacia arriba) o lo aleja: se cancela
// para que la página no se desplace mientras el cursor está encima.
function onWheel(event: WheelEvent) {
  const delta = event.deltaMode === 1 ? event.deltaY * WHEEL_LINE_HEIGHT : event.deltaY
  const rate = event.ctrlKey ? PINCH_WHEEL_ZOOM_RATE : WHEEL_ZOOM_RATE
  setZoom(zoom.value * Math.exp(-delta * rate))
}

// Pellizco de dos dedos: el factor es la razón entre la distancia actual y la
// inicial. El CSS (touch-action: none) deja los gestos táctiles al calendario;
// sin él el navegador los toma como desplazamiento y cancela los punteros.
const pointers = new Map<number, { x: number; y: number }>()
let pinchStartDistance = 0
let pinchStartZoom = 1
let swallowClick = false

function pointerDistance(): number {
  const [a, b] = [...pointers.values()]
  return a && b ? Math.hypot(a.x - b.x, a.y - b.y) : 0
}

function onPointerDown(event: PointerEvent) {
  if (event.pointerType === 'mouse') return
  pointers.set(event.pointerId, { x: event.clientX, y: event.clientY })
  if (pointers.size === 2) {
    pinchStartDistance = pointerDistance()
    pinchStartZoom = zoom.value
  }
}

function onPointerMove(event: PointerEvent) {
  const pointer = pointers.get(event.pointerId)
  if (!pointer) return
  pointer.x = event.clientX
  pointer.y = event.clientY
  if (pointers.size === 2 && pinchStartDistance > 0) {
    swallowClick = true
    setZoom(pinchStartZoom * (pointerDistance() / pinchStartDistance))
  }
}

function onPointerEnd(event: PointerEvent) {
  pointers.delete(event.pointerId)
  // Al soltar un pellizco el segundo dedo puede acabar sobre un día y generar
  // un clic que elegiría fecha sin querer: se descarta el del mismo gesto.
  if (swallowClick && pointers.size === 0) {
    setTimeout(() => {
      swallowClick = false
    }, 300)
  }
}

function onDialogClickCapture(event: MouseEvent) {
  if (!swallowClick) return
  event.preventDefault()
  event.stopPropagation()
}

function openCalendar() {
  if (props.disabled) return
  focusedDate.value =
    props.modelValue && isCivilDateString(props.modelValue) ? props.modelValue : (props.today ?? '')
  if (!focusedDate.value) return
  open.value = true
  focusDay()
}

function close(returnFocus = true) {
  open.value = false
  if (returnFocus) triggerRef.value?.focus()
}

function toggle() {
  if (open.value) close()
  else openCalendar()
}

function pick(date: string) {
  emit('update:modelValue', date)
  close()
}

function moveFocus(next: string) {
  focusedDate.value = next
  focusDay()
}

function onGridKeydown(event: KeyboardEvent) {
  const current = focusedDate.value
  let next: string | null = null
  switch (event.key) {
    case 'ArrowLeft':
      next = shiftCivilDate(current, -1)
      break
    case 'ArrowRight':
      next = shiftCivilDate(current, 1)
      break
    case 'ArrowUp':
      next = shiftCivilDate(current, -7)
      break
    case 'ArrowDown':
      next = shiftCivilDate(current, 7)
      break
    case 'Home':
      next = shiftCivilDate(current, -((new Date(`${current}T12:00:00Z`).getUTCDay() + 6) % 7))
      break
    case 'End':
      next = shiftCivilDate(current, 6 - ((new Date(`${current}T12:00:00Z`).getUTCDay() + 6) % 7))
      break
    case 'PageUp':
      next = addMonths(current, event.shiftKey ? -12 : -1)
      break
    case 'PageDown':
      next = addMonths(current, event.shiftKey ? 12 : 1)
      break
  }
  if (next) {
    event.preventDefault()
    moveFocus(next)
  }
}

// Foco atrapado (diálogo): Tab/Mayús+Tab cicla dentro del calendario y Esc lo
// cierra devolviendo el foco al disparador.
function onDialogKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    event.preventDefault()
    event.stopPropagation()
    close()
    return
  }
  if (event.key === '+' || event.key === '=') {
    event.preventDefault()
    setZoom(zoom.value * ZOOM_KEY_STEP)
    return
  }
  if (event.key === '-') {
    event.preventDefault()
    setZoom(zoom.value / ZOOM_KEY_STEP)
    return
  }
  if (event.key === '0') {
    event.preventDefault()
    setZoom(1)
    return
  }
  if (event.key !== 'Tab' || !dialogRef.value) return
  event.stopPropagation()
  const focusables = [
    ...dialogRef.value.querySelectorAll<HTMLElement>('button:not(:disabled):not([tabindex="-1"])'),
  ]
  const first = focusables[0]
  const last = focusables[focusables.length - 1]
  if (!first || !last) return
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault()
    last.focus()
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault()
    first.focus()
  }
}

function onTriggerKeydown(event: KeyboardEvent) {
  if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
    event.preventDefault()
    openCalendar()
  }
}

watch(
  () => props.disabled,
  (disabled) => {
    if (disabled) close(false)
  },
)

const triggerLabelledBy = computed(() =>
  props.labelId ? `${props.labelId} ${effectiveTriggerId.value}` : undefined,
)
</script>

<template>
  <div ref="rootRef" class="agenda-date-picker" :aria-owns="open ? dialogId : undefined">
    <button
      :id="effectiveTriggerId"
      ref="triggerRef"
      type="button"
      class="agenda-date-picker__trigger"
      :disabled="disabled"
      aria-haspopup="dialog"
      :aria-expanded="open"
      :aria-controls="dialogId"
      :aria-labelledby="triggerLabelledBy"
      :aria-describedby="required ? `${dialogId}-required` : undefined"
      @click="toggle"
      @keydown="onTriggerKeydown"
    >
      <span class="agenda-date-picker__value">{{ modelValue || placeholder }}</span>
      <svg
        class="agenda-date-picker__icon"
        viewBox="0 0 16 16"
        width="16"
        height="16"
        fill="none"
        stroke="currentColor"
        stroke-width="1.25"
        aria-hidden="true"
        focusable="false"
      >
        <rect x="2" y="3.25" width="12" height="10.5" />
        <path d="M2 6.75h12M5.25 1.75v3M10.75 1.75v3" />
      </svg>
    </button>

    <span v-if="required" :id="`${dialogId}-required`" class="agenda-date-picker__sr-only"
      >Fecha obligatoria.</span
    >
    <Teleport to="body">
      <Transition name="agenda-date-picker-pop">
        <div
          v-if="open"
          :id="dialogId"
          ref="dialogRef"
          class="agenda-date-picker__dialog"
          role="dialog"
          aria-label="Elegir fecha"
          :style="{ ...popoverStyle, '--dp-zoom': zoom }"
          @keydown="onDialogKeydown"
          @wheel.prevent="onWheel"
          @pointerdown="onPointerDown"
          @pointermove="onPointerMove"
          @pointerup="onPointerEnd"
          @pointercancel="onPointerEnd"
          @click.capture="onDialogClickCapture"
        >
          <div class="agenda-date-picker__header">
            <button
              type="button"
              class="agenda-date-picker__nav"
              aria-label="Mes anterior"
              @click="focusedDate = addMonths(focusedDate, -1)"
            >
              <span class="agenda-date-picker__nav-arrow agenda-date-picker__nav-arrow--prev" />
            </button>
            <div :id="titleId" class="agenda-date-picker__title" aria-live="polite">
              {{ monthTitle }}
            </div>
            <button
              type="button"
              class="agenda-date-picker__nav"
              aria-label="Mes siguiente"
              @click="focusedDate = addMonths(focusedDate, 1)"
            >
              <span class="agenda-date-picker__nav-arrow agenda-date-picker__nav-arrow--next" />
            </button>
          </div>

          <table class="agenda-date-picker__grid" role="grid" :aria-labelledby="titleId">
            <thead>
              <tr>
                <th
                  v-for="weekday in WEEKDAYS"
                  :key="weekday.abbr"
                  scope="col"
                  :abbr="weekday.abbr"
                >
                  <span aria-hidden="true">{{ weekday.short }}</span>
                  <span class="agenda-date-picker__sr-only">{{ weekday.abbr }}</span>
                </th>
              </tr>
            </thead>
            <tbody @keydown="onGridKeydown">
              <tr v-for="week in weeks" :key="week[0]!.date">
                <td
                  v-for="cell in week"
                  :key="cell.date"
                  role="gridcell"
                  :aria-selected="cell.isSelected"
                >
                  <button
                    type="button"
                    class="agenda-date-picker__day"
                    :class="{
                      'agenda-date-picker__day--outside': !cell.inMonth,
                      'agenda-date-picker__day--today': cell.isToday,
                      'agenda-date-picker__day--selected': cell.isSelected,
                    }"
                    :data-date="cell.date"
                    :tabindex="cell.date === focusedDate ? 0 : -1"
                    :aria-label="formatCivilDateFull(cell.date)"
                    :aria-current="cell.isToday ? 'date' : undefined"
                    @click="pick(cell.date)"
                  >
                    {{ cell.day }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>

          <div v-if="today" class="agenda-date-picker__footer">
            <button type="button" class="agenda-date-picker__today" @click="pick(today)">
              Hoy
            </button>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.agenda-date-picker {
  position: relative;
}

/* Campo hundido: más oscuro que la tinta de la página (--color-field-strong)
   con filete de latón, nunca un velo claro sobre el azul. Mismo lenguaje y
   misma altura que BarberSelect. */
.agenda-date-picker__trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
  width: 100%;
  min-height: 40px;
  padding: 0 var(--space-3);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-on-strong);
  text-align: left;
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-bottom: var(--border-width-emphasis) solid var(--color-accent-brass);
  border-radius: 2px;
  cursor: pointer;
}

.agenda-date-picker__trigger:hover:not(:disabled),
.agenda-date-picker__trigger[aria-expanded='true'] {
  border-color: rgb(184 149 90 / 55%);
  border-bottom-color: var(--color-brand-accent-surface);
}

.agenda-date-picker__trigger:disabled {
  cursor: not-allowed;
  color: var(--color-on-strong-muted);
  opacity: 0.42;
}

.agenda-date-picker__trigger:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus);
}

.agenda-date-picker__value {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.agenda-date-picker__icon {
  flex-shrink: 0;
  color: var(--color-brand-accent-surface);
}

.agenda-date-picker__dialog {
  position: fixed;
  z-index: var(--layer-dialog);
  max-height: calc(100dvh - 16px);
  overflow-y: auto;
  overscroll-behavior: contain;
  --color-accent-brass: var(--color-brand-accent-surface);
  box-sizing: border-box;
  width: min(calc(336px * var(--dp-zoom, 1)), calc(100vw - 16px));
  padding: var(--space-2) var(--space-3) var(--space-3);
  /* Los gestos táctiles sobre el calendario son suyos (pellizco = zoom): sin
     esto el navegador los toma como desplazamiento y cancela el puntero. */
  touch-action: none;
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid rgb(244 240 231 / 16%);
  border-top: var(--border-width-emphasis) solid var(--color-accent-brass);
  border-radius: 2px;
  box-shadow: var(--shadow-dialog);
  /* Reserva bajo el calendario la barra de navegación fija del móvil. */
  scroll-margin-bottom: 96px;
}

.agenda-date-picker__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
  margin-bottom: var(--space-1);
}

.agenda-date-picker__title {
  flex: 1;
  font-family: var(--font-display);
  font-size: calc(20px * var(--dp-zoom, 1));
  line-height: 1.2;
  color: var(--color-on-strong);
  text-align: center;
}

.agenda-date-picker__nav {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: calc(44px * var(--dp-zoom, 1));
  height: calc(44px * var(--dp-zoom, 1));
  padding: 0;
  background: transparent;
  border: var(--border-width-normal) solid transparent;
  border-radius: 2px;
  cursor: pointer;
}

.agenda-date-picker__nav:hover {
  background-color: var(--color-field-strong-raised);
  border-color: rgb(184 149 90 / 55%);
}

.agenda-date-picker__nav:focus-visible,
.agenda-date-picker__day:focus-visible,
.agenda-date-picker__today:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-field-strong),
    0 0 0 4px var(--color-brand-accent-surface);
}

.agenda-date-picker__nav-arrow {
  width: 8px;
  height: 8px;
  border-top: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  border-left: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
}

.agenda-date-picker__nav-arrow--prev {
  transform: translateX(2px) rotate(-45deg);
}

.agenda-date-picker__nav-arrow--next {
  transform: translateX(-2px) rotate(135deg);
}

.agenda-date-picker__grid {
  width: 100%;
  table-layout: fixed;
  border-collapse: collapse;
}

.agenda-date-picker__grid th {
  padding: var(--space-1) 0 var(--space-2);
  font-size: calc(13px * var(--dp-zoom, 1));
  font-weight: 500;
  letter-spacing: 0.12em;
  color: var(--color-brand-accent-surface);
  text-align: center;
  border-bottom: var(--border-width-normal) solid rgb(244 240 231 / 10%);
}

.agenda-date-picker__grid td {
  padding: 1px 0;
  text-align: center;
}

.agenda-date-picker__day {
  position: relative;
  width: 100%;
  min-width: 0;
  height: calc(42px * var(--dp-zoom, 1));
  padding: 0;
  font-family: var(--font-family-base);
  font-size: calc(var(--font-size-body) * var(--dp-zoom, 1));
  font-variant-numeric: tabular-nums;
  color: var(--color-on-strong);
  background: transparent;
  border: var(--border-width-normal) solid transparent;
  border-radius: 2px;
  cursor: pointer;
}

.agenda-date-picker__day:hover {
  background-color: var(--color-field-strong-raised);
  border-color: rgb(184 149 90 / 55%);
}

.agenda-date-picker__day--outside {
  color: var(--color-on-strong-muted);
}

/* Hoy: rombo de latón bajo la cifra, el motivo de la marca, sin competir con
   el día elegido. */
.agenda-date-picker__day--today::after {
  position: absolute;
  bottom: 3px;
  left: 50%;
  width: 5px;
  height: 5px;
  content: '';
  background-color: var(--color-brand-accent-surface);
  transform: translateX(-50%) rotate(45deg);
}

.agenda-date-picker__day--selected,
.agenda-date-picker__day--selected:hover {
  font-weight: 600;
  color: var(--color-surface-strong);
  background-color: var(--color-brand-accent-surface);
  border-color: var(--color-brand-accent-surface);
}

.agenda-date-picker__day--selected.agenda-date-picker__day--today::after {
  background-color: var(--color-surface-strong);
}

.agenda-date-picker__footer {
  display: flex;
  justify-content: center;
  margin-top: var(--space-2);
  padding-top: var(--space-2);
  border-top: var(--border-width-normal) solid rgb(244 240 231 / 10%);
}

.agenda-date-picker__today {
  min-height: calc(40px * var(--dp-zoom, 1));
  padding: 0 var(--space-4);
  font-family: var(--font-family-base);
  font-size: calc(var(--font-size-body-sm) * var(--dp-zoom, 1));
  font-weight: 500;
  color: var(--color-brand-accent-surface);
  background: transparent;
  border: var(--border-width-normal) solid rgb(184 149 90 / 50%);
  border-bottom-color: var(--color-brand-accent-surface);
  border-radius: 2px;
  cursor: pointer;
}

.agenda-date-picker__today:hover {
  background-color: rgb(184 149 90 / 12%);
}

.agenda-date-picker__sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}

.agenda-date-picker-pop-enter-active,
.agenda-date-picker-pop-leave-active {
  transition:
    opacity 0.12s ease,
    margin-top 0.12s ease;
}

.agenda-date-picker-pop-enter-from,
.agenda-date-picker-pop-leave-to {
  margin-top: -4px;
  opacity: 0;
}

/* La columna de la fecha mide ~136 px a 360 px y ~118 px a 320 px: con menos
   relleno lateral el ISO (~92 px) cabe con el icono, y a 320 px se prescinde
   del icono; el campo sigue siendo un botón con su nombre accesible. */
@media (max-width: 399px) {
  .agenda-date-picker__trigger {
    gap: var(--space-1);
    padding: 0 var(--space-2);
  }
}

@media (max-width: 359px) {
  .agenda-date-picker__icon {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .agenda-date-picker-pop-enter-active,
  .agenda-date-picker-pop-leave-active {
    transition: none;
  }
}
</style>
