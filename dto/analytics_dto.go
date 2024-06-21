package dto

type UsageDTO struct {
	Images       []interface{} `json:"images"`
	Containers   []interface{} `json:"containers"`
	Volumes      []interface{} `json:"volumes"`
	BuildCache   []interface{} `json:"buildCache"`
	TotalUsage   string        `json:"totalUsage"`
	UsagePercent string        `json:"usagePercent"`
}

type ContainerMemoryUsageDTO struct {
	ContainerName string  `json:"containerName"`
	MemoryUsage   float64 `json:"memoryUsageMB"`
}

type ContainersMemoryUsageDTO struct {
	MostUsageContainer  ContainerMemoryUsageDTO   `json:"mostUsageContainer"`
	LeastUsageContainer ContainerMemoryUsageDTO   `json:"leastUsageContainer"`
	RestUsageContainer  []ContainerMemoryUsageDTO `json:"restUsageContainer"`
}

type ContainerCpuUsageDTO struct {
	ContainerName string  `json:"containerName"`
	CpuUsage      float64 `json:"cpuUsage"`
}

type ContainersCpuUsageDTO struct {
	MostUsageContainer  ContainerCpuUsageDTO   `json:"mostUsageContainer"`
	LeastUsageContainer ContainerCpuUsageDTO   `json:"leastUsageContainer"`
	RestUsageContainer  []ContainerCpuUsageDTO `json:"restUsageContainer"`
}

type ContainersNetworkUsageDTO struct {
	TotalReceived     float64 `json:"totalReceived"`
	TotalSent         float64 `json:"totalSent"`
	InPackets         uint64  `json:"inPackets"`
	OutPackets        uint64  `json:"outPackets"`
	ReceivedErrors    uint64  `json:"receivedErrors"`
	SentErrors        uint64  `json:"sentErrors"`
	InPacketsDropped  uint64  `json:"inPacketsDropped"`
	OutPacketsDropped uint64  `json:"outPacketsDropped"`
}
