// Resultados discriminados de las operaciones de HU-022, tal como los
// recibe la página coordinadora. Ninguno expone `Problem`, `status` HTTP
// crudo ni cabeceras (docs/03-desarrollo/estandar-frontend-vue.md §6),
// mismo criterio que staffOutcome.ts. Un 401 real de estas rutas ya lo
// intercepta la coordinación única de `installSessionHandling`; todos
// igual declaran `unexpected-error` como frontera defensiva por si esa
// solicitud concreta llegara a resolverse antes de que la redirección
// navegue.
import type { Service, ServicePage } from './service'

export type FetchServicesOutcome =
  { kind: 'success'; page: ServicePage } | { kind: 'network-error' } | { kind: 'unexpected-error' }

export type CreateServiceOutcome =
  | { kind: 'success'; service: Service }
  | { kind: 'validation-error' }
  // DEC-067: el nombre ya lo usa otro servicio activo de la misma barbería.
  // Distinto de idempotency-conflict: no depende de la clave de
  // idempotencia ni de un reintento, así que reintentar con la MISMA clave
  // y OTRO nombre sí puede tener éxito.
  | { kind: 'name-conflict' }
  // RN-IDE-01: la misma clave de idempotencia ya se usó con un contenido
  // distinto, o la operación con esa clave sigue en curso (DEC-043). No es
  // un error de validación de campo: el formulario no cambió, es un
  // reintento el que debe generar una clave nueva.
  | { kind: 'idempotency-conflict' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type UpdateServiceOutcome =
  | { kind: 'success'; service: Service }
  | { kind: 'validation-error' }
  | { kind: 'name-conflict' }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

// --- HU-024: ciclo de vida --------------------------------------------------

export type PreviewDeactivationOutcome =
  | { kind: 'success'; affectedAppointments: number }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

// transition-conflict (RN-IDE-01 con una clave nueva sobre una transición
// que ya no aplica, CA-024-06): el servicio ya estaba en el estado destino.
// Distinto de idempotency-conflict (misma clave, otro contenido/operación,
// o en curso, DEC-043): ese caso nunca debería alcanzar la interfaz porque
// cada intento lógico genera su propia clave nueva; se conserva como
// frontera defensiva.
export type DeactivateServiceOutcome =
  | { kind: 'success'; service: Service; affectedAppointments: number }
  | { kind: 'not-found' }
  | { kind: 'transition-conflict' }
  | { kind: 'idempotency-conflict' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type ReactivateServiceOutcome =
  | { kind: 'success'; service: Service }
  | { kind: 'not-found' }
  | { kind: 'transition-conflict' }
  | { kind: 'idempotency-conflict' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }
