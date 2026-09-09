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
// antes- y un error recuperable (incluida la última asignación activa,
// DEC-068) revierte la casilla a su estado real sin perder la selección de
// barbero ni el resto de casillas ya marcadas.
import { computed, ref, onMounted } from 'vue'
import { BaseAlert, BaseButton, PageHeader } from '@/shared/ui'
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
    case 'last-active-conflict':
      toggleError.value = `No puedes retirar "${service.name}": es el único barbero asignado a este servicio activo. Asigna otro barbero antes de retirar este.`
      break
    case 'not-found':
      toggleError.value = 'Este barbero o servicio ya no está disponible. Recarga la página.'
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
  <section class="barber-services-page" aria-labelledby="barber-services-page-title">
    <PageHeader
      title-id="barber-services-page-title"
      title="Servicios por barbero"
      subtitle="Gestiona los servicios que presta cada barbero."
    />

    <div
      v-if="pageStatus === 'loading'"
      class="barber-services-page__state"
      role="status"
      aria-live="polite"
    >
      <p>Cargando barberos y servicios…</p>
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
      <p v-if="barbers.length === 0" class="barber-services-page__empty">
        Aún no tienes barberos registrados. Agrega uno en la sección “Barberos” antes de asignarle
        servicios.
      </p>
      <p v-else-if="services.length === 0" class="barber-services-page__empty">
        Aún no tienes servicios en el catálogo. Agrega uno en la sección “Servicios” antes de
        asignarlo a un barbero.
      </p>

      <template v-else>
        <div class="barber-services-page__picker">
          <label for="barber-services-barber-select" class="barber-services-page__label">
            Barbero
          </label>
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

        <div
          v-if="assignmentsStatus === 'loading'"
          class="barber-services-page__state"
          role="status"
          aria-live="polite"
        >
          <p>Cargando los servicios de {{ selectedBarber?.fullName }}…</p>
        </div>

        <BaseAlert
          v-else-if="assignmentsStatus === 'error'"
          variant="warning"
          title="No pudimos cargar los servicios de este barbero"
          role="alert"
        >
          Revisa tu conexión e inténtalo de nuevo.
          <template #action>
            <BaseButton type="button" variant="secondary" @click="onRetryAssignments">
              Reintentar
            </BaseButton>
          </template>
        </BaseAlert>

        <fieldset v-else-if="assignmentsStatus === 'ready'" class="barber-services-page__fieldset">
          <legend class="barber-services-page__legend">
            Servicios que presta {{ selectedBarber?.fullName }}
          </legend>

          <BaseAlert
            v-if="toggleError"
            variant="danger"
            title="No puedes retirar la última asignación activa de este servicio"
            role="alert"
            class="barber-services-page__toggle-error"
          >
            {{ toggleError }}
          </BaseAlert>

          <ul class="barber-services-page__list" aria-label="Catálogo de servicios">
            <li v-for="service in services" :key="service.id" class="barber-services-page__row">
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
                <span class="barber-services-page__service-icon" aria-hidden="true">▧</span>
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
                class="barber-services-page__assignment"
                :class="{ 'barber-services-page__assignment--assigned': isAssigned(service.id) }"
              >
                {{ isAssigned(service.id) ? 'Asignado' : 'No asignado' }}
              </span>
            </li>
          </ul>
        </fieldset>
      </template>
    </template>
  </section>
</template>

<style scoped>
.barber-services-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  max-width: 640px;
  padding: var(--space-4);
  margin: 0 auto;
}

.barber-services-page__state {
  padding: var(--space-4);
  color: var(--color-text-secondary);
}

.barber-services-page__empty {
  padding: var(--space-4);
  color: var(--color-text-secondary);
}

.barber-services-page__picker {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.barber-services-page__label {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  font-weight: 500;
  color: var(--color-text-primary);
}

.barber-services-page__select {
  min-height: 44px;
  padding: var(--space-2) var(--space-3);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-text-primary);
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: var(--radius-md);
}

.barber-services-page__fieldset {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding: 0;
  margin: 0;
  border: none;
}

.barber-services-page__legend {
  padding: 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  font-weight: 500;
  color: var(--color-text-primary);
}

.barber-services-page__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding: 0;
  margin: 0;
  list-style: none;
}

.barber-services-page__checkbox {
  width: 24px;
  height: 24px;
  min-width: 24px;
  flex-shrink: 0;
  accent-color: var(--color-action-primary);
}

/* El label ENVUELVE el checkbox y el nombre: es el objetivo táctil real de
   al menos 44x44px (estandar-diseno-visual.md), no solo el checkbox visual
   de 24x24. */
.barber-services-page__item-label {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-3);
  min-height: 44px;
  flex: 1 1 auto;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-text-primary);
  cursor: pointer;
  overflow-wrap: anywhere;
}

.barber-services-page__pending {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  color: var(--color-text-secondary);
}

.barber-services-page__toggle-error {
  margin: 0;
}
</style>

<style scoped>
.barber-services-page {
  --barber-services-width: 478px;
  gap: 16px;
  min-height: 100%;
  max-width: none;
  padding: 34px 32px 48px;
  color: var(--color-text-primary);
  background: var(--color-surface);
}

.barber-services-page > :deep(.page-header),
.barber-services-page__state,
.barber-services-page__empty,
.barber-services-page > :deep(.base-alert),
.barber-services-page__picker,
.barber-services-page__fieldset {
  width: min(100%, var(--barber-services-width));
  margin-inline: auto;
}

.barber-services-page > :deep(.page-header) {
  padding-bottom: 16px;
  margin-bottom: 0;
}

.barber-services-page :deep(.page-header__title) {
  font-size: 30px;
  line-height: 1.14;
}

.barber-services-page :deep(.page-header__subtitle) {
  font-size: 11px;
  line-height: 16px;
}

.barber-services-page__picker {
  gap: 6px;
}

.barber-services-page__label {
  font-size: 10px;
  font-weight: 600;
}

.barber-services-page__select {
  min-height: 32px;
  padding: 6px 10px;
  font-size: 11px;
  border-radius: 4px;
}

.barber-services-page__fieldset {
  gap: 12px;
}

.barber-services-page__legend {
  font-family: var(--font-display);
  font-size: 17px;
}

.barber-services-page__list {
  gap: 0;
  overflow: hidden;
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: 3px;
}

.barber-services-page__row {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 71px;
  padding: 11px 12px;
  border-bottom: var(--border-width-normal) solid var(--color-border-subtle);
}

.barber-services-page__row:last-child {
  border-bottom: none;
}

.barber-services-page__checkbox {
  width: 18px;
  height: 18px;
  min-width: 18px;
}

.barber-services-page__item-label {
  flex-wrap: nowrap;
  gap: 11px;
  min-width: 0;
  flex: 1;
  font-size: 12px;
  font-weight: 500;
}

.barber-services-page__service-icon {
  flex: 0 0 auto;
  color: var(--color-text-secondary);
  font-size: 20px;
  font-weight: 400;
  line-height: 1;
}

.barber-services-page__service-name {
  min-width: 0;
  overflow-wrap: anywhere;
}

.barber-services-page__assignment,
.barber-services-page__pending {
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  min-height: 22px;
  padding: 3px 8px;
  border-radius: 3px;
  font-family: var(--font-family-base);
  font-size: 10px;
  line-height: 1.2;
  color: var(--color-text-secondary);
  background: var(--color-surface-muted);
}

.barber-services-page__assignment--assigned {
  color: var(--color-success-text);
  background: var(--color-success-surface);
}

.barber-services-page__pending {
  gap: 6px;
  background: transparent;
}

.barber-services-page__spinner {
  width: 11px;
  height: 11px;
  border: 1px solid var(--color-border-control);
  border-top-color: transparent;
  border-radius: 50%;
}

.barber-services-page__toggle-error {
  margin: 0 0 4px;
}

@media (max-width: 640px) {
  .barber-services-page {
    gap: 12px;
    padding: 16px 16px 28px;
  }

  .barber-services-page > :deep(.page-header) {
    padding-bottom: 10px;
  }

  .barber-services-page :deep(.page-header__title) {
    font-size: 21px;
  }

  .barber-services-page :deep(.page-header__subtitle) {
    font-size: 10px;
  }

  .barber-services-page__legend {
    font-size: 12px;
  }

  .barber-services-page__row {
    min-height: 62px;
    gap: 8px;
    padding: 8px 9px;
  }

  .barber-services-page__item-label {
    gap: 8px;
    font-size: 11px;
  }

  .barber-services-page__service-icon {
    display: none;
  }

  .barber-services-page__assignment,
  .barber-services-page__pending {
    padding: 3px 6px;
    font-size: 8px;
  }
}
</style>
