-- Propósito
--   DEC-131 / HU-103: agrega el servicio a domicilio. `barbershop` gana un
--   interruptor general (`offers_home_service`) y `appointment` guarda la
--   modalidad del turno y, cuando es a domicilio, la dirección del cliente ya
--   geocodificada (`service_address`, `service_latitude`, `service_longitude`).
--
-- Reglas y decisiones
--   RN-TEN-01, RN-DAT-02, DEC-024 (esquema compartido con RLS), DEC-035/DEC-036
--   (estándar de base de datos, Atlas, migración inmutable), DEC-110 (marca y
--   configuración de la barbería) y DEC-131 (domicilio: interruptor general).
--
-- Modelo
--   `barbershop.offers_home_service boolean NOT NULL DEFAULT false`: ajuste
--   explícito de la barbería, apagado por defecto. Apagarlo no toca los turnos
--   ya creados con domicilio.
--   `appointment.is_home_service boolean NOT NULL DEFAULT false` y los tres
--   datos de dirección `NULL`. `appointment_home_service_ck` exige los tres
--   juntos cuando `is_home_service` es verdadero y los tres nulos cuando es
--   falso: un turno a domicilio nunca existe sin dirección geocodificada
--   completa y un turno en el local nunca arrastra datos de dirección, aun
--   ante un error de la aplicación. La dirección es un atributo del turno (la
--   del cliente ese día, no del cliente), así que vive en la propia fila: 3FN
--   intacta.
--
-- Qué NO hace esta migración
--   No cambia datos existentes: toda barbería queda sin domicilio y todo turno
--   existente queda en el local. No cambia RLS ni privilegios: ambas tablas ya
--   tienen RLS forzada y `barberia_app` ya tiene sus privilegios de columna
--   o de tabla (columnas nuevas incluidas). No agrega índices (no se filtra
--   por domicilio). No afecta `appointment_barber_interval_excl`.
--
-- Clasificación de datos
--   `service_address`, `service_latitude` y `service_longitude` son dato
--   personal del cliente (RN-DAT-02): solo el panel privado de esa barbería
--   los lee, nunca la reserva pública, las reseñas ni ningún registro. La
--   anonimización de un cliente (RN-DAT-03) debe limpiarlos juntos con
--   `is_home_service = false` para respetar el CHECK.
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
    WHERE table_schema = 'public' AND table_name = 'barbershop' AND column_name = 'longitude'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20261030120000_add_barbershop_location.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public'
      AND ((table_name = 'barbershop' AND column_name = 'offers_home_service')
        OR (table_name = 'appointment' AND column_name IN
             ('is_home_service', 'service_address', 'service_latitude', 'service_longitude')))
  ) THEN
    RAISE EXCEPTION 'El servicio a domicilio ya existe: esta migración no puede aplicarse dos veces.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Columnas y restricciones
-- ---------------------------------------------------------------------------

ALTER TABLE barbershop
  ADD COLUMN offers_home_service boolean NOT NULL DEFAULT false;

ALTER TABLE appointment
  ADD COLUMN is_home_service   boolean       NOT NULL DEFAULT false,
  ADD COLUMN service_address   text          NULL,
  ADD COLUMN service_latitude  numeric(9, 6) NULL,
  ADD COLUMN service_longitude numeric(9, 6) NULL;

ALTER TABLE appointment
  ADD CONSTRAINT appointment_home_service_ck
  CHECK (
    (is_home_service
       AND service_address IS NOT NULL
       AND btrim(service_address) <> ''
       AND char_length(service_address) <= 200
       AND service_latitude  IS NOT NULL AND service_latitude  BETWEEN -90 AND 90
       AND service_longitude IS NOT NULL AND service_longitude BETWEEN -180 AND 180)
    OR
    (NOT is_home_service
       AND service_address IS NULL
       AND service_latitude IS NULL
       AND service_longitude IS NULL)
  );

COMMENT ON COLUMN barbershop.offers_home_service IS
  'Interruptor general del servicio a domicilio (DEC-131). true permite elegir «a domicilio» al reservar '
  'cualquier servicio. Apagarlo no afecta turnos ya creados con domicilio.';
COMMENT ON COLUMN appointment.is_home_service IS
  'Modalidad del turno (DEC-131): true = a domicilio, false = en el local. Con true, appointment_home_service_ck '
  'exige la dirección completa; con false, exige los tres datos de dirección nulos.';
COMMENT ON COLUMN appointment.service_address IS
  'Dirección del cliente para un turno a domicilio (DEC-131), de hasta 200 caracteres. Dato personal (RN-DAT-02): '
  'solo la lee el panel privado de la barbería dueña del turno; nunca se registra en logs.';
COMMENT ON COLUMN appointment.service_latitude IS
  'Latitud (grados decimales, -90..90) de la dirección elegida con el autocompletado de Google Places. '
  'Siempre junto con service_address y service_longitude (appointment_home_service_ck).';
COMMENT ON COLUMN appointment.service_longitude IS
  'Longitud (grados decimales, -180..180); ver service_latitude.';

-- ---------------------------------------------------------------------------
-- 3. Postcondiciones
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conrelid = 'appointment'::regclass AND conname = 'appointment_home_service_ck'
  ) THEN
    RAISE EXCEPTION 'appointment_home_service_ck no quedó creada.';
  END IF;
END
$$;
