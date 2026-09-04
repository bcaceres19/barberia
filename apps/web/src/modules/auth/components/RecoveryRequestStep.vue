<script setup lang="ts">
// Paso 1/3 de HU-011: solicitar el código de recuperación. Dueño de su
// propio envío (mismo patrón que `PhoneChallengeForm.vue`): valida en
// cliente, llama al API real y decide cuándo avanzar al paso 2. Un 202
// (DEC-065, no enumeración) siempre avanza; un fallo de transporte real
// (red o 5xx) no revela nada sobre la cuenta, así que sí se distingue con
// un error accionable (a diferencia del "siempre avanzar" de
// `PhoneChallengeForm`, que ahí es correcto porque ese reto ya está detrás
// de un intento de acceso autenticable).
import { computed, ref } from 'vue'
import { BaseAlert, BaseButton, BaseInput } from '@/shared/ui'
import { requestRecovery } from '../api/recoveryApi'
import { validateRecoveryEmail } from '../validation/recoveryValidation'

const emit = defineEmits<{
  advance: [payload: { email: string }]
}>()

const email = ref('')
const fieldError = ref<string | undefined>(undefined)
const attemptedSubmit = ref(false)
const status = ref<'idle' | 'submitting' | 'network-error' | 'unexpected-error'>('idle')

const isSubmitting = computed(() => status.value === 'submitting')

const handleEmailInput = (value: string | number) => {
  email.value = String(value)
  if (attemptedSubmit.value) {
    fieldError.value = validateRecoveryEmail(email.value)
  }
}

async function onSubmit() {
  if (isSubmitting.value) return

  attemptedSubmit.value = true
  const error = validateRecoveryEmail(email.value)
  fieldError.value = error
  if (error) return

  status.value = 'submitting'
  const outcome = await requestRecovery(email.value.trim())

  switch (outcome.kind) {
    case 'accepted':
      emit('advance', { email: email.value.trim() })
      return
    case 'network-error':
      status.value = 'network-error'
      return
    case 'unexpected-error':
      status.value = 'unexpected-error'
  }
}
</script>

<template>
  <form class="recovery-request" novalidate @submit.prevent="onSubmit">
    <p class="recovery-request__hint">
      Escribe el correo de tu cuenta. Si existe, te enviaremos un código de un solo uso por WhatsApp
      y correo.
    </p>

    <BaseInput
      :model-value="email"
      type="email"
      name="email"
      label="Correo"
      autocomplete="username"
      placeholder="tu-correo@ejemplo.com"
      required
      :show-required-marker="false"
      :disabled="isSubmitting"
      :error="fieldError"
      @update:model-value="handleEmailInput"
    />

    <BaseButton
      type="submit"
      variant="primary"
      size="lg"
      :loading="isSubmitting"
      :disabled="isSubmitting"
      class="recovery-request__submit"
    >
      {{ isSubmitting ? 'Enviando…' : 'Enviar código' }}
    </BaseButton>

    <p class="recovery-back">
      <RouterLink :to="{ name: 'acceso' }">Volver al acceso</RouterLink>
    </p>

    <!-- Ancla de alertas: después del grupo de acciones y de "Volver al
         acceso" (trabajo requerido §3, auth-eventos/README.md). -->
    <BaseAlert
      v-if="status === 'network-error'"
      variant="warning"
      title="No pudimos conectar"
      role="alert"
    >
      Revisa tu conexión e inténtalo de nuevo.
      <template #action>
        <BaseButton variant="secondary" size="md" type="button" @click="onSubmit">
          Reintentar
        </BaseButton>
      </template>
    </BaseAlert>
    <BaseAlert
      v-if="status === 'unexpected-error'"
      variant="danger"
      title="Ocurrió un error inesperado"
      role="alert"
    >
      Inténtalo de nuevo en unos segundos.
    </BaseAlert>
  </form>
</template>

<style scoped>
.recovery-request {
  display: flex;
  flex-direction: column;
  gap: var(--space-8);
  width: 100%;
  margin-top: var(--space-8);
}

.recovery-request :deep(.base-input) {
  height: 56px;
  font-size: 18px;
}

.recovery-request :deep(.base-input__label) {
  font-size: 16px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.recovery-request__submit {
  width: 100%;
  min-height: 56px;
  font-size: 18px;
}

@media (min-width: 1024px) {
  .recovery-request {
    margin-top: 0;
  }
}

.recovery-request__hint {
  margin: 0;
  font-size: 20px;
  line-height: 28px;
  color: var(--color-text-secondary);
}

@media (min-width: 1024px) {
  .recovery-request__hint {
    font-size: 18px;
    line-height: 26px;
  }
}

.recovery-back {
  margin: 0;
  text-align: center;
  font-size: var(--font-size-body-sm);
}

.recovery-back a {
  color: var(--color-action-primary);
  font-weight: 500;
}
</style>
