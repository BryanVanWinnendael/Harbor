package dto

type ContainerEnvironmentVariable struct {
	Key      string
	Value    string
	IsSecret bool
}

type ContainerEnvironment struct {
	ContainerID   string
	ContainerName string
	Variables     []*ContainerEnvironmentVariable
}
