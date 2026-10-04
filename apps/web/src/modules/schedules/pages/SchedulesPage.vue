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
// (trabajo requerido §4.3/§4.4 de HU-021, mismo criterio aquí). Cada cambio
// confirmado añade un aviso emergente (DEC-095); los errores siguen en línea.
// La semana se dibuja como un tablero (WeeklyBoard): barras de latón sobre una
// regla horaria común, con la lista real de tramos debajo de cada día.
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useToast } from '@/shared/composables'
import { isSoloProfile } from '@/shared/model'
import { getCivilDateInTimezone } from '@/shared/time/civilDate'
import {
  BarberAvatar,
  BaseAlert,
  BaseButton,
  BaseDatePicker,
  BaseDialog,
  BaseInput,
  BaseSelect,
  BaseTimePicker,
  DiamondLoader,
  EmptyScene,
} from '@/shared/ui'
import AccordionItem from '../components/AccordionItem.vue'
import DayTrack from '../components/DayTrack.vue'
import WeekdayPicker from '../components/WeekdayPicker.vue'
import WeeklyBoard from '../components/WeeklyBoard.vue'
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
import {
  civilDateTile as dateTile,
  computeScale,
  endLabel,
  formatHours,
  nowInTimezone,
  type BoardNow,
} from '../model/weekBoard'
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

// Perfil de barbero individual (DEC-115): con un solo barbero no hay selector.
const hideBarberPicker = computed(() => isSoloProfile.value && barbers.value.length === 1)

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

const toast = useToast()

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

function onBarberChange(barberId: string) {
  if (barberId === selectedBarberId.value) return
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

function openCreateDialog(isoWeekday?: number) {
  createISOWeekday.value = isoWeekday ?? 1
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

function onCreateWeekdayInput(value: number) {
  createISOWeekday.value = value
  revalidateCreate()
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
      toast.success('Tramo agregado', { detail: 'El tramo ya forma parte de la jornada semanal.' })
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

function onEditWeekdayInput(value: number) {
  editISOWeekday.value = value
  revalidateEdit()
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
      toast.success('Tramo actualizado', { detail: 'Guardamos los cambios del tramo.' })
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
    toast.success('Tramo retirado', { detail: 'El tramo ya no forma parte de la jornada.' })
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
    toast.success(outcome.enabled ? 'Festivos activados' : 'Festivos desactivados', {
      detail: outcome.enabled
        ? 'Los festivos colombianos quedan cerrados, salvo una excepción manual.'
        : 'Los festivos ya no agregan ningún bloqueo automático.',
    })
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

function onCreateEffectiveDateInput(value: string) {
  createEffectiveDate.value = value
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
      toast.success('Excepción agregada', { detail: 'La excepción ya está en el calendario.' })
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

function onEditEffectiveDateInput(value: string) {
  editEffectiveDate.value = value
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
      toast.success('Excepción actualizada', { detail: 'Guardamos los cambios de la excepción.' })
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
    toast.success('Excepción retirada', { detail: 'La excepción ya no está en el calendario.' })
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

// --- Tablero semanal, cifras y reloj de la barbería -------------------------

const LOADING_PHRASES = [
  'Revisando la semana',
  'Afilando la navaja',
  'Alineando los turnos',
  'Todo a su hora',
] as const

// Duraciones que un barbero escribe una y otra vez: atajos, no una regla.
const DURATION_PRESETS = [
  { minutes: 30, label: '30 min' },
  { minutes: 45, label: '45 min' },
  { minutes: 60, label: '1 h' },
  { minutes: 120, label: '2 h' },
  { minutes: 240, label: '4 h' },
  { minutes: 480, label: '8 h' },
] as const

// Panel lateral en acordeón: un solo apartado abierto a la vez para que la
// pantalla quepa sin desplazarse. Las excepciones abren primero: son lo que
// más se edita; el estado de los festivos se lee en la insignia del apartado.
type AsidePanel = 'holidays' | 'exceptions' | 'upcoming'
const openPanel = ref<AsidePanel | null>('exceptions')

function togglePanel(panel: AsidePanel) {
  openPanel.value = openPanel.value === panel ? null : panel
}

const barberOptions = computed(() =>
  barbers.value.map((barber) => ({ value: barber.id, label: barber.fullName })),
)

const boardScale = computed(() => computeScale(workingHours.value))
const weekMinutes = computed(() =>
  workingHours.value.reduce((sum, wh) => sum + wh.durationMinutes, 0),
)
const activeDays = computed(() => new Set(workingHours.value.map((wh) => wh.isoWeekday)).size)

// "Hoy" y "ahora" se leen en la zona de la barbería, nunca en la del
// dispositivo (CA-040-06). Sin zona conocida no se marca ninguno.
const clock = ref<BoardNow | null>(null)

function refreshClock() {
  if (!barbershopTimezone.value) {
    clock.value = null
    return
  }
  try {
    clock.value = nowInTimezone(barbershopTimezone.value)
  } catch {
    clock.value = null
  }
}

watch(barbershopTimezone, refreshClock, { immediate: true })

let clockTimer: ReturnType<typeof setInterval> | undefined
onMounted(() => {
  clockTimer = setInterval(refreshClock, 30_000)
})
onUnmounted(() => {
  if (clockTimer !== undefined) clearInterval(clockTimer)
})

const todayIsoWeekday = computed(() => clock.value?.isoWeekday ?? null)
const nowMinute = computed(() => clock.value?.minute ?? null)
const todayCivilDate = computed(() =>
  barbershopTimezone.value && clock.value ? getCivilDateInTimezone(barbershopTimezone.value) : null,
)

const exceptionDates = computed(() => new Set(exceptions.value.map((e) => e.effectiveDate)))

function weekdayInitial(isoWeekday: number): string {
  return weekdayLabel(isoWeekday).charAt(0)
}

// buildPreview dibuja el tramo que se está escribiendo sobre la pista de su
// día, junto a los que ese día ya tiene (atenuados): así un solape se ve antes
// de guardar. El servidor sigue siendo quien decide (CA-040-04).
function buildPreview(
  isoWeekday: number,
  startsTime: string,
  durationMinutes: number,
  excludeId: string | null,
) {
  const context = workingHours.value
    .filter((wh) => wh.isoWeekday === isoWeekday && wh.id !== excludeId)
    .map((wh) => ({
      id: wh.id,
      startsTime: wh.startsTime,
      durationMinutes: wh.durationMinutes,
      muted: true,
    }))
  const valid =
    /^\d{2}:\d{2}$/.test(startsTime) &&
    Number.isInteger(durationMinutes) &&
    durationMinutes >= 1 &&
    durationMinutes <= 1440
  const draft = { id: 'draft', startsTime, durationMinutes, muted: false }
  return {
    segments: valid ? [...context, draft] : context,
    scale: computeScale(valid ? [...context, draft] : context),
    text: valid
      ? `${weekdayLabel(isoWeekday)} · ${startsTime} – ${endLabel(startsTime, durationMinutes)} · ${formatHours(durationMinutes)}`
      : 'Elige la hora y la duración para ver cómo queda.',
  }
}

const createPreview = computed(() =>
  buildPreview(createISOWeekday.value, createStartsTime.value, createDurationMinutes.value, null),
)
const editPreview = computed(() =>
  buildPreview(
    editISOWeekday.value,
    editStartsTime.value,
    editDurationMinutes.value,
    editTarget.value?.id ?? null,
  ),
)
</script>

<template>
  <section class="schedules-page" aria-labelledby="schedules-page-title">
    <header class="schedules-page__header">
      <div class="schedules-page__heading">
        <h1 id="schedules-page-title" class="schedules-page__title">Horarios</h1>
        <p class="schedules-page__subtitle">Jornada semanal, festivos y excepciones.</p>
      </div>
      <!-- Controles de la jornada junto al título: sin barberos no hay jornada
           que armar, así que la acción principal tampoco aparece. -->
      <div v-if="pageStatus === 'ready' && barbers.length > 0" class="schedules-page__controls">
        <div class="schedules-page__who">
          <BarberAvatar
            class="schedules-page__portrait"
            size="row"
            :full-name="selectedBarber?.fullName ?? ''"
            :photo-url="selectedBarber?.photoUrl ?? null"
          />
          <BaseSelect
            v-if="!hideBarberPicker"
            id="schedules-barber-select"
            class="schedules-page__select"
            label="Barbero"
            :model-value="selectedBarberId ?? ''"
            :options="barberOptions"
            @update:model-value="onBarberChange"
          />
          <!-- Perfil de barbero individual (DEC-115): con un solo barbero no hay nada que elegir. -->
          <span v-else class="schedules-page__who-name">{{ selectedBarber?.fullName }}</span>
        </div>

        <dl v-if="workingHoursStatus === 'ready'" class="schedules-page__stats">
          <div>
            <dt>Horas por semana</dt>
            <dd>
              <Transition name="schedules-stat" mode="out-in">
                <span :key="weekMinutes">{{ formatHours(weekMinutes) }}</span>
              </Transition>
            </dd>
          </div>
          <div>
            <dt>Días con jornada</dt>
            <dd>
              <Transition name="schedules-stat" mode="out-in">
                <span :key="activeDays"
                  >{{ activeDays }}<small class="schedules-page__stat-total">/7</small></span
                >
              </Transition>
            </dd>
          </div>
        </dl>

        <BaseButton
          type="button"
          variant="primary"
          class="schedules-page__create"
          @click="openCreateDialog()"
        >
          Agregar tramo
        </BaseButton>
      </div>
    </header>

    <!-- Un solo fundido entre carga/error/listo (mismo criterio que Barberos). -->
    <Transition name="schedules-content" mode="out-in">
      <div
        v-if="pageStatus === 'loading'"
        class="schedules-page__state"
        role="status"
        aria-live="polite"
      >
        <DiamondLoader label="Cargando barberos…" layout="inline" :phrases="LOADING_PHRASES" />
        <div class="schedules-page__skeleton" aria-hidden="true">
          <span v-for="n in 4" :key="n" class="schedules-page__skeleton-row">
            <span class="schedules-page__skeleton-bar schedules-page__skeleton-bar--day" />
            <span class="schedules-page__skeleton-bar schedules-page__skeleton-bar--track" />
          </span>
        </div>
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

      <div v-else class="schedules-page__ready">
        <EmptyScene v-if="barbers.length === 0" scene="agenda" class="schedules-page__empty">
          <template #title>{{
            isSoloProfile ? 'Aún no tienes tu perfil.' : 'Aún no tienes barberos registrados.'
          }}</template>
          <template #hint>
            <template v-if="isSoloProfile">
              Créalo en la sección
              <RouterLink :to="{ name: 'staff-barberos' }">“Mi perfil”</RouterLink>
              antes de configurar tu horario.
            </template>
            <template v-else>
              Agrega uno en la sección
              <RouterLink :to="{ name: 'staff-barberos' }">“Barberos”</RouterLink>
              antes de configurar su horario.
            </template>
          </template>
        </EmptyScene>

        <div v-else class="schedules-page__workspace">
          <div class="schedules-page__main">
            <div
              v-if="workingHoursStatus === 'loading'"
              class="schedules-page__state"
              role="status"
              aria-live="polite"
            >
              <DiamondLoader
                :label="`Cargando el horario de ${selectedBarber?.fullName}…`"
                layout="inline"
                :phrases="LOADING_PHRASES"
              />
              <div class="schedules-page__skeleton" aria-hidden="true">
                <span v-for="n in 7" :key="n" class="schedules-page__skeleton-row">
                  <span class="schedules-page__skeleton-bar schedules-page__skeleton-bar--day" />
                  <span class="schedules-page__skeleton-bar schedules-page__skeleton-bar--track" />
                </span>
              </div>
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

              <WeeklyBoard
                :days="groupedByWeekday"
                :scale="boardScale"
                :today-iso-weekday="todayIsoWeekday"
                :now-minute="nowMinute"
                :pending-delete-ids="pendingDeleteIds"
                @add="openCreateDialog"
                @edit="openEditDialog"
                @remove="onDelete"
              />
            </template>

            <p v-if="barbershopTimezone" class="schedules-page__timezone">
              <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.2">
                <circle cx="8" cy="8" r="6" />
                <path d="M8 4.5V8l2.4 1.6" />
              </svg>
              <span
                >Horas en la zona horaria de la barbería:
                <strong>{{ barbershopTimezone }}</strong></span
              >
            </p>
          </div>

          <!-- Ajustes del barbero en acordeón: festivos (HU-041), excepciones de
               jornada (HU-041) y festivos de referencia. -->
          <aside
            class="schedules-page__aside"
            :style="{ '--accordion-count': colombianHolidays.length > 0 ? 3 : 2 }"
            aria-label="Festivos y excepciones"
          >
            <AccordionItem
              id="schedules-holidays"
              title="Calendario de festivos colombianos"
              :open="openPanel === 'holidays'"
              :badge="
                holidayCalendarStatus === 'ready'
                  ? holidayCalendarEnabled
                    ? 'Activo'
                    : 'Inactivo'
                  : undefined
              "
              :badge-tone="holidayCalendarEnabled ? 'on' : 'off'"
              @toggle="togglePanel('holidays')"
            >
              <label
                class="schedules-switch"
                :class="{ 'schedules-switch--on': holidayCalendarEnabled }"
              >
                <input
                  type="checkbox"
                  class="schedules-switch__input"
                  :checked="holidayCalendarEnabled"
                  :disabled="holidayCalendarStatus !== 'ready' || holidayCalendarSaving"
                  @change="onToggleHolidayCalendar"
                />
                <span class="schedules-switch__track" aria-hidden="true">
                  <span class="schedules-switch__thumb" />
                </span>
                <span class="schedules-switch__text"
                  >Cerrar automáticamente los festivos colombianos de este barbero</span
                >
              </label>
              <BaseAlert v-if="holidayCalendarError" variant="warning" role="alert">
                {{ holidayCalendarError }}
              </BaseAlert>
              <p class="schedules-card__hint">
                <template v-if="holidayCalendarEnabled">
                  Activado: un festivo se cierra por defecto, salvo que exista una excepción manual
                  para esa fecha.
                </template>
                <template v-else>
                  Desactivado: los festivos no agregan ningún bloqueo automático.
                </template>
              </p>
            </AccordionItem>

            <AccordionItem
              id="schedules-exceptions"
              title="Excepciones de jornada"
              :open="openPanel === 'exceptions'"
              :badge="exceptionsStatus === 'ready' ? String(sortedExceptions.length) : undefined"
              @toggle="togglePanel('exceptions')"
            >
              <BaseButton
                type="button"
                variant="secondary"
                class="schedules-page__aside-action"
                @click="openExceptionCreateDialog()"
              >
                Agregar excepción
              </BaseButton>

              <div
                v-if="exceptionsStatus === 'loading'"
                class="schedules-page__state schedules-page__state--card"
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

                <p v-if="sortedExceptions.length === 0" class="schedules-card__empty">
                  Sin excepciones registradas.
                </p>

                <TransitionGroup
                  v-else
                  name="schedules-row"
                  tag="ul"
                  class="schedules-card__list"
                  aria-label="Excepciones de jornada"
                >
                  <li
                    v-for="exception in sortedExceptions"
                    :key="exception.id"
                    class="schedules-row"
                  >
                    <span class="schedules-tile" aria-hidden="true">
                      <b>{{ dateTile(exception.effectiveDate).day }}</b>
                      <i>{{ dateTile(exception.effectiveDate).month }}</i>
                    </span>
                    <div class="schedules-row__main">
                      <p class="schedules-row__title">
                        <span v-if="exception.isClosed" class="schedules-row__state">Cerrado</span>
                        <template v-else>
                          <span class="schedules-row__state schedules-row__state--open"
                            >Abierto</span
                          >
                          <span class="schedules-row__segments">{{
                            exception.segments
                              .map((s) => `${s.startsTime} (${s.durationMinutes} min)`)
                              .join(', ')
                          }}</span>
                        </template>
                      </p>
                      <p class="schedules-row__meta">
                        <span class="schedules-row__date">{{ exception.effectiveDate }}</span>
                        <span v-if="exception.reason"> · {{ exception.reason }}</span>
                      </p>
                    </div>
                    <div class="schedules-row__actions">
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
                        class="schedules-row__remove"
                        :loading="isExceptionDeletePending(exception.id)"
                        :disabled="isExceptionDeletePending(exception.id)"
                        :aria-label="`Retirar excepción del ${exception.effectiveDate}`"
                        @click="onDeleteException(exception)"
                      >
                        Retirar
                      </BaseButton>
                    </div>
                  </li>
                </TransitionGroup>
              </template>
            </AccordionItem>

            <AccordionItem
              v-if="colombianHolidays.length > 0"
              id="schedules-upcoming"
              title="Próximos festivos colombianos"
              :open="openPanel === 'upcoming'"
              :badge="String(colombianHolidays.length)"
              @toggle="togglePanel('upcoming')"
            >
              <ul class="schedules-card__list">
                <li
                  v-for="holiday in colombianHolidays"
                  :key="holiday.date"
                  class="schedules-row schedules-row--static"
                >
                  <span class="schedules-tile" aria-hidden="true">
                    <b>{{ dateTile(holiday.date).day }}</b>
                    <i>{{ dateTile(holiday.date).month }}</i>
                  </span>
                  <div class="schedules-row__main">
                    <p class="schedules-row__title">{{ holiday.name }}</p>
                    <p class="schedules-row__meta">
                      <span class="schedules-row__date">{{ holiday.date }}</span>
                    </p>
                  </div>
                  <span v-if="exceptionDates.has(holiday.date)" class="schedules-row__tag">
                    Con excepción
                  </span>
                  <BaseButton
                    v-else
                    type="button"
                    variant="secondary"
                    :aria-label="`Registrar una excepción para el ${holiday.date}, ${holiday.name}`"
                    @click="openExceptionCreateDialog(holiday.date)"
                  >
                    Registrar excepción
                  </BaseButton>
                </li>
              </ul>
            </AccordionItem>
          </aside>
        </div>
      </div>
    </Transition>

    <!-- Alta de tramo -->
    <BaseDialog
      v-model="isCreateOpen"
      title="Agregar tramo"
      :description="`Define cuándo trabaja ${selectedBarber?.fullName ?? 'el barbero'} ese día.`"
      size="md"
      content-class="schedules-page__dialog"
      @close="onCreateDialogClosed"
    >
      <template #icon>
        <span class="schedules-page__dialog-chip" aria-hidden="true">+</span>
      </template>
      <form
        name="createWorkingHour"
        class="schedules-page__ink schedules-page__form"
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

        <WeekdayPicker
          :model-value="createISOWeekday"
          label="Día"
          :disabled="createStatus === 'saving'"
          :error="createWeekdayError"
          @update:model-value="onCreateWeekdayInput"
        />

        <div class="schedules-page__pair">
          <div class="schedules-page__field">
            <BaseTimePicker
              :model-value="createStartsTime"
              label="Hora de inicio"
              required
              :disabled="createStatus === 'saving'"
              @update:model-value="onCreateStartsTimeInput"
            />
            <div v-if="createStartsTimeError" class="schedules-page__field-error" role="alert">
              {{ createStartsTimeError }}
            </div>
          </div>
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
            :class="{ 'schedules-page__input--filled': !!createDurationMinutes }"
            @update:model-value="onCreateDurationInput"
          />
        </div>

        <div class="schedules-page__presets" role="group" aria-label="Duraciones habituales">
          <button
            v-for="preset in DURATION_PRESETS"
            :key="preset.minutes"
            type="button"
            class="schedules-page__preset"
            :aria-pressed="createDurationMinutes === preset.minutes"
            :disabled="createStatus === 'saving'"
            @click="onCreateDurationInput(preset.minutes)"
          >
            {{ preset.label }}
          </button>
        </div>

        <div class="schedules-page__preview" aria-live="polite">
          <span class="schedules-page__eyebrow">Así queda</span>
          <DayTrack
            :segments="createPreview.segments"
            :scale="createPreview.scale"
            :order="0"
            preview
          />
          <p>{{ createPreview.text }}</p>
        </div>

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

    <!-- Edición de tramo -->
    <BaseDialog
      v-model="isEditOpen"
      title="Editar tramo"
      :description="`Ajusta el horario de ${editTarget ? weekdayLabel(editTarget.isoWeekday) : 'ese día'}.`"
      size="md"
      content-class="schedules-page__dialog"
      @close="onEditDialogClosed"
    >
      <template #icon>
        <span class="schedules-page__dialog-chip" aria-hidden="true">{{
          weekdayInitial(editISOWeekday)
        }}</span>
      </template>
      <form
        name="updateWorkingHour"
        class="schedules-page__ink schedules-page__form"
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

        <WeekdayPicker
          :model-value="editISOWeekday"
          label="Día"
          :disabled="editStatus === 'saving'"
          :error="editWeekdayError"
          @update:model-value="onEditWeekdayInput"
        />

        <div class="schedules-page__pair">
          <div class="schedules-page__field">
            <BaseTimePicker
              :model-value="editStartsTime"
              label="Hora de inicio"
              required
              :disabled="editStatus === 'saving'"
              @update:model-value="onEditStartsTimeInput"
            />
            <div v-if="editStartsTimeError" class="schedules-page__field-error" role="alert">
              {{ editStartsTimeError }}
            </div>
          </div>
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
            :class="{ 'schedules-page__input--filled': !!editDurationMinutes }"
            @update:model-value="onEditDurationInput"
          />
        </div>

        <div class="schedules-page__presets" role="group" aria-label="Duraciones habituales">
          <button
            v-for="preset in DURATION_PRESETS"
            :key="preset.minutes"
            type="button"
            class="schedules-page__preset"
            :aria-pressed="editDurationMinutes === preset.minutes"
            :disabled="editStatus === 'saving'"
            @click="onEditDurationInput(preset.minutes)"
          >
            {{ preset.label }}
          </button>
        </div>

        <div class="schedules-page__preview" aria-live="polite">
          <span class="schedules-page__eyebrow">Así queda</span>
          <DayTrack
            :segments="editPreview.segments"
            :scale="editPreview.scale"
            :order="0"
            preview
          />
          <p>{{ editPreview.text }}</p>
        </div>

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
      description="Un día que se sale de la jornada habitual."
      size="md"
      content-class="schedules-page__dialog"
      @close="onExceptionCreateDialogClosed"
    >
      <template #icon>
        <span class="schedules-page__dialog-chip" aria-hidden="true">+</span>
      </template>
      <form
        name="createScheduleException"
        class="schedules-page__ink schedules-page__form"
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

        <div class="schedules-page__field">
          <label
            id="schedules-create-date-label"
            class="schedules-page__label"
            for="schedules-create-date"
            >Fecha <span aria-hidden="true">*</span></label
          >
          <BaseDatePicker
            :model-value="createEffectiveDate"
            trigger-id="schedules-create-date"
            label-id="schedules-create-date-label"
            :today="todayCivilDate"
            required
            :disabled="createExceptionStatus === 'saving'"
            @update:model-value="onCreateEffectiveDateInput"
          />
          <div v-if="createEffectiveDateError" class="schedules-page__field-error" role="alert">
            {{ createEffectiveDateError }}
          </div>
        </div>

        <div class="schedules-page__field">
          <span class="schedules-page__label" aria-hidden="true">Estado del día</span>
          <div class="schedules-page__choice" role="radiogroup" aria-label="Estado del día">
            <label class="schedules-page__choice-option">
              <input
                type="radio"
                name="createExceptionShape"
                :checked="createIsClosed"
                :disabled="createExceptionStatus === 'saving'"
                @change="onCreateIsClosedChange(true)"
              />
              <span>Cerrado</span>
            </label>
            <label class="schedules-page__choice-option">
              <input
                type="radio"
                name="createExceptionShape"
                :checked="!createIsClosed"
                :disabled="createExceptionStatus === 'saving'"
                @change="onCreateIsClosedChange(false)"
              />
              <span>Abierto con tramos especiales</span>
            </label>
          </div>
        </div>

        <div v-if="!createIsClosed" class="schedules-page__segments">
          <div
            v-for="(segment, index) in createSegments"
            :key="index"
            class="schedules-page__segment"
          >
            <BaseTimePicker
              :model-value="segment.startsTime"
              label="Hora de inicio"
              required
              :disabled="createExceptionStatus === 'saving'"
              @update:model-value="
                (v) => {
                  segment.startsTime = v
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
            class="schedules-page__segment-add"
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
          :class="{ 'schedules-page__input--filled': !!createReason }"
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
      description="Ajusta la fecha, el estado o los tramos de ese día."
      size="md"
      content-class="schedules-page__dialog"
      @close="onExceptionEditDialogClosed"
    >
      <template #icon>
        <span class="schedules-page__dialog-chip" aria-hidden="true">{{
          editEffectiveDate ? dateTile(editEffectiveDate).day : ''
        }}</span>
      </template>
      <form
        name="updateScheduleException"
        class="schedules-page__ink schedules-page__form"
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

        <div class="schedules-page__field">
          <label
            id="schedules-edit-date-label"
            class="schedules-page__label"
            for="schedules-edit-date"
            >Fecha <span aria-hidden="true">*</span></label
          >
          <BaseDatePicker
            :model-value="editEffectiveDate"
            trigger-id="schedules-edit-date"
            label-id="schedules-edit-date-label"
            :today="todayCivilDate"
            required
            :disabled="editExceptionStatus === 'saving'"
            @update:model-value="onEditEffectiveDateInput"
          />
          <div v-if="editEffectiveDateError" class="schedules-page__field-error" role="alert">
            {{ editEffectiveDateError }}
          </div>
        </div>

        <div class="schedules-page__field">
          <span class="schedules-page__label" aria-hidden="true">Estado del día</span>
          <div class="schedules-page__choice" role="radiogroup" aria-label="Estado del día">
            <label class="schedules-page__choice-option">
              <input
                type="radio"
                name="editExceptionShape"
                :checked="editIsClosed"
                :disabled="editExceptionStatus === 'saving'"
                @change="onEditIsClosedChange(true)"
              />
              <span>Cerrado</span>
            </label>
            <label class="schedules-page__choice-option">
              <input
                type="radio"
                name="editExceptionShape"
                :checked="!editIsClosed"
                :disabled="editExceptionStatus === 'saving'"
                @change="onEditIsClosedChange(false)"
              />
              <span>Abierto con tramos especiales</span>
            </label>
          </div>
        </div>

        <div v-if="!editIsClosed" class="schedules-page__segments">
          <div
            v-for="(segment, index) in editSegments"
            :key="index"
            class="schedules-page__segment"
          >
            <BaseTimePicker
              :model-value="segment.startsTime"
              label="Hora de inicio"
              required
              :disabled="editExceptionStatus === 'saving'"
              @update:model-value="
                (v) => {
                  segment.startsTime = v
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
            class="schedules-page__segment-add"
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
          :class="{ 'schedules-page__input--filled': !!editReason }"
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
/* Superficie tinta de punta a punta (estandar-diseno-visual.md §3), el mismo
   canvas que Agenda, Servicios y Barberos; columna de lectura de 980px. */
.schedules-page {
  --schedules-width: 980px;

  /* Los controles compartidos (BaseSelect) se pintan con tokens de superficie
     clara: aquí se reasignan a la tinta, igual que en el panel de bloqueos. */
  --color-text-primary: var(--color-on-strong);
  --color-text-secondary: var(--color-on-strong-muted);
  --color-accent-brass: var(--color-brand-accent-surface);
  --color-border-control: var(--color-field-strong-border);
  --color-surface: var(--color-field-strong);

  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 100%;
  padding: 34px 32px 48px;
  color: var(--color-on-strong);
  /* Transparente: la tinta y el fondo animado los pone el cascarón. */
  background: transparent;
  box-sizing: border-box;
}

.schedules-page__header,
.schedules-page__state,
.schedules-page__ready,
.schedules-page > :deep(.base-alert) {
  width: min(100%, var(--schedules-width));
  margin-inline: auto;
}

.schedules-page__ready {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* Cabecera: título a la izquierda y, en escritorio, los controles de la
   jornada (barbero, cifras y acción principal) en la misma franja. */
.schedules-page__header {
  /* Por encima del tablero: la lista del selector de barbero se abre sobre él. */
  position: relative;
  z-index: 3;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 14px var(--space-4);
  padding-bottom: 14px;
  border-bottom: var(--border-width-normal) solid var(--color-field-strong-border);
}

.schedules-page__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h1);
  line-height: var(--font-size-h1-line);
  font-weight: var(--font-weight-h1);
}

@media (min-width: 1024px) {
  .schedules-page__title {
    font-size: var(--font-size-title-page);
    line-height: 46px;
  }
}

.schedules-page__subtitle {
  margin: 4px 0 0;
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
  color: var(--color-on-strong-muted);
}

/* Angosto: los controles se reparten en la cabecera (acción principal junto al
   título, luego el selector y las cifras); en escritorio forman una franja. */
.schedules-page__controls {
  display: contents;
}

.schedules-page__heading {
  flex: 1 1 auto;
}

/* CTA "Agregar tramo": el relleno tinta de BaseButton--primary es el mismo
   color que la página, así que se levanta con el dorado de marca. */
.schedules-page__create.base-button {
  height: 40px;
  padding-inline: 18px;
  margin-left: auto;
  order: 1;
  font-size: var(--font-size-body-sm);
  font-weight: 600;
}

.schedules-page__create :deep(.base-button__content)::before {
  content: '+';
  margin-right: 6px;
}

.schedules-page__create.base-button--primary {
  background-color: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  border-color: var(--color-brand-accent-surface);
}

.schedules-page__create.base-button--primary:hover:not(:disabled):not(.base-button--loading) {
  filter: brightness(92%);
}

.schedules-page__create.base-button--primary:active:not(:disabled):not(.base-button--loading) {
  filter: brightness(84%);
}

/* Fundido entre carga/error/listo: --motion-duration-base, el mismo token que
   el resto de transiciones de estado (estandar-diseno-visual.md §12). */
.schedules-content-enter-active,
.schedules-content-leave-active {
  transition: opacity var(--motion-duration-base) var(--motion-easing-standard);
}

.schedules-content-enter-from,
.schedules-content-leave-to {
  opacity: 0;
}

.schedules-page__state {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  color: var(--color-on-strong-muted);
}

.schedules-page__state--card {
  width: auto;
  padding: var(--space-3) 0;
  margin: 0;
}

.schedules-page__state--card p {
  margin: 0;
}

/* Esqueleto de la semana: una fila por día con su pista, con el pulso de
   Barberos/Servicios/Agenda. */
.schedules-page__skeleton {
  overflow: hidden;
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: 3px;
}

.schedules-page__skeleton-row {
  display: flex;
  align-items: center;
  gap: 18px;
  height: 64px;
  padding: 0 18px;
  border-top: var(--border-width-normal) solid var(--color-field-strong-border);
}

.schedules-page__skeleton-row:first-child {
  border-top: none;
}

.schedules-page__skeleton-bar {
  display: block;
  background-color: color-mix(in srgb, var(--color-on-strong) 14%, transparent);
  border-radius: 2px;
  animation: schedules-skeleton-pulse 1400ms ease-in-out infinite;
}

.schedules-page__skeleton-bar--day {
  flex: 0 0 110px;
  height: 18px;
}

.schedules-page__skeleton-bar--track {
  flex: 1;
  height: 20px;
  background-color: color-mix(in srgb, var(--color-on-strong) 8%, transparent);
}

@keyframes schedules-skeleton-pulse {
  0%,
  100% {
    opacity: 0.6;
  }

  50% {
    opacity: 1;
  }
}

/* Controles: quién (retrato + selector) y cuánto (cifras de la semana). */
.schedules-page__who {
  display: flex;
  order: 2;
  flex: 1 1 300px;
  align-items: flex-end;
  gap: 14px;
  min-width: 0;
  animation: schedules-rise 360ms var(--motion-easing-standard) both;
}

/* Perfil de barbero individual (DEC-115): sin selector, el nombre acompaña al retrato. */
.schedules-page__who-name {
  align-self: center;
  min-width: 0;
  overflow: hidden;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-lg);
  font-weight: 500;
  color: var(--color-on-strong);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.schedules-page__portrait {
  flex: 0 0 auto;
  transition: transform var(--motion-duration-base) cubic-bezier(0.34, 1.56, 0.64, 1);
}

@media (hover: hover) {
  .schedules-page__who:hover .schedules-page__portrait {
    transform: scale(1.05);
  }
}

.schedules-page__select {
  flex: 1;
  max-width: 340px;
  --color-focus: var(--color-brand-accent-surface);
}

.schedules-page__select :deep(.base-select__trigger) {
  background-color: color-mix(in srgb, var(--color-on-strong) 5%, transparent);
  border-color: color-mix(in srgb, var(--color-on-strong) 16%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.schedules-page__select :deep(.base-select__trigger:hover:not(:disabled)),
.schedules-page__select :deep(.base-select__trigger[aria-expanded='true']) {
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
}

.schedules-page__select :deep(.base-select__option:hover),
.schedules-page__select :deep(.base-select__option:focus-visible) {
  background-color: var(--color-field-strong-raised);
}

.schedules-page__select :deep(.base-select__option) {
  border-left-color: transparent;
}

.schedules-page__stats {
  display: flex;
  order: 3;
  flex: 0 0 auto;
  gap: 8px;
  margin: 0;
  animation: schedules-rise 360ms var(--motion-easing-standard) 60ms both;
}

.schedules-page__stats > div {
  display: flex;
  min-width: 112px;
  flex-direction: column;
  gap: 0;
  padding: 6px 12px;
  background-color: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-on-strong) 10%, transparent);
  border-left: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  border-radius: 2px;
}

.schedules-page__stats dt {
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.schedules-page__stats dd {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-title-item);
  line-height: 26px;
  font-variant-numeric: tabular-nums;
}

.schedules-page__stat-total {
  margin-left: 2px;
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-body-sm);
}

.schedules-stat-enter-active,
.schedules-stat-leave-active {
  display: inline-block;
  transition:
    opacity var(--motion-duration-fast) var(--motion-easing-standard),
    transform var(--motion-duration-fast) var(--motion-easing-standard);
}

.schedules-stat-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

.schedules-stat-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

.schedules-page__timezone {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  line-height: var(--font-size-caption-line);
}

.schedules-page__timezone svg {
  flex: 0 0 auto;
  width: 14px;
  height: 14px;
  color: var(--color-brand-accent-surface);
}

.schedules-page__timezone strong {
  color: var(--color-on-strong);
  font-weight: 600;
}

.schedules-page__delete-error {
  margin: 0;
}

/* Espacio de trabajo: el tablero y, aparte, el panel de ajustes en acordeón
   (festivos, excepciones, referencia). En pantallas angostas se apilan. */
.schedules-page__workspace {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.schedules-page__main {
  display: flex;
  min-width: 0;
  min-height: 0;
  flex-direction: column;
  gap: 10px;
}

.schedules-page__aside {
  --accordion-count: 3;

  display: flex;
  min-width: 0;
  flex-direction: column;
  align-self: start;
  width: 100%;
  overflow: hidden;
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-top: var(--border-width-emphasis) solid
    color-mix(in srgb, var(--color-brand-accent-surface) 70%, transparent);
  border-radius: 3px;
  container: aside / inline-size;
  animation: schedules-rise 400ms var(--motion-easing-standard) 180ms both;
}

.schedules-page__aside-action {
  align-self: flex-start;
}

/* Escritorio: todo cabe en la pantalla. El tablero y el panel comparten la
   altura que deja la cabecera; la lista del apartado abierto se desplaza
   dentro de su propio cuerpo si no cabe, nunca la página. */
@media (min-width: 1100px) {
  .schedules-page {
    --schedules-width: 1240px;

    height: 100%;
    min-height: 620px;
    padding: 16px 32px;
    gap: 12px;
  }

  .schedules-page__header {
    flex-wrap: nowrap;
    align-items: center;
    padding-bottom: 12px;
  }

  .schedules-page__title {
    font-size: var(--font-size-title-page);
    line-height: 40px;
  }

  .schedules-page__controls {
    display: flex;
    flex: 1 1 auto;
    flex-wrap: nowrap;
    align-items: center;
    justify-content: flex-end;
    gap: 16px;
  }

  .schedules-page__heading {
    flex: 0 0 auto;
  }

  .schedules-page__who {
    flex: 0 1 auto;
    align-items: center;
  }

  .schedules-page__select {
    width: 260px;
    flex: 0 1 260px;
  }

  .schedules-page__who {
    order: 1;
  }

  .schedules-page__stats {
    order: 2;
  }

  .schedules-page__create.base-button {
    order: 3;
    margin-left: 0;
  }

  .schedules-page__ready {
    flex: 1 1 0;
    min-height: 0;
  }

  .schedules-page__workspace {
    display: grid;
    flex: 1 1 0;
    min-height: 0;
    grid-template-columns: minmax(0, 1fr) clamp(320px, 29vw, 380px);
    grid-template-rows: minmax(0, 1fr);
    gap: 16px;
  }

  .schedules-page__main > :deep(.week-board) {
    flex: 1 1 0;
    overflow-y: auto;
  }

  .schedules-page__aside {
    --accordion-body-max: max(140px, calc(100cqh - var(--accordion-count) * 56px - 4px));

    align-self: stretch;
    container-type: size;
  }
}

.schedules-card__hint,
.schedules-card__empty {
  margin: 0;
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
}

.schedules-card__empty {
  padding: 18px 12px;
  text-align: center;
  border: var(--border-width-normal) dashed
    color-mix(in srgb, var(--color-on-strong) 16%, transparent);
  border-radius: 2px;
}

.schedules-card__list {
  display: flex;
  flex-direction: column;
  padding: 0;
  margin: 0;
  list-style: none;
}

/* Acciones del panel y de fila: botón fantasma de latón, igual que en Barberos. */
.schedules-page__aside-action.base-button,
.schedules-row :deep(.base-button) {
  --btn-focus-ring: 0 0 0 2px var(--color-field-strong), 0 0 0 4px var(--color-focus);

  height: 34px;
  padding-inline: 14px;
  font-size: var(--font-size-body-sm);
}

.schedules-page__aside-action.base-button--secondary,
.schedules-row :deep(.base-button--secondary) {
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.schedules-page__aside-action.base-button--secondary:hover:not(:disabled):not(
    .base-button--loading
  ),
.schedules-row :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 14%, transparent);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
}

.schedules-row :deep(.base-button--secondary:active:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 22%, transparent);
}

.schedules-row
  :deep(
    .schedules-row__remove.base-button--secondary:hover:not(:disabled):not(.base-button--loading)
  ) {
  color: var(--color-danger-on-strong);
  background-color: color-mix(in srgb, var(--color-danger-on-strong) 10%, transparent);
  border-color: color-mix(in srgb, var(--color-danger-on-strong) 55%, transparent);
}

/* Interruptor de festivos: un carril con una pieza que se desliza y se ilumina. */
.schedules-switch {
  display: inline-flex;
  max-width: 100%;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}

.schedules-switch__input {
  position: absolute;
  width: 1px;
  height: 1px;
  margin: -1px;
  overflow: hidden;
  clip-path: inset(50%);
  opacity: 0;
}

.schedules-switch__track {
  position: relative;
  flex: 0 0 auto;
  width: 46px;
  height: 24px;
  background-color: color-mix(in srgb, var(--color-on-strong) 6%, transparent);
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-on-strong) 26%, transparent);
  border-radius: 2px;
  transition:
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard);
}

.schedules-switch__thumb {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 16px;
  height: 16px;
  background-color: var(--color-on-strong-muted);
  transform: rotate(45deg) scale(0.78);
  transition:
    left var(--motion-duration-base) cubic-bezier(0.34, 1.56, 0.64, 1),
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    transform var(--motion-duration-base) cubic-bezier(0.34, 1.56, 0.64, 1);
}

.schedules-switch--on .schedules-switch__track {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 22%, transparent);
  border-color: var(--color-brand-accent-surface);
}

.schedules-switch--on .schedules-switch__thumb {
  left: 25px;
  background-color: var(--color-brand-accent-surface);
  transform: rotate(45deg) scale(0.9);
}

.schedules-switch__input:focus-visible + .schedules-switch__track {
  box-shadow:
    0 0 0 2px var(--color-field-strong),
    0 0 0 4px var(--color-brand-accent-surface);
}

.schedules-switch__input:disabled ~ * {
  opacity: 0.5;
}

.schedules-switch__input:disabled ~ .schedules-switch__text {
  cursor: not-allowed;
}

.schedules-switch__text {
  color: var(--color-on-strong);
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
}

/* Filas de excepciones y festivos: ficha de fecha, dato principal y acciones. */
.schedules-row {
  position: relative;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 6px 12px;
  padding: 10px 4px;
  border-top: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-on-strong) 9%, transparent);
  transition: background-color var(--motion-duration-fast) var(--motion-easing-standard);
}

.schedules-row:first-child {
  border-top: none;
}

.schedules-row--static {
  animation: schedules-rise 320ms var(--motion-easing-standard) both;
  animation-delay: calc(min(var(--row-index, 0), 8) * 45ms + 360ms);
}

@media (hover: hover) {
  .schedules-row:hover {
    background-color: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  }

  .schedules-row:hover .schedules-tile {
    border-color: var(--color-brand-accent-surface);
    transform: translateY(-2px);
  }
}

.schedules-tile {
  display: flex;
  width: 44px;
  height: 48px;
  flex: 0 0 auto;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1px;
  background-color: var(--color-surface-strong);
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent);
  border-top: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  border-radius: 2px;
  transition:
    border-color var(--motion-duration-fast) var(--motion-easing-standard),
    transform var(--motion-duration-base) cubic-bezier(0.34, 1.56, 0.64, 1);
}

.schedules-tile b {
  color: var(--color-on-strong);
  font-family: var(--font-display);
  font-size: var(--font-size-title-item);
  font-weight: 400;
  line-height: 24px;
  font-variant-numeric: tabular-nums;
}

.schedules-tile i {
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  font-style: normal;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.schedules-row__main {
  min-width: 0;
}

.schedules-row__title,
.schedules-row__meta {
  margin: 0;
  overflow-wrap: anywhere;
}

.schedules-row__title {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 4px 8px;
  color: var(--color-on-strong);
  font-size: var(--font-size-body);
  font-weight: 500;
  line-height: 22px;
}

.schedules-row__state {
  color: var(--color-danger-on-strong);
  font-size: var(--font-size-caption);
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.schedules-row__state::before {
  content: '';
  display: inline-block;
  width: 5px;
  height: 5px;
  margin-right: 6px;
  vertical-align: 1px;
  background-color: currentColor;
  transform: rotate(45deg);
}

.schedules-row__state--open {
  color: var(--color-success-on-strong);
}

.schedules-row__segments {
  font-size: var(--font-size-body-sm);
  font-variant-numeric: tabular-nums;
}

.schedules-row__meta {
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  line-height: 18px;
}

.schedules-row__date {
  font-variant-numeric: tabular-nums;
}

.schedules-row__actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.schedules-row__tag {
  color: var(--color-success-on-strong);
  font-size: var(--font-size-caption);
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  white-space: nowrap;
}

/* Alta y retiro de una excepción: la fila entra desde abajo y sale hacia un
   lado; las demás se reacomodan con suavidad. */
.schedules-row-enter-active,
.schedules-row-leave-active,
.schedules-row-move {
  transition:
    opacity 260ms var(--motion-easing-standard),
    transform 260ms var(--motion-easing-standard);
}

.schedules-row-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.schedules-row-leave-to {
  opacity: 0;
  transform: translateX(16px);
}

.schedules-row-leave-active {
  position: absolute;
  right: 0;
  left: 0;
}

/* Diálogos: mismo tinte, filete de latón y campos reglados que los de Barberos.
   Solo aplica dentro de .schedules-page__ink (el contenido que esta plantilla
   pone en el diálogo); el fondo y el encabezado, que BaseDialog renderiza en
   <Teleport to="body">, se sobrescriben en el bloque sin "scoped" del final. */
.schedules-page__dialog-chip {
  position: relative;
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  overflow: hidden;
  background-color: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  border-radius: var(--radius-md);
  font-family: var(--font-display);
  font-size: var(--font-size-title-item);
  line-height: 1;
  animation: schedules-chip-pop var(--motion-duration-base) cubic-bezier(0.34, 1.56, 0.64, 1) 60ms
    both;
}

.schedules-page__dialog-chip::after {
  content: '';
  position: absolute;
  inset: -40% -60%;
  background: linear-gradient(
    75deg,
    transparent 40%,
    color-mix(in srgb, var(--color-on-strong) 50%, transparent) 50%,
    transparent 60%
  );
  transform: translateX(-100%);
  animation: schedules-chip-glint 480ms cubic-bezier(0.5, 0, 0.3, 1) 260ms both;
}

@keyframes schedules-chip-pop {
  from {
    opacity: 0;
    transform: scale(0.5) rotate(-20deg);
  }

  to {
    opacity: 1;
    transform: scale(1) rotate(0deg);
  }
}

@keyframes schedules-chip-glint {
  from {
    transform: translateX(-100%);
  }

  to {
    transform: translateX(100%);
  }
}

.schedules-page__ink {
  --color-text-primary: var(--color-on-strong);
  --color-text-secondary: var(--color-on-strong-muted);
  --color-accent-brass: var(--color-brand-accent-surface);
  --color-border-control: var(--color-field-strong-border);
  --color-surface: var(--color-field-strong);

  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  color: var(--color-on-strong);
}

.schedules-page__ink :deep(.base-input) {
  --input-bg: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  --input-border-color: color-mix(in srgb, var(--color-on-strong) 12%, transparent);
  --input-border-base-color: color-mix(in srgb, var(--color-on-strong) 30%, transparent);
  --input-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);

  color: var(--color-on-strong);
}

.schedules-page__ink :deep(.base-input__label),
.schedules-page__ink :deep(.base-input__required) {
  color: var(--color-brand-accent-surface);
}

.schedules-page__ink :deep(.base-input__required) {
  margin-left: 2px;
}

.schedules-page__ink :deep(.base-input__hint) {
  color: var(--color-on-strong-muted);
}

.schedules-page__ink :deep(.base-input:hover:not(:disabled):not(.base-input--invalid)) {
  border-color: color-mix(in srgb, var(--color-on-strong) 26%, transparent);
}

/* El filete dorado inferior marca el campo YA RESUELTO (mismo criterio que
   Agenda, Servicios y Barberos). */
.schedules-page__ink :deep(.schedules-page__input--filled .base-input) {
  background-color: color-mix(in srgb, var(--color-on-strong) 6%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.schedules-page__ink :deep(.base-input:disabled),
.schedules-page__ink :deep(.base-input--disabled) {
  background-color: var(--input-bg);
  border-color: var(--input-border-color);
  border-bottom-color: color-mix(in srgb, var(--color-on-strong) 20%, transparent);
  color: var(--color-on-strong-muted);
  opacity: 0.45;
}

.schedules-page__ink :deep(.base-input--invalid) {
  background-color: color-mix(in srgb, var(--color-danger-on-strong) 7%, transparent);
  border-color: var(--input-border-color);
  border-bottom-color: var(--color-danger-on-strong);
}

.schedules-page__ink :deep(.base-input__error) {
  color: var(--color-danger-on-strong);
}

.schedules-page__ink :deep(.base-button) {
  --btn-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);
}

.schedules-page__ink :deep(.base-button--primary) {
  background-color: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  border-color: var(--color-brand-accent-surface);
}

.schedules-page__ink :deep(.base-button--primary:hover:not(:disabled):not(.base-button--loading)) {
  filter: brightness(92%);
}

.schedules-page__ink :deep(.base-button--primary:active:not(:disabled):not(.base-button--loading)) {
  filter: brightness(84%);
}

.schedules-page__ink :deep(.base-button--secondary) {
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.schedules-page__ink
  :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
}

.schedules-page__ink
  :deep(.base-button--secondary:active:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 20%, transparent);
}

.schedules-page__form {
  gap: var(--space-5);
}

.schedules-page__field {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 8px;
}

.schedules-page__label {
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  font-weight: 600;
  line-height: 14px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.schedules-page__label span {
  color: var(--color-danger-on-strong);
}

.schedules-page__field-error {
  color: var(--color-danger-on-strong);
  font-size: var(--font-size-body-sm);
}

.schedules-page__pair {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  align-items: start;
  gap: var(--space-4);
}

/* Duraciones habituales: fichas que se encienden de latón al elegirlas. */
.schedules-page__presets {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: calc(var(--space-2) * -1);
}

.schedules-page__preset {
  min-height: 32px;
  padding: 0 12px;
  color: var(--color-brand-accent-surface);
  font-family: var(--font-sans);
  font-size: var(--font-size-caption);
  font-variant-numeric: tabular-nums;
  background: transparent;
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-brand-accent-surface) 40%, transparent);
  border-radius: 2px;
  cursor: pointer;
  transition:
    background-color var(--motion-duration-fast) var(--motion-easing-standard),
    color var(--motion-duration-fast) var(--motion-easing-standard),
    transform var(--motion-duration-base) cubic-bezier(0.34, 1.56, 0.64, 1);
}

@media (hover: hover) {
  .schedules-page__preset:hover:not(:disabled):not([aria-pressed='true']) {
    background-color: color-mix(in srgb, var(--color-brand-accent-surface) 14%, transparent);
    transform: translateY(-2px);
  }
}

.schedules-page__preset[aria-pressed='true'] {
  color: var(--color-brand-accent-text);
  background-color: var(--color-brand-accent-surface);
  border-color: var(--color-brand-accent-surface);
}

.schedules-page__preset:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus);
}

.schedules-page__preset:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

/* "Así queda": la pista del día con la propuesta en latón y lo ya existente
   atenuado, para ver el solape antes de guardar. */
.schedules-page__preview {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px 16px;
  background: var(--color-field-strong);
  border-left: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
}

.schedules-page__eyebrow {
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.schedules-page__preview p {
  margin: 0;
  color: var(--color-on-strong);
  font-size: var(--font-size-body-sm);
  font-variant-numeric: tabular-nums;
}

/* Estado del día: dos opciones como fichas; la elegida se enciende de latón. */
.schedules-page__choice {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.schedules-page__choice-option {
  position: relative;
  display: flex;
  min-height: 48px;
  align-items: center;
  justify-content: center;
  padding: 8px 12px;
  color: var(--color-on-strong);
  font-size: var(--font-size-body-sm);
  text-align: center;
  background-color: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-on-strong) 14%, transparent);
  border-bottom: var(--border-width-emphasis) solid
    color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent);
  border-radius: 2px;
  cursor: pointer;
  transition:
    background-color var(--motion-duration-fast) var(--motion-easing-standard),
    border-color var(--motion-duration-fast) var(--motion-easing-standard);
}

.schedules-page__choice-option input {
  position: absolute;
  inset: 0;
  margin: 0;
  opacity: 0;
  cursor: pointer;
}

.schedules-page__choice-option:has(input:checked) {
  color: var(--color-brand-accent-text);
  font-weight: 600;
  background-color: var(--color-brand-accent-surface);
  border-color: var(--color-brand-accent-surface);
}

.schedules-page__choice-option:has(input:focus-visible) {
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus);
}

.schedules-page__choice-option:has(input:disabled) {
  cursor: not-allowed;
  opacity: 0.45;
}

@media (hover: hover) {
  .schedules-page__choice-option:hover:not(:has(input:checked)):not(:has(input:disabled)) {
    background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
    border-color: color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
  }
}

.schedules-page__segments {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.schedules-page__segment {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
  align-items: end;
  gap: var(--space-3);
  padding: 12px;
  background-color: color-mix(in srgb, var(--color-on-strong) 3%, transparent);
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-on-strong) 10%, transparent);
  border-radius: 2px;
  animation: schedules-rise 260ms var(--motion-easing-standard) both;
}

.schedules-page__segment :deep(.base-button) {
  height: 44px;
}

.schedules-page__segment-add {
  align-self: flex-start;
}

.schedules-page__dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  margin-top: var(--space-2);
  flex-wrap: wrap;
}

/* Filas del panel lateral: si el panel es angosto (columna de escritorio o
   móvil) las acciones bajan bajo el dato en lugar de apretarlo. */
@container aside (max-width: 520px) {
  .schedules-row {
    grid-template-columns: auto minmax(0, 1fr);
    padding-inline: 4px;
  }

  .schedules-row__actions,
  .schedules-row > :deep(.base-button),
  .schedules-row__tag {
    grid-column: 2;
    justify-self: start;
  }
}

@media (max-width: 760px) {
  .schedules-page {
    gap: 12px;
    padding: 16px 16px 28px;
  }

  .schedules-page__header {
    padding-bottom: 10px;
  }

  /* CTA compacto: un "+" en un círculo (el nombre accesible sigue siendo el
     texto del botón). */
  .schedules-page__create.base-button {
    width: var(--control-height-icon);
    height: var(--control-height-icon);
    padding: 0;
    overflow: hidden;
    font-size: 0;
  }

  .schedules-page__create :deep(.base-button__content) {
    font-size: 0;
  }

  .schedules-page__create :deep(.base-button__content)::before {
    margin: 0;
    font-size: var(--font-size-title-item);
  }

  .schedules-page__select {
    max-width: none;
  }

  .schedules-page__stats {
    flex: 1 1 100%;
  }

  .schedules-page__stats > div {
    flex: 1;
    min-width: 0;
  }

  .schedules-page__skeleton-row {
    padding: 0 12px;
  }

  .schedules-page__pair,
  .schedules-page__choice {
    grid-template-columns: minmax(0, 1fr);
  }

  .schedules-page__segment {
    grid-template-columns: minmax(0, 1fr);
  }

  .schedules-page__dialog-actions :deep(.base-button) {
    flex: 1;
  }
}

@media (prefers-reduced-motion: reduce) {
  .schedules-content-enter-active,
  .schedules-content-leave-active,
  .schedules-stat-enter-active,
  .schedules-stat-leave-active,
  .schedules-row-enter-active,
  .schedules-row-leave-active,
  .schedules-row-move,
  .schedules-page__portrait,
  .schedules-tile,
  .schedules-row,
  .schedules-switch__track,
  .schedules-switch__thumb,
  .schedules-page__preset,
  .schedules-page__choice-option {
    transition: none;
  }

  .schedules-page__who,
  .schedules-page__stats,
  .schedules-page__aside,
  .schedules-row--static,
  .schedules-page__segment,
  .schedules-page__dialog-chip,
  .schedules-page__dialog-chip::after {
    animation: none;
  }

  .schedules-page__skeleton-bar {
    animation: none;
    opacity: 0.8;
  }
}
</style>

<style>
/* SIN "scoped" a propósito, por el mismo motivo que en Barberos: BaseDialog.vue
   renderiza su tarjeta, encabezado, título, descripción y botón de cerrar en
   <Teleport to="body">, así que dejan de ser descendientes de .schedules-page
   en el DOM real y un :deep() con alcance de componente nunca los alcanza.
   .schedules-page__dialog es exclusivo de esta pantalla. El !important es
   necesario porque la regla propia de BaseDialog tiene la misma
   especificidad y el orden de inserción de los <style> no está garantizado. */
.schedules-page__dialog.base-dialog {
  background-color: var(--color-surface-strong) !important;
  border: var(--border-width-normal) solid var(--color-field-strong-border) !important;
}

/* Mismo rebote de apertura que los diálogos de Servicios y Barberos. */
.schedules-page__dialog.base-dialog--open {
  transition-timing-function: cubic-bezier(0.34, 1.56, 0.64, 1) !important;
}

@media (prefers-reduced-motion: reduce) {
  .schedules-page__dialog.base-dialog--open {
    transition-timing-function: var(--motion-easing-standard) !important;
  }
}

.schedules-page__dialog .base-dialog__header {
  border-bottom-width: var(--border-width-emphasis) !important;
  border-bottom-color: var(--color-brand-accent-surface) !important;
}

.schedules-page__dialog .base-dialog__title {
  color: var(--color-on-strong) !important;
  overflow-wrap: anywhere;
}

.schedules-page__dialog .base-dialog__description {
  color: var(--color-on-strong-muted) !important;
}

.schedules-page__dialog .base-dialog__close {
  color: var(--color-on-strong-muted) !important;
}

.schedules-page__dialog .base-dialog__close:hover {
  background-color: color-mix(in srgb, var(--color-on-strong) 8%, transparent) !important;
  color: var(--color-on-strong) !important;
}

.schedules-page__dialog .base-dialog__close:focus-visible {
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus) !important;
}
</style>
