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
// "Servicios por barbero" no es uno de los 5 destinos P0 del dock NAVA
// (especificacion-frontend-nava.md §5.1 agrupa "catálogo y asignación de
// servicios por barbero" bajo un único destino "Servicios"). La Fase 2
// (issue #135) intentó mover el enlace a un texto dentro de
// `CatalogPage.vue`, pero esa pantalla tiene una prueba de HU-022 que
// verifica expresamente que su copy nunca mencione "barbero" (fuera de
// su alcance de una sola preocupación); forzar el enlace ahí habría
// violado esa disciplina. Hasta que la Fase 5 fusione de verdad esta
// capacidad dentro de "Servicios", esta entrada se conserva como sexto
// ítem temporal del dock en vez de quedar inalcanzable.
import type { NavItem } from '@/shared/navigation/navItem'

export { barberServicesPrivateShellChildRoutes } from './routes'

export const barberServicesNavItems: NavItem[] = [
  { to: { name: 'barber-services' }, label: 'Servicios por barbero' },
]
