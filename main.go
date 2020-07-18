package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo"
	kevents "github.com/pepelias/kevents/client"
	"github.com/pepelias/tocopicadas/controllers/response"
)

func main() {
	// Recuperar copia de seguridad
	err := GetRecovery()
	if err != nil {
		log.Fatalf("Error: %q", err)
	}
	e := echo.New()
	// Suscribirse a un evento
	e.POST("/:EVENT/suscribe", suscribeRequest)
	// Crear un evento
	e.POST("/:EVENT/create", createRequest)
	// Disparar event
	e.POST("/:EVENT", dispatchRequest)

	// Crear Copia de seguridad
	e.POST("/make-recovery", func(c echo.Context) error {
		err := MakeRecovery()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, response.Model{
				Error: response.Error{
					Code:    http.StatusInternalServerError,
					Message: err.Error(),
				},
			})
		}
		return c.JSON(http.StatusCreated, response.Model{
			Ok: response.Ok{
				Code:    http.StatusOK,
				Message: "Archivo almacenado con éxito!",
			},
		})
	})

	e.Start(":8080")
}

// Suscribirse por HTTP
func suscribeRequest(c echo.Context) error {
	observer := &Observer{}
	err := json.NewDecoder(c.Request().Body).Decode(observer)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Model{
			Error: response.Error{
				Code:    http.StatusBadRequest,
				Message: "El body tiene un formato incorrecto",
			},
		})
	}
	// Verificar que estén todos los datos
	if observer.Name == "" || observer.Email == "" || observer.NotifyAddr == "" {
		return c.JSON(http.StatusBadRequest, response.Model{
			Error: response.Error{
				Code:    http.StatusBadRequest,
				Message: "Estructura de datos incorrecta. Asegurese de enviar su nombre, email y dirección URL.",
			},
		})
	}

	// Proceder a crear el evento
	err = Suscribe(c.Param("EVENT"), observer)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Model{
			Error: response.Error{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			},
		})
	}

	return c.JSON(http.StatusOK, response.Model{
		Ok: response.Ok{
			Code:    http.StatusOK,
			Message: "Suscrito al evento con éxito",
		},
		Data: map[string]interface{}{
			"name": c.Param("EVENT"),
		},
	})
}

// Crear por HTTP
func createRequest(c echo.Context) error {
	observable := &kevents.Observable{}
	err := json.NewDecoder(c.Request().Body).Decode(observable)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Model{
			Error: response.Error{
				Code:    http.StatusBadRequest,
				Message: "El body tiene un formato incorrecto",
			},
		})
	}
	// Verificar que estén todos los datos
	if observable.Name == "" || observable.Email == "" || observable.Addr == "" {
		return c.JSON(http.StatusBadRequest, response.Model{
			Error: response.Error{
				Code:    http.StatusBadRequest,
				Message: "Estructura de datos incorrecta. Asegurese de enviar su nombre, email y dirección URL.",
			},
		})
	}

	// Proceder a crear el evento
	err = CreateEvent(c.Param("EVENT"), observable)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Model{
			Error: response.Error{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			},
		})
	}

	return c.JSON(http.StatusCreated, response.Model{
		Ok: response.Ok{
			Code:    http.StatusCreated,
			Message: "Evento creado con éxito!",
		},
		Data: map[string]interface{}{
			"name":     c.Param("EVENT"),
			"createat": Events[c.Param("EVENT")].Createat,
		},
	})
}

// Disparar por HTTP
func dispatchRequest(c echo.Context) error {
	data := map[string]interface{}{}
	err := json.NewDecoder(c.Request().Body).Decode(&data)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Model{
			Error: response.Error{
				Code:    http.StatusBadRequest,
				Message: "El body tiene un formato incorrecto",
			},
		})
	}
	// Preparamos el error
	preError := response.Model{
		Error: response.Error{
			Code:    http.StatusBadRequest,
			Message: "Estructura de datos incorrecta. Se requieren los datos 'own' (Creador del evento) y 'data' (datos a enviar)",
		},
	}
	// Verificamos los dos datos
	if data["own"] == nil || data["data"] == nil {
		return c.JSON(http.StatusBadRequest, preError)
	}
	// Datos preprocesados del observable
	own := data["own"].(map[string]interface{})

	// Verificar datos del own
	if own["name"] == nil || own["name"] == "" ||
		own["email"] == nil || own["email"] == "" ||
		own["address"] == nil || own["address"] == "" {
		return c.JSON(http.StatusBadRequest, preError)
	}

	// Seteamos el observable
	observable := &kevents.Observable{
		Name:  own["name"].(string),
		Email: own["email"].(string),
		Addr:  own["address"].(string),
	}

	// Proceder disparar
	defer func() {
		fmt.Println("Se disparó")
		go Dispatch(c.Param("EVENT"), data["data"].(map[string]interface{}), observable)
	}()

	return c.JSON(http.StatusOK, response.Model{
		Ok: response.Ok{
			Code:    http.StatusOK,
			Message: "Evento disparado con éxito!",
		},
		Data: map[string]interface{}{
			"name": c.Param("EVENT"),
		},
	})
}
