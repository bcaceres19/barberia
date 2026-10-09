<script setup lang="ts">
// HU-042 / #100: alta de una serie por fechas explícitas (`date_list`, RN-BLQ-01
// «bloqueo de varios días») y edición de cualquier serie por alcance. El
// dueño de la lista es BarberBlocksPanel: aquí solo se pide el cambio y se
// devuelve la serie que respondió la API.
import { computed, ref, useId, watch } from 'vue'
import { shiftCivilDate } from '@/shared/time/civilDate'
import { useToast } from '@/shared/composables'
import {
  BaseAlert,
  BaseButton,
  BaseDatePicker,
  BaseDialog,
  BaseInput,
  BaseSelect,
  BaseTimePicker,
} from '@/shared/ui'
import { createDateListSeries, updateTimeBlockSeries } from '../api/timeBlocksApi'
import { displayCivilDate } from '../model/displayDate'
import { newIdempotencyKey } from '../model/idempotencyKey'
import {
  BLOCK_TYPES,
  blockTypeLabel,
  type BlockType,
  type TimeBlockSeries,
} from '../model/timeBlock'
import { weekdayLabel } from '../model/workingHour'
import { validateDurationMinutes } from '../validation/scheduleValidation'
import {
  validateEffectiveFrom,
  validateReason,
  validateStartsTime,
} from '../validation/timeBlockValidation'

type SaveStatus =
  | 'idle'
  | 'saving'
  | 'validation-error'
  | 'idempotency-conflict'
  | 'not-found'
  | 'network-error'
  | 'unexpected-error'
type Scope = 'whole' | 'this_and_following'
type EndMode = 'open' | 'until'

const props = defineProps<{
  modelValue: boolean
  mode: 'create-dates' | 'edit'
  barberId: string
  barberName: string
  series?: TimeBlockSeries | null
  today: string | null
}>()
const emit = defineEmits<{
  'update:modelValue': [open: boolean]
  created: [series: TimeBlockSeries]
  // `splitFrom` es la fecha de corte cuando el alcance fue «esta y las
  // siguientes»: la serie original queda recortada el día anterior y `series`
  // es la nueva.
  updated: [payload: { originalId: string; series: TimeBlockSeries; splitFrom: string | null }]
}>()

const uid = useId()
const toast = useToast()
const isEdit = computed(() => props.mode === 'edit')
const isWeekly = computed(() => props.series?.recurrenceKind === 'weekly')

const blockType = ref<BlockType>('vacation')
const startsTime = ref('00:00')
const duration = ref(1440)
const effectiveFrom = ref('')
const endMode = ref<EndMode>('open')
const effectiveUntil = ref('')
const reason = ref('')
const dates = ref<string[]>([])
const pendingDate = ref('')
const scope = ref<Scope>('whole')
const splitDate = ref('')
const error = ref<string | undefined>(undefined)
const status = ref<SaveStatus>('idle')
let idempotencyKey = newIdempotencyKey()

const END_OPTIONS = [
  { value: 'open' as EndMode, label: 'Sin fecha de fin' },
  { value: 'until' as EndMode, label: 'Hasta una fecha' },
]
const SCOPE_OPTIONS = [
  { value: 'whole' as Scope, label: 'Toda la serie' },
  { value: 'this_and_following' as Scope, label: 'Esta fecha y las siguientes' },
]

const title = computed(() =>
  isEdit.value ? 'Editar serie de bloqueo' : 'Agregar bloqueo por fechas',
)
const description = computed(() =>
  isEdit.value
    ? `Cambia el tipo, la hora o la vigencia de esta serie de ${props.barberName}.`
    : `Una pausa en fechas elegidas, sin repetirse cada semana, para ${props.barberName}.`,
)
const splitting = computed(
  () => isEdit.value && isWeekly.value && scope.value === 'this_and_following',
)
const busy = computed(() => status.value === 'saving')

function reset() {
  error.value = undefined
  status.value = 'idle'
  pendingDate.value = ''
  scope.value = 'whole'
  splitDate.value = ''
  idempotencyKey = newIdempotencyKey()
  const current = props.series
  if (isEdit.value && current) {
    blockType.value = current.blockType
    startsTime.value = current.startsTime
    duration.value = current.durationMinutes
    effectiveFrom.value = current.effectiveFrom
    endMode.value = current.effectiveUntil ? 'until' : 'open'
    effectiveUntil.value = current.effectiveUntil ?? ''
    reason.value = current.reason ?? ''
    dates.value = []
    return
  }
  blockType.value = 'vacation'
  startsTime.value = '00:00'
  duration.value = 1440
  effectiveFrom.value = ''
  endMode.value = 'open'
  effectiveUntil.value = ''
  reason.value = ''
  dates.value = []
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) reset()
  },
  { immediate: true },
)

function addPendingDate() {
  error.value = undefined
  const date = pendingDate.value
  if (!date) return
  if (dates.value.includes(date)) {
    error.value = `El ${displayCivilDate(date)} ya está en la lista.`
    return
  }
  dates.value = [...dates.value, date].sort()
  pendingDate.value = ''
}
function removePendingDate(date: string) {
  dates.value = dates.value.filter((item) => item !== date)
}

const summary = computed(() => {
  const kind = blockTypeLabel(blockType.value)
  const clock = `${startsTime.value || '—'} · ${duration.value || '—'} min`
  if (isEdit.value && isWeekly.value && props.series?.isoWeekday) {
    return `${kind} · ${weekdayLabel(props.series.isoWeekday)} · ${clock}`
  }
  return `${kind} · ${clock}`
})

function validate(): string | undefined {
  const base =
    validateStartsTime(startsTime.value) ??
    validateDurationMinutes(duration.value) ??
    validateReason(reason.value)
  if (base) return base

  const from = splitting.value ? splitDate.value : effectiveFrom.value
  const fromError = validateEffectiveFrom(from)
  if (fromError) return splitting.value ? 'Elige desde qué fecha aplica el cambio.' : fromError
  if (splitting.value && props.series) {
    if (splitDate.value <= props.series.effectiveFrom) {
      return `El corte debe ser posterior al inicio de la serie (${displayCivilDate(props.series.effectiveFrom)}).`
    }
    if (props.series.effectiveUntil && splitDate.value > props.series.effectiveUntil) {
      return `El corte debe caer dentro de la vigencia (hasta ${displayCivilDate(props.series.effectiveUntil)}).`
    }
  }
  if (endMode.value === 'until') {
    const untilError = validateEffectiveFrom(effectiveUntil.value)
    if (untilError) return 'Elige la fecha de fin de la vigencia.'
    if (effectiveUntil.value < from) return 'La vigencia no puede terminar antes de empezar.'
  }
  if (!isEdit.value) {
    if (dates.value.length === 0) return 'Agrega al menos una fecha.'
    const outside = dates.value.find(
      (date) =>
        date < effectiveFrom.value || (endMode.value === 'until' && date > effectiveUntil.value),
    )
    if (outside) return `El ${displayCivilDate(outside)} queda fuera de la vigencia de la serie.`
  }
  return undefined
}

function fail(kind: Exclude<SaveStatus, 'idle' | 'saving'>) {
  status.value = kind
}

async function onSubmit() {
  if (busy.value) return
  error.value = undefined
  const problem = validate()
  if (problem) {
    error.value = problem
    return
  }
  const trimmed = reason.value.trim()
  const common = {
    blockType: blockType.value,
    startsTime: startsTime.value,
    durationMinutes: duration.value,
    effectiveUntil: endMode.value === 'until' ? effectiveUntil.value : null,
    reason: trimmed === '' ? null : trimmed,
  }

  status.value = 'saving'
  if (!isEdit.value) {
    const outcome = await createDateListSeries(
      props.barberId,
      { ...common, effectiveFrom: effectiveFrom.value, explicitDates: dates.value },
      idempotencyKey,
    )
    if (outcome.kind === 'success') {
      status.value = 'idle'
      emit('created', outcome.series)
      emit('update:modelValue', false)
      toast.success('Bloqueo por fechas agregado', {
        detail: `Las ${outcome.series.dates.length} fechas ya están en el calendario.`,
      })
      return
    }
    fail(outcome.kind)
    return
  }

  const current = props.series
  if (!current) return
  const cut = splitting.value ? splitDate.value : null
  const outcome = await updateTimeBlockSeries(
    props.barberId,
    current.id,
    { ...common, effectiveFrom: cut ?? effectiveFrom.value },
    cut ? { effectiveDate: cut } : null,
  )
  if (outcome.kind === 'success') {
    status.value = 'idle'
    emit('updated', { originalId: current.id, series: outcome.series, splitFrom: cut })
    emit('update:modelValue', false)
    toast.success(cut ? 'Serie dividida' : 'Serie actualizada', {
      detail: cut
        ? `El cambio aplica desde el ${displayCivilDate(cut)}; antes se conserva lo anterior.`
        : 'Los cambios ya están en el calendario.',
    })
    return
  }
  fail(outcome.kind)
}

const previousDay = computed(() =>
  splitting.value && splitDate.value ? shiftCivilDate(splitDate.value, -1) : null,
)
</script>

<template>
  <BaseDialog
    :model-value="modelValue"
    :title="title"
    :description="description"
    size="md"
    content-class="block-dialog"
    :close-on-backdrop="!busy"
    :close-on-escape="!busy"
    :show-close="!busy"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <form :id="`series-form-${uid}`" class="block-form" @submit.prevent="onSubmit">
      <BaseSelect
        :id="`series-type-${uid}`"
        v-model="blockType"
        label="Tipo de bloqueo"
        :options="BLOCK_TYPES"
        :disabled="busy"
      />
      <BaseSelect
        v-if="isEdit && isWeekly"
        :id="`series-scope-${uid}`"
        v-model="scope"
        label="Qué cambia"
        :options="SCOPE_OPTIONS"
        :disabled="busy"
      />
      <div v-if="splitting" class="block-form__date-field">
        <label :id="`series-split-label-${uid}`" :for="`series-split-${uid}`"
          >Aplicar desde <span aria-hidden="true">*</span></label
        >
        <BaseDatePicker
          v-model="splitDate"
          :trigger-id="`series-split-${uid}`"
          :label-id="`series-split-label-${uid}`"
          :today="today"
          :disabled="busy"
          required
        />
      </div>
      <div class="block-form__row">
        <BaseTimePicker v-model="startsTime" label="Hora de inicio" required :disabled="busy" />
        <BaseInput
          v-model.number="duration"
          type="number"
          label="Duración (minutos)"
          required
          :disabled="busy"
        />
      </div>
      <div v-if="!splitting" class="block-form__date-field">
        <label :id="`series-from-label-${uid}`" :for="`series-from-${uid}`"
          >Vigente desde <span aria-hidden="true">*</span></label
        >
        <BaseDatePicker
          v-model="effectiveFrom"
          :trigger-id="`series-from-${uid}`"
          :label-id="`series-from-label-${uid}`"
          :today="today"
          :disabled="busy"
          required
        />
      </div>
      <BaseSelect
        :id="`series-end-${uid}`"
        v-model="endMode"
        label="Fin de la vigencia"
        :options="END_OPTIONS"
        :disabled="busy"
      />
      <div v-if="endMode === 'until'" class="block-form__date-field">
        <label :id="`series-until-label-${uid}`" :for="`series-until-${uid}`"
          >Vigente hasta <span aria-hidden="true">*</span></label
        >
        <BaseDatePicker
          v-model="effectiveUntil"
          :trigger-id="`series-until-${uid}`"
          :label-id="`series-until-label-${uid}`"
          :today="today"
          :disabled="busy"
          required
        />
      </div>

      <fieldset v-if="!isEdit" :disabled="busy">
        <legend>Fechas del bloqueo</legend>
        <ul v-if="dates.length" class="block-chips" aria-label="Fechas elegidas">
          <li v-for="date in dates" :key="date">
            <span>{{ displayCivilDate(date) }}</span>
            <BaseButton
              variant="secondary"
              :aria-label="`Quitar la fecha ${displayCivilDate(date)}`"
              @click="removePendingDate(date)"
              >Quitar</BaseButton
            >
          </li>
        </ul>
        <p v-else class="block-chips__empty">Aún no elegiste ninguna fecha.</p>
        <div class="block-form__date-field">
          <label :id="`series-date-label-${uid}`" :for="`series-date-${uid}`"
            >Elegir una fecha</label
          >
          <BaseDatePicker
            v-model="pendingDate"
            :trigger-id="`series-date-${uid}`"
            :label-id="`series-date-label-${uid}`"
            :today="today"
            :disabled="busy"
          />
        </div>
        <BaseButton variant="secondary" :disabled="!pendingDate || busy" @click="addPendingDate"
          >Añadir a la lista</BaseButton
        >
      </fieldset>

      <BaseInput v-model="reason" type="text" label="Motivo (opcional)" :disabled="busy" />

      <div class="block-form__summary" aria-live="polite">
        <span class="block-form__eyebrow">Así queda</span>
        <p>{{ summary }}</p>
        <p v-if="previousDay">
          La serie actual termina el {{ displayCivilDate(previousDay) }} y esta empieza el
          {{ displayCivilDate(splitDate) }}.
        </p>
        <p v-else-if="isEdit && !isWeekly">
          Las fechas y excepciones se gestionan desde «Fechas y excepciones».
        </p>
      </div>

      <BaseAlert v-if="error" variant="danger">{{ error }}</BaseAlert>
      <BaseAlert v-else-if="status === 'validation-error'" variant="danger"
        >Revisa los datos: las fechas deben quedar dentro de la vigencia y la hora dentro del
        día.</BaseAlert
      >
      <BaseAlert v-else-if="status === 'idempotency-conflict'" variant="danger"
        >Este intento ya se envió con otros datos. Cierra y vuelve a abrir el formulario para
        empezar de nuevo.</BaseAlert
      >
      <BaseAlert v-else-if="status === 'not-found'" variant="danger"
        >Esta serie ya no está disponible. Cierra el formulario y actualiza la lista.</BaseAlert
      >
      <BaseAlert
        v-else-if="status === 'network-error' || status === 'unexpected-error'"
        variant="danger"
        >No pudimos guardar los cambios. Tus datos siguen aquí; inténtalo de nuevo.</BaseAlert
      >
    </form>
    <template #footer>
      <div class="block-form__footer">
        <BaseButton variant="secondary" :disabled="busy" @click="emit('update:modelValue', false)"
          >Cancelar</BaseButton
        >
        <BaseButton
          type="submit"
          :form="`series-form-${uid}`"
          variant="primary"
          :loading="busy"
          :disabled="busy"
          >{{ isEdit ? 'Guardar cambios' : 'Guardar bloqueo' }}</BaseButton
        >
      </div>
    </template>
  </BaseDialog>
</template>

<style scoped src="./block-form.css"></style>
<style scoped>
.block-form__eyebrow {
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  letter-spacing: 0.12em;
  text-transform: uppercase;
}
</style>
