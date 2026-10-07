# Inventario y reparto

17 pantallas verificadas contra apps/web/src/modules/*/routes.ts. Diálogos/pasos pertenecen a su ficha. Un tenant por scope, máximo dos agentes simultáneos.

| Scope/ficha | Pantalla | Ruta | HU |
| --- | --- | --- | --- |
| [access](pantallas/access.md) | Acceso y sesión | /acceso | HU-005,HU-006,HU-007,HU-010,HU-012 |
| [recovery](pantallas/recovery.md) | Recuperación | /recuperar-acceso | HU-008,HU-011 |
| [shop](pantallas/shop.md) | Barbería, marca y perfil | /panel/barberia | HU-020,HU-025,HU-012 |
| [policy](pantallas/policy.md) | Política pública | /panel/reserva-publica | HU-093 |
| [staff](pantallas/staff.md) | Barberos, fotos y bloqueos | /panel/barberos | HU-021,HU-042 |
| [catalog](pantallas/catalog.md) | Servicios y ciclo de vida | /panel/servicios | HU-022,HU-024 |
| [assignments](pantallas/assignments.md) | Servicios por barbero | /panel/servicios-por-barbero | HU-023 |
| [schedules](pantallas/schedules.md) | Horarios, excepciones y festivos | /panel/horarios | HU-040,HU-041 |
| [agenda](pantallas/agenda.md) | Agenda diaria | /panel | HU-062,HU-063 |
| [manual](pantallas/manual.md) | Nuevo turno manual | /panel/turnos/nuevo | HU-061 |
| [detail](pantallas/detail.md) | Detalle e historial | /panel/turnos/:appointmentId | HU-064,HU-065,HU-066,HU-067,HU-068 |
| [public-entry](pantallas/public-entry.md) | Entrada pública | /reservar/:slug | HU-090 |
| [public-catalog](pantallas/public-catalog.md) | Catálogo público | /reservar/:slug/servicios | HU-091 |
| [public-barber](pantallas/public-barber.md) | Selección pública de profesional | /reservar/:slug/servicios/:serviceId/barbero | HU-092 |
| [public-slots](pantallas/public-slots.md) | Disponibilidad pública | /reservar/:slug/servicios/:serviceId/barbero/:barberId/horario | HU-094,HU-095 |
| [public-confirm](pantallas/public-confirm.md) | Cliente y confirmación pública | /reservar/:slug/servicios/:serviceId/barbero/:barberId/horario/:startsAt/cliente | HU-096,HU-097 |
| [customer](pantallas/customer.md) | Mi turno y cancelación pública | /mi-turno/:token | HU-098,HU-099 |

## Rutas y variantes

/panel/bloqueos redirige a Barberos DEC-105. Perfil individual es variante DEC-115, no ruta nueva. Reto pertenece a acceso; recuperación tiene tres pasos en una ruta. Resumen/confirmación están en public-confirm, no /confirmar. Mi turno recibe credencial de cita; cliente no crea cuenta.

## Orden

Smoke, access y recovery seriales. Cola de pantallas privadas/públicas con tenants independientes, máximo dos agentes. Cada uno prepara sus escenarios; no depende de la mutación de otro. Al final journeys y aislamiento en serie, usando journeys/isolation-a/isolation-b. Cruces intencionados solo entre estos tenants QA.

Revisar rutas en cada SHA. Nueva pantalla requiere ficha y prompt trazable antes de afirmar cobertura; retirada queda N/A con razón. Las tareas son particiones de una campaña de exploración, no entregas de implementación independientes.
