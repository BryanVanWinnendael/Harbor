package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/a-h/templ"
	"harbor/views/docker_views"
)

type DockerService interface {
}

func NewDockerHandler(ps DockerService) *DockerHandler {

	return &DockerHandler{
		DockerServices: ps,
	}
}

type DockerHandler struct {
	DockerServices DockerService
}

func renderView(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)

	return cmp.Render(c.Request().Context(), c.Response().Writer)
}

func (ds *DockerHandler) mainHandler(c echo.Context) error {
	return renderView(c, docker_views.DockerIndex(
		"| Create Post",
		"",
		fromProtected,
		isError,
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		docker_views.Home(),
	))
}

