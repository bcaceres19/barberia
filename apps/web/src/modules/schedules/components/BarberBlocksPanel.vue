<script setup lang="ts">
// HU-042 / DEC-105: una instancia por barbero, compuesta en app y montada
// con key=barberId. Los formularios y las peticiones conservan ese dueño.
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { getCivilDateInTimezone, shiftCivilDate } from '@/shared/time/civilDate'
import { useToast, useVocabulary } from '@/shared/composables'
import {
  BaseAlert,
  BaseButton,
  BaseDialog,
  BaseInput,
  BaseDatePicker,
  BaseTimePicker,
  BaseSelect,
  BarberAvatar,
  BaseSpinner,
} from '@/shared/ui'
import { fetchBarbershopTimezone } from '../api/schedulesApi'
import {
  createTimeBlock,
  createTimeBlockSeries,
  deleteTimeBlock,
  deleteTimeBlockSeries,
  fetchTimeBlockSeries,
  fetchTimeBlocks,
} from '../api/timeBlocksApi'
import { civilDateTimeToInstant, formatInstantInTimezone } from '../model/civilTime'
import { displayCivilDate } from '../model/displayDate'
import { newIdempotencyKey } from '../model/idempotencyKey'
import {
  BLOCK_TYPES,
  blockTypeLabel,
  type BlockType,
  type TimeBlock,
  type TimeBlockSeries,
} from '../model/timeBlock'
import { ISO_WEEKDAYS, weekdayLabel } from '../model/workingHour'
import {
  validateBlockInterval,
  validateEffectiveFrom,
  validateISOWeekday,
  validateInstant,
  validateReason,
  validateStartsTime,
} from '../validation/timeBlockValidation'
import { validateDurationMinutes } from '../validation/scheduleValidation'
import SeriesFormDialog from './SeriesFormDialog.vue'
import SeriesInstancesDialog from './SeriesInstancesDialog.vue'

type ListStatus = 'idle' | 'loading' | 'ready' | 'error'
type SaveStatus =
  | 'idle'
  | 'saving'
  | 'validation-error'
  | 'idempotency-conflict'
  | 'not-found'
  | 'network-error'
  | 'unexpected-error'

const props = defineProps<{ barberId: string; barberName: string; photoUrl?: string | null }>()
const selectedBarberId = computed(() => props.barberId)
const barbershopTimezone = ref<string | null>(null)
const timezoneLoading = ref(false)
const todayCivilDate = computed(() =>
  barbershopTimezone.value ? getCivilDateInTimezone(barbershopTimezone.value) : null,
)
const blocksCursor = ref<string | null>(null)
const seriesCursor = ref<string | null>(null)
const loadingMoreBlocks = ref(false)
const loadingMoreSeries = ref(false)
const moreBlocksError = ref(false)
const moreSeriesError = ref(false)
let requestToken = 0
let alive = true
onUnmounted(() => {
  alive = false
  requestToken++
})

const blocksStatus = ref<ListStatus>('idle')
const blocks = ref<TimeBlock[]>([])
const seriesStatus = ref<ListStatus>('idle')
const series = ref<TimeBlockSeries[]>([])
const pendingDeleteBlockIds = ref<Set<string>>(new Set())
const pendingDeleteSeriesIds = ref<Set<string>>(new Set())

// visibleBlocks: bloqueos NO retirados lógicamente, ordenados por inicio
// (ya vienen así de la API); RN-BLQ-04 exige que un bloqueo retirado deje
// de mostrarse aquí sin desaparecer del sistema.
const visibleBlocks = computed(() => blocks.value.filter((b) => !b.deletedAt))
const activeSeries = computed(() => series.value.filter((s) => !s.deletedAt))

const toast = useToast()
const v = useVocabulary()

async function loadTimezone() {
  timezoneLoading.value = true
  const outcome = await fetchBarbershopTimezone()
  if (!alive) return
  barbershopTimezone.value = outcome.kind === 'success' ? outcome.timezone : null
  timezoneLoading.value = false
}

async function loadLists() {
  const token = ++requestToken
  blocksStatus.value = 'loading'
  seriesStatus.value = 'loading'
  const [blocksOutcome, seriesOutcome] = await Promise.all([
    fetchTimeBlocks(props.barberId),
    fetchTimeBlockSeries(props.barberId),
  ])
  if (!alive || token !== requestToken) return
  if (blocksOutcome.kind === 'success') {
    blocks.value = blocksOutcome.page.items
    blocksCursor.value = blocksOutcome.page.nextCursor
    blocksStatus.value = 'ready'
  } else blocksStatus.value = 'error'
  if (seriesOutcome.kind === 'success') {
    series.value = seriesOutcome.page.items
    seriesCursor.value = seriesOutcome.page.nextCursor
    seriesStatus.value = 'ready'
  } else seriesStatus.value = 'error'
}

onMounted(() => {
  void loadTimezone()
  void loadLists()
})

async function loadMore(kind: 'blocks' | 'series') {
  const isBlocks = kind === 'blocks'
  const cursor = isBlocks ? blocksCursor.value : seriesCursor.value
  const busy = isBlocks ? loadingMoreBlocks : loadingMoreSeries
  const error = isBlocks ? moreBlocksError : moreSeriesError
  if (!cursor || busy.value) return
  busy.value = true
  error.value = false
  const token = requestToken
  if (isBlocks) {
    const outcome = await fetchTimeBlocks(props.barberId, cursor)
    if (alive && token === requestToken) {
      if (outcome.kind === 'success') {
        const known = new Set(blocks.value.map((item) => item.id))
        blocks.value.push(...outcome.page.items.filter((item) => !known.has(item.id)))
        blocksCursor.value = outcome.page.nextCursor
      } else error.value = true
    }
  } else {
    const outcome = await fetchTimeBlockSeries(props.barberId, cursor)
    if (alive && token === requestToken) {
      if (outcome.kind === 'success') {
        const known = new Set(series.value.map((item) => item.id))
        series.value.push(...outcome.page.items.filter((item) => !known.has(item.id)))
        seriesCursor.value = outcome.page.nextCursor
      } else error.value = true
    }
  }
  busy.value = false
}

function isPast(block: TimeBlock): boolean {
  return Date.parse(block.endsAt) <= Date.now()
}
const displayDate = displayCivilDate
const blockPreview = computed(() =>
  createStartDate.value && createEndDate.value
    ? `${displayDate(createStartDate.value)} · ${createStartTime.value || '—'} → ${displayDate(createEndDate.value)} · ${createEndTime.value || '—'}`
    : 'Elige cuándo empieza y termina este tiempo fuera.',
)
const seriesPreview = computed(
  () =>
    `${weekdayLabel(createSeriesWeekday.value)} · ${createSeriesStartsTime.value || '—'} · ${createSeriesDuration.value || '—'} min`,
)

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
  createBlockIdempotencyKey = newIdempotencyKey()
  isCreateBlockOpen.value = true
}

function onCreateBlockDialogClosed() {
  createBlockStatus.value = 'idle'
}

async function onSubmitCreateBlock() {
  if (createBlockStatus.value === 'saving' || !selectedBarberId.value) return
  createBlockError.value = undefined

  const timeZone = barbershopTimezone.value
  if (!timeZone) {
    createBlockError.value = `No pudimos determinar la zona horaria ${v.value.ofTheBusiness}. Reintenta.`
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
    createBlockError.value =
      validateInstant(startInstantRaw) ??
      validateInstant(endInstantRaw) ??
      `La fecha u hora no es válida en la zona ${v.value.ofTheBusiness}.`
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
      requestToken++
      blocks.value.push(outcome.block)
      blocks.value.sort((a, b) => a.startsAt.localeCompare(b.startsAt))
      isCreateBlockOpen.value = false
      createBlockStatus.value = 'idle'
      toast.success('Bloqueo agregado', {
        detail: `El bloqueo ya aparece en el calendario ${v.value.ofTheProfessional}.`,
      })
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
    requestToken++
    blocks.value = blocks.value.filter((b) => b.id !== block.id)
    if (outcome.kind === 'success') toast.success('Bloqueo retirado')
    return
  }

  // No hay formulario donde mostrar el fallo: el aviso ofrece reintentar.
  toast.error('No pudimos retirar el bloqueo', {
    detail: 'Revisa tu conexión e inténtalo de nuevo.',
    action: { label: 'Reintentar', icon: 'retry', run: () => void onDeleteBlock(block) },
  })
}

function displayInstant(instant: string): string {
  if (!barbershopTimezone.value) return instant
  const { date, time } = formatInstantInTimezone(instant, barbershopTimezone.value)
  return `${displayDate(date)} · ${time}`
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
      requestToken++
      series.value.push(outcome.series)
      isCreateSeriesOpen.value = false
      createSeriesStatus.value = 'idle'
      toast.success('Serie agregada', { detail: 'La serie semanal ya está en el calendario.' })
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
    requestToken++
    series.value = series.value.filter((s) => s.id !== item.id)
    if (outcome.kind === 'success') toast.success('Serie retirada')
    return
  }

  toast.error('No pudimos retirar la serie', {
    detail: 'Revisa tu conexión e inténtalo de nuevo.',
    action: { label: 'Reintentar', icon: 'retry', run: () => void onDeleteSeries(item) },
  })
}

// --- Series por fechas, edición y excepciones (#100) ---------------------

const isDateListOpen = ref(false)
const editingSeries = ref<TimeBlockSeries | null>(null)
const isEditOpen = ref(false)
const managingSeriesId = ref<string | null>(null)
const isInstancesOpen = ref(false)
const managingSeries = computed(
  () => series.value.find((item) => item.id === managingSeriesId.value) ?? null,
)

function onDateListCreated(created: TimeBlockSeries) {
  requestToken++
  series.value.push(created)
}

function openEditSeries(item: TimeBlockSeries) {
  editingSeries.value = item
  isEditOpen.value = true
}

function onSeriesUpdated(payload: {
  originalId: string
  series: TimeBlockSeries
  splitFrom: string | null
}) {
  requestToken++
  if (payload.splitFrom === null) {
    series.value = series.value.map((item) =>
      item.id === payload.originalId ? payload.series : item,
    )
    return
  }
  // «Esta y las siguientes»: la API recorta la original el día anterior al
  // corte y responde la serie nueva; la lista refleja las dos.
  const until = shiftCivilDate(payload.splitFrom, -1)
  series.value = series.value
    .map((item) => (item.id === payload.originalId ? { ...item, effectiveUntil: until } : item))
    .concat(payload.series)
}

function openInstances(item: TimeBlockSeries) {
  managingSeriesId.value = item.id
  isInstancesOpen.value = true
}

function onSeriesChanged(next: TimeBlockSeries) {
  requestToken++
  series.value = series.value.map((item) => (item.id === next.id ? next : item))
}

function seriesName(item: TimeBlockSeries): string {
  return `${blockTypeLabel(item.blockType)} ${item.isoWeekday ? weekdayLabel(item.isoWeekday) : 'por fechas'}`
}

function seriesCounts(item: TimeBlockSeries): string {
  const parts: string[] = []
  if (item.recurrenceKind === 'date_list') {
    parts.push(item.dates.length === 1 ? '1 fecha' : `${item.dates.length} fechas`)
  }
  if (item.exceptions.length > 0) {
    parts.push(
      item.exceptions.length === 1 ? '1 excepción' : `${item.exceptions.length} excepciones`,
    )
  }
  return parts.join(' · ')
}
</script>

<template>
  <section class="blocks-panel" :aria-labelledby="`blocks-title-${barberId}`">
    <header class="blocks-panel__header">
      <div class="blocks-panel__identity">
        <BarberAvatar :full-name="barberName" :photo-url="photoUrl" size="row" />
        <div>
          <p class="blocks-panel__eyebrow">Disponibilidad · {{ barberName }}</p>
          <h2 :id="`blocks-title-${barberId}`">Tiempo fuera de agenda</h2>
        </div>
      </div>
      <div class="blocks-panel__actions">
        <BaseButton
          variant="primary"
          :disabled="!barbershopTimezone || blocksStatus !== 'ready'"
          @click="openCreateBlockDialog"
          >Agregar bloqueo</BaseButton
        >
        <BaseButton
          variant="secondary"
          :disabled="!barbershopTimezone || seriesStatus !== 'ready'"
          @click="openCreateSeriesDialog"
          >Agregar serie semanal</BaseButton
        >
        <BaseButton
          variant="secondary"
          :disabled="!barbershopTimezone || seriesStatus !== 'ready'"
          @click="isDateListOpen = true"
          >Agregar bloqueo por fechas</BaseButton
        >
      </div>
    </header>
    <p class="blocks-panel__intro">
      Un descanso, un día libre o un imprevisto. Reserva ese tiempo para {{ barberName }}.
    </p>
    <p v-if="barbershopTimezone" class="blocks-panel__timezone">
      Horas {{ v.ofTheBusiness }} · {{ barbershopTimezone }}
    </p>
    <p v-else-if="timezoneLoading" role="status">Consultando la zona horaria…</p>
    <BaseAlert v-else variant="warning" title="No pudimos consultar la zona horaria">
      Reintenta antes de crear un bloqueo. Tus registros siguen disponibles.
      <template #action
        ><BaseButton variant="secondary" @click="loadTimezone"
          >Reintentar zona horaria</BaseButton
        ></template
      >
    </BaseAlert>

    <div class="blocks-panel__ledger">
      <section aria-labelledby="single-blocks-title">
        <div class="blocks-panel__section-head">
          <span aria-hidden="true">01</span>
          <h3 id="single-blocks-title">Bloqueos puntuales</h3>
        </div>
        <div v-if="blocksStatus === 'loading'" class="blocks-panel__state" role="status">
          <BaseSpinner size="sm" />
          <p>Cargando bloqueos…</p>
        </div>
        <BaseAlert
          v-else-if="blocksStatus === 'error'"
          variant="warning"
          title="No pudimos cargar los bloqueos"
        >
          Revisa tu conexión e inténtalo de nuevo.
          <template #action
            ><BaseButton variant="secondary" @click="loadLists"
              >Reintentar bloqueos</BaseButton
            ></template
          >
        </BaseAlert>
        <template v-else-if="blocksStatus === 'ready'">
          <div v-if="visibleBlocks.length === 0" class="blocks-panel__empty">
            <span aria-hidden="true">◇</span>
            <p>Sin bloqueos puntuales</p>
            <small>Cuando necesite una pausa, aparecerá aquí.</small>
          </div>
          <TransitionGroup
            v-else
            name="block-record"
            tag="ul"
            class="blocks-panel__list"
            aria-label="Bloqueos puntuales"
          >
            <li
              v-for="block in visibleBlocks"
              :key="block.id"
              class="blocks-panel__record"
              :class="{ 'blocks-panel__record--past': isPast(block) }"
            >
              <div class="blocks-panel__record-top">
                <strong>{{ blockTypeLabel(block.blockType) }}</strong
                ><span class="blocks-panel__stamp">{{ isPast(block) ? 'Pasado' : 'Puntual' }}</span>
              </div>
              <p class="blocks-panel__interval">
                <time :datetime="block.startsAt">{{ displayInstant(block.startsAt) }}</time
                ><span aria-hidden="true"> → </span><span class="sr-only"> hasta </span
                ><time :datetime="block.endsAt">{{ displayInstant(block.endsAt) }}</time>
              </p>
              <p v-if="block.reason" class="blocks-panel__reason">{{ block.reason }}</p>
              <div class="blocks-panel__record-footer">
                <span>{{
                  block.source === 'holiday_calendar' ? 'Calendario de festivos' : 'Bloqueo manual'
                }}</span
                ><BaseButton
                  variant="secondary"
                  :aria-label="`Retirar bloqueo ${blockTypeLabel(block.blockType)} del ${displayInstant(block.startsAt)}`"
                  :loading="pendingDeleteBlockIds.has(block.id)"
                  :disabled="pendingDeleteBlockIds.has(block.id)"
                  @click="onDeleteBlock(block)"
                  >Retirar</BaseButton
                >
              </div>
            </li>
          </TransitionGroup>
          <BaseAlert v-if="moreBlocksError" variant="warning"
            >No pudimos cargar más bloqueos. Puedes reintentar.</BaseAlert
          >
          <BaseButton
            v-if="blocksCursor"
            variant="secondary"
            :loading="loadingMoreBlocks"
            :disabled="loadingMoreBlocks"
            @click="loadMore('blocks')"
            >Cargar más bloqueos</BaseButton
          >
        </template>
      </section>
      <section aria-labelledby="weekly-blocks-title">
        <div class="blocks-panel__section-head">
          <span aria-hidden="true">02</span>
          <h3 id="weekly-blocks-title">Series de bloqueo</h3>
        </div>
        <div v-if="seriesStatus === 'loading'" class="blocks-panel__state" role="status">
          <BaseSpinner size="sm" />
          <p>Cargando series…</p>
        </div>
        <BaseAlert
          v-else-if="seriesStatus === 'error'"
          variant="warning"
          title="No pudimos cargar las series"
        >
          Revisa tu conexión e inténtalo de nuevo.
          <template #action
            ><BaseButton variant="secondary" @click="loadLists"
              >Reintentar series</BaseButton
            ></template
          >
        </BaseAlert>
        <template v-else-if="seriesStatus === 'ready'">
          <div v-if="activeSeries.length === 0" class="blocks-panel__empty">
            <span aria-hidden="true">↻</span>
            <p>Sin pausas recurrentes</p>
            <small>Una misma pausa, cada semana.</small>
          </div>
          <TransitionGroup
            v-else
            name="block-record"
            tag="ul"
            class="blocks-panel__list"
            aria-label="Series de bloqueo"
          >
            <li v-for="item in activeSeries" :key="item.id" class="blocks-panel__record">
              <div class="blocks-panel__record-top">
                <strong>{{ blockTypeLabel(item.blockType) }}</strong
                ><span class="blocks-panel__stamp">{{
                  item.recurrenceKind === 'weekly' ? 'Semanal' : 'Fechas'
                }}</span>
              </div>
              <p class="blocks-panel__interval">
                <template v-if="item.recurrenceKind === 'weekly' && item.isoWeekday"
                  >{{ weekdayLabel(item.isoWeekday) }} · {{ item.startsTime }} ·
                  {{ item.durationMinutes }} min</template
                ><template v-else
                  >Lista de fechas · {{ item.startsTime }} ·
                  {{ item.durationMinutes }} min</template
                >
              </p>
              <p class="blocks-panel__validity">
                Desde {{ displayDate(item.effectiveFrom)
                }}<template v-if="item.effectiveUntil">
                  hasta {{ displayDate(item.effectiveUntil) }}</template
                ><template v-else> · Sin fecha de fin</template>
              </p>
              <p v-if="item.reason" class="blocks-panel__reason">{{ item.reason }}</p>
              <p v-if="seriesCounts(item)" class="blocks-panel__validity">
                {{ seriesCounts(item) }}
              </p>
              <div class="blocks-panel__record-footer">
                <span>{{
                  item.recurrenceKind === 'weekly' ? 'Se repite cada semana' : 'Fechas explícitas'
                }}</span>
                <div class="blocks-panel__record-actions">
                  <BaseButton
                    variant="secondary"
                    :aria-label="`${item.recurrenceKind === 'date_list' ? 'Fechas y excepciones' : 'Excepciones'} de la serie ${seriesName(item)}`"
                    @click="openInstances(item)"
                    >{{
                      item.recurrenceKind === 'date_list' ? 'Fechas' : 'Excepciones'
                    }}</BaseButton
                  >
                  <BaseButton
                    variant="secondary"
                    :aria-label="`Editar serie ${seriesName(item)}`"
                    @click="openEditSeries(item)"
                    >Editar</BaseButton
                  >
                  <BaseButton
                    variant="secondary"
                    :aria-label="`Retirar serie ${seriesName(item)}`"
                    :loading="pendingDeleteSeriesIds.has(item.id)"
                    :disabled="pendingDeleteSeriesIds.has(item.id)"
                    @click="onDeleteSeries(item)"
                    >Retirar</BaseButton
                  >
                </div>
              </div>
            </li>
          </TransitionGroup>
          <BaseAlert v-if="moreSeriesError" variant="warning"
            >No pudimos cargar más series. Puedes reintentar.</BaseAlert
          >
          <BaseButton
            v-if="seriesCursor"
            variant="secondary"
            :loading="loadingMoreSeries"
            :disabled="loadingMoreSeries"
            @click="loadMore('series')"
            >Cargar más series</BaseButton
          >
        </template>
      </section>
    </div>
    <p class="blocks-panel__note">Un bloqueo no cancela ni reprograma los turnos ya agendados.</p>

    <BaseDialog
      v-model="isCreateBlockOpen"
      title="Agregar bloqueo puntual"
      description="Define el tiempo que no estará disponible en la agenda."
      size="md"
      content-class="block-dialog"
      :close-on-backdrop="createBlockStatus !== 'saving'"
      :close-on-escape="createBlockStatus !== 'saving'"
      :show-close="createBlockStatus !== 'saving'"
      @close="onCreateBlockDialogClosed"
    >
      <form
        id="create-block-form"
        :aria-label="`Bloqueo puntual para ${barberName}`"
        class="blocks-page__form block-form"
        @submit.prevent="onSubmitCreateBlock"
      >
        <BaseSelect
          id="create-block-type"
          v-model="createBlockType"
          label="Tipo de bloqueo"
          :options="BLOCK_TYPES"
          :disabled="createBlockStatus === 'saving'"
        />
        <fieldset :disabled="createBlockStatus === 'saving'">
          <legend>01 · Inicio</legend>
          <div class="block-form__row">
            <div class="block-form__date-field">
              <label id="block-start-date-label" for="block-start-date"
                >Fecha de inicio <span aria-hidden="true">*</span></label
              >
              <BaseDatePicker
                v-model="createStartDate"
                trigger-id="block-start-date"
                label-id="block-start-date-label"
                :today="todayCivilDate"
                :disabled="createBlockStatus === 'saving'"
                required
              />
            </div>
            <BaseTimePicker
              v-model="createStartTime"
              label="Hora de inicio"
              :disabled="createBlockStatus === 'saving'"
              required
            />
          </div>
        </fieldset>
        <fieldset :disabled="createBlockStatus === 'saving'">
          <legend>02 · Fin</legend>
          <div class="block-form__row">
            <div class="block-form__date-field">
              <label id="block-end-date-label" for="block-end-date"
                >Fecha de fin <span aria-hidden="true">*</span></label
              >
              <BaseDatePicker
                v-model="createEndDate"
                trigger-id="block-end-date"
                label-id="block-end-date-label"
                :today="todayCivilDate"
                :disabled="createBlockStatus === 'saving'"
                required
              />
            </div>
            <BaseTimePicker
              v-model="createEndTime"
              label="Hora de fin"
              :disabled="createBlockStatus === 'saving'"
              required
            />
          </div>
        </fieldset>
        <BaseInput
          v-model="createBlockReason"
          type="text"
          label="Motivo (opcional)"
          :disabled="createBlockStatus === 'saving'"
        />
        <div class="block-form__preview" aria-live="polite">
          <span class="blocks-panel__eyebrow">Así queda el bloqueo</span>
          <p>{{ blockTypeLabel(createBlockType) }} · {{ blockPreview }}</p>
          <small>Horas {{ v.ofTheBusiness }} · {{ barbershopTimezone }}</small>
        </div>
        <BaseAlert v-if="createBlockError" variant="danger">{{ createBlockError }}</BaseAlert>
        <BaseAlert v-else-if="createBlockStatus === 'validation-error'" variant="danger"
          >Revisa los datos del bloqueo.</BaseAlert
        >
        <BaseAlert v-else-if="createBlockStatus === 'idempotency-conflict'" variant="danger"
          >Este intento ya se envió con otros datos. Cierra y vuelve a abrir el formulario para
          iniciar un bloqueo nuevo.</BaseAlert
        >
        <BaseAlert v-else-if="createBlockStatus === 'not-found'" variant="danger"
          >{{ v.ThisProfessional }} ya no está disponible.</BaseAlert
        >
        <BaseAlert
          v-else-if="
            createBlockStatus === 'network-error' || createBlockStatus === 'unexpected-error'
          "
          variant="danger"
          >No pudimos guardar el bloqueo. Tus datos siguen aquí; inténtalo de nuevo.</BaseAlert
        >
      </form>
      <template #footer
        ><div class="block-form__footer">
          <BaseButton
            variant="secondary"
            :disabled="createBlockStatus === 'saving'"
            @click="isCreateBlockOpen = false"
            >Cancelar</BaseButton
          ><BaseButton
            type="submit"
            form="create-block-form"
            variant="primary"
            :loading="createBlockStatus === 'saving'"
            :disabled="createBlockStatus === 'saving'"
            >Guardar bloqueo</BaseButton
          >
        </div></template
      >
    </BaseDialog>

    <BaseDialog
      v-model="isCreateSeriesOpen"
      title="Agregar serie semanal"
      :description="`Una pausa recurrente para ${barberName}.`"
      size="md"
      content-class="block-dialog"
      :close-on-backdrop="createSeriesStatus !== 'saving'"
      :close-on-escape="createSeriesStatus !== 'saving'"
      :show-close="createSeriesStatus !== 'saving'"
      @close="onCreateSeriesDialogClosed"
    >
      <form
        id="create-series-form"
        class="blocks-page__form block-form"
        @submit.prevent="onSubmitCreateSeries"
      >
        <div class="block-form__identity">
          <BarberAvatar :full-name="barberName" :photo-url="photoUrl" size="row" />
          <div>
            <span class="blocks-panel__eyebrow">Serie semanal</span
            ><strong>{{ barberName }}</strong>
          </div>
        </div>
        <BaseSelect
          id="create-series-type"
          v-model="createSeriesType"
          label="Tipo de bloqueo"
          :options="BLOCK_TYPES"
          :disabled="createSeriesStatus === 'saving'"
        />
        <BaseSelect
          id="create-series-weekday"
          v-model="createSeriesWeekday"
          label="Día de la semana"
          :options="ISO_WEEKDAYS"
          :disabled="createSeriesStatus === 'saving'"
        />
        <div class="block-form__row">
          <BaseTimePicker
            v-model="createSeriesStartsTime"
            label="Hora de inicio"
            required
            :disabled="createSeriesStatus === 'saving'"
          /><BaseInput
            v-model.number="createSeriesDuration"
            type="number"
            label="Duración (minutos)"
            required
            :disabled="createSeriesStatus === 'saving'"
          />
        </div>
        <div class="block-form__date-field">
          <label id="series-from-label" for="series-from"
            >Vigente desde <span aria-hidden="true">*</span></label
          >
          <BaseDatePicker
            v-model="createSeriesEffectiveFrom"
            trigger-id="series-from"
            label-id="series-from-label"
            :today="todayCivilDate"
            :disabled="createSeriesStatus === 'saving'"
            required
          />
        </div>
        <BaseInput
          v-model="createSeriesReason"
          type="text"
          label="Motivo (opcional)"
          :disabled="createSeriesStatus === 'saving'"
        />
        <div class="block-form__preview" aria-live="polite">
          <span class="blocks-panel__eyebrow">Cada semana</span>
          <p>{{ blockTypeLabel(createSeriesType) }} · {{ seriesPreview }}</p>
          <small>{{ barbershopTimezone }}</small>
        </div>
        <BaseAlert v-if="createSeriesError" variant="danger">{{ createSeriesError }}</BaseAlert>
        <BaseAlert v-else-if="createSeriesStatus === 'validation-error'" variant="danger"
          >Revisa los datos de la serie.</BaseAlert
        >
        <BaseAlert v-else-if="createSeriesStatus === 'idempotency-conflict'" variant="danger"
          >Este intento ya se envió con otros datos. Cierra y vuelve a abrir el formulario para
          iniciar una serie nueva.</BaseAlert
        >
        <BaseAlert v-else-if="createSeriesStatus === 'not-found'" variant="danger"
          >{{ v.ThisProfessional }} ya no está disponible.</BaseAlert
        >
        <BaseAlert
          v-else-if="
            createSeriesStatus === 'network-error' || createSeriesStatus === 'unexpected-error'
          "
          variant="danger"
          >No pudimos guardar la serie. Tus datos siguen aquí; inténtalo de nuevo.</BaseAlert
        >
      </form>
      <template #footer
        ><div class="block-form__footer">
          <BaseButton
            variant="secondary"
            :disabled="createSeriesStatus === 'saving'"
            @click="isCreateSeriesOpen = false"
            >Cancelar</BaseButton
          ><BaseButton
            type="submit"
            form="create-series-form"
            variant="primary"
            :loading="createSeriesStatus === 'saving'"
            :disabled="createSeriesStatus === 'saving'"
            >Guardar serie</BaseButton
          >
        </div></template
      >
    </BaseDialog>

    <SeriesFormDialog
      v-model="isDateListOpen"
      mode="create-dates"
      :barber-id="barberId"
      :barber-name="barberName"
      :today="todayCivilDate"
      @created="onDateListCreated"
    />
    <SeriesFormDialog
      v-model="isEditOpen"
      mode="edit"
      :barber-id="barberId"
      :barber-name="barberName"
      :series="editingSeries"
      :today="todayCivilDate"
      @updated="onSeriesUpdated"
    />
    <SeriesInstancesDialog
      v-model="isInstancesOpen"
      :barber-id="barberId"
      :series="managingSeries"
      :today="todayCivilDate"
      @update:series="onSeriesChanged"
    />
  </section>
</template>

<style scoped>
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
  border: 0;
}
.blocks-panel__record :deep(.base-button--secondary) {
  background: transparent;
}
.blocks-panel__record--past :deep(.base-button--secondary) {
  color: var(--color-on-strong);
  border-color: var(--color-brand-accent-surface);
}

.blocks-panel {
  --color-text-primary: var(--color-on-strong);
  --color-text-secondary: var(--color-on-strong-muted);
  --color-accent-brass: var(--color-brand-accent-surface);
  color: var(--color-on-strong);
  padding: 28px;
  background: var(--color-surface-strong);
  border: 1px solid var(--color-field-strong-border);
  border-top: 2px solid var(--color-brand-accent-surface);
}
.blocks-panel__header,
.blocks-panel__identity,
.blocks-panel__actions {
  display: flex;
  align-items: center;
  gap: 16px;
}
.blocks-panel__header {
  justify-content: space-between;
  flex-wrap: wrap;
}
.blocks-panel__identity {
  min-width: 0;
}
.blocks-panel__identity > div {
  min-width: 0;
}
.blocks-panel h2 {
  margin: 5px 0 0;
  font-family: var(--font-display);
  font-size: clamp(24px, 3vw, 32px);
  font-weight: 400;
  line-height: 1.2;
}
.blocks-panel__eyebrow {
  margin: 0;
  color: var(--color-brand-accent-surface);
  font-family: var(--font-sans);
  font-size: var(--font-size-caption);
  letter-spacing: 0.12em;
  text-transform: uppercase;
  overflow-wrap: anywhere;
}
.blocks-panel__intro {
  margin: 20px 0 8px;
  font-size: var(--font-size-body-sm);
  color: var(--color-on-strong-muted);
}
.blocks-panel__timezone {
  margin: 0 0 24px;
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
}
.blocks-panel__ledger {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 28px;
}
.blocks-panel__ledger > section {
  min-width: 0;
}
.blocks-panel__section-head {
  display: flex;
  align-items: baseline;
  gap: 12px;
  border-bottom: 1px solid var(--color-field-strong-border);
  padding-bottom: 14px;
  margin-bottom: 16px;
}
.blocks-panel__section-head > span {
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  letter-spacing: 0.1em;
}
.blocks-panel h3 {
  margin: 0;
  font-family: var(--font-display);
  font-weight: 400;
  font-size: var(--font-size-title-item);
}
.blocks-panel__list {
  list-style: none;
  padding: 0;
  margin: 0 0 16px;
  display: grid;
  gap: 12px;
}
.blocks-panel__record {
  background: var(--color-surface-muted);
  color: var(--color-brand-accent-text);
  border-left: 3px solid var(--color-brand-accent-surface);
  padding: 18px;
  overflow-wrap: anywhere;
  --color-text-primary: var(--color-brand-accent-text);
  --color-text-secondary: var(--color-inactive-text);
  --color-accent-brass: var(--color-focus);
}
.blocks-panel__record--past {
  background: transparent;
  color: var(--color-on-strong-muted);
  border: 1px solid var(--color-field-strong-border);
}
.blocks-panel__record-top {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 12px;
}
.blocks-panel__record-top strong {
  font-family: var(--font-display);
  font-size: var(--font-size-title-item);
  font-weight: 400;
}
.blocks-panel__stamp {
  font-size: var(--font-size-caption);
  text-transform: uppercase;
  letter-spacing: 0.12em;
  border: 1px solid currentColor;
  padding: 3px 6px;
}
.blocks-panel__interval {
  font-size: var(--font-size-body-sm);
  line-height: 1.7;
  margin: 12px 0 6px;
}
.blocks-panel__reason {
  font-size: var(--font-size-body-sm);
  margin: 12px 0;
  line-height: 1.5;
}
.blocks-panel__validity {
  font-size: var(--font-size-caption);
  margin: 6px 0;
}
.blocks-panel__record-footer {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid var(--color-border-subtle);
}
.blocks-panel__record-footer > span {
  font-size: var(--font-size-caption);
}
.blocks-panel__record-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}
.blocks-panel__empty,
.blocks-panel__state {
  min-height: 150px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 18px;
  text-align: center;
  border: 1px dashed var(--color-field-strong-border);
  color: var(--color-on-strong-muted);
}
.blocks-panel__empty > span {
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-title-section);
}
.blocks-panel__empty p {
  margin: 10px 0;
  color: var(--color-on-strong);
  font-family: var(--font-display);
  font-size: var(--font-size-title-item);
}
.blocks-panel__empty small {
  font-size: var(--font-size-caption);
  line-height: 1.6;
}
.blocks-panel__note {
  margin: 24px 0 0;
  padding-top: 16px;
  border-top: 1px solid var(--color-field-strong-border);
  font-size: var(--font-size-caption);
  color: var(--color-on-strong-muted);
  line-height: 1.6;
}
.blocks-panel__actions :deep(.base-button--primary) {
  --color-action-primary-hover: var(--color-brand-accent-surface);
  --color-action-primary-active: var(--color-brand-accent-surface);
  background: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  border-color: var(--color-brand-accent-surface);
}
.blocks-panel__actions :deep(.base-button--secondary),
.blocks-panel__ledger > section > :deep(.base-button--secondary) {
  --color-surface-muted: var(--color-field-strong);
  --color-border-subtle: var(--color-field-strong-border);
  color: var(--color-on-strong);
  border-color: var(--color-brand-accent-surface);
  background: transparent;
}
.block-form {
  display: grid;
  gap: 22px;
}
.block-form__identity {
  display: flex;
  gap: 14px;
  align-items: center;
  padding-bottom: 18px;
  border-bottom: 1px solid var(--color-field-strong-border);
}
.block-form__identity strong {
  display: block;
  font-family: var(--font-display);
  font-size: var(--font-size-body-lg);
  font-weight: 400;
  margin-top: 5px;
  overflow-wrap: anywhere;
}
.block-form fieldset {
  min-width: 0;
  padding: 0;
  border: 0;
  margin: 0;
}
.block-form legend {
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  text-transform: uppercase;
  letter-spacing: 0.1em;
  margin-bottom: 14px;
  display: flex;
  align-items: center;
  gap: 12px;
}
.block-form__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.block-form__date-field {
  display: grid;
  align-content: start;
  gap: 8px;
  min-width: 0;
}
.block-form__date-field > label {
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  font-weight: 600;
  line-height: 14px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
.block-form__date-field > label span {
  color: var(--color-danger-on-strong);
}
.block-form__date-field :deep(.agenda-date-picker__trigger) {
  min-height: 44px;
}
.block-form__row > * {
  min-width: 0;
}
.block-form__preview {
  border-left: 2px solid var(--color-brand-accent-surface);
  background: var(--color-field-strong);
  padding: 14px 16px;
}
.block-form__preview p {
  font-size: var(--font-size-body-sm);
  line-height: 1.7;
  margin: 8px 0;
}
.block-form__preview small {
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
}
.block-form__footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
.block-record-enter-active,
.block-record-leave-active,
.block-record-move {
  transition:
    opacity 220ms ease,
    transform 220ms ease;
}
.block-record-enter-from,
.block-record-leave-to {
  opacity: 0;
  transform: translateY(8px);
}
@media (max-width: 767px) {
  .blocks-panel {
    padding: 20px 16px;
  }
  .blocks-panel__ledger {
    grid-template-columns: 1fr;
    gap: 24px;
  }
  .blocks-panel__actions {
    width: 100%;
    flex-wrap: wrap;
    gap: 10px;
  }
  .blocks-panel__actions :deep(.base-button) {
    flex: 1 1 auto;
  }
}
@media (max-width: 380px) {
  .block-form__row {
    grid-template-columns: 1fr;
    gap: 14px;
  }
  .blocks-panel__record {
    padding: 14px;
  }
  .blocks-panel__identity {
    align-items: flex-start;
  }
  .block-form__footer :deep(.base-button) {
    flex: 1;
  }
}
@media (prefers-reduced-motion: reduce) {
  .block-record-enter-active,
  .block-record-leave-active,
  .block-record-move {
    transition: none;
  }
}
</style>

<style>
/* BaseDialog se teleporta al body: este ámbito solo gobierna el panel de HU-042. */
.block-dialog {
  --color-text-primary: var(--color-on-strong);
  --color-text-secondary: var(--color-on-strong-muted);
  --color-accent-brass: var(--color-brand-accent-surface);
  --color-border-control: var(--color-field-strong-border);
  --color-surface: var(--color-field-strong);
  --color-focus: var(--color-brand-accent-surface);
  background: var(--color-surface-strong);
  color: var(--color-on-strong);
  border: 1px solid var(--color-field-strong-border);
  border-top: 2px solid var(--color-brand-accent-surface);
  box-shadow: var(--shadow-dialog);
}
.block-dialog .base-dialog__title {
  font-family: var(--font-display);
  font-size: var(--font-size-title-section);
  font-weight: 400;
  line-height: 1.2;
}
.block-dialog .base-dialog__header,
.block-dialog .base-dialog__footer {
  border-color: var(--color-field-strong-border);
}
.block-dialog .base-input {
  --input-bg: var(--color-field-strong);
  --input-border-color: var(--color-field-strong-border);
  --input-border-base-color: var(--color-brand-accent-surface);
  --input-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);
  color-scheme: dark;
}
.block-dialog .base-input__label {
  color: var(--color-on-strong-muted);
}
.block-dialog .base-button--primary {
  --color-action-primary-hover: var(--color-brand-accent-surface);
  --color-action-primary-active: var(--color-brand-accent-surface);
  background: var(--color-brand-accent-surface);
  border-color: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
}
.block-dialog .base-button--secondary {
  background: transparent;
  border-color: var(--color-brand-accent-surface);
  color: var(--color-on-strong);
}
/* El hover/active base del botón secundario usa superficies claras: sobre la
   tinta del diálogo dejaban texto claro sobre fondo claro (contraste 1,13). */
.block-dialog .base-button--secondary:hover:not(:disabled):not(.base-button--loading),
.block-dialog .base-button--secondary:active:not(:disabled):not(.base-button--loading) {
  background: var(--color-field-strong);
  border-color: var(--color-brand-accent-surface);
}
.block-dialog .base-dialog__close {
  color: var(--color-on-strong);
}
</style>
