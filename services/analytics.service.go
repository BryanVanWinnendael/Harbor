package services

import (
	"context"
	"log"
	"strconv"

	"github.com/BryanVanWinnendael/Harbor/dto"
	"github.com/docker/docker/api/types"
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
