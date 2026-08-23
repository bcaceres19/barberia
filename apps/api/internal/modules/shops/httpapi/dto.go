package httpapi

// BarbershopSettingsResponse es la representación canónica de la
// configuración básica de la barbería (CA-020-01, CA-020-07): exactamente
// los cuatro campos autorizados. ContactEmail/ContactPhone son punteros
// para que un contacto ausente serialice como `null` explícito (no se
// omiten del cuerpo, docs/06-api/estandar-openapi.md §9.3), nunca como
// cadena vacía.
type BarbershopSettingsResponse struct {
	Name         string  `json:"name"`
	Timezone     string  `json:"timezone"`
	ContactEmail *string `json:"contactEmail"`
	ContactPhone *string `json:"contactPhone"`
}

// UpdateBarbershopSettingsRequest es el cuerpo de
// PATCH /private/settings/barbershop (CA-020-02, CA-020-06). Cerrado: el
// handler rechaza cualquier campo desconocido, igual que LoginRequest.
// ContactEmail/ContactPhone vacíos representan "sin contacto" en el
// alambre; shops.Service los normaliza a ausencia antes de persistir.
// Nunca declara barbershopId: el tenant se deriva exclusivamente de
// auth.Principal (CA-020-05).
type UpdateBarbershopSettingsRequest struct {
	Name         string `json:"name"`
	Timezone     string `json:"timezone"`
	ContactEmail string `json:"contactEmail"`
	ContactPhone string `json:"contactPhone"`
}
