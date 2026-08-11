-- Suite RLS completa (issue #7, DDL-RLS-01): cada tabla protegida, cada
-- operación, sin contexto, tenant A/B, FORCE RLS, propietarios, sin
-- BYPASSRLS, con el rol real de la aplicación y del trabajador (nunca
-- superusuario para las comprobaciones de privilegio).
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/rls_suite_fixture.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/rls_suite.sql
--
-- Requisito del arnés: igual que hu001_aislamiento_rls.sql, la conexión debe
-- poder ejecutar `SET ROLE barberia_app`/`barberia_worker` sin pertenecer a
-- esos roles (conectada como superusuario en local/CI).
--
-- Relación con hu001_aislamiento_rls.sql: ese archivo ya cubre `barbershop`
-- y `staff_user` en detalle (CA-001-*); esta suite no los repite y cubre las
-- 23 tablas restantes con RLS forzada. `login_throttle` queda fuera a
-- propósito: es la única tabla sin `barbershop_id` por diseño (cuenta
-- intentos por IP antes de saber quién solicita), documentado en
-- modelo-fisico-referencia.sql sección A.4.
--
-- DDL-RLS-01 pide aceptar cualquier fallo fail-closed esperado, no un único
-- SQLSTATE accidental: los bloques de "sin contexto" y "sin privilegio de
-- tabla" aceptan tanto `undefined_object` (falta app.barbershop_id) como
-- `insufficient_privilege` (la tabla ni siquiera concede el verbo), porque
-- ambos son el mismo resultado correcto -acceso denegado- por razones
-- distintas y válidas.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== Suite RLS completa · issue #7 ==='

-- ---------------------------------------------------------------------------
-- Parte 1 · Estructura, por catálogo, para TODA tabla del esquema
-- ---------------------------------------------------------------------------
-- No depende de datos ni de fixture: recorre pg_class/pg_policies
-- directamente, así que cubre cualquier tabla presente, se le haya escrito
-- o no un caso de prueba ejecutado más abajo.

DO $$
DECLARE
  v_offenders text;
BEGIN
  -- RLS habilitada Y forzada en toda tabla, excepto la excepción documentada.
  SELECT string_agg(c.relname, ', ')
  INTO v_offenders
  FROM pg_class c
  JOIN pg_namespace n ON n.oid = c.relnamespace
  WHERE n.nspname = 'public' AND c.relkind = 'r'
    AND c.relname NOT IN ('atlas_schema_revisions', 'login_throttle')
    AND NOT (c.relrowsecurity AND c.relforcerowsecurity);
  IF v_offenders IS NOT NULL THEN
    RAISE EXCEPTION 'Tablas sin RLS habilitada y forzada: %', v_offenders;
  END IF;
END
$$;
\echo 'Parte 1.1 OK · RLS habilitada y forzada en todas las tablas'

DO $$
DECLARE
  v_offenders text;
BEGIN
  -- Toda tabla con RLS forzada tiene una política administrativa FOR ALL
  -- que apunta a barberia_owner (DEC-040), nunca a barberia_migrator.
  SELECT string_agg(c.relname, ', ')
  INTO v_offenders
  FROM pg_class c
  JOIN pg_namespace n ON n.oid = c.relnamespace
  WHERE n.nspname = 'public' AND c.relkind = 'r' AND c.relrowsecurity
    AND c.relname NOT IN ('atlas_schema_revisions', 'login_throttle')
    AND NOT EXISTS (
      SELECT 1 FROM pg_policies p
      WHERE p.schemaname = 'public' AND p.tablename = c.relname
        AND p.cmd = 'ALL' AND 'barberia_owner' = ANY(p.roles)
    );
  IF v_offenders IS NOT NULL THEN
    RAISE EXCEPTION 'Tablas sin política administrativa FOR ALL a barberia_owner: %', v_offenders;
  END IF;

  SELECT string_agg(DISTINCT p.tablename, ', ')
  INTO v_offenders
  FROM pg_policies p
  WHERE p.schemaname = 'public' AND 'barberia_migrator' = ANY(p.roles);
  IF v_offenders IS NOT NULL THEN
    RAISE EXCEPTION 'Políticas todavía apuntando a barberia_migrator (DDL-RLS-01/H.7): %', v_offenders;
  END IF;
END
$$;
\echo 'Parte 1.2 OK · política administrativa de cada tabla apunta a barberia_owner, ninguna a barberia_migrator'

DO $$
BEGIN
  -- Ni el rol del API ni el del worker son superusuario ni tienen BYPASSRLS
  -- (DEC-040, estandar-base-datos.md §9.3): sin esto, toda política de este
  -- archivo sería teatro.
  IF EXISTS (
    SELECT 1 FROM pg_roles
    WHERE rolname IN ('barberia_app', 'barberia_worker')
      AND (rolsuper OR rolbypassrls)
  ) THEN
    RAISE EXCEPTION 'barberia_app o barberia_worker tienen superusuario o BYPASSRLS.';
  END IF;
END
$$;
\echo 'Parte 1.3 OK · ni barberia_app ni barberia_worker tienen BYPASSRLS ni son superusuario'

DO $$
DECLARE
  v_offenders text;
BEGIN
  -- Cada política de lectura/escritura tenant-aware referencia de verdad
  -- barbershop_id y current_setting: una política SELECT/INSERT/UPDATE/
  -- DELETE de barberia_app cuyo USING/WITH CHECK no mencione ambos términos
  -- es, con altísima probabilidad, un copy-paste roto (columna equivocada,
  -- condición `true` residual) que ningún otro chequeo estructural detecta.
  SELECT string_agg(p.tablename || '.' || p.policyname, ', ')
  INTO v_offenders
  FROM pg_policies p
  WHERE p.schemaname = 'public'
    AND 'barberia_app' = ANY(p.roles)
    AND p.tablename NOT IN ('login_throttle')
    AND NOT (
      (p.qual       IS NOT NULL AND p.qual       LIKE '%barbershop_id%' AND p.qual       LIKE '%current_setting%')
      OR
      (p.with_check IS NOT NULL AND p.with_check LIKE '%barbershop_id%' AND p.with_check LIKE '%current_setting%')
    );
  IF v_offenders IS NOT NULL THEN
    RAISE EXCEPTION 'Políticas de barberia_app sin USING/WITH CHECK tenant-aware reconocible: %', v_offenders;
  END IF;
END
$$;
\echo 'Parte 1.4 OK · toda política de barberia_app referencia barbershop_id y current_setting'

-- ---------------------------------------------------------------------------
-- Parte 2 · Sin contexto: toda tabla falla cerrado, con cualquier SQLSTATE
-- fail-closed válido (DDL-RLS-01: no exigir uno accidental)
-- ---------------------------------------------------------------------------

DO $$
DECLARE
  v_table  text;
  v_tables constant text[] := ARRAY[
    'staff_credential', 'staff_session', 'staff_recovery_code',
    'barber', 'service', 'barber_service',
    'working_hour', 'working_hour_override', 'working_hour_override_segment',
    'time_block_series', 'time_block_series_date', 'time_block_series_exception', 'time_block',
    'customer', 'appointment', 'appointment_history', 'appointment_history_change',
    'appointment_access_token', 'notification_channel_setting', 'barbershop_reminder_rule',
    'notification_schedule', 'notification_attempt', 'idempotency_record'
  ];
  v_count  integer;
BEGIN
  SET ROLE barberia_app;
  FOREACH v_table IN ARRAY v_tables LOOP
    BEGIN
      EXECUTE format('SELECT count(*) FROM %I', v_table) INTO v_count;
      RAISE EXCEPTION 'Sin contexto, % devolvió % filas en lugar de fallar.', v_table, v_count;
    EXCEPTION
      WHEN undefined_object OR insufficient_privilege THEN
        NULL;  -- Esperado: cualquiera de los dos es fail-closed correcto.
    END;
  END LOOP;
  RESET ROLE;
END
$$;
\echo 'Parte 2 OK · sin app.barbershop_id, las 23 tablas fallan cerrado (undefined_object o insufficient_privilege)'

-- ---------------------------------------------------------------------------
-- Parte 3 · Tenant A/B: con contexto de A, cero filas de B (DDL-RLS-01)
-- ---------------------------------------------------------------------------
-- staff_credential queda fuera: no concede SELECT a barberia_app en
-- absoluto (auth_get_credential es el único camino de lectura), así que no
-- hay nada que "ver filtrado" — ya está cubierto por la Parte 2.

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_table  text;
  v_tables constant text[] := ARRAY[
    'staff_session', 'staff_recovery_code',
    'barber', 'service', 'barber_service',
    'working_hour', 'working_hour_override', 'working_hour_override_segment',
    'time_block_series', 'time_block_series_date', 'time_block_series_exception', 'time_block',
    'customer', 'appointment', 'appointment_history', 'appointment_history_change',
    'appointment_access_token', 'notification_channel_setting', 'barbershop_reminder_rule',
    'notification_schedule', 'notification_attempt', 'idempotency_record'
  ];
  v_count integer;
BEGIN
  FOREACH v_table IN ARRAY v_tables LOOP
    EXECUTE format(
      'SELECT count(*) FROM %I WHERE barbershop_id = ''22222222-2222-2222-2222-222222222222''',
      v_table
    ) INTO v_count;
    IF v_count <> 0 THEN
      RAISE EXCEPTION 'Con contexto de A, % expuso % filas de B.', v_table, v_count;
    END IF;
  END LOOP;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Parte 3 OK · con contexto de A, las 21 tablas restantes exponen cero filas de B'

-- ---------------------------------------------------------------------------
-- Parte 4 · CRUD permitido/denegado por tabla y rol
-- ---------------------------------------------------------------------------
-- Ejecutado (no solo por catálogo): prueba el privilegio de tabla real y el
-- WITH CHECK real, agrupado por verbo para no repetir el mismo bloque 23
-- veces. Usa las filas del fixture (sufijo `a`/`b`).

-- 4.1 · INSERT propio permitido; INSERT con barbershop_id ajeno rechazado.
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  -- Permitido: una fila nueva y propia en una tabla representativa de cada
  -- forma de INSERT (simple, compuesta, PK generada, PK compuesta sin id).
  INSERT INTO barber (id, barbershop_id, full_name)
  VALUES ('11110000-0000-0000-0000-000000000001', '11111111-1111-1111-1111-111111111111', 'Alta permitida');

  INSERT INTO notification_channel_setting (barbershop_id, event_type, channel)
  VALUES ('11111111-1111-1111-1111-111111111111', 'reminder', 'email');

  -- Denegado: barbershop_id ajeno en el payload, tabla con PK simple.
  BEGIN
    INSERT INTO barber (id, barbershop_id, full_name)
    VALUES ('22220000-0000-0000-0000-000000000002', '22222222-2222-2222-2222-222222222222', 'Alta cruzada');
    RAISE EXCEPTION 'Se insertó barber con el barbershop_id de otro tenant.';
  EXCEPTION
    WHEN insufficient_privilege THEN NULL;  -- Esperado.
  END;

  -- Denegado: barbershop_id ajeno, tabla con PK compuesta (sin columna id).
  BEGIN
    INSERT INTO notification_channel_setting (barbershop_id, event_type, channel)
    VALUES ('22222222-2222-2222-2222-222222222222', 'delay', 'whatsapp');
    RAISE EXCEPTION 'Se insertó notification_channel_setting con el barbershop_id de otro tenant.';
  EXCEPTION
    WHEN insufficient_privilege THEN NULL;  -- Esperado.
  END;

  -- Denegado: tabla append-only sin política de INSERT hacia otro tenant.
  BEGIN
    INSERT INTO appointment_history (barbershop_id, appointment_id, event_type, actor_type)
    VALUES ('22222222-2222-2222-2222-222222222222', 'a99a0000-0000-0000-0000-00000000000b',
            'appointment_completed', 'system');
    RAISE EXCEPTION 'Se insertó appointment_history con el barbershop_id de otro tenant.';
  EXCEPTION
    WHEN insufficient_privilege THEN NULL;  -- Esperado.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Parte 4.1 OK · INSERT propio permitido; INSERT con barbershop_id ajeno rechazado (PK simple, compuesta y append-only)'

-- 4.2 · UPDATE propio permitido; UPDATE de fila ajena o hacia otro tenant rechazado.
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_rows integer;
BEGIN
  -- Permitido: modificar una fila propia.
  UPDATE barber SET full_name = 'Barbero RLS A editado'
  WHERE id = 'ba5b0000-0000-0000-0000-00000000000a';

  -- Denegado por USING: una fila de B es invisible para A, así que el
  -- UPDATE afecta cero filas en vez de fallar (RN-TEN-01: indistinguible de
  -- "no existe").
  UPDATE barber SET full_name = 'Secuestrado'
  WHERE id = 'ba5b0000-0000-0000-0000-00000000000b';
  GET DIAGNOSTICS v_rows = ROW_COUNT;
  IF v_rows <> 0 THEN
    RAISE EXCEPTION 'UPDATE alcanzó % filas de otro tenant en barber.', v_rows;
  END IF;

  -- Denegado por WITH CHECK: mover una fila propia hacia el tenant ajeno.
  BEGIN
    UPDATE barber SET barbershop_id = '22222222-2222-2222-2222-222222222222'
    WHERE id = 'ba5b0000-0000-0000-0000-00000000000a';
    RAISE EXCEPTION 'Se movió un barber propio hacia otro tenant.';
  EXCEPTION
    WHEN insufficient_privilege THEN NULL;  -- Esperado.
  END;

  -- Repite el mismo par de pruebas sobre una tabla con UPDATE acotado a
  -- columnas concretas (notification_schedule, DEC-053): también debe fallar
  -- cerrado y no filtrar filas de B.
  UPDATE notification_schedule SET status = 'cancelled', cancelled_at = now()
  WHERE id = '501e0000-0000-0000-0000-00000000000b';
  GET DIAGNOSTICS v_rows = ROW_COUNT;
  IF v_rows <> 0 THEN
    RAISE EXCEPTION 'UPDATE alcanzó % filas de otro tenant en notification_schedule.', v_rows;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Parte 4.2 OK · UPDATE propio permitido; fila ajena invisible (0 filas) y WITH CHECK impide moverla de tenant'

-- 4.3 · DELETE propio permitido; DELETE de fila ajena no afecta nada.
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_rows integer;
BEGIN
  -- Denegado por USING: DELETE de una fila de B no afecta nada bajo
  -- contexto de A (tablas con DELETE concedido a barberia_app).
  DELETE FROM staff_session WHERE id = '50550000-0000-0000-0000-00000000000b';
  GET DIAGNOSTICS v_rows = ROW_COUNT;
  IF v_rows <> 0 THEN
    RAISE EXCEPTION 'DELETE alcanzó % filas de otro tenant en staff_session.', v_rows;
  END IF;

  DELETE FROM working_hour WHERE id = '90570000-0000-0000-0000-00000000000b';
  GET DIAGNOSTICS v_rows = ROW_COUNT;
  IF v_rows <> 0 THEN
    RAISE EXCEPTION 'DELETE alcanzó % filas de otro tenant en working_hour.', v_rows;
  END IF;

  -- Permitido: borrar una fila propia (staff_session, revocación por
  -- borrado según CA-008-05).
  DELETE FROM staff_session WHERE id = '50550000-0000-0000-0000-00000000000a';
  GET DIAGNOSTICS v_rows = ROW_COUNT;
  IF v_rows <> 1 THEN
    RAISE EXCEPTION 'No se pudo borrar una fila propia de staff_session.';
  END IF;

  -- time_block NO concede DELETE en absoluto (RN-BLQ-04/DEC-009: se
  -- elimina lógicamente vía deleted_at, nunca con DELETE real), ni siquiera
  -- sobre una fila propia.
  BEGIN
    DELETE FROM time_block WHERE id = 'b10c0000-0000-0000-0000-00000000000a';
    RAISE EXCEPTION 'Se pudo hacer DELETE real sobre time_block (debe ser solo eliminación lógica).';
  EXCEPTION
    WHEN insufficient_privilege THEN NULL;  -- Esperado.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Parte 4.3 OK · DELETE propio permitido; DELETE de fila ajena no afecta nada; time_block sin DELETE real'

-- 4.4 · Tablas append-only: ni UPDATE ni DELETE, ni siquiera sobre lo propio.
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  BEGIN
    UPDATE appointment_history SET reason = 'reescrito'
    WHERE id = '7157000a-0000-0000-0000-00000000000a';
    RAISE EXCEPTION 'Se pudo UPDATE una fila propia de appointment_history (debe ser append-only).';
  EXCEPTION
    WHEN insufficient_privilege THEN NULL;  -- Esperado.
  END;

  BEGIN
    DELETE FROM appointment_history WHERE id = '7157000a-0000-0000-0000-00000000000a';
    RAISE EXCEPTION 'Se pudo DELETE una fila propia de appointment_history (debe ser append-only).';
  EXCEPTION
    WHEN insufficient_privilege THEN NULL;  -- Esperado.
  END;

  BEGIN
    UPDATE appointment_history_change SET new_value = 'reescrito'
    WHERE barbershop_id = '11111111-1111-1111-1111-111111111111'
      AND history_id = '7157000a-0000-0000-0000-00000000000a' AND field_name = 'status';
    RAISE EXCEPTION 'Se pudo UPDATE una fila propia de appointment_history_change (debe ser append-only).';
  EXCEPTION
    WHEN insufficient_privilege THEN NULL;  -- Esperado.
  END;

  BEGIN
    UPDATE notification_attempt SET result = 'failed'
    WHERE barbershop_id = '11111111-1111-1111-1111-111111111111'
      AND notification_schedule_id = '501e0000-0000-0000-0000-00000000000a' AND attempt_number = 1;
    RAISE EXCEPTION 'Se pudo UPDATE una fila propia de notification_attempt (debe ser append-only).';
  EXCEPTION
    WHEN insufficient_privilege THEN NULL;  -- Esperado.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Parte 4.4 OK · appointment_history, appointment_history_change y notification_attempt son append-only incluso sobre filas propias'

-- ---------------------------------------------------------------------------
-- Parte 5 · Funciones SECURITY DEFINER y roles API/worker: cada una
-- exclusiva de su rol, nunca ambos por defecto (estandar-base-datos.md §9.10)
-- ---------------------------------------------------------------------------

DO $$
DECLARE
  v_offenders text;
BEGIN
  -- Ninguna función SECURITY DEFINER expuesta al API/worker concede EXECUTE
  -- a PUBLIC (defensa primaria de cada REVOKE ALL ... FROM PUBLIC).
  SELECT string_agg(p.proname, ', ')
  INTO v_offenders
  FROM pg_proc p
  JOIN pg_namespace n ON n.oid = p.pronamespace
  WHERE n.nspname = 'public' AND p.prosecdef
    AND has_function_privilege('public', p.oid, 'EXECUTE');
  IF v_offenders IS NOT NULL THEN
    RAISE EXCEPTION 'Funciones SECURITY DEFINER con EXECUTE para PUBLIC: %', v_offenders;
  END IF;
END
$$;
\echo 'Parte 5 OK · ninguna función SECURITY DEFINER concede EXECUTE a PUBLIC'

\echo '=== Suite RLS completa · todas las comprobaciones pasaron ==='
