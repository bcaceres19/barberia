/*
 * Modelo de contenido de los mockups de `/acceso` y `/recuperar-acceso`
 * (docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03).
 *
 * Cada evento se declara una sola vez y se renderiza en los dos viewports,
 * de modo que escritorio y móvil no pueden divergir en copy, orden ni
 * estado. Los textos provienen del módulo real `apps/web/src/modules/auth`
 * (validaciones, resúmenes de error y rótulos de botón) y de `DEC-081`
 * para el canal de entrega del código. Ningún texto se inventa aquí: si
 * una cadena no existe todavía en el código, procede de una decisión
 * `DEC-*` registrada.
 */

// Copys canónicos tomados del código real. Se centralizan para que un
// cambio en la app se refleje en los 46 PNG con una sola edición.
export const COPY = {
  // apps/web/src/modules/auth/components/LoginForm.vue
  loginSubmit: 'Iniciar sesión',
  loginSubmitting: 'Iniciando sesión…',
  recoveryLink: '¿Olvidaste tu contraseña?',
  summaryTitle: 'Revisa estos campos',

  // apps/web/src/modules/auth/validation/loginValidation.ts
  emailEmpty: 'Escribe tu correo.',
  emailShape: 'Escribe un correo con formato válido.',
  passwordEmpty: 'Escribe tu contraseña.',

  // apps/web/src/modules/auth/validation/recoveryValidation.ts
  newPasswordShort: 'La contraseña debe tener al menos 10 caracteres.',
  confirmMismatch: 'Las dos contraseñas no coinciden.',

  // apps/web/src/modules/auth/pages/LoginPage.vue
  invalidCredentialsTitle: 'No pudimos iniciar tu sesión',
  invalidCredentialsBody: 'Revisa tu correo y contraseña e inténtalo de nuevo.',
  networkTitle: 'No pudimos conectar',
  networkBody: 'Revisa tu conexión e inténtalo de nuevo.',
  networkAction: 'Reintentar',
  sessionExpiredTitle: 'Tu sesión venció',
  sessionExpiredBody: 'Inicia sesión de nuevo para continuar.',

  // Reto adicional: HU-007 / DEC-062 con el canal resuelto por DEC-081.
  challengeTitle: 'Necesitamos verificar este intento',
  challengeBody: 'Superaste el límite de intentos. Confirma el código para continuar.',
  otpLabel: 'Código de 6 dígitos',
  otpVerify: 'Verificar código',
  otpInvalid: 'El código no es válido o venció. Solicita uno nuevo si lo necesitas.',

  // apps/web/src/modules/auth/components/RecoveryVerifyStep.vue
  recoveryResend: 'Reenviar código',
  recoveryResendWait: (s) => `Reenviar en ${s} s`,
  backToLogin: 'Volver al acceso',

  // apps/web/src/modules/auth/components/RecoveryResetStep.vue
  policyTitle: 'Requisitos de la contraseña',
  policyBody:
    'Debe tener entre 10 y 128 caracteres, y ser diferente de tu correo y de tu contraseña actual.',
  resetSubmit: 'Guardar contraseña nueva',
  resetSubmitting: 'Guardando contraseña…',
  expiredTitle: 'El enlace de recuperación venció',
  expiredBody: 'Este paso ya no es válido. Solicita un código nuevo para continuar.',
  expiredAction: 'Solicitar un código nuevo',
  doneTitle: 'Contraseña actualizada',
  doneBody:
    'Actualizamos tu contraseña y cerramos todas tus sesiones activas. Inicia sesión de nuevo con la contraseña nueva.',
  doneAction: 'Ir al acceso',
}

// DEC-081: el canal lo resuelve el servidor desde contactos verificados.
// El texto es condicional y uniforme en las tres variantes para no
// enumerar cuentas; solo cambia el canal configurado que se informa.
const CHANNEL = {
  correo: {
    chip: 'Correo · canal predeterminado',
    note: 'Si la cuenta puede continuar, enviamos el código al correo verificado de la cuenta.',
  },
  whatsapp: {
    chip: 'WhatsApp oficial',
    note: 'Si la cuenta puede continuar, enviamos el código al WhatsApp oficial verificado de la cuenta.',
  },
  ambos: {
    chip: 'Correo y WhatsApp oficial',
    note: 'Si la cuenta puede continuar, enviamos el mismo código al correo y al WhatsApp verificados de la cuenta.',
  },
}

const EMAIL = 'barbero@nava.co'
const PASSWORD = '•'.repeat(11)

const emailField = (over = {}) => ({
  label: 'Correo',
  icon: 'mail',
  placeholder: 'tu-correo@ejemplo.com',
  value: '',
  ...over,
})

const passwordField = (over = {}) => ({
  label: 'Contraseña',
  icon: 'lock',
  toggle: true,
  placeholder: 'Tu contraseña',
  value: '',
  ...over,
})

// Bloque de credenciales del reto: el intento sigue en curso, así que los
// campos quedan bloqueados y la acción principal la toma el reto. Evita
// dos botones primarios compitiendo en la misma pantalla.
const lockedCredentials = () => [
  { type: 'fields', fields: [
    emailField({ value: EMAIL, state: 'disabled' }),
    passwordField({ value: PASSWORD, state: 'disabled' }),
  ] },
  { type: 'submit', label: COPY.loginSubmit, state: 'disabled' },
  { type: 'link', label: COPY.recoveryLink, state: 'muted' },
]

const challengeScreen = (channel, invalid) => ({
  route: 'acceso',
  blocks: [
    ...lockedCredentials(),
    { type: 'alert', variant: 'warning', title: COPY.challengeTitle, body: COPY.challengeBody },
    {
      type: 'challenge',
      chip: CHANNEL[channel].chip,
      note: CHANNEL[channel].note,
      otp: invalid ? { digits: '482913', state: 'error' } : { digits: '', state: 'idle' },
      error: invalid ? COPY.otpInvalid : undefined,
      primary: COPY.otpVerify,
      secondary: COPY.recoveryResendWait(60),
    },
  ],
})

export const ACCESO = {
  title: 'Accede a NAVA',
  subtitle: 'Tu agenda y tu equipo, en un solo lugar.',
  caption: 'ACCESO SEGURO',
  screens: {
    '01-inicial': {
      blocks: [
        { type: 'fields', fields: [emailField(), passwordField()] },
        { type: 'submit', label: COPY.loginSubmit },
        { type: 'link', label: COPY.recoveryLink },
      ],
    },

    '02-enviando': {
      blocks: [
        { type: 'fields', fields: [
          emailField({ value: EMAIL, state: 'disabled' }),
          passwordField({ value: PASSWORD, state: 'disabled' }),
        ] },
        { type: 'submit', label: COPY.loginSubmitting, state: 'loading' },
        { type: 'link', label: COPY.recoveryLink, state: 'muted' },
      ],
    },

    // Validación local con los dos campos vacíos: cada error vive bajo su
    // campo y el resumen global se ancla después del enlace, sin repetir
    // literalmente los mensajes locales.
    '03-validacion': {
      blocks: [
        { type: 'fields', fields: [
          emailField({ state: 'error', error: COPY.emailEmpty }),
          passwordField({ state: 'error', error: COPY.passwordEmpty }),
        ] },
        { type: 'submit', label: COPY.loginSubmit },
        { type: 'link', label: COPY.recoveryLink },
        { type: 'alert', variant: 'danger', title: COPY.summaryTitle, body: 'Faltan el correo y la contraseña para continuar.' },
      ],
    },

    // CA-010-02: el servidor evaluó la credencial y la rechazó, así que la
    // contraseña se limpia; el correo se conserva.
    '04-credenciales-invalidas': {
      blocks: [
        { type: 'fields', fields: [emailField({ value: EMAIL }), passwordField()] },
        { type: 'submit', label: COPY.loginSubmit },
        { type: 'link', label: COPY.recoveryLink },
        { type: 'alert', variant: 'danger', title: COPY.invalidCredentialsTitle, body: COPY.invalidCredentialsBody },
      ],
    },

    // CA-010-03: el servidor nunca evaluó la credencial, así que ambos
    // datos se conservan y la alerta ofrece reintentar en un solo toque.
    '05-sin-conexion': {
      blocks: [
        { type: 'fields', fields: [
          emailField({ value: EMAIL }),
          passwordField({ value: PASSWORD }),
        ] },
        { type: 'submit', label: COPY.loginSubmit },
        { type: 'link', label: COPY.recoveryLink },
        { type: 'alert', variant: 'warning', title: COPY.networkTitle, body: COPY.networkBody, action: COPY.networkAction },
      ],
    },

    // Llegada desde una ruta privada tras un 401: es una pantalla nueva,
    // no un reintento, así que los campos están vacíos.
    '06-sesion-vencida': {
      blocks: [
        { type: 'fields', fields: [emailField(), passwordField()] },
        { type: 'submit', label: COPY.loginSubmit },
        { type: 'link', label: COPY.recoveryLink },
        { type: 'alert', variant: 'info', title: COPY.sessionExpiredTitle, body: COPY.sessionExpiredBody, dismissible: true },
      ],
    },

    '07-reto-envio-correo': challengeScreen('correo', false),
    '08-reto-envio-whatsapp': challengeScreen('whatsapp', false),
    '09-reto-envio-ambos': challengeScreen('ambos', false),
    '10-reto-codigo-invalido-correo': challengeScreen('correo', true),
    '11-reto-codigo-invalido-whatsapp': challengeScreen('whatsapp', true),
    '12-reto-codigo-invalido-ambos': challengeScreen('ambos', true),
  },
}

const REQUEST_HINT =
  'Escribe el correo de tu cuenta. Si existe, enviamos un código de un solo uso por los canales configurados para este evento; correo es el predeterminado.'
const VERIFY_HINT =
  'Si la cuenta puede continuar, ya enviamos un código de 6 dígitos por los canales configurados para este evento.'

export const RECUPERACION = {
  title: null, // cada paso trae su propio título
  caption: 'RECUPERACIÓN SEGURA',
  screens: {
    '01-solicitud-inicial': {
      step: 1,
      title: 'Solicita tu código',
      blocks: [
        { type: 'note', text: REQUEST_HINT },
        { type: 'fields', fields: [emailField()] },
        { type: 'submit', label: 'Enviar código' },
        { type: 'link', label: COPY.backToLogin },
      ],
    },

    '02-solicitud-validacion': {
      step: 1,
      title: 'Solicita tu código',
      blocks: [
        { type: 'note', text: REQUEST_HINT },
        { type: 'fields', fields: [emailField({ state: 'error', error: COPY.emailEmpty })] },
        { type: 'submit', label: 'Enviar código' },
        { type: 'link', label: COPY.backToLogin },
      ],
    },

    '03-solicitud-enviando': {
      step: 1,
      title: 'Solicita tu código',
      blocks: [
        { type: 'note', text: REQUEST_HINT },
        { type: 'fields', fields: [emailField({ value: EMAIL, state: 'disabled' })] },
        { type: 'submit', label: 'Enviando código…', state: 'loading' },
        { type: 'link', label: COPY.backToLogin, state: 'muted' },
      ],
    },

    '04-solicitud-sin-conexion': {
      step: 1,
      title: 'Solicita tu código',
      blocks: [
        { type: 'note', text: REQUEST_HINT },
        { type: 'fields', fields: [emailField({ value: EMAIL })] },
        { type: 'submit', label: 'Enviar código' },
        { type: 'link', label: COPY.backToLogin },
        { type: 'alert', variant: 'warning', title: COPY.networkTitle, body: COPY.networkBody, action: COPY.networkAction },
      ],
    },

    '05-verificacion-inicial': {
      step: 2,
      title: 'Verifica el código',
      blocks: [
        { type: 'note', text: VERIFY_HINT },
        { type: 'otp', label: COPY.otpLabel, digits: '', state: 'idle' },
        { type: 'submit', label: COPY.otpVerify },
        { type: 'ghost', label: COPY.recoveryResendWait(42) },
        { type: 'link', label: COPY.backToLogin },
      ],
    },

    // Mismo tratamiento que el reto de acceso: el error específico vive
    // bajo las casillas, sin una alerta global que lo repita.
    '06-verificacion-codigo-invalido': {
      step: 2,
      title: 'Verifica el código',
      blocks: [
        { type: 'note', text: VERIFY_HINT },
        { type: 'otp', label: COPY.otpLabel, digits: '482913', state: 'error', error: COPY.otpInvalid },
        { type: 'submit', label: COPY.otpVerify },
        { type: 'ghost', label: COPY.recoveryResend },
        { type: 'link', label: COPY.backToLogin },
      ],
    },

    '07-contrasena-inicial': {
      step: 3,
      title: 'Establece tu contraseña nueva',
      blocks: [
        { type: 'note', text: 'Verificamos tu código, enviado a ••• ••• 4821 y b•••••@nava.co.' },
        { type: 'alert', variant: 'info', title: COPY.policyTitle, body: COPY.policyBody, plain: true },
        { type: 'fields', fields: [
          passwordField({ label: 'Contraseña nueva', placeholder: 'Contraseña nueva' }),
          passwordField({ label: 'Confirma la contraseña nueva', placeholder: 'Repite la contraseña' }),
        ] },
        { type: 'submit', label: COPY.resetSubmit },
        { type: 'link', label: COPY.backToLogin },
      ],
    },

    '08-contrasena-validacion': {
      step: 3,
      title: 'Establece tu contraseña nueva',
      blocks: [
        { type: 'alert', variant: 'info', title: COPY.policyTitle, body: COPY.policyBody, plain: true },
        { type: 'fields', fields: [
          passwordField({ label: 'Contraseña nueva', value: '•'.repeat(5), state: 'error', error: COPY.newPasswordShort }),
          passwordField({ label: 'Confirma la contraseña nueva', value: '•'.repeat(6), state: 'error', error: COPY.confirmMismatch }),
        ] },
        { type: 'submit', label: COPY.resetSubmit },
        { type: 'link', label: COPY.backToLogin },
        { type: 'alert', variant: 'danger', title: COPY.summaryTitle, body: 'Ajusta la contraseña nueva y su confirmación para continuar.' },
      ],
    },

    '09-contrasena-enviando': {
      step: 3,
      title: 'Establece tu contraseña nueva',
      blocks: [
        { type: 'alert', variant: 'info', title: COPY.policyTitle, body: COPY.policyBody, plain: true },
        { type: 'fields', fields: [
          passwordField({ label: 'Contraseña nueva', value: '•'.repeat(14), state: 'disabled' }),
          passwordField({ label: 'Confirma la contraseña nueva', value: '•'.repeat(14), state: 'disabled' }),
        ] },
        { type: 'submit', label: COPY.resetSubmitting, state: 'loading' },
        { type: 'link', label: COPY.backToLogin, state: 'muted' },
      ],
    },

    // El motivo se explica antes del remedio: la alerta precede al botón
    // que reinicia el flujo.
    '10-contrasena-enlace-vencido': {
      step: 3,
      title: 'Establece tu contraseña nueva',
      blocks: [
        { type: 'alert', variant: 'danger', title: COPY.expiredTitle, body: COPY.expiredBody },
        { type: 'submit', label: COPY.expiredAction },
        { type: 'link', label: COPY.backToLogin },
      ],
    },

    '11-completado': {
      step: 3,
      title: COPY.doneTitle,
      blocks: [
        { type: 'alert', variant: 'success', title: 'Listo', body: COPY.doneBody },
        { type: 'submit', label: COPY.doneAction },
      ],
    },
  },
}
