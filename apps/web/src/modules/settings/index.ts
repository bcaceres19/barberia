// Módulo: configuración básica de la barbería (HU-020). API pública mínima
// para `app`, según docs/03-desarrollo/estandar-frontend-vue.md: la ruta
// hija privada (cargada de forma diferida) y la entrada de navegación que
// la representa. `app/router/index.ts` combina ambas con las de `auth`
// (y de cualquier otro módulo con pantalla privada) antes de pasarlas a
// `auth.privateShellRoute`; este módulo no monta su propio cascarón ni
// importa internos de `auth`. Formulario, validación y cliente API
// internos permanecen privados.
import type { NavItem } from '@/shared/navigation/navItem'

export { settingsPrivateShellChildRoutes } from './routes'

export const settingsNavItems: NavItem[] = [
  { to: { name: 'configuracion-barberia' }, label: 'Barbería' },
]
