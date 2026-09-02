---
titulo: "Estándar de código del frontend en Vue"
version: "1.3"
estado: "Obligatorio para desarrollo"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-09-01"
documentos_relacionados:
  - "../00-control/registro-decisiones.md"
  - "../01-producto/alcance-mvp.md"
  - "../04-arquitectura/frontend.md"
  - "estandar-diseno-visual.md"
  - "especificacion-frontend-nava.md"
  - "estrategia-pruebas.md"
  - "../06-api/estandar-openapi.md"
---

# Estándar de código del frontend en Vue

## 1. Propósito y alcance

Estas reglas aplican a la aplicación Vue del flujo público y del panel del barbero. Buscan que una persona pueda entender, probar y modificar una funcionalidad sin recorrer toda la aplicación ni conocer detalles internos de otras funcionalidades.

Código limpio significa aquí comportamiento explícito, componentes cohesionados, contratos tipados, accesibilidad y dependencias contenidas. No significa fragmentar cada archivo ni convertir toda lógica en un patrón genérico.

Fuentes normativas: `DEC-035`, `DEC-039` y `DEC-077`.

## 2. Ubicación y estructura

El frontend vive en `apps/web`.

```text
apps/web/
  e2e/
  public/
  src/
    app/
      bootstrap/
      router/
    modules/
      auth/
      public-booking/
      agenda/
      catalog/
      staff/
      schedules/
      notifications/
      settings/
    shared/
      api/
      composables/
      model/
      ui/
      validation/
    styles/
  package.json
  tsconfig.json
  vite.config.ts
```

Una funcionalidad agrupa lo que cambia por la misma razón:

```text
modules/public-booking/
  api/
  components/
  composables/
  model/
  pages/
  validation/
  index.ts
```

No se crean carpetas vacías. Los archivos de prueba de unidad o componente se ubican junto al código como `*.spec.ts`; los recorridos completos viven en `e2e/`.

## 3. Límites y dirección de dependencias

```text
app → modules → shared
app → shared
```

1. `app` compone router, proveedores y arranque; no contiene reglas del negocio.
2. Un módulo puede importar `shared`, pero no los archivos internos de otro módulo.
3. Una capacidad compartida entre módulos expone una API pública mínima desde `index.ts` o se orquesta en `app`.
4. `shared` no importa `modules` ni contiene términos que pertenecen a una sola funcionalidad.
5. No existen carpetas genéricas `helpers`, `misc` o `utils`. Un elemento compartido conserva un nombre y una responsabilidad específicos.
6. Los alias de importación se usan de forma uniforme; no se crean cadenas frágiles de rutas relativas profundas.
7. Los archivos barril no reexportan árboles completos ni ocultan ciclos; la superficie pública es explícita.

## 4. Componentes, páginas y composables

### Componentes

- Un componente representa una responsabilidad visual o interacción nombrable.
- Las páginas coordinan carga, navegación y componentes; no concentran toda la lógica de negocio.
- Props y emits se tipan de forma explícita. El componente no modifica una prop ni un objeto propiedad del padre.
- Los nombres de componentes son descriptivos y en `PascalCase`: `AppointmentStatusBadge`, no `Card2` ni `CommonModal`.
- Los componentes base usan el prefijo acordado `Base`; los componentes de negocio permanecen dentro de su módulo.
- Los componentes consumen los tokens y variantes de [estandar-diseno-visual.md](estandar-diseno-visual.md), y las pantallas respetan arquitectura, estados y frontera de alcance de [especificacion-frontend-nava.md](especificacion-frontend-nava.md); una prop expresa intención y no acepta colores o medidas libres.
- Slots, eventos y estados visibles forman la API pública del componente y se mantienen pequeños.
- No se divide un componente solo por cantidad de líneas. Se divide cuando mezcla responsabilidades, repite una unidad o dificulta probarla.

### Composables

- Un composable encapsula lógica reactiva reutilizable o una interacción con ciclo de vida; no es un destino automático para toda función.
- Su nombre comienza con `use` y describe capacidad: `useAppointmentForm`.
- Expone el mínimo estado necesario y no devuelve un objeto interno mutable sin control.
- Una función pura de formato o validación vive en `model` o `validation`, no en un composable.
- Efectos, suscripciones y temporizadores se limpian al desmontar.

### Reactividad

- Derivar valores con `computed`; usar `watch` para efectos reales, no para mantener dos estados equivalentes sincronizados.
- Mantener el estado cerca de quien lo usa. Pinia solo se incorpora con estado compartido real entre rutas o sesiones de navegación.
- Evitar reactividad profunda sobre respuestas grandes; mapear la respuesta al modelo mínimo de la pantalla.
- Cada lista usa una clave estable del dominio, nunca el índice si los elementos pueden cambiar.
- El DOM se controla declarativamente; acceso manual solo cuando una API del navegador lo exige.

## 5. TypeScript y modelos

1. `strict` permanece habilitado; no se acepta `any` sin una frontera temporal documentada.
2. Datos externos entran como `unknown` y se validan antes de tratarlos como modelos confiables.
3. Usar uniones discriminadas para estados como carga, éxito, vacío y error.
4. Evitar aserciones `as` para ocultar un contrato incompleto; corregir o validar la fuente.
5. Preferir tipos inmutables para props, resultados y configuraciones.
6. Nombrar tipos por significado: `CreateAppointmentInput`, no `Payload` o `ResponseData`.
7. No duplicar manualmente contratos API. Los tipos derivados del bundle OpenAPI se aíslan en `shared/api/generated` y el código generado no se edita.
8. Separar DTO de transporte y modelo de vista cuando tengan semántica diferente.
9. Fechas e importes no se transportan como objetos implícitos: se parsean y formatean en una frontera conocida.
10. Los términos visibles usan “turno”; nombres técnicos de API y datos conservan `appointment`, según `DEC-016`.

## 6. Flujo de datos, API y errores

- Todo acceso HTTP pasa por el cliente tipado de `shared/api`; los componentes no llaman `fetch` directamente.
- Un módulo define funciones API por intención: `createAppointment`, no `postData`.
- La autoridad de disponibilidad, permisos y conflictos es el backend. El frontend solo anticipa errores para mejorar la experiencia.
- Toda operación asíncrona representa estados de carga, éxito, vacío y error cuando apliquen.
- Deshabilitar un botón no reemplaza una clave de idempotencia en operaciones críticas.
- Se cancelan solicitudes obsoletas cuando abandonar o repetir una búsqueda pueda producir resultados fuera de orden.
- Los errores de dominio conocidos se traducen a mensajes accionables; los errores inesperados muestran una salida segura y conservan `request_id` para soporte.
- No se almacenan tokens sensibles en registros, mensajes de consola ni almacenamiento persistente accesible sin necesidad.
- La vista no conoce la forma interna de errores de cada proveedor; usa el contrato estable del API.

## 7. Formularios, accesibilidad y rendimiento

1. Cada control tiene etiqueta accesible y mensaje de error asociado.
2. La navegación completa funciona con teclado y mantiene foco visible.
3. Diálogos administran foco, cierre y nombre accesible; no son un `div` con clic.
4. El color no es la única señal de estado.
5. Errores se muestran cerca del campo y el resumen mueve el foco cuando sea necesario.
6. La validación del navegador no sustituye la del backend.
7. Evitar doble envío y conservar campos no sensibles después de un error recuperable.
8. Cada ruta se carga de forma diferida y el flujo público no descarga el panel privado.
9. Una dependencia visual o de calendario requiere comparar tamaño, accesibilidad y alternativa nativa.
10. Las imágenes declaran dimensiones y tamaño apropiado; no se agregan animaciones que bloqueen la tarea principal.
11. La interfaz se verifica en el ancho y teléfono real del piloto; un emulador no es la única evidencia.
12. Colores, tipografía, espaciado, radios, sombras y tamaños proceden del sistema visual; un módulo no crea su propia paleta ni redefine una primitiva compartida.

## 8. Nombres, estilo y comentarios

- Componentes y tipos: `PascalCase`.
- Variables y funciones: `camelCase`.
- Composables: `useNombre`.
- Booleanos: `isLoading`, `hasConflict`, `canCancel`.
- Manejadores: `onSubmit`, `onStatusChanged`; evitar `handleThing` cuando puede expresarse la intención.
- Constantes verdaderamente globales: `UPPER_SNAKE_CASE`; no usar mayúsculas para cada `const` local.
- Archivos de componentes: `PascalCase.vue`; archivos TypeScript no componentes: nombre descriptivo en `kebab-case.ts`.

Reglas de documentación:

1. Props, eventos y tipos deben expresar el contrato antes de añadir comentarios.
2. Usar TSDoc/JSDoc en APIs compartidas cuando exista una precondición, efecto o caso límite no evidente.
3. Comentar por qué existe una excepción, un workaround o una regla temporal; no narrar el template.
4. Un `TODO` incluye referencia y condición de retiro.
5. Cada módulo complejo puede tener README corto con propósito, API pública y dependencias; no duplicar el documento de arquitectura.
6. No conservar código comentado: el historial de versiones será la fuente cuando se inicialice Git.
7. Ejemplos y fixtures usan datos ficticios y nunca información del piloto.

## 9. Controles automáticos

Cuando exista el scaffold, todo cambio frontend debe pasar:

```text
Prettier sin diferencias
ESLint sin errores
vue-tsc --noEmit
Vitest
build de producción
Playwright para los flujos exigidos por el cambio
```

- No se desactiva una regla de lint a nivel global para resolver un caso local.
- ESLint usa las reglas de Vue y TypeScript; Prettier se limita al formato y no sustituye el análisis estático.
- No se aprueba un error de TypeScript porque “funciona en el navegador”.
- El análisis de bundle forma parte de candidatos al piloto y de cambios de dependencia relevantes.

## 10. Lista de revisión del frontend

- [ ] La funcionalidad vive en su módulo y respeta la dirección de dependencias.
- [ ] Componentes, props, eventos y estados tienen nombres de dominio claros.
- [ ] No se introdujo `any`, estado global o dependencia sin necesidad demostrable.
- [ ] Carga, vacío, error, reintento y doble envío están resueltos donde aplican.
- [ ] La acción funciona con teclado, lector semántico y ancho móvil.
- [ ] La pantalla aplica tokens y tamaños del [estándar visual](estandar-diseno-visual.md), y composición, estados y alcance de la [especificación NAVA](especificacion-frontend-nava.md), con evidencia en los anchos exigidos.
- [ ] No se decidió disponibilidad ni permiso únicamente en el navegador.
- [ ] Se agregaron pruebas unitarias, de componente o E2E según [estrategia-pruebas.md](estrategia-pruebas.md).
- [ ] Los comentarios explican decisiones y no repiten la implementación.
- [ ] No se exponen datos personales, tokens ni detalles internos del API.
- [ ] El cambio conserva carga diferida y no aumenta el bundle sin justificación.

## 11. Referencias oficiales

- [Vue con TypeScript](https://vuejs.org/guide/typescript/overview.html)
- [Composables de Vue](https://vuejs.org/guide/reusability/composables.html)
- [Rendimiento en Vue](https://vuejs.org/guide/best-practices/performance)
- [Pruebas en Vue](https://vuejs.org/guide/scaling-up/testing.html)
