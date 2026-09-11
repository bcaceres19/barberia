package httpapi

// PublicBarbershopProfileResponse es la representación pública mínima de
// una barbería habilitada (HU-090, CA-090-01, CA-090-04), forma exacta del
// componente OpenAPI PublicBarbershopProfile.yaml. contactEmail/
// contactPhone viajan siempre presentes, con valor null explícito cuando la
// barbería no tiene ese contacto configurado, igual que
// shops/httpapi.BarbershopSettingsResponse.
type PublicBarbershopProfileResponse struct {
	Name         string  `json:"name"`
	Timezone     string  `json:"timezone"`
	ContactEmail *string `json:"contactEmail"`
	ContactPhone *string `json:"contactPhone"`
}

// PublicServiceResponse es la proyección pública mínima de un servicio
// (HU-091, CA-091-01), forma exacta del componente OpenAPI
// PublicServiceResponse.yaml: nunca isActive, deactivatedAt, createdAt ni
// updatedAt (esos campos son alcance exclusivo del catálogo privado,
// HU-022/HU-024).
type PublicServiceResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Description     *string `json:"description"`
	DurationMinutes int     `json:"durationMinutes"`
	Price           string  `json:"price"`
	Currency        string  `json:"currency"`
}

// PublicServiceListResponse es una página del catálogo público de
// servicios (HU-091), misma forma que ServiceListResponse.yaml del
// catálogo privado.
type PublicServiceListResponse struct {
	Items      []PublicServiceResponse `json:"items"`
	NextCursor *string                 `json:"nextCursor"`
}
