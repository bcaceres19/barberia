<script setup lang="ts">
// Rombo de selección de una opción de la reserva (servicio, barbero). Es
// decorativo: el estado lo comunica el `aria-checked` de la fila que lo
// contiene, y el aspecto (contorno, relleno, check dibujado) lo gobierna
// `.pb-choice--selected` en `styles/publicBooking.css`.
</script>

<template>
  <span class="pb-mark" aria-hidden="true">
    <svg viewBox="0 0 24 24" focusable="false">
      <path pathLength="1" d="M7.6 12.4l3 3 5.8-6.4" />
    </svg>
  </span>
</template>
