package idempotency

import (
	"context"
	"fmt"
	"time"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/database"
)

// SQLCoordinator es el adaptador PostgreSQL concreto de Coordinator: llama
// exactamente a idempotency_begin, idempotency_complete e idempotency_abort
// (database/migrations/20260811154100_harden_idempotency_concurrency.sql),
// sin duplicar en Go ninguna decisión que esas funciones ya garantizan de
// forma atómica (bloqueo consultivo, transición de estado, límites de
// tamaño). No conserva estado propio: toda la coordinación vive en
// PostgreSQL, por eso el valor cero ya es válido.
type SQLCoordinator struct{}

// NewSQLCoordinator construye el adaptador PostgreSQL.
func NewSQLCoordinator() SQLCoordinator { return SQLCoordinator{} }

var _ Coordinator = SQLCoordinator{}

func (SQLCoordinator) Begin(
	ctx context.Context,
	q database.Queries,
	shop database.BarbershopID,
	key Key,
	operation Operation,
	fingerprint Fingerprint,
	ttl time.Duration,
) (Decision, error) {
	if !fingerprint.valid() {
		// Defensa contra un Fingerprint construido fuera de
		// ComputeFingerprint (ver su comentario): nunca debería ocurrir en
		// código correcto, así que esto es un error del llamador, no una
		// entrada de cliente a reportar como tal.
		return Decision{}, apperr.Internal(
			fmt.Errorf("idempotency: fingerprint con formato inválido: %q", string(fingerprint)))
	}

	ttlSeconds := int(ttl.Round(time.Second) / time.Second)

	var (
		outcomeText string
		respStatus  *int16
		respType    *string
		respBody    *string
	)
	row := q.QueryRow(ctx,
		`SELECT outcome, response_status, response_content_type, response_body
		 FROM idempotency_begin($1, $2, $3, $4, $5)`,
		string(shop), string(key), string(operation), string(fingerprint), ttlSeconds,
	)
	if err := row.Scan(&outcomeText, &respStatus, &respType, &respBody); err != nil {
		return Decision{}, fmt.Errorf("idempotency: begin: %w", err)
	}

	outcome, err := parseOutcome(outcomeText)
	if err != nil {
		return Decision{}, err
	}

	decision := Decision{Outcome: outcome}
	if outcome == OutcomeReplay {
		if respStatus == nil || respType == nil || respBody == nil {
			return Decision{}, apperr.Internal(
				fmt.Errorf("idempotency: begin devolvió replay sin respuesta almacenada completa"))
		}
		decision.Response = StoredResponse{
			Status:      int(*respStatus),
			ContentType: *respType,
			Body:        *respBody,
		}
	}
	return decision, nil
}

func (SQLCoordinator) Complete(
	ctx context.Context,
	q database.Queries,
	shop database.BarbershopID,
	key Key,
	response StoredResponse,
) (bool, error) {
	var ok bool
	row := q.QueryRow(ctx,
		`SELECT idempotency_complete($1, $2, $3, $4, $5)`,
		string(shop), string(key), response.Status, response.ContentType, response.Body,
	)
	if err := row.Scan(&ok); err != nil {
		return false, fmt.Errorf("idempotency: complete: %w", err)
	}
	return ok, nil
}

func (SQLCoordinator) Abort(
	ctx context.Context,
	q database.Queries,
	shop database.BarbershopID,
	key Key,
) (bool, error) {
	var ok bool
	row := q.QueryRow(ctx,
		`SELECT idempotency_abort($1, $2)`,
		string(shop), string(key),
	)
	if err := row.Scan(&ok); err != nil {
		return false, fmt.Errorf("idempotency: abort: %w", err)
	}
	return ok, nil
}

// parseOutcome traduce el texto que devuelve idempotency_begin al pequeño
// catálogo cerrado Outcome. Un valor no reconocido significa que la
// migración cambió el protocolo sin que este adaptador se actualizara: se
// reporta como error interno, nunca como un Outcome inventado.
func parseOutcome(text string) (Outcome, error) {
	switch text {
	case "proceed":
		return OutcomeProceed, nil
	case "replay":
		return OutcomeReplay, nil
	case "conflict_operation":
		return OutcomeConflictOperation, nil
	case "conflict_fingerprint":
		return OutcomeConflictFingerprint, nil
	case "conflict_in_progress":
		return OutcomeConflictInProgress, nil
	case "locked":
		return OutcomeLocked, nil
	default:
		return 0, apperr.Internal(fmt.Errorf("idempotency: outcome desconocido de idempotency_begin: %q", text))
	}
}
