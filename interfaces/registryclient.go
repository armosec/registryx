package interfaces

import (
	"context"
	dockerregistry "github.com/docker/docker/api/types/registry"
)

type RegistryClient interface {
	GetAllRepositories(ctx context.Context) ([]string, error)
	GetImagesToScan(ctx context.Context) (map[string]string, error)
	GetDockerAuth() (*dockerregistry.AuthConfig, error)
}
