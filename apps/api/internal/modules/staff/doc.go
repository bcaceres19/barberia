// Package staff administra los barberos y su pertenencia a la barbería,
// según docs/03-desarrollo/estandar-backend-go.md.
//
// HU-021 entrega el primer alcance real: consultar, listar, registrar y
// renombrar barberos (registro/listado/lectura/renombrado). domain.go,
// errors.go y ports.go definen el núcleo (sin Chi, net/http, pgx ni
// internal/platform/database, CA-002-06); postgres/ y httpapi/ traducen
// ese núcleo hacia PostgreSQL y HTTP respectivamente. Deliberadamente fuera
// de alcance (DEC-047): borrado, desactivación, orden manual, credenciales,
// roles, servicios, horario y agenda; una historia futura los agrega si
// llega a hacer falta.
package staff
