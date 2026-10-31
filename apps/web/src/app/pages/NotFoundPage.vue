<script setup lang="ts">
// Destino de cualquier ruta desconocida (docs/03-desarrollo/especificacion-frontend-nava.md §7.13).
// Es la misma pantalla para `/nada` y para `/panel/nada`: no distingue rutas privadas
// inexistentes de públicas, así que no revela cuáles existen. La acción lleva a `/`, que
// redirige al acceso (o al panel si ya hay sesión), sin inventar otra navegación.
import { PageState } from '@/shared/ui'
</script>

<template>
  <main class="not-found">
    <h1 class="not-found__title">Página no encontrada</h1>
    <PageState variant="info" headline="No encontramos esta página">
      La dirección no existe o cambió de lugar. Vuelve al inicio para continuar.
      <template #action>
        <RouterLink class="not-found__action" to="/">Ir al inicio</RouterLink>
      </template>
    </PageState>
  </main>
</template>

<style scoped>
.not-found {
  display: flex;
  min-height: 100dvh;
  flex-direction: column;
  background-color: var(--color-surface-strong);
}

/* `PageState` solo aporta un <h2>; el <h1> es accesible pero no se ve. */
.not-found__title {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}

.not-found__action {
  display: inline-flex;
  min-height: var(--control-height);
  align-items: center;
  padding: 0 var(--space-4);
  border: 1px solid var(--color-brand-accent-surface);
  color: var(--color-brand-accent-surface);
  font-family: var(--font-family-base);
  font-weight: 600;
  text-decoration: none;
}

.not-found__action:hover {
  text-decoration: underline;
}

.not-found__action:focus-visible {
  outline: 2px solid var(--color-focus);
  outline-offset: 2px;
}
</style>
