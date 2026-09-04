<script setup lang="ts">
/**
 * AgendaSkeleton - Espera con la geometría del contenido por llegar (issue
 * #189, atlas panel-agenda-eventos evento 05): se usa cuando barbero y fecha
 * ya están resueltos, para que la pantalla no salte al llegar la agenda real
 * (a diferencia del contexto inicial, evento 02, que no tiene nada que
 * anticipar y usa PageState + BaseSpinner en su lugar).
 *
 * Local al módulo `agenda` porque su geometría (carril + lista) es la de
 * esta pantalla — no se promueve a shared/ui.
 *
 * El texto de espera nunca desaparece: sigue anunciándose vía
 * role="status"/aria-live="polite" aunque la representación visual sea
 * geometría, no un spinner. `prefers-reduced-motion: reduce` deja los
 * bloques en opacidad fija, sin pulso.
 */
interface Props {
  /** Texto exacto ya usado por el código, p. ej. "Cargando la agenda de
   * {barbero}…" — este componente no lo redacta. */
  label: string
  /** Cantidad de filas de placeholder a dibujar. */
  entryCount?: number
}

withDefaults(defineProps<Props>(), {
  entryCount: 4,
})
</script>

<template>
  <div class="agenda-skeleton" role="status" aria-live="polite">
    <p class="agenda-skeleton__label">{{ label }}</p>

    <div class="agenda-skeleton__timeline" aria-hidden="true">
      <span
        v-for="n in entryCount"
        :key="n"
        class="agenda-skeleton__slip"
        :style="{ left: `${(n - 1) * (90 / entryCount) + 2}%`, width: `${60 / entryCount}%` }"
      />
    </div>

    <ul class="agenda-skeleton__list" aria-hidden="true">
      <li v-for="n in entryCount" :key="n" class="agenda-skeleton__item">
        <span class="agenda-skeleton__bar agenda-skeleton__bar--name" />
        <span class="agenda-skeleton__bar agenda-skeleton__bar--badge" />
      </li>
    </ul>
  </div>
</template>

<style scoped>
.agenda-skeleton {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.agenda-skeleton__label {
  margin: 0;
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-accent-brass);
}

.agenda-skeleton__timeline {
  position: relative;
  display: none;
  height: 64px;
}

@media (min-width: 1024px) {
  .agenda-skeleton__timeline {
    display: block;
  }
}

.agenda-skeleton__slip {
  position: absolute;
  top: 0;
  bottom: 0;
  background-color: rgb(244 240 231 / 12%);
  border-left: var(--border-width-emphasis) solid var(--color-accent-brass);
  border-radius: var(--radius-sm);
  animation: agenda-skeleton-pulse 1400ms ease-in-out infinite;
}

.agenda-skeleton__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding: 0;
  margin: 0;
  list-style: none;
}

.agenda-skeleton__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 64px;
  padding: var(--space-4);
  background-color: rgb(244 240 231 / 6%);
  border-radius: var(--radius-md);
}

.agenda-skeleton__bar {
  display: block;
  height: 12px;
  background-color: rgb(244 240 231 / 16%);
  border-radius: var(--radius-sm);
  animation: agenda-skeleton-pulse 1400ms ease-in-out infinite;
}

.agenda-skeleton__bar--name {
  width: 40%;
}

.agenda-skeleton__bar--badge {
  width: 72px;
}

@keyframes agenda-skeleton-pulse {
  0%,
  100% {
    opacity: 0.6;
  }
  50% {
    opacity: 1;
  }
}

@media (prefers-reduced-motion: reduce) {
  .agenda-skeleton__slip,
  .agenda-skeleton__bar {
    animation: none;
    opacity: 0.8;
  }
}
</style>
