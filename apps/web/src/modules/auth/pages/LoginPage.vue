<script setup lang="ts">
// Página coordinadora de la pantalla de acceso (P0, estandar-diseno-visual.md
// §10): posee estado y navegación; `LoginForm.vue` solo presenta. Traduce
// `LoginOutcome` (ya libre de detalles de transporte) a texto seguro para
// el barbero, nunca al revés.
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import LoginForm, { type LoginServerErrorSummary } from '../components/LoginForm.vue'
import { login } from '../api/loginApi'
import { rememberSessionUntil } from '../model/sessionMarker'
import type { LoginOutcome } from '../model/loginOutcome'
import type { LoginScreenState } from '../model/loginScreenState'

const router = useRouter()

const email = ref('')
const password = ref('')
const screenState = ref<LoginScreenState>({ status: 'idle' })

const isSubmitting = computed(() => screenState.value.status === 'submitting')

// Convierte el resultado ya mapeado (`LoginOutcome`) en el resumen que
// `LoginForm` puede mostrar sin conocer status HTTP ni `Problem`.
const serverError = computed<LoginServerErrorSummary | null>(() => {
  const state = screenState.value
  switch (state.status) {
    case 'invalid-credentials':
      return {
        tone: 'danger',
        title: 'No pudimos iniciar tu sesión',
        // CA-005-02/CA-010-02: el mismo mensaje cubre correo inexistente y
        // contraseña incorrecta; nunca distingue el motivo.
        message: 'Revisa tu correo y contraseña e inténtalo de nuevo.',
      }
    case 'validation-error':
      return {
        tone: 'danger',
        title: 'Revisa los datos ingresados',
        message: 'El correo o la contraseña no tienen un formato válido.',
      }
    case 'network-error':
      return {
        tone: 'warning',
        title: 'No pudimos conectar',
        message: 'Revisa tu conexión e inténtalo de nuevo.',
        actionLabel: 'Reintentar',
      }
    case 'rate-limited': {
      const seconds = state.retryAfterSeconds
      return {
        tone: 'warning',
        title: 'Demasiados intentos',
        message:
          seconds && seconds > 0
            ? `Espera unos ${Math.ceil(seconds / 60) || 1} minuto(s) antes de volver a intentarlo.`
            : 'Espera un momento antes de volver a intentarlo.',
      }
    }
    case 'unexpected-error':
      return {
        tone: 'danger',
        title: 'Ocurrió un error inesperado',
        message: state.requestId
          ? `Inténtalo de nuevo. Si continúa, comparte este código con soporte: ${state.requestId}.`
          : 'Inténtalo de nuevo en unos segundos.',
      }
    default:
      return null
  }
})

async function attemptLogin() {
  // Guardia de doble envío (CA-010-04): la asignación a un ref es
  // síncrona, así que una segunda llamada reentrante (doble clic, Enter
  // repetido) ve el estado ya actualizado antes de proceder.
  if (screenState.value.status === 'submitting') return

  screenState.value = { status: 'submitting' }

  const outcome: LoginOutcome = await login({ email: email.value, password: password.value })

  switch (outcome.kind) {
    case 'success':
      rememberSessionUntil(outcome.expiresAt)
      // CA-010-01/DEC-056: navega exactamente a /panel; el cascarón
      // completo (cabecera, navegación) es responsabilidad de HU-012.
      await router.push({ name: 'panel' })
      return
    case 'invalid-credentials':
      // CA-010-02: conserva el correo escrito, no el criterio de si la
      // contraseña se limpia. Se limpia aquí a propósito: el intento fue
      // rechazado por el servidor (credencial evaluada e incorrecta), así
      // que no hay razón funcional para reenviar el mismo valor sin que
      // el barbero lo revise. CA-010-03 (fallo de red) es distinto: ahí
      // el servidor nunca evaluó la credencial, así que sí se conserva
      // para permitir un reintento de un solo toque.
      password.value = ''
      screenState.value = { status: 'invalid-credentials' }
      return
    case 'validation-error':
      screenState.value = { status: 'validation-error' }
      return
    case 'network-error':
      screenState.value = { status: 'network-error' }
      return
    case 'rate-limited':
      screenState.value = { status: 'rate-limited', retryAfterSeconds: outcome.retryAfterSeconds }
      return
    case 'unexpected-error':
      screenState.value = { status: 'unexpected-error', requestId: outcome.requestId }
  }
}

const onSubmit = () => {
  void attemptLogin()
}
</script>

<template>
  <main class="login-page">
    <div class="login-page__card">
      <p class="login-page__brand">Barbería</p>
      <h1 class="login-page__title">Inicia sesión</h1>

      <LoginForm
        :email="email"
        :password="password"
        :submitting="isSubmitting"
        :server-error="serverError"
        recovery-href="/recuperar-acceso"
        @update:email="(value) => (email = value)"
        @update:password="(value) => (password = value)"
        @submit="onSubmit"
        @retry="onSubmit"
      />
    </div>
  </main>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100dvh;
  padding: var(--space-6) var(--space-4);
}

.login-page__card {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  width: 100%;
  max-width: 360px;
}

.login-page__brand {
  margin: 0;
  font-size: var(--font-size-body-sm);
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.login-page__title {
  margin: 0 0 var(--space-2) 0;
  font-size: var(--font-size-h2);
  line-height: var(--font-size-h2-line);
  color: var(--color-text-primary);
}
</style>
