<script setup lang="ts">
// Pantalla provisional que confirma que el arranque de la aplicación
// funciona. Se reemplaza por la primera pantalla real (docs/03-desarrollo/
// estandar-diseno-visual.md, sección 10) cuando comience el desarrollo de un
// módulo.
</script>

<template>
  <main class="status">
    <h1>Base del proyecto lista</h1>
    <p>Todavía no hay funcionalidades implementadas.</p>
  </main>
</template>

<style scoped>
.status {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  max-width: 640px;
  margin: var(--space-12) auto;
  padding: var(--space-6);
  color: var(--color-text-primary);
}
</style>
