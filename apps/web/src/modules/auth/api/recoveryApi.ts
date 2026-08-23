// Único punto del módulo `auth` que llama al cliente HTTP tipado para la
// recuperación de acceso (HU-008, DEC-063–DEC-066), mismo patrón que
// `loginApi.ts`/`challengeApi.ts`. Traduce cada respuesta real a la unión
// discriminada correspondiente de `model/recoveryOutcome.ts`; mapea por
// `status` (nunca por `detail`), igual que el resto del módulo.
import { httpClient } from '@/shared/api/httpClient'
import { isProblem } from '@/shared/api/problem'
import type {
  RecoveryRequestOutcome,
  RecoveryResetPasswordOutcome,
  RecoveryVerifyOutcome,
} from '../model/recoveryOutcome'

/**
 * Solicita el código de recuperación. Siempre `accepted` ante un 202
 * (DEC-065: no enumeración — el cuerpo genérico no revela si la cuenta
 * existe). 400/422 no deberían ocurrir con un correo ya validado en
 * cliente, así que se tratan como `unexpected-error`, igual que cualquier
 * otro estado no documentado.
 */
export async function requestRecovery(email: string): Promise<RecoveryRequestOutcome> {
  try {
    const { error, response } = await httpClient.POST('/public/auth/recovery/request', {
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
 * Verifica el código de 6 dígitos. 200 es éxito y trae el token de
 * reinicio y el destino enmascarado; 401 cubre código incorrecto,
 * vencido, agotado o cuenta inexistente de forma indistinguible
 * (DEC-064/DEC-065); 400/422 es un defecto de forma que la validación de
 * cliente debería haber evitado.
 */
export async function verifyRecovery(email: string, code: string): Promise<RecoveryVerifyOutcome> {
  try {
    const { data, error, response } = await httpClient.POST('/public/auth/recovery/verify', {
      body: { email, code },
    })

    if (response.ok && data) {
      return {
        kind: 'verified',
        resetToken: data.resetToken,
        maskedPhone: data.maskedPhone,
        maskedEmail: data.maskedEmail,
      }
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

/**
 * Establece la contraseña nueva con el token de reinicio devuelto por la
 * verificación. 204 es éxito (contraseña actualizada, todas las sesiones
 * revocadas, CA-008-05); 401 cubre token desconocido, vencido, consumido o
 * de otra cuenta, siempre uniforme; 422 cubre un incumplimiento de la
 * política de `DEC-063` — el backend no distingue cuál regla exacta falló
 * (mismo `code` de validación genérico), así que este resultado tampoco lo
 * hace.
 */
export async function resetRecoveryPassword(
  email: string,
  resetToken: string,
  newPassword: string,
): Promise<RecoveryResetPasswordOutcome> {
  try {
    const { error, response } = await httpClient.POST('/public/auth/recovery/reset-password', {
      body: { email, resetToken, newPassword },
    })

    if (response.status === 204) {
      return { kind: 'success' }
    }

    const requestId = isProblem(error) ? error.requestId : undefined

    switch (response.status) {
      case 401:
        return { kind: 'invalid-token' }
      case 422:
        return { kind: 'policy-violation' }
      case 400:
        return { kind: 'validation-error' }
      default:
        return { kind: 'unexpected-error', requestId }
    }
  } catch {
    return { kind: 'network-error' }
  }
}
