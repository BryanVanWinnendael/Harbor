package main

import (
	"os"

	"github.com/BryanVanWinnendael/Harbor/handlers"
	"github.com/BryanVanWinnendael/Harbor/services"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func main() {
	var (
		SECRET_KEY string = os.Getenv("SECRET_KEY")
	)

	e := echo.New()

	e.Static("/", "assets")
	e.Static("/css", "css")
	e.Static("/static", "static")

	e.HTTPErrorHandler = handlers.CustomHTTPErrorHandler

	e.Use(session.Middleware(sessions.NewCookieStore([]byte(SECRET_KEY))))

	ds := services.NewDockerServices()

	dh := handlers.NewDockerHandler(ds)

	// Setting Routes
	handlers.SetupRoutes(e, dh)

	// Start Server
	e.Logger.Fatal(e.Start(":3000"))
}
