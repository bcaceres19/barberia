<script setup lang="ts">
// Paso 1/3 de HU-011: solicitar el código de recuperación eligiendo el
// canal (DEC-092). Dos botones —WhatsApp y Correo—; hasta elegir uno no hay
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
import { BaseAlert, BaseButton, BaseInput } from '@/shared/ui'
import { requestRecovery } from '../api/recoveryApi'
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
  { value: 'whatsapp', label: 'WhatsApp' },
  { value: 'email', label: 'Correo' },
]

// Nada elegido al entrar: ni campo ni acción de envío hasta que la persona
// decide el canal (CA-011-09).
const channel = ref<RecoveryChannel | null>(null)
// Lo escrito en cada canal se conserva al alternar entre ellos.
const values = reactive<Record<RecoveryChannel, string>>({ whatsapp: '', email: '' })
const fieldError = ref<string | undefined>(undefined)
const attemptedSubmit = ref(false)
const status = ref<'idle' | 'submitting' | 'network-error' | 'unexpected-error'>('idle')

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
      status.value = 'network-error'
      return
    case 'unexpected-error':
      status.value = 'unexpected-error'
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
            <path d="M3.5 20.5l1.3-4.3A8.5 8.5 0 1 1 8 19.3l-4.5 1.2z" />
            <path
              d="M9 8.6c.2 2.9 2.5 5.2 5.4 5.4l1.3-1.4-2.1-1-.9.7a3.6 3.6 0 0 1-1.6-1.6l.7-.9-1-2.1L9 8.6z"
            />
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
            <circle cx="12" cy="12" r="9.5" />
            <path d="M7.8 12.4l3 3 5.4-6" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Región `aria-live`: al elegir un canal el mensaje nuevo sustituye al
         anterior y se anuncia sin mover el foco (CA-011-10). Menciona solo
         el canal elegido (DEC-092). -->
    <div class="recovery-request__message" aria-live="polite">
      <p v-if="channel === null" class="recovery-request__hint">
        Elige por dónde quieres recibir tu código de un solo uso.
      </p>
      <p v-else-if="channel === 'whatsapp'" class="recovery-request__hint">
        Escribe el número de WhatsApp de tu cuenta. Si existe, te enviaremos un código de un solo
        uso por <strong>WhatsApp</strong>.
      </p>
      <p v-else class="recovery-request__hint">
        Escribe el correo de tu cuenta. Si existe, te enviaremos un código de un solo uso por
        <strong>correo</strong>.
      </p>
    </div>

    <template v-if="channel !== null">
      <!-- `key` fuerza un campo nuevo por canal: cambia tipo, etiqueta y
           autocompletado sin arrastrar el estado del anterior. -->
      <BaseInput
        v-if="channel === 'whatsapp'"
        key="whatsapp"
        :model-value="values.whatsapp"
        type="tel"
        name="whatsapp"
        label="WhatsApp"
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
    </template>

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
   sin recortar «WhatsApp» (320 px): el icono pasa sobre el rótulo y el check
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
