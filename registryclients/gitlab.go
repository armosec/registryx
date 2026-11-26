package registryclients

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/armosec/armoapi-go/armotypes"
	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/registries/defaultregistry"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	dockerregistry "github.com/moby/moby/api/types/registry"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type GitLabRegistryClient struct {
	Registry *armotypes.GitlabImageRegistry
	Options  *common.RegistryOptions
}

// GitLabProject represents a GitLab project from the API
type gitLabProject struct {
	ID                int    `json:"id"`
	PathWithNamespace string `json:"path_with_namespace"`
	// ContainerRegistryAccessLevel is "disabled", "private" or "enabled"; older GitLab
	// versions do not return it.
	ContainerRegistryAccessLevel string `json:"container_registry_access_level"`
}

// hasContainerRegistry reports whether a project may have container repositories.
// GitLab answers 403 for the repositories of a project whose registry is disabled, so
// those projects are skipped rather than read. A missing value counts as enabled.
func (p gitLabProject) hasContainerRegistry() bool {
	return p.ContainerRegistryAccessLevel != "disabled"
}

// GitLabRepository represents a container repository from GitLab API
type gitLabRepository struct {
	ID       int    `json:"id"`
	Path     string `json:"path"`
	Location string `json:"location"`
}

func (g *GitLabRegistryClient) GetAllRepositories(ctx context.Context) ([]string, error) {
	return g.getRepositoriesFromGitLabAPI(ctx)
}

func (g *GitLabRegistryClient) getRepositoriesFromGitLabAPI(ctx context.Context) ([]string, error) {
	return g.listRepositories(ctx, g.getGitLabAPIBaseURL())
}

// listRepositories lists the container repositories of every project the token can
// reach. A project whose repositories cannot be read fails the whole listing: callers
// use the result as the complete repository list, so a partial one must not pass as
// success. A project without a container registry (404) has no repositories.
func (g *GitLabRegistryClient) listRepositories(ctx context.Context, baseURL string) ([]string, error) {
	projects, err := g.getUserProjects(ctx, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user projects: %w", err)
	}

	var allRepos []string
	httpClient := &http.Client{}

	for _, project := range projects {
		if !project.hasContainerRegistry() {
			continue
		}
		repos, err := g.getProjectRepositories(ctx, httpClient, baseURL, project.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to list container repositories of project %s: %w", project.PathWithNamespace, err)
		}
		for _, repo := range repos {
			allRepos = append(allRepos, repo.Path)
		}
	}

	return allRepos, nil
}

func (g *GitLabRegistryClient) getGitLabAPIBaseURL() string {
	trimmed := strings.TrimSpace(g.Registry.RegistryURL)
	raw := trimmed
	if !strings.HasPrefix(strings.ToLower(trimmed), "https://") && !strings.HasPrefix(strings.ToLower(trimmed), "http://") {
		raw = "https://" + trimmed
	}
	parsedURL, err := url.Parse(raw)
	var host string
	if err == nil && parsedURL.Host != "" {
		host = parsedURL.Host
	} else {
		host = strings.TrimPrefix(trimmed, "https://")
		host = strings.TrimPrefix(host, "http://")
		for _, sep := range []string{"/", "?", "#"} {
			if idx := strings.Index(host, sep); idx != -1 {
				host = host[:idx]
			}
		}
	}

	host = strings.TrimPrefix(host, "registry.")

	if !hostLooksLikeGitLab(host) {
		host = "gitlab." + host
	}

	return fmt.Sprintf("https://%s/api/v4", host)
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

// getProjectRepositories returns all container repositories of a project, following
// pages: the API returns 20 per page by default. A 404 on the first page means the
// project has no container registry; a 404 on a later page is an error, so the pages
// already read are not passed off as the complete list.
func (g *GitLabRegistryClient) getProjectRepositories(ctx context.Context, httpClient *http.Client, baseURL string, projectID int) ([]gitLabRepository, error) {
	const perPage = 100
	var allRepos []gitLabRepository

	for page := 1; page > 0; {
		repos, nextPage, found, err := g.getProjectRepositoriesPage(ctx, httpClient, baseURL, projectID, page, perPage)
		if err != nil {
			return nil, err
		}
		if !found {
			if page == 1 {
				return []gitLabRepository{}, nil
			}
			return nil, fmt.Errorf("GitLab API error: 404 on page %d of the project's container repositories", page)
		}
		allRepos = append(allRepos, repos...)
		page = nextPage
	}
	return allRepos, nil
}

// getProjectRepositoriesPage reads one page of a project's container repositories.
// nextPage is 0 on the last page: it comes from GitLab's X-Next-Page header (empty on the
// last page), or, when the header is absent, from whether the page was full. found is
// false when the page answered 404.
func (g *GitLabRegistryClient) getProjectRepositoriesPage(ctx context.Context, httpClient *http.Client, baseURL string, projectID, page, perPage int) (repos []gitLabRepository, nextPage int, found bool, err error) {
	url := fmt.Sprintf("%s/projects/%d/registry/repositories?page=%d&per_page=%d", baseURL, projectID, page, perPage)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, false, err
	}

	req.Header.Set("PRIVATE-TOKEN", g.Registry.AccessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, 0, false, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, false, fmt.Errorf("GitLab API error: %d - %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, 0, false, err
	}

	if values, ok := resp.Header["X-Next-Page"]; ok {
		if next, convErr := strconv.Atoi(strings.TrimSpace(values[0])); convErr == nil && next > page {
			return repos, next, true, nil
		}
		return repos, 0, true, nil
	}
	if len(repos) == perPage {
		return repos, page + 1, true, nil
	}
	return repos, 0, true, nil
}

// discoverRegistryHost queries the GitLab API to find the actual container registry
// hostname from the `location` field of a registry repository. This handles self-hosted
// GitLab instances where the registry hostname differs from the GitLab web URL
// (e.g. "gitlab-reg.example.com" vs "gitlab.example.com").
func (g *GitLabRegistryClient) discoverRegistryHost(ctx context.Context, baseURL string) string {
	if len(g.Registry.Repositories) == 0 {
		return ""
	}

	projects, err := g.getUserProjects(ctx, baseURL)
	if err != nil {
		return ""
	}

	selectedSet := make(map[string]struct{}, len(g.Registry.Repositories))
	for _, r := range g.Registry.Repositories {
		selectedSet[r] = struct{}{}
	}

	httpClient := &http.Client{}
	for _, project := range projects {
		if !project.hasContainerRegistry() {
			continue
		}
		repos, err := g.getProjectRepositories(ctx, httpClient, baseURL, project.ID)
		if err != nil {
			continue
		}
		for _, repo := range repos {
			if _, ok := selectedSet[repo.Path]; ok && repo.Location != "" {
				// Location format: "registry-host.example.com/group/project"
				loc := repo.Location
				for _, prefix := range []string{"https://", "http://"} {
					loc = strings.TrimPrefix(loc, prefix)
				}
				if host, _, ok := strings.Cut(loc, "/"); ok && host != "" {
					return host
				}
			}
		}
	}

	return ""
}

func (g *GitLabRegistryClient) GetImagesToScan(ctx context.Context) (map[string]string, error) {
	// Try to discover the actual container registry hostname via the GitLab API.
	// Self-hosted GitLab instances can have a separate registry hostname
	// (e.g. "gitlab-reg.example.com") that differs from the GitLab web URL.
	registryHost := g.discoverRegistryHost(ctx, g.getGitLabAPIBaseURL())
	if registryHost == "" {
		registryHost = g.Registry.RegistryURL
	}

	registry, err := name.NewRegistry(registryHost)
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
			images[fmt.Sprintf("%s/%s", registryHost, repository)] = tag
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
