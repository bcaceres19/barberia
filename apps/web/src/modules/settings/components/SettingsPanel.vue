<script setup lang="ts">
// Sección numerada de Configuración (DEC-110): el mismo lenguaje reglado de
// Agenda, Servicios y Barberos —superficie tinta, filete de latón, versalitas
// espaciadas, rombo de estado— con un número editorial en serif como ancla. El
// encabezado deja claro DÓNDE vale lo que se cambia (este dispositivo o toda la
// barbería) y si aplica al instante o necesita guardarse.
defineProps<{
  id: string
  number: string
  title: string
  description?: string
  /** `device`: vale en este aparato y se aplica al instante. `shop`: vale para toda la barbería y se guarda. */
  scope?: 'device' | 'shop'
  scopeLabel?: string
}>()
</script>

<template>
  <section :id="id" class="settings-panel" :aria-labelledby="`${id}-title`">
    <header class="settings-panel__header">
      <span class="settings-panel__number" aria-hidden="true">{{ number }}</span>
      <div class="settings-panel__heading">
        <h2 :id="`${id}-title`" class="settings-panel__title">{{ title }}</h2>
        <p v-if="description" class="settings-panel__description">{{ description }}</p>
      </div>
      <span
        v-if="scope && scopeLabel"
        class="settings-panel__scope"
        :class="`settings-panel__scope--${scope}`"
      >
        {{ scopeLabel }}
      </span>
    </header>
    <div class="settings-panel__body">
      <slot />
    </div>
  </section>
</template>

<style scoped>
.settings-panel {
  position: relative;
  scroll-margin-top: 72px;
  background: color-mix(in srgb, var(--color-on-strong) 3%, transparent);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-top: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  border-radius: 3px;
  animation: settings-panel-enter 420ms var(--motion-easing-standard) both;
  animation-delay: calc(var(--panel-index, 0) * 70ms);
}

/* La animación de entrada deja a cada panel en su propio contexto de
   apilamiento: sin esto, el desplegable de la zona horaria quedaría tapado por
   el panel siguiente. El panel con el foco sube por encima de sus vecinos. */
.settings-panel:focus-within {
  z-index: 2;
}

.settings-panel__header {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding: 20px 24px 16px;
  border-bottom: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-brand-accent-surface) 35%, transparent);
}

.settings-panel__number {
  flex: 0 0 auto;
  font-family: var(--font-display);
  font-size: 34px;
  line-height: 1;
  color: var(--color-brand-accent-surface);
}

.settings-panel__heading {
  min-width: 0;
  flex: 1;
}

.settings-panel__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: 24px;
  font-weight: 400;
  line-height: 28px;
  color: var(--color-on-strong);
}

.settings-panel__description {
  margin: 4px 0 0;
  max-width: 56ch;
  font-size: 13px;
  line-height: 19px;
  color: var(--color-on-strong-muted);
}

/* Palabra de alcance en versalitas con rombo, como los estados de la agenda. */
.settings-panel__scope {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 6px;
  margin-top: 6px;
  color: var(--color-on-strong-muted);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  line-height: 16px;
  text-transform: uppercase;
  white-space: nowrap;
}

.settings-panel__scope::before {
  content: '';
  width: 5px;
  height: 5px;
  background: currentColor;
  transform: rotate(45deg);
  animation: settings-scope-blink 7s ease-in-out infinite;
}

.settings-panel__scope--device {
  color: var(--color-success-on-strong);
}

.settings-panel__scope--shop {
  color: var(--color-brand-accent-surface);
}

.settings-panel__body {
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 22px 24px 26px;
}

@keyframes settings-panel-enter {
  from {
    opacity: 0;
    transform: translateY(14px);
  }

  to {
    opacity: 1;
    transform: none;
  }
}

@keyframes settings-scope-blink {
  0%,
  62%,
  100% {
    opacity: 1;
    transform: rotate(45deg) scale(1);
  }

  80% {
    opacity: 0.55;
    transform: rotate(45deg) scale(0.85);
  }
}

@media (max-width: 640px) {
  .settings-panel__header {
    flex-wrap: wrap;
    gap: 8px 12px;
    padding: 16px 16px 14px;
  }

  .settings-panel__number {
    font-size: 28px;
  }

  .settings-panel__title {
    font-size: 21px;
    line-height: 25px;
  }

  .settings-panel__scope {
    flex-basis: 100%;
    margin-top: 0;
    white-space: normal;
  }

  .settings-panel__body {
    gap: 20px;
    padding: 18px 16px 20px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .settings-panel,
  .settings-panel__scope::before {
    animation: none;
  }
}
</style>
