// Módulo: agenda diaria del barbero y ciclo de vida de los turnos. HU-061
// entrega la primera pantalla real: creación manual de turnos ("Nuevo
// turno"). La agenda diaria en sí (HU-062, listar/ver el día) todavía no
// existe; API pública mínima según docs/03-desarrollo/estandar-frontend-vue.md.
import type { NavItem } from '@/shared/navigation/navItem'

export { agendaPrivateShellChildRoutes } from './routes'

export const agendaNavItems: NavItem[] = [
  { to: { name: 'agenda-nuevo-turno' }, label: 'Nuevo turno' },
]
