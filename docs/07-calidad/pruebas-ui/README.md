---
titulo: "Campañas de pruebas de interfaz con Luna"
version: "1.0"
estado: "Procedimiento operativo derivado"
ultima_actualizacion: "2026-10-07"
issue: 300
---
# Pruebas de la app real

Entrada: pedir **«Prueba toda la app»** o **«Prueba la pantalla de …»**. El coordinador carga [ui-app-testing](../../../.agents/skills/ui-app-testing/SKILL.md), prepara el entorno y ejecuta los encargos persistentes. No es necesario redactar los casos de nuevo.

El catálogo cubre las 17 pantallas implementadas en el checkout de #300, diálogos de bloqueos, perfil individual y recorridos entre pantallas. La UI ejecuta las acciones; API/PostgreSQL reales son el destino. Una respuesta simulada no demuestra guardado. No garantiza ausencia de bugs ni reemplaza las pruebas Go/SQL.

| Documento | Lectura |
| --- | --- |
| [Entorno y cuentas](01-entorno-cuentas.md) | Coordinador antes de delegar |
| [Protocolo y consumo](02-protocolo-agentes.md) | Todos; después solo su ficha |
| [Inventario y reparto](03-inventario.md) | Coordinador; agentes solo su fila |
| [Casos transversales](04-transversal.md) | Todos, según aplicabilidad |
| [Recorridos cruzados](05-recorridos.md) | Agente journeys y cierre |
| [Informe y evidencia](06-informe.md) | Todos al registrar |
| [Combinaciones](07-combinaciones.md) | Factores de su pantalla |
| [Cobertura y límites](08-cobertura.md) | Coordinador al cerrar |

Primera campaña: [#301](https://github.com/bcaceres19/barberia/issues/301), pendiente de ejecución. Kit: [#300](https://github.com/bcaceres19/barberia/issues/300). Campañas posteriores tienen issue real, SHA y prompts propios; no reutilizar un issue cerrado.

Los procedimientos son derivados de RN/DEC/HU. Si difieren, prevalece la norma y se registra la duda. Un issue/prompt/instrucción que asigna una referencia exacta activa fidelidad; una captura histórica no la asigna automáticamente. Prompts: [orquestación](../../10-backlog/prompts/orchestration/pruebas-ui-luna.md). La evidencia cruda es privada; solo una copia sanitizada se entrega públicamente.
