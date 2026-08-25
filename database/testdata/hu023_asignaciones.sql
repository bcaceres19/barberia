-- Fixture propio de HU-023: dos barberías DEDICADAS a las pruebas de
-- catalog/postgres.AssignmentRepository (asignar, desasignar, DEC-068,
-- aislamiento por tenant), separadas de dos_barberias.sql/
-- hu021_barberos.sql/hu022_catalogo.sql para que un conteo exacto de
-- asignaciones no dependa del estado que dejen otras suites (settings,
-- auth, staff, catalog) sobre shopA-shopF.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu023_asignaciones.sql
--
-- Se ejecuta con el rol migrador (o superusuario en local/CI efímero,
-- mismo criterio que dos_barberias.sql), el único con política
-- administrativa sobre barbershop.
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

INSERT INTO barbershop (id, name, timezone) VALUES
  ('77777777-7777-7777-7777-777777777777', 'Barbería de prueba G (asignaciones)', 'America/Bogota'),
  ('88888888-8888-8888-8888-888888888888', 'Barbería de prueba H (asignaciones)', 'America/Bogota')
ON CONFLICT (id) DO NOTHING;

COMMIT;
