<script setup lang="ts">
// Paso 1/3 de HU-011: solicitar el código de recuperación eligiendo el
// canal (DEC-092). Dos botones —Teléfono y Correo—; hasta elegir uno no hay
// campo ni envío. Al elegir, un mensaje menciona únicamente ese canal y se
// pide su valor (número o correo), que solo identifica la cuenta: el código
// llega al contacto verificado que el servidor ya tiene (DEC-093).
// Dueño de su propio envío (mismo patrón que `PhoneChallengeForm.vue`):
// valida en cliente, llama al API real y decide cuándo avanzar al paso 2. Un
// 202 (DEC-065, no enumeración) siempre avanza; un fallo de transporte real
// (red o 5xx) no revela nada sobre la cuenta, así que sí se distingue con
// un error accionable (a diferencia del "siempre avanzar" de
// `PhoneChallengeForm`, que ahí es correcto porque ese reto ya está detrás
// de un intento de acceso autenticable).
import { computed, reactive, ref } from 'vue'
import { BaseButton, BaseInput } from '@/shared/ui'
import { requestRecovery } from '../api/recoveryApi'
import { notifyConnectionLost, notifyUnexpectedError } from '../model/authFeedback'
import type { RecoveryChannel, RecoveryTarget } from '../model/recoveryTarget'
import {
  normalizeRecoveryPhone,
  validateRecoveryEmail,
  validateRecoveryPhone,
} from '../validation/recoveryValidation'

const emit = defineEmits<{
  advance: [payload: { target: RecoveryTarget }]
}>()

const CHANNELS: { value: RecoveryChannel; label: string }[] = [
  // El literal "whatsapp" pertenece al contrato API vigente; la interfaz se
  // mantiene agnóstica al proveedor y lo presenta como teléfono.
  { value: 'whatsapp', label: 'Teléfono' },
  { value: 'email', label: 'Correo' },
]

// Nada elegido al entrar: ni campo ni acción de envío hasta que la persona
// decide el canal (CA-011-09).
const channel = ref<RecoveryChannel | null>(null)
// Lo escrito en cada canal se conserva al alternar entre ellos.
const values = reactive<Record<RecoveryChannel, string>>({ whatsapp: '', email: '' })
const fieldError = ref<string | undefined>(undefined)
const attemptedSubmit = ref(false)
const status = ref<'idle' | 'submitting'>('idle')

const isSubmitting = computed(() => status.value === 'submitting')

function validateCurrent(): string | undefined {
  if (channel.value === 'whatsapp') return validateRecoveryPhone(values.whatsapp)
  if (channel.value === 'email') return validateRecoveryEmail(values.email)
  return undefined
}

function chooseChannel(next: RecoveryChannel) {
  if (isSubmitting.value || channel.value === next) return
  channel.value = next
  // Cambiar de canal descarta el error y el aviso del canal anterior: ya no
  // describen el campo que se ve. El foco se queda en el botón pulsado
  // (CA-011-10); el mensaje nuevo se anuncia por su región `aria-live`.
  attemptedSubmit.value = false
  fieldError.value = undefined
  status.value = 'idle'
}

const handleInput = (value: string | number) => {
  if (!channel.value) return
  values[channel.value] = String(value)
  if (attemptedSubmit.value) {
    fieldError.value = validateCurrent()
  }
}

async function onSubmit() {
  if (isSubmitting.value || !channel.value) return

  attemptedSubmit.value = true
  const error = validateCurrent()
  fieldError.value = error
  if (error) return

  const target: RecoveryTarget =
    channel.value === 'whatsapp'
      ? { channel: 'whatsapp', value: normalizeRecoveryPhone(values.whatsapp) }
      : { channel: 'email', value: values.email.trim() }

  status.value = 'submitting'
  const outcome = await requestRecovery(target)

  switch (outcome.kind) {
    case 'accepted':
      emit('advance', { target })
      return
    case 'network-error':
      status.value = 'idle'
      notifyConnectionLost(() => void onSubmit())
      return
    case 'unexpected-error':
      status.value = 'idle'
      notifyUnexpectedError(outcome.requestId)
  }
}
</script>

<template>
  <form class="recovery-request" novalidate @submit.prevent="onSubmit">
    <div class="recovery-request__group" role="group" aria-labelledby="recovery-channel-label">
      <p id="recovery-channel-label" class="recovery-request__group-label">Recibir código por</p>
      <div class="recovery-request__choices">
        <button
          v-for="option in CHANNELS"
          :key="option.value"
          type="button"
          class="recovery-channel"
          :class="{ 'recovery-channel--selected': channel === option.value }"
          :aria-pressed="channel === option.value"
          :disabled="isSubmitting"
          @click="chooseChannel(option.value)"
        >
          <svg
            v-if="option.value === 'whatsapp'"
            class="recovery-channel__icon"
            width="26"
            height="26"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="1.8"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"
          >
            <rect x="7" y="2.5" width="10" height="19" rx="2" />
            <path d="M10.5 18.5h3" />
          </svg>
          <svg
            v-else
            class="recovery-channel__icon"
            width="26"
            height="26"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="1.8"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"
          >
            <rect x="3" y="5" width="18" height="14" rx="2" />
            <path d="M3.5 6.5l8.5 6.5 8.5-6.5" />
          </svg>
          <span class="recovery-channel__label">{{ option.label }}</span>
          <!-- Señal distinta del color para el canal elegido (CA-011-09). -->
          <svg
            v-if="channel === option.value"
            class="recovery-channel__check"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.2"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"
          >
            <circle class="recovery-channel__check-ring" cx="12" cy="12" r="9.5" />
            <path class="recovery-channel__check-tick" d="M7.8 12.4l3 3 5.4-6" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Región `aria-live`: al elegir un canal el mensaje nuevo sustituye al
         anterior y se anuncia sin mover el foco (CA-011-10). -->
    <div class="recovery-request__message" aria-live="polite">
      <p v-if="channel === null" class="recovery-request__hint">
        Elige por dónde quieres recibir tu código de un solo uso.
      </p>
      <p
        v-else-if="channel === 'whatsapp'"
        class="recovery-request__hint recovery-request__hint--swap"
      >
        Escribe el número de teléfono de tu cuenta. Si existe, te enviaremos un código de un solo
        uso a tu <strong>teléfono</strong>.
      </p>
      <p v-else class="recovery-request__hint recovery-request__hint--swap">
        Escribe el correo de tu cuenta. Si existe, te enviaremos un código de un solo uso por
        <strong>correo</strong>.
      </p>
    </div>

    <!-- Al elegir canal el campo y el botón se despliegan (la altura crece
         desde cero) en vez de aparecer de golpe y empujar el enlace de abajo. -->
    <Transition name="recovery-reveal">
      <div v-if="channel !== null" class="recovery-reveal">
        <div class="recovery-reveal__inner">
          <!-- `key` fuerza un campo nuevo por canal: cambia tipo, etiqueta y
               autocompletado sin arrastrar el estado del anterior. -->
          <BaseInput
            v-if="channel === 'whatsapp'"
            key="whatsapp"
            class="recovery-request__field"
            :model-value="values.whatsapp"
            type="tel"
            name="whatsapp"
            label="Teléfono"
            autocomplete="tel"
            placeholder="+573001234567"
            required
            :show-required-marker="false"
            :disabled="isSubmitting"
            :error="fieldError"
            @update:model-value="handleInput"
          />
          <BaseInput
            v-else
            key="email"
            class="recovery-request__field"
            :model-value="values.email"
            type="email"
            name="email"
            label="Correo"
            autocomplete="username"
            placeholder="tu-correo@ejemplo.com"
            required
            :show-required-marker="false"
            :disabled="isSubmitting"
            :error="fieldError"
            @update:model-value="handleInput"
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
        </div>
      </div>
    </Transition>

    <p class="recovery-back">
      <RouterLink :to="{ name: 'acceso' }">Volver al acceso</RouterLink>
    </p>
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

.recovery-request__group {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

/* Versalitas de latón, mismo tratamiento que los rótulos de campo. */
.recovery-request__group-label {
  margin: 0;
  font-family: var(--font-sans);
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-accent-brass);
}

.recovery-request__choices {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-3);
}

.recovery-channel {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  min-height: 64px;
  padding: 0 var(--space-4);
  font-family: inherit;
  font-size: 18px;
  font-weight: 600;
  color: var(--color-action-primary);
  cursor: pointer;
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-control);
  border-radius: var(--radius-sm);
}

.recovery-channel:focus-visible {
  outline: var(--border-width-emphasis) solid var(--color-focus);
  outline-offset: 3px;
}

.recovery-channel:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.recovery-channel--selected {
  color: var(--color-on-strong);
  background-color: var(--color-action-primary);
  border: var(--border-width-emphasis) solid var(--color-action-primary);
}

.recovery-channel__icon,
.recovery-channel__check {
  flex-shrink: 0;
}

.recovery-channel__label {
  flex-grow: 1;
  text-align: left;
}

/* Por debajo de 480 px el rótulo, el icono y el check no caben en una fila
   sin recortar «Teléfono» (320 px): el icono pasa sobre el rótulo y el check
   sale del flujo a la esquina, así el botón conserva un área táctil amplia y
   el texto completo. */
@media (max-width: 479px) {
  .recovery-channel {
    position: relative;
    flex-direction: column;
    justify-content: center;
    gap: var(--space-1);
    min-height: 84px;
    padding: var(--space-3) var(--space-2);
    font-size: 16px;
  }

  .recovery-channel__label {
    flex-grow: 0;
    text-align: center;
  }

  .recovery-channel__check {
    position: absolute;
    top: var(--space-2);
    right: var(--space-2);
  }
}

.recovery-request__hint {
  margin: 0;
  font-size: 20px;
  line-height: 28px;
  color: var(--color-text-secondary);
}

.recovery-request__hint strong {
  font-weight: 600;
  color: var(--color-text-primary);
}

@media (min-width: 1024px) {
  .recovery-request__hint {
    font-size: 18px;
    line-height: 26px;
  }

  .recovery-channel {
    min-height: 72px;
  }
}

/* Selector de canal: el relleno de tinta se ilumina en vez de saltar, el botón
   se alza al pasar y se hunde al pulsar, y la marca de elegido se dibuja
   (aro y visto) con un pequeño rebote. */
.recovery-channel {
  transition:
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard),
    color var(--motion-duration-base) var(--motion-easing-standard),
    transform var(--motion-duration-base) cubic-bezier(0.2, 0.7, 0.2, 1),
    box-shadow var(--motion-duration-base) var(--motion-easing-standard);
}

.recovery-channel:hover:not(:disabled):not(.recovery-channel--selected) {
  border-color: var(--color-accent-brass);
  box-shadow: 0 6px 14px -8px color-mix(in srgb, var(--color-action-primary) 45%, transparent);
  transform: translateY(-2px);
}

.recovery-channel:active:not(:disabled) {
  transform: translateY(0) scale(0.985);
}

.recovery-channel__icon {
  transition: transform 320ms cubic-bezier(0.3, 1.4, 0.5, 1);
}

.recovery-channel:hover:not(:disabled) .recovery-channel__icon,
.recovery-channel--selected .recovery-channel__icon {
  transform: scale(1.1) rotate(-4deg);
}

.recovery-channel__check {
  animation: recovery-check-pop 0.5s cubic-bezier(0.3, 1.4, 0.5, 1) backwards;
}

.recovery-channel__check-ring {
  stroke-dasharray: 60;
  animation: recovery-check-draw 0.45s cubic-bezier(0.65, 0, 0.2, 1) backwards;
}

.recovery-channel__check-tick {
  stroke-dasharray: 18;
  animation: recovery-check-draw 0.35s cubic-bezier(0.65, 0, 0.2, 1) 0.2s backwards;
}

@keyframes recovery-check-pop {
  from {
    opacity: 0;
    transform: scale(0.4) rotate(-40deg);
  }
}

@keyframes recovery-check-draw {
  from {
    stroke-dashoffset: 60;
  }
}

/* El mensaje del canal elegido sustituye al anterior subiendo con suavidad. */
.recovery-request__hint--swap,
.recovery-request__field {
  animation: recovery-hint-in 0.45s cubic-bezier(0.2, 0.7, 0.2, 1) backwards;
}

@keyframes recovery-hint-in {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
}

/* Despliegue del campo y el botón: la fila de la rejilla pasa de 0fr a 1fr. El
   margen negativo y el relleno superior reservan el hueco del `gap` del
   formulario dentro del propio bloque, para que el enlace de abajo no salte
   al terminar. El recorte solo existe mientras dura el movimiento, así el aro
   de foco del campo no queda cortado en reposo. */
.recovery-reveal {
  display: grid;
  grid-template-rows: 1fr;
  margin-top: calc(-1 * var(--space-8));
}

.recovery-reveal__inner {
  display: flex;
  flex-direction: column;
  gap: var(--space-8);
  min-height: 0;
  padding-top: var(--space-8);
}

.recovery-reveal-enter-active,
.recovery-reveal-leave-active {
  transition:
    grid-template-rows 360ms cubic-bezier(0.2, 0.7, 0.2, 1),
    opacity 280ms var(--motion-easing-standard);
}

.recovery-reveal-enter-active > .recovery-reveal__inner,
.recovery-reveal-leave-active > .recovery-reveal__inner {
  overflow: hidden;
}

.recovery-reveal-enter-from,
.recovery-reveal-leave-to {
  grid-template-rows: 0fr;
  opacity: 0;
}

/* Cada campo del bloque entra con un pequeño retraso, como la cascada de la
   tarjeta de acceso. */
.recovery-reveal-enter-active .recovery-reveal__inner > * {
  animation: recovery-hint-in 0.55s cubic-bezier(0.2, 0.7, 0.2, 1) backwards;
}

.recovery-reveal-enter-active .recovery-reveal__inner > :nth-child(2) {
  animation-delay: 0.1s;
}

@media (prefers-reduced-motion: reduce) {
  .recovery-channel,
  .recovery-channel__icon,
  .recovery-reveal-enter-active,
  .recovery-reveal-leave-active {
    transition: none;
  }

  .recovery-channel:hover:not(:disabled),
  .recovery-channel:active:not(:disabled) {
    transform: none;
  }

  .recovery-channel__check,
  .recovery-channel__check-ring,
  .recovery-channel__check-tick,
  .recovery-request__hint--swap,
  .recovery-request__field,
  .recovery-reveal-enter-active .recovery-reveal__inner > * {
    animation: none;
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
