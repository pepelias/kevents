package main

import (
	"fmt"

	"github.com/labstack/echo"
	kevents "github.com/pepelias/kevents/client"
)

func main() {
	e := echo.New()
	e.POST("/events", func(c echo.Context) error {
		fmt.Println("Recibimos un evento!")
		_, err := kevents.ListenHandler(c.Response().Writer, c.Request())
		return err
	})
	fmt.Println("Ejecutandose!")
	kevents.On("LiveGo-connected", func(info map[string]interface{}) {
		fmt.Println("Se conectó el stream")
		fmt.Println(info)
	})
	kevents.On("LiveGo-disconnected", func(info map[string]interface{}) {
		fmt.Println("Se desconectó el stream")
		fmt.Println(info)
	})
	kevents.On("LiveGo-end-stream", func(info map[string]interface{}) {
		fmt.Println("Terminó un stream")
		fmt.Println(info)
	})

	kevents.Me.Name = "José Avello"
	kevents.Me.Email = "j.avellogomez@gmail.com"
	kevents.Me.Addr = "https://jovenescp.cl"

	e.Start(":8085")
}
