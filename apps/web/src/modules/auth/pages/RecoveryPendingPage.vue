<script setup lang="ts">
// Destino del enlace de recuperación de CA-010-08 mientras HU-011 (la
// pantalla real de recuperación) no existe. `DP-UX-06`
// (docs/00-control/dudas-pendientes.md) registra que ninguna fuente
// aprobada define esta transición todavía: esta página es la
// interpretación aplicada mientras el propietario la confirma, no una
// simulación de la funcionalidad de HU-011. Por eso NO reproduce la
// composición P0 de "Recuperación" (estandar-diseno-visual.md §10) — ese
// paso actual/código/reenvío pertenece por completo a HU-011 — y declara
// con BaseAlert, de forma explícita, que todavía no está disponible.
import { BaseAlert } from '@/shared/ui'
</script>

<template>
  <main class="recovery-pending">
    <h1 class="recovery-pending__title">Recuperación de acceso</h1>
    <BaseAlert variant="info" title="Todavía no está disponible" role="status">
      La recuperación de acceso se está construyendo (HU-011). Mientras tanto, vuelve a la pantalla
      de acceso.
    </BaseAlert>
    <p class="recovery-pending__back">
      <RouterLink :to="{ name: 'acceso' }">Volver al acceso</RouterLink>
    </p>
  </main>
</template>

<style scoped>
.recovery-pending {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  max-width: 480px;
  margin: var(--space-12) auto;
  padding: var(--space-6);
}

.recovery-pending__title {
  margin: 0;
  font-size: var(--font-size-h2);
  line-height: var(--font-size-h2-line);
  color: var(--color-text-primary);
}

.recovery-pending__back {
  margin: 0;
}

.recovery-pending__back a {
  color: var(--color-action-primary);
  font-weight: 500;
}
</style>
