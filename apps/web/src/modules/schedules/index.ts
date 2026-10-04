// Módulo: horario laboral recurrente de cada barbero (HU-040) y bloqueos de
// agenda (HU-042). API pública mínima para `app`, según
// docs/03-desarrollo/estandar-frontend-vue.md: las rutas hijas privadas
// (cargadas de forma diferida) y las entradas de navegación que las
// representan. `schedules-bloqueos` ya existía en el router pero no tenía
// entrada de navegación. DEC-105 integra los bloqueos en Barberos; la
// ruta antigua se mantiene mediante redirección. Formulario,
// validación y cliente API internos permanecen privados.
import { defineAsyncComponent } from 'vue'
import type { NavItem } from '@/shared/navigation/navItem'

export { schedulesPrivateShellChildRoutes } from './routes'

export const schedulesNavItems: NavItem[] = [
  {
    to: { name: 'schedules-horarios' },
    label: 'Horarios',
    // Quien trabaja solo vive de su horario: sube al dock (DEC-115).
    profiles: { solo: { primary: true } },
  },
]

export const BarberBlocksPanel = defineAsyncComponent(
  () => import('./components/BarberBlocksPanel.vue'),
)
