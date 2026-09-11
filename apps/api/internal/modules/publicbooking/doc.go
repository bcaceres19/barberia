// Package publicbooking resuelve el enlace público de reservas de una
// barbería (HU-090, F-PUB-01), según docs/03-desarrollo/estandar-backend-go.md.
//
// [Service] no conoce Chi, net/http, JSON, pgx ni internal/platform/database
// (CA-002-06): recibe sus dependencias por constructor y solo habla con
// [Repository], igual que shops.Service. La persistencia vive en postgres/,
// el adaptador HTTP en httpapi/.
//
// Deliberadamente NO importa el paquete shops: aunque ambos leen hoy los
// mismos cuatro campos de barbershop, son contratos distintos con ciclos de
// vida distintos (uno privado y autenticado, HU-020; otro público y sin
// sesión, HU-090). Cada uno mantiene su propia consulta mínima en su propio
// adaptador postgres/ para que un campo nuevo agregado al contrato privado
// de HU-020 nunca se filtre por accidente a la respuesta pública.
package publicbooking
