-- Escenario controlado de HU-001: dos barberías con usuarios propios.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--
-- Se ejecuta con el rol migrador, que es el único con política administrativa
-- sobre las tablas con RLS forzada.
--
-- Identificadores fijos a propósito: las pruebas de aislamiento necesitan
-- conocer un identificador de la barbería ajena para intentar alcanzarlo
-- (base-datos.md §4.8).
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

INSERT INTO barbershop (id, name, timezone) VALUES
  ('11111111-1111-1111-1111-111111111111', 'Barbería de prueba A', 'America/Bogota'),
  ('22222222-2222-2222-2222-222222222222', 'Barbería de prueba B', 'America/Bogota')
ON CONFLICT (id) DO NOTHING;

INSERT INTO staff_user (id, barbershop_id, email, full_name, is_active) VALUES
  ('aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaa1',
   '11111111-1111-1111-1111-111111111111',
   'duena.a@ejemplo.test',  'Dueña A',   true),

  ('aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2',
   '11111111-1111-1111-1111-111111111111',
   'barbero.a@ejemplo.test', 'Barbero A', true),

  ('bbbbbbb1-bbbb-bbbb-bbbb-bbbbbbbbbbb1',
   '22222222-2222-2222-2222-222222222222',
   'dueno.b@ejemplo.test',  'Dueño B',   true),

  ('bbbbbbb2-bbbb-bbbb-bbbb-bbbbbbbbbbb2',
   '22222222-2222-2222-2222-222222222222',
   'barbero.b@ejemplo.test', 'Barbero B', false)
ON CONFLICT (id) DO NOTHING;

COMMIT;
