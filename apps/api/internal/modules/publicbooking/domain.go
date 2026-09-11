package publicbooking

// BarbershopProfile es la representación pública mínima de una barbería
// habilitada (HU-090, CA-090-01, CA-090-04): nombre, zona horaria IANA y
// contacto público opcional. Nunca incluye un identificador interno, el
// slug mismo, ni ningún otro dato de configuración privada.
type BarbershopProfile struct {
	Name         string
	Timezone     string
	ContactEmail *string
	ContactPhone *string
}

// MaxSlugLength acota la longitud aceptada del identificador antes de
// consultar la base de datos: cualquier valor más largo se rechaza de
// inmediato con la misma respuesta uniforme que un slug desconocido
// (CA-090-02), sin gastar una consulta. Igual a shops.SlugMaxLength
// (barbershop_public_slug_ck): ningún slug generado supera ese largo, así
// que uno más largo nunca puede coincidir.
const MaxSlugLength = 40
