package booking

import (
	"context"
	"time"

	"system-barbershop/internal/platform/idempotency"
)

// PublicAppointmentAdapter satisface, de forma puramente estructural,
// publicbooking.PublicAppointmentPort (HU-097): traduce los tipos
// universales que ese puerto declara hacia CreatePublicInput y llama a
// Repository.CreatePublic, sin duplicar SQL ni la traducción de errores de
// booking/postgres. publicbooking nunca importa booking (CA-002-06); este
// adaptador vive en booking exactamente por el motivo contrario: booking sí
// puede conocer los tipos universales que su propio puerto expone.
type PublicAppointmentAdapter struct {
	repo Repository
}

// NewPublicAppointmentAdapter construye el adaptador a partir del
// repositorio real (booking/postgres.New(db, coord), el mismo que ya usan
// ManualBookingService/AgendaService/DetailService en cmd/api).
func NewPublicAppointmentAdapter(repo Repository) PublicAppointmentAdapter {
	return PublicAppointmentAdapter{repo: repo}
}

// CreatePublicAppointment implementa publicbooking.PublicAppointmentPort.
func (a PublicAppointmentAdapter) CreatePublicAppointment(
	ctx context.Context,
	barbershopID, barbershopName, timezone string,
	barberID, serviceID string,
	startsAt, endsAt time.Time,
	attendeeName string,
	serviceName string,
	durationMinutes int,
	priceAmountCents int64,
	currency string,
	customerNote *string,
	customerExistingID *string,
	customerUpdatePhone *string,
	customerUpdateEmail *string,
	customerFullName, customerPhone, customerEmail string,
	tokenPlain, tokenHash string,
	tokenIssuedAt, tokenExpiresAt time.Time,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (idempotency.Decision, idempotency.StoredResponse, error) {
	customer := CustomerInput{}
	if customerExistingID != nil {
		id := *customerExistingID
		customer.ExistingID = &id
	} else {
		customer.New = &NewCustomerInput{
			FullName: customerFullName,
			Phone:    stringPtrOrNil(customerPhone),
			Email:    stringPtrOrNil(customerEmail),
		}
	}

	input := CreatePublicInput{
		BarberID:     barberID,
		ServiceID:    serviceID,
		AttendeeName: attendeeName,
		StartsAt:     startsAt,
		EndsAt:       endsAt,
		Service: ServiceSnapshot{
			Name:             serviceName,
			DurationMinutes:  durationMinutes,
			PriceAmountCents: priceAmountCents,
			Currency:         currency,
		},
		CustomerNote:        customerNote,
		BarbershopName:      barbershopName,
		Timezone:            timezone,
		TokenPlain:          tokenPlain,
		Customer:            customer,
		CustomerUpdatePhone: customerUpdatePhone,
		CustomerUpdateEmail: customerUpdateEmail,
		TokenHash:           tokenHash,
		TokenIssuedAt:       tokenIssuedAt,
		TokenExpiresAt:      tokenExpiresAt,
	}
	if err := input.validate(); err != nil {
		return idempotency.Decision{}, idempotency.StoredResponse{}, err
	}

	result, err := a.repo.CreatePublic(ctx, barbershopID, input, key, fingerprint)
	if err != nil {
		return idempotency.Decision{}, idempotency.StoredResponse{}, err
	}
	return result.Decision, result.Response, nil
}

// stringPtrOrNil convierte una cadena vacía en nil (NewCustomerInput.Phone/
// Email son opcionales, mismo criterio que ManualBookingService.
// normalizeOptional): el puerto público solo declara string porque
// customerFullName siempre es obligatorio, pero phone/email pueden llegar
// vacíos cuando DEC-085 no encontró coincidencia y HU-096 igual los exige
// -en ese caso SIEMPRE llegan no vacíos; esta función cubre, de forma
// defensiva, el caso teórico de una cadena vacía.
func stringPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	v := s
	return &v
}
