<script setup lang="ts">
/**
 * NavaWordmark - Identidad de plataforma (estandar-diseno-visual.md §2.1,
 * especificacion-frontend-nava.md §6). Texto accesible, nunca imagen
 * raster; variantes tinta (lienzo claro) e invertida (superficie tinta).
 * NAVA identifica la plataforma; el nombre de la barbería es dato del
 * tenant y se muestra por separado, nunca sustituido por este componente.
 */
interface Props {
  /** tinta: para lienzo/superficie clara. invertida: para superficie tinta. */
  variant?: 'ink' | 'inverted'
  /** sm: chip de identidad en cabeceras (20px). lg: marca editorial de un
   * panel de marca (paneles partidos de acceso/recuperación). */
  size?: 'sm' | 'lg'
}

withDefaults(defineProps<Props>(), {
  variant: 'ink',
  size: 'sm',
})
</script>

<template>
  <span class="nava-wordmark" :class="[`nava-wordmark--${variant}`, `nava-wordmark--${size}`]"
    >NAVA</span
  >
</template>

<style scoped>
.nava-wordmark {
  font-family: var(--font-display);
  line-height: 1;
  font-weight: 400;
  letter-spacing: 0.04em;
  color: var(--color-action-primary);
}

.nava-wordmark--inverted {
  color: var(--color-on-strong);
}

.nava-wordmark--sm {
  font-size: 20px;
}

.nava-wordmark--lg {
  font-size: var(--font-size-display);
  line-height: var(--font-size-display-line);
}
</style>
