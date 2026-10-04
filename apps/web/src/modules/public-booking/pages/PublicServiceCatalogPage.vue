<script setup lang="ts">
// Página coordinadora del catálogo público de servicios (HU-091,
// CA-091-01 a CA-091-05). Sin mockup asignado: composición libre dentro de
// NAVA / Tailored Grid (DEC-078), alojada en el cascarón de la reserva
// pública (DEC-111). Posee estado y reintento; no conoce la forma RFC 9457
// del contrato (eso queda dentro de `api/listPublicServicesApi.ts`).
//
// Selección: lista `role="radiogroup"` con roving tabindex (WAI-ARIA APG,
// mismo patrón de teclado que BarberSelect.vue, sin el popup/colapso que
// ese widget sí necesita). La selección se indica con `aria-checked`, un
// filete lateral, el rombo relleno con su check -nunca solo color (CA-091-04)-. Esta
// historia no crea ninguna cita (fuera de alcance de HU-091 y HU-092); tras
// elegir un servicio, el botón "Continuar" (HU-092) navega a la selección
// pública de barbero de ESE servicio.
import { computed, nextTick, onMounted, ref } from 'vue'
import BookingStateScreen from '../components/BookingStateScreen.vue'
import ChoiceMark from '../components/ChoiceMark.vue'
import { listPublicServices } from '../api/listPublicServicesApi'
import type { PublicService } from '../model/publicServiceListOutcome'

interface Props {
  /** Identificador del enlace público, tal como llega del parámetro de
   * ruta `:slug` (app/router, `props: true`). No confiado: el servidor lo
   * vuelve a resolver desde cero, igual que PublicBarbershopEntryPage. */
  slug: string
}
const props = defineProps<Props>()

type ScreenState =
  | { status: 'loading' }
  | { status: 'success'; services: PublicService[] }
  | { status: 'not-found' }
  | { status: 'network-error' }
  | { status: 'unexpected-error'; requestId?: string }

const screenState = ref<ScreenState>({ status: 'loading' })
const selectedServiceId = ref<string | null>(null)
const activeIndex = ref(0)
const optionRefs = ref<HTMLElement[]>([])

function setOptionRef(el: Element | { $el?: Element } | null, index: number) {
  if (el instanceof HTMLElement) optionRefs.value[index] = el
}

async function load() {
  screenState.value = { status: 'loading' }
  selectedServiceId.value = null
  const outcome = await listPublicServices(props.slug)
  switch (outcome.kind) {
    case 'success':
      screenState.value = { status: 'success', services: outcome.services }
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

const services = computed(() =>
  screenState.value.status === 'success' ? screenState.value.services : [],
)

function formatPrice(service: PublicService): string {
  return `${service.price} ${service.currency}`
}

function selectService(serviceId: string) {
  selectedServiceId.value = serviceId
  const index = services.value.findIndex((s) => s.id === serviceId)
  if (index >= 0) activeIndex.value = index
}

async function focusActive() {
  await nextTick()
  optionRefs.value[activeIndex.value]?.focus()
}

function moveActive(delta: number) {
  const count = services.value.length
  if (count === 0) return
  activeIndex.value = (activeIndex.value + delta + count) % count
  void focusActive()
}

function onListKeydown(event: KeyboardEvent) {
  switch (event.key) {
    case 'ArrowDown':
    case 'ArrowRight':
      event.preventDefault()
      moveActive(1)
      break
    case 'ArrowUp':
    case 'ArrowLeft':
      event.preventDefault()
      moveActive(-1)
      break
    case 'Home':
      event.preventDefault()
      activeIndex.value = 0
      void focusActive()
      break
    case 'End':
      event.preventDefault()
      activeIndex.value = services.value.length - 1
      void focusActive()
      break
    case ' ':
    case 'Enter': {
      event.preventDefault()
      const service = services.value[activeIndex.value]
      if (service) selectService(service.id)
      break
    }
  }
}

const failedRequestId = computed(() =>
  screenState.value.status === 'unexpected-error' ? screenState.value.requestId : undefined,
)

const selectedService = computed(() => services.value.find((s) => s.id === selectedServiceId.value))
</script>

<template>
  <BookingStateScreen
    v-if="screenState.status !== 'success'"
    :status="screenState.status"
    :request-id="failedRequestId"
    title="Elegir servicio"
    loading-headline="Cargando servicios…"
    @retry="retry"
  />

  <main v-else class="pb-page pb-page--bar">
    <p class="pb-eyebrow">Servicios</p>
    <h1 class="pb-title">Elige tu servicio</h1>
    <div class="pb-rule" aria-hidden="true"></div>

    <div v-if="services.length === 0" class="pb-empty">
      <span class="pb-empty__mark" aria-hidden="true"></span>
      <p class="pb-empty__message">
        Esta barbería todavía no tiene servicios disponibles para reservar.
      </p>
    </div>

    <ul
      v-else
      class="pb-choices"
      role="radiogroup"
      aria-label="Servicios disponibles"
      @keydown="onListKeydown"
    >
      <li
        v-for="(service, index) in services"
        :key="service.id"
        :ref="(el) => setOptionRef(el as Element | null, index)"
        class="pb-choice"
        :class="{ 'pb-choice--selected': service.id === selectedServiceId }"
        :style="{ '--pb-i': index }"
        role="radio"
        :aria-checked="service.id === selectedServiceId"
        :tabindex="index === activeIndex ? 0 : -1"
        @click="selectService(service.id)"
      >
        <ChoiceMark />
        <span class="pb-choice__body">
          <span class="pb-choice__name">{{ service.name }}</span>
          <span v-if="service.description" class="pb-choice__description">{{
            service.description
          }}</span>
          <span class="pb-choice__meta">
            <span>{{ service.durationMinutes }} min</span>
          </span>
        </span>
        <span class="pb-choice__aside">
          <span class="pb-choice__price">{{ formatPrice(service) }}</span>
        </span>
      </li>
    </ul>

    <Transition name="pb-bar">
      <div v-if="selectedService" class="pb-actionbar">
        <p class="pb-actionbar__summary">
          <span class="pb-actionbar__label">Tu selección</span>
          <span class="pb-actionbar__value">{{ selectedService.name }}</span>
        </p>
        <RouterLink
          :to="{
            name: 'reserva-publica-barbero',
            params: { slug: props.slug, serviceId: selectedService.id },
          }"
          class="pb-cta"
        >
          <span>Continuar</span>
          <span class="pb-cta__arrow" aria-hidden="true">→</span>
        </RouterLink>
      </div>
    </Transition>
  </main>
</template>
