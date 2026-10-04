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
| `rls_suite_fixture.sql` | Una fila por tabla protegida por RLS, en ambas barberías, para `tests/rls_suite.sql` (issue #7) |

| `ui_bloqueos_barberos_286.sql` | Dos barberías y usuarios sintéticos dedicados al E2E de integración de bloqueos en Barberos (#286). Aplicar únicamente en BD local de pruebas con rol migrador; incluye hashes argon2id reales de una contraseña ficticia documentada. |

- `ui_paginacion_barberos_288.sql`: dos tenants dedicados para el E2E de paginación numerada de Barberos (#288), con 17 y 3 personas sintéticas. Solo entorno local/test; no aplicar en producción.
- `ui_barbero_individual_294.sql`: una barbería con UN solo barbero, perfil `solo`, cuatro servicios (tres ya ofrecidos), horario de lunes a sábado y tres turnos de hoy, para ver y probar el perfil de barbero individual (#294, `DEC-115`). Cuenta ficticia `mateo.demo@ejemplo.test` con la contraseña de desarrollo documentada en el archivo (hash argon2id real). Repetible: recrea los turnos de hoy. Solo entorno local/test; no aplicar en producción.
