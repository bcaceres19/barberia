-- Fixture propio de HU-092: dos barberías DEDICADAS a las pruebas de
-- publicbooking/postgres.ListPublicBarbers (selección pública de barbero,
-- matriz 0/1/N y aislamiento por tenant), separadas de
-- hu090_reserva_publica.sql/hu091_catalogo_publico.sql por el mismo motivo:
-- un conteo exacto de barberos elegibles no debe depender del estado que
-- dejen otras suites (catalog, staff) sobre servicios/asignaciones.
--
-- IDs '0092...' (mnemónico: HU-092), distintos de '1.../2...'
-- (dos_barberias.sql), '3.../4...' (HU-021), '5.../6...' (HU-022),
-- '7.../8...' (HU-023), 'b10c...' (HU-042), 'c17a...' (HU-060),
-- 'c.../d...'/'e.../f...' (HU-040/HU-041), '0090...'
-- (hu090_reserva_publica.sql) y '0091...' (hu091_catalogo_publico.sql).
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu092_seleccion_barbero.sql
--
-- Se ejecuta con el rol migrador (o superusuario en local/CI efímero,
-- mismo criterio que dos_barberias.sql), el único con política
-- administrativa sobre barbershop/barber/service/barber_service.
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

-- Barbería Uno: publicable, con los cuatro servicios que cubren la matriz
-- 0/1/N (CA-092-01) y el filtro de servicio inactivo (CA-092-02).
INSERT INTO barbershop (id, name, timezone, public_slug) VALUES
  ('00920001-0092-0092-0092-009200010001',
   'Barbería de prueba HU-092 Uno', 'America/Bogota', 'barberia-hu092-uno')
ON CONFLICT (id) DO NOTHING;

-- Barbería Dos: publicable, usada exclusivamente para el aislamiento de
-- tenant (CA-092-02/CA-092-03): su barbero y su servicio nunca deben
-- aparecer al resolver un serviceId de la barbería Uno, y viceversa.
INSERT INTO barbershop (id, name, timezone, public_slug) VALUES
  ('00920002-0092-0092-0092-009200020002',
   'Barbería de prueba HU-092 Dos', 'America/Bogota', 'barberia-hu092-dos')
ON CONFLICT (id) DO NOTHING;

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('00920011-0092-0092-0092-009200110011',
   '00920001-0092-0092-0092-009200010001', 'Barbero HU-092 Uno-1'),
  ('00920012-0092-0092-0092-009200120012',
   '00920001-0092-0092-0092-009200010001', 'Barbero HU-092 Uno-2'),
  ('00920013-0092-0092-0092-009200130013',
   '00920002-0092-0092-0092-009200020002', 'Barbero HU-092 Dos-1')
ON CONFLICT (id) DO NOTHING;

-- Uno: activo, con exactamente UN barbero asignado (CA-092-01, caso "1":
-- preselección automática).
INSERT INTO service (id, barbershop_id, name, description, duration_minutes, price_amount, price_currency, is_active) VALUES
  ('00920101-0092-0092-0092-009201010101',
   '00920001-0092-0092-0092-009200010001',
   'Servicio un barbero HU-092', NULL, 30, 35000.00, 'COP', true)
ON CONFLICT (id) DO NOTHING;

-- Uno: activo, con DOS barberos asignados (CA-092-01, caso "N": elección
-- explícita).
INSERT INTO service (id, barbershop_id, name, description, duration_minutes, price_amount, price_currency, is_active) VALUES
  ('00920102-0092-0092-0092-009201020102',
   '00920001-0092-0092-0092-009200010001',
   'Servicio varios barberos HU-092', NULL, 45, 50000.00, 'COP', true)
ON CONFLICT (id) DO NOTHING;

-- Uno: activo, SIN ningún barbero asignado (CA-092-01, caso "0": estado
-- vacío distinto de carga/error).
INSERT INTO service (id, barbershop_id, name, description, duration_minutes, price_amount, price_currency, is_active) VALUES
  ('00920103-0092-0092-0092-009201030103',
   '00920001-0092-0092-0092-009200010001',
   'Servicio sin barberos HU-092', NULL, 20, 15000.00, 'COP', true)
ON CONFLICT (id) DO NOTHING;

-- Uno: INACTIVO pero con un barbero asignado (CA-092-02: un servicio
-- inactivo nunca expone barberos, aunque conserve la asignación).
INSERT INTO service (id, barbershop_id, name, description, duration_minutes, price_amount, price_currency, is_active, deactivated_at) VALUES
  ('00920104-0092-0092-0092-009201040104',
   '00920001-0092-0092-0092-009200010001',
   'Servicio inactivo asignado HU-092', NULL, 40, 40000.00, 'COP', false, '2026-09-01T00:00:00Z')
ON CONFLICT (id) DO NOTHING;

-- Dos: activo, con su propio barbero asignado. Su serviceId NUNCA debe
-- devolver barberos al consultarse a través del slug de la barbería Uno
-- (CA-092-02/CA-092-03, aislamiento de tenant).
INSERT INTO service (id, barbershop_id, name, description, duration_minutes, price_amount, price_currency, is_active) VALUES
  ('00920201-0092-0092-0092-009202010201',
   '00920002-0092-0092-0092-009200020002',
   'Servicio HU-092 Dos', NULL, 25, 22000.00, 'COP', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO barber_service (barbershop_id, barber_id, service_id) VALUES
  ('00920001-0092-0092-0092-009200010001',
   '00920011-0092-0092-0092-009200110011',
   '00920101-0092-0092-0092-009201010101'),
  ('00920001-0092-0092-0092-009200010001',
   '00920011-0092-0092-0092-009200110011',
   '00920102-0092-0092-0092-009201020102'),
  ('00920001-0092-0092-0092-009200010001',
   '00920012-0092-0092-0092-009200120012',
   '00920102-0092-0092-0092-009201020102'),
  ('00920001-0092-0092-0092-009200010001',
   '00920011-0092-0092-0092-009200110011',
   '00920104-0092-0092-0092-009201040104'),
  ('00920002-0092-0092-0092-009200020002',
   '00920013-0092-0092-0092-009200130013',
   '00920201-0092-0092-0092-009202010201')
ON CONFLICT DO NOTHING;

COMMIT;
