<script setup lang="ts">
// Página coordinadora de la entrada pública de reservas (HU-090, CA-090-01
// a CA-090-05). Sin mockup asignado: composición libre dentro de NAVA /
// Tailored Grid (DEC-078). Posee estado y reintento; no conoce la forma
// RFC 9457 del contrato (eso queda dentro de `api/resolveBarbershopApi.ts`).
import { computed, onMounted, ref } from 'vue'
import { BaseAlert, BaseButton, NavaWordmark, PageState } from '@/shared/ui'
import { formatInstantInTimezone } from '@/shared/time/formatInstant'
import { resolveBarbershop } from '../api/resolveBarbershopApi'
import type { PublicBarbershopProfile } from '../model/barbershopProfileOutcome'

interface Props {
  /** Identificador del enlace público, tal como llega del parámetro de
   * ruta `:slug` (app/router, `props: true`). No confiado: el servidor lo
   * vuelve a resolver desde cero (CA-090-03). */
  slug: string
}
const props = defineProps<Props>()

type ScreenState =
  | { status: 'loading' }
  | { status: 'success'; profile: PublicBarbershopProfile }
  | { status: 'not-found' }
  | { status: 'network-error' }
  | { status: 'unexpected-error'; requestId?: string }

const screenState = ref<ScreenState>({ status: 'loading' })

async function load() {
  screenState.value = { status: 'loading' }
  const outcome = await resolveBarbershop(props.slug)
  switch (outcome.kind) {
    case 'success':
      screenState.value = { status: 'success', profile: outcome.profile }
      return
    case 'not-found':
      screenState.value = { status: 'not-found' }
      return
    case 'network-error':
      screenState.value = { status: 'network-error' }
      return
    case 'unexpected-error':
      screenState.value = { status: 'unexpected-error', requestId: outcome.requestId }
  }
}

onMounted(load)

const retry = () => {
  void load()
}

// CA-090-01: "presenta la hora en su zona". El reloj del dispositivo solo
// construye el instante a formatear; el día/hora civil resultante lo
// decide `timezone` (la de la barbería), nunca la zona del dispositivo
// (mismo criterio que formatFullDateInTimezone, HU-062).
const currentTimeLabel = computed(() => {
  if (screenState.value.status !== 'success') return ''
  return formatInstantInTimezone(new Date().toISOString(), screenState.value.profile.timezone)
})

const unexpectedErrorMessage = computed(() => {
  const state = screenState.value
  if (state.status !== 'unexpected-error') return ''
  return state.requestId
    ? `Inténtalo de nuevo. Si continúa, comparte este código con soporte: ${state.requestId}.`
    : 'Inténtalo de nuevo en unos segundos.'
})
</script>

<template>
  <main v-if="screenState.status !== 'success'" class="public-entry public-entry--state">
    <!-- Encabezado de página accesible pero visualmente oculto: PageState
         solo aporta un <h2> (subordinado a un <h1> que su consumidor
         habitual ya tiene en otra parte del panel privado); esta pantalla
         pública NO tiene ningún otro <h1>, así que lo provee aquí para que
         la página siempre tenga uno (page-has-heading-one, axe-core). -->
    <h1 class="visually-hidden">Entrada pública de reservas</h1>
    <PageState
      v-if="screenState.status === 'loading'"
      variant="loading"
      headline="Abriendo tu barbería…"
      role="status"
    />
    <PageState
      v-else-if="screenState.status === 'not-found'"
      variant="danger"
      status-label="Error"
      headline="No encontramos ese enlace"
      role="alert"
    >
      <template #default
        >Revisa que copiaste la dirección completa, o pídele al barbero que te la vuelva a
        compartir.</template
      >
      <template #action>
        <BaseButton variant="secondary" @click="retry">Reintentar</BaseButton>
      </template>
    </PageState>
    <PageState
      v-else-if="screenState.status === 'network-error'"
      variant="warning"
      status-label="Atención"
      headline="No pudimos conectar"
      role="alert"
    >
      <template #default>Revisa tu conexión e inténtalo de nuevo.</template>
      <template #action>
        <BaseButton variant="primary" @click="retry">Reintentar</BaseButton>
      </template>
    </PageState>
    <PageState
      v-else
      variant="danger"
      status-label="Error"
      headline="Ocurrió un error inesperado"
      role="alert"
    >
      <template #default>{{ unexpectedErrorMessage }}</template>
      <template #action>
        <BaseButton variant="primary" @click="retry">Reintentar</BaseButton>
      </template>
    </PageState>
  </main>

  <main v-else class="public-entry public-entry--profile">
    <div class="public-entry__card">
      <NavaWordmark variant="ink" size="lg" />
      <div class="public-entry__header">
        <h1 class="public-entry__name">{{ screenState.profile.name }}</h1>
        <p class="public-entry__time">Hora local de la barbería: {{ currentTimeLabel }}</p>
      </div>

      <BaseAlert
        v-if="screenState.profile.contactEmail || screenState.profile.contactPhone"
        variant="plain"
        role="status"
      >
        <p v-if="screenState.profile.contactPhone" class="public-entry__contact-line">
          Teléfono: {{ screenState.profile.contactPhone }}
        </p>
        <p v-if="screenState.profile.contactEmail" class="public-entry__contact-line">
          Correo: {{ screenState.profile.contactEmail }}
        </p>
      </BaseAlert>

      <p class="public-entry__coming-soon">
        Muy pronto podrás elegir servicio, barbero y horario desde aquí mismo.
      </p>
    </div>
  </main>
</template>

<style scoped>
.public-entry {
  display: flex;
  min-height: 100dvh;
  align-items: center;
  justify-content: center;
  padding: var(--space-6) var(--space-4);
}

.public-entry--state {
  background-color: var(--color-surface-strong);
}

.public-entry--profile {
  background-color: var(--color-canvas);
}

/* Técnica estándar "sr-only": presente para tecnología de asistencia,
   invisible y sin ocupar espacio para el resto de personas usuarias. */
.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  margin: -1px;
  padding: 0;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.public-entry__card {
  display: flex;
  width: 100%;
  max-width: 480px;
  flex-direction: column;
  align-items: center;
  gap: var(--space-5);
  text-align: center;
}

.public-entry__header {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.public-entry__name {
  margin: 0;
  font-family: var(--font-display);
  font-size: 40px;
  line-height: 48px;
  letter-spacing: -0.015em;
  color: var(--color-text-primary);
}

.public-entry__time {
  margin: 0;
  font-size: var(--font-size-body);
  color: var(--color-text-secondary);
}

.public-entry__contact-line {
  margin: 0;
}

.public-entry__contact-line + .public-entry__contact-line {
  margin-top: var(--space-1);
}

.public-entry__coming-soon {
  margin: 0;
  font-size: var(--font-size-body);
  color: var(--color-text-secondary);
}

@media (min-width: 1024px) {
  .public-entry__name {
    font-size: 48px;
    line-height: 56px;
  }
}
</style>
