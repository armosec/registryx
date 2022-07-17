package quay

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

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
	if err := common.ValidateAuth(reg.GetAuth()); err != nil && !options.IsPublic && options.Namespaces == "" {
		return nil, nil, fmt.Errorf("quay.io supports no/empty auth information only for public/namespaced registries")
	}

	//auth part not working though w/o removing scope
	if err := common.ValidateAuth(reg.GetAuth()); err == nil {
		if authenticator == nil {
			authenticator = authn.FromConfig(*reg.GetAuth())
		}

		return reg.catalogQuayV2Auth(pagination, options)

		// if err != nil {
		// 	return nil, nil, err
		// }
		// fmt.Printf("%v", res)
		// return res, common.CalcNextV2Pagination(res, pagination.Size), err
	} else {
		options.IsPublic = true
	}

	return reg.catalogQuayProprietery(pagination, options)
}

func (reg *QuayioRegistry) catalogQuayV2Auth(pagination common.PaginationOption, options common.CatalogOption) ([]string, *common.PaginationOption, error) {
	client := http.Client{Timeout: time.Duration(150) * time.Second}

	//Token Request
	token, err := reg.GetV2Token(&client, AUTH_URL)
	if err != nil {
		return nil, nil, err
	}

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
	resp, err := client.Do(req)
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
	for data.Cursor != "" {
		pagination.Cursor = data.Cursor
		data, err := reg.CatalogAux(pagination, options)
		if err != nil {
			return repositories, common.CalcNextV2Pagination(repositories, pagination.Size), fmt.Errorf("partial success, failed due to %s", err.Error())
		}
		repositories = append(repositories, data.Transform(0)...)

	}
	return repositories, nil, nil
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
