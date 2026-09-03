<script setup lang="ts">
// Cascarón compartido de las pantallas públicas de acceso/recuperación
// (Fase 3 del rediseño Tailored Grid, issue #188): panel de marca en tinta
// (editorial, decorativo) junto al formulario real sobre marfil. En
// escritorio (>= 1024px, mismo punto de corte que el resto de NAVA) el
// panel ocupa la mitad izquierda; en móvil se condensa a una franja
// superior con solo la marca, sin repetir el wordmark dentro de la
// tarjeta del formulario. `LoginPage`/`RecoveryPage` conservan su propio
// estado y encabezado real (`<h1>`); este componente no es un landmark de
// contenido, por eso su panel de marca lleva `aria-hidden`.
import { NavaWordmark } from '@/shared/ui'

interface Props {
  tagline?: string
}

withDefaults(defineProps<Props>(), {
  tagline: 'Gestión precisa para tu barbería.',
})
</script>

<template>
  <main class="auth-split">
    <div class="auth-split__frame">
      <div class="auth-split__brand" aria-hidden="true">
        <NavaWordmark variant="inverted" size="lg" />
        <span class="auth-split__rule"></span>
        <p class="auth-split__tagline">{{ tagline }}</p>
      </div>
      <div class="auth-split__content">
        <div class="auth-split__card">
          <slot />
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
.auth-split {
  --color-canvas: #f5f1ec;
  --color-surface: #ffffff;
  --color-surface-strong: #03182e;
  --color-text-primary: #2a2d32;
  --color-border-subtle: #c9c0b2;
  --color-action-primary: #03182e;
  --color-action-primary-hover: #072139;
  --color-action-primary-active: #021224;
  --color-on-strong: #f5f1ec;
  --color-brand-accent-surface: #b8955a;

  display: grid;
  place-items: center;
  min-height: 100dvh;
  background-color: #f3efeb;
}

.auth-split__frame {
  display: grid;
  grid-template-rows: 88px minmax(0, 1fr);
  width: 100%;
  min-height: 100dvh;
  background-color: var(--color-canvas);
}

.auth-split__brand {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-5) var(--space-6);
  background-color: var(--color-surface-strong);
  border-bottom: var(--border-width-normal) solid var(--color-action-soft-border);
  text-align: center;
}

.auth-split__brand :deep(.nava-wordmark--lg) {
  font-size: 44px;
  line-height: 48px;
  letter-spacing: 0.035em;
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
  display: none;
  margin: 0;
  max-width: 15ch;
  font-family: var(--font-display);
  font-size: 36px;
  line-height: 42px;
  color: var(--color-on-strong);
  opacity: 0.92;
}

.auth-split__content {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 48px clamp(var(--space-5), 6vw, var(--space-10)) var(--space-12);
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
    padding: var(--space-8);
  }

  .auth-split__frame {
    grid-template-rows: none;
    grid-template-columns: minmax(390px, 46fr) minmax(480px, 54fr);
    width: min(1440px, 100%);
    min-height: min(960px, calc(100dvh - 64px));
    overflow: hidden;
    border: var(--border-width-normal) solid var(--color-border-subtle);
    border-radius: 10px;
    box-shadow: var(--shadow-dialog);
  }

  .auth-split__brand {
    --auth-wordmark-font-size: clamp(120px, 11.5vw, 168px);

    padding: var(--space-16) var(--space-10);
    border-right: var(--border-width-normal) solid var(--color-action-soft-border);
    border-bottom: 0;
  }

  .auth-split__brand :deep(.nava-wordmark--lg) {
    font-size: var(--auth-wordmark-font-size);
    line-height: 1;
    letter-spacing: 0.025em;
  }

  .auth-split__rule {
    display: block;
    width: calc(var(--auth-wordmark-font-size) * 2.875);
  }

  .auth-split__tagline {
    display: block;
  }

  .auth-split__content {
    align-items: center;
    padding: var(--space-16) clamp(48px, 5vw, 64px);
  }

  .auth-split__card {
    max-width: 620px;
  }

  .auth-split__card :deep(.base-alert) {
    padding: 18px 20px;
    font-size: 17px;
    line-height: 24px;
  }

  .auth-split__card :deep(.base-alert__icon) {
    width: 24px;
    height: 24px;
  }
}

@media (max-width: 359px) {
  .auth-split__content {
    padding-right: var(--space-4);
    padding-left: var(--space-4);
  }
}
</style>
