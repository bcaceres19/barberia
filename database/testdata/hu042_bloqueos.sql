-- Fixture propio de HU-042: dos barberías DEDICADAS a las pruebas de
-- schedule/postgres (bloqueos puntuales y series recurrentes, aislamiento
-- por tenant), separadas de dos_barberias.sql/hu021_barberos.sql/
-- hu040_horario.sql/hu041_excepciones.sql por el mismo motivo: un conteo
-- exacto de bloqueos por barbero no debe depender del estado que dejen
-- otras suites.
--
-- IDs 'b10c...' ("bloc", mnemónico de bloqueo; solo dígitos hexadecimales
-- válidos b/1/0/c), distintos de '1...'/'2...' (dos_barberias.sql),
-- '3...'/'4...' (HU-021), '5...'/'6...' (HU-022), '7...'/'8...' (HU-023),
-- 'c...'/'d...' (HU-040), 'e...'/'f...' (HU-041), y de los ids
-- 'aaaaaaa1...'/'aaaaaaa2...'/'bbbbbbb1...'/'bbbbbbb2...' que
-- dos_barberias.sql/rls_suite_fixture.sql ya usan para staff_user.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu042_bloqueos.sql
--
-- Se ejecuta con el rol migrador (o superusuario en local/CI efímero,
-- mismo criterio que dos_barberias.sql), el único con política
-- administrativa sobre barbershop/barber.
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

INSERT INTO barbershop (id, name, timezone) VALUES
  ('b10c0001-b10c-b10c-b10c-b10c00010001', 'Barbería de prueba O (schedule bloqueos)', 'America/Bogota'),
  ('b10c0002-b10c-b10c-b10c-b10c00020002', 'Barbería de prueba P (schedule bloqueos)', 'America/Bogota')
ON CONFLICT (id) DO NOTHING;

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('b10c1001-1001-1001-1001-100110011001', 'b10c0001-b10c-b10c-b10c-b10c00010001', 'Barbero O-1'),
  ('b10c1002-1002-1002-1002-100210021002', 'b10c0001-b10c-b10c-b10c-b10c00010001', 'Barbero O-2'),
  ('b10c2001-2001-2001-2001-200120012001', 'b10c0002-b10c-b10c-b10c-b10c00020002', 'Barbero P-1')
ON CONFLICT (id) DO NOTHING;

-- staff_user propio de cada barbería, exigido por
-- time_block_barbershop_id_deleted_by_fk (RN-BLQ-04: quién retiró el
-- bloqueo).
INSERT INTO staff_user (id, barbershop_id, email, full_name, is_active) VALUES
  ('b10c9001-9001-9001-9001-900190019001',
   'b10c0001-b10c-b10c-b10c-b10c00010001',
   'dueno.o@ejemplo.test', 'Dueño O', true),
  ('b10c9002-9002-9002-9002-900290029002',
   'b10c0002-b10c-b10c-b10c-b10c00020002',
   'dueno.p@ejemplo.test', 'Dueño P', true)
ON CONFLICT (id) DO NOTHING;

COMMIT;
