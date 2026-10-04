-- Propósito
--   Issue #294 (DEC-115): agrega a la tabla ya aplicada `barbershop` el perfil
--   del panel: `shop` (barbería con equipo, el panel completo de siempre) o
--   `solo` (barbero individual: el mismo panel sin la gestión de equipo ni el
--   selector de barbero). Es una preferencia de presentación de la barbería,
--   del mismo tipo que la marca y el vocabulario de DEC-110.
--
-- Reglas y decisiones
--   RN-TEN-01, RN-DAT-02, DEC-019 (uno o varios barberos, sin caso especial),
--   DEC-024 (esquema compartido con RLS), DEC-035 (estándar de base de datos),
--   DEC-036 (Atlas, migración inmutable), DEC-110 (marca por barbería),
--   DEC-115 (perfil del panel).
--
-- Qué NO hace esta migración
--   No crea un caso especial en el modelo: un barbero independiente sigue
--   siendo una barbería con un solo barbero (DEC-019, glosario). El perfil no
--   restringe ni valida cuántos barberos existen, no toca `barber` ni ninguna
--   regla de agenda y no concede permisos. No agrega ninguna tabla ni política
--   RLS: `barbershop` ya tiene RLS forzada y las políticas de selección y
--   actualización cubren la fila completa, columnas incluidas. No concede
--   ningún GRANT nuevo (un privilegio de tabla cubre las columnas agregadas
--   después).
--
-- Valores por defecto
--   `shop` reproduce exactamente el panel actual: ninguna barbería existente
--   cambia al aplicar la migración. `ADD COLUMN ... NOT NULL DEFAULT
--   <constante>` no reescribe la tabla en PostgreSQL 11 o posterior.
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
    WHERE table_schema = 'public' AND table_name = 'barbershop' AND column_name = 'panel_profile'
  ) THEN
    RAISE EXCEPTION 'barbershop.panel_profile ya existe: esta migración no puede aplicarse dos veces.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Columna nueva
-- ---------------------------------------------------------------------------

ALTER TABLE barbershop
  ADD COLUMN panel_profile text NOT NULL DEFAULT 'shop';

-- Lista cerrada de perfiles (DEC-115). Añadir uno exige una migración, porque
-- cada valor tiene que existir también en la navegación del cliente.
ALTER TABLE barbershop
  ADD CONSTRAINT barbershop_panel_profile_ck CHECK (panel_profile IN ('shop', 'solo'));

COMMENT ON COLUMN barbershop.panel_profile IS
  'Perfil del panel de la barbería (issue #294, DEC-115): shop (barbería con equipo, panel '
  'completo, valor inicial) o solo (barbero individual: sin gestión de equipo ni selector de '
  'barbero). Es presentación pura: no limita cuántos barberos existen ni cambia reglas de '
  'agenda, y un barbero independiente sigue siendo una barbería con un solo barbero (DEC-019). '
  'Propietario funcional: barbería. Clasificación: configuración.';
