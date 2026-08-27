// Package booking es el núcleo persistente de citas (HU-060), según
// docs/03-desarrollo/estandar-backend-go.md.
//
// HU-060 entrega únicamente la base persistente y la primitiva
// transaccional interna: tipos de dominio cerrados para estado, origen,
// intervalo y snapshot de servicio, y una operación CreateInternal capaz de
// guardar cliente, cita `confirmed` e historial inicial en una sola
// transacción tenant-aware. domain.go, errors.go y ports.go definen el
// núcleo (sin Chi, net/http, pgx ni internal/platform/database, CA-002-06);
// postgres/ traduce ese núcleo hacia PostgreSQL. Deliberadamente fuera de
// alcance: cualquier endpoint HTTP, formulario Vue o pantalla de agenda; la
// política de creación manual completa, incluida la reconciliación de un
// cliente sin teléfono (DP-CIT-01) y la validación de jornada, bloqueos o
// asignación servicio-barbero (HU-061); modificar, reprogramar, cancelar,
// completar o corregir una cita existente; tokens públicos, recordatorios,
// notificaciones, cierre automático, anonimización o métricas del piloto.
// `schedule` y `catalog` no importan este paquete ni sus tablas, y este
// paquete no importa `schedule` ni `catalog`.
package booking
