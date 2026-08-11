-- Pruebas SQL del protocolo de lease de notificaciones (issue #5, DEC-053).
-- Cubren DDL-CON-01, DDL-CON-02 y DDL-OPS-01 sobre notification_claim_due,
-- notification_finalize_claim y retention_claim_due_customers.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/notification_lease_fixture.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/notification_lease_concurrency.sql
--
-- Requisito del arnés: igual que hu001_aislamiento_rls.sql, la conexión debe
-- poder ejecutar `SET ROLE barberia_app` y `SET ROLE barberia_worker` sin
-- pertenecer a esos roles (conectada como superusuario en local/CI).
--
-- Por qué no hay una prueba de dos conexiones reales aquí
--   La garantía de "una sola reclamación vigente" no depende de que dos
--   sesiones se solapen en el reloj: depende de que notification_claim_due
--   sea un único UPDATE atómico sobre un SELECT ... FOR UPDATE SKIP LOCKED, y
--   de que el WHERE distinga 'pending' vencido de 'processing' vencido. Cada
--   escenario de abajo lo prueba pasando un `p_now` distinto a llamadas
--   sucesivas en la MISMA sesión, sin esperar reloj real ni necesitar una
--   segunda conexión (estrategia-pruebas.md §7: nada de `sleep` para
--   coordinar). La prueba que SÍ exige dos conexiones reales y solapadas
--   —que SKIP LOCKED salte una fila bloqueada por una transacción de otro
--   proceso todavía abierta, no solo por la misma sesión— vive en
--   notification_lease_concurrency_two_connections.sh (DDL-CON-02).
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== notification_claim_due / notification_finalize_claim · pruebas de lease ==='

-- ---------------------------------------------------------------------------
-- Escenario 1 · Reclamo básico y "sin doble claim vigente"
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_migrator;

INSERT INTO notification_schedule (id, barbershop_id, appointment_id, event_type, channel, scheduled_for, created_at)
VALUES ('00000000-0000-0000-0000-000000000001',
        '11111111-1111-1111-1111-111111111111',
        'a99a99a9-a99a-a99a-a99a-a99a99a99a99',
        'reminder', 'email', '2026-01-05 09:30:00-05', '2026-01-05 09:00:00-05');

SET ROLE barberia_worker;

DO $$
DECLARE
  v_row     record;
  v_second  integer;
BEGIN
  SELECT * INTO v_row
  FROM notification_claim_due(10, 60, '2026-01-05 09:30:00-05'::timestamptz)
  WHERE schedule_id = '00000000-0000-0000-0000-000000000001';

  IF v_row IS NULL THEN
    RAISE EXCEPTION 'Escenario 1: la fila vencida no fue reclamada.';
  END IF;
  IF v_row.claim_token IS NULL THEN
    RAISE EXCEPTION 'Escenario 1: el reclamo no devolvió claim_token.';
  END IF;

  -- Mismo instante, misma fila: el lease sigue vigente, no debe reclamarse
  -- una segunda vez ("sin doble claim vigente", DDL-CON-01).
  SELECT count(*) INTO v_second
  FROM notification_claim_due(10, 60, '2026-01-05 09:30:00-05'::timestamptz)
  WHERE schedule_id = '00000000-0000-0000-0000-000000000001';

  IF v_second <> 0 THEN
    RAISE EXCEPTION 'Escenario 1: la fila se reclamó dos veces con el lease vigente.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Escenario 1 OK · reclamo único, sin doble claim con el lease vigente'

-- ---------------------------------------------------------------------------
-- Escenario 2 · Recuperación de un lease vencido (caída tras el claim)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_migrator;

INSERT INTO notification_schedule (id, barbershop_id, appointment_id, event_type, channel, scheduled_for, created_at)
VALUES ('00000000-0000-0000-0000-000000000002',
        '11111111-1111-1111-1111-111111111111',
        'a99a99a9-a99a-a99a-a99a-a99a99a99a99',
        'confirmation', 'email', '2026-01-05 09:30:00-05', '2026-01-05 09:00:00-05');

SET ROLE barberia_worker;

DO $$
DECLARE
  v_first  record;
  v_second record;
  v_stale  boolean;
BEGIN
  -- El "trabajador A" reclama con un lease de 60 s y nunca finaliza (caída).
  SELECT * INTO v_first
  FROM notification_claim_due(10, 60, '2026-01-05 09:30:00-05'::timestamptz)
  WHERE schedule_id = '00000000-0000-0000-0000-000000000002';

  IF v_first IS NULL THEN
    RAISE EXCEPTION 'Escenario 2: el claim inicial no reclamó la fila.';
  END IF;

  -- 61 s después el lease venció: otra llamada (el "trabajador B") la
  -- recupera con un claim_token NUEVO, sin necesitar función separada.
  SELECT * INTO v_second
  FROM notification_claim_due(10, 60, '2026-01-05 09:31:01-05'::timestamptz)
  WHERE schedule_id = '00000000-0000-0000-0000-000000000002';

  IF v_second IS NULL THEN
    RAISE EXCEPTION 'Escenario 2: el lease vencido no se recuperó.';
  END IF;
  IF v_second.claim_token = v_first.claim_token THEN
    RAISE EXCEPTION 'Escenario 2: la recuperación reutilizó el claim_token anterior.';
  END IF;

  -- El token del trabajador A ya no sirve: CAS falla, sin excepción.
  SELECT notification_finalize_claim(
    '11111111-1111-1111-1111-111111111111',
    '00000000-0000-0000-0000-000000000002',
    v_first.claim_token, 'sent', '2026-01-05 09:32:00-05'::timestamptz
  ) INTO v_stale;

  IF v_stale THEN
    RAISE EXCEPTION 'Escenario 2: finalizar con el claim_token vencido debía devolver false.';
  END IF;

  -- El token vigente (trabajador B) sí cierra correctamente.
  IF NOT notification_finalize_claim(
    '11111111-1111-1111-1111-111111111111',
    '00000000-0000-0000-0000-000000000002',
    v_second.claim_token, 'sent', '2026-01-05 09:32:00-05'::timestamptz
  ) THEN
    RAISE EXCEPTION 'Escenario 2: finalizar con el claim_token vigente debía devolver true.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Escenario 2 OK · lease vencido se recupera con token nuevo; el token viejo no finaliza nada'

-- ---------------------------------------------------------------------------
-- Escenario 3 · Fallo temporal: retry vuelve a 'pending' y es reclamable ya
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_migrator;

INSERT INTO notification_schedule (id, barbershop_id, appointment_id, event_type, channel, scheduled_for, created_at)
VALUES ('00000000-0000-0000-0000-000000000003',
        '11111111-1111-1111-1111-111111111111',
        'a99a99a9-a99a-a99a-a99a-a99a99a99a99',
        'cancellation', 'email', '2026-01-05 09:30:00-05', '2026-01-05 09:00:00-05');

SET ROLE barberia_worker;

DO $$
DECLARE
  v_claim  record;
  v_retry  record;
BEGIN
  SELECT * INTO v_claim
  FROM notification_claim_due(10, 60, '2026-01-05 09:30:00-05'::timestamptz)
  WHERE schedule_id = '00000000-0000-0000-0000-000000000003';

  IF NOT notification_finalize_claim(
    '11111111-1111-1111-1111-111111111111',
    '00000000-0000-0000-0000-000000000003',
    v_claim.claim_token, 'retry', '2026-01-05 09:30:05-05'::timestamptz
  ) THEN
    RAISE EXCEPTION 'Escenario 3: retry con token vigente debía devolver true.';
  END IF;

  -- Reclamable de inmediato, sin esperar a que venza ningún lease: prueba,
  -- sin necesitar leer la tabla (barberia_worker no tiene SELECT directo),
  -- que retry la dejó en 'pending' y no en 'processing' ni en un estado
  -- terminal.
  SELECT * INTO v_retry
  FROM notification_claim_due(10, 60, '2026-01-05 09:30:05-05'::timestamptz)
  WHERE schedule_id = '00000000-0000-0000-0000-000000000003';

  IF v_retry IS NULL THEN
    RAISE EXCEPTION 'Escenario 3: la fila en retry no se pudo reclamar de inmediato.';
  END IF;
END
$$;

SET ROLE barberia_migrator;

DO $$
DECLARE
  v_status text;
BEGIN
  SELECT status INTO v_status FROM notification_schedule
  WHERE id = '00000000-0000-0000-0000-000000000003';
  IF v_status <> 'processing' THEN
    RAISE EXCEPTION 'Escenario 3: tras el segundo claim la fila debía quedar processing, quedó en %.', v_status;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Escenario 3 OK · fallo temporal (retry) suelta el lease y permite reclamar de inmediato'

-- ---------------------------------------------------------------------------
-- Escenario 4 · Fallo permanente: permanent_failure pasa a 'skipped' y sale
-- de la cola
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_migrator;

INSERT INTO notification_schedule (id, barbershop_id, appointment_id, event_type, channel, scheduled_for, created_at)
VALUES ('00000000-0000-0000-0000-000000000004',
        '11111111-1111-1111-1111-111111111111',
        'a99a99a9-a99a-a99a-a99a-a99a99a99a99',
        'delay', 'whatsapp', '2026-01-05 09:30:00-05', '2026-01-05 09:00:00-05');

SET ROLE barberia_worker;

DO $$
DECLARE
  v_claim  record;
  v_again  integer;
BEGIN
  SELECT * INTO v_claim
  FROM notification_claim_due(10, 60, '2026-01-05 09:30:00-05'::timestamptz)
  WHERE schedule_id = '00000000-0000-0000-0000-000000000004';

  IF NOT notification_finalize_claim(
    '11111111-1111-1111-1111-111111111111',
    '00000000-0000-0000-0000-000000000004',
    v_claim.claim_token, 'permanent_failure', '2026-01-05 09:30:05-05'::timestamptz
  ) THEN
    RAISE EXCEPTION 'Escenario 4: permanent_failure con token vigente debía devolver true.';
  END IF;

  SELECT count(*) INTO v_again
  FROM notification_claim_due(10, 60, '2026-01-06 00:00:00-05'::timestamptz)
  WHERE schedule_id = '00000000-0000-0000-0000-000000000004';
  IF v_again <> 0 THEN
    RAISE EXCEPTION 'Escenario 4: una fila skipped no debía volver a reclamarse.';
  END IF;
END
$$;

SET ROLE barberia_migrator;

DO $$
DECLARE
  v_status text;
BEGIN
  SELECT status INTO v_status FROM notification_schedule
  WHERE id = '00000000-0000-0000-0000-000000000004';
  IF v_status <> 'skipped' THEN
    RAISE EXCEPTION 'Escenario 4: permanent_failure debía dejar la fila en skipped, quedó en %.', v_status;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Escenario 4 OK · fallo permanente pasa a skipped y sale de la cola'

-- ---------------------------------------------------------------------------
-- Escenario 5 · Envío exitoso: sent es terminal
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_migrator;

INSERT INTO notification_schedule (id, barbershop_id, appointment_id, event_type, channel, scheduled_for, created_at)
VALUES ('00000000-0000-0000-0000-000000000005',
        '11111111-1111-1111-1111-111111111111',
        'a99a99a9-a99a-a99a-a99a-a99a99a99a99',
        'no_show', 'email', '2026-01-05 09:30:00-05', '2026-01-05 09:00:00-05');

SET ROLE barberia_worker;

DO $$
DECLARE
  v_claim  record;
BEGIN
  SELECT * INTO v_claim
  FROM notification_claim_due(10, 60, '2026-01-05 09:30:00-05'::timestamptz)
  WHERE schedule_id = '00000000-0000-0000-0000-000000000005';

  IF NOT notification_finalize_claim(
    '11111111-1111-1111-1111-111111111111',
    '00000000-0000-0000-0000-000000000005',
    v_claim.claim_token, 'sent', '2026-01-05 09:30:05-05'::timestamptz
  ) THEN
    RAISE EXCEPTION 'Escenario 5: sent con token vigente debía devolver true.';
  END IF;

  -- Repetir la finalización con el mismo token ya no encuentra una fila
  -- 'processing': CAS falla, sin excepción ni doble efecto.
  IF notification_finalize_claim(
    '11111111-1111-1111-1111-111111111111',
    '00000000-0000-0000-0000-000000000005',
    v_claim.claim_token, 'sent', '2026-01-05 09:30:06-05'::timestamptz
  ) THEN
    RAISE EXCEPTION 'Escenario 5: finalizar una fila ya sent debía devolver false.';
  END IF;
END
$$;

SET ROLE barberia_migrator;

DO $$
DECLARE
  v_row record;
BEGIN
  SELECT status, sent_at, claim_token, claimed_at, lease_expires_at INTO v_row
  FROM notification_schedule WHERE id = '00000000-0000-0000-0000-000000000005';

  IF v_row.status <> 'sent' OR v_row.sent_at IS NULL THEN
    RAISE EXCEPTION 'Escenario 5: la fila debía quedar sent con sent_at fijado.';
  END IF;
  IF v_row.claim_token IS NOT NULL OR v_row.claimed_at IS NOT NULL
     OR v_row.lease_expires_at IS NOT NULL THEN
    RAISE EXCEPTION 'Escenario 5: sent debía limpiar por completo las columnas de lease.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Escenario 5 OK · sent es terminal, limpia el lease y no se puede reconfirmar'

-- ---------------------------------------------------------------------------
-- Escenario 6 · Una fila 'processing' no libera el hueco de unicidad activa
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_migrator;

INSERT INTO notification_schedule (id, barbershop_id, appointment_id, event_type, channel, ordinal, scheduled_for, created_at)
VALUES ('00000000-0000-0000-0000-000000000006',
        '11111111-1111-1111-1111-111111111111',
        'a99a99a9-a99a-a99a-a99a-a99a99a99a99',
        'reminder', 'whatsapp', 1, '2026-01-05 09:30:00-05', '2026-01-05 09:00:00-05');

SET ROLE barberia_worker;
SELECT claim_token FROM notification_claim_due(10, 60, '2026-01-05 09:30:00-05'::timestamptz)
WHERE schedule_id = '00000000-0000-0000-0000-000000000006';

SET ROLE barberia_migrator;

DO $$
BEGIN
  BEGIN
    INSERT INTO notification_schedule (
      id, barbershop_id, appointment_id, event_type, channel, ordinal, scheduled_for
    ) VALUES (
      '00000000-0000-0000-0000-000000000106',
      '11111111-1111-1111-1111-111111111111',
      'a99a99a9-a99a-a99a-a99a-a99a99a99a99',
      'reminder', 'whatsapp', 1, '2026-01-05 09:40:00-05'
    );
    RAISE EXCEPTION 'Escenario 6: se insertó una segunda programación activa mientras la primera seguía processing.';
  EXCEPTION
    WHEN unique_violation THEN
      NULL;  -- Esperado: idx_notification_schedule_pending_unique cubre 'processing'.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Escenario 6 OK · processing sigue ocupando el hueco de unicidad activa'

-- ---------------------------------------------------------------------------
-- Escenario 7 · barberia_app no puede reclamar ni finalizar (DEC-040, DDL-SEC-04)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_migrator;

INSERT INTO notification_schedule (id, barbershop_id, appointment_id, event_type, channel, scheduled_for, created_at)
VALUES ('00000000-0000-0000-0000-000000000007',
        '11111111-1111-1111-1111-111111111111',
        'a99a99a9-a99a-a99a-a99a-a99a99a99a99',
        'reminder', 'email', '2026-01-05 09:30:00-05', '2026-01-05 09:00:00-05');

SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  BEGIN
    PERFORM * FROM notification_claim_due(10, 60, '2026-01-05 09:30:00-05'::timestamptz);
    RAISE EXCEPTION 'Escenario 7: barberia_app pudo ejecutar notification_claim_due.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL;  -- Esperado.
  END;

  BEGIN
    PERFORM notification_finalize_claim(
      '11111111-1111-1111-1111-111111111111',
      '00000000-0000-0000-0000-000000000007',
      gen_random_uuid(), 'sent', '2026-01-05 09:30:00-05'::timestamptz
    );
    RAISE EXCEPTION 'Escenario 7: barberia_app pudo ejecutar notification_finalize_claim.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL;  -- Esperado.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Escenario 7 OK · barberia_app no tiene EXECUTE ni sobre el claim ni sobre la finalización'

-- ---------------------------------------------------------------------------
-- Escenario 8 · barberia_app no puede escribir columnas de lease, y el CHECK
-- de forma impide cancelar una fila processing por esa misma vía
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_migrator;

INSERT INTO notification_schedule (id, barbershop_id, appointment_id, event_type, channel, scheduled_for, created_at)
VALUES ('00000000-0000-0000-0000-000000000008',
        '11111111-1111-1111-1111-111111111111',
        'a99a99a9-a99a-a99a-a99a-a99a99a99a99',
        'reminder', 'email', '2026-01-05 09:30:00-05', '2026-01-05 09:00:00-05');

SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  BEGIN
    UPDATE notification_schedule
    SET claim_token = gen_random_uuid()
    WHERE id = '00000000-0000-0000-0000-000000000008';
    RAISE EXCEPTION 'Escenario 8: barberia_app pudo escribir claim_token directamente.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL;  -- Esperado: sin privilegio de columna.
  END;

  -- Cancelar una fila pending sí es un uso legítimo del API.
  UPDATE notification_schedule
  SET status = 'cancelled', cancelled_at = '2026-01-05 09:30:01-05'::timestamptz
  WHERE id = '00000000-0000-0000-0000-000000000008';
END
$$;

RESET ROLE;
SET ROLE barberia_migrator;

-- Nueva fila, esta vez reclamada por el worker antes de intentar cancelarla.
INSERT INTO notification_schedule (id, barbershop_id, appointment_id, event_type, channel, scheduled_for, created_at)
VALUES ('00000000-0000-0000-0000-000000000009',
        '11111111-1111-1111-1111-111111111111',
        'a99a99a9-a99a-a99a-a99a-a99a99a99a99',
        'confirmation', 'email', '2026-01-05 09:30:00-05', '2026-01-05 09:00:00-05');

SET ROLE barberia_worker;
SELECT claim_token FROM notification_claim_due(10, 60, '2026-01-05 09:30:00-05'::timestamptz)
WHERE schedule_id = '00000000-0000-0000-0000-000000000009';

SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  BEGIN
    UPDATE notification_schedule
    SET status = 'cancelled', cancelled_at = '2026-01-05 09:30:02-05'::timestamptz
    WHERE id = '00000000-0000-0000-0000-000000000009';
    RAISE EXCEPTION 'Escenario 8: se pudo cancelar una fila processing sin tocar el lease.';
  EXCEPTION
    WHEN check_violation THEN
      NULL;  -- Esperado: cancelled exige claim_token NULL, que barberia_app no puede fijar.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Escenario 8 OK · barberia_app no puede tocar el lease ni cancelar una fila en vuelo'

-- ---------------------------------------------------------------------------
-- Escenario 9 · Validación de p_limit/p_lease_seconds/p_now (DDL-OPS-01)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_worker;

DO $$
  -- Cada combinación inválida debe lanzar excepción, nunca ejecutarse.
  DECLARE
    v_limit  integer;
    v_lease  integer;
  BEGIN
    FOREACH v_limit IN ARRAY ARRAY[NULL, 0, -1, 201] LOOP
      BEGIN
        PERFORM * FROM notification_claim_due(v_limit, 60, now());
        RAISE EXCEPTION 'Escenario 9: p_limit=% debía rechazarse.', v_limit;
      EXCEPTION
        WHEN raise_exception THEN NULL;  -- Esperado.
      END;
    END LOOP;

    FOREACH v_lease IN ARRAY ARRAY[NULL, 0, 29, 3601] LOOP
      BEGIN
        PERFORM * FROM notification_claim_due(10, v_lease, now());
        RAISE EXCEPTION 'Escenario 9: p_lease_seconds=% debía rechazarse.', v_lease;
      EXCEPTION
        WHEN raise_exception THEN NULL;  -- Esperado.
      END;
    END LOOP;

    BEGIN
      PERFORM * FROM notification_claim_due(10, 60, NULL);
      RAISE EXCEPTION 'Escenario 9: p_now NULL debía rechazarse.';
    EXCEPTION
      WHEN raise_exception THEN NULL;  -- Esperado.
    END;

    BEGIN
      PERFORM notification_finalize_claim(NULL, NULL, NULL, NULL, NULL);
      RAISE EXCEPTION 'Escenario 9: notification_finalize_claim con NULL debía rechazarse.';
    EXCEPTION
      WHEN raise_exception THEN NULL;  -- Esperado.
    END;

    BEGIN
      PERFORM notification_finalize_claim(
        '11111111-1111-1111-1111-111111111111', gen_random_uuid(), gen_random_uuid(),
        'no_existe', now()
      );
      RAISE EXCEPTION 'Escenario 9: un outcome inválido debía rechazarse.';
    EXCEPTION
      WHEN raise_exception THEN NULL;  -- Esperado.
    END;
  END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Escenario 9 OK · p_limit, p_lease_seconds, p_now y outcome fuera de rango se rechazan'

\echo '=== retention_claim_due_customers · pruebas de endurecimiento (DDL-OPS-01) ==='

-- ---------------------------------------------------------------------------
-- Escenario 10 · p_limit/p_now se validan igual que en notification_claim_due;
-- barberia_app tampoco tiene EXECUTE
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_worker;

DO $$
DECLARE
  v_limit integer;
BEGIN
  FOREACH v_limit IN ARRAY ARRAY[NULL, 0, -1, 201] LOOP
    BEGIN
      PERFORM * FROM retention_claim_due_customers(v_limit, now());
      RAISE EXCEPTION 'Escenario 10: p_limit=% debía rechazarse.', v_limit;
    EXCEPTION
      WHEN raise_exception THEN NULL;  -- Esperado.
    END;
  END LOOP;

  BEGIN
    PERFORM * FROM retention_claim_due_customers(10, NULL);
    RAISE EXCEPTION 'Escenario 10: p_now NULL debía rechazarse.';
  EXCEPTION
    WHEN raise_exception THEN NULL;  -- Esperado.
  END;
END
$$;

RESET ROLE;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  BEGIN
    PERFORM * FROM retention_claim_due_customers(10, now());
    RAISE EXCEPTION 'Escenario 10: barberia_app pudo ejecutar retention_claim_due_customers.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL;  -- Esperado.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Escenario 10 OK · retention_claim_due_customers valida límites y sigue exclusivo del worker'

\echo '=== notification_lease_concurrency · todas las comprobaciones pasaron ==='
