-- Fixture mínimo para database/tests/notification_lease_concurrency.sql.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/notification_lease_fixture.sql
--
-- Se ejecuta con el rol migrador (política administrativa sobre tablas con
-- RLS forzada). Identificadores fijos e idempotente (ON CONFLICT DO NOTHING),
-- igual que dos_barberias.sql: una sola cita conocida sobre la que las
-- pruebas de lease crean y destruyen sus propias filas de
-- notification_schedule dentro de transacciones que revierten.
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

-- Reutiliza la barbería A de dos_barberias.sql si ya está cargada; si no,
-- la crea aquí para que este fixture también funcione de forma aislada.
INSERT INTO barbershop (id, name, timezone) VALUES
  ('11111111-1111-1111-1111-111111111111', 'Barbería de prueba A', 'America/Bogota')
ON CONFLICT (id) DO NOTHING;

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1',
   '11111111-1111-1111-1111-111111111111',
   'Barbero de prueba de lease')
ON CONFLICT (id) DO NOTHING;

INSERT INTO service (id, barbershop_id, name, duration_minutes, price_amount, price_currency) VALUES
  ('5e5e5e5e-5e5e-5e5e-5e5e-5e5e5e5e5e5e',
   '11111111-1111-1111-1111-111111111111',
   'Corte de prueba de lease', 30, 20000.00, 'COP')
ON CONFLICT (id) DO NOTHING;

INSERT INTO customer (id, barbershop_id, full_name) VALUES
  ('c05c05c0-c05c-c05c-c05c-c05c05c05c05',
   '11111111-1111-1111-1111-111111111111',
   'Cliente de prueba de lease')
ON CONFLICT (id) DO NOTHING;

INSERT INTO appointment (
  id, barbershop_id, barber_id, service_id, customer_id, attendee_name,
  starts_at, ends_at, origin,
  service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
) VALUES (
  'a99a99a9-a99a-a99a-a99a-a99a99a99a99',
  '11111111-1111-1111-1111-111111111111',
  'c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1',
  '5e5e5e5e-5e5e-5e5e-5e5e-5e5e5e5e5e5e',
  'c05c05c0-c05c-c05c-c05c-c05c05c05c05',
  'Cliente de prueba de lease',
  '2026-01-05 10:00:00-05', '2026-01-05 10:30:00-05', 'manual',
  'Corte de prueba de lease', 30, 20000.00, 'COP'
)
ON CONFLICT (id) DO NOTHING;

COMMIT;
