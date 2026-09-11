<script setup lang="ts">
// Página coordinadora del catálogo público de servicios (HU-091,
// CA-091-01 a CA-091-05). Sin mockup asignado: composición libre dentro de
// NAVA / Tailored Grid (DEC-078). Posee estado y reintento; no conoce la
// forma RFC 9457 del contrato (eso queda dentro de
// `api/listPublicServicesApi.ts`).
//
// Selección: lista `role="radiogroup"` con roving tabindex (WAI-ARIA APG,
// mismo patrón de teclado que BarberSelect.vue, sin el popup/colapso que
// ese widget sí necesita). La selección se indica con `aria-checked`, un
// borde de énfasis y un ícono de marca -nunca solo color (CA-091-04)-. Esta
// historia no crea ninguna cita ni navega a la siguiente pantalla (fuera de
// alcance de HU-091, HU-092 en adelante); elegir un servicio solo deja
// constancia local de cuál quedó marcado.
import { computed, nextTick, onMounted, ref } from 'vue'
import { BaseButton, EmptyState, PageState } from '@/shared/ui'
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

const unexpectedErrorMessage = computed(() => {
  const state = screenState.value
  if (state.status !== 'unexpected-error') return ''
  return state.requestId
    ? `Inténtalo de nuevo. Si continúa, comparte este código con soporte: ${state.requestId}.`
    : 'Inténtalo de nuevo en unos segundos.'
})
</script>

<template>
  <main v-if="screenState.status !== 'success'" class="service-catalog service-catalog--state">
    <h1 class="visually-hidden">Elegir servicio</h1>
    <PageState
      v-if="screenState.status === 'loading'"
      variant="loading"
      headline="Cargando servicios…"
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

  <main v-else class="service-catalog service-catalog--list">
    <div class="service-catalog__container">
      <h1 class="service-catalog__title">Elige tu servicio</h1>

      <EmptyState
        v-if="services.length === 0"
        message="Esta barbería todavía no tiene servicios disponibles para reservar."
      />

      <ul
        v-else
        class="service-catalog__list"
        role="radiogroup"
        aria-label="Servicios disponibles"
        @keydown="onListKeydown"
      >
        <li
          v-for="(service, index) in services"
          :key="service.id"
          :ref="(el) => setOptionRef(el as Element | null, index)"
          class="service-catalog__item"
          :class="{ 'service-catalog__item--selected': service.id === selectedServiceId }"
          role="radio"
          :aria-checked="service.id === selectedServiceId"
          :tabindex="index === activeIndex ? 0 : -1"
          @click="selectService(service.id)"
        >
          <span class="service-catalog__item-check" aria-hidden="true"></span>
          <span class="service-catalog__item-body">
            <span class="service-catalog__item-name">{{ service.name }}</span>
            <span v-if="service.description" class="service-catalog__item-description">{{
              service.description
            }}</span>
            <span class="service-catalog__item-meta">
              <span>{{ service.durationMinutes }} min</span>
              <span aria-hidden="true">·</span>
              <span>{{ formatPrice(service) }}</span>
            </span>
          </span>
        </li>
      </ul>
    </div>
  </main>
</template>

<style scoped>
.service-catalog {
  display: flex;
  min-height: 100dvh;
  justify-content: center;
  padding: var(--space-6) var(--space-4);
}

.service-catalog--state {
  align-items: center;
  background-color: var(--color-surface-strong);
}

.service-catalog--list {
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

.service-catalog__container {
  display: flex;
  width: 100%;
  max-width: 640px;
  flex-direction: column;
  gap: var(--space-5);
}

.service-catalog__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: 32px;
  line-height: 40px;
  letter-spacing: -0.015em;
  color: var(--color-text-primary);
}

.service-catalog__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding: 0;
  margin: 0;
  list-style: none;
}

.service-catalog__item {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  padding: var(--space-4);
  cursor: pointer;
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: 4px;
}

.service-catalog__item:hover {
  border-color: var(--color-accent-brass);
}

.service-catalog__item:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-canvas),
    0 0 0 4px var(--color-focus);
}

/* La selección se marca con el ícono de check Y el borde de énfasis, nunca
   solo con un cambio de color (CA-091-04). */
.service-catalog__item--selected {
  border-color: var(--color-accent-brass);
  border-width: var(--border-width-emphasis);
}

.service-catalog__item-check {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  margin-top: 2px;
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: 50%;
}

.service-catalog__item--selected .service-catalog__item-check {
  background-color: var(--color-accent-brass);
  border-color: var(--color-accent-brass);
}

.service-catalog__item-body {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: var(--space-1);
  min-width: 0;
}

.service-catalog__item-name {
  font-weight: 600;
  color: var(--color-text-primary);
}

.service-catalog__item-description {
  color: var(--color-text-secondary);
  font-size: var(--font-size-body-sm);
  overflow-wrap: break-word;
}

.service-catalog__item-meta {
  display: flex;
  gap: var(--space-2);
  color: var(--color-text-secondary);
  font-size: var(--font-size-body-sm);
}

@media (min-width: 1024px) {
  .service-catalog__title {
    font-size: 40px;
    line-height: 48px;
  }
}
</style>
