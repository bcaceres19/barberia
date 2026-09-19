// API pública para emitir avisos emergentes desde cualquier módulo. Envuelve
// la cola de `shared/model/toastStore` para que las pantallas no dependan de
// su forma interna: una pantalla dice qué ocurrió, no cómo se muestra.
import {
  dismissAllToasts,
  dismissToast,
  pushToast,
  type ToastAction,
  type ToastInput,
  type ToastVariant,
} from '@/shared/model/toastStore'

export type { ToastAction, ToastInput, ToastVariant }

export interface ToastOptions {
  detail?: string
  reference?: string
  action?: ToastAction
}

export interface UseToast {
  /** Aviso con la variante indicada; devuelve su identificador. */
  notify: (input: ToastInput) => number
  success: (title: string, options?: ToastOptions) => number
  info: (title: string, options?: ToastOptions) => number
  warning: (title: string, options?: ToastOptions) => number
  error: (title: string, options?: ToastOptions) => number
  dismiss: (id: number) => void
  dismissAll: () => void
}

function withVariant(variant: ToastVariant) {
  return (title: string, options: ToastOptions = {}): number =>
    pushToast({ variant, title, ...options })
}

export function useToast(): UseToast {
  return {
    notify: pushToast,
    success: withVariant('success'),
    info: withVariant('info'),
    warning: withVariant('warning'),
    error: withVariant('danger'),
    dismiss: dismissToast,
    dismissAll: dismissAllToasts,
  }
}
