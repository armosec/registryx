package registryclients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
	// Use GitLab Projects API instead of Docker Registry v2 _catalog endpoint
	// because GitLab doesn't support _catalog for personal access tokens
	return g.getRepositoriesFromGitLabAPI(ctx)
}

// getRepositoriesFromGitLabAPI uses GitLab's REST API to list container repositories
func (g *GitLabRegistryClient) getRepositoriesFromGitLabAPI(ctx context.Context) ([]string, error) {
	baseURL := "https://gitlab.com/api/v4"

	// Get all projects for the user
	projects, err := g.getUserProjects(ctx, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user projects: %w", err)
	}

	var allRepos []string
	httpClient := &http.Client{}

	// For each project, get its container repositories
	for _, project := range projects {
		repos, err := g.getProjectRepositories(ctx, httpClient, baseURL, project.ID)
		if err != nil {
			// Continue if a single project fails (e.g., no registry enabled)
			continue
		}
		for _, repo := range repos {
			allRepos = append(allRepos, repo.Path)
		}
	}

	return allRepos, nil
}

// getUserProjects gets all projects owned by or accessible to the authenticated user
func (g *GitLabRegistryClient) getUserProjects(ctx context.Context, baseURL string) ([]gitLabProject, error) {
	var allProjects []gitLabProject
	page := 1
	perPage := 100
	httpClient := &http.Client{}

	for {
		// Use /projects endpoint which returns projects the authenticated user (via token) has access to
		// This works better than /users/{username}/projects for personal access tokens
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

		// Check if there are more pages
		if len(projects) < perPage {
			break
		}
		page++
	}

	return allProjects, nil
}

// getProjectRepositories gets all container repositories for a specific project
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
		// Project has no container registry
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
