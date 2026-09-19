// Destino elegido en el paso 1 de la recuperación de acceso (DEC-092,
// DEC-093): el canal por el que la persona quiere recibir el código y el
// valor que escribió para ese canal. Viaja sin cambios por los tres pasos
// (DP-SEG-15). El valor solo identifica la cuenta: el servidor entrega el
// código al contacto verificado que ya tiene, nunca a lo escrito aquí.

export type RecoveryChannel = 'email' | 'whatsapp'

export interface RecoveryTarget {
  channel: RecoveryChannel
  /** Correo o número de WhatsApp, según `channel`, ya recortado y, para el
   * teléfono, sin separadores de presentación. */
  value: string
}

/** Cuerpo común de los tres pasos, cerrado por canal (`RecoveryRequestRequest`
 * y equivalentes del contrato). */
export type RecoveryTargetBody =
  { channel: 'email'; email: string } | { channel: 'whatsapp'; phone: string }

export function toRecoveryTargetBody(target: RecoveryTarget): RecoveryTargetBody {
  return target.channel === 'email'
    ? { channel: 'email', email: target.value }
    : { channel: 'whatsapp', phone: target.value }
}

/** Correo de la cuenta cuando el canal es correo; vacío con WhatsApp, porque
 * el cliente no conoce el correo hasta que el servidor lo enmascara. */
export function recoveryEmailOf(target: RecoveryTarget): string {
  return target.channel === 'email' ? target.value : ''
}
