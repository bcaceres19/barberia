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
  height: 100dvh;
  background-color: var(--color-surface-strong);
}

/* min-height: 0 es necesario para que este hijo flex pueda encogerse por
   debajo de la altura de su contenido: sin esto, un contenido largo crece
   la columna entera más allá de 100dvh y arrastra el header y el dock
   fuera de pantalla en vez de quedarse fijos mientras solo el contenido
   se desplaza (reporte en vivo, issue #189). */
.private-shell__content {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  /* Barra de scroll propia (latón sobre tinta), no la gris genérica del
     navegador — mismo acento que el resto del cascarón. */
  scrollbar-width: thin;
  scrollbar-color: rgb(184 149 90 / 45%) transparent;
}

.private-shell__content::-webkit-scrollbar {
  width: 10px;
}

.private-shell__content::-webkit-scrollbar-track {
  background: transparent;
}

.private-shell__content::-webkit-scrollbar-thumb {
  background-color: rgb(184 149 90 / 45%);
  border: 2px solid var(--color-surface-strong);
  border-radius: var(--radius-pill);
}

.private-shell__content::-webkit-scrollbar-thumb:hover {
  background-color: rgb(184 149 90 / 70%);
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
