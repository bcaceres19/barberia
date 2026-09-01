---
titulo: "Dudas pendientes y resoluciones"
version: "2.11"
estado: "Sin duda abierta"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-09-01"
documentos_relacionados:
  - "registro-decisiones.md"
  - "contradicciones.md"
  - "../01-producto/reglas-negocio.md"
  - "../01-producto/alcance-mvp.md"
  - "../02-requisitos/estados-citas.md"
  - "../../respuesta-manuales/respuesta-dudas-pendientes.txt"
---

# Dudas pendientes y resoluciones

## 1. Estado

No hay ninguna duda abierta. La última, `DP-CIT-06` (reprogramación voluntaria hacia un intervalo bloqueado), quedó resuelta el 1 de septiembre de 2026 como `DEC-076` (bloqueo duro, mismo tratamiento que un cruce de citas).

Las cinco dudas anteriores del lote (`DP-CIT-01`–`DP-CIT-05`) se detectaron el 26 de agosto de 2026 al preparar `HU-060`, `HU-061` y `HU-062`: las fuentes confirmaban el modelo de datos, la exclusión de cruces, la creación manual y la agenda diaria, pero no fijaban la reconciliación de clientes manuales sin teléfono, el efecto de la asignación servicio-barbero sobre una cita manual, la conducta frente a un bloqueo vigente, la vista inicial en una barbería con varios barberos ni el día o días en que aparece un turno que cruza medianoche. El propietario resolvió las tres primeras el 27 de agosto de 2026 como `DEC-071`–`DEC-073`, desbloqueando `HU-061`. `DP-CIT-04` y `DP-CIT-05` quedaron resueltas el 28 de agosto de 2026 como `DEC-074`–`DEC-075`, desbloqueando `HU-062`. `DP-CIT-06`, detectada el 31 de agosto de 2026 al redactar `HU-065`, quedó resuelta el 1 de septiembre de 2026 como `DEC-076`, desbloqueando la creación del issue real de `HU-065`.

Las tres dudas anteriores de B1 (`DP-SER-01`–`DP-SER-03`) se detectaron el 24 de agosto de 2026 al preparar `HU-022`, `HU-023` y `HU-024` y sus prompts persistentes (issue documental [#73](https://github.com/bcaceres19/barberia/issues/73)); `F-SERV-01`, `F-SERV-02`, `RN-SER-01`–`RN-SER-04` y el modelo físico de referencia no fijaban por sí solos todas las decisiones observables necesarias para contrato, interfaz y pruebas. El propietario las resolvió el mismo 24 de agosto de 2026 como `DEC-067`–`DEC-069`, y se crearon los issues reales de implementación: [#75](https://github.com/bcaceres19/barberia/issues/75) (`HU-022`), [#76](https://github.com/bcaceres19/barberia/issues/76) (`HU-023`) y [#77](https://github.com/bcaceres19/barberia/issues/77) (`HU-024`).

Antes de este lote no quedaba ninguna duda abierta. Al preparar los prompts de `HU-012`, `HU-007` y `HU-008` (issue documental `#55`) el 14 de agosto de 2026, la comparación entre los criterios, el contrato y el código integrado mostró cinco decisiones que todavía no podían inferirse sin ampliar o debilitar el alcance: el bootstrap autoritativo del panel (`DP-SEG-09`), el reto telefónico completo de `HU-007` (`DP-SEG-10`), la política de contraseña nueva (`DP-SEG-11`), los parámetros del código de recuperación (`DP-SEG-12`) y el proveedor oficial concreto de correo/WhatsApp (`DP-NOT-05`). El propietario resolvió las cinco el 17 de agosto de 2026 como `DEC-060`, `DEC-062`–`DEC-064` y `DEC-066`.

El 7 de agosto de 2026, al redactar las historias del bloque B0 ([historias-usuario.md](../02-requisitos/historias-usuario.md)), se detectaron tres vacíos que `DEC-026` y `DEC-027` no cubrían (`DP-SEG-04`, `DP-SEG-05`, `DP-SEG-06`); el 11 de agosto de 2026 el propietario las resolvió como `DEC-050`, `DEC-051` y `DEC-052`, desbloqueando `HU-005`–`HU-008` y `HU-011`. Al implementar `HU-005` (issue `#44`) el 13 de agosto de 2026 aparecieron dos vacíos más (`DP-SEG-07`, `DP-SEG-08`), resueltos el mismo día como `DEC-057` y `DEC-058`.

Las 41 dudas y la contradicción `CT-001` recibieron respuesta del propietario en [respuesta-dudas-pendientes.txt](../../respuesta-manuales/respuesta-dudas-pendientes.txt). No queda ninguna decisión abierta de ese lote.

Este archivo conserva los códigos originales y la normalización aplicada. La fuente normativa es [registro-decisiones.md](registro-decisiones.md); el archivo de respuestas se conserva como evidencia literal.

Al implementar `HU-010` (issue `#46`) el 13 de agosto de 2026 apareció una duda nueva, `DP-UX-06`, resuelta el mismo día como `DEC-059`.

Cuando la respuesta dio un rango o delegó una decisión, se escogió una configuración concreta y se documentó el límite de interpretación:

- ventana máxima inicial: **3 días**, dentro del rango de 1 a 3 días y compatible con reservar un miércoles para el viernes;
- recordatorios: **1 por defecto**, configurable hasta **3**;
- backend: **Go + Chi v5 sobre `net/http`** en un monolito modular (`DEC-034`) y frontend en **Vue 3 + TypeScript + Vite** (`DEC-033`);
- alojamiento: se interpreta “VPN dedicado” como **VPS dedicado**, que es el recurso de cómputo coherente con el contexto;
- “turnos” se usa en la interfaz y la comunicación; “citas” se conserva como término técnico para no renombrar la entidad `appointment`.

## 2. Dudas abiertas

Ninguna. `DP-CIT-06` (última duda abierta del lote `HU-063`–`HU-065`) quedó resuelta el 1 de septiembre de 2026 como `DEC-076` (sección 3).

## 3. Resoluciones

| Código | Pregunta resumida | Resolución incorporada | Decisión |
| --- | --- | --- | --- |
| `DP-PRD-01` | ¿Turnos o citas? | “Turno” para el usuario; “cita” para documentación técnica, API y datos. | `DEC-016` |
| `DP-PRD-02` | ¿Construir `no_show`? | Sí, como P0 antes del piloto. | `DEC-017` |
| `DP-PRD-03` | Plazo de cancelación | 20 minutos iniciales, configurable por barbería. | `DEC-018` |
| `DP-PRD-04` | Anticipación mínima | 60 minutos iniciales, configurable. | `DEC-018` |
| `DP-PRD-05` | Ventana máxima | 3 días iniciales, configurable. | `DEC-018` |
| `DP-PRD-06` | Uno o varios barberos | El MVP soporta ambos; el aislamiento de agenda es por barbero. | `DEC-019` |
| `DP-PRD-07` | Festivos colombianos | Calendario activable por barbero, bloqueo automático al activarlo y excepción manual para trabajar. | `DEC-020` |
| `DP-PRD-08` | Retrasos que empujan la agenda | El barbero decide; puede avisar, reprogramar o cancelar. No hay desplazamiento automático en cadena. | `DEC-021` |
| `DP-PRD-09` | Corregir un estado marcado por error | Sí, mediante corrección auditada y sin borrar el estado ni el historial anterior. | `DEC-017` |
| `DP-PRD-10` | Cierre de citas abiertas | Modo configurable: cierre manual asistido o cierre automático después de X horas. | `DEC-017` |
| `DP-UX-01` | Selección de barbero | El cliente elige cuando hay varios; con uno queda preseleccionado sin paso adicional. | `DEC-019` |
| `DP-UX-02` | Campos del formulario | Nombre, teléfono y correo obligatorios; nota opcional; persona atendida solo se pide aparte si difiere. | `DEC-022` |
| `DP-UX-03` | Validación del nombre | Texto no vacío con ejemplo claro; sin exigir dos palabras. | `DEC-022` |
| `DP-UX-04` | Acceso del cliente a su turno | Enlace aleatorio largo enviado por correo; portal por teléfono/correo y bot quedan para después. | `DEC-022` |
| `DP-UX-05` | Paso de la rejilla | 15 minutos iniciales, configurable. | `DEC-018` |
| `DP-BE-01` | Tecnología de implementación | Go + Chi v5 sobre `net/http`, monolito modular; frontend en Vue 3 + TypeScript + Vite. | `DEC-023`, `DEC-033`, `DEC-034` |
| `DP-BE-02` | Bloqueos puntuales o recurrentes | Ambos: puntual, recurrencia y listas explícitas de fechas, con excepciones. | `DEC-020` |
| `DP-BE-03` | Cruce de medianoche | Permitido si todo el intervalo cabe en el horario configurado del barbero. | `DEC-020` |
| `DP-BE-04` | Extremos del intervalo | Intervalos semiabiertos `[inicio, fin)`. | `DEC-020` |
| `DP-BD-01` | Motor de base de datos | PostgreSQL. | `DEC-024` |
| `DP-BD-02` | Aislamiento entre barberías | Esquema compartido con seguridad a nivel de fila (RLS) y `barbershop_id`. | `DEC-024` |
| `DP-BD-03` | Conservación de datos | 24 meses por defecto, configurable; después se anonimizan los datos personales. | `DEC-025` |
| `DP-SEG-01` | Autenticación del barbero | Correo y contraseña, sesión larga y recuperación con código al teléfono verificado. | `DEC-026` |
| `DP-SEG-02` | Protección contra abuso | Límite configurable por IP; al superarlo se exige código al teléfono. Valor inicial: 5 solicitudes. | `DEC-026` |
| `DP-SEG-03` | Incidente en el piloto | Procedimiento escrito de contención, evaluación, aviso y registro. | `DEC-026` |
| `DP-NOT-01` | Canales iniciales | WhatsApp oficial y correo, habilitables por barbería y por tipo de evento. | `DEC-027` |
| `DP-NOT-02` | Anticipación del recordatorio | 30 minutos iniciales, configurable. | `DEC-018` |
| `DP-NOT-03` | Cantidad de recordatorios | 1 por defecto, configurable entre 0 y 3. | `DEC-018` |
| `DP-NOT-04` | Aviso de inasistencia | Configurable por barbería; desactivado por defecto. | `DEC-018` |
| `DP-NOT-05` | ¿Qué proveedor/adaptador oficial concreto implementará el envío de códigos de seguridad por WhatsApp y correo? | Meta WhatsApp Cloud API directo (plantilla "Authentication") para el teléfono; Resend (o SES si ya hay AWS) para correo; puerto único en el módulo `notification` con timeout de 5 s y sin reintento síncrono. | `DEC-066` |
| `DP-PIL-01` | Duración del piloto | 4 semanas. | `DEC-028` |
| `DP-PIL-02` | Interrupción del piloto | El propietario entrega manualmente al barbero las citas futuras y pendientes. | `DEC-028` |
| `DP-PIL-03` | Método anterior en paralelo | Solo durante la primera semana; después se usa la plataforma. | `DEC-028` |
| `DP-COM-01` | Hipótesis de precio | El precio se define al finalizar el piloto según costos y utilidad. | `DEC-029` |
| `DP-COM-02` | Comisión del colaborador | Diferida hasta acordar por escrito porcentaje o monto fijo antes del primer pago asociado. | `DEC-029` |
| `DP-COM-03` | Gratuidad del piloto | Gratuito para 2 o 3 participantes durante un mes; posible extensión de un mes y 3 meses de membresía posterior. | `DEC-028` |
| `DP-LEG-01` | Política de datos | Aviso y política adaptada antes del piloto; revisión jurídica antes de producción comercial. | `DEC-030` |
| `DP-LEG-02` | Eliminación de datos | Anonimizar datos personales y conservar el historial operativo. | `DEC-025` |
| `DP-LEG-03` | Acuerdo con el colaborador | Acuerdo breve y escrito antes del piloto. | `DEC-030` |
| `DP-OPS-01` | Copias de seguridad | Diarias, con retención de 30 días y restauración probada. | `DEC-031` |
| `DP-OPS-02` | Alojamiento | Servicios gratuitos gestionados en desarrollo/piloto; VPS dedicado económico en producción. | `DEC-031` |
| `DP-OPS-03` | Soporte del piloto | Menos de una hora la primera semana; mismo día después, atendiendo cualquier alerta. | `DEC-031` |
| `CT-001` | Recordatorios P0 o P1 | `F-NOT-03` se eleva a P0 y se construye junto con su maquinaria. | `DEC-032` |
| `DP-DDL-01` | Vocabulario de `event_type` | Inglés `snake_case`, con prefijo `appointment_`. | `DEC-041` |
| `DP-DDL-02` | Fecha ancla de retención/anonimización | Última actividad del cliente (cita o contacto más reciente). | `DEC-042` |
| `DP-DDL-03` | Semántica concurrente de idempotencia | Bloqueo consultivo transaccional por tenant+clave; `409` inmediato sin espera. | `DEC-043` |
| `DP-DDL-04` | Versión mínima de PostgreSQL | Se confirma 14, sin cambio. | `DEC-044` |
| `DP-DDL-05` | Identidad/reutilización de `customer` por teléfono | Único por barbería, con upsert. | `DEC-045` |
| `DP-DDL-06` | Unicidad de correo de `customer` | Por barbería, no global. | `DEC-046` |
| `DP-DDL-07` | Ciclo de vida de `barber` | Recortado al alcance de `HU-021`: alta, listado, renombrar. | `DEC-047` |
| `DP-DDL-08` | Valores del 2do y 3er recordatorio | 24 horas y 2 horas antes de la cita. | `DEC-048` |
| `DP-DDL-09` | Matriz completa de anonimización | Cubre todas las copias de datos personales, no solo `customer`. | `DEC-049` |
| `DP-DDL-10` | Modelo de roles PostgreSQL | Propietario `NOLOGIN`, `barberia_app` y `barberia_worker` separados. | `DEC-040` |
| `DP-SEG-04` | Mecanismo y duración de la sesión larga del barbero | Cookie `HttpOnly`+`Secure`+`SameSite`, token opaco revocable en `staff_session`, 30 días con renovación por uso. | `DEC-050` |
| `DP-SEG-05` | Canal del código de recuperación de acceso | WhatsApp oficial y correo, mismo proveedor de `DEC-027`. | `DEC-051` |
| `DP-SEG-06` | Ventana y escalamiento del límite de acceso por IP | Ventana de 15 minutos; escalamiento a verificación telefónica de 24 horas. | `DEC-052` |
| `DP-SEG-07` | Atributos exactos de la cookie de sesión (`SameSite`, `Path`, `Domain`, nombre) | `barberia_session`, `Path=/api/v1`, `SameSite=Lax`, sin `Domain`, 30 días. | `DEC-057` |
| `DP-SEG-08` | Evidencia de aislamiento de `CA-005-05` sin un endpoint privado real todavía | Dividida: `HU-005` prueba aislamiento a nivel PostgreSQL/RLS; `HU-006` prueba end-to-end contra el logout real (`CA-006-07`). | `DEC-058` |
| `DP-UX-06` | Destino del enlace de recuperación de `CA-010-08` mientras `HU-011` no existe | Ruta real `/recuperar-acceso`, cargada de forma diferida, que declara explícitamente que la recuperación aún no está disponible; no simula el flujo de `HU-011`. | `DEC-059` |
| `DP-SER-01` | ¿Moneda COP fija o configurable; se permiten servicios gratuitos; nombres duplicados entre servicios activos? | COP fija, precio estrictamente mayor que cero (sin gratuitos), nombre único entre servicios activos de la misma barbería. | `DEC-067` |
| `DP-SER-02` | ¿Se permite retirar la última asignación de barbero de un servicio activo? | No; un servicio activo debe conservar al menos un barbero asignado, la operación se rechaza (bloqueo duro). | `DEC-068` |
| `DP-SER-03` | ¿Qué parte de la advertencia de citas futuras cierra B1 al desactivar un servicio, antes de `appointment` en B3? | B1 construye un recuento simple sobre el estado real (0 hasta que exista `appointment`), sin protección de concurrencia; cancelación selectiva y bloqueo optimista quedan para B3. | `DEC-069` |
| `DP-SEG-09` | ¿Cuál es la operación y el payload autoritativos con que `HU-012` rehidrata una cookie `HttpOnly` y obtiene la barbería activa al abrir la aplicación? | Nueva operación `GET /api/v1/private/auth/session`, protegida por `SessionCookie` sobre el mismo `SessionMiddleware` existente, con payload mínimo `{ barbershop: { id, name }, expiresAt }`. | `DEC-060` |
| `DP-SEG-10` | ¿Cómo funciona de extremo a extremo el reto telefónico de `HU-007` después del umbral? | Dos operaciones nuevas, `POST /api/v1/public/auth/challenge` y `.../challenge/verify`, código de 6 dígitos por WhatsApp oficial, vigencia 5 min, 5 intentos, límite propio de reenvío; verificar con éxito limpia `escalated_until` de esa IP. | `DEC-062` |
| `DP-SEG-11` | ¿Cuál es la política mínima exacta para la contraseña nueva de `CA-008-08`? | Longitud 10–128, sin exigencia de composición, rechazo si es igual al correo o a la contraseña actual; sin verificación contra lista externa de contraseñas filtradas en el MVP. | `DEC-063` |
| `DP-SEG-12` | ¿Cuáles son el formato/entropía y los valores iniciales del código de recuperación, su vigencia, máximo de intentos, control de reenvío y autorización entre “verificar” y “cambiar contraseña”? | Código de 6 dígitos, `HMAC-SHA256` con secreto de despliegue, vigencia 15 min, 5 intentos, reenvío con cooldown de 60 s y máximo 3/hora; token opaco de reinicio (mismo patrón que el token de sesión) de un solo uso entre verificar y cambiar contraseña. | `DEC-064` |
| `DP-CIT-01` | En una cita manual, ¿cómo se identifica o reutiliza `customer` cuando falta el teléfono? | Con correo presente, se reconcilia por correo dentro de la barbería; sin teléfono ni correo, siempre fila nueva, nunca por nombre. | `DEC-071` |
| `DP-CIT-02` | ¿Una cita manual solo puede usar servicios activos asignados al barbero elegido, o basta que el servicio esté activo en la barbería? | Solo servicios activos ya asignados al barbero elegido (`barber_service` vigente). | `DEC-072` |
| `DP-CIT-03` | Si el intervalo coincide con un bloqueo vigente, ¿la creación manual se rechaza, se permite con advertencia o exige retirar/exceptuar el bloqueo? | Se rechaza (bloqueo duro), mismo tratamiento que un cruce de citas; el barbero retira o exceptúa el bloqueo desde `HU-041`/`HU-042` antes de crear la cita. | `DEC-073` |
| `DP-CIT-04` | En una barbería con varios barberos y sin vínculo automático `staff_user`→`barber`, ¿la agenda abre con selector obligatorio de un barbero, con una vista consolidada o con otra selección inicial? | Selector obligatorio de un barbero; sin vista consolidada. Con un solo barbero, preselección sin paso adicional. | `DEC-074` |
| `DP-CIT-05` | Si un turno cruza medianoche, ¿aparece solo en el día civil de inicio o en cada agenda diaria cuyo intervalo intersecta? | En cada agenda diaria cuyo rango civil interseca el intervalo del turno (día de inicio y día siguiente cuando corresponda). | `DEC-075` |
| `DP-CIT-06` | Si el nuevo intervalo de una reprogramación (T2) coincide con un bloqueo vigente del mismo barbero, ¿se rechaza, se permite con advertencia o exige retirar/exceptuar el bloqueo? | Se rechaza (bloqueo duro), mismo tratamiento que un cruce de citas; el barbero retira o exceptúa el bloqueo desde `HU-041`/`HU-042` antes de reprogramar hacia ese intervalo. | `DEC-076` |

Al resolverse cada duda se aplica el flujo de la sección 3: `DEC-*`, propagación, conservación de la fila y `CT-*` si revela un conflicto.

## 4. Criterio para nuevas dudas

Una duda nueva recibe un código `DP-*` y se añade aquí solo mientras esté abierta. Al resolverse:

1. se registra una decisión `DEC-*`;
2. se propaga a reglas, alcance, requisitos y arquitectura;
3. se conserva la fila con su resolución;
4. si revela un conflicto entre documentos, se registra además una `CT-*`.
