// Package observability construye el logger estructurado compartido por los
// procesos api y worker. No decide qué campos de negocio se registran; cada
// llamador es responsable de no incluir nombre, teléfono, correo, tokens ni
// contenido de mensajes, según docs/04-arquitectura/backend-go.md.
package observability

import (
	"log/slog"
	"os"
)

// NewLogger crea un logger JSON estructurado con el nivel indicado. Un nivel
// desconocido cae en slog.LevelInfo.
func NewLogger(level string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLevel(level),
	})
	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		return slog.LevelInfo
	}
	return l
}
