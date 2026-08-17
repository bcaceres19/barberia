-- ===========================================================================
-- MODELO FÍSICO DE REFERENCIA — NO ES UNA MIGRACIÓN
-- ===========================================================================
--
-- Este archivo NO está en `database/migrations/`, NO entra en `atlas.sum` y NO
-- se aplica con Atlas. Existe por una razón concreta:
--
--   DEC-036 establece que una migración aplicada es inmutable. Escribir hoy
--   como migraciones el esquema completo de B1 a B6 produciría archivos que
--   habría que corregir con otras migraciones antes de aplicar la primera
--   línea de negocio. Este archivo conserva el diseño completo —tipos, claves,
--   índices, restricciones y el criterio de "ocupa agenda"— sin congelarlo.
--
-- Cómo se usa
--   Al abrir la HU dueña de cada sección: `atlas migrate new <descripcion>`,
--   copiar la sección al archivo generado, revisar, `atlas migrate hash`,
--   `atlas migrate validate` y escribir sus pruebas en `database/tests/`.
--   Cada sección indica su HU o bloque dueño y las reglas que sostiene.
--
-- Ya aplicado (no está aquí)
--   HU-001 · 20260807170000_create_tenant_foundation.sql
--   HU-004 · 20260807170100_create_idempotency_record.sql
--
-- Convenciones heredadas
--   estandar-base-datos.md §4 (nombres), §5 (tipos), §6 (integridad
--   tenant-aware), §7 (restricciones), §8 (tiempo y concurrencia), §9 (RLS).
--   Toda tabla de negocio repite el bloque final de RLS + privilegios; se
--   escribe completo en cada sección para que cada migración sea copiable
--   entera sin reconstruirlo de memoria.
--
-- ===========================================================================


-- ===========================================================================
-- SECCIÓN A · B0 — Autenticación (dudas resueltas)
-- ===========================================================================
--
-- Las tres dudas que bloqueaban esta sección (dudas-pendientes.md §2 bis) se
-- resolvieron el 2026-08-11:
--
--   DP-SEG-04 -> HU-005, HU-006 : sesión de 30 días, token opaco revocable
--                                 en staff_session, renovación por uso (DEC-050).
--   DP-SEG-05 -> HU-008, HU-011 : código de recuperación por WhatsApp oficial
--                                 y correo, proveedor de DEC-027 (DEC-051).
--   DP-SEG-06 -> HU-007         : ventana de 15 minutos, escalamiento a
--                                 verificación telefónica de 24 horas (DEC-052).
--
-- `staff_credential` y `login_throttle` además incorporan el endurecimiento
-- de `DDL-AUT-01` (docs/05-backend/revision-ddl-seguridad-2026-08-11.md):
-- ninguna de las dos concede ya SELECT/DML directo y amplio a `barberia_app`;
-- se exponen funciones `SECURITY DEFINER` estrechas y revisadas en su lugar.
--
-- ---------------------------------------------------------------------------
-- A.0 · El problema de resolver el tenant ANTES de tener contexto
-- ---------------------------------------------------------------------------
--
-- RLS forzada exige `app.barbershop_id` fijado, pero el inicio de sesión debe
-- encontrar al usuario a partir del correo, cuando todavía no se sabe a qué
-- barbería pertenece. Lo mismo ocurre al validar una cookie de sesión y al
-- abrir un enlace público (§E.3).
--
-- La salida es una función `SECURITY DEFINER` mínima, propiedad del rol
-- migrador, que devuelve exclusivamente el identificador de barbería. Es la
-- excepción prevista por estandar-base-datos.md §9.10: fija `search_path`,
-- revoca ejecución pública y requiere revisión de seguridad. No devuelve el
-- hash de la contraseña ni ningún dato personal: solo permite abrir la
-- transacción con el contexto correcto y volver a la ruta protegida por RLS.

CREATE OR REPLACE FUNCTION authn_resolve_login_tenant(p_email text)
RETURNS uuid
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = ''
AS $$
  SELECT barbershop_id
  FROM public.staff_user
  WHERE email = pg_catalog.lower(p_email)
    AND is_active
$$;

REVOKE ALL     ON FUNCTION authn_resolve_login_tenant(text) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION authn_resolve_login_tenant(text) TO barberia_app;

COMMENT ON FUNCTION authn_resolve_login_tenant(text) IS
  'SECURITY DEFINER acotada: resuelve la barbería de un correo de acceso para poder '
  'fijar app.barbershop_id. Devuelve NULL si no existe o está inactivo; quien la llama '
  'debe responder igual en ambos casos y consumir el mismo tiempo (CA-005-02). '
  'No expone credenciales ni datos personales.';

-- ---------------------------------------------------------------------------
-- A.1 · `staff_credential` — HU-005
-- ---------------------------------------------------------------------------
-- La credencial vive separada de `staff_user` para que ninguna consulta
-- ordinaria de perfil arrastre el material secreto.

CREATE TABLE staff_credential (
  staff_user_id       uuid        NOT NULL,
  barbershop_id       uuid        NOT NULL,
  password_hash       text        NOT NULL,
  password_algorithm  text        NOT NULL DEFAULT 'argon2id',
  password_updated_at timestamptz NOT NULL DEFAULT now(),
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT staff_credential_staff_user_id_pk PRIMARY KEY (staff_user_id),

  -- Composición tenant-aware: la FK simple por staff_user_id no demostraría
  -- que el usuario pertenece a esta barbería (estandar-base-datos.md §6).
  CONSTRAINT staff_credential_barbershop_id_staff_user_id_fk
    FOREIGN KEY (barbershop_id, staff_user_id)
    REFERENCES staff_user (barbershop_id, id) ON DELETE CASCADE,

  -- CASCADE autorizado: la credencial no tiene vida ni retención propias.
  CONSTRAINT staff_credential_password_hash_ck CHECK (
    char_length(password_hash) BETWEEN 32 AND 512
    -- El hash codificado de argon2id incluye sal y parámetros; no existe
    -- columna de sal separada a propósito (CA-005-03).
  ),
  CONSTRAINT staff_credential_password_algorithm_ck CHECK (
    password_algorithm IN ('argon2id', 'bcrypt')
  )
);

COMMENT ON TABLE staff_credential IS
  'Material de autenticación del área privada. Propietario funcional: barbería. '
  'Retención: mientras exista el usuario. Clasificación: secreto. Nunca se lee en '
  'consultas de listado ni se registra.';

CREATE TRIGGER staff_credential_set_updated_at
  BEFORE UPDATE ON staff_credential
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE staff_credential ENABLE ROW LEVEL SECURITY;
ALTER TABLE staff_credential FORCE  ROW LEVEL SECURITY;

CREATE POLICY staff_credential_all_admin_policy ON staff_credential
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY staff_credential_insert_tenant_policy ON staff_credential
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY staff_credential_update_tenant_policy ON staff_credential
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- DDL-AUT-01: sin política ni GRANT de SELECT para barberia_app. Un `SELECT *
-- FROM staff_credential` directo (por ejemplo, desde una inyección SQL)
-- dejaría de ser posible; la lectura pasa por auth_get_credential(), que
-- devuelve el material de un único usuario, no de toda la barbería.
GRANT INSERT, UPDATE ON TABLE staff_credential TO barberia_app;

-- Lectura estrecha y revisada (DDL-AUT-01): un único staff_user_id, siempre
-- dentro del tenant vigente. No hay forma de pedir "todas las credenciales".
CREATE OR REPLACE FUNCTION auth_get_credential(p_staff_user_id uuid)
RETURNS TABLE (password_hash text, password_algorithm text)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = ''
AS $$
  SELECT c.password_hash, c.password_algorithm
  FROM public.staff_credential c
  WHERE c.staff_user_id = p_staff_user_id
    AND c.barbershop_id = current_setting('app.barbershop_id')::uuid
$$;

REVOKE ALL     ON FUNCTION auth_get_credential(uuid) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION auth_get_credential(uuid) TO barberia_app;

COMMENT ON FUNCTION auth_get_credential(uuid) IS
  'Lectura estrecha de material de autenticación (DDL-AUT-01): un único staff_user_id, '
  'acotado al tenant vigente por app.barbershop_id. Sustituye el SELECT directo sobre '
  'staff_credential, que quedó sin GRANT ni política para barberia_app.';

-- ---------------------------------------------------------------------------
-- A.2 · `staff_session` — HU-005 / HU-006
-- ---------------------------------------------------------------------------
-- Token opaco revocable en base de datos (DEC-050): la única forma compatible
-- con CA-006-02 (revocación inmediata). Vigencia de 30 días desde el último
-- uso; la aplicación extiende `expires_at` en cada solicitud autenticada
-- (renovación deslizante, DEC-050). El plazo vive en configuración de la
-- aplicación, no en una columna: `expires_at` ya lo expresa por fila.

CREATE TABLE staff_session (
  id            uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id uuid        NOT NULL,
  staff_user_id uuid        NOT NULL,
  token_hash    text        NOT NULL,
  issued_at     timestamptz NOT NULL DEFAULT now(),
  expires_at    timestamptz NOT NULL,
  last_used_at  timestamptz NOT NULL DEFAULT now(),
  revoked_at    timestamptz,

  CONSTRAINT staff_session_id_pk PRIMARY KEY (id),

  CONSTRAINT staff_session_barbershop_id_staff_user_id_fk
    FOREIGN KEY (barbershop_id, staff_user_id)
    REFERENCES staff_user (barbershop_id, id) ON DELETE CASCADE,

  -- Unicidad GLOBAL, por el mismo motivo que el correo de acceso: la sesión se
  -- resuelve antes de conocer la barbería.
  CONSTRAINT staff_session_token_hash_uk UNIQUE (token_hash),

  CONSTRAINT staff_session_token_hash_ck  CHECK (char_length(token_hash) = 64),
  CONSTRAINT staff_session_expires_at_ck  CHECK (expires_at > issued_at),
  CONSTRAINT staff_session_last_used_at_ck CHECK (last_used_at >= issued_at),
  -- DDL-PER-01/DDL-TMP-01: la revocación no puede preceder a la emisión.
  -- Faltaba en esta tabla aunque ya existía en appointment_access_token.
  CONSTRAINT staff_session_revoked_at_ck CHECK (revoked_at IS NULL OR revoked_at >= issued_at)
);

COMMENT ON TABLE staff_session IS
  'Sesiones vigentes del área privada. Propietario funcional: barbería. '
  'Retención: hasta expires_at o revoked_at + ventana de auditoría. Clasificación: secreto. '
  'Se almacena solo el hash SHA-256 del token opaco; el valor original vive únicamente en '
  'la cookie del navegador (CA-006-05).';

-- Consulta de "mis sesiones activas", revocación masiva al cambiar contraseña
-- (CA-008-05) y auditoría de sesiones vencidas/revocadas del mismo usuario
-- (DDL-PER-01): sin predicado parcial a propósito. Un índice parcial
-- `WHERE revoked_at IS NULL` junto a este sería redundante -el planificador
-- ya sirve la consulta de sesiones activas con este mismo índice completo,
-- filtrando revoked_at en el mismo escaneo- y crearía dos índices que
-- cubren el mismo prefijo (prohibido explícitamente por el hallazgo).
CREATE INDEX idx_staff_session_shop_user
  ON staff_session (barbershop_id, staff_user_id);

-- Limpieza de sesiones vencidas.
CREATE INDEX idx_staff_session_expires_at ON staff_session (expires_at);

ALTER TABLE staff_session ENABLE ROW LEVEL SECURITY;
ALTER TABLE staff_session FORCE  ROW LEVEL SECURITY;

CREATE POLICY staff_session_all_admin_policy ON staff_session
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY staff_session_select_tenant_policy ON staff_session
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY staff_session_insert_tenant_policy ON staff_session
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY staff_session_update_tenant_policy ON staff_session
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY staff_session_delete_tenant_policy ON staff_session
  FOR DELETE TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE staff_session TO barberia_app;

-- Resolución de sesión previa al contexto, análoga a A.0.
CREATE OR REPLACE FUNCTION authn_resolve_session_tenant(p_token_hash text)
RETURNS uuid
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = ''
AS $$
  SELECT barbershop_id
  FROM public.staff_session
  WHERE token_hash = p_token_hash
    AND revoked_at IS NULL
    AND expires_at > pg_catalog.now()
$$;

REVOKE ALL     ON FUNCTION authn_resolve_session_tenant(text) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION authn_resolve_session_tenant(text) TO barberia_app;

-- ---------------------------------------------------------------------------
-- A.3 · `staff_recovery_code` — HU-008
-- ---------------------------------------------------------------------------

-- El teléfono verificado es requisito del mecanismo (DEC-026) y no existe en
-- HU-001; entra con esta migración. El envío del código usa WhatsApp oficial
-- y correo, mismo proveedor de DEC-027 (DEC-051); el destino de esta tabla es
-- el teléfono, y el correo se toma de staff_user.email.
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
  'Teléfono en formato E.164. Destino del código de recuperación; se muestra siempre '
  'enmascarado y nunca completo en respuestas ni registros (CA-008-06).';

-- DDL-INT-05: cambiar el teléfono invalida la verificación anterior; centralizado
-- aquí para que ningún flujo de actualización pueda olvidarlo.
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

CREATE TABLE staff_recovery_code (
  id             uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id  uuid        NOT NULL,
  staff_user_id  uuid        NOT NULL,
  code_hash      text        NOT NULL,
  attempt_count  smallint    NOT NULL DEFAULT 0,
  max_attempts   smallint    NOT NULL DEFAULT 5,
  expires_at     timestamptz NOT NULL,
  consumed_at    timestamptz,
  invalidated_at timestamptz,
  created_at     timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT staff_recovery_code_id_pk PRIMARY KEY (id),

  CONSTRAINT staff_recovery_code_barbershop_id_staff_user_id_fk
    FOREIGN KEY (barbershop_id, staff_user_id)
    REFERENCES staff_user (barbershop_id, id) ON DELETE CASCADE,

  CONSTRAINT staff_recovery_code_code_hash_ck     CHECK (char_length(code_hash) = 64),
  CONSTRAINT staff_recovery_code_attempt_count_ck CHECK (attempt_count >= 0 AND attempt_count <= max_attempts),
  CONSTRAINT staff_recovery_code_max_attempts_ck  CHECK (max_attempts BETWEEN 1 AND 10),
  CONSTRAINT staff_recovery_code_expires_at_ck    CHECK (expires_at > created_at),

  -- Un código consumido no puede además estar invalidado: son dos finales
  -- distintos del mismo objeto y confundirlos falsea la auditoría.
  CONSTRAINT staff_recovery_code_outcome_ck CHECK (
    NOT (consumed_at IS NOT NULL AND invalidated_at IS NOT NULL)
  ),
  -- DDL-PER-01/DDL-TMP-01: ningún desenlace puede preceder a la creación.
  CONSTRAINT staff_recovery_code_consumed_at_ck CHECK (
    consumed_at IS NULL OR consumed_at >= created_at
  ),
  CONSTRAINT staff_recovery_code_invalidated_at_ck CHECK (
    invalidated_at IS NULL OR invalidated_at >= created_at
  )
);

COMMENT ON TABLE staff_recovery_code IS
  'Códigos de recuperación de acceso. Propietario funcional: barbería. '
  'Retención: corta; se purgan tras vencer o consumirse. Clasificación: secreto. '
  'Solo se almacena el hash: la base de datos nunca contiene el valor enviado (CA-008-04).';

-- Un solo código vigente por usuario. Es la restricción que hace que un
-- reenvío deba invalidar el anterior en lugar de acumular códigos válidos
-- (CA-008-07).
CREATE UNIQUE INDEX idx_staff_recovery_code_active_per_user
  ON staff_recovery_code (barbershop_id, staff_user_id)
  WHERE consumed_at IS NULL AND invalidated_at IS NULL;

-- DDL-PER-01: el índice de arriba es parcial (solo el código vigente) y no
-- puede ensancharse sin romper la unicidad de negocio que garantiza. La
-- auditoría de un usuario ("todos sus códigos, consumidos e invalidados
-- incluidos") necesita uno aparte, sin predicado.
CREATE INDEX idx_staff_recovery_code_shop_user
  ON staff_recovery_code (barbershop_id, staff_user_id);

ALTER TABLE staff_recovery_code ENABLE ROW LEVEL SECURITY;
ALTER TABLE staff_recovery_code FORCE  ROW LEVEL SECURITY;

CREATE POLICY staff_recovery_code_all_admin_policy ON staff_recovery_code
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY staff_recovery_code_select_tenant_policy ON staff_recovery_code
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY staff_recovery_code_insert_tenant_policy ON staff_recovery_code
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY staff_recovery_code_update_tenant_policy ON staff_recovery_code
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY staff_recovery_code_delete_tenant_policy ON staff_recovery_code
  FOR DELETE TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE staff_recovery_code TO barberia_app;

-- ---------------------------------------------------------------------------
-- A.4 · `login_throttle` — HU-007
-- ---------------------------------------------------------------------------
-- Única tabla del sistema SIN `barbershop_id` y SIN RLS, y la excepción está
-- razonada: el conteo ocurre antes de saber quién es el solicitante, y
-- asociarlo a una barbería permitiría a un atacante repartir sus intentos
-- entre barberías para diluir el límite.
--
-- Ventana de 15 minutos y escalamiento de 24 horas (DEC-052), valores
-- iniciales configurables por la aplicación (CA-007-05 exige poder
-- cambiarlos sin recompilar); el esquema solo registra el resultado
-- (`window_started_at`, `escalated_until`), no la duración en sí.

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
  'HMAC-SHA256(secreto de despliegue, ip) (DDL-AUT-01): un hash simple con sal, aunque la '
  'sal no sea pública, sigue siendo reconstruible por fuerza bruta porque el espacio de IP '
  'es pequeño (~2^32 para IPv4); HMAC con clave secreta de despliegue no lo es. La clave '
  'nunca vive en el repositorio ni en esta base de datos. Sin barbershop_id ni RLS por '
  'diseño: ver comentario de la sección A.4.';

CREATE INDEX idx_login_throttle_expires_at ON login_throttle (expires_at);

-- DDL-AUT-01: sin GRANT directo a barberia_app. DML completo (en particular
-- UPDATE/DELETE) permitiría a una inyección SQL reiniciar el contador de
-- cualquier IP y anular el control de fuerza bruta. Toda la lógica del
-- protocolo vive en las dos funciones siguientes.

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
  -- Probado bajo concurrencia real: una versión con `SELECT ... FOR UPDATE`
  -- seguido de INSERT/UPDATE perdía incrementos, porque `FOR UPDATE` no
  -- bloquea nada cuando la fila todavía no existe — dos sesiones nuevas para
  -- la misma IP podían "empatar" en NOT FOUND y una sobrescribía a la otra
  -- con attempt_count=1 en vez de sumar. El UPSERT deja que PostgreSQL
  -- serialice por fila desde el primer intento.
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
    -- evalúan con normalidad; la solicitud p_threshold+1 es la que exige
    -- el reto. p_threshold se interpreta como "cantidad de solicitudes
    -- permitidas sin reto", nunca como "cantidad que ya exige reto".
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


-- ===========================================================================
-- SECCIÓN B · B1 — Identidad de la barbería y catálogo
-- ===========================================================================
-- Funciones F-CONF-01, F-CONF-02, F-SERV-01, F-SERV-02.
-- Reglas RN-SER-01 a RN-SER-04, RN-DIS-07, DEC-019.

-- ---------------------------------------------------------------------------
-- B.1 · Datos de contacto y enlace público de la barbería (F-CONF-01, F-PUB-01)
-- ---------------------------------------------------------------------------

ALTER TABLE barbershop
  ADD COLUMN contact_email text,
  ADD COLUMN contact_phone text,
  ADD COLUMN public_slug   text,
  ADD CONSTRAINT barbershop_contact_email_ck CHECK (
    contact_email IS NULL
    OR (contact_email = lower(contact_email) AND contact_email LIKE '_%@_%._%')
  ),
  ADD CONSTRAINT barbershop_contact_phone_ck CHECK (
    contact_phone IS NULL OR contact_phone ~ '^\+[1-9][0-9]{7,14}$'
  ),
  ADD CONSTRAINT barbershop_public_slug_ck CHECK (
    public_slug IS NULL OR public_slug ~ '^[a-z0-9]([a-z0-9-]{1,38}[a-z0-9])$'
  );

-- Índice único en lugar de restricción de tabla: permite comparar el
-- identificador público sin distinguir mayúsculas.
CREATE UNIQUE INDEX idx_barbershop_public_slug
  ON barbershop (lower(public_slug))
  WHERE public_slug IS NOT NULL;

COMMENT ON COLUMN barbershop.public_slug IS
  'Segmento estable del enlace público de reservas (F-PUB-01). Unicidad GLOBAL: el '
  'enlace se resuelve antes de existir contexto de tenant (ver función de la sección E.3).';

-- ---------------------------------------------------------------------------
-- B.2 · `barber` — F-CONF-02, DEC-019
-- ---------------------------------------------------------------------------
-- `barbershop` y `barber` son entidades distintas (base-datos.md §1). Una
-- barbería con un solo barbero tiene igualmente su fila en `barber`: es lo que
-- evita casos especiales en el código (criterio de salida de B1).

CREATE TABLE barber (
  id            uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id uuid        NOT NULL,
  full_name     text        NOT NULL,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT barber_id_pk               PRIMARY KEY (id),
  CONSTRAINT barber_barbershop_id_fk    FOREIGN KEY (barbershop_id)
                                        REFERENCES barbershop (id) ON DELETE RESTRICT,
  CONSTRAINT barber_barbershop_id_id_uk UNIQUE (barbershop_id, id),

  CONSTRAINT barber_full_name_ck CHECK (btrim(full_name) <> '' AND char_length(full_name) <= 120)
);

COMMENT ON TABLE barber IS
  'Persona que presta el servicio y unidad de exclusión de agenda (DEC-019). '
  'Propietario funcional: barbería. Retención: mientras exista la barbería; nunca se borra '
  'si tiene citas. Clasificación: datos personales (nombre). '
  'Alcance recortado a HU-021 (DEC-047, DDL-BIZ-02): sin borrado, desactivación, orden manual '
  'ni vínculo con staff_user; una historia futura los aprueba explícitamente si llegan a hacer falta.';

CREATE TRIGGER barber_set_updated_at
  BEFORE UPDATE ON barber
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE barber ENABLE ROW LEVEL SECURITY;
ALTER TABLE barber FORCE  ROW LEVEL SECURITY;

CREATE POLICY barber_all_admin_policy ON barber
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY barber_select_tenant_policy ON barber
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY barber_insert_tenant_policy ON barber
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY barber_update_tenant_policy ON barber
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Sin política ni privilegio de DELETE: HU-021 lo excluye explícitamente (DEC-047).
GRANT SELECT, INSERT, UPDATE ON TABLE barber TO barberia_app;

-- ---------------------------------------------------------------------------
-- B.3 · `service` — F-SERV-01, F-SERV-02, RN-SER-01 a RN-SER-04
-- ---------------------------------------------------------------------------

CREATE TABLE service (
  id               uuid           NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id    uuid           NOT NULL,
  name             text           NOT NULL,
  description      text,
  duration_minutes integer        NOT NULL,
  price_amount     numeric(12, 2) NOT NULL,
  price_currency   char(3)        NOT NULL DEFAULT 'COP',
  is_active        boolean        NOT NULL DEFAULT true,
  deactivated_at   timestamptz,
  created_at       timestamptz    NOT NULL DEFAULT now(),
  updated_at       timestamptz    NOT NULL DEFAULT now(),

  CONSTRAINT service_id_pk               PRIMARY KEY (id),
  CONSTRAINT service_barbershop_id_fk    FOREIGN KEY (barbershop_id)
                                         REFERENCES barbershop (id) ON DELETE RESTRICT,
  CONSTRAINT service_barbershop_id_id_uk UNIQUE (barbershop_id, id),

  CONSTRAINT service_name_ck        CHECK (btrim(name) <> '' AND char_length(name) <= 120),
  CONSTRAINT service_description_ck CHECK (description IS NULL OR char_length(description) <= 500),

  -- RN-SER-01: minutos enteros positivos. El tope evita una duración absurda
  -- por error de digitación sin cerrar la lista de valores (RN-SER-02).
  CONSTRAINT service_duration_minutes_ck CHECK (duration_minutes BETWEEN 1 AND 1440),

  -- `numeric`, nunca coma flotante (estandar-base-datos.md §5).
  CONSTRAINT service_price_amount_ck   CHECK (price_amount >= 0),
  CONSTRAINT service_price_currency_ck CHECK (price_currency ~ '^[A-Z]{3}$'),

  -- RN-SER-03: un servicio no se elimina; se desactiva y conserva su historia.
  CONSTRAINT service_deactivated_at_ck CHECK (
    (is_active AND deactivated_at IS NULL) OR (NOT is_active AND deactivated_at IS NOT NULL)
  )
);

COMMENT ON TABLE service IS
  'Catálogo de servicios de la barbería. Propietario funcional: barbería. '
  'Retención: permanente; nunca se borra físicamente (RN-SER-03). Clasificación: negocio. '
  'Los cambios de este catálogo NO se propagan a citas existentes: la cita guarda su propio '
  'snapshot (DEC-004, ver sección D.1).';

-- Nombres únicos entre los servicios ofrecidos; un servicio desactivado libera
-- su nombre para poder reemplazarlo.
CREATE UNIQUE INDEX idx_service_active_name
  ON service (barbershop_id, lower(name))
  WHERE is_active;

CREATE TRIGGER service_set_updated_at
  BEFORE UPDATE ON service
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE service ENABLE ROW LEVEL SECURITY;
ALTER TABLE service FORCE  ROW LEVEL SECURITY;

-- DDL-RLS-01/H.7: apuntaba a barberia_migrator, inconsistente con DEC-040
-- (política administrativa de tabla nueva -> barberia_owner; barberia_migrator
-- la satisface por membresía heredada, sin SET ROLE).
CREATE POLICY service_all_admin_policy ON service
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY service_select_tenant_policy ON service
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY service_insert_tenant_policy ON service
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY service_update_tenant_policy ON service
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Sin política de DELETE: RN-SER-03 prohíbe el borrado físico.
GRANT SELECT, INSERT, UPDATE ON TABLE service TO barberia_app;

-- ---------------------------------------------------------------------------
-- B.4 · `barber_service` — F-CONF-02 ("servicios por barbero")
-- ---------------------------------------------------------------------------
-- Tabla de asociación pura: solo contiene datos de la asociación, nunca
-- atributos que pertenecen a uno de sus extremos (2FN).

CREATE TABLE barber_service (
  barbershop_id uuid        NOT NULL,
  barber_id     uuid        NOT NULL,
  service_id    uuid        NOT NULL,
  created_at    timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT barber_service_pk PRIMARY KEY (barbershop_id, barber_id, service_id),

  CONSTRAINT barber_service_barbershop_id_barber_id_fk
    FOREIGN KEY (barbershop_id, barber_id)
    REFERENCES barber (barbershop_id, id) ON DELETE CASCADE,

  CONSTRAINT barber_service_barbershop_id_service_id_fk
    FOREIGN KEY (barbershop_id, service_id)
    REFERENCES service (barbershop_id, id) ON DELETE CASCADE
);

COMMENT ON TABLE barber_service IS
  'Qué servicios presta cada barbero. Propietario funcional: barbería. Retención: mientras '
  'exista la asociación. Clasificación: negocio. CASCADE autorizado: la fila no tiene vida '
  'propia y su desaparición no borra citas, que guardan su snapshot.';

-- Consulta inversa: qué barberos prestan un servicio (selección pública).
CREATE INDEX idx_barber_service_shop_service
  ON barber_service (barbershop_id, service_id);

ALTER TABLE barber_service ENABLE ROW LEVEL SECURITY;
ALTER TABLE barber_service FORCE  ROW LEVEL SECURITY;

-- DDL-RLS-01/H.7: ídem service_all_admin_policy.
CREATE POLICY barber_service_all_admin_policy ON barber_service
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY barber_service_select_tenant_policy ON barber_service
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY barber_service_insert_tenant_policy ON barber_service
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY barber_service_delete_tenant_policy ON barber_service
  FOR DELETE TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT, DELETE ON TABLE barber_service TO barberia_app;


-- ===========================================================================
-- SECCIÓN C · B2 — Horario laboral y bloqueos
-- ===========================================================================
-- Funciones F-HOR-01, F-HOR-02. Reglas RN-BLQ-01 a RN-BLQ-04, RN-DIS-05,
-- RN-DIS-07, DEC-020.

-- ---------------------------------------------------------------------------
-- C.1 · Calendario de festivos por barbero (RN-BLQ-02)
-- ---------------------------------------------------------------------------

ALTER TABLE barber
  ADD COLUMN holiday_calendar_enabled boolean NOT NULL DEFAULT false;

COMMENT ON COLUMN barber.holiday_calendar_enabled IS
  'Activa el calendario colombiano de festivos para ESTE barbero (RN-BLQ-02). '
  'La decisión de un barbero no altera el calendario de otro de la misma barbería.';

-- ---------------------------------------------------------------------------
-- C.2 · `working_hour` — F-HOR-01
-- ---------------------------------------------------------------------------
-- Representa tiempo civil local recurrente por día de semana. Se interpreta
-- junto con `barbershop.timezone`; no se almacena un instante calculado
-- (estandar-base-datos.md §5.5). Nombre en singular (DDL-NAM-01): antes de
-- existir migraciones aplicadas, `working_hours`/`working_hours_override` se
-- renombran a `working_hour`/`working_hour_override`.
--
-- Cruce de medianoche (DEC-020): en lugar de `ends_time`, que sería ambiguo
-- cuando el fin cae en el día siguiente, se almacena inicio + duración. Un
-- tramo 22:00 + 300 min termina a las 03:00 del día siguiente sin ambigüedad,
-- y `duration_minutes` cumple la regla de tipos para duraciones.

CREATE TABLE working_hour (
  id               uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id    uuid        NOT NULL,
  barber_id        uuid        NOT NULL,
  iso_weekday      smallint    NOT NULL,
  starts_time      time        NOT NULL,
  duration_minutes integer     NOT NULL,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT working_hour_id_pk PRIMARY KEY (id),

  CONSTRAINT working_hour_barbershop_id_barber_id_fk
    FOREIGN KEY (barbershop_id, barber_id)
    REFERENCES barber (barbershop_id, id) ON DELETE CASCADE,

  -- ISO 8601: 1 = lunes … 7 = domingo. No se usa 0-6 para no heredar la
  -- ambigüedad de qué día es el cero.
  CONSTRAINT working_hour_iso_weekday_ck      CHECK (iso_weekday BETWEEN 1 AND 7),
  CONSTRAINT working_hour_duration_minutes_ck CHECK (duration_minutes BETWEEN 1 AND 1440),

  -- Un mismo barbero no repite el mismo inicio en el mismo día de semana.
  CONSTRAINT working_hour_shop_barber_weekday_start_uk
    UNIQUE (barbershop_id, barber_id, iso_weekday, starts_time)
);

COMMENT ON TABLE working_hour IS
  'Jornada laboral recurrente por barbero y día de semana (F-HOR-01). Propietario funcional: '
  'barbería. Retención: mientras exista el barbero. Clasificación: negocio. '
  'Varios tramos por día representan jornada partida.';

COMMENT ON COLUMN working_hour.duration_minutes IS
  'Duración del tramo desde starts_time. Un valor que empuje el fin más allá de medianoche '
  'representa una jornada nocturna, permitida por DEC-020.';

-- El solapamiento entre tramos del MISMO barbero y día es un error de
-- configuración, no una invariante de integridad de agenda: no se expresa como
-- restricción de exclusión porque el cálculo con envolvente semanal y jornadas
-- nocturnas la haría frágil. Lo valida el servicio de horario y lo cubre su
-- prueba (estandar-base-datos.md §7, último párrafo). La invariante dura —que
-- dos citas no se crucen— sí vive en la base de datos (sección D.1).

CREATE INDEX idx_working_hour_shop_barber_weekday
  ON working_hour (barbershop_id, barber_id, iso_weekday);

CREATE TRIGGER working_hour_set_updated_at
  BEFORE UPDATE ON working_hour
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE working_hour ENABLE ROW LEVEL SECURITY;
ALTER TABLE working_hour FORCE  ROW LEVEL SECURITY;

CREATE POLICY working_hour_all_admin_policy ON working_hour
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY working_hour_select_tenant_policy ON working_hour
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY working_hour_insert_tenant_policy ON working_hour
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY working_hour_update_tenant_policy ON working_hour
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY working_hour_delete_tenant_policy ON working_hour
  FOR DELETE TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE working_hour TO barberia_app;

-- ---------------------------------------------------------------------------
-- C.3 · `working_hour_override` — horario especial por fecha (RN-BLQ-02)
-- ---------------------------------------------------------------------------
-- Cubre el festivo que el barbero decide trabajar con horario reducido y el
-- día suelto con jornada distinta. Prevalece sobre `working_hour` para esa
-- fecha y sobre el bloqueo automático de festivo.
--
-- DDL-INT-01: la versión anterior tenía starts_time/duration_minutes en la
-- misma fila que is_closed y solo dos índices parciales; nada impedía que un
-- mismo día tuviera a la vez una fila cerrada y filas abiertas, ni que dos
-- tramos abiertos se solaparan. Se separa la cabecera (una fila por barbero y
-- fecha; is_closed decide la forma) de sus tramos (tabla hija, solo válidos
-- si el día está abierto). Una restricción de exclusión impide el solape entre
-- tramos de la misma cabecera, y un disparador impide crear un tramo bajo una
-- cabecera cerrada: son las dos invariantes que la forma anterior no cubría.

CREATE TABLE working_hour_override (
  id             uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id  uuid        NOT NULL,
  barber_id      uuid        NOT NULL,
  effective_date date        NOT NULL,
  is_closed      boolean     NOT NULL,
  reason         text,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT working_hour_override_id_pk               PRIMARY KEY (id),
  CONSTRAINT working_hour_override_barbershop_id_id_uk UNIQUE (barbershop_id, id),

  CONSTRAINT working_hour_override_barbershop_id_barber_id_fk
    FOREIGN KEY (barbershop_id, barber_id)
    REFERENCES barber (barbershop_id, id) ON DELETE CASCADE,

  -- Una sola cabecera por barbero y fecha: cierra la mitad de DDL-INT-01 que
  -- exigía elegir entre cerrado y abierto sin admitir ambos ni ninguno dos veces.
  CONSTRAINT working_hour_override_shop_barber_date_uk
    UNIQUE (barbershop_id, barber_id, effective_date),

  CONSTRAINT working_hour_override_reason_ck CHECK (
    reason IS NULL OR char_length(reason) <= 200
  )
);

COMMENT ON TABLE working_hour_override IS
  'Cabecera de excepción de jornada para una fecha concreta (RN-BLQ-02): is_closed decide si '
  'el día está cerrado o tiene tramos propios en working_hour_override_segment. '
  'Propietario funcional: barbería. Retención: mientras exista el barbero. Clasificación: negocio.';

CREATE INDEX idx_working_hour_override_shop_barber_date
  ON working_hour_override (barbershop_id, barber_id, effective_date);

CREATE TRIGGER working_hour_override_set_updated_at
  BEFORE UPDATE ON working_hour_override
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE working_hour_override ENABLE ROW LEVEL SECURITY;
ALTER TABLE working_hour_override FORCE  ROW LEVEL SECURITY;

CREATE POLICY working_hour_override_all_admin_policy ON working_hour_override
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY working_hour_override_select_tenant_policy ON working_hour_override
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY working_hour_override_insert_tenant_policy ON working_hour_override
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY working_hour_override_update_tenant_policy ON working_hour_override
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY working_hour_override_delete_tenant_policy ON working_hour_override
  FOR DELETE TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE working_hour_override TO barberia_app;

-- ---------------------------------------------------------------------------
-- C.3b · `working_hour_override_segment` — tramos de una cabecera abierta
-- ---------------------------------------------------------------------------
-- Solo tiene sentido bajo una cabecera con is_closed = false; el disparador
-- de más abajo lo exige porque ninguna restricción declarativa puede mirar
-- otra tabla.

CREATE TABLE working_hour_override_segment (
  id               uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id    uuid        NOT NULL,
  override_id      uuid        NOT NULL,
  starts_time      time        NOT NULL,
  duration_minutes integer     NOT NULL,
  created_at       timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT working_hour_override_segment_id_pk PRIMARY KEY (id),

  CONSTRAINT working_hour_override_segment_barbershop_id_override_id_fk
    FOREIGN KEY (barbershop_id, override_id)
    REFERENCES working_hour_override (barbershop_id, id) ON DELETE CASCADE,

  CONSTRAINT working_hour_override_segment_duration_minutes_ck CHECK (
    duration_minutes BETWEEN 1 AND 1440
  ),
  CONSTRAINT working_hour_override_segment_override_start_uk
    UNIQUE (override_id, starts_time)
);

COMMENT ON TABLE working_hour_override_segment IS
  'Tramo horario de una cabecera de excepción abierta (RN-BLQ-02). Clasificación: negocio.';

-- Ningún tramo se solapa con otro de la MISMA cabecera. La fecha vive en la
-- cabecera (fija por fila), así que un ancla arbitraria común (2000-01-01)
-- basta para comparar los tramos entre sí; un tramo nocturno que cruza
-- medianoche sigue comparándose bien porque todos usan la misma ancla. Mismo
-- intervalo semiabierto `[)` que appointment (sección D.1).
ALTER TABLE working_hour_override_segment
  ADD CONSTRAINT working_hour_override_segment_no_overlap_excl
  EXCLUDE USING gist (
    override_id WITH =,
    tsrange(
      date '2000-01-01' + starts_time,
      date '2000-01-01' + starts_time + make_interval(mins => duration_minutes),
      '[)'
    ) WITH &&
  );

CREATE OR REPLACE FUNCTION working_hour_override_segment_check_open()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $$
DECLARE
  v_is_closed boolean;
BEGIN
  SELECT is_closed INTO v_is_closed
  FROM public.working_hour_override
  WHERE id = NEW.override_id AND barbershop_id = NEW.barbershop_id;

  IF v_is_closed IS NULL THEN
    RAISE EXCEPTION 'working_hour_override % no existe en la barbería %',
      NEW.override_id, NEW.barbershop_id;
  END IF;

  IF v_is_closed THEN
    RAISE EXCEPTION 'working_hour_override % está cerrada (is_closed=true): no admite tramos',
      NEW.override_id;
  END IF;

  RETURN NEW;
END;
$$;

REVOKE ALL     ON FUNCTION working_hour_override_segment_check_open() FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION working_hour_override_segment_check_open() TO barberia_app;

COMMENT ON FUNCTION working_hour_override_segment_check_open() IS
  'Impide insertar/mover un tramo bajo una cabecera cerrada (DDL-INT-01). La FK ya garantiza '
  'que la cabecera existe en la misma barbería; esto añade la condición is_closed = false.';

CREATE TRIGGER working_hour_override_segment_check_open_trg
  BEFORE INSERT OR UPDATE ON working_hour_override_segment
  FOR EACH ROW EXECUTE FUNCTION working_hour_override_segment_check_open();

ALTER TABLE working_hour_override_segment ENABLE ROW LEVEL SECURITY;
ALTER TABLE working_hour_override_segment FORCE  ROW LEVEL SECURITY;

CREATE POLICY working_hour_override_segment_all_admin_policy ON working_hour_override_segment
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY working_hour_override_segment_select_tenant_policy ON working_hour_override_segment
  FOR SELECT TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY working_hour_override_segment_insert_tenant_policy ON working_hour_override_segment
  FOR INSERT TO barberia_app WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY working_hour_override_segment_update_tenant_policy ON working_hour_override_segment
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY working_hour_override_segment_delete_tenant_policy ON working_hour_override_segment
  FOR DELETE TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE working_hour_override_segment TO barberia_app;

-- ---------------------------------------------------------------------------
-- C.4 · `time_block_series` — bloqueos recurrentes y listas de fechas
-- ---------------------------------------------------------------------------
-- DEC-020 exige tres formas: instancia puntual, recurrencia y lista explícita
-- de fechas, con excepciones individuales. AGENTS.md prohíbe representar
-- horarios en `jsonb`, así que la recurrencia se modela como entidad con dos
-- tablas hijas.
--
-- La disponibilidad expande las series en la consulta del rango pedido; NO se
-- materializan instancias en `time_block`. Materializarlas obligaría a
-- mantener un horizonte y a repararlo tras cada cambio de configuración.

CREATE TABLE time_block_series (
  id               uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id    uuid        NOT NULL,
  barber_id        uuid        NOT NULL,
  block_type       text        NOT NULL,
  recurrence_kind  text        NOT NULL,
  iso_weekday      smallint,
  starts_time      time        NOT NULL,
  duration_minutes integer     NOT NULL,
  effective_from   date        NOT NULL,
  effective_until  date,
  reason           text,
  deleted_at       timestamptz,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT time_block_series_id_pk               PRIMARY KEY (id),
  CONSTRAINT time_block_series_barbershop_id_id_uk UNIQUE (barbershop_id, id),

  CONSTRAINT time_block_series_barbershop_id_barber_id_fk
    FOREIGN KEY (barbershop_id, barber_id)
    REFERENCES barber (barbershop_id, id) ON DELETE CASCADE,

  -- Los siete tipos del criterio de salida de B2 (RN-BLQ-01).
  CONSTRAINT time_block_series_block_type_ck CHECK (
    block_type IN ('break', 'lunch', 'unavailable', 'day_off', 'holiday', 'vacation', 'emergency')
  ),
  CONSTRAINT time_block_series_recurrence_kind_ck CHECK (
    recurrence_kind IN ('weekly', 'date_list')
  ),

  -- Una recurrencia semanal necesita día de semana; una lista de fechas no.
  CONSTRAINT time_block_series_weekday_shape_ck CHECK (
    (recurrence_kind = 'weekly'    AND iso_weekday IS NOT NULL AND iso_weekday BETWEEN 1 AND 7)
    OR
    (recurrence_kind = 'date_list' AND iso_weekday IS NULL)
  ),
  CONSTRAINT time_block_series_duration_minutes_ck CHECK (duration_minutes BETWEEN 1 AND 1440),
  CONSTRAINT time_block_series_effective_range_ck  CHECK (
    effective_until IS NULL OR effective_until >= effective_from
  ),
  CONSTRAINT time_block_series_reason_ck CHECK (reason IS NULL OR char_length(reason) <= 200),

  -- DDL-TMP-01: la eliminación lógica no puede preceder a la creación.
  CONSTRAINT time_block_series_deleted_at_ck CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

COMMENT ON TABLE time_block_series IS
  'Definición de un bloqueo recurrente o por lista de fechas (RN-BLQ-01, DEC-020). '
  'Propietario funcional: barbería. Retención: permanente; eliminación lógica (RN-BLQ-04). '
  'Clasificación: negocio. Nunca se materializa en time_block: la disponibilidad la expande.';

-- DDL-PER-01: sin predicado parcial. Un `WHERE deleted_at IS NULL` aparte
-- sería un segundo índice con el mismo prefijo (duplicado prohibido por el
-- hallazgo); este mismo índice completo sirve tanto la agenda vigente como
-- la auditoría de series borradas lógicamente por barbero.
CREATE INDEX idx_time_block_series_shop_barber
  ON time_block_series (barbershop_id, barber_id, effective_from);

CREATE TRIGGER time_block_series_set_updated_at
  BEFORE UPDATE ON time_block_series
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE time_block_series ENABLE ROW LEVEL SECURITY;
ALTER TABLE time_block_series FORCE  ROW LEVEL SECURITY;

CREATE POLICY time_block_series_all_admin_policy ON time_block_series
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY time_block_series_select_tenant_policy ON time_block_series
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY time_block_series_insert_tenant_policy ON time_block_series
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY time_block_series_update_tenant_policy ON time_block_series
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Sin DELETE: RN-BLQ-04 exige eliminación lógica.
GRANT SELECT, INSERT, UPDATE ON TABLE time_block_series TO barberia_app;

-- Fechas explícitas de una serie `date_list`.
CREATE TABLE time_block_series_date (
  barbershop_id uuid NOT NULL,
  series_id     uuid NOT NULL,
  block_date    date NOT NULL,

  CONSTRAINT time_block_series_date_pk PRIMARY KEY (barbershop_id, series_id, block_date),
  CONSTRAINT time_block_series_date_barbershop_id_series_id_fk
    FOREIGN KEY (barbershop_id, series_id)
    REFERENCES time_block_series (barbershop_id, id) ON DELETE CASCADE
);

COMMENT ON TABLE time_block_series_date IS
  'Fechas explícitas de una serie de tipo date_list (por ejemplo, vacaciones del 15 al 30). '
  'Clasificación: negocio.';

-- DDL-INT-03: ninguna restricción declarativa puede comprobar a la vez que el
-- padre es de tipo date_list y que la fecha cae dentro de su rango efectivo
-- (effective_from/effective_until). Disparador pequeño y documentado, como
-- pide el hallazgo, en vez de dividir el modelo por tipo.
CREATE OR REPLACE FUNCTION time_block_series_date_check_parent()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $$
DECLARE
  v_series public.time_block_series%ROWTYPE;
BEGIN
  SELECT * INTO v_series
  FROM public.time_block_series
  WHERE id = NEW.series_id AND barbershop_id = NEW.barbershop_id;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'time_block_series % no existe en la barbería %', NEW.series_id, NEW.barbershop_id;
  END IF;

  IF v_series.recurrence_kind <> 'date_list' THEN
    RAISE EXCEPTION 'time_block_series % no es date_list (es %): no admite fechas explícitas',
      NEW.series_id, v_series.recurrence_kind;
  END IF;

  IF NEW.block_date < v_series.effective_from
     OR (v_series.effective_until IS NOT NULL AND NEW.block_date > v_series.effective_until) THEN
    RAISE EXCEPTION 'block_date % fuera del rango efectivo [%, %] de time_block_series %',
      NEW.block_date, v_series.effective_from, v_series.effective_until, NEW.series_id;
  END IF;

  RETURN NEW;
END;
$$;

REVOKE ALL     ON FUNCTION time_block_series_date_check_parent() FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION time_block_series_date_check_parent() TO barberia_app;

COMMENT ON FUNCTION time_block_series_date_check_parent() IS
  'Valida que la fecha explícita pertenezca a una serie date_list vigente en ese rango '
  '(DDL-INT-03). No puede expresarse como CHECK: necesita leer otra tabla.';

CREATE TRIGGER time_block_series_date_check_parent_trg
  BEFORE INSERT OR UPDATE ON time_block_series_date
  FOR EACH ROW EXECUTE FUNCTION time_block_series_date_check_parent();

-- Excepciones individuales: "esta instancia no" sin desarmar la serie.
CREATE TABLE time_block_series_exception (
  barbershop_id uuid        NOT NULL,
  series_id     uuid        NOT NULL,
  excluded_date date        NOT NULL,
  reason        text,
  created_at    timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT time_block_series_exception_pk PRIMARY KEY (barbershop_id, series_id, excluded_date),
  CONSTRAINT time_block_series_exception_barbershop_id_series_id_fk
    FOREIGN KEY (barbershop_id, series_id)
    REFERENCES time_block_series (barbershop_id, id) ON DELETE CASCADE,
  CONSTRAINT time_block_series_exception_reason_ck CHECK (
    reason IS NULL OR char_length(reason) <= 200
  )
);

COMMENT ON TABLE time_block_series_exception IS
  'Instancias suprimidas de una serie recurrente ("esta instancia", RN-BLQ-01). '
  'Clasificación: negocio.';

ALTER TABLE time_block_series_date      ENABLE ROW LEVEL SECURITY;
ALTER TABLE time_block_series_date      FORCE  ROW LEVEL SECURITY;
ALTER TABLE time_block_series_exception ENABLE ROW LEVEL SECURITY;
ALTER TABLE time_block_series_exception FORCE  ROW LEVEL SECURITY;

CREATE POLICY time_block_series_date_all_admin_policy ON time_block_series_date
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY time_block_series_date_select_tenant_policy ON time_block_series_date
  FOR SELECT TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY time_block_series_date_insert_tenant_policy ON time_block_series_date
  FOR INSERT TO barberia_app WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY time_block_series_date_delete_tenant_policy ON time_block_series_date
  FOR DELETE TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY time_block_series_exception_all_admin_policy ON time_block_series_exception
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY time_block_series_exception_select_tenant_policy ON time_block_series_exception
  FOR SELECT TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY time_block_series_exception_insert_tenant_policy ON time_block_series_exception
  FOR INSERT TO barberia_app WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY time_block_series_exception_delete_tenant_policy ON time_block_series_exception
  FOR DELETE TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT, DELETE ON TABLE time_block_series_date      TO barberia_app;
GRANT SELECT, INSERT, DELETE ON TABLE time_block_series_exception TO barberia_app;

-- ---------------------------------------------------------------------------
-- C.5 · `time_block` — bloqueo puntual, incluida la emergencia (RN-BLQ-03)
-- ---------------------------------------------------------------------------

CREATE TABLE time_block (
  id            uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id uuid        NOT NULL,
  barber_id     uuid        NOT NULL,
  block_type    text        NOT NULL,
  source        text        NOT NULL DEFAULT 'manual',
  starts_at     timestamptz NOT NULL,
  ends_at       timestamptz NOT NULL,
  reason        text,
  deleted_at    timestamptz,
  deleted_by    uuid,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT time_block_id_pk PRIMARY KEY (id),

  CONSTRAINT time_block_barbershop_id_barber_id_fk
    FOREIGN KEY (barbershop_id, barber_id)
    REFERENCES barber (barbershop_id, id) ON DELETE CASCADE,

  -- RESTRICT, no SET NULL (DDL-INT-02): barbershop_id es NOT NULL y una FK
  -- compuesta con SET NULL anularía ambas columnas a la vez, violando el
  -- NOT NULL y haciendo fallar el DELETE del staff_user. staff_user nunca se
  -- borra físicamente en este diseño (se desactiva, no se elimina).
  CONSTRAINT time_block_barbershop_id_deleted_by_fk
    FOREIGN KEY (barbershop_id, deleted_by)
    REFERENCES staff_user (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT time_block_block_type_ck CHECK (
    block_type IN ('break', 'lunch', 'unavailable', 'day_off', 'holiday', 'vacation', 'emergency')
  ),
  CONSTRAINT time_block_source_ck CHECK (source IN ('manual', 'holiday_calendar')),

  -- Intervalo semiabierto [starts_at, ends_at) con fin estrictamente posterior
  -- (RN-DIS-05).
  CONSTRAINT time_block_interval_ck   CHECK (ends_at > starts_at),
  CONSTRAINT time_block_reason_ck     CHECK (reason IS NULL OR char_length(reason) <= 200),
  CONSTRAINT time_block_deleted_by_ck CHECK (
    (deleted_at IS NULL AND deleted_by IS NULL) OR deleted_at IS NOT NULL
  ),
  -- DDL-TMP-01: la eliminación lógica no puede preceder a la creación.
  CONSTRAINT time_block_deleted_at_ck CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

COMMENT ON TABLE time_block IS
  'Bloqueo puntual de agenda, incluida la emergencia (RN-BLQ-01, RN-BLQ-03). '
  'Propietario funcional: barbería. Retención: permanente; eliminación lógica (RN-BLQ-04, DEC-009). '
  'Clasificación: negocio.';

COMMENT ON COLUMN time_block.starts_at IS
  'DELIBERADAMENTE SIN restricción de exclusión contra appointment: RN-BLQ-03 y DEC-008 exigen '
  'que un bloqueo urgente SIEMPRE se pueda crear aunque haya citas encima. Impedirlo sería lo '
  'contrario de lo que el barbero necesita en una emergencia. El flujo asistido de citas '
  'afectadas vive en la aplicación.';

-- Índice de base-datos.md §12. DDL-PER-01: sin predicado parcial, por el
-- mismo motivo que idx_time_block_series_shop_barber -sirve la agenda
-- vigente y la auditoría de bloqueos borrados lógicamente por barbero sin
-- duplicar el índice.
CREATE INDEX idx_time_block_shop_barber_starts_at
  ON time_block (barbershop_id, barber_id, starts_at);

-- DDL-PER-01: lado referenciante de deleted_by sin índice ("¿qué bloqueos
-- eliminó este miembro del staff?"). Parcial: la mayoría de filas nunca se
-- eliminan (deleted_by IS NULL).
CREATE INDEX idx_time_block_shop_deleted_by
  ON time_block (barbershop_id, deleted_by)
  WHERE deleted_by IS NOT NULL;

CREATE TRIGGER time_block_set_updated_at
  BEFORE UPDATE ON time_block
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE time_block ENABLE ROW LEVEL SECURITY;
ALTER TABLE time_block FORCE  ROW LEVEL SECURITY;

CREATE POLICY time_block_all_admin_policy ON time_block
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY time_block_select_tenant_policy ON time_block
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY time_block_insert_tenant_policy ON time_block
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY time_block_update_tenant_policy ON time_block
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Sin DELETE: RN-BLQ-04.
GRANT SELECT, INSERT, UPDATE ON TABLE time_block TO barberia_app;


-- ===========================================================================
-- SECCIÓN D · B3 — Agenda, estados e integridad
-- ===========================================================================
-- El corazón del producto. Funciones F-CITA-01 a F-CITA-06, F-EST-01 a
-- F-EST-04, F-DISP-02. Reglas RN-CON-01, RN-CON-03, RN-CIT-*, RN-HIS-*,
-- RN-RES-*, y la máquina completa de estados-citas.md.
--
-- DDL-INT-00: `customer` vive aquí, en D.0, y no en la sección E de reserva
-- pública donde pertenece conceptualmente (RN-DAT-01, RN-DAT-03, DEC-022),
-- porque toda cita —también la manual— la referencia. El archivo debe
-- ejecutarse en orden desde una base vacía; si `customer` se copiara junto al
-- resto de la sección E, `appointment` fallaría por una FK a una tabla que
-- todavía no existe.

-- ---------------------------------------------------------------------------
-- D.0 · `customer` — RN-DAT-01, RN-DAT-03, DEC-022
-- ---------------------------------------------------------------------------
-- Toda cita referencia un cliente, también la manual: un único lugar donde
-- anonimizar (RN-DAT-03) en lugar de datos personales repartidos entre tablas.
-- Por eso `phone` y `email` son opcionales: la reserva pública los exige
-- (DEC-022) pero la cita manual puede omitirlos (RN-CIT-02).

CREATE TABLE customer (
  id             uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id  uuid        NOT NULL,
  full_name      text        NOT NULL,
  phone          text,
  email          text,
  anonymized_at  timestamptz,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT customer_id_pk               PRIMARY KEY (id),
  CONSTRAINT customer_barbershop_id_id_uk UNIQUE (barbershop_id, id),
  CONSTRAINT customer_barbershop_id_fk    FOREIGN KEY (barbershop_id)
                                          REFERENCES barbershop (id) ON DELETE RESTRICT,

  CONSTRAINT customer_full_name_ck CHECK (
    btrim(full_name) <> '' AND char_length(full_name) <= 120
  ),
  CONSTRAINT customer_phone_ck CHECK (phone IS NULL OR phone ~ '^\+[1-9][0-9]{7,14}$'),
  -- DDL-VAL-01: forma canónica explícita (sin espacios, ya recortado), no
  -- solo el patrón mínimo de arroba y punto.
  CONSTRAINT customer_email_ck CHECK (
    email IS NULL
    OR (
      email = lower(email)
      AND email = btrim(email)
      AND email !~ '\s'
      AND email LIKE '_%@_%._%'
      AND char_length(email) <= 254
    )
  ),

  -- La anonimización es idempotente y verificable (RN-DAT-03, DEC-025,
  -- DEC-049): una fila anonimizada no conserva teléfono, correo ni nombre
  -- real. El marcador es literal y único (customer_anonymize lo usa; ver
  -- comentario de esa función) para que "¿está anonimizado?" sea una
  -- comparación exacta, no una convención de texto libre.
  CONSTRAINT customer_anonymized_ck CHECK (
    anonymized_at IS NULL
    OR (phone IS NULL AND email IS NULL AND full_name = 'Cliente anonimizado')
  )
);

COMMENT ON TABLE customer IS
  'Persona que reserva. Propietario funcional: barbería. Retención: configurable (1-120 meses), '
  'contada desde la última actividad, no desde la creación (DEC-042); después, anonimización '
  'completa vía customer_anonymize, que también cubre appointment.attendee_name/customer_note, '
  'appointment_history.reason y appointment_history_change, y revoca appointment_access_token '
  '(DEC-049). Clasificación: DATO PERSONAL. '
  'La persona atendida NO se guarda aquí: vive en appointment.attendee_name (RN-RES-03).';

-- Permite reutilizar el cliente por teléfono en el flujo público sin una
-- carrera entre SELECT e INSERT. Consecuencia aceptada: un padre que reserva
-- para su hijo con el mismo teléfono es UN cliente con varias citas de
-- distintos `attendee_name`, que es exactamente el modelo de RN-RES-03.
CREATE UNIQUE INDEX idx_customer_shop_phone
  ON customer (barbershop_id, phone)
  WHERE phone IS NOT NULL AND anonymized_at IS NULL;

-- DEC-046: el correo es único por barbería, no global (a diferencia de
-- staff_user.email, que sí lo es por DEC-016/H.5). El CHECK de email ya
-- exige forma canónica en minúsculas y sin espacios, así que el índice no
-- necesita envolver lower(email).
CREATE UNIQUE INDEX idx_customer_shop_email
  ON customer (barbershop_id, email)
  WHERE email IS NOT NULL AND anonymized_at IS NULL;

CREATE INDEX idx_customer_shop_anonymization_due
  ON customer (barbershop_id, created_at)
  WHERE anonymized_at IS NULL;

CREATE TRIGGER customer_set_updated_at
  BEFORE UPDATE ON customer
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE customer ENABLE ROW LEVEL SECURITY;
ALTER TABLE customer FORCE  ROW LEVEL SECURITY;

CREATE POLICY customer_all_admin_policy ON customer
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY customer_select_tenant_policy ON customer
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY customer_insert_tenant_policy ON customer
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY customer_update_tenant_policy ON customer
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Sin DELETE: los datos personales se anonimizan, no se borran (RN-DAT-03).
GRANT SELECT, INSERT, UPDATE ON TABLE customer TO barberia_app;

-- ---------------------------------------------------------------------------
-- D.1 · `appointment`
-- ---------------------------------------------------------------------------

CREATE TABLE appointment (
  id            uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id uuid        NOT NULL,
  barber_id     uuid        NOT NULL,
  service_id    uuid        NOT NULL,
  customer_id   uuid        NOT NULL,

  -- RN-RES-02 / RN-RES-03: quien reserva y quien es atendido pueden diferir.
  -- El nombre se resuelve al crear la cita y se guarda aquí, no se deriva del
  -- cliente en tiempo de lectura.
  attendee_name text        NOT NULL,

  starts_at     timestamptz NOT NULL,
  ends_at       timestamptz NOT NULL,

  status        text        NOT NULL DEFAULT 'confirmed',

  -- estados-citas.md §11 pide explícitamente una columna derivada que exprese
  -- "ocupa agenda", para que la restricción de exclusión y el cálculo de
  -- disponibilidad no puedan divergir. Columna generada y almacenada: un
  -- único criterio, imposible de contradecir desde la aplicación.
  occupies_schedule boolean GENERATED ALWAYS AS (
    status IN ('confirmed', 'completed', 'no_show')
  ) STORED,

  origin text NOT NULL,

  -- Snapshots autorizados por DEC-004 y por la excepción 2 de
  -- estandar-base-datos.md §3. Se fijan al crear o al aplicar un cambio
  -- explícito; NUNCA se sincronizan en segundo plano con `service`.
  service_name_snapshot     text           NOT NULL,
  duration_minutes_snapshot integer        NOT NULL,
  price_amount_snapshot     numeric(12, 2) NOT NULL,
  price_currency_snapshot   char(3)        NOT NULL,

  customer_note       text,
  cancellation_reason text,

  -- Instante efectivo del resultado, para métricas. En el cierre automático
  -- (RN-CIT-05) es `ends_at`, aunque el trabajo se ejecute después.
  resolved_at timestamptz,

  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT appointment_id_pk               PRIMARY KEY (id),
  CONSTRAINT appointment_barbershop_id_id_uk UNIQUE (barbershop_id, id),

  -- Pertenencia tenant-aware en ambos lados. Una FK simple por barber_id no
  -- demostraría que el barbero es de esta barbería (estandar-base-datos.md §6).
  CONSTRAINT appointment_barbershop_id_barber_id_fk
    FOREIGN KEY (barbershop_id, barber_id)
    REFERENCES barber (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT appointment_barbershop_id_service_id_fk
    FOREIGN KEY (barbershop_id, service_id)
    REFERENCES service (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT appointment_barbershop_id_customer_id_fk
    FOREIGN KEY (barbershop_id, customer_id)
    REFERENCES customer (barbershop_id, id) ON DELETE RESTRICT,

  -- Los cinco estados confirmados. Texto, nunca números: un `estado = 3` es
  -- ilegible en soporte (estados-citas.md §11).
  CONSTRAINT appointment_status_ck CHECK (
    status IN ('confirmed', 'completed', 'cancelled_by_customer', 'cancelled_by_barber', 'no_show')
  ),
  CONSTRAINT appointment_origin_ck CHECK (origin IN ('public', 'manual')),

  CONSTRAINT appointment_interval_ck      CHECK (ends_at > starts_at),
  CONSTRAINT appointment_attendee_name_ck CHECK (
    btrim(attendee_name) <> '' AND char_length(attendee_name) <= 120
  ),
  CONSTRAINT appointment_duration_snapshot_ck CHECK (duration_minutes_snapshot BETWEEN 1 AND 1440),
  CONSTRAINT appointment_price_snapshot_ck    CHECK (price_amount_snapshot >= 0),
  CONSTRAINT appointment_currency_snapshot_ck CHECK (price_currency_snapshot ~ '^[A-Z]{3}$'),
  CONSTRAINT appointment_service_name_snapshot_ck CHECK (
    btrim(service_name_snapshot) <> '' AND char_length(service_name_snapshot) <= 120
  ),
  CONSTRAINT appointment_customer_note_ck CHECK (
    customer_note IS NULL OR char_length(customer_note) <= 500
  ),

  -- DDL-INT-04: la resta de dos timestamptz es la diferencia exacta entre dos
  -- instantes absolutos, no una suma calendárica; EXTRACT(EPOCH ...) de esa
  -- resta no depende de la zona de sesión ni de DST. Probado contra
  -- PostgreSQL real cruzando medianoche y un cambio de horario de verano.
  CONSTRAINT appointment_duration_matches_snapshot_ck CHECK (
    EXTRACT(EPOCH FROM (ends_at - starts_at)) = duration_minutes_snapshot * 60
  ),

  -- Un motivo de cancelación solo tiene sentido en una cita cancelada, y con cota (DDL-VAL-01).
  CONSTRAINT appointment_cancellation_reason_ck CHECK (
    cancellation_reason IS NULL
    OR (
      status IN ('cancelled_by_customer', 'cancelled_by_barber')
      AND char_length(cancellation_reason) <= 500
    )
  ),

  -- `confirmed` es el único estado no terminal; los cuatro restantes tienen
  -- instante de resolución, y ese instante no puede preceder a la creación
  -- de la cita (DDL-TMP-01).
  CONSTRAINT appointment_resolved_at_ck CHECK (
    (status = 'confirmed' AND resolved_at IS NULL)
    OR
    (status <> 'confirmed' AND resolved_at IS NOT NULL AND resolved_at >= created_at)
  )
);

COMMENT ON TABLE appointment IS
  'Cita (turno en la interfaz, DEC-016). Propietario funcional: barbería. '
  'Retención: permanente en lo operativo; los datos personales asociados se anonimizan según '
  'RN-DAT-03. Clasificación: negocio + dato personal (attendee_name, customer_note). '
  'Nunca se borra en cascada desde barbería, servicio ni barbero.';

COMMENT ON COLUMN appointment.occupies_schedule IS
  'Criterio ÚNICO de "ocupa agenda" (estados-citas.md §4 y §11). Lo usan la restricción de '
  'exclusión y el cálculo de disponibilidad. Cambiar el conjunto de estados que ocupan agenda '
  'se hace aquí y en ningún otro lugar.';

COMMENT ON COLUMN appointment.duration_minutes_snapshot IS
  'Duración con la que se creó la cita (DEC-004). La igualdad '
  'ends_at = starts_at + duration_minutes_snapshot no puede expresarse en un CHECK porque la '
  'suma timestamptz + interval es STABLE, no IMMUTABLE: la garantiza el servicio de citas y '
  'la verifica su prueba de integración.';

-- -------------------------------------------------------------------
-- La restricción central del producto (RN-CON-01, RN-CON-03, F-DISP-02)
-- -------------------------------------------------------------------
-- Es la ÚLTIMA línea de defensa: aunque dos procesos comprueben a la vez que
-- la franja está libre, PostgreSQL acepta uno solo (RN-CON-02). No puede
-- descansar en el frontend, en un SELECT previo ni en un bloqueo en memoria.
--
-- `'[)'` implementa el intervalo semiabierto: dos citas contiguas (10:00-11:00
-- y 11:00-12:00) NO se cruzan (RN-DIS-05).
--
-- El predicado usa la columna generada, no una lista de estados repetida: si
-- mañana `pending` ocupa agenda (estados-citas.md §2.3), se cambia la columna
-- y la restricción sigue siendo correcta sin reescribirla.
ALTER TABLE appointment
  ADD CONSTRAINT appointment_barber_interval_excl
  EXCLUDE USING gist (
    barbershop_id WITH =,
    barber_id     WITH =,
    tstzrange(starts_at, ends_at, '[)') WITH &&
  ) WHERE (occupies_schedule);

-- Agenda diaria del barbero (F-CITA-01, F-CITA-02) y cálculo de disponibilidad.
CREATE INDEX idx_appointment_shop_barber_starts_at
  ON appointment (barbershop_id, barber_id, starts_at);

-- Historial del cliente y vista de sus turnos, del más reciente al más antiguo.
CREATE INDEX idx_appointment_shop_customer_starts_at
  ON appointment (barbershop_id, customer_id, starts_at DESC);

-- DDL-PER-01: "citas por servicio" (ej. cuántas citas usó cada servicio,
-- para desactivarlo con RN-SER-03) no tenía índice; sin él, la única forma
-- de resolverlo era recorrer la tabla completa.
CREATE INDEX idx_appointment_shop_service
  ON appointment (barbershop_id, service_id);

-- Cola del cierre automático de citas vencidas (RN-CIT-05): solo las activas.
CREATE INDEX idx_appointment_open_ends_at
  ON appointment (barbershop_id, ends_at)
  WHERE status = 'confirmed';

CREATE TRIGGER appointment_set_updated_at
  BEFORE UPDATE ON appointment
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE appointment ENABLE ROW LEVEL SECURITY;
ALTER TABLE appointment FORCE  ROW LEVEL SECURITY;

CREATE POLICY appointment_all_admin_policy ON appointment
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY appointment_select_tenant_policy ON appointment
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY appointment_insert_tenant_policy ON appointment
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY appointment_update_tenant_policy ON appointment
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Sin política ni privilegio de DELETE: una cita no se borra, se cancela.
GRANT SELECT, INSERT, UPDATE ON TABLE appointment TO barberia_app;

-- ---------------------------------------------------------------------------
-- D.2 · `appointment_history` — RN-HIS-01, RN-HIS-02, DEC-014
-- ---------------------------------------------------------------------------
-- Append-only para el rol de aplicación, garantizado por dos mecanismos
-- independientes: no se conceden privilegios UPDATE/DELETE, y no existe
-- política RLS que los permita. Una entrada errónea se corrige con otra
-- entrada, nunca editando la anterior.

CREATE TABLE appointment_history (
  id                   uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id        uuid        NOT NULL,
  appointment_id       uuid        NOT NULL,
  event_type           text        NOT NULL,
  actor_type           text        NOT NULL,
  actor_staff_user_id  uuid,
  actor_customer_id    uuid,
  reason               text,
  occurred_at          timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT appointment_history_id_pk               PRIMARY KEY (id),
  CONSTRAINT appointment_history_barbershop_id_id_uk UNIQUE (barbershop_id, id),

  CONSTRAINT appointment_history_barbershop_id_appointment_id_fk
    FOREIGN KEY (barbershop_id, appointment_id)
    REFERENCES appointment (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT appointment_history_barbershop_id_actor_staff_user_id_fk
    FOREIGN KEY (barbershop_id, actor_staff_user_id)
    REFERENCES staff_user (barbershop_id, id) ON DELETE RESTRICT,

  -- DDL-HIS-01: actor_customer_id no tenía FK propia; un cliente borrado (no
  -- ocurre hoy, pero nada lo impedía a nivel de tipos) habría dejado actores
  -- huérfanos. Tenant-aware, igual que el resto de referencias de esta tabla.
  CONSTRAINT appointment_history_barbershop_id_actor_customer_id_fk
    FOREIGN KEY (barbershop_id, actor_customer_id)
    REFERENCES customer (barbershop_id, id) ON DELETE RESTRICT,

  -- Vocabulario en inglés y snake_case (DEC-041/CT-002): estados-citas.md §9
  -- usaba puntos en español, en contradicción con la regla general de la
  -- sección 11 del mismo documento. Resuelto explícitamente por el propietario.
  CONSTRAINT appointment_history_event_type_ck CHECK (
    event_type IN (
      'appointment_created',
      'appointment_rescheduled',
      'appointment_service_changed',
      'appointment_completed',
      'appointment_cancelled_by_customer',
      'appointment_cancelled_by_barber',
      'appointment_no_show',
      'appointment_status_corrected'
    )
  ),

  -- RN-HIS-01: el actor puede ser el barbero, el cliente o el sistema.
  CONSTRAINT appointment_history_actor_type_ck CHECK (
    actor_type IN ('staff', 'customer', 'system')
  ),
  CONSTRAINT appointment_history_actor_shape_ck CHECK (
    (actor_type = 'staff'    AND actor_staff_user_id IS NOT NULL AND actor_customer_id IS NULL)
    OR
    (actor_type = 'customer' AND actor_customer_id IS NOT NULL   AND actor_staff_user_id IS NULL)
    OR
    (actor_type = 'system'   AND actor_staff_user_id IS NULL     AND actor_customer_id IS NULL)
  ),

  -- La corrección auditada (T8) exige motivo obligatorio (RN-CIT-04); en los
  -- demás casos es opcional pero acotado (DDL-VAL-01).
  CONSTRAINT appointment_history_reason_ck CHECK (
    (event_type = 'appointment_status_corrected'
       AND reason IS NOT NULL AND btrim(reason) <> '' AND char_length(reason) <= 500)
    OR
    (event_type <> 'appointment_status_corrected'
       AND (reason IS NULL OR (btrim(reason) <> '' AND char_length(reason) <= 500)))
  )
);

COMMENT ON TABLE appointment_history IS
  'Historial inmutable de cambios de una cita (RN-HIS-01, RN-HIS-02, DEC-014). '
  'Propietario funcional: barbería. Retención: permanente, incluso tras anonimizar al cliente. '
  'Clasificación: auditoría. APPEND-ONLY: el rol de aplicación no tiene UPDATE ni DELETE.';

CREATE INDEX idx_appointment_history_shop_appointment_occurred
  ON appointment_history (barbershop_id, appointment_id, occurred_at);

-- DDL-HIS-01: historial de un cliente por actor (por ejemplo, para atender un
-- derecho de acceso). Parcial porque la mayoría de filas son actor_type
-- staff/system, sin actor_customer_id.
CREATE INDEX idx_appointment_history_shop_actor_customer
  ON appointment_history (barbershop_id, actor_customer_id)
  WHERE actor_customer_id IS NOT NULL;

-- DDL-PER-01: "historial por actor" también cubre al staff, el caso más
-- frecuente (la mayoría de eventos los genera el barbero o un proceso del
-- sistema, no el cliente) y que no tenía índice propio.
CREATE INDEX idx_appointment_history_shop_actor_staff
  ON appointment_history (barbershop_id, actor_staff_user_id)
  WHERE actor_staff_user_id IS NOT NULL;

ALTER TABLE appointment_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE appointment_history FORCE  ROW LEVEL SECURITY;

CREATE POLICY appointment_history_all_admin_policy ON appointment_history
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY appointment_history_select_tenant_policy ON appointment_history
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY appointment_history_insert_tenant_policy ON appointment_history
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Deliberadamente sin políticas de UPDATE ni DELETE.
GRANT SELECT, INSERT ON TABLE appointment_history TO barberia_app;
REVOKE UPDATE, DELETE ON TABLE appointment_history FROM barberia_app;

-- Valores anterior y nuevo de cada campo tocado (RN-HIS-01). Tabla hija en
-- lugar de `jsonb`: mantiene 1FN y evita la dependencia prohibida por AGENTS.md.
CREATE TABLE appointment_history_change (
  id              uuid NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id   uuid NOT NULL,
  history_id      uuid NOT NULL,
  field_name      text NOT NULL,
  previous_value  text,
  new_value       text,

  CONSTRAINT appointment_history_change_id_pk PRIMARY KEY (id),

  CONSTRAINT appointment_history_change_barbershop_id_history_id_fk
    FOREIGN KEY (barbershop_id, history_id)
    REFERENCES appointment_history (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT appointment_history_change_field_name_uk
    UNIQUE (barbershop_id, history_id, field_name),

  CONSTRAINT appointment_history_change_field_name_ck CHECK (
    btrim(field_name) <> '' AND char_length(field_name) <= 60
  ),
  -- Una entrada sin ningún valor no aporta nada.
  CONSTRAINT appointment_history_change_value_ck CHECK (
    previous_value IS NOT NULL OR new_value IS NOT NULL
  ),
  -- DDL-VAL-01: cota generosa pero explícita; antes no tenían límite.
  CONSTRAINT appointment_history_change_previous_value_ck CHECK (
    previous_value IS NULL OR char_length(previous_value) <= 1000
  ),
  CONSTRAINT appointment_history_change_new_value_ck CHECK (
    new_value IS NULL OR char_length(new_value) <= 1000
  )
);

COMMENT ON TABLE appointment_history_change IS
  'Campos modificados con su valor anterior y nuevo (RN-HIS-01). APPEND-ONLY, igual que su padre. '
  'Clasificación: auditoría; puede contener datos personales y entra en la anonimización de RN-DAT-03.';

ALTER TABLE appointment_history_change ENABLE ROW LEVEL SECURITY;
ALTER TABLE appointment_history_change FORCE  ROW LEVEL SECURITY;

CREATE POLICY appointment_history_change_all_admin_policy ON appointment_history_change
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY appointment_history_change_select_tenant_policy ON appointment_history_change
  FOR SELECT TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY appointment_history_change_insert_tenant_policy ON appointment_history_change
  FOR INSERT TO barberia_app WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT ON TABLE appointment_history_change TO barberia_app;
REVOKE UPDATE, DELETE ON TABLE appointment_history_change FROM barberia_app;


-- ===========================================================================
-- SECCIÓN E · B4 — Reserva pública y disponibilidad
-- ===========================================================================
-- Funciones F-PUB-01 a F-PUB-08, F-DISP-01, F-DISP-03 a F-DISP-05, F-CITA-07.

-- ---------------------------------------------------------------------------
-- E.1 · `appointment_access_token` — F-PUB-07, DEC-022
-- ---------------------------------------------------------------------------
-- `customer` se trasladó a D.0 (DDL-INT-00): toda cita la referencia y el
-- archivo debe ejecutarse en orden desde una base vacía.

CREATE TABLE appointment_access_token (
  id             uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id  uuid        NOT NULL,
  appointment_id uuid        NOT NULL,
  token_hash     text        NOT NULL,
  issued_at      timestamptz NOT NULL DEFAULT now(),
  expires_at     timestamptz,
  revoked_at     timestamptz,

  CONSTRAINT appointment_access_token_id_pk PRIMARY KEY (id),

  CONSTRAINT appointment_access_token_barbershop_id_appointment_id_fk
    FOREIGN KEY (barbershop_id, appointment_id)
    REFERENCES appointment (barbershop_id, id) ON DELETE RESTRICT,

  -- Unicidad global: el enlace se resuelve antes de conocer la barbería.
  CONSTRAINT appointment_access_token_token_hash_uk UNIQUE (token_hash),
  -- DDL-VAL-01: hexadecimal minúscula de 64 caracteres (SHA-256), no solo longitud.
  CONSTRAINT appointment_access_token_token_hash_ck CHECK (token_hash ~ '^[0-9a-f]{64}$'),
  CONSTRAINT appointment_access_token_expires_at_ck CHECK (
    expires_at IS NULL OR expires_at > issued_at
  ),
  -- DDL-TMP-01: la revocación no puede preceder a la emisión.
  CONSTRAINT appointment_access_token_revoked_at_ck CHECK (
    revoked_at IS NULL OR revoked_at >= issued_at
  )
);

COMMENT ON TABLE appointment_access_token IS
  'Token del enlace aleatorio largo con el que el cliente consulta o cancela su turno (DEC-022). '
  'Propietario funcional: barbería. Retención: hasta la anonimización de la cita, que lo revoca '
  '(RN-DAT-03). Clasificación: secreto. Solo se almacena el hash; el valor viaja una vez, por correo.';

-- DDL-PER-01: sin predicado parcial. Un `WHERE revoked_at IS NULL` aparte
-- duplicaría el prefijo; este índice completo sirve tanto la resolución
-- pública (busca uno vigente) como la auditoría de tokens ya revocados de
-- una cita.
CREATE INDEX idx_appointment_access_token_shop_appointment
  ON appointment_access_token (barbershop_id, appointment_id);

ALTER TABLE appointment_access_token ENABLE ROW LEVEL SECURITY;
ALTER TABLE appointment_access_token FORCE  ROW LEVEL SECURITY;

CREATE POLICY appointment_access_token_all_admin_policy ON appointment_access_token
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY appointment_access_token_select_tenant_policy ON appointment_access_token
  FOR SELECT TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY appointment_access_token_insert_tenant_policy ON appointment_access_token
  FOR INSERT TO barberia_app WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY appointment_access_token_update_tenant_policy ON appointment_access_token
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT, UPDATE ON TABLE appointment_access_token TO barberia_app;

-- ---------------------------------------------------------------------------
-- E.2 · Resolución pública previa al contexto de tenant
-- ---------------------------------------------------------------------------
-- Mismo patrón acotado de la sección A.0, por el mismo motivo: el enlace
-- público y el enlace del turno llegan sin sesión y sin barbería conocida.

CREATE OR REPLACE FUNCTION public_resolve_barbershop_by_slug(p_slug text)
RETURNS uuid
LANGUAGE sql STABLE SECURITY DEFINER
SET search_path = ''
AS $$
  SELECT id FROM public.barbershop
  WHERE pg_catalog.lower(public_slug) = pg_catalog.lower(p_slug)
$$;

CREATE OR REPLACE FUNCTION public_resolve_appointment_token_tenant(p_token_hash text)
RETURNS uuid
LANGUAGE sql STABLE SECURITY DEFINER
SET search_path = ''
AS $$
  SELECT barbershop_id
  FROM public.appointment_access_token
  WHERE token_hash = p_token_hash
    AND revoked_at IS NULL
    AND (expires_at IS NULL OR expires_at > pg_catalog.now())
$$;

REVOKE ALL     ON FUNCTION public_resolve_barbershop_by_slug(text)          FROM PUBLIC;
REVOKE ALL     ON FUNCTION public_resolve_appointment_token_tenant(text)    FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION public_resolve_barbershop_by_slug(text)          TO barberia_app;
GRANT  EXECUTE ON FUNCTION public_resolve_appointment_token_tenant(text)    TO barberia_app;

-- ---------------------------------------------------------------------------
-- E.3 · Parámetros configurables de reserva y cancelación — DEC-018
-- ---------------------------------------------------------------------------
-- Columnas tipadas en `barbershop`, no una tabla clave/valor ni `jsonb`: cada
-- parámetro depende de la identidad de la barbería (3FN), tiene su propio tipo
-- y su propio rango verificable.

ALTER TABLE barbershop
  ADD COLUMN min_lead_minutes                     integer NOT NULL DEFAULT 60,
  ADD COLUMN max_booking_window_days              integer NOT NULL DEFAULT 3,
  ADD COLUMN slot_step_minutes                    integer NOT NULL DEFAULT 15,
  ADD COLUMN cancellation_deadline_minutes        integer NOT NULL DEFAULT 20,
  ADD COLUMN late_cancellation_policy             text    NOT NULL DEFAULT 'barber_only',
  ADD COLUMN late_cancellation_reason_required    boolean NOT NULL DEFAULT false,
  ADD COLUMN appointment_closing_mode             text    NOT NULL DEFAULT 'manual',
  ADD COLUMN auto_close_delay_hours               integer,

  -- Valores iniciales de DEC-018. Los rangos impiden una configuración que
  -- deje la agenda inoperante, sin cerrar la libertad del barbero.
  ADD CONSTRAINT barbershop_min_lead_minutes_ck        CHECK (min_lead_minutes BETWEEN 0 AND 10080),
  ADD CONSTRAINT barbershop_max_booking_window_days_ck CHECK (max_booking_window_days BETWEEN 1 AND 365),
  ADD CONSTRAINT barbershop_slot_step_minutes_ck       CHECK (slot_step_minutes BETWEEN 5 AND 120),
  ADD CONSTRAINT barbershop_cancellation_deadline_ck   CHECK (cancellation_deadline_minutes BETWEEN 0 AND 10080),
  ADD CONSTRAINT barbershop_late_cancellation_policy_ck CHECK (
    late_cancellation_policy IN ('barber_only', 'customer_allowed')
  ),
  ADD CONSTRAINT barbershop_appointment_closing_mode_ck CHECK (
    appointment_closing_mode IN ('manual', 'automatic')
  ),
  -- El retraso del cierre automático existe si y solo si el modo lo usa.
  ADD CONSTRAINT barbershop_auto_close_delay_ck CHECK (
    (appointment_closing_mode = 'automatic' AND auto_close_delay_hours IS NOT NULL
       AND auto_close_delay_hours BETWEEN 1 AND 168)
    OR
    (appointment_closing_mode = 'manual' AND auto_close_delay_hours IS NULL)
  ),
  -- DDL-CFG-01: la anticipación mínima no puede superar toda la ventana
  -- pública de reserva; si lo hiciera, ninguna franja quedaría reservable.
  ADD CONSTRAINT barbershop_min_lead_vs_window_ck CHECK (
    min_lead_minutes <= max_booking_window_days * 1440
  );


-- ===========================================================================
-- SECCIÓN F · B5 — Notificaciones y recordatorios
-- ===========================================================================
-- Funciones F-NOT-01 a F-NOT-03. Reglas RN-REC-01 a RN-REC-06, RN-DAT-02,
-- DEC-015, DEC-027, DEC-032.

-- ---------------------------------------------------------------------------
-- F.1 · Canales por evento (RN-REC-05, DEC-027)
-- ---------------------------------------------------------------------------

CREATE TABLE notification_channel_setting (
  barbershop_id uuid        NOT NULL,
  event_type    text        NOT NULL,
  channel       text        NOT NULL,
  is_enabled    boolean     NOT NULL DEFAULT false,
  updated_at    timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT notification_channel_setting_pk PRIMARY KEY (barbershop_id, event_type, channel),
  CONSTRAINT notification_channel_setting_barbershop_id_fk
    FOREIGN KEY (barbershop_id) REFERENCES barbershop (id) ON DELETE CASCADE,

  CONSTRAINT notification_channel_setting_event_type_ck CHECK (
    event_type IN ('confirmation', 'reminder', 'reschedule', 'cancellation', 'delay', 'no_show')
  ),
  -- DEC-027: solo correo y la interfaz OFICIAL de WhatsApp.
  CONSTRAINT notification_channel_setting_channel_ck CHECK (channel IN ('email', 'whatsapp'))
);

COMMENT ON TABLE notification_channel_setting IS
  'Matriz evento x canal por barbería (RN-REC-05). Propietario funcional: barbería. '
  'Retención: mientras exista la barbería. Clasificación: configuración. '
  'El aviso de inasistencia nace desactivado (DEC-018): is_enabled = false por defecto.';

CREATE TRIGGER notification_channel_setting_set_updated_at
  BEFORE UPDATE ON notification_channel_setting
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE notification_channel_setting ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_channel_setting FORCE  ROW LEVEL SECURITY;

-- DDL-RLS-01/H.7: ídem service_all_admin_policy.
CREATE POLICY notification_channel_setting_all_admin_policy ON notification_channel_setting
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY notification_channel_setting_select_tenant_policy ON notification_channel_setting
  FOR SELECT TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY notification_channel_setting_insert_tenant_policy ON notification_channel_setting
  FOR INSERT TO barberia_app WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY notification_channel_setting_update_tenant_policy ON notification_channel_setting
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY notification_channel_setting_delete_tenant_policy ON notification_channel_setting
  FOR DELETE TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE notification_channel_setting TO barberia_app;

-- ---------------------------------------------------------------------------
-- F.2 · Reglas de recordatorio (DEC-018, RN-REC-06)
-- ---------------------------------------------------------------------------
-- DEC-018 fija "1 recordatorio, 30 minutos antes, configurable entre 0 y 3"
-- pero NO dice qué anticipación tienen el segundo y el tercero. Una tabla hija
-- resuelve el vacío sin inventar valores: cada recordatorio configurado lleva
-- su propia anticipación, y "cantidad = 0" es simplemente ninguna fila.

CREATE TABLE barbershop_reminder_rule (
  barbershop_id uuid     NOT NULL,
  ordinal       smallint NOT NULL,
  lead_minutes  integer  NOT NULL,

  CONSTRAINT barbershop_reminder_rule_pk PRIMARY KEY (barbershop_id, ordinal),
  CONSTRAINT barbershop_reminder_rule_barbershop_id_fk
    FOREIGN KEY (barbershop_id) REFERENCES barbershop (id) ON DELETE CASCADE,

  CONSTRAINT barbershop_reminder_rule_ordinal_ck      CHECK (ordinal BETWEEN 1 AND 3),
  CONSTRAINT barbershop_reminder_rule_lead_minutes_ck CHECK (lead_minutes BETWEEN 5 AND 10080),
  CONSTRAINT barbershop_reminder_rule_lead_uk         UNIQUE (barbershop_id, lead_minutes)
);

COMMENT ON TABLE barbershop_reminder_rule IS
  'Cuántos recordatorios y con qué anticipación (DEC-018: 1 recordatorio a 30 minutos por '
  'defecto, hasta 3). Cero filas significa "sin recordatorios". Clasificación: configuración.';

ALTER TABLE barbershop_reminder_rule ENABLE ROW LEVEL SECURITY;
ALTER TABLE barbershop_reminder_rule FORCE  ROW LEVEL SECURITY;

CREATE POLICY barbershop_reminder_rule_all_admin_policy ON barbershop_reminder_rule
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY barbershop_reminder_rule_select_tenant_policy ON barbershop_reminder_rule
  FOR SELECT TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY barbershop_reminder_rule_insert_tenant_policy ON barbershop_reminder_rule
  FOR INSERT TO barberia_app WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY barbershop_reminder_rule_update_tenant_policy ON barbershop_reminder_rule
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY barbershop_reminder_rule_delete_tenant_policy ON barbershop_reminder_rule
  FOR DELETE TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE barbershop_reminder_rule TO barberia_app;

-- DDL-BIZ-01: cero filas de recordatorio para una barbería nueva significaba
-- "sin recordatorios", contradiciendo el valor inicial de DEC-018 (1
-- recordatorio a 30 minutos por defecto). El disparador lo crea al nacer la
-- barbería; el propietario puede borrar la fila después si de verdad no
-- quiere recordatorios, y esa elección explícita persiste.
CREATE OR REPLACE FUNCTION barbershop_seed_default_reminder_rule()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $$
BEGIN
  INSERT INTO public.barbershop_reminder_rule (barbershop_id, ordinal, lead_minutes)
  VALUES (NEW.id, 1, 30);
  RETURN NEW;
END;
$$;

REVOKE ALL     ON FUNCTION barbershop_seed_default_reminder_rule() FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION barbershop_seed_default_reminder_rule() TO barberia_app;

COMMENT ON FUNCTION barbershop_seed_default_reminder_rule() IS
  'Crea el recordatorio ordinal 1 a 30 minutos por defecto al crear una barbería '
  '(DEC-018, DDL-BIZ-01).';

CREATE TRIGGER barbershop_seed_default_reminder_rule_trg
  AFTER INSERT ON barbershop
  FOR EACH ROW EXECUTE FUNCTION barbershop_seed_default_reminder_rule();

-- Alcance retroactivo (DDL-BIZ-01): barberías creadas antes de que esta
-- sección se aplicara reciben la misma regla por defecto si aún no tienen
-- ninguna configuración de recordatorio ordinal 1. Sentencia de una sola vez,
-- segura de repetir: no toca una barbería que ya tiene su ordinal 1 definido,
-- sea cual sea su valor, ni una que borró la fila a propósito.
INSERT INTO barbershop_reminder_rule (barbershop_id, ordinal, lead_minutes)
SELECT b.id, 1, 30
FROM barbershop b
WHERE NOT EXISTS (
  SELECT 1 FROM barbershop_reminder_rule r
  WHERE r.barbershop_id = b.id AND r.ordinal = 1
);

-- ---------------------------------------------------------------------------
-- F.3 · `notification_schedule` — RN-REC-01, RN-REC-06, DEC-032
-- ---------------------------------------------------------------------------
-- La programación nace en la MISMA transacción que crea o modifica la cita.
-- El trabajador solo ejecuta lo ya programado; nunca descubre tardíamente qué
-- citas debían tener recordatorio.

CREATE TABLE notification_schedule (
  id                uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id     uuid        NOT NULL,
  appointment_id    uuid        NOT NULL,
  event_type        text        NOT NULL,
  channel           text        NOT NULL,
  ordinal           smallint    NOT NULL DEFAULT 1,
  scheduled_for     timestamptz NOT NULL,
  status            text        NOT NULL DEFAULT 'pending',
  -- DDL-CON-01: lease de reclamación. Solo tienen valor mientras
  -- status = 'processing'; notification_claim_due los asigna y
  -- notification_finalize_claim los limpia al cerrar, sea cual sea el
  -- desenlace.
  claim_token       uuid,
  claimed_at        timestamptz,
  lease_expires_at  timestamptz,
  cancelled_at      timestamptz,
  sent_at           timestamptz,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT notification_schedule_id_pk               PRIMARY KEY (id),
  CONSTRAINT notification_schedule_barbershop_id_id_uk UNIQUE (barbershop_id, id),

  CONSTRAINT notification_schedule_barbershop_id_appointment_id_fk
    FOREIGN KEY (barbershop_id, appointment_id)
    REFERENCES appointment (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT notification_schedule_event_type_ck CHECK (
    event_type IN ('confirmation', 'reminder', 'reschedule', 'cancellation', 'delay', 'no_show')
  ),
  CONSTRAINT notification_schedule_channel_ck CHECK (channel IN ('email', 'whatsapp')),
  CONSTRAINT notification_schedule_ordinal_ck CHECK (ordinal BETWEEN 1 AND 3),

  -- `processing` es el estado de lease (DEC-053, DDL-CON-01): un trabajador
  -- la reclamó y aún no confirma el desenlace. `skipped` cubre dos casos que
  -- comparten la misma forma de "no se va a enviar": la excepción de
  -- RN-REC-06 (cita sin canal habilitado) y un fallo permanente de envío que
  -- agotó reintentos (notification_finalize_claim, outcome
  -- 'permanent_failure'); no se crea un estado terminal nuevo solo para
  -- distinguir el motivo, que ya vive en notification_attempt.
  CONSTRAINT notification_schedule_status_ck CHECK (
    status IN ('pending', 'processing', 'sent', 'cancelled', 'skipped')
  ),
  -- DDL-TMP-01 y DDL-CON-01: además de exigir el instante correcto por
  -- estado, cada estado fija la forma exacta de las columnas de lease para
  -- que ninguna fila quede a medio reclamar ni conserve un token vencido.
  CONSTRAINT notification_schedule_outcome_ck CHECK (
    (status = 'pending'
       AND claim_token IS NULL AND claimed_at IS NULL AND lease_expires_at IS NULL
       AND sent_at IS NULL AND cancelled_at IS NULL)
    OR
    (status = 'processing'
       AND claim_token IS NOT NULL AND claimed_at IS NOT NULL
       AND lease_expires_at IS NOT NULL AND lease_expires_at > claimed_at
       AND sent_at IS NULL AND cancelled_at IS NULL)
    OR
    (status = 'sent'
       AND claim_token IS NULL AND claimed_at IS NULL AND lease_expires_at IS NULL
       AND sent_at IS NOT NULL AND cancelled_at IS NULL AND sent_at >= created_at)
    OR
    (status = 'cancelled'
       AND claim_token IS NULL AND claimed_at IS NULL AND lease_expires_at IS NULL
       AND cancelled_at IS NOT NULL AND cancelled_at >= created_at)
    OR
    (status = 'skipped'
       AND claim_token IS NULL AND claimed_at IS NULL AND lease_expires_at IS NULL)
  )
);

COMMENT ON TABLE notification_schedule IS
  'Programación transaccional de envíos (RN-REC-06, DEC-032). Propietario funcional: plataforma. '
  'Retención: la de la cita. Clasificación: técnico. NO almacena el contenido del mensaje: '
  'RN-REC-03 exige construirlo en el instante del envío, nunca al programarlo. '
  'El lease de reclamación (claim_token/claimed_at/lease_expires_at) es de barberia_worker: '
  'barberia_app no tiene privilegio de columna sobre esas tres columnas ni sobre sent_at (DEC-053).';

-- Una sola programación lógica vigente por cita, tipo, canal y ordinal
-- (estandar-base-datos.md §7). Es la restricción que impide que dos
-- reprogramaciones seguidas dejen tres recordatorios (RN-REC-01). Incluye
-- 'processing': una fila reclamada por el trabajador sigue vigente y no debe
-- dejar hueco para que una reprogramación cree una segunda fila activa.
CREATE UNIQUE INDEX idx_notification_schedule_pending_unique
  ON notification_schedule (barbershop_id, appointment_id, event_type, channel, ordinal)
  WHERE status IN ('pending', 'processing');

-- DDL-PER-01: el índice de arriba es parcial y no puede ensancharse sin
-- romper la unicidad de negocio que garantiza. El historial de
-- notificaciones de una cita (enviadas, canceladas, omitidas incluidas)
-- necesita uno aparte, sin predicado.
CREATE INDEX idx_notification_schedule_shop_appointment
  ON notification_schedule (barbershop_id, appointment_id);

-- Cola del trabajador (base-datos.md §12). El índice NO lleva barbershop_id
-- delante: el trabajador reclama trabajo pendiente de todas las barberías.
CREATE INDEX idx_notification_schedule_due
  ON notification_schedule (scheduled_for)
  WHERE status = 'pending';

-- DDL-CON-01: recuperación de leases vencidos. notification_claim_due la usa
-- para encontrar filas 'processing' cuyo trabajador desapareció sin finalizar
-- (caída tras el claim), en la misma llamada que atiende lo nuevo pendiente.
CREATE INDEX idx_notification_schedule_lease_expired
  ON notification_schedule (lease_expires_at)
  WHERE status = 'processing';

CREATE TRIGGER notification_schedule_set_updated_at
  BEFORE UPDATE ON notification_schedule
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE notification_schedule ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_schedule FORCE  ROW LEVEL SECURITY;

CREATE POLICY notification_schedule_all_admin_policy ON notification_schedule
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY notification_schedule_select_tenant_policy ON notification_schedule
  FOR SELECT TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY notification_schedule_insert_tenant_policy ON notification_schedule
  FOR INSERT TO barberia_app WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY notification_schedule_update_tenant_policy ON notification_schedule
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- DDL-CON-01/DEC-053: barberia_app NUNCA obtiene privilegio de columna sobre
-- claim_token/claimed_at/lease_expires_at/sent_at. Solo puede cancelar una
-- programación propia (status, cancelled_at); el CHECK de forma ya impide
-- cancelar una fila 'processing' porque exigiría claim_token NULL a la vez
-- que status = 'cancelled'. Reclamar y finalizar son exclusivos de
-- barberia_worker, a través de las funciones SECURITY DEFINER de más abajo.
GRANT SELECT, INSERT ON TABLE notification_schedule TO barberia_app;
GRANT UPDATE (status, cancelled_at) ON TABLE notification_schedule TO barberia_app;

-- -------------------------------------------------------------------
-- El trabajador, el aislamiento por barbería y el lease de reclamación
-- -------------------------------------------------------------------
-- Problema (DDL-CON-01): RLS forzada exige `app.barbershop_id`, pero el
-- trabajador debe encontrar envíos vencidos SIN saber de antemano de qué
-- barberías son, y el envío es una llamada de red que NUNCA debe ejecutarse
-- dentro de una transacción de base de datos abierta (base-datos.md §11).
-- Con el `SELECT ... FOR UPDATE` original, confirmar la transacción de
-- reclamo liberaba la fila de inmediato: si el envío tardaba, otro
-- trabajador podía reclamarla y enviarla dos veces; si el proceso caía
-- entre el claim y el envío, nada la recuperaba.
--
-- Protocolo de lease (DEC-053):
--   1. `notification_claim_due` hace un `UPDATE ... SET status =
--      'processing', claim_token, claimed_at, lease_expires_at` sobre un
--      `SELECT ... FOR UPDATE SKIP LOCKED` sin esperar; confirma la
--      transacción de inmediato y devuelve el token. El envío ocurre FUERA
--      de esa transacción.
--   2. La misma llamada recupera leases vencidos: su WHERE acepta tanto
--      'pending' vencido como 'processing' cuyo lease_expires_at ya pasó
--      (un trabajador que cayó tras el claim no bloquea el envío para
--      siempre; `idx_notification_schedule_lease_expired` la sirve).
--   3. `notification_finalize_claim` cierra por CAS de `claim_token`:
--      'sent' (terminal), 'retry' (vuelve a 'pending', reclamable de
--      inmediato: fallo temporal) o 'permanent_failure' (pasa a 'skipped':
--      fallo permanente, agotó reintentos). Si el CAS no encuentra la fila
--      con ese token —porque el lease ya venció y otro trabajador la
--      reclamó de nuevo— devuelve `false` sin lanzar excepción: quien
--      llegó tarde no debe tratarlo como error, solo dejar de insistir.
CREATE OR REPLACE FUNCTION notification_claim_due(
  p_limit         integer,
  p_lease_seconds integer,
  p_now           timestamptz
)
RETURNS TABLE (schedule_id uuid, barbershop_id uuid, claim_token uuid)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
BEGIN
  IF p_limit IS NULL OR p_limit < 1 OR p_limit > 200 THEN
    RAISE EXCEPTION 'notification_claim_due: p_limit fuera de rango (1-200).';
  END IF;
  IF p_lease_seconds IS NULL OR p_lease_seconds < 30 OR p_lease_seconds > 3600 THEN
    RAISE EXCEPTION 'notification_claim_due: p_lease_seconds fuera de rango (30-3600).';
  END IF;
  IF p_now IS NULL THEN
    RAISE EXCEPTION 'notification_claim_due: p_now no admite NULL.';
  END IF;

  RETURN QUERY
  WITH due AS (
    SELECT ns.id
    FROM public.notification_schedule ns
    WHERE (ns.status = 'pending'    AND ns.scheduled_for    <= p_now)
       OR (ns.status = 'processing' AND ns.lease_expires_at <= p_now)
    ORDER BY ns.scheduled_for
    LIMIT p_limit
    FOR UPDATE SKIP LOCKED
  )
  UPDATE public.notification_schedule ns
  SET status           = 'processing',
      claim_token      = pg_catalog.gen_random_uuid(),
      claimed_at       = p_now,
      lease_expires_at = p_now + pg_catalog.make_interval(secs => p_lease_seconds)
  FROM due
  WHERE ns.id = due.id
  RETURNING ns.id, ns.barbershop_id, ns.claim_token;
END;
$$;

-- Función de worker (DEC-040, DDL-SEC-04): solo barberia_worker la ejecuta,
-- nunca barberia_app.
REVOKE ALL     ON FUNCTION notification_claim_due(integer, integer, timestamptz) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION notification_claim_due(integer, integer, timestamptz) TO barberia_worker;

COMMENT ON FUNCTION notification_claim_due(integer, integer, timestamptz) IS
  'Reclama un lote pequeño de envíos vencidos entre barberías con un lease de p_lease_seconds '
  '(DEC-053, DDL-CON-01); en la misma llamada recupera leases vencidos de intentos anteriores. '
  'Devuelve SOLO identificadores y el claim_token para finalizar. El trabajador vuelve a validar '
  'el estado vigente de la cita antes de enviar (RN-REC-02, RN-REC-03), envía FUERA de esta '
  'transacción y cierra con notification_finalize_claim dentro de una transacción con '
  'app.barbershop_id fijado.';

CREATE OR REPLACE FUNCTION notification_finalize_claim(
  p_barbershop_id uuid,
  p_schedule_id   uuid,
  p_claim_token   uuid,
  p_outcome       text,
  p_now           timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_updated integer;
BEGIN
  IF p_barbershop_id IS NULL OR p_schedule_id IS NULL OR p_claim_token IS NULL
     OR p_outcome IS NULL OR p_now IS NULL THEN
    RAISE EXCEPTION 'notification_finalize_claim: ningún argumento admite NULL.';
  END IF;
  IF p_outcome NOT IN ('sent', 'retry', 'permanent_failure') THEN
    RAISE EXCEPTION 'notification_finalize_claim: p_outcome inválido (%).', p_outcome;
  END IF;

  IF p_outcome = 'sent' THEN
    UPDATE public.notification_schedule
    SET status = 'sent', sent_at = p_now,
        claim_token = NULL, claimed_at = NULL, lease_expires_at = NULL
    WHERE barbershop_id = p_barbershop_id AND id = p_schedule_id
      AND status = 'processing' AND claim_token = p_claim_token;
  ELSIF p_outcome = 'retry' THEN
    -- Fallo temporal: suelta el lease y vuelve a 'pending' para que
    -- cualquier trabajador la reclame de nuevo sin esperar a que venza.
    UPDATE public.notification_schedule
    SET status = 'pending',
        claim_token = NULL, claimed_at = NULL, lease_expires_at = NULL
    WHERE barbershop_id = p_barbershop_id AND id = p_schedule_id
      AND status = 'processing' AND claim_token = p_claim_token;
  ELSE
    -- Fallo permanente: reutiliza 'skipped', la misma forma que RN-REC-06
    -- (DEC-053); el motivo distingue una de otra en notification_attempt.
    UPDATE public.notification_schedule
    SET status = 'skipped',
        claim_token = NULL, claimed_at = NULL, lease_expires_at = NULL
    WHERE barbershop_id = p_barbershop_id AND id = p_schedule_id
      AND status = 'processing' AND claim_token = p_claim_token;
  END IF;

  GET DIAGNOSTICS v_updated = ROW_COUNT;
  -- CAS fallido (token distinto, ya finalizada, o el lease venció y otro
  -- trabajador la reclamó de nuevo): devuelve false, no lanza excepción.
  -- Quien llega tarde debe descartar su intento, no tratarlo como error.
  RETURN v_updated = 1;
END;
$$;

REVOKE ALL     ON FUNCTION notification_finalize_claim(uuid, uuid, uuid, text, timestamptz) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION notification_finalize_claim(uuid, uuid, uuid, text, timestamptz) TO barberia_worker;

COMMENT ON FUNCTION notification_finalize_claim(uuid, uuid, uuid, text, timestamptz) IS
  'Cierra por CAS de claim_token la reclamación abierta por notification_claim_due (DEC-053, '
  'DDL-CON-01). outcome=''sent'' es terminal; ''retry'' suelta el lease para un fallo temporal; '
  '''permanent_failure'' pasa a ''skipped'' cuando se agotan los reintentos. Devuelve false, sin '
  'lanzar excepción, si el token ya no coincide (finalización tardía tras recuperación de lease).';

-- ---------------------------------------------------------------------------
-- F.4 · `notification_attempt` — RN-REC-04, DEC-015
-- ---------------------------------------------------------------------------
-- Registro TÉCNICO, separado del historial de negocio de la cita (DEC-015), y
-- sin un solo dato personal (RN-DAT-02): ni teléfono, ni correo, ni contenido.

CREATE TABLE notification_attempt (
  id                       uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id            uuid        NOT NULL,
  notification_schedule_id uuid        NOT NULL,
  attempt_number           smallint    NOT NULL,
  result                   text        NOT NULL,
  template_key             text        NOT NULL,
  template_version         text        NOT NULL,
  provider_message_id      text,
  error_code               text,
  attempted_at             timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT notification_attempt_id_pk PRIMARY KEY (id),

  CONSTRAINT notification_attempt_barbershop_id_schedule_id_fk
    FOREIGN KEY (barbershop_id, notification_schedule_id)
    REFERENCES notification_schedule (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT notification_attempt_number_uk
    UNIQUE (barbershop_id, notification_schedule_id, attempt_number),

  CONSTRAINT notification_attempt_number_ck  CHECK (attempt_number BETWEEN 1 AND 20),
  CONSTRAINT notification_attempt_result_ck  CHECK (
    result IN ('delivered', 'failed', 'rejected')
  ),
  -- DDL-VAL-01: error_code también acotado, no solo NULL-eable según result.
  CONSTRAINT notification_attempt_error_code_ck CHECK (
    (result = 'delivered' AND error_code IS NULL)
    OR (result <> 'delivered' AND (error_code IS NULL OR char_length(error_code) <= 100))
  ),
  CONSTRAINT notification_attempt_template_ck CHECK (
    btrim(template_key) <> '' AND char_length(template_key) <= 80
    AND btrim(template_version) <> '' AND char_length(template_version) <= 20
  ),
  CONSTRAINT notification_attempt_provider_message_id_ck CHECK (
    provider_message_id IS NULL OR char_length(provider_message_id) <= 200
  )
);

COMMENT ON TABLE notification_attempt IS
  'Evidencia de cada intento de envío, incluidos los fallidos (RN-REC-04, DEC-015). '
  'Propietario funcional: plataforma. Retención: la de la cita. Clasificación: técnico. '
  'APPEND-ONLY. Guarda plantilla y versión, NUNCA el contenido, el teléfono ni el correo '
  '(RN-DAT-02). Un error del proveedor se depura antes de escribirlo en error_code. '
  'Sin columna channel propia (DDL-NOR-01): se lee de notification_schedule.channel a través '
  'de notification_schedule_id, evitando que las dos copias puedan discrepar.';

CREATE INDEX idx_notification_attempt_shop_schedule
  ON notification_attempt (barbershop_id, notification_schedule_id, attempted_at);

ALTER TABLE notification_attempt ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_attempt FORCE  ROW LEVEL SECURITY;

CREATE POLICY notification_attempt_all_admin_policy ON notification_attempt
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY notification_attempt_select_tenant_policy ON notification_attempt
  FOR SELECT TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY notification_attempt_insert_tenant_policy ON notification_attempt
  FOR INSERT TO barberia_app WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT ON TABLE notification_attempt TO barberia_app;
REVOKE UPDATE, DELETE ON TABLE notification_attempt FROM barberia_app;


-- ===========================================================================
-- SECCIÓN G · B6 — Retención y anonimización
-- ===========================================================================
-- Funciones F-SEG-02, F-OPS-05, F-OPS-06, F-OPS-09. Reglas RN-DAT-01 a
-- RN-DAT-03, DEC-025.

ALTER TABLE barbershop
  ADD COLUMN personal_data_retention_months integer NOT NULL DEFAULT 24,
  ADD CONSTRAINT barbershop_retention_months_ck CHECK (
    personal_data_retention_months BETWEEN 1 AND 120
  );

COMMENT ON COLUMN barbershop.personal_data_retention_months IS
  'Plazo efectivo de conservación de datos personales (DEC-025). Valor inicial 24 meses. '
  'AMPLIARLO exige base legal documentada: cambiar este número no basta.';

-- El trabajador de anonimización tiene el mismo problema de alcance que el de
-- notificaciones, pero NO el mismo problema de concurrencia (DEC-053,
-- DDL-CON-01): el efecto es una escritura SQL sobre esta misma base de
-- datos, no una llamada de red, así que corre dentro de la MISMA
-- transacción que hace el claim (customer_anonymize, más abajo). El
-- `FOR UPDATE OF c SKIP LOCKED` mantiene la fila bloqueada hasta que esa
-- transacción confirma; no hace falta un lease con claim_token porque no
-- hay una ventana externa que proteger. Si la anonimización completa
-- terminara partiéndose en varias transacciones, esta decisión debe
-- revisarse y adoptar el mismo protocolo que notification_claim_due.
--
-- DEC-042: la fecha ancla es la ÚLTIMA ACTIVIDAD, no c.created_at. Se
-- calcula como el mayor entre: la creación del propio cliente (piso, para
-- un cliente sin ninguna cita), la creación de su cita más reciente, la
-- transición de historial más reciente sobre alguna de sus citas, y la
-- emisión del enlace público más reciente sobre alguna de sus citas (mejor
-- proxy disponible de "acceso por token": el esquema no registra el
-- instante en que el cliente USA el enlace, solo cuándo el sistema lo
-- emite; ver DEC-054).
CREATE OR REPLACE FUNCTION retention_claim_due_customers(p_limit integer, p_now timestamptz)
RETURNS TABLE (customer_id uuid, barbershop_id uuid)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
BEGIN
  IF p_limit IS NULL OR p_limit < 1 OR p_limit > 200 THEN
    RAISE EXCEPTION 'retention_claim_due_customers: p_limit fuera de rango (1-200).';
  END IF;
  IF p_now IS NULL THEN
    RAISE EXCEPTION 'retention_claim_due_customers: p_now no admite NULL.';
  END IF;

  RETURN QUERY
  SELECT c.id, c.barbershop_id
  FROM public.customer c
  JOIN public.barbershop b ON b.id = c.barbershop_id
  WHERE c.anonymized_at IS NULL
    AND GREATEST(
          c.created_at,
          (SELECT max(a.created_at) FROM public.appointment a
             WHERE a.barbershop_id = c.barbershop_id AND a.customer_id = c.id),
          (SELECT max(ah.occurred_at) FROM public.appointment_history ah
             JOIN public.appointment a
               ON a.barbershop_id = ah.barbershop_id AND a.id = ah.appointment_id
             WHERE a.barbershop_id = c.barbershop_id AND a.customer_id = c.id),
          (SELECT max(t.issued_at) FROM public.appointment_access_token t
             JOIN public.appointment a
               ON a.barbershop_id = t.barbershop_id AND a.id = t.appointment_id
             WHERE a.barbershop_id = c.barbershop_id AND a.customer_id = c.id)
        ) < p_now - pg_catalog.make_interval(months => b.personal_data_retention_months)
  -- Orden por creación, no por la actividad calculada: evita repetir el
  -- GREATEST cuatro veces más solo para ordenar un lote que ya es pequeño
  -- (p_limit <= 200) y que se reintenta en cada ciclo del trabajador.
  ORDER BY c.created_at
  LIMIT p_limit
  FOR UPDATE OF c SKIP LOCKED;
END;
$$;

-- Función de worker (DEC-040, DDL-SEC-04): solo barberia_worker la ejecuta.
REVOKE ALL     ON FUNCTION retention_claim_due_customers(integer, timestamptz) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION retention_claim_due_customers(integer, timestamptz) TO barberia_worker;

COMMENT ON FUNCTION retention_claim_due_customers(integer, timestamptz) IS
  'Reclama clientes con datos personales vencidos por última actividad, no por creación '
  '(DEC-042). Devuelve SOLO identificadores; p_limit 1-200 (DDL-OPS-01). El trabajador anonimiza '
  'con customer_anonymize dentro de la MISMA transacción que este claim (DEC-053: sin lease, a '
  'diferencia de notification_claim_due, porque no hay E/S externa entre reclamar y escribir).';

-- ---------------------------------------------------------------------------
-- G.1 · `customer_anonymize` — DEC-025, DEC-049, DDL-PRI-01
-- ---------------------------------------------------------------------------
-- Cubre TODAS las copias de datos personales enumeradas por DEC-049, no solo
-- `customer`: una anonimización parcial deja identificable a la persona a
-- través del resto de copias de su nombre y de texto libre asociado a su
-- cita. Transaccional (una sola función = una sola transacción) e
-- idempotente: repetirla sobre un cliente ya anonimizado no hace nada
-- (guardado por `anonymized_at IS NULL` en el primer UPDATE, que además es
-- la señal de "hay algo que hacer" para el resto de la función).
CREATE OR REPLACE FUNCTION customer_anonymize(
  p_barbershop_id uuid,
  p_customer_id   uuid,
  p_now           timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  -- Marcador único y literal (customer_anonymized_ck lo exige tal cual):
  -- "¿está anonimizado?" es una comparación exacta, no una convención de
  -- texto libre que un futuro cambio podría romper en silencio.
  v_placeholder constant text := 'Cliente anonimizado';
  v_claimed     integer;
BEGIN
  IF p_barbershop_id IS NULL OR p_customer_id IS NULL OR p_now IS NULL THEN
    RAISE EXCEPTION 'customer_anonymize: ningún argumento admite NULL.';
  END IF;

  UPDATE public.customer
  SET full_name = v_placeholder,
      phone = NULL,
      email = NULL,
      anonymized_at = p_now
  WHERE barbershop_id = p_barbershop_id
    AND id = p_customer_id
    AND anonymized_at IS NULL;

  GET DIAGNOSTICS v_claimed = ROW_COUNT;
  IF v_claimed = 0 THEN
    RETURN false;  -- Ya anonimizado, o no existe: nada más que hacer.
  END IF;

  -- RN-RES-03: quien reserva y quien es atendido pueden diferir, pero
  -- DEC-049 cubre attendee_name igual que el nombre del cliente.
  UPDATE public.appointment
  SET attendee_name = v_placeholder,
      customer_note = NULL
  WHERE barbershop_id = p_barbershop_id
    AND customer_id = p_customer_id
    AND (attendee_name <> v_placeholder OR customer_note IS NOT NULL);

  -- El motivo de una transición puede contener texto libre personal
  -- ("canceló porque su hijo..."); se redacta, nunca se borra la fila
  -- (el historial es append-only e inmutable, RN-HIS-02).
  UPDATE public.appointment_history ah
  SET reason = v_placeholder
  FROM public.appointment a
  WHERE ah.barbershop_id = p_barbershop_id
    AND ah.appointment_id = a.id
    AND a.barbershop_id = p_barbershop_id
    AND a.customer_id = p_customer_id
    AND ah.reason IS NOT NULL
    AND ah.reason <> v_placeholder;

  -- Solo los campos que SÍ son datos personales (DEC-054): un cambio de
  -- `status` o `starts_at` no identifica a nadie y no se toca.
  UPDATE public.appointment_history_change ahc
  SET previous_value = CASE WHEN ahc.previous_value IS NOT NULL THEN v_placeholder END,
      new_value      = CASE WHEN ahc.new_value      IS NOT NULL THEN v_placeholder END
  FROM public.appointment_history ah
  JOIN public.appointment a
    ON a.barbershop_id = ah.barbershop_id AND a.id = ah.appointment_id
  WHERE ahc.barbershop_id = p_barbershop_id
    AND ahc.history_id = ah.id
    AND a.barbershop_id = p_barbershop_id
    AND a.customer_id = p_customer_id
    AND ahc.field_name IN ('attendee_name', 'customer_note')
    AND (ahc.previous_value IS DISTINCT FROM v_placeholder
         OR ahc.new_value IS DISTINCT FROM v_placeholder);

  -- El enlace público ya no debe abrir nada de un cliente anonimizado.
  UPDATE public.appointment_access_token t
  SET revoked_at = p_now
  FROM public.appointment a
  WHERE t.barbershop_id = p_barbershop_id
    AND t.appointment_id = a.id
    AND a.barbershop_id = p_barbershop_id
    AND a.customer_id = p_customer_id
    AND t.revoked_at IS NULL;

  -- idempotency_record.response_body NO se toca aquí a propósito (DEC-054):
  -- su TTL máximo (p_ttl_seconds <= 86400, un día) es siempre menor que el
  -- mínimo de retención posible (personal_data_retention_months >= 1 mes),
  -- así que cualquier respuesta idempotente ligada a la actividad de este
  -- cliente ya fue purgada por idempotency_purge_expired mucho antes de que
  -- el cliente llegara a ser candidato de esta función. No hay columna que
  -- correlacione idempotency_record con customer_id, y añadir una solo para
  -- este caso ya imposible sería alcance fuera de este issue.

  RETURN true;
END;
$$;

-- Función de worker (DEC-040, DDL-SEC-04): solo barberia_worker la ejecuta,
-- nunca barberia_app. El alcance de escritura de esta función es amplio
-- (varias tablas, cross-tenant en el sentido de que no depende de
-- app.barbershop_id), así que su superficie de privilegios es tan estrecha
-- como sea posible: ni EXECUTE público ni acceso directo del API.
REVOKE ALL     ON FUNCTION customer_anonymize(uuid, uuid, timestamptz) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION customer_anonymize(uuid, uuid, timestamptz) TO barberia_worker;

COMMENT ON FUNCTION customer_anonymize(uuid, uuid, timestamptz) IS
  'Anonimización completa de un cliente vencido (DEC-025, DEC-049, DDL-PRI-01): customer '
  '(nombre/teléfono/correo), appointment.attendee_name/customer_note, appointment_history.reason, '
  'appointment_history_change (solo campos personales) y revocación de appointment_access_token. '
  'Transaccional e idempotente. Se llama dentro de la misma transacción que '
  'retention_claim_due_customers reclamó la fila (DEC-053). No valida por sí misma que el cliente '
  'esté vencido: confía en que solo se invoca sobre lo que ese claim ya seleccionó.';


-- ===========================================================================
-- SECCIÓN H · Hallazgos que este diseño deja abiertos
-- ===========================================================================
--
-- No son decisiones que el DDL pueda tomar. Cada una necesita un `DEC-*` o una
-- corrección documental, conforme a AGENTS.md ("no inventar una respuesta a una
-- duda o contradicción: registrarla antes de codificar").
--
-- Resueltos desde la primera versión de este archivo (se conservan para no
-- perder el rastro, no porque sigan abiertos):
--   H.1 · Versión mínima de PostgreSQL — DEC-044 la fija en 14.
--   H.2 · Idioma de `appointment_history.event_type` — DEC-041/CT-002, ya
--         aplicado en la sección D.2 y en estados-citas.md §9.
--   H.3 · Anticipación del 2.º y 3.er recordatorio — DEC-048 (24h y 2h),
--         con el ordinal 1/30min sembrado por disparador (DDL-BIZ-01).
--   H.8 · Protocolo de lease de notification_claim_due, recuperación de
--         leases vencidos y validación de p_limit — DEC-053 (DDL-CON-01,
--         DDL-OPS-01). retention_claim_due_customers se revisa y se
--         mantiene sin lease por la misma decisión.
--   H.7 · Admin RLS policy apuntando a barberia_migrator en service,
--         barber_service y notification_channel_setting — corregido a
--         barberia_owner (DDL-RLS-01, issue #7), consistente con DEC-040.
--
-- H.4 · Validación de la zona IANA de la barbería
--       No es expresable en un `CHECK` (pg_timezone_names no es inmutable).
--       Queda en el servicio de configuración. Alternativa si se quisiera en
--       base de datos: tabla de referencia poblada por migración, con FK.
--
-- H.5 · Unicidad global del correo de acceso
--       `staff_user.email` es único en todo el sistema para poder resolver la
--       barbería antes de autenticar. Es una excepción consciente a la
--       unicidad tenant-aware de estandar-base-datos.md §6, aceptable mientras
--       no exista auto-registro. Si un usuario llegara a pertenecer a varias
--       barberías (previsto como caso futuro en RN-TEN-01), este diseño cambia.
--
-- H.6 · Bloqueos fuera de la restricción de exclusión
--       `time_block` NO comparte la restricción con `appointment`, por
--       exigencia de RN-BLQ-03 y DEC-008. La consecuencia —una cita puede
--       quedar dentro de un bloqueo— es deliberada (RN-CON-06) y el flujo
--       asistido que la resuelve es responsabilidad de la aplicación.
--       `docs/05-backend/concurrencia.md` sigue pendiente de creación.
--
-- H.9 · Índices parciales ensanchados a completos (DDL-PER-01, issue #7)
--       staff_session, staff_recovery_code (índice nuevo aparte, no
--       ensanchado: el parcial existente impone unicidad de negocio),
--       time_block_series, time_block, appointment_access_token y
--       notification_schedule (ídem staff_recovery_code) ganaron cobertura
--       para consultas que también necesitan filas históricas/revocadas/
--       eliminadas lógicamente, sin duplicar índices sobre el mismo
--       prefijo. `appointment` (por service_id) y `appointment_history`
--       (por actor_staff_user_id) ganaron un índice que no existía.
-- ===========================================================================
