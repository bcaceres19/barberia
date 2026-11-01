-- Pruebas SQL del issue #463 (HU-130, DEC-142) y #467 (HU-131, DEC-144): cola del
-- aviso al cliente por cancelación o reprogramación del barbero, del recordatorio
-- previo al turno, y evidencia de cada intento. Cubren los
-- privilegios (el worker solo ejecuta funciones; la aplicación solo encola), el RLS
-- entre dos barberías, un solo aviso pendiente por turno, el reclamo con lease y el
-- cierre por CAS (DDL-CON-01), el contexto con estado vigente y la evidencia sin
-- datos de contacto. La orquestación del envío vive en Go
-- (internal/modules/appointmentnotice), no aquí.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/appointment_notice.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app` y
-- `SET ROLE barberia_worker` (superusuario en local/CI efímero). Cada escenario usa
-- un `p_now` explícito en vez de esperar al reloj (estrategia-pruebas.md §7).
-- Todo se revierte con ROLLBACK.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== Issue #463 · pruebas de la cola de avisos al cliente ==='

-- ---------------------------------------------------------------------------
-- Privilegios
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_table text;
  v_fn    text;
BEGIN
  FOREACH v_table IN ARRAY ARRAY['appointment_notice', 'appointment_notice_attempt', 'appointment', 'customer']
  LOOP
    IF has_table_privilege('barberia_worker', v_table, 'SELECT')
       OR has_table_privilege('barberia_worker', v_table, 'INSERT')
       OR has_table_privilege('barberia_worker', v_table, 'UPDATE')
       OR has_table_privilege('barberia_worker', v_table, 'DELETE') THEN
      RAISE EXCEPTION 'barberia_worker no debe tener acceso directo a % (DEC-040).', v_table;
    END IF;
    IF NOT (SELECT relrowsecurity AND relforcerowsecurity FROM pg_class WHERE relname = v_table) THEN
      RAISE EXCEPTION '% debe tener RLS habilitada y forzada.', v_table;
    END IF;
  END LOOP;

  FOREACH v_fn IN ARRAY ARRAY[
    'notice_claim(integer, integer, integer, timestamptz)',
    'notice_context(uuid, uuid)',
    'notice_finish(uuid, uuid, text, text, integer, integer, timestamptz, timestamptz)']
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

  -- La aplicación encola y lee; nada más.
  IF NOT has_table_privilege('barberia_app', 'appointment_notice', 'SELECT')
     OR has_table_privilege('barberia_app', 'appointment_notice', 'DELETE')
     OR has_column_privilege('barberia_app', 'appointment_notice', 'status', 'INSERT')
     OR has_column_privilege('barberia_app', 'appointment_notice', 'status', 'UPDATE')
     OR has_column_privilege('barberia_app', 'appointment_notice', 'claim_token', 'UPDATE')
     OR has_column_privilege('barberia_app', 'appointment_notice', 'covers_state_at', 'INSERT')
     OR NOT has_column_privilege('barberia_app', 'appointment_notice', 'kind', 'UPDATE') THEN
    RAISE EXCEPTION 'barberia_app solo debe encolar (kind/state_at/run_at) y leer.';
  END IF;
  IF has_table_privilege('barberia_app', 'appointment_notice_attempt', 'INSERT')
     OR has_table_privilege('barberia_app', 'appointment_notice_attempt', 'UPDATE')
     OR has_table_privilege('barberia_app', 'appointment_notice_attempt', 'DELETE') THEN
    RAISE EXCEPTION 'la evidencia solo la escribe notice_finish: barberia_app solo lee.';
  END IF;
END
$$;
\echo 'privilegios OK · worker solo funciones, aplicación solo encola y lee'

-- ---------------------------------------------------------------------------
-- Escenario base: una cita en cada barbería
-- ---------------------------------------------------------------------------
BEGIN;
RESET ROLE; -- el arnés conecta como superusuario: crea fixtures sin pasar por RLS

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('b4630001-0000-4000-8000-00000000000a', '11111111-1111-1111-1111-111111111111', 'Barbero aviso A'),
  ('b4630001-0000-4000-8000-00000000000b', '22222222-2222-2222-2222-222222222222', 'Barbero aviso B');
INSERT INTO service (id, barbershop_id, name, duration_minutes, price_amount, price_currency) VALUES
  ('b4630002-0000-4000-8000-00000000000a', '11111111-1111-1111-1111-111111111111', 'Corte aviso A', 30, 20000.00, 'COP'),
  ('b4630002-0000-4000-8000-00000000000b', '22222222-2222-2222-2222-222222222222', 'Corte aviso B', 30, 20000.00, 'COP');
INSERT INTO customer (id, barbershop_id, full_name, email) VALUES
  ('b4630003-0000-4000-8000-00000000000a', '11111111-1111-1111-1111-111111111111', 'Cliente aviso A', 'cliente.a@ejemplo.test'),
  ('b4630003-0000-4000-8000-00000000000b', '22222222-2222-2222-2222-222222222222', 'Cliente aviso B', 'cliente.b@ejemplo.test');
INSERT INTO appointment (
  id, barbershop_id, barber_id, service_id, customer_id, attendee_name,
  starts_at, ends_at, status, origin,
  service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
) VALUES
  ('b4630004-0000-4000-8000-00000000000a', '11111111-1111-1111-1111-111111111111', 'b4630001-0000-4000-8000-00000000000a',
   'b4630002-0000-4000-8000-00000000000a', 'b4630003-0000-4000-8000-00000000000a', 'Cliente aviso A',
   '2099-01-07T10:00:00Z', '2099-01-07T10:30:00Z', 'confirmed', 'manual', 'Corte aviso A', 30, 20000.00, 'COP'),
  ('b4630004-0000-4000-8000-00000000000b', '22222222-2222-2222-2222-222222222222', 'b4630001-0000-4000-8000-00000000000b',
   'b4630002-0000-4000-8000-00000000000b', 'b4630003-0000-4000-8000-00000000000b', 'Cliente aviso B',
   '2099-01-07T10:00:00Z', '2099-01-07T10:30:00Z', 'confirmed', 'manual', 'Corte aviso B', 30, 20000.00, 'COP');

-- ---------------------------------------------------------------------------
-- La aplicación encola; RLS aísla; un solo pendiente por turno que se reemplaza
-- ---------------------------------------------------------------------------
SET ROLE barberia_app;
SELECT set_config('app.barbershop_id', '11111111-1111-1111-1111-111111111111', true);

DO $$
DECLARE
  v_count integer;
  v_kind  text;
BEGIN
  INSERT INTO appointment_notice (barbershop_id, appointment_id, kind, state_at, run_at)
  VALUES ('11111111-1111-1111-1111-111111111111', 'b4630004-0000-4000-8000-00000000000a',
          'rescheduled', '2026-01-05 09:00:00-05', '2026-01-05 09:01:00-05')
  ON CONFLICT (appointment_id, (kind = 'reminder')) WHERE status = 'pending'
  DO UPDATE SET kind = EXCLUDED.kind, state_at = EXCLUDED.state_at, run_at = EXCLUDED.run_at;

  -- Un cambio posterior REEMPLAZA al pendiente (la cancelación sustituye a la reprogramación).
  INSERT INTO appointment_notice (barbershop_id, appointment_id, kind, state_at, run_at)
  VALUES ('11111111-1111-1111-1111-111111111111', 'b4630004-0000-4000-8000-00000000000a',
          'cancelled_by_barber', '2026-01-05 09:00:20-05', '2026-01-05 09:01:20-05')
  ON CONFLICT (appointment_id, (kind = 'reminder')) WHERE status = 'pending'
  DO UPDATE SET kind = EXCLUDED.kind, state_at = EXCLUDED.state_at, run_at = EXCLUDED.run_at;

  SELECT count(*), max(kind) INTO v_count, v_kind FROM appointment_notice;
  IF v_count <> 1 OR v_kind <> 'cancelled_by_barber' THEN
    RAISE EXCEPTION 'debe quedar un solo aviso pendiente, el más reciente (hay %, kind %).', v_count, v_kind;
  END IF;

  -- La aplicación no puede fijar estado ni fabricar un enviado.
  BEGIN
    INSERT INTO appointment_notice (barbershop_id, appointment_id, kind, state_at, status)
    VALUES ('11111111-1111-1111-1111-111111111111', 'b4630004-0000-4000-8000-00000000000a',
            'rescheduled', now(), 'sent');
    RAISE EXCEPTION 'la aplicación no debe poder insertar un aviso con estado.';
  EXCEPTION WHEN insufficient_privilege THEN NULL;
  END;
  BEGIN
    UPDATE appointment_notice SET status = 'sent';
    RAISE EXCEPTION 'la aplicación no debe poder actualizar el estado.';
  EXCEPTION WHEN insufficient_privilege THEN NULL;
  END;

  -- Aislamiento: encolar sobre la barbería B desde el tenant A falla por RLS.
  BEGIN
    INSERT INTO appointment_notice (barbershop_id, appointment_id, kind, state_at)
    VALUES ('22222222-2222-2222-2222-222222222222', 'b4630004-0000-4000-8000-00000000000b',
            'rescheduled', now());
    RAISE EXCEPTION 'encolar en otra barbería debe rechazarse.';
  EXCEPTION WHEN insufficient_privilege THEN NULL;
  END;
END
$$;
\echo 'cola OK · un pendiente por turno, reemplazo por el más reciente y aplicación sin estado'

RESET ROLE;
-- Un aviso pendiente de B, para comprobar el aislamiento de lectura.
INSERT INTO appointment_notice (id, barbershop_id, appointment_id, kind, state_at, run_at)
VALUES ('b4630005-0000-4000-8000-00000000000b', '22222222-2222-2222-2222-222222222222',
        'b4630004-0000-4000-8000-00000000000b', 'rescheduled', '2026-01-05 09:00:00-05', '2026-01-05 09:01:00-05');

SET ROLE barberia_app;
SELECT set_config('app.barbershop_id', '11111111-1111-1111-1111-111111111111', true);
DO $$
BEGIN
  IF (SELECT count(*) FROM appointment_notice) <> 1
     OR EXISTS (SELECT 1 FROM appointment_notice WHERE barbershop_id <> '11111111-1111-1111-1111-111111111111') THEN
    RAISE EXCEPTION 'el tenant A solo debe ver sus avisos.';
  END IF;
END
$$;
\echo 'RLS OK · cada barbería ve solo sus avisos'

-- ---------------------------------------------------------------------------
-- Reclamo con lease, contexto y cierre por CAS
-- ---------------------------------------------------------------------------
SET ROLE barberia_worker;

DO $$
DECLARE
  v_first  record;
  v_second record;
  v_ctx    record;
  v_ok     boolean;
  v_count  integer;
BEGIN
  SELECT count(*) INTO v_count FROM notice_claim(10, 60, 5, '2026-01-05 09:01:00-05'::timestamptz);
  IF v_count <> 1 THEN
    RAISE EXCEPTION 'antes del run_at del aviso A solo vence el de B (reclamados: %).', v_count;
  END IF;

  SELECT * INTO v_first FROM notice_claim(10, 60, 5, '2026-01-05 09:01:30-05'::timestamptz);
  IF v_first IS NULL OR v_first.attempts <> 1 THEN
    RAISE EXCEPTION 'el aviso de A debe reclamarse con su contador en 1.';
  END IF;

  SELECT count(*) INTO v_count FROM notice_claim(10, 60, 5, '2026-01-05 09:01:45-05'::timestamptz);
  IF v_count <> 0 THEN
    RAISE EXCEPTION 'un aviso con lease vigente no debe reclamarse (reclamados: %).', v_count;
  END IF;

  -- Contexto: estado VIGENTE del turno, con el token reclamado.
  SELECT * INTO v_ctx FROM notice_context(v_first.notice_id, v_first.claim_token);
  IF v_ctx IS NULL OR v_ctx.kind <> 'cancelled_by_barber' OR v_ctx.customer_email <> 'cliente.a@ejemplo.test'
     OR v_ctx.service_name <> 'Corte aviso A' OR v_ctx.appointment_status <> 'confirmed'
     OR v_ctx.already_covered THEN
    RAISE EXCEPTION 'el contexto debe describir el turno vigente de A y no estar cubierto.';
  END IF;
  IF EXISTS (SELECT 1 FROM notice_context(v_first.notice_id, gen_random_uuid())) THEN
    RAISE EXCEPTION 'un token ajeno no debe leer el contexto.';
  END IF;

  -- Vencido el lease, se reclama con un token NUEVO y el primero ya no cierra.
  SELECT * INTO v_second FROM notice_claim(10, 60, 5, '2026-01-05 09:05:00-05'::timestamptz)
   WHERE notice_id = v_first.notice_id;
  IF v_second IS NULL OR v_second.claim_token = v_first.claim_token OR v_second.attempts <> 2 THEN
    RAISE EXCEPTION 'con el lease vencido debe reclamarse con otro token y attempts=2.';
  END IF;
  IF notice_finish(v_first.notice_id, v_first.claim_token, 'sent', NULL, NULL, 5,
                   '2026-01-05 09:00:20-05', '2026-01-05 09:05:01-05') THEN
    RAISE EXCEPTION 'el claim_token vencido no debe poder cerrar.';
  END IF;

  -- Un resultado inventado y un envío sin estado cubierto se rechazan.
  BEGIN
    PERFORM notice_finish(v_second.notice_id, v_second.claim_token, 'inventado');
    RAISE EXCEPTION 'un resultado fuera de la lista debe rechazarse.';
  EXCEPTION WHEN raise_exception THEN
    IF SQLERRM NOT LIKE 'notice_finish: resultado inválido%' THEN RAISE; END IF;
  END;
  BEGIN
    PERFORM notice_finish(v_second.notice_id, v_second.claim_token, 'sent');
    RAISE EXCEPTION 'sent sin p_covers_state_at debe rechazarse.';
  EXCEPTION WHEN raise_exception THEN
    IF SQLERRM NOT LIKE 'notice_finish: sent exige%' THEN RAISE; END IF;
  END;

  -- Un fallo transitorio vuelve a pending con espera y deja su intento.
  v_ok := notice_finish(v_second.notice_id, v_second.claim_token, 'retry', 'provider_unavailable', 120, 5,
                        NULL, '2026-01-05 09:05:02-05');
  IF NOT v_ok THEN RAISE EXCEPTION 'el claim_token vigente debe poder cerrar.'; END IF;
  IF notice_finish(v_second.notice_id, v_second.claim_token, 'sent', NULL, NULL, 5, now()) THEN
    RAISE EXCEPTION 'un aviso ya cerrado en esta reclamación no se cierra de nuevo.';
  END IF;
END
$$;
\echo 'lease OK · un solo claim vigente, token nuevo al vencer, contexto vigente y CAS'

-- ---------------------------------------------------------------------------
-- Reintento, envío, intentos agotados y estado ya cubierto
-- ---------------------------------------------------------------------------
RESET ROLE;
DO $$
DECLARE
  v_notice public.appointment_notice%ROWTYPE;
BEGIN
  SELECT * INTO v_notice FROM appointment_notice WHERE appointment_id = 'b4630004-0000-4000-8000-00000000000a';
  IF v_notice.status <> 'pending' OR v_notice.attempts <> 2 OR v_notice.outcome_code <> 'provider_unavailable'
     OR v_notice.run_at <> '2026-01-05 09:07:02-05' OR v_notice.claim_token IS NOT NULL THEN
    RAISE EXCEPTION 'el reintento debe dejar el aviso pending con espera y causa (estado %, intentos %).',
      v_notice.status, v_notice.attempts;
  END IF;
  IF (SELECT count(*) FROM appointment_notice_attempt WHERE notice_id = v_notice.id) <> 1
     OR NOT EXISTS (SELECT 1 FROM appointment_notice_attempt
                     WHERE notice_id = v_notice.id AND attempt_no = 2 AND outcome = 'retry'
                       AND error_code = 'provider_unavailable' AND channel = 'email') THEN
    RAISE EXCEPTION 'el intento 2 debe quedar como retry; el 1 (worker caído) no deja fila.';
  END IF;
END
$$;

SET ROLE barberia_worker;
CREATE TEMP TABLE claimed_a AS
  SELECT * FROM notice_claim(10, 60, 5, '2026-01-05 09:08:00-05'::timestamptz);
CREATE TEMP TABLE ctx_a AS
  SELECT c.notice_id, x.* FROM claimed_a c, LATERAL notice_context(c.notice_id, c.claim_token) x
   WHERE x.customer_email = 'cliente.a@ejemplo.test';
DO $$
DECLARE
  v_count integer;
BEGIN
  SELECT count(*) INTO v_count FROM ctx_a;
  IF v_count <> 1 THEN
    RAISE EXCEPTION 'vencida la espera, el aviso de A debe reclamarse (contextos de A: %).', v_count;
  END IF;
END
$$;
-- Envío: guarda el estado que cubrió.
SELECT notice_finish(c.notice_id, c.claim_token, 'sent', NULL, NULL, 5, x.appointment_state_at, '2026-01-05 09:08:05-05')
  FROM claimed_a c JOIN ctx_a x ON x.notice_id = c.notice_id;

RESET ROLE;
DO $$
DECLARE
  v_notice public.appointment_notice%ROWTYPE;
BEGIN
  SELECT * INTO v_notice FROM appointment_notice WHERE appointment_id = 'b4630004-0000-4000-8000-00000000000a';
  IF v_notice.status <> 'sent' OR v_notice.attempts <> 3 OR v_notice.covers_state_at IS NULL
     OR v_notice.finished_at IS NULL OR v_notice.outcome_code IS NOT NULL THEN
    RAISE EXCEPTION 'el envío debe dejar el aviso sent con el estado cubierto.';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM appointment_notice_attempt
                  WHERE notice_id = v_notice.id AND attempt_no = 3 AND outcome = 'sent' AND error_code IS NULL) THEN
    RAISE EXCEPTION 'el envío debe dejar su intento sent.';
  END IF;
END
$$;
\echo 'envío OK · reintento con espera, sent con el estado cubierto y su evidencia'

-- Intentos agotados: el aviso de B (con 3 reclamos hechos por los barridos anteriores) falla en el quinto.
SET ROLE barberia_worker;
DO $$
DECLARE
  v_c  record;
  v_at timestamptz;
BEGIN
  FOREACH v_at IN ARRAY ARRAY['2026-01-05 10:00:00-05', '2026-01-05 11:00:00-05']::timestamptz[] LOOP
    SELECT * INTO v_c FROM notice_claim(10, 60, 5, v_at) WHERE notice_id = 'b4630005-0000-4000-8000-00000000000b';
    IF v_c IS NULL THEN
      RAISE EXCEPTION 'el aviso de B debe reclamarse en %.', v_at;
    END IF;
    PERFORM notice_finish(v_c.notice_id, v_c.claim_token, 'retry', 'provider_unavailable', 60, 5, NULL, v_at);
  END LOOP;
END
$$;
RESET ROLE;
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM appointment_notice
                  WHERE id = 'b4630005-0000-4000-8000-00000000000b'
                    AND status = 'failed' AND attempts = 5 AND outcome_code = 'provider_unavailable'
                    AND finished_at IS NOT NULL) THEN
    RAISE EXCEPTION 'agotados los 5 intentos el aviso debe quedar failed.';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM appointment_notice_attempt
                  WHERE notice_id = 'b4630005-0000-4000-8000-00000000000b' AND attempt_no = 5 AND outcome = 'failed') THEN
    RAISE EXCEPTION 'el último intento debe quedar como failed.';
  END IF;
END
$$;
\echo 'agotamiento OK · failed al quinto intento'

-- Estado ya cubierto: otro aviso de A con state_at <= covers_state_at se descarta;
-- uno más reciente no.
INSERT INTO appointment_notice (id, barbershop_id, appointment_id, kind, state_at, run_at) VALUES
  ('b4630006-0000-4000-8000-00000000000a', '11111111-1111-1111-1111-111111111111',
   'b4630004-0000-4000-8000-00000000000a', 'rescheduled', '2026-01-05 09:00:10-05', '2026-01-05 13:00:00-05');
SET ROLE barberia_worker;
CREATE TEMP TABLE claimed_covered AS SELECT * FROM notice_claim(10, 60, 5, '2026-01-05 13:00:00-05'::timestamptz);
DO $$
DECLARE
  v_ctx record;
BEGIN
  SELECT x.* INTO v_ctx FROM claimed_covered c, LATERAL notice_context(c.notice_id, c.claim_token) x
   WHERE x.customer_email = 'cliente.a@ejemplo.test';
  IF v_ctx IS NULL OR NOT v_ctx.already_covered THEN
    RAISE EXCEPTION 'un aviso anterior a un envío ya hecho debe verse cubierto.';
  END IF;
END
$$;
SELECT notice_finish(c.notice_id, c.claim_token, 'skipped', 'superseded', NULL, 5, NULL, '2026-01-05 13:00:01-05')
  FROM claimed_covered c WHERE c.notice_id = 'b4630006-0000-4000-8000-00000000000a';
RESET ROLE;
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM appointment_notice WHERE id = 'b4630006-0000-4000-8000-00000000000a'
                  AND status = 'skipped' AND outcome_code = 'superseded' AND finished_at IS NOT NULL)
     OR NOT EXISTS (SELECT 1 FROM appointment_notice_attempt
                     WHERE notice_id = 'b4630006-0000-4000-8000-00000000000a' AND outcome = 'skipped'
                       AND error_code = 'superseded') THEN
    RAISE EXCEPTION 'el aviso cubierto debe quedar skipped/superseded con su intento.';
  END IF;
END
$$;

INSERT INTO appointment_notice (id, barbershop_id, appointment_id, kind, state_at, run_at) VALUES
  ('b4630007-0000-4000-8000-00000000000a', '11111111-1111-1111-1111-111111111111',
   'b4630004-0000-4000-8000-00000000000a', 'cancelled_by_barber', '2100-01-01T00:00:00Z', '2026-01-05 14:00:00-05');
SET ROLE barberia_worker;
DO $$
DECLARE
  v_c   record;
  v_ctx record;
BEGIN
  SELECT * INTO v_c FROM notice_claim(10, 60, 5, '2026-01-05 14:00:00-05'::timestamptz)
   WHERE notice_id = 'b4630007-0000-4000-8000-00000000000a';
  SELECT * INTO v_ctx FROM notice_context(v_c.notice_id, v_c.claim_token);
  IF v_ctx.already_covered THEN
    RAISE EXCEPTION 'un aviso más reciente que lo enviado NO debe verse cubierto.';
  END IF;
  PERFORM notice_finish(v_c.notice_id, v_c.claim_token, 'permanent', 'provider_rejected', NULL, 5, NULL, '2026-01-05 14:00:01-05');
END
$$;
RESET ROLE;
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM appointment_notice WHERE id = 'b4630007-0000-4000-8000-00000000000a'
                  AND status = 'failed' AND outcome_code = 'provider_rejected' AND attempts = 1)
     OR NOT EXISTS (SELECT 1 FROM appointment_notice_attempt
                     WHERE notice_id = 'b4630007-0000-4000-8000-00000000000a' AND attempt_no = 1
                       AND outcome = 'failed' AND error_code = 'provider_rejected') THEN
    RAISE EXCEPTION 'un rechazo permanente debe cerrar failed al primer intento, sin reintento.';
  END IF;
END
$$;
\echo 'cubierto OK · skipped/superseded y rechazo permanente sin reintento'

-- Expiración con intentos agotados: un worker que cae siempre no reclama para siempre.
INSERT INTO appointment_notice (id, barbershop_id, appointment_id, kind, state_at, run_at, status, attempts,
                                claim_token, lease_expires_at)
SELECT 'b4630008-0000-4000-8000-00000000000a', '11111111-1111-1111-1111-111111111111',
       'b4630004-0000-4000-8000-00000000000a', 'rescheduled', now(), now(), 'processing', 5,
       gen_random_uuid(), '2026-01-05 15:00:00-05';
SET ROLE barberia_worker;
SELECT count(*) FROM notice_claim(10, 60, 5, '2026-01-05 16:00:00-05'::timestamptz);
RESET ROLE;
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM appointment_notice WHERE id = 'b4630008-0000-4000-8000-00000000000a'
                  AND status = 'failed' AND outcome_code = 'lease_expired') THEN
    RAISE EXCEPTION 'un lease vencido con los intentos agotados debe cerrarse failed/lease_expired.';
  END IF;
END
$$;
\echo 'lease agotado OK · failed/lease_expired'

-- ---------------------------------------------------------------------------
-- Restricciones y evidencia sin datos de contacto
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  -- Un único pendiente por turno aun para el propietario.
  INSERT INTO appointment_notice (barbershop_id, appointment_id, kind, state_at)
  VALUES ('11111111-1111-1111-1111-111111111111', 'b4630004-0000-4000-8000-00000000000a', 'rescheduled', now());
  BEGIN
    INSERT INTO appointment_notice (barbershop_id, appointment_id, kind, state_at)
    VALUES ('11111111-1111-1111-1111-111111111111', 'b4630004-0000-4000-8000-00000000000a', 'rescheduled', now());
    RAISE EXCEPTION 'dos avisos pendientes del mismo turno deben rechazarse.';
  EXCEPTION WHEN unique_violation THEN NULL;
  END;

  -- Un aviso no puede apuntar a una cita de otra barbería.
  BEGIN
    INSERT INTO appointment_notice (barbershop_id, appointment_id, kind, state_at)
    VALUES ('11111111-1111-1111-1111-111111111111', 'b4630004-0000-4000-8000-00000000000b', 'rescheduled', now());
    RAISE EXCEPTION 'un aviso no puede cruzar tenants.';
  EXCEPTION WHEN foreign_key_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO appointment_notice (barbershop_id, appointment_id, kind, state_at)
    VALUES ('22222222-2222-2222-2222-222222222222', 'b4630004-0000-4000-8000-00000000000b', 'inventado', now());
    RAISE EXCEPTION 'un kind fuera de la lista cerrada debe rechazarse.';
  EXCEPTION WHEN check_violation THEN NULL;
  END;

  -- processing exige reclamación; sent exige el estado cubierto; terminado exige fecha.
  BEGIN
    UPDATE appointment_notice SET status = 'processing'
     WHERE appointment_id = 'b4630004-0000-4000-8000-00000000000a' AND status = 'pending';
    RAISE EXCEPTION 'processing sin claim_token debe rechazarse.';
  EXCEPTION WHEN check_violation THEN NULL;
  END;
  BEGIN
    UPDATE appointment_notice SET status = 'sent', finished_at = now()
     WHERE appointment_id = 'b4630004-0000-4000-8000-00000000000a' AND status = 'pending';
    RAISE EXCEPTION 'sent sin covers_state_at debe rechazarse.';
  EXCEPTION WHEN check_violation THEN NULL;
  END;

  -- La evidencia no admite un destino ni texto libre: error_code es un código corto.
  BEGIN
    INSERT INTO appointment_notice_attempt (barbershop_id, notice_id, attempt_no, channel, outcome, error_code)
    VALUES ('11111111-1111-1111-1111-111111111111', 'b4630006-0000-4000-8000-00000000000a', 9, 'email', 'failed',
            'cliente.a@ejemplo.test');
    RAISE EXCEPTION 'un correo en error_code debe rechazarse.';
  EXCEPTION WHEN check_violation THEN NULL;
  END;
  BEGIN
    INSERT INTO appointment_notice_attempt (barbershop_id, notice_id, attempt_no, channel, outcome)
    VALUES ('11111111-1111-1111-1111-111111111111', 'b4630006-0000-4000-8000-00000000000a', 9, 'whatsapp', 'sent');
    RAISE EXCEPTION 'WhatsApp no está en este tramo (DEC-027).';
  EXCEPTION WHEN check_violation THEN NULL;
  END;

  -- Ninguna columna de la evidencia puede contener contacto: se fija el conjunto.
  IF (SELECT array_agg(column_name::text ORDER BY column_name) FROM information_schema.columns
       WHERE table_schema = 'public' AND table_name = 'appointment_notice_attempt')
     IS DISTINCT FROM ARRAY['attempt_no', 'barbershop_id', 'channel', 'created_at', 'error_code', 'id',
                            'notice_id', 'outcome']::text[] THEN
    RAISE EXCEPTION 'appointment_notice_attempt no debe ganar columnas sin revisar RN-DAT-02.';
  END IF;
END
$$;
\echo 'restricciones OK · un pendiente por turno, sin cruce de tenants y evidencia sin contacto'

-- ---------------------------------------------------------------------------
-- Issue #467 (HU-131, DEC-144): el recordatorio comparte la cola
-- ---------------------------------------------------------------------------
RESET ROLE;
INSERT INTO appointment (
  id, barbershop_id, barber_id, service_id, customer_id, attendee_name,
  starts_at, ends_at, status, origin,
  service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
) VALUES
  ('b4630004-0000-4000-8000-0000000000c1', '11111111-1111-1111-1111-111111111111', 'b4630001-0000-4000-8000-00000000000a',
   'b4630002-0000-4000-8000-00000000000a', 'b4630003-0000-4000-8000-00000000000a', 'Cliente recordatorio',
   '2099-01-09T10:00:00Z', '2099-01-09T10:30:00Z', 'confirmed', 'manual', 'Corte aviso A', 30, 20000.00, 'COP');

SET ROLE barberia_app;
SELECT set_config('app.barbershop_id', '11111111-1111-1111-1111-111111111111', true);
DO $$
DECLARE
  v_count integer;
BEGIN
  -- Un recordatorio y un aviso de cambio pueden esperar a la vez: ninguno sustituye al otro.
  INSERT INTO appointment_notice (barbershop_id, appointment_id, kind, state_at, run_at)
  VALUES ('11111111-1111-1111-1111-111111111111', 'b4630004-0000-4000-8000-0000000000c1',
          'reminder', '2099-01-09T10:00:00Z', '2099-01-09T09:30:00Z');
  INSERT INTO appointment_notice (barbershop_id, appointment_id, kind, state_at, run_at)
  VALUES ('11111111-1111-1111-1111-111111111111', 'b4630004-0000-4000-8000-0000000000c1',
          'rescheduled', '2026-01-05 09:00:00-05', '2026-01-05 09:01:00-05')
  ON CONFLICT (appointment_id, (kind = 'reminder')) WHERE status = 'pending'
  DO UPDATE SET kind = EXCLUDED.kind, state_at = EXCLUDED.state_at, run_at = EXCLUDED.run_at;

  -- Reprogramar regenera el recordatorio: se actualiza el pendiente, no se duplica.
  INSERT INTO appointment_notice (barbershop_id, appointment_id, kind, state_at, run_at)
  VALUES ('11111111-1111-1111-1111-111111111111', 'b4630004-0000-4000-8000-0000000000c1',
          'reminder', '2099-01-09T13:00:00Z', '2099-01-09T12:30:00Z')
  ON CONFLICT (appointment_id, (kind = 'reminder')) WHERE status = 'pending'
  DO UPDATE SET state_at = EXCLUDED.state_at, run_at = EXCLUDED.run_at;

  SELECT count(*) INTO v_count FROM appointment_notice
   WHERE appointment_id = 'b4630004-0000-4000-8000-0000000000c1' AND status = 'pending';
  IF v_count <> 2 THEN
    RAISE EXCEPTION 'debe haber un recordatorio y un aviso de cambio pendientes (hay %).', v_count;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM appointment_notice WHERE appointment_id = 'b4630004-0000-4000-8000-0000000000c1'
                  AND kind = 'reminder' AND state_at = '2099-01-09T13:00:00Z' AND run_at = '2099-01-09T12:30:00Z') THEN
    RAISE EXCEPTION 'el recordatorio regenerado debe llevar el instante nuevo del turno.';
  END IF;

  -- Dos recordatorios pendientes del mismo turno se rechazan.
  BEGIN
    INSERT INTO appointment_notice (barbershop_id, appointment_id, kind, state_at, run_at)
    VALUES ('11111111-1111-1111-1111-111111111111', 'b4630004-0000-4000-8000-0000000000c1',
            'reminder', '2099-01-09T14:00:00Z', '2099-01-09T13:30:00Z');
    RAISE EXCEPTION 'un segundo recordatorio pendiente debe rechazarse.';
  EXCEPTION WHEN unique_violation THEN NULL;
  END;

  -- Una clase desconocida sigue rechazándose.
  BEGIN
    INSERT INTO appointment_notice (barbershop_id, appointment_id, kind, state_at)
    VALUES ('11111111-1111-1111-1111-111111111111', 'b4630004-0000-4000-8000-00000000000a', 'whatsapp_reminder', now());
    RAISE EXCEPTION 'una clase de aviso desconocida debe rechazarse.';
  EXCEPTION WHEN check_violation THEN NULL;
  END;
END
$$;

-- Contexto y cobertura: la cobertura es por clase y, para el recordatorio, exacta.
RESET ROLE;
UPDATE appointment_notice SET status = 'sent', attempts = 1, covers_state_at = '2100-01-01T00:00:00Z', finished_at = now()
 WHERE appointment_id = 'b4630004-0000-4000-8000-0000000000c1' AND kind = 'rescheduled';
DELETE FROM appointment_notice WHERE appointment_id = 'b4630004-0000-4000-8000-0000000000c1' AND kind = 'reminder';
INSERT INTO appointment_notice (id, barbershop_id, appointment_id, kind, state_at, run_at) VALUES
  ('b4630011-0000-4000-8000-0000000000c1', '11111111-1111-1111-1111-111111111111',
   'b4630004-0000-4000-8000-0000000000c1', 'reminder', '2099-01-09T10:00:00Z', '2099-01-09T09:30:00Z');
SET ROLE barberia_worker;
DO $$
DECLARE
  v_c   record;
  v_ctx record;
BEGIN
  SELECT * INTO v_c FROM notice_claim(100, 60, 5, '2099-01-09T09:31:00Z'::timestamptz)
   WHERE notice_id = 'b4630011-0000-4000-8000-0000000000c1';
  SELECT * INTO v_ctx FROM notice_context(v_c.notice_id, v_c.claim_token);
  IF v_ctx.notice_state_at <> '2099-01-09T10:00:00Z' OR v_ctx.starts_at <> '2099-01-09T10:00:00Z' THEN
    RAISE EXCEPTION 'el contexto debe devolver el instante que cubre el recordatorio.';
  END IF;
  IF v_ctx.already_covered THEN
    RAISE EXCEPTION 'un aviso de cambio enviado NO cubre un recordatorio (clase distinta).';
  END IF;
  PERFORM notice_finish(v_c.notice_id, v_c.claim_token, 'sent', NULL, NULL, 5, v_ctx.notice_state_at, '2099-01-09T09:31:01Z');
END
$$;
RESET ROLE;
INSERT INTO appointment_notice (id, barbershop_id, appointment_id, kind, state_at, run_at) VALUES
  ('b4630012-0000-4000-8000-0000000000c1', '11111111-1111-1111-1111-111111111111',
   'b4630004-0000-4000-8000-0000000000c1', 'reminder', '2099-01-09T10:00:00Z', '2099-01-09T09:30:00Z');
SET ROLE barberia_worker;
DO $$
DECLARE
  v_c record;
BEGIN
  SELECT * INTO v_c FROM notice_claim(100, 60, 5, '2099-01-09T09:32:00Z'::timestamptz)
   WHERE notice_id = 'b4630012-0000-4000-8000-0000000000c1';
  IF NOT (SELECT already_covered FROM notice_context(v_c.notice_id, v_c.claim_token)) THEN
    RAISE EXCEPTION 'un recordatorio del mismo instante ya enviado debe verse cubierto.';
  END IF;
  PERFORM notice_finish(v_c.notice_id, v_c.claim_token, 'skipped', 'superseded', NULL, 5, NULL, '2099-01-09T09:32:01Z');
END
$$;
RESET ROLE;
INSERT INTO appointment_notice (id, barbershop_id, appointment_id, kind, state_at, run_at) VALUES
  ('b4630013-0000-4000-8000-0000000000c1', '11111111-1111-1111-1111-111111111111',
   'b4630004-0000-4000-8000-0000000000c1', 'reminder', '2099-01-09T08:00:00Z', '2099-01-09T07:30:00Z');
SET ROLE barberia_worker;
DO $$
DECLARE
  v_c record;
BEGIN
  SELECT * INTO v_c FROM notice_claim(100, 60, 5, '2099-01-09T09:33:00Z'::timestamptz)
   WHERE notice_id = 'b4630013-0000-4000-8000-0000000000c1';
  IF (SELECT already_covered FROM notice_context(v_c.notice_id, v_c.claim_token)) THEN
    RAISE EXCEPTION 'un recordatorio de otro instante NO debe verse cubierto, aunque sea anterior (cobertura exacta).';
  END IF;
END
$$;
RESET ROLE;
\echo 'recordatorio OK · convive con el aviso de cambio, se regenera sin duplicar y su cobertura es exacta'

ROLLBACK;
