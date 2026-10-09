# Cobertura y límites

«Toda la app» activa todas las pantallas/recorridos del inventario en el SHA elegido. Casos no terminados quedan pendientes explícitos. No se limita al panel ni a happy paths.

UI: sesión/reto, recuperación, marca/perfil, política pública, profesionales/fotos/bloqueos, servicios/ciclo de vida, asignaciones, jornadas/excepciones/festivos, agenda, alta manual, detalle/historial/reprogramación/cancelación/cierre/corrección, cinco pasos públicos y Mi turno/cancelación cliente.

API real: status/respuesta mínima/request_id del recorrido. SQL: consultas/preparación estrechas en tenants QA para hash/expiración/renovación/revocación, con roles autorizados. La UI no prueba toda RLS ni atomicidad: complementar suites Go/SQL de dos tenants si se pide auditoría de backend.

Worker, recordatorios/reintentos, leases, anonimización/retención, recuperación operativa y migraciones no reciben PASS por navegar: BACKEND_ONLY/NOT_RUN y verificación separada según estrategia-pruebas.md. Google Calendar DEC-099–102 es trabajo documentado: inventariar implementación antes de incluirlo, sin exigir botón inexistente.

El marcador local permite completar OTP/correo, pero no prueba entrega externa. E2E existente con mocks puede probar una pantalla, no su backend: clasificarlo antes de reutilizar evidencia.

Cierre: todos los casos tienen estado, FAIL con evidencia, pendientes y SHA conocidos, cobertura por pantalla/factor y severidad. Resultado correcto: «No se encontró fallo en los casos ejecutados». Nunca «sin bugs», «todo full» o «100% seguro».
