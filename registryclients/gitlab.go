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
	ID       int    `json:"id"`
	Path     string `json:"path"`
	Location string `json:"location"`
}

func (g *GitLabRegistryClient) GetAllRepositories(ctx context.Context) ([]string, error) {
	return g.getRepositoriesFromGitLabAPI(ctx)
}

func (g *GitLabRegistryClient) getRepositoriesFromGitLabAPI(ctx context.Context) ([]string, error) {
	baseURL := g.getGitLabAPIBaseURL()

	projects, err := g.getUserProjects(ctx, baseURL)
	if err != nil {
		if rawBaseURL := g.getRawAPIBaseURL(); rawBaseURL != baseURL {
			projects, err = g.getUserProjects(ctx, rawBaseURL)
			if err != nil {
				return nil, fmt.Errorf("failed to get user projects: %w", err)
			}
			baseURL = rawBaseURL
		} else {
			return nil, fmt.Errorf("failed to get user projects: %w", err)
		}
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

// getRawAPIBaseURL builds the GitLab API base URL from RegistryURL as-given,
// preserving the original hostname without heuristic transformations. This is
// used for registry host discovery, where the user-entered URL should be tried
// verbatim before falling back to heuristic transforms.
func (g *GitLabRegistryClient) getRawAPIBaseURL() string {
	raw := strings.TrimSpace(g.Registry.RegistryURL)
	if lower := strings.ToLower(raw); !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		raw = "https://" + raw
	}
	if u, err := url.Parse(raw); err == nil && u.Host != "" {
		return fmt.Sprintf("%s://%s/api/v4", u.Scheme, u.Host)
	}
	host := strings.TrimPrefix(strings.TrimPrefix(raw, "https://"), "http://")
	if idx := strings.IndexAny(host, "/?#"); idx != -1 {
		host = host[:idx]
	}
	return fmt.Sprintf("https://%s/api/v4", host)
}

// extractRegistryHost returns the bare hostname from RegistryURL with scheme,
// path, query, and fragment stripped. Used as fallback when discovery fails.
func (g *GitLabRegistryClient) extractRegistryHost() string {
	raw := strings.TrimSpace(g.Registry.RegistryURL)
	raw = strings.TrimPrefix(raw, "https://")
	raw = strings.TrimPrefix(raw, "http://")
	if idx := strings.IndexAny(raw, "/?#"); idx != -1 {
		raw = raw[:idx]
	}
	return raw
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
	// (e.g. "gitlab-si-reg.hefr.ch") that differs from the GitLab web URL
	// (e.g. "gitlab-si.hefr.ch"). Using the wrong host causes Docker auth to
	// request service=dependency_proxy instead of service=container_registry → 403.
	//
	// Try the raw URL first (preserves the user-entered host verbatim), then fall
	// back to the heuristic URL (strips "registry.", prepends "gitlab.") which
	// handles cases like "registry.gitlab.example.com" → "gitlab.example.com".
	rawBaseURL := g.getRawAPIBaseURL()
	registryHost := g.discoverRegistryHost(ctx, rawBaseURL)
	if registryHost == "" {
		if heuristicBaseURL := g.getGitLabAPIBaseURL(); heuristicBaseURL != rawBaseURL {
			registryHost = g.discoverRegistryHost(ctx, heuristicBaseURL)
		}
	}
	if registryHost == "" {
		registryHost = g.extractRegistryHost()
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
