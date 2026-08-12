---
titulo: "Prompts detallados de implementación · B0 en curso"
version: "1.1"
estado: "Propuesta"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-12"
documentos_relacionados:
  - "prompts-implementacion.md"
  - "../02-requisitos/historias-usuario.md"
  - "../03-desarrollo/estandar-backend-go.md"
  - "../03-desarrollo/estandar-frontend-vue.md"
  - "../03-desarrollo/estandar-diseno-visual.md"
  - "../03-desarrollo/estrategia-pruebas.md"
  - "../04-arquitectura/backend-go.md"
  - "prompts/README.md"
---

# Prompts detallados de implementación · B0 en curso

> **Estado: Propuesta.** Un prompt es una herramienta de trabajo, **no** una fuente
> normativa. Si contradice `AGENTS.md`, una regla `RN-*`, una decisión `DEC-*` o un estándar
> de `docs/03-desarrollo/`, manda el documento normativo y el prompt se corrige.

> **Catálogo vigente:** este archivo se conserva como antecedente detallado de B0. Todo prompt nuevo o revisado vive individualmente en el [catálogo de prompts persistentes](prompts/README.md), que registra issue, versión, dependencias y estado de ejecución.

---

## 1. Relación con `prompts-implementacion.md`

[prompts-implementacion.md](prompts-implementacion.md) contiene el **índice** de prompts de
las doce historias de B0 y el **preámbulo obligatorio**. Ese documento no se reemplaza.

Este archivo expande, para las historias que están **entrando en construcción ahora**, el
nivel de detalle que un prompt corto no alcanza: decisiones técnicas concretas, trampas
conocidas del lenguaje o del motor, y la forma exacta de verificar cada criterio.

Se expande una historia cuando se va a construir, no antes. Detallar por anticipado una
historia lejana produce texto que hay que reescribir.

| Historia | Área | Estado | Dónde está su prompt |
| --- | --- | --- | --- |
| `HU-001` | Base de datos | **Terminada** | `prompts-implementacion.md` §3 |
| `HU-002` | Backend | En construcción | **Sección 3 de este archivo** |
| `HU-009` | Frontend | En construcción | **Sección 4 de este archivo** |
| `HU-003`, `HU-004` | Backend | Siguientes | `prompts-implementacion.md` §3 |
| `HU-005` – `HU-008` | Backend | Bloqueadas | `prompts-implementacion.md` §3 y [prompt-llaves-asimetricas.md](prompt-llaves-asimetricas.md) |

### Por qué estas dos y en paralelo

`HU-002` es el paso siguiente estricto del orden de construcción de B0. `HU-009` es la única
historia del bloque cuya tabla de cabecera dice **"Depende de: —"**, y bloquea `HU-010`,
`HU-011`, `HU-012` y toda pantalla posterior.

Son las dos únicas historias que se pueden empezar hoy, y no comparten ni un archivo: una
vive en `apps/api`, la otra en `apps/web`. Trabajarlas en paralelo no crea conflicto.

---

## 2. Estado real del repositorio al escribir estos prompts

El agente que reciba estos prompts debe partir de hechos, no de suposiciones.

**Base de datos — lista y verificada.**

- Atlas CLI **v1.3.0** instalado y fijado. `latest` entrega compilaciones *canary*, que
  `migraciones-atlas.md` §3.3 prohíbe.
- `database/migrations/` tiene dos migraciones aplicadas y `atlas.sum` generado y validado.
- Existen las tablas `barbershop`, `staff_user` e `idempotency_record`, con RLS habilitada
  y forzada.
- Existen los roles `barberia_migrator` (propietario) y `barberia_app` (**sin** `BYPASSRLS`,
  sin propiedad, sin DDL).
- `database/testdata/dos_barberias.sql` carga las barberías `1111…1111` y `2222…2222`.
- `database/tests/hu001_aislamiento_rls.sql` verifica `CA-001-01` a `CA-001-06`.
- El resto del modelo está diseñado en `database/modelo-fisico-referencia.sql`, que **no es
  una migración**.

**Backend — andamiaje sin implementación.**

- Un módulo Go, `module system-barbershop`, Go 1.23. **Cero dependencias externas.**
- `internal/platform/config` es el único paquete autorizado a leer variables de entorno.
- `internal/platform/database/doc.go` existe con solo el comentario de paquete.
- `internal/platform/httpserver/` tiene `server.go`, `health.go`, `recover.go`,
  `requestid.go`. Chi **todavía no** está incorporado: llega con `HU-003`.

**Frontend — andamiaje sin componentes.**

- Vue 3, TypeScript, Vite, Vitest, Playwright, ESLint y Prettier ya configurados.
- Dependencias de producción: solo `vue` y `vue-router`.
- `src/styles/tokens.css` **ya existe** con color, espaciado, forma, profundidad y tamaños
  de control.
- `src/shared/ui/index.ts` está vacío (`export {}`) con un comentario que nombra los
  componentes previstos.
- `src/app/BaseStatusView.vue` es una pantalla provisional que se reemplaza más adelante.

---

## 3. Prompt detallado · `HU-002` — Contexto de barbería en cada solicitud

> Precede este prompt con el **preámbulo obligatorio** de
> [prompts-implementacion.md](prompts-implementacion.md) sección 2.

```text
Implementa HU-002 (docs/02-requisitos/historias-usuario.md).
HU-001 está TERMINADA: el esquema, los roles y las políticas RLS ya existen y están
verificados. No los rehagas ni los modifiques.

OBJETIVO
Que toda operación de datos ocurra dentro de una transacción con app.barbershop_id fijado
con alcance local, y que sea IMPOSIBLE escribir una consulta de negocio sin contexto. No
"difícil de olvidar": imposible por la forma de la API.

────────────────────────────────────────────────────────────────────────
1. DEPENDENCIA: EL DRIVER DE POSTGRESQL
────────────────────────────────────────────────────────────────────────
El proyecto no tiene ninguna dependencia externa todavía. Esta es la primera, y
estandar-backend-go.md §5.21 exige justificarla por escrito: necesidad, mantenimiento,
licencia, superficie transitiva y efecto en seguridad.

Evalúa jackc/pgx v5 frente a database/sql con lib/pq. Considera al menos:
  - pgx habla el protocolo nativo y maneja uuid, timestamptz, numeric y rangos sin
    conversiones a texto. Este proyecto usa los cuatro tipos de forma intensiva;
  - pgxpool ofrece el pool con hooks de adquisición y liberación que la sección 4 de este
    prompt necesita;
  - lib/pq está en modo mantenimiento desde hace años;
  - database/sql añade una capa de abstracción que este proyecto no usa, porque
    DEC-035 excluye ORM y las consultas son SQL visible.
Escribe la justificación en el comentario de paquete o en un ADR corto. Fija la versión
menor en go.mod.

NO agregues: ORM, generador de consultas, contenedor de inyección, ni una biblioteca de
migraciones. Atlas ya gobierna las migraciones y NO se ejecutan al arrancar (CA-001-07).

────────────────────────────────────────────────────────────────────────
2. CONFIGURACIÓN
────────────────────────────────────────────────────────────────────────
Extiende internal/platform/config con lo que el pool necesita, siguiendo el estilo del
Config existente: campos agregados solo cuando un componente los consume, valores por
defecto seguros para local, y error explícito cuando falte un valor obligatorio.

Como mínimo: URL de conexión, máximo y mínimo de conexiones, tiempo de vida y de inactividad
de una conexión, timeout de conexión y timeout de sentencia.

La URL de conexión es un secreto: el tipo que la contiene implementa String() redactando
la contraseña. Una URL completa en un log de arranque es una fuga de credenciales, y es
el error más común de esta capa.

La aplicación se conecta como barberia_app. NUNCA como barberia_migrator ni como
superusuario: si las pruebas pasan con un rol privilegiado, no prueban nada sobre RLS.
NO uses SET ROLE para simular el rol de aplicación en producción.

────────────────────────────────────────────────────────────────────────
3. FIJACIÓN DEL CONTEXTO — LEE ESTO CON ATENCIÓN
────────────────────────────────────────────────────────────────────────
Existe una trampa concreta y peligrosa aquí.

SET LOCAL NO ACEPTA PARÁMETROS. Esto NO compila como consulta parametrizada:

    SET LOCAL app.barbershop_id = $1        -- imposible

La tentación es concatenar el identificador en la cadena SQL. Eso introduce inyección SQL
en el punto exacto del que depende TODO el aislamiento entre barberías. Es el peor lugar
posible del sistema para esa clase de defecto.

Usa la función, que sí es parametrizable, con alcance local a la transacción:

    SELECT set_config('app.barbershop_id', $1::text, true)

El tercer argumento `true` significa "local a la transacción": revierte solo al terminar
la transacción, sin necesidad de RESET. Ese comportamiento es justamente lo que verifica
CA-002-03.

Nunca construyas esa sentencia por concatenación, ni siquiera "porque el valor ya es un
uuid validado". La regla es la forma de la llamada, no la confianza en el valor.

────────────────────────────────────────────────────────────────────────
4. LA API QUE HACE IMPOSIBLE OLVIDAR EL CONTEXTO
────────────────────────────────────────────────────────────────────────
Este es el corazón de la historia. CA-002-04 exige que no exista NINGUNA función exportada
que ejecute consultas de negocio sin recibir el contexto de barbería.

Diseña internal/platform/database para que la única forma de tocar datos sea una función
que recibe el identificador de barbería y entrega el ejecutor de consultas SOLO dentro del
alcance de una transacción ya configurada. Idea de forma, adáptala a lo que resulte claro:

    // El tipo del ejecutor NO se puede obtener de otra manera ni guardar fuera del
    // callback: el pool no lo expone.
    func (d *DB) InTenantTx(
        ctx  context.Context,
        shop BarbershopID,
        fn   func(ctx context.Context, q Queries) error,
    ) error

Reglas de diseño, todas verificables:
  a) El pool NO expone Query, QueryRow, Exec ni Begin como métodos exportados. Si un módulo
     puede obtener una conexión suelta, la restricción es decorativa.
  b) BarbershopID es un tipo propio, no un uuid.UUID desnudo y mucho menos un string. Un
     tipo distinto impide pasar por error el id de un usuario o de una cita
     (estandar-backend-go.md §5.10).
  c) La secuencia interna es: adquirir conexión -> BEGIN -> set_config local -> ejecutar el
     callback -> COMMIT, o ROLLBACK ante cualquier error o pánico.
  d) Si set_config falla, se hace ROLLBACK y se devuelve error SIN ejecutar el callback.
     Jamás se continúa con contexto vacío (CA-002-05).
  e) El ejecutor entregado al callback deja de ser válido al retornar. Si alguien lo guarda
     en un struct, la transacción ya estará cerrada: documenta la invariante en el comentario
     del tipo, porque el compilador no puede impedirlo.
  f) context.Context es el primer parámetro, no se guarda en structs y se propaga hasta
     PostgreSQL (estandar-backend-go.md §5.18).
  g) No inicies goroutines dentro de la transacción.
  h) Nada de llamadas de red dentro de la transacción (base-datos.md §4.19).

Defensa adicional recomendada: usa el hook de liberación del pool para verificar que la
conexión devuelta no conserva app.barbershop_id. Convierte CA-002-03 en una garantía de
tiempo de ejecución, no solo en una prueba.

────────────────────────────────────────────────────────────────────────
5. TIMEOUTS Y CANCELACIÓN
────────────────────────────────────────────────────────────────────────
Fija statement_timeout a nivel de conexión desde la configuración. Una consulta que se
cuelga sin límite bloquea una conexión del pool y degrada todo el proceso.

Verifica en una prueba que cancelar el context cancela la consulta en curso: los criterios
técnicos de backend-go.md §9 lo exigen explícitamente. Una forma directa es lanzar
pg_sleep con un context de vida corta y comprobar que retorna por cancelación y no por
haber dormido.

────────────────────────────────────────────────────────────────────────
6. COMPROBACIÓN DE SALUD
────────────────────────────────────────────────────────────────────────
Ya existe internal/platform/httpserver/health.go. Añade la verificación de conectividad con
timeout corto y propio.

La respuesta indica disponible o no disponible y NADA más: sin versión de PostgreSQL, sin
nombre de base, sin host, sin usuario, sin texto del error del driver. Un endpoint de salud
es público de hecho aunque no se documente; su mensaje de error es superficie de
reconocimiento gratuita para un atacante.

La salud NO abre una transacción de tenant: no tiene barbería. Es la única lectura
autorizada fuera del patrón, y su comentario debe decirlo.

────────────────────────────────────────────────────────────────────────
7. DIRECCIÓN DE DEPENDENCIAS (CA-002-06)
────────────────────────────────────────────────────────────────────────
El dominio y los servicios no importan el paquete de base de datos ni tipos de pgx.

Escribe una prueba que lo verifique de forma automática, no un acuerdo verbal: recorre los
paquetes bajo internal/modules/ y falla si alguno importa internal/platform/database o
github.com/jackc/pgx. Puedes usar go/packages o parsear la salida de `go list -deps -json`.

Esta prueba existe para el import que alguien agregará dentro de seis meses con prisa. Debe
fallar con un mensaje que diga qué paquete importó qué y por qué está prohibido.

────────────────────────────────────────────────────────────────────────
8. PRUEBAS OBLIGATORIAS
────────────────────────────────────────────────────────────────────────
Integración con PostgreSQL REAL. Nada de SQLite ni de dobles: estrategia-pruebas.md §2 lo
prohíbe expresamente porque no reproducen RLS.

Prepara el esquema aplicando las migraciones con Atlas v1.3.0 sobre una base efímera, y
carga database/testdata/dos_barberias.sql. Barbería A = 11111111-1111-1111-1111-111111111111
con 2 usuarios; barbería B = 22222222-2222-2222-2222-222222222222 con 2 usuarios.
Conéctate como barberia_app.

Casos exigidos:
  1. CA-002-01 — Toda operación ocurre en transacción con contexto local. Verifica dentro
     del callback que current_setting('app.barbershop_id') devuelve el valor esperado.
  2. CA-002-02 — Dos barberías concurrentes sobre el MISMO pool. Lanza N goroutines
     alternando A y B, con varias iteraciones, y comprueba que ninguna observó filas de la
     otra. Ejecútalo con -race. Con una sola iteración esta prueba no demuestra nada:
     necesita concurrencia real y repetición.
  3. CA-002-03 — Residuo de contexto. Tras cerrar la transacción, adquiere conexiones del
     pool hasta cubrir el tamaño máximo y comprueba en cada una que
     current_setting('app.barbershop_id', true) es NULL o vacío. El segundo argumento true
     evita que la función lance error cuando el ajuste no existe: sin él, la prueba fallaría
     por la excepción en lugar de comprobar lo que quiere comprobar.
  4. CA-002-05 — Fallo al fijar el contexto. Fuerza el fallo (por ejemplo con un
     identificador inválido) y comprueba que el callback NUNCA se ejecutó y que la
     transacción quedó revertida.
  5. CA-002-06 — La prueba estructural de imports de la sección 7.
  6. Cancelación de context, según la sección 5.
  7. Que el rol usado en las pruebas NO tiene BYPASSRLS. Si alguien cambia el DSN a un
     superusuario, todas las pruebas de aislamiento pasarían siendo falsas. Esta
     comprobación protege a las demás.

────────────────────────────────────────────────────────────────────────
9. DOCUMENTACIÓN
────────────────────────────────────────────────────────────────────────
Crea o actualiza apps/api/README.md con el patrón obligatorio: cómo se abre una operación
tenant-aware, por qué no existe una vía alternativa, y qué hacer con la única excepción
autorizada (la comprobación de salud). Incluye el ejemplo de set_config y la advertencia de
por qué no se concatena.

────────────────────────────────────────────────────────────────────────
10. FUERA DE ALCANCE
────────────────────────────────────────────────────────────────────────
  - Autenticación y sesión: son HU-005 y HU-006, y están bloqueadas por DP-SEG-04. En esta
    historia el identificador de barbería llega como PARÁMETRO y las pruebas lo inyectan.
  - Chi y los grupos de rutas: llegan con HU-003.
  - Cualquier repositorio de negocio: no existen tablas de negocio todavía.
  - Migraciones: HU-001 ya las hizo. No crees ninguna.

Al terminar, enumera CA-002-01 a CA-002-06 y di con qué prueba concreta se verifica cada
uno. Un criterio sin prueba nombrada es un criterio incumplido.
```

---

## 4. Prompt detallado · `HU-009` — Sistema visual base en componentes

> Precede este prompt con el **preámbulo obligatorio** de
> [prompts-implementacion.md](prompts-implementacion.md) sección 2.

```text
Implementa HU-009 (docs/02-requisitos/historias-usuario.md).
No depende de ninguna historia y no toca el backend. Puede construirse en paralelo con
HU-002.

OBJETIVO
Que existan los componentes base que las pantallas de B0 van a usar, y que ninguna pantalla
posterior necesite inventar un estilo propio.

────────────────────────────────────────────────────────────────────────
0. HALLAZGO QUE DEBES RESOLVER PRIMERO
────────────────────────────────────────────────────────────────────────
apps/web/src/styles/tokens.css YA EXISTE y define color, espaciado, forma, profundidad,
capas y tamaños de control. Consúmelo; no lo recrees.

PERO le falta la escala tipográfica. tokens.css solo declara --font-family-base, mientras
que estandar-diseno-visual.md §5.2 define ocho roles con tamaño, alto de línea y peso:
display 36/44, h1 30/38, h2 24/32, h3 20/28, body-lg 18/28, body 16/24, body-sm 14/20 y
caption 12/16.

Sin esos tokens es IMPOSIBLE cumplir CA-009-01, que prohíbe cualquier tamaño en píxeles
literal dentro de un componente.

Antes de escribir el primer componente:
  a) agrega a tokens.css los tokens de tamaño, alto de línea y peso de los ocho roles,
     con los valores exactos de §5.2. No inventes valores intermedios ni redondees;
  b) respeta el comentario de cabecera del archivo, que exige actualizar el documento
     normativo al tocar un valor. Aquí no cambias ningún valor: trasladas al código una
     tabla que ya está decidida. Dilo así en el pull request;
  c) si al implementar descubres que falta OTRO token, no lo resuelvas con un literal:
     agrégalo a tokens.css y justifícalo. Ese es exactamente el mecanismo que el estándar
     define para las excepciones.

────────────────────────────────────────────────────────────────────────
1. ALCANCE EXACTO: CINCO COMPONENTES, NI UNO MÁS
────────────────────────────────────────────────────────────────────────
En apps/web/src/shared/ui/ crea SOLO los que HU-010, HU-011 y HU-012 van a usar:

    BaseButton   BaseInput   BaseAlert   BaseBadge   BaseDialog

El comentario actual de shared/ui/index.ts menciona además BaseSelect y BaseIconButton. NO
los construyas: ninguna pantalla de B0 los necesita todavía y la historia dice
explícitamente "No se crean componentes sin uso real". Deja el comentario actualizado
indicando que llegan con la pantalla que los use.

Un componente sin consumidor real es código que se diseña a ciegas y se rehace al primer
uso.

────────────────────────────────────────────────────────────────────────
2. CONTRATO DE CADA COMPONENTE
────────────────────────────────────────────────────────────────────────
Regla transversal: las props expresan INTENCIÓN, nunca valores. `tone="danger"` sí;
`color="#b91c1c"` no. Un componente que acepte un color libre rompe DEC-039 y hace inútil
todo el sistema de tokens.

BaseButton — estandar-diseno-visual.md §8.1
  - variantes: primary, secondary, text, danger. Las cuatro, no más;
  - estados obligatorios: default, hover, active, focus-visible, loading, disabled;
  - en carga MANTIENE EL ANCHO y cambia el texto a una acción explícita ("Confirmando…").
    Un botón que se encoge al cargar desplaza el layout bajo el dedo del usuario;
  - en carga bloquea envíos repetidos, y el comentario debe decir que esto NO sustituye la
    idempotencia del servidor (RN-IDE-01). Es una mejora de experiencia, no una garantía;
  - altura 44 px; 48 px para el primario en móvil (§6.2);
  - tipo de botón explícito; nunca un div con click;
  - no hagas que un botón parezca enlace ni al revés.

BaseInput — estandar-diseno-visual.md §8.2
  - orden estricto: etiqueta, ayuda opcional, control, mensaje de error o confirmación;
  - la etiqueta es VISIBLE y va encima. El placeholder no sustituye la etiqueta jamás;
  - el error se asocia al control de forma accesible y el control se marca como inválido;
  - el estado de error usa borde de énfasis de 2 px, icono y texto: el color no puede ser la
    única señal (§7.4 del estándar frontend);
  - se marca "Opcional" cuando aplique, en lugar de llenar el formulario de asteriscos;
  - altura 44 px, radio md;
  - el identificador del control y el del mensaje se generan de forma única por instancia:
    dos inputs en la misma pantalla no pueden compartirlos.

BaseAlert — estandar-diseno-visual.md §8.4
  - tonos: info, success, warning, danger;
  - admite una acción de recuperación opcional ("Reintentar");
  - los mensajes críticos NO desaparecen solos;
  - los errores de campo NO se muestran aquí: van junto al campo;
  - soporta mostrar un request_id copiable para soporte (§8.4, último punto);
  - el rol accesible corresponde a la gravedad, para que un lector de pantalla anuncie un
    error sin que el usuario tenga que buscarlo.

BaseBadge — estandar-diseno-visual.md §8.4 y §4.4
  - etiqueta un estado; NO es un botón;
  - los estados de turno ya tienen tokens dedicados en tokens.css:
    --color-status-confirmed-*, --color-status-completed-*,
    --color-status-cancelled-customer-*, --color-status-cancelled-barber-*,
    --color-status-no-show-*. Úsalos; no derives colores por tu cuenta;
  - el texto visible va en español y usa "turno" (DEC-016): Confirmada, Atendida, Cancelada
    por el cliente, Cancelada por la barbería, No asistió;
  - radio pill; el color nunca es la única señal.

BaseDialog — estandar-diseno-visual.md §8.5 y §12
  - título descriptivo, consecuencia, acción primaria y cancelación VISIBLE;
  - una acción destructiva nombra el objeto: "Cancelar turno", nunca "Aceptar";
  - el foco entra al abrir, queda contenido dentro, y VUELVE al elemento que lo disparó al
    cerrar. Esto no es opcional ni "se mejora después": un diálogo que pierde el foco deja
    la aplicación inutilizable con teclado;
  - cierra con Escape y con la acción de cancelar;
  - usa --layer-dialog; ningún componente inventa un z-index;
  - evalúa el elemento <dialog> nativo antes de construir el manejo de foco a mano: el
    navegador ya resuelve la contención de foco y la capa superior. Si lo descartas, escribe
    por qué;
  - en móvil puede presentarse como panel inferior, conservando la misma semántica y sin
    depender de arrastrar.

────────────────────────────────────────────────────────────────────────
3. PATRONES DE ESTADO DE PANTALLA — §9 DEL ESTÁNDAR VISUAL
────────────────────────────────────────────────────────────────────────
La historia exige los ocho patrones: inicial, carga, actualización, vacío, error
recuperable, error de campo, conflicto y éxito.

Cuidado con el alcance: NO son ocho componentes nuevos. Algunos son composición de los
cinco componentes y otros son un patrón documentado. Resuélvelo así:
  - error de campo -> ya vive en BaseInput;
  - error recuperable, conflicto, éxito, info -> son tonos y composición de BaseAlert;
  - carga inicial -> un esqueleto con la geometría aproximada. Si necesitas un componente,
    créalo, pero solo si HU-010 a HU-012 lo van a usar;
  - actualización -> el contenido anterior permanece con indicador local, NUNCA un
    bloqueador a pantalla completa;
  - vacío e inicial -> patrón de composición, documentado con ejemplo.

Los esqueletos respetan prefers-reduced-motion y no simulan contenido que la pantalla nunca
tendrá.

Documenta los ocho patrones con un ejemplo mínimo cada uno, para que HU-010 los aplique sin
reinterpretarlos.

────────────────────────────────────────────────────────────────────────
4. ACCESIBILIDAD E INTERACCIÓN
────────────────────────────────────────────────────────────────────────
  - Foco visible: anillo exterior de 2 px con --color-focus y separación de 2 px (§12).
    Usa focus-visible, no focus, para no mostrar el anillo al hacer clic con el ratón;
  - todo componente interactivo es operable SOLO con teclado (CA-009-02);
  - el orden de tabulación coincide con el orden visual;
  - área táctil mínima real de 44 × 44 px: el área interactiva no se reduce al dibujo del
    icono;
  - hover nunca es el único modo de revelar una acción necesaria;
  - transiciones, si existen, duran entre 120 y 200 ms y desaparecen por completo con
    prefers-reduced-motion (CA-009-07);
  - contraste WCAG 2.2 AA verificado, incluidos los bordes de control.

────────────────────────────────────────────────────────────────────────
5. TIPOS Y ESTILO DE CÓDIGO
────────────────────────────────────────────────────────────────────────
  - Composition API con <script setup lang="ts">;
  - props y emits tipados explícitamente. Prohibido `any` (estandar-frontend-vue.md §5.1);
  - las variantes se tipan como uniones literales, no como string. `tone: 'info' | 'success'
    | 'warning' | 'danger'` hace que un valor inválido falle en vue-tsc, no en producción;
  - el componente no modifica una prop ni un objeto del padre;
  - archivos PascalCase.vue; index.ts exporta la superficie pública explícita, sin
    reexportar árboles completos;
  - shared/ui NO importa nada de modules/ (dirección de dependencias §3);
  - estilos con <style scoped>, consumiendo únicamente variables de tokens.css;
  - los textos visibles van en español y usan "turno", no "cita" (DEC-016).

────────────────────────────────────────────────────────────────────────
6. PRUEBAS OBLIGATORIAS — estrategia-pruebas.md §5.2
────────────────────────────────────────────────────────────────────────
Vitest con Vue Test Utils, observando desde la perspectiva del usuario. Usa locators por
rol y por etiqueta accesible, no por clase CSS ni por estructura interna.

Por componente, como mínimo:
  - BaseButton: cada variante se renderiza; disabled y loading no emiten click; el ancho no
    cambia entre reposo y carga; es alcanzable y activable con teclado;
  - BaseInput: la etiqueta está asociada al control; el mensaje de error se anuncia y se
    vincula; el estado inválido se refleja de forma accesible; dos instancias en la misma
    pantalla no comparten identificadores;
  - BaseAlert: cada tono; la acción de recuperación emite su evento; el rol accesible
    corresponde a la gravedad;
  - BaseBadge: los cinco estados de turno muestran el texto español correcto y aplican su
    token;
  - BaseDialog: el foco entra al abrir, no escapa con Tab ni con Shift+Tab, vuelve al
    disparador al cerrar, y Escape cierra.

Prohibido: afirmar sobre `ref` internos, métodos privados o clases CSS, salvo que la clase
SEA el contrato visual. Los snapshots no valen como única evidencia (§5.2 del estándar).

Verificación accesible automatizada más revisión manual con teclado.

Evidencia visual en 320, 360, 768 y 1280 px, con los estados afectados: foco, carga, error,
conflicto, éxito e inactivo. Ninguna vista produce desplazamiento horizontal (CA-009-06).

Antes de entregar deben pasar: Prettier sin diferencias, ESLint sin errores,
vue-tsc --noEmit, Vitest y el build de producción.

────────────────────────────────────────────────────────────────────────
7. PROHIBIDO POR DEC-039
────────────────────────────────────────────────────────────────────────
  - modo oscuro y selector de tema;
  - colores por barbería;
  - CSS libre fuera de los tokens;
  - biblioteca visual externa (nada de Vuetify, PrimeVue, Tailwind ni similares);
  - emoji como iconografía de producto;
  - una dependencia de iconos sin justificar tamaño, licencia y alternativa nativa. Para
    los pocos iconos de B0, un SVG en línea bajo shared/ui es suficiente: caja de 24 px,
    trazo de 2 px, decorativos ocultos del árbol accesible.

────────────────────────────────────────────────────────────────────────
8. CRITERIO REAL DE TERMINADO
────────────────────────────────────────────────────────────────────────
La historia dice: "Terminado cuando las pantallas de HU-010 a HU-012 se construyen sin
agregar un solo estilo local nuevo."

No puedes verificarlo todavía porque esas pantallas no existen. Lo que SÍ debes hacer es
recorrer sus plantillas en estandar-diseno-visual.md §10 —Acceso, Recuperación y el
cascarón privado— y confirmar por escrito que los cinco componentes y los ocho patrones
cubren lo que esas tres pantallas necesitan. Si detectas un hueco, dilo en el pull request
en lugar de construir un sexto componente por si acaso.

Al terminar, enumera CA-009-01 a CA-009-07 y di con qué prueba o evidencia concreta se
verifica cada uno.
```

---

## 5. Qué hacer al cerrar estas dos historias

1. Marcar `HU-002` y `HU-009` en la matriz de trazabilidad y en el historial de cambios.
2. Expandir en este archivo el prompt de `HU-003`, que es la siguiente del orden de
   construcción y desbloquea `HU-004`.
3. `HU-010` no puede empezar aunque `HU-009` termine: depende de `HU-005`, bloqueada por
   `DP-SEG-04`. Ver [prompt-llaves-asimetricas.md](prompt-llaves-asimetricas.md).
