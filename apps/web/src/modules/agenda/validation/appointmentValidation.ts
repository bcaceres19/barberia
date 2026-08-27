// Validación de forma del formulario "Nuevo turno" (HU-061), espejo de las
// mismas reglas que el backend ya aplica
// (api/openapi/components/schemas/CreateManualAppointmentRequest.yaml):
// una entrada inválida se detiene aquí antes de gastar una solicitud, pero
// el servidor sigue siendo la fuente de verdad final.
const PHONE_PATTERN = /^\+[1-9][0-9]{7,14}$/
const CIVIL_DATETIME_PATTERN = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}$/

export function validateAttendeeName(value: string): string | undefined {
  const trimmed = value.trim()
  if (trimmed === '') return 'El nombre de la persona atendida es obligatorio.'
  if (trimmed.length > 120) return 'El nombre de la persona atendida es demasiado largo.'
  return undefined
}

export function validateCustomerFullName(value: string): string | undefined {
  const trimmed = value.trim()
  if (trimmed === '') return 'El nombre del cliente es obligatorio.'
  if (trimmed.length > 120) return 'El nombre del cliente es demasiado largo.'
  return undefined
}

export function validateCustomerPhone(value: string): string | undefined {
  if (value.trim() === '') return undefined
  return PHONE_PATTERN.test(value.trim())
    ? undefined
    : 'El teléfono debe incluir el indicativo de país, por ejemplo +573001234567.'
}

export function validateCustomerEmail(value: string): string | undefined {
  if (value.trim() === '') return undefined
  const v = value.trim()
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v) ? undefined : 'El correo tiene un formato inválido.'
}

export function validateCustomerNote(value: string): string | undefined {
  if (value.length > 500) return 'La nota no puede superar los 500 caracteres.'
  return undefined
}

export function validateStartsAt(civilDateTime: string): string | undefined {
  if (civilDateTime === '') return 'Elige la fecha y la hora del turno.'
  return CIVIL_DATETIME_PATTERN.test(civilDateTime)
    ? undefined
    : 'La fecha y la hora no tienen un formato válido.'
}

// buildStartsAt compone el valor "AAAA-MM-DDTHH:MM:SS" que el contrato
// exige a partir de los campos separados de fecha y hora del formulario
// (inputs type="date"/type="time" nativos): sin conversión de zona, el
// servidor interpreta este valor civil contra la zona IANA de la barbería
// (RN-DIS-07).
export function buildStartsAt(date: string, time: string): string {
  const withSeconds = time.length === 5 ? `${time}:00` : time
  return `${date}T${withSeconds}`
}
