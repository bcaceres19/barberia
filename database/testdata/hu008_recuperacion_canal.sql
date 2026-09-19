-- Escenario controlado de HU-008 con elección de canal (DEC-092, DEC-093):
-- cuentas con teléfono verificado propio para el recorrido por WhatsApp y un
-- número compartido entre dos barberías para probar la ambigüedad
-- (DP-SEG-14). Usa cuentas que ninguna otra suite de recuperación toca
-- (duena.a y dueno.b siguen reservadas para el recorrido por correo, con su
-- cooldown propio). Requiere testdata/dos_barberias.sql y
-- testdata/hu042_bloqueos.sql y testdata/hu060_citas.sql cargados primero.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu008_recuperacion_canal.sql
--
-- Los números son ficticios (E.164 válido, sin corresponder a una persona
-- real, AGENTS.md).

BEGIN;

-- Dos UPDATE por cuenta a propósito: staff_user_reset_phone_verification_trg
-- anula phone_verified_at cada vez que phone participa en el UPDATE
-- (DDL-INT-05), así que se fija el teléfono y luego se verifica.

-- dueno.o (barbería O): número único → identificable por WhatsApp.
UPDATE staff_user SET phone = '+573000000011'
WHERE email = 'dueno.o@ejemplo.test';
UPDATE staff_user SET phone_verified_at = now()
WHERE email = 'dueno.o@ejemplo.test';

-- Credencial ficticia para poder completar el paso 3 por WhatsApp. Igual que
-- en hu005_credenciales_sesiones.sql, no es un hash argon2id real: solo
-- cumple staff_credential_password_hash_ck.
INSERT INTO staff_credential (staff_user_id, barbershop_id, password_hash, password_algorithm)
VALUES ('b10c9001-9001-9001-9001-900190019001',
        'b10c0001-b10c-b10c-b10c-b10c00010001',
        'fixture-no-es-un-hash-real-de-prueba-dueno-o-0011', 'argon2id')
ON CONFLICT (staff_user_id) DO NOTHING;

-- dueno.p (barbería P) y dueno.q (barbería Q): MISMO número verificado →
-- ambiguo, no debe identificar ninguna cuenta.
UPDATE staff_user SET phone = '+573000000012'
WHERE email IN ('dueno.p@ejemplo.test', 'dueno.q@ejemplo.test');
UPDATE staff_user SET phone_verified_at = now()
WHERE email IN ('dueno.p@ejemplo.test', 'dueno.q@ejemplo.test');

COMMIT;
