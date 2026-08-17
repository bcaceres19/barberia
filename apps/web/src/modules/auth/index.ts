// Módulo: acceso del barbero (inicio de sesión, sesión privada y
// recuperación de acceso). API pública mínima para `app`, según
// docs/03-desarrollo/estandar-frontend-vue.md: las rutas (cargadas de forma
// diferida) y, desde HU-012, la instalación de la coordinación de sesión
// que solo `app` puede conectar al router real. Los componentes, el estado
// y el cliente API internos permanecen privados.
export { authRoutes } from './routes'
export { installSessionHandling } from './bootstrap/installSessionHandling'
