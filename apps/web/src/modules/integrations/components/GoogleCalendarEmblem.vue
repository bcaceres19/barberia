<script setup lang="ts">
// Emblema de la pantalla «Google Calendar» (issue #331): una hoja de calendario con el
// rombo de NAVA y una insignia de estado. Es solo ilustración (aria-hidden): el estado
// real lo comunican el badge y el texto de la pantalla. Todo el movimiento es CSS sobre
// `transform`, `opacity` y `stroke-dashoffset`; con `prefers-reduced-motion` o
// `data-motion="reduced"` queda dibujado y quieto.
import { computed } from 'vue'
import type { GoogleCalendarView } from '../model/googleCalendar'

const props = defineProps<{ view: GoogleCalendarView }>()

const isLive = computed(() => props.view === 'connected' || props.view === 'syncing')
const isAlert = computed(() => props.view === 'reauth' || props.view === 'sync-error')
</script>

<template>
  <svg
    class="gcal-emblem"
    :class="[`gcal-emblem--${view}`]"
    viewBox="0 0 112 112"
    width="112"
    height="112"
    aria-hidden="true"
    focusable="false"
  >
    <!-- Órbita: gira solo mientras se sincroniza. -->
    <circle
      class="gcal-emblem__orbit"
      cx="56"
      cy="56"
      r="52"
      fill="none"
      pathLength="100"
      stroke-dasharray="26 74"
      stroke-linecap="round"
    />
    <!-- Anillo que respira cuando la conexión está viva. -->
    <circle v-if="isLive" class="gcal-emblem__pulse" cx="56" cy="56" r="40" fill="none" />

    <g class="gcal-emblem__sheet">
      <rect
        class="gcal-emblem__body gcal-draw"
        x="22"
        y="28"
        width="68"
        height="64"
        rx="4"
        fill="none"
        pathLength="1"
      />
      <path class="gcal-emblem__band" d="M22 46h68" fill="none" pathLength="1" />
      <path class="gcal-emblem__ring" d="M40 21v14M72 21v14" fill="none" />
      <!-- Rejilla de días: se enciende en cascada. -->
      <g class="gcal-emblem__days">
        <rect
          v-for="(d, i) in [0, 1, 2, 3, 4, 5]"
          :key="d"
          class="gcal-emblem__day"
          :style="{ '--d': i }"
          :x="32 + (i % 3) * 18"
          :y="56 + Math.floor(i / 3) * 14"
          width="10"
          height="8"
          rx="1.5"
        />
      </g>
      <!-- Rombo NAVA: hueco hasta que la agenda se publica. -->
      <rect
        class="gcal-emblem__diamond"
        :class="{ 'gcal-emblem__diamond--filled': isLive }"
        x="51"
        y="71"
        width="10"
        height="10"
        transform="rotate(45 56 76)"
      />
    </g>

    <!-- Insignia de estado. -->
    <g :key="view" class="gcal-emblem__badge">
      <circle cx="88" cy="88" r="15" class="gcal-emblem__badge-disc" />
      <path
        v-if="view === 'connected'"
        class="gcal-emblem__glyph"
        d="M81 88.5l5 5 9-10"
        fill="none"
      />
      <path
        v-else-if="view === 'syncing'"
        class="gcal-emblem__glyph gcal-emblem__glyph--spin"
        d="M81 88a7 7 0 0 1 12-5M95 88a7 7 0 0 1-12 5"
        fill="none"
      />
      <path v-else-if="isAlert" class="gcal-emblem__glyph" d="M88 80v9M88 94.5v.5" fill="none" />
      <path
        v-else-if="view === 'unavailable'"
        class="gcal-emblem__glyph"
        d="M81 88h14"
        fill="none"
      />
      <path v-else class="gcal-emblem__glyph" d="M88 80v16M80 88h16" fill="none" />
    </g>
  </svg>
</template>

<style scoped>
.gcal-emblem {
  --emblem-tone: var(--gcal-tone, var(--color-brand-accent-surface));

  display: block;
  flex: 0 0 auto;
  overflow: visible;
}

.gcal-emblem__orbit {
  stroke: var(--emblem-tone);
  stroke-width: 1.5;
  opacity: 0.35;
  transform-origin: 56px 56px;
}

.gcal-emblem--syncing .gcal-emblem__orbit {
  opacity: 0.95;
  animation: gcal-orbit 1.6s linear infinite;
}

.gcal-emblem__pulse {
  stroke: var(--emblem-tone);
  stroke-width: 1.5;
  transform-box: fill-box;
  transform-origin: center;
  animation: gcal-breathe 3.2s var(--motion-ease-out, ease-out) infinite;
}

.gcal-emblem__body,
.gcal-emblem__band,
.gcal-emblem__ring {
  stroke: var(--color-on-strong);
  stroke-width: 2;
  stroke-linecap: round;
  opacity: 0.88;
}

.gcal-emblem__band {
  stroke: var(--emblem-tone);
  opacity: 1;
}

.gcal-emblem__ring {
  stroke: var(--emblem-tone);
  stroke-width: 3;
  opacity: 1;
}

.gcal-draw {
  stroke-dasharray: 1;
  stroke-dashoffset: 0;
  animation: gcal-draw 900ms var(--motion-ease-out, ease-out) backwards;
}

.gcal-emblem__band {
  stroke-dasharray: 1;
  animation: gcal-draw 700ms var(--motion-ease-out, ease-out) 250ms backwards;
}

.gcal-emblem__day {
  fill: color-mix(in srgb, var(--color-on-strong) 22%, transparent);
  transform-box: fill-box;
  transform-origin: center;
  animation: nava-pop 420ms var(--motion-ease-spring, ease-out) backwards;
  animation-delay: calc(500ms + var(--d, 0) * 70ms);
}

.gcal-emblem--connected .gcal-emblem__day,
.gcal-emblem--syncing .gcal-emblem__day {
  fill: color-mix(in srgb, var(--emblem-tone) 55%, transparent);
}

.gcal-emblem__diamond {
  fill: transparent;
  stroke: var(--emblem-tone);
  stroke-width: 1.6;
  transition:
    fill 400ms var(--motion-ease-out, ease-out),
    transform 400ms var(--motion-ease-out, ease-out);
}

.gcal-emblem__diamond--filled {
  fill: var(--emblem-tone);
}

.gcal-emblem__badge {
  transform-box: fill-box;
  transform-origin: center;
  animation: nava-pop 520ms var(--motion-ease-spring, ease-out) 300ms backwards;
}

.gcal-emblem__badge-disc {
  fill: var(--color-surface-strong);
  stroke: var(--emblem-tone);
  stroke-width: 2;
}

.gcal-emblem__glyph {
  stroke: var(--emblem-tone);
  stroke-width: 2.4;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.gcal-emblem--connected .gcal-emblem__glyph {
  stroke-dasharray: 24;
  animation: gcal-check 520ms var(--motion-ease-out, ease-out) 600ms backwards;
}

.gcal-emblem__glyph--spin {
  transform-box: fill-box;
  transform-origin: center;
  animation: gcal-orbit 1.2s linear infinite;
}

@keyframes gcal-orbit {
  to {
    transform: rotate(360deg);
  }
}

@keyframes gcal-breathe {
  0% {
    opacity: 0.55;
    transform: scale(0.9);
  }

  80%,
  100% {
    opacity: 0;
    transform: scale(1.3);
  }
}

@keyframes gcal-draw {
  from {
    stroke-dashoffset: 1;
  }

  to {
    stroke-dashoffset: 0;
  }
}

@keyframes gcal-check {
  from {
    stroke-dashoffset: 24;
  }

  to {
    stroke-dashoffset: 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .gcal-emblem *,
  .gcal-emblem--syncing .gcal-emblem__orbit {
    animation: none !important;
    transition: none !important;
  }

  .gcal-emblem__pulse {
    opacity: 0.25;
  }
}
</style>

<style>
:root[data-motion='reduced'] .gcal-emblem *,
:root[data-motion='reduced'] .gcal-emblem--syncing .gcal-emblem__orbit {
  animation: none !important;
  transition: none !important;
}

:root[data-motion='reduced'] .gcal-emblem__pulse {
  opacity: 0.25;
}
</style>
