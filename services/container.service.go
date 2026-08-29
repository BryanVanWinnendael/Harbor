package services

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"strings"

	"github.com/BryanVanWinnendael/Harbor/dto"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func NewContainerServices(cli *client.Client) *ContainerServices {
	return &ContainerServices{
		cli:  cli,
		cwds: make(map[string]string),
	}
}

type ContainerServices struct {
	cli   *client.Client
	cwdMu sync.RWMutex
	cwds  map[string]string
}

func (cs *ContainerServices) GetContainers() ([]types.Container, error) {
	containers, err := cs.cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	return containers, nil
}

func (cs *ContainerServices) GetContainer(id string) (types.ContainerJSON, [][]string, error) {
	container, err := cs.cli.ContainerInspect(context.Background(), id)
	if err != nil {
		return types.ContainerJSON{}, [][]string{}, err
	}

	ip := getip2()

	var foundPorts = make(map[string]string)

	for port, bindings := range container.NetworkSettings.Ports {
		if len(bindings) == 0 {
			foundPorts[port.Port()] = port.Port()
		} else {
			for _, binding := range bindings {
				foundPorts[binding.HostPort] = binding.HostPort
			}
		}

	}

	urls := [][]string{}

	for port, hostPort := range foundPorts {
		url := fmt.Sprintf("http://%s:%s", ip, hostPort)
		url_port := fmt.Sprintf(":%s", port)
		urls = append(urls, []string{url, url_port})
	}

	return container, urls, nil
}

func (cs *ContainerServices) RestartContainer(id string) error {
	err := cs.cli.ContainerRestart(context.Background(), id, container.StopOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (cs *ContainerServices) PauseContainer(id string) error {
	err := cs.cli.ContainerPause(context.Background(), id)
	if err != nil {
		return err
	}
	return nil
}

func (cs *ContainerServices) UnpauseContainer(id string) error {
	err := cs.cli.ContainerUnpause(context.Background(), id)
	if err != nil {
		return err
	}
	return nil
}

func (cs *ContainerServices) GetContainerLogs(id string) (string, error) {
	options := container.LogsOptions{ShowStdout: true, ShowStderr: true}

	logs, err := cs.cli.ContainerLogs(context.Background(), id, options)
	if err != nil {
		return "", err
	}
	defer logs.Close()

	var lines []string

	hdr := make([]byte, 8)
	for {
		_, err := logs.Read(hdr)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		count := binary.BigEndian.Uint32(hdr[4:])
		dat := make([]byte, count)
		_, err = logs.Read(dat)
		if err != nil {
			return "", err
		}

		lines = append(lines, string(dat))
	}

	return strings.Join(lines, "\n"), nil
}

func (cs *ContainerServices) StopContainer(id string) error {
	err := cs.cli.ContainerStop(context.Background(), id, container.StopOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (cs *ContainerServices) StartContainer(id string) error {
	err := cs.cli.ContainerStart(context.Background(), id, container.StartOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (cs *ContainerServices) RecreateContainer(id string) error {
	containerJSON, err := cs.cli.ContainerInspect(context.Background(), id)
	if err != nil {
		return err
	}
	err = cs.StopContainer(id)
	if err != nil {
		return err
	}

	removeOptions := container.RemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	err = cs.cli.ContainerRemove(context.Background(), id, removeOptions)
	if err != nil {
		return err
	}

	config := container.Config{
		Image:        containerJSON.Config.Image,
		Env:          containerJSON.Config.Env,
		Cmd:          containerJSON.Config.Cmd,
		ExposedPorts: containerJSON.Config.ExposedPorts,
		Labels:       containerJSON.Config.Labels,
	}

	hostConfig := container.HostConfig{
		Binds:        containerJSON.HostConfig.Binds,
		PortBindings: containerJSON.HostConfig.PortBindings,
	}

	networkingConfig := network.NetworkingConfig{
		EndpointsConfig: containerJSON.NetworkSettings.Networks,
	}

	newContainer, err := cs.cli.ContainerCreate(context.Background(), &config, &hostConfig, &networkingConfig, nil, containerJSON.Name)
	if err != nil {
		return err
	}

	err = cs.cli.ContainerStart(context.Background(), newContainer.ID, container.StartOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (cs *ContainerServices) RemoveContainer(id string) error {
	removeOptions := container.RemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	err := cs.cli.ContainerRemove(context.Background(), id, removeOptions)
	if err != nil {
		return err
	}

	return nil
}

func (cs *ContainerServices) SetupMySQLContainer(containerName, rootPassword, databaseName, hostPort string) error {
	imageName := "mysql:latest"
	reader, err := cs.cli.ImagePull(context.Background(), imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("error pulling image: %w", err)
	}
	defer reader.Close()

	config := &container.Config{
		Image: imageName,
		Env: []string{
			fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", rootPassword),
			fmt.Sprintf("MYSQL_DATABASE=%s", databaseName),
		},
		ExposedPorts: map[nat.Port]struct{}{
			"3306/tcp": {},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{
			"3306/tcp": {
				{
					HostIP:   "0.0.0.0",
					HostPort: hostPort,
				},
			},
		},
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		LogConfig: container.LogConfig{
			Type:   "json-file",
			Config: map[string]string{},
		},
	}

	resp, err := cs.cli.ContainerCreate(context.Background(), config, hostConfig, nil, nil, containerName)
	if err != nil {
		return fmt.Errorf("error creating container: %w", err)
	}

	if err := cs.cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("error starting container: %w", err)
	}

	return nil
}

func (cs *ContainerServices) getCWD(id string) string {
	cs.cwdMu.RLock()
	defer cs.cwdMu.RUnlock()

	if cwd, ok := cs.cwds[id]; ok {
		return cwd
	}

	return "/"
}

func (cs *ContainerServices) setCWD(id, cwd string) {
	cs.cwdMu.Lock()
	defer cs.cwdMu.Unlock()

	cs.cwds[id] = cwd
}

func (cs *ContainerServices) RemoveCWD(id string) {
	cs.cwdMu.Lock()
	defer cs.cwdMu.Unlock()

	delete(cs.cwds, id)
}

func (cs *ContainerServices) ExecCommandInContainer(id, cmd string) (string, string, error) {
	cwd := cs.getCWD(id)

	// We use a marker so that we can reliably separate the command's
	// output from the final working directory.
	const marker = "__HARBOR_CWD_7f3c9a__"

	script := fmt.Sprintf(
		`cd %s 2>/dev/null || exit $?; %s; printf '\n%s%%s\n' "$PWD"`,
		shellQuote(cwd),
		cmd,
		marker,
	)

	execConfig := container.ExecOptions{
		Cmd:          strslice.StrSlice([]string{"/bin/sh", "-c", script}),
		AttachStdout: true,
		AttachStderr: true,
	}

	resp, err := cs.cli.ContainerExecCreate(
		context.Background(),
		id,
		execConfig,
	)
	if err != nil {
		return "", cwd, err
	}

	attachResp, err := cs.cli.ContainerExecAttach(
		context.Background(),
		resp.ID,
		container.ExecStartOptions{},
	)
	if err != nil {
		return "", cwd, err
	}
	defer attachResp.Close()

	var output strings.Builder

	// Docker multiplexes stdout/stderr. stdcopy handles the Docker
	// headers for us.
	_, err = stdcopy.StdCopy(
		&output,
		&output,
		attachResp.Reader,
	)
	if err != nil {
		return "", cwd, err
	}

	result := output.String()

	newCWD := cwd

	if index := strings.LastIndex(result, "\n"+marker); index != -1 {
		commandOutput := result[:index]

		pathStart := index + len("\n"+marker)
		path := strings.TrimSpace(result[pathStart:])

		if path != "" {
			newCWD = path
		}

		result = strings.TrimRight(commandOutput, "\n")
	}

	cs.setCWD(id, newCWD)

	return result, newCWD, nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
func (cs *ContainerServices) GetContainerStats(containerID string) (dto.ContainerStats, error) {
	statsResponse, err := cs.cli.ContainerStats(context.Background(), containerID, false)
	if err != nil {
		return dto.ContainerStats{}, err
	}
	defer statsResponse.Body.Close()

	var v *types.StatsJSON
	dec := json.NewDecoder(statsResponse.Body)
	if err := dec.Decode(&v); err != nil {
		return dto.ContainerStats{}, err
	}

	cpuDelta := float64(v.CPUStats.CPUUsage.TotalUsage) - float64(v.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(v.CPUStats.SystemUsage) - float64(v.PreCPUStats.SystemUsage)
	numberOfCores := float64(v.CPUStats.OnlineCPUs)

	cpuPercentage := (cpuDelta / systemDelta) * numberOfCores * 100.0

	memoryUsageMB := float64(v.MemoryStats.Usage / 1024 / 1024)
	memoryLimitMB := float64(v.MemoryStats.Limit / 1024 / 1024)

	memoryUsagePercentage := (memoryUsageMB / memoryLimitMB) * 100

	stats := dto.ContainerStats{
		CPUPercentage: fmt.Sprintf("%.2f%%", cpuPercentage),
		MemoryUsageMB: fmt.Sprintf("%.2f%%", memoryUsagePercentage),
		MemoryLimitMB: fmt.Sprintf("%.2f", memoryLimitMB),
	}

	return stats, nil
}

func (cs *ContainerServices) PruneContainers() (string, error) {
	rep, err := cs.cli.ContainersPrune(context.Background(), filters.Args{})
	if err != nil {
		return "", err
	}

	res := fmt.Sprintf("Space reclaimed: %d bytes\n", rep.SpaceReclaimed)

	return res, nil
}

func (cs *ContainerServices) PruneImages() (string, error) {
	rep, err := cs.cli.ImagesPrune(context.Background(), filters.Args{})
	if err != nil {
		return "", err
	}

	res := fmt.Sprintf("Space reclaimed: %d bytes\n", rep.SpaceReclaimed)

	return res, nil
}

func (cs *ContainerServices) PruneVolumes() (string, error) {
	rep, err := cs.cli.VolumesPrune(context.Background(), filters.Args{})
	if err != nil {
		return "", err
	}

	res := fmt.Sprintf("Space reclaimed: %d bytes\n", rep.SpaceReclaimed)

	return res, nil
}

func (cs *ContainerServices) PruneNetworks() (string, error) {
	rep, err := cs.cli.NetworksPrune(context.Background(), filters.Args{})
	if err != nil {
		return "", err
	}

	res := ""
	for _, container := range rep.NetworksDeleted {
		res += fmt.Sprintf("Volume ID: %s\n", container)
	}

	return res, nil
}

func (cs *ContainerServices) PruneAll() error {
	_, err := cs.PruneContainers()
	if err != nil {
		return err
	}

	_, err = cs.PruneImages()
	if err != nil {
		return err
	}

	_, err = cs.PruneVolumes()
	if err != nil {
		return err
	}

	_, err = cs.PruneNetworks()
	if err != nil {
		return err
	}

	return nil
}

type IP struct {
	Query string
}

func getip2() string {
	req, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return err.Error()
	}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err.Error()
	}

	var ip IP
	json.Unmarshal(body, &ip)

	return ip.Query
}
