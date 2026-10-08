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
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useToast, useVocabulary } from '@/shared/composables'
import { BaseAlert, BaseButton, BaseInput, DiamondLoader } from '@/shared/ui'
import { fetchBookingPolicy, saveBookingPolicy } from '../api/bookingPolicyApi'
import BookingPolicyPreview from '../components/BookingPolicyPreview.vue'
import OptionGroup from '../components/OptionGroup.vue'
import SaveBar from '../components/SaveBar.vue'
import SettingsPanel from '../components/SettingsPanel.vue'
import SwitchField from '../components/SwitchField.vue'
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

const LOADING_PHRASES = [
  'Midiendo la agenda',
  'Afinando la rejilla',
  'Ajustando los plazos',
  'Todo a su hora',
] as const

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
// Último valor confirmado por el servidor: contra él se mide qué cambió y a él
// vuelve «Descartar».
const savedForm = ref<BookingPolicyFormValues>({ ...form })
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
// La confirmación persistente sigue en la pantalla; el aviso emergente la
// acompaña (DEC-095).
const toast = useToast()
const v = useVocabulary()

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
    savedForm.value = { ...form }
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

const SLOT_GRID_OPTIONS = ALLOWED_SLOT_GRID_MINUTES.map((minutes) => ({
  value: String(minutes),
  label: `${minutes} min`,
}))

function onSlotGridInput(value: string) {
  form.slotGridMinutes = Number(value)
  if (attemptedSubmit.value) fieldErrors.value = runValidation()
}

function onLateCancellationClientAllowedInput(value: boolean) {
  form.lateCancellationClientAllowed = value
  if (attemptedSubmit.value) fieldErrors.value = runValidation()
}

function onLateCancellationReasonRequiredInput(value: boolean) {
  form.lateCancellationReasonRequired = value
  if (attemptedSubmit.value) fieldErrors.value = runValidation()
}

const windowDirty = computed(
  () =>
    form.minAdvanceMinutes !== savedForm.value.minAdvanceMinutes ||
    form.maxAdvanceDays !== savedForm.value.maxAdvanceDays ||
    form.slotGridMinutes !== savedForm.value.slotGridMinutes,
)
const cancellationDirty = computed(
  () =>
    form.cancellationDeadlineMinutes !== savedForm.value.cancellationDeadlineMinutes ||
    form.lateCancellationClientAllowed !== savedForm.value.lateCancellationClientAllowed ||
    form.lateCancellationReasonRequired !== savedForm.value.lateCancellationReasonRequired,
)
const dirty = computed(() => windowDirty.value || cancellationDirty.value)
const dirtySummary = computed(() =>
  [windowDirty.value && 'Ventana de reserva', cancellationDirty.value && 'Cancelación']
    .filter(Boolean)
    .join(' · '),
)

// Una confirmación de «Guardado» deja de ser cierta en cuanto vuelve a haber
// cambios sin guardar.
watch(dirty, (isDirty) => {
  if (isDirty && saveStatus.value === 'saved') saveStatus.value = 'idle'
})

function onDiscard() {
  Object.assign(form, savedForm.value)
  fieldErrors.value = {}
  attemptedSubmit.value = false
  if (saveStatus.value === 'validation-error') saveStatus.value = 'idle'
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
      savedForm.value = { ...form }
      currentVersionToken = outcome.policy.versionToken
      saveStatus.value = 'saved'
      toast.success('Política de reservas guardada', {
        detail: 'Las nuevas reglas se aplican a las próximas reservas.',
      })
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
      <p class="booking-policy-page__subtitle">
        Define cuándo pueden reservar tus clientes y hasta cuándo pueden cancelar.
      </p>
      <span class="booking-policy-page__rule" aria-hidden="true"
        ><span class="booking-policy-page__rule-line" /><span
          class="booking-policy-page__rule-diamond" /><span class="booking-policy-page__rule-line"
      /></span>
    </header>

    <Transition name="booking-policy-content" mode="out-in">
      <div
        v-if="loadStatus === 'loading'"
        key="loading"
        class="booking-policy-page__state"
        role="status"
        aria-live="polite"
      >
        <DiamondLoader
          label="Cargando la configuración…"
          layout="inline"
          :phrases="LOADING_PHRASES"
        />
        <div class="booking-policy-page__skeletons" aria-hidden="true">
          <span
            v-for="n in 2"
            :key="n"
            class="booking-policy-page__skeleton"
            :style="{ '--n': n }"
          />
        </div>
      </div>

      <BaseAlert
        v-else-if="loadStatus === 'load-error'"
        key="error"
        variant="warning"
        title="No pudimos cargar la configuración"
        role="alert"
      >
        Revisa tu conexión e inténtalo de nuevo.
        <template #action>
          <BaseButton variant="secondary" type="button" @click="onRetryLoad">Reintentar</BaseButton>
        </template>
      </BaseAlert>

      <div v-else key="ready" class="booking-policy-page__layout">
        <form class="booking-policy-page__form policy-ink" novalidate @submit.prevent="onSubmit">
          <div v-if="saveStatus === 'saved'" class="booking-policy-page__saved" role="status">
            <span class="booking-policy-page__saved-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
                <path d="M5 12.5l4.5 4.5L19 7.5" pathLength="1" />
              </svg>
            </span>
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
            Alguien más actualizó esta configuración mientras la editabas. Recarga para ver la
            versión vigente; no perdiste lo que escribiste, pero debes revisarlo contra los valores
            nuevos.
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

          <!-- 01 · Ventana de reserva -->
          <SettingsPanel
            id="reserva-ventana"
            number="01"
            title="Ventana de reserva"
            description="Desde cuándo y hasta cuándo puede reservar un cliente, y cada cuánto se ofrece una franja."
            scope="shop"
            :scope-label="`Para ${v.allTheBusiness}`"
            style="--panel-index: 0"
          >
            <div class="policy-grid">
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
            </div>

            <div class="policy-field">
              <span id="field-slot-grid" class="policy-field__label"
                >Rejilla de horarios (minutos)</span
              >
              <p class="policy-field__hint">
                Cada cuántos minutos se ofrece una franja al cliente.
              </p>
              <OptionGroup
                :model-value="String(form.slotGridMinutes)"
                :options="SLOT_GRID_OPTIONS"
                label="Rejilla de horarios (minutos)"
                :disabled="saveStatus === 'saving'"
                class="policy-grid-options"
                @update:model-value="onSlotGridInput"
              >
                <template #default="{ option, selected }">
                  <span class="grid-chip" :class="{ 'grid-chip--selected': selected }">
                    <span class="grid-chip__value">{{ option.label.replace(' min', '') }}</span>
                    <span class="grid-chip__unit">min</span>
                  </span>
                </template>
              </OptionGroup>
              <p v-if="fieldErrors.slotGridMinutes" class="policy-field__error" role="alert">
                {{ fieldErrors.slotGridMinutes }}
              </p>
            </div>
          </SettingsPanel>

          <!-- 02 · Cancelación del cliente -->
          <SettingsPanel
            id="reserva-cancelacion"
            number="02"
            title="Cancelación del cliente"
            description="Hasta cuándo puede cancelar por su cuenta y qué pasa si lo hace fuera de plazo."
            scope="shop"
            :scope-label="`Para ${v.allTheBusiness}`"
            style="--panel-index: 1"
          >
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
              @update:model-value="
                (value) => onNumberFieldInput('cancellationDeadlineMinutes', value)
              "
            />

            <fieldset class="policy-fieldset">
              <legend class="policy-field__label">Cancelación fuera de plazo</legend>
              <SwitchField
                :model-value="form.lateCancellationClientAllowed"
                label="El cliente puede cancelar vencido el plazo"
                :disabled="saveStatus === 'saving'"
                @update:model-value="onLateCancellationClientAllowedInput"
              />
              <SwitchField
                :model-value="form.lateCancellationReasonRequired"
                label="Esa cancelación exige un motivo"
                :disabled="saveStatus === 'saving'"
                @update:model-value="onLateCancellationReasonRequiredInput"
              />
              <p v-if="fieldErrors.lateCancellation" class="policy-field__error" role="alert">
                {{ fieldErrors.lateCancellation }}
              </p>
            </fieldset>
          </SettingsPanel>

          <div class="booking-policy-page__save">
            <SaveBar
              v-if="dirty || saveStatus === 'saving'"
              :summary="dirtySummary"
              :saving="saveStatus === 'saving'"
              @discard="onDiscard"
              @save="onSubmit"
            />
          </div>
        </form>

        <BookingPolicyPreview :policy="form" />
      </div>
    </Transition>
  </section>
</template>

<style scoped>
/* Superficie tinta de punta a punta, el mismo canvas que Agenda, Servicios,
   Barberos y Configuración (estandar-diseno-visual.md §3). */
.booking-policy-page {
  --policy-width: 1040px;

  display: flex;
  flex-direction: column;
  gap: 24px;
  min-height: 100%;
  padding: 34px 32px 48px;
  color: var(--color-on-strong);
  /* Transparente: la tinta y el fondo animado los pone el cascarón. */
  background: transparent;
  box-sizing: border-box;
}

.booking-policy-page__header,
.booking-policy-page__state,
.booking-policy-page__layout,
.booking-policy-page > :deep(.base-alert) {
  width: min(100%, var(--policy-width));
  margin-inline: auto;
}

.booking-policy-page__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h1);
  font-weight: var(--font-weight-h1);
  line-height: var(--font-size-h1-line);
  animation: policy-title-enter 520ms var(--motion-easing-standard) both;
}

@media (min-width: 1024px) {
  .booking-policy-page__title {
    font-size: var(--font-size-title-page);
    line-height: 46px;
  }
}

.booking-policy-page__title--mobile {
  display: none;
}

.booking-policy-page__subtitle {
  margin: 4px 0 0;
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
  color: var(--color-on-strong-muted);
  animation: policy-title-enter 520ms var(--motion-easing-standard) 80ms both;
}

/* Regla – rombo – regla, el divisor de la casa, que se dibuja al entrar. */
.booking-policy-page__rule {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 18px;
}

.booking-policy-page__rule-line {
  height: 1px;
  flex: 1;
  background: color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent);
  transform: scaleX(0);
  transform-origin: left center;
  animation: policy-rule-draw 700ms var(--motion-easing-standard) 160ms both;
}

.booking-policy-page__rule-line:last-child {
  transform-origin: right center;
}

.booking-policy-page__rule-diamond {
  width: 7px;
  height: 7px;
  background: var(--color-brand-accent-surface);
  transform: rotate(45deg);
  animation: policy-diamond-pop 420ms cubic-bezier(0.34, 1.56, 0.64, 1) 260ms both;
}

.booking-policy-page__state {
  display: flex;
  flex-direction: column;
  gap: 16px;
  color: var(--color-on-strong-muted);
}

.booking-policy-page__skeletons {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.booking-policy-page__skeleton {
  display: block;
  height: 132px;
  background: color-mix(in srgb, var(--color-on-strong) 8%, transparent);
  border-radius: 3px;
  animation: policy-skeleton-pulse 1400ms ease-in-out infinite;
  animation-delay: calc(var(--n) * 120ms);
}

/* Formulario a la izquierda, vista previa fija a la derecha. */
.booking-policy-page__layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 24px;
  align-items: start;
}

@media (min-width: 1100px) {
  .booking-policy-page__layout {
    grid-template-columns: minmax(0, 1fr) 320px;
    gap: 32px;
  }
}

.booking-policy-page__form {
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-width: 0;
}

.booking-policy-page__save {
  position: sticky;
  bottom: 16px;
  z-index: var(--layer-sticky);
}

/* Confirmación persistente: nota al margen de éxito con el check que se dibuja. */
.booking-policy-page__saved {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 13px 16px;
  color: var(--color-success-on-strong);
  background: color-mix(in srgb, var(--color-success-on-strong) 9%, transparent);
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-success-on-strong) 35%, transparent);
  border-left: 4px solid var(--color-success-on-strong);
  border-radius: 3px;
  animation: policy-title-enter 360ms var(--motion-easing-standard) both;
}

.booking-policy-page__saved-icon {
  display: grid;
  flex: 0 0 auto;
  width: 22px;
  height: 22px;
  place-items: center;
  border: var(--border-width-normal) solid currentColor;
  border-radius: 50%;
}

.booking-policy-page__saved-icon path {
  stroke: currentColor;
  stroke-width: 2.4;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-dasharray: 1;
  stroke-dashoffset: 1;
  animation: policy-check-draw 420ms var(--motion-easing-standard) 120ms forwards;
}

.booking-policy-page__saved-title,
.booking-policy-page__saved-copy {
  margin: 0;
}

.booking-policy-page__saved-title {
  font-size: var(--font-size-body-sm);
  font-weight: 600;
  line-height: 20px;
}

.booking-policy-page__saved-copy {
  font-size: var(--font-size-caption);
  line-height: 18px;
  color: var(--color-on-strong-muted);
}

/* --- Campos de una sección ------------------------------------------------- */

.policy-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
}

.policy-field,
.policy-fieldset {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-width: 0;
  padding: 0;
  margin: 0;
  border: none;
}

.policy-fieldset {
  gap: 14px;
}

.policy-field__label {
  padding: 0;
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.08em;
  line-height: 14px;
  text-transform: uppercase;
}

.policy-field__hint {
  margin: -4px 0 0;
  max-width: 60ch;
  font-size: var(--font-size-caption);
  line-height: 18px;
  color: var(--color-on-strong-muted);
}

.policy-field__error {
  margin: 0;
  color: var(--color-danger-on-strong);
  font-size: var(--font-size-caption);
  line-height: 18px;
}

/* Rejilla: seis fichas con el valor en serif, como las franjas del cliente. */
.policy-grid-options {
  --option-group-gap: 10px;
}

.grid-chip {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-width: 68px;
  min-height: 64px;
  padding: 8px 12px;
  background: color-mix(in srgb, var(--color-on-strong) 3%, transparent);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: 3px;
  transition:
    transform var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard),
    background-color var(--motion-duration-base) var(--motion-easing-standard);
}

.grid-chip:hover {
  transform: translateY(-3px);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
}

.grid-chip--selected {
  background: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
  border-color: var(--color-brand-accent-surface);
  box-shadow: 0 0 0 1px var(--color-brand-accent-surface);
}

.grid-chip__value {
  font-family: var(--font-display);
  font-size: var(--font-size-title-item);
  line-height: 24px;
  color: var(--color-on-strong);
}

.grid-chip__unit {
  font-size: var(--font-size-caption);
  letter-spacing: 0.08em;
  line-height: 14px;
  text-transform: uppercase;
  color: var(--color-on-strong-muted);
}

.grid-chip--selected .grid-chip__unit {
  color: var(--color-brand-accent-surface);
}

/* --- Campos reglados: mismo tratamiento que Servicios, Barberos y Configuración --- */

.policy-ink :deep(.base-input) {
  --input-bg: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  --input-border-color: color-mix(in srgb, var(--color-on-strong) 12%, transparent);
  --input-border-base-color: color-mix(in srgb, var(--color-on-strong) 30%, transparent);
  --input-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);

  height: 44px;
  color: var(--color-on-strong);
}

.policy-ink :deep(.base-input__label),
.policy-ink :deep(.base-input__required) {
  color: var(--color-brand-accent-surface);
}

.policy-ink :deep(.base-input__hint) {
  color: var(--color-on-strong-muted);
}

.policy-ink :deep(.base-input:hover:not(:disabled):not(.base-input--invalid)) {
  border-color: color-mix(in srgb, var(--color-on-strong) 26%, transparent);
}

.policy-ink :deep(.base-input:disabled),
.policy-ink :deep(.base-input--disabled) {
  background-color: var(--input-bg);
  border-color: var(--input-border-color);
  border-bottom-color: color-mix(in srgb, var(--color-on-strong) 20%, transparent);
  color: var(--color-on-strong-muted);
  opacity: 0.45;
}

.policy-ink :deep(.base-input--invalid) {
  background-color: color-mix(in srgb, var(--color-danger-on-strong) 7%, transparent);
  border-color: var(--input-border-color);
  border-bottom-color: var(--color-danger-on-strong);
}

.policy-ink :deep(.base-input__error) {
  color: var(--color-danger-on-strong);
}

.policy-ink :deep(.base-button) {
  --btn-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);
}

.policy-ink :deep(.base-button--secondary) {
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.policy-ink :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
}

.policy-ink :deep(.base-button:disabled) {
  opacity: 0.4;
}

/* Fundido entre carga/error/listo. */
.booking-policy-content-enter-active,
.booking-policy-content-leave-active {
  transition: opacity var(--motion-duration-base) var(--motion-easing-standard);
}

.booking-policy-content-enter-from,
.booking-policy-content-leave-to {
  opacity: 0;
}

@keyframes policy-title-enter {
  from {
    opacity: 0;
    transform: translateY(8px);
  }

  to {
    opacity: 1;
    transform: none;
  }
}

@keyframes policy-rule-draw {
  to {
    transform: scaleX(1);
  }
}

@keyframes policy-diamond-pop {
  from {
    opacity: 0;
    transform: rotate(-45deg) scale(0);
  }

  to {
    opacity: 1;
    transform: rotate(45deg) scale(1);
  }
}

@keyframes policy-skeleton-pulse {
  0%,
  100% {
    opacity: 0.6;
  }

  50% {
    opacity: 1;
  }
}

@keyframes policy-check-draw {
  to {
    stroke-dashoffset: 0;
  }
}

@media (max-width: 640px) {
  .booking-policy-page {
    gap: 16px;
    padding: 16px 16px 28px;
  }

  .booking-policy-page__title {
    font-size: var(--font-size-title-item);
    line-height: 1.2;
  }

  .booking-policy-page__title--desktop {
    display: none;
  }

  .booking-policy-page__title--mobile {
    display: inline;
  }

  .booking-policy-page__save {
    bottom: 8px;
  }

  .policy-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .grid-chip {
    min-width: 60px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .booking-policy-page__title,
  .booking-policy-page__subtitle,
  .booking-policy-page__rule-line,
  .booking-policy-page__rule-diamond,
  .booking-policy-page__skeleton,
  .booking-policy-page__saved {
    animation: none;
  }

  .booking-policy-page__rule-line {
    transform: none;
  }

  .booking-policy-page__saved-icon path {
    animation: none;
    stroke-dashoffset: 0;
  }
}
</style>
