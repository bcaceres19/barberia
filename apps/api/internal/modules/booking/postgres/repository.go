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

// ---------------------------------------------------------------------------
// HU-065: reprogramación auditada (T2)
// ---------------------------------------------------------------------------

// rescheduleAppointmentOperation identifica, para el protocolo de
// idempotencia (RN-IDE-01, DEC-043), la operación de HU-065: una clave ya
// usada para crear un turno o para cualquier otra operación no puede
// reutilizarse aquí (idempotency_begin responde OutcomeConflictOperation).
const rescheduleAppointmentOperation idempotency.Operation = "reschedule_appointment"

// rescheduleAppointmentIdempotencyTTL es la vigencia de una reclamación
// OutcomeProceed sin completar todavía, mismo valor que
// createManualAppointmentIdempotencyTTL: sin ninguna razón de negocio para
// diferir de ese precedente.
const rescheduleAppointmentIdempotencyTTL = 10 * time.Minute

// currentAppointmentForReschedule es la fila que lockAppointmentForReschedule
// lee con `FOR UPDATE`: exactamente los campos que Reschedule necesita para
// verificar la precondición y reconstruir la representación final, sin
// columnas de control innecesarias (occupies_schedule, barbershop_id).
type currentAppointmentForReschedule struct {
	id                      string
	barberID                string
	serviceID               string
	customerID              string
	attendeeName            string
	startsAt                time.Time
	endsAt                  time.Time
	status                  string
	origin                  string
	serviceNameSnapshot     string
	durationMinutesSnapshot int
	priceAmountSnapshot     string
	currencySnapshot        string
	customerNote            *string
	createdAt               time.Time
	updatedAt               time.Time
}

// lockAppointmentForReschedule bloquea la fila (`FOR UPDATE`) dentro de la
// transacción vigente: ninguna otra transacción puede leer-modificar la
// misma fila hasta que esta transacción termine (COMMIT o ROLLBACK), lo que
// hace segura la verificación de versión/estado que Reschedule ejecuta a
// continuación frente a una segunda solicitud concurrente sobre la misma
// cita. found=false cubre appointmentID inexistente o de otra barbería.
func lockAppointmentForReschedule(
	ctx context.Context,
	q database.Queries,
	barbershopID, appointmentID string,
) (currentAppointmentForReschedule, bool, error) {
	var row currentAppointmentForReschedule
	err := q.QueryRow(ctx, `
		SELECT id, barber_id, service_id, customer_id, attendee_name,
		       starts_at, ends_at, status, origin,
		       service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot,
		       customer_note, created_at, updated_at
		  FROM appointment
		 WHERE barbershop_id = $1 AND id = $2
		 FOR UPDATE`,
		barbershopID, appointmentID,
	).Scan(
		&row.id, &row.barberID, &row.serviceID, &row.customerID, &row.attendeeName,
		&row.startsAt, &row.endsAt, &row.status, &row.origin,
		&row.serviceNameSnapshot, &row.durationMinutesSnapshot, &row.priceAmountSnapshot, &row.currencySnapshot,
		&row.customerNote, &row.createdAt, &row.updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return currentAppointmentForReschedule{}, false, nil
		}
		return currentAppointmentForReschedule{}, false, err
	}
	return row, true, nil
}

// updateAppointmentInterval aplica el nuevo intervalo sobre la fila YA
// bloqueada por lockAppointmentForReschedule, dentro de la misma
// transacción: el disparador appointment_set_updated_at recalcula
// updated_at, y la misma restricción de exclusión GiST que protege un
// INSERT (appointment_barber_interval_excl) se evalúa igual sobre este
// UPDATE, excluyendo automáticamente la propia fila del chequeo.
func updateAppointmentInterval(
	ctx context.Context,
	q database.Queries,
	barbershopID, appointmentID string,
	newStartsAt, newEndsAt time.Time,
) (time.Time, error) {
	var updatedAt time.Time
	err := q.QueryRow(ctx, `
		UPDATE appointment
		   SET starts_at = $3, ends_at = $4
		 WHERE barbershop_id = $1 AND id = $2
		RETURNING updated_at`,
		barbershopID, appointmentID, newStartsAt, newEndsAt,
	).Scan(&updatedAt)
	if err != nil {
		if isConstraintViolation(err, "23P01", appointmentBarberIntervalExclConstraint) {
			return time.Time{}, errScheduleConflict()
		}
		if isDeadlockDetected(err) {
			// Mismo criterio que insertAppointment: PostgreSQL puede
			// reportar 40P01 en vez de 23P01 para la transacción perdedora
			// de dos escrituras concurrentes que se disputan el mismo
			// índice GiST.
			return time.Time{}, errScheduleConflict()
		}
		return time.Time{}, err
	}
	return updatedAt, nil
}

// insertAppointmentRescheduledHistory inserta, dentro de la misma
// transacción, el evento appointment_rescheduled y sus dos cambios
// anterior/nuevo (starts_at, ends_at), como instantes RFC 3339 en UTC: un
// valor determinista y sin ambigüedad de zona, apto para
// appointment_history_change.previous_value/new_value (texto, hasta 1000
// caracteres).
func insertAppointmentRescheduledHistory(
	ctx context.Context,
	q database.Queries,
	barbershopID, appointmentID string,
	actor booking.Actor,
	prevStartsAt, prevEndsAt, newStartsAt, newEndsAt time.Time,
) error {
	var historyID string
	err := q.QueryRow(ctx, `
		INSERT INTO appointment_history (
			barbershop_id, appointment_id, event_type, actor_type, actor_staff_user_id, actor_customer_id
		) VALUES ($1, $2, 'appointment_rescheduled', $3, $4, $5)
		RETURNING id`,
		barbershopID, appointmentID, string(actor.Type), actor.StaffUserID, actor.CustomerID,
	).Scan(&historyID)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `
		INSERT INTO appointment_history_change (barbershop_id, history_id, field_name, previous_value, new_value)
		VALUES
			($1, $2, 'starts_at', $3, $4),
			($1, $2, 'ends_at', $5, $6)`,
		barbershopID, historyID,
		prevStartsAt.UTC().Format(time.RFC3339), newStartsAt.UTC().Format(time.RFC3339),
		prevEndsAt.UTC().Format(time.RFC3339), newEndsAt.UTC().Format(time.RFC3339),
	)
	return err
}

func newRescheduleAppointment(row currentAppointmentForReschedule) (booking.RescheduleAppointment, error) {
	cents, err := parsePriceAmount(row.priceAmountSnapshot)
	if err != nil {
		return booking.RescheduleAppointment{}, fmt.Errorf("booking/postgres: price_amount_snapshot ilegible: %w", err)
	}
	return booking.RescheduleAppointment{
		ID:                       row.id,
		BarberID:                 row.barberID,
		ServiceID:                row.serviceID,
		CustomerID:               row.customerID,
		AttendeeName:             row.attendeeName,
		StartsAt:                 row.startsAt,
		EndsAt:                   row.endsAt,
		Status:                   booking.Status(row.status),
		Origin:                   booking.Origin(row.origin),
		ServiceNameSnapshot:      row.serviceNameSnapshot,
		DurationMinutesSnapshot:  row.durationMinutesSnapshot,
		PriceAmountCentsSnapshot: cents,
		CurrencySnapshot:         row.currencySnapshot,
		CustomerNote:             row.customerNote,
		VersionToken:             booking.EncodeVersionToken(row.id, row.updatedAt),
		CreatedAt:                row.createdAt,
	}, nil
}

// rescheduleAppointmentResponseWire es la forma exacta que Reschedule
// serializa como cuerpo almacenado de idempotencia: debe coincidir campo a
// campo con httpapi.AppointmentRescheduledResponse para que una repetición
// exacta reproduzca bytes idénticos a los de la respuesta original (mismo
// criterio que manualAppointmentResponseWire).
type rescheduleAppointmentResponseWire struct {
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
	VersionToken    string  `json:"versionToken"`
	CreatedAt       string  `json:"createdAt"`
}

func newRescheduleAppointmentResponseWire(a booking.RescheduleAppointment) rescheduleAppointmentResponseWire {
	return rescheduleAppointmentResponseWire{
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
		VersionToken:    a.VersionToken,
		CreatedAt:       a.CreatedAt.Format(time.RFC3339),
	}
}

func errAppointmentNotFound() error {
	return apperr.NotFound("no existe una cita con ese identificador")
}

// errVersionConflict cubre HU-065: el token opaco de versión que el cliente
// envió (cabecera If-Match) ya no coincide con la representación vigente,
// verificado con la fila ya bloqueada dentro de la transacción.
func errVersionConflict() error {
	return apperr.VersionConflict("el turno cambió desde que se leyó; recarga antes de reintentar")
}

// errAppointmentNotConfirmed cubre HU-065: T2 solo aplica sobre una cita
// `confirmed`, verificado con la fila ya bloqueada dentro de la transacción
// (nunca confía en el estado que RescheduleService leyó antes de abrir
// esta transacción).
func errAppointmentNotConfirmed() error {
	return apperr.InvalidState("el turno ya no está confirmado; recarga para ver su estado actual")
}

// errAppointmentNotStarted cubre HU-067 (CA-067-03): T4 manual/T7 solo
// aplican sobre una cita `confirmed` cuyo starts_at ya pasó, verificado con
// la fila ya bloqueada dentro de la transacción contra el instante que el
// dominio resolvió (input.Now). apperr.Validation (422), nunca un estado
// de conflicto: la operación es inválida, no está en disputa con otra
// escritura.
func errAppointmentNotStarted() error {
	return apperr.Validation("el turno todavía no comienza; espera hasta su hora de inicio")
}

// errAppointmentAlreadyClosed cubre HU-067 (CA-067-05): el turno, con la
// fila ya bloqueada, tiene un resultado terminal distinto del que el
// comando intenta aplicar (el resultado contrario, o cualquier estado
// cancelado). Repetir el MISMO resultado nunca llega aquí: se resuelve como
// no-op antes de esta rama.
func errAppointmentAlreadyClosed() error {
	return apperr.InvalidState("el turno ya tiene un resultado terminal registrado; una futura corrección (T8) permitirá cambiarlo")
}

// Reschedule implementa booking.Repository.Reschedule (HU-065, T2): Begin,
// bloquear la fila, verificar versión/estado con la fila ya bloqueada,
// aplicar el nuevo intervalo (o detectar el no-op del mismo intervalo),
// insertar el evento appointment_rescheduled y Complete, todo dentro de UNA
// sola InTenantTx.
func (r *Repository) Reschedule(
	ctx context.Context,
	barbershopID string,
	input booking.RescheduleInput,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (booking.RescheduleResult, error) {
	shop, err := database.ValidBarbershopID(barbershopID)
	if err != nil {
		return booking.RescheduleResult{}, fmt.Errorf("booking/postgres: %w", err)
	}

	var result booking.RescheduleResult

	err = r.db.InTenantTx(ctx, shop, func(ctx context.Context, q database.Queries) error {
		decision, err := r.coord.Begin(ctx, q, shop, key, rescheduleAppointmentOperation, fingerprint, rescheduleAppointmentIdempotencyTTL)
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

		current, found, err := lockAppointmentForReschedule(ctx, q, barbershopID, input.AppointmentID)
		if err != nil {
			return err
		}
		if !found {
			return errAppointmentNotFound()
		}

		if booking.EncodeVersionToken(current.id, current.updatedAt) != input.ExpectedVersionToken {
			return errVersionConflict()
		}
		if current.status != string(booking.StatusConfirmed) {
			return errAppointmentNotConfirmed()
		}

		if !current.startsAt.Equal(input.NewStartsAt) || !current.endsAt.Equal(input.NewEndsAt) {
			updatedAt, err := updateAppointmentInterval(ctx, q, barbershopID, current.id, input.NewStartsAt, input.NewEndsAt)
			if err != nil {
				return err
			}
			if err := insertAppointmentRescheduledHistory(
				ctx, q, barbershopID, current.id, input.Actor,
				current.startsAt, current.endsAt, input.NewStartsAt, input.NewEndsAt,
			); err != nil {
				return err
			}
			current.startsAt = input.NewStartsAt
			current.endsAt = input.NewEndsAt
			current.updatedAt = updatedAt
		}
		// El intervalo idéntico es un no-op exitoso (trabajo requerido
		// §2.6): ni UPDATE ni historial, la fila conserva su updated_at
		// vigente.

		final, err := newRescheduleAppointment(current)
		if err != nil {
			return err
		}

		body, err := json.Marshal(newRescheduleAppointmentResponseWire(final))
		if err != nil {
			return fmt.Errorf("marshal rescheduled appointment: %w", err)
		}
		stored := idempotency.StoredResponse{
			Status:      200,
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

		result.Appointment = final
		result.Response = stored
		return nil
	})
	if err != nil {
		return booking.RescheduleResult{}, err
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// HU-066: cancelación auditada por el barbero (T6)
// ---------------------------------------------------------------------------

// cancelAppointmentByBarberOperation identifica, para el protocolo de
// idempotencia (RN-IDE-01, DEC-043), la operación de HU-066: una clave ya
// usada para crear o reprogramar un turno no puede reutilizarse aquí
// (idempotency_begin responde OutcomeConflictOperation).
const cancelAppointmentByBarberOperation idempotency.Operation = "cancel_appointment_by_barber"

// cancelAppointmentByBarberIdempotencyTTL es la vigencia de una reclamación
// OutcomeProceed sin completar todavía, mismo valor que
// rescheduleAppointmentIdempotencyTTL: sin ninguna razón de negocio para
// diferir de ese precedente.
const cancelAppointmentByBarberIdempotencyTTL = 10 * time.Minute

// updateAppointmentStatusCancelledByBarber aplica el estado terminal sobre
// la fila YA bloqueada por lockAppointmentForReschedule, dentro de la misma
// transacción. appointment_resolved_at_ck exige resolved_at NOT NULL (y
// >= created_at) para cualquier estado distinto de `confirmed`: por eso
// status y resolved_at se fijan en la MISMA sentencia, nunca en dos pasos
// que dejarían una fila transitoriamente inconsistente con el CHECK.
// Cancelar nunca puede violar appointment_barber_interval_excl (mover un
// turno FUERA de occupies_schedule solo puede liberar la exclusión, nunca
// competir por ella), así que este UPDATE no necesita traducir 23P01.
func updateAppointmentStatusCancelledByBarber(
	ctx context.Context,
	q database.Queries,
	barbershopID, appointmentID string,
) (time.Time, error) {
	var updatedAt time.Time
	err := q.QueryRow(ctx, `
		UPDATE appointment
		   SET status = 'cancelled_by_barber', resolved_at = now()
		 WHERE barbershop_id = $1 AND id = $2
		RETURNING updated_at`,
		barbershopID, appointmentID,
	).Scan(&updatedAt)
	if err != nil {
		return time.Time{}, err
	}
	return updatedAt, nil
}

// insertAppointmentCancelledByBarberHistory inserta, dentro de la misma
// transacción, el evento appointment_cancelled_by_barber y su único cambio
// (status: confirmed -> cancelled_by_barber), mismo criterio de dos INSERT
// que insertAppointmentRescheduledHistory.
func insertAppointmentCancelledByBarberHistory(
	ctx context.Context,
	q database.Queries,
	barbershopID, appointmentID string,
	actor booking.Actor,
) error {
	var historyID string
	err := q.QueryRow(ctx, `
		INSERT INTO appointment_history (
			barbershop_id, appointment_id, event_type, actor_type, actor_staff_user_id, actor_customer_id
		) VALUES ($1, $2, 'appointment_cancelled_by_barber', $3, $4, $5)
		RETURNING id`,
		barbershopID, appointmentID, string(actor.Type), actor.StaffUserID, actor.CustomerID,
	).Scan(&historyID)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `
		INSERT INTO appointment_history_change (barbershop_id, history_id, field_name, previous_value, new_value)
		VALUES ($1, $2, 'status', 'confirmed', 'cancelled_by_barber')`,
		barbershopID, historyID,
	)
	return err
}

func newCancelledAppointment(row currentAppointmentForReschedule) (booking.CancelledAppointment, error) {
	cents, err := parsePriceAmount(row.priceAmountSnapshot)
	if err != nil {
		return booking.CancelledAppointment{}, fmt.Errorf("booking/postgres: price_amount_snapshot ilegible: %w", err)
	}
	return booking.CancelledAppointment{
		ID:                       row.id,
		BarberID:                 row.barberID,
		ServiceID:                row.serviceID,
		CustomerID:               row.customerID,
		AttendeeName:             row.attendeeName,
		StartsAt:                 row.startsAt,
		EndsAt:                   row.endsAt,
		Status:                   booking.Status(row.status),
		Origin:                   booking.Origin(row.origin),
		ServiceNameSnapshot:      row.serviceNameSnapshot,
		DurationMinutesSnapshot:  row.durationMinutesSnapshot,
		PriceAmountCentsSnapshot: cents,
		CurrencySnapshot:         row.currencySnapshot,
		CustomerNote:             row.customerNote,
		VersionToken:             booking.EncodeVersionToken(row.id, row.updatedAt),
		CreatedAt:                row.createdAt,
	}, nil
}

// cancelAppointmentByBarberResponseWire es la forma exacta que CancelByBarber
// serializa como cuerpo almacenado de idempotencia: debe coincidir campo a
// campo con httpapi.CancelAppointmentByBarberResponse para que una
// repetición exacta reproduzca bytes idénticos a los de la respuesta
// original (mismo criterio que rescheduleAppointmentResponseWire).
type cancelAppointmentByBarberResponseWire struct {
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
	VersionToken    string  `json:"versionToken"`
	CreatedAt       string  `json:"createdAt"`
}

func newCancelAppointmentByBarberResponseWire(a booking.CancelledAppointment) cancelAppointmentByBarberResponseWire {
	return cancelAppointmentByBarberResponseWire{
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
		VersionToken:    a.VersionToken,
		CreatedAt:       a.CreatedAt.Format(time.RFC3339),
	}
}

// CancelByBarber implementa booking.Repository.CancelByBarber (HU-066, T6):
// Begin, bloquear la fila, verificar de nuevo con la fila ya bloqueada
// (no-op si ya está cancelled_by_barber, InvalidState si es cualquier otro
// estado terminal, VersionConflict si el token no coincide mientras sigue
// confirmed), aplicar el estado terminal + insertar
// appointment_cancelled_by_barber, y Complete, todo dentro de UNA sola
// InTenantTx. Reutiliza lockAppointmentForReschedule/
// currentAppointmentForReschedule (HU-065): misma fila, mismas columnas,
// ningún motivo para duplicar el SELECT ... FOR UPDATE.
func (r *Repository) CancelByBarber(
	ctx context.Context,
	barbershopID string,
	input booking.CancelAppointmentByBarberInput,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (booking.CancelAppointmentByBarberResult, error) {
	shop, err := database.ValidBarbershopID(barbershopID)
	if err != nil {
		return booking.CancelAppointmentByBarberResult{}, fmt.Errorf("booking/postgres: %w", err)
	}

	var result booking.CancelAppointmentByBarberResult

	err = r.db.InTenantTx(ctx, shop, func(ctx context.Context, q database.Queries) error {
		decision, err := r.coord.Begin(ctx, q, shop, key, cancelAppointmentByBarberOperation, fingerprint, cancelAppointmentByBarberIdempotencyTTL)
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

		current, found, err := lockAppointmentForReschedule(ctx, q, barbershopID, input.AppointmentID)
		if err != nil {
			return err
		}
		if !found {
			return errAppointmentNotFound()
		}

		switch {
		case current.status == string(booking.StatusCancelledByBarber):
			// CA-066-04: una cancelación repetida por CUALQUIER barbero de
			// la barbería siempre tiene éxito, sin comparar versionToken
			// (el que el cliente conserva ya quedó obsoleto tras la
			// primera cancelación) y sin duplicar el evento de historial.
		case current.status != string(booking.StatusConfirmed):
			return errAppointmentNotConfirmed()
		default:
			if booking.EncodeVersionToken(current.id, current.updatedAt) != input.ExpectedVersionToken {
				return errVersionConflict()
			}
			updatedAt, err := updateAppointmentStatusCancelledByBarber(ctx, q, barbershopID, current.id)
			if err != nil {
				return err
			}
			if err := insertAppointmentCancelledByBarberHistory(ctx, q, barbershopID, current.id, input.Actor); err != nil {
				return err
			}
			current.status = string(booking.StatusCancelledByBarber)
			current.updatedAt = updatedAt
		}

		final, err := newCancelledAppointment(current)
		if err != nil {
			return err
		}

		body, err := json.Marshal(newCancelAppointmentByBarberResponseWire(final))
		if err != nil {
			return fmt.Errorf("marshal cancelled appointment: %w", err)
		}
		stored := idempotency.StoredResponse{
			Status:      200,
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

		result.Appointment = final
		result.Response = stored
		return nil
	})
	if err != nil {
		return booking.CancelAppointmentByBarberResult{}, err
	}
	return result, nil
}

// completeAppointmentOperation identifica, para el protocolo de
// idempotencia (RN-IDE-01, DEC-043), el comando T4 manual de HU-067 frente
// a cualquier otra operación de booking, mismo criterio que
// cancelAppointmentByBarberOperation.
const completeAppointmentOperation idempotency.Operation = "complete_appointment"

// completeAppointmentIdempotencyTTL es la vigencia de una reclamación de
// idempotencia de T4 manual, mismo criterio que las demás operaciones de
// escritura de booking.
const completeAppointmentIdempotencyTTL = 10 * time.Minute

// markAppointmentNoShowOperation identifica el comando T7 de HU-067.
const markAppointmentNoShowOperation idempotency.Operation = "mark_appointment_no_show"

// markAppointmentNoShowIdempotencyTTL es la vigencia de una reclamación de
// idempotencia de T7.
const markAppointmentNoShowIdempotencyTTL = 10 * time.Minute

// updateAppointmentStatusCompleted aplica `completed` sobre la fila YA
// bloqueada por lockAppointmentForReschedule, dentro de la misma
// transacción: mismo criterio que updateAppointmentStatusCancelledByBarber.
// `completed` sigue ocupando agenda (occupies_schedule ya lo incluye desde
// HU-060), así que este UPDATE tampoco puede violar
// appointment_barber_interval_excl (mismo intervalo, mismo barbero, nunca
// compite por la exclusión).
func updateAppointmentStatusCompleted(
	ctx context.Context,
	q database.Queries,
	barbershopID, appointmentID string,
) (time.Time, error) {
	var updatedAt time.Time
	err := q.QueryRow(ctx, `
		UPDATE appointment
		   SET status = 'completed', resolved_at = now()
		 WHERE barbershop_id = $1 AND id = $2
		RETURNING updated_at`,
		barbershopID, appointmentID,
	).Scan(&updatedAt)
	if err != nil {
		return time.Time{}, err
	}
	return updatedAt, nil
}

// updateAppointmentStatusNoShow aplica `no_show` sobre la fila YA bloqueada,
// mismo criterio que updateAppointmentStatusCompleted.
func updateAppointmentStatusNoShow(
	ctx context.Context,
	q database.Queries,
	barbershopID, appointmentID string,
) (time.Time, error) {
	var updatedAt time.Time
	err := q.QueryRow(ctx, `
		UPDATE appointment
		   SET status = 'no_show', resolved_at = now()
		 WHERE barbershop_id = $1 AND id = $2
		RETURNING updated_at`,
		barbershopID, appointmentID,
	).Scan(&updatedAt)
	if err != nil {
		return time.Time{}, err
	}
	return updatedAt, nil
}

// insertAppointmentCompletedHistory inserta, dentro de la misma
// transacción, el evento appointment_completed y su único cambio (status:
// confirmed -> completed), mismo criterio de dos INSERT que
// insertAppointmentCancelledByBarberHistory.
func insertAppointmentCompletedHistory(
	ctx context.Context,
	q database.Queries,
	barbershopID, appointmentID string,
	actor booking.Actor,
) error {
	var historyID string
	err := q.QueryRow(ctx, `
		INSERT INTO appointment_history (
			barbershop_id, appointment_id, event_type, actor_type, actor_staff_user_id, actor_customer_id
		) VALUES ($1, $2, 'appointment_completed', $3, $4, $5)
		RETURNING id`,
		barbershopID, appointmentID, string(actor.Type), actor.StaffUserID, actor.CustomerID,
	).Scan(&historyID)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `
		INSERT INTO appointment_history_change (barbershop_id, history_id, field_name, previous_value, new_value)
		VALUES ($1, $2, 'status', 'confirmed', 'completed')`,
		barbershopID, historyID,
	)
	return err
}

// insertAppointmentNoShowHistory inserta el evento appointment_no_show,
// mismo criterio que insertAppointmentCompletedHistory.
func insertAppointmentNoShowHistory(
	ctx context.Context,
	q database.Queries,
	barbershopID, appointmentID string,
	actor booking.Actor,
) error {
	var historyID string
	err := q.QueryRow(ctx, `
		INSERT INTO appointment_history (
			barbershop_id, appointment_id, event_type, actor_type, actor_staff_user_id, actor_customer_id
		) VALUES ($1, $2, 'appointment_no_show', $3, $4, $5)
		RETURNING id`,
		barbershopID, appointmentID, string(actor.Type), actor.StaffUserID, actor.CustomerID,
	).Scan(&historyID)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `
		INSERT INTO appointment_history_change (barbershop_id, history_id, field_name, previous_value, new_value)
		VALUES ($1, $2, 'status', 'confirmed', 'no_show')`,
		barbershopID, historyID,
	)
	return err
}

func newClosedAppointment(row currentAppointmentForReschedule) (booking.ClosedAppointment, error) {
	cents, err := parsePriceAmount(row.priceAmountSnapshot)
	if err != nil {
		return booking.ClosedAppointment{}, fmt.Errorf("booking/postgres: price_amount_snapshot ilegible: %w", err)
	}
	return booking.ClosedAppointment{
		ID:                       row.id,
		BarberID:                 row.barberID,
		ServiceID:                row.serviceID,
		CustomerID:               row.customerID,
		AttendeeName:             row.attendeeName,
		StartsAt:                 row.startsAt,
		EndsAt:                   row.endsAt,
		Status:                   booking.Status(row.status),
		Origin:                   booking.Origin(row.origin),
		ServiceNameSnapshot:      row.serviceNameSnapshot,
		DurationMinutesSnapshot:  row.durationMinutesSnapshot,
		PriceAmountCentsSnapshot: cents,
		CurrencySnapshot:         row.currencySnapshot,
		CustomerNote:             row.customerNote,
		VersionToken:             booking.EncodeVersionToken(row.id, row.updatedAt),
		CreatedAt:                row.createdAt,
	}, nil
}

// closedAppointmentResponseWire es la forma exacta que CompleteAppointment y
// MarkNoShow serializan como cuerpo almacenado de idempotencia: debe
// coincidir campo a campo con httpapi.AppointmentClosedResponse (ambas
// rutas comparten la misma forma de respuesta, solo status difiere) para
// que una repetición exacta reproduzca bytes idénticos a los de la
// respuesta original, mismo criterio que cancelAppointmentByBarberResponseWire.
type closedAppointmentResponseWire struct {
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
	VersionToken    string  `json:"versionToken"`
	CreatedAt       string  `json:"createdAt"`
}

func newClosedAppointmentResponseWire(a booking.ClosedAppointment) closedAppointmentResponseWire {
	return closedAppointmentResponseWire{
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
		VersionToken:    a.VersionToken,
		CreatedAt:       a.CreatedAt.Format(time.RFC3339),
	}
}

// CompleteAppointment implementa booking.Repository.CompleteAppointment
// (HU-067, T4 manual): Begin, bloquear la fila, verificar de nuevo con la
// fila ya bloqueada (no-op si ya está `completed`, Validation 422 si sigue
// `confirmed` pero starts_at > input.Now, InvalidState si es cualquier otro
// estado terminal, VersionConflict si el token no coincide mientras sigue
// `confirmed` y ya puede cerrarse), aplicar `completed` + insertar
// `appointment_completed`, y Complete, todo dentro de UNA sola InTenantTx.
// Reutiliza lockAppointmentForReschedule/currentAppointmentForReschedule
// (HU-065/HU-066): misma fila, mismas columnas, ningún motivo para duplicar
// el SELECT ... FOR UPDATE.
func (r *Repository) CompleteAppointment(
	ctx context.Context,
	barbershopID string,
	input booking.CloseAppointmentInput,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (booking.CompleteAppointmentResult, error) {
	shop, err := database.ValidBarbershopID(barbershopID)
	if err != nil {
		return booking.CompleteAppointmentResult{}, fmt.Errorf("booking/postgres: %w", err)
	}

	var result booking.CompleteAppointmentResult

	err = r.db.InTenantTx(ctx, shop, func(ctx context.Context, q database.Queries) error {
		decision, err := r.coord.Begin(ctx, q, shop, key, completeAppointmentOperation, fingerprint, completeAppointmentIdempotencyTTL)
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

		current, found, err := lockAppointmentForReschedule(ctx, q, barbershopID, input.AppointmentID)
		if err != nil {
			return err
		}
		if !found {
			return errAppointmentNotFound()
		}

		switch {
		case current.status == string(booking.StatusCompleted):
			// CA-067-05: repetir el mismo resultado siempre tiene éxito,
			// sin comparar versionToken y sin duplicar el evento.
		case current.status != string(booking.StatusConfirmed):
			return errAppointmentAlreadyClosed()
		case current.startsAt.After(input.Now):
			// CA-067-03: antes de starts_at la operación es inválida; no
			// se toca la fila ni el historial.
			return errAppointmentNotStarted()
		default:
			if booking.EncodeVersionToken(current.id, current.updatedAt) != input.ExpectedVersionToken {
				return errVersionConflict()
			}
			updatedAt, err := updateAppointmentStatusCompleted(ctx, q, barbershopID, current.id)
			if err != nil {
				return err
			}
			if err := insertAppointmentCompletedHistory(ctx, q, barbershopID, current.id, input.Actor); err != nil {
				return err
			}
			current.status = string(booking.StatusCompleted)
			current.updatedAt = updatedAt
		}

		final, err := newClosedAppointment(current)
		if err != nil {
			return err
		}

		body, err := json.Marshal(newClosedAppointmentResponseWire(final))
		if err != nil {
			return fmt.Errorf("marshal completed appointment: %w", err)
		}
		stored := idempotency.StoredResponse{
			Status:      200,
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

		result.Appointment = final
		result.Response = stored
		return nil
	})
	if err != nil {
		return booking.CompleteAppointmentResult{}, err
	}
	return result, nil
}

// MarkNoShow implementa booking.Repository.MarkNoShow (HU-067, T7): mismo
// criterio exacto que CompleteAppointment, con `no_show` +
// `appointment_no_show` en vez de `completed` + `appointment_completed`.
func (r *Repository) MarkNoShow(
	ctx context.Context,
	barbershopID string,
	input booking.CloseAppointmentInput,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (booking.MarkNoShowResult, error) {
	shop, err := database.ValidBarbershopID(barbershopID)
	if err != nil {
		return booking.MarkNoShowResult{}, fmt.Errorf("booking/postgres: %w", err)
	}

	var result booking.MarkNoShowResult

	err = r.db.InTenantTx(ctx, shop, func(ctx context.Context, q database.Queries) error {
		decision, err := r.coord.Begin(ctx, q, shop, key, markAppointmentNoShowOperation, fingerprint, markAppointmentNoShowIdempotencyTTL)
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

		current, found, err := lockAppointmentForReschedule(ctx, q, barbershopID, input.AppointmentID)
		if err != nil {
			return err
		}
		if !found {
			return errAppointmentNotFound()
		}

		switch {
		case current.status == string(booking.StatusNoShow):
			// CA-067-05: repetir el mismo resultado siempre tiene éxito,
			// sin comparar versionToken y sin duplicar el evento.
		case current.status != string(booking.StatusConfirmed):
			return errAppointmentAlreadyClosed()
		case current.startsAt.After(input.Now):
			// CA-067-03: antes de starts_at la operación es inválida; no
			// se toca la fila ni el historial.
			return errAppointmentNotStarted()
		default:
			if booking.EncodeVersionToken(current.id, current.updatedAt) != input.ExpectedVersionToken {
				return errVersionConflict()
			}
			updatedAt, err := updateAppointmentStatusNoShow(ctx, q, barbershopID, current.id)
			if err != nil {
				return err
			}
			if err := insertAppointmentNoShowHistory(ctx, q, barbershopID, current.id, input.Actor); err != nil {
				return err
			}
			current.status = string(booking.StatusNoShow)
			current.updatedAt = updatedAt
		}

		final, err := newClosedAppointment(current)
		if err != nil {
			return err
		}

		body, err := json.Marshal(newClosedAppointmentResponseWire(final))
		if err != nil {
			return fmt.Errorf("marshal no-show appointment: %w", err)
		}
		stored := idempotency.StoredResponse{
			Status:      200,
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

		result.Appointment = final
		result.Response = stored
		return nil
	})
	if err != nil {
		return booking.MarkNoShowResult{}, err
	}
	return result, nil
}
