<script setup lang="ts">
/**
 * WeeklyBoard - La jornada semanal como un tablero: siete días, cada uno con
 * su pista horaria (barras de latón sobre una regla común) y, debajo, la lista
 * real de tramos con sus acciones. Los siete días siempre aparecen, con o sin
 * tramos (CA-040-01). El tablero no conoce la API: quien lo monta recibe los
 * eventos `add`/`edit`/`remove` y decide qué hacer (mismo contrato de estado
 * que la lista de antes: solo una respuesta exitosa cambia una fila).
 */
import { computed, ref } from 'vue'
import { BaseButton } from '@/shared/ui'
import { endLabel, formatHours, scaleTicks, type BoardScale } from '../model/weekBoard'
import type { WorkingHour } from '../model/workingHour'
import DayTrack from './DayTrack.vue'

interface BoardDay {
  value: number
  label: string
  items: WorkingHour[]
}

interface Props {
  days: readonly BoardDay[]
  scale: BoardScale
  /** Día ISO vigente en la barbería, o null si su zona no se conoce. */
  todayIsoWeekday: number | null
  /** Minuto del día vigente en la barbería, para el marcador "ahora". */
  nowMinute: number | null
  pendingDeleteIds: ReadonlySet<string>
}

const props = defineProps<Props>()

const emit = defineEmits<{
  add: [isoWeekday: number]
  edit: [workingHour: WorkingHour]
  remove: [workingHour: WorkingHour]
}>()

const ticks = computed(() => scaleTicks(props.scale))
const activeId = ref<string | null>(null)

function dayMinutes(day: BoardDay): number {
  return day.items.reduce((sum, item) => sum + item.durationMinutes, 0)
}

function countLabel(day: BoardDay): string {
  return day.items.length === 1 ? '1 tramo' : `${day.items.length} tramos`
}
</script>

<template>
  <div class="week-board">
    <div class="week-board__axis" aria-hidden="true">
      <div class="week-board__axis-scale">
        <span
          v-for="tick in ticks"
          :key="tick.minute"
          class="week-board__tick"
          :style="{ left: `${tick.percent}%` }"
          >{{ tick.label }}</span
        >
      </div>
    </div>

    <section
      v-for="(day, index) in days"
      :key="day.value"
      class="week-board__day"
      :class="{
        'week-board__day--today': day.value === todayIsoWeekday,
        'week-board__day--free': day.items.length === 0,
      }"
      :style="{ '--day-index': index }"
      :aria-labelledby="`schedules-day-${day.value}-title`"
    >
      <header class="week-board__day-header">
        <h2 :id="`schedules-day-${day.value}-title`" class="week-board__day-title">
          {{ day.label }}
        </h2>
        <span v-if="day.value === todayIsoWeekday" class="week-board__today">Hoy</span>
        <span class="week-board__day-count">
          <template v-if="day.items.length === 0">Día libre</template>
          <template v-else>
            {{ countLabel(day)
            }}<span class="week-board__day-hours">{{ formatHours(dayMinutes(day)) }}</span>
          </template>
        </span>
      </header>

      <div class="week-board__body">
        <DayTrack
          :segments="day.items"
          :scale="scale"
          :order="index"
          :active-id="activeId"
          :now-minute="day.value === todayIsoWeekday ? nowMinute : null"
        />

        <div class="week-board__line">
          <p v-if="day.items.length === 0" class="week-board__empty">Sin tramos.</p>

          <ul v-else class="week-board__list" :aria-label="`Tramos del ${day.label}`">
            <li
              v-for="wh in day.items"
              :key="wh.id"
              class="week-board__chip"
              @mouseenter="activeId = wh.id"
              @mouseleave="activeId = null"
              @focusin="activeId = wh.id"
              @focusout="activeId = null"
            >
              <span class="week-board__chip-text">
                <span class="week-board__chip-time">
                  {{ wh.startsTime }} · {{ wh.durationMinutes }} min
                </span>
                <span class="week-board__chip-until">
                  hasta {{ endLabel(wh.startsTime, wh.durationMinutes) }}
                </span>
              </span>
              <div class="week-board__chip-actions">
                <BaseButton
                  type="button"
                  variant="secondary"
                  :aria-label="`Editar tramo de ${day.label} a las ${wh.startsTime}`"
                  @click="emit('edit', wh)"
                >
                  Editar
                </BaseButton>
                <BaseButton
                  type="button"
                  variant="secondary"
                  class="week-board__remove"
                  :loading="pendingDeleteIds.has(wh.id)"
                  :disabled="pendingDeleteIds.has(wh.id)"
                  :aria-label="`Retirar tramo de ${day.label} a las ${wh.startsTime}`"
                  @click="emit('remove', wh)"
                >
                  Retirar
                </BaseButton>
              </div>
            </li>
          </ul>

          <button
            type="button"
            class="week-board__add"
            :aria-label="`Agregar un tramo el ${day.label}`"
            :title="`Agregar un tramo el ${day.label}`"
            @click="emit('add', day.value)"
          >
            <span aria-hidden="true">+</span>
          </button>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
/* Panel hundido sobre tinta: el mismo hueco que la tabla de Barberos. */
.week-board {
  --track-height: 14px;

  display: flex;
  min-height: 0;
  flex-direction: column;
  overflow: hidden;
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: 3px;
}

/* Regla horaria: sus etiquetas quedan sobre la columna de pistas, alineadas
   con las líneas de hora de cada día. */
.week-board__axis {
  position: sticky;
  top: 0;
  z-index: 2;
  display: grid;
  flex: none;
  grid-template-columns: 132px minmax(0, 1fr);
  gap: 18px;
  height: 30px;
  padding: 0 18px;
  align-items: end;
  background-color: var(--color-field-strong);
}

.week-board__axis::before {
  content: '';
}

.week-board__axis-scale {
  position: relative;
  height: 20px;
}

.week-board__tick {
  position: absolute;
  bottom: 6px;
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  letter-spacing: 0.08em;
  transform: translateX(-50%);
  opacity: 0.85;
}

.week-board__tick::after {
  content: '';
  position: absolute;
  bottom: -6px;
  left: 50%;
  width: 1px;
  height: 4px;
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 60%, transparent);
}

.week-board__day {
  position: relative;
  display: grid;
  flex: 1 0 auto;
  grid-template-columns: 132px minmax(0, 1fr);
  gap: 18px;
  padding: 7px 18px;
  border-top: var(--border-width-normal) solid var(--color-field-strong-border);
  transition: background-color var(--motion-duration-fast) var(--motion-easing-standard);
  /* Entrada escalonada: cada día arranca un paso después del anterior. */
  animation: week-board-day-in 360ms var(--motion-easing-standard) both;
  animation-delay: calc(var(--day-index, 0) * 55ms);
}

/* Hoy: filete de latón a la izquierda y un velo que se apaga hacia la derecha. */
.week-board__day--today {
  background-image: linear-gradient(
    90deg,
    color-mix(in srgb, var(--color-brand-accent-surface) 11%, transparent),
    color-mix(in srgb, var(--color-brand-accent-surface) 0%, transparent) 55%
  );
}

.week-board__day--today::before {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: 3px;
  background-color: var(--color-brand-accent-surface);
  transform-origin: top;
  animation: week-board-rule-in 520ms cubic-bezier(0.2, 0.8, 0.2, 1) 240ms both;
}

@media (hover: hover) {
  .week-board__day:hover {
    background-color: color-mix(in srgb, var(--color-on-strong) 3.5%, transparent);
  }
}

.week-board__day-header {
  display: flex;
  flex-wrap: wrap;
  align-content: flex-start;
  align-items: baseline;
  gap: 2px 8px;
  min-width: 0;
}

.week-board__day-title {
  margin: 0;
  color: var(--color-on-strong);
  font-family: var(--font-display);
  font-size: var(--font-size-title-item);
  font-weight: 400;
  line-height: 26px;
}

.week-board__today {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.week-board__today::before {
  content: '';
  width: 5px;
  height: 5px;
  background-color: currentColor;
  transform: rotate(45deg);
  animation: week-board-diamond 2.6s ease-in-out infinite;
}

.week-board__day-count {
  flex-basis: 100%;
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  line-height: var(--font-size-caption-line);
}

.week-board__day-hours {
  margin-left: 6px;
  padding-left: 6px;
  color: var(--color-on-strong);
  font-variant-numeric: tabular-nums;
  border-left: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-on-strong) 22%, transparent);
}

.week-board__day--free .week-board__day-title {
  color: var(--color-on-strong-muted);
}

.week-board__body {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 6px;
  justify-content: center;
}

.week-board__line {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.week-board__list {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  gap: 8px;
  padding: 0;
  margin: 0;
  list-style: none;
}

.week-board__chip {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 42px;
  padding: 3px 4px 3px 14px;
  background-color: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-on-strong) 12%, transparent);
  border-left: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  border-radius: 2px;
  transition:
    background-color var(--motion-duration-fast) var(--motion-easing-standard),
    border-color var(--motion-duration-fast) var(--motion-easing-standard),
    transform var(--motion-duration-base) var(--motion-easing-standard);
}

@media (hover: hover) {
  .week-board__chip:hover {
    background-color: color-mix(in srgb, var(--color-brand-accent-surface) 10%, transparent);
    border-color: color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent);
    transform: translateY(-1px);
  }
}

.week-board__chip:focus-within {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 10%, transparent);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent);
}

/* Hora y fin apilados: el tramo ocupa poco ancho y caben varios por día. */
.week-board__chip-text {
  display: flex;
  flex-direction: column;
  justify-content: center;
  line-height: 1.15;
}

.week-board__chip-time {
  color: var(--color-on-strong);
  font-size: var(--font-size-body-sm);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.week-board__chip-until {
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.week-board__chip-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-left: 4px;
}

/* Acciones de fila: mismo par fantasma de latón que Barberos y Servicios. */
.week-board__chip-actions :deep(.base-button) {
  --btn-focus-ring: 0 0 0 2px var(--color-field-strong), 0 0 0 4px var(--color-focus);

  height: 32px;
  padding-inline: 12px;
  font-size: var(--font-size-body-sm);
}

.week-board__chip-actions :deep(.base-button--secondary) {
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 40%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.week-board__chip-actions
  :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 14%, transparent);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
}

.week-board__chip-actions
  :deep(.base-button--secondary:active:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 22%, transparent);
}

/* Retirar es la acción destructiva: se tiñe de rosa solo al señalarla. */
.week-board__chip-actions
  :deep(
    .week-board__remove.base-button--secondary:hover:not(:disabled):not(.base-button--loading)
  ) {
  color: var(--color-danger-on-strong);
  background-color: color-mix(in srgb, var(--color-danger-on-strong) 10%, transparent);
  border-color: color-mix(in srgb, var(--color-danger-on-strong) 55%, transparent);
}

.week-board__empty {
  display: flex;
  align-items: center;
  min-height: 42px;
  padding: 0 14px;
  margin: 0;
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-body-sm);
  border: var(--border-width-normal) dashed
    color-mix(in srgb, var(--color-on-strong) 16%, transparent);
  border-radius: 2px;
}

/* "+" por día: ficha punteada que se rellena de latón al señalarla. */
.week-board__add {
  display: inline-grid;
  place-items: center;
  width: 42px;
  height: 42px;
  padding: 0;
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-title-item);
  line-height: 1;
  background: transparent;
  border: var(--border-width-normal) dashed
    color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-radius: 2px;
  cursor: pointer;
  transition:
    background-color var(--motion-duration-fast) var(--motion-easing-standard),
    color var(--motion-duration-fast) var(--motion-easing-standard),
    border-color var(--motion-duration-fast) var(--motion-easing-standard),
    transform var(--motion-duration-base) cubic-bezier(0.34, 1.56, 0.64, 1);
}

@media (hover: hover) {
  .week-board__add:hover {
    color: var(--color-brand-accent-text);
    background-color: var(--color-brand-accent-surface);
    border-style: solid;
    transform: rotate(90deg) scale(1.06);
  }
}

.week-board__add:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-field-strong),
    0 0 0 4px var(--color-focus);
}

@keyframes week-board-day-in {
  from {
    opacity: 0;
    transform: translateY(10px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes week-board-rule-in {
  from {
    transform: scaleY(0);
  }

  to {
    transform: scaleY(1);
  }
}

@keyframes week-board-diamond {
  0%,
  62%,
  100% {
    opacity: 1;
    transform: rotate(45deg) scale(1);
  }

  80% {
    opacity: 0.5;
    transform: rotate(45deg) scale(0.8);
  }
}

@media (max-width: 760px) {
  .week-board__axis {
    position: static;
    grid-template-columns: minmax(0, 1fr);
    padding: 0 12px;
  }

  .week-board__axis::before {
    display: none;
  }

  .week-board__day {
    grid-template-columns: minmax(0, 1fr);
    gap: 10px;
    padding: 14px 12px;
  }

  .week-board__day-header {
    flex-wrap: nowrap;
    justify-content: flex-start;
  }

  .week-board__day-title {
    font-size: var(--font-size-title-item);
  }

  .week-board__day-count {
    flex: 1;
    flex-basis: auto;
    text-align: right;
  }

  .week-board__chip {
    flex: 1 1 100%;
    flex-wrap: wrap;
    row-gap: 4px;
  }

  .week-board__chip-actions {
    margin-left: auto;
  }

  .week-board__line {
    gap: 8px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .week-board__day,
  .week-board__day--today::before,
  .week-board__today::before {
    animation: none;
  }

  .week-board__day,
  .week-board__chip,
  .week-board__add {
    transition: none;
  }
}
</style>
