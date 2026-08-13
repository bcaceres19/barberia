// Package idempotency implementa la capacidad reutilizable de idempotencia
// que RN-IDE-01 exige para toda escritura crítica: interpretar y validar la
// cabecera Idempotency-Key, calcular la huella canónica del contenido, y
// coordinar con las funciones PostgreSQL ya aprobadas (DEC-043) que
// garantizan que un efecto se ejecuta como máximo una vez por
// (barbería, clave), incluso con concurrencia real.
//
// Este paquete es deliberadamente ajeno a HTTP: no importa net/http ni Chi.
// Un módulo de dominio futuro (por ejemplo booking) lo invoca desde su
// propio caso de uso, dentro de la MISMA transacción que
// database.DB.InTenantTx entrega a su callback — ver Coordinator para la
// razón exacta. La cabecera HTTP se interpreta en
// internal/platform/httpserver (ver KeyFromRequest), que traduce hacia y
// desde este paquete, igual que ya hace con apperr y Translate.
//
// Este paquete NO expone un endpoint ni un handler de producción: HU-004
// entrega el mecanismo, no una operación de negocio. Las pruebas de
// integración usan un handler que existe solo dentro del paquete de
// pruebas de httpserver.
package idempotency
