package quay

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/LiorAlafiArmo/registryx/common"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/remote"
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
func (reg *QuayioRegistry) Catalog(ctx context.Context, pagination common.PaginationOption, options common.CatalogOption) ([]string, error) {
	if err := common.ValidateAuth(reg.auth); err != nil && !options.IsPublic && options.Namespaces == "" {
		return nil, fmt.Errorf("quay.io supports no/empty auth information only for public/namespaced registries")
	}

	//auth part not working though w/o removing scope
	if err := common.ValidateAuth(reg.auth); err == nil {
		res, err := remote.CatalogPage(*reg.registry, pagination.Cursor, pagination.Size, remote.WithAuth(authn.FromConfig(*reg.auth)))
		if err != nil {
			return nil, err
		}
		fmt.Printf("%v", res)
		return res, err
	}

	uri := reg.getURL("repository")
	uri = catalogOptionsToQuery(uri, pagination, options)
	client := http.Client{}

	req, err := http.NewRequest("GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}
	// req = req.WithContext(ctx)

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
	repositories := data.Transform(pagination.Size)
	if len(repositories) < pagination.Size && data.Cursor != "" {

	}
	return repositories, nil
}
