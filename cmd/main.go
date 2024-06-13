package main

import (
	"harbor/handlers"
	"github.com/labstack/echo/v4"
	"harbor/services"
)

func main() {
	e := echo.New()

	e.Static("/", "assets")
	e.Static("/css", "css")
	e.Static("/static", "static")

	e.HTTPErrorHandler = handlers.CustomHTTPErrorHandler

	
	ds := services.NewDockerServices()

	dh := handlers.NewDockerHandler(ds)

	// Setting Routes
	handlers.SetupRoutes(e, dh)

	// Start Server
	e.Logger.Fatal(e.Start(":3000"))
}
