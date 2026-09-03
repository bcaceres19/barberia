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
import { NavaWordmark } from '@/shared/ui'

interface Props {
  tagline?: string
  /** Leyenda reglada del panel de marca / pie móvil: "ACCESO SEGURO" o
   * "RECUPERACIÓN SEGURA" (contrato visual, auth-eventos/README.md). */
  caption: string
}

withDefaults(defineProps<Props>(), {
  tagline: 'Gestión precisa para tu barbería.',
})
</script>

<template>
  <main class="auth-split">
    <div class="auth-split__frame">
      <div class="auth-split__brand" aria-hidden="true">
        <div class="auth-split__wordmark-group">
          <NavaWordmark variant="inverted" size="lg" />
          <span class="auth-split__rule"></span>
        </div>
        <p class="auth-split__tagline">{{ tagline }}</p>
        <p class="auth-split__caption">{{ caption }} · NAVA</p>
      </div>
      <div class="auth-split__content">
        <div class="auth-split__card">
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
  place-items: center;
  min-height: 100dvh;
  background-color: var(--color-canvas);
}

.auth-split__frame {
  display: grid;
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
  max-width: 470px;
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
</style>
