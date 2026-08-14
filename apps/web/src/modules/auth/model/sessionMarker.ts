// Marcador local mínimo de sesión para el guard de `/panel` (DEC-056).
//
// Por qué existe: la cookie de sesión es HttpOnly (JavaScript no puede
// leerla, DEC-050) y todavía no existe ningún endpoint privado real contra
// el que verificar la sesión (`DP-SEG-08`: el primer endpoint privado real
// es `POST /api/v1/private/auth/logout` de HU-006, fuera de alcance de
// HU-010). Sin ninguna de las dos vías, el guard de esta historia no puede
// verificar la sesión contra el servidor todavía.
//
// Qué se guarda y por qué es seguro: únicamente `expiresAt`, el mismo
// valor no sensible que `LoginResponse` ya expone en el cuerpo 200 de
// `POST /public/auth/login` (api/openapi/components/schemas/LoginResponse.yaml,
// "el estado de la sesión vive en la cookie, no en este cuerpo"). No es el
// token, no es la contraseña, y `sessionStorage` no sobrevive al cierre de
// la pestaña.
//
// Qué NO es esto: no es una verificación de seguridad. La autoridad real
// de la sesión sigue siendo, en cada solicitud, la cookie HttpOnly que el
// navegador adjunta automáticamente. Este marcador solo evita que, dentro
// de la misma pestaña, alguien sin haber iniciado sesión recientemente
// vea el marcador de posición de `/panel`; un guard que valide la sesión
// contra un endpoint privado real le corresponde a una historia posterior
// (`HU-012` generaliza y endurece este guard, según su fila en
// docs/02-requisitos/historias-usuario.md).
const STORAGE_KEY = 'barberia:session-expires-at'

function readStorage(): Storage | null {
  try {
    return window.sessionStorage
  } catch {
    // Modo privado estricto, cuota agotada o entorno sin `sessionStorage`:
    // se trata como si no hubiera sesión recordada, nunca como error fatal.
    return null
  }
}

/** Se llama exactamente una vez, justo después de un `LoginOutcome`
 * `success` (CA-010-01). */
export function rememberSessionUntil(expiresAt: string): void {
  const storage = readStorage()
  storage?.setItem(STORAGE_KEY, expiresAt)
}

/** Usado por el guard de `/panel`. `true` solo si hay un `expiresAt`
 * recordado y todavía no venció según el reloj del navegador. */
export function hasRememberedSession(): boolean {
  const storage = readStorage()
  const value = storage?.getItem(STORAGE_KEY)
  if (!value) return false
  const expiresAt = Date.parse(value)
  return Number.isFinite(expiresAt) && expiresAt > Date.now()
}

/** Disponible para un cierre de sesión futuro (HU-006/HU-012); HU-010 no
 * lo invoca todavía porque no construye logout. */
export function forgetSession(): void {
  const storage = readStorage()
  storage?.removeItem(STORAGE_KEY)
}
