package quay

/*
see https://docs.quay.io/api/swagger/
5/7/2022
*/
import (
	"fmt"
	"net/url"

	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/interfaces"
	"github.com/armosec/registryx/registries/defaultregistry"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

func NewQuayIORegistry(auth *authn.AuthConfig, registry *name.Registry, registryCfg *common.RegistryOptions) (interfaces.IRegistry, error) {
	if registry.Name() == "" {
		return nil, fmt.Errorf("must provide a non empty registry")
	}
	reg := &QuayioRegistry{DefaultRegistry: defaultregistry.DefaultRegistry{Registry: registry, Auth: auth}}
	reg.This = reg
	return reg, nil

}

type QuayioRegistry struct {
	defaultregistry.DefaultRegistry
}

func (reg *QuayioRegistry) GetAuth() *authn.AuthConfig {
	return reg.DefaultRegistry.GetAuth()
}
func (reg *QuayioRegistry) GetRegistry() *name.Registry {
	return reg.DefaultRegistry.GetRegistry()
}

func (reg *QuayioRegistry) getURL(urlSuffix string) *url.URL {

	return &url.URL{
		Scheme: reg.GetRegistry().Scheme(),
		Host:   reg.GetRegistry().RegistryStr(),
		Path:   fmt.Sprintf("/api/v1/%s", urlSuffix),
	}
}
