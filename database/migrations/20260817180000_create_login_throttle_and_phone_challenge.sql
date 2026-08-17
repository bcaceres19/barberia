-- Propósito
--   HU-007: defensa escalonada contra abuso del inicio de sesión. Crea
--   `login_throttle` (conteo por IP, sin barbershop_id ni RLS por diseño) y
--   `auth_phone_challenge` (reto telefónico que desbloquea el login tras el
--   escalamiento). Adelanta desde A.3 las columnas `staff_user.phone` y
--   `phone_verified_at` (y su trigger de reinicio), porque HU-007 las
--   necesita como condición del reto y se construye antes que HU-008 en el
--   orden de B0; HU-008 las reutilizará sin volver a crearlas. Copia
--   revisada de las secciones A.3 (solo el prerrequisito de teléfono) y A.5
--   de `database/modelo-fisico-referencia.sql` contra el estado real
--   aplicado (seis migraciones previas, incluida HU-012).
--
-- Reglas y decisiones
--   RN-DAT-02 (nada sensible en logs), DEC-026 (defensa contra abuso con
--   reto telefónico), DEC-040 (roles separados: `barberia_app` sin DML
--   directo sobre `login_throttle`/`auth_phone_challenge`, solo funciones
--   `SECURITY DEFINER` estrechas; `barberia_worker` solo purga), DEC-052
--   (ventana 15 min, escalamiento 24 h), DEC-061 (la sexta solicitud exige
--   el reto, no la quinta: `attempt_count + 1 > p_threshold`), DEC-062
--   (reto telefónico completo: `auth_phone_challenge`, HMAC-SHA256 del
--   código con secreto de despliegue, límite propio de reenvío/solicitudes
--   activas, verificación que limpia el escalamiento en la misma
--   transacción), DDL-AUT-01 (sin GRANT directo a `barberia_app` sobre
--   ninguna de las dos tablas: toda la lógica vive en funciones estrechas).
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE), igual que las
--   migraciones anteriores. Los objetos nuevos quedan owned por
--   `barberia_migrator`; la política RLS administrativa de
--   `auth_phone_challenge` se escribe `FOR ALL TO barberia_owner`.
--   `login_throttle_register_attempt`/`purge_expired` y
--   `auth_phone_challenge_request`/`verify`/`purge_expired` son
--   `SECURITY DEFINER` con `search_path=''`, nombres calificados y grants
--   mínimos: ninguna devuelve más que el mínimo necesario (un booleano y,
--   solo en el camino aceptado de la solicitud del reto, el teléfono de UN
--   único usuario ya resuelto por correo, nunca una lista).
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
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'staff_user'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260807170000_create_tenant_foundation.sql (staff_user); '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'staff_user' AND column_name = 'phone'
  ) THEN
    RAISE EXCEPTION
      'staff_user.phone ya existe: esta migración no puede aplicarse dos veces ni después '
      'de una migración de HU-008 que ya la haya creado.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Prerrequisito: teléfono verificado en `staff_user` (adelantado de A.3)
-- ---------------------------------------------------------------------------

ALTER TABLE staff_user
  ADD COLUMN phone             text,
  ADD COLUMN phone_verified_at timestamptz,
  ADD CONSTRAINT staff_user_phone_ck CHECK (
    phone IS NULL OR (phone ~ '^\+[1-9][0-9]{7,14}$')
  ),
  ADD CONSTRAINT staff_user_phone_verified_at_ck CHECK (
    phone_verified_at IS NULL OR phone IS NOT NULL
  );

COMMENT ON COLUMN staff_user.phone IS
  'Teléfono en formato E.164. Destino del reto de HU-007 y del código de recuperación de '
  'HU-008; se muestra siempre enmascarado y nunca completo en respuestas ni registros '
  '(CA-008-06).';

CREATE OR REPLACE FUNCTION staff_user_reset_phone_verification()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $$
BEGIN
  IF NEW.phone IS DISTINCT FROM OLD.phone THEN
    NEW.phone_verified_at := NULL;
  END IF;
  RETURN NEW;
END;
$$;

REVOKE ALL     ON FUNCTION staff_user_reset_phone_verification() FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION staff_user_reset_phone_verification() TO barberia_app;

COMMENT ON FUNCTION staff_user_reset_phone_verification() IS
  'Anula phone_verified_at cuando phone cambia de valor (DDL-INT-05). Sin esto, un número '
  'nuevo heredaría la verificación del número anterior.';

CREATE TRIGGER staff_user_reset_phone_verification_trg
  BEFORE UPDATE OF phone ON staff_user
  FOR EACH ROW EXECUTE FUNCTION staff_user_reset_phone_verification();

-- ---------------------------------------------------------------------------
-- 3. A.4 · `login_throttle`
-- ---------------------------------------------------------------------------
-- Única tabla del sistema SIN barbershop_id y SIN RLS: el conteo ocurre
-- antes de saber quién es el solicitante, y asociarlo a una barbería
-- permitiría a un atacante repartir sus intentos entre barberías para
-- diluir el límite.

CREATE TABLE login_throttle (
  ip_hash            text        NOT NULL,
  window_started_at  timestamptz NOT NULL DEFAULT now(),
  attempt_count      integer     NOT NULL DEFAULT 0,
  escalated_until    timestamptz,
  expires_at         timestamptz NOT NULL,

  CONSTRAINT login_throttle_ip_hash_pk    PRIMARY KEY (ip_hash),
  CONSTRAINT login_throttle_ip_hash_ck    CHECK (char_length(ip_hash) = 64),
  CONSTRAINT login_throttle_attempt_ck    CHECK (attempt_count >= 0),
  CONSTRAINT login_throttle_expires_at_ck CHECK (expires_at > window_started_at)
);

COMMENT ON TABLE login_throttle IS
  'Contadores del límite por IP del formulario de acceso (DEC-026, DEC-052). Propietario '
  'funcional: plataforma. Retención: hasta expires_at. Clasificación: técnico. ip_hash es '
  'HMAC-SHA256(secreto de despliegue, ip) (DDL-AUT-01): un hash simple, aunque con sal no '
  'pública, sigue siendo reconstruible por fuerza bruta porque el espacio de IP es pequeño '
  '(~2^32 para IPv4); HMAC con clave secreta de despliegue no lo es. La clave nunca vive en '
  'el repositorio ni en esta base de datos. Sin barbershop_id ni RLS por diseño.';

CREATE INDEX idx_login_throttle_expires_at ON login_throttle (expires_at);

-- DDL-AUT-01: sin GRANT directo a barberia_app. DML completo (en particular
-- UPDATE/DELETE) permitiría a una inyección SQL reiniciar el contador de
-- cualquier IP y anular el control de fuerza bruta.

CREATE OR REPLACE FUNCTION login_throttle_register_attempt(
  p_ip_hash            text,
  p_window_seconds     integer,
  p_escalation_seconds integer,
  p_threshold          integer,
  p_retention_seconds  integer
)
RETURNS TABLE (attempt_count integer, escalated boolean, retry_after timestamptz)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_now  timestamptz := pg_catalog.now();
  v_row  public.login_throttle%ROWTYPE;
BEGIN
  IF p_ip_hash IS NULL OR p_window_seconds IS NULL OR p_escalation_seconds IS NULL
     OR p_threshold IS NULL OR p_retention_seconds IS NULL THEN
    RAISE EXCEPTION 'login_throttle_register_attempt: ningún argumento admite NULL.';
  END IF;
  IF char_length(p_ip_hash) <> 64 THEN
    RAISE EXCEPTION 'login_throttle_register_attempt: ip_hash debe ser HMAC-SHA256 (64 hex).';
  END IF;
  IF p_window_seconds < 1 OR p_escalation_seconds < 1 OR p_threshold < 1
     OR p_retention_seconds < p_escalation_seconds THEN
    RAISE EXCEPTION 'login_throttle_register_attempt: parámetros fuera de rango.';
  END IF;

  -- Un solo INSERT ... ON CONFLICT DO UPDATE atómico, sin SELECT previo.
  -- Probado bajo concurrencia real en la revisión DDL: una versión con
  -- SELECT ... FOR UPDATE seguido de INSERT/UPDATE perdía incrementos,
  -- porque FOR UPDATE no bloquea nada cuando la fila todavía no existe.
  INSERT INTO public.login_throttle (ip_hash, window_started_at, attempt_count, escalated_until, expires_at)
  VALUES (p_ip_hash, v_now, 1, NULL, v_now + pg_catalog.make_interval(secs => p_retention_seconds))
  ON CONFLICT (ip_hash) DO UPDATE SET
    window_started_at = CASE
      WHEN login_throttle.window_started_at + pg_catalog.make_interval(secs => p_window_seconds) <= v_now
        THEN v_now
      ELSE login_throttle.window_started_at
    END,
    attempt_count = CASE
      WHEN login_throttle.window_started_at + pg_catalog.make_interval(secs => p_window_seconds) <= v_now
        THEN 1
      ELSE login_throttle.attempt_count + 1
    END,
    -- CA-007-04: la ventana vencida reinicia el conteo, pero no exime de un
    -- escalamiento todavía vigente. Dentro de la ventana, escala al
    -- SUPERAR el umbral (DEC-061/CT-005): p_threshold solicitudes se
    -- evalúan con normalidad; la solicitud p_threshold+1 exige el reto.
    escalated_until = CASE
      WHEN login_throttle.window_started_at + pg_catalog.make_interval(secs => p_window_seconds) <= v_now
        THEN CASE WHEN login_throttle.escalated_until > v_now THEN login_throttle.escalated_until END
      WHEN login_throttle.attempt_count + 1 > p_threshold
        THEN v_now + pg_catalog.make_interval(secs => p_escalation_seconds)
      ELSE login_throttle.escalated_until
    END,
    expires_at = v_now + pg_catalog.make_interval(secs => p_retention_seconds)
  RETURNING * INTO v_row;

  RETURN QUERY SELECT
    v_row.attempt_count,
    (v_row.escalated_until IS NOT NULL AND v_row.escalated_until > v_now),
    v_row.escalated_until;
END;
$$;

REVOKE ALL     ON FUNCTION login_throttle_register_attempt(text, integer, integer, integer, integer) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION login_throttle_register_attempt(text, integer, integer, integer, integer) TO barberia_app;

COMMENT ON FUNCTION login_throttle_register_attempt(text, integer, integer, integer, integer) IS
  'Único punto de escritura de login_throttle (DDL-AUT-01). Incrementa o reinicia el '
  'contador de una IP de forma atómica y calcula el escalamiento; barberia_app no tiene '
  'INSERT/UPDATE/DELETE directo sobre la tabla.';

CREATE OR REPLACE FUNCTION login_throttle_purge_expired(p_limit integer)
RETURNS integer
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_deleted integer;
BEGIN
  IF p_limit IS NULL OR p_limit < 1 OR p_limit > 1000 THEN
    RAISE EXCEPTION 'login_throttle_purge_expired: p_limit fuera de rango (1-1000).';
  END IF;

  WITH due AS (
    SELECT ip_hash FROM public.login_throttle
    WHERE expires_at <= pg_catalog.now()
    ORDER BY expires_at
    LIMIT p_limit
    FOR UPDATE SKIP LOCKED
  )
  DELETE FROM public.login_throttle
  WHERE ip_hash IN (SELECT ip_hash FROM due);

  GET DIAGNOSTICS v_deleted = ROW_COUNT;
  RETURN v_deleted;
END;
$$;

REVOKE ALL     ON FUNCTION login_throttle_purge_expired(integer) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION login_throttle_purge_expired(integer) TO barberia_worker;

COMMENT ON FUNCTION login_throttle_purge_expired(integer) IS
  'Mantenimiento del worker: purga en lote los contadores vencidos. Exclusivo de barberia_worker.';

-- ---------------------------------------------------------------------------
-- 4. A.5 · `auth_phone_challenge`
-- ---------------------------------------------------------------------------

CREATE TABLE auth_phone_challenge (
  id             uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id  uuid        NOT NULL,
  staff_user_id  uuid        NOT NULL,
  ip_hash        text        NOT NULL,
  code_hash      text        NOT NULL,
  attempt_count  smallint    NOT NULL DEFAULT 0,
  max_attempts   smallint    NOT NULL DEFAULT 5,
  expires_at     timestamptz NOT NULL,
  consumed_at    timestamptz,
  invalidated_at timestamptz,
  created_at     timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT auth_phone_challenge_id_pk PRIMARY KEY (id),

  CONSTRAINT auth_phone_challenge_barbershop_id_staff_user_id_fk
    FOREIGN KEY (barbershop_id, staff_user_id)
    REFERENCES staff_user (barbershop_id, id) ON DELETE CASCADE,

  CONSTRAINT auth_phone_challenge_ip_hash_ck CHECK (char_length(ip_hash) = 64),
  CONSTRAINT auth_phone_challenge_code_hash_ck CHECK (char_length(code_hash) = 64),
  CONSTRAINT auth_phone_challenge_attempt_count_ck CHECK (attempt_count >= 0 AND attempt_count <= max_attempts),
  CONSTRAINT auth_phone_challenge_max_attempts_ck CHECK (max_attempts BETWEEN 1 AND 10),
  CONSTRAINT auth_phone_challenge_expires_at_ck CHECK (expires_at > created_at),

  CONSTRAINT auth_phone_challenge_outcome_ck CHECK (
    NOT (consumed_at IS NOT NULL AND invalidated_at IS NOT NULL)
  ),
  CONSTRAINT auth_phone_challenge_consumed_at_ck CHECK (
    consumed_at IS NULL OR consumed_at >= created_at
  ),
  CONSTRAINT auth_phone_challenge_invalidated_at_ck CHECK (
    invalidated_at IS NULL OR invalidated_at >= created_at
  )
);

COMMENT ON TABLE auth_phone_challenge IS
  'Reto telefónico de HU-007 que desbloquea el login tras superar el umbral de intentos '
  '(DEC-061, DEC-062). Propietario funcional: plataforma. Retención: corta, se purga tras '
  'vencer. Clasificación: secreto. Solo se almacena el hash del código; la base de datos '
  'nunca contiene el valor enviado por WhatsApp. No es staff_recovery_code (HU-008): un '
  'reto de acceso y un código de recuperación de contraseña son artefactos distintos.';

CREATE INDEX idx_auth_phone_challenge_ip_created ON auth_phone_challenge (ip_hash, created_at);
CREATE INDEX idx_auth_phone_challenge_shop_user ON auth_phone_challenge (barbershop_id, staff_user_id);
CREATE INDEX idx_auth_phone_challenge_expires_at ON auth_phone_challenge (expires_at);

ALTER TABLE auth_phone_challenge ENABLE ROW LEVEL SECURITY;
ALTER TABLE auth_phone_challenge FORCE  ROW LEVEL SECURITY;

CREATE POLICY auth_phone_challenge_all_admin_policy ON auth_phone_challenge
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

-- DDL-AUT-01: sin política de SELECT/INSERT/UPDATE ni GRANT para
-- barberia_app. Las dos funciones siguientes son el único punto de acceso;
-- ambas corren ANTES de resolver contexto de tenant (igual que A.0/A.2).

CREATE OR REPLACE FUNCTION auth_phone_challenge_request(
  p_email                text,
  p_ip_hash               text,
  p_code_hash             text,
  p_expires_seconds       integer,
  p_rate_window_seconds   integer,
  p_rate_max_active       integer,
  p_rate_cooldown_seconds integer
)
RETURNS TABLE (accepted boolean, phone text)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_now       timestamptz := pg_catalog.now();
  v_user      record;
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

  SELECT u.id, u.barbershop_id, u.phone
  INTO v_user
  FROM public.staff_user u
  WHERE u.email = pg_catalog.lower(p_email)
    AND u.is_active
    AND u.phone_verified_at IS NOT NULL;

  IF NOT FOUND THEN
    RETURN QUERY SELECT false, NULL::text;
    RETURN;
  END IF;

  SELECT (t.escalated_until IS NOT NULL AND t.escalated_until > v_now)
  INTO v_escalated
  FROM public.login_throttle t
  WHERE t.ip_hash = p_ip_hash;

  IF NOT FOUND OR NOT v_escalated THEN
    RETURN QUERY SELECT false, NULL::text;
    RETURN;
  END IF;

  IF EXISTS (
    SELECT 1 FROM public.auth_phone_challenge c
    WHERE c.ip_hash = p_ip_hash
      AND c.created_at > v_now - pg_catalog.make_interval(secs => p_rate_cooldown_seconds)
  ) THEN
    RETURN QUERY SELECT false, NULL::text;
    RETURN;
  END IF;

  SELECT count(*) INTO v_recent
  FROM public.auth_phone_challenge c
  WHERE c.ip_hash = p_ip_hash
    AND c.created_at > v_now - pg_catalog.make_interval(secs => p_rate_window_seconds);

  IF v_recent >= p_rate_max_active THEN
    RETURN QUERY SELECT false, NULL::text;
    RETURN;
  END IF;

  INSERT INTO public.auth_phone_challenge
    (barbershop_id, staff_user_id, ip_hash, code_hash, expires_at)
  VALUES
    (v_user.barbershop_id, v_user.id, p_ip_hash, p_code_hash,
     v_now + pg_catalog.make_interval(secs => p_expires_seconds));

  RETURN QUERY SELECT true, v_user.phone;
END;
$$;

REVOKE ALL     ON FUNCTION auth_phone_challenge_request(text, text, text, integer, integer, integer, integer) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION auth_phone_challenge_request(text, text, text, integer, integer, integer, integer) TO barberia_app;

COMMENT ON FUNCTION auth_phone_challenge_request(text, text, text, integer, integer, integer, integer) IS
  'Único punto de escritura de auth_phone_challenge para solicitudes (DDL-AUT-01). accepted '
  'y phone solo se completan cuando la cuenta existe con teléfono verificado, la IP está '
  'realmente escalada y no se excede el límite propio de reenvío/solicitudes activas; en '
  'cualquier otro caso devuelve (false, NULL) sin crear fila, para que la capa HTTP '
  'responda siempre 202 sin distinguir el motivo (no enumeración).';

CREATE OR REPLACE FUNCTION auth_phone_challenge_verify(
  p_email    text,
  p_ip_hash  text,
  p_code_hash text
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
  v_row  public.auth_phone_challenge%ROWTYPE;
BEGIN
  IF p_email IS NULL OR p_ip_hash IS NULL OR p_code_hash IS NULL THEN
    RAISE EXCEPTION 'auth_phone_challenge_verify: ningún argumento admite NULL.';
  END IF;
  IF char_length(p_ip_hash) <> 64 OR char_length(p_code_hash) <> 64 THEN
    RAISE EXCEPTION 'auth_phone_challenge_verify: ip_hash/code_hash deben ser HMAC-SHA256 (64 hex).';
  END IF;

  SELECT u.id, u.barbershop_id
  INTO v_user
  FROM public.staff_user u
  WHERE u.email = pg_catalog.lower(p_email)
    AND u.is_active;

  IF NOT FOUND THEN
    RETURN false;
  END IF;

  SELECT * INTO v_row
  FROM public.auth_phone_challenge c
  WHERE c.barbershop_id = v_user.barbershop_id
    AND c.staff_user_id = v_user.id
    AND c.ip_hash = p_ip_hash
    AND c.consumed_at IS NULL
    AND c.invalidated_at IS NULL
    AND c.expires_at > v_now
  ORDER BY c.created_at DESC
  LIMIT 1
  FOR UPDATE;

  IF NOT FOUND THEN
    RETURN false;
  END IF;

  IF v_row.code_hash <> p_code_hash THEN
    UPDATE public.auth_phone_challenge
    SET attempt_count = attempt_count + 1,
        invalidated_at = CASE WHEN attempt_count + 1 >= max_attempts THEN v_now END
    WHERE id = v_row.id;
    RETURN false;
  END IF;

  UPDATE public.auth_phone_challenge SET consumed_at = v_now WHERE id = v_row.id;

  UPDATE public.login_throttle
  SET escalated_until = NULL,
      attempt_count   = 0,
      window_started_at = v_now
  WHERE ip_hash = p_ip_hash;

  RETURN true;
END;
$$;

REVOKE ALL     ON FUNCTION auth_phone_challenge_verify(text, text, text) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION auth_phone_challenge_verify(text, text, text) TO barberia_app;

COMMENT ON FUNCTION auth_phone_challenge_verify(text, text, text) IS
  'Único punto de escritura de verificación (DDL-AUT-01). Código incorrecto incrementa '
  'attempt_count y, al agotar max_attempts, invalida; código correcto consume el reto Y '
  'limpia login_throttle de la misma IP en la misma transacción (DEC-062). Cuenta '
  'inexistente, código incorrecto, vencido, agotado o de otra IP devuelven exactamente '
  'false: la capa HTTP no puede distinguirlos (no enumeración).';

CREATE OR REPLACE FUNCTION auth_phone_challenge_purge_expired(p_limit integer)
RETURNS integer
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_deleted integer;
BEGIN
  IF p_limit IS NULL OR p_limit < 1 OR p_limit > 1000 THEN
    RAISE EXCEPTION 'auth_phone_challenge_purge_expired: p_limit fuera de rango (1-1000).';
  END IF;

  WITH due AS (
    SELECT id FROM public.auth_phone_challenge
    WHERE expires_at <= pg_catalog.now()
    ORDER BY expires_at
    LIMIT p_limit
    FOR UPDATE SKIP LOCKED
  )
  DELETE FROM public.auth_phone_challenge
  WHERE id IN (SELECT id FROM due);

  GET DIAGNOSTICS v_deleted = ROW_COUNT;
  RETURN v_deleted;
END;
$$;

REVOKE ALL     ON FUNCTION auth_phone_challenge_purge_expired(integer) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION auth_phone_challenge_purge_expired(integer) TO barberia_worker;

COMMENT ON FUNCTION auth_phone_challenge_purge_expired(integer) IS
  'Mantenimiento del worker: purga en lote los retos vencidos. Exclusivo de barberia_worker.';
