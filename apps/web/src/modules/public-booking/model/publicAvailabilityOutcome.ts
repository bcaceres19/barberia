// Modelo de dominio de la exploración pública de fechas y horarios
// (HU-095). `PublicAvailabilitySlot`/`PublicAvailabilityWindow` son la
// forma que la página consume; `ListPublicAvailabilityOutcome` es lo único
// que `api/listPublicAvailabilityApi.ts` entrega. Ninguno de los dos
// expone la forma RFC 9457 (`Problem`) del contrato (docs/03-desarrollo/
// estandar-frontend-vue.md §6), mismo criterio que
// model/publicBarberListOutcome.ts.

/** Inicio público válido, ya resuelto por el motor de disponibilidad de
 * HU-094 (CA-094-01 a CA-094-03). A qué día civil pertenece se decide al
 * agrupar (ver `availabilityCalendar.ts`), nunca aquí. */
export interface PublicAvailabilitySlot {
  startsAt: string
}

/** Ventana pública de disponibilidad ya resuelta contra el servidor para un
 * servicio y un barbero concretos: los inicios válidos en orden estable
 * (cronológico), la duración planificada del servicio (DEC-002) y la zona
 * horaria de la barbería (RN-DIS-07, nunca la del dispositivo) con las que
 * interpretarlos, y el paso de rejilla vigente (RN-DIS-06, DEC-083). El
 * servidor no expone los límites de anticipación/ventana por separado de
 * los inicios que produce: la ventana navegable en pantalla se deriva
 * enteramente de `slots` (ver `groupSlotsByCivilDate`). */
export interface PublicAvailabilityWindow {
  slots: PublicAvailabilitySlot[]
  durationMinutes: number
  timezone: string
  slotGridMinutes: number
}

/** Resultado discriminado de consultar la disponibilidad real contra el
 * API. `not-found` cubre las mismas tres causas indistinguibles de
 * CA-090-02 (slug con forma inválida, desconocido, no publicable) que el
 * resto del módulo; un `serviceId`/`barberId` ajeno, inexistente, inactivo
 * o sin asignación vigente NUNCA produce `not-found`: llega como `success`
 * con `slots: []` (mismo criterio que CA-092-03), indistinguible de "sin
 * disponibilidad en la ventana pública vigente". `network-error` y
 * `unexpected-error` son las mismas fronteras defensivas del resto de la
 * aplicación. */
export type ListPublicAvailabilityOutcome =
  | { kind: 'success'; availability: PublicAvailabilityWindow }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error'; requestId?: string }
