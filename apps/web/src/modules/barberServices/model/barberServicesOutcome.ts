// Resultados discriminados de las operaciones de HU-023, tal como los recibe
// la página coordinadora. Ninguno expone `Problem`, `status` HTTP crudo ni
// cabeceras (docs/03-desarrollo/estandar-frontend-vue.md §6), mismo criterio
// que staffOutcome.ts/catalogOutcome.ts. Un 401 real de estas rutas ya lo
// intercepta la coordinación única de `installSessionHandling`; todos igual
// declaran `unexpected-error` como frontera defensiva.
import type { Assignment, AssignmentPage, BarberSummary, ServiceSummary } from './assignment'

export type FetchBarberSummariesOutcome =
  | { kind: 'success'; items: BarberSummary[] }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type FetchServiceSummariesOutcome =
  | { kind: 'success'; items: ServiceSummary[] }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type FetchAssignmentsOutcome =
  | { kind: 'success'; page: AssignmentPage }
  // El propio barbero seleccionado ya no existe/es de otra barbería
  // (RN-TEN-01): la pantalla lo trata igual que una carga fallida.
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type AssignServiceOutcome =
  | { kind: 'success'; assignment: Assignment }
  // CA-023-04: barbero o servicio inexistente/ajeno, indistinguibles.
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type UnassignServiceOutcome =
  | { kind: 'success' }
  | { kind: 'not-found' }
  // DEC-068/CA-023-05: retirar la última asignación activa de un servicio
  // activo se rechaza; el estado del checkbox no cambia.
  | { kind: 'last-active-conflict' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }
