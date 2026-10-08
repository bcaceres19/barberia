package postgres

// WithSlugCode reemplaza el generador de códigos para pruebas de colisión
// deterministas. Vive en un archivo _test: el código de producción no puede
// cambiar el generador.
func (r *Repository) WithSlugCode(next func() (string, error)) *Repository {
	r.newCode = next
	return r
}
