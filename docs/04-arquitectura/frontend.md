---
titulo: "Arquitectura del frontend"
version: "1.6"
estado: "Decisión confirmada"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-09-02"
documentos_relacionados:
  - "../00-control/registro-decisiones.md"
  - "../01-producto/alcance-mvp.md"
  - "../03-desarrollo/estandar-frontend-vue.md"
  - "../03-desarrollo/estandar-diseno-visual.md"
  - "../03-desarrollo/especificacion-frontend-nava.md"
  - "../03-desarrollo/estrategia-pruebas.md"
  - "../06-api/estandar-openapi.md"
  - "stack-despliegue-operacion.md"
---

# Arquitectura del frontend

## 1. Stack confirmado

| Pieza | Elección |
| --- | --- |
| Framework | Vue 3 |
| Lenguaje | TypeScript con modo estricto |
| Build y desarrollo | Vite |
| Estilo de componentes | Composition API y `<script setup>` |
| Navegación | Vue Router |
| Estado | Estado local y composables; Pinia solo si aparece estado global real |
| Comunicación | `fetch` encapsulado en un cliente tipado |
| Entrega | Aplicación web responsive con archivos estáticos versionados |

Fuente normativa: `DEC-033`.

## 2. Motivo de la elección

Vue 3 ofrece un punto medio adecuado para este proyecto:

- menos estructura obligatoria y herramientas iniciales que Angular;
- más decisiones integradas para construir una aplicación que React usado como biblioteca aislada;
- plantillas precompiladas y APIs que permiten tree-shaking con un build moderno;
- integración oficial con TypeScript y scaffolding basado en Vite;
- curva operativa apropiada para un único desarrollador.

La decisión busca reducir **tiempo total de construcción y mantenimiento**, no ganar un benchmark aislado. El rendimiento real se verificará en el teléfono de gama media definido por `SUP-008`.

## 3. Reglas para mantener bajo el consumo

1. Cargar cada ruta de forma diferida.
2. No instalar una dependencia para algo que pueda resolverse de forma clara con la plataforma web.
3. Revisar el impacto en el bundle antes de incorporar calendarios, editores o bibliotecas visuales.
4. Evitar una biblioteca de componentes completa durante el flujo público; priorizar componentes propios pequeños.
5. No usar Pinia para estado que pertenece a una sola pantalla.
6. No conservar listas completas si la API puede paginarlas.
7. Evitar reactividad profunda sobre estructuras grandes; normalizar los datos recibidos.
8. Servir imágenes en dimensiones adecuadas y no cargar recursos del área privada en el flujo público.
9. Generar bundles separados para el flujo público y el panel del barbero mediante rutas diferidas.
10. Medir tamaño, carga e interacción en cada build de candidato al piloto.
11. Todo rediseño o pantalla nueva conserva los mockups aprobados, la firma cromática y el contraste editorial/funcional de NAVA / Tailored Grid; la composición y la tecnología de estilos son libres y cada issue elige la solución que mejor equilibre identidad, accesibilidad, rendimiento y mantenimiento.
12. Una herramienta visual o dependencia nueva —incluido Tailwind, una biblioteca de componentes o iconos— requiere justificar licencia, mantenimiento, accesibilidad e impacto en el bundle.
13. Mantener separados los bundles del shell privado y la reserva pública; el wordmark compartido no autoriza importar navegación privada en el flujo del cliente.

## 4. Estructura inicial

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
```

Cada módulo agrupa vistas, componentes y lógica de su capacidad. `shared/` solo recibe elementos reutilizados de verdad; no se convierte en un depósito genérico.

## 5. Límites deliberados del MVP

- no aplicación móvil nativa;
- no funcionamiento completo sin conexión;
- no Nuxt ni renderizado del lado del servidor para el panel operativo;
- no microfrontends;
- no gestor de estado global obligatorio;
- no se agregan al MVP configuradores de tema, personalización por barbería ni otras capacidades de producto solo por una decisión visual; la implementación interna puede usar la expresión visual que mejor materialice NAVA / Tailored Grid.

Las páginas públicas de mercadeo pueden generarse como HTML estático separado. El flujo de reservas sigue siendo interactivo y se entrega con Vue.

## 6. Criterios de aceptación técnicos

- funciona con una sola mano en el teléfono real del participante;
- navegación inicial no descarga módulos privados que el cliente no usa;
- una conexión lenta muestra estado de carga y permite reintentar sin duplicar operaciones;
- cada formulario muestra errores por campo y conserva datos no sensibles;
- doble toque no crea dos citas;
- ninguna decisión de disponibilidad se toma solo en el navegador;
- los tipos del cliente HTTP derivan del bundle OpenAPI y no se duplican manualmente;
- el build de producción se analiza antes del piloto y toda dependencia grande queda justificada.

## 7. Estándares de implementación

La organización por funcionalidades, las reglas de TypeScript, componentes, documentación y accesibilidad son obligatorias según [estandar-frontend-vue.md](../03-desarrollo/estandar-frontend-vue.md). La dirección visual y las libertades de implementación siguen [estandar-diseno-visual.md](../03-desarrollo/estandar-diseno-visual.md); arquitectura de información, shell, inventario de capacidades y frontera P0/P1/P2 siguen [especificacion-frontend-nava.md](../03-desarrollo/especificacion-frontend-nava.md). Las pruebas unitarias, de componente y E2E se definen en [estrategia-pruebas.md](../03-desarrollo/estrategia-pruebas.md). Los DTO, errores y operaciones HTTP siguen [estandar-openapi.md](../06-api/estandar-openapi.md).

La organización interna detallada y la dirección de dependencias están en el estándar de desarrollo.

## 8. Referencias oficiales

- [Vue con TypeScript](https://vuejs.org/guide/typescript/overview.html)
- [Rendimiento en Vue](https://vuejs.org/guide/best-practices/performance)
- [Vite](https://vite.dev/guide/)
