// Único punto del módulo `settings` que llama al cliente HTTP tipado
// (docs/03-desarrollo/estandar-frontend-vue.md §6). Traduce las respuestas
// reales de GET/PATCH /private/settings/barbershop a los outcomes
// discriminados que la página consume; ningún componente ve `Problem`,
// `status` HTTP crudo ni cabeceras.
import { httpClient } from '@/shared/api/httpClient'
import type { BarbershopSettings, BarbershopSettingsFormValues } from '../model/barbershopSettings'
import type {
  FetchBarbershopSettingsOutcome,
  SaveBarbershopSettingsOutcome,
} from '../model/settingsOutcome'

export async function fetchBarbershopSettings(): Promise<FetchBarbershopSettingsOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/settings/barbershop')

    if (response.ok && data) {
      return { kind: 'success', settings: toSettings(data) }
    }

    // 401 lo intercepta la coordinación única de installSessionHandling
    // (redirige a acceso); 404 es defensivo (nunca debería ocurrir para un
    // principal autenticado real). Ambos, junto con cualquier otro estado
    // no-2xx, se tratan aquí como un error genérico recuperable.
    return { kind: 'unexpected-error' }
  } catch {
    // `fetch` en sí lanzó (red caída, DNS, CORS bloqueado): no hubo
    // respuesta HTTP que traducir (mismo criterio que loginApi.ts).
    return { kind: 'network-error' }
  }
}

export async function saveBarbershopSettings(
  values: BarbershopSettingsFormValues,
): Promise<SaveBarbershopSettingsOutcome> {
  try {
    const { data, response } = await httpClient.PATCH('/private/settings/barbershop', {
      body: {
        name: values.name,
        timezone: values.timezone,
        contactEmail: values.contactEmail,
        contactPhone: values.contactPhone,
      },
    })

    if (response.ok && data) {
      return { kind: 'success', settings: toSettings(data) }
    }

    switch (response.status) {
      case 400:
      case 422:
        return { kind: 'validation-error' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

function toSettings(data: {
  name: string
  timezone: string
  contactEmail: string | null
  contactPhone: string | null
}): BarbershopSettings {
  return {
    name: data.name,
    timezone: data.timezone,
    contactEmail: data.contactEmail,
    contactPhone: data.contactPhone,
  }
}
