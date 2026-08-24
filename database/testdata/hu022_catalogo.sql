-- Fixture propio de HU-022: dos barberías DEDICADAS a las pruebas de
-- catalog/postgres (alta, listado, aislamiento por tenant e idempotencia),
-- separadas de las de dos_barberias.sql y hu021_barberos.sql para que un
-- conteo exacto de servicios en una página no dependa del estado que dejen
-- otras suites (settings, auth, staff) sobre shopA/shopB/shopC/shopD.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu022_catalogo.sql
--
-- Se ejecuta con el rol migrador (o superusuario en local/CI efímero,
-- mismo criterio que dos_barberias.sql), el único con política
-- administrativa sobre barbershop.
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

INSERT INTO barbershop (id, name, timezone) VALUES
  ('55555555-5555-5555-5555-555555555555', 'Barbería de prueba E (catalog)', 'America/Bogota'),
  ('66666666-6666-6666-6666-666666666666', 'Barbería de prueba F (catalog)', 'America/Bogota')
ON CONFLICT (id) DO NOTHING;

COMMIT;
