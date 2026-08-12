-- Prueba del protocolo de idempotencia (DEC-043, migración aplicada
-- 20260811154100_harden_idempotency_concurrency.sql). Cubre HU-004,
-- CA-004-01, CA-004-02, CA-004-04, CA-004-05, CA-004-06 y RN-IDE-01.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/idempotency_concurrency.sql
--
-- Requisito del arnés: igual que hu001_aislamiento_rls.sql, la conexión
-- debe poder ejecutar `SET ROLE barberia_app`/`barberia_worker`/
-- `barberia_owner` sin pertenecer a esos roles (conectada como superusuario
-- en local/CI). barberia_owner ya llega concedido a barberia_migrator desde
-- la migración de endurecimiento de roles; barberia_app/barberia_worker
-- necesitan el mismo GRANT puntual que ya exige hu001_aislamiento_rls.sql si
-- la conexión no es superusuario.
--
-- Qué NO cubre este archivo: el resultado 'locked' de idempotency_begin
-- (bloqueo consultivo transaccional real) exige DOS conexiones físicas
-- distintas, porque un mismo backend nunca bloquea su propio
-- pg_try_advisory_xact_lock. Eso lo cubre
-- idempotency_concurrency_two_connections.sh (CA-004-03), igual que
-- notification_lease_concurrency.sql/_two_connections.sh se dividen por el
-- mismo motivo.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== idempotency_begin/complete/abort/purge_expired · protocolo (DEC-043) ==='

-- ---------------------------------------------------------------------------
-- CA-004-01 · Camino feliz: proceed -> complete -> replay con la misma huella
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
DECLARE
  v_outcome   text;
  v_completed boolean;
BEGIN
  PERFORM set_config('app.barbershop_id', '11111111-1111-1111-1111-111111111111', true);

  SELECT outcome INTO v_outcome
  FROM idempotency_begin(
    '11111111-1111-1111-1111-111111111111'::uuid,
    'audit-test-happy-path', 'audit_test_operation', repeat('a', 64), 60
  );
  IF v_outcome <> 'proceed' THEN
    RAISE EXCEPTION 'CA-004-01: se esperaba proceed en la primera solicitud, se obtuvo %.', v_outcome;
  END IF;

  SELECT idempotency_complete(
    '11111111-1111-1111-1111-111111111111'::uuid,
    'audit-test-happy-path', 200, 'application/json', '{"id":"ok"}'
  ) INTO v_completed;
  IF NOT v_completed THEN
    RAISE EXCEPTION 'CA-004-01: idempotency_complete no transicionó la fila in_progress.';
  END IF;
END
$$;

COMMIT;

BEGIN;
SET ROLE barberia_app;

DO $$
DECLARE
  v_outcome text;
  v_status  smallint;
  v_body    text;
BEGIN
  PERFORM set_config('app.barbershop_id', '11111111-1111-1111-1111-111111111111', true);

  SELECT outcome, response_status, response_body INTO v_outcome, v_status, v_body
  FROM idempotency_begin(
    '11111111-1111-1111-1111-111111111111'::uuid,
    'audit-test-happy-path', 'audit_test_operation', repeat('a', 64), 60
  );
  IF v_outcome <> 'replay' THEN
    RAISE EXCEPTION 'CA-004-01: se esperaba replay con la misma huella, se obtuvo %.', v_outcome;
  END IF;
  IF v_status <> 200 OR v_body <> '{"id":"ok"}' THEN
    RAISE EXCEPTION 'CA-004-01: replay devolvió una respuesta distinta a la original (% / %).', v_status, v_body;
  END IF;
END
$$;

ROLLBACK;
\echo 'CA-004-01 OK · proceed -> complete -> replay con la misma huella'

-- ---------------------------------------------------------------------------
-- CA-004-02 · Misma clave, huella distinta -> conflict_fingerprint, sin efecto
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
DECLARE
  v_outcome text;
BEGIN
  PERFORM set_config('app.barbershop_id', '11111111-1111-1111-1111-111111111111', true);

  SELECT outcome INTO v_outcome
  FROM idempotency_begin(
    '11111111-1111-1111-1111-111111111111'::uuid,
    'audit-test-happy-path', 'audit_test_operation', repeat('b', 64), 60
  );
  IF v_outcome <> 'conflict_fingerprint' THEN
    RAISE EXCEPTION 'CA-004-02: se esperaba conflict_fingerprint, se obtuvo %.', v_outcome;
  END IF;
END
$$;

ROLLBACK;
\echo 'CA-004-02 OK · huella distinta con la misma clave -> conflict_fingerprint'

-- ---------------------------------------------------------------------------
-- RN-IDE-01 · Misma clave, otra operación -> conflict_operation
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
DECLARE
  v_outcome text;
BEGIN
  PERFORM set_config('app.barbershop_id', '11111111-1111-1111-1111-111111111111', true);

  SELECT outcome INTO v_outcome
  FROM idempotency_begin(
    '11111111-1111-1111-1111-111111111111'::uuid,
    'audit-test-happy-path', 'otra_operacion_distinta', repeat('a', 64), 60
  );
  IF v_outcome <> 'conflict_operation' THEN
    RAISE EXCEPTION 'RN-IDE-01: se esperaba conflict_operation, se obtuvo %.', v_outcome;
  END IF;
END
$$;

ROLLBACK;
\echo 'RN-IDE-01 OK · misma clave para otra operación -> conflict_operation, nunca se ejecuta'

-- ---------------------------------------------------------------------------
-- CA-004-06 · idempotency_abort limpia una fila in_progress y permite
-- reintentar; jamás borra una fila completed (DDL-IDEM-01).
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
DECLARE
  v_outcome text;
  v_aborted boolean;
BEGIN
  PERFORM set_config('app.barbershop_id', '11111111-1111-1111-1111-111111111111', true);

  SELECT outcome INTO v_outcome
  FROM idempotency_begin(
    '11111111-1111-1111-1111-111111111111'::uuid,
    'audit-test-abort', 'audit_test_operation', repeat('c', 64), 60
  );
  IF v_outcome <> 'proceed' THEN
    RAISE EXCEPTION 'CA-004-06: se esperaba proceed, se obtuvo %.', v_outcome;
  END IF;

  SELECT idempotency_abort('11111111-1111-1111-1111-111111111111'::uuid, 'audit-test-abort')
  INTO v_aborted;
  IF NOT v_aborted THEN
    RAISE EXCEPTION 'CA-004-06: idempotency_abort no borró la fila in_progress.';
  END IF;

  SELECT outcome INTO v_outcome
  FROM idempotency_begin(
    '11111111-1111-1111-1111-111111111111'::uuid,
    'audit-test-abort', 'audit_test_operation', repeat('c', 64), 60
  );
  IF v_outcome <> 'proceed' THEN
    RAISE EXCEPTION
      'CA-004-06: tras abortar, un reintento legítimo debía obtener proceed, obtuvo %.', v_outcome;
  END IF;

  -- Se completa la fila y se comprueba que abort YA NO puede borrarla.
  PERFORM idempotency_complete(
    '11111111-1111-1111-1111-111111111111'::uuid,
    'audit-test-abort', 201, 'application/json', '{"id":"c"}'
  );

  SELECT idempotency_abort('11111111-1111-1111-1111-111111111111'::uuid, 'audit-test-abort')
  INTO v_aborted;
  IF v_aborted THEN
    RAISE EXCEPTION 'DDL-IDEM-01: idempotency_abort borró una fila completed; nunca debe hacerlo.';
  END IF;
END
$$;

ROLLBACK;
\echo 'CA-004-06 OK · abort limpia solo in_progress; jamás borra una fila completed (DDL-IDEM-01)'

-- ---------------------------------------------------------------------------
-- CA-004-04 · Aislamiento tenant-aware: la misma clave literal en dos
-- barberías son registros independientes.
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
DECLARE
  v_outcome_a text;
  v_outcome_b text;
BEGIN
  PERFORM set_config('app.barbershop_id', '11111111-1111-1111-1111-111111111111', true);
  SELECT outcome INTO v_outcome_a
  FROM idempotency_begin(
    '11111111-1111-1111-1111-111111111111'::uuid,
    'audit-test-shared-key', 'audit_test_operation', repeat('d', 64), 60
  );

  PERFORM set_config('app.barbershop_id', '22222222-2222-2222-2222-222222222222', true);
  SELECT outcome INTO v_outcome_b
  FROM idempotency_begin(
    '22222222-2222-2222-2222-222222222222'::uuid,
    'audit-test-shared-key', 'audit_test_operation', repeat('e', 64), 60
  );

  IF v_outcome_a <> 'proceed' OR v_outcome_b <> 'proceed' THEN
    RAISE EXCEPTION
      'CA-004-04: la misma clave literal en dos barberías no debía interferir (A=% B=%).',
      v_outcome_a, v_outcome_b;
  END IF;
END
$$;

ROLLBACK;
\echo 'CA-004-04 OK · la misma clave literal en dos barberías son registros independientes'

-- ---------------------------------------------------------------------------
-- CA-004-05 · Una fila vencida no revive una respuesta antigua: se reclama
-- de nuevo como si no existiera.
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
DECLARE
  v_outcome text;
BEGIN
  PERFORM set_config('app.barbershop_id', '11111111-1111-1111-1111-111111111111', true);

  SELECT outcome INTO v_outcome
  FROM idempotency_begin(
    '11111111-1111-1111-1111-111111111111'::uuid,
    'audit-test-expired', 'audit_test_operation', repeat('f', 64), 1
  );
  IF v_outcome <> 'proceed' THEN
    RAISE EXCEPTION 'CA-004-05: se esperaba proceed, se obtuvo %.', v_outcome;
  END IF;

  PERFORM idempotency_complete(
    '11111111-1111-1111-1111-111111111111'::uuid,
    'audit-test-expired', 200, 'application/json', '{"id":"expired"}'
  );
END
$$;

COMMIT;

-- Espera a que venza de verdad, en vez de forzar expires_at con UPDATE:
-- barberia_app ya no tiene UPDATE/DELETE directo sobre la tabla desde la
-- migración de endurecimiento de concurrencia, y retroceder expires_at
-- manualmente puede violar idempotency_record_expires_at_ck (exige
-- expires_at > created_at, y created_at ya quedó fijado por el INSERT de
-- arriba). El TTL de 1 segundo de la llamada anterior hace este único
-- pg_sleep determinista y acotado, no una espera para coordinar con otro
-- proceso (estrategia-pruebas.md §7).
SELECT pg_sleep(1.1);

BEGIN;
SET ROLE barberia_app;

DO $$
DECLARE
  v_outcome text;
BEGIN
  PERFORM set_config('app.barbershop_id', '11111111-1111-1111-1111-111111111111', true);

  SELECT outcome INTO v_outcome
  FROM idempotency_begin(
    '11111111-1111-1111-1111-111111111111'::uuid,
    'audit-test-expired', 'audit_test_operation', repeat('f', 64), 60
  );
  IF v_outcome <> 'proceed' THEN
    RAISE EXCEPTION
      'CA-004-05: una clave vencida debía tratarse como inexistente, se obtuvo %.', v_outcome;
  END IF;
END
$$;

ROLLBACK;
\echo 'CA-004-05 OK · una fila vencida no revive una respuesta antigua'

-- ---------------------------------------------------------------------------
-- idempotency_purge_expired · limpieza acotada, exclusiva del worker
-- ---------------------------------------------------------------------------
SET ROLE barberia_owner;
DELETE FROM idempotency_record
WHERE barbershop_id = '11111111-1111-1111-1111-111111111111'
  AND idempotency_key IN ('audit-test-happy-path', 'audit-test-abort', 'audit-test-expired');
-- created_at y expires_at se fijan ambos en el pasado, con created_at más
-- atrás: idempotency_record_expires_at_ck exige expires_at > created_at, y
-- dentro de una misma sentencia now() es constante, así que "now() - 1s"
-- para ambas columnas violaría el CHECK (quedarían iguales, no expires_at
-- después de created_at).
INSERT INTO idempotency_record (
  barbershop_id, idempotency_key, operation, request_fingerprint, status, created_at, expires_at
)
VALUES (
  '11111111-1111-1111-1111-111111111111', 'audit-test-purge-1', 'audit_test_operation',
  repeat('a', 64), 'in_progress', now() - interval '10 seconds', now() - interval '5 seconds'
);
RESET ROLE;

BEGIN;
SET ROLE barberia_worker;

-- barberia_worker solo tiene EXECUTE sobre esta función, NADA de privilegio
-- directo sobre la tabla (mínimo privilegio real, no solo documentado): la
-- comprobación de ausencia de la fila se hace después, con un rol que sí
-- tiene acceso a la tabla, no aquí.
DO $$
DECLARE
  v_purged integer;
BEGIN
  SELECT idempotency_purge_expired(1000) INTO v_purged;
  IF v_purged < 1 THEN
    RAISE EXCEPTION 'idempotency_purge_expired: se esperaba reciclar al menos 1 fila vencida, recicló %.', v_purged;
  END IF;
END
$$;

SET ROLE barberia_owner;

DO $$
DECLARE
  v_count integer;
BEGIN
  SELECT count(*) INTO v_count FROM idempotency_record
  WHERE barbershop_id = '11111111-1111-1111-1111-111111111111'
    AND idempotency_key = 'audit-test-purge-1';
  IF v_count <> 0 THEN
    RAISE EXCEPTION 'idempotency_purge_expired: la fila vencida seguía presente tras la limpieza.';
  END IF;
END
$$;

RESET ROLE;

-- barberia_app NUNCA debe poder ejecutar la limpieza (exclusiva del worker).
-- SET ROLE es una sentencia de utilidad: no puede aparecer dentro de un
-- bloque PL/pgSQL, así que el cambio de rol queda fuera del DO.
SET ROLE barberia_app;
DO $$
BEGIN
  PERFORM idempotency_purge_expired(10);
  RAISE EXCEPTION 'idempotency_purge_expired: barberia_app pudo ejecutar la limpieza exclusiva del worker.';
EXCEPTION
  WHEN insufficient_privilege THEN
    NULL; -- Esperado: sin EXECUTE concedido a barberia_app.
END
$$;
RESET ROLE;

COMMIT;
\echo 'idempotency_purge_expired OK · recicla filas vencidas, exclusivo de barberia_worker'

\echo '=== idempotency_begin/complete/abort/purge_expired · todas las comprobaciones pasaron ==='
