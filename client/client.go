package kevents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Observable .
type Observable struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Addr  string `json:"address"`
	Token string `json:"token,omitempty"`
}

// Event evento recibido
type Event struct {
	Event    string                 `json:"event"`
	Origin   Observable             `json:"origin"`
	Data     map[string]interface{} `json:"data"`
	Sendedat time.Time              `json:"sendedat"`
}

// Slice con escuchadores de eventos
var events = map[string][]func(map[string]interface{}){}

// Me Emisor del evento. (Mismos datos del creador del evento)
var Me = &Observable{}

// ServerAddr es la dirección del manejador de eventos
var ServerAddr = "http://localhost:8080/"

// ListenHandler debe ponerse en el endpint de escucha de eventos
func ListenHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	event := &Event{}
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		fmt.Println("Falló la decodificación")
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Fprintf(w, "{message: \"Internal Server error\"}")
	}

	fmt.Println(events)
	fmt.Println(event.Event)
	event.Data["Request"] = r
	// Verificamos que estemos escuchando
	if events[event.Event] != nil {
		// Disparamos cada evento
		for _, handler := range events[event.Event] {
			handler(event.Data)
		}
	}

	w.WriteHeader(http.StatusOK)
	return fmt.Fprintf(w, "{message: \"Recibido con éxito\"}")
}

// On Escuchar un evento
func On(event string, handler func(map[string]interface{})) {
	if events[event] == nil {
		events[event] = []func(map[string]interface{}){}
	}
	events[event] = append(events[event], handler)
}

// Notify Disparar un evento
func Notify(event string, message interface{}) error {
	msg := map[string]interface{}{
		"own":  Me,
		"data": message,
	}
	// Convertir a JSON
	requestBody, _ := json.Marshal(msg)

	// Cliente HTTP
	client := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	// Petición
	req, err := http.NewRequest("POST", ServerAddr+event, bytes.NewBuffer(requestBody))
	// Headers
	req.Header.Set("Content-type", "Application/Json")
	if err != nil {
		return err
	}

	res, err := client.Do(req)

	if err != nil {
		return err
	}
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("Falló envío: Código de respuesta: %q", strconv.Itoa(res.StatusCode))
}
