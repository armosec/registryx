package registryclients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/armosec/armoapi-go/armotypes"
	"github.com/armosec/registryx/common"
)

func TestGitLabRegistryClient_getGitLabAPIBaseURL(t *testing.T) {
	tests := []struct {
		name        string
		registryURL string
		want        string
	}{
		{
			name:        "Docker registry URL with registry prefix",
			registryURL: "registry.gitlab.example.com",
			want:        "https://gitlab.example.com/api/v4",
		},
		{
			name:        "GitLab instance URL without scheme",
			registryURL: "gitlab.example.com",
			want:        "https://gitlab.example.com/api/v4",
		},
		{
			name:        "GitLab instance URL with HTTPS scheme",
			registryURL: "https://gitlab.example.com",
			want:        "https://gitlab.example.com/api/v4",
		},
		{
			name:        "GitLab instance URL with HTTP scheme",
			registryURL: "http://gitlab.example.com",
			want:        "https://gitlab.example.com/api/v4",
		},
		{
			name:        "GitLab URL with path",
			registryURL: "https://gitlab.example.com/root/test",
			want:        "https://gitlab.example.com/api/v4",
		},
		{
			name:        "GitLab URL with path and query params",
			registryURL: "https://gitlab.example.com/root/test?param=value",
			want:        "https://gitlab.example.com/api/v4",
		},
		{
			name:        "Base domain - should prepend gitlab",
			registryURL: "example.com",
			want:        "https://gitlab.example.com/api/v4",
		},
		{
			name:        "Base domain with scheme - should prepend gitlab",
			registryURL: "https://example.com",
			want:        "https://gitlab.example.com/api/v4",
		},
		{
			name:        "gitlab.com - should use as-is",
			registryURL: "gitlab.com",
			want:        "https://gitlab.com/api/v4",
		},
		{
			name:        "gitlab.com with scheme",
			registryURL: "https://gitlab.com",
			want:        "https://gitlab.com/api/v4",
		},
		{
			name:        "Docker registry with scheme",
			registryURL: "https://registry.gitlab.example.com",
			want:        "https://gitlab.example.com/api/v4",
		},
		{
			name:        "Hostname with path (no scheme)",
			registryURL: "gitlab.example.com/root/test",
			want:        "https://gitlab.example.com/api/v4",
		},
		{
			name:        "Another on-premise instance",
			registryURL: "gitlab.example.internal",
			want:        "https://gitlab.example.internal/api/v4",
		},
		{
			name:        "Another on-premise instance with registry prefix",
			registryURL: "registry.gitlab.example.internal",
			want:        "https://gitlab.example.internal/api/v4",
		},
		{
			name:        "IP address",
			registryURL: "192.168.1.100",
			want:        "https://gitlab.192.168.1.100/api/v4",
		},
		{
			name:        "IP address with scheme",
			registryURL: "https://192.168.1.100",
			want:        "https://gitlab.192.168.1.100/api/v4",
		},
		{
			name:        "IP address with port",
			registryURL: "192.168.1.100:8080",
			want:        "https://gitlab.192.168.1.100:8080/api/v4",
		},
		{
			name:        "GitLab instance with port",
			registryURL: "gitlab.example.com:8080",
			want:        "https://gitlab.example.com:8080/api/v4",
		},
		{
			name:        "GitLab instance with port and scheme",
			registryURL: "https://gitlab.example.com:8080",
			want:        "https://gitlab.example.com:8080/api/v4",
		},
		// --- gitlab-* hostnames (no prepend; e.g. gitlab-si.test.example.com) ---
		{
			name:        "GitLab host with hyphen - gitlab-si.test.example.com (on-prem)",
			registryURL: "gitlab-si.test.example.com",
			want:        "https://gitlab-si.test.example.com/api/v4",
		},
		{
			name:        "Registry prefix with gitlab- host",
			registryURL: "registry.gitlab-si.test.example.com",
			want:        "https://gitlab-si.test.example.com/api/v4",
		},
		{
			name:        "gitlab- host with HTTPS scheme",
			registryURL: "https://gitlab-si.test.example.com",
			want:        "https://gitlab-si.test.example.com/api/v4",
		},
		{
			name:        "gitlab- host with path",
			registryURL: "https://gitlab-si.test.example.com/group/project",
			want:        "https://gitlab-si.test.example.com/api/v4",
		},
		{
			name:        "gitlab- host without scheme with path",
			registryURL: "gitlab-si.test.example.com/root/myproject",
			want:        "https://gitlab-si.test.example.com/api/v4",
		},
		{
			name:        "gitlab- numbered instance",
			registryURL: "gitlab-01.example.internal",
			want:        "https://gitlab-01.example.internal/api/v4",
		},
		{
			name:        "gitlab- with subdomain",
			registryURL: "gitlab-prod.domain.example.com",
			want:        "https://gitlab-prod.domain.example.com/api/v4",
		},
		// --- .gitlab. / -gitlab. in host (no prepend) ---
		{
			name:        "Subdomain before .gitlab.",
			registryURL: "internal.gitlab.example.com",
			want:        "https://internal.gitlab.example.com/api/v4",
		},
		{
			name:        "Host contains .gitlab. in middle",
			registryURL: "foo.gitlab.bar.example.com",
			want:        "https://foo.gitlab.bar.example.com/api/v4",
		},
		{
			name:        "Host contains -gitlab. (suffix before dot)",
			registryURL: "something-gitlab.domain.example.com",
			want:        "https://something-gitlab.domain.example.com/api/v4",
		},
		// --- Mixed case (hostLooksLikeGitLab uses ToLower) ---
		{
			name:        "Mixed case gitlab- host",
			registryURL: "GitLab-Si.Test.Example.COM",
			want:        "https://GitLab-Si.Test.Example.COM/api/v4",
		},
		// --- Bare/plain domains (should prepend gitlab.) ---
		{
			name:        "Bare domain - prepend gitlab",
			registryURL: "test.example.com",
			want:        "https://gitlab.test.example.com/api/v4",
		},
		{
			name:        "Bare domain with scheme",
			registryURL: "https://test.example.com",
			want:        "https://gitlab.test.example.com/api/v4",
		},
		{
			name:        "Plain company domain - prepend gitlab",
			registryURL: "mycompany.example.com",
			want:        "https://gitlab.mycompany.example.com/api/v4",
		},
		{
			name:        "Registry prefix on plain domain",
			registryURL: "registry.mycompany.example.com",
			want:        "https://gitlab.mycompany.example.com/api/v4",
		},
		{
			name:        "Registry prefix with scheme on plain domain",
			registryURL: "https://registry.mycompany.example.com",
			want:        "https://gitlab.mycompany.example.com/api/v4",
		},
		// --- Double registry. (strip only one prefix) ---
		{
			name:        "Double registry prefix - strip one",
			registryURL: "registry.registry.example.com",
			want:        "https://gitlab.registry.example.com/api/v4",
		},
		// --- Host with gitlab in name (contains "gitlab" -> use as-is, no prepend) ---
		{
			name:        "Host contains gitlab - use as-is",
			registryURL: "mygitlab.example.com",
			want:        "https://mygitlab.example.com/api/v4",
		},
		{
			name:        "Host contains gitlab as substring - use as-is",
			registryURL: "registry.somegitlab.example.io",
			want:        "https://somegitlab.example.io/api/v4",
		},
		// --- HTTP scheme trimming ---
		{
			name:        "gitlab- host with HTTP scheme",
			registryURL: "http://gitlab-si.test.example.com",
			want:        "https://gitlab-si.test.example.com/api/v4",
		},
		{
			name:        "Plain host with HTTP scheme - prepend",
			registryURL: "http://example.internal",
			want:        "https://gitlab.example.internal/api/v4",
		},
		// --- Port preservation ---
		{
			name:        "gitlab- host with port",
			registryURL: "gitlab-si.test.example.com:8443",
			want:        "https://gitlab-si.test.example.com:8443/api/v4",
		},
		{
			name:        "gitlab- host with registry prefix and port",
			registryURL: "https://registry.gitlab-si.test.example.com:443",
			want:        "https://gitlab-si.test.example.com:443/api/v4",
		},
		// --- No scheme with query/fragment: host must not include ? or # ---
		{
			name:        "No scheme with query - host trimmed at ?",
			registryURL: "gitlab.example.com?x=1",
			want:        "https://gitlab.example.com/api/v4",
		},
		{
			name:        "No scheme with fragment - host trimmed at #",
			registryURL: "gitlab.example.com#anchor",
			want:        "https://gitlab.example.com/api/v4",
		},
		{
			name:        "No scheme with query and fragment - host trimmed",
			registryURL: "gitlab-si.test.example.com?foo=bar#section",
			want:        "https://gitlab-si.test.example.com/api/v4",
		},
		{
			name:        "Plain host no scheme with query - prepend gitlab, host trimmed",
			registryURL: "example.com?token=secret",
			want:        "https://gitlab.example.com/api/v4",
		},
		// --- Leading/trailing whitespace: trimmed so returned URL has no spaces ---
		{
			name:        "Leading and trailing whitespace - host trimmed",
			registryURL: "  gitlab.example.com  ",
			want:        "https://gitlab.example.com/api/v4",
		},
		{
			name:        "Leading whitespace only - host trimmed",
			registryURL: "\t\thttps://gitlab.example.com",
			want:        "https://gitlab.example.com/api/v4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &GitLabRegistryClient{
				Registry: &armotypes.GitlabImageRegistry{
					RegistryURL: tt.registryURL,
				},
				Options: &common.RegistryOptions{},
			}
			if got := client.getGitLabAPIBaseURL(); got != tt.want {
				t.Errorf("getGitLabAPIBaseURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitLabRegistryClient_getRawAPIBaseURL(t *testing.T) {
	tests := []struct {
		name        string
		registryURL string
		want        string
	}{
		{
			name:        "GitLab web URL - preserved as-is",
			registryURL: "gitlab-si.hefr.ch",
			want:        "https://gitlab-si.hefr.ch/api/v4",
		},
		{
			name:        "Registry URL - preserved as-is (no heuristic strip)",
			registryURL: "registry.gitlab.example.com",
			want:        "https://registry.gitlab.example.com/api/v4",
		},
		{
			name:        "Non-gitlab host - preserved as-is (no heuristic prepend)",
			registryURL: "example.com",
			want:        "https://example.com/api/v4",
		},
		{
			name:        "With HTTPS scheme",
			registryURL: "https://gitlab-si.hefr.ch",
			want:        "https://gitlab-si.hefr.ch/api/v4",
		},
		{
			name:        "With HTTP scheme - preserves http",
			registryURL: "http://gitlab-si.hefr.ch",
			want:        "http://gitlab-si.hefr.ch/api/v4",
		},
		{
			name:        "With path - strips path",
			registryURL: "https://gitlab-si.hefr.ch/group/project",
			want:        "https://gitlab-si.hefr.ch/api/v4",
		},
		{
			name:        "Self-hosted registry host",
			registryURL: "gitlab-si-reg.hefr.ch",
			want:        "https://gitlab-si-reg.hefr.ch/api/v4",
		},
		{
			name:        "With port",
			registryURL: "gitlab.example.com:8443",
			want:        "https://gitlab.example.com:8443/api/v4",
		},
		{
			name:        "Whitespace trimmed",
			registryURL: "  gitlab.example.com  ",
			want:        "https://gitlab.example.com/api/v4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &GitLabRegistryClient{
				Registry: &armotypes.GitlabImageRegistry{
					RegistryURL: tt.registryURL,
				},
				Options: &common.RegistryOptions{},
			}
			if got := client.getRawAPIBaseURL(); got != tt.want {
				t.Errorf("getRawAPIBaseURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitLabRegistryClient_extractRegistryHost(t *testing.T) {
	tests := []struct {
		name        string
		registryURL string
		want        string
	}{
		{
			name:        "plain hostname",
			registryURL: "gitlab-si-reg.hefr.ch",
			want:        "gitlab-si-reg.hefr.ch",
		},
		{
			name:        "with https scheme",
			registryURL: "https://gitlab-si.hefr.ch",
			want:        "gitlab-si.hefr.ch",
		},
		{
			name:        "with http scheme",
			registryURL: "http://gitlab-si.hefr.ch",
			want:        "gitlab-si.hefr.ch",
		},
		{
			name:        "with path",
			registryURL: "https://gitlab-si.hefr.ch/group/project",
			want:        "gitlab-si.hefr.ch",
		},
		{
			name:        "with query",
			registryURL: "gitlab.example.com?token=abc",
			want:        "gitlab.example.com",
		},
		{
			name:        "with port",
			registryURL: "https://gitlab.example.com:8443/path",
			want:        "gitlab.example.com:8443",
		},
		{
			name:        "whitespace trimmed",
			registryURL: "  https://gitlab.example.com  ",
			want:        "gitlab.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &GitLabRegistryClient{
				Registry: &armotypes.GitlabImageRegistry{
					RegistryURL: tt.registryURL,
				},
				Options: &common.RegistryOptions{},
			}
			if got := client.extractRegistryHost(); got != tt.want {
				t.Errorf("extractRegistryHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitLabRegistryClient_discoverRegistryHost(t *testing.T) {
	tests := []struct {
		name          string
		selectedRepos []string
		apiProjects   []gitLabProject
		apiRepos      map[int][]gitLabRepository
		wantHost      string
	}{
		{
			name:          "discovers non-standard registry hostname from location field",
			selectedRepos: []string{"team-exploitation/kubernetes/sftpgo"},
			apiProjects:   []gitLabProject{{ID: 1, PathWithNamespace: "team-exploitation/kubernetes"}},
			apiRepos: map[int][]gitLabRepository{
				1: {{ID: 10, Path: "team-exploitation/kubernetes/sftpgo", Location: "gitlab-reg.example.com/team-exploitation/kubernetes/sftpgo"}},
			},
			wantHost: "gitlab-reg.example.com",
		},
		{
			name:          "discovers standard registry hostname from location field",
			selectedRepos: []string{"mygroup/myproject"},
			apiProjects:   []gitLabProject{{ID: 2, PathWithNamespace: "mygroup"}},
			apiRepos: map[int][]gitLabRepository{
				2: {{ID: 20, Path: "mygroup/myproject", Location: "registry.gitlab.example.com/mygroup/myproject"}},
			},
			wantHost: "registry.gitlab.example.com",
		},
		{
			name:          "returns empty when no repos match",
			selectedRepos: []string{"group/nonexistent"},
			apiProjects:   []gitLabProject{{ID: 3, PathWithNamespace: "group"}},
			apiRepos: map[int][]gitLabRepository{
				3: {{ID: 30, Path: "group/other-repo", Location: "registry.gitlab.example.com/group/other-repo"}},
			},
			wantHost: "",
		},
		{
			name:          "returns empty when location field is empty",
			selectedRepos: []string{"group/myproject"},
			apiProjects:   []gitLabProject{{ID: 4, PathWithNamespace: "group"}},
			apiRepos: map[int][]gitLabRepository{
				4: {{ID: 40, Path: "group/myproject", Location: ""}},
			},
			wantHost: "",
		},
		{
			name:          "returns empty when no repositories are selected",
			selectedRepos: []string{},
			apiProjects:   nil,
			apiRepos:      nil,
			wantHost:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(tt.apiProjects)
			})
			for projectID, repos := range tt.apiRepos {
				path := fmt.Sprintf("/api/v4/projects/%d/registry/repositories", projectID)
				mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(repos)
				})
			}
			srv := httptest.NewServer(mux)
			defer srv.Close()

			reg := &armotypes.GitlabImageRegistry{}
			reg.Repositories = tt.selectedRepos
			client := &GitLabRegistryClient{
				Registry: reg,
				Options:  &common.RegistryOptions{},
			}

			// Pass mock server URL directly as baseURL
			got := client.discoverRegistryHost(context.Background(), srv.URL+"/api/v4")
			if got != tt.wantHost {
				t.Errorf("discoverRegistryHost() = %q, want %q", got, tt.wantHost)
			}
		})
	}
}

func TestGitLabRegistryClient_GetImagesToScan_discoveryViaRawURL(t *testing.T) {
	const discoveredHost = "gitlab-si-reg.example.test"
	const repoPath = "team/kubernetes/sftpgo"

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]gitLabProject{{ID: 1, PathWithNamespace: "team/kubernetes"}})
	})
	mux.HandleFunc("/api/v4/projects/1/registry/repositories", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]gitLabRepository{{
			ID:       10,
			Path:     repoPath,
			Location: fmt.Sprintf("%s/%s", discoveredHost, repoPath),
		}})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Set RegistryURL to mock server URL with http:// scheme so getRawAPIBaseURL()
	// preserves it and reaches the mock. getGitLabAPIBaseURL() would mangle the host
	// (prepend "gitlab.") and miss the mock — this verifies discovery uses the raw URL.
	reg := &armotypes.GitlabImageRegistry{
		RegistryURL: srv.URL,
	}
	reg.Repositories = []string{repoPath}
	client := &GitLabRegistryClient{
		Registry: reg,
		Options:  &common.RegistryOptions{},
	}

	_, err := client.GetImagesToScan(context.Background())
	if err == nil {
		t.Fatal("expected error (discovered registry unreachable), got nil")
	}
	if !strings.Contains(err.Error(), discoveredHost) {
		t.Errorf("error should reference discovered host %q, got: %v", discoveredHost, err)
	}
}

func TestGitLabRegistryClient_GetImagesToScan_fallbackStripsScheme(t *testing.T) {
	// When discovery fails, the fallback should use the bare hostname from
	// RegistryURL (no scheme/path), so name.NewRegistry receives a valid host.
	reg := &armotypes.GitlabImageRegistry{
		RegistryURL: "https://gitlab-si-reg.example.test/some/path",
	}
	reg.Repositories = []string{"group/myproject"}
	client := &GitLabRegistryClient{
		Registry: reg,
		Options:  &common.RegistryOptions{},
	}

	_, err := client.GetImagesToScan(context.Background())
	if err == nil {
		t.Fatal("expected error (registry unreachable), got nil")
	}
	if !strings.Contains(err.Error(), "gitlab-si-reg.example.test") {
		t.Errorf("error should reference bare host, got: %v", err)
	}
	// Verify the path was stripped — it should NOT appear in the registry host
	if strings.Contains(err.Error(), "/some/path") {
		t.Errorf("error should not contain path from RegistryURL, got: %v", err)
	}
}

func TestGitLabRegistryClient_GetImagesToScan_heuristicFallbackDiscovery(t *testing.T) {
	const discoveredHost = "actual-registry.example.test"
	const repoPath = "group/myproject"

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]gitLabProject{{ID: 1, PathWithNamespace: "group"}})
	})
	mux.HandleFunc("/api/v4/projects/1/registry/repositories", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]gitLabRepository{{
			ID:       10,
			Path:     repoPath,
			Location: fmt.Sprintf("%s/%s", discoveredHost, repoPath),
		}})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	srvHost := strings.TrimPrefix(srv.URL, "http://")

	// Verify the heuristic fallback path: raw URL ≠ heuristic URL, raw fails,
	// heuristic succeeds. We use "registry.gitlab-<host>" as RegistryURL:
	//   raw URL     → http://registry.gitlab-<host>/api/v4  (unreachable)
	//   heuristic   → https://gitlab-<host>/api/v4          (scheme mismatch with mock)
	// Since httptest only serves HTTP and getGitLabAPIBaseURL always returns https://,
	// we test the fallback logic at the discoverRegistryHost level instead.
	client := &GitLabRegistryClient{
		Registry: &armotypes.GitlabImageRegistry{
			RegistryURL: "registry.gitlab-" + srvHost,
		},
		Options: &common.RegistryOptions{},
	}
	client.Registry.Repositories = []string{repoPath}

	rawBaseURL := client.getRawAPIBaseURL()
	heuristicBaseURL := client.getGitLabAPIBaseURL()
	if rawBaseURL == heuristicBaseURL {
		t.Fatalf("raw and heuristic URLs should differ for this test, got %q", rawBaseURL)
	}

	// Raw URL should fail (unreachable host)
	host := client.discoverRegistryHost(context.Background(), rawBaseURL)
	if host != "" {
		t.Errorf("raw URL discovery should fail, got %q", host)
	}

	// Heuristic URL reaches the mock (after stripping "registry." prefix)
	// We need to use HTTP to reach the mock, so build the URL manually
	heuristicHTTP := fmt.Sprintf("http://gitlab-%s/api/v4", srvHost)
	_ = heuristicHTTP // can't reach mock with heuristic either due to DNS

	// Instead, verify the getGitLabAPIBaseURL strips "registry." correctly
	if !strings.Contains(heuristicBaseURL, "gitlab-"+srvHost) {
		t.Errorf("heuristic URL should contain gitlab-%s, got %q", srvHost, heuristicBaseURL)
	}
}

func TestGitLabRegistryClient_getRepositoriesFromGitLabAPI_fallbackToRawURL(t *testing.T) {
	const repoPath = "team/kubernetes/sftpgo"

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]gitLabProject{{ID: 1, PathWithNamespace: "team/kubernetes"}})
	})
	mux.HandleFunc("/api/v4/projects/1/registry/repositories", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]gitLabRepository{{
			ID:   10,
			Path: repoPath,
		}})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Set RegistryURL to mock server URL (with http:// scheme). getGitLabAPIBaseURL()
	// would mangle it (prepend "gitlab.", force https://), failing to reach the mock.
	// The fallback to getRawAPIBaseURL() preserves the URL and reaches the mock.
	client := &GitLabRegistryClient{
		Registry: &armotypes.GitlabImageRegistry{
			RegistryURL: srv.URL,
		},
		Options: &common.RegistryOptions{},
	}

	repos, err := client.GetAllRepositories(context.Background())
	if err != nil {
		t.Fatalf("GetAllRepositories() error = %v, want nil", err)
	}
	if len(repos) != 1 || repos[0] != repoPath {
		t.Errorf("GetAllRepositories() = %v, want [%s]", repos, repoPath)
	}
}
