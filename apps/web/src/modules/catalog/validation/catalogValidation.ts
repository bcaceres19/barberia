// Validación de forma en cliente (docs/03-desarrollo/estandar-frontend-vue.md
// §7.6): ayuda a corregir antes de enviar, nunca sustituye al backend (la
// validación real, incluida la unicidad de nombre entre activos DEC-067,
// vive en el servicio Go). Duraciones sin lista cerrada (RN-SER-02): 25 o
// 45 minutos son igual de válidos que 30, 60 o 90.
export const NAME_MAX_LENGTH = 120
export const DESCRIPTION_MAX_LENGTH = 500
export const DURATION_MIN_MINUTES = 1
export const DURATION_MAX_MINUTES = 1440

// pricePattern es EXACTAMENTE la misma forma que el contrato documenta
// (`^[0-9]{1,10}(\.[0-9]{1,2})?$`, api/openapi/components/schemas/
// CreateServiceRequest.yaml): un decimal sin signo, hasta dos cifras de
// fracción. Validar aquí evita un 422 evitable por un formato claramente
// mal escrito, sin duplicar la regla de "mayor que cero" (DEC-067), que
// depende del valor numérico, no solo de la forma.
const pricePattern = /^[0-9]{1,10}(\.[0-9]{1,2})?$/

export function validateName(name: string): string | undefined {
  const trimmed = name.trim()
  if (!trimmed) return 'Escribe el nombre del servicio.'
  if (trimmed.length > NAME_MAX_LENGTH) {
    return `El nombre no puede superar ${NAME_MAX_LENGTH} caracteres.`
  }
  return undefined
}

export function validateDescription(description: string): string | undefined {
  if (description.trim().length > DESCRIPTION_MAX_LENGTH) {
    return `La descripción no puede superar ${DESCRIPTION_MAX_LENGTH} caracteres.`
  }
  return undefined
}

export function validateDurationMinutes(raw: string): string | undefined {
  const trimmed = raw.trim()
  if (!trimmed) return 'Escribe la duración en minutos.'
  if (!/^\d+$/.test(trimmed)) return 'La duración debe ser un número entero de minutos.'
  const minutes = Number(trimmed)
  if (minutes < DURATION_MIN_MINUTES || minutes > DURATION_MAX_MINUTES) {
    return `La duración debe estar entre ${DURATION_MIN_MINUTES} y ${DURATION_MAX_MINUTES} minutos.`
  }
  return undefined
}

export function validatePrice(raw: string): string | undefined {
  const trimmed = raw.trim()
  if (!trimmed) return 'Escribe el precio en pesos colombianos.'
  if (!pricePattern.test(trimmed)) {
    return 'Escribe un precio válido, con hasta dos cifras decimales (por ejemplo 45000 o 45000.50).'
  }
  // DEC-067: sin servicios gratuitos. parseFloat es seguro AQUÍ porque solo
  // se usa para comparar contra cero después de que pricePattern ya
  // confirmó la forma exacta del texto; el valor que se envía al servidor
  // sigue siendo el string original, nunca este número (nunca se
  // reserializa con coma flotante).
  if (Number.parseFloat(trimmed) <= 0) {
    return 'El precio debe ser mayor que cero.'
  }
  return undefined
}
