package main

import (
	"os"

	"github.com/BryanVanWinnendael/Harbor/db"
	"github.com/BryanVanWinnendael/Harbor/handlers"
	"github.com/BryanVanWinnendael/Harbor/services"
	"github.com/docker/docker/client"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func main() {
	godotenv.Load()

	var (
		SECRET_KEY string = os.Getenv("SECRET_KEY")
		DB_NAME    string = "app_data.db"
	)

	e := echo.New()

	e.Static("/", "assets")
	e.Static("/css", "css")
	e.Static("/static", "static")

	e.HTTPErrorHandler = handlers.CustomHTTPErrorHandler

	e.Use(session.Middleware(sessions.NewCookieStore([]byte(SECRET_KEY))))

	cli, err := client.NewClientWithOpts(
		client.WithHost("unix:///var/run/docker.sock"),
		client.WithVersion("1.45"),
	)

	if err != nil {
		e.Logger.Fatal(err)
	} else {
		e.Logger.Infof("Connected to Docker Daemon")
	}

	store, err := db.NewStore(DB_NAME)
	if err != nil {
		e.Logger.Fatalf("failed to create store: %s", err)
	} else {
		e.Logger.Infof("Connected to Database")
	}

	us := services.NewUserServices(services.User{}, store)
	cs := services.NewContainerServices(cli)
	is := services.NewImageServices(cli)
	as := services.NewAnalyticsServices(cli)

	ch := handlers.NewContainerHandler(cs)
	ih := handlers.NewImageHandler(is)
	anh := handlers.NewAnalyticsHandler(as)
	ah := handlers.NewAuthHandler(us)

	// Setting Routes
	handlers.SetupRoutes(e, ah, ch, ih, anh)

	// Start Server
	e.Logger.Fatal(e.Start(":3000"))
}
