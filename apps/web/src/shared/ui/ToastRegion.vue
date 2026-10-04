<script setup lang="ts">
/**
 * ToastRegion - Región única que muestra la cola de avisos emergentes
 * (DEC-095). Se monta una sola vez por pantalla:
 * - `viewport`: anclada a la ventana (acceso, recuperación, reserva pública).
 * - `host`: anclada a su contenedor posicionado, que en el shell privado es
 *   el espacio entre la cabecera y la navegación; así nunca tapa ninguna de
 *   las dos ni depende de su altura.
 * La región es persistente para que los lectores de pantalla anuncien cada
 * aviso al insertarse; solo Escape descarta el aviso con foco, sin mover el
 * foco de la tarea en curso.
 */
import {
  dismissAllToasts,
  dismissToast,
  setToastHeld,
  toastState,
  toggleToast,
} from '@/shared/model/toastStore'
import BaseToast from './BaseToast.vue'

withDefaults(defineProps<{ placement?: 'viewport' | 'host' }>(), { placement: 'viewport' })

function onAction(id: number, run: (() => void) | undefined) {
  // Se descarta primero para que un aviso emitido por la acción no compita
  // con el que la originó por el cupo máximo de la cola.
  dismissToast(id)
  run?.()
}
</script>

<template>
  <div :class="['toast-region', `toast-region--${placement}`]" role="region" aria-label="Avisos">
    <div
      v-for="toast in toastState.items"
      :key="toast.id"
      class="toast-region__item"
      @keydown.esc="dismissToast(toast.id)"
    >
      <BaseToast
        :variant="toast.variant"
        :title="toast.title"
        :detail="toast.detail"
        :reference="toast.reference"
        :action-label="toast.action?.label"
        :action-icon="toast.action?.icon"
        :expanded="toast.expanded"
        :held="toast.expanded || toast.hovered || toast.focused"
        :duration-ms="toast.durationMs"
        @toggle="toggleToast(toast.id)"
        @dismiss="dismissToast(toast.id)"
        @action="onAction(toast.id, toast.action?.run)"
        @hover="setToastHeld(toast.id, 'hovered', $event)"
        @focus="setToastHeld(toast.id, 'focused', $event)"
      />
    </div>

    <button
      v-if="toastState.items.length > 1"
      type="button"
      class="toast-region__clear"
      @click="dismissAllToasts()"
    >
      <svg
        width="12"
        height="12"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2.4"
        stroke-linecap="round"
        aria-hidden="true"
      >
        <path d="M6 6l12 12M18 6L6 18" />
      </svg>
      <span>Descartar todos</span>
      <span class="toast-region__count">{{ toastState.items.length }}</span>
    </button>
  </div>
</template>

<style scoped>
/* Móvil: ancho completo con margen de 12px, bajo la cabecera. Escritorio:
   columna de 380px a la derecha. La región no captura el puntero: solo los
   avisos y el botón de descarte lo hacen, para no bloquear la pantalla. */
.toast-region {
  position: fixed;
  top: var(--space-4);
  right: 12px;
  left: 12px;
  z-index: var(--layer-toast);
  display: flex;
  flex-direction: column;
  gap: 10px;
  pointer-events: none;
}

.toast-region--host {
  position: absolute;
  top: var(--space-2);
}

.toast-region__item {
  pointer-events: auto;
}

@media (min-width: 1024px) {
  .toast-region {
    right: var(--space-10);
    left: auto;
    width: 380px;
  }

  .toast-region--host {
    top: var(--space-4);
  }
}

.toast-region__clear {
  position: relative;
  display: inline-flex;
  align-items: center;
  align-self: flex-end;
  gap: var(--space-2);
  height: 36px;
  padding: 0 var(--space-3);
  background-color: var(--color-surface-strong);
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent);
  border-radius: 2px;
  box-shadow: 0 8px 20px -8px rgb(16 27 43 / 50%);
  color: var(--color-on-strong);
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  cursor: pointer;
  pointer-events: auto;
  transition:
    background-color var(--motion-duration-fast) var(--motion-easing-standard),
    border-color var(--motion-duration-fast) var(--motion-easing-standard);
}

/* Botón de 36px con zona activa de 44px (estandar-diseno-visual.md §8). */
.toast-region__clear::after {
  position: absolute;
  top: -4px;
  right: 0;
  bottom: -4px;
  left: 0;
  content: '';
}

.toast-region__clear:hover {
  background-color: var(--color-action-primary-hover);
  border-color: color-mix(in srgb, var(--color-on-strong) 32%, transparent);
}

.toast-region__clear svg {
  flex: none;
  color: var(--color-brand-accent-surface);
}

.toast-region__clear:focus-visible {
  outline: 2px solid var(--color-focus);
  outline-offset: 2px;
}

.toast-region__count {
  padding-left: var(--space-2);
  border-left: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-on-strong) 20%, transparent);
  color: var(--color-brand-accent-surface);
}

@media (prefers-reduced-motion: reduce) {
  .toast-region__clear {
    transition: none;
  }
}
</style>
