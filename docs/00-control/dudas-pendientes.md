---
titulo: "Dudas pendientes y resoluciones"
version: "1.3"
estado: "Tres dudas abiertas"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-07"
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

**Del lote respondido el 5 de agosto de 2026 no queda ninguna duda abierta.** El 7 de agosto de 2026, al redactar las historias del bloque B0 ([historias-usuario.md](../02-requisitos/historias-usuario.md)), se detectaron **tres vacíos nuevos** que `DEC-026` y `DEC-027` no cubren: se registran en la sección 2 bis y bloquean historias concretas.

Las 41 dudas y la contradicción `CT-001` recibieron respuesta del propietario en [respuesta-dudas-pendientes.txt](../../respuesta-manuales/respuesta-dudas-pendientes.txt). No queda ninguna decisión abierta de este lote.

Este archivo conserva los códigos originales y la normalización aplicada. La fuente normativa es [registro-decisiones.md](registro-decisiones.md); el archivo de respuestas se conserva como evidencia literal.

Cuando la respuesta dio un rango o delegó una decisión, se escogió una configuración concreta y se documentó el límite de interpretación:

- ventana máxima inicial: **3 días**, dentro del rango de 1 a 3 días y compatible con reservar un miércoles para el viernes;
- recordatorios: **1 por defecto**, configurable hasta **3**;
- backend: **Go + Chi v5 sobre `net/http`** en un monolito modular (`DEC-034`) y frontend en **Vue 3 + TypeScript + Vite** (`DEC-033`);
- alojamiento: se interpreta “VPN dedicado” como **VPS dedicado**, que es el recurso de cómputo coherente con el contexto;
- “turnos” se usa en la interfaz y la comunicación; “citas” se conserva como término técnico para no renombrar la entidad `appointment`.

## 2. Resoluciones

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

## 2 bis. Dudas abiertas

**Estado: Duda pendiente.** Las tres corresponden al propietario y no se resuelven escribiendo código. Mientras no exista su `DEC-*`, las historias señaladas no deben implementarse.

| Código | Pregunta | Por qué no está resuelta | Qué bloquea | Opciones a considerar |
| --- | --- | --- | --- | --- |
| `DP-SEG-04` | ¿Qué mecanismo sostiene la sesión larga del barbero y cuánto dura exactamente? | `DEC-026` confirma "sesión de larga duración", pero no el mecanismo ni el plazo. `DEC-037` deja expresamente la selección del mecanismo de autenticación "al diseñar la primera entrega". | `HU-005`, `HU-006` | Cookie `HttpOnly` + `Secure` + `SameSite` con token opaco revocable en base de datos, frente a un token firmado sin revocación inmediata. La revocabilidad es requisito de `CA-006-02`. |
| `DP-SEG-05` | ¿Por qué canal y con qué proveedor viaja el código de recuperación al teléfono verificado? | `DEC-026` fija el mecanismo, no el canal. `DEC-027` habilita correo y WhatsApp oficial **para notificaciones de turnos**; extenderlo a códigos de seguridad es una decisión nueva, con costo y requisitos de proveedor propios. | `HU-008`, `HU-011` | WhatsApp oficial reutilizando el proveedor de `DEC-027`, frente a SMS con un proveedor adicional. Afecta costo mensual y verificación previa al piloto. |
| `DP-SEG-06` | ¿Cuánto dura la ventana del límite de 5 solicitudes por IP y cuánto dura el escalamiento? | `DEC-026` fija el umbral inicial (5 solicitudes por IP) pero no la ventana ni la duración de la exigencia telefónica. Sin ese valor, el límite no se puede implementar ni probar. | `HU-007` | Ventanas cortas protegen menos pero molestan menos al barbero legítimo; el valor debe elegirse sabiendo que un barbero bloqueado en plena jornada es un fallo operativo grave. |

Al resolverse cada una se aplica el flujo de la sección 3: `DEC-*`, propagación, conservación de la fila y `CT-*` si revela un conflicto.

## 3. Criterio para nuevas dudas

Una duda nueva recibe un código `DP-*` y se añade aquí solo mientras esté abierta. Al resolverse:

1. se registra una decisión `DEC-*`;
2. se propaga a reglas, alcance, requisitos y arquitectura;
3. se conserva la fila con su resolución;
4. si revela un conflicto entre documentos, se registra además una `CT-*`.
