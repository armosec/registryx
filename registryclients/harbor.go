package registryclients

import (
	"context"
	"fmt"
	"github.com/armosec/armoapi-go/armotypes"
	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/registries/harbor"
	dockerregistry "github.com/docker/docker/api/types/registry"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

type HarborRegistryClient struct {
	Registry *armotypes.HarborImageRegistry
}

func (h *HarborRegistryClient) GetAllRepositories(ctx context.Context) ([]string, error) {
	registry, err := name.NewRegistry(h.Registry.InstanceURL)
	if err != nil {
		return nil, err
	}
	iRegistry, err := harbor.NewHarborRegistry(&authn.AuthConfig{Username: h.Registry.Username, Password: h.Registry.Password}, &registry, &common.RegistryOptions{})
	if err != nil {
		return nil, err
	}

	return getAllRepositories(ctx, iRegistry)
}

func (h *HarborRegistryClient) GetImagesToScan(_ context.Context) (map[string]string, error) {
	registry, err := name.NewRegistry(h.Registry.InstanceURL)
	if err != nil {
		return nil, err
	}
	iRegistry, err := harbor.NewHarborRegistry(&authn.AuthConfig{Username: h.Registry.Username, Password: h.Registry.Password}, &registry, &common.RegistryOptions{})
	if err != nil {
		return nil, err
	}

	images := make(map[string]string, len(h.Registry.Repositories))
	for _, repository := range h.Registry.Repositories {
		tag, err := getImageLatestTag(repository, iRegistry)
		if err != nil {
			return nil, err
		} else if tag == "" {
			return nil, fmt.Errorf("failed to find latest tag for repository %s", repository)
		}
		images[fmt.Sprintf("%s/%s", h.Registry.InstanceURL, repository)] = tag
	}
	return images, nil
}

func (h *HarborRegistryClient) GetDockerAuth() *dockerregistry.AuthConfig {
	return &dockerregistry.AuthConfig{
		Username: h.Registry.Username,
		Password: h.Registry.Password,
	}
}
