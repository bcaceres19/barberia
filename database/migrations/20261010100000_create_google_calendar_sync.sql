-- Propósito
--   Issue #324 (DEC-099, DEC-101, DEC-102, DEC-122): cola propia y vínculos de
--   evento para publicar las citas y los bloqueos de NAVA en el Google Calendar
--   del barbero. Crea `google_calendar_sync_job` (outbox con lease), `google_calendar_event_link`
--   (qué evento de Google corresponde a qué cita o bloqueo) y las funciones con
--   las que el worker reclama y finaliza trabajos SIN contexto de tenant.
--
-- Reglas y decisiones
--   RN-TEN-01, DEC-019/DEC-024 (RLS), DEC-035/DEC-036 (estándar y Atlas), DEC-040
--   (el worker solo ejecuta funciones, no tiene acceso directo a tablas),
--   DDL-CON-01 (reclamo con lease y finalización por CAS sobre `claim_token`),
--   DEC-099 (NAVA → Google), DEC-101 (semántica de la publicación), DEC-102
--   (cola propia escrita en la misma transacción del cambio de negocio).
--
-- Cola a nivel de ESTADO, no de operación
--   Un trabajo dice «reconcilia este recurso», no «crea» o «cancela». Al
--   ejecutarlo el worker lee el estado ACTUAL de la cita o del bloqueo y deja a
--   Google igual: varios cambios seguidos se funden en un trabajo `pending` (índice
--   único parcial) y un reintento o una ejecución fuera de orden nunca publica un
--   estado viejo. El vínculo persistido decide entre crear y actualizar, así que
--   reintentar jamás duplica un evento.
--
-- Acceso del worker (DEC-040)
--   `barberia_worker` no recibe SELECT/DML sobre ninguna tabla. Solo EXECUTE sobre
--   funciones SECURITY DEFINER acotadas: reclamar trabajos vencidos (SKIP LOCKED
--   con lease), leer el contexto de UN trabajo que reclamó (exige su `claim_token`),
--   finalizarlo (CAS sobre `claim_token`) y las tres funciones del chequeo periódico
--   de eventos borrados. Ninguna función acepta un tenant del cliente.
--
-- Qué NO hace esta migración
--   No publica series de bloqueo (DP-INT-02), no lee cambios de Google y no agrega
--   DELETE físico sobre `appointment` ni `time_block`. Los vínculos solo los
--   escribe el worker (las funciones); la aplicación únicamente los lee.
--
-- Clasificación de datos
--   El trabajo y el vínculo solo guardan identificadores opacos y códigos de
--   error cortos, sin datos personales. `gcal_job_context` devuelve, solo para un
--   trabajo reclamado, los campos mínimos para armar el evento (nombre y servicio de
--   la cita y, por DEC-122, el correo del cliente para invitarlo).
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator` (hereda `barberia_owner`).
--
-- Plan de avance
--   Una corrección posterior se hace con otra migración (DEC-036). Sin `down`.

-- ---------------------------------------------------------------------------
-- 1. Precondiciones
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'google_calendar_connection'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20261009130000_create_google_calendar_connection.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public'
      AND table_name IN ('google_calendar_sync_job', 'google_calendar_event_link')
  ) THEN
    RAISE EXCEPTION 'Las tablas de sincronización ya existen: esta migración no puede aplicarse dos veces.';
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_worker') THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260811145252_harden_roles_and_definer_functions.sql '
      '(rol barberia_worker); esa migración debe estar aplicada antes.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Chequeo periódico de eventos borrados: marca por conexión
-- ---------------------------------------------------------------------------

ALTER TABLE google_calendar_connection
  ADD COLUMN last_restore_check_at timestamptz;

COMMENT ON COLUMN google_calendar_connection.last_restore_check_at IS
  'Última vez que el worker comprobó que los eventos publicados siguen existiendo en Google '
  '(DEC-101). Sirve de lease implícito: la conexión se reserva al elegirla, no al terminar.';

-- ---------------------------------------------------------------------------
-- 3. Vínculo evento ↔ recurso
-- ---------------------------------------------------------------------------

CREATE TABLE google_calendar_event_link (
  connection_id    uuid        NOT NULL,
  resource_type    text        NOT NULL,
  resource_id      uuid        NOT NULL,
  barbershop_id    uuid        NOT NULL,
  google_event_id  text        NOT NULL,
  etag             text,
  event_generation integer     NOT NULL DEFAULT 1,
  visible_hash     text,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),

  -- Un evento como máximo por recurso y conexión.
  CONSTRAINT google_calendar_event_link_pk PRIMARY KEY (connection_id, resource_type, resource_id),

  CONSTRAINT google_calendar_event_link_connection_fk FOREIGN KEY (barbershop_id, connection_id)
    REFERENCES google_calendar_connection (barbershop_id, id) ON DELETE RESTRICT,

  -- El mismo evento de Google no puede representar dos recursos.
  CONSTRAINT google_calendar_event_link_event_uk UNIQUE (connection_id, google_event_id),

  CONSTRAINT google_calendar_event_link_resource_type_ck
    CHECK (resource_type IN ('appointment', 'time_block')),
  CONSTRAINT google_calendar_event_link_generation_ck CHECK (event_generation >= 1),
  CONSTRAINT google_calendar_event_link_event_id_ck
    CHECK (char_length(google_event_id) BETWEEN 5 AND 1024)
);

COMMENT ON TABLE google_calendar_event_link IS
  'Qué evento de Google Calendar corresponde a qué cita o bloqueo de NAVA (issue #324, DEC-101). '
  'Nunca se identifica por título ni por hora. Solo la escribe el worker mediante '
  'gcal_finish_job; la aplicación únicamente la lee. Cancelar en la app borra el vínculo para que '
  'el chequeo periódico no recree el evento. Sin datos personales.';

COMMENT ON COLUMN google_calendar_event_link.event_generation IS
  'Generación del identificador del evento. Google no permite reutilizar el id de un evento '
  'borrado, así que recrear un evento (por ejemplo, tras borrarlo el barbero) usa la generación '
  'siguiente. El id de Google es una función determinista de (conexión, recurso, generación), de '
  'modo que reintentar tras una caída entre «crear en Google» y «guardar el vínculo» no duplica.';

COMMENT ON COLUMN google_calendar_event_link.visible_hash IS
  'Huella de lo que el invitado ve del evento (título, horario, correo invitado). Solo si cambia se '
  'notifica al invitado: refrescar el recordatorio de todos los eventos no envía correos (DEC-122).';

CREATE TRIGGER google_calendar_event_link_set_updated_at
  BEFORE UPDATE ON google_calendar_event_link
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE google_calendar_event_link ENABLE ROW LEVEL SECURITY;
ALTER TABLE google_calendar_event_link FORCE  ROW LEVEL SECURITY;

CREATE POLICY google_calendar_event_link_all_admin_policy ON google_calendar_event_link
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY google_calendar_event_link_select_tenant_policy ON google_calendar_event_link
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT ON TABLE google_calendar_event_link TO barberia_app;

-- ---------------------------------------------------------------------------
-- 4. Cola de trabajos
-- ---------------------------------------------------------------------------

CREATE TABLE google_calendar_sync_job (
  id               uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id    uuid        NOT NULL,
  connection_id    uuid        NOT NULL,
  resource_type    text        NOT NULL,
  resource_id      uuid        NOT NULL,
  status           text        NOT NULL DEFAULT 'pending',
  attempts         integer     NOT NULL DEFAULT 0,
  run_at           timestamptz NOT NULL DEFAULT now(),
  claim_token      uuid,
  lease_expires_at timestamptz,
  last_error_code  text,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT google_calendar_sync_job_id_pk PRIMARY KEY (id),

  CONSTRAINT google_calendar_sync_job_connection_fk FOREIGN KEY (barbershop_id, connection_id)
    REFERENCES google_calendar_connection (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT google_calendar_sync_job_resource_type_ck
    CHECK (resource_type IN ('appointment', 'time_block')),
  CONSTRAINT google_calendar_sync_job_status_ck
    CHECK (status IN ('pending', 'processing', 'failed')),
  CONSTRAINT google_calendar_sync_job_attempts_ck CHECK (attempts >= 0),

  -- Un trabajo en proceso siempre tiene su reclamación (DDL-CON-01).
  CONSTRAINT google_calendar_sync_job_claim_ck CHECK (
    (status = 'processing' AND claim_token IS NOT NULL AND lease_expires_at IS NOT NULL)
    OR (status <> 'processing' AND claim_token IS NULL AND lease_expires_at IS NULL)
  )
);

COMMENT ON TABLE google_calendar_sync_job IS
  'Cola propia de la publicación en Google Calendar (issue #324, DEC-102): «reconcilia este recurso». '
  'Se escribe en la misma transacción que el cambio de negocio y la consume el worker con reclamo '
  'por lease (SKIP LOCKED) y finalización por CAS sobre claim_token (DDL-CON-01). Un trabajo '
  'terminado se borra; uno que agota los intentos queda `failed` hasta «Sincronizar ahora». Sin '
  'datos personales.';

COMMENT ON COLUMN google_calendar_sync_job.claim_token IS
  'Reclamación vigente. Solo quien la tiene puede finalizar el trabajo; si el lease vence, otro '
  'worker lo reclama con un token nuevo y el primero ya no puede finalizarlo.';

-- Varios cambios seguidos del mismo recurso se funden en UN trabajo pendiente.
CREATE UNIQUE INDEX google_calendar_sync_job_pending_uk
  ON google_calendar_sync_job (connection_id, resource_type, resource_id)
  WHERE status = 'pending';

CREATE INDEX idx_google_calendar_sync_job_due
  ON google_calendar_sync_job (run_at) WHERE status = 'pending';
CREATE INDEX idx_google_calendar_sync_job_lease
  ON google_calendar_sync_job (lease_expires_at) WHERE status = 'processing';
CREATE INDEX idx_google_calendar_sync_job_connection
  ON google_calendar_sync_job (barbershop_id, connection_id, status);

CREATE TRIGGER google_calendar_sync_job_set_updated_at
  BEFORE UPDATE ON google_calendar_sync_job
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE google_calendar_sync_job ENABLE ROW LEVEL SECURITY;
ALTER TABLE google_calendar_sync_job FORCE  ROW LEVEL SECURITY;

CREATE POLICY google_calendar_sync_job_all_admin_policy ON google_calendar_sync_job
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY google_calendar_sync_job_select_tenant_policy ON google_calendar_sync_job
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY google_calendar_sync_job_insert_tenant_policy ON google_calendar_sync_job
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY google_calendar_sync_job_update_tenant_policy ON google_calendar_sync_job
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY google_calendar_sync_job_delete_tenant_policy ON google_calendar_sync_job
  FOR DELETE TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- La aplicación encola (en la transacción del cambio), refresca la cola de una
-- conexión («Sincronizar ahora») y la vacía al desconectar.
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE google_calendar_sync_job TO barberia_app;

-- ---------------------------------------------------------------------------
-- 5. Funciones del worker (SECURITY DEFINER, sin contexto de tenant)
-- ---------------------------------------------------------------------------

-- 5.1 Reclamo con lease. Un único UPDATE atómico sobre un SELECT ... FOR UPDATE SKIP
-- LOCKED: dos workers nunca reclaman el mismo trabajo, y un `processing` con lease
-- vencido (worker caído) vuelve a ser elegible con un claim_token nuevo.
CREATE FUNCTION gcal_claim_jobs(
  p_limit         integer,
  p_lease_seconds integer,
  p_now           timestamptz DEFAULT pg_catalog.now()
)
RETURNS TABLE (
  job_id        uuid,
  claim_token   uuid,
  barbershop_id uuid,
  connection_id uuid,
  resource_type text,
  resource_id   uuid,
  attempts      integer
)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
#variable_conflict use_column
BEGIN
  IF p_limit IS NULL OR p_limit < 1 OR p_limit > 100 THEN
    RAISE EXCEPTION 'gcal_claim_jobs: p_limit fuera de rango (1-100).';
  END IF;
  IF p_lease_seconds IS NULL OR p_lease_seconds < 10 OR p_lease_seconds > 3600 THEN
    RAISE EXCEPTION 'gcal_claim_jobs: p_lease_seconds fuera de rango (10-3600).';
  END IF;

  RETURN QUERY
  WITH due AS (
    SELECT j.id
      FROM public.google_calendar_sync_job j
      JOIN public.google_calendar_connection c
        ON c.barbershop_id = j.barbershop_id AND c.id = j.connection_id
     WHERE c.status IN ('connected', 'error')
       AND (
            (j.status = 'pending' AND j.run_at <= p_now)
         OR (j.status = 'processing' AND j.lease_expires_at <= p_now)
       )
     ORDER BY j.run_at, j.created_at
     LIMIT p_limit
       FOR UPDATE OF j SKIP LOCKED
  )
  UPDATE public.google_calendar_sync_job j
     SET status = 'processing',
         claim_token = pg_catalog.gen_random_uuid(),
         lease_expires_at = p_now + pg_catalog.make_interval(secs => p_lease_seconds),
         attempts = j.attempts + 1,
         updated_at = p_now
    FROM due
   WHERE j.id = due.id
  RETURNING j.id, j.claim_token, j.barbershop_id, j.connection_id, j.resource_type, j.resource_id, j.attempts;
END;
$$;

REVOKE ALL     ON FUNCTION gcal_claim_jobs(integer, integer, timestamptz) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION gcal_claim_jobs(integer, integer, timestamptz) TO barberia_worker;

COMMENT ON FUNCTION gcal_claim_jobs(integer, integer, timestamptz) IS
  'Reclamo con lease de la cola de Google Calendar (DDL-CON-01): SKIP LOCKED sobre trabajos '
  'pending vencidos o processing con lease vencido de conexiones que siguen publicando. '
  'Exclusivo de barberia_worker.';

-- 5.2 Contexto de UN trabajo reclamado: lo mínimo para armar el evento.
CREATE FUNCTION gcal_job_context(p_job_id uuid, p_claim_token uuid)
RETURNS TABLE (
  barber_id           uuid,
  connection_status   text,
  calendar_id         text,
  reminder_minutes    integer,
  token_ciphertext    bytea,
  token_key_id        text,
  barbershop_timezone text,
  link_event_id       text,
  link_generation     integer,
  link_visible_hash   text,
  resource_type       text,
  resource_id         uuid,
  resource_found      boolean,
  barber_matches      boolean,
  appt_status         text,
  starts_at           timestamptz,
  ends_at             timestamptz,
  attendee_name       text,
  service_name        text,
  customer_email      text,
  block_type          text,
  block_source        text,
  block_deleted       boolean
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = ''
AS $$
  SELECT c.barber_id, c.status, c.calendar_id, c.reminder_minutes, c.refresh_token_ciphertext, c.token_key_id,
         b.timezone,
         l.google_event_id, l.event_generation, l.visible_hash,
         j.resource_type, j.resource_id,
         (a.id IS NOT NULL OR t.id IS NOT NULL),
         COALESCE(a.barber_id, t.barber_id) = c.barber_id,
         a.status, COALESCE(a.starts_at, t.starts_at), COALESCE(a.ends_at, t.ends_at),
         a.attendee_name, a.service_name_snapshot, cu.email,
         t.block_type, t.source, (t.deleted_at IS NOT NULL)
    FROM public.google_calendar_sync_job j
    JOIN public.google_calendar_connection c
      ON c.barbershop_id = j.barbershop_id AND c.id = j.connection_id
    JOIN public.barbershop b ON b.id = j.barbershop_id
    LEFT JOIN public.google_calendar_event_link l
      ON l.connection_id = j.connection_id
     AND l.resource_type = j.resource_type AND l.resource_id = j.resource_id
    LEFT JOIN public.appointment a
      ON j.resource_type = 'appointment' AND a.barbershop_id = j.barbershop_id AND a.id = j.resource_id
    LEFT JOIN public.customer cu
      ON cu.barbershop_id = a.barbershop_id AND cu.id = a.customer_id
    LEFT JOIN public.time_block t
      ON j.resource_type = 'time_block' AND t.barbershop_id = j.barbershop_id AND t.id = j.resource_id
   WHERE j.id = p_job_id AND j.claim_token = p_claim_token AND j.status = 'processing';
$$;

REVOKE ALL     ON FUNCTION gcal_job_context(uuid, uuid) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION gcal_job_context(uuid, uuid) TO barberia_worker;

COMMENT ON FUNCTION gcal_job_context(uuid, uuid) IS
  'Datos mínimos para publicar UN trabajo reclamado; exige su claim_token vigente, así que no sirve '
  'para leer citas arbitrarias. Incluye el correo del cliente solo para invitarlo (DEC-122). '
  'Exclusivo de barberia_worker.';

-- 5.3 Finalización por CAS sobre claim_token.
CREATE FUNCTION gcal_finish_job(
  p_job_id        uuid,
  p_claim_token   uuid,
  p_outcome       text,
  p_event_id      text DEFAULT NULL,
  p_etag          text DEFAULT NULL,
  p_generation    integer DEFAULT NULL,
  p_visible_hash  text DEFAULT NULL,
  p_error_code    text DEFAULT NULL,
  p_retry_seconds integer DEFAULT NULL,
  p_max_attempts  integer DEFAULT 8,
  p_now           timestamptz DEFAULT pg_catalog.now()
)
RETURNS boolean
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_job public.google_calendar_sync_job%ROWTYPE;
BEGIN
  IF p_outcome NOT IN ('published', 'removed', 'skipped', 'retry', 'reauth', 'permanent') THEN
    RAISE EXCEPTION 'gcal_finish_job: resultado inválido %.', p_outcome;
  END IF;

  -- CAS: solo quien tiene la reclamación vigente finaliza.
  SELECT * INTO v_job
    FROM public.google_calendar_sync_job
   WHERE id = p_job_id AND claim_token = p_claim_token AND status = 'processing'
   FOR UPDATE;
  IF NOT FOUND THEN
    RETURN false;
  END IF;

  IF p_outcome = 'published' THEN
    IF p_event_id IS NULL OR p_generation IS NULL THEN
      RAISE EXCEPTION 'gcal_finish_job: published exige p_event_id y p_generation.';
    END IF;
    INSERT INTO public.google_calendar_event_link
           (connection_id, resource_type, resource_id, barbershop_id, google_event_id, etag,
            event_generation, visible_hash)
    VALUES (v_job.connection_id, v_job.resource_type, v_job.resource_id, v_job.barbershop_id,
            p_event_id, p_etag, p_generation, p_visible_hash)
    ON CONFLICT (connection_id, resource_type, resource_id) DO UPDATE
       SET google_event_id = EXCLUDED.google_event_id, etag = EXCLUDED.etag,
           event_generation = EXCLUDED.event_generation, visible_hash = EXCLUDED.visible_hash;
    UPDATE public.google_calendar_connection
       SET last_synced_at = p_now, last_error_code = NULL,
           status = CASE WHEN status = 'error' THEN 'connected' ELSE status END
     WHERE barbershop_id = v_job.barbershop_id AND id = v_job.connection_id;
    DELETE FROM public.google_calendar_sync_job WHERE id = v_job.id;

  ELSIF p_outcome = 'removed' THEN
    DELETE FROM public.google_calendar_event_link
     WHERE connection_id = v_job.connection_id
       AND resource_type = v_job.resource_type AND resource_id = v_job.resource_id;
    DELETE FROM public.google_calendar_sync_job WHERE id = v_job.id;

  ELSIF p_outcome = 'skipped' THEN
    DELETE FROM public.google_calendar_sync_job WHERE id = v_job.id;

  ELSIF p_outcome = 'retry' THEN
    IF v_job.attempts >= COALESCE(p_max_attempts, 8) THEN
      UPDATE public.google_calendar_sync_job
         SET status = 'failed', claim_token = NULL, lease_expires_at = NULL,
             last_error_code = p_error_code
       WHERE id = v_job.id;
    ELSE
      UPDATE public.google_calendar_sync_job
         SET status = 'pending', claim_token = NULL, lease_expires_at = NULL,
             run_at = p_now + pg_catalog.make_interval(secs => COALESCE(p_retry_seconds, 60)),
             last_error_code = p_error_code
       WHERE id = v_job.id;
    END IF;

  ELSIF p_outcome = 'reauth' THEN
    -- Google ya no acepta el permiso: se borran las credenciales y se vacía la cola de esa
    -- conexión (al reconectar, la publicación inicial vuelve a encolar lo vigente).
    UPDATE public.google_calendar_connection
       SET status = 'reauth_required', refresh_token_ciphertext = NULL, token_key_id = NULL,
           last_error_code = COALESCE(p_error_code, 'token_revoked')
     WHERE barbershop_id = v_job.barbershop_id AND id = v_job.connection_id;
    DELETE FROM public.google_calendar_sync_job
     WHERE barbershop_id = v_job.barbershop_id AND connection_id = v_job.connection_id;

  ELSE  -- 'permanent': permiso denegado, calendario eliminado... el token sigue siendo válido
    UPDATE public.google_calendar_sync_job
       SET status = 'failed', claim_token = NULL, lease_expires_at = NULL,
           last_error_code = p_error_code
     WHERE id = v_job.id;
    UPDATE public.google_calendar_connection
       SET status = 'error', last_error_code = p_error_code
     WHERE barbershop_id = v_job.barbershop_id AND id = v_job.connection_id
       AND status IN ('connected', 'error');
  END IF;

  RETURN true;
END;
$$;

REVOKE ALL     ON FUNCTION gcal_finish_job(uuid, uuid, text, text, text, integer, text, text, integer, integer, timestamptz) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION gcal_finish_job(uuid, uuid, text, text, text, integer, text, text, integer, integer, timestamptz) TO barberia_worker;

COMMENT ON FUNCTION gcal_finish_job(uuid, uuid, text, text, text, integer, text, text, integer, integer, timestamptz) IS
  'Finalización por CAS sobre claim_token (DDL-CON-01): published (crea o actualiza el vínculo), '
  'removed (borra el vínculo), skipped, retry (con backoff; failed al agotar intentos), reauth '
  '(credenciales borradas y cola de la conexión vaciada) y permanent. Devuelve false si la '
  'reclamación ya no es vigente. Exclusivo de barberia_worker.';

-- 5.4 Chequeo periódico de eventos borrados (DEC-101): reservar una conexión, listar sus
-- vínculos vigentes y reencolar los que faltan en Google.
CREATE FUNCTION gcal_restore_claim_connection(
  p_interval_seconds integer,
  p_now              timestamptz DEFAULT pg_catalog.now()
)
RETURNS TABLE (
  connection_id    uuid,
  barbershop_id    uuid,
  barber_id        uuid,
  calendar_id      text,
  token_ciphertext bytea,
  token_key_id     text
)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
#variable_conflict use_column
BEGIN
  IF p_interval_seconds IS NULL OR p_interval_seconds < 30 OR p_interval_seconds > 86400 THEN
    RAISE EXCEPTION 'gcal_restore_claim_connection: p_interval_seconds fuera de rango (30-86400).';
  END IF;

  RETURN QUERY
  WITH due AS (
    SELECT c.id
      FROM public.google_calendar_connection c
     WHERE c.status = 'connected'
       AND (c.last_restore_check_at IS NULL
            OR c.last_restore_check_at <= p_now - pg_catalog.make_interval(secs => p_interval_seconds))
     ORDER BY c.last_restore_check_at NULLS FIRST
     LIMIT 1
       FOR UPDATE SKIP LOCKED
  )
  UPDATE public.google_calendar_connection c
     SET last_restore_check_at = p_now
    FROM due
   WHERE c.id = due.id
  RETURNING c.id, c.barbershop_id, c.barber_id, c.calendar_id, c.refresh_token_ciphertext, c.token_key_id;
END;
$$;

REVOKE ALL     ON FUNCTION gcal_restore_claim_connection(integer, timestamptz) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION gcal_restore_claim_connection(integer, timestamptz) TO barberia_worker;

COMMENT ON FUNCTION gcal_restore_claim_connection(integer, timestamptz) IS
  'Reserva (SKIP LOCKED) la siguiente conexión conectada cuyo chequeo de eventos borrados venció. '
  'Exclusivo de barberia_worker.';

CREATE FUNCTION gcal_restore_links(p_connection_id uuid, p_now timestamptz DEFAULT pg_catalog.now())
RETURNS TABLE (
  resource_type   text,
  resource_id     uuid,
  google_event_id text
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = ''
AS $$
  -- Solo recursos que SIGUEN vigentes y futuros: una cita cancelada en la app, terminal o
  -- pasada y un bloqueo retirado nunca se recrean (DEC-101).
  SELECT l.resource_type, l.resource_id, l.google_event_id
    FROM public.google_calendar_event_link l
    LEFT JOIN public.appointment a
      ON l.resource_type = 'appointment' AND a.barbershop_id = l.barbershop_id AND a.id = l.resource_id
    LEFT JOIN public.time_block t
      ON l.resource_type = 'time_block' AND t.barbershop_id = l.barbershop_id AND t.id = l.resource_id
   WHERE l.connection_id = p_connection_id
     AND (
          (a.id IS NOT NULL AND a.status = 'confirmed' AND a.ends_at > p_now)
       OR (t.id IS NOT NULL AND t.deleted_at IS NULL AND t.ends_at > p_now)
     );
$$;

REVOKE ALL     ON FUNCTION gcal_restore_links(uuid, timestamptz) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION gcal_restore_links(uuid, timestamptz) TO barberia_worker;

COMMENT ON FUNCTION gcal_restore_links(uuid, timestamptz) IS
  'Vínculos de una conexión cuyo recurso sigue vigente y futuro: los únicos eventos que el chequeo '
  'periódico puede restaurar. Exclusivo de barberia_worker.';

CREATE FUNCTION gcal_enqueue_missing(p_connection_id uuid, p_resource_type text, p_resource_id uuid)
RETURNS boolean
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_inserted integer;
BEGIN
  IF p_resource_type NOT IN ('appointment', 'time_block') THEN
    RAISE EXCEPTION 'gcal_enqueue_missing: tipo de recurso inválido %.', p_resource_type;
  END IF;

  INSERT INTO public.google_calendar_sync_job (barbershop_id, connection_id, resource_type, resource_id)
  SELECT c.barbershop_id, c.id, p_resource_type, p_resource_id
    FROM public.google_calendar_connection c
   WHERE c.id = p_connection_id AND c.status IN ('connected', 'error')
  ON CONFLICT (connection_id, resource_type, resource_id) WHERE status = 'pending' DO NOTHING;

  GET DIAGNOSTICS v_inserted = ROW_COUNT;
  RETURN v_inserted > 0;
END;
$$;

REVOKE ALL     ON FUNCTION gcal_enqueue_missing(uuid, text, uuid) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION gcal_enqueue_missing(uuid, text, uuid) TO barberia_worker;

COMMENT ON FUNCTION gcal_enqueue_missing(uuid, text, uuid) IS
  'Reencola un recurso cuyo evento falta en Google. Solo encola para conexiones que siguen '
  'publicando y se funde con un trabajo pendiente. Exclusivo de barberia_worker.';
