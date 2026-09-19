-- Propósito
--   HU-008 (DEC-092, DEC-093): agrega `auth_recovery_resolve_phone`, la
--   función `SECURITY DEFINER` estrecha que resuelve un número de WhatsApp
--   escrito en el paso 1 de la recuperación de acceso al correo de la única
--   cuenta activa que lo tiene verificado. El servicio traduce el teléfono a
--   ese correo y reutiliza sin cambios `auth_recovery_request/verify/
--   current_credential/change_password`, que siguen identificando la cuenta
--   por correo.
--
-- Reglas y decisiones
--   RN-DAT-01/RN-DAT-02 (minimización, sin dato personal en registros),
--   RN-TEN-01 (aislamiento por barbería: el número no es único entre
--   barberías, así que una coincidencia ambigua no resuelve nada),
--   DEC-024 (esquema compartido con RLS), DEC-035/DEC-036 (estándar y roles
--   de Atlas), DEC-064 (parámetros del código), DEC-065 (respuesta
--   idéntica exista o no la cuenta), DEC-092 (la persona elige el canal y
--   escribe su valor), DEC-093 (resolución de `DP-SEG-14`: el teléfono
--   identifica la cuenta solo si coincide con exactamente una cuenta activa
--   con ese número verificado en todo el sistema).
--
-- Comportamiento
--   Devuelve el correo de la cuenta cuando hay exactamente UNA fila activa
--   con `phone = p_phone` y `phone_verified_at IS NOT NULL`. Con cero o con
--   más de una devuelve NULL: quien la llama no puede distinguir ambos casos
--   ni tampoco una cuenta inexistente de una inactiva (DEC-065). No expone
--   el identificador de barbería, el nombre ni ningún otro dato.
--
-- Qué NO hace esta migración
--   No modifica `staff_user` ni las funciones de recuperación existentes, no
--   agrega una restricción de unicidad del teléfono (varias barberías pueden
--   registrar el mismo número; DP-SEG-14) y no envía ningún código.
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE,
--   20260811145252_harden_roles_and_definer_functions.sql).
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
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'staff_user' AND column_name = 'phone_verified_at'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260817180000_create_login_throttle_and_phone_challenge.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM pg_proc
    WHERE proname = 'auth_recovery_resolve_phone'
  ) THEN
    RAISE EXCEPTION 'auth_recovery_resolve_phone ya existe: esta migración no puede aplicarse dos veces.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Índice de apoyo para la búsqueda por teléfono verificado
-- ---------------------------------------------------------------------------

-- Parcial y no único a propósito: el número puede repetirse entre barberías.
-- Cubre exactamente el predicado de la función; la tabla es pequeña, pero la
-- búsqueda corre en un endpoint público sin sesión y no debe degradarse con
-- el crecimiento de barberías.
CREATE INDEX idx_staff_user_verified_phone
  ON staff_user (phone)
  WHERE is_active AND phone_verified_at IS NOT NULL;

-- ---------------------------------------------------------------------------
-- 3. `auth_recovery_resolve_phone` — F-AUTH-02, DEC-092, DEC-093
-- ---------------------------------------------------------------------------

CREATE FUNCTION auth_recovery_resolve_phone(p_phone text)
RETURNS text
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = ''
AS $$
  -- `count(*) = 1` deja fuera tanto ninguna coincidencia como una ambigua.
  -- `min(email)` solo elige el único valor cuando hay una fila.
  SELECT CASE WHEN pg_catalog.count(*) = 1 THEN pg_catalog.min(u.email) END
  FROM public.staff_user u
  WHERE u.phone = p_phone
    AND u.is_active
    AND u.phone_verified_at IS NOT NULL
$$;

REVOKE ALL     ON FUNCTION auth_recovery_resolve_phone(text) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION auth_recovery_resolve_phone(text) TO barberia_app;

COMMENT ON FUNCTION auth_recovery_resolve_phone(text) IS
  'SECURITY DEFINER acotada: resuelve un número E.164 verificado al correo de la única cuenta '
  'activa que lo tiene (recuperación de acceso, DEC-092/DEC-093). Devuelve NULL con cero o con '
  'más de una coincidencia; quien la llama responde igual en ambos casos y ante una cuenta '
  'inexistente (DEC-065, RN-TEN-01). No expone barbería, nombre ni otro dato.';
