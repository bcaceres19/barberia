-- Pruebas SQL del issue #324 (DEC-099, DEC-101, DEC-102): cola de publicación en
-- Google Calendar. Cubren a nivel de base de datos los privilegios del worker
-- (solo funciones, nada de tablas), el reclamo con lease y la finalización por CAS
-- (DDL-CON-01), el único trabajo pendiente por recurso, el CHECK de la
-- reclamación, el aislamiento RLS entre dos barberías y las funciones del chequeo
-- de eventos borrados. La orquestación de Google vive en Go
-- (internal/modules/googlecalendar), no aquí.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/google_calendar_sync.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app` y
-- `SET ROLE barberia_worker` (superusuario en local/CI efímero). Cada escenario
-- usa un `p_now` explícito en vez de esperar al reloj (estrategia-pruebas.md §7).
-- Todo se revierte con ROLLBACK.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== Issue #324 · pruebas de la cola de Google Calendar ==='

-- ---------------------------------------------------------------------------
-- Privilegios: el worker solo ejecuta funciones; la aplicación no las ejecuta
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_table text;
  v_fn    text;
BEGIN
  FOREACH v_table IN ARRAY ARRAY['google_calendar_connection', 'google_calendar_sync_job',
                                 'google_calendar_event_link', 'appointment', 'customer', 'time_block']
  LOOP
    IF has_table_privilege('barberia_worker', v_table, 'SELECT')
       OR has_table_privilege('barberia_worker', v_table, 'INSERT')
       OR has_table_privilege('barberia_worker', v_table, 'UPDATE')
       OR has_table_privilege('barberia_worker', v_table, 'DELETE') THEN
      RAISE EXCEPTION 'barberia_worker no debe tener acceso directo a % (DEC-040).', v_table;
    END IF;
  END LOOP;

  FOREACH v_fn IN ARRAY ARRAY[
    'gcal_claim_jobs(integer, integer, timestamptz)',
    'gcal_job_context(uuid, uuid)',
    'gcal_finish_job(uuid, uuid, text, text, text, integer, text, text, integer, integer, timestamptz)',
    'gcal_restore_claim_connection(integer, timestamptz)',
    'gcal_restore_links(uuid, timestamptz)',
    'gcal_enqueue_missing(uuid, text, uuid)']
  LOOP
    IF NOT has_function_privilege('barberia_worker', v_fn, 'EXECUTE') THEN
      RAISE EXCEPTION 'barberia_worker debe poder ejecutar %.', v_fn;
    END IF;
    IF has_function_privilege('barberia_app', v_fn, 'EXECUTE') THEN
      RAISE EXCEPTION 'barberia_app no debe poder ejecutar % (exclusivo del worker).', v_fn;
    END IF;
    IF NOT (SELECT prosecdef FROM pg_proc WHERE oid = v_fn::regprocedure) THEN
      RAISE EXCEPTION '% debe ser SECURITY DEFINER.', v_fn;
    END IF;
  END LOOP;

  IF has_table_privilege('barberia_app', 'google_calendar_event_link', 'INSERT')
     OR has_table_privilege('barberia_app', 'google_calendar_event_link', 'UPDATE')
     OR has_table_privilege('barberia_app', 'google_calendar_event_link', 'DELETE') THEN
    RAISE EXCEPTION 'los vínculos solo los escribe el worker: barberia_app solo lee.';
  END IF;
END
$$;
\echo 'privilegios OK · worker solo funciones, aplicación sin funciones y vínculos de solo lectura'

-- ---------------------------------------------------------------------------
-- Escenario base: un barbero conectado, una cita y su trabajo (como el hook)
-- ---------------------------------------------------------------------------
BEGIN;
RESET ROLE; -- el arnés conecta como superusuario: crea fixtures sin pasar por RLS

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('c3240001-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111', 'Barbero cola'),
  ('c3240001-0000-4000-8000-0000000000b1', '22222222-2222-2222-2222-222222222222', 'Barbero cola B');
INSERT INTO google_calendar_connection
  (id, barbershop_id, barber_id, status, google_account_email, refresh_token_ciphertext, token_key_id)
VALUES
  ('c3240002-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111',
   'c3240001-0000-4000-8000-000000000001', 'connected', 'barbero@ejemplo.test', '\x0102'::bytea, 'v1'),
  ('c3240002-0000-4000-8000-0000000000b1', '22222222-2222-2222-2222-222222222222',
   'c3240001-0000-4000-8000-0000000000b1', 'connected', 'b@ejemplo.test', '\x0102'::bytea, 'v1');

INSERT INTO google_calendar_sync_job (id, barbershop_id, connection_id, resource_type, resource_id, run_at) VALUES
  ('c3240003-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111',
   'c3240002-0000-4000-8000-000000000001', 'appointment', 'a3240000-0000-4000-8000-000000000001',
   '2026-01-05 09:00:00-05'),
  ('c3240003-0000-4000-8000-0000000000b1', '22222222-2222-2222-2222-222222222222',
   'c3240002-0000-4000-8000-0000000000b1', 'appointment', 'a3240000-0000-4000-8000-0000000000b1',
   '2026-01-05 09:00:00-05');

-- Un solo trabajo PENDIENTE por recurso; uno en proceso o fallido no compite.
DO $$
BEGIN
  BEGIN
    INSERT INTO google_calendar_sync_job (barbershop_id, connection_id, resource_type, resource_id)
    VALUES ('11111111-1111-1111-1111-111111111111', 'c3240002-0000-4000-8000-000000000001',
            'appointment', 'a3240000-0000-4000-8000-000000000001');
    RAISE EXCEPTION 'dos trabajos pendientes del mismo recurso deben rechazarse.';
  EXCEPTION WHEN unique_violation THEN NULL;
  END;

  -- processing exige su reclamación (y pending/failed la prohíben).
  BEGIN
    UPDATE google_calendar_sync_job SET status = 'processing'
     WHERE id = 'c3240003-0000-4000-8000-000000000001';
    RAISE EXCEPTION 'processing sin claim_token debe rechazarse.';
  EXCEPTION WHEN check_violation THEN NULL;
  END;
  BEGIN
    UPDATE google_calendar_sync_job SET claim_token = gen_random_uuid(), lease_expires_at = now()
     WHERE id = 'c3240003-0000-4000-8000-000000000001';
    RAISE EXCEPTION 'un trabajo pendiente no puede llevar claim_token.';
  EXCEPTION WHEN check_violation THEN NULL;
  END;
  BEGIN
    UPDATE google_calendar_sync_job SET status = 'cancelado'
     WHERE id = 'c3240003-0000-4000-8000-000000000001';
    RAISE EXCEPTION 'un estado fuera de la lista cerrada debe rechazarse.';
  EXCEPTION WHEN check_violation THEN NULL;
  END;
END
$$;
\echo 'restricciones OK · un pendiente por recurso y reclamación ligada al estado'

-- ---------------------------------------------------------------------------
-- Reclamo con lease, sin doble claim vigente, y finalización por CAS
-- ---------------------------------------------------------------------------
SET ROLE barberia_worker;

DO $$
DECLARE
  v_first   record;
  v_second  record;
  v_third   record;
  v_ok      boolean;
  v_count   integer;
BEGIN
  -- Antes de su run_at no hay nada que reclamar.
  SELECT count(*) INTO v_count FROM gcal_claim_jobs(10, 60, '2026-01-05 08:59:00-05'::timestamptz);
  IF v_count <> 0 THEN
    RAISE EXCEPTION 'un trabajo no vencido no debe reclamarse.';
  END IF;

  -- Reclamo: de ambas barberías, cada una con su token (el worker no tiene tenant).
  SELECT count(*) INTO v_count FROM gcal_claim_jobs(1, 60, '2026-01-05 09:00:00-05'::timestamptz);
  IF v_count <> 1 THEN
    RAISE EXCEPTION 'con p_limit=1 se reclama exactamente uno, no %.', v_count;
  END IF;
  SELECT * INTO v_first FROM gcal_claim_jobs(10, 60, '2026-01-05 09:00:30-05'::timestamptz);
  IF v_first IS NULL OR v_first.attempts <> 1 THEN
    RAISE EXCEPTION 'el segundo trabajo debe reclamarse con su contador en 1.';
  END IF;

  -- Con ambos con lease vigente, nadie los reclama de nuevo.
  SELECT count(*) INTO v_count FROM gcal_claim_jobs(10, 60, '2026-01-05 09:00:45-05'::timestamptz);
  IF v_count <> 0 THEN
    RAISE EXCEPTION 'un trabajo con lease vigente no debe reclamarse (reclamados: %).', v_count;
  END IF;

  -- Vencido el lease, se reclama con un token NUEVO y el contador sube.
  SELECT * INTO v_second FROM gcal_claim_jobs(10, 60, '2026-01-05 09:05:00-05'::timestamptz)
   WHERE job_id = v_first.job_id;
  IF v_second IS NULL OR v_second.claim_token = v_first.claim_token OR v_second.attempts <> 2 THEN
    RAISE EXCEPTION 'con el lease vencido debe reclamarse con otro token y attempts=2.';
  END IF;

  -- El primer token ya no finaliza (CAS); el vigente sí.
  v_ok := gcal_finish_job(v_first.job_id, v_first.claim_token, 'skipped');
  IF v_ok THEN
    RAISE EXCEPTION 'el claim_token vencido no debe poder finalizar.';
  END IF;
  v_ok := gcal_finish_job(v_second.job_id, v_second.claim_token, 'skipped');
  IF NOT v_ok THEN
    RAISE EXCEPTION 'el claim_token vigente debe poder finalizar.';
  END IF;
  -- Finalizar dos veces con el mismo token es falso la segunda vez.
  v_ok := gcal_finish_job(v_second.job_id, v_second.claim_token, 'skipped');
  IF v_ok THEN
    RAISE EXCEPTION 'un trabajo ya finalizado no se finaliza de nuevo.';
  END IF;

  -- Un resultado inventado se rechaza.
  BEGIN
    PERFORM gcal_finish_job(v_second.job_id, v_second.claim_token, 'inventado');
    RAISE EXCEPTION 'un resultado fuera de la lista debe rechazarse.';
  EXCEPTION WHEN raise_exception THEN
    IF SQLERRM NOT LIKE 'gcal_finish_job: resultado inválido%' THEN RAISE; END IF;
  END;
END
$$;
\echo 'lease OK · un solo claim vigente, token nuevo al vencer y CAS sobre claim_token'

-- ---------------------------------------------------------------------------
-- Finalización: published crea el vínculo; retry agota intentos; reauth vacía la cola
-- ---------------------------------------------------------------------------
RESET ROLE; -- el arnés conecta como superusuario: crea fixtures sin pasar por RLS
INSERT INTO google_calendar_sync_job (id, barbershop_id, connection_id, resource_type, resource_id, run_at) VALUES
  ('c3240004-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111',
   'c3240002-0000-4000-8000-000000000001', 'appointment', 'a3240000-0000-4000-8000-000000000002',
   '2026-02-01 09:00:00-05'),
  ('c3240004-0000-4000-8000-000000000002', '11111111-1111-1111-1111-111111111111',
   'c3240002-0000-4000-8000-000000000001', 'appointment', 'a3240000-0000-4000-8000-000000000003',
   '2026-02-01 09:00:00-05');
SET ROLE barberia_worker;

DO $$
DECLARE
  v_job   record;
  v_ok    boolean;
  v_row   record;
BEGIN
  SELECT * INTO v_job FROM gcal_claim_jobs(1, 60, '2026-02-01 09:00:00-05'::timestamptz);

  -- published sin identificador del evento se rechaza.
  BEGIN
    PERFORM gcal_finish_job(v_job.job_id, v_job.claim_token, 'published');
    RAISE EXCEPTION 'published exige p_event_id y p_generation.';
  EXCEPTION WHEN raise_exception THEN
    IF SQLERRM NOT LIKE 'gcal_finish_job: published exige%' THEN RAISE; END IF;
  END;

  v_ok := gcal_finish_job(v_job.job_id, v_job.claim_token, 'published',
                          'navaevento0001', '"e1"', 1, 'huella', NULL, NULL, 8,
                          '2026-02-01 09:00:05-05'::timestamptz);
  IF NOT v_ok THEN RAISE EXCEPTION 'published debe finalizar.'; END IF;

  -- El contexto del siguiente trabajo del MISMO recurso ve el vínculo (crear o actualizar).
  SELECT * INTO v_job FROM gcal_claim_jobs(1, 60, '2026-02-01 09:00:10-05'::timestamptz);
  v_ok := gcal_finish_job(v_job.job_id, v_job.claim_token, 'retry', NULL, NULL, NULL, NULL,
                          'google_unavailable', 60, 2, '2026-02-01 09:00:10-05'::timestamptz);
  IF NOT v_ok THEN RAISE EXCEPTION 'retry debe finalizar.'; END IF;
  -- Antes del backoff no vuelve; pasado, sí (attempts=2 alcanza el máximo de 2).
  IF EXISTS (SELECT 1 FROM gcal_claim_jobs(50, 60, '2026-02-01 09:00:30-05'::timestamptz)
              WHERE job_id = v_job.job_id) THEN
    RAISE EXCEPTION 'antes del backoff no debe reclamarse.';
  END IF;
  SELECT * INTO v_job FROM gcal_claim_jobs(50, 60, '2026-02-01 09:01:30-05'::timestamptz)
   WHERE job_id = v_job.job_id;
  IF v_job.attempts <> 2 THEN RAISE EXCEPTION 'segundo intento esperado, fue %.', v_job.attempts; END IF;
  v_ok := gcal_finish_job(v_job.job_id, v_job.claim_token, 'retry', NULL, NULL, NULL, NULL,
                          'google_unavailable', 60, 2, '2026-02-01 09:01:30-05'::timestamptz);
  -- Agotados los intentos queda failed y deja de reclamarse (sin bucle infinito).
  IF EXISTS (SELECT 1 FROM gcal_claim_jobs(10, 60, '2027-01-01 00:00:00-05'::timestamptz)
              WHERE job_id = v_job.job_id) THEN
    RAISE EXCEPTION 'un trabajo failed no debe reclamarse jamás.';
  END IF;
END
$$;

RESET ROLE; -- el arnés conecta como superusuario: crea fixtures sin pasar por RLS
DO $$
DECLARE
  v_status text;
  v_event  text;
BEGIN
  SELECT status INTO v_status FROM google_calendar_sync_job WHERE id = 'c3240004-0000-4000-8000-000000000002'
     OR id = 'c3240004-0000-4000-8000-000000000001' ORDER BY updated_at DESC LIMIT 1;
  SELECT google_event_id INTO v_event FROM google_calendar_event_link
   WHERE connection_id = 'c3240002-0000-4000-8000-000000000001' AND google_event_id = 'navaevento0001';
  IF v_event IS DISTINCT FROM 'navaevento0001' THEN
    RAISE EXCEPTION 'published debe crear el vínculo (leyó %).', v_event;
  END IF;
  IF (SELECT last_synced_at FROM google_calendar_connection WHERE id = 'c3240002-0000-4000-8000-000000000001') IS NULL THEN
    RAISE EXCEPTION 'published debe fijar last_synced_at.';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM google_calendar_sync_job
                  WHERE connection_id = 'c3240002-0000-4000-8000-000000000001' AND status = 'failed') THEN
    RAISE EXCEPTION 'agotar los intentos debe dejar el trabajo failed.';
  END IF;
END
$$;
\echo 'finalización OK · published crea el vínculo, retry con backoff y failed al agotar intentos'

-- reauth: credenciales borradas y cola de la conexión vaciada.
INSERT INTO google_calendar_sync_job (id, barbershop_id, connection_id, resource_type, resource_id, run_at) VALUES
  ('c3240005-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111',
   'c3240002-0000-4000-8000-000000000001', 'appointment', 'a3240000-0000-4000-8000-000000000004',
   '2026-03-01 09:00:00-05');
SET ROLE barberia_worker;
DO $$
DECLARE
  v_job record;
BEGIN
  SELECT * INTO v_job FROM gcal_claim_jobs(1, 60, '2026-03-01 09:00:00-05'::timestamptz);
  IF NOT gcal_finish_job(v_job.job_id, v_job.claim_token, 'reauth', NULL, NULL, NULL, NULL, 'token_revoked') THEN
    RAISE EXCEPTION 'reauth debe finalizar.';
  END IF;
END
$$;
RESET ROLE; -- el arnés conecta como superusuario: crea fixtures sin pasar por RLS
DO $$
DECLARE
  v_conn record;
BEGIN
  SELECT * INTO v_conn FROM google_calendar_connection WHERE id = 'c3240002-0000-4000-8000-000000000001';
  IF v_conn.status <> 'reauth_required' OR v_conn.refresh_token_ciphertext IS NOT NULL
     OR v_conn.last_error_code <> 'token_revoked' THEN
    RAISE EXCEPTION 'reauth debe dejar reauth_required sin credenciales (estado %).', v_conn.status;
  END IF;
  IF EXISTS (SELECT 1 FROM google_calendar_sync_job WHERE connection_id = 'c3240002-0000-4000-8000-000000000001') THEN
    RAISE EXCEPTION 'reauth debe vaciar la cola de la conexión.';
  END IF;
  IF (SELECT status FROM google_calendar_connection WHERE id = 'c3240002-0000-4000-8000-0000000000b1') <> 'connected' THEN
    RAISE EXCEPTION 'reauth de A no debe tocar la conexión de B.';
  END IF;
END
$$;
\echo 'reauth OK · credenciales borradas y cola de esa conexión vacía, sin tocar la conexión de otra barbería'

-- Una conexión sin permiso deja de reclamar trabajos aunque queden filas.
INSERT INTO google_calendar_sync_job (id, barbershop_id, connection_id, resource_type, resource_id, run_at) VALUES
  ('c3240006-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111',
   'c3240002-0000-4000-8000-000000000001', 'appointment', 'a3240000-0000-4000-8000-000000000005',
   '2026-03-01 09:00:00-05');
SET ROLE barberia_worker;
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM gcal_claim_jobs(50, 60, '2026-03-02 09:00:00-05'::timestamptz)
              WHERE connection_id = 'c3240002-0000-4000-8000-000000000001') THEN
    RAISE EXCEPTION 'una conexión reauth_required no debe reclamar trabajos.';
  END IF;
END
$$;
\echo 'estado de la conexión OK · reauth_required no reclama'

RESET ROLE;
ROLLBACK;

-- ---------------------------------------------------------------------------
-- Aislamiento RLS de la aplicación y chequeo de eventos borrados
-- ---------------------------------------------------------------------------
BEGIN;
RESET ROLE; -- el arnés conecta como superusuario: crea fixtures sin pasar por RLS
INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('c3240011-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111', 'Barbero restaurar'),
  ('c3240011-0000-4000-8000-0000000000b1', '22222222-2222-2222-2222-222222222222', 'Barbero restaurar B');
INSERT INTO google_calendar_connection
  (id, barbershop_id, barber_id, status, refresh_token_ciphertext, token_key_id)
VALUES
  ('c3240012-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111',
   'c3240011-0000-4000-8000-000000000001', 'connected', '\x0102'::bytea, 'v1'),
  ('c3240012-0000-4000-8000-0000000000b1', '22222222-2222-2222-2222-222222222222',
   'c3240011-0000-4000-8000-0000000000b1', 'connected', '\x0102'::bytea, 'v1');
INSERT INTO google_calendar_sync_job (barbershop_id, connection_id, resource_type, resource_id) VALUES
  ('11111111-1111-1111-1111-111111111111', 'c3240012-0000-4000-8000-000000000001', 'appointment',
   'a3240000-0000-4000-8000-0000000000c1');

SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '22222222-2222-2222-2222-222222222222';
DO $$
DECLARE
  v_rows integer;
BEGIN
  IF EXISTS (SELECT 1 FROM google_calendar_sync_job WHERE connection_id = 'c3240012-0000-4000-8000-000000000001') THEN
    RAISE EXCEPTION 'la barbería B no debe ver la cola de la A.';
  END IF;
  UPDATE google_calendar_sync_job SET run_at = now() WHERE connection_id = 'c3240012-0000-4000-8000-000000000001';
  GET DIAGNOSTICS v_rows = ROW_COUNT;
  IF v_rows <> 0 THEN RAISE EXCEPTION 'B no debe reordenar la cola de A.'; END IF;
  DELETE FROM google_calendar_sync_job WHERE connection_id = 'c3240012-0000-4000-8000-000000000001';
  GET DIAGNOSTICS v_rows = ROW_COUNT;
  IF v_rows <> 0 THEN RAISE EXCEPTION 'B no debe borrar trabajos de A.'; END IF;

  -- B no puede encolar un recurso en la conexión de A (FK compuesta + RLS).
  BEGIN
    INSERT INTO google_calendar_sync_job (barbershop_id, connection_id, resource_type, resource_id)
    VALUES ('22222222-2222-2222-2222-222222222222', 'c3240012-0000-4000-8000-000000000001',
            'appointment', 'a3240000-0000-4000-8000-0000000000c2');
    RAISE EXCEPTION 'B no debe poder encolar sobre la conexión de A.';
  EXCEPTION WHEN foreign_key_violation THEN NULL;
  END;
END
$$;
RESET ROLE;

-- Chequeo de eventos borrados: se reserva una conexión vencida y no vuelve hasta el intervalo.
SET ROLE barberia_worker;
DO $$
DECLARE
  v_target record;
  v_again  integer;
  v_queued boolean;
BEGIN
  SELECT * INTO v_target FROM gcal_restore_claim_connection(300, '2026-04-01 09:00:00-05'::timestamptz)
   WHERE connection_id = 'c3240012-0000-4000-8000-000000000001';
  -- (la base compartida puede tener otras conexiones por delante: se reservan hasta dar con esta)
  IF v_target IS NULL THEN
    FOR i IN 1..200 LOOP
      SELECT * INTO v_target FROM gcal_restore_claim_connection(300, '2026-04-01 09:00:00-05'::timestamptz);
      EXIT WHEN v_target IS NULL OR v_target.connection_id = 'c3240012-0000-4000-8000-000000000001';
    END LOOP;
  END IF;
  IF v_target IS NULL OR v_target.connection_id <> 'c3240012-0000-4000-8000-000000000001' THEN
    RAISE EXCEPTION 'debe reservarse la conexión pendiente de comprobar.';
  END IF;
  IF v_target.barber_id IS NULL OR v_target.token_ciphertext IS NULL THEN
    RAISE EXCEPTION 'la reserva debe devolver barbero y credenciales cifradas.';
  END IF;

  SELECT count(*) INTO v_again FROM gcal_restore_claim_connection(300, '2026-04-01 09:01:00-05'::timestamptz)
   WHERE connection_id = 'c3240012-0000-4000-8000-000000000001';
  IF v_again <> 0 THEN RAISE EXCEPTION 'recién reservada no vuelve antes del intervalo.'; END IF;

  v_queued := gcal_enqueue_missing('c3240012-0000-4000-8000-000000000001', 'appointment',
                                    'a3240000-0000-4000-8000-0000000000c1');
  IF v_queued THEN RAISE EXCEPTION 'reencolar lo que ya está pendiente no debe duplicar.'; END IF;
  v_queued := gcal_enqueue_missing('c3240012-0000-4000-8000-000000000001', 'appointment',
                                    'a3240000-0000-4000-8000-0000000000c3');
  IF NOT v_queued THEN RAISE EXCEPTION 'un recurso nuevo sí se encola.'; END IF;
  BEGIN
    PERFORM gcal_enqueue_missing('c3240012-0000-4000-8000-000000000001', 'inventado', gen_random_uuid());
    RAISE EXCEPTION 'un tipo de recurso inválido debe rechazarse.';
  EXCEPTION WHEN raise_exception THEN
    IF SQLERRM NOT LIKE 'gcal_enqueue_missing: tipo de recurso inválido%' THEN RAISE; END IF;
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'aislamiento y chequeo OK · B no ve ni altera la cola de A; la reserva respeta el intervalo'

\echo '=== Issue #324 · todas las comprobaciones pasaron ==='
