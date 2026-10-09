-- Pruebas SQL del issue #323 (DEC-099, DEC-102): conexión del barbero con Google
-- Calendar y estado OAuth. Cubren a nivel de base de datos la unicidad de una
-- conexión por barbero, la FK compuesta por tenant, el CHECK que liga las
-- credenciales al estado, el rango del recordatorio, la ausencia de DELETE y el
-- aislamiento RLS entre dos barberías. La orquestación OAuth vive en el servicio
-- Go (internal/modules/googlecalendar), no aquí.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/google_calendar_conexion.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- (estandar-base-datos.md §9.8). Todo se revierte con ROLLBACK.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== Issue #323 · pruebas de la conexión con Google Calendar ==='

-- ---------------------------------------------------------------------------
-- RLS forzada y privilegios: sin DELETE en la conexión, con DELETE en el estado
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_table text;
BEGIN
  FOREACH v_table IN ARRAY ARRAY['google_calendar_connection', 'google_calendar_oauth_state']
  LOOP
    IF NOT (SELECT relrowsecurity AND relforcerowsecurity FROM pg_class
             WHERE relname = v_table AND relnamespace = 'public'::regnamespace) THEN
      RAISE EXCEPTION '% debe tener RLS habilitada Y forzada.', v_table;
    END IF;
    IF (SELECT rolbypassrls FROM pg_roles WHERE rolname = 'barberia_app') THEN
      RAISE EXCEPTION 'barberia_app no debe poder saltarse RLS.';
    END IF;
  END LOOP;

  IF has_table_privilege('barberia_app', 'google_calendar_connection', 'DELETE') THEN
    RAISE EXCEPTION 'la conexión se desconecta, no se elimina: barberia_app no debe tener DELETE.';
  END IF;
  IF NOT has_table_privilege('barberia_app', 'google_calendar_oauth_state', 'DELETE') THEN
    RAISE EXCEPTION 'el estado OAuth vencido se purga: barberia_app necesita DELETE.';
  END IF;
  IF has_table_privilege('barberia_worker', 'google_calendar_connection', 'SELECT')
     OR has_table_privilege('barberia_worker', 'google_calendar_oauth_state', 'SELECT') THEN
    RAISE EXCEPTION 'el worker no recibe privilegios hasta que llegue la cola (issue #324).';
  END IF;
END
$$;
\echo 'RLS y privilegios OK · conexión sin DELETE, estado con DELETE, worker sin acceso'

-- ---------------------------------------------------------------------------
-- Una conexión por barbero, credenciales ligadas al estado y rango del recordatorio
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('c3230001-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111', 'Barbero gcal 1');

DO $$
BEGIN
  INSERT INTO google_calendar_connection
    (barbershop_id, barber_id, status, google_account_email, refresh_token_ciphertext, token_key_id)
  VALUES ('11111111-1111-1111-1111-111111111111', 'c3230001-0000-4000-8000-000000000001',
          'connected', 'barbero@ejemplo.test', '\x0102'::bytea, 'v1');

  -- Una sola conexión por barbero.
  BEGIN
    INSERT INTO google_calendar_connection
      (barbershop_id, barber_id, status, refresh_token_ciphertext, token_key_id)
    VALUES ('11111111-1111-1111-1111-111111111111', 'c3230001-0000-4000-8000-000000000001',
            'connected', '\x0102'::bytea, 'v1');
    RAISE EXCEPTION 'un barbero no debe poder tener dos conexiones.';
  EXCEPTION WHEN unique_violation THEN NULL;
  END;

  -- connected y error exigen credenciales; reauth_required y disconnected las prohíben.
  BEGIN
    UPDATE google_calendar_connection SET refresh_token_ciphertext = NULL, token_key_id = NULL
     WHERE barber_id = 'c3230001-0000-4000-8000-000000000001';
    RAISE EXCEPTION 'connected sin credenciales debe rechazarse.';
  EXCEPTION WHEN check_violation THEN NULL;
  END;
  BEGIN
    UPDATE google_calendar_connection SET status = 'disconnected'
     WHERE barber_id = 'c3230001-0000-4000-8000-000000000001';
    RAISE EXCEPTION 'disconnected con credenciales debe rechazarse.';
  EXCEPTION WHEN check_violation THEN NULL;
  END;
  BEGIN
    UPDATE google_calendar_connection SET status = 'sincronizando'
     WHERE barber_id = 'c3230001-0000-4000-8000-000000000001';
    RAISE EXCEPTION 'un estado fuera de la lista cerrada debe rechazarse.';
  EXCEPTION WHEN check_violation THEN NULL;
  END;

  -- Desconectar de forma coherente (sin credenciales) sí se admite.
  UPDATE google_calendar_connection
     SET status = 'disconnected', refresh_token_ciphertext = NULL, token_key_id = NULL,
         google_account_email = NULL
   WHERE barber_id = 'c3230001-0000-4000-8000-000000000001';

  -- Recordatorio: 0 a 40320 o NULL.
  UPDATE google_calendar_connection SET reminder_minutes = 0
   WHERE barber_id = 'c3230001-0000-4000-8000-000000000001';
  UPDATE google_calendar_connection SET reminder_minutes = 40320
   WHERE barber_id = 'c3230001-0000-4000-8000-000000000001';
  UPDATE google_calendar_connection SET reminder_minutes = NULL
   WHERE barber_id = 'c3230001-0000-4000-8000-000000000001';
  BEGIN
    UPDATE google_calendar_connection SET reminder_minutes = 40321
     WHERE barber_id = 'c3230001-0000-4000-8000-000000000001';
    RAISE EXCEPTION 'un recordatorio mayor de 40320 debe rechazarse.';
  EXCEPTION WHEN check_violation THEN NULL;
  END;
  BEGIN
    UPDATE google_calendar_connection SET reminder_minutes = -1
     WHERE barber_id = 'c3230001-0000-4000-8000-000000000001';
    RAISE EXCEPTION 'un recordatorio negativo debe rechazarse.';
  EXCEPTION WHEN check_violation THEN NULL;
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'restricciones OK · una conexión por barbero, credenciales ligadas al estado, recordatorio acotado'

-- ---------------------------------------------------------------------------
-- La FK compuesta rechaza conectar el barbero de otra barbería
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('c3230002-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111', 'Barbero gcal A');

RESET ROLE;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '22222222-2222-2222-2222-222222222222';

DO $$
BEGIN
  -- La barbería B intenta conectar el barbero de la A: la RLS lo permite (es su
  -- tenant) pero la FK compuesta (barbershop_id, barber_id) lo rechaza.
  BEGIN
    INSERT INTO google_calendar_connection
      (barbershop_id, barber_id, status, refresh_token_ciphertext, token_key_id)
    VALUES ('22222222-2222-2222-2222-222222222222', 'c3230002-0000-4000-8000-000000000001',
            'connected', '\x0102'::bytea, 'v1');
    RAISE EXCEPTION 'la barbería B no debe poder conectar al barbero de la A.';
  EXCEPTION WHEN foreign_key_violation THEN NULL;
  END;

  -- Escribir con un barbershop_id distinto del contexto lo rechaza la RLS.
  BEGIN
    INSERT INTO google_calendar_connection
      (barbershop_id, barber_id, status, refresh_token_ciphertext, token_key_id)
    VALUES ('11111111-1111-1111-1111-111111111111', 'c3230002-0000-4000-8000-000000000001',
            'connected', '\x0102'::bytea, 'v1');
    RAISE EXCEPTION 'la RLS debe rechazar una fila de otra barbería.';
  EXCEPTION WHEN insufficient_privilege THEN NULL;
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'FK compuesta y RLS OK · el barbero de otra barbería no se conecta ni se escribe'

-- ---------------------------------------------------------------------------
-- Aislamiento RLS en lectura y actualización cruzada
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('c3230003-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111', 'Barbero gcal RLS');
INSERT INTO google_calendar_connection
  (barbershop_id, barber_id, status, refresh_token_ciphertext, token_key_id)
VALUES ('11111111-1111-1111-1111-111111111111', 'c3230003-0000-4000-8000-000000000001',
        'connected', '\x0102'::bytea, 'v1');

RESET ROLE;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '22222222-2222-2222-2222-222222222222';

DO $$
DECLARE
  v_rows integer;
BEGIN
  IF EXISTS (SELECT 1 FROM google_calendar_connection
              WHERE barber_id = 'c3230003-0000-4000-8000-000000000001') THEN
    RAISE EXCEPTION 'la barbería B no debe ver la conexión de la A.';
  END IF;
  UPDATE google_calendar_connection SET reminder_minutes = 5
   WHERE barber_id = 'c3230003-0000-4000-8000-000000000001';
  GET DIAGNOSTICS v_rows = ROW_COUNT;
  IF v_rows <> 0 THEN
    RAISE EXCEPTION 'la barbería B no debe poder modificar la conexión de la A (% filas).', v_rows;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'RLS OK · la barbería B no ve ni modifica la conexión de la A'

-- ---------------------------------------------------------------------------
-- Estado OAuth: hash único de 64 caracteres y ligado a barbería, usuario y sesión
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('c3230004-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111', 'Barbero gcal estado');
INSERT INTO staff_session (id, barbershop_id, staff_user_id, token_hash, expires_at) VALUES
  ('5e550323-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111',
   'aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2', repeat('a', 64), now() + interval '1 hour');

DO $$
DECLARE
  v_hash text := repeat('b', 64);
BEGIN
  INSERT INTO google_calendar_oauth_state
    (barbershop_id, barber_id, staff_user_id, session_id, state_hash,
     code_verifier_ciphertext, verifier_key_id, expires_at)
  VALUES ('11111111-1111-1111-1111-111111111111', 'c3230004-0000-4000-8000-000000000001',
          'aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2', '5e550323-0000-4000-8000-000000000001',
          v_hash, '\x0102'::bytea, 'v1', now() + interval '10 minutes');

  BEGIN
    INSERT INTO google_calendar_oauth_state
      (barbershop_id, barber_id, staff_user_id, session_id, state_hash,
       code_verifier_ciphertext, verifier_key_id, expires_at)
    VALUES ('11111111-1111-1111-1111-111111111111', 'c3230004-0000-4000-8000-000000000001',
            'aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2', '5e550323-0000-4000-8000-000000000001',
            v_hash, '\x0102'::bytea, 'v1', now() + interval '10 minutes');
    RAISE EXCEPTION 'un state_hash duplicado debe rechazarse.';
  EXCEPTION WHEN unique_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO google_calendar_oauth_state
      (barbershop_id, barber_id, staff_user_id, session_id, state_hash,
       code_verifier_ciphertext, verifier_key_id, expires_at)
    VALUES ('11111111-1111-1111-1111-111111111111', 'c3230004-0000-4000-8000-000000000001',
            'aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2', '5e550323-0000-4000-8000-000000000001',
            'corto', '\x0102'::bytea, 'v1', now() + interval '10 minutes');
    RAISE EXCEPTION 'un state_hash que no mide 64 caracteres debe rechazarse.';
  EXCEPTION WHEN check_violation THEN NULL;
  END;

  -- Un usuario de otra barbería no puede ser dueño del estado (FK compuesta).
  BEGIN
    INSERT INTO google_calendar_oauth_state
      (barbershop_id, barber_id, staff_user_id, session_id, state_hash,
       code_verifier_ciphertext, verifier_key_id, expires_at)
    VALUES ('11111111-1111-1111-1111-111111111111', 'c3230004-0000-4000-8000-000000000001',
            'bbbbbbb2-bbbb-bbbb-bbbb-bbbbbbbbbbb2', '5e550323-0000-4000-8000-000000000001',
            repeat('c', 64), '\x0102'::bytea, 'v1', now() + interval '10 minutes');
    RAISE EXCEPTION 'un usuario de otra barbería no debe poder ser dueño del state.';
  EXCEPTION WHEN foreign_key_violation THEN NULL;
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'estado OAuth OK · hash único de 64, ligado a barbería, usuario y sesión'

\echo '=== Issue #323 · todas las comprobaciones pasaron ==='
