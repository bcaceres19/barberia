<script setup lang="ts">
// Pantalla "Servicios por barbero" (HU-023): elegir un barbero y marcar qué
// servicios presta, mediante una casilla por servicio del catálogo. El
// mismo componente funciona con un barbero que presta todo el catálogo,
// varios con servicios compartidos, y un equipo con especialidades
// distintas (CA-023-01/02/03): nunca hay una rama especial según el número
// de barberos o servicios. Estados discriminados: carga inicial, listo,
// vacío (sin barberos o sin servicios), error recuperable, cargando
// asignaciones del barbero elegido. Cada casilla se deshabilita mientras su
// propia solicitud está en curso (evita doble envío, CA-023-08); el estado
// visual solo cambia después de la respuesta real del servidor -nunca
// antes- y un error recuperable revierte la casilla a su estado real sin
// perder la selección de barbero ni el resto de casillas ya marcadas. Se
// puede retirar a cualquier barbero, también al último de un servicio (DEC-114). Cada asignación o retiro
// confirmado añade un aviso emergente (DEC-095); el error sigue en línea.
import { computed, nextTick, ref, onMounted, onUnmounted, watch } from 'vue'
import { useToast, useVocabulary } from '@/shared/composables'
import { capitalize } from '@/shared/model'
import { BarberAvatar, BaseAlert, BaseButton, EmptyScene } from '@/shared/ui'
import {
  assignService,
  fetchAssignments,
  fetchBarberSummaries,
  fetchServiceSummaries,
  unassignService,
  type BarberSummary,
  type ServiceSummary,
} from '../api/barberServicesApi'

type PageStatus = 'loading' | 'ready' | 'load-error'
type AssignmentsStatus = 'idle' | 'loading' | 'ready' | 'error'

const pageStatus = ref<PageStatus>('loading')
const barbers = ref<BarberSummary[]>([])
const services = ref<ServiceSummary[]>([])

const selectedBarberId = ref<string | null>(null)
const assignmentsStatus = ref<AssignmentsStatus>('idle')
const assignedServiceIds = ref<Set<string>>(new Set())
const pendingServiceIds = ref<Set<string>>(new Set())
const toggleError = ref<string | null>(null)

const selectedBarber = computed(
  () => barbers.value.find((b) => b.id === selectedBarberId.value) ?? null,
)

// Cuántos servicios del catálogo presta la persona elegida: alimenta el
// medidor del panel lateral. Solo cuenta ids que existen en el catálogo, así
// una asignación a un servicio ya retirado del listado no infla el total.
const assignedCount = computed(
  () => services.value.filter((s) => assignedServiceIds.value.has(s.id)).length,
)
const assignedRatio = computed(() =>
  services.value.length === 0 ? 0 : assignedCount.value / services.value.length,
)

// Tabla paginada como Barberos y Servicios (DEC-107): el catálogo ya llega
// completo (selector de hasta 50), así que se pagina en el cliente y el tamaño
// de página es el número de filas que caben en el visor real.
const FALLBACK_PAGE_SIZE = 20
// En móvil la ficha de la persona va encima y deja poco alto: un piso evita
// una página de una sola fila; ahí el cascarón simplemente se desplaza.
const MIN_FIT_ROWS = 4
const currentPage = ref(1)
const pageSize = ref(FALLBACK_PAGE_SIZE)
const fitPending = ref(true)
const viewportWidth = ref(window.innerWidth)
const pageRef = ref<HTMLElement | null>(null)
const listRef = ref<HTMLElement | null>(null)
const footerRef = ref<HTMLElement | null>(null)
let resizeTimer: ReturnType<typeof setTimeout> | undefined
let observer: ResizeObserver | undefined

const totalPages = computed(() => Math.max(1, Math.ceil(services.value.length / pageSize.value)))
const pagedServices = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return services.value.slice(start, start + pageSize.value)
})
const countLabel = computed(
  () => `${services.value.length} ${services.value.length === 1 ? 'servicio' : 'servicios'}`,
)
const pageTokens = computed(() => {
  const total = totalPages.value
  const current = currentPage.value
  if (total <= 5) return Array.from({ length: total }, (_, i) => i + 1)
  const pages = new Set(
    viewportWidth.value <= 640
      ? [1, total, current]
      : [1, total, current, Math.max(1, current - 1), Math.min(total, current + 1)],
  )
  const tokens: (number | 'ellipsis')[] = []
  for (const p of [...pages].sort((a, b) => a - b)) {
    const last = tokens[tokens.length - 1]
    if (typeof last === 'number' && p - last > 1) tokens.push('ellipsis')
    tokens.push(p)
  }
  return tokens
})

function goToPage(page: number) {
  if (page < 1 || page > totalPages.value || page === currentPage.value) return
  const hadFocusInNav =
    document.activeElement instanceof HTMLElement &&
    document.activeElement.closest('.barber-services-page__pagination-nav') !== null
  currentPage.value = page
  if (!hadFocusInNav) return
  // Si el botón enfocado queda deshabilitado (primera/última página), el foco
  // se devuelve a la página vigente en vez de perderse en el <body>.
  void nextTick(() => {
    const active = document.activeElement
    if (active === document.body || active === null) {
      footerRef.value?.querySelector<HTMLButtonElement>('[aria-current="page"]')?.focus()
    }
  })
}

function scrollParent(el: HTMLElement): HTMLElement | null {
  let node = el.parentElement
  while (node) {
    if (['auto', 'scroll'].includes(getComputedStyle(node).overflowY)) return node
    node = node.parentElement
  }
  return null
}

// Mide el visor, la fila real y el pie; no oculta desborde para aparentar que
// caben filas. El padding cambia en móvil y también se mide.
async function fitPage() {
  await nextTick()
  const list = listRef.value
  const footer = footerRef.value
  const page = pageRef.value
  const parent = list ? scrollParent(list) : null
  const row = list?.querySelector<HTMLElement>('.barber-services-page__row')
  const height = row?.getBoundingClientRect().height
  if (!list || !footer || !page || !parent || !height) {
    fitPending.value = false
    return
  }
  const listRect = list.getBoundingClientRect()
  const footerRect = footer.getBoundingClientRect()
  const gap = Math.max(0, footerRect.top - listRect.bottom)
  const bottomPadding = parseFloat(getComputedStyle(page).paddingBottom)
  const available =
    parent.clientHeight -
    (listRect.top - parent.getBoundingClientRect().top + parent.scrollTop) -
    gap -
    footerRect.height -
    bottomPadding -
    2
  const fitting = Math.max(MIN_FIT_ROWS, Math.min(50, Math.floor(available / height)))
  if (fitting !== pageSize.value) {
    // Conserva la primera fila visible al cambiar el tamaño de página.
    const firstIndex = (currentPage.value - 1) * pageSize.value
    pageSize.value = fitting
    currentPage.value = Math.floor(firstIndex / fitting) + 1
  }
  parent.scrollTop = 0
  fitPending.value = false
}

function scheduleFit() {
  viewportWidth.value = window.innerWidth
  if (resizeTimer !== undefined) clearTimeout(resizeTimer)
  resizeTimer = setTimeout(() => {
    void fitPage()
  }, 120)
}

watch(
  listRef,
  async (el) => {
    if (!el) return
    await fitPage()
    if (typeof ResizeObserver === 'undefined') return
    observer?.disconnect()
    observer = new ResizeObserver(scheduleFit)
    const parent = scrollParent(el)
    if (parent) observer.observe(parent)
    const header = pageRef.value?.querySelector('.barber-services-page__header')
    if (header) observer.observe(header)
  },
  { flush: 'post' },
)

onMounted(() => window.addEventListener('resize', scheduleFit))
onUnmounted(() => {
  observer?.disconnect()
  window.removeEventListener('resize', scheduleFit)
  if (resizeTimer !== undefined) clearTimeout(resizeTimer)
})

async function loadPage() {
  pageStatus.value = 'loading'
  const [barbersOutcome, servicesOutcome] = await Promise.all([
    fetchBarberSummaries(),
    fetchServiceSummaries(),
  ])

  if (barbersOutcome.kind !== 'success' || servicesOutcome.kind !== 'success') {
    pageStatus.value = 'load-error'
    return
  }

  barbers.value = barbersOutcome.items
  services.value = servicesOutcome.items
  currentPage.value = 1
  pageStatus.value = 'ready'

  // El selector solo se muestra (y solo entonces tiene sentido pedir
  // asignaciones) cuando existe AL MENOS un barbero Y un servicio: con
  // cualquiera de las dos colecciones vacía, la plantilla muestra el estado
  // vacío correspondiente en vez del selector.
  if (barbers.value.length > 0 && services.value.length > 0) {
    await selectBarber(barbers.value[0]!.id)
  }
}

onMounted(loadPage)

function onRetryLoad() {
  void loadPage()
}

const toast = useToast()
// Palabras de la barbería (DEC-110): con los valores iniciales, el texto de siempre.
const v = useVocabulary()

async function selectBarber(barberId: string) {
  selectedBarberId.value = barberId
  toggleError.value = null
  assignmentsStatus.value = 'loading'

  const outcome = await fetchAssignments(barberId)
  // El barbero seleccionado pudo cambiar mientras la solicitud estaba en
  // vuelo (cambio rápido en el selector): descarta una respuesta obsoleta.
  if (selectedBarberId.value !== barberId) return

  if (outcome.kind === 'success') {
    assignedServiceIds.value = new Set(outcome.page.items.map((a) => a.serviceId))
    assignmentsStatus.value = 'ready'
    return
  }
  assignmentsStatus.value = 'error'
}

function onBarberSelectChange(event: Event) {
  const barberId = (event.target as HTMLSelectElement).value
  void selectBarber(barberId)
}

function onRetryAssignments() {
  if (selectedBarberId.value) void selectBarber(selectedBarberId.value)
}

function isPending(serviceId: string): boolean {
  return pendingServiceIds.value.has(serviceId)
}

function isAssigned(serviceId: string): boolean {
  return assignedServiceIds.value.has(serviceId)
}

function setPending(serviceId: string, pending: boolean) {
  const next = new Set(pendingServiceIds.value)
  if (pending) next.add(serviceId)
  else next.delete(serviceId)
  pendingServiceIds.value = next
}

function setAssigned(serviceId: string, assigned: boolean) {
  const next = new Set(assignedServiceIds.value)
  if (assigned) next.add(serviceId)
  else next.delete(serviceId)
  assignedServiceIds.value = next
}

async function onToggleService(service: ServiceSummary, event: Event) {
  const barberId = selectedBarberId.value
  const checkbox = event.target as HTMLInputElement

  if (!barberId || isPending(service.id)) {
    // Defensivo: el checkbox está deshabilitado mientras está pendiente
    // (evita doble envío en la práctica), pero si de algún modo llegara un
    // evento igual, se revierte de inmediato al estado real.
    checkbox.checked = isAssigned(service.id)
    return
  }

  const wantsAssigned = checkbox.checked

  setPending(service.id, true)
  toggleError.value = null

  const outcome = wantsAssigned
    ? await assignService(barberId, service.id)
    : await unassignService(barberId, service.id)

  setPending(service.id, false)

  if (outcome.kind === 'success') {
    setAssigned(service.id, wantsAssigned)
    toast.success(wantsAssigned ? 'Servicio asignado' : 'Servicio retirado', {
      detail: wantsAssigned
        ? `«${service.name}» quedó asignado a ${v.value.thisProfessional}.`
        : `«${service.name}» ya no está asignado a ${v.value.thisProfessional}.`,
    })
    return
  }

  // Ningún error recuperable cambia el estado real: se revierte la casilla
  // al valor que el servidor ya había confirmado antes de este intento.
  // Se asigna DIRECTAMENTE sobre el elemento del DOM que disparó el evento
  // (no solo la fuente reactiva `:checked`): si el valor reactivo no
  // cambió de contenido (por ejemplo, seguía asignado antes y sigue
  // asignado después de un rechazo), Vue no vuelve a tocar la propiedad
  // `checked` del elemento porque, desde su óptica, nada cambió -aunque el
  // navegador ya la haya alternado al hacer clic-; fijarla aquí garantiza
  // que la casilla siempre refleje el estado real, sin depender de esa
  // optimización de Vue.
  checkbox.checked = isAssigned(service.id)

  switch (outcome.kind) {
    case 'not-found':
      toggleError.value = `${capitalize(v.value.thisProfessional)} o servicio ya no está disponible. Recarga la página.`
      break
    case 'network-error':
      toggleError.value = 'No pudimos conectar. Revisa tu conexión e inténtalo de nuevo.'
      break
    case 'unexpected-error':
      toggleError.value = 'Ocurrió un error inesperado. Inténtalo de nuevo en unos segundos.'
      break
  }
}
</script>

<template>
  <section ref="pageRef" class="barber-services-page" aria-labelledby="barber-services-page-title">
    <header class="barber-services-page__header nv-rise">
      <h1 id="barber-services-page-title" class="barber-services-page__title">
        Servicios por {{ v.professional }}
      </h1>
      <p class="barber-services-page__subtitle">
        Gestiona los servicios que presta cada {{ v.professional }}.
      </p>
    </header>

    <div
      v-if="pageStatus === 'loading'"
      class="barber-services-page__state"
      role="status"
      aria-live="polite"
    >
      <p class="barber-services-page__state-text">Cargando {{ v.professionals }} y servicios…</p>
      <ul class="barber-services-page__skeleton" aria-hidden="true">
        <li v-for="n in 3" :key="n" class="barber-services-page__skeleton-row" />
      </ul>
    </div>

    <BaseAlert
      v-else-if="pageStatus === 'load-error'"
      variant="warning"
      title="No pudimos cargar esta sección"
      role="alert"
    >
      Revisa tu conexión e inténtalo de nuevo.
      <template #action>
        <BaseButton type="button" variant="secondary" @click="onRetryLoad">Reintentar</BaseButton>
      </template>
    </BaseAlert>

    <template v-else>
      <EmptyScene v-if="barbers.length === 0" scene="team" class="barber-services-page__empty">
        <template #title>Aún no tienes {{ v.professionalsRegistered }}.</template>
        <template #hint>
          Agrega {{ v.oneProfessional }} en la sección
          <RouterLink :to="{ name: 'staff-barberos' }">“{{ v.Professionals }}”</RouterLink>
          antes de asignarle servicios.
        </template>
      </EmptyScene>
      <EmptyScene
        v-else-if="services.length === 0"
        scene="services"
        class="barber-services-page__empty"
      >
        <template #title>Aún no tienes servicios en el catálogo.</template>
        <template #hint>
          Agrega uno en la sección
          <RouterLink :to="{ name: 'catalog-servicios' }">“Servicios”</RouterLink>
          antes de asignarlo a {{ v.aProfessional }}.
        </template>
      </EmptyScene>

      <div v-else class="barber-services-page__workspace">
        <!-- Ficha de la persona elegida: retrato, selector y medidor. Cambiar de
             persona reemplaza la ficha con una entrada breve (la clave fuerza el
             re-montaje) para que se lea como "otra persona", no como un dato
             que cambia de sitio. -->
        <aside class="barber-services-page__profile nv-rise" style="--i: 1">
          <div :key="selectedBarberId ?? 'none'" class="barber-services-page__identity nv-pop">
            <BarberAvatar
              v-if="selectedBarber"
              :full-name="selectedBarber.fullName"
              size="hero"
              class="barber-services-page__avatar"
            />
            <p class="barber-services-page__identity-name">{{ selectedBarber?.fullName }}</p>
          </div>

          <div class="barber-services-page__picker">
            <label for="barber-services-barber-select" class="barber-services-page__label">
              {{ v.Professional }}
            </label>
            <div class="barber-services-page__select-wrap">
              <select
                id="barber-services-barber-select"
                class="barber-services-page__select"
                :value="selectedBarberId ?? ''"
                @change="onBarberSelectChange"
              >
                <option v-for="barber in barbers" :key="barber.id" :value="barber.id">
                  {{ barber.fullName }}
                </option>
              </select>
            </div>
          </div>

          <div
            v-if="assignmentsStatus === 'ready'"
            class="barber-services-page__meter"
            role="group"
            aria-label="Servicios asignados"
          >
            <p class="barber-services-page__meter-figure">
              <span class="barber-services-page__meter-count">{{ assignedCount }}</span>
              <span class="barber-services-page__meter-total">/ {{ services.length }}</span>
              <span class="barber-services-page__meter-unit">servicios</span>
            </p>
            <span class="barber-services-page__meter-track" aria-hidden="true">
              <span
                class="barber-services-page__meter-fill"
                :style="{ transform: `scaleX(${assignedRatio})` }"
              />
            </span>
          </div>
        </aside>

        <div class="barber-services-page__panel">
          <div
            v-if="assignmentsStatus === 'loading'"
            class="barber-services-page__state"
            role="status"
            aria-live="polite"
          >
            <p class="barber-services-page__state-text">
              Cargando los servicios de {{ selectedBarber?.fullName }}…
            </p>
            <ul class="barber-services-page__skeleton" aria-hidden="true">
              <li
                v-for="n in Math.min(services.length, 4)"
                :key="n"
                class="barber-services-page__skeleton-row"
              />
            </ul>
          </div>

          <BaseAlert
            v-else-if="assignmentsStatus === 'error'"
            variant="warning"
            :title="`No pudimos cargar los servicios ${v.ofThisProfessional}`"
            role="alert"
          >
            Revisa tu conexión e inténtalo de nuevo.
            <template #action>
              <BaseButton type="button" variant="secondary" @click="onRetryAssignments">
                Reintentar
              </BaseButton>
            </template>
          </BaseAlert>

          <fieldset
            v-else-if="assignmentsStatus === 'ready'"
            :key="selectedBarberId ?? 'none'"
            class="barber-services-page__fieldset"
            :class="{ 'barber-services-page__fieldset--fitting': fitPending }"
          >
            <legend class="barber-services-page__legend">
              Servicios que presta {{ selectedBarber?.fullName }}
            </legend>

            <BaseAlert
              v-if="toggleError"
              variant="danger"
              title="No pudimos guardar el cambio"
              role="alert"
              class="barber-services-page__toggle-error"
            >
              {{ toggleError }}
            </BaseAlert>

            <div class="barber-services-page__table">
              <div class="barber-services-page__columns" aria-hidden="true">
                <span>Servicio</span><span>Asignación</span>
              </div>
              <ul
                ref="listRef"
                class="barber-services-page__list"
                aria-label="Catálogo de servicios"
              >
                <li
                  v-for="(service, index) in pagedServices"
                  :key="service.id"
                  class="barber-services-page__row nv-rise"
                  :class="{
                    'barber-services-page__row--assigned': isAssigned(service.id),
                    'barber-services-page__row--pending': isPending(service.id),
                  }"
                  :style="{ '--i': index }"
                >
                  <label
                    :for="`barber-services-service-${service.id}`"
                    class="barber-services-page__item-label"
                  >
                    <input
                      :id="`barber-services-service-${service.id}`"
                      type="checkbox"
                      class="barber-services-page__checkbox"
                      :checked="isAssigned(service.id)"
                      :disabled="isPending(service.id)"
                      @change="onToggleService(service, $event)"
                    />
                    <span class="barber-services-page__service-icon" aria-hidden="true">{{
                      service.name.trim().charAt(0).toUpperCase()
                    }}</span>
                    <span class="barber-services-page__service-name">{{ service.name }}</span>
                  </label>
                  <span
                    v-if="isPending(service.id)"
                    class="barber-services-page__pending"
                    aria-live="polite"
                  >
                    <span class="barber-services-page__spinner" aria-hidden="true" /> Guardando…
                  </span>
                  <span
                    v-else
                    :key="isAssigned(service.id) ? 'on' : 'off'"
                    class="barber-services-page__assignment"
                    :class="{
                      'barber-services-page__assignment--assigned': isAssigned(service.id),
                    }"
                  >
                    {{ isAssigned(service.id) ? 'Asignado' : 'No asignado' }}
                  </span>
                </li>
              </ul>
            </div>

            <div ref="footerRef" class="barber-services-page__footer">
              <span class="barber-services-page__count" role="status">{{ countLabel }}</span>
              <nav
                class="barber-services-page__pagination-nav"
                aria-label="Paginación de servicios"
              >
                <BaseButton
                  type="button"
                  variant="secondary"
                  aria-label="Página anterior"
                  :disabled="currentPage === 1"
                  @click="goToPage(currentPage - 1)"
                >
                  ‹
                </BaseButton>
                <template v-for="(token, index) in pageTokens" :key="`${token}-${index}`">
                  <span
                    v-if="token === 'ellipsis'"
                    class="barber-services-page__pagination-ellipsis"
                    aria-hidden="true"
                  >
                    …
                  </span>
                  <BaseButton
                    v-else
                    type="button"
                    :variant="token === currentPage ? 'primary' : 'secondary'"
                    :aria-label="`Página ${token}`"
                    :aria-current="token === currentPage ? 'page' : undefined"
                    @click="goToPage(token)"
                  >
                    {{ token }}
                  </BaseButton>
                </template>
                <BaseButton
                  type="button"
                  variant="secondary"
                  aria-label="Página siguiente"
                  :disabled="currentPage === totalPages"
                  @click="goToPage(currentPage + 1)"
                >
                  ›
                </BaseButton>
              </nav>
            </div>
          </fieldset>
        </div>
      </div>
    </template>
  </section>
</template>

<style scoped>
/* Lienzo de tinta de punta a punta, igual que Agenda, Servicios y Barberos:
   esta pantalla se había quedado en la superficie clara previa al rediseño.
   Composición propia (identidad guiada): a la izquierda la ficha de la
   persona con su medidor, a la derecha el catálogo como filas de
   interruptor; en móvil la ficha pasa arriba y el catálogo debajo. */
.barber-services-page {
  --barber-services-width: 920px;
  --barber-services-row-height: 72px;

  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 100%;
  padding: 34px 32px 48px;
  color: var(--color-on-strong);
  /* Transparente: la tinta y el fondo animado los pone el cascarón. */
  background: transparent;
}

.barber-services-page__header,
.barber-services-page__state,
.barber-services-page__empty,
.barber-services-page > :deep(.base-alert),
.barber-services-page__workspace {
  width: min(100%, var(--barber-services-width));
  margin-inline: auto;
}

.barber-services-page__header {
  padding-bottom: 16px;
  border-bottom: var(--border-width-normal) solid var(--color-field-strong-border);
}

.barber-services-page__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h1);
  line-height: var(--font-size-h1-line);
  font-weight: var(--font-weight-h1);
}

@media (min-width: 1024px) {
  .barber-services-page__title {
    font-size: var(--font-size-title-page);
    line-height: 46px;
  }
}

.barber-services-page__subtitle {
  margin: 4px 0 0;
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
  color: var(--color-on-strong-muted);
}

.barber-services-page__state {
  padding: var(--space-4) 0;
  color: var(--color-on-strong-muted);
}

.barber-services-page__state-text {
  margin: 0 0 var(--space-3);
}

/* Esqueleto: filas del tamaño de las reales con el brillo que cruza de
   izquierda a derecha, para que la carga ocupe el lugar del contenido. */
.barber-services-page__skeleton {
  display: flex;
  flex-direction: column;
  gap: 1px;
  padding: 0;
  margin: 0;
  list-style: none;
  overflow: hidden;
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: 3px;
}

.barber-services-page__skeleton-row {
  height: 68px;
  background: linear-gradient(
    100deg,
    transparent 30%,
    color-mix(in srgb, var(--color-on-strong) 8%, transparent) 50%,
    transparent 70%
  );
  background-size: 200% 100%;
  animation: nava-shimmer 1.6s linear infinite;
}

.barber-services-page__workspace {
  display: grid;
  align-items: start;
  gap: 24px;
  grid-template-columns: minmax(0, 1fr);
}

@media (min-width: 900px) {
  .barber-services-page__workspace {
    grid-template-columns: 280px minmax(0, 1fr);
    gap: 32px;
  }

  .barber-services-page__profile {
    position: sticky;
    top: 0;
  }
}

/* Ficha de la persona */
.barber-services-page__profile {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 20px;
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-top: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  border-radius: 3px;
}

.barber-services-page__identity {
  display: flex;
  align-items: center;
  gap: 16px;
}

.barber-services-page__identity-name {
  margin: 0;
  min-width: 0;
  overflow-wrap: anywhere;
  font-family: var(--font-display);
  font-size: var(--font-size-title-item);
  line-height: var(--font-size-title-item-line);
  color: var(--color-on-strong);
}

@media (max-width: 899px) {
  .barber-services-page .barber-services-page__avatar {
    --avatar-size: 64px;
  }
}

@media (min-width: 900px) {
  .barber-services-page__identity {
    flex-direction: column;
    align-items: flex-start;
  }
}

.barber-services-page__picker {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.barber-services-page__label {
  font-family: var(--font-sans);
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--color-brand-accent-surface);
}

.barber-services-page__select-wrap {
  position: relative;
}

/* Rombo de despliegue dibujado con bordes: el <select> nativo conserva su
   semántica y su teclado; solo se le quita la apariencia del sistema. */
.barber-services-page__select-wrap::after {
  content: '';
  position: absolute;
  top: 50%;
  right: 16px;
  width: 8px;
  height: 8px;
  border-right: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  border-bottom: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  pointer-events: none;
  transform: translateY(-70%) rotate(45deg);
  transition: transform var(--motion-duration-base) var(--motion-ease-out);
}

.barber-services-page__select-wrap:focus-within::after {
  transform: translateY(-30%) rotate(225deg);
}

.barber-services-page__select {
  appearance: none;
  width: 100%;
  min-height: var(--control-height);
  padding: 0 44px 0 var(--space-4);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-on-strong);
  cursor: pointer;
  background-color: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-bottom: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  border-radius: 2px;
  transition:
    background-color var(--motion-duration-fast) var(--motion-easing-standard),
    border-color var(--motion-duration-fast) var(--motion-easing-standard);
}

.barber-services-page__select option {
  color: var(--color-text-primary);
  background-color: var(--color-surface);
}

.barber-services-page__select:hover {
  background-color: color-mix(in srgb, var(--color-on-strong) 8%, transparent);
}

.barber-services-page__select:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus);
}

/* Medidor: cifra serif grande y una regla de latón que se llena. */
.barber-services-page__meter {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding-top: 16px;
  border-top: var(--border-width-normal) solid var(--color-field-strong-border);
}

.barber-services-page__meter-figure {
  display: flex;
  align-items: baseline;
  gap: 6px;
  margin: 0;
}

.barber-services-page__meter-count {
  font-family: var(--font-display);
  font-size: var(--font-size-title-page);
  line-height: 1;
  color: var(--color-brand-accent-surface);
  font-variant-numeric: tabular-nums;
}

.barber-services-page__meter-total {
  font-family: var(--font-display);
  font-size: var(--font-size-title-item);
  color: var(--color-on-strong-muted);
  font-variant-numeric: tabular-nums;
}

.barber-services-page__meter-unit {
  margin-left: 2px;
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--color-on-strong-muted);
}

.barber-services-page__meter-track {
  display: block;
  height: 4px;
  overflow: hidden;
  background-color: color-mix(in srgb, var(--color-on-strong) 12%, transparent);
  border-radius: 2px;
}

.barber-services-page__meter-fill {
  display: block;
  width: 100%;
  height: 100%;
  background-color: var(--color-brand-accent-surface);
  transform-origin: left center;
  transition: transform 520ms var(--motion-ease-out);
}

/* Catálogo */
.barber-services-page__fieldset {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 0;
  margin: 0;
  min-width: 0;
  border: none;
}

.barber-services-page__legend {
  padding: 0;
  margin-bottom: 14px;
  font-family: var(--font-display);
  font-size: var(--font-size-title-item);
  line-height: var(--font-size-title-item-line);
  color: var(--color-on-strong);
}

.barber-services-page__toggle-error {
  margin: 0;
}

/* Mientras se mide cuántas filas caben, el contenido ocupa su lugar pero no se
   ve: así la tabla no salta de tamaño ante la persona. */
.barber-services-page__fieldset--fitting {
  visibility: hidden;
}

/* Encabezado reglado: versalitas espaciadas de latón, igual que Barberos y
   Servicios. Comparte columnas con la fila. */
.barber-services-page__columns,
.barber-services-page__row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 140px;
  align-items: center;
  gap: 12px;
}

.barber-services-page__columns {
  min-height: 44px;
  padding: 0 16px;
  font-family: var(--font-sans);
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--color-brand-accent-surface);
}

.barber-services-page__columns span:last-child {
  text-align: right;
}

.barber-services-page__list {
  display: flex;
  flex-direction: column;
  padding: 0;
  margin: 0;
  list-style: none;
  overflow: hidden;
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: 3px;
}

.barber-services-page__row {
  position: relative;
  height: var(--barber-services-row-height);
  padding: 0 16px;
  border-top: var(--border-width-normal) solid var(--color-field-strong-border);
  transition: background-color var(--motion-duration-base) var(--motion-easing-standard);
}

.barber-services-page__row:first-child {
  border-top: none;
}

/* Filete de latón que se dibuja en el borde izquierdo al asignar. */
.barber-services-page__row::before {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: 3px;
  background-color: var(--color-brand-accent-surface);
  transform: scaleY(0);
  transform-origin: center;
  transition: transform var(--motion-duration-base) var(--motion-ease-out);
}

.barber-services-page__row--assigned {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 7%, transparent);
}

.barber-services-page__row--assigned::before {
  transform: scaleY(1);
}

@media (hover: hover) {
  .barber-services-page__row:hover {
    background-color: color-mix(in srgb, var(--color-on-strong) 5%, transparent);
  }

  .barber-services-page__row--assigned:hover {
    background-color: color-mix(in srgb, var(--color-brand-accent-surface) 11%, transparent);
  }
}

/* El label ENVUELVE casilla y nombre; su ::after cubre toda la fila para que
   cualquier clic o toque alterne el servicio (objetivo de 72px de alto). */
.barber-services-page__item-label {
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
  height: 100%;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-lg);
  font-weight: 500;
  color: var(--color-on-strong);
  cursor: pointer;
}

.barber-services-page__item-label::after {
  content: '';
  position: absolute;
  inset: 0;
}

/* Casilla propia: cuadro de latón que se rellena y dibuja la marca. */
.barber-services-page__checkbox {
  position: relative;
  appearance: none;
  flex: 0 0 auto;
  width: 24px;
  height: 24px;
  margin: 0;
  cursor: pointer;
  background-color: transparent;
  border: var(--border-width-emphasis) solid
    color-mix(in srgb, var(--color-brand-accent-surface) 70%, transparent);
  border-radius: 2px;
  transition:
    background-color var(--motion-duration-fast) var(--motion-easing-standard),
    border-color var(--motion-duration-fast) var(--motion-easing-standard),
    transform var(--motion-duration-base) var(--motion-ease-spring);
}

.barber-services-page__checkbox::after {
  content: '';
  position: absolute;
  top: 3px;
  left: 7px;
  width: 6px;
  height: 11px;
  border-right: var(--border-width-emphasis) solid var(--color-brand-accent-text);
  border-bottom: var(--border-width-emphasis) solid var(--color-brand-accent-text);
  opacity: 0;
  transform: rotate(45deg) scale(0.4);
  transition:
    opacity var(--motion-duration-fast) var(--motion-easing-standard),
    transform var(--motion-duration-base) var(--motion-ease-spring);
}

.barber-services-page__checkbox:checked {
  background-color: var(--color-brand-accent-surface);
  border-color: var(--color-brand-accent-surface);
  transform: scale(1.06);
}

.barber-services-page__checkbox:checked::after {
  opacity: 1;
  transform: rotate(45deg) scale(1);
}

.barber-services-page__checkbox:disabled {
  cursor: progress;
  opacity: 0.6;
}

.barber-services-page__checkbox:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-field-strong),
    0 0 0 4px var(--color-focus);
}

.barber-services-page__service-icon {
  display: grid;
  flex: 0 0 auto;
  place-items: center;
  width: 36px;
  height: 36px;
  font-family: var(--font-display);
  font-size: var(--font-size-body-lg);
  line-height: 1;
  color: var(--color-brand-accent-surface);
  background-color: var(--color-field-strong-raised);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: 2px;
}

.barber-services-page__service-name {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: color var(--motion-duration-fast) var(--motion-easing-standard);
}

.barber-services-page__row--assigned .barber-services-page__service-name {
  color: var(--color-brand-accent-surface);
}

/* Estado como rombo con rótulo (mismo lenguaje que Servicios y Barberos). */
.barber-services-page__assignment,
.barber-services-page__pending {
  position: relative;
  z-index: 1;
  justify-self: end;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-family: var(--font-family-base);
  font-size: var(--font-size-caption);
  font-weight: 600;
  line-height: var(--font-size-caption-line);
  letter-spacing: 0.08em;
  text-transform: uppercase;
  white-space: nowrap;
  color: var(--color-on-strong-muted);
  pointer-events: none;
}

.barber-services-page__assignment {
  animation: nava-pop 360ms var(--motion-ease-spring) backwards;
}

.barber-services-page__assignment::before {
  content: '';
  width: 6px;
  height: 6px;
  background-color: currentColor;
  transform: rotate(45deg);
}

.barber-services-page__assignment--assigned {
  color: var(--color-success-on-strong);
}

/* Pie de la tabla: conteo a la izquierda, paginador a la derecha. */
.barber-services-page__footer {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.barber-services-page__count {
  font-size: var(--font-size-body-sm);
  color: var(--color-on-strong-muted);
}

.barber-services-page__pagination-nav {
  display: flex;
  align-items: center;
  gap: 6px;
}

.barber-services-page__pagination-ellipsis {
  padding-inline: 4px;
  color: var(--color-on-strong-muted);
}

/* Mismo par dorado-sólido (página vigente) / tinta-fantasma (el resto) que el
   paginador de Barberos y Servicios. */
.barber-services-page__pagination-nav :deep(.base-button) {
  height: 38px;
  min-width: 38px;
  padding-inline: 10px;
  font-size: var(--font-size-body-sm);
  --btn-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);
}

.barber-services-page__pagination-nav :deep(.base-button--primary) {
  background-color: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  border-color: var(--color-brand-accent-surface);
}

.barber-services-page__pagination-nav
  :deep(.base-button--primary:hover:not(:disabled):not(.base-button--loading)) {
  filter: brightness(92%);
}

.barber-services-page__pagination-nav :deep(.base-button--secondary) {
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.barber-services-page__pagination-nav
  :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
}

.barber-services-page__pagination-nav :deep(.base-button:disabled) {
  opacity: 0.4;
}

.barber-services-page__spinner {
  width: 12px;
  height: 12px;
  border: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  border-top-color: transparent;
  border-radius: 50%;
  animation: barber-services-spin 800ms linear infinite;
}

@keyframes barber-services-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .barber-services-page__assignment,
  .barber-services-page__spinner,
  .barber-services-page__skeleton-row {
    animation: none;
  }

  .barber-services-page__checkbox,
  .barber-services-page__checkbox::after,
  .barber-services-page__row::before,
  .barber-services-page__meter-fill,
  .barber-services-page__select-wrap::after {
    transition: none;
  }
}

@media (max-width: 640px) {
  .barber-services-page {
    gap: 12px;
    padding: 20px 16px 28px;
  }

  .barber-services-page__profile {
    padding: 16px;
  }

  .barber-services-page {
    --barber-services-row-height: 64px;
  }

  /* Sin encabezado en móvil (como Barberos): la fila ya se explica sola. */
  .barber-services-page__columns {
    display: none;
  }

  .barber-services-page__row {
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 8px;
    padding: 0 12px;
  }

  .barber-services-page__item-label {
    gap: 10px;
    font-size: var(--font-size-body);
  }

  .barber-services-page__footer {
    justify-content: center;
    text-align: center;
  }

  .barber-services-page__service-icon {
    display: none;
  }
}
</style>
