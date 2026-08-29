// Módulo: acceso del barbero (inicio de sesión, sesión privada y
// recuperación de acceso). API pública mínima para `app`, según
// docs/03-desarrollo/estandar-frontend-vue.md: las rutas (cargadas de forma
// diferida), la instalación de la coordinación de sesión que solo `app`
// puede conectar al router real, la fábrica del único cascarón privado
// (`privateShellRoute`, HU-020: `app` la invoca combinando
// `privateShellChildRoutes` con las hijas privadas de cualquier otro
// módulo, sin crear un segundo guard ni un árbol de rutas paralelo;
// `privateShellChildRoutes` está vacío desde HU-062, porque `/panel` ahora
// lo sirve `agenda`) y una función estrecha para que un módulo hermano
// actualice el nombre de la barbería ya confirmado por el servidor
// (`updateBarbershopName`, HU-020). Los componentes, el estado y el cliente
// API internos permanecen privados.
export { authRoutes, privateShellChildRoutes, privateShellRoute } from './routes'
export { installSessionHandling } from './bootstrap/installSessionHandling'
export { updateBarbershopName } from './model/sessionStore'
