// Genera la clave de idempotencia de UN intento lógico de alta (RN-IDE-01,
// DEC-043, HU-004): un mismo intento (por ejemplo, un reintento de red tras
// un timeout mientras el diálogo sigue abierto con los mismos datos)
// reutiliza la MISMA clave, para que el servidor lo trate como la misma
// solicitud en vez de crear un segundo servicio. Un intento lógico nuevo
// (abrir el formulario de nuevo, o el siguiente envío tras un éxito o tras
// cerrar y reabrir el diálogo) exige una clave nueva.
//
// Duplicado deliberado de staff/model/idempotencyKey.ts (docs/03-desarrollo/
// estandar-backend-go.md §5.23, mismo criterio aplicado en frontend:
// duplicación pequeña y clara preferible a que un módulo importe archivos
// internos de otro, docs/03-desarrollo/estandar-frontend-vue.md §3).
//
// crypto.randomUUID() (Web Crypto, disponible en todo navegador que este
// proyecto soporta) ya cumple el patrón que el contrato exige
// (^[A-Za-z0-9._-]{1,255}$): un UUID solo usa dígitos, letras minúsculas y
// guiones.
export function newIdempotencyKey(): string {
  return crypto.randomUUID()
}
