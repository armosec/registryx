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
			registryURL: "registry.gitlab.eudev1.cyberarmorsoft.com",
			want:        "https://gitlab.eudev1.cyberarmorsoft.com/api/v4",
		},
		{
			name:        "GitLab instance URL without scheme",
			registryURL: "gitlab.eudev1.cyberarmorsoft.com",
			want:        "https://gitlab.eudev1.cyberarmorsoft.com/api/v4",
		},
		{
			name:        "GitLab instance URL with HTTPS scheme",
			registryURL: "https://gitlab.eudev1.cyberarmorsoft.com",
			want:        "https://gitlab.eudev1.cyberarmorsoft.com/api/v4",
		},
		{
			name:        "GitLab instance URL with HTTP scheme",
			registryURL: "http://gitlab.eudev1.cyberarmorsoft.com",
			want:        "https://gitlab.eudev1.cyberarmorsoft.com/api/v4",
		},
		{
			name:        "GitLab URL with path",
			registryURL: "https://gitlab.eudev1.cyberarmorsoft.com/root/test",
			want:        "https://gitlab.eudev1.cyberarmorsoft.com/api/v4",
		},
		{
			name:        "GitLab URL with path and query params",
			registryURL: "https://gitlab.eudev1.cyberarmorsoft.com/root/test?param=value",
			want:        "https://gitlab.eudev1.cyberarmorsoft.com/api/v4",
		},
		{
			name:        "Base domain - should prepend gitlab",
			registryURL: "eudev1.cyberarmorsoft.com",
			want:        "https://gitlab.eudev1.cyberarmorsoft.com/api/v4",
		},
		{
			name:        "Base domain with scheme - should prepend gitlab",
			registryURL: "https://eudev1.cyberarmorsoft.com",
			want:        "https://gitlab.eudev1.cyberarmorsoft.com/api/v4",
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
			registryURL: "https://registry.gitlab.eudev1.cyberarmorsoft.com",
			want:        "https://gitlab.eudev1.cyberarmorsoft.com/api/v4",
		},
		{
			name:        "Hostname with path (no scheme)",
			registryURL: "gitlab.eudev1.cyberarmorsoft.com/root/test",
			want:        "https://gitlab.eudev1.cyberarmorsoft.com/api/v4",
		},
		{
			name:        "Another on-premise instance",
			registryURL: "gitlab.company.internal",
			want:        "https://gitlab.company.internal/api/v4",
		},
		{
			name:        "Another on-premise instance with registry prefix",
			registryURL: "registry.gitlab.company.internal",
			want:        "https://gitlab.company.internal/api/v4",
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
			registryURL: "gitlab.eudev1.cyberarmorsoft.com:8080",
			want:        "https://gitlab.eudev1.cyberarmorsoft.com:8080/api/v4",
		},
		{
			name:        "GitLab instance with port and scheme",
			registryURL: "https://gitlab.eudev1.cyberarmorsoft.com:8080",
			want:        "https://gitlab.eudev1.cyberarmorsoft.com:8080/api/v4",
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
