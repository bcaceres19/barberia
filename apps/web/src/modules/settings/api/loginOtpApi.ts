// Cliente de `GET`/`PATCH /private/settings/login-otp` (DEC-139): el canal del
// código del reto adicional de acceso. El cuerpo es solo el canal; el servidor
// deriva los destinos de cada cuenta y nunca los recibe.
import { httpClient } from '@/shared/api/httpClient'
import {
  isLoginOtpChannel,
  type FetchLoginOtpOutcome,
  type LoginOtpChannel,
  type SaveLoginOtpOutcome,
} from '../model/loginOtpOutcome'

export async function fetchLoginOtpChannel(): Promise<FetchLoginOtpOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/settings/login-otp')
    if (response.ok && data && isLoginOtpChannel(data.channel)) {
      return { kind: 'success', channel: data.channel }
    }
    return { kind: 'unexpected-error' }
  } catch {
    return { kind: 'network-error' }
  }
}

export async function saveLoginOtpChannel(channel: LoginOtpChannel): Promise<SaveLoginOtpOutcome> {
  try {
    const { data, response } = await httpClient.PATCH('/private/settings/login-otp', {
      body: { channel },
    })
    if (response.ok && data && isLoginOtpChannel(data.channel)) {
      return { kind: 'success', channel: data.channel }
    }
    return { kind: 'unexpected-error' }
  } catch {
    return { kind: 'network-error' }
  }
}
