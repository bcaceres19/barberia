-- Propósito
--   HU-008 (DEC-094, issue #278): `auth_recovery_request` deja de exigir un
--   teléfono verificado para aceptar la solicitud. El canal Correo solo
--   necesita una cuenta activa con ese correo; el requisito de contacto
--   verificado pasa a evaluarse únicamente sobre el canal elegido. El correo
--   de la cuenta se devuelve siempre; el teléfono se devuelve SOLO si está
--   verificado, de modo que el canal WhatsApp conserva la garantía de
--   `DEC-093` (nunca se envía a un número sin verificar) sin que la función
--   tenga que conocer el canal.
--
-- Reglas y decisiones
--   RN-DAT-01/RN-DAT-02 (minimización, sin dato personal en registros),
--   DEC-064 (cooldown, límite de reenvío y vigencia: sin cambios), DEC-065
--   (respuesta idéntica exista o no la cuenta), DEC-092 (canal elegido),
--   DEC-093 (WhatsApp exige número verificado con coincidencia única),
--   DEC-094 (correo basta).
--
-- Comportamiento
--   Antes: `accepted = true` solo con `is_active AND phone_verified_at IS NOT
--   NULL`. Ahora: `accepted = true` con `is_active`, sin mirar el teléfono.
--   `phone` del resultado es NULL cuando la cuenta no tiene teléfono o no
--   está verificado. Cooldown, límite por ventana e invalidación del código
--   vigente anterior no cambian.
--
-- Qué NO hace esta migración
--   No modifica `staff_user`, `staff_recovery_code` ni las demás funciones de
--   recuperación (`verify`, `current_credential`, `change_password` ya no
--   dependían del teléfono verificado), no agrega una marca de correo
--   verificado (alternativa descartada por DEC-094) y no cambia los grants.
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE,
--   20260811145252_harden_roles_and_definer_functions.sql). La función sigue
--   siendo `SECURITY DEFINER` con `search_path` vacío y `EXECUTE` solo para
--   `barberia_app`.
--
-- Plan de avance
--   Una corrección posterior se hace con otra migración, nunca editando esta
--   (DEC-036). No existe archivo `down`: el mecanismo normal es roll-forward.

-- ---------------------------------------------------------------------------
-- 1. Precondiciones
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF to_regprocedure('public.auth_recovery_request(text, text, integer, integer, integer, integer, integer)') IS NULL THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260817190000_create_staff_recovery_code.sql '
      '(auth_recovery_request); esa migración debe estar aplicada antes.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. `auth_recovery_request` sin exigir teléfono verificado
-- ---------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION auth_recovery_request(
  p_email                  text,
  p_code_hash              text,
  p_expires_seconds        integer,
  p_max_attempts           integer,
  p_resend_cooldown_seconds integer,
  p_resend_window_seconds  integer,
  p_resend_max_per_window  integer
)
RETURNS TABLE (accepted boolean, phone text, email text)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_now          timestamptz := pg_catalog.now();
  v_user         record;
  v_last_created timestamptz;
  v_recent       integer;
BEGIN
  IF p_email IS NULL OR p_code_hash IS NULL OR p_expires_seconds IS NULL
     OR p_max_attempts IS NULL OR p_resend_cooldown_seconds IS NULL
     OR p_resend_window_seconds IS NULL OR p_resend_max_per_window IS NULL THEN
    RAISE EXCEPTION 'auth_recovery_request: ningún argumento admite NULL.';
  END IF;
  IF char_length(p_code_hash) <> 64 THEN
    RAISE EXCEPTION 'auth_recovery_request: code_hash debe ser HMAC-SHA256 (64 hex).';
  END IF;
  IF p_expires_seconds < 1 OR p_max_attempts NOT BETWEEN 1 AND 10
     OR p_resend_cooldown_seconds < 1 OR p_resend_window_seconds < 1
     OR p_resend_max_per_window < 1 THEN
    RAISE EXCEPTION 'auth_recovery_request: parámetros fuera de rango.';
  END IF;

  -- DEC-094: el teléfono ya no filtra la cuenta. Se lee su verificación solo
  -- para decidir más abajo si el número puede devolverse como destino.
  SELECT u.id, u.barbershop_id, u.phone, u.phone_verified_at, u.email
  INTO v_user
  FROM public.staff_user u
  WHERE u.email = pg_catalog.lower(p_email)
    AND u.is_active;

  IF NOT FOUND THEN
    RETURN QUERY SELECT false, NULL::text, NULL::text;
    RETURN;
  END IF;

  SELECT max(created_at) INTO v_last_created
  FROM public.staff_recovery_code
  WHERE staff_user_id = v_user.id;

  IF v_last_created IS NOT NULL
     AND v_last_created > v_now - pg_catalog.make_interval(secs => p_resend_cooldown_seconds) THEN
    RETURN QUERY SELECT false, NULL::text, NULL::text;
    RETURN;
  END IF;

  SELECT count(*) INTO v_recent
  FROM public.staff_recovery_code
  WHERE staff_user_id = v_user.id
    AND created_at > v_now - pg_catalog.make_interval(secs => p_resend_window_seconds);

  IF v_recent >= p_resend_max_per_window THEN
    RETURN QUERY SELECT false, NULL::text, NULL::text;
    RETURN;
  END IF;

  -- Reenvío: invalida atómicamente cualquier código vigente anterior en la
  -- misma escritura que crea el nuevo (DEC-064, CA-008-07). El índice único
  -- parcial de staff_recovery_code exige esto: sin este UPDATE, el INSERT de
  -- abajo violaría la unicidad si ya existiera un código vigente.
  UPDATE public.staff_recovery_code
  SET invalidated_at = v_now
  WHERE staff_user_id = v_user.id
    AND verified_at IS NULL
    AND invalidated_at IS NULL;

  INSERT INTO public.staff_recovery_code
    (barbershop_id, staff_user_id, code_hash, max_attempts, expires_at)
  VALUES
    (v_user.barbershop_id, v_user.id, p_code_hash, p_max_attempts,
     v_now + pg_catalog.make_interval(secs => p_expires_seconds));

  -- Un teléfono sin verificar nunca sale como destino (DEC-093, DEC-094).
  RETURN QUERY SELECT
    true,
    CASE WHEN v_user.phone_verified_at IS NOT NULL THEN v_user.phone END,
    v_user.email;
END;
$$;

REVOKE ALL     ON FUNCTION auth_recovery_request(text, text, integer, integer, integer, integer, integer) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION auth_recovery_request(text, text, integer, integer, integer, integer, integer) TO barberia_app;

COMMENT ON FUNCTION auth_recovery_request(text, text, integer, integer, integer, integer, integer) IS
  'Único punto de escritura de solicitud de staff_recovery_code (DDL-AUT-01). accepted y '
  'email se completan cuando la cuenta activa existe y no se excede el cooldown/límite propio '
  'de reenvío; phone solo cuando además está verificado (DEC-093, DEC-094). En cualquier otro '
  'caso devuelve (false, NULL, NULL) sin crear fila, para que la capa HTTP responda siempre '
  'igual (CA-008-01, no enumeración). Invalida atómicamente el código vigente anterior del '
  'mismo usuario antes de insertar el nuevo (CA-008-07).';
