package httpapi

// BrandResponse es la representación canónica de la marca y el vocabulario
// de la barbería (issue #292, DEC-110): misma forma para GET y para la
// respuesta 200 de PATCH. Accent es una clave de la lista cerrada, nunca un
// color.
type BrandResponse struct {
	Accent                 string `json:"accent"`
	BusinessTerm           string `json:"businessTerm"`
	BusinessTermGender     string `json:"businessTermGender"`
	ProfessionalTerm       string `json:"professionalTerm"`
	ProfessionalTermPlural string `json:"professionalTermPlural"`
	ProfessionalTermGender string `json:"professionalTermGender"`
}

// UpdateBrandRequest es el cuerpo de PATCH /private/settings/brand. Cerrado:
// el handler rechaza cualquier campo desconocido, igual que
// UpdateBarbershopSettingsRequest. Los seis campos viajan siempre presentes.
// Nunca declara barbershopId: el tenant se deriva exclusivamente de
// auth.Principal.
type UpdateBrandRequest struct {
	Accent                 string `json:"accent"`
	BusinessTerm           string `json:"businessTerm"`
	BusinessTermGender     string `json:"businessTermGender"`
	ProfessionalTerm       string `json:"professionalTerm"`
	ProfessionalTermPlural string `json:"professionalTermPlural"`
	ProfessionalTermGender string `json:"professionalTermGender"`
}
