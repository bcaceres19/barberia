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
import AuthSplitLayout from '../components/AuthSplitLayout.vue'
import RecoveryRequestStep from '../components/RecoveryRequestStep.vue'
import RecoveryVerifyStep from '../components/RecoveryVerifyStep.vue'
import RecoveryResetStep from '../components/RecoveryResetStep.vue'
import type { RecoveryTarget } from '../model/recoveryTarget'

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
// Canal y valor elegidos en el paso 1; los pasos 2 y 3 los reutilizan
// (DEC-093, DP-SEG-15).
const target = ref<RecoveryTarget | null>(null)
const resetToken = ref('')
const maskedPhone = ref('')
const maskedEmail = ref('')

const headingRef = ref<HTMLHeadingElement | null>(null)

// Dirección del movimiento entre pasos: avanzar entra desde la derecha,
// volver (otro canal, solicitar de nuevo) desde la izquierda. El título solo
// anima al cambiar de paso, no al abrir la pantalla, donde ya manda la entrada
// escalonada de la tarjeta.
const direction = ref<'fwd' | 'back'>('fwd')
const stepChanged = ref(false)

// Foco al encabezado del paso en cada transición, anunciado a lector de
// pantalla (CA-011-07): un cambio de paso es equivalente a una navegación,
// no a una simple actualización de contenido dentro del mismo paso.
watch(step, async () => {
  stepChanged.value = true
  await nextTick()
  headingRef.value?.focus()
})

function onRequestAdvance(payload: { target: RecoveryTarget }) {
  direction.value = 'fwd'
  target.value = payload.target
  step.value = 'verify'
}

function onVerifyAdvance(payload: {
  resetToken: string
  maskedPhone: string
  maskedEmail: string
}) {
  direction.value = 'fwd'
  resetToken.value = payload.resetToken
  maskedPhone.value = payload.maskedPhone
  maskedEmail.value = payload.maskedEmail
  step.value = 'reset'
}

function onResetDone() {
  direction.value = 'fwd'
  step.value = 'done'
}

// Desde el paso 2 (aún sin código verificado) se puede volver a elegir canal:
// no hay progreso confirmado por el servidor que conservar.
function onChangeChannel() {
  direction.value = 'back'
  target.value = null
  step.value = 'request'
}

function onRestart() {
  direction.value = 'back'
  target.value = null
  resetToken.value = ''
  maskedPhone.value = ''
  maskedEmail.value = ''
  step.value = 'request'
}
</script>

<template>
  <AuthSplitLayout class="recovery-page" caption="RECUPERACIÓN SEGURA">
    <div class="recovery-page__header">
      <p
        :key="`progress-${step}`"
        class="recovery-page__progress"
        :class="{ 'recovery-page__swap': stepChanged }"
      >
        Paso {{ STEP_NUMBERS[step] }} de 3
      </p>
      <h1
        :key="`title-${step}`"
        ref="headingRef"
        class="recovery-page__title"
        :class="{ 'recovery-page__swap recovery-page__swap--title': stepChanged }"
        tabindex="-1"
      >
        {{ STEP_TITLES[step] }}
      </h1>
    </div>

    <!-- Un paso a la vez: el saliente se desvanece rápido y el entrante llega
         desde el lado al que se avanza (o desde el contrario al volver). -->
    <Transition :name="`recovery-step-${direction}`" mode="out-in">
      <RecoveryRequestStep v-if="step === 'request'" @advance="onRequestAdvance" />

      <RecoveryVerifyStep
        v-else-if="step === 'verify' && target"
        :target="target"
        @advance="onVerifyAdvance"
        @change-channel="onChangeChannel"
      />

      <RecoveryResetStep
        v-else-if="step === 'reset' && target"
        :target="target"
        :reset-token="resetToken"
        :masked-phone="maskedPhone"
        :masked-email="maskedEmail"
        @done="onResetDone"
        @restart="onRestart"
      />

      <div v-else class="recovery-page__done">
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
      </div>
    </Transition>
  </AuthSplitLayout>
</template>

<style scoped>
.recovery-page__header {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

/* Versalitas de latón, mismo tratamiento reglado que "ACCESO SEGURO" y
   los rótulos de campo (contrato visual, issue #212/#213). */
.recovery-page__progress {
  margin: 0;
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  text-align: center;
  color: var(--color-accent-brass);
}

.recovery-page__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: 52px;
  line-height: 60px;
  text-align: center;
  color: var(--color-text-primary);
}

@media (min-width: 1024px) {
  .recovery-page__header {
    gap: var(--space-2);
  }

  .recovery-page__title {
    font-size: 36px;
    line-height: 42px;
  }
}

.recovery-page__title:focus-visible {
  outline: none;
}

.recovery-page__done {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.recovery-page__done-action {
  width: 100%;
  text-align: center;
  animation: recovery-rise 0.6s cubic-bezier(0.2, 0.7, 0.2, 1) 0.25s backwards;
}

/* Movimiento de la recuperación: mismo lenguaje que la entrada de acceso
   (subida corta con la curva de la marca, filete y reflejo de latón). Todo
   termina en el estado normal del elemento, así que sin animación (reduced
   motion) la pantalla queda ya compuesta. */
.recovery-page__swap {
  animation: recovery-rise 0.45s cubic-bezier(0.2, 0.7, 0.2, 1) backwards;
}

.recovery-page__swap--title {
  animation-delay: 0.06s;
}

.recovery-step-fwd-leave-active,
.recovery-step-back-leave-active {
  transition:
    opacity 140ms ease-in,
    transform 140ms ease-in;
}

.recovery-step-fwd-enter-active,
.recovery-step-back-enter-active {
  transition:
    opacity 380ms cubic-bezier(0.2, 0.7, 0.2, 1),
    transform 380ms cubic-bezier(0.2, 0.7, 0.2, 1);
}

.recovery-step-fwd-enter-from {
  opacity: 0;
  transform: translateX(18px);
}

.recovery-step-fwd-leave-to {
  opacity: 0;
  transform: translateX(-18px);
}

.recovery-step-back-enter-from {
  opacity: 0;
  transform: translateX(-18px);
}

.recovery-step-back-leave-to {
  opacity: 0;
  transform: translateX(18px);
}

/* Campos y botón principal de los tres pasos: el rótulo pasa de latón a tinta
   y un filete de latón recorre la base al enfocar; el botón recibe un reflejo
   de latón al pasar o enfocar. Es el mismo tratamiento de `LoginForm`. */
.recovery-page :deep(.base-input__label) {
  transition: color var(--motion-duration-base) var(--motion-easing-standard);
}

.recovery-page :deep(.base-input__wrapper:focus-within .base-input__label) {
  color: var(--color-action-primary);
}

.recovery-page :deep(.base-input__input-wrapper)::after {
  content: '';
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: var(--border-width-emphasis);
  background-color: var(--color-brand-accent-surface);
  border-radius: 0 0 var(--radius-sm) var(--radius-sm);
  transform: scaleX(0);
  transform-origin: left;
  transition: transform 320ms cubic-bezier(0.2, 0.7, 0.2, 1);
  pointer-events: none;
}

.recovery-page :deep(.base-input__input-wrapper:focus-within)::after {
  transform: scaleX(1);
}

.recovery-page :deep(.base-button--primary) {
  position: relative;
  overflow: hidden;
  isolation: isolate;
}

.recovery-page :deep(.base-button--primary)::after {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  left: -40%;
  z-index: -1;
  width: 32%;
  background: linear-gradient(
    100deg,
    transparent,
    color-mix(in srgb, var(--color-brand-accent-surface) 38%, transparent),
    transparent
  );
  transform: translateX(0) skewX(-18deg);
  transition: transform 700ms cubic-bezier(0.2, 0.7, 0.2, 1);
  pointer-events: none;
}

.recovery-page :deep(.base-button--primary:hover:not(:disabled))::after,
.recovery-page :deep(.base-button--primary:focus-visible)::after {
  transform: translateX(460%) skewX(-18deg);
}

@keyframes recovery-rise {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
}

@media (prefers-reduced-motion: reduce) {
  .recovery-page__swap,
  .recovery-page__done-action {
    animation: none;
  }

  .recovery-step-fwd-enter-active,
  .recovery-step-fwd-leave-active,
  .recovery-step-back-enter-active,
  .recovery-step-back-leave-active,
  .recovery-page :deep(.base-input__label),
  .recovery-page :deep(.base-input__input-wrapper)::after,
  .recovery-page :deep(.base-button--primary)::after {
    transition: none;
  }
}
</style>
