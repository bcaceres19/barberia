-- Propósito
--   HU-093: agrega la configuración pública de reserva y cancelación de la
--   barbería (anticipación mínima, ventana máxima, rejilla de horarios y
--   política de cancelación tardía) a la tabla ya aplicada `barbershop`.
--   `name`, `timezone` y el contacto ya existen desde
--   20260807170000_create_tenant_foundation.sql /
--   20260823120000_add_barbershop_contact_info.sql; esta migración no los
--   toca. `public_slug` (HU-090) tampoco.
--
-- Reglas y decisiones
--   RN-DIS-04 (anticipación mínima y ventana máxima), RN-DIS-06 (paso de
--   rejilla), RN-CAN-01 (plazo de cancelación del cliente), RN-CAN-02
--   (política de cancelación fuera de plazo configurable), RN-TEN-01,
--   RN-DAT-02, DEC-005/DEC-006/DEC-010/DEC-018 (valores iniciales),
--   DEC-083 (rangos permitidos y default de la política, resuelve
--   DP-PUB-02), DEC-024 (esquema compartido con RLS), DEC-035 (estándar de
--   base de datos), DEC-036 (Atlas, migración inmutable).
--
-- Relación con database/modelo-fisico-referencia.sql §E.3
--   La sección E.3 ya proponía columnas tipadas en `barbershop` (no una
--   tabla clave/valor ni `jsonb`) para estos mismos parámetros, con nombres
--   parecidos (`min_lead_minutes`, `max_booking_window_days`,
--   `slot_step_minutes`) que esta migración reutiliza donde el concepto
--   coincide exactamente. Los RANGOS y la forma de la política de
--   cancelación NO se copian de esa sección como norma (el prompt de
--   HU-093 lo prohíbe explícitamente): DEC-083 fija rangos más estrechos
--   (`min_lead_minutes` tope 1440, no 10080; `max_booking_window_days`
--   tope 90, no 365; `slot_step_minutes` restringido a un conjunto discreto
--   {5,10,15,20,30,60}, no un rango continuo 5-120) y sustituye el enum
--   `late_cancellation_policy` por dos booleanos independientes
--   (`late_cancellation_client_allowed`, `late_cancellation_reason_required`),
--   decisión explícita del propietario (DEC-083, alternativa descartada:
--   "solo barbero" como default). Esta migración NO agrega
--   `appointment_closing_mode` ni `auto_close_delay_hours`: pertenecen a
--   una historia futura de cierre automático, fuera del alcance de HU-093.
--
-- Qué NO hace esta migración
--   No agrega ninguna tabla nueva ni política RLS nueva: `barbershop` ya
--   tiene RLS forzada y `barbershop_select_tenant_policy`/
--   `barbershop_update_tenant_policy` ya cubren la fila completa, columnas
--   incluidas (mismo criterio que
--   20260823120000_add_barbershop_contact_info.sql). No concede ningún
--   GRANT nuevo: `barbershop` ya tiene `GRANT SELECT, UPDATE ... TO
--   barberia_app`. No cambia citas existentes, cierre automático ni
--   notificaciones.
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE, ver
--   20260811145252_harden_roles_and_definer_functions.sql). Las columnas
--   nuevas quedan en la misma tabla, ya propiedad de `barberia_owner`.
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
    WHERE table_schema = 'public' AND table_name = 'barbershop' AND column_name = 'min_advance_minutes'
  ) THEN
    RAISE EXCEPTION 'barbershop.min_advance_minutes ya existe: esta migración no puede aplicarse dos veces.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Columnas nuevas
-- ---------------------------------------------------------------------------

ALTER TABLE barbershop
  ADD COLUMN min_advance_minutes               integer NOT NULL DEFAULT 60,
  ADD COLUMN max_advance_days                  integer NOT NULL DEFAULT 3,
  ADD COLUMN slot_grid_minutes                 integer NOT NULL DEFAULT 15,
  ADD COLUMN cancellation_deadline_minutes     integer NOT NULL DEFAULT 20,
  ADD COLUMN late_cancellation_client_allowed  boolean NOT NULL DEFAULT true,
  ADD COLUMN late_cancellation_reason_required boolean NOT NULL DEFAULT true;

-- Rangos de DEC-083 (resuelve DP-PUB-02): generosos alrededor de los
-- valores iniciales de DEC-005/DEC-006/DEC-010/DEC-018, sin permitir
-- configuraciones absurdas.
ALTER TABLE barbershop
  ADD CONSTRAINT barbershop_min_advance_minutes_ck CHECK (min_advance_minutes BETWEEN 0 AND 1440),
  ADD CONSTRAINT barbershop_max_advance_days_ck    CHECK (max_advance_days BETWEEN 1 AND 90),
  -- Conjunto discreto, no un rango continuo (DEC-083): una rejilla no
  -- múltiplo de 5 complica la lectura humana de las franjas sin beneficio.
  ADD CONSTRAINT barbershop_slot_grid_minutes_ck   CHECK (slot_grid_minutes IN (5, 10, 15, 20, 30, 60)),
  ADD CONSTRAINT barbershop_cancellation_deadline_ck CHECK (cancellation_deadline_minutes BETWEEN 0 AND 10080),
  -- CA-093-02: exigir motivo para una cancelación tardía que el cliente ni
  -- siquiera puede hacer es una combinación incoherente (DEC-083 no la
  -- contempla como estado válido).
  ADD CONSTRAINT barbershop_late_cancellation_coherence_ck CHECK (
    late_cancellation_client_allowed OR NOT late_cancellation_reason_required
  ),
  -- Misma protección que DDL-CFG-01 de modelo-fisico-referencia.sql §E.3,
  -- adaptada a los rangos de DEC-083: la anticipación mínima no puede
  -- alcanzar ni superar toda la ventana pública de reserva, o ninguna
  -- franja quedaría reservable.
  ADD CONSTRAINT barbershop_min_advance_vs_window_ck CHECK (
    min_advance_minutes < max_advance_days * 1440
  );

COMMENT ON COLUMN barbershop.min_advance_minutes IS
  'Anticipación mínima en minutos para reservar públicamente (HU-093, RN-DIS-04). Default y rango '
  'fijados por DEC-005/DEC-018/DEC-083. No aplica a la creación manual (RN-DIS-04, excepción). '
  'Propietario funcional: barbería. Clasificación: configuración.';

COMMENT ON COLUMN barbershop.max_advance_days IS
  'Ventana máxima en días para reservar públicamente (HU-093, RN-DIS-04). Default y rango fijados '
  'por DEC-005/DEC-018/DEC-083. No aplica a la creación manual. Propietario funcional: barbería. '
  'Clasificación: configuración.';

COMMENT ON COLUMN barbershop.slot_grid_minutes IS
  'Paso en minutos de la rejilla de franjas ofrecidas al cliente (HU-093, RN-DIS-06), restringido '
  'al conjunto {5,10,15,20,30,60} (DEC-083). Default fijado por DEC-006/DEC-018. No aplica a la '
  'creación manual, que puede fijar cualquier hora. Propietario funcional: barbería. Clasificación: '
  'configuración.';

COMMENT ON COLUMN barbershop.cancellation_deadline_minutes IS
  'Plazo en minutos antes de la cita hasta el cual el cliente puede cancelar por su cuenta (HU-093, '
  'RN-CAN-01). Default y rango fijados por DEC-010/DEC-018/DEC-083. El barbero conserva siempre su '
  'facultad de cancelar sin este plazo (RN-CAN-03). Propietario funcional: barbería. Clasificación: '
  'configuración.';

COMMENT ON COLUMN barbershop.late_cancellation_client_allowed IS
  'Si el cliente puede cancelar vencido el plazo de cancellation_deadline_minutes (HU-093, '
  'RN-CAN-02). Default true (DEC-083). Propietario funcional: barbería. Clasificación: '
  'configuración.';

COMMENT ON COLUMN barbershop.late_cancellation_reason_required IS
  'Si una cancelación tardía del cliente (cuando late_cancellation_client_allowed es true) exige '
  'motivo (HU-093, RN-CAN-02). Default true (DEC-083); no puede ser true si '
  'late_cancellation_client_allowed es false (barbershop_late_cancellation_coherence_ck). '
  'Propietario funcional: barbería. Clasificación: configuración.';
