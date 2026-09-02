<script setup lang="ts">
// Cabecera del cascarón privado (HU-012, CA-012-04): muestra siempre la
// barbería activa y permite cerrar sesión. El nombre viene únicamente de la
// fuente aprobada por DEC-060 (prop obligatoria, nunca un valor local ni un
// texto fijo). Recompuesta a la identidad NAVA en la Fase 2
// (estandar-diseno-visual.md §5.3): wordmark + barbería a la izquierda,
// "Nuevo turno" como acción primaria global (no un destino del dock,
// especificacion-frontend-nava.md §5.1) y cierre de sesión conservado tal
// cual (misma lógica y pruebas, sin convertirlo en un ítem del dock).
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { BaseButton, NavaWordmark } from '@/shared/ui'
import { httpClient } from '@/shared/api/httpClient'
import { resetForFreshLogin } from '../model/sessionStore'

defineProps<{
  barbershopName: string
}>()

const router = useRouter()
const loggingOut = ref(false)

async function onLogout() {
  // Guardia de doble envío, mismo patrón que LoginPage (CA-010-04): la
  // asignación síncrona bloquea un segundo clic mientras la solicitud está
  // en curso.
  if (loggingOut.value) return
  loggingOut.value = true

  try {
    await httpClient.POST('/private/auth/logout')
  } catch {
    // Un fallo de red ambiguo no debe fingir éxito ni bloquear el cierre
    // local: la sesión igual se trata como terminada del lado del cliente,
    // porque no hay una forma segura de reintentar automáticamente una
    // acción destructiva sin saber si el servidor ya la aplicó.
  } finally {
    // Limpia el estado local no sensible (nunca hubo token en JS) y navega
    // a acceso; una próxima entrada a una ruta privada vuelve a consultar
    // el servidor en vez de reutilizar este resultado (mismo mecanismo que
    // tras un login exitoso).
    resetForFreshLogin()
    loggingOut.value = false
    await router.push({ name: 'acceso' })
  }
}
</script>

<template>
  <header class="app-header">
    <div class="app-header__identity">
      <NavaWordmark />
      <span class="app-header__divider" aria-hidden="true">·</span>
      <p class="app-header__barbershop" data-testid="barbershop-name">{{ barbershopName }}</p>
    </div>
    <div class="app-header__actions">
      <RouterLink :to="{ name: 'agenda-nuevo-turno' }" class="app-header__cta">
        Nuevo turno
      </RouterLink>
      <BaseButton
        variant="secondary"
        :disabled="loggingOut"
        :loading="loggingOut"
        @click="onLogout"
      >
        Cerrar sesión
      </BaseButton>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  border-bottom: var(--border-width-normal) solid var(--color-border-subtle);
  background-color: var(--color-surface);
}

.app-header__identity {
  display: flex;
  min-width: 0;
  align-items: baseline;
  gap: var(--space-2);
}

.app-header__divider {
  color: var(--color-border-control);
}

.app-header__barbershop {
  margin: 0;
  overflow: hidden;
  font-size: var(--font-size-body);
  font-weight: 600;
  color: var(--color-text-primary);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-header__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2);
}

.app-header__cta {
  display: inline-flex;
  height: var(--control-height);
  align-items: center;
  padding: 0 var(--space-4);
  border-radius: var(--radius-sm);
  background-color: var(--color-action-primary);
  color: var(--color-on-strong);
  font-family: var(--font-sans);
  font-size: var(--font-size-body);
  font-weight: 500;
  text-decoration: none;
  transition: background-color var(--motion-duration-fast) var(--motion-easing-standard);
}

@media (prefers-reduced-motion: reduce) {
  .app-header__cta {
    transition: none;
  }
}

.app-header__cta:hover {
  background-color: var(--color-action-primary-hover);
}

.app-header__cta:active {
  background-color: var(--color-action-primary-active);
}

.app-header__cta:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-surface),
    0 0 0 4px var(--color-focus);
}
</style>
