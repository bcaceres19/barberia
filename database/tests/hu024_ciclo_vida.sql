-- Pruebas SQL de HU-024: desactivación y reactivación de servicios. Cubren,
-- a nivel de base de datos, lo que HU-022 (20260824140000_create_service.sql)
-- ya preparó pero no ejercía todavía con el rol de aplicación real:
-- service_deactivated_at_ck en ambas direcciones, la transición real
-- is_active/deactivated_at con barberia_app, aislamiento RLS entre
-- barberías sobre esa misma transición, y la reutilización de nombre que
-- CA-024-05/idx_service_active_name habilitan. La orquestación completa
-- (idempotencia, transición inválida, la carrera real de dos conexiones)
-- vive en el repositorio Go (internal/modules/catalog/postgres), probada
-- aparte contra PostgreSQL real. RLS forzada, ausencia de GRANT/política
-- DELETE y el esquema mínimo de `service` ya están cubiertos por
-- tests/hu022_catalogo.sql: este archivo no los repite.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/hu024_ciclo_vida.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- (estandar-base-datos.md §9.8), igual que hu022_catalogo.sql/hu023_asignaciones.sql.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== HU-024 · pruebas de ciclo de vida de servicios ==='

-- ---------------------------------------------------------------------------
-- service_deactivated_at_ck existe (defensa estructural; DEC-069/RN-SER-03)
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint c
    JOIN pg_class t ON t.oid = c.conrelid
    WHERE t.relname = 'service' AND c.conname = 'service_deactivated_at_ck'
  ) THEN
    RAISE EXCEPTION 'service_deactivated_at_ck debe existir sobre service.';
  END IF;
END
$$;
\echo 'constraint OK · service_deactivated_at_ck existe'

-- ---------------------------------------------------------------------------
-- service_deactivated_at_ck rechaza is_active=false con deactivated_at NULL
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  BEGIN
    INSERT INTO service (barbershop_id, name, duration_minutes, price_amount, is_active, deactivated_at)
    VALUES ('11111111-1111-1111-1111-111111111111', 'HU024 Inactivo sin marca ' || gen_random_uuid(), 30, 20000, false, NULL);
    RAISE EXCEPTION 'service_deactivated_at_ck: se permitió is_active=false con deactivated_at NULL.';
  EXCEPTION
    WHEN check_violation THEN NULL; -- Esperado.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'service_deactivated_at_ck OK · is_active=false exige deactivated_at NOT NULL'

-- ---------------------------------------------------------------------------
-- service_deactivated_at_ck rechaza is_active=true con deactivated_at fijado
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  BEGIN
    INSERT INTO service (barbershop_id, name, duration_minutes, price_amount, is_active, deactivated_at)
    VALUES ('11111111-1111-1111-1111-111111111111', 'HU024 Activo con marca ' || gen_random_uuid(), 30, 20000, true, now());
    RAISE EXCEPTION 'service_deactivated_at_ck: se permitió is_active=true con deactivated_at fijado.';
  EXCEPTION
    WHEN check_violation THEN NULL; -- Esperado.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'service_deactivated_at_ck OK · is_active=true exige deactivated_at NULL'

-- ---------------------------------------------------------------------------
-- CA-024-02 · barberia_app puede transicionar activo -> inactivo, misma fila
-- (RN-SER-03: nunca otra fila, nunca DELETE)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_service   uuid;
  v_is_active boolean;
  v_deact     timestamptz;
  v_count     integer;
BEGIN
  INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
  VALUES ('11111111-1111-1111-1111-111111111111', 'HU024 A desactivar ' || gen_random_uuid(), 30, 20000)
  RETURNING id INTO v_service;

  UPDATE service SET is_active = false, deactivated_at = now()
   WHERE id = v_service AND barbershop_id = '11111111-1111-1111-1111-111111111111';

  SELECT is_active, deactivated_at INTO v_is_active, v_deact FROM service WHERE id = v_service;
  IF v_is_active OR v_deact IS NULL THEN
    RAISE EXCEPTION 'CA-024-02: esperado is_active=false y deactivated_at fijado tras la transición, got is_active=% deactivated_at=%', v_is_active, v_deact;
  END IF;

  SELECT count(*) INTO v_count FROM service WHERE id = v_service;
  IF v_count <> 1 THEN
    RAISE EXCEPTION 'RN-SER-03: se esperaba EXACTAMENTE 1 fila (nunca duplicada ni borrada), hay %.', v_count;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-024-02 OK · barberia_app transiciona activo -> inactivo sobre la MISMA fila'

-- ---------------------------------------------------------------------------
-- CA-024-05 · barberia_app puede transicionar inactivo -> activo sin crear
-- otra fila ni alterar duración/precio
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_service   uuid;
  v_duration  integer;
  v_price     numeric(12,2);
  v_is_active boolean;
  v_deact     timestamptz;
BEGIN
  INSERT INTO service (barbershop_id, name, duration_minutes, price_amount, is_active, deactivated_at)
  VALUES ('11111111-1111-1111-1111-111111111111', 'HU024 A reactivar ' || gen_random_uuid(), 45, 30000, false, now())
  RETURNING id, duration_minutes, price_amount INTO v_service, v_duration, v_price;

  UPDATE service SET is_active = true, deactivated_at = NULL
   WHERE id = v_service AND barbershop_id = '11111111-1111-1111-1111-111111111111';

  SELECT is_active, deactivated_at INTO v_is_active, v_deact FROM service WHERE id = v_service;
  IF NOT v_is_active OR v_deact IS NOT NULL THEN
    RAISE EXCEPTION 'CA-024-05: esperado is_active=true y deactivated_at=NULL tras reactivar, got is_active=% deactivated_at=%', v_is_active, v_deact;
  END IF;

  IF (SELECT duration_minutes FROM service WHERE id = v_service) <> v_duration
     OR (SELECT price_amount FROM service WHERE id = v_service) <> v_price THEN
    RAISE EXCEPTION 'CA-024-05: reactivar no debe alterar duration_minutes ni price_amount.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-024-05 OK · barberia_app reactiva sin crear otra fila ni alterar duración/precio'

-- ---------------------------------------------------------------------------
-- CA-024-07/RN-TEN-01 · la transición nunca cruza de barbería, aun con el
-- id real de la fila ajena (defensa en profundidad + RLS forzada)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';
INSERT INTO service (id, barbershop_id, name, duration_minutes, price_amount)
VALUES ('77777777-2222-2222-2222-777777777772', '11111111-1111-1111-1111-111111111111', 'HU024 Servicio de A ' || gen_random_uuid(), 30, 20000);
RESET ROLE;
COMMIT;

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '22222222-2222-2222-2222-222222222222';

DO $$
DECLARE
  v_is_active boolean;
BEGIN
  UPDATE service SET is_active = false, deactivated_at = now()
   WHERE id = '77777777-2222-2222-2222-777777777772' AND barbershop_id = '22222222-2222-2222-2222-222222222222';
  IF FOUND THEN
    RAISE EXCEPTION 'CA-024-07/RN-TEN-01: barbería B pudo transicionar un servicio de la barbería A.';
  END IF;

  SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';
  SELECT is_active INTO v_is_active FROM service WHERE id = '77777777-2222-2222-2222-777777777772';
  IF NOT v_is_active THEN
    RAISE EXCEPTION 'CA-024-07: el servicio de A debe permanecer activo tras el intento cruzado.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-024-07 OK · ninguna transición cruza de barbería (RLS forzada + filtro explícito)'

-- ---------------------------------------------------------------------------
-- CA-024-05 · desactivar libera el nombre para un servicio ACTIVO nuevo
-- (idx_service_active_name, DEC-067), ejercido con la transición real de
-- barberia_app (no con una fila insertada ya inactiva por el propietario,
-- a diferencia de tests/hu022_catalogo.sql)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_name    text := 'HU024 Nombre Reutilizable ' || gen_random_uuid();
  v_first   uuid;
  v_second  uuid;
BEGIN
  INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
  VALUES ('11111111-1111-1111-1111-111111111111', v_name, 30, 20000)
  RETURNING id INTO v_first;

  -- Mientras está activo, un segundo servicio con el mismo nombre choca.
  BEGIN
    INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
    VALUES ('11111111-1111-1111-1111-111111111111', v_name, 30, 20000);
    RAISE EXCEPTION 'idx_service_active_name: se permitieron dos servicios ACTIVOS con el mismo nombre.';
  EXCEPTION
    WHEN unique_violation THEN NULL; -- Esperado.
  END;

  UPDATE service SET is_active = false, deactivated_at = now()
   WHERE id = v_first AND barbershop_id = '11111111-1111-1111-1111-111111111111';

  -- Ya desactivado, el nombre queda libre para un servicio nuevo.
  INSERT INTO service (barbershop_id, name, duration_minutes, price_amount)
  VALUES ('11111111-1111-1111-1111-111111111111', v_name, 30, 20000)
  RETURNING id INTO v_second;

  IF v_second = v_first THEN
    RAISE EXCEPTION 'esperada una fila NUEVA, no la misma que se desactivó.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-024-05 OK · desactivar libera el nombre para un servicio activo nuevo (DEC-067)'

\echo '=== HU-024 · todas las pruebas pasaron ==='
