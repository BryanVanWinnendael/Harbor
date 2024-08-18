package handlers

import (
	"net/http"

	"github.com/BryanVanWinnendael/Harbor/views/image_views"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"

	"github.com/labstack/echo/v4"
)

type ImageService interface {
	GetImages() ([]image.Summary, error)
	RemoveImage(id string) error
	GetUsedContainers(id string) ([]types.Container, error)
	PullImage(imageString string) error
	CreateContainer(id, mappedPort, name, env string) error
}

func NewImageHandler(is ImageService) *ImageHandler {
	return &ImageHandler{
		ImageServices: is,
	}
}

type ImageHandler struct {
	ImageServices ImageService
}

func (is *ImageHandler) getImages(c echo.Context) error {
	images, err := is.ImageServices.GetImages()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get images"})
	}

	return renderView(c, image_views.ImageTable(images))
}

func (is *ImageHandler) listImages(c echo.Context) error {
	images, err := is.ImageServices.GetImages()
	if err != nil {
		isError = true
		setFlashmessages(c, "error", "Failed to get images")
		return c.Redirect(http.StatusSeeOther, "/")
	}
	return renderView(c, image_views.ImageIndex(
		"Images |",
		fromProtected,
		isError,
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		image_views.ImageHome(images),
	))
}

func (is *ImageHandler) removeImage(c echo.Context) error {
	id := c.Param("id")
	err := is.ImageServices.RemoveImage(id)
	if err != nil {
		isError = true
		setFlashmessages(c, "error", "Failed to remove image")
		return c.Redirect(http.StatusSeeOther, "/images")
	}

	setFlashmessages(c, "success", "Image removed successfully")
	return c.Redirect(http.StatusSeeOther, "/images")
}

func (is *ImageHandler) getUsedContainers(c echo.Context) error {
	id := c.Param("id")
	containers, err := is.ImageServices.GetUsedContainers(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get containers"})
	}

	return renderView(c, image_views.ImageContainers(containers))
}

func (is *ImageHandler) pullImage(c echo.Context) error {
	imageString := c.FormValue("image")
	err := is.ImageServices.PullImage(imageString)
	if err != nil {
		isError = true
		setFlashmessages(c, "error", "Failed to pull image")
		return c.Redirect(http.StatusSeeOther, "/images")
	}

	setFlashmessages(c, "success", "Image pulled successfully")
	return c.Redirect(http.StatusSeeOther, "/images")
}

func (is *ImageHandler) createContainer(c echo.Context) error {
	id := c.Param("id")
	mappedPort := c.FormValue("mappedPort")
	env := c.FormValue("env")
	name := c.FormValue("name")
	err := is.ImageServices.CreateContainer(id, mappedPort, name, env)
	if err != nil {
		isError = true
		setFlashmessages(c, "error", "Failed to create container")
		return c.Redirect(http.StatusSeeOther, "/images")
	}

	setFlashmessages(c, "success", "Container created successfully")
	return c.Redirect(http.StatusSeeOther, "/")
}
