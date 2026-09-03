// Módulo: gestión de los barberos de la barbería (HU-021). API pública
// mínima para `app`, según docs/03-desarrollo/estandar-frontend-vue.md: la
// ruta hija privada (cargada de forma diferida) y la entrada de navegación
// que la representa. `app/router/index.ts` combina esto con las rutas de
// `auth`/`settings` (y de cualquier otro módulo con pantalla privada) antes
// de pasarlas a `auth.privateShellRoute`; este módulo no monta su propio
// cascarón ni importa internos de otro módulo. Formulario, validación y
// cliente API internos permanecen privados.
import type { NavItem } from '@/shared/navigation/navItem'

export { staffPrivateShellChildRoutes } from './routes'

export const staffNavItems: NavItem[] = [
  { to: { name: 'staff-barberos' }, label: 'Barberos', primary: true },
]
