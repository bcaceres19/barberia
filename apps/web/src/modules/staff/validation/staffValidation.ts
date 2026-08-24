// Validación de forma en cliente (docs/03-desarrollo/estandar-frontend-vue.md
// §7.6): ayuda a corregir antes de enviar, nunca sustituye al backend
// (CA-021-03 real vive en el servicio Go). No exige dos palabras ni
// unicidad: un nombre repetido dentro de la misma barbería es válido.
export const FULL_NAME_MAX_LENGTH = 120

export function validateFullName(fullName: string): string | undefined {
  const trimmed = fullName.trim()
  if (!trimmed) return 'Escribe el nombre del barbero.'
  if (trimmed.length > FULL_NAME_MAX_LENGTH) {
    return `El nombre no puede superar ${FULL_NAME_MAX_LENGTH} caracteres.`
  }
  return undefined
}
