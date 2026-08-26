-- Pruebas SQL de HU-041: excepciones de jornada y festivos. Cubren
-- CA-041-04, CA-041-05, CA-041-06 y DEC-070 (CT-008: FK en
-- ON DELETE RESTRICT) a nivel de base de datos (esquema, RLS, grants, FK,
-- EXCLUDE, disparador); el cálculo del calendario colombiano de festivos,
-- la precedencia de resolución (CA-041-01/02/03/07), la paginación por
-- cursor y el protocolo de idempotencia completo viven en el servicio Go
-- (internal/modules/schedule), probados aparte contra PostgreSQL real.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu041_excepciones.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/hu041_excepciones.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- y `RESET ROLE` hacia un rol con privilegios administrativos (mismo
-- criterio que hu040_horario.sql).
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== HU-041 · pruebas de excepciones de jornada y festivos ==='

-- ---------------------------------------------------------------------------
-- Esquema mínimo
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barber'
      AND column_name = 'holiday_calendar_enabled' AND is_nullable = 'NO'
  ) THEN
    RAISE EXCEPTION 'barber.holiday_calendar_enabled debe existir y ser NOT NULL.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barber'
      AND column_name = 'holiday_calendar_enabled' AND column_default = 'false'
  ) THEN
    RAISE EXCEPTION 'barber.holiday_calendar_enabled debe tener DEFAULT false.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'working_hour_override'
  ) THEN
    RAISE EXCEPTION 'working_hour_override debe existir.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'working_hour_override_segment'
  ) THEN
    RAISE EXCEPTION 'working_hour_override_segment debe existir.';
  END IF;
END
$$;
\echo 'esquema OK · barber.holiday_calendar_enabled y ambas tablas de excepción existen'

-- ---------------------------------------------------------------------------
-- CA-041-05 · Una sola cabecera por barbero y fecha
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee';

DO $$
DECLARE
  v_barber uuid := 'eeee0001-0001-0001-0001-000100010001';
BEGIN
  INSERT INTO working_hour_override (barbershop_id, barber_id, effective_date, is_closed)
  VALUES ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', v_barber, '2026-12-08', true);

  BEGIN
    INSERT INTO working_hour_override (barbershop_id, barber_id, effective_date, is_closed)
    VALUES ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', v_barber, '2026-12-08', false);
    RAISE EXCEPTION 'working_hour_override_shop_barber_date_uk: se aceptó una segunda cabecera para el mismo barbero y fecha.';
  EXCEPTION
    WHEN unique_violation THEN NULL;
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-041-05 OK · working_hour_override_shop_barber_date_uk rechaza una segunda cabecera del mismo barbero+fecha'

-- ---------------------------------------------------------------------------
-- CA-041-04 · Un día cerrado no admite tramos (disparador check_open)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee';

DO $$
DECLARE
  v_barber uuid := 'eeee0001-0001-0001-0001-000100010001';
  v_override_id uuid;
BEGIN
  INSERT INTO working_hour_override (barbershop_id, barber_id, effective_date, is_closed)
  VALUES ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', v_barber, '2026-11-01', true)
  RETURNING id INTO v_override_id;

  BEGIN
    INSERT INTO working_hour_override_segment (barbershop_id, override_id, starts_time, duration_minutes)
    VALUES ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', v_override_id, '08:00', 60);
    RAISE EXCEPTION 'working_hour_override_segment_check_open_trg: se aceptó un tramo bajo una cabecera cerrada.';
  EXCEPTION
    WHEN raise_exception THEN NULL; -- Esperado: el disparador lanza RAISE EXCEPTION.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-041-04 OK · un tramo bajo una cabecera cerrada (is_closed=true) se rechaza'

-- ---------------------------------------------------------------------------
-- CA-041-04 · Un día abierto acepta varios tramos sin solape, y rechaza el
-- solape entre tramos de la MISMA cabecera (EXCLUDE, sin caso especial de
-- Go: esta invariante SÍ vive en la base porque la fecha es fija)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee';

DO $$
DECLARE
  v_barber uuid := 'eeee0001-0001-0001-0001-000100010001';
  v_override_id uuid;
  v_count integer;
BEGIN
  INSERT INTO working_hour_override (barbershop_id, barber_id, effective_date, is_closed)
  VALUES ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', v_barber, '2026-07-20', false)
  RETURNING id INTO v_override_id;

  INSERT INTO working_hour_override_segment (barbershop_id, override_id, starts_time, duration_minutes)
  VALUES ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', v_override_id, '09:00', 180);
  INSERT INTO working_hour_override_segment (barbershop_id, override_id, starts_time, duration_minutes)
  VALUES ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', v_override_id, '14:00', 120);

  SELECT count(*) INTO v_count FROM working_hour_override_segment WHERE override_id = v_override_id;
  IF v_count <> 2 THEN
    RAISE EXCEPTION 'CA-041-04: se esperaban 2 tramos sin solape, hay %.', v_count;
  END IF;

  BEGIN
    INSERT INTO working_hour_override_segment (barbershop_id, override_id, starts_time, duration_minutes)
    VALUES ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', v_override_id, '10:00', 60);
    RAISE EXCEPTION 'working_hour_override_segment_no_overlap_excl: se aceptó un tramo solapado con otro de la misma cabecera.';
  EXCEPTION
    WHEN exclusion_violation THEN NULL;
  END;

  -- Tramo nocturno: 22:00 + 180 min termina a las 01:00 del día siguiente,
  -- válido en una cabecera distinta de la misma fecha lógica (DEC-020).
  DECLARE
    v_override_id2 uuid;
  BEGIN
    INSERT INTO working_hour_override (barbershop_id, barber_id, effective_date, is_closed)
    VALUES ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', v_barber, '2026-12-31', false)
    RETURNING id INTO v_override_id2;

    INSERT INTO working_hour_override_segment (barbershop_id, override_id, starts_time, duration_minutes)
    VALUES ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', v_override_id2, '22:00', 180);
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-041-04 OK · varios tramos sin solape se aceptan, el solape entre tramos de la misma cabecera se rechaza (EXCLUDE), y un tramo nocturno es válido'

-- ---------------------------------------------------------------------------
-- CHECK de duración [1,1440] y de reason (<=200 caracteres)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee';

DO $$
DECLARE
  v_barber uuid := 'eeee0002-0002-0002-0002-000200020002';
  v_override_id uuid;
BEGIN
  BEGIN
    INSERT INTO working_hour_override (barbershop_id, barber_id, effective_date, is_closed, reason)
    VALUES ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', v_barber, '2026-05-01', true, repeat('a', 201));
    RAISE EXCEPTION 'working_hour_override_reason_ck: se aceptó un motivo de 201 caracteres.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  INSERT INTO working_hour_override (barbershop_id, barber_id, effective_date, is_closed)
  VALUES ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', v_barber, '2026-05-01', false)
  RETURNING id INTO v_override_id;

  BEGIN
    INSERT INTO working_hour_override_segment (barbershop_id, override_id, starts_time, duration_minutes)
    VALUES ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', v_override_id, '08:00', 0);
    RAISE EXCEPTION 'working_hour_override_segment_duration_minutes_ck: se aceptó duration_minutes = 0.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO working_hour_override_segment (barbershop_id, override_id, starts_time, duration_minutes)
    VALUES ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', v_override_id, '08:00', 1441);
    RAISE EXCEPTION 'working_hour_override_segment_duration_minutes_ck: se aceptó duration_minutes = 1441.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'constraint OK · working_hour_override_reason_ck y working_hour_override_segment_duration_minutes_ck rechazan fuera de rango'

-- ---------------------------------------------------------------------------
-- RLS forzada, sin BYPASSRLS (ambas tablas)
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_relrowsecurity boolean;
  v_relforcerowsecurity boolean;
BEGIN
  FOR v_relrowsecurity, v_relforcerowsecurity IN
    SELECT relrowsecurity, relforcerowsecurity FROM pg_class
    WHERE relname IN ('working_hour_override', 'working_hour_override_segment')
      AND relnamespace = 'public'::regnamespace
  LOOP
    IF NOT v_relrowsecurity OR NOT v_relforcerowsecurity THEN
      RAISE EXCEPTION 'working_hour_override/working_hour_override_segment deben tener RLS habilitada Y forzada.';
    END IF;
  END LOOP;

  IF EXISTS (
    SELECT 1 FROM pg_roles WHERE rolname = 'barberia_app' AND rolbypassrls
  ) THEN
    RAISE EXCEPTION 'barberia_app no puede tener BYPASSRLS.';
  END IF;
END
$$;
\echo 'RLS OK · ambas tablas tienen RLS habilitada y forzada, barberia_app sin BYPASSRLS'

-- ---------------------------------------------------------------------------
-- Grants de barberia_app: SELECT/INSERT/UPDATE/DELETE en ambas tablas
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_missing text;
  v_table text;
BEGIN
  FOREACH v_table IN ARRAY ARRAY['working_hour_override', 'working_hour_override_segment']
  LOOP
    SELECT string_agg(expected, ', ') INTO v_missing
    FROM unnest(ARRAY['SELECT', 'INSERT', 'UPDATE', 'DELETE']) AS expected
    WHERE NOT EXISTS (
      SELECT 1 FROM information_schema.table_privileges
      WHERE table_schema = 'public' AND table_name = v_table
        AND grantee = 'barberia_app' AND privilege_type = expected
    );
    IF v_missing IS NOT NULL THEN
      RAISE EXCEPTION 'barberia_app debe tener % sobre %.', v_missing, v_table;
    END IF;
  END LOOP;
END
$$;
\echo 'grants OK · barberia_app tiene SELECT/INSERT/UPDATE/DELETE sobre ambas tablas'

-- ---------------------------------------------------------------------------
-- CA-041-06 · El contexto de M no ve, no edita ni retira una excepción (ni
-- sus tramos) de N
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'ffffffff-ffff-ffff-ffff-ffffffffffff';

INSERT INTO working_hour_override (id, barbershop_id, barber_id, effective_date, is_closed)
VALUES (
  'dead0001-0001-0001-0001-000100010001', 'ffffffff-ffff-ffff-ffff-ffffffffffff',
  'ffff0001-0001-0001-0001-000100010001', '2026-08-07', false
);
INSERT INTO working_hour_override_segment (id, barbershop_id, override_id, starts_time, duration_minutes)
VALUES (
  'beef0001-0001-0001-0001-000100010001', 'ffffffff-ffff-ffff-ffff-ffffffffffff',
  'dead0001-0001-0001-0001-000100010001', '08:00', 60
);

RESET ROLE;
COMMIT;

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee';

DO $$
DECLARE
  v_override_n_id uuid := 'dead0001-0001-0001-0001-000100010001';
  v_segment_n_id  uuid := 'beef0001-0001-0001-0001-000100010001';
  v_visible integer;
  v_updated integer;
  v_deleted integer;
BEGIN
  SELECT count(*) INTO v_visible FROM working_hour_override WHERE id = v_override_n_id;
  IF v_visible <> 0 THEN
    RAISE EXCEPTION 'CA-041-06: el contexto de M puede ver la excepción de N por su id real.';
  END IF;

  SELECT count(*) INTO v_visible FROM working_hour_override_segment WHERE id = v_segment_n_id;
  IF v_visible <> 0 THEN
    RAISE EXCEPTION 'CA-041-06: el contexto de M puede ver el tramo de N por su id real.';
  END IF;

  UPDATE working_hour_override SET reason = 'intento cruzado' WHERE id = v_override_n_id;
  GET DIAGNOSTICS v_updated = ROW_COUNT;
  IF v_updated <> 0 THEN
    RAISE EXCEPTION 'CA-041-06: el contexto de M pudo editar la excepción de N (% filas).', v_updated;
  END IF;

  DELETE FROM working_hour_override_segment WHERE id = v_segment_n_id;
  GET DIAGNOSTICS v_deleted = ROW_COUNT;
  IF v_deleted <> 0 THEN
    RAISE EXCEPTION 'CA-041-06: el contexto de M pudo retirar el tramo de N (% filas).', v_deleted;
  END IF;

  DELETE FROM working_hour_override WHERE id = v_override_n_id;
  GET DIAGNOSTICS v_deleted = ROW_COUNT;
  IF v_deleted <> 0 THEN
    RAISE EXCEPTION 'CA-041-06: el contexto de M pudo retirar la excepción de N (% filas).', v_deleted;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-041-06 OK · el contexto de M no ve, no edita y no retira la excepción ni el tramo de N'

-- Limpieza de las filas de N confirmadas arriba (segmento antes que
-- cabecera: ON DELETE RESTRICT, no se puede borrar la cabecera primero).
DELETE FROM working_hour_override_segment WHERE id = 'beef0001-0001-0001-0001-000100010001';
DELETE FROM working_hour_override WHERE id = 'dead0001-0001-0001-0001-000100010001';

-- ---------------------------------------------------------------------------
-- DEC-070/CT-008 · Ambas FK usan ON DELETE RESTRICT, no CASCADE
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee';

INSERT INTO working_hour_override (id, barbershop_id, barber_id, effective_date, is_closed)
VALUES (
  'dead0002-0002-0002-0002-000200020002', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
  'eeee0002-0002-0002-0002-000200020002', '2026-06-29', false
);
INSERT INTO working_hour_override_segment (id, barbershop_id, override_id, starts_time, duration_minutes)
VALUES (
  'beef0002-0002-0002-0002-000200020002', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
  'dead0002-0002-0002-0002-000200020002', '08:00', 60
);

RESET ROLE;
COMMIT;

DO $$
BEGIN
  -- La cabecera no se puede borrar mientras tenga un tramo (RESTRICT de
  -- working_hour_override_segment hacia working_hour_override).
  BEGIN
    DELETE FROM working_hour_override WHERE id = 'dead0002-0002-0002-0002-000200020002';
    RAISE EXCEPTION 'DEC-070/CT-008: se permitió borrar una cabecera con un tramo asociado (CASCADE en vez de RESTRICT).';
  EXCEPTION
    WHEN foreign_key_violation THEN NULL; -- Esperado: working_hour_override_segment_..._fk en RESTRICT.
  END;

  -- El barbero no se puede borrar mientras tenga una excepción (RESTRICT
  -- de working_hour_override hacia barber).
  BEGIN
    DELETE FROM barber WHERE id = 'eeee0002-0002-0002-0002-000200020002';
    RAISE EXCEPTION 'DEC-070/CT-008: se permitió borrar un barber con una excepción asociada (CASCADE en vez de RESTRICT).';
  EXCEPTION
    WHEN foreign_key_violation THEN NULL; -- Esperado: working_hour_override_..._fk en RESTRICT.
  END;
END
$$;

-- Limpieza de las filas confirmadas arriba (segmento antes que cabecera).
DELETE FROM working_hour_override_segment WHERE id = 'beef0002-0002-0002-0002-000200020002';
DELETE FROM working_hour_override WHERE id = 'dead0002-0002-0002-0002-000200020002';
\echo 'DEC-070/CT-008 OK · ninguna de las dos FK permite un borrado en cascada; ambas rechazan (ON DELETE RESTRICT)'

-- ---------------------------------------------------------------------------
-- Trigger de updated_at sobre working_hour_override
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee';

INSERT INTO working_hour_override (id, barbershop_id, barber_id, effective_date, is_closed)
VALUES (
  'dead0003-0003-0003-0003-000300030003', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
  'eeee0001-0001-0001-0001-000100010001', '2026-01-06', true
);

RESET ROLE;
COMMIT;

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee';

DO $$
DECLARE
  v_updated_before timestamptz;
  v_updated_after timestamptz;
BEGIN
  SELECT updated_at INTO v_updated_before FROM working_hour_override WHERE id = 'dead0003-0003-0003-0003-000300030003';

  PERFORM pg_sleep(0.01);

  UPDATE working_hour_override SET reason = 'Actualizado' WHERE id = 'dead0003-0003-0003-0003-000300030003'
  RETURNING updated_at INTO v_updated_after;

  IF v_updated_after <= v_updated_before THEN
    RAISE EXCEPTION 'working_hour_override_set_updated_at: updated_at no avanzó tras el UPDATE (antes=%, después=%).',
      v_updated_before, v_updated_after;
  END IF;
END
$$;

RESET ROLE;
COMMIT;

-- Limpieza de la fila confirmada para el trigger.
DELETE FROM working_hour_override WHERE id = 'dead0003-0003-0003-0003-000300030003';

\echo 'trigger OK · working_hour_override_set_updated_at avanza updated_at en cada UPDATE'

\echo '=== HU-041 · todas las comprobaciones pasaron ==='
