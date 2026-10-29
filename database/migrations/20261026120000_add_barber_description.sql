-- Propósito
--   DEC-127: agrega a la tabla ya aplicada `barber` la descripción corta y
--   opcional del barbero (`description`), que la reserva pública muestra junto
--   a su nombre y fotografía para que el cliente elija con más información
--   que un nombre suelto.
--
-- Reglas y decisiones
--   RN-TEN-01, DEC-019, DEC-024 (esquema compartido con RLS), DEC-035/DEC-036
--   (estándar de base de datos, Atlas, migración inmutable), DEC-047 (la
--   ficha del barbero se amplía por decisión explícita), DEC-104 (fotografía
--   opcional) y DEC-127 (descripción y fotografía visibles en la reserva
--   pública).
--
-- Modelo
--   `description text NULL`: sin descripción es NULL, nunca una cadena vacía.
--   La restricción rechaza la cadena vacía o de solo espacios y limita el
--   largo a 280 caracteres (un texto de presentación, no una biografía). Vive
--   en la propia fila porque es un atributo de una sola columna, de cardinalidad
--   uno a uno con el barbero y sin ciclo de vida propio (a diferencia de la
--   fotografía, que es un binario con tabla aparte, DEC-104): 3FN intacta.
--
-- Qué NO hace esta migración
--   No rellena descripciones existentes: todas las filas quedan en NULL y solo
--   el propio equipo del panel las escribe. No cambia RLS ni privilegios:
--   `barber` ya tiene RLS forzada y `barberia_app` ya tiene SELECT/UPDATE sobre
--   la tabla, columnas nuevas incluidas. No agrega índices (no se filtra ni
--   ordena por descripción).
--
-- Clasificación de datos
--   Texto libre escrito por la barbería para presentar a su barbero. Es
--   contenido público por diseño (DEC-127): se muestra a cualquier visitante de
--   la reserva pública, así que no debe contener datos personales sensibles; el
--   panel lo advierte y ningún log lo registra (RN-DAT-02).
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE, ver
--   20260811145252_harden_roles_and_definer_functions.sql).
--
-- Plan de avance
--   Una corrección posterior se hace con otra migración, nunca editando
--   esta (DEC-036). No existe archivo `down`: el mecanismo normal es
--   roll-forward.

-- ---------------------------------------------------------------------------
-- 1. Precondiciones
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barber' AND column_name = 'staff_user_id'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20261009120000_add_barber_staff_user_link.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barber' AND column_name = 'description'
  ) THEN
    RAISE EXCEPTION 'barber.description ya existe: esta migración no puede aplicarse dos veces.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Columna y restricción
-- ---------------------------------------------------------------------------

ALTER TABLE barber
  ADD COLUMN description text NULL;

-- NULL = sin descripción. Una cadena vacía o de solo espacios no es "sin
-- descripción" sino un error de captura: la aplicación la convierte en NULL
-- antes de guardar y la base lo exige.
ALTER TABLE barber
  ADD CONSTRAINT barber_description_ck
  CHECK (description IS NULL OR (btrim(description) <> '' AND char_length(description) <= 280));

COMMENT ON COLUMN barber.description IS
  'Descripción corta y opcional del barbero (DEC-127), visible para cualquier visitante de la reserva '
  'pública. NULL = sin descripción. Texto libre de hasta 280 caracteres, sin datos personales '
  'sensibles; nunca una cadena vacía (barber_description_ck).';

-- ---------------------------------------------------------------------------
-- 3. Postcondiciones
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conrelid = 'barber'::regclass AND conname = 'barber_description_ck'
  ) THEN
    RAISE EXCEPTION 'barber_description_ck no quedó creada.';
  END IF;
END
$$;
