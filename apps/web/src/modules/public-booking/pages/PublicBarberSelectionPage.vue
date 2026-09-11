<script setup lang="ts">
// Página coordinadora de la selección pública de barbero (HU-092, CA-092-01
// a CA-092-05). Sin mockup asignado: composición libre dentro de NAVA /
// Tailored Grid (DEC-078). Posee estado y reintento; no conoce la forma
// RFC 9457 del contrato (eso queda dentro de `api/listPublicBarbersApi.ts`).
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
import { BaseButton, EmptyState, PageState } from '@/shared/ui'
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

const unexpectedErrorMessage = computed(() => {
  const state = screenState.value
  if (state.status !== 'unexpected-error') return ''
  return state.requestId
    ? `Inténtalo de nuevo. Si continúa, comparte este código con soporte: ${state.requestId}.`
    : 'Inténtalo de nuevo en unos segundos.'
})
</script>

<template>
  <main v-if="screenState.status !== 'success'" class="barber-selection barber-selection--state">
    <h1 class="visually-hidden">Elegir barbero</h1>
    <PageState
      v-if="screenState.status === 'loading'"
      variant="loading"
      headline="Cargando barberos…"
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

  <main v-else class="barber-selection barber-selection--list">
    <div class="barber-selection__container">
      <h1 class="barber-selection__title">Elige tu barbero</h1>

      <EmptyState
        v-if="barbers.length === 0"
        message="Este servicio no tiene barberos disponibles en este momento."
      />

      <!-- CA-092-01: un único barbero elegible se informa sin presentarse
           como una elección pendiente (sin radiogroup, sin paso adicional). -->
      <p v-else-if="isSinglePreselected" class="barber-selection__preselected" role="status">
        Te atenderá <strong>{{ barbers[0]!.fullName }}</strong>
      </p>

      <ul
        v-else
        class="barber-selection__list"
        role="radiogroup"
        aria-label="Barberos disponibles"
        @keydown="onListKeydown"
      >
        <li
          v-for="(barber, index) in barbers"
          :key="barber.id"
          :ref="(el) => setOptionRef(el as Element | null, index)"
          class="barber-selection__item"
          :class="{ 'barber-selection__item--selected': barber.id === selectedBarberId }"
          role="radio"
          :aria-checked="barber.id === selectedBarberId"
          :tabindex="index === activeIndex ? 0 : -1"
          @click="selectBarber(barber.id)"
        >
          <span class="barber-selection__item-check" aria-hidden="true"></span>
          <span class="barber-selection__item-name">{{ barber.fullName }}</span>
        </li>
      </ul>
    </div>
  </main>
</template>

<style scoped>
.barber-selection {
  display: flex;
  min-height: 100dvh;
  justify-content: center;
  padding: var(--space-6) var(--space-4);
}

.barber-selection--state {
  align-items: center;
  background-color: var(--color-surface-strong);
}

.barber-selection--list {
  background-color: var(--color-canvas);
}

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

.barber-selection__container {
  display: flex;
  width: 100%;
  max-width: 640px;
  flex-direction: column;
  gap: var(--space-5);
}

.barber-selection__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: 32px;
  line-height: 40px;
  letter-spacing: -0.015em;
  color: var(--color-text-primary);
}

.barber-selection__preselected {
  margin: 0;
  padding: var(--space-4);
  color: var(--color-text-primary);
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: 4px;
}

.barber-selection__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding: 0;
  margin: 0;
  list-style: none;
}

.barber-selection__item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4);
  cursor: pointer;
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: 4px;
}

.barber-selection__item:hover {
  border-color: var(--color-accent-brass);
}

.barber-selection__item:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-canvas),
    0 0 0 4px var(--color-focus);
}

/* La selección se marca con el ícono de check Y el borde de énfasis, nunca
   solo con un cambio de color (CA-092-05, mismo criterio que
   PublicServiceCatalogPage.vue/CA-091-04). */
.barber-selection__item--selected {
  border-color: var(--color-accent-brass);
  border-width: var(--border-width-emphasis);
}

.barber-selection__item-check {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: 50%;
}

.barber-selection__item--selected .barber-selection__item-check {
  background-color: var(--color-accent-brass);
  border-color: var(--color-accent-brass);
}

.barber-selection__item-name {
  font-weight: 600;
  color: var(--color-text-primary);
}

@media (min-width: 1024px) {
  .barber-selection__title {
    font-size: 40px;
    line-height: 48px;
  }
}
</style>
