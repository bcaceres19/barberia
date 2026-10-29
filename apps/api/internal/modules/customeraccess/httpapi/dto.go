package httpapi

import "time"

// CustomerAppointmentResponse es la proyección mínima que HU-098 expone
// (CA-098-01, CA-098-03), forma exacta del componente OpenAPI
// CustomerAppointmentResponse.yaml: nunca un identificador interno,
// contacto del cliente, nota, historial ni ningún otro turno.
// cancellationDeadlineMinutes/lateCancellation* son los mismos nombres que
// shops/httpapi.BookingPolicyResponse (HU-093): la política vigente de
// cancelación, no un snapshot fijado al reservar.
type CustomerAppointmentResponse struct {
	BarbershopName                 string    `json:"barbershopName"`
	Timezone                       string    `json:"timezone"`
	AttendeeName                   string    `json:"attendeeName"`
	ServiceName                    string    `json:"serviceName"`
	DurationMinutes                int       `json:"durationMinutes"`
	BarberName                     string    `json:"barberName"`
	StartsAt                       time.Time `json:"startsAt"`
	EndsAt                         time.Time `json:"endsAt"`
	Status                         string    `json:"status"`
	CancellationDeadlineMinutes    int       `json:"cancellationDeadlineMinutes"`
	LateCancellationClientAllowed  bool      `json:"lateCancellationClientAllowed"`
	LateCancellationReasonRequired bool      `json:"lateCancellationReasonRequired"`

	Vocabulary PublicVocabularyResponse `json:"vocabulary"`

	// CanReview y Review (HU-100): la acción «Calificar» solo se ofrece con el
	// turno completado y sin reseña; Review es la que el cliente ya dejó.
	CanReview bool                    `json:"canReview"`
	Review    *CustomerReviewResponse `json:"review"`
}

// CustomerReviewResponse es la forma de CustomerReview.yaml: lo que el cliente
// dejó sobre su turno, sin identificadores internos.
type CustomerReviewResponse struct {
	Rating    int       `json:"rating"`
	Comment   *string   `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}

// PublicVocabularyResponse es la forma exacta del componente OpenAPI
// PublicVocabulary.yaml (DEC-119): las palabras con las que la barbería nombra a
// su negocio y a su profesional, para que el cliente lea «tu manicurista» y no
// «tu barbero». Nunca el acento ni el perfil del panel.
type PublicVocabularyResponse struct {
	BusinessTerm           string `json:"businessTerm"`
	BusinessTermGender     string `json:"businessTermGender"`
	ProfessionalTerm       string `json:"professionalTerm"`
	ProfessionalTermPlural string `json:"professionalTermPlural"`
	ProfessionalTermGender string `json:"professionalTermGender"`
}
