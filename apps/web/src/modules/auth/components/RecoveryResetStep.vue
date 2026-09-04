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
const errorCount = computed(() => Object.keys(fieldErrors.value).length)
// Resumen visible cuando hay más de un error de campo a la vez (mockup
// 08-contrasena-validacion), mismo umbral que LoginForm (CA-010-05).
const showSummary = computed(() => attemptedSubmit.value && errorCount.value > 1)

function runValidation() {
  const errors: { newPassword?: string; confirmPassword?: string } = {}
  const passwordError = validateNewPassword(newPassword.value, props.email)
  if (passwordError) errors.newPassword = passwordError
  // Ambos chequeos son independientes (longitud/política de la contraseña
  // nueva y coincidencia con su confirmación): el mockup 08 los muestra a
  // la vez, así que uno no suprime al otro.
  if (confirmPassword.value !== newPassword.value) {
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

    <!-- Excepción del ancla de alertas (trabajo requerido §3): la alerta
         explica por qué la pantalla está en este estado y la acción es su
         remedio, así que precede al botón en vez de seguirlo. -->
    <template v-if="status === 'invalid-token'">
      <BaseAlert variant="danger" title="El enlace de recuperación venció" role="alert">
        Este paso ya no es válido. Solicita un código nuevo para continuar.
      </BaseAlert>
      <BaseButton
        type="button"
        variant="primary"
        size="lg"
        class="recovery-reset__submit"
        @click="onRestart"
      >
        Solicitar de nuevo
      </BaseButton>
      <p class="recovery-back">
        <RouterLink :to="{ name: 'acceso' }">Volver al acceso</RouterLink>
      </p>
    </template>

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
        :show-required-marker="false"
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
        :show-required-marker="false"
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
        {{ isSubmitting ? 'Guardando contraseña…' : 'Guardar contraseña nueva' }}
      </BaseButton>

      <p class="recovery-back">
        <RouterLink :to="{ name: 'acceso' }">Volver al acceso</RouterLink>
      </p>

      <!-- Ancla de alertas: después del grupo de acciones y de "Volver al
           acceso" (trabajo requerido §3, auth-eventos/README.md). Resumen
           enumerado (mockup 08-contrasena-validacion), mismo tratamiento
           que LoginForm cuando hay más de un error de campo. -->
      <BaseAlert v-if="showSummary" variant="danger" title="Revisa estos campos" role="alert">
        <ul class="recovery-reset__summary-list">
          <li v-if="fieldErrors.newPassword">{{ fieldErrors.newPassword }}</li>
          <li v-if="fieldErrors.confirmPassword">{{ fieldErrors.confirmPassword }}</li>
        </ul>
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
  font-size: 18px;
  line-height: 26px;
  color: var(--color-text-secondary);
}

.recovery-reset__submit {
  width: 100%;
  min-height: 56px;
}

.recovery-reset__summary-list {
  margin: 0;
  padding-left: var(--space-5);
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
