<script setup lang="ts">
// Página coordinadora de la entrada pública de reservas (HU-090, CA-090-01
// a CA-090-05). Sin mockup asignado: composición libre dentro de NAVA /
// Tailored Grid (DEC-078), alojada en el cascarón de la reserva pública
// (`layouts/PublicBookingLayout.vue`, DEC-111). Posee estado y reintento; no
// conoce la forma RFC 9457 del contrato (eso queda dentro de
// `api/resolveBarbershopApi.ts`).
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { formatInstantInTimezone } from '@/shared/time/formatInstant'
import BookingStateScreen from '../components/BookingStateScreen.vue'
import { loadPublicProfile } from '../model/publicProfile'
import type { PublicBarbershopProfile } from '../model/barbershopProfileOutcome'
import { usePublicVocabulary } from '../model/publicVocabulary'

interface Props {
  /** Identificador del enlace público, tal como llega del parámetro de
   * ruta `:slug` (app/router, `props: true`). No confiado: el servidor lo
   * vuelve a resolver desde cero (CA-090-03). */
  slug: string
}
const props = defineProps<Props>()
const v = usePublicVocabulary()

type ScreenState =
  | { status: 'loading' }
  | { status: 'success'; profile: PublicBarbershopProfile }
  | { status: 'not-found' }
  | { status: 'network-error' }
  | { status: 'unexpected-error'; requestId?: string }

const screenState = ref<ScreenState>({ status: 'loading' })

async function load() {
  screenState.value = { status: 'loading' }
  const outcome = await loadPublicProfile(props.slug)
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

// El reloj avanza solo: la hora de la barbería es lo que la persona compara
// con su propia hora antes de elegir, y una hora congelada engaña.
const now = ref(new Date())
let clockTimer: ReturnType<typeof setInterval> | undefined

onMounted(() => {
  void load()
  clockTimer = setInterval(() => {
    now.value = new Date()
  }, 15_000)
})
onBeforeUnmount(() => clearInterval(clockTimer))

const retry = () => {
  void load()
}

// CA-090-01: "presenta la hora en su zona". El reloj del dispositivo solo
// construye el instante a formatear; el día/hora civil resultante lo
// decide `timezone` (la de la barbería), nunca la zona del dispositivo
// (mismo criterio que formatFullDateInTimezone, HU-062).
const currentTimeLabel = computed(() => {
  if (screenState.value.status !== 'success') return ''
  return formatInstantInTimezone(now.value.toISOString(), screenState.value.profile.timezone)
})

const failedRequestId = computed(() =>
  screenState.value.status === 'unexpected-error' ? screenState.value.requestId : undefined,
)
</script>

<template>
  <BookingStateScreen
    v-if="screenState.status !== 'success'"
    :status="screenState.status"
    :request-id="failedRequestId"
    title="Entrada pública de reservas"
    loading-headline="Abriendo tu reserva…"
    @retry="retry"
  />

  <main v-else class="pb-page pb-hero">
    <p class="pb-eyebrow">Reserva en línea</p>
    <h1 class="pb-hero__name">{{ screenState.profile.name }}</h1>
    <div class="pb-rule" aria-hidden="true"></div>

    <p class="pb-clock">Hora local {{ v.ofTheBusiness }}: {{ currentTimeLabel }}</p>

    <dl
      v-if="screenState.profile.contactEmail || screenState.profile.contactPhone"
      class="pb-contact"
    >
      <div v-if="screenState.profile.contactPhone" class="pb-contact__row">
        <dt>Teléfono</dt>
        <dd>{{ screenState.profile.contactPhone }}</dd>
      </div>
      <div v-if="screenState.profile.contactEmail" class="pb-contact__row">
        <dt>Correo</dt>
        <dd>{{ screenState.profile.contactEmail }}</dd>
      </div>
    </dl>

    <RouterLink
      :to="{ name: 'reserva-publica-servicios', params: { slug: props.slug } }"
      class="pb-cta pb-cta--wide"
    >
      <span>Reservar un turno</span>
      <span class="pb-cta__arrow" aria-hidden="true">→</span>
    </RouterLink>
  </main>
</template>
