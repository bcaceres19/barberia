// Validación de forma del formulario de datos del cliente (HU-096), espejo
// de las mismas reglas que apps/api/internal/modules/publicbooking/
// identity.go aplica (NormalizeAndValidateCustomerIdentity): una entrada
// inválida se detiene aquí antes de gastar una interacción, pero el
// servidor sigue siendo la fuente de verdad final cuando HU-097 la reciba.
//
// A diferencia de agenda/validation/appointmentValidation.ts (HU-061,
// teléfono/correo opcionales por RN-CIT-02), aquí los tres campos de
// contacto son obligatorios (CA-096-02): el núcleo de public-booking no
// importa el módulo agenda (mismo criterio de independencia entre módulos
// que CA-002-06 aplica en el backend), así que las reglas se repiten aquí.
const PHONE_PATTERN = /^\+[1-9][0-9]{7,14}$/
const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export const CUSTOMER_FULL_NAME_MAX_LENGTH = 120
export const ATTENDEE_NAME_MAX_LENGTH = 120
export const CUSTOMER_NOTE_MAX_LENGTH = 500

export function validateCustomerFullName(value: string): string | undefined {
  const trimmed = value.trim()
  if (trimmed === '') return 'El nombre es obligatorio.'
  if (trimmed.length > CUSTOMER_FULL_NAME_MAX_LENGTH) return 'El nombre es demasiado largo.'
  return undefined
}

export function validateCustomerPhone(value: string): string | undefined {
  const trimmed = value.trim()
  if (trimmed === '') return 'El teléfono es obligatorio.'
  return PHONE_PATTERN.test(trimmed)
    ? undefined
    : 'El teléfono debe incluir el indicativo de país, por ejemplo +573001234567.'
}

export function validateCustomerEmail(value: string): string | undefined {
  const trimmed = value.trim()
  if (trimmed === '') return 'El correo es obligatorio.'
  return EMAIL_PATTERN.test(trimmed) ? undefined : 'El correo tiene un formato inválido.'
}

export function validateCustomerNote(value: string): string | undefined {
  if (value.length > CUSTOMER_NOTE_MAX_LENGTH) return 'La nota no puede superar los 500 caracteres.'
  return undefined
}

// validateAttendeeName solo se evalúa cuando forSomeoneElse es true
// (CA-096-01): reservar para sí mismo nunca pide -ni valida- este campo.
export function validateAttendeeName(value: string): string | undefined {
  const trimmed = value.trim()
  if (trimmed === '') return 'El nombre de la persona atendida es obligatorio.'
  if (trimmed.length > ATTENDEE_NAME_MAX_LENGTH)
    return 'El nombre de la persona atendida es demasiado largo.'
  return undefined
}
