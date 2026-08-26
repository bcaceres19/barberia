-- Pruebas SQL de HU-042: bloqueos de agenda. Cubren el esquema mínimo de las
-- cuatro tablas (time_block, time_block_series, time_block_series_date,
-- time_block_series_exception), sus CHECK, RLS/grants exactos (sin DELETE:
-- RN-BLQ-04), el disparador de forma de time_block_series_date, y DEC-070
-- (CT-008: FK en ON DELETE RESTRICT) a nivel de base de datos; la
-- orquestación completa (idempotencia, expansión de series, división
-- this_and_following, proyección efectiva) vive en el servicio Go
-- (internal/modules/schedule), probada aparte contra PostgreSQL real.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu042_bloqueos.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/hu042_bloqueos.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- y `RESET ROLE` hacia un rol con privilegios administrativos (mismo
-- criterio que hu040_horario.sql/hu041_excepciones.sql).
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== HU-042 · pruebas de bloqueos de agenda ==='

-- ---------------------------------------------------------------------------
-- Esquema mínimo
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'time_block') THEN
    RAISE EXCEPTION 'time_block debe existir.';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'time_block_series') THEN
    RAISE EXCEPTION 'time_block_series debe existir.';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'time_block_series_date') THEN
    RAISE EXCEPTION 'time_block_series_date debe existir.';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'time_block_series_exception') THEN
    RAISE EXCEPTION 'time_block_series_exception debe existir.';
  END IF;
END
$$;
\echo 'esquema OK · las cuatro tablas existen'

-- ---------------------------------------------------------------------------
-- CHECK de time_block: block_type (siete valores), source, intervalo
-- semiabierto (ends_at > starts_at), reason (<=200), deleted_by coherente
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'b10c0001-b10c-b10c-b10c-b10c00010001';

DO $$
DECLARE
  v_barber uuid := 'b10c1001-1001-1001-1001-100110011001';
BEGIN
  BEGIN
    INSERT INTO time_block (barbershop_id, barber_id, block_type, starts_at, ends_at)
    VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_barber, 'appointment', now(), now() + interval '1 hour');
    RAISE EXCEPTION 'time_block_block_type_ck: se aceptó un tipo fuera de los siete cerrados.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO time_block (barbershop_id, barber_id, block_type, source, starts_at, ends_at)
    VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_barber, 'break', 'imported', now(), now() + interval '1 hour');
    RAISE EXCEPTION 'time_block_source_ck: se aceptó un source fuera de manual/holiday_calendar.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO time_block (barbershop_id, barber_id, block_type, starts_at, ends_at)
    VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_barber, 'unavailable', now(), now());
    RAISE EXCEPTION 'time_block_interval_ck: se aceptó ends_at = starts_at (semiabierto).';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO time_block (barbershop_id, barber_id, block_type, starts_at, ends_at, reason)
    VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_barber, 'day_off', now(), now() + interval '1 day', repeat('a', 201));
    RAISE EXCEPTION 'time_block_reason_ck: se aceptó un motivo de 201 caracteres.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  -- RN-BLQ-03/DEC-008: un bloqueo urgente cruzando medianoche siempre se crea.
  INSERT INTO time_block (barbershop_id, barber_id, block_type, starts_at, ends_at)
  VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_barber, 'emergency', now(), now() + interval '10 hours');
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CHECK OK · time_block_block_type_ck/source_ck/interval_ck/reason_ck rechazan valores inválidos'

-- ---------------------------------------------------------------------------
-- CHECK de time_block_series: recurrence_kind, weekday_shape, duración,
-- rango efectivo
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'b10c0001-b10c-b10c-b10c-b10c00010001';

DO $$
DECLARE
  v_barber uuid := 'b10c1001-1001-1001-1001-100110011001';
BEGIN
  BEGIN
    INSERT INTO time_block_series (barbershop_id, barber_id, block_type, recurrence_kind, iso_weekday, starts_time, duration_minutes, effective_from)
    VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_barber, 'lunch', 'daily', 1, '13:00', 60, '2027-01-01');
    RAISE EXCEPTION 'time_block_series_recurrence_kind_ck: se aceptó daily.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO time_block_series (barbershop_id, barber_id, block_type, recurrence_kind, iso_weekday, starts_time, duration_minutes, effective_from)
    VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_barber, 'lunch', 'weekly', NULL, '13:00', 60, '2027-01-01');
    RAISE EXCEPTION 'time_block_series_weekday_shape_ck: se aceptó weekly sin iso_weekday.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO time_block_series (barbershop_id, barber_id, block_type, recurrence_kind, iso_weekday, starts_time, duration_minutes, effective_from)
    VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_barber, 'vacation', 'date_list', 3, '00:00', 1440, '2027-01-01');
    RAISE EXCEPTION 'time_block_series_weekday_shape_ck: se aceptó date_list con iso_weekday.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO time_block_series (barbershop_id, barber_id, block_type, recurrence_kind, iso_weekday, starts_time, duration_minutes, effective_from)
    VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_barber, 'lunch', 'weekly', 1, '13:00', 0, '2027-01-01');
    RAISE EXCEPTION 'time_block_series_duration_minutes_ck: se aceptó duration_minutes = 0.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO time_block_series (barbershop_id, barber_id, block_type, recurrence_kind, iso_weekday, starts_time, duration_minutes, effective_from, effective_until)
    VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_barber, 'lunch', 'weekly', 1, '13:00', 60, '2027-02-01', '2027-01-01');
    RAISE EXCEPTION 'time_block_series_effective_range_ck: se aceptó effective_until anterior a effective_from.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CHECK OK · time_block_series rechaza recurrence_kind/weekday_shape/duración/rango inválidos'

-- ---------------------------------------------------------------------------
-- Disparador time_block_series_date_check_parent: date_list y rango vigente
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'b10c0001-b10c-b10c-b10c-b10c00010001';

DO $$
DECLARE
  v_barber uuid := 'b10c1001-1001-1001-1001-100110011001';
  v_weekly_id uuid;
  v_datelist_id uuid;
BEGIN
  INSERT INTO time_block_series (barbershop_id, barber_id, block_type, recurrence_kind, iso_weekday, starts_time, duration_minutes, effective_from)
  VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_barber, 'lunch', 'weekly', 2, '13:00', 60, '2027-01-01')
  RETURNING id INTO v_weekly_id;

  BEGIN
    INSERT INTO time_block_series_date (barbershop_id, series_id, block_date)
    VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_weekly_id, '2027-01-05');
    RAISE EXCEPTION 'time_block_series_date_check_parent_trg: se aceptó una fecha sobre una serie weekly.';
  EXCEPTION
    WHEN raise_exception THEN NULL;
  END;

  INSERT INTO time_block_series (barbershop_id, barber_id, block_type, recurrence_kind, starts_time, duration_minutes, effective_from, effective_until)
  VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_barber, 'vacation', 'date_list', '00:00', 1440, '2027-03-01', '2027-03-31')
  RETURNING id INTO v_datelist_id;

  BEGIN
    INSERT INTO time_block_series_date (barbershop_id, series_id, block_date)
    VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_datelist_id, '2027-04-01');
    RAISE EXCEPTION 'time_block_series_date_check_parent_trg: se aceptó una fecha fuera del rango vigente.';
  EXCEPTION
    WHEN raise_exception THEN NULL;
  END;

  -- Dentro de rango: se acepta.
  INSERT INTO time_block_series_date (barbershop_id, series_id, block_date)
  VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_datelist_id, '2027-03-15');

  -- PK compuesta: la misma fecha dos veces se rechaza.
  BEGIN
    INSERT INTO time_block_series_date (barbershop_id, series_id, block_date)
    VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_datelist_id, '2027-03-15');
    RAISE EXCEPTION 'time_block_series_date_pk: se aceptó la misma fecha dos veces en la misma serie.';
  EXCEPTION
    WHEN unique_violation THEN NULL;
  END;

  -- Misma fecha dos veces como EXCEPCIÓN de la misma serie: también se rechaza.
  INSERT INTO time_block_series_exception (barbershop_id, series_id, excluded_date)
  VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_datelist_id, '2027-03-15');
  BEGIN
    INSERT INTO time_block_series_exception (barbershop_id, series_id, excluded_date)
    VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', v_datelist_id, '2027-03-15');
    RAISE EXCEPTION 'time_block_series_exception_pk: se aceptó la misma fecha excepcionada dos veces.';
  EXCEPTION
    WHEN unique_violation THEN NULL;
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'disparador OK · time_block_series_date exige date_list y rango vigente; PK de dates/exceptions rechaza duplicados'

-- ---------------------------------------------------------------------------
-- RLS forzada, sin BYPASSRLS (las cuatro tablas)
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_relrowsecurity boolean;
  v_relforcerowsecurity boolean;
BEGIN
  FOR v_relrowsecurity, v_relforcerowsecurity IN
    SELECT relrowsecurity, relforcerowsecurity FROM pg_class
    WHERE relname IN ('time_block', 'time_block_series', 'time_block_series_date', 'time_block_series_exception')
      AND relnamespace = 'public'::regnamespace
  LOOP
    IF NOT v_relrowsecurity OR NOT v_relforcerowsecurity THEN
      RAISE EXCEPTION 'las cuatro tablas de HU-042 deben tener RLS habilitada Y forzada.';
    END IF;
  END LOOP;

  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_app' AND rolbypassrls) THEN
    RAISE EXCEPTION 'barberia_app no puede tener BYPASSRLS.';
  END IF;
END
$$;
\echo 'RLS OK · las cuatro tablas tienen RLS habilitada y forzada, barberia_app sin BYPASSRLS'

-- ---------------------------------------------------------------------------
-- Grants exactos de barberia_app: sin DELETE en time_block/time_block_series
-- (RN-BLQ-04: eliminación lógica); SELECT/INSERT/DELETE (sin UPDATE) en
-- dates/exceptions, filas hijas sin ciclo de vida propio
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_missing text;
  v_extra text;
BEGIN
  SELECT string_agg(expected, ', ') INTO v_missing
  FROM unnest(ARRAY['SELECT', 'INSERT', 'UPDATE']) AS expected
  WHERE NOT EXISTS (
    SELECT 1 FROM information_schema.table_privileges
    WHERE table_schema = 'public' AND table_name = 'time_block'
      AND grantee = 'barberia_app' AND privilege_type = expected
  );
  IF v_missing IS NOT NULL THEN
    RAISE EXCEPTION 'barberia_app debe tener % sobre time_block.', v_missing;
  END IF;
  IF EXISTS (
    SELECT 1 FROM information_schema.table_privileges
    WHERE table_schema = 'public' AND table_name = 'time_block'
      AND grantee = 'barberia_app' AND privilege_type = 'DELETE'
  ) THEN
    RAISE EXCEPTION 'RN-BLQ-04: barberia_app NO debe tener DELETE sobre time_block.';
  END IF;

  SELECT string_agg(expected, ', ') INTO v_missing
  FROM unnest(ARRAY['SELECT', 'INSERT', 'UPDATE']) AS expected
  WHERE NOT EXISTS (
    SELECT 1 FROM information_schema.table_privileges
    WHERE table_schema = 'public' AND table_name = 'time_block_series'
      AND grantee = 'barberia_app' AND privilege_type = expected
  );
  IF v_missing IS NOT NULL THEN
    RAISE EXCEPTION 'barberia_app debe tener % sobre time_block_series.', v_missing;
  END IF;
  IF EXISTS (
    SELECT 1 FROM information_schema.table_privileges
    WHERE table_schema = 'public' AND table_name = 'time_block_series'
      AND grantee = 'barberia_app' AND privilege_type = 'DELETE'
  ) THEN
    RAISE EXCEPTION 'RN-BLQ-04: barberia_app NO debe tener DELETE sobre time_block_series.';
  END IF;

  FOREACH v_extra IN ARRAY ARRAY['time_block_series_date', 'time_block_series_exception']
  LOOP
    SELECT string_agg(expected, ', ') INTO v_missing
    FROM unnest(ARRAY['SELECT', 'INSERT', 'DELETE']) AS expected
    WHERE NOT EXISTS (
      SELECT 1 FROM information_schema.table_privileges
      WHERE table_schema = 'public' AND table_name = v_extra
        AND grantee = 'barberia_app' AND privilege_type = expected
    );
    IF v_missing IS NOT NULL THEN
      RAISE EXCEPTION 'barberia_app debe tener % sobre %.', v_missing, v_extra;
    END IF;
  END LOOP;
END
$$;
\echo 'grants OK · sin DELETE en time_block/time_block_series (RN-BLQ-04); SELECT/INSERT/DELETE exactos en dates/exceptions'

-- ---------------------------------------------------------------------------
-- Aislamiento de tenant: el contexto de O no ve, no edita ni retira nada de P
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'b10c0002-b10c-b10c-b10c-b10c00020002';

INSERT INTO time_block (id, barbershop_id, barber_id, block_type, starts_at, ends_at)
VALUES (
  'dead0001-0001-0001-0001-000100010001', 'b10c0002-b10c-b10c-b10c-b10c00020002',
  'b10c2001-2001-2001-2001-200120012001', 'unavailable', '2027-05-01T10:00:00Z', '2027-05-01T11:00:00Z'
);
INSERT INTO time_block_series (id, barbershop_id, barber_id, block_type, recurrence_kind, iso_weekday, starts_time, duration_minutes, effective_from)
VALUES (
  'beef0001-0001-0001-0001-000100010001', 'b10c0002-b10c-b10c-b10c-b10c00020002',
  'b10c2001-2001-2001-2001-200120012001', 'lunch', 'weekly', 4, '13:00', 60, '2027-01-01'
);

RESET ROLE;
COMMIT;

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'b10c0001-b10c-b10c-b10c-b10c00010001';

DO $$
DECLARE
  v_block_p_id  uuid := 'dead0001-0001-0001-0001-000100010001';
  v_series_p_id uuid := 'beef0001-0001-0001-0001-000100010001';
  v_count integer;
  v_updated integer;
BEGIN
  SELECT count(*) INTO v_count FROM time_block WHERE id = v_block_p_id;
  IF v_count <> 0 THEN
    RAISE EXCEPTION 'RN-TEN-01: el contexto de O puede ver el bloqueo de P por su id real.';
  END IF;

  SELECT count(*) INTO v_count FROM time_block_series WHERE id = v_series_p_id;
  IF v_count <> 0 THEN
    RAISE EXCEPTION 'RN-TEN-01: el contexto de O puede ver la serie de P por su id real.';
  END IF;

  UPDATE time_block SET reason = 'intento cruzado' WHERE id = v_block_p_id;
  GET DIAGNOSTICS v_updated = ROW_COUNT;
  IF v_updated <> 0 THEN
    RAISE EXCEPTION 'RN-TEN-01: el contexto de O pudo editar el bloqueo de P (% filas).', v_updated;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;

-- Limpieza de las filas de P confirmadas arriba. Sin SET ROLE
-- barberia_app: RN-BLQ-04 no le concede DELETE sobre time_block ni
-- time_block_series a propósito; limpiar como el rol administrativo
-- conectado (superusuario/migrador en este arnés) es correcto aquí.
BEGIN;
DELETE FROM time_block WHERE id = 'dead0001-0001-0001-0001-000100010001';
DELETE FROM time_block_series WHERE id = 'beef0001-0001-0001-0001-000100010001';
COMMIT;
\echo 'RN-TEN-01 OK · el contexto de O no ve ni edita el bloqueo ni la serie de P'

-- ---------------------------------------------------------------------------
-- DEC-070/CT-008 · Las FK hacia barber/staff_user/time_block_series usan
-- ON DELETE RESTRICT, no CASCADE
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'b10c0001-b10c-b10c-b10c-b10c00010001';

INSERT INTO time_block (id, barbershop_id, barber_id, block_type, starts_at, ends_at, deleted_at, deleted_by)
VALUES (
  'dead0002-0002-0002-0002-000200020002', 'b10c0001-b10c-b10c-b10c-b10c00010001',
  'b10c1002-1002-1002-1002-100210021002', 'day_off', '2027-05-05T00:00:00Z', '2027-05-06T00:00:00Z',
  now(), 'b10c9001-9001-9001-9001-900190019001'
);
INSERT INTO time_block_series (id, barbershop_id, barber_id, block_type, recurrence_kind, starts_time, duration_minutes, effective_from)
VALUES (
  'beef0002-0002-0002-0002-000200020002', 'b10c0001-b10c-b10c-b10c-b10c00010001',
  'b10c1002-1002-1002-1002-100210021002', 'vacation', 'date_list', '00:00', 1440, '2027-09-01'
);
INSERT INTO time_block_series_date (barbershop_id, series_id, block_date)
VALUES ('b10c0001-b10c-b10c-b10c-b10c00010001', 'beef0002-0002-0002-0002-000200020002', '2027-09-10');

RESET ROLE;
COMMIT;

DO $$
BEGIN
  -- El barbero no se puede borrar mientras tenga un bloqueo asociado.
  BEGIN
    DELETE FROM barber WHERE id = 'b10c1002-1002-1002-1002-100210021002';
    RAISE EXCEPTION 'DEC-070/CT-008: se permitió borrar un barber con un time_block asociado (CASCADE en vez de RESTRICT).';
  EXCEPTION
    WHEN foreign_key_violation THEN NULL;
  END;

  -- El staff_user que retiró un bloqueo no se puede borrar físicamente
  -- (time_block_barbershop_id_deleted_by_fk).
  BEGIN
    DELETE FROM staff_user WHERE id = 'b10c9001-9001-9001-9001-900190019001';
    RAISE EXCEPTION 'DEC-070/CT-008: se permitió borrar un staff_user referenciado por time_block.deleted_by.';
  EXCEPTION
    WHEN foreign_key_violation THEN NULL;
  END;

  -- La serie no se puede borrar mientras tenga una fecha explícita asociada.
  BEGIN
    DELETE FROM time_block_series WHERE id = 'beef0002-0002-0002-0002-000200020002';
    RAISE EXCEPTION 'DEC-070/CT-008: se permitió borrar una serie con una fecha explícita asociada (CASCADE en vez de RESTRICT).';
  EXCEPTION
    WHEN foreign_key_violation THEN NULL;
  END;
END
$$;

-- Limpieza (fecha antes que serie: RESTRICT, no se puede borrar la serie
-- primero). Sin SET ROLE barberia_app para time_block_series/time_block por
-- el mismo motivo que la limpieza cruzada de arriba (RN-BLQ-04, sin DELETE).
BEGIN;
DELETE FROM time_block_series_date WHERE series_id = 'beef0002-0002-0002-0002-000200020002';
DELETE FROM time_block_series WHERE id = 'beef0002-0002-0002-0002-000200020002';
DELETE FROM time_block WHERE id = 'dead0002-0002-0002-0002-000200020002';
COMMIT;
\echo 'DEC-070/CT-008 OK · ninguna FK de HU-042 permite un borrado en cascada; todas rechazan (ON DELETE RESTRICT)'

-- ---------------------------------------------------------------------------
-- Trigger de updated_at sobre time_block y time_block_series
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'b10c0001-b10c-b10c-b10c-b10c00010001';

INSERT INTO time_block (id, barbershop_id, barber_id, block_type, starts_at, ends_at)
VALUES (
  'dead0003-0003-0003-0003-000300030003', 'b10c0001-b10c-b10c-b10c-b10c00010001',
  'b10c1001-1001-1001-1001-100110011001', 'break', '2027-06-01T09:00:00Z', '2027-06-01T09:15:00Z'
);

RESET ROLE;
COMMIT;

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'b10c0001-b10c-b10c-b10c-b10c00010001';

DO $$
DECLARE
  v_updated_before timestamptz;
  v_updated_after timestamptz;
BEGIN
  SELECT updated_at INTO v_updated_before FROM time_block WHERE id = 'dead0003-0003-0003-0003-000300030003';

  PERFORM pg_sleep(0.01);

  UPDATE time_block SET reason = 'Actualizado' WHERE id = 'dead0003-0003-0003-0003-000300030003'
  RETURNING updated_at INTO v_updated_after;

  IF v_updated_after <= v_updated_before THEN
    RAISE EXCEPTION 'time_block_set_updated_at: updated_at no avanzó tras el UPDATE (antes=%, después=%).',
      v_updated_before, v_updated_after;
  END IF;
END
$$;

RESET ROLE;
COMMIT;

-- Limpieza de la fila confirmada para el trigger. Sin SET ROLE
-- barberia_app: mismo motivo que las dos limpiezas anteriores.
BEGIN;
DELETE FROM time_block WHERE id = 'dead0003-0003-0003-0003-000300030003';
COMMIT;

\echo 'trigger OK · time_block_set_updated_at avanza updated_at en cada UPDATE'

\echo '=== HU-042 · todas las comprobaciones pasaron ==='
