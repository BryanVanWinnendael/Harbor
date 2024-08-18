package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"

	"github.com/docker/docker/client"
)

func NewImageServices(cli *client.Client) *ImageServices {
	return &ImageServices{
		cli: cli,
	}
}

type ImageServices struct {
	cli *client.Client
}

func (is *ImageServices) GetImages() ([]image.Summary, error) {
	images, err := is.cli.ImageList(context.Background(), image.ListOptions{})

	for i := range images {
		if len(images[i].RepoTags) == 0 {
			images[i].RepoTags = []string{"<none>"}
		}
	}

	if err != nil {
		return nil, err
	}

	return images, nil
}

func (is *ImageServices) RemoveImage(id string) error {
	_, err := is.cli.ImageRemove(context.Background(), id, image.RemoveOptions{})

	if err != nil {
		return err
	}

	return nil
}

func (is *ImageServices) GetUsedContainers(id string) ([]types.Container, error) {
	containers, err := is.cli.ContainerList(context.Background(), container.ListOptions{})

	if err != nil {
		return nil, err
	}

	var usedContainers []types.Container

	for _, container := range containers {
		if container.ImageID == id {
			usedContainers = append(usedContainers, container)
		}
	}

	return usedContainers, nil
}

func (is *ImageServices) PullImage(imageString string) error {
	_, err := is.cli.ImagePull(context.Background(), imageString, image.PullOptions{})

	if err != nil {
		return err
	}

	return nil
}

func (is *ImageServices) CreateContainer(id, mappedPort, name, env string) error {
	envArray := []string{}

	// Split and trim environment variables
	if env != "" {
		envArray = strings.Split(env, ",")
		for i := range envArray {
			envArray[i] = strings.TrimSpace(envArray[i])
		}
	}

	hostPort := ""
	containerPort := ""

	// Parse mappedPort into hostPort and containerPort if provided
	if mappedPort != "" {
		parts := strings.Split(mappedPort, ":")
		if len(parts) == 2 {
			hostPort = parts[0]
			containerPort = parts[1]
		} else if len(parts) == 1 {
			hostPort = parts[0]
		} else {
			return fmt.Errorf("invalid mappedPort format: %s", mappedPort)
		}
	}

	// If containerPort is not provided, determine exposed ports from image metadata
	var exposedPorts map[nat.Port]struct{}
	var portBindings nat.PortMap
	if mappedPort != "" {
		if containerPort == "" {
			// Get image information to determine exposed ports
			imgInspect, _, err := is.cli.ImageInspectWithRaw(context.Background(), id)
			if err != nil {
				return fmt.Errorf("failed to inspect image %s: %v", id, err)
			}

			// Convert nat.PortSet to map[nat.Port]struct{}
			exposedPorts = make(map[nat.Port]struct{})
			for port := range imgInspect.Config.ExposedPorts {
				exposedPorts[port] = struct{}{}
			}

			for port := range exposedPorts {
				portBindings = nat.PortMap{
					port: []nat.PortBinding{
						{
							HostIP:   "0.0.0.0",
							HostPort: hostPort,
						},
					},
				}
			}

		} else {
			// Create port bindings for the specified containerPort
			exposedPort := nat.Port(containerPort + "/tcp")
			exposedPorts = map[nat.Port]struct{}{
				exposedPort: {},
			}
			portBindings = nat.PortMap{
				exposedPort: []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: hostPort,
					},
				},
			}
		}
	}

	// Configure the host configuration for the container
	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		LogConfig: container.LogConfig{
			Type:   "json-file",
			Config: map[string]string{},
		},
	}

	// Configure network settings for the container
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}

	// Container configuration
	config := &container.Config{
		Image:        id,
		ExposedPorts: exposedPorts,
		Env:          envArray,
	}

	resp, err := is.cli.ContainerCreate(context.Background(),
		config,
		hostConfig,
		networkConfig,
		nil,
		name)

	if err != nil {
		return err
	}

	if err := is.cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		fmt.Println("Error starting container:", err.Error())
		return err
	}

	fmt.Println("Container started successfully:", resp.ID)
	return nil
}
