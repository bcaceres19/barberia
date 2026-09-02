<script setup lang="ts">
// Pantalla "Horarios" (HU-040): elegir un barbero y mantener sus tramos de
// horario laboral recurrente, agrupados por día ISO de la semana (los
// siete días siempre aparecen, con o sin tramos, CA-040-01). Un barbero con
// un solo tramo y uno con jornada partida en varios días usan exactamente
// el mismo componente (CA-040-02): nunca hay una rama especial. Estados
// discriminados: carga inicial, listo, vacío (sin barberos), error
// recuperable, cargando tramos del barbero elegido, guardando. Un error
// recuperable nunca borra lo que ya se escribió en el diálogo abierto; solo
// una respuesta exitosa del servidor cierra el diálogo o cambia una fila
// (trabajo requerido §4.3/§4.4 de HU-021, mismo criterio aquí).
import { computed, onMounted, ref } from 'vue'
import { BaseAlert, BaseButton, BaseDialog, BaseInput } from '@/shared/ui'
import {
  createWorkingHour,
  deleteWorkingHour,
  fetchBarberSummaries,
  fetchBarbershopTimezone,
  fetchWorkingHours,
  updateWorkingHour,
} from '../api/schedulesApi'
import {
  createScheduleException,
  deleteScheduleException,
  fetchColombianHolidays,
  fetchHolidayCalendar,
  fetchScheduleExceptions,
  updateHolidayCalendar,
  updateScheduleException,
  type ScheduleExceptionSegmentInput,
} from '../api/scheduleExceptionsApi'
import { newIdempotencyKey } from '../model/idempotencyKey'
import type { ColombianHoliday, ScheduleException } from '../model/scheduleException'
import {
  ISO_WEEKDAYS,
  weekdayLabel,
  type BarberSummary,
  type WorkingHour,
} from '../model/workingHour'
import {
  validateDurationMinutes,
  validateISOWeekday,
  validateStartsTime,
} from '../validation/scheduleValidation'
import {
  validateEffectiveDate,
  validateExceptionShape,
  validateReason,
} from '../validation/exceptionValidation'

type PageStatus = 'loading' | 'ready' | 'load-error'
type WorkingHoursStatus = 'idle' | 'loading' | 'ready' | 'error'
type SaveStatus =
  | 'idle'
  | 'saving'
  | 'validation-error'
  | 'overlap-conflict'
  | 'idempotency-conflict'
  | 'not-found'
  | 'network-error'
  | 'unexpected-error'

const pageStatus = ref<PageStatus>('loading')
const barbers = ref<BarberSummary[]>([])
const selectedBarberId = ref<string | null>(null)
// CA-040-06: zona IANA de la barbería, nunca la del dispositivo. null
// mientras carga o si la consulta falla; no bloquea la pantalla (la hora
// civil ya es correcta sin conversión alguna).
const barbershopTimezone = ref<string | null>(null)

const workingHoursStatus = ref<WorkingHoursStatus>('idle')
const workingHours = ref<WorkingHour[]>([])
const pendingDeleteIds = ref<Set<string>>(new Set())

const selectedBarber = computed(
  () => barbers.value.find((b) => b.id === selectedBarberId.value) ?? null,
)

// groupedByWeekday siempre incluye los siete días, en orden, con o sin
// tramos (CA-040-01): un día sin ninguno simplemente muestra su lista
// vacía, nunca desaparece de la pantalla.
const groupedByWeekday = computed(() => {
  return ISO_WEEKDAYS.map((day) => ({
    ...day,
    items: workingHours.value
      .filter((wh) => wh.isoWeekday === day.value)
      .sort((a, b) => a.startsTime.localeCompare(b.startsTime)),
  }))
})

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
  workingHoursStatus.value = 'loading'

  const outcome = await fetchWorkingHours(barberId)
  // El barbero seleccionado pudo cambiar mientras la solicitud estaba en
  // vuelo (cambio rápido en el selector): descarta una respuesta obsoleta.
  if (selectedBarberId.value !== barberId) return

  if (outcome.kind === 'success') {
    workingHours.value = outcome.page.items
    workingHoursStatus.value = 'ready'
  } else {
    workingHoursStatus.value = 'error'
  }

  void loadHolidayCalendar(barberId)
  void loadExceptions(barberId)
}

function onBarberSelectChange(event: Event) {
  const barberId = (event.target as HTMLSelectElement).value
  void selectBarber(barberId)
}

function onRetryWorkingHours() {
  if (selectedBarberId.value) void selectBarber(selectedBarberId.value)
}

// --- Alta -------------------------------------------------------------

const isCreateOpen = ref(false)
const createISOWeekday = ref(1)
const createStartsTime = ref('')
const createDurationMinutes = ref(60)
const createWeekdayError = ref<string | undefined>(undefined)
const createStartsTimeError = ref<string | undefined>(undefined)
const createDurationError = ref<string | undefined>(undefined)
const createStatus = ref<SaveStatus>('idle')
const createAttempted = ref(false)
// Clave de idempotencia del intento lógico vigente (RN-IDE-01): se genera
// al abrir el diálogo y se reutiliza en cada reintento del mismo intento
// (mismo criterio que staff/pages/StaffPage.vue).
let createIdempotencyKey = newIdempotencyKey()

function openCreateDialog() {
  createISOWeekday.value = 1
  createStartsTime.value = ''
  createDurationMinutes.value = 60
  createWeekdayError.value = undefined
  createStartsTimeError.value = undefined
  createDurationError.value = undefined
  createStatus.value = 'idle'
  createAttempted.value = false
  createIdempotencyKey = newIdempotencyKey()
  isCreateOpen.value = true
}

function onCreateDialogClosed() {
  createStatus.value = 'idle'
}

function revalidateCreate() {
  if (!createAttempted.value) return
  createWeekdayError.value = validateISOWeekday(createISOWeekday.value)
  createStartsTimeError.value = validateStartsTime(createStartsTime.value)
  createDurationError.value = validateDurationMinutes(createDurationMinutes.value)
}

function onCreateStartsTimeInput(value: string | number) {
  createStartsTime.value = String(value)
  revalidateCreate()
}

function onCreateDurationInput(value: string | number) {
  createDurationMinutes.value = Number(value)
  revalidateCreate()
}

async function onSubmitCreate() {
  if (createStatus.value === 'saving' || !selectedBarberId.value) return

  createAttempted.value = true
  const weekdayError = validateISOWeekday(createISOWeekday.value)
  const startsTimeError = validateStartsTime(createStartsTime.value)
  const durationError = validateDurationMinutes(createDurationMinutes.value)
  createWeekdayError.value = weekdayError
  createStartsTimeError.value = startsTimeError
  createDurationError.value = durationError
  if (weekdayError || startsTimeError || durationError) return

  createStatus.value = 'saving'
  const outcome = await createWorkingHour(
    selectedBarberId.value,
    createISOWeekday.value,
    createStartsTime.value,
    createDurationMinutes.value,
    createIdempotencyKey,
  )

  switch (outcome.kind) {
    case 'success':
      workingHours.value.push(outcome.workingHour)
      isCreateOpen.value = false
      createStatus.value = 'idle'
      return
    case 'validation-error':
      createStatus.value = 'validation-error'
      return
    case 'overlap-conflict':
      createStatus.value = 'overlap-conflict'
      return
    case 'idempotency-conflict':
      createStatus.value = 'idempotency-conflict'
      return
    case 'not-found':
      createStatus.value = 'not-found'
      return
    case 'network-error':
      createStatus.value = 'network-error'
      return
    case 'unexpected-error':
      createStatus.value = 'unexpected-error'
  }
}

// --- Edición ------------------------------------------------------------

const isEditOpen = ref(false)
const editTarget = ref<WorkingHour | null>(null)
const editISOWeekday = ref(1)
const editStartsTime = ref('')
const editDurationMinutes = ref(60)
const editWeekdayError = ref<string | undefined>(undefined)
const editStartsTimeError = ref<string | undefined>(undefined)
const editDurationError = ref<string | undefined>(undefined)
const editStatus = ref<SaveStatus>('idle')
const editAttempted = ref(false)

function openEditDialog(wh: WorkingHour) {
  editTarget.value = wh
  editISOWeekday.value = wh.isoWeekday
  editStartsTime.value = wh.startsTime
  editDurationMinutes.value = wh.durationMinutes
  editWeekdayError.value = undefined
  editStartsTimeError.value = undefined
  editDurationError.value = undefined
  editStatus.value = 'idle'
  editAttempted.value = false
  isEditOpen.value = true
}

function onEditDialogClosed() {
  editStatus.value = 'idle'
}

function revalidateEdit() {
  if (!editAttempted.value) return
  editWeekdayError.value = validateISOWeekday(editISOWeekday.value)
  editStartsTimeError.value = validateStartsTime(editStartsTime.value)
  editDurationError.value = validateDurationMinutes(editDurationMinutes.value)
}

function onEditStartsTimeInput(value: string | number) {
  editStartsTime.value = String(value)
  revalidateEdit()
}

function onEditDurationInput(value: string | number) {
  editDurationMinutes.value = Number(value)
  revalidateEdit()
}

async function onSubmitEdit() {
  if (editStatus.value === 'saving' || !editTarget.value || !selectedBarberId.value) return

  editAttempted.value = true
  const weekdayError = validateISOWeekday(editISOWeekday.value)
  const startsTimeError = validateStartsTime(editStartsTime.value)
  const durationError = validateDurationMinutes(editDurationMinutes.value)
  editWeekdayError.value = weekdayError
  editStartsTimeError.value = startsTimeError
  editDurationError.value = durationError
  if (weekdayError || startsTimeError || durationError) return

  editStatus.value = 'saving'
  const outcome = await updateWorkingHour(
    selectedBarberId.value,
    editTarget.value.id,
    editISOWeekday.value,
    editStartsTime.value,
    editDurationMinutes.value,
  )

  switch (outcome.kind) {
    case 'success': {
      const index = workingHours.value.findIndex((wh) => wh.id === outcome.workingHour.id)
      if (index !== -1) workingHours.value[index] = outcome.workingHour
      isEditOpen.value = false
      editStatus.value = 'idle'
      return
    }
    case 'validation-error':
      editStatus.value = 'validation-error'
      return
    case 'overlap-conflict':
      editStatus.value = 'overlap-conflict'
      return
    case 'not-found':
      editStatus.value = 'not-found'
      return
    case 'network-error':
      editStatus.value = 'network-error'
      return
    case 'unexpected-error':
      editStatus.value = 'unexpected-error'
  }
}

// --- Retiro ---------------------------------------------------------------

const deleteError = ref<string | null>(null)

function isDeletePending(id: string): boolean {
  return pendingDeleteIds.value.has(id)
}

async function onDelete(wh: WorkingHour) {
  const barberId = selectedBarberId.value
  if (!barberId || isDeletePending(wh.id)) return

  const next = new Set(pendingDeleteIds.value)
  next.add(wh.id)
  pendingDeleteIds.value = next
  deleteError.value = null

  const outcome = await deleteWorkingHour(barberId, wh.id)

  const after = new Set(pendingDeleteIds.value)
  after.delete(wh.id)
  pendingDeleteIds.value = after

  if (outcome.kind === 'success') {
    workingHours.value = workingHours.value.filter((item) => item.id !== wh.id)
    return
  }

  switch (outcome.kind) {
    case 'not-found':
      // Ya no existe (retirado antes, quizá en otra pestaña): se refleja
      // igual, sin tratarlo como un error nuevo que exija reintento.
      workingHours.value = workingHours.value.filter((item) => item.id !== wh.id)
      break
    case 'network-error':
      deleteError.value = 'No pudimos conectar. Revisa tu conexión e inténtalo de nuevo.'
      break
    case 'unexpected-error':
      deleteError.value = 'Ocurrió un error inesperado. Inténtalo de nuevo en unos segundos.'
  }
}

// --- HU-041: calendario de festivos colombianos (CA-041-01/02) -----------

type HolidayCalendarStatus = 'idle' | 'loading' | 'ready' | 'error'
const holidayCalendarStatus = ref<HolidayCalendarStatus>('idle')
const holidayCalendarEnabled = ref(false)
const holidayCalendarSaving = ref(false)
const holidayCalendarError = ref<string | null>(null)

async function loadHolidayCalendar(barberId: string) {
  holidayCalendarStatus.value = 'loading'
  const outcome = await fetchHolidayCalendar(barberId)
  if (selectedBarberId.value !== barberId) return

  if (outcome.kind === 'success') {
    holidayCalendarEnabled.value = outcome.enabled
    holidayCalendarStatus.value = 'ready'
    return
  }
  holidayCalendarStatus.value = 'error'
}

async function onToggleHolidayCalendar(event: Event) {
  const barberId = selectedBarberId.value
  const checked = (event.target as HTMLInputElement).checked
  if (!barberId || holidayCalendarSaving.value) return

  holidayCalendarSaving.value = true
  holidayCalendarError.value = null
  const outcome = await updateHolidayCalendar(barberId, checked)
  holidayCalendarSaving.value = false

  if (outcome.kind === 'success') {
    holidayCalendarEnabled.value = outcome.enabled
    return
  }
  // La casilla vuelve a su valor real: ninguna respuesta distinta de
  // `success` cambió el estado persistido.
  holidayCalendarError.value =
    outcome.kind === 'network-error'
      ? 'No pudimos conectar. Revisa tu conexión e inténtalo de nuevo.'
      : 'Ocurrió un error inesperado. Inténtalo de nuevo en unos segundos.'
}

// --- HU-041: festivos colombianos de referencia (RN-BLQ-02) ---------------

// upcomingColombianHolidays es un dato de referencia único para toda la
// pantalla (no depende del barbero elegido): se carga una sola vez.
const colombianHolidays = ref<ColombianHoliday[]>([])

async function loadColombianHolidays() {
  const year = new Date().getFullYear()
  const outcome = await fetchColombianHolidays(year)
  if (outcome.kind === 'success') {
    const today = new Date().toISOString().slice(0, 10)
    colombianHolidays.value = outcome.items.filter((h) => h.date >= today)
  }
}

// --- HU-041: excepciones de jornada (CA-041-04/05) -------------------------

type ExceptionsStatus = 'idle' | 'loading' | 'ready' | 'error'
const exceptionsStatus = ref<ExceptionsStatus>('idle')
const exceptions = ref<ScheduleException[]>([])
const pendingDeleteExceptionIds = ref<Set<string>>(new Set())
const exceptionDeleteError = ref<string | null>(null)

const sortedExceptions = computed(() =>
  [...exceptions.value].sort((a, b) => a.effectiveDate.localeCompare(b.effectiveDate)),
)

async function loadExceptions(barberId: string) {
  exceptionsStatus.value = 'loading'
  const outcome = await fetchScheduleExceptions(barberId)
  if (selectedBarberId.value !== barberId) return

  if (outcome.kind === 'success') {
    exceptions.value = outcome.page.items
    exceptionsStatus.value = 'ready'
    return
  }
  exceptionsStatus.value = 'error'
}

function onRetryExceptions() {
  if (selectedBarberId.value) void loadExceptions(selectedBarberId.value)
}

type ExceptionSaveStatus =
  | 'idle'
  | 'saving'
  | 'validation-error'
  | 'date-conflict'
  | 'idempotency-conflict'
  | 'not-found'
  | 'network-error'
  | 'unexpected-error'

interface ExceptionSegmentRow {
  startsTime: string
  durationMinutes: number
}

function newSegmentRow(): ExceptionSegmentRow {
  return { startsTime: '', durationMinutes: 60 }
}

// --- Alta de excepción -----------------------------------------------------

const isExceptionCreateOpen = ref(false)
const createEffectiveDate = ref('')
const createIsClosed = ref(true)
const createReason = ref('')
const createSegments = ref<ExceptionSegmentRow[]>([])
const createEffectiveDateError = ref<string | undefined>(undefined)
const createReasonError = ref<string | undefined>(undefined)
const createShapeError = ref<string | undefined>(undefined)
const createExceptionStatus = ref<ExceptionSaveStatus>('idle')
const createExceptionAttempted = ref(false)
let createExceptionIdempotencyKey = newIdempotencyKey()

function openExceptionCreateDialog(prefillDate?: string) {
  createEffectiveDate.value = prefillDate ?? ''
  createIsClosed.value = true
  createReason.value = ''
  createSegments.value = []
  createEffectiveDateError.value = undefined
  createReasonError.value = undefined
  createShapeError.value = undefined
  createExceptionStatus.value = 'idle'
  createExceptionAttempted.value = false
  createExceptionIdempotencyKey = newIdempotencyKey()
  isExceptionCreateOpen.value = true
}

function onExceptionCreateDialogClosed() {
  createExceptionStatus.value = 'idle'
}

function onCreateIsClosedChange(isClosed: boolean) {
  createIsClosed.value = isClosed
  if (!isClosed && createSegments.value.length === 0) {
    createSegments.value = [newSegmentRow()]
  }
  revalidateExceptionCreate()
}

function addCreateSegment() {
  createSegments.value = [...createSegments.value, newSegmentRow()]
}

function removeCreateSegment(index: number) {
  createSegments.value = createSegments.value.filter((_, i) => i !== index)
  revalidateExceptionCreate()
}

function revalidateExceptionCreate() {
  if (!createExceptionAttempted.value) return
  createEffectiveDateError.value = validateEffectiveDate(createEffectiveDate.value)
  createReasonError.value = validateReason(createReason.value)
  createShapeError.value = validateExceptionShape(createIsClosed.value, createSegments.value)
}

async function onSubmitExceptionCreate() {
  if (createExceptionStatus.value === 'saving' || !selectedBarberId.value) return

  createExceptionAttempted.value = true
  const effectiveDateError = validateEffectiveDate(createEffectiveDate.value)
  const reasonError = validateReason(createReason.value)
  const shapeError = validateExceptionShape(createIsClosed.value, createSegments.value)
  createEffectiveDateError.value = effectiveDateError
  createReasonError.value = reasonError
  createShapeError.value = shapeError
  if (effectiveDateError || reasonError || shapeError) return

  createExceptionStatus.value = 'saving'
  const segments: ScheduleExceptionSegmentInput[] = createIsClosed.value
    ? []
    : createSegments.value.map((s) => ({
        startsTime: s.startsTime,
        durationMinutes: s.durationMinutes,
      }))
  const outcome = await createScheduleException(
    selectedBarberId.value,
    createEffectiveDate.value,
    createIsClosed.value,
    createReason.value.trim() === '' ? null : createReason.value.trim(),
    segments,
    createExceptionIdempotencyKey,
  )

  switch (outcome.kind) {
    case 'success':
      exceptions.value.push(outcome.exception)
      isExceptionCreateOpen.value = false
      createExceptionStatus.value = 'idle'
      return
    case 'validation-error':
      createExceptionStatus.value = 'validation-error'
      return
    case 'date-conflict':
      createExceptionStatus.value = 'date-conflict'
      return
    case 'idempotency-conflict':
      createExceptionStatus.value = 'idempotency-conflict'
      return
    case 'not-found':
      createExceptionStatus.value = 'not-found'
      return
    case 'network-error':
      createExceptionStatus.value = 'network-error'
      return
    case 'unexpected-error':
      createExceptionStatus.value = 'unexpected-error'
  }
}

// --- Edición de excepción ---------------------------------------------------

const isExceptionEditOpen = ref(false)
const editExceptionTarget = ref<ScheduleException | null>(null)
const editEffectiveDate = ref('')
const editIsClosed = ref(true)
const editReason = ref('')
const editSegments = ref<ExceptionSegmentRow[]>([])
const editEffectiveDateError = ref<string | undefined>(undefined)
const editReasonError = ref<string | undefined>(undefined)
const editShapeError = ref<string | undefined>(undefined)
const editExceptionStatus = ref<ExceptionSaveStatus>('idle')
const editExceptionAttempted = ref(false)

function openExceptionEditDialog(exception: ScheduleException) {
  editExceptionTarget.value = exception
  editEffectiveDate.value = exception.effectiveDate
  editIsClosed.value = exception.isClosed
  editReason.value = exception.reason ?? ''
  editSegments.value = exception.segments.map((s) => ({
    startsTime: s.startsTime,
    durationMinutes: s.durationMinutes,
  }))
  editEffectiveDateError.value = undefined
  editReasonError.value = undefined
  editShapeError.value = undefined
  editExceptionStatus.value = 'idle'
  editExceptionAttempted.value = false
  isExceptionEditOpen.value = true
}

function onExceptionEditDialogClosed() {
  editExceptionStatus.value = 'idle'
}

function onEditIsClosedChange(isClosed: boolean) {
  editIsClosed.value = isClosed
  if (!isClosed && editSegments.value.length === 0) {
    editSegments.value = [newSegmentRow()]
  }
  revalidateExceptionEdit()
}

function addEditSegment() {
  editSegments.value = [...editSegments.value, newSegmentRow()]
}

function removeEditSegment(index: number) {
  editSegments.value = editSegments.value.filter((_, i) => i !== index)
  revalidateExceptionEdit()
}

function revalidateExceptionEdit() {
  if (!editExceptionAttempted.value) return
  editEffectiveDateError.value = validateEffectiveDate(editEffectiveDate.value)
  editReasonError.value = validateReason(editReason.value)
  editShapeError.value = validateExceptionShape(editIsClosed.value, editSegments.value)
}

async function onSubmitExceptionEdit() {
  if (
    editExceptionStatus.value === 'saving' ||
    !editExceptionTarget.value ||
    !selectedBarberId.value
  )
    return

  editExceptionAttempted.value = true
  const effectiveDateError = validateEffectiveDate(editEffectiveDate.value)
  const reasonError = validateReason(editReason.value)
  const shapeError = validateExceptionShape(editIsClosed.value, editSegments.value)
  editEffectiveDateError.value = effectiveDateError
  editReasonError.value = reasonError
  editShapeError.value = shapeError
  if (effectiveDateError || reasonError || shapeError) return

  editExceptionStatus.value = 'saving'
  const segments: ScheduleExceptionSegmentInput[] = editIsClosed.value
    ? []
    : editSegments.value.map((s) => ({
        startsTime: s.startsTime,
        durationMinutes: s.durationMinutes,
      }))
  const outcome = await updateScheduleException(
    selectedBarberId.value,
    editExceptionTarget.value.id,
    editEffectiveDate.value,
    editIsClosed.value,
    editReason.value.trim() === '' ? null : editReason.value.trim(),
    segments,
  )

  switch (outcome.kind) {
    case 'success': {
      const index = exceptions.value.findIndex((e) => e.id === outcome.exception.id)
      if (index !== -1) exceptions.value[index] = outcome.exception
      isExceptionEditOpen.value = false
      editExceptionStatus.value = 'idle'
      return
    }
    case 'validation-error':
      editExceptionStatus.value = 'validation-error'
      return
    case 'date-conflict':
      editExceptionStatus.value = 'date-conflict'
      return
    case 'not-found':
      editExceptionStatus.value = 'not-found'
      return
    case 'network-error':
      editExceptionStatus.value = 'network-error'
      return
    case 'unexpected-error':
      editExceptionStatus.value = 'unexpected-error'
  }
}

// --- Retiro de excepción ----------------------------------------------------

function isExceptionDeletePending(id: string): boolean {
  return pendingDeleteExceptionIds.value.has(id)
}

async function onDeleteException(exception: ScheduleException) {
  const barberId = selectedBarberId.value
  if (!barberId || isExceptionDeletePending(exception.id)) return

  const next = new Set(pendingDeleteExceptionIds.value)
  next.add(exception.id)
  pendingDeleteExceptionIds.value = next
  exceptionDeleteError.value = null

  const outcome = await deleteScheduleException(barberId, exception.id)

  const after = new Set(pendingDeleteExceptionIds.value)
  after.delete(exception.id)
  pendingDeleteExceptionIds.value = after

  if (outcome.kind === 'success') {
    exceptions.value = exceptions.value.filter((item) => item.id !== exception.id)
    return
  }

  switch (outcome.kind) {
    case 'not-found':
      exceptions.value = exceptions.value.filter((item) => item.id !== exception.id)
      break
    case 'network-error':
      exceptionDeleteError.value = 'No pudimos conectar. Revisa tu conexión e inténtalo de nuevo.'
      break
    case 'unexpected-error':
      exceptionDeleteError.value =
        'Ocurrió un error inesperado. Inténtalo de nuevo en unos segundos.'
  }
}

onMounted(loadColombianHolidays)
</script>

<template>
  <section class="schedules-page" aria-labelledby="schedules-page-title">
    <header class="schedules-page__header">
      <h1 id="schedules-page-title" class="schedules-page__title">Horarios</h1>
      <BaseButton
        v-if="pageStatus === 'ready' && barbers.length > 0"
        type="button"
        variant="primary"
        @click="openCreateDialog"
      >
        Agregar tramo
      </BaseButton>
    </header>

    <p v-if="barbershopTimezone" class="schedules-page__timezone">
      Horas en la zona horaria de la barbería: {{ barbershopTimezone }}
    </p>

    <div
      v-if="pageStatus === 'loading'"
      class="schedules-page__state"
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
      <p v-if="barbers.length === 0" class="schedules-page__empty">
        Aún no tienes barberos registrados. Agrega uno en la sección "Barberos" antes de configurar
        su horario.
      </p>

      <template v-else>
        <div class="schedules-page__picker">
          <label for="schedules-barber-select" class="schedules-page__label">Barbero</label>
          <select
            id="schedules-barber-select"
            class="schedules-page__select"
            :value="selectedBarberId ?? ''"
            @change="onBarberSelectChange"
          >
            <option v-for="barber in barbers" :key="barber.id" :value="barber.id">
              {{ barber.fullName }}
            </option>
          </select>
        </div>

        <div
          v-if="workingHoursStatus === 'loading'"
          class="schedules-page__state"
          role="status"
          aria-live="polite"
        >
          <p>Cargando el horario de {{ selectedBarber?.fullName }}…</p>
        </div>

        <BaseAlert
          v-else-if="workingHoursStatus === 'error'"
          variant="warning"
          title="No pudimos cargar el horario de este barbero"
          role="alert"
        >
          Revisa tu conexión e inténtalo de nuevo.
          <template #action>
            <BaseButton type="button" variant="secondary" @click="onRetryWorkingHours">
              Reintentar
            </BaseButton>
          </template>
        </BaseAlert>

        <template v-else-if="workingHoursStatus === 'ready'">
          <BaseAlert
            v-if="deleteError"
            variant="danger"
            role="alert"
            class="schedules-page__delete-error"
          >
            {{ deleteError }}
          </BaseAlert>

          <div class="schedules-page__days">
            <section
              v-for="day in groupedByWeekday"
              :key="day.value"
              class="schedules-page__day"
              :aria-labelledby="`schedules-day-${day.value}-title`"
            >
              <h2 :id="`schedules-day-${day.value}-title`" class="schedules-page__day-title">
                {{ day.label }}
              </h2>

              <p v-if="day.items.length === 0" class="schedules-page__day-empty">Sin tramos.</p>

              <ul v-else class="schedules-page__list" :aria-label="`Tramos del ${day.label}`">
                <li v-for="wh in day.items" :key="wh.id" class="schedules-page__item">
                  <span class="schedules-page__item-time">
                    {{ wh.startsTime }} · {{ wh.durationMinutes }} min
                  </span>
                  <div class="schedules-page__item-actions">
                    <BaseButton
                      type="button"
                      variant="secondary"
                      :aria-label="`Editar tramo de ${weekdayLabel(wh.isoWeekday)} a las ${wh.startsTime}`"
                      @click="openEditDialog(wh)"
                    >
                      Editar
                    </BaseButton>
                    <BaseButton
                      type="button"
                      variant="secondary"
                      :loading="isDeletePending(wh.id)"
                      :disabled="isDeletePending(wh.id)"
                      :aria-label="`Retirar tramo de ${weekdayLabel(wh.isoWeekday)} a las ${wh.startsTime}`"
                      @click="onDelete(wh)"
                    >
                      Retirar
                    </BaseButton>
                  </div>
                </li>
              </ul>
            </section>
          </div>
        </template>

        <!-- HU-041: calendario de festivos colombianos -->
        <section class="schedules-page__holiday-calendar" aria-labelledby="holiday-calendar-title">
          <h2 id="holiday-calendar-title" class="schedules-page__day-title">
            Calendario de festivos colombianos
          </h2>
          <BaseAlert v-if="holidayCalendarError" variant="warning" role="alert">
            {{ holidayCalendarError }}
          </BaseAlert>
          <label class="schedules-page__checkbox-label">
            <input
              type="checkbox"
              :checked="holidayCalendarEnabled"
              :disabled="holidayCalendarStatus !== 'ready' || holidayCalendarSaving"
              @change="onToggleHolidayCalendar"
            />
            Cerrar automáticamente los festivos colombianos de este barbero
          </label>
          <p class="schedules-page__day-empty">
            Desactivado: los festivos no agregan ningún bloqueo automático. Activado: un festivo
            queda cerrado por defecto, salvo que exista una excepción manual para esa fecha.
          </p>
        </section>

        <!-- HU-041: excepciones de jornada -->
        <section class="schedules-page__exceptions" aria-labelledby="exceptions-title">
          <header class="schedules-page__exceptions-header">
            <h2 id="exceptions-title" class="schedules-page__day-title">Excepciones de jornada</h2>
            <BaseButton type="button" variant="secondary" @click="openExceptionCreateDialog()">
              Agregar excepción
            </BaseButton>
          </header>

          <div
            v-if="exceptionsStatus === 'loading'"
            class="schedules-page__state"
            role="status"
            aria-live="polite"
          >
            <p>Cargando excepciones…</p>
          </div>

          <BaseAlert v-else-if="exceptionsStatus === 'error'" variant="warning" role="alert">
            No pudimos cargar las excepciones de este barbero. Revisa tu conexión e inténtalo de
            nuevo.
            <template #action>
              <BaseButton type="button" variant="secondary" @click="onRetryExceptions">
                Reintentar
              </BaseButton>
            </template>
          </BaseAlert>

          <template v-else-if="exceptionsStatus === 'ready'">
            <BaseAlert
              v-if="exceptionDeleteError"
              variant="danger"
              role="alert"
              class="schedules-page__delete-error"
            >
              {{ exceptionDeleteError }}
            </BaseAlert>

            <p v-if="sortedExceptions.length === 0" class="schedules-page__day-empty">
              Sin excepciones registradas.
            </p>

            <ul v-else class="schedules-page__list" aria-label="Excepciones de jornada">
              <li
                v-for="exception in sortedExceptions"
                :key="exception.id"
                class="schedules-page__item"
              >
                <div>
                  <span class="schedules-page__item-time">{{ exception.effectiveDate }}</span>
                  <span v-if="exception.isClosed"> · Cerrado</span>
                  <span v-else>
                    ·
                    {{
                      exception.segments
                        .map((s) => `${s.startsTime} (${s.durationMinutes} min)`)
                        .join(', ')
                    }}
                  </span>
                  <span v-if="exception.reason"> · {{ exception.reason }}</span>
                </div>
                <div class="schedules-page__item-actions">
                  <BaseButton
                    type="button"
                    variant="secondary"
                    :aria-label="`Editar excepción del ${exception.effectiveDate}`"
                    @click="openExceptionEditDialog(exception)"
                  >
                    Editar
                  </BaseButton>
                  <BaseButton
                    type="button"
                    variant="secondary"
                    :loading="isExceptionDeletePending(exception.id)"
                    :disabled="isExceptionDeletePending(exception.id)"
                    :aria-label="`Retirar excepción del ${exception.effectiveDate}`"
                    @click="onDeleteException(exception)"
                  >
                    Retirar
                  </BaseButton>
                </div>
              </li>
            </ul>
          </template>
        </section>

        <!-- HU-041: festivos colombianos de referencia -->
        <section
          v-if="colombianHolidays.length > 0"
          class="schedules-page__holidays-reference"
          aria-labelledby="holidays-reference-title"
        >
          <h2 id="holidays-reference-title" class="schedules-page__day-title">
            Próximos festivos colombianos
          </h2>
          <ul class="schedules-page__list" aria-label="Próximos festivos colombianos">
            <li
              v-for="holiday in colombianHolidays"
              :key="holiday.date"
              class="schedules-page__item"
            >
              <span class="schedules-page__item-time">{{ holiday.date }} · {{ holiday.name }}</span>
              <BaseButton
                type="button"
                variant="secondary"
                :aria-label="`Registrar una excepción para el ${holiday.date}, ${holiday.name}`"
                @click="openExceptionCreateDialog(holiday.date)"
              >
                Registrar excepción
              </BaseButton>
            </li>
          </ul>
        </section>
      </template>
    </template>

    <!-- Alta -->
    <BaseDialog
      v-model="isCreateOpen"
      title="Agregar tramo"
      size="sm"
      @close="onCreateDialogClosed"
    >
      <form
        name="createWorkingHour"
        class="schedules-page__form"
        novalidate
        @submit.prevent="onSubmitCreate"
      >
        <BaseAlert
          v-if="createStatus === 'overlap-conflict'"
          variant="danger"
          title="Este tramo se solapa con otro existente"
          role="alert"
        >
          Elige un horario distinto para ese día.
        </BaseAlert>
        <BaseAlert
          v-if="createStatus === 'not-found'"
          variant="warning"
          title="Este barbero ya no está disponible"
          role="alert"
        >
          Cierra este diálogo y recarga la lista.
        </BaseAlert>
        <BaseAlert
          v-if="createStatus === 'idempotency-conflict'"
          variant="danger"
          title="No pudimos completar el intento anterior"
          role="alert"
        >
          Inténtalo de nuevo.
        </BaseAlert>
        <BaseAlert
          v-if="createStatus === 'network-error'"
          variant="warning"
          title="No pudimos conectar"
          role="alert"
        >
          Revisa tu conexión e inténtalo de nuevo. No perdiste lo que escribiste.
        </BaseAlert>
        <BaseAlert
          v-if="createStatus === 'unexpected-error'"
          variant="danger"
          title="Ocurrió un error inesperado"
          role="alert"
        >
          Inténtalo de nuevo en unos segundos. No perdiste lo que escribiste.
        </BaseAlert>

        <div class="schedules-page__field">
          <label for="schedules-create-weekday" class="schedules-page__label">
            Día
            <span class="schedules-page__required" aria-hidden="true">*</span>
          </label>
          <select
            id="schedules-create-weekday"
            v-model.number="createISOWeekday"
            class="schedules-page__select"
            :disabled="createStatus === 'saving'"
            @change="revalidateCreate"
          >
            <option v-for="day in ISO_WEEKDAYS" :key="day.value" :value="day.value">
              {{ day.label }}
            </option>
          </select>
          <div v-if="createWeekdayError" class="schedules-page__field-error" role="alert">
            {{ createWeekdayError }}
          </div>
        </div>

        <BaseInput
          :model-value="createStartsTime"
          type="time"
          name="startsTime"
          label="Hora de inicio"
          required
          :disabled="createStatus === 'saving'"
          :error="createStartsTimeError"
          @update:model-value="onCreateStartsTimeInput"
        />

        <BaseInput
          :model-value="createDurationMinutes"
          type="number"
          name="durationMinutes"
          label="Duración (minutos)"
          required
          :min="1"
          :max="1440"
          :disabled="createStatus === 'saving'"
          :error="createDurationError"
          @update:model-value="onCreateDurationInput"
        />

        <div class="schedules-page__dialog-actions">
          <BaseButton type="button" variant="secondary" @click="isCreateOpen = false">
            Cancelar
          </BaseButton>
          <BaseButton
            type="submit"
            variant="primary"
            :loading="createStatus === 'saving'"
            :disabled="createStatus === 'saving'"
          >
            Guardar
          </BaseButton>
        </div>
      </form>
    </BaseDialog>

    <!-- Edición -->
    <BaseDialog v-model="isEditOpen" title="Editar tramo" size="sm" @close="onEditDialogClosed">
      <form
        name="updateWorkingHour"
        class="schedules-page__form"
        novalidate
        @submit.prevent="onSubmitEdit"
      >
        <BaseAlert
          v-if="editStatus === 'overlap-conflict'"
          variant="danger"
          title="Este tramo se solapa con otro existente"
          role="alert"
        >
          Elige un horario distinto para ese día.
        </BaseAlert>
        <BaseAlert
          v-if="editStatus === 'not-found'"
          variant="warning"
          title="Este tramo ya no está disponible"
          role="alert"
        >
          Cierra este diálogo y recarga la lista.
        </BaseAlert>
        <BaseAlert
          v-if="editStatus === 'network-error'"
          variant="warning"
          title="No pudimos conectar"
          role="alert"
        >
          Revisa tu conexión e inténtalo de nuevo. No perdiste lo que escribiste.
        </BaseAlert>
        <BaseAlert
          v-if="editStatus === 'unexpected-error'"
          variant="danger"
          title="Ocurrió un error inesperado"
          role="alert"
        >
          Inténtalo de nuevo en unos segundos. No perdiste lo que escribiste.
        </BaseAlert>

        <div class="schedules-page__field">
          <label for="schedules-edit-weekday" class="schedules-page__label">
            Día
            <span class="schedules-page__required" aria-hidden="true">*</span>
          </label>
          <select
            id="schedules-edit-weekday"
            v-model.number="editISOWeekday"
            class="schedules-page__select"
            :disabled="editStatus === 'saving'"
            @change="revalidateEdit"
          >
            <option v-for="day in ISO_WEEKDAYS" :key="day.value" :value="day.value">
              {{ day.label }}
            </option>
          </select>
          <div v-if="editWeekdayError" class="schedules-page__field-error" role="alert">
            {{ editWeekdayError }}
          </div>
        </div>

        <BaseInput
          :model-value="editStartsTime"
          type="time"
          name="startsTime"
          label="Hora de inicio"
          required
          :disabled="editStatus === 'saving'"
          :error="editStartsTimeError"
          @update:model-value="onEditStartsTimeInput"
        />

        <BaseInput
          :model-value="editDurationMinutes"
          type="number"
          name="durationMinutes"
          label="Duración (minutos)"
          required
          :min="1"
          :max="1440"
          :disabled="editStatus === 'saving'"
          :error="editDurationError"
          @update:model-value="onEditDurationInput"
        />

        <div class="schedules-page__dialog-actions">
          <BaseButton type="button" variant="secondary" @click="isEditOpen = false">
            Cancelar
          </BaseButton>
          <BaseButton
            type="submit"
            variant="primary"
            :loading="editStatus === 'saving'"
            :disabled="editStatus === 'saving'"
          >
            Guardar
          </BaseButton>
        </div>
      </form>
    </BaseDialog>

    <!-- HU-041: alta de excepción -->
    <BaseDialog
      v-model="isExceptionCreateOpen"
      title="Agregar excepción"
      size="sm"
      @close="onExceptionCreateDialogClosed"
    >
      <form
        name="createScheduleException"
        class="schedules-page__form"
        novalidate
        @submit.prevent="onSubmitExceptionCreate"
      >
        <BaseAlert
          v-if="createExceptionStatus === 'date-conflict'"
          variant="danger"
          title="Ya existe una excepción para esa fecha"
          role="alert"
        >
          Edita la excepción existente en vez de crear una nueva.
        </BaseAlert>
        <BaseAlert
          v-if="createExceptionStatus === 'not-found'"
          variant="warning"
          title="Este barbero ya no está disponible"
          role="alert"
        >
          Cierra este diálogo y recarga la lista.
        </BaseAlert>
        <BaseAlert
          v-if="createExceptionStatus === 'idempotency-conflict'"
          variant="danger"
          title="No pudimos completar el intento anterior"
          role="alert"
        >
          Inténtalo de nuevo.
        </BaseAlert>
        <BaseAlert
          v-if="createExceptionStatus === 'network-error'"
          variant="warning"
          title="No pudimos conectar"
          role="alert"
        >
          Revisa tu conexión e inténtalo de nuevo. No perdiste lo que escribiste.
        </BaseAlert>
        <BaseAlert
          v-if="createExceptionStatus === 'unexpected-error'"
          variant="danger"
          title="Ocurrió un error inesperado"
          role="alert"
        >
          Inténtalo de nuevo en unos segundos. No perdiste lo que escribiste.
        </BaseAlert>

        <BaseInput
          :model-value="createEffectiveDate"
          type="date"
          name="effectiveDate"
          label="Fecha"
          required
          :disabled="createExceptionStatus === 'saving'"
          :error="createEffectiveDateError"
          @update:model-value="
            (v) => {
              createEffectiveDate = String(v)
              revalidateExceptionCreate()
            }
          "
        />

        <div class="schedules-page__field">
          <span class="schedules-page__label">Estado del día</span>
          <div class="schedules-page__radio-group" role="radiogroup" aria-label="Estado del día">
            <label class="schedules-page__checkbox-label">
              <input
                type="radio"
                name="createExceptionShape"
                :checked="createIsClosed"
                :disabled="createExceptionStatus === 'saving'"
                @change="onCreateIsClosedChange(true)"
              />
              Cerrado
            </label>
            <label class="schedules-page__checkbox-label">
              <input
                type="radio"
                name="createExceptionShape"
                :checked="!createIsClosed"
                :disabled="createExceptionStatus === 'saving'"
                @change="onCreateIsClosedChange(false)"
              />
              Abierto con tramos especiales
            </label>
          </div>
        </div>

        <div v-if="!createIsClosed" class="schedules-page__segments">
          <div
            v-for="(segment, index) in createSegments"
            :key="index"
            class="schedules-page__segment-row"
          >
            <BaseInput
              :model-value="segment.startsTime"
              type="time"
              :name="`createSegmentStart${index}`"
              label="Hora de inicio"
              required
              :disabled="createExceptionStatus === 'saving'"
              @update:model-value="
                (v) => {
                  segment.startsTime = String(v)
                  revalidateExceptionCreate()
                }
              "
            />
            <BaseInput
              :model-value="segment.durationMinutes"
              type="number"
              :name="`createSegmentDuration${index}`"
              label="Duración (minutos)"
              required
              :min="1"
              :max="1440"
              :disabled="createExceptionStatus === 'saving'"
              @update:model-value="
                (v) => {
                  segment.durationMinutes = Number(v)
                  revalidateExceptionCreate()
                }
              "
            />
            <BaseButton
              type="button"
              variant="secondary"
              :disabled="createExceptionStatus === 'saving'"
              aria-label="Quitar este tramo"
              @click="removeCreateSegment(index)"
            >
              Quitar
            </BaseButton>
          </div>
          <BaseButton
            type="button"
            variant="secondary"
            :disabled="createExceptionStatus === 'saving'"
            @click="addCreateSegment"
          >
            Agregar tramo
          </BaseButton>
          <div v-if="createShapeError" class="schedules-page__field-error" role="alert">
            {{ createShapeError }}
          </div>
        </div>

        <BaseInput
          :model-value="createReason"
          type="text"
          name="reason"
          label="Motivo (opcional)"
          :maxlength="200"
          :disabled="createExceptionStatus === 'saving'"
          :error="createReasonError"
          @update:model-value="
            (v) => {
              createReason = String(v)
              revalidateExceptionCreate()
            }
          "
        />

        <div class="schedules-page__dialog-actions">
          <BaseButton type="button" variant="secondary" @click="isExceptionCreateOpen = false">
            Cancelar
          </BaseButton>
          <BaseButton
            type="submit"
            variant="primary"
            :loading="createExceptionStatus === 'saving'"
            :disabled="createExceptionStatus === 'saving'"
          >
            Guardar
          </BaseButton>
        </div>
      </form>
    </BaseDialog>

    <!-- HU-041: edición de excepción -->
    <BaseDialog
      v-model="isExceptionEditOpen"
      title="Editar excepción"
      size="sm"
      @close="onExceptionEditDialogClosed"
    >
      <form
        name="updateScheduleException"
        class="schedules-page__form"
        novalidate
        @submit.prevent="onSubmitExceptionEdit"
      >
        <BaseAlert
          v-if="editExceptionStatus === 'date-conflict'"
          variant="danger"
          title="Ya existe otra excepción para esa fecha"
          role="alert"
        >
          Elige una fecha distinta.
        </BaseAlert>
        <BaseAlert
          v-if="editExceptionStatus === 'not-found'"
          variant="warning"
          title="Esta excepción ya no está disponible"
          role="alert"
        >
          Cierra este diálogo y recarga la lista.
        </BaseAlert>
        <BaseAlert
          v-if="editExceptionStatus === 'network-error'"
          variant="warning"
          title="No pudimos conectar"
          role="alert"
        >
          Revisa tu conexión e inténtalo de nuevo. No perdiste lo que escribiste.
        </BaseAlert>
        <BaseAlert
          v-if="editExceptionStatus === 'unexpected-error'"
          variant="danger"
          title="Ocurrió un error inesperado"
          role="alert"
        >
          Inténtalo de nuevo en unos segundos. No perdiste lo que escribiste.
        </BaseAlert>

        <BaseInput
          :model-value="editEffectiveDate"
          type="date"
          name="effectiveDate"
          label="Fecha"
          required
          :disabled="editExceptionStatus === 'saving'"
          :error="editEffectiveDateError"
          @update:model-value="
            (v) => {
              editEffectiveDate = String(v)
              revalidateExceptionEdit()
            }
          "
        />

        <div class="schedules-page__field">
          <span class="schedules-page__label">Estado del día</span>
          <div class="schedules-page__radio-group" role="radiogroup" aria-label="Estado del día">
            <label class="schedules-page__checkbox-label">
              <input
                type="radio"
                name="editExceptionShape"
                :checked="editIsClosed"
                :disabled="editExceptionStatus === 'saving'"
                @change="onEditIsClosedChange(true)"
              />
              Cerrado
            </label>
            <label class="schedules-page__checkbox-label">
              <input
                type="radio"
                name="editExceptionShape"
                :checked="!editIsClosed"
                :disabled="editExceptionStatus === 'saving'"
                @change="onEditIsClosedChange(false)"
              />
              Abierto con tramos especiales
            </label>
          </div>
        </div>

        <div v-if="!editIsClosed" class="schedules-page__segments">
          <div
            v-for="(segment, index) in editSegments"
            :key="index"
            class="schedules-page__segment-row"
          >
            <BaseInput
              :model-value="segment.startsTime"
              type="time"
              :name="`editSegmentStart${index}`"
              label="Hora de inicio"
              required
              :disabled="editExceptionStatus === 'saving'"
              @update:model-value="
                (v) => {
                  segment.startsTime = String(v)
                  revalidateExceptionEdit()
                }
              "
            />
            <BaseInput
              :model-value="segment.durationMinutes"
              type="number"
              :name="`editSegmentDuration${index}`"
              label="Duración (minutos)"
              required
              :min="1"
              :max="1440"
              :disabled="editExceptionStatus === 'saving'"
              @update:model-value="
                (v) => {
                  segment.durationMinutes = Number(v)
                  revalidateExceptionEdit()
                }
              "
            />
            <BaseButton
              type="button"
              variant="secondary"
              :disabled="editExceptionStatus === 'saving'"
              aria-label="Quitar este tramo"
              @click="removeEditSegment(index)"
            >
              Quitar
            </BaseButton>
          </div>
          <BaseButton
            type="button"
            variant="secondary"
            :disabled="editExceptionStatus === 'saving'"
            @click="addEditSegment"
          >
            Agregar tramo
          </BaseButton>
          <div v-if="editShapeError" class="schedules-page__field-error" role="alert">
            {{ editShapeError }}
          </div>
        </div>

        <BaseInput
          :model-value="editReason"
          type="text"
          name="reason"
          label="Motivo (opcional)"
          :maxlength="200"
          :disabled="editExceptionStatus === 'saving'"
          :error="editReasonError"
          @update:model-value="
            (v) => {
              editReason = String(v)
              revalidateExceptionEdit()
            }
          "
        />

        <div class="schedules-page__dialog-actions">
          <BaseButton type="button" variant="secondary" @click="isExceptionEditOpen = false">
            Cancelar
          </BaseButton>
          <BaseButton
            type="submit"
            variant="primary"
            :loading="editExceptionStatus === 'saving'"
            :disabled="editExceptionStatus === 'saving'"
          >
            Guardar
          </BaseButton>
        </div>
      </form>
    </BaseDialog>
  </section>
</template>

<style scoped>
.schedules-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  max-width: 720px;
  padding: var(--space-4);
  margin: 0 auto;
}

.schedules-page__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.schedules-page__title {
  margin: 0;
  font-size: var(--font-size-h1);
  line-height: var(--font-size-h1-line);
  font-weight: var(--font-weight-h1);
  color: var(--color-text-primary);
}

.schedules-page__state {
  padding: var(--space-4);
  color: var(--color-text-secondary);
}

.schedules-page__timezone {
  margin: 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  color: var(--color-text-secondary);
}

.schedules-page__empty {
  padding: var(--space-4);
  color: var(--color-text-secondary);
}

.schedules-page__picker {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.schedules-page__label {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  font-weight: 500;
  color: var(--color-text-primary);
}

.schedules-page__required {
  color: var(--color-danger-action);
}

.schedules-page__select {
  min-height: 44px;
  padding: var(--space-2) var(--space-3);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-text-primary);
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: var(--radius-md);
}

.schedules-page__days {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.schedules-page__day {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.schedules-page__day-title {
  margin: 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  font-weight: 600;
  color: var(--color-text-primary);
}

.schedules-page__day-empty {
  margin: 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  color: var(--color-text-secondary);
}

.schedules-page__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding: 0;
  margin: 0;
  list-style: none;
}

.schedules-page__holiday-calendar,
.schedules-page__exceptions,
.schedules-page__holidays-reference {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding-top: var(--space-4);
  border-top: var(--border-width-normal) solid var(--color-border-subtle);
}

.schedules-page__exceptions-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.schedules-page__checkbox-label {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-text-primary);
}

.schedules-page__radio-group {
  display: flex;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.schedules-page__segments {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.schedules-page__segment-row {
  display: flex;
  align-items: flex-end;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.schedules-page__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-4);
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  flex-wrap: wrap;
}

.schedules-page__item-time {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-text-primary);
}

.schedules-page__item-actions {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.schedules-page__delete-error {
  margin: 0;
}

.schedules-page__form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.schedules-page__field {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.schedules-page__field-error {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  color: var(--color-danger-action);
}

.schedules-page__dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  margin-top: var(--space-5);
  flex-wrap: wrap;
}
</style>
