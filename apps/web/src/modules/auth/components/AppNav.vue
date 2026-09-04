<script setup lang="ts">
// Dock de navegación primaria del cascarón (HU-012; recompuesto a la
// identidad NAVA en la Fase 2, especificacion-frontend-nava.md §5.2-§5.3,
// §6 `DesktopDock`/`MobileDock`). Fijo al pie en todos los anchos ("No hay
// sidebar lateral en la dirección NAVA"): icono + texto, 64px, con
// `env(safe-area-inset-bottom)`. "Agenda" es la única entrada propia de
// `auth`; cualquier módulo hermano con pantalla privada aporta la suya
// mediante `extraItems` (mismo mecanismo de composición app → modules →
// shared que ya usaba la navegación anterior). `RouterLink` ya marca
// `aria-current="page"` en la entrada activa sin código adicional.
//
// Fase 2 del rediseño Tailored Grid (issue #187, estandar-diseno-
// visual.md §6.5): el dock ya no listaba los seis destinos en línea plana
// (Agenda/Servicios/Barberos/Horarios/Configuración/Servicios por
// barbero), lo que trunca las etiquetas en anchos angostos y nunca incluyó
// "Bloqueos" (ruta real sin entrada de navegación). Ahora solo los
// destinos marcados `primary` viven en el dock; el resto se expone dentro
// de "Más" (BaseDialog, reutiliza su foco atrapado/Escape/restauración),
// junto con "Cerrar sesión" para que ese cierre también sea alcanzable sin
// depender del ancho de AppHeader.
import { computed, ref } from 'vue'
import type { NavItem } from '@/shared/navigation/navItem'
import { BaseDialog, RecordRow } from '@/shared/ui'
import { useLogout } from '../model/logout'

const props = defineProps<{ extraItems?: NavItem[] }>()

const baseItems: NavItem[] = [{ to: { name: 'panel' }, label: 'Agenda', primary: true }]
const items = computed<NavItem[]>(() => [...baseItems, ...(props.extraItems ?? [])])
const primaryItems = computed(() => items.value.filter((item) => item.primary))
const secondaryItems = computed(() => items.value.filter((item) => !item.primary))

// Un icono por destino P0 conocido; un destino futuro sin entrada aquí cae
// al icono de Agenda en vez de romper (mismo criterio defensivo que
// `getWeekdayLabel`/`getBlockTypeLabel` de `schedules`). Decorativo:
// aria-hidden porque el nombre accesible del enlace ya lo da el texto.
const ICONS: Record<string, string> = {
  panel: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="17" rx="2"></rect><line x1="3" y1="9" x2="21" y2="9"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="16" y1="2" x2="16" y2="6"></line></svg>`,
  'catalog-servicios': `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4h10l6 6v10a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V5a1 1 0 0 1 1-1z"></path><path d="M14 4v6h6"></path><line x1="8" y1="13" x2="15" y2="13"></line><line x1="8" y1="17" x2="12" y2="17"></line></svg>`,
  'staff-barberos': `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="8" r="4"></circle><path d="M4 21c0-4 3.6-7 8-7s8 3 8 7"></path></svg>`,
  'schedules-horarios': `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9"></circle><path d="M12 7v5l3.5 2"></path></svg>`,
  'schedules-bloqueos': `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="10" width="16" height="10" rx="1.5"></rect><path d="M8 10V7a4 4 0 0 1 8 0v3"></path></svg>`,
  'configuracion-barberia': `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M4 6h11M19 6h1M4 12h1M9 12h11M4 18h6M14 18h6"></path><circle cx="17" cy="6" r="2"></circle><circle cx="7" cy="12" r="2"></circle><circle cx="12" cy="18" r="2"></circle></svg>`,
  'barber-services': `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><circle cx="9" cy="7" r="3"></circle><path d="M2 21c0-3.5 3-6 7-6s7 2.5 7 6"></path><path d="M16 4a3 3 0 0 1 0 6"></path><path d="M22 21c0-2.7-1.7-4.8-4-5.6"></path></svg>`,
}
const MORE_ICON = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><circle cx="5" cy="12" r="1.5"></circle><circle cx="12" cy="12" r="1.5"></circle><circle cx="19" cy="12" r="1.5"></circle></svg>`

function iconFor(routeName: string): string {
  return ICONS[routeName] ?? ICONS.panel
}

const moreOpen = ref(false)
const { loggingOut, logout } = useLogout()

function closeMore() {
  moreOpen.value = false
}

async function onLogoutFromMore() {
  await logout()
  closeMore()
}
</script>

<template>
  <nav class="app-nav" aria-label="Navegación principal">
    <ul class="app-nav__list app-nav__list--desktop">
      <li v-for="item in items" :key="item.label" class="app-nav__item">
        <RouterLink :to="item.to" class="app-nav__link" active-class="app-nav__link--active">
          <span class="app-nav__icon" aria-hidden="true" v-html="iconFor(item.to.name)" />
          <span class="app-nav__label">{{ item.label }}</span>
        </RouterLink>
      </li>
    </ul>
    <ul class="app-nav__list app-nav__list--mobile">
      <li v-for="item in primaryItems" :key="item.label" class="app-nav__item">
        <RouterLink :to="item.to" class="app-nav__link" active-class="app-nav__link--active">
          <span class="app-nav__icon" aria-hidden="true" v-html="iconFor(item.to.name)" />
          <span class="app-nav__label">{{ item.label }}</span>
        </RouterLink>
      </li>
      <li v-if="secondaryItems.length" class="app-nav__item">
        <button
          type="button"
          class="app-nav__link app-nav__link--button"
          aria-haspopup="dialog"
          :aria-expanded="moreOpen"
          @click="moreOpen = true"
        >
          <span class="app-nav__icon" aria-hidden="true" v-html="MORE_ICON" />
          <span class="app-nav__label">Más</span>
        </button>
      </li>
    </ul>

    <BaseDialog v-model="moreOpen" title="Más" size="sm">
      <ul class="app-nav__more-list" aria-label="Más destinos y cierre de sesión">
        <RecordRow v-for="item in secondaryItems" :key="item.label">
          <RouterLink
            :to="item.to"
            class="app-nav__more-link"
            active-class="app-nav__more-link--active"
            @click="closeMore"
          >
            <span class="app-nav__icon" aria-hidden="true" v-html="iconFor(item.to.name)" />
            {{ item.label }}
          </RouterLink>
        </RecordRow>
        <RecordRow>
          <button
            type="button"
            class="app-nav__more-link app-nav__more-logout"
            :disabled="loggingOut"
            :aria-busy="loggingOut"
            @click="onLogoutFromMore"
          >
            Cerrar sesión
          </button>
        </RecordRow>
      </ul>
    </BaseDialog>
  </nav>
</template>

<style scoped>
.app-nav {
  margin-top: auto;
  /* Superficie tinta, igual que AppHeader: la navegación es orientación
     operativa (estandar-diseno-visual.md §3), no un formulario. */
  background-color: var(--color-surface-strong);
  box-shadow: var(--shadow-raised);
  padding-bottom: env(safe-area-inset-bottom, 0px);
}

.app-nav__list {
  display: flex;
  align-items: stretch;
  justify-content: space-around;
  height: 64px;
  margin: 0;
  padding: 0 var(--space-1);
  list-style: none;
}

.app-nav__list--desktop {
  display: none;
}

.app-nav__item {
  flex: 1;
  min-width: 0;
}

.app-nav__link {
  display: flex;
  width: 100%;
  height: 100%;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-1);
  padding: var(--space-1);
  background: transparent;
  border: none;
  border-bottom: var(--border-width-emphasis) solid transparent;
  color: var(--color-on-strong);
  opacity: 0.64;
  font-family: var(--font-family-base);
  font-size: var(--font-size-caption);
  font-weight: 500;
  text-align: center;
  text-decoration: none;
  cursor: pointer;
  transition:
    color var(--motion-duration-fast) var(--motion-easing-standard),
    opacity var(--motion-duration-fast) var(--motion-easing-standard),
    border-color var(--motion-duration-fast) var(--motion-easing-standard);
}

.app-nav__link--button {
  font: inherit;
}

@media (prefers-reduced-motion: reduce) {
  .app-nav__link {
    transition: none;
  }
}

.app-nav__icon :deep(svg) {
  width: 20px;
  height: 20px;
}

.app-nav__label {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-nav__link:hover {
  opacity: 1;
}

.app-nav__link:focus-visible {
  outline: none;
  opacity: 1;
  box-shadow: inset 0 0 0 var(--border-width-emphasis) var(--color-focus);
}

.app-nav__link--active,
.app-nav__link[aria-expanded='true'] {
  border-bottom-color: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-surface);
  opacity: 1;
}

@media (min-width: 1024px) {
  .app-nav__list--mobile {
    display: none;
  }

  .app-nav__list--desktop {
    display: flex;
    justify-content: flex-start;
    height: 64px;
    padding: 0 40px;
  }

  .app-nav__item {
    flex: 1;
  }

  .app-nav__link {
    flex-direction: row;
    gap: 10px;
    padding: 0 18px;
    font-size: 14px;
  }
}

.app-nav__more-list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.app-nav__more-link {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  width: 100%;
  min-height: var(--control-height);
  padding: var(--space-2) 0;
  background: transparent;
  border: none;
  color: var(--color-text-primary);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  text-align: left;
  text-decoration: none;
  cursor: pointer;
}

.app-nav__more-link:hover {
  color: var(--color-action-primary);
}

.app-nav__more-link:focus-visible {
  outline: none;
  box-shadow: 0 0 0 var(--border-width-emphasis) var(--color-focus);
}

.app-nav__more-link--active {
  color: var(--color-action-primary);
  font-weight: 600;
}

.app-nav__more-logout {
  color: var(--color-danger-text);
}

.app-nav__more-logout:disabled {
  opacity: 0.64;
  cursor: not-allowed;
}
</style>
