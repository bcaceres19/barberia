-- Propósito
--   Issue #357 (HU-100, DEC-128): crea `barber_review`, la reseña interna que un
--   cliente deja sobre el barbero que lo atendió, una sola por turno y solo
--   cuando el turno está `completed`. Las estrellas del barbero (`rating`,
--   `ratingCount`) se calculan en cada consulta sobre las reseñas no ocultas;
--   no hay columna derivada ni disparador de agregados.
--
-- Reglas y decisiones
--   RN-TEN-01, RN-DAT-02, DEC-024 (esquema compartido con RLS), DEC-035/DEC-036
--   (estándar de base de datos, Atlas, migración inmutable), DEC-089 (token de
--   acceso al turno), DEC-104 (ocultar reversible sin borrado físico) y DEC-128.
--
-- Modelo
--   `barber_review (id, barbershop_id, barber_id, appointment_id, rating,
--   comment, hidden_at, created_at)`.
--   * `UNIQUE (appointment_id)`: como máximo una reseña por turno, nunca por
--     cliente ni por barbero (un cliente que vuelve genera otro turno).
--   * Claves foráneas compuestas (barbershop_id, …) hacia `barber` y
--     `appointment` con ON DELETE RESTRICT: ninguna reseña apunta a otra
--     barbería aun con SQL directo bajo el rol de aplicación.
--   * `rating` 1–5 y `comment` NULL o 1–500 caracteres sin espacios sobrantes en
--     los extremos; vacío o solo espacios llega como NULL desde el servicio.
--   * `hidden_at` no nulo = oculta (DEC-128 §4). Ocultar y volver a mostrar es un
--     UPDATE de esa única columna; la fila nunca se borra.
--   * Un disparador BEFORE INSERT exige que el turno esté `completed` y que
--     `barber_id` sea el barbero de ese turno: la regla de DEC-128 §1 queda
--     garantizada también contra SQL directo, no solo en el servicio Go.
--
-- Privilegios y RLS
--   RLS forzada. La aplicación puede SELECT, INSERT y UPDATE solo de `hidden_at`
--   (privilegio por columna); sin DELETE.
--
-- Qué NO hace esta migración
--   No guarda al cliente (ni nombre ni contacto), no agrega respuesta del
--   barbero, edición de la reseña ni cola de aprobación (fuera de alcance de
--   DEC-128) y no toca `appointment` ni `barber`.
--
-- Clasificación de datos
--   `comment` es texto libre que el propio cliente eligió publicar sobre su
--   experiencia; el formulario advierte que no escriba teléfonos, direcciones ni
--   nombres de terceros y nunca se registra en logs (RN-DAT-02). La reseña no
--   guarda ningún dato de contacto del cliente.
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE, ver
--   20260811145252_harden_roles_and_definer_functions.sql).
--
-- Plan de avance
--   Una corrección posterior se hace con otra migración, nunca editando esta
--   (DEC-036). No existe archivo `down`: el mecanismo normal es roll-forward.

-- ---------------------------------------------------------------------------
-- 1. Precondiciones
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'appointment_access_token'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260915190000_create_appointment_access_token.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'barber_review'
  ) THEN
    RAISE EXCEPTION 'barber_review ya existe: esta migración no puede aplicarse dos veces.';
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_owner') THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260811145252_harden_roles_and_definer_functions.sql '
      '(rol barberia_owner); esa migración debe estar aplicada antes.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Tabla
-- ---------------------------------------------------------------------------

CREATE TABLE barber_review (
  id             uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id  uuid        NOT NULL,
  barber_id      uuid        NOT NULL,
  appointment_id uuid        NOT NULL,
  rating         smallint    NOT NULL,
  comment        text,
  hidden_at      timestamptz,
  created_at     timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT barber_review_id_pk PRIMARY KEY (id),
  CONSTRAINT barber_review_barbershop_id_id_uk UNIQUE (barbershop_id, id),

  -- Una reseña por turno (DEC-128 §2).
  CONSTRAINT barber_review_appointment_id_uk UNIQUE (appointment_id),

  -- Sin CASCADE: ni el barbero ni el turno se borran.
  CONSTRAINT barber_review_barber_fk FOREIGN KEY (barbershop_id, barber_id)
    REFERENCES barber (barbershop_id, id) ON DELETE RESTRICT,
  CONSTRAINT barber_review_appointment_fk FOREIGN KEY (barbershop_id, appointment_id)
    REFERENCES appointment (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT barber_review_rating_ck CHECK (rating BETWEEN 1 AND 5),
  CONSTRAINT barber_review_comment_ck CHECK (
    comment IS NULL OR (comment = btrim(comment) AND char_length(comment) BETWEEN 1 AND 500)
  ),
  -- DDL-TMP-01: ocultar no puede preceder a la creación.
  CONSTRAINT barber_review_hidden_at_ck CHECK (hidden_at IS NULL OR hidden_at >= created_at)
);

COMMENT ON TABLE barber_review IS
  'Reseña interna del cliente sobre el barbero que lo atendió (DEC-128): una por turno completado, '
  'publicada de inmediato y ocultable de forma reversible. Propietario funcional: barbería. '
  'Clasificación: contenido publicado por el cliente (sin datos de contacto). Retención: mientras '
  'exista el barbero; la fila no se borra, se oculta.';

COMMENT ON COLUMN barber_review.rating IS 'Calificación entera de 1 a 5.';
COMMENT ON COLUMN barber_review.comment IS
  'Comentario opcional de hasta 500 caracteres, sin espacios en los extremos; NULL si el cliente no '
  'escribió nada. Nunca se registra en logs.';
COMMENT ON COLUMN barber_review.hidden_at IS
  'Momento en que el barbero o el propietario la ocultó; NULL = visible. Una oculta no cuenta en '
  'el promedio ni en el total públicos y el panel la conserva con su fecha (DEC-128 §4).';

-- Listado por barbero, más recientes primero (público y panel).
CREATE INDEX idx_barber_review_shop_barber_created
  ON barber_review (barbershop_id, barber_id, created_at DESC);

-- ---------------------------------------------------------------------------
-- 3. Disparador: turno completado y del mismo barbero
-- ---------------------------------------------------------------------------

CREATE FUNCTION barber_review_check_appointment() RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
  v_status    text;
  v_barber_id uuid;
BEGIN
  SELECT status, barber_id INTO v_status, v_barber_id
  FROM appointment
  WHERE barbershop_id = NEW.barbershop_id AND id = NEW.appointment_id;

  IF v_status IS DISTINCT FROM 'completed' THEN
    RAISE EXCEPTION 'barber_review: el turno debe estar completed (estado: %).', coalesce(v_status, 'inexistente')
      USING ERRCODE = 'check_violation', CONSTRAINT = 'barber_review_appointment_completed_ck';
  END IF;
  IF v_barber_id IS DISTINCT FROM NEW.barber_id THEN
    RAISE EXCEPTION 'barber_review: el barbero debe ser el del turno.'
      USING ERRCODE = 'check_violation', CONSTRAINT = 'barber_review_barber_matches_appointment_ck';
  END IF;
  RETURN NEW;
END
$$;

CREATE TRIGGER barber_review_check_appointment_trg
  BEFORE INSERT ON barber_review
  FOR EACH ROW EXECUTE FUNCTION barber_review_check_appointment();

-- ---------------------------------------------------------------------------
-- 4. RLS y privilegios
-- ---------------------------------------------------------------------------

ALTER TABLE barber_review ENABLE ROW LEVEL SECURITY;
ALTER TABLE barber_review FORCE  ROW LEVEL SECURITY;

CREATE POLICY barber_review_all_admin_policy ON barber_review
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY barber_review_select_tenant_policy ON barber_review
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY barber_review_insert_tenant_policy ON barber_review
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY barber_review_update_tenant_policy ON barber_review
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Sin DELETE; el único campo editable es `hidden_at` (nada de reescribir la
-- calificación ni el comentario de otra persona).
GRANT SELECT, INSERT ON TABLE barber_review TO barberia_app;
GRANT UPDATE (hidden_at) ON TABLE barber_review TO barberia_app;

-- ---------------------------------------------------------------------------
-- 5. Postcondiciones
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_class
    WHERE oid = 'barber_review'::regclass AND relrowsecurity AND relforcerowsecurity
  ) THEN
    RAISE EXCEPTION 'barber_review debe tener RLS habilitada y forzada.';
  END IF;
END
$$;
