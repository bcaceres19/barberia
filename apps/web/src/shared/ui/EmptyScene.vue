<script setup lang="ts">
/**
 * EmptyScene - Estado vacío del panel sobre tinta (DEC-113). Sustituye al
 * párrafo suelto o al divisor solitario de cada pantalla por una escena:
 * una ilustración de línea de latón que se dibuja sola (la agenda sin
 * turnos es una regla de horas con huecos, el catálogo una pila de fichas,
 * el equipo tres marcos de retrato), un titular serif, una ayuda y la acción
 * real, que el consumidor entrega por slot. Ningún motivo de barbería
 * (tijeras, navajas, postes): las ilustraciones son del mismo lenguaje de
 * regla, rombo y marco que el resto del panel.
 *
 * Es decorativo salvo por el texto: el SVG va `aria-hidden` y el titular, la
 * ayuda y la acción son contenido normal. `prefers-reduced-motion` y
 * `data-motion="reduced"` dejan la ilustración dibujada y quieta.
 */
withDefaults(
  defineProps<{
    /** Qué ilustra: huecos de agenda, fichas de catálogo, marcos de equipo,
     * un vínculo entre dos marcos o una búsqueda sin resultado. */
    scene: 'agenda' | 'services' | 'team' | 'assignments' | 'search'
  }>(),
  {},
)
</script>

<template>
  <div class="empty-scene" :class="`empty-scene--${scene}`">
    <span class="empty-scene__glow" aria-hidden="true" />

    <div class="empty-scene__art nv-pop" aria-hidden="true">
      <i class="empty-scene__spark empty-scene__spark--a" />
      <i class="empty-scene__spark empty-scene__spark--b" />
      <i class="empty-scene__spark empty-scene__spark--c" />

      <svg viewBox="0 0 240 150" fill="none" stroke="currentColor" stroke-width="1.5">
        <!-- Agenda: regla de horas, tres huecos libres y el «ahora» -->
        <g v-if="scene === 'agenda'">
          <path class="es-draw" pathLength="1" style="--d: 100" d="M16 108H224" />
          <path
            class="es-draw es-soft"
            pathLength="1"
            style="--d: 500"
            d="M16 108v8M50 108v8M84 108v8M118 108v8M152 108v8M186 108v8M220 108v8"
          />
          <rect class="es-slot" style="--d: 450" x="26" y="56" width="50" height="40" rx="2" />
          <rect
            class="es-slot es-slot--lead"
            style="--d: 650"
            x="96"
            y="44"
            width="60"
            height="52"
            rx="2"
          />
          <rect class="es-slot" style="--d: 850" x="174" y="62" width="42" height="34" rx="2" />
          <path class="es-plus" d="M126 64v18M117 73h18" />
          <path class="es-draw es-now-line" pathLength="1" style="--d: 1000" d="M84 28V108" />
          <circle class="es-ring" cx="84" cy="108" r="6" />
          <rect
            class="es-diamond es-pop"
            style="--d: 1250"
            x="78"
            y="102"
            width="12"
            height="12"
            transform="rotate(45 84 108)"
          />
        </g>

        <!-- Servicios: pila de fichas que sube y una acción por añadir -->
        <g v-else-if="scene === 'services'">
          <rect
            class="es-card es-soft es-rise"
            style="--d: 150"
            x="64"
            y="26"
            width="112"
            height="46"
            rx="2"
          />
          <rect
            class="es-card es-rise"
            style="--d: 350; opacity: 0.6"
            x="52"
            y="42"
            width="136"
            height="46"
            rx="2"
          />
          <rect
            class="es-card es-front es-rise"
            style="--d: 550"
            x="40"
            y="58"
            width="160"
            height="56"
            rx="2"
          />
          <path class="es-draw" pathLength="1" style="--d: 900" d="M58 78h64M58 94h38" />
          <circle class="es-plus-ring" cx="174" cy="86" r="14" />
          <path class="es-plus" d="M174 79v14M167 86h14" />
        </g>

        <!-- Equipo: tres marcos de retrato, el central por llenar -->
        <g v-else-if="scene === 'team'">
          <g class="es-rise es-soft" style="--d: 150">
            <rect x="24" y="48" width="54" height="66" rx="2" />
            <circle cx="51" cy="74" r="9" />
            <path d="M34 108c1-12 8-18 17-18s16 6 17 18" />
          </g>
          <g class="es-rise es-soft" style="--d: 350">
            <rect x="162" y="48" width="54" height="66" rx="2" />
            <circle cx="189" cy="74" r="9" />
            <path d="M172 108c1-12 8-18 17-18s16 6 17 18" />
          </g>
          <rect
            class="es-slot es-slot--lead es-rise"
            style="--d: 550"
            x="91"
            y="30"
            width="58"
            height="84"
            rx="2"
          />
          <path class="es-corner" d="M87 46V26h20M153 98v20h-20" />
          <path class="es-plus" d="M120 62v22M109 73h22" />
        </g>

        <!-- Servicios por barbero: dos marcos y un vínculo que late -->
        <g v-else-if="scene === 'assignments'">
          <g class="es-rise" style="--d: 150">
            <rect x="24" y="44" width="58" height="68" rx="2" />
            <circle cx="53" cy="72" r="10" />
            <path d="M34 106c1-13 9-20 19-20s18 7 19 20" />
          </g>
          <g class="es-rise" style="--d: 350">
            <rect class="es-card es-front" x="158" y="44" width="58" height="68" rx="2" />
            <path d="M172 66h30M172 80h30M172 94h18" />
          </g>
          <path class="es-link" d="M88 78H152" />
          <rect
            class="es-diamond es-travel"
            x="115"
            y="72"
            width="12"
            height="12"
            transform="rotate(45 121 78)"
          />
        </g>

        <!-- Búsqueda sin resultado -->
        <g v-else>
          <circle class="es-draw" pathLength="1" style="--d: 100" cx="108" cy="68" r="32" />
          <path
            class="es-draw"
            pathLength="1"
            style="--d: 700"
            d="M132 92L164 124"
            stroke-width="3"
          />
          <path class="es-draw es-soft" pathLength="1" style="--d: 900" d="M92 68h32" />
          <rect
            class="es-diamond es-pop"
            style="--d: 1200"
            x="102"
            y="82"
            width="10"
            height="10"
            transform="rotate(45 107 87)"
          />
        </g>
      </svg>
    </div>

    <p class="empty-scene__title nv-rise" style="--i: 2"><slot name="title" /></p>
    <p v-if="$slots.hint" class="empty-scene__hint nv-rise" style="--i: 3"><slot name="hint" /></p>
    <div v-if="$slots.action" class="empty-scene__action nv-rise" style="--i: 4">
      <slot name="action" />
    </div>
  </div>
</template>

<style scoped>
.empty-scene {
  position: relative;
  display: flex;
  min-height: 360px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  padding: var(--space-10) var(--space-5);
  overflow: hidden;
  text-align: center;
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: 3px;
  color: var(--color-brand-accent-surface);
}

/* Cuadrícula de puntos que se desvanece hacia los bordes: da al vacío un
   fondo de «hoja de trabajo» sin competir con el texto. */
.empty-scene::before {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background-image: radial-gradient(
    color-mix(in srgb, var(--color-on-strong) 12%, transparent) 1px,
    transparent 1.5px
  );
  background-size: 22px 22px;
  mask-image: radial-gradient(ellipse 60% 55% at 50% 38%, #000 0%, transparent 100%);
}

/* Filete de latón superior que se dibuja de izquierda a derecha. */
.empty-scene::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  left: 0;
  height: var(--border-width-emphasis);
  background: linear-gradient(
    90deg,
    transparent,
    var(--color-brand-accent-surface) 30%,
    var(--color-brand-accent-surface) 70%,
    transparent
  );
  animation: nava-wipe 900ms var(--motion-ease-out) backwards;
}

.empty-scene__glow {
  position: absolute;
  top: 46px;
  left: 50%;
  width: 280px;
  height: 200px;
  margin-left: -140px;
  pointer-events: none;
  background: radial-gradient(
    closest-side,
    color-mix(in srgb, var(--color-brand-accent-surface) 20%, transparent),
    transparent
  );
  animation: es-glow 5s steps(40, end) infinite;
}

.empty-scene__art {
  position: relative;
  width: 240px;
  height: 150px;
}

.empty-scene__art svg {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: visible;
}

/* Chispas: rombos pequeños que flotan alrededor de la ilustración. */
.empty-scene__spark {
  position: absolute;
  width: 7px;
  height: 7px;
  background-color: var(--color-brand-accent-surface);
  transform: rotate(45deg);
  animation: es-float 4.6s steps(46, end) infinite;
}

.empty-scene__spark--a {
  top: 10px;
  left: -6px;
  opacity: 0.7;
}

.empty-scene__spark--b {
  top: 30px;
  right: -14px;
  width: 5px;
  height: 5px;
  opacity: 0.5;
  animation-delay: -1.6s;
  animation-duration: 5.4s;
}

.empty-scene__spark--c {
  bottom: 14px;
  left: 20px;
  width: 4px;
  height: 4px;
  opacity: 0.6;
  animation-delay: -3s;
}

.empty-scene__title {
  position: relative;
  max-width: 30ch;
  margin: var(--space-2) 0 0;
  font-family: var(--font-display);
  font-size: var(--font-size-title-section);
  line-height: var(--font-size-title-section-line);
  color: var(--color-on-strong);
}

.empty-scene__hint {
  position: relative;
  max-width: 44ch;
  margin: 0;
  font-family: var(--font-sans);
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
  color: var(--color-on-strong-muted);
}

/* Enlace dentro de la ayuda (a otra sección del panel): latón subrayado. */
.empty-scene__hint :slotted(a) {
  color: var(--color-brand-accent-surface);
  text-decoration: underline;
  text-underline-offset: 3px;
}

.empty-scene__hint :slotted(a:hover) {
  color: var(--color-on-strong);
}

.empty-scene__action {
  position: relative;
  margin-top: var(--space-3);
}

/* La única acción de la escena: latón lleno con un destello que cruza cada
   pocos segundos (el primario de BaseButton es tinta y desaparecería). */
.empty-scene__action :deep(.base-button--primary) {
  position: relative;
  height: var(--control-height);
  padding-inline: 22px;
  overflow: hidden;
  font-size: var(--font-size-body-sm);
  font-weight: 600;
  color: var(--color-brand-accent-text);
  background-color: var(--color-brand-accent-surface);
  border-color: var(--color-brand-accent-surface);
}

.empty-scene__action :deep(.base-button--primary::after) {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: linear-gradient(105deg, transparent 35%, rgb(255 255 255 / 38%) 50%, transparent 65%);
  background-size: 250% 100%;
  background-position: -140% 0;
  animation: nava-glint 4.5s steps(45, end) 1.6s 3;
}

/* Costo de CPU (medido en Firefox): los bucles por pasos (`steps`) se recomponen
   unas pocas veces por segundo, no 60, y los que mueven el SVG o repintan el
   botón (anillo, «+», vínculo, destello) corren solo unas veces tras entrar y
   se detienen: el SVG se rasteriza en el hilo principal. Solo el resplandor y
   las chispas, que son capas HTML, siguen en bucle. */

/* Dibujo de la ilustración. `pathLength="1"` normaliza el trazo para que el
   mismo keyframe sirva a cualquier forma. */
.es-draw {
  stroke-dasharray: 1;
  stroke-dashoffset: 1;
  animation: es-draw 900ms var(--motion-ease-out) forwards;
  animation-delay: calc(var(--d, 0) * 1ms);
}

.es-soft {
  opacity: 0.55;
}

.es-slot {
  fill: color-mix(in srgb, var(--color-brand-accent-surface) 6%, transparent);
  stroke-dasharray: 4 4;
  opacity: 0;
  animation: es-fade 600ms var(--motion-ease-out) forwards;
  animation-delay: calc(var(--d, 0) * 1ms);
}

.es-slot--lead {
  fill: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
  stroke-width: 2;
}

.es-plus,
.es-plus-ring {
  opacity: 0;
  animation:
    es-fade 500ms var(--motion-ease-out) 1000ms forwards,
    es-breathe 2.8s steps(28, end) 1500ms 3;
  transform-box: fill-box;
  transform-origin: center;
  stroke-width: 2;
}

.es-plus-ring {
  stroke-width: 1.5;
  stroke-dasharray: 3 3;
}

.es-card {
  fill: var(--color-field-strong);
}

.es-front {
  fill: color-mix(in srgb, var(--color-brand-accent-surface) 8%, var(--color-field-strong));
}

.es-corner {
  stroke-width: 2;
  opacity: 0;
  animation: es-fade 500ms var(--motion-ease-out) 800ms forwards;
}

.es-rise,
.es-pop {
  opacity: 0;
  transform-box: fill-box;
  transform-origin: center;
  animation: es-rise 640ms var(--motion-ease-out) forwards;
  animation-delay: calc(var(--d, 0) * 1ms);
}

.es-pop {
  animation-name: es-pop;
}

.es-diamond {
  fill: var(--color-brand-accent-surface);
  stroke: none;
}

.es-now-line {
  stroke-width: 2;
}

.es-ring {
  fill: none;
  stroke: var(--color-brand-accent-surface);
  opacity: 0;
  transform-box: fill-box;
  transform-origin: center;
  animation: es-ring 2.4s steps(24, end) 1.6s 3;
}

.es-link {
  stroke-dasharray: 4 5;
  animation: es-march 1.2s steps(9, end) 6;
}

.es-travel {
  opacity: 0;
  animation: es-travel 3s steps(60, end) 600ms 4;
}

@keyframes es-draw {
  to {
    stroke-dashoffset: 0;
  }
}

@keyframes es-fade {
  to {
    opacity: 1;
  }
}

@keyframes es-rise {
  from {
    opacity: 0;
    transform: translateY(10px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes es-pop {
  0% {
    opacity: 0;
    transform: rotate(45deg) scale(0);
  }

  70% {
    opacity: 1;
    transform: rotate(45deg) scale(1.25);
  }

  100% {
    opacity: 1;
    transform: rotate(45deg) scale(1);
  }
}

@keyframes es-breathe {
  0%,
  100% {
    transform: scale(1);
  }

  50% {
    transform: scale(1.14);
  }
}

@keyframes es-ring {
  0% {
    opacity: 0.9;
    transform: scale(1);
  }

  100% {
    opacity: 0;
    transform: scale(3.4);
  }
}

@keyframes es-march {
  to {
    stroke-dashoffset: -9;
  }
}

@keyframes es-travel {
  0% {
    opacity: 0;
    transform: translateX(-34px) rotate(45deg);
  }

  20%,
  80% {
    opacity: 1;
  }

  100% {
    opacity: 0;
    transform: translateX(34px) rotate(45deg);
  }
}

@keyframes es-float {
  0%,
  100% {
    transform: translateY(0) rotate(45deg);
  }

  50% {
    transform: translateY(-9px) rotate(45deg);
  }
}

@keyframes es-glow {
  0%,
  100% {
    opacity: 0.7;
    transform: scale(1);
  }

  50% {
    opacity: 1;
    transform: scale(1.08);
  }
}

@media (max-width: 640px) {
  .empty-scene {
    min-height: 320px;
    padding: var(--space-8) var(--space-4);
  }

  .empty-scene__title {
    font-size: var(--font-size-title-item);
    line-height: var(--font-size-title-item-line);
  }
}

/* Movimiento reducido: la ilustración queda dibujada y quieta. */
@media (prefers-reduced-motion: reduce) {
  .es-draw {
    stroke-dashoffset: 0;
  }

  .es-slot,
  .es-plus,
  .es-plus-ring,
  .es-corner,
  .es-rise {
    opacity: 1;
  }

  .es-pop {
    opacity: 1;
    transform: rotate(45deg);
  }

  .es-travel {
    opacity: 1;
  }

  .es-draw,
  .es-slot,
  .es-plus,
  .es-plus-ring,
  .es-corner,
  .es-rise,
  .es-pop,
  .es-ring,
  .es-link,
  .es-travel,
  .empty-scene::after,
  .empty-scene__glow,
  .empty-scene__spark,
  .empty-scene__action :deep(.base-button--primary::after) {
    animation: none;
  }

  .es-ring {
    opacity: 0;
  }
}

:root[data-motion='reduced'] .es-draw {
  stroke-dashoffset: 0;
}

:root[data-motion='reduced'] .es-slot,
:root[data-motion='reduced'] .es-plus,
:root[data-motion='reduced'] .es-plus-ring,
:root[data-motion='reduced'] .es-corner,
:root[data-motion='reduced'] .es-rise {
  opacity: 1;
}

:root[data-motion='reduced'] .es-pop {
  opacity: 1;
  transform: rotate(45deg);
}

:root[data-motion='reduced'] .es-travel {
  opacity: 1;
}

:root[data-motion='reduced'] .es-draw,
:root[data-motion='reduced'] .es-slot,
:root[data-motion='reduced'] .es-plus,
:root[data-motion='reduced'] .es-plus-ring,
:root[data-motion='reduced'] .es-corner,
:root[data-motion='reduced'] .es-rise,
:root[data-motion='reduced'] .es-pop,
:root[data-motion='reduced'] .es-ring,
:root[data-motion='reduced'] .es-link,
:root[data-motion='reduced'] .es-travel,
:root[data-motion='reduced'] .empty-scene__glow,
:root[data-motion='reduced'] .empty-scene__spark {
  animation: none;
}

:root[data-motion='reduced'] .es-ring {
  opacity: 0;
}
</style>
