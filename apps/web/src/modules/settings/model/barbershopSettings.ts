// Espejo tipado del cuerpo real del contrato (CA-020-07): exactamente los
// cuatro campos autorizados. contactEmail/contactPhone son `string | null`
// porque el servidor los devuelve como `null` explícito cuando la barbería
// no tiene ese contacto configurado (nunca cadena vacía, CA-020-06).
export interface BarbershopSettings {
  name: string
  timezone: string
  contactEmail: string | null
  contactPhone: string | null
}

/** Forma editable del formulario: los cuatro campos como texto, contacto
 * vacío representando "sin contacto" en el alambre (igual convención que
 * el request real, ver UpdateBarbershopSettingsRequest.yaml). */
export interface BarbershopSettingsFormValues {
  name: string
  timezone: string
  contactEmail: string
  contactPhone: string
}

export function toFormValues(settings: BarbershopSettings): BarbershopSettingsFormValues {
  return {
    name: settings.name,
    timezone: settings.timezone,
    contactEmail: settings.contactEmail ?? '',
    contactPhone: settings.contactPhone ?? '',
  }
}
