// Único punto del módulo `auth` que llama al cliente HTTP tipado
// (docs/03-desarrollo/estandar-frontend-vue.md §6: "un módulo define
// funciones API por intención"). Traduce la respuesta real de
// `POST /public/auth/login` a `LoginOutcome`: la página coordinadora y el
// formulario nunca ven `Problem`, `status` HTTP crudo ni cabeceras.
import { httpClient } from '@/shared/api/httpClient'
import { isProblem } from '@/shared/api/problem'
import type { LoginCredentials, LoginOutcome } from '../model/loginOutcome'

/**
 * Intenta iniciar sesión contra el API real. Mapea por `status` (nunca por
 * `detail`, trabajo requerido §8): 401 son credenciales inválidas o correo
 * inexistente (indistinguibles, CA-005-02/CA-010-02); 400/422 son un
 * defecto de forma que la validación de cliente debería haber evitado;
 * 429 es el bloqueo por umbral que HU-007 implementará por completo (esta
 * función solo interpreta la forma del contrato genérico `Problem`, sin
 * afirmar que el escalamiento de HU-007 existe); cualquier otro estado no
 *-2xx es inesperado. Un fallo de `fetch` en sí (sin respuesta HTTP) es
 * error de red, distinto de un 5xx del servidor.
 */
export async function login(credentials: LoginCredentials): Promise<LoginOutcome> {
  try {
    const { data, error, response } = await httpClient.POST('/public/auth/login', {
      body: credentials,
    })

    if (response.ok && data) {
      return { kind: 'success', expiresAt: data.expiresAt }
    }

    const requestId = isProblem(error) ? error.requestId : undefined

    switch (response.status) {
      case 401:
        return { kind: 'invalid-credentials' }
      case 400:
      case 422:
        return { kind: 'validation-error' }
      case 429: {
        const retryAfterHeader = response.headers.get('Retry-After')
        const parsedRetryAfter = retryAfterHeader ? Number(retryAfterHeader) : NaN
        return {
          kind: 'rate-limited',
          retryAfterSeconds: Number.isFinite(parsedRetryAfter) ? parsedRetryAfter : undefined,
        }
      }
      default:
        return { kind: 'unexpected-error', requestId }
    }
  } catch {
    // `fetch` en sí lanzó (red caída, DNS, CORS bloqueado): no hubo
    // respuesta HTTP que traducir, así que nunca llega a la rama anterior.
    return { kind: 'network-error' }
  }
}
