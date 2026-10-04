<script setup lang="ts">
// Página coordinadora de la selección pública de barbero (HU-092, CA-092-01
// a CA-092-05). Sin mockup asignado: composición libre dentro de NAVA /
// Tailored Grid (DEC-078), alojada en el cascarón de la reserva pública
// (DEC-111). Posee estado y reintento; no conoce la forma RFC 9457 del
// contrato (eso queda dentro de `api/listPublicBarbersApi.ts`).
//
// Matriz 0/1/N (CA-092-01): con un único barbero elegible se preselecciona
// automáticamente y se informa sin exigir una interacción adicional; con
// varios, se exige elección explícita mediante un `role="radiogroup"` con
// roving tabindex (mismo patrón de teclado que PublicServiceCatalogPage.vue);
// con cero, un EmptyState distinto de carga o error.
//
// Revalidación (CA-092-04): un cambio de `serviceId` (mismo componente
// reutilizado por vue-router al cambiar solo el parámetro de ruta) dispara
// una nueva carga; una respuesta tardía de una carga ya reemplazada se
// descarta con un token monótono, y la selección previa solo se conserva si
// el barbero sigue en la lista fresca -nunca se asume compatible sin
// revalidarla contra la respuesta del servidor.
import { computed, nextTick, ref, watch } from 'vue'
import { BarberAvatar } from '@/shared/ui'
import BookingStateScreen from '../components/BookingStateScreen.vue'
import ChoiceMark from '../components/ChoiceMark.vue'
import { listPublicBarbers } from '../api/listPublicBarbersApi'
import type { PublicBarber } from '../model/publicBarberListOutcome'

interface Props {
  /** Identificador del enlace público, tal como llega del parámetro de
   * ruta `:slug`. No confiado: el servidor lo vuelve a resolver desde cero. */
  slug: string
  /** Identificador del servicio activo ya elegido, tal como llega del
   * parámetro de ruta `:serviceId`. No confiado: el servidor revalida
   * pertenencia, vigencia y asignación en cada lectura (CA-092-03). */
  serviceId: string
}
const props = defineProps<Props>()

type ScreenState =
  | { status: 'loading' }
  | { status: 'success'; barbers: PublicBarber[] }
  | { status: 'not-found' }
  | { status: 'network-error' }
  | { status: 'unexpected-error'; requestId?: string }

const screenState = ref<ScreenState>({ status: 'loading' })
const selectedBarberId = ref<string | null>(null)
const activeIndex = ref(0)
const optionRefs = ref<HTMLElement[]>([])

function setOptionRef(el: Element | { $el?: Element } | null, index: number) {
  if (el instanceof HTMLElement) optionRefs.value[index] = el
}

let requestToken = 0

async function load() {
  const token = ++requestToken
  screenState.value = { status: 'loading' }
  optionRefs.value = []
  activeIndex.value = 0

  const outcome = await listPublicBarbers(props.slug, props.serviceId)
  // Una respuesta tardía de una carga ya reemplazada (serviceId cambió de
  // nuevo mientras esta esperaba) se descarta: nunca sobrescribe el estado
  // de la carga vigente (CA-092-04).
  if (token !== requestToken) return

  switch (outcome.kind) {
    case 'success': {
      const barbers = outcome.barbers
      screenState.value = { status: 'success', barbers }
      if (barbers.length === 1) {
        // CA-092-01: exactamente un barbero elegible se preselecciona sin
        // paso adicional.
        selectedBarberId.value = barbers[0].id
      } else if (
        selectedBarberId.value !== null &&
        !barbers.some((b) => b.id === selectedBarberId.value)
      ) {
        // CA-092-04: una selección previa incompatible con la lista fresca
        // se limpia; una compatible se conserva tal cual, ya revalidada.
        selectedBarberId.value = null
      }
      return
    }
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

watch(() => [props.slug, props.serviceId], load, { immediate: true })

const retry = () => {
  void load()
}

const barbers = computed(() =>
  screenState.value.status === 'success' ? screenState.value.barbers : [],
)

const isSinglePreselected = computed(() => barbers.value.length === 1)

function selectBarber(barberId: string) {
  selectedBarberId.value = barberId
  const index = barbers.value.findIndex((b) => b.id === barberId)
  if (index >= 0) activeIndex.value = index
}

async function focusActive() {
  await nextTick()
  optionRefs.value[activeIndex.value]?.focus()
}

function moveActive(delta: number) {
  const count = barbers.value.length
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
      activeIndex.value = barbers.value.length - 1
      void focusActive()
      break
    case ' ':
    case 'Enter': {
      event.preventDefault()
      const barber = barbers.value[activeIndex.value]
      if (barber) selectBarber(barber.id)
      break
    }
  }
}

const failedRequestId = computed(() =>
  screenState.value.status === 'unexpected-error' ? screenState.value.requestId : undefined,
)

const selectedBarber = computed(() => barbers.value.find((b) => b.id === selectedBarberId.value))
</script>

<template>
  <BookingStateScreen
    v-if="screenState.status !== 'success'"
    :status="screenState.status"
    :request-id="failedRequestId"
    title="Elegir barbero"
    loading-headline="Cargando barberos…"
    @retry="retry"
  />

  <main v-else class="pb-page pb-page--bar">
    <p class="pb-eyebrow">Tu barbero</p>
    <h1 class="pb-title">Elige tu barbero</h1>
    <div class="pb-rule" aria-hidden="true"></div>

    <div v-if="barbers.length === 0" class="pb-empty">
      <span class="pb-empty__mark" aria-hidden="true"></span>
      <p class="pb-empty__message">Este servicio no tiene barberos disponibles en este momento.</p>
    </div>

    <!-- CA-092-01: un único barbero elegible se informa sin presentarse
         como una elección pendiente (sin radiogroup, sin paso adicional). -->
    <div v-else-if="isSinglePreselected" class="pb-ficha">
      <BarberAvatar :full-name="barbers[0]!.fullName" size="hero" />
      <p class="pb-ficha__text" role="status">
        Te atenderá <strong>{{ barbers[0]!.fullName }}</strong>
      </p>
    </div>

    <ul
      v-else
      class="pb-choices"
      role="radiogroup"
      aria-label="Barberos disponibles"
      @keydown="onListKeydown"
    >
      <li
        v-for="(barber, index) in barbers"
        :key="barber.id"
        :ref="(el) => setOptionRef(el as Element | null, index)"
        class="pb-choice pb-choice--person"
        :class="{ 'pb-choice--selected': barber.id === selectedBarberId }"
        :style="{ '--pb-i': index }"
        role="radio"
        :aria-checked="barber.id === selectedBarberId"
        :tabindex="index === activeIndex ? 0 : -1"
        @click="selectBarber(barber.id)"
      >
        <ChoiceMark />
        <span class="pb-choice__body pb-choice__body--person">
          <BarberAvatar :full-name="barber.fullName" size="row" />
          <span class="pb-choice__name">{{ barber.fullName }}</span>
        </span>
      </li>
    </ul>

    <Transition name="pb-bar">
      <div v-if="selectedBarber" class="pb-actionbar">
        <!-- Con un único barbero la ficha de arriba ya dice quién atiende:
             repetir su nombre aquí solo duplicaría el mismo dato. -->
        <p v-if="isSinglePreselected" class="pb-actionbar__summary">
          <span class="pb-actionbar__label">Siguiente</span>
          <span class="pb-actionbar__value">Fecha y hora</span>
        </p>
        <p v-else class="pb-actionbar__summary">
          <span class="pb-actionbar__label">Tu barbero</span>
          <span class="pb-actionbar__value">{{ selectedBarber.fullName }}</span>
        </p>
        <RouterLink
          :to="{
            name: 'reserva-publica-horario',
            params: { slug: props.slug, serviceId: props.serviceId, barberId: selectedBarber.id },
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
