-- Fixture propio de HU-060: dos barberías DEDICADAS a las pruebas de
-- booking/postgres (núcleo de citas, exclusión anti-cruces, aislamiento por
-- tenant), separadas de dos_barberias.sql y de las demás suites por el
-- mismo motivo que hu042_bloqueos.sql: una carrera de exclusión o un
-- conteo de citas por barbero no debe depender del estado que dejen otras
-- suites.
--
-- IDs 'c17a...' ("cita", mnemónico; solo dígitos hexadecimales válidos
-- c/1/7/a), distintos de '1.../2...' (dos_barberias.sql), '3.../4...'
-- (HU-021), '5.../6...' (HU-022), '7.../8...' (HU-023), 'b10c...' (HU-042),
-- 'c.../d...'/'e.../f...' (HU-040/HU-041, repetidos completos) y de
-- 'b1a0.../5e10...' (customer_anonymization_fixture.sql).
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu060_citas.sql
--
-- Se ejecuta con el rol migrador (o superusuario en local/CI efímero,
-- mismo criterio que dos_barberias.sql), el único con política
-- administrativa sobre barbershop/barber/service/staff_user.
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

INSERT INTO barbershop (id, name, timezone) VALUES
  ('c17a0001-c17a-c17a-c17a-c17a00010001', 'Barbería de prueba Q (booking núcleo)', 'America/Bogota'),
  ('c17a0002-c17a-c17a-c17a-c17a00020002', 'Barbería de prueba R (booking núcleo)', 'America/Bogota')
ON CONFLICT (id) DO NOTHING;

INSERT INTO staff_user (id, barbershop_id, email, full_name, is_active) VALUES
  ('c17a9001-9001-9001-9001-900190019001',
   'c17a0001-c17a-c17a-c17a-c17a00010001',
   'dueno.q@ejemplo.test', 'Dueño Q', true),
  ('c17a9002-9002-9002-9002-900290029002',
   'c17a0002-c17a-c17a-c17a-c17a00020002',
   'dueno.r@ejemplo.test', 'Dueño R', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('c17a1001-1001-1001-1001-100110011001', 'c17a0001-c17a-c17a-c17a-c17a00010001', 'Barbero Q-1'),
  ('c17a1002-1002-1002-1002-100210021002', 'c17a0001-c17a-c17a-c17a-c17a00010001', 'Barbero Q-2'),
  ('c17a2001-2001-2001-2001-200120012001', 'c17a0002-c17a-c17a-c17a-c17a00020002', 'Barbero R-1')
ON CONFLICT (id) DO NOTHING;

INSERT INTO service (id, barbershop_id, name, duration_minutes, price_amount, price_currency) VALUES
  ('c17a5001-5001-5001-5001-500150015001',
   'c17a0001-c17a-c17a-c17a-c17a00010001',
   'Corte de prueba Q', 30, 20000.00, 'COP'),
  ('c17a5002-5002-5002-5002-500250025002',
   'c17a0002-c17a-c17a-c17a-c17a00020002',
   'Corte de prueba R', 45, 25000.00, 'COP')
ON CONFLICT (id) DO NOTHING;

COMMIT;
