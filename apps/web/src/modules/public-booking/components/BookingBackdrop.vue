<script setup lang="ts">
// Fondo decorativo de la reserva pública: la misma mesa de patronaje del
// acceso (cuadrícula de trazo, curva de patrón que se dibuja sola, compás y
// regla con calibrador) recompuesta para una columna de lectura, más un
// foco de latón que sigue al cursor. Es puramente ornamental: sin texto, sin
// foco, oculto a las tecnologías de apoyo y sin capturar el puntero.
// Todo el movimiento es CSS sobre transform/opacity/stroke-dashoffset; el
// único JS es el foco, que solo escribe un `transform` por fotograma.
// `prefers-reduced-motion` deja cada pieza en su estado final y apaga el foco.
import { onBeforeUnmount, onMounted, ref } from 'vue'

const spot = ref<HTMLElement | null>(null)
const tracking = ref(false)
let frame = 0

function onPointerMove(event: PointerEvent) {
  // Un dedo no "apunta": en táctil el foco se quedaría clavado donde se tocó.
  if (event.pointerType === 'touch') return
  const { clientX, clientY } = event
  cancelAnimationFrame(frame)
  frame = requestAnimationFrame(() => {
    if (spot.value) spot.value.style.transform = `translate3d(${clientX}px, ${clientY}px, 0)`
    tracking.value = true
  })
}

onMounted(() => {
  const reduced =
    typeof window.matchMedia === 'function' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  if (reduced) return
  window.addEventListener('pointermove', onPointerMove, { passive: true })
})

onBeforeUnmount(() => {
  window.removeEventListener('pointermove', onPointerMove)
  cancelAnimationFrame(frame)
})
</script>

<template>
  <div class="pb-backdrop" aria-hidden="true">
    <span class="pb-backdrop__glow"></span>

    <span class="pb-backdrop__grid-mask">
      <span class="pb-backdrop__grid"></span>
    </span>

    <svg
      class="pb-backdrop__draft"
      viewBox="0 0 600 800"
      preserveAspectRatio="xMaxYMid slice"
      focusable="false"
    >
      <!-- Curva maestra: baja por el margen derecho como el contorno de una
           sisa, sin invadir la columna de lectura. -->
      <path
        class="pb-backdrop__curve pb-backdrop__curve--master"
        pathLength="1"
        d="M 640 60 C 520 120, 470 280, 505 430 S 470 700, 260 830"
      />
      <!-- Margen de costura: la misma curva desplazada, en puntada. -->
      <path
        class="pb-backdrop__curve pb-backdrop__curve--stitch"
        d="M 640 84 C 540 140, 494 282, 528 430 S 494 704, 290 830"
      />
      <!-- Compás con su centro marcado, abajo a la izquierda. -->
      <circle
        class="pb-backdrop__curve pb-backdrop__curve--compass"
        pathLength="1"
        cx="40"
        cy="760"
        r="210"
      />
      <circle
        class="pb-backdrop__curve pb-backdrop__curve--compass-inner"
        pathLength="1"
        cx="40"
        cy="760"
        r="130"
      />
      <g class="pb-backdrop__crosshair">
        <path d="M 28 760 H 52 M 40 748 V 772" />
      </g>
      <circle class="pb-backdrop__node" cx="505" cy="430" r="3.5" />
      <circle class="pb-backdrop__node pb-backdrop__node--late" cx="482" cy="610" r="2.5" />
    </svg>

    <span class="pb-backdrop__ruler">
      <span class="pb-backdrop__caliper"></span>
    </span>

    <span
      ref="spot"
      class="pb-backdrop__spot"
      :class="{ 'pb-backdrop__spot--on': tracking }"
    ></span>
  </div>
</template>

<style scoped>
.pb-backdrop {
  --backdrop-brass: var(--color-brand-accent-surface);

  position: fixed;
  inset: 0;
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
}

/* Resplandor de latón detrás del título: respira muy despacio para que el
   lienzo de tinta no se sienta plano, sin llegar a ser un efecto. */
.pb-backdrop__glow {
  position: absolute;
  top: 18%;
  left: 50%;
  width: min(760px, 120%);
  aspect-ratio: 1;
  transform: translate(-50%, -50%);
  background: radial-gradient(
    closest-side,
    color-mix(in srgb, var(--backdrop-brass) 14%, transparent),
    transparent
  );
  animation: pb-breathe 9s ease-in-out infinite alternate;
}

/* La máscara es estática; la rejilla interior es la que se desplaza (solo
   transform, compuesto en GPU), así el movimiento no repinta el degradado. */
.pb-backdrop__grid-mask {
  position: absolute;
  inset: 0;
  overflow: hidden;
  mask-image: radial-gradient(ellipse 78% 64% at 50% 28%, #000 18%, transparent 80%);
}

.pb-backdrop__grid {
  position: absolute;
  inset: -56px 0;
  background-image:
    linear-gradient(
      to right,
      color-mix(in srgb, var(--color-on-strong) 6%, transparent) 1px,
      transparent 1px
    ),
    linear-gradient(
      to bottom,
      color-mix(in srgb, var(--color-on-strong) 6%, transparent) 1px,
      transparent 1px
    );
  background-size: 56px 56px;
  animation:
    pb-fade 1.6s ease-out backwards,
    pb-drift 52s linear infinite;
}

.pb-backdrop__draft {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  opacity: 0.3;
}

.pb-backdrop__curve {
  fill: none;
  stroke: var(--backdrop-brass);
  stroke-width: 0.9;
  stroke-linecap: round;
}

/* `pathLength="1"` normaliza el trazo: el dibujado va de 1 a 0 sin medir. */
.pb-backdrop__curve--master {
  stroke-dasharray: 1;
  opacity: 0.6;
  animation: pb-draw 2.8s cubic-bezier(0.55, 0, 0.2, 1) 0.5s backwards;
}

.pb-backdrop__curve--stitch {
  stroke-dasharray: 5 8;
  opacity: 0.38;
  animation:
    pb-fade 1.4s ease-out 1.9s backwards,
    pb-march 14s linear 1.9s infinite;
}

.pb-backdrop__curve--compass,
.pb-backdrop__curve--compass-inner {
  stroke-dasharray: 1;
  opacity: 0.24;
  transform-origin: 40px 760px;
  transform: rotate(-90deg);
  animation: pb-draw 3.2s cubic-bezier(0.55, 0, 0.2, 1) 1.1s backwards;
}

.pb-backdrop__curve--compass-inner {
  opacity: 0.16;
  animation-delay: 1.5s;
}

.pb-backdrop__crosshair path {
  fill: none;
  stroke: var(--backdrop-brass);
  stroke-width: 0.9;
  opacity: 0.5;
  animation: pb-fade 0.8s ease-out 2.7s backwards;
}

.pb-backdrop__node {
  fill: var(--color-surface-strong);
  stroke: var(--backdrop-brass);
  stroke-width: 1;
  transform-box: fill-box;
  transform-origin: center;
  animation:
    pb-pop 0.6s cubic-bezier(0.3, 1.6, 0.5, 1) 2.3s backwards,
    pb-pulse 4.5s ease-in-out 3s infinite;
}

.pb-backdrop__node--late {
  animation-delay: 2.6s, 3.4s;
}

/* Regla de sastre en el borde izquierdo, solo en pantallas anchas: marcas
   cortas cada 12 px y largas cada 60 px, con un calibrador de latón que la
   recorre muy despacio. */
.pb-backdrop__ruler {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  display: none;
  width: 32px;
  background-image:
    linear-gradient(to bottom, var(--backdrop-brass) 1px, transparent 1px),
    linear-gradient(to bottom, var(--backdrop-brass) 1px, transparent 1px);
  background-size:
    9px 12px,
    18px 60px;
  background-repeat: repeat-y;
  opacity: 0.3;
  mask-image: linear-gradient(to bottom, transparent, #000 14%, #000 86%, transparent);
  animation: pb-fade 1.6s ease-out 0.6s backwards;
}

.pb-backdrop__caliper {
  position: absolute;
  top: 0;
  left: 20px;
  width: 7px;
  height: 7px;
  background-color: var(--backdrop-brass);
  transform: translateY(12vh) rotate(45deg);
  box-shadow: 0 0 10px 1px color-mix(in srgb, var(--backdrop-brass) 70%, transparent);
  animation: pb-caliper 22s ease-in-out infinite alternate;
}

/* Foco de latón que sigue al cursor. Un elemento de tamaño fijo que solo se
   traslada: el degradado no se repinta, y la transición suaviza el rastro. */
.pb-backdrop__spot {
  position: absolute;
  top: -260px;
  left: -260px;
  width: 520px;
  height: 520px;
  opacity: 0;
  background: radial-gradient(
    closest-side,
    color-mix(in srgb, var(--backdrop-brass) 13%, transparent),
    transparent
  );
  transition:
    transform 480ms cubic-bezier(0.2, 0.7, 0.2, 1),
    opacity 600ms ease;
  will-change: transform;
}

.pb-backdrop__spot--on {
  opacity: 1;
}

@keyframes pb-draw {
  from {
    stroke-dashoffset: 1;
  }
  to {
    stroke-dashoffset: 0;
  }
}

@keyframes pb-march {
  to {
    stroke-dashoffset: -26;
  }
}

@keyframes pb-fade {
  from {
    opacity: 0;
  }
}

@keyframes pb-pop {
  from {
    opacity: 0;
    transform: scale(0);
  }
}

@keyframes pb-pulse {
  50% {
    transform: scale(1.5);
  }
}

@keyframes pb-breathe {
  from {
    opacity: 0.55;
    transform: translate(-50%, -50%) scale(0.94);
  }
  to {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1.06);
  }
}

@keyframes pb-drift {
  to {
    transform: translateY(56px);
  }
}

@keyframes pb-caliper {
  from {
    transform: translateY(12vh) rotate(45deg);
  }
  to {
    transform: translateY(86vh) rotate(45deg);
  }
}

@media (min-width: 768px) {
  .pb-backdrop__draft {
    opacity: 0.8;
  }
}

@media (min-width: 1024px) {
  .pb-backdrop__ruler {
    display: block;
  }

  .pb-backdrop__draft {
    opacity: 1;
  }
}

/* Estado final sin movimiento: curvas dibujadas, puntos visibles, calibrador
   quieto a media altura y sin foco de cursor. */
@media (prefers-reduced-motion: reduce) {
  .pb-backdrop__glow,
  .pb-backdrop__grid,
  .pb-backdrop__curve,
  .pb-backdrop__crosshair path,
  .pb-backdrop__node,
  .pb-backdrop__ruler,
  .pb-backdrop__caliper {
    animation: none;
  }

  .pb-backdrop__caliper {
    transform: translateY(46vh) rotate(45deg);
  }

  .pb-backdrop__spot {
    display: none;
  }
}
</style>
