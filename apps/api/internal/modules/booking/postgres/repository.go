// Package postgres traduce el núcleo booking hacia PostgreSQL, según
// docs/03-desarrollo/estandar-backend-go.md. Repository es la única pieza
// que conoce pgx, database.DB y los nombres reales de tablas/constraints de
// database/migrations/20260827110000_create_appointment_core.sql.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

// Nombres de las restricciones de 20260827110000_create_appointment_core.sql
// que un exclusion_violation/foreign_key_violation/unique_violation puede
// reportar. Solo estos se traducen a un error tipado del dominio; cualquier
// otra violación se propaga sin traducir (defecto de forma que el dominio
// ya debió rechazar antes de llegar aquí).
const (
	appointmentBarberIntervalExclConstraint = "appointment_barber_interval_excl"
	appointmentCustomerFKConstraint         = "appointment_barbershop_id_customer_id_fk"
	appointmentBarberFKConstraint           = "appointment_barbershop_id_barber_id_fk"
	appointmentServiceFKConstraint          = "appointment_barbershop_id_service_id_fk"
	customerShopPhoneUniqueIndex            = "idx_customer_shop_phone"
	customerShopEmailUniqueIndex            = "idx_customer_shop_email"
)

// createManualAppointmentOperation identifica, para el protocolo de
// idempotencia (RN-IDE-01, DEC-043), la operación de HU-061: una clave ya
// usada para crear un tramo, una excepción o cualquier otra operación no
// puede reutilizarse aquí (idempotency_begin responde
// OutcomeConflictOperation).
const createManualAppointmentOperation idempotency.Operation = "create_manual_appointment"

// createManualAppointmentIdempotencyTTL es la vigencia de una reclamación
// OutcomeProceed sin completar todavía, mismo valor que
// createWorkingHourIdempotencyTTL (schedule/postgres) y sin ninguna razón
// de negocio para diferir de ese precedente.
const createManualAppointmentIdempotencyTTL = 10 * time.Minute

// Repository implementa booking.Repository contra PostgreSQL real.
type Repository struct {
	db    *database.DB
	coord idempotency.Coordinator
}

// New construye el repositorio a partir del pool tenant-aware compartido y
// el coordinador de idempotencia real (mismo patrón que
// schedulepostgres.New/catalogpostgres.New).
func New(db *database.DB, coord idempotency.Coordinator) *Repository {
	return &Repository{db: db, coord: coord}
}

func isConstraintViolation(err error, code, constraint string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == code && pgErr.ConstraintName == constraint
}

// isDeadlockDetected reconoce el SQLSTATE 40P01 (deadlock_detected) que
// PostgreSQL real puede reportar para la transacción perdedora de dos
// INSERT concurrentes que se disputan el mismo índice GiST de
// appointment_barber_interval_excl (ver el comentario en insertAppointment).
func isDeadlockDetected(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "40P01"
}

// CreateInternal implementa booking.Repository.CreateInternal: crea o
// vincula el cliente que input.Customer ya decidió, inserta la cita
// `confirmed` y el evento appointment_created, todo dentro de UNA sola
// InTenantTx (§2.3 del prompt de HU-060). Un fallo en cualquier paso
// revierte los tres.
func (r *Repository) CreateInternal(
	ctx context.Context,
	barbershopID string,
	input booking.CreateInternalInput,
) (booking.CreateInternalResult, error) {
	shop, err := database.ValidBarbershopID(barbershopID)
	if err != nil {
		return booking.CreateInternalResult{}, fmt.Errorf("booking/postgres: %w", err)
	}

	var result booking.CreateInternalResult

	err = r.db.InTenantTx(ctx, shop, func(ctx context.Context, q database.Queries) error {
		customer, err := resolveCustomer(ctx, q, barbershopID, input.Customer)
		if err != nil {
			return err
		}

		appointment, err := insertAppointment(ctx, q, barbershopID, customer.ID, input)
		if err != nil {
			return err
		}

		if err := insertAppointmentCreatedHistory(ctx, q, barbershopID, appointment.ID, input.Actor); err != nil {
			return err
		}

		result = booking.CreateInternalResult{Appointment: appointment, Customer: customer}
		return nil
	})
	if err != nil {
		return booking.CreateInternalResult{}, err
	}
	return result, nil
}

// FindCustomerForReconciliation implementa
// booking.Repository.FindCustomerForReconciliation (DEC-071): por teléfono
// si phone no es nil (DEC-045), si no por correo si email no es nil
// (DEC-046). found=false cuando no hay coincidencia o cuando ambos llegan
// nil.
func (r *Repository) FindCustomerForReconciliation(
	ctx context.Context,
	barbershopID string,
	phone, email *string,
) (booking.Customer, bool, error) {
	shop, err := database.ValidBarbershopID(barbershopID)
	if err != nil {
		return booking.Customer{}, false, fmt.Errorf("booking/postgres: %w", err)
	}

	var (
		customer booking.Customer
		found    bool
	)
	err = r.db.InTenantTx(ctx, shop, func(ctx context.Context, q database.Queries) error {
		c, ok, err := findCustomerForReconciliation(ctx, q, barbershopID, phone, email)
		if err != nil {
			return err
		}
		customer, found = c, ok
		return nil
	})
	if err != nil {
		return booking.Customer{}, false, err
	}
	return customer, found, nil
}

// findCustomerForReconciliation es la implementación compartida que tanto
// FindCustomerForReconciliation (fuera de transacción propia, la usa
// ManualBookingService ANTES de decidir CustomerInput) como CreateManual
// (dentro de su propia transacción, para acotar la carrera) reutilizan.
func findCustomerForReconciliation(
	ctx context.Context,
	q database.Queries,
	barbershopID string,
	phone, email *string,
) (booking.Customer, bool, error) {
	var (
		row  pgx.Row
		none = booking.Customer{}
	)
	switch {
	case phone != nil:
		row = q.QueryRow(ctx,
			`SELECT id, barbershop_id, full_name, phone, email, created_at, updated_at
			   FROM customer
			  WHERE barbershop_id = $1 AND phone = $2`,
			barbershopID, *phone)
	case email != nil:
		row = q.QueryRow(ctx,
			`SELECT id, barbershop_id, full_name, phone, email, created_at, updated_at
			   FROM customer
			  WHERE barbershop_id = $1 AND email = $2`,
			barbershopID, *email)
	default:
		return none, false, nil
	}

	var c booking.Customer
	err := row.Scan(&c.ID, &c.BarbershopID, &c.FullName, &c.Phone, &c.Email, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return none, false, nil
		}
		return none, false, err
	}
	return c, true, nil
}

// CreateManual implementa booking.Repository.CreateManual (HU-061): Begin,
// la reconciliación de cliente de DEC-071 (repetida aquí solo cuando
// input.Customer.ExistingID llega nil, para acotar la carrera entre el
// Find previo de ManualBookingService y este INSERT), el INSERT de la cita
// `confirmed`, el INSERT del evento appointment_created y Complete, todo
// dentro de UNA sola InTenantTx (mismo patrón que
// schedulepostgres.Repository.Create).
func (r *Repository) CreateManual(
	ctx context.Context,
	barbershopID string,
	input booking.CreateInternalInput,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (booking.CreateManualResult, error) {
	shop, err := database.ValidBarbershopID(barbershopID)
	if err != nil {
		return booking.CreateManualResult{}, fmt.Errorf("booking/postgres: %w", err)
	}

	var result booking.CreateManualResult

	err = r.db.InTenantTx(ctx, shop, func(ctx context.Context, q database.Queries) error {
		decision, err := r.coord.Begin(ctx, q, shop, key, createManualAppointmentOperation, fingerprint, createManualAppointmentIdempotencyTTL)
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

		customerInput := input.Customer
		if customerInput.ExistingID == nil && customerInput.New != nil {
			if reconciled, found, err := findCustomerForReconciliation(ctx, q, barbershopID, customerInput.New.Phone, customerInput.New.Email); err != nil {
				return err
			} else if found {
				id := reconciled.ID
				customerInput = booking.CustomerInput{ExistingID: &id}
			}
		}

		customer, err := resolveCustomer(ctx, q, barbershopID, customerInput)
		if err != nil {
			return err
		}

		effectiveInput := input
		effectiveInput.Customer = customerInput

		appointment, err := insertAppointment(ctx, q, barbershopID, customer.ID, effectiveInput)
		if err != nil {
			return err
		}

		if err := insertAppointmentCreatedHistory(ctx, q, barbershopID, appointment.ID, input.Actor); err != nil {
			return err
		}

		body, err := json.Marshal(newManualAppointmentResponseWire(appointment))
		if err != nil {
			return fmt.Errorf("marshal created appointment: %w", err)
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

		result.Appointment = appointment
		result.Customer = customer
		result.Response = stored
		return nil
	})
	if err != nil {
		return booking.CreateManualResult{}, err
	}
	return result, nil
}

// manualAppointmentResponseWire es la forma exacta que CreateManual
// serializa como cuerpo almacenado de idempotencia (idempotency_record.
// response_body): debe coincidir campo a campo con
// httpapi.AppointmentResponse para que una repetición exacta reproduzca
// bytes idénticos a los de la respuesta original (mismo criterio que
// schedulepostgres.workingHourResponseWire).
type manualAppointmentResponseWire struct {
	ID              string  `json:"id"`
	BarberID        string  `json:"barberId"`
	ServiceID       string  `json:"serviceId"`
	CustomerID      string  `json:"customerId"`
	AttendeeName    string  `json:"attendeeName"`
	StartsAt        string  `json:"startsAt"`
	EndsAt          string  `json:"endsAt"`
	Status          string  `json:"status"`
	Origin          string  `json:"origin"`
	ServiceName     string  `json:"serviceName"`
	DurationMinutes int     `json:"durationMinutes"`
	PriceAmount     string  `json:"priceAmount"`
	Currency        string  `json:"currency"`
	CustomerNote    *string `json:"customerNote"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

func newManualAppointmentResponseWire(a booking.Appointment) manualAppointmentResponseWire {
	return manualAppointmentResponseWire{
		ID:              a.ID,
		BarberID:        a.BarberID,
		ServiceID:       a.ServiceID,
		CustomerID:      a.CustomerID,
		AttendeeName:    a.AttendeeName,
		StartsAt:        a.StartsAt.Format(time.RFC3339),
		EndsAt:          a.EndsAt.Format(time.RFC3339),
		Status:          string(a.Status),
		Origin:          string(a.Origin),
		ServiceName:     a.ServiceNameSnapshot,
		DurationMinutes: a.DurationMinutesSnapshot,
		PriceAmount:     formatPriceAmount(a.PriceAmountCentsSnapshot),
		Currency:        a.CurrencySnapshot,
		CustomerNote:    a.CustomerNote,
		CreatedAt:       a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       a.UpdatedAt.Format(time.RFC3339),
	}
}

// resolveCustomer aplica CustomerInput ya validado por booking.CreateInternalInput.validate:
// exactamente uno de ExistingID/New está presente.
func resolveCustomer(
	ctx context.Context,
	q database.Queries,
	barbershopID string,
	in booking.CustomerInput,
) (booking.Customer, error) {
	if in.ExistingID != nil {
		return loadExistingCustomer(ctx, q, barbershopID, *in.ExistingID)
	}
	return insertNewCustomer(ctx, q, barbershopID, *in.New)
}

func loadExistingCustomer(
	ctx context.Context,
	q database.Queries,
	barbershopID, customerID string,
) (booking.Customer, error) {
	var c booking.Customer
	err := q.QueryRow(ctx,
		`SELECT id, barbershop_id, full_name, phone, email, created_at, updated_at
		   FROM customer
		  WHERE barbershop_id = $1 AND id = $2`,
		barbershopID, customerID,
	).Scan(&c.ID, &c.BarbershopID, &c.FullName, &c.Phone, &c.Email, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Cubre tanto un cliente inexistente como uno de otra barbería
			// (RN-TEN-01): la consulta filtra por barbershopID además de RLS,
			// sin distinguir la causa en el error.
			return booking.Customer{}, errCustomerNotFound()
		}
		return booking.Customer{}, err
	}
	return c, nil
}

func insertNewCustomer(
	ctx context.Context,
	q database.Queries,
	barbershopID string,
	in booking.NewCustomerInput,
) (booking.Customer, error) {
	var c booking.Customer
	err := q.QueryRow(ctx,
		`INSERT INTO customer (barbershop_id, full_name, phone, email)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, barbershop_id, full_name, phone, email, created_at, updated_at`,
		barbershopID, in.FullName, in.Phone, in.Email,
	).Scan(&c.ID, &c.BarbershopID, &c.FullName, &c.Phone, &c.Email, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
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

// appointmentResponseFields es la lista de columnas que toda lectura de
// appointment de este paquete usa: mantiene la RETURNING del INSERT y el
// Scan del struct siempre alineados.
const appointmentResponseFields = `
	id, barbershop_id, barber_id, service_id, customer_id, attendee_name,
	starts_at, ends_at, status, occupies_schedule, origin,
	service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot,
	customer_note, created_at, updated_at`

func scanAppointment(row pgx.Row) (booking.Appointment, error) {
	var (
		a           booking.Appointment
		priceAmount string
		status      string
		origin      string
	)
	err := row.Scan(
		&a.ID, &a.BarbershopID, &a.BarberID, &a.ServiceID, &a.CustomerID, &a.AttendeeName,
		&a.StartsAt, &a.EndsAt, &status, &a.OccupiesSchedule, &origin,
		&a.ServiceNameSnapshot, &a.DurationMinutesSnapshot, &priceAmount, &a.CurrencySnapshot,
		&a.CustomerNote, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return booking.Appointment{}, err
	}
	a.Status = booking.Status(status)
	a.Origin = booking.Origin(origin)
	cents, err := parsePriceAmount(priceAmount)
	if err != nil {
		return booking.Appointment{}, fmt.Errorf("booking/postgres: price_amount_snapshot ilegible: %w", err)
	}
	a.PriceAmountCentsSnapshot = cents
	return a, nil
}

func insertAppointment(
	ctx context.Context,
	q database.Queries,
	barbershopID, customerID string,
	input booking.CreateInternalInput,
) (booking.Appointment, error) {
	row := q.QueryRow(ctx, `
		INSERT INTO appointment (
			barbershop_id, barber_id, service_id, customer_id, attendee_name,
			starts_at, ends_at, status, origin,
			service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot,
			customer_note
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, 'confirmed', $8,
			$9, $10, $11, $12,
			$13
		)
		RETURNING`+appointmentResponseFields,
		barbershopID, input.BarberID, input.ServiceID, customerID, input.AttendeeName,
		input.StartsAt, input.EndsAt, string(input.Origin),
		input.Service.Name, input.Service.DurationMinutes, formatPriceAmount(input.Service.PriceAmountCents), input.Service.Currency,
		input.CustomerNote,
	)

	appointment, err := scanAppointment(row)
	if err != nil {
		if isConstraintViolation(err, "23P01", appointmentBarberIntervalExclConstraint) {
			return booking.Appointment{}, errScheduleConflict()
		}
		if isDeadlockDetected(err) {
			// PostgreSQL resuelve algunas carreras de inserción concurrente
			// sobre el mismo índice GiST de appointment_barber_interval_excl
			// como "deadlock detected" (40P01) en vez de exclusion_violation
			// (23P01) para la transacción perdedora: la comprobación
			// especulativa del índice puede crear un ciclo de espera entre
			// dos INSERT simultáneos que se cruzan, y el detector de
			// interbloqueos de PostgreSQL aborta uno de los dos. Verificado
			// contra PostgreSQL 14 real con una carrera de dos conexiones
			// (TestCreateInternal_ConcurrentOverlap_ExactlyOneSucceeds): en
			// este método, la ÚNICA fuente posible de un interbloqueo es esa
			// contención de exclusión, así que se traduce igual que un
			// exclusion_violation limpio.
			return booking.Appointment{}, errScheduleConflict()
		}
		if isConstraintViolation(err, "23503", appointmentCustomerFKConstraint) {
			return booking.Appointment{}, errCustomerNotFound()
		}
		if isConstraintViolation(err, "23503", appointmentBarberFKConstraint) {
			return booking.Appointment{}, errBarberNotFound()
		}
		if isConstraintViolation(err, "23503", appointmentServiceFKConstraint) {
			return booking.Appointment{}, errServiceNotFound()
		}
		return booking.Appointment{}, err
	}
	return appointment, nil
}

func insertAppointmentCreatedHistory(
	ctx context.Context,
	q database.Queries,
	barbershopID, appointmentID string,
	actor booking.Actor,
) error {
	_, err := q.Exec(ctx, `
		INSERT INTO appointment_history (
			barbershop_id, appointment_id, event_type, actor_type, actor_staff_user_id, actor_customer_id
		) VALUES ($1, $2, 'appointment_created', $3, $4, $5)`,
		barbershopID, appointmentID, string(actor.Type), actor.StaffUserID, actor.CustomerID,
	)
	return err
}

// formatPriceAmount convierte centavos (entero exacto, nunca coma
// flotante) al string decimal que price_amount_snapshot (numeric(12,2))
// espera, mismo criterio que catalog.FormatPriceCOP.
func formatPriceAmount(cents int64) string {
	whole := cents / 100
	frac := cents % 100
	return fmt.Sprintf("%d.%02d", whole, frac)
}

// parsePriceAmount invierte formatPriceAmount al leer de vuelta el valor
// que PostgreSQL devuelve para una columna numeric(12,2) (siempre con
// exactamente dos decimales, sin signo: price_amount_snapshot_ck exige
// >= 0).
func parsePriceAmount(raw string) (int64, error) {
	var whole, frac int64
	if _, err := fmt.Sscanf(raw, "%d.%d", &whole, &frac); err != nil {
		return 0, err
	}
	return whole*100 + frac, nil
}

func errScheduleConflict() error {
	return apperr.Conflict("el barbero ya tiene una cita en ese intervalo")
}

func errCustomerNotFound() error {
	return apperr.NotFound("no existe un cliente con ese identificador")
}

func errBarberNotFound() error {
	return apperr.NotFound("no existe un barbero con ese identificador")
}

func errServiceNotFound() error {
	return apperr.NotFound("no existe un servicio con ese identificador")
}

// maxAppointmentDurationLookback acota, además de starts_at < rangeEnd, el
// extremo inferior del rango explorado por ListDailyAgenda a
// rangeStart - 24h: MaxDurationMinutesSnapshot (booking.domain.go) nunca
// deja una cita durar más de un día, así que ninguna cita que interseque
// [rangeStart, rangeEnd) puede empezar antes de ese límite. Sin esta cota,
// "starts_at < rangeEnd" por sí solo obligaría a recorrer TODO el historial
// de citas anteriores del barbero antes de aplicar el filtro ends_at >
// rangeStart; con ella, PostgreSQL usa idx_appointment_shop_barber_starts_at
// (barbershop_id, barber_id, starts_at) como un rango acotado en ambos
// extremos (verificado con EXPLAIN (ANALYZE, BUFFERS), ver postgres/README
// o el reporte de HU-062).
const maxAppointmentDurationLookback = 24 * time.Hour

// ListDailyAgenda implementa booking.Repository.ListDailyAgenda (HU-062):
// lee, sin transacción explícita (una sola SELECT de solo lectura no
// necesita InTenantTx), las citas de barberID cuyo intervalo interseca
// [rangeStart, rangeEnd) (DEC-075), tenant-aware por barbershopID además de
// RLS.
func (r *Repository) ListDailyAgenda(
	ctx context.Context,
	barbershopID, barberID string,
	rangeStart, rangeEnd time.Time,
) ([]booking.DailyAgendaEntry, error) {
	shop, err := database.ValidBarbershopID(barbershopID)
	if err != nil {
		return nil, fmt.Errorf("booking/postgres: %w", err)
	}

	var entries []booking.DailyAgendaEntry

	err = r.db.InTenantTx(ctx, shop, func(ctx context.Context, q database.Queries) error {
		rows, err := q.Query(ctx, `
			SELECT id, attendee_name, starts_at, ends_at, status, origin,
			       service_name_snapshot, duration_minutes_snapshot,
			       price_amount_snapshot, price_currency_snapshot
			  FROM appointment
			 WHERE barbershop_id = $1
			   AND barber_id = $2
			   AND starts_at >= $3
			   AND starts_at < $4
			   AND ends_at > $5
			 ORDER BY starts_at, id
			 LIMIT $6`,
			barbershopID, barberID,
			rangeStart.Add(-maxAppointmentDurationLookback), rangeEnd, rangeStart,
			booking.DailyAgendaLimit,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			entry, err := scanDailyAgendaEntry(rows)
			if err != nil {
				return err
			}
			entries = append(entries, entry)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func scanDailyAgendaEntry(row pgx.Row) (booking.DailyAgendaEntry, error) {
	var (
		e           booking.DailyAgendaEntry
		priceAmount string
		status      string
		origin      string
	)
	err := row.Scan(
		&e.ID, &e.AttendeeName, &e.StartsAt, &e.EndsAt, &status, &origin,
		&e.ServiceNameSnapshot, &e.DurationMinutesSnapshot, &priceAmount, &e.CurrencySnapshot,
	)
	if err != nil {
		return booking.DailyAgendaEntry{}, err
	}
	e.Status = booking.Status(status)
	e.Origin = booking.Origin(origin)
	cents, err := parsePriceAmount(priceAmount)
	if err != nil {
		return booking.DailyAgendaEntry{}, fmt.Errorf("booking/postgres: price_amount_snapshot ilegible: %w", err)
	}
	e.PriceAmountCentsSnapshot = cents
	return e, nil
}

func errCustomerPhoneTaken() error {
	return apperr.Conflict("ya existe un cliente con ese teléfono en esta barbería")
}

func errCustomerEmailTaken() error {
	return apperr.Conflict("ya existe un cliente con ese correo en esta barbería")
}

// ---------------------------------------------------------------------------
// HU-064: detalle e historial de un turno
// ---------------------------------------------------------------------------

// GetAppointmentDetail implementa booking.Repository.GetAppointmentDetail
// (CA-064-01 a CA-064-04): une `appointment` con `customer` (ambas tablas
// propias de booking, HU-060) dentro de barbershopID, tenant-aware además de
// RLS. found=false cubre appointmentID inexistente o de otra barbería.
func (r *Repository) GetAppointmentDetail(
	ctx context.Context,
	barbershopID, appointmentID string,
) (booking.AppointmentDetail, bool, error) {
	shop, err := database.ValidBarbershopID(barbershopID)
	if err != nil {
		return booking.AppointmentDetail{}, false, fmt.Errorf("booking/postgres: %w", err)
	}

	var (
		detail booking.AppointmentDetail
		found  bool
	)
	err = r.db.InTenantTx(ctx, shop, func(ctx context.Context, q database.Queries) error {
		row := q.QueryRow(ctx, `
			SELECT a.id, a.barber_id, a.attendee_name,
			       c.full_name, c.phone, c.email,
			       a.customer_note, a.starts_at, a.ends_at, a.status, a.origin,
			       a.service_name_snapshot, a.duration_minutes_snapshot,
			       a.price_amount_snapshot, a.price_currency_snapshot,
			       a.created_at, a.updated_at
			  FROM appointment a
			  JOIN customer c ON c.barbershop_id = a.barbershop_id AND c.id = a.customer_id
			 WHERE a.barbershop_id = $1 AND a.id = $2`,
			barbershopID, appointmentID)

		d, ok, err := scanAppointmentDetail(row)
		if err != nil {
			return err
		}
		detail, found = d, ok
		return nil
	})
	if err != nil {
		return booking.AppointmentDetail{}, false, err
	}
	return detail, found, nil
}

func scanAppointmentDetail(row pgx.Row) (booking.AppointmentDetail, bool, error) {
	var (
		d           booking.AppointmentDetail
		priceAmount string
		status      string
		origin      string
		updatedAt   time.Time
	)
	err := row.Scan(
		&d.ID, &d.BarberID, &d.AttendeeName,
		&d.CustomerFullName, &d.CustomerPhone, &d.CustomerEmail,
		&d.CustomerNote, &d.StartsAt, &d.EndsAt, &status, &origin,
		&d.ServiceNameSnapshot, &d.DurationMinutesSnapshot,
		&priceAmount, &d.CurrencySnapshot,
		&d.CreatedAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return booking.AppointmentDetail{}, false, nil
		}
		return booking.AppointmentDetail{}, false, err
	}
	d.Status = booking.Status(status)
	d.Origin = booking.Origin(origin)
	cents, err := parsePriceAmount(priceAmount)
	if err != nil {
		return booking.AppointmentDetail{}, false, fmt.Errorf("booking/postgres: price_amount_snapshot ilegible: %w", err)
	}
	d.PriceAmountCentsSnapshot = cents
	d.VersionToken = booking.EncodeVersionToken(d.ID, updatedAt)
	return d, true, nil
}

// appointmentExists confirma, dentro de la misma InTenantTx que la consulta
// de historial, que appointmentID pertenece a barbershopID: sin este
// chequeo explícito, un appointmentID inexistente y uno sin ningún evento
// todavía serían indistinguibles (cero filas en ambos casos).
func appointmentExists(ctx context.Context, q database.Queries, barbershopID, appointmentID string) (bool, error) {
	var exists bool
	err := q.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM appointment WHERE barbershop_id = $1 AND id = $2)`,
		barbershopID, appointmentID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// ListAppointmentHistory implementa booking.Repository.ListAppointmentHistory
// (CA-064-05): pide LIMIT+1 filas para saber si hay página siguiente sin una
// segunda consulta COUNT, ordenadas por (occurred_at, id) — los dos últimos
// campos de idx_appointment_history_shop_appointment_occurred
// (barbershop_id, appointment_id, occurred_at), verificado con
// EXPLAIN (ANALYZE, BUFFERS) como Index Scan (ver apps/api/README.md). Las
// filas de appointment_history_change de toda la página se leen en UNA
// segunda consulta con `history_id = ANY(...)`, nunca una por fila (trabajo
// requerido §2.3, "sin generar N+1").
func (r *Repository) ListAppointmentHistory(
	ctx context.Context,
	barbershopID, appointmentID string,
	cursor *booking.HistoryCursor,
	limit int,
) ([]booking.HistoryRow, *booking.HistoryCursor, bool, error) {
	shop, err := database.ValidBarbershopID(barbershopID)
	if err != nil {
		return nil, nil, false, fmt.Errorf("booking/postgres: %w", err)
	}

	var (
		items []booking.HistoryRow
		next  *booking.HistoryCursor
		found bool
	)
	err = r.db.InTenantTx(ctx, shop, func(ctx context.Context, q database.Queries) error {
		exists, err := appointmentExists(ctx, q, barbershopID, appointmentID)
		if err != nil {
			return err
		}
		if !exists {
			return nil
		}
		found = true

		var afterOccurredAt any
		var afterID any
		if cursor != nil {
			afterOccurredAt = cursor.OccurredAt
			afterID = cursor.ID
		}

		rows, err := q.Query(ctx, `
			SELECT id, event_type, actor_type, actor_staff_user_id, actor_customer_id,
			       reason, occurred_at
			  FROM appointment_history
			 WHERE barbershop_id = $1 AND appointment_id = $2
			   AND ($3::timestamptz IS NULL OR (occurred_at, id) > ($3::timestamptz, $4::uuid))
			 ORDER BY occurred_at, id
			 LIMIT $5`,
			barbershopID, appointmentID, afterOccurredAt, afterID, limit+1,
		)
		if err != nil {
			return err
		}
		fetched, err := scanHistoryRows(rows)
		if err != nil {
			return err
		}

		hasMore := len(fetched) > limit
		if hasMore {
			fetched = fetched[:limit]
		}
		if len(fetched) > 0 && hasMore {
			last := fetched[len(fetched)-1]
			next = &booking.HistoryCursor{OccurredAt: last.OccurredAt, ID: last.ID}
		}

		changesByHistoryID, err := loadHistoryChanges(ctx, q, barbershopID, historyIDs(fetched))
		if err != nil {
			return err
		}
		for i := range fetched {
			fetched[i].Changes = changesByHistoryID[fetched[i].ID]
		}

		items = fetched
		return nil
	})
	if err != nil {
		return nil, nil, false, err
	}
	return items, next, found, nil
}

func scanHistoryRows(rows pgx.Rows) ([]booking.HistoryRow, error) {
	defer rows.Close()
	var out []booking.HistoryRow
	for rows.Next() {
		var (
			row       booking.HistoryRow
			eventType string
			actorType string
		)
		if err := rows.Scan(
			&row.ID, &eventType, &actorType, &row.ActorStaffUserID, &row.ActorCustomerID,
			&row.Reason, &row.OccurredAt,
		); err != nil {
			return nil, err
		}
		row.EventType = booking.EventType(eventType)
		row.ActorType = booking.ActorType(actorType)
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func historyIDs(rows []booking.HistoryRow) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids
}

// loadHistoryChanges lee, en una sola consulta, los campos modificados de
// TODAS las entradas de historial ids ya cargadas (trabajo requerido §2.3):
// orden estable (history_id, field_name) para que la representación no
// dependa del orden físico de inserción.
func loadHistoryChanges(
	ctx context.Context,
	q database.Queries,
	barbershopID string,
	ids []string,
) (map[string][]booking.HistoryChange, error) {
	byHistory := make(map[string][]booking.HistoryChange)
	if len(ids) == 0 {
		return byHistory, nil
	}

	rows, err := q.Query(ctx, `
		SELECT history_id, field_name, previous_value, new_value
		  FROM appointment_history_change
		 WHERE barbershop_id = $1 AND history_id = ANY($2)
		 ORDER BY history_id, field_name`,
		barbershopID, ids,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			historyID string
			change    booking.HistoryChange
		)
		if err := rows.Scan(&historyID, &change.FieldName, &change.PreviousValue, &change.NewValue); err != nil {
			return nil, err
		}
		byHistory[historyID] = append(byHistory[historyID], change)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return byHistory, nil
}

// CustomerNames implementa booking.Repository.CustomerNames (HU-064): una
// sola consulta por lote con `= ANY($2)`, tenant-aware por barbershop_id
// además de RLS. Un id sin coincidencia simplemente está ausente del mapa
// devuelto.
func (r *Repository) CustomerNames(ctx context.Context, barbershopID string, customerIDs []string) (map[string]string, error) {
	names := make(map[string]string, len(customerIDs))
	if len(customerIDs) == 0 {
		return names, nil
	}

	shop, err := database.ValidBarbershopID(barbershopID)
	if err != nil {
		return nil, fmt.Errorf("booking/postgres: %w", err)
	}

	err = r.db.InTenantTx(ctx, shop, func(ctx context.Context, q database.Queries) error {
		rows, err := q.Query(ctx,
			`SELECT id, full_name FROM customer WHERE barbershop_id = $1 AND id = ANY($2)`,
			barbershopID, customerIDs)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id, name string
			if err := rows.Scan(&id, &name); err != nil {
				return err
			}
			names[id] = name
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return names, nil
}
