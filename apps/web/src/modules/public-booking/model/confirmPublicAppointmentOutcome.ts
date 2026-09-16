// Modelo de dominio de la confirmación pública concurrente (HU-097).
// `ConfirmedPublicAppointment`/`PublicAppointmentAlternative` son la forma
// que la página consume; `ConfirmPublicAppointmentOutcome` es lo único que
// `api/confirmPublicAppointmentApi.ts` entrega. Ninguno de los dos expone
// la forma RFC 9457 (`Problem`) del contrato, mismo criterio que
// model/publicAvailabilityOutcome.ts.

/** Turno público ya confirmado (CA-097-01): proyección mínima, nunca un
 * identificador interno (RN-DAT-02) porque el cliente público no consulta
 * esta cita por id -HU-098 la resolverá por `accessToken`. */
export interface ConfirmedPublicAppointment {
  attendeeName: string
  barbershopName: string
  serviceName: string
  durationMinutes: number
  priceAmount: string
  currency: string
  startsAt: string
  endsAt: string
  timezone: string
  /** Credencial de acceso al turno EN CLARO (DEC-089): esta es la ÚNICA vez
   * que aparece en cualquier respuesta. */
  accessToken: string
  customerNote: string | null
}

/** Franja alternativa cronológicamente cercana a la elegida (RN-CON-05,
 * DEC-090), mismo barbero y mismo servicio. */
export interface PublicAppointmentAlternative {
  startsAt: string
}

/** Resultado discriminado de confirmar públicamente el turno elegido
 * contra el API real (HU-097). `not-found` cubre la misma familia
 * indistinguible de causas que el resto del módulo (CA-090-02, CA-092-03):
 * barbería no publicable, o servicio/barbero ajeno, inexistente, inactivo
 * o sin asignación vigente -a diferencia de la disponibilidad de solo
 * lectura, esta escritura no tiene un "vacío" que devolver, así que la
 * causa uniforme aquí SÍ produce un error, nunca un éxito con datos
 * vacíos. `schedule-conflict` es RN-CON-05/DEC-090: la franja ya no está
 * disponible, con hasta 3 alternativas cronológicamente cercanas.
 * `validation-error` cubre un dato de identidad inválido (HU-096, HU-097
 * revalida en servidor). `idempotency-conflict` cubre RN-IDE-01 (reintento
 * con contenido distinto, o clave en curso). */
export type ConfirmPublicAppointmentOutcome =
  | { kind: 'success'; appointment: ConfirmedPublicAppointment }
  | { kind: 'not-found' }
  | { kind: 'schedule-conflict'; alternatives: PublicAppointmentAlternative[] }
  | { kind: 'validation-error'; detail: string }
  | { kind: 'idempotency-conflict' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error'; requestId?: string }
