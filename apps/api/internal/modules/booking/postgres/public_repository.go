package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

// createPublicAppointmentOperation identifica, para el protocolo de
// idempotencia (RN-IDE-01, DEC-043), la operación de HU-097: una clave ya
// usada para cualquier otra operación (incluida create_manual_appointment)
// no puede reutilizarse aquí (idempotency_begin responde
// OutcomeConflictOperation).
const createPublicAppointmentOperation idempotency.Operation = "create_public_appointment"

// createPublicAppointmentIdempotencyTTL es la vigencia de una reclamación
// OutcomeProceed sin completar todavía, mismo valor que
// createManualAppointmentIdempotencyTTL: HU-097 no documenta una razón de
// negocio para diferir de ese precedente.
const createPublicAppointmentIdempotencyTTL = 10 * time.Minute

// CreatePublic implementa booking.Repository.CreatePublic (HU-097): Begin,
// resuelve/crea `customer` según la decisión ya tomada por el llamador
// (DEC-085, sin repetir el Find dentro de la transacción -ver el
// comentario de la interfaz en booking/ports.go-), inserta la cita
// `confirmed` de origen `public`, el evento appointment_created con actor
// `customer` y el token de acceso (DEC-089), y Complete, todo dentro de UNA
// sola InTenantTx (mismo patrón que CreateManual).
func (r *Repository) CreatePublic(
	ctx context.Context,
	barbershopID string,
	input booking.CreatePublicInput,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (booking.CreatePublicResult, error) {
	shop, err := database.ValidBarbershopID(barbershopID)
	if err != nil {
		return booking.CreatePublicResult{}, fmt.Errorf("booking/postgres: %w", err)
	}

	var result booking.CreatePublicResult

	err = r.db.InTenantTx(ctx, shop, func(ctx context.Context, q database.Queries) error {
		decision, err := r.coord.Begin(ctx, q, shop, key, createPublicAppointmentOperation, fingerprint, createPublicAppointmentIdempotencyTTL)
		if err != nil {
			return fmt.Errorf("idempotency begin: %w", err)
		}
		result.Decision = decision

		if decision.Outcome != idempotency.OutcomeProceed {
			if decision.Outcome == idempotency.OutcomeReplay {
				result.Response = decision.Response
			}
			return nil
		}

		customer, err := resolvePublicCustomer(ctx, q, barbershopID, input.Customer, input.CustomerUpdatePhone, input.CustomerUpdateEmail)
		if err != nil {
			return err
		}

		internalInput := booking.CreateInternalInput{
			BarberID:     input.BarberID,
			ServiceID:    input.ServiceID,
			AttendeeName: input.AttendeeName,
			StartsAt:     input.StartsAt,
			EndsAt:       input.EndsAt,
			Origin:       booking.OriginPublic,
			Service:      input.Service,
			CustomerNote: input.CustomerNote,
		}

		appointment, err := insertAppointment(ctx, q, barbershopID, customer.ID, internalInput)
		if err != nil {
			return err
		}

		customerID := customer.ID
		actor := booking.Actor{Type: booking.ActorTypeCustomer, CustomerID: &customerID}
		if err := insertAppointmentCreatedHistory(ctx, q, barbershopID, appointment.ID, actor); err != nil {
			return err
		}

		if err := insertAppointmentAccessToken(ctx, q, barbershopID, appointment.ID, input.TokenHash, input.TokenIssuedAt, input.TokenExpiresAt); err != nil {
			return err
		}

		body, err := json.Marshal(newPublicAppointmentResponseWire(appointment, input))
		if err != nil {
			return fmt.Errorf("marshal confirmed public appointment: %w", err)
		}
		stored := idempotency.StoredResponse{
			Status:      201,
			ContentType: "application/json",
			Body:        string(body),
		}

		ok, err := r.coord.Complete(ctx, q, shop, key, stored)
		if err != nil {
			return fmt.Errorf("idempotency complete: %w", err)
		}
		if !ok {
			return fmt.Errorf("idempotency complete: la reclamación ya no estaba in_progress")
		}

		result.Response = stored
		return nil
	})
	if err != nil {
		return booking.CreatePublicResult{}, err
	}
	return result, nil
}

// publicAppointmentResponseWire es la forma exacta que CreatePublic
// serializa como cuerpo almacenado de idempotencia (idempotency_record.
// response_body): debe coincidir campo a campo con
// publicbookinghttpapi.ConfirmedPublicAppointmentResponse para que una
// repetición exacta reproduzca bytes idénticos a los de la respuesta
// original (mismo criterio que manualAppointmentResponseWire). accessToken
// viaja EN CLARO: es la única vez que existe fuera de esta transacción
// (DEC-089); idempotency_record hereda la misma vigencia corta
// (createPublicAppointmentIdempotencyTTL) que cualquier otra reclamación en
// curso, y una vez `completed` su retención sigue el mismo criterio que el
// resto de operaciones críticas del proyecto.
type publicAppointmentResponseWire struct {
	AttendeeName    string  `json:"attendeeName"`
	BarbershopName  string  `json:"barbershopName"`
	ServiceName     string  `json:"serviceName"`
	DurationMinutes int     `json:"durationMinutes"`
	PriceAmount     string  `json:"priceAmount"`
	Currency        string  `json:"currency"`
	StartsAt        string  `json:"startsAt"`
	EndsAt          string  `json:"endsAt"`
	Timezone        string  `json:"timezone"`
	AccessToken     string  `json:"accessToken"`
	CustomerNote    *string `json:"customerNote"`
}

func newPublicAppointmentResponseWire(a booking.Appointment, input booking.CreatePublicInput) publicAppointmentResponseWire {
	return publicAppointmentResponseWire{
		AttendeeName:    a.AttendeeName,
		BarbershopName:  input.BarbershopName,
		ServiceName:     a.ServiceNameSnapshot,
		DurationMinutes: a.DurationMinutesSnapshot,
		PriceAmount:     formatPriceAmount(a.PriceAmountCentsSnapshot),
		Currency:        a.CurrencySnapshot,
		StartsAt:        a.StartsAt.Format(time.RFC3339),
		EndsAt:          a.EndsAt.Format(time.RFC3339),
		Timezone:        input.Timezone,
		AccessToken:     input.TokenPlain,
		CustomerNote:    a.CustomerNote,
	}
}

// resolvePublicCustomer aplica la decisión de reconciliación de DEC-085 ya
// tomada por publicbooking.ReconcilePublicCustomer: reutiliza
// input.ExistingID (sobrescribiendo updatePhone/updateEmail cuando no son
// nil, el campo que NO participó en la coincidencia) o crea un cliente
// nuevo con input.New.
func resolvePublicCustomer(
	ctx context.Context,
	q database.Queries,
	barbershopID string,
	in booking.CustomerInput,
	updatePhone, updateEmail *string,
) (booking.Customer, error) {
	if in.ExistingID != nil {
		if updatePhone != nil || updateEmail != nil {
			return updateExistingCustomerContact(ctx, q, barbershopID, *in.ExistingID, updatePhone, updateEmail)
		}
		return loadExistingCustomer(ctx, q, barbershopID, *in.ExistingID)
	}
	return insertNewCustomer(ctx, q, barbershopID, *in.New)
}

// updateExistingCustomerContact aplica DEC-085 ("sobrescribir con el valor
// nuevo el campo que no participó en la coincidencia"): COALESCE dentro del
// UPDATE deja intacto el campo cuyo puntero llega nil.
func updateExistingCustomerContact(
	ctx context.Context,
	q database.Queries,
	barbershopID, customerID string,
	updatePhone, updateEmail *string,
) (booking.Customer, error) {
	var c booking.Customer
	err := q.QueryRow(ctx, `
		UPDATE customer
		   SET phone = COALESCE($3, phone),
		       email = COALESCE($4, email)
		 WHERE barbershop_id = $1 AND id = $2
		RETURNING id, barbershop_id, full_name, phone, email, created_at, updated_at`,
		barbershopID, customerID, updatePhone, updateEmail,
	).Scan(&c.ID, &c.BarbershopID, &c.FullName, &c.Phone, &c.Email, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return booking.Customer{}, errCustomerNotFound()
		}
		if isConstraintViolation(err, "23505", customerShopPhoneUniqueIndex) {
			return booking.Customer{}, errCustomerPhoneTaken()
		}
		if isConstraintViolation(err, "23505", customerShopEmailUniqueIndex) {
			return booking.Customer{}, errCustomerEmailTaken()
		}
		return booking.Customer{}, err
	}
	return c, nil
}

// insertAppointmentAccessToken persiste el token de acceso emitido dentro
// de la misma transacción de confirmación (DEC-089): solo el hash llega a
// PostgreSQL, nunca el valor en claro.
func insertAppointmentAccessToken(
	ctx context.Context,
	q database.Queries,
	barbershopID, appointmentID, tokenHash string,
	issuedAt, expiresAt time.Time,
) error {
	_, err := q.Exec(ctx, `
		INSERT INTO appointment_access_token (barbershop_id, appointment_id, token_hash, issued_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)`,
		barbershopID, appointmentID, tokenHash, issuedAt, expiresAt,
	)
	return err
}
