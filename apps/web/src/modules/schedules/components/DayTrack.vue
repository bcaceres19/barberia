<script setup lang="ts">
/**
 * DayTrack - Pista horaria de un día: una regla con marcas por hora y una
 * barra de latón por cada tramo. Decorativa (aria-hidden): los tramos reales
 * se leen en la lista que la acompaña. Las barras crecen desde el inicio al
 * montarse, escalonadas por `order`; `activeId` resalta la barra del tramo que
 * el usuario señala en la lista. Con `nowMinute` dibuja el marcador de "ahora".
 */
import { computed } from 'vue'
import { segmentGeometry, type BoardScale } from '../model/weekBoard'

interface TrackSegment {
  id: string
  startsTime: string
  durationMinutes: number
  /** Tramo de contexto (ya existente) bajo el que se está proponiendo otro. */
  muted?: boolean
}

interface Props {
  segments: readonly TrackSegment[]
  scale: BoardScale
  /** Posición del día en la semana: escalona la entrada de las barras. */
  order?: number
  activeId?: string | null
  nowMinute?: number | null
  /** Pista del diálogo: la barra es la propuesta y late con suavidad. */
  preview?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  order: 0,
  activeId: null,
  nowMinute: null,
  preview: false,
})

const hours = computed(() => (props.scale.endMinute - props.scale.startMinute) / 60)

const bars = computed(() =>
  props.segments
    .filter((s) => /^\d{2}:\d{2}$/.test(s.startsTime) && s.durationMinutes > 0)
    .map((s) => ({
      id: s.id,
      muted: s.muted === true,
      ...segmentGeometry(s.startsTime, s.durationMinutes, props.scale),
    })),
)

const nowPercent = computed(() => {
  if (props.nowMinute === null) return null
  const { startMinute, endMinute } = props.scale
  if (props.nowMinute < startMinute || props.nowMinute > endMinute) return null
  return ((props.nowMinute - startMinute) / (endMinute - startMinute)) * 100
})
</script>

<template>
  <div
    class="day-track"
    :class="{ 'day-track--preview': preview }"
    :style="{ '--track-hours': hours, '--track-order': order }"
    aria-hidden="true"
  >
    <span
      v-for="bar in bars"
      :key="bar.id"
      class="day-track__bar"
      :class="{
        'day-track__bar--active': bar.id === activeId,
        'day-track__bar--cut': bar.crossesMidnight,
        'day-track__bar--muted': bar.muted,
      }"
      :style="{ '--bar-from': `${bar.from}%`, '--bar-span': `${bar.span}%` }"
    />
    <span v-if="nowPercent !== null" class="day-track__now" :style="{ left: `${nowPercent}%` }" />
  </div>
</template>

<style scoped>
.day-track {
  position: relative;
  height: var(--track-height, 20px);
  background-color: color-mix(in srgb, var(--color-on-strong) 3%, transparent);
  /* Una línea por hora: la regla se lee sin necesidad de etiquetas en cada día. */
  background-image: linear-gradient(
    90deg,
    color-mix(in srgb, var(--color-on-strong) 10%, transparent) 1px,
    transparent 1px
  );
  background-size: calc(100% / var(--track-hours)) 100%;
  border-bottom: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-on-strong) 14%, transparent);
  border-radius: 2px;
}

.day-track__bar {
  position: absolute;
  top: 3px;
  bottom: 3px;
  left: var(--bar-from);
  width: var(--bar-span);
  overflow: hidden;
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--color-brand-accent-surface) 90%, black),
    var(--color-brand-accent-surface) 55%,
    color-mix(in srgb, var(--color-brand-accent-surface) 75%, white)
  );
  border-radius: 2px;
  box-shadow: 0 0 0 0 color-mix(in srgb, var(--color-brand-accent-surface) 0%, transparent);
  transform-origin: left center;
  transition:
    top var(--motion-duration-fast) var(--motion-easing-standard),
    bottom var(--motion-duration-fast) var(--motion-easing-standard),
    box-shadow var(--motion-duration-base) var(--motion-easing-standard),
    filter var(--motion-duration-fast) var(--motion-easing-standard);
  animation: day-track-grow 640ms cubic-bezier(0.2, 0.8, 0.2, 1) both;
  animation-delay: calc(var(--track-order) * 70ms + 180ms);
}

/* Un solo destello cruza la barra cuando termina de crecer. */
.day-track__bar::after {
  content: '';
  position: absolute;
  inset: 0 -60%;
  background: linear-gradient(
    75deg,
    transparent 40%,
    color-mix(in srgb, var(--color-on-strong) 55%, transparent) 50%,
    transparent 60%
  );
  transform: translateX(-100%);
  animation: day-track-glint 700ms cubic-bezier(0.5, 0, 0.3, 1) both;
  animation-delay: calc(var(--track-order) * 70ms + 760ms);
}

.day-track__bar--cut {
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
  -webkit-mask-image: linear-gradient(90deg, #000 88%, transparent);
  mask-image: linear-gradient(90deg, #000 88%, transparent);
}

.day-track__bar--muted {
  background: color-mix(in srgb, var(--color-brand-accent-surface) 28%, transparent);
  animation-delay: 0ms;
}

.day-track__bar--muted::after {
  display: none;
}

.day-track__bar--active {
  top: 0;
  bottom: 0;
  box-shadow: 0 0 14px 0 color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
  filter: brightness(1.12);
}

/* "Ahora": filete de marfil con un rombo que respira, sobre la barra. */
.day-track__now {
  position: absolute;
  top: -4px;
  bottom: -4px;
  width: 2px;
  margin-left: -1px;
  background-color: var(--color-on-strong);
  box-shadow: 0 0 8px color-mix(in srgb, var(--color-on-strong) 45%, transparent);
  animation: day-track-fade 400ms var(--motion-easing-standard) 900ms both;
}

.day-track__now::before {
  content: '';
  position: absolute;
  top: -3px;
  left: -3px;
  width: 8px;
  height: 8px;
  background-color: var(--color-on-strong);
  transform: rotate(45deg);
  animation: day-track-now-pulse 2.6s ease-in-out 1.3s infinite;
}

.day-track--preview .day-track__bar {
  animation-delay: 0ms;
  transition:
    left 280ms cubic-bezier(0.2, 0.8, 0.2, 1),
    width 280ms cubic-bezier(0.2, 0.8, 0.2, 1);
}

.day-track--preview .day-track__bar::after {
  animation: none;
}

@keyframes day-track-grow {
  from {
    opacity: 0;
    transform: scaleX(0);
  }

  to {
    opacity: 1;
    transform: scaleX(1);
  }
}

@keyframes day-track-glint {
  from {
    transform: translateX(-100%);
  }

  to {
    transform: translateX(100%);
  }
}

@keyframes day-track-fade {
  from {
    opacity: 0;
  }

  to {
    opacity: 1;
  }
}

@keyframes day-track-now-pulse {
  0%,
  100% {
    opacity: 1;
    transform: rotate(45deg) scale(1);
  }

  50% {
    opacity: 0.55;
    transform: rotate(45deg) scale(0.8);
  }
}

@media (prefers-reduced-motion: reduce) {
  .day-track__bar,
  .day-track__bar::after,
  .day-track__now,
  .day-track__now::before {
    animation: none;
    transition: none;
  }

  .day-track__bar::after {
    display: none;
  }
}
</style>
