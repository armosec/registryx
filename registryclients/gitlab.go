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

type GitLabRegistryClient struct {
	Registry *armotypes.GitlabImageRegistry
	Options  *common.RegistryOptions
}

func (g *GitLabRegistryClient) GetAllRepositories(ctx context.Context) ([]string, error) {
	registry, err := name.NewRegistry(g.Registry.RegistryURL)
	if err != nil {
		return nil, err
	}
	iRegistry, err := defaultregistry.NewRegistry(&authn.AuthConfig{Username: g.Registry.Username, Password: g.Registry.AccessToken}, &registry, g.Options)
	if err != nil {
		return nil, err
	}

	iRegistry.SetMaxPageSize(1000)
	return getAllRepositories(ctx, iRegistry)
}

func (g *GitLabRegistryClient) GetImagesToScan(_ context.Context) (map[string]string, error) {
	registry, err := name.NewRegistry(g.Registry.RegistryURL)
	if err != nil {
		return nil, err
	}
	iRegistry, err := defaultregistry.NewRegistry(&authn.AuthConfig{Username: g.Registry.Username, Password: g.Registry.AccessToken}, &registry, g.Options)
	if err != nil {
		return nil, err
	}

	images := make(map[string]string, len(g.Registry.Repositories))
	for _, repository := range g.Registry.Repositories {
		tag, err := getImageLatestTag(repository, iRegistry)
		if err != nil {
			return nil, err
		}
		if tag != "" {
			images[fmt.Sprintf("%s/%s", g.Registry.RegistryURL, repository)] = tag
		}
	}
	return images, nil
}

func (g *GitLabRegistryClient) GetDockerAuth() (*dockerregistry.AuthConfig, error) {
	return &dockerregistry.AuthConfig{
		Username: g.Registry.Username,
		Password: g.Registry.AccessToken,
	}, nil
}
