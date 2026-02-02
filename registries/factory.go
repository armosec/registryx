package registries

import (
	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/interfaces"
	"github.com/armosec/registryx/registries/defaultregistry"
	"github.com/armosec/registryx/registries/harbor"
	"github.com/armosec/registryx/registries/quay"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

func Factory(auth *authn.AuthConfig, registryName string, registryOptions *common.RegistryOptions) (interfaces.IRegistry, error) {
	kind, registry, err := makeRegistry(registryOptions, registryName)
	if err != nil {
		return nil, err
	}
	switch kind {
	case common.Quay:
		return quay.NewQuayIORegistry(auth, registry, registryOptions)
	case common.Harbor:
		return harbor.NewHarborRegistry(auth, registry, registryOptions)
	default:
		return defaultregistry.NewRegistry(auth, registry, registryOptions)
	}
}

func makeRegistry(registryOptions *common.RegistryOptions, registryName string) (common.RegistryKind, *name.Registry, error) {
	opts := []name.Option{}
	var kind common.RegistryKind
	var err error
	if registryOptions != nil {
		kind = registryOptions.Kind()
		opts = append(opts, name.WithDefaultRegistry(registryOptions.DefaultRegistry()))
		opts = append(opts, name.WithDefaultTag(registryOptions.DefaultTag()))
		if registryOptions.Insecure() {
			opts = append(opts, name.Insecure)
		}
		if registryOptions.Strict() {
			opts = append(opts, name.StrictValidation)
		} else {
			opts = append(opts, name.WeakValidation)
		}
	} else if kind, err = common.GetRegistryKind(registryName); err != nil {
		return kind, nil, err
	}
	registry, err := name.NewRegistry(registryName, opts...)

	return kind, &registry, err
}
