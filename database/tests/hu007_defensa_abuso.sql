-- Pruebas SQL de HU-007: login_throttle, auth_phone_challenge y las cinco
-- funciones SECURITY DEFINER (login_throttle_register_attempt/purge_expired,
-- auth_phone_challenge_request/verify/purge_expired). Cubren DDL-AUT-01
-- (sin GRANT directo a barberia_app), DEC-061/CT-005 (la sexta solicitud
-- escala, no la quinta) y DEC-062 (reto telefónico: tres condiciones para
-- aceptar, ip_hash ata el reto a su IP, verificar limpia el escalamiento).
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu005_credenciales_sesiones.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu007_reto_telefonico.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/hu007_defensa_abuso.sql
--
-- Mismo requisito de arnés que hu001_aislamiento_rls.sql: la conexión debe
-- poder ejecutar `SET ROLE barberia_app` y `SET ROLE barberia_worker`.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== HU-007 · pruebas de defensa escalonada contra abuso ==='

-- ---------------------------------------------------------------------------
-- DDL-AUT-01 · barberia_app no puede leer/escribir login_throttle ni
-- auth_phone_challenge directamente
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
BEGIN
  BEGIN
    PERFORM * FROM login_throttle LIMIT 1;
    RAISE EXCEPTION 'DDL-AUT-01: barberia_app pudo leer login_throttle directamente.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL; -- Esperado.
  END;

  BEGIN
    PERFORM * FROM auth_phone_challenge LIMIT 1;
    RAISE EXCEPTION 'DDL-AUT-01: barberia_app pudo leer auth_phone_challenge directamente.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL; -- Esperado.
  END;

  BEGIN
    DELETE FROM login_throttle;
    RAISE EXCEPTION 'DDL-AUT-01: barberia_app pudo borrar login_throttle directamente.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL; -- Esperado.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'DDL-AUT-01 OK · login_throttle/auth_phone_challenge no admiten DML directo de barberia_app'

-- ---------------------------------------------------------------------------
-- DDL-AUT-01 · las funciones de purga son exclusivas de barberia_worker
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
BEGIN
  BEGIN
    PERFORM login_throttle_purge_expired(10);
    RAISE EXCEPTION 'DDL-AUT-01: barberia_app pudo ejecutar login_throttle_purge_expired.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL; -- Esperado.
  END;

  BEGIN
    PERFORM auth_phone_challenge_purge_expired(10);
    RAISE EXCEPTION 'DDL-AUT-01: barberia_app pudo ejecutar auth_phone_challenge_purge_expired.';
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
  PERFORM login_throttle_purge_expired(10);
  PERFORM auth_phone_challenge_purge_expired(10);
END
$$;
RESET ROLE;
ROLLBACK;
\echo 'OK · barberia_worker sí puede purgar ambas tablas'

-- ---------------------------------------------------------------------------
-- DEC-061/CT-005 · la sexta solicitud escala, no la quinta
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
DECLARE
  v_ip_hash text := encode(sha256('hu007-sql-test-sixth-escalates'::bytea), 'hex');
  v_count   integer;
  v_esc     boolean;
BEGIN
  FOR i IN 1..5 LOOP
    SELECT attempt_count, escalated INTO v_count, v_esc
    FROM login_throttle_register_attempt(v_ip_hash, 900, 86400, 5, 172800);
    IF v_count <> i THEN
      RAISE EXCEPTION 'DEC-061: intento % esperaba attempt_count=%, obtuvo %', i, i, v_count;
    END IF;
    IF v_esc THEN
      RAISE EXCEPTION 'DEC-061/CT-005: el intento % escaló antes de superar el umbral.', i;
    END IF;
  END LOOP;

  SELECT attempt_count, escalated INTO v_count, v_esc
  FROM login_throttle_register_attempt(v_ip_hash, 900, 86400, 5, 172800);
  IF v_count <> 6 OR NOT v_esc THEN
    RAISE EXCEPTION 'DEC-061/CT-005: el sexto intento debía escalar (count=6, escalated=true), obtuvo count=% escalated=%', v_count, v_esc;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'DEC-061/CT-005 OK · la sexta solicitud escala, no la quinta'

-- ---------------------------------------------------------------------------
-- DEC-062 · auth_phone_challenge_request exige las tres condiciones
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
DECLARE
  v_ip_escalated  text := encode(sha256('hu007-sql-test-escalated-ip'::bytea), 'hex');
  v_ip_fresh      text := encode(sha256('hu007-sql-test-fresh-ip'::bytea), 'hex');
  v_code_hash     text := encode(sha256('hu007-sql-test-code'::bytea), 'hex');
  v_accepted      boolean;
  v_phone         text;
BEGIN
  -- Escala v_ip_escalated (6 intentos).
  FOR i IN 1..6 LOOP
    PERFORM login_throttle_register_attempt(v_ip_escalated, 900, 86400, 5, 172800);
  END LOOP;

  -- Teléfono no verificado (barbero.a, testdata/hu007_reto_telefonico.sql) +
  -- IP escalada: NO se acepta.
  SELECT accepted, phone INTO v_accepted, v_phone
  FROM auth_phone_challenge_request('barbero.a@ejemplo.test', v_ip_escalated, v_code_hash, 300, 900, 3, 60);
  IF v_accepted OR v_phone IS NOT NULL THEN
    RAISE EXCEPTION 'DEC-062: se aceptó un reto para un teléfono no verificado.';
  END IF;

  -- Teléfono verificado (duena.a) + IP SIN escalar: NO se acepta.
  SELECT accepted, phone INTO v_accepted, v_phone
  FROM auth_phone_challenge_request('duena.a@ejemplo.test', v_ip_fresh, v_code_hash, 300, 900, 3, 60);
  IF v_accepted OR v_phone IS NOT NULL THEN
    RAISE EXCEPTION 'DEC-062: se aceptó un reto para una IP no escalada.';
  END IF;

  -- Cuenta inexistente + IP escalada: NO se acepta.
  SELECT accepted, phone INTO v_accepted, v_phone
  FROM auth_phone_challenge_request('no-existe@ejemplo.test', v_ip_escalated, v_code_hash, 300, 900, 3, 60);
  IF v_accepted OR v_phone IS NOT NULL THEN
    RAISE EXCEPTION 'DEC-062: se aceptó un reto para una cuenta inexistente.';
  END IF;

  -- Las tres condiciones se cumplen: SÍ se acepta y devuelve el teléfono.
  SELECT accepted, phone INTO v_accepted, v_phone
  FROM auth_phone_challenge_request('duena.a@ejemplo.test', v_ip_escalated, v_code_hash, 300, 900, 3, 60);
  IF NOT v_accepted OR v_phone IS DISTINCT FROM '+573000000001' THEN
    RAISE EXCEPTION 'DEC-062: no se aceptó un reto válido (accepted=%, phone=%).', v_accepted, v_phone;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'DEC-062 OK · auth_phone_challenge_request exige cuenta+teléfono verificado+IP escalada'

-- ---------------------------------------------------------------------------
-- DEC-062 · verificar limpia el escalamiento; el código está atado a su IP
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
DECLARE
  v_ip          text := encode(sha256('hu007-sql-test-verify-ip'::bytea), 'hex');
  v_other_ip    text := encode(sha256('hu007-sql-test-verify-other-ip'::bytea), 'hex');
  v_code_hash   text := encode(sha256('hu007-sql-test-verify-code'::bytea), 'hex');
  v_wrong_hash  text := encode(sha256('hu007-sql-test-verify-wrong'::bytea), 'hex');
  v_accepted    boolean;
  v_ok          boolean;
BEGIN
  FOR i IN 1..6 LOOP
    PERFORM login_throttle_register_attempt(v_ip, 900, 86400, 5, 172800);
  END LOOP;

  SELECT accepted INTO v_accepted
  FROM auth_phone_challenge_request('dueno.b@ejemplo.test', v_ip, v_code_hash, 300, 900, 3, 60);
  IF NOT v_accepted THEN
    RAISE EXCEPTION 'DEC-062: preparación del escenario de verificación falló (no aceptado).';
  END IF;

  -- Código correcto desde OTRA IP: falla (atado a la IP que lo pidió).
  SELECT auth_phone_challenge_verify('dueno.b@ejemplo.test', v_other_ip, v_code_hash) INTO v_ok;
  IF v_ok THEN
    RAISE EXCEPTION 'DEC-062: un código correcto verificó desde una IP distinta a la que lo pidió.';
  END IF;

  -- Código incorrecto desde la IP correcta: falla.
  SELECT auth_phone_challenge_verify('dueno.b@ejemplo.test', v_ip, v_wrong_hash) INTO v_ok;
  IF v_ok THEN
    RAISE EXCEPTION 'DEC-062: un código incorrecto verificó con éxito.';
  END IF;

  -- Código correcto desde la IP correcta: éxito, y limpia el escalamiento.
  SELECT auth_phone_challenge_verify('dueno.b@ejemplo.test', v_ip, v_code_hash) INTO v_ok;
  IF NOT v_ok THEN
    RAISE EXCEPTION 'DEC-062: el código correcto desde la IP correcta debía verificar.';
  END IF;

  -- barberia_app no puede leer login_throttle directamente (DDL-AUT-01, ya
  -- probado arriba): se verifica la limpieza del escalamiento registrando
  -- UN intento más. Si verify no hubiera limpiado escalated_until/
  -- attempt_count, este intento arrastraría el conteo alto anterior y
  -- seguiría escalado; si sí limpió, es un conteo fresco (1) y no escala.
  -- (No se usa auth_phone_challenge_request aquí a propósito: su propio
  -- cooldown de reenvío de 60s confundiría el resultado con la solicitud
  -- ya hecha arriba para esta misma IP.)
  DECLARE
    v_fresh_count integer;
    v_fresh_esc   boolean;
  BEGIN
    SELECT attempt_count, escalated INTO v_fresh_count, v_fresh_esc
    FROM login_throttle_register_attempt(v_ip, 900, 86400, 5, 172800);
    IF v_fresh_count <> 1 OR v_fresh_esc THEN
      RAISE EXCEPTION 'DEC-062: la IP no quedó limpia tras verificar (attempt_count=%, escalated=%), se esperaba (1, false).',
        v_fresh_count, v_fresh_esc;
    END IF;
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'DEC-062 OK · verificar limpia el escalamiento; el código está atado a su IP'
