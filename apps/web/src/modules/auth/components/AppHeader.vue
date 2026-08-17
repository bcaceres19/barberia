<script setup lang="ts">
// Cabecera del cascarón privado (HU-012, CA-012-04): muestra siempre la
// barbería activa y permite cerrar sesión. El nombre viene únicamente de la
// fuente aprobada por DEC-060 (prop obligatoria, nunca un valor local ni un
// texto fijo).
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { BaseButton } from '@/shared/ui'
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
    <p class="app-header__barbershop" data-testid="barbershop-name">{{ barbershopName }}</p>
    <BaseButton variant="secondary" :disabled="loggingOut" :loading="loggingOut" @click="onLogout">
      Cerrar sesión
    </BaseButton>
  </header>
</template>

<style scoped>
.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-3) var(--space-4);
  border-bottom: var(--border-width-normal) solid var(--color-border-subtle);
  background-color: var(--color-surface);
}

.app-header__barbershop {
  margin: 0;
  font-size: var(--font-size-body);
  font-weight: 600;
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
