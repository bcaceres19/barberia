<script setup lang="ts">
// Pantalla "Barbería" (HU-020): consultar y actualizar nombre, zona horaria
// y contacto opcional de la barbería activa. Estados discriminados: carga
// inicial, listo, guardando, error de campo, error recuperable y éxito
// (trabajo requerido §4.4). Un error recuperable NUNCA borra lo que el
// barbero ya escribió (CA-020-08); solo un guardado exitoso confirmado por
// el servidor reemplaza los valores del formulario.
import { onMounted, reactive, ref } from 'vue'
import { BaseAlert, BaseButton, BaseInput, PageHeader } from '@/shared/ui'
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
    <PageHeader title-id="settings-page-title" title="Barbería" />

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
      <BaseAlert v-if="saveStatus === 'saved'" variant="success" title="Guardado" role="status">
        Los cambios se guardaron correctamente.
      </BaseAlert>
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
        :maxlength="120"
        :disabled="saveStatus === 'saving'"
        :error="fieldErrors.name"
        @update:model-value="(value) => onFieldInput('name', value)"
      />

      <BaseInput
        :model-value="form.timezone"
        name="timezone"
        label="Zona horaria"
        hint="Identificador IANA, por ejemplo America/Bogota. Gobierna toda hora que se muestre de esta barbería."
        required
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
        hint="Opcional. Déjalo vacío si la barbería no tiene uno."
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
        hint="Opcional, formato internacional (ej. +573001234567). Déjalo vacío si la barbería no tiene uno."
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
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  max-width: 480px;
  padding: var(--space-4);
  margin: 0 auto;
}

.settings-page__state {
  padding: var(--space-4);
  color: var(--color-text-secondary);
}

.settings-page__form {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.settings-page__submit {
  width: 100%;
}
</style>
