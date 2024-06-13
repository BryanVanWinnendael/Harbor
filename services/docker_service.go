package services

import (
	"github.com/labstack/echo/v4"
)

func NewDockerServices() *DockerServices {

	return &DockerServices{
	}
}

type DockerServices struct {
}

func (ps *DockerServices) UploadPost(c echo.Context) error {
	return nil
}

