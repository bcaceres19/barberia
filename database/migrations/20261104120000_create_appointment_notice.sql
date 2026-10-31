-- Propósito
--   Issue #463 (HU-130, DEC-142): cola del aviso por correo al cliente cuando el
--   barbero cancela o reprograma su turno (F-NOT-02, primer tramo de B5), y la
--   evidencia de cada intento de envío (RN-REC-04).
--
-- Reglas y decisiones
--   RN-REC-01 (un cambio esencial notifica sin duplicar), RN-REC-02 y RN-REC-03
--   (sin avisos de citas canceladas ni de horarios viejos: el mensaje se redacta
--   al enviar), RN-REC-04 (evidencia sin datos de contacto), RN-REC-06 (se
--   persiste en la transacción del cambio), RN-DAT-02, RN-TEN-01, DEC-024, DEC-035
--   a DEC-040 (worker sin tenant, solo funciones SECURITY DEFINER), DEC-102
--   (patrón de cola con lease) y DEC-142.
--
-- Modelo
--   `appointment_notice (id, barbershop_id, appointment_id, kind, status,
--   state_at, attempts, run_at, claim_token, lease_expires_at, outcome_code,
--   covers_state_at, finished_at, created_at, updated_at)`: un aviso por cambio.
--   * Como máximo un aviso `pending` por turno (índice único parcial): un cambio
--     nuevo reemplaza al pendiente y reinicia la espera, de modo que dos
--     reprogramaciones seguidas dan un solo correo.
--   * `state_at` es el `updated_at` del turno cuando se encoló; `covers_state_at`
--     es el que tenía cuando el correo salió. Si un aviso ya envió un estado igual
--     o más reciente, el que esperaba se descarta (`superseded`).
--   * Los avisos terminados (`sent`, `skipped`, `failed`) se conservan: son la
--     evidencia. Su purga corresponde a B6 (RN-DAT-03).
--   `appointment_notice_attempt (id, barbershop_id, notice_id, attempt_no, channel,
--   outcome, error_code, created_at)`: una fila por intento, sin correo, teléfono
--   ni nombre. La escribe únicamente `notice_finish`.
--
-- Privilegios y RLS
--   RLS forzada en ambas. La aplicación encola (INSERT de cinco columnas y UPDATE
--   de tres, para reemplazar el pendiente) y lee; no escribe estado, intentos ni
--   resultados. El worker no toca tablas: opera con `notice_claim`,
--   `notice_context` y `notice_finish`, exclusivas de `barberia_worker`.
--
-- Qué NO hace esta migración
--   No programa recordatorios (F-NOT-03), no configura canales por evento
--   (RN-REC-05), no cubre WhatsApp (DEC-027) y no purga la evidencia.
--
-- Clasificación de datos
--   Ninguna columna guarda correo, teléfono ni nombre. `notice_context` devuelve
--   el correo y el nombre del cliente al worker únicamente para redactar y enviar
--   el mensaje de un aviso con su `claim_token` vigente.
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE, ver
--   20260811145252_harden_roles_and_definer_functions.sql).
--
-- Plan de avance
--   Una corrección posterior se hace con otra migración, nunca editando esta
--   (DEC-036). No existe archivo `down`: el mecanismo normal es roll-forward.

-- ---------------------------------------------------------------------------
-- 1. Precondiciones
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'appointment'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260827110000_create_appointment_core.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'appointment_notice'
  ) THEN
    RAISE EXCEPTION 'appointment_notice ya existe: esta migración no puede aplicarse dos veces.';
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_owner')
     OR NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_worker') THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260811145252_harden_roles_and_definer_functions.sql '
      '(roles barberia_owner y barberia_worker); esa migración debe estar aplicada antes.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Cola de avisos
-- ---------------------------------------------------------------------------

CREATE TABLE appointment_notice (
  id               uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id    uuid        NOT NULL,
  appointment_id   uuid        NOT NULL,
  kind             text        NOT NULL,
  status           text        NOT NULL DEFAULT 'pending',
  state_at         timestamptz NOT NULL,
  attempts         integer     NOT NULL DEFAULT 0,
  run_at           timestamptz NOT NULL DEFAULT now() + interval '60 seconds',
  claim_token      uuid,
  lease_expires_at timestamptz,
  outcome_code     text,
  covers_state_at  timestamptz,
  finished_at      timestamptz,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT appointment_notice_id_pk PRIMARY KEY (id),
  CONSTRAINT appointment_notice_barbershop_id_id_uk UNIQUE (barbershop_id, id),

  -- Sin CASCADE: el turno no se borra (RN-HIS-02) y la evidencia lo sobrevive.
  CONSTRAINT appointment_notice_appointment_fk FOREIGN KEY (barbershop_id, appointment_id)
    REFERENCES appointment (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT appointment_notice_kind_ck CHECK (kind IN ('cancelled_by_barber', 'rescheduled')),
  CONSTRAINT appointment_notice_status_ck
    CHECK (status IN ('pending', 'processing', 'sent', 'skipped', 'failed')),
  CONSTRAINT appointment_notice_attempts_ck CHECK (attempts >= 0),
  CONSTRAINT appointment_notice_outcome_code_ck
    CHECK (outcome_code IS NULL OR outcome_code ~ '^[a-z0-9_]{1,40}$'),

  -- Un aviso en proceso siempre tiene su reclamación (DDL-CON-01).
  CONSTRAINT appointment_notice_claim_ck CHECK (
    (status = 'processing' AND claim_token IS NOT NULL AND lease_expires_at IS NOT NULL)
    OR (status <> 'processing' AND claim_token IS NULL AND lease_expires_at IS NULL)
  ),
  -- Terminado ⇔ tiene fecha de cierre; un enviado declara qué estado cubrió.
  CONSTRAINT appointment_notice_finished_ck CHECK (
    (status IN ('sent', 'skipped', 'failed')) = (finished_at IS NOT NULL)
  ),
  CONSTRAINT appointment_notice_covers_ck CHECK (
    (status = 'sent') = (covers_state_at IS NOT NULL)
  )
);

COMMENT ON TABLE appointment_notice IS
  'Aviso por correo al cliente cuando el barbero cancela o reprograma su turno (HU-130, DEC-142). '
  'Se escribe en la misma transacción que el cambio y lo consume el worker con lease (SKIP LOCKED) y '
  'cierre por CAS sobre claim_token. Los terminados se conservan como evidencia (RN-REC-04). Sin '
  'correo, teléfono ni nombre; propietario funcional: barbería. Retención: la fija B6 (RN-DAT-03).';

COMMENT ON COLUMN appointment_notice.state_at IS
  'updated_at del turno cuando se encoló el aviso. Sirve para descartar un aviso cuando otro ya envió '
  'un estado igual o más reciente (covers_state_at >= state_at).';

COMMENT ON COLUMN appointment_notice.claim_token IS
  'Reclamación vigente. Solo quien la tiene puede cerrar el aviso; si el lease vence, otro worker lo '
  'reclama con un token nuevo y el primero ya no puede cerrarlo.';

-- Como máximo UN aviso pendiente por turno (RN-REC-01: sin duplicados).
CREATE UNIQUE INDEX appointment_notice_pending_uk
  ON appointment_notice (appointment_id) WHERE status = 'pending';

CREATE INDEX idx_appointment_notice_due
  ON appointment_notice (run_at) WHERE status = 'pending';
CREATE INDEX idx_appointment_notice_lease
  ON appointment_notice (lease_expires_at) WHERE status = 'processing';
CREATE INDEX idx_appointment_notice_appointment
  ON appointment_notice (barbershop_id, appointment_id, status);

CREATE TRIGGER appointment_notice_set_updated_at
  BEFORE UPDATE ON appointment_notice
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE appointment_notice ENABLE ROW LEVEL SECURITY;
ALTER TABLE appointment_notice FORCE  ROW LEVEL SECURITY;

CREATE POLICY appointment_notice_all_admin_policy ON appointment_notice
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY appointment_notice_select_tenant_policy ON appointment_notice
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY appointment_notice_insert_tenant_policy ON appointment_notice
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY appointment_notice_update_tenant_policy ON appointment_notice
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- La aplicación solo encola: no fija estado, resultado ni reclamación.
GRANT SELECT ON TABLE appointment_notice TO barberia_app;
GRANT INSERT (barbershop_id, appointment_id, kind, state_at, run_at)
  ON TABLE appointment_notice TO barberia_app;
GRANT UPDATE (kind, state_at, run_at) ON TABLE appointment_notice TO barberia_app;

-- ---------------------------------------------------------------------------
-- 3. Evidencia de cada intento
-- ---------------------------------------------------------------------------

CREATE TABLE appointment_notice_attempt (
  id            uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id uuid        NOT NULL,
  notice_id     uuid        NOT NULL,
  attempt_no    integer     NOT NULL,
  channel       text        NOT NULL,
  outcome       text        NOT NULL,
  error_code    text,
  created_at    timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT appointment_notice_attempt_id_pk PRIMARY KEY (id),
  CONSTRAINT appointment_notice_attempt_notice_fk FOREIGN KEY (barbershop_id, notice_id)
    REFERENCES appointment_notice (barbershop_id, id) ON DELETE RESTRICT,
  CONSTRAINT appointment_notice_attempt_no_uk UNIQUE (notice_id, attempt_no),

  CONSTRAINT appointment_notice_attempt_no_ck CHECK (attempt_no >= 1),
  CONSTRAINT appointment_notice_attempt_channel_ck CHECK (channel IN ('email')),
  CONSTRAINT appointment_notice_attempt_outcome_ck
    CHECK (outcome IN ('sent', 'retry', 'failed', 'skipped')),
  CONSTRAINT appointment_notice_attempt_error_code_ck
    CHECK (error_code IS NULL OR error_code ~ '^[a-z0-9_]{1,40}$')
);

COMMENT ON TABLE appointment_notice_attempt IS
  'Evidencia técnica de cada intento de envío de un aviso (RN-REC-04), separada del historial de '
  'negocio del turno. Sin correo, teléfono ni nombre (RN-DAT-02). La escribe solo notice_finish.';

CREATE INDEX idx_appointment_notice_attempt_notice
  ON appointment_notice_attempt (barbershop_id, notice_id);

ALTER TABLE appointment_notice_attempt ENABLE ROW LEVEL SECURITY;
ALTER TABLE appointment_notice_attempt FORCE  ROW LEVEL SECURITY;

CREATE POLICY appointment_notice_attempt_all_admin_policy ON appointment_notice_attempt
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY appointment_notice_attempt_select_tenant_policy ON appointment_notice_attempt
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT ON TABLE appointment_notice_attempt TO barberia_app;

-- ---------------------------------------------------------------------------
-- 4. Funciones del worker (SECURITY DEFINER, sin contexto de tenant)
-- ---------------------------------------------------------------------------

-- 4.1 Reclamo con lease. Primero cierra como `failed` los avisos cuyo lease venció
-- habiendo agotado los intentos (un worker que cae siempre al enviar no puede
-- reclamar el mismo aviso para siempre); luego reclama en un único UPDATE
-- atómico sobre un SELECT ... FOR UPDATE SKIP LOCKED.
CREATE FUNCTION notice_claim(
  p_limit         integer,
  p_lease_seconds integer,
  p_max_attempts  integer,
  p_now           timestamptz DEFAULT pg_catalog.now()
)
RETURNS TABLE (
  notice_id   uuid,
  claim_token uuid,
  attempts    integer
)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
#variable_conflict use_column
BEGIN
  IF p_limit IS NULL OR p_limit < 1 OR p_limit > 100 THEN
    RAISE EXCEPTION 'notice_claim: p_limit fuera de rango (1-100).';
  END IF;
  IF p_lease_seconds IS NULL OR p_lease_seconds < 10 OR p_lease_seconds > 3600 THEN
    RAISE EXCEPTION 'notice_claim: p_lease_seconds fuera de rango (10-3600).';
  END IF;
  IF p_max_attempts IS NULL OR p_max_attempts < 1 OR p_max_attempts > 20 THEN
    RAISE EXCEPTION 'notice_claim: p_max_attempts fuera de rango (1-20).';
  END IF;

  UPDATE public.appointment_notice n
     SET status = 'failed', claim_token = NULL, lease_expires_at = NULL,
         outcome_code = 'lease_expired', finished_at = p_now
   WHERE n.status = 'processing' AND n.lease_expires_at <= p_now AND n.attempts >= p_max_attempts;

  RETURN QUERY
  WITH due AS (
    SELECT n.id
      FROM public.appointment_notice n
     WHERE (n.status = 'pending' AND n.run_at <= p_now)
        OR (n.status = 'processing' AND n.lease_expires_at <= p_now)
     ORDER BY n.run_at, n.created_at
     LIMIT p_limit
       FOR UPDATE OF n SKIP LOCKED
  )
  UPDATE public.appointment_notice n
     SET status = 'processing',
         claim_token = pg_catalog.gen_random_uuid(),
         lease_expires_at = p_now + pg_catalog.make_interval(secs => p_lease_seconds),
         attempts = n.attempts + 1,
         updated_at = p_now
    FROM due
   WHERE n.id = due.id
  RETURNING n.id, n.claim_token, n.attempts;
END;
$$;

REVOKE ALL     ON FUNCTION notice_claim(integer, integer, integer, timestamptz) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION notice_claim(integer, integer, integer, timestamptz) TO barberia_worker;

COMMENT ON FUNCTION notice_claim(integer, integer, integer, timestamptz) IS
  'Reclamo con lease de la cola de avisos (DDL-CON-01): SKIP LOCKED sobre avisos pending vencidos o '
  'processing con lease vencido; cierra como failed los que ya agotaron los intentos. Exclusivo de '
  'barberia_worker.';

-- 4.2 Contexto de UN aviso reclamado: lo mínimo para redactar el correo con el
-- estado VIGENTE del turno (RN-REC-03), nunca el que tenía al encolarse.
CREATE FUNCTION notice_context(p_notice_id uuid, p_claim_token uuid)
RETURNS TABLE (
  kind                 text,
  appointment_status   text,
  starts_at            timestamptz,
  appointment_state_at timestamptz,
  customer_email       text,
  attendee_name        text,
  service_name         text,
  barbershop_name      text,
  barbershop_timezone  text,
  already_covered      boolean
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = ''
AS $$
  SELECT n.kind, a.status, a.starts_at, a.updated_at,
         cu.email, a.attendee_name, a.service_name_snapshot, b.name, b.timezone,
         EXISTS (
           SELECT 1 FROM public.appointment_notice o
            WHERE o.barbershop_id = n.barbershop_id AND o.appointment_id = n.appointment_id
              AND o.id <> n.id AND o.status = 'sent' AND o.covers_state_at >= n.state_at
         )
    FROM public.appointment_notice n
    JOIN public.appointment a ON a.barbershop_id = n.barbershop_id AND a.id = n.appointment_id
    JOIN public.customer cu ON cu.barbershop_id = a.barbershop_id AND cu.id = a.customer_id
    JOIN public.barbershop b ON b.id = n.barbershop_id
   WHERE n.id = p_notice_id AND n.claim_token = p_claim_token AND n.status = 'processing';
$$;

REVOKE ALL     ON FUNCTION notice_context(uuid, uuid) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION notice_context(uuid, uuid) TO barberia_worker;

COMMENT ON FUNCTION notice_context(uuid, uuid) IS
  'Datos mínimos para redactar y enviar UN aviso reclamado; exige su claim_token vigente, así que no '
  'sirve para leer citas arbitrarias. Incluye el correo y el nombre del cliente solo para ese envío '
  '(RN-DAT-02). Exclusivo de barberia_worker.';

-- 4.3 Cierre por CAS sobre claim_token. Registra el intento en la misma sentencia
-- lógica que cambia el estado, de modo que ningún resultado queda sin evidencia.
CREATE FUNCTION notice_finish(
  p_notice_id       uuid,
  p_claim_token     uuid,
  p_outcome         text,
  p_code            text DEFAULT NULL,
  p_retry_seconds   integer DEFAULT NULL,
  p_max_attempts    integer DEFAULT 5,
  p_covers_state_at timestamptz DEFAULT NULL,
  p_now             timestamptz DEFAULT pg_catalog.now()
)
RETURNS boolean
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_notice  public.appointment_notice%ROWTYPE;
  v_attempt text;
BEGIN
  IF p_outcome NOT IN ('sent', 'skipped', 'retry', 'permanent') THEN
    RAISE EXCEPTION 'notice_finish: resultado inválido %.', p_outcome;
  END IF;

  SELECT * INTO v_notice
    FROM public.appointment_notice
   WHERE id = p_notice_id AND claim_token = p_claim_token AND status = 'processing'
   FOR UPDATE;
  IF NOT FOUND THEN
    RETURN false;
  END IF;

  IF p_outcome = 'sent' THEN
    IF p_covers_state_at IS NULL THEN
      RAISE EXCEPTION 'notice_finish: sent exige p_covers_state_at.';
    END IF;
    v_attempt := 'sent';
    UPDATE public.appointment_notice
       SET status = 'sent', claim_token = NULL, lease_expires_at = NULL,
           outcome_code = NULL, covers_state_at = p_covers_state_at, finished_at = p_now
     WHERE id = v_notice.id;

  ELSIF p_outcome = 'skipped' THEN
    v_attempt := 'skipped';
    UPDATE public.appointment_notice
       SET status = 'skipped', claim_token = NULL, lease_expires_at = NULL,
           outcome_code = p_code, finished_at = p_now
     WHERE id = v_notice.id;

  ELSIF p_outcome = 'retry' AND v_notice.attempts < COALESCE(p_max_attempts, 5) THEN
    v_attempt := 'retry';
    UPDATE public.appointment_notice
       SET status = 'pending', claim_token = NULL, lease_expires_at = NULL,
           run_at = p_now + pg_catalog.make_interval(secs => COALESCE(p_retry_seconds, 60)),
           outcome_code = p_code
     WHERE id = v_notice.id;

  ELSE  -- 'permanent', o 'retry' con los intentos agotados
    v_attempt := 'failed';
    UPDATE public.appointment_notice
       SET status = 'failed', claim_token = NULL, lease_expires_at = NULL,
           outcome_code = p_code, finished_at = p_now
     WHERE id = v_notice.id;
  END IF;

  INSERT INTO public.appointment_notice_attempt
         (barbershop_id, notice_id, attempt_no, channel, outcome, error_code, created_at)
  VALUES (v_notice.barbershop_id, v_notice.id, v_notice.attempts, 'email', v_attempt,
          CASE WHEN v_attempt = 'sent' THEN NULL ELSE p_code END, p_now);

  RETURN true;
END;
$$;

REVOKE ALL     ON FUNCTION notice_finish(uuid, uuid, text, text, integer, integer, timestamptz, timestamptz) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION notice_finish(uuid, uuid, text, text, integer, integer, timestamptz, timestamptz) TO barberia_worker;

COMMENT ON FUNCTION notice_finish(uuid, uuid, text, text, integer, integer, timestamptz, timestamptz) IS
  'Cierre por CAS sobre claim_token (DDL-CON-01): sent, skipped, retry (vuelve a pending con espera, o '
  'failed si agotó los intentos) o permanent. Inserta el intento en appointment_notice_attempt '
  '(RN-REC-04). Exclusivo de barberia_worker.';
