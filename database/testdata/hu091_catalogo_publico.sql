-- Fixture propio de HU-091: dos barberías DEDICADAS a las pruebas de
-- publicbooking/postgres.ListPublicServices (catálogo público de servicios
-- activos y asignados, aislamiento por tenant), separadas de
-- hu090_reserva_publica.sql y de las demás suites por el mismo motivo: un
-- conteo exacto de servicios ofrecidos no debe depender del estado que
-- dejen otras suites (catalog, staff) sobre servicios/asignaciones.
--
-- IDs '0091...' (mnemónico: HU-091; solo dígitos hexadecimales válidos
-- 0/9/1), distintos de '1.../2...' (dos_barberias.sql), '3.../4...'
-- (HU-021), '5.../6...' (HU-022), '7.../8...' (HU-023), 'b10c...' (HU-042),
-- 'c17a...' (HU-060), 'c.../d...'/'e.../f...' (HU-040/HU-041) y de '0090...'
-- (hu090_reserva_publica.sql).
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu091_catalogo_publico.sql
--
-- Se ejecuta con el rol migrador (o superusuario en local/CI efímero,
-- mismo criterio que dos_barberias.sql), el único con política
-- administrativa sobre barbershop/barber/service/barber_service.
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

-- Barbería Uno: publicable, con los tres casos que CA-091-01/CA-091-02
-- deben distinguir (activo+asignado, activo+sin asignar, inactivo+asignado).
INSERT INTO barbershop (id, name, timezone, public_slug) VALUES
  ('00910001-0091-0091-0091-009100010001',
   'Barbería de prueba HU-091 Uno', 'America/Bogota', 'barberia-hu091-uno')
ON CONFLICT (id) DO NOTHING;

-- Barbería Dos: publicable, usada exclusivamente para CA-091-03
-- (aislamiento de tenant: su único servicio nunca debe aparecer al
-- resolver el slug de la barbería Uno, y viceversa).
INSERT INTO barbershop (id, name, timezone, public_slug) VALUES
  ('00910002-0091-0091-0091-009100020002',
   'Barbería de prueba HU-091 Dos', 'America/Bogota', 'barberia-hu091-dos')
ON CONFLICT (id) DO NOTHING;

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('00910011-0091-0091-0091-009100110011',
   '00910001-0091-0091-0091-009100010001', 'Barbero HU-091 Uno-1'),
  ('00910012-0091-0091-0091-009100120012',
   '00910002-0091-0091-0091-009100020002', 'Barbero HU-091 Dos-1')
ON CONFLICT (id) DO NOTHING;

-- Uno: activo y asignado (DEBE aparecer, CA-091-01).
INSERT INTO service (id, barbershop_id, name, description, duration_minutes, price_amount, price_currency, is_active) VALUES
  ('00910101-0091-0091-0091-009101010101',
   '00910001-0091-0091-0091-009100010001',
   'Corte activo asignado HU-091', 'Servicio público visible de prueba', 30, 35000.00, 'COP', true)
ON CONFLICT (id) DO NOTHING;

-- Uno: segundo servicio activo y asignado (DEBE aparecer). Existe junto al
-- primero para poder probar la paginación por cursor con dos páginas reales
-- (repository_test.go: TestListPublicServices_CursorPastLastItem_...).
INSERT INTO service (id, barbershop_id, name, description, duration_minutes, price_amount, price_currency, is_active) VALUES
  ('00910105-0091-0091-0091-009101050105',
   '00910001-0091-0091-0091-009100010001',
   'Corte activo asignado HU-091 B', NULL, 45, 50000.00, 'COP', true)
ON CONFLICT (id) DO NOTHING;

-- Uno: activo pero SIN ninguna asignación (NO debe aparecer, CA-091-01: "al
-- menos una asignación vigente").
INSERT INTO service (id, barbershop_id, name, description, duration_minutes, price_amount, price_currency, is_active) VALUES
  ('00910102-0091-0091-0091-009101020102',
   '00910001-0091-0091-0091-009100010001',
   'Corte activo sin asignar HU-091', NULL, 20, 15000.00, 'COP', true)
ON CONFLICT (id) DO NOTHING;

-- Uno: inactivo pero asignado (NO debe aparecer, CA-091-02: desactivar lo
-- retira de nuevas lecturas aunque conserve asignación).
INSERT INTO service (id, barbershop_id, name, description, duration_minutes, price_amount, price_currency, is_active, deactivated_at) VALUES
  ('00910103-0091-0091-0091-009101030103',
   '00910001-0091-0091-0091-009100010001',
   'Corte inactivo asignado HU-091', NULL, 40, 40000.00, 'COP', false, '2026-09-01T00:00:00Z')
ON CONFLICT (id) DO NOTHING;

-- Dos: activo y asignado, único servicio de esta barbería (CA-091-03).
INSERT INTO service (id, barbershop_id, name, description, duration_minutes, price_amount, price_currency, is_active) VALUES
  ('00910201-0091-0091-0091-009102010201',
   '00910002-0091-0091-0091-009100020002',
   'Corte activo asignado HU-091 Dos', NULL, 25, 22000.00, 'COP', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO barber_service (barbershop_id, barber_id, service_id) VALUES
  ('00910001-0091-0091-0091-009100010001',
   '00910011-0091-0091-0091-009100110011',
   '00910101-0091-0091-0091-009101010101'),
  ('00910001-0091-0091-0091-009100010001',
   '00910011-0091-0091-0091-009100110011',
   '00910105-0091-0091-0091-009101050105'),
  ('00910001-0091-0091-0091-009100010001',
   '00910011-0091-0091-0091-009100110011',
   '00910103-0091-0091-0091-009101030103'),
  ('00910002-0091-0091-0091-009100020002',
   '00910012-0091-0091-0091-009100120012',
   '00910201-0091-0091-0091-009102010201')
ON CONFLICT DO NOTHING;

COMMIT;
