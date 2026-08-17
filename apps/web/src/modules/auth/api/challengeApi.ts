// Único punto del módulo `auth` que llama al cliente HTTP tipado para el
// reto telefónico de HU-007 (DEC-062), mismo patrón que `loginApi.ts`.
import { httpClient } from '@/shared/api/httpClient'
import { isProblem } from '@/shared/api/problem'
import type { ChallengeRequestOutcome, ChallengeVerifyOutcome } from '../model/challengeOutcome'

/**
 * Solicita el código del reto telefónico. Siempre `accepted` ante un 202
 * (DEC-062: no enumeración — el servidor responde igual exista o no la
 * cuenta, esté o no el teléfono verificado). Un fallo de forma (400/422) no
 * debería ocurrir con un correo ya validado por el formulario de acceso, así
 * que se trata como `unexpected-error`, igual que cualquier otro estado no
 * documentado.
 */
export async function requestChallenge(email: string): Promise<ChallengeRequestOutcome> {
  try {
    const { error, response } = await httpClient.POST('/public/auth/challenge', {
      body: { email },
    })

    if (response.ok) {
      return { kind: 'accepted' }
    }

    const requestId = isProblem(error) ? error.requestId : undefined
    return { kind: 'unexpected-error', requestId }
  } catch {
    return { kind: 'network-error' }
  }
}

/**
 * Verifica el código de 6 dígitos. Mapea por `status` (nunca por `detail`):
 * 204 es éxito, 401 cubre código incorrecto/vencido/agotado/cuenta
 * inexistente de forma indistinguible (DEC-062), 400/422 es un defecto de
 * forma que la validación de cliente debería haber evitado.
 */
export async function verifyChallenge(
  email: string,
  code: string,
): Promise<ChallengeVerifyOutcome> {
  try {
    const { error, response } = await httpClient.POST('/public/auth/challenge/verify', {
      body: { email, code },
    })

    if (response.status === 204) {
      return { kind: 'verified' }
    }

    const requestId = isProblem(error) ? error.requestId : undefined

    switch (response.status) {
      case 401:
        return { kind: 'invalid-code' }
      case 400:
      case 422:
        return { kind: 'validation-error' }
      default:
        return { kind: 'unexpected-error', requestId }
    }
  } catch {
    return { kind: 'network-error' }
  }
}
