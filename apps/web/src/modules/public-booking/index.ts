// Módulo: flujo público de reserva de turnos, sin sesión del barbero.
// API pública mínima para `app`, según docs/03-desarrollo/
// estandar-frontend-vue.md: las rutas (cargadas de forma diferida).
// HU-090 entrega la primera capacidad real: `/reservar/:slug` abre el
// contexto público de una barbería a partir de su enlace de reservas. Los
// componentes, el estado y el cliente API internos permanecen privados.
export { publicBookingRoutes } from './routes'
