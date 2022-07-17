package quay

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/registries/defaultregistry"
	"github.com/google/go-containerregistry/pkg/authn"
)

func catalogOptionsToQuery(uri *url.URL, pagination common.PaginationOption, options common.CatalogOption) *url.URL {
	q := uri.Query()
	if options.IsPublic {
		q.Add("public", "true")
	}

	if options.IncludeLastModified {
		q.Add("last_modified", "true")
	}

	if options.Namespaces != "" {
		q.Add("namespace", options.Namespaces)
	}

	if pagination.Cursor != "" {
		q.Add("next_page", pagination.Cursor)
	}

	uri.RawQuery = q.Encode()

	return uri
}
func (reg *QuayioRegistry) Catalog(ctx context.Context, pagination common.PaginationOption, options common.CatalogOption, authenticator authn.Authenticator) ([]string, *common.PaginationOption, error) {
	//if quay is invalid we use it as public!!!!
	if err := common.ValidateAuth(reg.GetAuth()); err == nil {
		if authenticator == nil {
			authenticator = authn.FromConfig(*reg.GetAuth())
		}

		return reg.catalogQuayV2Auth(pagination, options)

	} else {
		options.IsPublic = true

	}

	return reg.catalogQuayProprietery(pagination, options)
}

func (reg *QuayioRegistry) catalogQuayV2Auth(pagination common.PaginationOption, options common.CatalogOption) ([]string, *common.PaginationOption, error) {

	//Token Request
	token, err := reg.GetV2Token(reg.HTTPClient, AUTH_URL)
	if err != nil {
		return nil, nil, err
	}
	reg.GetAuth().RegistryToken = token.Token
	uri := reg.DefaultRegistry.GetURL("_catalog")
	q := uri.Query()
	if pagination.Cursor != "" {
		q.Add("last", pagination.Cursor)
	}
	if pagination.Size > 0 {
		q.Add("n", fmt.Sprintf("%d", pagination.Size))
	}
	uri.RawQuery = q.Encode()
	req, err := http.NewRequest(http.MethodGet, uri.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.Token))
	resp, err := reg.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()
	repos := &defaultregistry.CatalogV2Response{}
	if err := json.NewDecoder(resp.Body).Decode(repos); err != nil {
		return nil, nil, err
	}
	pgn := &common.PaginationOption{Size: pagination.Size}
	if len(repos.Repositories) > 0 && pagination.Size > 0 {
		pgn.Cursor = repos.Repositories[len(repos.Repositories)-1]
	}

	return repos.Repositories, pgn, nil
}

func (reg *QuayioRegistry) catalogQuayProprietery(pagination common.PaginationOption, options common.CatalogOption) ([]string, *common.PaginationOption, error) {
	data, err := reg.CatalogAux(pagination, options)
	if err != nil {
		return nil, nil, err
	}
	repositories := data.Transform(pagination.Size)
	var pgn *common.PaginationOption = nil
	for data.Cursor != "" {
		pgn = &common.PaginationOption{Cursor: data.Cursor, Size: pagination.Size}

		repositories = append(repositories, data.Transform(0)...)

	}
	return repositories, pgn, nil
}

func (reg *QuayioRegistry) CatalogAux(pagination common.PaginationOption, options common.CatalogOption) (*QuayCatalogResponse, error) {
	uri := reg.getURL("repository")
	uri = catalogOptionsToQuery(uri, pagination, options)
	client := http.Client{}

	req, err := http.NewRequest("GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	data := &QuayCatalogResponse{}
	if err := json.Unmarshal(body, data); err != nil {
		return nil, err
	}
	body = nil
	return data, nil
}
