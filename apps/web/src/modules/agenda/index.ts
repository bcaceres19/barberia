// Módulo: agenda diaria del barbero y ciclo de vida de los turnos. HU-061
// entregó la creación manual de turnos ("Nuevo turno"); HU-062 agrega la
// agenda diaria real en la raíz del panel privado (`/panel`, name 'panel').
// API pública mínima según docs/03-desarrollo/estandar-frontend-vue.md.
// "Nuevo turno" es la acción primaria global del header NAVA, no un
// destino de navegación (especificacion-frontend-nava.md §5.1); por eso
// este módulo no aporta ninguna entrada al dock. "Agenda" (name 'panel')
// vive como entrada base en `AppNav.vue`, dueño de `auth`.
import type { NavItem } from '@/shared/navigation/navItem'

export { agendaPrivateShellChildRoutes } from './routes'

export const agendaNavItems: NavItem[] = []
