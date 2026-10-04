// Módulo: catálogo básico de servicios de la barbería (HU-022). API
// pública mínima para `app`, según docs/03-desarrollo/estandar-frontend-vue.md:
// la ruta hija privada (cargada de forma diferida) y la entrada de
// navegación que la representa. `app/router/index.ts` combina esto con las
// rutas de `auth`/`settings`/`staff` (y de cualquier otro módulo con
// pantalla privada) antes de pasarlas a `auth.privateShellRoute`; este
// módulo no monta su propio cascarón ni importa internos de otro módulo.
// Formulario, validación y cliente API internos permanecen privados.
import { defineAsyncComponent } from 'vue'
import type { NavItem } from '@/shared/navigation/navItem'

export { catalogPrivateShellChildRoutes } from './routes'

// `app` compone esta pantalla con aportes de otros módulos (slot `service-offer`,
// DEC-115) sin que `catalog` conozca a ninguno.
export const CatalogPage = defineAsyncComponent(() => import('./pages/CatalogPage.vue'))
export type { Service as CatalogService } from './model/service'

export const catalogNavItems: NavItem[] = [
  { to: { name: 'catalog-servicios' }, label: 'Servicios', primary: true },
]
