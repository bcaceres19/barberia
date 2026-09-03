// Módulo: horario laboral recurrente de cada barbero (HU-040) y bloqueos de
// agenda (HU-042). API pública mínima para `app`, según
// docs/03-desarrollo/estandar-frontend-vue.md: las rutas hijas privadas
// (cargadas de forma diferida) y las entradas de navegación que las
// representan. `schedules-bloqueos` ya existía en el router pero no tenía
// entrada de navegación (Fase 2 del rediseño NAVA, issue #187): sin ella la
// ruta solo era alcanzable escribiendo la URL a mano. Formulario,
// validación y cliente API internos permanecen privados.
import type { NavItem } from '@/shared/navigation/navItem'

export { schedulesPrivateShellChildRoutes } from './routes'

export const schedulesNavItems: NavItem[] = [
  { to: { name: 'schedules-horarios' }, label: 'Horarios' },
  { to: { name: 'schedules-bloqueos' }, label: 'Bloqueos' },
]
