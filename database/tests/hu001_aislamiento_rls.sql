-- Pruebas SQL de aislamiento de HU-001. Cubren CA-001-01 a CA-001-06.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/hu001_aislamiento_rls.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`,
-- porque las políticas se prueban con el rol real (estandar-base-datos.md §9.8).
-- En local y en CI, con PostgreSQL efímero, se conecta como superusuario. Si se
-- conecta como `barberia_migrator`, hace falta `GRANT barberia_app TO
-- barberia_migrator` una sola vez; con NOINHERIT solo surte efecto vía SET ROLE.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== HU-001 · pruebas de aislamiento ==='

-- ---------------------------------------------------------------------------
-- CA-001-04 · El rol de aplicación no tiene BYPASSRLS ni es propietario
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_owned integer;
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_app') THEN
    RAISE EXCEPTION 'CA-001-04: el rol barberia_app no existe.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM pg_roles
    WHERE rolname = 'barberia_app' AND (rolsuper OR rolbypassrls)
  ) THEN
    RAISE EXCEPTION 'CA-001-04: barberia_app tiene superusuario o BYPASSRLS.';
  END IF;

  SELECT count(*) INTO v_owned
  FROM pg_class c
  JOIN pg_namespace n ON n.oid = c.relnamespace
  WHERE n.nspname = 'public'
    AND c.relkind = 'r'
    AND pg_get_userbyid(c.relowner) = 'barberia_app';

  IF v_owned > 0 THEN
    RAISE EXCEPTION 'CA-001-04: barberia_app es propietario de % tablas.', v_owned;
  END IF;
END
$$;
\echo 'CA-001-04 OK · rol de aplicación sin BYPASSRLS ni propiedad'

-- ---------------------------------------------------------------------------
-- CA-001-06 · Ninguna columna temporal sin zona horaria
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_offenders text;
BEGIN
  SELECT string_agg(table_name || '.' || column_name, ', ')
  INTO v_offenders
  FROM information_schema.columns
  WHERE table_schema = 'public'
    AND data_type IN ('timestamp without time zone', 'time with time zone');

  IF v_offenders IS NOT NULL THEN
    RAISE EXCEPTION 'CA-001-06: columnas temporales inválidas: %', v_offenders;
  END IF;
END
$$;
\echo 'CA-001-06 OK · todo instante almacena zona horaria'

-- ---------------------------------------------------------------------------
-- RLS habilitada Y forzada en toda tabla tenant-aware
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_offenders text;
BEGIN
  SELECT string_agg(c.relname, ', ')
  INTO v_offenders
  FROM pg_class c
  JOIN pg_namespace n ON n.oid = c.relnamespace
  WHERE n.nspname = 'public'
    AND c.relkind = 'r'
    -- `atlas_schema_revisions` pertenece a Atlas y no lleva barbershop_id.
    -- `login_throttle` es la única excepción de diseño: cuenta intentos por IP
    -- ANTES de saber quién solicita, y asociarla a una barbería permitiría
    -- repartir los intentos entre barberías para diluir el límite.
    -- Ver modelo-fisico-referencia.sql, sección A.4.
    AND c.relname NOT IN ('atlas_schema_revisions', 'login_throttle')
    AND NOT (c.relrowsecurity AND c.relforcerowsecurity);

  IF v_offenders IS NOT NULL THEN
    RAISE EXCEPTION 'Tablas sin RLS habilitada y forzada: %', v_offenders;
  END IF;
END
$$;
\echo 'RLS OK · habilitada y forzada en todas las tablas'

-- ---------------------------------------------------------------------------
-- CA-001-03 · Sin contexto, la consulta falla de forma explícita
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
DECLARE
  v_count integer;
BEGIN
  BEGIN
    SELECT count(*) INTO v_count FROM staff_user;
    RAISE EXCEPTION
      'CA-001-03: una consulta sin app.barbershop_id devolvió % filas en lugar de fallar.',
      v_count;
  EXCEPTION
    WHEN undefined_object THEN
      NULL;  -- Comportamiento esperado: el ajuste no existe y la consulta aborta.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-001-03 OK · sin contexto la operación falla, no devuelve todo'

-- ---------------------------------------------------------------------------
-- CA-001-01 · Con contexto de A solo se ven filas de A
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_visible_a integer;
  v_visible_b integer;
  v_shops     integer;
BEGIN
  SELECT count(*) INTO v_visible_a
  FROM staff_user
  WHERE barbershop_id = '11111111-1111-1111-1111-111111111111';

  SELECT count(*) INTO v_visible_b
  FROM staff_user
  WHERE barbershop_id = '22222222-2222-2222-2222-222222222222';

  SELECT count(*) INTO v_shops FROM barbershop;

  IF v_visible_a <> 3 THEN
    RAISE EXCEPTION 'CA-001-01: se esperaban 3 usuarios de A y se vieron %.', v_visible_a;
  END IF;

  IF v_visible_b <> 0 THEN
    RAISE EXCEPTION 'CA-001-01: se vieron % usuarios de B con contexto de A.', v_visible_b;
  END IF;

  IF v_shops <> 1 THEN
    RAISE EXCEPTION 'CA-001-01: se vieron % barberías; debía verse solo la propia.', v_shops;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-001-01 OK · el contexto de A no revela nada de B'

-- ---------------------------------------------------------------------------
-- CA-001-02 · Escritura cruzada rechazada por WITH CHECK
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  -- INSERT con el tenant ajeno.
  BEGIN
    INSERT INTO staff_user (barbershop_id, email, full_name)
    VALUES ('22222222-2222-2222-2222-222222222222',
            'intruso@ejemplo.test', 'Intruso');
    RAISE EXCEPTION 'CA-001-02: se pudo insertar una fila con el barbershop_id de B.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL;  -- Esperado: la política WITH CHECK rechaza la fila.
  END;

  -- UPDATE que intenta mover una fila propia hacia el tenant ajeno.
  BEGIN
    UPDATE staff_user
    SET barbershop_id = '22222222-2222-2222-2222-222222222222'
    WHERE id = 'aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2';
    RAISE EXCEPTION 'CA-001-02: se pudo mover una fila de A hacia B.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL;  -- Esperado.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-001-02 OK · WITH CHECK impide escribir hacia otra barbería'

-- ---------------------------------------------------------------------------
-- Un recurso de otra barbería es indistinguible de uno inexistente
-- (base de RN-TEN-01 y de la respuesta 404 de CA-003-03)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_known_of_b integer;
  v_inexistent integer;
BEGIN
  SELECT count(*) INTO v_known_of_b
  FROM staff_user WHERE id = 'bbbbbbb1-bbbb-bbbb-bbbb-bbbbbbbbbbb1';

  SELECT count(*) INTO v_inexistent
  FROM staff_user WHERE id = '99999999-9999-9999-9999-999999999999';

  IF v_known_of_b <> v_inexistent THEN
    RAISE EXCEPTION
      'RN-TEN-01: conocer el identificador de B produce un resultado distinto (% frente a %).',
      v_known_of_b, v_inexistent;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'RN-TEN-01 OK · adivinar un identificador ajeno no revela su existencia'

\echo '=== HU-001 · todas las comprobaciones pasaron ==='
