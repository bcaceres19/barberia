-- Fixture propio de HU-093: dos barberías DEDICADAS a las pruebas de
-- shops/postgres.BookingPolicyRepository (lectura/actualización de la
-- política pública de reserva y cancelación, concurrencia optimista y
-- aislamiento por tenant), separadas de dos_barberias.sql y de las demás
-- suites por el mismo motivo que hu091_catalogo_publico.sql/
-- hu092_seleccion_barbero.sql: estas pruebas EJECUTAN UPDATE real sobre
-- `barbershop.min_advance_minutes`/etc. y no deben competir con otro
-- paquete de test que también escriba esa fila (issue #158, ver
-- shops/postgres/repository_test.go).
--
-- IDs '0093...' (mnemónico: HU-093), distintos de '1.../2...'
-- (dos_barberias.sql), '3.../4...' (HU-021), '5.../6...' (HU-022),
-- '7.../8...' (HU-023), 'b10c...' (HU-042), 'c17a...' (HU-060),
-- 'c.../d...'/'e.../f...' (HU-040/HU-041), '0090...' (hu090), '0091...'
-- (hu091) y '0092...' (hu092).
--
-- Ambas barberías se dejan en sus valores DEFAULT (DEC-083, CA-093-01: sin
-- overrides explícitos, para poder verificar exactamente esos defaults).
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu093_politica_reserva.sql
--
-- Se ejecuta con el rol migrador (o superusuario en local/CI efímero,
-- mismo criterio que dos_barberias.sql).
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

INSERT INTO barbershop (id, name, timezone) VALUES
  ('00930001-0093-0093-0093-009300010001', 'Barbería de prueba HU-093 Uno', 'America/Bogota'),
  ('00930002-0093-0093-0093-009300020002', 'Barbería de prueba HU-093 Dos', 'America/Bogota')
ON CONFLICT (id) DO NOTHING;

COMMIT;
