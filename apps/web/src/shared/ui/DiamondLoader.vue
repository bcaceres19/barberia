<script setup lang="ts">
/**
 * DiamondLoader - Espera con el rombo NAVA como protagonista.
 *
 * El rombo es la marca de la casa (divisor regla-rombo-regla, rombos de
 * estado de la agenda). Aquí se "talla" mientras la pantalla espera: el
 * contorno se dibuja solo, aparecen las facetas, un destello cruza la
 * piedra al girar y vuelve a empezar. Debajo, la regla NAVA hace de barra de
 * progreso: un rombo la recorre dejando estela de latón. Una frase de
 * barbería va cambiando cada pocos segundos para que la espera tenga voz.
 *
 * Accesibilidad: solo el `label` (texto exacto que decide el consumidor) es
 * accesible; el gráfico y las frases son decorativos (aria-hidden) para que
 * un lector no repita una frase distinta cada 2,6 s. El consumidor aporta el
 * role="status"/aria-live que le corresponda, igual que con AgendaSkeleton.
 * `prefers-reduced-motion: reduce` deja el rombo tallado y estático, sin
 * destello, sin rombo viajero y sin rotar frases.
 */
import { onBeforeUnmount, onMounted, ref, useId } from 'vue'

interface Props {
  /** Texto accesible de la espera, p. ej. "Cargando la agenda de …". */
  label: string
  /** Frases decorativas que rotan; por defecto, las de la casa. */
  phrases?: readonly string[]
  /** stacked: centrado en toda el área; inline: rombo a la izquierda. */
  layout?: 'stacked' | 'inline'
}

const props = withDefaults(defineProps<Props>(), {
  phrases: () => [
    'Afilando la navaja',
    'Tallando cada turno',
    'Puliendo el filo',
    'Alineando la agenda',
    'Todo a su hora',
  ],
  layout: 'stacked',
})

const PHRASE_INTERVAL_MS = 2600

const uid = useId()
const clipId = `${uid}-gem-clip`
const glintId = `${uid}-gem-glint`

const phraseIndex = ref(0)
let timer: ReturnType<typeof setInterval> | undefined

onMounted(() => {
  const reducedMotion =
    typeof window.matchMedia === 'function' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  if (reducedMotion || props.phrases.length < 2) return
  timer = setInterval(() => {
    phraseIndex.value = (phraseIndex.value + 1) % props.phrases.length
  }, PHRASE_INTERVAL_MS)
})

onBeforeUnmount(() => {
  if (timer !== undefined) clearInterval(timer)
})
</script>

<template>
  <div class="diamond-loader" :class="`diamond-loader--${layout}`">
    <span class="diamond-loader__label">{{ label }}</span>

    <div class="diamond-loader__stage" aria-hidden="true">
      <span class="diamond-loader__halo" />
      <svg class="diamond-loader__gem" viewBox="0 0 64 64" fill="none">
        <defs>
          <clipPath :id="clipId">
            <path d="M32 4 60 32 32 60 4 32Z" />
          </clipPath>
          <linearGradient :id="glintId" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0" stop-color="#f6e6c2" stop-opacity="0" />
            <stop offset="0.5" stop-color="#f6e6c2" stop-opacity="0.9" />
            <stop offset="1" stop-color="#f6e6c2" stop-opacity="0" />
          </linearGradient>
        </defs>
        <path class="diamond-loader__inner" d="M32 17 47 32 32 47 17 32Z" />
        <path class="diamond-loader__facets" d="M32 4V60M4 32H60M18 18 46 46M46 18 18 46" />
        <path class="diamond-loader__outline" d="M32 4 60 32 32 60 4 32Z" pathLength="100" />
        <g :clip-path="`url(#${clipId})`">
          <g transform="skewX(-22)">
            <rect
              class="diamond-loader__glint"
              x="-4"
              y="0"
              width="18"
              height="64"
              :fill="`url(#${glintId})`"
            />
          </g>
        </g>
      </svg>
      <span class="diamond-loader__spark diamond-loader__spark--a" />
      <span class="diamond-loader__spark diamond-loader__spark--b" />
      <span class="diamond-loader__spark diamond-loader__spark--c" />
    </div>

    <div class="diamond-loader__text" aria-hidden="true">
      <div class="diamond-loader__phrase-slot">
        <Transition name="diamond-phrase" mode="out-in">
          <span :key="phraseIndex" class="diamond-loader__phrase">{{ phrases[phraseIndex] }}</span>
        </Transition>
      </div>
      <span class="diamond-loader__rule">
        <span class="diamond-loader__rule-trail" />
        <span class="diamond-loader__rule-diamond" />
      </span>
    </div>
  </div>
</template>

<style scoped>
.diamond-loader {
  --gem-size: 100px;
  --gem-brass: var(--color-brand-accent-surface);

  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-4);
  padding: var(--space-8) var(--space-4);
  color: var(--color-on-strong);
}

.diamond-loader--stacked {
  flex: 1;
  flex-direction: column;
  text-align: center;
}

.diamond-loader--inline {
  --gem-size: 56px;

  flex-direction: row;
  justify-content: flex-start;
  padding: var(--space-2) 0;
  text-align: left;
}

.diamond-loader__label {
  position: absolute;
  width: 1px;
  height: 1px;
  margin: -1px;
  padding: 0;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}

.diamond-loader__stage {
  position: relative;
  flex: none;
  width: var(--gem-size);
  height: var(--gem-size);
}

.diamond-loader__halo {
  position: absolute;
  inset: 12%;
  background: radial-gradient(circle, rgb(184 149 90 / 34%) 0%, rgb(184 149 90 / 0%) 70%);
  border-radius: 50%;
  animation: diamond-halo 3.2s ease-in-out infinite;
}

.diamond-loader__gem {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: visible;
  animation: diamond-turn 3.2s cubic-bezier(0.45, 0, 0.25, 1) infinite;
}

.diamond-loader__outline {
  stroke: var(--gem-brass);
  stroke-width: 2;
  stroke-linejoin: miter;
  stroke-dasharray: 100;
  stroke-dashoffset: 100;
  animation: diamond-draw 3.2s cubic-bezier(0.55, 0, 0.2, 1) infinite;
}

.diamond-loader__facets {
  stroke: var(--gem-brass);
  stroke-width: 0.8;
  opacity: 0;
  animation: diamond-facets 3.2s ease-in-out infinite;
}

.diamond-loader__inner {
  fill: rgb(184 149 90 / 26%);
  stroke: rgb(184 149 90 / 70%);
  stroke-width: 1;
  opacity: 0;
  transform-box: fill-box;
  transform-origin: center;
  animation: diamond-inner 3.2s cubic-bezier(0.3, 0.7, 0.2, 1) infinite;
}

.diamond-loader__glint {
  transform: translateX(-40px);
  animation: diamond-glint 3.2s cubic-bezier(0.5, 0, 0.3, 1) infinite;
}

/* Destellos: rombos diminutos (el mismo motivo de los estados de la agenda)
   que se encienden alrededor de la piedra, cada uno a su ritmo. */
.diamond-loader__spark {
  position: absolute;
  width: 6px;
  height: 6px;
  background-color: var(--gem-brass);
  opacity: 0;
  transform: rotate(45deg) scale(0);
  animation: diamond-spark 3.2s ease-in-out infinite;
}

.diamond-loader__spark--a {
  top: 2%;
  right: -4%;
  animation-delay: 1.5s;
}

.diamond-loader__spark--b {
  bottom: 6%;
  left: -6%;
  width: 4px;
  height: 4px;
  animation-delay: 1.85s;
}

.diamond-loader__spark--c {
  top: 30%;
  left: -12%;
  width: 3px;
  height: 3px;
  animation-delay: 2.1s;
}

.diamond-loader__text {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-3);
}

.diamond-loader--inline .diamond-loader__text {
  align-items: flex-start;
}

.diamond-loader__phrase-slot {
  display: flex;
  align-items: center;
  min-width: 22ch;
  height: 30px;
}

.diamond-loader--stacked .diamond-loader__phrase-slot {
  justify-content: center;
}

.diamond-loader__phrase {
  font-family: var(--font-display);
  font-size: 24px;
  line-height: 30px;
  white-space: nowrap;
  color: var(--color-on-strong);
}

.diamond-phrase-enter-active,
.diamond-phrase-leave-active {
  transition:
    opacity 0.28s ease,
    transform 0.28s ease,
    letter-spacing 0.4s ease;
}

.diamond-phrase-enter-from {
  opacity: 0;
  transform: translateY(8px);
  letter-spacing: 0.08em;
}

.diamond-phrase-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* Regla NAVA como barra de progreso: línea tenue, estela de latón y un rombo
   que la recorre de punta a punta. */
.diamond-loader__rule {
  position: relative;
  display: block;
  width: 220px;
  height: 2px;
  background-color: rgb(244 240 231 / 14%);
}

.diamond-loader__rule-trail {
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, rgb(184 149 90 / 0%), var(--gem-brass));
  transform-origin: left center;
  animation: diamond-trail 2.6s cubic-bezier(0.6, 0, 0.3, 1) infinite;
}

.diamond-loader__rule-diamond {
  position: absolute;
  top: 50%;
  left: 0;
  width: 9px;
  height: 9px;
  margin: -4.5px 0 0 -4.5px;
  background-color: var(--gem-brass);
  box-shadow: 0 0 10px rgb(184 149 90 / 70%);
  transform: rotate(45deg);
  animation: diamond-travel 2.6s cubic-bezier(0.6, 0, 0.3, 1) infinite;
}

@keyframes diamond-draw {
  0% {
    stroke-dashoffset: 100;
  }
  42%,
  76% {
    stroke-dashoffset: 0;
  }
  100% {
    stroke-dashoffset: -100;
  }
}

@keyframes diamond-facets {
  0%,
  30% {
    opacity: 0;
  }
  50%,
  72% {
    opacity: 0.5;
  }
  92%,
  100% {
    opacity: 0;
  }
}

@keyframes diamond-inner {
  0%,
  34% {
    opacity: 0;
    transform: scale(0.2);
  }
  54%,
  74% {
    opacity: 1;
    transform: scale(1);
  }
  94%,
  100% {
    opacity: 0;
    transform: scale(1.15);
  }
}

@keyframes diamond-glint {
  0%,
  56% {
    transform: translateX(-40px);
  }
  80%,
  100% {
    transform: translateX(96px);
  }
}

/* La piedra "gira": se estrecha en el momento del destello, como si la luz
   la hiciera voltear. */
@keyframes diamond-turn {
  0%,
  52% {
    transform: scaleX(1);
  }
  66% {
    transform: scaleX(0.42);
  }
  80%,
  100% {
    transform: scaleX(1);
  }
}

@keyframes diamond-halo {
  0%,
  100% {
    opacity: 0.35;
    transform: scale(0.85);
  }
  60% {
    opacity: 1;
    transform: scale(1.1);
  }
}

@keyframes diamond-spark {
  0%,
  100% {
    opacity: 0;
    transform: rotate(45deg) scale(0);
  }
  10% {
    opacity: 1;
    transform: rotate(45deg) scale(1.2);
  }
  24% {
    opacity: 0;
    transform: rotate(45deg) scale(0.3);
  }
}

@keyframes diamond-trail {
  0% {
    transform: scaleX(0);
    opacity: 1;
  }
  70% {
    transform: scaleX(1);
    opacity: 1;
  }
  100% {
    transform: scaleX(1);
    opacity: 0;
  }
}

@keyframes diamond-travel {
  0% {
    left: 0;
    opacity: 1;
  }
  70% {
    left: 100%;
    opacity: 1;
  }
  100% {
    left: 100%;
    opacity: 0;
  }
}

@media (max-width: 480px) {
  .diamond-loader--inline {
    --gem-size: 44px;
  }

  .diamond-loader__phrase {
    font-size: 20px;
  }

  .diamond-loader__rule {
    width: 160px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .diamond-loader__halo,
  .diamond-loader__gem,
  .diamond-loader__facets,
  .diamond-loader__inner,
  .diamond-loader__glint,
  .diamond-loader__spark,
  .diamond-loader__rule-trail,
  .diamond-loader__rule-diamond {
    animation: none;
  }

  /* Rombo ya tallado, sin destello ni rombos viajeros. */
  .diamond-loader__outline {
    animation: none;
    stroke-dashoffset: 0;
  }

  .diamond-loader__inner {
    opacity: 1;
  }

  .diamond-loader__glint,
  .diamond-loader__spark {
    display: none;
  }

  .diamond-loader__rule-diamond {
    left: 50%;
  }

  .diamond-phrase-enter-active,
  .diamond-phrase-leave-active {
    transition: none;
  }
}
</style>
