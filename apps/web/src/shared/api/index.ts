// Cliente HTTP tipado y contratos derivados del bundle OpenAPI. Todo
// acceso HTTP de un módulo pasa por aquí; ningún componente llama `fetch`
// directamente (docs/03-desarrollo/estandar-frontend-vue.md §6). Los tipos
// nacen de `generated/openapi.d.ts` (nunca editado a mano, ver
// `httpClient.ts` para el comando reproducible de generación).
export { httpClient } from './httpClient'
export { isProblem, type Problem } from './problem'
