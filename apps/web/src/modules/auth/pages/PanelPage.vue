<script setup lang="ts">
// Contenido de /panel dentro del cascarón de HU-012 (PrivateShell ya pone
// cabecera y navegación). Deliberadamente mínimo: no representa una agenda
// inexistente (fuera de alcance de esta historia, llega en B3). Sustituye a
// PanelPlaceholderPage.vue de HU-010, cuyo marcador de posición autenticado
// ya cumplió su propósito de demostrar la ruta privada real (DEC-056).
</script>

<template>
  <div class="panel-page">
    <h1 class="panel-page__title">Panel del barbero</h1>
    <p class="panel-page__hint">
      El contenido completo de esta sección (agenda, servicios, horarios) llega en bloques
      posteriores.
    </p>
  </div>
</template>

<style scoped>
.panel-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  max-width: 640px;
  margin: 0 auto;
  padding: var(--space-6) var(--space-4);
}

.panel-page__title {
  margin: 0;
  font-size: var(--font-size-h2);
  line-height: var(--font-size-h2-line);
  color: var(--color-text-primary);
}

.panel-page__hint {
  margin: 0;
  color: var(--color-text-secondary);
}
</style>
