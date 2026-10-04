<script setup lang="ts">
// Fondo decorativo del panel de marca de acceso/recuperación: la mesa de
// patronaje del sastre (Tailored Grid) — cuadrícula de trazo, curvas de
// patrón que se dibujan solas, una costura que avanza y una regla con
// calibrador. Es puramente ornamental: sin texto, sin foco y oculto a las
// tecnologías de apoyo; el contenido real vive en el resto del panel.
// Todo el movimiento es CSS sobre transform/opacity/stroke-dashoffset (sin
// JS ni dependencias) y `prefers-reduced-motion` lo deja en su estado final.
</script>

<template>
  <div class="brand-backdrop" aria-hidden="true">
    <span class="brand-backdrop__glow"></span>

    <span class="brand-backdrop__grid-mask">
      <span class="brand-backdrop__grid"></span>
    </span>

    <svg
      class="brand-backdrop__draft"
      viewBox="0 0 600 800"
      preserveAspectRatio="xMidYMid slice"
      focusable="false"
    >
      <!-- Curva maestra del patrón: entra por abajo-izquierda y sale por
           arriba-derecha, rodeando la marca sin tocarla. -->
      <path
        class="brand-backdrop__curve brand-backdrop__curve--master"
        pathLength="1"
        d="M -40 780 C 60 700, 130 540, 110 340 S 230 70, 640 40"
      />
      <!-- Margen de costura: la misma curva desplazada, en puntada. -->
      <path
        class="brand-backdrop__curve brand-backdrop__curve--stitch"
        d="M -40 802 C 70 722, 152 540, 132 340 S 250 94, 640 62"
      />
      <!-- Compás: circunferencia con su centro marcado. -->
      <circle
        class="brand-backdrop__curve brand-backdrop__curve--compass"
        pathLength="1"
        cx="560"
        cy="790"
        r="230"
      />
      <circle
        class="brand-backdrop__curve brand-backdrop__curve--compass-inner"
        pathLength="1"
        cx="560"
        cy="790"
        r="150"
      />
      <g class="brand-backdrop__crosshair">
        <path d="M 548 790 H 572 M 560 778 V 802" />
      </g>
      <!-- Puntos de construcción sobre la curva maestra. -->
      <circle class="brand-backdrop__node" cx="110" cy="340" r="3.5" />
      <circle class="brand-backdrop__node brand-backdrop__node--late" cx="80" cy="605" r="2.5" />
    </svg>

    <span class="brand-backdrop__ruler">
      <span class="brand-backdrop__caliper"></span>
    </span>
  </div>
</template>

<style scoped>
.brand-backdrop {
  --backdrop-brass: var(--color-brand-accent-surface);

  position: absolute;
  inset: 0;
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
}

/* Resplandor de latón detrás de la marca: respira muy despacio para que el
   panel de tinta no se sienta plano, sin llegar a ser un efecto. */
.brand-backdrop__glow {
  position: absolute;
  top: 38%;
  left: 50%;
  width: min(640px, 90%);
  aspect-ratio: 1;
  transform: translate(-50%, -50%);
  background: radial-gradient(
    closest-side,
    color-mix(in srgb, var(--backdrop-brass) 16%, transparent),
    transparent
  );
  animation: backdrop-breathe 9s ease-in-out infinite alternate;
}

/* La máscara es estática; la rejilla interior es la que se desplaza (solo
   transform, compuesto en GPU), así el movimiento no repinta el degradado. */
.brand-backdrop__grid-mask {
  position: absolute;
  inset: 0;
  overflow: hidden;
  mask-image: radial-gradient(ellipse 70% 62% at 50% 44%, #000 25%, transparent 78%);
}

.brand-backdrop__grid {
  position: absolute;
  inset: -56px 0;
  background-image:
    linear-gradient(
      to right,
      color-mix(in srgb, var(--color-on-strong) 7%, transparent) 1px,
      transparent 1px
    ),
    linear-gradient(
      to bottom,
      color-mix(in srgb, var(--color-on-strong) 7%, transparent) 1px,
      transparent 1px
    );
  background-size: 56px 56px;
  animation:
    backdrop-fade 1.6s ease-out backwards,
    backdrop-drift 48s linear infinite;
}

.brand-backdrop__draft {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

.brand-backdrop__curve {
  fill: none;
  stroke: var(--backdrop-brass);
  stroke-width: 0.9;
  stroke-linecap: round;
}

/* `pathLength="1"` normaliza el trazo: el dibujado va de 1 a 0 sin medir. */
.brand-backdrop__curve--master {
  stroke-dasharray: 1;
  opacity: 0.6;
  animation: backdrop-draw 2.8s cubic-bezier(0.55, 0, 0.2, 1) 0.4s backwards;
}

.brand-backdrop__curve--stitch {
  stroke-dasharray: 5 8;
  opacity: 0.4;
  animation:
    backdrop-fade 1.4s ease-out 1.8s backwards,
    backdrop-march 14s linear 1.8s infinite;
}

.brand-backdrop__curve--compass,
.brand-backdrop__curve--compass-inner {
  stroke-dasharray: 1;
  opacity: 0.26;
  transform-origin: 560px 790px;
  transform: rotate(-90deg);
  animation: backdrop-draw 3.2s cubic-bezier(0.55, 0, 0.2, 1) 1s backwards;
}

.brand-backdrop__curve--compass-inner {
  opacity: 0.18;
  animation-delay: 1.4s;
}

.brand-backdrop__crosshair path {
  fill: none;
  stroke: var(--backdrop-brass);
  stroke-width: 0.9;
  opacity: 0.5;
  animation: backdrop-fade 0.8s ease-out 2.6s backwards;
}

.brand-backdrop__node {
  fill: var(--color-surface-strong);
  stroke: var(--backdrop-brass);
  stroke-width: 1;
  transform-box: fill-box;
  transform-origin: center;
  animation:
    backdrop-pop 0.6s cubic-bezier(0.3, 1.6, 0.5, 1) 2.2s backwards,
    backdrop-pulse 4.5s ease-in-out 3s infinite;
}

.brand-backdrop__node--late {
  animation-delay: 2.5s, 3.4s;
}

/* Regla de sastre en el borde izquierdo: marcas cortas cada 12 px y largas
   cada 60 px, con un calibrador de latón que la recorre muy despacio. */
.brand-backdrop__ruler {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: 32px;
  background-image:
    linear-gradient(to bottom, var(--backdrop-brass) 1px, transparent 1px),
    linear-gradient(to bottom, var(--backdrop-brass) 1px, transparent 1px);
  background-size:
    9px 12px,
    18px 60px;
  background-repeat: repeat-y;
  opacity: 0.34;
  mask-image: linear-gradient(to bottom, transparent, #000 14%, #000 86%, transparent);
  animation: backdrop-fade 1.6s ease-out 0.6s backwards;
}

.brand-backdrop__caliper {
  position: absolute;
  top: 0;
  left: 20px;
  width: 7px;
  height: 7px;
  background-color: var(--backdrop-brass);
  transform: translateY(12vh) rotate(45deg);
  box-shadow: 0 0 10px 1px color-mix(in srgb, var(--backdrop-brass) 70%, transparent);
  animation: backdrop-caliper 22s ease-in-out infinite alternate;
}

@keyframes backdrop-draw {
  from {
    stroke-dashoffset: 1;
  }
  to {
    stroke-dashoffset: 0;
  }
}

@keyframes backdrop-march {
  to {
    stroke-dashoffset: -26;
  }
}

@keyframes backdrop-fade {
  from {
    opacity: 0;
  }
}

@keyframes backdrop-pop {
  from {
    opacity: 0;
    transform: scale(0);
  }
}

@keyframes backdrop-pulse {
  50% {
    transform: scale(1.5);
  }
}

@keyframes backdrop-breathe {
  from {
    opacity: 0.55;
    transform: translate(-50%, -50%) scale(0.94);
  }
  to {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1.06);
  }
}

@keyframes backdrop-drift {
  to {
    transform: translateY(56px);
  }
}

@keyframes backdrop-caliper {
  from {
    transform: translateY(12vh) rotate(45deg);
  }
  to {
    transform: translateY(86vh) rotate(45deg);
  }
}

/* Estado final sin movimiento: curvas dibujadas, puntos visibles, calibrador
   quieto a media altura. Nada depende de que la animación corra. */
@media (prefers-reduced-motion: reduce) {
  .brand-backdrop__glow,
  .brand-backdrop__grid,
  .brand-backdrop__curve,
  .brand-backdrop__crosshair path,
  .brand-backdrop__node,
  .brand-backdrop__ruler,
  .brand-backdrop__caliper {
    animation: none;
  }

  .brand-backdrop__caliper {
    transform: translateY(46vh) rotate(45deg);
  }
}
</style>
