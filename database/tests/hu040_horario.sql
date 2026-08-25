-- Pruebas SQL de HU-040: horario laboral recurrente. Cubren CA-040-01,
-- CA-040-03, CA-040-04, CA-040-05 y CA-070 (DEC-070/CT-008: FK en
-- ON DELETE RESTRICT) a nivel de base de datos (esquema, RLS, grants, FK);
-- la validación de solape entre tramos, la paginación por cursor y el
-- protocolo de idempotencia completo viven en el servicio Go
-- (internal/modules/schedule), probados aparte contra PostgreSQL real.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu040_horario.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/hu040_horario.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- y `RESET ROLE` hacia un rol con privilegios administrativos (mismo
-- criterio que hu021_barberos.sql).
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== HU-040 · pruebas de horario laboral recurrente ==='

-- ---------------------------------------------------------------------------
-- Esquema mínimo: columnas, tipos y CHECK esperados
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'working_hour'
  ) THEN
    RAISE EXCEPTION 'working_hour debe existir.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'working_hour'
      AND column_name = 'iso_weekday' AND is_nullable = 'NO'
  ) THEN
    RAISE EXCEPTION 'working_hour.iso_weekday debe existir y ser NOT NULL.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'working_hour'
      AND column_name = 'starts_time' AND data_type = 'time without time zone'
  ) THEN
    RAISE EXCEPTION 'working_hour.starts_time debe existir y ser time sin zona (hora civil).';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'working_hour'
      AND column_name = 'duration_minutes' AND is_nullable = 'NO'
  ) THEN
    RAISE EXCEPTION 'working_hour.duration_minutes debe existir y ser NOT NULL.';
  END IF;
END
$$;
\echo 'esquema OK · working_hour tiene iso_weekday, starts_time (time) y duration_minutes NOT NULL'

-- ---------------------------------------------------------------------------
-- CA-040-04 (nivel base de datos) · CHECK de día [1,7] y duración [1,1440]
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'cccccccc-cccc-cccc-cccc-cccccccccccc';

DO $$
DECLARE
  v_barber uuid := 'cccc0001-0001-0001-0001-000100010001';
BEGIN
  BEGIN
    INSERT INTO working_hour (barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
    VALUES ('cccccccc-cccc-cccc-cccc-cccccccccccc', v_barber, 0, '08:00', 60);
    RAISE EXCEPTION 'working_hour_iso_weekday_ck: se aceptó iso_weekday = 0.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO working_hour (barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
    VALUES ('cccccccc-cccc-cccc-cccc-cccccccccccc', v_barber, 8, '08:00', 60);
    RAISE EXCEPTION 'working_hour_iso_weekday_ck: se aceptó iso_weekday = 8.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO working_hour (barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
    VALUES ('cccccccc-cccc-cccc-cccc-cccccccccccc', v_barber, 1, '08:00', 0);
    RAISE EXCEPTION 'working_hour_duration_minutes_ck: se aceptó duration_minutes = 0.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO working_hour (barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
    VALUES ('cccccccc-cccc-cccc-cccc-cccccccccccc', v_barber, 1, '08:00', 1441);
    RAISE EXCEPTION 'working_hour_duration_minutes_ck: se aceptó duration_minutes = 1441.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  -- Límites inclusivos válidos: día 1 y 7, duración 1 y 1440.
  INSERT INTO working_hour (barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
  VALUES ('cccccccc-cccc-cccc-cccc-cccccccccccc', v_barber, 1, '00:00', 1);
  INSERT INTO working_hour (barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
  VALUES ('cccccccc-cccc-cccc-cccc-cccccccccccc', v_barber, 7, '22:00', 1440);
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'constraint OK · working_hour_iso_weekday_ck/duration_minutes_ck rechazan fuera de rango y aceptan los límites'

-- ---------------------------------------------------------------------------
-- CA-040-02/03 · Jornada partida (varios tramos no solapados el mismo día) y
-- tramo nocturno que cruza medianoche (DEC-020) persisten sin caso especial
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'cccccccc-cccc-cccc-cccc-cccccccccccc';

DO $$
DECLARE
  v_barber uuid := 'cccc0001-0001-0001-0001-000100010001';
  v_count integer;
BEGIN
  -- Jornada partida: mañana y tarde, mismo barbero y día, sin solape.
  INSERT INTO working_hour (barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
  VALUES ('cccccccc-cccc-cccc-cccc-cccccccccccc', v_barber, 1, '08:00', 240);
  INSERT INTO working_hour (barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
  VALUES ('cccccccc-cccc-cccc-cccc-cccccccccccc', v_barber, 1, '14:00', 240);

  SELECT count(*) INTO v_count FROM working_hour
  WHERE barbershop_id = 'cccccccc-cccc-cccc-cccc-cccccccccccc' AND barber_id = v_barber AND iso_weekday = 1;
  IF v_count <> 2 THEN
    RAISE EXCEPTION 'CA-040-02: se esperaban 2 tramos de jornada partida, hay %.', v_count;
  END IF;

  -- Tramo nocturno: 22:00 + 300 min termina a las 03:00 del día siguiente,
  -- sin ambigüedad porque se almacena inicio + duración, no ends_time.
  INSERT INTO working_hour (barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
  VALUES ('cccccccc-cccc-cccc-cccc-cccccccccccc', v_barber, 5, '22:00', 300);
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-040-02/03 OK · jornada partida y tramo nocturno cruzando medianoche persisten sin rama especial'

-- ---------------------------------------------------------------------------
-- CA-040-04 · Un mismo barbero no repite exactamente el mismo día e inicio
-- (working_hour_shop_barber_weekday_start_uk). El solape PARCIAL entre dos
-- tramos que no comparten starts_time se valida en el servicio Go, no aquí
-- (ver comentario de la migración): una restricción de exclusión sobre una
-- envolvente semanal con jornadas nocturnas sería frágil.
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'cccccccc-cccc-cccc-cccc-cccccccccccc';

DO $$
DECLARE
  v_barber uuid := 'cccc0001-0001-0001-0001-000100010001';
BEGIN
  INSERT INTO working_hour (barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
  VALUES ('cccccccc-cccc-cccc-cccc-cccccccccccc', v_barber, 2, '08:00', 240);

  BEGIN
    INSERT INTO working_hour (barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
    VALUES ('cccccccc-cccc-cccc-cccc-cccccccccccc', v_barber, 2, '08:00', 60);
    RAISE EXCEPTION 'working_hour_shop_barber_weekday_start_uk: se aceptó repetir día e inicio exactos.';
  EXCEPTION
    WHEN unique_violation THEN NULL;
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-040-04 OK · working_hour_shop_barber_weekday_start_uk rechaza el mismo barbero+día+inicio exacto'

-- ---------------------------------------------------------------------------
-- CA-040-06 (nivel base de datos) · RLS forzada, sin BYPASSRLS
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_relrowsecurity boolean;
  v_relforcerowsecurity boolean;
BEGIN
  SELECT relrowsecurity, relforcerowsecurity INTO v_relrowsecurity, v_relforcerowsecurity
  FROM pg_class
  WHERE relname = 'working_hour' AND relnamespace = 'public'::regnamespace;

  IF NOT v_relrowsecurity OR NOT v_relforcerowsecurity THEN
    RAISE EXCEPTION 'working_hour debe tener RLS habilitada Y forzada.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM pg_roles WHERE rolname = 'barberia_app' AND rolbypassrls
  ) THEN
    RAISE EXCEPTION 'barberia_app no puede tener BYPASSRLS.';
  END IF;
END
$$;
\echo 'RLS OK · working_hour tiene RLS habilitada y forzada, barberia_app sin BYPASSRLS'

-- ---------------------------------------------------------------------------
-- Grants de barberia_app: SELECT/INSERT/UPDATE/DELETE, a diferencia de
-- barber (sin ciclo de vida, DEC-047): un tramo sí se edita y se retira.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_missing text;
BEGIN
  SELECT string_agg(expected, ', ') INTO v_missing
  FROM unnest(ARRAY['SELECT', 'INSERT', 'UPDATE', 'DELETE']) AS expected
  WHERE NOT EXISTS (
    SELECT 1 FROM information_schema.table_privileges
    WHERE table_schema = 'public' AND table_name = 'working_hour'
      AND grantee = 'barberia_app' AND privilege_type = expected
  );
  IF v_missing IS NOT NULL THEN
    RAISE EXCEPTION 'barberia_app debe tener % sobre working_hour.', v_missing;
  END IF;
END
$$;
\echo 'grants OK · barberia_app tiene SELECT/INSERT/UPDATE/DELETE sobre working_hour'

-- ---------------------------------------------------------------------------
-- CA-040-05 · El contexto de K no ve, no puede actualizar ni retirar un
-- tramo de L (aislamiento de tenant con identificador real)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'dddddddd-dddd-dddd-dddd-dddddddddddd';

INSERT INTO working_hour (id, barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
VALUES (
  'ffff0001-0001-0001-0001-000100010001', 'dddddddd-dddd-dddd-dddd-dddddddddddd',
  'dddd0001-0001-0001-0001-000100010001', 3, '09:00', 120
);

RESET ROLE;
COMMIT;

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'cccccccc-cccc-cccc-cccc-cccccccccccc';

DO $$
DECLARE
  v_tramo_l_id uuid := 'ffff0001-0001-0001-0001-000100010001';
  v_visible integer;
  v_updated integer;
  v_deleted integer;
BEGIN
  SELECT count(*) INTO v_visible FROM working_hour WHERE id = v_tramo_l_id;
  IF v_visible <> 0 THEN
    RAISE EXCEPTION 'CA-040-05: el contexto de K puede ver el tramo de L por su id real.';
  END IF;

  UPDATE working_hour SET duration_minutes = 999 WHERE id = v_tramo_l_id;
  GET DIAGNOSTICS v_updated = ROW_COUNT;
  IF v_updated <> 0 THEN
    RAISE EXCEPTION 'CA-040-05: el contexto de K pudo editar el tramo de L (% filas).', v_updated;
  END IF;

  DELETE FROM working_hour WHERE id = v_tramo_l_id;
  GET DIAGNOSTICS v_deleted = ROW_COUNT;
  IF v_deleted <> 0 THEN
    RAISE EXCEPTION 'CA-040-05: el contexto de K pudo retirar el tramo de L (% filas).', v_deleted;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-040-05 OK · el contexto de K no ve, no edita y no retira el tramo de L'

-- Limpieza del tramo de L confirmado arriba con COMMIT.
DELETE FROM working_hour WHERE id = 'ffff0001-0001-0001-0001-000100010001';

-- ---------------------------------------------------------------------------
-- DEC-070/CT-008 · La FK hacia barber usa ON DELETE RESTRICT, no CASCADE:
-- un barbero con al menos un tramo no se puede borrar físicamente. Se
-- ejecuta con el rol administrativo (barberia_app no tiene GRANT DELETE
-- sobre barber, DEC-047): esto prueba la FK en sí, no un permiso.
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'cccccccc-cccc-cccc-cccc-cccccccccccc';

INSERT INTO working_hour (barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
VALUES (
  'cccccccc-cccc-cccc-cccc-cccccccccccc', 'cccc0002-0002-0002-0002-000200020002', 4, '10:00', 60
);

RESET ROLE;
COMMIT;

DO $$
BEGIN
  BEGIN
    DELETE FROM barber WHERE id = 'cccc0002-0002-0002-0002-000200020002';
    RAISE EXCEPTION 'DEC-070/CT-008: se permitió borrar un barber con un working_hour asociado (CASCADE en vez de RESTRICT).';
  EXCEPTION
    WHEN foreign_key_violation THEN NULL; -- Esperado: working_hour_barbershop_id_barber_id_fk en RESTRICT.
  END;
END
$$;

-- Limpieza del tramo confirmado arriba.
DELETE FROM working_hour WHERE barber_id = 'cccc0002-0002-0002-0002-000200020002';
\echo 'DEC-070/CT-008 OK · borrar un barber con horario asociado se rechaza (ON DELETE RESTRICT, no CASCADE)'

-- ---------------------------------------------------------------------------
-- Trigger de updated_at (dos transacciones separadas, mismo criterio que
-- hu021_barberos.sql: now() es transaction_timestamp())
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'cccccccc-cccc-cccc-cccc-cccccccccccc';

INSERT INTO working_hour (id, barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
VALUES (
  'eeee0001-0001-0001-0001-000100010001', 'cccccccc-cccc-cccc-cccc-cccccccccccc',
  'cccc0001-0001-0001-0001-000100010001', 6, '08:00', 120
);

RESET ROLE;
COMMIT;

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'cccccccc-cccc-cccc-cccc-cccccccccccc';

DO $$
DECLARE
  v_updated_before timestamptz;
  v_updated_after timestamptz;
BEGIN
  SELECT updated_at INTO v_updated_before FROM working_hour WHERE id = 'eeee0001-0001-0001-0001-000100010001';

  PERFORM pg_sleep(0.01);

  UPDATE working_hour SET duration_minutes = 180 WHERE id = 'eeee0001-0001-0001-0001-000100010001'
  RETURNING updated_at INTO v_updated_after;

  IF v_updated_after <= v_updated_before THEN
    RAISE EXCEPTION 'working_hour_set_updated_at: updated_at no avanzó tras el UPDATE (antes=%, después=%).',
      v_updated_before, v_updated_after;
  END IF;
END
$$;

RESET ROLE;
COMMIT;

-- Limpieza de la fila confirmada para el trigger.
DELETE FROM working_hour WHERE id = 'eeee0001-0001-0001-0001-000100010001';

\echo 'trigger OK · working_hour_set_updated_at avanza updated_at en cada UPDATE'

\echo '=== HU-040 · todas las comprobaciones pasaron ==='
