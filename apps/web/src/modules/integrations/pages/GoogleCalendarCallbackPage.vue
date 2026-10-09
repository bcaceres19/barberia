<script setup lang="ts">
// Pantalla de retorno de Google (issue #325, DEC-102): es el `redirect_uri`
// registrado en Google Cloud. Google redirige aquí con `state` y `code` (o
// `error` si el barbero rechazó); la pantalla los reenvía una sola vez al API,
// que valida el `state` (de un solo uso, ligado a esta sesión), canjea el código
// con PKCE y guarda el token cifrado. El código nunca se guarda ni se registra
// y la URL se limpia en cuanto se lee.
import { onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useToast } from '@/shared/composables'
import { BaseAlert, BaseButton, DiamondLoader } from '@/shared/ui'
import { completeCallback } from '../api/googleCalendarApi'

type Phase = 'working' | 'invalid' | 'denied' | 'failed' | 'error'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const phase = ref<Phase>('working')
const LOADING_PHRASES = ['Hablando con Google', 'Guardando tu permiso', 'Todo a su hora'] as const

function first(value: unknown): string {
  if (Array.isArray(value)) return typeof value[0] === 'string' ? value[0] : ''
  return typeof value === 'string' ? value : ''
}

async function finish() {
  const state = first(route.query.state)
  const code = first(route.query.code)
  const error = first(route.query.error)

  // El código y el state no deben quedarse en la URL ni en el historial.
  await router.replace({ name: 'integraciones-google-calendar-retorno', query: {} })

  if (!state) {
    phase.value = 'invalid'
    return
  }
  const outcome = await completeCallback({ state, code, error })
  if (outcome.kind === 'success') {
    if (outcome.result === 'connected') {
      toast.success('Google Calendar conectado', {
        detail: 'Tus turnos y bloqueos aparecerán en tu calendario.',
      })
      await router.replace({ name: 'integraciones-google-calendar' })
      return
    }
    phase.value = outcome.result === 'denied' ? 'denied' : 'failed'
    return
  }
  phase.value = outcome.kind === 'invalid' ? 'invalid' : 'error'
}

// Una sola vez: el state solo sirve una vez y un segundo envío lo reportaría inválido.
let started = false
onMounted(() => {
  if (started) return
  started = true
  void finish()
})
</script>

<template>
  <section class="gcal-return" aria-labelledby="gcal-return-title">
    <h1 id="gcal-return-title" class="gcal-return__title">Google Calendar</h1>

    <div v-if="phase === 'working'" class="gcal-return__state" role="status" aria-live="polite">
      <DiamondLoader label="Conectando con Google…" layout="inline" :phrases="LOADING_PHRASES" />
    </div>

    <BaseAlert v-else :variant="phase === 'denied' ? 'warning' : 'danger'" role="alert">
      <strong>{{
        phase === 'denied'
          ? 'No diste el permiso.'
          : phase === 'invalid'
            ? 'La autorización venció.'
            : 'No pudimos conectar.'
      }}</strong>
      <template v-if="phase === 'denied'">
        Sin el permiso de calendario NAVA no puede publicar tu agenda. Puedes intentarlo de nuevo
        cuando quieras.
      </template>
      <template v-else-if="phase === 'invalid'">
        El enlace de Google ya se usó o venció. Inicia la conexión de nuevo desde NAVA.
      </template>
      <template v-else-if="phase === 'failed'">
        Google no completó la conexión. Inténtalo de nuevo; tus turnos no se tocaron.
      </template>
      <template v-else>
        Hubo un problema al hablar con NAVA. Revisa tu conexión e inténtalo de nuevo.
      </template>
      <template #action>
        <RouterLink :to="{ name: 'integraciones-google-calendar' }" class="gcal-return__link">
          <BaseButton type="button" variant="secondary">Volver a Google Calendar</BaseButton>
        </RouterLink>
      </template>
    </BaseAlert>
  </section>
</template>

<style scoped>
.gcal-return {
  display: flex;
  flex-direction: column;
  gap: 20px;
  width: min(100%, 640px);
  margin: 0 auto;
  padding: 48px 16px;
  color: var(--color-on-strong);
  box-sizing: border-box;
}

.gcal-return__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h1);
  font-weight: var(--font-weight-h1);
  line-height: var(--font-size-h1-line);
}

.gcal-return__state {
  color: var(--color-on-strong-muted);
}

.gcal-return__link {
  text-decoration: none;
}
</style>
