package registryclients

import (
	"context"
	"fmt"
	"github.com/armosec/armoapi-go/armotypes"
	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/registries/quay"
	dockerregistry "github.com/docker/docker/api/types/registry"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

type QuayRegistryClient struct {
	Registry *armotypes.QuayImageRegistry
	Options  *common.RegistryOptions
}

func (q *QuayRegistryClient) GetAllRepositories(ctx context.Context) ([]string, error) {
	registry, err := name.NewRegistry(q.Registry.ContainerRegistryName)
	if err != nil {
		return nil, err
	}
	iRegistry, err := quay.NewQuayIORegistry(&authn.AuthConfig{Username: q.Registry.RobotAccountName, Password: q.Registry.RobotAccountToken}, &registry, q.Options)
	if err != nil {
		return nil, err
	}
	return getAllRepositories(ctx, iRegistry)
}

func (q *QuayRegistryClient) GetImagesToScan(_ context.Context) (map[string]string, error) {
	registry, err := name.NewRegistry(q.Registry.ContainerRegistryName)
	if err != nil {
		return nil, err
	}
	iRegistry, err := quay.NewQuayIORegistry(&authn.AuthConfig{Username: q.Registry.RobotAccountName, Password: q.Registry.RobotAccountToken}, &registry, q.Options)
	if err != nil {
		return nil, err
	}

	images := make(map[string]string, len(q.Registry.Repositories))
	for _, repository := range q.Registry.Repositories {
		tag, err := getImageLatestTag(repository, iRegistry)
		if err != nil {
			return nil, err
		}
		if tag != "" {
			images[fmt.Sprintf("%s/%s", q.Registry.ContainerRegistryName, repository)] = tag
		}
	}
	return images, nil
}

func (q *QuayRegistryClient) GetDockerAuth() (*dockerregistry.AuthConfig, error) {
	return &dockerregistry.AuthConfig{
		Username: q.Registry.RobotAccountName,
		Password: q.Registry.RobotAccountToken,
	}, nil
}
