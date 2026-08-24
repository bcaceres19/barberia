-- Propósito
--   HU-022: crea la tabla `service` (catálogo básico de servicios de la
--   barbería): nombre, descripción opcional, duración planificada en
--   minutos y precio informativo en COP. Alta, lectura, listado y edición
--   parcial; sin ciclo de activación expuesto (HU-024) ni asignación a
--   barberos (HU-023, tabla `barber_service`, fuera de esta migración).
--
-- Reglas y decisiones
--   RN-SER-01 (duración configurable, tiempo planificado no garantizado),
--   RN-SER-02 (duración no limitada a 30/60/90), RN-SER-03 (un servicio no
--   se borra físicamente; el ciclo de desactivación llega con HU-024),
--   RN-SER-04 (un cambio de catálogo nunca modifica una cita existente),
--   RN-TEN-01 (ninguna barbería accede a datos de otra), RN-DAT-02
--   (registros técnicos sin datos personales), DEC-024 (esquema compartido
--   con RLS), DEC-035 (estándar de base de datos), DEC-036 (Atlas,
--   migración inmutable), DEC-040 (modelo de roles: política administrativa
--   a `barberia_owner`, no `barberia_migrator`), DEC-067 (moneda COP fija
--   sin columna editable por el cliente, precio estrictamente mayor que
--   cero, nombre único entre servicios activos de la misma barbería).
--
-- Relación con database/modelo-fisico-referencia.sql §B.3
--   Esta migración toma como base §B.3 (confirmada como decisión de
--   producto por DEC-067) con TRES diferencias deliberadas, documentadas
--   también en ese archivo:
--     1. service_price_amount_ck exige `price_amount > 0` (§B.3 tenía
--        `>= 0`): DEC-067 prohíbe servicios gratuitos.
--     2. service_price_currency_ck exige `price_currency = 'COP'` (§B.3
--        aceptaba cualquier código de tres letras mayúsculas): DEC-067 fija
--        la moneda, sin columna editable por el cliente. `price_currency`
--        sigue existiendo como columna (numeric + moneda ISO,
--        docs/05-backend/estandar-base-datos.md §5) para no romper esa
--        forma si una decisión futura la abre, pero ningún endpoint de
--        HU-022 la acepta como entrada: siempre 'COP', puesto por el
--        servidor.
--     3. idx_service_active_name se crea sobre `(barbershop_id, name)`,
--        comparación exacta (§B.3 usaba `lower(name)`, insensible a
--        mayúsculas): DEC-067 especifica literalmente
--        "`(barbershop_id, name) WHERE is_active`" como el índice a
--        implementar; esta migración sigue esa redacción exacta en vez de
--        la propuesta técnica de §B.3.
--
-- Qué NO hace esta migración
--   No agrega `barber_service` (HU-023) ni ninguna tabla de horario,
--   disponibilidad o cita (secciones posteriores de
--   modelo-fisico-referencia.sql). No agrega política ni GRANT de DELETE
--   (RN-SER-03): el borrado físico nunca se expone. No agrega ningún
--   endpoint ni columna que permita leer o escribir `is_active`/
--   `deactivated_at` desde HTTP: existen para que el esquema no necesite
--   otra migración cuando HU-024 los active, pero HU-022 no los expone.
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE,
--   20260811145252_harden_roles_and_definer_functions.sql). La política
--   administrativa se escribe `FOR ALL TO barberia_owner`: por membresía,
--   `barberia_migrator` la satisface sin que el objeto quede literalmente
--   owned por `barberia_owner` (mismo criterio documentado en
--   20260823130000_create_barber.sql).
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
    WHERE table_schema = 'public' AND table_name = 'service'
  ) THEN
    RAISE EXCEPTION 'service ya existe: esta migración no puede aplicarse dos veces.';
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_owner') THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260811145252_harden_roles_and_definer_functions.sql '
      '(rol barberia_owner); esa migración debe estar aplicada antes.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_proc p
    JOIN pg_namespace n ON n.oid = p.pronamespace
    WHERE n.nspname = 'public' AND p.proname = 'set_updated_at'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de la función set_updated_at() '
      '(20260807170000_create_tenant_foundation.sql); esa migración debe estar aplicada antes.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Tabla `service`
-- ---------------------------------------------------------------------------

CREATE TABLE service (
  id               uuid           NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id    uuid           NOT NULL,
  name             text           NOT NULL,
  description      text,
  duration_minutes integer        NOT NULL,
  price_amount     numeric(12, 2) NOT NULL,
  price_currency   char(3)        NOT NULL DEFAULT 'COP',
  is_active        boolean        NOT NULL DEFAULT true,
  deactivated_at   timestamptz,
  created_at       timestamptz    NOT NULL DEFAULT now(),
  updated_at       timestamptz    NOT NULL DEFAULT now(),

  CONSTRAINT service_id_pk               PRIMARY KEY (id),
  CONSTRAINT service_barbershop_id_fk    FOREIGN KEY (barbershop_id)
                                         REFERENCES barbershop (id) ON DELETE RESTRICT,
  CONSTRAINT service_barbershop_id_id_uk UNIQUE (barbershop_id, id),

  CONSTRAINT service_name_ck        CHECK (btrim(name) <> '' AND char_length(name) <= 120),
  CONSTRAINT service_description_ck CHECK (description IS NULL OR char_length(description) <= 500),

  -- RN-SER-01: minutos enteros positivos. El tope evita una duración absurda
  -- por error de digitación sin cerrar la lista de valores (RN-SER-02).
  CONSTRAINT service_duration_minutes_ck CHECK (duration_minutes BETWEEN 1 AND 1440),

  -- `numeric`, nunca coma flotante (estandar-base-datos.md §5). DEC-067:
  -- estrictamente mayor que cero, nunca servicios gratuitos.
  CONSTRAINT service_price_amount_ck   CHECK (price_amount > 0),
  -- DEC-067: moneda COP fija, sin columna editable por el cliente (a
  -- diferencia de modelo-fisico-referencia.sql §B.3, que aceptaba
  -- cualquier código ISO de tres letras mayúsculas).
  CONSTRAINT service_price_currency_ck CHECK (price_currency = 'COP'),

  -- RN-SER-03: un servicio no se elimina; se desactiva y conserva su
  -- historial. HU-022 no expone ningún endpoint que cambie is_active (nace
  -- en true); la columna y este CHECK preparan HU-024 sin exponer esa
  -- transición todavía.
  CONSTRAINT service_deactivated_at_ck CHECK (
    (is_active AND deactivated_at IS NULL) OR (NOT is_active AND deactivated_at IS NOT NULL)
  )
);

COMMENT ON TABLE service IS
  'Catálogo de servicios de la barbería (HU-022). Propietario funcional: barbería. '
  'Retención: permanente; nunca se borra físicamente (RN-SER-03). Clasificación: negocio. '
  'Los cambios de este catálogo NO se propagan a citas existentes (RN-SER-04, DEC-004): '
  'una futura tabla de citas guardará su propio snapshot de nombre, duración y precio.';

COMMENT ON COLUMN service.name IS
  'Nombre visible del servicio, recortado por el servicio Go antes de persistir (nunca vacío '
  'ni solo espacios, máximo 120 caracteres). Único entre servicios ACTIVOS de la misma '
  'barbería (idx_service_active_name, DEC-067): un servicio desactivado libera su nombre.';

COMMENT ON COLUMN service.price_amount IS
  'Precio informativo en price_currency, estrictamente mayor que cero (DEC-067: sin '
  'servicios gratuitos). No garantiza el precio final de una cita: RN-SER-04/DEC-004 exigen '
  'que un cambio posterior no reescriba citas ya creadas.';

COMMENT ON COLUMN service.price_currency IS
  'Fija en COP para todo el MVP (DEC-067): ningún endpoint de HU-022 acepta este valor como '
  'entrada del cliente. Existe como columna (no una constante de aplicación) para que una '
  'decisión de producto futura pueda abrirla sin otra migración estructural.';

COMMENT ON COLUMN service.is_active IS
  'Nace en true; HU-022 no expone ningún endpoint que la cambie. Prepara el ciclo de '
  'activación de HU-024 (RN-SER-03) sin exponer esa transición todavía.';

CREATE TRIGGER service_set_updated_at
  BEFORE UPDATE ON service
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- 3. Seguridad a nivel de fila
-- ---------------------------------------------------------------------------

ALTER TABLE service ENABLE ROW LEVEL SECURITY;
ALTER TABLE service FORCE  ROW LEVEL SECURITY;

-- Rol administrativo: FORCE ROW LEVEL SECURITY también alcanza al
-- propietario, así que barberia_owner necesita una política explícita para
-- cargar datos de referencia, semillas y correcciones auditadas
-- (estandar-base-datos.md §9.5, mismo patrón que barber/barbershop).
CREATE POLICY service_all_admin_policy ON service
  FOR ALL TO barberia_owner
  USING (true) WITH CHECK (true);

CREATE POLICY service_select_tenant_policy ON service
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY service_insert_tenant_policy ON service
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY service_update_tenant_policy ON service
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Sin política ni GRANT de DELETE: RN-SER-03 prohíbe el borrado físico.

-- ---------------------------------------------------------------------------
-- 4. Índices
-- ---------------------------------------------------------------------------

-- Sirve tanto el filtro de RLS/WHERE barbershop_id = ... como el orden
-- estable (created_at, id) que la paginación por cursor de CA-022-01
-- necesita para recorrer páginas contiguas sin duplicados ni omisiones.
CREATE INDEX idx_service_shop_created_id ON service (barbershop_id, created_at, id);

-- DEC-067: nombre único entre servicios ACTIVOS de la misma barbería,
-- comparación exacta (no lower(name), a diferencia de la propuesta técnica
-- de modelo-fisico-referencia.sql §B.3: ver la nota al inicio del archivo).
-- Un servicio desactivado libera su nombre para que HU-024 permita
-- reutilizarlo sin que esta historia exponga esa transición.
CREATE UNIQUE INDEX idx_service_active_name
  ON service (barbershop_id, name)
  WHERE is_active;

-- ---------------------------------------------------------------------------
-- 5. Privilegios del rol de aplicación
-- ---------------------------------------------------------------------------

GRANT SELECT, INSERT, UPDATE ON TABLE service TO barberia_app;
