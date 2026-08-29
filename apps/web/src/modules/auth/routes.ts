// Rutas del módulo auth, cargadas de forma diferida (docs/03-desarrollo/
// estandar-frontend-vue.md §7.8) y registradas desde `app/router`
// (trabajo requerido §1). Este archivo es la API pública de rutas del
// módulo: `app/router/index.ts` las importa y las compone, sin conocer
// los componentes internos de `auth`.
import type { RouteRecordRaw } from 'vue-router'
import type { NavItem } from '@/shared/navigation/navItem'
import { requireSession } from './guards/requireSession'

export const authRoutes: RouteRecordRaw[] = [
  {
    path: '/acceso',
    name: 'acceso',
    component: () => import('./pages/LoginPage.vue'),
  },
  {
    // Destino real del enlace de recuperación de CA-010-08 (antes
    // provisional bajo `DP-UX-06`, resuelta por `HU-011`).
    path: '/recuperar-acceso',
    name: 'recuperar-acceso',
    component: () => import('./pages/RecoveryPage.vue'),
  },
]

// Hijas privadas propias de `auth`: ninguna desde HU-062. `/panel` (path
// vacío, name 'panel') lo registra ahora `agenda` (mismo criterio de "cada
// módulo contribuye sus propias hijas" que ya usan `catalog`/`schedules`),
// porque abrir el panel privado ahora muestra la agenda diaria real en vez
// del marcador de posición de HU-012 (`auth/pages/PanelPage.vue`, retirado).
// Se conserva el arreglo vacío para que `app/router/index.ts` siga
// componiendo con el mismo patrón `...privateShellChildRoutes` sin un caso
// especial si `auth` vuelve a necesitar una hija propia más adelante.
export const privateShellChildRoutes: RouteRecordRaw[] = []

// Fábrica del único cascarón privado (HU-012, DEC-056; composición
// multi-módulo desde HU-020): un único guard generalizado, montado una
// sola vez sobre `/panel`, cubre toda ruta privada hija, sin repetirlo por
// pantalla ni crear un segundo cascarón. `app/router/index.ts` es el único
// llamador: le pasa `children` y `extraNavItems` ya combinadas desde la
// API pública de cada módulo que necesite una pantalla privada, sin que
// `auth` importe internos de ningún otro módulo (app → modules → shared).
// `extraNavItems` viaja como prop ESTÁTICA de la ruta (no depende de
// params): PrivateShell la reenvía a AppNav sin que ninguno de los dos
// conozca qué módulo la originó.
export function privateShellRoute(
  children: RouteRecordRaw[],
  extraNavItems: NavItem[] = [],
): RouteRecordRaw {
  return {
    path: '/panel',
    component: () => import('./layouts/PrivateShell.vue'),
    beforeEnter: requireSession,
    props: () => ({ extraNavItems }),
    children,
  }
}
