package handlers

import (
	"net/http"

	"github.com/BryanVanWinnendael/Harbor/dto"
	"github.com/BryanVanWinnendael/Harbor/views/container_views"
	"github.com/docker/docker/api/types"

	"github.com/labstack/echo/v4"
)

type ContainerService interface {
	GetContainers() ([]types.Container, error)
	GetContainer(id string) (types.ContainerJSON, [][]string, error)
	RestartContainer(id string) error
	PauseContainer(id string) error
	UnpauseContainer(id string) error
	GetContainerLogs(id string) (string, error)
	StopContainer(id string) error
	StartContainer(id string) error
	RecreateContainer(id string) error
	RemoveContainer(id string) error
	SetupMySQLContainer(containerName, rootPassword, databaseName, hostPort string) error
	ExecCommandInContainer(id, cmd string) (string, error)
	GetContainerStats(containerID string) (dto.ContainerStats, error)
	PruneContainers() (string, error)
	PruneImages() (string, error)
	PruneVolumes() (string, error)
	PruneNetworks() (string, error)
	PruneAll() error
}

func NewContainerHandler(cs ContainerService) *ContainerHandler {
	return &ContainerHandler{
		ContainerServices: cs,
	}
}

type ContainerHandler struct {
	ContainerServices ContainerService
}

func (cs *ContainerHandler) getContainers(c echo.Context) error {
	containers, err := cs.ContainerServices.GetContainers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get containers"})
	}

	return renderView(c, container_views.ContainerTable(containers))
}

func (cs *ContainerHandler) listContainers(c echo.Context) error {
	isError = false

	containers, err := cs.ContainerServices.GetContainers()
	if err != nil {
		isError = true
		setFlashmessages(c, "error", "Failed to get containers")
		return c.Redirect(http.StatusSeeOther, "/")
	}

	return renderView(c, container_views.ContainerIndex(
		"Containers |",
		fromProtected,
		isError,
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		container_views.ContainerHome(containers),
	))
}

func (cs *ContainerHandler) getContainer(c echo.Context) error {
	id := c.Param("id")

	container, url, err := cs.ContainerServices.GetContainer(id)
	if err != nil {
		setFlashmessages(c, "error", "Failed to get container")
		return c.Redirect(http.StatusSeeOther, "/")
	}

	return renderView(c, container_views.ContainerIndex(
		container.Name+" |",
		fromProtected,
		isError,
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		container_views.Container(container, url),
	))
}

func (cs *ContainerHandler) getContainerStatus(c echo.Context) error {
	id := c.Param("id")

	container, _, err := cs.ContainerServices.GetContainer(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get container"})
	}

	return renderView(c, container_views.ContainerStatus(container.State.Status))
}

func (cs *ContainerHandler) getContainerLogs(c echo.Context) error {
	id := c.Param("id")

	logs, err := cs.ContainerServices.GetContainerLogs(id)
	if err != nil {
		setFlashmessages(c, "error", "Failed to get container logs")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get container logs"})
	}
	return renderView(c, container_views.ContainerLogs(logs))
}

func (cs *ContainerHandler) restartContainer(c echo.Context) error {
	id := c.Param("id")

	err := cs.ContainerServices.RestartContainer(id)
	if err != nil {
		setFlashmessages(c, "error", "Failed to restart container")
		return c.Redirect(http.StatusSeeOther, "/containers/"+id)
	}

	setFlashmessages(c, "success", "Container restarted successfully")

	return c.Redirect(http.StatusSeeOther, "/containers/"+id)
}

func (cs *ContainerHandler) pauseContainer(c echo.Context) error {
	id := c.Param("id")

	err := cs.ContainerServices.PauseContainer(id)
	if err != nil {
		setFlashmessages(c, "error", "Failed to pause container")
		return c.Redirect(http.StatusSeeOther, "/containers/"+id)
	}

	setFlashmessages(c, "success", "Container paused successfully")

	return c.Redirect(http.StatusSeeOther, "/containers/"+id)
}

func (cs *ContainerHandler) unpauseContainer(c echo.Context) error {
	id := c.Param("id")

	err := cs.ContainerServices.UnpauseContainer(id)
	if err != nil {
		setFlashmessages(c, "error", "Failed to unpause container")
		return c.Redirect(http.StatusSeeOther, "/containers/"+id)
	}

	setFlashmessages(c, "success", "Container unpaused successfully")

	return c.Redirect(http.StatusSeeOther, "/containers/"+id)
}

func (cs *ContainerHandler) stopContainer(c echo.Context) error {
	id := c.Param("id")

	err := cs.ContainerServices.StopContainer(id)
	if err != nil {
		setFlashmessages(c, "error", "Failed to stop container")
		return c.Redirect(http.StatusSeeOther, "/containers/"+id)
	}

	setFlashmessages(c, "success", "Container stopped successfully")

	return c.Redirect(http.StatusSeeOther, "/containers/"+id)
}

func (cs *ContainerHandler) startContainer(c echo.Context) error {
	id := c.Param("id")

	err := cs.ContainerServices.StartContainer(id)
	if err != nil {
		setFlashmessages(c, "error", "Failed to start container")
		return c.Redirect(http.StatusSeeOther, "/containers/"+id)
	}

	setFlashmessages(c, "success", "Container started successfully")

	return c.Redirect(http.StatusSeeOther, "/containers/"+id)
}

func (cs *ContainerHandler) recreateContainer(c echo.Context) error {
	id := c.Param("id")

	err := cs.ContainerServices.RecreateContainer(id)
	if err != nil {
		setFlashmessages(c, "error", "Failed to recreate container")
		return c.Redirect(http.StatusSeeOther, "/containers/"+id)
	}

	setFlashmessages(c, "success", "Container recreated successfully")

	return c.Redirect(http.StatusSeeOther, "/")
}

func (cs *ContainerHandler) removeContainer(c echo.Context) error {
	id := c.Param("id")

	err := cs.ContainerServices.RemoveContainer(id)
	if err != nil {
		setFlashmessages(c, "error", "Failed to remove container")
		return c.Redirect(http.StatusSeeOther, "/containers/"+id)
	}

	setFlashmessages(c, "success", "Container removed successfully")

	return c.Redirect(http.StatusSeeOther, "/")
}

func (cs *ContainerHandler) setupMySQLContainer(c echo.Context) error {
	containerName := c.FormValue("containerName")
	rootPassword := c.FormValue("rootPassword")
	databaseName := c.FormValue("databaseName")
	hostPort := c.FormValue("hostPort")

	err := cs.ContainerServices.SetupMySQLContainer(containerName, rootPassword, databaseName, hostPort)
	if err != nil {
		setFlashmessages(c, "error", "Failed to setup MySQL container")
		return c.Redirect(http.StatusSeeOther, "/")
	}

	setFlashmessages(c, "success", "MySQL container setup successfully")

	return c.Redirect(http.StatusSeeOther, "/")
}

func (cs *ContainerHandler) execCommandInContainer(c echo.Context) error {
	id := c.Param("id")
	cmd := c.FormValue("cmd")

	output, err := cs.ContainerServices.ExecCommandInContainer(id, cmd)
	if err != nil {
		setFlashmessages(c, "error", "Failed to run command in container")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to run command in container"})
	}

	return renderView(c, container_views.ContainerBashLine(output))
}

func (cs *ContainerHandler) getContainerStats(c echo.Context) error {
	id := c.Param("id")

	stats, err := cs.ContainerServices.GetContainerStats(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get container stats"})
	}

	return renderView(c, container_views.ContainerStats(stats))
}

func (cs *ContainerHandler) pruneContainers(c echo.Context) error {
	res, err := cs.ContainerServices.PruneContainers()
	if err != nil {
		setFlashmessages(c, "error", "Failed to prune containers")
		return c.Redirect(http.StatusSeeOther, "/")
	}

	setFlashmessages(c, "success", res)

	return c.Redirect(http.StatusSeeOther, "/")
}

func (cs *ContainerHandler) pruneImages(c echo.Context) error {
	res, err := cs.ContainerServices.PruneImages()
	if err != nil {
		setFlashmessages(c, "error", "Failed to prune images")
		return c.Redirect(http.StatusSeeOther, "/")
	}

	setFlashmessages(c, "success", res)

	return c.Redirect(http.StatusSeeOther, "/")
}

func (cs *ContainerHandler) pruneVolumes(c echo.Context) error {
	res, err := cs.ContainerServices.PruneVolumes()
	if err != nil {
		setFlashmessages(c, "error", "Failed to prune volumes")
		return c.Redirect(http.StatusSeeOther, "/")
	}

	setFlashmessages(c, "success", res)

	return c.Redirect(http.StatusSeeOther, "/")
}

func (cs *ContainerHandler) pruneNetworks(c echo.Context) error {
	_, err := cs.ContainerServices.PruneNetworks()
	if err != nil {
		setFlashmessages(c, "error", "Failed to prune networks")
		return c.Redirect(http.StatusSeeOther, "/")
	}

	setFlashmessages(c, "success", "Networks pruned successfully")

	return c.Redirect(http.StatusSeeOther, "/")
}

func (cs *ContainerHandler) pruneAll(c echo.Context) error {
	err := cs.ContainerServices.PruneAll()
	if err != nil {
		setFlashmessages(c, "error", "Failed to prune all")
		return c.Redirect(http.StatusSeeOther, "/")
	}

	setFlashmessages(c, "success", "All pruned successfully")

	return c.Redirect(http.StatusSeeOther, "/")
}
