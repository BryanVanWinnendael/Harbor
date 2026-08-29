package services

import (
	"context"
	"strings"

	"github.com/BryanVanWinnendael/Harbor/dto"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type ContainerEnvironmentServices struct {
	cli *client.Client
}

func NewContainerEnvironmentServices(
	cli *client.Client,
) *ContainerEnvironmentServices {
	return &ContainerEnvironmentServices{
		cli: cli,
	}
}

func (es *ContainerEnvironmentServices) GetEnvironments() ([]*dto.ContainerEnvironment, error) {
	ctx := context.Background()

	containers, err := es.cli.ContainerList(
		ctx,
		container.ListOptions{
			All: true,
		},
	)
	if err != nil {
		return nil, err
	}

	environments := make(
		[]*dto.ContainerEnvironment,
		0,
		len(containers),
	)

	for _, c := range containers {
		inspect, err := es.cli.ContainerInspect(ctx, c.ID)
		if err != nil {
			return nil, err
		}

		variables := make(
			[]*dto.ContainerEnvironmentVariable,
			0,
			len(inspect.Config.Env),
		)

		for _, env := range inspect.Config.Env {
			parts := strings.SplitN(env, "=", 2)

			key := parts[0]

			value := ""
			if len(parts) == 2 {
				value = parts[1]
			}

			variables = append(
				variables,
				&dto.ContainerEnvironmentVariable{
					Key:      key,
					Value:    value,
					IsSecret: isSecretEnvironmentVariable(key),
				},
			)
		}

		name := strings.TrimPrefix(c.Names[0], "/")

		environments = append(
			environments,
			&dto.ContainerEnvironment{
				ContainerID:   c.ID,
				ContainerName: name,
				Variables:     variables,
			},
		)
	}

	return environments, nil
}

func isSecretEnvironmentVariable(key string) bool {
	key = strings.ToUpper(key)

	secretKeywords := []string{
		"PASSWORD",
		"PASSWD",
		"SECRET",
		"TOKEN",
		"API_KEY",
		"APIKEY",
		"PRIVATE_KEY",
		"PRIVATEKEY",
		"CREDENTIAL",
		"CREDENTIALS",
		"AUTH",
	}

	for _, keyword := range secretKeywords {
		if strings.Contains(key, keyword) {
			return true
		}
	}

	return false
}
