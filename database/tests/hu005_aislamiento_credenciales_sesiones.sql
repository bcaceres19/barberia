-- Pruebas SQL de HU-005: staff_credential, staff_session y las tres
-- funciones SECURITY DEFINER (authn_resolve_login_tenant, auth_get_credential,
-- authn_resolve_session_tenant). Cubren la parte de CA-005-01, CA-005-03,
-- CA-005-05 (a nivel de base de datos: ver DP-SEG-08 en dudas-pendientes.md
-- para la parte de "endpoint privado real" que queda pendiente) y CA-005-07
-- verificable sin el hasher Go real.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu005_credenciales_sesiones.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/hu005_aislamiento_credenciales_sesiones.sql
--
-- Mismo requisito de arnés que hu001_aislamiento_rls.sql: la conexión debe
-- poder ejecutar `SET ROLE barberia_app`.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== HU-005 · pruebas de credenciales y sesiones ==='

-- ---------------------------------------------------------------------------
-- DDL-AUT-01 · barberia_app no puede leer staff_credential directamente
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  BEGIN
    PERFORM * FROM staff_credential LIMIT 1;
    RAISE EXCEPTION 'DDL-AUT-01: barberia_app pudo leer staff_credential directamente.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL; -- Esperado: sin GRANT ni política de SELECT.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'DDL-AUT-01 OK · staff_credential no admite SELECT directo de barberia_app'

-- ---------------------------------------------------------------------------
-- CA-005-03 · auth_get_credential expone el hash de un único usuario, dentro
-- del tenant vigente, y CA-005-05 (parte de base de datos) · no cruza tenant
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_hash text;
  v_alg  text;
BEGIN
  SELECT password_hash, password_algorithm
  INTO v_hash, v_alg
  FROM auth_get_credential('aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaa1');

  IF v_hash IS NULL OR v_alg <> 'argon2id' THEN
    RAISE EXCEPTION 'CA-005-03: auth_get_credential no devolvió la credencial esperada de A.';
  END IF;

  -- Un staff_user_id real de LA OTRA barbería, con el contexto de A: debe
  -- devolver cero filas, nunca el hash de B (RN-TEN-01).
  PERFORM 1 FROM auth_get_credential('bbbbbbb1-bbbb-bbbb-bbbb-bbbbbbbbbbb1');
  IF FOUND THEN
    RAISE EXCEPTION 'CA-005-05: auth_get_credential devolvió una credencial de B con contexto de A.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-005-03/05 OK · auth_get_credential no cruza tenant'

-- ---------------------------------------------------------------------------
-- CA-005-02/07 · authn_resolve_login_tenant unifica correo inexistente e
-- inactivo (ambos NULL); solo el usuario activo resuelve su barbería
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_shop_active   uuid;
  v_shop_inactive uuid;
  v_shop_unknown  uuid;
BEGIN
  SELECT authn_resolve_login_tenant('duena.a@ejemplo.test') INTO v_shop_active;
  IF v_shop_active <> '11111111-1111-1111-1111-111111111111' THEN
    RAISE EXCEPTION 'CA-005-01: authn_resolve_login_tenant no resolvió el correo activo de A.';
  END IF;

  -- barbero.b@ejemplo.test existe pero está inactivo (dos_barberias.sql).
  SELECT authn_resolve_login_tenant('barbero.b@ejemplo.test') INTO v_shop_inactive;
  SELECT authn_resolve_login_tenant('no-existe@ejemplo.test') INTO v_shop_unknown;

  IF v_shop_inactive IS NOT NULL THEN
    RAISE EXCEPTION 'CA-005-07: authn_resolve_login_tenant resolvió tenant para un usuario inactivo.';
  END IF;
  IF v_shop_unknown IS NOT NULL THEN
    RAISE EXCEPTION 'CA-005-02: authn_resolve_login_tenant resolvió tenant para un correo inexistente.';
  END IF;
  IF v_shop_inactive IS DISTINCT FROM v_shop_unknown THEN
    RAISE EXCEPTION 'CA-005-02: usuario inactivo y correo inexistente producen resultados distintos.';
  END IF;

  -- Mayúsculas/espacios: la normalización de correo también ocurre en SQL
  -- (lower() dentro de la función), como defensa en profundidad además de
  -- la normalización que aplica el servicio Go.
  IF authn_resolve_login_tenant('DUENA.A@EJEMPLO.TEST') <> '11111111-1111-1111-1111-111111111111' THEN
    RAISE EXCEPTION 'authn_resolve_login_tenant no normaliza mayúsculas.';
  END IF;
END
$$;
\echo 'CA-005-01/02/07 OK · authn_resolve_login_tenant unifica inexistente e inactivo'

-- ---------------------------------------------------------------------------
-- CA-005-01/05 · staff_session: contexto de A no ve la sesión de B
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_count_a integer;
  v_count_b integer;
BEGIN
  SELECT count(*) INTO v_count_a FROM staff_session
    WHERE barbershop_id = '11111111-1111-1111-1111-111111111111';
  SELECT count(*) INTO v_count_b FROM staff_session
    WHERE barbershop_id = '22222222-2222-2222-2222-222222222222';

  IF v_count_a <> 1 THEN
    RAISE EXCEPTION 'CA-005-01: se esperaba 1 sesión de A y se vieron %.', v_count_a;
  END IF;
  IF v_count_b <> 0 THEN
    RAISE EXCEPTION 'CA-005-05: el contexto de A vio % sesiones de B.', v_count_b;
  END IF;

  -- Intentar insertar una sesión con el tenant ajeno falla (WITH CHECK).
  BEGIN
    INSERT INTO staff_session (barbershop_id, staff_user_id, token_hash, expires_at)
    VALUES ('22222222-2222-2222-2222-222222222222',
            'bbbbbbb1-bbbb-bbbb-bbbb-bbbbbbbbbbb1',
            encode(sha256('intruso'::bytea), 'hex'),
            now() + interval '30 days');
    RAISE EXCEPTION 'CA-005-05: se pudo insertar una sesión con el barbershop_id de B.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL; -- Esperado.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-005-01/05 OK · staff_session aísla por tenant'

-- ---------------------------------------------------------------------------
-- authn_resolve_session_tenant: token vigente resuelve, revocado/vencido no
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_shop uuid;
BEGIN
  SELECT authn_resolve_session_tenant(encode(sha256('token-ficticio-a'::bytea), 'hex')) INTO v_shop;
  IF v_shop <> '11111111-1111-1111-1111-111111111111' THEN
    RAISE EXCEPTION 'authn_resolve_session_tenant no resolvió el token vigente de A.';
  END IF;

  SELECT authn_resolve_session_tenant(encode(sha256('token-desconocido'::bytea), 'hex')) INTO v_shop;
  IF v_shop IS NOT NULL THEN
    RAISE EXCEPTION 'authn_resolve_session_tenant resolvió un token desconocido.';
  END IF;
END
$$;
\echo 'authn_resolve_session_tenant OK · token vigente resuelve, desconocido no'

-- ---------------------------------------------------------------------------
-- Formato de datos: password_hash 32-512 caracteres
-- ---------------------------------------------------------------------------
-- BEGIN/ROLLBACK aislado: borra la fila fixture solo dentro de esta
-- transacción para poder reinsertar con un hash inválido sin chocar con la
-- PK, sin afectar el resto del archivo (se revierte al final).
BEGIN;
DELETE FROM staff_credential WHERE staff_user_id = 'aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaa1';

DO $$
BEGIN
  BEGIN
    INSERT INTO staff_credential (staff_user_id, barbershop_id, password_hash)
    VALUES ('aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaa1',
            '11111111-1111-1111-1111-111111111111', 'corto');
    RAISE EXCEPTION 'staff_credential_password_hash_ck no rechazó un hash demasiado corto.';
  EXCEPTION
    WHEN check_violation THEN
      NULL; -- Esperado.
  END;
END
$$;

ROLLBACK;
\echo 'Formato OK · restricción de longitud de password_hash activa'

\echo '=== HU-005 · todas las comprobaciones pasaron ==='
