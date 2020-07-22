package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	kevents "github.com/pepelias/kevents/client"
	kqueue "github.com/pepelias/kevents/client"
)

// Event es una cola de mensajes
type Event struct {
	Name       string `json:"-"`
	Observable int    `json:"observable"`
	Observers  []int  `json:"observers"`
}

// Observer .
type Observer struct {
	Name       string                   `json:"name"`
	NotifyAddr string                   `json:"notify_address,omitempty"`
	Email      string                   `json:"email"`
	Queue      []map[string]interface{} `json:"queue,omitempty"`
}

// Database .
var Database = struct {
	Events      map[string]*Event     `json:"events"`
	Observers   []*Observer           `json:"observers"`
	Observables []*kevents.Observable `json:"observables"`
}{
	Events:      map[string]*Event{},
	Observers:   make([]*Observer, 0),
	Observables: make([]*kevents.Observable, 0),
}

// Dispatch .
func Dispatch(event string, data map[string]interface{}, own *kqueue.Observable) error {
	fmt.Println("Disparando...")
	fmt.Println(event, data, own)
	if Database.Events[event] == nil || len(Database.Events[event].Observers) == 0 {
		return fmt.Errorf("No hay suscritos al evento")
	}
	Database.Events[event].NotifyObservers(data)
	return nil
}

// NotifyObservers .
func (q *Event) NotifyObservers(data map[string]interface{}) {
	delete(data, "response_error")
	msg := map[string]interface{}{
		"event":    q.Name,
		"origin":   Database.Observables[q.Observable],
		"data":     data,
		"sendedat": time.Now(),
	}
	for _, o := range q.Observers {
		ob := Database.Observers[o]
		err := ob.Notify(msg)
		if err != nil {
			fmt.Printf("[ERROR]: %q \n", err.Error())
		} else {
			fmt.Printf("[SUCCESS]: %q \n", ob.NotifyAddr)
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
	return fmt.Errorf("[Error]: %q", strconv.Itoa(res.StatusCode))
}

// MakeRecovery .
func MakeRecovery() error {
	data, err := json.Marshal(Database)
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
	err = json.Unmarshal(data, &Database)
	for name, event := range Database.Events {
		event.Name = name
	}
	if err != nil {
		return err
	}
	return nil
}
