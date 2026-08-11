# tests

Comprobaciones SQL globales de restricciones, RLS y migración con al menos
dos tenants y el rol real de la aplicación, según
[`docs/05-backend/estandar-base-datos.md`](../../docs/05-backend/estandar-base-datos.md)
secciones 9 y 13.

| Archivo | Cubre |
| --- | --- |
| `hu001_aislamiento_rls.sql` | CA-001-01 a CA-001-06: aislamiento RLS de HU-001 |
| `customer_anonymization.sql` | Anonimización completa de clientes vencidos (issue #6, `DDL-PRI-01`, `DEC-042`, `DEC-049`): solicitud individual, vencimiento masivo, dos tenants, repetición, caída a mitad de lote, fecha ancla por última actividad |
