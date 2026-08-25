// Package schedule administra horarios, excepciones, festivos y bloqueos,
// según docs/03-desarrollo/estandar-backend-go.md.
//
// HU-040 agrega la primera capacidad: working_hour, el tramo recurrente de
// jornada laboral de un barbero por día ISO de la semana. Excepciones por
// fecha/festivos (HU-041) y bloqueos (HU-042) se agregan a este mismo
// paquete cuando existan; ninguna de las dos vive aquí todavía.
package schedule
