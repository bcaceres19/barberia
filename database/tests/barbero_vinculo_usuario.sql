-- Pruebas SQL del issue #322 (DEC-100): vínculo opcional barber.staff_user_id.
-- Cubren a nivel de base de datos la unicidad parcial (un usuario es como
-- máximo un barbero y un barbero como máximo un usuario), la clave foránea
-- compuesta que impide vincular un usuario de otra barbería y el aislamiento
-- RLS entre dos barberías. La orquestación (seleccionar, cambiar, quitar) vive
-- en el servicio Go (internal/modules/staff), no aquí.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/barbero_vinculo_usuario.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- (estandar-base-datos.md §9.8). Todo se revierte con ROLLBACK.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== Issue #322 · pruebas del vínculo barbero–usuario ==='

-- ---------------------------------------------------------------------------
-- Columna nullable: sin vínculo por defecto, nunca inferido
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_col record;
BEGIN
  SELECT is_nullable, data_type INTO v_col
  FROM information_schema.columns
  WHERE table_schema = 'public' AND table_name = 'barber' AND column_name = 'staff_user_id';

  IF NOT FOUND THEN
    RAISE EXCEPTION 'barber.staff_user_id debe existir.';
  END IF;
  IF v_col.is_nullable <> 'YES' OR v_col.data_type <> 'uuid' THEN
    RAISE EXCEPTION 'barber.staff_user_id debe ser uuid NULL (leyó % %).', v_col.data_type, v_col.is_nullable;
  END IF;
  IF EXISTS (SELECT 1 FROM barber WHERE staff_user_id IS NOT NULL) THEN
    RAISE EXCEPTION 'ningún barbero existente debe tener vínculo inferido.';
  END IF;
END
$$;
\echo 'columna OK · uuid NULL y sin vínculos inferidos'

-- ---------------------------------------------------------------------------
-- Vincular, no duplicar y poder tener varios barberos sin usuario
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('c3220001-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111', 'Barbero vínculo 1'),
  ('c3220001-0000-4000-8000-000000000002', '11111111-1111-1111-1111-111111111111', 'Barbero vínculo 2'),
  ('c3220001-0000-4000-8000-000000000003', '11111111-1111-1111-1111-111111111111', 'Barbero vínculo 3');

DO $$
BEGIN
  -- Varios barberos sin usuario conviven: NULL no cuenta como duplicado.
  IF (SELECT count(*) FROM barber WHERE staff_user_id IS NULL
        AND id::text LIKE 'c3220001-%') <> 3 THEN
    RAISE EXCEPTION 'tres barberos sin usuario deben convivir.';
  END IF;

  UPDATE barber SET staff_user_id = 'aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2'
   WHERE id = 'c3220001-0000-4000-8000-000000000001';

  -- Un usuario no puede ser dos barberos.
  BEGIN
    UPDATE barber SET staff_user_id = 'aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2'
     WHERE id = 'c3220001-0000-4000-8000-000000000002';
    RAISE EXCEPTION 'un usuario no debe poder vincularse a dos barberos.';
  EXCEPTION WHEN unique_violation THEN
    NULL;
  END;

  -- Liberar el vínculo permite tomarlo con otro barbero.
  UPDATE barber SET staff_user_id = NULL
   WHERE id = 'c3220001-0000-4000-8000-000000000001';
  UPDATE barber SET staff_user_id = 'aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2'
   WHERE id = 'c3220001-0000-4000-8000-000000000002';
  IF (SELECT count(*) FROM barber WHERE staff_user_id = 'aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2') <> 1 THEN
    RAISE EXCEPTION 'tras liberar, el usuario debe quedar vinculado a un solo barbero.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'unicidad OK · un usuario, un barbero; NULL repetible; liberar y retomar'

-- ---------------------------------------------------------------------------
-- La clave foránea compuesta rechaza un usuario de otra barbería
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('c3220002-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111', 'Barbero A');

DO $$
BEGIN
  BEGIN
    UPDATE barber SET staff_user_id = 'bbbbbbb1-bbbb-bbbb-bbbb-bbbbbbbbbbb1'
     WHERE id = 'c3220002-0000-4000-8000-000000000001';
    RAISE EXCEPTION 'un usuario de otra barbería no debe poder vincularse.';
  EXCEPTION WHEN foreign_key_violation THEN
    NULL;
  END;

  -- Un usuario inexistente tampoco.
  BEGIN
    UPDATE barber SET staff_user_id = 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee'
     WHERE id = 'c3220002-0000-4000-8000-000000000001';
    RAISE EXCEPTION 'un usuario inexistente no debe poder vincularse.';
  EXCEPTION WHEN foreign_key_violation THEN
    NULL;
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'FK compuesta OK · usuario de otra barbería o inexistente rechazado'

-- ---------------------------------------------------------------------------
-- Aislamiento RLS: la barbería B ni ve ni modifica el vínculo de la A
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

INSERT INTO barber (id, barbershop_id, full_name, staff_user_id) VALUES
  ('c3220003-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111',
   'Barbero A vinculado', 'aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2');

RESET ROLE;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '22222222-2222-2222-2222-222222222222';

DO $$
DECLARE
  v_rows integer;
BEGIN
  IF EXISTS (SELECT 1 FROM barber WHERE id = 'c3220003-0000-4000-8000-000000000001') THEN
    RAISE EXCEPTION 'la barbería B no debe ver al barbero de la A.';
  END IF;

  UPDATE barber SET staff_user_id = NULL WHERE id = 'c3220003-0000-4000-8000-000000000001';
  GET DIAGNOSTICS v_rows = ROW_COUNT;
  IF v_rows <> 0 THEN
    RAISE EXCEPTION 'la barbería B no debe poder liberar el vínculo de la A (afectó % filas).', v_rows;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'RLS OK · la barbería B no ve ni libera el vínculo de la A'

\echo '=== Issue #322 · todas las comprobaciones pasaron ==='
