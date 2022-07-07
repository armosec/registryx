package defaultregistry

import (
	"context"
	"fmt"
	"net/url"

	"github.com/LiorAlafiArmo/registryx/common"
	"github.com/LiorAlafiArmo/registryx/interfaces"
	"github.com/LiorAlafiArmo/registryx/registries/dockerregistry"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// this is just a wrapper around the go-container & remote catalog
type DefaultRegistry struct {
	Registry *name.Registry
	Auth     *authn.AuthConfig
	Cfg      *common.RegistryOptions
}

func NewRegistry(auth *authn.AuthConfig, registry *name.Registry, registryCfg *common.RegistryOptions) (interfaces.IRegistry, error) {
	if registry.Name() == "" {
		return nil, fmt.Errorf("must provide a non empty registry")
	}

	return &DefaultRegistry{Auth: auth, Registry: registry, Cfg: registryCfg}, nil

}

func (reg *DefaultRegistry) GetAuth() *authn.AuthConfig {
	return reg.Auth
}
func (reg *DefaultRegistry) GetRegistry() *name.Registry {
	return reg.Registry
}

func (reg *DefaultRegistry) getURL(urlSuffix string) *url.URL {

	return &url.URL{
		Scheme: reg.Registry.Scheme(),
		Host:   reg.Registry.RegistryStr(),
		Path:   fmt.Sprintf("/v2/%s", urlSuffix),
	}
}

func (reg *DefaultRegistry) Catalog(ctx context.Context, pagination common.PaginationOption, options common.CatalogOption, authenticator authn.Authenticator) ([]string, error) {
	regis := reg.GetRegistry().Name()
	if regis == "index.docker.io" && authenticator == nil {
		token, err := dockerregistry.Token(reg.GetAuth(), reg.GetRegistry())
		if err != nil {
			return nil, err
		}
		if reg.Auth == nil {
			reg.Auth = &authn.AuthConfig{}
		}
		reg.Auth.RegistryToken = token.Token
	}
	//auth part not working though w/o removing scope
	if err := common.ValidateAuth(reg.GetAuth()); err == nil {
		// res, err := remote.CatalogPage(*reg.GetRegistry(), pagination.Cursor, pagination.Size, remote.WithAuth(authn.FromConfig(*reg.GetAuth())))
		if authenticator == nil {
			authenticator = authn.FromConfig(*reg.GetAuth())
		}
		res, err := remote.Catalog(ctx, *reg.GetRegistry(), remote.WithAuth(authenticator))
		if err != nil {
			return nil, err
		}
		return res, err
	}
	if regis == "index.docker.io" {
		return remote.Catalog(ctx, *reg.GetRegistry(), remote.WithAuth(authn.FromConfig(*reg.GetAuth())))
	}
	return remote.CatalogPage(*reg.GetRegistry(), pagination.Cursor, pagination.Size, remote.WithAuth(authn.Anonymous))
}
