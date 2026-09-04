<script setup lang="ts">
// Paso 2/3 de HU-011: verificar el código de 6 dígitos. El destino
// (teléfono/correo) NO se muestra aquí: el backend solo lo revela
// enmascarado en la respuesta exitosa de esta misma verificación
// (CA-008-06, DEC-065) — mostrarlo antes abriría el oráculo que esa
// decisión evita. Por eso el paso 3 (`RecoveryResetStep.vue`), no este,
// es quien confirma "enviamos tu código a b***@c***.test" como contexto.
//
// El reenvío es un cooldown rastreado en cliente (60 s por defecto,
// `APP_RECOVERY_RESEND_COOLDOWN_SECONDS`), mismo patrón que
// `PhoneChallengeForm.vue`, pero con cuenta regresiva visible (CA-011-04
// exige "cuánto falta", no solo un botón deshabilitado).
import { computed, onUnmounted, ref } from 'vue'
import { BaseAlert, BaseButton, OtpInput } from '@/shared/ui'
import { requestRecovery, verifyRecovery } from '../api/recoveryApi'
import { validateRecoveryCode } from '../validation/recoveryValidation'

const RESEND_COOLDOWN_SECONDS = 60

const props = defineProps<{
  email: string
}>()

const emit = defineEmits<{
  advance: [payload: { resetToken: string; maskedPhone: string; maskedEmail: string }]
}>()

const code = ref('')
const fieldError = ref<string | undefined>(undefined)
const attemptedSubmit = ref(false)
const verifyStatus = ref<
  'idle' | 'verifying' | 'invalid-code' | 'network-error' | 'unexpected-error'
>('idle')
const resendCooldownRemaining = ref(0)
let cooldownTimer: ReturnType<typeof setInterval> | undefined

const isVerifying = computed(() => verifyStatus.value === 'verifying')
const canResend = computed(() => resendCooldownRemaining.value <= 0)
// El error del código vive bajo las casillas, sin una alerta global que lo
// repita (mockup 06-verificacion-codigo-invalido, trabajo requerido §6):
// mismo tratamiento que el reto de acceso. Conserva el texto actual del
// código (CA-011-*: no se toca contenido, solo su ancla).
const otpError = computed(() =>
  verifyStatus.value === 'invalid-code'
    ? 'El código no es correcto o ya venció. Puedes reenviarlo o revisar lo que escribiste.'
    : fieldError.value,
)

function startResendCooldown() {
  resendCooldownRemaining.value = RESEND_COOLDOWN_SECONDS
  cooldownTimer = setInterval(() => {
    resendCooldownRemaining.value = Math.max(0, resendCooldownRemaining.value - 1)
    if (resendCooldownRemaining.value === 0 && cooldownTimer) {
      clearInterval(cooldownTimer)
      cooldownTimer = undefined
    }
  }, 1_000)
}

onUnmounted(() => {
  if (cooldownTimer) clearInterval(cooldownTimer)
})

// Cooldown activo desde el momento en que se entra a este paso: el
// código ya se envió una vez al avanzar del paso 1 (trabajo requerido §4).
startResendCooldown()

async function onResend() {
  if (!canResend.value) return
  startResendCooldown()
  // Mismo trato que el paso 1: la solicitud siempre se intenta contra el
  // API real, pero el resultado no cambia el cooldown de cliente ni
  // revela nada distinto (DEC-065).
  await requestRecovery(props.email)
}

const handleCodeInput = (value: string) => {
  code.value = value
  if (attemptedSubmit.value) {
    fieldError.value = validateRecoveryCode(code.value)
  }
}

async function onSubmit() {
  if (isVerifying.value) return

  attemptedSubmit.value = true
  const error = validateRecoveryCode(code.value)
  fieldError.value = error
  if (error) return

  verifyStatus.value = 'verifying'
  const outcome = await verifyRecovery(props.email, code.value)

  switch (outcome.kind) {
    case 'verified':
      emit('advance', {
        resetToken: outcome.resetToken,
        maskedPhone: outcome.maskedPhone,
        maskedEmail: outcome.maskedEmail,
      })
      return
    case 'invalid-code':
      // Mismo error uniforme para incorrecto, vencido, agotado o cuenta
      // inexistente (DEC-064/DEC-065): el cliente tampoco puede
      // distinguir el motivo. CT-007 (docs/00-control/contradicciones.md)
      // registra el conflicto entre esta uniformidad y el texto literal
      // de CA-011-03.
      // El estado rechazado conserva las seis ranuras llenas y en error,
      // como el contrato visual de los eventos 06/10–12. La uniformidad de
      // seguridad reside en el mensaje y el resultado, no en borrar el
      // valor que la persona acaba de revisar.
      verifyStatus.value = 'invalid-code'
      return
    case 'validation-error':
      verifyStatus.value = 'invalid-code'
      return
    case 'network-error':
      verifyStatus.value = 'network-error'
      return
    case 'unexpected-error':
      verifyStatus.value = 'unexpected-error'
  }
}
</script>

<template>
  <form class="recovery-verify" novalidate @submit.prevent="onSubmit">
    <p class="recovery-verify__sent" role="status">
      Si tu cuenta existe, recibirás un código de 6 dígitos por WhatsApp y correo.
    </p>

    <OtpInput
      :model-value="code"
      label="Código de 6 dígitos"
      :disabled="isVerifying"
      :error="otpError"
      @update:model-value="handleCodeInput"
    />

    <div class="recovery-verify__actions">
      <BaseButton
        type="submit"
        variant="primary"
        size="lg"
        :loading="isVerifying"
        :disabled="isVerifying"
        class="recovery-verify__submit"
      >
        Verificar código
      </BaseButton>

      <BaseButton
        type="button"
        variant="secondary"
        size="md"
        :disabled="!canResend"
        @click="onResend"
      >
        {{ canResend ? 'Reenviar código' : `Reenviar en ${resendCooldownRemaining} s` }}
      </BaseButton>
    </div>

    <p class="recovery-back">
      <RouterLink :to="{ name: 'acceso' }">Volver al acceso</RouterLink>
    </p>

    <!-- Ancla de alertas: después del grupo de acciones y de "Volver al
         acceso" (trabajo requerido §3, auth-eventos/README.md). -->
    <BaseAlert
      v-if="verifyStatus === 'network-error'"
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
      v-if="verifyStatus === 'unexpected-error'"
      variant="danger"
      title="Ocurrió un error inesperado"
      role="alert"
    >
      Inténtalo de nuevo en unos segundos.
    </BaseAlert>
  </form>
</template>

<style scoped>
.recovery-verify {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  width: 100%;
}

.recovery-verify__sent {
  margin: 0;
  font-size: 18px;
  line-height: 26px;
  color: var(--color-text-secondary);
}

.recovery-verify__actions {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.recovery-verify__submit {
  width: 100%;
  min-height: 56px;
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
