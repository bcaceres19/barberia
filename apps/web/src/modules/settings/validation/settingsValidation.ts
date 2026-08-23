// Validación de forma en cliente (docs/03-desarrollo/estandar-frontend-vue.md
// §7.6): ayuda a corregir antes de enviar, nunca sustituye al backend. La
// zona horaria NO se valida contra un catálogo IANA aquí a propósito
// (fuera de alcance: "catálogo nuevo de zonas servido por un endpoint");
// solo se comprueba que no esté vacía ni exceda el largo. La confirmación
// real contra pg_timezone_names ocurre en el servidor (CA-020-03) y su
// 422 se muestra como un error general del formulario, no de este campo.
export const NAME_MAX_LENGTH = 120
export const TIMEZONE_MAX_LENGTH = 64
export const CONTACT_EMAIL_MAX_LENGTH = 254

// Misma forma mínima que barbershop_contact_email_ck en la base: sin
// espacios, con '@' y un '.' posterior (no una validación RFC 5322
// completa, docs/06-api/estandar-openapi.md "no crear reglas más
// restrictivas que las fuentes").
const CONTACT_EMAIL_SHAPE_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

// Exactamente la misma expresión que barbershop_contact_phone_ck.
const CONTACT_PHONE_E164_PATTERN = /^\+[1-9][0-9]{7,14}$/

export function validateName(name: string): string | undefined {
  const trimmed = name.trim()
  if (!trimmed) return 'Escribe el nombre de la barbería.'
  if (trimmed.length > NAME_MAX_LENGTH)
    return `El nombre no puede superar ${NAME_MAX_LENGTH} caracteres.`
  return undefined
}

export function validateTimezone(timezone: string): string | undefined {
  const trimmed = timezone.trim()
  if (!trimmed) return 'Escribe la zona horaria.'
  if (trimmed.length > TIMEZONE_MAX_LENGTH)
    return `La zona horaria no puede superar ${TIMEZONE_MAX_LENGTH} caracteres.`
  return undefined
}

export function validateContactEmail(contactEmail: string): string | undefined {
  const trimmed = contactEmail.trim()
  if (!trimmed) return undefined // vacío es válido: significa "sin correo de contacto".
  if (trimmed.length > CONTACT_EMAIL_MAX_LENGTH || !CONTACT_EMAIL_SHAPE_PATTERN.test(trimmed)) {
    return 'Escribe un correo con formato válido, o déjalo vacío.'
  }
  return undefined
}

export function validateContactPhone(contactPhone: string): string | undefined {
  const trimmed = contactPhone.trim()
  if (!trimmed) return undefined // vacío es válido: significa "sin teléfono de contacto".
  if (!CONTACT_PHONE_E164_PATTERN.test(trimmed)) {
    return 'Escribe un teléfono en formato internacional (ej. +573001234567), o déjalo vacío.'
  }
  return undefined
}
