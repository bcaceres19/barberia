// Package postgres traduce el núcleo booking hacia PostgreSQL, según
// docs/03-desarrollo/estandar-backend-go.md. Repository es la única pieza
// que conoce pgx, database.DB y los nombres reales de tablas/constraints de
// database/migrations/20260827110000_create_appointment_core.sql.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/database"
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

// Repository implementa booking.Repository contra PostgreSQL real.
type Repository struct {
	db *database.DB
}

// New construye el repositorio a partir del pool tenant-aware compartido.
func New(db *database.DB) *Repository {
	return &Repository{db: db}
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

func errCustomerPhoneTaken() error {
	return apperr.Conflict("ya existe un cliente con ese teléfono en esta barbería")
}

func errCustomerEmailTaken() error {
	return apperr.Conflict("ya existe un cliente con ese correo en esta barbería")
}
