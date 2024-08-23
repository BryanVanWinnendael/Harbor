package services

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"

	"github.com/BryanVanWinnendael/Harbor/dto"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/shirou/gopsutil/disk"

	"github.com/docker/docker/client"
)

func NewAnalyticsServices(cli *client.Client) *AnalyticsServices {
	return &AnalyticsServices{
		cli: cli,
	}
}

type AnalyticsServices struct {
	cli *client.Client
}

func (is *AnalyticsServices) GetTotalUsage() (dto.UsageDTO, error) {
	usage, err := is.cli.DiskUsage(context.Background(), types.DiskUsageOptions{})
	if err != nil {
		return dto.UsageDTO{}, err
	}

	var totalDiskUsage int64

	var imageUsage int64
	var imageTotal int
	var imageActive int

	var containerUsage int64
	var containerTotal int
	var containerActive int

	var volumeUsage int64
	var volumeTotal int
	var volumeActive int

	var buildCacheUsage int64
	var buildCacheTotal int
	var buildCacheActive int

	totalDiskUsage += usage.LayersSize

	// Add the size of images
	for _, image := range usage.Images {
		imageUsage += image.Size
		if image.Containers > 0 {
			imageActive++
		}
	}
	imageTotal += len(usage.Images)

	// Add the size of containers
	for _, container := range usage.Containers {
		containerUsage += container.SizeRw
		if container.State == "running" {
			containerActive++
		}
	}
	containerTotal += len(usage.Containers)

	// Add the size of volumes
	for _, volume := range usage.Volumes {
		if volume.UsageData.Size != -1 {
			volumeUsage += volume.UsageData.Size
		}
		if volume.UsageData.RefCount != 0 {
			volumeActive++
		}
	}
	volumeTotal += len(usage.Volumes)

	// Add the size of build cache
	for _, cache := range usage.BuildCache {
		buildCacheUsage += cache.Size
		if cache.InUse {
			buildCacheActive++
		}
	}
	buildCacheTotal += len(usage.BuildCache)

	totalDiskUsage += containerUsage + imageUsage + volumeUsage + buildCacheUsage

	totalImageUsageMB := imageUsage / 1000 / 1000
	totalContainerUsageMB := containerUsage / 1000 / 1000
	totalVolumeUsageMB := volumeUsage / 1000 / 1000
	totalBuildCacheUsageMB := buildCacheUsage / 1000 / 1000
	totalDiskUsageMB := totalDiskUsage / 1000 / 1000

	diskStat, err := disk.Usage("/")
	if err != nil {
		log.Fatalf("Error getting disk usage: %v", err)
	}

	totalDiskSpaceMB := diskStat.Total / 1000 / 1000
	usagePercent := (totalDiskUsageMB * 100) / int64(totalDiskSpaceMB)

	result := dto.UsageDTO{
		Images:       []interface{}{strconv.FormatInt(totalImageUsageMB, 10), imageTotal, imageActive},
		Containers:   []interface{}{strconv.FormatInt(totalContainerUsageMB, 10), containerTotal, containerActive},
		Volumes:      []interface{}{strconv.FormatInt(totalVolumeUsageMB, 10), volumeTotal, volumeActive},
		BuildCache:   []interface{}{strconv.FormatInt(totalBuildCacheUsageMB, 10), buildCacheTotal, buildCacheActive},
		TotalUsage:   strconv.FormatInt(totalDiskUsageMB, 10),
		UsagePercent: strconv.FormatInt(usagePercent, 10),
	}

	return result, nil
}

func (is *AnalyticsServices) GetContainersCpuUsage() (dto.ContainersCpuUsageDTO, error) {
	containers, err := is.cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return dto.ContainersCpuUsageDTO{}, err
	}

	type usageResult struct {
		containerCpuUsage dto.ContainerCpuUsageDTO
		err               error
	}

	// Create channels to collect results
	results := make(chan usageResult, len(containers))

	// Use a sync.WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	for _, container := range containers {
		wg.Add(1)
		go func(container types.Container) {
			defer wg.Done()

			statsResponse, err := is.cli.ContainerStats(context.Background(), container.ID, false)
			if err != nil {
				results <- usageResult{dto.ContainerCpuUsageDTO{}, err}
				return
			}
			defer statsResponse.Body.Close()

			var stats *types.StatsJSON
			dec := json.NewDecoder(statsResponse.Body)
			if err := dec.Decode(&stats); err != nil {
				results <- usageResult{dto.ContainerCpuUsageDTO{}, err}
				return
			}

			cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)
			systemDelta := float64(stats.CPUStats.SystemUsage) - float64(stats.PreCPUStats.SystemUsage)
			numberOfCores := float64(stats.CPUStats.OnlineCPUs)

			cpuPercentage := (cpuDelta / systemDelta) * numberOfCores * 100.0

			results <- usageResult{
				containerCpuUsage: dto.ContainerCpuUsageDTO{
					ContainerName: container.Names[0],
					CpuUsage:      cpuPercentage,
				},
				err: nil,
			}
		}(container)
	}

	// Close the results channel when all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results as they come in
	var mostUsageContainer dto.ContainerCpuUsageDTO
	var leastUsageContainer dto.ContainerCpuUsageDTO
	var restUsageContainer []dto.ContainerCpuUsageDTO
	first := true

	for result := range results {
		if result.err != nil {
			return dto.ContainersCpuUsageDTO{}, result.err
		}

		containerCpuUsage := result.containerCpuUsage

		if first {
			mostUsageContainer = containerCpuUsage
			leastUsageContainer = containerCpuUsage
			first = false
		} else {
			if mostUsageContainer.CpuUsage < containerCpuUsage.CpuUsage {
				mostUsageContainer = containerCpuUsage
			} else if leastUsageContainer.CpuUsage > containerCpuUsage.CpuUsage {
				leastUsageContainer = containerCpuUsage
			} else {
				restUsageContainer = append(restUsageContainer, containerCpuUsage)
			}
		}
	}

	return dto.ContainersCpuUsageDTO{
		MostUsageContainer:  mostUsageContainer,
		LeastUsageContainer: leastUsageContainer,
		RestUsageContainer:  restUsageContainer,
	}, nil
}

func (is *AnalyticsServices) GetContainersMemoryUsage() (dto.ContainersMemoryUsageDTO, error) {
	containers, err := is.cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return dto.ContainersMemoryUsageDTO{}, err
	}

	type usageResult struct {
		containerMemoryUsage dto.ContainerMemoryUsageDTO
		err                  error
	}

	results := make(chan usageResult, len(containers))
	var wg sync.WaitGroup

	for _, container := range containers {
		wg.Add(1)
		go func(container types.Container) {
			defer wg.Done()

			statsResponse, err := is.cli.ContainerStats(context.Background(), container.ID, false)
			if err != nil {
				results <- usageResult{dto.ContainerMemoryUsageDTO{}, err}
				return
			}
			defer statsResponse.Body.Close()

			var stats *types.StatsJSON
			dec := json.NewDecoder(statsResponse.Body)
			if err := dec.Decode(&stats); err != nil {
				results <- usageResult{dto.ContainerMemoryUsageDTO{}, err}
				return
			}

			memoryUsageMB := float64(stats.MemoryStats.Usage) / 1024 / 1024
			memoryLimitMB := float64(stats.MemoryStats.Limit) / 1024 / 1024

			memoryUsagePercentage := (memoryUsageMB / memoryLimitMB) * 100

			results <- usageResult{
				containerMemoryUsage: dto.ContainerMemoryUsageDTO{
					ContainerName: container.Names[0],
					MemoryUsage:   memoryUsagePercentage,
				},
				err: nil,
			}
		}(container)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var mostUsageContainer dto.ContainerMemoryUsageDTO
	var leastUsageContainer dto.ContainerMemoryUsageDTO
	var restUsageContainer []dto.ContainerMemoryUsageDTO
	first := true

	for result := range results {
		if result.err != nil {
			return dto.ContainersMemoryUsageDTO{}, result.err
		}

		containerMemoryUsage := result.containerMemoryUsage

		if first {
			mostUsageContainer = containerMemoryUsage
			leastUsageContainer = containerMemoryUsage
			first = false
		} else {
			if mostUsageContainer.MemoryUsage < containerMemoryUsage.MemoryUsage {
				mostUsageContainer = containerMemoryUsage
			} else if leastUsageContainer.MemoryUsage > containerMemoryUsage.MemoryUsage {
				leastUsageContainer = containerMemoryUsage
			} else {
				restUsageContainer = append(restUsageContainer, containerMemoryUsage)
			}
		}
	}

	return dto.ContainersMemoryUsageDTO{
		MostUsageContainer:  mostUsageContainer,
		LeastUsageContainer: leastUsageContainer,
		RestUsageContainer:  restUsageContainer,
	}, nil
}

func (is *AnalyticsServices) GetContainersNetworkUsage() (dto.ContainersNetworkUsageDTO, error) {
	containers, err := is.cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return dto.ContainersNetworkUsageDTO{}, err
	}

	type networkUsageResult struct {
		totalReceived     float64
		totalSent         float64
		inPackets         uint64
		outPackets        uint64
		receivedErrors    uint64
		sentErrors        uint64
		inPacketsDropped  uint64
		outPacketsDropped uint64
		err               error
	}

	results := make(chan networkUsageResult, len(containers))
	var wg sync.WaitGroup

	for _, container := range containers {
		wg.Add(1)
		go func(container types.Container) {
			defer wg.Done()

			statsResponse, err := is.cli.ContainerStats(context.Background(), container.ID, false)
			if err != nil {
				results <- networkUsageResult{err: err}
				return
			}
			defer statsResponse.Body.Close()

			var stats *types.StatsJSON
			dec := json.NewDecoder(statsResponse.Body)
			if err := dec.Decode(&stats); err != nil {
				results <- networkUsageResult{err: err}
				return
			}

			var totalReceived, totalSent float64
			var inPackets, outPackets, receivedErrors, sentErrors, inPacketsDropped, outPacketsDropped uint64

			for _, network := range stats.Networks {
				totalReceived += float64(network.RxBytes)
				totalSent += float64(network.TxBytes)
				inPackets += network.RxPackets
				outPackets += network.TxPackets
				receivedErrors += network.RxErrors
				sentErrors += network.TxErrors
				inPacketsDropped += network.RxDropped
				outPacketsDropped += network.TxDropped
			}

			results <- networkUsageResult{
				totalReceived:     totalReceived,
				totalSent:         totalSent,
				inPackets:         inPackets,
				outPackets:        outPackets,
				receivedErrors:    receivedErrors,
				sentErrors:        sentErrors,
				inPacketsDropped:  inPacketsDropped,
				outPacketsDropped: outPacketsDropped,
				err:               nil,
			}
		}(container)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var totalReceived, totalSent float64
	var inPackets, outPackets, receivedErrors, sentErrors, inPacketsDropped, outPacketsDropped uint64

	for result := range results {
		if result.err != nil {
			return dto.ContainersNetworkUsageDTO{}, result.err
		}

		totalReceived += result.totalReceived
		totalSent += result.totalSent
		inPackets += result.inPackets
		outPackets += result.outPackets
		receivedErrors += result.receivedErrors
		sentErrors += result.sentErrors
		inPacketsDropped += result.inPacketsDropped
		outPacketsDropped += result.outPacketsDropped
	}

	return dto.ContainersNetworkUsageDTO{
		TotalReceived:     totalReceived / 1024 / 1024,
		TotalSent:         totalSent / 1024 / 1024,
		InPackets:         inPackets,
		OutPackets:        outPackets,
		ReceivedErrors:    receivedErrors,
		SentErrors:        sentErrors,
		InPacketsDropped:  inPacketsDropped,
		OutPacketsDropped: outPacketsDropped,
	}, nil
}
