package registryclients

import (
	"context"
	"fmt"

	"github.com/armosec/armoapi-go/armotypes"
	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/registries/defaultregistry"
	dockerregistry "github.com/docker/docker/api/types/registry"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

type NexusRegistryClient struct {
	Registry *armotypes.NexusImageRegistry
}

func (n *NexusRegistryClient) GetAllRepositories(ctx context.Context) ([]string, error) {
	registry, err := name.NewRegistry(n.Registry.RegistryURL)
	if err != nil {
		return nil, err
	}
	iRegistry, err := defaultregistry.NewRegistry(&authn.AuthConfig{Username: n.Registry.Username, Password: n.Registry.Password}, &registry, &common.RegistryOptions{})
	if err != nil {
		return nil, err
	}

	iRegistry.SetMaxPageSize(1000)
	return getAllRepositories(ctx, iRegistry)
}

func (n *NexusRegistryClient) GetImagesToScan(_ context.Context) (map[string]string, error) {
	registry, err := name.NewRegistry(n.Registry.RegistryURL)
	if err != nil {
		return nil, err
	}
	iRegistry, err := defaultregistry.NewRegistry(&authn.AuthConfig{Username: n.Registry.Username, Password: n.Registry.Password}, &registry, &common.RegistryOptions{})
	if err != nil {
		return nil, err
	}

	images := make(map[string]string, len(n.Registry.Repositories))
	for _, repository := range n.Registry.Repositories {
		tag, err := getImageLatestTag(repository, iRegistry)
		if err != nil {
			return nil, err
		}
		if tag != "" {
			images[fmt.Sprintf("%s/%s", n.Registry.RegistryURL, repository)] = tag
		}
	}
	return images, nil
}

func (n *NexusRegistryClient) GetDockerAuth() (*dockerregistry.AuthConfig, error) {
	return &dockerregistry.AuthConfig{
		Username: n.Registry.Username,
		Password: n.Registry.Password,
	}, nil
}
