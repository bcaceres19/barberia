// Avisos emergentes de los fallos de sistema en acceso y recuperación
// (DEC-095): una conexión perdida o un error inesperado no pertenecen a
// ningún campo, así que se avisan como toast con su acción en vez de una
// alerta fija bajo el formulario. Los errores de credencial, código o
// política siguen en línea, junto al campo que los origina.
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
