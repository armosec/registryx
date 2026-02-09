package registryclients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

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

// GitLabProject represents a GitLab project from the API
type gitLabProject struct {
	ID                int    `json:"id"`
	PathWithNamespace string `json:"path_with_namespace"`
}

// GitLabRepository represents a container repository from GitLab API
type gitLabRepository struct {
	ID   int    `json:"id"`
	Path string `json:"path"`
}

func (g *GitLabRegistryClient) GetAllRepositories(ctx context.Context) ([]string, error) {
	return g.getRepositoriesFromGitLabAPI(ctx)
}

func (g *GitLabRegistryClient) getRepositoriesFromGitLabAPI(ctx context.Context) ([]string, error) {
	baseURL := g.getGitLabAPIBaseURL()

	projects, err := g.getUserProjects(ctx, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user projects: %w", err)
	}

	var allRepos []string
	httpClient := &http.Client{}

	for _, project := range projects {
		repos, err := g.getProjectRepositories(ctx, httpClient, baseURL, project.ID)
		if err != nil {
			continue
		}
		for _, repo := range repos {
			allRepos = append(allRepos, repo.Path)
		}
	}

	return allRepos, nil
}

func (g *GitLabRegistryClient) getGitLabAPIBaseURL() string {
	registryURL := g.Registry.RegistryURL

	parsedURL, err := url.Parse(registryURL)
	if err == nil && parsedURL.Host != "" {
		registryURL = parsedURL.Host
	} else {
		registryURL = strings.TrimPrefix(registryURL, "https://")
		registryURL = strings.TrimPrefix(registryURL, "http://")
		if idx := strings.Index(registryURL, "/"); idx != -1 {
			registryURL = registryURL[:idx]
		}
	}

	registryURL = strings.TrimPrefix(registryURL, "registry.")

	if !hostLooksLikeGitLab(registryURL) {
		registryURL = "gitlab." + registryURL
	}

	return fmt.Sprintf("https://%s/api/v4", registryURL)
}

func hostLooksLikeGitLab(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return false
	}
	return strings.Contains(host, "gitlab")
}

func (g *GitLabRegistryClient) getUserProjects(ctx context.Context, baseURL string) ([]gitLabProject, error) {
	var allProjects []gitLabProject
	page := 1
	perPage := 100
	httpClient := &http.Client{}

	for {
		url := fmt.Sprintf("%s/projects?page=%d&per_page=%d&min_access_level=30&membership=true",
			baseURL, page, perPage)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("PRIVATE-TOKEN", g.Registry.AccessToken)

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GitLab API error (status %d): %s", resp.StatusCode, string(body))
		}

		var projects []gitLabProject
		if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
			return nil, fmt.Errorf("failed to decode projects: %w", err)
		}

		if len(projects) == 0 {
			break
		}

		allProjects = append(allProjects, projects...)

		if len(projects) < perPage {
			break
		}
		page++
	}

	return allProjects, nil
}

func (g *GitLabRegistryClient) getProjectRepositories(ctx context.Context, httpClient *http.Client, baseURL string, projectID int) ([]gitLabRepository, error) {
	url := fmt.Sprintf("%s/projects/%d/registry/repositories", baseURL, projectID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("PRIVATE-TOKEN", g.Registry.AccessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []gitLabRepository{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitLab API error: %d - %s", resp.StatusCode, string(body))
	}

	var repos []gitLabRepository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
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
