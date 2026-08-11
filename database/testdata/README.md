# testdata

Escenarios controlados para pruebas de integración y volumen (por ejemplo,
datos representativos para revisar planes de consulta), según
[`docs/05-backend/estandar-base-datos.md`](../../docs/05-backend/estandar-base-datos.md)
sección 2.

| Archivo | Escenario |
| --- | --- |
| `dos_barberias.sql` | Dos barberías con usuarios propios (HU-001, aislamiento RLS) |
| `customer_anonymization_fixture.sql` | Barbero y servicio fijos por barbería para `tests/customer_anonymization.sql` (issue #6) |
