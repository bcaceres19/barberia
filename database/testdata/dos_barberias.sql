-- Escenario controlado de HU-001: dos barberías con usuarios propios.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--
-- Se ejecuta con el rol migrador, que es el único con política administrativa
-- sobre las tablas con RLS forzada.
--
-- Identificadores fijos a propósito: las pruebas de aislamiento necesitan
-- conocer un identificador de la barbería ajena para intentar alcanzarlo
-- (base-datos.md §4.8).
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

INSERT INTO barbershop (id, name, timezone) VALUES
  ('11111111-1111-1111-1111-111111111111', 'Barbería de prueba A', 'America/Bogota'),
  ('22222222-2222-2222-2222-222222222222', 'Barbería de prueba B', 'America/Bogota')
ON CONFLICT (id) DO NOTHING;

-- Par dedicado (NO 33333333.../44444444...: esos ids ya son "shopC/shopD" de
-- testdata/hu021_barberos.sql, reutilizados por hu022/hu023; un conflicto de
-- nombre ahí es un fixture roto, no una fila nueva): reservado en exclusiva
-- para pruebas de integración de PAQUETES Go distintos a los que ya usan A/B
-- para las mismas columnas mutables de `barbershop`
-- (name/timezone/contact_*). `go test ./...` ejecuta paquetes en binarios
-- separados EN PARALELO por defecto (issue #158): si dos paquetes hacen
-- UPDATE confirmado sobre la misma fila A o B, uno puede leer el estado a
-- medio escribir del otro. Úsalo solo si tu paquete escribe esas columnas Y
-- ya existe otro paquete que también las escribe sobre A/B (verifica antes
-- de reutilizar A/B para ese propósito), y antes de acuñar un id nuevo aquí
-- confirma con grep que ningún otro testdata/*.sql ya lo reclama.
INSERT INTO barbershop (id, name, timezone) VALUES
  ('12121212-1212-1212-1212-121212121212', 'Barbería de prueba (aislamiento de paquete) 1', 'America/Bogota'),
  ('34343434-3434-3434-3434-343434343434', 'Barbería de prueba (aislamiento de paquete) 2', 'America/Bogota')
ON CONFLICT (id) DO NOTHING;

INSERT INTO staff_user (id, barbershop_id, email, full_name, is_active) VALUES
  ('aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaa1',
   '11111111-1111-1111-1111-111111111111',
   'duena.a@ejemplo.test',  'Dueña A',   true),

  ('aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2',
   '11111111-1111-1111-1111-111111111111',
   'barbero.a@ejemplo.test', 'Barbero A', true),

  ('bbbbbbb1-bbbb-bbbb-bbbb-bbbbbbbbbbb1',
   '22222222-2222-2222-2222-222222222222',
   'dueno.b@ejemplo.test',  'Dueño B',   true),

  ('bbbbbbb2-bbbb-bbbb-bbbb-bbbbbbbbbbb2',
   '22222222-2222-2222-2222-222222222222',
   'barbero.b@ejemplo.test', 'Barbero B', false)
ON CONFLICT (id) DO NOTHING;

-- Tercer usuario por barbería (aaaaaaa3/bbbbbbb3): activo, SIN teléfono
-- verificado, dedicado a pruebas de cmd/api que solo necesitan UNA sesión
-- válida para ejercitar una feature ajena a recuperación (catálogo,
-- barberos, turnos, servicios...). auth_recovery_change_password (issue
-- #158) revoca TODAS las sesiones activas del usuario cuya contraseña
-- cambia (CA-008-05); como duena.a/dueno.b son las únicas cuentas con
-- teléfono verificado (hu007_reto_telefonico.sql) y las únicas que ejercitan
-- recuperación en cualquier paquete, cualquier sesión de cmd/api abierta
-- sobre ELLAS puede morir a mitad de prueba si un paquete concurrente
-- termina un flujo de recuperación real. aaaaaaa3/bbbbbbb3 nunca son el
-- objetivo de una prueba de recuperación en ningún paquete: mantenlo así.
-- internal/platform/database/database_test.go cuenta USUARIOS POR
-- BARBERÍA bajo RLS (no total): añadir esta fila exige mantener ese conteo
-- en 3, no en 2.
INSERT INTO staff_user (id, barbershop_id, email, full_name, is_active) VALUES
  ('aaaaaaa3-aaaa-aaaa-aaaa-aaaaaaaaaaa3',
   '11111111-1111-1111-1111-111111111111',
   'sesion.a@ejemplo.test', 'Sesión de prueba A', true),

  ('bbbbbbb3-bbbb-bbbb-bbbb-bbbbbbbbbbb3',
   '22222222-2222-2222-2222-222222222222',
   'sesion.b@ejemplo.test', 'Sesión de prueba B', true)
ON CONFLICT (id) DO NOTHING;

COMMIT;
