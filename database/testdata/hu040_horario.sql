-- Fixture propio de HU-040: dos barberías DEDICADAS a las pruebas de
-- schedule/postgres (horario laboral recurrente, aislamiento por tenant),
-- separadas de dos_barberias.sql/hu021_barberos.sql/hu022_catalogo.sql/
-- hu023_asignaciones.sql por el mismo motivo: un conteo exacto de tramos
-- por barbero no debe depender del estado que dejen otras suites.
--
-- IDs 'c...'/'d...' (K/L), no '9...'/'a...' (I/J): esos dos ya están
-- reservados como identificadores de barbería DELIBERADAMENTE inexistentes
-- por internal/modules/auth/postgres/session_repository_test.go
-- (TestBarbershopName_UnknownID_ReturnsError) y database/tests/
-- hu001_aislamiento_rls.sql/hu021_barberos.sql (ids de otras tablas);
-- reutilizarlos como barbershop.id real habría roto esa prueba.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu040_horario.sql
--
-- Se ejecuta con el rol migrador (o superusuario en local/CI efímero,
-- mismo criterio que dos_barberias.sql), el único con política
-- administrativa sobre barbershop/barber.
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

INSERT INTO barbershop (id, name, timezone) VALUES
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'Barbería de prueba K (schedule)', 'America/Bogota'),
  ('dddddddd-dddd-dddd-dddd-dddddddddddd', 'Barbería de prueba L (schedule)', 'America/Bogota')
ON CONFLICT (id) DO NOTHING;

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('cccc0001-0001-0001-0001-000100010001', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'Barbero K-1'),
  ('cccc0002-0002-0002-0002-000200020002', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'Barbero K-2'),
  ('dddd0001-0001-0001-0001-000100010001', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'Barbero L-1')
ON CONFLICT (id) DO NOTHING;

COMMIT;
