// Package google es el adaptador de googlecalendar hacia Google OAuth 2.0. Es
// el ÚNICO paquete que importa golang.org/x/oauth2 (DEC-102): el núcleo y las
// pruebas del servicio usan el puerto googlecalendar.OAuthProvider.
package google

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"

	"system-barbershop/internal/modules/googlecalendar"
)

// Alcances mínimos (DEC-102): eventos del calendario más la identidad básica
// para mostrar qué cuenta quedó conectada. Nunca el alcance completo `calendar`.
const (
	ScopeCalendarEvents = "https://www.googleapis.com/auth/calendar.events"
	scopeOpenID         = "openid"
	scopeEmail          = "email"
)

// DefaultRevokeURL es el endpoint de revocación de tokens de Google.
const DefaultRevokeURL = "https://oauth2.googleapis.com/revoke"

// ErrScopeMissing indica que el barbero no concedió el permiso de calendario
// (Google permite desmarcarlo en el consentimiento): sin él no hay nada que
// publicar, así que la conexión no se considera hecha.
var ErrScopeMissing = errors.New("google: el permiso de calendario no fue concedido")

// Config son los datos de la aplicación registrada en Google Cloud.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string

	// Endpoint y RevokeURL existen para apuntar a un servidor Google falso en
	// pruebas; vacíos usan los de Google.
	Endpoint  *oauth2.Endpoint
	RevokeURL string
	// HTTPClient permite inyectar transporte y tiempo de espera; nil usa uno
	// con timeout de 10 s.
	HTTPClient *http.Client
}

// Client implementa googlecalendar.OAuthProvider.
type Client struct {
	oauth     oauth2.Config
	revokeURL string
	http      *http.Client
}

var _ googlecalendar.OAuthProvider = (*Client)(nil)

// New construye el cliente.
func New(cfg Config) *Client {
	endpoint := googleoauth.Endpoint
	if cfg.Endpoint != nil {
		endpoint = *cfg.Endpoint
	}
	revokeURL := cfg.RevokeURL
	if revokeURL == "" {
		revokeURL = DefaultRevokeURL
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		oauth: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Endpoint:     endpoint,
			Scopes:       []string{scopeOpenID, scopeEmail, ScopeCalendarEvents},
		},
		revokeURL: revokeURL,
		http:      httpClient,
	}
}

func (c *Client) withHTTP(ctx context.Context) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, c.http)
}

// AuthorizationURL construye la URL de consentimiento: código de autorización
// con PKCE S256, acceso offline (para recibir refresh token) y consentimiento
// forzado (Google solo entrega refresh token en el primer consentimiento si no).
func (c *Client) AuthorizationURL(state, codeVerifier string) string {
	return c.oauth.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.S256ChallengeOption(codeVerifier),
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
}

// Exchange canjea el código y devuelve el refresh token y la cuenta conectada.
func (c *Client) Exchange(ctx context.Context, code, codeVerifier string) (googlecalendar.Tokens, error) {
	token, err := c.oauth.Exchange(c.withHTTP(ctx), code, oauth2.VerifierOption(codeVerifier))
	if err != nil {
		return googlecalendar.Tokens{}, fmt.Errorf("google: canjear código: %w", sanitize(err))
	}
	if granted, ok := token.Extra("scope").(string); ok && !hasScope(granted, ScopeCalendarEvents) {
		return googlecalendar.Tokens{}, ErrScopeMissing
	}
	if token.RefreshToken == "" {
		return googlecalendar.Tokens{}, errors.New("google: la respuesta no trae refresh token")
	}
	email := ""
	if idToken, ok := token.Extra("id_token").(string); ok {
		email = emailFromIDToken(idToken)
	}
	return googlecalendar.Tokens{RefreshToken: token.RefreshToken, AccountEmail: email}, nil
}

// RefreshAccessToken renueva el access token. invalid_grant (permiso
// revocado, token caducado o contraseña cambiada) es permanente.
func (c *Client) RefreshAccessToken(ctx context.Context, refreshToken string) (string, time.Time, error) {
	source := c.oauth.TokenSource(c.withHTTP(ctx), &oauth2.Token{RefreshToken: refreshToken})
	token, err := source.Token()
	if err != nil {
		var retrieveErr *oauth2.RetrieveError
		if errors.As(err, &retrieveErr) && retrieveErr.ErrorCode == "invalid_grant" {
			return "", time.Time{}, googlecalendar.ErrTokenRevoked
		}
		return "", time.Time{}, fmt.Errorf("google: renovar access token: %w", sanitize(err))
	}
	return token.AccessToken, token.Expiry, nil
}

// Revoke revoca el refresh token. Un token que Google ya no reconoce (400) es
// el estado que se buscaba, no un error.
func (c *Client) Revoke(ctx context.Context, refreshToken string) error {
	form := url.Values{"token": {refreshToken}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.revokeURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("google: preparar revocación: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("google: revocar token: %w", sanitize(err))
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest {
		return nil
	}
	return fmt.Errorf("google: revocar token: estado %d", resp.StatusCode)
}

func hasScope(granted, want string) bool {
	for _, scope := range strings.Fields(granted) {
		if scope == want {
			return true
		}
	}
	return false
}

// emailFromIDToken lee el correo del id_token. El token llega directamente del
// endpoint de Google por TLS en el canje del código, así que, según OpenID
// Connect Core §3.1.3.7, no hace falta verificar su firma para este fin
// (mostrar la cuenta conectada); nunca se usa para autenticar a nadie.
func emailFromIDToken(idToken string) string {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		Email string `json:"email"`
	}
	if json.Unmarshal(payload, &claims) != nil {
		return ""
	}
	return claims.Email
}

// sanitize quita del error lo que el SDK pueda adjuntar de la respuesta de
// Google (que a veces cita partes del código o del token): solo se conserva
// el código de error estándar de OAuth.
func sanitize(err error) error {
	var retrieveErr *oauth2.RetrieveError
	if errors.As(err, &retrieveErr) {
		if retrieveErr.ErrorCode != "" {
			return fmt.Errorf("respuesta de Google: %s", retrieveErr.ErrorCode)
		}
		return fmt.Errorf("respuesta de Google: estado %d", retrieveErr.Response.StatusCode)
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		// url.Error incluye la URL completa, con parámetros sensibles.
		return fmt.Errorf("%s: %w", urlErr.Op, urlErr.Err)
	}
	return err
}
