package registries

import (
	"github.com/LiorAlafiArmo/registryx/common"
	"github.com/LiorAlafiArmo/registryx/interfaces"
	"github.com/LiorAlafiArmo/registryx/registries/defaultregistry"
	"github.com/LiorAlafiArmo/registryx/registries/quay"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

func Factory(auth *authn.AuthConfig, registry *name.Registry, registryOptions *common.RegistryOptions) (interfaces.IRegistry, error) {
	switch registry.Name() {
	case "quay.io":
		return quay.NewQuayIORegistry(auth, registry, registryOptions)
	default:
		return defaultregistry.NewRegistry(auth, registry, registryOptions)
	}
}
