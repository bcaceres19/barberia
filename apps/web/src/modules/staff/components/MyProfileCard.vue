<script setup lang="ts">
// Ficha del perfil de panel `solo` (DEC-115): una sola persona no es una fila
// sin cabecera de una tabla de equipo, es su propia tarjeta. StaffPage.vue
// sigue siendo dueño de los datos, la carga y los diálogos (alta, edición);
// este componente es puramente de presentación y emite `edit` para abrir el
// diálogo de edición que StaffPage ya tiene. Mismo lenguaje visual que el
// detalle de barbero y los diálogos de Servicios: tinta hundida, filete de
// latón, retrato con marco de esquinas y rombo de estado.
import { BarberAvatar, BaseButton } from '@/shared/ui'

defineProps<{
  fullName: string
  /** Descripción pública (DEC-127); null mientras no tiene. */
  description: string | null
  photoUrl: string | null
  hasPhoto: boolean
  sinceLabel: string
  roleLabel: string
}>()

defineEmits<{ edit: [] }>()
</script>

<template>
  <section class="my-profile-card" aria-labelledby="my-profile-card-name">
    <div class="my-profile-card__hero">
      <span class="my-profile-card__frame">
        <BarberAvatar size="hero" :full-name="fullName" :photo-url="photoUrl" />
      </span>
      <p class="my-profile-card__role">{{ roleLabel }}</p>
      <h2 id="my-profile-card-name" class="my-profile-card__name">{{ fullName }}</h2>
      <p v-if="description" class="my-profile-card__description">{{ description }}</p>
    </div>

    <dl class="my-profile-card__facts">
      <div>
        <dt>En NAVA desde</dt>
        <dd>{{ sinceLabel }}</dd>
      </div>
      <div>
        <dt>Foto</dt>
        <dd
          class="my-profile-card__photo-state"
          :class="{ 'my-profile-card__photo-state--has': hasPhoto }"
        >
          {{ hasPhoto ? 'Con foto' : 'Sin foto' }}
        </dd>
      </div>
    </dl>

    <slot name="reviews" />

    <div class="my-profile-card__actions">
      <slot name="actions" />
      <BaseButton
        variant="secondary"
        class="my-profile-card__edit"
        :aria-label="`Editar ${fullName}`"
        @click="$emit('edit')"
      >
        <svg
          class="my-profile-card__edit-icon"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
          aria-hidden="true"
        >
          <path d="M4 20h4L19 9a2.8 2.8 0 0 0-4-4L4 16v4Z" />
          <path d="m13.5 6.5 4 4" />
        </svg>
        Editar mi perfil
      </BaseButton>
    </div>
  </section>
</template>

<style scoped>
/* Tarjeta centrada en vez de una fila de tabla sin cabecera: una persona sola
   es su propia pantalla, no el resto de un equipo que ya no está. */
.my-profile-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-5);
  width: min(100%, 460px);
  margin-inline: auto;
  padding: 40px 32px 32px;
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-top: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  border-radius: 4px;
  animation: my-profile-card-enter 380ms var(--motion-easing-standard) both;
}

@keyframes my-profile-card-enter {
  from {
    opacity: 0;
    transform: translateY(10px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.my-profile-card__hero {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
}

.my-profile-card__frame {
  position: relative;
  display: inline-flex;
  padding: 12px;
  margin-bottom: var(--space-1);
  background-color: var(--color-surface-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: 2px;
}

.my-profile-card__frame::before,
.my-profile-card__frame::after {
  content: '';
  position: absolute;
  width: 18px;
  height: 18px;
  border: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
}

.my-profile-card__frame::before {
  top: -5px;
  left: -5px;
  border-right: 0;
  border-bottom: 0;
}

.my-profile-card__frame::after {
  right: -5px;
  bottom: -5px;
  border-top: 0;
  border-left: 0;
}

.my-profile-card__role {
  margin: 0;
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.my-profile-card__description {
  max-width: 36ch;
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-body);
  line-height: var(--font-size-body-line);
  color: var(--color-on-strong-muted);
  text-align: center;
  overflow-wrap: anywhere;
  white-space: pre-line;
}

.my-profile-card__name {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h2);
  line-height: 1.15;
  font-weight: var(--font-weight-h2);
  color: var(--color-on-strong);
  text-align: center;
  overflow-wrap: anywhere;
}

.my-profile-card__facts {
  display: grid;
  grid-auto-flow: column;
  grid-auto-columns: minmax(0, 1fr);
  gap: var(--space-3);
  width: 100%;
  margin: 0;
  padding-top: var(--space-4);
  border-top: var(--border-width-normal) solid var(--color-field-strong-border);
}

.my-profile-card__facts > div {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-1);
  text-align: center;
}

.my-profile-card__facts dt {
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.my-profile-card__facts dd {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h3);
  line-height: 1.2;
  color: var(--color-on-strong);
  font-variant-numeric: tabular-nums;
}

/* Mismo rombo con rótulo que el estado de la foto en el equipo: sin caja,
   versalitas de latón. */
.my-profile-card__photo-state {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: var(--font-size-body);
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--color-on-strong-muted);
}

.my-profile-card__photo-state::before {
  content: '';
  flex-shrink: 0;
  width: 6px;
  height: 6px;
  transform: rotate(45deg);
  background-color: currentColor;
}

.my-profile-card__photo-state--has {
  color: var(--color-success-on-strong);
}

.my-profile-card__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  width: 100%;
}

/* «Editar mi perfil» y «Bloquear» (que llega por slot desde StaffWorkspacePage)
   comparten forma: BaseButton secundario, fantasma con filete inferior de color
   e icono de línea. Aquí el color es el latón de la marca; el de Bloquear es el
   de peligro. */
.my-profile-card__edit.base-button.base-button--secondary {
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.my-profile-card__edit.base-button.base-button--secondary:hover:not(:disabled):not(
    .base-button--loading
  ) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.my-profile-card__edit.base-button.base-button--secondary:active:not(:disabled):not(
    .base-button--loading
  ) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 20%, transparent);
}

.my-profile-card__edit :deep(.base-button__content) {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.my-profile-card__edit-icon {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
}

@media (max-width: 640px) {
  .my-profile-card {
    width: 100%;
    padding: 32px 20px 24px;
  }

  .my-profile-card__facts {
    grid-auto-flow: row;
  }

  .my-profile-card__actions {
    flex-direction: column;
  }

  .my-profile-card__actions > * {
    width: 100%;
  }
}

@media (prefers-reduced-motion: reduce) {
  .my-profile-card {
    animation: none;
  }
}
</style>
