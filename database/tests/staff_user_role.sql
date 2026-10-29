-- Pruebas SQL del issue #365 (DEC-134): roles del usuario del área privada.
-- Cubren a nivel de base de datos la lista cerrada de roles, la unicidad por
-- (usuario, rol) con varios roles por usuario, la FK compuesta que impide un rol
-- para un usuario de otra barbería, el aislamiento RLS entre dos barberías y que
-- la aplicación solo puede leer. La derivación del panel vive en el servicio Go
-- (internal/modules/shops), no aquí.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/staff_user_role_365.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/staff_user_role.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- (estandar-base-datos.md §9.8), igual que panel_perfil_barbero_individual.sql.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== Issue #365 · pruebas de los roles del usuario ==='

-- ---------------------------------------------------------------------------
-- Lista cerrada de roles y unicidad por (usuario, rol)
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  BEGIN
    INSERT INTO staff_user_role (barbershop_id, staff_user_id, role)
    VALUES ('12121212-1212-1212-1212-121212121212', '365a0001-0000-4000-8000-000000000005', 'administrador');
    RAISE EXCEPTION 'un rol fuera de la lista cerrada debía rechazarse.';
  EXCEPTION WHEN check_violation THEN
    NULL;
  END;

  BEGIN
    INSERT INTO staff_user_role (barbershop_id, staff_user_id, role)
    VALUES ('12121212-1212-1212-1212-121212121212', '365a0001-0000-4000-8000-000000000004', 'empresa');
    RAISE EXCEPTION 'el mismo rol dos veces para un usuario debía rechazarse.';
  EXCEPTION WHEN unique_violation THEN
    NULL;
  END;
END
$$;
\echo 'CHECK y unicidad OK'

-- Un usuario puede ser empleado e independiente a la vez.
DO $$
DECLARE
  v_roles text[];
BEGIN
  SELECT array_agg(role ORDER BY role) INTO v_roles
    FROM staff_user_role
   WHERE staff_user_id = '365a0001-0000-4000-8000-000000000002';
  IF v_roles IS DISTINCT FROM ARRAY['empleado', 'independiente'] THEN
    RAISE EXCEPTION 'el usuario mixto debe tener empleado e independiente, tiene %.', v_roles;
  END IF;
END
$$;
\echo 'varios roles por usuario OK'

-- ---------------------------------------------------------------------------
-- FK compuesta: no se asigna un rol a un usuario de otra barbería
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  BEGIN
    INSERT INTO staff_user_role (barbershop_id, staff_user_id, role)
    VALUES ('12121212-1212-1212-1212-121212121212', '365a0001-0000-4000-8000-000000000006', 'empleado');
    RAISE EXCEPTION 'un rol para un usuario de otra barbería debía rechazarse.';
  EXCEPTION WHEN foreign_key_violation THEN
    NULL;
  END;
END
$$;
\echo 'FK compuesta OK'

-- ---------------------------------------------------------------------------
-- RLS: cada barbería ve solo los roles de sus usuarios
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '12121212-1212-1212-1212-121212121212';

DO $$
DECLARE
  v_total integer;
  v_other integer;
BEGIN
  SELECT count(*) INTO v_total FROM staff_user_role;
  IF v_total <> 5 THEN
    RAISE EXCEPTION 'la barbería A debe ver exactamente los 5 roles de sus usuarios (vio %).', v_total;
  END IF;
  SELECT count(*) INTO v_other FROM staff_user_role
   WHERE staff_user_id = '365a0001-0000-4000-8000-000000000006';
  IF v_other <> 0 THEN
    RAISE EXCEPTION 'la barbería A no debe ver los roles de un usuario de B (vio %).', v_other;
  END IF;
END
$$;
ROLLBACK;

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '34343434-3434-3434-3434-343434343434';

DO $$
DECLARE
  v_a integer;
  v_own integer;
BEGIN
  SELECT count(*) INTO v_a FROM staff_user_role
   WHERE staff_user_id IN ('365a0001-0000-4000-8000-000000000001', '365a0001-0000-4000-8000-000000000002',
                           '365a0001-0000-4000-8000-000000000003', '365a0001-0000-4000-8000-000000000004');
  IF v_a <> 0 THEN
    RAISE EXCEPTION 'la barbería B no debe ver los roles de usuarios de A (vio %).', v_a;
  END IF;
  SELECT count(*) INTO v_own FROM staff_user_role
   WHERE staff_user_id = '365a0001-0000-4000-8000-000000000006';
  IF v_own <> 1 THEN
    RAISE EXCEPTION 'la barbería B debe ver el rol de su usuario (vio %).', v_own;
  END IF;
END
$$;
ROLLBACK;
\echo 'RLS entre dos barberías OK'

-- ---------------------------------------------------------------------------
-- La aplicación solo lee: sin INSERT, UPDATE ni DELETE
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '12121212-1212-1212-1212-121212121212';

DO $$
BEGIN
  BEGIN
    INSERT INTO staff_user_role (barbershop_id, staff_user_id, role)
    VALUES ('12121212-1212-1212-1212-121212121212', '365a0001-0000-4000-8000-000000000005', 'empresa');
    RAISE EXCEPTION 'la aplicación no debe poder insertar roles.';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;

  BEGIN
    UPDATE staff_user_role SET role = 'empresa'
     WHERE staff_user_id = '365a0001-0000-4000-8000-000000000001';
    RAISE EXCEPTION 'la aplicación no debe poder cambiar roles.';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;

  BEGIN
    DELETE FROM staff_user_role WHERE staff_user_id = '365a0001-0000-4000-8000-000000000001';
    RAISE EXCEPTION 'la aplicación no debe poder borrar roles.';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;
END
$$;
ROLLBACK;
\echo 'solo lectura para la aplicación OK'

\echo '=== Issue #365 · todas las pruebas pasaron ==='
