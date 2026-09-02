<script setup lang="ts">
// Cascarón reutilizable de toda ruta privada (HU-012). Compone cabecera,
// navegación y área de contenido solo cuando el estado es `authenticated`;
// renderiza los estados globales de carga y conexión perdida en su lugar,
// nunca una pantalla en blanco (objetivo de la historia). No construye
// agenda ni ninguna capacidad de B1-B3: `<RouterView />` es el único punto
// de extensión para las pantallas que las historias siguientes agreguen
// como hijas de esta misma ruta.
import { BaseAlert, BaseButton } from '@/shared/ui'
import type { NavItem } from '@/shared/navigation/navItem'
import { retryBootstrap, sessionState } from '../model/sessionStore'
import AppHeader from '../components/AppHeader.vue'
import AppNav from '../components/AppNav.vue'

// extraNavItems llega como prop estática de ruta (auth.privateShellRoute,
// HU-020): `app/router/index.ts` combina las entradas de cada módulo con
// una pantalla privada, sin que este componente ni `auth` importen esos
// módulos.
defineProps<{ extraNavItems?: NavItem[] }>()

function onRetry() {
  void retryBootstrap()
}
</script>

<template>
  <div class="private-shell">
    <template v-if="sessionState.bootstrap.status === 'authenticated'">
      <AppHeader :barbershop-name="sessionState.bootstrap.barbershopName" />
      <main class="private-shell__content">
        <RouterView />
      </main>
      <AppNav :extra-items="extraNavItems" />
    </template>

    <div
      v-else-if="sessionState.bootstrap.status === 'connection-lost'"
      class="private-shell__state"
    >
      <BaseAlert variant="warning" title="No pudimos conectar" role="status">
        Revisa tu conexión e inténtalo de nuevo.
        <template #action>
          <BaseButton variant="secondary" @click="onRetry">Reintentar</BaseButton>
        </template>
      </BaseAlert>
    </div>

    <div v-else class="private-shell__state" role="status" aria-live="polite">
      <p class="private-shell__loading-text">Comprobando tu sesión…</p>
    </div>
  </div>
</template>

<style scoped>
.private-shell {
  display: flex;
  flex-direction: column;
  min-height: 100dvh;
}

.private-shell__content {
  flex: 1;
  /* El dock (AppNav) queda fijo al pie (64px) fuera del flujo; este padding
     evita que la última fila de contenido quede oculta debajo
     (estandar-diseno-visual.md §5.3, §8.3). */
  padding-bottom: calc(64px + env(safe-area-inset-bottom, 0px) + var(--space-4));
}

.private-shell__state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  min-height: 100dvh;
  padding: var(--space-6) var(--space-4);
  gap: var(--space-4);
}

.private-shell__loading-text {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: var(--font-size-body);
}
</style>
