-- Fixture propio de HU-098: dos barberías DEDICADAS a las pruebas de
-- customeraccess/postgres (lectura del turno por token de acceso),
-- separadas de dos_barberias.sql y de las demás suites por el mismo motivo
-- que hu060_citas.sql/hu090_reserva_publica.sql: una prueba de aislamiento
-- de tenant no debe depender del estado que dejen otras suites, y estas
-- filas necesitan una política de reserva/cancelación DISTINTA por
-- barbería para que una fuga cruzada de CancellationDeadlineMinutes/
-- LateCancellation* sea detectable (CA-098-05).
--
-- IDs '0098...' (mnemónico: HU-098; solo dígitos hexadecimales válidos
-- 0/9/8), distintos de '1.../2...' (dos_barberias.sql), '3.../4...'
-- (HU-021), '5.../6...' (HU-022), '7.../8...' (HU-023), 'b10c...'
-- (HU-042), 'c17a...' (HU-060), 'c.../d...'/'e.../f...' (HU-040/HU-041),
-- '0090...'/'0091...'/'0092...' (HU-090/091/092) y de
-- 'b1a0.../5e10...' (customer_anonymization_fixture.sql).
--
-- El cliente y la cita de cada barbería NO se insertan aquí: cada prueba
-- de customeraccess/postgres los crea con
-- booking/postgres.Repository.CreateInternal (la misma primitiva probada
-- en booking/postgres/repository_test.go), y agrega su propia fila
-- appointment_access_token con SQL directo dentro de InTenantTx -esta
-- última tabla es responsabilidad de customeraccess/publicbooking, no de
-- booking, mismo criterio que insertSyntheticHistoryRow en
-- booking/postgres/detail_repository_test.go.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu098_acceso_turno.sql
--
-- Se ejecuta con el rol migrador (o superusuario en local/CI efímero,
-- mismo criterio que dos_barberias.sql), el único con política
-- administrativa sobre barbershop/barber/service.
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

INSERT INTO barbershop (
  id, name, timezone,
  cancellation_deadline_minutes, late_cancellation_client_allowed, late_cancellation_reason_required
) VALUES
  ('00980001-0098-0098-0098-009800010001',
   'Barbería de prueba HU-098 Uno', 'America/Bogota',
   30, true, false),
  ('00980002-0098-0098-0098-009800020002',
   'Barbería de prueba HU-098 Dos', 'America/Bogota',
   90, false, false)
ON CONFLICT (id) DO NOTHING;

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('00980011-0098-0098-0098-009800110011',
   '00980001-0098-0098-0098-009800010001', 'Barbero HU-098 Uno'),
  ('00980021-0098-0098-0098-009800210021',
   '00980002-0098-0098-0098-009800020002', 'Barbero HU-098 Dos')
ON CONFLICT (id) DO NOTHING;

INSERT INTO service (id, barbershop_id, name, duration_minutes, price_amount, price_currency) VALUES
  ('00980012-0098-0098-0098-009800120012',
   '00980001-0098-0098-0098-009800010001',
   'Corte de prueba HU-098 Uno', 30, 20000.00, 'COP'),
  ('00980022-0098-0098-0098-009800220022',
   '00980002-0098-0098-0098-009800020002',
   'Corte de prueba HU-098 Dos', 45, 25000.00, 'COP')
ON CONFLICT (id) DO NOTHING;

COMMIT;
