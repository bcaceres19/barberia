# Recorridos entre pantallas

Scope journeys, tenant propio, precondiciones creadas por UI. No depender de datos que otro agente esté editando.

| ID | Secuencia | Esperado |
| --- | --- | --- |
| UI-J-01 | Configuración → servicio → profesional → asignación → horario → reserva → agenda/detalle | Cita coherente, snapshots y zona, RN-RES-02/RN-HIS-01 |
| UI-J-02 | Alta manual próxima fuera de rejilla pública → agenda | Excepción manual válida, RN-CIT-02/DEC-072/073 |
| UI-J-03 | Dos contextos consultan y confirman misma franja | Una ganadora, conflicto controlado, RN-CON-01/02/05 |
| UI-J-04 | Reserva → cambiar catálogo → detalle | Snapshot previo intacto, RN-SER-04/DEC-004 |
| UI-J-05 | Reserva → reprogramación → Mi turno | Nueva hora e historial coherentes |
| UI-J-06 | Reserva → cancelar cliente → disponibilidad | Estado e historial únicos y hueco futuro liberado, RN-CAN-04 |
| UI-J-07 | Turno → bloqueo encima → agenda | No cancelación automática, RN-BLQ-03/RN-CON-06 |
| UI-J-08 | Retirar última asignación → catálogo → reasignar | Retirada permitida, desaparece/reaparece, DEC-114 |
| UI-J-09 | Guardar → perder respuesta → reintentar → F5 | Un registro/evento, RN-IDE-01 |
| UI-J-10 | Perfil individual → oferta/horarios/agenda → equipo | Profesional real sin duplicados, DEC-115 |
| UI-J-11 | Reservar para otra persona → detalles | Cliente y atendido distintos, RN-RES-03 |
| UI-J-12 | Dos tenants con IDs conocidos cruzados | No fuga y 404, RN-TEN-01 |

## Sesión: scope access y coordinador

- **UI-S-01:** login por formulario, panel y cookie HttpOnly/Secure/SameSite=Lax/Path=/api/v1; token opaco. Hash coincide en BD, sin token en claro. No imprimir cookie ni volcar tabla.
- **UI-S-02:** usar sesión activa y comprobar last_used_at/expires_at y Expires de cookie tras solicitud real. Deben concordar con DEC-050; renovación solo en BD no basta.
- **UI-S-03:** expirar solo sesión QA y recargar/guardar; 401, no revival. Falta permiso de preparación = BLOCKED.
- **UI-S-04:** logout y reutilizar cookie anterior conservada solo en memoria; 401. No exigir 200 a segundo logout sin sesión si el contrato exige autenticación.
- **UI-S-05:** dos sesiones, cerrar una mantiene otra; reset por recuperación revoca ambas de ESA cuenta, CA-008-05.
- **UI-S-06:** renovación/logout simultáneos; estado final revocado y próximo uso rechazado, apoyo SQL para acreditar interior.

La sesión no usa JWT/refresh token: renovación deslizante de token opaco. Estos apoyos no sustituyen pruebas Go/PostgreSQL de las invariantes.
