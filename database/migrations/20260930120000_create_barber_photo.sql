-- Propósito
--   DEC-100: crea `barber_photo`, la fotografía opcional de cada barbero de la
--   barbería (retrato para la lista de equipo y el selector de barbero). Una
--   fila por barbero como máximo; sin fila = el barbero usa el monograma.
--   La imagen vive en PostgreSQL (`bytea`, tope de 512 KiB) y NO en `barber`:
--   la lista de barberos nunca lee estos bytes (solo `updated_at`, vía LEFT
--   JOIN) y una fotografía es un dato propio con ciclo de vida propio
--   (subir, reemplazar, quitar), no un atributo del nombre.
--
-- Reglas y decisiones
--   RN-TEN-01 (ninguna barbería accede a datos de otra), DEC-019/DEC-024
--   (esquema compartido con RLS), DEC-035/DEC-036 (estándar de base de datos,
--   Atlas, migración inmutable), DEC-040 (modelo de roles), DEC-047 (`barber`
--   se mantiene sin desactivación ni borrado), DEC-100 (fotografía de barbero
--   con almacenamiento en PostgreSQL, contrato binario aprobado por el
--   propietario).
--
-- Decisión de almacenamiento (DEC-100)
--   No existe hoy almacenamiento de objetos en la infraestructura del
--   proyecto y añadirlo (bucket, credenciales, URL firmadas, ciclo de vida)
--   sería una dependencia nueva sin necesidad demostrada (AGENTS.md,
--   "Calidad"). El cliente recorta y reduce la imagen a un cuadrado de
--   512 px antes de enviarla y el servidor la valida (formato real, tamaño,
--   dimensiones): una fotografía pesa decenas de KiB, así que `bytea` con un
--   CHECK de 512 KiB es proporcionado. Se revisará si el volumen real lo
--   justifica.
--
-- Qué NO hace esta migración
--   No modifica `barber` (sigue siendo dueño exclusivo del nombre). No agrega
--   `ON DELETE CASCADE`: `barber` no expone DELETE (CA-021-07), así que la
--   FK compuesta usa `ON DELETE RESTRICT` para documentar esa imposibilidad.
--   No guarda el nombre original del archivo, EXIF ni ningún metadato del
--   dispositivo: el cliente re-codifica la imagen y el servidor solo persiste
--   el tipo de contenido y los bytes ya validados.
--
-- Clasificación de datos
--   Datos personales (imagen de una persona). Retención: mientras el barbero
--   conserve la fotografía; quitarla borra la fila (a diferencia de `barber`,
--   esta tabla SÍ expone DELETE al rol de aplicación: retirar la imagen de
--   una persona debe eliminarla, no solo ocultarla).
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE,
--   20260811145252_harden_roles_and_definer_functions.sql). La política
--   administrativa se escribe `FOR ALL TO barberia_owner`: mismo criterio que
--   20260823130000_create_barber.sql y 20260824150000_create_barber_service.sql.
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
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'barber'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260823130000_create_barber.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'barber_photo'
  ) THEN
    RAISE EXCEPTION 'barber_photo ya existe: esta migración no puede aplicarse dos veces.';
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_owner') THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260811145252_harden_roles_and_definer_functions.sql '
      '(rol barberia_owner); esa migración debe estar aplicada antes.';
  END IF;

  -- La FK compuesta por tenant necesita el mismo `UNIQUE (barbershop_id, id)`
  -- que `barber` ya declaró en HU-021.
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'barber_barbershop_id_id_uk'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de barber_barbershop_id_id_uk '
      '(20260823130000_create_barber.sql); esa restricción debe existir antes.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Tabla `barber_photo`
-- ---------------------------------------------------------------------------

CREATE TABLE barber_photo (
  barbershop_id uuid        NOT NULL,
  barber_id     uuid        NOT NULL,
  content_type  text        NOT NULL,
  image         bytea       NOT NULL,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),

  -- Una fotografía por barbero: la propia PK impide una segunda fila y sirve
  -- de índice para toda consulta (siempre por barbero dentro de su tenant).
  CONSTRAINT barber_photo_pk PRIMARY KEY (barbershop_id, barber_id),

  -- FK compuesta por barbershop_id: impide asociar una imagen a un barbero
  -- de otra barbería AUN con SQL directo bajo el rol de aplicación (RN-TEN-01).
  CONSTRAINT barber_photo_barber_fk FOREIGN KEY (barbershop_id, barber_id)
                                    REFERENCES barber (barbershop_id, id)
                                    ON DELETE RESTRICT,

  -- El servidor solo persiste imágenes cuyo formato real ya verificó; el
  -- CHECK es la última defensa, no la validación de negocio.
  CONSTRAINT barber_photo_content_type_ck CHECK (content_type IN ('image/jpeg', 'image/png')),
  CONSTRAINT barber_photo_image_size_ck   CHECK (octet_length(image) BETWEEN 1 AND 524288)
);

COMMENT ON TABLE barber_photo IS
  'Fotografía opcional de un barbero (DEC-100), una fila como máximo por barbero. '
  'Propietario funcional: módulo staff. Clasificación: datos personales (imagen de una '
  'persona). Retención: mientras el barbero la conserve; quitarla borra la fila. Sin '
  'nombre de archivo, EXIF ni metadatos del dispositivo: solo content_type y bytes ya '
  'validados y re-codificados por el cliente. Vive aparte de `barber` para que el listado '
  'del equipo no lea los bytes.';

COMMENT ON COLUMN barber_photo.image IS
  'Imagen JPEG o PNG ya validada por el servidor (formato real, dimensiones acotadas). '
  'Tope de 512 KiB (barber_photo_image_size_ck).';

COMMENT ON COLUMN barber_photo.updated_at IS
  'Instante de la última fotografía. La API lo expone como `photoUpdatedAt` y el cliente '
  'lo usa como versión para invalidar la caché de la imagen y como ETag.';

CREATE TRIGGER barber_photo_set_updated_at
  BEFORE UPDATE ON barber_photo
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- 3. Seguridad a nivel de fila
-- ---------------------------------------------------------------------------

ALTER TABLE barber_photo ENABLE ROW LEVEL SECURITY;
ALTER TABLE barber_photo FORCE  ROW LEVEL SECURITY;

-- Rol administrativo: FORCE ROW LEVEL SECURITY también alcanza al
-- propietario, así que barberia_owner necesita una política explícita
-- (estandar-base-datos.md §9.5, mismo patrón que barber/barber_service).
CREATE POLICY barber_photo_all_admin_policy ON barber_photo
  FOR ALL TO barberia_owner
  USING (true) WITH CHECK (true);

CREATE POLICY barber_photo_select_tenant_policy ON barber_photo
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY barber_photo_insert_tenant_policy ON barber_photo
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY barber_photo_update_tenant_policy ON barber_photo
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY barber_photo_delete_tenant_policy ON barber_photo
  FOR DELETE TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- ---------------------------------------------------------------------------
-- 4. Privilegios del rol de aplicación
-- ---------------------------------------------------------------------------

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE barber_photo TO barberia_app;
