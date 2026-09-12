-- Fixture propio de HU-094: dos barberías DEDICADAS a las pruebas de
-- publicbooking/postgres (ResolveBarbershopID, ListOccupiedIntervals) y a
-- las pruebas HTTP de disponibilidad pública en cmd/api, separadas de las
-- demás suites por el mismo motivo que hu091_catalogo_publico.sql/
-- hu092_seleccion_barbero.sql: un cálculo de franjas no debe depender de
-- horarios, bloqueos o citas que deje otra suite sobre un barbero
-- compartido.
--
-- IDs '0094...' (mnemónico: HU-094), distintos de '1.../2...'
-- (dos_barberias.sql), '3.../4...' (HU-021), '5.../6...' (HU-022),
-- '7.../8...' (HU-023), 'b10c...' (HU-042), 'c17a...' (HU-060),
-- 'c.../d...'/'e.../f...' (HU-040/HU-041), '0090...'/'0091...'/'0092...'/
-- '0093...' (hu090-hu093).
--
-- La jornada laboral cubre los SIETE días ISO (9:00-18:00) a propósito:
-- las pruebas HTTP de HU-094 corren en tiempo real ("hoy" varía según
-- cuándo se ejecuten) y no pueden depender de que el día de la corrida
-- tenga jornada configurada.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu094_disponibilidad.sql
--
-- Se ejecuta con el rol migrador (o superusuario en local/CI efímero,
-- mismo criterio que dos_barberias.sql).
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

-- Barbería Uno: publicable, con anticipación mínima 0 (para que "ahora
-- mismo" nunca descarte por RN-DIS-04 en una prueba) y ventana amplia.
INSERT INTO barbershop (
  id, name, timezone, public_slug,
  min_advance_minutes, max_advance_days, slot_grid_minutes
) VALUES
  ('00940001-0094-0094-0094-009400010001',
   'Barbería de prueba HU-094 Uno', 'America/Bogota', 'barberia-hu094-uno',
   0, 7, 30)
ON CONFLICT (id) DO NOTHING;

-- Barbería Dos: publicable, usada exclusivamente para el aislamiento de
-- tenant (CA-094-05): su barbero y su servicio nunca deben producir
-- disponibilidad al consultarse a través del slug de la barbería Uno.
INSERT INTO barbershop (
  id, name, timezone, public_slug,
  min_advance_minutes, max_advance_days, slot_grid_minutes
) VALUES
  ('00940002-0094-0094-0094-009400020002',
   'Barbería de prueba HU-094 Dos', 'America/Bogota', 'barberia-hu094-dos',
   0, 7, 30)
ON CONFLICT (id) DO NOTHING;

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('00940011-0094-0094-0094-009400110011',
   '00940001-0094-0094-0094-009400010001', 'Barbero HU-094 Uno-1'),
  ('00940012-0094-0094-0094-009400120012',
   '00940002-0094-0094-0094-009400020002', 'Barbero HU-094 Dos-1')
ON CONFLICT (id) DO NOTHING;

INSERT INTO service (id, barbershop_id, name, description, duration_minutes, price_amount, price_currency, is_active) VALUES
  ('00940101-0094-0094-0094-009401010101',
   '00940001-0094-0094-0094-009400010001',
   'Servicio disponibilidad HU-094', NULL, 30, 35000.00, 'COP', true),
  ('00940201-0094-0094-0094-009402010201',
   '00940002-0094-0094-0094-009400020002',
   'Servicio disponibilidad HU-094 Dos', NULL, 30, 35000.00, 'COP', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO barber_service (barbershop_id, barber_id, service_id) VALUES
  ('00940001-0094-0094-0094-009400010001',
   '00940011-0094-0094-0094-009400110011',
   '00940101-0094-0094-0094-009401010101'),
  ('00940002-0094-0094-0094-009400020002',
   '00940012-0094-0094-0094-009400120012',
   '00940201-0094-0094-0094-009402010201')
ON CONFLICT DO NOTHING;

-- Jornada 9:00-18:00 los siete días ISO, para ambos barberos.
INSERT INTO working_hour (barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
SELECT '00940001-0094-0094-0094-009400010001', '00940011-0094-0094-0094-009400110011', d, '09:00', 540
  FROM generate_series(1, 7) AS d
ON CONFLICT DO NOTHING;

INSERT INTO working_hour (barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
SELECT '00940002-0094-0094-0094-009400020002', '00940012-0094-0094-0094-009400120012', d, '09:00', 540
  FROM generate_series(1, 7) AS d
ON CONFLICT DO NOTHING;

COMMIT;
