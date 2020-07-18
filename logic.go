package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	kqueue "github.com/pepelias/kevents/client"
)

// Event es una cola de mensajes
type Event struct {
	Name       string
	Observable *kqueue.Observable
	Observers  []*Observer
	Createat   time.Time
}

// Observer .
type Observer struct {
	Name       string                   `json:"name"`
	NotifyAddr string                   `json:"notify_address,omitempty"`
	Email      string                   `json:"email"`
	Queue      []map[string]interface{} `json:"queue,omitempty"`
}

// Events son colas indexadas
var Events = map[string]*Event{}

// Suscribe .
func Suscribe(event string, o *Observer) error {
	if Events[event] == nil || Events[event].Observable == nil {
		return fmt.Errorf("No existe este evento")
	}
	err := Events[event].Suscribe(o)
	return err
}

// CreateEvent .
func CreateEvent(event string, own *kqueue.Observable) error {
	if own == nil {
		return fmt.Errorf("Se necesita un observable")
	}
	if Events[event] != nil && Events[event].Observable != nil {
		return fmt.Errorf("El evento ya existe")
	}
	if Events[event] != nil {
		Events[event].Observable = own
		return nil
	}
	Events[event] = &Event{
		Name:       event,
		Observable: own,
		Createat:   time.Now(),
	}
	defer saveRecovery()
	return nil
}

// Dispatch .
func Dispatch(event string, data map[string]interface{}, own *kqueue.Observable) error {
	if Events[event] == nil {
		return fmt.Errorf("No hay suscritos al evento")
	}
	Events[event].NotifyObservers(data)
	return nil
}

// Suscribe to Event
func (q *Event) Suscribe(o *Observer) error {
	for _, ob := range q.Observers {
		if ob.NotifyAddr == o.NotifyAddr {
			return fmt.Errorf("Esta dirección de notificación ya está registrada para este evento")
		}
	}
	q.Observers = append(q.Observers, o)
	defer saveRecovery()
	return nil
}

// NotifyObservers .
func (q *Event) NotifyObservers(data map[string]interface{}) {
	delete(data, "response_error")
	msg := map[string]interface{}{
		"event":    q.Name,
		"origin":   q.Observable,
		"data":     data,
		"sendedat": time.Now(),
	}
	for _, o := range q.Observers {
		err := o.Notify(msg)
		if err != nil {
			fmt.Printf("[ERROR]: %q \n", err.Error())
		} else {
			fmt.Printf("[SUCCESS]: %q \n", o.NotifyAddr)
		}
	}
}

// Notify .
func (o *Observer) Notify(message map[string]interface{}) error {
	// fmt.Printf("%q notificó a %q con el mensaje: %q", message["origin"].(*kqueue.Observable).Name, o.Name, message["message"].(string))
	err := SendMessage(o.NotifyAddr, message)
	if err != nil {
		message["response_error"] = err.Error()
		o.Queue = append(o.Queue, message)
		return err
	}
	// TODO: Si falla debe agregarse al queue
	return nil
}

// SendMessage ...
func SendMessage(destination string, content map[string]interface{}) error {
	if content["event"] == nil || content["origin"] == nil || content["data"] == nil {
		return fmt.Errorf("La estructura del mensaje es incorrecta")
	}
	// Convertir a JSON
	requestBody, _ := json.Marshal(content)

	// Cliente HTTP
	client := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	// Petición
	req, err := http.NewRequest("POST", destination, bytes.NewBuffer(requestBody))
	// Headers
	req.Header.Set("Content-type", "Application/Json")
	// req.Header.Set("Authorization", "key=AAAA680Bld8:APA91bHjFbzvXRSUN6-mF2gxp5x1_PcfunXcQU9epXR6kJSlIC2f_uacEX60Q_9yYljowcSnZYSjui7tizczIzfAAmt7jd4_MP4dtJLMUxLajvhCACqvY9JCvJLZVBM2qbfoWukyMRQQ")

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

// MakeRecovery .
func MakeRecovery() error {
	data, err := json.Marshal(Events)
	if err != nil {
		return err
	}
	b := []byte(data)
	err = ioutil.WriteFile("recovery.json", b, 0644)
	if err != nil {
		return err
	}
	return nil
}

// GetRecovery .
func GetRecovery() error {
	data, err := ioutil.ReadFile("recovery.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &Events)
	if err != nil {
		return err
	}
	return nil
}

func saveRecovery() {
	err := MakeRecovery()
	if err != nil {
		log.Printf("Error al crear copia: %q", err.Error())
		return
	}
	log.Printf("Copia generada con éxito!")
}
