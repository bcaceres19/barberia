-- Pruebas SQL de DEC-139 / DEC-081: barbershop.login_otp_channel y los destinos
-- que devuelve auth_phone_challenge_request. Cubren el valor por defecto, la
-- restricción CHECK, qué destinos salen con cada canal (correo, WhatsApp, ambos),
-- que nunca se sustituye un canal por otro, que una cuenta sin destino utilizable
-- o inactiva responde igual que una inexistente y que el canal de una barbería no
-- afecta a la otra. El envío en sí vive en el servicio Go (internal/modules/auth).
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu005_credenciales_sesiones.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu007_reto_telefonico.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/login_otp_channel.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- (estandar-base-datos.md §9.8). Todo se revierte con ROLLBACK.
--
-- Cuentas (testdata/hu007_reto_telefonico.sql): duena.a y dueno.b tienen
-- teléfono verificado; barbero.a no lo tiene; barbero.b está inactivo.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== DEC-139 · pruebas de login_otp_channel ==='

-- ---------------------------------------------------------------------------
-- Valor por defecto y restricción
-- ---------------------------------------------------------------------------
BEGIN;

DO $$
DECLARE
  v_id text;
  v_channel text;
BEGIN
  INSERT INTO barbershop (id, name, timezone)
  VALUES ('99999999-9999-4999-8999-999999999999', 'Barbería nueva', 'America/Bogota');
  SELECT login_otp_channel INTO v_channel FROM barbershop WHERE id = '99999999-9999-4999-8999-999999999999';
  IF v_channel IS DISTINCT FROM 'email' THEN
    RAISE EXCEPTION 'una barbería nueva debe nacer con correo (DEC-081), leyó %.', v_channel;
  END IF;

  FOREACH v_id IN ARRAY ARRAY['email', 'whatsapp', 'both'] LOOP
    UPDATE barbershop SET login_otp_channel = v_id WHERE id = '99999999-9999-4999-8999-999999999999';
  END LOOP;

  FOREACH v_id IN ARRAY ARRAY['sms', 'Email', 'WHATSAPP', '', 'email,whatsapp', 'push'] LOOP
    BEGIN
      UPDATE barbershop SET login_otp_channel = v_id WHERE id = '99999999-9999-4999-8999-999999999999';
      RAISE EXCEPTION 'el CHECK debió rechazar el canal "%".', v_id;
    EXCEPTION WHEN check_violation THEN
      NULL;
    END;
  END LOOP;

  BEGIN
    UPDATE barbershop SET login_otp_channel = NULL WHERE id = '99999999-9999-4999-8999-999999999999';
    RAISE EXCEPTION 'NOT NULL debió rechazar el canal nulo.';
  EXCEPTION WHEN not_null_violation THEN
    NULL;
  END;
END
$$;

ROLLBACK;
\echo 'defaults OK · correo para las nuevas, CHECK y NOT NULL'

-- ---------------------------------------------------------------------------
-- Destinos por canal (la IP escalada es distinta en cada caso: la solicitud
-- acepta una por IP dentro del enfriamiento)
-- ---------------------------------------------------------------------------
BEGIN;

UPDATE barbershop SET login_otp_channel = 'email'    WHERE id = '11111111-1111-1111-1111-111111111111';
UPDATE barbershop SET login_otp_channel = 'whatsapp' WHERE id = '22222222-2222-2222-2222-222222222222';

SET ROLE barberia_app;

DO $$
DECLARE
  v_code_hash text := encode(sha256('login-otp-channel-code'::bytea), 'hex');
  v_ip        text;
  v_ok        boolean;
  v_phone     text;
  v_email     text;
  v_n         integer := 0;
BEGIN
  -- Canal correo (barbería A).
  v_n := v_n + 1; v_ip := encode(sha256(('ip-' || v_n)::bytea), 'hex');
  FOR i IN 1..6 LOOP PERFORM login_throttle_register_attempt(v_ip, 900, 86400, 5, 172800); END LOOP;
  SELECT accepted, phone, email INTO v_ok, v_phone, v_email
    FROM auth_phone_challenge_request('duena.a@ejemplo.test', v_ip, v_code_hash, 300, 900, 3, 60);
  IF NOT v_ok OR v_phone IS NOT NULL OR v_email IS DISTINCT FROM 'duena.a@ejemplo.test' THEN
    RAISE EXCEPTION 'canal email: debía devolver solo el correo de la cuenta, obtuvo (%, %, %).', v_ok, v_phone, v_email;
  END IF;

  -- Canal correo: una cuenta SIN teléfono verificado igual recibe el código.
  v_n := v_n + 1; v_ip := encode(sha256(('ip-' || v_n)::bytea), 'hex');
  FOR i IN 1..6 LOOP PERFORM login_throttle_register_attempt(v_ip, 900, 86400, 5, 172800); END LOOP;
  SELECT accepted, phone, email INTO v_ok, v_phone, v_email
    FROM auth_phone_challenge_request('barbero.a@ejemplo.test', v_ip, v_code_hash, 300, 900, 3, 60);
  IF NOT v_ok OR v_phone IS NOT NULL OR v_email IS DISTINCT FROM 'barbero.a@ejemplo.test' THEN
    RAISE EXCEPTION 'canal email sin teléfono: debía devolver solo el correo, obtuvo (%, %, %).', v_ok, v_phone, v_email;
  END IF;

  -- Canal WhatsApp (barbería B): solo teléfono verificado, sin correo.
  v_n := v_n + 1; v_ip := encode(sha256(('ip-' || v_n)::bytea), 'hex');
  FOR i IN 1..6 LOOP PERFORM login_throttle_register_attempt(v_ip, 900, 86400, 5, 172800); END LOOP;
  SELECT accepted, phone, email INTO v_ok, v_phone, v_email
    FROM auth_phone_challenge_request('dueno.b@ejemplo.test', v_ip, v_code_hash, 300, 900, 3, 60);
  IF NOT v_ok OR v_phone IS NULL OR v_email IS NOT NULL THEN
    RAISE EXCEPTION 'canal whatsapp: debía devolver solo el teléfono, obtuvo (%, %, %).', v_ok, v_phone, v_email;
  END IF;

  -- Canal WhatsApp: cuenta inactiva (barbero.b) no recibe nada.
  v_n := v_n + 1; v_ip := encode(sha256(('ip-' || v_n)::bytea), 'hex');
  FOR i IN 1..6 LOOP PERFORM login_throttle_register_attempt(v_ip, 900, 86400, 5, 172800); END LOOP;
  SELECT accepted, phone, email INTO v_ok, v_phone, v_email
    FROM auth_phone_challenge_request('barbero.b@ejemplo.test', v_ip, v_code_hash, 300, 900, 3, 60);
  IF v_ok OR v_phone IS NOT NULL OR v_email IS NOT NULL THEN
    RAISE EXCEPTION 'cuenta inactiva: no debía recibir destino, obtuvo (%, %, %).', v_ok, v_phone, v_email;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'canales email y whatsapp OK · destinos del canal, cuenta inactiva sin destino, aislamiento entre barberías'

-- ---------------------------------------------------------------------------
-- Ambos canales y WhatsApp sin teléfono verificado
-- ---------------------------------------------------------------------------
BEGIN;

UPDATE barbershop SET login_otp_channel = 'both'     WHERE id = '11111111-1111-1111-1111-111111111111';
UPDATE barbershop SET login_otp_channel = 'whatsapp' WHERE id = '22222222-2222-2222-2222-222222222222';

SET ROLE barberia_app;

DO $$
DECLARE
  v_code_hash text := encode(sha256('login-otp-channel-code'::bytea), 'hex');
  v_ip        text;
  v_ok        boolean;
  v_phone     text;
  v_email     text;
BEGIN
  -- Ambos: teléfono verificado y correo juntos, UNA sola fila de reto (un código).
  v_ip := encode(sha256('ip-both-1'::bytea), 'hex');
  FOR i IN 1..6 LOOP PERFORM login_throttle_register_attempt(v_ip, 900, 86400, 5, 172800); END LOOP;
  SELECT accepted, phone, email INTO v_ok, v_phone, v_email
    FROM auth_phone_challenge_request('duena.a@ejemplo.test', v_ip, v_code_hash, 300, 900, 3, 60);
  IF NOT v_ok OR v_phone IS NULL OR v_email IS DISTINCT FROM 'duena.a@ejemplo.test' THEN
    RAISE EXCEPTION 'canal both: debía devolver teléfono y correo, obtuvo (%, %, %).', v_ok, v_phone, v_email;
  END IF;

  -- Ambos con teléfono sin verificar: sale solo el correo; no se inventa un teléfono.
  v_ip := encode(sha256('ip-both-2'::bytea), 'hex');
  FOR i IN 1..6 LOOP PERFORM login_throttle_register_attempt(v_ip, 900, 86400, 5, 172800); END LOOP;
  SELECT accepted, phone, email INTO v_ok, v_phone, v_email
    FROM auth_phone_challenge_request('barbero.a@ejemplo.test', v_ip, v_code_hash, 300, 900, 3, 60);
  IF NOT v_ok OR v_phone IS NOT NULL OR v_email IS DISTINCT FROM 'barbero.a@ejemplo.test' THEN
    RAISE EXCEPTION 'canal both sin teléfono: debía devolver solo el correo, obtuvo (%, %, %).', v_ok, v_phone, v_email;
  END IF;
END
$$;

RESET ROLE;

-- WhatsApp sin teléfono verificado: no hay destino utilizable y NO se cae a correo.
UPDATE barbershop SET login_otp_channel = 'whatsapp' WHERE id = '11111111-1111-1111-1111-111111111111';
SET ROLE barberia_app;
DO $$
DECLARE
  v_code_hash text := encode(sha256('login-otp-channel-code'::bytea), 'hex');
  v_ip        text := encode(sha256('ip-wa-1'::bytea), 'hex');
  v_ok        boolean;
  v_phone     text;
  v_email     text;
BEGIN
  FOR i IN 1..6 LOOP PERFORM login_throttle_register_attempt(v_ip, 900, 86400, 5, 172800); END LOOP;
  SELECT accepted, phone, email INTO v_ok, v_phone, v_email
    FROM auth_phone_challenge_request('barbero.a@ejemplo.test', v_ip, v_code_hash, 300, 900, 3, 60);
  IF v_ok OR v_phone IS NOT NULL OR v_email IS NOT NULL THEN
    RAISE EXCEPTION 'whatsapp sin teléfono verificado: no debía haber destino ni respaldo por correo, obtuvo (%, %, %).', v_ok, v_phone, v_email;
  END IF;
END
$$;
RESET ROLE;

ROLLBACK;
\echo 'canal both y whatsapp sin teléfono OK · un solo reto, sin respaldo a un canal no configurado'

-- ---------------------------------------------------------------------------
-- Una solicitud aceptada crea una sola fila de reto aunque haya dos destinos
-- ---------------------------------------------------------------------------
BEGIN;

UPDATE barbershop SET login_otp_channel = 'both' WHERE id = '11111111-1111-1111-1111-111111111111';
SET ROLE barberia_app;
DO $$
DECLARE
  v_code_hash text := encode(sha256('login-otp-single-row'::bytea), 'hex');
  v_ip        text := encode(sha256('ip-single-row'::bytea), 'hex');
  v_ok        boolean;
BEGIN
  FOR i IN 1..6 LOOP PERFORM login_throttle_register_attempt(v_ip, 900, 86400, 5, 172800); END LOOP;
  SELECT accepted INTO v_ok
    FROM auth_phone_challenge_request('duena.a@ejemplo.test', v_ip, v_code_hash, 300, 900, 3, 60);
  IF NOT v_ok THEN
    RAISE EXCEPTION 'el reto con ambos canales debía aceptarse.';
  END IF;
END
$$;
RESET ROLE;

DO $$
DECLARE
  v_rows integer;
BEGIN
  SELECT count(*) INTO v_rows FROM auth_phone_challenge
   WHERE code_hash = encode(sha256('login-otp-single-row'::bytea), 'hex');
  IF v_rows <> 1 THEN
    RAISE EXCEPTION 'un reto con dos destinos debe ser UNA operación lógica (1 fila), hay %.', v_rows;
  END IF;
END
$$;

ROLLBACK;
\echo 'operación lógica única OK · una fila por solicitud aunque haya dos destinos'

\echo '=== DEC-139 · todas las pruebas pasaron ==='
