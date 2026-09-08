<script setup lang="ts">
// Pantalla "Barbería" (HU-020): consultar y actualizar nombre, zona horaria
// y contacto opcional de la barbería activa. Estados discriminados: carga
// inicial, listo, guardando, error de campo, error recuperable y éxito
// (trabajo requerido §4.4). Un error recuperable NUNCA borra lo que el
// barbero ya escribió (CA-020-08); solo un guardado exitoso confirmado por
// el servidor reemplaza los valores del formulario.
import { onMounted, reactive, ref } from 'vue'
import { BaseAlert, BaseButton, BaseInput } from '@/shared/ui'
import { updateBarbershopName } from '@/modules/auth'
import { fetchBarbershopSettings, saveBarbershopSettings } from '../api/settingsApi'
import { toFormValues, type BarbershopSettingsFormValues } from '../model/barbershopSettings'
import {
  validateContactEmail,
  validateContactPhone,
  validateName,
  validateTimezone,
} from '../validation/settingsValidation'

type LoadStatus = 'loading' | 'ready' | 'load-error'
type SaveStatus =
  'idle' | 'saving' | 'saved' | 'validation-error' | 'network-error' | 'unexpected-error'

const loadStatus = ref<LoadStatus>('loading')
const saveStatus = ref<SaveStatus>('idle')

const form = reactive<BarbershopSettingsFormValues>({
  name: '',
  timezone: '',
  contactEmail: '',
  contactPhone: '',
})

type FieldErrors = Partial<Record<keyof BarbershopSettingsFormValues, string>>
const fieldErrors = ref<FieldErrors>({})
const attemptedSubmit = ref(false)

// requestToken evita que una carga inicial obsoleta (el barbero recargó la
// sección antes de que la primera respuesta llegara) sobrescriba el
// formulario con datos viejos (trabajo requerido §4.4: "cancela o ignora
// respuestas obsoletas si el usuario repite la carga").
let requestToken = 0

async function load() {
  const token = ++requestToken
  loadStatus.value = 'loading'
  const outcome = await fetchBarbershopSettings()
  if (token !== requestToken) return

  if (outcome.kind === 'success') {
    Object.assign(form, toFormValues(outcome.settings))
    loadStatus.value = 'ready'
    return
  }
  loadStatus.value = 'load-error'
}

onMounted(load)

function runValidation(): FieldErrors {
  const errors: FieldErrors = {}
  const nameError = validateName(form.name)
  if (nameError) errors.name = nameError
  const timezoneError = validateTimezone(form.timezone)
  if (timezoneError) errors.timezone = timezoneError
  const contactEmailError = validateContactEmail(form.contactEmail)
  if (contactEmailError) errors.contactEmail = contactEmailError
  const contactPhoneError = validateContactPhone(form.contactPhone)
  if (contactPhoneError) errors.contactPhone = contactPhoneError
  return errors
}

function onFieldInput<K extends keyof BarbershopSettingsFormValues>(
  field: K,
  value: string | number,
) {
  form[field] = String(value)
  if (attemptedSubmit.value) fieldErrors.value = runValidation()
}

async function onSubmit() {
  // Guardia de doble envío (mismo patrón que LoginForm/RecoveryResetStep):
  // la asignación síncrona bloquea un segundo clic mientras la solicitud
  // está en curso.
  if (saveStatus.value === 'saving') return

  attemptedSubmit.value = true
  const errors = runValidation()
  fieldErrors.value = errors
  if (Object.keys(errors).length > 0) return

  saveStatus.value = 'saving'
  const outcome = await saveBarbershopSettings(form)

  switch (outcome.kind) {
    case 'success':
      Object.assign(form, toFormValues(outcome.settings))
      // Solo el valor confirmado por el servidor cambia la cabecera
      // (CA-020-02): ningún camino de este componente la actualiza antes
      // de este punto.
      updateBarbershopName(outcome.settings.name)
      saveStatus.value = 'saved'
      return
    case 'validation-error':
      // La zona no reconocida (CA-020-03) es el caso típico que llega
      // hasta aquí: el cliente ya valida forma de nombre/correo/teléfono
      // antes de enviar, pero no puede confirmar el catálogo IANA real.
      saveStatus.value = 'validation-error'
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
</script>

<template>
  <section class="settings-page" aria-labelledby="settings-page-title">
    <header class="settings-page__header">
      <h1
        id="settings-page-title"
        class="settings-page__title"
        aria-label="Configuración de barbería"
      >
        <span class="settings-page__title--desktop">Configuración de barbería</span>
        <span class="settings-page__title--mobile">Barbería</span>
      </h1>
    </header>

    <div
      v-if="loadStatus === 'loading'"
      class="settings-page__state"
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

    <form v-else class="settings-page__form" novalidate @submit.prevent="onSubmit">
      <div v-if="saveStatus === 'saved'" class="settings-page__saved" role="status">
        <span class="settings-page__saved-icon" aria-hidden="true">✓</span>
        <div>
          <p class="settings-page__saved-title">Guardado</p>
          <p class="settings-page__saved-copy">Los cambios se guardaron correctamente.</p>
        </div>
      </div>
      <BaseAlert
        v-if="saveStatus === 'validation-error'"
        variant="danger"
        title="No pudimos guardar los cambios"
        role="alert"
      >
        Revisa los datos, especialmente la zona horaria: debe ser una zona IANA reconocida (ej.
        America/Bogota).
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
        :model-value="form.name"
        name="name"
        label="Nombre"
        required
        :show-required-marker="false"
        :maxlength="120"
        :disabled="saveStatus === 'saving'"
        :error="fieldErrors.name"
        @update:model-value="(value) => onFieldInput('name', value)"
      />

      <BaseInput
        :model-value="form.timezone"
        name="timezone"
        label="Zona horaria"
        required
        :show-required-marker="false"
        :maxlength="64"
        :disabled="saveStatus === 'saving'"
        :error="fieldErrors.timezone"
        @update:model-value="(value) => onFieldInput('timezone', value)"
      />

      <BaseInput
        :model-value="form.contactEmail"
        type="email"
        name="contactEmail"
        label="Correo de contacto"
        :maxlength="254"
        :disabled="saveStatus === 'saving'"
        :error="fieldErrors.contactEmail"
        @update:model-value="(value) => onFieldInput('contactEmail', value)"
      />

      <BaseInput
        :model-value="form.contactPhone"
        type="tel"
        name="contactPhone"
        label="Teléfono de contacto"
        :disabled="saveStatus === 'saving'"
        :error="fieldErrors.contactPhone"
        @update:model-value="(value) => onFieldInput('contactPhone', value)"
      />

      <BaseButton
        type="submit"
        variant="primary"
        size="lg"
        :loading="saveStatus === 'saving'"
        :disabled="saveStatus === 'saving'"
        class="settings-page__submit"
      >
        Guardar cambios
      </BaseButton>
    </form>
  </section>
</template>

<style scoped>
.settings-page {
  --settings-form-width: 540px;
  display: flex;
  flex-direction: column;
  min-height: 100%;
  gap: 0;
  max-width: none;
  padding: 34px 32px 48px;
  margin: 0 auto;
  color: var(--color-text-primary);
  background: var(--color-surface);
}

.settings-page__header,
.settings-page__form,
.settings-page__state,
.settings-page > :deep(.base-alert) {
  width: min(100%, var(--settings-form-width));
  margin-inline: auto;
}

.settings-page__header {
  margin-bottom: 12px;
}

.settings-page__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(28px, 2.3vw, 34px);
  font-weight: var(--font-weight-h1);
  line-height: 1.14;
  color: var(--color-text-primary);
}

.settings-page__title--mobile {
  display: none;
}

.settings-page__state {
  padding: var(--space-5) 0;
  color: var(--color-text-secondary);
}

.settings-page__form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.settings-page__saved {
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

.settings-page__saved-icon {
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

.settings-page__saved-title,
.settings-page__saved-copy {
  margin: 0;
}

.settings-page__saved-title {
  font-size: 12px;
  font-weight: 600;
  line-height: 16px;
}

.settings-page__saved-copy {
  font-size: 11px;
  line-height: 16px;
}

.settings-page__form :deep(.base-input__wrapper) {
  gap: 4px;
}

.settings-page__form :deep(.base-input__label) {
  color: var(--color-text-primary);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0;
  line-height: 14px;
  text-transform: none;
}

.settings-page__form :deep(.base-input) {
  height: 34px;
  padding-inline: 10px;
  font-size: 12px;
  border-color: var(--color-border-subtle);
  border-bottom-width: var(--border-width-normal);
  border-radius: 3px;
}

.settings-page__form :deep(.base-input:focus-visible) {
  border-color: var(--color-focus);
  box-shadow:
    0 0 0 2px var(--color-surface),
    0 0 0 4px var(--color-focus);
}

.settings-page__form :deep(.base-button) {
  height: 34px;
  font-size: 12px;
  font-weight: 600;
  border-radius: 3px;
}

.settings-page__submit {
  width: 100%;
  margin-top: 4px;
}

@media (max-width: 640px) {
  .settings-page {
    padding: 16px 16px 28px;
  }

  .settings-page__header {
    margin-bottom: 10px;
  }

  .settings-page__title {
    font-size: 20px;
    line-height: 1.2;
  }

  .settings-page__title--desktop {
    display: none;
  }

  .settings-page__title--mobile {
    display: inline;
  }

  .settings-page__form {
    gap: 10px;
  }

  .settings-page__saved {
    gap: 10px;
    min-height: 58px;
    padding: 9px 10px;
  }

  .settings-page__saved-title {
    font-size: 10px;
    line-height: 14px;
  }

  .settings-page__saved-copy {
    font-size: 9px;
    line-height: 13px;
  }

  .settings-page__form :deep(.base-input) {
    height: 28px;
    padding-inline: 8px;
    font-size: 11px;
  }

  .settings-page__form :deep(.base-input__label) {
    font-size: 10px;
    line-height: 12px;
  }

  .settings-page__form :deep(.base-button) {
    height: 30px;
    font-size: 11px;
  }
}
</style>
