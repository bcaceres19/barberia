# testdata

Escenarios controlados para pruebas de integración y volumen (por ejemplo,
datos representativos para revisar planes de consulta), según
[`docs/05-backend/estandar-base-datos.md`](../../docs/05-backend/estandar-base-datos.md)
sección 2.

| Archivo | Escenario |
| --- | --- |
| `dos_barberias.sql` | Dos barberías con usuarios propios (HU-001, aislamiento RLS) |
| `notification_lease_fixture.sql` | Barbero, servicio, cliente y cita fijos para `tests/notification_lease_concurrency*.sql` (issue #5) |
| `customer_anonymization_fixture.sql` | Barbero y servicio fijos por barbería para `tests/customer_anonymization.sql` (issue #6) |
