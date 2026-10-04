// Carga la marca de la barbería activa en el estado compartido (DEC-110).
// `app` la invoca al autenticarse la sesión; un fallo no bloquea el panel: la
// copia guardada en el dispositivo (o los valores iniciales) sigue pintándose.
import { setBrand } from '@/shared/model'
import { fetchBrand } from '../api/brandApi'

let requestToken = 0

export async function loadWorkspaceBrand(): Promise<void> {
  const token = ++requestToken
  const outcome = await fetchBrand()
  // Una respuesta obsoleta (otra sesión inició mientras esta volaba) no
  // sobrescribe la marca de la barbería vigente.
  if (token !== requestToken) return
  if (outcome.kind === 'success') setBrand(outcome.brand)
}
