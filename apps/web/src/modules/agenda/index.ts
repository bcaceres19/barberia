// Módulo: agenda diaria del barbero y ciclo de vida de los turnos. HU-061
// entregó la creación manual de turnos ("Nuevo turno"); HU-062 agrega la
// agenda diaria real en la raíz del panel privado (`/panel`, name 'panel').
// API pública mínima según docs/03-desarrollo/estandar-frontend-vue.md.
import type { NavItem } from '@/shared/navigation/navItem'

export { agendaPrivateShellChildRoutes } from './routes'

export const agendaNavItems: NavItem[] = [
  { to: { name: 'agenda-nuevo-turno' }, label: 'Nuevo turno' },
]
