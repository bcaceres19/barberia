-- Propósito
--   Issue #355 (DEC-126): almacena lo que Meta notifica por webhook. Crea
--   `whatsapp_conversation_window` (última interacción entrante por teléfono, de
--   la que depende la ventana de atención de 24 horas del texto libre) y
--   `whatsapp_message_status` (estados de entrega de los mensajes que NAVA envía,
--   por `wamid`), con las funciones SECURITY DEFINER con las que la API escribe y
--   consulta, y la de purga del worker.
--
-- Reglas y decisiones
--   RN-DAT-02 (nada sensible en logs ni en claro), DEC-024 (RLS forzada en toda
--   tabla; estas no tienen `barbershop_id`, así que solo llevan la política
--   administrativa de `barberia_owner`, igual que `auth_phone_challenge`), DEC-027
--   (WhatsApp oficial), DEC-035/DEC-036 (estándar y Atlas), DEC-040 (roles
--   separados), DDL-AUT-01 (sin GRANT directo a `barberia_app`: solo funciones
--   estrechas), DEC-123/DEC-124 (OTP por Meta, modo texto), DEC-126.
--
-- Por qué no llevan barbershop_id
--   El número emisor es único para toda la plataforma y un mismo teléfono puede
--   existir en varias barberías: la ventana pertenece a la conversación entre ese
--   teléfono y el número de NAVA, no a un tenant. La tabla guarda únicamente el
--   HMAC del teléfono, de modo que no revela a qué barbería ni a qué persona
--   corresponde.
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda `barberia_owner`; la política
--   administrativa se escribe `FOR ALL TO barberia_owner` (DEC-040). Las cuatro funciones son SECURITY DEFINER
--   con `search_path=''` y nombres calificados. `barberia_app` ejecuta las de
--   escritura/lectura; solo `barberia_worker` ejecuta la purga. Ninguna tabla se
--   concede a ningún rol.
--
-- Plan de avance
--   Una corrección posterior se hace con otra migración (DEC-036). Sin archivo
--   `down`: el mecanismo normal es roll-forward.

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_owner')
     OR NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_app')
     OR NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_worker') THEN
    RAISE EXCEPTION
      'Esta migración depende de los roles barberia_owner, barberia_app y barberia_worker '
      '(20260811145252_harden_roles_and_definer_functions.sql).';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 1. whatsapp_conversation_window
-- ---------------------------------------------------------------------------

CREATE TABLE whatsapp_conversation_window (
  phone_hash      text        NOT NULL,
  last_inbound_at timestamptz NOT NULL,
  expires_at      timestamptz NOT NULL,

  CONSTRAINT whatsapp_conversation_window_pk       PRIMARY KEY (phone_hash),
  CONSTRAINT whatsapp_conversation_window_hash_ck  CHECK (char_length(phone_hash) = 64),
  CONSTRAINT whatsapp_conversation_window_range_ck CHECK (expires_at > last_inbound_at)
);

COMMENT ON TABLE whatsapp_conversation_window IS
  'Última interacción entrante de cada teléfono con el número de WhatsApp de NAVA (DEC-126). '
  'Propietario funcional: plataforma. Retención: hasta expires_at (24 h después del último '
  'mensaje entrante, el límite de Meta para texto libre). Clasificación: dato personal '
  'seudonimizado; phone_hash es HMAC-SHA256 con el secreto de despliegue, nunca el teléfono. '
  'Sin barbershop_id por diseño: el número emisor es único de la plataforma. RLS forzada '
  'con solo la política administrativa de barberia_owner.';

CREATE INDEX idx_whatsapp_conversation_window_expires_at
  ON whatsapp_conversation_window (expires_at);

ALTER TABLE whatsapp_conversation_window ENABLE ROW LEVEL SECURITY;
ALTER TABLE whatsapp_conversation_window FORCE  ROW LEVEL SECURITY;

CREATE POLICY whatsapp_conversation_window_all_admin_policy ON whatsapp_conversation_window
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

-- ---------------------------------------------------------------------------
-- 2. whatsapp_message_status
-- ---------------------------------------------------------------------------

CREATE TABLE whatsapp_message_status (
  wamid       text        NOT NULL,
  status      text        NOT NULL,
  occurred_at timestamptz NOT NULL,
  error_code  integer,
  received_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT whatsapp_message_status_pk        PRIMARY KEY (wamid, status),
  CONSTRAINT whatsapp_message_status_wamid_ck  CHECK (char_length(wamid) BETWEEN 1 AND 255),
  CONSTRAINT whatsapp_message_status_status_ck CHECK (status IN ('sent', 'delivered', 'read', 'failed'))
);

COMMENT ON TABLE whatsapp_message_status IS
  'Estados de entrega que Meta notifica de los mensajes enviados por NAVA (DEC-126). '
  'Propietario funcional: plataforma. Retención: 30 días desde received_at. Clasificación: '
  'técnico. Solo guarda el identificador del mensaje (wamid), el estado y el código de error '
  'de Meta: ni teléfono ni contenido. La clave (wamid, status) hace idempotente la repetición '
  'de una notificación. Sin barbershop_id por diseño; RLS forzada con solo la política '
  'administrativa de barberia_owner.';

CREATE INDEX idx_whatsapp_message_status_received_at
  ON whatsapp_message_status (received_at);

ALTER TABLE whatsapp_message_status ENABLE ROW LEVEL SECURITY;
ALTER TABLE whatsapp_message_status FORCE  ROW LEVEL SECURITY;

CREATE POLICY whatsapp_message_status_all_admin_policy ON whatsapp_message_status
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

-- DDL-AUT-01: ningún GRANT sobre las tablas ni política para barberia_app o
-- barberia_worker. Todo acceso es por las funciones.

-- ---------------------------------------------------------------------------
-- 3. Funciones
-- ---------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION whatsapp_window_record(p_phone_hash text, p_at timestamptz)
RETURNS void
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_now timestamptz := pg_catalog.now();
  v_at  timestamptz;
BEGIN
  IF p_phone_hash IS NULL OR p_at IS NULL OR pg_catalog.char_length(p_phone_hash) <> 64 THEN
    RAISE EXCEPTION 'whatsapp_window_record: argumentos inválidos.';
  END IF;
  -- Un reloj adelantado del emisor no puede abrir una ventana más allá de ahora.
  v_at := LEAST(p_at, v_now);

  INSERT INTO public.whatsapp_conversation_window (phone_hash, last_inbound_at, expires_at)
  VALUES (p_phone_hash, v_at, v_at + pg_catalog.make_interval(hours => 24))
  ON CONFLICT (phone_hash) DO UPDATE
  SET last_inbound_at = GREATEST(public.whatsapp_conversation_window.last_inbound_at, EXCLUDED.last_inbound_at),
      expires_at      = GREATEST(public.whatsapp_conversation_window.last_inbound_at, EXCLUDED.last_inbound_at)
                        + pg_catalog.make_interval(hours => 24);
END;
$$;

REVOKE ALL     ON FUNCTION whatsapp_window_record(text, timestamptz) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION whatsapp_window_record(text, timestamptz) TO barberia_app;

COMMENT ON FUNCTION whatsapp_window_record(text, timestamptz) IS
  'Registra un mensaje entrante: conserva el más reciente por phone_hash, de forma que una '
  'notificación repetida o tardía no retrocede la ventana. Se aplica un tope a "ahora".';

CREATE OR REPLACE FUNCTION whatsapp_window_is_open(p_phone_hash text)
RETURNS boolean
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = ''
AS $$
BEGIN
  IF p_phone_hash IS NULL OR pg_catalog.char_length(p_phone_hash) <> 64 THEN
    RAISE EXCEPTION 'whatsapp_window_is_open: argumentos inválidos.';
  END IF;
  RETURN EXISTS (
    SELECT 1 FROM public.whatsapp_conversation_window w
    WHERE w.phone_hash = p_phone_hash AND w.expires_at > pg_catalog.now()
  );
END;
$$;

REVOKE ALL     ON FUNCTION whatsapp_window_is_open(text) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION whatsapp_window_is_open(text) TO barberia_app;

COMMENT ON FUNCTION whatsapp_window_is_open(text) IS
  'true solo si hay un mensaje entrante registrado en las últimas 24 horas. false significa '
  '"sin registro vigente", no "Meta lo rechazará": un webhook perdido no cierra la ventana '
  'real, por eso ningún envío se decide con este valor (DEC-126).';

CREATE OR REPLACE FUNCTION whatsapp_status_record(
  p_wamid       text,
  p_status      text,
  p_occurred_at timestamptz,
  p_error_code  integer
)
RETURNS boolean
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_inserted integer;
BEGIN
  IF p_wamid IS NULL OR p_status IS NULL OR p_occurred_at IS NULL
     OR pg_catalog.char_length(p_wamid) NOT BETWEEN 1 AND 255
     OR p_status NOT IN ('sent', 'delivered', 'read', 'failed') THEN
    RAISE EXCEPTION 'whatsapp_status_record: argumentos inválidos.';
  END IF;

  INSERT INTO public.whatsapp_message_status (wamid, status, occurred_at, error_code)
  VALUES (p_wamid, p_status, p_occurred_at, p_error_code)
  ON CONFLICT (wamid, status) DO NOTHING;

  GET DIAGNOSTICS v_inserted = ROW_COUNT;
  RETURN v_inserted = 1;
END;
$$;

REVOKE ALL     ON FUNCTION whatsapp_status_record(text, text, timestamptz, integer) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION whatsapp_status_record(text, text, timestamptz, integer) TO barberia_app;

COMMENT ON FUNCTION whatsapp_status_record(text, text, timestamptz, integer) IS
  'Registra un estado de entrega. Devuelve true si es nuevo y false si ya estaba registrado '
  '(Meta repite notificaciones): el llamador no vuelve a reaccionar al duplicado.';

CREATE OR REPLACE FUNCTION whatsapp_webhook_purge_expired(p_limit integer, p_status_retention_seconds integer)
RETURNS integer
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_windows  integer;
  v_statuses integer;
BEGIN
  IF p_limit IS NULL OR p_limit < 1 OR p_limit > 1000
     OR p_status_retention_seconds IS NULL OR p_status_retention_seconds < 60 THEN
    RAISE EXCEPTION 'whatsapp_webhook_purge_expired: parámetros fuera de rango.';
  END IF;

  WITH due AS (
    SELECT phone_hash FROM public.whatsapp_conversation_window
    WHERE expires_at <= pg_catalog.now()
    ORDER BY expires_at
    LIMIT p_limit
    FOR UPDATE SKIP LOCKED
  )
  DELETE FROM public.whatsapp_conversation_window
  WHERE phone_hash IN (SELECT phone_hash FROM due);
  GET DIAGNOSTICS v_windows = ROW_COUNT;

  WITH due AS (
    SELECT wamid, status FROM public.whatsapp_message_status
    WHERE received_at <= pg_catalog.now() - pg_catalog.make_interval(secs => p_status_retention_seconds)
    ORDER BY received_at
    LIMIT p_limit
    FOR UPDATE SKIP LOCKED
  )
  DELETE FROM public.whatsapp_message_status s
  USING due
  WHERE s.wamid = due.wamid AND s.status = due.status;
  GET DIAGNOSTICS v_statuses = ROW_COUNT;

  RETURN v_windows + v_statuses;
END;
$$;

REVOKE ALL     ON FUNCTION whatsapp_webhook_purge_expired(integer, integer) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION whatsapp_webhook_purge_expired(integer, integer) TO barberia_worker;

COMMENT ON FUNCTION whatsapp_webhook_purge_expired(integer, integer) IS
  'Purga por lote ventanas vencidas y estados más antiguos que la retención. Exclusiva del worker.';
