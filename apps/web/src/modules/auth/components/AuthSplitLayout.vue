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
  </main>
</template>

<style scoped>
.auth-split {
  display: grid;
  grid-template-rows: auto 1fr;
  min-height: 100dvh;
}

.auth-split__brand {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: var(--space-2);
  padding: var(--space-6) var(--space-6);
  background-color: var(--color-surface-strong);
}

.auth-split__rule {
  display: none;
  width: 64px;
  height: 1px;
  margin: var(--space-3) 0;
  background-color: var(--color-brand-accent-surface);
  position: relative;
}

.auth-split__rule::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 28px;
  width: 6px;
  height: 6px;
  transform: translateY(-50%) rotate(45deg);
  background-color: var(--color-brand-accent-surface);
}

.auth-split__tagline {
  display: none;
  margin: 0;
  max-width: 32ch;
  font-family: var(--font-sans);
  font-size: var(--font-size-body-lg);
  line-height: var(--font-size-body-lg-line);
  color: var(--color-on-strong);
  opacity: 0.8;
}

.auth-split__content {
  display: flex;
  /* En móvil el contenido arranca justo debajo de la franja de marca en
     vez de centrarse en el espacio restante: centrarlo dejaba un vacío
     grande entre la franja y el formulario en viewports altos (320-768,
     confirmado con evidencia Playwright). En escritorio sí se centra
     respecto a la altura completa, igual que el panel de marca. */
  align-items: flex-start;
  justify-content: center;
  padding: var(--space-8) var(--space-4);
}

.auth-split__card {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  width: 100%;
  max-width: 360px;
}

@media (min-width: 1024px) {
  .auth-split {
    grid-template-rows: none;
    grid-template-columns: 1fr 1fr;
  }

  .auth-split__brand {
    padding: var(--space-12);
  }

  .auth-split__rule {
    display: block;
  }

  .auth-split__tagline {
    display: block;
  }

  .auth-split__content {
    align-items: center;
  }
}
</style>
