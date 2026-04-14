package registryclients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
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

func TestGitLabRegistryClient_discoverRegistryHost(t *testing.T) {
	tests := []struct {
		name              string
		selectedRepos     []string
		apiProjects       []gitLabProject
		apiRepos          map[int][]gitLabRepository
		wantHost          string
		wantFallbackToURL bool
	}{
		{
			name:          "discovers non-standard registry hostname from location field",
			selectedRepos: []string{"team-exploitation/kubernetes/sftpgo"},
			apiProjects:   []gitLabProject{{ID: 1, PathWithNamespace: "team-exploitation/kubernetes"}},
			apiRepos: map[int][]gitLabRepository{
				1: {{ID: 10, Path: "team-exploitation/kubernetes/sftpgo", Location: "gitlab-reg.example.com/team-exploitation/kubernetes/sftpgo"}},
			},
			wantHost:          "gitlab-reg.example.com",
			wantFallbackToURL: false,
		},
		{
			name:          "discovers standard registry hostname from location field",
			selectedRepos: []string{"mygroup/myproject"},
			apiProjects:   []gitLabProject{{ID: 2, PathWithNamespace: "mygroup"}},
			apiRepos: map[int][]gitLabRepository{
				2: {{ID: 20, Path: "mygroup/myproject", Location: "registry.gitlab.example.com/mygroup/myproject"}},
			},
			wantHost:          "registry.gitlab.example.com",
			wantFallbackToURL: false,
		},
		{
			name:          "falls back to RegistryURL when no repos match",
			selectedRepos: []string{"group/nonexistent"},
			apiProjects:   []gitLabProject{{ID: 3, PathWithNamespace: "group"}},
			apiRepos: map[int][]gitLabRepository{
				3: {{ID: 30, Path: "group/other-repo", Location: "registry.gitlab.example.com/group/other-repo"}},
			},
			wantHost:          "",
			wantFallbackToURL: true,
		},
		{
			name:          "falls back when location field is empty",
			selectedRepos: []string{"group/myproject"},
			apiProjects:   []gitLabProject{{ID: 4, PathWithNamespace: "group"}},
			apiRepos: map[int][]gitLabRepository{
				4: {{ID: 40, Path: "group/myproject", Location: ""}},
			},
			wantHost:          "",
			wantFallbackToURL: true,
		},
		{
			name:              "returns empty string when no repositories are selected",
			selectedRepos:     []string{},
			apiProjects:       nil,
			apiRepos:          nil,
			wantHost:          "",
			wantFallbackToURL: true,
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

			reg := &armotypes.GitlabImageRegistry{
				RegistryURL: srv.URL,
			}
			reg.Repositories = tt.selectedRepos
			client := &GitLabRegistryClient{
				Registry: reg,
				Options:  &common.RegistryOptions{},
			}

			got, err := client.discoverRegistryHost(context.Background())

			if tt.wantFallbackToURL {
				if err != nil {
					t.Fatalf("discoverRegistryHost() unexpected error in fallback case: %v", err)
				}
				if got != "" {
					t.Errorf("discoverRegistryHost() = %q, want empty string (fallback)", got)
				}
			} else {
				if err != nil {
					t.Fatalf("discoverRegistryHost() unexpected error: %v", err)
				}
				if got != tt.wantHost {
					t.Errorf("discoverRegistryHost() = %q, want %q", got, tt.wantHost)
				}
			}
		})
	}
}

func TestGitLabRegistryClient_resolveRegistryHost(t *testing.T) {
	const discoveredHost = "gitlab-si-reg.example.test"
	const repoPath = "group/myproject"

	makeAPIServer := func(t *testing.T, location string) (*httptest.Server, *atomic.Int32) {
		t.Helper()
		var hits atomic.Int32
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
			hits.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]gitLabProject{{ID: 1, PathWithNamespace: "group"}})
		})
		mux.HandleFunc("/api/v4/projects/1/registry/repositories", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]gitLabRepository{{ID: 10, Path: repoPath, Location: location}})
		})
		return httptest.NewServer(mux), &hits
	}

	t.Run("prefers discovered host over RegistryURL", func(t *testing.T) {
		srv, _ := makeAPIServer(t, fmt.Sprintf("%s/%s", discoveredHost, repoPath))
		defer srv.Close()

		reg := &armotypes.GitlabImageRegistry{RegistryURL: srv.URL}
		reg.Repositories = []string{repoPath}
		client := &GitLabRegistryClient{Registry: reg, Options: &common.RegistryOptions{}}

		got := client.resolveRegistryHost(context.Background())
		if got != discoveredHost {
			t.Errorf("resolveRegistryHost() = %q, want %q", got, discoveredHost)
		}
	})

	t.Run("falls back to RegistryURL when discovery fails", func(t *testing.T) {
		// Use an already-closed server to force a discovery error.
		srv, _ := makeAPIServer(t, "")
		srv.Close()

		reg := &armotypes.GitlabImageRegistry{RegistryURL: srv.URL}
		reg.Repositories = []string{repoPath}
		client := &GitLabRegistryClient{Registry: reg, Options: &common.RegistryOptions{}}

		got := client.resolveRegistryHost(context.Background())
		// The fallback strips the scheme from srv.URL (e.g. "http://127.0.0.1:PORT").
		wantHost := extractHostFromLocation(srv.URL)
		if got != wantHost {
			t.Errorf("resolveRegistryHost() = %q, want %q (normalized RegistryURL)", got, wantHost)
		}
	})

	t.Run("strips scheme from RegistryURL in fallback", func(t *testing.T) {
		// Server is immediately closed so discovery returns an error.
		srv, _ := makeAPIServer(t, "")
		srv.Close()

		reg := &armotypes.GitlabImageRegistry{RegistryURL: "https://gitlab.example.com"}
		reg.Repositories = []string{repoPath}
		client := &GitLabRegistryClient{Registry: reg, Options: &common.RegistryOptions{}}

		got := client.resolveRegistryHost(context.Background())
		if got != "gitlab.example.com" {
			t.Errorf("resolveRegistryHost() = %q, want %q", got, "gitlab.example.com")
		}
	})

	t.Run("caches result and avoids redundant API calls", func(t *testing.T) {
		srv, hits := makeAPIServer(t, fmt.Sprintf("%s/%s", discoveredHost, repoPath))
		defer srv.Close()

		reg := &armotypes.GitlabImageRegistry{RegistryURL: srv.URL}
		reg.Repositories = []string{repoPath}
		client := &GitLabRegistryClient{Registry: reg, Options: &common.RegistryOptions{}}

		first := client.resolveRegistryHost(context.Background())
		second := client.resolveRegistryHost(context.Background())

		if first != discoveredHost || second != discoveredHost {
			t.Errorf("resolveRegistryHost() = %q, %q; want %q both times", first, second, discoveredHost)
		}
		if n := hits.Load(); n != 1 {
			t.Errorf("GitLab API called %d time(s), want exactly 1 (cached after first call)", n)
		}
	})
}

func TestGitLabRegistryClient_GetImagesToScan_usesDiscoveredHost(t *testing.T) {
	const discoveredHost = "gitlab-reg.example.test"
	const repoPath = "group/myproject"

	// Mock GitLab API that returns location pointing to discoveredHost
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

	reg := &armotypes.GitlabImageRegistry{
		RegistryURL: srv.URL,
	}
	reg.Repositories = []string{repoPath}
	client := &GitLabRegistryClient{
		Registry: reg,
		Options:  &common.RegistryOptions{},
	}

	// GetImagesToScan will discover the registry host, then fail when trying to
	// list tags from the unreachable discovered host. The error should reference
	// the discovered host, proving GetImagesToScan used it instead of RegistryURL.
	_, err := client.GetImagesToScan(context.Background())
	if err == nil {
		t.Fatal("expected error (discovered registry unreachable), got nil")
	}
	if !strings.Contains(err.Error(), discoveredHost) {
		t.Errorf("error should reference discovered host %q, got: %v", discoveredHost, err)
	}
}
