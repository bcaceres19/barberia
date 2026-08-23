// Package shops administra la barbería, su configuración y su zona horaria,
// según docs/03-desarrollo/estandar-backend-go.md.
//
// HU-020 entrega la primera funcionalidad: consultar y actualizar el
// nombre, la zona horaria IANA y el contacto opcional (correo, teléfono) de
// la barbería activa. [Service] no conoce Chi, net/http, JSON, pgx ni
// internal/platform/database (CA-002-06): recibe sus dependencias por
// constructor y solo habla con [Repository], igual que auth.LoginService.
// La persistencia vive en postgres/, el adaptador HTTP en httpapi/.
package shops
