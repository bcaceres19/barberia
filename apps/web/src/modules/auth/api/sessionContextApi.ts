// Único punto del módulo `auth` que llama a GET /private/auth/session
// (docs/03-desarrollo/estandar-frontend-vue.md §6, mismo patrón que
// api/loginApi.ts): traduce la respuesta real del contrato a
// SessionContextOutcome. La página/guard nunca ven `Problem`, `status` HTTP
// crudo ni cabeceras.
import { httpClient } from '@/shared/api/httpClient'
import type { SessionContextOutcome } from '../model/sessionContextOutcome'

export async function fetchSessionContext(): Promise<SessionContextOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/auth/session')

    if (response.ok && data) {
      return {
        kind: 'authenticated',
        barbershopId: data.barbershop.id,
        barbershopName: data.barbershop.name,
        expiresAt: data.expiresAt,
      }
    }

    if (response.status === 401) {
      return { kind: 'unauthenticated' }
    }

    return { kind: 'unexpected-error' }
  } catch {
    // `fetch` en sí lanzó (red caída, DNS, CORS bloqueado): no hubo
    // respuesta HTTP que traducir, distinto de un 401/5xx real del
    // servidor (mismo criterio que loginApi.ts).
    return { kind: 'network-error' }
  }
}
