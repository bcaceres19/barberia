<script setup lang="ts">
/**
 * NewAppointmentSkeleton - Espera con la geometría del formulario por llegar
 * (issue #190, atlas nuevo-turno-eventos evento 01): mientras barberos y zona
 * horaria siguen sin resolverse, la pantalla conserva la hoja de cuatro
 * bandas numeradas y sus campos en vez de dejar un indicador suelto sobre una
 * pantalla vacía. Hermano de `AgendaSkeleton` (evento 05 del atlas de
 * `/panel`) y, como él, local al módulo `agenda`: su geometría es la de esta
 * pantalla, no se promueve a shared/ui.
 *
 * El texto de espera nunca desaparece: se sigue anunciando vía
 * role="status"/aria-live="polite" aunque la representación visual sea
 * geometría. `prefers-reduced-motion: reduce` deja los bloques en opacidad
 * fija, sin pulso.
 */
interface Props {
  /** Texto exacto ya usado por el código ("Cargando barberos…"): este
   * componente no lo redacta. */
  label: string
}

defineProps<Props>()

// Cada banda declara cuántos campos lleva cada una de sus filas, calcado del
// esqueleto del atlas (`skeletonHtml` en tools/mockups/nuevo-turno-eventos/
// render.mjs): 1 → dos campos, 2 → dos filas de dos, 3 → dos campos,
// 4 → un campo ancho.
const BANDS: readonly (readonly number[])[] = [[2], [2, 2], [2], [1]]
</script>

<template>
  <div class="nt-skeleton" role="status" aria-live="polite">
    <p class="nt-skeleton__status">
      <span class="nt-skeleton__spinner" aria-hidden="true" />{{ label }}
    </p>

    <div class="nt-skeleton__layout" aria-hidden="true">
      <div class="nt-skeleton__sheet">
        <div v-for="(rows, band) in BANDS" :key="band" class="nt-skeleton__band">
          <div class="nt-skeleton__head">
            <span class="nt-skeleton__bar nt-skeleton__bar--num" />
            <span class="nt-skeleton__stack">
              <span class="nt-skeleton__bar nt-skeleton__bar--title" />
              <span class="nt-skeleton__bar nt-skeleton__bar--hint" />
            </span>
          </div>
          <div class="nt-skeleton__body">
            <div v-for="(count, row) in rows" :key="row" class="nt-skeleton__row">
              <span v-for="n in count" :key="n" class="nt-skeleton__field">
                <span class="nt-skeleton__bar nt-skeleton__bar--label" />
                <span class="nt-skeleton__bar nt-skeleton__bar--control" />
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="nt-skeleton__actions">
        <span class="nt-skeleton__bar nt-skeleton__bar--cta" />
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Móvil primero (420px del atlas); la banda de escritorio del atlas
   (1440px) entra en @media (min-width: 1024px), igual que DailyAgendaPage. */
.nt-skeleton {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.nt-skeleton__status {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.14em;
  line-height: 1;
  text-transform: uppercase;
  color: var(--color-brand-accent-surface);
}

.nt-skeleton__spinner {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
  border: 2px solid rgb(184 149 90 / 22%);
  border-top-color: var(--color-brand-accent-surface);
  border-right-color: var(--color-brand-accent-surface);
  border-radius: 50%;
  transform: rotate(-38deg);
}

.nt-skeleton__layout {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* Misma hoja que el formulario real: una sola superficie translúcida con
   filete de latón a la izquierda, no cuatro tarjetas. */
.nt-skeleton__sheet {
  background-color: rgb(244 240 231 / 3.5%);
  border: var(--border-width-normal) solid rgb(244 240 231 / 10%);
  border-left: 3px solid rgb(184 149 90 / 55%);
  border-radius: 2px;
}

.nt-skeleton__band {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 18px 16px;
}

.nt-skeleton__band + .nt-skeleton__band {
  border-top: var(--border-width-normal) solid rgb(244 240 231 / 10%);
}

.nt-skeleton__head {
  display: flex;
  align-items: flex-start;
  gap: 11px;
}

.nt-skeleton__stack {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.nt-skeleton__body {
  display: flex;
  flex-direction: column;
  gap: 13px;
}

.nt-skeleton__row {
  display: flex;
  flex-direction: column;
  gap: 13px;
}

.nt-skeleton__field {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
}

.nt-skeleton__bar {
  display: block;
  background-color: rgb(244 240 231 / 14%);
  border-radius: 2px;
  animation: nt-skeleton-pulse 1400ms ease-in-out infinite;
}

.nt-skeleton__bar--num {
  flex: none;
  width: 26px;
  height: 26px;
}

.nt-skeleton__bar--title {
  width: 130px;
  height: 14px;
}

.nt-skeleton__bar--hint {
  width: 170px;
  height: 10px;
  margin-top: 8px;
  background-color: rgb(244 240 231 / 9%);
}

.nt-skeleton__bar--label {
  width: 80px;
  height: 9px;
}

/* El campo por llegar conserva su filete inferior de 2px: la espera anticipa
   la forma del control reglado, no un rectángulo neutro. */
.nt-skeleton__bar--control {
  width: 100%;
  height: 46px;
  background-color: rgb(244 240 231 / 7%);
  border-bottom: var(--border-width-emphasis) solid rgb(184 149 90 / 45%);
}

.nt-skeleton__actions {
  display: flex;
  flex-direction: column;
}

.nt-skeleton__bar--cta {
  width: 100%;
  height: 48px;
  background-color: rgb(184 149 90 / 35%);
}

@keyframes nt-skeleton-pulse {
  0%,
  100% {
    opacity: 0.6;
  }
  50% {
    opacity: 1;
  }
}

@media (prefers-reduced-motion: reduce) {
  .nt-skeleton__bar {
    animation: none;
    opacity: 0.8;
  }
}

@media (min-width: 1024px) {
  .nt-skeleton {
    gap: 18px;
  }

  .nt-skeleton__status {
    gap: 12px;
    font-size: 12px;
  }

  .nt-skeleton__spinner {
    width: 18px;
    height: 18px;
  }

  /* Misma sangría que el CTA real: rótulo (288px) + gutter (28px) + relleno
     de banda (26px) + filete de latón (3px). */
  .nt-skeleton__layout {
    gap: 0;
    max-width: 980px;
  }

  .nt-skeleton__actions {
    margin-top: 16px;
    padding-left: calc(26px + 288px + 28px + 3px);
  }

  .nt-skeleton__band {
    display: grid;
    grid-template-columns: 288px minmax(0, 1fr);
    gap: 28px;
    padding: 20px 26px;
  }

  .nt-skeleton__head {
    gap: 12px;
  }

  .nt-skeleton__body {
    gap: 14px;
  }

  .nt-skeleton__row {
    flex-direction: row;
    gap: 20px;
  }

  .nt-skeleton__field {
    gap: 9px;
  }

  .nt-skeleton__bar--num {
    width: 28px;
    height: 28px;
  }

  .nt-skeleton__bar--title {
    width: 150px;
    height: 15px;
  }

  .nt-skeleton__bar--hint {
    width: 190px;
    height: 10px;
    margin-top: 9px;
  }

  .nt-skeleton__bar--label {
    width: 88px;
  }

  .nt-skeleton__bar--cta {
    width: 240px;
    height: 50px;
  }
}
</style>
