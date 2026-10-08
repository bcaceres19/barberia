// URL completa del enlace público de reservas (issue #304, DEC-117). El API
// entrega solo el slug, porque no conoce el dominio del frontend; el origen sale
// del navegador, así que en cada ambiente el enlace apunta a donde se está
// usando la app. La forma `/reservar/:slug` es la entrada pública de
// `public-booking` (HU-090): una prueba la compara con su tabla de rutas real
// para que no diverjan.
const PUBLIC_BOOKING_PATH = '/reservar'

export function buildPublicLinkUrl(origin: string, slug: string): string {
  return `${origin.replace(/\/+$/, '')}${PUBLIC_BOOKING_PATH}/${encodeURIComponent(slug)}`
}
