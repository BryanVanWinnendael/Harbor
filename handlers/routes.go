package handlers

import "github.com/labstack/echo/v4"

var (
	fromProtected bool = false
	isError       bool = false
)

func SetupRoutes(e *echo.Echo, dh *DockerHandler) {
	e.GET("/", dh.mainHandler)
}
