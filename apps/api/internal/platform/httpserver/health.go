package httpserver

import (
	"context"
	"net/http"
	"time"

	"system-barbershop/internal/platform/database"
)

// HealthHandler responde la comprobación operativa de /health descrita en
// docs/04-arquitectura/backend-go.md. No depende de PostgreSQL ni de ningún
// módulo: solo confirma que el proceso acepta conexiones.
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}

// DatabaseHealthHandler responde la comprobación de conectividad a la base de
// datos. La respuesta indica disponible o no disponible y NADA más: sin
// versión de PostgreSQL, sin nombre de base, sin host, sin usuario, sin texto
// del error del driver. Un endpoint de salud es público de hecho aunque no se
// documente; su mensaje de error es superficie de reconocimiento gratuita para
// un atacante. La salud NO abre una transacción de tenant: no tiene barbería.
// Es la única lectura autorizada fuera del patrón.
func DatabaseHealthHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		if err := db.HealthCheck(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unavailable"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}
