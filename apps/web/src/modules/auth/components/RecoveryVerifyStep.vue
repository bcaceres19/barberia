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
import { BaseButton, OtpInput } from '@/shared/ui'
import { requestRecovery, verifyRecovery } from '../api/recoveryApi'
import { notifyConnectionLost, notifyUnexpectedError } from '../model/authFeedback'
import type { RecoveryTarget } from '../model/recoveryTarget'
import { validateRecoveryCode } from '../validation/recoveryValidation'

const RESEND_COOLDOWN_SECONDS = 60

// Mismo canal y valor del paso 1 (DEC-093, DP-SEG-15).
const props = defineProps<{
  target: RecoveryTarget
}>()

const emit = defineEmits<{
  advance: [payload: { resetToken: string; maskedPhone: string; maskedEmail: string }]
  // Volver al paso 1 para elegir otro canal (DP-SEG-16, DEC-093). Aún no hay
  // nada verificado que perder: es lo único que se ofrece cuando el código
  // no llega, sin decir por qué (no enumeración, DEC-065).
  'change-channel': []
}>()

// Solo el canal elegido en el paso 1 (DEC-092): el otro no recibe nada.
const channelLabel = computed(() => (props.target.channel === 'whatsapp' ? 'teléfono' : 'correo'))

const code = ref('')
const fieldError = ref<string | undefined>(undefined)
const attemptedSubmit = ref(false)
const verifyStatus = ref<'idle' | 'verifying' | 'invalid-code'>('idle')
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
  await requestRecovery(props.target)
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
  const outcome = await verifyRecovery(props.target, code.value)

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
      verifyStatus.value = 'idle'
      notifyConnectionLost(() => void onSubmit())
      return
    case 'unexpected-error':
      verifyStatus.value = 'idle'
      notifyUnexpectedError(outcome.requestId)
  }
}
</script>

<template>
  <form class="recovery-verify" novalidate @submit.prevent="onSubmit">
    <p class="recovery-verify__sent" role="status">
      Si tu cuenta existe, recibirás un código de 6 dígitos por {{ channelLabel }}.
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
        class="recovery-verify__resend"
        :class="{ 'recovery-verify__resend--ready': canResend }"
        :disabled="!canResend"
        @click="onResend"
      >
        {{ canResend ? 'Reenviar código' : `Reenviar en ${resendCooldownRemaining} s` }}
      </BaseButton>
    </div>

    <p class="recovery-back">
      <button type="button" class="recovery-back__link" @click="emit('change-channel')">
        ¿No te llegó? Elegir otro canal
      </button>
    </p>

    <p class="recovery-back">
      <RouterLink :to="{ name: 'acceso' }">Volver al acceso</RouterLink>
    </p>
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

/* Las seis casillas del código se componen una a una de izquierda a derecha,
   como las marcas de una regla. Un dígito escrito se asienta con un pequeño golpe y un código
   rechazado sacude las casillas; ambos movimientos se disparan al cambiar de
   clase, así que no se repiten mientras la persona sigue escribiendo. */
.recovery-verify :deep(.otp-input__slot) {
  animation: recovery-slot-in 0.5s cubic-bezier(0.2, 0.7, 0.2, 1) backwards;
}

.recovery-verify :deep(.otp-input__slot:nth-child(1)) {
  animation-delay: 0.12s;
}

.recovery-verify :deep(.otp-input__slot:nth-child(2)) {
  animation-delay: 0.18s;
}

.recovery-verify :deep(.otp-input__slot:nth-child(3)) {
  animation-delay: 0.24s;
}

.recovery-verify :deep(.otp-input__slot:nth-child(4)) {
  animation-delay: 0.3s;
}

.recovery-verify :deep(.otp-input__slot:nth-child(5)) {
  animation-delay: 0.36s;
}

.recovery-verify :deep(.otp-input__slot:nth-child(6)) {
  animation-delay: 0.42s;
}

.recovery-verify :deep(.otp-input__slot--filled) {
  animation: recovery-slot-fill 0.28s cubic-bezier(0.3, 1.4, 0.5, 1);
}

.recovery-verify :deep(.otp-input__slot--invalid) {
  animation: recovery-slot-shake 0.42s cubic-bezier(0.36, 0.07, 0.19, 0.97);
}

.recovery-verify :deep(.otp-input__error) {
  animation: recovery-slot-in 0.35s cubic-bezier(0.2, 0.7, 0.2, 1) backwards;
}

/* Al terminar la cuenta atrás el botón de reenvío avisa una vez con un aro de
   latón que se expande; no se repite ni compite con el foco. */
.recovery-verify__resend--ready {
  animation: recovery-ready-ring 0.9s ease-out 1;
}

@keyframes recovery-slot-in {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
}

@keyframes recovery-slot-fill {
  40% {
    transform: translateY(-3px) scale(1.06);
  }
}

@keyframes recovery-slot-shake {
  20%,
  60% {
    transform: translateX(-5px);
  }
  40%,
  80% {
    transform: translateX(5px);
  }
}

@keyframes recovery-ready-ring {
  from {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--color-accent-brass) 55%, transparent);
  }
  to {
    box-shadow: 0 0 0 12px transparent;
  }
}

@media (prefers-reduced-motion: reduce) {
  .recovery-verify :deep(.otp-input__slot),
  .recovery-verify :deep(.otp-input__slot--filled),
  .recovery-verify :deep(.otp-input__slot--invalid),
  .recovery-verify :deep(.otp-input__error),
  .recovery-verify__resend--ready {
    animation: none;
  }
}

.recovery-back {
  margin: 0;
  text-align: center;
  font-size: var(--font-size-body-sm);
}

.recovery-back__link {
  min-height: 44px;
  padding: 0 var(--space-2);
  font: inherit;
  font-weight: 500;
  color: var(--color-action-primary);
  text-decoration: underline;
  cursor: pointer;
  background: none;
  border: 0;
}

.recovery-back__link:focus-visible {
  outline: var(--border-width-emphasis) solid var(--color-focus);
  outline-offset: 2px;
}

.recovery-back a {
  color: var(--color-action-primary);
  font-weight: 500;
}
</style>
