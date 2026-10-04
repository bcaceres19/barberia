-- Propósito
--   Issue #292 (DEC-110): agrega a la tabla ya aplicada `barbershop` la marca
--   y el vocabulario que la barbería elige para su panel: el color de acento
--   (una paleta cerrada), la palabra con la que llama a su negocio
--   ("barbería", "salón", "estudio"...) y a su profesional ("barbero",
--   "estilista"...) con su plural y su género gramatical, para que la
--   interfaz concuerde ("este estilista" / "esta estilista").
--
-- Reglas y decisiones
--   RN-TEN-01, RN-DAT-02, DEC-024 (esquema compartido con RLS), DEC-035
--   (estándar de base de datos), DEC-036 (Atlas, migración inmutable),
--   DEC-110 (marca y vocabulario por barbería, paleta cerrada).
--
-- Qué NO hace esta migración
--   No toca `name`, `timezone`, el contacto (HU-020), `public_slug` (HU-090)
--   ni la política de reserva (HU-093). No admite un color libre: el acento
--   es una clave de una lista cerrada que el cliente traduce a una paleta
--   con contraste comprobado; ningún valor hexadecimal llega a la base. No
--   agrega ninguna tabla ni política RLS: `barbershop` ya tiene RLS forzada y
--   `barbershop_select_tenant_policy`/`barbershop_update_tenant_policy` cubren
--   la fila completa, columnas incluidas. No concede ningún GRANT nuevo
--   (`barbershop` ya tiene `GRANT SELECT, UPDATE ... TO barberia_app`; un
--   privilegio de tabla cubre las columnas agregadas después).
--
-- Valores por defecto
--   Reproducen exactamente la interfaz actual (latón, "barbería", "barbero"),
--   de modo que ninguna barbería existente cambia de aspecto ni de texto al
--   aplicar la migración. `ADD COLUMN ... NOT NULL DEFAULT <constante>` no
--   reescribe la tabla en PostgreSQL 11 o posterior.
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
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'barbershop'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260807170000_create_tenant_foundation.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barbershop' AND column_name = 'brand_accent'
  ) THEN
    RAISE EXCEPTION 'barbershop.brand_accent ya existe: esta migración no puede aplicarse dos veces.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Columnas nuevas
-- ---------------------------------------------------------------------------

ALTER TABLE barbershop
  ADD COLUMN brand_accent             text NOT NULL DEFAULT 'brass',
  ADD COLUMN business_term            text NOT NULL DEFAULT 'barbería',
  ADD COLUMN business_term_gender     text NOT NULL DEFAULT 'feminine',
  ADD COLUMN professional_term        text NOT NULL DEFAULT 'barbero',
  ADD COLUMN professional_term_plural text NOT NULL DEFAULT 'barberos',
  ADD COLUMN professional_term_gender text NOT NULL DEFAULT 'masculine';

-- Lista cerrada de acentos (DEC-110). Añadir uno exige una migración, porque
-- cada clave tiene que existir también en la paleta del cliente con su
-- contraste comprobado en los dos modos.
ALTER TABLE barbershop
  ADD CONSTRAINT barbershop_brand_accent_ck CHECK (
    brand_accent IN ('brass', 'emerald', 'sapphire', 'ruby', 'amethyst', 'copper')
  ),
  ADD CONSTRAINT barbershop_business_term_gender_ck CHECK (
    business_term_gender IN ('masculine', 'feminine')
  ),
  ADD CONSTRAINT barbershop_professional_term_gender_ck CHECK (
    professional_term_gender IN ('masculine', 'feminine')
  ),
  -- Última línea de defensa de la forma de un término: recortado, en
  -- minúsculas (la interfaz capitaliza donde corresponde), de 2 a 30
  -- caracteres y sin dígitos ni caracteres de control. El servicio aplica una
  -- regla más estricta (solo letras, espacios, guion y apóstrofo) con la
  -- clasificación Unicode de Go; la base no repite esa clasificación porque
  -- `[[:alpha:]]` depende de la configuración regional del clúster.
  ADD CONSTRAINT barbershop_business_term_ck CHECK (
    business_term = lower(btrim(business_term))
    AND char_length(business_term) BETWEEN 2 AND 30
    AND business_term !~ '[[:cntrl:][:digit:]]'
  ),
  ADD CONSTRAINT barbershop_professional_term_ck CHECK (
    professional_term = lower(btrim(professional_term))
    AND char_length(professional_term) BETWEEN 2 AND 30
    AND professional_term !~ '[[:cntrl:][:digit:]]'
  ),
  ADD CONSTRAINT barbershop_professional_term_plural_ck CHECK (
    professional_term_plural = lower(btrim(professional_term_plural))
    AND char_length(professional_term_plural) BETWEEN 2 AND 30
    AND professional_term_plural !~ '[[:cntrl:][:digit:]]'
  );

COMMENT ON COLUMN barbershop.brand_accent IS
  'Clave del color de acento del panel de la barbería (issue #292, DEC-110): brass, emerald, '
  'sapphire, ruby, amethyst o copper. Nunca un valor de color: el cliente la traduce a una paleta '
  'cerrada con contraste AA comprobado en modo Tinta y Marfil. Default brass (el latón NAVA). '
  'Propietario funcional: barbería. Clasificación: configuración.';

COMMENT ON COLUMN barbershop.business_term IS
  'Palabra con la que la barbería llama a su negocio en la interfaz (issue #292, DEC-110), en '
  'minúsculas, p. ej. barbería o salón. Default barbería. No es el nombre comercial (name). '
  'Propietario funcional: barbería. Clasificación: configuración.';

COMMENT ON COLUMN barbershop.business_term_gender IS
  'Género gramatical de business_term (masculine o feminine) para que los artículos de la '
  'interfaz concuerden (la barbería / el salón). Default feminine. Propietario funcional: '
  'barbería. Clasificación: configuración.';

COMMENT ON COLUMN barbershop.professional_term IS
  'Palabra en singular con la que la barbería llama a quien atiende los turnos (issue #292, '
  'DEC-110), en minúsculas, p. ej. barbero o estilista. Default barbero. No renombra la tabla '
  'barber ni el contrato de la API. Propietario funcional: barbería. Clasificación: configuración.';

COMMENT ON COLUMN barbershop.professional_term_plural IS
  'Plural de professional_term (barberos, estilistas). Se guarda en vez de derivarlo porque el '
  'plural español no es regular (estilista → estilistas, pintor → pintores). Default barberos. '
  'Propietario funcional: barbería. Clasificación: configuración.';

COMMENT ON COLUMN barbershop.professional_term_gender IS
  'Género gramatical de professional_term (masculine o feminine) para la concordancia de la '
  'interfaz (este barbero / esta barbera). Default masculine. Propietario funcional: barbería. '
  'Clasificación: configuración.';
