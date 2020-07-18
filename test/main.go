package main

import (
	"fmt"

	"github.com/labstack/echo"
	kevents "github.com/pepelias/kevents/client"
)

func main() {
	e := echo.New()
	e.POST("/events", func(c echo.Context) error {
		_, err := kevents.ListenHandler(c.Response().Writer, c.Request())
		return err
	})

	kevents.On("streaming-start", func(info map[string]interface{}) {
		fmt.Println("Inició el streaming")
		fmt.Printf("El streaming %q ya comenzó", info["streaming_id"].(string))
	})

	kevents.Me.Name = "José Avello"
	kevents.Me.Email = "j.avellogomez@gmail.com"
	kevents.Me.Addr = "https://jovenescp.cl"

	err := kevents.Notify("streaming-start", struct{ Message string }{"Hola Mundo!"})
	if err != nil {
		fmt.Println(err)
	}
	e.Start(":8085")
}
