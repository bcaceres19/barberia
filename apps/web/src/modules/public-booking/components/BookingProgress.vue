<script setup lang="ts">
// Progreso de la reserva: una regla de latón con un rombo por paso. El tramo
// recorrido se dibuja con una transición de `transform` cada vez que cambia
// el paso, así que persistir entre rutas (vive en el cascarón) es lo que lo
// hace perceptible. El estado nunca depende solo del color: el paso actual
// lleva `aria-current="step"`, halo y etiqueta más clara; los completados van
// rellenos y se anuncian como «completado»; los pendientes son contorno.
import { computed, onMounted, ref, watch } from 'vue'

interface Props {
  /** Paso actual, 1-based. Un valor mayor que el último paso significa que
   * la reserva terminó: todos los rombos quedan completados. */
  step: number
}

const props = defineProps<Props>()

const STEPS = ['Servicio', 'Barbero', 'Horario', 'Tus datos'] as const

// El tramo parte de cero en el primer montaje para que la regla se dibuje al
// abrir la reserva; después sigue a `step` con la misma transición.
const shownStep = ref(1)
onMounted(() => {
  requestAnimationFrame(() => {
    shownStep.value = props.step
  })
})
watch(
  () => props.step,
  (next) => {
    shownStep.value = next
  },
)

const fill = computed(() => {
  const clamped = Math.min(Math.max(shownStep.value, 1), STEPS.length)
  return (clamped - 1) / (STEPS.length - 1)
})

function stateOf(index: number): 'done' | 'current' | 'upcoming' {
  const position = index + 1
  if (position < props.step) return 'done'
  return position === props.step ? 'current' : 'upcoming'
}
</script>

<template>
  <nav class="pb-progress" aria-label="Progreso de la reserva" :style="{ '--pb-fill': fill }">
    <span class="pb-progress__track" aria-hidden="true">
      <span class="pb-progress__fill"></span>
    </span>
    <ol class="pb-progress__list">
      <li
        v-for="(label, index) in STEPS"
        :key="label"
        class="pb-progress__item"
        :class="`pb-progress__item--${stateOf(index)}`"
        :aria-current="stateOf(index) === 'current' ? 'step' : undefined"
      >
        <span class="pb-progress__node" aria-hidden="true"></span>
        <span class="pb-progress__label">{{ label }}</span>
        <span v-if="stateOf(index) === 'done'" class="pb-sr-only"> (completado)</span>
      </li>
    </ol>
  </nav>
</template>

<style scoped>
.pb-progress {
  position: relative;
  padding-top: var(--space-5);
}

.pb-progress__list {
  position: relative;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  padding: 0;
  margin: 0;
  list-style: none;
}

.pb-progress__item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
  text-align: center;
}

/* La regla cruza el centro de los rombos: de la mitad de la primera columna a
   la mitad de la última (12,5 % a cada lado con cuatro columnas iguales). */
.pb-progress__track {
  position: absolute;
  top: calc(var(--space-5) + 5px);
  right: 12.5%;
  left: 12.5%;
  height: 1px;
  background-color: var(--pb-line-strong);
}

.pb-progress__fill {
  display: block;
  height: 100%;
  background-color: var(--pb-brass);
  transform: scaleX(var(--pb-fill, 0));
  transform-origin: left;
  transition: transform 780ms cubic-bezier(0.65, 0, 0.2, 1) 120ms;
}

.pb-progress__node {
  position: relative;
  box-sizing: border-box;
  width: 11px;
  height: 11px;
  background-color: var(--color-surface-strong);
  border: 1px solid var(--pb-line-strong);
  transform: rotate(45deg);
  transition:
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard),
    transform 480ms cubic-bezier(0.3, 1.5, 0.5, 1);
}

.pb-progress__item--done .pb-progress__node {
  background-color: var(--pb-brass);
  border-color: var(--pb-brass);
}

.pb-progress__item--current .pb-progress__node {
  background-color: var(--pb-brass);
  border-color: var(--pb-brass);
  transform: rotate(45deg) scale(1.25);
}

/* Halo del paso actual: un segundo rombo que se abre y se desvanece. */
.pb-progress__item--current .pb-progress__node::after {
  content: '';
  position: absolute;
  inset: -1px;
  border: 1px solid var(--pb-brass);
  animation: pb-halo 2.6s ease-out 0.9s infinite;
}

.pb-progress__label {
  font-family: var(--font-sans);
  font-size: 10px;
  font-weight: 600;
  line-height: 14px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--pb-muted);
  transition: color var(--motion-duration-base) var(--motion-easing-standard);
}

.pb-progress__item--current .pb-progress__label,
.pb-progress__item--done .pb-progress__label {
  color: var(--pb-text);
}

.pb-progress__item--current .pb-progress__label {
  color: var(--pb-brass);
}

@keyframes pb-halo {
  0% {
    opacity: 0.8;
    transform: scale(1);
  }
  70%,
  100% {
    opacity: 0;
    transform: scale(2.6);
  }
}

@media (min-width: 480px) {
  .pb-progress__label {
    font-size: 11px;
    letter-spacing: 0.1em;
  }
}

@media (prefers-reduced-motion: reduce) {
  .pb-progress__fill,
  .pb-progress__node,
  .pb-progress__label {
    transition: none;
  }

  .pb-progress__item--current .pb-progress__node::after {
    animation: none;
    opacity: 0;
  }
}
</style>
