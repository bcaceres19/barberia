<script setup lang="ts">
// Pantalla "Horarios y bloqueos" (HU-042): elegir un barbero y mantener sus
// bloqueos puntuales y sus series recurrentes semanales. Alcance de esta
// primera entrega de pantalla (el backend ya soporta más, ver prompt
// HU-042): alta/listado/retiro lógico de bloqueos puntuales y de series
// weekly. Fechas explícitas de una serie date_list, excepciones ("esta
// instancia no") y la edición de series (scope whole/this_and_following)
// todavía no tienen controles propios aquí; quedan como seguimiento
// explícito, no como huecos silenciosos.
import { computed, onMounted, ref } from 'vue'
import { BaseAlert, BaseButton, BaseDialog, BaseInput, PageHeader, RecordRow } from '@/shared/ui'
import { fetchBarberSummaries, fetchBarbershopTimezone } from '../api/schedulesApi'
import {
  createTimeBlock,
  createTimeBlockSeries,
  deleteTimeBlock,
  deleteTimeBlockSeries,
  fetchTimeBlockSeries,
  fetchTimeBlocks,
} from '../api/timeBlocksApi'
import { civilDateTimeToInstant, formatInstantInTimezone } from '../model/civilTime'
import { newIdempotencyKey } from '../model/idempotencyKey'
import {
  BLOCK_TYPES,
  blockTypeLabel,
  type BlockType,
  type TimeBlock,
  type TimeBlockSeries,
} from '../model/timeBlock'
import { ISO_WEEKDAYS, weekdayLabel, type BarberSummary } from '../model/workingHour'
import {
  validateBlockInterval,
  validateEffectiveFrom,
  validateISOWeekday,
  validateInstant,
  validateReason,
  validateStartsTime,
} from '../validation/timeBlockValidation'
import { validateDurationMinutes } from '../validation/scheduleValidation'

type PageStatus = 'loading' | 'ready' | 'load-error'
type ListStatus = 'idle' | 'loading' | 'ready' | 'error'
type SaveStatus =
  | 'idle'
  | 'saving'
  | 'validation-error'
  | 'idempotency-conflict'
  | 'not-found'
  | 'network-error'
  | 'unexpected-error'

const pageStatus = ref<PageStatus>('loading')
const barbers = ref<BarberSummary[]>([])
const selectedBarberId = ref<string | null>(null)
const barbershopTimezone = ref<string | null>(null)

const blocksStatus = ref<ListStatus>('idle')
const blocks = ref<TimeBlock[]>([])
const seriesStatus = ref<ListStatus>('idle')
const series = ref<TimeBlockSeries[]>([])
const pendingDeleteBlockIds = ref<Set<string>>(new Set())
const pendingDeleteSeriesIds = ref<Set<string>>(new Set())

// upcomingBlocks: bloqueos NO retirados lógicamente, ordenados por inicio
// (ya vienen así de la API); RN-BLQ-04 exige que un bloqueo retirado deje
// de mostrarse aquí sin desaparecer del sistema.
const upcomingBlocks = computed(() => blocks.value.filter((b) => !b.deletedAt))
const activeSeries = computed(() => series.value.filter((s) => !s.deletedAt))

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
  blocksStatus.value = 'loading'
  seriesStatus.value = 'loading'

  const [blocksOutcome, seriesOutcome] = await Promise.all([
    fetchTimeBlocks(barberId),
    fetchTimeBlockSeries(barberId),
  ])
  if (selectedBarberId.value !== barberId) return

  if (blocksOutcome.kind === 'success') {
    blocks.value = blocksOutcome.page.items
    blocksStatus.value = 'ready'
  } else {
    blocksStatus.value = 'error'
  }

  if (seriesOutcome.kind === 'success') {
    series.value = seriesOutcome.page.items
    seriesStatus.value = 'ready'
  } else {
    seriesStatus.value = 'error'
  }
}

function onBarberSelectChange(event: Event) {
  const barberId = (event.target as HTMLSelectElement).value
  void selectBarber(barberId)
}

function onRetryBlocks() {
  if (selectedBarberId.value) void selectBarber(selectedBarberId.value)
}

// --- Alta de bloqueo puntual --------------------------------------------

const isCreateBlockOpen = ref(false)
const createBlockType = ref<BlockType>('emergency')
const createStartDate = ref('')
const createStartTime = ref('')
const createEndDate = ref('')
const createEndTime = ref('')
const createBlockReason = ref('')
const createBlockError = ref<string | undefined>(undefined)
const createBlockStatus = ref<SaveStatus>('idle')
const createBlockAttempted = ref(false)
let createBlockIdempotencyKey = newIdempotencyKey()

function openCreateBlockDialog() {
  createBlockType.value = 'emergency'
  createStartDate.value = ''
  createStartTime.value = ''
  createEndDate.value = ''
  createEndTime.value = ''
  createBlockReason.value = ''
  createBlockError.value = undefined
  createBlockStatus.value = 'idle'
  createBlockAttempted.value = false
  createBlockIdempotencyKey = newIdempotencyKey()
  isCreateBlockOpen.value = true
}

function onCreateBlockDialogClosed() {
  createBlockStatus.value = 'idle'
}

async function onSubmitCreateBlock() {
  if (createBlockStatus.value === 'saving' || !selectedBarberId.value) return
  createBlockAttempted.value = true
  createBlockError.value = undefined

  const timeZone = barbershopTimezone.value
  if (!timeZone) {
    createBlockError.value = 'No pudimos determinar la zona horaria de la barbería. Reintenta.'
    return
  }

  const startInstantRaw = `${createStartDate.value}T${createStartTime.value}`
  const endInstantRaw = `${createEndDate.value}T${createEndTime.value}`
  if (
    !createStartDate.value ||
    !createStartTime.value ||
    !createEndDate.value ||
    !createEndTime.value
  ) {
    createBlockError.value = 'Completa fecha y hora de inicio y de fin.'
    return
  }

  let startsAt: string
  let endsAt: string
  try {
    startsAt = civilDateTimeToInstant(createStartDate.value, createStartTime.value, timeZone)
    endsAt = civilDateTimeToInstant(createEndDate.value, createEndTime.value, timeZone)
  } catch {
    createBlockError.value = validateInstant(startInstantRaw) ?? validateInstant(endInstantRaw)
    return
  }

  const intervalError = validateBlockInterval(startsAt, endsAt)
  if (intervalError) {
    createBlockError.value = intervalError
    return
  }
  const reasonError = validateReason(createBlockReason.value)
  if (reasonError) {
    createBlockError.value = reasonError
    return
  }

  createBlockStatus.value = 'saving'
  const outcome = await createTimeBlock(
    selectedBarberId.value,
    createBlockType.value,
    startsAt,
    endsAt,
    createBlockReason.value.trim() === '' ? null : createBlockReason.value.trim(),
    createBlockIdempotencyKey,
  )

  switch (outcome.kind) {
    case 'success':
      blocks.value.push(outcome.block)
      isCreateBlockOpen.value = false
      createBlockStatus.value = 'idle'
      return
    case 'validation-error':
      createBlockStatus.value = 'validation-error'
      return
    case 'idempotency-conflict':
      createBlockStatus.value = 'idempotency-conflict'
      return
    case 'not-found':
      createBlockStatus.value = 'not-found'
      return
    case 'network-error':
      createBlockStatus.value = 'network-error'
      return
    case 'unexpected-error':
      createBlockStatus.value = 'unexpected-error'
  }
}

async function onDeleteBlock(block: TimeBlock) {
  if (!selectedBarberId.value || pendingDeleteBlockIds.value.has(block.id)) return
  pendingDeleteBlockIds.value = new Set(pendingDeleteBlockIds.value).add(block.id)

  const outcome = await deleteTimeBlock(selectedBarberId.value, block.id)
  const next = new Set(pendingDeleteBlockIds.value)
  next.delete(block.id)
  pendingDeleteBlockIds.value = next

  if (outcome.kind === 'success' || outcome.kind === 'not-found') {
    blocks.value = blocks.value.filter((b) => b.id !== block.id)
  }
}

function displayInstant(instant: string): string {
  if (!barbershopTimezone.value) return instant
  const { date, time } = formatInstantInTimezone(instant, barbershopTimezone.value)
  return `${date} ${time}`
}

// --- Alta de serie semanal ------------------------------------------------

const isCreateSeriesOpen = ref(false)
const createSeriesType = ref<BlockType>('lunch')
const createSeriesWeekday = ref(1)
const createSeriesStartsTime = ref('')
const createSeriesDuration = ref(60)
const createSeriesEffectiveFrom = ref('')
const createSeriesReason = ref('')
const createSeriesError = ref<string | undefined>(undefined)
const createSeriesStatus = ref<SaveStatus>('idle')
let createSeriesIdempotencyKey = newIdempotencyKey()

function openCreateSeriesDialog() {
  createSeriesType.value = 'lunch'
  createSeriesWeekday.value = 1
  createSeriesStartsTime.value = ''
  createSeriesDuration.value = 60
  createSeriesEffectiveFrom.value = ''
  createSeriesReason.value = ''
  createSeriesError.value = undefined
  createSeriesStatus.value = 'idle'
  createSeriesIdempotencyKey = newIdempotencyKey()
  isCreateSeriesOpen.value = true
}

function onCreateSeriesDialogClosed() {
  createSeriesStatus.value = 'idle'
}

async function onSubmitCreateSeries() {
  if (createSeriesStatus.value === 'saving' || !selectedBarberId.value) return
  createSeriesError.value = undefined

  const weekdayError = validateISOWeekday(createSeriesWeekday.value)
  const startsTimeError = validateStartsTime(createSeriesStartsTime.value)
  const durationError = validateDurationMinutes(createSeriesDuration.value)
  const effectiveFromError = validateEffectiveFrom(createSeriesEffectiveFrom.value)
  const reasonError = validateReason(createSeriesReason.value)
  const firstError =
    weekdayError ?? startsTimeError ?? durationError ?? effectiveFromError ?? reasonError
  if (firstError) {
    createSeriesError.value = firstError
    return
  }

  createSeriesStatus.value = 'saving'
  const outcome = await createTimeBlockSeries(
    selectedBarberId.value,
    createSeriesType.value,
    createSeriesWeekday.value,
    createSeriesStartsTime.value,
    createSeriesDuration.value,
    createSeriesEffectiveFrom.value,
    createSeriesReason.value.trim() === '' ? null : createSeriesReason.value.trim(),
    createSeriesIdempotencyKey,
  )

  switch (outcome.kind) {
    case 'success':
      series.value.push(outcome.series)
      isCreateSeriesOpen.value = false
      createSeriesStatus.value = 'idle'
      return
    case 'validation-error':
      createSeriesStatus.value = 'validation-error'
      return
    case 'idempotency-conflict':
      createSeriesStatus.value = 'idempotency-conflict'
      return
    case 'not-found':
      createSeriesStatus.value = 'not-found'
      return
    case 'network-error':
      createSeriesStatus.value = 'network-error'
      return
    case 'unexpected-error':
      createSeriesStatus.value = 'unexpected-error'
  }
}

async function onDeleteSeries(item: TimeBlockSeries) {
  if (!selectedBarberId.value || pendingDeleteSeriesIds.value.has(item.id)) return
  pendingDeleteSeriesIds.value = new Set(pendingDeleteSeriesIds.value).add(item.id)

  const outcome = await deleteTimeBlockSeries(selectedBarberId.value, item.id)
  const next = new Set(pendingDeleteSeriesIds.value)
  next.delete(item.id)
  pendingDeleteSeriesIds.value = next

  if (outcome.kind === 'success' || outcome.kind === 'not-found') {
    series.value = series.value.filter((s) => s.id !== item.id)
  }
}
</script>

<template>
  <section class="blocks-page" aria-labelledby="blocks-page-title">
    <PageHeader title-id="blocks-page-title" title="Horarios y bloqueos">
      <template v-if="pageStatus === 'ready' && barbers.length > 0" #actions>
        <BaseButton type="button" variant="primary" @click="openCreateBlockDialog">
          Agregar bloqueo
        </BaseButton>
        <BaseButton type="button" variant="secondary" @click="openCreateSeriesDialog">
          Agregar serie semanal
        </BaseButton>
      </template>
    </PageHeader>

    <p v-if="barbershopTimezone" class="blocks-page__timezone">
      Horas en la zona horaria de la barbería: {{ barbershopTimezone }}
    </p>

    <div
      v-if="pageStatus === 'loading'"
      class="blocks-page__state"
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
      <p v-if="barbers.length === 0" class="blocks-page__empty">
        Aún no tienes barberos registrados. Agrega uno en la sección "Barberos" antes de configurar
        sus bloqueos.
      </p>

      <template v-else>
        <div class="blocks-page__picker">
          <label for="blocks-barber-select" class="blocks-page__label">Barbero</label>
          <select
            id="blocks-barber-select"
            class="blocks-page__select"
            :value="selectedBarberId ?? ''"
            @change="onBarberSelectChange"
          >
            <option v-for="barber in barbers" :key="barber.id" :value="barber.id">
              {{ barber.fullName }}
            </option>
          </select>
        </div>

        <div v-if="blocksStatus === 'loading'" role="status" aria-live="polite">
          <p>Cargando bloqueos…</p>
        </div>
        <BaseAlert v-else-if="blocksStatus === 'error'" variant="warning" role="alert">
          No pudimos cargar los bloqueos de este barbero.
          <template #action>
            <BaseButton type="button" variant="secondary" @click="onRetryBlocks"
              >Reintentar</BaseButton
            >
          </template>
        </BaseAlert>
        <template v-else-if="blocksStatus === 'ready'">
          <h2 class="blocks-page__section-title">Próximos bloqueos puntuales</h2>
          <p v-if="upcomingBlocks.length === 0" class="blocks-page__empty">
            Este barbero no tiene bloqueos puntuales vigentes.
          </p>
          <ul v-else class="blocks-page__list">
            <RecordRow v-for="block in upcomingBlocks" :key="block.id">
              <strong>{{ blockTypeLabel(block.blockType) }}</strong>
              <span>
                — {{ displayInstant(block.startsAt) }} a {{ displayInstant(block.endsAt) }}</span
              >
              <p v-if="block.reason" class="blocks-page__reason">{{ block.reason }}</p>
              <template #trailing>
                <BaseButton
                  type="button"
                  variant="danger"
                  :disabled="pendingDeleteBlockIds.has(block.id)"
                  @click="onDeleteBlock(block)"
                >
                  Retirar
                </BaseButton>
              </template>
            </RecordRow>
          </ul>

          <h2 class="blocks-page__section-title">Series semanales</h2>
          <p v-if="activeSeries.length === 0" class="blocks-page__empty">
            Este barbero no tiene series de bloqueo configuradas.
          </p>
          <ul v-else class="blocks-page__list">
            <RecordRow v-for="item in activeSeries" :key="item.id">
              <strong>{{ blockTypeLabel(item.blockType) }}</strong>
              <span v-if="item.recurrenceKind === 'weekly' && item.isoWeekday">
                — {{ weekdayLabel(item.isoWeekday) }}, {{ item.startsTime }} ({{
                  item.durationMinutes
                }}
                min)
              </span>
              <span v-else> — lista de fechas explícitas</span>
              <p v-if="item.reason" class="blocks-page__reason">{{ item.reason }}</p>
              <template #trailing>
                <BaseButton
                  type="button"
                  variant="danger"
                  :disabled="pendingDeleteSeriesIds.has(item.id)"
                  @click="onDeleteSeries(item)"
                >
                  Retirar
                </BaseButton>
              </template>
            </RecordRow>
          </ul>
        </template>
      </template>
    </template>

    <BaseDialog
      v-model="isCreateBlockOpen"
      title="Agregar bloqueo puntual"
      size="md"
      @close="onCreateBlockDialogClosed"
    >
      <form class="blocks-page__form" @submit.prevent="onSubmitCreateBlock">
        <label for="create-block-type" class="blocks-page__label">Tipo de bloqueo</label>
        <select id="create-block-type" v-model="createBlockType" class="blocks-page__select">
          <option v-for="t in BLOCK_TYPES" :key="t.value" :value="t.value">{{ t.label }}</option>
        </select>

        <div class="blocks-page__form-row">
          <BaseInput v-model="createStartDate" type="date" label="Fecha de inicio" required />
          <BaseInput v-model="createStartTime" type="time" label="Hora de inicio" required />
        </div>
        <div class="blocks-page__form-row">
          <BaseInput v-model="createEndDate" type="date" label="Fecha de fin" required />
          <BaseInput v-model="createEndTime" type="time" label="Hora de fin" required />
        </div>
        <BaseInput v-model="createBlockReason" type="text" label="Motivo (opcional)" />

        <BaseAlert v-if="createBlockError" variant="danger" role="alert">{{
          createBlockError
        }}</BaseAlert>
        <BaseAlert
          v-else-if="createBlockStatus === 'validation-error'"
          variant="danger"
          role="alert"
        >
          Revisa los datos del bloqueo.
        </BaseAlert>
        <BaseAlert v-else-if="createBlockStatus === 'not-found'" variant="danger" role="alert">
          Este barbero ya no está disponible.
        </BaseAlert>
        <BaseAlert
          v-else-if="
            createBlockStatus === 'network-error' || createBlockStatus === 'unexpected-error'
          "
          variant="danger"
          role="alert"
        >
          No pudimos guardar el bloqueo. Inténtalo de nuevo.
        </BaseAlert>

        <BaseButton type="submit" variant="primary" :disabled="createBlockStatus === 'saving'">
          {{ createBlockStatus === 'saving' ? 'Guardando…' : 'Guardar bloqueo' }}
        </BaseButton>
      </form>
    </BaseDialog>

    <BaseDialog
      v-model="isCreateSeriesOpen"
      title="Agregar serie semanal"
      size="md"
      @close="onCreateSeriesDialogClosed"
    >
      <form class="blocks-page__form" @submit.prevent="onSubmitCreateSeries">
        <label for="create-series-type" class="blocks-page__label">Tipo de bloqueo</label>
        <select id="create-series-type" v-model="createSeriesType" class="blocks-page__select">
          <option v-for="t in BLOCK_TYPES" :key="t.value" :value="t.value">{{ t.label }}</option>
        </select>

        <label for="create-series-weekday" class="blocks-page__label">Día de la semana</label>
        <select
          id="create-series-weekday"
          v-model.number="createSeriesWeekday"
          class="blocks-page__select"
        >
          <option v-for="d in ISO_WEEKDAYS" :key="d.value" :value="d.value">{{ d.label }}</option>
        </select>

        <BaseInput v-model="createSeriesStartsTime" type="time" label="Hora de inicio" required />
        <BaseInput
          v-model.number="createSeriesDuration"
          type="number"
          label="Duración (minutos)"
          required
        />
        <BaseInput v-model="createSeriesEffectiveFrom" type="date" label="Vigente desde" required />
        <BaseInput v-model="createSeriesReason" type="text" label="Motivo (opcional)" />

        <BaseAlert v-if="createSeriesError" variant="danger" role="alert">{{
          createSeriesError
        }}</BaseAlert>
        <BaseAlert
          v-else-if="createSeriesStatus === 'validation-error'"
          variant="danger"
          role="alert"
        >
          Revisa los datos de la serie.
        </BaseAlert>
        <BaseAlert v-else-if="createSeriesStatus === 'not-found'" variant="danger" role="alert">
          Este barbero ya no está disponible.
        </BaseAlert>
        <BaseAlert
          v-else-if="
            createSeriesStatus === 'network-error' || createSeriesStatus === 'unexpected-error'
          "
          variant="danger"
          role="alert"
        >
          No pudimos guardar la serie. Inténtalo de nuevo.
        </BaseAlert>

        <BaseButton type="submit" variant="primary" :disabled="createSeriesStatus === 'saving'">
          {{ createSeriesStatus === 'saving' ? 'Guardando…' : 'Guardar serie' }}
        </BaseButton>
      </form>
    </BaseDialog>
  </section>
</template>

<style scoped>
.blocks-page__picker {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-block: 1rem;
  max-width: 24rem;
}

.blocks-page__select {
  min-height: 44px;
  padding: 0.5rem 0.75rem;
  border: var(--border-width-normal) solid var(--color-border-control);
  border-radius: var(--radius-md);
}

.blocks-page__list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.blocks-page__reason {
  margin: var(--space-1) 0 0;
  color: var(--color-text-secondary);
}

.blocks-page__section-title {
  margin-top: var(--space-6);
  font-size: var(--font-size-h2);
  line-height: var(--font-size-h2-line);
  font-weight: var(--font-weight-h2);
  color: var(--color-text-primary);
}

.blocks-page__form {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.blocks-page__form-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.blocks-page__form-row > * {
  flex: 1 1 10rem;
}
</style>
