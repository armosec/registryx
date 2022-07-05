package quay

/*
see https://docs.quay.io/api/swagger/
5/7/2022
*/
import (
	"fmt"
	"net/url"

	"github.com/LiorAlafiArmo/registryx/interfaces"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

func NewQuayIORegistry(auth *authn.AuthConfig, registry *name.Registry) (interfaces.IRegistry, error) {
	if registry.Name() == "" {
		return nil, fmt.Errorf("must provide a non empty registry")
	}

	return &QuayioRegistry{auth: auth, registry: registry}, nil

}

type QuayioRegistry struct {
	auth     *authn.AuthConfig
	registry *name.Registry
}

func (reg *QuayioRegistry) GetAuth() *authn.AuthConfig {
	return reg.auth
}
func (reg *QuayioRegistry) GetRegistry() *name.Registry {
	return reg.registry
}

func (reg *QuayioRegistry) getURL(urlSuffix string) *url.URL {

	return &url.URL{
		Scheme: reg.registry.Scheme(),
		Host:   reg.registry.RegistryStr(),
		Path:   fmt.Sprintf("/api/v1/%s", urlSuffix),
	}
}
