package registryclients

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/armosec/armoapi-go/armotypes"
	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/registries/defaultregistry"
	dockerregistry "github.com/docker/docker/api/types/registry"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net/http"
)

const (
	gcpScope        = "https://www.googleapis.com/auth/cloud-platform"
	oauth2user      = "oauth2accesstoken"
	accessTokenAuth = "accesstoken"
)

type GoogleArtifactRegistryClient struct {
	Registry   *armotypes.GoogleImageRegistry
	httpClient *http.Client
	projectID  string
	ts         oauth2.TokenSource
}

func NewGoogleArtifactRegistryClient(registry *armotypes.GoogleImageRegistry) (*GoogleArtifactRegistryClient, error) {
	jsonData, err := json.Marshal(registry.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json key: %w", err)
	}
	creds, err := google.CredentialsFromJSON(context.Background(), jsonData, gcpScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return &GoogleArtifactRegistryClient{
		Registry:   registry,
		httpClient: oauth2.NewClient(context.Background(), creds.TokenSource),
		projectID:  creds.ProjectID,
		ts:         creds.TokenSource,
	}, nil
}

func (g *GoogleArtifactRegistryClient) GetAllRepositories(ctx context.Context) ([]string, error) {
	registry, err := name.NewRegistry(g.Registry.RegistryURI)
	if err != nil {
		return nil, err
	}
	token, err := g.ts.Token()
	if err != nil {
		return nil, err
	}
	iRegistry, err := defaultregistry.NewRegistry(&authn.AuthConfig{Username: oauth2user, Password: token.AccessToken}, &registry, &common.RegistryOptions{})
	if err != nil {
		return nil, err
	}

	return getAllRepositories(ctx, iRegistry)
}

func (g *GoogleArtifactRegistryClient) GetImagesToScan(_ context.Context) (map[string]string, error) {
	registry, err := name.NewRegistry(g.Registry.RegistryURI)
	if err != nil {
		return nil, err
	}
	token, err := g.ts.Token()
	if err != nil {
		return nil, err
	}
	iRegistry, err := defaultregistry.NewRegistry(&authn.AuthConfig{Username: oauth2user, Password: token.AccessToken}, &registry, &common.RegistryOptions{})
	if err != nil {
		return nil, err
	}

	images := make(map[string]string, len(g.Registry.Repositories))
	for _, repository := range g.Registry.Repositories {
		tag, err := getImageLatestTag(repository, iRegistry)
		if err != nil {
			return nil, err
		} else if tag == "" {
			return nil, fmt.Errorf("failed to find latest tag for repository %s", repository)
		}
		images[fmt.Sprintf("%s/%s", g.Registry.RegistryURI, repository)] = tag
	}
	return images, nil
}

func (g *GoogleArtifactRegistryClient) GetDockerAuth() (*dockerregistry.AuthConfig, error) {
	token, err := g.ts.Token()
	if err != nil {
		return nil, err
	}
	return &dockerregistry.AuthConfig{
		Username: oauth2user,
		Password: token.AccessToken,
		Auth:     accessTokenAuth,
	}, nil
}
