<script setup lang="ts">
/**
 * PageHeader - Cabecera de página del sistema visual NAVA / Tailored Grid
 * (estandar-diseno-visual.md §6.6). Título editorial (serif), contexto
 * opcional, enlace de retorno opcional y una acción primaria por región.
 * Se repite igual en las pantallas privadas para dar jerarquía y densidad
 * ordenada sin que cada pantalla reinvente su propio encabezado.
 */
import type { RouteLocationRaw } from 'vue-router'

interface Props {
  /** Título de la pantalla. Único <h1> real de la vista. */
  title: string
  /** Línea de contexto bajo el título (fecha, zona horaria, alcance). */
  subtitle?: string
  /** Destino del enlace de retorno. Sin backTo, no se muestra el enlace. */
  backTo?: RouteLocationRaw
  /** Texto del enlace de retorno. */
  backLabel?: string
}

withDefaults(defineProps<Props>(), {
  subtitle: undefined,
  backTo: undefined,
  backLabel: 'Volver',
})
</script>

<template>
  <header class="page-header">
    <RouterLink v-if="backTo" :to="backTo" class="page-header__back">
      ← {{ backLabel }}
    </RouterLink>
    <div class="page-header__row">
      <div class="page-header__heading">
        <h1 class="page-header__title">
          {{ title }}
          <slot name="badge" />
        </h1>
        <p v-if="subtitle" class="page-header__subtitle">{{ subtitle }}</p>
      </div>
      <div v-if="$slots.actions" class="page-header__actions">
        <slot name="actions" />
      </div>
    </div>
  </header>
</template>

<style scoped>
.page-header {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  padding-bottom: var(--space-5);
  margin-bottom: var(--space-6);
  border-bottom: var(--border-width-normal) solid var(--color-border-subtle);
}

.page-header__back {
  align-self: flex-start;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  color: var(--color-text-secondary);
  text-decoration: none;
}

.page-header__back:hover {
  color: var(--color-text-primary);
  text-decoration: underline;
}

.page-header__row {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
}

.page-header__heading {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  min-width: 0;
}

.page-header__title {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h1);
  line-height: var(--font-size-h1-line);
  font-weight: var(--font-weight-h1);
  color: var(--color-text-primary);
}

.page-header__subtitle {
  margin: 0;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
  color: var(--color-text-secondary);
}

.page-header__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

@media (max-width: 480px) {
  .page-header__row {
    flex-direction: column;
    align-items: stretch;
  }

  .page-header__actions {
    flex-direction: column;
  }
}
</style>
