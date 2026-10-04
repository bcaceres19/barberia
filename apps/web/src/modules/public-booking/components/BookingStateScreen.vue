<script setup lang="ts">
// Pantalla de estado de página de los pasos públicos: carga, enlace no
// encontrado, error de red y error inesperado. Las cuatro pantallas de la
// reserva repetían este bloque al pie de la letra (mismo copy, misma
// jerarquía de roles); aquí vive una sola vez y cada página solo aporta su
// título oculto y el texto de carga. `PageState` ya compone el divisor, el
// rótulo y el titular sobre tinta; este componente solo le da a la acción
// (Reintentar) el tratamiento de latón de la reserva, porque `BaseButton`
// sin ajustar es blanco o tinta sobre tinta.
import { computed } from 'vue'
import { BaseButton, PageState } from '@/shared/ui'

interface Props {
  status: 'loading' | 'not-found' | 'network-error' | 'unexpected-error'
  /** Título `<h1>` accesible pero oculto: `PageState` solo aporta un `<h2>`
   * y estas pantallas no tienen otro `<h1>` (page-has-heading-one, axe). */
  title: string
  loadingHeadline: string
  requestId?: string
}

const props = defineProps<Props>()
const emit = defineEmits<{ retry: [] }>()

const unexpectedErrorMessage = computed(() =>
  props.requestId
    ? `Inténtalo de nuevo. Si continúa, comparte este código con soporte: ${props.requestId}.`
    : 'Inténtalo de nuevo en unos segundos.',
)
</script>

<template>
  <main class="pb-page pb-page--state">
    <h1 class="pb-sr-only">{{ title }}</h1>
    <PageState
      v-if="status === 'loading'"
      class="pb-state"
      variant="loading"
      :headline="loadingHeadline"
      role="status"
    />
    <PageState
      v-else-if="status === 'not-found'"
      class="pb-state pb-state--danger"
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
        <BaseButton variant="secondary" @click="emit('retry')">Reintentar</BaseButton>
      </template>
    </PageState>
    <PageState
      v-else-if="status === 'network-error'"
      class="pb-state pb-state--warning"
      variant="warning"
      status-label="Atención"
      headline="No pudimos conectar"
      role="alert"
    >
      <template #default>Revisa tu conexión e inténtalo de nuevo.</template>
      <template #action>
        <BaseButton variant="primary" @click="emit('retry')">Reintentar</BaseButton>
      </template>
    </PageState>
    <PageState
      v-else
      class="pb-state pb-state--danger"
      variant="danger"
      status-label="Error"
      headline="Ocurrió un error inesperado"
      role="alert"
    >
      <template #default>{{ unexpectedErrorMessage }}</template>
      <template #action>
        <BaseButton variant="primary" @click="emit('retry')">Reintentar</BaseButton>
      </template>
    </PageState>
  </main>
</template>

<style scoped>
/* El estado entra con la misma subida suave que el resto de la reserva. */
.pb-state {
  animation: pb-rise 0.6s var(--pb-ease, ease) backwards;
}

/* `PageState` pinta el rótulo y el divisor con el rojo/ámbar de las
   superficies claras (--color-danger-action, --color-warning-border), que
   sobre tinta no llegan a 3:1. Aquí se sustituyen por los tintes LEVANTADOS
   al lienzo de tinta que el sistema ya define para ese fin. */
.pb-state--danger :deep(.page-state__status) {
  color: var(--color-danger-on-strong);
}

.pb-state--danger :deep(.page-state__divider),
.pb-state--danger :deep(.page-state__divider)::after {
  background-color: var(--color-danger-on-strong);
}

.pb-state--warning :deep(.page-state__status) {
  color: var(--color-warning-on-strong);
}

.pb-state--warning :deep(.page-state__divider),
.pb-state--warning :deep(.page-state__divider)::after {
  background-color: var(--color-warning-on-strong);
}

/* Reintentar: latón sobre tinta. La primaria se llena; la secundaria queda
   en contorno con el filete inferior acentuado del lenguaje reglado. */
.pb-state :deep(.base-button) {
  min-width: 148px;
  height: var(--control-height-primary-mobile);
  font-weight: 600;
}

.pb-state :deep(.base-button--primary) {
  color: var(--color-brand-accent-text);
  background-color: var(--color-brand-accent-surface);
  border-color: var(--color-brand-accent-surface);
  border-bottom-width: var(--border-width-emphasis);
  border-bottom-color: color-mix(in srgb, var(--color-brand-accent-surface) 70%, #000);
}

.pb-state :deep(.base-button--primary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 86%, #fff);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 86%, #fff);
}

.pb-state :deep(.base-button--secondary) {
  color: var(--color-brand-accent-surface);
  background-color: transparent;
  border-color: var(--color-brand-accent-surface);
  border-bottom-width: var(--border-width-emphasis);
}

.pb-state :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 14%, transparent);
}

.pb-state :deep(.base-button:focus-visible) {
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-on-strong);
}

@media (prefers-reduced-motion: reduce) {
  .pb-state {
    animation: none;
  }
}
</style>
