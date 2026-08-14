<script setup lang="ts">
// Marcador de posición autenticado de `/panel` (DEC-056, CA-010-01).
// Deliberadamente mínimo: sin cabecera ni navegación general — ese
// cascarón completo lo construye HU-012 reutilizando el guard de esta
// ruta (`guards/requireSession.ts`), no uno paralelo. Esta página solo
// demuestra que la navegación posterior a un acceso válido llegó a una
// ruta privada real y protegida.
</script>

<template>
  <main class="panel-placeholder">
    <p class="panel-placeholder__eyebrow">Sesión iniciada</p>
    <h1 class="panel-placeholder__title">Panel del barbero</h1>
    <p class="panel-placeholder__hint">
      El contenido completo de esta sección lo construye HU-012.
    </p>
  </main>
</template>

<style scoped>
.panel-placeholder {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  max-width: 480px;
  margin: var(--space-12) auto;
  padding: var(--space-6);
  color: var(--color-text-primary);
}

.panel-placeholder__eyebrow {
  margin: 0;
  font-size: var(--font-size-body-sm);
  font-weight: 600;
  color: var(--color-success-text);
}

.panel-placeholder__title {
  margin: 0;
  font-size: var(--font-size-h2);
  line-height: var(--font-size-h2-line);
}

.panel-placeholder__hint {
  margin: 0;
  color: var(--color-text-secondary);
}
</style>
