-- Propósito
--   DEC-129 / HU-101: agrega `barbershop.google_place_id`, el identificador de
--   la ficha de Google (Place ID) de la barbería. Con él, el servidor consulta
--   a Google Places la calificación y las reseñas públicas del negocio y la
--   reserva pública las muestra.
--
-- Reglas y decisiones
--   RN-TEN-01, DEC-024 (esquema compartido con RLS), DEC-035/DEC-036 (estándar
--   de base de datos, Atlas, migración inmutable), DEC-082 (resolución pública
--   sin tenant), DEC-110 (marca y configuración de la barbería) y DEC-129
--   (reseñas de Google: una sola ficha por barbería).
--
-- Modelo
--   `google_place_id text NULL`: sin ficha configurada no hay reseñas de
--   Google y la reserva pública no muestra el apartado. Es un atributo de la
--   barbería (una ficha por barbería, DEC-129), así que vive en su fila: 3FN
--   intacta. `barbershop_google_place_id_ck` limita la forma a letras, dígitos,
--   `_` y `-` (la de un Place ID de Google) y 1–200 caracteres: el valor viaja
--   luego a una URL de Google, y la base nunca debe aceptar uno que no pueda
--   usarse como segmento de ruta.
--
-- Qué NO hace esta migración
--   No guarda calificación ni texto de reseñas: Google exige no persistirlos
--   más allá de una caché corta, que vive en memoria del servidor (6 horas).
--   No cambia RLS ni privilegios: `barbershop` ya tiene RLS forzada y
--   `barberia_app` ya puede actualizar la fila de su barbería. No agrega
--   índices (nunca se busca por este valor).
--
-- Clasificación de datos
--   El Place ID es un identificador público de una ficha pública de negocio,
--   no un dato personal ni un secreto. La clave de servidor de Google vive en
--   el ambiente, nunca en la base.
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
    WHERE table_schema = 'public' AND table_name = 'barbershop' AND column_name = 'offers_home_service'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20261101120000_add_home_service.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barbershop' AND column_name = 'google_place_id'
  ) THEN
    RAISE EXCEPTION 'barbershop.google_place_id ya existe: esta migración no puede aplicarse dos veces.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Columna y restricción
-- ---------------------------------------------------------------------------

ALTER TABLE barbershop
  ADD COLUMN google_place_id text NULL;

ALTER TABLE barbershop
  ADD CONSTRAINT barbershop_google_place_id_ck
  CHECK (google_place_id IS NULL OR google_place_id ~ '^[A-Za-z0-9_-]{1,200}$');

COMMENT ON COLUMN barbershop.google_place_id IS
  'Place ID de la ficha de Google de la barbería (DEC-129): letras, dígitos, _ y -, de 1 a 200 caracteres. '
  'NULL = sin ficha: la reserva pública no muestra reseñas de Google. Las reseñas nunca se persisten.';

-- ---------------------------------------------------------------------------
-- 3. Postcondiciones
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conrelid = 'barbershop'::regclass AND conname = 'barbershop_google_place_id_ck'
  ) THEN
    RAISE EXCEPTION 'barbershop_google_place_id_ck no quedó creada.';
  END IF;
END
$$;
