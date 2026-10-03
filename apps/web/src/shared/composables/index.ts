// Composables reutilizados por más de un módulo. Una función pura de
// formato o validación vive en shared/model o shared/validation, no aquí,
// según docs/03-desarrollo/estandar-frontend-vue.md.
export { useToast } from './useToast'
export type { ToastOptions, UseToast } from './useToast'
export { DEFAULT_MIN_HOLD_MS, PAGE_MIN_HOLD_MS, useMinHoldLoading } from './useMinHoldLoading'
export type { MinHoldLoading } from './useMinHoldLoading'
