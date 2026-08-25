package catalog

import (
	"regexp"
	"time"
)

// barberIDPattern es la misma forma canónica 8-4-4-4-12 que
// database.ValidBarbershopID/staff.LooksLikeBarberID exigen (el núcleo no
// puede importar ninguno de esos dos paquetes: database.ValidBarbershopID
// por CA-002-06, staff.LooksLikeBarberID porque catalog no importa el
// núcleo de staff, mismo criterio de "colaboración solo por puerto
// explícito" de HU-023). Duplicar esta única expresión regular es la misma
// duplicación pequeña y clara que domain.go ya documenta para
// serviceIDPattern frente a staff.barberIDPattern.
var barberIDPattern = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
)

// LooksLikeBarberID informa si id tiene la forma de un UUID válido. Un
// identificador que no cumple esta forma no puede corresponder a ninguna
// fila real: se trata igual que "no existe" (mismo apperr.NotFound) en vez
// de dejar que la consulta a staff.BarberPort o a PostgreSQL falle con un
// error de tipo.
func LooksLikeBarberID(id string) bool {
	return barberIDPattern.MatchString(id)
}

// Assignment es la asociación PURA entre un barbero y un servicio de la
// misma barbería (HU-023, CA-023-07): exactamente BarberID, ServiceID y
// CreatedAt. Nunca nombre, duración, precio ni estado: esos campos siguen
// siendo responsabilidad exclusiva de staff.Barber y catalog.Service: un
// consumidor que necesite mostrarlos los obtiene de esas otras dos listas y
// cruza por identificador, nunca de este tipo.
type Assignment struct {
	BarberID  string
	ServiceID string
	CreatedAt time.Time
}

// AssignOutcome distingue, sin ambigüedad, qué pasó dentro de la
// transacción de Repository.Assign.
type AssignOutcome int

const (
	// AssignOutcomeCreated: la fila no existía y se insertó (CA-023-02).
	AssignOutcomeCreated AssignOutcome = iota
	// AssignOutcomeAlreadyExists: la fila ya existía (misma asignación
	// repetida); no se insertó una segunda (CA-023-02, semántica PUT
	// naturalmente repetible).
	AssignOutcomeAlreadyExists
	// AssignOutcomeServiceNotFound: serviceID no existe en esta barbería
	// (CA-023-04). La existencia/pertenencia del barbero ya se verificó
	// ANTES de esta llamada mediante BarberPort (AssignmentService.Assign);
	// esta variante solo cubre el servicio, propio de este módulo.
	AssignOutcomeServiceNotFound
)

// AssignResult es el desenlace completo de Repository.Assign.
type AssignResult struct {
	Outcome    AssignOutcome
	Assignment Assignment
}

// UnassignOutcome distingue, sin ambigüedad, qué pasó dentro de la
// transacción de Repository.Unassign.
type UnassignOutcome int

const (
	// UnassignOutcomeDeleted: la fila existía, DEC-068 lo permitía, y se
	// borró.
	UnassignOutcomeDeleted UnassignOutcome = iota
	// UnassignOutcomeNotFound: no existe tal asignación (barbero
	// inexistente/ajeno, servicio inexistente/ajeno, o la asociación
	// concreta nunca existió/ya se había retirado). Las tres causas
	// producen el mismo 404 uniforme (CA-023-04), sin distinguirse en el
	// resultado.
	UnassignOutcomeNotFound
	// UnassignOutcomeLastActiveConflict: la fila existe, el servicio está
	// activo, y es la última asignación activa de ese servicio (DEC-068,
	// CA-023-05/CA-023-06): se rechaza sin borrar nada.
	UnassignOutcomeLastActiveConflict
)

// UnassignResult es el desenlace completo de Repository.Unassign.
type UnassignResult struct {
	Outcome UnassignOutcome
}

// AssignmentListResult es una página de asignaciones de un barbero, ya
// ordenada de forma estable (created_at, service_id). NextCursor es "" para
// la última página (CA-023-01).
type AssignmentListResult struct {
	Items      []Assignment
	NextCursor string
}
