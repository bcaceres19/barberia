<script setup lang="ts">
/**
 * AccordionItem - Un apartado plegable del panel lateral de Horarios. El
 * encabezado es un botón con `aria-expanded` dentro de un `h2` (patrón WAI-ARIA
 * de acordeón); el cuerpo es una región etiquetada por ese botón. Un cuerpo
 * cerrado queda con `visibility: hidden`, así que no recibe foco ni se lee con
 * lector de pantalla, pero su altura se anima (filas de grid 0fr → 1fr) en vez
 * de saltar. Quien lo monta decide cuál está abierto (un solo apartado a la
 * vez), así el panel nunca crece más que la pantalla.
 */
interface Props {
  /** Prefijo de los ids del botón y de la región (`<id>-trigger`, `<id>-panel`). */
  id: string
  title: string
  open: boolean
  /** Dato corto junto al título: un conteo o el estado de un ajuste. */
  badge?: string
  badgeTone?: 'neutral' | 'on' | 'off'
}

withDefaults(defineProps<Props>(), { badge: undefined, badgeTone: 'neutral' })

defineEmits<{ toggle: [] }>()
</script>

<template>
  <section class="accordion-item" :class="{ 'accordion-item--open': open }">
    <h2 class="accordion-item__heading">
      <button
        :id="`${id}-trigger`"
        type="button"
        class="accordion-item__trigger"
        :aria-expanded="open"
        :aria-controls="`${id}-panel`"
        @click="$emit('toggle')"
      >
        <span class="accordion-item__title">{{ title }}</span>
        <span
          v-if="badge"
          class="accordion-item__badge"
          :class="`accordion-item__badge--${badgeTone}`"
          >{{ badge }}</span
        >
        <svg
          class="accordion-item__chevron"
          viewBox="0 0 12 12"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
          aria-hidden="true"
        >
          <path d="M2.5 4.5 6 8l3.5-3.5" />
        </svg>
      </button>
    </h2>
    <div
      :id="`${id}-panel`"
      class="accordion-item__panel"
      role="region"
      :aria-labelledby="`${id}-trigger`"
    >
      <div class="accordion-item__clip">
        <div class="accordion-item__body">
          <slot />
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.accordion-item {
  position: relative;
  border-top: var(--border-width-normal) solid var(--color-field-strong-border);
  transition: background-color var(--motion-duration-base) var(--motion-easing-standard);
}

.accordion-item:first-child {
  border-top: none;
}

/* Filete de latón del apartado abierto: crece desde arriba al abrirse. */
.accordion-item::before {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: 3px;
  background-color: var(--color-brand-accent-surface);
  transform: scaleY(0);
  transform-origin: top;
  transition: transform var(--motion-duration-base) var(--motion-easing-standard);
}

.accordion-item--open {
  background-color: color-mix(in srgb, var(--color-on-strong) 3%, transparent);
}

.accordion-item--open::before {
  transform: scaleY(1);
}

.accordion-item__heading {
  margin: 0;
}

.accordion-item__trigger {
  display: flex;
  width: 100%;
  min-height: 54px;
  align-items: center;
  gap: 10px;
  padding: 8px 16px;
  color: var(--color-on-strong);
  text-align: left;
  background: transparent;
  border: 0;
  cursor: pointer;
  transition: background-color var(--motion-duration-fast) var(--motion-easing-standard);
}

@media (hover: hover) {
  .accordion-item__trigger:hover {
    background-color: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  }
}

.accordion-item__trigger:focus-visible {
  outline: none;
  box-shadow:
    inset 0 0 0 2px var(--color-field-strong),
    inset 0 0 0 4px var(--color-brand-accent-surface);
}

.accordion-item__title {
  flex: 1;
  min-width: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-body-lg);
  font-weight: 400;
  line-height: 22px;
}

.accordion-item__badge {
  flex: 0 0 auto;
  padding: 2px 8px;
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  font-variant-numeric: tabular-nums;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-on-strong) 20%, transparent);
  border-radius: 2px;
  transition:
    color var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard),
    background-color var(--motion-duration-base) var(--motion-easing-standard);
}

.accordion-item__badge--on {
  color: var(--color-success-on-strong);
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 10%, transparent);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
}

.accordion-item--open .accordion-item__badge--neutral {
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
}

.accordion-item__chevron {
  flex: 0 0 auto;
  width: 14px;
  height: 14px;
  color: var(--color-brand-accent-surface);
  transition: transform var(--motion-duration-base) cubic-bezier(0.34, 1.56, 0.64, 1);
}

.accordion-item--open .accordion-item__chevron {
  transform: rotate(180deg);
}

/* La altura se anima con filas de grid (0fr → 1fr); la visibilidad se retrasa
   al cerrar para que el contenido no desaparezca antes de plegarse. */
.accordion-item__panel {
  display: grid;
  grid-template-rows: 0fr;
  visibility: hidden;
  transition:
    grid-template-rows 340ms var(--motion-easing-standard),
    visibility 0s linear 340ms;
}

.accordion-item--open .accordion-item__panel {
  grid-template-rows: 1fr;
  visibility: visible;
  transition-delay: 0s;
}

.accordion-item__clip {
  min-height: 0;
  overflow: hidden;
}

/* --accordion-body-max la define el panel lateral: en escritorio es lo que
   queda de la altura disponible; en pantallas angostas no hay tope. */
.accordion-item__body {
  display: flex;
  max-height: var(--accordion-body-max, none);
  flex-direction: column;
  gap: 12px;
  padding: 4px 16px 16px;
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-color: color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent)
    transparent;
  scrollbar-width: thin;
}

.accordion-item--open .accordion-item__body {
  animation: accordion-item-body-in 380ms var(--motion-easing-standard) 90ms both;
}

@keyframes accordion-item-body-in {
  from {
    opacity: 0;
    transform: translateY(-6px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .accordion-item,
  .accordion-item::before,
  .accordion-item__trigger,
  .accordion-item__badge,
  .accordion-item__chevron,
  .accordion-item__panel {
    transition: none;
  }

  .accordion-item--open .accordion-item__body {
    animation: none;
  }
}
</style>
