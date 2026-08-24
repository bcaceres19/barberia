<script setup lang="ts">
// Navegación principal del cascarón (HU-012). "Panel" es la única entrada
// que `auth` conoce de por sí; cualquier módulo hermano con una pantalla
// privada (HU-020 en adelante) aporta la suya mediante `extraItems`, que
// `app/router/index.ts` compone y pasa a través de `PrivateShell` (app →
// modules → shared): `auth` nunca importa el módulo que la declaró.
import { computed } from 'vue'
import type { NavItem } from '@/shared/navigation/navItem'

const props = defineProps<{ extraItems?: NavItem[] }>()

const baseItems: NavItem[] = [{ to: { name: 'panel' }, label: 'Panel' }]
const items = computed<NavItem[]>(() => [...baseItems, ...(props.extraItems ?? [])])
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
  /* HU-022 agregó una cuarta entrada ("Servicios"): a 320/360 px, cuatro
     enlaces ya no caben en una sola fila sin desbordar el documento
     (docs/03-desarrollo/estandar-diseno-visual.md §16, CA-022-08). En vez
     de envolver a varias líneas -que cambia el ritmo vertical del
     cascarón entre barberías con más o menos módulos- la barra se
     desplaza horizontalmente dentro de sí misma, un patrón de pestañas
     estándar que escala a un quinto o sexto módulo futuro sin volver a
     tocar este componente.
   */
  overflow-x: auto;
  overscroll-behavior-x: contain;
}

.app-nav__item {
  flex-shrink: 0;
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
