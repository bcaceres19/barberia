<script setup lang="ts">
// Índice de secciones de Configuración (DEC-110): en escritorio una columna
// fija con un filete de latón que se desliza hasta la sección visible; en
// pantallas angostas, una fila de pestañas que se desplaza en horizontal. Es
// navegación por anclas (enlaces reales), no pestañas que oculten contenido:
// toda la información sigue en la página.
export interface SettingsIndexItem {
  id: string
  number: string
  label: string
}

defineProps<{ items: readonly SettingsIndexItem[]; active: string }>()
const emit = defineEmits<{ select: [id: string] }>()

function onClick(event: MouseEvent, id: string) {
  event.preventDefault()
  emit('select', id)
}
</script>

<template>
  <nav class="settings-index" aria-label="Secciones de configuración">
    <ol class="settings-index__list">
      <li v-for="item in items" :key="item.id" class="settings-index__item">
        <a
          :href="`#${item.id}`"
          class="settings-index__link"
          :class="{ 'settings-index__link--active': item.id === active }"
          :aria-current="item.id === active ? 'location' : undefined"
          @click="onClick($event, item.id)"
        >
          <span class="settings-index__number" aria-hidden="true">{{ item.number }}</span>
          <span class="settings-index__label">{{ item.label }}</span>
        </a>
      </li>
    </ol>
  </nav>
</template>

<style scoped>
.settings-index {
  position: sticky;
  top: 24px;
  align-self: start;
}

.settings-index__list {
  display: flex;
  flex-direction: column;
  margin: 0;
  padding: 0;
  list-style: none;
  border-left: var(--border-width-normal) solid var(--color-field-strong-border);
}

.settings-index__link {
  position: relative;
  display: flex;
  align-items: baseline;
  gap: 10px;
  min-height: 40px;
  padding: 9px 12px;
  color: var(--color-on-strong-muted);
  font-size: 13px;
  line-height: 20px;
  text-decoration: none;
  transition:
    color var(--motion-duration-base) var(--motion-easing-standard),
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    transform var(--motion-duration-base) var(--motion-easing-standard);
}

/* El filete de latón crece desde el centro al activarse la sección. */
.settings-index__link::before {
  content: '';
  position: absolute;
  top: 6px;
  bottom: 6px;
  left: -2px;
  width: 2px;
  background: var(--color-brand-accent-surface);
  transform: scaleY(0);
  transition: transform 260ms cubic-bezier(0.34, 1.56, 0.64, 1);
}

.settings-index__link:hover {
  color: var(--color-on-strong);
  background: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  transform: translateX(2px);
}

.settings-index__link:focus-visible {
  outline: var(--border-width-emphasis) solid var(--color-focus);
  outline-offset: -2px;
}

.settings-index__link--active {
  color: var(--color-on-strong);
  font-weight: 600;
}

.settings-index__link--active::before {
  transform: scaleY(1);
}

.settings-index__number {
  font-family: var(--font-display);
  font-size: 15px;
  color: var(--color-brand-accent-surface);
}

@media (max-width: 1099px) {
  .settings-index {
    position: sticky;
    top: 0;
    z-index: var(--layer-sticky);
    margin-inline: -16px;
    padding: 8px 16px;
    background: var(--color-surface-strong);
    border-bottom: var(--border-width-normal) solid var(--color-field-strong-border);
  }

  .settings-index__list {
    flex-direction: row;
    gap: 4px;
    overflow-x: auto;
    border-left: 0;
    scrollbar-width: none;
  }

  .settings-index__list::-webkit-scrollbar {
    display: none;
  }

  .settings-index__link {
    flex: 0 0 auto;
    min-height: 44px;
    padding: 10px 12px;
    white-space: nowrap;
  }

  .settings-index__link::before {
    inset: auto 12px 4px 12px;
    top: auto;
    left: 12px;
    right: 12px;
    width: auto;
    height: 2px;
    transform: scaleX(0);
  }

  .settings-index__link--active::before {
    transform: scaleX(1);
  }

  .settings-index__link:hover {
    transform: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .settings-index__link,
  .settings-index__link::before {
    transition: none;
  }
}
</style>
