-- Propósito
--   Issue #322 (DEC-100): agrega a la tabla ya aplicada `barber` el vínculo
--   explícito y opcional con el usuario del área privada que es ese barbero
--   (`staff_user_id`). Resuelve la excepción que DEC-047 dejaba a «una
--   historia futura». Lo necesita la integración con Google Calendar
--   (DEC-099) para saber de qué barbero es la sesión autenticada; sin vínculo
--   el usuario simplemente no tiene barbero (p. ej. el dueño que no atiende).
--
-- Reglas y decisiones
--   RN-TEN-01, DEC-019, DEC-024 (esquema compartido con RLS), DEC-035/DEC-036
--   (estándar de base de datos, Atlas, migración inmutable), DEC-047, DEC-099
--   y DEC-100 (vínculo opcional, único en ambos sentidos).
--
-- Modelo
--   `staff_user_id uuid NULL` con clave foránea compuesta
--   (barbershop_id, staff_user_id) → staff_user (barbershop_id, id): un barbero
--   solo puede vincularse a un usuario de su misma barbería (estandar-base-datos
--   §6). La unicidad parcial por (barbershop_id, staff_user_id) cuando no es
--   nulo garantiza que cada usuario es como máximo un barbero; que cada barbero
--   tenga como máximo un usuario es inherente a la columna única.
--
-- Qué NO hace esta migración
--   No infiere vínculos por nombre, correo ni teléfono: todas las filas
--   existentes quedan en NULL y el vínculo solo lo establece el propio usuario
--   desde la aplicación. No agrega borrado ni desactivación de `barber`
--   (DEC-047 sigue vigente) y por eso la FK usa ON DELETE RESTRICT. No cambia
--   RLS ni privilegios: `barber` ya tiene RLS forzada y `barberia_app` ya tiene
--   UPDATE sobre la tabla, columnas nuevas incluidas.
--
-- Clasificación de datos
--   Identificador interno que relaciona dos filas del mismo tenant; no es dato
--   personal por sí mismo y la API nunca lo expone (DEC-047, DEC-100).
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
    WHERE table_schema = 'public' AND table_name = 'barbershop' AND column_name = 'panel_profile'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20261004120000_add_barbershop_panel_profile.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barber' AND column_name = 'staff_user_id'
  ) THEN
    RAISE EXCEPTION 'barber.staff_user_id ya existe: esta migración no puede aplicarse dos veces.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Columna, clave foránea compuesta y unicidad parcial
-- ---------------------------------------------------------------------------

ALTER TABLE barber
  ADD COLUMN staff_user_id uuid NULL;

ALTER TABLE barber
  ADD CONSTRAINT barber_barbershop_id_staff_user_id_fk
  FOREIGN KEY (barbershop_id, staff_user_id)
  REFERENCES staff_user (barbershop_id, id) ON DELETE RESTRICT;

-- Cada usuario es como máximo un barbero. Parcial: la mayoría de las filas
-- (sin vínculo) no entran al índice y NULL no cuenta como duplicado.
CREATE UNIQUE INDEX barber_barbershop_id_staff_user_id_uk
  ON barber (barbershop_id, staff_user_id)
  WHERE staff_user_id IS NOT NULL;

COMMENT ON COLUMN barber.staff_user_id IS
  'Usuario del área privada que es este barbero (issue #322, DEC-100). Opcional: NULL = el barbero '
  'no tiene usuario, o el usuario no atiende clientes. Lo establece, cambia o quita únicamente el '
  'propio usuario autenticado; nunca se infiere por nombre, correo ni teléfono, ni se asume igual a '
  'barber.id. Único por barbería en ambos sentidos. Clave foránea compuesta: solo puede apuntar a un '
  'usuario de la misma barbería. Nunca se expone por la API.';

COMMENT ON TABLE barber IS
  'Persona que presta el servicio de la barbería (HU-021, DEC-019). '
  'Propietario funcional: barbería. Retención: mientras exista la barbería; nunca se borra '
  'si tiene citas. Clasificación: datos personales (nombre). '
  'Alcance recortado a HU-021 (DEC-047, DDL-BIZ-02): sin borrado, desactivación ni orden manual; '
  'DEC-100 levantó parcialmente esa restricción solo para el vínculo opcional con staff_user.';
