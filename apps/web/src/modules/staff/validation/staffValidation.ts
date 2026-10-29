// Validación de forma en cliente (docs/03-desarrollo/estandar-frontend-vue.md
// §7.6): ayuda a corregir antes de enviar, nunca sustituye al backend
// (CA-021-03 real vive en el servicio Go). No exige dos palabras ni
// unicidad: un nombre repetido dentro de la misma barbería es válido.
import type { Vocabulary } from '@/shared/model'

export const FULL_NAME_MAX_LENGTH = 120

export function validateFullName(fullName: string, v: Vocabulary): string | undefined {
  const trimmed = fullName.trim()
  if (!trimmed) return `Escribe el nombre ${v.ofTheProfessional}.`
  if (trimmed.length > FULL_NAME_MAX_LENGTH) {
    return `El nombre no puede superar ${FULL_NAME_MAX_LENGTH} caracteres.`
  }
  return undefined
}

// Mismo tope que `barber_description_ck` y el contrato (DEC-127). El servidor
// cuenta caracteres Unicode, no unidades UTF-16: aquí igual, para que un emoji
// no cuente doble y el cliente rechace exactamente lo que el servidor rechaza.
export const DESCRIPTION_MAX_LENGTH = 280

export function descriptionLength(description: string): number {
  return Array.from(description.trim()).length
}

export function validateDescription(description: string): string | undefined {
  if (descriptionLength(description) > DESCRIPTION_MAX_LENGTH) {
    return `La descripción no puede superar ${DESCRIPTION_MAX_LENGTH} caracteres.`
  }
  return undefined
}
