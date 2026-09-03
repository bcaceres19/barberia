<script setup lang="ts">
/**
 * RecordRow - Fila de carril para listas de registros (estandar-diseno-
 * visual.md §6.6). Sustituye la tarjeta blanca con borde por un renglón
 * de divisor fino, más cercano a la densidad Tailored Grid. No es
 * interactivo por sí mismo: cuando la fila completa navega, el consumidor
 * coloca un RouterLink/botón dentro de la región `main` o `trailing`, igual
 * que ya hace la agenda con el turno listado.
 */
interface Props {
  /** Elemento HTML raíz: 'li' dentro de una <ul>, 'div' en otro contenedor. */
  tag?: 'li' | 'div'
}

withDefaults(defineProps<Props>(), {
  tag: 'li',
})
</script>

<template>
  <component :is="tag" class="record-row">
    <span v-if="$slots.leading" class="record-row__leading">
      <slot name="leading" />
    </span>
    <div class="record-row__main">
      <slot />
    </div>
    <span v-if="$slots.trailing" class="record-row__trailing">
      <slot name="trailing" />
    </span>
  </component>
</template>

<style scoped>
.record-row {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-4) 0;
  border-bottom: var(--border-width-normal) solid var(--color-border-subtle);
}

.record-row:first-child {
  padding-top: 0;
}

.record-row:last-child {
  border-bottom: none;
}

.record-row__leading {
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.record-row__main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.record-row__trailing {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

@media (max-width: 480px) {
  .record-row {
    flex-wrap: wrap;
  }

  .record-row__trailing {
    flex-basis: 100%;
    justify-content: flex-start;
  }
}
</style>
