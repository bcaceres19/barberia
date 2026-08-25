// Módulo: horario laboral recurrente de cada barbero (HU-040). API pública
// mínima para `app`, según docs/03-desarrollo/estandar-frontend-vue.md: la
// ruta hija privada (cargada de forma diferida) y la entrada de navegación
// que la representa. Excepciones por fecha/festivos (HU-041) y bloqueos
// (HU-042) se agregan a este mismo módulo cuando existan. Formulario,
// validación y cliente API internos permanecen privados.
import type { NavItem } from '@/shared/navigation/navItem'

export { schedulesPrivateShellChildRoutes } from './routes'

export const schedulesNavItems: NavItem[] = [
  { to: { name: 'schedules-horarios' }, label: 'Horarios' },
]
