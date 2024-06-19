package dto

type ContainerDto struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ContainerStats struct {
	CPUPercentage string
	MemoryUsageMB string
	MemoryLimitMB string
}
