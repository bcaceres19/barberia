// Package catalog administra los servicios ofrecidos por la barbería, su
// duración planificada y su precio informativo, según
// docs/03-desarrollo/estandar-backend-go.md.
//
// HU-022 entrega el primer alcance real: listar, consultar, crear y editar
// servicios del catálogo básico (nombre, descripción opcional, duración en
// minutos, precio en COP). domain.go, errors.go y ports.go definen el
// núcleo (sin Chi, net/http, pgx ni internal/platform/database, CA-002-06);
// postgres/ y httpapi/ traducen ese núcleo hacia PostgreSQL y HTTP
// respectivamente. Deliberadamente fuera de alcance: asignar servicios a
// barberos (HU-023), desactivar/reactivar/eliminar servicios (HU-024),
// horario, disponibilidad, citas y cualquier propagación de un cambio de
// catálogo hacia una cita existente (RN-SER-04, DEC-004).
package catalog
