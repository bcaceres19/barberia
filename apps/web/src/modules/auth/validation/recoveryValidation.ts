// Validación de forma en cliente para los tres pasos de HU-011 (trabajo
// requerido §3/§5, docs/03-desarrollo/estandar-frontend-vue.md §7.6): ayuda
// a corregir antes de enviar, nunca sustituye al backend. La política de
// contraseña (`DEC-063`) se explica ANTES de escribir (CA-011-06) y se
// valida aquí solo en lo que el cliente puede anticipar sin conocer datos
// del servidor (longitud, igual al correo); "igual a la contraseña
// actual" solo el backend puede evaluarlo (ver `recoveryOutcome.ts`).
import { EMAIL_MAX_LENGTH, EMAIL_SHAPE_PATTERN } from './loginValidation'

// RecoveryResetPasswordRequest.yaml fija minLength: 10, maxLength: 128.
export const NEW_PASSWORD_MIN_LENGTH = 10
export const NEW_PASSWORD_MAX_LENGTH = 128

export function validateRecoveryEmail(email: string): string | undefined {
  const trimmed = email.trim()
  if (!trimmed) return 'Escribe tu correo.'
  if (trimmed.length > EMAIL_MAX_LENGTH || !EMAIL_SHAPE_PATTERN.test(trimmed)) {
    return 'Escribe un correo con formato válido.'
  }
  return undefined
}

// Formato E.164 que exige el contrato y `staff_user_phone_ck`.
const PHONE_E164_PATTERN = /^\+[1-9][0-9]{7,14}$/
// Separadores de presentación que el servidor también descarta.
const PHONE_SEPARATORS = /[\s\-.()]/g

/** Deja el número como lo espera el servidor (`+573001234567`): recorta y
 * quita espacios, guiones, puntos y paréntesis. No valida. */
export function normalizeRecoveryPhone(phone: string): string {
  return phone.trim().replace(PHONE_SEPARATORS, '')
}

export function validateRecoveryPhone(phone: string): string | undefined {
  if (!phone.trim()) return 'Escribe tu número de WhatsApp.'
  if (!PHONE_E164_PATTERN.test(normalizeRecoveryPhone(phone))) {
    return 'Escribe el número con el indicativo de tu país, por ejemplo +573001234567.'
  }
  return undefined
}

export function validateRecoveryCode(code: string): string | undefined {
  if (!code) return 'Escribe el código.'
  if (!/^[0-9]{6}$/.test(code)) return 'El código tiene 6 dígitos numéricos.'
  return undefined
}

export function validateNewPassword(password: string, email: string): string | undefined {
  if (!password) return 'Escribe la contraseña nueva.'
  if (password.length < NEW_PASSWORD_MIN_LENGTH) {
    return `La contraseña debe tener al menos ${NEW_PASSWORD_MIN_LENGTH} caracteres.`
  }
  if (password.length > NEW_PASSWORD_MAX_LENGTH) {
    return `La contraseña no puede superar ${NEW_PASSWORD_MAX_LENGTH} caracteres.`
  }
  // Con WhatsApp el cliente no conoce el correo (email vacío): esa
  // comparación la hace el servidor, que sí lo conoce.
  if (email && password.toLowerCase() === email.trim().toLowerCase()) {
    return 'La contraseña no puede ser igual a tu correo.'
  }
  return undefined
}
