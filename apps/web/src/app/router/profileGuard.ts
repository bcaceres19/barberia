// Perfil de barbero individual (DEC-115): «Servicios por barbero» deja de ser un
// destino porque con un solo barbero cada servicio se ofrece desde «Servicios».
// La ruta sigue existiendo para el panel completo; aquí solo se desvía a quien
// trabaja solo y entra por un enlace guardado, para que nunca vea una pantalla
// que no es suya ni una en blanco.
import { watch } from 'vue'
import type { RouteLocationNormalized, RouteLocationRaw, Router } from 'vue-router'
import { panelProfile, type PanelProfile } from '@/shared/model'

const RETIRED_ROUTES: Record<PanelProfile, Record<string, string>> = {
  shop: {},
  solo: { 'barber-services': 'catalog-servicios' },
}

/** Destino alternativo para `to` en `profile`, o `null` si la ruta sigue disponible. */
export function profileRedirect(
  to: Pick<RouteLocationNormalized, 'name'>,
  profile: PanelProfile,
): RouteLocationRaw | null {
  if (typeof to.name !== 'string') return null
  const target = RETIRED_ROUTES[profile][to.name]
  return target ? { name: target } : null
}

/**
 * Aplica `profileRedirect` en cada navegación y también cuando el perfil llega
 * DESPUÉS: en una primera visita sin copia local la marca viaja con la sesión y
 * todavía es `shop` al evaluar la ruta, así que sin esta segunda vía un enlace
 * guardado a una ruta retirada se quedaría en ella. La segunda vía espera a que
 * termine la navegación inicial (`isReady`): la marca suele llegar mientras esa
 * navegación aún espera la sesión, y entonces no hay ruta actual que corregir.
 */
export function installProfileRedirect(router: Router): void {
  router.beforeEach((to) => profileRedirect(to, panelProfile.value) ?? true)

  const reconcile = () => {
    const target = profileRedirect(router.currentRoute.value, panelProfile.value)
    if (target) void router.replace(target)
  }
  watch(panelProfile, () => {
    void router.isReady().then(reconcile)
  })
}
