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
