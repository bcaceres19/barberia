// Package reviews implementa HU-100 (DEC-128): la reseña interna que un cliente
// deja sobre el barbero que lo atendió, una sola por turno y solo si el turno
// está `completed`, y su moderación reversible desde el panel.
//
// El cliente se identifica únicamente con el token opaco del enlace de su turno
// (DEC-089), el mismo de customeraccess; este paquete no lo importa: repite su
// propia forma mínima (hash SHA-256 del token) como el resto de módulos
// independientes. Las estrellas del barbero (`rating`, `ratingCount`) no viven
// aquí: staff y publicbooking las calculan en su propia consulta sobre
// barber_review. Fuera de alcance: reseñas del negocio (DEC-129), respuesta
// del barbero, edición de una reseña enviada y cola de aprobación.
package reviews
