-- Fixture propio de HU-021: dos barberías DEDICADAS a las pruebas de
-- staff/postgres (registro, listado, aislamiento por tenant e
-- idempotencia), separadas de las de dos_barberias.sql para que un conteo
-- exacto de barberos en una página no dependa del estado que dejen otras
-- suites (settings, auth) sobre shopA/shopB.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu021_barberos.sql
--
-- Se ejecuta con el rol migrador (o superusuario en local/CI efímero,
-- mismo criterio que dos_barberias.sql), el único con política
-- administrativa sobre barbershop.
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

INSERT INTO barbershop (id, name, timezone) VALUES
  ('33333333-3333-3333-3333-333333333333', 'Barbería de prueba C (staff)', 'America/Bogota'),
  ('44444444-4444-4444-4444-444444444444', 'Barbería de prueba D (staff)', 'America/Bogota')
ON CONFLICT (id) DO NOTHING;

-- Issue #322 (DEC-100): usuarios del área privada que las pruebas de
-- staff/postgres vinculan con barberos (dos en la C para probar cambiar y
-- competir por un barbero; uno en la D para probar el aislamiento).
INSERT INTO staff_user (id, barbershop_id, email, full_name, is_active) VALUES
  ('cccccc01-cccc-4ccc-8ccc-cccccccccc01', '33333333-3333-3333-3333-333333333333',
   'vinculo.c1@ejemplo.test', 'Usuario vínculo C1', true),
  ('cccccc02-cccc-4ccc-8ccc-cccccccccc02', '33333333-3333-3333-3333-333333333333',
   'vinculo.c2@ejemplo.test', 'Usuario vínculo C2', true),
  ('dddddd01-dddd-4ddd-8ddd-dddddddddd01', '44444444-4444-4444-4444-444444444444',
   'vinculo.d1@ejemplo.test', 'Usuario vínculo D1', true)
ON CONFLICT (id) DO NOTHING;

COMMIT;
