-- Propósito
--   Issue #365 (DEC-134): crea los roles del usuario del área privada
--   (`empresa`, `empleado`, `independiente`), que deciden qué tipo de panel
--   ve cada persona. Sustituye en la práctica a `barbershop.panel_profile`
--   (DEC-115), que era un ajuste de la barbería editable desde Configuración;
--   la columna sigue existiendo y se retira en una migración posterior.
--
-- Reglas y decisiones
--   RN-TEN-01, DEC-019, DEC-024 (esquema compartido con RLS), DEC-035/DEC-036
--   (estándar de base de datos, Atlas, migración inmutable), DEC-040, DEC-115
--   y DEC-134.
--
-- Modelo
--   `staff_user_role (barbershop_id, staff_user_id, role)`. La clave primaria es
--   (staff_user_id, role): un usuario puede tener varios roles a la vez (un
--   barbero empleado que además tiene su propia cuenta es `empleado` e
--   `independiente`) y nunca el mismo dos veces. La clave foránea compuesta
--   (barbershop_id, staff_user_id) → staff_user (barbershop_id, id) impide
--   asignar un rol a un usuario de otra barbería aun con SQL directo bajo el
--   rol de aplicación (estandar-base-datos §6). `role` es texto con CHECK sobre
--   la lista cerrada, como `panel_profile`.
--
-- Datos existentes
--   Cada usuario de una barbería con `panel_profile = 'solo'` recibe el rol
--   `independiente`; cada usuario de una barbería con `'shop'`, el rol
--   `empresa`. Es la correspondencia que fijó el propietario en DEC-134. Un
--   usuario creado después sin ningún rol se muestra con el panel completo
--   (`shop`), igual que antes.
--
-- Privilegios y RLS
--   RLS forzada. La aplicación solo LEE los roles de su tenant; no hay política
--   ni GRANT de INSERT, UPDATE ni DELETE porque ninguna pantalla los cambia: se
--   asignan en la base de datos con el rol administrativo (DEC-134).
--
-- Qué NO hace esta migración
--   No retira `barbershop.panel_profile`, no concede ni retira permisos de API
--   (los roles solo deciden la presentación) y no infiere roles por nombre o
--   correo.
--
-- Clasificación de datos
--   Relaciona un identificador interno con un rol de presentación; no es dato
--   personal por sí mismo y la API nunca expone el identificador (RN-DAT-02).
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
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barbershop' AND column_name = 'panel_profile'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20261004120000_add_barbershop_panel_profile.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'staff_user_role'
  ) THEN
    RAISE EXCEPTION 'staff_user_role ya existe: esta migración no puede aplicarse dos veces.';
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

CREATE TABLE staff_user_role (
  barbershop_id  uuid        NOT NULL,
  staff_user_id  uuid        NOT NULL,
  role           text        NOT NULL,
  created_at     timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT staff_user_role_pk PRIMARY KEY (staff_user_id, role),

  -- Sin CASCADE: `staff_user` no se borra.
  CONSTRAINT staff_user_role_staff_user_fk FOREIGN KEY (barbershop_id, staff_user_id)
    REFERENCES staff_user (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT staff_user_role_role_ck CHECK (role IN ('empresa', 'empleado', 'independiente'))
);

COMMENT ON TABLE staff_user_role IS
  'Roles de un usuario del área privada (DEC-134): empresa, empleado o independiente; un usuario '
  'puede tener varios (empleado e independiente a la vez). Deciden el tipo de panel que ve, solo '
  'como presentación. Se asignan en la base de datos, no desde la aplicación. Propietario '
  'funcional: barbería. Clasificación: configuración. Retención: mientras exista el usuario.';

COMMENT ON COLUMN staff_user_role.role IS
  'empresa: la barbería como negocio con equipo (panel completo). empleado: barbero que trabaja '
  'para la barbería (panel individual limitado). independiente: barbero que trabaja por su cuenta '
  '(panel individual completo). Prioridad al elegir el panel: empresa, independiente, empleado.';

ALTER TABLE staff_user_role ENABLE ROW LEVEL SECURITY;
ALTER TABLE staff_user_role FORCE  ROW LEVEL SECURITY;

-- Rol administrativo: FORCE ROW LEVEL SECURITY también alcanza al propietario
-- (estandar-base-datos.md §9.5, mismo patrón que barber/barber_photo).
CREATE POLICY staff_user_role_all_admin_policy ON staff_user_role
  FOR ALL TO barberia_owner
  USING (true) WITH CHECK (true);

CREATE POLICY staff_user_role_select_tenant_policy ON staff_user_role
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Solo lectura para la aplicación: los roles no se editan desde ninguna pantalla.
GRANT SELECT ON TABLE staff_user_role TO barberia_app;

-- ---------------------------------------------------------------------------
-- 3. Datos existentes: panel_profile → roles
-- ---------------------------------------------------------------------------

INSERT INTO staff_user_role (barbershop_id, staff_user_id, role)
SELECT s.barbershop_id,
       s.id,
       CASE b.panel_profile WHEN 'solo' THEN 'independiente' ELSE 'empresa' END
FROM staff_user s
JOIN barbershop b ON b.id = s.barbershop_id;

-- ---------------------------------------------------------------------------
-- 4. Postcondiciones
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM staff_user s
    WHERE NOT EXISTS (SELECT 1 FROM staff_user_role r WHERE r.staff_user_id = s.id)
  ) THEN
    RAISE EXCEPTION 'Quedaron usuarios sin rol tras la migración de panel_profile.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_class
    WHERE oid = 'staff_user_role'::regclass AND relrowsecurity AND relforcerowsecurity
  ) THEN
    RAISE EXCEPTION 'staff_user_role debe tener RLS habilitada y forzada.';
  END IF;
END
$$;
