---
titulo: "Estándar de diseño y evolución de base de datos"
version: "1.1"
estado: "Obligatorio para desarrollo"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-06"
documentos_relacionados:
  - "../00-control/registro-decisiones.md"
  - "../01-producto/reglas-negocio.md"
  - "../02-requisitos/estados-citas.md"
  - "base-datos.md"
  - "migraciones-atlas.md"
  - "../03-desarrollo/estrategia-pruebas.md"
  - "../03-desarrollo/estandar-backend-go.md"
---

# Estándar de diseño y evolución de base de datos

## 1. Propósito y alcance

Este estándar gobierna el modelo PostgreSQL, las migraciones, consultas, índices, roles y pruebas de datos del MVP. La base de datos protege invariantes que no pueden depender de una sola instancia del API: pertenencia al tenant, referencias válidas, estados permitidos, agenda sin cruces e historial inmutable.

El diseño parte de tercera forma normal y solo se desnormaliza cuando existe una necesidad histórica, de seguridad o rendimiento demostrada y probada.

Fuente normativa: `DEC-035`, en complemento de `DEC-024` y `DEC-025`.

## 2. Organización de artefactos

```text
database/
  atlas.hcl
  migrations/
    atlas.sum
  seeds/
  testdata/
  tests/
  README.md
```

- `migrations/`: cambios de esquema, roles, funciones, políticas y datos de referencia indispensables.
- `seeds/`: datos ficticios repetibles para desarrollo; nunca se ejecutan automáticamente en producción.
- `testdata/`: escenarios controlados para pruebas de integración y volumen.
- `tests/`: comprobaciones SQL globales de restricciones, RLS y migración.
- `README.md`: versión mínima de PostgreSQL, herramienta elegida, comandos de aplicar/verificar y procedimiento de recuperación.

El modelo se documenta además con un diagrama y diccionario cuando se diseñen las primeras tablas. Cada tabla debe indicar propósito, propietario funcional, retención y clasificación de datos.

## 3. Normalización obligatoria

### Primera forma normal

- Cada fila representa una instancia y cada columna un valor del atributo definido.
- No usar columnas repetidas como `phone_1`, `phone_2` ni listas separadas por comas.
- Relaciones de varios elementos viven en tablas hijas o de asociación.
- `jsonb` y arreglos no almacenan relaciones principales, permisos, horarios o estados solo para evitar diseñar tablas.

### Segunda forma normal

- En una clave compuesta, todo atributo no clave depende de la clave completa.
- Una tabla de asociación contiene datos propios de la asociación, no atributos que pertenecen solo a uno de sus extremos.
- Las claves tenant-aware pueden ser compuestas por integridad sin obligar a repetir descripciones o configuración del tenant.

### Tercera forma normal

- Un atributo no clave depende de la identidad de la fila, no de otro atributo no clave.
- Zona horaria, configuración y datos de la barbería viven en `barbershop`; no se copian en cada barbero o cita como fuente de verdad.
- Los datos del servicio viven en `service`; una cita solo conserva snapshots aprobados para historia.
- Estados editables o con metadatos propios se modelan como entidad; vocabularios cerrados y estables pueden usar `CHECK`.

### Excepciones autorizadas para este proyecto

1. `barbershop_id` se repite en tablas de negocio y claves relacionadas para RLS, índices y FK compuestas. Es una duplicación deliberada de seguridad, no una fuente alternativa de la barbería.
2. La cita conserva snapshots de duración y precio, y de nombre si la historia visible lo exige, para que un cambio de servicio no reescriba el pasado (`DEC-004`). Se establecen al crear o aplicar un cambio explícito y no se sincronizan silenciosamente.
3. Contadores o vistas materializadas solo se incorporan después de medir una consulta y definir reconstrucción, consistencia y dueño.

Toda otra desnormalización documenta consulta que la requiere, volumen observado, costo de consistencia, mecanismo de reparación y prueba. “Evitar un join” no es justificación suficiente.

## 4. Convenciones de nombres

- Identificadores en inglés técnico, `snake_case`, minúscula y sin comillas.
- Tablas en singular: `appointment`, `time_block`, `notification_attempt`.
- Clave primaria: `id`.
- Clave foránea: `<tabla>_id`; tenant: `barbershop_id`.
- Instantes: `<evento>_at`; fechas locales: `<evento>_date`; booleanos: `is_*` o `has_*` cuando no exista un nombre de estado mejor.
- Restricciones: `<tabla>_<columnas>_<tipo>`, con sufijos `pk`, `fk`, `uk`, `ck`, `excl`.
- Índices: `idx_<tabla>_<columnas>` y, si es parcial, un sufijo que describa el predicado.
- Políticas: `<tabla>_<operacion>_<alcance>_policy`.
- Migraciones: timestamp ordenable generado por Atlas y descripción, por ejemplo `20260806143000_create_appointment.sql`.
- Evitar abreviaturas locales y nombres reservados como `user`, `order` o `group`.
- Mantener nombres por debajo del límite de PostgreSQL y suficientemente descriptivos.

Las palabras SQL se escriben en mayúscula y los identificadores en minúscula para lectura uniforme.

## 5. Tipos de datos

| Dato | Tipo preferido | Regla |
| --- | --- | --- |
| Identificador de dominio | `uuid` | no exponer secuencias predecibles en enlaces públicos |
| Instante real | `timestamptz` | almacenar el instante; mostrar en zona IANA de barbería |
| Fecha o hora civil de horario | `date`, `time` | interpretar junto con zona de la barbería |
| Duración o paso | entero en minutos | `CHECK (value > 0)` y rango de negocio |
| Dinero | `numeric` + moneda ISO | nunca `real` o `double precision` |
| Teléfono | `text` | normalizar a formato acordado; no tratar como número |
| Correo | `text` | normalizar para comparación sin perder valor de presentación cuando aplique |
| Estado cerrado | `text` + `CHECK` | no usar booleanos múltiples contradictorios |
| Contenido flexible | `jsonb` | solo con esquema de aplicación, motivo y política de indexación |

Reglas adicionales:

1. `NOT NULL` es el valor inicial; `NULL` solo representa ausencia significativa y documentada.
2. Cadena vacía, cero y fecha centinela no sustituyen `NULL`.
3. Toda longitud o rango importante se protege con `CHECK`, no solo con el formulario.
4. No usar `timestamp without time zone` para citas, intentos o auditoría.
5. No guardar hora local calculada si puede derivarse del instante y la zona, salvo snapshot justificado.
6. Contenido de proveedor en `jsonb` se minimiza y no conserva secretos ni datos personales innecesarios.

## 6. Claves, relaciones e integridad tenant-aware

- Toda tabla de negocio tiene `barbershop_id NOT NULL`.
- Cada tabla con ID global agrega `UNIQUE (barbershop_id, id)` cuando vaya a ser destino de una FK compuesta.
- Las relaciones de negocio incluyen el tenant en ambos lados:

```sql
FOREIGN KEY (barbershop_id, barber_id)
REFERENCES barber (barbershop_id, id)
```

- Una FK simple por `barber_id` no basta para demostrar que el barbero pertenece a la misma barbería.
- Toda FK define acción de borrado explícita. `CASCADE` se reserva para hijos sin vida ni retención propias.
- Citas, historial, intentos y auditoría no se borran en cascada desde una barbería o servicio.
- El borrado lógico se usa solo donde una regla lo exige; no se agrega `deleted_at` a todas las tablas.
- Las restricciones de unicidad dentro de una barbería incluyen `barbershop_id`.
- Crear índice en el lado referenciante de una FK cuando las consultas o borrados lo necesiten; PostgreSQL no lo crea automáticamente.

## 7. Restricciones de negocio en PostgreSQL

La base de datos debe proteger como mínimo:

- fin posterior al inicio;
- duración, rejilla, anticipación y valores configurables dentro de rangos válidos;
- estados permitidos;
- claves de idempotencia únicas en su alcance;
- tokens almacenados como hash y con unicidad apropiada;
- pertenencia tenant-aware de barbero, servicio, cliente y cita;
- ausencia de cruces para citas que ocupan agenda;
- historial sin actualización o borrado por el rol de aplicación;
- una sola programación lógica vigente por cita, tipo y canal cuando corresponda.

Se prefieren `NOT NULL`, `CHECK`, `UNIQUE`, FK y restricciones de exclusión a triggers. Un trigger solo se usa cuando la integridad no puede expresarse declarativamente y documenta orden, concurrencia y prueba.

Las reglas que cambian con frecuencia o requieren contexto del actor permanecen en el servicio de aplicación, con restricciones de base de datos que impidan corrupción estructural.

## 8. Tiempo, agenda y concurrencia

- Los intervalos son semiabiertos `[starts_at, ends_at)` en todas las capas.
- La exclusión usa `tstzrange` y el conjunto exacto de estados que ocupan agenda.
- Contigüidad es válida: una cita puede empezar cuando termina otra.
- La zona IANA se almacena en la barbería; cambios de zona no reinterpretan instantes históricos.
- Horarios recurrentes representan tiempo civil local y sus excepciones tienen fecha y zona inequívocas.
- El API no usa un bloqueo en memoria como defensa de agenda; múltiples procesos dependen de la misma restricción PostgreSQL.
- Una violación de exclusión se traduce a conflicto de negocio seguro.
- Los trabajadores reclaman lotes pequeños con bloqueo de fila y `SKIP LOCKED`, y vuelven a validar estado vigente.

## 9. Seguridad, RLS y privilegios

1. RLS se habilita y fuerza en toda tabla tenant-aware.
2. Las políticas definen `USING` y `WITH CHECK` según lectura y escritura.
3. El rol del API no es propietario, superusuario ni posee `BYPASSRLS`.
4. El tenant se fija con alcance local dentro de la transacción; nunca permanece en una conexión reutilizada.
5. Procesos administrativos usan rol separado, acceso mínimo, auditoría y ninguna exposición a handlers ordinarios.
6. Respaldos verifican que RLS no omita filas silenciosamente.
7. Las políticas simples que comparan columnas de la fila con el contexto son preferibles a subconsultas complejas.
8. Toda política se prueba con al menos dos barberías usando el rol real.
9. `search_path` se fija explícitamente para roles y funciones; no se confía en esquemas escribibles por usuarios no autorizados.
10. Funciones `SECURITY DEFINER` son excepcionales, fijan `search_path`, revocan ejecución pública y tienen revisión de seguridad.

## 10. Migraciones

### Reglas generales

- Atlas CLI y el flujo SQL versionado de [migraciones-atlas.md](migraciones-atlas.md) son obligatorios.
- Cada cambio de esquema existe como migración versionada; no se modifica producción manualmente.
- Una migración aplicada en un entorno compartido es inmutable. Una corrección crea otra migración.
- Cada versión usa un archivo SQL ascendente; no se crean parejas `.up.sql`/`.down.sql` ni se depende de reversión automática en producción.
- `atlas.sum` forma parte del cambio y debe coincidir con el contenido y orden del directorio.
- Una migración tiene una responsabilidad coherente y un nombre descriptivo.
- DDL transaccional se ejecuta en transacción. Operaciones incompatibles, como ciertos índices concurrentes, viven en una migración separada y explícitamente no transaccional.
- No mezclar una gran transformación de datos con muchos cambios estructurales sin puntos de validación.
- La promoción usa exactamente el mismo archivo y hash en local, CI, piloto y producción.

### Evolución segura

1. Preferir expandir, migrar datos, cambiar aplicación y contraer en una versión posterior.
2. Agregar columna requerida primero como nullable o con estrategia que evite reescribir/bloquear toda la tabla; rellenar por lotes, validar y después fijar `NOT NULL`.
3. No renombrar y eliminar una columna consumida en la misma versión desplegable.
4. Crear índices grandes de forma concurrente cuando el entorno tenga escrituras, considerando que no puede combinarse como DDL transaccional normal.
5. Revisar el bloqueo que toma cada `ALTER TABLE`; configurar límites operativos y abortar de forma segura antes de bloquear el servicio indefinidamente.
6. Una migración de datos declara reanudación, tamaño de lote, validación y comportamiento ante fallo.
7. Antes de una operación destructiva se verifica copia restaurable, conteos y plan de recuperación.
8. No eliminar datos personales para “reintentar” una migración fallida sin autorización y evidencia.

### Encabezado documental

Cada migración incluye comentarios breves con propósito, decisión o regla relacionada, condición de seguridad y plan de reversión/avance cuando no sea obvio. No duplica cada sentencia SQL.

## 11. Consultas, transacciones e índices

- Seleccionar columnas explícitas y aplicar límites.
- Toda consulta de negocio filtra por `barbershop_id`, aunque RLS también proteja.
- Mantener transacciones cortas; no llamar correo, WhatsApp u otra red dentro de ellas.
- Usar el nivel de aislamiento mínimo que preserve la regla y apoyarse en restricciones para carreras.
- Reintentar solo errores transitorios identificados y con operación idempotente.
- No concatenar SQL ni identificadores a partir de entrada; usar parámetros y listas permitidas.
- Un índice responde a una consulta o restricción concreta. Documentar consulta, orden de columnas y predicado parcial.
- Revisar `EXPLAIN (ANALYZE, BUFFERS)` con datos representativos en disponibilidad, agenda diaria, historial y cola de recordatorios.
- Eliminar índices redundantes solo después de verificar uso y efecto en restricciones.
- Mantener autovacuum y estadísticas; cambios de parámetros requieren medición.
- No paginar colecciones crecientes con offsets grandes si una clave estable permite paginación por cursor.

## 12. Datos personales, auditoría y retención

- Clasificar nombre, teléfono, correo, tokens y contenido de mensajes como sensibles.
- Guardar tokens públicos mediante hash; el valor original solo se entrega en la frontera necesaria.
- No almacenar contenido completo de notificación si basta plantilla, versión y metadatos mínimos.
- Auditoría conserva actor, instante, operación, entidad y cambios necesarios sin duplicar secretos.
- El historial es append-only para el rol de aplicación; una corrección crea otra entrada.
- La anonimización es idempotente, revoca tokens y conserva solo datos operativos autorizados por `DEC-025`.
- Datos vencidos no permanecen en tablas auxiliares, índices de búsqueda, colas o fixtures.
- Producción y piloto nunca se copian directamente a desarrollo.

## 13. Pruebas obligatorias de datos

- migraciones completas sobre base vacía;
- actualización desde la versión anterior con datos;
- `NOT NULL`, `CHECK`, `UNIQUE`, FK y exclusiones;
- RLS para `SELECT`, `INSERT`, `UPDATE` y `DELETE` con dos tenants;
- rol normal sin `BYPASSRLS` ni propiedad de tablas;
- relaciones cruzadas rechazadas incluso si se intenta SQL directo;
- concurrencia de reservas, bloqueos y trabajadores;
- contigüidad, cruce de medianoche y estados que liberan agenda;
- anonimización repetida sin pérdida de métricas autorizadas;
- plan y tiempo de consultas críticas;
- restauración de respaldo antes del piloto y después de cambios materiales.

El detalle de ejecución vive en [estrategia-pruebas.md](../03-desarrollo/estrategia-pruebas.md).

## 14. Lista de revisión de un cambio de base de datos

- [ ] El modelo cumple 1FN, 2FN y 3FN o documenta una excepción real.
- [ ] Nombres, tipos, `NULL` y valores iniciales expresan la semántica correcta.
- [ ] Tenant, FK compuestas, RLS y privilegios impiden relaciones o lecturas cruzadas.
- [ ] Restricciones protegen la invariante aun con dos procesos concurrentes.
- [ ] Acciones de borrado y retención son explícitas.
- [ ] La migración funciona en vacío y sobre la versión anterior con datos.
- [ ] Bloqueos, duración, reversión y recuperación fueron evaluados.
- [ ] Índices corresponden a consultas verificables y no duplican otros.
- [ ] No aparecen secretos o datos personales innecesarios.
- [ ] Existen pruebas de integración con PostgreSQL y rol real.
- [ ] Diagrama, diccionario, consultas o decisiones afectadas quedaron actualizados.

## 15. Referencias oficiales

- [Restricciones en PostgreSQL](https://www.postgresql.org/docs/current/ddl-constraints.html)
- [Políticas de seguridad por fila](https://www.postgresql.org/docs/current/ddl-rowsecurity.html)
- [Bloqueos explícitos](https://www.postgresql.org/docs/current/explicit-locking.html)
- [`CREATE INDEX`](https://www.postgresql.org/docs/current/sql-createindex.html)
- [Estructura léxica y nombres SQL](https://www.postgresql.org/docs/current/sql-syntax-lexical.html)
