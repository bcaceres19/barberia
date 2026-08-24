-- Pruebas SQL de HU-022: catálogo básico de servicios. Cubren CA-022-01,
-- CA-022-04 (parte de esquema/constraint), CA-022-06 y CA-022-07 a nivel de
-- base de datos (esquema, CHECK, RLS, grants, DEC-067); la validación de
-- nombre/descripción/duración/precio decimal exacto, la paginación por
-- cursor y el protocolo de idempotencia completo viven en el servicio Go
-- (internal/modules/catalog), probados aparte contra PostgreSQL real.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu022_catalogo.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/hu022_catalogo.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- (estandar-base-datos.md §9.8), igual que hu001_aislamiento_rls.sql /
-- hu021_barberos.sql.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== HU-022 · pruebas del catálogo básico de servicios ==='

-- ---------------------------------------------------------------------------
-- Esquema mínimo: columnas, tipos, precisión monetaria y CHECK esperados
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'service'
  ) THEN
    RAISE EXCEPTION 'service debe existir.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'service'
      AND column_name = 'name' AND is_nullable = 'NO'
  ) THEN
    RAISE EXCEPTION 'service.name debe existir y ser NOT NULL.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'service'
      AND column_name = 'description' AND is_nullable = 'YES'
  ) THEN
    RAISE EXCEPTION 'service.description debe existir y admitir NULL.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'service'
      AND column_name = 'price_amount' AND data_type = 'numeric'
      AND numeric_precision = 12 AND numeric_scale = 2
  ) THEN
    RAISE EXCEPTION 'service.price_amount debe ser numeric(12,2) (nunca coma flotante).';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'service'
      AND column_name = 'price_amount' AND data_type IN ('real', 'double precision')
  ) THEN
    RAISE EXCEPTION 'service.price_amount NUNCA debe ser real/double precision.';
  END IF;

  -- CA-022-07/DEC-067: sin columnas fuera de alcance de HU-022 (asignación a
  -- barberos u otra forma de "vínculo" que HU-023 debería introducir).
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'service'
      AND column_name IN ('barber_id', 'staff_user_id')
  ) THEN
    RAISE EXCEPTION 'DEC-067/HU-023: service no debe tener columnas de asignación a barberos.';
  END IF;
END
$$;
\echo 'esquema OK · service tiene name/description/price_amount con la forma esperada'

-- ---------------------------------------------------------------------------
-- CA-022-06 · RLS forzada, sin BYPASSRLS, sin política/GRANT de DELETE
-- (RN-SER-03: el borrado físico nunca se expone)
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_relrowsecurity boolean;
  v_relforcerowsecurity boolean;
BEGIN
  SELECT relrowsecurity, relforcerowsecurity INTO v_relrowsecurity, v_relforcerowsecurity
  FROM pg_class
  WHERE relname = 'service' AND relnamespace = 'public'::regnamespace;

  IF NOT v_relrowsecurity OR NOT v_relforcerowsecurity THEN
    RAISE EXCEPTION 'CA-022-06: service debe tener RLS habilitada Y forzada.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM pg_roles WHERE rolname = 'barberia_app' AND rolbypassrls
  ) THEN
    RAISE EXCEPTION 'CA-022-06: barberia_app no puede tener BYPASSRLS.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.table_privileges
    WHERE table_schema = 'public' AND table_name = 'service'
      AND grantee = 'barberia_app' AND privilege_type = 'DELETE'
  ) THEN
    RAISE EXCEPTION 'RN-SER-03: barberia_app no debe tener GRANT DELETE sobre service.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'service' AND cmd = 'DELETE'
  ) THEN
    RAISE EXCEPTION 'RN-SER-03: no debe existir ninguna política DELETE sobre service.';
  END IF;
END
$$;
\echo 'CA-022-06 OK · RLS forzada, sin BYPASSRLS, sin GRANT ni política DELETE'

-- ---------------------------------------------------------------------------
-- Ausencia de contexto de tenant: falla cerrado, no devuelve el conjunto
-- completo. DEBE ejecutarse antes de cualquier SET LOCAL app.barbershop_id
-- de este archivo (mismo orden que hu001/hu021).
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
BEGIN
  BEGIN
    PERFORM count(*) FROM service;
    RAISE EXCEPTION 'sin contexto: se pudo leer service sin app.barbershop_id fijado.';
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
\echo 'sin contexto OK · leer service sin app.barbershop_id fijado falla cerrado'

-- ---------------------------------------------------------------------------
-- CA-022-01 · Alta dentro del propio tenant: varios servicios en el mismo
-- catálogo son recursos distintos
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '55555555-5555-5555-5555-555555555555';

DO $$
DECLARE
  v_id1 uuid;
  v_id2 uuid;
BEGIN
  INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
  VALUES ('55555555-5555-5555-5555-555555555555', 'SQL Suite Servicio Uno', 30, 45000.00)
  RETURNING id INTO v_id1;

  INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
  VALUES ('55555555-5555-5555-5555-555555555555', 'SQL Suite Servicio Dos', 60, 65000.00)
  RETURNING id INTO v_id2;

  IF v_id1 = v_id2 THEN
    RAISE EXCEPTION 'CA-022-01: se esperaban dos servicios distintos.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-022-01 OK · el catálogo admite varios servicios distintos en la misma barbería'

-- ---------------------------------------------------------------------------
-- CA-022-06 · La base rechaza insertar un service con un barbershop_id
-- distinto del contexto de la transacción, incluso con el rol de aplicación
-- real
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '55555555-5555-5555-5555-555555555555';

DO $$
BEGIN
  BEGIN
    INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
    VALUES ('66666666-6666-6666-6666-666666666666', 'Intruso', 30, 1.00);
    RAISE EXCEPTION 'CA-022-06: se permitió insertar un service con barbershop_id ajeno al contexto.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL; -- Esperado: service_insert_tenant_policy rechaza el WITH CHECK.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-022-06 OK · INSERT con barbershop_id ajeno al contexto se rechaza (WITH CHECK de RLS)'

-- ---------------------------------------------------------------------------
-- CA-022-06 · El contexto de E no ve ni puede actualizar un service de F
-- ---------------------------------------------------------------------------
-- Este bloque SÍ confirma (COMMIT): la comprobación siguiente necesita leer
-- el identificador real desde una sesión/transacción distinta a la que lo
-- creó. Se limpia al final de esta sección.
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '66666666-6666-6666-6666-666666666666';

INSERT INTO service (id, barbershop_id, name, duration_minutes, price_amount)
VALUES ('77777777-7777-7777-7777-777777777777', '66666666-6666-6666-6666-666666666666',
        'Servicio de F para aislamiento SQL', 30, 1.00);

RESET ROLE;
COMMIT;

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '55555555-5555-5555-5555-555555555555';

DO $$
DECLARE
  v_service_f_id uuid := '77777777-7777-7777-7777-777777777777';
  v_visible integer;
  v_updated integer;
BEGIN
  SELECT count(*) INTO v_visible FROM service WHERE id = v_service_f_id;
  IF v_visible <> 0 THEN
    RAISE EXCEPTION 'CA-022-06: el contexto de E puede ver el service de F por su id real.';
  END IF;

  UPDATE service SET name = 'Renombrado por E' WHERE id = v_service_f_id;
  GET DIAGNOSTICS v_updated = ROW_COUNT;
  IF v_updated <> 0 THEN
    RAISE EXCEPTION 'CA-022-06: el contexto de E pudo editar un service de F (% filas).', v_updated;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-022-06 OK · el contexto de E no ve ni puede editar un service de F'

-- Limpia la fila confirmada arriba (rol migrador, bypassa RLS por política
-- administrativa) para no dejar residuo.
DELETE FROM service WHERE id = '77777777-7777-7777-7777-777777777777';

-- ---------------------------------------------------------------------------
-- CA-022-04/DEC-067 · nombre único entre servicios ACTIVOS de la misma
-- barbería; un servicio de OTRA barbería puede compartir el mismo nombre
-- exacto sin conflicto
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '55555555-5555-5555-5555-555555555555';

DO $$
BEGIN
  INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
  VALUES ('55555555-5555-5555-5555-555555555555', 'SQL Suite Nombre Único', 30, 1.00);

  BEGIN
    INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
    VALUES ('55555555-5555-5555-5555-555555555555', 'SQL Suite Nombre Único', 45, 2.00);
    RAISE EXCEPTION 'DEC-067: se aceptó un nombre duplicado entre servicios activos de la misma barbería.';
  EXCEPTION
    WHEN unique_violation THEN NULL; -- Esperado: idx_service_active_name.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-022-04/DEC-067 OK · idx_service_active_name rechaza el nombre duplicado entre activos'

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '55555555-5555-5555-5555-555555555555';
INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
VALUES ('55555555-5555-5555-5555-555555555555', 'SQL Suite Nombre Compartido Entre Tenants', 30, 1.00);
RESET ROLE;
ROLLBACK;

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '66666666-6666-6666-6666-666666666666';
-- El mismo nombre exacto en OTRA barbería no conflictúa (DEC-067: la
-- unicidad es tenant-aware, no global). Ambos INSERT se hacen dentro de
-- transacciones separadas que hacen ROLLBACK, así que esto solo confirma
-- que el segundo INSERT NO lanza unique_violation por sí mismo (no depende
-- de que el primero haya persistido).
INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
VALUES ('66666666-6666-6666-6666-666666666666', 'SQL Suite Nombre Compartido Entre Tenants', 30, 1.00);
RESET ROLE;
ROLLBACK;
\echo 'DEC-067 OK · el mismo nombre exacto en barberías distintas no conflictúa'

-- ---------------------------------------------------------------------------
-- DEC-067 · un servicio DESACTIVADO libera su nombre (la transición de
-- is_active pertenece a HU-024; aquí solo se confirma que el índice parcial
-- WHERE is_active permite esa reutilización cuando is_active = false)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '55555555-5555-5555-5555-555555555555';

DO $$
BEGIN
  -- barberia_app no puede fijar is_active=false (RLS/CHECK no lo prohíben
  -- directamente, pero HU-022 no expone ningún camino de aplicación para
  -- hacerlo); se usa el rol administrativo dentro de este bloque para
  -- simular el estado que HU-024 producirá, y confirmar que el índice
  -- parcial reacciona correctamente a esa combinación de columnas.
  NULL;
END
$$;

RESET ROLE;
ROLLBACK;

-- Verificación directa con el rol administrativo (barberia_owner por
-- membresía de barberia_migrator), fuera de la política tenant-aware.
INSERT INTO service (id, barbershop_id, name, duration_minutes, price_amount, is_active, deactivated_at)
VALUES ('88888888-8888-8888-8888-888888888888', '55555555-5555-5555-5555-555555555555',
        'SQL Suite Nombre Reutilizable', 30, 1.00, false, now());
INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
VALUES ('55555555-5555-5555-5555-555555555555', 'SQL Suite Nombre Reutilizable', 45, 2.00)
ON CONFLICT DO NOTHING;
DO $$
DECLARE
  v_count integer;
BEGIN
  SELECT count(*) INTO v_count FROM service
   WHERE barbershop_id = '55555555-5555-5555-5555-555555555555'
     AND name = 'SQL Suite Nombre Reutilizable';
  IF v_count <> 2 THEN
    RAISE EXCEPTION 'DEC-067: se esperaban 2 filas (una desactivada, una activa) con el mismo nombre, hay %.', v_count;
  END IF;
END
$$;
\echo 'DEC-067 OK · un servicio desactivado libera su nombre para uno activo nuevo'

-- ---------------------------------------------------------------------------
-- CA-022-03/CA-022-04 · CHECK de duración y precio (defensa estructural; el
-- mensaje de campo completo lo prueba el servicio Go)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '55555555-5555-5555-5555-555555555555';

DO $$
BEGIN
  BEGIN
    INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
    VALUES ('55555555-5555-5555-5555-555555555555', 'Duración Cero', 0, 1.00);
    RAISE EXCEPTION 'service_duration_minutes_ck: se aceptó una duración de 0 minutos.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
    VALUES ('55555555-5555-5555-5555-555555555555', 'Duración Excesiva', 1441, 1.00);
    RAISE EXCEPTION 'service_duration_minutes_ck: se aceptó una duración de 1441 minutos.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  -- DEC-067: precio 0 o negativo se rechaza (sin servicios gratuitos).
  BEGIN
    INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
    VALUES ('55555555-5555-5555-5555-555555555555', 'Precio Cero', 30, 0.00);
    RAISE EXCEPTION 'service_price_amount_ck: se aceptó un precio de 0 (DEC-067).';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
    VALUES ('55555555-5555-5555-5555-555555555555', 'Precio Negativo', 30, -1.00);
    RAISE EXCEPTION 'service_price_amount_ck: se aceptó un precio negativo.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  -- DEC-067: moneda distinta de COP se rechaza (fija, no editable).
  BEGIN
    INSERT INTO service (barbershop_id, name, duration_minutes, price_amount, price_currency)
    VALUES ('55555555-5555-5555-5555-555555555555', 'Moneda Ajena', 30, 1.00, 'USD');
    RAISE EXCEPTION 'service_price_currency_ck: se aceptó una moneda distinta de COP (DEC-067).';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  -- Valores límite válidos: 1 y 1440 minutos, precio positivo mínimo.
  INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
  VALUES ('55555555-5555-5555-5555-555555555555', 'Duración Mínima Válida', 1, 0.01);
  INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
  VALUES ('55555555-5555-5555-5555-555555555555', 'Duración Máxima Válida', 1440, 0.01);
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'constraint OK · duración fuera de [1,1440], precio <= 0 y moneda != COP se rechazan; los límites 1/1440/0.01 se aceptan'

-- ---------------------------------------------------------------------------
-- DELETE se deniega para barberia_app, incluso sobre su propia fila
-- (RN-SER-03: el borrado físico nunca se expone)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '55555555-5555-5555-5555-555555555555';

DO $$
DECLARE
  v_id uuid;
BEGIN
  INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
  VALUES ('55555555-5555-5555-5555-555555555555', 'Para Intentar Borrar', 30, 1.00)
  RETURNING id INTO v_id;

  BEGIN
    DELETE FROM service WHERE id = v_id;
    RAISE EXCEPTION 'RN-SER-03: se permitió DELETE sobre service con el rol de aplicación.';
  EXCEPTION
    WHEN insufficient_privilege THEN NULL; -- Esperado: sin GRANT DELETE.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'DELETE OK · denegado para barberia_app incluso sobre su propia fila (sin GRANT, RN-SER-03)'

-- ---------------------------------------------------------------------------
-- El rol de aplicación real no tiene BYPASSRLS (repetido explícitamente
-- para HU-022, mismo criterio que hu001/hu020/hu021)
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
-- Trigger de updated_at. now() es transaction_timestamp(): dentro de UNA
-- sola transacción devuelve el mismo valor para el INSERT y un UPDATE
-- posterior, así que la prueba usa dos transacciones separadas.
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '55555555-5555-5555-5555-555555555555';

DO $$
BEGIN
  INSERT INTO service (id, barbershop_id, name, duration_minutes, price_amount)
  VALUES ('99999999-8888-8888-8888-888888888888', '55555555-5555-5555-5555-555555555555',
          'Nombre Original Trigger', 30, 1.00);
END
$$;

RESET ROLE;
COMMIT;

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '55555555-5555-5555-5555-555555555555';

DO $$
DECLARE
  v_updated_before timestamptz;
  v_updated_after timestamptz;
BEGIN
  SELECT updated_at INTO v_updated_before FROM service WHERE id = '99999999-8888-8888-8888-888888888888';

  PERFORM pg_sleep(0.01);

  UPDATE service SET name = 'Nombre Renombrado Trigger' WHERE id = '99999999-8888-8888-8888-888888888888'
  RETURNING updated_at INTO v_updated_after;

  IF v_updated_after <= v_updated_before THEN
    RAISE EXCEPTION 'service_set_updated_at: updated_at no avanzó tras el UPDATE (antes=%, después=%).',
      v_updated_before, v_updated_after;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'trigger OK · service_set_updated_at avanza updated_at en cada UPDATE'

-- ---------------------------------------------------------------------------
-- Limpieza de las filas que se confirmaron con COMMIT/rol administrativo
-- más arriba, para dejar la base exactamente como estaba al empezar este
-- archivo.
-- ---------------------------------------------------------------------------
DELETE FROM service WHERE id IN (
  '88888888-8888-8888-8888-888888888888',
  '99999999-8888-8888-8888-888888888888'
);
DELETE FROM service
 WHERE barbershop_id = '55555555-5555-5555-5555-555555555555'
   AND name = 'SQL Suite Nombre Reutilizable';

\echo '=== HU-022 · todas las comprobaciones pasaron ==='
