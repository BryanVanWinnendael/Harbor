package handlers

import "github.com/labstack/echo/v4"

var (
	isError bool = false
)

func SetupRoutes(e *echo.Echo, dh *DockerHandler) {
	e.GET("/", dh.mainHandler)
}
