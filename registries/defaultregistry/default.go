package defaultregistry

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/interfaces"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type CatalogV2Response struct {
	Repositories []string `json:"repositories"`
}

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

func (reg *DefaultRegistry) GetMaxPageSize() int {
	return 1000
}

func (reg *DefaultRegistry) GetAuth() *authn.AuthConfig {
	return reg.Auth
}
func (reg *DefaultRegistry) GetRegistry() *name.Registry {
	return reg.Registry
}

func (reg *DefaultRegistry) GetURL(urlSuffix string) *url.URL {

	return &url.URL{
		Scheme: reg.Registry.Scheme(),
		Host:   reg.Registry.RegistryStr(),
		Path:   fmt.Sprintf("/v2/%s", urlSuffix),
	}
}

func (reg *DefaultRegistry) List(repoName string, pagination common.PaginationOption, options ...remote.Option) ([]string, *common.PaginationOption, error) {
	repoData, err := common.MakeRepoWithRegistry(repoName, reg.Registry)
	if err != nil {
		return nil, nil, err
	}
	tags, err := remote.List(*repoData, options...)
	//TODO handle pagination
	return tags, nil, err
}

// this is the default catalog implementation uses remote(for now)
func (reg *DefaultRegistry) Catalog(ctx context.Context, pagination common.PaginationOption, options common.CatalogOption, authenticator authn.Authenticator) ([]string, *common.PaginationOption, error) {

	if err := common.ValidateAuth(reg.GetAuth()); err == nil {
		if authenticator == nil {
			authenticator = authn.FromConfig(*reg.GetAuth())
		}
		res, err := remote.CatalogPage(*reg.GetRegistry(), pagination.Cursor, pagination.Size, remote.WithAuth(authn.FromConfig(*reg.GetAuth())))
		if err != nil {
			return nil, nil, err
		}
		return res, nil, err
	}
	repos, err := remote.CatalogPage(*reg.GetRegistry(), pagination.Cursor, pagination.Size, remote.WithAuth(authn.Anonymous))
	return repos, common.CalcNextV2Pagination(repos, pagination.Size), err
}

func (reg *DefaultRegistry) GetV2Token(client *http.Client, url string) (*common.V2TokenResponse, error) {
	if reg.GetAuth() == nil {
		return nil, fmt.Errorf("no authorization found")
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if reg.GetAuth().Username != "" && reg.GetAuth().Password != "" {
		usrpwd := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", reg.GetAuth().Username, reg.GetAuth().Password)))
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", usrpwd))
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	token := &common.V2TokenResponse{}

	if err := json.NewDecoder(resp.Body).Decode(token); err != nil {
		return nil, err
	}

	if token.Token == "" {
		return nil, fmt.Errorf("recieved an empty token")
	}
	return token, nil
}
