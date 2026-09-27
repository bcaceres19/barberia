-- HU-007 / issue #282: Twilio Verify must receive the stored verified phone
-- before it can validate a provider-managed OTP. This narrow function never
-- exposes the number to HTTP and only returns it for an active challenge
-- already bound to the same email and IP.
CREATE OR REPLACE FUNCTION auth_phone_challenge_destination(
  p_email text,
  p_ip_hash text
)
RETURNS text
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_now timestamptz := pg_catalog.now();
  v_user record;
BEGIN
  IF p_email IS NULL OR p_ip_hash IS NULL THEN
    RAISE EXCEPTION 'auth_phone_challenge_destination: ningún argumento admite NULL.';
  END IF;
  IF char_length(p_ip_hash) <> 64 THEN
    RAISE EXCEPTION 'auth_phone_challenge_destination: ip_hash debe ser HMAC-SHA256 (64 hex).';
  END IF;

  SELECT u.id, u.barbershop_id, u.phone INTO v_user
  FROM public.staff_user u
  WHERE u.email = pg_catalog.lower(p_email)
    AND u.is_active
    AND u.phone_verified_at IS NOT NULL;
  IF NOT FOUND THEN
    RETURN NULL;
  END IF;

  PERFORM 1
  FROM public.auth_phone_challenge c
  WHERE c.barbershop_id = v_user.barbershop_id
    AND c.staff_user_id = v_user.id
    AND c.ip_hash = p_ip_hash
    AND c.consumed_at IS NULL
    AND c.invalidated_at IS NULL
    AND c.expires_at > v_now
  ORDER BY c.created_at DESC
  LIMIT 1;
  IF NOT FOUND THEN
    RETURN NULL;
  END IF;
  RETURN v_user.phone;
END;
$$;

REVOKE ALL ON FUNCTION auth_phone_challenge_destination(text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION auth_phone_challenge_destination(text, text) TO barberia_app;

COMMENT ON FUNCTION auth_phone_challenge_destination(text, text) IS
  'Lectura SECURITY DEFINER estrecha para un proveedor externo de OTP: devuelve el teléfono verificado solo si hay reto activo ligado al mismo correo e IP; no expone datos a HTTP.';
