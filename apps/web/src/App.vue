<script setup lang="ts">
// Composición raíz: monta el router y la región de avisos de las pantallas
// sin cascarón (acceso, recuperación, reserva pública). El cascarón privado
// aloja su propia región entre la cabecera y la navegación, y lo declara con
// `meta.toastHost` para que solo haya una región a la vez (DEC-095). Cualquier
// otro proveedor global se agrega en app/bootstrap, no aquí, según
// docs/04-arquitectura/frontend.md.
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { ToastRegion } from '@/shared/ui'

const route = useRoute()
const shellHostsToasts = computed(() =>
  route.matched.some((record) => record.meta.toastHost === 'shell'),
)
</script>

<template>
  <RouterView />
  <ToastRegion v-if="!shellHostsToasts" placement="viewport" />
</template>
