<script setup lang="ts">
// Cascarón de la reserva pública. Aloja las cinco pantallas del recorrido y
// conserva entre ellas lo que no debe reiniciarse: el fondo animado, la
// cabecera con la marca, el enlace de retorno y el progreso, que se dibuja
// de un paso al siguiente. Cada pantalla conserva su propio `<main>`, su
// estado y su `<h1>`; el cascarón solo aporta el marco (el `<header>` es el
// único landmark de banner) y la transición de entrada/salida entre pasos.
//
// Es un lienzo de tinta fijo: no hereda `data-app-theme` ni los acentos del
// panel privado (DEC-110: la reserva pública nunca los hereda) y NAVA aparece
// como firma secundaria, porque el nombre de la barbería es la jerarquía
// principal de la entrada (estandar-diseno-visual.md §3.2).
import { computed, provide, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, type RouteLocationRaw } from 'vue-router'
import { NavaWordmark } from '@/shared/ui'
import BookingBackdrop from '../components/BookingBackdrop.vue'
import BookingProgress from '../components/BookingProgress.vue'
import { bookingChromeKey } from '../model/bookingChrome'
import '../styles/publicBooking.css'

const route = useRoute()

const completed = ref(false)
provide(bookingChromeKey, { completed })

const TOTAL_STEPS = 4

const step = computed<number | null>(() => {
  const value = route.meta.bookingStep
  return typeof value === 'number' ? value : null
})

// Confirmado: el progreso pasa del último rombo y todos quedan completados.
const progressStep = computed(() => (completed.value ? TOTAL_STEPS + 1 : (step.value ?? 1)))

// Un paso atrás deshace exactamente el anterior: mismos parámetros menos el
// último. Es navegación entre rutas que ya existen, no una función nueva.
const back = computed<{ to: RouteLocationRaw; label: string } | null>(() => {
  if (completed.value) return null
  const slug = String(route.params.slug ?? '')
  const serviceId = String(route.params.serviceId ?? '')
  const barberId = String(route.params.barberId ?? '')
  switch (route.name) {
    case 'reserva-publica-servicios':
      return { to: { name: 'reserva-publica-entrada', params: { slug } }, label: 'Inicio' }
    case 'reserva-publica-barbero':
      return { to: { name: 'reserva-publica-servicios', params: { slug } }, label: 'Servicios' }
    case 'reserva-publica-horario':
      return {
        to: { name: 'reserva-publica-barbero', params: { slug, serviceId } },
        label: 'Barbero',
      }
    case 'reserva-publica-cliente':
      return {
        to: { name: 'reserva-publica-horario', params: { slug, serviceId, barberId } },
        label: 'Horario',
      }
    default:
      return null
  }
})

// La entrada no es un paso y ya se rotula «Reserva en línea» en su cuerpo:
// repetirlo en la cabecera solo quitaría ancho a la marca en 320 px.
const caption = computed(() => {
  if (completed.value) return 'Confirmado'
  return step.value === null ? '' : `Paso ${step.value} de ${TOTAL_STEPS}`
})

// La página sale hacia el lado contrario al volver atrás. Se decide con
// `flush: 'sync'` para que el nombre de la transición ya sea el correcto
// cuando Vue pinta la pantalla que sale.
const direction = ref<'forward' | 'back'>('forward')
watch(
  step,
  (next, previous) => {
    direction.value = (next ?? 0) < (previous ?? 0) ? 'back' : 'forward'
  },
  { flush: 'sync' },
)
</script>

<template>
  <div class="pb-layout">
    <BookingBackdrop />

    <header class="pb-header">
      <div class="pb-header__bar">
        <RouterLink v-if="back" :to="back.to" class="pb-back">
          <span class="pb-back__arrow" aria-hidden="true">‹</span>
          <span>{{ back.label }}</span>
        </RouterLink>
        <span v-else class="pb-header__spacer" aria-hidden="true"></span>

        <span class="pb-header__brand">
          <NavaWordmark variant="inverted" size="sm" />
        </span>

        <span class="pb-header__caption">{{ caption }}</span>
      </div>

      <BookingProgress v-if="step !== null" :step="progressStep" />
    </header>

    <RouterView v-slot="{ Component, route: current }">
      <Transition :name="`pb-step-${direction}`" mode="out-in">
        <component :is="Component" :key="String(current.name)" />
      </Transition>
    </RouterView>

    <!-- Firma secundaria (especificacion-frontend-nava.md §2.1): la barbería
         lleva la jerarquía principal; NAVA solo firma al pie, para que nadie
         crea que reserva en una barbería llamada NAVA. -->
    <footer class="pb-footer">
      <span>Reservas con</span>
      <NavaWordmark variant="inverted" size="sm" />
    </footer>
  </div>
</template>
