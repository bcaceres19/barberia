<script setup lang="ts">
// Cascarón compartido de las pantallas públicas de acceso/recuperación
// (Fase 3 del rediseño Tailored Grid, issue #188; contrato reglado del
// issue #212/#213). En escritorio (>= 1024px, mismo punto de corte que el
// resto de NAVA) un panel de marca en tinta (editorial, decorativo) ocupa
// la mitad izquierda; en móvil ese panel se oculta por completo y el
// lienzo entero queda en marfil, con la marca reglada (wordmark + regla de
// latón a su ancho) como encabezado de la propia columna de contenido —
// misma marca que en escritorio, ya no una franja de tinta aparte.
// `LoginPage`/`RecoveryPage` conservan su propio estado y encabezado real
// (`<h1>`); ni el panel de marca ni la marca reglada del encabezado son un
// landmark de contenido, por eso ambos llevan `aria-hidden`.
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { NavaWordmark } from '@/shared/ui'
import AuthBrandBackdrop from './AuthBrandBackdrop.vue'

interface Props {
  tagline?: string
  /** Leyenda reglada del panel de marca / pie móvil: "ACCESO SEGURO" o
   * "RECUPERACIÓN SEGURA" (contrato visual, auth-eventos/README.md). */
  caption: string
}

withDefaults(defineProps<Props>(), {
  tagline: 'Gestión precisa para tu barbería.',
})

// La entrada escalonada de la tarjeta solo aplica a lo que ya está montado
// al abrir la pantalla. Se retira al terminar para que una alerta o el reto
// telefónico que aparezcan después usen su propia transición corta (estándar
// visual §8: 160 ms, sin retrasos no esenciales) y no el escalón de la
// entrada. El estado final de cada elemento es el normal, así que quitar la
// clase no produce salto alguno.
const introActive = ref(true)
let introTimer: ReturnType<typeof setTimeout> | undefined
onMounted(() => {
  introTimer = setTimeout(() => {
    introActive.value = false
  }, 1600)
})
onBeforeUnmount(() => clearTimeout(introTimer))
</script>

<template>
  <main class="auth-split">
    <div class="auth-split__frame">
      <div class="auth-split__brand" aria-hidden="true">
        <AuthBrandBackdrop />
        <div class="auth-split__wordmark-group">
          <NavaWordmark variant="inverted" size="lg" />
          <span class="auth-split__rule"></span>
        </div>
        <p class="auth-split__tagline">{{ tagline }}</p>
        <p class="auth-split__caption">{{ caption }} · NAVA</p>
      </div>
      <div class="auth-split__content">
        <div class="auth-split__card" :class="{ 'auth-split__card--intro': introActive }">
          <div class="auth-split__mark" aria-hidden="true">
            <NavaWordmark variant="ink" size="lg" />
            <span class="auth-split__mark-rule"></span>
          </div>
          <slot />
          <p class="auth-split__mobile-caption" aria-hidden="true">{{ caption }} · NAVA</p>
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
.auth-split {
  display: grid;
  /* Sin esta pista explícita, la columna implícita de este grid se
     dimensiona a su contenido (min-content/max-content) en vez de al
     ancho disponible: el reto telefónico y sus seis ranuras de OTP tienen
     contenido intrínseco más ancho que 420 px, y el grid se desborda en
     vez de encogerse (issue #213, evidencia de auth-eventos-fidelidad). */
  grid-template-columns: minmax(0, 1fr);
  place-items: center;
  min-height: 100dvh;
  background-color: var(--color-canvas);
}

.auth-split__frame {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  grid-template-rows: auto minmax(0, 1fr);
  width: 100%;
  min-height: 100dvh;
  background-color: var(--color-canvas);
}

/* Oculto por completo en móvil (< 1024px): el mockup mobile no repite una
   franja de tinta, la marca reglada de `.auth-split__mark` ya cumple ese
   rol dentro de la columna de contenido. */
.auth-split__brand {
  display: none;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-5) var(--space-6);
  background-color: var(--color-surface-strong);
  border-bottom: var(--border-width-normal) solid var(--color-action-soft-border);
  text-align: center;
  position: relative;
  overflow: hidden;
}

/* El contenido real del panel queda por encima del fondo ornamental. */
.auth-split__wordmark-group,
.auth-split__tagline,
.auth-split__caption {
  position: relative;
  z-index: 1;
}

.auth-split__brand :deep(.nava-wordmark--lg) {
  font-size: 76px;
  line-height: 1;
  letter-spacing: 0.03em;
}

.auth-split__wordmark-group {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.auth-split__rule {
  display: none;
  width: 220px;
  height: 2px;
  margin: var(--space-6) 0 var(--space-5);
  background-color: var(--color-brand-accent-surface);
  position: relative;
}

.auth-split__rule::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 9px;
  height: 9px;
  transform: translateY(-50%) rotate(45deg);
  background-color: var(--color-brand-accent-surface);
  box-shadow: 0 0 0 8px var(--color-surface-strong);
}

.auth-split__tagline {
  margin: 0;
  max-width: 15ch;
  font-family: var(--font-display);
  font-size: 36px;
  line-height: 42px;
  color: var(--color-on-strong);
  opacity: 0.92;
}

/* Leyenda reglada del panel de marca ("ACCESO SEGURO · NAVA" /
   "RECUPERACIÓN SEGURA · NAVA"): versalitas de baja énfasis al pie del
   panel de tinta, mismo tratamiento tipográfico que el resto de las
   etiquetas regladas del contrato visual. */
.auth-split__caption {
  margin: var(--space-10) 0 0;
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--color-on-strong);
  opacity: 0.56;
}

/* Marca reglada de la columna de contenido: wordmark a escala de chip
   editorial + regla de latón a su propio ancho, encabezado compartido por
   `/acceso` y `/recuperar-acceso` en los dos viewports (issue #213,
   auth-eventos/README.md "Ambas rutas muestran el wordmark con la regla
   de latón a su ancho completo"). */
.auth-split__mark {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: fit-content;
  margin: 0 auto 48px;
}

.auth-split__mark :deep(.nava-wordmark--lg) {
  font-size: 80px;
  line-height: 1;
  letter-spacing: 0.03em;
}

.auth-split__mark-rule {
  width: 100%;
  height: 2px;
  margin-top: var(--space-4);
  background-color: var(--color-accent-brass);
}

/* Leyenda reglada al pie de la tarjeta, solo en móvil (el panel de tinta
   la reemplaza en escritorio). */
.auth-split__mobile-caption {
  margin: var(--space-8) 0 0;
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  text-align: center;
  color: var(--color-text-secondary);
}

.auth-split__content {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 88px clamp(var(--space-5), 6vw, var(--space-10)) var(--space-12);
  background-color: var(--color-canvas);
}

.auth-split__card {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
  width: 100%;
  min-width: 0;
  max-width: 470px;
}

/* Entrada del panel de marca: la marca se descubre de izquierda a derecha,
   la regla se abre desde el rombo, el lema y la leyenda suben con suavidad.
   Cada animación termina en el estado normal del elemento (`backwards`), así
   que sin animación —reduced motion— todo queda ya compuesto. */
.auth-split__wordmark-group :deep(.nava-wordmark) {
  animation: auth-reveal-wipe 1.1s cubic-bezier(0.65, 0, 0.2, 1) 0.15s backwards;
}

.auth-split__rule {
  animation: auth-rule-open 1s cubic-bezier(0.65, 0, 0.2, 1) 0.9s backwards;
}

.auth-split__rule::after {
  animation: auth-diamond-in 0.9s cubic-bezier(0.3, 1.4, 0.5, 1) 1.1s backwards;
}

.auth-split__tagline {
  animation: auth-rise 0.9s cubic-bezier(0.2, 0.7, 0.2, 1) 1.2s backwards;
}

.auth-split__caption {
  animation: auth-rise 0.9s cubic-bezier(0.2, 0.7, 0.2, 1) 1.6s backwards;
}

/* Entrada de la tarjeta: cada bloque sube 12 px y aparece, en cascada corta
   (marca, encabezado, formulario, resto). */
.auth-split__card--intro > * {
  animation: auth-rise 0.7s cubic-bezier(0.2, 0.7, 0.2, 1) backwards;
}

.auth-split__card--intro > :nth-child(1) {
  animation-delay: 0.1s;
}

.auth-split__card--intro > :nth-child(2) {
  animation-delay: 0.22s;
}

.auth-split__card--intro > :nth-child(3) {
  animation-delay: 0.34s;
}

.auth-split__card--intro > :nth-child(n + 4) {
  animation-delay: 0.46s;
}

/* La regla de latón bajo la marca de la tarjeta se traza desde el centro. */
.auth-split__card--intro .auth-split__mark-rule {
  transform-origin: center;
  animation: auth-rule-grow 0.9s cubic-bezier(0.65, 0, 0.2, 1) 0.45s backwards;
}

/* Mesa de patronaje también en el lienzo claro: una rejilla casi invisible
   que se desvanece desde la esquina superior y una cinta métrica de latón
   en el borde superior. Pura ornamentación, detrás del contenido. */
.auth-split__content {
  position: relative;
  isolation: isolate;
}

.auth-split__content::before,
.auth-split__content::after {
  content: '';
  position: absolute;
  z-index: -1;
  pointer-events: none;
}

.auth-split__content::before {
  inset: 0;
  background-image:
    linear-gradient(
      to right,
      color-mix(in srgb, var(--color-action-primary) 5%, transparent) 1px,
      transparent 1px
    ),
    linear-gradient(
      to bottom,
      color-mix(in srgb, var(--color-action-primary) 5%, transparent) 1px,
      transparent 1px
    );
  background-size: 56px 56px;
  mask-image: radial-gradient(ellipse 80% 70% at 100% 0%, #000 0%, transparent 70%);
  animation: auth-fade 1.8s ease-out 0.3s backwards;
}

.auth-split__content::after {
  top: 0;
  right: 0;
  left: 0;
  height: 14px;
  background-image:
    linear-gradient(to right, var(--color-accent-brass) 1px, transparent 1px),
    linear-gradient(to right, var(--color-accent-brass) 1px, transparent 1px);
  background-size:
    12px 7px,
    60px 14px;
  background-repeat: repeat-x;
  opacity: 0.3;
  mask-image: linear-gradient(to right, transparent, #000 18%, #000 82%, transparent);
  animation: auth-fade 1.6s ease-out 0.6s backwards;
}

@keyframes auth-reveal-wipe {
  from {
    clip-path: inset(0 100% 0 0);
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    clip-path: inset(0 0 0 0);
  }
}

@keyframes auth-rule-open {
  from {
    clip-path: inset(-20px 50% -20px 50%);
  }
  to {
    clip-path: inset(-20px 0 -20px 0);
  }
}

@keyframes auth-rule-grow {
  from {
    transform: scaleX(0);
  }
}

@keyframes auth-diamond-in {
  from {
    opacity: 0;
    transform: translateY(-50%) rotate(-135deg) scale(0);
  }
}

@keyframes auth-rise {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
}

@keyframes auth-fade {
  from {
    opacity: 0;
  }
}

@keyframes auth-seam-glint {
  from {
    background-position: 0 100%;
  }
  to {
    background-position: 0 0;
  }
}

@media (min-width: 1024px) {
  .auth-split {
    padding: 0;
  }

  .auth-split__frame {
    grid-template-rows: none;
    grid-template-columns: 43.0556fr 56.9444fr;
    width: 100%;
    min-height: 100dvh;
  }

  .auth-split__brand {
    --auth-wordmark-font-size: clamp(120px, 9vw, 132px);

    display: flex;
    padding: var(--space-16) var(--space-10);
    border-right: var(--border-width-normal) solid var(--color-action-soft-border);
    border-bottom: 0;
  }

  /* Destello que recorre la costura entre los dos paneles, de arriba abajo. */
  .auth-split__brand::after {
    content: '';
    position: absolute;
    top: 0;
    right: 0;
    bottom: 0;
    width: 1px;
    z-index: 1;
    background-image: linear-gradient(
      to bottom,
      transparent 0 42%,
      var(--color-brand-accent-surface) 50%,
      transparent 58% 100%
    );
    background-size: 100% 300%;
    animation: auth-seam-glint 7s ease-in-out 2s infinite;
  }

  .auth-split__brand :deep(.nava-wordmark--lg) {
    font-size: var(--auth-wordmark-font-size);
    line-height: 1;
    letter-spacing: 0.025em;
  }

  .auth-split__tagline {
    font-size: 44px;
    line-height: 56px;
  }

  .auth-split__caption {
    margin-top: 64px;
  }

  .auth-split__wordmark-group {
    width: fit-content;
  }

  .auth-split__rule {
    display: block;
    width: 100%;
    margin: 48px 0 56px;
    background-color: var(--color-accent-brass);
  }

  .auth-split__mobile-caption {
    display: none;
  }

  .auth-split__content {
    align-items: center;
    padding: var(--space-16) clamp(48px, 5vw, 64px);
  }

  .auth-split__card {
    max-width: 580px;
  }

  .auth-split__card :deep(.base-alert) {
    padding: 18px 20px;
    font-size: 17px;
    line-height: 24px;
  }

  .auth-split__mark :deep(.nava-wordmark--lg) {
    font-size: 56px;
  }

  .auth-split__mark {
    margin-bottom: var(--space-8);
  }
}

@media (max-width: 359px) {
  .auth-split__content {
    padding-right: var(--space-4);
    padding-left: var(--space-4);
  }
}

@media (prefers-reduced-motion: reduce) {
  .auth-split__wordmark-group :deep(.nava-wordmark),
  .auth-split__rule,
  .auth-split__rule::after,
  .auth-split__tagline,
  .auth-split__caption,
  .auth-split__card--intro > *,
  .auth-split__card--intro .auth-split__mark-rule,
  .auth-split__content::before,
  .auth-split__content::after,
  .auth-split__brand::after {
    animation: none;
  }
}
</style>
