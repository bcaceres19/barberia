-- Pruebas SQL de DEC-131 / HU-103: servicio a domicilio.
-- Cubren a nivel de base de datos los valores por defecto (barbería sin domicilio,
-- turno en el local), `appointment_home_service_ck` (dirección completa o ausente,
-- nunca parcial, con rangos y longitud) y el aislamiento RLS de la dirección entre
-- dos barberías. Qué se acepta desde la reserva pública vive en el servicio Go.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/appointment_domicilio.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- (estandar-base-datos.md §9.8). Crea sus propios datos y todo se revierte con
-- ROLLBACK. Cada comprobación falla lanzando una excepción; si el archivo
-- termina, pasó.

\set ON_ERROR_STOP on

\echo '=== DEC-131 · pruebas de domicilio ==='

BEGIN;

INSERT INTO barbershop (id, name, timezone) VALUES
  ('01030001-0100-4100-8100-010300000001', 'Barbería domicilio A', 'America/Bogota'),
  ('01030002-0100-4100-8100-010300000002', 'Barbería domicilio B', 'America/Bogota');

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('01030011-0100-4100-8100-010300000011', '01030001-0100-4100-8100-010300000001', 'Barbero A'),
  ('01030021-0100-4100-8100-010300000021', '01030002-0100-4100-8100-010300000002', 'Barbero B');

INSERT INTO service (id, barbershop_id, name, duration_minutes, price_amount, price_currency) VALUES
  ('01030031-0100-4100-8100-010300000031', '01030001-0100-4100-8100-010300000001', 'Corte A', 30, 20000.00, 'COP'),
  ('01030032-0100-4100-8100-010300000032', '01030002-0100-4100-8100-010300000002', 'Corte B', 30, 20000.00, 'COP');

INSERT INTO customer (id, barbershop_id, full_name) VALUES
  ('01030041-0100-4100-8100-010300000041', '01030001-0100-4100-8100-010300000001', 'Cliente de prueba A'),
  ('01030042-0100-4100-8100-010300000042', '01030002-0100-4100-8100-010300000002', 'Cliente de prueba B');

-- ---------------------------------------------------------------------------
-- Valores por defecto
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  IF (SELECT offers_home_service FROM barbershop WHERE id = '01030001-0100-4100-8100-010300000001') THEN
    RAISE EXCEPTION 'una barbería nueva no debe ofrecer domicilio por defecto.';
  END IF;
END
$$;

-- Turno en el local: sin dirección, valores por defecto.
INSERT INTO appointment (
  id, barbershop_id, barber_id, service_id, customer_id, attendee_name,
  starts_at, ends_at, status, origin,
  service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
) VALUES
  ('01030051-0100-4100-8100-010300000051', '01030001-0100-4100-8100-010300000001', '01030011-0100-4100-8100-010300000011',
   '01030031-0100-4100-8100-010300000031', '01030041-0100-4100-8100-010300000041', 'Cliente de prueba A',
   '2099-01-05T10:00:00Z', '2099-01-05T10:30:00Z', 'confirmed', 'public', 'Corte A', 30, 20000.00, 'COP');

DO $$
DECLARE r record;
BEGIN
  SELECT is_home_service, service_address, service_latitude, service_longitude INTO r
    FROM appointment WHERE id = '01030051-0100-4100-8100-010300000051';
  IF r.is_home_service OR r.service_address IS NOT NULL
     OR r.service_latitude IS NOT NULL OR r.service_longitude IS NOT NULL THEN
    RAISE EXCEPTION 'un turno nuevo debe nacer en el local y sin datos de dirección.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- Domicilio completo: válido
-- ---------------------------------------------------------------------------
INSERT INTO appointment (
  id, barbershop_id, barber_id, service_id, customer_id, attendee_name,
  starts_at, ends_at, status, origin,
  service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot,
  is_home_service, service_address, service_latitude, service_longitude
) VALUES
  ('01030052-0100-4100-8100-010300000052', '01030001-0100-4100-8100-010300000001', '01030011-0100-4100-8100-010300000011',
   '01030031-0100-4100-8100-010300000031', '01030041-0100-4100-8100-010300000041', 'Cliente de prueba A',
   '2099-01-06T10:00:00Z', '2099-01-06T10:30:00Z', 'confirmed', 'public', 'Corte A', 30, 20000.00, 'COP',
   true, 'Calle 10 # 5-20, Bogotá', 4.598056, -74.076111),
  ('01030061-0100-4100-8100-010300000061', '01030002-0100-4100-8100-010300000002', '01030021-0100-4100-8100-010300000021',
   '01030032-0100-4100-8100-010300000032', '01030042-0100-4100-8100-010300000042', 'Cliente de prueba B',
   '2099-01-06T10:00:00Z', '2099-01-06T10:30:00Z', 'confirmed', 'public', 'Corte B', 30, 20000.00, 'COP',
   true, 'Carrera 7 # 71-21, Bogotá', 4.6486, -74.0628);

-- ---------------------------------------------------------------------------
-- Combinaciones inconsistentes: todas deben rechazarse (CA-103-04)
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_case text;
  v_rejected boolean;
  v_sql text;
  v_base text := $b$
    INSERT INTO appointment (
      id, barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot,
      is_home_service, service_address, service_latitude, service_longitude
    ) VALUES (
      gen_random_uuid(), '01030001-0100-4100-8100-010300000001', '01030011-0100-4100-8100-010300000011',
      '01030031-0100-4100-8100-010300000031', '01030041-0100-4100-8100-010300000041', 'Cliente de prueba A',
      '2099-02-01T10:00:00Z', '2099-02-01T10:30:00Z', 'confirmed', 'public', 'Corte A', 30, 20000.00, 'COP',
      %s)
  $b$;
  v_values text[] := ARRAY[
    'true, NULL, NULL, NULL',
    'true, ''Calle 1'', NULL, NULL',
    'true, ''Calle 1'', 4.6, NULL',
    'true, NULL, 4.6, -74.0',
    'true, '' '', 4.6, -74.0',
    'true, ''Calle 1'', 91, -74.0',
    'true, ''Calle 1'', 4.6, -181',
    'false, ''Calle 1'', NULL, NULL',
    'false, NULL, 4.6, NULL',
    'false, NULL, NULL, -74.0'
  ];
BEGIN
  FOREACH v_case IN ARRAY v_values LOOP
    v_sql := format(v_base, v_case);
    v_rejected := false;
    BEGIN
      EXECUTE v_sql;
    EXCEPTION WHEN check_violation THEN
      v_rejected := true;
    END;
    IF NOT v_rejected THEN
      RAISE EXCEPTION 'appointment_home_service_ck debió rechazar la combinación: %', v_case;
    END IF;
  END LOOP;

  -- Una dirección de más de 200 caracteres también se rechaza.
  v_rejected := false;
  BEGIN
    EXECUTE format(v_base, format('true, %L, 4.6, -74.0', repeat('x', 201)));
  EXCEPTION WHEN check_violation THEN
    v_rejected := true;
  END;
  IF NOT v_rejected THEN
    RAISE EXCEPTION 'appointment_home_service_ck debió rechazar una dirección de más de 200 caracteres.';
  END IF;
END
$$;

-- Convertir un turno existente en uno parcial también falla.
DO $$
DECLARE v_rejected boolean := false;
BEGIN
  BEGIN
    UPDATE appointment SET service_address = NULL WHERE id = '01030052-0100-4100-8100-010300000052';
  EXCEPTION WHEN check_violation THEN
    v_rejected := true;
  END;
  IF NOT v_rejected THEN
    RAISE EXCEPTION 'quitar solo la dirección de un turno a domicilio debió rechazarse.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- Aislamiento RLS: la dirección de A nunca es visible desde B (CA-103-06)
-- ---------------------------------------------------------------------------
SET LOCAL ROLE barberia_app;
SET LOCAL app.barbershop_id = '01030002-0100-4100-8100-010300000002';

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM appointment WHERE service_address = 'Calle 10 # 5-20, Bogotá') THEN
    RAISE EXCEPTION 'la barbería B no debe ver la dirección a domicilio de A.';
  END IF;
  IF (SELECT count(*) FROM appointment WHERE is_home_service) <> 1 THEN
    RAISE EXCEPTION 'la barbería B debe ver solo su propio turno a domicilio.';
  END IF;
END
$$;

RESET ROLE;

ROLLBACK;

\echo '=== DEC-131 · domicilio: OK ==='
