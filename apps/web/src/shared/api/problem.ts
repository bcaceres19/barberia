// Forma de error uniforme del API (RFC 9457, docs/06-api/estandar-openapi.md
// §12). `Problem` nace del bundle OpenAPI (nunca se declara a mano): toda
// respuesta que no sea 2xx del cliente tipado de `shared/api` tiene esta
// forma, para cualquier módulo que la necesite, no solo `auth`.
import type { components } from './generated/openapi.d.ts'

export type Problem = components['schemas']['Problem']

/**
 * Comprueba en tiempo de ejecución que un valor `unknown` (por ejemplo, el
 * cuerpo decodificado de una respuesta de error) tiene la forma mínima de
 * `Problem`. Un módulo mapea errores por `status`/`code`, nunca por
 * `detail` (docs/03-desarrollo/estandar-frontend-vue.md §6): esta función
 * solo valida la forma, no decide el mensaje.
 */
export function isProblem(value: unknown): value is Problem {
  if (typeof value !== 'object' || value === null) return false
  const candidate = value as Record<string, unknown>
  return (
    typeof candidate.status === 'number' &&
    typeof candidate.code === 'string' &&
    typeof candidate.title === 'string'
  )
}
