<script setup lang="ts">
// Página coordinadora de la pantalla de acceso (P0, estandar-diseno-visual.md
// §10): posee estado y navegación; `LoginForm.vue` solo presenta. Traduce
// `LoginOutcome` (ya libre de detalles de transporte) a texto seguro para
// el barbero, nunca al revés.
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { BaseAlert } from '@/shared/ui'
import AuthSplitLayout from '../components/AuthSplitLayout.vue'
import LoginForm, { type LoginServerErrorSummary } from '../components/LoginForm.vue'
import PhoneChallengeForm from '../components/PhoneChallengeForm.vue'
import { login } from '../api/loginApi'
import { resetForFreshLogin } from '../model/sessionStore'
import { isSafeInternalRedirect } from '../model/redirectTarget'
import type { LoginOutcome } from '../model/loginOutcome'
import type { LoginScreenState } from '../model/loginScreenState'

const router = useRouter()
const route = useRoute()

const email = ref('')
const password = ref('')
const screenState = ref<LoginScreenState>({ status: 'idle' })

const isSubmitting = computed(() => screenState.value.status === 'submitting')

// "Sesión vencida" (especificacion-frontend-nava.md §7.1): installSessionHandling.ts
// ya empuja este query param al redirigir tras un 401 fuera de acceso; esta
// pantalla solo le da el mensaje en contexto que el estándar exige. No
// distingue más motivo que este (nunca revela detalle de seguridad) y deja
// de mostrarse en cuanto el barbero reintenta un envío.
const showSessionExpired = computed(
  () => route.query.motivo === 'sesion-expirada' && screenState.value.status === 'idle',
)

// HU-007 (DEC-062): un 429 ofrece el reto telefónico como salida inmediata,
// sin esperar el escalamiento de 24 horas. `PhoneChallengeForm` es dueño de
// su propio flujo (solicitar/verificar código); esta página solo decide
// cuándo mostrarlo y qué hacer cuando se verifica con éxito.
const showPhoneChallenge = computed(() => screenState.value.status === 'rate-limited')

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
    case 'success': {
      // HU-012 (DEC-060): ya no existe un marcador local de sesión; la
      // próxima ruta privada consulta la fuente real
      // (GET /private/auth/session) en vez de reutilizar un resultado
      // `unauthenticated` previo a este mismo inicio de sesión.
      resetForFreshLogin()
      // CA-012-02: vuelve al destino pretendido si el guard lo conservó y
      // es una ruta interna segura; de lo contrario, /panel (CA-010-01).
      const redirect = route.query.redirect
      const destination = isSafeInternalRedirect(redirect) ? redirect : { name: 'panel' }
      await router.push(destination)
      return
    }
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

// El servidor ya limpió el escalamiento de esta IP en la misma verificación
// exitosa (DEC-062): reintenta el login normalmente, con las credenciales
// que el barbero ya escribió, sin pedirle que las repita a mano.
const onChallengeVerified = () => {
  void attemptLogin()
}
</script>

<template>
  <AuthSplitLayout>
    <h1 class="login-page__title">Accede a NAVA</h1>

    <BaseAlert v-if="showSessionExpired" variant="info" title="Tu sesión venció" role="status">
      Inicia sesión de nuevo para continuar.
    </BaseAlert>

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

    <PhoneChallengeForm
      v-if="showPhoneChallenge"
      :email="email"
      :disabled="isSubmitting"
      @verified="onChallengeVerified"
    />
  </AuthSplitLayout>
</template>

<style scoped>
.login-page__title {
  margin: 0 0 var(--space-2) 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h2);
  line-height: var(--font-size-h2-line);
  color: var(--color-text-primary);
}
</style>
