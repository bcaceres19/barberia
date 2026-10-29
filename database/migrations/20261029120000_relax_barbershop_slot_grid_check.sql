-- Propósito
--   DEC-133: sustituye el conjunto discreto {5, 10, 15, 20, 30, 60} que
--   `DEC-083` impuso a `barbershop.slot_grid_minutes` por un entero de 1 a 60.
--   La barbería elige el paso de su rejilla de horarios (RN-DIS-06) con un campo
--   numérico libre dentro de ese rango; el valor por defecto sigue siendo 15.
--
-- Reglas y decisiones
--   RN-DIS-06, DEC-083 (rango original, que esta migración relaja), DEC-135
--   (las franjas caen en la rejilla fija del tramo, así que funciona con
--   cualquier paso positivo), DEC-035/DEC-036 (estándar de
--   base de datos, Atlas, migración inmutable) y DEC-133.
--
-- Modelo
--   Se reemplaza `barbershop_slot_grid_minutes_ck`
--   (`slot_grid_minutes IN (5, 10, 15, 20, 30, 60)`) por la misma restricción
--   con `BETWEEN 1 AND 60`. El nombre se conserva para que los mensajes de error
--   y la documentación sigan apuntando a la misma regla. La columna, su default
--   y su comentario de tipo no cambian de forma, salvo el texto del comentario.
--
-- Compatibilidad y datos existentes
--   Es una relajación: todo valor que cumplía la restricción anterior cumple la
--   nueva (5, 10, 15, 20, 30 y 60 están en 1..60), así que ninguna fila
--   existente puede violarla. No cambia RLS ni privilegios. El `ALTER TABLE`
--   toma un bloqueo breve sobre `barbershop` (tabla de una fila por barbería)
--   mientras valida la restricción nueva.
--
-- Qué NO hace esta migración
--   No modifica datos, no cambia el default de 15 ni toca las demás restricciones
--   de la política de reserva (anticipación, ventana, plazo de cancelación).
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE, ver
--   20260811145252_harden_roles_and_definer_functions.sql).
--
-- Plan de avance
--   Una corrección posterior se hace con otra migración, nunca editando esta
--   (DEC-036). No existe archivo `down`: el mecanismo normal es roll-forward. Si
--   el negocio volviera a restringir el paso, la nueva migración debe comprobar
--   antes que ninguna fila quede fuera del conjunto.

-- ---------------------------------------------------------------------------
-- 1. Precondiciones
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conrelid = 'barbershop'::regclass AND conname = 'barbershop_slot_grid_minutes_ck'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260911060000_add_barbershop_booking_policy.sql; '
      'barbershop_slot_grid_minutes_ck debe existir antes.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Restricción
-- ---------------------------------------------------------------------------

ALTER TABLE barbershop
  DROP CONSTRAINT barbershop_slot_grid_minutes_ck,
  ADD CONSTRAINT barbershop_slot_grid_minutes_ck CHECK (slot_grid_minutes BETWEEN 1 AND 60);

COMMENT ON COLUMN barbershop.slot_grid_minutes IS
  'Paso en minutos de la rejilla de franjas ofrecidas al cliente (HU-093, RN-DIS-06): entero de 1 a 60 '
  '(DEC-133, que sustituye el conjunto discreto de DEC-083). Default fijado por DEC-006/DEC-018. No aplica '
  'a la creación manual, que puede fijar cualquier hora. Propietario funcional: barbería. Clasificación: '
  'configuración.';

-- ---------------------------------------------------------------------------
-- 3. Postcondiciones
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conrelid = 'barbershop'::regclass
      AND conname = 'barbershop_slot_grid_minutes_ck'
      -- PostgreSQL normaliza BETWEEN a «>= 1 AND <= 60».
      AND pg_get_constraintdef(oid) LIKE '%>= 1%'
      AND pg_get_constraintdef(oid) LIKE '%<= 60%'
  ) THEN
    RAISE EXCEPTION 'barbershop_slot_grid_minutes_ck no quedó con el rango 1..60.';
  END IF;
END
$$;
