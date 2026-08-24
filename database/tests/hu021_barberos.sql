-- Pruebas SQL de HU-021: registro y listado de barberos. Cubren CA-021-01,
-- CA-021-02, CA-021-05, CA-021-06 y CA-021-07 a nivel de base de datos
-- (esquema, RLS, grants); la validación de fullName, la paginación por
-- cursor y el protocolo de idempotencia completo viven en el servicio Go
-- (internal/modules/staff), probados aparte contra PostgreSQL real.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/hu021_barberos.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- (estandar-base-datos.md §9.8), igual que hu001_aislamiento_rls.sql /
-- hu020_configuracion_barberia.sql.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== HU-021 · pruebas de registro y listado de barberos ==='

-- ---------------------------------------------------------------------------
-- Esquema mínimo: columnas, tipos y CHECK esperados (DEC-047: sin active,
-- deleted_at, sort_order ni staff_user_id)
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_offenders text;
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'barber'
  ) THEN
    RAISE EXCEPTION 'barber debe existir.';
  END IF;

  SELECT string_agg(column_name, ', ') INTO v_offenders
  FROM information_schema.columns
  WHERE table_schema = 'public' AND table_name = 'barber'
    AND column_name IN ('active', 'deleted_at', 'sort_order', 'staff_user_id', 'is_active');
  IF v_offenders IS NOT NULL THEN
    RAISE EXCEPTION 'DEC-047/CA-021-07: barber tiene columnas fuera de alcance: %', v_offenders;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barber'
      AND column_name = 'full_name' AND is_nullable = 'NO'
  ) THEN
    RAISE EXCEPTION 'barber.full_name debe existir y ser NOT NULL.';
  END IF;
END
$$;
\echo 'esquema OK · barber tiene exactamente las columnas de HU-021, sin ciclo de vida'

-- ---------------------------------------------------------------------------
-- CA-021-06 · RLS forzada, sin BYPASSRLS, sin política/GRANT de DELETE
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_relrowsecurity boolean;
  v_relforcerowsecurity boolean;
BEGIN
  SELECT relrowsecurity, relforcerowsecurity INTO v_relrowsecurity, v_relforcerowsecurity
  FROM pg_class
  WHERE relname = 'barber' AND relnamespace = 'public'::regnamespace;

  IF NOT v_relrowsecurity OR NOT v_relforcerowsecurity THEN
    RAISE EXCEPTION 'CA-021-06: barber debe tener RLS habilitada Y forzada.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM pg_roles WHERE rolname = 'barberia_app' AND rolbypassrls
  ) THEN
    RAISE EXCEPTION 'CA-021-06: barberia_app no puede tener BYPASSRLS.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.table_privileges
    WHERE table_schema = 'public' AND table_name = 'barber'
      AND grantee = 'barberia_app' AND privilege_type = 'DELETE'
  ) THEN
    RAISE EXCEPTION 'CA-021-07/DEC-047: barberia_app no debe tener GRANT DELETE sobre barber.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'barber' AND cmd = 'DELETE'
  ) THEN
    RAISE EXCEPTION 'CA-021-07/DEC-047: no debe existir ninguna política DELETE sobre barber.';
  END IF;
END
$$;
\echo 'CA-021-06/07 OK · RLS forzada, sin BYPASSRLS, sin GRANT ni política DELETE'

-- ---------------------------------------------------------------------------
-- Ausencia de contexto de tenant: falla cerrado, no devuelve el conjunto
-- completo. DEBE ejecutarse antes de cualquier SET LOCAL app.barbershop_id
-- de este archivo: una vez que la sesión usa ese nombre de parámetro
-- personalizado al menos una vez, PostgreSQL lo recuerda por el resto de la
-- sesión y current_setting() deja de lanzar "unrecognized configuration
-- parameter" tras el ROLLBACK (vuelve a su valor de arranque, cadena
-- vacía), aunque la garantía real en producción (una conexión nueva del
-- pool, jamás usada antes) siga intacta. Mismo orden que
-- hu001_aislamiento_rls.sql (CA-001-03 antes de su primer SET LOCAL).
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
BEGIN
  BEGIN
    PERFORM count(*) FROM barber;
    RAISE EXCEPTION 'sin contexto: se pudo leer barber sin app.barbershop_id fijado.';
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
\echo 'sin contexto OK · leer barber sin app.barbershop_id fijado falla cerrado'

-- ---------------------------------------------------------------------------
-- CA-021-01/02 · Alta dentro del propio tenant: una fila para barbería
-- unipersonal, cuatro filas distintas para un equipo
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_id1 uuid;
  v_id2 uuid;
  v_id3 uuid;
  v_id4 uuid;
  v_count integer;
BEGIN
  INSERT INTO barber (barbershop_id, full_name)
  VALUES ('11111111-1111-1111-1111-111111111111', 'Barbero Único')
  RETURNING id INTO v_id1;

  SELECT count(*) INTO v_count FROM barber WHERE barbershop_id = '11111111-1111-1111-1111-111111111111';
  IF v_count <> 1 THEN
    RAISE EXCEPTION 'CA-021-01: barbería unipersonal debe tener exactamente 1 fila barber, tiene %.', v_count;
  END IF;

  -- CA-021-02: agregar tres más produce cuatro recursos distintos, mismo
  -- schema/tabla, sin caso especial de "barbero único".
  INSERT INTO barber (barbershop_id, full_name) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Segundo Barbero')
    RETURNING id INTO v_id2;
  INSERT INTO barber (barbershop_id, full_name) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Tercer Barbero')
    RETURNING id INTO v_id3;
  -- Nombres duplicados están permitidos: HU-021 no impone unicidad.
  INSERT INTO barber (barbershop_id, full_name) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Segundo Barbero')
    RETURNING id INTO v_id4;

  SELECT count(*) INTO v_count FROM barber WHERE barbershop_id = '11111111-1111-1111-1111-111111111111';
  IF v_count <> 4 THEN
    RAISE EXCEPTION 'CA-021-02: se esperaban 4 barberos distintos, hay %.', v_count;
  END IF;

  IF v_id1 = v_id2 OR v_id1 = v_id3 OR v_id1 = v_id4 OR v_id2 = v_id3 OR v_id2 = v_id4 OR v_id3 = v_id4 THEN
    RAISE EXCEPTION 'CA-021-02: los cuatro identificadores deben ser distintos.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-021-01/02 OK · barbería unipersonal y de equipo usan la misma tabla, sin rama especial'

-- ---------------------------------------------------------------------------
-- CA-021-06 · La base rechaza insertar/relacionar un barber con un
-- barbershop_id distinto del contexto de la transacción, incluso con el rol
-- de aplicación real
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  BEGIN
    INSERT INTO barber (barbershop_id, full_name)
    VALUES ('22222222-2222-2222-2222-222222222222', 'Intruso');
    RAISE EXCEPTION 'CA-021-06: se permitió insertar un barber con barbershop_id ajeno al contexto.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL; -- Esperado: barber_insert_tenant_policy rechaza el WITH CHECK.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-021-06 OK · INSERT con barbershop_id ajeno al contexto se rechaza (WITH CHECK de RLS)'

-- ---------------------------------------------------------------------------
-- CA-021-05 · El contexto de A no ve, no puede actualizar y no puede
-- "adoptar" (reasignar hacia sí) un barber de B
-- ---------------------------------------------------------------------------
-- Este bloque SÍ confirma (COMMIT): la comprobación siguiente necesita leer
-- el identificador real desde una sesión/transacción distinta a la que lo
-- creó, para que "el contexto de A no ve nada" no dependa de ver su propia
-- escritura todavía no confirmada. Se limpia explícitamente al final de esta
-- sección (superusuario, bypassa RLS) para no dejar residuo.
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '22222222-2222-2222-2222-222222222222';

INSERT INTO barber (id, barbershop_id, full_name)
VALUES ('88888888-8888-8888-8888-888888888888', '22222222-2222-2222-2222-222222222222', 'Barbero de B');

RESET ROLE;
COMMIT;

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_barber_b_id uuid := '88888888-8888-8888-8888-888888888888';
  v_visible integer;
BEGIN
  -- Identificador REAL de un barbero de B (creado y confirmado arriba),
  -- consultado directamente por id bajo el contexto de A: no debe
  -- encontrarse (RLS, no solo "adivinar mal el id").
  SELECT count(*) INTO v_visible FROM barber WHERE id = v_barber_b_id;
  IF v_visible <> 0 THEN
    RAISE EXCEPTION 'CA-021-05: el contexto de A puede ver el barber de B por su id real.';
  END IF;

  SELECT count(*) INTO v_visible FROM barber WHERE barbershop_id = '22222222-2222-2222-2222-222222222222';
  IF v_visible <> 0 THEN
    RAISE EXCEPTION 'CA-021-05: el contexto de A ve % filas de la barbería de B.', v_visible;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-021-05 OK · el contexto de A no ve ningún barber de B'

-- Repite el intento de escritura cruzada usando el mismo identificador REAL
-- del barbero de B, para que el rechazo del UPDATE de abajo sea la parte
-- bajo prueba, no la ausencia de una fila que adivinar.
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_barber_b_id uuid := '88888888-8888-8888-8888-888888888888';
  v_updated integer;
BEGIN
  UPDATE barber SET full_name = 'Renombrado por A' WHERE id = v_barber_b_id;
  GET DIAGNOSTICS v_updated = ROW_COUNT;
  IF v_updated <> 0 THEN
    RAISE EXCEPTION 'CA-021-05: el contexto de A pudo renombrar un barber de B (% filas).', v_updated;
  END IF;

  -- WITH CHECK también rechaza "adoptar" la fila de B moviéndola hacia el
  -- tenant de A, aunque de algún modo fuera visible.
  BEGIN
    UPDATE barber SET barbershop_id = '11111111-1111-1111-1111-111111111111'
    WHERE id = v_barber_b_id;
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL; -- Aceptable: bloqueado por RLS de una forma u otra.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-021-05 OK · UPDATE cruzado sobre un identificador real de B se rechaza (cero filas o WITH CHECK)'

-- ---------------------------------------------------------------------------
-- CA-021-03 (nivel base de datos) · full_name vacío/solo espacios/121+
-- caracteres se rechaza en el CHECK; el 422 completo (mensaje de campo) lo
-- prueba el servicio Go, esto es la defensa estructural.
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  BEGIN
    INSERT INTO barber (barbershop_id, full_name) VALUES ('11111111-1111-1111-1111-111111111111', '');
    RAISE EXCEPTION 'barber_full_name_ck: se aceptó un nombre vacío.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO barber (barbershop_id, full_name) VALUES ('11111111-1111-1111-1111-111111111111', '   ');
    RAISE EXCEPTION 'barber_full_name_ck: se aceptó un nombre compuesto solo por espacios.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO barber (barbershop_id, full_name)
    VALUES ('11111111-1111-1111-1111-111111111111', repeat('a', 121));
    RAISE EXCEPTION 'barber_full_name_ck: se aceptó un nombre de 121 caracteres.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  -- 120 caracteres exactos SÍ se aceptan (límite inclusivo).
  INSERT INTO barber (barbershop_id, full_name)
  VALUES ('11111111-1111-1111-1111-111111111111', repeat('a', 120));
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'constraint OK · barber_full_name_ck rechaza vacío/solo espacios/121+ y acepta exactamente 120'

-- ---------------------------------------------------------------------------
-- DELETE se deniega para barberia_app, incluso sobre su propia fila
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_id uuid;
BEGIN
  INSERT INTO barber (barbershop_id, full_name)
  VALUES ('11111111-1111-1111-1111-111111111111', 'Para Borrar')
  RETURNING id INTO v_id;

  BEGIN
    DELETE FROM barber WHERE id = v_id;
    RAISE EXCEPTION 'CA-021-07/DEC-047: se permitió DELETE sobre barber con el rol de aplicación.';
  EXCEPTION
    WHEN insufficient_privilege THEN NULL; -- Esperado: sin GRANT DELETE.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'DELETE OK · denegado para barberia_app incluso sobre su propia fila (sin GRANT, CA-021-07)'

-- ---------------------------------------------------------------------------
-- El rol de aplicación real no tiene BYPASSRLS (repetido explícitamente
-- para HU-021, mismo criterio que hu001/hu020)
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
-- posterior, así que la prueba usa dos transacciones separadas (mismo
-- patrón real: alta y renombrado siempre son dos InTenantTx distintas,
-- nunca la misma).
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  INSERT INTO barber (id, barbershop_id, full_name)
  VALUES ('99999999-9999-9999-9999-999999999999', '11111111-1111-1111-1111-111111111111', 'Nombre Original');
END
$$;

RESET ROLE;
COMMIT;

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_updated_before timestamptz;
  v_updated_after timestamptz;
BEGIN
  SELECT updated_at INTO v_updated_before FROM barber WHERE id = '99999999-9999-9999-9999-999999999999';

  PERFORM pg_sleep(0.01);

  UPDATE barber SET full_name = 'Nombre Renombrado' WHERE id = '99999999-9999-9999-9999-999999999999'
  RETURNING updated_at INTO v_updated_after;

  IF v_updated_after <= v_updated_before THEN
    RAISE EXCEPTION 'barber_set_updated_at: updated_at no avanzó tras el UPDATE (antes=%, después=%).',
      v_updated_before, v_updated_after;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'trigger OK · barber_set_updated_at avanza updated_at en cada UPDATE'

-- ---------------------------------------------------------------------------
-- Limpieza de las dos filas que se confirmaron con COMMIT más arriba (el
-- barbero de B usado para CA-021-05 y el barbero fijo usado para el trigger
-- de updated_at): con el rol de sesión sin RLS forzada, para dejar la base
-- exactamente como estaba al empezar este archivo.
-- ---------------------------------------------------------------------------
DELETE FROM barber WHERE id IN (
  '88888888-8888-8888-8888-888888888888',
  '99999999-9999-9999-9999-999999999999'
);

\echo '=== HU-021 · todas las comprobaciones pasaron ==='
