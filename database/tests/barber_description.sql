-- Pruebas SQL de DEC-127: descripción opcional del barbero (barber.description).
-- Cubren a nivel de base de datos la nulabilidad, la restricción
-- barber_description_ck (ni vacía ni de solo espacios, máximo 280 caracteres) y
-- el aislamiento RLS entre dos barberías. Normalizar (recortar, vacío = NULL) y
-- decidir qué se escribe viven en el servicio Go (internal/modules/staff).
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/barber_description.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- (estandar-base-datos.md §9.8). Todo se revierte con ROLLBACK.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== DEC-127 · pruebas de barber.description ==='

-- ---------------------------------------------------------------------------
-- Columna nullable de texto: sin descripción por defecto
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_col record;
BEGIN
  SELECT is_nullable, data_type INTO v_col
  FROM information_schema.columns
  WHERE table_schema = 'public' AND table_name = 'barber' AND column_name = 'description';

  IF NOT FOUND THEN
    RAISE EXCEPTION 'barber.description debe existir.';
  END IF;
  IF v_col.is_nullable <> 'YES' OR v_col.data_type <> 'text' THEN
    RAISE EXCEPTION 'barber.description debe ser text NULL (leyó % %).', v_col.data_type, v_col.is_nullable;
  END IF;
  IF EXISTS (SELECT 1 FROM barber WHERE description IS NOT NULL AND btrim(description) = '') THEN
    RAISE EXCEPTION 'ningún barbero puede tener una descripción vacía.';
  END IF;
END
$$;
\echo 'columna OK · text NULL'

-- ---------------------------------------------------------------------------
-- Restricción: NULL y 1..280 caracteres válidos; vacía, espacios o 281 no.
-- (btrim recorta solo espacios: tabuladores y saltos de línea los recorta la
-- aplicación antes de guardar, internal/modules/staff.NormalizeDescription.)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

INSERT INTO barber (id, barbershop_id, full_name, description) VALUES
  ('d1270001-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111', 'Barbero sin descripción', NULL),
  ('d1270001-0000-4000-8000-000000000002', '11111111-1111-1111-1111-111111111111', 'Barbero con descripción', 'Fade, barba y navaja.'),
  ('d1270001-0000-4000-8000-000000000003', '11111111-1111-1111-1111-111111111111', 'Barbero con 280', repeat('ñ', 280));

DO $$
DECLARE
  v_rejected boolean;
  v_value text;
BEGIN
  FOREACH v_value IN ARRAY ARRAY['', '   ', repeat('a', 281)] LOOP
    v_rejected := false;
    BEGIN
      INSERT INTO barber (barbershop_id, full_name, description)
      VALUES ('11111111-1111-1111-1111-111111111111', 'Descripción inválida', v_value);
    EXCEPTION WHEN check_violation THEN
      v_rejected := true;
    END;
    IF NOT v_rejected THEN
      RAISE EXCEPTION 'barber_description_ck debió rechazar una descripción de % caracteres.', char_length(v_value);
    END IF;
  END LOOP;

  v_rejected := false;
  BEGIN
    UPDATE barber SET description = '' WHERE id = 'd1270001-0000-4000-8000-000000000002';
  EXCEPTION WHEN check_violation THEN
    v_rejected := true;
  END;
  IF NOT v_rejected THEN
    RAISE EXCEPTION 'barber_description_ck debió rechazar vaciar una descripción con cadena vacía (se quita con NULL).';
  END IF;

  UPDATE barber SET description = NULL WHERE id = 'd1270001-0000-4000-8000-000000000002';
  IF (SELECT description FROM barber WHERE id = 'd1270001-0000-4000-8000-000000000002') IS NOT NULL THEN
    RAISE EXCEPTION 'quitar la descripción con NULL debe funcionar.';
  END IF;
END
$$;
\echo 'restricción OK · NULL y 1..280 caracteres; vacía, espacios y 281 rechazadas'
ROLLBACK;

-- ---------------------------------------------------------------------------
-- RLS: la barbería B no ve ni modifica la descripción de la A
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

INSERT INTO barber (id, barbershop_id, full_name, description) VALUES
  ('d1270002-0000-4000-8000-000000000001', '11111111-1111-1111-1111-111111111111', 'Barbero A', 'Descripción de A');

SET LOCAL app.barbershop_id = '22222222-2222-2222-2222-222222222222';

DO $$
DECLARE
  v_rows integer;
BEGIN
  IF EXISTS (SELECT 1 FROM barber WHERE id = 'd1270002-0000-4000-8000-000000000001') THEN
    RAISE EXCEPTION 'la barbería B no debe ver el barbero (ni la descripción) de la A.';
  END IF;

  UPDATE barber SET description = 'Pisada por B' WHERE id = 'd1270002-0000-4000-8000-000000000001';
  GET DIAGNOSTICS v_rows = ROW_COUNT;
  IF v_rows <> 0 THEN
    RAISE EXCEPTION 'la barbería B no debe poder modificar la descripción de un barbero de la A.';
  END IF;
END
$$;
\echo 'RLS OK · la descripción no cruza barberías'
ROLLBACK;

\echo '=== DEC-127 · todas las pruebas pasaron ==='
