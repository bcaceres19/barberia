<script setup lang="ts">
// Cabecera del cascarón privado (HU-012, CA-012-04): muestra siempre la
// barbería activa y permite cerrar sesión. El nombre viene únicamente de la
// fuente aprobada por DEC-060 (prop obligatoria, nunca un valor local ni un
// texto fijo). Recompuesta a la identidad NAVA en la Fase 2
// (estandar-diseno-visual.md §5.3): wordmark + barbería a la izquierda,
// "Nuevo turno" como acción primaria global (no un destino del dock,
// especificacion-frontend-nava.md §5.1) y cierre de sesión conservado tal
// cual (misma lógica y pruebas, sin convertirlo en un ítem del dock).
// Fase 2 del rediseño Tailored Grid (issue #187): la cabecera pasa a
// superficie tinta (orientación/navegación, estandar-diseno-visual.md §3)
// y "Nuevo turno" usa el acento latón reservado para foco destacado; la
// lógica de cierre de sesión vive en `useLogout` (compartida con el menú
// "Más" del dock en móvil, ver AppNav.vue).
import { BaseButton, NavaWordmark } from '@/shared/ui'
import { useLogout } from '../model/logout'

defineProps<{
  barbershopName: string
}>()

const { loggingOut, logout: onLogout } = useLogout()
</script>

<template>
  <header class="app-header">
    <div class="app-header__identity">
      <NavaWordmark variant="inverted" />
      <span class="app-header__divider" aria-hidden="true">·</span>
      <p class="app-header__barbershop" data-testid="barbershop-name">{{ barbershopName }}</p>
    </div>
    <div class="app-header__actions">
      <RouterLink :to="{ name: 'agenda-nuevo-turno' }" class="app-header__cta">
        Nuevo turno
      </RouterLink>
      <BaseButton
        variant="secondary"
        :disabled="loggingOut"
        :loading="loggingOut"
        @click="onLogout"
      >
        Cerrar sesión
      </BaseButton>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background-color: var(--color-surface-strong);
}

.app-header__identity {
  display: flex;
  min-width: 0;
  align-items: baseline;
  gap: var(--space-2);
}

.app-header__divider {
  color: var(--color-on-strong);
  opacity: 0.48;
}

.app-header__barbershop {
  margin: 0;
  overflow: hidden;
  font-size: var(--font-size-body);
  font-weight: 600;
  color: var(--color-on-strong);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-header__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2);
}

.app-header__cta {
  display: inline-flex;
  height: var(--control-height);
  align-items: center;
  padding: 0 var(--space-4);
  border-radius: var(--radius-sm);
  background-color: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  font-family: var(--font-sans);
  font-size: var(--font-size-body);
  font-weight: 500;
  text-decoration: none;
  transition: filter var(--motion-duration-fast) var(--motion-easing-standard);
}

@media (prefers-reduced-motion: reduce) {
  .app-header__cta {
    transition: none;
  }
}

.app-header__cta:hover {
  filter: brightness(94%);
}

.app-header__cta:active {
  filter: brightness(88%);
}

.app-header__cta:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus);
}
</style>
