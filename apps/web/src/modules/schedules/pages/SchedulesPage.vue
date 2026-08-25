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
import { newIdempotencyKey } from '../model/idempotencyKey'
import { ISO_WEEKDAYS, weekdayLabel, type BarberSummary, type WorkingHour } from '../model/workingHour'
import {
  validateDurationMinutes,
  validateISOWeekday,
  validateStartsTime,
} from '../validation/scheduleValidation'

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
    return
  }
  workingHoursStatus.value = 'error'
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
        Aún no tienes barberos registrados. Agrega uno en la sección "Barberos" antes de
        configurar su horario.
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
          <BaseAlert v-if="deleteError" variant="danger" role="alert" class="schedules-page__delete-error">
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
      </template>
    </template>

    <!-- Alta -->
    <BaseDialog v-model="isCreateOpen" title="Agregar tramo" size="sm" @close="onCreateDialogClosed">
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
  font-size: var(--font-size-heading-lg);
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
