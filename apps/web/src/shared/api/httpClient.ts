// Cliente HTTP tipado único de la aplicación (docs/03-desarrollo/
// estandar-frontend-vue.md §6: "todo acceso HTTP pasa por el cliente
// tipado de shared/api; los componentes no llaman fetch directamente").
//
// Herramienta elegida (DEC-037, "el mecanismo exacto ... se seleccionan al
// diseñar la primera entrega, siempre desde el bundle OpenAPI"):
// `openapi-typescript` + `openapi-fetch`.
//   - Necesidad: los tipos de request/response/error deben nacer del bundle
//     OpenAPI real (api/openapi/dist/openapi.yaml), sin DTO manuales
//     paralelos que puedan desviarse del contrato.
//   - Alternativas descartadas: un cliente generado de mayor superficie
//     (openapi-generator, orval con hooks React-only) añade codegen de
//     lógica de negocio y dependencias pesadas que este proyecto no
//     necesita (DEC-035: sin abstracciones sin necesidad demostrada);
//     escribir DTO e `fetch` a mano duplicaría el contrato y podría
//     desviarse de él en silencio.
//   - Mantenimiento: ambos paquetes son mantenidos activamente por
//     openapi-ts (antes drwpow/openapi-typescript), con adopción amplia.
//   - Licencia: MIT ambos.
//   - Superficie transitiva: openapi-typescript es solo de build (no viaja
//     al bundle de producción); openapi-fetch en runtime no trae
//     dependencias transitivas propias más allá de tipos.
//   - Reproducibilidad: `pnpm run generate:api` (ver package.json) genera
//     `generated/openapi.d.ts` desde el bundle ya lintado/generado por
//     `pnpm run openapi:bundle` en la raíz del repositorio. El archivo
//     generado se versiona (igual que un lockfile) y NUNCA se edita a
//     mano; un cambio de contrato se refleja regenerando este archivo.
import createClient from 'openapi-fetch'
import type { paths } from './generated/openapi.d.ts'

// Ruta relativa: coincide con `servers: - url: /api/v1` del contrato y con
// el proxy de desarrollo de vite.config.ts, que reproduce el mismo origen
// que un reverse proxy real usaría en despliegue (evita CORS y fricción de
// SameSite con la cookie de sesión HttpOnly).
const API_BASE_URL = '/api/v1'

/**
 * Cliente HTTP tipado compartido por toda la aplicación. `credentials:
 * 'include'` envía y recibe la cookie de sesión `barberia_session`
 * (HttpOnly): JavaScript nunca lee su valor, solo reacciona al resultado
 * HTTP (docs/03-desarrollo/estandar-frontend-vue.md §6).
 */
export const httpClient = createClient<paths>({
  baseUrl: API_BASE_URL,
  credentials: 'include',
})
