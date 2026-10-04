// Avisos emergentes de acceso y recuperación (DEC-095): una conexión
// perdida o un error inesperado no pertenecen a ningún campo, así que se
// avisan como toast con su acción en vez de una alerta fija bajo el
// formulario. El rechazo de credenciales en `/acceso` también es un toast
// (DEC-108): aparece arriba a la derecha y se retira solo. Los errores de
// código o política de recuperación siguen en línea, junto al campo que los
// origina.
import { useToast } from '@/shared/composables'

/** Conexión perdida: el formulario conserva lo escrito y "Reintentar" repite el envío. */
export function notifyConnectionLost(retry: () => void): void {
  useToast().warning('No pudimos conectar', {
    detail: 'Revisa tu conexión e inténtalo de nuevo.',
    action: { label: 'Reintentar', icon: 'retry', run: retry },
  })
}

/** Error inesperado: mensaje genérico y, si el contrato lo dio, el código seguro para soporte. */
export function notifyUnexpectedError(requestId?: string): void {
  useToast().error('Ocurrió un error inesperado', {
    detail: 'Inténtalo de nuevo en unos segundos.',
    reference: requestId,
  })
}

/** Credenciales rechazadas: el mismo texto para correo inexistente y contraseña incorrecta (CA-005-02, CA-010-02). */
export function notifyInvalidCredentials(): void {
  useToast().error('No pudimos iniciar tu sesión', {
    detail: 'Revisa tu correo y contraseña e inténtalo de nuevo.',
  })
}

/** El servidor rechazó la forma del correo o la contraseña. */
export function notifyInvalidLoginFormat(): void {
  useToast().error('Revisa los datos ingresados', {
    detail: 'El correo o la contraseña no tienen un formato válido.',
  })
}
