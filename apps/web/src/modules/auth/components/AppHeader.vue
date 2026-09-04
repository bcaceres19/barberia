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
import { NavaWordmark } from '@/shared/ui'
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
      <p class="app-header__barbershop" data-testid="barbershop-name">{{ barbershopName }}</p>
    </div>
    <!-- El atlas de /panel no dibuja un botón con relleno en la barra, pero
         "no elimines... logout... existentes solo porque un PNG estático no
         muestre la acción; la función prevalece y se compone de la forma
         menos intrusiva posible" (autorización del propietario, 2026-09-03):
         texto discreto en vez del botón con filete anterior, siempre visible
         para cualquier persona con mouse, no solo al enfocar con teclado. -->
    <button type="button" class="app-header__logout" :disabled="loggingOut" @click="onLogout">
      Cerrar sesión
    </button>
  </header>
</template>

<style scoped>
.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 75px;
  padding: 20px 40px;
  background-color: var(--color-surface-strong);
  /* Filete inferior del atlas panel-agenda-eventos (issue #189):
     rgba(184,149,90,.45), el mismo latón de marca a baja opacidad. */
  border-bottom: 1px solid rgb(184 149 90 / 45%);
}

.app-header__identity {
  display: flex;
  min-width: 0;
  align-items: baseline;
  gap: 14px;
}

.app-header__identity :deep(.nava-wordmark) {
  font-size: 34px;
}

.app-header__barbershop {
  margin: 0;
  overflow: hidden;
  font-size: 16px;
  font-weight: 400;
  color: var(--color-on-strong-muted);
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Texto discreto, no un botón con filete: pesa menos que "Nuevo turno" de
   la ficha de la página, pero sigue siendo un control real y visible en
   todo momento — nunca solo al enfocar (issue #189, ver comentario del
   template). */
.app-header__logout {
  flex-shrink: 0;
  height: var(--control-height);
  padding: 0 var(--space-2);
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  color: var(--color-on-strong);
  opacity: 0.72;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  cursor: pointer;
  transition: opacity var(--motion-duration-fast) var(--motion-easing-standard);
}

@media (prefers-reduced-motion: reduce) {
  .app-header__logout {
    transition: none;
  }
}

.app-header__logout:hover {
  opacity: 1;
}

.app-header__logout:disabled {
  cursor: not-allowed;
  opacity: 0.48;
}

.app-header__logout:focus-visible {
  outline: none;
  opacity: 1;
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus);
}

@media (max-width: 1023px) {
  .app-header {
    min-height: 59px;
    padding: 16px 20px;
  }

  .app-header__barbershop {
    font-size: 13px;
  }
}
</style>
