package common

import (
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
)

func TestMakeRepoWithRegistry(t *testing.T) {
	tests := []struct {
		name           string
		repoName       string
		registryHost   string
		wantRegistry   string
		wantRepository string
		wantErr        bool
	}{
		// --- Basic cases across registry providers ---
		{
			name:           "simple repo name with ACR registry",
			repoName:       "myapp",
			registryHost:   "myacr.azurecr.io",
			wantRegistry:   "myacr.azurecr.io",
			wantRepository: "myapp",
		},
		{
			name:           "nested repo name with ACR registry",
			repoName:       "team/myapp",
			registryHost:   "myacr.azurecr.io",
			wantRegistry:   "myacr.azurecr.io",
			wantRepository: "team/myapp",
		},
		{
			name:           "simple repo name with quay.io registry",
			repoName:       "org/myrepo",
			registryHost:   "quay.io",
			wantRegistry:   "quay.io",
			wantRepository: "org/myrepo",
		},
		{
			name:           "simple repo name with ECR registry",
			repoName:       "my-service",
			registryHost:   "123456789012.dkr.ecr.us-east-1.amazonaws.com",
			wantRegistry:   "123456789012.dkr.ecr.us-east-1.amazonaws.com",
			wantRepository: "my-service",
		},
		{
			name:           "nested repo with ECR registry",
			repoName:       "team/my-service",
			registryHost:   "123456789012.dkr.ecr.us-east-1.amazonaws.com",
			wantRegistry:   "123456789012.dkr.ecr.us-east-1.amazonaws.com",
			wantRepository: "team/my-service",
		},
		{
			name:           "simple repo with GCR registry",
			repoName:       "my-project/my-image",
			registryHost:   "gcr.io",
			wantRegistry:   "gcr.io",
			wantRepository: "my-project/my-image",
		},
		{
			name:           "simple repo with Harbor registry",
			repoName:       "project/repo",
			registryHost:   "harbor.example.com",
			wantRegistry:   "harbor.example.com",
			wantRepository: "project/repo",
		},
		{
			name:           "simple repo with Nexus registry (port in host)",
			repoName:       "my-image",
			registryHost:   "nexus.example.com:8443",
			wantRegistry:   "nexus.example.com:8443",
			wantRepository: "my-image",
		},
		{
			name:           "simple repo with GitLab registry",
			repoName:       "group/project",
			registryHost:   "registry.gitlab.com",
			wantRegistry:   "registry.gitlab.com",
			wantRepository: "group/project",
		},

		// --- The bug scenario: mirrored images with registry-like prefixes in ACR ---
		{
			name:           "docker.io prefixed repo in ACR",
			repoName:       "docker.io/prom/blackbox-exporter",
			registryHost:   "escdev.azurecr.io",
			wantRegistry:   "escdev.azurecr.io",
			wantRepository: "docker.io/prom/blackbox-exporter",
		},
		{
			name:           "docker.io prefixed library image in ACR",
			repoName:       "docker.io/library/nginx",
			registryHost:   "myacr.azurecr.io",
			wantRegistry:   "myacr.azurecr.io",
			wantRepository: "docker.io/library/nginx",
		},
		{
			name:           "docker.io prefixed grafana image in ACR",
			repoName:       "docker.io/grafana/grafana-oss",
			registryHost:   "myacr.azurecr.io",
			wantRegistry:   "myacr.azurecr.io",
			wantRepository: "docker.io/grafana/grafana-oss",
		},
		{
			name:           "quay.io prefixed repo in ACR (mirrored image)",
			repoName:       "quay.io/prometheus/node-exporter",
			registryHost:   "myacr.azurecr.io",
			wantRegistry:   "myacr.azurecr.io",
			wantRepository: "quay.io/prometheus/node-exporter",
		},
		{
			name:           "gcr.io prefixed repo in ACR (mirrored image)",
			repoName:       "gcr.io/google-containers/pause",
			registryHost:   "myacr.azurecr.io",
			wantRegistry:   "myacr.azurecr.io",
			wantRepository: "gcr.io/google-containers/pause",
		},
		{
			name:           "ghcr.io prefixed repo in ACR (mirrored image)",
			repoName:       "ghcr.io/external-secrets/external-secrets",
			registryHost:   "myacr.azurecr.io",
			wantRegistry:   "myacr.azurecr.io",
			wantRepository: "ghcr.io/external-secrets/external-secrets",
		},

		// --- Single-component repo names (no slash) ---
		{
			name:           "single-component repo with ACR",
			repoName:       "nginx",
			registryHost:   "myacr.azurecr.io",
			wantRegistry:   "myacr.azurecr.io",
			wantRepository: "nginx",
		},
		{
			name:           "single-component repo with ECR",
			repoName:       "myapp",
			registryHost:   "123456789012.dkr.ecr.eu-west-1.amazonaws.com",
			wantRegistry:   "123456789012.dkr.ecr.eu-west-1.amazonaws.com",
			wantRepository: "myapp",
		},

		// --- Deeply nested paths ---
		{
			name:           "deeply nested repo path in ACR",
			repoName:       "a/b/c/d/my-image",
			registryHost:   "myacr.azurecr.io",
			wantRegistry:   "myacr.azurecr.io",
			wantRepository: "a/b/c/d/my-image",
		},

		// --- Docker Hub as the actual target registry ---
		{
			name:           "Docker Hub as target with standard repo",
			repoName:       "library/nginx",
			registryHost:   "index.docker.io",
			wantRegistry:   "index.docker.io",
			wantRepository: "library/nginx",
		},
		{
			name:           "Docker Hub as target with org repo",
			repoName:       "prom/blackbox-exporter",
			registryHost:   "index.docker.io",
			wantRegistry:   "index.docker.io",
			wantRepository: "prom/blackbox-exporter",
		},

		// --- Repos with special characters (dots, underscores, hyphens) ---
		{
			name:           "repo name with dots and hyphens",
			repoName:       "my-team/my.dotted.image-name",
			registryHost:   "myacr.azurecr.io",
			wantRegistry:   "myacr.azurecr.io",
			wantRepository: "my-team/my.dotted.image-name",
		},
		{
			name:           "repo name with underscores",
			repoName:       "my_org/my_image",
			registryHost:   "myacr.azurecr.io",
			wantRegistry:   "myacr.azurecr.io",
			wantRepository: "my_org/my_image",
		},

		// --- Error cases ---
		{
			name:         "empty repo name",
			repoName:     "",
			registryHost: "myacr.azurecr.io",
			wantErr:      true,
		},
		{
			name:         "repo name with invalid characters",
			repoName:     "INVALID/UPPERCASE",
			registryHost: "myacr.azurecr.io",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var registry *name.Registry
			if tt.registryHost != "" {
				reg, err := name.NewRegistry(tt.registryHost)
				if err != nil {
					t.Fatalf("failed to create registry: %v", err)
				}
				registry = &reg
			}

			repo, err := MakeRepoWithRegistry(tt.repoName, registry)
			if (err != nil) != tt.wantErr {
				t.Fatalf("MakeRepoWithRegistry() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			gotRegistry := repo.RegistryStr()
			if gotRegistry != tt.wantRegistry {
				t.Errorf("registry = %q, want %q", gotRegistry, tt.wantRegistry)
			}

			gotRepo := repo.RepositoryStr()
			if gotRepo != tt.wantRepository {
				t.Errorf("repository = %q, want %q", gotRepo, tt.wantRepository)
			}
		})
	}
}

func TestMakeRepoWithRegistry_NilRegistry(t *testing.T) {
	tests := []struct {
		name           string
		repoName       string
		wantRegistry   string
		wantRepository string
	}{
		{
			name:           "library image defaults to Docker Hub",
			repoName:       "library/nginx",
			wantRegistry:   "index.docker.io",
			wantRepository: "library/nginx",
		},
		{
			name:           "org image defaults to Docker Hub",
			repoName:       "prom/blackbox-exporter",
			wantRegistry:   "index.docker.io",
			wantRepository: "prom/blackbox-exporter",
		},
		{
			name:           "fully qualified reference keeps its own registry",
			repoName:       "myacr.azurecr.io/myapp",
			wantRegistry:   "myacr.azurecr.io",
			wantRepository: "myapp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := MakeRepoWithRegistry(tt.repoName, nil)
			if err != nil {
				t.Fatalf("MakeRepoWithRegistry() error = %v", err)
			}

			if repo.RegistryStr() != tt.wantRegistry {
				t.Errorf("registry = %q, want %q", repo.RegistryStr(), tt.wantRegistry)
			}
			if repo.RepositoryStr() != tt.wantRepository {
				t.Errorf("repository = %q, want %q", repo.RepositoryStr(), tt.wantRepository)
			}
		})
	}
}
