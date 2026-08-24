-- Propósito
--   HU-021: crea la tabla mínima `barber`, distinta de `barbershop` y de
--   `staff_user`, con registro, listado, lectura individual y renombrado.
--   Copia B.2 de modelo-fisico-referencia.sql tal cual (DEC-047 ya la
--   recortó al alcance de esta historia): sin borrado, desactivación, orden
--   manual ni vínculo con staff_user.
--
-- Reglas y decisiones
--   RN-TEN-01 (ninguna barbería accede a datos de otra), RN-DAT-02
--   (registros técnicos sin datos personales), DEC-019 (modelo
--   multi-barbero: una barbería unipersonal y una de varios barberos usan
--   la misma fila `barber`, sin caso especial), DEC-024 (esquema compartido
--   con RLS), DEC-035 (estándar de base de datos), DEC-036 (Atlas,
--   migración inmutable), DEC-040 (modelo de roles: política administrativa
--   a `barberia_owner`, no `barberia_migrator`), DEC-047 (alcance de
--   `barber` recortado a esta historia).
--
-- Qué NO hace esta migración
--   No agrega `service`, `barber_service`, horario, disponibilidad ni
--   agenda (secciones B.3+ de modelo-fisico-referencia.sql pertenecen a
--   historias futuras). No agrega borrado, `active`, `deleted_at`,
--   `sort_order` ni columna que vincule con `staff_user` (DEC-047,
--   CA-021-07). No agrega unicidad de `full_name`: dos barberos de la misma
--   barbería pueden compartir nombre.
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE,
--   20260811145252_harden_roles_and_definer_functions.sql). La política
--   administrativa se escribe `FOR ALL TO barberia_owner`: por membresía,
--   `barberia_migrator` la satisface sin que el objeto quede literalmente
--   owned por `barberia_owner` (mismo criterio documentado en esa
--   migración, "Plan de avance").
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
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'barber'
  ) THEN
    RAISE EXCEPTION 'barber ya existe: esta migración no puede aplicarse dos veces.';
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_owner') THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260811145252_harden_roles_and_definer_functions.sql '
      '(rol barberia_owner); esa migración debe estar aplicada antes.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Tabla `barber`
-- ---------------------------------------------------------------------------

CREATE TABLE barber (
  id            uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id uuid        NOT NULL,
  full_name     text        NOT NULL,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT barber_id_pk               PRIMARY KEY (id),
  CONSTRAINT barber_barbershop_id_fk    FOREIGN KEY (barbershop_id)
                                        REFERENCES barbershop (id) ON DELETE RESTRICT,
  CONSTRAINT barber_barbershop_id_id_uk UNIQUE (barbershop_id, id),

  CONSTRAINT barber_full_name_ck CHECK (btrim(full_name) <> '' AND char_length(full_name) <= 120)
);

COMMENT ON TABLE barber IS
  'Persona que presta el servicio de la barbería (HU-021, DEC-019). '
  'Propietario funcional: barbería. Retención: mientras exista la barbería; nunca se borra '
  'si tiene citas (fuera de alcance de esta migración: aún no existen citas). '
  'Clasificación: datos personales (nombre). '
  'Alcance recortado a HU-021 (DEC-047, DDL-BIZ-02): sin borrado, desactivación, orden manual '
  'ni vínculo con staff_user; una historia futura los aprueba explícitamente si llegan a hacer falta.';

COMMENT ON COLUMN barber.full_name IS
  'Nombre visible del barbero, recortado por el servicio antes de persistir (nunca vacío ni '
  'solo espacios, máximo 120 caracteres). Sin unicidad: dos barberos de la misma barbería '
  'pueden compartir nombre (HU-021 no impone esa regla).';

CREATE TRIGGER barber_set_updated_at
  BEFORE UPDATE ON barber
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- 3. Seguridad a nivel de fila
-- ---------------------------------------------------------------------------

ALTER TABLE barber ENABLE ROW LEVEL SECURITY;
ALTER TABLE barber FORCE  ROW LEVEL SECURITY;

-- Rol administrativo: FORCE ROW LEVEL SECURITY también alcanza al
-- propietario, así que barberia_owner necesita una política explícita para
-- cargar datos de referencia, semillas y correcciones auditadas
-- (estandar-base-datos.md §9.5, mismo patrón que barbershop/staff_user).
CREATE POLICY barber_all_admin_policy ON barber
  FOR ALL TO barberia_owner
  USING (true) WITH CHECK (true);

CREATE POLICY barber_select_tenant_policy ON barber
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY barber_insert_tenant_policy ON barber
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY barber_update_tenant_policy ON barber
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Sin política ni GRANT de DELETE: HU-021 lo excluye explícitamente (DEC-047, CA-021-07).

-- ---------------------------------------------------------------------------
-- 4. Índices
-- ---------------------------------------------------------------------------

-- Sirve tanto el filtro de RLS/WHERE barbershop_id = ... como el orden
-- estable (created_at, id) que la paginación por cursor de CA-021-02
-- necesita para recorrer páginas contiguas sin duplicados ni omisiones.
CREATE INDEX idx_barber_shop_created_id ON barber (barbershop_id, created_at, id);

-- ---------------------------------------------------------------------------
-- 5. Privilegios del rol de aplicación
-- ---------------------------------------------------------------------------

GRANT SELECT, INSERT, UPDATE ON TABLE barber TO barberia_app;
