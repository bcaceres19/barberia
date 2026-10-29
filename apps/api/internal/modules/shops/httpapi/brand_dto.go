package httpapi

import "encoding/json"

// BrandResponse es la representación canónica de la marca y el vocabulario
// de la barbería (issue #292, DEC-110): misma forma para GET y para la
// respuesta 200 de PATCH. Accent es una clave de la lista cerrada, nunca un
// color. PanelProfile es el panel del usuario que consulta (`shop`, `solo` o
// `employee`), derivado de sus roles (DEC-134), no un dato de la marca.
type BrandResponse struct {
	Accent                 string `json:"accent"`
	BusinessTerm           string `json:"businessTerm"`
	BusinessTermGender     string `json:"businessTermGender"`
	ProfessionalTerm       string `json:"professionalTerm"`
	ProfessionalTermPlural string `json:"professionalTermPlural"`
	ProfessionalTermGender string `json:"professionalTermGender"`
	PanelProfile           string `json:"panelProfile"`
	// Location es la ubicación física que la reserva pública muestra (DEC-130).
	// hasPhysicalLocation, address, latitude y longitude viajan siempre, con
	// null explícito cuando no hay dirección o coordenadas.
	HasPhysicalLocation bool     `json:"hasPhysicalLocation"`
	Address             *string  `json:"address"`
	Latitude            *float64 `json:"latitude"`
	Longitude           *float64 `json:"longitude"`
}

// UpdateBrandRequest es el cuerpo de PATCH /private/settings/brand. Cerrado:
// el handler rechaza cualquier campo desconocido, igual que
// UpdateBarbershopSettingsRequest. Los seis campos de la marca viajan siempre
// presentes. Ya no admite panelProfile (DEC-134): el panel lo deciden los
// roles de la base de datos. Nunca declara barbershopId: el tenant se deriva
// exclusivamente de auth.Principal.
type UpdateBrandRequest struct {
	Accent                 string `json:"accent"`
	BusinessTerm           string `json:"businessTerm"`
	BusinessTermGender     string `json:"businessTermGender"`
	ProfessionalTerm       string `json:"professionalTerm"`
	ProfessionalTermPlural string `json:"professionalTermPlural"`
	ProfessionalTermGender string `json:"professionalTermGender"`
	// HasPhysicalLocation y Address son opcionales (DEC-130): ausentes no tocan la
	// ubicación. Address distingue ausente de null (quita la dirección y sus
	// coordenadas). Las coordenadas no se envían: las calcula el servidor al
	// geocodificar una dirección nueva.
	HasPhysicalLocation *bool        `json:"hasPhysicalLocation"`
	Address             OptionalText `json:"address"`
}

// OptionalText distingue un campo ausente de uno enviado como null: ausente
// conserva el valor actual y null lo quita. `encoding/json` llama a
// UnmarshalJSON solo cuando la clave viene en el cuerpo, así que Set queda en
// false si no vino.
type OptionalText struct {
	Set   bool
	Value *string
}

// UnmarshalJSON implementa json.Unmarshaler. Un valor que no es cadena ni null
// falla con el mismo error de JSON inválido que cualquier otro campo.
func (o *OptionalText) UnmarshalJSON(data []byte) error {
	o.Set = true
	if string(data) == "null" {
		o.Value = nil
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	o.Value = &value
	return nil
}
