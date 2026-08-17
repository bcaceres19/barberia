-- Propósito
--   HU-008: recuperación de acceso con código de un solo uso. Crea
--   `staff_recovery_code` y sus cuatro funciones `SECURITY DEFINER`
--   (solicitar, verificar, leer la credencial vigente para la política de
--   contraseña, cambiar contraseña + revocar sesiones, purgar). Se aparta
--   deliberadamente de la sección A.3 de `database/modelo-fisico-referencia.sql`:
--   esa sección da a `barberia_app` GRANT directo de SELECT/INSERT/UPDATE/
--   DELETE bajo política tenant-aware, pero solicitar y verificar ocurren
--   por correo, ANTES de que exista `app.barbershop_id` (el mismo problema
--   que resolvieron `staff_credential`/`login_throttle`/
--   `auth_phone_challenge`); un GRANT directo tenant-aware no es alcanzable
--   en ese punto. Esta migración sigue en cambio el patrón ya probado de
--   `20260817180000_create_login_throttle_and_phone_challenge.sql`: sin
--   GRANT directo, todo el acceso pasa por funciones `SECURITY DEFINER`
--   estrechas que resuelven el usuario por correo internamente
--   (DDL-AUT-01).
--
-- Reglas y decisiones
--   RN-TEN-01, RN-DAT-01, RN-DAT-02, RN-REC-04, RN-REC-05, DEC-024 (esquema
--   compartido con RLS), DEC-026 (recuperación con código al teléfono
--   verificado), DEC-040 (roles separados), DEC-050 (mismo patrón de token
--   opaco + hash que la sesión), DEC-063 (política de contraseña, aplicada
--   en Go, no en SQL), DEC-064 (formato/vigencia/intentos/reenvío del
--   código y el token de reinicio), DEC-065 (destino enmascarado solo tras
--   verificar, aplicado en Go), DEC-066 (proveedor de entrega, fuera de
--   esta migración).
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE). El objeto nuevo
--   queda owned por `barberia_migrator`; la política RLS administrativa se
--   escribe `FOR ALL TO barberia_owner`. Las cuatro funciones son
--   `SECURITY DEFINER` con `search_path=''`, nombres calificados y grants
--   mínimos: `auth_recovery_request`/`verify`/`current_credential` van a
--   `barberia_app`; `auth_recovery_purge_expired` va exclusivamente a
--   `barberia_worker`. Ninguna función devuelve más que el mínimo
--   necesario: `auth_recovery_current_credential` expone el hash ya
--   codificado (nunca la contraseña) de UN único usuario ya autorizado por
--   un token de reinicio vigente, exactamente el mismo contrato que
--   `auth_get_credential` (HU-005) aplicado a este flujo.
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
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_owner')
     OR NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_app')
     OR NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_worker')
     OR NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_migrator') THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260811145252_harden_roles_and_definer_functions.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'staff_user' AND column_name = 'phone'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260817180000_create_login_throttle_and_phone_challenge.sql '
      '(staff_user.phone/phone_verified_at); esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'staff_recovery_code'
  ) THEN
    RAISE EXCEPTION
      'staff_recovery_code ya existe: esta migración no puede aplicarse dos veces.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. `staff_recovery_code`
-- ---------------------------------------------------------------------------
-- Una sola fila atraviesa las dos primeras etapas del flujo: nace al
-- solicitar (code_hash, expires_at), gana verified_at/reset_token_* al
-- verificar con éxito, y gana reset_token_consumed_at al cambiar la
-- contraseña. No es dos tablas separadas porque el token de reinicio no
-- tiene sentido fuera del código que lo autorizó (DEC-064).

CREATE TABLE staff_recovery_code (
  id                       uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id            uuid        NOT NULL,
  staff_user_id            uuid        NOT NULL,
  code_hash                text        NOT NULL,
  attempt_count            smallint    NOT NULL DEFAULT 0,
  max_attempts             smallint    NOT NULL DEFAULT 5,
  expires_at               timestamptz NOT NULL,
  verified_at              timestamptz,
  invalidated_at           timestamptz,
  reset_token_hash         text,
  reset_token_expires_at   timestamptz,
  reset_token_consumed_at  timestamptz,
  created_at               timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT staff_recovery_code_id_pk PRIMARY KEY (id),

  CONSTRAINT staff_recovery_code_barbershop_id_staff_user_id_fk
    FOREIGN KEY (barbershop_id, staff_user_id)
    REFERENCES staff_user (barbershop_id, id) ON DELETE CASCADE,

  CONSTRAINT staff_recovery_code_code_hash_ck     CHECK (char_length(code_hash) = 64),
  CONSTRAINT staff_recovery_code_attempt_count_ck CHECK (attempt_count >= 0 AND attempt_count <= max_attempts),
  CONSTRAINT staff_recovery_code_max_attempts_ck  CHECK (max_attempts BETWEEN 1 AND 10),
  CONSTRAINT staff_recovery_code_expires_at_ck    CHECK (expires_at > created_at),

  -- Un código no puede estar verificado e invalidado a la vez: son dos
  -- desenlaces distintos del mismo objeto (igual que auth_phone_challenge).
  CONSTRAINT staff_recovery_code_outcome_ck CHECK (
    NOT (verified_at IS NOT NULL AND invalidated_at IS NOT NULL)
  ),
  CONSTRAINT staff_recovery_code_verified_at_ck CHECK (
    verified_at IS NULL OR verified_at >= created_at
  ),
  CONSTRAINT staff_recovery_code_invalidated_at_ck CHECK (
    invalidated_at IS NULL OR invalidated_at >= created_at
  ),

  -- El token de reinicio solo existe si el código se verificó, y solo se
  -- consume si existe (DEC-064: "el tercer paso" no puede alcanzarse sin
  -- haber pasado por el segundo).
  CONSTRAINT staff_recovery_code_reset_token_requires_verified_ck CHECK (
    reset_token_hash IS NULL OR verified_at IS NOT NULL
  ),
  CONSTRAINT staff_recovery_code_reset_token_hash_ck CHECK (
    reset_token_hash IS NULL OR char_length(reset_token_hash) = 64
  ),
  CONSTRAINT staff_recovery_code_reset_token_expires_at_ck CHECK (
    reset_token_expires_at IS NULL OR (verified_at IS NOT NULL AND reset_token_expires_at > verified_at)
  ),
  CONSTRAINT staff_recovery_code_reset_token_consumed_ck CHECK (
    reset_token_consumed_at IS NULL
    OR (reset_token_hash IS NOT NULL AND reset_token_consumed_at >= verified_at)
  )
);

COMMENT ON TABLE staff_recovery_code IS
  'Códigos de recuperación de acceso y su token de reinicio de un solo uso (HU-008, '
  'DEC-064). Propietario funcional: barbería. Retención: corta, se purga tras vencer o '
  'consumirse. Clasificación: secreto. Solo se almacenan hashes: ni el código ni el token '
  'de reinicio en claro llegan nunca a esta tabla (CA-008-04).';

-- Un solo código vigente por usuario (CA-008-07): "vigente" significa que
-- todavía no se verificó ni se invalidó. Una vez verified_at se fija, la
-- fila deja de contar para esta unicidad (ya avanzó a la etapa del token de
-- reinicio), así que un reenvío posterior a una verificación exitosa no
-- choca con ella.
CREATE UNIQUE INDEX idx_staff_recovery_code_active_per_user
  ON staff_recovery_code (barbershop_id, staff_user_id)
  WHERE verified_at IS NULL AND invalidated_at IS NULL;

-- Auditoría completa (todos los desenlaces) y resolución de cooldown/tasa de
-- reenvío por staff_user_id: sin predicado, distinto del índice parcial de
-- arriba (DDL-PER-01).
CREATE INDEX idx_staff_recovery_code_shop_user_created
  ON staff_recovery_code (staff_user_id, created_at);

CREATE INDEX idx_staff_recovery_code_expires_at ON staff_recovery_code (expires_at);

ALTER TABLE staff_recovery_code ENABLE ROW LEVEL SECURITY;
ALTER TABLE staff_recovery_code FORCE  ROW LEVEL SECURITY;

CREATE POLICY staff_recovery_code_all_admin_policy ON staff_recovery_code
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

-- DDL-AUT-01: sin política de SELECT/INSERT/UPDATE ni GRANT para
-- barberia_app. Las cuatro funciones siguientes son el único punto de
-- acceso; las tres primeras corren ANTES de resolver contexto de tenant
-- (igual que A.0/A.2/A.4/A.5).

-- ---------------------------------------------------------------------------
-- 3. `auth_recovery_request`
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

  SELECT u.id, u.barbershop_id, u.phone, u.email
  INTO v_user
  FROM public.staff_user u
  WHERE u.email = pg_catalog.lower(p_email)
    AND u.is_active
    AND u.phone_verified_at IS NOT NULL;

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
  -- parcial de arriba exige esto: sin este UPDATE, el INSERT de abajo
  -- violaría la unicidad si ya existiera un código vigente.
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

  RETURN QUERY SELECT true, v_user.phone, v_user.email;
END;
$$;

REVOKE ALL     ON FUNCTION auth_recovery_request(text, text, integer, integer, integer, integer, integer) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION auth_recovery_request(text, text, integer, integer, integer, integer, integer) TO barberia_app;

COMMENT ON FUNCTION auth_recovery_request(text, text, integer, integer, integer, integer, integer) IS
  'Único punto de escritura de solicitud de staff_recovery_code (DDL-AUT-01). accepted, '
  'phone y email solo se completan cuando la cuenta existe con teléfono verificado y no se '
  'excede el cooldown/límite propio de reenvío; en cualquier otro caso devuelve '
  '(false, NULL, NULL) sin crear fila, para que la capa HTTP responda siempre igual '
  '(CA-008-01, no enumeración). Invalida atómicamente el código vigente anterior del mismo '
  'usuario antes de insertar el nuevo (CA-008-07).';

-- ---------------------------------------------------------------------------
-- 4. `auth_recovery_verify`
-- ---------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION auth_recovery_verify(
  p_email                text,
  p_code_hash            text,
  p_reset_token_hash     text,
  p_reset_expires_seconds integer
)
RETURNS TABLE (ok boolean, phone text, email text)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_now  timestamptz := pg_catalog.now();
  v_user record;
  v_row  public.staff_recovery_code%ROWTYPE;
BEGIN
  IF p_email IS NULL OR p_code_hash IS NULL OR p_reset_token_hash IS NULL
     OR p_reset_expires_seconds IS NULL THEN
    RAISE EXCEPTION 'auth_recovery_verify: ningún argumento admite NULL.';
  END IF;
  IF char_length(p_code_hash) <> 64 OR char_length(p_reset_token_hash) <> 64 THEN
    RAISE EXCEPTION 'auth_recovery_verify: code_hash/reset_token_hash deben tener 64 hex.';
  END IF;
  IF p_reset_expires_seconds < 1 THEN
    RAISE EXCEPTION 'auth_recovery_verify: p_reset_expires_seconds fuera de rango.';
  END IF;

  SELECT u.id, u.barbershop_id, u.phone, u.email
  INTO v_user
  FROM public.staff_user u
  WHERE u.email = pg_catalog.lower(p_email)
    AND u.is_active;

  IF NOT FOUND THEN
    RETURN QUERY SELECT false, NULL::text, NULL::text;
    RETURN;
  END IF;

  SELECT * INTO v_row
  FROM public.staff_recovery_code c
  WHERE c.barbershop_id = v_user.barbershop_id
    AND c.staff_user_id = v_user.id
    AND c.verified_at IS NULL
    AND c.invalidated_at IS NULL
    AND c.expires_at > v_now
  ORDER BY c.created_at DESC
  LIMIT 1
  FOR UPDATE;

  IF NOT FOUND THEN
    RETURN QUERY SELECT false, NULL::text, NULL::text;
    RETURN;
  END IF;

  IF v_row.code_hash <> p_code_hash THEN
    UPDATE public.staff_recovery_code
    SET attempt_count = attempt_count + 1,
        invalidated_at = CASE WHEN attempt_count + 1 >= max_attempts THEN v_now END
    WHERE id = v_row.id;
    RETURN QUERY SELECT false, NULL::text, NULL::text;
    RETURN;
  END IF;

  UPDATE public.staff_recovery_code
  SET verified_at            = v_now,
      reset_token_hash       = p_reset_token_hash,
      reset_token_expires_at = v_now + pg_catalog.make_interval(secs => p_reset_expires_seconds)
  WHERE id = v_row.id;

  RETURN QUERY SELECT true, v_user.phone, v_user.email;
END;
$$;

REVOKE ALL     ON FUNCTION auth_recovery_verify(text, text, text, integer) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION auth_recovery_verify(text, text, text, integer) TO barberia_app;

COMMENT ON FUNCTION auth_recovery_verify(text, text, text, integer) IS
  'Único punto de escritura de verificación (DDL-AUT-01). Código incorrecto incrementa '
  'attempt_count y, al agotar max_attempts, invalida; código correcto marca verified_at y '
  'emite el token de reinicio (hash ya calculado por Go) en la misma fila. Cuenta '
  'inexistente, código incorrecto, vencido o agotado devuelven exactamente '
  '(false, NULL, NULL): la capa HTTP no puede distinguirlos (CA-008-01, no enumeración). '
  'phone/email solo viajan en el desenlace exitoso, para que Go los enmascare (DEC-065): '
  'nunca se exponen en el desenlace false.';

-- ---------------------------------------------------------------------------
-- 5. `auth_recovery_current_credential`
-- ---------------------------------------------------------------------------
-- Lectura estrecha (mismo contrato que auth_get_credential de HU-005):
-- entrega el hash YA codificado de un único usuario, autorizado por un
-- token de reinicio vigente y no consumido, para que Go pueda rechazar en
-- DP-SEG-11/DEC-063 una contraseña nueva idéntica a la actual sin volver a
-- implementar Argon2id en SQL. No consume el token: eso lo hace
-- auth_recovery_change_password.

CREATE OR REPLACE FUNCTION auth_recovery_current_credential(
  p_email             text,
  p_reset_token_hash  text
)
RETURNS TABLE (found boolean, password_hash text, password_algorithm text)
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_user record;
BEGIN
  IF p_email IS NULL OR p_reset_token_hash IS NULL THEN
    RAISE EXCEPTION 'auth_recovery_current_credential: ningún argumento admite NULL.';
  END IF;
  IF char_length(p_reset_token_hash) <> 64 THEN
    RAISE EXCEPTION 'auth_recovery_current_credential: reset_token_hash debe tener 64 hex.';
  END IF;

  SELECT u.id, u.barbershop_id
  INTO v_user
  FROM public.staff_user u
  WHERE u.email = pg_catalog.lower(p_email)
    AND u.is_active;

  IF NOT FOUND THEN
    RETURN QUERY SELECT false, NULL::text, NULL::text;
    RETURN;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM public.staff_recovery_code c
    WHERE c.barbershop_id = v_user.barbershop_id
      AND c.staff_user_id = v_user.id
      AND c.reset_token_hash = p_reset_token_hash
      AND c.reset_token_consumed_at IS NULL
      AND c.reset_token_expires_at > pg_catalog.now()
  ) THEN
    RETURN QUERY SELECT false, NULL::text, NULL::text;
    RETURN;
  END IF;

  RETURN QUERY
    SELECT true, cr.password_hash, cr.password_algorithm
    FROM public.staff_credential cr
    WHERE cr.staff_user_id = v_user.id
      AND cr.barbershop_id = v_user.barbershop_id;
END;
$$;

REVOKE ALL     ON FUNCTION auth_recovery_current_credential(text, text) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION auth_recovery_current_credential(text, text) TO barberia_app;

COMMENT ON FUNCTION auth_recovery_current_credential(text, text) IS
  'Lectura estrecha y revisada (DDL-AUT-01), mismo contrato que auth_get_credential: '
  'expone el hash codificado de un único usuario, solo cuando p_reset_token_hash '
  'corresponde a un token de reinicio vigente y no consumido de ese usuario. No consume '
  'el token.';

-- ---------------------------------------------------------------------------
-- 6. `auth_recovery_change_password`
-- ---------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION auth_recovery_change_password(
  p_email             text,
  p_reset_token_hash  text,
  p_new_password_hash text
)
RETURNS boolean
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_now  timestamptz := pg_catalog.now();
  v_user record;
  v_row  public.staff_recovery_code%ROWTYPE;
BEGIN
  IF p_email IS NULL OR p_reset_token_hash IS NULL OR p_new_password_hash IS NULL THEN
    RAISE EXCEPTION 'auth_recovery_change_password: ningún argumento admite NULL.';
  END IF;
  IF char_length(p_reset_token_hash) <> 64 THEN
    RAISE EXCEPTION 'auth_recovery_change_password: reset_token_hash debe tener 64 hex.';
  END IF;
  IF char_length(p_new_password_hash) NOT BETWEEN 32 AND 512 THEN
    RAISE EXCEPTION 'auth_recovery_change_password: new_password_hash fuera de rango.';
  END IF;

  SELECT u.id, u.barbershop_id
  INTO v_user
  FROM public.staff_user u
  WHERE u.email = pg_catalog.lower(p_email)
    AND u.is_active;

  IF NOT FOUND THEN
    RETURN false;
  END IF;

  -- El artefacto se consume una vez incluso bajo dos solicitudes
  -- concurrentes: FOR UPDATE serializa la segunda detrás de la primera, que
  -- ya habrá fijado reset_token_consumed_at cuando la segunda llegue a esta
  -- fila (DEC-064, CA-008-05).
  SELECT * INTO v_row
  FROM public.staff_recovery_code c
  WHERE c.barbershop_id = v_user.barbershop_id
    AND c.staff_user_id = v_user.id
    AND c.reset_token_hash = p_reset_token_hash
    AND c.reset_token_consumed_at IS NULL
    AND c.reset_token_expires_at > v_now
  FOR UPDATE;

  IF NOT FOUND THEN
    RETURN false;
  END IF;

  UPDATE public.staff_recovery_code
  SET reset_token_consumed_at = v_now
  WHERE id = v_row.id;

  UPDATE public.staff_credential
  SET password_hash       = p_new_password_hash,
      password_algorithm  = 'argon2id',
      password_updated_at = v_now
  WHERE staff_user_id = v_user.id
    AND barbershop_id = v_user.barbershop_id;

  UPDATE public.staff_session
  SET revoked_at = v_now
  WHERE barbershop_id = v_user.barbershop_id
    AND staff_user_id = v_user.id
    AND revoked_at IS NULL;

  RETURN true;
END;
$$;

REVOKE ALL     ON FUNCTION auth_recovery_change_password(text, text, text) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION auth_recovery_change_password(text, text, text) TO barberia_app;

COMMENT ON FUNCTION auth_recovery_change_password(text, text, text) IS
  'Único punto de escritura del cambio de contraseña (DDL-AUT-01). Consume el token de '
  'reinicio, actualiza staff_credential y revoca TODAS las sesiones activas del usuario en '
  'la misma transacción implícita de la función (CA-008-05). Token desconocido, ya '
  'consumido o vencido devuelve false sin modificar nada.';

-- ---------------------------------------------------------------------------
-- 7. `auth_recovery_purge_expired`
-- ---------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION auth_recovery_purge_expired(p_limit integer)
RETURNS integer
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_deleted integer;
  v_now     timestamptz := pg_catalog.now();
BEGIN
  IF p_limit IS NULL OR p_limit < 1 OR p_limit > 1000 THEN
    RAISE EXCEPTION 'auth_recovery_purge_expired: p_limit fuera de rango (1-1000).';
  END IF;

  WITH due AS (
    SELECT id FROM public.staff_recovery_code
    WHERE invalidated_at IS NOT NULL
       OR (verified_at IS NULL AND expires_at <= v_now)
       OR (verified_at IS NOT NULL
           AND (reset_token_consumed_at IS NOT NULL OR reset_token_expires_at <= v_now))
    ORDER BY created_at
    LIMIT p_limit
    FOR UPDATE SKIP LOCKED
  )
  DELETE FROM public.staff_recovery_code
  WHERE id IN (SELECT id FROM due);

  GET DIAGNOSTICS v_deleted = ROW_COUNT;
  RETURN v_deleted;
END;
$$;

REVOKE ALL     ON FUNCTION auth_recovery_purge_expired(integer) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION auth_recovery_purge_expired(integer) TO barberia_worker;

COMMENT ON FUNCTION auth_recovery_purge_expired(integer) IS
  'Mantenimiento del worker: purga en lote los códigos de recuperación en un desenlace '
  'terminal (invalidado, expirado sin verificar, o con token de reinicio consumido/vencido). '
  'Exclusivo de barberia_worker.';
