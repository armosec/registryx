package harbor

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/armosec/registryx/common"
	"github.com/armosec/registryx/interfaces"
	"github.com/armosec/registryx/registries/defaultregistry"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

func NewHarborRegistry(auth *authn.AuthConfig, registry *name.Registry, registryCfg *common.RegistryOptions) (interfaces.IRegistry, error) {
	if registry.Name() == "" {
		return nil, fmt.Errorf("must provide a non empty registry")
	}
	return &HarborRegistry{DefaultRegistry: defaultregistry.DefaultRegistry{Registry: registry, Auth: auth, Cfg: registryCfg}}, nil
}

type HarborRegistry struct {
	defaultregistry.DefaultRegistry
}

func (h *HarborRegistry) Catalog(ctx context.Context, pagination common.PaginationOption, options common.CatalogOption, authenticator authn.Authenticator) ([]string, *common.PaginationOption, error) {
	//if first pagination request set the page number (cursor) to 1
	if len(pagination.Cursor) == 0 {
		pagination.Cursor = "1"
	} else {
		//ensure pagination Cursor is a number
		_, err := strconv.Atoi(pagination.Cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid pagination, cursor must be an integer")
		}
	}
	//create list repos request
	req, err := h.repositoriesRequest(strconv.Itoa(pagination.Size), pagination.Cursor)
	if err != nil {
		return nil, nil, err
	}
	//create client according to registry configuration
	res, err := h.getClient().Do(req)
	if err != nil {
		return nil, nil, err
	}

	if err := transport.CheckError(res, http.StatusOK); err != nil {
		return nil, nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}
	//decode the repositories names from the response
	if repos, err := decodeObjectsNames(body); err != nil {
		return nil, nil, err
	} else {
		//get next pagination (can be nill if this is the last one)
		nextPagination, err := getNextPageOption(res)
		return repos, nextPagination, err
	}
}

func (h *HarborRegistry) List(repo name.Repository, pagination common.PaginationOption, options ...remote.Option) ([]string, *common.PaginationOption, error) {
	//create list tag request
	req, err := h.listTagsRequest(repo, strconv.Itoa(pagination.Size), pagination.Cursor)
	if err != nil {
		return nil, nil, err
	}
	//create client according to registry configuration
	res, err := h.getClient().Do(req)
	if err != nil {
		return nil, nil, err
	}

	defer res.Body.Close()

	//decode tags
	type tags struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}
	tagList := []string{}
	parsed := tags{}

	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, nil, err
	}
	//append tags names to tagList
	tagList = append(tagList, parsed.Tags...)
	//get next pagination option
	nextPagination, err := common.GetNextV2Pagination(res)
	return tagList, nextPagination, err

}

func getNextPageOption(res *http.Response) (*common.PaginationOption, error) {
	paginationHeaders := res.Header["Link"]
	//in harbor the link header should always exist, but check anyway
	if len(paginationHeaders) == 0 {
		return nil, nil
	}
	links := strings.Split(paginationHeaders[0], ",")
	//get the last header value
	lastLink := links[len(links)-1]
	lastLink = strings.Trim(lastLink, " ")
	//if the last link is pointing to the prev page then the current page was the last one
	if strings.Contains(lastLink, `rel="prev"`) {
		return nil, nil
	}
	//extract link
	if lastLink[0] != '<' {
		return nil, fmt.Errorf("failed to parse link header: missing '<' in: %s", lastLink)
	}
	end := strings.Index(lastLink, ">")
	if end == -1 {
		return nil, fmt.Errorf("failed to parse link header: missing '>' in: %s", lastLink)
	}
	lastLink = lastLink[1:end]

	//parse URL and get pagination query params
	linkURL, err := url.Parse(lastLink)
	if err != nil {
		return nil, err
	}
	queryParams := linkURL.Query()
	nextPagination := common.PaginationOption{}
	//get page_size and parse it as int
	if pageSizeValues, ok := queryParams["page_size"]; !ok {
		return nil, fmt.Errorf("page size is missing in next page header")
	} else {
		if pageSize, err := strconv.Atoi(pageSizeValues[0]); err != nil {
			return nil, fmt.Errorf("page size is not an integer in next page header")
		} else {
			nextPagination.Size = pageSize
		}
	}
	//get current page num
	if pageValues, ok := queryParams["page"]; !ok {
		return nil, fmt.Errorf("page number is missing in next page header")
	} else {
		if _, err := strconv.Atoi(pageValues[0]); err != nil {
			return nil, fmt.Errorf("page number is not an integer in next page header")
		} else {
			nextPagination.Cursor = pageValues[0]
		}
	}
	return &nextPagination, nil

}

func (h *HarborRegistry) getClient() *http.Client {
	var client *http.Client
	if h.Cfg.SkipTLSVerify() {
		skipVerifyTransport := &(*http.DefaultTransport.(*http.Transport)) // make shallow copy
		skipVerifyTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client = &http.Client{Transport: skipVerifyTransport}
	} else {
		client = &http.Client{}
	}
	return client
}

func (h *HarborRegistry) repositoriesRequest(pageSize string, pageNum string) (*http.Request, error) {
	uri := &url.URL{
		Scheme: h.requestScheme(),
		Host:   h.Registry.RegistryStr(),
	}
	//if project is specified then use the projects api to list repos in the project
	if len(h.Cfg.Project()) != 0 {
		uri.Path = fmt.Sprintf("/api/v2.0/projects/%s/repositories", h.Cfg.Project())
	} else {
		//no project - get all visible repositories
		uri.Path = "/api/v2.0/repositories"
	}

	if pageSize != "0" {
		//add pagination params
		paginationParams := url.Values{}
		paginationParams.Add("page_size", pageSize)
		paginationParams.Add("page", pageNum)
		uri.RawQuery = paginationParams.Encode()
	}

	req, err := http.NewRequest("GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}
	h.addAuthHeader(req)
	return req, nil
}

func (h *HarborRegistry) listTagsRequest(repo name.Repository, size string, cursor string) (*http.Request, error) {
	uri := &url.URL{
		Scheme: h.requestScheme(),
		Host:   repo.Registry.RegistryStr(),
		Path:   fmt.Sprintf("/v2/%s/tags/list", repo.RepositoryStr()),
	}

	if size != "0" {
		paginationParams := url.Values{}
		paginationParams.Add("n", size)
		if cursor != "" {
			paginationParams.Add("last", cursor)
		}
		uri.RawQuery = paginationParams.Encode()
	}

	req, err := http.NewRequest("GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}
	h.addAuthHeader(req)
	return req, nil
}

func (h *HarborRegistry) addAuthHeader(req *http.Request) {
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(h.Auth.Username+":"+h.Auth.Password)))
}

func (h *HarborRegistry) requestScheme() string {
	if h.Cfg.Insecure() {
		return "http"
	}
	return "https"
}
