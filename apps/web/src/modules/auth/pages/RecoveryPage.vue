<script setup lang="ts">
// Página coordinadora de HU-011 (P0 "Recuperación",
// estandar-diseno-visual.md §10): posee el estado y la navegación entre
// los tres pasos; cada `Recovery*Step.vue` solo presenta su paso y llama
// al API real. Reemplaza el destino provisional de `DP-UX-06`
// (`RecoveryPendingPage.vue`, ya retirada) ahora que el flujo real existe.
//
// Sin retroceso entre pasos completados (trabajo requerido §2): una vez
// verificado el código, no hay botón para volver al paso 1/2 y perder el
// progreso que el servidor ya confirmó. El único camino hacia atrás es
// `restart`, que solo ocurre cuando el paso 3 informa que el token ya no
// es válido (no queda nada que conservar).
import { nextTick, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { BaseAlert, BaseButton } from '@/shared/ui'
import RecoveryRequestStep from '../components/RecoveryRequestStep.vue'
import RecoveryVerifyStep from '../components/RecoveryVerifyStep.vue'
import RecoveryResetStep from '../components/RecoveryResetStep.vue'

type Step = 'request' | 'verify' | 'reset' | 'done'

const STEP_NUMBERS: Record<Step, number> = { request: 1, verify: 2, reset: 3, done: 3 }
const STEP_TITLES: Record<Step, string> = {
  request: 'Solicita tu código',
  verify: 'Verifica el código',
  reset: 'Establece tu contraseña nueva',
  done: 'Contraseña actualizada',
}

const router = useRouter()

const step = ref<Step>('request')
const email = ref('')
const resetToken = ref('')
const maskedPhone = ref('')
const maskedEmail = ref('')

const headingRef = ref<HTMLHeadingElement | null>(null)

// Foco al encabezado del paso en cada transición, anunciado a lector de
// pantalla (CA-011-07): un cambio de paso es equivalente a una navegación,
// no a una simple actualización de contenido dentro del mismo paso.
watch(step, async () => {
  await nextTick()
  headingRef.value?.focus()
})

function onRequestAdvance(payload: { email: string }) {
  email.value = payload.email
  step.value = 'verify'
}

function onVerifyAdvance(payload: {
  resetToken: string
  maskedPhone: string
  maskedEmail: string
}) {
  resetToken.value = payload.resetToken
  maskedPhone.value = payload.maskedPhone
  maskedEmail.value = payload.maskedEmail
  step.value = 'reset'
}

function onResetDone() {
  step.value = 'done'
}

function onRestart() {
  resetToken.value = ''
  maskedPhone.value = ''
  maskedEmail.value = ''
  step.value = 'request'
}
</script>

<template>
  <main class="recovery-page">
    <div class="recovery-page__card">
      <p class="recovery-page__brand">Barbería</p>
      <p class="recovery-page__progress">Paso {{ STEP_NUMBERS[step] }} de 3</p>
      <h1 ref="headingRef" class="recovery-page__title" tabindex="-1">{{ STEP_TITLES[step] }}</h1>

      <RecoveryRequestStep v-if="step === 'request'" @advance="onRequestAdvance" />

      <RecoveryVerifyStep v-else-if="step === 'verify'" :email="email" @advance="onVerifyAdvance" />

      <RecoveryResetStep
        v-else-if="step === 'reset'"
        :email="email"
        :reset-token="resetToken"
        :masked-phone="maskedPhone"
        :masked-email="maskedEmail"
        @done="onResetDone"
        @restart="onRestart"
      />

      <template v-else>
        <BaseAlert variant="success" title="Listo" role="status">
          Actualizamos tu contraseña y cerramos todas tus sesiones activas. Inicia sesión de nuevo
          con la contraseña nueva.
        </BaseAlert>
        <BaseButton
          type="button"
          variant="primary"
          size="lg"
          class="recovery-page__done-action"
          @click="router.push({ name: 'acceso' })"
        >
          Ir al acceso
        </BaseButton>
      </template>

      <p v-if="step !== 'done'" class="recovery-page__back">
        <RouterLink :to="{ name: 'acceso' }">Volver al acceso</RouterLink>
      </p>
    </div>
  </main>
</template>

<style scoped>
.recovery-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100dvh;
  padding: var(--space-6) var(--space-4);
}

.recovery-page__card {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  width: 100%;
  max-width: 360px;
}

.recovery-page__brand {
  margin: 0;
  font-size: var(--font-size-body-sm);
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.recovery-page__progress {
  margin: 0;
  font-size: var(--font-size-body-sm);
  color: var(--color-text-secondary);
}

.recovery-page__title {
  margin: 0 0 var(--space-2) 0;
  font-size: var(--font-size-h2);
  line-height: var(--font-size-h2-line);
  color: var(--color-text-primary);
}

.recovery-page__title:focus-visible {
  outline: none;
}

.recovery-page__done-action {
  width: 100%;
  text-align: center;
}

.recovery-page__back {
  margin: 0;
  text-align: center;
  font-size: var(--font-size-body-sm);
}

.recovery-page__back a {
  color: var(--color-action-primary);
  font-weight: 500;
}
</style>
