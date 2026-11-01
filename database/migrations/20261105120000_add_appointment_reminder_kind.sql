-- Propósito
--   Issue #467 (HU-131, DEC-144): el recordatorio por correo 30 minutos antes del
--   turno (F-NOT-03, primer tramo) reutiliza la cola de avisos de HU-130
--   (`appointment_notice`): mismo lease, mismo worker, misma evidencia por
--   intento. Esta migración solo amplía lo que esa cola asumía.
--
-- Reglas y decisiones
--   RN-REC-01 (un cambio esencial invalida y regenera el recordatorio, y queda
--   uno solo vigente), RN-REC-02 y RN-REC-03 (se decide con el estado vigente al
--   enviar), RN-REC-04, RN-DAT-02, RN-TEN-01, DEC-032, DEC-040, DEC-142, DEC-144.
--
-- Cambios
--   1. `kind` admite `reminder`.
--   2. Un `reminder` y un aviso de cambio pueden estar `pending` a la vez para el
--      mismo turno (antes el índice único era solo por turno): el recordatorio
--      no debe sustituir ni ser sustituido por una cancelación o reprogramación.
--      Sigue habiendo como máximo UN pendiente de cada clase por turno.
--   3. Para un `reminder`, `state_at` es el instante del turno (`starts_at`) para
--      el que se programó: si el turno cambia de hora, el recordatorio viejo ya
--      no coincide con el vigente y se descarta al enviar (RN-REC-01/03). El
--      contexto del worker devuelve `notice_state_at` para esa comparación y la
--      cobertura (`already_covered`) separa las dos clases: un correo de cambio
--      enviado no cubre un recordatorio ni al revés, y entre recordatorios solo
--      cubre el del mismo instante (los avisos de cambio conservan la regla de
--      HU-130: un estado igual o más reciente).
--
-- Datos existentes
--   Solo filas con kind cancelled_by_barber o rescheduled: ninguna cambia.
--
-- Riesgo
--   El índice se recrea (tabla de una fase recién introducida, sin volumen); el
--   cambio de retorno de `notice_context` exige DROP + CREATE. Ambos corren en la
--   transacción de la migración, y el despliegue de la API que usa el índice
--   nuevo debe ir después de aplicarla.
--
-- Recuperación
--   Antes de usar `reminder`: restaurar el CHECK, el índice por turno y la
--   función anteriores. Con filas `reminder` presentes, primero eliminarlas.

ALTER TABLE appointment_notice DROP CONSTRAINT appointment_notice_kind_ck;
ALTER TABLE appointment_notice
  ADD CONSTRAINT appointment_notice_kind_ck
  CHECK (kind IN ('cancelled_by_barber', 'rescheduled', 'reminder'));

DROP INDEX appointment_notice_pending_uk;
CREATE UNIQUE INDEX appointment_notice_pending_uk
  ON appointment_notice (appointment_id, (kind = 'reminder')) WHERE status = 'pending';

COMMENT ON INDEX appointment_notice_pending_uk IS
  'Como máximo un aviso de cambio y un recordatorio pending por turno (RN-REC-01: sin duplicados).';

DROP FUNCTION notice_context(uuid, uuid);

CREATE FUNCTION notice_context(p_notice_id uuid, p_claim_token uuid)
RETURNS TABLE (
  kind                 text,
  appointment_status   text,
  starts_at            timestamptz,
  appointment_state_at timestamptz,
  customer_email       text,
  attendee_name        text,
  service_name         text,
  barbershop_name      text,
  barbershop_timezone  text,
  already_covered      boolean,
  notice_state_at      timestamptz
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = ''
AS $$
  SELECT n.kind, a.status, a.starts_at, a.updated_at,
         cu.email, a.attendee_name, a.service_name_snapshot, b.name, b.timezone,
         EXISTS (
           SELECT 1 FROM public.appointment_notice o
            WHERE o.barbershop_id = n.barbershop_id AND o.appointment_id = n.appointment_id
              AND o.id <> n.id AND o.status = 'sent' AND (o.kind = 'reminder') = (n.kind = 'reminder')
              AND CASE WHEN n.kind = 'reminder'
                       THEN o.covers_state_at = n.state_at
                       ELSE o.covers_state_at >= n.state_at END
         ),
         n.state_at
    FROM public.appointment_notice n
    JOIN public.appointment a ON a.barbershop_id = n.barbershop_id AND a.id = n.appointment_id
    JOIN public.customer cu ON cu.barbershop_id = a.barbershop_id AND cu.id = a.customer_id
    JOIN public.barbershop b ON b.id = n.barbershop_id
   WHERE n.id = p_notice_id AND n.claim_token = p_claim_token AND n.status = 'processing';
$$;

REVOKE ALL     ON FUNCTION notice_context(uuid, uuid) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION notice_context(uuid, uuid) TO barberia_worker;

COMMENT ON FUNCTION notice_context(uuid, uuid) IS
  'Datos mínimos para redactar y enviar UN aviso reclamado; exige su claim_token vigente, así que no '
  'sirve para leer citas arbitrarias. Incluye el correo y el nombre del cliente solo para ese envío '
  '(RN-DAT-02). notice_state_at es el estado que el aviso cubre: para un recordatorio, el instante '
  'del turno para el que se programó. Exclusivo de barberia_worker.';
