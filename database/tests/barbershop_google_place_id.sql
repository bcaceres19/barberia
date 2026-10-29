-- Pruebas SQL de DEC-129 / HU-101: barbershop.google_place_id. Cubren a nivel de
-- base de datos el valor por defecto (sin ficha), la forma aceptada (letras,
-- dígitos, `_` y `-`, de 1 a 200 caracteres), el rechazo de cualquier otra
-- (vacío, espacios, `/`, `?`, `%`, acentos, más de 200) y el aislamiento RLS
-- entre dos barberías. Consultar a Google y cachear viven en el servicio Go
-- (internal/modules/publicbooking/google).
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/barbershop_google_place_id.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- (estandar-base-datos.md §9.8). Todo se revierte con ROLLBACK.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== DEC-129 · pruebas de google_place_id ==='

-- ---------------------------------------------------------------------------
-- Defaults: sin ficha
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v text;
BEGIN
  SELECT google_place_id INTO v FROM barbershop WHERE id = '11111111-1111-1111-1111-111111111111';
  IF v IS NOT NULL THEN
    RAISE EXCEPTION 'una barbería existente no debe tener ficha de Google (leyó %).', v;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'defaults OK · sin ficha de Google'

-- ---------------------------------------------------------------------------
-- Formas válidas y quitar la ficha
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_ok text;
  v text;
BEGIN
  FOREACH v_ok IN ARRAY ARRAY['ChIJN1t_tDeuEmsRUsoyG83frY4', 'a', 'A-b_9', repeat('x', 200)]
  LOOP
    UPDATE barbershop SET google_place_id = v_ok WHERE id = '11111111-1111-1111-1111-111111111111';
    SELECT google_place_id INTO v FROM barbershop WHERE id = '11111111-1111-1111-1111-111111111111';
    IF v <> v_ok THEN
      RAISE EXCEPTION 'no se conservó el Place ID válido %.', v_ok;
    END IF;
  END LOOP;

  UPDATE barbershop SET google_place_id = NULL WHERE id = '11111111-1111-1111-1111-111111111111';
  SELECT google_place_id INTO v FROM barbershop WHERE id = '11111111-1111-1111-1111-111111111111';
  IF v IS NOT NULL THEN
    RAISE EXCEPTION 'quitar la ficha debe dejar NULL.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'forma OK · Place ID válidos se guardan y la ficha puede quitarse'

-- ---------------------------------------------------------------------------
-- Restricción: cada forma inválida se rechaza en la base
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_bad text;
BEGIN
  FOREACH v_bad IN ARRAY ARRAY[
    '', ' ', 'ChIJ abc', 'ChIJ/abc', 'ChIJ?x=1', 'ChIJ%2F', E'ChIJ\nabc', 'Chíj', repeat('x', 201)
  ]
  LOOP
    BEGIN
      UPDATE barbershop SET google_place_id = v_bad WHERE id = '11111111-1111-1111-1111-111111111111';
      RAISE EXCEPTION 'se aceptó un Place ID inválido: [%]', v_bad;
    EXCEPTION
      WHEN check_violation THEN
        NULL; -- Esperado.
    END;
  END LOOP;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CHECK OK · formas inválidas rechazadas en la base'

-- ---------------------------------------------------------------------------
-- RLS: un tenant no lee ni modifica la ficha de otro
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_rows integer;
  v_visible integer;
BEGIN
  UPDATE barbershop SET google_place_id = 'Intrusion'
   WHERE id = '22222222-2222-2222-2222-222222222222';
  GET DIAGNOSTICS v_rows = ROW_COUNT;
  IF v_rows <> 0 THEN
    RAISE EXCEPTION 'RLS: el tenant A modificó la ficha del tenant B (% filas).', v_rows;
  END IF;

  SELECT count(*) INTO v_visible FROM barbershop
   WHERE id = '22222222-2222-2222-2222-222222222222';
  IF v_visible <> 0 THEN
    RAISE EXCEPTION 'RLS: el tenant A ve la barbería del tenant B.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'RLS OK · un tenant no lee ni modifica la ficha de otro'

\echo '=== DEC-129 · todas las pruebas pasaron ==='
