package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/armosec/registryx/common"
)

// GitLabClient implements registry client for GitLab Container Registry
type GitLabClient struct {
	baseURL     string
	username    string
	accessToken string
	httpClient  *http.Client
	options     *common.RegistryOptions
}

// GitLabProject represents a GitLab project from the API
type GitLabProject struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
}

// GitLabRepository represents a container repository from GitLab API
type GitLabRepository struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Location string `json:"location"`
}

// GitLabTag represents an image tag from GitLab API
type GitLabTag struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Location  string `json:"location"`
	CreatedAt string `json:"created_at"`
}

// NewGitLabClient creates a new GitLab registry client
func NewGitLabClient(registryURL, username, accessToken string, options *common.RegistryOptions) *GitLabClient {
	if options == nil {
		options = &common.RegistryOptions{}
	}

	return &GitLabClient{
		baseURL:     "https://gitlab.com/api/v4",
		username:    username,
		accessToken: accessToken,
		httpClient:  &http.Client{},
		options:     options,
	}
}

// GetAllRepositories retrieves all container repositories for the user
func (c *GitLabClient) GetAllRepositories(ctx context.Context) ([]string, error) {
	projects, err := c.getUserProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user projects: %w", err)
	}

	var allRepos []string
	for _, project := range projects {
		repos, err := c.getProjectRepositories(ctx, project.ID)
		if err != nil {
			// Continue if a single project fails
			continue
		}
		for _, repo := range repos {
			// Format: username/project-name
			allRepos = append(allRepos, repo.Path)
		}
	}

	return allRepos, nil
}

// getUserProjects gets all projects owned by the user
func (c *GitLabClient) getUserProjects(ctx context.Context) ([]GitLabProject, error) {
	var allProjects []GitLabProject
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("%s/users/%s/projects?page=%d&per_page=%d&min_access_level=30",
			c.baseURL, c.username, page, perPage)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("PRIVATE-TOKEN", c.accessToken)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GitLab API error: %d - %s", resp.StatusCode, string(body))
		}

		var projects []GitLabProject
		if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
			return nil, err
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
func (c *GitLabClient) getProjectRepositories(ctx context.Context, projectID int) ([]GitLabRepository, error) {
	url := fmt.Sprintf("%s/projects/%d/registry/repositories", c.baseURL, projectID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("PRIVATE-TOKEN", c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Project has no container registry
		return []GitLabRepository{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitLab API error: %d - %s", resp.StatusCode, string(body))
	}

	var repos []GitLabRepository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
}

// GetRepositoryTags gets all tags for a specific repository
func (c *GitLabClient) GetRepositoryTags(ctx context.Context, projectID, repositoryID int) ([]string, error) {
	url := fmt.Sprintf("%s/projects/%d/registry/repositories/%d/tags", c.baseURL, projectID, repositoryID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("PRIVATE-TOKEN", c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitLab API error: %d - %s", resp.StatusCode, string(body))
	}

	var tags []GitLabTag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	var tagNames []string
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}

	return tagNames, nil
}

// ParseRegistryURL extracts registry domain from URL
func ParseRegistryURL(registryURL string) string {
	registryURL = strings.TrimPrefix(registryURL, "https://")
	registryURL = strings.TrimPrefix(registryURL, "http://")
	registryURL = strings.TrimSuffix(registryURL, "/")

	if u, err := url.Parse("https://" + registryURL); err == nil {
		return u.Host
	}

	return registryURL
}

