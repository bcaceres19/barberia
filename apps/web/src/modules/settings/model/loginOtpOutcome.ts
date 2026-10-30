// Canal del código del reto de acceso (DEC-139) y resultados de leerlo y
// guardarlo, mismo criterio que `brandOutcome.ts`: sin `Problem`, `status`
// HTTP crudo ni cabeceras. Un 401 real lo intercepta la coordinación única de
// `installSessionHandling`.
export const LOGIN_OTP_CHANNELS = ['email', 'whatsapp', 'both'] as const
export type LoginOtpChannel = (typeof LOGIN_OTP_CHANNELS)[number]

export function isLoginOtpChannel(value: unknown): value is LoginOtpChannel {
  return typeof value === 'string' && (LOGIN_OTP_CHANNELS as readonly string[]).includes(value)
}

export type FetchLoginOtpOutcome =
  | { kind: 'success'; channel: LoginOtpChannel }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type SaveLoginOtpOutcome =
  | { kind: 'success'; channel: LoginOtpChannel }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }
