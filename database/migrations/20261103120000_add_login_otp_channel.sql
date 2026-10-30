-- Propósito
--   DEC-139 / DEC-081 / HU-007: cada barbería elige por qué canal le llega el
--   código del reto adicional de acceso: correo, WhatsApp oficial o ambos. Se
--   agrega `barbershop.login_otp_channel` y `auth_phone_challenge_request`
--   devuelve los destinos que corresponden a ese canal (teléfono verificado y/o
--   correo de la cuenta) en lugar de exigir siempre un teléfono verificado.
--
-- Reglas y decisiones
--   RN-TEN-01, RN-DAT-02, DEC-024 (esquema compartido con RLS), DEC-035/DEC-036
--   (estándar de base de datos, Atlas, migración inmutable), DEC-062 (reto
--   telefónico), DEC-065 (no enumeración), DEC-081 (canal configurable) y
--   DEC-139 (decisión de esta migración).
--
-- Modelo
--   `login_otp_channel text NOT NULL` con `CHECK` sobre `email`, `whatsapp` y
--   `both`. Es un atributo de la barbería (una preferencia por barbería, la
--   recuperación de acceso no se configura: DEC-092), así que vive en su fila:
--   3FN intacta, sin `jsonb` ni tabla nueva. Las barberías que ya existen
--   quedan en `whatsapp`, que es como se entregaba hasta ahora; las nuevas
--   nacen en `email` (DEC-081: correo por defecto).
--
-- Cambio de la función
--   `auth_phone_challenge_request` pasa de `(accepted, phone)` a
--   `(accepted, phone, email)`. `phone` solo viene lleno si el canal incluye
--   WhatsApp y la cuenta tiene teléfono verificado; `email` solo si el canal
--   incluye correo. Sin ningún destino utilizable devuelve `(false, NULL, NULL)`
--   igual que una cuenta inexistente, una IP no escalada o un límite excedido:
--   la capa HTTP responde siempre 202 sin distinguir el motivo (DEC-065). Nunca
--   cae a un canal no configurado (DEC-081). Cambiar las columnas de salida
--   obliga a eliminar y volver a crear la función; se restablecen sus
--   privilegios exactos.
--
-- Qué NO hace esta migración
--   No toca `auth_phone_challenge_verify` (la verificación no depende del
--   canal), ni la recuperación de acceso, ni RLS: `barbershop` ya tiene RLS
--   forzada y `barberia_app` ya actualiza la fila de su barbería.
--
-- Clasificación de datos
--   El canal elegido no es un dato personal. El teléfono y el correo salen de la
--   función solo hacia el servidor, que los usa para entregar el código y no los
--   registra (RN-DAT-02).
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE, ver
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
    WHERE table_schema = 'public' AND table_name = 'barbershop' AND column_name = 'google_place_id'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20261102120000_add_google_place_id.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barbershop' AND column_name = 'login_otp_channel'
  ) THEN
    RAISE EXCEPTION
      'barbershop.login_otp_channel ya existe; esta migración no se puede aplicar dos veces.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_catalog.pg_proc p
    JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
    WHERE n.nspname = 'public' AND p.proname = 'auth_phone_challenge_request'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260817180000_create_login_throttle_and_phone_challenge.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Canal por barbería
-- ---------------------------------------------------------------------------

-- Las filas existentes toman 'whatsapp' (como se entregaba hasta ahora) y el
-- valor por defecto para las nuevas pasa a 'email' (DEC-081).
ALTER TABLE barbershop
  ADD COLUMN login_otp_channel text NOT NULL DEFAULT 'whatsapp',
  ADD CONSTRAINT barbershop_login_otp_channel_ck
    CHECK (login_otp_channel IN ('email', 'whatsapp', 'both'));

ALTER TABLE barbershop ALTER COLUMN login_otp_channel SET DEFAULT 'email';

COMMENT ON COLUMN barbershop.login_otp_channel IS
  'Canal del código del reto adicional de acceso (DEC-139): email, whatsapp o both. '
  'Correo es el valor de las barberías nuevas; las anteriores quedaron en whatsapp.';

-- ---------------------------------------------------------------------------
-- 3. auth_phone_challenge_request devuelve los destinos del canal
-- ---------------------------------------------------------------------------

DROP FUNCTION auth_phone_challenge_request(text, text, text, integer, integer, integer, integer);

CREATE FUNCTION auth_phone_challenge_request(
  p_email                text,
  p_ip_hash               text,
  p_code_hash             text,
  p_expires_seconds       integer,
  p_rate_window_seconds   integer,
  p_rate_max_active       integer,
  p_rate_cooldown_seconds integer
)
RETURNS TABLE (accepted boolean, phone text, email text)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_now       timestamptz := pg_catalog.now();
  v_user      record;
  v_phone     text;
  v_email     text;
  v_escalated boolean;
  v_recent    integer;
BEGIN
  IF p_email IS NULL OR p_ip_hash IS NULL OR p_code_hash IS NULL
     OR p_expires_seconds IS NULL OR p_rate_window_seconds IS NULL
     OR p_rate_max_active IS NULL OR p_rate_cooldown_seconds IS NULL THEN
    RAISE EXCEPTION 'auth_phone_challenge_request: ningún argumento admite NULL.';
  END IF;
  IF char_length(p_ip_hash) <> 64 OR char_length(p_code_hash) <> 64 THEN
    RAISE EXCEPTION 'auth_phone_challenge_request: ip_hash/code_hash deben ser HMAC-SHA256 (64 hex).';
  END IF;
  IF p_expires_seconds < 1 OR p_rate_window_seconds < 1
     OR p_rate_max_active < 1 OR p_rate_cooldown_seconds < 1 THEN
    RAISE EXCEPTION 'auth_phone_challenge_request: parámetros fuera de rango.';
  END IF;

  SELECT u.id, u.barbershop_id, u.email, u.phone,
         (u.phone_verified_at IS NOT NULL) AS phone_verified,
         b.login_otp_channel AS channel
  INTO v_user
  FROM public.staff_user u
  JOIN public.barbershop b ON b.id = u.barbershop_id
  WHERE u.email = pg_catalog.lower(p_email)
    AND u.is_active;

  IF NOT FOUND THEN
    RETURN QUERY SELECT false, NULL::text, NULL::text;
    RETURN;
  END IF;

  -- Solo el canal configurado y solo contactos de la cuenta; ningún destino
  -- viene del llamador ni se sustituye por otro canal (DEC-081).
  IF v_user.channel IN ('whatsapp', 'both') AND v_user.phone_verified THEN
    v_phone := v_user.phone;
  END IF;
  IF v_user.channel IN ('email', 'both') THEN
    v_email := v_user.email;
  END IF;

  IF v_phone IS NULL AND v_email IS NULL THEN
    RETURN QUERY SELECT false, NULL::text, NULL::text;
    RETURN;
  END IF;

  SELECT (t.escalated_until IS NOT NULL AND t.escalated_until > v_now)
  INTO v_escalated
  FROM public.login_throttle t
  WHERE t.ip_hash = p_ip_hash;

  IF NOT FOUND OR NOT v_escalated THEN
    RETURN QUERY SELECT false, NULL::text, NULL::text;
    RETURN;
  END IF;

  IF EXISTS (
    SELECT 1 FROM public.auth_phone_challenge c
    WHERE c.ip_hash = p_ip_hash
      AND c.created_at > v_now - pg_catalog.make_interval(secs => p_rate_cooldown_seconds)
  ) THEN
    RETURN QUERY SELECT false, NULL::text, NULL::text;
    RETURN;
  END IF;

  SELECT count(*) INTO v_recent
  FROM public.auth_phone_challenge c
  WHERE c.ip_hash = p_ip_hash
    AND c.created_at > v_now - pg_catalog.make_interval(secs => p_rate_window_seconds);

  IF v_recent >= p_rate_max_active THEN
    RETURN QUERY SELECT false, NULL::text, NULL::text;
    RETURN;
  END IF;

  INSERT INTO public.auth_phone_challenge
    (barbershop_id, staff_user_id, ip_hash, code_hash, expires_at)
  VALUES
    (v_user.barbershop_id, v_user.id, p_ip_hash, p_code_hash,
     v_now + pg_catalog.make_interval(secs => p_expires_seconds));

  RETURN QUERY SELECT true, v_phone, v_email;
END;
$$;

REVOKE ALL     ON FUNCTION auth_phone_challenge_request(text, text, text, integer, integer, integer, integer) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION auth_phone_challenge_request(text, text, text, integer, integer, integer, integer) TO barberia_app;

COMMENT ON FUNCTION auth_phone_challenge_request(text, text, text, integer, integer, integer, integer) IS
  'Único punto de escritura de auth_phone_challenge para solicitudes (DDL-AUT-01). accepted '
  'solo es true cuando la cuenta está activa y el canal de su barbería (login_otp_channel, '
  'DEC-139) tiene un destino utilizable (teléfono verificado para WhatsApp, correo de la '
  'cuenta para correo), la IP está realmente escalada y no se excede el límite propio de '
  'reenvío/solicitudes activas; phone y email solo se completan para los canales '
  'configurados. En cualquier otro caso devuelve (false, NULL, NULL) sin crear fila, para '
  'que la capa HTTP responda siempre 202 sin distinguir el motivo (no enumeración).';
