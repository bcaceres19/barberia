<script setup lang="ts">
// Cascarón reutilizable de toda ruta privada (HU-012). Compone cabecera,
// navegación y área de contenido solo cuando el estado es `authenticated`;
// renderiza los estados globales de carga y conexión perdida en su lugar,
// nunca una pantalla en blanco (objetivo de la historia). No construye
// agenda ni ninguna capacidad de B1-B3: `<RouterView />` es el único punto
// de extensión para las pantallas que las historias siguientes agreguen
// como hijas de esta misma ruta.
import { onBeforeUnmount } from 'vue'
import { attachWorkspaceTheme } from '@/shared/model'
import { BaseAlert, BaseButton, ToastRegion } from '@/shared/ui'
import type { NavItem } from '@/shared/navigation/navItem'
import { retryBootstrap, sessionState } from '../model/sessionStore'
import AppHeader from '../components/AppHeader.vue'
import AppNav from '../components/AppNav.vue'

// extraNavItems llega como prop estática de ruta (auth.privateShellRoute,
// HU-020): `app/router/index.ts` combina las entradas de cada módulo con
// una pantalla privada, sin que este componente ni `auth` importen esos
// módulos.
defineProps<{ extraNavItems?: NavItem[] }>()

// Modo, tamaño de texto, animaciones y acento de la barbería (DEC-110) se
// pintan en <html> solo mientras el cascarón está montado; al salir se retiran,
// así el acceso y la reserva pública nunca heredan el tema del panel.
const detachWorkspaceTheme = attachWorkspaceTheme()
onBeforeUnmount(detachWorkspaceTheme)

function onRetry() {
  void retryBootstrap()
}
</script>

<template>
  <div class="private-shell">
    <template v-if="sessionState.bootstrap.status === 'authenticated'">
      <AppHeader :barbershop-name="sessionState.bootstrap.barbershopName" />
      <!-- Escenario posicionado entre cabecera y navegación: ancla la región
           de avisos (DEC-095) sin depender de la altura de ninguna de las dos
           ni desplazarse con el contenido. -->
      <div class="private-shell__stage">
        <!-- Fondo vivo del panel: dos luces suaves que derivan despacio y una
             cuadrícula de patronaje que avanza una casilla por ciclo. Es una
             capa decorativa detrás del contenido (sin eventos ni lectura de
             pantalla); las pantallas son transparentes y la tinta la pone el
             escenario. -->
        <div class="shell-ambient" aria-hidden="true">
          <i class="shell-ambient__glow shell-ambient__glow--brass" />
          <i class="shell-ambient__glow shell-ambient__glow--sage" />
          <i class="shell-ambient__grid" />
        </div>
        <main class="private-shell__content">
          <RouterView />
        </main>
        <ToastRegion placement="host" />
      </div>
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
  height: var(--viewport-height);
  background-color: var(--color-chrome-surface);
}

/* min-height: 0 es necesario para que este hijo flex pueda encogerse por
   debajo de la altura de su contenido: sin esto, un contenido largo crece
   la columna entera más allá de 100dvh y arrastra el header y el dock
   fuera de pantalla en vez de quedarse fijos mientras solo el contenido
   se desplaza (reporte en vivo, issue #189). */
.private-shell__stage {
  position: relative;
  background-color: var(--color-surface-strong);
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.private-shell__content {
  position: relative;
  z-index: 1;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  /* Barra de scroll propia (latón sobre tinta), no la gris genérica del
     navegador — mismo acento que el resto del cascarón. */
  scrollbar-width: thin;
  scrollbar-color: color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent)
    transparent;
}

.private-shell__content::-webkit-scrollbar {
  width: 10px;
}

.private-shell__content::-webkit-scrollbar-track {
  background: transparent;
}

.private-shell__content::-webkit-scrollbar-thumb {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent);
  border: 2px solid var(--color-surface-strong);
  border-radius: var(--radius-pill);
}

.private-shell__content::-webkit-scrollbar-thumb:hover {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 70%, transparent);
}

.shell-ambient {
  position: absolute;
  inset: 0;
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
}

/* Luces: manchas radiales muy tenues de latón y salvia. Solo se trasladan,
   en ciclos de casi un minuto, y lo hacen POR PASOS (`steps`): el fondo se
   mueve menos de un píxel por paso, imperceptible, pero se recompone ~5
   veces por segundo en vez de 60. Medido en Firefox: de ~10-20 % de un
   núcleo a ~3 % con la página quieta. */
.shell-ambient__glow {
  position: absolute;
  width: max(70vmax, 520px);
  height: max(70vmax, 520px);
  border-radius: 50%;
  will-change: transform;
}

.shell-ambient__glow--brass {
  top: -34vmax;
  left: -22vmax;
  background: radial-gradient(
    closest-side,
    color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent),
    transparent
  );
  animation: shell-drift-a 52s steps(260, end) infinite alternate;
}

.shell-ambient__glow--sage {
  right: -26vmax;
  bottom: -38vmax;
  background: radial-gradient(
    closest-side,
    color-mix(in srgb, var(--color-brand-sage) 14%, transparent),
    transparent
  );
  animation: shell-drift-b 64s steps(320, end) infinite alternate;
}

/* Cuadrícula de patronaje: líneas casi invisibles que se desvanecen hacia los
   bordes y se desplazan una casilla por ciclo (el bucle es continuo porque
   el desplazamiento iguala el tamaño de la celda). */
.shell-ambient__grid {
  position: absolute;
  inset: -56px;
  background-image:
    linear-gradient(
      to right,
      color-mix(in srgb, var(--color-on-strong) 4%, transparent) 1px,
      transparent 1px
    ),
    linear-gradient(
      to bottom,
      color-mix(in srgb, var(--color-on-strong) 4%, transparent) 1px,
      transparent 1px
    );
  background-size: 56px 56px;
  mask-image: radial-gradient(ellipse 80% 70% at 50% 40%, #000 20%, transparent 100%);
  animation: shell-grid 70s steps(56, end) infinite;
  will-change: transform;
}

@keyframes shell-drift-a {
  from {
    transform: translate3d(0, 0, 0) scale(1);
  }

  to {
    transform: translate3d(14vmax, 10vmax, 0) scale(1.15);
  }
}

@keyframes shell-drift-b {
  from {
    transform: translate3d(0, 0, 0) scale(1.1);
  }

  to {
    transform: translate3d(-12vmax, -8vmax, 0) scale(0.95);
  }
}

@keyframes shell-grid {
  from {
    transform: translate3d(0, 0, 0);
  }

  to {
    transform: translate3d(56px, 56px, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .shell-ambient__glow,
  .shell-ambient__grid {
    animation: none;
  }
}

:root[data-motion='reduced'] .shell-ambient__glow,
:root[data-motion='reduced'] .shell-ambient__grid {
  animation: none;
}

.private-shell__state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  min-height: var(--viewport-height);
  padding: var(--space-6) var(--space-4);
  gap: var(--space-4);
}

.private-shell__loading-text {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: var(--font-size-body);
}
</style>
