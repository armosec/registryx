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

type AzureRegistryClient struct {
	Registry *armotypes.AzureImageRegistry
}

func (a *AzureRegistryClient) GetAllRepositories(ctx context.Context) ([]string, error) {
	registry, err := name.NewRegistry(a.Registry.LoginServer)
	if err != nil {
		return nil, err
	}
	iRegistry, err := defaultregistry.NewRegistry(&authn.AuthConfig{Username: a.Registry.Username, Password: a.Registry.AccessToken}, &registry, &common.RegistryOptions{})
	if err != nil {
		return nil, err
	}

	return getAllRepositories(ctx, iRegistry)
}

func (a *AzureRegistryClient) GetImagesToScan(_ context.Context) (map[string]string, error) {
	registry, err := name.NewRegistry(a.Registry.LoginServer)
	if err != nil {
		return nil, err
	}
	iRegistry, err := defaultregistry.NewRegistry(&authn.AuthConfig{Username: a.Registry.Username, Password: a.Registry.AccessToken}, &registry, &common.RegistryOptions{})
	if err != nil {
		return nil, err
	}
	images := make(map[string]string, len(a.Registry.Repositories))
	for _, repository := range a.Registry.Repositories {
		tag, err := getImageLatestTag(repository, iRegistry)
		if err != nil {
			return nil, err
		}
		if tag != "" {
			images[fmt.Sprintf("%s/%s", a.Registry.LoginServer, repository)] = tag
		}
	}
	return images, nil
}

func (a *AzureRegistryClient) GetDockerAuth() (*dockerregistry.AuthConfig, error) {
	return &dockerregistry.AuthConfig{
		Username: a.Registry.Username,
		Password: a.Registry.AccessToken,
	}, nil
}
