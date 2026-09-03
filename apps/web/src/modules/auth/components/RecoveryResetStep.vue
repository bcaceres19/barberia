<script setup lang="ts">
// Paso 3/3 de HU-011: establecer la contraseña nueva. Confirma el destino
// enmascarado que devolvió la verificación (CA-008-06/CA-011-02: única vez
// que aparece, y llegar aquí ya exigió el código real) y explica la
// política de `DEC-063` ANTES del campo (CA-011-06), no solo al fallar.
import { computed, ref } from 'vue'
import { BaseAlert, BaseButton, BaseInput } from '@/shared/ui'
import { resetRecoveryPassword } from '../api/recoveryApi'
import { validateNewPassword } from '../validation/recoveryValidation'

const props = defineProps<{
  email: string
  resetToken: string
  maskedPhone: string
  maskedEmail: string
}>()

const emit = defineEmits<{
  done: []
  restart: []
}>()

const newPassword = ref('')
const confirmPassword = ref('')
const fieldErrors = ref<{ newPassword?: string; confirmPassword?: string }>({})
const attemptedSubmit = ref(false)
const status = ref<
  | 'idle'
  | 'submitting'
  | 'invalid-token'
  | 'policy-violation'
  | 'network-error'
  | 'unexpected-error'
>('idle')

const isSubmitting = computed(() => status.value === 'submitting')

function runValidation() {
  const errors: { newPassword?: string; confirmPassword?: string } = {}
  const passwordError = validateNewPassword(newPassword.value, props.email)
  if (passwordError) errors.newPassword = passwordError
  if (!errors.newPassword && confirmPassword.value !== newPassword.value) {
    errors.confirmPassword = 'Las dos contraseñas no coinciden.'
  }
  return errors
}

const handleNewPasswordInput = (value: string | number) => {
  newPassword.value = String(value)
  if (attemptedSubmit.value) fieldErrors.value = runValidation()
}
const handleConfirmPasswordInput = (value: string | number) => {
  confirmPassword.value = String(value)
  if (attemptedSubmit.value) fieldErrors.value = runValidation()
}

async function onSubmit() {
  if (isSubmitting.value) return

  attemptedSubmit.value = true
  const errors = runValidation()
  fieldErrors.value = errors
  if (Object.keys(errors).length > 0) return

  status.value = 'submitting'
  const outcome = await resetRecoveryPassword(props.email, props.resetToken, newPassword.value)

  switch (outcome.kind) {
    case 'success':
      emit('done')
      return
    case 'invalid-token':
      status.value = 'invalid-token'
      return
    case 'policy-violation':
    case 'validation-error':
      status.value = 'policy-violation'
      return
    case 'network-error':
      status.value = 'network-error'
      return
    case 'unexpected-error':
      status.value = 'unexpected-error'
  }
}

const onRestart = () => emit('restart')
</script>

<template>
  <form class="recovery-reset" novalidate @submit.prevent="onSubmit">
    <p class="recovery-reset__confirmed" role="status">
      Verificamos tu código, enviado a {{ maskedPhone }} y {{ maskedEmail }}.
    </p>

    <BaseAlert
      v-if="status === 'invalid-token'"
      variant="danger"
      title="El enlace de recuperación venció"
      role="alert"
    >
      Este paso ya no es válido. Solicita un código nuevo para continuar.
      <template #action>
        <BaseButton variant="secondary" size="md" type="button" @click="onRestart">
          Solicitar de nuevo
        </BaseButton>
      </template>
    </BaseAlert>
    <BaseAlert
      v-if="status === 'policy-violation'"
      variant="danger"
      title="La contraseña no cumple la política"
      role="alert"
    >
      Revisa los requisitos e inténtalo de nuevo.
    </BaseAlert>
    <BaseAlert
      v-if="status === 'network-error'"
      variant="warning"
      title="No pudimos conectar"
      role="alert"
    >
      Revisa tu conexión e inténtalo de nuevo.
    </BaseAlert>
    <BaseAlert
      v-if="status === 'unexpected-error'"
      variant="danger"
      title="Ocurrió un error inesperado"
      role="alert"
    >
      Inténtalo de nuevo en unos segundos.
    </BaseAlert>

    <BaseAlert
      v-if="status !== 'invalid-token'"
      variant="plain"
      title="Requisitos de la contraseña"
    >
      Debe tener entre 10 y 128 caracteres, y ser diferente de tu correo y de tu contraseña actual.
    </BaseAlert>

    <template v-if="status !== 'invalid-token'">
      <BaseInput
        :model-value="newPassword"
        type="password"
        name="newPassword"
        label="Contraseña nueva"
        autocomplete="new-password"
        required
        :disabled="isSubmitting"
        :error="fieldErrors.newPassword"
        @update:model-value="handleNewPasswordInput"
      />

      <BaseInput
        :model-value="confirmPassword"
        type="password"
        name="confirmPassword"
        label="Confirma la contraseña nueva"
        autocomplete="new-password"
        required
        :disabled="isSubmitting"
        :error="fieldErrors.confirmPassword"
        @update:model-value="handleConfirmPasswordInput"
      />

      <BaseButton
        type="submit"
        variant="primary"
        size="lg"
        :loading="isSubmitting"
        :disabled="isSubmitting"
        class="recovery-reset__submit"
      >
        Guardar contraseña nueva
      </BaseButton>
    </template>
  </form>
</template>

<style scoped>
.recovery-reset {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  width: 100%;
}

.recovery-reset__confirmed {
  margin: 0;
  font-size: var(--font-size-body-sm);
  color: var(--color-text-secondary);
}

.recovery-reset__submit {
  width: 100%;
}
</style>
