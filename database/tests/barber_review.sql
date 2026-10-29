-- Pruebas SQL de DEC-128 / HU-100: reseña interna por barbero (barber_review).
-- Cubren a nivel de base de datos la unicidad por turno, el disparador que exige
-- un turno `completed` del mismo barbero, los CHECK de calificación y comentario,
-- el ocultamiento reversible (único campo editable), la ausencia de borrado y el
-- aislamiento RLS entre dos barberías. Normalizar el comentario y decidir qué se
-- publica viven en el servicio Go (internal/modules/reviews).
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/barber_review.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- (estandar-base-datos.md §9.8). Crea sus propios datos y todo se revierte con
-- ROLLBACK. Cada comprobación falla lanzando una excepción; si el archivo
-- termina, pasó.

\set ON_ERROR_STOP on

\echo '=== DEC-128 · pruebas de barber_review ==='

BEGIN;

-- Datos de dos barberías (rol administrativo de la sesión, antes de SET ROLE).
INSERT INTO barbershop (id, name, timezone) VALUES
  ('01000001-0100-4100-8100-010000000001', 'Barbería reseñas A', 'America/Bogota'),
  ('01000002-0100-4100-8100-010000000002', 'Barbería reseñas B', 'America/Bogota');

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('01000011-0100-4100-8100-010000000011', '01000001-0100-4100-8100-010000000001', 'Barbero A1'),
  ('01000012-0100-4100-8100-010000000012', '01000001-0100-4100-8100-010000000001', 'Barbero A2'),
  ('01000021-0100-4100-8100-010000000021', '01000002-0100-4100-8100-010000000002', 'Barbero B1');

INSERT INTO service (id, barbershop_id, name, duration_minutes, price_amount, price_currency) VALUES
  ('01000031-0100-4100-8100-010000000031', '01000001-0100-4100-8100-010000000001', 'Corte A', 30, 20000.00, 'COP'),
  ('01000032-0100-4100-8100-010000000032', '01000002-0100-4100-8100-010000000002', 'Corte B', 30, 20000.00, 'COP');

INSERT INTO customer (id, barbershop_id, full_name) VALUES
  ('01000041-0100-4100-8100-010000000041', '01000001-0100-4100-8100-010000000001', 'Cliente de prueba A'),
  ('01000042-0100-4100-8100-010000000042', '01000002-0100-4100-8100-010000000002', 'Cliente de prueba B');

-- Citas de A: una completada (A1), otra completada (A1), una confirmada (A1) y una
-- completada con el barbero A2. Una completada de B (B1).
INSERT INTO appointment (
  id, barbershop_id, barber_id, service_id, customer_id, attendee_name,
  starts_at, ends_at, status, origin, resolved_at,
  service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
) VALUES
  ('01000051-0100-4100-8100-010000000051', '01000001-0100-4100-8100-010000000001', '01000011-0100-4100-8100-010000000011',
   '01000031-0100-4100-8100-010000000031', '01000041-0100-4100-8100-010000000041', 'Cliente de prueba A',
   '2026-01-05T10:00:00Z', '2026-01-05T10:30:00Z', 'completed', 'manual', now(), 'Corte A', 30, 20000.00, 'COP'),
  ('01000052-0100-4100-8100-010000000052', '01000001-0100-4100-8100-010000000001', '01000011-0100-4100-8100-010000000011',
   '01000031-0100-4100-8100-010000000031', '01000041-0100-4100-8100-010000000041', 'Cliente de prueba A',
   '2026-01-06T10:00:00Z', '2026-01-06T10:30:00Z', 'completed', 'manual', now(), 'Corte A', 30, 20000.00, 'COP'),
  ('01000053-0100-4100-8100-010000000053', '01000001-0100-4100-8100-010000000001', '01000011-0100-4100-8100-010000000011',
   '01000031-0100-4100-8100-010000000031', '01000041-0100-4100-8100-010000000041', 'Cliente de prueba A',
   '2099-01-07T10:00:00Z', '2099-01-07T10:30:00Z', 'confirmed', 'manual', NULL, 'Corte A', 30, 20000.00, 'COP'),
  ('01000054-0100-4100-8100-010000000054', '01000001-0100-4100-8100-010000000001', '01000012-0100-4100-8100-010000000012',
   '01000031-0100-4100-8100-010000000031', '01000041-0100-4100-8100-010000000041', 'Cliente de prueba A',
   '2026-01-08T10:00:00Z', '2026-01-08T10:30:00Z', 'completed', 'manual', now(), 'Corte A', 30, 20000.00, 'COP'),
  ('01000061-0100-4100-8100-010000000061', '01000002-0100-4100-8100-010000000002', '01000021-0100-4100-8100-010000000021',
   '01000032-0100-4100-8100-010000000032', '01000042-0100-4100-8100-010000000042', 'Cliente de prueba B',
   '2026-01-05T10:00:00Z', '2026-01-05T10:30:00Z', 'completed', 'manual', now(), 'Corte B', 30, 20000.00, 'COP');

SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '01000001-0100-4100-8100-010000000001';

-- ---------------------------------------------------------------------------
-- Una reseña válida; comentario NULL permitido
-- ---------------------------------------------------------------------------
INSERT INTO barber_review (id, barbershop_id, barber_id, appointment_id, rating, comment)
VALUES ('01000071-0100-4100-8100-010000000071', '01000001-0100-4100-8100-010000000001',
        '01000011-0100-4100-8100-010000000011', '01000051-0100-4100-8100-010000000051', 5, 'Excelente corte.');
INSERT INTO barber_review (id, barbershop_id, barber_id, appointment_id, rating, comment)
VALUES ('01000072-0100-4100-8100-010000000072', '01000001-0100-4100-8100-010000000001',
        '01000011-0100-4100-8100-010000000011', '01000052-0100-4100-8100-010000000052', 3, NULL);
\echo 'inserción OK · con y sin comentario'

-- ---------------------------------------------------------------------------
-- Unicidad por turno (CA-100-03)
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_rejected boolean := false;
BEGIN
  BEGIN
    INSERT INTO barber_review (barbershop_id, barber_id, appointment_id, rating)
    VALUES ('01000001-0100-4100-8100-010000000001', '01000011-0100-4100-8100-010000000011',
            '01000051-0100-4100-8100-010000000051', 1);
  EXCEPTION WHEN unique_violation THEN
    v_rejected := true;
  END;
  IF NOT v_rejected THEN
    RAISE EXCEPTION 'barber_review_appointment_id_uk debió rechazar una segunda reseña del mismo turno.';
  END IF;
  IF (SELECT rating FROM barber_review WHERE appointment_id = '01000051-0100-4100-8100-010000000051') <> 5 THEN
    RAISE EXCEPTION 'el segundo intento no debe modificar la reseña existente.';
  END IF;
END
$$;
\echo 'unicidad OK · una reseña por turno'

-- ---------------------------------------------------------------------------
-- Solo un turno completado y del mismo barbero (disparador)
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_rejected boolean := false;
BEGIN
  -- Turno confirmado.
  BEGIN
    INSERT INTO barber_review (barbershop_id, barber_id, appointment_id, rating)
    VALUES ('01000001-0100-4100-8100-010000000001', '01000011-0100-4100-8100-010000000011',
            '01000053-0100-4100-8100-010000000053', 4);
  EXCEPTION WHEN check_violation THEN
    v_rejected := true;
  END;
  IF NOT v_rejected THEN
    RAISE EXCEPTION 'el disparador debió rechazar una reseña de un turno no completado.';
  END IF;

  -- Barbero distinto al del turno (turno de A2 reseñando a A1).
  v_rejected := false;
  BEGIN
    INSERT INTO barber_review (barbershop_id, barber_id, appointment_id, rating)
    VALUES ('01000001-0100-4100-8100-010000000001', '01000011-0100-4100-8100-010000000011',
            '01000054-0100-4100-8100-010000000054', 4);
  EXCEPTION WHEN check_violation THEN
    v_rejected := true;
  END;
  IF NOT v_rejected THEN
    RAISE EXCEPTION 'el disparador debió rechazar un barbero distinto al del turno.';
  END IF;
END
$$;
\echo 'disparador OK · turno completado y barbero del turno'

-- ---------------------------------------------------------------------------
-- CHECK de calificación y comentario (CA-100-02)
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_rejected boolean;
  v_rating   smallint;
  v_comment  text;
BEGIN
  FOREACH v_rating IN ARRAY ARRAY[0, 6, -1]::smallint[] LOOP
    v_rejected := false;
    BEGIN
      INSERT INTO barber_review (barbershop_id, barber_id, appointment_id, rating)
      VALUES ('01000001-0100-4100-8100-010000000001', '01000012-0100-4100-8100-010000000012',
              '01000054-0100-4100-8100-010000000054', v_rating);
    EXCEPTION WHEN check_violation THEN
      v_rejected := true;
    END;
    IF NOT v_rejected THEN
      RAISE EXCEPTION 'barber_review_rating_ck debió rechazar la calificación %.', v_rating;
    END IF;
  END LOOP;

  FOREACH v_comment IN ARRAY ARRAY['', '   ', ' con espacios ', repeat('a', 501)] LOOP
    v_rejected := false;
    BEGIN
      INSERT INTO barber_review (barbershop_id, barber_id, appointment_id, rating, comment)
      VALUES ('01000001-0100-4100-8100-010000000001', '01000012-0100-4100-8100-010000000012',
              '01000054-0100-4100-8100-010000000054', 4, v_comment);
    EXCEPTION WHEN check_violation THEN
      v_rejected := true;
    END;
    IF NOT v_rejected THEN
      RAISE EXCEPTION 'barber_review_comment_ck debió rechazar un comentario de % caracteres.', char_length(v_comment);
    END IF;
  END LOOP;

  -- 500 caracteres exactos son válidos.
  INSERT INTO barber_review (barbershop_id, barber_id, appointment_id, rating, comment)
  VALUES ('01000001-0100-4100-8100-010000000001', '01000012-0100-4100-8100-010000000012',
          '01000054-0100-4100-8100-010000000054', 4, repeat('ñ', 500));
END
$$;
\echo 'CHECK OK · calificación 1–5, comentario NULL o 1–500 sin espacios en los extremos'

-- ---------------------------------------------------------------------------
-- Ocultar es reversible y solo toca hidden_at; no hay DELETE (CA-100-05)
-- ---------------------------------------------------------------------------
UPDATE barber_review SET hidden_at = now() WHERE id = '01000071-0100-4100-8100-010000000071';
UPDATE barber_review SET hidden_at = NULL  WHERE id = '01000071-0100-4100-8100-010000000071';

DO $$
DECLARE
  v_denied boolean;
BEGIN
  IF (SELECT comment FROM barber_review WHERE id = '01000071-0100-4100-8100-010000000071') <> 'Excelente corte.' THEN
    RAISE EXCEPTION 'ocultar y mostrar no debe perder el contenido de la reseña.';
  END IF;

  v_denied := false;
  BEGIN
    UPDATE barber_review SET rating = 1 WHERE id = '01000071-0100-4100-8100-010000000071';
  EXCEPTION WHEN insufficient_privilege THEN
    v_denied := true;
  END;
  IF NOT v_denied THEN
    RAISE EXCEPTION 'la aplicación no debe poder cambiar la calificación de una reseña.';
  END IF;

  v_denied := false;
  BEGIN
    DELETE FROM barber_review WHERE id = '01000071-0100-4100-8100-010000000071';
  EXCEPTION WHEN insufficient_privilege THEN
    v_denied := true;
  END;
  IF NOT v_denied THEN
    RAISE EXCEPTION 'la aplicación no debe poder borrar una reseña.';
  END IF;
END
$$;
\echo 'ocultar OK · reversible, solo hidden_at, sin borrado'

-- ---------------------------------------------------------------------------
-- RLS: la barbería B no ve, no oculta y no reseña turnos de A
-- ---------------------------------------------------------------------------
SET LOCAL app.barbershop_id = '01000002-0100-4100-8100-010000000002';

DO $$
DECLARE
  v_rows     integer;
  v_rejected boolean := false;
BEGIN
  IF (SELECT count(*) FROM barber_review) <> 0 THEN
    RAISE EXCEPTION 'la barbería B no debe ver reseñas de A.';
  END IF;

  UPDATE barber_review SET hidden_at = now() WHERE id = '01000071-0100-4100-8100-010000000071';
  GET DIAGNOSTICS v_rows = ROW_COUNT;
  IF v_rows <> 0 THEN
    RAISE EXCEPTION 'la barbería B no debe poder ocultar una reseña de A.';
  END IF;

  BEGIN
    INSERT INTO barber_review (barbershop_id, barber_id, appointment_id, rating)
    VALUES ('01000001-0100-4100-8100-010000000001', '01000011-0100-4100-8100-010000000011',
            '01000051-0100-4100-8100-010000000051', 1);
  EXCEPTION WHEN insufficient_privilege OR unique_violation OR check_violation THEN
    v_rejected := true;
  END;
  IF NOT v_rejected THEN
    RAISE EXCEPTION 'la barbería B no debe poder insertar una reseña con el identificador de A.';
  END IF;

  -- B sí reseña su propio turno.
  INSERT INTO barber_review (barbershop_id, barber_id, appointment_id, rating)
  VALUES ('01000002-0100-4100-8100-010000000002', '01000021-0100-4100-8100-010000000021',
          '01000061-0100-4100-8100-010000000061', 2);
  IF (SELECT count(*) FROM barber_review) <> 1 THEN
    RAISE EXCEPTION 'la barbería B debe ver solo su propia reseña.';
  END IF;
END
$$;
\echo 'RLS OK · dos barberías aisladas'

ROLLBACK;

\echo '=== barber_review · todas las comprobaciones pasaron ==='
