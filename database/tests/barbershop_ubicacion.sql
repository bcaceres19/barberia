-- Pruebas SQL de DEC-130 / HU-102: ubicación de la barbería (barbershop.
-- has_physical_location, address, latitude, longitude). Cubren a nivel de base
-- de datos los defaults, las restricciones (dirección ni vacía ni de más de 200
-- caracteres, rangos de coordenadas, coordenadas en pareja y solo con
-- dirección), que apagar el interruptor conserva el dato y el aislamiento RLS
-- entre dos barberías. Geocodificar y decidir qué se guarda viven en el
-- servicio Go (internal/modules/shops).
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/barbershop_ubicacion.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- (estandar-base-datos.md §9.8). Todo se revierte con ROLLBACK.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== DEC-130 · pruebas de la ubicación de barbershop ==='

-- ---------------------------------------------------------------------------
-- Defaults: interruptor activo, sin dirección ni coordenadas
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v record;
BEGIN
  SELECT has_physical_location, address, latitude, longitude INTO v
    FROM barbershop WHERE id = '11111111-1111-1111-1111-111111111111';
  IF v.has_physical_location IS NOT TRUE THEN
    RAISE EXCEPTION 'has_physical_location debe ser true por defecto (leyó %).', v.has_physical_location;
  END IF;
  IF v.address IS NOT NULL OR v.latitude IS NOT NULL OR v.longitude IS NOT NULL THEN
    RAISE EXCEPTION 'una barbería existente no debe tener dirección ni coordenadas.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'defaults OK · interruptor activo, sin dirección ni coordenadas'

-- ---------------------------------------------------------------------------
-- Guardar, apagar y reactivar: el dato se conserva (CA-102-05)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v record;
BEGIN
  UPDATE barbershop
     SET address = 'Calle 10 # 5-20, Bogotá', latitude = 4.598056, longitude = -74.076111
   WHERE id = '11111111-1111-1111-1111-111111111111';

  UPDATE barbershop SET has_physical_location = false
   WHERE id = '11111111-1111-1111-1111-111111111111';
  SELECT has_physical_location, address, latitude, longitude INTO v
    FROM barbershop WHERE id = '11111111-1111-1111-1111-111111111111';
  IF v.has_physical_location OR v.address IS NULL OR v.latitude <> 4.598056 OR v.longitude <> -74.076111 THEN
    RAISE EXCEPTION 'apagar el interruptor no debe borrar address ni coordenadas.';
  END IF;

  UPDATE barbershop SET has_physical_location = true
   WHERE id = '11111111-1111-1111-1111-111111111111';
  SELECT has_physical_location, address, latitude, longitude INTO v
    FROM barbershop WHERE id = '11111111-1111-1111-1111-111111111111';
  IF NOT v.has_physical_location OR v.address <> 'Calle 10 # 5-20, Bogotá' THEN
    RAISE EXCEPTION 'reactivar el interruptor debe devolver la dirección guardada.';
  END IF;

  -- Borrar la dirección exige borrar también las coordenadas.
  UPDATE barbershop SET address = NULL, latitude = NULL, longitude = NULL
   WHERE id = '11111111-1111-1111-1111-111111111111';
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'ciclo OK · apagar conserva address y coordenadas, reactivar las devuelve'

-- ---------------------------------------------------------------------------
-- Restricciones: cada estado inválido se rechaza en la base
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_set text;
BEGIN
  FOREACH v_set IN ARRAY ARRAY[
    $s$address = ''$s$,
    $s$address = '   '$s$,
    $s$address = repeat('x', 201)$s$,
    $s$address = 'Calle 1', latitude = 91, longitude = 0$s$,
    $s$address = 'Calle 1', latitude = -91, longitude = 0$s$,
    $s$address = 'Calle 1', latitude = 0, longitude = 181$s$,
    $s$address = 'Calle 1', latitude = 0, longitude = -181$s$,
    $s$address = 'Calle 1', latitude = 4.5, longitude = NULL$s$,
    $s$address = 'Calle 1', latitude = NULL, longitude = -74.0$s$,
    $s$address = NULL, latitude = 4.5, longitude = -74.0$s$
  ]
  LOOP
    BEGIN
      EXECUTE format('UPDATE barbershop SET %s WHERE id = %L', v_set,
                     '11111111-1111-1111-1111-111111111111');
      RAISE EXCEPTION 'se aceptó un estado inválido: %', v_set;
    EXCEPTION
      WHEN check_violation THEN
        NULL; -- Esperado.
    END;
  END LOOP;

  -- Los límites exactos sí son válidos.
  UPDATE barbershop
     SET address = repeat('ñ', 200), latitude = 90, longitude = -180
   WHERE id = '11111111-1111-1111-1111-111111111111';
  UPDATE barbershop
     SET address = 'Calle 1', latitude = -90, longitude = 180
   WHERE id = '11111111-1111-1111-1111-111111111111';
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CHECK OK · dirección, rangos y pareja de coordenadas se validan en la base'

-- ---------------------------------------------------------------------------
-- RLS: un tenant no lee ni modifica la ubicación de otro
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_rows integer;
  v_visible integer;
BEGIN
  UPDATE barbershop SET address = 'Intrusión', latitude = 1, longitude = 1
   WHERE id = '22222222-2222-2222-2222-222222222222';
  GET DIAGNOSTICS v_rows = ROW_COUNT;
  IF v_rows <> 0 THEN
    RAISE EXCEPTION 'RLS: el tenant A modificó la ubicación del tenant B (% filas).', v_rows;
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
\echo 'RLS OK · un tenant no lee ni modifica la ubicación de otro'

\echo '=== DEC-130 · todas las pruebas pasaron ==='
