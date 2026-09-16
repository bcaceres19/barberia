// Módulo: acceso del cliente a su turno mediante la credencial del enlace,
// sin sesión ni cuenta. API pública mínima para `app`, según docs/03-
// desarrollo/estandar-frontend-vue.md: las rutas (cargadas de forma
// diferida). HU-098 entrega la primera capacidad real: `/mi-turno/:token`
// consulta la cita correspondiente a esa credencial. Los componentes, el
// estado y el cliente API internos permanecen privados.
export { customerAccessRoutes } from './routes'
