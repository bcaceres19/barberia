<script setup lang="ts">
// app compone «Servicios» por los índices públicos de `catalog` y
// `barberServices`; ninguno conoce al otro (mismo patrón que StaffWorkspacePage).
// En el perfil de barbero individual (DEC-115) cada servicio gana un
// interruptor «Lo ofrezco» —la asignación de HU-023 al único barbero— y un
// servicio nuevo queda ofrecido sin pasar por la matriz «Servicios por barbero».
// En el panel completo no se aporta nada: la pantalla es la de siempre.
import { onMounted, watch } from 'vue'
import { CatalogPage } from '@/modules/catalog'
import type { CatalogService } from '@/modules/catalog'
import { OfferServiceSwitch, useBarberOffering } from '@/modules/barberServices'
import { useToast } from '@/shared/composables'
import { isSoloProfile } from '@/shared/model'

const toast = useToast()
const offering = useBarberOffering()

async function loadOffering() {
  await offering.load()
  if (offering.status.value === 'error') {
    toast.error('No pudimos cargar qué servicios ofreces', {
      detail: 'Recarga la página para volver a intentarlo.',
    })
  }
}

// La marca llega del servidor poco después de entrar: si el perfil solo aparece cuando
// esta pantalla ya está montada, la carga empieza entonces (y no antes de saberlo).
onMounted(() => {
  if (isSoloProfile.value) void loadOffering()
})
watch(isSoloProfile, (solo) => {
  if (solo && offering.status.value === 'idle') void loadOffering()
})

// Ofrecer el servicio recién creado al único barbero; si falla, el servicio existe y la
// persona puede activarlo con su interruptor (el aviso lo dice).
async function onServiceCreated(service: CatalogService) {
  if (!isSoloProfile.value || offering.status.value !== 'ready') return
  const outcome = await offering.setOffered(service.id, true)
  if (outcome === 'network-error' || outcome === 'error') {
    toast.warning('El servicio se creó, pero aún no lo ofreces', {
      detail: 'Actívalo con «Lo ofrezco» para que aparezca en tu reserva pública.',
    })
  }
}
</script>

<template>
  <CatalogPage @service-created="onServiceCreated">
    <template
      v-if="isSoloProfile && offering.status.value === 'ready'"
      #service-offer="{ service }"
    >
      <OfferServiceSwitch
        compact
        :offering="offering"
        :service-id="service.id"
        :service-name="service.name"
      />
    </template>
  </CatalogPage>
</template>
