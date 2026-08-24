-- Pruebas SQL de HU-023: asignación de servicios a barberos. Cubren
-- CA-023-04, CA-023-06 y CA-023-07 a nivel de base de datos (esquema, FK
-- compuestas, RLS, grants); la orquestación completa de asignar/desasignar,
-- la carrera de DEC-068 y la traducción a apperr viven en el servicio y el
-- repositorio Go (internal/modules/catalog), probados aparte contra
-- PostgreSQL real.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu021_barberos.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu022_catalogo.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu023_asignaciones.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/hu023_asignaciones.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- (estandar-base-datos.md §9.8), igual que hu001_aislamiento_rls.sql /
-- hu021_barberos.sql / hu022_catalogo.sql.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== HU-023 · pruebas de asignación de servicios a barberos ==='

-- ---------------------------------------------------------------------------
-- Esquema mínimo: columnas exactas (CA-023-07: sin nombre, duración, precio
-- ni estado propios)
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_columns text;
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'barber_service'
  ) THEN
    RAISE EXCEPTION 'barber_service debe existir.';
  END IF;

  SELECT string_agg(column_name, ', ' ORDER BY ordinal_position) INTO v_columns
  FROM information_schema.columns
  WHERE table_schema = 'public' AND table_name = 'barber_service';

  IF v_columns IS DISTINCT FROM 'barbershop_id, barber_id, service_id, created_at' THEN
    RAISE EXCEPTION 'CA-023-07: barber_service debe tener EXACTAMENTE barbershop_id, barber_id, '
      'service_id, created_at (en ese orden), tiene: %', v_columns;
  END IF;
END
$$;
\echo 'esquema OK · barber_service tiene exactamente las columnas de HU-023, sin duplicar atributos'

-- ---------------------------------------------------------------------------
-- PK compuesta y FK compuestas por barbershop_id
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE table_schema = 'public' AND table_name = 'barber_service'
      AND constraint_type = 'PRIMARY KEY' AND constraint_name = 'barber_service_pk'
  ) THEN
    RAISE EXCEPTION 'barber_service_pk debe existir como PRIMARY KEY.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE table_schema = 'public' AND table_name = 'barber_service'
      AND constraint_type = 'FOREIGN KEY' AND constraint_name = 'barber_service_barber_fk'
  ) THEN
    RAISE EXCEPTION 'barber_service_barber_fk debe existir.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE table_schema = 'public' AND table_name = 'barber_service'
      AND constraint_type = 'FOREIGN KEY' AND constraint_name = 'barber_service_service_fk'
  ) THEN
    RAISE EXCEPTION 'barber_service_service_fk debe existir.';
  END IF;
END
$$;
\echo 'constraint OK · PK compuesta y FK compuestas de barber_service existen'

-- ---------------------------------------------------------------------------
-- CA-023-06 (nivel base de datos) · RLS forzada, sin BYPASSRLS
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_relrowsecurity boolean;
  v_relforcerowsecurity boolean;
BEGIN
  SELECT relrowsecurity, relforcerowsecurity INTO v_relrowsecurity, v_relforcerowsecurity
  FROM pg_class
  WHERE relname = 'barber_service' AND relnamespace = 'public'::regnamespace;

  IF NOT v_relrowsecurity OR NOT v_relforcerowsecurity THEN
    RAISE EXCEPTION 'barber_service debe tener RLS habilitada Y forzada.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM pg_roles WHERE rolname = 'barberia_app' AND rolbypassrls
  ) THEN
    RAISE EXCEPTION 'barberia_app no puede tener BYPASSRLS.';
  END IF;
END
$$;
\echo 'RLS OK · barber_service tiene RLS habilitada y forzada, sin BYPASSRLS'

-- ---------------------------------------------------------------------------
-- Grants exactos: SELECT/INSERT/DELETE, nunca UPDATE (CA-023-07: sin campo
-- editable)
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_privs text;
BEGIN
  SELECT string_agg(privilege_type, ', ' ORDER BY privilege_type) INTO v_privs
  FROM information_schema.table_privileges
  WHERE table_schema = 'public' AND table_name = 'barber_service' AND grantee = 'barberia_app';

  IF v_privs IS DISTINCT FROM 'DELETE, INSERT, SELECT' THEN
    RAISE EXCEPTION 'barberia_app debe tener EXACTAMENTE SELECT/INSERT/DELETE sobre barber_service, tiene: %', v_privs;
  END IF;
END
$$;
\echo 'grants OK · barberia_app tiene exactamente SELECT/INSERT/DELETE sobre barber_service, sin UPDATE'

-- ---------------------------------------------------------------------------
-- Ausencia de contexto de tenant: falla cerrado
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
BEGIN
  BEGIN
    PERFORM count(*) FROM barber_service;
    RAISE EXCEPTION 'sin contexto: se pudo leer barber_service sin app.barbershop_id fijado.';
  EXCEPTION
    WHEN OTHERS THEN
      IF SQLERRM NOT LIKE '%app.barbershop_id%' THEN
        RAISE EXCEPTION 'sin contexto: error inesperado distinto al de falta de configuración: %', SQLERRM;
      END IF;
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'sin contexto OK · leer barber_service sin app.barbershop_id fijado falla cerrado'

-- ---------------------------------------------------------------------------
-- CA-023-02/03 · Asignar dentro del propio tenant: repetible sin duplicar,
-- un mismo servicio a varios barberos
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_barber1 uuid;
  v_barber2 uuid;
  v_service uuid;
  v_count integer;
BEGIN
  INSERT INTO barber (barbershop_id, full_name)
  VALUES ('11111111-1111-1111-1111-111111111111', 'HU023 Barbero SQL Uno')
  RETURNING id INTO v_barber1;
  INSERT INTO barber (barbershop_id, full_name)
  VALUES ('11111111-1111-1111-1111-111111111111', 'HU023 Barbero SQL Dos')
  RETURNING id INTO v_barber2;
  INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
  VALUES ('11111111-1111-1111-1111-111111111111', 'HU023 Servicio SQL ' || gen_random_uuid(), 30, 20000)
  RETURNING id INTO v_service;

  INSERT INTO barber_service (barbershop_id, barber_id, service_id)
  VALUES ('11111111-1111-1111-1111-111111111111', v_barber1, v_service);

  -- CA-023-02: repetir exactamente la misma fila no crea una segunda
  -- (ON CONFLICT es responsabilidad del repositorio Go; aquí se confirma
  -- que el propio esquema -la PK- es lo que lo hace posible).
  INSERT INTO barber_service (barbershop_id, barber_id, service_id)
  VALUES ('11111111-1111-1111-1111-111111111111', v_barber1, v_service)
  ON CONFLICT (barbershop_id, barber_id, service_id) DO NOTHING;

  SELECT count(*) INTO v_count FROM barber_service
   WHERE barbershop_id = '11111111-1111-1111-1111-111111111111' AND service_id = v_service AND barber_id = v_barber1;
  IF v_count <> 1 THEN
    RAISE EXCEPTION 'CA-023-02: se esperaba exactamente 1 fila tras repetir la asignación, hay %.', v_count;
  END IF;

  -- CA-023-03: el mismo servicio a un segundo barbero es una fila
  -- independiente.
  INSERT INTO barber_service (barbershop_id, barber_id, service_id)
  VALUES ('11111111-1111-1111-1111-111111111111', v_barber2, v_service);

  SELECT count(*) INTO v_count FROM barber_service
   WHERE barbershop_id = '11111111-1111-1111-1111-111111111111' AND service_id = v_service;
  IF v_count <> 2 THEN
    RAISE EXCEPTION 'CA-023-03: se esperaban 2 asignaciones distintas para el mismo servicio, hay %.', v_count;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-023-02/03 OK · asignación repetida no duplica; un servicio compartido por varios barberos son filas independientes'

-- ---------------------------------------------------------------------------
-- CA-023-04 · La base rechaza asociar un barbero y un servicio de
-- barberías DISTINTAS, aun con el rol de aplicación real (FK compuesta)
-- ---------------------------------------------------------------------------
-- Prepara un barbero real de A y un servicio real de B en transacciones
-- confirmadas por separado (para no depender de visibilidad entre RLS y la
-- transacción de prueba).
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';
INSERT INTO barber (id, barbershop_id, full_name)
VALUES ('77777777-1111-1111-1111-777777777771', '11111111-1111-1111-1111-111111111111', 'HU023 Barbero Cruzado A');
RESET ROLE;
COMMIT;

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '22222222-2222-2222-2222-222222222222';
INSERT INTO service (id, barbershop_id, name, duration_minutes, price_amount)
VALUES ('77777777-2222-2222-2222-777777777772', '22222222-2222-2222-2222-222222222222',
        'HU023 Servicio Cruzado B ' || gen_random_uuid(), 30, 20000);
RESET ROLE;
COMMIT;

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  BEGIN
    -- barbershop_id = A (contexto vigente, pasa WITH CHECK de RLS), pero
    -- service_id pertenece a B: la FK compuesta barber_service_service_fk
    -- NO encuentra (A, service_id_de_B) en service(barbershop_id, id) y
    -- rechaza.
    INSERT INTO barber_service (barbershop_id, barber_id, service_id)
    VALUES ('11111111-1111-1111-1111-111111111111',
            '77777777-1111-1111-1111-777777777771',
            '77777777-2222-2222-2222-777777777772');
    RAISE EXCEPTION 'CA-023-04: se permitió asociar un barbero de A con un servicio de B.';
  EXCEPTION
    WHEN foreign_key_violation THEN NULL; -- Esperado.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-023-04 OK · la FK compuesta rechaza asociar un barbero y un servicio de barberías distintas'

-- ---------------------------------------------------------------------------
-- CA-023-06 (nivel base de datos) · El contexto de A no ve ni puede
-- insertar/borrar una fila de barber_service de B
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '22222222-2222-2222-2222-222222222222';

DO $$
DECLARE
  v_barber uuid;
  v_service uuid;
BEGIN
  INSERT INTO barber (barbershop_id, full_name)
  VALUES ('22222222-2222-2222-2222-222222222222', 'HU023 Barbero B Aislamiento')
  RETURNING id INTO v_barber;
  INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
  VALUES ('22222222-2222-2222-2222-222222222222', 'HU023 Servicio B Aislamiento ' || gen_random_uuid(), 30, 20000)
  RETURNING id INTO v_service;
  INSERT INTO barber_service (barbershop_id, barber_id, service_id)
  VALUES ('22222222-2222-2222-2222-222222222222', v_barber, v_service);

  PERFORM set_config('hu023.barber_b', v_barber::text, false);
  PERFORM set_config('hu023.service_b', v_service::text, false);
END
$$;

RESET ROLE;
COMMIT;

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_visible integer;
BEGIN
  SELECT count(*) INTO v_visible FROM barber_service
   WHERE barbershop_id = '22222222-2222-2222-2222-222222222222';
  IF v_visible <> 0 THEN
    RAISE EXCEPTION 'CA-023-06: el contexto de A ve % filas de barber_service de B.', v_visible;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-023-06 OK · el contexto de A no ve ninguna fila de barber_service de B'

-- ---------------------------------------------------------------------------
-- DELETE se permite para barberia_app SOBRE SU PROPIO tenant (a diferencia
-- de barber/service): retirar una asignación es el propósito de esta tabla
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_barber uuid;
  v_service uuid;
  v_deleted integer;
BEGIN
  INSERT INTO barber (barbershop_id, full_name)
  VALUES ('11111111-1111-1111-1111-111111111111', 'HU023 Barbero Delete')
  RETURNING id INTO v_barber;
  INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
  VALUES ('11111111-1111-1111-1111-111111111111', 'HU023 Servicio Delete ' || gen_random_uuid(), 30, 20000)
  RETURNING id INTO v_service;
  INSERT INTO barber_service (barbershop_id, barber_id, service_id)
  VALUES ('11111111-1111-1111-1111-111111111111', v_barber, v_service);

  DELETE FROM barber_service WHERE barbershop_id = '11111111-1111-1111-1111-111111111111'
    AND barber_id = v_barber AND service_id = v_service;
  GET DIAGNOSTICS v_deleted = ROW_COUNT;
  IF v_deleted <> 1 THEN
    RAISE EXCEPTION 'se esperaba poder borrar la propia fila de barber_service, ROW_COUNT=%.', v_deleted;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'DELETE OK · barberia_app puede retirar su propia asignación (a diferencia de barber/service)'

-- ---------------------------------------------------------------------------
-- Sin UPDATE: ningún campo de barber_service es editable (CA-023-07)
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.table_privileges
    WHERE table_schema = 'public' AND table_name = 'barber_service'
      AND grantee = 'barberia_app' AND privilege_type = 'UPDATE'
  ) THEN
    RAISE EXCEPTION 'CA-023-07: barberia_app no debe tener GRANT UPDATE sobre barber_service.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'barber_service' AND cmd = 'UPDATE'
  ) THEN
    RAISE EXCEPTION 'CA-023-07: no debe existir ninguna política UPDATE sobre barber_service.';
  END IF;
END
$$;
\echo 'CA-023-07 OK · sin GRANT ni política UPDATE sobre barber_service (nada editable)'

-- ---------------------------------------------------------------------------
-- El rol de aplicación real no tiene BYPASSRLS (repetido explícitamente
-- para HU-023, mismo criterio que hu001/hu021/hu022)
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_app' AND rolbypassrls) THEN
    RAISE EXCEPTION 'barberia_app no puede tener BYPASSRLS.';
  END IF;
END
$$;
\echo 'rol OK · barberia_app sin BYPASSRLS'

-- ---------------------------------------------------------------------------
-- Limpieza de las dos filas que se confirmaron con COMMIT más arriba (el
-- barbero de A y el servicio de B usados para CA-023-04): con el rol de
-- sesión sin RLS forzada (superusuario/`barberia_migrator`), para dejar la
-- base exactamente como estaba al empezar este archivo. barberia_app nunca
-- podría hacer esta limpieza (sin GRANT DELETE sobre barber/service,
-- RN-SER-03/CA-021-07); no hace falta limpiar barber_service porque su
-- única fila potencial (la que el propio bloque CA-023-04 intentó insertar)
-- nunca llegó a confirmarse (ROLLBACK).
-- ---------------------------------------------------------------------------
DELETE FROM barber WHERE id = '77777777-1111-1111-1111-777777777771';
DELETE FROM service WHERE id = '77777777-2222-2222-2222-777777777772';

\echo '=== HU-023 · todas las comprobaciones pasaron ==='
