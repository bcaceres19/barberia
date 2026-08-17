<script setup lang="ts">
// Reto telefónico de HU-007 (DEC-062): se muestra cuando el acceso responde
// 429 (umbral superado). Pide el código de 6 dígitos enviado por WhatsApp y,
// al verificarlo, avisa a `LoginPage` para que reintente el acceso — el
// servidor ya limpió el escalamiento de esa IP en la misma verificación, así
// que no hace falta ningún dato adicional para ese reintento.
//
// El mensaje de "código enviado" es deliberadamente genérico (DEC-062, no
// enumeración): nunca afirma que un mensaje real llegó, porque el servidor
// tampoco lo confirma.
import { computed, ref } from 'vue'
import { BaseAlert, BaseButton, BaseInput } from '@/shared/ui'
import { requestChallenge, verifyChallenge } from '../api/challengeApi'

const props = defineProps<{
  email: string
  /** Se deshabilitan las acciones mientras el formulario de acceso está
   * reintentando el login tras un reto verificado (evita doble envío). */
  disabled: boolean
}>()

const emit = defineEmits<{
  verified: []
}>()

type RequestPhase = 'idle' | 'requesting' | 'sent' | 'error'
type VerifyPhase = 'idle' | 'verifying' | 'invalid' | 'error'

const requestPhase = ref<RequestPhase>('idle')
const verifyPhase = ref<VerifyPhase>('idle')
const code = ref('')
const resendCooldownActive = ref(false)

const codeDigitsOnly = computed(() => /^[0-9]{6}$/.test(code.value))
const canVerify = computed(
  () => codeDigitsOnly.value && verifyPhase.value !== 'verifying' && !props.disabled,
)
const canRequest = computed(
  () => requestPhase.value !== 'requesting' && !resendCooldownActive.value && !props.disabled,
)

async function onRequestCode() {
  if (!canRequest.value) return
  requestPhase.value = 'requesting'

  const outcome = await requestChallenge(props.email)

  // Siempre se trata como "solicitud recibida", incluso ante un error de
  // red o inesperado: reintentar el envío (con el mismo cooldown del
  // servidor) es más seguro que bloquear al barbero en un estado sin
  // salida por un fallo transitorio de nuestro propio resumen de UI.
  requestPhase.value = outcome.kind === 'accepted' ? 'sent' : 'error'

  // Cooldown de cliente (60 s, DEC-062) para que el botón no invite a un
  // reenvío que el servidor descartará en silencio; el servidor sigue
  // siendo la única autoridad real del límite.
  resendCooldownActive.value = true
  setTimeout(() => {
    resendCooldownActive.value = false
  }, 60_000)
}

async function onVerifyCode() {
  if (!canVerify.value) return
  verifyPhase.value = 'verifying'

  const outcome = await verifyChallenge(props.email, code.value)

  switch (outcome.kind) {
    case 'verified':
      emit('verified')
      return
    case 'invalid-code':
      verifyPhase.value = 'invalid'
      code.value = ''
      return
    default:
      verifyPhase.value = 'error'
  }
}
</script>

<template>
  <div class="phone-challenge">
    <BaseAlert variant="info" title="Verifica tu teléfono">
      Para continuar, confirma tu identidad con el código que enviamos por WhatsApp a tu teléfono
      verificado.
    </BaseAlert>

    <BaseButton
      v-if="requestPhase === 'idle' || requestPhase === 'error'"
      type="button"
      variant="secondary"
      size="md"
      :disabled="!canRequest"
      @click="onRequestCode"
    >
      Enviar código por WhatsApp
    </BaseButton>

    <template v-else>
      <p class="phone-challenge__sent" role="status">
        Si tu cuenta existe y tu teléfono está verificado, recibirás un código por WhatsApp.
      </p>

      <BaseInput
        :model-value="code"
        type="text"
        name="challengeCode"
        label="Código de 6 dígitos"
        pattern="[0-9]*"
        :maxlength="6"
        autocomplete="one-time-code"
        placeholder="000000"
        :disabled="verifyPhase === 'verifying' || disabled"
        :error="
          verifyPhase === 'invalid'
            ? 'El código no es válido o venció. Inténtalo de nuevo.'
            : undefined
        "
        @update:model-value="
          (value) =>
            (code = String(value)
              .replace(/[^0-9]/g, '')
              .slice(0, 6))
        "
      />

      <div class="phone-challenge__actions">
        <BaseButton
          type="button"
          variant="primary"
          size="md"
          :loading="verifyPhase === 'verifying'"
          :disabled="!canVerify"
          @click="onVerifyCode"
        >
          Verificar código
        </BaseButton>
        <BaseButton
          type="button"
          variant="ghost"
          size="md"
          :disabled="!canRequest"
          @click="onRequestCode"
        >
          Reenviar código
        </BaseButton>
      </div>

      <p v-if="verifyPhase === 'error'" class="phone-challenge__error" role="alert">
        No pudimos verificar el código. Revisa tu conexión e inténtalo de nuevo.
      </p>
    </template>
  </div>
</template>

<style scoped>
.phone-challenge {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding-top: var(--space-2);
}

.phone-challenge__sent {
  margin: 0;
  font-size: var(--font-size-body-sm);
  color: var(--color-text-secondary);
}

.phone-challenge__actions {
  display: flex;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.phone-challenge__error {
  margin: 0;
  font-size: var(--font-size-body-sm);
  color: var(--color-danger-text);
}
</style>
