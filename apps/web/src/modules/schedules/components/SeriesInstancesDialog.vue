<script setup lang="ts">
// HU-042 / #100: fechas explícitas de una serie `date_list` y excepciones
// («esta instancia no») de cualquier serie. La API responde 204 sin cuerpo,
// así que el diálogo aplica el cambio sobre una copia de la serie y se la
// devuelve a BarberBlocksPanel, dueño de la lista.
import { computed, ref, useId, watch } from 'vue'
import { useToast } from '@/shared/composables'
import { BaseAlert, BaseButton, BaseDatePicker, BaseDialog, BaseInput } from '@/shared/ui'
import {
  addSeriesDate,
  addSeriesException,
  removeSeriesDate,
  removeSeriesException,
} from '../api/timeBlocksApi'
import type { SeriesChildOutcome } from '../model/blockOutcome'
import { displayCivilDate } from '../model/displayDate'
import { blockTypeLabel, type TimeBlockSeries } from '../model/timeBlock'
import { weekdayLabel } from '../model/workingHour'
import { validateEffectiveFrom, validateReason } from '../validation/timeBlockValidation'

const props = defineProps<{
  modelValue: boolean
  barberId: string
  series: TimeBlockSeries | null
  today: string | null
}>()
const emit = defineEmits<{
  'update:modelValue': [open: boolean]
  'update:series': [series: TimeBlockSeries]
}>()

const uid = useId()
const toast = useToast()
const isDateList = computed(() => props.series?.recurrenceKind === 'date_list')
const sortedDates = computed(() =>
  [...(props.series?.dates ?? [])].sort((a, b) => a.blockDate.localeCompare(b.blockDate)),
)
const sortedExceptions = computed(() =>
  [...(props.series?.exceptions ?? [])].sort((a, b) =>
    a.excludedDate.localeCompare(b.excludedDate),
  ),
)

const newDate = ref('')
const exceptionDate = ref('')
const exceptionReason = ref('')
const error = ref<string | undefined>(undefined)
const pending = ref<Set<string>>(new Set())

watch(
  () => props.modelValue,
  (open) => {
    if (!open) return
    newDate.value = ''
    exceptionDate.value = ''
    exceptionReason.value = ''
    error.value = undefined
  },
)

const heading = computed(() => {
  const current = props.series
  if (!current) return ''
  const kind = blockTypeLabel(current.blockType)
  return current.recurrenceKind === 'weekly' && current.isoWeekday
    ? `${kind} · ${weekdayLabel(current.isoWeekday)} · ${current.startsTime}`
    : `${kind} · ${current.startsTime}`
})

function track(key: string, on: boolean) {
  const next = new Set(pending.value)
  if (on) next.add(key)
  else next.delete(key)
  pending.value = next
}

function describe(outcome: SeriesChildOutcome, what: 'fecha' | 'excepción'): string | undefined {
  switch (outcome.kind) {
    case 'success':
      return undefined
    case 'duplicate':
      return what === 'fecha'
        ? 'Esa fecha ya está en la serie.'
        : 'Esa instancia ya tiene una excepción.'
    case 'validation-error':
      return what === 'fecha'
        ? 'La fecha debe quedar dentro de la vigencia de la serie.'
        : 'Revisa la fecha y el motivo de la excepción.'
    case 'not-found':
      return 'Esta serie ya no está disponible. Cierra el diálogo y actualiza la lista.'
    default:
      return 'No pudimos guardar el cambio. Revisa tu conexión e inténtalo de nuevo.'
  }
}

async function onAddDate() {
  const current = props.series
  if (!current || pending.value.has('add-date')) return
  error.value = undefined
  const invalid = validateEffectiveFrom(newDate.value)
  if (invalid) {
    error.value = 'Elige la fecha que quieres agregar.'
    return
  }
  const date = newDate.value
  track('add-date', true)
  const outcome = await addSeriesDate(props.barberId, current.id, date)
  track('add-date', false)
  const problem = describe(outcome, 'fecha')
  if (problem) {
    error.value = problem
    return
  }
  emit('update:series', { ...current, dates: [...current.dates, { blockDate: date }] })
  newDate.value = ''
  toast.success('Fecha agregada', { detail: displayCivilDate(date) })
}

async function onRemoveDate(date: string) {
  const current = props.series
  const key = `date-${date}`
  if (!current || pending.value.has(key)) return
  error.value = undefined
  track(key, true)
  const outcome = await removeSeriesDate(props.barberId, current.id, date)
  track(key, false)
  // not-found: ya no existía; la pantalla coincide con el servidor al quitarla.
  if (outcome.kind !== 'success' && outcome.kind !== 'not-found') {
    error.value = describe(outcome, 'fecha')
    return
  }
  emit('update:series', {
    ...current,
    dates: current.dates.filter((item) => item.blockDate !== date),
  })
  if (outcome.kind === 'success')
    toast.success('Fecha retirada', { detail: displayCivilDate(date) })
}

async function onAddException() {
  const current = props.series
  if (!current || pending.value.has('add-exception')) return
  error.value = undefined
  const invalid =
    validateEffectiveFrom(exceptionDate.value) === undefined
      ? validateReason(exceptionReason.value)
      : 'Elige la fecha de la instancia que no se aplica.'
  if (invalid) {
    error.value = invalid
    return
  }
  const date = exceptionDate.value
  const reason = exceptionReason.value.trim() === '' ? null : exceptionReason.value.trim()
  track('add-exception', true)
  const outcome = await addSeriesException(props.barberId, current.id, date, reason)
  track('add-exception', false)
  const problem = describe(outcome, 'excepción')
  if (problem) {
    error.value = problem
    return
  }
  emit('update:series', {
    ...current,
    exceptions: [
      ...current.exceptions,
      { excludedDate: date, reason, createdAt: new Date().toISOString() },
    ],
  })
  exceptionDate.value = ''
  exceptionReason.value = ''
  toast.success('Excepción agregada', { detail: `El ${displayCivilDate(date)} no se bloquea.` })
}

async function onRemoveException(date: string) {
  const current = props.series
  const key = `exception-${date}`
  if (!current || pending.value.has(key)) return
  error.value = undefined
  track(key, true)
  const outcome = await removeSeriesException(props.barberId, current.id, date)
  track(key, false)
  if (outcome.kind !== 'success' && outcome.kind !== 'not-found') {
    error.value = describe(outcome, 'excepción')
    return
  }
  emit('update:series', {
    ...current,
    exceptions: current.exceptions.filter((item) => item.excludedDate !== date),
  })
  if (outcome.kind === 'success') {
    toast.success('Instancia restaurada', {
      detail: `El ${displayCivilDate(date)} vuelve a bloquearse.`,
    })
  }
}
</script>

<template>
  <BaseDialog
    :model-value="modelValue && series !== null"
    :title="isDateList ? 'Fechas y excepciones' : 'Excepciones de la serie'"
    :description="heading"
    size="md"
    content-class="block-dialog"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div v-if="series" class="block-form">
      <fieldset v-if="isDateList">
        <legend>Fechas del bloqueo</legend>
        <ul v-if="sortedDates.length" class="block-chips" aria-label="Fechas de la serie">
          <li v-for="item in sortedDates" :key="item.blockDate">
            <span>{{ displayCivilDate(item.blockDate) }}</span>
            <BaseButton
              variant="secondary"
              :aria-label="`Quitar la fecha ${displayCivilDate(item.blockDate)}`"
              :loading="pending.has(`date-${item.blockDate}`)"
              :disabled="pending.has(`date-${item.blockDate}`)"
              @click="onRemoveDate(item.blockDate)"
              >Quitar</BaseButton
            >
          </li>
        </ul>
        <p v-else class="block-chips__empty">Esta serie no tiene fechas todavía.</p>
        <div class="block-form__date-field">
          <label :id="`new-date-label-${uid}`" :for="`new-date-${uid}`">Agregar una fecha</label>
          <BaseDatePicker
            v-model="newDate"
            :trigger-id="`new-date-${uid}`"
            :label-id="`new-date-label-${uid}`"
            :today="today"
            :disabled="pending.has('add-date')"
          />
        </div>
        <BaseButton
          variant="secondary"
          :loading="pending.has('add-date')"
          :disabled="!newDate || pending.has('add-date')"
          @click="onAddDate"
          >Agregar fecha</BaseButton
        >
      </fieldset>

      <fieldset>
        <legend>Excepciones · «esta instancia no»</legend>
        <ul v-if="sortedExceptions.length" class="block-chips" aria-label="Excepciones de la serie">
          <li v-for="item in sortedExceptions" :key="item.excludedDate">
            <span
              >{{ displayCivilDate(item.excludedDate) }}
              <small v-if="item.reason">{{ item.reason }}</small></span
            >
            <BaseButton
              variant="secondary"
              :aria-label="`Restaurar la instancia del ${displayCivilDate(item.excludedDate)}`"
              :loading="pending.has(`exception-${item.excludedDate}`)"
              :disabled="pending.has(`exception-${item.excludedDate}`)"
              @click="onRemoveException(item.excludedDate)"
              >Restaurar</BaseButton
            >
          </li>
        </ul>
        <p v-else class="block-chips__empty">Ninguna instancia está exceptuada.</p>
        <div class="block-form__date-field">
          <label :id="`exception-label-${uid}`" :for="`exception-${uid}`"
            >Instancia que no se bloquea</label
          >
          <BaseDatePicker
            v-model="exceptionDate"
            :trigger-id="`exception-${uid}`"
            :label-id="`exception-label-${uid}`"
            :today="today"
            :disabled="pending.has('add-exception')"
          />
        </div>
        <BaseInput
          v-model="exceptionReason"
          type="text"
          label="Motivo de la excepción (opcional)"
          :disabled="pending.has('add-exception')"
        />
        <BaseButton
          variant="secondary"
          :loading="pending.has('add-exception')"
          :disabled="!exceptionDate || pending.has('add-exception')"
          @click="onAddException"
          >Agregar excepción</BaseButton
        >
      </fieldset>

      <BaseAlert v-if="error" variant="danger">{{ error }}</BaseAlert>
    </div>
    <template #footer>
      <div class="block-form__footer">
        <BaseButton variant="secondary" @click="emit('update:modelValue', false)"
          >Cerrar</BaseButton
        >
      </div>
    </template>
  </BaseDialog>
</template>

<style scoped src="./block-form.css"></style>
