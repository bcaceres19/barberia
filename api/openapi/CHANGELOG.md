# Registro de cambios del contrato OpenAPI

Formato basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/).
Ver [`docs/06-api/estandar-openapi.md`](../../docs/06-api/estandar-openapi.md)
sección 18 para qué cuenta como cambio compatible o incompatible.

## 0.47.4 · 2026-10-30

### Corregido

- Las rutas bajo `/api/v1/private` reemiten la cookie de sesión con el `Expires`
  renovado cuando la sesión es válida (`DEC-050`, renovación deslizante en cada
  uso); antes el vencimiento del navegador quedaba fijo en el del login. La
  descripción de `GET /private/auth/session` ya no dice que no emite `Set-Cookie`.

### Compatibilidad

- Compatible: no cambia esquemas ni códigos; el token y los atributos de la
  cookie son los del login.

## 0.47.3 · 2026-10-30

### Corregido

- `POST /private/services/{serviceId}/reactivate` responde `409`
  (`ServiceConflictProblem`, `RN-SER-01`) cuando otro servicio activo ya usa el
  nombre del servicio inactivo; antes respondía `500`. No escribe nada ni consume
  la clave de idempotencia. La descripción de la operación lo documenta.

### Compatibilidad

- Sin cambios de esquema: `409` ya era una respuesta declarada de la operación.

## 0.47.2 · 2026-10-30

### Corregido

- `GET /private/services/{serviceId}/deactivation-impact` y la respuesta de
  `POST .../deactivate` informan el recuento real de citas `confirmed` futuras
  del servicio (`RN-SER-03`, `CA-024-01`, `CA-024-04`); antes devolvían `0` fijo,
  valor honesto solo mientras `appointment` no existía (`DEC-069`, B1).

### Compatibilidad

- Sin cambios de esquema ni de códigos de estado; solo cambia el valor de
  `affectedAppointments`, que ya podía ser cualquier entero `>= 0`.

## 0.47.1 · 2026-10-30

### Corregido

- `POST /private/appointments` y `POST /private/appointments/{appointmentId}/reschedule`
  rechazan con `422` y sin escribir un intervalo que no cabe completo en la
  jornada efectiva del barbero (`CA-061-02`, `CA-065-04`); antes solo se
  comprobaban bloqueos y cruces. La descripción de ambas operaciones y de la
  respuesta `AppointmentValidationProblem` lo documenta.

### Compatibilidad

- Sin cambios de esquema ni de códigos de estado: `422` ya era una respuesta
  declarada de ambas operaciones. Compatible a nivel de contrato; endurece una
  validación que la norma ya exigía.

## 0.47.0 · 2026-10-30

### Agregado

- `GET`/`PATCH /private/settings/login-otp` (`getLoginOtpSettings`,
  `updateLoginOtpSettings`, `DEC-139`): el canal del código del reto adicional de
  acceso de la barbería de la sesión, `email`, `whatsapp` o `both`. El cuerpo es
  solo `channel`; un correo o teléfono de destino se rechaza como campo
  desconocido (`400`) y un canal fuera de la lista, con `422`. Esquemas nuevos
  `LoginOtpChannel`, `LoginOtpSettings` y `UpdateLoginOtpRequest`.

### Compatibilidad

- Compatible: solo agrega operaciones y esquemas; ninguna operación existente
  cambia. El reto público `POST /public/auth/challenge` mantiene su contrato y su
  respuesta uniforme `202`; cambia únicamente por dónde sale el código.

## 0.46.0 · 2026-10-30

### Cambiado

- Solo descripciones de `HomeServiceAddress` (`DEC-138`, reemplaza la parte de
  `DEC-131` que exigía una sugerencia del autocompletado): `address` puede ser la
  sugerencia elegida, una dirección aproximada del punto marcado en el mapa o
  «Ubicación marcada en el mapa». Ninguna forma, límite ni regla cambia.

### Compatibilidad

- Compatible: no cambia el esquema ni su validación.

## 0.45.0 · 2026-10-29

### Eliminado

- `googlePlaceId` sale de `BrandResponse` (`GET`/`PATCH /private/settings/brand`)
  y de `UpdateBrandRequest` (`DEC-137`, reemplaza lo que `DEC-129` pedía del
  panel). El identificador de la ficha de Google lo fija quien opera la base de
  datos; la reserva pública (`GET /public/barbershops/{slug}/reviews`) lo sigue
  leyendo de la base y no cambia.

### Compatibilidad

- Es un cambio incompatible para quien lea o envíe `googlePlaceId` en la marca:
  la respuesta ya no lo trae y la solicitud lo rechaza como cualquier campo
  desconocido (`400`). El único consumidor es el cliente tipado de `apps/web`,
  que cambia en el mismo PR.

## 0.44.0 · 2026-10-29

### Cambiado

- Cancelar un turno ya no revoca el enlace del cliente (`DEC-136`, reemplaza esa
  parte de `DEC-089`): `GET /customer/appointments/{token}` de un turno cancelado
  responde `200` con el estado cancelado y repetir `cancelCustomerAppointment` con
  la misma `Idempotency-Key` reproduce la respuesta original. Solo cambian
  descripciones; ninguna forma de solicitud o respuesta se modifica.

## 0.43.0 · 2026-11-03

### Añadido

- Cancelación pública (`HU-099`, `T5`). Nueva operación `cancelCustomerAppointment`
  (`POST /customer/appointments/{token}/cancel`, sin sesión, con
  `Idempotency-Key` e `If-Match`) y cuerpo opcional `CancelCustomerAppointmentRequest`
  (`reason`, de hasta 500 caracteres). Responde `200` con
  `CustomerAppointmentCancelled` (`status: cancelled_by_customer`). Los rechazos
  de política responden `422` con `code` distinguible: `cancellation-window-closed`
  (plazo vencido sin política tardía, o turno ya iniciado) y
  `cancellation-reason-required` (motivo obligatorio ausente o en blanco); versión
  obsoleta y estado no confirmado responden `409` (`version-conflict`,
  `invalid-state`); un enlace inexistente, vencido o ya revocado, el `404` uniforme.
  Cancelar revoca el enlace (`DEC-089`).
- `CustomerAppointmentResponse` suma `versionToken` (el `If-Match` de la cancelación),
  `barbershopContactEmail` y `barbershopContactPhone` (contacto público opcional,
  `null` sin configurar) para mostrar el contacto cuando la política rechaza.

### Compatibilidad

- La operación y los campos son aditivos; los tres campos nuevos de la respuesta
  de `GET` son obligatorios en la respuesta, no en ninguna solicitud, así que un
  cliente que ignora propiedades desconocidas no se rompe.

## 0.42.0 · 2026-11-02

### Añadido

- Reseñas y estrellas de Google (`DEC-129`, `HU-101`). Nueva operación pública
  `getPublicGoogleReviews` (`GET /public/barbershops/{slug}/reviews`, sin sesión):
  calificación, total, enlace a la ficha real y hasta 5 reseñas
  (`PublicGoogleReviews`, `PublicGoogleReview`). Sin ficha, sin credencial de
  Google o ante una consulta fallida responde `200` con la estructura vacía
  explícita (`rating: null`, `ratingCount: 0`, `mapsUrl: null`, `reviews: []`);
  un `slug` inválido o no publicable, el mismo `404` uniforme.
- `BrandResponse` suma `googlePlaceId` (obligatorio, `null` sin ficha) y
  `UpdateBrandRequest` lo acepta opcional: ausente conserva la ficha, `null` o
  vacío la quita y una forma imposible responde `422`.

### Compatibilidad

- El campo y la operación son aditivos; un cliente que ignora propiedades
  desconocidas no se rompe y `googlePlaceId` en la solicitud es opcional.

## 0.41.0 · 2026-11-01

### Añadido

- Servicio a domicilio (`DEC-131`, `HU-103`). `BrandResponse` suma
  `offersHomeService` (obligatorio) y `UpdateBrandRequest` lo acepta opcional:
  ausente conserva el interruptor y apagarlo no afecta turnos ya creados.
  `PublicBarbershopProfile` suma `offersHomeService` (obligatorio).
- `ConfirmPublicAppointmentRequest` acepta `homeService` (`HomeServiceAddress`:
  `address`, `latitude`, `longitude`), `null` o ausente para el local. Con
  domicilio, el servidor revalida el interruptor al confirmar: si la barbería no
  lo ofrece responde `422` sin crear el turno.
- `DailyAgendaEntry` y `AppointmentDetailResponse` suman `homeService` (obligatorio,
  `null` en el local) para que el panel de esa barbería sepa a dónde ir. La
  dirección es dato personal y nunca aparece en la reserva pública.

### Compatibilidad

- Los campos nuevos de respuesta son aditivos; un cliente que ignora
  propiedades desconocidas no se rompe. `homeService` en la solicitud es
  opcional, así que un cliente anterior sigue reservando en el local.

## 0.40.0 · 2026-10-31

### Añadido

- Reseñas internas por barbero (`DEC-128`, `HU-100`).
  `POST /customer/appointments/{token}/review` guarda la calificación (1–5) y el
  comentario opcional (≤500 caracteres) de un turno `completed` sin reseña:
  `404` uniforme para un enlace inválido, `409` uniforme para un turno que no
  admite reseña (no completado, cancelado o ya calificado) y `422` para una
  calificación ausente/fuera de rango o un comentario largo.
- `GET /public/barbershops/{slug}/barbers/{barberId}/reviews` devuelve el
  promedio, el total y hasta 20 reseñas visibles.
- `GET /private/barbers/{barberId}/reviews` y
  `PATCH /private/barbers/{barberId}/reviews/{reviewId}` (`hidden`) listan y
  moderan las reseñas; ocultar es reversible y no borra la fila.
- `BarberResponse` y `PublicBarberResponse` suman `rating` (promedio de 1 decimal
  o `null`) y `ratingCount`, ambos obligatorios, calculados sobre las reseñas no
  ocultas. `CustomerAppointmentResponse` suma `canReview` y `review`, también
  obligatorios.

### Compatibilidad

- Los campos nuevos de respuesta son aditivos; un cliente que ignora
  propiedades desconocidas no se rompe. Las cuatro operaciones son nuevas.

## 0.39.0 · 2026-10-30

### Añadido

- Ubicación física de la barbería (`DEC-130`, `HU-102`). `BrandResponse` suma
  `hasPhysicalLocation`, `address`, `latitude` y `longitude` (siempre presentes,
  con `null` explícito sin dirección) y `UpdateBrandRequest` suma
  `hasPhysicalLocation` y `address`, ambos opcionales: ausentes no tocan la
  ubicación. Una dirección nueva se geocodifica en el servidor y, si no se puede
  ubicar con precisión de calle, `PATCH /private/settings/brand` responde `422`
  sin escribir nada; apagar el interruptor conserva dirección y coordenadas.
  Las coordenadas son solo lectura.
- `PublicBarbershopProfile` suma la propiedad obligatoria `location`
  (`PublicLocation` o `null`): solo viaja con el interruptor activo y
  coordenadas guardadas.

### Compatibilidad

- Compatible para quien envía `PATCH` sin los campos nuevos. Un cliente que
  valida estrictamente la respuesta de `GET /private/settings/brand` o de
  `GET /public/barbershops/{slug}` debe conocer las propiedades nuevas.

## 0.38.0 · 2026-10-28

### Cambiado

- `PanelProfile` añade el valor `employee` (panel individual limitado del rol
  `empleado`) y `panelProfile` de `BrandResponse` deja de ser un dato de la
  barbería: se deriva de los roles del usuario que consulta (`DEC-134`).

- La descripción de `GET .../availability` documenta que las franjas caen en la
  rejilla fija del tramo laboral y no se reinician donde termina un turno
  (`DEC-135`, que sustituye `DEC-084`). Sin cambio de forma; cambia qué franjas
  se ofrecen cuando hay turnos ocupados.

### Eliminado

- `panelProfile` sale de `UpdateBrandRequest`: `PATCH /private/settings/brand`
  rechaza el campo como forma inválida. El panel lo deciden los roles en la base
  de datos, no un ajuste de Configuración.

### Compatibilidad

- **Incompatible** para un cliente que enviaba `panelProfile` en el `PATCH`
  (ahora responde `400`) o que no conoce el valor `employee` al leerlo. El único
  cliente conocido, la web de NAVA, se actualiza en la misma entrega.

## 0.37.0 · 2026-10-28

### Cambiado

- `slotGridMinutes` en `UpdateBookingPolicyRequest` y `BookingPolicyResponse`
  deja de ser el conjunto `{5, 10, 15, 20, 30, 60}` y pasa a ser un entero de 1 a
  60, `DEC-133` (sustituye el conjunto discreto de `DEC-083`). El valor por
  defecto sigue siendo 15. Un valor fuera de ese rango responde `422` como antes.

### Compatibilidad

Compatible para quien escribe: todo valor que se enviaba antes sigue siendo
válido. En la respuesta ahora pueden llegar enteros que antes no existían (por
ejemplo 25 o 45); un cliente que validara el campo contra el conjunto anterior
debe aceptar el rango 1..60.

## 0.36.0 · 2026-10-26

### Agregado

- `description` (cadena de hasta 280 caracteres o `null`) en `BarberResponse`
  (obligatorio, nullable) y, opcional y nullable, en `CreateBarberRequest` y
  `UpdateBarberRequest`, `DEC-127`: descripción corta del barbero que escribe el
  equipo en el panel. En el `PATCH`, ausente conserva la actual y `null` o vacía
  la quita.
- `description` y `photoUpdatedAt` (ambos obligatorios, nullable) en
  `PublicBarberResponse`, `DEC-127`: la reserva pública muestra la descripción y
  la fotografía del barbero.
- `GET /public/barbershops/{slug}/barbers/{barberId}/photo`
  (`getPublicBarberPhoto`), `DEC-127`: bytes de la fotografía de un barbero que
  la reserva pública ya muestra, sin sesión, con `ETag` y `304`. `404` uniforme
  para slug inválido, barbero ajeno, sin fotografía o sin asignación vigente.

### Compatibilidad

Cambio compatible para quien lee: campos nuevos en respuestas y una operación
nueva; los clientes que ignoran campos desconocidos no se rompen. El alta
idempotente ya persistida antes de este cambio (vigencia 24 h) se repite sin
`description`; el cliente lo trata como «sin descripción». Los campos de
solicitud son opcionales.

## 0.35.0 · 2026-10-10

### Agregado

- Etiqueta `Webhooks` y `GET`/`POST /public/webhooks/meta/whatsapp`
  (`verifyMetaWhatsAppWebhook`, `receiveMetaWhatsAppWebhook`), `DEC-126`, issue
  #355: verificación de la suscripción con un token propio y notificaciones de
  Meta (estados de entrega y mensajes entrantes) autenticadas con la firma
  `X-Hub-Signature-256`. Lo llama Meta, no el frontend. Cambio compatible:
  operaciones y esquema nuevos; la ruta solo existe si el despliegue define sus
  dos secretos.

## 0.34.0 · 2026-10-10

### Agregado

- `POST /private/integrations/google-calendar/sync` (`syncGoogleCalendarNow`,
  «Sincronizar ahora»), `DEC-101`, issue #324: hace vencer los cambios pendientes
  de la conexión y reintenta los fallidos; no llama a Google ni toca citas.
- `pendingSyncJobs` y `failedSyncJobs` en `GoogleCalendarConnectionResponse`
  (obligatorios en la respuesta): permiten mostrar «Sincronizando» y ofrecer el
  reintento. Cambio compatible: agrega propiedades a respuestas con
  `additionalProperties: false` que el cliente del repositorio actualiza en el
  mismo cambio.

## 0.33.0 · 2026-10-09

### Agregado

- Etiqueta `Integrations` y las operaciones de la conexión del barbero con su
  Google Calendar (`DEC-099`, `DEC-102`, issue #323): `GET`/`PATCH`/`DELETE
  /private/integrations/google-calendar` (`getGoogleCalendarConnection`,
  `updateGoogleCalendarReminder`, `disconnectGoogleCalendar`), `POST .../connect`
  (`startGoogleCalendarConnection`, devuelve la URL de consentimiento) y `POST
  .../callback` (`completeGoogleCalendarConnection`, la pantalla de retorno del
  frontend reenvía `state` y `code`; responde `connected`, `denied` o `failed`). Actúan
  sobre el barbero vinculado al usuario autenticado y nunca aceptan un
  identificador de barbero, usuario ni barbería; ninguna respuesta lleva tokens,
  secretos, el código OAuth ni el `state`. Cambio compatible: operaciones y
  esquemas nuevos.

## 0.32.0 · 2026-10-09

### Agregado

- `GET`, `PUT` y `DELETE /private/me/barber` (`getMyBarber`, `linkMyBarber`,
  `unlinkMyBarber`), `DEC-100`, issue #322: el usuario autenticado consulta,
  declara y quita cuál barbero es él. Actúa siempre sobre el principal de la
  sesión y el cuerpo (`LinkMyBarberRequest`) solo lleva `barberId`, nunca un
  identificador de usuario. `409` si el barbero ya pertenece a otro usuario.
  Cambio compatible: operaciones nuevas; `BarberResponse` no cambia y sigue sin
  exponer `staffUserId`.

## 0.31.0 · 2026-10-08

### Agregado

- Esquema `PublicVocabulary` (`businessTerm`, `businessTermGender`,
  `professionalTerm`, `professionalTermPlural`, `professionalTermGender`) y
  propiedad obligatoria `vocabulary` en las respuestas `PublicBarbershopProfile`
  (`GET /public/barbershops/{slug}`) y `CustomerAppointmentResponse`
  (`GET /customer/appointments/{token}`), `DEC-119`, issue #309. La reserva
  pública y el acceso al turno nombran al negocio y a su profesional con la
  palabra que la barbería configuró (`DEC-110`). Cambio compatible: solo agrega
  una propiedad a respuestas con `additionalProperties: false` que el cliente
  del repositorio actualiza en el mismo cambio; no expone el acento ni el
  perfil del panel.

## 0.30.1 · 2026-10-08

### Cambiado (comportamiento, no forma del contrato)

- El slug que se genera pasa de `<nombre>-<código de 6>` a
  `<nombre><código de 8>` sin guiones, p. ej. `mateobarberok7x2m9q4`
  (`DEC-118`, issue #306). Mismo patrón `barbershop_public_slug_ck`, así que el
  esquema `PublicLinkResponse` no cambia salvo su descripción y ejemplo. Los
  slugs ya generados se conservan y siguen resolviendo.

## 0.30.0 · 2026-10-08

### Agregado

- `GET /private/settings/public-link` (`operationId: getPublicLink`, tag
  `Settings`, `DEC-117`, issue #304): enlace público de reservas de la
  barbería de la sesión, `{ "slug": "corte-fino-k7x2m9" }`. Si la barbería aún
  no tiene enlace, la primera lectura lo genera y las siguientes devuelven el
  mismo. Respuesta con `Cache-Control: no-store`. Cambio compatible: es una
  operación nueva; no cambia ninguna otra.

### Cambiado (comportamiento, no forma del contrato)

- El slug que se genera al guardar `PUT /private/settings/barbershop` o al
  leer el enlace pasa de `<nombre>`, `<nombre>-2`, … a
  `<nombre>-<código aleatorio de 6>` (`DEC-117`, sustituye el sufijo numérico
  de `DEC-082`). Los slugs ya existentes no cambian y siguen resolviendo en
  `/public/barbershops/{slug}`.

## 0.29.0 · 2026-10-04

### Cambiado

- `DELETE /private/barbers/{barberId}/services/{serviceId}` deja de responder
  `409` cuando la asignación es la última de un servicio activo (`DEC-114`,
  que sustituye a `DEC-068`): se puede retirar a cualquier barbero y el
  servicio queda sin asignaciones, sin ofrecerse al público hasta que se
  asigne otro. Se elimina la response `LastActiveAssignmentConflictProblem`.
  Consumidor afectado: la pantalla «Servicios por barbero», que ya no trata
  ese `409`. El resto de respuestas (`204`, `401`, `404`, `500`) no cambia y
  repetir la operación sigue respondiendo `404`.

## 0.28.0 · 2026-10-04

### Agregado

- `panelProfile` (`shop` | `solo`, `DEC-115`, issue #294) en `BrandResponse`
  (siempre presente) y en `UpdateBrandRequest` (optativo): perfil del panel de
  la barbería. `solo` es el panel de un barbero individual, sin gestión de
  equipo ni selector de barbero; `shop`, el valor inicial, es el panel de
  siempre. Es presentación pura: no limita cuántos barberos existen. Cambio
  compatible: si `PATCH /private/settings/brand` omite el campo, el perfil
  guardado no cambia; un valor fuera de la lista responde `422` sin escribir.

## 0.27.0 · 2026-10-03

### Agregado

- `GET` y `PATCH /private/settings/brand` (`operationId: getBrand`,
  `updateBrand`, tag `Settings`, `DEC-110`, issue #292): marca y vocabulario
  de la barbería. Seis campos siempre presentes: `accent` (lista cerrada
  `brass`, `emerald`, `sapphire`, `ruby`, `amethyst`, `copper`, nunca un
  color libre), `businessTerm`, `professionalTerm` y `professionalTermPlural`
  (2 a 30 caracteres, solo letras, espacios, guion o apóstrofo; el servidor
  los recorta, colapsa y pasa a minúsculas) y los géneros gramaticales
  `businessTermGender` y `professionalTermGender` (`masculine` o
  `feminine`). Un valor inválido responde `422` sin escribir. Recurso propio:
  `BarbershopSettingsResponse` conserva sus cuatro campos (`CA-020-07`). Sin
  `If-Match`: la última escritura gana. Una barbería que nunca lo cambió lee
  los valores iniciales, idénticos a la interfaz anterior.

## 0.26.0 · 2026-10-02

- DEC-107 / #288: `GET /private/barbers` añade modo optativo `page`/`pageSize` y metadatos `page`, `pageSize`, `total`, `totalPages`; conserva el modo cursor existente. Mezclar modos o parámetros numerados inválidos produce 400.

## [0.25.0] - 2026-09-30

### Agregado

- `PUT`, `GET` y `DELETE /private/barbers/{barberId}/photo`
  (`operationId: putBarberPhoto`, `getBarberPhoto`, `deleteBarberPhoto`,
  tag `Staff`, `DEC-104`): fotografía opcional de cada barbero. El `PUT` recibe
  la imagen binaria (`image/jpeg` o `image/png`, máximo 512 KiB, entre 64 y
  1024 px por lado; el formato se decide por los bytes, no por la cabecera) y
  responde el barbero con su nuevo `photoUpdatedAt`; el `GET` sirve los bytes
  con `ETag` y responde `304` a `If-None-Match`; el `DELETE` elimina la imagen
  y es idempotente (`204` también sin fotografía). Sin `Idempotency-Key`: el
  `PUT` reemplaza y es idempotente por naturaleza. Primer contenido binario
  del contrato: caso aprobado por el propietario el 2026-09-30 (excepción
  documentada en `DEC-104` a "archivos o contenido binario solo cuando exista
  un caso aprobado", `docs/06-api/estandar-openapi.md` §10).

### Modificado (compatible)

- `BarberResponse` gana `photoUpdatedAt` (`date-time` o `null`, obligatorio):
  `null` mientras el barbero no tiene fotografía, y su versión cuando existe.
  Lo devuelven `GET`/`POST`/`PATCH` de `/private/barbers` y el listado. Un
  cliente que lo ignora sigue funcionando: el resto de campos no cambia.

## [0.24.0] - 2026-09-28

### Modificado (incompatible)

- `GET /private/services` (`operationId: listServices`, `HU-022`,
  `CA-022-01`, `DEC-103`): pasa de paginación por cursor a paginación por
  número de página. Los parámetros `cursor`/`limit` se reemplazan por
  `page`/`pageSize`; se agrega `search` (filtro opcional por nombre,
  coincidencia parcial insensible a mayúsculas). La respuesta reemplaza
  `nextCursor` por `page`, `pageSize`, `total` y `totalPages`. Excepción
  explícita del propietario a `docs/06-api/estandar-openapi.md` §6.10
  ("listas potencialmente crecientes usan paginación por cursor"), acotada
  a este único endpoint: el panel privado necesita un paginador numerado
  con total exacto para su tabla de servicios, algo que un cursor opaco no
  puede representar sin una consulta `COUNT` aparte de todos modos. Ningún
  otro endpoint de listado del contrato cambia (`BarberListResponse`,
  `AssignmentListResponse`, etc. conservan cursor sin cambios). Consumidor
  único: `apps/web` (módulo `catalog`), migrado en el mismo cambio.

## [0.23.0] - 2026-09-19

### Modificado (compatible)

- `POST /public/auth/recovery/request` (`operationId: requestRecovery`,
  `HU-008`, `CA-008-01`, `CA-008-11`): con el canal `email` se envía el código
  a toda cuenta activa con ese correo, aunque no tenga teléfono o su teléfono
  no esté verificado (`DEC-094`); con `whatsapp` se mantiene el número
  verificado con coincidencia única (`DEC-093`). No cambian el esquema, los
  códigos de estado ni la respuesta `202` idéntica de `DEC-065`: solo se
  amplía el conjunto de cuentas a las que llega el código por correo.

## [0.22.0] - 2026-09-19

### Modificado (incompatible)

- `POST /public/auth/recovery/request`, `/verify` y `/reset-password`
  (`operationId: requestRecovery`, `verifyRecovery`,
  `resetPasswordWithRecoveryToken`, `HU-008`, `CA-008-01`, `CA-008-09`,
  `CA-008-10`): el cuerpo pasa de `{email}` a `{channel, email | phone}`,
  cerrado por canal (`DEC-092`, `DEC-093`). `channel` es `email` o
  `whatsapp`; `email` es obligatorio con `email` y `phone` (E.164) con
  `whatsapp`, y el campo del otro canal no se admite. El código se entrega
  únicamente por el canal elegido, sin respaldo por el otro; los pasos 2 y 3
  reutilizan el mismo canal y valor. La respuesta `202` sigue siendo idéntica
  exista o no la cuenta y sea o no el número ambiguo entre barberías
  (`DEC-065`). Cambio incompatible interno: la pantalla `HU-011` y la API se
  entregan juntas, sin otros consumidores.

## [0.21.0] - 2026-09-15

### Agregado

- `POST /public/barbershops/{slug}/services/{serviceId}/barbers/{barberId}/appointments`
  (`operationId: confirmPublicAppointment`, `CA-097-01` a `CA-097-07`, tag
  `PublicBooking`): confirmación pública concurrente e idempotente (`HU-097`,
  T1 pública), protegida con `Idempotency-Key` (`RN-IDE-01`, `DEC-043`).
  Revalida barbería/servicio/asignación, jornada, bloqueos y política de
  reserva vigentes contra PostgreSQL real antes de persistir (`CA-097-02`);
  la restricción de exclusión de PostgreSQL (`RN-CON-03`) es la última
  defensa contra una carrera entre confirmaciones simultáneas por la misma
  franja (`RN-CON-01`, `RN-CON-02`). Reconcilia el cliente por teléfono y
  correo (`DEC-085`), emite el token de acceso al turno
  (`appointment_access_token`, 256 bits, hash SHA-256, vigencia 90 días,
  `DEC-089`) dentro de la misma transacción atómica, y envía un único
  correo de confirmación síncrono con el enlace de acceso (Resend,
  `DEC-091`). `201` (`ConfirmedPublicAppointmentResponse`, con
  `accessToken` en claro una única vez), `400`, `404` (mismo criterio
  uniforme que `CA-090-02`/`CA-092-03`), `409` (`slot-conflict` con hasta 3
  `alternatives` cronológicamente cercanas, `RN-CON-05`/`DEC-090`, o los
  conflictos de idempotencia ya conocidos), `422`, `500`.

## [0.20.0] - 2026-09-11

### Agregado

- `GET /public/barbershops/{slug}/services` (`operationId:
  listPublicServices`, `CA-091-01` a `CA-091-03`, tag `PublicBooking`):
  catálogo público de servicios (`HU-091`). Lista paginada por cursor de los
  servicios ACTIVOS con al menos una asignación vigente a un barbero de la
  barbería resuelta por el mismo `slug` de `GET /public/barbershops/{slug}`
  (misma resolución sin sesión, mismo `404` uniforme para forma inválida,
  desconocida o no publicable). Proyección mínima:
  `PublicServiceResponse` (`id`, `name`, `description`, `durationMinutes`,
  `price`, `currency`) — nunca estado de ciclo de vida, auditoría ni ningún
  identificador de tenant. Resuelve `DP-PUB-02` como `DEC-083`.

## [0.19.0] - 2026-09-11

### Agregado

- `GET /public/barbershops/{slug}` (`operationId: resolvePublicBarbershop`,
  `CA-090-01` a `CA-090-04`, tag `PublicBooking`): entrada pública de
  reservas (`HU-090`). Resuelve `slug` -el identificador del enlace
  público generado automáticamente desde el nombre de la barbería
  (`DEC-082`, resuelve `DP-PUB-01`)- a la barbería habilitada
  correspondiente, sin sesión y sin que el cliente pueda fijar
  `barbershopId` (`CA-090-03`), mediante la función `SECURITY DEFINER`
  `public_resolve_barbershop_by_slug`
  (`20260911045044_add_barbershop_public_slug.sql`). Un `slug` mal
  formado, desconocido o de una barbería no publicable produce
  EXACTAMENTE la misma respuesta `404` uniforme (`CA-090-02`, `RN-TEN-01`).
  `200`, `404`, `500`.

## [0.18.0] - 2026-09-10

### Agregado

- `POST /private/appointments/{appointmentId}/complete` (`operationId:
  completeAppointment`) y `POST /private/appointments/{appointmentId}/no-show`
  (`operationId: markAppointmentNoShow`) (`CA-067-01` a `CA-067-08`): `T4`
  manual y `T7` (`HU-067`), cierran una cita `confirmed` cuyo `starts_at` ya
  pasó según el reloj del servidor como `completed` o `no_show`
  respectivamente, insertando un único evento (`appointment_completed` o
  `appointment_no_show`) dentro de la misma transacción que la
  actualización de estado. Antes de `starts_at` responden `422` sin
  persistir nada, ni siquiera una reclamación de idempotencia
  (`CA-067-03`). Protegidas con `Idempotency-Key` (`RN-IDE-01`, `DEC-043`)
  y con la precondición `If-Match` (el `versionToken` de `HU-064`). Sin
  cuerpo de solicitud. Repetir el mismo resultado es un no-op exitoso sin
  duplicar el evento; el resultado contrario o cualquier otro estado
  terminal responde `409` orientando a la futura corrección `T8`
  (`CA-067-05`). `200`, `400`, `401`, `404`, `409`
  (versión/estado/idempotencia, códigos distinguibles), `422`, `500`.

## [0.17.0] - 2026-09-10

### Agregado

- `POST /private/appointments/{appointmentId}/cancel` (`operationId:
  cancelAppointmentByBarber`, `CA-066-01` a `CA-066-08`): T6 (`HU-066`),
  cancela una cita `confirmed` en cualquier momento (sin ventana temporal),
  cambiándola a `cancelled_by_barber` e insertando un único evento
  `appointment_cancelled_by_barber`, protegida con `Idempotency-Key`
  (`RN-IDE-01`, `DEC-043`) y con la precondición `If-Match` (el
  `versionToken` de `HU-064`). Sin cuerpo de solicitud. Repetir la
  cancelación sobre una cita ya `cancelled_by_barber` es un no-op exitoso
  sin duplicar el evento (`CA-066-04`); cualquier otro estado terminal
  responde `409`. `200`, `400`, `401`, `404`, `409`
  (versión/estado/idempotencia, códigos distinguibles), `500`.

## [0.16.0] - 2026-09-01

### Agregado

- `POST /private/appointments/{appointmentId}/reschedule` (`operationId:
  rescheduleAppointment`, `CA-065-01` a `CA-065-08`): T2 (`HU-065`), mueve
  el intervalo de una cita `confirmed` a un nuevo inicio futuro, protegida
  con `Idempotency-Key` (`RN-IDE-01`, `DEC-043`) y con la precondición
  `If-Match` (el `versionToken` de `HU-064`). Un bloqueo vigente en el
  nuevo intervalo rechaza con el mismo `409` que un cruce de citas
  (`DEC-076`); un intervalo idéntico es un no-op exitoso. `200`, `400`,
  `401`, `404`, `409` (agenda/versión/estado/idempotencia, códigos
  distinguibles), `422`, `500`.

## [0.15.0] - 2026-09-01

### Agregado

- `GET /private/appointments/{appointmentId}` (`operationId:
  getAppointmentDetail`, `CA-064-01` a `CA-064-04`): detalle completo de un
  turno (`HU-064`), incluidos contacto y nota (solo visibles aquí, nunca en
  la agenda diaria). Nunca incluye `customerId`, ningún identificador de
  actor, `barbershopId` ni `updatedAt` crudo; `versionToken` es el único
  dato de concurrencia expuesto, opaco. `200`, `401`, `404`, `500`.
- `GET /private/appointments/{appointmentId}/history` (`operationId:
  listAppointmentHistory`, `CA-064-05`): historial inmutable paginado por
  cursor (`HU-060`, `HU-064`), orden estable por instante y luego por
  identificador. Cada entrada trae `actorLabel`, un nombre visible seguro
  (nunca correo ni identificador de actor). `200`, `400`, `401`, `404`,
  `500`.

## [0.14.0] - 2026-08-29

### Agregado

- `GET /private/barbers/{barberId}/appointments/daily-agenda`
  (`operationId: listDailyAgenda`, `CA-062-01` a `CA-062-07`): agenda
  diaria de lectura de un único barbero (`HU-062`), calculada en la zona
  IANA de la barbería. La ruta exige `barberId` explícito, sin vista
  consolidada de varios barberos (`DEC-074`); un turno que cruza
  medianoche aparece en cada agenda diaria cuyo rango interseca su
  intervalo (`DEC-075`). Respuesta mínima sin teléfono, correo, nota ni
  `customerId` (`CA-062-05`). `200`, `400`, `401`, `404`, `500`.

## [0.13.0] - 2026-08-27

### Agregado

- `POST /private/appointments` (`operationId: createManualAppointment`,
  `CA-061-01` a `CA-061-08`): alta de un turno manual `confirmed` a
  partir de un turno recibido por teléfono, WhatsApp o en persona
  (`HU-061`), protegida con `Idempotency-Key` (`RN-IDE-01`, `DEC-043`).
  Sin anticipación mínima ni ventana máxima (`RN-DIS-04`, `DEC-005`,
  `DEC-018`); el servicio debe estar activo y asignado al barbero elegido
  (`DEC-072`); un cruce con otra cita (`RN-CON-01`) o con un bloqueo
  vigente (`DEC-073`) responde `409`. El cliente se reconcilia por
  teléfono o correo (`DEC-045`, `DEC-046`, `DEC-071`). `201`, `400`,
  `401`, `404`, `409` (idempotencia o conflicto), `422`, `500`.

## [0.12.0] - 2026-08-25

### Agregado

- `GET`/`PATCH /private/barbers/{barberId}/holiday-calendar`
  (`operationId: getHolidayCalendar`/`updateHolidayCalendar`,
  `CA-041-01`, `CA-041-02`): interruptor de calendario colombiano de
  festivos por barbero. `200`, `400`, `401`, `404`, `500`.
- `GET /private/barbers/{barberId}/schedule-exceptions`
  (`operationId: listScheduleExceptions`, `CA-041-04`, `CA-041-05`): lista
  paginada por cursor de excepciones de jornada, ordenada por fecha
  efectiva. `200`, `400`, `401`, `404`, `500`.
- `POST /private/barbers/{barberId}/schedule-exceptions`
  (`operationId: createScheduleException`, `CA-041-04`, `CA-041-05`): alta
  de una excepción (día cerrado o abierto con tramos), protegida con
  `Idempotency-Key` (`RN-IDE-01`, `DEC-043`). Una excepción ya existente
  para esa fecha, o tramos que se solapan entre sí, responden `409`
  (`code: conflict`). `201`, `400`, `401`, `404`, `409` (idempotencia o
  conflicto), `422`, `500`.
- `GET /private/barbers/{barberId}/schedule-exceptions/{exceptionId}`
  (`operationId: getScheduleException`, `CA-041-06`). `200`, `401`,
  `404`, `500`.
- `PATCH /private/barbers/{barberId}/schedule-exceptions/{exceptionId}`
  (`operationId: updateScheduleException`, `CA-041-04`, `CA-041-05`,
  `CA-041-06`): reemplaza la excepción completa. `200`, `400`, `401`,
  `404`, `409`, `422`, `500`.
- `DELETE /private/barbers/{barberId}/schedule-exceptions/{exceptionId}`
  (`operationId: deleteScheduleException`, `CA-041-06`): retiro físico;
  reintentar tras un `404` es seguro. `204`, `401`, `404`, `500`.
- `GET /private/schedule/colombian-holidays`
  (`operationId: listColombianHolidays`): festivos colombianos de un año
  calendario, calculados de forma determinista (Ley 51 de 1983, "Ley
  Emiliani"), sin tabla ni dependencia nueva. `200`, `400`, `401`, `500`.

No mezcla con horario semanal (`working-hours`, `HU-040`) ni con bloqueos
(`HU-042`). La precedencia de resolución (excepción manual >
festivo automático > horario semanal) vive en un puerto interno de Go, sin
endpoint propio: la disponibilidad (B4) lo consumirá cuando exista.

## [0.11.0] - 2026-08-25

### Agregado

- Tag `Schedule` (HU-040): horario laboral recurrente de cada barbero.
- `GET /private/barbers/{barberId}/working-hours`
  (`operationId: listWorkingHours`, `CA-040-01`): lista paginada por cursor
  de los tramos del barbero de la ruta, ordenada por día ISO y hora de
  inicio. `200` (`WorkingHourListResponse`), `400`, `401`, `404`, `500`.
- `POST /private/barbers/{barberId}/working-hours`
  (`operationId: createWorkingHour`, `CA-040-02`, `CA-040-03`,
  `CA-040-04`): alta de un tramo, protegida con `Idempotency-Key`
  (`RN-IDE-01`, `DEC-043`). Un tramo cuyo `startsTime` + `durationMinutes`
  cruza medianoche es válido (`DEC-020`). Un solape o una hora de inicio
  repetida con otro tramo del mismo barbero y día responde `409`
  (`code: conflict`). `201` (`WorkingHourResponse`), `400`, `401`, `404`,
  `409` (idempotencia o solape), `422`, `500`.
- `GET /private/barbers/{barberId}/working-hours/{workingHourId}`
  (`operationId: getWorkingHour`, `CA-040-05`): lectura individual; un
  tramo inexistente, de otro barbero o de otra barbería responde `404`.
  `200`, `401`, `404`, `500`.
- `PATCH /private/barbers/{barberId}/working-hours/{workingHourId}`
  (`operationId: updateWorkingHour`, `CA-040-04`, `CA-040-05`): reemplaza
  el intervalo completo del tramo (`isoWeekday`, `startsTime`,
  `durationMinutes` obligatorios). `200`, `400`, `401`, `404`, `409`,
  `422`, `500`.
- `DELETE /private/barbers/{barberId}/working-hours/{workingHourId}`
  (`operationId: deleteWorkingHour`, `CA-040-05`): retiro físico
  (`working_hour` no tiene eliminación lógica); reintentar tras un `404` es
  seguro. `204`, `401`, `404`, `500`.

No mezcla con excepciones por fecha/festivos (`HU-041`) ni con bloqueos
(`HU-042`): sus operaciones llegan con sus propias historias. `CT-008` se
resolvió como `DEC-070` antes de esta migración: las FK del modelo físico
de B2 usan `ON DELETE RESTRICT`, no `CASCADE`.

## [0.10.0] - 2026-08-25

### Agregado

- `GET /private/services/{serviceId}/deactivation-impact` (HU-024,
  `operationId: getServiceDeactivationImpact`, `CA-024-01`): previsualiza el
  impacto real de desactivar el servicio, calculado en el momento de la
  solicitud. Siempre `0` en B1 (`DEC-069`): `appointment` no existe todavía
  en la cadena migrada. `200` (`ServiceDeactivationImpactResponse`), `401`,
  `404`, `500`.
- `POST /private/services/{serviceId}/deactivate`
  (`operationId: deactivateService`, `CA-024-02`, `CA-024-03`, `CA-024-04`):
  transición activo → inactivo, protegida con `Idempotency-Key`
  (`RN-IDE-01`, `DEC-043`). Vuelve a consultar el impacto real dentro de la
  misma operación, sin bloqueo optimista (`DEC-069`). Un servicio ya
  inactivo con una clave nueva responde `409` (`code: conflict`,
  transición inválida). `200` (`ServiceDeactivationResponse`), `400`, `401`,
  `404`, `409` (idempotencia u transición inválida), `500`.
- `POST /private/services/{serviceId}/reactivate`
  (`operationId: reactivateService`, `CA-024-05`): transición inactivo →
  activo, misma protección de idempotencia; nunca crea otra fila ni altera
  duración, precio o asignaciones. Un servicio ya activo con una clave nueva
  responde `409`. `200` (`ServiceResponse`), `400`, `401`, `404`, `409`,
  `500`.
- `ServiceResponse` gana `isActive`/`deactivatedAt`, de solo lectura
  (`CA-024-02`, `CA-024-05`): ningún endpoint de HU-022 (`create`/`update`)
  los acepta como entrada; solo cambian mediante las dos operaciones
  anteriores.
- Esquemas `ServiceDeactivationImpactResponse`, `ServiceDeactivationResponse`.
- Responses `ServiceDeactivationImpact`, `ServiceDeactivated`,
  `ServiceReactivated`.
- Tag `Catalog` ampliado para cubrir también el ciclo de vida de servicios.

## [0.9.0] - 2026-08-24

### Agregado

- `GET /private/barbers/{barberId}/services` (HU-023,
  `operationId: listBarberServices`): lista paginada por cursor de los
  servicios asignados a un barbero de la barbería activa. Un `barberId`
  inexistente o de otra barbería responde `404` uniforme (`RN-TEN-01`).
  `200` (`AssignmentListResponse`), `400`, `401`, `404`, `500`.
- `PUT /private/barbers/{barberId}/services/{serviceId}`
  (`operationId: assignServiceToBarber`): asigna un servicio a un barbero.
  Semántica HTTP naturalmente repetible (sin `Idempotency-Key`): repetir la
  misma operación no crea una segunda fila (`CA-023-02`), responde `200` en
  vez de `201` conservando el `createdAt` original. Un mismo servicio puede
  asignarse a varios barberos (`CA-023-03`). Un `barberId`/`serviceId`
  ajeno o inexistente responde el mismo `404` uniforme (`CA-023-04`).
  `200`/`201` (`AssignmentResponse` + `Location`), `401`, `404`, `500`.
- `DELETE /private/barbers/{barberId}/services/{serviceId}`
  (`operationId: unassignServiceFromBarber`): retira la asignación.
  Retirar la última asignación ACTIVA de un servicio ACTIVO se rechaza con
  `409` (`DEC-068`, `code: conflict`); nunca borra al barbero, al servicio
  ni una cita. `204`, `401`, `404`, `409`, `500`.
- Esquemas `AssignmentResponse`, `AssignmentListResponse`: exactamente
  `barberId`, `serviceId`, `createdAt` (`CA-023-07`) — nunca nombre,
  duración, precio ni estado del barbero o del servicio.
- Responses `AssignmentListSuccess`, `AssignmentAssigned`,
  `AssignmentUnassigned`, `LastActiveAssignmentConflictProblem` (nuevo uso
  de `code: conflict` sobre esta relación, `DEC-068`).
- Tag `Catalog` ampliado para cubrir también la asignación de servicios a
  barberos.

## [0.8.0] - 2026-08-24

### Agregado

- `GET /private/services` (HU-022, `operationId: listServices`): lista
  paginada por cursor de los servicios del catálogo de la barbería activa.
  `200` (`ServiceListResponse`), `400` (cursor/limit inválido), `401`,
  `500`.
- `POST /private/services` (`operationId: createService`): alta de un
  servicio, protegida con `Idempotency-Key` (`RN-IDE-01`, `DEC-043`).
  `201` (`ServiceResponse` + `Location`), `400`, `401`, `409` (conflicto u
  operación en curso de idempotencia, `code: idempotency-conflict`/
  `idempotency-locked`; o nombre ya usado por otro servicio activo de la
  misma barbería, `code: conflict`, `DEC-067`), `422` (nombre/descripción/
  duración/precio inválidos), `500`.
- `GET /private/services/{serviceId}` (`operationId: getService`): lectura
  individual; un identificador inexistente o de otra barbería responde el
  mismo `404` (`CA-022-06`, `RN-TEN-01`).
- `PATCH /private/services/{serviceId}` (`operationId: updateService`):
  edición parcial de `name`/`description`/`durationMinutes`/`price`;
  rechaza objetos vacíos y campos desconocidos (nunca `isActive`,
  `currency`, asignaciones a barberos ni alcance de propagación hacia una
  cita, `RN-SER-04`). `200`, `400`, `401`, `404`, `409` (nombre duplicado),
  `422`, `500`.
- Esquemas `ServiceResponse`, `ServiceListResponse`, `CreateServiceRequest`,
  `UpdateServiceRequest`. `price` viaja como string decimal exacto;
  `currency` siempre `"COP"` (`DEC-067`), informativo, nunca aceptado como
  entrada.
- Responses `ServiceListSuccess`, `ServiceSuccess`, `ServiceCreated`,
  `ServiceUpdated`, `ServiceValidationProblem`, `ServiceConflictProblem`
  (nuevo problem type `conflict`, distinto de `idempotency-conflict`).
- Tag `Catalog`.

## [0.7.0] - 2026-08-23

### Agregado

- `GET /private/barbers` (HU-021, `operationId: listBarbers`): lista
  paginada por cursor de los barberos de la barbería activa. `200`
  (`BarberListResponse`), `400` (cursor/limit inválido), `401`, `500`.
- `POST /private/barbers` (`operationId: createBarber`): alta de un barbero,
  protegida con `Idempotency-Key` (`RN-IDE-01`, `DEC-043`). `201`
  (`BarberResponse` + `Location`), `400`, `401`, `409` (conflicto u
  operación en curso de idempotencia), `422` (nombre vacío/solo
  espacios/mayor de 120 caracteres), `500`. No impone unicidad de
  `fullName`.
- `GET /private/barbers/{barberId}` (`operationId: getBarber`): lectura
  individual; un identificador inexistente o de otra barbería responde el
  mismo `404` (`CA-021-05`, `RN-TEN-01`).
- `PATCH /private/barbers/{barberId}` (`operationId: renameBarber`):
  renombrado limitado a `fullName`; rechaza objetos vacíos y campos
  desconocidos (nunca `active`, `deletedAt`, `sortOrder`, `staffUserId`,
  servicios ni horarios, `DEC-047`). `200`, `400`, `401`, `404`, `422`,
  `500`.
- Esquemas `BarberResponse`, `BarberListResponse`, `CreateBarberRequest`,
  `UpdateBarberRequest`.
- Responses `BarberListSuccess`, `BarberSuccess`, `BarberCreated`,
  `BarberUpdated`, `BarberValidationProblem`.
- Header `Location` (componente compartido, primer uso de un `201` en este
  contrato).
- Tag `Staff`.

## [0.6.0] - 2026-08-23

### Agregado

- `GET /private/settings/barbershop` (HU-020, `operationId:
  getBarbershopSettings`): lectura autenticada de la configuración básica
  de la barbería activa (`name`, `timezone`, `contactEmail`,
  `contactPhone`). `200` (`BarbershopSettingsResponse`), `401`, `404`
  defensivo, `500`.
- `PATCH /private/settings/barbershop` (`operationId:
  updateBarbershopSettings`): actualiza los mismos cuatro campos.
  `barbershopId` nunca es un campo aceptado (`CA-020-05`). `200`
  (representación canónica guardada), `400` (JSON/campo desconocido),
  `401`, `404` defensivo, `422` (validación de campo, incluida zona IANA no
  reconocida, `CA-020-03`), `500`.
- Esquemas `BarbershopSettingsResponse`, `UpdateBarbershopSettingsRequest`.
- Responses `BarbershopSettingsSuccess`, `BarbershopSettingsUpdated`,
  `BarbershopSettingsValidationProblem`.
- Tag `Settings`.

## [0.5.0] - 2026-08-17

### Agregado

- `POST /public/auth/recovery/request` (`operationId: requestRecovery`,
  `DEC-064`/`DEC-065`/`DEC-066`): solicita el código de recuperación de
  acceso. Siempre `202` (`RecoveryRequestAcceptedResponse`), sin destino,
  exista o no la cuenta (no enumeración, `CA-008-01`).
- `POST /public/auth/recovery/verify` (`operationId: verifyRecovery`,
  `DEC-064`/`DEC-065`): verifica el código de 6 dígitos. `200` en éxito
  (`RecoveryVerifyResponse`: token de reinicio de un solo uso + destino
  enmascarado, `CA-008-06`); `401` uniforme para código incorrecto, vencido,
  agotado o cuenta inexistente.
- `POST /public/auth/recovery/reset-password`
  (`operationId: resetPasswordWithRecoveryToken`, `DEC-063`/`DEC-064`):
  establece la contraseña nueva con el token de reinicio. `204` en éxito
  (revoca todas las sesiones activas, `CA-008-05`); `401` uniforme para
  token inválido/vencido/reutilizado; `422` si la contraseña incumple la
  política de `DEC-063`.

## [0.4.0] - 2026-08-17

### Agregado

- `429` en `POST /public/auth/login` (HU-007, `DEC-061`/`DEC-062`): la sexta
  solicitud desde la misma IP dentro de la ventana de 15 minutos, y
  cualquier otra mientras el escalamiento siga vigente, responde
  `ChallengeRequiredProblem` (`code: challenge-required`, cabecera
  `Retry-After`) sin evaluar la contraseña.
- `POST /public/auth/challenge` (`operationId: requestPhoneChallenge`,
  `DEC-062`): solicita el código del reto telefónico. Siempre `202`
  (`ChallengeAcceptedResponse`), exista o no la cuenta (no enumeración).
- `POST /public/auth/challenge/verify` (`operationId: verifyPhoneChallenge`,
  `DEC-062`): verifica el código de 6 dígitos. `204` en éxito (limpia el
  escalamiento de la IP); `401` uniforme (`UnauthorizedProblem`) para código
  incorrecto, vencido, agotado o cuenta inexistente.
- Esquemas `ChallengeRequest`, `ChallengeVerifyRequest`,
  `ChallengeAcceptedResponse`. Responses `ChallengeAccepted`,
  `ChallengeVerifySuccess`, `ChallengeRequiredProblem`. Header `Retry-After`.

## [0.3.0] - 2026-08-17

### Agregado

- `GET /private/auth/session` (HU-012, `operationId: getSessionContext`,
  `DEC-060`): lectura no destructiva del contexto de sesión vigente, con
  seguridad `SessionCookie`. Respuesta `200` (`SessionContextResponse`):
  `barbershop.id`/`barbershop.name` y `expiresAt`, sin `Set-Cookie` (no
  emite ni modifica la cookie, a diferencia de login/logout). Errores `401`
  (sesión ausente/inválida/vencida/revocada, mismo `UnauthorizedProblem`
  uniforme que logout), `500`.
- Esquema `SessionContextResponse`.
- Response `SessionContextSuccess`.

### Nota

- Este registro no documenta la adición de `POST /private/auth/logout`
  (HU-006) como una versión propia: quedó dentro de la entrada `0.2.0` de
  abajo aunque `openapi.yaml` ya reflejaba `version: 0.2.0` desde antes de
  esa operación. Se detectó al preparar esta entrada; se deja constancia
  aquí en vez de reescribir el historial de una versión ya publicada.

## [0.2.0] - 2026-08-13

### Agregado

- `POST /public/auth/login` (HU-005, `operationId: loginWithPassword`):
  inicio de sesión con correo y contraseña, sin seguridad (`security: []`,
  `DEC-055`). Respuesta `200` sin secretos (`LoginResponse`, solo
  `expiresAt`); la sesión se fija en `Set-Cookie`, documentada sin su valor
  funcional. Errores `400` (JSON/campo desconocido), `401` (credenciales
  inválidas, unificado para correo inexistente/contraseña
  incorrecta/usuario inactivo), `422` (validación de campo), `500`.
- Esquemas `LoginRequest`, `LoginResponse`.
- Responses `LoginSuccess`, `UnauthorizedProblem`, `ValidationProblem`.
- Header `Set-Cookie` (`SetCookieSession.yaml`).
- Esquema de seguridad `SessionCookie` (`components/security-schemes/`), sin
  registrar todavía en `openapi.yaml` porque ninguna operación lo referencia
  aún (HU-006 lo activa junto con la primera operación privada real).
- Tag `Auth`.

## [0.1.0] - 2026-08-07

`openapi.yaml` existe como documento de entrada sin operaciones (HU-003).
