-- Solo datos sintéticos del issue #365 (DEC-134): usuarios con cada combinación
-- de roles sobre el par de barberías dedicado a las pruebas del paquete
-- internal/modules/shops/postgres (dos_barberias.sql), más un usuario de la otra
-- barbería para comprobar que los roles nunca cruzan el tenant. Los roles solo
-- se pueden asignar con el rol administrativo, que es con el que se carga este
-- archivo (la aplicación solo los lee).
--
-- Es repetible (ON CONFLICT DO NOTHING). Ningún dato corresponde a una persona
-- real (AGENTS.md).
BEGIN;

INSERT INTO staff_user (id, barbershop_id, email, full_name, is_active) VALUES
  ('365a0001-0000-4000-8000-000000000001', '12121212-1212-1212-1212-121212121212',
   'rol.empleado@ejemplo.test', 'Rol empleado de prueba', true),
  ('365a0001-0000-4000-8000-000000000002', '12121212-1212-1212-1212-121212121212',
   'rol.mixto@ejemplo.test', 'Rol empleado e independiente de prueba', true),
  ('365a0001-0000-4000-8000-000000000003', '12121212-1212-1212-1212-121212121212',
   'rol.independiente@ejemplo.test', 'Rol independiente de prueba', true),
  ('365a0001-0000-4000-8000-000000000004', '12121212-1212-1212-1212-121212121212',
   'rol.empresa@ejemplo.test', 'Rol empresa de prueba', true),
  ('365a0001-0000-4000-8000-000000000005', '12121212-1212-1212-1212-121212121212',
   'rol.ninguno@ejemplo.test', 'Sin rol de prueba', true),
  ('365a0001-0000-4000-8000-000000000006', '34343434-3434-3434-3434-343434343434',
   'rol.otra.barberia@ejemplo.test', 'Rol de otra barbería de prueba', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO staff_user_role (barbershop_id, staff_user_id, role) VALUES
  ('12121212-1212-1212-1212-121212121212', '365a0001-0000-4000-8000-000000000001', 'empleado'),
  ('12121212-1212-1212-1212-121212121212', '365a0001-0000-4000-8000-000000000002', 'empleado'),
  ('12121212-1212-1212-1212-121212121212', '365a0001-0000-4000-8000-000000000002', 'independiente'),
  ('12121212-1212-1212-1212-121212121212', '365a0001-0000-4000-8000-000000000003', 'independiente'),
  ('12121212-1212-1212-1212-121212121212', '365a0001-0000-4000-8000-000000000004', 'empresa'),
  ('34343434-3434-3434-3434-343434343434', '365a0001-0000-4000-8000-000000000006', 'independiente')
ON CONFLICT DO NOTHING;

COMMIT;
