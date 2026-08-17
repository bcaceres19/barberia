-- Escenario controlado de HU-007: teléfono verificado para el reto que
-- desbloquea el login (DEC-062). Requiere testdata/dos_barberias.sql y
-- testdata/hu005_credenciales_sesiones.sql cargados primero.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu005_credenciales_sesiones.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu007_reto_telefonico.sql
--
-- Se ejecuta con el rol migrador/administrador, igual que los dos archivos
-- anteriores.
--
-- Los números de teléfono son ficticios (formato E.164 válido, sin
-- corresponder a una persona real, AGENTS.md).

BEGIN;

-- Dos UPDATE separados a propósito: staff_user_reset_phone_verification_trg
-- (BEFORE UPDATE OF phone) anula phone_verified_at cada vez que la columna
-- phone participa en el UPDATE, aunque el mismo UPDATE también intente
-- fijar phone_verified_at en la misma sentencia (DDL-INT-05). Fijar el
-- teléfono y verificarlo en un solo UPDATE dejaría phone_verified_at en
-- NULL sin ningún error visible.
UPDATE staff_user SET phone = '+573000000001'
WHERE id = 'aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaa1';
UPDATE staff_user SET phone_verified_at = now()
WHERE id = 'aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaa1';

UPDATE staff_user SET phone = '+573000000003'
WHERE id = 'bbbbbbb1-bbbb-bbbb-bbbb-bbbbbbbbbbb1';
UPDATE staff_user SET phone_verified_at = now()
WHERE id = 'bbbbbbb1-bbbb-bbbb-bbbb-bbbbbbbbbbb1';

-- aaaaaaa2 (Barbero A) y bbbbbbb2 (Barbero B, inactivo) quedan SIN teléfono
-- verificado a propósito: las pruebas de auth_phone_challenge_request deben
-- demostrar que una cuenta sin teléfono verificado no recibe código, con la
-- misma respuesta 202 genérica que una cuenta inexistente.

COMMIT;
