package handlers

import (
	"net/http"

	"github.com/BryanVanWinnendael/Harbor/dto"
	"github.com/BryanVanWinnendael/Harbor/views/container_env_views"
	"github.com/labstack/echo/v4"
)

type ContainerEnvironmentService interface {
	GetEnvironments() ([]*dto.ContainerEnvironment, error)
}

type ContainerEnvironmentHandler struct {
	EnvironmentService ContainerEnvironmentService
}

func NewContainerEnvironmentHandler(
	service ContainerEnvironmentService,
) *ContainerEnvironmentHandler {
	return &ContainerEnvironmentHandler{
		EnvironmentService: service,
	}
}

func (h *ContainerEnvironmentHandler) GetEnvironments(
	c echo.Context,
) error {
	environments, err := h.EnvironmentService.GetEnvironments()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return renderView(c, container_env_views.ContainerEnvIndex(
		"Container Environments |",
		fromProtected,
		isError,
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		container_env_views.ContainerEnvHome(environments),
	))
}
