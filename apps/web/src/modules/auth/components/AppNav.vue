<script setup lang="ts">
// Navegación principal del cascarón (HU-012). Estructura lista para B1 en
// adelante, pero sin inventar destinos que todavía no existen: hoy la única
// ruta privada real es /panel (fuera de alcance: "adelantar la edición de
// HU-020" ni ninguna otra capacidad de B1-B3). Cada entrada nueva se agrega
// aquí declarando solo su ruta y su texto, según el criterio de terminado
// de esta historia.
interface NavItem {
  to: { name: string }
  label: string
}

const items: NavItem[] = [{ to: { name: 'panel' }, label: 'Panel' }]
</script>

<template>
  <nav class="app-nav" aria-label="Navegación principal">
    <ul class="app-nav__list">
      <li v-for="item in items" :key="item.label" class="app-nav__item">
        <RouterLink :to="item.to" class="app-nav__link" active-class="app-nav__link--active">
          {{ item.label }}
        </RouterLink>
      </li>
    </ul>
  </nav>
</template>

<style scoped>
.app-nav__list {
  display: flex;
  gap: var(--space-2);
  margin: 0;
  padding: var(--space-2) var(--space-4);
  list-style: none;
  border-bottom: var(--border-width-normal) solid var(--color-border-subtle);
}

.app-nav__link {
  display: inline-flex;
  align-items: center;
  height: var(--control-height);
  padding: 0 var(--space-3);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  font-size: var(--font-size-body);
  font-weight: 500;
  text-decoration: none;
  transition: background-color var(--motion-duration-fast) var(--motion-easing-standard);
}

@media (prefers-reduced-motion: reduce) {
  .app-nav__link {
    transition: none;
  }
}

.app-nav__link:hover {
  background-color: var(--color-surface-muted);
}

.app-nav__link:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-surface),
    0 0 0 4px var(--color-focus);
}

.app-nav__link--active {
  color: var(--color-action-primary);
  background-color: var(--color-action-soft);
}
</style>
