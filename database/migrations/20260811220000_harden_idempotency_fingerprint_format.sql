-- Propósito
--   Corrección roll-forward pendiente desde el issue #3, registrada como
--   comentario en el issue #7 (docs/05-backend/revision-ddl-seguridad-2026-08-11.md,
--   DDL-VAL-01): `idempotency_record.request_fingerprint` se validaba solo
--   por longitud, no por formato ni codificación.
--
-- Reglas y decisiones
--   DDL-VAL-01 (forma canónica explícita, no solo longitud).
--
-- Qué NO hace esta migración
--   No cambia el alcance de la huella (sigue sin incluir `operation`,
--   DEC-036: inmutable el diseño original). No cambia el algoritmo de hash
--   usado por el llamador: solo exige que, sea cual sea, se codifique como
--   hexadecimal en minúsculas antes de guardarse, igual que
--   `appointment_access_token.token_hash` y `staff_session.token_hash`.
--
-- Condición de seguridad
--   `ALTER TABLE ... DROP/ADD CONSTRAINT` no reescribe la tabla (no cambia
--   tipo ni agrega columna); valida el `CHECK` nuevo contra las filas
--   existentes. Si una fila ya guardada no fuera hexadecimal en minúsculas,
--   la migración falla en vez de aplicarse a medias: preferible a aceptar
--   datos que ya incumplían el formato documentado.

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'idempotency_record'
  ) THEN
    RAISE EXCEPTION
      'Esta migración corrige 20260807170100_create_idempotency_record.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;
END
$$;

-- La longitud (32-128) ya cubría MD5/SHA-1/SHA-256/SHA-384/SHA-512 en
-- hexadecimal; el CHECK original solo comprobaba el rango, no que el
-- contenido fuera realmente hexadecimal en minúsculas.
ALTER TABLE idempotency_record
  DROP CONSTRAINT idempotency_record_request_fingerprint_ck,
  ADD CONSTRAINT idempotency_record_request_fingerprint_ck CHECK (
    request_fingerprint ~ '^[0-9a-f]{32,128}$'
  );

COMMENT ON COLUMN idempotency_record.request_fingerprint IS
  'Huella del contenido de la solicitud (hash hexadecimal en minúsculas de método, ruta y cuerpo '
  'canonizado; 32-128 caracteres, según el algoritmo). Comparar huellas es lo que distingue '
  '"mismo contenido" de "contenido distinto" sin volver a guardar la solicitud completa. '
  'DDL-VAL-01: formato validado explícitamente, no solo longitud.';
