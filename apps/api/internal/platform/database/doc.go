// Package database aloja la apertura del pool de PostgreSQL y el
// establecimiento del contexto de barbería (app.barbershop_id) requerido por
// RLS en cada transacción, según docs/05-backend/base-datos.md.
//
// El pool NO expone Query, QueryRow, Exec ni Begin como métodos exportados. La
// única forma de ejecutar trabajo de negocio es InTenantTx, que recibe el
// identificador de barbería y entrega el ejecutor SOLO dentro del alcance de
// una transacción ya configurada. Esto hace IMPOSIBLE omitir el contexto.
//
// Justificación de la dependencia (estandar-backend-go.md §5.21):
//   - Necesidad: driver nativo de PostgreSQL para pool con hooks de adquisición
//     y liberación, tipos uuid/timestamptz/numeric/rangos sin conversión, y
//     set_config parametrizable para fijar app.barbershop_id sin inyección SQL.
//   - Mantenimiento: jackc/pgx v5 es el driver estándar de facto, desarrollo
//     activo, versión semver estable.
//   - Licencia: MIT.
//   - Superficie transitiva: ninguna dependencia externa más allá de la librería
//     estándar de Go y golang.org/x/* internos.
//   - Seguridad: protocolo nativo, sin capas de abstracción que oculten
//     comportamiento; consultas SQL visibles y revisables. Atlas gobierna
//     migraciones; NO se ejecutan al arrancar (CA-001-07).
package database
