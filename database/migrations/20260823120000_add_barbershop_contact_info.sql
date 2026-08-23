-- Propósito
--   HU-020: agrega el contacto opcional de la barbería (correo y teléfono)
--   a la tabla ya aplicada `barbershop`. `name` y `timezone` ya existen desde
--   20260807170000_create_tenant_foundation.sql; esta migración no los toca.
--   `public_slug` (sección B.1 de modelo-fisico-referencia.sql) queda fuera:
--   pertenece a F-PUB-01, no a esta historia.
--
-- Reglas y decisiones
--   RN-DIS-07 (toda hora respeta la zona configurada, sin cambios aquí),
--   RN-TEN-01, RN-DAT-02, DEC-024 (esquema compartido con RLS), DEC-035
--   (estándar de base de datos), DEC-036 (Atlas, migración inmutable),
--   DEC-040 (modelo de roles).
--
-- Qué NO hace esta migración
--   No valida la zona IANA (ya lo advierte el comentario de
--   barbershop.timezone en la migración fundacional: no es expresable en un
--   CHECK inmutable; lo valida la aplicación). No agrega `public_slug`, logo
--   ni ningún campo de F-PUB-01. No concede ningún GRANT nuevo: `barbershop`
--   ya tiene `GRANT SELECT, UPDATE ... TO barberia_app`
--   (20260807170000_create_tenant_foundation.sql), y un privilegio a nivel
--   de tabla cubre columnas agregadas después sin necesidad de repetirlo.
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE, ver
--   20260811145252_harden_roles_and_definer_functions.sql). Las columnas
--   nuevas quedan en la misma tabla, ya propiedad de `barberia_owner`; no se
--   crea ninguna política nueva porque `barbershop_select_tenant_policy` y
--   `barbershop_update_tenant_policy` ya cubren la fila completa, columnas
--   incluidas.
--
-- Plan de avance
--   Una corrección posterior se hace con otra migración, nunca editando
--   esta (DEC-036). No existe archivo `down`: el mecanismo normal es
--   roll-forward.

-- ---------------------------------------------------------------------------
-- 1. Precondiciones
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'barbershop'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260807170000_create_tenant_foundation.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barbershop' AND column_name = 'contact_email'
  ) THEN
    RAISE EXCEPTION 'barbershop.contact_email ya existe: esta migración no puede aplicarse dos veces.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Columnas nuevas
-- ---------------------------------------------------------------------------

ALTER TABLE barbershop
  ADD COLUMN contact_email text,
  ADD COLUMN contact_phone text;

-- Mismo patrón de normalización/forma que staff_user_email_ck (migración
-- fundacional): minúsculas, forma mínima de correo, largo acotado. Anulable:
-- un valor vacío del formulario se persiste como NULL (CA-020-06), nunca
-- como cadena vacía.
ALTER TABLE barbershop
  ADD CONSTRAINT barbershop_contact_email_ck CHECK (
    contact_email IS NULL OR (
      contact_email = lower(contact_email)
      AND contact_email LIKE '_%@_%._%'
      AND char_length(contact_email) <= 254
    )
  );

-- Mismo patrón E.164 que staff_user.phone
-- (20260817180000_create_login_throttle_and_phone_challenge.sql): '+',
-- indicativo sin cero inicial, 8 a 15 dígitos en total.
ALTER TABLE barbershop
  ADD CONSTRAINT barbershop_contact_phone_ck CHECK (
    contact_phone IS NULL OR contact_phone ~ '^\+[1-9][0-9]{7,14}$'
  );

COMMENT ON COLUMN barbershop.contact_email IS
  'Correo de contacto opcional de la barbería (HU-020, CA-020-06). Normalizado a minúsculas por '
  'el servicio antes de persistir; NULL cuando el formulario lo deja vacío. Propietario funcional: '
  'barbería. Retención: mientras exista la cuenta. Clasificación: datos de negocio (no es el correo '
  'de acceso de ningún staff_user).';

COMMENT ON COLUMN barbershop.contact_phone IS
  'Teléfono de contacto opcional de la barbería (HU-020, CA-020-06), en formato E.164, normalizado '
  'por el servicio antes de persistir. NULL cuando el formulario lo deja vacío. Propietario '
  'funcional: barbería. Retención: mientras exista la cuenta. Clasificación: datos de negocio.';
