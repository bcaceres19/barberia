-- Propósito
--   DEC-130 / HU-102: agrega a la tabla ya aplicada `barbershop` la ubicación
--   física opcional de la barbería: un interruptor explícito
--   (`has_physical_location`), la dirección en texto (`address`) y sus
--   coordenadas ya geocodificadas (`latitude`, `longitude`). La reserva
--   pública muestra un mapa con marcador solo cuando el interruptor está
--   activo y hay coordenadas.
--
-- Reglas y decisiones
--   RN-TEN-01, DEC-024 (esquema compartido con RLS), DEC-035/DEC-036
--   (estándar de base de datos, Atlas, migración inmutable), DEC-110 (marca
--   y configuración de la barbería) y DEC-130 (ubicación configurable).
--
-- Modelo
--   `has_physical_location boolean NOT NULL DEFAULT true`: ajuste explícito,
--   nunca derivado de si `address` está vacía (mismo criterio que
--   `panel_profile` en DEC-115). `address text NULL`: la dirección tal como la
--   escribió la barbería. `latitude`/`longitude numeric(9,6) NULL`: el
--   resultado de geocodificar UNA vez en el servidor al guardar; sin precisión
--   falsa (6 decimales ≈ 0,1 m). Las coordenadas son atributos de una barbería
--   y de una sola ubicación (varias sedes están fuera de alcance), así que
--   viven en la propia fila: 3FN intacta.
--
-- Qué NO hace esta migración
--   No rellena ubicaciones existentes: toda barbería queda con el interruptor
--   activo pero sin dirección ni coordenadas, y por eso no muestra mapa hasta
--   que alguien capture una dirección. No cambia RLS ni privilegios:
--   `barbershop` ya tiene RLS forzada y `barberia_app` ya tiene SELECT/UPDATE,
--   columnas nuevas incluidas. No agrega índices (no se filtra por ubicación).
--
-- Clasificación de datos
--   La dirección de un negocio es contenido público por diseño (DEC-130): se
--   muestra a cualquier visitante de la reserva pública. Si el titular es un
--   barbero independiente que trabaja desde su casa, la dirección puede ser un
--   domicilio particular: la pantalla de Configuración lo advierte y permite
--   apagar el interruptor sin borrar el dato. Ningún log registra la dirección
--   (RN-DAT-02).
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
    WHERE table_schema = 'public' AND table_name = 'barbershop' AND column_name = 'brand_accent'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20261003120000_add_barbershop_brand.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barbershop'
      AND column_name IN ('has_physical_location', 'address', 'latitude', 'longitude')
  ) THEN
    RAISE EXCEPTION 'La ubicación de barbershop ya existe: esta migración no puede aplicarse dos veces.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Columnas y restricciones
-- ---------------------------------------------------------------------------

ALTER TABLE barbershop
  ADD COLUMN has_physical_location boolean NOT NULL DEFAULT true,
  ADD COLUMN address   text          NULL,
  ADD COLUMN latitude  numeric(9, 6) NULL,
  ADD COLUMN longitude numeric(9, 6) NULL;

-- NULL = sin dirección. Una cadena vacía o de solo espacios es un error de
-- captura: la aplicación la convierte en NULL antes de guardar.
ALTER TABLE barbershop
  ADD CONSTRAINT barbershop_address_ck
  CHECK (address IS NULL OR (btrim(address) <> '' AND char_length(address) <= 200));

ALTER TABLE barbershop
  ADD CONSTRAINT barbershop_latitude_ck
  CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90);

ALTER TABLE barbershop
  ADD CONSTRAINT barbershop_longitude_ck
  CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180);

-- Las coordenadas viajan juntas y solo existen con una dirección: un punto sin
-- dirección ni una dirección con media coordenada son estados inválidos.
ALTER TABLE barbershop
  ADD CONSTRAINT barbershop_coordinates_ck
  CHECK (
    (latitude IS NULL) = (longitude IS NULL)
    AND (latitude IS NULL OR address IS NOT NULL)
  );

COMMENT ON COLUMN barbershop.has_physical_location IS
  'Interruptor explícito (DEC-130): true = la barbería tiene un local que mostrar en la reserva pública. '
  'false conserva address y coordenadas sin borrarlas. Nunca se deriva de si address está vacía.';
COMMENT ON COLUMN barbershop.address IS
  'Dirección del local en texto (DEC-130), de hasta 200 caracteres. NULL = sin dirección capturada. '
  'Contenido público: la reserva pública la muestra junto al mapa.';
COMMENT ON COLUMN barbershop.latitude IS
  'Latitud (grados decimales, -90..90) resultante de geocodificar address una vez en el servidor al guardar (DEC-130). '
  'Siempre junto con longitude y solo con address no nula (barbershop_coordinates_ck).';
COMMENT ON COLUMN barbershop.longitude IS
  'Longitud (grados decimales, -180..180); ver latitude.';

-- ---------------------------------------------------------------------------
-- 3. Postcondiciones
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF (SELECT count(*) FROM pg_constraint
       WHERE conrelid = 'barbershop'::regclass
         AND conname IN ('barbershop_address_ck', 'barbershop_latitude_ck',
                         'barbershop_longitude_ck', 'barbershop_coordinates_ck')) <> 4 THEN
    RAISE EXCEPTION 'Las cuatro restricciones de ubicación de barbershop no quedaron creadas.';
  END IF;
END
$$;
