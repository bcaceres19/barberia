// Validación de forma en cliente (trabajo requerido §5): ayuda a corregir
// antes de enviar, nunca sustituye la validación del backend
// (docs/03-desarrollo/estandar-frontend-vue.md §7.6). No intenta decidir
// "¿existe este correo?" — CA-005-02/CA-010-02 exigen que esa respuesta
// nazca únicamente del backend, indistinguible entre correo inexistente y
// contraseña incorrecta.

export interface LoginFieldErrors {
  email?: string
  password?: string
}

// Patrón mínimo de forma (contiene "@" y un dominio con punto), igual de
// permisivo que `LoginRequest.yaml` (format: email, sin regex adicional
// del lado del servidor): más estricto aquí solo rechazaría formularios
// válidos que el backend sí aceptaría.
const EMAIL_SHAPE_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

// LoginRequest.yaml fija maxLength: 254 (email) y 256 (password).
const EMAIL_MAX_LENGTH = 254
const PASSWORD_MAX_LENGTH = 256

export function validateLoginForm(email: string, password: string): LoginFieldErrors {
  const errors: LoginFieldErrors = {}
  const trimmedEmail = email.trim()

  if (!trimmedEmail) {
    errors.email = 'Escribe tu correo.'
  } else if (trimmedEmail.length > EMAIL_MAX_LENGTH || !EMAIL_SHAPE_PATTERN.test(trimmedEmail)) {
    errors.email = 'Escribe un correo con formato válido.'
  }

  if (!password) {
    errors.password = 'Escribe tu contraseña.'
  } else if (password.length > PASSWORD_MAX_LENGTH) {
    errors.password = 'La contraseña es demasiado larga.'
  }

  return errors
}

export function hasLoginFieldErrors(errors: LoginFieldErrors): boolean {
  return Object.keys(errors).length > 0
}
