-- Solo datos sintéticos para el E2E de #286. Aplicar con rol migrador en BD local de pruebas.
-- Contraseña ficticia: PruebaBloqueos286!; no usar estas cuentas en producción.
BEGIN;
INSERT INTO barbershop (id,name,timezone) VALUES ('286a0001-0000-4000-8000-000000000001','Equipo de prueba bloqueos 1','America/Bogota') ON CONFLICT (id) DO NOTHING;
INSERT INTO staff_user (id,barbershop_id,email,full_name,is_active) VALUES ('286a0001-0000-4000-8000-000000000002','286a0001-0000-4000-8000-000000000001','bloqueos.286.1@ejemplo.test','Operador de prueba 1',true) ON CONFLICT (id) DO NOTHING;
INSERT INTO staff_credential (staff_user_id,barbershop_id,password_hash,password_algorithm) VALUES ('286a0001-0000-4000-8000-000000000002','286a0001-0000-4000-8000-000000000001','$argon2id$v=19$m=19456,t=2,p=1$Zml4dHVyZTI4NnNhbHQxNg$056gU87nsT2DvgY8THBcSZ7mMnfsLXrOXQZc0bKDjDY','argon2id') ON CONFLICT (staff_user_id) DO NOTHING;
INSERT INTO barber (id,barbershop_id,full_name) VALUES ('286b0001-0000-4000-8000-000000000001','286a0001-0000-4000-8000-000000000001','Alex de prueba') ON CONFLICT (id) DO NOTHING;
INSERT INTO barber (id,barbershop_id,full_name) VALUES ('286b0001-0000-4000-8000-000000000002','286a0001-0000-4000-8000-000000000001','Samuel de prueba') ON CONFLICT (id) DO NOTHING;
INSERT INTO barbershop (id,name,timezone) VALUES ('286a0002-0000-4000-8000-000000000001','Equipo de prueba bloqueos 2','America/Bogota') ON CONFLICT (id) DO NOTHING;
INSERT INTO staff_user (id,barbershop_id,email,full_name,is_active) VALUES ('286a0002-0000-4000-8000-000000000002','286a0002-0000-4000-8000-000000000001','bloqueos.286.2@ejemplo.test','Operador de prueba 2',true) ON CONFLICT (id) DO NOTHING;
INSERT INTO staff_credential (staff_user_id,barbershop_id,password_hash,password_algorithm) VALUES ('286a0002-0000-4000-8000-000000000002','286a0002-0000-4000-8000-000000000001','$argon2id$v=19$m=19456,t=2,p=1$Zml4dHVyZTI4NnNhbHQxNg$056gU87nsT2DvgY8THBcSZ7mMnfsLXrOXQZc0bKDjDY','argon2id') ON CONFLICT (staff_user_id) DO NOTHING;
INSERT INTO barber (id,barbershop_id,full_name) VALUES ('286b0002-0000-4000-8000-000000000001','286a0002-0000-4000-8000-000000000001','Alex de prueba') ON CONFLICT (id) DO NOTHING;
INSERT INTO barber (id,barbershop_id,full_name) VALUES ('286b0002-0000-4000-8000-000000000002','286a0002-0000-4000-8000-000000000001','Samuel de prueba') ON CONFLICT (id) DO NOTHING;
COMMIT;
