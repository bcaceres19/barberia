-- Fixture propio de HU-100: dos barberías DEDICADAS, publicables, a las pruebas
-- de reviews/postgres (reseñas internas por barbero), separadas de las demás
-- suites por el mismo motivo que hu098_acceso_turno.sql/hu092_seleccion_barbero.sql:
-- una prueba de aislamiento de tenant no debe depender del estado que dejen
-- otras suites, y las filas de `barber_review` y de citas completadas no se
-- pueden borrar (RN-HIS-02), así que las pruebas usan barberos propios.
--
-- IDs '0101...' (mnemónico: HU-100 → 0100 + dígito), distintos de los demás
-- fixtures.
--
-- Cada prueba crea su propio barbero, sus citas (booking/postgres.CreateInternal)
-- y su fila appointment_access_token con SQL directo, para que ninguna dependa
-- del promedio que dejó otra ni de una corrida anterior contra la misma base.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu100_resenas.sql
--
-- Se ejecuta con el rol migrador (o superusuario en local/CI efímero).
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

INSERT INTO barbershop (id, name, timezone, public_slug) VALUES
  ('01010001-0101-0101-0101-010100010001', 'Barbería de prueba HU-100 Uno', 'America/Bogota', 'barberia-hu100-uno'),
  ('01010002-0101-0101-0101-010100020002', 'Barbería de prueba HU-100 Dos', 'America/Bogota', 'barberia-hu100-dos')
ON CONFLICT (id) DO NOTHING;

INSERT INTO service (id, barbershop_id, name, duration_minutes, price_amount, price_currency) VALUES
  ('01010031-0101-0101-0101-010100310031', '01010001-0101-0101-0101-010100010001', 'Corte de prueba HU-100 Uno', 30, 20000.00, 'COP'),
  ('01010032-0101-0101-0101-010100320032', '01010002-0101-0101-0101-010100020002', 'Corte de prueba HU-100 Dos', 30, 20000.00, 'COP')
ON CONFLICT (id) DO NOTHING;

-- Un usuario del área privada por barbería, para abrir sesión en las pruebas de
-- cmd/api (moderación de reseñas desde el panel).
INSERT INTO staff_user (id, barbershop_id, email, full_name, is_active) VALUES
  ('01019001-0101-0101-0101-010190019001', '01010001-0101-0101-0101-010100010001', 'moderador.uno@ejemplo.test', 'Moderador Uno', true),
  ('01019002-0101-0101-0101-010190029002', '01010002-0101-0101-0101-010100020002', 'moderador.dos@ejemplo.test', 'Moderador Dos', true)
ON CONFLICT (id) DO NOTHING;

COMMIT;
