// Estado visible de la pantalla de acceso (unión discriminada,
// docs/03-desarrollo/estandar-frontend-vue.md §5.3). `LoginPage.vue` es la
// única propietaria; `LoginForm.vue` solo recibe props derivadas de este
// estado, nunca el estado completo con detalles de transporte.
export type LoginScreenState =
  | { status: 'idle' }
  | { status: 'submitting' }
  | { status: 'invalid-credentials' }
  | { status: 'validation-error' }
  | { status: 'network-error' }
  | { status: 'rate-limited'; retryAfterSeconds?: number }
  | { status: 'unexpected-error'; requestId?: string }
