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
    serverError?: LoginServerErrorSummary | null
    recoveryHref: string
  }>(),
  {
    serverError: null,
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
const summaryRef = ref<HTMLElement | null>(null)

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
      summaryRef.value?.focus()
    }
    return
  }

  emit('submit')
}

const onRetry = () => emit('retry')
</script>

<template>
  <form class="login-form" novalidate @submit.prevent="onSubmit">
    <div v-if="showSummary" ref="summaryRef" class="login-form__summary" role="alert" tabindex="-1">
      <p class="login-form__summary-title">Revisa estos campos:</p>
      <ul>
        <li v-if="fieldErrors.email">{{ fieldErrors.email }}</li>
        <li v-if="fieldErrors.password">{{ fieldErrors.password }}</li>
      </ul>
    </div>

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

    <BaseInput
      :model-value="email"
      type="email"
      name="email"
      label="Correo"
      autocomplete="username"
      placeholder="tu-correo@ejemplo.com"
      required
      :disabled="submitting"
      :error="emailError"
      @update:model-value="handleEmailInput"
    >
      <template #leading>
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.75"
          width="18"
          height="18"
        >
          <rect x="3" y="5" width="18" height="14" rx="2" />
          <path d="m4 7 8 6 8-6" />
        </svg>
      </template>
    </BaseInput>

    <BaseInput
      :model-value="password"
      type="password"
      name="password"
      label="Contraseña"
      autocomplete="current-password"
      required
      :disabled="submitting"
      :error="passwordError"
      @update:model-value="handlePasswordInput"
    >
      <template #leading>
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.75"
          width="18"
          height="18"
        >
          <rect x="4" y="11" width="16" height="9" rx="1.5" />
          <path d="M8 11V7a4 4 0 0 1 8 0v4" />
        </svg>
      </template>
    </BaseInput>

    <BaseButton
      type="submit"
      variant="primary"
      size="lg"
      :loading="submitting"
      :disabled="submitting"
      class="login-form__submit"
    >
      {{ submitting ? 'Iniciando sesión…' : 'Iniciar sesión' }}
    </BaseButton>

    <p class="login-form__recovery">
      <RouterLink :to="recoveryHref">¿Olvidaste tu contraseña?</RouterLink>
    </p>
  </form>
</template>

<style scoped>
.login-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  width: 100%;
}

.login-form__summary {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  padding: var(--space-4);
  background-color: var(--color-danger-surface);
  border: var(--border-width-normal) solid var(--color-danger-border);
  border-radius: var(--radius-md);
  color: var(--color-danger-text);
}

.login-form__summary:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-surface),
    0 0 0 4px var(--color-focus);
}

.login-form__summary-title {
  margin: 0;
  font-weight: 600;
}

.login-form__summary ul {
  margin: 0;
  padding-left: var(--space-5);
}

.login-form__submit {
  width: 100%;
}

.login-form__recovery {
  margin: 0;
  text-align: center;
  font-size: var(--font-size-body-sm);
}

.login-form__recovery a {
  color: var(--color-action-primary);
  font-weight: 500;
}
</style>
