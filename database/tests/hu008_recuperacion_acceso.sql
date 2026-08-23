-- Pruebas SQL de HU-008: staff_recovery_code y sus cuatro funciones
-- SECURITY DEFINER (auth_recovery_request/verify/current_credential/
-- change_password). Cubren DDL-AUT-01 (sin GRANT directo a barberia_app),
-- DEC-064 (código de un solo uso, vencimiento, intentos, cooldown/límite de
-- reenvío, token de reinicio) y CA-008-05 (cambio de contraseña + revocación
-- de TODAS las sesiones en la misma operación atómica).
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu005_credenciales_sesiones.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu007_reto_telefonico.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/hu008_recuperacion_acceso.sql
--
-- Mismo requisito de arnés que hu007_defensa_abuso.sql: la conexión debe
-- poder ejecutar `SET ROLE barberia_app` y `SET ROLE barberia_worker`. Cada
-- escenario que usa las cuentas reales (duena.a, dueno.b) se envuelve en
-- BEGIN...ROLLBACK a propósito, igual que hu007_defensa_abuso.sql: así
-- nunca deja un código de recuperación real persistido que compita por
-- cooldown con las pruebas Go de internal/modules/auth/postgres o
-- cmd/api, que corren en un job de CI distinto con su propio PostgreSQL
-- efímero (sin solaparse con este archivo de todos modos).
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== HU-008 · pruebas de recuperación de acceso con código ==='

-- ---------------------------------------------------------------------------
-- DDL-AUT-01 · barberia_app no puede leer/escribir staff_recovery_code
-- directamente
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
BEGIN
  BEGIN
    PERFORM * FROM staff_recovery_code LIMIT 1;
    RAISE EXCEPTION 'DDL-AUT-01: barberia_app pudo leer staff_recovery_code directamente.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL; -- Esperado.
  END;

  BEGIN
    DELETE FROM staff_recovery_code;
    RAISE EXCEPTION 'DDL-AUT-01: barberia_app pudo borrar staff_recovery_code directamente.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL; -- Esperado.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'DDL-AUT-01 OK · staff_recovery_code no admite DML directo de barberia_app'

-- ---------------------------------------------------------------------------
-- DDL-AUT-01 · auth_recovery_purge_expired es exclusiva de barberia_worker
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
BEGIN
  BEGIN
    PERFORM auth_recovery_purge_expired(10);
    RAISE EXCEPTION 'DDL-AUT-01: barberia_app pudo ejecutar auth_recovery_purge_expired.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL; -- Esperado.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'DDL-AUT-01 OK · purga exclusiva de barberia_worker'

BEGIN;
SET ROLE barberia_worker;
DO $$
BEGIN
  PERFORM auth_recovery_purge_expired(10);
END
$$;
RESET ROLE;
ROLLBACK;
\echo 'OK · barberia_worker sí puede purgar staff_recovery_code'

-- ---------------------------------------------------------------------------
-- CA-008-01 · auth_recovery_request no enumera: cuenta inexistente,
-- teléfono no verificado y cuenta válida producen la forma de resultado
-- correcta en cada caso
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
DECLARE
  v_code_hash text := encode(sha256('hu008-sql-test-request-code'::bytea), 'hex');
  v_accepted  boolean;
  v_phone     text;
  v_email     text;
BEGIN
  -- Cuenta inexistente: no acepta, sin phone/email.
  SELECT accepted, phone, email INTO v_accepted, v_phone, v_email
  FROM auth_recovery_request('no-existe@ejemplo.test', v_code_hash, 900, 5, 60, 3600, 3);
  IF v_accepted OR v_phone IS NOT NULL OR v_email IS NOT NULL THEN
    RAISE EXCEPTION 'CA-008-01: se aceptó una solicitud para una cuenta inexistente.';
  END IF;

  -- Teléfono no verificado (barbero.a, testdata/hu007_reto_telefonico.sql):
  -- no acepta.
  SELECT accepted, phone, email INTO v_accepted, v_phone, v_email
  FROM auth_recovery_request('barbero.a@ejemplo.test', v_code_hash, 900, 5, 60, 3600, 3);
  IF v_accepted OR v_phone IS NOT NULL OR v_email IS NOT NULL THEN
    RAISE EXCEPTION 'CA-008-01: se aceptó una solicitud para un teléfono no verificado.';
  END IF;

  -- Cuenta válida con teléfono verificado (duena.a): sí acepta y devuelve
  -- phone/email.
  SELECT accepted, phone, email INTO v_accepted, v_phone, v_email
  FROM auth_recovery_request('duena.a@ejemplo.test', v_code_hash, 900, 5, 60, 3600, 3);
  IF NOT v_accepted OR v_phone IS DISTINCT FROM '+573000000001' OR v_email IS DISTINCT FROM 'duena.a@ejemplo.test' THEN
    RAISE EXCEPTION 'CA-008-01: no se aceptó una solicitud válida (accepted=%, phone=%, email=%).', v_accepted, v_phone, v_email;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-008-01 OK · auth_recovery_request no enumera cuentas'

-- ---------------------------------------------------------------------------
-- CA-008-07 · cooldown de reenvío y un solo código vigente por usuario
-- ---------------------------------------------------------------------------
-- Nota: las lecturas directas de staff_recovery_code que verifican "un solo
-- código vigente" corren DESPUÉS de RESET ROLE (superusuario de la
-- conexión, que hace bypass de RLS): barberia_app no tiene ningún GRANT
-- sobre esta tabla (DDL-AUT-01, ya probado arriba), así que una lectura de
-- verificación con ese rol fallaría por privilegios en vez de por RLS.
BEGIN;
SET ROLE barberia_app;

DO $$
DECLARE
  v_first_hash text := encode(sha256('hu008-sql-test-resend-first'::bytea), 'hex');
  v_accepted   boolean;
BEGIN
  SELECT accepted INTO v_accepted
  FROM auth_recovery_request('dueno.b@ejemplo.test', v_first_hash, 900, 5, 60, 3600, 3);
  IF NOT v_accepted THEN
    RAISE EXCEPTION 'CA-008-07: preparación del escenario de cooldown falló (no aceptado).';
  END IF;
END
$$;

DO $$
DECLARE
  v_second_hash text := encode(sha256('hu008-sql-test-resend-second'::bytea), 'hex');
  v_accepted    boolean;
  v_phone       text;
BEGIN
  -- Reenvío inmediato: cooldown de 60s lo rechaza.
  SELECT accepted, phone INTO v_accepted, v_phone
  FROM auth_recovery_request('dueno.b@ejemplo.test', v_second_hash, 900, 5, 60, 3600, 3);
  IF v_accepted OR v_phone IS NOT NULL THEN
    RAISE EXCEPTION 'CA-008-07: el reenvío inmediato debía rechazarse por cooldown.';
  END IF;
END
$$;

RESET ROLE;
DO $$
DECLARE
  v_active integer;
BEGIN
  -- Un solo código vigente: el rechazo por cooldown no creó una segunda fila
  -- activa (índice único parcial idx_staff_recovery_code_active_per_user).
  SELECT count(*) INTO v_active
  FROM staff_recovery_code
  WHERE staff_user_id = (SELECT id FROM staff_user WHERE email = 'dueno.b@ejemplo.test')
    AND verified_at IS NULL AND invalidated_at IS NULL;
  IF v_active <> 1 THEN
    RAISE EXCEPTION 'CA-008-07: se esperaba exactamente 1 código vigente, había %.', v_active;
  END IF;
END
$$;

-- `now()` es fijo durante toda la transacción en PostgreSQL (hora de inicio
-- de la transacción), así que un cooldown_seconds menor no simula el paso
-- del tiempo aquí dentro; en su lugar, se retrasa manualmente created_at
-- del código vigente (como superusuario, tras RESET ROLE) lo suficiente
-- para limpiar el cooldown de 60s sin salir de la ventana de 3600s del
-- límite de reenvío. El paso del tiempo real ya se prueba, determinista,
-- con reloj inyectado en las pruebas Go (auth.RecoveryService).
RESET ROLE;
DO $$
BEGIN
  UPDATE staff_recovery_code
  SET created_at = created_at - interval '65 seconds'
  WHERE staff_user_id = (SELECT id FROM staff_user WHERE email = 'dueno.b@ejemplo.test')
    AND verified_at IS NULL AND invalidated_at IS NULL;
END
$$;

SET ROLE barberia_app;
DO $$
DECLARE
  v_second_hash text := encode(sha256('hu008-sql-test-resend-second'::bytea), 'hex');
  v_accepted    boolean;
BEGIN
  -- Con el cooldown ya vencido (created_at retrasado arriba), la solicitud
  -- siguiente SÍ se acepta e invalida atómicamente la anterior.
  SELECT accepted INTO v_accepted
  FROM auth_recovery_request('dueno.b@ejemplo.test', v_second_hash, 900, 5, 60, 3600, 2);
  IF NOT v_accepted THEN
    RAISE EXCEPTION 'CA-008-07: el reenvío tras el cooldown debía aceptarse.';
  END IF;
END
$$;

RESET ROLE;
DO $$
DECLARE
  v_active integer;
BEGIN
  SELECT count(*) INTO v_active
  FROM staff_recovery_code
  WHERE staff_user_id = (SELECT id FROM staff_user WHERE email = 'dueno.b@ejemplo.test')
    AND verified_at IS NULL AND invalidated_at IS NULL;
  IF v_active <> 1 THEN
    RAISE EXCEPTION 'CA-008-07: el reenvío debía dejar exactamente 1 código vigente (el nuevo), había %.', v_active;
  END IF;

  -- Retrasa el nuevo código igual que arriba, para el tercer intento.
  UPDATE staff_recovery_code
  SET created_at = created_at - interval '65 seconds'
  WHERE staff_user_id = (SELECT id FROM staff_user WHERE email = 'dueno.b@ejemplo.test')
    AND verified_at IS NULL AND invalidated_at IS NULL;
END
$$;

SET ROLE barberia_app;
DO $$
DECLARE
  v_accepted boolean;
BEGIN
  -- Límite propio de reenvío: con ResendMaxPerWindow=2, un tercer intento
  -- dentro de la ventana se rechaza aunque el cooldown ya haya pasado (los
  -- dos códigos anteriores, aunque uno ya esté invalidado, siguen contando
  -- para la ventana de 3600s).
  SELECT accepted INTO v_accepted
  FROM auth_recovery_request('dueno.b@ejemplo.test', encode(sha256('hu008-sql-test-resend-third'::bytea), 'hex'), 900, 5, 60, 3600, 2);
  IF v_accepted THEN
    RAISE EXCEPTION 'CA-008-07: el tercer reenvío debía rechazarse por el límite propio de la ventana.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-008-07 OK · cooldown, un solo código vigente y límite de reenvío'

-- ---------------------------------------------------------------------------
-- CA-008-02/03 · intentos, agotamiento y vencimiento en auth_recovery_verify
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
DECLARE
  v_code_hash   text := encode(sha256('hu008-sql-test-verify-code'::bytea), 'hex');
  v_wrong_hash  text := encode(sha256('hu008-sql-test-verify-wrong'::bytea), 'hex');
  v_token_hash  text := encode(sha256('hu008-sql-test-verify-token'::bytea), 'hex');
  v_accepted    boolean;
BEGIN
  SELECT accepted INTO v_accepted
  FROM auth_recovery_request('duena.a@ejemplo.test', v_code_hash, 900, 3, 60, 3600, 3);
  IF NOT v_accepted THEN
    RAISE EXCEPTION 'CA-008-02: preparación del escenario de verificación falló (no aceptado).';
  END IF;

  -- Código incorrecto: no avanza, incrementa attempt_count.
  PERFORM ok FROM auth_recovery_verify('duena.a@ejemplo.test', v_wrong_hash, v_token_hash, 300);
END
$$;

DO $$
DECLARE
  v_wrong_hash text := encode(sha256('hu008-sql-test-verify-wrong'::bytea), 'hex');
  v_token_hash text := encode(sha256('hu008-sql-test-verify-token'::bytea), 'hex');
  v_ok         boolean;
  v_phone      text;
  v_email      text;
BEGIN
  SELECT ok, phone, email INTO v_ok, v_phone, v_email
  FROM auth_recovery_verify('duena.a@ejemplo.test', v_wrong_hash, v_token_hash, 300);
  IF v_ok OR v_phone IS NOT NULL OR v_email IS NOT NULL THEN
    RAISE EXCEPTION 'CA-008-02: un código incorrecto no debía verificar.';
  END IF;
END
$$;

RESET ROLE;
DO $$
DECLARE
  v_attempts smallint;
BEGIN
  SELECT attempt_count INTO v_attempts
  FROM staff_recovery_code
  WHERE staff_user_id = (SELECT id FROM staff_user WHERE email = 'duena.a@ejemplo.test')
    AND verified_at IS NULL AND invalidated_at IS NULL;
  IF v_attempts <> 2 THEN
    RAISE EXCEPTION 'CA-008-02: se esperaba attempt_count=2 tras dos códigos incorrectos, obtuvo %.', v_attempts;
  END IF;
END
$$;

SET ROLE barberia_app;
DO $$
DECLARE
  v_wrong_hash text := encode(sha256('hu008-sql-test-verify-wrong'::bytea), 'hex');
  v_code_hash  text := encode(sha256('hu008-sql-test-verify-code'::bytea), 'hex');
  v_token_hash text := encode(sha256('hu008-sql-test-verify-token'::bytea), 'hex');
  v_ok         boolean;
BEGIN
  -- Agotar el tercer intento (max_attempts=3): el código queda invalidado.
  PERFORM ok FROM auth_recovery_verify('duena.a@ejemplo.test', v_wrong_hash, v_token_hash, 300);

  -- El código correcto, ya agotado/invalidado, tampoco verifica (CA-008-02:
  -- vencido/usado/inválido no revive).
  SELECT ok INTO v_ok
  FROM auth_recovery_verify('duena.a@ejemplo.test', v_code_hash, v_token_hash, 300);
  IF v_ok THEN
    RAISE EXCEPTION 'CA-008-02: un código agotado no debía revivir con el valor correcto.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-008-02/03 OK · intentos, agotamiento e invalidez terminal'

-- ---------------------------------------------------------------------------
-- CA-008-05 · verificación exitosa, credencial vigente y cambio de
-- contraseña con revocación de TODAS las sesiones en la misma operación
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

-- Nota: igual que en el bloque CA-008-07, las lecturas directas de
-- staff_credential/staff_session (para confirmar el efecto del cambio de
-- contraseña) corren tras RESET ROLE: barberia_app solo alcanza esas tablas
-- a través de las funciones SECURITY DEFINER aprobadas, nunca por SQL
-- directo en este flujo previo a resolver tenant.
DO $$
DECLARE
  v_code_hash  text := encode(sha256('hu008-sql-test-success-code'::bytea), 'hex');
  v_token_hash text := encode(sha256('hu008-sql-test-success-token'::bytea), 'hex');
  v_accepted   boolean;
  v_ok         boolean;
  v_found      boolean;
  v_password   text;
  v_algorithm  text;
BEGIN
  SELECT accepted INTO v_accepted
  FROM auth_recovery_request('duena.a@ejemplo.test', v_code_hash, 900, 5, 60, 3600, 3);
  IF NOT v_accepted THEN
    RAISE EXCEPTION 'CA-008-05: preparación del escenario de éxito falló (no aceptado).';
  END IF;

  -- Verificar con el código correcto: éxito, devuelve phone/email reales.
  SELECT ok INTO v_ok
  FROM auth_recovery_verify('duena.a@ejemplo.test', v_code_hash, v_token_hash, 300);
  IF NOT v_ok THEN
    RAISE EXCEPTION 'CA-008-05: el código correcto debía verificar.';
  END IF;

  -- auth_recovery_current_credential expone el hash vigente sin consumir el
  -- token. Alias de tabla obligatorio: "found" colisiona con la variable
  -- especial FOUND de PL/pgSQL si se referencia sin calificar.
  SELECT cc.found, cc.password_hash, cc.password_algorithm INTO v_found, v_password, v_algorithm
  FROM auth_recovery_current_credential('duena.a@ejemplo.test', v_token_hash) AS cc;
  IF NOT v_found OR v_password IS NULL OR v_algorithm <> 'argon2id' THEN
    RAISE EXCEPTION 'CA-008-05: auth_recovery_current_credential debía devolver la credencial vigente.';
  END IF;

  -- El token sigue sin consumirse: repetir current_credential funciona
  -- igual.
  SELECT cc.found INTO v_found
  FROM auth_recovery_current_credential('duena.a@ejemplo.test', v_token_hash) AS cc;
  IF NOT v_found THEN
    RAISE EXCEPTION 'CA-008-05: current_credential no debía consumir el token.';
  END IF;
END
$$;

DO $$
DECLARE
  v_token_hash text := encode(sha256('hu008-sql-test-success-token'::bytea), 'hex');
  v_new_hash   text := 'fixture-nuevo-hash-de-prueba-sql-suficientemente-largo';
  v_changed    boolean;
BEGIN
  -- Cambiar la contraseña: éxito, revoca TODAS las sesiones activas de
  -- duena.a (fixture: una sesión vigente por
  -- testdata/hu005_credenciales_sesiones.sql).
  SELECT auth_recovery_change_password('duena.a@ejemplo.test', v_token_hash, v_new_hash) INTO v_changed;
  IF NOT v_changed THEN
    RAISE EXCEPTION 'CA-008-05: el cambio de contraseña con un token válido debía tener éxito.';
  END IF;
END
$$;

RESET ROLE;
DO $$
DECLARE
  v_password    text;
  v_active_sess integer;
BEGIN
  SELECT password_hash INTO v_password
  FROM staff_credential
  WHERE staff_user_id = (SELECT id FROM staff_user WHERE email = 'duena.a@ejemplo.test');
  IF v_password IS DISTINCT FROM 'fixture-nuevo-hash-de-prueba-sql-suficientemente-largo' THEN
    RAISE EXCEPTION 'CA-008-05: staff_credential.password_hash debía quedar en el nuevo valor.';
  END IF;

  SELECT count(*) INTO v_active_sess
  FROM staff_session
  WHERE staff_user_id = (SELECT id FROM staff_user WHERE email = 'duena.a@ejemplo.test')
    AND revoked_at IS NULL;
  IF v_active_sess <> 0 THEN
    RAISE EXCEPTION 'CA-008-05: se esperaba 0 sesiones activas tras el cambio de contraseña, había %.', v_active_sess;
  END IF;
END
$$;

SET ROLE barberia_app;
DO $$
DECLARE
  v_token_hash text := encode(sha256('hu008-sql-test-success-token'::bytea), 'hex');
  v_changed    boolean;
BEGIN
  -- Reutilizar el token ya consumido falla (cualquier hash nuevo).
  SELECT auth_recovery_change_password('duena.a@ejemplo.test', v_token_hash, 'fixture-hash-reuso-suficientemente-largo') INTO v_changed;
  IF v_changed THEN
    RAISE EXCEPTION 'CA-008-05: reutilizar un token ya consumido no debía tener éxito.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-008-05 OK · verificación, lectura de credencial y cambio de contraseña con revocación total de sesiones'
