<script setup lang="ts">
// Formulario de acceso: correo y contraseña, sin conocer transporte HTTP
// ni la forma RFC 9457 del error (docs/03-desarrollo/estandar-frontend-vue.md
// §6). Valida forma en cliente y emite `submit` solo cuando es localmente
// válido; `LoginPage.vue` decide qué ocurre después (llamar al API,
// navegar, mostrar un resultado del servidor).
import { computed, nextTick, ref } from 'vue'
import { BaseAlert, BaseButton, BaseInput } from '@/shared/ui'
import {
  hasLoginFieldErrors,
  validateLoginForm,
  type LoginFieldErrors,
} from '../validation/loginValidation'

/** Resumen de error ya traducido a texto seguro por la página: el
 * formulario solo lo muestra, no interpreta `status`/`code`/`Problem`. */
export interface LoginServerErrorSummary {
  title: string
  message: string
  tone: 'danger' | 'warning' | 'info'
  /** Acción de recuperación opcional (por ejemplo, "Reintentar" ante un
   * error de red, CA-010-03). */
  actionLabel?: string
}

const props = withDefaults(
  defineProps<{
    email: string
    password: string
    submitting: boolean
    /** El reto adicional deja las credenciales visibles pero bloqueadas:
     * mientras esté activo, su única acción primaria es verificar el OTP. */
    challengeActive?: boolean
    serverError?: LoginServerErrorSummary | null
    recoveryHref: string
  }>(),
  {
    serverError: null,
    challengeActive: false,
  },
)

const emit = defineEmits<{
  'update:email': [value: string]
  'update:password': [value: string]
  submit: []
  retry: []
}>()

const fieldErrors = ref<LoginFieldErrors>({})
const attemptedSubmit = ref(false)
const summaryRef = ref<{ focus: () => void } | null>(null)

// Solo se muestran errores de campo tras el primer intento de envío: no se
// regaña al barbero mientras todavía está escribiendo por primera vez.
const showFieldErrors = computed(() => attemptedSubmit.value)
const emailError = computed(() => (showFieldErrors.value ? fieldErrors.value.email : undefined))
const passwordError = computed(() =>
  showFieldErrors.value ? fieldErrors.value.password : undefined,
)

const errorCount = computed(() => Object.keys(fieldErrors.value).length)
const showSummary = computed(() => showFieldErrors.value && errorCount.value > 1)

// Revalida en vivo, tras un intento fallido, en respuesta a la escritura
// real del barbero (eventos de entrada de BaseInput), NUNCA a un cambio
// de prop en sí mismo. Esto importa porque LoginPage limpia la
// contraseña por su cuenta ante credenciales inválidas (CA-010-02): un
// `watch` sobre `props.email`/`props.password` reaccionaría también a
// ESE reinicio programático y mostraría "Escribe tu contraseña" al mismo
// tiempo que el servidor ya explicó el problema — dos mensajes
// contradictorios sobre el mismo campo. Validar con el valor recién
// escrito (no con la prop, que todavía no se actualizó) evita depender
// de un `nextTick` adicional.
const handleEmailInput = (value: string | number) => {
  const nextValue = String(value)
  emit('update:email', nextValue)
  if (attemptedSubmit.value) {
    fieldErrors.value = validateLoginForm(nextValue, props.password)
  }
}
const handlePasswordInput = (value: string | number) => {
  const nextValue = String(value)
  emit('update:password', nextValue)
  if (attemptedSubmit.value) {
    fieldErrors.value = validateLoginForm(props.email, nextValue)
  }
}

const onSubmit = async () => {
  const errors = validateLoginForm(props.email, props.password)
  fieldErrors.value = errors
  attemptedSubmit.value = true

  if (hasLoginFieldErrors(errors)) {
    // Foco al resumen cuando hay más de un error (CA-010-05, estándar
    // visual §16); con un único error el foco nativo del campo inválido
    // ya es suficiente y un resumen de un solo ítem sería ruido.
    if (errorCount.value > 1) {
      await nextTick()
      summaryRef.value?.focus?.()
    }
    return
  }

  emit('submit')
}

const onRetry = () => emit('retry')
</script>

<template>
  <form class="login-form" novalidate @submit.prevent="onSubmit">
    <BaseInput
      :model-value="email"
      type="email"
      name="email"
      label="Correo"
      autocomplete="username"
      placeholder="tu-correo@ejemplo.com"
      required
      :show-required-marker="false"
      :disabled="submitting || challengeActive"
      :error="emailError"
      @update:model-value="handleEmailInput"
    />

    <BaseInput
      :model-value="password"
      type="password"
      name="password"
      label="Contraseña"
      autocomplete="current-password"
      required
      :show-required-marker="false"
      :disabled="submitting || challengeActive"
      :error="passwordError"
      @update:model-value="handlePasswordInput"
    />

    <BaseButton
      type="submit"
      variant="primary"
      size="lg"
      :loading="submitting"
      :disabled="submitting || challengeActive"
      class="login-form__submit"
    >
      {{ submitting ? 'Iniciando sesión…' : 'Iniciar sesión' }}
    </BaseButton>

    <p class="login-form__recovery">
      <RouterLink :to="recoveryHref">¿Olvidaste tu contraseña?</RouterLink>
    </p>

    <!-- Ancla de alertas (issue #213, auth-eventos/README.md): la alerta
         global va después del botón y del enlace de recuperación, nunca
         entre el título y los campos. -->
    <BaseAlert
      v-if="showSummary"
      ref="summaryRef"
      variant="danger"
      title="Revisa estos campos"
      role="alert"
      tabindex="-1"
    >
      <!-- Desviación conocida (trabajo requerido §7): el mockup muestra un
           resumen que explica qué falta; el código sigue enumerando cada
           error de campo tal como ya se comporta hoy. -->
      <ul class="login-form__summary-list">
        <li v-if="fieldErrors.email">{{ fieldErrors.email }}</li>
        <li v-if="fieldErrors.password">{{ fieldErrors.password }}</li>
      </ul>
    </BaseAlert>

    <BaseAlert
      v-if="serverError"
      :variant="serverError.tone"
      :title="serverError.title"
      role="alert"
    >
      {{ serverError.message }}
      <template v-if="serverError.actionLabel" #action>
        <BaseButton variant="secondary" size="md" type="button" @click="onRetry">
          {{ serverError.actionLabel }}
        </BaseButton>
      </template>
    </BaseAlert>
  </form>
</template>

<style scoped>
.login-form {
  display: flex;
  flex-direction: column;
  gap: 26px;
  width: 100%;
}

.login-form :deep(.base-input__wrapper) {
  gap: 6px;
}

.login-form :deep(.base-input__label) {
  font-size: 16px;
  font-weight: 600;
  line-height: 22px;
}

.login-form :deep(.base-input) {
  height: 56px;
  padding-right: var(--space-4);
  padding-left: var(--space-4);
  font-size: 18px;
  line-height: 24px;
  background-color: var(--color-surface);
  border-color: var(--color-border-subtle);
  border-radius: var(--radius-sm);
}

/* "OCULTAR" (la palabra más larga del conmutador reglado) necesita más
   ancho reservado que el icono de 22px que ocupaba este mismo lugar. */
.login-form :deep(.base-input__input-wrapper:has(.base-input__toggle) .base-input) {
  padding-right: 92px;
}

.login-form :deep(.base-input__toggle) {
  min-width: 84px;
}

.login-form :deep(.base-input__toggle:hover:not(:disabled)) {
  background-color: var(--color-overlay-hover);
}

.login-form :deep(.base-input__toggle:focus-visible) {
  outline: var(--border-width-emphasis) solid var(--color-focus);
  outline-offset: -4px;
  box-shadow: none;
}

.login-form :deep(.base-input__toggle:disabled) {
  opacity: 0.48;
}

.login-form__summary-list {
  margin: 0;
  padding-left: var(--space-5);
}

.login-form__submit {
  width: 100%;
  height: 56px;
  border-radius: var(--radius-sm);
  font-size: 18px;
  font-weight: 600;
  letter-spacing: 0.01em;
}

.login-form__recovery {
  margin: 0;
  text-align: center;
  font-size: 18px;
  line-height: 26px;
}

.login-form__recovery a {
  color: var(--color-action-primary);
  font-weight: 500;
  text-decoration-thickness: 1px;
  text-underline-offset: 3px;
}

@media (min-width: 1024px) {
  .login-form {
    gap: 28px;
  }

  .login-form :deep(.base-input__wrapper) {
    gap: 8px;
  }

  .login-form :deep(.base-input__label) {
    font-size: 18px;
    line-height: 24px;
  }

  .login-form :deep(.base-input__input-wrapper:has(.base-input__toggle) .base-input) {
    padding-right: 96px;
  }

  .login-form :deep(.base-input__toggle) {
    min-width: 88px;
  }
}
</style>
