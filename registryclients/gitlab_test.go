package registryclients

import (
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
			name:        "Host ends with gitlab - contains gitlab, use as-is",
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
