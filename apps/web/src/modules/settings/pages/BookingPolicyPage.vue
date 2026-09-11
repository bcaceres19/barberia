<script setup lang="ts">
// Pantalla "Reserva pública" (HU-093): consultar y actualizar anticipación,
// ventana, rejilla y política de cancelación tardía de la barbería activa.
// Estados discriminados: carga inicial, listo, guardando, error de campo,
// conflicto de versión, error recuperable y éxito (mismo criterio que
// SettingsPage.vue/HU-020). Un error recuperable NUNCA borra lo que el
// barbero ya escribió; solo un guardado exitoso confirmado por el servidor
// reemplaza los valores del formulario. `versionToken` viaja aparte del
// formulario (precondición `If-Match` de la siguiente escritura, CA-093-02),
// nunca como un campo editable.
import { onMounted, reactive, ref } from 'vue'
import { BaseAlert, BaseButton, BaseInput } from '@/shared/ui'
import { fetchBookingPolicy, saveBookingPolicy } from '../api/bookingPolicyApi'
import {
  ALLOWED_SLOT_GRID_MINUTES,
  toFormValues,
  type BookingPolicyFormValues,
} from '../model/bookingPolicy'
import {
  validateAdvanceWithinWindow,
  validateCancellationDeadlineMinutes,
  validateLateCancellationCoherence,
  validateMaxAdvanceDays,
  validateMinAdvanceMinutes,
  validateSlotGridMinutes,
} from '../validation/bookingPolicyValidation'

type LoadStatus = 'loading' | 'ready' | 'load-error'
type SaveStatus =
  | 'idle'
  | 'saving'
  | 'saved'
  | 'validation-error'
  | 'version-conflict'
  | 'network-error'
  | 'unexpected-error'

const loadStatus = ref<LoadStatus>('loading')
const saveStatus = ref<SaveStatus>('idle')

const form = reactive<BookingPolicyFormValues>({
  minAdvanceMinutes: 60,
  maxAdvanceDays: 3,
  slotGridMinutes: 15,
  cancellationDeadlineMinutes: 20,
  lateCancellationClientAllowed: true,
  lateCancellationReasonRequired: true,
})
// El token de la última lectura confirmada: precondición de la siguiente
// escritura (CA-093-02), nunca expuesto como campo del formulario.
let currentVersionToken = ''

type FieldErrors = Partial<
  Record<
    | 'minAdvanceMinutes'
    | 'maxAdvanceDays'
    | 'slotGridMinutes'
    | 'cancellationDeadlineMinutes'
    | 'lateCancellation',
    string
  >
>
const fieldErrors = ref<FieldErrors>({})
const attemptedSubmit = ref(false)

// requestToken evita que una carga inicial obsoleta sobrescriba el
// formulario con datos viejos (mismo criterio que SettingsPage.vue).
let requestToken = 0

async function load() {
  const token = ++requestToken
  loadStatus.value = 'loading'
  const outcome = await fetchBookingPolicy()
  if (token !== requestToken) return

  if (outcome.kind === 'success') {
    Object.assign(form, toFormValues(outcome.policy))
    currentVersionToken = outcome.policy.versionToken
    loadStatus.value = 'ready'
    return
  }
  loadStatus.value = 'load-error'
}

onMounted(load)

function runValidation(): FieldErrors {
  const errors: FieldErrors = {}
  const minAdvanceError = validateMinAdvanceMinutes(form.minAdvanceMinutes)
  if (minAdvanceError) errors.minAdvanceMinutes = minAdvanceError
  const maxAdvanceError = validateMaxAdvanceDays(form.maxAdvanceDays)
  if (maxAdvanceError) errors.maxAdvanceDays = maxAdvanceError
  const gridError = validateSlotGridMinutes(form.slotGridMinutes)
  if (gridError) errors.slotGridMinutes = gridError
  const deadlineError = validateCancellationDeadlineMinutes(form.cancellationDeadlineMinutes)
  if (deadlineError) errors.cancellationDeadlineMinutes = deadlineError
  if (!errors.minAdvanceMinutes && !errors.maxAdvanceDays) {
    const windowError = validateAdvanceWithinWindow(form.minAdvanceMinutes, form.maxAdvanceDays)
    if (windowError) errors.minAdvanceMinutes = windowError
  }
  const coherenceError = validateLateCancellationCoherence(
    form.lateCancellationClientAllowed,
    form.lateCancellationReasonRequired,
  )
  if (coherenceError) errors.lateCancellation = coherenceError
  return errors
}

function onNumberFieldInput(
  field: 'minAdvanceMinutes' | 'maxAdvanceDays' | 'cancellationDeadlineMinutes',
  value: string | number,
) {
  form[field] = Number(value)
  if (attemptedSubmit.value) fieldErrors.value = runValidation()
}

function onSlotGridInput(event: Event) {
  form.slotGridMinutes = Number((event.target as HTMLSelectElement).value)
  if (attemptedSubmit.value) fieldErrors.value = runValidation()
}

function onLateCancellationClientAllowedInput(event: Event) {
  form.lateCancellationClientAllowed = (event.target as HTMLInputElement).checked
  if (attemptedSubmit.value) fieldErrors.value = runValidation()
}

function onLateCancellationReasonRequiredInput(event: Event) {
  form.lateCancellationReasonRequired = (event.target as HTMLInputElement).checked
  if (attemptedSubmit.value) fieldErrors.value = runValidation()
}

async function onSubmit() {
  // Guardia de doble envío (mismo patrón que SettingsPage.vue).
  if (saveStatus.value === 'saving') return

  attemptedSubmit.value = true
  const errors = runValidation()
  fieldErrors.value = errors
  if (Object.keys(errors).length > 0) return

  saveStatus.value = 'saving'
  const outcome = await saveBookingPolicy(form, currentVersionToken)

  switch (outcome.kind) {
    case 'success':
      Object.assign(form, toFormValues(outcome.policy))
      currentVersionToken = outcome.policy.versionToken
      saveStatus.value = 'saved'
      return
    case 'validation-error':
      saveStatus.value = 'validation-error'
      return
    case 'version-conflict':
      saveStatus.value = 'version-conflict'
      return
    case 'network-error':
      saveStatus.value = 'network-error'
      return
    case 'unexpected-error':
      saveStatus.value = 'unexpected-error'
  }
}

function onRetryLoad() {
  void load()
}

// CA-093-02: la única recuperación correcta de un conflicto de versión es
// recargar la representación vigente, nunca reintentar con el mismo token.
function onReloadAfterConflict() {
  void load()
}
</script>

<template>
  <section class="booking-policy-page" aria-labelledby="booking-policy-page-title">
    <header class="booking-policy-page__header">
      <h1
        id="booking-policy-page-title"
        class="booking-policy-page__title"
        aria-label="Configuración de reserva pública"
      >
        <span class="booking-policy-page__title--desktop">Configuración de reserva pública</span>
        <span class="booking-policy-page__title--mobile">Reserva pública</span>
      </h1>
    </header>

    <div
      v-if="loadStatus === 'loading'"
      class="booking-policy-page__state"
      role="status"
      aria-live="polite"
    >
      <p>Cargando la configuración…</p>
    </div>

    <BaseAlert
      v-else-if="loadStatus === 'load-error'"
      variant="warning"
      title="No pudimos cargar la configuración"
      role="alert"
    >
      Revisa tu conexión e inténtalo de nuevo.
      <template #action>
        <BaseButton variant="secondary" type="button" @click="onRetryLoad">Reintentar</BaseButton>
      </template>
    </BaseAlert>

    <form v-else class="booking-policy-page__form" novalidate @submit.prevent="onSubmit">
      <div v-if="saveStatus === 'saved'" class="booking-policy-page__saved" role="status">
        <span class="booking-policy-page__saved-icon" aria-hidden="true">✓</span>
        <div>
          <p class="booking-policy-page__saved-title">Guardado</p>
          <p class="booking-policy-page__saved-copy">Los cambios se guardaron correctamente.</p>
        </div>
      </div>
      <BaseAlert
        v-if="saveStatus === 'validation-error'"
        variant="danger"
        title="No pudimos guardar los cambios"
        role="alert"
      >
        Revisa los valores: cada campo tiene un rango permitido.
      </BaseAlert>
      <BaseAlert
        v-if="saveStatus === 'version-conflict'"
        variant="warning"
        title="La configuración cambió"
        role="alert"
      >
        Alguien más actualizó esta configuración mientras la editabas. Recarga para ver la versión
        vigente; no perdiste lo que escribiste, pero debes revisarlo contra los valores nuevos.
        <template #action>
          <BaseButton variant="secondary" type="button" @click="onReloadAfterConflict"
            >Recargar</BaseButton
          >
        </template>
      </BaseAlert>
      <BaseAlert
        v-if="saveStatus === 'network-error'"
        variant="warning"
        title="No pudimos conectar"
        role="alert"
      >
        Revisa tu conexión e inténtalo de nuevo. No perdiste lo que escribiste.
      </BaseAlert>
      <BaseAlert
        v-if="saveStatus === 'unexpected-error'"
        variant="danger"
        title="Ocurrió un error inesperado"
        role="alert"
      >
        Inténtalo de nuevo en unos segundos. No perdiste lo que escribiste.
      </BaseAlert>

      <BaseInput
        :model-value="form.minAdvanceMinutes"
        type="number"
        name="minAdvanceMinutes"
        label="Anticipación mínima (minutos)"
        hint="El cliente no puede reservar con menos anticipación que esto. 0 desactiva el mínimo."
        required
        :show-required-marker="false"
        :min="0"
        :max="1440"
        :disabled="saveStatus === 'saving'"
        :error="fieldErrors.minAdvanceMinutes"
        @update:model-value="(value) => onNumberFieldInput('minAdvanceMinutes', value)"
      />

      <BaseInput
        :model-value="form.maxAdvanceDays"
        type="number"
        name="maxAdvanceDays"
        label="Ventana máxima (días)"
        hint="El cliente no puede reservar más lejos en el futuro que esto."
        required
        :show-required-marker="false"
        :min="1"
        :max="90"
        :disabled="saveStatus === 'saving'"
        :error="fieldErrors.maxAdvanceDays"
        @update:model-value="(value) => onNumberFieldInput('maxAdvanceDays', value)"
      />

      <div class="booking-policy-page__field">
        <label for="slotGridMinutes" class="booking-policy-page__label"
          >Rejilla de horarios (minutos)</label
        >
        <p class="booking-policy-page__hint">
          Cada cuántos minutos se ofrece una franja al cliente.
        </p>
        <select
          id="slotGridMinutes"
          name="slotGridMinutes"
          class="booking-policy-page__select"
          :disabled="saveStatus === 'saving'"
          :aria-invalid="!!fieldErrors.slotGridMinutes"
          :value="form.slotGridMinutes"
          @change="onSlotGridInput"
        >
          <option v-for="grid in ALLOWED_SLOT_GRID_MINUTES" :key="grid" :value="grid">
            {{ grid }} minutos
          </option>
        </select>
        <p v-if="fieldErrors.slotGridMinutes" class="booking-policy-page__field-error" role="alert">
          {{ fieldErrors.slotGridMinutes }}
        </p>
      </div>

      <BaseInput
        :model-value="form.cancellationDeadlineMinutes"
        type="number"
        name="cancellationDeadlineMinutes"
        label="Plazo de cancelación del cliente (minutos)"
        hint="Hasta cuántos minutos antes de la cita el cliente puede cancelar por su cuenta. 0 desactiva la cancelación propia del cliente."
        required
        :show-required-marker="false"
        :min="0"
        :max="10080"
        :disabled="saveStatus === 'saving'"
        :error="fieldErrors.cancellationDeadlineMinutes"
        @update:model-value="(value) => onNumberFieldInput('cancellationDeadlineMinutes', value)"
      />

      <fieldset class="booking-policy-page__fieldset">
        <legend class="booking-policy-page__label">Cancelación fuera de plazo</legend>
        <label class="booking-policy-page__checkbox-row">
          <input
            type="checkbox"
            name="lateCancellationClientAllowed"
            :checked="form.lateCancellationClientAllowed"
            :disabled="saveStatus === 'saving'"
            @change="onLateCancellationClientAllowedInput"
          />
          El cliente puede cancelar vencido el plazo
        </label>
        <label class="booking-policy-page__checkbox-row">
          <input
            type="checkbox"
            name="lateCancellationReasonRequired"
            :checked="form.lateCancellationReasonRequired"
            :disabled="saveStatus === 'saving'"
            @change="onLateCancellationReasonRequiredInput"
          />
          Esa cancelación exige un motivo
        </label>
        <p
          v-if="fieldErrors.lateCancellation"
          class="booking-policy-page__field-error"
          role="alert"
        >
          {{ fieldErrors.lateCancellation }}
        </p>
      </fieldset>

      <BaseButton
        type="submit"
        variant="primary"
        size="lg"
        :loading="saveStatus === 'saving'"
        :disabled="saveStatus === 'saving'"
        class="booking-policy-page__submit"
      >
        Guardar cambios
      </BaseButton>
    </form>
  </section>
</template>

<style scoped>
.booking-policy-page {
  --booking-policy-form-width: 540px;
  display: flex;
  flex-direction: column;
  min-height: 100%;
  max-width: none;
  padding: 34px 32px 48px;
  margin: 0 auto;
  color: var(--color-text-primary);
  background: var(--color-surface);
}

.booking-policy-page__header,
.booking-policy-page__form,
.booking-policy-page__state,
.booking-policy-page > :deep(.base-alert) {
  width: min(100%, var(--booking-policy-form-width));
  margin-inline: auto;
}

.booking-policy-page__header {
  margin-bottom: 12px;
}

.booking-policy-page__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(28px, 2.3vw, 34px);
  font-weight: var(--font-weight-h1);
  line-height: 1.14;
  color: var(--color-text-primary);
}

.booking-policy-page__title--mobile {
  display: none;
}

.booking-policy-page__state {
  padding: var(--space-5) 0;
  color: var(--color-text-secondary);
}

.booking-policy-page__form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.booking-policy-page__saved {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  min-height: 60px;
  padding: 13px 14px;
  color: var(--color-success-text);
  background: var(--color-success-surface);
  border: var(--border-width-normal) solid var(--color-success-border);
  border-radius: 3px;
}

.booking-policy-page__saved-icon {
  display: grid;
  flex: 0 0 auto;
  width: 17px;
  height: 17px;
  place-items: center;
  margin-top: 1px;
  border: var(--border-width-normal) solid currentColor;
  border-radius: 50%;
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 700;
  line-height: 1;
}

.booking-policy-page__saved-title,
.booking-policy-page__saved-copy {
  margin: 0;
}

.booking-policy-page__saved-title {
  font-size: 12px;
  font-weight: 600;
  line-height: 16px;
}

.booking-policy-page__saved-copy {
  font-size: 11px;
  line-height: 16px;
}

.booking-policy-page__field,
.booking-policy-page__fieldset {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 0;
  margin: 0;
  border: none;
}

.booking-policy-page__label {
  color: var(--color-text-primary);
  font-size: 11px;
  font-weight: 600;
  line-height: 14px;
}

.booking-policy-page__hint {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 11px;
  line-height: 15px;
}

.booking-policy-page__select {
  height: 34px;
  padding-inline: 10px;
  font-size: 12px;
  font-family: var(--font-family-base);
  color: var(--color-text-primary);
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: 3px;
}

.booking-policy-page__select:focus-visible {
  outline: none;
  border-color: var(--color-focus);
  box-shadow:
    0 0 0 2px var(--color-surface),
    0 0 0 4px var(--color-focus);
}

.booking-policy-page__checkbox-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--color-text-primary);
}

.booking-policy-page__field-error {
  margin: 0;
  color: var(--color-danger-text, #b3261e);
  font-size: 11px;
  line-height: 15px;
}

.booking-policy-page__form :deep(.base-input__wrapper) {
  gap: 4px;
}

.booking-policy-page__form :deep(.base-input__label) {
  color: var(--color-text-primary);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0;
  line-height: 14px;
  text-transform: none;
}

.booking-policy-page__form :deep(.base-input) {
  height: 34px;
  padding-inline: 10px;
  font-size: 12px;
  border-color: var(--color-border-subtle);
  border-bottom-width: var(--border-width-normal);
  border-radius: 3px;
}

.booking-policy-page__form :deep(.base-input:focus-visible) {
  border-color: var(--color-focus);
  box-shadow:
    0 0 0 2px var(--color-surface),
    0 0 0 4px var(--color-focus);
}

.booking-policy-page__form :deep(.base-button) {
  height: 34px;
  font-size: 12px;
  font-weight: 600;
  border-radius: 3px;
}

.booking-policy-page__submit {
  width: 100%;
  margin-top: 4px;
}

@media (max-width: 640px) {
  .booking-policy-page {
    padding: 16px 16px 28px;
  }

  .booking-policy-page__header {
    margin-bottom: 10px;
  }

  .booking-policy-page__title {
    font-size: 20px;
    line-height: 1.2;
  }

  .booking-policy-page__title--desktop {
    display: none;
  }

  .booking-policy-page__title--mobile {
    display: inline;
  }

  .booking-policy-page__form {
    gap: 10px;
  }
}
</style>
