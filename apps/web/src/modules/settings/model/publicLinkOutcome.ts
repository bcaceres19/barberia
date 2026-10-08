// Resultado de leer el enlace público de reservas (issue #304, DEC-117), mismo
// criterio que `brandOutcome.ts`: no expone `Problem`, `status` HTTP crudo ni
// cabeceras. Un 401 real lo intercepta la coordinación única de
// `installSessionHandling`.
export type FetchPublicLinkOutcome =
  { kind: 'success'; slug: string } | { kind: 'network-error' } | { kind: 'unexpected-error' }
