// Módulo: asignación de servicios a barberos (HU-023). API pública mínima
// para `app`, según docs/03-desarrollo/estandar-frontend-vue.md: la ruta
// hija privada (cargada de forma diferida) y la entrada de navegación que la
// representa. `app/router/index.ts` combina esto con las rutas de
// `auth`/`settings`/`staff`/`catalog` (y de cualquier otro módulo con
// pantalla privada) antes de pasarlas a `auth.privateShellRoute`; este
// módulo no monta su propio cascarón ni importa internos de `staff` ni de
// `catalog` (colabora con ambos SOLO a través del cliente HTTP compartido,
// nunca de sus clientes API privados). Página, modelo y cliente API internos
// permanecen privados.
import type { NavItem } from '@/shared/navigation/navItem'

export { barberServicesPrivateShellChildRoutes } from './routes'

export const barberServicesNavItems: NavItem[] = [
  { to: { name: 'barber-services' }, label: 'Servicios por barbero' },
]
