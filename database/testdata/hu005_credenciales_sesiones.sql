-- Escenario controlado de HU-005: credenciales y sesiones para los cuatro
-- usuarios ya creados por testdata/dos_barberias.sql (que debe cargarse
-- primero).
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu005_credenciales_sesiones.sql
--
-- Se ejecuta con el rol migrador/administrador, igual que dos_barberias.sql:
-- es el único con política administrativa sobre tablas con RLS forzada.
--
-- El valor de password_hash NO es un hash argon2id real: es una cadena
-- ficticia que solo necesita cumplir staff_credential_password_hash_ck
-- (32-512 caracteres). Las pruebas de PostgreSQL de este archivo verifican
-- forma, RLS y privilegios, no el algoritmo de verificación en sí (eso lo
-- cubren las pruebas Go de internal/modules/auth con el hasher real).
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

INSERT INTO staff_credential (staff_user_id, barbershop_id, password_hash, password_algorithm) VALUES
  ('aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaa1',
   '11111111-1111-1111-1111-111111111111',
   'fixture-no-es-un-hash-real-de-prueba-duena-a-0001', 'argon2id'),

  ('aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2',
   '11111111-1111-1111-1111-111111111111',
   'fixture-no-es-un-hash-real-de-prueba-barbero-a-0002', 'argon2id'),

  ('bbbbbbb1-bbbb-bbbb-bbbb-bbbbbbbbbbb1',
   '22222222-2222-2222-2222-222222222222',
   'fixture-no-es-un-hash-real-de-prueba-dueno-b-0003', 'argon2id'),

  ('bbbbbbb2-bbbb-bbbb-bbbb-bbbbbbbbbbb2',
   '22222222-2222-2222-2222-222222222222',
   'fixture-no-es-un-hash-real-de-prueba-barbero-b-0004', 'argon2id')
ON CONFLICT (staff_user_id) DO NOTHING;

-- Una sesión vigente por barbería, con identificadores e hash de token fijos
-- para que las pruebas de aislamiento puedan referenciarlos directamente.
-- El hash es SHA-256 de un valor ficticio ("token-ficticio-a"/"...-b"), no un
-- token real emitido por la aplicación.
INSERT INTO staff_session (id, barbershop_id, staff_user_id, token_hash, issued_at, expires_at, last_used_at) VALUES
  ('cccccca1-cccc-cccc-cccc-ccccccccccc1',
   '11111111-1111-1111-1111-111111111111',
   'aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaa1',
   encode(sha256('token-ficticio-a'::bytea), 'hex'),
   now() - interval '1 day', now() + interval '29 days', now() - interval '1 day'),

  ('cccccca2-cccc-cccc-cccc-ccccccccccc2',
   '22222222-2222-2222-2222-222222222222',
   'bbbbbbb1-bbbb-bbbb-bbbb-bbbbbbbbbbb1',
   encode(sha256('token-ficticio-b'::bytea), 'hex'),
   now() - interval '1 day', now() + interval '29 days', now() - interval '1 day')
ON CONFLICT (id) DO NOTHING;

COMMIT;
