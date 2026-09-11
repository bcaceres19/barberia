-- Fixture propio de HU-090: dos barberías DEDICADAS a las pruebas de
-- publicbooking/postgres (resolución pública del enlace de reservas),
-- separadas de dos_barberias.sql y de las demás suites por el mismo motivo
-- que hu060_citas.sql: una prueba de aislamiento de tenant no debe depender
-- del estado que dejen otras suites, y estas filas SÍ requieren un
-- public_slug ya asignado (a diferencia de dos_barberias.sql, cuyas dos
-- barberías nacen sin slug para las pruebas de generación automática de
-- shops/postgres).
--
-- IDs '0090...' (mnemónico: HU-090; solo dígitos hexadecimales válidos
-- 0/9), distintos de '1.../2...' (dos_barberias.sql), '3.../4...'
-- (HU-021), '5.../6...' (HU-022), '7.../8...' (HU-023), 'b10c...' (HU-042),
-- 'c17a...' (HU-060), 'c.../d...'/'e.../f...' (HU-040/HU-041) y de
-- 'b1a0.../5e10...' (customer_anonymization_fixture.sql).
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu090_reserva_publica.sql
--
-- Se ejecuta con el rol migrador (o superusuario en local/CI efímero,
-- mismo criterio que dos_barberias.sql), el único con política
-- administrativa sobre barbershop.
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

INSERT INTO barbershop (id, name, timezone, contact_email, contact_phone, public_slug) VALUES
  ('00900001-0090-0090-0090-009000010001',
   'Barbería de prueba HU-090 Uno', 'America/Bogota',
   'contacto.hu090uno@ejemplo.test', '+573000000001', 'barberia-hu090-uno'),
  ('00900002-0090-0090-0090-009000020002',
   'Barbería de prueba HU-090 Dos', 'America/Bogota', NULL, NULL, 'barberia-hu090-dos')
ON CONFLICT (id) DO NOTHING;

-- Tercera barbería, deliberadamente SIN public_slug: cubre "existe pero no
-- es publicable" (CA-090-02), indistinguible desde fuera de "no existe".
INSERT INTO barbershop (id, name, timezone, public_slug) VALUES
  ('00900003-0090-0090-0090-009000030003',
   'Barbería de prueba HU-090 No Publicable', 'America/Bogota', NULL)
ON CONFLICT (id) DO NOTHING;

COMMIT;
