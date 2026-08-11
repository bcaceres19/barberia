# tests

Comprobaciones SQL globales de restricciones, RLS y migración con al menos
dos tenants y el rol real de la aplicación, según
[`docs/05-backend/estandar-base-datos.md`](../../docs/05-backend/estandar-base-datos.md)
secciones 9 y 13.

| Archivo | Cubre |
| --- | --- |
| `hu001_aislamiento_rls.sql` | CA-001-01 a CA-001-06: aislamiento RLS de HU-001 |
| `notification_lease_concurrency.sql` | Protocolo de lease de `notification_claim_due`/`notification_finalize_claim` y endurecimiento de `retention_claim_due_customers` (issue #5, `DDL-CON-01`, `DDL-OPS-01`); secuencial, un solo conexión, `p_now` controlado |
| `notification_lease_concurrency_two_connections.sh` | La misma reclamación, pero con dos conexiones `psql` reales y solapadas, para probar que `SKIP LOCKED` salta una fila bloqueada por una transacción ajena todavía abierta (`DDL-CON-02`); requiere `DATABASE_TEST_URL` y `psql` en el `PATH` |
