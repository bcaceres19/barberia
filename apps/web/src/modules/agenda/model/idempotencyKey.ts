// Genera la clave de idempotencia de UN intento lógico de alta (RN-IDE-01,
// DEC-043, HU-004), mismo criterio y misma forma que
// schedules/model/idempotencyKey.ts: duplicada aquí porque un módulo no
// importa archivos internos de otro (docs/03-desarrollo/estandar-frontend-vue.md
// §3).
export function newIdempotencyKey(): string {
  return crypto.randomUUID()
}
