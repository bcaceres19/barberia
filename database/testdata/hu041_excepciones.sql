-- Fixture propio de HU-041: dos barberías DEDICADAS a las pruebas de
-- schedule/postgres (excepciones de jornada y festivos, aislamiento por
-- tenant), separadas de dos_barberias.sql/hu021_barberos.sql/
-- hu040_horario.sql por el mismo motivo: un conteo exacto de excepciones
-- por barbero no debe depender del estado que dejen otras suites.
--
-- IDs 'e...'/'f...' (M/N), distintos de 'c...'/'d...' (K/L, HU-040) y de
-- '9...'/'a...'/'b...' (reservados como sentinelas o usados como ids de
-- otras tablas en otras suites).
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu041_excepciones.sql
--
-- Se ejecuta con el rol migrador (o superusuario en local/CI efímero,
-- mismo criterio que dos_barberias.sql), el único con política
-- administrativa sobre barbershop/barber.
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

INSERT INTO barbershop (id, name, timezone) VALUES
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'Barbería de prueba M (schedule excepciones)', 'America/Bogota'),
  ('ffffffff-ffff-ffff-ffff-ffffffffffff', 'Barbería de prueba N (schedule excepciones)', 'America/Bogota')
ON CONFLICT (id) DO NOTHING;

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('eeee0001-0001-0001-0001-000100010001', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'Barbero M-1'),
  ('eeee0002-0002-0002-0002-000200020002', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'Barbero M-2'),
  ('ffff0001-0001-0001-0001-000100010001', 'ffffffff-ffff-ffff-ffff-ffffffffffff', 'Barbero N-1')
ON CONFLICT (id) DO NOTHING;

COMMIT;
