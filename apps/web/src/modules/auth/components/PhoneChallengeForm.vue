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
import { BaseAlert, BaseButton, OtpInput } from '@/shared/ui'
import { requestChallenge, verifyChallenge } from '../api/challengeApi'

// DEC-081: el backend vigente solo resuelve un canal (WhatsApp). El
// mockup asignado (auth-eventos/README.md, eventos 07-12) representa tres
// variantes de canal; renderizar correo/ambos exigiría configuración,
// contrato o preferencia persistida que no existen hoy, así que esta
// pantalla solo compone la variante real (issue #213, trabajo requerido
// §4). Las variantes restantes quedan registradas como pendientes del
// issue funcional de DEC-081, no implementadas aquí.
const CHANNEL_CHIP = 'WhatsApp oficial'

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
      // El mockup de código rechazado conserva las seis ranuras llenas y
      // marcadas en error. No revela la causa (el mensaje sigue uniforme),
      // pero permite revisar visualmente qué código se rechazó.
      return
    default:
      verifyPhase.value = 'error'
  }
}
</script>

<template>
  <div class="phone-challenge">
    <!-- Fase previa a solicitar el código: no representada en los mockups
         asignados (trabajo requerido §4), conserva su composición actual. -->
    <template v-if="requestPhase === 'idle' || requestPhase === 'error'">
      <BaseAlert variant="info" title="Verifica tu teléfono">
        Para continuar, confirma tu identidad con el código que enviamos por WhatsApp a tu teléfono
        verificado.
      </BaseAlert>

      <BaseButton
        type="button"
        variant="secondary"
        size="md"
        :disabled="!canRequest"
        @click="onRequestCode"
      >
        Enviar código por WhatsApp
      </BaseButton>
    </template>

    <!-- Reto con código ya enviado (mockups 07-12): sección de la misma
         columna, con la regla de latón y rombo, seguida de la línea de
         canal en versalitas y su nota (trabajo requerido §4). -->
    <template v-else>
      <span class="phone-challenge__divider" aria-hidden="true"></span>

      <div class="phone-challenge__channel">
        <p class="phone-challenge__channel-chip">{{ CHANNEL_CHIP }}</p>
        <p class="phone-challenge__channel-note" role="status">
          Si tu cuenta existe y tu teléfono está verificado, recibirás un código por WhatsApp.
        </p>
      </div>

      <OtpInput
        v-model="code"
        label="Código de 6 dígitos"
        :disabled="verifyPhase === 'verifying' || disabled"
        :error="
          verifyPhase === 'invalid'
            ? 'El código no es válido o venció. Inténtalo de nuevo.'
            : undefined
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
          variant="secondary"
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

/* Regla de latón con rombo centrado: misma construcción que
   `.auth-split__rule` del cascarón, aquí a lo ancho de la columna
   (issue #213, trabajo requerido §4). */
.phone-challenge__divider {
  display: block;
  width: 100%;
  height: 2px;
  margin: var(--space-2) 0;
  background-color: var(--color-accent-brass);
  position: relative;
}

.phone-challenge__divider::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 8px;
  height: 8px;
  transform: translate(-50%, -50%) rotate(45deg);
  background-color: var(--color-accent-brass);
  box-shadow: 0 0 0 8px var(--color-canvas);
}

.phone-challenge__channel {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.phone-challenge__channel-chip {
  margin: 0;
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-accent-brass);
}

.phone-challenge__channel-note {
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
