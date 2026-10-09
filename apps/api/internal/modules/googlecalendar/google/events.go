package google

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"system-barbershop/internal/modules/googlecalendar"
)

// DefaultEventsBaseURL es la raíz de la API de Google Calendar v3.
const DefaultEventsBaseURL = "https://www.googleapis.com/calendar/v3"

// maxBodyBytes acota lo que se lee de una respuesta de Google.
const maxBodyBytes = 1 << 20

// maxListPages acota la paginación del chequeo de eventos borrados.
const maxListPages = 40

// EventsClient implementa googlecalendar.EventsAPI sobre la API REST de eventos
// de Google Calendar. Usa net/http y encoding/json: la superficie son cuatro
// llamadas y el SDK completo de Google no aporta nada que las justifique.
type EventsClient struct {
	baseURL string
	http    *http.Client
}

var _ googlecalendar.EventsAPI = (*EventsClient)(nil)

// NewEventsClient construye el cliente. baseURL vacío usa Google; httpClient nil
// usa uno con tiempo de espera de 20 s.
func NewEventsClient(baseURL string, httpClient *http.Client) *EventsClient {
	if baseURL == "" {
		baseURL = DefaultEventsBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &EventsClient{baseURL: baseURL, http: httpClient}
}

type eventTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone,omitempty"`
}

type reminderOverride struct {
	Method  string `json:"method"`
	Minutes int    `json:"minutes"`
}

type reminders struct {
	UseDefault bool               `json:"useDefault"`
	Overrides  []reminderOverride `json:"overrides,omitempty"`
}

type attendee struct {
	Email string `json:"email"`
}

type extendedProperties struct {
	Private map[string]string `json:"private"`
}

type eventBody struct {
	ID                      string             `json:"id,omitempty"`
	Status                  string             `json:"status"`
	Summary                 string             `json:"summary"`
	Description             string             `json:"description"`
	Start                   eventTime          `json:"start"`
	End                     eventTime          `json:"end"`
	ExtendedProperties      extendedProperties `json:"extendedProperties"`
	Reminders               reminders          `json:"reminders"`
	Attendees               []attendee         `json:"attendees"`
	GuestsCanModify         *bool              `json:"guestsCanModify,omitempty"`
	GuestsCanInviteOthers   *bool              `json:"guestsCanInviteOthers,omitempty"`
	GuestsCanSeeOtherGuests *bool              `json:"guestsCanSeeOtherGuests,omitempty"`
}

// toBody serializa el evento. El correo del cliente solo viaja en `attendees`
// (DEC-122) y el invitado nunca puede modificar el evento, invitar a otros ni
// ver a otros invitados. Sin invitado se envía `attendees: []` para que un
// PATCH retire a uno anterior.
func toBody(ev googlecalendar.Event, includeID bool) eventBody {
	loc, err := time.LoadLocation(ev.TimeZone)
	if err != nil {
		loc = time.UTC
	}
	body := eventBody{
		Status:             "confirmed",
		Summary:            ev.Summary,
		Description:        ev.Description,
		Start:              eventTime{DateTime: ev.StartsAt.In(loc).Format(time.RFC3339), TimeZone: loc.String()},
		End:                eventTime{DateTime: ev.EndsAt.In(loc).Format(time.RFC3339), TimeZone: loc.String()},
		ExtendedProperties: extendedProperties{Private: ev.Private},
		Attendees:          []attendee{},
	}
	if includeID {
		body.ID = ev.ID
	}
	if ev.ReminderMinutes != nil {
		body.Reminders = reminders{UseDefault: false, Overrides: []reminderOverride{{Method: "popup", Minutes: *ev.ReminderMinutes}}}
	} else {
		body.Reminders = reminders{UseDefault: true}
	}
	if ev.AttendeeEmail != "" {
		no := false
		body.Attendees = []attendee{{Email: ev.AttendeeEmail}}
		body.GuestsCanModify = &no
		body.GuestsCanInviteOthers = &no
		body.GuestsCanSeeOtherGuests = &no
	}
	return body
}

func sendUpdates(notify bool) string {
	if notify {
		return "all"
	}
	return "none"
}

func (c *EventsClient) eventsURL(calendarID string, parts ...string) string {
	u := c.baseURL + "/calendars/" + url.PathEscape(calendarID) + "/events"
	for _, part := range parts {
		u += "/" + url.PathEscape(part)
	}
	return u
}

// do ejecuta la solicitud y devuelve el cuerpo; cualquier estado fuera de 2xx se
// traduce a *APIError sin conservar el cuerpo ni el token.
func (c *EventsClient) do(ctx context.Context, method, rawURL, token string, body any) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("google: serializar evento: %w", err)
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, reader)
	if err != nil {
		return nil, fmt.Errorf("google: preparar solicitud: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		// url.Error incluye la URL completa: se reduce a la causa.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return nil, fmt.Errorf("google: %s: %w", urlErr.Op, urlErr.Err)
		}
		return nil, fmt.Errorf("google: solicitud: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, parseAPIError(resp, data)
	}
	return data, nil
}

func parseAPIError(resp *http.Response, data []byte) *googlecalendar.APIError {
	apiErr := &googlecalendar.APIError{Status: resp.StatusCode}
	var envelope struct {
		Error struct {
			Errors []struct {
				Reason string `json:"reason"`
			} `json:"errors"`
		} `json:"error"`
	}
	if json.Unmarshal(data, &envelope) == nil && len(envelope.Error.Errors) > 0 {
		apiErr.Reason = envelope.Error.Errors[0].Reason
	}
	if seconds, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && seconds > 0 {
		apiErr.RetryAfter = time.Duration(seconds) * time.Second
	}
	return apiErr
}

func decodePublished(data []byte) (googlecalendar.Published, error) {
	var out struct {
		ID   string `json:"id"`
		ETag string `json:"etag"`
	}
	if err := json.Unmarshal(data, &out); err != nil || out.ID == "" {
		return googlecalendar.Published{}, errors.New("google: respuesta de evento ilegible")
	}
	return googlecalendar.Published{EventID: out.ID, ETag: out.ETag}, nil
}

// Insert implementa googlecalendar.EventsAPI.
func (c *EventsClient) Insert(ctx context.Context, token, calendarID string, ev googlecalendar.Event, notify bool) (googlecalendar.Published, error) {
	data, err := c.do(ctx, http.MethodPost,
		c.eventsURL(calendarID)+"?sendUpdates="+sendUpdates(notify), token, toBody(ev, true))
	if err != nil {
		return googlecalendar.Published{}, err
	}
	return decodePublished(data)
}

// Patch implementa googlecalendar.EventsAPI.
func (c *EventsClient) Patch(ctx context.Context, token, calendarID, eventID string, ev googlecalendar.Event, notify bool) (googlecalendar.Published, error) {
	data, err := c.do(ctx, http.MethodPatch,
		c.eventsURL(calendarID, eventID)+"?sendUpdates="+sendUpdates(notify), token, toBody(ev, false))
	if err != nil {
		return googlecalendar.Published{}, err
	}
	return decodePublished(data)
}

// Delete implementa googlecalendar.EventsAPI.
func (c *EventsClient) Delete(ctx context.Context, token, calendarID, eventID string, notify bool) error {
	_, err := c.do(ctx, http.MethodDelete,
		c.eventsURL(calendarID, eventID)+"?sendUpdates="+sendUpdates(notify), token, nil)
	return err
}

// ListEventIDs implementa googlecalendar.EventsAPI: solo los eventos publicados
// por NAVA para esa conexión, por su propiedad extendida privada (nunca por
// título ni por hora) y sin syncToken.
func (c *EventsClient) ListEventIDs(ctx context.Context, token, calendarID, connectionID string, timeMin time.Time) (map[string]bool, error) {
	ids := map[string]bool{}
	pageToken := ""
	for range maxListPages {
		query := url.Values{
			"privateExtendedProperty": {googlecalendar.PropConnectionID + "=" + connectionID},
			"timeMin":                 {timeMin.UTC().Format(time.RFC3339)},
			"maxResults":              {"250"},
			"showDeleted":             {"false"},
			"singleEvents":            {"true"},
		}
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		data, err := c.do(ctx, http.MethodGet, c.eventsURL(calendarID)+"?"+query.Encode(), token, nil)
		if err != nil {
			return nil, err
		}
		var page struct {
			Items []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"items"`
			NextPageToken string `json:"nextPageToken"`
		}
		if err := json.Unmarshal(data, &page); err != nil {
			return nil, errors.New("google: listado de eventos ilegible")
		}
		for _, item := range page.Items {
			if item.Status != "cancelled" {
				ids[item.ID] = true
			}
		}
		if page.NextPageToken == "" {
			return ids, nil
		}
		pageToken = page.NextPageToken
	}
	return nil, errors.New("google: el listado de eventos excede el máximo de páginas")
}
