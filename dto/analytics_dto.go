package dto

type UsageDTO struct {
	Images       []interface{} `json:"images"`
	Containers   []interface{} `json:"containers"`
	Volumes      []interface{} `json:"volumes"`
	BuildCache   []interface{} `json:"buildCache"`
	TotalUsage   string        `json:"totalUsage"`
	UsagePercent string        `json:"usagePercent"`
}
